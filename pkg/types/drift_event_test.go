package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDriftEvent(t *testing.T) {
	event := NewDriftEvent("aws", "aws_security_group", "sg-12345", ChangeTypeModified)

	assert.Equal(t, "terraform_drift_detected", event.EventType)
	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "aws_security_group", event.ResourceType)
	assert.Equal(t, "sg-12345", event.ResourceID)
	assert.Equal(t, ChangeTypeModified, event.ChangeType)
	assert.Equal(t, "tfdrift-falco", event.Source)
	assert.Equal(t, SeverityMedium, event.Severity)
	assert.Equal(t, EventSchemaVersion, event.Version)
	assert.False(t, event.DetectedAt.IsZero())
}

func TestDriftEvent_ToJSON(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithRegion("us-west-2")
	event.WithUser("john@example.com")
	event.WithAccountID("123456789012")

	data, err := event.ToJSON()
	require.NoError(t, err)

	// Verify it's valid JSON
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "terraform_drift_detected", decoded["event_type"])
	assert.Equal(t, "aws", decoded["provider"])
	assert.Equal(t, "aws_instance", decoded["resource_type"])
	assert.Equal(t, "i-12345", decoded["resource_id"])
	assert.Equal(t, "us-west-2", decoded["region"])
	assert.Equal(t, "john@example.com", decoded["user"])
	assert.Equal(t, "123456789012", decoded["account_id"])
}

func TestDriftEvent_ToJSONString(t *testing.T) {
	event := NewDriftEvent("aws", "aws_db_instance", "db-12345", ChangeTypeDeleted)

	str, err := event.ToJSONString()
	require.NoError(t, err)
	assert.Contains(t, str, "terraform_drift_detected")
	assert.Contains(t, str, "aws_db_instance")
	assert.Contains(t, str, "db-12345")
}

func TestDriftEvent_WithSeverity(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithSeverity(SeverityCritical)

	assert.Equal(t, SeverityCritical, event.Severity)
}

func TestDriftEvent_WithRegion(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithRegion("ap-northeast-1")

	assert.Equal(t, "ap-northeast-1", event.Region)
}

func TestDriftEvent_WithUser(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithUser("alice@example.com")

	assert.Equal(t, "alice@example.com", event.User)
}

func TestDriftEvent_WithAccountID(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithAccountID("987654321098")

	assert.Equal(t, "987654321098", event.AccountID)
}

func TestDriftEvent_WithCloudTrailEvent(t *testing.T) {
	event := NewDriftEvent("aws", "aws_security_group", "sg-12345", ChangeTypeModified)
	event.WithCloudTrailEvent("AuthorizeSecurityGroupIngress", "req-123")

	assert.Equal(t, "AuthorizeSecurityGroupIngress", event.CloudTrailEvent)
	assert.Equal(t, "req-123", event.RequestID)
}

func TestDriftEvent_WithFalcoContext(t *testing.T) {
	event := NewDriftEvent("aws", "aws_iam_role", "role-admin", ChangeTypeModified)
	tags := []string{"security", "iam"}
	event.WithFalcoContext("IAM Role Modified", "WARNING", tags)

	assert.Equal(t, "IAM Role Modified", event.FalcoRule)
	assert.Equal(t, "WARNING", event.FalcoPriority)
	assert.Equal(t, tags, event.FalcoTags)
}

func TestDriftEvent_WithStateContext(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithStateContext("production", "s3")

	assert.Equal(t, "production", event.TerraformWorkspace)
	assert.Equal(t, "s3", event.StateBackend)
}

func TestDriftEvent_WithDiff(t *testing.T) {
	event := NewDriftEvent("aws", "aws_security_group", "sg-12345", ChangeTypeModified)
	expected := map[string]interface{}{"ingress": []string{"0.0.0.0/0"}}
	actual := map[string]interface{}{"ingress": []string{"10.0.0.0/8"}}
	diff := "- ingress: 0.0.0.0/0\n+ ingress: 10.0.0.0/8"

	event.WithDiff(expected, actual, diff)

	assert.Equal(t, expected, event.Expected)
	assert.Equal(t, actual, event.Actual)
	assert.Equal(t, diff, event.Diff)
}

func TestDriftEvent_WithLabel(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)
	event.WithLabel("environment", "production")
	event.WithLabel("team", "platform")

	assert.Equal(t, "production", event.Labels["environment"])
	assert.Equal(t, "platform", event.Labels["team"])
	assert.Len(t, event.Labels, 2)
}

