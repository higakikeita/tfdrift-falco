package falco

// mapEventToResourceType maps a CloudTrail event name to a Terraform resource type
func (s *Subscriber) mapEventToResourceType(eventName string) string {
	mapping := map[string]string{
		// EC2
		"ModifyInstanceAttribute": "aws_instance",
		"ModifyVolume":            "aws_ebs_volume",

		// VPC - Security Groups
		"AuthorizeSecurityGroupIngress": "aws_security_group",
		"AuthorizeSecurityGroupEgress":  "aws_security_group",
		"RevokeSecurityGroupIngress":    "aws_security_group",
		"RevokeSecurityGroupEgress":     "aws_security_group",
		"CreateSecurityGroup":           "aws_security_group",
		"DeleteSecurityGroup":           "aws_security_group",
		"ModifySecurityGroupRules":      "aws_security_group_rule",

		// VPC - Core
		"CreateVpc":           "aws_vpc",
		"DeleteVpc":           "aws_vpc",
		"ModifyVpcAttribute":  "aws_vpc",
		"CreateSubnet":        "aws_subnet",
		"DeleteSubnet":        "aws_subnet",
		"ModifySubnetAttribute": "aws_subnet",

		// VPC - Route Tables
		"CreateRoute":         "aws_route",
		"DeleteRoute":         "aws_route",
		"ReplaceRoute":        "aws_route",
		"CreateRouteTable":    "aws_route_table",
		"DeleteRouteTable":    "aws_route_table",
		"AssociateRouteTable": "aws_route_table_association",

		// VPC - Gateways
		"AttachInternetGateway": "aws_internet_gateway_attachment",
		"DetachInternetGateway": "aws_internet_gateway_attachment",
		"CreateNatGateway":      "aws_nat_gateway",
		"DeleteNatGateway":      "aws_nat_gateway",

		// VPC - Network ACLs
		"CreateNetworkAcl":       "aws_network_acl",
		"DeleteNetworkAcl":       "aws_network_acl",
		"CreateNetworkAclEntry":  "aws_network_acl_rule",
		"DeleteNetworkAclEntry":  "aws_network_acl_rule",
		"ReplaceNetworkAclEntry": "aws_network_acl_rule",

		// VPC - Endpoints
		"CreateVpcEndpoint": "aws_vpc_endpoint",
		"DeleteVpcEndpoint": "aws_vpc_endpoint",
		"ModifyVpcEndpoint": "aws_vpc_endpoint",

		// ELB/ALB - Load Balancers
		"CreateLoadBalancer":           "aws_lb",
		"DeleteLoadBalancer":           "aws_lb",
		"ModifyLoadBalancerAttributes": "aws_lb",

		// ELB/ALB - Target Groups
		"CreateTargetGroup":           "aws_lb_target_group",
		"DeleteTargetGroup":           "aws_lb_target_group",
		"ModifyTargetGroup":           "aws_lb_target_group",
		"ModifyTargetGroupAttributes": "aws_lb_target_group",
		"RegisterTargets":             "aws_lb_target_group_attachment",
		"DeregisterTargets":           "aws_lb_target_group_attachment",

		// ELB/ALB - Listeners & Rules
		"CreateListener": "aws_lb_listener",
		"DeleteListener": "aws_lb_listener",
		"ModifyListener": "aws_lb_listener",
		"CreateRule":     "aws_lb_listener_rule",
		"DeleteRule":     "aws_lb_listener_rule",
		"ModifyRule":     "aws_lb_listener_rule",

		// KMS
		"ScheduleKeyDeletion": "aws_kms_key",
		"DisableKey":          "aws_kms_key",
		"EnableKey":           "aws_kms_key",
		"PutKeyPolicy":        "aws_kms_key",
		"CreateKey":           "aws_kms_key",
		"CreateAlias":         "aws_kms_alias",
		"DeleteAlias":         "aws_kms_alias",
		"UpdateAlias":         "aws_kms_alias",
		"EnableKeyRotation":   "aws_kms_key",
		"DisableKeyRotation":  "aws_kms_key",

		// DynamoDB
		"CreateTable":             "aws_dynamodb_table",
		"DeleteTable":             "aws_dynamodb_table",
		"UpdateTable":             "aws_dynamodb_table",
		"UpdateTimeToLive":        "aws_dynamodb_table",
		"UpdateContinuousBackups": "aws_dynamodb_table",

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
		"PutBucketPolicy":               "aws_s3_bucket_policy",
		"PutBucketVersioning":           "aws_s3_bucket",
		"PutBucketEncryption":           "aws_s3_bucket",
		"DeleteBucketEncryption":        "aws_s3_bucket",
		"PutBucketPublicAccessBlock":    "aws_s3_bucket_public_access_block",
		"DeleteBucketPublicAccessBlock": "aws_s3_bucket_public_access_block",
		"PutBucketAcl":                  "aws_s3_bucket_acl",

		// RDS
		"ModifyDBInstance": "aws_db_instance",

		// Lambda
		"UpdateFunctionConfiguration": "aws_lambda_function",
		"AddPermission":               "aws_lambda_permission",
		"RemovePermission":            "aws_lambda_permission",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
