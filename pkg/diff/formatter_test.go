package diff

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name         string
		colorEnabled bool
	}{
		{"With colors", true},
		{"Without colors", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.colorEnabled)
			require.NotNil(t, formatter)
			assert.Equal(t, tt.colorEnabled, formatter.colorEnabled)
		})
	}
}

func TestFormatValue(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Nil value",
			value:    nil,
			expected: "null",
		},
		{
			name:     "String value",
			value:    "test-string",
			expected: "test-string",
		},
		{
			name:     "Integer value",
			value:    42,
			expected: "42",
		},
		{
			name:     "Boolean value",
			value:    true,
			expected: "true",
		},
		{
			name:     "Map value",
			value:    map[string]interface{}{"key": "value"},
			expected: "{\n  \"key\": \"value\"\n}",
		},
		{
			name:     "Slice value",
			value:    []string{"item1", "item2"},
			expected: "[\n  \"item1\",\n  \"item2\"\n]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsComplexType(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"Nil", nil, false},
		{"String", "text", false},
		{"Integer", 123, false},
		{"Boolean", true, false},
		{"Map", map[string]interface{}{}, true},
		{"Slice", []string{}, true},
		{"Struct", types.UserIdentity{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.isComplexType(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSeverityColor(t *testing.T) {
	formatter := NewFormatter(true)

	tests := []struct {
		severity string
		expected string
	}{
		{"critical", ColorRed},
		{"high", ColorYellow},
		{"medium", ColorBlue},
		{"low", ColorGreen},
		{"unknown", ColorGray},
	}

	for _, tt := range tests {
		t.Run(tt.severity, func(t *testing.T) {
			color := formatter.getSeverityColor(tt.severity)
			assert.Equal(t, tt.expected, color)
		})
	}
}

func TestColor(t *testing.T) {
	tests := []struct {
		name         string
		colorEnabled bool
		text         string
		colorCode    string
		expectColor  bool
	}{
		{
			name:         "Color enabled",
			colorEnabled: true,
			text:         "test",
			colorCode:    ColorRed,
			expectColor:  true,
		},
		{
			name:         "Color disabled",
			colorEnabled: false,
			text:         "test",
			colorCode:    ColorRed,
			expectColor:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.colorEnabled)
			result := formatter.color(tt.colorCode, tt.text)

			if tt.expectColor {
				assert.Contains(t, result, tt.colorCode)
				assert.Contains(t, result, ColorReset)
				assert.Contains(t, result, tt.text)
			} else {
				assert.Equal(t, tt.text, result)
			}
		})
	}
}

func TestFormatTerraformValue(t *testing.T) {
	formatter := NewFormatter(false)

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{
			name:     "Null",
			value:    nil,
			expected: "null",
		},
		{
			name:     "String",
			value:    "example",
			expected: "\"example\"",
		},
		{
			name:     "Boolean true",
			value:    true,
			expected: "true",
		},
		{
			name:     "Boolean false",
			value:    false,
			expected: "false",
		},
		{
			name:     "Integer",
			value:    123,
			expected: "123",
		},
		{
			name:     "Float",
			value:    123.45,
			expected: "123.45",
		},
		{
			name:     "Simple list",
			value:    []interface{}{"a", "b"},
			expected: "[\"a\", \"b\"]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.formatTerraformValue(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatConsole(t *testing.T) {
	formatter := NewFormatter(false) // Disable colors for easier testing

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-1234567890",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		UserIdentity: types.UserIdentity{
			UserName:  "john.doe",
			Type:      "IAMUser",
			ARN:       "arn:aws:iam::123456789012:user/john.doe",
			AccountID: "123456789012",
		},
		Timestamp:    "2025-01-15T10:00:00Z",
		MatchedRules: []string{"instance-type-change"},
	}

	result := formatter.FormatConsole(alert)

	// Verify key sections are present
	assert.Contains(t, result, "DRIFT DETECTED")
	assert.Contains(t, result, "aws_instance.web")
	assert.Contains(t, result, "HIGH")
	assert.Contains(t, result, "instance_type")
	assert.Contains(t, result, "t3.micro")
	assert.Contains(t, result, "t3.large")
	assert.Contains(t, result, "john.doe")
	assert.Contains(t, result, "2025-01-15T10:00:00Z")
	assert.Contains(t, result, "instance-type-change")
}

func TestFormatUnifiedDiff(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		ResourceType: "aws_s3_bucket",
		ResourceName: "data",
		OldValue:     "private",
		NewValue:     "public",
	}

	result := formatter.FormatUnifiedDiff(alert)

	assert.Contains(t, result, "--- terraform/aws_s3_bucket.data")
	assert.Contains(t, result, "+++ runtime/aws_s3_bucket.data")
	assert.Contains(t, result, "-private")
	assert.Contains(t, result, "+public")
}

func TestFormatMarkdown(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		Severity:     "critical",
		ResourceType: "aws_security_group",
		ResourceName: "app_sg",
		ResourceID:   "sg-123456",
		Attribute:    "ingress_rules",
		OldValue:     "restricted",
		NewValue:     "open",
		UserIdentity: types.UserIdentity{
			UserName:  "admin",
			AccountID: "123456789012",
		},
		Timestamp: "2025-01-15T12:00:00Z",
	}

	result := formatter.FormatMarkdown(alert)

	// Verify Markdown formatting
	assert.Contains(t, result, "## 🚨 Drift Detected:")
	assert.Contains(t, result, "`aws_security_group.app_sg`")
	assert.Contains(t, result, "**Severity:** 🔴 **CRITICAL**")
	assert.Contains(t, result, "**Changed Attribute:** `ingress_rules`")
	assert.Contains(t, result, "```diff")
	assert.Contains(t, result, "- restricted")
	assert.Contains(t, result, "+ open")
	assert.Contains(t, result, "**User:** admin")
	assert.Contains(t, result, "```hcl")
	assert.Contains(t, result, "- [ ]")
}

