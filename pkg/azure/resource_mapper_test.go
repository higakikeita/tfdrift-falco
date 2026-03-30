package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceMapper_MapEventToResource(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		operation string
		want      string
	}{
		// Compute
		{"Microsoft.Compute/virtualMachines/write", "azurerm_virtual_machine"},
		{"Microsoft.Compute/virtualMachines/delete", "azurerm_virtual_machine"},
		{"Microsoft.Compute/virtualMachines/start/action", "azurerm_virtual_machine"},
		{"Microsoft.Compute/virtualMachineScaleSets/write", "azurerm_virtual_machine_scale_set"},
		{"Microsoft.Compute/disks/write", "azurerm_managed_disk"},

		// Networking - NSGs
		{"Microsoft.Network/networkSecurityGroups/write", "azurerm_network_security_group"},
		{"Microsoft.Network/networkSecurityGroups/delete", "azurerm_network_security_group"},
		{"Microsoft.Network/networkSecurityGroups/securityRules/write", "azurerm_network_security_rule"},
		{"Microsoft.Network/networkSecurityGroups/securityRules/delete", "azurerm_network_security_rule"},

		// Networking - Virtual Networks
		{"Microsoft.Network/virtualNetworks/write", "azurerm_virtual_network"},
		{"Microsoft.Network/virtualNetworks/delete", "azurerm_virtual_network"},
		{"Microsoft.Network/virtualNetworks/subnets/write", "azurerm_subnet"},
		{"Microsoft.Network/virtualNetworks/subnets/delete", "azurerm_subnet"},

		// Networking - Load Balancer
		{"Microsoft.Network/loadBalancers/write", "azurerm_lb"},
		{"Microsoft.Network/loadBalancers/delete", "azurerm_lb"},
		{"Microsoft.Network/loadBalancers/backendAddressPools/write", "azurerm_lb_backend_address_pool"},

		// Networking - Public IP
		{"Microsoft.Network/publicIPAddresses/write", "azurerm_public_ip"},
		{"Microsoft.Network/publicIPAddresses/delete", "azurerm_public_ip"},

		// Networking - Network Interfaces
		{"Microsoft.Network/networkInterfaces/write", "azurerm_network_interface"},
		{"Microsoft.Network/networkInterfaces/delete", "azurerm_network_interface"},

		// Networking - Route Tables
		{"Microsoft.Network/routeTables/write", "azurerm_route_table"},
		{"Microsoft.Network/routeTables/delete", "azurerm_route_table"},
		{"Microsoft.Network/routeTables/routes/write", "azurerm_route"},
		{"Microsoft.Network/routeTables/routes/delete", "azurerm_route"},

		// Networking - Application Gateway
		{"Microsoft.Network/applicationGateways/write", "azurerm_application_gateway"},
		{"Microsoft.Network/applicationGateways/delete", "azurerm_application_gateway"},

		// Networking - Firewall
		{"Microsoft.Network/azureFirewalls/write", "azurerm_firewall"},
		{"Microsoft.Network/azureFirewalls/delete", "azurerm_firewall"},
		{"Microsoft.Network/firewallPolicies/write", "azurerm_firewall_policy"},

		// Networking - Private Endpoints
		{"Microsoft.Network/privateEndpoints/write", "azurerm_private_endpoint"},
		{"Microsoft.Network/privateEndpoints/delete", "azurerm_private_endpoint"},

		// Networking - CDN
		{"Microsoft.Cdn/profiles/write", "azurerm_cdn_profile"},
		{"Microsoft.Cdn/profiles/endpoints/write", "azurerm_cdn_endpoint"},

		// Networking - Front Door
		{"Microsoft.Network/frontDoors/write", "azurerm_frontdoor"},
		{"Microsoft.Network/frontDoors/delete", "azurerm_frontdoor"},

		// Storage
		{"Microsoft.Storage/storageAccounts/write", "azurerm_storage_account"},
		{"Microsoft.Storage/storageAccounts/delete", "azurerm_storage_account"},
		{"Microsoft.Storage/storageAccounts/blobServices/write", "azurerm_storage_blob"},
		{"Microsoft.Storage/storageAccounts/blobServices/containers/write", "azurerm_storage_container"},

		// SQL
		{"Microsoft.Sql/servers/write", "azurerm_mssql_server"},
		{"Microsoft.Sql/servers/databases/write", "azurerm_mssql_database"},
		{"Microsoft.Sql/servers/databases/delete", "azurerm_mssql_database"},
		{"Microsoft.Sql/servers/failoverGroups/write", "azurerm_mssql_failover_group"},
		{"Microsoft.Sql/servers/elasticPools/write", "azurerm_mssql_elasticpool"},

		// MySQL
		{"Microsoft.DBforMySQL/servers/write", "azurerm_mysql_server"},
		{"Microsoft.DBforMySQL/servers/databases/write", "azurerm_mysql_database"},

		// PostgreSQL
		{"Microsoft.DBforPostgreSQL/servers/write", "azurerm_postgresql_server"},
		{"Microsoft.DBforPostgreSQL/servers/databases/write", "azurerm_postgresql_database"},

		// Cosmos DB
		{"Microsoft.DocumentDB/databaseAccounts/write", "azurerm_cosmosdb_account"},
		{"Microsoft.DocumentDB/databaseAccounts/delete", "azurerm_cosmosdb_account"},
		{"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/write", "azurerm_cosmosdb_sql_database"},
		{"Microsoft.DocumentDB/databaseAccounts/sqlDatabases/containers/write", "azurerm_cosmosdb_sql_container"},

		// Key Vault
		{"Microsoft.KeyVault/vaults/write", "azurerm_key_vault"},
		{"Microsoft.KeyVault/vaults/delete", "azurerm_key_vault"},
		{"Microsoft.KeyVault/vaults/secrets/write", "azurerm_key_vault_secret"},
		{"Microsoft.KeyVault/vaults/secrets/delete", "azurerm_key_vault_secret"},
		{"Microsoft.KeyVault/vaults/keys/write", "azurerm_key_vault_key"},
		{"Microsoft.KeyVault/vaults/keys/delete", "azurerm_key_vault_key"},

		// Managed Identities
		{"Microsoft.ManagedIdentity/userAssignedIdentities/write", "azurerm_user_assigned_identity"},
		{"Microsoft.ManagedIdentity/userAssignedIdentities/delete", "azurerm_user_assigned_identity"},

		// App Service
		{"Microsoft.Web/sites/write", "azurerm_app_service"},
		{"Microsoft.Web/sites/delete", "azurerm_app_service"},
		{"Microsoft.Web/serverfarms/write", "azurerm_app_service_plan"},
		{"Microsoft.Web/serverfarms/delete", "azurerm_app_service_plan"},

		// Kubernetes
		{"Microsoft.ContainerService/managedClusters/write", "azurerm_kubernetes_cluster"},
		{"Microsoft.ContainerService/managedClusters/delete", "azurerm_kubernetes_cluster"},

		// Container Registry
		{"Microsoft.ContainerRegistry/registries/write", "azurerm_container_registry"},
		{"Microsoft.ContainerRegistry/registries/delete", "azurerm_container_registry"},

		// Container Instances
		{"Microsoft.ContainerInstance/containerGroups/write", "azurerm_container_group"},
		{"Microsoft.ContainerInstance/containerGroups/delete", "azurerm_container_group"},

		// Service Bus
		{"Microsoft.ServiceBus/namespaces/write", "azurerm_servicebus_namespace"},
		{"Microsoft.ServiceBus/namespaces/queues/write", "azurerm_servicebus_queue"},
		{"Microsoft.ServiceBus/namespaces/topics/write", "azurerm_servicebus_topic"},
		{"Microsoft.ServiceBus/namespaces/topics/subscriptions/write", "azurerm_servicebus_subscription"},

		// Event Grid
		{"Microsoft.EventGrid/topics/write", "azurerm_eventgrid_topic"},
		{"Microsoft.EventGrid/topics/delete", "azurerm_eventgrid_topic"},
		{"Microsoft.EventGrid/domains/write", "azurerm_eventgrid_domain"},

		// Event Hub
		{"Microsoft.EventHub/namespaces/write", "azurerm_eventhub_namespace"},
		{"Microsoft.EventHub/namespaces/eventhubs/write", "azurerm_eventhub"},

		// Monitoring
		{"Microsoft.Insights/metricAlerts/write", "azurerm_monitor_metric_alert"},
		{"Microsoft.Insights/metricAlerts/delete", "azurerm_monitor_metric_alert"},
		{"Microsoft.Insights/scheduledQueryRules/write", "azurerm_monitor_scheduled_query_rules_alert"},
		{"Microsoft.Insights/actionGroups/write", "azurerm_monitor_action_group"},
		{"Microsoft.Insights/diagnosticSettings/write", "azurerm_monitor_diagnostic_setting"},

		// Log Analytics
		{"Microsoft.OperationalInsights/workspaces/write", "azurerm_log_analytics_workspace"},
		{"Microsoft.OperationalInsights/workspaces/delete", "azurerm_log_analytics_workspace"},

		// Application Insights
		{"Microsoft.Insights/components/write", "azurerm_application_insights"},

		// Cache
		{"Microsoft.Cache/redis/write", "azurerm_redis_cache"},
		{"Microsoft.Cache/redis/delete", "azurerm_redis_cache"},

		// DNS
		{"Microsoft.Network/dnszones/write", "azurerm_dns_zone"},
		{"Microsoft.Network/dnszones/delete", "azurerm_dns_zone"},
		{"Microsoft.Network/dnszones/recordSets/write", "azurerm_dns_a_record"},
		{"Microsoft.Network/privateDnsZones/write", "azurerm_private_dns_zone"},

		// API Management
		{"Microsoft.ApiManagement/service/write", "azurerm_api_management"},
		{"Microsoft.ApiManagement/service/apis/write", "azurerm_api_management_api"},
		{"Microsoft.ApiManagement/service/apis/operations/write", "azurerm_api_management_api_operation"},

		// Authorization & Governance
		{"Microsoft.Authorization/roleAssignments/write", "azurerm_role_assignment"},
		{"Microsoft.Authorization/roleAssignments/delete", "azurerm_role_assignment"},
		{"Microsoft.Authorization/policyAssignments/write", "azurerm_resource_group_policy_assignment"},
		{"Microsoft.Authorization/policyAssignments/delete", "azurerm_resource_group_policy_assignment"},
		{"Microsoft.Authorization/locks/write", "azurerm_management_lock"},
		{"Microsoft.Authorization/locks/delete", "azurerm_management_lock"},

		// Batch
		{"Microsoft.Batch/batchAccounts/write", "azurerm_batch_account"},
		{"Microsoft.Batch/batchAccounts/pools/write", "azurerm_batch_pool"},

		// Data Factory
		{"Microsoft.DataFactory/factories/write", "azurerm_data_factory"},
		{"Microsoft.DataFactory/factories/pipelines/write", "azurerm_data_factory_pipeline"},

		// Search
		{"Microsoft.Search/searchServices/write", "azurerm_search_service"},

		// Stream Analytics
		{"Microsoft.StreamAnalytics/streamingjobs/write", "azurerm_stream_analytics_job"},

		// Synapse
		{"Microsoft.Synapse/workspaces/write", "azurerm_synapse_workspace"},
		{"Microsoft.Synapse/workspaces/bigDataPools/write", "azurerm_synapse_spark_pool"},

		// Logic Apps
		{"Microsoft.Logic/workflows/write", "azurerm_logic_app_workflow"},

		// Automation
		{"Microsoft.Automation/automationAccounts/write", "azurerm_automation_account"},
		{"Microsoft.Automation/automationAccounts/runbooks/write", "azurerm_automation_runbook"},

		// Resource Groups
		{"Microsoft.Resources/resourceGroups/write", "azurerm_resource_group"},

		// Non-existent operation
		{"nonexistent/operation", ""},
	}

	for _, tt := range tests {
		got := mapper.MapEventToResource(tt.operation)
		assert.Equal(t, tt.want, got, "operation: %s", tt.operation)
	}
}

