// Package falco provides event parsing and processing from Falco.
package falco

import (
	"encoding/json"
	"strings"

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

	// Extract user identity. resolveActor fills in a human "who" for
	// AssumedRole/SSO callers, where ct.user (userName) is empty and the
	// identifier lives in the ARN session name or principalId (#325).
	userIdentity := resolveActor(types.UserIdentity{
		Type:        getStringField(fields, "ct.user.type"),
		PrincipalID: getStringField(fields, "ct.user.principalid"),
		ARN:         getStringField(fields, "ct.user.arn"),
		AccountID:   getStringField(fields, "ct.user.accountid"),
		UserName:    getStringField(fields, "ct.user"),
	})

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

	// Fallback: cloudtrail plugin 0.17.1 does not expose per-service id fields
	// (ct.request.instanceid, ct.request.groupid, …) — only the aggregate
	// ct.request (full requestParameters JSON). Pull the id out of that JSON
	// using the same candidate field names (#362).
	if id := resourceIDFromRequestJSON(fields, possibleFields); id != "" {
		return id
	}

	return ""
}

// resourceIDFromRequestJSON extracts a resource ID from the plugin's aggregate
// ct.request field (the full requestParameters JSON). For each candidate field
// (e.g. "ct.request.instanceid") it strips the prefix and looks the key up
// case-insensitively in the JSON — CloudTrail keys are camelCase (instanceId,
// groupId, bucketName), while the mapping uses lowercase (#362).
func resourceIDFromRequestJSON(fields map[string]string, candidates []string) string {
	raw := getStringField(fields, "ct.request")
	if raw == "" {
		return ""
	}
	var req map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		return ""
	}
	// Index top-level keys case-insensitively once.
	byLower := make(map[string]interface{}, len(req))
	for k, v := range req {
		byLower[strings.ToLower(k)] = v
	}
	for _, cand := range candidates {
		key := cand
		if i := strings.LastIndex(key, "."); i >= 0 {
			key = key[i+1:]
		}
		if v, ok := byLower[strings.ToLower(key)]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