func TestFormatJSON(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		Severity:     "medium",
		ResourceType: "aws_instance",
		ResourceName: "app",
		ResourceID:   "i-abcdef123",
		Attribute:    "tags",
		OldValue:     map[string]string{"env": "dev"},
		NewValue:     map[string]string{"env": "prod"},
		UserIdentity: types.UserIdentity{
			UserName:    "user1",
			Type:        "IAMUser",
			ARN:         "arn:aws:iam::123:user/user1",
			AccountID:   "123456789012",
			PrincipalID: "AIDAI123",
		},
		Timestamp:    "2025-01-15T14:00:00Z",
		MatchedRules: []string{"tag-change"},
	}

	result, err := formatter.FormatJSON(alert)
	require.NoError(t, err)

	// Verify it's valid JSON
	var jsonData map[string]interface{}
	err = json.Unmarshal([]byte(result), &jsonData)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "medium", jsonData["severity"])
	assert.Equal(t, "aws_instance", jsonData["resource_type"])
	assert.Equal(t, "app", jsonData["resource_name"])
	assert.Equal(t, "i-abcdef123", jsonData["resource_id"])
	assert.Equal(t, "tags", jsonData["attribute"])
	assert.NotNil(t, jsonData["change"])
	assert.NotNil(t, jsonData["user"])
	assert.NotNil(t, jsonData["matched_rules"])
}

func TestFormatSideBySide(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		OldValue: "old-value",
		NewValue: "new-value",
	}

	result := formatter.FormatSideBySide(alert)

	assert.Contains(t, result, "Terraform State")
	assert.Contains(t, result, "Actual Configuration")
	assert.Contains(t, result, "old-value")
	assert.Contains(t, result, "new-value")
	assert.Contains(t, result, "│") // Column separator
}

func TestFormatUnmanagedResource(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.UnmanagedResourceAlert{
		Severity:     "warning",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged",
		EventName:    "RunInstances",
		UserIdentity: types.UserIdentity{
			UserName:    "developer",
			ARN:         "arn:aws:iam::123:user/developer",
			PrincipalID: "AIDAI456",
		},
		Timestamp: "2025-01-15T15:00:00Z",
		Reason:    "Resource not found in Terraform state",
		Changes: map[string]interface{}{
			"instance_type": "t3.medium",
		},
	}

	result := formatter.FormatUnmanagedResource(alert)

	assert.Contains(t, result, "UNMANAGED RESOURCE DETECTED")
	assert.Contains(t, result, "WARNING")
	assert.Contains(t, result, "aws_instance")
	assert.Contains(t, result, "i-unmanaged")
	assert.Contains(t, result, "RunInstances")
	assert.Contains(t, result, "developer")
	assert.Contains(t, result, "Resource not found in Terraform state")
	assert.Contains(t, result, "terraform import")
}

func TestFormatUnmanagedResourceMarkdown(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.UnmanagedResourceAlert{
		Severity:     "high",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "unmanaged-bucket",
		EventName:    "CreateBucket",
		UserIdentity: types.UserIdentity{
			UserName: "admin",
			ARN:      "arn:aws:iam::123:user/admin",
		},
		Timestamp: "2025-01-15T16:00:00Z",
		Reason:    "Created outside Terraform workflow",
	}

	result := formatter.FormatUnmanagedResourceMarkdown(alert)

	assert.Contains(t, result, "UNMANAGED RESOURCE DETECTED")
	assert.Contains(t, result, "HIGH")
	assert.Contains(t, result, "`aws_s3_bucket`")
	assert.Contains(t, result, "`unmanaged-bucket`")
	assert.Contains(t, result, "CreateBucket")
	assert.Contains(t, result, "admin")
	assert.Contains(t, result, "```\nterraform import")
}

func TestIndentLines(t *testing.T) {
	formatter := NewFormatter(false)

	text := "line1\nline2\nline3"
	result := formatter.indentLines(text, 4, "")

	lines := strings.Split(result, "\n")
	assert.Equal(t, 3, len(lines))
	assert.True(t, strings.HasPrefix(lines[0], "    "))
	assert.True(t, strings.HasPrefix(lines[1], "    "))
	assert.True(t, strings.HasPrefix(lines[2], "    "))
}

func TestFormatValueChange_Simple(t *testing.T) {
	formatter := NewFormatter(false)

	result := formatter.formatValueChange("old", "new")

	assert.Contains(t, result, "old")
	assert.Contains(t, result, "new")
	assert.Contains(t, result, "→")
}

func TestFormatValueChange_Complex(t *testing.T) {
	formatter := NewFormatter(false)

	oldValue := map[string]interface{}{"key": "old"}
	newValue := map[string]interface{}{"key": "new"}

	result := formatter.formatValueChange(oldValue, newValue)

	assert.Contains(t, result, "Old Value")
	assert.Contains(t, result, "New Value")
}

func TestFormatTerraformResource(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceName: "test",
		Attribute:    "instance_type",
	}

	result := formatter.formatTerraformResource(alert, "t3.micro")

	assert.Contains(t, result, "resource \"aws_instance\" \"test\"")
	assert.Contains(t, result, "instance_type = \"t3.micro\"")
}

func TestFormatRecommendations(t *testing.T) {
	formatter := NewFormatter(false)

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceName: "web",
	}

	result := formatter.formatRecommendations(alert)

	assert.Contains(t, result, "Review the change")
	assert.Contains(t, result, "terraform plan")
	assert.Contains(t, result, "terraform apply")
	assert.Contains(t, result, "-target=aws_instance.web")
}