func TestResourceMapper_GetAllSupportedEvents(t *testing.T) {
	mapper := NewResourceMapper()
	events := mapper.GetAllSupportedEvents()

	// Should have 185+ events (after deduplication of Microsoft.Web/sites entries)
	assert.Greater(t, len(events), 185, "Expected more than 185 supported events")

	// All events should be non-empty strings
	for _, event := range events {
		assert.NotEmpty(t, event, "Event should not be empty")
	}
}

func TestResourceMapper_GetResourceTypesForService(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		service string
		minCount int
	}{
		{"Microsoft.Compute/", 5},
		{"Microsoft.Network/", 15},
		{"Microsoft.Storage/", 3},
		{"Microsoft.Sql/", 4},
		{"Microsoft.KeyVault/", 4},
		{"Microsoft.Web/", 3},
		{"Microsoft.ContainerService/", 1},
	}

	for _, tt := range tests {
		resourceTypes := mapper.GetResourceTypesForService(tt.service)
		assert.GreaterOrEqual(t, len(resourceTypes), tt.minCount,
			"service %s should have at least %d resource types, got %d",
			tt.service, tt.minCount, len(resourceTypes))
	}
}

func TestResourceMapper_CreateNewResourceMapper(t *testing.T) {
	mapper := NewResourceMapper()
	assert.NotNil(t, mapper, "Mapper should not be nil")
	assert.Greater(t, len(mapper.eventToResource), 0, "Mapper should have events")
}

