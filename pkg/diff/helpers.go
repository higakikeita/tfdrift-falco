package diff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

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
func (f *DiffFormatter) formatTerraformCode(alert *types.DriftAlert) string {
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
func (f *DiffFormatter) formatTerraformResource(alert *types.DriftAlert, value interface{}) string {
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
func (f *DiffFormatter) formatRecommendations(alert *types.DriftAlert) string {
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
