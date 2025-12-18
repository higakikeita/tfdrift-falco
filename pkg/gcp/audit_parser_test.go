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

// ============================================================================
// Edge Case Tests - Added for comprehensive coverage
// ============================================================================

func TestAuditParser_Parse_MalformedJSON(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "compute.instances.setMetadata",
			"gcp.resource.name": "projects/my-project/zones/us-central1-a/instances/vm-1",
			"gcp.serviceName":   "compute.googleapis.com",
			"gcp.request":       `{"metadata": INVALID_JSON}`, // Malformed JSON
		},
	}

	event := parser.Parse(res)
	// Should still parse but changes might be limited
	assert.NotNil(t, event)
	assert.Equal(t, "gcp", event.Provider)
}

func TestAuditParser_Parse_VeryLongResourceName(t *testing.T) {
	parser := NewAuditParser()

	// Create a very long resource name
	longName := "projects/my-project/zones/us-central1-a/instances/"
	for i := 0; i < 100; i++ {
		longName += "very-long-name-"
	}

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "compute.instances.setMetadata",
			"gcp.resource.name": longName,
			"gcp.serviceName":   "compute.googleapis.com",
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.NotEmpty(t, event.ResourceID)
}

func TestAuditParser_Parse_SpecialCharactersInResourceID(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
	}{
		{
			"Hyphens and underscores",
			"projects/my-project/zones/us-central1-a/instances/vm-test_123-prod",
		},
		{
			"Dots",
			"projects/my-project/zones/us-central1-a/instances/vm.test.123",
		},
		{
			"Mixed special chars",
			"projects/my-project/zones/us-central1-a/instances/vm-test_123.prod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":    "compute.instances.setMetadata",
					"gcp.resource.name": tt.resourceName,
					"gcp.serviceName":   "compute.googleapis.com",
				},
			}

			event := parser.Parse(res)
			assert.NotNil(t, event)
			assert.NotEmpty(t, event.ResourceID)
		})
	}
}

func TestAuditParser_Parse_UnicodeCharacters(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.instances.setLabels",
			"gcp.resource.name":                     "projects/my-project/zones/us-central1-a/instances/vm-1",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "ユーザー@example.com", // Japanese characters
			"gcp.request":                           `{"labels": {"name": "テスト"}}`,
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Contains(t, event.UserIdentity.UserName, "ユーザー")
}

func TestAuditParser_Parse_EmptyOutputFields(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source:       "gcpaudit",
		OutputFields: map[string]string{},
	}

	event := parser.Parse(res)
	assert.Nil(t, event, "Should return nil for empty output fields")
}

func TestAuditParser_Parse_NilResponse(t *testing.T) {
	parser := NewAuditParser()

	event := parser.Parse(nil)
	assert.Nil(t, event, "Should handle nil response gracefully")
}

