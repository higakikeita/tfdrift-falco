package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewS3Backend(t *testing.T) {
	tests := []struct {
		name      string
		config    S3BackendConfig
		wantError bool
		checkFunc func(*testing.T, *S3Backend)
	}{
		{
			name: "valid config with all fields",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: false,
			checkFunc: func(t *testing.T, b *S3Backend) {
				assert.Equal(t, "my-bucket", b.bucket)
				assert.Equal(t, "terraform.tfstate", b.key)
				assert.Equal(t, "us-west-2", b.region)
				assert.Equal(t, "s3", b.Name())
			},
		},
		{
			name: "valid config with default region",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate",
				Region: "",
			},
			wantError: false,
			checkFunc: func(t *testing.T, b *S3Backend) {
				assert.Equal(t, "us-east-1", b.region)
			},
		},
		{
			name: "missing bucket",
			config: S3BackendConfig{
				Bucket: "",
				Key:    "terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: true,
		},
		{
			name: "missing key",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "",
				Region: "us-west-2",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewS3Backend(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.NotNil(t, backend.client)
				if tt.checkFunc != nil {
					tt.checkFunc(t, backend)
				}
			}
		})
	}
}

func TestS3Backend_Name(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
	})
	assert.NoError(t, err)
	assert.Equal(t, "s3", backend.Name())
}

// Note: S3Backend.Load() requires actual AWS credentials and S3 access
// For unit tests, we would need to mock the S3 client
// Integration tests with actual S3 should be in tests/integration/

// ============================================================================
// Edge Case and Error Handling Tests
// ============================================================================

func TestS3Backend_ConfigValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		config    S3BackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "Bucket with hyphens and numbers",
			config: S3BackendConfig{
				Bucket: "my-bucket-123",
				Key:    "terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: false,
		},
		{
			name: "Key with nested path",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "env/production/us-west-2/terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: false,
		},
		{
			name: "Very long bucket name (63 chars - max allowed)",
			config: S3BackendConfig{
				Bucket: "a123456789b123456789c123456789d123456789e123456789f12345678912",
				Key:    "terraform.tfstate",
				Region: "us-east-1",
			},
			wantError: false,
		},
		{
			name: "Very long key (1024 chars)",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    string(make([]byte, 1024)),
				Region: "ap-northeast-1",
			},
			wantError: false,
		},
		{
			name: "All valid AWS regions",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate",
				Region: "ap-southeast-2",
			},
			wantError: false,
		},
		{
			name: "Key with leading slash",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "/terraform.tfstate",
				Region: "eu-west-1",
			},
			wantError: false,
		},
		{
			name: "Key with trailing slash",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate/",
				Region: "us-east-2",
			},
			wantError: false,
		},
		{
			name: "Empty bucket",
			config: S3BackendConfig{
				Bucket: "",
				Key:    "terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: true,
			errorMsg:  "bucket is required",
		},
		{
			name: "Empty key",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "",
				Region: "us-west-2",
			},
			wantError: true,
			errorMsg:  "key is required",
		},
		{
			name: "Whitespace-only bucket",
			config: S3BackendConfig{
				Bucket: "   ",
				Key:    "terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: false, // AWS SDK will handle this
		},
		{
			name: "Whitespace-only key",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "   ",
				Region: "us-west-2",
			},
			wantError: false, // AWS SDK will handle this
		},
		{
			name: "Empty region defaults to us-east-1",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate",
				Region: "",
			},
			wantError: false,
		},
		{
			name: "Invalid region format",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform.tfstate",
				Region: "invalid-region-123",
			},
			wantError: false, // AWS SDK will validate
		},
		{
			name: "Bucket with dots (DNS-style)",
			config: S3BackendConfig{
				Bucket: "my.bucket.name",
				Key:    "terraform.tfstate",
				Region: "us-east-1",
			},
			wantError: false,
		},
		{
			name: "Key with special characters",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform_state-2024!@#$%^&*().tfstate",
				Region: "us-east-1",
			},
			wantError: false,
		},
		{
			name: "Unicode characters in key",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "terraform-日本語.tfstate",
				Region: "ap-northeast-1",
			},
			wantError: false, // S3 allows Unicode in keys
		},
		{
			name: "Multiple consecutive slashes in key",
			config: S3BackendConfig{
				Bucket: "my-bucket",
				Key:    "path//to///terraform.tfstate",
				Region: "us-west-2",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewS3Backend(tt.config)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, backend)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.NotNil(t, backend.client)

				// Verify config was set correctly
				assert.Equal(t, tt.config.Bucket, backend.bucket)
				assert.Equal(t, tt.config.Key, backend.key)

				// Check default region handling
				if tt.config.Region == "" {
					assert.Equal(t, "us-east-1", backend.region)
				} else {
					assert.Equal(t, tt.config.Region, backend.region)
				}
			}
		})
	}
}

