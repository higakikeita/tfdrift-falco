package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	log "github.com/sirupsen/logrus"
)

// Manager manages notification delivery
type Manager struct {
	cfg       config.NotificationsConfig
	formatter *diff.DiffFormatter
}

// NewManager creates a new notification manager
func NewManager(cfg config.NotificationsConfig) (*Manager, error) {
	formatter := diff.NewFormatter(false) // No colors for notifications

	return &Manager{
		cfg:       cfg,
		formatter: formatter,
	}, nil
}

// Send sends a drift alert to configured channels
func (m *Manager) Send(alert *detector.DriftAlert) error {
	var errors []error

	if m.cfg.Slack.Enabled {
		if err := m.sendSlack(alert); err != nil {
			errors = append(errors, fmt.Errorf("slack: %w", err))
		}
	}

	if m.cfg.Discord.Enabled {
		if err := m.sendDiscord(alert); err != nil {
			errors = append(errors, fmt.Errorf("discord: %w", err))
		}
	}

	if m.cfg.FalcoOutput.Enabled {
		if err := m.sendFalcoOutput(alert); err != nil {
			errors = append(errors, fmt.Errorf("falco: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("notification errors: %v", errors)
	}

	return nil
}

// sendSlack sends alert to Slack
func (m *Manager) sendSlack(alert *detector.DriftAlert) error {
	// Use Markdown formatter for Slack
	markdownText := m.formatter.FormatMarkdown(alert)

	// Also include Slack blocks for better formatting
	blocks := m.formatSlackMessage(alert)

	payload := map[string]interface{}{
		"channel": m.cfg.Slack.Channel,
		"text":    markdownText, // Fallback text
		"blocks":  blocks,
	}

	return m.sendWebhook(m.cfg.Slack.WebhookURL, payload)
}

// formatSlackMessage formats the alert for Slack
func (m *Manager) formatSlackMessage(alert *detector.DriftAlert) []map[string]interface{} {
	severityEmoji := map[string]string{
		"critical": ":rotating_light:",
		"high":     ":warning:",
		"medium":   ":large_orange_diamond:",
		"low":      ":information_source:",
	}

	emoji := severityEmoji[alert.Severity]

	return []map[string]interface{}{
		{
			"type": "header",
			"text": map[string]string{
				"type": "plain_text",
				"text": fmt.Sprintf("%s Drift Detected: %s.%s", emoji, alert.ResourceType, alert.ResourceName),
			},
		},
		{
			"type": "section",
			"fields": []map[string]string{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Attribute:*\n`%s`", alert.Attribute),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Severity:*\n`%s`", alert.Severity),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*Changed:*\n`%v` → `%v`", alert.OldValue, alert.NewValue),
				},
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("*User:*\n%s", alert.UserIdentity.UserName),
				},
			},
		},
		{
			"type": "context",
			"elements": []map[string]string{
				{
					"type": "mrkdwn",
					"text": fmt.Sprintf("Resource ID: `%s` | Time: %s", alert.ResourceID, alert.Timestamp),
				},
			},
		},
	}
}

// sendDiscord sends alert to Discord
func (m *Manager) sendDiscord(alert *detector.DriftAlert) error {
	severityColor := map[string]int{
		"critical": 0xFF0000, // Red
		"high":     0xFF8C00, // Orange
		"medium":   0xFFFF00, // Yellow
		"low":      0x00FF00, // Green
	}

	embed := map[string]interface{}{
		"title":       fmt.Sprintf("Drift Detected: %s.%s", alert.ResourceType, alert.ResourceName),
		"description": fmt.Sprintf("Attribute `%s` was modified", alert.Attribute),
		"color":       severityColor[alert.Severity],
		"fields": []map[string]interface{}{
			{
				"name":   "Old Value",
				"value":  fmt.Sprintf("`%v`", alert.OldValue),
				"inline": true,
			},
			{
				"name":   "New Value",
				"value":  fmt.Sprintf("`%v`", alert.NewValue),
				"inline": true,
			},
			{
				"name":  "User",
				"value": alert.UserIdentity.UserName,
			},
			{
				"name":  "Resource ID",
				"value": fmt.Sprintf("`%s`", alert.ResourceID),
			},
		},
		"timestamp": alert.Timestamp,
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{embed},
	}

	return m.sendWebhook(m.cfg.Discord.WebhookURL, payload)
}

// sendFalcoOutput sends alert as Falco-compatible output
func (m *Manager) sendFalcoOutput(alert *detector.DriftAlert) error {
	// Format as Falco JSON output
	falcoEvent := map[string]interface{}{
		"output": fmt.Sprintf("Terraform drift detected: %s.%s attribute %s changed from %v to %v (user=%s resource=%s)",
			alert.ResourceType, alert.ResourceName, alert.Attribute,
			alert.OldValue, alert.NewValue,
			alert.UserIdentity.UserName, alert.ResourceID),
		"priority":   m.mapSeverityToPriority(alert.Severity),
		"rule":       "Terraform Drift Detection",
		"time":       alert.Timestamp,
		"output_fields": map[string]interface{}{
			"resource.type":    alert.ResourceType,
			"resource.name":    alert.ResourceName,
			"resource.id":      alert.ResourceID,
			"drift.attribute":  alert.Attribute,
			"drift.old_value":  alert.OldValue,
			"drift.new_value":  alert.NewValue,
			"user.name":        alert.UserIdentity.UserName,
			"severity":         alert.Severity,
		},
	}

	// TODO: Send to Falco gRPC endpoint or write to stdout
	jsonData, _ := json.MarshalIndent(falcoEvent, "", "  ")
	log.Info(string(jsonData))

	return nil
}

// sendWebhook sends a generic webhook
func (m *Manager) sendWebhook(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned error status: %d", resp.StatusCode)
	}

	return nil
}

// mapSeverityToPriority maps severity to Falco priority
func (m *Manager) mapSeverityToPriority(severity string) string {
	mapping := map[string]string{
		"critical": "Critical",
		"high":     "Error",
		"medium":   "Warning",
		"low":      "Notice",
	}

	if priority, ok := mapping[severity]; ok {
		return priority
	}

	return "Info"
}
