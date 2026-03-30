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

func TestActivityParser_ExtractChanges(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		name          string
		operationName string
		fields        map[string]string
		checkFunc     func(t *testing.T, changes map[string]interface{})
	}{
		{
			name:          "write operation",
			operationName: "Microsoft.Compute/virtualMachines/write",
			fields: map[string]string{
				"azure.requestProperties": `{"vmSize":"Standard_D2s_v3"}`,
				"azure.responseProperties": `{"id":"vm-123"}`,
			},
			checkFunc: func(t *testing.T, changes map[string]interface{}) {
				if changes["_action"] != "write" {
					t.Errorf("expected _action=write")
				}
			},
		},
		{
			name:          "delete operation",
			operationName: "Microsoft.Compute/virtualMachines/delete",
			fields:        map[string]string{},
			checkFunc: func(t *testing.T, changes map[string]interface{}) {
				if changes["_action"] != "delete" {
					t.Errorf("expected _action=delete")
				}
			},
		},
		{
			name:          "action operation",
			operationName: "Microsoft.Compute/virtualMachines/start/action",
			fields:        map[string]string{},
			checkFunc: func(t *testing.T, changes map[string]interface{}) {
				// The extractChanges function returns "start/action" before trimming
				if _, ok := changes["_action"]; !ok {
					t.Errorf("expected _action to be present")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := parser.extractChanges(tt.operationName, tt.fields)
			if tt.checkFunc != nil {
				tt.checkFunc(t, changes)
			}
		})
	}
}

func TestActivityParser_IsRelevantEvent(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		name     string
		event    string
		expected bool
	}{
		{"VM write", "Microsoft.Compute/virtualMachines/write", true},
		{"VM read", "Microsoft.Compute/virtualMachines/read", false},
		{"NSG write", "Microsoft.Network/networkSecurityGroups/write", true},
		{"Storage delete", "Microsoft.Storage/storageAccounts/delete", true},
		{"Unknown operation", "Microsoft.Unknown/operation", false},
	}

	for _, tt := range tests {
		result := parser.isRelevantEvent(tt.event)
		if result != tt.expected {
			t.Errorf("isRelevantEvent(%s): expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

func TestActivityParser_Parse_MissingFields(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		name   string
		fields map[string]string
	}{
		{
			name: "missing operation name",
			fields: map[string]string{
				"azure.resourceId": "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm",
			},
		},
		{
			name: "missing resource ID",
			fields: map[string]string{
				"azure.operationName": "Microsoft.Compute/virtualMachines/write",
			},
		},
		{
			name: "empty operation name",
			fields: map[string]string{
				"azure.operationName": "",
				"azure.resourceId":    "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &outputs.Response{
				Source:       "azure_activity",
				OutputFields: tt.fields,
			}

			event := parser.Parse(res)
			if event != nil {
				t.Errorf("expected nil event for %s", tt.name)
			}
		})
	}
}

func TestActivityParser_Parse_CompleteEvent(t *testing.T) {
	parser := NewActivityParser()

	res := &outputs.Response{
		Source: "azure_activity",
		OutputFields: map[string]string{
			"azure.operationName":    "Microsoft.Compute/virtualMachines/write",
			"azure.resourceId":       "/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/my-vm",
			"azure.caller":           "user@example.com",
			"azure.resourceLocation": "eastus",
			"azure.correlationId":    "abc-123-def",
			"azure.status":           "Succeeded",
		},
	}

	event := parser.Parse(res)

	if event == nil {
		t.Fatalf("expected non-nil event")
	}

	assert.Equal(t, "azure", event.Provider)
	assert.Equal(t, "Microsoft.Compute/virtualMachines/write", event.EventName)
	assert.Equal(t, "azurerm_virtual_machine", event.ResourceType)
	assert.Equal(t, "my-vm", event.ResourceID)
	assert.Equal(t, "eastus", event.Region)
	assert.Equal(t, "user@example.com", event.UserIdentity.UserName)
	assert.Equal(t, "sub-123", event.UserIdentity.AccountID)
}

func TestActivityParser_ValidateEvent_InvalidProvider(t *testing.T) {
	parser := NewActivityParser()

	event := &types.Event{
		Provider:     "gcp",
		EventName:    "operation",
		ResourceType: "resource",
		ResourceID:   "id",
	}

	err := parser.ValidateEvent(event)
	if err == nil {
		t.Errorf("expected error for invalid provider")
	}
}

func TestGetStringField(t *testing.T) {
	fields := map[string]string{
		"key1": "value1",
		"key2": "",
	}

	tests := []struct {
		key      string
		expected string
	}{
		{"key1", "value1"},
		{"key2", ""},
		{"missing", ""},
	}

	for _, tt := range tests {
		result := getStringField(fields, tt.key)
		if result != tt.expected {
			t.Errorf("getStringField(%s): expected %q, got %q", tt.key, tt.expected, result)
		}
	}
}

func TestActivityParser_ExtractResourceName_EdgeCases(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		expected   string
	}{
		{"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/my-vm", "my-vm"},
		{"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Network/networkSecurityGroups/nsg/securityRules/allow-http", "allow-http"},
		{"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.KeyVault/vaults/vault/secrets/secret", "secret"},
		{"", ""},
		{"/", ""},
		{"/single", "single"},
	}

	for _, tt := range tests {
		result := parser.extractResourceName(tt.resourceID)
		if result != tt.expected {
			t.Errorf("extractResourceName(%s): expected %q, got %q", tt.resourceID, tt.expected, result)
		}
	}
}

func TestActivityParser_ExtractSubscriptionID_EdgeCases(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		expected   string
	}{
		{"/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", "sub-123"},
		{"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", "12345678-1234-1234-1234-123456789012"},
		{"/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := parser.extractSubscriptionID(tt.resourceID)
		if result != tt.expected {
			t.Errorf("extractSubscriptionID(%s): expected %q, got %q", tt.resourceID, tt.expected, result)
		}
	}
}

func TestActivityParser_ExtractResourceGroup_EdgeCases(t *testing.T) {
	parser := NewActivityParser()

	tests := []struct {
		resourceID string
		expected   string
	}{
		{"/subscriptions/sub-123/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/vm", "my-rg"},
		{"/subscriptions/sub-123/resourcegroups/MyRG/providers/Microsoft.Compute/virtualMachines/vm", ""}, // resourcegroups (lowercase) won't match
		{"/subscriptions/sub-123/providers/Microsoft.Compute/virtualMachines/vm", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := parser.extractResourceGroup(tt.resourceID)
		if result != tt.expected {
			t.Errorf("extractResourceGroup(%s): expected %q, got %q", tt.resourceID, tt.expected, result)
		}
	}
}
