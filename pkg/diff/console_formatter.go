// Package diff provides diff formatting utilities for drift alerts.
package diff

import (
	"fmt"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatConsole formats the drift for console output with colors.
func (f *Formatter) FormatConsole(alert *types.DriftAlert) string {
	var b strings.Builder

	// Header
	severityColor := f.getSeverityColor(alert.Severity)
	b.WriteString(f.color(severityColor, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	b.WriteString(f.color(ColorBold, fmt.Sprintf("🚨 DRIFT DETECTED: %s.%s\n", alert.ResourceType, alert.ResourceName)))
	b.WriteString(f.color(severityColor, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))

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

	// User Context - WHO made the change
	b.WriteString(f.color(ColorBold+ColorYellow, "\n👤 WHO Changed It:\n"))
	b.WriteString(fmt.Sprintf("  User:       %s\n", f.color(ColorPurple+ColorBold, alert.UserIdentity.UserName)))
	b.WriteString(fmt.Sprintf("  Type:       %s\n", alert.UserIdentity.Type))
	if alert.UserIdentity.ARN != "" {
		b.WriteString(fmt.Sprintf("  ARN:        %s\n", f.color(ColorGray, alert.UserIdentity.ARN)))
	}
	if alert.UserIdentity.PrincipalID != "" {
		b.WriteString(fmt.Sprintf("  Principal:  %s\n", f.color(ColorGray, alert.UserIdentity.PrincipalID)))
	}
	b.WriteString(fmt.Sprintf("  Account:    %s\n", alert.UserIdentity.AccountID))

	// Timestamp - WHEN the change happened
	b.WriteString(f.color(ColorBold+ColorYellow, "\n⏰ WHEN It Changed:\n"))
	b.WriteString(fmt.Sprintf("  %s\n", f.color(ColorCyan, alert.Timestamp)))

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

// FormatUnmanagedResource formats an unmanaged resource alert for console output.
func (f *Formatter) FormatUnmanagedResource(alert *types.UnmanagedResourceAlert) string {
	var b strings.Builder

	// Header
	severityColor := f.getSeverityColor(alert.Severity)
	b.WriteString(f.color(severityColor, "\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	b.WriteString(f.color(ColorBold, "⚠️  UNMANAGED RESOURCE DETECTED\n"))
	b.WriteString(f.color(severityColor, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))

	// Severity
	b.WriteString(f.color(ColorBold, "\n📊 Severity: "))
	b.WriteString(f.color(severityColor, strings.ToUpper(alert.Severity)))
	b.WriteString("\n")

	// Resource info
	b.WriteString(f.color(ColorBold, "\n📦 Resource:\n"))
	b.WriteString(fmt.Sprintf("   Type: %s\n", f.color(ColorCyan, alert.ResourceType)))
	b.WriteString(fmt.Sprintf("   ID:   %s\n", f.color(ColorYellow, alert.ResourceID)))

	// Event
	b.WriteString(f.color(ColorBold, "\n🔔 Event: "))
	b.WriteString(alert.EventName)
	b.WriteString("\n")

	// Timestamp
	b.WriteString(f.color(ColorBold, "\n🕐 When: "))
	b.WriteString(f.color(ColorGray, alert.Timestamp))
	b.WriteString("\n")

	// User identity - WHO made the change
	b.WriteString(f.color(ColorBold, "\n👤 Who Changed It:\n"))
	b.WriteString(fmt.Sprintf("   User:     %s\n", f.color(ColorPurple, alert.UserIdentity.UserName)))
	if alert.UserIdentity.ARN != "" {
		b.WriteString(fmt.Sprintf("   ARN:      %s\n", alert.UserIdentity.ARN))
	}
	if alert.UserIdentity.PrincipalID != "" {
		b.WriteString(fmt.Sprintf("   Principal: %s\n", alert.UserIdentity.PrincipalID))
	}

	// Reason
	b.WriteString(f.color(ColorBold, "\n⚠️  Reason:\n"))
	b.WriteString(f.color(ColorYellow, fmt.Sprintf("   %s\n", alert.Reason)))

	// Changes
	if len(alert.Changes) > 0 {
		b.WriteString(f.color(ColorBold, "\n🔄 Changes Made:\n"))
		for key, value := range alert.Changes {
			b.WriteString(fmt.Sprintf("   %s: %v\n", key, value))
		}
	}

	// Recommendation
	b.WriteString(f.color(ColorBold, "\n💡 Recommendation:\n"))
	b.WriteString(f.color(ColorYellow, "   This resource is not managed by Terraform.\n"))
	b.WriteString(f.color(ColorYellow, "   Consider importing it:\n"))
	b.WriteString(f.color(ColorGray, fmt.Sprintf("   terraform import %s.resource_name %s\n",
		alert.ResourceType, alert.ResourceID)))

	b.WriteString(f.color(severityColor, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n"))

	return b.String()
}
