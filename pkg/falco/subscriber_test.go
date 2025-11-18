package falco

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/api/schema"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSubscriber(t *testing.T) {
	cfg := config.FalcoConfig{
		Enabled:  true,
		Hostname: "localhost",
		Port:     5060,
	}

	sub, err := NewSubscriber(cfg)
	require.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, cfg, sub.cfg)
}

func TestIsRelevantEvent(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		want      bool
	}{
		// EC2 Events
		{"EC2 ModifyInstanceAttribute", "ModifyInstanceAttribute", true},
		{"EC2 ModifyVolume", "ModifyVolume", true},
		{"EC2 Irrelevant", "RunInstances", false},

		// IAM Policy Events
		{"IAM PutUserPolicy", "PutUserPolicy", true},
		{"IAM AttachRolePolicy", "AttachRolePolicy", true},
		{"IAM CreatePolicy", "CreatePolicy", true},
		{"IAM CreatePolicyVersion", "CreatePolicyVersion", true},

		// IAM Lifecycle Events
		{"IAM CreateRole", "CreateRole", true},
		{"IAM DeleteRole", "DeleteRole", true},
		{"IAM CreateUser", "CreateUser", true},
		{"IAM DeleteUser", "DeleteUser", true},
		{"IAM CreateAccessKey", "CreateAccessKey", true},
		{"IAM AddUserToGroup", "AddUserToGroup", true},

		// S3 Events
		{"S3 PutBucketPolicy", "PutBucketPolicy", true},
		{"S3 PutBucketEncryption", "PutBucketEncryption", true},
		{"S3 DeleteBucketEncryption", "DeleteBucketEncryption", true},
		{"S3 Irrelevant", "CreateBucket", false},

		// RDS Events
		{"RDS ModifyDBInstance", "ModifyDBInstance", true},
		{"RDS Irrelevant", "CreateDBInstance", false},

		// Lambda Events
		{"Lambda UpdateFunctionConfiguration", "UpdateFunctionConfiguration", true},
		{"Lambda Irrelevant", "CreateFunction", false},

		// Completely irrelevant
		{"Unknown Event", "SomeRandomEvent", false},
		{"Empty Event", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.isRelevantEvent(tt.eventName)
			assert.Equal(t, tt.want, got, "Event: %s", tt.eventName)
		})
	}
}

