package azure

import (
	"testing"
)

func TestNewResourceMapper(t *testing.T) {
	mapper := NewResourceMapper()
	if mapper == nil {
		t.Fatal("NewResourceMapper returned nil")
	}
}

func TestMapOperationToResource(t *testing.T) {
	mapper := NewResourceMapper()

	tests := []struct {
		name      string
		operation string
		want      string
	}{
		// Compute
		{"VM write", "Microsoft.Compute/virtualMachines/write", "azurerm_linux_virtual_machine"},
		{"VM delete", "Microsoft.Compute/virtualMachines/delete", "azurerm_linux_virtual_machine"},
		{"AKS write", "Microsoft.ContainerService/managedClusters/write", "azurerm_kubernetes_cluster"},
		{"App Service write", "Microsoft.Web/sites/write", "azurerm_linux_web_app"},

		// Networking
		{"VNet write", "Microsoft.Network/virtualNetworks/write", "azurerm_virtual_network"},
		{"NSG write", "Microsoft.Network/networkSecurityGroups/write", "azurerm_network_security_group"},
		{"NSG Rule write", "Microsoft.Network/networkSecurityGroups/securityRules/write", "azurerm_network_security_rule"},
		{"LB write", "Microsoft.Network/loadBalancers/write", "azurerm_lb"},
		{"DNS Zone write", "Microsoft.Network/dnsZones/write", "azurerm_dns_zone"},

		// Storage & Database
		{"Storage Account write", "Microsoft.Storage/storageAccounts/write", "azurerm_storage_account"},
		{"SQL Server write", "Microsoft.Sql/servers/write", "azurerm_mssql_server"},
		{"SQL DB write", "Microsoft.Sql/servers/databases/write", "azurerm_mssql_database"},
		{"CosmosDB write", "Microsoft.DocumentDB/databaseAccounts/write", "azurerm_cosmosdb_account"},
		{"Redis write", "Microsoft.Cache/redis/write", "azurerm_redis_cache"},
		{"PostgreSQL write", "Microsoft.DBforPostgreSQL/flexibleServers/write", "azurerm_postgresql_flexible_server"},

		// Security
		{"Key Vault write", "Microsoft.KeyVault/vaults/write", "azurerm_key_vault"},
		{"Managed Identity write", "Microsoft.ManagedIdentity/userAssignedIdentities/write", "azurerm_user_assigned_identity"},
		{"Role Assignment write", "Microsoft.Authorization/roleAssignments/write", "azurerm_role_assignment"},

		// Unknown
		{"Unknown operation", "Microsoft.Unknown/stuff/write", ""},
		{"Empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapper.MapOperationToResource(tt.operation)
			if got != tt.want {
				t.Errorf("MapOperationToResource(%q) = %q, want %q", tt.operation, got, tt.want)
			}
		})
	}
}

func TestGetSupportedServiceCount(t *testing.T) {
	mapper := NewResourceMapper()
	count := mapper.GetSupportedServiceCount()

	if count < 10 {
		t.Errorf("Expected at least 10 Azure services, got %d", count)
	}
}

func TestGetAllSupportedOperations(t *testing.T) {
	mapper := NewResourceMapper()
	ops := mapper.GetAllSupportedOperations()

	if len(ops) < 50 {
		t.Errorf("Expected at least 50 operations, got %d", len(ops))
	}
}
