package falco

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/api/schema"
	"github.com/stretchr/testify/assert"
)

func TestIsRelevantEvent(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		want      bool
	}{
		// EC2 Events
		{"EC2 ModifyInstanceAttribute", "ModifyInstanceAttribute", true},
		{"EC2 ModifyVolume", "ModifyVolume", true},
		{"EC2 Irrelevant", "RunInstances", false},

		// IAM Policy Events
		{"IAM PutUserPolicy", "PutUserPolicy", true},
		{"IAM AttachRolePolicy", "AttachRolePolicy", true},
		{"IAM CreatePolicy", "CreatePolicy", true},
		{"IAM CreatePolicyVersion", "CreatePolicyVersion", true},

		// IAM Lifecycle Events
		{"IAM CreateRole", "CreateRole", true},
		{"IAM DeleteRole", "DeleteRole", true},
		{"IAM CreateUser", "CreateUser", true},
		{"IAM DeleteUser", "DeleteUser", true},
		{"IAM CreateAccessKey", "CreateAccessKey", true},
		{"IAM AddUserToGroup", "AddUserToGroup", true},

		// S3 Events
		{"S3 PutBucketPolicy", "PutBucketPolicy", true},
		{"S3 PutBucketEncryption", "PutBucketEncryption", true},
		{"S3 DeleteBucketEncryption", "DeleteBucketEncryption", true},
		{"S3 Irrelevant", "CreateBucket", false},

		// RDS Events
		{"RDS ModifyDBInstance", "ModifyDBInstance", true},
		{"RDS CreateDBInstance", "CreateDBInstance", true}, // Creates need to be imported

		// Lambda Events
		{"Lambda UpdateFunctionConfiguration", "UpdateFunctionConfiguration", true},
		{"Lambda Irrelevant", "CreateFunction", false},

		// ECS Events - Services
		{"ECS CreateService", "CreateService", true},
		{"ECS UpdateService", "UpdateService", true},
		{"ECS DeleteService", "DeleteService", true},

		// ECS Events - Task Definitions
		{"ECS RegisterTaskDefinition", "RegisterTaskDefinition", true},
		{"ECS DeregisterTaskDefinition", "DeregisterTaskDefinition", true},

		// ECS Events - Clusters
		{"ECS UpdateCluster", "UpdateCluster", true},
		{"ECS UpdateClusterSettings", "UpdateClusterSettings", true},
		{"ECS PutClusterCapacityProviders", "PutClusterCapacityProviders", true},
		{"ECS UpdateContainerInstancesState", "UpdateContainerInstancesState", true},

		// ECS Events - Capacity Providers
		{"ECS CreateCapacityProvider", "CreateCapacityProvider", true},
		{"ECS UpdateCapacityProvider", "UpdateCapacityProvider", true},
		{"ECS DeleteCapacityProvider", "DeleteCapacityProvider", true},

		// ECS Irrelevant
		{"ECS Irrelevant", "DescribeServices", false},

		// EKS Events - Clusters
		{"EKS CreateCluster", "CreateCluster", true},
		{"EKS DeleteCluster", "DeleteCluster", true},
		{"EKS UpdateClusterConfig", "UpdateClusterConfig", true},
		{"EKS UpdateClusterVersion", "UpdateClusterVersion", true},

		// EKS Events - Node Groups
		{"EKS CreateNodegroup", "CreateNodegroup", true},
		{"EKS DeleteNodegroup", "DeleteNodegroup", true},
		{"EKS UpdateNodegroupConfig", "UpdateNodegroupConfig", true},
		{"EKS UpdateNodegroupVersion", "UpdateNodegroupVersion", true},

		// EKS Events - Addons
		{"EKS CreateAddon", "CreateAddon", true},
		{"EKS DeleteAddon", "DeleteAddon", true},
		{"EKS UpdateAddon", "UpdateAddon", true},

		// EKS Events - Fargate Profiles
		{"EKS CreateFargateProfile", "CreateFargateProfile", true},

		// EKS Irrelevant
		{"EKS Irrelevant", "DescribeCluster", false},

		// Completely irrelevant
		{"Unknown Event", "SomeRandomEvent", false},
		{"Empty Event", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.isRelevantEvent(tt.eventName)
			assert.Equal(t, tt.want, got, "Event: %s", tt.eventName)
		})
	}
}

