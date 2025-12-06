package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStateManager(t *testing.T) {
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "./terraform.tfstate",
	}

	sm, err := NewStateManager(cfg)

	require.NoError(t, err)
	require.NotNil(t, sm)
	assert.Equal(t, "local", sm.cfg.Backend)
	assert.Equal(t, "./terraform.tfstate", sm.cfg.LocalPath)
	assert.NotNil(t, sm.resources)
}

func TestStateManager_LoadLocal_Simple(t *testing.T) {
	path := filepath.Join("testdata", "simple.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, 3, sm.ResourceCount())

	// Check EC2 instance
	resource, exists := sm.GetResource("i-1234567890abcdef0")
	require.True(t, exists)
	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "web", resource.Name)
	assert.Equal(t, "managed", resource.Mode)

	// Check S3 bucket
	resource, exists = sm.GetResource("my-data-bucket")
	require.True(t, exists)
	assert.Equal(t, "aws_s3_bucket", resource.Type)
	assert.Equal(t, "data", resource.Name)

	// Check Security Group
	resource, exists = sm.GetResource("sg-0123456789abcdef0")
	require.True(t, exists)
	assert.Equal(t, "aws_security_group", resource.Type)
	assert.Equal(t, "web_sg", resource.Name)
}

func TestStateManager_LoadLocal_Empty(t *testing.T) {
	path := filepath.Join("testdata", "empty.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, 0, sm.ResourceCount())
}

func TestStateManager_LoadLocal_FileNotFound(t *testing.T) {
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "nonexistent.tfstate",
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	// The backend creation should fail when file doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "state file not found")
}

func TestStateManager_LoadLocal_InvalidJSON(t *testing.T) {
	path := filepath.Join("testdata", "invalid.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse state file")
}

func TestStateManager_LoadLocal_DefaultPath(t *testing.T) {
	// Create a temporary terraform.tfstate in current directory
	tmpState := "terraform.tfstate"
	content := `{
		"version": 4,
		"terraform_version": "1.5.0",
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "test",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"attributes": {
							"id": "i-test123"
						}
					}
				]
			}
		]
	}`
	err := os.WriteFile(tmpState, []byte(content), 0600)
	require.NoError(t, err)
	defer func() { _ = os.Remove(tmpState) }()

	// Test with empty LocalPath (should use default)
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "",
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, sm.ResourceCount())
}

