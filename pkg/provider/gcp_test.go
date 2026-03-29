package provider

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- GCP Provider Option Tests ---

func TestWithGCPProjectID(t *testing.T) {
	p := NewGCPProvider(WithGCPProjectID("my-project"))
	assert.Equal(t, "my-project", p.projectID)
}

func TestWithGCPRegions(t *testing.T) {
	regions := []string{"us-central1", "us-west1", "europe-west1"}
	p := NewGCPProvider(WithGCPRegions(regions))
	assert.Equal(t, regions, p.regions)
}

func TestNewGCPProvider_DefaultValues(t *testing.T) {
	p := NewGCPProvider()
	assert.NotNil(t, p.parser)
	assert.NotNil(t, p.mapper)
	assert.Empty(t, p.projectID)
	assert.Empty(t, p.regions)
}

func TestNewGCPProvider_MultipleOptions(t *testing.T) {
	regions := []string{"us-central1", "europe-west1"}
	p := NewGCPProvider(
		WithGCPProjectID("test-project"),
		WithGCPRegions(regions),
	)
	assert.Equal(t, "test-project", p.projectID)
	assert.Equal(t, regions, p.regions)
}

// --- GCP IsRelevantEvent Tests ---

func TestGCPIsRelevantEvent_ValidEvent(t *testing.T) {
	p := NewGCPProvider()
	// Google Compute Engine operations should be relevant
	result := p.IsRelevantEvent("compute.instances.insert")
	assert.True(t, result)
}

func TestGCPIsRelevantEvent_InvalidEvent(t *testing.T) {
	p := NewGCPProvider()
	result := p.IsRelevantEvent("nonexistent.operation")
	assert.False(t, result)
}

func TestGCPIsRelevantEvent_EmptyEvent(t *testing.T) {
	p := NewGCPProvider()
	result := p.IsRelevantEvent("")
	assert.False(t, result)
}

// --- GCP MapEventToResource Tests ---

func TestGCPMapEventToResource_ValidEvent(t *testing.T) {
	p := NewGCPProvider()
	resourceType := p.MapEventToResource("compute.instances.insert", "")
	assert.NotEmpty(t, resourceType)
}

func TestGCPMapEventToResource_UnknownEvent(t *testing.T) {
	p := NewGCPProvider()
	resourceType := p.MapEventToResource("unknown.event", "")
	assert.Equal(t, "", resourceType)
}

func TestGCPMapEventToResource_EventSourceIgnored(t *testing.T) {
	p := NewGCPProvider()
	// eventSource parameter should be ignored for GCP
	resourceType1 := p.MapEventToResource("compute.instances.insert", "source1")
	resourceType2 := p.MapEventToResource("compute.instances.insert", "source2")
	assert.Equal(t, resourceType1, resourceType2)
}

// --- GCP ExtractChanges Tests ---

func TestGCPExtractChanges_WithRequestAndResponse(t *testing.T) {
	p := NewGCPProvider()
	changes := p.ExtractChanges("compute.instances.insert", map[string]string{
		"gcp.request":  `{"machineType":"n1-standard-1"}`,
		"gcp.response": `{"id":"1234567890","name":"instance-1"}`,
	})

	assert.NotEmpty(t, changes)
	assert.Equal(t, `{"machineType":"n1-standard-1"}`, changes["request"])
	assert.Equal(t, `{"id":"1234567890","name":"instance-1"}`, changes["response"])
}

func TestGCPExtractChanges_EmptyFields(t *testing.T) {
	p := NewGCPProvider()
	changes := p.ExtractChanges("compute.instances.insert", map[string]string{})
	assert.Empty(t, changes)
}

func TestGCPExtractChanges_OnlyRequest(t *testing.T) {
	p := NewGCPProvider()
	changes := p.ExtractChanges("compute.instances.insert", map[string]string{
		"gcp.request": `{"name":"instance-1"}`,
	})

	assert.Len(t, changes, 1)
	assert.Equal(t, `{"name":"instance-1"}`, changes["request"])
}

func TestGCPExtractChanges_OnlyResponse(t *testing.T) {
	p := NewGCPProvider()
	changes := p.ExtractChanges("compute.instances.insert", map[string]string{
		"gcp.response": `{"id":"123"}`,
	})

	assert.Len(t, changes, 1)
	assert.Equal(t, `{"id":"123"}`, changes["response"])
}

