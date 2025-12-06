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
