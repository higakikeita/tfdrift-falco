package gcp

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewAuditParser(t *testing.T) {
	parser := NewAuditParser()
	assert.NotNil(t, parser)
	assert.NotNil(t, parser.mapper)
}

func TestAuditParser_Parse_ValidEvent(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.instances.setMetadata",
			"gcp.resource.name":                     "projects/my-project-123/zones/us-central1-a/instances/vm-1",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "user@example.com",
			"gcp.request":                           `{"metadata": {"items": [{"key": "ssh-keys", "value": "..."}]}}`,
		},
	}

	event := parser.Parse(res)

	assert.NotNil(t, event)
	assert.Equal(t, "gcp", event.Provider)
	assert.Equal(t, "compute.instances.setMetadata", event.EventName)
	assert.Equal(t, "google_compute_instance", event.ResourceType)
	assert.Equal(t, "vm-1", event.ResourceID)
	assert.Equal(t, "my-project-123", event.ProjectID)
	assert.Equal(t, "us-central1", event.Region) // Region is extracted from zone (us-central1-a -> us-central1)
	assert.Equal(t, "compute.googleapis.com", event.ServiceName)
	assert.Equal(t, "user@example.com", event.UserIdentity.UserName)
	assert.NotEmpty(t, event.Changes)
}

func TestAuditParser_Parse_NonGCPEvent(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name": "ModifyInstanceAttribute",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for non-GCP events")
}

func TestAuditParser_Parse_MissingMethodName(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source:       "gcpaudit",
		OutputFields: map[string]string{},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil when methodName is missing")
}

func TestAuditParser_Parse_IrrelevantEvent(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "storage.objects.get",
			"gcp.resource.name": "projects/my-project/buckets/my-bucket/objects/file.txt",
		},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for irrelevant events")
}

func TestAuditParser_isRelevantEvent(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name       string
		methodName string
		want       bool
	}{
		{"Compute Instance SetMetadata", "compute.instances.setMetadata", true},
		{"Compute Firewall Insert", "compute.firewalls.insert", true},
		{"SetIamPolicy", "SetIamPolicy", true},
		{"Storage Bucket Create", "storage.buckets.create", true},
		{"Irrelevant Event", "storage.objects.get", false},
		{"Unknown Event", "unknown.method", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.isRelevantEvent(tt.methodName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractResourceID(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
		want         string
	}{
		{
			"Compute Instance",
			"projects/my-project/zones/us-central1-a/instances/vm-1",
			"vm-1",
		},
		{
			"Cloud Storage Bucket",
			"projects/_/buckets/my-bucket",
			"my-bucket",
		},
		{
			"Single Name",
			"my-resource",
			"my-resource",
		},
		{
			"Empty",
			"",
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractResourceID(tt.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractProjectID(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
		fields       map[string]string
		want         string
	}{
		{
			"From Resource Name",
			"projects/my-project-123/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"my-project-123",
		},
		{
			"From Fields",
			"invalid/path",
			map[string]string{"gcp.resource.labels.project_id": "my-project-456"},
			"my-project-456",
		},
		{
			"Not Found",
			"invalid/path",
			map[string]string{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractProjectID(tt.resourceName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractZone(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
		fields       map[string]string
		want         string
	}{
		{
			"From Resource Name",
			"projects/my-project/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"us-central1-a",
		},
		{
			"From Fields",
			"projects/my-project/instances/vm-1",
			map[string]string{"gcp.resource.labels.zone": "us-west1-b"},
			"us-west1-b",
		},
		{
			"Not Found",
			"projects/my-project/instances/vm-1",
			map[string]string{},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractZone(tt.resourceName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractRegion(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name string
		zone string
		want string
	}{
		{"Valid Zone", "us-central1-a", "us-central1"},
		{"Another Zone", "europe-west1-b", "europe-west1"},
		{"Asia Zone", "asia-northeast1-c", "asia-northeast1"},
		{"Empty Zone", "", ""},
		{"Invalid Format", "invalid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractRegion(tt.zone)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractChanges(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name       string
		methodName string
		fields     map[string]string
		wantKeys   []string
	}{
		{
			"SetMetadata",
			"compute.instances.setMetadata",
			map[string]string{"gcp.request": `{"metadata": "..."}`},
			[]string{"metadata", "_raw_request"},
		},
		{
			"SetLabels",
			"compute.instances.setLabels",
			map[string]string{"gcp.request": `{"labels": "..."}`},
			[]string{"labels", "_raw_request"},
		},
		{
			"Create Instance",
			"compute.instances.insert",
			map[string]string{"gcp.response": `{"id": "123"}`},
			[]string{"_action", "_created_resource", "_raw_response"},
		},
		{
			"Delete Instance",
			"compute.instances.delete",
			map[string]string{},
			[]string{"_action"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractChanges(tt.methodName, tt.fields)
			assert.NotNil(t, got)

			for _, key := range tt.wantKeys {
				assert.Contains(t, got, key, "Expected key %s not found", key)
			}
		})
	}
}

func TestAuditParser_ValidateEvent(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name      string
		event     *types.Event
		wantError bool
	}{
		{
			"Valid Event",
			&types.Event{
				Provider:     "gcp",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
			},
			false,
		},
		{
			"Nil Event",
			nil,
			true,
		},
		{
			"Invalid Provider",
			&types.Event{
				Provider:     "aws",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
			},
			true,
		},
		{
			"Missing EventName",
			&types.Event{
				Provider:     "gcp",
				EventName:    "",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
			},
			true,
		},
		{
			"Missing ResourceType",
			&types.Event{
				Provider:     "gcp",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "",
				ResourceID:   "vm-1",
			},
			true,
		},
		{
			"Missing ResourceID",
			&types.Event{
				Provider:     "gcp",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "",
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateEvent(tt.event)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func Test_getStringField(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		key    string
		want   string
	}{
		{"Existing Key", map[string]string{"foo": "bar"}, "foo", "bar"},
		{"Missing Key", map[string]string{"foo": "bar"}, "baz", ""},
		{"Empty Map", map[string]string{}, "foo", ""},
		{"Empty Value", map[string]string{"foo": ""}, "foo", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringField(tt.fields, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}
