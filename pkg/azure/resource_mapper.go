// Package azure provides Azure Activity Log event parsing and Terraform resource mapping.
//
// This package enables TFDrift-Falco to detect infrastructure drift in Azure
// by correlating Azure Activity Log operations with Terraform-managed resources.
//
// Supported services (v0.7.0):
//   - Compute: VMs, VM Scale Sets, AKS, Container Instances, App Service, Functions
//   - Networking: VNets, Subnets, NSGs, Load Balancers, Application Gateways, DNS
//   - Storage & Database: Storage Accounts, Azure SQL, CosmosDB, Redis Cache, PostgreSQL
//   - Security & Identity: Key Vault, Managed Identity, Role Assignments
package azure

// ResourceMapper maps Azure Activity Log operation names to Terraform resource types.
type ResourceMapper struct {
	operationToResource map[string]string
}

// NewResourceMapper creates a new resource mapper with pre-initialized mappings.
func NewResourceMapper() *ResourceMapper {
	return &ResourceMapper{
		operationToResource: initializeOperationMapping(),
	}
}

// MapOperationToResource maps an Azure Activity Log operation name to its Terraform resource type.
//
// Parameters:
//   - operationName: Azure Activity Log operation (e.g., "Microsoft.Compute/virtualMachines/write")
//
// Returns:
//   - Terraform resource type (e.g., "azurerm_virtual_machine"), or empty string if no mapping exists
func (m *ResourceMapper) MapOperationToResource(operationName string) string {
	if resourceType, ok := m.operationToResource[operationName]; ok {
		return resourceType
	}
	return ""
}

// GetAllSupportedOperations returns all supported Azure operation names
func (m *ResourceMapper) GetAllSupportedOperations() []string {
	ops := make([]string, 0, len(m.operationToResource))
	for op := range m.operationToResource {
		ops = append(ops, op)
	}
	return ops
}

// GetSupportedServiceCount returns the number of unique Azure services supported
func (m *ResourceMapper) GetSupportedServiceCount() int {
	services := make(map[string]bool)
	for op := range m.operationToResource {
		// Extract service from operation: "Microsoft.Compute/virtualMachines/write" -> "Microsoft.Compute"
		for i, c := range op {
			if c == '/' {
				services[op[:i]] = true
				break
			}
		}
	}
	return len(services)
}

