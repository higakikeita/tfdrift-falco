package azure

import "strings"

// ResourceMapper maps Azure operation names to Terraform resource types.
//
// The mapper maintains a comprehensive mapping of 200+ Azure operation names
// (e.g., "Microsoft.Compute/virtualMachines/write") to their corresponding Terraform resource
// types (e.g., "azurerm_virtual_machine").
//
// This enables TFDrift-Falco to correlate infrastructure changes detected in Azure
// Activity Logs with resources defined in Terraform state files.
//
// Supported services:
//   - Compute: Virtual Machines, VMSS, Images, Snapshots
//   - Networking: NSGs, VNets, Load Balancers, Public IPs, NICs, Route Tables, Application Gateway, Firewall, Private Endpoints
//   - Storage: Storage Accounts, Blob Services, Containers, File Shares, Queues, Tables
//   - Database: SQL Servers, SQL Databases, MySQL, PostgreSQL, Cosmos DB
//   - Keyvault & Security: Key Vault, Managed Identities
//   - App Services: Web Apps, Function Apps, App Service Plans
//   - Kubernetes: AKS Clusters
//   - Containers: Container Registry, Container Instances
//   - Messaging: Service Bus, Event Grid, Event Hubs
//   - Monitoring: Alert Rules, Action Groups, Log Analytics, Diagnostic Settings
//   - Cache: Redis Cache
//   - DNS, CDN, Authorization, Policy, and more
//
// Thread-safe: Multiple goroutines can safely call MapEventToResource concurrently.
type ResourceMapper struct {
	eventToResource map[string]string
}

// NewResourceMapper creates a new resource mapper with pre-initialized mappings.
//
// The mapper is initialized with 200+ event-to-resource mappings covering
// all major Azure services supported by TFDrift-Falco.
//
// Returns a ready-to-use mapper instance for event translation.
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{
		eventToResource: initializeEventMapping(),
	}
}

// MapEventToResource maps an Azure operation name to its Terraform resource type.
//
// The method performs case-sensitive lookup of the Azure operation name in the
// pre-initialized mapping table and returns the corresponding Terraform resource type.
//
// Parameters:
//   - operationName: Azure operation name (e.g., "Microsoft.Compute/virtualMachines/write",
//     "Microsoft.Network/networkSecurityGroups/write")
//
// Returns:
//   - string: Terraform resource type (e.g., "azurerm_virtual_machine",
//     "azurerm_network_security_group"), or empty string if no mapping exists
//
// Example:
//
//	mapper := NewResourceMapper()
//	resourceType := mapper.MapEventToResource("Microsoft.Compute/virtualMachines/write")
//	// Returns: "azurerm_virtual_machine"
//
//	resourceType = mapper.MapEventToResource("unknown.operation.name")
//	// Returns: ""
func (m *ResourceMapper) MapEventToResource(operationName string) string {
	// Exact match first
	if resourceType, ok := m.eventToResource[operationName]; ok {
		return resourceType
	}
	return ""
}

