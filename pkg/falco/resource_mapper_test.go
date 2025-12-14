package falco

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapEventToResourceType(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		want      string
	}{
		// EC2 - Instance Management
		{"EC2 Instance Run", "RunInstances", "aws_instance"},
		{"EC2 Instance Terminate", "TerminateInstances", "aws_instance"},
		{"EC2 Instance Start", "StartInstances", "aws_instance"},
		{"EC2 Instance Stop", "StopInstances", "aws_instance"},
		{"EC2 Instance Modify", "ModifyInstanceAttribute", "aws_instance"},

		// EC2 - AMI Management
		{"EC2 AMI Create", "CreateImage", "aws_ami"},
		{"EC2 AMI Deregister", "DeregisterImage", "aws_ami"},

		// EC2 - EBS Volume Management
		{"EBS Volume Create", "CreateVolume", "aws_ebs_volume"},
		{"EBS Volume Delete", "DeleteVolume", "aws_ebs_volume"},
		{"EBS Volume Attach", "AttachVolume", "aws_volume_attachment"},
		{"EBS Volume Detach", "DetachVolume", "aws_volume_attachment"},
		{"EBS Volume Modify", "ModifyVolume", "aws_ebs_volume"},

		// EC2 - Snapshot Management
		{"EBS Snapshot Create", "CreateSnapshot", "aws_ebs_snapshot"},
		{"EBS Snapshot Delete", "DeleteSnapshot", "aws_ebs_snapshot"},

		// EC2 - Network Interface Management
		{"EC2 Network Interface Create", "CreateNetworkInterface", "aws_network_interface"},
		{"EC2 Network Interface Delete", "DeleteNetworkInterface", "aws_network_interface"},
		{"EC2 Network Interface Attach", "AttachNetworkInterface", "aws_network_interface_attachment"},

		// IAM Roles
		{"IAM Role Policy", "PutRolePolicy", "aws_iam_role_policy"},
		{"IAM Role", "CreateRole", "aws_iam_role"},
		{"IAM Role Assume Policy", "UpdateAssumeRolePolicy", "aws_iam_role"},
		{"IAM Role Policy Attachment", "AttachRolePolicy", "aws_iam_role_policy_attachment"},

		// IAM Users
		{"IAM User Policy", "PutUserPolicy", "aws_iam_user_policy"},
		{"IAM User", "CreateUser", "aws_iam_user"},
		{"IAM Access Key", "CreateAccessKey", "aws_iam_access_key"},

		// IAM Groups
		{"IAM Group Policy", "PutGroupPolicy", "aws_iam_group_policy"},

		// IAM Policies
		{"IAM Policy", "CreatePolicy", "aws_iam_policy"},
		{"IAM Policy Version", "CreatePolicyVersion", "aws_iam_policy"},

		// IAM Account
		{"IAM Account Password Policy", "UpdateAccountPasswordPolicy", "aws_iam_account_password_policy"},

		// S3
		{"S3 Bucket Policy", "PutBucketPolicy", "aws_s3_bucket_policy"},
		{"S3 Bucket Encryption", "PutBucketEncryption", "aws_s3_bucket"},

		// RDS - DB Instances
		{"RDS Instance Create", "CreateDBInstance", "aws_db_instance"},
		{"RDS Instance Delete", "DeleteDBInstance", "aws_db_instance"},
		{"RDS Instance Modify", "ModifyDBInstance", "aws_db_instance"},
		{"RDS Instance Reboot", "RebootDBInstance", "aws_db_instance"},
		{"RDS Instance Start", "StartDBInstance", "aws_db_instance"},
		{"RDS Instance Stop", "StopDBInstance", "aws_db_instance"},
		{"RDS Instance Attribute Modify", "ModifyDBInstanceAttribute", "aws_db_instance"},
		{"RDS Instance Read Replica Create", "CreateDBInstanceReadReplica", "aws_db_instance"},

		// RDS - DB Clusters (Aurora)
		{"RDS Cluster Create", "CreateDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Delete", "DeleteDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Modify", "ModifyDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Start", "StartDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Stop", "StopDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Failover", "FailoverDBCluster", "aws_rds_cluster"},
		{"RDS Cluster Role Add", "AddRoleToDBCluster", "aws_rds_cluster_role_association"},
		{"RDS Cluster Role Remove", "RemoveRoleFromDBCluster", "aws_rds_cluster_role_association"},
		{"RDS Cluster Endpoint Modify", "ModifyDBClusterEndpoint", "aws_rds_cluster_endpoint"},
		{"RDS Cluster Endpoint Create", "CreateDBClusterEndpoint", "aws_rds_cluster_endpoint"},
		{"RDS Cluster Endpoint Delete", "DeleteDBClusterEndpoint", "aws_rds_cluster_endpoint"},
		{"RDS Global Cluster Modify", "ModifyGlobalCluster", "aws_rds_global_cluster"},

		// RDS - Snapshots
		{"RDS DB Snapshot Create", "CreateDBSnapshot", "aws_db_snapshot"},
		{"RDS DB Snapshot Delete", "DeleteDBSnapshot", "aws_db_snapshot"},
		{"RDS DB Snapshot Attribute Modify", "ModifyDBSnapshotAttribute", "aws_db_snapshot"},
		{"RDS Cluster Snapshot Create", "CreateDBClusterSnapshot", "aws_db_cluster_snapshot"},
		{"RDS Cluster Snapshot Delete", "DeleteDBClusterSnapshot", "aws_db_cluster_snapshot"},

		// RDS - Parameter Groups
		{"RDS Parameter Group Create", "CreateDBParameterGroup", "aws_db_parameter_group"},
		{"RDS Parameter Group Delete", "DeleteDBParameterGroup", "aws_db_parameter_group"},
		{"RDS Parameter Group Modify", "ModifyDBParameterGroup", "aws_db_parameter_group"},

		// RDS - Subnet Groups
		{"RDS Subnet Group Create", "CreateDBSubnetGroup", "aws_db_subnet_group"},
		{"RDS Subnet Group Delete", "DeleteDBSubnetGroup", "aws_db_subnet_group"},
		{"RDS Subnet Group Modify", "ModifyDBSubnetGroup", "aws_db_subnet_group"},

		// RDS - Restore
		{"RDS Instance Restore From Snapshot", "RestoreDBInstanceFromDBSnapshot", "aws_db_instance"},
		{"RDS Instance Restore To Point In Time", "RestoreDBInstanceToPointInTime", "aws_db_instance"},
		{"RDS Cluster Restore From Snapshot", "RestoreDBClusterFromSnapshot", "aws_rds_cluster"},

		// RDS - Option Groups
		{"RDS Option Group Create", "CreateOptionGroup", "aws_db_option_group"},
		{"RDS Option Group Delete", "DeleteOptionGroup", "aws_db_option_group"},
		{"RDS Option Group Modify", "ModifyOptionGroup", "aws_db_option_group"},

		// Lambda - Function Management
		{"Lambda Function Create", "CreateFunction", "aws_lambda_function"},
		{"Lambda Function Delete", "DeleteFunction", "aws_lambda_function"},
		{"Lambda Function Code Update", "UpdateFunctionCode", "aws_lambda_function"},
		{"Lambda Function Config", "UpdateFunctionConfiguration", "aws_lambda_function"},

		// Lambda - Permissions
		{"Lambda Permission Add", "AddPermission", "aws_lambda_permission"},
		{"Lambda Permission Remove", "RemovePermission", "aws_lambda_permission"},

		// Lambda - Event Source Mappings
		{"Lambda Event Source Create", "CreateEventSourceMapping", "aws_lambda_event_source_mapping"},
		{"Lambda Event Source Delete", "DeleteEventSourceMapping", "aws_lambda_event_source_mapping"},
		{"Lambda Event Source Update", "UpdateEventSourceMapping", "aws_lambda_event_source_mapping"},

		// Lambda - Concurrency
		{"Lambda Concurrency Put", "PutFunctionConcurrency", "aws_lambda_function"},

		// Note: Lambda alias events (CreateAlias, DeleteAlias, UpdateAlias) are tested under KMS
		// as they share the same event names and cannot be distinguished without eventSource

		// Auto Scaling - Auto Scaling Groups
		{"Auto Scaling Group Create", "CreateAutoScalingGroup", "aws_autoscaling_group"},
		{"Auto Scaling Group Delete", "DeleteAutoScalingGroup", "aws_autoscaling_group"},
		{"Auto Scaling Group Update", "UpdateAutoScalingGroup", "aws_autoscaling_group"},
		{"Auto Scaling Set Desired Capacity", "SetDesiredCapacity", "aws_autoscaling_group"},

		// Auto Scaling - Launch Configurations
		{"Launch Configuration Create", "CreateLaunchConfiguration", "aws_launch_configuration"},
		{"Launch Configuration Delete", "DeleteLaunchConfiguration", "aws_launch_configuration"},

		// Auto Scaling - Scaling Policies
		{"Auto Scaling Policy Put", "PutScalingPolicy", "aws_autoscaling_policy"},
		{"Auto Scaling Policy Delete", "DeletePolicy", "aws_autoscaling_policy"},

		// Auto Scaling - Scheduled Actions
		{"Auto Scaling Scheduled Action Put", "PutScheduledUpdateGroupAction", "aws_autoscaling_schedule"},
		{"Auto Scaling Scheduled Action Delete", "DeleteScheduledAction", "aws_autoscaling_schedule"},

		// ECS - Services
		{"ECS Service Create", "CreateService", "aws_ecs_service"},
		{"ECS Service Update", "UpdateService", "aws_ecs_service"},
		{"ECS Service Delete", "DeleteService", "aws_ecs_service"},

		// ECS - Task Definitions
		{"ECS Task Definition Register", "RegisterTaskDefinition", "aws_ecs_task_definition"},
		{"ECS Task Definition Deregister", "DeregisterTaskDefinition", "aws_ecs_task_definition"},

		// ECS - Clusters
		{"ECS Cluster Update", "UpdateCluster", "aws_ecs_cluster"},
		{"ECS Cluster Settings Update", "UpdateClusterSettings", "aws_ecs_cluster"},
		{"ECS Cluster Capacity Providers", "PutClusterCapacityProviders", "aws_ecs_cluster_capacity_providers"},
		{"ECS Container Instance State", "UpdateContainerInstancesState", "aws_ecs_container_instance"},

		// ECS - Capacity Providers
		{"ECS Capacity Provider Create", "CreateCapacityProvider", "aws_ecs_capacity_provider"},
		{"ECS Capacity Provider Update", "UpdateCapacityProvider", "aws_ecs_capacity_provider"},
		{"ECS Capacity Provider Delete", "DeleteCapacityProvider", "aws_ecs_capacity_provider"},

		// EKS - Clusters
		{"EKS Cluster Create", "CreateCluster", "aws_eks_cluster"},
		{"EKS Cluster Delete", "DeleteCluster", "aws_eks_cluster"},
		{"EKS Cluster Config Update", "UpdateClusterConfig", "aws_eks_cluster"},
		{"EKS Cluster Version Update", "UpdateClusterVersion", "aws_eks_cluster"},

		// EKS - Node Groups
		{"EKS Node Group Create", "CreateNodegroup", "aws_eks_node_group"},
		{"EKS Node Group Delete", "DeleteNodegroup", "aws_eks_node_group"},
		{"EKS Node Group Config Update", "UpdateNodegroupConfig", "aws_eks_node_group"},
		{"EKS Node Group Version Update", "UpdateNodegroupVersion", "aws_eks_node_group"},

		// EKS - Addons
		{"EKS Addon Create", "CreateAddon", "aws_eks_addon"},
		{"EKS Addon Delete", "DeleteAddon", "aws_eks_addon"},
		{"EKS Addon Update", "UpdateAddon", "aws_eks_addon"},

		// EKS - Fargate Profiles
		{"EKS Fargate Profile Create", "CreateFargateProfile", "aws_eks_fargate_profile"},

		// ElastiCache - Cache Clusters
		{"ElastiCache Cluster Create", "CreateCacheCluster", "aws_elasticache_cluster"},
		{"ElastiCache Cluster Delete", "DeleteCacheCluster", "aws_elasticache_cluster"},
		{"ElastiCache Cluster Modify", "ModifyCacheCluster", "aws_elasticache_cluster"},
		{"ElastiCache Cluster Reboot", "RebootCacheCluster", "aws_elasticache_cluster"},

		// ElastiCache - Replication Groups
		{"ElastiCache Replication Group Create", "CreateReplicationGroup", "aws_elasticache_replication_group"},
		{"ElastiCache Replication Group Delete", "DeleteReplicationGroup", "aws_elasticache_replication_group"},
		{"ElastiCache Replication Group Modify", "ModifyReplicationGroup", "aws_elasticache_replication_group"},
		{"ElastiCache Replica Count Increase", "IncreaseReplicaCount", "aws_elasticache_replication_group"},
		{"ElastiCache Replica Count Decrease", "DecreaseReplicaCount", "aws_elasticache_replication_group"},

		// ElastiCache - Parameter Groups
		{"ElastiCache Parameter Group Create", "CreateCacheParameterGroup", "aws_elasticache_parameter_group"},
		{"ElastiCache Parameter Group Delete", "DeleteCacheParameterGroup", "aws_elasticache_parameter_group"},
		{"ElastiCache Parameter Group Modify", "ModifyCacheParameterGroup", "aws_elasticache_parameter_group"},

		// DynamoDB - Point-in-Time Recovery
		{"DynamoDB Restore To Point In Time", "RestoreTableToPointInTime", "aws_dynamodb_table"},

		// DynamoDB - Backups
		{"DynamoDB Backup Create", "CreateBackup", "aws_dynamodb_table_backup"},
		{"DynamoDB Backup Delete", "DeleteBackup", "aws_dynamodb_table_backup"},
		{"DynamoDB Restore From Backup", "RestoreTableFromBackup", "aws_dynamodb_table"},

		// DynamoDB - Global Tables
		{"DynamoDB Global Table Create", "CreateGlobalTable", "aws_dynamodb_global_table"},
		{"DynamoDB Global Table Update", "UpdateGlobalTable", "aws_dynamodb_global_table"},

		// DynamoDB - Streams
		{"DynamoDB Kinesis Streaming Enable", "EnableKinesisStreamingDestination", "aws_dynamodb_kinesis_streaming_destination"},
		{"DynamoDB Kinesis Streaming Disable", "DisableKinesisStreamingDestination", "aws_dynamodb_kinesis_streaming_destination"},

		// DynamoDB - Monitoring
		{"DynamoDB Contributor Insights Update", "UpdateContributorInsights", "aws_dynamodb_contributor_insights"},

		// VPC - Peering
		{"VPC Peering Connection Create", "CreateVpcPeeringConnection", "aws_vpc_peering_connection"},
		{"VPC Peering Connection Accept", "AcceptVpcPeeringConnection", "aws_vpc_peering_connection_accepter"},
		{"VPC Peering Connection Delete", "DeleteVpcPeeringConnection", "aws_vpc_peering_connection"},

		// VPC - Transit Gateway
		{"Transit Gateway Create", "CreateTransitGateway", "aws_ec2_transit_gateway"},
		{"Transit Gateway Delete", "DeleteTransitGateway", "aws_ec2_transit_gateway"},
		{"Transit Gateway VPC Attachment Create", "CreateTransitGatewayVpcAttachment", "aws_ec2_transit_gateway_vpc_attachment"},

		// VPC - Flow Logs
		{"VPC Flow Logs Create", "CreateFlowLogs", "aws_flow_log"},
		{"VPC Flow Logs Delete", "DeleteFlowLogs", "aws_flow_log"},

		// VPC - Network Firewall
		{"Network Firewall Delete", "DeleteFirewall", "aws_networkfirewall_firewall"},

		// SageMaker - Endpoints
		{"SageMaker Endpoint Create", "CreateEndpoint", "aws_sagemaker_endpoint"},
		{"SageMaker Endpoint Delete", "DeleteEndpoint", "aws_sagemaker_endpoint"},
		{"SageMaker Endpoint Update", "UpdateEndpoint", "aws_sagemaker_endpoint"},
		{"SageMaker Endpoint Config Create", "CreateEndpointConfig", "aws_sagemaker_endpoint_configuration"},

		// SageMaker - Training Jobs
		{"SageMaker Training Job Create", "CreateTrainingJob", "aws_sagemaker_training_job"},
		{"SageMaker Training Job Stop", "StopTrainingJob", "aws_sagemaker_training_job"},

		// SageMaker - Model Packages
		{"SageMaker Model Package Create", "CreateModelPackage", "aws_sagemaker_model_package"},
		{"SageMaker Model Package Delete", "DeleteModelPackage", "aws_sagemaker_model_package"},
		{"SageMaker Model Package Update", "UpdateModelPackage", "aws_sagemaker_model_package"},
		{"SageMaker Model Package Group Create", "CreateModelPackageGroup", "aws_sagemaker_model_package_group"},
		{"SageMaker Model Package Group Delete", "DeleteModelPackageGroup", "aws_sagemaker_model_package_group"},

		// SageMaker - Notebook Instances
		{"SageMaker Notebook Instance Create", "CreateNotebookInstance", "aws_sagemaker_notebook_instance"},
		{"SageMaker Notebook Instance Delete", "DeleteNotebookInstance", "aws_sagemaker_notebook_instance"},
		{"SageMaker Notebook Instance Stop", "StopNotebookInstance", "aws_sagemaker_notebook_instance"},
		{"SageMaker Notebook Instance Start", "StartNotebookInstance", "aws_sagemaker_notebook_instance"},
		{"SageMaker Notebook Instance Update", "UpdateNotebookInstance", "aws_sagemaker_notebook_instance"},

		// Unknown
		{"Unknown Event", "UnknownEvent", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.mapEventToResourceType(tt.eventName)
			assert.Equal(t, tt.want, got)
		})
	}
}
