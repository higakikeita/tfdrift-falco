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

		// RDS
		{"RDS Instance", "ModifyDBInstance", "aws_db_instance"},

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
