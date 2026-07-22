package detector

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/require"
)

// spyNotifier records every alert the detector actually delivers, so tests can
// assert on count and fields instead of the old assert.True(true) style that
// stayed green even if the tool alerted on nothing (pus #1, #327).
type spyNotifier struct{ sent []*types.DriftAlert }

func (s *spyNotifier) Send(a *types.DriftAlert) error {
	s.sent = append(s.sent, a)
	return nil
}

// newTestDetector wires a real StateManager (loaded from a temp tfstate with a
// single aws_instance keyed by attributes.id) to a spy notifier.
func newTestDetector(t *testing.T, rules []config.DriftRule, attrs map[string]interface{}) (*Detector, *spyNotifier) {
	t.Helper()

	state := map[string]interface{}{
		"version":           4,
		"terraform_version": "1.5.0",
		"resources": []map[string]interface{}{{
			"mode":      "managed",
			"type":      "aws_instance",
			"name":      "web",
			"provider":  `provider["registry.terraform.io/hashicorp/aws"]`,
			"instances": []map[string]interface{}{{"attributes": attrs}},
		}},
	}
	data, err := json.Marshal(state)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "terraform.tfstate")
	require.NoError(t, os.WriteFile(path, data, 0o600))

	sm, err := terraform.NewStateManager(config.TerraformStateConfig{Backend: "local", LocalPath: path})
	require.NoError(t, err)
	require.NoError(t, sm.Load(context.Background()))

	spy := &spyNotifier{}
	d := &Detector{
		cfg:          &config.Config{DriftRules: rules},
		stateManager: sm,
		formatter:    diff.NewFormatter(false),
		notifier:     spy,
	}
	return d, spy
}

func modifyEvent(resourceID string, changes map[string]interface{}) types.Event {
	return types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   resourceID,
		UserIdentity: types.UserIdentity{UserName: "alice", Type: "IAMUser"},
		Changes:      changes,
		RawEvent:     map[string]interface{}{"eventTime": "2026-07-22T00:00:00Z"},
	}
}

func TestHandleEvent_DeliversAlertOnRealChange(t *testing.T) {
	rules := []config.DriftRule{{
		Name:              "EC2 Instance Type Change",
		ResourceTypes:     []string{"aws_instance"},
		WatchedAttributes: []string{"instance_type"},
		Severity:          "high",
	}}
	d, spy := newTestDetector(t, rules, map[string]interface{}{"id": "i-123", "instance_type": "t2.micro"})

	d.handleEvent(modifyEvent("i-123", map[string]interface{}{"instance_type": "t2.large"}))

	require.Len(t, spy.sent, 1, "a real change with a matching rule must produce exactly one alert")
	a := spy.sent[0]
	require.Equal(t, "high", a.Severity)
	require.Equal(t, "instance_type", a.Attribute)
	require.Equal(t, "t2.micro", a.OldValue)
	require.Equal(t, "t2.large", a.NewValue)
	require.Equal(t, "alice", a.UserIdentity.UserName)
	require.Equal(t, "i-123", a.ResourceID)
}

func TestHandleEvent_UnmatchedRuleStillAlerts(t *testing.T) {
	// Silent-drop fix (#327): a real change whose attribute matches NO
	// configured drift_rule must still be surfaced (default severity), not
	// silently `continue`d away.
	d, spy := newTestDetector(t, nil, map[string]interface{}{"id": "i-123", "instance_type": "t2.micro"})

	d.handleEvent(modifyEvent("i-123", map[string]interface{}{"disable_api_termination": true}))

	require.Len(t, spy.sent, 1, "a detected change must alert even with no matching drift_rule")
	require.Equal(t, "medium", spy.sent[0].Severity, "unclassified change defaults to medium, not dropped")
	require.Equal(t, "disable_api_termination", spy.sent[0].Attribute)
}

func TestHandleEvent_NoChangeNoAlert(t *testing.T) {
	// False-positive guard: an event whose value equals current state is not drift.
	rules := []config.DriftRule{{
		Name: "EC2 Instance Type Change", ResourceTypes: []string{"aws_instance"},
		WatchedAttributes: []string{"instance_type"}, Severity: "high",
	}}
	d, spy := newTestDetector(t, rules, map[string]interface{}{"id": "i-123", "instance_type": "t2.micro"})

	d.handleEvent(modifyEvent("i-123", map[string]interface{}{"instance_type": "t2.micro"}))

	require.Empty(t, spy.sent, "no attribute changed, so no alert should be sent")
}

func TestHandleEvent_UnmanagedResourceAlerts(t *testing.T) {
	d, spy := newTestDetector(t, nil, map[string]interface{}{"id": "i-123", "instance_type": "t2.micro"})

	d.handleEvent(modifyEvent("i-999", map[string]interface{}{"instance_type": "t2.large"}))

	require.Len(t, spy.sent, 1, "an event for a resource absent from state is an unmanaged-resource alert")
}

func TestDetectDrifts_ScalarAndSliceComparison(t *testing.T) {
	d := &Detector{}
	res := &terraform.Resource{
		Type: "aws_security_group",
		Attributes: map[string]interface{}{
			"instance_type": "t2.micro",
			"ingress":       []interface{}{"22", "443"},
		},
	}

	// Identical scalar -> no drift.
	require.Empty(t, d.detectDrifts(res, map[string]interface{}{"instance_type": "t2.micro"}))
	// Different scalar -> drift.
	require.Len(t, d.detectDrifts(res, map[string]interface{}{"instance_type": "t2.large"}), 1)
	// Equal slices -> no drift (DeepEqual, not ==).
	require.Empty(t, d.detectDrifts(res, map[string]interface{}{"ingress": []interface{}{"22", "443"}}))
	// Different slices -> drift, and crucially NO panic (reflect.DeepEqual).
	require.NotPanics(t, func() {
		got := d.detectDrifts(res, map[string]interface{}{"ingress": []interface{}{"22", "443", "3306"}})
		require.Len(t, got, 1)
	})
}
