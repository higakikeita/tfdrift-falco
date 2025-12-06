package backend

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalBackend(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		setup     func(string) error
		wantError bool
	}{
		{
			name: "existing file",
			path: "test.tfstate",
			setup: func(path string) error {
				return os.WriteFile(path, []byte(`{"version": 4}`), 0600)
			},
			wantError: false,
		},
		{
			name:      "non-existent file",
			path:      "nonexistent.tfstate",
			setup:     func(path string) error { return nil },
			wantError: true,
		},
		{
			name: "empty path defaults to terraform.tfstate",
			path: "",
			setup: func(path string) error {
				return os.WriteFile("./terraform.tfstate", []byte(`{"version": 4}`), 0600)
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			var testPath string
			if tt.path == "" {
				// For empty path test, create in current directory temporarily
				testPath = "./terraform.tfstate"
				defer func() { _ = os.Remove(testPath) }()
			} else {
				testPath = filepath.Join(tmpDir, tt.path)
			}

			if tt.setup != nil {
				err := tt.setup(testPath)
				require.NoError(t, err)
			}

			// Test
			backend, err := NewLocalBackend(testPath)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.Equal(t, "local", backend.Name())
			}
		})
	}
}

func TestLocalBackend_Load(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantError bool
	}{
		{
			name:      "valid state file",
			content:   `{"version": 4, "resources": []}`,
			wantError: false,
		},
		{
			name:      "empty state file",
			content:   "",
			wantError: false,
		},
		{
			name:      "large state file",
			content:   string(make([]byte, 10000)),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			statePath := filepath.Join(tmpDir, "terraform.tfstate")
			err := os.WriteFile(statePath, []byte(tt.content), 0600)
			require.NoError(t, err)

			backend, err := NewLocalBackend(statePath)
			require.NoError(t, err)

			// Test
			ctx := context.Background()
			data, err := backend.Load(ctx)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, []byte(tt.content), data)
			}
		})
	}
}

func TestLocalBackend_Load_FileDeleted(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	// Delete file after backend creation
	err = os.Remove(statePath)
	require.NoError(t, err)

	// Test - should fail since file was deleted
	ctx := context.Background()
	data, err := backend.Load(ctx)
	assert.Error(t, err)
	assert.Nil(t, data)
}

func TestLocalBackend_Name(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	assert.Equal(t, "local", backend.Name())
}
