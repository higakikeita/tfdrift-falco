package provider

import (
	"github.com/keitahigaki/tfdrift-falco/pkg/azure"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Compile-time interface check
var _ Provider = (*AzureProvider)(nil)

// AzureProvider implements the Provider interface for Microsoft Azure.
type AzureProvider struct {
	parser *azure.AuditParser
	mapper *azure.ResourceMapper
}

// AzureProviderOption configures the Azure provider.
type AzureProviderOption func(*AzureProvider)

// NewAzureProvider creates a new Azure provider instance.
func NewAzureProvider(opts ...AzureProviderOption) *AzureProvider {
	p := &AzureProvider{
		parser: azure.NewAuditParser(),
		mapper: azure.NewResourceMapper(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *AzureProvider) Name() string { return "azure" }

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

func (p *AzureProvider) IsRelevantEvent(eventName string) bool {
	return p.mapper.MapOperationToResource(eventName) != ""
}

func (p *AzureProvider) MapEventToResource(eventName string, eventSource string) string {
	return p.mapper.MapOperationToResource(eventName)
}

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

func (p *AzureProvider) SupportedEventCount() int {
	return len(p.mapper.GetAllSupportedOperations())
}

func (p *AzureProvider) SupportedResourceTypes() []string {
	typeSet := make(map[string]bool)
	for _, event := range p.mapper.GetAllSupportedOperations() {
		if rt := p.mapper.MapOperationToResource(event); rt != "" {
			typeSet[rt] = true
		}
	}
	result := make([]string, 0, len(typeSet))
	for rt := range typeSet {
		result = append(result, rt)
	}
	return result
}
