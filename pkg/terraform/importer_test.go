package terraform

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImporter(t *testing.T) {
	importer := NewImporter("/tmp/terraform", true)

	assert.NotNil(t, importer)
	assert.Equal(t, "terraform", importer.terraformBinary)
	assert.Equal(t, "/tmp/terraform", importer.workingDir)
	assert.True(t, importer.dryRun)
}

func TestGenerateResourceName(t *testing.T) {
	importer := NewImporter(".", false)

	tests := []struct {
		name       string
		resourceID string
		want       string
	}{
		{
			name:       "EC2 Instance ID",
			resourceID: "i-1234567890abcdef0",
			want:       "i_1234567890abcdef0",
		},
		{
			name:       "S3 Bucket Name",
			resourceID: "my-test-bucket",
			want:       "my_test_bucket",
		},
		{
			name:       "IAM Role ARN",
			resourceID: "arn:aws:iam::123456789012:role/MyRole",
			want:       "arn_aws_iam__123456789012_role_MyRole",
		},
		{
			name:       "RDS Instance with dots",
			resourceID: "my.database.instance",
			want:       "my_database_instance",
		},
		{
			name:       "Security Group starting with number",
			resourceID: "sg-0123456789",
			want:       "sg_0123456789",
		},
		{
			name:       "Resource starting with number",
			resourceID: "123-resource",
			want:       "r_123_resource",
		},
		{
			name:       "Long resource name (over 64 chars)",
			resourceID: "very-long-resource-name-that-exceeds-sixty-four-characters-limit-and-should-be-truncated",
			// After replacing hyphens with underscores, the string becomes longer than 64 chars
			// The function truncates to exactly 64 characters
			want: func() string {
				result := strings.ReplaceAll("very-long-resource-name-that-exceeds-sixty-four-characters-limit-and-should-be-truncated", "-", "_")
				if len(result) > 64 {
					result = result[:64]
				}
				return result
			}(),
		},
		{
			name:       "Resource with multiple special chars",
			resourceID: "test-resource:2024/01/01.backup",
			want:       "test_resource_2024_01_01_backup",
		},
		{
			name:       "Empty string",
			resourceID: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := importer.generateResourceName(tt.resourceID)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateImportCommand(t *testing.T) {
	importer := NewImporter(".", false)

	tests := []struct {
		name         string
		resourceType string
		resourceID   string
		wantType     string
		wantID       string
	}{
		{
			name:         "EC2 Instance",
			resourceType: "aws_instance",
			resourceID:   "i-1234567890abcdef0",
			wantType:     "aws_instance",
			wantID:       "i-1234567890abcdef0",
		},
		{
			name:         "S3 Bucket",
			resourceType: "aws_s3_bucket",
			resourceID:   "my-bucket",
			wantType:     "aws_s3_bucket",
			wantID:       "my-bucket",
		},
		{
			name:         "IAM Role",
			resourceType: "aws_iam_role",
			resourceID:   "arn:aws:iam::123456789012:role/MyRole",
			wantType:     "aws_iam_role",
			wantID:       "arn:aws:iam::123456789012:role/MyRole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := importer.GenerateImportCommand(tt.resourceType, tt.resourceID)

			assert.NotNil(t, cmd)
			assert.Equal(t, tt.wantType, cmd.ResourceType)
			assert.Equal(t, tt.wantID, cmd.ResourceID)
			assert.NotEmpty(t, cmd.ResourceName)
		})
	}
}

func TestImportCommand_String(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *ImportCommand
		wantContains []string
	}{
		{
			name: "EC2 Instance",
			cmd: &ImportCommand{
				ResourceType: "aws_instance",
				ResourceName: "web",
				ResourceID:   "i-1234567890abcdef0",
			},
			wantContains: []string{
				"terraform import",
				"aws_instance.web",
				"i-1234567890abcdef0",
			},
		},
		{
			name: "S3 Bucket",
			cmd: &ImportCommand{
				ResourceType: "aws_s3_bucket",
				ResourceName: "data_bucket",
				ResourceID:   "my-data-bucket",
			},
			wantContains: []string{
				"terraform import",
				"aws_s3_bucket.data_bucket",
				"my-data-bucket",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cmd.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, got, want)
			}
		})
	}
}

