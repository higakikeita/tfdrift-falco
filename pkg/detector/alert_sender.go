// Package detector implements drift detection logic for TFDrift-Falco.
package detector

import (
	"fmt"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
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

	// Broadcast to WebSocket clients
	if d.broadcaster != nil {
		d.broadcaster.Broadcast(broadcaster.Event{
			Type:      "drift",
			Timestamp: time.Now().Format(time.RFC3339),
			Payload: map[string]interface{}{
				"severity":      alert.Severity,
				"resource_type": alert.ResourceType,
				"resource_name": alert.ResourceName,
				"resource_id":   alert.ResourceID,
				"attribute":     alert.Attribute,
				"old_value":     alert.OldValue,
				"new_value":     alert.NewValue,
				"user_identity": alert.UserIdentity,
				"matched_rules": alert.MatchedRules,
				"timestamp":     alert.Timestamp,
			},
		})
	}

	// Add to graph store for visualization
	if d.graphStore != nil {
		d.graphStore.AddDrift(*alert)
		log.Debug("Added drift alert to graph store")
	}

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

	// Broadcast to WebSocket clients
	if d.broadcaster != nil {
		d.broadcaster.Broadcast(broadcaster.Event{
			Type:      "unmanaged",
			Timestamp: time.Now().Format(time.RFC3339),
			Payload: map[string]interface{}{
				"severity":      alert.Severity,
				"resource_type": alert.ResourceType,
				"resource_id":   alert.ResourceID,
				"event_name":    alert.EventName,
				"user_identity": alert.UserIdentity,
				"changes":       alert.Changes,
				"reason":        alert.Reason,
				"timestamp":     alert.Timestamp,
			},
		})
	}

	// Add to graph store for visualization
	if d.graphStore != nil {
		d.graphStore.AddUnmanaged(*alert)
		log.Debug("Added unmanaged resource alert to graph store")
	}

	if d.cfg.DryRun {
		log.Info("[DRY-RUN] Unmanaged resource alert notification skipped")
		fmt.Println("\n=== Markdown Format (for Slack) ===")
		fmt.Println(d.formatter.FormatUnmanagedResourceMarkdown(alert))
		return
	}

	// TODO: Send to notification channels
	// For now, console output is enough
}
