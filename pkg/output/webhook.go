package output

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// WebhookConfig contains webhook configuration
type WebhookConfig struct {
	URL         string            `yaml:"url" json:"url"`
	Method      string            `yaml:"method" json:"method"`           // POST, PUT (default: POST)
	Headers     map[string]string `yaml:"headers" json:"headers"`         // Custom headers
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`         // Request timeout (default: 10s)
	MaxRetries  int               `yaml:"max_retries" json:"max_retries"` // Max retry attempts (default: 3)
	RetryDelay  time.Duration     `yaml:"retry_delay" json:"retry_delay"` // Initial retry delay (default: 1s)
	ContentType string            `yaml:"content_type" json:"content_type"` // Content-Type header (default: application/json)
}

// WebhookOutput sends drift events to a webhook endpoint
type WebhookOutput struct {
	config WebhookConfig
	client *http.Client
	mu     sync.Mutex
}

// NewWebhookOutput creates a new webhook output
func NewWebhookOutput(config WebhookConfig) *WebhookOutput {
	// Set defaults
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.ContentType == "" {
		config.ContentType = "application/json"
	}

	return &WebhookOutput{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Write sends a drift event to the webhook endpoint
func (w *WebhookOutput) Write(event *types.DriftEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Serialize event to JSON
	jsonData, err := event.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Send with retries
	return w.sendWithRetry(jsonData)
}

// sendWithRetry sends the request with exponential backoff retry
func (w *WebhookOutput) sendWithRetry(jsonData []byte) error {
	var lastErr error

	for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: delay * 2^(attempt-1)
			delay := w.config.RetryDelay * time.Duration(1<<uint(attempt-1))
			log.Debugf("Webhook retry attempt %d/%d after %v", attempt, w.config.MaxRetries, delay)
			time.Sleep(delay)
		}

		err := w.send(jsonData)
		if err == nil {
			if attempt > 0 {
				log.Infof("Webhook succeeded after %d retries", attempt)
			}
			return nil
		}

		lastErr = err
		log.Warnf("Webhook attempt %d failed: %v", attempt+1, err)
	}

	return fmt.Errorf("webhook failed after %d attempts: %w", w.config.MaxRetries+1, lastErr)
}

// send performs a single HTTP request
func (w *WebhookOutput) send(jsonData []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), w.config.Timeout)
	defer cancel()

	// Create request
	req, err := http.NewRequestWithContext(ctx, w.config.Method, w.config.URL, bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set Content-Type
	req.Header.Set("Content-Type", w.config.ContentType)

	// Set custom headers
	for key, value := range w.config.Headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Close closes the webhook output
func (w *WebhookOutput) Close() error {
	w.client.CloseIdleConnections()
	return nil
}

// WebhookPayload is a wrapper for drift events sent to webhooks
// Some systems (like Slack) expect a specific format
type WebhookPayload struct {
	Event *types.DriftEvent `json:"event"`
	Text  string            `json:"text,omitempty"`  // For Slack compatibility
	Title string            `json:"title,omitempty"` // For general use
}

// FormatSlackPayload formats a drift event as a Slack message
func FormatSlackPayload(event *types.DriftEvent) map[string]interface{} {
	severity := getSeverityColor(event.Severity)

	text := fmt.Sprintf("*Terraform Drift Detected*\n"+
		"Resource: `%s` (%s)\n"+
		"Change: %s\n"+
		"Severity: %s",
		event.ResourceType,
		event.ResourceID,
		event.ChangeType,
		event.Severity)

	if event.Region != "" {
		text += fmt.Sprintf("\nRegion: %s", event.Region)
	}
	if event.User != "" {
		text += fmt.Sprintf("\nUser: %s", event.User)
	}
	if event.CloudTrailEvent != "" {
		text += fmt.Sprintf("\nCloudTrail: %s", event.CloudTrailEvent)
	}

	return map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      severity,
				"text":       text,
				"footer":     "TFDrift-Falco",
				"footer_icon": "https://falco.org/img/brand/falco-logo.png",
				"ts":         event.DetectedAt.Unix(),
			},
		},
	}
}

// FormatTeamsPayload formats a drift event as a Microsoft Teams message
func FormatTeamsPayload(event *types.DriftEvent) map[string]interface{} {
	title := fmt.Sprintf("Terraform Drift Detected: %s", event.ResourceType)

	text := fmt.Sprintf("**Resource ID**: %s\n\n"+
		"**Change Type**: %s\n\n"+
		"**Severity**: %s",
		event.ResourceID,
		event.ChangeType,
		event.Severity)

	if event.Region != "" {
		text += fmt.Sprintf("\n\n**Region**: %s", event.Region)
	}
	if event.User != "" {
		text += fmt.Sprintf("\n\n**User**: %s", event.User)
	}

	return map[string]interface{}{
		"@type":    "MessageCard",
		"@context": "https://schema.org/extensions",
		"summary":  title,
		"title":    title,
		"text":     text,
		"themeColor": getTeamsColor(event.Severity),
	}
}

// getSeverityColor returns a color code for Slack attachments
func getSeverityColor(severity string) string {
	switch severity {
	case types.SeverityCritical:
		return "danger" // Red
	case types.SeverityHigh:
		return "warning" // Orange
	case types.SeverityMedium:
		return "#439FE0" // Blue
	case types.SeverityLow:
		return "good" // Green
	default:
		return "#808080" // Gray
	}
}

// getTeamsColor returns a color code for Microsoft Teams
func getTeamsColor(severity string) string {
	switch severity {
	case types.SeverityCritical:
		return "FF0000" // Red
	case types.SeverityHigh:
		return "FFA500" // Orange
	case types.SeverityMedium:
		return "0078D7" // Blue
	case types.SeverityLow:
		return "28A745" // Green
	default:
		return "808080" // Gray
	}
}

// SendToSlack is a convenience function to send a drift event to Slack
func SendToSlack(webhookURL string, event *types.DriftEvent) error {
	payload := FormatSlackPayload(event)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	config := WebhookConfig{
		URL:    webhookURL,
		Method: "POST",
	}

	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	return webhook.send(jsonData)
}

// SendToTeams is a convenience function to send a drift event to Microsoft Teams
func SendToTeams(webhookURL string, event *types.DriftEvent) error {
	payload := FormatTeamsPayload(event)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Teams payload: %w", err)
	}

	config := WebhookConfig{
		URL:    webhookURL,
		Method: "POST",
	}

	webhook := NewWebhookOutput(config)
	defer webhook.Close()

	return webhook.send(jsonData)
}
