package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

func sampleDrift() *types.DriftResult {
	return &types.DriftResult{
		Provider: "aws",
		UnmanagedResources: []*types.DiscoveredResource{
			{Type: "aws_security_group", ID: "sg-123", Region: "us-east-1"},
		},
		MissingResources: []*types.TerraformResource{
			{Type: "aws_instance", Name: "web", ID: "i-999"},
		},
		ModifiedResources: []*types.ResourceDiff{
			{ResourceType: "aws_db_instance", ResourceID: "db-1", Differences: []types.FieldDiff{
				{Field: "instance_class", TerraformValue: "db.t3.small", ActualValue: "db.t3.large"},
			}},
		},
	}
}

func TestDriftTotal(t *testing.T) {
	if got := driftTotal(sampleDrift()); got != 3 {
		t.Fatalf("driftTotal = %d, want 3", got)
	}
	if got := driftTotal(nil); got != 0 {
		t.Fatalf("driftTotal(nil) = %d, want 0", got)
	}
	if got := driftTotal(&types.DriftResult{}); got != 0 {
		t.Fatalf("driftTotal(empty) = %d, want 0", got)
	}
}

func TestExitCodeForDrift(t *testing.T) {
	cases := []struct {
		total       int
		failOnDrift bool
		want        int
	}{
		{0, true, 0},                   // clean -> 0
		{3, true, 3},                   // drift -> count
		{5, false, 0},                  // reporting only -> 0 even with drift
		{9999, true, maxDriftExitCode}, // capped
	}
	for _, c := range cases {
		if got := exitCodeForDrift(c.total, c.failOnDrift); got != c.want {
			t.Errorf("exitCodeForDrift(%d,%v) = %d, want %d", c.total, c.failOnDrift, got, c.want)
		}
	}
}

func TestRenderDriftReport_HumanCleanVsDrift(t *testing.T) {
	clean := renderDriftReport(&types.DriftResult{}, "human", 10, 10, []string{"us-east-1"})
	if !strings.Contains(clean, "No drift") {
		t.Errorf("clean report should say No drift, got:\n%s", clean)
	}

	rep := renderDriftReport(sampleDrift(), "human", 10, 11, []string{"us-east-1"})
	for _, want := range []string{
		"Drift detected: 3", "unmanaged=1", "missing=1", "modified=1",
		"sg-123", "aws_instance.web", "db-1", "instance_class",
	} {
		if !strings.Contains(rep, want) {
			t.Errorf("human report missing %q; got:\n%s", want, rep)
		}
	}
}

func TestRenderDriftReport_JSONShape(t *testing.T) {
	out := renderDriftReport(sampleDrift(), "json", 10, 11, []string{"us-east-1", "ap-northeast-1"})
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("json output must parse: %v\n%s", err, out)
	}
	summary, ok := parsed["summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("json output missing summary object")
	}
	if summary["total_drift"].(float64) != 3 {
		t.Errorf("summary.total_drift = %v, want 3", summary["total_drift"])
	}
	if _, ok := parsed["drift"]; !ok {
		t.Errorf("json output must include drift detail")
	}
}
