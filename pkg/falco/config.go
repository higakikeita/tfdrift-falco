package falco

import (
	"embed"
	"fmt"
	"sync"

	"gopkg.in/yaml.v3"
)

//go:embed configs/event_mappings.yaml
var embeddedConfigs embed.FS

// EventConfig contains all event mapping configurations loaded from YAML
type EventConfig struct {
	RelevantEvents       []string                     `yaml:"relevant_events"`
	ResourceIDFields     map[string][]string          `yaml:"resource_id_fields"`
	EventToResourceType  map[string]string            `yaml:"event_to_resource_type"`
	EventSourceConflicts map[string]map[string]string `yaml:"event_source_conflicts"`

	// Pre-computed map for O(1) lookup of relevant events
	relevantEventsMap map[string]bool
}

var (
	globalConfig *EventConfig
	configMutex  sync.RWMutex
)

// LoadEventConfig loads the event configuration from the embedded YAML file.
// This function is idempotent and thread-safe - it only loads the config once.
func LoadEventConfig() (*EventConfig, error) {
	configMutex.RLock()
	if globalConfig != nil {
		configMutex.RUnlock()
		return globalConfig, nil
	}
	configMutex.RUnlock()

	// Read embedded YAML file
	data, err := embeddedConfigs.ReadFile("configs/event_mappings.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	// Parse YAML
	var cfg EventConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse event config: %w", err)
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
func (ec *EventConfig) GetRelevantEventsMap() map[string]bool {
	return ec.relevantEventsMap
}

// IsRelevantEvent checks if an event is relevant for drift detection
func (ec *EventConfig) IsRelevantEvent(eventName string) bool {
	return ec.relevantEventsMap[eventName]
}

// GetResourceIDFields returns the ordered list of field paths for a given event
func (ec *EventConfig) GetResourceIDFields(eventName string) []string {
	if fields, ok := ec.ResourceIDFields[eventName]; ok {
		return fields
	}
	// Return default fields if not found
	return []string{"ct.resource.id", "ct.request.resource"}
}

// GetResourceType returns the Terraform resource type for a given event
func (ec *EventConfig) GetResourceType(eventName string) string {
	if resourceType, ok := ec.EventToResourceType[eventName]; ok {
		return resourceType
	}
	return "unknown"
}

// ResolveEventSourceConflict resolves conflicts when multiple AWS services
// use the same CloudTrail event name
func (ec *EventConfig) ResolveEventSourceConflict(eventName string, eventSource string) string {
	if conflicts, ok := ec.EventSourceConflicts[eventName]; ok {
		if resourceType, ok := conflicts[eventSource]; ok {
			return resourceType
		}
	}
	return ""
}
