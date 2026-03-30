package gcp

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// Additional comprehensive tests for edge cases and better coverage

func TestValuesEqual_NilBothNil(t *testing.T) {
	assert.True(t, valuesEqual(nil, nil))
}

func TestValuesEqual_BooleanStringComparisons(t *testing.T) {
	tests := []struct {
		name string
		a    interface{}
		b    interface{}
		want bool
	}{
		{"bool_true_string_true", true, "true", true},
		{"bool_false_string_false", false, "false", true},
		{"bool_true_string_false", true, "false", false},
		{"bool_false_string_true", false, "true", false},
		{"int_string_match", 42, "42", true},
		{"float_string_match", 3.14, "3.14", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.a, tt.b)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetNestedValue_DeepNesting(t *testing.T) {
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": "deep value",
			},
		},
	}
	result := getNestedValue(data, "level1.level2.level3")
	assert.Equal(t, "deep value", result)
}

func TestGetNestedValue_NonMapIntermediate(t *testing.T) {
	data := map[string]interface{}{
		"key": "string value",
	}
	result := getNestedValue(data, "key.nested")
	assert.Nil(t, result)
}

func TestLabelsEqual_WithGoogLabels(t *testing.T) {
	tfAttrs := map[string]interface{}{
		"labels": map[string]interface{}{
			"env":  "prod",
			"team": "backend",
		},
	}
	gcpLabels := map[string]string{
		"env":              "prod",
		"team":             "backend",
		"goog-managed-by":  "terraform",
		"gke-cluster-name": "my-cluster",
	}
	assert.True(t, labelsEqual(tfAttrs, gcpLabels))
}

func TestCompareResourceAttributes_TypeMismatch(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_instance",
		Attributes: map[string]interface{}{
			"name": "test",
		},
	}
	gcpRes := &DiscoveredResource{
		Type: "google_storage_bucket",
		Attributes: map[string]interface{}{
			"name": "test",
		},
	}
	result := compareResourceAttributes(tfRes, gcpRes)
	assert.Nil(t, result)
}

func TestFindByName_MultipleMatches(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_instance",
		Attributes: map[string]interface{}{
			"name": "web-server",
		},
	}
	gcpMap := map[string]*DiscoveredResource{
		"id1": {
			Type: "google_compute_instance",
			Name: "web-server",
		},
		"id2": {
			Type: "google_compute_instance",
			Name: "web-server",
		},
	}
	// Should return the first match
	result := findByName(tfRes, gcpMap)
	assert.NotNil(t, result)
	assert.Equal(t, "google_compute_instance", result.Type)
	assert.Equal(t, "web-server", result.Name)
}

func TestCompareResourceAttributes_AllDifferences(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_storage_bucket",
		Attributes: map[string]interface{}{
			"location":      "US",
			"storage_class": "STANDARD",
			"versioning":    true,
			"labels": map[string]interface{}{
				"env": "prod",
			},
		},
	}
	gcpRes := &DiscoveredResource{
		Type: "google_storage_bucket",
		ID:   "bucket-123",
		Attributes: map[string]interface{}{
			"location":      "EU",
			"storage_class": "NEARLINE",
			"versioning":    false,
		},
		Labels: map[string]string{
			"env": "dev",
		},
	}
	diffs := compareResourceAttributes(tfRes, gcpRes)
	assert.NotNil(t, diffs)
	assert.Greater(t, len(diffs), 0)

	// Should have differences
	fields := make(map[string]bool)
	for _, d := range diffs {
		fields[d.Field] = true
	}
	assert.True(t, fields["location"])
	assert.True(t, fields["storage_class"])
	assert.True(t, fields["versioning"])
}

func TestAuditParser_Parse_MissingResourceName(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName": "compute.instances.insert",
		},
	}
	event := parser.Parse(res)
	assert.Nil(t, event)
}

func TestAuditParser_Parse_EmptyMethodName(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "",
			"gcp.resource.name": "projects/123/zones/us-central1-a/instances/vm-1",
		},
	}
	event := parser.Parse(res)
	assert.Nil(t, event)
}

func TestAuditParser_Parse_UnmappedEvent(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":    "unknown.service.method",
			"gcp.resource.name": "projects/123/zones/us-central1-a/instances/vm-1",
		},
	}
	event := parser.Parse(res)
	assert.Nil(t, event)
}

func TestAuditParser_ExtractChanges_InsertAction(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.response": `{"id": "new-instance", "status": "RUNNING"}`,
	}
	changes := parser.extractChanges("compute.instances.insert", fields)
	assert.NotEmpty(t, changes)
	assert.Equal(t, "create", changes["_action"])
}

