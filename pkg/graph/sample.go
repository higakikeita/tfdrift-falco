package graph

import (
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// PopulateSampleData adds sample data to the store for testing
func (s *Store) PopulateSampleData() {
	// Sample Drift Alert 1: IAM Policy Modified
	s.AddDrift(types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_iam_policy",
		ResourceName: "admin-policy",
		ResourceID:   "arn:aws:iam::123456789012:policy/admin-policy",
		Attribute:    "policy_document",
		OldValue:     `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Resource":"*"}]}`,
		NewValue:     `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
			UserName:    "admin",
		},
		MatchedRules: []string{"iam_policy_modified"},
		Timestamp:    time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
		AlertType:    "drift",
	})

	// Sample Drift Alert 2: EC2 Security Group Modified
	s.AddDrift(types.DriftAlert{
		Severity:     "critical",
		ResourceType: "aws_security_group",
		ResourceName: "web-sg",
		ResourceID:   "sg-0123456789abcdef0",
		Attribute:    "ingress_rules",
		OldValue:     `[{"protocol":"tcp","from_port":443,"to_port":443,"cidr_blocks":["10.0.0.0/8"]}]`,
		NewValue:     `[{"protocol":"tcp","from_port":443,"to_port":443,"cidr_blocks":["0.0.0.0/0"]}]`,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/devops",
			AccountID:   "123456789012",
			UserName:    "devops",
		},
		MatchedRules: []string{"security_group_ingress_changed"},
		Timestamp:    time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		AlertType:    "drift",
	})

	// Sample Drift Alert 3: S3 Bucket Encryption Disabled
	s.AddDrift(types.DriftAlert{
		Severity:     "critical",
		ResourceType: "aws_s3_bucket",
		ResourceName: "sensitive-data-bucket",
		ResourceID:   "arn:aws:s3:::sensitive-data-bucket",
		Attribute:    "server_side_encryption_configuration",
		OldValue:     `{"rule":{"apply_server_side_encryption_by_default":{"sse_algorithm":"AES256"}}}`,
		NewValue:     `null`,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/developer",
			AccountID:   "123456789012",
			UserName:    "developer",
		},
		MatchedRules: []string{"s3_bucket_encryption_disabled"},
		Timestamp:    time.Now().Add(-15 * time.Minute).Format(time.RFC3339),
		AlertType:    "drift",
	})

	// Sample Falco Event 1: IAM Policy Update
	s.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "PutUserPolicy",
		ResourceType: "aws_iam_policy",
		ResourceID:   "arn:aws:iam::123456789012:policy/admin-policy",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
			UserName:    "admin",
		},
		Changes: map[string]interface{}{
			"policyDocument": `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"*","Resource":"*"}]}`,
		},
		Region: "us-east-1",
	})

	// Sample Falco Event 2: Security Group Modification
	s.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "AuthorizeSecurityGroupIngress",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-0123456789abcdef0",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/devops",
			AccountID:   "123456789012",
			UserName:    "devops",
		},
		Changes: map[string]interface{}{
			"ipPermissions": `[{"protocol":"tcp","fromPort":443,"toPort":443,"ipRanges":[{"cidrIp":"0.0.0.0/0"}]}]`,
		},
		Region: "us-east-1",
	})

	// Sample Unmanaged Resource: Manually Created EC2 Instance
	s.AddUnmanaged(types.UnmanagedResourceAlert{
		Severity:     "medium",
		ResourceType: "aws_instance",
		ResourceID:   "i-0123456789abcdef0",
		EventName:    "RunInstances",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/developer",
			AccountID:   "123456789012",
			UserName:    "developer",
		},
		Changes: map[string]interface{}{
			"instanceType": "t3.medium",
			"imageId":      "ami-0c55b159cbfafe1f0",
		},
		Timestamp: time.Now().Add(-20 * time.Minute).Format(time.RFC3339),
		Reason:    "Resource not found in Terraform state",
	})
}
