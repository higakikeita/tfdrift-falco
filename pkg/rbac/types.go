package rbac

// Role represents a user role in the system
type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
	RoleViewer Role = "viewer"
)

// Permission represents an action that can be performed
type Permission string

const (
	PermRead        Permission = "read"
	PermWrite       Permission = "write"
	PermDelete      Permission = "delete"
	PermAdmin       Permission = "admin"
	PermManageUsers Permission = "manage_users"
)

// Resource represents an entity that can be accessed
type Resource string

const (
	ResourceDrifts    Resource = "drifts"
	ResourceEvents    Resource = "events"
	ResourceSettings  Resource = "settings"
	ResourceUsers     Resource = "users"
	ResourceProviders Resource = "providers"
)

// RolePermissions defines what each role can do on each resource
var RolePermissions = map[Role]map[Resource][]Permission{
	RoleAdmin: {
		ResourceDrifts:    {PermRead, PermWrite, PermDelete, PermAdmin},
		ResourceEvents:    {PermRead, PermWrite, PermDelete, PermAdmin},
		ResourceSettings:  {PermRead, PermWrite, PermDelete, PermAdmin},
		ResourceUsers:     {PermRead, PermWrite, PermDelete, PermManageUsers, PermAdmin},
		ResourceProviders: {PermRead, PermWrite, PermDelete, PermAdmin},
	},
	RoleEditor: {
		ResourceDrifts:    {PermRead, PermWrite},
		ResourceEvents:    {PermRead, PermWrite},
		ResourceSettings:  {PermRead},
		ResourceUsers:     {},
		ResourceProviders: {PermRead, PermWrite},
	},
	RoleViewer: {
		ResourceDrifts:    {PermRead},
		ResourceEvents:    {PermRead},
		ResourceSettings:  {},
		ResourceUsers:     {},
		ResourceProviders: {PermRead},
	},
}

// RoleHierarchy defines the hierarchy of roles (higher index = higher privilege)
// This allows checking if a role is at least as privileged as another
var RoleHierarchy = map[Role]int{
	RoleViewer: 0,
	RoleEditor: 1,
	RoleAdmin:  2,
}

