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

func TestHandleEvent_ResourceNotFound(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	// Create minimal detector with mock state manager
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

	// Should not panic when resource not found
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

	// Create state manager with test resource
	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	// Load the state
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
			"instance_type": "t3.large", // Changed from t3.micro in state
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic when processing drift
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

	// Load the state
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
		Changes:      map[string]interface{}{
			// No actual changes
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic or produce alerts
	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}
