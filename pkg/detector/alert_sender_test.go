package detector

import (
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/notifier"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendAlert_DryRun(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
	}

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
	}

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-0cea65ac652556767",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic in dry-run mode
	assert.NotPanics(t, func() {
		detector.sendAlert(alert)
	})
}

func TestSendUnmanagedResourceAlert_DryRun(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
	}

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic in dry-run mode
	assert.NotPanics(t, func() {
		detector.sendUnmanagedResourceAlert(event)
	})
}

func TestSendAlert_WithNotifier(t *testing.T) {
	cfg := &config.Config{
		DryRun: false, // Not dry-run
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled:    true,
				WebhookURL: "http://invalid-url",
				Channel:    "#test",
			},
		},
	}

	formatter := diff.NewFormatter(false)
	// Create notifier (will fail to send but shouldn't panic)
	notifierMgr, err := notifier.NewManager(cfg.Notifications)
	require.NoError(t, err)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
		notifier:  notifierMgr,
	}

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-123",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic even if notification fails
	assert.NotPanics(t, func() {
		detector.sendAlert(alert)
	})
}

func TestSendUnmanagedResourceAlert_NotifierError(t *testing.T) {
	cfg := &config.Config{
		DryRun: false, // Not dry-run, so it will try to send notifications
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled:    true,
				WebhookURL: "https://invalid-webhook-url-that-will-fail.example.com/webhook",
				Channel:    "#test",
			},
		},
	}

	notifierMgr, err := notifier.NewManager(cfg.Notifications)
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:       cfg,
		notifier:  notifierMgr,
		formatter: formatter,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged-123",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "test-user",
			ARN:      "arn:aws:iam::123456789012:user/test-user",
		},
		RawEvent: map[string]interface{}{
			"eventTime": time.Now().Format(time.RFC3339),
		},
	}

	// Should not panic even if notifier fails
	assert.NotPanics(t, func() {
		detector.sendUnmanagedResourceAlert(event)
	})
}
