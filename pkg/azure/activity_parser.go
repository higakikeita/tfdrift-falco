// Package azure provides Azure Activity Log parsing and resource mapping
// for TFDrift-Falco. It processes Azure Activity Logs received via Falco's
// azure_activity plugin and maps them to Terraform resource types for drift detection.
//
// The package supports 100+ Azure operation types across 20+ services including:
//   - Compute (Virtual Machines, Virtual Machine Scale Sets)
//   - Networking (NSGs, Virtual Networks, Load Balancers, CDN, Front Door, Application Gateway, Firewall)
//   - Storage (Storage Accounts, Storage Account Failover)
//   - Database (SQL Servers, SQL Databases, Azure Cosmos DB)
//   - Keyvault & Security (Key Vault, Managed Identities)
//   - App Services (Web Apps, Function Apps, App Service Plans)
//   - Kubernetes (AKS Clusters, AKS Node Pools)
//   - Containers (Container Registry)
//   - Messaging (Service Bus, Event Grid)
//   - Monitoring (Monitor, Log Analytics, Alert Rules, Diagnostics)
//   - Cache (Redis Cache)
//   - Other services (DNS, Private Endpoints, Policy, Role Assignments, Resource Locks)
//
// Example usage:
//
//	parser := azure.NewActivityParser()
//	event := parser.Parse(falcoResponse)
//	if event != nil {
//	    // Process drift detection event
//	}
package azure

