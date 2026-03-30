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

// Test LocalBackend interface compliance
func TestLocalBackend_ImplementsBackendInterface(t *testing.T) {
	var _ Backend = (*LocalBackend)(nil)

	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)
	assert.NotNil(t, backend)
}

// Test LocalBackend initialization with valid path
func TestLocalBackend_InitializationWithValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	content := `{"version": 4}`
	err := os.WriteFile(statePath, []byte(content), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, statePath, backend.path)
	assert.Equal(t, "local", backend.Name())
}

// Test LocalBackend with various file permissions
func TestLocalBackend_FilePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions os.FileMode
		wantError   bool
	}{
		{
			name:        "standard permissions",
			permissions: 0644,
			wantError:   false,
		},
		{
			name:        "restricted permissions",
			permissions: 0600,
			wantError:   false,
		},
		{
			name:        "read-only",
			permissions: 0444,
			wantError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			statePath := filepath.Join(tmpDir, "terraform.tfstate")
			err := os.WriteFile(statePath, []byte(`{"version": 4}`), tt.permissions)
			require.NoError(t, err)

			backend, err := NewLocalBackend(statePath)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
			}
		})
	}
}

// Test LocalBackend Load with JSON parsing
func TestLocalBackend_Load_ComplexJSON(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")

	complexState := `{
		"version": 4,
		"terraform_version": "1.5.0",
		"serial": 42,
		"lineage": "abc-123",
		"outputs": {
			"instance_id": {
				"value": "i-1234567890abcdef0"
			}
		},
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"instances": [
					{
						"attributes": {
							"id": "i-1234567890abcdef0",
							"instance_type": "t3.micro"
						}
					}
				]
			}
		]
	}`

	err := os.WriteFile(statePath, []byte(complexState), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := backend.Load(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, data)
	assert.Equal(t, []byte(complexState), data)
}

// Test LocalBackend Load with binary data
func TestLocalBackend_Load_BinaryContent(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")

	// Create file with various byte values
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	err := os.WriteFile(statePath, binaryData, 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := backend.Load(ctx)

	assert.NoError(t, err)
	assert.Equal(t, binaryData, data)
}

// Test LocalBackend with deeply nested path
func TestLocalBackend_DeeplyNestedPath(t *testing.T) {
	tmpDir := t.TempDir()
	deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e")
	err := os.MkdirAll(deepPath, 0755)
	require.NoError(t, err)

	statePath := filepath.Join(deepPath, "terraform.tfstate")
	err = os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, statePath, backend.path)
}

// Test LocalBackend Load context usage
func TestLocalBackend_Load_ContextBehavior(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	// Test with background context
	ctx := context.Background()
	data, err := backend.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// The context parameter is not actually used in Load, but test it still works
	// when context is cancelled
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should still work because LocalBackend doesn't use context for file I/O
	data, err = backend.Load(cancelCtx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

// Test LocalBackend with symlink
func TestLocalBackend_SymlinkFile(t *testing.T) {
	tmpDir := t.TempDir()
	actualPath := filepath.Join(tmpDir, "actual.tfstate")
	linkPath := filepath.Join(tmpDir, "link.tfstate")

	// Create actual file
	err := os.WriteFile(actualPath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	// Create symlink
	err = os.Symlink(actualPath, linkPath)
	if err != nil {
		t.Skip("Symlink creation not supported on this platform")
	}

	backend, err := NewLocalBackend(linkPath)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := backend.Load(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, data)
}

// Test LocalBackend Name method
func TestLocalBackend_NameMethod(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	assert.Equal(t, "local", backend.Name())
	// Name should always return the same value
	assert.Equal(t, "local", backend.Name())
}

// Test LocalBackend error messages
func TestLocalBackend_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectedError string
	}{
		{
			name:          "Non-existent file",
			path:          "/nonexistent/path/terraform.tfstate",
			expectedError: "state file not found",
		},
		{
			name:          "Empty path with no default file",
			path:          "",
			expectedError: "state file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewLocalBackend(tt.path)

			assert.Error(t, err)
			assert.Nil(t, backend)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// Test LocalBackend path storage
func TestLocalBackend_PathStorage(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err := os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	// Verify path is stored correctly
	assert.Equal(t, statePath, backend.path)
}

// Test LocalBackend Load with large file
func TestLocalBackend_Load_LargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")

	// Create a 1MB file
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = 'a' + byte(i%26)
	}

	err := os.WriteFile(statePath, largeContent, 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend(statePath)
	require.NoError(t, err)

	ctx := context.Background()
	data, err := backend.Load(ctx)

	assert.NoError(t, err)
	assert.Equal(t, largeContent, data)
	assert.Equal(t, 1024*1024, len(data))
}

// Test LocalBackend default path
func TestLocalBackend_DefaultPath(t *testing.T) {
	// Create default terraform.tfstate in current directory
	tmpDir := t.TempDir()
	oldCwd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldCwd)

	os.Chdir(tmpDir)

	// Create terraform.tfstate
	err = os.WriteFile("terraform.tfstate", []byte(`{"version": 4}`), 0600)
	require.NoError(t, err)

	backend, err := NewLocalBackend("")
	require.NoError(t, err)
	assert.NotNil(t, backend)
}

// Test LocalBackend with various state file formats
func TestLocalBackend_Load_ValidStateFormats(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "minimal state",
			content: `{"version": 4}`,
		},
		{
			name:    "state with outputs",
			content: `{"version": 4, "outputs": {"instance_id": {"value": "i-12345"}}}`,
		},
		{
			name:    "state with resources",
			content: `{"version": 4, "resources": [{"type": "aws_instance", "name": "example"}]}`,
		},
		{
			name:    "empty content",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statePath := filepath.Join(tmpDir, "state_"+tt.name+".tfstate")
			err := os.WriteFile(statePath, []byte(tt.content), 0600)
			require.NoError(t, err)

			backend, err := NewLocalBackend(statePath)
			require.NoError(t, err)

			data, err := backend.Load(context.Background())
			assert.NoError(t, err)
			assert.Equal(t, []byte(tt.content), data)
		})
	}
}

// Test LocalBackend with invalid paths
func TestLocalBackend_InvalidPaths(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"absolute path to nonexistent", "/tmp/nonexistent/terraform.tfstate"},
		{"relative path nonexistent", "nonexistent/terraform.tfstate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewLocalBackend(tt.path)
			assert.Error(t, err)
			assert.Nil(t, backend)
		})
	}
}

// Test LocalBackend Name consistency
func TestLocalBackend_NameConsistency(t *testing.T) {
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	os.WriteFile(statePath, []byte(`{"version": 4}`), 0600)

	backend, _ := NewLocalBackend(statePath)

	assert.Equal(t, "local", backend.Name())
	assert.Equal(t, "local", backend.Name())
	assert.Equal(t, "local", backend.Name())
}
