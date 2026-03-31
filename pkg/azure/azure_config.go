package azure

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
	RelevantEvents      []string          `yaml:"relevant_events"`
	EventToResourceType map[string]string `yaml:"event_to_resource_type"`
	relevantEventsMap   map[string]bool
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

	// Build the relevant events map for O(1) lookup
	cfg.relevantEventsMap = make(map[string]bool, len(cfg.RelevantEvents))
	for _, event := range cfg.RelevantEvents {
		cfg.relevantEventsMap[event] = true
	}

	// Store in global variable under write lock
	configMutex.Lock()
	globalConfig = &cfg
	configMutex.Unlock()

	return &cfg, nil
}

// GetRelevantEventsMap returns a map for O(1) lookup of relevant events
func (rc *ResourceConfig) GetRelevantEventsMap() map[string]bool {
	return rc.relevantEventsMap
}

// IsRelevantEvent checks if an operation is relevant for drift detection
func (rc *ResourceConfig) IsRelevantEvent(operationName string) bool {
	return rc.relevantEventsMap[operationName]
}

// GetEventToResourceTypeMap returns the event-to-resource type mapping
func (rc *ResourceConfig) GetEventToResourceTypeMap() map[string]string {
	return rc.EventToResourceType
}

// GetResourceType returns the Terraform resource type for a given event
func (rc *ResourceConfig) GetResourceType(operationName string) string {
	if resourceType, ok := rc.EventToResourceType[operationName]; ok {
		return resourceType
	}
	return ""
}
