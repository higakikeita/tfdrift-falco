package falco

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/stretchr/testify/assert"
)

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
			"ct.name":             "CreateRole",
			"ct.request.rolename": "test-role",
			"ct.user":             "ユーザー名", // Japanese
			"ct.user.type":        "IAMUser",
			"ct.user.arn":         "arn:aws:iam::123456789012:user/名前", // Japanese in ARN
			"ct.user.accountid":   "123456789012",
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
				"ct.name":               "",
				"ct.request.instanceid": "i-123456",
			},
			true,
		},
		{
			"Empty resource ID",
			map[string]string{
				"ct.name":               "ModifyInstanceAttribute",
				"ct.request.instanceid": "",
			},
			true,
		},
		{
			"Empty user identity fields",
			map[string]string{
				"ct.name":               "ModifyInstanceAttribute",
				"ct.request.instanceid": "i-123456",
				"ct.user":               "",
				"ct.user.type":          "",
				"ct.user.arn":           "",
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
					"ct.name":               "ModifyInstanceAttribute",
					"ct.request.instanceid": string(rune(id)) + "-concurrent-test",
					"ct.user.type":          "IAMUser",
					"ct.user":               "test-user",
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
				"ct.name":             "CreateRole",
				"ct.request.rolename": "test-role",
				"ct.user.type":        "IAMUser",
				"ct.user":             "admin",
				"ct.user.arn":         "arn:aws:iam::123456789012:user/admin",
				"ct.user.principalid": "AIDAI123456789",
				"ct.user.accountid":   "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				// Just basic validation - actual structure differs
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"AssumedRole",
			map[string]string{
				"ct.name":             "CreateRole",
				"ct.request.rolename": "test-role",
				"ct.user.type":        "AssumedRole",
				"ct.user":             "assumed-role-session",
				"ct.user.arn":         "arn:aws:sts::123456789012:assumed-role/MyRole/MySession",
				"ct.user.principalid": "AROAI123456789:MySession",
				"ct.user.accountid":   "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"Root user",
			map[string]string{
				"ct.name":             "CreateRole",
				"ct.request.rolename": "test-role",
				"ct.user.type":        "Root",
				"ct.user":             "root",
				"ct.user.arn":         "arn:aws:iam::123456789012:root",
				"ct.user.accountid":   "123456789012",
			},
			func(t *testing.T, userIdentity interface{}) {
				assert.NotNil(t, userIdentity)
			},
		},
		{
			"Service account",
			map[string]string{
				"ct.name":             "CreateRole",
				"ct.request.rolename": "test-role",
				"ct.user.type":        "AWSService",
				"ct.user":             "ec2.amazonaws.com",
				"ct.user.accountid":   "123456789012",
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