func TestAuditParser_ExtractChanges_DeleteAction(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{}
	changes := parser.extractChanges("compute.instances.delete", fields)
	assert.Equal(t, "delete", changes["_action"])
}

func TestAuditParser_ExtractChanges_UpdateAction(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"labels": {"env": "prod"}}`,
	}
	changes := parser.extractChanges("compute.instances.update", fields)
	assert.Equal(t, "update", changes["_action"])
}

func TestAuditParser_ExtractChanges_SetMetadata(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"items": [{"key": "startup-script"}]}`,
	}
	changes := parser.extractChanges("compute.instances.setMetadata", fields)
	assert.Contains(t, changes, "metadata")
}

func TestAuditParser_ExtractChanges_SetLabels(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"labels": {"env": "production"}}`,
	}
	changes := parser.extractChanges("compute.instances.setLabels", fields)
	assert.Contains(t, changes, "labels")
}

func TestAuditParser_ExtractChanges_SetTags(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `["http", "https"]`,
	}
	changes := parser.extractChanges("compute.instances.setTags", fields)
	assert.Contains(t, changes, "tags")
}

func TestAuditParser_ExtractChanges_SetIamPolicy(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"bindings": []}`,
	}
	changes := parser.extractChanges("SetIamPolicy", fields)
	assert.Contains(t, changes, "policy")
}

func TestAuditParser_ExtractChanges_CloudSQL(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"settings": {"tier": "db-f1-micro"}}`,
	}
	changes := parser.extractChanges("cloudsql.instances.patch", fields)
	assert.Equal(t, "update", changes["_action"])
}

func TestAuditParser_ExtractChanges_PublishTopic(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"topic": "projects/123/topics/my-topic"}`,
	}
	changes := parser.extractChanges("google.pubsub.v1.Publisher.CreateTopic", fields)
	assert.Equal(t, "create", changes["_action"])
}

