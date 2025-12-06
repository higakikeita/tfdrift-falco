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

		// RDS - DB Instances
		"CreateDBInstance":           "aws_db_instance",
		"DeleteDBInstance":           "aws_db_instance",
		"ModifyDBInstance":           "aws_db_instance",
		"RebootDBInstance":           "aws_db_instance",
		"StartDBInstance":            "aws_db_instance",
		"StopDBInstance":             "aws_db_instance",
		"ModifyDBInstanceAttribute":  "aws_db_instance",

		// RDS - DB Clusters (Aurora)
		"CreateDBCluster":            "aws_rds_cluster",
		"DeleteDBCluster":            "aws_rds_cluster",
		"ModifyDBCluster":            "aws_rds_cluster",
		"StartDBCluster":             "aws_rds_cluster",
		"StopDBCluster":              "aws_rds_cluster",
		"FailoverDBCluster":          "aws_rds_cluster",
		"AddRoleToDBCluster":         "aws_rds_cluster_role_association",
		"RemoveRoleFromDBCluster":    "aws_rds_cluster_role_association",
		"ModifyDBClusterEndpoint":    "aws_rds_cluster_endpoint",
		"CreateDBClusterEndpoint":    "aws_rds_cluster_endpoint",
		"DeleteDBClusterEndpoint":    "aws_rds_cluster_endpoint",
		"ModifyGlobalCluster":        "aws_rds_global_cluster",

		// RDS - Snapshots
		"CreateDBSnapshot":           "aws_db_snapshot",
		"DeleteDBSnapshot":           "aws_db_snapshot",
		"ModifyDBSnapshotAttribute":  "aws_db_snapshot",
		"CreateDBClusterSnapshot":    "aws_db_cluster_snapshot",
		"DeleteDBClusterSnapshot":    "aws_db_cluster_snapshot",

		// RDS - Parameter Groups
		"CreateDBParameterGroup":     "aws_db_parameter_group",
		"DeleteDBParameterGroup":     "aws_db_parameter_group",
		"ModifyDBParameterGroup":     "aws_db_parameter_group",

		// RDS - Subnet Groups
		"CreateDBSubnetGroup":        "aws_db_subnet_group",
		"DeleteDBSubnetGroup":        "aws_db_subnet_group",
		"ModifyDBSubnetGroup":        "aws_db_subnet_group",

		// RDS - Restore
		"RestoreDBInstanceFromDBSnapshot": "aws_db_instance",
		"RestoreDBClusterFromSnapshot":    "aws_rds_cluster",

		// Lambda
		"UpdateFunctionConfiguration": "aws_lambda_function",
		"AddPermission":               "aws_lambda_permission",
		"RemovePermission":            "aws_lambda_permission",

		// API Gateway - REST API
		"CreateRestApi":               "aws_api_gateway_rest_api",
		"DeleteRestApi":               "aws_api_gateway_rest_api",
		"UpdateRestApi":               "aws_api_gateway_rest_api",
		"CreateResource":              "aws_api_gateway_resource",
		"DeleteResource":              "aws_api_gateway_resource",
		"CreateMethod":                "aws_api_gateway_method",
		"DeleteMethod":                "aws_api_gateway_method",
		"PutMethod":                   "aws_api_gateway_method",
		"UpdateMethod":                "aws_api_gateway_method",
		"CreateDeployment":            "aws_api_gateway_deployment",
		"DeleteDeployment":            "aws_api_gateway_deployment",
		"CreateStage":                 "aws_api_gateway_stage",
		"DeleteStage":                 "aws_api_gateway_stage",
		"UpdateStage":                 "aws_api_gateway_stage",

		// API Gateway - Authorizers & Models
		"CreateAuthorizer":            "aws_api_gateway_authorizer",
		"DeleteAuthorizer":            "aws_api_gateway_authorizer",
		"UpdateAuthorizer":            "aws_api_gateway_authorizer",
		"CreateModel":                 "aws_api_gateway_model",
		"DeleteModel":                 "aws_api_gateway_model",

		// API Gateway - API Keys & Usage Plans
		"CreateApiKey":                "aws_api_gateway_api_key",
		"DeleteApiKey":                "aws_api_gateway_api_key",
		"UpdateApiKey":                "aws_api_gateway_api_key",
		"CreateUsagePlan":             "aws_api_gateway_usage_plan",
		"DeleteUsagePlan":             "aws_api_gateway_usage_plan",
		"UpdateUsagePlan":             "aws_api_gateway_usage_plan",

		// API Gateway v2 (HTTP/WebSocket)
		"CreateApi":                   "aws_apigatewayv2_api",
		"DeleteApi":                   "aws_apigatewayv2_api",
		"UpdateApi":                   "aws_apigatewayv2_api",
		"CreateRoute":                 "aws_apigatewayv2_route",
		"DeleteRoute":                 "aws_apigatewayv2_route",
		"UpdateRoute":                 "aws_apigatewayv2_route",
		"CreateIntegration":           "aws_apigatewayv2_integration",
		"DeleteIntegration":           "aws_apigatewayv2_integration",
		"UpdateIntegration":           "aws_apigatewayv2_integration",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
