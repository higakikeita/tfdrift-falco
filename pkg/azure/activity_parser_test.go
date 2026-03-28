package azure

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestActivityParser_Parse(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		name     string
		source   string
		fields   map[string]string
		wantNil  bool
		wantType string
	}{
		{
			name:   "VM write event",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.Compute/virtualMachines/write",
				"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
				"azure.caller":           "user@example.com",
				"azure.resourceLocation": "eastus",
			},
			wantNil:  false,
			wantType: "azurerm_virtual_machine",
		},
		{
			name:   "NSG rule write",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.Network/networkSecurityGroups/securityRules/write",
				"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Network/networkSecurityGroups/nsg-test/securityRules/allow-ssh",
				"azure.caller":           "admin@example.com",
				"azure.resourceLocation": "westus2",
			},
			wantNil:  false,
			wantType: "azurerm_network_security_rule",
		},
		{
			name:   "Storage account write",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.Storage/storageAccounts/write",
				"azure.resourceId":       "/subscriptions/sub-456/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/mystg",
				"azure.caller":           "deployer@example.com",
				"azure.resourceLocation": "centralus",
			},
			wantNil:  false,
			wantType: "azurerm_storage_account",
		},
		{
			name:   "SQL database write",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.Sql/servers/databases/write",
				"azure.resourceId":       "/subscriptions/sub-789/resourceGroups/rg-sql/providers/Microsoft.Sql/servers/mysqlserver/databases/mydb",
				"azure.caller":           "dba@example.com",
				"azure.resourceLocation": "southcentralus",
			},
			wantNil:  false,
			wantType: "azurerm_mssql_database",
		},
		{
			name:    "Non-azure source",
			source:  "aws_cloudtrail",
			fields:  map[string]string{},
			wantNil: true,
		},
		{
			name:    "Nil response",
			wantNil: true,
		},
		{
			name:   "Irrelevant read operation",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName": "Microsoft.Compute/virtualMachines/read",
			},
			wantNil: true,
		},
		{
			name:   "Missing operation name",
			source: "azure_activity",
			fields: map[string]string{
				"azure.resourceId": "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
			},
			wantNil: true,
		},
		{
			name:   "Missing resource ID",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName": "Microsoft.Compute/virtualMachines/write",
			},
			wantNil: true,
		},
		{
			name:   "VM start action",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.Compute/virtualMachines/start/action",
				"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
				"azure.caller":           "operator@example.com",
				"azure.resourceLocation": "eastus",
			},
			wantNil:  false,
			wantType: "azurerm_virtual_machine",
		},
		{
			name:   "Key Vault secret write",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.KeyVault/vaults/secrets/write",
				"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/rg-vault/providers/Microsoft.KeyVault/vaults/myvault/secrets/mysecret",
				"azure.caller":           "admin@example.com",
				"azure.resourceLocation": "eastus",
			},
			wantNil:  false,
			wantType: "azurerm_key_vault_secret",
		},
		{
			name:   "AKS cluster write",
			source: "azure_activity",
			fields: map[string]string{
				"azure.operationName":    "Microsoft.ContainerService/managedClusters/write",
				"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/rg-aks/providers/Microsoft.ContainerService/managedClusters/myakscluster",
				"azure.caller":           "devops@example.com",
				"azure.resourceLocation": "westus2",
			},
			wantNil:  false,
			wantType: "azurerm_kubernetes_cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var res *outputs.Response
			if tt.source != "" {
				res = &outputs.Response{
					Source:       tt.source,
					OutputFields: tt.fields,
				}
			}

			event := parser.Parse(res)
			if tt.wantNil {
				assert.Nil(t, event, "Expected nil event")
			} else {
				assert.NotNil(t, event, "Expected non-nil event")
				assert.Equal(t, "azure", event.Provider, "Provider should be azure")
				assert.Equal(t, tt.wantType, event.ResourceType, "Resource type mismatch")
				assert.NotEmpty(t, event.ResourceID, "Resource ID should not be empty")
			}
		})
	}
}

