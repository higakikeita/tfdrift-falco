package falco

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				"ct.request.rolename":                 "my-role",
				"ct.request.assumerolepolicydocument": `{"Version":"2012-10-17"}`,
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
				"ct.request.username":               "service-user",
				"ct.response.accesskey.accesskeyid": "AKIAIOSFODNN7EXAMPLE",
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
