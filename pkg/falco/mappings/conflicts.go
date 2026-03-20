package mappings

// ResolveEventSourceConflict resolves conflicts when multiple AWS services use the same CloudTrail event name
// Returns empty string if no conflict resolution is needed
func ResolveEventSourceConflict(eventName string, eventSource string) string {
	// For events that conflict across services, use eventSource to disambiguate
	switch eventName {
	case "CreateAlias", "DeleteAlias", "UpdateAlias":
		if eventSource == "lambda.amazonaws.com" {
			return "aws_lambda_alias"
		}
		if eventSource == "kms.amazonaws.com" || eventSource == "" {
			// Default to KMS for backward compatibility
			return "aws_kms_alias"
		}

	case "DeleteRule":
		if eventSource == "events.amazonaws.com" {
			return "aws_cloudwatch_event_rule"
		}
		if eventSource == "elasticloadbalancing.amazonaws.com" || eventSource == "" {
			// Default to ALB for backward compatibility
			return "aws_lb_listener_rule"
		}

	case "CreateTable", "DeleteTable", "UpdateTable":
		if eventSource == "glue.amazonaws.com" {
			return "aws_glue_catalog_table"
		}
		if eventSource == "dynamodb.amazonaws.com" || eventSource == "" {
			// Default to DynamoDB for backward compatibility
			return "aws_dynamodb_table"
		}

	case "CreateDatabase", "DeleteDatabase", "UpdateDatabase":
		if eventSource == "glue.amazonaws.com" {
			return "aws_glue_catalog_database"
		}
		if eventSource == "rds.amazonaws.com" {
			// RDS doesn't use CreateDatabase/DeleteDatabase for drift detection
			// RDS uses CreateDBInstance/DeleteDBInstance instead
			return "unknown"
		}

	case "TagResource", "UntagResource":
		// Generic tagging operations - use eventSource to determine resource type
		if eventSource == "states.amazonaws.com" {
			return "aws_sfn_state_machine"
		}
		// For other services, return unknown and rely on base event mapping
		return "unknown"

	case "CreateModel", "DeleteModel":
		if eventSource == "sagemaker.amazonaws.com" {
			return "aws_sagemaker_model"
		}
		if eventSource == "apigateway.amazonaws.com" || eventSource == "" {
			// Default to API Gateway for backward compatibility
			return "aws_api_gateway_model"
		}

	case "CreateCluster", "DeleteCluster", "ModifyCluster", "RebootCluster", "ResizeCluster":
		if eventSource == "kafka.amazonaws.com" {
			return "aws_msk_cluster"
		}
		if eventSource == "eks.amazonaws.com" {
			return "aws_eks_cluster"
		}
		if eventSource == "ecs.amazonaws.com" {
			return "aws_ecs_cluster"
		}
		if eventSource == "redshift.amazonaws.com" {
			return "aws_redshift_cluster"
		}
		if eventSource == "elasticache.amazonaws.com" {
			return "aws_elasticache_cluster"
		}
		// Default to EKS for backward compatibility
		return "aws_eks_cluster"

	case "PutResourcePolicy", "DeleteResourcePolicy":
		if eventSource == "secretsmanager.amazonaws.com" {
			return "aws_secretsmanager_secret_policy"
		}
		if eventSource == "backup.amazonaws.com" {
			return "aws_backup_vault_policy"
		}
		// Add more services as needed
		return "unknown"

	// New conflicts for v0.6.0 expanded services
	case "CreateDomain", "DeleteDomain":
		if eventSource == "es.amazonaws.com" || eventSource == "opensearch.amazonaws.com" {
			return "aws_opensearch_domain"
		}
		return "unknown"

	case "CreatePipeline", "DeletePipeline", "UpdatePipeline":
		if eventSource == "codepipeline.amazonaws.com" {
			return "aws_codepipeline"
		}
		return "unknown"

	case "CreateProject", "DeleteProject", "UpdateProject":
		if eventSource == "codebuild.amazonaws.com" {
			return "aws_codebuild_project"
		}
		return "unknown"

	case "CreateWebhook", "DeleteWebhook", "UpdateWebhook":
		if eventSource == "codebuild.amazonaws.com" {
			return "aws_codebuild_webhook"
		}
		if eventSource == "codepipeline.amazonaws.com" {
			return "aws_codepipeline_webhook"
		}
		return "unknown"

	case "CreateServer", "DeleteServer", "UpdateServer":
		if eventSource == "transfer.amazonaws.com" {
			return "aws_transfer_server"
		}
		return "unknown"

	case "CreateUser", "DeleteUser", "UpdateUser":
		if eventSource == "transfer.amazonaws.com" {
			return "aws_transfer_user"
		}
		if eventSource == "cognito-idp.amazonaws.com" {
			return "unknown" // User management, not Terraform resource
		}
		// Default to IAM for backward compatibility
		return ""

	case "CreateConfiguration", "DeleteConfiguration":
		if eventSource == "kafka.amazonaws.com" {
			return "aws_msk_configuration"
		}
		return "unknown"

	case "CreateDetector", "DeleteDetector", "UpdateDetector":
		if eventSource == "guardduty.amazonaws.com" {
			return "aws_guardduty_detector"
		}
		return "unknown"

	case "CreateFilter", "DeleteFilter", "UpdateFilter":
		if eventSource == "guardduty.amazonaws.com" {
			return "aws_guardduty_filter"
		}
		return "unknown"

	case "CreateIPSet", "DeleteIPSet", "UpdateIPSet":
		if eventSource == "guardduty.amazonaws.com" {
			return "aws_guardduty_ipset"
		}
		if eventSource == "wafv2.amazonaws.com" {
			return "aws_wafv2_ip_set"
		}
		return "unknown"

	case "CreateFileSystem", "DeleteFileSystem", "UpdateFileSystem":
		if eventSource == "elasticfilesystem.amazonaws.com" {
			return "aws_efs_file_system"
		}
		return "unknown"

	case "CreateApplication", "DeleteApplication":
		if eventSource == "codedeploy.amazonaws.com" {
			return "aws_codedeploy_app"
		}
		if eventSource == "kinesisanalytics.amazonaws.com" {
			return "aws_kinesis_analytics_application"
		}
		return "unknown"

	case "CreateApiKey", "DeleteApiKey":
		if eventSource == "appsync.amazonaws.com" {
			return "aws_appsync_api_key"
		}
		if eventSource == "apigateway.amazonaws.com" {
			return "aws_api_gateway_api_key"
		}
		return "unknown"
	}

	// No conflict resolution needed
	return ""
}
