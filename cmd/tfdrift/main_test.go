package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitLogger(t *testing.T) {
	initLogger()

	// Verify logger is configured
	assert.Equal(t, log.InfoLevel, log.GetLevel())

	// Test that logging works with output capture
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	log.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "test message")
	assert.Contains(t, output, "level=info")
}

func TestNewApprovalCmd(t *testing.T) {
	cmd := newApprovalCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "approval", cmd.Use)
	assert.Equal(t, "Manage import approval requests", cmd.Short)
	assert.True(t, cmd.HasSubCommands())

	// Check that all subcommands are present
	subcommands := cmd.Commands()
	assert.Len(t, subcommands, 4)

	subcommandNames := make([]string, len(subcommands))
	for i, subcmd := range subcommands {
		subcommandNames[i] = subcmd.Use
	}

	assert.Contains(t, subcommandNames, "list")
	assert.Contains(t, subcommandNames, "approve [request-id]")
	assert.Contains(t, subcommandNames, "reject [request-id]")
	assert.Contains(t, subcommandNames, "cleanup")
}

func TestNewApprovalListCmd(t *testing.T) {
	cmd := newApprovalListCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List pending approval requests", cmd.Short)
	assert.NotNil(t, cmd.Run)
}

func TestNewApprovalListCmd_Execute(t *testing.T) {
	cmd := newApprovalListCmd()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the command
	cmd.Run(cmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify output contains expected messages
	assert.Contains(t, output, "This feature requires a running TFDrift-Falco instance")
	assert.Contains(t, output, "approval requests are only available during interactive sessions")
	assert.Contains(t, output, "To use approval workflow:")
	assert.Contains(t, output, "Enable auto_import in config.yaml")
	assert.Contains(t, output, "require_approval: true")
	assert.Contains(t, output, "tfdrift --config config.yaml --interactive")
}

func TestNewApprovalApproveCmd(t *testing.T) {
	cmd := newApprovalApproveCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "approve [request-id]", cmd.Use)
	assert.Equal(t, "Approve a specific import request", cmd.Short)
	assert.NotNil(t, cmd.Run)

	// Verify it requires exactly one argument
	cmd.SetArgs([]string{"test-request-123"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewApprovalApproveCmd_NoArgs(t *testing.T) {
	cmd := newApprovalApproveCmd()

	// Test without arguments (should fail)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestNewApprovalRejectCmd(t *testing.T) {
	cmd := newApprovalRejectCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "reject [request-id]", cmd.Use)
	assert.Equal(t, "Reject a specific import request", cmd.Short)
	assert.NotNil(t, cmd.Run)

	// Test with valid argument
	cmd.SetArgs([]string{"test-request-456"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewApprovalRejectCmd_WithReason(t *testing.T) {
	cmd := newApprovalRejectCmd()

	// Test with reason flag
	cmd.SetArgs([]string{"test-request-789", "--reason", "Security policy violation"})
	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewApprovalRejectCmd_NoArgs(t *testing.T) {
	cmd := newApprovalRejectCmd()

	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.Error(t, err)
}

func TestNewApprovalCleanupCmd(t *testing.T) {
	cmd := newApprovalCleanupCmd()

	assert.NotNil(t, cmd)
	assert.Equal(t, "cleanup", cmd.Use)
	assert.Equal(t, "Clean up expired approval requests", cmd.Short)
	assert.NotNil(t, cmd.Run)

	// Test with default older-than value
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestNewApprovalCleanupCmd_CustomDuration(t *testing.T) {
	cmd := newApprovalCleanupCmd()

	cmd.SetArgs([]string{"--older-than", "48h"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestVersion(t *testing.T) {
	assert.NotEmpty(t, version)
	assert.Equal(t, "0.4.1", version)
}

func TestLoggerOutput(t *testing.T) {
	// Test that logger can write to different outputs
	var buf bytes.Buffer

	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
	log.Info("test info message")

	output := buf.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "level=info")
}

func TestLoggerFormatter(t *testing.T) {
	// Test that text formatter is used
	var buf bytes.Buffer

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	log.Info("formatted message")

	output := buf.String()
	assert.Contains(t, output, "formatted message")
	// Should contain timestamp with the specified format
	assert.Contains(t, output, "time=")
}

func TestApprovalSubcommandIntegration(t *testing.T) {
	// Test that approval subcommand can be created and has all expected subcommands
	approvalCmd := newApprovalCmd()

	// Verify all subcommands
	subcommands := approvalCmd.Commands()
	assert.Len(t, subcommands, 4)

	// Test that each subcommand can be retrieved
	listCmd, _, err := approvalCmd.Find([]string{"list"})
	assert.NoError(t, err)
	assert.NotNil(t, listCmd)

	approveCmd, _, err := approvalCmd.Find([]string{"approve"})
	assert.NoError(t, err)
	assert.NotNil(t, approveCmd)

	rejectCmd, _, err := approvalCmd.Find([]string{"reject"})
	assert.NoError(t, err)
	assert.NotNil(t, rejectCmd)

	cleanupCmd, _, err := approvalCmd.Find([]string{"cleanup"})
	assert.NoError(t, err)
	assert.NotNil(t, cleanupCmd)
}

func TestApprovalCommandHelp(t *testing.T) {
	// Test that help text is available
	cmd := newApprovalCmd()

	assert.NotEmpty(t, cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Equal(t, "Manage import approval requests", cmd.Short)
}

// Tests for L1 Semi-Auto Mode overrides

func TestApplyConfigOverrides_NoOverrides(t *testing.T) {
	// Reset global flags
	regionOverride = nil
	falcoEndpoint = ""
	statePathOverride = ""
	backendTypeOverride = ""

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend: "local",
				},
			},
		},
		Falco: config.FalcoConfig{
			Hostname: "localhost",
			Port:     5060,
		},
	}

	err := applyConfigOverrides(cfg)
	require.NoError(t, err)

	// Config should remain unchanged
	assert.Equal(t, []string{"us-east-1"}, cfg.Providers.AWS.Regions)
	assert.Equal(t, "localhost", cfg.Falco.Hostname)
	assert.Equal(t, uint16(5060), cfg.Falco.Port)
	assert.Equal(t, "local", cfg.Providers.AWS.State.Backend)
}

func TestApplyConfigOverrides_RegionOverride(t *testing.T) {
	// Set region override
	regionOverride = []string{"us-west-2", "ap-northeast-1"}
	defer func() { regionOverride = nil }()

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Regions: []string{"us-east-1"},
			},
		},
		Falco: config.FalcoConfig{
			Hostname: "localhost",
			Port:     5060,
		},
	}

	err := applyConfigOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, []string{"us-west-2", "ap-northeast-1"}, cfg.Providers.AWS.Regions)
}

func TestApplyConfigOverrides_FalcoEndpoint(t *testing.T) {
	// Set Falco endpoint override
	falcoEndpoint = "prod-falco:5061"
	defer func() { falcoEndpoint = "" }()

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Regions: []string{"us-east-1"},
			},
		},
		Falco: config.FalcoConfig{
			Hostname: "localhost",
			Port:     5060,
		},
	}

	err := applyConfigOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "prod-falco", cfg.Falco.Hostname)
	assert.Equal(t, uint16(5061), cfg.Falco.Port)
}

