package falco

// mapEventToResourceType maps a CloudTrail event name to a Terraform resource type
func (s *Subscriber) mapEventToResourceType(eventName string) string {
	mapping := map[string]string{
		// EC2 - Instance Management
		"RunInstances":            "aws_instance",
		"TerminateInstances":      "aws_instance",
		"StartInstances":          "aws_instance",
		"StopInstances":           "aws_instance",
		"ModifyInstanceAttribute": "aws_instance",

		// EC2 - AMI Management
		"CreateImage":      "aws_ami",
		"DeregisterImage":  "aws_ami",

		// EC2 - EBS Volume Management
		"CreateVolume":  "aws_ebs_volume",
		"DeleteVolume":  "aws_ebs_volume",
		"AttachVolume":  "aws_volume_attachment",
		"DetachVolume":  "aws_volume_attachment",
		"ModifyVolume":  "aws_ebs_volume",

		// EC2 - Snapshot Management
		"CreateSnapshot": "aws_ebs_snapshot",
		"DeleteSnapshot": "aws_ebs_snapshot",

		// EC2 - Network Interface Management
		"CreateNetworkInterface": "aws_network_interface",
		"DeleteNetworkInterface": "aws_network_interface",
		"AttachNetworkInterface": "aws_network_interface_attachment",

		// VPC - Security Groups
		"AuthorizeSecurityGroupIngress": "aws_security_group",
		"AuthorizeSecurityGroupEgress":  "aws_security_group",
		"RevokeSecurityGroupIngress":    "aws_security_group",
		"RevokeSecurityGroupEgress":     "aws_security_group",
		"CreateSecurityGroup":           "aws_security_group",
		"DeleteSecurityGroup":           "aws_security_group",
		"ModifySecurityGroupRules":      "aws_security_group_rule",

		// VPC - Core
		"CreateVpc":             "aws_vpc",
		"DeleteVpc":             "aws_vpc",
		"ModifyVpcAttribute":    "aws_vpc",
		"CreateSubnet":          "aws_subnet",
		"DeleteSubnet":          "aws_subnet",
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
		"CreateDBInstance":          "aws_db_instance",
		"DeleteDBInstance":          "aws_db_instance",
		"ModifyDBInstance":          "aws_db_instance",
		"RebootDBInstance":          "aws_db_instance",
		"StartDBInstance":           "aws_db_instance",
		"StopDBInstance":            "aws_db_instance",
		"ModifyDBInstanceAttribute": "aws_db_instance",

		// RDS - DB Clusters (Aurora)
		"CreateDBCluster":         "aws_rds_cluster",
		"DeleteDBCluster":         "aws_rds_cluster",
		"ModifyDBCluster":         "aws_rds_cluster",
		"StartDBCluster":          "aws_rds_cluster",
		"StopDBCluster":           "aws_rds_cluster",
		"FailoverDBCluster":       "aws_rds_cluster",
		"AddRoleToDBCluster":      "aws_rds_cluster_role_association",
		"RemoveRoleFromDBCluster": "aws_rds_cluster_role_association",
		"ModifyDBClusterEndpoint": "aws_rds_cluster_endpoint",
		"CreateDBClusterEndpoint": "aws_rds_cluster_endpoint",
		"DeleteDBClusterEndpoint": "aws_rds_cluster_endpoint",
		"ModifyGlobalCluster":     "aws_rds_global_cluster",

		// RDS - Snapshots
		"CreateDBSnapshot":          "aws_db_snapshot",
		"DeleteDBSnapshot":          "aws_db_snapshot",
		"ModifyDBSnapshotAttribute": "aws_db_snapshot",
		"CreateDBClusterSnapshot":   "aws_db_cluster_snapshot",
		"DeleteDBClusterSnapshot":   "aws_db_cluster_snapshot",

		// RDS - Parameter Groups
		"CreateDBParameterGroup": "aws_db_parameter_group",
		"DeleteDBParameterGroup": "aws_db_parameter_group",
		"ModifyDBParameterGroup": "aws_db_parameter_group",

		// RDS - Subnet Groups
		"CreateDBSubnetGroup": "aws_db_subnet_group",
		"DeleteDBSubnetGroup": "aws_db_subnet_group",
		"ModifyDBSubnetGroup": "aws_db_subnet_group",

		// RDS - Restore
		"RestoreDBInstanceFromDBSnapshot": "aws_db_instance",
		"RestoreDBClusterFromSnapshot":    "aws_rds_cluster",

		// Lambda - Function Management
		"CreateFunction":              "aws_lambda_function",
		"DeleteFunction":              "aws_lambda_function",
		"UpdateFunctionCode":          "aws_lambda_function",
		"UpdateFunctionConfiguration": "aws_lambda_function",

		// Lambda - Permissions
		"AddPermission":    "aws_lambda_permission",
		"RemovePermission": "aws_lambda_permission",

		// Lambda - Event Source Mappings
		"CreateEventSourceMapping": "aws_lambda_event_source_mapping",
		"DeleteEventSourceMapping": "aws_lambda_event_source_mapping",
		"UpdateEventSourceMapping": "aws_lambda_event_source_mapping",

		// Lambda - Concurrency
		"PutFunctionConcurrency": "aws_lambda_function",

		// Note: CreateAlias, DeleteAlias, UpdateAlias are mapped to aws_kms_alias (line 80-82)
		// Lambda aliases cannot be distinguished from KMS aliases without eventSource field

		// API Gateway - REST API
		"CreateRestApi":    "aws_api_gateway_rest_api",
		"DeleteRestApi":    "aws_api_gateway_rest_api",
		"UpdateRestApi":    "aws_api_gateway_rest_api",
		"CreateResource":   "aws_api_gateway_resource",
		"DeleteResource":   "aws_api_gateway_resource",
		"CreateMethod":     "aws_api_gateway_method",
		"DeleteMethod":     "aws_api_gateway_method",
		"PutMethod":        "aws_api_gateway_method",
		"UpdateMethod":     "aws_api_gateway_method",
		"CreateDeployment": "aws_api_gateway_deployment",
		"DeleteDeployment": "aws_api_gateway_deployment",
		"CreateStage":      "aws_api_gateway_stage",
		"DeleteStage":      "aws_api_gateway_stage",
		"UpdateStage":      "aws_api_gateway_stage",

		// API Gateway - Authorizers & Models
		"CreateAuthorizer": "aws_api_gateway_authorizer",
		"DeleteAuthorizer": "aws_api_gateway_authorizer",
		"UpdateAuthorizer": "aws_api_gateway_authorizer",
		"CreateModel":      "aws_api_gateway_model",
		"DeleteModel":      "aws_api_gateway_model",

		// API Gateway - API Keys & Usage Plans
		"CreateApiKey":    "aws_api_gateway_api_key",
		"DeleteApiKey":    "aws_api_gateway_api_key",
		"UpdateApiKey":    "aws_api_gateway_api_key",
		"CreateUsagePlan": "aws_api_gateway_usage_plan",
		"DeleteUsagePlan": "aws_api_gateway_usage_plan",
		"UpdateUsagePlan": "aws_api_gateway_usage_plan",

		// API Gateway v2 (HTTP/WebSocket)
		"CreateApi": "aws_apigatewayv2_api",
		"DeleteApi": "aws_apigatewayv2_api",
		"UpdateApi": "aws_apigatewayv2_api",
		// Note: CreateRoute/DeleteRoute mapped to aws_route (VPC) above
		"UpdateRoute":       "aws_apigatewayv2_route",
		"CreateIntegration": "aws_apigatewayv2_integration",
		"DeleteIntegration": "aws_apigatewayv2_integration",
		"UpdateIntegration": "aws_apigatewayv2_integration",

		// CloudWatch - Alarms
		"PutMetricAlarm":      "aws_cloudwatch_metric_alarm",
		"DeleteAlarms":        "aws_cloudwatch_metric_alarm",
		"DisableAlarmActions": "aws_cloudwatch_metric_alarm",
		"EnableAlarmActions":  "aws_cloudwatch_metric_alarm",
		"SetAlarmState":       "aws_cloudwatch_metric_alarm",

		// CloudWatch - Logs
		"CreateLogGroup":        "aws_cloudwatch_log_group",
		"DeleteLogGroup":        "aws_cloudwatch_log_group",
		"PutRetentionPolicy":    "aws_cloudwatch_log_group",
		"DeleteRetentionPolicy": "aws_cloudwatch_log_group",
		"AssociateKmsKey":       "aws_cloudwatch_log_group",
		"DisassociateKmsKey":    "aws_cloudwatch_log_group",
		"PutMetricFilter":       "aws_cloudwatch_log_metric_filter",
		"DeleteMetricFilter":    "aws_cloudwatch_log_metric_filter",
		"CreateLogStream":       "aws_cloudwatch_log_stream",
		"DeleteLogStream":       "aws_cloudwatch_log_stream",
		"PutDashboard":          "aws_cloudwatch_dashboard",
		"DeleteDashboards":      "aws_cloudwatch_dashboard",

		// SNS
		"CreateTopic":         "aws_sns_topic",
		"DeleteTopic":         "aws_sns_topic",
		"SetTopicAttributes":  "aws_sns_topic",
		"Subscribe":           "aws_sns_topic_subscription",
		"Unsubscribe":         "aws_sns_topic_subscription",
		"ConfirmSubscription": "aws_sns_topic_subscription",

		// SQS
		"CreateQueue":        "aws_sqs_queue",
		"DeleteQueue":        "aws_sqs_queue",
		"SetQueueAttributes": "aws_sqs_queue",
		"PurgeQueue":         "aws_sqs_queue",

		// Route53
		"ChangeResourceRecordSets":      "aws_route53_record",
		"CreateHostedZone":              "aws_route53_zone",
		"DeleteHostedZone":              "aws_route53_zone",
		"ChangeTagsForResource":         "aws_route53_zone",
		"AssociateVPCWithHostedZone":    "aws_route53_zone_association",
		"DisassociateVPCFromHostedZone": "aws_route53_zone_association",

		// ECR
		"PutImageScanningConfiguration": "aws_ecr_repository",
		"PutImageTagMutability":         "aws_ecr_repository",
		"PutLifecyclePolicy":            "aws_ecr_lifecycle_policy",
		"DeleteLifecyclePolicy":         "aws_ecr_lifecycle_policy",
		"SetRepositoryPolicy":           "aws_ecr_repository_policy",
		"DeleteRepositoryPolicy":        "aws_ecr_repository_policy",
		"CreateRepository":              "aws_ecr_repository",
		"DeleteRepository":              "aws_ecr_repository",
		"PutReplicationConfiguration":   "aws_ecr_replication_configuration",

		// SSM Parameter Store
		"PutParameter":          "aws_ssm_parameter",
		"DeleteParameter":       "aws_ssm_parameter",
		"DeleteParameters":      "aws_ssm_parameter",
		"LabelParameterVersion": "aws_ssm_parameter",

		// Secrets Manager
		"CreateSecret":             "aws_secretsmanager_secret",
		"DeleteSecret":             "aws_secretsmanager_secret",
		"UpdateSecret":             "aws_secretsmanager_secret",
		"PutSecretValue":           "aws_secretsmanager_secret_version",
		"RotateSecret":             "aws_secretsmanager_secret_rotation",
		"CancelRotateSecret":       "aws_secretsmanager_secret_rotation",
		"UpdateSecretVersionStage": "aws_secretsmanager_secret_version",
		"PutResourcePolicy":        "aws_secretsmanager_secret_policy",
		"DeleteResourcePolicy":     "aws_secretsmanager_secret_policy",

		// CloudFront
		"CreateDistribution": "aws_cloudfront_distribution",
		"DeleteDistribution": "aws_cloudfront_distribution",
		"UpdateDistribution": "aws_cloudfront_distribution",
		"CreateInvalidation": "aws_cloudfront_invalidation",

		// CloudTrail
		"CreateTrail":         "aws_cloudtrail",
		"DeleteTrail":         "aws_cloudtrail",
		"UpdateTrail":         "aws_cloudtrail",
		"StartLogging":        "aws_cloudtrail",
		"StopLogging":         "aws_cloudtrail",
		"PutEventSelectors":   "aws_cloudtrail_event_data_store",
		"PutInsightSelectors": "aws_cloudtrail",

		// ECS - Services
		"CreateService": "aws_ecs_service",
		"UpdateService": "aws_ecs_service",
		"DeleteService": "aws_ecs_service",

		// ECS - Task Definitions
		"RegisterTaskDefinition":   "aws_ecs_task_definition",
		"DeregisterTaskDefinition": "aws_ecs_task_definition",

		// ECS - Clusters
		// Note: CreateCluster/DeleteCluster can be ECS, EKS, or Redshift - context-dependent
		// For ECS clusters, use UpdateCluster, UpdateClusterSettings instead
		"UpdateCluster":                 "aws_ecs_cluster",
		"UpdateClusterSettings":         "aws_ecs_cluster",
		"PutClusterCapacityProviders":   "aws_ecs_cluster_capacity_providers",
		"UpdateContainerInstancesState": "aws_ecs_container_instance",

		// ECS - Capacity Providers
		"CreateCapacityProvider": "aws_ecs_capacity_provider",
		"UpdateCapacityProvider": "aws_ecs_capacity_provider",
		"DeleteCapacityProvider": "aws_ecs_capacity_provider",

		// EKS - Clusters
		"CreateCluster":        "aws_eks_cluster",
		"DeleteCluster":        "aws_eks_cluster",
		"UpdateClusterConfig":  "aws_eks_cluster",
		"UpdateClusterVersion": "aws_eks_cluster",

		// EKS - Node Groups
		"CreateNodegroup":        "aws_eks_node_group",
		"DeleteNodegroup":        "aws_eks_node_group",
		"UpdateNodegroupConfig":  "aws_eks_node_group",
		"UpdateNodegroupVersion": "aws_eks_node_group",

		// EKS - Addons
		"CreateAddon": "aws_eks_addon",
		"DeleteAddon": "aws_eks_addon",
		"UpdateAddon": "aws_eks_addon",

		// EKS - Fargate Profiles
		"CreateFargateProfile": "aws_eks_fargate_profile",

		// Redshift
		"ModifyCluster":               "aws_redshift_cluster",
		"ModifyClusterParameterGroup": "aws_redshift_parameter_group",
		"CreateClusterParameterGroup": "aws_redshift_parameter_group",
		"DeleteClusterParameterGroup": "aws_redshift_parameter_group",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