func TestAuditParser_extractResourceID_MultipleSlashes(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
		want         string
	}{
		{
			"Multiple consecutive slashes",
			"projects/my-project///zones//us-central1-a///instances///vm-1",
			"vm-1",
		},
		{
			"Trailing slash",
			"projects/my-project/zones/us-central1-a/instances/vm-1/",
			"",
		},
		{
			"Leading slash",
			"/projects/my-project/zones/us-central1-a/instances/vm-1",
			"vm-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractResourceID(tt.resourceName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractProjectID_EdgeCases(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name         string
		resourceName string
		fields       map[string]string
		want         string
	}{
		{
			"Project ID with numbers",
			"projects/my-project-12345/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"my-project-12345",
		},
		{
			"Very short project ID",
			"projects/p1/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"p1",
		},
		{
			"Project ID with underscores",
			"projects/my_project_123/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"my_project_123",
		},
		{
			"Multiple projects keyword",
			"projects/projects/my-project/zones/us-central1-a/instances/vm-1",
			map[string]string{},
			"projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractProjectID(tt.resourceName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractRegion_EdgeCases(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name string
		zone string
		want string
	}{
		{"Single letter zone", "us-central1-z", "us-central1"},
		{"Multi-digit zone", "us-central1-123", "us-central1"},
		{"Zone with only region", "us-central1", "us-central1"},
		{"Very long region name", "australia-southeast1-a", "australia-southeast1"},
		{"Special format", "us-east4-c", "us-east4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.extractRegion(tt.zone)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_extractChanges_LargePayload(t *testing.T) {
	parser := NewAuditParser()

	// Create a large JSON payload
	largePayload := `{"metadata": {"items": [`
	for i := 0; i < 100; i++ {
		if i > 0 {
			largePayload += ","
		}
		largePayload += `{"key": "key` + string(rune(i)) + `", "value": "value` + string(rune(i)) + `"}`
	}
	largePayload += `]}}`

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName": "compute.instances.setMetadata",
			"gcp.request":    largePayload,
		},
	}

	changes := parser.extractChanges("compute.instances.setMetadata", res.OutputFields)
	assert.NotNil(t, changes)
	assert.NotEmpty(t, changes)
}

func TestAuditParser_isRelevantEvent_EdgeCases(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name       string
		methodName string
		want       bool
	}{
		{"Empty method name", "", false},
		{"Very long method name", "compute.instances.setMetadata.with.many.dots.and.long.name.that.exceeds.normal.length", false},
		{"Method with spaces", "compute instances setMetadata", false},
		{"Method with special chars", "compute.instances.set@Metadata", false},
		{"Case sensitivity", "COMPUTE.INSTANCES.SETMETADATA", false},
		{"Partial match", "compute.instances", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.isRelevantEvent(tt.methodName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuditParser_Parse_MissingOptionalFields(t *testing.T) {
	parser := NewAuditParser()

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "compute.instances.setMetadata",
			"gcp.resource.name": "projects/my-project/zones/us-central1-a/instances/vm-1",
			// Missing: serviceName, principalEmail, request, etc.
		},
	}

	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "gcp", event.Provider)
	assert.Equal(t, "google_compute_instance", event.ResourceType)
	// Optional fields should have default/empty values
	assert.Empty(t, event.UserIdentity.UserName)
}

func TestAuditParser_Parse_AllServiceTypes(t *testing.T) {
	parser := NewAuditParser()

	services := []struct {
		service      string
		methodName   string
		resourceName string
		wantType     string
	}{
		{"compute.googleapis.com", "compute.instances.insert", "projects/p/zones/z/instances/i", "google_compute_instance"},
		{"compute.googleapis.com", "compute.firewalls.update", "projects/p/global/firewalls/f", "google_compute_firewall"},
		{"sqladmin.googleapis.com", "cloudsql.instances.create", "projects/p/instances/db", "google_sql_database_instance"},
		{"container.googleapis.com", "container.clusters.create", "projects/p/locations/l/clusters/c", "google_container_cluster"},
		{"storage.googleapis.com", "storage.buckets.create", "projects/_/buckets/b", "google_storage_bucket"},
	}

	for _, s := range services {
		t.Run(s.service, func(t *testing.T) {
			res := &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":    s.methodName,
					"gcp.resource.name": s.resourceName,
					"gcp.serviceName":   s.service,
				},
			}

			event := parser.Parse(res)
			assert.NotNil(t, event)
			assert.Equal(t, s.wantType, event.ResourceType)
		})
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestAuditParser_Parse_ConcurrentAccess(t *testing.T) {
	parser := NewAuditParser()

	// Test concurrent parsing
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			res := &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":    "compute.instances.setMetadata",
					"gcp.resource.name": "projects/my-project/zones/us-central1-a/instances/vm-" + string(rune(id)),
					"gcp.serviceName":   "compute.googleapis.com",
				},
			}
			event := parser.Parse(res)
			assert.NotNil(t, event)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAuditParser_ValidateEvent_ExtensiveCases(t *testing.T) {
	parser := NewAuditParser()

	tests := []struct {
		name      string
		event     *types.Event
		wantError bool
		errorMsg  string
	}{
		{
			"All fields populated",
			&types.Event{
				Provider:     "gcp",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
				ProjectID:    "my-project",
				Region:       "us-central1",
				ServiceName:  "compute.googleapis.com",
			},
			false,
			"",
		},
		{
			"Empty provider",
			&types.Event{
				Provider:     "",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
			},
			true,
			"invalid provider",
		},
		{
			"Wrong provider",
			&types.Event{
				Provider:     "azure",
				EventName:    "compute.instances.setMetadata",
				ResourceType: "google_compute_instance",
				ResourceID:   "vm-1",
			},
			true,
			"invalid provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parser.ValidateEvent(tt.event)
			if tt.wantError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
