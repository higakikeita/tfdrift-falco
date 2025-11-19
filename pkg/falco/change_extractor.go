package falco

import (
	"encoding/json"
)

// extractChanges extracts the changed attributes from Falco output
func (s *Subscriber) extractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	switch eventName {
	case "ModifyInstanceAttribute":
		if val, ok := fields["ct.request.disableapitermination"]; ok && val != "" {
			changes["disable_api_termination"] = val
		}
		if val, ok := fields["ct.request.instancetype"]; ok && val != "" {
			changes["instance_type"] = val
		}

	case "PutBucketEncryption":
		// Encryption enabled
		if config, ok := fields["ct.request.serversideencryptionconfiguration"]; ok && config != "" {
			changes["server_side_encryption_configuration"] = config
		}

	case "DeleteBucketEncryption":
		// Encryption disabled
		changes["server_side_encryption_configuration"] = nil

	case "UpdateFunctionConfiguration":
		if val, ok := fields["ct.request.timeout"]; ok && val != "" {
			changes["timeout"] = val
		}
		if val, ok := fields["ct.request.memorysize"]; ok && val != "" {
			changes["memory_size"] = val
		}

	case "UpdateAssumeRolePolicy":
		if policy := getStringField(fields, "ct.request.policydocument"); policy != "" {
			var policyDoc map[string]interface{}
			if err := json.Unmarshal([]byte(policy), &policyDoc); err == nil {
				changes["assume_role_policy"] = policyDoc
			}
		}

	// IAM Policy attachments
	case "AttachRolePolicy", "AttachUserPolicy", "AttachGroupPolicy":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["attached_policy_arn"] = policyArn
		}

	// IAM Inline policies
	case "PutRolePolicy", "PutUserPolicy", "PutGroupPolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["inline_policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy creation
	case "CreatePolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy version
	case "CreatePolicyVersion":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["policy_arn"] = policyArn
		}
		if setDefault, ok := fields["ct.request.setasdefault"]; ok && setDefault != "" {
			changes["set_as_default"] = setDefault
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Role creation
	case "CreateRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["role_name"] = roleName
		}
		if assumePolicy := getStringField(fields, "ct.request.assumerolepolicydocument"); assumePolicy != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(assumePolicy), &doc); err == nil {
				changes["assume_role_policy"] = doc
			}
		}

	// IAM User/Role deletion
	case "DeleteRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["deleted_role"] = roleName
		}
	case "DeleteUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["deleted_user"] = userName
		}

	// IAM User creation
	case "CreateUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}

	// IAM Access Key creation
	case "CreateAccessKey":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if accessKeyId := getStringField(fields, "ct.response.accesskey.accesskeyid"); accessKeyId != "" {
			changes["access_key_id"] = accessKeyId
		}

	// IAM Group membership
	case "AddUserToGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}
	case "RemoveUserFromGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}

	// Account password policy
	case "UpdateAccountPasswordPolicy":
		// Extract various password policy fields if available
		if minLength, ok := fields["ct.request.minimumpasswordlength"]; ok && minLength != "" {
			changes["minimum_password_length"] = minLength
		}
		if requireSymbols, ok := fields["ct.request.requiresymbols"]; ok && requireSymbols != "" {
			changes["require_symbols"] = requireSymbols
		}
	}

	return changes
}