func TestResourceMapper_MapEventToResource_NotFound(t *testing.T) {
	mapper := NewResourceMapper()

	result := mapper.MapEventToResource("Microsoft.Unknown/unknown/write")
	assert.Equal(t, "", result, "Expected empty string for unmapped event")
}

func TestResourceMapper_MapEventToResource_CaseSensitive(t *testing.T) {
	mapper := NewResourceMapper()

	// Exact match
	result := mapper.MapEventToResource("Microsoft.Compute/virtualMachines/write")
	assert.Equal(t, "azurerm_virtual_machine", result)

	// Different case should not match (case-sensitive)
	result = mapper.MapEventToResource("microsoft.compute/virtualmachines/write")
	assert.Equal(t, "", result)
}

func TestResourceMapper_GetResourceTypesForService_AllServices(t *testing.T) {
	mapper := NewResourceMapper()

	services := []string{
		"Microsoft.Compute/",
		"Microsoft.Network/",
		"Microsoft.Storage/",
		"Microsoft.Sql/",
		"Microsoft.DocumentDB/",
		"Microsoft.KeyVault/",
		"Microsoft.Web/",
		"Microsoft.Cache/",
	}

	for _, service := range services {
		types := mapper.GetResourceTypesForService(service)
		assert.Greater(t, len(types), 0, "Service %s should have at least one resource type", service)

		// Verify no empty strings
		for _, rt := range types {
			assert.NotEmpty(t, rt, "Resource type should not be empty for service %s", service)
		}
	}
}

