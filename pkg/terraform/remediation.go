// Package terraform provides Terraform-related functionality for TFDrift-Falco.
package terraform

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// RemediationGenerator generates remediation proposals for detected drifts and unmanaged resources
type RemediationGenerator struct{}

// NewRemediationGenerator creates a new RemediationGenerator
func NewRemediationGenerator() *RemediationGenerator {
	return &RemediationGenerator{}
}

// GenerateForDrift creates a RemediationProposal for a detected drift
func (g *RemediationGenerator) GenerateForDrift(alert *types.DriftAlert) *types.RemediationProposal {
	if alert == nil {
		return nil
	}

	proposal := &types.RemediationProposal{
		ID:           uuid.New().String(),
		AlertType:    "drift",
		ResourceType: alert.ResourceType,
		ResourceID:   alert.ResourceID,
		ResourceName: alert.ResourceName,
		Severity:     alert.Severity,
		Status:       types.RemediationPending,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		Attributes: map[string]interface{}{
			"attribute":  alert.Attribute,
			"old_value":  alert.OldValue,
			"new_value":  alert.NewValue,
		},
	}

	proposal.Description = fmt.Sprintf(
		"Drift detected in %s.%s: attribute '%s' changed from %v to %v",
		alert.ResourceType, alert.ResourceName, alert.Attribute, alert.OldValue, alert.NewValue,
	)

	proposal.TerraformCode = g.generateDriftFixHCL(
		alert.ResourceType, alert.ResourceName, alert.Attribute, alert.OldValue, alert.NewValue,
	)
	proposal.ImportCommand = g.generateImportCommand(alert.ResourceType, alert.ResourceName, alert.ResourceID)
	proposal.PlanCommand = g.generatePlanCommand(alert.ResourceType, alert.ResourceName)

	return proposal
}

// GenerateForUnmanaged creates a RemediationProposal for an unmanaged resource
func (g *RemediationGenerator) GenerateForUnmanaged(event *types.Event) *types.RemediationProposal {
	if event == nil {
		return nil
	}

	proposal := &types.RemediationProposal{
		ID:           uuid.New().String(),
		AlertType:    "unmanaged",
		Provider:     event.Provider,
		ResourceType: event.ResourceType,
		ResourceID:   event.ResourceID,
		Severity:     "medium", // Default severity for unmanaged resources
		Status:       types.RemediationPending,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
		Attributes:   event.Changes,
	}

	proposal.Description = fmt.Sprintf(
		"Unmanaged resource detected: %s (ID: %s) in provider %s",
		event.ResourceType, event.ResourceID, event.Provider,
	)

	proposal.TerraformCode = g.generateHCL(event.ResourceType, "", event.Changes)
	proposal.ImportCommand = g.generateImportCommand(event.ResourceType, "", event.ResourceID)
	proposal.PlanCommand = g.generatePlanCommand(event.ResourceType, "")

	return proposal
}

// generateImportCommand generates `terraform import <type>.<name> <id>`
func (g *RemediationGenerator) generateImportCommand(resourceType, resourceName, resourceID string) string {
	if resourceName == "" {
		resourceName = "imported_resource"
	}
	return fmt.Sprintf("terraform import %s.%s %s", resourceType, resourceName, resourceID)
}

// generatePlanCommand generates `terraform plan -target=<type>.<name>`
func (g *RemediationGenerator) generatePlanCommand(resourceType, resourceName string) string {
	if resourceName == "" {
		resourceName = "imported_resource"
	}
	return fmt.Sprintf("terraform plan -target=%s.%s", resourceType, resourceName)
}

// generateHCL generates a skeleton HCL resource block
func (g *RemediationGenerator) generateHCL(resourceType, resourceName string, attributes map[string]interface{}) string {
	if resourceName == "" {
		resourceName = "imported_resource"
	}

	var hcl strings.Builder
	hcl.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resourceType, resourceName))

	// Add a comment about attributes
	hcl.WriteString("  # Add required attributes below\n")

	// Add attributes if provided
	if len(attributes) > 0 {
		hcl.WriteString("  # Attributes from cloud provider:\n")
		for key, value := range attributes {
			hcl.WriteString(fmt.Sprintf("  # %s = %q\n", key, fmt.Sprintf("%v", value)))
		}
	}

	hcl.WriteString("}\n")
	return hcl.String()
}

// generateDriftFixHCL generates HCL that fixes a specific attribute drift
func (g *RemediationGenerator) generateDriftFixHCL(resourceType, resourceName string, attribute string, oldValue, newValue interface{}) string {
	if resourceName == "" {
		resourceName = "resource"
	}

	var hcl strings.Builder
	hcl.WriteString("# Drift remediation: update the following attribute in your Terraform configuration\n\n")
	hcl.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resourceType, resourceName))
	hcl.WriteString(fmt.Sprintf("  %s = %v\n", attribute, formatValue(newValue)))
	hcl.WriteString("}\n")

	return hcl.String()
}

// formatValue formats a value for HCL output
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case bool:
		return fmt.Sprintf("%v", val)
	case float64:
		return fmt.Sprintf("%v", val)
	case int:
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%q", fmt.Sprintf("%v", val))
	}
}
