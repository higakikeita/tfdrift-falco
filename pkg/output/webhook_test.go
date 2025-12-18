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

// ============================================================================
// Edge Case and Error Handling Tests
// ============================================================================

func TestWebhookOutput_NilEvent(t *testing.T) {
	config := WebhookConfig{
		URL: "https://example.com/webhook",
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send nil event
	err := webhook.Write(nil)
	// Should handle nil gracefully (may error or ignore)
	_ = err
}

func TestWebhookOutput_EmptyURL(t *testing.T) {
	config := WebhookConfig{
		URL: "",
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should fail with empty URL
	assert.Error(t, err)
}

func TestWebhookOutput_InvalidURL(t *testing.T) {
	config := WebhookConfig{
		URL: "not-a-valid-url",
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should fail with invalid URL
	assert.Error(t, err)
}

func TestWebhookOutput_VeryLongURL(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create webhook with very long URL (with query parameters)
	longQuery := ""
	for i := 0; i < 100; i++ {
		if i > 0 {
			longQuery += "&"
		}
		longQuery += "param" + string(rune(i)) + "=value" + string(rune(i))
	}

	config := WebhookConfig{
		URL: server.URL + "?" + longQuery,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)
	// Should handle long URLs (may succeed or fail depending on limits)
	_ = err
}

func TestWebhookOutput_UnicodeInPayload(t *testing.T) {
	var receivedEvent types.DriftEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&receivedEvent)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL: server.URL,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Event with Unicode characters
	event := types.NewDriftEvent("gcp", "gcp_compute_instance", "インスタンス-12345", types.ChangeTypeModified).
		WithUser("ユーザー@example.com").
		WithRegion("asia-northeast1")

	err := webhook.Write(event)
	require.NoError(t, err)
	assert.Contains(t, receivedEvent.ResourceID, "インスタンス")
}

func TestWebhookOutput_HTTP4xxErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"400 Bad Request", http.StatusBadRequest},
		{"401 Unauthorized", http.StatusUnauthorized},
		{"403 Forbidden", http.StatusForbidden},
		{"404 Not Found", http.StatusNotFound},
		{"409 Conflict", http.StatusConflict},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			config := WebhookConfig{
				URL:        server.URL,
				MaxRetries: 0, // No retries for faster test
			}
			webhook := NewWebhookOutput(config)
			defer webhook.Close()

			event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
			err := webhook.Write(event)

			// Should fail with 4xx errors
			assert.Error(t, err)
		})
	}
}

func TestWebhookOutput_HTTP5xxErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"502 Bad Gateway", http.StatusBadGateway},
		{"503 Service Unavailable", http.StatusServiceUnavailable},
		{"504 Gateway Timeout", http.StatusGatewayTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			config := WebhookConfig{
				URL:        server.URL,
				MaxRetries: 1,
				RetryDelay: 10 * time.Millisecond,
			}
			webhook := NewWebhookOutput(config)
			defer webhook.Close()

			event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
			err := webhook.Write(event)

			// Should fail with 5xx errors after retries
			assert.Error(t, err)
		})
	}
}

func TestWebhookOutput_ConnectionRefused(t *testing.T) {
	// Use a port that's not listening
	config := WebhookConfig{
		URL:        "http://localhost:9999",
		MaxRetries: 0,
		Timeout:    100 * time.Millisecond,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should fail with connection refused
	assert.Error(t, err)
}

func TestWebhookOutput_ConcurrentWrites(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL: server.URL,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send multiple events concurrently
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-"+string(rune(id)), types.ChangeTypeCreated)
			done <- webhook.Write(event)
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// All requests should have been received
	assert.Equal(t, numGoroutines, requestCount)
}

func TestWebhookOutput_MultipleConsecutiveWrites(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL: server.URL,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	// Send multiple events sequentially
	for i := 0; i < 5; i++ {
		event := types.NewDriftEvent("aws", "aws_instance", "i-"+string(rune(i)), types.ChangeTypeCreated)
		err := webhook.Write(event)
		require.NoError(t, err)
	}

	assert.Equal(t, 5, requestCount)
}

func TestWebhookOutput_CloseIdempotent(t *testing.T) {
	config := WebhookConfig{
		URL: "https://example.com/webhook",
	}
	webhook := NewWebhookOutput(config)

	// Close multiple times should be safe
	err1 := webhook.Close()
	err2 := webhook.Close()
	err3 := webhook.Close()

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
}

func TestWebhookOutput_CustomContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-custom", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL:         server.URL,
		ContentType: "application/x-custom",
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)
	require.NoError(t, err)
}