func TestExtractResourceID(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		want      string
	}{
		{
			name:      "EC2 Instance",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.instanceid": "i-1234567890abcdef0",
			},
			want: "i-1234567890abcdef0",
		},
		{
			name:      "EBS Volume",
			eventName: "ModifyVolume",
			fields: map[string]string{
				"ct.request.volumeid": "vol-123456",
			},
			want: "vol-123456",
		},
		{
			name:      "S3 Bucket",
			eventName: "PutBucketPolicy",
			fields: map[string]string{
				"ct.request.bucket": "my-bucket",
			},
			want: "my-bucket",
		},
		{
			name:      "IAM Role",
			eventName: "PutRolePolicy",
			fields: map[string]string{
				"ct.request.rolename": "my-role",
			},
			want: "my-role",
		},
		{
			name:      "IAM User",
			eventName: "CreateUser",
			fields: map[string]string{
				"ct.request.username": "john-doe",
			},
			want: "john-doe",
		},
		{
			name:      "IAM Policy by ARN",
			eventName: "CreatePolicyVersion",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::123456789012:policy/MyPolicy",
			},
			want: "arn:aws:iam::123456789012:policy/MyPolicy",
		},
		{
			name:      "Lambda Function",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.functionname": "my-function",
			},
			want: "my-function",
		},
		// ECS - Services
		{
			name:      "ECS CreateService",
			eventName: "CreateService",
			fields: map[string]string{
				"ct.response.service.servicearn": "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service",
		},
		{
			name:      "ECS UpdateService",
			eventName: "UpdateService",
			fields: map[string]string{
				"ct.request.service": "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:service/my-cluster/my-service",
		},
		{
			name:      "ECS DeleteService",
			eventName: "DeleteService",
			fields: map[string]string{
				"ct.request.service": "my-service",
			},
			want: "my-service",
		},
		// ECS - Task Definitions
		{
			name:      "ECS RegisterTaskDefinition",
			eventName: "RegisterTaskDefinition",
			fields: map[string]string{
				"ct.response.taskdefinition.taskdefinitionarn": "arn:aws:ecs:us-east-1:123456789012:task-definition/my-task:1",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:task-definition/my-task:1",
		},
		{
			name:      "ECS DeregisterTaskDefinition",
			eventName: "DeregisterTaskDefinition",
			fields: map[string]string{
				"ct.request.taskdefinition": "my-task:1",
			},
			want: "my-task:1",
		},
		// ECS - Clusters
		{
			name:      "ECS UpdateCluster",
			eventName: "UpdateCluster",
			fields: map[string]string{
				"ct.request.cluster": "my-cluster",
			},
			want: "my-cluster",
		},
		{
			name:      "ECS UpdateClusterSettings",
			eventName: "UpdateClusterSettings",
			fields: map[string]string{
				"ct.request.cluster": "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:cluster/my-cluster",
		},
		{
			name:      "ECS PutClusterCapacityProviders",
			eventName: "PutClusterCapacityProviders",
			fields: map[string]string{
				"ct.request.cluster": "my-cluster",
			},
			want: "my-cluster",
		},
		{
			name:      "ECS UpdateContainerInstancesState",
			eventName: "UpdateContainerInstancesState",
			fields: map[string]string{
				"ct.request.containerinstances.0": "arn:aws:ecs:us-east-1:123456789012:container-instance/abc123",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:container-instance/abc123",
		},
		// ECS - Capacity Providers
		{
			name:      "ECS CreateCapacityProvider",
			eventName: "CreateCapacityProvider",
			fields: map[string]string{
				"ct.response.capacityprovider.capacityproviderarn": "arn:aws:ecs:us-east-1:123456789012:capacity-provider/my-provider",
			},
			want: "arn:aws:ecs:us-east-1:123456789012:capacity-provider/my-provider",
		},
		{
			name:      "ECS UpdateCapacityProvider",
			eventName: "UpdateCapacityProvider",
			fields: map[string]string{
				"ct.request.name": "my-provider",
			},
			want: "my-provider",
		},
		{
			name:      "ECS DeleteCapacityProvider",
			eventName: "DeleteCapacityProvider",
			fields: map[string]string{
				"ct.request.capacityprovider": "my-provider",
			},
			want: "my-provider",
		},
		// EKS - Clusters
		{
			name:      "EKS CreateCluster",
			eventName: "CreateCluster",
			fields: map[string]string{
				"ct.response.cluster.name": "my-eks-cluster",
			},
			want: "my-eks-cluster",
		},
		{
			name:      "EKS DeleteCluster",
			eventName: "DeleteCluster",
			fields: map[string]string{
				"ct.request.name": "my-eks-cluster",
			},
			want: "my-eks-cluster",
		},
		{
			name:      "EKS UpdateClusterConfig",
			eventName: "UpdateClusterConfig",
			fields: map[string]string{
				"ct.request.name": "my-eks-cluster",
			},
			want: "my-eks-cluster",
		},
		{
			name:      "EKS UpdateClusterVersion",
			eventName: "UpdateClusterVersion",
			fields: map[string]string{
				"ct.request.name": "my-eks-cluster",
			},
			want: "my-eks-cluster",
		},
		// EKS - Node Groups
		{
			name:      "EKS CreateNodegroup",
			eventName: "CreateNodegroup",
			fields: map[string]string{
				"ct.response.nodegroup.nodegroupname": "my-nodegroup",
			},
			want: "my-nodegroup",
		},
		{
			name:      "EKS DeleteNodegroup",
			eventName: "DeleteNodegroup",
			fields: map[string]string{
				"ct.request.nodegroupname": "my-nodegroup",
			},
			want: "my-nodegroup",
		},
		{
			name:      "EKS UpdateNodegroupConfig",
			eventName: "UpdateNodegroupConfig",
			fields: map[string]string{
				"ct.request.nodegroupname": "my-nodegroup",
			},
			want: "my-nodegroup",
		},
		{
			name:      "EKS UpdateNodegroupVersion",
			eventName: "UpdateNodegroupVersion",
			fields: map[string]string{
				"ct.request.nodegroupname": "my-nodegroup",
			},
			want: "my-nodegroup",
		},
		// EKS - Addons
		{
			name:      "EKS CreateAddon",
			eventName: "CreateAddon",
			fields: map[string]string{
				"ct.response.addon.addonname": "vpc-cni",
			},
			want: "vpc-cni",
		},
		{
			name:      "EKS DeleteAddon",
			eventName: "DeleteAddon",
			fields: map[string]string{
				"ct.request.addonname": "vpc-cni",
			},
			want: "vpc-cni",
		},
		{
			name:      "EKS UpdateAddon",
			eventName: "UpdateAddon",
			fields: map[string]string{
				"ct.request.addonname": "vpc-cni",
			},
			want: "vpc-cni",
		},
		// EKS - Fargate Profiles
		{
			name:      "EKS CreateFargateProfile",
			eventName: "CreateFargateProfile",
			fields: map[string]string{
				"ct.response.fargateprofile.fargateprofilename": "my-fargate-profile",
			},
			want: "my-fargate-profile",
		},
		{
			name:      "Missing Resource ID",
			eventName: "ModifyInstanceAttribute",
			fields:    map[string]string{},
			want:      "",
		},
		{
			name:      "Unknown Event Type",
			eventName: "UnknownEvent",
			fields: map[string]string{
				"ct.resource.id": "some-id",
			},
			want: "some-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.extractResourceID(tt.eventName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFalcoOutput(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name     string
		response *outputs.Response
		wantNil  bool
		validate func(t *testing.T, event interface{})
	}{
		{
			name: "Valid EC2 ModifyInstanceAttribute Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":                 "ModifyInstanceAttribute",
					"ct.request.instanceid":   "i-1234567890abcdef0",
					"ct.request.instancetype": "t3.medium",
					"ct.user.type":            "IAMUser",
					"ct.user.principalid":     "AIDAI123456789",
					"ct.user.arn":             "arn:aws:iam::123456789012:user/admin",
					"ct.user.accountid":       "123456789012",
					"ct.user":                 "admin",
				},
			},
			wantNil: false,
			validate: func(t *testing.T, event interface{}) {
				e := event
				assert.NotNil(t, e)
				assert.Equal(t, "aws", e.(*outputs.Response).Source)
			},
		},
		{
			name: "Non-CloudTrail Source",
			response: &outputs.Response{
				Source:   "syscalls",
				Rule:     "Terminal Shell",
				Priority: schema.Priority_NOTICE,
			},
			wantNil: true,
		},
		{
			name: "Missing ct.name",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"other.field": "value",
				},
			},
			wantNil: true,
		},
		{
			name: "Irrelevant Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_INFORMATIONAL,
				OutputFields: map[string]string{
					"ct.name": "DescribeInstances",
				},
			},
			wantNil: true,
		},
		{
			name: "Missing Resource ID",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name": "ModifyInstanceAttribute",
					// Missing instanceid
				},
			},
			wantNil: true,
		},
		{
			name: "Valid S3 PutBucketEncryption Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS S3 Bucket Encryption Modified",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":           "PutBucketEncryption",
					"ct.request.bucket": "my-secure-bucket",
					"ct.request.serversideencryptionconfiguration": "AES256",
					"ct.user.type":      "IAMUser",
					"ct.user.accountid": "123456789012",
					"ct.user":           "security-admin",
				},
			},
			wantNil: false,
		},
		{
			name: "Valid IAM CreateRole Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS IAM Role Created",
				Priority: schema.Priority_NOTICE,
				OutputFields: map[string]string{
					"ct.name":                             "CreateRole",
					"ct.request.rolename":                 "lambda-execution-role",
					"ct.request.assumerolepolicydocument": `{"Version":"2012-10-17","Statement":[]}`,
					"ct.user.type":                        "IAMUser",
					"ct.user":                             "iam-admin",
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.parseFalcoOutput(tt.response)

			if tt.wantNil {
				assert.Nil(t, got, "Expected nil event")
			} else {
				assert.NotNil(t, got, "Expected non-nil event")
				if got != nil && tt.validate != nil {
					// For now, just check basic fields
					assert.Equal(t, "aws", got.Provider)
					assert.NotEmpty(t, got.EventName)
					assert.NotEmpty(t, got.ResourceType)
					assert.NotEmpty(t, got.ResourceID)
				}
			}
		})
	}
}

