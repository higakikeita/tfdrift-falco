package detector

import (
	"context"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/policy"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// evaluatePolicy runs the policy engine against a drift alert and returns
// the policy decision. If the policy engine is not configured, returns nil
// (meaning no policy override).
func (d *Detector) evaluatePolicy(ctx context.Context, alert *types.DriftAlert) *policy.EvalResult {
	if d.policyEngine == nil {
		return nil
	}

	input := &policy.DriftInput{
		Type:         alert.AlertType,
		ResourceType: alert.ResourceType,
		ResourceID:   alert.ResourceID,
		ResourceName: alert.ResourceName,
		Attribute:    alert.Attribute,
		OldValue:     alert.OldValue,
		NewValue:     alert.NewValue,
		Severity:     alert.Severity,
		Timestamp:    alert.Timestamp,
		UserIdentity: policy.UserInput{
			Type:        alert.UserIdentity.Type,
			PrincipalID: alert.UserIdentity.PrincipalID,
			ARN:         alert.UserIdentity.ARN,
			AccountID:   alert.UserIdentity.AccountID,
			UserName:    alert.UserIdentity.UserName,
		},
	}

	result, err := d.policyEngine.Evaluate(ctx, input)
	if err != nil {
		log.WithError(err).Warn("Policy evaluation failed, falling through to default alert")
		return nil
	}

	log.WithFields(log.Fields{
		"resource_id":   alert.ResourceID,
		"resource_type": alert.ResourceType,
		"decision":      result.Decision,
		"reason":        result.Reason,
	}).Debug("Policy evaluation result")

	// Broadcast policy decision
	if d.broadcaster != nil {
		d.broadcaster.Broadcast(broadcaster.Event{
			Type: "policy_decision",
			Payload: map[string]interface{}{
				"resource_type": alert.ResourceType,
				"resource_id":   alert.ResourceID,
				"attribute":     alert.Attribute,
				"decision":      string(result.Decision),
				"reason":        result.Reason,
				"severity":      result.Severity,
				"labels":        result.Labels,
			},
		})
	}

	return result
}

// evaluateUnmanagedPolicy runs the policy engine for an unmanaged resource event.
func (d *Detector) evaluateUnmanagedPolicy(ctx context.Context, event *types.Event) *policy.EvalResult {
	if d.policyEngine == nil {
		return nil
	}

	input := &policy.DriftInput{
		Type:         "unmanaged",
		Provider:     event.Provider,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		Severity:     "medium",
		Changes:      event.Changes,
		UserIdentity: policy.UserInput{
			Type:        event.UserIdentity.Type,
			PrincipalID: event.UserIdentity.PrincipalID,
			ARN:         event.UserIdentity.ARN,
			AccountID:   event.UserIdentity.AccountID,
			UserName:    event.UserIdentity.UserName,
		},
	}

	result, err := d.policyEngine.Evaluate(ctx, input)
	if err != nil {
		log.WithError(err).Warn("Policy evaluation failed for unmanaged resource")
		return nil
	}

	log.WithFields(log.Fields{
		"resource_id":   event.ResourceID,
		"resource_type": event.ResourceType,
		"decision":      result.Decision,
		"reason":        result.Reason,
	}).Debug("Unmanaged resource policy evaluation result")

	return result
}
