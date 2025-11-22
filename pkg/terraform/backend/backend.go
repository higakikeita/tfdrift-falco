package backend

import (
	"context"
)

// Backend is the interface for Terraform state backends
type Backend interface {
	// Load reads the Terraform state from the backend
	Load(ctx context.Context) ([]byte, error)

	// Name returns the backend name for logging
	Name() string
}

// StateData represents the raw state data and metadata
type StateData struct {
	Data         []byte
	LastModified string
	Version      string
}
