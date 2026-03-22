package util

import "testing"

func TestGetStringField(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]string
		key      string
		expected string
	}{
		{
			name:     "exact match",
			fields:   map[string]string{"test": "value"},
			key:      "test",
			expected: "value",
		},
		{
			name:     "case insensitive match",
			fields:   map[string]string{"Test": "value"},
			key:      "test",
			expected: "value",
		},
		{
			name:     "key not found",
			fields:   map[string]string{"other": "value"},
			key:      "test",
			expected: "",
		},
		{
			name:     "empty map",
			fields:   map[string]string{},
			key:      "test",
			expected: "",
		},
		{
			name:     "exact match takes precedence",
			fields:   map[string]string{"test": "exact", "TEST": "fallback"},
			key:      "test",
			expected: "exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStringField(tt.fields, tt.key)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetInterfaceStringField(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "string value",
			m:        map[string]interface{}{"test": "value"},
			key:      "test",
			expected: "value",
		},
		{
			name:     "non-string value",
			m:        map[string]interface{}{"test": 123},
			key:      "test",
			expected: "",
		},
		{
			name:     "key not found",
			m:        map[string]interface{}{"other": "value"},
			key:      "test",
			expected: "",
		},
		{
			name:     "empty map",
			m:        map[string]interface{}{},
			key:      "test",
			expected: "",
		},
		{
			name:     "nil value",
			m:        map[string]interface{}{"test": nil},
			key:      "test",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetInterfaceStringField(tt.m, tt.key)
			if result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
	}
}
