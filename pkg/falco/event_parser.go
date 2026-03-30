// Package falco provides event parsing and processing from Falco.
package falco

import (
	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// parseFalcoOutput parses a Falco output response into a TFDrift event
// Supports AWS CloudTrail, GCP Audit Log, and Azure Activity Log events
func (s *Subscriber) parseFalcoOutput(res *outputs.Response) *types.Event {
	// Handle nil response
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	switch res.Source {
	case "aws_cloudtrail":
		return s.parseAWSEvent(res)
	case "gcpaudit":
		return s.gcpParser.Parse(res)
	case "azure_activity":
		return s.azureParser.Parse(res)
	default:
		log.Debugf("Unknown Falco source: %s", res.Source)
		return nil
	}
}

// parseAWSEvent parses AWS CloudTrail events
func (s *Subscriber) parseAWSEvent(res *outputs.Response) *types.Event {

	// Parse output fields
	fields := res.OutputFields

	// Extract CloudTrail event name
	eventName, ok := fields["ct.name"]
	if !ok || eventName == "" {
		log.Warnf("Missing ct.name in Falco output")
		return nil
	}

	// Extract event source (e.g., "ec2.amazonaws.com", "lambda.amazonaws.com")
	eventSource := getStringField(fields, "ct.src")

	// Check if this is a relevant event for drift detection
	if !s.isRelevantEvent(eventName) {
		return nil
	}

	// Extract resource ID
	resourceID := s.extractResourceID(eventName, fields)
	if resourceID == "" {
		log.Debugf("Could not extract resource ID from event %s", eventName)
		return nil
	}

	// Extract resource type (using eventSource for disambiguation)
	resourceType := s.mapEventToResourceType(eventName, eventSource)

	// Extract user identity
	userIdentity := types.UserIdentity{
		Type:        getStringField(fields, "ct.user.type"),
		PrincipalID: getStringField(fields, "ct.user.principalid"),
		ARN:         getStringField(fields, "ct.user.arn"),
		AccountID:   getStringField(fields, "ct.user.accountid"),
		UserName:    getStringField(fields, "ct.user"),
	}

	// Extract changes based on event type
	changes := s.extractChanges(eventName, fields)

	return &types.Event{
		Provider:     "aws",
		EventName:    eventName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     res,
	}
}

// isRelevantEvent checks if an event is relevant for drift detection
// Uses configuration loaded from YAML for flexibility and maintainability
func (s *Subscriber) isRelevantEvent(eventName string) bool {
	cfg, err := LoadEventConfig()
	if err != nil {
		log.Warnf("Failed to load event config: %v, falling back to deny", err)
		return false
	}
	return cfg.IsRelevantEvent(eventName)
}

// extractResourceID extracts the resource ID from Falco output fields
// Uses configuration loaded from YAML for flexibility and maintainability
func (s *Subscriber) extractResourceID(eventName string, fields map[string]string) string {
	cfg, err := LoadEventConfig()
	if err != nil {
		log.Warnf("Failed to load event config: %v, using default fields", err)
		// Fall back to default fields if config cannot be loaded
		possibleFields := []string{"ct.resource.id", "ct.request.resource"}
		for _, fieldName := range possibleFields {
			if id := getStringField(fields, fieldName); id != "" {
				return id
			}
		}
		return ""
	}

	// Get possible field names for this event from config
	possibleFields := cfg.GetResourceIDFields(eventName)

	// Try each field
	for _, fieldName := range possibleFields {
		if id := getStringField(fields, fieldName); id != "" {
			return id
		}
	}

	return ""
}