func TestAuditParser_ExtractChanges_SecurityPolicy(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"rules": [{"action": "allow"}]}`,
	}
	changes := parser.extractChanges("compute.securityPolicies.patch", fields)
	assert.Contains(t, changes, "rules")
}

func TestAuditParser_ExtractProjectID_WithPrefix(t *testing.T) {
	parser := NewAuditParser()
	resourceName := "projects/test-project-123/zones/us-central1-a/instances/vm-1"
	projectID := parser.extractProjectIDFromName(resourceName, map[string]string{})
	assert.Equal(t, "test-project-123", projectID)
}

func TestAuditParser_ExtractZone_WithPrefix(t *testing.T) {
	parser := NewAuditParser()
	resourceName := "projects/123/zones/europe-west1-d/instances/vm-1"
	zone := parser.extractZoneFromName(resourceName, map[string]string{})
	assert.Equal(t, "europe-west1-d", zone)
}

func TestResourceMapper_GetResourceTypesForService_Compute(t *testing.T) {
	mapper := NewResourceMapper()

	// Test compute service
	computeTypes := mapper.GetResourceTypesForService("compute")
	assert.NotEmpty(t, computeTypes)
	assert.Contains(t, computeTypes, "google_compute_instance")
	assert.Contains(t, computeTypes, "google_compute_firewall")
	assert.Contains(t, computeTypes, "google_compute_network")
}

func TestResourceMapper_GetResourceTypesForService_NoMatches(t *testing.T) {
	mapper := NewResourceMapper()
	types := mapper.GetResourceTypesForService("nonexistent")
	assert.Empty(t, types)
}

func TestResourceMapper_GetAllSupportedEvents_Count(t *testing.T) {
	mapper := NewResourceMapper()
	events := mapper.GetAllSupportedEvents()
	assert.Greater(t, len(events), 100)
}

func TestAuditParser_ValidateEvent_MissingProvider(t *testing.T) {
	parser := NewAuditParser()
	event := &types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
	}
	err := parser.ValidateEvent(event)
	assert.Error(t, err)
}

func TestResourceMapper_MultipleServices(t *testing.T) {
	mapper := NewResourceMapper()

	// Verify that different services return correct types
	instanceTypes := mapper.GetResourceTypesForService("compute.instances")
	assert.Contains(t, instanceTypes, "google_compute_instance")

	storageTypes := mapper.GetResourceTypesForService("storage")
	assert.Contains(t, storageTypes, "google_storage_bucket")
}

func TestAuditParser_Parse_ComplexResourceName(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "container.clusters.create",
			"gcp.resource.name":                     "projects/my-gcp-project/zones/us-central1-a/clusters/prod-cluster",
			"gcp.serviceName":                       "container.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "terraform@my-project.iam.gserviceaccount.com",
		},
	}
	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "prod-cluster", event.ResourceID)
	assert.Equal(t, "my-gcp-project", event.ProjectID)
	assert.Equal(t, "us-central1", event.Region)
}

func TestGetTerraformLabels_NonStringValues(t *testing.T) {
	attrs := map[string]interface{}{
		"labels": map[string]interface{}{
			"env":     "prod",
			"count":   42,
			"enabled": true,
			"version": 3.14,
		},
	}
	labels := getTerraformLabels(attrs)
	assert.Equal(t, "prod", labels["env"])
	assert.NotContains(t, labels, "count")
	assert.NotContains(t, labels, "enabled")
	assert.NotContains(t, labels, "version")
}

func TestCompareResourceAttributes_FirewallResource(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_firewall",
		Attributes: map[string]interface{}{
			"network":   "default",
			"direction": "INGRESS",
			"priority":  1000,
			"disabled":  false,
		},
	}
	gcpRes := &DiscoveredResource{
		Type: "google_compute_firewall",
		ID:   "default-allow-http",
		Attributes: map[string]interface{}{
			"network":   "default",
			"direction": "INGRESS",
			"priority":  1000,
			"disabled":  false,
		},
	}
	diffs := compareResourceAttributes(tfRes, gcpRes)
	if diffs != nil {
		assert.Equal(t, 0, len(diffs))
	}
}

func TestCompareResourceAttributes_SQLInstance(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_sql_database_instance",
		Attributes: map[string]interface{}{
			"database_version":  "MYSQL_8_0",
			"region":            "us-central1",
			"tier":              "db-f1-micro",
			"availability_type": "ZONAL",
		},
	}
	gcpRes := &DiscoveredResource{
		Type: "google_sql_database_instance",
		ID:   "prod-db",
		Attributes: map[string]interface{}{
			"database_version":  "MYSQL_8_0",
			"region":            "us-central1",
			"tier":              "db-f1-micro",
			"availability_type": "ZONAL",
		},
	}
	diffs := compareResourceAttributes(tfRes, gcpRes)
	if diffs != nil {
		assert.Equal(t, 0, len(diffs))
	}
}

func TestCompareResourceAttributes_SubnetworkResource(t *testing.T) {
	tfRes := &TerraformResource{
		Type: "google_compute_subnetwork",
		Attributes: map[string]interface{}{
			"ip_cidr_range":            "10.0.0.0/24",
			"network":                  "default",
			"private_ip_google_access": true,
		},
	}
	gcpRes := &DiscoveredResource{
		Type: "google_compute_subnetwork",
		ID:   "default-us-central1",
		Attributes: map[string]interface{}{
			"ip_cidr_range":            "10.0.0.0/24",
			"network":                  "default",
			"private_ip_google_access": true,
		},
	}
	diffs := compareResourceAttributes(tfRes, gcpRes)
	if diffs != nil {
		assert.Equal(t, 0, len(diffs))
	}
}

func TestCompareStateWithActual_ComplexDrift(t *testing.T) {
	tfResources := []*TerraformResource{
		{
			Type: "google_compute_instance",
			Name: "web",
			Attributes: map[string]interface{}{
				"id":           "projects/my-project/zones/us-central1-a/instances/web",
				"name":         "web",
				"machine_type": "n1-standard-1",
			},
		},
	}
	gcpResources := []*DiscoveredResource{
		{
			ID:   "projects/my-project/zones/us-central1-a/instances/web",
			Type: "google_compute_instance",
			Name: "web",
			Attributes: map[string]interface{}{
				"machine_type": "n1-standard-2",
				"zone":         "us-central1-a",
			},
		},
		{
			ID:   "projects/my-project/zones/us-central1-a/instances/db",
			Type: "google_compute_instance",
			Name: "db",
			Attributes: map[string]interface{}{
				"machine_type": "n1-highmem-4",
				"zone":         "us-central1-a",
			},
		},
	}
	result := CompareStateWithActual(tfResources, gcpResources)
	assert.Greater(t, len(result.UnmanagedResources), 0)
	assert.Greater(t, len(result.ModifiedResources), 0)
}

// Additional extractChanges edge cases for full coverage
func TestAuditParser_ExtractChanges_SetMachineType(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"machine_type": "n1-standard-2"}`,
	}
	changes := parser.extractChanges("compute.instances.setMachineType", fields)
	assert.Contains(t, changes, "machine_type")
}

func TestAuditParser_ExtractChanges_SetServiceAccount(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"service_account": "sa@project.iam.gserviceaccount.com"}`,
	}
	changes := parser.extractChanges("compute.instances.setServiceAccount", fields)
	assert.Contains(t, changes, "service_account")
}

func TestAuditParser_ExtractChanges_SetDeletionProtection(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"deletion_protection": true}`,
	}
	changes := parser.extractChanges("compute.instances.setDeletionProtection", fields)
	assert.Contains(t, changes, "deletion_protection")
}

