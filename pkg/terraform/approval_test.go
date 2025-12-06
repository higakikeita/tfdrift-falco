package terraform

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockStdin creates a mock stdin reader with predefined input for testing
func mockStdin(input string) io.Reader {
	if input != "" && input[len(input)-1] != '\n' {
		input += "\n"
	}
	return bytes.NewBufferString(input)
}

func TestNewApprovalManager(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.importer)
	assert.NotNil(t, manager.pendingRequests)
	assert.False(t, manager.interactiveMode)
}

func TestRequestApproval(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	resourceType := "aws_instance"
	resourceID := "i-1234567890abcdef0"
	changes := map[string]interface{}{
		"instance_type": "t3.medium",
	}
	userIdentity := "admin@example.com"

	request := manager.RequestApproval(resourceType, resourceID, changes, userIdentity)

	assert.NotNil(t, request)
	assert.NotEmpty(t, request.ID)
	assert.Equal(t, resourceType, request.ResourceType)
	assert.Equal(t, resourceID, request.ResourceID)
	assert.NotEmpty(t, request.ResourceName)
	assert.Equal(t, userIdentity, request.UserIdentity)
	assert.Equal(t, changes, request.Changes)
	assert.Equal(t, ApprovalPending, request.Status)
	assert.NotNil(t, request.ImportCommand)

	// Verify it's stored in pending requests
	stored := manager.pendingRequests[request.ID]
	assert.Equal(t, request, stored)
}

func TestListPending(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	// Create multiple requests
	req1 := manager.RequestApproval("aws_instance", "i-111", nil, "user1")
	req2 := manager.RequestApproval("aws_s3_bucket", "bucket-222", nil, "user2")
	req3 := manager.RequestApproval("aws_iam_role", "role-333", nil, "user3")

	// Mark one as approved
	req2.Status = ApprovalApproved

	pending := manager.ListPending()

	assert.Len(t, pending, 2) // Only pending ones

	// Check that req2 is not in the list
	found := false
	for _, p := range pending {
		if p.ID == req2.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Approved request should not be in pending list")

	// Check that req1 and req3 are in the list
	ids := []string{pending[0].ID, pending[1].ID}
	assert.Contains(t, ids, req1.ID)
	assert.Contains(t, ids, req3.ID)
}

func TestReject(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-123", nil, "user")
	requestID := request.ID

	err := manager.Reject(requestID, "Not needed")
	assert.NoError(t, err)

	// Verify request is removed
	_, exists := manager.pendingRequests[requestID]
	assert.False(t, exists)

	// Verify status was updated before removal
	assert.Equal(t, ApprovalRejected, request.Status)
}

func TestReject_NotFound(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	err := manager.Reject("nonexistent-id", "reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "approval request not found")
}

func TestApproveAndExecute_DryRun(t *testing.T) {
	importer := NewImporter(".", true) // dry-run mode
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-123", map[string]interface{}{
		"ami": "ami-123",
	}, "admin")

	ctx := context.Background()
	result, err := manager.ApproveAndExecute(ctx, request.ID, "approver@example.com")

	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verify request was updated
	assert.Equal(t, ApprovalApproved, request.Status)
	assert.Equal(t, "approver@example.com", request.ApprovedBy)
	assert.False(t, request.ApprovedAt.IsZero())

	// Verify request was removed from pending
	_, exists := manager.pendingRequests[request.ID]
	assert.False(t, exists)
}

func TestApproveAndExecute_NotFound(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	ctx := context.Background()
	result, err := manager.ApproveAndExecute(ctx, "nonexistent-id", "approver")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "approval request not found")
}

func TestApproveAndExecute_NotPending(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-123", nil, "user")
	request.Status = ApprovalRejected // Already rejected

	ctx := context.Background()
	result, err := manager.ApproveAndExecute(ctx, request.ID, "approver")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "request is not pending")
}

