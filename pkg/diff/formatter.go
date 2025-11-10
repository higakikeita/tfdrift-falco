package diff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
)

// DiffFormatter formats drift differences in various output formats
type DiffFormatter struct {
	colorEnabled bool
}

// NewFormatter creates a new diff formatter
func NewFormatter(colorEnabled bool) *DiffFormatter {
	return &DiffFormatter{
		colorEnabled: colorEnabled,
	}
}

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorBold   = "\033[1m"
)

// FormatConsole formats the drift for console output with colors
func (f *DiffFormatter) FormatConsole(alert *detector.DriftAlert) string {
	var b strings.Builder

	// Header
	severityColor := f.getSeverityColor(alert.Severity)
	b.WriteString(f.color(severityColor, fmt.Sprintf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")))
	b.WriteString(f.color(ColorBold, fmt.Sprintf("🚨 DRIFT DETECTED: %s.%s\n", alert.ResourceType, alert.ResourceName)))
	b.WriteString(f.color(severityColor, fmt.Sprintf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")))

	// Severity
	b.WriteString(f.color(ColorBold, "\n📊 Severity: "))
	b.WriteString(f.color(severityColor, strings.ToUpper(alert.Severity)))
	b.WriteString("\n")

	// Resource Info
	b.WriteString(f.color(ColorBold, "\n📦 Resource:\n"))
	b.WriteString(fmt.Sprintf("  Type:       %s\n", f.color(ColorCyan, alert.ResourceType)))
	b.WriteString(fmt.Sprintf("  Name:       %s\n", f.color(ColorCyan, alert.ResourceName)))
	b.WriteString(fmt.Sprintf("  ID:         %s\n", f.color(ColorGray, alert.ResourceID)))

	// Changed Attribute
	b.WriteString(f.color(ColorBold, "\n🔄 Changed Attribute:\n"))
	b.WriteString(fmt.Sprintf("  %s\n", f.color(ColorYellow, alert.Attribute)))

	// Value Change
	b.WriteString(f.color(ColorBold, "\n📝 Value Change:\n"))
	b.WriteString(f.formatValueChange(alert.OldValue, alert.NewValue))

	// User Context
	b.WriteString(f.color(ColorBold, "\n👤 Changed By:\n"))
	b.WriteString(fmt.Sprintf("  User:       %s\n", f.color(ColorPurple, alert.UserIdentity.UserName)))
	b.WriteString(fmt.Sprintf("  Type:       %s\n", alert.UserIdentity.Type))
	if alert.UserIdentity.ARN != "" {
		b.WriteString(fmt.Sprintf("  ARN:        %s\n", f.color(ColorGray, alert.UserIdentity.ARN)))
	}
	b.WriteString(fmt.Sprintf("  Account:    %s\n", alert.UserIdentity.AccountID))

	// Timestamp
	b.WriteString(f.color(ColorBold, "\n⏰ Timestamp:\n"))
	b.WriteString(fmt.Sprintf("  %s\n", alert.Timestamp))

	// Matched Rules
	if len(alert.MatchedRules) > 0 {
		b.WriteString(f.color(ColorBold, "\n📋 Matched Rules:\n"))
		for _, rule := range alert.MatchedRules {
			b.WriteString(fmt.Sprintf("  • %s\n", rule))
		}
	}

	// Terraform Code Reference (if available)
	b.WriteString(f.color(ColorBold, "\n📄 Terraform Code:\n"))
	b.WriteString(f.formatTerraformCode(alert))

	// Recommendations
	b.WriteString(f.color(ColorBold, "\n💡 Recommendations:\n"))
	b.WriteString(f.formatRecommendations(alert))

	b.WriteString(f.color(severityColor, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))

	return b.String()
}

// formatValueChange formats the old -> new value change
func (f *DiffFormatter) formatValueChange(oldValue, newValue interface{}) string {
	var b strings.Builder

	// Handle different types
	oldStr := f.formatValue(oldValue)
	newStr := f.formatValue(newValue)

	// Simple scalar values
	if !f.isComplexType(oldValue) && !f.isComplexType(newValue) {
		b.WriteString(fmt.Sprintf("  %s  →  %s\n",
			f.color(ColorRed, fmt.Sprintf("- %s", oldStr)),
			f.color(ColorGreen, fmt.Sprintf("+ %s", newStr)),
		))
		return b.String()
	}

	// Complex types (maps, slices)
	b.WriteString("  " + f.color(ColorRed, "- Old Value:\n"))
	b.WriteString(f.indentLines(oldStr, 4, ColorRed))
	b.WriteString("\n")
	b.WriteString("  " + f.color(ColorGreen, "+ New Value:\n"))
	b.WriteString(f.indentLines(newStr, 4, ColorGreen))
	b.WriteString("\n")

	return b.String()
}

// formatValue formats a value for display
func (f *DiffFormatter) formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	// Try JSON formatting for complex types
	if f.isComplexType(value) {
		jsonBytes, err := json.MarshalIndent(value, "", "  ")
		if err == nil {
			return string(jsonBytes)
		}
	}

	// Fallback to string representation
	return fmt.Sprintf("%v", value)
}

// isComplexType checks if a value is a complex type (map, slice, struct)
func (f *DiffFormatter) isComplexType(value interface{}) bool {
	if value == nil {
		return false
	}

	v := reflect.ValueOf(value)
	kind := v.Kind()

	return kind == reflect.Map || kind == reflect.Slice || kind == reflect.Struct
}

// formatTerraformCode formats the Terraform code reference
func (f *DiffFormatter) formatTerraformCode(alert *detector.DriftAlert) string {
	var b strings.Builder

	// Example Terraform code showing the current state
	b.WriteString(f.color(ColorGray, "  # Current Terraform Definition:\n"))
	b.WriteString(f.formatTerraformResource(alert, alert.OldValue))

	b.WriteString("\n")
	b.WriteString(f.color(ColorGray, "  # Actual Runtime Configuration:\n"))
	b.WriteString(f.formatTerraformResource(alert, alert.NewValue))

	return b.String()
}

// formatTerraformResource formats a Terraform resource block
func (f *DiffFormatter) formatTerraformResource(alert *detector.DriftAlert, value interface{}) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("  resource \"%s\" \"%s\" {\n",
		alert.ResourceType, alert.ResourceName))
	b.WriteString(fmt.Sprintf("    %s = %s\n",
		alert.Attribute, f.formatTerraformValue(value)))
	b.WriteString("    # ... other attributes ...\n")
	b.WriteString("  }\n")

	return b.String()
}

// formatTerraformValue formats a value in Terraform syntax
func (f *DiffFormatter) formatTerraformValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	case []interface{}:
		var items []string
		for _, item := range v {
			items = append(items, f.formatTerraformValue(item))
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	case map[string]interface{}:
		var pairs []string
		for key, val := range v {
			pairs = append(pairs, fmt.Sprintf("%s = %s", key, f.formatTerraformValue(val)))
		}
		return fmt.Sprintf("{\n    %s\n  }", strings.Join(pairs, "\n    "))
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatRecommendations formats recommended actions
func (f *DiffFormatter) formatRecommendations(alert *detector.DriftAlert) string {
	var b strings.Builder

	b.WriteString("  1. Review the change with the user who made it\n")
	b.WriteString("  2. Determine if the change is authorized\n")
	b.WriteString("  3. Update Terraform code if the change is intentional:\n")
	b.WriteString(f.color(ColorCyan, "     terraform plan && terraform apply\n"))
	b.WriteString("  4. Or revert the manual change to match IaC:\n")
	b.WriteString(f.color(ColorCyan, fmt.Sprintf("     terraform apply -target=%s.%s\n",
		alert.ResourceType, alert.ResourceName)))

	return b.String()
}

// FormatUnifiedDiff formats the drift as a unified diff (Git-style)
func (f *DiffFormatter) FormatUnifiedDiff(alert *detector.DriftAlert) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("--- terraform/%s.%s\t(Terraform State)\n",
		alert.ResourceType, alert.ResourceName))
	b.WriteString(fmt.Sprintf("+++ runtime/%s.%s\t(Actual Configuration)\n",
		alert.ResourceType, alert.ResourceName))
	b.WriteString("@@ -1,1 +1,1 @@\n")

	oldStr := f.formatValue(alert.OldValue)
	newStr := f.formatValue(alert.NewValue)

	// Format as unified diff
	for _, line := range strings.Split(oldStr, "\n") {
		b.WriteString(f.color(ColorRed, fmt.Sprintf("-%s\n", line)))
	}
	for _, line := range strings.Split(newStr, "\n") {
		b.WriteString(f.color(ColorGreen, fmt.Sprintf("+%s\n", line)))
	}

	return b.String()
}

