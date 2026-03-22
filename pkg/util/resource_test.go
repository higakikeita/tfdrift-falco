package util

import "testing"

func TestExtractLastPathSegment(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "simple path",
			path:     "a/b/c",
			expected: "c",
		},
		{
			name:     "Azure resource ID",
			path:     "/subscriptions/abc-123/resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM",
			expected: "myVM",
		},
		{
			name:     "GCP resource name",
			path:     "projects/123/zones/us-central1-a/instances/vm-1",
			expected: "vm-1",
		},
		{
			name:     "trailing slash",
			path:     "a/b/c/",
			expected: "c",
		},
		{
			name:     "single segment",
			path:     "test",
			expected: "test",
		},
		{
			name:     "empty string",
			path:     "",
			expected: "",
		},
		{
			name:     "only slashes",
			path:     "///",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractLastPathSegment(tt.path)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractPathSegment(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		key      string
		expected string
	}{
		{
			name:     "resourcegroups",
			path:     "/subscriptions/abc-123/resourceGroups/myRG/providers/test",
			key:      "resourcegroups",
			expected: "myRG",
		},
		{
			name:     "subscriptions",
			path:     "/subscriptions/sub-123/resourceGroups/myRG",
			key:      "subscriptions",
			expected: "sub-123",
		},
		{
			name:     "zones",
			path:     "projects/123/zones/us-central1-a/instances/vm-1",
			key:      "zones",
			expected: "us-central1-a",
		},
		{
			name:     "key not found",
			path:     "/subscriptions/abc-123/resourceGroups/myRG",
			key:      "providers",
			expected: "",
		},
		{
			name:     "key at end of path",
			path:     "/subscriptions/abc-123/resourceGroups",
			key:      "resourcegroups",
			expected: "",
		},
		{
			name:     "case insensitive",
			path:     "/SUBSCRIPTIONS/abc-123/RESOURCEGROUPS/myRG",
			key:      "subscriptions",
			expected: "abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractPathSegment(tt.path, tt.key)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractResourceIDFromAttributes(t *testing.T) {
	tests := []struct {
		name     string
		attrs    map[string]interface{}
		expected string
	}{
		{
			name:     "id field",
			attrs:    map[string]interface{}{"id": "test-id"},
			expected: "test-id",
		},
		{
			name:     "arn field",
			attrs:    map[string]interface{}{"arn": "arn:aws:test"},
			expected: "arn:aws:test",
		},
		{
			name:     "name field",
			attrs:    map[string]interface{}{"name": "test-name"},
			expected: "test-name",
		},
		{
			name:     "self_link field",
			attrs:    map[string]interface{}{"self_link": "https://example.com/resource"},
			expected: "https://example.com/resource",
		},
		{
			name:     "id takes precedence",
			attrs:    map[string]interface{}{"id": "id-value", "arn": "arn-value"},
			expected: "id-value",
		},
		{
			name:     "no matching fields",
			attrs:    map[string]interface{}{"other": "value"},
			expected: "",
		},
		{
			name:     "empty attributes",
			attrs:    map[string]interface{}{},
			expected: "",
		},
		{
			name:     "non-string value",
			attrs:    map[string]interface{}{"id": 123},
			expected: "",
		},
		{
			name:     "empty string value",
			attrs:    map[string]interface{}{"id": ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractResourceIDFromAttributes(tt.attrs)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
