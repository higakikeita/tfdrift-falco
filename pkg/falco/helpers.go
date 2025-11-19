package falco

import (
	"strings"
)

// getStringField safely gets a string field from the map
func getStringField(fields map[string]string, key string) string {
	// Try direct lookup
	if val, ok := fields[key]; ok {
		return val
	}

	// Try case-insensitive lookup (Falco might use different casing)
	lowerKey := strings.ToLower(key)
	for k, v := range fields {
		if strings.ToLower(k) == lowerKey {
			return v
		}
	}

	return ""
}
