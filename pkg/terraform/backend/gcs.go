package backend

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

// GCSBackend implements Google Cloud Storage backend
type GCSBackend struct {
	bucket string
	prefix string
	client *storage.Client
}

// GCSBackendConfig contains GCS backend configuration
type GCSBackendConfig struct {
	Bucket string
	Prefix string
	// CredentialsFile is optional - if not provided, uses Application Default Credentials
	CredentialsFile string
}

// NewGCSBackend creates a new GCS backend
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

// Load reads the state file from GCS
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

// Name returns the backend name
func (b *GCSBackend) Name() string {
	return "gcs"
}

// Close closes the GCS client
func (b *GCSBackend) Close() error {
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}
