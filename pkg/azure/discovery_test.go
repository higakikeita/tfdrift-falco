package azure

import (
	"testing"
)

// Tests for DiscoveredResource structure

func TestDiscoveredResource_VirtualMachine(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/web-vm",
		Type:   "azurerm_virtual_machine",
		Name:   "web-vm",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"vm_size":             "Standard_B2s",
			"location":            "eastus",
			"provisioning_state":  "Succeeded",
		},
		Tags: map[string]string{
			"environment": "production",
		},
	}

	if resource.Type != "azurerm_virtual_machine" {
		t.Errorf("expected type azurerm_virtual_machine")
	}
	if resource.Attributes["vm_size"] != "Standard_B2s" {
		t.Errorf("expected vm_size Standard_B2s")
	}
	if resource.Tags["environment"] != "production" {
		t.Errorf("expected environment tag")
	}
}

func TestDiscoveredResource_StorageAccount(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/mystorageacct",
		Type:   "azurerm_storage_account",
		Name:   "mystorageacct",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":  "eastus",
			"sku_name":  "Standard_LRS",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_storage_account" {
		t.Errorf("expected type azurerm_storage_account")
	}
	if resource.Attributes["sku_name"] != "Standard_LRS" {
		t.Errorf("expected sku_name Standard_LRS")
	}
}

func TestDiscoveredResource_NetworkSecurityGroup(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Network/networkSecurityGroups/my-nsg",
		Type:   "azurerm_network_security_group",
		Name:   "my-nsg",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location": "eastus",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_network_security_group" {
		t.Errorf("expected type azurerm_network_security_group")
	}
}

func TestDiscoveredResource_VirtualNetwork(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Network/virtualNetworks/my-vnet",
		Type:   "azurerm_virtual_network",
		Name:   "my-vnet",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":     "eastus",
			"address_space": []string{"10.0.0.0/16"},
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_virtual_network" {
		t.Errorf("expected type azurerm_virtual_network")
	}
	if resource.Attributes["location"] != "eastus" {
		t.Errorf("expected location eastus")
	}
}

func TestDiscoveredResource_KubernetesCluster(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.ContainerService/managedClusters/my-aks",
		Type:   "azurerm_kubernetes_cluster",
		Name:   "my-aks",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":           "eastus",
			"kubernetes_version": "1.24.6",
			"dns_prefix":         "my-aks-dns",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_kubernetes_cluster" {
		t.Errorf("expected type azurerm_kubernetes_cluster")
	}
	if resource.Attributes["kubernetes_version"] != "1.24.6" {
		t.Errorf("expected kubernetes_version 1.24.6")
	}
}

func TestDiscoveredResource_SQLServer(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Sql/servers/my-sqlserver",
		Type:   "azurerm_mssql_server",
		Name:   "my-sqlserver",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":              "eastus",
			"version":               "12.0",
			"administrator_login":   "sqladmin",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_mssql_server" {
		t.Errorf("expected type azurerm_mssql_server")
	}
	if resource.Attributes["version"] != "12.0" {
		t.Errorf("expected version 12.0")
	}
}

func TestDiscoveredResource_KeyVault(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.KeyVault/vaults/my-keyvault",
		Type:   "azurerm_key_vault",
		Name:   "my-keyvault",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location": "eastus",
			"sku_name": "premium",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_key_vault" {
		t.Errorf("expected type azurerm_key_vault")
	}
	if resource.Attributes["sku_name"] != "premium" {
		t.Errorf("expected sku_name premium")
	}
}

func TestDiscoveredResource_AppService(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Web/sites/my-appservice",
		Type:   "azurerm_app_service",
		Name:   "my-appservice",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location": "eastus",
			"state":    "Running",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_app_service" {
		t.Errorf("expected type azurerm_app_service")
	}
	if resource.Attributes["state"] != "Running" {
		t.Errorf("expected state Running")
	}
}

func TestDiscoveredResource_ContainerRegistry(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.ContainerRegistry/registries/myregistry",
		Type:   "azurerm_container_registry",
		Name:   "myregistry",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":       "eastus",
			"sku_name":       "Premium",
			"admin_enabled":  true,
			"login_server":   "myregistry.azurecr.io",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_container_registry" {
		t.Errorf("expected type azurerm_container_registry")
	}
	if resource.Attributes["admin_enabled"] != true {
		t.Errorf("expected admin_enabled true")
	}
}

func TestDiscoveredResource_RedisCache(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Cache/redis/my-redis",
		Type:   "azurerm_redis_cache",
		Name:   "my-redis",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":   "eastus",
			"sku_name":   "Premium",
			"hostname":   "my-redis.redis.cache.windows.net",
			"port":       6379,
			"ssl_port":   6380,
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_redis_cache" {
		t.Errorf("expected type azurerm_redis_cache")
	}
	if resource.Attributes["port"] != 6379 {
		t.Errorf("expected port 6379")
	}
}

func TestDiscoveredResource_LoadBalancer(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Network/loadBalancers/my-lb",
		Type:   "azurerm_lb",
		Name:   "my-lb",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location": "eastus",
			"sku_name": "Standard",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_lb" {
		t.Errorf("expected type azurerm_lb")
	}
	if resource.Attributes["sku_name"] != "Standard" {
		t.Errorf("expected sku_name Standard")
	}
}