func TestCleanupExpired(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	// Create requests with different detection times
	req1 := manager.RequestApproval("aws_instance", "i-111", nil, "user")
	req2 := manager.RequestApproval("aws_s3_bucket", "bucket-222", nil, "user")
	req3 := manager.RequestApproval("aws_iam_role", "role-333", nil, "user")

	// Manually set detection times
	req1.DetectedAt = time.Now().Add(-2 * time.Hour)    // Old
	req2.DetectedAt = time.Now().Add(-30 * time.Minute) // Recent
	req3.DetectedAt = time.Now().Add(-3 * time.Hour)    // Very old

	// Cleanup requests older than 1 hour
	count := manager.CleanupExpired(1 * time.Hour)

	assert.Equal(t, 2, count) // req1 and req3 should be cleaned up

	// Verify only req2 remains
	assert.Len(t, manager.pendingRequests, 1)
	_, exists := manager.pendingRequests[req2.ID]
	assert.True(t, exists)

	// Verify expired requests were marked as expired
	assert.Equal(t, ApprovalExpired, req1.Status)
	assert.Equal(t, ApprovalExpired, req3.Status)
	assert.Equal(t, ApprovalPending, req2.Status) // Still pending
}

func TestCleanupExpired_NoExpiredRequests(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	// Create recent request
	manager.RequestApproval("aws_instance", "i-111", nil, "user")

	// Cleanup with long expiry duration
	count := manager.CleanupExpired(24 * time.Hour)

	assert.Equal(t, 0, count)
	assert.Len(t, manager.pendingRequests, 1)
}

func TestAutoApproveIfAllowed_EmptyAllowList(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-123", map[string]interface{}{
		"ami": "ami-123",
	}, "user")

	ctx := context.Background()
	// Empty allow list means all resources are allowed
	result, err := manager.AutoApproveIfAllowed(ctx, request, []string{})

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ApprovalApproved, request.Status)
	assert.Equal(t, "auto-approval", request.ApprovedBy)
}

func TestAutoApproveIfAllowed_ResourceInList(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-123", map[string]interface{}{
		"ami": "ami-123",
	}, "user")

	ctx := context.Background()
	allowedResources := []string{"aws_instance", "aws_s3_bucket"}
	result, err := manager.AutoApproveIfAllowed(ctx, request, allowedResources)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, ApprovalApproved, request.Status)
}

func TestAutoApproveIfAllowed_ResourceNotInList(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_iam_role", "role-123", nil, "user")

	ctx := context.Background()
	allowedResources := []string{"aws_instance", "aws_s3_bucket"} // role not included
	result, err := manager.AutoApproveIfAllowed(ctx, request, allowedResources)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not in auto-approve list")
	assert.Equal(t, ApprovalPending, request.Status) // Status unchanged
}

func TestFormatApprovalSummary(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	tests := []struct {
		name         string
		setupRequest func() *ApprovalRequest
		wantContains []string
	}{
		{
			name: "Pending Request",
			setupRequest: func() *ApprovalRequest {
				return manager.RequestApproval("aws_instance", "i-123", map[string]interface{}{
					"ami": "ami-123",
				}, "admin@example.com")
			},
			wantContains: []string{
				"Approval Request:",
				"Status:          pending",
				"Resource:        i-123",
				"aws_instance",
				"terraform import",
				"Detected At:",
			},
		},
		{
			name: "Approved Request",
			setupRequest: func() *ApprovalRequest {
				req := manager.RequestApproval("aws_s3_bucket", "my-bucket", nil, "user")
				req.Status = ApprovalApproved
				req.ApprovedBy = "approver@example.com"
				req.ApprovedAt = time.Now()
				return req
			},
			wantContains: []string{
				"Approval Request:",
				"Status:          approved",
				"Resource:        my-bucket",
				"aws_s3_bucket",
				"Approved By:     approver@example.com",
				"Approved At:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := tt.setupRequest()
			summary := request.FormatApprovalSummary()

			for _, want := range tt.wantContains {
				assert.Contains(t, summary, want, "Summary should contain: %s", want)
			}
		})
	}
}

