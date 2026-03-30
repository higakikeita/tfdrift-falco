// Package detector provides drift detection and analysis.
package detector

import (
	"context"
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/keitahigaki/tfdrift-falco/pkg/vcs"
	log "github.com/sirupsen/logrus"
)

// handleRemediation generates a remediation proposal for a drift alert,
// optionally creates a GitHub PR, and broadcasts the proposal.
func (d *Detector) handleRemediation(ctx context.Context, alert *types.DriftAlert) {
	if !d.cfg.Remediation.Enabled {
		return
	}

	gen := terraform.NewRemediationGenerator()
	proposal := gen.GenerateForDrift(alert)
	if proposal == nil {
		return
	}

	log.WithFields(log.Fields{
		"resource_id":   proposal.ResourceID,
		"resource_type": proposal.ResourceType,
		"severity":      proposal.Severity,
	}).Info("Remediation proposal generated")

	// Create GitHub PR if configured
	if d.cfg.Remediation.CreatePRs && d.cfg.GitHub.Enabled {
		d.createRemediationPR(ctx, proposal)
	}

	// Broadcast the proposal via WebSocket/SSE
	if d.broadcaster != nil {
		d.broadcaster.Broadcast(broadcaster.Event{
			Type: "remediation_proposal",
			Payload: map[string]interface{}{
				"id":             proposal.ID,
				"resource_type":  proposal.ResourceType,
				"resource_id":    proposal.ResourceID,
				"severity":       proposal.Severity,
				"status":         proposal.Status,
				"description":    proposal.Description,
				"import_command": proposal.ImportCommand,
				"plan_command":   proposal.PlanCommand,
				"pr_url":         proposal.PRUrl,
			},
		})
	}
}

// handleUnmanagedRemediation generates a remediation proposal for an unmanaged resource.
func (d *Detector) handleUnmanagedRemediation(ctx context.Context, event *types.Event) {
	if !d.cfg.Remediation.Enabled {
		return
	}

	gen := terraform.NewRemediationGenerator()
	proposal := gen.GenerateForUnmanaged(event)
	if proposal == nil {
		return
	}

	log.WithFields(log.Fields{
		"resource_id":   proposal.ResourceID,
		"resource_type": proposal.ResourceType,
	}).Info("Unmanaged resource remediation proposal generated")

	if d.cfg.Remediation.CreatePRs && d.cfg.GitHub.Enabled {
		d.createRemediationPR(ctx, proposal)
	}

	if d.broadcaster != nil {
		d.broadcaster.Broadcast(broadcaster.Event{
			Type: "remediation_proposal",
			Payload: map[string]interface{}{
				"id":             proposal.ID,
				"resource_type":  proposal.ResourceType,
				"resource_id":    proposal.ResourceID,
				"severity":       proposal.Severity,
				"status":         proposal.Status,
				"description":    proposal.Description,
				"import_command": proposal.ImportCommand,
				"plan_command":   proposal.PlanCommand,
				"pr_url":         proposal.PRUrl,
			},
		})
	}
}

func (d *Detector) createRemediationPR(ctx context.Context, proposal *types.RemediationProposal) {
	if d.cfg.Remediation.DryRun {
		log.Info("Dry run: skipping GitHub PR creation")
		return
	}

	token := d.cfg.GitHub.Token
	if token == "" {
		log.Warn("GitHub token not configured, skipping PR creation")
		return
	}

	client := vcs.NewGitHubClient(
		d.cfg.GitHub.Owner,
		d.cfg.GitHub.Repo,
		d.cfg.GitHub.Branch,
		token,
	)

	branchName := fmt.Sprintf("remediation/drift-%s-%s", proposal.ResourceType, proposal.ID[:8])
	body := terraform.FormatProposalMarkdown(proposal)

	files := map[string]string{}
	if proposal.TerraformCode != "" {
		fileName := fmt.Sprintf("remediation/%s_%s.tf", proposal.ResourceType, proposal.ResourceName)
		files[fileName] = proposal.TerraformCode
	}

	if len(files) == 0 {
		log.Debug("No terraform files to commit, skipping PR")
		return
	}

	result, err := client.CreatePR(ctx, &vcs.PRRequest{
		Title:      fmt.Sprintf("fix: remediate drift in %s.%s", proposal.ResourceType, proposal.ResourceName),
		Body:       body,
		BranchName: branchName,
		Files:      files,
		CommitMsg:  fmt.Sprintf("fix: auto-remediation for %s drift\n\n%s", proposal.ResourceID, proposal.Description),
	})
	if err != nil {
		log.WithError(err).Error("Failed to create remediation PR")
		return
	}

	proposal.PRUrl = result.URL
	proposal.PRNumber = result.Number
	proposal.Status = types.RemediationApproved

	log.WithFields(log.Fields{
		"pr_url":    result.URL,
		"pr_number": result.Number,
	}).Info("Remediation PR created")
}
