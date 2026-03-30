package provider

import (
	"context"
	"fmt"
	"strings"

	azurepkg "github.com/keitahigaki/tfdrift-falco/pkg/azure"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Compile-time interface checks
var (
	_ Provider           = (*AzureProvider)(nil)
	_ ResourceDiscoverer = (*AzureProvider)(nil)
	_ StateComparator    = (*AzureProvider)(nil)
)

// AzureProvider implements Provider, ResourceDiscoverer, and StateComparator
// for Microsoft Azure. It wraps existing Azure Activity Log parsing,
// resource discovery, and state comparison logic.
type AzureProvider struct {
	parser         *azurepkg.ActivityParser
	mapper         *azurepkg.ResourceMapper
	subscriptionID string   // Azure subscription ID for resource discovery
	regions        []string // configured Azure regions/locations
	resourceGroup  string   // optional: limit discovery to a resource group
	resourceLister azurepkg.ResourceLister // Azure SDK resource lister
}

// AzureProviderOption configures the Azure provider.
type AzureProviderOption func(*AzureProvider)

// WithAzureSubscriptionID sets the subscription ID for resource discovery.
func WithAzureSubscriptionID(subscriptionID string) AzureProviderOption {
	return func(p *AzureProvider) {
		p.subscriptionID = subscriptionID
	}
}

// WithAzureRegions sets the regions for resource discovery.
func WithAzureRegions(regions []string) AzureProviderOption {
	return func(p *AzureProvider) {
		p.regions = regions
	}
}

// WithAzureResourceGroup limits discovery to a specific resource group.
func WithAzureResourceGroup(resourceGroup string) AzureProviderOption {
	return func(p *AzureProvider) {
		p.resourceGroup = resourceGroup
	}
}

// WithAzureResourceLister sets the resource lister for Azure SDK integration.
func WithAzureResourceLister(lister azurepkg.ResourceLister) AzureProviderOption {
	return func(p *AzureProvider) {
		p.resourceLister = lister
	}
}

// NewAzureProvider creates a new Azure provider instance.
func NewAzureProvider(opts ...AzureProviderOption) *AzureProvider {
	p := &AzureProvider{
		parser:  azurepkg.NewActivityParser(),
		mapper:  azurepkg.NewResourceMapper(),
		regions: []string{}, // empty = all regions
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Name returns the provider name.
func (p *AzureProvider) Name() string { return "azure" }

// ParseEvent parses an Azure event into a normalized event.
func (p *AzureProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	if source != "azure_activity" {
		return nil
	}

	// Handle pre-parsed events (from Falco subscriber)
	if azureEvent, ok := rawEvent.(*types.Event); ok {
		// Ensure metadata is populated
		if azureEvent.Metadata == nil {
			azureEvent.Metadata = make(map[string]string)
		}
		return azureEvent
	}

	// Field-based parsing for direct API calls
	operationName := fields["azure.operationName"]
	if operationName == "" {
		return nil
	}

	if !p.IsRelevantEvent(operationName) {
		return nil
	}

	resourceType := p.MapEventToResource(operationName, "")

	event := &types.Event{
		Provider:     "azure",
		EventName:    operationName,
		ResourceType: resourceType,
		ResourceID:   fields["azure.resourceId"],
		UserIdentity: types.UserIdentity{
			UserName: fields["azure.caller"],
			Type:     "AzureAD",
		},
		Changes:  p.ExtractChanges(operationName, fields),
		RawEvent: rawEvent,
		Metadata: make(map[string]string),
	}

	// Populate Azure metadata
	if subscriptionID := fields["azure.subscriptionId"]; subscriptionID != "" {
		event.Metadata["subscription_id"] = subscriptionID
	}
	if resourceGroup := fields["azure.resourceGroup"]; resourceGroup != "" {
		event.Metadata["resource_group"] = resourceGroup
	}
	if region := fields["azure.region"]; region != "" {
		event.Metadata["region"] = region
	}
	if correlationID := fields["azure.correlationId"]; correlationID != "" {
		event.Metadata["correlation_id"] = correlationID
	}

	return event
}

// IsRelevantEvent checks if an event is relevant for drift tracking.
func (p *AzureProvider) IsRelevantEvent(eventName string) bool {
	return p.mapper.MapEventToResource(eventName) != ""
}

// MapEventToResource maps an Azure event to a Terraform resource type.
func (p *AzureProvider) MapEventToResource(eventName string, eventSource string) string {
	return p.mapper.MapEventToResource(eventName)
}

// ExtractChanges extracts change details from event fields.
func (p *AzureProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})
	if fields["azure.requestProperties"] != "" {
		changes["request_properties"] = fields["azure.requestProperties"]
	}
	if fields["azure.responseProperties"] != "" {
		changes["response_properties"] = fields["azure.responseProperties"]
	}
	if fields["azure.status"] != "" {
		changes["status"] = fields["azure.status"]
	}
	return changes
}

// SupportedEventCount returns the number of supported event types.
func (p *AzureProvider) SupportedEventCount() int {
	return len(p.mapper.GetAllSupportedEvents())
}

// SupportedResourceTypes returns the list of supported Terraform resource types.
func (p *AzureProvider) SupportedResourceTypes() []string {
	typeSet := make(map[string]bool)
	for _, event := range p.mapper.GetAllSupportedEvents() {
		if rt := p.mapper.MapEventToResource(event); rt != "" {
			typeSet[rt] = true
		}
	}
	result := make([]string, 0, len(typeSet))
	for rt := range typeSet {
		result = append(result, rt)
	}
	return result
}

// --- ResourceDiscoverer implementation ---

// DiscoverResources enumerates actual Azure resources across configured regions.
func (p *AzureProvider) DiscoverResources(ctx context.Context, opts DiscoveryOptions) ([]*types.DiscoveredResource, error) {
	subscriptionID := p.subscriptionID
	if subscriptionID == "" {
		return nil, fmt.Errorf("Azure subscription ID is required for resource discovery; use WithAzureSubscriptionID option")
	}

	regions := opts.Regions
	if len(regions) == 0 {
		regions = p.regions
	}

	client, err := azurepkg.NewDiscoveryClient(subscriptionID, regions, p.resourceLister)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure discovery client: %w", err)
	}

	if p.resourceGroup != "" {
		client.WithResourceGroup(p.resourceGroup)
	}

	azureResources, err := client.DiscoverAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover Azure resources: %w", err)
	}

	// Convert Azure-specific DiscoveredResource to common type
	var allResources []*types.DiscoveredResource
	for _, r := range azureResources {
		metadata := map[string]string{
			"subscription_id": subscriptionID,
		}
		if rg := extractResourceGroupFromAzureID(r.ID); rg != "" {
			metadata["resource_group"] = rg
		}

		allResources = append(allResources, &types.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Provider:   "azure",
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Tags,
			Metadata:   metadata,
		})
	}

	return allResources, nil
}

