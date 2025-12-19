package mappings

// SecurityMappings contains CloudTrail event to Terraform resource mappings for security services
var SecurityMappings = map[string]string{
	// IAM - Policy modifications
	"PutUserPolicy":          "aws_iam_user_policy",
	"PutRolePolicy":          "aws_iam_role_policy",
	"PutGroupPolicy":         "aws_iam_group_policy",
	"UpdateAssumeRolePolicy": "aws_iam_role",
	"AttachUserPolicy":       "aws_iam_user_policy_attachment",
	"AttachRolePolicy":       "aws_iam_role_policy_attachment",
	"AttachGroupPolicy":      "aws_iam_group_policy_attachment",
	"CreatePolicy":           "aws_iam_policy",
	"CreatePolicyVersion":    "aws_iam_policy",

	// IAM - User/Role/Group lifecycle
	"CreateRole":                  "aws_iam_role",
	"DeleteRole":                  "aws_iam_role",
	"UpdateRole":                  "aws_iam_role",
	"TagRole":                     "aws_iam_role",
	"UntagRole":                   "aws_iam_role",
	"CreateUser":                  "aws_iam_user",
	"DeleteUser":                  "aws_iam_user",
	"UpdateUser":                  "aws_iam_user",
	"CreateGroup":                 "aws_iam_group",
	"DeleteGroup":                 "aws_iam_group",
	"UpdateGroup":                 "aws_iam_group",
	"CreateAccessKey":             "aws_iam_access_key",
	"AddUserToGroup":              "aws_iam_user_group_membership",
	"RemoveUserFromGroup":         "aws_iam_user_group_membership",
	"UpdateAccountPasswordPolicy": "aws_iam_account_password_policy",

	// IAM - Instance Profiles
	"CreateInstanceProfile":         "aws_iam_instance_profile",
	"DeleteInstanceProfile":         "aws_iam_instance_profile",
	"AddRoleToInstanceProfile":      "aws_iam_instance_profile",
	"RemoveRoleFromInstanceProfile": "aws_iam_instance_profile",

	// IAM - OIDC Providers
	"AddClientIDToOpenIDConnectProvider":    "aws_iam_openid_connect_provider",
	"RemoveClientIDFromOpenIDConnectProvider": "aws_iam_openid_connect_provider",

	// KMS - Key Management
	"CreateKey":           "aws_kms_key",
	"ScheduleKeyDeletion": "aws_kms_key",
	"DisableKey":          "aws_kms_key",
	"EnableKey":           "aws_kms_key",
	"PutKeyPolicy":        "aws_kms_key",
	"EnableKeyRotation":   "aws_kms_key",
	"DisableKeyRotation":  "aws_kms_key",

	// ACM - Certificate Management
	"RequestCertificate":        "aws_acm_certificate",
	"DeleteCertificate":         "aws_acm_certificate",
	"AddTagsToCertificate":      "aws_acm_certificate",
	"RemoveTagsFromCertificate": "aws_acm_certificate",
	"ImportCertificate":         "aws_acm_certificate",
	"ExportCertificate":         "aws_acm_certificate",

	// WAF / WAFv2 - Web ACLs
	"CreateWebACL": "aws_wafv2_web_acl",
	"UpdateWebACL": "aws_wafv2_web_acl",
	"DeleteWebACL": "aws_wafv2_web_acl",

	// WAF - Rule Groups
	"CreateRuleGroup": "aws_wafv2_rule_group",
	"UpdateRuleGroup": "aws_wafv2_rule_group",
	"DeleteRuleGroup": "aws_wafv2_rule_group",

	// WAF - IP Sets
	"CreateIPSet": "aws_wafv2_ip_set",
	"UpdateIPSet": "aws_wafv2_ip_set",
	"DeleteIPSet": "aws_wafv2_ip_set",

	// WAF - Regex Pattern Sets
	"CreateRegexPatternSet": "aws_wafv2_regex_pattern_set",
	"UpdateRegexPatternSet": "aws_wafv2_regex_pattern_set",
	"DeleteRegexPatternSet": "aws_wafv2_regex_pattern_set",

	// WAF - Associations
	"AssociateWebACL":    "aws_wafv2_web_acl_association",
	"DisassociateWebACL": "aws_wafv2_web_acl_association",

	// Secrets Manager - Secrets
	"CreateSecret":             "aws_secretsmanager_secret",
	"DeleteSecret":             "aws_secretsmanager_secret",
	"UpdateSecret":             "aws_secretsmanager_secret",
	"PutSecretValue":           "aws_secretsmanager_secret_version",
	"RotateSecret":             "aws_secretsmanager_secret_rotation",
	"CancelRotateSecret":       "aws_secretsmanager_secret_rotation",
	"UpdateSecretVersionStage": "aws_secretsmanager_secret_version",
	"PutResourcePolicy":        "aws_secretsmanager_secret_policy",
	"DeleteResourcePolicy":     "aws_secretsmanager_secret_policy",

	// SSM Parameter Store - Parameters
	"PutParameter":          "aws_ssm_parameter",
	"DeleteParameter":       "aws_ssm_parameter",
	"DeleteParameters":      "aws_ssm_parameter",
	"LabelParameterVersion": "aws_ssm_parameter",

	// CloudTrail - Trails
	"CreateTrail":         "aws_cloudtrail",
	"DeleteTrail":         "aws_cloudtrail",
	"UpdateTrail":         "aws_cloudtrail",
	"StartLogging":        "aws_cloudtrail",
	"StopLogging":         "aws_cloudtrail",
	"PutEventSelectors":   "aws_cloudtrail",
	"PutInsightSelectors": "aws_cloudtrail",
}