func TestWebhookOutput_ZeroMaxRetries(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL:        server.URL,
		MaxRetries: 0, // MaxRetries=0 means use default (3)
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)

	// Should fail after default max retries (1 initial attempt + 3 retries = 4 total)
	assert.Error(t, err)
	assert.Equal(t, 4, attempts)
}

func TestFormatSlackPayload_AllSeverities(t *testing.T) {
	severities := []string{
		types.SeverityCritical,
		types.SeverityHigh,
		types.SeverityMedium,
		types.SeverityLow,
	}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated).
				WithSeverity(severity)

			payload := FormatSlackPayload(event)
			assert.NotNil(t, payload)
			assert.Contains(t, payload, "attachments")

			attachments := payload["attachments"].([]map[string]interface{})
			assert.Len(t, attachments, 1)
			assert.Contains(t, attachments[0], "color")
		})
	}
}

func TestFormatSlackPayload_AllChangeTypes(t *testing.T) {
	changeTypes := []string{
		types.ChangeTypeCreated,
		types.ChangeTypeModified,
		types.ChangeTypeDeleted,
	}

	for _, changeType := range changeTypes {
		t.Run(changeType, func(t *testing.T) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-12345", changeType)

			payload := FormatSlackPayload(event)
			assert.NotNil(t, payload)

			attachments := payload["attachments"].([]map[string]interface{})
			text := attachments[0]["text"].(string)
			assert.Contains(t, text, changeType)
		})
	}
}

func TestFormatTeamsPayload_AllSeverities(t *testing.T) {
	severities := []string{
		types.SeverityCritical,
		types.SeverityHigh,
		types.SeverityMedium,
		types.SeverityLow,
	}

	for _, severity := range severities {
		t.Run(severity, func(t *testing.T) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated).
				WithSeverity(severity)

			payload := FormatTeamsPayload(event)
			assert.NotNil(t, payload)
			assert.Equal(t, "MessageCard", payload["@type"])
			assert.Contains(t, payload, "themeColor")
		})
	}
}

func TestFormatSlackPayload_MinimalEvent(t *testing.T) {
	// Event with only required fields
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)

	payload := FormatSlackPayload(event)
	assert.NotNil(t, payload)
	assert.Contains(t, payload, "attachments")

	attachments := payload["attachments"].([]map[string]interface{})
	assert.Len(t, attachments, 1)
}

func TestFormatSlackPayload_CompleteEvent(t *testing.T) {
	// Event with all optional fields
	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified).
		WithSeverity(types.SeverityCritical).
		WithRegion("us-west-2").
		WithUser("admin@example.com").
		WithCloudTrailEvent("AuthorizeSecurityGroupIngress", "req-123")

	payload := FormatSlackPayload(event)
	assert.NotNil(t, payload)

	attachments := payload["attachments"].([]map[string]interface{})
	text := attachments[0]["text"].(string)

	assert.Contains(t, text, "aws_security_group")
	assert.Contains(t, text, "sg-12345")
	assert.Contains(t, text, "us-west-2")
	assert.Contains(t, text, "admin@example.com")
	assert.Contains(t, text, "AuthorizeSecurityGroupIngress")
}

func TestSendToSlack_InvalidURL(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := SendToSlack("not-a-valid-url", event)
	assert.Error(t, err)
}

func TestSendToSlack_EmptyURL(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := SendToSlack("", event)
	assert.Error(t, err)
}

func TestSendToTeams_InvalidURL(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := SendToTeams("not-a-valid-url", event)
	assert.Error(t, err)
}

func TestSendToTeams_EmptyURL(t *testing.T) {
	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := SendToTeams("", event)
	assert.Error(t, err)
}

func TestWebhookOutput_AllProviders(t *testing.T) {
	providers := []string{"aws", "gcp", "azure"}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL: server.URL,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			event := types.NewDriftEvent(provider, provider+"_resource", "resource-123", types.ChangeTypeCreated)
			err := webhook.Write(event)
			assert.NoError(t, err)
		})
	}
}

func TestWebhookOutput_RetryBackoff(t *testing.T) {
	attempts := 0
	var attemptTimes []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		attemptTimes = append(attemptTimes, time.Now())

		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := WebhookConfig{
		URL:        server.URL,
		MaxRetries: 3,
		RetryDelay: 50 * time.Millisecond,
	}
	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
	err := webhook.Write(event)
	require.NoError(t, err)

	// Verify attempts were made with delays
	assert.Equal(t, 3, attempts)
	assert.Len(t, attemptTimes, 3)

	// Verify delays between attempts (with some tolerance)
	for i := 1; i < len(attemptTimes); i++ {
		delay := attemptTimes[i].Sub(attemptTimes[i-1])
		assert.GreaterOrEqual(t, delay.Milliseconds(), int64(40)) // 50ms - 10ms tolerance
	}
}
