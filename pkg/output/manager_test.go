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
	manager := NewManager(ModeJSON)
	assert.NotNil(t, manager)
	assert.Equal(t, ModeJSON, manager.mode)
}

func TestManager_EmitDriftEvent_JSONMode(t *testing.T) {
	manager := NewManager(ModeJSON)

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
	manager := NewManager(ModeHuman)

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
	manager := NewManager(ModeBoth)

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
	manager := NewManager(ModeHuman)

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
	manager := NewManager(ModeHuman)

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
	manager := NewManager(ModeHuman)
	assert.Equal(t, ModeHuman, manager.mode)

	manager.SetMode(ModeJSON)
	assert.Equal(t, ModeJSON, manager.mode)

	manager.SetMode(ModeBoth)
	assert.Equal(t, ModeBoth, manager.mode)
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		input    string
		expected Mode
		wantErr  bool
	}{
		{"human", ModeHuman, false},
		{"json", ModeJSON, false},
		{"both", ModeBoth, false},
		{"invalid", ModeHuman, true},
		{"", ModeHuman, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			mode, err := ParseMode(tt.input)
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
	manager := NewManager(ModeJSON)
	err := manager.Close()
	assert.NoError(t, err)
}

func TestManager_Concurrent(t *testing.T) {
	manager := NewManager(ModeBoth)

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