import (
	"fmt"
	"strings"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/parser"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// ActivityParser parses Azure Activity Log events from Falco into TFDrift events.
//
// The parser extracts relevant information from Falco's azure_activity plugin output,
// including resource identifiers, user identity, subscription/resource group information,
// and maps Azure operation names to corresponding Terraform resource types.
//
// Thread-safe: Multiple goroutines can safely call Parse concurrently.
type ActivityParser struct {
	mapper *ResourceMapper
	base   *parser.BaseEventParser
}

// NewActivityParser creates a new Azure Activity Log parser with pre-initialized resource mappings.
//
// The parser is configured with 200+ event-to-resource mappings covering major Azure services.
// Returns a ready-to-use parser instance that can process Falco azure_activity events.
func NewActivityParser() *ActivityParser {
	ap := &ActivityParser{
		mapper: NewResourceMapper(),
	}
	ap.base = parser.NewBaseEventParser(ap.createConfig())
	return ap
}

// Parse converts a Falco output response into a TFDrift event for drift detection.
//
// The method performs the following operations:
//   - Validates the response is from the azure_activity source
//   - Extracts Azure operation name (e.g., "Microsoft.Compute/virtualMachines/write")
//   - Filters irrelevant events (read-only operations, non-infrastructure changes)
//   - Maps the operation to a Terraform resource type (e.g., "azurerm_virtual_machine")
//   - Extracts resource identifiers, subscription, resource group, and user identity
//   - Parses event-specific changes from the activity log
//
// Parameters:
//   - res: Falco output response containing Azure activity log data
//
// Returns:
//   - *types.Event: Parsed drift detection event, or nil if:
//   - Response is nil
//   - Source is not "azure_activity"
//   - Event is not relevant for drift detection
//   - Required fields are missing
//   - No resource mapping exists for the event type
//
// Example:
//
//	parser := NewActivityParser()
//	event := parser.Parse(falcoResponse)
//	if event != nil {
//	    fmt.Printf("Drift detected: %s on %s.%s\n",
//	        event.EventName, event.ResourceType, event.ResourceID)
//	}
func (p *ActivityParser) Parse(res *outputs.Response) *types.Event {
	event := p.base.Parse(res)
	if event == nil {
		return nil
	}

	// Set deprecated field for backward compatibility
	if event.Metadata != nil {
		if region, ok := event.Metadata["region"]; ok {
			event.Region = region
		}
	}

	return event
}

// createConfig creates the parser configuration for the BaseEventParser
func (p *ActivityParser) createConfig() parser.EventParserConfig {
	return parser.EventParserConfig{
		Provider:       "azure",
		ExpectedSource: "azure_activity",
		ExtractEventName: func(fields map[string]string) string {
			return parser.GetStringField(fields, "azure.operationName")
		},
		IsRelevantEvent: p.isRelevantEvent,
		ExtractResourceID: func(eventName string, fields map[string]string) string {
			resourceID := parser.GetStringField(fields, "azure.resourceId")
			if resourceID == "" {
				return ""
			}
			return p.extractResourceNameFromID(resourceID)
		},
		MapResourceType: func(eventName string, fields map[string]string) string {
			return p.mapper.MapEventToResource(eventName)
		},
		ExtractUserIdentity: func(fields map[string]string) types.UserIdentity {
			resourceID := parser.GetStringField(fields, "azure.resourceId")
			subscriptionID := p.extractSubscriptionIDFromID(resourceID)
			return types.UserIdentity{
				Type:      "AzureAD",
				UserName:  parser.GetStringField(fields, "azure.caller"),
				AccountID: subscriptionID,
			}
		},
		ExtractChanges: p.extractChanges,
		ExtractMetadata: func(eventName string, fields map[string]string) map[string]string {
			resourceID := parser.GetStringField(fields, "azure.resourceId")
			metadata := make(map[string]string)
			if region := parser.GetStringField(fields, "azure.resourceLocation"); region != "" {
				metadata["region"] = region
			}
			if subscriptionID := p.extractSubscriptionIDFromID(resourceID); subscriptionID != "" {
				metadata["subscription_id"] = subscriptionID
			}
			if resourceGroup := p.extractResourceGroupFromID(resourceID); resourceGroup != "" {
				metadata["resource_group"] = resourceGroup
			}
			return metadata
		},
	}
}

// isRelevantEvent checks if an Azure operation is relevant for drift detection
func (p *ActivityParser) isRelevantEvent(operationName string) bool {
	cfg, err := LoadResourceConfig()
	if err != nil {
		// Fallback to empty map if config loading fails
		return false
	}
	return cfg.IsRelevantEvent(operationName)
}

// extractResourceNameFromID extracts the resource name from an Azure resource ID
// Example: "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm" -> "my-vm"
func (p *ActivityParser) extractResourceNameFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// extractSubscriptionIDFromID extracts the subscription ID from an Azure resource ID
// Example: "/subscriptions/sub-123/resourceGroups/..." -> "sub-123"
func (p *ActivityParser) extractSubscriptionIDFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if part == "subscriptions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractResourceGroupFromID extracts the resource group name from an Azure resource ID
// Example: "/subscriptions/sub-123/resourceGroups/rg-test/..." -> "rg-test"
func (p *ActivityParser) extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractChanges extracts attribute changes from Azure Activity Log
func (p *ActivityParser) extractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	// Extract request properties (what was requested/changed)
	requestProperties := parser.GetStringField(fields, "azure.requestProperties")

	// Extract response properties (what resulted)
	responseProperties := parser.GetStringField(fields, "azure.responseProperties")

	// Event-specific extraction
	switch {
	case strings.Contains(eventName, "/write"):
		// Resource modification/creation
		changes["_action"] = "write"
		if requestProperties != "" {
			changes["_properties"] = requestProperties
		}
		if responseProperties != "" {
			changes["_response"] = responseProperties
		}

	case strings.Contains(eventName, "/delete"):
		// Resource deletion
		changes["_action"] = "delete"

	case strings.HasSuffix(eventName, "/action"):
		// Action operation (start, stop, restart, deallocate, etc.)
		actionType := eventName[strings.LastIndex(eventName, "/")+1:]
		actionType = strings.TrimSuffix(actionType, "/action")
		changes["_action"] = actionType
	}

	// Include correlation ID for tracking
	correlationID := parser.GetStringField(fields, "azure.correlationId")
	if correlationID != "" {
		changes["_correlation_id"] = correlationID
	}

	// Include operation status
	status := parser.GetStringField(fields, "azure.status")
	if status != "" {
		changes["_status"] = status
	}

	return changes
}

// ValidateEvent performs validation on parsed event
func (p *ActivityParser) ValidateEvent(event *types.Event) error {
	if event == nil {
		return fmt.Errorf("event is nil")
	}

	if event.Provider != "azure" {
		return fmt.Errorf("invalid provider: %s (expected: azure)", event.Provider)
	}

	if event.EventName == "" {
		return fmt.Errorf("event name is empty")
	}

	if event.ResourceType == "" {
		return fmt.Errorf("resource type is empty")
	}

	if event.ResourceID == "" {
		return fmt.Errorf("resource ID is empty")
	}

	return nil
}