func initializeOperationMapping() map[string]string {
	return map[string]string{
		// ========== Compute ==========

		// Virtual Machines
		"Microsoft.Compute/virtualMachines/write":             "azurerm_linux_virtual_machine",
		"Microsoft.Compute/virtualMachines/delete":            "azurerm_linux_virtual_machine",
		"Microsoft.Compute/virtualMachines/start/action":      "azurerm_linux_virtual_machine",
		"Microsoft.Compute/virtualMachines/deallocate/action": "azurerm_linux_virtual_machine",
		"Microsoft.Compute/virtualMachines/restart/action":    "azurerm_linux_virtual_machine",

		// VM Scale Sets
		"Microsoft.Compute/virtualMachineScaleSets/write":  "azurerm_linux_virtual_machine_scale_set",
		"Microsoft.Compute/virtualMachineScaleSets/delete": "azurerm_linux_virtual_machine_scale_set",

		// Disks
		"Microsoft.Compute/disks/write":  "azurerm_managed_disk",
		"Microsoft.Compute/disks/delete": "azurerm_managed_disk",

		// AKS (Azure Kubernetes Service)
		"Microsoft.ContainerService/managedClusters/write":  "azurerm_kubernetes_cluster",
		"Microsoft.ContainerService/managedClusters/delete": "azurerm_kubernetes_cluster",

		// Container Instances
		"Microsoft.ContainerInstance/containerGroups/write":  "azurerm_container_group",
		"Microsoft.ContainerInstance/containerGroups/delete": "azurerm_container_group",

		// Container Registry
		"Microsoft.ContainerRegistry/registries/write":  "azurerm_container_registry",
		"Microsoft.ContainerRegistry/registries/delete": "azurerm_container_registry",

		// App Service
		"Microsoft.Web/sites/write":  "azurerm_linux_web_app",
		"Microsoft.Web/sites/delete": "azurerm_linux_web_app",

		// App Service Plans
		"Microsoft.Web/serverFarms/write":  "azurerm_service_plan",
		"Microsoft.Web/serverFarms/delete": "azurerm_service_plan",

		// Function Apps (handled by kind=functionapp in Activity Log)
		"Microsoft.Web/sites/functions/write":  "azurerm_linux_function_app",
		"Microsoft.Web/sites/functions/delete": "azurerm_linux_function_app",

		// ========== Networking ==========

		// Virtual Networks
		"Microsoft.Network/virtualNetworks/write":  "azurerm_virtual_network",
		"Microsoft.Network/virtualNetworks/delete": "azurerm_virtual_network",

		// Subnets
		"Microsoft.Network/virtualNetworks/subnets/write":  "azurerm_subnet",
		"Microsoft.Network/virtualNetworks/subnets/delete": "azurerm_subnet",

		// Network Security Groups
		"Microsoft.Network/networkSecurityGroups/write":  "azurerm_network_security_group",
		"Microsoft.Network/networkSecurityGroups/delete": "azurerm_network_security_group",

		// NSG Rules
		"Microsoft.Network/networkSecurityGroups/securityRules/write":  "azurerm_network_security_rule",
		"Microsoft.Network/networkSecurityGroups/securityRules/delete": "azurerm_network_security_rule",

		// Public IP Addresses
		"Microsoft.Network/publicIPAddresses/write":  "azurerm_public_ip",
		"Microsoft.Network/publicIPAddresses/delete": "azurerm_public_ip",

		// Load Balancers
		"Microsoft.Network/loadBalancers/write":  "azurerm_lb",
		"Microsoft.Network/loadBalancers/delete": "azurerm_lb",

		// Application Gateways
		"Microsoft.Network/applicationGateways/write":  "azurerm_application_gateway",
		"Microsoft.Network/applicationGateways/delete": "azurerm_application_gateway",

		// Network Interfaces
		"Microsoft.Network/networkInterfaces/write":  "azurerm_network_interface",
		"Microsoft.Network/networkInterfaces/delete": "azurerm_network_interface",

		// DNS Zones
		"Microsoft.Network/dnsZones/write":  "azurerm_dns_zone",
		"Microsoft.Network/dnsZones/delete": "azurerm_dns_zone",

		// DNS Record Sets
		"Microsoft.Network/dnsZones/A/write":     "azurerm_dns_a_record",
		"Microsoft.Network/dnsZones/CNAME/write": "azurerm_dns_cname_record",

		// Private DNS Zones
		"Microsoft.Network/privateDnsZones/write":  "azurerm_private_dns_zone",
		"Microsoft.Network/privateDnsZones/delete": "azurerm_private_dns_zone",

		// Route Tables
		"Microsoft.Network/routeTables/write":  "azurerm_route_table",
		"Microsoft.Network/routeTables/delete": "azurerm_route_table",

		// Front Door
		"Microsoft.Network/frontDoors/write":  "azurerm_frontdoor",
		"Microsoft.Network/frontDoors/delete": "azurerm_frontdoor",

		// ========== Storage & Database ==========

		// Storage Accounts
		"Microsoft.Storage/storageAccounts/write":  "azurerm_storage_account",
		"Microsoft.Storage/storageAccounts/delete": "azurerm_storage_account",

		// Storage Containers
		"Microsoft.Storage/storageAccounts/blobServices/containers/write":  "azurerm_storage_container",
		"Microsoft.Storage/storageAccounts/blobServices/containers/delete": "azurerm_storage_container",

		// Azure SQL Server
		"Microsoft.Sql/servers/write":  "azurerm_mssql_server",
		"Microsoft.Sql/servers/delete": "azurerm_mssql_server",

		// Azure SQL Database
		"Microsoft.Sql/servers/databases/write":  "azurerm_mssql_database",
		"Microsoft.Sql/servers/databases/delete": "azurerm_mssql_database",

		// CosmosDB
		"Microsoft.DocumentDB/databaseAccounts/write":  "azurerm_cosmosdb_account",
		"Microsoft.DocumentDB/databaseAccounts/delete": "azurerm_cosmosdb_account",

		// Redis Cache
		"Microsoft.Cache/redis/write":  "azurerm_redis_cache",
		"Microsoft.Cache/redis/delete": "azurerm_redis_cache",

		// PostgreSQL Flexible Server
		"Microsoft.DBforPostgreSQL/flexibleServers/write":  "azurerm_postgresql_flexible_server",
		"Microsoft.DBforPostgreSQL/flexibleServers/delete": "azurerm_postgresql_flexible_server",

		// MySQL Flexible Server
		"Microsoft.DBforMySQL/flexibleServers/write":  "azurerm_mysql_flexible_server",
		"Microsoft.DBforMySQL/flexibleServers/delete": "azurerm_mysql_flexible_server",

		// ========== Security & Identity ==========

		// Key Vault
		"Microsoft.KeyVault/vaults/write":  "azurerm_key_vault",
		"Microsoft.KeyVault/vaults/delete": "azurerm_key_vault",

		// Key Vault Secrets
		"Microsoft.KeyVault/vaults/secrets/write": "azurerm_key_vault_secret",

		// Key Vault Keys
		"Microsoft.KeyVault/vaults/keys/write": "azurerm_key_vault_key",

		// Managed Identity
		"Microsoft.ManagedIdentity/userAssignedIdentities/write":  "azurerm_user_assigned_identity",
		"Microsoft.ManagedIdentity/userAssignedIdentities/delete": "azurerm_user_assigned_identity",

		// Role Assignments
		"Microsoft.Authorization/roleAssignments/write":  "azurerm_role_assignment",
		"Microsoft.Authorization/roleAssignments/delete": "azurerm_role_assignment",

		// Role Definitions
		"Microsoft.Authorization/roleDefinitions/write":  "azurerm_role_definition",
		"Microsoft.Authorization/roleDefinitions/delete": "azurerm_role_definition",

		// ========== Monitoring & Management ==========

		// Monitor Action Groups
		"Microsoft.Insights/actionGroups/write":  "azurerm_monitor_action_group",
		"Microsoft.Insights/actionGroups/delete": "azurerm_monitor_action_group",

		// Monitor Metric Alerts
		"Microsoft.Insights/metricAlerts/write":  "azurerm_monitor_metric_alert",
		"Microsoft.Insights/metricAlerts/delete": "azurerm_monitor_metric_alert",

		// Log Analytics Workspace
		"Microsoft.OperationalInsights/workspaces/write":  "azurerm_log_analytics_workspace",
		"Microsoft.OperationalInsights/workspaces/delete": "azurerm_log_analytics_workspace",

		// Resource Groups
		"Microsoft.Resources/subscriptions/resourceGroups/write":  "azurerm_resource_group",
		"Microsoft.Resources/subscriptions/resourceGroups/delete": "azurerm_resource_group",
	}
}
