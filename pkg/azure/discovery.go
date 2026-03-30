package azure

import (
	"context"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// DiscoveryClient handles Azure resource discovery across a subscription.
type DiscoveryClient struct {
	subscriptionID string
	resourceGroup  string // optional: limit to a single resource group
	regions        []string

	// Azure service clients (interfaces for testability)
	resourceLister ResourceLister
}

// ResourceLister abstracts Azure resource listing for testability.
// In production, this is backed by Azure SDK's armresources client.
type ResourceLister interface {
	// ListResources lists Azure resources, optionally filtered by resource group.
	ListResources(ctx context.Context, subscriptionID string, resourceGroup string) ([]*AzureResource, error)
}

// AzureResource represents a raw Azure resource from the ARM API.
type AzureResource struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`       // e.g., "Microsoft.Compute/virtualMachines"
	Location   string            `json:"location"`
	Tags       map[string]string `json:"tags,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	SKU        *SKU              `json:"sku,omitempty"`
	Kind       string            `json:"kind,omitempty"`
}

// SKU represents the Azure resource SKU.
type SKU struct {
	Name     string `json:"name,omitempty"`
	Tier     string `json:"tier,omitempty"`
	Capacity int64  `json:"capacity,omitempty"`
}

// Type aliases to use shared types from pkg/types
// This allows discovery.go to continue using the old names while using the shared implementations
type DiscoveredResource = types.DiscoveredResource
type DriftResult = types.DriftResult
type TerraformResource = types.TerraformResource
type ResourceDiff = types.ResourceDiff
type FieldDiff = types.FieldDiff

// azureTypeToTerraform maps Azure resource types to Terraform resource types.
var azureTypeToTerraform = map[string]string{
	// Compute
	"microsoft.compute/virtualmachines":         "azurerm_virtual_machine",
	"microsoft.compute/virtualmachinescalesets":  "azurerm_virtual_machine_scale_set",
	"microsoft.compute/disks":                   "azurerm_managed_disk",
	"microsoft.compute/images":                  "azurerm_image",
	"microsoft.compute/snapshots":               "azurerm_snapshot",

	// Networking
	"microsoft.network/networksecuritygroups":    "azurerm_network_security_group",
	"microsoft.network/virtualnetworks":          "azurerm_virtual_network",
	"microsoft.network/publicipaddresses":        "azurerm_public_ip",
	"microsoft.network/networkinterfaces":        "azurerm_network_interface",
	"microsoft.network/loadbalancers":            "azurerm_lb",
	"microsoft.network/routetables":              "azurerm_route_table",
	"microsoft.network/applicationgateways":      "azurerm_application_gateway",
	"microsoft.network/azurefirewalls":           "azurerm_firewall",
	"microsoft.network/firewallpolicies":         "azurerm_firewall_policy",
	"microsoft.network/privateendpoints":         "azurerm_private_endpoint",
	"microsoft.network/frontdoors":               "azurerm_frontdoor",
	"microsoft.network/trafficmanagerprofiles":   "azurerm_traffic_manager_profile",
	"microsoft.network/expressroutecircuits":     "azurerm_express_route_circuit",
	"microsoft.network/vpngateways":              "azurerm_vpn_gateway",
	"microsoft.network/virtualnetworkgateways":   "azurerm_virtual_network_gateway",
	"microsoft.network/localnetworkgateways":     "azurerm_local_network_gateway",
	"microsoft.network/dnszones":                 "azurerm_dns_zone",
	"microsoft.network/privatednszones":          "azurerm_private_dns_zone",
	"microsoft.cdn/profiles":                     "azurerm_cdn_profile",

	// Storage
	"microsoft.storage/storageaccounts":          "azurerm_storage_account",

	// Database
	"microsoft.sql/servers":                      "azurerm_mssql_server",
	"microsoft.sql/servers/databases":            "azurerm_mssql_database",
	"microsoft.dbformysql/servers":               "azurerm_mysql_server",
	"microsoft.dbforpostgresql/servers":          "azurerm_postgresql_server",
	"microsoft.documentdb/databaseaccounts":      "azurerm_cosmosdb_account",

	// Key Vault
	"microsoft.keyvault/vaults":                  "azurerm_key_vault",

	// App Service
	"microsoft.web/sites":                        "azurerm_app_service",
	"microsoft.web/serverfarms":                  "azurerm_app_service_plan",
	"microsoft.web/staticsites":                  "azurerm_static_web_app",

	// Kubernetes
	"microsoft.containerservice/managedclusters": "azurerm_kubernetes_cluster",

	// Containers
	"microsoft.containerregistry/registries":     "azurerm_container_registry",
	"microsoft.containerinstance/containergroups": "azurerm_container_group",

	// Messaging
	"microsoft.servicebus/namespaces":            "azurerm_servicebus_namespace",
	"microsoft.eventgrid/topics":                 "azurerm_eventgrid_topic",
	"microsoft.eventgrid/domains":                "azurerm_eventgrid_domain",
	"microsoft.eventhub/namespaces":              "azurerm_eventhub_namespace",

	// Monitoring
	"microsoft.insights/components":              "azurerm_application_insights",
	"microsoft.operationalinsights/workspaces":   "azurerm_log_analytics_workspace",

	// Cache
	"microsoft.cache/redis":                      "azurerm_redis_cache",

	// Other
	"microsoft.resources/resourcegroups":         "azurerm_resource_group",
	"microsoft.automation/automationaccounts":    "azurerm_automation_account",
	"microsoft.logic/workflows":                  "azurerm_logic_app_workflow",
	"microsoft.datafactory/factories":            "azurerm_data_factory",
	"microsoft.search/searchservices":            "azurerm_search_service",
	"microsoft.batch/batchaccounts":              "azurerm_batch_account",
	"microsoft.synapse/workspaces":               "azurerm_synapse_workspace",
	"microsoft.apimanagement/service":            "azurerm_api_management",
	"microsoft.managedidentity/userassignedidentities": "azurerm_user_assigned_identity",
}

