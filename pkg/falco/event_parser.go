package falco

import (
	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// parseFalcoOutput parses a Falco output response into a TFDrift event
// Supports both AWS CloudTrail and GCP Audit Log events
func (s *Subscriber) parseFalcoOutput(res *outputs.Response) *types.Event {
	// Handle nil response
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	switch res.Source {
	case "aws_cloudtrail":
		return s.parseAWSEvent(res)
	case "gcpaudit":
		return s.gcpParser.Parse(res)
	default:
		log.Debugf("Unknown Falco source: %s", res.Source)
		return nil
	}
}

// parseAWSEvent parses AWS CloudTrail events
func (s *Subscriber) parseAWSEvent(res *outputs.Response) *types.Event {

	// Parse output fields
	fields := res.OutputFields

	// Extract CloudTrail event name
	eventName, ok := fields["ct.name"]
	if !ok || eventName == "" {
		log.Warnf("Missing ct.name in Falco output")
		return nil
	}

	// Extract event source (e.g., "ec2.amazonaws.com", "lambda.amazonaws.com")
	eventSource := getStringField(fields, "ct.src")

	// Check if this is a relevant event for drift detection
	if !s.isRelevantEvent(eventName) {
		return nil
	}

	// Extract resource ID
	resourceID := s.extractResourceID(eventName, fields)
	if resourceID == "" {
		log.Debugf("Could not extract resource ID from event %s", eventName)
		return nil
	}

	// Extract resource type (using eventSource for disambiguation)
	resourceType := s.mapEventToResourceType(eventName, eventSource)

	// Extract user identity
	userIdentity := types.UserIdentity{
		Type:        getStringField(fields, "ct.user.type"),
		PrincipalID: getStringField(fields, "ct.user.principalid"),
		ARN:         getStringField(fields, "ct.user.arn"),
		AccountID:   getStringField(fields, "ct.user.accountid"),
		UserName:    getStringField(fields, "ct.user"),
	}

	// Extract changes based on event type
	changes := s.extractChanges(eventName, fields)

	return &types.Event{
		Provider:     "aws",
		EventName:    eventName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     res,
	}
}

// isRelevantEvent checks if an event is relevant for drift detection
func (s *Subscriber) isRelevantEvent(eventName string) bool {
	relevantEvents := map[string]bool{
		// EC2
		"ModifyInstanceAttribute":         true,
		"ModifyNetworkInterfaceAttribute": true,
		"ModifyVolume":                    true,

		// VPC - Security Groups (Critical)
		"AuthorizeSecurityGroupIngress": true,
		"AuthorizeSecurityGroupEgress":  true,
		"RevokeSecurityGroupIngress":    true,
		"RevokeSecurityGroupEgress":     true,
		"CreateSecurityGroup":           true,
		"DeleteSecurityGroup":           true,
		"ModifySecurityGroupRules":      true,

		// VPC - Core
		"CreateVpc":             true,
		"DeleteVpc":             true,
		"ModifyVpcAttribute":    true,
		"CreateSubnet":          true,
		"DeleteSubnet":          true,
		"ModifySubnetAttribute": true,

		// VPC - Route Tables (Critical)
		"CreateRoute":         true,
		"DeleteRoute":         true,
		"ReplaceRoute":        true,
		"CreateRouteTable":    true,
		"DeleteRouteTable":    true,
		"AssociateRouteTable": true,

		// VPC - Internet/NAT Gateways
		"AttachInternetGateway": true,
		"DetachInternetGateway": true,
		"CreateNatGateway":      true,
		"DeleteNatGateway":      true,

		// VPC - Network ACLs
		"CreateNetworkAcl":       true,
		"DeleteNetworkAcl":       true,
		"CreateNetworkAclEntry":  true,
		"DeleteNetworkAclEntry":  true,
		"ReplaceNetworkAclEntry": true,

		// VPC - VPC Endpoints
		"CreateVpcEndpoint": true,
		"DeleteVpcEndpoint": true,
		"ModifyVpcEndpoint": true,

		// ELB/ALB - Load Balancers
		"CreateLoadBalancer":           true,
		"DeleteLoadBalancer":           true,
		"ModifyLoadBalancerAttributes": true,

		// ELB/ALB - Target Groups
		"CreateTargetGroup":           true,
		"DeleteTargetGroup":           true,
		"ModifyTargetGroup":           true,
		"ModifyTargetGroupAttributes": true,
		"RegisterTargets":             true,
		"DeregisterTargets":           true,

		// ELB/ALB - Listeners & Rules (Critical)
		"CreateListener": true,
		"DeleteListener": true,
		"ModifyListener": true,
		"CreateRule":     true,
		"DeleteRule":     true,
		"ModifyRule":     true,

		// KMS (Critical)
		"ScheduleKeyDeletion": true,
		"DisableKey":          true,
		"EnableKey":           true,
		"PutKeyPolicy":        true,
		"CreateKey":           true,
		"CreateAlias":         true,
		"DeleteAlias":         true,
		"UpdateAlias":         true,
		"EnableKeyRotation":   true,
		"DisableKeyRotation":  true,

		// DynamoDB
		"CreateTable":             true,
		"DeleteTable":             true,
		"UpdateTable":             true,
		"UpdateTimeToLive":        true,
		"UpdateContinuousBackups": true,

		// IAM - Policy modifications
		"PutUserPolicy":          true,
		"PutRolePolicy":          true,
		"PutGroupPolicy":         true,
		"UpdateAssumeRolePolicy": true,
		"AttachUserPolicy":       true,
		"AttachRolePolicy":       true,
		"AttachGroupPolicy":      true,
		"CreatePolicy":           true,
		"CreatePolicyVersion":    true,

		// IAM - User/Role/Group lifecycle
		"CreateRole":                  true,
		"DeleteRole":                  true,
		"CreateUser":                  true,
		"DeleteUser":                  true,
		"CreateAccessKey":             true,
		"AddUserToGroup":              true,
		"RemoveUserFromGroup":         true,
		"UpdateAccountPasswordPolicy": true,

		// S3
		"PutBucketPolicy":               true,
		"PutBucketVersioning":           true,
		"PutBucketEncryption":           true,
		"DeleteBucketEncryption":        true,
		"PutBucketLogging":              true,
		"PutBucketPublicAccessBlock":    true,
		"DeleteBucketPublicAccessBlock": true,
		"PutBucketAcl":                  true,

		// RDS - DB Instances
		"CreateDBInstance": true,
		"DeleteDBInstance": true,
		"ModifyDBInstance": true,
		"RebootDBInstance": true,
		"StartDBInstance":  true,
		"StopDBInstance":   true,

		// RDS - DB Clusters (including Aurora)
		"CreateDBCluster":   true,
		"DeleteDBCluster":   true,
		"ModifyDBCluster":   true,
		"StartDBCluster":    true,
		"StopDBCluster":     true,
		"FailoverDBCluster": true,

		// RDS - Aurora Specific
		"AddRoleToDBCluster":      true,
		"RemoveRoleFromDBCluster": true,
		"ModifyDBClusterEndpoint": true,
		"CreateDBClusterEndpoint": true,
		"DeleteDBClusterEndpoint": true,
		"ModifyGlobalCluster":     true,

		// RDS - Snapshots
		"CreateDBSnapshot":          true,
		"DeleteDBSnapshot":          true,
		"ModifyDBSnapshotAttribute": true,
		"CreateDBClusterSnapshot":   true,
		"DeleteDBClusterSnapshot":   true,

		// RDS - Parameter Groups
		"CreateDBParameterGroup": true,
		"DeleteDBParameterGroup": true,
		"ModifyDBParameterGroup": true,

		// RDS - Subnet Groups
		"CreateDBSubnetGroup": true,
		"DeleteDBSubnetGroup": true,
		"ModifyDBSubnetGroup": true,

		// RDS - Security & Backup
		"ModifyDBInstanceAttribute":       true,
		"RestoreDBInstanceFromDBSnapshot": true,
		"RestoreDBClusterFromSnapshot":    true,

		// Lambda
		"UpdateFunctionConfiguration": true,
		"UpdateFunctionCode":          true,
		"AddPermission":               true,
		"RemovePermission":            true,

		// API Gateway - REST API
		"CreateRestApi":    true,
		"DeleteRestApi":    true,
		"UpdateRestApi":    true,
		"CreateResource":   true,
		"DeleteResource":   true,
		"CreateMethod":     true,
		"DeleteMethod":     true,
		"PutMethod":        true,
		"UpdateMethod":     true,
		"CreateDeployment": true,
		"DeleteDeployment": true,
		"CreateStage":      true,
		"DeleteStage":      true,
		"UpdateStage":      true,

		// API Gateway - Authorizers & Models
		"CreateAuthorizer": true,
		"DeleteAuthorizer": true,
		"UpdateAuthorizer": true,
		"CreateModel":      true,
		"DeleteModel":      true,

		// API Gateway - API Keys & Usage Plans
		"CreateApiKey":    true,
		"DeleteApiKey":    true,
		"UpdateApiKey":    true,
		"CreateUsagePlan": true,
		"DeleteUsagePlan": true,
		"UpdateUsagePlan": true,

		// API Gateway v2 (HTTP/WebSocket)
		"CreateApi": true,
		"DeleteApi": true,
		"UpdateApi": true,
		// Note: CreateRoute/DeleteRoute covered by VPC Route Tables section
		"UpdateRoute":       true,
		"CreateIntegration": true,
		"DeleteIntegration": true,
		"UpdateIntegration": true,

		// CloudWatch - Alarms (Critical for monitoring)
		"PutMetricAlarm":      true,
		"DeleteAlarms":        true,
		"DisableAlarmActions": true,
		"EnableAlarmActions":  true,
		"SetAlarmState":       true,

		// CloudWatch - Logs (Critical for auditing)
		"CreateLogGroup":        true,
		"DeleteLogGroup":        true,
		"PutRetentionPolicy":    true,
		"DeleteRetentionPolicy": true,
		"AssociateKmsKey":       true,
		"DisassociateKmsKey":    true,

		// CloudWatch - Metric Filters
		"PutMetricFilter":    true,
		"DeleteMetricFilter": true,

		// CloudWatch - Log Streams
		"CreateLogStream": true,
		"DeleteLogStream": true,

		// CloudWatch - Dashboards
		"PutDashboard":     true,
		"DeleteDashboards": true,

		// SNS (Critical for alerting)
		"CreateTopic":         true,
		"DeleteTopic":         true,
		"SetTopicAttributes":  true,
		"Subscribe":           true,
		"Unsubscribe":         true,
		"ConfirmSubscription": true,
		// Note: AddPermission/RemovePermission covered by Lambda section

		// SQS (Critical for async processing)
		"CreateQueue":        true,
		"DeleteQueue":        true,
		"SetQueueAttributes": true,
		// Note: AddPermission/RemovePermission covered by Lambda section
		"PurgeQueue": true,

		// Route53 (Critical for DNS)
		"ChangeResourceRecordSets":      true,
		"CreateHostedZone":              true,
		"DeleteHostedZone":              true,
		"ChangeTagsForResource":         true,
		"AssociateVPCWithHostedZone":    true,
		"DisassociateVPCFromHostedZone": true,

		// ECR (Critical for container security)
		"PutImageScanningConfiguration": true,
		"PutImageTagMutability":         true,
		"PutLifecyclePolicy":            true,
		"DeleteLifecyclePolicy":         true,
		"SetRepositoryPolicy":           true,
		"DeleteRepositoryPolicy":        true,
		"CreateRepository":              true,
		"DeleteRepository":              true,
		"PutReplicationConfiguration":   true,

		// SSM Parameter Store (Critical for secrets)
		"PutParameter":          true,
		"DeleteParameter":       true,
		"DeleteParameters":      true,
		"LabelParameterVersion": true,

		// Secrets Manager (Critical for secrets)
		"CreateSecret":             true,
		"DeleteSecret":             true,
		"UpdateSecret":             true,
		"PutSecretValue":           true,
		"RotateSecret":             true,
		"CancelRotateSecret":       true,
		"UpdateSecretVersionStage": true,
		"PutResourcePolicy":        true,
		"DeleteResourcePolicy":     true,

		// CloudFront (Critical for CDN)
		"CreateDistribution": true,
		"DeleteDistribution": true,
		"UpdateDistribution": true,
		"CreateInvalidation": true,

		// CloudTrail (Critical for auditing)
		"CreateTrail":         true,
		"DeleteTrail":         true,
		"UpdateTrail":         true,
		"StartLogging":        true,
		"StopLogging":         true,
		"PutEventSelectors":   true,
		"PutInsightSelectors": true,

		// ECS - Services (Critical for container orchestration)
		"CreateService": true,
		"UpdateService": true,
		"DeleteService": true,

		// ECS - Task Definitions
		"RegisterTaskDefinition":   true,
		"DeregisterTaskDefinition": true,

		// ECS - Clusters
		// Note: CreateCluster/DeleteCluster are handled by context (ECS vs EKS vs Redshift)
		"UpdateCluster":                 true,
		"UpdateClusterSettings":         true,
		"PutClusterCapacityProviders":   true,
		"UpdateContainerInstancesState": true,

		// ECS - Capacity Providers
		"CreateCapacityProvider": true,
		"UpdateCapacityProvider": true,
		"DeleteCapacityProvider": true,

		// EKS - Clusters
		"CreateCluster":        true,
		"DeleteCluster":        true,
		"UpdateClusterConfig":  true,
		"UpdateClusterVersion": true,

		// EKS - Node Groups
		"CreateNodegroup":        true,
		"DeleteNodegroup":        true,
		"UpdateNodegroupConfig":  true,
		"UpdateNodegroupVersion": true,

		// EKS - Addons
		"CreateAddon": true,
		"DeleteAddon": true,
		"UpdateAddon": true,

		// EKS - Fargate Profiles
		"CreateFargateProfile": true,

		// Redshift
		// Note: CreateCluster/DeleteCluster covered by EKS section
		"ModifyCluster":               true,
		"ModifyClusterParameterGroup": true,
		"CreateClusterParameterGroup": true,
		"DeleteClusterParameterGroup": true,

		// S3 - Additional events (v0.5.0)
		"CreateBucket":              true,
		"DeleteBucket":              true,
		"PutBucketLifecycle":        true,
		"DeleteBucketLifecycle":     true,
		"PutBucketReplication":      true,
		"DeleteBucketReplication":   true,
		"PutBucketCors":             true,
		"DeleteBucketCors":          true,
		"PutBucketWebsite":          true,
		"DeleteBucketWebsite":       true,
		"PutBucketTagging":          true,
		"DeleteBucketTagging":       true,

		// Lambda - Additional events (v0.5.0)
		"CreateFunction":                  true,
		"DeleteFunction":                  true,
		"PublishVersion":                  true,
		"PutProvisionedConcurrencyConfig": true,
		"PutFunctionEventInvokeConfig":    true,
		"CreateEventSourceMapping":        true,
		"DeleteEventSourceMapping":        true,
		"UpdateEventSourceMapping":        true,
		"PutFunctionConcurrency":          true,
		"DeleteFunctionConcurrency":       true,

		// IAM - Additional events (v0.5.0)
		"UpdateRole":                    true,
		"TagRole":                       true,
		"UntagRole":                     true,
		"UpdateUser":                    true,
		"UpdateGroup":                   true,
		"CreateInstanceProfile":         true,
		"DeleteInstanceProfile":         true,
		"AddRoleToInstanceProfile":      true,
		"RemoveRoleFromInstanceProfile": true,

		// CloudFormation (v0.5.0) - Critical for IaC
		"CreateStack":        true,
		"UpdateStack":        true,
		"DeleteStack":        true,
		"CreateChangeSet":    true,
		"ExecuteChangeSet":   true,
		"DeleteChangeSet":    true,
		"CreateStackSet":     true,
		"UpdateStackSet":     true,
		"DeleteStackSet":     true,
		"SetStackPolicy":     true,

		// EventBridge (v0.5.0) - Critical for event-driven
		"PutRule":       true,
		"PutTargets":    true,
		"RemoveTargets": true,

		// Step Functions (v0.5.0) - Critical for workflow
		"CreateStateMachine": true,
		"UpdateStateMachine": true,
		"DeleteStateMachine": true,
		"StartExecution":     true,
		"StopExecution":      true,

		// AWS Glue (v0.5.0) - Critical for data pipeline
		"CreateJob":    true,
		"UpdateJob":    true,
		"DeleteJob":    true,
		"CreateCrawler": true,
		"UpdateCrawler": true,
		"DeleteCrawler": true,

		// Kinesis (v0.5.0) - Critical for streaming
		"CreateStream":                true,
		"DeleteStream":                true,
		"UpdateShardCount":            true,
		"EnableEnhancedMonitoring":    true,
		"DisableEnhancedMonitoring":   true,
		"StartStreamEncryption":       true,
		"StopStreamEncryption":        true,
		"RegisterStreamConsumer":      true,
		"DeregisterStreamConsumer":    true,
		"CreateDeliveryStream":        true, // Kinesis Firehose
		"DeleteDeliveryStream":        true, // Kinesis Firehose
		"UpdateDestination":           true, // Kinesis Firehose

		// ACM (v0.5.0) - Critical for security
		"RequestCertificate":        true,
		"DeleteCertificate":         true,
		"AddTagsToCertificate":      true,
		"RemoveTagsFromCertificate": true,
		"ImportCertificate":         true,

		// WAF / WAFv2 (v0.5.0) - Critical for security
		"CreateWebACL":          true,
		"UpdateWebACL":          true,
		"DeleteWebACL":          true,
		"CreateRuleGroup":       true,
		"UpdateRuleGroup":       true,
		"DeleteRuleGroup":       true,
		"CreateIPSet":           true,
		"UpdateIPSet":           true,
		"DeleteIPSet":           true,
		"AssociateWebACL":       true,
		"DisassociateWebACL":    true,

		// AWS Backup (v0.5.0) - Critical for disaster recovery
		"CreateBackupPlan":              true,
		"UpdateBackupPlan":              true,
		"DeleteBackupPlan":              true,
		"CreateBackupVault":             true,
		"DeleteBackupVault":             true,
		"PutBackupVaultAccessPolicy":    true,
		"DeleteBackupVaultAccessPolicy": true,
	}

	return relevantEvents[eventName]
}

// extractResourceID extracts the resource ID from Falco output fields
func (s *Subscriber) extractResourceID(eventName string, fields map[string]string) string {
	// Try different field names based on event type
	idFieldMap := map[string][]string{
		"ModifyInstanceAttribute":     {"ct.request.instanceid", "ct.resource.instanceid"},
		"ModifyVolume":                {"ct.request.volumeid"},
		"PutBucketPolicy":             {"ct.request.bucket", "ct.resource.bucket"},
		"PutBucketEncryption":         {"ct.request.bucket"},
		"DeleteBucketEncryption":      {"ct.request.bucket"},
		"ModifyDBInstance":            {"ct.request.dbinstanceidentifier"},
		"UpdateFunctionConfiguration": {"ct.request.functionname"},

		// VPC - Security Groups
		"AuthorizeSecurityGroupIngress": {"ct.request.groupid", "ct.request.groupname"},
		"AuthorizeSecurityGroupEgress":  {"ct.request.groupid", "ct.request.groupname"},
		"RevokeSecurityGroupIngress":    {"ct.request.groupid", "ct.request.groupname"},
		"RevokeSecurityGroupEgress":     {"ct.request.groupid", "ct.request.groupname"},
		"CreateSecurityGroup":           {"ct.response.groupid", "ct.request.groupname"},
		"DeleteSecurityGroup":           {"ct.request.groupid", "ct.request.groupname"},
		"ModifySecurityGroupRules":      {"ct.request.groupid"},

		// VPC - Core
		"CreateVpc":             {"ct.response.vpcid", "ct.response.vpc.vpcid"},
		"DeleteVpc":             {"ct.request.vpcid"},
		"ModifyVpcAttribute":    {"ct.request.vpcid"},
		"CreateSubnet":          {"ct.response.subnetid", "ct.response.subnet.subnetid"},
		"DeleteSubnet":          {"ct.request.subnetid"},
		"ModifySubnetAttribute": {"ct.request.subnetid"},

		// VPC - Route Tables
		"CreateRoute":         {"ct.request.routetableid"},
		"DeleteRoute":         {"ct.request.routetableid"},
		"ReplaceRoute":        {"ct.request.routetableid"},
		"CreateRouteTable":    {"ct.response.routetableid", "ct.response.routetable.routetableid"},
		"DeleteRouteTable":    {"ct.request.routetableid"},
		"AssociateRouteTable": {"ct.request.routetableid"},

		// VPC - Gateways
		"AttachInternetGateway": {"ct.request.internetgatewayid"},
		"DetachInternetGateway": {"ct.request.internetgatewayid"},
		"CreateNatGateway":      {"ct.response.natgatewayid", "ct.response.natgateway.natgatewayid"},
		"DeleteNatGateway":      {"ct.request.natgatewayid"},

		// VPC - Network ACLs
		"CreateNetworkAcl":       {"ct.response.networkaclid"},
		"DeleteNetworkAcl":       {"ct.request.networkaclid"},
		"CreateNetworkAclEntry":  {"ct.request.networkaclid"},
		"DeleteNetworkAclEntry":  {"ct.request.networkaclid"},
		"ReplaceNetworkAclEntry": {"ct.request.networkaclid"},

		// VPC - Endpoints
		"CreateVpcEndpoint": {"ct.response.vpcendpointid"},
		"DeleteVpcEndpoint": {"ct.request.vpcendpointid"},
		"ModifyVpcEndpoint": {"ct.request.vpcendpointid"},

		// ELB/ALB
		"CreateLoadBalancer":           {"ct.response.loadbalancers.0.loadbalancerarn"},
		"DeleteLoadBalancer":           {"ct.request.loadbalancerarn"},
		"ModifyLoadBalancerAttributes": {"ct.request.loadbalancerarn"},
		"CreateTargetGroup":            {"ct.response.targetgroups.0.targetgrouparn"},
		"DeleteTargetGroup":            {"ct.request.targetgrouparn"},
		"ModifyTargetGroup":            {"ct.request.targetgrouparn"},
		"ModifyTargetGroupAttributes":  {"ct.request.targetgrouparn"},
		"RegisterTargets":              {"ct.request.targetgrouparn"},
		"DeregisterTargets":            {"ct.request.targetgrouparn"},
		"CreateListener":               {"ct.response.listeners.0.listenerarn"},
		"DeleteListener":               {"ct.request.listenerarn"},
		"ModifyListener":               {"ct.request.listenerarn"},
		"CreateRule":                   {"ct.response.rules.0.rulearn"},
		"DeleteRule":                   {"ct.request.rulearn"},
		"ModifyRule":                   {"ct.request.rulearn"},

		// KMS
		"ScheduleKeyDeletion": {"ct.request.keyid"},
		"DisableKey":          {"ct.request.keyid"},
		"EnableKey":           {"ct.request.keyid"},
		"PutKeyPolicy":        {"ct.request.keyid"},
		"CreateKey":           {"ct.response.keymetadata.keyid"},
		"CreateAlias":         {"ct.request.aliasname"},
		"DeleteAlias":         {"ct.request.aliasname"},
		"UpdateAlias":         {"ct.request.aliasname"},
		"EnableKeyRotation":   {"ct.request.keyid"},
		"DisableKeyRotation":  {"ct.request.keyid"},

		// DynamoDB
		"CreateTable":             {"ct.request.tablename"},
		"DeleteTable":             {"ct.request.tablename"},
		"UpdateTable":             {"ct.request.tablename"},
		"UpdateTimeToLive":        {"ct.request.tablename"},
		"UpdateContinuousBackups": {"ct.request.tablename"},

		// IAM - Roles
		"PutRolePolicy":          {"ct.request.rolename"},
		"UpdateAssumeRolePolicy": {"ct.request.rolename"},
		"AttachRolePolicy":       {"ct.request.rolename"},
		"CreateRole":             {"ct.request.rolename"},
		"DeleteRole":             {"ct.request.rolename"},

		// IAM - Users
		"PutUserPolicy":       {"ct.request.username"},
		"AttachUserPolicy":    {"ct.request.username"},
		"CreateUser":          {"ct.request.username"},
		"DeleteUser":          {"ct.request.username"},
		"CreateAccessKey":     {"ct.request.username"},
		"AddUserToGroup":      {"ct.request.username"},
		"RemoveUserFromGroup": {"ct.request.username"},

		// IAM - Groups
		"PutGroupPolicy":    {"ct.request.groupname"},
		"AttachGroupPolicy": {"ct.request.groupname"},

		// IAM - Policies
		"CreatePolicy":        {"ct.request.policyname"},
		"CreatePolicyVersion": {"ct.request.policyarn"},

		// S3
		"PutBucketPublicAccessBlock":    {"ct.request.bucket"},
		"DeleteBucketPublicAccessBlock": {"ct.request.bucket"},
		"PutBucketAcl":                  {"ct.request.bucket"},

		// Lambda
		"AddPermission":    {"ct.request.functionname"},
		"RemovePermission": {"ct.request.functionname"},

		// RDS - DB Instances
		"CreateDBInstance":          {"ct.response.dbinstanceidentifier", "ct.request.dbinstanceidentifier"},
		"DeleteDBInstance":          {"ct.request.dbinstanceidentifier"},
		"RebootDBInstance":          {"ct.request.dbinstanceidentifier"},
		"StartDBInstance":           {"ct.request.dbinstanceidentifier"},
		"StopDBInstance":            {"ct.request.dbinstanceidentifier"},
		"ModifyDBInstanceAttribute": {"ct.request.dbinstanceidentifier"},

		// RDS - DB Clusters
		"CreateDBCluster":         {"ct.response.dbclusteridentifier", "ct.request.dbclusteridentifier"},
		"DeleteDBCluster":         {"ct.request.dbclusteridentifier"},
		"StartDBCluster":          {"ct.request.dbclusteridentifier"},
		"StopDBCluster":           {"ct.request.dbclusteridentifier"},
		"FailoverDBCluster":       {"ct.request.dbclusteridentifier"},
		"AddRoleToDBCluster":      {"ct.request.dbclusteridentifier"},
		"RemoveRoleFromDBCluster": {"ct.request.dbclusteridentifier"},
		"ModifyDBClusterEndpoint": {"ct.request.dbclusterendpointidentifier"},
		"CreateDBClusterEndpoint": {"ct.response.dbclusterendpointidentifier"},
		"DeleteDBClusterEndpoint": {"ct.request.dbclusterendpointidentifier"},
		"ModifyGlobalCluster":     {"ct.request.globalclusteridentifier"},

		// RDS - Snapshots
		"CreateDBSnapshot":          {"ct.request.dbsnapshotidentifier"},
		"DeleteDBSnapshot":          {"ct.request.dbsnapshotidentifier"},
		"ModifyDBSnapshotAttribute": {"ct.request.dbsnapshotidentifier"},
		"CreateDBClusterSnapshot":   {"ct.request.dbclustersnapshotidentifier"},
		"DeleteDBClusterSnapshot":   {"ct.request.dbclustersnapshotidentifier"},

		// RDS - Parameter Groups
		"CreateDBParameterGroup": {"ct.request.dbparametergroupname"},
		"DeleteDBParameterGroup": {"ct.request.dbparametergroupname"},
		"ModifyDBParameterGroup": {"ct.request.dbparametergroupname"},

		// RDS - Subnet Groups
		"CreateDBSubnetGroup": {"ct.request.dbsubnetgroupname"},
		"DeleteDBSubnetGroup": {"ct.request.dbsubnetgroupname"},
		"ModifyDBSubnetGroup": {"ct.request.dbsubnetgroupname"},

		// RDS - Restore
		"RestoreDBInstanceFromDBSnapshot": {"ct.request.dbinstanceidentifier"},
		"RestoreDBClusterFromSnapshot":    {"ct.request.dbclusteridentifier"},

		// API Gateway - REST API
		"CreateRestApi":    {"ct.response.id", "ct.response.restapiid"},
		"DeleteRestApi":    {"ct.request.restapiid"},
		"UpdateRestApi":    {"ct.request.restapiid"},
		"CreateResource":   {"ct.response.id"},
		"DeleteResource":   {"ct.request.resourceid"},
		"CreateMethod":     {"ct.request.resourceid"},
		"DeleteMethod":     {"ct.request.resourceid"},
		"PutMethod":        {"ct.request.resourceid"},
		"UpdateMethod":     {"ct.request.resourceid"},
		"CreateDeployment": {"ct.response.id"},
		"DeleteDeployment": {"ct.request.deploymentid"},
		"CreateStage":      {"ct.request.stagename"},
		"DeleteStage":      {"ct.request.stagename"},
		"UpdateStage":      {"ct.request.stagename"},

		// API Gateway - Authorizers & Models
		"CreateAuthorizer": {"ct.response.id"},
		"DeleteAuthorizer": {"ct.request.authorizerid"},
		"UpdateAuthorizer": {"ct.request.authorizerid"},
		"CreateModel":      {"ct.response.name"},
		"DeleteModel":      {"ct.request.modelname"},

		// API Gateway - API Keys & Usage Plans
		"CreateApiKey":    {"ct.response.id"},
		"DeleteApiKey":    {"ct.request.apikeyid"},
		"UpdateApiKey":    {"ct.request.apikeyid"},
		"CreateUsagePlan": {"ct.response.id"},
		"DeleteUsagePlan": {"ct.request.usageplanid"},
		"UpdateUsagePlan": {"ct.request.usageplanid"},

		// API Gateway v2
		"CreateApi": {"ct.response.apiid"},
		"DeleteApi": {"ct.request.apiid"},
		"UpdateApi": {"ct.request.apiid"},
		// Note: CreateRoute/DeleteRoute covered by VPC Route Tables section above
		"UpdateRoute":       {"ct.request.routeid"},
		"CreateIntegration": {"ct.response.integrationid"},
		"DeleteIntegration": {"ct.request.integrationid"},
		"UpdateIntegration": {"ct.request.integrationid"},

		// CloudWatch - Alarms
		"PutMetricAlarm":      {"ct.request.alarmname"},
		"DeleteAlarms":        {"ct.request.alarmnames.0"},
		"DisableAlarmActions": {"ct.request.alarmnames.0"},
		"EnableAlarmActions":  {"ct.request.alarmnames.0"},
		"SetAlarmState":       {"ct.request.alarmname"},

		// CloudWatch - Logs
		"CreateLogGroup":        {"ct.request.loggroupname"},
		"DeleteLogGroup":        {"ct.request.loggroupname"},
		"PutRetentionPolicy":    {"ct.request.loggroupname"},
		"DeleteRetentionPolicy": {"ct.request.loggroupname"},
		"AssociateKmsKey":       {"ct.request.loggroupname"},
		"DisassociateKmsKey":    {"ct.request.loggroupname"},
		"PutMetricFilter":       {"ct.request.loggroupname"},
		"DeleteMetricFilter":    {"ct.request.loggroupname"},
		"CreateLogStream":       {"ct.request.logstreamname"},
		"DeleteLogStream":       {"ct.request.logstreamname"},
		"PutDashboard":          {"ct.request.dashboardname"},
		"DeleteDashboards":      {"ct.request.dashboardnames.0"},

		// SNS
		"CreateTopic":         {"ct.response.topicarn"},
		"DeleteTopic":         {"ct.request.topicarn"},
		"SetTopicAttributes":  {"ct.request.topicarn"},
		"Subscribe":           {"ct.request.topicarn"},
		"Unsubscribe":         {"ct.request.subscriptionarn"},
		"ConfirmSubscription": {"ct.request.topicarn"},

		// SQS
		"CreateQueue":        {"ct.response.queueurl"},
		"DeleteQueue":        {"ct.request.queueurl"},
		"SetQueueAttributes": {"ct.request.queueurl"},
		"PurgeQueue":         {"ct.request.queueurl"},

		// Route53
		"ChangeResourceRecordSets":      {"ct.request.hostedzoneid"},
		"CreateHostedZone":              {"ct.response.hostedzone.id"},
		"DeleteHostedZone":              {"ct.request.id"},
		"ChangeTagsForResource":         {"ct.request.resourceid"},
		"AssociateVPCWithHostedZone":    {"ct.request.hostedzoneid"},
		"DisassociateVPCFromHostedZone": {"ct.request.hostedzoneid"},

		// ECR
		"PutImageScanningConfiguration": {"ct.request.repositoryname"},
		"PutImageTagMutability":         {"ct.request.repositoryname"},
		"PutLifecyclePolicy":            {"ct.request.repositoryname"},
		"DeleteLifecyclePolicy":         {"ct.request.repositoryname"},
		"SetRepositoryPolicy":           {"ct.request.repositoryname"},
		"DeleteRepositoryPolicy":        {"ct.request.repositoryname"},
		"CreateRepository":              {"ct.request.repositoryname"},
		"DeleteRepository":              {"ct.request.repositoryname"},
		"PutReplicationConfiguration":   {"ct.request.repositoryname"},

		// SSM Parameter Store
		"PutParameter":          {"ct.request.name"},
		"DeleteParameter":       {"ct.request.name"},
		"DeleteParameters":      {"ct.request.names.0"},
		"LabelParameterVersion": {"ct.request.name"},

		// Secrets Manager
		"CreateSecret":             {"ct.response.arn", "ct.response.name"},
		"DeleteSecret":             {"ct.request.secretid"},
		"UpdateSecret":             {"ct.request.secretid"},
		"PutSecretValue":           {"ct.request.secretid"},
		"RotateSecret":             {"ct.request.secretid"},
		"CancelRotateSecret":       {"ct.request.secretid"},
		"UpdateSecretVersionStage": {"ct.request.secretid"},
		"PutResourcePolicy":        {"ct.request.secretid"},
		"DeleteResourcePolicy":     {"ct.request.secretid"},

		// CloudFront
		"CreateDistribution": {"ct.response.distribution.id"},
		"DeleteDistribution": {"ct.request.id"},
		"UpdateDistribution": {"ct.request.id"},
		"CreateInvalidation": {"ct.request.distributionid"},

		// CloudTrail
		"CreateTrail":         {"ct.response.trailarn", "ct.response.name"},
		"DeleteTrail":         {"ct.request.name"},
		"UpdateTrail":         {"ct.request.name"},
		"StartLogging":        {"ct.request.name"},
		"StopLogging":         {"ct.request.name"},
		"PutEventSelectors":   {"ct.request.trailname"},
		"PutInsightSelectors": {"ct.request.trailname"},

		// ECS - Services
		"CreateService": {"ct.response.service.servicearn"},
		"UpdateService": {"ct.request.service"},
		"DeleteService": {"ct.request.service"},

		// ECS - Task Definitions
		"RegisterTaskDefinition":   {"ct.response.taskdefinition.taskdefinitionarn"},
		"DeregisterTaskDefinition": {"ct.request.taskdefinition"},

		// ECS - Clusters
		// Note: CreateCluster/DeleteCluster are context-dependent (ECS vs EKS vs Redshift)
		// ECS-specific cluster operations use different event names
		"UpdateCluster":                 {"ct.request.cluster"},
		"UpdateClusterSettings":         {"ct.request.cluster"},
		"PutClusterCapacityProviders":   {"ct.request.cluster"},
		"UpdateContainerInstancesState": {"ct.request.containerinstances.0"},

		// ECS - Capacity Providers
		"CreateCapacityProvider": {"ct.response.capacityprovider.capacityproviderarn"},
		"UpdateCapacityProvider": {"ct.request.name"},
		"DeleteCapacityProvider": {"ct.request.capacityprovider"},

		// EKS - Clusters
		"CreateCluster":        {"ct.response.cluster.name", "ct.request.name"},
		"DeleteCluster":        {"ct.request.name"},
		"UpdateClusterConfig":  {"ct.request.name"},
		"UpdateClusterVersion": {"ct.request.name"},

		// EKS - Node Groups
		"CreateNodegroup":        {"ct.response.nodegroup.nodegroupname", "ct.request.nodegroupname"},
		"DeleteNodegroup":        {"ct.request.nodegroupname"},
		"UpdateNodegroupConfig":  {"ct.request.nodegroupname"},
		"UpdateNodegroupVersion": {"ct.request.nodegroupname"},

		// EKS - Addons
		"CreateAddon": {"ct.response.addon.addonname", "ct.request.addonname"},
		"DeleteAddon": {"ct.request.addonname"},
		"UpdateAddon": {"ct.request.addonname"},

		// EKS - Fargate Profiles
		"CreateFargateProfile": {"ct.response.fargateprofile.fargateprofilename", "ct.request.fargateprofilename"},

		// Redshift
		"ModifyCluster":               {"ct.request.clusteridentifier"},
		"ModifyClusterParameterGroup": {"ct.request.parametergroupname"},
		"CreateClusterParameterGroup": {"ct.request.parametergroupname"},
		"DeleteClusterParameterGroup": {"ct.request.parametergroupname"},
	}

	// Get possible field names for this event
	possibleFields := idFieldMap[eventName]
	if possibleFields == nil {
		// Default fields to try
		possibleFields = []string{"ct.resource.id", "ct.request.resource"}
	}

	// Try each field
	for _, fieldName := range possibleFields {
		if id := getStringField(fields, fieldName); id != "" {
			return id
		}
	}

	return ""
}
