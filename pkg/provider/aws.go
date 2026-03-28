package provider

import (
	"context"
	"fmt"

	"github.com/keitahigaki/tfdrift-falco/pkg/aws"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco/mappings"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Compile-time interface checks
var (
	_ Provider           = (*AWSProvider)(nil)
	_ ResourceDiscoverer = (*AWSProvider)(nil)
	_ StateComparator    = (*AWSProvider)(nil)
)

// AWSProvider implements Provider, ResourceDiscoverer, and StateComparator
// for Amazon Web Services. It wraps existing AWS CloudTrail event parsing,
// resource discovery, and state comparison logic.
type AWSProvider struct {
	relevantEvents map[string]bool
	regions        []string // configured AWS regions
}

// AWSProviderOption configures the AWS provider.
type AWSProviderOption func(*AWSProvider)

// WithAWSRegions sets the regions for resource discovery.
func WithAWSRegions(regions []string) AWSProviderOption {
	return func(p *AWSProvider) {
		p.regions = regions
	}
}

// NewAWSProvider creates a new AWS provider instance.
func NewAWSProvider(opts ...AWSProviderOption) *AWSProvider {
	p := &AWSProvider{
		relevantEvents: falco.GetAWSRelevantEvents(),
		regions:        []string{"us-east-1"}, // default region
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *AWSProvider) Name() string { return "aws" }

func (p *AWSProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	if source != "aws_cloudtrail" {
		return nil
	}

	eventName := fields["ct.name"]
	if eventName == "" {
		return nil
	}

	if !p.IsRelevantEvent(eventName) {
		return nil
	}

	resourceID := falco.ExtractAWSResourceID(eventName, fields)
	if resourceID == "" {
		return nil
	}

	eventSource := fields["ct.src"]
	resourceType := p.MapEventToResource(eventName, eventSource)

	userIdentity := types.UserIdentity{
		Type:        fields["ct.user.type"],
		PrincipalID: fields["ct.user.principalid"],
		ARN:         fields["ct.user.arn"],
		AccountID:   fields["ct.user.accountid"],
		UserName:    fields["ct.user"],
	}

	changes := p.ExtractChanges(eventName, fields)

	event := &types.Event{
		Provider:     "aws",
		EventName:    eventName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     rawEvent,
		Metadata:     make(map[string]string),
	}

	// Populate metadata with AWS-specific fields
	if region := fields["ct.region"]; region != "" {
		event.Metadata["region"] = region
		event.Region = region // backward compatibility
	}
	if accountID := fields["ct.user.accountid"]; accountID != "" {
		event.Metadata["account_id"] = accountID
	}
	if eventSource := fields["ct.src"]; eventSource != "" {
		event.Metadata["event_source"] = eventSource
	}

	return event
}

func (p *AWSProvider) IsRelevantEvent(eventName string) bool {
	return p.relevantEvents[eventName]
}

func (p *AWSProvider) MapEventToResource(eventName string, eventSource string) string {
	// First try conflict resolution
	if resolved := mappings.ResolveEventSourceConflict(eventName, eventSource); resolved != "" {
		return resolved
	}

	allMappings := []map[string]string{
		mappings.ComputeMappings,
		mappings.NetworkingMappings,
		mappings.StorageAndDatabaseMappings,
		mappings.SecurityMappings,
		mappings.OtherServicesMappings,
	}

	for _, mapping := range allMappings {
		if resourceType, ok := mapping[eventName]; ok {
			return resourceType
		}
	}

	return "unknown"
}

func (p *AWSProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	return falco.ExtractAWSChanges(eventName, fields)
}

func (p *AWSProvider) SupportedEventCount() int {
	return len(p.relevantEvents)
}

func (p *AWSProvider) SupportedResourceTypes() []string {
	typeSet := make(map[string]bool)
	allMappings := []map[string]string{
		mappings.ComputeMappings,
		mappings.NetworkingMappings,
		mappings.StorageAndDatabaseMappings,
		mappings.SecurityMappings,
		mappings.OtherServicesMappings,
	}
	for _, m := range allMappings {
		for _, rt := range m {
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

// DiscoverResources enumerates actual AWS resources across configured regions.
func (p *AWSProvider) DiscoverResources(ctx context.Context, opts DiscoveryOptions) ([]*types.DiscoveredResource, error) {
	regions := opts.Regions
	if len(regions) == 0 {
		regions = p.regions
	}

	var allResources []*types.DiscoveredResource
	for _, region := range regions {
		client, err := aws.NewDiscoveryClient(ctx, region)
		if err != nil {
			return nil, fmt.Errorf("failed to create discovery client for region %s: %w", region, err)
		}

		awsResources, err := client.DiscoverAll(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to discover resources in region %s: %w", region, err)
		}

		// Convert AWS-specific DiscoveredResource to common type
		for _, r := range awsResources {
			allResources = append(allResources, &types.DiscoveredResource{
				ID:         r.ID,
				Type:       r.Type,
				Provider:   "aws",
				ARN:        r.ARN,
				Name:       r.Name,
				Region:     r.Region,
				Attributes: r.Attributes,
				Tags:       r.Tags,
			})
		}
	}

	return allResources, nil
}

// SupportedDiscoveryTypes returns the Terraform resource types that AWS can discover.
func (p *AWSProvider) SupportedDiscoveryTypes() []string {
	return []string{
		"aws_vpc",
		"aws_subnet",
		"aws_security_group",
		"aws_instance",
		"aws_db_instance",
		"aws_eks_cluster",
		"aws_elasticache_replication_group",
		"aws_lb",
	}
}

// --- StateComparator implementation ---

// CompareState compares Terraform resources with discovered AWS resources.
func (p *AWSProvider) CompareState(tfResources []*types.TerraformResource, actualResources []*types.DiscoveredResource, opts CompareOptions) *types.DriftResult {
	// Convert common types to AWS-specific types for the existing comparator
	awsTFResources := make([]*terraform.Resource, 0, len(tfResources))
	for _, r := range tfResources {
		awsTFResources = append(awsTFResources, &terraform.Resource{
			Type:       r.Type,
			Name:       r.Name,
			Attributes: r.Attributes,
		})
	}

	awsDiscovered := make([]*aws.DiscoveredResource, 0, len(actualResources))
	for _, r := range actualResources {
		awsDiscovered = append(awsDiscovered, &aws.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			ARN:        r.ARN,
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Tags,
		})
	}

	// Use existing comparator
	awsResult := aws.CompareStateWithActual(awsTFResources, awsDiscovered)

	// Convert AWS-specific result to common type
	result := &types.DriftResult{
		Provider:           "aws",
		UnmanagedResources: make([]*types.DiscoveredResource, 0, len(awsResult.UnmanagedResources)),
		MissingResources:   make([]*types.TerraformResource, 0, len(awsResult.MissingResources)),
		ModifiedResources:  make([]*types.ResourceDiff, 0, len(awsResult.ModifiedResources)),
	}

	for _, r := range awsResult.UnmanagedResources {
		result.UnmanagedResources = append(result.UnmanagedResources, &types.DiscoveredResource{
			ID:         r.ID,
			Type:       r.Type,
			Provider:   "aws",
			ARN:        r.ARN,
			Name:       r.Name,
			Region:     r.Region,
			Attributes: r.Attributes,
			Tags:       r.Tags,
		})
	}

	for _, r := range awsResult.MissingResources {
		result.MissingResources = append(result.MissingResources, &types.TerraformResource{
			Type:       r.Type,
			Name:       r.Name,
			Provider:   "aws",
			Attributes: r.Attributes,
		})
	}

	for _, d := range awsResult.ModifiedResources {
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
			Provider:       "aws",
			TerraformState: d.TerraformState,
			ActualState:    d.ActualState,
			Differences:    diffs,
		})
	}

	return result
}
