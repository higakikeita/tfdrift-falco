package provider

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mock providers for testing ---

// mockBasicProvider only implements Provider (no discovery/comparison)
type mockBasicProvider struct {
	name          string
	source        string
	eventCount    int
	resourceTypes []string
}

func (m *mockBasicProvider) Name() string { return m.name }
func (m *mockBasicProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	if source != m.source {
		return nil
	}
	return &types.Event{
		Provider:  m.name,
		EventName: fields["event"],
		Metadata:  make(map[string]string),
	}
}
func (m *mockBasicProvider) IsRelevantEvent(eventName string) bool                              { return true }
func (m *mockBasicProvider) MapEventToResource(eventName string, eventSource string) string      { return "mock_resource" }
func (m *mockBasicProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	return nil
}
func (m *mockBasicProvider) SupportedEventCount() int         { return m.eventCount }
func (m *mockBasicProvider) SupportedResourceTypes() []string { return m.resourceTypes }

// mockFullProvider implements Provider + ResourceDiscoverer + StateComparator
type mockFullProvider struct {
	mockBasicProvider
	discoveredResources []*types.DiscoveredResource
}

func (m *mockFullProvider) DiscoverResources(ctx context.Context, opts DiscoveryOptions) ([]*types.DiscoveredResource, error) {
	return m.discoveredResources, nil
}

func (m *mockFullProvider) SupportedDiscoveryTypes() []string {
	return []string{"mock_instance", "mock_bucket"}
}

func (m *mockFullProvider) CompareState(tfResources []*types.TerraformResource, actualResources []*types.DiscoveredResource, opts CompareOptions) *types.DriftResult {
	return &types.DriftResult{
		Provider:           m.name,
		UnmanagedResources: []*types.DiscoveredResource{},
		MissingResources:   []*types.TerraformResource{},
		ModifiedResources:  []*types.ResourceDiff{},
	}
}

// --- Interface tests ---

func TestGetCapabilities_BasicProvider(t *testing.T) {
	p := &mockBasicProvider{name: "basic"}
	caps := GetCapabilities(p)

	assert.False(t, caps.Discovery, "basic provider should not have discovery")
	assert.False(t, caps.Comparison, "basic provider should not have comparison")
}

func TestGetCapabilities_FullProvider(t *testing.T) {
	p := &mockFullProvider{
		mockBasicProvider: mockBasicProvider{name: "full"},
	}
	caps := GetCapabilities(p)

	assert.True(t, caps.Discovery, "full provider should have discovery")
	assert.True(t, caps.Comparison, "full provider should have comparison")
}

func TestRegistryGetDiscoverer(t *testing.T) {
	r := NewRegistry()

	basic := &mockBasicProvider{name: "basic", source: "basic_src", eventCount: 5}
	full := &mockFullProvider{
		mockBasicProvider: mockBasicProvider{name: "full", source: "full_src", eventCount: 10},
	}

	require.NoError(t, r.Register(basic))
	require.NoError(t, r.Register(full))

	// Basic provider should not be a discoverer
	_, ok := r.GetDiscoverer("basic")
	assert.False(t, ok)

	// Full provider should be a discoverer
	d, ok := r.GetDiscoverer("full")
	assert.True(t, ok)
	assert.Equal(t, []string{"mock_instance", "mock_bucket"}, d.SupportedDiscoveryTypes())
}

func TestRegistryGetComparator(t *testing.T) {
	r := NewRegistry()

	full := &mockFullProvider{
		mockBasicProvider: mockBasicProvider{name: "full", source: "full_src", eventCount: 10},
	}
	require.NoError(t, r.Register(full))

	c, ok := r.GetComparator("full")
	assert.True(t, ok)

	result := c.CompareState(nil, nil, CompareOptions{})
	assert.Equal(t, "full", result.Provider)
}

func TestRegistryGetAllCapabilities(t *testing.T) {
	r := NewRegistry()

	basic := &mockBasicProvider{name: "basic", source: "basic_src", eventCount: 5}
	full := &mockFullProvider{
		mockBasicProvider: mockBasicProvider{name: "full", source: "full_src", eventCount: 10},
	}

	require.NoError(t, r.Register(basic))
	require.NoError(t, r.Register(full))

	caps := r.GetAllCapabilities()
	assert.Len(t, caps, 2)

	assert.False(t, caps["basic"].Discovery)
	assert.False(t, caps["basic"].Comparison)
	assert.True(t, caps["full"].Discovery)
	assert.True(t, caps["full"].Comparison)
}

func TestRegistryDiscoverAll(t *testing.T) {
	r := NewRegistry()

	discovered := []*types.DiscoveredResource{
		{ID: "i-123", Type: "mock_instance", Provider: "full"},
	}
	full := &mockFullProvider{
		mockBasicProvider:   mockBasicProvider{name: "full", source: "full_src", eventCount: 10},
		discoveredResources: discovered,
	}
	basic := &mockBasicProvider{name: "basic", source: "basic_src", eventCount: 5}

	require.NoError(t, r.Register(full))
	require.NoError(t, r.Register(basic))

	results, err := r.DiscoverAll(context.Background(), DiscoveryOptions{})
	require.NoError(t, err)

	// Only the full provider should have results
	assert.Len(t, results, 1)
	assert.Contains(t, results, "full")
	assert.Len(t, results["full"], 1)
	assert.Equal(t, "i-123", results["full"][0].ID)
}

