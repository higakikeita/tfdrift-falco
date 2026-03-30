// Package policy provides OPA/Rego-based drift policy evaluation.
package policy

// Decision represents the policy engine's verdict for a drift event.
type Decision string

const (
	// DecisionAllow means the drift is acceptable and should be silently ignored.
	DecisionAllow Decision = "allow"
	// DecisionAlert means the drift should trigger an alert/notification.
	DecisionAlert Decision = "alert"
	// DecisionRemediate means the drift should be auto-remediated.
	DecisionRemediate Decision = "remediate"
	// DecisionDeny means the drift is a policy violation and must be escalated.
	DecisionDeny Decision = "deny"
)

// EvalResult holds the output of a policy evaluation.
type EvalResult struct {
	Decision    Decision          `json:"decision"`
	Reason      string            `json:"reason,omitempty"`
	Severity    string            `json:"severity,omitempty"`    // override severity from policy
	Labels      map[string]string `json:"labels,omitempty"`      // additional labels for routing
	Suppressors []string          `json:"suppressors,omitempty"` // reasons why the drift is suppressed
}

// DriftInput is the input document passed to Rego policies for drift evaluation.
type DriftInput struct {
	Type         string                 `json:"type"` // "drift" or "unmanaged"
	Provider     string                 `json:"provider"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	ResourceName string                 `json:"resource_name"`
	Attribute    string                 `json:"attribute,omitempty"`
	OldValue     interface{}            `json:"old_value,omitempty"`
	NewValue     interface{}            `json:"new_value,omitempty"`
	Severity     string                 `json:"severity"`
	UserIdentity UserInput              `json:"user_identity"`
	Timestamp    string                 `json:"timestamp,omitempty"`
	Changes      map[string]interface{} `json:"changes,omitempty"`
}

// UserInput is the identity portion of the input document.
type UserInput struct {
	Type        string `json:"type"`
	PrincipalID string `json:"principal_id"`
	ARN         string `json:"arn"`
	AccountID   string `json:"account_id"`
	UserName    string `json:"user_name"`
}
