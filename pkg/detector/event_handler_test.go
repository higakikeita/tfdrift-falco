package detector

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandleEvent_TableDriven covers event handling with various scenarios
func TestHandleEvent_TableDriven(t *testing.T) {
	// Set up state manager with test data
	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)
	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	tests := []struct {
		name        string
		rules       []config.DriftRule
		event       types.Event
		expectPanic bool
	}{
		{
			name: "drift_detected_with_matching_rule",
			rules: []config.DriftRule{
				{
					Name:              "instance-type-change",
					ResourceTypes:     []string{"aws_instance"},
					WatchedAttributes: []string{"instance_type"},
					Severity:          "high",
				},
			},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "ModifyInstanceAttribute",
				Changes:      map[string]interface{}{"instance_type": "t3.large"},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
			},
			expectPanic: false,
		},
		{
			name:  "resource_not_found_unmanaged",
			rules: []config.DriftRule{},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-nonexistent",
				EventName:    "RunInstances",
				Changes:      map[string]interface{}{"instance_type": "t3.micro"},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
			},
			expectPanic: false,
		},
		{
			name:  "no_changes_in_event",
			rules: []config.DriftRule{},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "DescribeInstances",
				Changes:      map[string]interface{}{},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "readonly"},
			},
			expectPanic: false,
		},
		{
			name:  "nil_changes_map",
			rules: []config.DriftRule{},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "DescribeInstances",
				Changes:      nil,
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "readonly"},
			},
			expectPanic: false,
		},
		{
			name: "drift_with_no_matching_rules",
			rules: []config.DriftRule{
				{
					Name:              "sg-ingress",
					ResourceTypes:     []string{"aws_security_group"},
					WatchedAttributes: []string{"ingress"},
					Severity:          "critical",
				},
			},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "ModifyInstanceAttribute",
				Changes:      map[string]interface{}{"instance_type": "t3.large"},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
			},
			expectPanic: false,
		},
		{
			name: "event_with_raw_event_timestamp",
			rules: []config.DriftRule{
				{
					Name:              "instance-type-change",
					ResourceTypes:     []string{"aws_instance"},
					WatchedAttributes: []string{"instance_type"},
					Severity:          "medium",
				},
			},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "ModifyInstanceAttribute",
				Changes:      map[string]interface{}{"instance_type": "t3.xlarge"},
				UserIdentity: types.UserIdentity{Type: "AssumedRole", UserName: "deploy-role"},
				RawEvent: map[string]interface{}{
					"eventTime":   "2026-03-21T10:00:00Z",
					"eventSource": "ec2.amazonaws.com",
				},
			},
			expectPanic: false,
		},
		{
			name: "event_with_invalid_raw_event_type",
			rules: []config.DriftRule{
				{
					Name:              "instance-type-change",
					ResourceTypes:     []string{"aws_instance"},
					WatchedAttributes: []string{"instance_type"},
					Severity:          "high",
				},
			},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "ModifyInstanceAttribute",
				Changes:      map[string]interface{}{"instance_type": "t3.2xlarge"},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
				RawEvent:     "not-a-map", // Invalid type for timestamp extraction
			},
			expectPanic: false,
		},
		{
			name:  "empty_resource_id",
			rules: []config.DriftRule{},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "",
				EventName:    "RunInstances",
				Changes:      map[string]interface{}{"instance_type": "t3.micro"},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
			},
			expectPanic: false,
		},
		{
			name:  "empty_user_identity",
			rules: []config.DriftRule{},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-nonexistent",
				EventName:    "RunInstances",
				Changes:      map[string]interface{}{"instance_type": "t3.micro"},
				UserIdentity: types.UserIdentity{},
			},
			expectPanic: false,
		},
		{
			name: "multiple_changes_with_mixed_rules",
			rules: []config.DriftRule{
				{
					Name:              "instance-type-change",
					ResourceTypes:     []string{"aws_instance"},
					WatchedAttributes: []string{"instance_type"},
					Severity:          "high",
				},
				{
					Name:              "tag-change",
					ResourceTypes:     []string{"aws_instance"},
					WatchedAttributes: []string{"tags"},
					Severity:          "low",
				},
			},
			event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-0cea65ac652556767",
				EventName:    "ModifyInstanceAttribute",
				Changes: map[string]interface{}{
					"instance_type": "t3.large",
					"tags":          "Name=updated",
					"unmatched_key": "some_value",
				},
				UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
			},
			expectPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				DriftRules: tt.rules,
				DryRun:     true,
				AutoImport: config.AutoImportConfig{Enabled: false},
			}
			formatter := diff.NewFormatter(false)

			detector := &Detector{
				cfg:          cfg,
				stateManager: stateManager,
				formatter:    formatter,
			}

			assert.NotPanics(t, func() {
				detector.handleEvent(tt.event)
			})
		})
	}
}

// TestHandleEvent_EmptyState tests event handling when state has no resources
func TestHandleEvent_EmptyState(t *testing.T) {
	// Create state manager without loading state (empty state)
	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)
	// Note: intentionally NOT calling stateManager.Load() to test empty state

	cfg := &config.Config{
		DriftRules: []config.DriftRule{
			{
				Name:              "instance-type-change",
				ResourceTypes:     []string{"aws_instance"},
				WatchedAttributes: []string{"instance_type"},
				Severity:          "high",
			},
		},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{Enabled: false},
	}
	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	// All events should be treated as unmanaged when state is empty
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "ModifyInstanceAttribute",
		Changes:      map[string]interface{}{"instance_type": "t3.large"},
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "admin"},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

// Keep original non-table-driven tests for backwards compatibility
func TestHandleEvent_ResourceNotFound(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateManager := &terraform.StateManager{}
	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-nonexistent",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleEvent_DriftDetected(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "instance-type-change",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"instance_type"},
			Severity:          "high",
		},
	}

	cfg := &config.Config{
		DriftRules: driftRules,
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"instance_type": "t3.large",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleEvent_NoDrift(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "DescribeInstances",
		Changes:      map[string]interface{}{},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}