// --- Event Metadata tests ---

func TestEventMetadata(t *testing.T) {
	event := &types.Event{
		Provider:  "aws",
		EventName: "CreateBucket",
	}

	// GetMetadata on nil map should return empty string
	assert.Equal(t, "", event.GetMetadata("region"))

	// SetMetadata should initialize map
	event.SetMetadata("region", "us-east-1")
	assert.Equal(t, "us-east-1", event.GetMetadata("region"))

	// Multiple metadata entries
	event.SetMetadata("account_id", "123456789")
	assert.Equal(t, "123456789", event.GetMetadata("account_id"))
	assert.Equal(t, "us-east-1", event.GetMetadata("region"))
}

// --- AWS Provider specific tests ---

func TestAWSProviderImplementsFullProvider(t *testing.T) {
	p := NewAWSProvider()

	// Check core interface
	var _ Provider = p

	// Check optional interfaces
	var _ ResourceDiscoverer = p
	var _ StateComparator = p

	caps := GetCapabilities(p)
	assert.True(t, caps.Discovery)
	assert.True(t, caps.Comparison)
}

func TestAWSProviderName(t *testing.T) {
	p := NewAWSProvider()
	assert.Equal(t, "aws", p.Name())
}

func TestAWSProviderWithRegions(t *testing.T) {
	p := NewAWSProvider(WithAWSRegions([]string{"us-west-2", "eu-west-1"}))
	assert.Equal(t, []string{"us-west-2", "eu-west-1"}, p.regions)
}

func TestAWSProviderParseEventMetadata(t *testing.T) {
	p := NewAWSProvider()

	// Use AuthorizeSecurityGroupIngress which has explicit field mapping
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"ct.name":              "AuthorizeSecurityGroupIngress",
		"ct.src":               "ec2.amazonaws.com",
		"ct.region":            "us-west-2",
		"ct.request.groupid":   "sg-12345678",
		"ct.user.type":         "IAMUser",
		"ct.user.principalid":  "AIDAEXAMPLE",
		"ct.user.arn":          "arn:aws:iam::123456:user/test",
		"ct.user.accountid":    "123456789012",
		"ct.user":              "test-user",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "AuthorizeSecurityGroupIngress", event.EventName)
	assert.Equal(t, "sg-12345678", event.ResourceID)
	assert.Equal(t, "us-west-2", event.GetMetadata("region"))
	assert.Equal(t, "123456789012", event.GetMetadata("account_id"))
	assert.Equal(t, "ec2.amazonaws.com", event.GetMetadata("event_source"))
	// Backward compatibility
	assert.Equal(t, "us-west-2", event.Region)
}

func TestAWSProviderRejectsNonCloudtrail(t *testing.T) {
	p := NewAWSProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{"ct.name": "RunInstances"}, nil)
	assert.Nil(t, event)
}

func TestAWSProviderSupportedDiscoveryTypes(t *testing.T) {
	p := NewAWSProvider()
	types := p.SupportedDiscoveryTypes()
	assert.Contains(t, types, "aws_vpc")
	assert.Contains(t, types, "aws_instance")
	assert.Contains(t, types, "aws_db_instance")
}

// --- GCP Provider tests ---

func TestGCPProviderImplementsProvider(t *testing.T) {
	p := NewGCPProvider()
	var _ Provider = p
	assert.Equal(t, "gcp", p.Name())

	// GCP now implements discovery and comparison
	caps := GetCapabilities(p)
	assert.True(t, caps.Discovery)
	assert.True(t, caps.Comparison)
}

func TestGCPProviderRejectsNonAudit(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{}, nil)
	assert.Nil(t, event)
}

func TestGCPProviderSupportedEventCount(t *testing.T) {
	p := NewGCPProvider()
	count := p.SupportedEventCount()
	assert.Greater(t, count, 0, "GCP should report supported events from mapper")
}

func TestGCPProviderSupportedResourceTypes(t *testing.T) {
	p := NewGCPProvider()
	rtypes := p.SupportedResourceTypes()
	assert.Greater(t, len(rtypes), 0, "GCP should report supported resource types from mapper")
}

// --- Azure Provider tests ---

func TestAzureProviderImplementsProvider(t *testing.T) {
	p := NewAzureProvider()
	var _ Provider = p
	assert.Equal(t, "azure", p.Name())

	caps := GetCapabilities(p)
	assert.False(t, caps.Discovery)
	assert.False(t, caps.Comparison)
}

func TestAzureProviderRejectsNonActivity(t *testing.T) {
	p := NewAzureProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{}, nil)
	assert.Nil(t, event)
}

func TestAzureProviderSupportedEventCount(t *testing.T) {
	p := NewAzureProvider()
	count := p.SupportedEventCount()
	assert.Greater(t, count, 0, "Azure should report supported events from mapper")
}

func TestAzureProviderSupportedResourceTypes(t *testing.T) {
	p := NewAzureProvider()
	rtypes := p.SupportedResourceTypes()
	assert.Greater(t, len(rtypes), 0, "Azure should report supported resource types from mapper")
}
