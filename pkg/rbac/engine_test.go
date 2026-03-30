package rbac

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.rolePermissions)
	assert.NotNil(t, engine.roleHierarchy)
}

func TestAssignRole(t *testing.T) {
	engine := NewEngine()
	userID := "user123"

	engine.AssignRole(userID, RoleAdmin)
	role, ok := engine.GetUserRole(userID)
	assert.True(t, ok)
	assert.Equal(t, RoleAdmin, role)
}

func TestGetUserRole(t *testing.T) {
	engine := NewEngine()

	// Test non-existent user
	_, ok := engine.GetUserRole("nonexistent")
	assert.False(t, ok)

	// Test existing user
	engine.AssignRole("user1", RoleViewer)
	role, ok := engine.GetUserRole("user1")
	assert.True(t, ok)
	assert.Equal(t, RoleViewer, role)
}

func TestRemoveUser(t *testing.T) {
	engine := NewEngine()
	userID := "user123"

	engine.AssignRole(userID, RoleEditor)
	_, ok := engine.GetUserRole(userID)
	assert.True(t, ok)

	engine.RemoveUser(userID)
	_, ok = engine.GetUserRole(userID)
	assert.False(t, ok)
}

func TestHasPermission(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		role       Role
		resource   Resource
		permission Permission
		expected   bool
	}{
		// Admin permissions
		{name: "Admin read drifts", role: RoleAdmin, resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Admin write drifts", role: RoleAdmin, resource: ResourceDrifts, permission: PermWrite, expected: true},
		{name: "Admin delete drifts", role: RoleAdmin, resource: ResourceDrifts, permission: PermDelete, expected: true},
		{name: "Admin manage users", role: RoleAdmin, resource: ResourceUsers, permission: PermManageUsers, expected: true},
		{name: "Admin admin on settings", role: RoleAdmin, resource: ResourceSettings, permission: PermAdmin, expected: true},

		// Editor permissions
		{name: "Editor read drifts", role: RoleEditor, resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Editor write drifts", role: RoleEditor, resource: ResourceDrifts, permission: PermWrite, expected: true},
		{name: "Editor cannot delete drifts", role: RoleEditor, resource: ResourceDrifts, permission: PermDelete, expected: false},
		{name: "Editor read settings", role: RoleEditor, resource: ResourceSettings, permission: PermRead, expected: true},
		{name: "Editor cannot write settings", role: RoleEditor, resource: ResourceSettings, permission: PermWrite, expected: false},
		{name: "Editor cannot manage users", role: RoleEditor, resource: ResourceUsers, permission: PermManageUsers, expected: false},
		{name: "Editor read providers", role: RoleEditor, resource: ResourceProviders, permission: PermRead, expected: true},
		{name: "Editor write providers", role: RoleEditor, resource: ResourceProviders, permission: PermWrite, expected: true},

		// Viewer permissions
		{name: "Viewer read drifts", role: RoleViewer, resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Viewer cannot write drifts", role: RoleViewer, resource: ResourceDrifts, permission: PermWrite, expected: false},
		{name: "Viewer cannot delete drifts", role: RoleViewer, resource: ResourceDrifts, permission: PermDelete, expected: false},
		{name: "Viewer cannot read settings", role: RoleViewer, resource: ResourceSettings, permission: PermRead, expected: false},
		{name: "Viewer read providers", role: RoleViewer, resource: ResourceProviders, permission: PermRead, expected: true},
		{name: "Viewer cannot write providers", role: RoleViewer, resource: ResourceProviders, permission: PermWrite, expected: false},

		// Edge cases
		{name: "Unknown role", role: Role("unknown"), resource: ResourceDrifts, permission: PermRead, expected: false},
		{name: "Unknown resource", role: RoleAdmin, resource: Resource("unknown"), permission: PermRead, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.HasPermission(tt.role, tt.resource, tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPermissions(t *testing.T) {
	engine := NewEngine()

	// Test admin permissions
	adminPerms := engine.GetPermissions(RoleAdmin)
	assert.NotNil(t, adminPerms)
	assert.Equal(t, 5, len(adminPerms)) // All 5 resources
	assert.Contains(t, adminPerms[ResourceDrifts], PermRead)
	assert.Contains(t, adminPerms[ResourceDrifts], PermWrite)
	assert.Contains(t, adminPerms[ResourceDrifts], PermDelete)

	// Test editor permissions
	editorPerms := engine.GetPermissions(RoleEditor)
	assert.NotNil(t, editorPerms)
	assert.Greater(t, len(editorPerms), 0)
	assert.Contains(t, editorPerms[ResourceDrifts], PermRead)
	assert.Contains(t, editorPerms[ResourceDrifts], PermWrite)
	assert.NotContains(t, editorPerms[ResourceDrifts], PermDelete)

	// Test viewer permissions
	viewerPerms := engine.GetPermissions(RoleViewer)
	assert.NotNil(t, viewerPerms)
	assert.Greater(t, len(viewerPerms), 0)
	assert.Contains(t, viewerPerms[ResourceDrifts], PermRead)
	assert.NotContains(t, viewerPerms[ResourceDrifts], PermWrite)

	// Test unknown role
	unknownPerms := engine.GetPermissions(Role("unknown"))
	assert.Empty(t, unknownPerms)
}

func TestIsRoleAtLeast(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name        string
		role        Role
		minimumRole Role
		expected    bool
	}{
		// Admin tests
		{name: "Admin >= Admin", role: RoleAdmin, minimumRole: RoleAdmin, expected: true},
		{name: "Admin >= Editor", role: RoleAdmin, minimumRole: RoleEditor, expected: true},
		{name: "Admin >= Viewer", role: RoleAdmin, minimumRole: RoleViewer, expected: true},

		// Editor tests
		{name: "Editor >= Editor", role: RoleEditor, minimumRole: RoleEditor, expected: true},
		{name: "Editor >= Viewer", role: RoleEditor, minimumRole: RoleViewer, expected: true},
		{name: "Editor >= Admin", role: RoleEditor, minimumRole: RoleAdmin, expected: false},

		// Viewer tests
		{name: "Viewer >= Viewer", role: RoleViewer, minimumRole: RoleViewer, expected: true},
		{name: "Viewer >= Editor", role: RoleViewer, minimumRole: RoleEditor, expected: false},
		{name: "Viewer >= Admin", role: RoleViewer, minimumRole: RoleAdmin, expected: false},

		// Edge cases
		{name: "Unknown role", role: Role("unknown"), minimumRole: RoleViewer, expected: false},
		{name: "Unknown minimum role", role: RoleAdmin, minimumRole: Role("unknown"), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.IsRoleAtLeast(tt.role, tt.minimumRole)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUserCanAccess(t *testing.T) {
	engine := NewEngine()

	// Setup users
	engine.AssignRole("admin_user", RoleAdmin)
	engine.AssignRole("editor_user", RoleEditor)
	engine.AssignRole("viewer_user", RoleViewer)

	tests := []struct {
		name       string
		userID     string
		resource   Resource
		permission Permission
		expected   bool
	}{
		// Admin user
		{name: "Admin can read drifts", userID: "admin_user", resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Admin can delete drifts", userID: "admin_user", resource: ResourceDrifts, permission: PermDelete, expected: true},
		{name: "Admin can manage users", userID: "admin_user", resource: ResourceUsers, permission: PermManageUsers, expected: true},

		// Editor user
		{name: "Editor can read drifts", userID: "editor_user", resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Editor can write drifts", userID: "editor_user", resource: ResourceDrifts, permission: PermWrite, expected: true},
		{name: "Editor cannot delete drifts", userID: "editor_user", resource: ResourceDrifts, permission: PermDelete, expected: false},
		{name: "Editor cannot manage users", userID: "editor_user", resource: ResourceUsers, permission: PermManageUsers, expected: false},

		// Viewer user
		{name: "Viewer can read drifts", userID: "viewer_user", resource: ResourceDrifts, permission: PermRead, expected: true},
		{name: "Viewer cannot write drifts", userID: "viewer_user", resource: ResourceDrifts, permission: PermWrite, expected: false},
		{name: "Viewer cannot delete drifts", userID: "viewer_user", resource: ResourceDrifts, permission: PermDelete, expected: false},

		// Non-existent user
		{name: "Non-existent user", userID: "nonexistent", resource: ResourceDrifts, permission: PermRead, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.UserCanAccess(tt.userID, tt.resource, tt.permission)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRoleHierarchyLevel(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		role      Role
		expected  int
		shouldErr bool
	}{
		{name: "Admin level", role: RoleAdmin, expected: 2, shouldErr: false},
		{name: "Editor level", role: RoleEditor, expected: 1, shouldErr: false},
		{name: "Viewer level", role: RoleViewer, expected: 0, shouldErr: false},
		{name: "Unknown role", role: Role("unknown"), expected: 0, shouldErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level, ok := engine.GetRoleHierarchyLevel(tt.role)
			if tt.shouldErr {
				assert.False(t, ok)
			} else {
				assert.True(t, ok)
				assert.Equal(t, tt.expected, level)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	engine := NewEngine()

	// Assign roles concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			engine.AssignRole("user"+string(rune(id)), RoleAdmin)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all assignments
	for i := 0; i < 10; i++ {
		role, ok := engine.GetUserRole("user" + string(rune(i)))
		require.True(t, ok)
		assert.Equal(t, RoleAdmin, role)
	}
}

func TestRolePermissionsCompleteness(t *testing.T) {
	// Verify that all resources have some definition for each role
	for role := range RoleHierarchy {
		rolePerms, ok := RolePermissions[role]
		assert.True(t, ok, "Role %s missing from RolePermissions", role)

		// Admin should have permissions on all resources
		if role == RoleAdmin {
			for resource := range map[Resource]bool{
				ResourceDrifts:    true,
				ResourceEvents:    true,
				ResourceSettings:  true,
				ResourceUsers:     true,
				ResourceProviders: true,
			} {
				assert.Contains(t, rolePerms, resource, "Admin missing resource %s", resource)
			}
		}
	}
}
