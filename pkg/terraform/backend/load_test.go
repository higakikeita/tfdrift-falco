package backend

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// mockS3Client implements the s3Client interface for testing
type mockS3Client struct {
	getObjectFunc func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	if m.getObjectFunc != nil {
		return m.getObjectFunc(ctx, input, optFns...)
	}
	return nil, fmt.Errorf("not implemented")
}

// TestS3Backend_Load_MockSuccess tests successful S3 Load
func TestS3Backend_Load_MockSuccess(t *testing.T) {
	stateData := []byte(`{"version": 4, "terraform_version": "1.5.0"}`)

	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "terraform.tfstate",
		region: "us-east-1",
		client: &mockS3Client{
			getObjectFunc: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				if *input.Bucket != "test-bucket" {
					t.Errorf("expected bucket test-bucket, got %s", *input.Bucket)
				}
				if *input.Key != "terraform.tfstate" {
					t.Errorf("expected key terraform.tfstate, got %s", *input.Key)
				}
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(stateData)),
				}, nil
			},
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if !bytes.Equal(data, stateData) {
		t.Errorf("expected %s, got %s", stateData, data)
	}
}

// TestS3Backend_Load_MockGetObjectError tests S3 Load when GetObject fails
func TestS3Backend_Load_MockGetObjectError(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "terraform.tfstate",
		region: "us-east-1",
		client: &mockS3Client{
			getObjectFunc: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return nil, fmt.Errorf("NoSuchKey: The specified key does not exist")
			},
		},
	}

	_, err := backend.Load(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "failed to get object from S3") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// TestS3Backend_Load_MockReadBodyError tests S3 Load when reading body fails
func TestS3Backend_Load_MockReadBodyError(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "terraform.tfstate",
		region: "us-east-1",
		client: &mockS3Client{
			getObjectFunc: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(&errorReader{}),
				}, nil
			},
		},
	}

	_, err := backend.Load(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "failed to read S3 object body") {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

// TestS3Backend_Load_MockLargeState tests S3 Load with large state data
func TestS3Backend_Load_MockLargeState(t *testing.T) {
	// Create a 1MB state file
	largeData := bytes.Repeat([]byte("A"), 1024*1024)

	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "large-state.tfstate",
		region: "eu-west-1",
		client: &mockS3Client{
			getObjectFunc: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader(largeData)),
				}, nil
			},
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(data) != 1024*1024 {
		t.Errorf("expected %d bytes, got %d", 1024*1024, len(data))
	}
}

// TestS3Backend_Load_MockEmptyState tests S3 Load with empty state
func TestS3Backend_Load_MockEmptyState(t *testing.T) {
	backend := &S3Backend{
		bucket: "test-bucket",
		key:    "empty.tfstate",
		region: "us-east-1",
		client: &mockS3Client{
			getObjectFunc: func(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
				return &s3.GetObjectOutput{
					Body: io.NopCloser(bytes.NewReader([]byte{})),
				}, nil
			},
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty data, got %d bytes", len(data))
	}
}

// TestGCSBackend_Load_Success tests successful GCS Load using loadFunc override
func TestGCSBackend_Load_Success(t *testing.T) {
	stateData := []byte(`{"version": 4, "terraform_version": "1.5.0"}`)

	backend := &GCSBackend{
		bucket: "test-bucket",
		prefix: "terraform.tfstate",
		loadFunc: func(ctx context.Context) ([]byte, error) {
			return stateData, nil
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !bytes.Equal(data, stateData) {
		t.Errorf("expected %s, got %s", stateData, data)
	}
}

// TestGCSBackend_Load_Error tests GCS Load when reading fails
func TestGCSBackend_Load_Error(t *testing.T) {
	backend := &GCSBackend{
		bucket: "test-bucket",
		prefix: "terraform.tfstate",
		loadFunc: func(ctx context.Context) ([]byte, error) {
			return nil, fmt.Errorf("storage: object doesn't exist")
		},
	}

	_, err := backend.Load(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !contains(err.Error(), "object doesn't exist") {
		t.Errorf("unexpected error: %s", err.Error())
	}
}

// TestGCSBackend_Load_LargeState tests GCS Load with large state
func TestGCSBackend_Load_LargeState(t *testing.T) {
	largeData := bytes.Repeat([]byte("B"), 2*1024*1024)

	backend := &GCSBackend{
		bucket: "test-bucket",
		prefix: "large.tfstate",
		loadFunc: func(ctx context.Context) ([]byte, error) {
			return largeData, nil
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(data) != 2*1024*1024 {
		t.Errorf("expected %d bytes, got %d", 2*1024*1024, len(data))
	}
}

// TestGCSBackend_Load_EmptyState tests GCS Load with empty state
func TestGCSBackend_Load_EmptyState(t *testing.T) {
	backend := &GCSBackend{
		bucket: "test-bucket",
		prefix: "empty.tfstate",
		loadFunc: func(ctx context.Context) ([]byte, error) {
			return []byte{}, nil
		},
	}

	data, err := backend.Load(context.Background())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("expected empty data, got %d bytes", len(data))
	}
}

// errorReader is a reader that always returns an error
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("read error")
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
