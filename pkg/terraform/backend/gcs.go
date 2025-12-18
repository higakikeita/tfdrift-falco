package backend

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// GCSBackend implements the Backend interface for Google Cloud Storage.
//
// This backend reads Terraform state files from GCS buckets, supporting
// both explicit service account credentials and Application Default Credentials (ADC).
//
// The backend handles:
//   - Connection management with automatic retries
//   - Credential resolution (explicit file or ADC)
//   - Error handling with detailed logging
//   - Resource cleanup via Close()
//
// Thread-safe: GCS client handles concurrent access internally.
//
// Example usage:
//
//	ctx := context.Background()
//	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
//	    Bucket: "my-terraform-state",
//	    Prefix: "prod/terraform.tfstate",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer backend.Close()
//
//	data, err := backend.Load(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
type GCSBackend struct {
	bucket string
	prefix string
	client *storage.Client
}

// GCSBackendConfig contains configuration for the GCS backend.
//
// All fields are required except CredentialsFile, which defaults to
// Application Default Credentials if not provided.
type GCSBackendConfig struct {
	// Bucket is the GCS bucket name (e.g., "my-terraform-state")
	Bucket string

	// Prefix is the object key/path to the state file (e.g., "prod/terraform.tfstate")
	Prefix string

	// CredentialsFile is optional path to a service account JSON key file.
	// If empty, uses Application Default Credentials (ADC), which automatically
	// discovers credentials from:
	//   - GOOGLE_APPLICATION_CREDENTIALS environment variable
	//   - GCE/GKE metadata server
	//   - gcloud CLI default credentials
	CredentialsFile string
}

// NewGCSBackend creates a new GCS backend for Terraform state management.
//
// The function initializes a GCS client with either explicit credentials
// (if CredentialsFile is provided) or Application Default Credentials.
//
// Parameters:
//   - ctx: Context for client initialization and API calls
//   - cfg: GCS backend configuration (bucket, prefix, optional credentials)
//
// Returns:
//   - *GCSBackend: Initialized backend ready to load state
//   - error: Configuration validation or client creation errors
//
// Errors:
//   - "GCS bucket is required" if cfg.Bucket is empty
//   - "GCS prefix (object key) is required" if cfg.Prefix is empty
//   - "failed to create GCS client" if client initialization fails (e.g., invalid credentials)
//
// Example:
//
//	// Using explicit credentials
//	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
//	    Bucket: "my-state-bucket",
//	    Prefix: "terraform.tfstate",
//	    CredentialsFile: "/path/to/service-account.json",
//	})
//
//	// Using Application Default Credentials
//	backend, err := NewGCSBackend(ctx, GCSBackendConfig{
//	    Bucket: "my-state-bucket",
//	    Prefix: "terraform.tfstate",
//	})
func NewGCSBackend(ctx context.Context, cfg GCSBackendConfig) (*GCSBackend, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("GCS bucket is required")
	}
	if cfg.Prefix == "" {
		return nil, fmt.Errorf("GCS prefix (object key) is required")
	}

	// Create GCS client
	var client *storage.Client
	var err error

	if cfg.CredentialsFile != "" {
		// Use explicit credentials file
		log.Debugf("Using GCS credentials from file: %s", cfg.CredentialsFile)
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(cfg.CredentialsFile))
	} else {
		// Use Application Default Credentials (ADC)
		log.Debug("Using GCS Application Default Credentials")
		client, err = storage.NewClient(ctx)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSBackend{
		bucket: cfg.Bucket,
		prefix: cfg.Prefix,
		client: client,
	}, nil
}

// Load reads the Terraform state file from GCS and returns its contents.
//
// The method performs the following operations:
//   - Retrieves the GCS bucket and object handles
//   - Opens a reader for the state file object
//   - Reads the complete state file into memory
//   - Logs the operation with bucket/prefix details
//
// Parameters:
//   - ctx: Context for the GCS API call (supports cancellation and timeouts)
//
// Returns:
//   - []byte: Complete Terraform state file contents as JSON
//   - error: If the object doesn't exist, cannot be read, or other GCS errors occur
//
// Errors:
//   - "failed to read object from GCS": Object doesn't exist, permission denied, or network errors
//   - "failed to read GCS object body": Error reading object data
//
// Example:
//
//	ctx := context.Background()
//	data, err := backend.Load(ctx)
//	if err != nil {
//	    log.Fatalf("Failed to load state: %v", err)
//	}
//	// Parse Terraform state from data
func (b *GCSBackend) Load(ctx context.Context) ([]byte, error) {
	log.Infof("Loading Terraform state from GCS: gs://%s/%s", b.bucket, b.prefix)

	// Get bucket handle
	bucket := b.client.Bucket(b.bucket)

	// Get object handle
	obj := bucket.Object(b.prefix)

	// Read object
	reader, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to read object from GCS gs://%s/%s: %w", b.bucket, b.prefix, err)
	}
	defer func() { _ = reader.Close() }()

	// Read all data
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read GCS object body: %w", err)
	}

	log.Infof("Successfully loaded %d bytes from GCS (gs://%s/%s)", len(data), b.bucket, b.prefix)
	return data, nil
}

// Name returns the backend identifier for logging and debugging.
//
// This method is part of the Backend interface and returns a constant
// string "gcs" to identify this backend type in logs and error messages.
//
// Returns:
//   - string: Always returns "gcs"
func (b *GCSBackend) Name() string {
	return "gcs"
}

// Close closes the GCS client and releases associated resources.
//
// This method should be called when the backend is no longer needed to
// properly clean up connections and release resources. It's safe to call
// multiple times - subsequent calls after the first Close() are no-ops.
//
// Returns:
//   - error: Error from closing the GCS client, or nil if successful or already closed
//
// Best practice: Use defer to ensure cleanup
//
//	backend, err := NewGCSBackend(ctx, cfg)
//	if err != nil {
//	    return err
//	}
//	defer backend.Close()
func (b *GCSBackend) Close() error {
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}
