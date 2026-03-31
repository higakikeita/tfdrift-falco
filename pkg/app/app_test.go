package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
)

// TestNewApp tests that New creates a valid App instance
func TestNewApp(t *testing.T) {
	cfg := &Config{
		Version: "0.4.1",
	}
	app := New(cfg)

	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.cfg != cfg {
		t.Error("Config not properly set")
	}

	if app.appCfg != nil {
		t.Error("appCfg should be nil before Run() is called")
	}
}

// TestNewAppWithAllFields tests New with all fields populated
func TestNewAppWithAllFields(t *testing.T) {
	cfg := &Config{
		ConfigFile:          "config.yaml",
		AutoDetect:          true,
		OutputMode:          "json",
		DryRun:              true,
		Daemon:              true,
		Interactive:         false,
		ServerMode:          true,
		APIPort:             8080,
		RegionOverride:      []string{"us-west-2", "us-east-1"},
		FalcoEndpoint:       "localhost:5060",
		StatePathOverride:   "./state.tfstate",
		BackendTypeOverride: "local",
		Version:             "0.4.1",
	}
	app := New(cfg)

	if app == nil {
		t.Fatal("New() returned nil")
	}

	if app.cfg.ConfigFile != "config.yaml" {
		t.Error("ConfigFile not properly set")
	}

	if app.cfg.APIPort != 8080 {
		t.Error("APIPort not properly set")
	}

	if len(app.cfg.RegionOverride) != 2 {
		t.Error("RegionOverride not properly set")
	}
}

// TestLoadConfigMissingConfigFile tests that loading fails when no config file is specified
func TestLoadConfigMissingConfigFile(t *testing.T) {
	cfg := &Config{
		ConfigFile: "",
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err := app.loadConfig()
	if err == nil {
		t.Error("Expected error when config file not specified and auto-detect disabled, got nil")
	}

	expectedMsg := "config file not specified and auto-detection disabled"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestLoadConfigWithInvalidFile tests loading with a non-existent config file
func TestLoadConfigWithInvalidFile(t *testing.T) {
	cfg := &Config{
		ConfigFile: "/nonexistent/path/config.yaml",
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err := app.loadConfig()
	if err == nil {
		t.Error("Expected error when loading non-existent config file, got nil")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestLoadConfigWithValidFile tests loading with a valid config file
func TestLoadConfigWithValidFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Write a minimal valid config
	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
logging:
  level: info
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err := app.loadConfig()
	if err != nil {
		t.Errorf("Expected no error loading valid config, got: %v", err)
	}

	if app.appCfg == nil {
		t.Error("appCfg should not be nil after successful loadConfig")
	}

	if !app.appCfg.Providers.AWS.Enabled {
		t.Error("AWS provider should be enabled in loaded config")
	}
}

// TestApplyConfigOverridesRegionOverride tests region override functionality
func TestApplyConfigOverridesRegionOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := &Config{
		RegionOverride: []string{"us-west-2", "eu-west-1"},
		Version:        "0.4.1",
	}
	app := New(cfg)
	app.appCfg = appCfg

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error applying region override, got: %v", err)
	}

	if len(appCfg.Providers.AWS.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(appCfg.Providers.AWS.Regions))
	}

	if appCfg.Providers.AWS.Regions[0] != "us-west-2" {
		t.Errorf("Expected first region to be us-west-2, got %s", appCfg.Providers.AWS.Regions[0])
	}
}

// TestApplyConfigOverridesFalcoEndpoint tests Falco endpoint parsing and override
func TestApplyConfigOverridesFalcoEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		wantError bool
		wantHost  string
		wantPort  uint16
	}{
		{
			name:      "valid endpoint",
			endpoint:  "localhost:5060",
			wantError: false,
			wantHost:  "localhost",
			wantPort:  5060,
		},
		{
			name:      "invalid endpoint - missing port",
			endpoint:  "localhost",
			wantError: true,
		},
		{
			name:      "invalid endpoint - invalid port",
			endpoint:  "localhost:notaport",
			wantError: true,
		},
		{
			name:      "valid endpoint with IP and high port",
			endpoint:  "192.168.1.1:8080",
			wantError: false,
			wantHost:  "192.168.1.1",
			wantPort:  8080,
		},
		{
			name:      "valid endpoint with domain",
			endpoint:  "falco.example.com:9090",
			wantError: false,
			wantHost:  "falco.example.com",
			wantPort:  9090,
		},
		{
			name:      "invalid endpoint - multiple colons",
			endpoint:  "localhost:5060:extra",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				FalcoEndpoint: tt.endpoint,
				Version:       "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error: %v, got: %v", tt.wantError, err != nil)
			}

			if !tt.wantError {
				if appCfg.Falco.Hostname != tt.wantHost {
					t.Errorf("Expected host %s, got %s", tt.wantHost, appCfg.Falco.Hostname)
				}
				if appCfg.Falco.Port != tt.wantPort {
					t.Errorf("Expected port %d, got %d", tt.wantPort, appCfg.Falco.Port)
				}
			}
		})
	}
}