// NewDiscoveryClient creates a new Azure discovery client.
// subscriptionID is the Azure subscription to discover resources in.
// regions specifies which locations to scan (empty = all locations).
func NewDiscoveryClient(subscriptionID string, regions []string, lister ResourceLister) (*DiscoveryClient, error) {
	if subscriptionID == "" {
		return nil, fmt.Errorf("Azure subscription ID is required")
	}

	return &DiscoveryClient{
		subscriptionID: subscriptionID,
		regions:        regions,
		resourceLister: lister,
	}, nil
}

// WithResourceGroup limits discovery to a specific resource group.
func (d *DiscoveryClient) WithResourceGroup(rg string) *DiscoveryClient {
	d.resourceGroup = rg
	return d
}

// DiscoverAll discovers all supported Azure resources in the subscription.
func (d *DiscoveryClient) DiscoverAll(ctx context.Context) ([]*DiscoveredResource, error) {
	log.Infof("Starting Azure resource discovery in subscription %s", d.subscriptionID)

	if d.resourceLister == nil {
		return nil, fmt.Errorf("resource lister is not configured; Azure SDK credentials may be required")
	}

	azureResources, err := d.resourceLister.ListResources(ctx, d.subscriptionID, d.resourceGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to list Azure resources: %w", err)
	}

	var allResources []*DiscoveredResource
	for _, res := range azureResources {
		discovered := d.convertResource(res)
		if discovered == nil {
			continue
		}

		// Apply region filter
		if len(d.regions) > 0 && !containsString(d.regions, discovered.Region) {
			continue
		}

		allResources = append(allResources, discovered)
	}

	log.Infof("Azure discovery completed: %d total resources discovered", len(allResources))
	return allResources, nil
}

// convertResource converts an Azure ARM resource to a DiscoveredResource.
// Returns nil if the resource type is not supported for Terraform mapping.
func (d *DiscoveryClient) convertResource(res *AzureResource) *DiscoveredResource {
	// Normalize the Azure type to lowercase for matching
	azureType := strings.ToLower(res.Type)

	tfType, ok := azureTypeToTerraform[azureType]
	if !ok {
		return nil
	}

	attrs := make(map[string]interface{})
	attrs["name"] = res.Name
	attrs["location"] = res.Location

	// Extract resource group from Azure resource ID
	rg := extractResourceGroupFromID(res.ID)
	if rg != "" {
		attrs["resource_group_name"] = rg
	}

	// Extract SKU information
	if res.SKU != nil {
		if res.SKU.Name != "" {
			attrs["sku_name"] = res.SKU.Name
		}
		if res.SKU.Tier != "" {
			attrs["sku_tier"] = res.SKU.Tier
		}
		if res.SKU.Capacity > 0 {
			attrs["sku_capacity"] = res.SKU.Capacity
		}
	}

	// Extract kind (for App Service: web vs function)
	if res.Kind != "" {
		attrs["kind"] = res.Kind
	}

	// Copy relevant properties
	if res.Properties != nil {
		d.extractProperties(tfType, res.Properties, attrs)
	}

	return &DiscoveredResource{
		ID:         res.ID,
		Type:       tfType,
		Name:       res.Name,
		Region:     strings.ToLower(res.Location),
		Attributes: attrs,
		Tags:       res.Tags,
	}
}

