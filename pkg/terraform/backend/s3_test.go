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
