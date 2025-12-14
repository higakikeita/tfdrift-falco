package output

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWebhookOutput(t *testing.T) {
	config := WebhookConfig{
		URL: "https://example.com/webhook",
	}

	webhook := NewWebhookOutput(config)
	assert.NotNil(t, webhook)
	assert.Equal(t, "POST", webhook.config.Method)
	assert.Equal(t, 10*time.Second, webhook.config.Timeout)
	assert.Equal(t, 3, webhook.config.MaxRetries)
	assert.Equal(t, "application/json", webhook.config.ContentType)
}

func TestWebhookOutput_Write_Success(t *testing.T) {
	// Create test server
	var receivedEvent types.DriftEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Decode event
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook
	config := WebhookConfig{
		URL: server.URL,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified)
	err := webhook.Write(event)
	require.NoError(t, err)

	// Verify received event
	assert.Equal(t, "aws_security_group", receivedEvent.ResourceType)
	assert.Equal(t, "sg-12345", receivedEvent.ResourceID)
}

func TestWebhookOutput_Write_CustomHeaders(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers
		assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with custom headers
	config := WebhookConfig{
		URL: server.URL,
		Headers: map[string]string{
			"Authorization":   "Bearer secret-token",
			"X-Custom-Header": "custom-value",
		},
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)
	require.NoError(t, err)
}

func TestWebhookOutput_Write_Retry(t *testing.T) {
	attempts := 0

	// Create test server that fails first 2 attempts
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with fast retries for testing
	config := WebhookConfig{
		URL:        server.URL,
		MaxRetries: 3,
		RetryDelay: 10 * time.Millisecond,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_db_instance", "db-12345", types.ChangeTypeDeleted)
	err := webhook.Write(event)
	require.NoError(t, err)

	// Should have retried twice before success
	assert.Equal(t, 3, attempts)
}

func TestWebhookOutput_Write_MaxRetriesExceeded(t *testing.T) {
	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Create webhook with fast retries for testing
	config := WebhookConfig{
		URL:        server.URL,
		MaxRetries: 2,
		RetryDelay: 10 * time.Millisecond,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should fail after max retries
	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook failed after")
}

func TestWebhookOutput_Write_Timeout(t *testing.T) {
	// Create test server that takes too long
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with short timeout
	config := WebhookConfig{
		URL:        server.URL,
		Timeout:    50 * time.Millisecond,
		MaxRetries: 0, // No retries for faster test
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should timeout
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestFormatSlackPayload(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified).
		WithSeverity(types.SeverityCritical).
		WithRegion("us-west-2").
		WithUser("admin@example.com").
		WithCloudTrailEvent("AuthorizeSecurityGroupIngress", "req-123")

	payload := FormatSlackPayload(event)

	assert.NotNil(t, payload)
	assert.Contains(t, payload, "attachments")

	attachments := payload["attachments"].([]map[string]interface{})
	assert.Len(t, attachments, 1)

	attachment := attachments[0]
	assert.Equal(t, "danger", attachment["color"]) // Critical = danger
	assert.Contains(t, attachment["text"], "aws_security_group")
	assert.Contains(t, attachment["text"], "sg-12345")
	assert.Contains(t, attachment["text"], "us-west-2")
	assert.Contains(t, attachment["text"], "admin@example.com")
}

func TestFormatTeamsPayload(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeDeleted).
		WithSeverity(types.SeverityHigh).
		WithRegion("ap-northeast-1")

	payload := FormatTeamsPayload(event)

	assert.NotNil(t, payload)
	assert.Equal(t, "MessageCard", payload["@type"])
	assert.Contains(t, payload["title"], "aws_instance")
	assert.Contains(t, payload["text"], "i-12345")
	assert.Contains(t, payload["text"], "ap-northeast-1")
	assert.Equal(t, "FFA500", payload["themeColor"]) // High = orange
}

func TestGetSeverityColor(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{types.SeverityCritical, "danger"},
		{types.SeverityHigh, "warning"},
		{types.SeverityMedium, "#439FE0"},
		{types.SeverityLow, "good"},
		{"unknown", "#808080"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			color := getSeverityColor(tt.severity)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestGetTeamsColor(t *testing.T) {
	tests := []struct {
		severity string
		expected string
	}{
		{types.SeverityCritical, "FF0000"},
		{types.SeverityHigh, "FFA500"},
		{types.SeverityMedium, "0078D7"},
		{types.SeverityLow, "28A745"},
		{"unknown", "808080"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			color := getTeamsColor(tt.severity)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestSendToSlack(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		// Verify Slack format
		assert.Contains(t, payload, "attachments")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified)
	err := SendToSlack(server.URL, event)
	require.NoError(t, err)
}

func TestSendToTeams(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&payload)
		require.NoError(t, err)

		// Verify Teams format
		assert.Equal(t, "MessageCard", payload["@type"])
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := SendToTeams(server.URL, event)
	require.NoError(t, err)
}

func TestWebhookOutput_CustomMethod(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with PUT method
	config := WebhookConfig{
		URL:    server.URL,
		Method: "PUT",
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send event
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)
	require.NoError(t, err)
}