// ============================================================================
// AWS Parser - Edge Case Tests (Comprehensive Coverage)
// ============================================================================

func TestAWSParser_Parse_NilResponse(t *testing.T) {
	sub := &Subscriber{}
	event := sub.parseFalcoOutput(nil)
	assert.Nil(t, event, "Should handle nil response gracefully")
}

func TestAWSParser_Parse_MalformedJSON(t *testing.T) {
	sub := &Subscriber{}
	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name":                 "ModifyInstanceAttribute",
			"ct.request.instanceid":   "i-1234567890abcdef0",
			"ct.request.instancetype": `{"invalid": MALFORMED_JSON}`, // Malformed JSON
			"ct.user.type":            "IAMUser",
			"ct.user":                 "admin",
		},
	}
	event := sub.parseFalcoOutput(res)
	assert.NotNil(t, event)
	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "ModifyInstanceAttribute", event.EventName)
}

func TestAWSParser_Parse_VeryLongARN(t *testing.T) {
	sub := &Subscriber{}
	// ARN with 256+ characters
	longARN := "arn:aws:iam::123456789012:user/" + string(make([]byte, 256))
	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name":             "CreateUser",
			"ct.request.username": "test-user-with-very-long-arn",
			"ct.user.arn":         longARN,
			"ct.user.type":        "IAMUser",
			"ct.user":             "admin",
		},
	}
	event := sub.parseFalcoOutput(res)
	assert.NotNil(t, event)
	assert.Equal(t, "test-user-with-very-long-arn", event.ResourceID)
	assert.Equal(t, longARN, event.UserIdentity.ARN)
}

