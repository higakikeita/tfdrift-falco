package detector

import (
	"context"
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// handleAutoImport handles automatic terraform import for unmanaged resources
func (d *Detector) handleAutoImport(ctx context.Context, event *types.Event) {
	log.Infof("Auto-import triggered for %s (%s)", event.ResourceID, event.ResourceType)

	// Create approval request
	userIdentity := fmt.Sprintf("%s (%s)", event.UserIdentity.UserName, event.UserIdentity.ARN)
	request := d.approvalManager.RequestApproval(
		event.ResourceType,
		event.ResourceID,
		event.Changes,
		userIdentity,
	)

	var result *terraform.ImportResult
	var err error

	// Handle based on approval mode
	if d.cfg.AutoImport.RequireApproval {
		// Manual approval mode - prompt user
		approved, promptErr := d.approvalManager.PromptForApproval(ctx, request)
		if promptErr != nil {
			log.Errorf("Failed to prompt for approval: %v", promptErr)
			return
		}

		if approved {
			fmt.Printf("🚀 Executing: %s\n", request.ImportCommand.String())
			result, err = d.approvalManager.ApproveAndExecute(ctx, request.ID, "console-user")
		} else {
			log.Info("Import rejected by user")
			return
		}
	} else {
		// Auto-approval mode - check whitelist
		result, err = d.approvalManager.AutoApproveIfAllowed(ctx, request, d.cfg.AutoImport.AllowedResources)
		if err != nil {
			log.Warnf("Auto-approval denied: %v", err)
			return
		}
	}

	// Handle result
	if err != nil {
		log.Errorf("Import failed: %v", err)
		fmt.Printf("❌ Import failed: %v\n", err)
		return
	}

	if result.Success {
		fmt.Println("✅ Import successful!")
		if result.GeneratedCode != "" {
			// Save the generated code to output directory
			outputFile := fmt.Sprintf("%s/%s_%s.tf",
				d.cfg.AutoImport.OutputDir,
				event.ResourceType,
				result.Command.ResourceName)
			fmt.Printf("📄 Generated Terraform code:\n%s\n", result.GeneratedCode)
			fmt.Printf("💡 Save this to: %s\n", outputFile)
		}
		log.Infof("Successfully imported %s", event.ResourceID)
	} else {
		fmt.Printf("❌ Import failed: %s\n", result.Error)
		log.Errorf("Import failed for %s: %s", event.ResourceID, result.Error)
	}
}