func TestResourceMapper_GetResourceTypesForService_NonExistent(t *testing.T) {
	mapper := NewResourceMapper()

	types := mapper.GetResourceTypesForService("Microsoft.NonExistent/")
	assert.Equal(t, 0, len(types), "Should return empty list for non-existent service")
}

func TestResourceMapper_GetAllSupportedEvents_Uniqueness(t *testing.T) {
	mapper := NewResourceMapper()
	events := mapper.GetAllSupportedEvents()

	// Check for duplicates
	seen := make(map[string]bool)
	duplicates := 0
	for _, event := range events {
		if seen[event] {
			duplicates++
		}
		seen[event] = true
	}

	assert.Equal(t, 0, duplicates, "Should have no duplicate events")
}

func TestResourceMapper_GetAllSupportedEvents_NoEmpty(t *testing.T) {
	mapper := NewResourceMapper()
	events := mapper.GetAllSupportedEvents()

	for _, event := range events {
		assert.NotEmpty(t, event, "Event should not be empty")
		assert.True(t, len(event) > 0, "Event should have non-zero length")
	}
}

func TestResourceMapper_MultipleMappers(t *testing.T) {
	mapper1 := NewResourceMapper()
	mapper2 := NewResourceMapper()

	// Both should have the same mappings
	assert.Equal(t, len(mapper1.eventToResource), len(mapper2.eventToResource))

	// Test a specific mapping on both
	assert.Equal(t,
		mapper1.MapEventToResource("Microsoft.Compute/virtualMachines/write"),
		mapper2.MapEventToResource("Microsoft.Compute/virtualMachines/write"),
	)
}