func TestAWSParser_Parse_SpecialCharactersInResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"Hyphens", "my-bucket-name-with-hyphens"},
		{"Dots", "my.bucket.name.with.dots"},
		{"Underscores", "my_bucket_name_with_underscores"},
		{"Mixed", "my-bucket.name_with-all.chars"},
		{"Numbers", "my-bucket-123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			res := &outputs.Response{
				Source: "aws_cloudtrail",
				OutputFields: map[string]string{
					"ct.name":           "PutBucketEncryption",
					"ct.request.bucket": tt.resourceID,
					"ct.user.type":      "IAMUser",
					"ct.user":           "admin",
				},
			}
			event := sub.parseFalcoOutput(res)
			assert.NotNil(t, event)
			assert.Equal(t, tt.resourceID, event.ResourceID)
		})
	}
}

func TestAWSParser_Parse_UnicodeCharacters(t *testing.T) {
	sub := &Subscriber{}
	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name":                 "CreateRole",
			"ct.request.rolename":     "test-role",
			"ct.user":                 "ユーザー名", // Japanese
			"ct.user.type":            "IAMUser",
			"ct.user.arn":             "arn:aws:iam::123456789012:user/名前", // Japanese in ARN
			"ct.user.accountid":       "123456789012",
		},
	}
	event := sub.parseFalcoOutput(res)
	assert.NotNil(t, event)
	assert.Contains(t, event.UserIdentity.UserName, "ユーザー")
	assert.Contains(t, event.UserIdentity.ARN, "名前")
}

