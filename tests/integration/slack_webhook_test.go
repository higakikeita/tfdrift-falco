//go:build ignore
// +build ignore

package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/notifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSlackWebhook_SendAlert(t *testing.T) {
	var receivedPayload map[string]interface{}

	// Create mock Slack server
	mockServer := NewMockWebhookServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		assert.Equal(t, "POST", r.Method)

		// Verify Content-Type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Parse payload
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		err = json.Unmarshal(body, &receivedPayload)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	})
	defer mockServer.Close()

	// Create Slack notifier with mock URL
	cfg := config.SlackConfig{
		Enabled:    true,
		WebhookURL: mockServer.Server.URL,
		Channel:    "#test-alerts",
	}

	slackNotifier := notifier.NewSlackNotifier(cfg)

	// Create test alert
	alert := CreateTestAlert()

	// Send alert
	err := slackNotifier.Send(alert)
	require.NoError(t, err)

	// Verify payload
	assert.NotNil(t, receivedPayload)
	assert.Contains(t, receivedPayload, "text")
	assert.Contains(t, receivedPayload, "attachments")

	// Verify alert details in payload
	text := receivedPayload["text"].(string)
	assert.Contains(t, text, "drift", "Payload should mention drift")
	assert.Contains(t, text, "aws_instance", "Payload should contain resource type")
}

func TestSlackWebhook_SendUnmanagedAlert(t *testing.T) {
	var receivedPayload map[string]interface{}

	mockServer := NewMockWebhookServer(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedPayload)
		w.WriteHeader(http.StatusOK)
	})
	defer mockServer.Close()

	cfg := config.SlackConfig{
		Enabled:    true,
		WebhookURL: mockServer.Server.URL,
	}

	slackNotifier := notifier.NewSlackNotifier(cfg)

	// Create unmanaged resource alert
	alert := CreateTestUnmanagedAlert()

	// Send alert
	err := slackNotifier.SendUnmanaged(alert)
	require.NoError(t, err)

	// Verify payload
	assert.NotNil(t, receivedPayload)
	text := receivedPayload["text"].(string)
	assert.Contains(t, text, "unmanaged", "Payload should mention unmanaged resource")
	assert.Contains(t, text, "terraform import", "Payload should contain import command")
}

func TestSlackWebhook_RetryOnFailure(t *testing.T) {
	attemptCount := 0

	mockServer := NewMockWebhookServer(t, func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			// Fail first 2 attempts
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Succeed on 3rd attempt
			w.WriteHeader(http.StatusOK)
		}
	})
	defer mockServer.Close()

	cfg := config.SlackConfig{
		Enabled:    true,
		WebhookURL: mockServer.Server.URL,
	}

	slackNotifier := notifier.NewSlackNotifier(cfg)
	alert := CreateTestAlert()

	// Should eventually succeed after retries
	err := slackNotifier.Send(alert)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, attemptCount, 3, "Should have retried")
}

func TestSlackWebhook_InvalidURL(t *testing.T) {
	cfg := config.SlackConfig{
		Enabled:    true,
		WebhookURL: "invalid-url",
	}

	slackNotifier := notifier.NewSlackNotifier(cfg)
	alert := CreateTestAlert()

	// Should return error for invalid URL
	err := slackNotifier.Send(alert)
	assert.Error(t, err)
}

func TestSlackWebhook_Disabled(t *testing.T) {
	cfg := config.SlackConfig{
		Enabled:    false,
		WebhookURL: "http://example.com",
	}

	slackNotifier := notifier.NewSlackNotifier(cfg)
	alert := CreateTestAlert()

	// Should not send if disabled
	err := slackNotifier.Send(alert)
	assert.NoError(t, err) // Not an error, just skipped
}