// FormatMarkdown formats the drift for Markdown (GitHub, Slack, etc.)
func (f *DiffFormatter) FormatMarkdown(alert *detector.DriftAlert) string {
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

// FormatJSON formats the drift as JSON
func (f *DiffFormatter) FormatJSON(alert *detector.DriftAlert) (string, error) {
	// Create a structured diff object
	diff := map[string]interface{}{
		"severity":      alert.Severity,
		"resource_type": alert.ResourceType,
		"resource_name": alert.ResourceName,
		"resource_id":   alert.ResourceID,
		"attribute":     alert.Attribute,
		"change": map[string]interface{}{
			"old_value": alert.OldValue,
			"new_value": alert.NewValue,
		},
		"user": map[string]string{
			"name":        alert.UserIdentity.UserName,
			"type":        alert.UserIdentity.Type,
			"arn":         alert.UserIdentity.ARN,
			"account_id":  alert.UserIdentity.AccountID,
			"principal_id": alert.UserIdentity.PrincipalID,
		},
		"timestamp":     alert.Timestamp,
		"matched_rules": alert.MatchedRules,
		"terraform_code": map[string]string{
			"state_definition": f.formatTerraformResource(alert, alert.OldValue),
			"actual_config":    f.formatTerraformResource(alert, alert.NewValue),
		},
	}

	jsonBytes, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// color applies ANSI color codes if color is enabled
func (f *DiffFormatter) color(colorCode, text string) string {
	if !f.colorEnabled {
		return text
	}
	return colorCode + text + ColorReset
}

// getSeverityColor returns the appropriate color for a severity level
func (f *DiffFormatter) getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return ColorRed
	case "high":
		return ColorYellow
	case "medium":
		return ColorBlue
	case "low":
		return ColorGreen
	default:
		return ColorGray
	}
}

// indentLines indents all lines of a string
func (f *DiffFormatter) indentLines(text string, spaces int, color string) string {
	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	result := strings.Join(lines, "\n")
	return f.color(color, result)
}

// FormatSideBySide formats the drift as side-by-side comparison
func (f *DiffFormatter) FormatSideBySide(alert *detector.DriftAlert) string {
	var b strings.Builder

	oldStr := f.formatValue(alert.OldValue)
	newStr := f.formatValue(alert.NewValue)

	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	// Header
	b.WriteString(fmt.Sprintf("%-40s | %s\n", "Terraform State", "Actual Configuration"))
	b.WriteString(strings.Repeat("─", 40) + "─┼─" + strings.Repeat("─", 40) + "\n")

	// Side-by-side comparison
	for i := 0; i < maxLines; i++ {
		oldLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}

		newLine := ""
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			b.WriteString(fmt.Sprintf("  %-38s │   %-38s\n", oldLine, newLine))
		} else {
			b.WriteString(fmt.Sprintf("%s │ %s\n",
				f.color(ColorRed, fmt.Sprintf("- %-38s", oldLine)),
				f.color(ColorGreen, fmt.Sprintf("+ %-38s", newLine)),
			))
		}
	}

	return b.String()
}
