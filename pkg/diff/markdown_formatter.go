package diff

import (
	"fmt"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatMarkdown formats the drift for Markdown (GitHub, Slack, etc.)
func (f *DiffFormatter) FormatMarkdown(alert *types.DriftAlert) string {
	var b strings.Builder

	// Title
	b.WriteString(fmt.Sprintf("## 🚨 Drift Detected: `%s.%s`\n\n", alert.ResourceType, alert.ResourceName))

	// Severity Badge
	severityEmoji := map[string]string{
		"critical": "🔴",
		"high":     "🟠",
		"medium":   "🟡",
		"low":      "🟢",
	}
	b.WriteString(fmt.Sprintf("**Severity:** %s **%s**\n\n", severityEmoji[alert.Severity], strings.ToUpper(alert.Severity)))

	// Changed Attribute
	b.WriteString(fmt.Sprintf("**Changed Attribute:** `%s`\n\n", alert.Attribute))

	// Diff
	b.WriteString("### Value Change\n\n")
	b.WriteString("```diff\n")
	b.WriteString(fmt.Sprintf("- %s\n", f.formatValue(alert.OldValue)))
	b.WriteString(fmt.Sprintf("+ %s\n", f.formatValue(alert.NewValue)))
	b.WriteString("```\n\n")

	// User Info
	b.WriteString("### Changed By\n\n")
	b.WriteString(fmt.Sprintf("- **User:** %s\n", alert.UserIdentity.UserName))
	b.WriteString(fmt.Sprintf("- **Account:** %s\n", alert.UserIdentity.AccountID))
	b.WriteString(fmt.Sprintf("- **Time:** %s\n\n", alert.Timestamp))

	// Terraform Code
	b.WriteString("### Terraform State\n\n")
	b.WriteString("```hcl\n")
	b.WriteString(f.formatTerraformResource(alert, alert.OldValue))
	b.WriteString("```\n\n")

	b.WriteString("### Actual Configuration\n\n")
	b.WriteString("```hcl\n")
	b.WriteString(f.formatTerraformResource(alert, alert.NewValue))
	b.WriteString("```\n\n")

	// Actions
	b.WriteString("### Recommended Actions\n\n")
	b.WriteString("- [ ] Review change with user\n")
	b.WriteString("- [ ] Update Terraform code if intentional\n")
	b.WriteString(fmt.Sprintf("- [ ] Run `terraform apply -target=%s.%s` to revert\n\n",
		alert.ResourceType, alert.ResourceName))

	return b.String()
}

// FormatUnmanagedResourceMarkdown formats an unmanaged resource alert for Slack/Discord
func (f *DiffFormatter) FormatUnmanagedResourceMarkdown(alert *types.UnmanagedResourceAlert) string {
	var b strings.Builder

	emoji := "⚠️"
	if alert.Severity == "critical" {
		emoji = "🚨"
	}

	b.WriteString(fmt.Sprintf("%s **UNMANAGED RESOURCE DETECTED**\n\n", emoji))
	b.WriteString(fmt.Sprintf("**Severity:** %s\n", strings.ToUpper(alert.Severity)))
	b.WriteString(fmt.Sprintf("**Resource Type:** `%s`\n", alert.ResourceType))
	b.WriteString(fmt.Sprintf("**Resource ID:** `%s`\n", alert.ResourceID))
	b.WriteString(fmt.Sprintf("**Event:** %s\n", alert.EventName))
	b.WriteString(fmt.Sprintf("**When:** %s\n", alert.Timestamp))
	b.WriteString(fmt.Sprintf("**Who:** %s (`%s`)\n", alert.UserIdentity.UserName, alert.UserIdentity.ARN))
	b.WriteString(fmt.Sprintf("\n**Reason:** %s\n\n", alert.Reason))
	b.WriteString("This resource is not managed by Terraform. Consider importing it:\n")
	b.WriteString(fmt.Sprintf("```\nterraform import %s.resource_name %s\n```\n",
		alert.ResourceType, alert.ResourceID))

	return b.String()
}
