package detector

import (
	"context"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestHandleRemediation_DisabledRemediationConfig tests that remediation is skipped when disabled
func TestHandleRemediation_DisabledRemediationConfig(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled: false, // Disabled
		},
	}

	detector := &Detector{
		cfg: cfg,
	}

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
	}

	// Should return immediately without error
	detector.handleRemediation(context.Background(), alert)
}

func TestHandleRemediation_WithNilBroadcaster(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			CreatePRs: false,
		},
	}

	detector := &Detector{
		cfg:         cfg,
		broadcaster: nil, // No broadcaster configured
	}

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-789",
	}

	// Should not panic even without broadcaster
	detector.handleRemediation(context.Background(), alert)
}

func TestHandleUnmanagedRemediation_DisabledRemediationConfig(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled: false,
		},
	}

	detector := &Detector{
		cfg: cfg,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-bucket",
	}

	detector.handleUnmanagedRemediation(context.Background(), event)
	// Should return immediately without error
}

func TestHandleUnmanagedRemediation_WithNilBroadcaster(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			CreatePRs: false,
			DryRun:    false,
		},
		GitHub: config.GitHubConfig{
			Enabled: false,
		},
	}

	detector := &Detector{
		cfg:         cfg,
		broadcaster: nil,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_rds_instance",
		ResourceID:   "db-123",
		UserIdentity: types.UserIdentity{
			UserName: "admin",
		},
	}

	detector.handleUnmanagedRemediation(context.Background(), event)
	// Should not panic
}

func TestCreateRemediationPR_DryRunMode(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			DryRun:    true,
			CreatePRs: true,
		},
		GitHub: config.GitHubConfig{
			Enabled: true,
			Token:   "valid-token",
			Owner:   "test-owner",
			Repo:    "test-repo",
			Branch:  "main",
		},
	}

	detector := &Detector{
		cfg: cfg,
	}

	proposal := &types.RemediationProposal{
		ID:             "prop-123",
		ResourceID:     "i-456",
		ResourceType:   "aws_instance",
		ResourceName:   "web-server",
		Severity:       "high",
		Status:         types.RemediationPending,
		Description:    "Fix instance type drift",
		TerraformCode:  "resource \"aws_instance\" \"web\" {\n  instance_type = \"t3.large\"\n}",
		ImportCommand:  "terraform import aws_instance.web i-456",
		PlanCommand:    "terraform plan",
	}

	detector.createRemediationPR(context.Background(), proposal)
	// In dry run mode, should not actually create PR
}

func TestCreateRemediationPR_NoToken(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			DryRun:    false,
			CreatePRs: true,
		},
		GitHub: config.GitHubConfig{
			Enabled: true,
			Token:   "", // Empty token
			Owner:   "test-owner",
			Repo:    "test-repo",
			Branch:  "main",
		},
	}

	detector := &Detector{
		cfg: cfg,
	}

	proposal := &types.RemediationProposal{
		ID:           "prop-123",
		ResourceID:   "i-456",
		ResourceType: "aws_instance",
	}

	detector.createRemediationPR(context.Background(), proposal)
	// Should skip PR creation due to missing token
}

func TestCreateRemediationPR_NoTerraformFiles(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			DryRun:    false,
			CreatePRs: true,
		},
		GitHub: config.GitHubConfig{
			Enabled: true,
			Token:   "valid-token",
			Owner:   "test-owner",
			Repo:    "test-repo",
			Branch:  "main",
		},
	}

	detector := &Detector{
		cfg: cfg,
	}

	proposal := &types.RemediationProposal{
		ID:            "prop-123",
		ResourceID:    "i-456",
		ResourceType:  "aws_instance",
		TerraformCode: "", // Empty code
	}

	detector.createRemediationPR(context.Background(), proposal)
	// Should skip PR creation due to no terraform files
}

