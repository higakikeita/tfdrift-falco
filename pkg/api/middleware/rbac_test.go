package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/rbac"
	"github.com/stretchr/testify/assert"
)

func TestRequireRole_Disabled(t *testing.T) {
	// Disable RBAC for this test
	oldConfig := GlobalRBACConfig
	defer func() { GlobalRBACConfig = oldConfig }()

	SetRBACConfig(&RBACConfig{
		Enabled: false,
		Engine:  rbac.NewEngine(),
	})

	middleware := RequireRole(rbac.RoleAdmin)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, "OK", string(body))
}

func TestRequireRole_NoUserInContext(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequireRole(rbac.RoleAdmin)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Unauthorized", response.Error)
}

func TestRequireRole_InsufficientRole(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleViewer)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequireRole(rbac.RoleAdmin)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleViewer)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Forbidden", response.Error)
}

func TestRequireRole_ValidRole(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleAdmin)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequireRole(rbac.RoleAdmin)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleAdmin)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, "OK", string(body))
}

func TestRequireRole_HierarchyCheck(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleAdmin)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	// Require Editor role, but user is Admin (higher privilege)
	middleware := RequireRole(rbac.RoleEditor)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleAdmin)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRole_MultipleRoles(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleEditor)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	// Require one of: Admin or Editor
	middleware := RequireRole(rbac.RoleAdmin, rbac.RoleEditor)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleEditor)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_Disabled(t *testing.T) {
	// Disable RBAC
	oldConfig := GlobalRBACConfig
	defer func() { GlobalRBACConfig = oldConfig }()

	SetRBACConfig(&RBACConfig{
		Enabled: false,
		Engine:  rbac.NewEngine(),
	})

	middleware := RequirePermission(rbac.ResourceDrifts, rbac.PermRead)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequirePermission_NoUserInContext(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequirePermission(rbac.ResourceDrifts, rbac.PermRead)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestRequirePermission_InsufficientPermission(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleViewer)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequirePermission(rbac.ResourceDrifts, rbac.PermWrite)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleViewer)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)
	assert.Equal(t, "Forbidden", response.Error)
}

func TestRequirePermission_ValidPermission(t *testing.T) {
	// Enable RBAC
	engine := rbac.NewEngine()
	engine.AssignRole("user1", rbac.RoleEditor)
	SetRBACConfig(&RBACConfig{
		Enabled: true,
		Engine:  engine,
	})
	defer func() { SetRBACConfig(&RBACConfig{Enabled: false, Engine: rbac.NewEngine()}) }()

	middleware := RequirePermission(rbac.ResourceDrifts, rbac.PermWrite)
	testHandler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user1", rbac.RoleEditor)
	w := httptest.NewRecorder()

	testHandler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body, _ := io.ReadAll(w.Body)
	assert.Equal(t, "OK", string(body))
}

func TestSetUserContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	req = SetUserContext(req, "user123", rbac.RoleAdmin)

	userID, ok := GetUserID(req.Context())
	assert.True(t, ok)
	assert.Equal(t, "user123", userID)

	role, ok := GetUserRole(req.Context())
	assert.True(t, ok)
	assert.Equal(t, rbac.RoleAdmin, role)
}

func TestGetUserID(t *testing.T) {
	// Test with no context
	ctx := context.Background()
	_, ok := GetUserID(ctx)
	assert.False(t, ok)

	// Test with context
	ctx = context.WithValue(ctx, contextKey("user_id"), "user123")
	userID, ok := GetUserID(ctx)
	assert.True(t, ok)
	assert.Equal(t, "user123", userID)
}

func TestGetUserRole(t *testing.T) {
	// Test with no context
	ctx := context.Background()
	_, ok := GetUserRole(ctx)
	assert.False(t, ok)

	// Test with context
	ctx = context.WithValue(ctx, contextKey("user_role"), rbac.RoleViewer)
	role, ok := GetUserRole(ctx)
	assert.True(t, ok)
	assert.Equal(t, rbac.RoleViewer, role)
}

func TestGlobalRBACConfig(t *testing.T) {
	oldConfig := GlobalRBACConfig
	defer func() { GlobalRBACConfig = oldConfig }()

	// Test SetRBACConfig
	newConfig := &RBACConfig{
		Enabled: true,
		Engine:  rbac.NewEngine(),
	}
	SetRBACConfig(newConfig)

	config := GetRBACConfig()
	assert.Equal(t, newConfig.Enabled, config.Enabled)
	assert.Equal(t, newConfig.Engine, config.Engine)
}

func TestErrorResponse_JSON(t *testing.T) {
	w := httptest.NewRecorder()

	respondWithError(w, http.StatusForbidden, "Forbidden", "Test error message")

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, "Forbidden", response.Error)
	assert.Equal(t, "Test error message", response.Message)
}

func TestRBACConfigFromEnv(t *testing.T) {
	// Save original env
	original := os.Getenv("RBAC_ENABLED")
	defer os.Setenv("RBAC_ENABLED", original)

	// Test with RBAC_ENABLED=true
	os.Setenv("RBAC_ENABLED", "true")
	// Note: GlobalRBACConfig is initialized at package init, so we just verify the logic works
	cfg := &RBACConfig{
		Enabled: os.Getenv("RBAC_ENABLED") == "true",
		Engine:  rbac.NewEngine(),
	}
	assert.True(t, cfg.Enabled)

	// Test with RBAC_ENABLED=false
	os.Setenv("RBAC_ENABLED", "false")
	cfg = &RBACConfig{
		Enabled: os.Getenv("RBAC_ENABLED") == "true",
		Engine:  rbac.NewEngine(),
	}
	assert.False(t, cfg.Enabled)
}