func TestResourceMapper_CommonOperations(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		operation        string
		expectedResource string
	}{
		// Write operations
		{"Microsoft.Compute/virtualMachines/write", "azurerm_virtual_machine"},
		{"Microsoft.Network/virtualNetworks/write", "azurerm_virtual_network"},
		{"Microsoft.Storage/storageAccounts/write", "azurerm_storage_account"},

		// Delete operations
		{"Microsoft.Compute/virtualMachines/delete", "azurerm_virtual_machine"},
		{"Microsoft.Network/networkSecurityGroups/delete", "azurerm_network_security_group"},

		// Action operations
		{"Microsoft.Compute/virtualMachines/start/action", "azurerm_virtual_machine"},
		{"Microsoft.Compute/virtualMachines/powerOff/action", "azurerm_virtual_machine"},
		{"Microsoft.Compute/virtualMachines/deallocate/action", "azurerm_virtual_machine"},
	}

	for _, tt := range tests {
		result := mapper.MapEventToResource(tt.operation)
		assert.Equal(t, tt.expectedResource, result, "Operation %s should map to %s", tt.operation, tt.expectedResource)
	}
}

func TestResourceMapper_NestedResources(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		operation        string
		expectedResource string
	}{
		{"Microsoft.Network/networkSecurityGroups/securityRules/write", "azurerm_network_security_rule"},
		{"Microsoft.Network/loadBalancers/backendAddressPools/write", "azurerm_lb_backend_address_pool"},
		{"Microsoft.Network/virtualNetworks/subnets/write", "azurerm_subnet"},
		{"Microsoft.Sql/servers/databases/write", "azurerm_mssql_database"},
		{"Microsoft.KeyVault/vaults/secrets/write", "azurerm_key_vault_secret"},
	}

	for _, tt := range tests {
		result := mapper.MapEventToResource(tt.operation)
		assert.Equal(t, tt.expectedResource, result, "Nested operation %s should map to %s", tt.operation, tt.expectedResource)
	}
}

func TestResourceMapper_ServiceCoverage(t *testing.T) {
	mapper := NewResourceMapper()

	// Verify coverage for major Azure services
	requiredServices := map[string]int{
		"Microsoft.Compute/": 5,
		"Microsoft.Network/": 15,
		"Microsoft.Storage/": 3,
		"Microsoft.Sql/": 4,
		"Microsoft.KeyVault/": 4,
		"Microsoft.Web/": 3,
		"Microsoft.ContainerService/": 1,
	}

	for service, minCount := range requiredServices {
		types := mapper.GetResourceTypesForService(service)
		assert.GreaterOrEqual(t, len(types), minCount, "Service %s should have at least %d resource types", service, minCount)
	}
}
