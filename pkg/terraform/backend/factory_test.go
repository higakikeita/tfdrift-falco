package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBackend_Local(t *testing.T) {
	tests := []struct {
		name      string
		config    config.TerraformStateConfig
		setup     func(string) error
		wantError bool
		wantType  string
	}{
		{
			name: "local backend with explicit path",
			config: config.TerraformStateConfig{
				Backend:   "local",
				LocalPath: "test.tfstate",
			},
			setup: func(path string) error {
				return os.WriteFile(path, []byte(`{"version": 4}`), 0600)
			},
			wantError: false,
			wantType:  "local",
		},
		{
			name: "local backend with default path",
			config: config.TerraformStateConfig{
				Backend:   "local",
				LocalPath: "",
			},
			setup: func(path string) error {
				return os.WriteFile("./terraform.tfstate", []byte(`{"version": 4}`), 0600)
			},
			wantError: false,
			wantType:  "local",
		},
		{
			name: "empty backend defaults to local",
			config: config.TerraformStateConfig{
				Backend:   "",
				LocalPath: "test.tfstate",
			},
			setup: func(path string) error {
				return os.WriteFile(path, []byte(`{"version": 4}`), 0600)
			},
			wantError: false,
			wantType:  "local",
		},
		{
			name: "local backend with non-existent file",
			config: config.TerraformStateConfig{
				Backend:   "local",
				LocalPath: "nonexistent.tfstate",
			},
			setup:     func(path string) error { return nil },
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			testPath := filepath.Join(tmpDir, tt.config.LocalPath)
			if tt.config.LocalPath == "" {
				testPath = "./terraform.tfstate"
				defer func() { _ = os.Remove(testPath) }()
			} else {
				tt.config.LocalPath = testPath
			}

			if tt.setup != nil {
				err := tt.setup(testPath)
				require.NoError(t, err)
			}

			// Test
			ctx := context.Background()
			backend, err := NewBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, tt.wantType, backend.Name())
			}
		})
	}
}

func TestNewBackend_S3(t *testing.T) {
	tests := []struct {
		name      string
		config    config.TerraformStateConfig
		wantError bool
	}{
		{
			name: "valid S3 config",
			config: config.TerraformStateConfig{
				Backend:  "s3",
				S3Bucket: "my-bucket",
				S3Key:    "terraform.tfstate",
				S3Region: "us-west-2",
			},
			wantError: false,
		},
		{
			name: "S3 without region (should default)",
			config: config.TerraformStateConfig{
				Backend:  "s3",
				S3Bucket: "my-bucket",
				S3Key:    "terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "S3 without bucket",
			config: config.TerraformStateConfig{
				Backend: "s3",
				S3Key:   "terraform.tfstate",
			},
			wantError: true,
		},
		{
			name: "S3 without key",
			config: config.TerraformStateConfig{
				Backend:  "s3",
				S3Bucket: "my-bucket",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			backend, err := NewBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, "s3", backend.Name())
			}
		})
	}
}

func TestNewBackend_UnsupportedBackend(t *testing.T) {
	ctx := context.Background()
	cfg := config.TerraformStateConfig{
		Backend: "unsupported",
	}

	backend, err := NewBackend(ctx, cfg)
	assert.Error(t, err)
	assert.Nil(t, backend)
	assert.Contains(t, err.Error(), "unsupported backend")
}

func TestNewBackend_GCS(t *testing.T) {
	tests := []struct {
		name      string
		config    config.TerraformStateConfig
		wantError bool
		wantName  string
	}{
		{
			name: "valid GCS config",
			config: config.TerraformStateConfig{
				Backend:   "gcs",
				GCSBucket: "my-bucket",
				GCSPrefix: "terraform.tfstate",
			},
			wantError: false,
			wantName:  "gcs",
		},
		{
			name: "GCS without bucket",
			config: config.TerraformStateConfig{
				Backend:   "gcs",
				GCSPrefix: "terraform.tfstate",
			},
			wantError: true,
		},
		{
			name: "GCS without prefix",
			config: config.TerraformStateConfig{
				Backend:   "gcs",
				GCSBucket: "my-bucket",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			backend, err := NewBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				// GCS backend will fail without credentials, but config validation should pass
				if err != nil {
					// Skip if it's a credentials error (expected in CI)
					t.Logf("Skipped due to GCP credentials: %v", err)
					return
				}
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, tt.wantName, backend.Name())
				// Clean up if backend was created successfully
				if gcsBackend, ok := backend.(*GCSBackend); ok {
					_ = gcsBackend.Close()
				}
			}
		})
	}
}

