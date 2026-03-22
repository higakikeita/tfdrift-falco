// Package util provides shared utility functions used across TFDrift-Falco packages.
package util

import "strings"

// GetStringField retrieves a string value from a map with case-insensitive key lookup.
// Returns empty string if key not found.
func GetStringField(fields map[string]string, key string) string {
	// Try exact match first (fast path)
	if v, ok := fields[key]; ok {
		return v
	}
	// Fall back to case-insensitive lookup
	for k, v := range fields {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

// GetInterfaceStringField retrieves a string value from a map[string]interface{}.
// Performs type assertion and returns empty string if key not found or value is not a string.
func GetInterfaceStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
