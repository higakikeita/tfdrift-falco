package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONOutput_Write(t *testing.T) {
	var buf bytes.Buffer
	output := NewJSONOutputWithWriter(&buf)

	event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified).
		WithRegion("us-west-2").
		WithUser("admin@example.com")

	err := output.Write(event)
	require.NoError(t, err)

	result := buf.String()
	assert.Contains(t, result, "terraform_drift_detected")
	assert.Contains(t, result, "aws_security_group")
	assert.Contains(t, result, "sg-12345")
	assert.Contains(t, result, "us-west-2")
	assert.Contains(t, result, "admin@example.com")
	// Should end with newline (NDJSON format)
	assert.True(t, strings.HasSuffix(result, "\n"))
}

func TestJSONOutput_WriteMultiple(t *testing.T) {
	var buf bytes.Buffer
	output := NewJSONOutputWithWriter(&buf)

	event1 := types.NewDriftEvent("aws", "aws_instance", "i-111", types.ChangeTypeCreated)
	event2 := types.NewDriftEvent("aws", "aws_db_instance", "db-222", types.ChangeTypeModified)
	event3 := types.NewDriftEvent("aws", "aws_security_group", "sg-333", types.ChangeTypeDeleted)

	require.NoError(t, output.Write(event1))
	require.NoError(t, output.Write(event2))
	require.NoError(t, output.Write(event3))

	result := buf.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should have 3 JSON lines
	assert.Len(t, lines, 3)

	// Each line should be valid JSON with correct content
	assert.Contains(t, lines[0], "i-111")
	assert.Contains(t, lines[1], "db-222")
	assert.Contains(t, lines[2], "sg-333")
}

func TestJSONOutput_Concurrent(t *testing.T) {
	var buf bytes.Buffer
	output := NewJSONOutputWithWriter(&buf)

	done := make(chan bool)
	count := 10

	for i := 0; i < count; i++ {
		go func(id int) {
			event := types.NewDriftEvent("aws", "aws_instance", "i-"+string(rune(id)), types.ChangeTypeCreated)
			err := output.Write(event)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	for i := 0; i < count; i++ {
		<-done
	}

	result := buf.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")

	// Should have written all events
	assert.Len(t, lines, count)
}

func TestNewJSONOutput(t *testing.T) {
	output := NewJSONOutput()
	assert.NotNil(t, output)
	assert.NotNil(t, output.writer)
}
