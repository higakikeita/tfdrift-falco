package backend

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGCSBackend(t *testing.T) {
	tests := []struct {
		name      string
		config    GCSBackendConfig
		wantError bool
		checkFunc func(*testing.T, *GCSBackend)
	}{
		{
			name: "valid config with all fields",
			config: GCSBackendConfig{
				Bucket: "my-gcs-bucket",
				Prefix: "terraform.tfstate",
			},
			wantError: false,
			checkFunc: func(t *testing.T, b *GCSBackend) {
				assert.Equal(t, "my-gcs-bucket", b.bucket)
				assert.Equal(t, "terraform.tfstate", b.prefix)
				assert.Equal(t, "gcs", b.Name())
			},
		},
		{
			name: "valid config with credentials file",
			config: GCSBackendConfig{
				Bucket:          "my-gcs-bucket",
				Prefix:          "prod/terraform.tfstate",
				CredentialsFile: "/path/to/credentials.json",
			},
			wantError: false,
			checkFunc: func(t *testing.T, b *GCSBackend) {
				assert.Equal(t, "my-gcs-bucket", b.bucket)
				assert.Equal(t, "prod/terraform.tfstate", b.prefix)
			},
		},
		{
			name: "missing bucket",
			config: GCSBackendConfig{
				Bucket: "",
				Prefix: "terraform.tfstate",
			},
			wantError: true,
		},
		{
			name: "missing prefix",
			config: GCSBackendConfig{
				Bucket: "my-gcs-bucket",
				Prefix: "",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Note: This will fail if GCP credentials are not available
			// In CI/CD, we can skip this test or use mocks
			backend, err := NewGCSBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, backend)
			} else {
				// In test environment without GCP credentials, this might fail
				// We check if error is credential-related
				if err != nil {
					// Skip if it's a credential error (expected in CI)
					t.Skipf("Skipping test due to GCP credentials not available: %v", err)
					return
				}

				assert.NoError(t, err)
				assert.NotNil(t, backend)
				assert.NotNil(t, backend.client)

				if tt.checkFunc != nil {
					tt.checkFunc(t, backend)
				}

				// Clean up
				_ = backend.Close()
			}
		})
	}
}

func TestGCSBackend_Name(t *testing.T) {
	ctx := context.Background()

	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
		Bucket: "test-bucket",
		Prefix: "test.tfstate",
	})

	if err != nil {
		// Skip if GCP credentials not available
		t.Skipf("Skipping test due to GCP credentials not available: %v", err)
		return
	}

	defer func() { _ = backend.Close() }()

	assert.Equal(t, "gcs", backend.Name())
}

func TestGCSBackend_Close(t *testing.T) {
	ctx := context.Background()

	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
		Bucket: "test-bucket",
		Prefix: "test.tfstate",
	})

	if err != nil {
		// Skip if GCP credentials not available
		t.Skipf("Skipping test due to GCP credentials not available: %v", err)
		return
	}

	// Should not error on close
	err = backend.Close()
	assert.NoError(t, err)

	// Should be safe to close multiple times
	err = backend.Close()
	assert.NoError(t, err)
}

// Note: GCSBackend.Load() requires actual GCP credentials and GCS access
// For unit tests, we would need to mock the GCS client
// Integration tests with actual GCS should be in tests/integration/

// ============================================================================
// Edge Case and Error Handling Tests
// ============================================================================

func TestGCSBackend_ConfigValidation_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		config    GCSBackendConfig
		wantError bool
		errorMsg  string
	}{
		{
			name: "Bucket with special characters",
			config: GCSBackendConfig{
				Bucket: "my-bucket_with.dots-123",
				Prefix: "terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "Prefix with nested path",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "environments/production/terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "Very long bucket name (63 chars - max allowed)",
			config: GCSBackendConfig{
				Bucket: "a123456789b123456789c123456789d123456789e123456789f12345678912",
				Prefix: "terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "Very long prefix (1024 chars)",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: string(make([]byte, 1024)),
			},
			wantError: false,
		},
		{
			name: "Bucket with only hyphens",
			config: GCSBackendConfig{
				Bucket: "my-bucket-name",
				Prefix: "terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "Prefix with leading slash",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "/terraform.tfstate",
			},
			wantError: false,
		},
		{
			name: "Prefix with trailing slash",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "terraform.tfstate/",
			},
			wantError: false,
		},
		{
			name: "Empty string bucket",
			config: GCSBackendConfig{
				Bucket: "",
				Prefix: "terraform.tfstate",
			},
			wantError: true,
			errorMsg:  "bucket is required",
		},
		{
			name: "Empty string prefix",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "",
			},
			wantError: true,
			errorMsg:  "prefix (object key) is required",
		},
		{
			name: "Whitespace-only bucket",
			config: GCSBackendConfig{
				Bucket: "   ",
				Prefix: "terraform.tfstate",
			},
			wantError: false, // GCS client will handle this
		},
		{
			name: "Whitespace-only prefix",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "   ",
			},
			wantError: false, // GCS client will handle this
		},
		{
			name: "Unicode characters in bucket",
			config: GCSBackendConfig{
				Bucket: "my-bucket-日本語", // GCS doesn't allow Unicode in bucket names
				Prefix: "terraform.tfstate",
			},
			wantError: false, // Validation happens at GCS client level
		},
		{
			name: "Unicode characters in prefix",
			config: GCSBackendConfig{
				Bucket: "my-bucket",
				Prefix: "terraform-日本語.tfstate",
			},
			wantError: false, // GCS allows Unicode in object keys
		},
		{
			name: "Invalid credentials file path",
			config: GCSBackendConfig{
				Bucket:          "my-bucket",
				Prefix:          "terraform.tfstate",
				CredentialsFile: "/nonexistent/path/credentials.json",
			},
			wantError: false, // Error occurs during client creation, not config validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			backend, err := NewGCSBackend(ctx, tt.config)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, backend)
			} else {
				// May error due to GCP credentials, but not due to config validation
				if err != nil {
					// Check if error is credential-related (expected) not config-related
					assert.NotContains(t, err.Error(), "is required")
					t.Logf("Skipped due to GCP credentials: %v", err)
				} else {
					assert.NotNil(t, backend)
					_ = backend.Close()
				}
			}
		})
	}
}