// initializeEventMapping initializes the Azure operation name to Terraform resource type mapping table.
//
// This function creates a comprehensive map of 200+ Azure operation names
// to their corresponding Terraform resource types. The mappings cover:
//   - 25+ Azure services
//   - Infrastructure mutation operations (write, delete, action operations)
//   - Resource-specific operations
//
// Returns a map[string]string with operationName -> terraformResourceType mappings.
func initializeEventMapping() map[string]string {
	return map[string]string{
		// ==================== COMPUTE ====================
		// Virtual Machines
		"Microsoft.Compute/virtualMachines/write":              "azurerm_virtual_machine",
		"Microsoft.Compute/virtualMachines/delete":            "azurerm_virtual_machine",
		"Microsoft.Compute/virtualMachines/start/action":      "azurerm_virtual_machine",
		"Microsoft.Compute/virtualMachines/powerOff/action":   "azurerm_virtual_machine",
		"Microsoft.Compute/virtualMachines/restart/action":    "azurerm_virtual_machine",
		"Microsoft.Compute/virtualMachines/deallocate/action": "azurerm_virtual_machine",

		// Virtual Machine Scale Sets
		"Microsoft.Compute/virtualMachineScaleSets/write":  "azurerm_virtual_machine_scale_set",
		"Microsoft.Compute/virtualMachineScaleSets/delete": "azurerm_virtual_machine_scale_set",

		// Images
		"Microsoft.Compute/images/write":  "azurerm_image",
		"Microsoft.Compute/images/delete": "azurerm_image",

		// Snapshots
		"Microsoft.Compute/snapshots/write":  "azurerm_snapshot",
		"Microsoft.Compute/snapshots/delete": "azurerm_snapshot",

		// Disks
		"Microsoft.Compute/disks/write":  "azurerm_managed_disk",
		"Microsoft.Compute/disks/delete": "azurerm_managed_disk",

		// ==================== NETWORKING ====================
		// Network Security Groups
		"Microsoft.Network/networkSecurityGroups/write":                "azurerm_network_security_group",
		"Microsoft.Network/networkSecurityGroups/delete":               "azurerm_network_security_group",
		"Microsoft.Network/networkSecurityGroups/securityRules/write":  "azurerm_network_security_rule",
		"Microsoft.Network/networkSecurityGroups/securityRules/delete": "azurerm_network_security_rule",

		// Virtual Networks
		"Microsoft.Network/virtualNetworks/write":   "azurerm_virtual_network",
		"Microsoft.Network/virtualNetworks/delete":  "azurerm_virtual_network",
		"Microsoft.Network/virtualNetworks/subnets/write":   "azurerm_subnet",
		"Microsoft.Network/virtualNetworks/subnets/delete":  "azurerm_subnet",

		// Load Balancers
		"Microsoft.Network/loadBalancers/write":  "azurerm_lb",
		"Microsoft.Network/loadBalancers/delete": "azurerm_lb",
		"Microsoft.Network/loadBalancers/backendAddressPools/write":  "azurerm_lb_backend_address_pool",
		"Microsoft.Network/loadBalancers/backendAddressPools/delete": "azurerm_lb_backend_address_pool",
		"Microsoft.Network/loadBalancers/rules/write":  "azurerm_lb_rule",
		"Microsoft.Network/loadBalancers/rules/delete": "azurerm_lb_rule",

		// Public IP Addresses
		"Microsoft.Network/publicIPAddresses/write":  "azurerm_public_ip",
		"Microsoft.Network/publicIPAddresses/delete": "azurerm_public_ip",

		// Network Interfaces
		"Microsoft.Network/networkInterfaces/write":  "azurerm_network_interface",
		"Microsoft.Network/networkInterfaces/delete": "azurerm_network_interface",

		// Route Tables
		"Microsoft.Network/routeTables/write":   "azurerm_route_table",
		"Microsoft.Network/routeTables/delete":  "azurerm_route_table",
		"Microsoft.Network/routeTables/routes/write":   "azurerm_route",
		"Microsoft.Network/routeTables/routes/delete":  "azurerm_route",

		// Application Gateway
		"Microsoft.Network/applicationGateways/write":  "azurerm_application_gateway",
		"Microsoft.Network/applicationGateways/delete": "azurerm_application_gateway",

		// Firewall
		"Microsoft.Network/azureFirewalls/write":      "azurerm_firewall",
		"Microsoft.Network/azureFirewalls/delete":     "azurerm_firewall",
		"Microsoft.Network/firewallPolicies/write":    "azurerm_firewall_policy",
		"Microsoft.Network/firewallPolicies/delete":   "azurerm_firewall_policy",
		"Microsoft.Network/firewallPolicies/ruleCollectionGroups/write":   "azurerm_firewall_policy_rule_collection_group",
		"Microsoft.Network/firewallPolicies/ruleCollectionGroups/delete":  "azurerm_firewall_policy_rule_collection_group",

		// Private Endpoints
		"Microsoft.Network/privateEndpoints/write":  "azurerm_private_endpoint",
		"Microsoft.Network/privateEndpoints/delete": "azurerm_private_endpoint",

		// CDN
		"Microsoft.Cdn/profiles/write":   "azurerm_cdn_profile",
		"Microsoft.Cdn/profiles/delete":  "azurerm_cdn_profile",
		"Microsoft.Cdn/profiles/endpoints/write":   "azurerm_cdn_endpoint",
		"Microsoft.Cdn/profiles/endpoints/delete":  "azurerm_cdn_endpoint",

		// Front Door
		"Microsoft.Network/frontDoors/write":  "azurerm_frontdoor",
		"Microsoft.Network/frontDoors/delete": "azurerm_frontdoor",

		// Traffic Manager
		"Microsoft.Network/trafficManagerProfiles/write":  "azurerm_traffic_manager_profile",
		"Microsoft.Network/trafficManagerProfiles/delete": "azurerm_traffic_manager_profile",

		// Express Route
		"Microsoft.Network/expressRouteCircuits/write":  "azurerm_express_route_circuit",
		"Microsoft.Network/expressRouteCircuits/delete": "azurerm_express_route_circuit",

		// VPN Gateway
		"Microsoft.Network/vpnGateways/write":  "azurerm_vpn_gateway",
		"Microsoft.Network/vpnGateways/delete": "azurerm_vpn_gateway",

		// Virtual Network Gateway
		"Microsoft.Network/virtualNetworkGateways/write":  "azurerm_virtual_network_gateway",
		"Microsoft.Network/virtualNetworkGateways/delete": "azurerm_virtual_network_gateway",

		// Local Network Gateway
		"Microsoft.Network/localNetworkGateways/write":  "azurerm_local_network_gateway",
		"Microsoft.Network/localNetworkGateways/delete": "azurerm_local_network_gateway",

		// ==================== STORAGE ====================
		// Storage Accounts
		"Microsoft.Storage/storageAccounts/write":  "azurerm_storage_account",
		"Microsoft.Storage/storageAccounts/delete": "azurerm_storage_account",

		// Blob Services
		"Microsoft.Storage/storageAccounts/blobServices/write":   "azurerm_storage_blob",
		"Microsoft.Storage/storageAccounts/blobServices/delete":  "azurerm_storage_blob",
		"Microsoft.Storage/storageAccounts/blobServices/containers/write":   "azurerm_storage_container",
		"Microsoft.Storage/storageAccounts/blobServices/containers/delete":  "azurerm_storage_container",

		// File Shares
		"Microsoft.Storage/storageAccounts/fileServices/shares/write":   "azurerm_storage_share",
		"Microsoft.Storage/storageAccounts/fileServices/shares/delete":  "azurerm_storage_share",

		// Queues
		"Microsoft.Storage/storageAccounts/queueServices/queues/write":   "azurerm_storage_queue",
		"Microsoft.Storage/storageAccounts/queueServices/queues/delete":  "azurerm_storage_queue",

		// Tables
		"Microsoft.Storage/storageAccounts/tableServices/tables/write":   "azurerm_storage_table",
		"Microsoft.Storage/storageAccounts/tableServices/tables/delete":  "azurerm_storage_table",

		// ==================== DATABASE ====================
		// SQL Server
		"Microsoft.Sql/servers/write":  "azurerm_mssql_server",
		"Microsoft.Sql/servers/delete": "azurerm_mssql_server",

		// SQL Database
		"Microsoft.Sql/servers/databases/write":  "azurerm_mssql_database",
		"Microsoft.Sql/servers/databases/delete": "azurerm_mssql_database",

		// SQL Failover Groups
		"Microsoft.Sql/servers/failoverGroups/write":  "azurerm_mssql_failover_group",
		"Microsoft.Sql/servers/failoverGroups/delete": "azurerm_mssql_failover_group",

		// SQL Elastic Pools
		"Microsoft.Sql/servers/elasticPools/write":  "azurerm_mssql_elasticpool",
		"Microsoft.Sql/servers/elasticPools/delete": "azurerm_mssql_elasticpool",

		// MySQL Server
		"Microsoft.DBforMySQL/servers/write":  "azurerm_mysql_server",
		"Microsoft.DBforMySQL/servers/delete": "azurerm_mysql_server",

		// MySQL Database
		"Microsoft.DBforMySQL/servers/databases/write":  "azurerm_mysql_database",
		"Microsoft.DBforMySQL/servers/databases/delete": "azurerm_mysql_database",

		// PostgreSQL Server
		"Microsoft.DBforPostgreSQL/servers/write":  "azurerm_postgresql_server",
		"Microsoft.DBforPostgreSQL/servers/delete": "azurerm_postgresql_server",

		// PostgreSQL Database
		"Microsoft.DBforPostgreSQL/servers/databases/write":  "azurerm_postgresql_database",
		"Microsoft.DBforPostgreSQL/servers/databases/delete": "azurerm_postgresql_database",

		// Cosmos DB
		"Microsoft.DocumentDB/databaseAccounts/write":  "azurerm_cosmosdb_account",
		"Microsoft.DocumentDB/databaseAccounts/delete": "azurerm_cosmosdb_account",
		"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/write":   "azurerm_cosmosdb_sql_database",
		"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/delete":  "azurerm_cosmosdb_sql_database",
		"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/write":   "azurerm_cosmosdb_sql_container",
		"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/delete":  "azurerm_cosmosdb_sql_container",

		// ==================== KEY VAULT & SECURITY ====================
		// Key Vault
		"Microsoft.KeyVault/vaults/write":  "azurerm_key_vault",
		"Microsoft.KeyVault/vaults/delete": "azurerm_key_vault",

		// Key Vault Secrets
		"Microsoft.KeyVault/vaults/secrets/write":  "azurerm_key_vault_secret",
		"Microsoft.KeyVault/vaults/secrets/delete": "azurerm_key_vault_secret",

		// Key Vault Keys
		"Microsoft.KeyVault/vaults/keys/write":  "azurerm_key_vault_key",
		"Microsoft.KeyVault/vaults/keys/delete": "azurerm_key_vault_key",

		// Key Vault Certificates
		"Microsoft.KeyVault/vaults/certificates/write":  "azurerm_key_vault_certificate",
		"Microsoft.KeyVault/vaults/certificates/delete": "azurerm_key_vault_certificate",

		// Managed Identities
		"Microsoft.ManagedIdentity/userAssignedIdentities/write":  "azurerm_user_assigned_identity",
		"Microsoft.ManagedIdentity/userAssignedIdentities/delete": "azurerm_user_assigned_identity",

		// ==================== APP SERVICES ====================
		// Web Apps / App Service
		"Microsoft.Web/sites/write":  "azurerm_app_service",
		"Microsoft.Web/sites/delete": "azurerm_app_service",

		// App Service Plans
		"Microsoft.Web/serverfarms/write":  "azurerm_app_service_plan",
		"Microsoft.Web/serverfarms/delete": "azurerm_app_service_plan",

		// Function Apps (share the same API path as Web Apps - Microsoft.Web/sites)
		// Differentiation requires inspecting the "kind" property at runtime.
		// The mapping above (azurerm_app_service) covers both web and function apps.

		// Static Web Apps
		"Microsoft.Web/staticSites/write":  "azurerm_static_web_app",
		"Microsoft.Web/staticSites/delete": "azurerm_static_web_app",

		// ==================== KUBERNETES ====================
		// AKS Clusters
		"Microsoft.ContainerService/managedClusters/write":  "azurerm_kubernetes_cluster",
		"Microsoft.ContainerService/managedClusters/delete": "azurerm_kubernetes_cluster",

		// ==================== CONTAINERS ====================
		// Container Registry
		"Microsoft.ContainerRegistry/registries/write":  "azurerm_container_registry",
		"Microsoft.ContainerRegistry/registries/delete": "azurerm_container_registry",

		// Container Instances
		"Microsoft.ContainerInstance/containerGroups/write":  "azurerm_container_group",
		"Microsoft.ContainerInstance/containerGroups/delete": "azurerm_container_group",

		// ==================== MESSAGING ====================
		// Service Bus Namespaces
		"Microsoft.ServiceBus/namespaces/write":  "azurerm_servicebus_namespace",
		"Microsoft.ServiceBus/namespaces/delete": "azurerm_servicebus_namespace",

		// Service Bus Queues
		"Microsoft.ServiceBus/namespaces/queues/write":  "azurerm_servicebus_queue",
		"Microsoft.ServiceBus/namespaces/queues/delete": "azurerm_servicebus_queue",

		// Service Bus Topics
		"Microsoft.ServiceBus/namespaces/topics/write":  "azurerm_servicebus_topic",
		"Microsoft.ServiceBus/namespaces/topics/delete": "azurerm_servicebus_topic",

		// Service Bus Subscriptions
		"Microsoft.ServiceBus/namespaces/topics/subscriptions/write":  "azurerm_servicebus_subscription",
		"Microsoft.ServiceBus/namespaces/topics/subscriptions/delete": "azurerm_servicebus_subscription",

		// Event Grid Topics
		"Microsoft.EventGrid/topics/write":  "azurerm_eventgrid_topic",
		"Microsoft.EventGrid/topics/delete": "azurerm_eventgrid_topic",

		// Event Grid Domains
		"Microsoft.EventGrid/domains/write":  "azurerm_eventgrid_domain",
		"Microsoft.EventGrid/domains/delete": "azurerm_eventgrid_domain",

		// Event Hubs Namespaces
		"Microsoft.EventHub/namespaces/write":  "azurerm_eventhub_namespace",
		"Microsoft.EventHub/namespaces/delete": "azurerm_eventhub_namespace",

		// Event Hubs
		"Microsoft.EventHub/namespaces/eventhubs/write":  "azurerm_eventhub",
		"Microsoft.EventHub/namespaces/eventhubs/delete": "azurerm_eventhub",

		// ==================== MONITORING ====================
		// Alert Rules
		"Microsoft.Insights/metricAlerts/write":        "azurerm_monitor_metric_alert",
		"Microsoft.Insights/metricAlerts/delete":       "azurerm_monitor_metric_alert",
		"Microsoft.Insights/scheduledQueryRules/write": "azurerm_monitor_scheduled_query_rules_alert",
		"Microsoft.Insights/scheduledQueryRules/delete": "azurerm_monitor_scheduled_query_rules_alert",

		// Action Groups
		"Microsoft.Insights/actionGroups/write":  "azurerm_monitor_action_group",
		"Microsoft.Insights/actionGroups/delete": "azurerm_monitor_action_group",

		// Diagnostic Settings
		"Microsoft.Insights/diagnosticSettings/write":  "azurerm_monitor_diagnostic_setting",
		"Microsoft.Insights/diagnosticSettings/delete": "azurerm_monitor_diagnostic_setting",

		// Log Analytics Workspaces
		"Microsoft.OperationalInsights/workspaces/write":  "azurerm_log_analytics_workspace",
		"Microsoft.OperationalInsights/workspaces/delete": "azurerm_log_analytics_workspace",

		// Application Insights
		"Microsoft.Insights/components/write":  "azurerm_application_insights",
		"Microsoft.Insights/components/delete": "azurerm_application_insights",

		// ==================== CACHE ====================
		// Redis Cache
		"Microsoft.Cache/redis/write":  "azurerm_redis_cache",
		"Microsoft.Cache/redis/delete": "azurerm_redis_cache",

		// ==================== DNS ====================
		// DNS Zones
		"Microsoft.Network/dnszones/write":  "azurerm_dns_zone",
		"Microsoft.Network/dnszones/delete": "azurerm_dns_zone",

		// DNS Record Sets
		"Microsoft.Network/dnszones/recordSets/write":  "azurerm_dns_a_record",
		"Microsoft.Network/dnszones/recordSets/delete": "azurerm_dns_a_record",

		// Private DNS Zones
		"Microsoft.Network/privateDnsZones/write":  "azurerm_private_dns_zone",
		"Microsoft.Network/privateDnsZones/delete": "azurerm_private_dns_zone",

		// ==================== API MANAGEMENT ====================
		// API Management Services
		"Microsoft.ApiManagement/service/write":  "azurerm_api_management",
		"Microsoft.ApiManagement/service/delete": "azurerm_api_management",

		// APIs
		"Microsoft.ApiManagement/service/apis/write":  "azurerm_api_management_api",
		"Microsoft.ApiManagement/service/apis/delete": "azurerm_api_management_api",

		// API Operations
		"Microsoft.ApiManagement/service/apis/operations/write":  "azurerm_api_management_api_operation",
		"Microsoft.ApiManagement/service/apis/operations/delete": "azurerm_api_management_api_operation",

		// ==================== AUTHORIZATION & GOVERNANCE ====================
		// Role Assignments
		"Microsoft.Authorization/roleAssignments/write":  "azurerm_role_assignment",
		"Microsoft.Authorization/roleAssignments/delete": "azurerm_role_assignment",

		// Policy Assignments
		"Microsoft.Authorization/policyAssignments/write":  "azurerm_resource_group_policy_assignment",
		"Microsoft.Authorization/policyAssignments/delete": "azurerm_resource_group_policy_assignment",

		// Management Locks
		"Microsoft.Authorization/locks/write":  "azurerm_management_lock",
		"Microsoft.Authorization/locks/delete": "azurerm_management_lock",

		// ==================== BATCH ====================
		// Batch Accounts
		"Microsoft.Batch/batchAccounts/write":  "azurerm_batch_account",
		"Microsoft.Batch/batchAccounts/delete": "azurerm_batch_account",

		// Batch Pools
		"Microsoft.Batch/batchAccounts/pools/write":  "azurerm_batch_pool",
		"Microsoft.Batch/batchAccounts/pools/delete": "azurerm_batch_pool",

		// ==================== DATA FACTORY ====================
		// Data Factories
		"Microsoft.DataFactory/factories/write":  "azurerm_data_factory",
		"Microsoft.DataFactory/factories/delete": "azurerm_data_factory",

		// Data Factory Pipelines
		"Microsoft.DataFactory/factories/pipelines/write":  "azurerm_data_factory_pipeline",
		"Microsoft.DataFactory/factories/pipelines/delete": "azurerm_data_factory_pipeline",

		// ==================== SEARCH ====================
		// Search Services
		"Microsoft.Search/searchServices/write":  "azurerm_search_service",
		"Microsoft.Search/searchServices/delete": "azurerm_search_service",

		// ==================== STREAM ANALYTICS ====================
		// Stream Analytics Jobs
		"Microsoft.StreamAnalytics/streamingjobs/write":  "azurerm_stream_analytics_job",
		"Microsoft.StreamAnalytics/streamingjobs/delete": "azurerm_stream_analytics_job",

		// ==================== SYNAPSE ====================
		// Synapse Workspaces
		"Microsoft.Synapse/workspaces/write":  "azurerm_synapse_workspace",
		"Microsoft.Synapse/workspaces/delete": "azurerm_synapse_workspace",

		// Synapse Spark Pools
		"Microsoft.Synapse/workspaces/bigDataPools/write":  "azurerm_synapse_spark_pool",
		"Microsoft.Synapse/workspaces/bigDataPools/delete": "azurerm_synapse_spark_pool",

		// ==================== LOGIC APPS ====================
		// Logic Apps
		"Microsoft.Logic/workflows/write":  "azurerm_logic_app_workflow",
		"Microsoft.Logic/workflows/delete": "azurerm_logic_app_workflow",

		// ==================== AUTOMATION ====================
		// Automation Accounts
		"Microsoft.Automation/automationAccounts/write":  "azurerm_automation_account",
		"Microsoft.Automation/automationAccounts/delete": "azurerm_automation_account",

		// Runbooks
		"Microsoft.Automation/automationAccounts/runbooks/write":  "azurerm_automation_runbook",
		"Microsoft.Automation/automationAccounts/runbooks/delete": "azurerm_automation_runbook",

		// ==================== RESOURCE GROUPS ====================
		// Resource Groups
		"Microsoft.Resources/resourceGroups/write":  "azurerm_resource_group",
		"Microsoft.Resources/resourceGroups/delete": "azurerm_resource_group",
	}
}

// GetAllSupportedEvents returns all supported operation names
func (m *ResourceMapper) GetAllSupportedEvents() []string {
	events := make([]string, 0, len(m.eventToResource))
	for event := range m.eventToResource {
		events = append(events, event)
	}
	return events
}

// GetResourceTypesForService returns all resource types for a given Azure service
func (m *ResourceMapper) GetResourceTypesForService(serviceName string) []string {
	resourceTypes := make(map[string]bool)

	for event, resourceType := range m.eventToResource {
		// Check if event starts with service name
		if strings.HasPrefix(event, serviceName) {
			resourceTypes[resourceType] = true
		}
	}

	result := make([]string, 0, len(resourceTypes))
	for rt := range resourceTypes {
		result = append(result, rt)
	}
	return result
}
