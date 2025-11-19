package falco

import (
	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// parseFalcoOutput parses a Falco output response into a TFDrift event
func (s *Subscriber) parseFalcoOutput(res *outputs.Response) *types.Event {
	// Check if this is a CloudTrail event
	if res.Source != "aws_cloudtrail" {
		return nil
	}

	// Parse output fields
	fields := res.OutputFields

	// Extract CloudTrail event name
	eventName, ok := fields["ct.name"]
	if !ok || eventName == "" {
		log.Warnf("Missing ct.name in Falco output")
		return nil
	}

	// Check if this is a relevant event for drift detection
	if !s.isRelevantEvent(eventName) {
		return nil
	}

	// Extract resource ID
	resourceID := s.extractResourceID(eventName, fields)
	if resourceID == "" {
		log.Debugf("Could not extract resource ID from event %s", eventName)
		return nil
	}

	// Extract resource type
	resourceType := s.mapEventToResourceType(eventName)

	// Extract user identity
	userIdentity := types.UserIdentity{
		Type:        getStringField(fields, "ct.user.type"),
		PrincipalID: getStringField(fields, "ct.user.principalid"),
		ARN:         getStringField(fields, "ct.user.arn"),
		AccountID:   getStringField(fields, "ct.user.accountid"),
		UserName:    getStringField(fields, "ct.user"),
	}

	// Extract changes based on event type
	changes := s.extractChanges(eventName, fields)

	return &types.Event{
		Provider:     "aws",
		EventName:    eventName,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		UserIdentity: userIdentity,
		Changes:      changes,
		RawEvent:     res,
	}
}

// isRelevantEvent checks if an event is relevant for drift detection
func (s *Subscriber) isRelevantEvent(eventName string) bool {
	relevantEvents := map[string]bool{
		// EC2
		"ModifyInstanceAttribute":         true,
		"ModifyNetworkInterfaceAttribute": true,
		"ModifyVolume":                    true,

		// IAM - Policy modifications
		"PutUserPolicy":          true,
		"PutRolePolicy":          true,
		"PutGroupPolicy":         true,
		"UpdateAssumeRolePolicy": true,
		"AttachUserPolicy":       true,
		"AttachRolePolicy":       true,
		"AttachGroupPolicy":      true,
		"CreatePolicy":           true,
		"CreatePolicyVersion":    true,

		// IAM - User/Role/Group lifecycle
		"CreateRole":                  true,
		"DeleteRole":                  true,
		"CreateUser":                  true,
		"DeleteUser":                  true,
		"CreateAccessKey":             true,
		"AddUserToGroup":              true,
		"RemoveUserFromGroup":         true,
		"UpdateAccountPasswordPolicy": true,

		// S3
		"PutBucketPolicy":        true,
		"PutBucketVersioning":    true,
		"PutBucketEncryption":    true,
		"DeleteBucketEncryption": true,
		"PutBucketLogging":       true,

		// RDS
		"ModifyDBInstance": true,
		"ModifyDBCluster":  true,

		// Lambda
		"UpdateFunctionConfiguration": true,
		"UpdateFunctionCode":          true,
	}

	return relevantEvents[eventName]
}

// extractResourceID extracts the resource ID from Falco output fields
func (s *Subscriber) extractResourceID(eventName string, fields map[string]string) string {
	// Try different field names based on event type
	idFieldMap := map[string][]string{
		"ModifyInstanceAttribute":     {"ct.request.instanceid", "ct.resource.instanceid"},
		"ModifyVolume":                {"ct.request.volumeid"},
		"PutBucketPolicy":             {"ct.request.bucket", "ct.resource.bucket"},
		"PutBucketEncryption":         {"ct.request.bucket"},
		"DeleteBucketEncryption":      {"ct.request.bucket"},
		"ModifyDBInstance":            {"ct.request.dbinstanceidentifier"},
		"UpdateFunctionConfiguration": {"ct.request.functionname"},

		// IAM - Roles
		"PutRolePolicy":          {"ct.request.rolename"},
		"UpdateAssumeRolePolicy": {"ct.request.rolename"},
		"AttachRolePolicy":       {"ct.request.rolename"},
		"CreateRole":             {"ct.request.rolename"},
		"DeleteRole":             {"ct.request.rolename"},

		// IAM - Users
		"PutUserPolicy":       {"ct.request.username"},
		"AttachUserPolicy":    {"ct.request.username"},
		"CreateUser":          {"ct.request.username"},
		"DeleteUser":          {"ct.request.username"},
		"CreateAccessKey":     {"ct.request.username"},
		"AddUserToGroup":      {"ct.request.username"},
		"RemoveUserFromGroup": {"ct.request.username"},

		// IAM - Groups
		"PutGroupPolicy":    {"ct.request.groupname"},
		"AttachGroupPolicy": {"ct.request.groupname"},

		// IAM - Policies
		"CreatePolicy":        {"ct.request.policyname"},
		"CreatePolicyVersion": {"ct.request.policyarn"},
	}

	// Get possible field names for this event
	possibleFields := idFieldMap[eventName]
	if possibleFields == nil {
		// Default fields to try
		possibleFields = []string{"ct.resource.id", "ct.request.resource"}
	}

	// Try each field
	for _, fieldName := range possibleFields {
		if id := getStringField(fields, fieldName); id != "" {
			return id
		}
	}

	return ""
}
