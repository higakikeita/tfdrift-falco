// Package provider defines the common interface for cloud providers in TFDrift-Falco.
//
// The provider system uses a layered interface design:
//   - Provider: core interface for event parsing and resource mapping (required)
//   - ResourceDiscoverer: optional interface for discovering actual cloud resources
//   - StateComparator: optional interface for comparing Terraform state with actual state
//   - FullProvider: composite interface combining all capabilities
//
// Providers implement the interfaces they support. Use type assertions to check
// for optional capabilities:
//
//	if discoverer, ok := p.(provider.ResourceDiscoverer); ok {
//	    resources, err := discoverer.DiscoverResources(ctx, opts)
//	}
package provider

import (
	"context"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// Provider defines the core interface that all cloud providers must implement.
// Each provider is responsible for parsing events, mapping resources, and managing state.
type Provider interface {
	// Name returns the provider identifier (e.g., "aws", "gcp", "azure")
	Name() string

	// ParseEvent parses a raw Falco event into a typed Event.
	// Returns nil if the event is not relevant or not from this provider.
	ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event

	// IsRelevantEvent checks whether the given event name should trigger drift detection.
	IsRelevantEvent(eventName string) bool

	// MapEventToResource maps a cloud event name to a Terraform resource type.
	MapEventToResource(eventName string, eventSource string) string

	// ExtractChanges extracts attribute changes from event fields.
	ExtractChanges(eventName string, fields map[string]string) map[string]interface{}

	// SupportedEventCount returns the number of event types this provider can detect.
	SupportedEventCount() int

	// SupportedResourceTypes returns a list of Terraform resource types this provider covers.
	SupportedResourceTypes() []string
}

// DiscoveryOptions configures resource discovery behavior.
type DiscoveryOptions struct {
	// Regions to discover resources in (empty = all configured regions)
	Regions []string

	// ResourceTypes to discover (empty = all supported types)
	ResourceTypes []string

	// Tags filter: only discover resources matching these tags
	Tags map[string]string
}

// ResourceDiscoverer is an optional interface for providers that can enumerate
// actual cloud resources. This enables drift detection by comparing discovered
// resources against Terraform state.
type ResourceDiscoverer interface {
	// DiscoverResources enumerates actual cloud resources.
	// Returns provider-agnostic DiscoveredResource structs.
	DiscoverResources(ctx context.Context, opts DiscoveryOptions) ([]*types.DiscoveredResource, error)

	// SupportedDiscoveryTypes returns the Terraform resource types that this
	// provider can discover. An empty slice means discovery is not yet implemented.
	SupportedDiscoveryTypes() []string
}

// CompareOptions configures state comparison behavior.
type CompareOptions struct {
	// IgnoredAttributes are attribute names to skip during comparison.
	IgnoredAttributes []string

	// IgnoredTagPrefixes are tag key prefixes to skip (e.g., "aws:", "kubernetes.io/").
	IgnoredTagPrefixes []string
}

// StateComparator is an optional interface for providers that can compare
// Terraform state with actual cloud state to produce drift results.
type StateComparator interface {
	// CompareState compares Terraform resources with discovered cloud resources
	// and returns a unified DriftResult.
	CompareState(tfResources []*types.TerraformResource, actualResources []*types.DiscoveredResource, opts CompareOptions) *types.DriftResult
}

// FullProvider combines all provider interfaces.
// Providers that implement all capabilities can satisfy this interface.
type FullProvider interface {
	Provider
	ResourceDiscoverer
	StateComparator
}

// Capabilities returns a summary of what optional interfaces a provider implements.
type Capabilities struct {
	Discovery  bool
	Comparison bool
}

// GetCapabilities checks which optional interfaces a provider implements.
func GetCapabilities(p Provider) Capabilities {
	_, hasDiscovery := p.(ResourceDiscoverer)
	_, hasComparison := p.(StateComparator)
	return Capabilities{
		Discovery:  hasDiscovery,
		Comparison: hasComparison,
	}
}

// StateConfig holds the Terraform state configuration for a provider.
type StateConfig struct {
	Backend   string
	LocalPath string
	S3Bucket  string
	S3Key     string
	S3Region  string
	GCSBucket string
	GCSPrefix string
	// Azure Blob Storage
	AzureStorageAccount string
	AzureContainerName  string
	AzureBlobName       string
}
