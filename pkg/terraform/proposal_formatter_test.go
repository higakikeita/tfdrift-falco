package terraform

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

func TestFormatProposalMarkdown(t *testing.T) {
	proposal := &types.RemediationProposal{
		ID:            "test-id-123",
		AlertType:     "drift",
		Provider:      "aws",
		ResourceType:  "aws_instance",
		ResourceID:    "i-1234567890abcdef0",
		ResourceName:  "web_server",
		Severity:      "high",
		Description:   "Drift detected in instance type",
		TerraformCode: `resource "aws_instance" "web_server" {\n  instance_type = "t2.medium"\n}`,
		ImportCommand: "terraform import aws_instance.web_server i-1234567890abcdef0",
		PlanCommand:   "terraform plan -target=aws_instance.web_server",
		Status:        types.RemediationPending,
		CreatedAt:     "2026-03-29T10:00:00Z",
		Attributes: map[string]interface{}{
			"instance_type": "t2.medium",
			"ami":           "ami-12345678",
		},
	}

	md := FormatProposalMarkdown(proposal)

	// Test that required sections are present
	requiredSections := []string{
		"## Drift Auto-Remediation Proposal",
		"### Summary",
		"### Details",
		"### Proposed Terraform Code",
		"### Remediation Commands",
		"### Import Command",
		"### Plan Command",
		"### Detected Attributes",
		"### Next Steps",
		"Proposal ID: test-id-123",
	}

	for _, section := range requiredSections {
		if !strings.Contains(md, section) {
			t.Errorf("Markdown should contain section: %q", section)
		}
	}

	// Test that key details are present
	requiredDetails := []string{
		"drift",
		"high",
		"aws_instance",
		"web_server",
		"terraform import",
		"terraform plan",
	}

	for _, detail := range requiredDetails {
		if !strings.Contains(md, detail) {
			t.Errorf("Markdown should contain detail: %q, but got:\n%s", detail, md)
		}
	}

	// Test nil proposal
	emptyMD := FormatProposalMarkdown(nil)
	if emptyMD != "" {
		t.Errorf("expected empty string for nil proposal, got: %q", emptyMD)
	}
}

func TestFormatProposalJSON(t *testing.T) {
	proposal := &types.RemediationProposal{
		ID:            "test-id-456",
		AlertType:     "unmanaged",
		Provider:      "gcp",
		ResourceType:  "google_compute_instance",
		ResourceID:    "my-instance",
		ResourceName:  "app-server",
		Severity:      "medium",
		Description:   "Unmanaged resource detected",
		TerraformCode: `resource "google_compute_instance" "app_server" {}`,
		ImportCommand: "terraform import google_compute_instance.app_server my-instance",
		PlanCommand:   "terraform plan -target=google_compute_instance.app_server",
		Status:        types.RemediationApproved,
		CreatedAt:     "2026-03-29T11:00:00Z",
		PRUrl:         "https://github.com/example/repo/pull/123",
		PRNumber:      123,
		Attributes: map[string]interface{}{
			"machine_type": "n1-standard-1",
			"zone":         "us-central1-a",
		},
	}

	jsonBytes, err := FormatProposalJSON(proposal)
	if err != nil {
		t.Fatalf("failed to format proposal: %v", err)
	}

	// Parse and verify the JSON structure
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Verify required fields
	requiredFields := []string{
		"id", "alert_type", "provider", "resource_type", "resource_id",
		"resource_name", "severity", "description", "terraform_code",
		"import_command", "plan_command", "status", "created_at",
	}

	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON should contain field: %q", field)
		}
	}

	// Verify field values
	if result["id"] != "test-id-456" {
		t.Errorf("expected id 'test-id-456', got %v", result["id"])
	}
	if result["alert_type"] != "unmanaged" {
		t.Errorf("expected alert_type 'unmanaged', got %v", result["alert_type"])
	}
	if result["severity"] != "medium" {
		t.Errorf("expected severity 'medium', got %v", result["severity"])
	}
	if result["pr_number"] != float64(123) {
		t.Errorf("expected pr_number 123, got %v", result["pr_number"])
	}
	if result["pr_url"] != "https://github.com/example/repo/pull/123" {
		t.Errorf("expected pr_url to be set, got %v", result["pr_url"])
	}

	// Test nil proposal
	_, err = FormatProposalJSON(nil)
	if err == nil {
		t.Errorf("expected error for nil proposal, got nil")
	}
}

func TestFormatProposalJSONWithoutPR(t *testing.T) {
	proposal := &types.RemediationProposal{
		ID:            "test-id-789",
		AlertType:     "drift",
		Provider:      "aws",
		ResourceType:  "aws_s3_bucket",
		ResourceID:    "my-bucket",
		ResourceName:  "data-bucket",
		Severity:      "low",
		Description:   "Bucket versioning drift",
		TerraformCode: `resource "aws_s3_bucket" "data_bucket" {}`,
		ImportCommand: "terraform import aws_s3_bucket.data_bucket my-bucket",
		PlanCommand:   "terraform plan -target=aws_s3_bucket.data_bucket",
		Status:        types.RemediationPending,
		CreatedAt:     "2026-03-29T12:00:00Z",
		Attributes:    map[string]interface{}{},
	}

	jsonBytes, err := FormatProposalJSON(proposal)
	if err != nil {
		t.Fatalf("failed to format proposal: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	// Verify PR fields are not present or empty
	if prUrl, ok := result["pr_url"]; ok && prUrl != "" {
		t.Errorf("expected pr_url to be absent or empty, got %v", prUrl)
	}
	if prNumber, ok := result["pr_number"]; ok && prNumber != float64(0) {
		t.Errorf("expected pr_number to be absent or 0, got %v", prNumber)
	}
}

func TestFormatProposalMarkdownWithoutAttributes(t *testing.T) {
	proposal := &types.RemediationProposal{
		ID:            "test-id-999",
		AlertType:     "drift",
		ResourceType:  "aws_instance",
		ResourceID:    "i-999999",
		ResourceName:  "test",
		Severity:      "low",
		Description:   "Test drift",
		TerraformCode: `resource "aws_instance" "test" {}`,
		ImportCommand: "terraform import aws_instance.test i-999999",
		PlanCommand:   "terraform plan -target=aws_instance.test",
		Status:        types.RemediationPending,
		CreatedAt:     "2026-03-29T13:00:00Z",
		Attributes:    nil,
	}

	md := FormatProposalMarkdown(proposal)

	// Should not have Detected Attributes section if attributes are nil
	if strings.Contains(md, "### Detected Attributes") {
		t.Errorf("Markdown should not have Detected Attributes section for nil attributes")
	}

	// But should have all other sections
	if !strings.Contains(md, "### Summary") {
		t.Errorf("Markdown should have Summary section")
	}
	if !strings.Contains(md, "### Next Steps") {
		t.Errorf("Markdown should have Next Steps section")
	}
}
