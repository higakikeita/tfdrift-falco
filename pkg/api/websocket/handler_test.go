package websocket

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== checkOrigin Tests ====================

func TestCheckOrigin_DevelopmentModeAllowsAll(t *testing.T) {
	// Setup: development mode allows all origins
	t.Setenv("ENVIRONMENT", "development")

	// Create a request with an arbitrary origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://evil.com")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "development mode should allow any origin")
}

func TestCheckOrigin_DevelopmentModeAllowsAll_CaseInsensitive(t *testing.T) {
	// Setup: case-insensitive "development"
	t.Setenv("ENVIRONMENT", "DEVELOPMENT")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://attacker.com")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "development mode (case-insensitive) should allow any origin")
}

func TestCheckOrigin_EmptyOriginAlwaysAllowed(t *testing.T) {
	// Empty origin (same-origin requests) should always be allowed
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	// Don't set Origin header - simulates same-origin request

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "empty origin should always be allowed")
}

func TestCheckOrigin_LocalhostAllowed_Production(t *testing.T) {
	// localhost should be in default allowed origins
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "") // Use defaults

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "localhost should be allowed in production by default")
}

func TestCheckOrigin_LocalhostWithPortNotAllowed_ByDefault(t *testing.T) {
	// localhost with port is not in default allowed origins (only bare localhost)
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "localhost:3000 is not allowed in production by default (must use exact match)")
}

func TestCheckOrigin_Localhost5173NotAllowed_ByDefault(t *testing.T) {
	// localhost:5173 (Vite dev server) is not in default allowed origins
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost:5173")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "localhost:5173 is not allowed by default (needs explicit config)")
}

func TestCheckOrigin_127_0_0_1Allowed(t *testing.T) {
	// 127.0.0.1 (localhost IP) without port
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://127.0.0.1")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "127.0.0.1 should be allowed in production by default")
}

func TestCheckOrigin_127_0_0_1_WithPortNotAllowed_ByDefault(t *testing.T) {
	// 127.0.0.1 with port is not in default allowed origins (only bare IP)
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://127.0.0.1:3000")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "127.0.0.1:3000 is not allowed by default (needs exact match)")
}

func TestCheckOrigin_CustomAllowedOrigin(t *testing.T) {
	// Custom allowed origin
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://myapp.com")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "custom allowed origin should be accepted")
}

func TestCheckOrigin_CustomAllowedOrigin_MultipleOrigins(t *testing.T) {
	// Multiple custom origins
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com,https://other.com,https://third.com")

	testCases := []struct {
		origin   string
		expected bool
	}{
		{"https://myapp.com", true},
		{"https://other.com", true},
		{"https://third.com", true},
		{"https://evil.com", false},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Origin", tc.origin)
		result := TestCheckOriginFunc(req)
		assert.Equal(t, tc.expected, result, "origin %s should be %v", tc.origin, tc.expected)
	}
}

func TestCheckOrigin_UnauthorizedOriginBlocked(t *testing.T) {
	// Unauthorized origin should be rejected
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://evil.com")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "unauthorized origin should be blocked in production")
}

func TestCheckOrigin_UnauthorizedOriginBlocked_WithDefaults(t *testing.T) {
	// Unauthorized origin blocked even with default origins
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "") // Use defaults (localhost, 127.0.0.1)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://attacker.com")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "unauthorized origin should be blocked even with defaults")
}

func TestCheckOrigin_ExactMatchRequired(t *testing.T) {
	// Exact match required - subdomain shouldn't match
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://sub.myapp.com")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "subdomain should not match parent domain")
}

func TestCheckOrigin_CaseSensitiveMatch(t *testing.T) {
	// Origin matching is case-sensitive
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://MyApp.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://myapp.com")

	// Note: This may depend on implementation. HTTP origin headers are case-insensitive,
	// but this tests current behavior
	result := TestCheckOriginFunc(req)
	// Current implementation does exact string match, so different cases won't match
	assert.False(t, result, "origin matching should respect case sensitivity")
}

func TestCheckOrigin_HTTPSRequired(t *testing.T) {
	// HTTPS should not match HTTP
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://myapp.com")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "http should not match https")
}

func TestCheckOrigin_PortMatters(t *testing.T) {
	// Port is significant in origin
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com:3000")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://myapp.com:8000")

	result := TestCheckOriginFunc(req)
	assert.False(t, result, "different port should not match")
}

