// Package detector implements drift detection logic for TFDrift-Falco.
package detector

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// sendAlert sends a drift alert
func (d *Detector) sendAlert(alert *types.DriftAlert) {
	// Format and display the drift in console
	consoleDiff := d.formatter.FormatConsole(alert)
	fmt.Println(consoleDiff)

	// Also log in traditional format
	log.Warnf("DRIFT DETECTED: %s.%s - %s: %v → %v",
		alert.ResourceType, alert.ResourceName, alert.Attribute,
		alert.OldValue, alert.NewValue)

	if d.cfg.DryRun {
		log.Info("[DRY-RUN] Alert notification skipped")

		// In dry-run, also show other formats as examples
		fmt.Println("\n=== Unified Diff Format ===")
		fmt.Println(d.formatter.FormatUnifiedDiff(alert))

		fmt.Println("\n=== Side-by-Side Format ===")
		fmt.Println(d.formatter.FormatSideBySide(alert))

		fmt.Println("\n=== Markdown Format (for Slack/GitHub) ===")
		fmt.Println(d.formatter.FormatMarkdown(alert))

		return
	}

	if err := d.notifier.Send(alert); err != nil {
		log.Errorf("Failed to send alert: %v", err)
	}
}

// sendUnmanagedResourceAlert sends an alert for unmanaged resources
func (d *Detector) sendUnmanagedResourceAlert(event *types.Event) {
	// Extract timestamp from raw event
	timestamp := ""
	if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
		if eventTime, ok := rawEvent["eventTime"].(string); ok {
			timestamp = eventTime
		}
	}

	alert := &types.UnmanagedResourceAlert{
		Severity:     "warning", // Default severity for unmanaged resources
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		EventName:    event.EventName,
		UserIdentity: event.UserIdentity,
		Changes:      event.Changes,
		Timestamp:    timestamp,
		Reason:       fmt.Sprintf("Resource %s (%s) is not found in Terraform state", event.ResourceID, event.ResourceType),
	}

	// Format and display
	consoleOutput := d.formatter.FormatUnmanagedResource(alert)
	fmt.Println(consoleOutput)

	// Also log
	log.Warnf("UNMANAGED RESOURCE: %s (%s) - Event: %s by %s",
		alert.ResourceID, alert.ResourceType, alert.EventName, alert.UserIdentity.UserName)

	if d.cfg.DryRun {
		log.Info("[DRY-RUN] Unmanaged resource alert notification skipped")
		fmt.Println("\n=== Markdown Format (for Slack) ===")
		fmt.Println(d.formatter.FormatUnmanagedResourceMarkdown(alert))
		return
	}

	// TODO: Send to notification channels
	// For now, console output is enough
}
