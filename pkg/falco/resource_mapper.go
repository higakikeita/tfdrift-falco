package falco

import (
	log "github.com/sirupsen/logrus"
)

// mapEventToResourceType maps a CloudTrail event name and source to a Terraform resource type
// eventSource examples: "ec2.amazonaws.com", "lambda.amazonaws.com", "kms.amazonaws.com"
// Uses configuration loaded from YAML for flexibility and maintainability
func (s *Subscriber) mapEventToResourceType(eventName string, eventSource string) string {
	cfg, err := LoadEventConfig()
	if err != nil {
		log.Warnf("Failed to load event config: %v, returning unknown", err)
		return "unknown"
	}

	// First, try to resolve conflicts using eventSource
	if resolved := cfg.ResolveEventSourceConflict(eventName, eventSource); resolved != "" {
		return resolved
	}

	// Get resource type from config mapping
	return cfg.GetResourceType(eventName)
}
