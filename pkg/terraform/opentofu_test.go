package terraform

import (
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// The IaC-tool selection must reach every user-visible command string:
// the importer's ImportCommand, the remediation generator's import/plan
// commands, and the Markdown proposal rendered into PRs.

func TestImportCommand_String_Binary(t *testing.T) {
	tf := &ImportCommand{ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-123", Binary: "terraform"}
	assert.Equal(t, "terraform import -- aws_instance.web i-123", tf.String())

	tofu := &ImportCommand{ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-123", Binary: "tofu"}
	assert.Equal(t, "tofu import -- aws_instance.web i-123", tofu.String())

	// Empty binary is backward-compatible with terraform
	legacy := &ImportCommand{ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-123"}
	assert.Equal(t, "terraform import -- aws_instance.web i-123", legacy.String())
}

func TestImporter_GenerateImportCommand_CarriesBinary(t *testing.T) {
	imp := NewImporterWithBinary(".", true, "tofu")
	cmd := imp.GenerateImportCommand("aws_s3_bucket", "my-bucket")
	assert.Equal(t, "tofu", cmd.Binary)
	assert.True(t, strings.HasPrefix(cmd.String(), "tofu import "))
}

func TestNewImporterWithBinary_EmptyFallsBackToTerraform(t *testing.T) {
	imp := NewImporterWithBinary(".", true, "")
	assert.Equal(t, "terraform", imp.terraformBinary)
}

func TestRemediationGenerator_Tofu_Commands(t *testing.T) {
	gen := NewRemediationGeneratorWithTool("tofu")
	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-abc",
		Severity:     "high",
		Attribute:    "instance_type",
	}
	proposal := gen.GenerateForDrift(alert)
	assert.Equal(t, "tofu import aws_instance.web i-abc", proposal.ImportCommand)
	assert.Equal(t, "tofu plan -target=aws_instance.web", proposal.PlanCommand)
}

func TestRemediationGenerator_DefaultIsTerraform(t *testing.T) {
	gen := NewRemediationGenerator()
	alert := &types.DriftAlert{ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-abc"}
	proposal := gen.GenerateForDrift(alert)
	assert.True(t, strings.HasPrefix(proposal.ImportCommand, "terraform import "))
	assert.True(t, strings.HasPrefix(proposal.PlanCommand, "terraform plan "))
}

func TestFormatProposalMarkdown_TofuMode(t *testing.T) {
	gen := NewRemediationGeneratorWithTool("tofu")
	proposal := gen.GenerateForDrift(&types.DriftAlert{
		ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-abc", Severity: "high",
	})
	md := FormatProposalMarkdown(proposal)
	assert.Contains(t, md, "tofu import aws_instance.web i-abc")
	assert.Contains(t, md, "tofu plan -target=aws_instance.web")
	assert.Contains(t, md, "`tofu apply`")
	assert.Contains(t, md, "Proposed OpenTofu Code")
	assert.NotContains(t, md, "terraform apply")
}

func TestFormatProposalMarkdown_TerraformModeUnchanged(t *testing.T) {
	gen := NewRemediationGenerator()
	proposal := gen.GenerateForDrift(&types.DriftAlert{
		ResourceType: "aws_instance", ResourceName: "web", ResourceID: "i-abc", Severity: "high",
	})
	md := FormatProposalMarkdown(proposal)
	assert.Contains(t, md, "`terraform apply`")
	assert.Contains(t, md, "Proposed Terraform Code")
}
