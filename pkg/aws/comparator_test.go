package aws

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
)

// --- CompareStateWithActual ---

func TestCompareStateWithActual_AllMatching(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-1", "cidr_block": "10.0.0.0/16"}},
	}
	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc", Attributes: map[string]interface{}{"cidr_block": "10.0.0.0/16"}, Tags: map[string]string{}},
	}

	result := CompareStateWithActual(tfResources, awsResources)
	if len(result.UnmanagedResources) != 0 {
		t.Errorf("unmanaged = %d, want 0", len(result.UnmanagedResources))
	}
	if len(result.MissingResources) != 0 {
		t.Errorf("missing = %d, want 0", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_UnmanagedResource(t *testing.T) {
	tfResources := []*terraform.Resource{}
	awsResources := []*DiscoveredResource{
		{ID: "vpc-orphan", Type: "aws_vpc", Attributes: map[string]interface{}{}, Tags: map[string]string{}},
	}

	result := CompareStateWithActual(tfResources, awsResources)
	if len(result.UnmanagedResources) != 1 {
		t.Errorf("unmanaged = %d, want 1", len(result.UnmanagedResources))
	}
}

func TestCompareStateWithActual_MissingResource(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-deleted"}},
	}
	awsResources := []*DiscoveredResource{}

	result := CompareStateWithActual(tfResources, awsResources)
	if len(result.MissingResources) != 1 {
		t.Errorf("missing = %d, want 1", len(result.MissingResources))
	}
}

func TestCompareStateWithActual_ModifiedResource(t *testing.T) {
	tfResources := []*terraform.Resource{
		{Type: "aws_vpc", Attributes: map[string]interface{}{"id": "vpc-1", "cidr_block": "10.0.0.0/16"}},
	}
	awsResources := []*DiscoveredResource{
		{ID: "vpc-1", Type: "aws_vpc", Attributes: map[string]interface{}{"cidr_block": "10.0.0.0/8"}, Tags: map[string]string{}},
	}

	result := CompareStateWithActual(tfResources, awsResources)
	if len(result.ModifiedResources) != 1 {
		t.Errorf("modified = %d, want 1", len(result.ModifiedResources))
	}
	if len(result.ModifiedResources) > 0 {
		diff := result.ModifiedResources[0]
		found := false
		for _, d := range diff.Differences {
			if d.Field == "cidr_block" {
				found = true
			}
		}
		if !found {
			t.Error("expected cidr_block difference")
		}
	}
}

// --- extractTFResourceID ---

func TestExtractTFResourceID(t *testing.T) {
	tests := []struct {
		name string
		res  *terraform.Resource
		want string
	}{
		{"id field", &terraform.Resource{Attributes: map[string]interface{}{"id": "vpc-1"}}, "vpc-1"},
		{"instance_id", &terraform.Resource{Attributes: map[string]interface{}{"instance_id": "i-1"}}, "i-1"},
		{"vpc_id", &terraform.Resource{Attributes: map[string]interface{}{"vpc_id": "vpc-2"}}, "vpc-2"},
		{"arn", &terraform.Resource{Attributes: map[string]interface{}{"arn": "arn:aws:..."}}, "arn:aws:..."},
		{"empty", &terraform.Resource{Attributes: map[string]interface{}{}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractTFResourceID(tt.res)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// --- getComparableFields ---

func TestGetComparableFields(t *testing.T) {
	tests := []struct {
		resType string
		minLen  int
	}{
		{"aws_vpc", 3},
		{"aws_subnet", 4},
		{"aws_instance", 4},
		{"aws_db_instance", 7},
		{"aws_unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.resType, func(t *testing.T) {
			fields := getComparableFields(tt.resType)
			if len(fields) < tt.minLen {
				t.Errorf("fields = %d, want >= %d", len(fields), tt.minLen)
			}
		})
	}
}

// --- getNestedValue ---

func TestGetNestedValue(t *testing.T) {
	data := map[string]interface{}{
		"cidr_block": "10.0.0.0/16",
		"nested": map[string]interface{}{
			"inner": "value",
		},
	}

	tests := []struct {
		path string
		want interface{}
	}{
		{"cidr_block", "10.0.0.0/16"},
		{"nested.inner", "value"},
		{"missing", nil},
		{"nested.missing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := getNestedValue(data, tt.path)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

// --- valuesEqual ---

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"both nil", nil, nil, true},
		{"a nil", nil, "x", false},
		{"b nil", "x", nil, false},
		{"same string", "hello", "hello", true},
		{"diff string", "hello", "world", false},
		{"same bool", true, true, true},
		{"diff bool", true, false, false},
		{"bool vs string true", true, "true", true},
		{"bool vs string false", false, "false", true},
		{"same int", 42, 42, true},
		{"string comparison", "10", "10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valuesEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("valuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// --- tagsEqual ---

func TestTagsEqual(t *testing.T) {
	// Exact match
	tfAttrs := map[string]interface{}{
		"tags": map[string]interface{}{"Name": "my-vpc", "Env": "prod"},
	}
	awsTags := map[string]string{"Name": "my-vpc", "Env": "prod"}
	if !tagsEqual(tfAttrs, awsTags) {
		t.Error("expected tags to be equal")
	}

	// Different tags
	awsTagsDiff := map[string]string{"Name": "my-vpc", "Env": "staging"}
	if tagsEqual(tfAttrs, awsTagsDiff) {
		t.Error("expected tags to differ")
	}

	// AWS-managed tags should be ignored
	awsTagsManaged := map[string]string{"Name": "my-vpc", "Env": "prod", "aws:cloudformation:stack": "foo"}
	if !tagsEqual(tfAttrs, awsTagsManaged) {
		t.Error("aws: prefixed tags should be ignored")
	}

	// k8s tags should be ignored
	awsTagsK8s := map[string]string{"Name": "my-vpc", "Env": "prod", "kubernetes.io/cluster/foo": "owned"}
	if !tagsEqual(tfAttrs, awsTagsK8s) {
		t.Error("kubernetes.io/ prefixed tags should be ignored")
	}
}

// --- getTerraformTags ---

func TestGetTerraformTags(t *testing.T) {
	// From "tags" field
	attrs := map[string]interface{}{
		"tags": map[string]interface{}{"Name": "test"},
	}
	tags := getTerraformTags(attrs)
	if tags["Name"] != "test" {
		t.Errorf("tags[Name] = %q, want test", tags["Name"])
	}

	// From "tags_all" field
	attrs2 := map[string]interface{}{
		"tags_all": map[string]interface{}{"Name": "test2"},
	}
	tags2 := getTerraformTags(attrs2)
	if tags2["Name"] != "test2" {
		t.Errorf("tags_all[Name] = %q, want test2", tags2["Name"])
	}

	// Empty
	tags3 := getTerraformTags(map[string]interface{}{})
	if len(tags3) != 0 {
		t.Errorf("expected empty tags, got %d", len(tags3))
	}
}
