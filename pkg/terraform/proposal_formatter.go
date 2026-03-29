package terraform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatProposalMarkdown formats a remediation proposal as Markdown suitable for PR descriptions
func FormatProposalMarkdown(proposal *types.RemediationProposal) string {
	if proposal == nil {
		return ""
	}

	var md strings.Builder

	md.WriteString("## Drift Auto-Remediation Proposal\n\n")

	// Summary section
	md.WriteString("### Summary\n\n")
	md.WriteString(proposal.Description)
	md.WriteString("\n\n")

	// Details section
	md.WriteString("### Details\n\n")
	md.WriteString(fmt.Sprintf("- **Type**: %s\n", proposal.AlertType))
	md.WriteString(fmt.Sprintf("- **Severity**: %s\n", proposal.Severity))
	md.WriteString(fmt.Sprintf("- **Resource Type**: %s\n", proposal.ResourceType))
	md.WriteString(fmt.Sprintf("- **Resource ID**: %s\n", proposal.ResourceID))
	if proposal.ResourceName != "" {
		md.WriteString(fmt.Sprintf("- **Resource Name**: %s\n", proposal.ResourceName))
	}
	if proposal.Provider != "" {
		md.WriteString(fmt.Sprintf("- **Provider**: %s\n", proposal.Provider))
	}
	md.WriteString(fmt.Sprintf("- **Status**: %s\n", proposal.Status))
	md.WriteString(fmt.Sprintf("- **Created**: %s\n", proposal.CreatedAt))
	md.WriteString("\n")

	// Proposed Terraform Code
	md.WriteString("### Proposed Terraform Code\n\n")
	md.WriteString("```hcl\n")
	md.WriteString(proposal.TerraformCode)
	md.WriteString("```\n\n")

	// Commands section
	md.WriteString("### Remediation Commands\n\n")
	md.WriteString("#### Import Command\n\n")
	md.WriteString("```bash\n")
	md.WriteString(proposal.ImportCommand)
	md.WriteString("\n```\n\n")

	md.WriteString("#### Plan Command\n\n")
	md.WriteString("```bash\n")
	md.WriteString(proposal.PlanCommand)
	md.WriteString("\n```\n\n")

	// Attributes section (if present)
	if len(proposal.Attributes) > 0 {
		md.WriteString("### Detected Attributes\n\n")
		for key, value := range proposal.Attributes {
			md.WriteString(fmt.Sprintf("- `%s`: %v\n", key, value))
		}
		md.WriteString("\n")
	}

	// Instructions
	md.WriteString("### Next Steps\n\n")
	md.WriteString("1. Review the proposed Terraform code above\n")
	md.WriteString("2. Run the import command to add the resource to your state (if unmanaged)\n")
	md.WriteString("3. Update your Terraform configuration with the proposed code\n")
	md.WriteString("4. Run the plan command to verify the changes\n")
	md.WriteString("5. Apply the changes using `terraform apply`\n\n")

	// Footer
	md.WriteString("---\n")
	md.WriteString(fmt.Sprintf("*Proposal ID: %s*\n", proposal.ID))

	return md.String()
}

// FormatProposalJSON formats a remediation proposal as JSON
func FormatProposalJSON(proposal *types.RemediationProposal) ([]byte, error) {
	if proposal == nil {
		return nil, fmt.Errorf("proposal cannot be nil")
	}

	// Create a JSON-serializable version with proper types
	jsonProposal := map[string]interface{}{
		"id":               proposal.ID,
		"alert_type":       proposal.AlertType,
		"provider":         proposal.Provider,
		"resource_type":    proposal.ResourceType,
		"resource_id":      proposal.ResourceID,
		"resource_name":    proposal.ResourceName,
		"severity":         proposal.Severity,
		"description":      proposal.Description,
		"terraform_code":   proposal.TerraformCode,
		"import_command":   proposal.ImportCommand,
		"plan_command":     proposal.PlanCommand,
		"status":           proposal.Status,
		"created_at":       proposal.CreatedAt,
		"attributes":       proposal.Attributes,
	}

	if proposal.PRUrl != "" {
		jsonProposal["pr_url"] = proposal.PRUrl
	}
	if proposal.PRNumber > 0 {
		jsonProposal["pr_number"] = proposal.PRNumber
	}

	return json.MarshalIndent(jsonProposal, "", "  ")
}