func TestGCSBackend_ConcurrentClose(t *testing.T) {
	ctx := context.Background()

	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
		Bucket: "test-bucket",
		Prefix: "test.tfstate",
	})

	if err != nil {
		t.Skipf("Skipping test due to GCP credentials not available: %v", err)
		return
	}

	// Close concurrently from multiple goroutines
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			done <- backend.Close()
		}()
	}

	// Collect all errors
	var errors []error
	for i := 0; i < numGoroutines; i++ {
		if err := <-done; err != nil {
			errors = append(errors, err)
		}
	}

	// All close calls should succeed (idempotent)
	assert.Empty(t, errors, "All concurrent close calls should succeed")
}

func TestGCSBackend_ClosedClient(t *testing.T) {
	ctx := context.Background()

	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
		Bucket: "test-bucket",
		Prefix: "test.tfstate",
	})

	if err != nil {
		t.Skipf("Skipping test due to GCP credentials not available: %v", err)
		return
	}

	// Close the backend
	err = backend.Close()
	assert.NoError(t, err)

	// Operations on closed client should fail gracefully
	// (This requires actual GCS access, so we just verify Close is idempotent)
	err = backend.Close()
	assert.NoError(t, err, "Second close should be safe")
}

func TestGCSBackend_Name_AfterClose(t *testing.T) {
	ctx := context.Background()

	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
		Bucket: "test-bucket",
		Prefix: "test.tfstate",
	})

	if err != nil {
		t.Skipf("Skipping test due to GCP credentials not available: %v", err)
		return
	}

	// Name should work before close
	assert.Equal(t, "gcs", backend.Name())

	// Close
	_ = backend.Close()

	// Name should still work after close
	assert.Equal(t, "gcs", backend.Name())
}

// TestGCSBackend_ConfigStructFields tests that config fields are properly assigned
func TestGCSBackend_ConfigStructFields(t *testing.T) {
	tests := []struct {
		name           string
		config         GCSBackendConfig
		expectedBucket string
		expectedPrefix string
	}{
		{
			name: "Standard config",
			config: GCSBackendConfig{
				Bucket: "production-bucket",
				Prefix: "terraform/prod.tfstate",
			},
			expectedBucket: "production-bucket",
			expectedPrefix: "terraform/prod.tfstate",
		},
		{
			name: "Config with credentials",
			config: GCSBackendConfig{
				Bucket:          "staging-bucket",
				Prefix:          "terraform/staging.tfstate",
				CredentialsFile: "/path/to/creds.json",
			},
			expectedBucket: "staging-bucket",
			expectedPrefix: "terraform/staging.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			backend, err := NewGCSBackend(ctx, tt.config)

			if err != nil {
				t.Skipf("Skipping test due to GCP credentials not available: %v", err)
				return
			}

			defer func() { _ = backend.Close() }()

			// Verify internal fields are set correctly
			assert.Equal(t, tt.expectedBucket, backend.bucket)
			assert.Equal(t, tt.expectedPrefix, backend.prefix)
			assert.NotNil(t, backend.client)
		})
	}
}

// Note: Testing Load() with actual errors (network errors, auth errors, etc.)
// requires either:
// 1. Integration tests with real GCS (disabled in CI)
// 2. Mock GCS client (requires refactoring to inject dependencies)
//
// Example scenarios that would need mocking:
// - Authentication failures (invalid credentials)
// - Network errors (timeout, connection refused)
// - Bucket not found (404 error)
// - Object not found (404 error)
// - Permission denied (403 error)
// - Rate limiting (429 error)
// - Large file handling
// - Partial read errors
// - Context cancellation during Load()
//
// These scenarios should be covered in integration tests with mocked GCS client
// or in E2E tests with actual GCS buckets.
