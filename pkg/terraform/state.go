package terraform

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform/backend"
	log "github.com/sirupsen/logrus"
)

// StateManager manages Terraform state
type StateManager struct {
	cfg       config.TerraformStateConfig
	resources map[string]*Resource
	mu        sync.RWMutex
}

// Resource represents a Terraform resource
type Resource struct {
	Mode       string                 `json:"mode"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Provider   string                 `json:"provider"`
	Attributes map[string]interface{} `json:"attributes"`
}

// State represents a Terraform state file
type State struct {
	Version          int                  `json:"version"`
	TerraformVersion string               `json:"terraform_version"`
	Resources        []ResourceDefinition `json:"resources"`
}

// ResourceDefinition represents a resource in the state file
type ResourceDefinition struct {
	Mode      string             `json:"mode"`
	Type      string             `json:"type"`
	Name      string             `json:"name"`
	Provider  string             `json:"provider"`
	Instances []ResourceInstance `json:"instances"`
}

// ResourceInstance represents an instance of a resource
type ResourceInstance struct {
	Attributes map[string]interface{} `json:"attributes"`
}

// NewStateManager creates a new StateManager
func NewStateManager(cfg config.TerraformStateConfig) (*StateManager, error) {
	return &StateManager{
		cfg:       cfg,
		resources: make(map[string]*Resource),
	}, nil
}

// Load loads the Terraform state
func (sm *StateManager) Load(ctx context.Context) error {
	// Create backend
	be, err := backend.NewBackend(ctx, sm.cfg)
	if err != nil {
		return fmt.Errorf("failed to create backend: %w", err)
	}

	log.Infof("Loading state from %s backend", be.Name())

	// Load state data
	data, err := be.Load(ctx)
	if err != nil {
		return fmt.Errorf("failed to load state from %s backend: %w", be.Name(), err)
	}

	// Parse state
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	return sm.indexState(state)
}

// indexState indexes the state resources for quick lookup
func (sm *StateManager) indexState(state State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.resources = make(map[string]*Resource)

	for _, resDef := range state.Resources {
		for _, instance := range resDef.Instances {
			resource := &Resource{
				Mode:       resDef.Mode,
				Type:       resDef.Type,
				Name:       resDef.Name,
				Provider:   resDef.Provider,
				Attributes: instance.Attributes,
			}

			// Generate resource ID based on attributes
			resourceID := sm.extractResourceID(resource)
			if resourceID != "" {
				sm.resources[resourceID] = resource
			}
		}
	}

	log.Infof("Indexed %d resources from Terraform state", len(sm.resources))
	return nil
}

// extractResourceID extracts a unique resource ID from attributes
func (sm *StateManager) extractResourceID(resource *Resource) string {
	// Try to get ID from attributes
	if id, ok := resource.Attributes["id"].(string); ok {
		return id
	}

	// Fallback to ARN for AWS resources
	if arn, ok := resource.Attributes["arn"].(string); ok {
		return arn
	}

	// Fallback to name
	if name, ok := resource.Attributes["name"].(string); ok {
		return name
	}

	return ""
}

// GetResource retrieves a resource by ID
func (sm *StateManager) GetResource(resourceID string) (*Resource, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	resource, exists := sm.resources[resourceID]
	return resource, exists
}

// ResourceCount returns the number of resources in the state
func (sm *StateManager) ResourceCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.resources)
}

// Refresh reloads the Terraform state
func (sm *StateManager) Refresh(ctx context.Context) error {
	log.Info("Refreshing Terraform state...")
	return sm.Load(ctx)
}
