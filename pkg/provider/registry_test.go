package provider

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockProvider is a test implementation of the Provider interface
type mockProvider struct {
	name           string
	source         string
	eventCount     int
	resourceTypes  []string
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	if source != m.source {
		return nil
	}
	return &types.Event{
		Provider:  m.name,
		EventName: fields["event"],
	}
}

func (m *mockProvider) IsRelevantEvent(eventName string) bool { return true }

func (m *mockProvider) MapEventToResource(eventName string, eventSource string) string {
	return "mock_resource"
}

func (m *mockProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	return nil
}

func (m *mockProvider) SupportedEventCount() int         { return m.eventCount }
func (m *mockProvider) SupportedResourceTypes() []string { return m.resourceTypes }

func TestRegistryRegister(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "test", source: "test_source", eventCount: 10}

	err := r.Register(p)
	require.NoError(t, err)
	assert.Equal(t, 1, r.Count())

	// Duplicate registration should fail
	err = r.Register(p)
	assert.Error(t, err)
}

func TestRegistryGet(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "test", source: "test_source"}
	_ = r.Register(p)

	got, ok := r.Get("test")
	assert.True(t, ok)
	assert.Equal(t, "test", got.Name())

	_, ok = r.Get("nonexistent")
	assert.False(t, ok)
}

func TestRegistryRouteEvent(t *testing.T) {
	r := NewRegistry()
	awsP := &mockProvider{name: "aws", source: "aws_cloudtrail"}
	gcpP := &mockProvider{name: "gcp", source: "gcpaudit"}
	_ = r.Register(awsP)
	_ = r.Register(gcpP)

	// Route AWS event
	event, providerName := r.RouteEvent("aws_cloudtrail", map[string]string{"event": "CreateBucket"}, nil)
	assert.NotNil(t, event)
	assert.Equal(t, "aws", providerName)

	// Route GCP event
	event, providerName = r.RouteEvent("gcpaudit", map[string]string{"event": "compute.instances.insert"}, nil)
	assert.NotNil(t, event)
	assert.Equal(t, "gcp", providerName)

	// Unknown source
	event, providerName = r.RouteEvent("unknown", map[string]string{}, nil)
	assert.Nil(t, event)
	assert.Equal(t, "", providerName)
}

func TestRegistryAll(t *testing.T) {
	r := NewRegistry()
	_ = r.Register(&mockProvider{name: "aws", source: "aws_cloudtrail"})
	_ = r.Register(&mockProvider{name: "gcp", source: "gcpaudit"})

	all := r.All()
	assert.Len(t, all, 2)

	names := r.Names()
	assert.Len(t, names, 2)
	assert.Contains(t, names, "aws")
	assert.Contains(t, names, "gcp")
}