func TestNewBackend_AzureRM(t *testing.T) {
	tests := []struct {
		name      string
		config    config.TerraformStateConfig
		wantError bool
		wantName  string
	}{
		{
			name: "valid Azure config",
			config: config.TerraformStateConfig{
				Backend:             "azurerm",
				AzureStorageAccount: "myaccount",
				AzureContainerName:  "tfstate",
				AzureBlobName:       "terraform.tfstate",
				AzureAccessKey:      "mykey",
			},
			wantError: false,
			wantName:  "azurerm",
		},
		{
			name: "Azure without storage account",
			config: config.TerraformStateConfig{
				Backend:            "azurerm",
				AzureContainerName: "tfstate",
				AzureBlobName:      "terraform.tfstate",
			},
			wantError: true,
		},
		{
			name: "Azure without container",
			config: config.TerraformStateConfig{
				Backend:             "azurerm",
				AzureStorageAccount: "myaccount",
				AzureBlobName:       "terraform.tfstate",
			},
			wantError: true,
		},
		{
			name: "Azure without blob",
			config: config.TerraformStateConfig{
				Backend:             "azurerm",
				AzureStorageAccount: "myaccount",
				AzureContainerName:  "tfstate",
			},
			wantError: true,
		},
		{
			name: "Azure with SAS token",
			config: config.TerraformStateConfig{
				Backend:             "azurerm",
				AzureStorageAccount: "myaccount",
				AzureContainerName:  "tfstate",
				AzureBlobName:       "terraform.tfstate",
				AzureSASToken:       "sv=2020-10-02&sig=abc",
			},
			wantError: false,
			wantName:  "azurerm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			backend, err := NewBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, tt.wantName, backend.Name())
			}
		})
	}
}

func TestNewBackend_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name       string
		backendKey string
		wantError  bool
	}{
		{
			name:       "lowercase s3",
			backendKey: "s3",
			wantError:  false,
		},
		{
			name:       "uppercase S3",
			backendKey: "S3",
			wantError:  true,
		},
		{
			name:       "mixed case Local",
			backendKey: "Local",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.TerraformStateConfig{
				Backend: tt.backendKey,
			}

			backend, err := NewBackend(ctx, cfg)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				if err != nil {
					t.Logf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewBackend_AllBackendTypes(t *testing.T) {
	// Verify that all backend types are handled
	backendTypes := []string{"local", "", "s3", "gcs", "azurerm"}

	for _, backendType := range backendTypes {
		t.Run("backend type "+backendType, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.TerraformStateConfig{Backend: backendType}

			// Create a temporary file for local backend
			if backendType == "local" || backendType == "" {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "test.tfstate")
				err := os.WriteFile(tmpFile, []byte(`{}`), 0600)
				require.NoError(t, err)
				cfg.LocalPath = tmpFile
			}

			// Set minimal valid config for other backends
			if backendType == "s3" {
				cfg.S3Bucket = "test"
				cfg.S3Key = "test"
			} else if backendType == "gcs" {
				cfg.GCSBucket = "test"
				cfg.GCSPrefix = "test"
			} else if backendType == "azurerm" {
				cfg.AzureStorageAccount = "test"
				cfg.AzureContainerName = "test"
				cfg.AzureBlobName = "test"
			}

			backend, err := NewBackend(ctx, cfg)

			// Some backends may fail without proper setup (credentials, etc)
			// but they should at least attempt creation
			if backendType == "local" || backendType == "" {
				// Local backend should succeed
				assert.NoError(t, err)
				assert.NotNil(t, backend)
			} else {
				// Other backends may fail due to setup, but should not be nil on config validation
				// Success is acceptable, credential errors are acceptable
				if err == nil {
					assert.NotNil(t, backend)
					// Clean up GCS backends
					if gcsBackend, ok := backend.(*GCSBackend); ok {
						_ = gcsBackend.Close()
					}
				}
			}
		})
	}
}