func TestExtractResourceID(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		want      string
	}{
		{
			name:      "EC2 Instance",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.instanceid": "i-1234567890abcdef0",
			},
			want: "i-1234567890abcdef0",
		},
		{
			name:      "EBS Volume",
			eventName: "ModifyVolume",
			fields: map[string]string{
				"ct.request.volumeid": "vol-123456",
			},
			want: "vol-123456",
		},
		{
			name:      "S3 Bucket",
			eventName: "PutBucketPolicy",
			fields: map[string]string{
				"ct.request.bucket": "my-bucket",
			},
			want: "my-bucket",
		},
		{
			name:      "IAM Role",
			eventName: "PutRolePolicy",
			fields: map[string]string{
				"ct.request.rolename": "my-role",
			},
			want: "my-role",
		},
		{
			name:      "IAM User",
			eventName: "CreateUser",
			fields: map[string]string{
				"ct.request.username": "john-doe",
			},
			want: "john-doe",
		},
		{
			name:      "IAM Policy by ARN",
			eventName: "CreatePolicyVersion",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::123456789012:policy/MyPolicy",
			},
			want: "arn:aws:iam::123456789012:policy/MyPolicy",
		},
		{
			name:      "Lambda Function",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.functionname": "my-function",
			},
			want: "my-function",
		},
		{
			name:      "Missing Resource ID",
			eventName: "ModifyInstanceAttribute",
			fields:    map[string]string{},
			want:      "",
		},
		{
			name:      "Unknown Event Type",
			eventName: "UnknownEvent",
			fields: map[string]string{
				"ct.resource.id": "some-id",
			},
			want: "some-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.extractResourceID(tt.eventName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapEventToResourceType(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		want      string
	}{
		// EC2
		{"EC2 Instance", "ModifyInstanceAttribute", "aws_instance"},
		{"EBS Volume", "ModifyVolume", "aws_ebs_volume"},

		// IAM Roles
		{"IAM Role Policy", "PutRolePolicy", "aws_iam_role_policy"},
		{"IAM Role", "CreateRole", "aws_iam_role"},
		{"IAM Role Assume Policy", "UpdateAssumeRolePolicy", "aws_iam_role"},
		{"IAM Role Policy Attachment", "AttachRolePolicy", "aws_iam_role_policy_attachment"},

		// IAM Users
		{"IAM User Policy", "PutUserPolicy", "aws_iam_user_policy"},
		{"IAM User", "CreateUser", "aws_iam_user"},
		{"IAM Access Key", "CreateAccessKey", "aws_iam_access_key"},

		// IAM Groups
		{"IAM Group Policy", "PutGroupPolicy", "aws_iam_group_policy"},

		// IAM Policies
		{"IAM Policy", "CreatePolicy", "aws_iam_policy"},
		{"IAM Policy Version", "CreatePolicyVersion", "aws_iam_policy"},

		// IAM Account
		{"IAM Account Password Policy", "UpdateAccountPasswordPolicy", "aws_iam_account_password_policy"},

		// S3
		{"S3 Bucket Policy", "PutBucketPolicy", "aws_s3_bucket_policy"},
		{"S3 Bucket Encryption", "PutBucketEncryption", "aws_s3_bucket"},

		// RDS
		{"RDS Instance", "ModifyDBInstance", "aws_db_instance"},

		// Lambda
		{"Lambda Function", "UpdateFunctionConfiguration", "aws_lambda_function"},

		// Unknown
		{"Unknown Event", "UnknownEvent", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.mapEventToResourceType(tt.eventName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetStringField(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		key    string
		want   string
	}{
		{
			name: "Direct Match",
			fields: map[string]string{
				"ct.user.name": "john",
			},
			key:  "ct.user.name",
			want: "john",
		},
		{
			name: "Case Insensitive Match",
			fields: map[string]string{
				"CT.User.Name": "jane",
			},
			key:  "ct.user.name",
			want: "jane",
		},
		{
			name: "Not Found",
			fields: map[string]string{
				"other.field": "value",
			},
			key:  "ct.user.name",
			want: "",
		},
		{
			name:   "Empty Map",
			fields: map[string]string{},
			key:    "ct.user.name",
			want:   "",
		},
		{
			name: "Multiple Fields with Case Variations",
			fields: map[string]string{
				"field1":      "value1",
				"Field2":      "value2",
				"ct.user.arn": "arn:aws:iam::123:user/test",
			},
			key:  "CT.USER.ARN",
			want: "arn:aws:iam::123:user/test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringField(tt.fields, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractChanges(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		wantKeys  []string
	}{
		{
			name:      "ModifyInstanceAttribute - Instance Type",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.instancetype": "t3.medium",
			},
			wantKeys: []string{"instance_type"},
		},
		{
			name:      "ModifyInstanceAttribute - API Termination",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.disableapitermination": "true",
			},
			wantKeys: []string{"disable_api_termination"},
		},
		{
			name:      "PutBucketEncryption",
			eventName: "PutBucketEncryption",
			fields: map[string]string{
				"ct.request.serversideencryptionconfiguration": "AES256",
			},
			wantKeys: []string{"server_side_encryption_configuration"},
		},
		{
			name:      "DeleteBucketEncryption",
			eventName: "DeleteBucketEncryption",
			fields:    map[string]string{},
			wantKeys:  []string{"server_side_encryption_configuration"},
		},
		{
			name:      "UpdateFunctionConfiguration - Timeout and Memory",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.timeout":    "60",
				"ct.request.memorysize": "512",
			},
			wantKeys: []string{"timeout", "memory_size"},
		},
		{
			name:      "AttachRolePolicy",
			eventName: "AttachRolePolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/ReadOnlyAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "PutRolePolicy",
			eventName: "PutRolePolicy",
			fields: map[string]string{
				"ct.request.policyname":     "inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "CreateRole",
			eventName: "CreateRole",
			fields: map[string]string{
				"ct.request.rolename":                   "my-role",
				"ct.request.assumerolepolicydocument":   `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"role_name", "assume_role_policy"},
		},
		{
			name:      "DeleteUser",
			eventName: "DeleteUser",
			fields: map[string]string{
				"ct.request.username": "old-user",
			},
			wantKeys: []string{"deleted_user"},
		},
		{
			name:      "CreateAccessKey",
			eventName: "CreateAccessKey",
			fields: map[string]string{
				"ct.request.username":                 "service-user",
				"ct.response.accesskey.accesskeyid":   "AKIAIOSFODNN7EXAMPLE",
			},
			wantKeys: []string{"user_name", "access_key_id"},
		},
		{
			name:      "AddUserToGroup",
			eventName: "AddUserToGroup",
			fields: map[string]string{
				"ct.request.username":  "john",
				"ct.request.groupname": "developers",
			},
			wantKeys: []string{"user_name", "group_name"},
		},
		{
			name:      "UpdateAccountPasswordPolicy",
			eventName: "UpdateAccountPasswordPolicy",
			fields: map[string]string{
				"ct.request.minimumpasswordlength": "14",
				"ct.request.requiresymbols":        "true",
			},
			wantKeys: []string{"minimum_password_length", "require_symbols"},
		},
		{
			name:      "DeleteRole",
			eventName: "DeleteRole",
			fields: map[string]string{
				"ct.request.rolename": "obsolete-role",
			},
			wantKeys: []string{"deleted_role"},
		},
		{
			name:      "CreateUser",
			eventName: "CreateUser",
			fields: map[string]string{
				"ct.request.username": "new-user",
			},
			wantKeys: []string{"user_name"},
		},
		{
			name:      "RemoveUserFromGroup",
			eventName: "RemoveUserFromGroup",
			fields: map[string]string{
				"ct.request.username":  "jane",
				"ct.request.groupname": "admins",
			},
			wantKeys: []string{"user_name", "group_name"},
		},
		{
			name:      "CreatePolicy",
			eventName: "CreatePolicy",
			fields: map[string]string{
				"ct.request.policyname":     "new-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17","Statement":[]}`,
			},
			wantKeys: []string{"policy_name", "policy_document"},
		},
		{
			name:      "CreatePolicyVersion",
			eventName: "CreatePolicyVersion",
			fields: map[string]string{
				"ct.request.policyarn":      "arn:aws:iam::123:policy/my-policy",
				"ct.request.setasdefault":   "true",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"policy_arn", "set_as_default", "policy_document"},
		},
		{
			name:      "AttachUserPolicy",
			eventName: "AttachUserPolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/PowerUserAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "AttachGroupPolicy",
			eventName: "AttachGroupPolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/ReadOnlyAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "PutUserPolicy",
			eventName: "PutUserPolicy",
			fields: map[string]string{
				"ct.request.policyname":     "user-inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "PutGroupPolicy",
			eventName: "PutGroupPolicy",
			fields: map[string]string{
				"ct.request.policyname":     "group-inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "UpdateFunctionConfiguration - Only Timeout",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.timeout": "30",
			},
			wantKeys: []string{"timeout"},
		},
		{
			name:      "UpdateFunctionConfiguration - Only Memory",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.memorysize": "1024",
			},
			wantKeys: []string{"memory_size"},
		},
		{
			name:      "Unknown Event",
			eventName: "UnknownEvent",
			fields: map[string]string{
				"some.field": "value",
			},
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.extractChanges(tt.eventName, tt.fields)

			// Check that all expected keys are present
			for _, key := range tt.wantKeys {
				assert.Contains(t, got, key, "Missing key: %s", key)
			}

			// Check that we don't have unexpected keys (except for special cases)
			if len(tt.wantKeys) > 0 {
				assert.Len(t, got, len(tt.wantKeys), "Unexpected number of keys")
			}
		})
	}
}

