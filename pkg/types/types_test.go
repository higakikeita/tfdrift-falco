package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIdentity_Creation(t *testing.T) {
	tests := []struct {
		name     string
		identity UserIdentity
		wantType string
	}{
		{
			name: "IAM User",
			identity: UserIdentity{
				Type:        "IAMUser",
				PrincipalID: "AIDAI123456789",
				ARN:         "arn:aws:iam::123456789012:user/admin",
				AccountID:   "123456789012",
				UserName:    "admin",
			},
			wantType: "IAMUser",
		},
		{
			name: "IAM Role",
			identity: UserIdentity{
				Type:        "AssumedRole",
				PrincipalID: "AIDAI987654321",
				ARN:         "arn:aws:sts::123456789012:assumed-role/MyRole/session",
				AccountID:   "123456789012",
				UserName:    "MyRole",
			},
			wantType: "AssumedRole",
		},
		{
			name: "Root User",
			identity: UserIdentity{
				Type:      "Root",
				ARN:       "arn:aws:iam::123456789012:root",
				AccountID: "123456789012",
				UserName:  "root",
			},
			wantType: "Root",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.identity.Type)
			assert.NotEmpty(t, tt.identity.AccountID)
			if tt.identity.Type != "Root" {
				assert.NotEmpty(t, tt.identity.PrincipalID)
			}
		})
	}
}

func TestUserIdentity_JSONSerialization(t *testing.T) {
	original := UserIdentity{
		Type:        "IAMUser",
		PrincipalID: "AIDAI123456789",
		ARN:         "arn:aws:iam::123456789012:user/admin",
		AccountID:   "123456789012",
		UserName:    "admin",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var decoded UserIdentity
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, original.Type, decoded.Type)
	assert.Equal(t, original.PrincipalID, decoded.PrincipalID)
	assert.Equal(t, original.ARN, decoded.ARN)
	assert.Equal(t, original.AccountID, decoded.AccountID)
	assert.Equal(t, original.UserName, decoded.UserName)
}

func TestEvent_Creation(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		wantName string
	}{
		{
			name: "EC2 Instance Termination",
			event: Event{
				Provider:     "aws",
				EventName:    "TerminateInstances",
				ResourceType: "AWS::EC2::Instance",
				ResourceID:   "i-1234567890abcdef0",
				UserIdentity: UserIdentity{
					Type:     "IAMUser",
					UserName: "admin",
				},
				Changes: map[string]interface{}{
					"state": "terminated",
				},
			},
			wantName: "TerminateInstances",
		},
		{
			name: "IAM Policy Change",
			event: Event{
				Provider:     "aws",
				EventName:    "PutUserPolicy",
				ResourceType: "AWS::IAM::User",
				ResourceID:   "test-user",
				UserIdentity: UserIdentity{
					Type:     "AssumedRole",
					UserName: "DevRole",
				},
				Changes: map[string]interface{}{
					"policyDocument": "...",
				},
			},
			wantName: "PutUserPolicy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantName, tt.event.EventName)
			assert.NotEmpty(t, tt.event.Provider)
			assert.NotEmpty(t, tt.event.ResourceType)
			assert.NotEmpty(t, tt.event.ResourceID)
			assert.NotNil(t, tt.event.Changes)
		})
	}
}

func TestEvent_JSONSerialization(t *testing.T) {
	original := Event{
		Provider:     "aws",
		EventName:    "TerminateInstances",
		ResourceType: "AWS::EC2::Instance",
		ResourceID:   "i-1234567890abcdef0",
		UserIdentity: UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI123456789",
			UserName:    "admin",
		},
		Changes: map[string]interface{}{
			"state":  "terminated",
			"reason": "User initiated",
		},
		RawEvent: map[string]interface{}{
			"eventSource": "ec2.amazonaws.com",
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var decoded Event
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, original.Provider, decoded.Provider)
	assert.Equal(t, original.EventName, decoded.EventName)
	assert.Equal(t, original.ResourceType, decoded.ResourceType)
	assert.Equal(t, original.ResourceID, decoded.ResourceID)
	assert.Equal(t, original.UserIdentity.Type, decoded.UserIdentity.Type)
	assert.NotNil(t, decoded.Changes)
	assert.Equal(t, "terminated", decoded.Changes["state"])
}

func TestDriftAlert_Creation(t *testing.T) {
	tests := []struct {
		name      string
		alert     DriftAlert
		wantType  string
		wantValue string
	}{
		{
			name: "Critical Drift - Security Group",
			alert: DriftAlert{
				Severity:     "critical",
				ResourceType: "AWS::EC2::SecurityGroup",
				ResourceName: "app-sg",
				ResourceID:   "sg-123456",
				Attribute:    "ingress_rules",
				OldValue:     []string{"22"},
				NewValue:     []string{"22", "80", "443"},
				UserIdentity: UserIdentity{
					Type:     "IAMUser",
					UserName: "operator",
				},
				MatchedRules: []string{"critical-port-change"},
				Timestamp:    "2025-01-15T10:00:00Z",
				AlertType:    "drift",
			},
			wantType:  "drift",
			wantValue: "critical",
		},
		{
			name: "Warning Drift - Tag Change",
			alert: DriftAlert{
				Severity:     "warning",
				ResourceType: "AWS::EC2::Instance",
				ResourceName: "web-server",
				ResourceID:   "i-abcdef123",
				Attribute:    "tags",
				OldValue:     map[string]string{"env": "prod"},
				NewValue:     map[string]string{"env": "prod", "team": "ops"},
				UserIdentity: UserIdentity{
					Type:     "AssumedRole",
					UserName: "DevRole",
				},
				MatchedRules: []string{"tag-modification"},
				Timestamp:    "2025-01-15T11:00:00Z",
				AlertType:    "drift",
			},
			wantType:  "drift",
			wantValue: "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantType, tt.alert.AlertType)
			assert.Equal(t, tt.wantValue, tt.alert.Severity)
			assert.NotEmpty(t, tt.alert.ResourceType)
			assert.NotEmpty(t, tt.alert.ResourceID)
			assert.NotNil(t, tt.alert.OldValue)
			assert.NotNil(t, tt.alert.NewValue)
			assert.NotEmpty(t, tt.alert.MatchedRules)
		})
	}
}

