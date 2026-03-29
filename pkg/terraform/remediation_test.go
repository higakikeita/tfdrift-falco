package terraform

import (
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

func TestRemediationGeneratorForDrift(t *testing.T) {
	tests := []struct {
		name           string
		alert          *types.DriftAlert
		expectNil      bool
		expectID       bool
		expectType     string
		checkTerraform bool
	}{
		{
			name: "valid drift alert",
			alert: &types.DriftAlert{
				Severity:     "high",
				ResourceType: "aws_instance",
				ResourceName: "web_server",
				ResourceID:   "i-1234567890abcdef0",
				Attribute:    "tags",
				OldValue:     map[string]string{"Name": "old"},
				NewValue:     map[string]string{"Name": "new"},
			},
			expectNil:      false,
			expectID:       true,
			expectType:     "drift",
			checkTerraform: true,
		},
		{
			name:      "nil alert",
			alert:     nil,
			expectNil: true,
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proposal := gen.GenerateForDrift(tc.alert)

			if tc.expectNil && proposal != nil {
				t.Errorf("expected nil proposal, got %v", proposal)
			}
			if !tc.expectNil && proposal == nil {
				t.Errorf("expected non-nil proposal")
			}

			if !tc.expectNil {
				if tc.expectID && proposal.ID == "" {
					t.Errorf("expected non-empty ID")
				}

				if proposal.AlertType != tc.expectType {
					t.Errorf("expected AlertType %q, got %q", tc.expectType, proposal.AlertType)
				}

				if proposal.Status != types.RemediationPending {
					t.Errorf("expected status %q, got %q", types.RemediationPending, proposal.Status)
				}

				if tc.checkTerraform && proposal.TerraformCode == "" {
					t.Errorf("expected non-empty TerraformCode")
				}

				if proposal.ImportCommand == "" {
					t.Errorf("expected non-empty ImportCommand")
				}

				if proposal.PlanCommand == "" {
					t.Errorf("expected non-empty PlanCommand")
				}
			}
		})
	}
}

func TestRemediationGeneratorForUnmanaged(t *testing.T) {
	tests := []struct {
		name       string
		event      *types.Event
		expectNil  bool
		expectType string
	}{
		{
			name: "valid unmanaged event",
			event: &types.Event{
				Provider:     "aws",
				EventName:    "RunInstances",
				ResourceType: "aws_instance",
				ResourceID:   "i-0987654321fedcba0",
				Changes: map[string]interface{}{
					"instance_type": "t2.micro",
					"ami":           "ami-12345678",
				},
			},
			expectNil:  false,
			expectType: "unmanaged",
		},
		{
			name:      "nil event",
			event:     nil,
			expectNil: true,
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proposal := gen.GenerateForUnmanaged(tc.event)

			if tc.expectNil && proposal != nil {
				t.Errorf("expected nil proposal, got %v", proposal)
			}
			if !tc.expectNil && proposal == nil {
				t.Errorf("expected non-nil proposal")
			}

			if !tc.expectNil {
				if proposal.ID == "" {
					t.Errorf("expected non-empty ID")
				}

				if proposal.AlertType != tc.expectType {
					t.Errorf("expected AlertType %q, got %q", tc.expectType, proposal.AlertType)
				}

				if proposal.Status != types.RemediationPending {
					t.Errorf("expected status %q, got %q", types.RemediationPending, proposal.Status)
				}
			}
		})
	}
}

func TestRemediation_GenerateImportCommand(t *testing.T) {
	tests := []struct {
		name              string
		resourceType      string
		resourceName      string
		resourceID        string
		expectedSubstring string
	}{
		{
			name:              "aws instance",
			resourceType:      "aws_instance",
			resourceName:      "web",
			resourceID:        "i-12345",
			expectedSubstring: "terraform import aws_instance.web i-12345",
		},
		{
			name:              "with empty resource name",
			resourceType:      "aws_instance",
			resourceName:      "",
			resourceID:        "i-12345",
			expectedSubstring: "terraform import aws_instance.imported_resource i-12345",
		},
		{
			name:              "gcp compute instance",
			resourceType:      "google_compute_instance",
			resourceName:      "app",
			resourceID:        "my-instance",
			expectedSubstring: "terraform import google_compute_instance.app my-instance",
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := gen.generateImportCommand(tc.resourceType, tc.resourceName, tc.resourceID)
			if cmd != tc.expectedSubstring {
				t.Errorf("expected %q, got %q", tc.expectedSubstring, cmd)
			}
		})
	}
}

