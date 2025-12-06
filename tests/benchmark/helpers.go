package benchmark

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// setupBenchmarkDetector creates a detector instance for benchmarking
func setupBenchmarkDetector(tb testing.TB) *detector.Detector {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "../e2e/terraform/testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false, // Disabled for benchmarking
			Hostname: "localhost",
			Port:     5060,
		},
		DriftRules: []config.DriftRule{
			{
				Name:              "test-rule",
				ResourceTypes:     []string{"aws_instance", "aws_iam_role", "aws_s3_bucket"},
				WatchedAttributes: []string{"*"},
				Severity:          "high",
			},
		},
		DryRun: true,
	}

	det, err := detector.New(cfg)
	if err != nil {
		tb.Fatalf("Failed to create detector: %v", err)
	}

	return det
}

// createBenchmarkEvent creates a sample event for benchmarking
func createBenchmarkEvent() types.Event {
	return types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-1234567890abcdef0",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"instance_type": "t3.large",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}
}

// parseEvent simulates event parsing for benchmarking
func parseEvent(rawEvent map[string]interface{}) types.Event {
	// Simplified parsing for benchmark
	return types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-test",
		EventName:    rawEvent["rule"].(string),
	}
}
