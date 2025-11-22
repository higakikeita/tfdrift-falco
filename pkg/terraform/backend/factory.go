package backend

import (
	"context"
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
)

// NewBackend creates a backend based on configuration
func NewBackend(ctx context.Context, cfg config.TerraformStateConfig) (Backend, error) {
	switch cfg.Backend {
	case "local", "":
		path := cfg.LocalPath
		if path == "" {
			path = "./terraform.tfstate"
		}
		return NewLocalBackend(path)

	case "s3":
		return NewS3Backend(S3BackendConfig{
			Bucket: cfg.S3Bucket,
			Key:    cfg.S3Key,
			Region: cfg.S3Region,
		})

	default:
		return nil, fmt.Errorf("unsupported backend: %s (supported: local, s3)", cfg.Backend)
	}
}