func TestGeneratePlanCommand(t *testing.T) {
	tests := []struct {
		name              string
		resourceType      string
		resourceName      string
		expectedSubstring string
	}{
		{
			name:              "aws instance",
			resourceType:      "aws_instance",
			resourceName:      "web",
			expectedSubstring: "terraform plan -target=aws_instance.web",
		},
		{
			name:              "with empty resource name",
			resourceType:      "aws_instance",
			resourceName:      "",
			expectedSubstring: "terraform plan -target=aws_instance.imported_resource",
		},
		{
			name:              "gcp instance",
			resourceType:      "google_compute_instance",
			resourceName:      "prod",
			expectedSubstring: "terraform plan -target=google_compute_instance.prod",
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := gen.generatePlanCommand(tc.resourceType, tc.resourceName)
			if cmd != tc.expectedSubstring {
				t.Errorf("expected %q, got %q", tc.expectedSubstring, cmd)
			}
		})
	}
}

func TestGenerateHCL(t *testing.T) {
	tests := []struct {
		name             string
		resourceType     string
		resourceName     string
		attributes       map[string]interface{}
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:         "basic resource",
			resourceType: "aws_instance",
			resourceName: "web",
			attributes:   nil,
			shouldContain: []string{
				`resource "aws_instance" "web"`,
				"# Add required attributes below",
			},
		},
		{
			name:         "resource with attributes",
			resourceType: "aws_instance",
			resourceName: "app",
			attributes: map[string]interface{}{
				"instance_type": "t2.micro",
				"ami":           "ami-12345678",
			},
			shouldContain: []string{
				`resource "aws_instance" "app"`,
				"# Attributes from cloud provider:",
				"instance_type",
				"ami",
			},
		},
		{
			name:         "empty resource name defaults",
			resourceType: "aws_instance",
			resourceName: "",
			attributes:   nil,
			shouldContain: []string{
				`resource "aws_instance" "imported_resource"`,
			},
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hcl := gen.generateHCL(tc.resourceType, tc.resourceName, tc.attributes)

			for _, substring := range tc.shouldContain {
				if !strings.Contains(hcl, substring) {
					t.Errorf("HCL should contain %q, but got:\n%s", substring, hcl)
				}
			}

			for _, substring := range tc.shouldNotContain {
				if strings.Contains(hcl, substring) {
					t.Errorf("HCL should not contain %q, but got:\n%s", substring, hcl)
				}
			}
		})
	}
}

func TestGenerateDriftFixHCL(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  string
		resourceName  string
		attribute     string
		oldValue      interface{}
		newValue      interface{}
		shouldContain []string
	}{
		{
			name:         "string attribute change",
			resourceType: "aws_instance",
			resourceName: "web",
			attribute:    "instance_type",
			oldValue:     "t2.small",
			newValue:     "t2.medium",
			shouldContain: []string{
				`resource "aws_instance" "web"`,
				"instance_type = \"t2.medium\"",
				"# Drift remediation",
			},
		},
		{
			name:         "numeric attribute change",
			resourceType: "aws_ebs_volume",
			resourceName: "data",
			attribute:    "size",
			oldValue:     10,
			newValue:     20,
			shouldContain: []string{
				`resource "aws_ebs_volume" "data"`,
				"size = 20",
			},
		},
		{
			name:         "bool attribute change",
			resourceType: "aws_instance",
			resourceName: "app",
			attribute:    "associate_public_ip_address",
			oldValue:     false,
			newValue:     true,
			shouldContain: []string{
				`resource "aws_instance" "app"`,
				"associate_public_ip_address = true",
			},
		},
		{
			name:         "empty resource name defaults",
			resourceType: "aws_instance",
			resourceName: "",
			attribute:    "tags",
			oldValue:     "old",
			newValue:     "new",
			shouldContain: []string{
				`resource "aws_instance" "resource"`,
			},
		},
	}

	gen := NewRemediationGenerator()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hcl := gen.generateDriftFixHCL(
				tc.resourceType, tc.resourceName, tc.attribute, tc.oldValue, tc.newValue,
			)

			for _, substring := range tc.shouldContain {
				if !strings.Contains(hcl, substring) {
					t.Errorf("HCL should contain %q, but got:\n%s", substring, hcl)
				}
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "string value",
			value:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "bool true",
			value:    true,
			expected: "true",
		},
		{
			name:     "bool false",
			value:    false,
			expected: "false",
		},
		{
			name:     "integer",
			value:    42,
			expected: "42",
		},
		{
			name:     "float",
			value:    3.14,
			expected: "3.14",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatValue(tc.value)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
