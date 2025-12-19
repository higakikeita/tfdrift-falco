package falco

import (
	"github.com/keitahigaki/tfdrift-falco/pkg/falco/mappings"
)

// mapEventToResourceType maps a CloudTrail event name and source to a Terraform resource type
// eventSource examples: "ec2.amazonaws.com", "lambda.amazonaws.com", "kms.amazonaws.com"
func (s *Subscriber) mapEventToResourceType(eventName string, eventSource string) string {
	// First, try to resolve conflicts using eventSource
	if resolved := mappings.ResolveEventSourceConflict(eventName, eventSource); resolved != "" {
		return resolved
	}

	// Try each category of mappings
	allMappings := []map[string]string{
		mappings.ComputeMappings,
		mappings.NetworkingMappings,
		mappings.StorageAndDatabaseMappings,
		mappings.SecurityMappings,
		mappings.OtherServicesMappings,
	}

	for _, mapping := range allMappings {
		if resourceType, ok := mapping[eventName]; ok {
			return resourceType
		}
	}

	// Event not found in any mapping
	return "unknown"
}