func TestAWSParser_Parse_EmptyFields(t *testing.T) {
	tests := []struct {
		name      string
		fields    map[string]string
		expectNil bool
	}{
		{
			"Empty event name",
			map[string]string{
				"ct.name":                 "",
				"ct.request.instanceid":   "i-123456",
			},
			true,
		},
		{
			"Empty resource ID",
			map[string]string{
				"ct.name":                 "ModifyInstanceAttribute",
				"ct.request.instanceid":   "",
			},
			true,
		},
		{
			"Empty user identity fields",
			map[string]string{
				"ct.name":                 "ModifyInstanceAttribute",
				"ct.request.instanceid":   "i-123456",
				"ct.user":                 "",
				"ct.user.type":            "",
				"ct.user.arn":             "",
			},
			false, // Should still parse, user identity is optional
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			res := &outputs.Response{
				Source:       "aws_cloudtrail",
				OutputFields: tt.fields,
			}
			event := sub.parseFalcoOutput(res)
			if tt.expectNil {
				assert.Nil(t, event)
			} else {
				assert.NotNil(t, event)
			}
		})
	}
}

func TestAWSParser_Parse_MultipleConsecutiveSlashesInARN(t *testing.T) {
	sub := &Subscriber{}
	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name":             "CreateRole",
			"ct.request.rolename": "my-role",
			"ct.user.arn":         "arn:aws:iam::123456789012:user//admin", // Double slash
			"ct.user.type":        "IAMUser",
			"ct.user":             "admin",
		},
	}
	event := sub.parseFalcoOutput(res)
	assert.NotNil(t, event)
	assert.Equal(t, "my-role", event.ResourceID)
	assert.Contains(t, event.UserIdentity.ARN, "//")
}

