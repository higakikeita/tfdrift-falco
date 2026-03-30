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
	relevantEvents := map[string]bool{
		// Compute - Virtual Machines
		"Microsoft.Compute/virtualMachines/write":             true,
		"Microsoft.Compute/virtualMachines/delete":            true,
		"Microsoft.Compute/virtualMachines/start/action":      true,
		"Microsoft.Compute/virtualMachines/powerOff/action":   true,
		"Microsoft.Compute/virtualMachines/restart/action":    true,
		"Microsoft.Compute/virtualMachines/deallocate/action": true,

		// Compute - Virtual Machine Scale Sets
		"Microsoft.Compute/virtualMachineScaleSets/write":  true,
		"Microsoft.Compute/virtualMachineScaleSets/delete": true,

		// Network - Security Groups
		"Microsoft.Network/networkSecurityGroups/write":                true,
		"Microsoft.Network/networkSecurityGroups/delete":               true,
		"Microsoft.Network/networkSecurityGroups/securityRules/write":  true,
		"Microsoft.Network/networkSecurityGroups/securityRules/delete": true,

		// Network - Virtual Networks
		"Microsoft.Network/virtualNetworks/write":          true,
		"Microsoft.Network/virtualNetworks/delete":         true,
		"Microsoft.Network/virtualNetworks/subnets/write":  true,
		"Microsoft.Network/virtualNetworks/subnets/delete": true,

		// Network - Load Balancer
		"Microsoft.Network/loadBalancers/write":                      true,
		"Microsoft.Network/loadBalancers/delete":                     true,
		"Microsoft.Network/loadBalancers/backendAddressPools/write":  true,
		"Microsoft.Network/loadBalancers/backendAddressPools/delete": true,

		// Network - Public IP
		"Microsoft.Network/publicIPAddresses/write":  true,
		"Microsoft.Network/publicIPAddresses/delete": true,

		// Network - Network Interfaces
		"Microsoft.Network/networkInterfaces/write":  true,
		"Microsoft.Network/networkInterfaces/delete": true,

		// Network - Route Tables
		"Microsoft.Network/routeTables/write":         true,
		"Microsoft.Network/routeTables/delete":        true,
		"Microsoft.Network/routeTables/routes/write":  true,
		"Microsoft.Network/routeTables/routes/delete": true,

		// Network - CDN
		"Microsoft.Cdn/profiles/write":            true,
		"Microsoft.Cdn/profiles/delete":           true,
		"Microsoft.Cdn/profiles/endpoints/write":  true,
		"Microsoft.Cdn/profiles/endpoints/delete": true,

		// Network - Front Door
		"Microsoft.Network/frontDoors/write":  true,
		"Microsoft.Network/frontDoors/delete": true,

		// Network - Application Gateway
		"Microsoft.Network/applicationGateways/write":  true,
		"Microsoft.Network/applicationGateways/delete": true,

		// Network - Firewall
		"Microsoft.Network/azureFirewalls/write":    true,
		"Microsoft.Network/azureFirewalls/delete":   true,
		"Microsoft.Network/firewallPolicies/write":  true,
		"Microsoft.Network/firewallPolicies/delete": true,

		// Network - Private Endpoints
		"Microsoft.Network/privateEndpoints/write":  true,
		"Microsoft.Network/privateEndpoints/delete": true,

		// Storage - Storage Accounts
		"Microsoft.Storage/storageAccounts/write":  true,
		"Microsoft.Storage/storageAccounts/delete": true,

		// Storage - Blob Services
		"Microsoft.Storage/storageAccounts/blobServices/write":  true,
		"Microsoft.Storage/storageAccounts/blobServices/delete": true,

		// SQL - Servers
		"Microsoft.Sql/servers/write":  true,
		"Microsoft.Sql/servers/delete": true,

		// SQL - Databases
		"Microsoft.Sql/servers/databases/write":  true,
		"Microsoft.Sql/servers/databases/delete": true,

		// Cosmos DB
		"Microsoft.DocumentDB/databaseAccounts/write":  true,
		"Microsoft.DocumentDB/databaseAccounts/delete": true,

		// Key Vault
		"Microsoft.KeyVault/vaults/write":          true,
		"Microsoft.KeyVault/vaults/delete":         true,
		"Microsoft.KeyVault/vaults/secrets/write":  true,
		"Microsoft.KeyVault/vaults/secrets/delete": true,
		"Microsoft.KeyVault/vaults/keys/write":     true,
		"Microsoft.KeyVault/vaults/keys/delete":    true,

		// App Service - Web Apps & Function Apps
		"Microsoft.Web/sites/write":  true,
		"Microsoft.Web/sites/delete": true,

		// App Service - App Service Plans
		"Microsoft.Web/serverfarms/write":  true,
		"Microsoft.Web/serverfarms/delete": true,

		// Kubernetes - AKS
		"Microsoft.ContainerService/managedClusters/write":  true,
		"Microsoft.ContainerService/managedClusters/delete": true,

		// Container Registry
		"Microsoft.ContainerRegistry/registries/write":  true,
		"Microsoft.ContainerRegistry/registries/delete": true,

		// DNS
		"Microsoft.Network/dnszones/write":             true,
		"Microsoft.Network/dnszones/delete":            true,
		"Microsoft.Network/dnszones/recordSets/write":  true,
		"Microsoft.Network/dnszones/recordSets/delete": true,

		// Redis Cache
		"Microsoft.Cache/redis/write":  true,
		"Microsoft.Cache/redis/delete": true,

		// Service Bus
		"Microsoft.ServiceBus/namespaces/write":         true,
		"Microsoft.ServiceBus/namespaces/delete":        true,
		"Microsoft.ServiceBus/namespaces/queues/write":  true,
		"Microsoft.ServiceBus/namespaces/queues/delete": true,
		"Microsoft.ServiceBus/namespaces/topics/write":  true,
		"Microsoft.ServiceBus/namespaces/topics/delete": true,

		// Event Grid
		"Microsoft.EventGrid/topics/write":  true,
		"Microsoft.EventGrid/topics/delete": true,

		// Monitoring - Alert Rules
		"Microsoft.Insights/metricAlerts/write":         true,
		"Microsoft.Insights/metricAlerts/delete":        true,
		"Microsoft.Insights/scheduledQueryRules/write":  true,
		"Microsoft.Insights/scheduledQueryRules/delete": true,

		// Monitoring - Diagnostic Settings
		"Microsoft.Insights/diagnosticSettings/write":  true,
		"Microsoft.Insights/diagnosticSettings/delete": true,

		// Monitoring - Action Groups
		"Microsoft.Insights/actionGroups/write":  true,
		"Microsoft.Insights/actionGroups/delete": true,

		// Log Analytics
		"Microsoft.OperationalInsights/workspaces/write":  true,
		"Microsoft.OperationalInsights/workspaces/delete": true,

		// Identity - Managed Identities
		"Microsoft.ManagedIdentity/userAssignedIdentities/write":  true,
		"Microsoft.ManagedIdentity/userAssignedIdentities/delete": true,

		// Policy - Policy Assignments
		"Microsoft.Authorization/policyAssignments/write":  true,
		"Microsoft.Authorization/policyAssignments/delete": true,

		// Authorization - Role Assignments
		"Microsoft.Authorization/roleAssignments/write":  true,
		"Microsoft.Authorization/roleAssignments/delete": true,

		// Authorization - Resource Locks
		"Microsoft.Authorization/locks/write":  true,
		"Microsoft.Authorization/locks/delete": true,

		// API Management
		"Microsoft.ApiManagement/service/write":       true,
		"Microsoft.ApiManagement/service/delete":      true,
		"Microsoft.ApiManagement/service/apis/write":  true,
		"Microsoft.ApiManagement/service/apis/delete": true,
	}

	return relevantEvents[operationName]
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
