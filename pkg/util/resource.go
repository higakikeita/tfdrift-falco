package util

import "strings"

// ExtractLastPathSegment returns the last segment of a path split by "/".
// Useful for extracting resource names from cloud resource IDs/ARNs.
func ExtractLastPathSegment(path string) string {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			return parts[i]
		}
	}
	return ""
}

// ExtractPathSegment finds a path key (case-insensitive) and returns the next segment.
// Example: ExtractPathSegment("/subscriptions/abc-123/resourceGroups/myRG", "resourcegroups") returns "myRG"
func ExtractPathSegment(path, key string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.EqualFold(part, key) && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// ExtractResourceIDFromAttributes extracts a resource identifier from a map of attributes.
// Tries common keys in order: id, arn, name, self_link.
func ExtractResourceIDFromAttributes(attrs map[string]interface{}) string {
	for _, key := range []string{"id", "arn", "name", "self_link"} {
		if v, ok := attrs[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
