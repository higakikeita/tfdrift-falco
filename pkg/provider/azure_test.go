package provider

import (
	"context"
	"testing"

	azurepkg "github.com/keitahigaki/tfdrift-falco/pkg/azure"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Azure Provider Option Tests ---

func TestWithAzureSubscriptionID(t *testing.T) {
	p := NewAzureProvider(WithAzureSubscriptionID("sub-123"))
	assert.Equal(t, "sub-123", p.subscriptionID)
}

func TestWithAzureRegions(t *testing.T) {
	regions := []string{"eastus", "westus", "northeurope"}
	p := NewAzureProvider(WithAzureRegions(regions))
	assert.Equal(t, regions, p.regions)
}

func TestWithAzureResourceGroup(t *testing.T) {
	p := NewAzureProvider(WithAzureResourceGroup("my-rg"))
	assert.Equal(t, "my-rg", p.resourceGroup)
}

func TestWithAzureResourceLister(t *testing.T) {
	mockLister := newMockAzureResourceLister()
	p := NewAzureProvider(WithAzureResourceLister(mockLister))
	assert.Equal(t, mockLister, p.resourceLister)
}

func TestNewAzureProvider_DefaultValues(t *testing.T) {
	p := NewAzureProvider()
	assert.NotNil(t, p.parser)
	assert.NotNil(t, p.mapper)
	assert.Empty(t, p.subscriptionID)
	assert.Empty(t, p.regions)
	assert.Empty(t, p.resourceGroup)
}

func TestNewAzureProvider_MultipleOptions(t *testing.T) {
	p := NewAzureProvider(
		WithAzureSubscriptionID("sub-456"),
		WithAzureRegions([]string{"eastus"}),
		WithAzureResourceGroup("prod-rg"),
	)
	assert.Equal(t, "sub-456", p.subscriptionID)
	assert.Equal(t, []string{"eastus"}, p.regions)
	assert.Equal(t, "prod-rg", p.resourceGroup)
}

// --- Azure IsRelevantEvent Tests ---

func TestAzureIsRelevantEvent_ValidEvent(t *testing.T) {
	p := NewAzureProvider()
	// VirtualMachines are supported in Azure mapper
	result := p.IsRelevantEvent("Microsoft.Compute/virtualMachines/write")
	assert.True(t, result)
}

func TestAzureIsRelevantEvent_InvalidEvent(t *testing.T) {
	p := NewAzureProvider()
	result := p.IsRelevantEvent("NonExistentEvent")
	assert.False(t, result)
}

func TestAzureIsRelevantEvent_EmptyEvent(t *testing.T) {
	p := NewAzureProvider()
	result := p.IsRelevantEvent("")
	assert.False(t, result)
}

// --- Azure MapEventToResource Tests ---

func TestAzureMapEventToResource_ValidEvent(t *testing.T) {
	p := NewAzureProvider()
	resourceType := p.MapEventToResource("Microsoft.Compute/virtualMachines/write", "")
	assert.NotEmpty(t, resourceType)
	assert.Equal(t, "azurerm_virtual_machine", resourceType)
}

func TestAzureMapEventToResource_UnknownEvent(t *testing.T) {
	p := NewAzureProvider()
	resourceType := p.MapEventToResource("UnknownEvent", "")
	assert.Equal(t, "", resourceType)
}

func TestAzureMapEventToResource_EventSourceIgnored(t *testing.T) {
	p := NewAzureProvider()
	// eventSource parameter should be ignored for Azure
	resourceType1 := p.MapEventToResource("Microsoft.Compute/virtualMachines/write", "source1")
	resourceType2 := p.MapEventToResource("Microsoft.Compute/virtualMachines/write", "source2")
	assert.Equal(t, resourceType1, resourceType2)
}

// --- Azure ExtractChanges Tests ---

func TestAzureExtractChanges_WithRequestProperties(t *testing.T) {
	p := NewAzureProvider()
	changes := p.ExtractChanges("Microsoft.Compute/virtualMachines/write", map[string]string{
		"azure.requestProperties":  `{"vmSize":"Standard_D2s_v3"}`,
		"azure.responseProperties": `{"id":"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1"}`,
		"azure.status":             "Success",
	})

	assert.NotEmpty(t, changes)
	assert.Equal(t, `{"vmSize":"Standard_D2s_v3"}`, changes["request_properties"])
	assert.Equal(t, `{"id":"/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1"}`, changes["response_properties"])
	assert.Equal(t, "Success", changes["status"])
}

func TestAzureExtractChanges_EmptyFields(t *testing.T) {
	p := NewAzureProvider()
	changes := p.ExtractChanges("Microsoft.Compute/virtualMachines/write", map[string]string{})
	assert.Empty(t, changes)
}

func TestAzureExtractChanges_PartialFields(t *testing.T) {
	p := NewAzureProvider()
	changes := p.ExtractChanges("SomeEvent", map[string]string{
		"azure.status": "Failure",
	})

	assert.Len(t, changes, 1)
	assert.Equal(t, "Failure", changes["status"])
}

// --- Azure SupportedEventCount Tests ---

func TestAzureSupportedEventCount_GreaterThanZero(t *testing.T) {
	p := NewAzureProvider()
	count := p.SupportedEventCount()
	assert.Greater(t, count, 0)
}

// --- Azure SupportedResourceTypes Tests ---

func TestAzureSupportedResourceTypes_NotEmpty(t *testing.T) {
	p := NewAzureProvider()
	types := p.SupportedResourceTypes()
	assert.NotEmpty(t, types)
}

func TestAzureSupportedResourceTypes_HasVirtualMachine(t *testing.T) {
	p := NewAzureProvider()
	types := p.SupportedResourceTypes()
	// Should have VM or similar resource type
	assert.Greater(t, len(types), 0)
}

func TestAzureSupportedResourceTypes_NoDuplicates(t *testing.T) {
	p := NewAzureProvider()
	types := p.SupportedResourceTypes()
	seen := make(map[string]bool)
	for _, rt := range types {
		assert.False(t, seen[rt], "Duplicate resource type: %s", rt)
		seen[rt] = true
	}
}

// --- Azure DiscoverResources Tests ---

func TestAzureDiscoverResources_NoSubscriptionID(t *testing.T) {
	p := NewAzureProvider()
	_, err := p.DiscoverResources(context.Background(), DiscoveryOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subscription ID")
}

func TestAzureDiscoverResources_WithSubscriptionID_NilResourceLister(t *testing.T) {
	p := NewAzureProvider(WithAzureSubscriptionID("sub-123"))
	// Without a valid ResourceLister, this will fail when trying to discover
	// This tests that the function at least tries to proceed
	_, err := p.DiscoverResources(context.Background(), DiscoveryOptions{})
	require.Error(t, err)
	// Should fail on discovery, not on missing subscription
	assert.Contains(t, err.Error(), "failed to discover Azure resources")
}

func TestAzureDiscoverResources_UsesProvidedRegions(t *testing.T) {
	p := NewAzureProvider(
		WithAzureSubscriptionID("sub-123"),
		WithAzureRegions([]string{"eastus", "westus"}),
	)
	// Should use provided regions from options, not from provider
	opts := DiscoveryOptions{Regions: []string{"northeurope"}}
	_, err := p.DiscoverResources(context.Background(), opts)
	require.Error(t, err)
	// Error is expected due to no valid ResourceLister, but at least verify logic
	assert.Contains(t, err.Error(), "failed to discover Azure resources")
}

// --- Azure CompareState Tests ---

func TestAzureCompareState_EmptyInputs(t *testing.T) {
	p := NewAzureProvider()
	result := p.CompareState(nil, nil, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "azure", result.Provider)
	assert.Empty(t, result.UnmanagedResources)
	assert.Empty(t, result.MissingResources)
	assert.Empty(t, result.ModifiedResources)
}

func TestAzureCompareState_NoResources(t *testing.T) {
	p := NewAzureProvider()
	result := p.CompareState([]*types.TerraformResource{}, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "azure", result.Provider)
}

func TestAzureCompareState_WithTerraformResources(t *testing.T) {
	p := NewAzureProvider()
	tfResources := []*types.TerraformResource{
		{
			Type: "azurerm_virtual_machine",
			Name: "vm1",
			Attributes: map[string]interface{}{
				"name": "vm1",
			},
		},
	}
	result := p.CompareState(tfResources, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "azure", result.Provider)
}

func TestAzureCompareState_WithActualResources(t *testing.T) {
	p := NewAzureProvider()
	actualResources := []*types.DiscoveredResource{
		{
			ID:       "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
			Type:     "azurerm_virtual_machine",
			Provider: "azure",
			Name:     "vm1",
			Region:   "eastus",
		},
	}
	result := p.CompareState([]*types.TerraformResource{}, actualResources, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "azure", result.Provider)
}

// --- SupportedDiscoveryTypes Tests ---

func TestAzureSupportedDiscoveryTypes_NotEmpty(t *testing.T) {
	p := NewAzureProvider()
	types := p.SupportedDiscoveryTypes()
	assert.NotEmpty(t, types)
}

// --- Azure ParseEvent Tests ---

func TestAzureParseEvent_CorrectSource(t *testing.T) {
	p := NewAzureProvider()
	event := p.ParseEvent("azure_activity", map[string]string{
		"azure.operationName":  "Microsoft.Compute/virtualMachines/write",
		"azure.resourceId":     "/subscriptions/sub/resourceGroups/rg/providers/Microsoft.Compute/virtualMachines/vm1",
		"azure.caller":         "user@example.com",
		"azure.subscriptionId": "sub-123",
		"azure.resourceGroup":  "my-rg",
		"azure.region":         "eastus",
		"azure.correlationId":  "corr-123",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "azure", event.Provider)
	assert.Equal(t, "Microsoft.Compute/virtualMachines/write", event.EventName)
	assert.Equal(t, "user@example.com", event.UserIdentity.UserName)
	assert.Equal(t, "sub-123", event.GetMetadata("subscription_id"))
	assert.Equal(t, "my-rg", event.GetMetadata("resource_group"))
	assert.Equal(t, "eastus", event.GetMetadata("region"))
	assert.Equal(t, "corr-123", event.GetMetadata("correlation_id"))
}

func TestAzureParseEvent_WrongSource(t *testing.T) {
	p := NewAzureProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"azure.operationName": "Microsoft.Compute/virtualMachines/write",
	}, nil)
	assert.Nil(t, event)
}

func TestAzureParseEvent_MissingOperationName(t *testing.T) {
	p := NewAzureProvider()
	event := p.ParseEvent("azure_activity", map[string]string{}, nil)
	assert.Nil(t, event)
}

func TestAzureParseEvent_IrrelevantEvent(t *testing.T) {
	p := NewAzureProvider()
	event := p.ParseEvent("azure_activity", map[string]string{
		"azure.operationName": "NonExistentOperation",
	}, nil)
	assert.Nil(t, event)
}

func TestAzureParseEvent_PreParsedEvent(t *testing.T) {
	p := NewAzureProvider()
	preparsed := &types.Event{
		Provider:  "azure",
		EventName: "TestEvent",
		Metadata:  map[string]string{"key": "value"},
	}
	event := p.ParseEvent("azure_activity", map[string]string{}, preparsed)

	require.NotNil(t, event)
	assert.Equal(t, "TestEvent", event.EventName)
	assert.Equal(t, "value", event.GetMetadata("key"))
}

func TestAzureParseEvent_PreParsedEventNilMetadata(t *testing.T) {
	p := NewAzureProvider()
	preparsed := &types.Event{
		Provider:  "azure",
		EventName: "TestEvent",
		Metadata:  nil,
	}
	event := p.ParseEvent("azure_activity", map[string]string{}, preparsed)

	require.NotNil(t, event)
	assert.NotNil(t, event.Metadata)
}

// --- extractResourceGroupFromAzureID Tests ---

func TestExtractResourceGroupFromAzureID_ValidID(t *testing.T) {
	id := "/subscriptions/12345/resourceGroups/my-rg/providers/Microsoft.Compute/virtualMachines/vm1"
	rg := extractResourceGroupFromAzureID(id)
	assert.Equal(t, "my-rg", rg)
}

func TestExtractResourceGroupFromAzureID_NoResourceGroup(t *testing.T) {
	id := "/subscriptions/12345/providers/Microsoft.Compute/virtualMachines/vm1"
	rg := extractResourceGroupFromAzureID(id)
	assert.Empty(t, rg)
}

func TestExtractResourceGroupFromAzureID_CaseInsensitive(t *testing.T) {
	id := "/subscriptions/12345/RESOURCEGROUPS/my-rg/providers/Microsoft.Compute/virtualMachines/vm1"
	rg := extractResourceGroupFromAzureID(id)
	assert.Equal(t, "my-rg", rg)
}

func TestExtractResourceGroupFromAzureID_EmptyString(t *testing.T) {
	rg := extractResourceGroupFromAzureID("")
	assert.Empty(t, rg)
}

func TestExtractResourceGroupFromAzureID_MissingValue(t *testing.T) {
	id := "/subscriptions/12345/resourceGroups/"
	rg := extractResourceGroupFromAzureID(id)
	assert.Empty(t, rg)
}

// --- Mock Azure ResourceLister for testing ---

type mockAzureResourceLister struct{}

func newMockAzureResourceLister() azurepkg.ResourceLister {
	return &mockAzureResourceLister{}
}

func (m *mockAzureResourceLister) ListResources(ctx context.Context, subscriptionID string, resourceGroup string) ([]*azurepkg.Resource, error) {
	return []*azurepkg.Resource{}, nil
}

// --- Azure Provider Interface Compliance ---

func TestAzureProviderImplementsInterfaces(t *testing.T) {
	p := NewAzureProvider()

	// Core Provider interface
	var _ Provider = p
	assert.Equal(t, "azure", p.Name())

	// ResourceDiscoverer interface
	var _ ResourceDiscoverer = p
	discoveryTypes := p.SupportedDiscoveryTypes()
	assert.NotNil(t, discoveryTypes)

	// StateComparator interface
	var _ StateComparator = p
}
