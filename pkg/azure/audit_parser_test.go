package azure

import (
	"testing"
)

func TestParseEvent_VMCreate(t *testing.T) {
	parser := NewAuditParser()

	event := map[string]interface{}{
		"operationName":  "Microsoft.Compute/virtualMachines/write",
		"resourceId":     "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
		"status":         "Succeeded",
		"caller":         "admin@example.com",
		"eventTimestamp": "2026-03-20T10:00:00Z",
	}

	result := parser.ParseEvent(event)
	if result == nil {
		t.Fatal("ParseEvent returned nil for valid VM create event")
	}
	if result.TerraformResourceType != "azurerm_linux_virtual_machine" {
		t.Errorf("Expected azurerm_linux_virtual_machine, got %s", result.TerraformResourceType)
	}
	if result.ResourceName != "myVM" {
		t.Errorf("Expected resource name myVM, got %s", result.ResourceName)
	}
	if result.ResourceGroup != "myrg" {
		t.Errorf("Expected resource group myrg, got %s", result.ResourceGroup)
	}
	if result.Caller != "admin@example.com" {
		t.Errorf("Expected caller admin@example.com, got %s", result.Caller)
	}
	if result.Provider != "azure" {
		t.Errorf("Expected provider azure, got %s", result.Provider)
	}
}

func TestParseEvent_NSGDelete(t *testing.T) {
	parser := NewAuditParser()

	event := map[string]interface{}{
		"operationName": "Microsoft.Network/networkSecurityGroups/delete",
		"resourceId":    "/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Network/networkSecurityGroups/myNSG",
		"status":        "Succeeded",
		"caller":        "ops@example.com",
	}

	result := parser.ParseEvent(event)
	if result == nil {
		t.Fatal("ParseEvent returned nil for NSG delete")
	}
	if result.ChangeType != "deleted" {
		t.Errorf("Expected change type deleted, got %s", result.ChangeType)
	}
}

func TestParseEvent_IrrelevantOperation(t *testing.T) {
	parser := NewAuditParser()

	event := map[string]interface{}{
		"operationName": "Microsoft.Unknown/something/write",
		"status":        "Succeeded",
	}

	result := parser.ParseEvent(event)
	if result != nil {
		t.Error("Expected nil for irrelevant operation")
	}
}

func TestParseEvent_FailedOperation(t *testing.T) {
	parser := NewAuditParser()

	event := map[string]interface{}{
		"operationName": "Microsoft.Compute/virtualMachines/write",
		"status":        "Failed",
	}

	result := parser.ParseEvent(event)
	if result != nil {
		t.Error("Expected nil for failed operation")
	}
}

func TestExtractResourceName(t *testing.T) {
	tests := []struct {
		resourceID string
		want       string
	}{
		{"/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM", "myVM"},
		{"/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Network/virtualNetworks/myVNet", "myVNet"},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractResourceName(tt.resourceID)
		if got != tt.want {
			t.Errorf("extractResourceName(%q) = %q, want %q", tt.resourceID, got, tt.want)
		}
	}
}

func TestExtractSubscriptionID(t *testing.T) {
	got := extractSubscriptionID("/subscriptions/sub-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM")
	if got != "sub-123" {
		t.Errorf("Expected sub-123, got %s", got)
	}
}