func TestApprovalStatus_Constants(t *testing.T) {
	assert.Equal(t, ApprovalStatus("pending"), ApprovalPending)
	assert.Equal(t, ApprovalStatus("approved"), ApprovalApproved)
	assert.Equal(t, ApprovalStatus("rejected"), ApprovalRejected)
	assert.Equal(t, ApprovalStatus("expired"), ApprovalExpired)
}

func TestApprovalRequest_Structure(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	changes := map[string]interface{}{
		"instance_type": "t3.micro",
		"ami":           "ami-123",
	}

	request := manager.RequestApproval("aws_instance", "i-test", changes, "test@example.com")

	// Verify all fields are populated
	assert.NotEmpty(t, request.ID)
	assert.Equal(t, "aws_instance", request.ResourceType)
	assert.Equal(t, "i-test", request.ResourceID)
	assert.NotEmpty(t, request.ResourceName)
	assert.False(t, request.DetectedAt.IsZero())
	assert.Equal(t, "test@example.com", request.UserIdentity)
	assert.Equal(t, changes, request.Changes)
	assert.NotNil(t, request.ImportCommand)
	assert.Equal(t, ApprovalPending, request.Status)
	assert.Empty(t, request.ApprovedBy)
	assert.True(t, request.ApprovedAt.IsZero())
}

func TestPromptForApproval_NotInteractiveMode(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false) // non-interactive

	request := manager.RequestApproval("aws_instance", "i-123", nil, "user")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.Error(t, err)
	assert.False(t, approved)
	assert.Contains(t, err.Error(), "not in interactive mode")
}

func TestPromptForApproval_UserApprovesWithY(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true) // interactive mode
	manager.stdin = mockStdin("y")

	request := manager.RequestApproval("aws_instance", "i-123", nil, "user@example.com")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.True(t, approved)
	assert.Equal(t, ApprovalApproved, request.Status)
	assert.Equal(t, "console-user", request.ApprovedBy)
	assert.False(t, request.ApprovedAt.IsZero())
}

func TestPromptForApproval_UserApprovesWithYes(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("yes")

	request := manager.RequestApproval("aws_s3_bucket", "my-bucket", map[string]interface{}{
		"versioning": true,
	}, "admin@example.com")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.True(t, approved)
	assert.Equal(t, ApprovalApproved, request.Status)
}

func TestPromptForApproval_UserRejectsWithN(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("n")

	request := manager.RequestApproval("aws_instance", "i-456", nil, "user")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.False(t, approved)
	assert.Equal(t, ApprovalRejected, request.Status)
	assert.Empty(t, request.ApprovedBy)
	assert.True(t, request.ApprovedAt.IsZero())
}

func TestPromptForApproval_UserRejectsWithNo(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("no")

	request := manager.RequestApproval("aws_iam_role", "role-789", nil, "user")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.False(t, approved)
	assert.Equal(t, ApprovalRejected, request.Status)
}

func TestPromptForApproval_EmptyInput(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("") // Just press Enter

	request := manager.RequestApproval("aws_instance", "i-empty", nil, "user")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	// Empty input (EOF) causes read error, which is expected
	assert.Error(t, err)
	assert.False(t, approved)
	assert.Contains(t, err.Error(), "failed to read input")
}

func TestPromptForApproval_InvalidInput(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("maybe")

	request := manager.RequestApproval("aws_instance", "i-invalid", nil, "user")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.False(t, approved) // Invalid input treated as reject
	assert.Equal(t, ApprovalRejected, request.Status)
}

func TestPromptForApproval_CaseInsensitive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Uppercase Y", "Y", true},
		{"Uppercase YES", "YES", true},
		{"Mixed case Yes", "Yes", true},
		{"Uppercase N", "N", false},
		{"Uppercase NO", "NO", false},
		{"Mixed case No", "No", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			importer := NewImporter(".", true)
			manager := NewApprovalManager(importer, true)
			manager.stdin = mockStdin(tt.input)

			request := manager.RequestApproval("aws_instance", "i-case", nil, "user")

			ctx := context.Background()
			approved, err := manager.PromptForApproval(ctx, request)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, approved)
			if tt.expected {
				assert.Equal(t, ApprovalApproved, request.Status)
			} else {
				assert.Equal(t, ApprovalRejected, request.Status)
			}
		})
	}
}

