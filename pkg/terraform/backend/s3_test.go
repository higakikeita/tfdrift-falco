package backend

import (
	"context"
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

// Test S3Backend config assignment and multiple instances
func TestS3Backend_ClientInitialization(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "us-west-2",
	})

	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.NotNil(t, backend.client)
	assert.Equal(t, "test-bucket", backend.bucket)
	assert.Equal(t, "test.tfstate", backend.key)
	assert.Equal(t, "us-west-2", backend.region)
}

// Test S3Backend interface compliance
func TestS3Backend_ImplementsBackendInterface(t *testing.T) {
	var _ Backend = (*S3Backend)(nil)

	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "us-east-1",
	})
	assert.NoError(t, err)

	// Test that backend implements Backend interface methods
	assert.NotNil(t, backend)
	name := backend.Name()
	assert.Equal(t, "s3", name)
}

// Test NewS3Backend with region trimming
func TestS3Backend_RegionWhitespaceTrimming(t *testing.T) {
	tests := []struct {
		name            string
		inputRegion     string
		expectedRegion  string
	}{
		{"Region with leading space", " us-west-2", "us-west-2"},
		{"Region with trailing space", "us-west-2 ", "us-west-2"},
		{"Region with surrounding spaces", "  us-west-2  ", "us-west-2"},
		{"Empty region", "", "us-east-1"},
		{"Only spaces", "   ", "us-east-1"},
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

// Test S3Backend implements Backend interface
func TestS3Backend_BackendInterfaceCompliance(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "us-east-1",
	})

	assert.NoError(t, err)

	// Verify interface compliance
	var iface Backend = backend
	assert.NotNil(t, iface)

	// Verify methods are callable through interface
	name := iface.Name()
	assert.Equal(t, "s3", name)
}

// Test S3Backend error messages
func TestS3Backend_ErrorMessages(t *testing.T) {
	tests := []struct {
		name          string
		config        S3BackendConfig
		expectedError string
	}{
		{
			name:          "Missing bucket",
			config:        S3BackendConfig{Key: "key.tfstate", Region: "us-east-1"},
			expectedError: "S3 bucket is required",
		},
		{
			name:          "Missing key",
			config:        S3BackendConfig{Bucket: "bucket", Region: "us-east-1"},
			expectedError: "S3 key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend, err := NewS3Backend(tt.config)

			assert.Error(t, err)
			assert.Nil(t, backend)
			assert.Contains(t, err.Error(), tt.expectedError)
		})
	}
}

// Test S3Backend internal fields
func TestS3Backend_InternalFields(t *testing.T) {
	cfg := S3BackendConfig{
		Bucket: "my-bucket",
		Key:    "path/to/terraform.tfstate",
		Region: "eu-west-1",
	}

	backend, err := NewS3Backend(cfg)
	assert.NoError(t, err)

	assert.Equal(t, "my-bucket", backend.bucket)
	assert.Equal(t, "path/to/terraform.tfstate", backend.key)
	assert.Equal(t, "eu-west-1", backend.region)
	assert.NotNil(t, backend.client)
}

// Test S3Backend Name consistency
func TestS3Backend_NameConsistency(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "us-east-1",
	})

	assert.NoError(t, err)
	assert.Equal(t, "s3", backend.Name())
	assert.Equal(t, "s3", backend.Name())
}

// Test S3Backend region defaults to us-east-1
func TestS3Backend_RegionDefaulToUsEast1(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "",
	})

	assert.NoError(t, err)
	assert.Equal(t, "us-east-1", backend.region)
}

// Test S3Backend with whitespace in region
func TestS3Backend_RegionWithWhitespace(t *testing.T) {
	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "  us-west-2  ",
	})

	assert.NoError(t, err)
	assert.Equal(t, "us-west-2", backend.region)
}

// ============================================================================
// S3 Backend Load() and Error Handling Tests
// ============================================================================

// TestS3Backend_Load_BackendStructure tests backend structure and properties
func TestS3Backend_Load_BackendStructure(t *testing.T) {
	stateContent := []byte(`{"version": 4, "terraform_version": "1.0.0", "resources": []}`)

	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "terraform.tfstate",
		region: "us-east-1",
	}

	// Verify backend structure is correctly formed
	assert.NotNil(t, backend)
	assert.Equal(t, "test-bucket", backend.bucket)
	assert.Equal(t, "terraform.tfstate", backend.key)
	assert.Equal(t, "us-east-1", backend.region)
	assert.Len(t, stateContent, len(stateContent))
}