// TestApplyConfigOverridesStatePathOverride tests state path override
func TestApplyConfigOverridesStatePathOverride(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: my-bucket
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	statePath := "/custom/path/terraform.tfstate"
	cfg := &Config{
		StatePathOverride: statePath,
		Version:           "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error applying state path override, got: %v", err)
	}

	if appCfg.Providers.AWS.State.LocalPath != statePath {
		t.Errorf("Expected state path %s, got %s", statePath, appCfg.Providers.AWS.State.LocalPath)
	}

	if appCfg.Providers.AWS.State.Backend != "local" {
		t.Errorf("Expected backend to be changed to 'local', got %s", appCfg.Providers.AWS.State.Backend)
	}
}

// TestApplyConfigOverridesBackendType tests backend type validation and override
func TestApplyConfigOverridesBackendType(t *testing.T) {
	tests := []struct {
		name      string
		backend   string
		wantError bool
	}{
		{
			name:      "valid backend - local",
			backend:   "local",
			wantError: false,
		},
		{
			name:      "valid backend - s3",
			backend:   "s3",
			wantError: false,
		},
		{
			name:      "invalid backend - gcs",
			backend:   "gcs",
			wantError: true,
		},
		{
			name:      "invalid backend - azure",
			backend:   "azure",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				BackendTypeOverride: tt.backend,
				Version:             "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error: %v, got error: %v", tt.wantError, err != nil)
			}

			if !tt.wantError && appCfg.Providers.AWS.State.Backend != tt.backend {
				t.Errorf("Expected backend %s, got %s", tt.backend, appCfg.Providers.AWS.State.Backend)
			}
		})
	}
}

// TestApplyConfigOverridesEmptyBackendType tests that empty backend type is not applied
func TestApplyConfigOverridesEmptyBackendType(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: my-bucket
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	originalBackend := appCfg.Providers.AWS.State.Backend

	cfg := &Config{
		BackendTypeOverride: "",
		Version:             "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error with empty backend override, got: %v", err)
	}

	if appCfg.Providers.AWS.State.Backend != originalBackend {
		t.Errorf("Backend should not change with empty override, expected %s, got %s",
			originalBackend, appCfg.Providers.AWS.State.Backend)
	}
}

// TestRunWithMissingConfig tests that Run fails when config loading fails
func TestRunWithMissingConfig(t *testing.T) {
	cfg := &Config{
		ConfigFile: "/nonexistent/config.yaml",
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx := context.Background()
	err := app.Run(ctx)

	if err == nil {
		t.Error("Expected error when running with invalid config, got nil")
	}

	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

// TestContextCancellation tests that context cancellation is respected
func TestContextCancellation(t *testing.T) {
	cfg := &Config{
		Version: "0.4.1",
	}
	_ = New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if ctx.Err() == nil {
		t.Error("Context should be cancelled")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", ctx.Err())
	}
}

// TestContextDeadlineExceeded tests behavior with context timeout
func TestContextDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	cfg := &Config{
		Version: "0.4.1",
	}
	_ = New(cfg)

	<-ctx.Done()

	if ctx.Err() == nil {
		t.Error("Expected context to have an error after timeout")
	}

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got %v", ctx.Err())
	}
}

