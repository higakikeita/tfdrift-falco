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
	}

	// No conflict resolution needed
	return ""
}
