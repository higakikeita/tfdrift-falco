package detector

import (
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test correlator.GetGroups function
func TestCrossCloudCorrelator_GetGroups(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Initially should be empty
	groups := c.GetGroups()
	assert.Empty(t, groups)

	// Add events that correlate
	c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	})

	c.AddEvent(types.Event{
		Provider:     "gcp",
		EventName:    "instances.insert",
		ResourceType: "google_compute_instance",
		ResourceID:   "instance-1",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	})

	// Now should have groups
	groups = c.GetGroups()
	assert.NotEmpty(t, groups, "Should have correlation groups after adding events")
	assert.Greater(t, len(groups), 0)

	// Verify it's a copy (modifying returned slice doesn't affect internal state)
	originalLen := len(groups)
	_ = append(groups, CorrelationGroup{ID: "fake"})
	groupsAgain := c.GetGroups()
	assert.Equal(t, originalLen, len(groupsAgain), "Returned slice should be a copy")
}

// Test correlator.GetGroupsByProvider function
func TestCrossCloudCorrelator_GetGroupsByProvider(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add AWS event
	c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "AuthorizeSecurityGroupIngress",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-123",
		UserIdentity: types.UserIdentity{UserName: "alice"},
	})

	// Add GCP event that correlates by resource pattern
	c.AddEvent(types.Event{
		Provider:     "gcp",
		EventName:    "compute.firewalls.create",
		ResourceType: "google_compute_firewall",
		ResourceID:   "gcp-fw",
		UserIdentity: types.UserIdentity{UserName: "bob"},
	})

	// Add Azure event
	c.AddEvent(types.Event{
		Provider:     "azure",
		EventName:    "Microsoft.Network/networkSecurityGroups/write",
		ResourceType: "azurerm_network_security_group",
		ResourceID:   "nsg-1",
		UserIdentity: types.UserIdentity{UserName: "charlie"},
	})

	// Test getting groups by AWS provider
	awsGroups := c.GetGroupsByProvider("aws")
	assert.NotEmpty(t, awsGroups, "Should have groups for AWS provider")
	for _, g := range awsGroups {
		found := false
		for _, p := range g.Providers {
			if p == "aws" {
				found = true
				break
			}
		}
		assert.True(t, found, "AWS group should contain AWS provider")
	}

	// Test getting groups by GCP provider
	gcpGroups := c.GetGroupsByProvider("gcp")
	assert.NotEmpty(t, gcpGroups, "Should have groups for GCP provider")

	// Test getting groups for non-existent provider
	nonExistentGroups := c.GetGroupsByProvider("nonexistent")
	assert.Empty(t, nonExistentGroups, "Should have no groups for non-existent provider")
}

// Test GetGroupsByProvider with multi-cloud groups
func TestCrossCloudCorrelator_GetGroupsByProvider_MultiCloud(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add compute resources across clouds to create multi-cloud group
	c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: "devops"},
	})

	c.AddEvent(types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.insert",
		ResourceType: "google_compute_instance",
		ResourceID:   "instance-456",
		UserIdentity: types.UserIdentity{UserName: "devops"},
	})

	c.AddEvent(types.Event{
		Provider:     "azure",
		EventName:    "Microsoft.Compute/virtualMachines/write",
		ResourceType: "azurerm_virtual_machine",
		ResourceID:   "vm-789",
		UserIdentity: types.UserIdentity{UserName: "devops"},
	})

	// Each provider should have groups that contain multi-cloud information
	awsGroups := c.GetGroupsByProvider("aws")
	assert.NotEmpty(t, awsGroups)
	// Verify these are multi-cloud groups
	for _, g := range awsGroups {
		assert.GreaterOrEqual(t, len(g.Providers), 2, "Groups should be multi-cloud")
	}

	gcpGroups := c.GetGroupsByProvider("gcp")
	azureGroups := c.GetGroupsByProvider("azure")
	assert.NotEmpty(t, gcpGroups)
	assert.NotEmpty(t, azureGroups)
}

// Test detector.GetStateManager
func TestDetector_GetStateManager(t *testing.T) {
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
			Enabled:  false,
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

	// Test GetStateManager
	sm := detector.GetStateManager()
	assert.NotNil(t, sm, "State manager should not be nil")
	assert.Equal(t, detector.stateManager, sm, "Should return the same state manager instance")
}

// Test detector.GetProviderRegistry
func TestDetector_GetProviderRegistry(t *testing.T) {
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
			Enabled:  false,
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

	// Test GetProviderRegistry
	registry := detector.GetProviderRegistry()
	assert.NotNil(t, registry, "Provider registry should not be nil")
	assert.Equal(t, detector.providerRegistry, registry, "Should return the same registry instance")
}

