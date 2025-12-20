package graph

import (
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
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