func TestDriftAlert_JSONSerialization(t *testing.T) {
	original := DriftAlert{
		Severity:     "critical",
		ResourceType: "AWS::EC2::SecurityGroup",
		ResourceName: "app-sg",
		ResourceID:   "sg-123456",
		Attribute:    "ingress_rules",
		OldValue:     "rule1",
		NewValue:     "rule2",
		UserIdentity: UserIdentity{
			Type:     "IAMUser",
			UserName: "operator",
		},
		MatchedRules: []string{"critical-port-change"},
		Timestamp:    "2025-01-15T10:00:00Z",
		AlertType:    "drift",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var decoded DriftAlert
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, original.Severity, decoded.Severity)
	assert.Equal(t, original.ResourceType, decoded.ResourceType)
	assert.Equal(t, original.ResourceName, decoded.ResourceName)
	assert.Equal(t, original.ResourceID, decoded.ResourceID)
	assert.Equal(t, original.Attribute, decoded.Attribute)
	assert.Equal(t, original.AlertType, decoded.AlertType)
	assert.Equal(t, original.Timestamp, decoded.Timestamp)
	assert.Equal(t, len(original.MatchedRules), len(decoded.MatchedRules))
}

func TestUnmanagedResourceAlert_Creation(t *testing.T) {
	tests := []struct {
		name       string
		alert      UnmanagedResourceAlert
		wantReason string
	}{
		{
			name: "Unmanaged EC2 Instance",
			alert: UnmanagedResourceAlert{
				Severity:     "high",
				ResourceType: "AWS::EC2::Instance",
				ResourceID:   "i-unmanaged123",
				EventName:    "RunInstances",
				UserIdentity: UserIdentity{
					Type:     "IAMUser",
					UserName: "developer",
				},
				Changes: map[string]interface{}{
					"instanceType": "t3.medium",
				},
				Timestamp: "2025-01-15T12:00:00Z",
				Reason:    "Resource not found in Terraform state",
			},
			wantReason: "Resource not found in Terraform state",
		},
		{
			name: "Unmanaged S3 Bucket",
			alert: UnmanagedResourceAlert{
				Severity:     "medium",
				ResourceType: "AWS::S3::Bucket",
				ResourceID:   "unmanaged-bucket-123",
				EventName:    "CreateBucket",
				UserIdentity: UserIdentity{
					Type:     "AssumedRole",
					UserName: "AdminRole",
				},
				Changes: map[string]interface{}{
					"bucketName": "unmanaged-bucket-123",
				},
				Timestamp: "2025-01-15T13:00:00Z",
				Reason:    "Created outside Terraform workflow",
			},
			wantReason: "Created outside Terraform workflow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantReason, tt.alert.Reason)
			assert.NotEmpty(t, tt.alert.Severity)
			assert.NotEmpty(t, tt.alert.ResourceType)
			assert.NotEmpty(t, tt.alert.ResourceID)
			assert.NotEmpty(t, tt.alert.EventName)
			assert.NotNil(t, tt.alert.Changes)
		})
	}
}

func TestUnmanagedResourceAlert_JSONSerialization(t *testing.T) {
	original := UnmanagedResourceAlert{
		Severity:     "high",
		ResourceType: "AWS::EC2::Instance",
		ResourceID:   "i-unmanaged123",
		EventName:    "RunInstances",
		UserIdentity: UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI999999999",
			UserName:    "developer",
		},
		Changes: map[string]interface{}{
			"instanceType": "t3.medium",
			"region":       "us-east-1",
		},
		Timestamp: "2025-01-15T12:00:00Z",
		Reason:    "Resource not found in Terraform state",
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(original)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Unmarshal back
	var decoded UnmanagedResourceAlert
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	// Verify fields
	assert.Equal(t, original.Severity, decoded.Severity)
	assert.Equal(t, original.ResourceType, decoded.ResourceType)
	assert.Equal(t, original.ResourceID, decoded.ResourceID)
	assert.Equal(t, original.EventName, decoded.EventName)
	assert.Equal(t, original.Reason, decoded.Reason)
	assert.Equal(t, original.Timestamp, decoded.Timestamp)
	assert.Equal(t, original.UserIdentity.Type, decoded.UserIdentity.Type)
	assert.NotNil(t, decoded.Changes)
	assert.Equal(t, "t3.medium", decoded.Changes["instanceType"])
}

func TestDriftAlert_SeverityLevels(t *testing.T) {
	severities := []string{"critical", "high", "medium", "low", "info"}

	for _, severity := range severities {
		t.Run("Severity_"+severity, func(t *testing.T) {
			alert := DriftAlert{
				Severity:     severity,
				ResourceType: "AWS::EC2::Instance",
				ResourceID:   "i-test",
				AlertType:    "drift",
			}
			assert.Equal(t, severity, alert.Severity)
		})
	}
}

func TestAlertType_Values(t *testing.T) {
	tests := []struct {
		name      string
		alertType string
		valid     bool
	}{
		{"Drift Alert", "drift", true},
		{"Unmanaged Alert", "unmanaged", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := DriftAlert{
				AlertType: tt.alertType,
			}
			assert.Equal(t, tt.alertType, alert.AlertType)
		})
	}
}
