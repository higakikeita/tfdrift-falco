package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTempDir(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	// Verify directory exists
	_, err := os.Stat(dir)
	assert.NoError(t, err)

	// Verify directory is writable
	testFile := filepath.Join(dir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)
}

func TestWriteTestFile(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	content := "test content"
	filename := "test.txt"

	path := WriteTestFile(t, dir, filename, content)

	// Verify file exists
	_, err := os.Stat(path)
	require.NoError(t, err)

	// Verify content
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, content, string(data))
}

func TestWriteTestFile_EmptyContent(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	path := WriteTestFile(t, dir, "empty.txt", "")

	// Verify file exists
	_, err := os.Stat(path)
	require.NoError(t, err)

	// Verify empty content
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Empty(t, string(data))
}

func TestCreateTestConfig_DefaultValues(t *testing.T) {
	cfg1 := CreateTestConfig()
	cfg2 := CreateTestConfig()

	// Both should be valid but independent
	assert.NotNil(t, cfg1)
	assert.NotNil(t, cfg2)
	assert.Equal(t, cfg1.Falco.Port, cfg2.Falco.Port)
}

func TestCreateTestConfig(t *testing.T) {
	cfg := CreateTestConfig()

	assert.NotNil(t, cfg)
	assert.True(t, cfg.Providers.AWS.Enabled)
	assert.NotEmpty(t, cfg.Providers.AWS.Regions)
	assert.True(t, cfg.Falco.Enabled)
	assert.NotEmpty(t, cfg.Falco.Hostname)
	assert.Greater(t, cfg.Falco.Port, uint16(0))
}

func TestCreateTestStateFile(t *testing.T) {
	dir, cleanup := CreateTempDir(t, "test-")
	defer cleanup()

	statePath := CreateTestStateFile(t, dir)

	// Verify file exists
	_, err := os.Stat(statePath)
	require.NoError(t, err)

	// Verify it's valid JSON
	data, err := os.ReadFile(statePath)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"version"`)
	assert.Contains(t, string(data), `"resources"`)
}

func TestCreateTestEvent(t *testing.T) {
	event := CreateTestEvent("aws_instance", "i-12345", "RunInstances")

	assert.NotNil(t, event)
	assert.Equal(t, "aws_instance", event.ResourceType)
	assert.Equal(t, "i-12345", event.ResourceID)
	assert.NotNil(t, event.Changes)
}

func TestCreateTestDriftAlert(t *testing.T) {
	alert := CreateTestDriftAlert()

	assert.NotNil(t, alert)
	assert.NotEmpty(t, alert.Severity)
	assert.NotEmpty(t, alert.ResourceType)
	assert.NotEmpty(t, alert.ResourceName)
	assert.NotEmpty(t, alert.ResourceID)
	assert.NotEmpty(t, alert.Attribute)
}

func TestCreateTestResource(t *testing.T) {
	resource := CreateTestResource("aws_instance", "test", "i-12345")

	assert.NotNil(t, resource)
	assert.Equal(t, "aws_instance", resource.Type)
	assert.Equal(t, "test", resource.Name)
	assert.NotNil(t, resource.Attributes)
}

func TestCreateTestUnmanagedAlert(t *testing.T) {
	alert := CreateTestUnmanagedAlert()

	assert.NotNil(t, alert)
	assert.NotEmpty(t, alert.Severity)
	assert.NotEmpty(t, alert.ResourceType)
	assert.NotEmpty(t, alert.ResourceID)
	assert.NotEmpty(t, alert.EventName)
}
