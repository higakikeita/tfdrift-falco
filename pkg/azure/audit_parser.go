package azure

import (
	"strings"
	"time"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
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

// Parse converts a Falco output response into a TFDrift event for drift detection.
//
// The method performs the following operations:
//   - Validates the response is from the azureaudit source
//   - Extracts Azure operation name (e.g., "Microsoft.Compute/virtualMachines/write")
//   - Filters irrelevant events (read-only operations, non-infrastructure changes)
//   - Maps the operation to a Terraform resource type (e.g., "azurerm_linux_virtual_machine")
//   - Extracts resource identifiers, subscription, and resource group
//   - Parses event-specific changes from the audit log
//
// Parameters:
//   - res: Falco output response containing Azure audit log data
//
// Returns:
//   - *types.Event: Parsed drift detection event, or nil if:
//   - Response is nil
//   - Source is not "azureaudit"
//   - Event is not relevant for drift detection
//   - Required fields are missing
//   - No resource mapping exists for the event type
func (p *AuditParser) Parse(res *outputs.Response) *types.Event {
	// Handle nil response
	if res == nil {
		log.Warn("Received nil response")
		return nil
	}

	// Check if this is an Azure Audit Log event
	if res.Source != "azureaudit" {
		return nil
	}

	// Parse output fields
	fields := res.OutputFields

	// Extract Azure operation name (equivalent to CloudTrail event name)
	operationName, ok := fields["azure.operationName"]
	if !ok || operationName == "" {
		log.Warnf("Missing azure.operationName in Falco output")
		return nil
	}

	// Check if this is a relevant event for drift detection
	if !p.isRelevantOperation(operationName) {
		log.Debugf("Operation %s is not relevant for drift detection", operationName)
		return nil
	}

	// Check if the operation succeeded
	status := getAzureStringField(fields, "azure.status")
	if status != "" && status != "Succeeded" && status != "Started" {
		log.Debugf("Operation %s failed or not applicable (status: %s)", operationName, status)
		return nil
	}

	// Map to Terraform resource type
	resourceType := p.mapper.MapOperationToResource(operationName)
	if resourceType == "" {
		log.Debugf("No mapping for Azure operation: %s", operationName)
		return nil
	}

	// Extract resource information
	resourceID := getAzureStringField(fields, "azure.resourceId")
	caller := getAzureStringField(fields, "azure.caller")
	subscriptionID := getAzureStringField(fields, "azure.subscriptionId")
	resourceGroup := getAzureStringField(fields, "azure.resourceGroup")

	// Extract resource name from resource ID
	resourceName := extractResourceName(resourceID)

	// Extract user identity
	userIdentity := types.UserIdentity{
		Type:     "User",
		UserName: caller,
	}

	// Extract changes based on operation type
	changes := p.extractChanges(operationName, fields)

	return &types.Event{
		Provider:       "azure",
		EventName:      operationName,
		ResourceType:   resourceType,
		ResourceID:     resourceName,
		SubscriptionID: subscriptionID,
		ResourceGroup:  resourceGroup,
		UserIdentity:   userIdentity,
		Changes:        changes,
		RawEvent:       res,
	}
}

// extractChanges extracts attribute changes from Azure Audit Log
func (p *AuditParser) extractChanges(operationName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	// Determine change type based on operation
	if strings.HasSuffix(operationName, "/delete") {
		changes["_action"] = "delete"
	} else if strings.HasSuffix(operationName, "/write") {
		changes["_action"] = "create"
	}

	// Extract relevant request/response data if available
	if properties, ok := fields["azure.properties"]; ok && properties != "" {
		changes["properties"] = properties
	}

	return changes
}

// getAzureStringField safely gets a string field from Falco output fields
func getAzureStringField(fields map[string]string, key string) string {
	if val, ok := fields[key]; ok {
		return val
	}
	return ""
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
