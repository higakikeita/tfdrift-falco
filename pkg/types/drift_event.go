package types

import (
	"encoding/json"
	"time"
)

// DriftEvent represents a Terraform drift detection event
// This is the core event model for system integrations (SIEM, SOAR, webhooks, etc.)
type DriftEvent struct {
	// Core fields (always present)
	EventType    string    `json:"event_type"`              // Always "terraform_drift_detected"
	Provider     string    `json:"provider"`                // Cloud provider (aws, gcp, azure)
	AccountID    string    `json:"account_id,omitempty"`    // Cloud account/project ID
	ResourceType string    `json:"resource_type"`           // Terraform resource type (e.g., aws_security_group)
	ResourceID   string    `json:"resource_id"`             // Cloud resource ID
	ChangeType   string    `json:"change_type"`             // created, modified, deleted, unknown
	DetectedAt   time.Time `json:"detected_at"`             // When drift was detected (RFC3339)
	Source       string    `json:"source"`                  // Always "tfdrift-falco"
	Severity     string    `json:"severity"`                // critical, high, medium, low, info

	// Optional context fields
	Region       string `json:"region,omitempty"`        // Cloud region
	User         string `json:"user,omitempty"`          // IAM user who made the change
	IPAddress    string `json:"ip_address,omitempty"`    // Source IP address
	UserAgent    string `json:"user_agent,omitempty"`    // User agent string

	// Terraform state context
	TerraformWorkspace string `json:"terraform_workspace,omitempty"` // Terraform workspace
	StateBackend       string `json:"state_backend,omitempty"`       // local, s3, etc.

	// Change details (optional, can be large)
	Expected     any    `json:"expected,omitempty"`      // Expected state from Terraform
	Actual       any    `json:"actual,omitempty"`        // Actual state from cloud
	Diff         any    `json:"diff,omitempty"`          // Human-readable diff

	// CloudTrail context
	CloudTrailEvent string `json:"cloudtrail_event,omitempty"` // CloudTrail event name
	RequestID       string `json:"request_id,omitempty"`       // CloudTrail request ID

	// Falco context
	FalcoRule     string            `json:"falco_rule,omitempty"`     // Falco rule that triggered
	FalcoPriority string            `json:"falco_priority,omitempty"` // Falco priority
	FalcoTags     []string          `json:"falco_tags,omitempty"`     // Falco rule tags
	FalcoFields   map[string]string `json:"falco_fields,omitempty"`   // Additional Falco fields

	// Metadata
	Version string            `json:"version"`           // Event schema version
	Labels  map[string]string `json:"labels,omitempty"`  // Custom labels
}

// ChangeType constants
const (
	ChangeTypeCreated  = "created"
	ChangeTypeModified = "modified"
	ChangeTypeDeleted  = "deleted"
	ChangeTypeUnknown  = "unknown"
)

// Severity constants
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// EventSchemaVersion is the current schema version
const EventSchemaVersion = "1.0.0"

// NewDriftEvent creates a new DriftEvent with required fields
func NewDriftEvent(provider, resourceType, resourceID, changeType string) *DriftEvent {
	return &DriftEvent{
		EventType:    "terraform_drift_detected",
		Provider:     provider,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		ChangeType:   changeType,
		DetectedAt:   time.Now().UTC(),
		Source:       "tfdrift-falco",
		Severity:     SeverityMedium, // Default severity
		Version:      EventSchemaVersion,
	}
}

// ToJSON serializes the event to JSON
func (e *DriftEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// ToJSONString serializes the event to a JSON string
func (e *DriftEvent) ToJSONString() (string, error) {
	data, err := e.ToJSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WithSeverity sets the severity level
func (e *DriftEvent) WithSeverity(severity string) *DriftEvent {
	e.Severity = severity
	return e
}

// WithRegion sets the cloud region
func (e *DriftEvent) WithRegion(region string) *DriftEvent {
	e.Region = region
	return e
}

// WithUser sets the IAM user
func (e *DriftEvent) WithUser(user string) *DriftEvent {
	e.User = user
	return e
}

// WithAccountID sets the cloud account ID
func (e *DriftEvent) WithAccountID(accountID string) *DriftEvent {
	e.AccountID = accountID
	return e
}

// WithCloudTrailEvent sets CloudTrail event details
func (e *DriftEvent) WithCloudTrailEvent(eventName, requestID string) *DriftEvent {
	e.CloudTrailEvent = eventName
	e.RequestID = requestID
	return e
}

// WithFalcoContext sets Falco-related context
func (e *DriftEvent) WithFalcoContext(rule, priority string, tags []string) *DriftEvent {
	e.FalcoRule = rule
	e.FalcoPriority = priority
	e.FalcoTags = tags
	return e
}

// WithStateContext sets Terraform state context
func (e *DriftEvent) WithStateContext(workspace, backend string) *DriftEvent {
	e.TerraformWorkspace = workspace
	e.StateBackend = backend
	return e
}

// WithDiff sets the diff information
func (e *DriftEvent) WithDiff(expected, actual, diff any) *DriftEvent {
	e.Expected = expected
	e.Actual = actual
	e.Diff = diff
	return e
}

// WithLabel adds a custom label
func (e *DriftEvent) WithLabel(key, value string) *DriftEvent {
	if e.Labels == nil {
		e.Labels = make(map[string]string)
	}
	e.Labels[key] = value
	return e
}

// DetermineSeverity determines severity based on resource type and change
// This can be customized based on organizational policies
func DetermineSeverity(resourceType, changeType string) string {
	// Critical resources
	criticalResources := map[string]bool{
		"aws_iam_role":             true,
		"aws_iam_policy":           true,
		"aws_security_group":       true,
		"aws_network_acl":          true,
		"aws_kms_key":              true,
		"aws_s3_bucket_policy":     true,
		"aws_iam_account_password_policy": true,
	}

	// High priority resources
	highResources := map[string]bool{
		"aws_instance":              true,
		"aws_db_instance":           true,
		"aws_rds_cluster":           true,
		"aws_lambda_function":       true,
		"aws_ecs_service":           true,
		"aws_eks_cluster":           true,
		"aws_elasticache_cluster":   true,
	}

	if criticalResources[resourceType] {
		return SeverityCritical
	}

	if highResources[resourceType] {
		return SeverityHigh
	}

	// Deletion is always more severe
	if changeType == ChangeTypeDeleted {
		return SeverityHigh
	}

	// Creation is usually less concerning
	if changeType == ChangeTypeCreated {
		return SeverityLow
	}

	// Default for modifications
	return SeverityMedium
}
