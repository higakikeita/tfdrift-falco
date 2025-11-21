package integration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/require"
)

// MockWebhookServer creates a mock HTTP server for webhook testing
type MockWebhookServer struct {
	Server          *httptest.Server
	ReceivedPayload interface{}
	StatusCode      int
	t               *testing.T
}

// NewMockWebhookServer creates a new mock webhook server
func NewMockWebhookServer(t *testing.T, handler http.HandlerFunc) *MockWebhookServer {
	mock := &MockWebhookServer{
		t:          t,
		StatusCode: http.StatusOK,
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(w, r)
		}
		w.WriteHeader(mock.StatusCode)
	}))

	return mock
}

// Close closes the mock server
func (m *MockWebhookServer) Close() {
	m.Server.Close()
}

// CreateTestAlert creates a test drift alert
func CreateTestAlert() types.DriftAlert {
	return types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-1234567890abcdef0",
		ResourceName: "web-server",
		Changes: map[string]types.Change{
			"disable_api_termination": {
				Attribute: "disable_api_termination",
				OldValue:  true,
				NewValue:  false,
			},
		},
		Severity: "high",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			UserName:    "admin@example.com",
			PrincipalID: "AIDAI1234567890ABCD",
			ARN:         "arn:aws:iam::123456789012:user/admin",
			AccountID:   "123456789012",
		},
		EventMetadata: types.EventMetadata{
			EventTime: "2025-11-20T10:30:00Z",
			EventName: "ModifyInstanceAttribute",
			SourceIP:  "203.0.113.42",
			UserAgent: "AWS Console",
		},
	}
}

// CreateTestUnmanagedAlert creates a test unmanaged resource alert
func CreateTestUnmanagedAlert() types.UnmanagedResourceAlert {
	return types.UnmanagedResourceAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged-instance",
		ResourceARN:  "arn:aws:ec2:us-east-1:123456789012:instance/i-unmanaged-instance",
		Region:       "us-east-1",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "developer@example.com",
		},
		EventMetadata: types.EventMetadata{
			EventTime: "2025-11-20T10:35:00Z",
			EventName: "RunInstances",
		},
		ImportCommand: "terraform import aws_instance.unmanaged i-unmanaged-instance",
	}
}

// MockFalcoEvent creates a mock Falco event
func MockFalcoEvent() map[string]interface{} {
	return map[string]interface{}{
		"source":   "aws_cloudtrail",
		"rule":     "EC2 Termination Protection Disabled",
		"priority": "CRITICAL",
		"output":   "EC2 termination protection disabled (user=admin instance=i-1234567890abcdef0)",
		"output_fields": map[string]string{
			"ct.name":                          "ModifyInstanceAttribute",
			"ct.request.instanceid":            "i-1234567890abcdef0",
			"ct.request.disableapitermination": "false",
			"ct.user":                          "admin@example.com",
			"ct.user.type":                     "IAMUser",
			"ct.user.arn":                      "arn:aws:iam::123456789012:user/admin",
			"ct.region":                        "us-east-1",
			"ct.account":                       "123456789012",
		},
	}
}

// RequireNoError is a helper that fails the test if error is not nil
func RequireNoError(t *testing.T, err error, msgAndArgs ...interface{}) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}

// AssertEventually retries assertion until it passes or timeout
func AssertEventually(t *testing.T, assertion func() bool, timeout, interval int) bool {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			if assertion() {
				return true
			}
		}
	}
}
