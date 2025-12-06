package backend

import (
	"context"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
)

// LocalBackend implements local file system backend
type LocalBackend struct {
	path string
}

// NewLocalBackend creates a new local backend
func NewLocalBackend(path string) (*LocalBackend, error) {
	if path == "" {
		path = "./terraform.tfstate"
	}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("state file not found: %s: %w", path, err)
	}

	return &LocalBackend{
		path: path,
	}, nil
}

// Load reads the state file from local filesystem
func (b *LocalBackend) Load(_ context.Context) ([]byte, error) {
	log.Infof("Loading Terraform state from local file: %s", b.path)

	data, err := os.ReadFile(b.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %s: %w", b.path, err)
	}

	log.Infof("Successfully loaded %d bytes from %s", len(data), b.path)
	return data, nil
}

// Name returns the backend name
func (b *LocalBackend) Name() string {
	return "local"
}
