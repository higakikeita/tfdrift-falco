package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key-for-jwt-signing-min32chars!"

func newTestAuth(enabled bool) *Auth {
	return NewAuth(AuthConfig{
		Enabled:   enabled,
		JWTSecret: testSecret,
		JWTIssuer: "tfdrift-falco-test",
		JWTExpiry: "1h",
		APIKeys: []APIKeyEntry{
			{
				Name:      "test-key",
				Key:       "tfd_testapikey1234567890abcdef",
				Scopes:    []string{"read", "write"},
				CreatedAt: "2026-03-22T00:00:00Z",
			},
		},
	})
}

func TestAuthMiddleware_Disabled(t *testing.T) {
	auth := newTestAuth(false)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "ok", rr.Body.String())
}

func TestAuthMiddleware_NoCredentials(t *testing.T) {
	auth := newTestAuth(true)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Header().Get("WWW-Authenticate"), "Bearer")

	var resp models.APIResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotNil(t, resp.Error)
	assert.Equal(t, 401, resp.Error.Code)
}

func TestAuthMiddleware_ValidJWT(t *testing.T) {
	auth := newTestAuth(true)

	// Generate a valid token
	token, err := auth.GenerateToken("test-user")
	require.NoError(t, err)

	var capturedInfo *AuthInfo
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInfo = GetAuthInfo(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedInfo)
	assert.Equal(t, "jwt", capturedInfo.Method)
	assert.Equal(t, "test-user", capturedInfo.Subject)
}

func TestAuthMiddleware_ExpiredJWT(t *testing.T) {
	auth := newTestAuth(true)

	// Create an expired token
	claims := jwt.MapClaims{
		"sub": "test-user",
		"iss": "tfdrift-falco-test",
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
		"exp": time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_InvalidJWTSignature(t *testing.T) {
	auth := newTestAuth(true)

	// Create token with wrong secret
	claims := jwt.MapClaims{
		"sub": "test-user",
		"iss": "tfdrift-falco-test",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("wrong-secret-key-that-is-long-enough"))
	require.NoError(t, err)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_WrongIssuer(t *testing.T) {
	auth := newTestAuth(true)

	claims := jwt.MapClaims{
		"sub": "test-user",
		"iss": "wrong-issuer",
		"exp": time.Now().Add(1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(testSecret))
	require.NoError(t, err)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestAuthMiddleware_ValidAPIKey(t *testing.T) {
	auth := newTestAuth(true)

	var capturedInfo *AuthInfo
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInfo = GetAuthInfo(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-API-Key", "tfd_testapikey1234567890abcdef")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedInfo)
	assert.Equal(t, "api_key", capturedInfo.Method)
	assert.Equal(t, "test-key", capturedInfo.Subject)
	assert.Equal(t, []string{"read", "write"}, capturedInfo.Scopes)
}

func TestAuthMiddleware_InvalidAPIKey(t *testing.T) {
	auth := newTestAuth(true)

	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("X-API-Key", "tfd_invalidkey")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestGenerateToken(t *testing.T) {
	auth := newTestAuth(true)

	token, err := auth.GenerateToken("admin")
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token is valid
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testSecret), nil
	})
	require.NoError(t, err)
	assert.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)

	sub, err := claims.GetSubject()
	require.NoError(t, err)
	assert.Equal(t, "admin", sub)

	iss, err := claims.GetIssuer()
	require.NoError(t, err)
	assert.Equal(t, "tfdrift-falco-test", iss)
}

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	require.NoError(t, err)
	assert.True(t, len(key) > 10)
	assert.True(t, key[:4] == "tfd_", "API key should start with tfd_ prefix")

	// Generate two keys and verify they are different
	key2, err := GenerateAPIKey()
	require.NoError(t, err)
	assert.NotEqual(t, key, key2)
}

func TestAddAndRemoveAPIKey(t *testing.T) {
	auth := newTestAuth(true)

	// Add a new key
	auth.AddAPIKey(APIKeyEntry{
		Name:      "new-key",
		Key:       "tfd_newkey1234567890",
		Scopes:    []string{"read"},
		CreatedAt: "2026-03-22T00:00:00Z",
	})

	// Verify the key works
	info, err := auth.validateAPIKey("tfd_newkey1234567890")
	require.NoError(t, err)
	assert.Equal(t, "new-key", info.Subject)

	// List keys
	keys := auth.ListAPIKeys()
	assert.Equal(t, 2, len(keys)) // original + new

	// Remove the key
	removed := auth.RemoveAPIKey("new-key")
	assert.True(t, removed)

	// Verify it no longer works
	_, err = auth.validateAPIKey("tfd_newkey1234567890")
	assert.Error(t, err)

	// Remove non-existent key
	removed = auth.RemoveAPIKey("non-existent")
	assert.False(t, removed)
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"tfd_a1b2c3d4e5f6g7h8", "tfd_a1b2...g7h8"},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "12345678...6789"},
	}

	for _, tc := range tests {
		result := maskKey(tc.input)
		assert.Equal(t, tc.expected, result, "maskKey(%q)", tc.input)
	}
}

func TestJWTPreferredOverAPIKey(t *testing.T) {
	auth := newTestAuth(true)

	token, err := auth.GenerateToken("jwt-user")
	require.NoError(t, err)

	var capturedInfo *AuthInfo
	handler := auth.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedInfo = GetAuthInfo(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	// Send both JWT and API Key - JWT should take precedence
	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-API-Key", "tfd_testapikey1234567890abcdef")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	require.NotNil(t, capturedInfo)
	assert.Equal(t, "jwt", capturedInfo.Method)
	assert.Equal(t, "jwt-user", capturedInfo.Subject)
}
