package detector

import (
	"context"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// handleEvent processes a single event
func (d *Detector) handleEvent(event types.Event) {
	log.Debugf("Processing event: %s - %s", event.EventName, event.ResourceID)

	// Look up resource in Terraform state
	resource, exists := d.stateManager.GetResource(event.ResourceID)
	if !exists {
		log.Warnf("Resource %s not found in Terraform state (unmanaged resource)", event.ResourceID)
		// Send alert for unmanaged resource
		d.sendUnmanagedResourceAlert(&event)

		// Handle auto-import if enabled
		if d.cfg.AutoImport.Enabled && d.importer != nil && d.approvalManager != nil {
			d.handleAutoImport(context.Background(), &event)
		}
		return
	}

	// Compare with state
	drifts := d.detectDrifts(resource, event.Changes)
	if len(drifts) == 0 {
		log.Debugf("No drift detected for %s", event.ResourceID)
		return
	}

	// Evaluate rules
	for _, drift := range drifts {
		matchedRules := d.evaluateRules(resource.Type, drift.Attribute)
		if len(matchedRules) == 0 {
			continue
		}

		// Extract timestamp safely
		timestamp := ""
		if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
			if eventTime, ok := rawEvent["eventTime"].(string); ok {
				timestamp = eventTime
			}
		}

		alert := &types.DriftAlert{
			Severity:     d.getSeverity(matchedRules),
			ResourceType: resource.Type,
			ResourceName: resource.Name,
			ResourceID:   event.ResourceID,
			Attribute:    drift.Attribute,
			OldValue:     drift.OldValue,
			NewValue:     drift.NewValue,
			UserIdentity: event.UserIdentity,
			MatchedRules: matchedRules,
			Timestamp:    timestamp,
			AlertType:    "drift", // Mark as drift alert
		}

		d.sendAlert(alert)
	}
}