func TestPromptForApproval_WithChanges(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)
	manager.stdin = mockStdin("y")

	changes := map[string]interface{}{
		"instance_type": "t3.large",
		"monitoring":    true,
		"tags": map[string]string{
			"Environment": "production",
			"Team":        "DevOps",
		},
	}

	request := manager.RequestApproval("aws_instance", "i-prod-001", changes, "admin@company.com")

	ctx := context.Background()
	approved, err := manager.PromptForApproval(ctx, request)

	assert.NoError(t, err)
	assert.True(t, approved)
	assert.Equal(t, ApprovalApproved, request.Status)
	assert.Equal(t, changes, request.Changes) // Changes preserved
}

func TestPromptForApproval_MultipleSequentialPrompts(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, true)

	// Create multiple requests
	req1 := manager.RequestApproval("aws_instance", "i-111", nil, "user1")
	req2 := manager.RequestApproval("aws_s3_bucket", "bucket-222", nil, "user2")
	req3 := manager.RequestApproval("aws_iam_role", "role-333", nil, "user3")

	ctx := context.Background()

	// Approve first
	manager.stdin = mockStdin("yes")
	approved1, err1 := manager.PromptForApproval(ctx, req1)
	assert.NoError(t, err1)
	assert.True(t, approved1)
	assert.Equal(t, ApprovalApproved, req1.Status)

	// Reject second
	manager.stdin = mockStdin("no")
	approved2, err2 := manager.PromptForApproval(ctx, req2)
	assert.NoError(t, err2)
	assert.False(t, approved2)
	assert.Equal(t, ApprovalRejected, req2.Status)

	// Approve third
	manager.stdin = mockStdin("y")
	approved3, err3 := manager.PromptForApproval(ctx, req3)
	assert.NoError(t, err3)
	assert.True(t, approved3)
	assert.Equal(t, ApprovalApproved, req3.Status)
}

func TestFormatApprovalSummary_WithChanges(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	changes := map[string]interface{}{
		"instance_type": "t3.large",
		"monitoring":    true,
		"tags":          map[string]string{"env": "prod"},
	}

	request := manager.RequestApproval("aws_instance", "i-prod-001", changes, "admin")

	summary := request.FormatApprovalSummary()

	assert.Contains(t, summary, "aws_instance")
	assert.Contains(t, summary, "i-prod-001")
	assert.Contains(t, summary, "pending")
}

func TestMultipleManagerInstances(t *testing.T) {
	importer1 := NewImporter(".", true)
	importer2 := NewImporter(".", true)

	manager1 := NewApprovalManager(importer1, false)
	manager2 := NewApprovalManager(importer2, true)

	// Create requests in different managers
	req1 := manager1.RequestApproval("aws_instance", "i-111", nil, "user1")
	req2 := manager2.RequestApproval("aws_instance", "i-222", nil, "user2")

	// Verify they are independent
	assert.Len(t, manager1.pendingRequests, 1)
	assert.Len(t, manager2.pendingRequests, 1)
	assert.NotEqual(t, req1.ID, req2.ID)

	// Verify interactive mode settings
	assert.False(t, manager1.interactiveMode)
	assert.True(t, manager2.interactiveMode)
}

func TestRequestApproval_IDFormat(t *testing.T) {
	importer := NewImporter(".", true)
	manager := NewApprovalManager(importer, false)

	request := manager.RequestApproval("aws_instance", "i-1234567890abcdef0", nil, "user")

	// ID should start with "import-" and contain the resource ID
	assert.True(t, strings.HasPrefix(request.ID, "import-i-1234567890abcdef0"))
	assert.Contains(t, request.ID, "i-1234567890abcdef0")
}
