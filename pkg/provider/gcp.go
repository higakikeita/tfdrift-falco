package provider

import (
	"context"
	"fmt"

	gcppkg "github.com/keitahigaki/tfdrift-falco/pkg/gcp"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Compile-time interface checks
var (
	_ Provider           = (*GCPProvider)(nil)
	_ ResourceDiscoverer = (*GCPProvider)(nil)
	_ StateComparator    = (*GCPProvider)(nil)
)

// GCPProvider implements Provider, ResourceDiscoverer, and StateComparator
// for Google Cloud Platform. It wraps existing GCP Audit Log parsing,
// resource discovery, and state comparison logic.
type GCPProvider struct {
	parser    *gcppkg.AuditParser
	mapper    *gcppkg.ResourceMapper
	projectID string   // GCP project ID for resource discovery
	regions   []string // configured GCP regions
}

// GCPProviderOption configures the GCP provider.
type GCPProviderOption func(*GCPProvider)

// WithGCPProjectID sets the project ID for resource discovery.
func WithGCPProjectID(projectID string) GCPProviderOption {
	return func(p *GCPProvider) {
		p.projectID = projectID
	}
}

// WithGCPRegions sets the regions for resource discovery.
func WithGCPRegions(regions []string) GCPProviderOption {
	return func(p *GCPProvider) {
		p.regions = regions
	}
}

// NewGCPProvider creates a new GCP provider instance.
func NewGCPProvider(opts ...GCPProviderOption) *GCPProvider {
	p := &GCPProvider{
		parser:  gcppkg.NewAuditParser(),
		mapper:  gcppkg.NewResourceMapper(),
		regions: []string{}, // empty = all regions
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *GCPProvider) Name() string { return "gcp" }

func (p *GCPProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	if source != "gcpaudit" {
		return nil
	}

	// Handle pre-parsed events (from Falco subscriber)
	if gcpEvent, ok := rawEvent.(*types.Event); ok {
		// Ensure metadata is populated for backward compatibility
		if gcpEvent.Metadata == nil {
			gcpEvent.Metadata = make(map[string]string)
		}
		if gcpEvent.ProjectID != "" {
			gcpEvent.Metadata["project_id"] = gcpEvent.ProjectID
		}
		if gcpEvent.ServiceName != "" {
			gcpEvent.Metadata["service_name"] = gcpEvent.ServiceName
		}
		return gcpEvent
	}

	// Field-based parsing for direct API calls
	methodName := fields["gcp.methodName"]
	if methodName == "" {
		return nil
	}

	if !p.IsRelevantEvent(methodName) {
		return nil
	}

	resourceType := p.MapEventToResource(methodName, "")

	event := &types.Event{
		Provider:     "gcp",
		EventName:    methodName,
		ResourceType: resourceType,
		ResourceID:   fields["gcp.resource.name"],
		UserIdentity: types.UserIdentity{
			UserName: fields["gcp.user"],
			Type:     "ServiceAccount",
		},
		Changes:  p.ExtractChanges(methodName, fields),
		RawEvent: rawEvent,
		Metadata: make(map[string]string),
	}

	// Populate GCP metadata
	if projectID := fields["gcp.projectId"]; projectID != "" {
		event.Metadata["project_id"] = projectID
		event.ProjectID = projectID // backward compatibility
	}
	if serviceName := fields["gcp.serviceName"]; serviceName != "" {
		event.Metadata["service_name"] = serviceName
		event.ServiceName = serviceName // backward compatibility
	}
	if zone := fields["gcp.zone"]; zone != "" {
		event.Metadata["zone"] = zone
	}
	if region := fields["gcp.region"]; region != "" {
		event.Metadata["region"] = region
	}

	return event
}

func (p *GCPProvider) IsRelevantEvent(eventName string) bool {
	// Delegate to the mapper: if the event maps to a resource, it's relevant
	return p.mapper.MapEventToResource(eventName) != ""
}

func (p *GCPProvider) MapEventToResource(eventName string, eventSource string) string {
	return p.mapper.MapEventToResource(eventName)
}

func (p *GCPProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	// Extract available change information from fields
	changes := make(map[string]interface{})
	if fields["gcp.request"] != "" {
		changes["request"] = fields["gcp.request"]
	}
	if fields["gcp.response"] != "" {
		changes["response"] = fields["gcp.response"]
	}
	return changes
}

func (p *GCPProvider) SupportedEventCount() int {
	return len(p.mapper.GetAllSupportedEvents())
}

func (p *GCPProvider) SupportedResourceTypes() []string {
	// Deduplicate resource types from the mapper
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

// DiscoverResources enumerates actual GCP resources across configured regions.
func (p *GCPProvider) DiscoverResources(ctx context.Context, opts DiscoveryOptions) ([]*types.DiscoveredResource, error) {
	projectID := p.projectID
	if projectID == "" {
		return nil, fmt.Errorf("GCP project ID is required for resource discovery; use WithGCPProjectID option")
	}

	regions := opts.Regions
	if len(regions) == 0 {
		regions = p.regions
	}

	client, err := gcppkg.NewDiscoveryClient(ctx, projectID, regions)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCP discovery client: %w", err)
	}

	gcpResources, err := client.DiscoverAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to discover GCP resources: %w", err)
	}

	// Convert GCP-specific DiscoveredResource to common type
	var allResources []*types.DiscoveredResource
	for _, r := range gcpResources {
		metadata := map[string]string{
			"project_id": projectID,
		}
		if r.SelfLink != "" {
			metadata["self_link"] = r.SelfLink
		}

		allResources = append(allResources, &types.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Provider:   "gcp",
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Labels,
			Metadata:   metadata,
		})
	}

	return allResources, nil
}

