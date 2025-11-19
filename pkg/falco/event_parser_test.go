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
		{"RDS Irrelevant", "CreateDBInstance", false},

		// Lambda Events
		{"Lambda UpdateFunctionConfiguration", "UpdateFunctionConfiguration", true},
		{"Lambda Irrelevant", "CreateFunction", false},

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
