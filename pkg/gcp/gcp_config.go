package gcp

import (
	"embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed configs/resource_mappings.yaml
var embeddedConfigs embed.FS

// ResourceConfig contains all resource mapping configurations loaded from YAML
type ResourceConfig struct {
	EventToResourceType map[string]string `yaml:"event_to_resource_type"`
}

var (
	globalConfig *ResourceConfig
	configMutex  sync.RWMutex
)

// LoadResourceConfig loads the resource configuration from the embedded YAML file.
// This function is idempotent and thread-safe - it only loads the config once.
func LoadResourceConfig() (*ResourceConfig, error) {
	configMutex.RLock()
	if globalConfig != nil {
		configMutex.RUnlock()
		return globalConfig, nil
	}
	configMutex.RUnlock()

	// Read embedded YAML file
	data, err := embeddedConfigs.ReadFile("configs/resource_mappings.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	// Parse YAML
	var cfg ResourceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse resource config: %w", err)
	}

	// Store in global variable under write lock
	configMutex.Lock()
	globalConfig = &cfg
	configMutex.Unlock()

	return &cfg, nil
}

// GetEventToResourceTypeMap returns the event-to-resource type mapping
func (rc *ResourceConfig) GetEventToResourceTypeMap() map[string]string {
	return rc.EventToResourceType
}

// GetResourceType returns the Terraform resource type for a given event
func (rc *ResourceConfig) GetResourceType(eventName string) string {
	if resourceType, ok := rc.EventToResourceType[eventName]; ok {
		return resourceType
	}
	return ""
}