// SupportedDiscoveryTypes returns the Terraform resource types that Azure can discover.
func (p *AzureProvider) SupportedDiscoveryTypes() []string {
	return azurepkg.SupportedDiscoveryTypes()
}

// --- StateComparator implementation ---

// CompareState compares Terraform resources with discovered Azure resources.
func (p *AzureProvider) CompareState(tfResources []*types.TerraformResource, actualResources []*types.DiscoveredResource, opts CompareOptions) *types.DriftResult {
	// Convert common types to Azure-specific types for the existing comparator
	azureTFResources := make([]*azurepkg.TerraformResource, 0, len(tfResources))
	for _, r := range tfResources {
		azureTFResources = append(azureTFResources, &azurepkg.TerraformResource{
			Type:       r.Type,
			Name:       r.Name,
			Attributes: r.Attributes,
		})
	}

	azureDiscovered := make([]*azurepkg.DiscoveredResource, 0, len(actualResources))
	for _, r := range actualResources {
		azureDiscovered = append(azureDiscovered, &azurepkg.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Tags,
		})
	}

	// Use existing Azure comparator
	azureResult := azurepkg.CompareStateWithActual(azureTFResources, azureDiscovered)

	// Convert Azure-specific result to common type
	result := &types.DriftResult{
		Provider:           "azure",
		UnmanagedResources: make([]*types.DiscoveredResource, 0, len(azureResult.UnmanagedResources)),
		MissingResources:   make([]*types.TerraformResource, 0, len(azureResult.MissingResources)),
		ModifiedResources:  make([]*types.ResourceDiff, 0, len(azureResult.ModifiedResources)),
	}

	for _, r := range azureResult.UnmanagedResources {
		result.UnmanagedResources = append(result.UnmanagedResources, &types.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Provider:   "azure",
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Tags,
		})
	}

	for _, r := range azureResult.MissingResources {
		result.MissingResources = append(result.MissingResources, &types.TerraformResource{
			Type:       r.Type,
			Name:       r.Name,
			Provider:   "azure",
			Attributes: r.Attributes,
		})
	}

	for _, d := range azureResult.ModifiedResources {
		diffs := make([]types.FieldDiff, 0, len(d.Differences))
		for _, fd := range d.Differences {
			diffs = append(diffs, types.FieldDiff{
				Field:          fd.Field,
				TerraformValue: fd.TerraformValue,
				ActualValue:    fd.ActualValue,
			})
		}
		result.ModifiedResources = append(result.ModifiedResources, &types.ResourceDiff{
			ResourceID:     d.ResourceID,
			ResourceType:   d.ResourceType,
			Provider:       "azure",
			TerraformState: d.TerraformState,
			ActualState:    d.ActualState,
			Differences:    diffs,
		})
	}

	return result
}

// extractResourceGroupFromAzureID extracts the resource group from an Azure resource ID.
func extractResourceGroupFromAzureID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if strings.EqualFold(part, "resourceGroups") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
