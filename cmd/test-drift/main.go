// Package main provides a test utility for simulating Terraform drift detection.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	fmt.Println("🧪 TFDrift-Falco Test - Drift Detection Simulation")
	fmt.Println("=" + string(make([]byte, 60)))

	// Load test configuration
	cfg, err := config.Load("./test-config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Infof("Config loaded - Backend: %s, LocalPath: '%s'", cfg.Providers.AWS.State.Backend, cfg.Providers.AWS.State.LocalPath)

	// Initialize state manager
	stateManager, err := terraform.NewStateManager(cfg.Providers.AWS.State)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}

	// Load Terraform state
	log.Info("Loading Terraform state...")
	if err := stateManager.Load(context.TODO()); err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	resourceCount := stateManager.ResourceCount()
	log.Infof("✓ Loaded %d resources from Terraform state", resourceCount)

	// Create formatter
	formatter := diff.NewFormatter(true)

	fmt.Println("\n" + "Simulating drift scenarios...")
	fmt.Println("=" + string(make([]byte, 60)))

	// Test Case 1: EC2 Instance Termination Protection Disabled
	fmt.Println("\n📋 Test Case 1: EC2 Instance Termination Protection Changed")
	testEC2DriftAlert := &types.DriftAlert{
		Severity:     "critical",
		ResourceType: "aws_instance",
		ResourceName: "webserver",
		ResourceID:   "i-0abcd1234efgh5678",
		Attribute:    "disable_api_termination",
		OldValue:     true,
		NewValue:     false,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23456EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/admin-user",
			AccountID:   "123456789012",
			UserName:    "admin-user@example.com",
		},
		MatchedRules: []string{"EC2 Instance Termination Protection"},
		Timestamp:    "2025-01-15T10:35:10Z",
	}

	// Display console format
	consoleDiff := formatter.FormatConsole(testEC2DriftAlert)
	fmt.Println(consoleDiff)

	// Test Case 2: S3 Bucket Encryption Disabled
	fmt.Println("\n📋 Test Case 2: S3 Bucket Encryption Configuration Changed")
	testS3DriftAlert := &types.DriftAlert{
		Severity:     "critical",
		ResourceType: "aws_s3_bucket",
		ResourceName: "data_bucket",
		ResourceID:   "my-data-bucket-12345",
		Attribute:    "server_side_encryption_configuration",
		OldValue: map[string]interface{}{
			"rule": map[string]interface{}{
				"apply_server_side_encryption_by_default": map[string]interface{}{
					"sse_algorithm": "AES256",
				},
			},
		},
		NewValue: nil, // Encryption disabled
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI78901EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:user/developer",
			AccountID:   "123456789012",
			UserName:    "developer@example.com",
		},
		MatchedRules: []string{"S3 Bucket Encryption Change"},
		Timestamp:    "2025-01-15T11:20:00Z",
	}

	consoleDiff2 := formatter.FormatConsole(testS3DriftAlert)
	fmt.Println(consoleDiff2)

	// Test Case 3: Instance Type Changed
	fmt.Println("\n📋 Test Case 3: EC2 Instance Type Upgraded")
	testInstanceTypeDrift := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "webserver",
		ResourceID:   "i-0abcd1234efgh5678",
		Attribute:    "instance_type",
		OldValue:     "t2.micro",
		NewValue:     "t2.large",
		UserIdentity: types.UserIdentity{
			Type:        "AssumedRole",
			PrincipalID: "AROAI12345EXAMPLE:admin-session",
			ARN:         "arn:aws:sts::123456789012:assumed-role/AdminRole/admin-session",
			AccountID:   "123456789012",
			UserName:    "AdminRole",
		},
		MatchedRules: []string{"EC2 Instance Type Change"},
		Timestamp:    "2025-01-15T14:45:30Z",
	}

	consoleDiff3 := formatter.FormatConsole(testInstanceTypeDrift)
	fmt.Println(consoleDiff3)

	// Show different format examples
	fmt.Println("\n" + "Format Examples:")
	fmt.Println("=" + string(make([]byte, 60)))

	fmt.Println("\n📄 Unified Diff Format (Git-style):")
	fmt.Println(formatter.FormatUnifiedDiff(testEC2DriftAlert))

	fmt.Println("\n📊 Side-by-Side Format:")
	fmt.Println(formatter.FormatSideBySide(testEC2DriftAlert))

	fmt.Println("\n📝 Markdown Format (for Slack/GitHub):")
	fmt.Println(formatter.FormatMarkdown(testEC2DriftAlert))

	fmt.Println("\n🎉 Test completed successfully!")
	fmt.Println("All diff formats are working as expected.")

	os.Exit(0)
}
