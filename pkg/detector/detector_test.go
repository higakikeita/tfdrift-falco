package detector

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNew_TableDriven covers various initialization scenarios using table-driven tests
func TestNew_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		cfg             *config.Config
		wantErr         bool
		wantImporter    bool
		wantApproval    bool
	}{
		{
			name: "basic_success_no_auto_import",
			cfg: &config.Config{
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
				AutoImport:    config.AutoImportConfig{Enabled: false},
			},
			wantErr:      false,
			wantImporter: false,
			wantApproval: false,
		},
		{
			name: "auto_import_with_manual_approval",
			cfg: &config.Config{
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
			},
			wantErr:      false,
			wantImporter: true,
			wantApproval: true,
		},
		{
			name: "auto_import_with_whitelist_auto_approval",
			cfg: &config.Config{
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
			},
			wantErr:      false,
			wantImporter: true,
			wantApproval: true,
		},
		{
			name: "auto_import_no_whitelist_auto_approve_all",
			cfg: &config.Config{
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
				Notifications: config.NotificationsConfig{},
				DriftRules:    []config.DriftRule{},
				DryRun:        true,
				AutoImport: config.AutoImportConfig{
					Enabled:          true,
					TerraformDir:     ".",
					OutputDir:        "/tmp",
					AllowedResources: []string{},
					RequireApproval:  false,
				},
			},
			wantErr:      false,
			wantImporter: true,
			wantApproval: true,
		},
		{
			name: "multiple_regions",
			cfg: &config.Config{
				Providers: config.ProvidersConfig{
					AWS: config.AWSConfig{
						Enabled: true,
						Regions: []string{"us-east-1", "us-west-2", "ap-northeast-1"},
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
				AutoImport:    config.AutoImportConfig{Enabled: false},
			},
			wantErr:      false,
			wantImporter: false,
			wantApproval: false,
		},
		{
			name: "with_drift_rules",
			cfg: &config.Config{
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
				DriftRules: []config.DriftRule{
					{
						Name:              "sg-rule",
						ResourceTypes:     []string{"aws_security_group"},
						WatchedAttributes: []string{"ingress", "egress"},
						Severity:          "critical",
					},
					{
						Name:              "instance-type",
						ResourceTypes:     []string{"aws_instance"},
						WatchedAttributes: []string{"instance_type"},
						Severity:          "high",
					},
				},
				DryRun:     true,
				AutoImport: config.AutoImportConfig{Enabled: false},
			},
			wantErr:      false,
			wantImporter: false,
			wantApproval: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detector, err := New(tt.cfg)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, detector)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, detector)
			assert.NotNil(t, detector.stateManager)
			assert.NotNil(t, detector.falcoSubscriber)
			assert.NotNil(t, detector.notifier)
			assert.NotNil(t, detector.formatter)
			assert.NotNil(t, detector.eventCh)

			if tt.wantImporter {
				assert.NotNil(t, detector.importer)
			} else {
				assert.Nil(t, detector.importer)
			}
			if tt.wantApproval {
				assert.NotNil(t, detector.approvalManager)
			} else {
				assert.Nil(t, detector.approvalManager)
			}
		})
	}
}

// TestDetector_Accessors tests getter/setter methods
func TestDetector_Accessors(t *testing.T) {
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
		AutoImport:    config.AutoImportConfig{Enabled: false},
	}

	d, err := New(cfg)
	require.NoError(t, err)

	// Test GetStateManager
	assert.NotNil(t, d.GetStateManager())
	assert.Equal(t, d.stateManager, d.GetStateManager())

	// Test Broadcaster getter/setter (initially nil)
	assert.Nil(t, d.GetBroadcaster())
	d.SetBroadcaster(nil) // should not panic
	assert.Nil(t, d.GetBroadcaster())

	// Test GraphStore getter/setter (initially nil)
	assert.Nil(t, d.GetGraphStore())
	d.SetGraphStore(nil) // should not panic
	assert.Nil(t, d.GetGraphStore())
}
