package terraform

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Importer handles automatic Terraform import operations
type Importer struct {
	terraformBinary string
	workingDir      string
	dryRun          bool
}

// NewImporter creates a new Terraform importer
func NewImporter(workingDir string, dryRun bool) *Importer {
	return &Importer{
		terraformBinary: "terraform",
		workingDir:      workingDir,
		dryRun:          dryRun,
	}
}

// ImportCommand represents a Terraform import command
type ImportCommand struct {
	ResourceType string
	ResourceName string
	ResourceID   string
}

// GenerateImportCommand generates a terraform import command string
func (i *Importer) GenerateImportCommand(resourceType, resourceID string) *ImportCommand {
	// Generate a suggested resource name based on the resource ID
	resourceName := i.generateResourceName(resourceID)

	return &ImportCommand{
		ResourceType: resourceType,
		ResourceName: resourceName,
		ResourceID:   resourceID,
	}
}

// generateResourceName generates a Terraform resource name from a resource ID
func (i *Importer) generateResourceName(resourceID string) string {
	// Remove special characters and replace with underscores
	name := strings.ReplaceAll(resourceID, "-", "_")
	name = strings.ReplaceAll(name, ":", "_")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ".", "_")

	// Remove leading numbers if any
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "r_" + name
	}

	// Truncate if too long
	if len(name) > 64 {
		name = name[:64]
	}

	return name
}

// String returns the command as a string
func (cmd *ImportCommand) String() string {
	return fmt.Sprintf("terraform import %s.%s %s",
		cmd.ResourceType, cmd.ResourceName, cmd.ResourceID)
}

// Execute runs the terraform import command
func (i *Importer) Execute(ctx context.Context, cmd *ImportCommand) error {
	if i.dryRun {
		log.Infof("[DRY-RUN] Would execute: %s", cmd.String())
		return nil
	}

	log.Infof("Executing Terraform import: %s", cmd.String())

	// Build the command
	args := []string{
		"import",
		fmt.Sprintf("%s.%s", cmd.ResourceType, cmd.ResourceName),
		cmd.ResourceID,
	}

	execCmd := exec.CommandContext(ctx, i.terraformBinary, args...)
	execCmd.Dir = i.workingDir

	// Capture output
	var stdout, stderr bytes.Buffer
	execCmd.Stdout = &stdout
	execCmd.Stderr = &stderr

	// Execute
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("terraform import failed: %w\nstderr: %s", err, stderr.String())
	}

	log.Infof("Import successful: %s", stdout.String())
	return nil
}

// GenerateTerraformCode generates a basic Terraform resource block for the imported resource
func (i *Importer) GenerateTerraformCode(resourceType, resourceName string, attributes map[string]interface{}) string {
	var b strings.Builder

	b.WriteString("# Auto-generated resource block for import\n")
	b.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resourceType, resourceName))

	// Add basic attributes
	for key, value := range attributes {
		switch v := value.(type) {
		case string:
			b.WriteString(fmt.Sprintf("  %s = \"%s\"\n", key, v))
		case int, int64, float64:
			b.WriteString(fmt.Sprintf("  %s = %v\n", key, v))
		case bool:
			b.WriteString(fmt.Sprintf("  %s = %t\n", key, v))
		default:
			// Complex types - add placeholder
			b.WriteString(fmt.Sprintf("  # %s = <complex value>\n", key))
		}
	}

	b.WriteString("}\n")

	return b.String()
}

// ValidateImport checks if the import would be successful without actually importing
func (i *Importer) ValidateImport(ctx context.Context, _ *ImportCommand) error {
	// Check if terraform binary exists
	if _, err := exec.LookPath(i.terraformBinary); err != nil {
		return fmt.Errorf("terraform binary not found: %w", err)
	}

	// Check if working directory exists and is initialized
	checkCmd := exec.CommandContext(ctx, i.terraformBinary, "version")
	checkCmd.Dir = i.workingDir

	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("terraform not properly initialized in %s: %w", i.workingDir, err)
	}

	return nil
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	Success       bool
	Command       *ImportCommand
	Output        string
	Error         error
	GeneratedCode string
}

// AutoImport attempts to automatically import a resource
func (i *Importer) AutoImport(ctx context.Context, resourceType, resourceID string, attributes map[string]interface{}) *ImportResult {
	result := &ImportResult{
		Success: false,
	}

	// Generate import command
	cmd := i.GenerateImportCommand(resourceType, resourceID)
	result.Command = cmd

	// Validate
	if err := i.ValidateImport(ctx, cmd); err != nil {
		result.Error = fmt.Errorf("validation failed: %w", err)
		return result
	}

	// Execute import
	if err := i.Execute(ctx, cmd); err != nil {
		result.Error = err
		return result
	}

	// Generate Terraform code
	result.GeneratedCode = i.GenerateTerraformCode(resourceType, cmd.ResourceName, attributes)
	result.Success = true

	return result
}

// BatchImport imports multiple resources at once
func (i *Importer) BatchImport(ctx context.Context, resources []struct {
	ResourceType string
	ResourceID   string
	Attributes   map[string]interface{}
}) []*ImportResult {
	results := make([]*ImportResult, 0, len(resources))

	for _, resource := range resources {
		result := i.AutoImport(ctx, resource.ResourceType, resource.ResourceID, resource.Attributes)
		results = append(results, result)
	}

	return results
}
