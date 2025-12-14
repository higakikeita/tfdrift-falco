package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoDetectTerraformState_NoTerraformDir(t *testing.T) {
	// Create temporary directory without .terraform
	tmpDir := t.TempDir()

	result, err := AutoDetectTerraformState(tmpDir)
	require.NoError(t, err)
	assert.False(t, result.Found)
	assert.Contains(t, result.Message, "No .terraform directory found")
	assert.Equal(t, tmpDir, result.DetectionPath)
}

func TestAutoDetectTerraformState_LocalState(t *testing.T) {
	// Create temporary directory with .terraform and terraform.tfstate
	tmpDir := t.TempDir()

	// Create .terraform directory
	terraformDir := filepath.Join(tmpDir, ".terraform")
	err := os.Mkdir(terraformDir, 0755)
	require.NoError(t, err)

	// Create terraform.tfstate file
	statePath := filepath.Join(tmpDir, "terraform.tfstate")
	err = os.WriteFile(statePath, []byte("{}"), 0644)
	require.NoError(t, err)

	result, err := AutoDetectTerraformState(tmpDir)
	require.NoError(t, err)
	assert.True(t, result.Found)
	assert.Equal(t, "local", result.Backend)
	assert.Equal(t, statePath, result.LocalPath)
	assert.Contains(t, result.Message, "Detected local Terraform state")
}

func TestAutoDetectTerraformState_S3Backend(t *testing.T) {
	// Create temporary directory with .terraform and backend.tf
	tmpDir := t.TempDir()

	// Create .terraform directory
	terraformDir := filepath.Join(tmpDir, ".terraform")
	err := os.Mkdir(terraformDir, 0755)
	require.NoError(t, err)

	// Create backend.tf with S3 configuration
	backendTF := `
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "prod/terraform.tfstate"
    region = "us-west-2"
  }
}
`
	backendPath := filepath.Join(tmpDir, "backend.tf")
	err = os.WriteFile(backendPath, []byte(backendTF), 0644)
	require.NoError(t, err)

	result, err := AutoDetectTerraformState(tmpDir)
	require.NoError(t, err)
	assert.True(t, result.Found)
	assert.Equal(t, "s3", result.Backend)
	assert.Equal(t, "my-terraform-state", result.S3Bucket)
	assert.Equal(t, "prod/terraform.tfstate", result.S3Key)
	assert.Equal(t, "us-west-2", result.S3Region)
	assert.Contains(t, result.Message, "Detected S3 backend")
}

func TestAutoDetectTerraformState_S3Backend_DefaultRegion(t *testing.T) {
	// Create temporary directory with .terraform and backend.tf without region
	tmpDir := t.TempDir()

	// Create .terraform directory
	terraformDir := filepath.Join(tmpDir, ".terraform")
	err := os.Mkdir(terraformDir, 0755)
	require.NoError(t, err)

	// Create backend.tf with S3 configuration (no region)
	backendTF := `
terraform {
  backend "s3" {
    bucket = "my-terraform-state"
    key    = "prod/terraform.tfstate"
  }
}
`
	backendPath := filepath.Join(tmpDir, "backend.tf")
	err = os.WriteFile(backendPath, []byte(backendTF), 0644)
	require.NoError(t, err)

	result, err := AutoDetectTerraformState(tmpDir)
	require.NoError(t, err)
	assert.True(t, result.Found)
	assert.Equal(t, "s3", result.Backend)
	assert.Equal(t, "us-east-1", result.S3Region) // Default region
}

func TestAutoDetectTerraformState_NoStateFound(t *testing.T) {
	// Create temporary directory with .terraform but no state
	tmpDir := t.TempDir()

	// Create .terraform directory
	terraformDir := filepath.Join(tmpDir, ".terraform")
	err := os.Mkdir(terraformDir, 0755)
	require.NoError(t, err)

	result, err := AutoDetectTerraformState(tmpDir)
	require.NoError(t, err)
	assert.False(t, result.Found)
	assert.Contains(t, result.Message, "Terraform initialized but no state found")
}

func TestAutoDetectTerraformState_EmptyPath(t *testing.T) {
	// Test with empty path (should use current directory)
	result, err := AutoDetectTerraformState("")
	require.NoError(t, err)
	assert.NotEmpty(t, result.DetectionPath)
}