func TestGenerateTerraformCode(t *testing.T) {
	importer := NewImporter(".", false)

	tests := []struct {
		name            string
		resourceType    string
		resourceName    string
		attributes      map[string]interface{}
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "Simple EC2 Instance",
			resourceType: "aws_instance",
			resourceName: "web",
			attributes: map[string]interface{}{
				"ami":           "ami-12345678",
				"instance_type": "t3.micro",
			},
			wantContains: []string{
				`resource "aws_instance" "web"`,
				`ami = "ami-12345678"`,
				`instance_type = "t3.micro"`,
			},
		},
		{
			name:         "Resource with boolean",
			resourceType: "aws_s3_bucket",
			resourceName: "data",
			attributes: map[string]interface{}{
				"bucket":         "my-bucket",
				"force_destroy":  true,
				"versioning_enabled": false,
			},
			wantContains: []string{
				`resource "aws_s3_bucket" "data"`,
				`bucket = "my-bucket"`,
				`force_destroy = true`,
				`versioning_enabled = false`,
			},
		},
		{
			name:         "Resource with numbers",
			resourceType: "aws_db_instance",
			resourceName: "main",
			attributes: map[string]interface{}{
				"identifier":       "mydb",
				"allocated_storage": 100,
				"port":             3306,
				"backup_retention": float64(7),
			},
			wantContains: []string{
				`resource "aws_db_instance" "main"`,
				`identifier = "mydb"`,
				`allocated_storage = 100`,
				`port = 3306`,
				`backup_retention = 7`,
			},
		},
		{
			name:         "Resource with complex types",
			resourceType: "aws_security_group",
			resourceName: "allow_http",
			attributes: map[string]interface{}{
				"name":        "allow-http",
				"description": "Allow HTTP traffic",
				"ingress": []map[string]interface{}{
					{"from_port": 80, "to_port": 80, "protocol": "tcp"},
				},
				"tags": map[string]string{"Environment": "production"},
			},
			wantContains: []string{
				`resource "aws_security_group" "allow_http"`,
				`name = "allow-http"`,
				`description = "Allow HTTP traffic"`,
				`# ingress = <complex value>`,
				`# tags = <complex value>`,
			},
		},
		{
			name:         "Empty attributes",
			resourceType: "aws_vpc",
			resourceName: "main",
			attributes:   map[string]interface{}{},
			wantContains: []string{
				`resource "aws_vpc" "main"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := importer.GenerateTerraformCode(tt.resourceType, tt.resourceName, tt.attributes)

			for _, want := range tt.wantContains {
				assert.Contains(t, got, want, "Generated code should contain: %s", want)
			}

			for _, notWant := range tt.wantNotContains {
				assert.NotContains(t, got, notWant, "Generated code should not contain: %s", notWant)
			}

			// Verify it's valid HCL-like syntax
			assert.True(t, strings.HasPrefix(got, "# Auto-generated"))
			assert.Contains(t, got, "resource \"")
			assert.Contains(t, got, " {\n")
			assert.True(t, strings.HasSuffix(strings.TrimSpace(got), "}"))
		})
	}
}

func TestExecute_DryRun(t *testing.T) {
	importer := NewImporter(".", true) // dry-run enabled

	cmd := &ImportCommand{
		ResourceType: "aws_instance",
		ResourceName: "test",
		ResourceID:   "i-123",
	}

	ctx := context.Background()
	err := importer.Execute(ctx, cmd)

	assert.NoError(t, err, "Dry-run should not produce errors")
}

func TestAutoImport_DryRun(t *testing.T) {
	importer := NewImporter(".", true) // dry-run enabled

	ctx := context.Background()
	attributes := map[string]interface{}{
		"ami":           "ami-12345678",
		"instance_type": "t3.micro",
	}

	result := importer.AutoImport(ctx, "aws_instance", "i-1234567890abcdef0", attributes)

	// In dry-run mode, it should succeed without actual terraform execution
	assert.NotNil(t, result)
	assert.NotNil(t, result.Command)
	assert.Equal(t, "aws_instance", result.Command.ResourceType)
	assert.Equal(t, "i-1234567890abcdef0", result.Command.ResourceID)
	assert.NotEmpty(t, result.GeneratedCode)
}

func TestBatchImport_DryRun(t *testing.T) {
	importer := NewImporter(".", true) // dry-run enabled

	ctx := context.Background()
	resources := []struct {
		ResourceType string
		ResourceID   string
		Attributes   map[string]interface{}
	}{
		{
			ResourceType: "aws_instance",
			ResourceID:   "i-111",
			Attributes: map[string]interface{}{
				"ami": "ami-111",
			},
		},
		{
			ResourceType: "aws_instance",
			ResourceID:   "i-222",
			Attributes: map[string]interface{}{
				"ami": "ami-222",
			},
		},
		{
			ResourceType: "aws_s3_bucket",
			ResourceID:   "my-bucket",
			Attributes: map[string]interface{}{
				"bucket": "my-bucket",
			},
		},
	}

	results := importer.BatchImport(ctx, resources)

	assert.Len(t, results, 3)

	for i, result := range results {
		assert.NotNil(t, result, "Result %d should not be nil", i)
		assert.NotNil(t, result.Command, "Result %d command should not be nil", i)
		assert.Equal(t, resources[i].ResourceType, result.Command.ResourceType)
		assert.Equal(t, resources[i].ResourceID, result.Command.ResourceID)
	}
}

func TestImportResult_Structure(t *testing.T) {
	result := &ImportResult{
		Success: true,
		Command: &ImportCommand{
			ResourceType: "aws_instance",
			ResourceName: "web",
			ResourceID:   "i-123",
		},
		Output:        "Import successful",
		GeneratedCode: `resource "aws_instance" "web" {}`,
		Error:         nil,
	}

	assert.True(t, result.Success)
	assert.NotNil(t, result.Command)
	assert.Equal(t, "aws_instance", result.Command.ResourceType)
	assert.NotEmpty(t, result.Output)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.NoError(t, result.Error)
}

func TestGenerateTerraformCode_AttributeTypes(t *testing.T) {
	importer := NewImporter(".", false)

	t.Run("String attributes", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"name": "my-instance",
		})
		assert.Contains(t, code, `name = "my-instance"`)
	})

	t.Run("Integer attributes", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"count": 5,
		})
		assert.Contains(t, code, `count = 5`)
	})

	t.Run("Int64 attributes", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"storage": int64(1000),
		})
		assert.Contains(t, code, `storage = 1000`)
	})

	t.Run("Float64 attributes", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"cpu_credits": float64(0.5),
		})
		assert.Contains(t, code, `cpu_credits = 0.5`)
	})

	t.Run("Boolean true", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"enabled": true,
		})
		assert.Contains(t, code, `enabled = true`)
	})

	t.Run("Boolean false", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"disabled": false,
		})
		assert.Contains(t, code, `disabled = false`)
	})

	t.Run("Complex slice type", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"security_groups": []string{"sg-1", "sg-2"},
		})
		assert.Contains(t, code, `# security_groups = <complex value>`)
	})

	t.Run("Complex map type", func(t *testing.T) {
		code := importer.GenerateTerraformCode("aws_instance", "test", map[string]interface{}{
			"tags": map[string]string{"env": "prod"},
		})
		assert.Contains(t, code, `# tags = <complex value>`)
	})
}