// SupportedDiscoveryTypes returns the Terraform resource types that GCP can discover.
func (p *GCPProvider) SupportedDiscoveryTypes() []string {
	return []string{
		"google_compute_network",
		"google_compute_subnetwork",
		"google_compute_firewall",
		"google_compute_instance",
		"google_storage_bucket",
		"google_sql_database_instance",
		"google_container_cluster",
		"google_cloud_run_v2_service",
	}
}

// --- StateComparator implementation ---

// CompareState compares Terraform resources with discovered GCP resources.
func (p *GCPProvider) CompareState(tfResources []*types.TerraformResource, actualResources []*types.DiscoveredResource, opts CompareOptions) *types.DriftResult {
	// Convert common types to GCP-specific types for the existing comparator
	gcpTFResources := make([]*gcppkg.TerraformResource, 0, len(tfResources))
	for _, r := range tfResources {
		gcpTFResources = append(gcpTFResources, &gcppkg.TerraformResource{
			Type:       r.Type,
			Name:       r.Name,
			Attributes: r.Attributes,
		})
	}

	gcpDiscovered := make([]*gcppkg.DiscoveredResource, 0, len(actualResources))
	for _, r := range actualResources {
		gcpDiscovered = append(gcpDiscovered, &gcppkg.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Labels:     r.Tags,
		})
	}

	// Use existing GCP comparator
	gcpResult := gcppkg.CompareStateWithActual(gcpTFResources, gcpDiscovered)

	// Convert GCP-specific result to common type
	result := &types.DriftResult{
		Provider:           "gcp",
		UnmanagedResources: make([]*types.DiscoveredResource, 0, len(gcpResult.UnmanagedResources)),
		MissingResources:   make([]*types.TerraformResource, 0, len(gcpResult.MissingResources)),
		ModifiedResources:  make([]*types.ResourceDiff, 0, len(gcpResult.ModifiedResources)),
	}

	for _, r := range gcpResult.UnmanagedResources {
		result.UnmanagedResources = append(result.UnmanagedResources, &types.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Provider:   "gcp",
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Labels,
		})
	}

	for _, r := range gcpResult.MissingResources {
		result.MissingResources = append(result.MissingResources, &types.TerraformResource{
			Type:       r.Type,
			Name:       r.Name,
			Provider:   "gcp",
			Attributes: r.Attributes,
		})
	}

	for _, d := range gcpResult.ModifiedResources {
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
			Provider:       "gcp",
			TerraformState: d.TerraformState,
			ActualState:    d.ActualState,
			Differences:    diffs,
		})
	}

	return result
}
