// Package types defines core data structures used throughout TFDrift-Falco.
package types

// Event represents a cloud event that might indicate drift
type Event struct {
	Provider     string
	EventName    string
	ResourceType string
	ResourceID   string
	UserIdentity UserIdentity
	Changes      map[string]interface{}
	RawEvent     interface{}
}

// UserIdentity represents the identity of the user who made the change
type UserIdentity struct {
	Type        string
	PrincipalID string
	ARN         string
	AccountID   string
	UserName    string
}

// DriftAlert represents a detected drift
type DriftAlert struct {
	Severity     string
	ResourceType string
	ResourceName string
	ResourceID   string
	Attribute    string
	OldValue     interface{}
	NewValue     interface{}
	UserIdentity UserIdentity
	MatchedRules []string
	Timestamp    string
	AlertType    string // "drift" or "unmanaged"
}

// UnmanagedResourceAlert represents a resource not managed by Terraform
type UnmanagedResourceAlert struct {
	Severity     string
	ResourceType string
	ResourceID   string
	EventName    string
	UserIdentity UserIdentity
	Changes      map[string]interface{}
	Timestamp    string
	Reason       string // Why it's considered unmanaged
}
