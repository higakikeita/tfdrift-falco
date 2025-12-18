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

		// VPC - Peering
		"CreateVpcPeeringConnection": "aws_vpc_peering_connection",
		"AcceptVpcPeeringConnection": "aws_vpc_peering_connection_accepter",
		"DeleteVpcPeeringConnection": "aws_vpc_peering_connection",

		// VPC - Transit Gateway
		"CreateTransitGateway":           "aws_ec2_transit_gateway",
		"DeleteTransitGateway":           "aws_ec2_transit_gateway",
		"CreateTransitGatewayVpcAttachment": "aws_ec2_transit_gateway_vpc_attachment",

		// VPC - Flow Logs
		"CreateFlowLogs": "aws_flow_log",
		"DeleteFlowLogs": "aws_flow_log",

		// VPC - Network Firewall
		"DeleteFirewall": "aws_networkfirewall_firewall",

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
		"EnableKeyRotation":   "aws_kms_key",
		"DisableKeyRotation":  "aws_kms_key",
		// Note: CreateAlias, DeleteAlias, UpdateAlias are mapped to Lambda aliases
		// KMS aliases currently not supported - requires eventSource field to distinguish

		// DynamoDB - Tables
		"CreateTable":  "aws_dynamodb_table",
		"DeleteTable":  "aws_dynamodb_table",
		"UpdateTable":  "aws_dynamodb_table",
		"UpdateTimeToLive": "aws_dynamodb_table",

		// DynamoDB - Point-in-Time Recovery
		"UpdateContinuousBackups":   "aws_dynamodb_table",
		"RestoreTableToPointInTime": "aws_dynamodb_table",

		// DynamoDB - Backups
		"CreateBackup":           "aws_dynamodb_table_backup",
		"DeleteBackup":           "aws_dynamodb_table_backup",
		"RestoreTableFromBackup": "aws_dynamodb_table",

		// DynamoDB - Global Tables
		"CreateGlobalTable": "aws_dynamodb_global_table",
		"UpdateGlobalTable": "aws_dynamodb_global_table",

		// DynamoDB - Streams
		"EnableKinesisStreamingDestination":  "aws_dynamodb_kinesis_streaming_destination",
		"DisableKinesisStreamingDestination": "aws_dynamodb_kinesis_streaming_destination",

		// DynamoDB - Monitoring
		"UpdateContributorInsights": "aws_dynamodb_contributor_insights",

		// IAM - Roles
		"PutRolePolicy":          "aws_iam_role_policy",
		"UpdateAssumeRolePolicy": "aws_iam_role",
		"AttachRolePolicy":       "aws_iam_role_policy_attachment",
		"CreateRole":             "aws_iam_role",
		"DeleteRole":             "aws_iam_role",
		"UpdateRole":             "aws_iam_role",
		"TagRole":                "aws_iam_role",
		"UntagRole":              "aws_iam_role",

		// IAM - Users
		"PutUserPolicy":       "aws_iam_user_policy",
		"AttachUserPolicy":    "aws_iam_user_policy_attachment",
		"CreateUser":          "aws_iam_user",
		"DeleteUser":          "aws_iam_user",
		"UpdateUser":          "aws_iam_user",
		"CreateAccessKey":     "aws_iam_access_key",
		"AddUserToGroup":      "aws_iam_user_group_membership",
		"RemoveUserFromGroup": "aws_iam_user_group_membership",

		// IAM - Groups
		"PutGroupPolicy":    "aws_iam_group_policy",
		"AttachGroupPolicy": "aws_iam_group_policy_attachment",
		"UpdateGroup":       "aws_iam_group",

		// IAM - Instance Profiles
		"CreateInstanceProfile":    "aws_iam_instance_profile",
		"DeleteInstanceProfile":    "aws_iam_instance_profile",
		"AddRoleToInstanceProfile": "aws_iam_instance_profile",

		// IAM - Policies
		"CreatePolicy":        "aws_iam_policy",
		"CreatePolicyVersion": "aws_iam_policy",

		// IAM - Account
		"UpdateAccountPasswordPolicy": "aws_iam_account_password_policy",

		// S3 - Bucket Management
		"CreateBucket": "aws_s3_bucket",
		"DeleteBucket": "aws_s3_bucket",

		// S3 - Bucket Configuration
		"PutBucketPolicy":               "aws_s3_bucket_policy",
		"PutBucketVersioning":           "aws_s3_bucket",
		"PutBucketEncryption":           "aws_s3_bucket",
		"DeleteBucketEncryption":        "aws_s3_bucket",
		"PutBucketPublicAccessBlock":    "aws_s3_bucket_public_access_block",
		"DeleteBucketPublicAccessBlock": "aws_s3_bucket_public_access_block",
		"PutBucketAcl":                  "aws_s3_bucket_acl",
		"PutBucketTagging":              "aws_s3_bucket",

		// S3 - Lifecycle Management
		"PutBucketLifecycle":    "aws_s3_bucket_lifecycle_configuration",
		"DeleteBucketLifecycle": "aws_s3_bucket_lifecycle_configuration",

		// S3 - Replication
		"PutBucketReplication":    "aws_s3_bucket_replication_configuration",
		"DeleteBucketReplication": "aws_s3_bucket_replication_configuration",

		// S3 - Logging & Notifications
		"PutBucketLogging":       "aws_s3_bucket_logging",
		"PutBucketNotification":  "aws_s3_bucket_notification",

		// S3 - Website Configuration
		"PutBucketWebsite":    "aws_s3_bucket_website_configuration",
		"DeleteBucketWebsite": "aws_s3_bucket_website_configuration",

		// S3 - CORS Configuration
		"PutBucketCors":    "aws_s3_bucket_cors_configuration",
		"DeleteBucketCors": "aws_s3_bucket_cors_configuration",

		// S3 - Object Management
		"PutObjectAcl": "aws_s3_object",

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
		"RestoreDBInstanceToPointInTime":  "aws_db_instance",
		"RestoreDBClusterFromSnapshot":    "aws_rds_cluster",

		// RDS - Read Replicas
		"CreateDBInstanceReadReplica": "aws_db_instance",

		// RDS - Option Groups
		"CreateOptionGroup": "aws_db_option_group",
		"DeleteOptionGroup": "aws_db_option_group",
		"ModifyOptionGroup": "aws_db_option_group",

		// Lambda - Function Management
		"CreateFunction":              "aws_lambda_function",
		"DeleteFunction":              "aws_lambda_function",
		"UpdateFunctionCode":          "aws_lambda_function",
		"UpdateFunctionConfiguration": "aws_lambda_function",
		"PublishVersion":              "aws_lambda_function",

		// Lambda - Permissions
		"AddPermission":    "aws_lambda_permission",
		"RemovePermission": "aws_lambda_permission",

		// Lambda - Event Source Mappings
		"CreateEventSourceMapping": "aws_lambda_event_source_mapping",
		"DeleteEventSourceMapping": "aws_lambda_event_source_mapping",
		"UpdateEventSourceMapping": "aws_lambda_event_source_mapping",

		// Lambda - Concurrency & Invoke Config
		"PutFunctionConcurrency":         "aws_lambda_function",
		"PutProvisionedConcurrencyConfig": "aws_lambda_provisioned_concurrency_config",
		"PutFunctionEventInvokeConfig":   "aws_lambda_function_event_invoke_config",

		// Lambda - Aliases
		// Note: CreateAlias, DeleteAlias, UpdateAlias conflict with KMS aliases
		// In production, distinguish using eventSource field (lambda.amazonaws.com vs kms.amazonaws.com)
		"CreateAlias": "aws_lambda_alias",
		"DeleteAlias": "aws_lambda_alias",
		"UpdateAlias": "aws_lambda_alias",

		// Auto Scaling - Auto Scaling Groups
		"CreateAutoScalingGroup": "aws_autoscaling_group",
		"DeleteAutoScalingGroup": "aws_autoscaling_group",
		"UpdateAutoScalingGroup": "aws_autoscaling_group",
		"SetDesiredCapacity":     "aws_autoscaling_group",

		// Auto Scaling - Launch Configurations
		"CreateLaunchConfiguration": "aws_launch_configuration",
		"DeleteLaunchConfiguration": "aws_launch_configuration",

		// Auto Scaling - Scaling Policies
		"PutScalingPolicy": "aws_autoscaling_policy",
		"DeletePolicy":     "aws_autoscaling_policy",

		// Auto Scaling - Scheduled Actions
		"PutScheduledUpdateGroupAction": "aws_autoscaling_schedule",
		"DeleteScheduledAction":         "aws_autoscaling_schedule",

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

		// CloudFormation - Stacks
		"CreateStack":                  "aws_cloudformation_stack",
		"DeleteStack":                  "aws_cloudformation_stack",
		"UpdateStack":                  "aws_cloudformation_stack",
		"CancelUpdateStack":            "aws_cloudformation_stack",
		"ContinueUpdateRollback":       "aws_cloudformation_stack",
		"SetStackPolicy":               "aws_cloudformation_stack",
		"UpdateTerminationProtection":  "aws_cloudformation_stack",

		// CloudFormation - Change Sets
		"CreateChangeSet":   "aws_cloudformation_change_set",
		"DeleteChangeSet":   "aws_cloudformation_change_set",
		"ExecuteChangeSet":  "aws_cloudformation_change_set",

		// EventBridge (CloudWatch Events) - Rules
		"PutRule":     "aws_cloudwatch_event_rule",
		"DeleteRule":  "aws_cloudwatch_event_rule",
		"EnableRule":  "aws_cloudwatch_event_rule",
		"DisableRule": "aws_cloudwatch_event_rule",

		// EventBridge - Targets
		"PutTargets":    "aws_cloudwatch_event_target",
		"RemoveTargets": "aws_cloudwatch_event_target",

		// EventBridge - Events
		"PutEvents":        "aws_cloudwatch_event_bus",
		"PutPartnerEvents": "aws_cloudwatch_event_bus",

		// Step Functions - State Machines
		"CreateStateMachine": "aws_sfn_state_machine",
		"DeleteStateMachine": "aws_sfn_state_machine",
		"UpdateStateMachine": "aws_sfn_state_machine",

		// Step Functions - Executions
		"StartExecution": "aws_sfn_state_machine",
		"StopExecution":  "aws_sfn_state_machine",

		// Step Functions - Tags
		// Note: TagResource, UntagResource are generic AWS API operations
		// For Step Functions: maps to aws_sfn_state_machine
		"TagResource":   "aws_sfn_state_machine",
		"UntagResource": "aws_sfn_state_machine",

		// AWS Glue - Databases
		"CreateDatabase": "aws_glue_catalog_database",
		"DeleteDatabase": "aws_glue_catalog_database",
		"UpdateDatabase": "aws_glue_catalog_database",

		// AWS Glue - Tables
		"CreateTable": "aws_glue_catalog_table",
		"DeleteTable": "aws_glue_catalog_table",
		"UpdateTable": "aws_glue_catalog_table",

		// AWS Glue - Jobs
		"CreateJob":   "aws_glue_job",
		"DeleteJob":   "aws_glue_job",
		"UpdateJob":   "aws_glue_job",
		"StartJobRun": "aws_glue_job",
		"StopJobRun":  "aws_glue_job",

		// Kinesis - Streams
		"CreateStream":     "aws_kinesis_stream",
		"DeleteStream":     "aws_kinesis_stream",
		"UpdateShardCount": "aws_kinesis_stream",

		// Kinesis - Data Operations
		"PutRecords": "aws_kinesis_stream",
		"PutRecord":  "aws_kinesis_stream",

		// Kinesis - Consumers
		"RegisterStreamConsumer":   "aws_kinesis_stream_consumer",
		"DeregisterStreamConsumer": "aws_kinesis_stream_consumer",

		// ACM (Certificate Manager)
		"RequestCertificate":        "aws_acm_certificate",
		"DeleteCertificate":         "aws_acm_certificate",
		"ImportCertificate":         "aws_acm_certificate",
		"ExportCertificate":         "aws_acm_certificate",
		"AddTagsToCertificate":      "aws_acm_certificate",
		"RemoveTagsFromCertificate": "aws_acm_certificate",

		// WAF / WAFv2 - Web ACLs
		"CreateWebACL": "aws_wafv2_web_acl",
		"DeleteWebACL": "aws_wafv2_web_acl",
		"UpdateWebACL": "aws_wafv2_web_acl",

		// WAF / WAFv2 - Rule Groups
		"CreateRuleGroup": "aws_wafv2_rule_group",
		"DeleteRuleGroup": "aws_wafv2_rule_group",
		"UpdateRuleGroup": "aws_wafv2_rule_group",

		// WAF / WAFv2 - Associations
		"AssociateWebACL":    "aws_wafv2_web_acl_association",
		"DisassociateWebACL": "aws_wafv2_web_acl_association",

		// AWS Backup - Backup Plans
		"CreateBackupPlan": "aws_backup_plan",
		"DeleteBackupPlan": "aws_backup_plan",
		"UpdateBackupPlan": "aws_backup_plan",

		// AWS Backup - Backup Vaults
		"CreateBackupVault": "aws_backup_vault",
		"DeleteBackupVault": "aws_backup_vault",

		// AWS Backup - Backup Jobs
		"StartBackupJob": "aws_backup_selection",
		"StopBackupJob":  "aws_backup_selection",

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

		// ElastiCache - Cache Clusters
		"CreateCacheCluster": "aws_elasticache_cluster",
		"DeleteCacheCluster": "aws_elasticache_cluster",
		"ModifyCacheCluster": "aws_elasticache_cluster",
		"RebootCacheCluster": "aws_elasticache_cluster",

		// ElastiCache - Replication Groups
		"CreateReplicationGroup":  "aws_elasticache_replication_group",
		"DeleteReplicationGroup":  "aws_elasticache_replication_group",
		"ModifyReplicationGroup":  "aws_elasticache_replication_group",
		"IncreaseReplicaCount":    "aws_elasticache_replication_group",
		"DecreaseReplicaCount":    "aws_elasticache_replication_group",

		// ElastiCache - Parameter Groups
		"CreateCacheParameterGroup": "aws_elasticache_parameter_group",
		"DeleteCacheParameterGroup": "aws_elasticache_parameter_group",
		"ModifyCacheParameterGroup": "aws_elasticache_parameter_group",

		// Redshift
		"ModifyCluster":               "aws_redshift_cluster",
		"ModifyClusterParameterGroup": "aws_redshift_parameter_group",
		"CreateClusterParameterGroup": "aws_redshift_parameter_group",
		"DeleteClusterParameterGroup": "aws_redshift_parameter_group",

		// SageMaker - Endpoints
		"CreateEndpoint":       "aws_sagemaker_endpoint",
		"DeleteEndpoint":       "aws_sagemaker_endpoint",
		"UpdateEndpoint":       "aws_sagemaker_endpoint",
		"CreateEndpointConfig": "aws_sagemaker_endpoint_configuration",

		// SageMaker - Training Jobs
		"CreateTrainingJob": "aws_sagemaker_training_job",
		"StopTrainingJob":   "aws_sagemaker_training_job",

		// SageMaker - Model Packages (Model Registry)
		"CreateModelPackage":        "aws_sagemaker_model_package",
		"DeleteModelPackage":        "aws_sagemaker_model_package",
		"UpdateModelPackage":        "aws_sagemaker_model_package",
		"CreateModelPackageGroup":   "aws_sagemaker_model_package_group",
		"DeleteModelPackageGroup":   "aws_sagemaker_model_package_group",

		// SageMaker - Notebook Instances
		"CreateNotebookInstance": "aws_sagemaker_notebook_instance",
		"DeleteNotebookInstance": "aws_sagemaker_notebook_instance",
		"StopNotebookInstance":   "aws_sagemaker_notebook_instance",
		"StartNotebookInstance":  "aws_sagemaker_notebook_instance",
		"UpdateNotebookInstance": "aws_sagemaker_notebook_instance",

		// Note: CreateModel/DeleteModel events conflict with API Gateway events (lines 300-301)
		// SageMaker model events cannot be distinguished from API Gateway model events without eventSource field
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