func TestDetectS3Backend_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main.tf with S3 backend
	mainTF := `
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}

terraform {
  backend "s3" {
    bucket = "test-bucket"
    key    = "test/terraform.tfstate"
    region = "eu-west-1"
  }
}
`
	mainPath := filepath.Join(tmpDir, "main.tf")
	err := os.WriteFile(mainPath, []byte(mainTF), 0644)
	require.NoError(t, err)

	config, err := detectS3Backend(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, config)
	assert.Equal(t, "test-bucket", config.Bucket)
	assert.Equal(t, "test/terraform.tfstate", config.Key)
	assert.Equal(t, "eu-west-1", config.Region)
}

func TestDetectS3Backend_NoBackend(t *testing.T) {
	tmpDir := t.TempDir()

	// Create main.tf without backend
	mainTF := `
resource "aws_instance" "example" {
  ami           = "ami-12345678"
  instance_type = "t2.micro"
}
`
	mainPath := filepath.Join(tmpDir, "main.tf")
	err := os.WriteFile(mainPath, []byte(mainTF), 0644)
	require.NoError(t, err)

	config, err := detectS3Backend(tmpDir)
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "no S3 backend configuration found")
}

func TestDetectS3Backend_IncompleteConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create backend.tf with incomplete S3 config (missing key)
	backendTF := `
terraform {
  backend "s3" {
    bucket = "test-bucket"
    region = "us-east-1"
  }
}
`
	backendPath := filepath.Join(tmpDir, "backend.tf")
	err := os.WriteFile(backendPath, []byte(backendTF), 0644)
	require.NoError(t, err)

	config, err := detectS3Backend(tmpDir)
	assert.Error(t, err)
	assert.Nil(t, config)
}

func TestCreateAutoConfig_LocalBackend(t *testing.T) {
	result := &AutoDetectResult{
		Found:     true,
		Backend:   "local",
		LocalPath: "/path/to/terraform.tfstate",
	}

	cfg, err := CreateAutoConfig(result)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Providers.AWS.Enabled)
	assert.Equal(t, "local", cfg.Providers.AWS.State.Backend)
	assert.Equal(t, "/path/to/terraform.tfstate", cfg.Providers.AWS.State.LocalPath)
	assert.True(t, cfg.Falco.Enabled)
	assert.Equal(t, "localhost", cfg.Falco.Hostname)
	assert.Equal(t, uint16(5060), cfg.Falco.Port)
}

func TestCreateAutoConfig_S3Backend(t *testing.T) {
	result := &AutoDetectResult{
		Found:    true,
		Backend:  "s3",
		S3Bucket: "my-bucket",
		S3Key:    "my-key",
		S3Region: "ap-northeast-1",
	}

	cfg, err := CreateAutoConfig(result)
	require.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.True(t, cfg.Providers.AWS.Enabled)
	assert.Equal(t, "s3", cfg.Providers.AWS.State.Backend)
	assert.Equal(t, "my-bucket", cfg.Providers.AWS.State.S3Bucket)
	assert.Equal(t, "my-key", cfg.Providers.AWS.State.S3Key)
	assert.Equal(t, "ap-northeast-1", cfg.Providers.AWS.State.S3Region)
	assert.Equal(t, []string{"ap-northeast-1"}, cfg.Providers.AWS.Regions)
}

func TestCreateAutoConfig_NotFound(t *testing.T) {
	result := &AutoDetectResult{
		Found:   false,
		Message: "No state found",
	}

	cfg, err := CreateAutoConfig(result)
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "no Terraform state detected")
}

func TestPrintAutoDetectHelp_Found(t *testing.T) {
	// Test that the function doesn't panic with valid result
	result := &AutoDetectResult{
		Found:     true,
		Backend:   "local",
		LocalPath: "/path/to/terraform.tfstate",
		Message:   "✓ Detected local Terraform state: /path/to/terraform.tfstate",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		PrintAutoDetectHelp(result)
	})
}

func TestPrintAutoDetectHelp_NotFound(t *testing.T) {
	// Test that the function doesn't panic with not found result
	result := &AutoDetectResult{
		Found:   false,
		Message: "No state found",
	}

	// Should not panic
	assert.NotPanics(t, func() {
		PrintAutoDetectHelp(result)
	})
}