// --- GCP SupportedEventCount Tests ---

func TestGCPSupportedEventCount_GreaterThanZero(t *testing.T) {
	p := NewGCPProvider()
	count := p.SupportedEventCount()
	assert.Greater(t, count, 0)
}

// --- GCP SupportedResourceTypes Tests ---

func TestGCPSupportedResourceTypes_NotEmpty(t *testing.T) {
	p := NewGCPProvider()
	types := p.SupportedResourceTypes()
	assert.NotEmpty(t, types)
}

func TestGCPSupportedResourceTypes_NoDuplicates(t *testing.T) {
	p := NewGCPProvider()
	types := p.SupportedResourceTypes()
	seen := make(map[string]bool)
	for _, rt := range types {
		assert.False(t, seen[rt], "Duplicate resource type: %s", rt)
		seen[rt] = true
	}
}

// --- GCP SupportedDiscoveryTypes Tests ---

func TestGCPSupportedDiscoveryTypes_ContainsExpected(t *testing.T) {
	p := NewGCPProvider()
	types := p.SupportedDiscoveryTypes()
	assert.Contains(t, types, "google_compute_network")
	assert.Contains(t, types, "google_compute_instance")
	assert.Contains(t, types, "google_storage_bucket")
}

func TestGCPSupportedDiscoveryTypes_Consistent(t *testing.T) {
	p := NewGCPProvider()
	types1 := p.SupportedDiscoveryTypes()
	types2 := p.SupportedDiscoveryTypes()
	assert.Equal(t, types1, types2)
}

// --- GCP DiscoverResources Tests ---

func TestGCPDiscoverResources_NoProjectID(t *testing.T) {
	p := NewGCPProvider()
	_, err := p.DiscoverResources(context.Background(), DiscoveryOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project ID")
}

func TestGCPDiscoverResources_ProjectIDRequired(t *testing.T) {
	p := NewGCPProvider(WithGCPProjectID(""))
	_, err := p.DiscoverResources(context.Background(), DiscoveryOptions{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "project ID")
}

func TestGCPDiscoverResources_ImplementsInterface(t *testing.T) {
	p := NewGCPProvider(WithGCPProjectID("test-project"))
	var _ ResourceDiscoverer = p
	assert.NotNil(t, p)
}

// --- GCP CompareState Tests ---

func TestGCPCompareState_EmptyInputs(t *testing.T) {
	p := NewGCPProvider()
	result := p.CompareState(nil, nil, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "gcp", result.Provider)
	assert.Empty(t, result.UnmanagedResources)
	assert.Empty(t, result.MissingResources)
	assert.Empty(t, result.ModifiedResources)
}

func TestGCPCompareState_NoResources(t *testing.T) {
	p := NewGCPProvider()
	result := p.CompareState([]*types.TerraformResource{}, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "gcp", result.Provider)
}

func TestGCPCompareState_WithTerraformResources(t *testing.T) {
	p := NewGCPProvider()
	tfResources := []*types.TerraformResource{
		{
			Type: "google_compute_instance",
			Name: "instance-1",
			Attributes: map[string]interface{}{
				"machine_type": "n1-standard-1",
			},
		},
	}
	result := p.CompareState(tfResources, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "gcp", result.Provider)
}

func TestGCPCompareState_WithActualResources(t *testing.T) {
	p := NewGCPProvider()
	actualResources := []*types.DiscoveredResource{
		{
			ID:       "projects/test-project/zones/us-central1-a/instances/instance-1",
			Type:     "google_compute_instance",
			Provider: "gcp",
			Name:     "instance-1",
			Region:   "us-central1",
		},
	}
	result := p.CompareState([]*types.TerraformResource{}, actualResources, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "gcp", result.Provider)
}

func TestGCPCompareState_BothResourcesEmpty(t *testing.T) {
	p := NewGCPProvider()
	result := p.CompareState([]*types.TerraformResource{}, []*types.DiscoveredResource{}, CompareOptions{})
	assert.NotNil(t, result)
	assert.Equal(t, "gcp", result.Provider)
	assert.Empty(t, result.UnmanagedResources)
	assert.Empty(t, result.MissingResources)
	assert.Empty(t, result.ModifiedResources)
}

// --- GCP ParseEvent Tests ---

func TestGCPParseEvent_CorrectSource(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{
		"gcp.methodName":  "compute.instances.insert",
		"gcp.resource.name": "projects/test-project/zones/us-central1-a/instances/instance-1",
		"gcp.user":        "user@example.com",
		"gcp.projectId":   "test-project",
		"gcp.serviceName": "compute.googleapis.com",
		"gcp.zone":        "us-central1-a",
		"gcp.region":      "us-central1",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "gcp", event.Provider)
	assert.Equal(t, "compute.instances.insert", event.EventName)
	assert.Equal(t, "user@example.com", event.UserIdentity.UserName)
	assert.Equal(t, "test-project", event.GetMetadata("project_id"))
	assert.Equal(t, "compute.googleapis.com", event.GetMetadata("service_name"))
	assert.Equal(t, "us-central1-a", event.GetMetadata("zone"))
	assert.Equal(t, "us-central1", event.GetMetadata("region"))
	// Backward compatibility
	assert.Equal(t, "test-project", event.ProjectID)
	assert.Equal(t, "compute.googleapis.com", event.ServiceName)
}

func TestGCPParseEvent_WrongSource(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("aws_cloudtrail", map[string]string{
		"gcp.methodName": "compute.instances.insert",
	}, nil)
	assert.Nil(t, event)
}

func TestGCPParseEvent_MissingMethodName(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{}, nil)
	assert.Nil(t, event)
}

func TestGCPParseEvent_IrrelevantEvent(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{
		"gcp.methodName": "unknown.operation",
	}, nil)
	assert.Nil(t, event)
}