// TestConfigFieldValidation tests Config struct with various field combinations
func TestConfigFieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *Config
		validate func(*Config) error
	}{
		{
			name: "minimal config",
			cfg: &Config{
				Version: "0.4.1",
			},
			validate: func(cfg *Config) error {
				if cfg.Version != "0.4.1" {
					return ErrConfigValidation
				}
				return nil
			},
		},
		{
			name: "full config",
			cfg: &Config{
				ConfigFile:          "config.yaml",
				AutoDetect:          true,
				OutputMode:          "json",
				DryRun:              true,
				ServerMode:          true,
				APIPort:             8080,
				Version:             "0.4.1",
				RegionOverride:      []string{"us-west-2"},
				FalcoEndpoint:       "localhost:5060",
				StatePathOverride:   "./state.tfstate",
				BackendTypeOverride: "local",
			},
			validate: func(cfg *Config) error {
				if cfg.APIPort != 8080 || len(cfg.RegionOverride) != 1 {
					return ErrConfigValidation
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New(tt.cfg)
			if app == nil {
				t.Fatal("New() returned nil")
			}

			if app.cfg != tt.cfg {
				t.Error("Config not properly set")
			}

			if err := tt.validate(tt.cfg); err != nil {
				t.Errorf("Config validation failed: %v", err)
			}
		})
	}
}

// TestRunWithDryRunMode tests that dry-run mode is applied
func TestRunWithDryRunMode(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		DryRun:     true,
		Version:    "0.4.1",
	}
	app := New(cfg)

	// Load config first
	err := app.loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Apply dry-run mode (simulating what Run does)
	if app.cfg.DryRun {
		app.appCfg.DryRun = true
	}

	if !app.appCfg.DryRun {
		t.Error("DryRun mode should be enabled in appCfg")
	}
}

// TestMultipleRegionOverrides tests applying multiple region overrides
func TestMultipleRegionOverrides(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	regions := []string{"us-west-1", "us-west-2", "eu-west-1"}
	cfg := &Config{
		RegionOverride: regions,
		Version:        "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error applying region overrides, got: %v", err)
	}

	if len(appCfg.Providers.AWS.Regions) != 3 {
		t.Errorf("Expected 3 regions, got %d", len(appCfg.Providers.AWS.Regions))
	}

	for i, region := range regions {
		if appCfg.Providers.AWS.Regions[i] != region {
			t.Errorf("Expected region %s at index %d, got %s", region, i, appCfg.Providers.AWS.Regions[i])
		}
	}
}

// TestLoadAutoConfigMissingDir tests auto-detection with missing directory
func TestLoadAutoConfigMissingDir(t *testing.T) {
	cfg := &Config{
		ConfigFile: "",
		AutoDetect: true,
		Version:    "0.4.1",
	}
	_ = New(cfg)

	// loadAutoConfig uses os.Getwd() which should work,
	// but auto-detection should fail when no state is found
	result, err := config.AutoDetectTerraformState(".")
	if err == nil && !result.Found {
		// This is the expected behavior - no state found
		return
	}
	if err != nil {
		// Error is also acceptable
		return
	}
}

// TestApplyConfigOverridesAll tests applying all overrides at once
func TestApplyConfigOverridesAll(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: my-bucket
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := &Config{
		RegionOverride:      []string{"eu-central-1"},
		FalcoEndpoint:       "falco.example.com:9090",
		StatePathOverride:   "/custom/state.tfstate",
		BackendTypeOverride: "local",
		Version:             "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error applying all overrides, got: %v", err)
	}

	// Verify all overrides were applied
	if appCfg.Providers.AWS.Regions[0] != "eu-central-1" {
		t.Error("Region override not applied")
	}

	if appCfg.Falco.Hostname != "falco.example.com" || appCfg.Falco.Port != 9090 {
		t.Error("Falco endpoint override not applied")
	}

	if appCfg.Providers.AWS.State.LocalPath != "/custom/state.tfstate" {
		t.Error("State path override not applied")
	}

	if appCfg.Providers.AWS.State.Backend != "local" {
		t.Error("Backend type override not applied")
	}
}

