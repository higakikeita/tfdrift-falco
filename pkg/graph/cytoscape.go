package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// ConvertDriftToCytoscape converts a drift alert to Cytoscape node
func ConvertDriftToCytoscape(drift types.DriftAlert) models.CytoscapeNode {
	return models.CytoscapeNode{
		Data: models.NodeData{
			ID:           drift.ResourceID,
			Label:        drift.ResourceName,
			Type:         "drift",
			ResourceType: drift.ResourceType,
			ResourceName: drift.ResourceName,
			Severity:     drift.Severity,
			Metadata: map[string]interface{}{
				"attribute":     drift.Attribute,
				"old_value":     drift.OldValue,
				"new_value":     drift.NewValue,
				"user":          drift.UserIdentity.UserName,
				"user_arn":      drift.UserIdentity.ARN,
				"matched_rules": drift.MatchedRules,
				"timestamp":     drift.Timestamp,
				"alert_type":    drift.AlertType,
			},
		},
	}
}

// ConvertEventToCytoscape converts a Falco event to Cytoscape node
func ConvertEventToCytoscape(event types.Event) models.CytoscapeNode {
	return models.CytoscapeNode{
		Data: models.NodeData{
			ID:           event.ResourceID,
			Label:        event.EventName,
			Type:         "falco_event",
			ResourceType: event.ResourceType,
			ResourceName: event.ResourceID,
			Severity:     "info", // Default severity for events
			Metadata: map[string]interface{}{
				"provider":     event.Provider,
				"event_name":   event.EventName,
				"user":         event.UserIdentity.UserName,
				"user_arn":     event.UserIdentity.ARN,
				"changes":      event.Changes,
				"region":       event.Region,
				"project_id":   event.ProjectID,
				"service_name": event.ServiceName,
			},
		},
	}
}

// ConvertUnmanagedToCytoscape converts unmanaged resource to Cytoscape node
func ConvertUnmanagedToCytoscape(unmanaged types.UnmanagedResourceAlert) models.CytoscapeNode {
	return models.CytoscapeNode{
		Data: models.NodeData{
			ID:           unmanaged.ResourceID,
			Label:        unmanaged.ResourceID,
			Type:         "unmanaged",
			ResourceType: unmanaged.ResourceType,
			ResourceName: unmanaged.ResourceID,
			Severity:     unmanaged.Severity,
			Metadata: map[string]interface{}{
				"event_name": unmanaged.EventName,
				"user":       unmanaged.UserIdentity.UserName,
				"user_arn":   unmanaged.UserIdentity.ARN,
				"changes":    unmanaged.Changes,
				"timestamp":  unmanaged.Timestamp,
				"reason":     unmanaged.Reason,
			},
		},
	}
}

// ConvertTerraformResourceToCytoscape converts a Terraform State resource to Cytoscape node
func ConvertTerraformResourceToCytoscape(resource *terraform.Resource, hasDrift bool) models.CytoscapeNode {
	resourceID := extractResourceIDFromAttributes(resource.Attributes)
	resourceName := extractResourceName(resource)

	// Determine node type and severity based on drift status
	nodeType := "terraform_resource"
	severity := "low" // Normal resources are low severity (green)
	if hasDrift {
		nodeType = "terraform_resource_drifted"
		severity = "high" // Drifted resources are high severity (red/orange)
	}

	return models.CytoscapeNode{
		Data: models.NodeData{
			ID:           resourceID,
			Label:        resourceName,
			Type:         nodeType,
			ResourceType: resource.Type,
			ResourceName: resourceName,
			Severity:     severity,
			Metadata: map[string]interface{}{
				"mode":       resource.Mode,
				"provider":   resource.Provider,
				"tf_name":    resource.Name,
				"has_drift":  hasDrift,
				"attributes": resource.Attributes,
			},
		},
	}
}

// extractResourceIDFromAttributes extracts a unique resource ID from Terraform attributes
func extractResourceIDFromAttributes(attributes map[string]interface{}) string {
	// Try to get ID from attributes
	if id, ok := attributes["id"].(string); ok && id != "" {
		return id
	}

	// Fallback to ARN for AWS resources
	if arn, ok := attributes["arn"].(string); ok && arn != "" {
		return arn
	}

	// Fallback to name
	if name, ok := attributes["name"].(string); ok && name != "" {
		return name
	}

	// Fallback to self_link for GCP resources
	if selfLink, ok := attributes["self_link"].(string); ok && selfLink != "" {
		return selfLink
	}

	return ""
}

// extractResourceName extracts a human-readable name from Terraform resource
func extractResourceName(resource *terraform.Resource) string {
	// Try name attribute
	if name, ok := resource.Attributes["name"].(string); ok && name != "" {
		return name
	}

	// Try tags.Name for AWS resources
	if tags, ok := resource.Attributes["tags"].(map[string]interface{}); ok {
		if name, ok := tags["Name"].(string); ok && name != "" {
			return name
		}
	}

	// Fallback to Terraform name
	if resource.Name != "" {
		return resource.Name
	}

	// Fallback to resource type
	return resource.Type
}

// CreateEdge creates a Cytoscape edge
func CreateEdge(source, target, label, edgeType, relationship string) models.CytoscapeEdge {
	return models.CytoscapeEdge{
		Data: models.EdgeData{
			ID:           fmt.Sprintf("%s-%s", source, target),
			Source:       source,
			Target:       target,
			Label:        label,
			Type:         edgeType,
			Relationship: relationship,
		},
	}
}
