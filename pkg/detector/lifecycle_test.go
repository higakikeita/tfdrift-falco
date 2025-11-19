package detector

import (
	"context"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessEvents(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{
			{
				Name:              "instance-type-change",
				ResourceTypes:     []string{"aws_instance"},
				WatchedAttributes: []string{"instance_type"},
				Severity:          "high",
			},
		},
		DryRun: true,
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
		eventCh:      make(chan types.Event, 10),
	}

	// Start processEvents in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go detector.processEvents(ctx)

	// Send a test event
	testEvent := types.Event{
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

	detector.eventCh <- testEvent

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop processEvents
	cancel()

	// Give it time to shutdown
	time.Sleep(50 * time.Millisecond)

	// Test passed if no panic occurred
	assert.True(t, true)
}

func TestProcessEvents_Cancellation(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
	}

	detector := &Detector{
		cfg:     cfg,
		eventCh: make(chan types.Event, 10),
	}

	// Start processEvents with a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return quickly without panic
	done := make(chan bool)
	go func() {
		detector.processEvents(ctx)
		done <- true
	}()

	select {
	case <-done:
		// Success - processEvents returned
		assert.True(t, true)
	case <-time.After(1 * time.Second):
		t.Fatal("processEvents did not return after context cancellation")
	}
}

func TestStartCollectors_Integration(t *testing.T) {
	// This is a minimal test since startCollectors calls Falco subscriber
	// which requires a real Falco connection
	cfg := &config.Config{
		Falco: config.FalcoConfig{
			Enabled:  false, // Disabled to avoid connection
			Hostname: "localhost",
			Port:     5060,
		},
	}

	falcoSub, err := falco.NewSubscriber(cfg.Falco)
	require.NoError(t, err)

	detector := &Detector{
		cfg:             cfg,
		falcoSubscriber: falcoSub,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should return error since Falco is not running, but shouldn't panic
	err = detector.startCollectors(ctx)
	// Error is expected since we're not running Falco
	assert.Error(t, err)
}

func TestStart_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false, // Disabled to avoid connection
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, detector)

	// Start with a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start returns nil (starts goroutines) even if Falco is not running
	// The collector error is logged but doesn't stop the detector
	err = detector.Start(ctx)
	assert.NoError(t, err) // Start() succeeds, errors are logged in goroutines

	// Wait for context timeout
	<-ctx.Done()

	// Give goroutines time to clean up
	time.Sleep(50 * time.Millisecond)
}

func TestStart_StateLoadError(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/nonexistent.tfstate", // Invalid path
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should return error immediately due to state load failure
	err = detector.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform state")
}

func TestProcessEvents_MultipleEvents(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{
			{
				Name:              "test-rule",
				ResourceTypes:     []string{"aws_instance"},
				WatchedAttributes: []string{"instance_type"},
				Severity:          "high",
			},
		},
		DryRun: true,
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
		eventCh:      make(chan types.Event, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go detector.processEvents(ctx)

	// Send multiple events
	for i := 0; i < 3; i++ {
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
		detector.eventCh <- event
	}

	// Give time to process all events
	time.Sleep(200 * time.Millisecond)

	cancel()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, true) // No panic = success
}
