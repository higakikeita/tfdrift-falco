package azure

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuditParser(t *testing.T) {
	parser := NewAuditParser()
	require.NotNil(t, parser)
	assert.NotNil(t, parser.mapper)
	assert.NotNil(t, parser.relevantOps)
}

func TestAuditParser_Parse_ValidEvent(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "azureaudit",
		OutputFields: map[string]string{
			"azure.operationName":  "Microsoft.Compute/virtualMachines/write",
			"azure.resourceId":     "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"azure.resourceGroup":  "myRG",
			"azure.subscriptionId": "sub-123",
			"azure.caller":         "user@contoso.com",
			"azure.status":         "Succeeded",
		},
	}

	event := parser.Parse(res)

	assert.NotNil(t, event)
	assert.Equal(t, "azure", event.Provider)
	assert.Equal(t, "Microsoft.Compute/virtualMachines/write", event.EventName)
	assert.Equal(t, "azurerm_linux_virtual_machine", event.ResourceType)
	assert.Equal(t, "myVM", event.ResourceID)
	assert.Equal(t, "myRG", event.ResourceGroup)
	assert.Equal(t, "sub-123", event.SubscriptionID)
	assert.Equal(t, "user@contoso.com", event.UserIdentity.UserName)
	assert.NotEmpty(t, event.Changes)
	assert.Equal(t, "create", event.Changes["_action"])
}

func TestAuditParser_Parse_NilResponse(t *testing.T) {
	parser := NewAuditParser()
	event := parser.Parse(nil)
	assert.Nil(t, event)
}

func TestAuditParser_Parse_WrongSource(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name": "ModifyInstanceAttribute",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for non-Azure events")
}

func TestAuditParser_Parse_UnknownOperation(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "azureaudit",
		OutputFields: map[string]string{
			"azure.operationName":  "Microsoft.Unknown/unknownResource/read",
			"azure.resourceId":     "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Unknown/unknownResource/myResource",
			"azure.resourceGroup":  "myRG",
			"azure.subscriptionId": "sub-123",
			"azure.caller":         "user@contoso.com",
			"azure.status":         "Succeeded",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for unknown operations")
}

func TestAuditParser_Parse_FailedStatus(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "azureaudit",
		OutputFields: map[string]string{
			"azure.operationName":  "Microsoft.Compute/virtualMachines/write",
			"azure.resourceId":     "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"azure.resourceGroup":  "myRG",
			"azure.subscriptionId": "sub-123",
			"azure.caller":         "user@contoso.com",
			"azure.status":         "Failed",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for failed operations")
}

func TestAuditParser_Parse_DeleteOperation(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "azureaudit",
		OutputFields: map[string]string{
			"azure.operationName":  "Microsoft.Compute/virtualMachines/delete",
			"azure.resourceId":     "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"azure.resourceGroup":  "myRG",
			"azure.subscriptionId": "sub-123",
			"azure.caller":         "admin@contoso.com",
			"azure.status":         "Succeeded",
		},
	}

	event := parser.Parse(res)

	assert.NotNil(t, event)
	assert.Equal(t, "Microsoft.Compute/virtualMachines/delete", event.EventName)
	assert.Equal(t, "delete", event.Changes["_action"])
}

func TestAuditParser_isRelevantOperation(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name      string
		operation string
		want      bool
	}{
		{"VM Write", "Microsoft.Compute/virtualMachines/write", true},
		{"VM Delete", "Microsoft.Compute/virtualMachines/delete", true},
		{"Storage Account Write", "Microsoft.Storage/storageAccounts/write", true},
		{"NSG Write", "Microsoft.Network/networkSecurityGroups/write", true},
		{"Key Vault Write", "Microsoft.KeyVault/vaults/write", true},
		{"Unknown Operation", "Microsoft.Unknown/resource/write", false},
		{"Read Operation", "Microsoft.Compute/virtualMachines/read", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.isRelevantOperation(tt.operation)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractResourceName(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		expected   string
	}{
		{
			"Standard VM",
			"/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"myVM",
		},
		{
			"Storage Account",
			"/subscriptions/sub-456/resourceGroups/prod/providers/Microsoft.Storage/storageAccounts/mystorageaccount",
			"mystorageaccount",
		},
		{
			"Empty Resource ID",
			"",
			"",
		},
		{
			"Single Component",
			"myResource",
			"myResource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResourceName(tt.resourceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractResourceGroup(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		expected   string
	}{
		{
			"Standard Resource",
			"/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"myRG",
		},
		{
			"Another Resource Group",
			"/subscriptions/sub-456/resourceGroups/prod-rg/providers/Microsoft.Storage/storageAccounts/storage",
			"prod-rg",
		},
		{
			"No Resource Group",
			"/subscriptions/sub-123/providers/Microsoft.Subscription/subscriptionDefinitions/def",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractResourceGroup(tt.resourceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractSubscriptionID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		expected   string
	}{
		{
			"Standard Resource",
			"/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"sub-123",
		},
		{
			"Different Subscription",
			"/subscriptions/my-prod-sub-456/resourceGroups/prod/providers/Microsoft.Storage/storageAccounts/storage",
			"my-prod-sub-456",
		},
		{
			"No Subscription",
			"/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractSubscriptionID(tt.resourceID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