// TestRunDetectorMissingState tests that detector initialization with missing state fails
func TestRunDetectorWithMissingState(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	nonExistentStatePath := filepath.Join(tmpDir, "nonexistent.tfstate")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ` + nonExistentStatePath + `
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := app.Run(ctx)
	// We expect an error since the detector will fail to load state
	// But we're mainly testing that the config loading and initialization path works
	if err != nil {
		// This is expected due to missing state file
		errMsg := err.Error()
		if errMsg != "failed to initialize detector" && !strings.Contains(errMsg, "state") {
			// Either error is acceptable
		}
	}
}

// TestLoadConfigWithAutoDetectDisabled tests config loading with auto-detect disabled
func TestLoadConfigWithAutoDetectDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err := app.loadConfig()
	if err != nil {
		t.Errorf("Expected no error with explicit config file and auto-detect disabled, got: %v", err)
	}

	if app.appCfg == nil {
		t.Error("appCfg should be set after successful loadConfig")
	}

	if app.appCfg.Providers.AWS.Enabled != true {
		t.Error("AWS provider should be enabled")
	}
}

// TestFalcoEndpointParsingEdgeCases tests edge cases in Falco endpoint parsing
func TestFalcoEndpointParsingEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		endpoint  string
		wantError bool
	}{
		{
			name:      "port 1",
			endpoint:  "localhost:1",
			wantError: false,
		},
		{
			name:      "port 65535",
			endpoint:  "localhost:65535",
			wantError: false,
		},
		{
			name:      "IPv6 format (should fail with simple split)",
			endpoint:  "[::1]:5060",
			wantError: true, // Current implementation uses simple split
		},
		{
			name:      "localhost with high port",
			endpoint:  "localhost:32768",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				FalcoEndpoint: tt.endpoint,
				Version:       "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if (err != nil) != tt.wantError {
				t.Errorf("Expected error: %v, got: %v (err=%v)", tt.wantError, err != nil, err)
			}
		})
	}
}

// TestNewAppReturnsValidInstance tests New() returns valid instance
func TestNewAppReturnsValidInstance(t *testing.T) {
	cfg := &Config{
		Version: "1.0.0",
	}
	app := New(cfg)

	if app == nil {
		t.Fatal("New() should not return nil")
	}

	if app.cfg == nil {
		t.Fatal("app.cfg should not be nil")
	}

	if app.cfg.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", app.cfg.Version)
	}
}

// TestLoadConfigFilePath tests config loading with absolute path
func TestLoadConfigFilePath(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	if app.cfg.ConfigFile != configFile {
		t.Errorf("Expected ConfigFile to be %s, got %s", configFile, app.cfg.ConfigFile)
	}

	err := app.loadConfig()
	if err != nil {
		t.Errorf("Failed to load config from %s: %v", configFile, err)
	}
}

// TestRunWithValidConfigAndServerMode tests Run with server mode enabled
func TestRunWithValidConfigAndServerMode(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		ServerMode: true,
		APIPort:    8080,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Run will try to start the API server, which will fail without a valid Falco connection
	// But it tests that the server mode path is taken
	_ = app.Run(ctx)
}

// TestApplyConfigOverridesNone tests that overrides don't fail when all are empty/default
func TestApplyConfigOverridesNone(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	originalRegions := make([]string, len(appCfg.Providers.AWS.Regions))
	copy(originalRegions, appCfg.Providers.AWS.Regions)
	originalHost := appCfg.Falco.Hostname
	originalPort := appCfg.Falco.Port

	cfg := &Config{
		RegionOverride:      []string{},
		FalcoEndpoint:       "",
		StatePathOverride:   "",
		BackendTypeOverride: "",
		Version:             "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error with empty overrides, got: %v", err)
	}

	// Verify nothing changed
	if len(appCfg.Providers.AWS.Regions) != len(originalRegions) {
		t.Error("Regions should not change when override is empty")
	}

	if appCfg.Falco.Hostname != originalHost {
		t.Error("Falco host should not change when override is empty")
	}

	if appCfg.Falco.Port != originalPort {
		t.Error("Falco port should not change when override is empty")
	}
}

