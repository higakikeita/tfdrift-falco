package azure

import (
	"testing"
)

func TestCompareStateWithActual_UnmanagedResources(t *testing.T) {
	tfResources := []*TerraformResource{}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-1",
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"name":     "vm-1",
				"location": "eastus",
			},
		},
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/sa1",
			Type: "azurerm_storage_account",
			Name: "sa1",
			Attributes: map[string]interface{}{
				"name":     "sa1",
				"location": "eastus",
			},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.UnmanagedResources) != 2 {
		t.Errorf("expected 2 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified resources, got %d", len(result.ModifiedResources))
	}
}

func TestCompareStateWithActual_MissingResources(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"id":   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-1",
				"name": "vm-1",
			},
		},
		{
			Type: "azurerm_storage_account",
			Name: "sa1",
			Attributes: map[string]interface{}{
				"id":   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/sa1",
				"name": "sa1",
			},
		},
	}

	azureResources := []*DiscoveredResource{}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.MissingResources) != 2 {
		t.Errorf("expected 2 missing resources, got %d", len(result.MissingResources))
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged resources, got %d", len(result.UnmanagedResources))
	}
}

func TestCompareStateWithActual_ModifiedResources(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"id":       "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-1",
				"name":     "vm-1",
				"vm_size":  "Standard_D2_v3",
				"location": "eastus",
			},
		},
	}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-1",
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"name":     "vm-1",
				"vm_size":  "Standard_D4_v3",
				"location": "eastus",
			},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.ModifiedResources) != 1 {
		t.Fatalf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}

	diff := result.ModifiedResources[0]
	if diff.ResourceType != "azurerm_virtual_machine" {
		t.Errorf("expected resource type azurerm_virtual_machine, got %s", diff.ResourceType)
	}

	foundVMSize := false
	for _, d := range diff.Differences {
		if d.Field == "vm_size" {
			foundVMSize = true
			if d.TerraformValue != "Standard_D2_v3" {
				t.Errorf("expected terraform value Standard_D2_v3, got %v", d.TerraformValue)
			}
			if d.ActualValue != "Standard_D4_v3" {
				t.Errorf("expected actual value Standard_D4_v3, got %v", d.ActualValue)
			}
		}
	}
	if !foundVMSize {
		t.Error("expected vm_size difference not found")
	}
}

func TestCompareStateWithActual_MatchByName(t *testing.T) {
	// Test that resources can be matched by name when IDs differ
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_storage_account",
			Name: "mystorage",
			Attributes: map[string]interface{}{
				"name":     "mystorage",
				"location": "eastus",
			},
		},
	}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/mystorage",
			Type: "azurerm_storage_account",
			Name: "mystorage",
			Attributes: map[string]interface{}{
				"name":     "mystorage",
				"location": "eastus",
			},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged (matched by name), got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing (matched by name), got %d", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_MixedDrift(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_virtual_network",
			Name: "existing-vnet",
			Attributes: map[string]interface{}{
				"id":   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Network/virtualNetworks/existing-vnet",
				"name": "existing-vnet",
			},
		},
		{
			Type: "azurerm_storage_account",
			Name: "deleted-sa",
			Attributes: map[string]interface{}{
				"id":   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Storage/storageAccounts/deleted-sa",
				"name": "deleted-sa",
			},
		},
	}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Network/virtualNetworks/existing-vnet",
			Type: "azurerm_virtual_network",
			Name: "existing-vnet",
			Attributes: map[string]interface{}{
				"name": "existing-vnet",
			},
		},
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Network/virtualNetworks/manual-vnet",
			Type: "azurerm_virtual_network",
			Name: "manual-vnet",
			Attributes: map[string]interface{}{
				"name": "manual-vnet",
			},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.UnmanagedResources) != 1 {
		t.Errorf("expected 1 unmanaged resource (manual-vnet), got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 1 {
		t.Errorf("expected 1 missing resource (deleted-sa), got %d", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_EmptyInputs(t *testing.T) {
	result := CompareStateWithActual(nil, nil)

	if result == nil {
		t.Fatal("expected non-nil result for empty inputs")
	}
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified, got %d", len(result.ModifiedResources))
	}
}

func TestCompareStateWithActual_TagDrift(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_storage_account",
			Name: "tagged-sa",
			Attributes: map[string]interface{}{
				"id":   "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/tagged-sa",
				"name": "tagged-sa",
				"tags": map[string]interface{}{"env": "prod", "team": "platform"},
			},
		},
	}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/tagged-sa",
			Type: "azurerm_storage_account",
			Name: "tagged-sa",
			Attributes: map[string]interface{}{
				"name": "tagged-sa",
			},
			Tags: map[string]string{"env": "staging", "team": "platform", "hidden-link:/something": "true"},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.ModifiedResources) != 1 {
		t.Fatalf("expected 1 modified resource, got %d", len(result.ModifiedResources))
	}

	// Should detect tag drift (env changed, hidden-* ignored)
	foundTagDiff := false
	for _, d := range result.ModifiedResources[0].Differences {
		if d.Field == "tags" {
			foundTagDiff = true
		}
	}
	if !foundTagDiff {
		t.Error("expected tags difference not found")
	}
}