func TestDiscoveredResource_PublicIP(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Network/publicIPAddresses/my-pip",
		Type:   "azurerm_public_ip",
		Name:   "my-pip",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":           "eastus",
			"allocation_method":  "Static",
			"ip_address":         "20.25.30.1",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_public_ip" {
		t.Errorf("expected type azurerm_public_ip")
	}
	if resource.Attributes["allocation_method"] != "Static" {
		t.Errorf("expected allocation_method Static")
	}
}

func TestDiscoveredResource_CosmosDB(t *testing.T) {
	resource := &DiscoveredResource{
		ID:     "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.DocumentDB/databaseAccounts/my-cosmosdb",
		Type:   "azurerm_cosmosdb_account",
		Name:   "my-cosmosdb",
		Region: "eastus",
		Attributes: map[string]interface{}{
			"location":             "eastus",
			"offer_type":           "Standard",
			"consistency_level":    "Session",
		},
		Tags: map[string]string{},
	}

	if resource.Type != "azurerm_cosmosdb_account" {
		t.Errorf("expected type azurerm_cosmosdb_account")
	}
	if resource.Attributes["consistency_level"] != "Session" {
		t.Errorf("expected consistency_level Session")
	}
}

// Tests for DriftResult

func TestDriftResult_AllEmpty(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{},
		MissingResources:   []*TerraformResource{},
		ModifiedResources:  []*ResourceDiff{},
	}

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources")
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources")
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources")
	}
}

func TestDriftResult_Mixed(t *testing.T) {
	result := &DriftResult{
		UnmanagedResources: []*DiscoveredResource{
			{
				ID:   "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Storage/storageAccounts/unmanaged",
				Type: "azurerm_storage_account",
			},
		},
		MissingResources: []*TerraformResource{
			{
				Type: "azurerm_virtual_machine",
				Name: "deleted-vm",
			},
		},
		ModifiedResources: []*ResourceDiff{
			{
				ResourceID:   "my-vnet",
				ResourceType: "azurerm_virtual_network",
			},
		},
	}

	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource")
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource")
	}
	if len(result.ModifiedResources) != 1 {
		t.Errorf("expected 1 modified resource")
	}
}

// Tests for TerraformResource

func TestTerraformResource_VirtualMachine(t *testing.T) {
	resource := &TerraformResource{
		Type: "azurerm_virtual_machine",
		Name: "web-vm",
		Attributes: map[string]interface{}{
			"vm_size": "Standard_D2s_v3",
			"name":    "web-vm",
		},
	}

	if resource.Type != "azurerm_virtual_machine" {
		t.Errorf("expected type azurerm_virtual_machine")
	}
	if resource.Name != "web-vm" {
		t.Errorf("expected name web-vm")
	}
}

// Tests for ResourceDiff

func TestResourceDiff_SingleDifference(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "my-vm",
		ResourceType: "azurerm_virtual_machine",
		TerraformState: map[string]interface{}{
			"vm_size": "Standard_B2s",
		},
		ActualState: map[string]interface{}{
			"vm_size": "Standard_D2s_v3",
		},
		Differences: []FieldDiff{
			{
				Field:          "vm_size",
				TerraformValue: "Standard_B2s",
				ActualValue:    "Standard_D2s_v3",
			},
		},
	}

	if diff.ResourceID != "my-vm" {
		t.Errorf("expected ResourceID my-vm")
	}
	if len(diff.Differences) != 1 {
		t.Errorf("expected 1 difference")
	}
}

func TestResourceDiff_MultipleDifferences(t *testing.T) {
	diff := &ResourceDiff{
		ResourceID:   "my-storage",
		ResourceType: "azurerm_storage_account",
		TerraformState: map[string]interface{}{
			"location": "westus",
			"sku_name": "Standard_LRS",
		},
		ActualState: map[string]interface{}{
			"location": "eastus",
			"sku_name": "Standard_GRS",
		},
		Differences: []FieldDiff{
			{
				Field:          "location",
				TerraformValue: "westus",
				ActualValue:    "eastus",
			},
			{
				Field:          "sku_name",
				TerraformValue: "Standard_LRS",
				ActualValue:    "Standard_GRS",
			},
		},
	}

	if len(diff.Differences) != 2 {
		t.Errorf("expected 2 differences")
	}
}

// Tests for FieldDiff

func TestFieldDiff_StringValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "location",
		TerraformValue: "westus",
		ActualValue:    "eastus",
	}

	if diff.Field != "location" {
		t.Errorf("expected field location")
	}
	if diff.TerraformValue != "westus" {
		t.Errorf("expected westus")
	}
}

func TestFieldDiff_NumericValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "port",
		TerraformValue: 6379,
		ActualValue:    6380,
	}

	if diff.TerraformValue != 6379 {
		t.Errorf("expected 6379")
	}
	if diff.ActualValue != 6380 {
		t.Errorf("expected 6380")
	}
}

func TestFieldDiff_BooleanValues(t *testing.T) {
	diff := FieldDiff{
		Field:          "admin_enabled",
		TerraformValue: false,
		ActualValue:    true,
	}

	if diff.TerraformValue != false {
		t.Errorf("expected false")
	}
	if diff.ActualValue != true {
		t.Errorf("expected true")
	}
}