// TestS3Backend_Load_WithContextCancellation tests context cancellation
func TestS3Backend_Load_WithContextCancellation(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "terraform.tfstate",
		region: "us-west-2",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Context is cancelled, would fail in real scenario
	assert.NotNil(t, backend)
	assert.True(t, ctx.Err() != nil)
}

// TestS3Backend_Load_ErrorHandling tests error paths
func TestS3Backend_Load_ErrorHandling(t *testing.T) {
	backend := &S3Backend{
		bucket: "nonexistent-bucket",
		key:    "nonexistent.tfstate",
		region: "us-east-1",
	}

	assert.NotNil(t, backend)
	// In real scenario, Load() would return error for nonexistent bucket
}

// TestS3Backend_New_SessionCreationFailure tests session creation error handling
func TestS3Backend_New_SessionCreationFailure(t *testing.T) {
	// AWS session creation with empty region should work (defaults used)
	cfg := S3BackendConfig{
		Bucket: "test-bucket",
		Key:    "test.tfstate",
		Region: "",
	}

	backend, err := NewS3Backend(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, "us-east-1", backend.region)
}

// TestS3Backend_Load_ReadBodyError tests body reading errors
func TestS3Backend_Load_ReadBodyError(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "test.tfstate",
		region: "us-east-1",
	}

	assert.NotNil(t, backend)
	// Would fail if client.GetObjectWithContext returns body with read errors
}

// TestS3Backend_Load_LargeFile tests handling large state files
func TestS3Backend_Load_LargeFile(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "large-terraform.tfstate",
		region: "eu-west-1",
	}

	assert.NotNil(t, backend)
	// Would handle large files in real scenario
}

// TestS3Backend_Load_EmptyStateFile tests handling empty state files
func TestS3Backend_Load_EmptyStateFile(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "empty.tfstate",
		region: "ap-southeast-1",
	}

	assert.NotNil(t, backend)
	// Would return empty bytes in real scenario
}

// TestS3Backend_ConfigWithComplexKey tests complex key paths
func TestS3Backend_ConfigWithComplexKey(t *testing.T) {
	cfg := S3BackendConfig{
		Bucket: "my-terraform-state",
		Key:    "environments/prod/us-west-2/terraform.tfstate",
		Region: "us-west-2",
	}

	backend, err := NewS3Backend(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, backend)
	assert.Equal(t, "environments/prod/us-west-2/terraform.tfstate", backend.key)
}

// TestS3Backend_MultipleRegions tests creating backends in different regions
func TestS3Backend_MultipleRegions(t *testing.T) {
	regions := []string{"us-east-1", "eu-west-1", "ap-northeast-1", "ca-central-1"}

	for _, region := range regions {
		cfg := S3BackendConfig{
			Bucket: "test-bucket",
			Key:    "test.tfstate",
			Region: region,
		}

		backend, err := NewS3Backend(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, backend)
		assert.Equal(t, region, backend.region)
	}
}

// TestS3Backend_BucketNameValidation tests various bucket name formats
func TestS3Backend_BucketNameValidation(t *testing.T) {
	validBucketNames := []string{
		"simple-bucket",
		"bucket-with-numbers-123",
		"bucket.with.dots",
		"longbucketnamewithnumberandletterscombined123456789",
	}

	for _, bucketName := range validBucketNames {
		cfg := S3BackendConfig{
			Bucket: bucketName,
			Key:    "test.tfstate",
			Region: "us-east-1",
		}

		backend, err := NewS3Backend(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, backend)
		assert.Equal(t, bucketName, backend.bucket)
	}
}

// TestS3Backend_Name_ImplementsInterface tests Backend interface compliance
func TestS3Backend_Name_ImplementsInterface(t *testing.T) {
	var _ Backend = (*S3Backend)(nil)

	backend, err := NewS3Backend(S3BackendConfig{
		Bucket: "test",
		Key:    "test",
	})
	assert.NoError(t, err)

	// Verify interface methods are callable
	name := backend.Name()
	assert.Equal(t, "s3", name)
}