func TestGenerateResourceName_EdgeCases(t *testing.T) {
	importer := NewImporter(".", false)

	t.Run("Only special characters", func(t *testing.T) {
		name := importer.generateResourceName("---:::///..")
		assert.Equal(t, "___________", name)
	})

	t.Run("Single character", func(t *testing.T) {
		name := importer.generateResourceName("a")
		assert.Equal(t, "a", name)
	})

	t.Run("Single number", func(t *testing.T) {
		name := importer.generateResourceName("1")
		assert.Equal(t, "r_1", name)
	})

	t.Run("Exactly 64 characters", func(t *testing.T) {
		longID := strings.Repeat("a", 64)
		name := importer.generateResourceName(longID)
		assert.Equal(t, 64, len(name))
	})

	t.Run("65 characters", func(t *testing.T) {
		longID := strings.Repeat("a", 65)
		name := importer.generateResourceName(longID)
		assert.Equal(t, 64, len(name))
	})

	t.Run("Mixed valid and invalid chars", func(t *testing.T) {
		name := importer.generateResourceName("valid_name-123:test/item.end")
		assert.Equal(t, "valid_name_123_test_item_end", name)
	})
}

func TestBatchImport_EmptyList(t *testing.T) {
	importer := NewImporter(".", true)

	ctx := context.Background()
	resources := []struct {
		ResourceType string
		ResourceID   string
		Attributes   map[string]interface{}
	}{}

	results := importer.BatchImport(ctx, resources)

	assert.NotNil(t, results)
	assert.Len(t, results, 0)
}