// extractProperties extracts Terraform-relevant properties from Azure resource properties.
func (d *DiscoveryClient) extractProperties(tfType string, properties map[string]interface{}, attrs map[string]interface{}) {
	switch tfType {
	case "azurerm_virtual_machine":
		copyIfExists(properties, attrs, "vmId", "vm_id")
		if hardwareProfile, ok := properties["hardwareProfile"].(map[string]interface{}); ok {
			copyIfExists(hardwareProfile, attrs, "vmSize", "vm_size")
		}
		if osProfile, ok := properties["osProfile"].(map[string]interface{}); ok {
			copyIfExists(osProfile, attrs, "computerName", "computer_name")
			copyIfExists(osProfile, attrs, "adminUsername", "admin_username")
		}
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_virtual_network":
		if addrSpace, ok := properties["addressSpace"].(map[string]interface{}); ok {
			if prefixes, ok := addrSpace["addressPrefixes"].([]interface{}); ok {
				attrs["address_space"] = prefixes
			}
		}
		if dhcpOpts, ok := properties["dhcpOptions"].(map[string]interface{}); ok {
			if dns, ok := dhcpOpts["dnsServers"].([]interface{}); ok {
				attrs["dns_servers"] = dns
			}
		}

	case "azurerm_network_security_group":
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_storage_account":
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")
		if primaryEndpoints, ok := properties["primaryEndpoints"].(map[string]interface{}); ok {
			copyIfExists(primaryEndpoints, attrs, "blob", "primary_blob_endpoint")
		}
		if encryption, ok := properties["encryption"].(map[string]interface{}); ok {
			if services, ok := encryption["services"].(map[string]interface{}); ok {
				if blob, ok := services["blob"].(map[string]interface{}); ok {
					if enabled, ok := blob["enabled"].(bool); ok {
						attrs["blob_encryption_enabled"] = enabled
					}
				}
			}
		}

	case "azurerm_kubernetes_cluster":
		copyIfExists(properties, attrs, "kubernetesVersion", "kubernetes_version")
		copyIfExists(properties, attrs, "dnsPrefix", "dns_prefix")
		copyIfExists(properties, attrs, "fqdn", "fqdn")
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")
		if networkProfile, ok := properties["networkProfile"].(map[string]interface{}); ok {
			copyIfExists(networkProfile, attrs, "networkPlugin", "network_plugin")
			copyIfExists(networkProfile, attrs, "serviceCidr", "service_cidr")
			copyIfExists(networkProfile, attrs, "podCidr", "pod_cidr")
		}

	case "azurerm_mssql_server":
		copyIfExists(properties, attrs, "version", "version")
		copyIfExists(properties, attrs, "administratorLogin", "administrator_login")
		copyIfExists(properties, attrs, "fullyQualifiedDomainName", "fully_qualified_domain_name")

	case "azurerm_mssql_database":
		copyIfExists(properties, attrs, "collation", "collation")
		copyIfExists(properties, attrs, "maxSizeBytes", "max_size_gb")
		copyIfExists(properties, attrs, "status", "status")

	case "azurerm_key_vault":
		if vaultProperties, ok := properties["properties"].(map[string]interface{}); ok {
			copyIfExists(vaultProperties, attrs, "tenantId", "tenant_id")
			if sku, ok := vaultProperties["sku"].(map[string]interface{}); ok {
				copyIfExists(sku, attrs, "name", "sku_name")
			}
			if enabledForDeployment, ok := vaultProperties["enabledForDeployment"].(bool); ok {
				attrs["enabled_for_deployment"] = enabledForDeployment
			}
		}

	case "azurerm_cosmosdb_account":
		copyIfExists(properties, attrs, "databaseAccountOfferType", "offer_type")
		copyIfExists(properties, attrs, "documentEndpoint", "endpoint")
		if consistencyPolicy, ok := properties["consistencyPolicy"].(map[string]interface{}); ok {
			copyIfExists(consistencyPolicy, attrs, "defaultConsistencyLevel", "consistency_level")
		}

	case "azurerm_lb":
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_public_ip":
		copyIfExists(properties, attrs, "publicIPAllocationMethod", "allocation_method")
		copyIfExists(properties, attrs, "ipAddress", "ip_address")
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_app_service", "azurerm_app_service_plan":
		copyIfExists(properties, attrs, "state", "state")
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_redis_cache":
		copyIfExists(properties, attrs, "hostName", "hostname")
		copyIfExists(properties, attrs, "port", "port")
		copyIfExists(properties, attrs, "sslPort", "ssl_port")
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")

	case "azurerm_container_registry":
		copyIfExists(properties, attrs, "loginServer", "login_server")
		copyIfExists(properties, attrs, "adminUserEnabled", "admin_enabled")
		copyIfExists(properties, attrs, "provisioningState", "provisioning_state")
	}
}

// copyIfExists copies a value from src to dst with a new key name.
func copyIfExists(src map[string]interface{}, dst map[string]interface{}, srcKey, dstKey string) {
	if val, ok := src[srcKey]; ok && val != nil {
		dst[dstKey] = val
	}
}

// extractResourceGroupFromID extracts the resource group from an Azure resource ID.
// Example: "/subscriptions/sub-123/resourceGroups/rg-test/providers/..." -> "rg-test"
func extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractSubscriptionFromID extracts the subscription ID from an Azure resource ID.
func extractSubscriptionFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "subscriptions") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// SupportedDiscoveryTypes returns the Terraform resource types that Azure can discover.
func SupportedDiscoveryTypes() []string {
	typeSet := make(map[string]bool)
	for _, tfType := range azureTypeToTerraform {
		typeSet[tfType] = true
	}
	result := make([]string, 0, len(typeSet))
	for rt := range typeSet {
		result = append(result, rt)
	}
	return result
}

// containsString checks if a string slice contains a given string (case-insensitive).
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