func TestDriftEvent_ChainedBuilders(t *testing.T) {
	event := NewDriftEvent("aws", "aws_security_group", "sg-12345", ChangeTypeModified).
		WithSeverity(SeverityCritical).
		WithRegion("us-east-1").
		WithUser("admin@example.com").
		WithAccountID("123456789012").
		WithCloudTrailEvent("AuthorizeSecurityGroupIngress", "req-456").
		WithLabel("team", "security")

	assert.Equal(t, SeverityCritical, event.Severity)
	assert.Equal(t, "us-east-1", event.Region)
	assert.Equal(t, "admin@example.com", event.User)
	assert.Equal(t, "123456789012", event.AccountID)
	assert.Equal(t, "AuthorizeSecurityGroupIngress", event.CloudTrailEvent)
	assert.Equal(t, "security", event.Labels["team"])
}

func TestDetermineSeverity_Critical(t *testing.T) {
	tests := []struct {
		resourceType string
	}{
		{"aws_iam_role"},
		{"aws_iam_policy"},
		{"aws_security_group"},
		{"aws_kms_key"},
		{"aws_s3_bucket_policy"},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			severity := DetermineSeverity(tt.resourceType, ChangeTypeModified)
			assert.Equal(t, SeverityCritical, severity)
		})
	}
}

func TestDetermineSeverity_High(t *testing.T) {
	tests := []struct {
		resourceType string
	}{
		{"aws_instance"},
		{"aws_db_instance"},
		{"aws_rds_cluster"},
		{"aws_lambda_function"},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			severity := DetermineSeverity(tt.resourceType, ChangeTypeModified)
			assert.Equal(t, SeverityHigh, severity)
		})
	}
}

func TestDetermineSeverity_Deleted(t *testing.T) {
	// Any deletion should be high severity
	severity := DetermineSeverity("aws_s3_bucket", ChangeTypeDeleted)
	assert.Equal(t, SeverityHigh, severity)
}

func TestDetermineSeverity_Created(t *testing.T) {
	// Creation is usually low severity
	severity := DetermineSeverity("aws_s3_bucket", ChangeTypeCreated)
	assert.Equal(t, SeverityLow, severity)
}

func TestDetermineSeverity_DefaultMedium(t *testing.T) {
	// Unknown resource type with modification
	severity := DetermineSeverity("aws_unknown_resource", ChangeTypeModified)
	assert.Equal(t, SeverityMedium, severity)
}

func TestDriftEvent_JSONRoundtrip(t *testing.T) {
	original := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeModified).
		WithSeverity(SeverityHigh).
		WithRegion("us-west-2").
		WithUser("user@example.com").
		WithAccountID("123456789012")

	// Serialize
	data, err := original.ToJSON()
	require.NoError(t, err)

	// Deserialize
	var decoded DriftEvent
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Compare
	assert.Equal(t, original.EventType, decoded.EventType)
	assert.Equal(t, original.Provider, decoded.Provider)
	assert.Equal(t, original.ResourceType, decoded.ResourceType)
	assert.Equal(t, original.ResourceID, decoded.ResourceID)
	assert.Equal(t, original.ChangeType, decoded.ChangeType)
	assert.Equal(t, original.Severity, decoded.Severity)
	assert.Equal(t, original.Region, decoded.Region)
	assert.Equal(t, original.User, decoded.User)
	assert.Equal(t, original.AccountID, decoded.AccountID)
	assert.Equal(t, original.Source, decoded.Source)
	assert.Equal(t, original.Version, decoded.Version)

	// Time comparison (within 1 second)
	assert.WithinDuration(t, original.DetectedAt, decoded.DetectedAt, time.Second)
}

func TestDriftEvent_EmptyFields(t *testing.T) {
	event := NewDriftEvent("aws", "aws_instance", "i-12345", ChangeTypeCreated)

	data, err := event.ToJSON()
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Optional fields should not be present in JSON when empty
	_, hasRegion := decoded["region"]
	assert.False(t, hasRegion, "Empty optional fields should be omitted")

	_, hasUser := decoded["user"]
	assert.False(t, hasUser, "Empty optional fields should be omitted")
}

func TestChangeTypeConstants(t *testing.T) {
	assert.Equal(t, "created", ChangeTypeCreated)
	assert.Equal(t, "modified", ChangeTypeModified)
	assert.Equal(t, "deleted", ChangeTypeDeleted)
	assert.Equal(t, "unknown", ChangeTypeUnknown)
}

func TestSeverityConstants(t *testing.T) {
	assert.Equal(t, "critical", SeverityCritical)
	assert.Equal(t, "high", SeverityHigh)
	assert.Equal(t, "medium", SeverityMedium)
	assert.Equal(t, "low", SeverityLow)
	assert.Equal(t, "info", SeverityInfo)
}
