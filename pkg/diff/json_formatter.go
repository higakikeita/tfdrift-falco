package diff

import (
	"encoding/json"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatJSON formats the drift as JSON
func (f *DiffFormatter) FormatJSON(alert *types.DriftAlert) (string, error) {
	// Create a structured diff object
	diff := map[string]interface{}{
		"severity":      alert.Severity,
		"resource_type": alert.ResourceType,
		"resource_name": alert.ResourceName,
		"resource_id":   alert.ResourceID,
		"attribute":     alert.Attribute,
		"change": map[string]interface{}{
			"old_value": alert.OldValue,
			"new_value": alert.NewValue,
		},
		"user": map[string]string{
			"name":         alert.UserIdentity.UserName,
			"type":         alert.UserIdentity.Type,
			"arn":          alert.UserIdentity.ARN,
			"account_id":   alert.UserIdentity.AccountID,
			"principal_id": alert.UserIdentity.PrincipalID,
		},
		"timestamp":     alert.Timestamp,
		"matched_rules": alert.MatchedRules,
		"terraform_code": map[string]string{
			"state_definition": f.formatTerraformResource(alert, alert.OldValue),
			"actual_config":    f.formatTerraformResource(alert, alert.NewValue),
		},
	}

	jsonBytes, err := json.MarshalIndent(diff, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