func TestImportCommand_StringFormat(t *testing.T) {
	tests := []struct {
		name string
		cmd  *ImportCommand
		want string
	}{
		{
			name: "Standard format",
			cmd: &ImportCommand{
				ResourceType: "aws_instance",
				ResourceName: "web",
				ResourceID:   "i-123",
			},
			want: "terraform import aws_instance.web i-123",
		},
		{
			name: "With special characters in ID",
			cmd: &ImportCommand{
				ResourceType: "aws_iam_role",
				ResourceName: "admin_role",
				ResourceID:   "arn:aws:iam::123456789012:role/AdminRole",
			},
			want: "terraform import aws_iam_role.admin_role arn:aws:iam::123456789012:role/AdminRole",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cmd.String()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAutoImport_ValidationFailure(t *testing.T) {
	// Use a non-existent working directory to trigger validation failure
	importer := NewImporter("/nonexistent/directory/for/testing/123456", false)

	ctx := context.Background()
	attributes := map[string]interface{}{
		"ami": "ami-12345678",
	}

	result := importer.AutoImport(ctx, "aws_instance", "i-test", attributes)

	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "validation failed")
}

func TestAutoImport_Success_DryRun(t *testing.T) {
	importer := NewImporter(".", true) // dry-run mode

	ctx := context.Background()
	attributes := map[string]interface{}{
		"ami":           "ami-abc123",
		"instance_type": "t3.small",
		"monitoring":    true,
	}

	result := importer.AutoImport(ctx, "aws_instance", "i-success-test", attributes)

	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NoError(t, result.Error)
	assert.NotNil(t, result.Command)
	assert.Equal(t, "aws_instance", result.Command.ResourceType)
	assert.Equal(t, "i-success-test", result.Command.ResourceID)
	assert.NotEmpty(t, result.GeneratedCode)
	assert.Contains(t, result.GeneratedCode, "aws_instance")
	assert.Contains(t, result.GeneratedCode, "ami-abc123")
}

func TestValidateImport_TerraformNotFound(t *testing.T) {
	// Create importer with non-existent binary
	importer := &Importer{
		terraformBinary: "/nonexistent/terraform/binary/path/123456",
		workingDir:      ".",
		dryRun:          false,
	}

	cmd := &ImportCommand{
		ResourceType: "aws_instance",
		ResourceName: "test",
		ResourceID:   "i-123",
	}

	ctx := context.Background()
	err := importer.ValidateImport(ctx, cmd)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform binary not found")
}

func TestExecute_NonDryRun_InvalidDirectory(t *testing.T) {
	// Use non-existent directory to ensure command fails
	importer := &Importer{
		terraformBinary: "terraform",
		workingDir:      "/nonexistent/directory/12345678",
		dryRun:          false,
	}

	cmd := &ImportCommand{
		ResourceType: "aws_instance",
		ResourceName: "test",
		ResourceID:   "i-123",
	}

	ctx := context.Background()
	err := importer.Execute(ctx, cmd)

	// Should fail because working directory doesn't exist
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform import failed")
}