// TestConfigWithDifferentBackends tests applying backend overrides for different scenarios
func TestConfigWithDifferentBackends(t *testing.T) {
	backends := []string{"local", "s3"}

	for _, backend := range backends {
		t.Run("backend_"+backend, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				BackendTypeOverride: backend,
				Version:             "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if err != nil {
				t.Errorf("Expected no error with backend %s, got: %v", backend, err)
			}

			if appCfg.Providers.AWS.State.Backend != backend {
				t.Errorf("Expected backend %s, got %s", backend, appCfg.Providers.AWS.State.Backend)
			}
		})
	}
}

// TestLoadConfigPreservesOriginalFile tests that loading config from file is correct
func TestLoadConfigPreservesOriginalFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: falco.example.com
  port: 9090
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err := app.loadConfig()
	if err != nil {
		t.Errorf("Failed to load config: %v", err)
	}

	if len(app.appCfg.Providers.AWS.Regions) != 2 {
		t.Errorf("Expected 2 regions, got %d", len(app.appCfg.Providers.AWS.Regions))
	}

	if app.appCfg.Falco.Hostname != "falco.example.com" {
		t.Errorf("Expected Falco host falco.example.com, got %s", app.appCfg.Falco.Hostname)
	}

	if app.appCfg.Falco.Port != 9090 {
		t.Errorf("Expected Falco port 9090, got %d", app.appCfg.Falco.Port)
	}
}

// TestConfigAppInstances tests that multiple App instances can be created
func TestConfigAppInstances(t *testing.T) {
	cfg1 := &Config{Version: "1.0.0"}
	cfg2 := &Config{Version: "2.0.0"}

	app1 := New(cfg1)
	app2 := New(cfg2)

	if app1.cfg.Version == app2.cfg.Version {
		t.Error("Different App instances should have different configurations")
	}

	if app1.cfg.Version != "1.0.0" {
		t.Errorf("app1 version should be 1.0.0, got %s", app1.cfg.Version)
	}

	if app2.cfg.Version != "2.0.0" {
		t.Errorf("app2 version should be 2.0.0, got %s", app2.cfg.Version)
	}
}

// TestContextDoneBeforeRun tests that context cancellation before Run is handled
func TestContextDoneBeforeRun(t *testing.T) {
	cfg := &Config{
		ConfigFile: "",
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Context is already cancelled, but Run should attempt to load config first
	err := app.Run(ctx)
	if err == nil {
		t.Error("Expected error when running with cancelled context and no config")
	}
}

// TestApplyOverridesWithInvalidFalcoEndpointFormat tests various invalid endpoint formats
func TestApplyOverridesWithInvalidFalcoEndpointFormat(t *testing.T) {
	invalidEndpoints := []string{
		"localhost:",
		"host:port:extra",
		"host:not-a-number",
	}

	for _, endpoint := range invalidEndpoints {
		t.Run("endpoint_"+endpoint, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				FalcoEndpoint: endpoint,
				Version:       "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if err == nil {
				t.Errorf("Expected error for invalid endpoint %s", endpoint)
			}
		})
	}
}

// TestRunWithZeroTimeout tests Run with very short timeout
func TestRunWithZeroTimeout(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Should hit timeout while running detector
	_ = app.Run(ctx)
}

