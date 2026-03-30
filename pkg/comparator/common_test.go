package comparator

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]interface{}
		path     string
		expected interface{}
	}{
		{
			name: "simple key",
			data: map[string]interface{}{
				"key": "value",
			},
			path:     "key",
			expected: "value",
		},
		{
			name: "nested key with dot notation",
			data: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "nested_value",
				},
			},
			path:     "parent.child",
			expected: "nested_value",
		},
		{
			name: "deeply nested key",
			data: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"level3": "deep_value",
					},
				},
			},
			path:     "level1.level2.level3",
			expected: "deep_value",
		},
		{
			name: "missing key",
			data: map[string]interface{}{
				"key": "value",
			},
			path:     "missing",
			expected: nil,
		},
		{
			name: "missing nested key",
			data: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "value",
				},
			},
			path:     "parent.missing",
			expected: nil,
		},
		{
			name: "path breaks at non-map type",
			data: map[string]interface{}{
				"parent": "string_value",
			},
			path:     "parent.child",
			expected: nil,
		},
		{
			name: "empty data",
			data: map[string]interface{}{},
			path: "key",
			expected: nil,
		},
		{
			name: "numeric value",
			data: map[string]interface{}{
				"number": 42,
			},
			path:     "number",
			expected: 42,
		},
		{
			name: "boolean value",
			data: map[string]interface{}{
				"flag": true,
			},
			path:     "flag",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetNestedValue(tt.data, tt.path)
			if result != tt.expected {
				t.Errorf("GetNestedValue(%v, %s) = %v, want %v", tt.data, tt.path, result, tt.expected)
			}
		})
	}
}

func TestValuesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        nil,
			b:        "value",
			expected: false,
		},
		{
			name:     "equal strings",
			a:        "value",
			b:        "value",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "value1",
			b:        "value2",
			expected: false,
		},
		{
			name:     "equal numbers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "different numbers",
			a:        42,
			b:        43,
			expected: false,
		},
		{
			name:     "equal booleans",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "different booleans",
			a:        true,
			b:        false,
			expected: false,
		},
		{
			name:     "boolean and string true",
			a:        true,
			b:        "true",
			expected: true,
		},
		{
			name:     "boolean and string false",
			a:        false,
			b:        "false",
			expected: true,
		},
		{
			name:     "boolean and wrong string",
			a:        true,
			b:        "false",
			expected: false,
		},
		{
			name:     "same type numbers",
			a:        int64(42),
			b:        int64(42),
			expected: true,
		},
		{
			name:     "different type numbers (string format)",
			a:        42,
			b:        "42",
			expected: true,
		},
		{
			name:     "empty slices",
			a:        []string{},
			b:        []string{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValuesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("ValuesEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestValuesEqualCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		a        interface{}
		b        interface{}
		expected bool
	}{
		{
			name:     "equal strings",
			a:        "value",
			b:        "value",
			expected: true,
		},
		{
			name:     "case-insensitive strings",
			a:        "eastUS",
			b:        "EastUs",
			expected: true,
		},
		{
			name:     "different strings",
			a:        "value1",
			b:        "value2",
			expected: false,
		},
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "one nil",
			a:        nil,
			b:        "value",
			expected: false,
		},
		{
			name:     "equal numbers",
			a:        42,
			b:        42,
			expected: true,
		},
		{
			name:     "equal booleans",
			a:        true,
			b:        true,
			expected: true,
		},
		{
			name:     "boolean and string true",
			a:        true,
			b:        "true",
			expected: true,
		},
		{
			name:     "mixed type case-insensitive",
			a:        "EastUS",
			b:        "eastus",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValuesEqualCaseInsensitive(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("ValuesEqualCaseInsensitive(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestExtractMapStringValues(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		expected map[string]string
	}{
		{
			name: "valid map",
			data: map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
			expected: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name: "mixed value types, only strings extracted",
			data: map[string]interface{}{
				"string": "value",
				"number": 42,
				"bool":   true,
			},
			expected: map[string]string{
				"string": "value",
			},
		},
		{
			name:     "nil data",
			data:     nil,
			expected: map[string]string{},
		},
		{
			name:     "not a map",
			data:     "string",
			expected: map[string]string{},
		},
		{
			name:     "empty map",
			data:     map[string]interface{}{},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractMapStringValues(tt.data)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("ExtractMapStringValues(%v) = %v, want %v", tt.data, result, tt.expected)
			}
		})
	}
}

func TestFilterManagedLabels(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		prefixes []string
		expected map[string]string
	}{
		{
			name: "filter AWS managed tags",
			labels: map[string]string{
				"user-tag":            "value1",
				"aws:cloudformation": "stack-id",
				"custom":              "value2",
				"kubernetes.io/name": "my-service",
			},
			prefixes: []string{"aws:", "kubernetes.io/"},
			expected: map[string]string{
				"user-tag": "value1",
				"custom":   "value2",
			},
		},
		{
			name: "no filtering needed",
			labels: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
			prefixes: []string{"aws:", "kubernetes.io/"},
			expected: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
		},
		{
			name:     "empty labels",
			labels:   map[string]string{},
			prefixes: []string{"aws:"},
			expected: map[string]string{},
		},
		{
			name: "filter all labels",
			labels: map[string]string{
				"aws:tag1": "value1",
				"aws:tag2": "value2",
			},
			prefixes: []string{"aws:"},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterManagedLabels(tt.labels, tt.prefixes)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("FilterManagedLabels(%v, %v) = %v, want %v", tt.labels, tt.prefixes, result, tt.expected)
			}
		})
	}
}

func TestFilterManagedLabelsCaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		prefixes []string
		expected map[string]string
	}{
		{
			name: "filter Azure managed tags case-insensitive",
			labels: map[string]string{
				"user-tag":          "value1",
				"hidden-something":  "hidden-value",
				"custom":            "value2",
				"HIDDEN-OTHER":      "another-hidden",
				"ms-resource-usage": "usage-data",
			},
			prefixes: []string{"hidden-", "ms-resource-usage"},
			expected: map[string]string{
				"user-tag": "value1",
				"custom":   "value2",
			},
		},
		{
			name: "no filtering needed",
			labels: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
			prefixes: []string{"hidden-"},
			expected: map[string]string{
				"tag1": "value1",
				"tag2": "value2",
			},
		},
		{
			name:     "empty labels",
			labels:   map[string]string{},
			prefixes: []string{"hidden-"},
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterManagedLabelsCaseInsensitive(tt.labels, tt.prefixes)
			if !mapsEqual(result, tt.expected) {
				t.Errorf("FilterManagedLabelsCaseInsensitive(%v, %v) = %v, want %v", tt.labels, tt.prefixes, result, tt.expected)
			}
		})
	}
}

// Helper function to compare two string maps
func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || bv != v {
			return false
		}
	}
	return true
}

func TestCompareResources(t *testing.T) {
	tests := []struct {
		name              string
		config            *ComparisonConfig
		tfResources       []interface{}
		cloudResources    []interface{}
		expectedUnmanaged int
		expectedMissing   int
		expectedModified  int
	}{
		{
			name: "all resources match",
			config: &ComparisonConfig{
				ExtractTFID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				ExtractCloudID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				CompareAttributes: func(tf, cloud interface{}) []types.FieldDiff {
					return nil // No differences
				},
				BuildUnmanaged: func(r interface{}) *types.ResourceDiff {
					return nil
				},
				BuildMissing: func(r interface{}) *types.TerraformResource {
					return nil
				},
			},
			tfResources: []interface{}{
				map[string]string{"id": "res1"},
				map[string]string{"id": "res2"},
			},
			cloudResources: []interface{}{
				map[string]string{"id": "res1"},
				map[string]string{"id": "res2"},
			},
			expectedUnmanaged: 0,
			expectedMissing:   0,
			expectedModified:  0,
		},
		{
			name: "unmanaged resource in cloud",
			config: &ComparisonConfig{
				ExtractTFID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				ExtractCloudID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				CompareAttributes: func(tf, cloud interface{}) []types.FieldDiff {
					return nil
				},
				BuildUnmanaged: func(r interface{}) *types.ResourceDiff {
					m := r.(map[string]string)
					return &types.ResourceDiff{
						ResourceID: m["id"],
					}
				},
				BuildMissing: func(r interface{}) *types.TerraformResource {
					return nil
				},
			},
			tfResources: []interface{}{
				map[string]string{"id": "res1"},
			},
			cloudResources: []interface{}{
				map[string]string{"id": "res1"},
				map[string]string{"id": "res2"}, // Unmanaged
			},
			expectedUnmanaged: 1,
			expectedMissing:   0,
			expectedModified:  0,
		},
		{
			name: "missing resource in cloud",
			config: &ComparisonConfig{
				ExtractTFID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				ExtractCloudID: func(r interface{}) string {
					return r.(map[string]string)["id"]
				},
				CompareAttributes: func(tf, cloud interface{}) []types.FieldDiff {
					return nil
				},
				BuildUnmanaged: func(r interface{}) *types.ResourceDiff {
					return nil
				},
				BuildMissing: func(r interface{}) *types.TerraformResource {
					m := r.(map[string]string)
					return &types.TerraformResource{
						ID: m["id"],
					}
				},
			},
			tfResources: []interface{}{
				map[string]string{"id": "res1"},
				map[string]string{"id": "res2"}, // Missing
			},
			cloudResources: []interface{}{
				map[string]string{"id": "res1"},
			},
			expectedUnmanaged: 0,
			expectedMissing:   1,
			expectedModified:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CompareResources(tt.config, tt.tfResources, tt.cloudResources)
			if len(result.UnmanagedResources) != tt.expectedUnmanaged {
				t.Errorf("unmanaged resources: got %d, want %d", len(result.UnmanagedResources), tt.expectedUnmanaged)
			}
			if len(result.MissingResources) != tt.expectedMissing {
				t.Errorf("missing resources: got %d, want %d", len(result.MissingResources), tt.expectedMissing)
			}
			if len(result.ModifiedResources) != tt.expectedModified {
				t.Errorf("modified resources: got %d, want %d", len(result.ModifiedResources), tt.expectedModified)
			}
		})
	}
}