func TestParseFalcoOutput(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name     string
		response *outputs.Response
		wantNil  bool
		validate func(t *testing.T, event interface{})
	}{
		{
			name: "Valid EC2 ModifyInstanceAttribute Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":                 "ModifyInstanceAttribute",
					"ct.request.instanceid":   "i-1234567890abcdef0",
					"ct.request.instancetype": "t3.medium",
					"ct.user.type":            "IAMUser",
					"ct.user.principalid":     "AIDAI123456789",
					"ct.user.arn":             "arn:aws:iam::123456789012:user/admin",
					"ct.user.accountid":       "123456789012",
					"ct.user":                 "admin",
				},
			},
			wantNil: false,
			validate: func(t *testing.T, event interface{}) {
				e := event
				assert.NotNil(t, e)
				assert.Equal(t, "aws", e.(*outputs.Response).Source)
			},
		},
		{
			name: "Non-CloudTrail Source",
			response: &outputs.Response{
				Source:   "syscalls",
				Rule:     "Terminal Shell",
				Priority: schema.Priority_NOTICE,
			},
			wantNil: true,
		},
		{
			name: "Missing ct.name",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"other.field": "value",
				},
			},
			wantNil: true,
		},
		{
			name: "Irrelevant Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_INFORMATIONAL,
				OutputFields: map[string]string{
					"ct.name": "DescribeInstances",
				},
			},
			wantNil: true,
		},
		{
			name: "Missing Resource ID",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name": "ModifyInstanceAttribute",
					// Missing instanceid
				},
			},
			wantNil: true,
		},
		{
			name: "Valid S3 PutBucketEncryption Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS S3 Bucket Encryption Modified",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":           "PutBucketEncryption",
					"ct.request.bucket": "my-secure-bucket",
					"ct.request.serversideencryptionconfiguration": "AES256",
					"ct.user.type":       "IAMUser",
					"ct.user.accountid":  "123456789012",
					"ct.user":            "security-admin",
				},
			},
			wantNil: false,
		},
		{
			name: "Valid IAM CreateRole Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS IAM Role Created",
				Priority: schema.Priority_NOTICE,
				OutputFields: map[string]string{
					"ct.name":                         "CreateRole",
					"ct.request.rolename":             "lambda-execution-role",
					"ct.request.assumerolepolicydocument": `{"Version":"2012-10-17","Statement":[]}`,
					"ct.user.type":                    "IAMUser",
					"ct.user":                         "iam-admin",
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.parseFalcoOutput(tt.response)

			if tt.wantNil {
				assert.Nil(t, got, "Expected nil event")
			} else {
				assert.NotNil(t, got, "Expected non-nil event")
				if got != nil && tt.validate != nil {
					// For now, just check basic fields
					assert.Equal(t, "aws", got.Provider)
					assert.NotEmpty(t, got.EventName)
					assert.NotEmpty(t, got.ResourceType)
					assert.NotEmpty(t, got.ResourceID)
				}
			}
		})
	}
}

func TestExtractChanges_JSONParsing(t *testing.T) {
	sub := &Subscriber{}

	t.Run("Valid JSON Policy Document", func(t *testing.T) {
		fields := map[string]string{
			"ct.request.policydocument": `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`,
		}

		changes := sub.extractChanges("UpdateAssumeRolePolicy", fields)

		assert.Contains(t, changes, "assume_role_policy")
		policy, ok := changes["assume_role_policy"].(map[string]interface{})
		require.True(t, ok, "Policy should be a map")
		assert.Equal(t, "2012-10-17", policy["Version"])
	})

	t.Run("Invalid JSON Policy Document", func(t *testing.T) {
		fields := map[string]string{
			"ct.request.policydocument": `{invalid json}`,
		}

		changes := sub.extractChanges("UpdateAssumeRolePolicy", fields)

		// Invalid JSON should not be added to changes
		assert.NotContains(t, changes, "assume_role_policy")
	})
}