func TestGCPParseEvent_PreParsedEvent(t *testing.T) {
	p := NewGCPProvider()
	preparsed := &types.Event{
		Provider:  "gcp",
		EventName: "compute.instances.insert",
		ProjectID: "test-project",
		ServiceName: "compute.googleapis.com",
		Metadata:  map[string]string{"key": "value"},
	}
	event := p.ParseEvent("gcpaudit", map[string]string{}, preparsed)

	require.NotNil(t, event)
	assert.Equal(t, "compute.instances.insert", event.EventName)
	assert.Equal(t, "test-project", event.GetMetadata("project_id"))
	assert.Equal(t, "compute.googleapis.com", event.GetMetadata("service_name"))
	assert.Equal(t, "value", event.GetMetadata("key"))
}

func TestGCPParseEvent_PreParsedEventNilMetadata(t *testing.T) {
	p := NewGCPProvider()
	preparsed := &types.Event{
		Provider:  "gcp",
		EventName: "compute.instances.insert",
		Metadata:  nil,
	}
	event := p.ParseEvent("gcpaudit", map[string]string{}, preparsed)

	require.NotNil(t, event)
	assert.NotNil(t, event.Metadata)
}

func TestGCPParseEvent_AllMetadataFields(t *testing.T) {
	p := NewGCPProvider()
	event := p.ParseEvent("gcpaudit", map[string]string{
		"gcp.methodName":   "storage.buckets.create",
		"gcp.resource.name": "projects/test-project/buckets/my-bucket",
		"gcp.user":         "user@example.com",
		"gcp.projectId":    "test-project",
		"gcp.serviceName":  "storage.googleapis.com",
		"gcp.zone":         "us-central1-a",
		"gcp.region":       "us-central1",
	}, nil)

	require.NotNil(t, event)
	assert.Equal(t, "test-project", event.GetMetadata("project_id"))
	assert.Equal(t, "storage.googleapis.com", event.GetMetadata("service_name"))
	assert.Equal(t, "us-central1-a", event.GetMetadata("zone"))
	assert.Equal(t, "us-central1", event.GetMetadata("region"))
}

// --- GCP Provider Interface Compliance ---

func TestGCPProviderImplementsInterfaces(t *testing.T) {
	p := NewGCPProvider()

	// Core Provider interface
	var _ Provider = p
	assert.Equal(t, "gcp", p.Name())

	// ResourceDiscoverer interface
	var _ ResourceDiscoverer = p
	discoveryTypes := p.SupportedDiscoveryTypes()
	assert.NotNil(t, discoveryTypes)

	// StateComparator interface
	var _ StateComparator = p
}

func TestGCPProviderWithRegionsOption(t *testing.T) {
	regions := []string{"europe-west1", "asia-east1"}
	p := NewGCPProvider(WithGCPRegions(regions))
	assert.Equal(t, regions, p.regions)
}