func TestGetComparableFields(t *testing.T) {
	tests := []struct {
		resourceType string
		minFields    int
	}{
		{"azurerm_virtual_machine", 3},
		{"azurerm_virtual_network", 2},
		{"azurerm_network_security_group", 1},
		{"azurerm_storage_account", 2},
		{"azurerm_kubernetes_cluster", 3},
		{"azurerm_mssql_server", 3},
		{"azurerm_key_vault", 2},
		{"azurerm_public_ip", 3},
		{"azurerm_redis_cache", 4},
		{"unknown_type", 1}, // defaults to ["location"]
	}

	for _, tt := range tests {
		fields := getComparableFields(tt.resourceType)
		if len(fields) < tt.minFields {
			t.Errorf("getComparableFields(%s): expected at least %d fields, got %d",
				tt.resourceType, tt.minFields, len(fields))
		}
	}
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a, b     interface{}
		expected bool
	}{
		{"nil-nil", nil, nil, true},
		{"nil-string", nil, "value", false},
		{"string-nil", "value", nil, false},
		{"same-string", "hello", "hello", true},
		{"diff-string", "hello", "world", false},
		{"bool-true", true, true, true},
		{"bool-false", false, false, true},
		{"bool-diff", true, false, false},
		{"bool-string-true", true, "true", true},
		{"bool-string-false", false, "false", true},
		{"int-same", 42, 42, true},
		{"int-diff", 42, 43, false},
		{"case-insensitive", "EastUS", "eastus", true},
	}

	for _, tt := range tests {
		result := valuesEqual(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("valuesEqual(%v, %v) [%s]: expected %v, got %v",
				tt.a, tt.b, tt.name, tt.expected, result)
		}
	}
}

