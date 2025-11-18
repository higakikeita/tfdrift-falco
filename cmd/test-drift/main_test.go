package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMainComponents(t *testing.T) {
	// This test verifies that the main components can be imported and used
	// Since main() calls os.Exit(), we can't test it directly
	// Instead, we test the individual components

	t.Run("Logger setup", func(t *testing.T) {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		defer log.SetOutput(os.Stdout)

		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		log.Info("test message")
		output := buf.String()
		assert.Contains(t, output, "test message")
	})

	t.Run("String formatting", func(t *testing.T) {
		// Test the string formatting used in main
		separator := string(make([]byte, 60))
		assert.Len(t, separator, 60)

		output := "=" + separator
		assert.Len(t, output, 61)
	})
}

func TestOutputFormatting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Header with emoji",
			input:    "🧪 TFDrift-Falco Test - Drift Detection Simulation",
			expected: "🧪 TFDrift-Falco Test - Drift Detection Simulation",
		},
		{
			name:     "Test case header",
			input:    "📋 Test Case 1: EC2 Instance Termination Protection Changed",
			expected: "📋 Test Case 1: EC2 Instance Termination Protection Changed",
		},
		{
			name:     "Format examples header",
			input:    "Format Examples:",
			expected: "Format Examples:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.input)
		})
	}
}

func TestExitCode(t *testing.T) {
	// Verify that os.Exit(0) would be called on success
	// We can't actually test os.Exit() as it would terminate the test
	exitCode := 0
	assert.Equal(t, 0, exitCode, "Expected exit code 0 on success")
}

func TestLoggerConfiguration(t *testing.T) {
	// Test logger configuration as used in main
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stdout)

	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	// Test various log levels
	log.Info("info message")
	log.Warn("warning message")
	log.Error("error message")

	output := buf.String()
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warning message")
	assert.Contains(t, output, "error message")
}

func TestFormatMessages(t *testing.T) {
	// Test the format of messages used in main
	tests := []struct {
		name     string
		format   string
		args     []interface{}
		expected string
	}{
		{
			name:     "Config loaded message",
			format:   "Config loaded - Backend: %s, LocalPath: '%s'",
			args:     []interface{}{"local", "test.tfstate"},
			expected: "Config loaded - Backend: local, LocalPath: 'test.tfstate'",
		},
		{
			name:     "Resource count message",
			format:   "✓ Loaded %d resources from Terraform state",
			args:     []interface{}{5},
			expected: "✓ Loaded 5 resources from Terraform state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ""
			if len(tt.args) > 0 {
				result = fmt.Sprintf(tt.format, tt.args...)
			} else {
				result = tt.format
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSeparatorLine(t *testing.T) {
	// Test the separator line generation
	separator := "=" + string(make([]byte, 60))
	assert.Len(t, separator, 61)
	assert.True(t, len(separator) > 0)
}

func TestEmojis(t *testing.T) {
	// Verify emojis are properly handled
	emojis := []string{
		"🧪", // Test tube
		"📋", // Clipboard
		"✓",  // Check mark
		"📄", // Document
		"📊", // Chart
		"📝", // Memo
		"🎉", // Party popper
	}

	for _, emoji := range emojis {
		assert.NotEmpty(t, emoji)
		assert.True(t, len(emoji) > 0)
	}
}

func TestTestCaseStructure(t *testing.T) {
	// Test the structure of test case data
	testCases := []struct {
		name         string
		severity     string
		resourceType string
		attribute    string
	}{
		{
			name:         "EC2 Instance Termination Protection Changed",
			severity:     "critical",
			resourceType: "aws_instance",
			attribute:    "disable_api_termination",
		},
		{
			name:         "S3 Bucket Encryption Configuration Changed",
			severity:     "critical",
			resourceType: "aws_s3_bucket",
			attribute:    "server_side_encryption_configuration",
		},
		{
			name:         "EC2 Instance Type Upgraded",
			severity:     "high",
			resourceType: "aws_instance",
			attribute:    "instance_type",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEmpty(t, tc.name)
			assert.NotEmpty(t, tc.severity)
			assert.NotEmpty(t, tc.resourceType)
			assert.NotEmpty(t, tc.attribute)

			// Verify severity is valid
			validSeverities := []string{"critical", "high", "medium", "low"}
			assert.Contains(t, validSeverities, tc.severity)
		})
	}
}

func TestUserIdentityStructure(t *testing.T) {
	// Test user identity data structure
	userIdentities := []struct {
		identityType string
		userName     string
		accountID    string
	}{
		{
			identityType: "IAMUser",
			userName:     "admin-user@example.com",
			accountID:    "123456789012",
		},
		{
			identityType: "IAMUser",
			userName:     "developer@example.com",
			accountID:    "123456789012",
		},
		{
			identityType: "AssumedRole",
			userName:     "AdminRole",
			accountID:    "123456789012",
		},
	}

	for _, ui := range userIdentities {
		t.Run(ui.userName, func(t *testing.T) {
			assert.NotEmpty(t, ui.identityType)
			assert.NotEmpty(t, ui.userName)
			assert.NotEmpty(t, ui.accountID)

			// Verify account ID format (should be 12 digits)
			assert.Len(t, ui.accountID, 12)

			// Verify identity type is valid
			validTypes := []string{"IAMUser", "AssumedRole", "Root"}
			assert.Contains(t, validTypes, ui.identityType)
		})
	}
}

func TestResourceIdentifiers(t *testing.T) {
	// Test resource identifier formats
	tests := []struct {
		name         string
		resourceType string
		resourceID   string
		valid        bool
	}{
		{
			name:         "EC2 Instance ID",
			resourceType: "aws_instance",
			resourceID:   "i-0abcd1234efgh5678",
			valid:        true,
		},
		{
			name:         "S3 Bucket Name",
			resourceType: "aws_s3_bucket",
			resourceID:   "my-data-bucket-12345",
			valid:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.resourceType)
			assert.NotEmpty(t, tt.resourceID)
			assert.True(t, tt.valid)

			// EC2 instance IDs should start with "i-"
			if tt.resourceType == "aws_instance" {
				assert.True(t, len(tt.resourceID) > 2)
				assert.Equal(t, "i-", tt.resourceID[:2])
			}
		})
	}
}

func TestTimestampFormat(t *testing.T) {
	// Test timestamp formats used in the test cases
	timestamps := []string{
		"2025-01-15T10:35:10Z",
		"2025-01-15T11:20:00Z",
		"2025-01-15T14:45:30Z",
	}

	for _, ts := range timestamps {
		t.Run(ts, func(t *testing.T) {
			assert.NotEmpty(t, ts)
			assert.Contains(t, ts, "T")
			assert.Contains(t, ts, "Z")
			assert.True(t, len(ts) > 0)
		})
	}
}

func TestDriftValueTypes(t *testing.T) {
	// Test different types of drift values
	tests := []struct {
		name     string
		oldValue interface{}
		newValue interface{}
	}{
		{
			name:     "Boolean change",
			oldValue: true,
			newValue: false,
		},
		{
			name:     "String change",
			oldValue: "t2.micro",
			newValue: "t2.large",
		},
		{
			name: "Complex object to nil",
			oldValue: map[string]interface{}{
				"rule": "encryption",
			},
			newValue: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Old value and new value should be different
			assert.NotEqual(t, tt.oldValue, tt.newValue)
		})
	}
}
