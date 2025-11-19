package falco

// mapEventToResourceType maps a CloudTrail event name to a Terraform resource type
func (s *Subscriber) mapEventToResourceType(eventName string) string {
	mapping := map[string]string{
		// EC2
		"ModifyInstanceAttribute": "aws_instance",
		"ModifyVolume":            "aws_ebs_volume",

		// IAM - Roles
		"PutRolePolicy":          "aws_iam_role_policy",
		"UpdateAssumeRolePolicy": "aws_iam_role",
		"AttachRolePolicy":       "aws_iam_role_policy_attachment",
		"CreateRole":             "aws_iam_role",
		"DeleteRole":             "aws_iam_role",

		// IAM - Users
		"PutUserPolicy":       "aws_iam_user_policy",
		"AttachUserPolicy":    "aws_iam_user_policy_attachment",
		"CreateUser":          "aws_iam_user",
		"DeleteUser":          "aws_iam_user",
		"CreateAccessKey":     "aws_iam_access_key",
		"AddUserToGroup":      "aws_iam_user_group_membership",
		"RemoveUserFromGroup": "aws_iam_user_group_membership",

		// IAM - Groups
		"PutGroupPolicy":    "aws_iam_group_policy",
		"AttachGroupPolicy": "aws_iam_group_policy_attachment",

		// IAM - Policies
		"CreatePolicy":        "aws_iam_policy",
		"CreatePolicyVersion": "aws_iam_policy",

		// IAM - Account
		"UpdateAccountPasswordPolicy": "aws_iam_account_password_policy",

		// S3
		"PutBucketPolicy":        "aws_s3_bucket_policy",
		"PutBucketVersioning":    "aws_s3_bucket",
		"PutBucketEncryption":    "aws_s3_bucket",
		"DeleteBucketEncryption": "aws_s3_bucket",

		// RDS
		"ModifyDBInstance": "aws_db_instance",

		// Lambda
		"UpdateFunctionConfiguration": "aws_lambda_function",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
