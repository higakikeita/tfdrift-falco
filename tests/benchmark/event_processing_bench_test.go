// +build ignore

package benchmark

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// BenchmarkEventProcessing_Single measures single event processing performance
func BenchmarkEventProcessing_Single(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := createBenchmarkEvent()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		det.HandleEvent(event)
	}
}

// BenchmarkEventProcessing_Batch measures batch event processing
func BenchmarkEventProcessing_Batch(b *testing.B) {
	det := setupBenchmarkDetector(b)
	events := make([]types.Event, 100)
	for i := range events {
		events[i] = createBenchmarkEvent()
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, event := range events {
			det.HandleEvent(event)
		}
	}
}

// BenchmarkEventParsing measures Falco event parsing performance
func BenchmarkEventParsing(b *testing.B) {
	rawEvent := map[string]interface{}{
		"source":   "aws_cloudtrail",
		"rule":     "EC2 Instance Modified",
		"priority": "WARNING",
		"output_fields": map[string]string{
			"ct.name":               "ModifyInstanceAttribute",
			"ct.request.instanceid": "i-1234567890abcdef0",
			"ct.user":               "admin",
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate parsing
		_ = parseEvent(rawEvent)
	}
}

// BenchmarkStateComparison measures state comparison performance
func BenchmarkStateComparison(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := createBenchmarkEvent()

	// Load state once
	ctx := context.Background()
	_ = det.GetStateManager().Load(ctx)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate state comparison
		resource := det.GetStateManager().GetResource(event.ResourceType, event.ResourceID)
		_ = resource != nil
	}
}

// BenchmarkDriftDetection_EC2 measures EC2 drift detection performance
func BenchmarkDriftDetection_EC2(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-1234567890abcdef0",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"disable_api_termination": false,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		det.HandleEvent(event)
	}
}

// BenchmarkDriftDetection_IAM measures IAM drift detection performance
func BenchmarkDriftDetection_IAM(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_iam_role",
		ResourceID:   "test-role",
		EventName:    "UpdateAssumeRolePolicy",
		Changes: map[string]interface{}{
			"assume_role_policy": map[string]interface{}{
				"Version": "2012-10-17",
			},
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		det.HandleEvent(event)
	}
}

// BenchmarkDriftDetection_S3 measures S3 drift detection performance
func BenchmarkDriftDetection_S3(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "test-bucket",
		EventName:    "DeleteBucketEncryption",
		Changes: map[string]interface{}{
			"server_side_encryption_configuration": nil,
		},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		det.HandleEvent(event)
	}
}

// BenchmarkConcurrentEvents measures concurrent event handling
func BenchmarkConcurrentEvents(b *testing.B) {
	det := setupBenchmarkDetector(b)
	event := createBenchmarkEvent()

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			det.HandleEvent(event)
		}
	})
}

// Helper functions

func setupBenchmarkDetector(b *testing.B) *detector.Detector {
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
		b.Fatalf("Failed to create detector: %v", err)
	}

	return det
}

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

func parseEvent(rawEvent map[string]interface{}) types.Event {
	// Simplified parsing for benchmark
	return types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-test",
		EventName:    rawEvent["rule"].(string),
	}
}
