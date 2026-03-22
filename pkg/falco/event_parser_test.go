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
		{"S3 CreateBucket", "CreateBucket", true},
		{"S3 Irrelevant", "GetBucketLocation", false},

		// RDS Events
		{"RDS ModifyDBInstance", "ModifyDBInstance", true},
		{"RDS CreateDBInstance", "CreateDBInstance", true}, // Creates need to be imported

		// Lambda Events
		{"Lambda UpdateFunctionConfiguration", "UpdateFunctionConfiguration", true},
		{"Lambda UpdateFunctionCode", "UpdateFunctionCode", true},
		{"Lambda AddPermission", "AddPermission", true},
		{"Lambda RemovePermission", "RemovePermission", true},
		{"Lambda Irrelevant", "GetFunction", false},

		// VPC Events
		{"VPC AuthorizeSecurityGroupIngress", "AuthorizeSecurityGroupIngress", true},
		{"VPC AuthorizeSecurityGroupEgress", "AuthorizeSecurityGroupEgress", true},
		{"VPC RevokeSecurityGroupIngress", "RevokeSecurityGroupIngress", true},
		{"VPC RevokeSecurityGroupEgress", "RevokeSecurityGroupEgress", true},
		{"VPC CreateSecurityGroup", "CreateSecurityGroup", true},
		{"VPC DeleteSecurityGroup", "DeleteSecurityGroup", true},
		{"VPC ModifySecurityGroupRules", "ModifySecurityGroupRules", true},
		{"VPC CreateVpc", "CreateVpc", true},
		{"VPC DeleteVpc", "DeleteVpc", true},
		{"VPC Irrelevant", "DescribeSecurityGroups", false},

		// KMS Events
		{"KMS ScheduleKeyDeletion", "ScheduleKeyDeletion", true},
		{"KMS PutKeyPolicy", "PutKeyPolicy", true},
		{"KMS CreateKey", "CreateKey", true},
		{"KMS Irrelevant", "DescribeKey", false},

		// CloudTrail Events
		{"CloudTrail CreateTrail", "CreateTrail", true},
		{"CloudTrail DeleteTrail", "DeleteTrail", true},
		{"CloudTrail UpdateTrail", "UpdateTrail", true},
		{"CloudTrail StartLogging", "StartLogging", true},
		{"CloudTrail StopLogging", "StopLogging", true},

		// API Gateway Events
		{"API Gateway CreateRestApi", "CreateRestApi", true},
		{"API Gateway DeleteRestApi", "DeleteRestApi", true},
		{"API Gateway UpdateRestApi", "UpdateRestApi", true},

		// ECS Events
		{"ECS CreateService", "CreateService", true},
		{"ECS UpdateService", "UpdateService", true},
		{"ECS DeleteService", "DeleteService", true},

		// SNS Events
		{"SNS CreateTopic", "CreateTopic", true},
		{"SNS DeleteTopic", "DeleteTopic", true},

		// SQS Events
		{"SQS CreateQueue", "CreateQueue", true},
		{"SQS DeleteQueue", "DeleteQueue", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.isRelevantEvent(tt.eventName)
			assert.Equal(t, tt.want, got, "isRelevantEvent(%q) = %v, want %v", tt.eventName, got, tt.want)
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
		// EC2 Events
		{
			name:      "ModifyInstanceAttribute with instance ID",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.instanceid": "i-1234567890abcdef0",
			},
			want: "i-1234567890abcdef0",
		},

		// S3 Events
		{
			name:      "PutBucketPolicy with bucket",
			eventName: "PutBucketPolicy",
			fields: map[string]string{
				"ct.request.bucket": "my-bucket",
			},
			want: "my-bucket",
		},

		// KMS Events
		{
			name:      "ScheduleKeyDeletion with key ID",
			eventName: "ScheduleKeyDeletion",
			fields: map[string]string{
				"ct.request.keyid": "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
			},
			want: "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
		},

		// RDS Events
		{
			name:      "CreateDBInstance with DB instance identifier",
			eventName: "CreateDBInstance",
			fields: map[string]string{
				"ct.response.dbinstanceidentifier": "my-db-instance",
			},
			want: "my-db-instance",
		},
		{
			name:      "DeleteDBInstance with DB instance identifier",
			eventName: "DeleteDBInstance",
			fields: map[string]string{
				"ct.request.dbinstanceidentifier": "my-db-instance",
			},
			want: "my-db-instance",
		},

		// VPC Events - Security Groups
		{
			name:      "AuthorizeSecurityGroupIngress with group ID",
			eventName: "AuthorizeSecurityGroupIngress",
			fields: map[string]string{
				"ct.request.groupid": "sg-1234567890abcdef0",
			},
			want: "sg-1234567890abcdef0",
		},
		{
			name:      "AuthorizeSecurityGroupIngress with group name fallback",
			eventName: "AuthorizeSecurityGroupIngress",
			fields: map[string]string{
				"ct.request.groupname": "my-security-group",
			},
			want: "my-security-group",
		},

		// VPC Events - VPC Endpoints
		{
			name:      "CreateVpcEndpoint with VPC endpoint ID in response",
			eventName: "CreateVpcEndpoint",
			fields: map[string]string{
				"ct.response.vpcendpointid": "vpce-12345678",
			},
			want: "vpce-12345678",
		},
		{
			name:      "DeleteVpcEndpoint with VPC endpoint ID in request",
			eventName: "DeleteVpcEndpoint",
			fields: map[string]string{
				"ct.request.vpcendpointid": "vpce-12345678",
			},
			want: "vpce-12345678",
		},

		// Lambda Events
		{
			name:      "UpdateFunctionConfiguration with function name",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.functionname": "my-function",
			},
			want: "my-function",
		},

		// API Gateway - REST API
		{
			name:      "CreateRestApi with API ID in response",
			eventName: "CreateRestApi",
			fields: map[string]string{
				"ct.response.id": "abcdef1234",
			},
			want: "abcdef1234",
		},

		// CloudTrail Events
		{
			name:      "CreateTrail with trail ARN in response",
			eventName: "CreateTrail",
			fields: map[string]string{
				"ct.response.trailarn": "arn:aws:cloudtrail:us-east-1:123456789012:trail/my-trail",
			},
			want: "arn:aws:cloudtrail:us-east-1:123456789012:trail/my-trail",
		},

		// Test fallback to default fields
		{
			name:      "Unknown event type with fallback",
			eventName: "UnknownEvent",
			fields: map[string]string{
				"ct.resource.id": "unknown-resource-id",
			},
			want: "unknown-resource-id",
		},

		// Test with no matching fields (should return empty string)
		{
			name:      "Event with no matching fields",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.other": "other-value",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.extractResourceID(tt.eventName, tt.fields)
			assert.Equal(t, tt.want, got, "extractResourceID(%q, ...) = %q, want %q", tt.eventName, got, tt.want)
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
			name: "Valid S3 PutBucketPolicy Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":         "PutBucketPolicy",
					"ct.request.bucket": "my-bucket",
					"ct.src":          "s3.amazonaws.com",
					"ct.user.type":    "IAMUser",
					"ct.user":         "admin",
				},
			},
			wantNil: false,
		},
		{
			name: "Valid IAM CreateRole Event",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.name":           "CreateRole",
					"ct.request.rolename": "MyRole",
					"ct.src":            "iam.amazonaws.com",
					"ct.user.type":      "IAMUser",
					"ct.user":           "admin",
				},
			},
			wantNil: false,
		},
		{
			name: "Irrelevant AWS Event (should be filtered)",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_NOTICE,
				OutputFields: map[string]string{
					"ct.name":   "DescribeInstances",
					"ct.src":    "ec2.amazonaws.com",
					"ct.user":   "admin",
				},
			},
			wantNil: true,
		},
		{
			name: "Missing ct.name Field",
			response: &outputs.Response{
				Source:   "aws_cloudtrail",
				Rule:     "AWS API Call",
				Priority: schema.Priority_WARNING,
				OutputFields: map[string]string{
					"ct.request.instanceid": "i-1234567890abcdef0",
					"ct.user":               "admin",
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
					"ct.user": "admin",
				},
			},
			wantNil: true,
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
