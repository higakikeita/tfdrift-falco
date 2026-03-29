package detector

import (
	"context"
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/policy"
	"github.com/keitahigaki/tfdrift-falco/pkg/telemetry"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// handleEvent processes a single event
func (d *Detector) handleEvent(event types.Event) {
	ctx := context.Background()
	ctx, span := telemetry.StartSpan(ctx, "detector.handle_event",
		trace.WithAttributes(
			telemetry.AttrProvider.String(event.Provider),
			telemetry.AttrEventName.String(event.EventName),
			telemetry.AttrResourceType.String(event.ResourceType),
			telemetry.AttrResourceID.String(event.ResourceID),
		),
	)
	defer span.End()

	log.Debugf("Processing event: %s - %s", event.EventName, event.ResourceID)

	// Look up resource in Terraform state
	resource, exists := d.stateManager.GetResource(event.ResourceID)
	if !exists {
		span.AddEvent("unmanaged_resource", trace.WithAttributes(
			attribute.String("resource_id", event.ResourceID),
		))
		log.Warnf("Resource %s not found in Terraform state (unmanaged resource)", event.ResourceID)

		// Evaluate policy for unmanaged resource
		policyResult := d.evaluateUnmanagedPolicy(ctx, &event)
		if policyResult != nil && policyResult.Decision == policy.DecisionAllow {
			log.Debugf("Policy allows unmanaged resource %s, skipping alert", event.ResourceID)
			telemetry.SetOK(span)
			return
		}

		// Send alert for unmanaged resource
		d.sendUnmanagedResourceAlert(&event)

		// Handle auto-import if enabled (or if policy says remediate)
		if d.cfg.AutoImport.Enabled && d.importer != nil && d.approvalManager != nil {
			d.handleAutoImport(ctx, &event)
		}

		// Generate remediation proposal for unmanaged resource
		if policyResult != nil && policyResult.Decision == policy.DecisionRemediate {
			d.handleUnmanagedRemediation(ctx, &event)
		} else if d.cfg.Remediation.Enabled {
			d.handleUnmanagedRemediation(ctx, &event)
		}
		return
	}

	// Compare with state
	_, detectSpan := telemetry.StartSpan(ctx, "detector.detect_drifts",
		trace.WithAttributes(
			telemetry.AttrResourceType.String(resource.Type),
			telemetry.AttrResourceID.String(event.ResourceID),
		),
	)
	drifts := d.detectDrifts(resource, event.Changes)
	detectSpan.SetAttributes(attribute.Int("drift_count", len(drifts)))
	detectSpan.End()

	if len(drifts) == 0 {
		log.Debugf("No drift detected for %s", event.ResourceID)
		telemetry.SetOK(span)
		return
	}

	span.AddEvent("drift_detected", trace.WithAttributes(
		telemetry.AttrDriftCount.Int(len(drifts)),
	))

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

		severity := d.getSeverity(matchedRules)
		alert := &types.DriftAlert{
			Severity:     severity,
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

		// Evaluate policy before alerting
		policyResult := d.evaluatePolicy(ctx, alert)
		if policyResult != nil {
			// Override severity if policy says so
			if policyResult.Severity != "" {
				alert.Severity = policyResult.Severity
				severity = policyResult.Severity
			}

			switch policyResult.Decision {
			case policy.DecisionAllow:
				log.Debugf("Policy allows drift on %s.%s, skipping alert", alert.ResourceType, alert.Attribute)
				span.AddEvent("policy_allow", trace.WithAttributes(
					attribute.String("reason", policyResult.Reason),
				))
				continue
			case policy.DecisionDeny:
				log.Warnf("Policy DENY: %s — %s", alert.ResourceID, policyResult.Reason)
				span.AddEvent("policy_deny", trace.WithAttributes(
					attribute.String("reason", policyResult.Reason),
				))
			}
		}

		span.AddEvent("alert_sent", trace.WithAttributes(
			telemetry.AttrSeverity.String(severity),
			attribute.String("attribute", fmt.Sprintf("%v", drift.Attribute)),
		))

		d.sendAlert(alert)

		// Generate remediation proposal for drift (if policy says remediate, or if remediation is enabled)
		if policyResult != nil && policyResult.Decision == policy.DecisionRemediate {
			d.handleRemediation(ctx, alert)
		} else {
			d.handleRemediation(ctx, alert)
		}
	}
}
