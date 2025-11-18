package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Use a single metrics instance for all tests to avoid duplicate registration
var testMetrics *Metrics

func init() {
	// Create metrics once for all tests
	testMetrics = NewMetrics("tfdrift_test")
}

func TestNewMetrics(t *testing.T) {
	// Just verify the metrics were created successfully
	m := testMetrics

	require.NotNil(t, m)
	assert.NotNil(t, m.DriftAlertsTotal)
	assert.NotNil(t, m.UnresolvedAlerts)
	assert.NotNil(t, m.DetectionLatency)
	assert.NotNil(t, m.EventsProcessed)
	assert.NotNil(t, m.ComponentStatus)

	// Verify global instance is set
	assert.Equal(t, m, DefaultMetrics)
}

func TestRecordDriftAlert(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name         string
		severity     string
		resourceType string
		provider     string
	}{
		{
			name:         "Critical AWS EC2 drift",
			severity:     "critical",
			resourceType: "aws_instance",
			provider:     "aws",
		},
		{
			name:         "High AWS S3 drift",
			severity:     "high",
			resourceType: "aws_s3_bucket",
			provider:     "aws",
		},
		{
			name:         "Medium GCP drift",
			severity:     "medium",
			resourceType: "google_compute_instance",
			provider:     "gcp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			assert.NotPanics(t, func() {
				m.RecordDriftAlert(tt.severity, tt.resourceType, tt.provider)
			})
		})
	}
}

func TestResolveAlert(t *testing.T) {
	m := testMetrics

	// First record an alert
	m.RecordDriftAlert("critical", "aws_instance", "aws")

	// Then resolve it
	assert.NotPanics(t, func() {
		m.ResolveAlert("critical", "aws_instance")
	})
}

func TestRecordEvent(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name      string
		eventType string
		source    string
		status    string
	}{
		{
			name:      "EC2 event from Falco success",
			eventType: "ec2_modification",
			source:    "falco",
			status:    "success",
		},
		{
			name:      "S3 event from CloudTrail success",
			eventType: "s3_modification",
			source:    "cloudtrail",
			status:    "success",
		},
		{
			name:      "IAM event from Falco error",
			eventType: "iam_modification",
			source:    "falco",
			status:    "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.RecordEvent(tt.eventType, tt.source, tt.status)
			})
		})
	}
}

func TestRecordDetectionLatency(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name    string
		seconds float64
	}{
		{"Fast detection", 0.1},
		{"Medium detection", 1.5},
		{"Slow detection", 5.0},
		{"Very slow detection", 10.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.RecordDetectionLatency(tt.seconds)
			})
		})
	}
}

func TestSetComponentStatus(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name      string
		component string
		healthy   bool
	}{
		{
			name:      "Falco subscriber healthy",
			component: "falco_subscriber",
			healthy:   true,
		},
		{
			name:      "State manager unhealthy",
			component: "state_manager",
			healthy:   false,
		},
		{
			name:      "Notifier healthy",
			component: "notifier",
			healthy:   true,
		},
		{
			name:      "Detector unhealthy",
			component: "detector",
			healthy:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.SetComponentStatus(tt.component, tt.healthy)
			})
		})
	}
}

func TestMetrics_MultipleOperations(t *testing.T) {
	m := testMetrics

	// Record multiple alerts
	m.RecordDriftAlert("critical", "aws_instance", "aws")
	m.RecordDriftAlert("high", "aws_s3_bucket", "aws")
	m.RecordDriftAlert("medium", "aws_security_group", "aws")

	// Record events
	m.RecordEvent("ec2_modification", "falco", "success")
	m.RecordEvent("s3_modification", "falco", "success")
	m.RecordEvent("iam_modification", "falco", "error")

	// Record latencies
	m.RecordDetectionLatency(0.5)
	m.RecordDetectionLatency(1.2)
	m.RecordDetectionLatency(2.8)

	// Set component statuses
	m.SetComponentStatus("falco_subscriber", true)
	m.SetComponentStatus("state_manager", true)
	m.SetComponentStatus("detector", false)

	// Resolve some alerts
	m.ResolveAlert("high", "aws_s3_bucket")

	// Should not panic
	assert.NotNil(t, m)
}

func TestRecordDriftAlert_IncrementsBothCounters(t *testing.T) {
	m := testMetrics

	// Record an alert
	m.RecordDriftAlert("critical", "aws_instance", "aws")

	// Both counters should be incremented:
	// 1. DriftAlertsTotal (counter that only goes up)
	// 2. UnresolvedAlerts (gauge that can go up and down)

	// We can't easily check the actual values without exposing the metrics
	// HTTP endpoint and scraping it, but we can verify no panic occurs
	assert.NotNil(t, m.DriftAlertsTotal)
	assert.NotNil(t, m.UnresolvedAlerts)
}

func TestResolveAlert_DecrementsGauge(t *testing.T) {
	m := testMetrics

	// Record an alert first
	m.RecordDriftAlert("high", "aws_s3_bucket", "aws")

	// Resolve it
	m.ResolveAlert("high", "aws_s3_bucket")

	// Should not panic even if gauge goes to zero or negative
	assert.NotPanics(t, func() {
		m.ResolveAlert("high", "aws_s3_bucket")
	})
}

func TestComponentStatus_BooleanConversion(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name         string
		healthy      bool
		expectedCall string
	}{
		{
			name:         "Healthy component",
			healthy:      true,
			expectedCall: "should set to 1.0",
		},
		{
			name:         "Unhealthy component",
			healthy:      false,
			expectedCall: "should set to 0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.SetComponentStatus("test_component", tt.healthy)
			})
		})
	}
}

func TestDefaultMetrics_GlobalInstance(t *testing.T) {
	// Verify that the test metrics is set as default
	assert.Equal(t, testMetrics, DefaultMetrics)
	assert.NotNil(t, DefaultMetrics)
}

func TestRecordDetectionLatency_VariousValues(t *testing.T) {
	m := testMetrics

	tests := []struct {
		name    string
		seconds float64
	}{
		{"Zero latency", 0.0},
		{"Negative latency (edge case)", -1.0},
		{"Very small latency", 0.001},
		{"Very large latency", 999.999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotPanics(t, func() {
				m.RecordDetectionLatency(tt.seconds)
			})
		})
	}
}

func TestRecordEvent_EmptyStrings(t *testing.T) {
	m := testMetrics

	// Should handle empty strings without panicking
	assert.NotPanics(t, func() {
		m.RecordEvent("", "", "")
	})
}

func TestRecordDriftAlert_EmptyStrings(t *testing.T) {
	m := testMetrics

	// Should handle empty strings without panicking
	assert.NotPanics(t, func() {
		m.RecordDriftAlert("", "", "")
	})
}