func TestS3Backend_Name_Consistency(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "us-west-2",
	})
	assert.NoError(t, err)

	// Name should always return "s3"
	assert.Equal(t, "s3", backend.Name())
	assert.Equal(t, "s3", backend.Name()) // Second call should be identical
}

func TestS3Backend_ConfigFieldsAssignment(t *testing.T) {
	tests := []struct {
		name           string
		config         S3BackendConfig
		expectedBucket string
		expectedKey    string
		expectedRegion string
	}{
		{
			name: "Standard config",
			config: S3BackendConfig{
				Bucket: "production-state",
				Key:    "prod/terraform.tfstate",
				Region: "us-west-2",
			},
			expectedBucket: "production-state",
			expectedKey:    "prod/terraform.tfstate",
			expectedRegion: "us-west-2",
		},
		{
			name: "Config with default region",
			config: S3BackendConfig{
				Bucket: "staging-state",
				Key:    "staging/terraform.tfstate",
				Region: "",
			},
			expectedBucket: "staging-state",
			expectedKey:    "staging/terraform.tfstate",
			expectedRegion: "us-east-1",
		},
		{
			name: "Config with different region",
			config: S3BackendConfig{
				Bucket: "eu-state",
				Key:    "terraform.tfstate",
				Region: "eu-central-1",
			},
			expectedBucket: "eu-state",
			expectedKey:    "terraform.tfstate",
			expectedRegion: "eu-central-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewS3Backend(tt.config)
			assert.NoError(t, err)
			assert.NotNil(t, backend)

			// Verify internal fields
			assert.Equal(t, tt.expectedBucket, backend.bucket)
			assert.Equal(t, tt.expectedKey, backend.key)
			assert.Equal(t, tt.expectedRegion, backend.region)
		})
	}
}

func TestS3Backend_MultipleInstances(t *testing.T) {
	// Create multiple backend instances with different configs
	configs := []S3BackendConfig{
		{Bucket: "bucket-1", Key: "state-1.tfstate", Region: "us-east-1"},
		{Bucket: "bucket-2", Key: "state-2.tfstate", Region: "us-west-2"},
		{Bucket: "bucket-3", Key: "state-3.tfstate", Region: "eu-west-1"},
	}

	var backends []*S3Backend
	for _, cfg := range configs {
		backend, err := NewS3Backend(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, backend)
		backends = append(backends, backend)
	}

	// Verify each instance is independent
	for i, backend := range backends {
		assert.Equal(t, configs[i].Bucket, backend.bucket)
		assert.Equal(t, configs[i].Key, backend.key)
		assert.Equal(t, configs[i].Region, backend.region)
	}
}

func TestS3Backend_RegionDefaults(t *testing.T) {
	tests := []struct {
		name           string
		inputRegion    string
		expectedRegion string
	}{
		{"Empty region defaults to us-east-1", "", "us-east-1"},
		{"Explicit us-east-1", "us-east-1", "us-east-1"},
		{"us-west-2", "us-west-2", "us-west-2"},
		{"ap-northeast-1", "ap-northeast-1", "ap-northeast-1"},
		{"eu-central-1", "eu-central-1", "eu-central-1"},
		{"Whitespace region defaults to us-east-1", "   ", "us-east-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewS3Backend(S3BackendConfig{
				Bucket: "test-bucket",
				Key:    "test.tfstate",
				Region: tt.inputRegion,
			})
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedRegion, backend.region)
		})
	}
}

// Note: Testing Load() with actual errors (network errors, auth errors, etc.)
// requires either:
// 1. Integration tests with real S3 (requires AWS credentials)
// 2. Mock S3 client (requires refactoring to inject dependencies)
//
// Example scenarios that would need mocking:
// - Authentication failures (invalid AWS credentials)
// - Network errors (timeout, connection refused)
// - Bucket not found (NoSuchBucket error)
// - Object not found (NoSuchKey error)
// - Permission denied (AccessDenied error)
// - Rate limiting (SlowDown error)
// - Large file handling (multi-part downloads)
// - Partial read errors
// - Context cancellation during Load()
// - S3 server errors (500, 503)
// - Encryption errors (KMS key access)
// - Bucket versioning scenarios
//
// These scenarios should be covered in integration tests with mocked S3 client
// or in E2E tests with actual S3 buckets.
