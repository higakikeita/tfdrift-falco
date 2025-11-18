// Package testutil provides common test utilities and fixtures for testing.
package testutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// CreateTempDir creates a temporary directory for testing and returns cleanup function
func CreateTempDir(t *testing.T, pattern string) (string, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	return dir, func() { os.RemoveAll(dir) }
}

// WriteTestFile writes content to a temporary file
func WriteTestFile(t *testing.T, dir, filename, content string) string {
	t.Helper()
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	return path
}

// CreateTestConfig returns a minimal valid configuration for testing
func CreateTestConfig() *config.Config {
	return &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  true,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled: false,
			},
			Discord: config.DiscordConfig{
				Enabled: false,
			},
			Webhook: config.WebhookConfig{
				Enabled: false,
			},
		},
		Logging: config.LoggingConfig{
			Level:  "info",
			Format: "text",
		},
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}
}

// CreateTestStateFile creates a minimal tfstate JSON file for testing
func CreateTestStateFile(t *testing.T, dir string) string {
	t.Helper()
	state := map[string]interface{}{
		"version":           4,
		"terraform_version": "1.5.0",
		"resources": []map[string]interface{}{
			{
				"mode":     "managed",
				"type":     "aws_instance",
				"name":     "test",
				"provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
				"instances": []map[string]interface{}{
					{
						"schema_version": 1,
						"attributes": map[string]interface{}{
							"id":            "i-1234567890abcdef0",
							"instance_type": "t3.micro",
							"ami":           "ami-12345678",
							"tags": map[string]interface{}{
								"Name": "test-instance",
							},
						},
					},
				},
			},
		},
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal state: %v", err)
	}

	return WriteTestFile(t, dir, "terraform.tfstate", string(data))
}

// CreateTestEvent creates a test event
func CreateTestEvent(resourceType, resourceID, eventType string) *types.Event {
	return &types.Event{
		Provider:     "aws",
		EventName:    eventType,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI123456789",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
			UserName:    "admin",
		},
		Changes: map[string]interface{}{
			"instance_type": "t3.small",
		},
	}
}

// CreateTestDriftAlert creates a test drift alert
func CreateTestDriftAlert() *types.DriftAlert {
	return &types.DriftAlert{
		Timestamp:    "2025-11-18T10:00:00Z",
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-1234567890abcdef0",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.small",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI123456789",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
			UserName:    "admin",
		},
		MatchedRules: []string{"instance_type_change"},
		AlertType:    "drift",
	}
}

// CreateTestResource creates a test Terraform resource
func CreateTestResource(resourceType, name, id string) *terraform.Resource {
	return &terraform.Resource{
		Mode:     "managed",
		Type:     resourceType,
		Name:     name,
		Provider: "provider[\"registry.terraform.io/hashicorp/aws\"]",
		Attributes: map[string]interface{}{
			"id":            id,
			"instance_type": "t3.micro",
			"ami":           "ami-12345678",
		},
	}
}

// CreateTestUnmanagedAlert creates a test unmanaged resource alert
func CreateTestUnmanagedAlert() *types.UnmanagedResourceAlert {
	return &types.UnmanagedResourceAlert{
		Timestamp:    "2025-11-18T10:00:00Z",
		Severity:     "medium",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "unmanaged-bucket-123",
		EventName:    "CreateBucket",
		Reason:       "Not found in Terraform state",
		Changes:      map[string]interface{}{"bucket_name": "unmanaged-bucket-123"},
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI123456789",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
			UserName:    "admin",
		},
	}
}
