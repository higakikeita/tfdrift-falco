package detector

import (
	"context"
	"fmt"
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/policy"
	"github.com/keitahigaki/tfdrift-falco/pkg/telemetry"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
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
		// A mutating event reached a *managed* resource but we couldn't extract
		// any attribute-level change (there's no change_extractor case for this
		// event yet). Reporting "no drift" here silently loses a real
		// out-of-band change — the core of the detection-ceiling bug (#324):
		// the rule now fires on 246 events but only ~41 have an extractor.
		// Surface a coarse "resource modified" drift instead of dropping it.
		// (Read-only events are excluded so a stray Describe/List can't
		// false-positive; in the live pipeline only relevant/mutating events
		// reach here anyway.)
		if len(event.Changes) == 0 && isMutatingEvent(event.EventName) {
			d.sendCoarseDriftAlert(ctx, resource, &event)
			telemetry.SetOK(span)
			return
		}
		log.Debugf("No drift detected for %s", event.ResourceID)
		telemetry.SetOK(span)
		return
	}

	span.AddEvent("drift_detected", trace.WithAttributes(
		telemetry.AttrDriftCount.Int(len(drifts)),
	))

	// Evaluate rules
	for _, drift := range drifts {
		// A detected change must never be silently dropped just because the
		// user did not configure a matching drift_rule. drift_rules only
		// classify severity; absence of a rule means "unclassified", not
		// "ignore". Previously an unmatched drift hit `continue` and vanished.
		matchedRules := d.evaluateRules(resource.Type, drift.Attribute)
		severity := "medium" // default for an unclassified but real change
		if len(matchedRules) > 0 {
			severity = d.getSeverity(matchedRules)
		}

		// Extract timestamp safely
		timestamp := ""
		if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
			if eventTime, ok := rawEvent["eventTime"].(string); ok {
				timestamp = eventTime
			}
		}

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

// readOnlyEventPrefixes are CloudTrail verb prefixes that never mutate state.
// Used as a defensive guard so a coarse drift alert is only raised for events
// that actually change something.
var readOnlyEventPrefixes = []string{"Describe", "List", "Get", "BatchGet", "Lookup", "Search", "Head", "Query", "Scan"}

// isMutatingEvent reports whether an event name denotes a state-changing API
// call (as opposed to a read). The live pipeline only forwards relevant
// (mutating) events, but handleEvent is also exercised directly, so this keeps
// coarse alerting honest regardless of caller.
func isMutatingEvent(eventName string) bool {
	if eventName == "" {
		return false
	}
	for _, p := range readOnlyEventPrefixes {
		if strings.HasPrefix(eventName, p) {
			return false
		}
	}
	return true
}

// sendCoarseDriftAlert emits a drift alert for a managed resource that was
// changed by a mutating event we can't yet diff at the attribute level. It is
// the honest fallback for the detection ceiling: "this managed resource was
// modified out-of-band by <event>" is the drift signal; attribute-level detail
// is enrichment that change_extractor adds for events it understands.
func (d *Detector) sendCoarseDriftAlert(ctx context.Context, resource *terraform.Resource, event *types.Event) {
	timestamp := ""
	if rawEvent, ok := event.RawEvent.(map[string]interface{}); ok {
		if eventTime, ok := rawEvent["eventTime"].(string); ok {
			timestamp = eventTime
		}
	}

	alert := &types.DriftAlert{
		Severity:     "medium",
		ResourceType: resource.Type,
		ResourceName: resource.Name,
		ResourceID:   event.ResourceID,
		Attribute:    "(resource modified out-of-band)",
		OldValue:     nil,
		NewValue:     event.EventName,
		UserIdentity: event.UserIdentity,
		Timestamp:    timestamp,
		AlertType:    "drift",
	}

	// Respect policy allow decisions so this path stays consistent with the
	// attribute-level path (evaluatePolicy is nil-safe).
	if policyResult := d.evaluatePolicy(ctx, alert); policyResult != nil {
		if policyResult.Severity != "" {
			alert.Severity = policyResult.Severity
		}
		if policyResult.Decision == policy.DecisionAllow {
			log.Debugf("Policy allows coarse drift on %s, skipping alert", alert.ResourceID)
			return
		}
	}

	d.sendAlert(alert)
}
