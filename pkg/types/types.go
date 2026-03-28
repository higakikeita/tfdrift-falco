// Package types defines core data structures used throughout TFDrift-Falco.
package types

// EventStatus represents the status of a drift event
type EventStatus string

const (
	EventStatusOpen         EventStatus = "open"
	EventStatusAcknowledged EventStatus = "acknowledged"
	EventStatusIgnored      EventStatus = "ignored"
	EventStatusResolved     EventStatus = "resolved"
)

// Event represents a cloud event that might indicate drift
type Event struct {
	Provider     string
	EventName    string
	ResourceType string
	ResourceID   string
	UserIdentity UserIdentity
	Changes      map[string]interface{}
	RawEvent     interface{}
	Timestamp    string      // ISO 8601 timestamp
	Severity     string      // critical, high, medium, low
	Status       EventStatus // open, acknowledged, ignored, resolved
	StatusReason string      // reason for status change (e.g., ignore reason)

	// AWS-specific fields
	Region string // AWS Region (optional)

	// GCP-specific fields
	ProjectID   string // GCP Project ID (optional)
	ServiceName string // GCP Service Name (e.g., compute.googleapis.com) (optional)

	// Azure-specific fields
	SubscriptionID string // Azure Subscription ID (optional)
	ResourceGroup  string // Azure Resource Group (optional)

	// Metadata holds provider-specific fields in a unified way.
	// AWS examples:  "region" -> "us-east-1"
	// GCP examples:  "project_id" -> "my-project", "zone" -> "us-central1-a", "service_name" -> "compute.googleapis.com"
	// Azure examples: "subscription_id" -> "...", "resource_group" -> "...", "region" -> "eastus"
	Metadata map[string]string
}

// GetMetadata returns a metadata value for the given key, or empty string if not found.
func (e *Event) GetMetadata(key string) string {
	if e.Metadata == nil {
		return ""
	}
	return e.Metadata[key]
}

// SetMetadata sets a metadata value. Initializes the map if nil.
func (e *Event) SetMetadata(key, value string) {
	if e.Metadata == nil {
		e.Metadata = make(map[string]string)
	}
	e.Metadata[key] = value
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

// DiscoveredResource represents a resource found in a cloud provider.
// This is the provider-agnostic version; each provider maps its native resources to this.
type DiscoveredResource struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`       // Terraform resource type (e.g., "aws_instance", "google_compute_instance")
	Provider   string                 `json:"provider"`   // Provider name (e.g., "aws", "gcp", "azure")
	ARN        string                 `json:"arn,omitempty"`
	Name       string                 `json:"name"`
	Region     string                 `json:"region"`
	Attributes map[string]interface{} `json:"attributes"`
	Tags       map[string]string      `json:"tags,omitempty"`
	Metadata   map[string]string      `json:"metadata,omitempty"` // Provider-specific metadata
}

// DriftResult represents the difference between Terraform state and actual cloud state.
// This is the provider-agnostic version used across all providers.
type DriftResult struct {
	Provider string `json:"provider"`

	// Resources in cloud but not in Terraform (manually created)
	UnmanagedResources []*DiscoveredResource `json:"unmanaged_resources"`

	// Resources in Terraform but not in cloud (manually deleted)
	MissingResources []*TerraformResource `json:"missing_resources"`

	// Resources with configuration differences
	ModifiedResources []*ResourceDiff `json:"modified_resources"`
}

// TerraformResource is a minimal representation of a Terraform-managed resource for drift results.
type TerraformResource struct {
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	ID         string                 `json:"id"`
	Provider   string                 `json:"provider"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ResourceDiff represents differences in a single resource between Terraform and actual state.
type ResourceDiff struct {
	ResourceID     string                 `json:"resource_id"`
	ResourceType   string                 `json:"resource_type"`
	Provider       string                 `json:"provider"`
	TerraformState map[string]interface{} `json:"terraform_state"`
	ActualState    map[string]interface{} `json:"actual_state"`
	Differences    []FieldDiff            `json:"differences"`
}

// FieldDiff represents a difference in a specific field.
type FieldDiff struct {
	Field          string      `json:"field"`
	TerraformValue interface{} `json:"terraform_value"`
	ActualValue    interface{} `json:"actual_value"`
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