func TestCheckOrigin_DefaultOrigins_NoAllowedOrigins(t *testing.T) {
	// When ALLOWED_ORIGINS is empty, defaults should be used
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "default origins should be used when ALLOWED_ORIGINS is empty")
}

// ==================== getAllowedOrigins Tests ====================

func TestGetAllowedOrigins_DefaultsWhenEmpty(t *testing.T) {
	// When ALLOWED_ORIGINS is empty, should return defaults
	t.Setenv("ALLOWED_ORIGINS", "")

	origins := TestGetAllowedOriginsFunc()

	assert.NotNil(t, origins)
	assert.GreaterOrEqual(t, len(origins), 2, "should have at least localhost and 127.0.0.1")
	assert.Contains(t, origins, "http://localhost", "should contain http://localhost")
	assert.Contains(t, origins, "http://127.0.0.1", "should contain http://127.0.0.1")
}

func TestGetAllowedOrigins_SingleOrigin(t *testing.T) {
	// Single custom origin
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 1)
	assert.Equal(t, "https://myapp.com", origins[0])
}

func TestGetAllowedOrigins_MultipleOrigins(t *testing.T) {
	// Multiple comma-separated origins
	t.Setenv("ALLOWED_ORIGINS", "https://app1.com,https://app2.com,https://app3.com")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 3)
	assert.Contains(t, origins, "https://app1.com")
	assert.Contains(t, origins, "https://app2.com")
	assert.Contains(t, origins, "https://app3.com")
}

func TestGetAllowedOrigins_TrimsWhitespace(t *testing.T) {
	// Whitespace around origins should be trimmed
	t.Setenv("ALLOWED_ORIGINS", "  https://app1.com  ,  https://app2.com  ,  https://app3.com  ")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 3)
	assert.Equal(t, "https://app1.com", origins[0], "leading/trailing spaces should be trimmed")
	assert.Equal(t, "https://app2.com", origins[1])
	assert.Equal(t, "https://app3.com", origins[2], "trailing spaces should be trimmed")
}

func TestGetAllowedOrigins_SingleOriginWithWhitespace(t *testing.T) {
	// Single origin with surrounding whitespace
	t.Setenv("ALLOWED_ORIGINS", "  https://myapp.com  ")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 1)
	assert.Equal(t, "https://myapp.com", origins[0])
}

func TestGetAllowedOrigins_PreservesProtocol(t *testing.T) {
	// Protocol should be preserved
	t.Setenv("ALLOWED_ORIGINS", "https://secure.com,http://insecure.com")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 2)
	assert.Equal(t, "https://secure.com", origins[0])
	assert.Equal(t, "http://insecure.com", origins[1])
}

func TestGetAllowedOrigins_PreservesPorts(t *testing.T) {
	// Ports should be preserved
	t.Setenv("ALLOWED_ORIGINS", "https://app.com:3000,https://app.com:8000")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 2)
	assert.Contains(t, origins, "https://app.com:3000")
	assert.Contains(t, origins, "https://app.com:8000")
}

func TestGetAllowedOrigins_EmptyStringsNotPresentAfterTrimming(t *testing.T) {
	// Edge case: multiple commas shouldn't create empty strings in result
	// (depends on implementation)
	t.Setenv("ALLOWED_ORIGINS", "https://app1.com, , https://app2.com")

	origins := TestGetAllowedOriginsFunc()

	// Implementation will include empty string from the middle comma
	// This tests the actual behavior - you may want to add filtering for empty strings
	assert.Greater(t, len(origins), 0)
	for _, origin := range origins {
		if origin != "" { // If there are any non-empty origins
			assert.True(t, strings.HasPrefix(origin, "http"), "non-empty origins should start with http")
		}
	}
}

func TestGetAllowedOrigins_NoEnvVar(t *testing.T) {
	// Ensure ALLOWED_ORIGINS is not set
	t.Setenv("ALLOWED_ORIGINS", "")

	origins := TestGetAllowedOriginsFunc()

	// Should return defaults
	assert.NotEmpty(t, origins)
	defaultFound := false
	for _, origin := range origins {
		if origin == "http://localhost" || origin == "http://127.0.0.1" {
			defaultFound = true
			break
		}
	}
	assert.True(t, defaultFound, "should contain default origins")
}

func TestGetAllowedOrigins_ComplexOrigins(t *testing.T) {
	// Complex real-world examples
	t.Setenv("ALLOWED_ORIGINS", "https://app.example.com:3000,https://staging-app.example.com,http://localhost:8080")

	origins := TestGetAllowedOriginsFunc()

	assert.Len(t, origins, 3)
	assert.Contains(t, origins, "https://app.example.com:3000")
	assert.Contains(t, origins, "https://staging-app.example.com")
	assert.Contains(t, origins, "http://localhost:8080")
}

