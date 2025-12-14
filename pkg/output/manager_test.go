package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	manager := NewManager(OutputModeJSON)
	assert.NotNil(t, manager)
	assert.Equal(t, OutputModeJSON, manager.mode)
}

func TestManager_EmitDriftEvent_JSONMode(t *testing.T) {
	manager := NewManager(OutputModeJSON)

	var jsonBuf bytes.Buffer
	manager.SetJSONWriter(&jsonBuf)

	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified)
	err := manager.EmitDriftEvent(event)
	require.NoError(t, err)

	jsonOutput := jsonBuf.String()
	assert.Contains(t, jsonOutput, "terraform_drift_detected")
	assert.Contains(t, jsonOutput, "aws_security_group")
	assert.Contains(t, jsonOutput, "sg-12345")
}

func TestManager_EmitDriftEvent_HumanMode(t *testing.T) {
	manager := NewManager(OutputModeHuman)

	var humanBuf bytes.Buffer
	manager.SetHumanWriter(&humanBuf)

	event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated).
		WithRegion("us-west-2").
		WithUser("admin@example.com")

	err := manager.EmitDriftEvent(event)
	require.NoError(t, err)

	humanOutput := humanBuf.String()
	assert.Contains(t, humanOutput, "aws_instance")
	assert.Contains(t, humanOutput, "i-12345")
	assert.Contains(t, humanOutput, "created")
	assert.Contains(t, humanOutput, "us-west-2")
	assert.Contains(t, humanOutput, "admin@example.com")
}

func TestManager_EmitDriftEvent_BothMode(t *testing.T) {
	manager := NewManager(OutputModeBoth)

	var jsonBuf, humanBuf bytes.Buffer
	manager.SetJSONWriter(&jsonBuf)
	manager.SetHumanWriter(&humanBuf)

	event := types.NewDriftEvent("aws", "aws_db_instance", "db-12345", types.ChangeTypeDeleted)
	err := manager.EmitDriftEvent(event)
	require.NoError(t, err)

	// Both outputs should have content
	jsonOutput := jsonBuf.String()
	assert.Contains(t, jsonOutput, "terraform_drift_detected")
	assert.Contains(t, jsonOutput, "db-12345")

	humanOutput := humanBuf.String()
	assert.Contains(t, humanOutput, "db-12345")
	assert.Contains(t, humanOutput, "deleted")
}

func TestManager_FormatHumanMessage(t *testing.T) {
	manager := NewManager(OutputModeHuman)

	tests := []struct {
		name     string
		event    *types.DriftEvent
		contains []string
	}{
		{
			name:  "basic event",
			event: types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated),
			contains: []string{
				"aws_instance",
				"i-12345",
				"created",
			},
		},
		{
			name: "event with region",
			event: types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified).
				WithRegion("us-east-1"),
			contains: []string{
				"aws_security_group",
				"sg-12345",
				"modified",
				"us-east-1",
			},
		},
		{
			name: "event with user",
			event: types.NewDriftEvent("aws", "aws_iam_role", "role-admin", types.ChangeTypeModified).
				WithUser("admin@example.com"),
			contains: []string{
				"aws_iam_role",
				"role-admin",
				"admin@example.com",
			},
		},
		{
			name: "event with CloudTrail",
			event: types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified).
				WithCloudTrailEvent("AuthorizeSecurityGroupIngress", "req-123"),
			contains: []string{
				"aws_security_group",
				"sg-12345",
				"AuthorizeSecurityGroupIngress",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := manager.formatHumanMessage(tt.event)
			for _, expected := range tt.contains {
				assert.Contains(t, msg, expected)
			}
		})
	}
}

func TestManager_GetSeverityEmoji(t *testing.T) {
	manager := NewManager(OutputModeHuman)

	tests := []struct {
		severity string
		expected string
	}{
		{types.SeverityCritical, "🚨"},
		{types.SeverityHigh, "⚠️ "},
		{types.SeverityMedium, "📊"},
		{types.SeverityLow, "ℹ️ "},
		{types.SeverityInfo, "💡"},
		{"unknown", "❓"},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			emoji := manager.getSeverityEmoji(tt.severity)
			assert.Equal(t, tt.expected, emoji)
		})
	}
}

func TestManager_SetMode(t *testing.T) {
	manager := NewManager(OutputModeHuman)
	assert.Equal(t, OutputModeHuman, manager.mode)

	manager.SetMode(OutputModeJSON)
	assert.Equal(t, OutputModeJSON, manager.mode)

	manager.SetMode(OutputModeBoth)
	assert.Equal(t, OutputModeBoth, manager.mode)
}

func TestParseOutputMode(t *testing.T) {
	tests := []struct {
		input    string
		expected OutputMode
		wantErr  bool
	}{
		{"human", OutputModeHuman, false},
		{"json", OutputModeJSON, false},
		{"both", OutputModeBoth, false},
		{"invalid", OutputModeHuman, true},
		{"", OutputModeHuman, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mode, err := ParseOutputMode(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, mode)
			}
		})
	}
}

func TestManager_Close(t *testing.T) {
	manager := NewManager(OutputModeJSON)
	err := manager.Close()
	assert.NoError(t, err)
}

func TestManager_Concurrent(t *testing.T) {
	manager := NewManager(OutputModeBoth)

	var jsonBuf, humanBuf bytes.Buffer
	manager.SetJSONWriter(&jsonBuf)
	manager.SetHumanWriter(&humanBuf)

	done := make(chan bool)
	count := 10

	for i := 0; i < count; i++ {
		go func(id int) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-"+string(rune(id)), types.ChangeTypeCreated)
			err := manager.EmitDriftEvent(event)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}

	// Both outputs should have written all events
	jsonLines := strings.Split(strings.TrimSpace(jsonBuf.String()), "\n")
	humanLines := strings.Split(strings.TrimSpace(humanBuf.String()), "\n")

	assert.Len(t, jsonLines, count)
	assert.Len(t, humanLines, count)
}