func TestTagsEqual(t *testing.T) {
	tests := []struct {
		name      string
		tfAttrs   map[string]interface{}
		azureTags map[string]string
		expected  bool
	}{
		{
			name:      "both-empty",
			tfAttrs:   map[string]interface{}{},
			azureTags: map[string]string{},
			expected:  true,
		},
		{
			name: "equal-tags",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"env": "prod"},
			},
			azureTags: map[string]string{"env": "prod"},
			expected:  true,
		},
		{
			name: "azure-managed-tags-ignored",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"env": "prod"},
			},
			azureTags: map[string]string{"env": "prod", "hidden-link:/something": "true", "ms-resource-usage": "azure-cloud-shell"},
			expected:  true,
		},
		{
			name: "different-tags",
			tfAttrs: map[string]interface{}{
				"tags": map[string]interface{}{"env": "prod"},
			},
			azureTags: map[string]string{"env": "staging"},
			expected:  false,
		},
	}

	for _, tt := range tests {
		result := tagsEqual(tt.tfAttrs, tt.azureTags)
		if result != tt.expected {
			t.Errorf("tagsEqual [%s]: expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

func TestExtractTFResourceID(t *testing.T) {
	tests := []struct {
		name     string
		resource *TerraformResource
		expected string
	}{
		{
			name: "id-field",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{"id": "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-1"},
			},
			expected: "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-1",
		},
		{
			name: "name-field-fallback",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{"name": "my-resource"},
			},
			expected: "my-resource",
		},
		{
			name: "empty-attributes",
			resource: &TerraformResource{
				Attributes: map[string]interface{}{},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		result := extractTFResourceID(tt.resource)
		if result != tt.expected {
			t.Errorf("extractTFResourceID [%s]: expected %q, got %q", tt.name, tt.expected, result)
		}
	}
}

func TestExtractResourceGroupFromID(t *testing.T) {
	tests := []struct {
		resourceID string
		expected   string
	}{
		{"/subscriptions/sub-123/resourceGroups/rg-test/providers/Microsoft.Compute/virtualMachines/vm-1", "rg-test"},
		{"/subscriptions/sub-123/resourceGroups/MyRG/providers/Microsoft.Storage/storageAccounts/sa1", "MyRG"},
		{"/subscriptions/sub-123", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := extractResourceGroupFromID(tt.resourceID)
		if result != tt.expected {
			t.Errorf("extractResourceGroupFromID(%s): expected %q, got %q", tt.resourceID, tt.expected, result)
		}
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"eastus", "westus2", "centralus"}

	if !containsString(slice, "eastus") {
		t.Error("expected containsString to find eastus")
	}
	if !containsString(slice, "EastUS") {
		t.Error("expected containsString to find EastUS (case-insensitive)")
	}
	if containsString(slice, "northeurope") {
		t.Error("expected containsString not to find northeurope")
	}
	if containsString(nil, "anything") {
		t.Error("expected containsString to return false for nil slice")
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected interface{}
	}{
		{
			name:     "top-level field",
			data:     map[string]interface{}{"location": "eastus"},
			path:     "location",
			expected: "eastus",
		},
		{
			name: "nested field",
			data: map[string]interface{}{
				"network": map[string]interface{}{
					"cidr": "10.0.0.0/16",
				},
			},
			path:     "network.cidr",
			expected: "10.0.0.0/16",
		},
		{
			name:     "missing field",
			data:     map[string]interface{}{"location": "eastus"},
			path:     "missing",
			expected: nil,
		},
		{
			name:     "path through non-map",
			data:     map[string]interface{}{"value": "string"},
			path:     "value.nested",
			expected: nil,
		},
	}

	for _, tt := range tests {
		result := getNestedValue(tt.data, tt.path)
		if result != tt.expected {
			t.Errorf("getNestedValue [%s]: expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

func TestValuesEqual_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		a, b     interface{}
		expected bool
	}{
		{"both nil", nil, nil, true},
		{"nil vs empty string", nil, "", false},
		{"empty strings", "", "", true},
		{"case-insensitive", "westus", "WestUS", true},
		{"int and int", 100, 100, true},
		{"int and float", 100, 100.0, true}, // String comparison makes them equal
		{"bool true", true, true, true},
		{"bool string true", true, "true", true},
		{"bool string false", false, "false", true},
		{"float comparison", 1.5, 1.5, true},
	}

	for _, tt := range tests {
		result := valuesEqual(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("valuesEqual [%s]: expected %v, got %v", tt.name, tt.expected, result)
		}
	}
}

func TestGetTerraformTags(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]interface{}
		expected map[string]string
	}{
		{
			name: "valid tags",
			attrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"env":  "prod",
					"team": "platform",
				},
			},
			expected: map[string]string{
				"env":  "prod",
				"team": "platform",
			},
		},
		{
			name:     "no tags field",
			attrs:    map[string]interface{}{},
			expected: map[string]string{},
		},
		{
			name: "tags with non-string values",
			attrs: map[string]interface{}{
				"tags": map[string]interface{}{
					"env":  "prod",
					"port": 8080, // This should be skipped
				},
			},
			expected: map[string]string{
				"env": "prod",
			},
		},
		{
			name: "tags not a map",
			attrs: map[string]interface{}{
				"tags": "not-a-map",
			},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		result := getTerraformTags(tt.attrs)
		if len(result) != len(tt.expected) {
			t.Errorf("getTerraformTags [%s]: expected %d tags, got %d", tt.name, len(tt.expected), len(result))
		}
		for k, v := range tt.expected {
			if result[k] != v {
				t.Errorf("getTerraformTags [%s]: tag %s expected %s, got %s", tt.name, k, v, result[k])
			}
		}
	}
}

func TestMatchesByName_CaseSensitivity(t *testing.T) {
	tfMap := map[string]*TerraformResource{
		"1": {
			Type: "azurerm_virtual_machine",
			Attributes: map[string]interface{}{
				"name": "MyVM",
			},
		},
	}

	azRes := &DiscoveredResource{
		Type: "azurerm_virtual_machine",
		Name: "myvm",
	}

	if !matchesByName(azRes, tfMap) {
		t.Errorf("expected case-insensitive match for names")
	}
}

func TestFindByName_EmptyTerraformName(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "azurerm_virtual_machine",
		Attributes: map[string]interface{}{
			"name": "",
		},
	}

	azureMap := map[string]*DiscoveredResource{
		"1": {
			Type: "azurerm_virtual_machine",
			Name: "vm1",
		},
	}

	result := findByName(tfRes, azureMap)
	if result != nil {
		t.Errorf("expected nil for empty terraform name")
	}
}

func TestCompareResourceAttributes_TypeMismatch(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "azurerm_virtual_machine",
	}

	azRes := &DiscoveredResource{
		Type: "azurerm_storage_account",
	}

	result := compareResourceAttributes(tfRes, azRes)
	if result != nil {
		t.Errorf("expected nil for type mismatch")
	}
}

func TestCompareResourceAttributes_WithTags(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "azurerm_storage_account",
		Attributes: map[string]interface{}{
			"name": "sa1",
			"tags": map[string]interface{}{
				"env": "prod",
			},
		},
	}

	azRes := &DiscoveredResource{
		ID:   "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Storage/storageAccounts/sa1",
		Type: "azurerm_storage_account",
		Name: "sa1",
		Attributes: map[string]interface{}{
			"name": "sa1",
		},
		Tags: map[string]string{
			"env": "staging",
		},
	}

	diffs := compareResourceAttributes(tfRes, azRes)
	if diffs == nil {
		t.Fatalf("expected non-nil diffs")
	}

	// Should find tag diff
	foundTagDiff := false
	for _, d := range diffs {
		if d.Field == "tags" {
			foundTagDiff = true
		}
	}

	if !foundTagDiff {
		t.Errorf("expected tag difference to be detected")
	}
}

