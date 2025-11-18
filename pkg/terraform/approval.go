package terraform

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// ApprovalRequest represents a request for import approval
type ApprovalRequest struct {
	ID            string
	ResourceType  string
	ResourceID    string
	ResourceName  string
	DetectedAt    time.Time
	UserIdentity  string
	Changes       map[string]interface{}
	ImportCommand *ImportCommand
	Status        ApprovalStatus
	ApprovedBy    string
	ApprovedAt    time.Time
}

// ApprovalStatus represents the status of an approval request
type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
	ApprovalExpired  ApprovalStatus = "expired"
)

// ApprovalManager manages import approval workflow
type ApprovalManager struct {
	pendingRequests map[string]*ApprovalRequest
	importer        *Importer
	interactiveMode bool
	mu              sync.RWMutex
	stdin           io.Reader // For testing: if nil, uses os.Stdin
}

// NewApprovalManager creates a new approval manager
func NewApprovalManager(importer *Importer, interactiveMode bool) *ApprovalManager {
	return &ApprovalManager{
		pendingRequests: make(map[string]*ApprovalRequest),
		importer:        importer,
		interactiveMode: interactiveMode,
	}
}

// RequestApproval creates a new approval request
func (am *ApprovalManager) RequestApproval(resourceType, resourceID string, changes map[string]interface{}, userIdentity string) *ApprovalRequest {
	cmd := am.importer.GenerateImportCommand(resourceType, resourceID)

	request := &ApprovalRequest{
		ID:            fmt.Sprintf("import-%s-%d", resourceID, time.Now().Unix()),
		ResourceType:  resourceType,
		ResourceID:    resourceID,
		ResourceName:  cmd.ResourceName,
		DetectedAt:    time.Now(),
		UserIdentity:  userIdentity,
		Changes:       changes,
		ImportCommand: cmd,
		Status:        ApprovalPending,
	}

	am.pendingRequests[request.ID] = request
	return request
}

// PromptForApproval prompts the user for approval in interactive mode
func (am *ApprovalManager) PromptForApproval(ctx context.Context, request *ApprovalRequest) (bool, error) {
	if !am.interactiveMode {
		return false, fmt.Errorf("not in interactive mode")
	}

	// Display the import request
	fmt.Println("\n" + strings.Repeat("━", 60))
	fmt.Println("🔔 IMPORT APPROVAL REQUIRED")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Printf("\n📦 Resource Type: %s\n", request.ResourceType)
	fmt.Printf("🆔 Resource ID:   %s\n", request.ResourceID)
	fmt.Printf("📝 Resource Name: %s (auto-generated)\n", request.ResourceName)
	fmt.Printf("👤 Detected By:   %s\n", request.UserIdentity)
	fmt.Printf("🕐 Detected At:   %s\n", request.DetectedAt.Format(time.RFC3339))

	if len(request.Changes) > 0 {
		fmt.Println("\n🔄 Changes:")
		for key, value := range request.Changes {
			fmt.Printf("   %s: %v\n", key, value)
		}
	}

	fmt.Printf("\n💻 Import Command:\n")
	fmt.Printf("   %s\n", request.ImportCommand.String())

	fmt.Printf("\n❓ Approve this import? [y/N]: ")

	// Read user input - use injected stdin for testing, fallback to os.Stdin
	stdinReader := am.stdin
	if stdinReader == nil {
		stdinReader = os.Stdin
	}
	reader := bufio.NewReader(stdinReader)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(strings.ToLower(input))
	approved := input == "y" || input == "yes"

	if approved {
		request.Status = ApprovalApproved
		request.ApprovedBy = "console-user" // TODO: Get actual user
		request.ApprovedAt = time.Now()
		fmt.Println("✅ Import approved!")
	} else {
		request.Status = ApprovalRejected
		fmt.Println("❌ Import rejected")
	}

	return approved, nil
}