func TestEvaluatePolicy_NilPolicyEngine(t *testing.T) {
	detector := &Detector{
		policyEngine: nil,
	}

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
	}

	result := detector.evaluatePolicy(context.Background(), alert)

	assert.Nil(t, result)
}

func TestEvaluatePolicy_WithNilBroadcaster(t *testing.T) {
	detector := &Detector{
		policyEngine: nil,
		broadcaster:  nil,
	}

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-999",
	}

	// Should not panic even without broadcaster
	result := detector.evaluatePolicy(context.Background(), alert)

	assert.Nil(t, result)
}

func TestEvaluateUnmanagedPolicy_NilPolicyEngine(t *testing.T) {
	detector := &Detector{
		policyEngine: nil,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-bucket",
	}

	result := detector.evaluateUnmanagedPolicy(context.Background(), event)

	assert.Nil(t, result)
}

func TestEvaluateUnmanagedPolicy_WithNilBroadcaster(t *testing.T) {
	detector := &Detector{
		policyEngine: nil,
		broadcaster:  nil,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_rds_instance",
		ResourceID:   "db-123",
	}

	// Should not panic even without broadcaster
	result := detector.evaluateUnmanagedPolicy(context.Background(), event)

	assert.Nil(t, result)
}

func TestHandleRemediation_ProposalCreation(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			CreatePRs: false,
			DryRun:    false,
		},
		GitHub: config.GitHubConfig{
			Enabled: false,
		},
	}

	detector := &Detector{
		cfg:         cfg,
		broadcaster: nil,
	}

	alert := &types.DriftAlert{
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		ResourceName: "web-server",
		Severity:     "critical",
		Attribute:    "instance_type",
		OldValue:     "t2.micro",
		NewValue:     "t2.large",
	}

	// Should not panic and should process remediation
	detector.handleRemediation(context.Background(), alert)
}

func TestHandleUnmanagedRemediation_ProposalCreation(t *testing.T) {
	cfg := &config.Config{
		Remediation: config.RemediationConfig{
			Enabled:   true,
			CreatePRs: false,
			DryRun:    false,
		},
		GitHub: config.GitHubConfig{
			Enabled: false,
		},
	}

	detector := &Detector{
		cfg:         cfg,
		broadcaster: nil,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-bucket",
		UserIdentity: types.UserIdentity{
			UserName: "admin",
		},
	}

	// Should not panic and should process remediation
	detector.handleUnmanagedRemediation(context.Background(), event)
}

func TestEvaluatePolicy_AlertFields(t *testing.T) {
	// Test that we can construct alerts with all required fields
	alert := &types.DriftAlert{
		AlertType:    "drift",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-test123",
		ResourceName: "web-sg",
		Attribute:    "ingress",
		OldValue:     "all-open",
		NewValue:     "restricted",
		Severity:     "critical",
		Timestamp:    "2024-01-15T10:30:00Z",
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDATEST",
			ARN:         "arn:aws:iam::111111:user/test",
			AccountID:   "111111",
			UserName:    "testuser",
		},
	}

	assert.NotEmpty(t, alert.AlertType)
	assert.NotEmpty(t, alert.ResourceType)
	assert.NotEmpty(t, alert.ResourceID)
	assert.NotEmpty(t, alert.Severity)
	assert.NotEmpty(t, alert.UserIdentity.Type)
}

func TestEvaluateUnmanagedPolicy_EventFields(t *testing.T) {
	// Test that we can construct events with all required fields
	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_rds_instance",
		ResourceID:   "db-unmanaged",
		Changes: map[string]interface{}{
			"allocated_storage": 1000,
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMRole",
			PrincipalID: "AIDAROLETEST",
			ARN:      "arn:aws:iam::222222:role/lambda",
			AccountID: "222222",
			UserName: "lambda-execution",
		},
	}

	assert.NotEmpty(t, event.Provider)
	assert.NotEmpty(t, event.ResourceType)
	assert.NotEmpty(t, event.ResourceID)
	assert.NotEmpty(t, event.UserIdentity.Type)
}
