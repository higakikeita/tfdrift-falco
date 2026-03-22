package azure

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceMapper(t *testing.T) {
	mapper := NewResourceMapper()
	require.NotNil(t, mapper)
	assert.NotNil(t, mapper.operationToResource)
}

func TestResourceMapper_MapOperationToResource(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		name         string
		operation    string
		expectedType string
	}{
		// Compute
		{"VM Write", "Microsoft.Compute/virtualMachines/write", "azurerm_linux_virtual_machine"},
		{"VM Delete", "Microsoft.Compute/virtualMachines/delete", "azurerm_linux_virtual_machine"},
		{"Disk Write", "Microsoft.Compute/disks/write", "azurerm_managed_disk"},

		// Networking
		{"VNet Write", "Microsoft.Network/virtualNetworks/write", "azurerm_virtual_network"},
		{"NSG Write", "Microsoft.Network/networkSecurityGroups/write", "azurerm_network_security_group"},
		{"Public IP Write", "Microsoft.Network/publicIPAddresses/write", "azurerm_public_ip"},
		{"Load Balancer Write", "Microsoft.Network/loadBalancers/write", "azurerm_lb"},
		{"VPN Gateway Write", "Microsoft.Network/virtualNetworkGateways/write", "azurerm_virtual_network_gateway"},
		{"Firewall Write", "Microsoft.Network/azureFirewalls/write", "azurerm_firewall"},
		{"NAT Gateway Write", "Microsoft.Network/natGateways/write", "azurerm_nat_gateway"},
		{"Private Endpoint Write", "Microsoft.Network/privateEndpoints/write", "azurerm_private_endpoint"},

		// Storage & Database
		{"Storage Account Write", "Microsoft.Storage/storageAccounts/write", "azurerm_storage_account"},
		{"SQL Server Write", "Microsoft.Sql/servers/write", "azurerm_mssql_server"},
		{"CosmosDB Write", "Microsoft.DocumentDB/databaseAccounts/write", "azurerm_cosmosdb_account"},
		{"Redis Write", "Microsoft.Cache/redis/write", "azurerm_redis_cache"},

		// Security & Identity
		{"Key Vault Write", "Microsoft.KeyVault/vaults/write", "azurerm_key_vault"},
		{"Role Assignment Write", "Microsoft.Authorization/roleAssignments/write", "azurerm_role_assignment"},
		{"Managed Identity Write", "Microsoft.ManagedIdentity/userAssignedIdentities/write", "azurerm_user_assigned_identity"},

		// Monitoring
		{"Action Group Write", "Microsoft.Insights/actionGroups/write", "azurerm_monitor_action_group"},
		{"Diagnostic Settings Write", "Microsoft.Insights/diagnosticSettings/write", "azurerm_monitor_diagnostic_setting"},

		// Analytics & Data
		{"Event Hub Namespace Write", "Microsoft.EventHub/namespaces/write", "azurerm_eventhub_namespace"},
		{"Service Bus Namespace Write", "Microsoft.ServiceBus/namespaces/write", "azurerm_servicebus_namespace"},
		{"Data Factory Write", "Microsoft.DataFactory/factories/write", "azurerm_data_factory"},
		{"Synapse Workspace Write", "Microsoft.Synapse/workspaces/write", "azurerm_synapse_workspace"},

		// Integration
		{"API Management Write", "Microsoft.ApiManagement/service/write", "azurerm_api_management"},
		{"Logic App Write", "Microsoft.Logic/workflows/write", "azurerm_logic_app_workflow"},

		// Unknown operation
		{"Unknown Operation", "Microsoft.Unknown/unknown/write", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MapOperationToResource(tt.operation)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestResourceMapper_GetAllSupportedOperations(t *testing.T) {
	mapper := NewResourceMapper()
	ops := mapper.GetAllSupportedOperations()

	assert.NotEmpty(t, ops)
	assert.Greater(t, len(ops), 50, "Should support at least 50 operations")

	// Check that some key operations are present
	opSet := make(map[string]bool)
	for _, op := range ops {
		opSet[op] = true
	}

	expectedOps := []string{
		"Microsoft.Compute/virtualMachines/write",
		"Microsoft.Network/virtualNetworks/write",
		"Microsoft.Storage/storageAccounts/write",
		"Microsoft.KeyVault/vaults/write",
		"Microsoft.Authorization/roleAssignments/write",
	}

	for _, op := range expectedOps {
		assert.True(t, opSet[op], "Operation %s should be present", op)
	}
}

func TestResourceMapper_GetSupportedServiceCount(t *testing.T) {
	mapper := NewResourceMapper()
	count := mapper.GetSupportedServiceCount()

	// Should support multiple Azure services
	assert.Greater(t, count, 10, "Should support at least 10 Azure services")
	assert.Less(t, count, 100, "Should not exceed 100 services")
}

func TestResourceMapper_AllOperationsHaveMappings(t *testing.T) {
	mapper := NewResourceMapper()
	ops := mapper.GetAllSupportedOperations()

	// Every operation should have a valid mapping
	for _, op := range ops {
		resourceType := mapper.MapOperationToResource(op)
		assert.NotEmpty(t, resourceType, "Operation %s should have a mapping", op)
		assert.True(t, len(resourceType) > 0, "Resource type for %s should not be empty", op)
	}
}

func TestResourceMapper_UnmappedOperationReturnsEmpty(t *testing.T) {
	mapper := NewResourceMapper()

	result := mapper.MapOperationToResource("Microsoft.Unknown/unknownService/write")
	assert.Equal(t, "", result)

	result = mapper.MapOperationToResource("Invalid/Operation")
	assert.Equal(t, "", result)
}

func TestResourceMapper_OperationsMatchPattern(t *testing.T) {
	mapper := NewResourceMapper()
	ops := mapper.GetAllSupportedOperations()

	// All operations should follow Microsoft pattern
	for _, op := range ops {
		assert.True(t, len(op) > 0, "Operation should not be empty")
		assert.Contains(t, op, "/", "Operation should contain forward slashes")
		// Most operations should start with Microsoft or be a generic term
		assert.True(t,
			op[0] >= 'A' && op[0] <= 'Z',
			"Operation %s should start with capital letter", op)
	}
}