// ApproveAndExecute approves and executes an import
func (am *ApprovalManager) ApproveAndExecute(ctx context.Context, requestID string, approvedBy string) (*ImportResult, error) {
	request, exists := am.pendingRequests[requestID]
	if !exists {
		return nil, fmt.Errorf("approval request not found: %s", requestID)
	}

	if request.Status != ApprovalPending {
		return nil, fmt.Errorf("request is not pending (status: %s)", request.Status)
	}

	// Mark as approved
	request.Status = ApprovalApproved
	request.ApprovedBy = approvedBy
	request.ApprovedAt = time.Now()

	log.Infof("Import approved by %s: %s", approvedBy, request.ImportCommand.String())

	// Execute import
	result := am.importer.AutoImport(ctx, request.ResourceType, request.ResourceID, request.Changes)

	// Clean up
	delete(am.pendingRequests, requestID)

	return result, nil
}

// Reject rejects an import request
func (am *ApprovalManager) Reject(requestID string, reason string) error {
	request, exists := am.pendingRequests[requestID]
	if !exists {
		return fmt.Errorf("approval request not found: %s", requestID)
	}

	request.Status = ApprovalRejected
	log.Infof("Import rejected: %s (reason: %s)", request.ImportCommand.String(), reason)

	delete(am.pendingRequests, requestID)
	return nil
}

// ListPending returns all pending approval requests
func (am *ApprovalManager) ListPending() []*ApprovalRequest {
	pending := make([]*ApprovalRequest, 0, len(am.pendingRequests))
	for _, req := range am.pendingRequests {
		if req.Status == ApprovalPending {
			pending = append(pending, req)
		}
	}
	return pending
}

// CleanupExpired removes expired approval requests
func (am *ApprovalManager) CleanupExpired(expiryDuration time.Duration) int {
	count := 0
	now := time.Now()

	for id, req := range am.pendingRequests {
		if req.Status == ApprovalPending && now.Sub(req.DetectedAt) > expiryDuration {
			req.Status = ApprovalExpired
			delete(am.pendingRequests, id)
			count++
			log.Infof("Expired import request: %s", req.ImportCommand.String())
		}
	}

	return count
}

// AutoApproveIfAllowed automatically approves if resource is in allowed list
func (am *ApprovalManager) AutoApproveIfAllowed(ctx context.Context, request *ApprovalRequest, allowedResources []string) (*ImportResult, error) {
	// Check if resource type is in allowed list
	allowed := false
	if len(allowedResources) == 0 {
		// Empty list means all resources are allowed
		allowed = true
	} else {
		for _, rt := range allowedResources {
			if rt == request.ResourceType {
				allowed = true
				break
			}
		}
	}

	if !allowed {
		return nil, fmt.Errorf("resource type %s is not in auto-approve list", request.ResourceType)
	}

	log.Infof("Auto-approving import for %s (allowed resource type)", request.ResourceType)
	return am.ApproveAndExecute(ctx, request.ID, "auto-approval")
}

// FormatApprovalSummary formats a summary of the approval request
func (request *ApprovalRequest) FormatApprovalSummary() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Approval Request: %s\n", request.ID))
	b.WriteString(fmt.Sprintf("Status:          %s\n", request.Status))
	b.WriteString(fmt.Sprintf("Resource:        %s (%s)\n", request.ResourceID, request.ResourceType))
	b.WriteString(fmt.Sprintf("Import Command:  %s\n", request.ImportCommand.String()))
	b.WriteString(fmt.Sprintf("Detected At:     %s\n", request.DetectedAt.Format(time.RFC3339)))

	if request.Status == ApprovalApproved {
		b.WriteString(fmt.Sprintf("Approved By:     %s\n", request.ApprovedBy))
		b.WriteString(fmt.Sprintf("Approved At:     %s\n", request.ApprovedAt.Format(time.RFC3339)))
	}

	return b.String()
}