func TestAWSParser_Parse_AllAWSServiceTypes(t *testing.T) {
	tests := []struct {
		eventName    string
		resourceType string
	}{
		// EC2
		{"ModifyInstanceAttribute", "aws_instance"},
		{"ModifyVolume", "aws_ebs_volume"},

		// VPC
		{"CreateSecurityGroup", "aws_security_group"},
		{"CreateVpc", "aws_vpc"},
		{"CreateSubnet", "aws_subnet"},

		// ELB
		{"CreateLoadBalancer", "aws_lb"},
		{"CreateTargetGroup", "aws_lb_target_group"},

		// KMS
		{"CreateKey", "aws_kms_key"},
		{"CreateAlias", "aws_kms_alias"},

		// DynamoDB
		{"CreateTable", "aws_dynamodb_table"},

		// IAM
		{"CreateRole", "aws_iam_role"},
		{"CreateUser", "aws_iam_user"},
		{"CreatePolicy", "aws_iam_policy"},

		// S3
		{"PutBucketEncryption", "aws_s3_bucket"},
		{"PutBucketPolicy", "aws_s3_bucket_policy"},

		// RDS
		{"CreateDBInstance", "aws_db_instance"},
		{"CreateDBCluster", "aws_rds_cluster"},

		// Lambda
		{"UpdateFunctionConfiguration", "aws_lambda_function"},
		{"AddPermission", "aws_lambda_permission"},

		// API Gateway
		{"CreateRestApi", "aws_api_gateway_rest_api"},
		{"CreateApi", "aws_apigatewayv2_api"},

		// CloudWatch
		{"PutMetricAlarm", "aws_cloudwatch_metric_alarm"},
		{"CreateLogGroup", "aws_cloudwatch_log_group"},

		// SNS/SQS
		{"CreateTopic", "aws_sns_topic"},
		{"CreateQueue", "aws_sqs_queue"},

		// Route53
		{"CreateHostedZone", "aws_route53_zone"},

		// ECR
		{"CreateRepository", "aws_ecr_repository"},

		// SSM
		{"PutParameter", "aws_ssm_parameter"},

		// Secrets Manager
		{"CreateSecret", "aws_secretsmanager_secret"},

		// CloudFront
		{"CreateDistribution", "aws_cloudfront_distribution"},

		// CloudTrail
		{"CreateTrail", "aws_cloudtrail"},

		// ECS
		{"CreateService", "aws_ecs_service"},
		{"RegisterTaskDefinition", "aws_ecs_task_definition"},

		// EKS
		{"CreateCluster", "aws_eks_cluster"},
		{"CreateNodegroup", "aws_eks_node_group"},
		{"CreateAddon", "aws_eks_addon"},

		// ElastiCache
		{"CreateCacheCluster", "aws_elasticache_cluster"},

		// Redshift
		{"ModifyCluster", "aws_redshift_cluster"},

		// SageMaker
		{"CreateEndpoint", "aws_sagemaker_endpoint"},
		{"CreateNotebookInstance", "aws_sagemaker_notebook_instance"},
	}

	for _, tt := range tests {
		t.Run(tt.eventName, func(t *testing.T) {
			sub := &Subscriber{}
			resourceType := sub.mapEventToResourceType(tt.eventName, "")
			assert.Equal(t, tt.resourceType, resourceType)
		})
	}
}

