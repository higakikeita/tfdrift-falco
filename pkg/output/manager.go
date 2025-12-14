package output

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// OutputMode defines the output mode
type OutputMode string

const (
	// OutputModeHuman outputs human-readable logs only
	OutputModeHuman OutputMode = "human"
	// OutputModeJSON outputs JSON events only (NDJSON)
	OutputModeJSON OutputMode = "json"
	// OutputModeBoth outputs both human-readable and JSON
	OutputModeBoth OutputMode = "both"
)

// Manager manages multiple output writers for drift events
type Manager struct {
	mode       OutputMode
	jsonOutput *JSONOutput
	humanOut   io.Writer
	mu         sync.Mutex
}

// NewManager creates a new output manager
func NewManager(mode OutputMode) *Manager {
	return &Manager{
		mode:       mode,
		jsonOutput: NewJSONOutput(),
		humanOut:   os.Stderr, // Human logs to stderr
	}
}

// EmitDriftEvent emits a drift event to all configured outputs
func (m *Manager) EmitDriftEvent(event *types.DriftEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// Emit JSON output
	if m.mode == OutputModeJSON || m.mode == OutputModeBoth {
		if err := m.jsonOutput.Write(event); err != nil {
			errors = append(errors, fmt.Errorf("JSON output error: %w", err))
		}
	}

	// Emit human-readable output
	if m.mode == OutputModeHuman || m.mode == OutputModeBoth {
		humanMsg := m.formatHumanMessage(event)
		if _, err := fmt.Fprintln(m.humanOut, humanMsg); err != nil {
			errors = append(errors, fmt.Errorf("human output error: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("output errors: %v", errors)
	}

	return nil
}

// formatHumanMessage formats a drift event as a human-readable log message
func (m *Manager) formatHumanMessage(event *types.DriftEvent) string {
	// Format: [SEVERITY] Drift detected: resource_type (resource_id) - change_type
	severity := m.getSeverityEmoji(event.Severity)
	msg := fmt.Sprintf("%s Drift detected: %s (%s) - %s",
		severity,
		event.ResourceType,
		event.ResourceID,
		event.ChangeType)

	// Add region if present
	if event.Region != "" {
		msg += fmt.Sprintf(" [%s]", event.Region)
	}

	// Add user if present
	if event.User != "" {
		msg += fmt.Sprintf(" by %s", event.User)
	}

	// Add CloudTrail event if present
	if event.CloudTrailEvent != "" {
		msg += fmt.Sprintf(" (CloudTrail: %s)", event.CloudTrailEvent)
	}

	return msg
}

// getSeverityEmoji returns an emoji for the severity level
func (m *Manager) getSeverityEmoji(severity string) string {
	switch severity {
	case types.SeverityCritical:
		return "🚨"
	case types.SeverityHigh:
		return "⚠️ "
	case types.SeverityMedium:
		return "📊"
	case types.SeverityLow:
		return "ℹ️ "
	case types.SeverityInfo:
		return "💡"
	default:
		return "❓"
	}
}

// SetMode sets the output mode
func (m *Manager) SetMode(mode OutputMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.mode = mode
}

// SetHumanWriter sets the writer for human-readable output
func (m *Manager) SetHumanWriter(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.humanOut = w
}

// SetJSONWriter sets the writer for JSON output
func (m *Manager) SetJSONWriter(w io.Writer) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jsonOutput = NewJSONOutputWithWriter(w)
}

// Close closes all outputs
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	if err := m.jsonOutput.Close(); err != nil {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		return fmt.Errorf("close errors: %v", errors)
	}

	return nil
}

// ParseOutputMode parses a string into OutputMode
func ParseOutputMode(mode string) (OutputMode, error) {
	switch mode {
	case "human":
		return OutputModeHuman, nil
	case "json":
		return OutputModeJSON, nil
	case "both":
		return OutputModeBoth, nil
	default:
		return OutputModeHuman, fmt.Errorf("invalid output mode: %s (valid: human, json, both)", mode)
	}
}

// LogDriftEvent is a convenience function to log a drift event using the logger
func LogDriftEvent(event *types.DriftEvent) {
	fields := log.Fields{
		"event_type":    event.EventType,
		"resource_type": event.ResourceType,
		"resource_id":   event.ResourceID,
		"change_type":   event.ChangeType,
		"severity":      event.Severity,
	}

	if event.Region != "" {
		fields["region"] = event.Region
	}
	if event.User != "" {
		fields["user"] = event.User
	}
	if event.CloudTrailEvent != "" {
		fields["cloudtrail_event"] = event.CloudTrailEvent
	}

	msg := fmt.Sprintf("Drift detected: %s (%s) - %s",
		event.ResourceType,
		event.ResourceID,
		event.ChangeType)

	switch event.Severity {
	case types.SeverityCritical, types.SeverityHigh:
		log.WithFields(fields).Warn(msg)
	case types.SeverityMedium:
		log.WithFields(fields).Info(msg)
	default:
		log.WithFields(fields).Debug(msg)
	}
}