func TestStateManager_LoadS3_NotImplemented(t *testing.T) {
	cfg := config.TerraformStateConfig{
		Backend:  "s3",
		S3Bucket: "my-bucket",
		S3Key:    "terraform.tfstate",
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	// S3 backend is implemented but will fail without AWS credentials
	// This test verifies the error contains credential issues
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get object from S3")
}

func TestStateManager_LoadUnsupportedBackend(t *testing.T) {
	cfg := config.TerraformStateConfig{
		Backend: "gcs",
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported backend: gcs")
}

func TestStateManager_GetResource(t *testing.T) {
	path := filepath.Join("testdata", "simple.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)
	require.NoError(t, err)

	tests := []struct {
		name        string
		resourceID  string
		shouldExist bool
		wantType    string
	}{
		{
			name:        "EC2 instance by ID",
			resourceID:  "i-1234567890abcdef0",
			shouldExist: true,
			wantType:    "aws_instance",
		},
		{
			name:        "S3 bucket by ID",
			resourceID:  "my-data-bucket",
			shouldExist: true,
			wantType:    "aws_s3_bucket",
		},
		{
			name:        "Security Group by ID",
			resourceID:  "sg-0123456789abcdef0",
			shouldExist: true,
			wantType:    "aws_security_group",
		},
		{
			name:        "Non-existent resource",
			resourceID:  "i-nonexistent",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, exists := sm.GetResource(tt.resourceID)

			assert.Equal(t, tt.shouldExist, exists)
			if tt.shouldExist {
				require.NotNil(t, resource)
				assert.Equal(t, tt.wantType, resource.Type)
			}
		})
	}
}

func TestStateManager_ExtractResourceID(t *testing.T) {
	sm := &StateManager{
		resources: make(map[string]*Resource),
	}

	tests := []struct {
		name       string
		resource   *Resource
		expectedID string
	}{
		{
			name: "Resource with ID",
			resource: &Resource{
				Type: "aws_instance",
				Attributes: map[string]interface{}{
					"id": "i-1234567890abcdef0",
				},
			},
			expectedID: "i-1234567890abcdef0",
		},
		{
			name: "Resource with ARN (no ID)",
			resource: &Resource{
				Type: "aws_iam_role",
				Attributes: map[string]interface{}{
					"arn": "arn:aws:iam::123456789012:role/MyRole",
				},
			},
			expectedID: "arn:aws:iam::123456789012:role/MyRole",
		},
		{
			name: "Resource with name only",
			resource: &Resource{
				Type: "aws_s3_bucket",
				Attributes: map[string]interface{}{
					"name": "my-bucket",
				},
			},
			expectedID: "my-bucket",
		},
		{
			name: "Resource with no identifiable attributes",
			resource: &Resource{
				Type:       "aws_unknown",
				Attributes: map[string]interface{}{},
			},
			expectedID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := sm.extractResourceID(tt.resource)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestStateManager_MultipleInstances(t *testing.T) {
	path := filepath.Join("testdata", "multiple_instances.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)

	require.NoError(t, err)
	assert.Equal(t, 2, sm.ResourceCount())

	// Check first instance
	resource, exists := sm.GetResource("i-0000000000000001")
	require.True(t, exists)
	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "app", resource.Name)

	// Check second instance
	resource, exists = sm.GetResource("i-0000000000000002")
	require.True(t, exists)
	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "app", resource.Name)
}

func TestStateManager_Refresh(t *testing.T) {
	// Create a temporary state file
	tmpDir := t.TempDir()
	statePath := filepath.Join(tmpDir, "terraform.tfstate")

	initialState := `{
		"version": 4,
		"terraform_version": "1.5.0",
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "test",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"attributes": {
							"id": "i-initial"
						}
					}
				]
			}
		]
	}`
	err := os.WriteFile(statePath, []byte(initialState), 0600)
	require.NoError(t, err)

	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: statePath,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, sm.ResourceCount())

	// Update state file
	updatedState := `{
		"version": 4,
		"terraform_version": "1.5.0",
		"resources": [
			{
				"mode": "managed",
				"type": "aws_instance",
				"name": "test",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": [
					{
						"attributes": {
							"id": "i-initial"
						}
					},
					{
						"attributes": {
							"id": "i-new"
						}
					}
				]
			}
		]
	}`
	err = os.WriteFile(statePath, []byte(updatedState), 0600)
	require.NoError(t, err)

	// Refresh state
	err = sm.Refresh(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, sm.ResourceCount())

	// Verify new resource exists
	_, exists := sm.GetResource("i-new")
	assert.True(t, exists)
}

func TestStateManager_ResourceCount(t *testing.T) {
	path := filepath.Join("testdata", "simple.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	// Before loading
	assert.Equal(t, 0, sm.ResourceCount())

	// After loading
	ctx := context.Background()
	err = sm.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, 3, sm.ResourceCount())
}

func TestStateManager_ThreadSafety(t *testing.T) {
	path := filepath.Join("testdata", "simple.tfstate")
	cfg := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: path,
	}

	sm, err := NewStateManager(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	err = sm.Load(ctx)
	require.NoError(t, err)

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			resource, exists := sm.GetResource("i-1234567890abcdef0")
			assert.True(t, exists)
			assert.NotNil(t, resource)
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify resource count is still correct
	assert.Equal(t, 3, sm.ResourceCount())
}

func TestResource_Structure(t *testing.T) {
	resource := Resource{
		Mode:     "managed",
		Type:     "aws_instance",
		Name:     "web",
		Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
		Attributes: map[string]interface{}{
			"id":            "i-test",
			"instance_type": "t3.micro",
		},
	}

	assert.Equal(t, "managed", resource.Mode)
	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "web", resource.Name)
	assert.NotNil(t, resource.Attributes)
	assert.Equal(t, "i-test", resource.Attributes["id"])
}

func TestState_Structure(t *testing.T) {
	state := State{
		Version:          4,
		TerraformVersion: "1.5.0",
		Resources: []ResourceDefinition{
			{
				Mode: "managed",
				Type: "aws_instance",
				Name: "test",
				Instances: []ResourceInstance{
					{
						Attributes: map[string]interface{}{
							"id": "i-test",
						},
					},
				},
			},
		},
	}

	assert.Equal(t, 4, state.Version)
	assert.Equal(t, "1.5.0", state.TerraformVersion)
	assert.Len(t, state.Resources, 1)
	assert.Equal(t, "aws_instance", state.Resources[0].Type)
}
