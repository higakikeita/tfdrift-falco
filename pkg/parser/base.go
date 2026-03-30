// Package parser provides event parsing infrastructure.
package parser

import (
	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// EventParserConfig defines the provider-specific parsing callbacks.
type EventParserConfig struct {
	Provider       string // "aws", "gcp", "azure"
	ExpectedSource string // "aws_cloudtrail", "gcpaudit", "azure_activity"

	// ExtractEventName gets the event/method name from output fields
	ExtractEventName func(fields map[string]string) string

	// IsRelevantEvent checks if this event should trigger drift detection
	IsRelevantEvent func(eventName string) bool

	// ExtractResourceID extracts the resource identifier
	ExtractResourceID func(eventName string, fields map[string]string) string

	// MapResourceType maps event name to Terraform resource type
	MapResourceType func(eventName string, fields map[string]string) string

	// ExtractUserIdentity extracts the user/service identity
	ExtractUserIdentity func(fields map[string]string) types.UserIdentity

	// ExtractChanges gets attribute changes from the event
	ExtractChanges func(eventName string, fields map[string]string) map[string]interface{}

	// ExtractMetadata extracts provider-specific metadata (region, project, subscription, etc.)
	// Optional - can be nil
	ExtractMetadata func(eventName string, fields map[string]string) map[string]string
}

// BaseEventParser provides the common parsing flow for all cloud providers.
type BaseEventParser struct {
	config EventParserConfig
}

// NewBaseEventParser creates a new base parser with the given config.
func NewBaseEventParser(config EventParserConfig) *BaseEventParser {
	return &BaseEventParser{config: config}
}

// Parse executes the common parsing flow using provider-specific callbacks.
func (p *BaseEventParser) Parse(res *outputs.Response) *types.Event {
	// Step 1: nil check
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	// Step 2: source validation
	if res.Source != p.config.ExpectedSource {
		return nil
	}

	fields := res.OutputFields

	// Step 3: extract event name
	eventName := p.config.ExtractEventName(fields)
	if eventName == "" {
		log.Warnf("Missing event name in %s output", p.config.Provider)
		return nil
	}

	// Step 4: relevance check
	if !p.config.IsRelevantEvent(eventName) {
		log.Debugf("Event %s is not relevant for drift detection", eventName)
		return nil
	}

	// Step 5: extract resource ID
	resourceID := p.config.ExtractResourceID(eventName, fields)
	if resourceID == "" {
		log.Debugf("Could not extract resource ID for event %s", eventName)
		return nil
	}

	// Step 6: map resource type
	resourceType := p.config.MapResourceType(eventName, fields)
	if resourceType == "" {
		log.Debugf("No resource mapping for event %s", eventName)
		return nil
	}

	// Step 7: extract user identity
	userIdentity := p.config.ExtractUserIdentity(fields)

	// Step 8: extract changes
	changes := p.config.ExtractChanges(eventName, fields)

	// Build event
	event := &types.Event{
		Provider:     p.config.Provider,
		EventName:    eventName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     res,
	}

	// Step 9: extract metadata if available
	if p.config.ExtractMetadata != nil {
		event.Metadata = p.config.ExtractMetadata(eventName, fields)
	}

	return event
}

// GetStringField safely extracts a string field from the output fields map.
// This is a shared utility used by all providers.
func GetStringField(fields map[string]string, key string) string {
	if val, ok := fields[key]; ok {
		return val
	}
	return ""
}
