package rbac

import (
	"sync"
)

// Engine provides role-based access control functionality
type Engine struct {
	mu              sync.RWMutex
	userRoles       map[string]Role // maps user ID to role
	rolePermissions map[Role]map[Resource][]Permission
	roleHierarchy   map[Role]int
}

// NewEngine creates a new RBAC engine
func NewEngine() *Engine {
	return &Engine{
		userRoles:       make(map[string]Role),
		rolePermissions: RolePermissions,
		roleHierarchy:   RoleHierarchy,
	}
}

// AssignRole assigns a role to a user
func (e *Engine) AssignRole(userID string, role Role) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.userRoles[userID] = role
}

// GetUserRole returns the role for a user
func (e *Engine) GetUserRole(userID string) (Role, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	role, ok := e.userRoles[userID]
	return role, ok
}

// RemoveUser removes a user from the system
func (e *Engine) RemoveUser(userID string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.userRoles, userID)
}

// HasPermission checks if a role has a specific permission on a resource
func (e *Engine) HasPermission(role Role, resource Resource, permission Permission) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	permissions, ok := e.rolePermissions[role]
	if !ok {
		return false
	}

	resourcePerms, ok := permissions[resource]
	if !ok {
		return false
	}

	for _, p := range resourcePerms {
		if p == permission {
			return true
		}
	}
	return false
}

// GetPermissions returns all permissions for a role on all resources
func (e *Engine) GetPermissions(role Role) map[Resource][]Permission {
	e.mu.RLock()
	defer e.mu.RUnlock()

	permissions, ok := e.rolePermissions[role]
	if !ok {
		return make(map[Resource][]Permission)
	}

	// Return a copy to avoid external modifications
	result := make(map[Resource][]Permission)
	for resource, perms := range permissions {
		result[resource] = append([]Permission(nil), perms...)
	}
	return result
}

// IsRoleAtLeast checks if a role is at least as privileged as a minimum role
// Returns true if role >= minimumRole in the hierarchy
func (e *Engine) IsRoleAtLeast(role, minimumRole Role) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	roleLevel, ok := e.roleHierarchy[role]
	if !ok {
		return false
	}

	minLevel, ok := e.roleHierarchy[minimumRole]
	if !ok {
		return false
	}

	return roleLevel >= minLevel
}

// UserCanAccess checks if a user can access a resource with a specific permission
func (e *Engine) UserCanAccess(userID string, resource Resource, permission Permission) bool {
	role, ok := e.GetUserRole(userID)
	if !ok {
		return false
	}

	return e.HasPermission(role, resource, permission)
}

// GetRoleHierarchyLevel returns the hierarchy level of a role
// Higher values mean higher privilege
func (e *Engine) GetRoleHierarchyLevel(role Role) (int, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	level, ok := e.roleHierarchy[role]
	return level, ok
}