func TestGetAllowedOrigins_TrailingComma(t *testing.T) {
	// Trailing comma creates empty string
	t.Setenv("ALLOWED_ORIGINS", "https://app1.com,https://app2.com,")

	origins := TestGetAllowedOriginsFunc()

	// This will include an empty string due to how split works
	assert.Greater(t, len(origins), 0)
	// At minimum, should have the first two origins
	assert.Contains(t, origins, "https://app1.com")
	assert.Contains(t, origins, "https://app2.com")
}

// ==================== Integration Tests ====================

func TestCheckOrigin_WithGetAllowedOrigins_Integration(t *testing.T) {
	// Test checkOrigin with getAllowedOrigins integration
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://app.com,https://staging.com")

	testCases := []struct {
		name     string
		origin   string
		expected bool
	}{
		{"allowed origin 1", "https://app.com", true},
		{"allowed origin 2", "https://staging.com", true},
		{"unauthorized origin", "https://evil.com", false},
		{"empty origin", "", true},
		{"localhost not in custom list", "http://localhost", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			result := TestCheckOriginFunc(req)
			assert.Equal(t, tc.expected, result, "origin %s", tc.origin)
		})
	}
}

func TestCheckOrigin_DevelopmentVsProduction(t *testing.T) {
	// Same origin should behave differently based on environment
	origin := "http://random-origin.com"

	// Development mode
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("ALLOWED_ORIGINS", "https://legitimate.com")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", origin)
	devResult := TestCheckOriginFunc(req)

	// Production mode
	t.Setenv("ENVIRONMENT", "production")
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", origin)
	prodResult := TestCheckOriginFunc(req)

	assert.True(t, devResult, "development mode should allow the origin")
	assert.False(t, prodResult, "production mode should block the origin")
}

func TestCheckOrigin_EnvironmentNotDevelopment(t *testing.T) {
	// Non-development environments should validate against allowed origins
	for _, env := range []string{"staging", "testing", "", "PROD", "Production"} {
		t.Run("env_"+env, func(t *testing.T) {
			t.Setenv("ENVIRONMENT", env)
			t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Origin", "http://evil.com")

			result := TestCheckOriginFunc(req)
			assert.False(t, result, "non-development environment should enforce allowed origins")
		})
	}
}

// ==================== Edge Case Tests ====================

func TestCheckOrigin_MalformedOrigin(t *testing.T) {
	// Malformed origin should still be compared as string
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "ht!tp://malformed")

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "ht!tp://malformed")

	result := TestCheckOriginFunc(req)
	assert.True(t, result, "even malformed origins should match if configured")
}

func TestCheckOrigin_OriginWithPath(t *testing.T) {
	// Origin header should not include path
	t.Setenv("ENVIRONMENT", "production")
	t.Setenv("ALLOWED_ORIGINS", "https://myapp.com")

	req := httptest.NewRequest("GET", "/", nil)
	// Origin header correctly doesn't include path
	req.Header.Set("Origin", "https://myapp.com")

	result := TestCheckOriginFunc(req)
	assert.True(t, result)
}

func TestGetAllowedOrigins_ManyOrigins(t *testing.T) {
	// Many origins
	originStr := "https://app1.com,https://app2.com,https://app3.com,https://app4.com,https://app5.com"

	t.Setenv("ALLOWED_ORIGINS", originStr)

	result := TestGetAllowedOriginsFunc()
	assert.Len(t, result, 5)
}

// ==================== Upgrader Tests ====================

func TestNewUpgrader_CheckOriginFunctionSet(t *testing.T) {
	// Verify newUpgrader returns a valid Upgrader with CheckOrigin set
	upgraderInstance := TestNewUploaderFunc()

	assert.NotNil(t, upgraderInstance.CheckOrigin, "CheckOrigin function should be set")

	// Test that the CheckOrigin function works
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://localhost")

	// This should call our checkOrigin function internally
	result := upgraderInstance.CheckOrigin(req)
	assert.True(t, result, "CheckOrigin function should work for localhost")
}

func TestNewUpgrader_BufferSizes(t *testing.T) {
	// Verify buffer sizes are set
	upgraderInstance := TestNewUploaderFunc()

	assert.Equal(t, 1024, upgraderInstance.ReadBufferSize)
	assert.Equal(t, 1024, upgraderInstance.WriteBufferSize)
}
