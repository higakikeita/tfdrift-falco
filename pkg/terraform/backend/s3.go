package backend

import (
	"context"
	"fmt"
	"io"

	// TODO: Migrate to aws-sdk-go-v2 (aws-sdk-go-v1 deprecated, EOL July 31, 2025)
	// See: https://aws.amazon.com/blogs/developer/announcing-end-of-support-for-aws-sdk-for-go-v1-on-july-31-2025/
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

// S3Backend implements AWS S3 backend
type S3Backend struct {
	bucket string
	key    string
	region string
	client *s3.S3
}

// S3BackendConfig contains S3 backend configuration
type S3BackendConfig struct {
	Bucket string
	Key    string
	Region string
}

// NewS3Backend creates a new S3 backend
func NewS3Backend(cfg S3BackendConfig) (*S3Backend, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("S3 bucket is required")
	}
	if cfg.Key == "" {
		return nil, fmt.Errorf("S3 key is required")
	}

	// Default region if not specified
	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &S3Backend{
		bucket: cfg.Bucket,
		key:    cfg.Key,
		region: region,
		client: s3.New(sess),
	}, nil
}

// Load reads the state file from S3
func (b *S3Backend) Load(ctx context.Context) ([]byte, error) {
	log.Infof("Loading Terraform state from S3: s3://%s/%s (region: %s)", b.bucket, b.key, b.region)

	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.key),
	}

	result, err := b.client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3 s3://%s/%s: %w", b.bucket, b.key, err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read S3 object body: %w", err)
	}

	log.Infof("Successfully loaded %d bytes from S3 (s3://%s/%s)", len(data), b.bucket, b.key)
	return data, nil
}

// Name returns the backend name
func (b *S3Backend) Name() string {
	return "s3"
}