func TestApplyConfigOverrides_FalcoEndpoint_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		wantErr  string
	}{
		{
			name:     "missing port",
			endpoint: "localhost",
			wantErr:  "invalid Falco endpoint format",
		},
		{
			name:     "invalid port",
			endpoint: "localhost:abc",
			wantErr:  "invalid port in Falco endpoint",
		},
		{
			name:     "empty",
			endpoint: "",
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			falcoEndpoint = tt.endpoint
			defer func() { falcoEndpoint = "" }()

			cfg := &config.Config{
				Providers: config.ProvidersConfig{
					AWS: config.AWSConfig{},
				},
				Falco: config.FalcoConfig{
					Hostname: "localhost",
					Port:     5060,
				},
			}

			err := applyConfigOverrides(cfg)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestApplyConfigOverrides_StatePath(t *testing.T) {
	// Set state path override
	statePathOverride = "/custom/terraform.tfstate"
	defer func() { statePathOverride = "" }()

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				State: config.TerraformStateConfig{
					Backend:   "s3",
					LocalPath: "",
				},
			},
		},
		Falco: config.FalcoConfig{
			Hostname: "localhost",
			Port:     5060,
		},
	}

	err := applyConfigOverrides(cfg)
	require.NoError(t, err)

	assert.Equal(t, "/custom/terraform.tfstate", cfg.Providers.AWS.State.LocalPath)
	assert.Equal(t, "local", cfg.Providers.AWS.State.Backend) // Should be changed to local
}

func TestApplyConfigOverrides_BackendType(t *testing.T) {
	tests := []struct {
		name        string
		backendType string
		wantErr     bool
	}{
		{
			name:        "local backend",
			backendType: "local",
			wantErr:     false,
		},
		{
			name:        "s3 backend",
			backendType: "s3",
			wantErr:     false,
		},
		{
			name:        "invalid backend",
			backendType: "gcs",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backendTypeOverride = tt.backendType
			defer func() { backendTypeOverride = "" }()

			cfg := &config.Config{
				Providers: config.ProvidersConfig{
					AWS: config.AWSConfig{
						State: config.TerraformStateConfig{
							Backend: "local",
						},
					},
				},
				Falco: config.FalcoConfig{
					Hostname: "localhost",
					Port:     5060,
				},
			}

			err := applyConfigOverrides(cfg)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid backend type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.backendType, cfg.Providers.AWS.State.Backend)
			}
		})
	}
}

func TestApplyConfigOverrides_MultipleOverrides(t *testing.T) {
	// Set multiple overrides
	regionOverride = []string{"us-west-2"}
	falcoEndpoint = "prod:5061"
	statePathOverride = "/custom/state.json"
	backendTypeOverride = "local"
	defer func() {
		regionOverride = nil
		falcoEndpoint = ""
		statePathOverride = ""
		backendTypeOverride = ""
	}()

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend: "s3",
				},
			},
		},
		Falco: config.FalcoConfig{
			Hostname: "localhost",
			Port:     5060,
		},
	}

	err := applyConfigOverrides(cfg)
	require.NoError(t, err)

	// All overrides should be applied
	assert.Equal(t, []string{"us-west-2"}, cfg.Providers.AWS.Regions)
	assert.Equal(t, "prod", cfg.Falco.Hostname)
	assert.Equal(t, uint16(5061), cfg.Falco.Port)
	assert.Equal(t, "/custom/state.json", cfg.Providers.AWS.State.LocalPath)
	assert.Equal(t, "local", cfg.Providers.AWS.State.Backend)
}