func TestActivityParser_ExtractResourceName(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		want       string
	}{
		{
			"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
			"my-vm",
		},
		{
			"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Network/networkSecurityGroups/nsg-test/securityRules/allow-ssh",
			"allow-ssh",
		},
		{
			"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.KeyVault/vaults/myvault/secrets/mysecret",
			"mysecret",
		},
		{
			"/subscriptions/sub-456/resourceGroups/rg-storage/providers/Microsoft.Storage/storageAccounts/mystg",
			"mystg",
		},
	}

	for _, tt := range tests {
		got := parser.extractResourceName(tt.resourceID)
		assert.Equal(t, tt.want, got, "resourceID: %s", tt.resourceID)
	}
}

func TestActivityParser_ExtractSubscriptionID(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		want       string
	}{
		{
			"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
			"sub-123",
		},
		{
			"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg-test/providers/Microsoft.Network/networkSecurityGroups/nsg-test",
			"12345678-1234-1234-1234-123456789012",
		},
	}

	for _, tt := range tests {
		got := parser.extractSubscriptionID(tt.resourceID)
		assert.Equal(t, tt.want, got, "resourceID: %s", tt.resourceID)
	}
}

func TestActivityParser_ExtractResourceGroup(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		want       string
	}{
		{
			"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/my-vm",
			"rg-test",
		},
		{
			"/subscriptions/sub-123/resourceGroups/production/providers/Microsoft.Network/networkSecurityGroups/nsg-test",
			"production",
		},
	}

	for _, tt := range tests {
		got := parser.extractResourceGroup(tt.resourceID)
		assert.Equal(t, tt.want, got, "resourceID: %s", tt.resourceID)
	}
}

func TestActivityParser_ValidateEvent(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		name    string
		event   *testEvent
		wantErr bool
	}{
		{
			name: "valid event",
			event: &testEvent{
				Provider:     "azure",
				EventName:    "Microsoft.Compute/virtualMachines/write",
				ResourceType: "azurerm_virtual_machine",
				ResourceID:   "my-vm",
			},
			wantErr: false,
		},
		{
			name:    "nil event",
			event:   nil,
			wantErr: true,
		},
		{
			name: "invalid provider",
			event: &testEvent{
				Provider:     "aws",
				EventName:    "Microsoft.Compute/virtualMachines/write",
				ResourceType: "azurerm_virtual_machine",
				ResourceID:   "my-vm",
			},
			wantErr: true,
		},
		{
			name: "empty event name",
			event: &testEvent{
				Provider:     "azure",
				EventName:    "",
				ResourceType: "azurerm_virtual_machine",
				ResourceID:   "my-vm",
			},
			wantErr: true,
		},
		{
			name: "empty resource type",
			event: &testEvent{
				Provider:     "azure",
				EventName:    "Microsoft.Compute/virtualMachines/write",
				ResourceType: "",
				ResourceID:   "my-vm",
			},
			wantErr: true,
		},
		{
			name: "empty resource ID",
			event: &testEvent{
				Provider:     "azure",
				EventName:    "Microsoft.Compute/virtualMachines/write",
				ResourceType: "azurerm_virtual_machine",
				ResourceID:   "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var event *testEvent
			if tt.event != nil {
				event = tt.event
			}
			err := parser.ValidateEvent(convertTestEventToTypes(event))
			if tt.wantErr {
				assert.Error(t, err, "Expected validation error")
			} else {
				assert.NoError(t, err, "Unexpected validation error")
			}
		})
	}
}

// Test helpers
type testEvent struct {
	Provider     string
	EventName    string
	ResourceType string
	ResourceID   string
}

func convertTestEventToTypes(te *testEvent) *types.Event {
	if te == nil {
		return nil
	}
	return &types.Event{
		Provider:     te.Provider,
		EventName:    te.EventName,
		ResourceType: te.ResourceType,
		ResourceID:   te.ResourceID,
	}
}
