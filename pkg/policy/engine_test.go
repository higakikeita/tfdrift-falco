package policy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine()
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	if e.ModuleCount() != 0 {
		t.Errorf("expected 0 modules, got %d", e.ModuleCount())
	}
}

func TestEvaluateNoPolicy(t *testing.T) {
	e := NewEngine()
	result, err := e.Evaluate(context.Background(), &DriftInput{
		Type:         "drift",
		ResourceType: "aws_instance",
		ResourceID:   "i-12345",
		Attribute:    "instance_type",
		Severity:     "high",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Decision != DecisionAlert {
		t.Errorf("expected decision %q, got %q", DecisionAlert, result.Decision)
	}
	if result.Reason != "no policy loaded" {
		t.Errorf("expected reason 'no policy loaded', got %q", result.Reason)
	}
}

const allowPolicy = `
package tfdrift

import rego.v1

default decision := "alert"
default reason := "default"

decision := "allow" if {
	input.resource_type == "aws_autoscaling_group"
	input.attribute == "desired_capacity"
}

reason := "autoscaling changes expected" if {
	input.resource_type == "aws_autoscaling_group"
	input.attribute == "desired_capacity"
}
`

const remediatePolicy = `
package tfdrift

import rego.v1

default decision := "alert"
default reason := "default"

decision := "remediate" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
}

reason := "open security group" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
}

severity := "critical" if {
	input.resource_type == "aws_security_group"
	input.attribute == "ingress"
}

labels := {"team": "security"} if {
	decision == "remediate"
}
`

const denyPolicy = `
package tfdrift

import rego.v1

default decision := "alert"
default reason := "default"

decision := "deny" if {
	startswith(input.resource_type, "aws_iam")
	input.user_identity.user_name == ""
}

reason := "IAM change by unknown user" if {
	startswith(input.resource_type, "aws_iam")
	input.user_identity.user_name == ""
}
`

func TestEvaluateAllowDecision(t *testing.T) {
	e := NewEngine()
	if err := e.LoadModule("allow.rego", allowPolicy); err != nil {
		t.Fatalf("load module: %v", err)
	}

	result, err := e.Evaluate(context.Background(), &DriftInput{
		Type:         "drift",
		ResourceType: "aws_autoscaling_group",
		ResourceID:   "asg-123",
		Attribute:    "desired_capacity",
		Severity:     "medium",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.Decision != DecisionAllow {
		t.Errorf("expected %q, got %q", DecisionAllow, result.Decision)
	}
	if result.Reason != "autoscaling changes expected" {
		t.Errorf("unexpected reason: %q", result.Reason)
	}
}

func TestEvaluateAlertDefault(t *testing.T) {
	e := NewEngine()
	if err := e.LoadModule("allow.rego", allowPolicy); err != nil {
		t.Fatalf("load module: %v", err)
	}

	// This input does NOT match the allow rule → default to alert
	result, err := e.Evaluate(context.Background(), &DriftInput{
		Type:         "drift",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		Attribute:    "instance_type",
		Severity:     "high",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.Decision != DecisionAlert {
		t.Errorf("expected %q, got %q", DecisionAlert, result.Decision)
	}
}

func TestEvaluateRemediateDecision(t *testing.T) {
	e := NewEngine()
	if err := e.LoadModule("remediate.rego", remediatePolicy); err != nil {
		t.Fatalf("load module: %v", err)
	}

	result, err := e.Evaluate(context.Background(), &DriftInput{
		Type:         "drift",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-123",
		Attribute:    "ingress",
		NewValue:     "0.0.0.0/0",
		Severity:     "medium",
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.Decision != DecisionRemediate {
		t.Errorf("expected %q, got %q", DecisionRemediate, result.Decision)
	}
	if result.Severity != "critical" {
		t.Errorf("expected severity override 'critical', got %q", result.Severity)
	}
	if result.Labels["team"] != "security" {
		t.Errorf("expected label team=security, got %v", result.Labels)
	}
}

func TestEvaluateDenyDecision(t *testing.T) {
	e := NewEngine()
	if err := e.LoadModule("deny.rego", denyPolicy); err != nil {
		t.Fatalf("load module: %v", err)
	}

	result, err := e.Evaluate(context.Background(), &DriftInput{
		Type:         "drift",
		ResourceType: "aws_iam_role",
		ResourceID:   "role-admin",
		Attribute:    "policy",
		Severity:     "high",
		UserIdentity: UserInput{UserName: ""},
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if result.Decision != DecisionDeny {
		t.Errorf("expected %q, got %q", DecisionDeny, result.Decision)
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()

	// Write a valid policy file
	content := `
package tfdrift

import rego.v1

default decision := "alert"
`
	if err := os.WriteFile(filepath.Join(dir, "test.rego"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Write a non-rego file (should be skipped)
	if err := os.WriteFile(filepath.Join(dir, "readme.txt"), []byte("not rego"), 0644); err != nil {
		t.Fatal(err)
	}

	e := NewEngine()
	if err := e.LoadDir(dir); err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if e.ModuleCount() != 1 {
		t.Errorf("expected 1 module, got %d", e.ModuleCount())
	}
}

func TestLoadDirNotFound(t *testing.T) {
	e := NewEngine()
	err := e.LoadDir("/nonexistent/path")
	if err == nil {
		t.Error("expected error for nonexistent dir")
	}
}

func TestLoadModuleInvalidRego(t *testing.T) {
	e := NewEngine()
	err := e.LoadModule("bad.rego", "this is not valid rego!!!")
	if err == nil {
		t.Error("expected error for invalid rego")
	}
}

func TestParseResultEdgeCases(t *testing.T) {
	// Unknown decision string → stays as default "alert"
	r := parseResult(map[string]interface{}{
		"decision": "unknown_decision",
	})
	if r.Decision != DecisionAlert {
		t.Errorf("expected default alert, got %q", r.Decision)
	}

	// Empty map
	r = parseResult(map[string]interface{}{})
	if r.Decision != DecisionAlert {
		t.Errorf("expected default alert, got %q", r.Decision)
	}

	// Suppressors as array
	r = parseResult(map[string]interface{}{
		"decision":    "allow",
		"suppressors": []interface{}{"autoscaling", "tag-only"},
	})
	if len(r.Suppressors) != 2 {
		t.Errorf("expected 2 suppressors, got %d", len(r.Suppressors))
	}
}
