package detector

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestHandleAutoImport(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance"},
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged-123",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_NotAllowed(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_s3_bucket"}, // Not aws_instance
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance", // Not in allowed list
		ResourceID:   "i-unmanaged-456",
		EventName:    "RunInstances",
		Changes:      map[string]interface{}{},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic even if not allowed
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_EmptyAllowList(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{}, // Empty = allow all
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_iam_role",
		ResourceID:   "role-123",
		EventName:    "CreateRole",
		Changes:      map[string]interface{}{},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic with empty allow list
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_WithApprovalRequired_NonInteractive(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:         true,
			TerraformDir:    ".",
			RequireApproval: true, // Requires approval
			OutputDir:       "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false) // NON-interactive mode

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-noninteractive-123",
		EventName:    "RunInstances",
		Changes:      map[string]interface{}{"instance_type": "t3.micro"},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should handle non-interactive mode gracefully (prompt will fail, should log error)
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_WithGeneratedCode(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_s3_bucket"},
			OutputDir:        "/tmp/terraform-imports",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-new-bucket",
		EventName:    "CreateBucket",
		Changes: map[string]interface{}{
			"bucket":        "my-new-bucket",
			"versioning":    true,
			"acl":           "private",
			"force_destroy": false,
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "devops",
			ARN:      "arn:aws:iam::123456789012:user/devops",
		},
	}

	// Should generate code and suggest output file
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleEvent_WithAutoImport(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance"},
		},
	}

	stateManager := &terraform.StateManager{}
	formatter := diff.NewFormatter(false)
	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		stateManager:    stateManager,
		formatter:       formatter,
		importer:        importer,
		approvalManager: approvalManager,
	}

	// Unmanaged resource should trigger auto-import
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-new-resource",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleAutoImport_DryRunMode(t *testing.T) {
	cfg := &config.Config{
		DryRun: true, // Dry-run mode
		AutoImport: config.AutoImportConfig{
			Enabled:      true,
			TerraformDir: ".",
			OutputDir:    "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-dryrun-123",
		EventName:    "RunInstances",
		Changes:      map[string]interface{}{"instance_type": "t3.micro"},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should handle dry-run mode gracefully
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}