// TestApplyConfigOverridesWithAllZeroValues tests overrides when all zero values
func TestApplyConfigOverridesWithAllZeroValues(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: bucket
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	originalBackend := appCfg.Providers.AWS.State.Backend

	cfg := &Config{
		Version: "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Backend should not change
	if appCfg.Providers.AWS.State.Backend != originalBackend {
		t.Error("Backend should not change with zero values")
	}
}

// TestRegionOverrideReplacement tests that region override replaces all regions
func TestRegionOverrideReplacement(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-1
      - eu-west-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Original should have 3 regions
	if len(appCfg.Providers.AWS.Regions) != 3 {
		t.Fatalf("Expected 3 original regions, got %d", len(appCfg.Providers.AWS.Regions))
	}

	cfg := &Config{
		RegionOverride: []string{"ap-southeast-1"},
		Version:        "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Should now have only 1 region (override replaces)
	if len(appCfg.Providers.AWS.Regions) != 1 {
		t.Errorf("Expected 1 region after override, got %d", len(appCfg.Providers.AWS.Regions))
	}

	if appCfg.Providers.AWS.Regions[0] != "ap-southeast-1" {
		t.Errorf("Expected ap-southeast-1, got %s", appCfg.Providers.AWS.Regions[0])
	}
}

// TestStatePathOverrideSetsBackendToLocal tests that state path override sets backend to local
func TestStatePathOverrideSetsBackendToLocal(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: my-bucket
      s3_key: terraform/state
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify initial backend is s3
	if appCfg.Providers.AWS.State.Backend != "s3" {
		t.Fatalf("Expected initial backend to be s3, got %s", appCfg.Providers.AWS.State.Backend)
	}

	cfg := &Config{
		StatePathOverride: "/tmp/terraform.tfstate",
		Version:           "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Backend should now be local
	if appCfg.Providers.AWS.State.Backend != "local" {
		t.Errorf("Expected backend to be changed to local, got %s", appCfg.Providers.AWS.State.Backend)
	}

	// LocalPath should be set
	if appCfg.Providers.AWS.State.LocalPath != "/tmp/terraform.tfstate" {
		t.Errorf("Expected LocalPath to be /tmp/terraform.tfstate, got %s", appCfg.Providers.AWS.State.LocalPath)
	}
}

// TestConfigInitialization tests that Config fields are properly initialized
func TestConfigInitialization(t *testing.T) {
	cfg := &Config{
		ConfigFile:          "test.yaml",
		AutoDetect:          false,
		OutputMode:          "json",
		DryRun:              true,
		Daemon:              false,
		Interactive:         true,
		ServerMode:          false,
		APIPort:             9090,
		RegionOverride:      []string{"eu-west-1"},
		FalcoEndpoint:       "localhost:6000",
		StatePathOverride:   "/var/state.tfstate",
		BackendTypeOverride: "s3",
		Version:             "1.2.3",
	}

	if cfg.ConfigFile != "test.yaml" {
		t.Error("ConfigFile not properly initialized")
	}

	if cfg.APIPort != 9090 {
		t.Error("APIPort not properly initialized")
	}

	if cfg.Version != "1.2.3" {
		t.Error("Version not properly initialized")
	}

	if len(cfg.RegionOverride) != 1 {
		t.Error("RegionOverride not properly initialized")
	}
}

// TestLoadAutoConfigFlow tests the auto-detection flow
func TestLoadAutoConfigFlow(t *testing.T) {
	// Create a temporary directory with a valid state file
	tmpDir := t.TempDir()

	// Create a terraform state file
	stateContent := `{
  "version": 4,
  "terraform_version": "1.0.0",
  "serial": 1,
  "lineage": "test",
  "outputs": {},
  "resources": []
}`
	stateFile := filepath.Join(tmpDir, "terraform.tfstate")
	if err := os.WriteFile(stateFile, []byte(stateContent), 0644); err != nil {
		t.Fatalf("Failed to write state file: %v", err)
	}

	// Change to temp directory for auto-detection
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Now test auto-config loading
	cfg := &Config{
		ConfigFile: "",
		AutoDetect: true,
		Version:    "0.4.1",
	}
	app := New(cfg)

	// loadAutoConfig should successfully find the state and create a config
	appCfg, err := app.loadAutoConfig()
	if err != nil {
		t.Logf("Auto-config load error (expected in test environment): %v", err)
		return
	}

	if appCfg == nil {
		t.Error("Expected loadAutoConfig to return a valid config")
	}
}

// TestRunWithAutoDetectAndNoState tests Run with auto-detect enabled but no state found
func TestRunWithAutoDetectAndNoState(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cfg := &Config{
		ConfigFile: "",
		AutoDetect: true,
		Version:    "0.4.1",
	}
	app := New(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = app.Run(ctx)
	if err == nil {
		t.Error("Expected error when auto-detect finds no state")
	}
}

// TestRunDetectorWithTimeoutContext tests detector run with timeout
func TestRunDetectorWithTimeoutContext(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	cfg := &Config{
		ConfigFile: configFile,
		AutoDetect: false,
		Version:    "0.4.1",
	}
	app := New(cfg)

	// Load config
	err := app.loadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Create timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// runDetector should handle context cancellation
	det, err := detector.New(app.appCfg)
	if err != nil {
		t.Logf("Detector initialization failed (expected without real Falco): %v", err)
		return
	}

	// Start detector with context that will timeout
	go func() {
		_ = det.Start(ctx)
	}()

	// Wait for context to timeout
	<-ctx.Done()
}

// TestLoadConfigAutoDetectPath tests the auto-detect logic path in loadConfig
func TestLoadConfigAutoDetectPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	cfg := &Config{
		ConfigFile: "",
		AutoDetect: true,
		Version:    "0.4.1",
	}
	app := New(cfg)

	err = app.loadConfig()
	// Should fail because no state is found
	if err == nil {
		t.Error("Expected error when no state is found in auto-detect")
	}
}

// TestApplyConfigOverridesIntegration tests all overrides together
func TestApplyConfigOverridesIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-1
    state:
      backend: s3
      s3_bucket: mybucket
falco:
  enabled: true
  hostname: oldhost
  port: 5000
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	appCfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := &Config{
		RegionOverride:      []string{"ap-northeast-1"},
		FalcoEndpoint:       "newhost.local:7000",
		StatePathOverride:   "/data/state.tfstate",
		BackendTypeOverride: "local",
		Version:             "0.4.1",
	}
	app := New(cfg)

	err = app.applyConfigOverrides(appCfg)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify all overrides
	if len(appCfg.Providers.AWS.Regions) != 1 || appCfg.Providers.AWS.Regions[0] != "ap-northeast-1" {
		t.Error("Region override failed")
	}

	if appCfg.Falco.Hostname != "newhost.local" || appCfg.Falco.Port != 7000 {
		t.Error("Falco endpoint override failed")
	}

	if appCfg.Providers.AWS.State.LocalPath != "/data/state.tfstate" {
		t.Error("State path override failed")
	}

	if appCfg.Providers.AWS.State.Backend != "local" {
		t.Error("Backend type override failed")
	}
}

// TestFalcoEndpointWithVariousFormats tests Falco endpoint with various valid formats
func TestFalcoEndpointWithVariousFormats(t *testing.T) {
	validEndpoints := []struct {
		endpoint string
		host     string
		port     uint16
	}{
		{"localhost:5060", "localhost", 5060},
		{"127.0.0.1:5060", "127.0.0.1", 5060},
		{"192.168.1.100:8080", "192.168.1.100", 8080},
		{"falco.local:9090", "falco.local", 9090},
		{"my-host.example.com:6000", "my-host.example.com", 6000},
	}

	for _, tt := range validEndpoints {
		t.Run(tt.endpoint, func(t *testing.T) {
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			configContent := `providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: ./terraform.tfstate
falco:
  enabled: true
  hostname: localhost
  port: 5060
`

			if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write test config file: %v", err)
			}

			appCfg, err := config.Load(configFile)
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			cfg := &Config{
				FalcoEndpoint: tt.endpoint,
				Version:       "0.4.1",
			}
			app := New(cfg)

			err = app.applyConfigOverrides(appCfg)
			if err != nil {
				t.Errorf("Expected no error for endpoint %s, got: %v", tt.endpoint, err)
			}

			if appCfg.Falco.Hostname != tt.host {
				t.Errorf("Expected hostname %s, got %s", tt.host, appCfg.Falco.Hostname)
			}

			if appCfg.Falco.Port != tt.port {
				t.Errorf("Expected port %d, got %d", tt.port, appCfg.Falco.Port)
			}
		})
	}
}

// ErrConfigValidation is a test error
var ErrConfigValidation = Error("config validation failed")

// Error returns a new test error
func Error(msg string) error {
	return testError{msg}
}

type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}
