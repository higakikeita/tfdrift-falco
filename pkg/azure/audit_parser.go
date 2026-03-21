package azure

import (
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// ActivityLogEvent represents a parsed Azure Activity Log event
type ActivityLogEvent struct {
	OperationName  string
	ResourceID     string
	ResourceType   string
	ResourceGroup  string
	SubscriptionID string
	Status         string
	Caller         string
	Timestamp      time.Time
	Properties     map[string]interface{}
}

// AuditParser parses Azure Activity Log events and maps them to Terraform resources
type AuditParser struct {
	mapper      *ResourceMapper
	relevantOps map[string]bool
}

// NewAuditParser creates a new Azure audit parser
func NewAuditParser() *AuditParser {
	mapper := NewResourceMapper()

	// Build relevance set from mapper
	relevantOps := make(map[string]bool)
	for _, op := range mapper.GetAllSupportedOperations() {
		relevantOps[op] = true
	}

	return &AuditParser{
		mapper:      mapper,
		relevantOps: relevantOps,
	}
}

// ParseEvent parses an Azure Activity Log event and returns drift event data
func (p *AuditParser) ParseEvent(event map[string]interface{}) *DriftEvent {
	operationName := getStringField(event, "operationName")
	if operationName == "" {
		return nil
	}

	// Check if this is a relevant operation
	if !p.isRelevantOperation(operationName) {
		return nil
	}

	// Check if the operation succeeded
	status := getStringField(event, "status")
	if status != "" && status != "Succeeded" && status != "Started" {
		return nil
	}

	// Map to Terraform resource type
	terraformResourceType := p.mapper.MapOperationToResource(operationName)
	if terraformResourceType == "" {
		log.Debugf("No mapping for Azure operation: %s", operationName)
		return nil
	}

	// Extract resource information
	resourceID := getStringField(event, "resourceId")
	caller := getStringField(event, "caller")
	timestampStr := getStringField(event, "eventTimestamp")

	var timestamp time.Time
	if timestampStr != "" {
		if t, err := time.Parse(time.RFC3339, timestampStr); err == nil {
			timestamp = t
		}
	}
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Determine change type
	changeType := "modified"
	if strings.HasSuffix(operationName, "/delete") {
		changeType = "deleted"
	} else if strings.HasSuffix(operationName, "/write") {
		changeType = "created"
	}

	// Extract resource name from resource ID
	resourceName := extractResourceName(resourceID)
	resourceGroup := extractResourceGroup(resourceID)
	subscriptionID := extractSubscriptionID(resourceID)

	return &DriftEvent{
		Provider:              "azure",
		OperationName:         operationName,
		TerraformResourceType: terraformResourceType,
		ResourceID:            resourceID,
		ResourceName:          resourceName,
		ResourceGroup:         resourceGroup,
		SubscriptionID:        subscriptionID,
		ChangeType:            changeType,
		Caller:                caller,
		Timestamp:             timestamp,
	}
}

// isRelevantOperation checks if an Azure operation is relevant for drift detection
func (p *AuditParser) isRelevantOperation(operationName string) bool {
	return p.relevantOps[operationName]
}

// DriftEvent represents a detected drift event from Azure
type DriftEvent struct {
	Provider              string
	OperationName         string
	TerraformResourceType string
	ResourceID            string
	ResourceName          string
	ResourceGroup         string
	SubscriptionID        string
	ChangeType            string
	Caller                string
	Timestamp             time.Time
}

func getStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// extractResourceName extracts the resource name from an Azure resource ID
// e.g., "/subscriptions/.../resourceGroups/myRG/providers/Microsoft.Compute/virtualMachines/myVM" -> "myVM"
func extractResourceName(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return resourceID
}

// extractResourceGroup extracts the resource group from an Azure resource ID
func extractResourceGroup(resourceID string) string {
	parts := strings.Split(strings.ToLower(resourceID), "/")
	for i, part := range parts {
		if part == "resourcegroups" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// extractSubscriptionID extracts the subscription ID from an Azure resource ID
func extractSubscriptionID(resourceID string) string {
	parts := strings.Split(strings.ToLower(resourceID), "/")
	for i, part := range parts {
		if part == "subscriptions" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
