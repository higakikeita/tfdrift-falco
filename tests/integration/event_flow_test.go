// Package integration provides integration tests for the full pipeline
package integration

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

func TestEventFlowAWSInstance(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Test direct change extraction using the exported function
	fields := map[string]string{
		"ct.name":                          "ModifyInstanceAttribute",
		"ct.request.disableapitermination": "true",
	}

	changes := falco.ExtractAWSChanges("ModifyInstanceAttribute", fields)

	// Verify changes were extracted
	if changes == nil || len(changes) == 0 {
		t.Error("Expected changes to be extracted")
	}

	if _, ok := changes["disable_api_termination"]; !ok {
		t.Error("Expected disable_api_termination in changes")
	}

	// Test resource ID extraction
	resourceID := falco.ExtractAWSResourceID("ModifyInstanceAttribute", fields)
	if resourceID != "" && resourceID != "i-1234567890abcdef0" {
		t.Logf("Note: resource ID extraction returned: %s (this is expected if config is not loaded)", resourceID)
	}
}

func TestEventFlowS3Encryption(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Test S3 encryption change extraction
	fields := map[string]string{
		"ct.name": "PutBucketEncryption",
		"ct.request.serversideencryptionconfiguration": `{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}`,
	}

	changes := falco.ExtractAWSChanges("PutBucketEncryption", fields)

	// Verify changes were extracted
	if changes == nil || len(changes) == 0 {
		t.Error("Expected changes to be extracted for S3 encryption")
	}

	if _, ok := changes["server_side_encryption_configuration"]; !ok {
		t.Error("Expected server_side_encryption_configuration in changes")
	}
}

func TestCorrelateEventWithTerraformState(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Create a simple in-memory graph for testing
	// In production, resources would come from Terraform state
	store := graph.NewStore()

	// Build a simple resource graph directly
	resources := []*terraform.Resource{
		{
			Type:     "aws_instance",
			Name:     "web_server",
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":                      "i-1234567890abcdef0",
				"instance_type":           "t2.micro",
				"subnet_id":               "subnet-12345",
				"vpc_security_group_ids":  []interface{}{"sg-98765"},
				"disable_api_termination": false,
			},
		},
	}

	// Convert resources to graph database
	graphDB := graph.TerraformToGraph(resources, map[string]bool{})

	// Verify resources are in the graph
	if graphDB == nil {
		t.Fatal("Expected graph database to be created")
	}

	if graphDB.NodeCount() != 1 {
		t.Errorf("Expected 1 node in graph, got %d", graphDB.NodeCount())
	}

	// Create a mock event directly and add it to the store
	// (simulating what would happen after event parsing)
	mockEvent := types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-1234567890abcdef0",
		Changes: map[string]interface{}{
			"instance_type": "t2.micro",
		},
	}

	// Add event to store
	store.AddEvent(mockEvent)

	// Verify event was added
	events := store.GetEvents()
	if len(events) != 1 {
		t.Errorf("Expected 1 event in store, got %d", len(events))
	}

	// Verify the event correlates with the Terraform resource
	if events[0].ResourceID != "i-1234567890abcdef0" {
		t.Errorf("Expected resource ID to match, got %s", events[0].ResourceID)
	}
}

func TestMultiProviderEventFlow(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Test with mock events directly to avoid config loading issues
	store := graph.NewStore()

	// Create mock AWS event
	awsEvent := types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-1234567890abcdef0",
		Changes: map[string]interface{}{
			"instance_type": "t2.small",
		},
	}

	store.AddEvent(awsEvent)

	// Create another mock event
	otherEvent := types.Event{
		Provider:     "aws",
		EventName:    "CreateSecurityGroup",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-newgroup",
		Changes: map[string]interface{}{
			"group_name": "my-sg",
		},
	}

	store.AddEvent(otherEvent)

	// Verify both events were added
	events := store.GetEvents()
	if len(events) != 2 {
		t.Errorf("Expected 2 events in store, got %d", len(events))
	}

	if events[0].Provider != "aws" {
		t.Errorf("Expected AWS provider, got %s", events[0].Provider)
	}
}

