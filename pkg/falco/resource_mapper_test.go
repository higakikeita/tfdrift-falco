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
		// EC2
		{"EC2 Instance", "ModifyInstanceAttribute", "aws_instance"},
		{"EBS Volume", "ModifyVolume", "aws_ebs_volume"},

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

		// Lambda
		{"Lambda Function", "UpdateFunctionConfiguration", "aws_lambda_function"},

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