func TestCompareStateWithActual_AllResourcesMatch(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"id":       "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-1",
				"name":     "vm-1",
				"location": "eastus",
				"vm_size":  "Standard_B2s",
			},
		},
	}

	azureResources := []*DiscoveredResource{
		{
			ID:   "/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm-1",
			Type: "azurerm_virtual_machine",
			Name: "vm-1",
			Attributes: map[string]interface{}{
				"name":     "vm-1",
				"location": "eastus",
				"vm_size":  "Standard_B2s",
			},
		},
	}

	result := CompareStateWithActual(tfResources, azureResources)

	if len(result.UnmanagedResources) != 0 {
		t.Errorf("expected 0 unmanaged, got %d", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("expected 0 missing, got %d", len(result.MissingResources))
	}
	if len(result.ModifiedResources) != 0 {
		t.Errorf("expected 0 modified, got %d", len(result.ModifiedResources))
	}
}

func TestGetComparableFields_DefaultFallback(t *testing.T) {
	fields := getComparableFields("unknown.resource.type")
	if len(fields) == 0 {
		t.Errorf("expected at least one default field")
	}
	if fields[0] != "location" {
		t.Errorf("expected 'location' as default field")
	}
}

func TestExtractSubscriptionFromID(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"/subscriptions/sub-123/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", "sub-123"},
		{"/subscriptions/12345678-1234-1234-1234-123456789012/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", "12345678-1234-1234-1234-123456789012"},
		{"/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm", ""},
		{"", ""},
		{"/subscriptions/", ""}, // Edge case: subscriptions without ID
	}

	for _, tt := range tests {
		result := extractSubscriptionFromID(tt.id)
		if result != tt.expected {
			t.Errorf("extractSubscriptionFromID(%s): expected %q, got %q", tt.id, tt.expected, result)
		}
	}
}