func TestAuditParser_ExtractChanges_UpdateDatabaseDdl(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"statements": ["CREATE TABLE ..."]}`,
	}
	changes := parser.extractChanges("google.spanner.admin.database.v1.DatabaseAdmin.UpdateDatabaseDdl", fields)
	assert.Contains(t, changes, "ddl")
}

func TestAuditParser_ExtractChanges_DNSChanges(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"changes": [{"rrsets": []}]}`,
	}
	changes := parser.extractChanges("dns.changes.create", fields)
	assert.Contains(t, changes, "rrdata")
}

func TestAuditParser_ExtractChanges_ModifyPushConfig(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.request": `{"push_config": {"push_endpoint": "https://example.com"}}`,
	}
	changes := parser.extractChanges("google.pubsub.v1.Subscriber.ModifyPushConfig", fields)
	assert.Contains(t, changes, "push_config")
}

func TestAuditParser_ExtractChanges_AllOptions(t *testing.T) {
	parser := NewAuditParser()
	tests := []struct {
		name       string
		methodName string
		fields     map[string]string
		wantField  string
	}{
		{
			name:       "createTopic",
			methodName: "google.pubsub.v1.Publisher.CreateTopic",
			fields:     map[string]string{"gcp.response": `{"name": "topic"}`},
			wantField:  "_action",
		},
		{
			name:       "deleteSubscription",
			methodName: "google.pubsub.v1.Subscriber.DeleteSubscription",
			fields:     map[string]string{},
			wantField:  "_action",
		},
		{
			name:       "updateDataset",
			methodName: "google.cloud.bigquery.v2.DatasetService.UpdateDataset",
			fields:     map[string]string{"gcp.request": `{"dataset": {}}`},
			wantField:  "_action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := parser.extractChanges(tt.methodName, tt.fields)
			assert.Contains(t, changes, tt.wantField)
		})
	}
}

func TestAuditParser_ExtractChanges_NoRequest(t *testing.T) {
	parser := NewAuditParser()
	changes := parser.extractChanges("compute.instances.delete", map[string]string{})
	assert.Equal(t, "delete", changes["_action"])
	assert.NotContains(t, changes, "_raw_request")
}

func TestAuditParser_ExtractChanges_OnlyResponse(t *testing.T) {
	parser := NewAuditParser()
	fields := map[string]string{
		"gcp.response": `{"id": "resource-123"}`,
	}
	changes := parser.extractChanges("storage.buckets.create", fields)
	assert.Contains(t, changes, "_created_resource")
	assert.Contains(t, changes, "_raw_response")
}

func TestAuditParser_Parse_WithAllFields(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.firewalls.insert",
			"gcp.resource.name":                     "projects/my-project/global/firewalls/allow-http",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "admin@company.com",
			"gcp.request":                           `{"sourceRanges": ["0.0.0.0/0"]}`,
			"gcp.response":                          `{"id": "3345"}`,
			"gcp.resource.labels.project_id":        "my-project",
		},
	}
	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "google_compute_firewall", event.ResourceType)
	assert.Equal(t, "allow-http", event.ResourceID)
	assert.Equal(t, "admin@company.com", event.UserIdentity.UserName)
	assert.Equal(t, "compute.googleapis.com", event.ServiceName)
	assert.NotEmpty(t, event.Changes)
}

func TestAuditParser_Parse_FirewallCreation(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.firewalls.insert",
			"gcp.resource.name":                     "projects/123/global/firewalls/allow-ssh",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "user@example.com",
		},
	}
	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "google_compute_firewall", event.ResourceType)
	assert.Equal(t, "allow-ssh", event.ResourceID)
}

func TestAuditParser_Parse_SubnetworkCreation(t *testing.T) {
	parser := NewAuditParser()
	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.subnetworks.insert",
			"gcp.resource.name":                     "projects/proj/regions/us-west1/subnetworks/subnet-1",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "terraform@proj.iam.gserviceaccount.com",
		},
	}
	event := parser.Parse(res)
	assert.NotNil(t, event)
	assert.Equal(t, "google_compute_subnetwork", event.ResourceType)
	assert.Equal(t, "subnet-1", event.ResourceID)
}

func TestAuditParser_ExtractResourceID_Empty(t *testing.T) {
	parser := NewAuditParser()
	id := parser.extractResourceIDFromName("")
	assert.Equal(t, "", id)
}

func TestAuditParser_ExtractResourceID_SingleSegment(t *testing.T) {
	parser := NewAuditParser()
	id := parser.extractResourceIDFromName("resource-name")
	assert.Equal(t, "resource-name", id)
}
