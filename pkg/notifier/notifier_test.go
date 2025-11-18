package notifier

import (
	"encoding/json"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/testutil"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled: true,
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.formatter)
}

func TestSend_Slack(t *testing.T) {
	mockServer := testutil.NewMockHTTPServer()
	defer mockServer.Close()

	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled:    true,
			WebhookURL: mockServer.URL(),
			Channel:    "#alerts",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	err = manager.Send(alert)
	assert.NoError(t, err)

	// Verify request was sent
	assert.Equal(t, 1, mockServer.GetRequestCount())

	// Verify request body
	body := mockServer.GetLastRequestBody()
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(body), &payload)
	require.NoError(t, err)

	assert.Equal(t, "#alerts", payload["channel"])
	assert.Contains(t, payload, "text")
	assert.Contains(t, payload, "blocks")
}

func TestSend_Discord(t *testing.T) {
	mockServer := testutil.NewMockHTTPServer()
	defer mockServer.Close()

	cfg := config.NotificationsConfig{
		Discord: config.DiscordConfig{
			Enabled:    true,
			WebhookURL: mockServer.URL(),
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	err = manager.Send(alert)
	assert.NoError(t, err)

	// Verify request was sent
	assert.Equal(t, 1, mockServer.GetRequestCount())

	// Verify request body
	body := mockServer.GetLastRequestBody()
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(body), &payload)
	require.NoError(t, err)

	assert.Contains(t, payload, "embeds")
	embeds := payload["embeds"].([]interface{})
	assert.Len(t, embeds, 1)
}

func TestSend_FalcoOutput(t *testing.T) {
	cfg := config.NotificationsConfig{
		FalcoOutput: config.FalcoOutputConfig{
			Enabled:  true,
			Priority: "Warning",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	// FalcoOutput logs to stdout, so no error expected
	err = manager.Send(alert)
	assert.NoError(t, err)
}

func TestSend_MultipleChannels(t *testing.T) {
	slackServer := testutil.NewMockHTTPServer()
	defer slackServer.Close()

	discordServer := testutil.NewMockHTTPServer()
	defer discordServer.Close()

	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled:    true,
			WebhookURL: slackServer.URL(),
			Channel:    "#alerts",
		},
		Discord: config.DiscordConfig{
			Enabled:    true,
			WebhookURL: discordServer.URL(),
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	err = manager.Send(alert)
	assert.NoError(t, err)

	// Verify both channels received notifications
	assert.Equal(t, 1, slackServer.GetRequestCount())
	assert.Equal(t, 1, discordServer.GetRequestCount())
}

func TestSend_ErrorHandling(t *testing.T) {
	mockServer := testutil.NewMockHTTPServer()
	mockServer.SetStatusCode(500) // Simulate server error
	defer mockServer.Close()

	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled:    true,
			WebhookURL: mockServer.URL(),
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	err = manager.Send(alert)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "notification errors")
}

func TestSend_NoChannelsEnabled(t *testing.T) {
	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled: false,
		},
		Discord: config.DiscordConfig{
			Enabled: false,
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := testutil.CreateTestDriftAlert()

	err = manager.Send(alert)
	assert.NoError(t, err) // No error when no channels enabled
}

func TestFormatSlackMessage(t *testing.T) {
	manager, err := NewManager(config.NotificationsConfig{})
	require.NoError(t, err)

	tests := []struct {
		name     string
		alert    *types.DriftAlert
		validate func(t *testing.T, blocks []map[string]interface{})
	}{
		{
			name: "Critical Severity",
			alert: &types.DriftAlert{
				Severity:     "critical",
				ResourceType: "aws_instance",
				ResourceName: "web",
				Attribute:    "instance_type",
				OldValue:     "t3.micro",
				NewValue:     "t3.large",
				ResourceID:   "i-123",
				Timestamp:    "2025-11-18T10:00:00Z",
				UserIdentity: types.UserIdentity{
					UserName: "admin",
				},
			},
			validate: func(t *testing.T, blocks []map[string]interface{}) {
				assert.Len(t, blocks, 3)

				// Check header
				header := blocks[0]
				assert.Equal(t, "header", header["type"])
				text := header["text"].(map[string]string)
				assert.Contains(t, text["text"], ":rotating_light:")
				assert.Contains(t, text["text"], "aws_instance.web")

				// Check fields section
				section := blocks[1]
				assert.Equal(t, "section", section["type"])
			},
		},
		{
			name: "Low Severity",
			alert: &types.DriftAlert{
				Severity:     "low",
				ResourceType: "aws_s3_bucket",
				ResourceName: "data",
				Attribute:    "tags",
				OldValue:     map[string]string{"env": "dev"},
				NewValue:     map[string]string{"env": "prod"},
				ResourceID:   "bucket-123",
				Timestamp:    "2025-11-18T10:00:00Z",
				UserIdentity: types.UserIdentity{
					UserName: "user",
				},
			},
			validate: func(t *testing.T, blocks []map[string]interface{}) {
				assert.Len(t, blocks, 3)

				header := blocks[0]
				text := header["text"].(map[string]string)
				assert.Contains(t, text["text"], ":information_source:")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := manager.formatSlackMessage(tt.alert)
			tt.validate(t, blocks)
		})
	}
}

func TestSendDiscord_EmbedFormat(t *testing.T) {
	mockServer := testutil.NewMockHTTPServer()
	defer mockServer.Close()

	cfg := config.NotificationsConfig{
		Discord: config.DiscordConfig{
			Enabled:    true,
			WebhookURL: mockServer.URL(),
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	tests := []struct {
		name     string
		alert    *types.DriftAlert
		validate func(t *testing.T, embed map[string]interface{})
	}{
		{
			name: "Critical Alert",
			alert: &types.DriftAlert{
				Severity:     "critical",
				ResourceType: "aws_security_group",
				ResourceName: "main",
				Attribute:    "ingress",
				OldValue:     []string{"80"},
				NewValue:     []string{"80", "443", "22"},
				ResourceID:   "sg-123",
				UserIdentity: types.UserIdentity{
					UserName: "attacker",
				},
				Timestamp: "2025-11-18T10:00:00Z",
			},
			validate: func(t *testing.T, embed map[string]interface{}) {
				// JSON unmarshal converts numbers to float64
				assert.Equal(t, float64(0xFF0000), embed["color"]) // Red
				assert.Contains(t, embed["title"], "aws_security_group.main")
				fields := embed["fields"].([]interface{})
				assert.Len(t, fields, 4)
			},
		},
		{
			name: "High Alert",
			alert: &types.DriftAlert{
				Severity:     "high",
				ResourceType: "aws_iam_role",
				ResourceName: "admin",
				Attribute:    "assume_role_policy",
				OldValue:     "old_policy",
				NewValue:     "new_policy",
				ResourceID:   "role-123",
				UserIdentity: types.UserIdentity{
					UserName: "user",
				},
				Timestamp: "2025-11-18T10:00:00Z",
			},
			validate: func(t *testing.T, embed map[string]interface{}) {
				assert.Equal(t, float64(0xFF8C00), embed["color"]) // Orange
			},
		},
		{
			name: "Medium Alert",
			alert: &types.DriftAlert{
				Severity:     "medium",
				ResourceType: "aws_s3_bucket",
				ResourceName: "logs",
				Attribute:    "versioning",
				OldValue:     "enabled",
				NewValue:     "disabled",
				ResourceID:   "bucket-logs",
				UserIdentity: types.UserIdentity{
					UserName: "ops",
				},
				Timestamp: "2025-11-18T10:00:00Z",
			},
			validate: func(t *testing.T, embed map[string]interface{}) {
				assert.Equal(t, float64(0xFFFF00), embed["color"]) // Yellow
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer.Reset()

			err := manager.sendDiscord(tt.alert)
			assert.NoError(t, err)

			body := mockServer.GetLastRequestBody()
			var payload map[string]interface{}
			err = json.Unmarshal([]byte(body), &payload)
			require.NoError(t, err)

			embeds := payload["embeds"].([]interface{})
			embed := embeds[0].(map[string]interface{})

			tt.validate(t, embed)
		})
	}
}

func TestMapSeverityToPriority(t *testing.T) {
	manager, _ := NewManager(config.NotificationsConfig{})

	tests := []struct {
		severity string
		want     string
	}{
		{"critical", "Critical"},
		{"high", "Error"},
		{"medium", "Warning"},
		{"low", "Notice"},
		{"unknown", "Info"},
		{"", "Info"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			got := manager.mapSeverityToPriority(tt.severity)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSendWebhook_InvalidURL(t *testing.T) {
	manager, _ := NewManager(config.NotificationsConfig{})

	payload := map[string]string{"test": "data"}
	err := manager.sendWebhook("http://invalid-url-that-does-not-exist:99999", payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send webhook")
}

func TestSendWebhook_InvalidJSON(t *testing.T) {
	manager, _ := NewManager(config.NotificationsConfig{})

	// Create a payload that cannot be marshaled to JSON
	payload := make(map[string]interface{})
	payload["invalid"] = make(chan int) // channels cannot be marshaled

	err := manager.sendWebhook("http://example.com", payload)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to marshal payload")
}

func TestSendSlack_Integration(t *testing.T) {
	mockServer := testutil.NewMockHTTPServer()
	defer mockServer.Close()

	cfg := config.NotificationsConfig{
		Slack: config.SlackConfig{
			Enabled:    true,
			WebhookURL: mockServer.URL(),
			Channel:    "#security-alerts",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_security_group",
		ResourceName: "production",
		Attribute:    "ingress",
		OldValue:     "restricted",
		NewValue:     "0.0.0.0/0",
		ResourceID:   "sg-prod-123",
		Timestamp:    "2025-11-18T10:00:00Z",
		UserIdentity: types.UserIdentity{
			UserName: "developer",
			ARN:      "arn:aws:iam::123456789012:user/developer",
		},
		MatchedRules: []string{"security_group_rule"},
	}

	err = manager.sendSlack(alert)
	assert.NoError(t, err)

	// Verify the webhook was called
	assert.Equal(t, 1, mockServer.GetRequestCount())

	// Verify content type
	lastReq := mockServer.GetLastRequest()
	assert.Equal(t, "application/json", lastReq.Header.Get("Content-Type"))

	// Verify payload structure
	body := mockServer.GetLastRequestBody()
	var payload map[string]interface{}
	err = json.Unmarshal([]byte(body), &payload)
	require.NoError(t, err)

	assert.Equal(t, "#security-alerts", payload["channel"])
	assert.NotEmpty(t, payload["text"])

	blocks := payload["blocks"].([]interface{})
	assert.Greater(t, len(blocks), 0)
}

func TestSendFalcoOutput_Format(t *testing.T) {
	cfg := config.NotificationsConfig{
		FalcoOutput: config.FalcoOutputConfig{
			Enabled:  true,
			Priority: "Error",
		},
	}

	manager, err := NewManager(cfg)
	require.NoError(t, err)

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web-server",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.2xlarge",
		ResourceID:   "i-prod-123",
		Timestamp:    "2025-11-18T10:00:00Z",
		UserIdentity: types.UserIdentity{
			UserName: "admin",
		},
	}

	// Just verify it doesn't error (output goes to log)
	err = manager.sendFalcoOutput(alert)
	assert.NoError(t, err)
}
