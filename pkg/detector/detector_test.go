package detector

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Success(t *testing.T) {
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

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.stateManager)
	assert.NotNil(t, detector.falcoSubscriber)
	assert.NotNil(t, detector.notifier)
	assert.NotNil(t, detector.formatter)
	assert.NotNil(t, detector.eventCh)
	assert.Nil(t, detector.importer)
	assert.Nil(t, detector.approvalManager)
}

func TestNew_WithAutoImport(t *testing.T) {
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
			Enabled:         true,
			TerraformDir:    ".",
			RequireApproval: true,
		},
	}

	detector, err := New(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestNew_WithAutoImportAutoApproval(t *testing.T) {
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
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance", "aws_s3_bucket"},
		},
	}

	detector, err := New(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestNew_WithAutoImportNoApproval(t *testing.T) {
	// Test New with auto-import enabled but no approval required
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
			Enabled:  true,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled: false,
			},
		},
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			OutputDir:        "/tmp",
			AllowedResources: []string{"aws_instance", "aws_s3_bucket"},
			RequireApproval:  false, // Auto-approve with whitelist
		},
		DryRun: true,
	}

	detector, err := New(cfg)

	require.NoError(t, err)
	require.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestNew_WithAutoImportWithApproval(t *testing.T) {
	// Test New with auto-import enabled and approval required
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
			Enabled:  true,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled: false,
			},
		},
		AutoImport: config.AutoImportConfig{
			Enabled:         true,
			TerraformDir:    ".",
			OutputDir:       "/tmp",
			RequireApproval: true, // Manual approval required
		},
		DryRun: true,
	}

	detector, err := New(cfg)

	require.NoError(t, err)
	require.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestNew_WithAutoImportNoWhitelist(t *testing.T) {
	// Test New with auto-import enabled, no approval, no whitelist (auto-approve all)
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
			Enabled:  true,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled: false,
			},
		},
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			OutputDir:        "/tmp",
			AllowedResources: []string{}, // Empty whitelist
			RequireApproval:  false,
		},
		DryRun: true,
	}

	detector, err := New(cfg)

	require.NoError(t, err)
	require.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}