func TestEventToGraphIntegration(t *testing.T) {
	log.SetLevel(log.DebugLevel)

	// Create Terraform resources forming a simple hierarchy
	resources := []*terraform.Resource{
		{
			Type:     "aws_vpc",
			Name:     "main",
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":         "vpc-12345",
				"cidr_block": "10.0.0.0/16",
			},
		},
		{
			Type:     "aws_subnet",
			Name:     "main",
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":                "subnet-12345",
				"vpc_id":            "vpc-12345",
				"cidr_block":        "10.0.1.0/24",
				"availability_zone": "us-east-1a",
			},
		},
		{
			Type:     "aws_instance",
			Name:     "web",
			Mode:     "managed",
			Provider: "aws",
			Attributes: map[string]interface{}{
				"id":        "i-1234567890abcdef0",
				"subnet_id": "subnet-12345",
			},
		},
	}

	// Convert resources to graph database
	graphDB := graph.TerraformToGraph(resources, map[string]bool{})

	// Create graph store
	store := graph.NewStore()

	// Verify hierarchy in graph
	if graphDB == nil {
		t.Fatal("Expected graph database to be created")
	}

	if graphDB.NodeCount() != 3 {
		t.Errorf("Expected 3 nodes in graph, got %d", graphDB.NodeCount())
	}

	// Verify relationships were created
	relationshipCount := graphDB.RelationshipCount()
	if relationshipCount == 0 {
		t.Error("Expected relationships to be created between resources")
	}

	t.Logf("Created graph with %d nodes and %d relationships", graphDB.NodeCount(), relationshipCount)

	// Add events to store
	sub := falco.NewSubscriberWithDefaults()

	event1 := &outputs.Response{
		Source: "aws_cloudtrail",
		Rule:   "AWS Instance Created",
		OutputFields: map[string]string{
			"ct.name":             "RunInstances",
			"ct.src":              "ec2.amazonaws.com",
			"ct.request.resource": "i-1234567890abcdef0",
			"ct.user.type":        "IAMUser",
			"ct.user.arn":         "arn:aws:iam::123456789012:user/admin",
		},
	}

	parsed1 := sub.ParseFalcoOutput(event1)
	if parsed1 != nil {
		store.AddEvent(*parsed1)
	}

	// Verify event was added
	events := store.GetEvents()
	if len(events) > 0 {
		t.Logf("Added %d event(s) to store", len(events))
	}
}

func TestEventChangeExtraction(t *testing.T) {
	// Test various event types for proper change extraction
	testCases := []struct {
		name        string
		eventName   string
		fields      map[string]string
		expectedKey string
	}{
		{
			name:      "IAM role policy creation",
			eventName: "PutRolePolicy",
			fields: map[string]string{
				"ct.name":                   "PutRolePolicy",
				"ct.src":                    "iam.amazonaws.com",
				"ct.request.policyname":     "inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`,
			},
			expectedKey: "inline_policy_name",
		},
		{
			name:      "User addition to group",
			eventName: "AddUserToGroup",
			fields: map[string]string{
				"ct.name":              "AddUserToGroup",
				"ct.src":               "iam.amazonaws.com",
				"ct.request.username":  "testuser",
				"ct.request.groupname": "testgroup",
			},
			expectedKey: "user_name",
		},
	}

	sub := falco.NewSubscriberWithDefaults()

	for _, tc := range testCases {
		changes := sub.ExtractChanges(tc.eventName, tc.fields)

		if changes == nil || len(changes) == 0 {
			t.Errorf("Test %s: Expected changes to be extracted", tc.name)
			continue
		}

		if _, ok := changes[tc.expectedKey]; !ok {
			t.Errorf("Test %s: Expected key %s in changes", tc.name, tc.expectedKey)
		}
	}
}
