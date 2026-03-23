package graph

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

func TestConvertDriftToCytoscape(t *testing.T) {
	drift := types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web-server",
		ResourceID:   "i-123",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		Timestamp:    "2024-01-01T00:00:00Z",
		AlertType:    "drift",
	}

	node := ConvertDriftToCytoscape(drift)
	if node.Data.ID != "i-123" {
		t.Errorf("ID = %q, want i-123", node.Data.ID)
	}
	if node.Data.Type != "drift" {
		t.Errorf("Type = %q, want drift", node.Data.Type)
	}
	if node.Data.Severity != "high" {
		t.Errorf("Severity = %q, want high", node.Data.Severity)
	}
	if node.Data.Metadata["attribute"] != "instance_type" {
		t.Errorf("attribute metadata = %v, want instance_type", node.Data.Metadata["attribute"])
	}
}

func TestConvertEventToCytoscape(t *testing.T) {
	event := types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-456",
		Region:       "us-east-1",
	}

	node := ConvertEventToCytoscape(event)
	if node.Data.ID != "i-456" {
		t.Errorf("ID = %q, want i-456", node.Data.ID)
	}
	if node.Data.Type != "falco_event" {
		t.Errorf("Type = %q, want falco_event", node.Data.Type)
	}
	if node.Data.Severity != "info" {
		t.Errorf("Severity = %q, want info", node.Data.Severity)
	}
}

func TestConvertUnmanagedToCytoscape(t *testing.T) {
	unmanaged := types.UnmanagedResourceAlert{
		Severity:     "medium",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-bucket",
		EventName:    "CreateBucket",
		Reason:       "not in terraform state",
	}

	node := ConvertUnmanagedToCytoscape(unmanaged)
	if node.Data.ID != "my-bucket" {
		t.Errorf("ID = %q, want my-bucket", node.Data.ID)
	}
	if node.Data.Type != "unmanaged" {
		t.Errorf("Type = %q, want unmanaged", node.Data.Type)
	}
	if node.Data.Metadata["reason"] != "not in terraform state" {
		t.Errorf("reason = %v", node.Data.Metadata["reason"])
	}
}

func TestConvertTerraformResourceToCytoscape(t *testing.T) {
	res := &terraform.Resource{
		Mode:     "managed",
		Type:     "aws_vpc",
		Name:     "main",
		Provider: "aws",
		Attributes: map[string]interface{}{
			"id":         "vpc-1",
			"name":       "main-vpc",
			"cidr_block": "10.0.0.0/16",
		},
	}

	// Not drifted
	node := ConvertTerraformResourceToCytoscape(res, false)
	if node.Data.Type != "terraform_resource" {
		t.Errorf("Type = %q, want terraform_resource", node.Data.Type)
	}
	if node.Data.Severity != "low" {
		t.Errorf("Severity = %q, want low", node.Data.Severity)
	}

	// Drifted
	nodeDrifted := ConvertTerraformResourceToCytoscape(res, true)
	if nodeDrifted.Data.Type != "terraform_resource_drifted" {
		t.Errorf("Type = %q, want terraform_resource_drifted", nodeDrifted.Data.Type)
	}
	if nodeDrifted.Data.Severity != "high" {
		t.Errorf("Severity = %q, want high", nodeDrifted.Data.Severity)
	}
}
