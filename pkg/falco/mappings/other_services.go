package mappings

// OtherServicesMappings contains CloudTrail event to Terraform resource mappings for other AWS services
var OtherServicesMappings = map[string]string{
	// CloudWatch - Alarms
	"PutMetricAlarm":      "aws_cloudwatch_metric_alarm",
	"DeleteAlarms":        "aws_cloudwatch_metric_alarm",
	"DisableAlarmActions": "aws_cloudwatch_metric_alarm",
	"EnableAlarmActions":  "aws_cloudwatch_metric_alarm",
	"SetAlarmState":       "aws_cloudwatch_metric_alarm",

	// CloudWatch - Composite Alarms
	"PutCompositeAlarm": "aws_cloudwatch_composite_alarm",

	// CloudWatch - Logs
	"CreateLogGroup":        "aws_cloudwatch_log_group",
	"DeleteLogGroup":        "aws_cloudwatch_log_group",
	"PutRetentionPolicy":    "aws_cloudwatch_log_group",
	"DeleteRetentionPolicy": "aws_cloudwatch_log_group",
	"AssociateKmsKey":       "aws_cloudwatch_log_group",
	"DisassociateKmsKey":    "aws_cloudwatch_log_group",

	// CloudWatch - Metric Filters
	"PutMetricFilter":    "aws_cloudwatch_log_metric_filter",
	"DeleteMetricFilter": "aws_cloudwatch_log_metric_filter",

	// CloudWatch - Log Streams
	"CreateLogStream": "aws_cloudwatch_log_stream",
	"DeleteLogStream": "aws_cloudwatch_log_stream",

	// CloudWatch - Dashboards
	"PutDashboard":     "aws_cloudwatch_dashboard",
	"DeleteDashboards": "aws_cloudwatch_dashboard",

	// CloudWatch - Metric Streams
	"PutMetricStream":    "aws_cloudwatch_metric_stream",
	"DeleteMetricStream": "aws_cloudwatch_metric_stream",

	// CloudWatch - Insights Rules
	"PutInsightRule":    "aws_cloudwatch_event_rule",
	"DeleteInsightRule": "aws_cloudwatch_event_rule",

	// SNS - Topics
	"CreateTopic":         "aws_sns_topic",
	"DeleteTopic":         "aws_sns_topic",
	"SetTopicAttributes":  "aws_sns_topic",
	"Subscribe":           "aws_sns_topic_subscription",
	"Unsubscribe":         "aws_sns_topic_subscription",
	"ConfirmSubscription": "aws_sns_topic_subscription",

	// SQS - Queues
	"CreateQueue":        "aws_sqs_queue",
	"DeleteQueue":        "aws_sqs_queue",
	"SetQueueAttributes": "aws_sqs_queue",
	"PurgeQueue":         "aws_sqs_queue",

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

	// API Gateway - API Keys & Usage Plans
	"CreateApiKey":    "aws_api_gateway_api_key",
	"DeleteApiKey":    "aws_api_gateway_api_key",
	"UpdateApiKey":    "aws_api_gateway_api_key",
	"CreateUsagePlan": "aws_api_gateway_usage_plan",
	"DeleteUsagePlan": "aws_api_gateway_usage_plan",
	"UpdateUsagePlan": "aws_api_gateway_usage_plan",

	// API Gateway v2 (HTTP/WebSocket)
	"CreateApi":        "aws_apigatewayv2_api",
	"DeleteApi":        "aws_apigatewayv2_api",
	"UpdateApi":        "aws_apigatewayv2_api",
	"UpdateRoute":      "aws_apigatewayv2_route",
	"CreateIntegration": "aws_apigatewayv2_integration",
	"DeleteIntegration": "aws_apigatewayv2_integration",
	"UpdateIntegration": "aws_apigatewayv2_integration",

	// CloudFormation - Stacks
	"CreateStack":        "aws_cloudformation_stack",
	"UpdateStack":        "aws_cloudformation_stack",
	"DeleteStack":        "aws_cloudformation_stack",
	"CreateChangeSet":    "aws_cloudformation_change_set",
	"ExecuteChangeSet":   "aws_cloudformation_change_set",
	"DeleteChangeSet":    "aws_cloudformation_change_set",
	"CreateStackSet":     "aws_cloudformation_stack_set",
	"UpdateStackSet":     "aws_cloudformation_stack_set",
	"DeleteStackSet":     "aws_cloudformation_stack_set",
	"SetStackPolicy":     "aws_cloudformation_stack",

	// CloudFormation - Stack Instances
	"CreateStackInstances": "aws_cloudformation_stack_set_instance",
	"DeleteStackInstances": "aws_cloudformation_stack_set_instance",
	"UpdateStackInstances": "aws_cloudformation_stack_set_instance",

	// EventBridge (CloudWatch Events) - Rules
	"PutRule":       "aws_cloudwatch_event_rule",
	"PutTargets":    "aws_cloudwatch_event_target",
	"RemoveTargets": "aws_cloudwatch_event_target",
	"PutEvents":     "aws_cloudwatch_event_rule",

	// Step Functions - State Machines
	"CreateStateMachine": "aws_sfn_state_machine",
	"UpdateStateMachine": "aws_sfn_state_machine",
	"DeleteStateMachine": "aws_sfn_state_machine",
	"StartExecution":     "aws_sfn_state_machine",
	"StopExecution":      "aws_sfn_state_machine",

	// AWS Glue - Jobs
	"CreateJob":    "aws_glue_job",
	"UpdateJob":    "aws_glue_job",
	"DeleteJob":    "aws_glue_job",
	"CreateCrawler": "aws_glue_crawler",
	"UpdateCrawler": "aws_glue_crawler",
	"DeleteCrawler": "aws_glue_crawler",

	// Kinesis - Streams
	"CreateStream":             "aws_kinesis_stream",
	"DeleteStream":             "aws_kinesis_stream",
	"UpdateShardCount":         "aws_kinesis_stream",
	"EnableEnhancedMonitoring": "aws_kinesis_stream",
	"DisableEnhancedMonitoring": "aws_kinesis_stream",
	"StartStreamEncryption":    "aws_kinesis_stream",
	"StopStreamEncryption":     "aws_kinesis_stream",

	// Kinesis - Consumers
	"RegisterStreamConsumer":   "aws_kinesis_stream_consumer",
	"DeregisterStreamConsumer": "aws_kinesis_stream_consumer",

	// Kinesis Firehose - Delivery Streams
	"CreateDeliveryStream": "aws_kinesis_firehose_delivery_stream",
	"DeleteDeliveryStream": "aws_kinesis_firehose_delivery_stream",
	"UpdateDestination":    "aws_kinesis_firehose_delivery_stream",

	// Kinesis Analytics - Applications
	"CreateApplication": "aws_kinesis_analytics_application",
	"UpdateApplication": "aws_kinesis_analytics_application",
	"DeleteApplication": "aws_kinesis_analytics_application",

	// AWS Backup - Backup Plans
	"CreateBackupPlan": "aws_backup_plan",
	"UpdateBackupPlan": "aws_backup_plan",
	"DeleteBackupPlan": "aws_backup_plan",

	// AWS Backup - Backup Vaults
	"CreateBackupVault":             "aws_backup_vault",
	"DeleteBackupVault":             "aws_backup_vault",
	"PutBackupVaultAccessPolicy":    "aws_backup_vault_policy",
	"DeleteBackupVaultAccessPolicy": "aws_backup_vault_policy",

	// AWS Backup - Jobs
	"StartBackupJob":   "aws_backup_selection",
	"StopBackupJob":    "aws_backup_selection",
	"StartRestoreJob":  "aws_backup_selection",

	// SageMaker - Endpoints
	"CreateEndpoint":       "aws_sagemaker_endpoint",
	"UpdateEndpoint":       "aws_sagemaker_endpoint",
	"DeleteEndpoint":       "aws_sagemaker_endpoint",
	"CreateEndpointConfig": "aws_sagemaker_endpoint_configuration",
	"DeleteEndpointConfig": "aws_sagemaker_endpoint_configuration",

	// SageMaker - Training
	"CreateTrainingJob": "aws_sagemaker_training_job",
	"StopTrainingJob":   "aws_sagemaker_training_job",

	// SageMaker - Notebook Instances
	"CreateNotebookInstance": "aws_sagemaker_notebook_instance",
	"UpdateNotebookInstance": "aws_sagemaker_notebook_instance",
	"DeleteNotebookInstance": "aws_sagemaker_notebook_instance",
	"StartNotebookInstance":  "aws_sagemaker_notebook_instance",
	"StopNotebookInstance":   "aws_sagemaker_notebook_instance",

	// SageMaker - Model Packages
	"CreateModelPackage":      "aws_sagemaker_model_package",
	"UpdateModelPackage":      "aws_sagemaker_model_package",
	"DeleteModelPackage":      "aws_sagemaker_model_package",
	"CreateModelPackageGroup": "aws_sagemaker_model_package_group",
	"DeleteModelPackageGroup": "aws_sagemaker_model_package_group",
}