// Test detector.GetStateManagers with multiple providers
func TestDetector_GetStateManagers(t *testing.T) {
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
			GCP: config.GCPConfig{
				Enabled: true,
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
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
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	// Test GetStateManagers
	managers := detector.GetStateManagers()
	assert.NotNil(t, managers)
	assert.Equal(t, detector.stateManagers, managers, "Should return the same managers map")
	assert.Contains(t, managers, "aws", "Should have AWS state manager")
	assert.Contains(t, managers, "gcp", "Should have GCP state manager")
}

// Test detector.GetStateManagers with single provider
func TestDetector_GetStateManagers_SingleProvider(t *testing.T) {
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
			Enabled:  false,
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

	managers := detector.GetStateManagers()
	assert.Len(t, managers, 1, "Should have exactly one state manager")
	assert.Contains(t, managers, "aws")
}

// Test detector.SetBroadcaster and GetBroadcaster
func TestDetector_SetBroadcaster_GetBroadcaster(t *testing.T) {
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
			Enabled:  false,
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

	// Initially broadcaster should be nil
	bc := detector.GetBroadcaster()
	assert.Nil(t, bc)

	// Set broadcaster
	newBC := broadcaster.NewBroadcaster()
	detector.SetBroadcaster(newBC)

	// Get broadcaster and verify it's the same instance
	retrievedBC := detector.GetBroadcaster()
	assert.NotNil(t, retrievedBC)
	assert.Equal(t, newBC, retrievedBC, "Should return the same broadcaster instance")
}

// Test detector.SetGraphStore and GetGraphStore
func TestDetector_SetGraphStore_GetGraphStore(t *testing.T) {
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
			Enabled:  false,
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

	// Initially graph store should be nil
	gs := detector.GetGraphStore()
	assert.Nil(t, gs)

	// Set graph store
	newGS := graph.NewStore()
	detector.SetGraphStore(newGS)

	// Get graph store and verify it's the same instance
	retrievedGS := detector.GetGraphStore()
	assert.NotNil(t, retrievedGS)
	assert.Equal(t, newGS, retrievedGS, "Should return the same graph store instance")
}

// Test testing.HandleEventForTest
func TestDetector_HandleEventForTest(t *testing.T) {
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
			Enabled:  false,
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

	// Create a test event
	event := types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-test-123",
		UserIdentity: types.UserIdentity{
			UserName: "testuser",
		},
	}

	// Call HandleEventForTest - it should not panic
	assert.NotPanics(t, func() {
		detector.HandleEventForTest(event)
	}, "HandleEventForTest should not panic")
}

// Test testing.GetStateManagerForTest
func TestDetector_GetStateManagerForTest(t *testing.T) {
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
			Enabled:  false,
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

	// Get state manager via test helper
	sm := detector.GetStateManagerForTest()
	assert.NotNil(t, sm, "Should return state manager")
	assert.Equal(t, detector.stateManager, sm, "Should return the internal state manager")
}

// Test multiple broadcaster/graph store operations
func TestDetector_BroadcasterAndGraphStoreOperations(t *testing.T) {
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
			Enabled:  false,
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

	// Set broadcaster multiple times
	bc1 := broadcaster.NewBroadcaster()
	detector.SetBroadcaster(bc1)
	assert.Equal(t, bc1, detector.GetBroadcaster())

	bc2 := broadcaster.NewBroadcaster()
	detector.SetBroadcaster(bc2)
	assert.Equal(t, bc2, detector.GetBroadcaster())

	// Set graph store multiple times
	gs1 := graph.NewStore()
	detector.SetGraphStore(gs1)
	assert.Equal(t, gs1, detector.GetGraphStore())

	gs2 := graph.NewStore()
	detector.SetGraphStore(gs2)
	assert.Equal(t, gs2, detector.GetGraphStore())
}

// Test all provider types enabled simultaneously
func TestDetector_AllProvidersEnabled(t *testing.T) {
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
			GCP: config.GCPConfig{
				Enabled: true,
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
			Azure: config.AzureConfig{
				Enabled: true,
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
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
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	// Verify all state managers are available
	managers := detector.GetStateManagers()
	assert.Len(t, managers, 3)
	assert.Contains(t, managers, "aws")
	assert.Contains(t, managers, "gcp")
	assert.Contains(t, managers, "azure")

	// Verify default state manager is set (AWS takes precedence)
	sm := detector.GetStateManager()
	assert.NotNil(t, sm)
	assert.Equal(t, managers["aws"], sm)
}

// Test state manager precedence (GCP as default when AWS disabled)
func TestDetector_StateManagerPrecedence_GCPDefault(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: false,
			},
			GCP: config.GCPConfig{
				Enabled: true,
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
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
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	// Verify GCP is the default when AWS is disabled
	managers := detector.GetStateManagers()
	assert.Contains(t, managers, "gcp")
	sm := detector.GetStateManager()
	assert.NotNil(t, sm)
	assert.Equal(t, managers["gcp"], sm)
}

// Test state manager precedence (Azure as default when AWS and GCP disabled)
func TestDetector_StateManagerPrecedence_AzureDefault(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: false,
			},
			GCP: config.GCPConfig{
				Enabled: false,
			},
			Azure: config.AzureConfig{
				Enabled: true,
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
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
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	// Verify Azure is the default when AWS and GCP are disabled
	managers := detector.GetStateManagers()
	assert.Contains(t, managers, "azure")
	sm := detector.GetStateManager()
	assert.NotNil(t, sm)
	assert.Equal(t, managers["azure"], sm)
}