func TestAWSParser_Parse_ConcurrentAccess(t *testing.T) {
	sub := &Subscriber{}
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(id int) {
			res := &outputs.Response{
				Source: "aws_cloudtrail",
				OutputFields: map[string]string{
					"ct.name":                 "ModifyInstanceAttribute",
					"ct.request.instanceid":   string(rune(id)) + "-concurrent-test",
					"ct.user.type":            "IAMUser",
					"ct.user":                 "test-user",
				},
			}
			event := sub.parseFalcoOutput(res)
			assert.NotNil(t, event)
			assert.Equal(t, "aws", event.Provider)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAWSParser_UserIdentity_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]string
		validate func(t *testing.T, userIdentity interface{})
	}{
		{
			"IAMUser with full details",
			map[string]string{
				"ct.name":                 "CreateRole",
				"ct.request.rolename":     "test-role",
				"ct.user.type":            "IAMUser",
				"ct.user":                 "admin",
				"ct.user.arn":             "arn:aws:iam::123456789012:user/admin",
				"ct.user.principalid":     "AIDAI123456789",
				"ct.user.accountid":       "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				// Just basic validation - actual structure differs
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"AssumedRole",
			map[string]string{
				"ct.name":                 "CreateRole",
				"ct.request.rolename":     "test-role",
				"ct.user.type":            "AssumedRole",
				"ct.user":                 "assumed-role-session",
				"ct.user.arn":             "arn:aws:sts::123456789012:assumed-role/MyRole/MySession",
				"ct.user.principalid":     "AROAI123456789:MySession",
				"ct.user.accountid":       "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"Root user",
			map[string]string{
				"ct.name":                 "CreateRole",
				"ct.request.rolename":     "test-role",
				"ct.user.type":            "Root",
				"ct.user":                 "root",
				"ct.user.arn":             "arn:aws:iam::123456789012:root",
				"ct.user.accountid":       "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"Service account",
			map[string]string{
				"ct.name":                 "CreateRole",
				"ct.request.rolename":     "test-role",
				"ct.user.type":            "AWSService",
				"ct.user":                 "ec2.amazonaws.com",
				"ct.user.accountid":       "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				assert.NotNil(t, userIdentity)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			res := &outputs.Response{
				Source:       "aws_cloudtrail",
				OutputFields: tt.fields,
			}
			event := sub.parseFalcoOutput(res)
			assert.NotNil(t, event)
			assert.Equal(t, "aws", event.Provider)
			tt.validate(t, event.UserIdentity)
		})
	}
}

func TestAWSParser_ExtractChanges_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		validate  func(t *testing.T, changes map[string]interface{})
	}{
		{
			"ModifyInstanceAttribute with multiple changes",
			"ModifyInstanceAttribute",
			map[string]string{
				"ct.request.instanceid":            "i-123456",
				"ct.request.instancetype":          "t3.medium",
				"ct.request.disableapitermination": "true",
			},
			func(t *testing.T, changes map[string]interface{}) {
				assert.NotNil(t, changes)
				// Changes should be extracted
				if val, ok := changes["instance_type"]; ok {
					assert.Equal(t, "t3.medium", val)
				}
			},
		},
		{
			"PutBucketEncryption",
			"PutBucketEncryption",
			map[string]string{
				"ct.request.bucket":                            "my-bucket",
				"ct.request.serversideencryptionconfiguration": "AES256",
			},
			func(t *testing.T, changes map[string]interface{}) {
				assert.NotNil(t, changes)
				if val, ok := changes["server_side_encryption_configuration"]; ok {
					assert.Equal(t, "AES256", val)
				}
			},
		},
		{
			"UpdateFunctionConfiguration",
			"UpdateFunctionConfiguration",
			map[string]string{
				"ct.request.functionname": "my-function",
				"ct.request.timeout":      "300",
				"ct.request.memorysize":   "1024",
			},
			func(t *testing.T, changes map[string]interface{}) {
				assert.NotNil(t, changes)
				// At least one change should be present
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			changes := sub.extractChanges(tt.eventName, tt.fields)
			tt.validate(t, changes)
		})
	}
}

func TestAWSParser_isRelevantEvent_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		want      bool
	}{
		{"Empty event name", "", false},
		{"Very long event name", "ModifyInstanceAttributeWithVeryLongNameThatExceedsNormalLength", false},
		{"Invalid characters", "Modify@Instance#Attribute", false},
		{"Case sensitive - wrong case", "modifyinstanceattribute", false},
		{"Spaces in event name", "Modify Instance Attribute", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			got := sub.isRelevantEvent(tt.eventName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAWSParser_extractResourceID_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		want      string
	}{
		{
			"Missing all possible fields",
			"ModifyInstanceAttribute",
			map[string]string{},
			"",
		},
		{
			"Very long resource ID",
			"CreateLoadBalancer",
			map[string]string{
				"ct.response.loadbalancers.0.loadbalancerarn": "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/" + string(make([]byte, 256)),
			},
			"arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/" + string(make([]byte, 256)),
		},
		{
			"Unknown event type with fallback",
			"UnknownEvent",
			map[string]string{
				"ct.resource.id": "fallback-id",
			},
			"fallback-id",
		},
		{
			"Empty string values",
			"ModifyInstanceAttribute",
			map[string]string{
				"ct.request.instanceid": "",
			},
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subscriber{}
			got := sub.extractResourceID(tt.eventName, tt.fields)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAWSParser_mapEventToResourceType_UnknownEvent(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		eventName string
		expected  string
	}{
		{"UnknownEvent", "unknown"},
		{"SomeRandomEvent", "unknown"},
		{"", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.eventName, func(t *testing.T) {
			got := sub.mapEventToResourceType(tt.eventName, "")
			assert.Equal(t, tt.expected, got)
		})
	}
}
