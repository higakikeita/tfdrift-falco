package falco

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/client"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Subscriber subscribes to Falco outputs via gRPC
type Subscriber struct {
	cfg    config.FalcoConfig
	client *client.Client
}

// NewSubscriber creates a new Falco subscriber
func NewSubscriber(cfg config.FalcoConfig) (*Subscriber, error) {
	return &Subscriber{
		cfg: cfg,
	}, nil
}

// Start starts subscribing to Falco outputs
func (s *Subscriber) Start(ctx context.Context, eventCh chan<- types.Event) error {
	log.Info("Starting Falco subscriber...")

	// Create Falco client configuration
	clientConfig := &client.Config{
		Hostname:   s.cfg.Hostname,
		Port:       s.cfg.Port,
		CertFile:   s.cfg.CertFile,
		KeyFile:    s.cfg.KeyFile,
		CARootFile: s.cfg.CARootFile,
	}

	// Create Falco gRPC client
	c, err := client.NewForConfig(ctx, clientConfig)
	if err != nil {
		return fmt.Errorf("failed to create Falco client: %w", err)
	}
	s.client = c
	defer c.Close()

	log.Infof("Connected to Falco at %s:%d", s.cfg.Hostname, s.cfg.Port)

	// Subscribe to outputs stream
	outputClient, err := c.Outputs()
	if err != nil {
		return fmt.Errorf("failed to get outputs client: %w", err)
	}

	// Start streaming using Sub method
	stream, err := outputClient.Sub(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe to Falco outputs: %w", err)
	}

	// Receive messages from stream
	for {
		select {
		case <-ctx.Done():
			log.Info("Falco subscriber stopping...")
			return ctx.Err()
		default:
			res, err := stream.Recv()
			if err != nil {
				return fmt.Errorf("error receiving Falco output: %w", err)
			}

			// Parse Falco output
			event := s.parseFalcoOutput(res)
			if event != nil {
				select {
				case eventCh <- *event:
					log.Debugf("Sent Falco event to channel: %s", res.Rule)
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}
	}

	log.Info("Falco subscriber stopped")
	return nil
}

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
		"ModifyInstanceAttribute":          true,
		"ModifyNetworkInterfaceAttribute":  true,
		"ModifyVolume":                     true,

		// IAM - Policy modifications
		"PutUserPolicy":                    true,
		"PutRolePolicy":                    true,
		"PutGroupPolicy":                   true,
		"UpdateAssumeRolePolicy":           true,
		"AttachUserPolicy":                 true,
		"AttachRolePolicy":                 true,
		"AttachGroupPolicy":                true,
		"CreatePolicy":                     true,
		"CreatePolicyVersion":              true,

		// IAM - User/Role/Group lifecycle
		"CreateRole":                       true,
		"DeleteRole":                       true,
		"CreateUser":                       true,
		"DeleteUser":                       true,
		"CreateAccessKey":                  true,
		"AddUserToGroup":                   true,
		"RemoveUserFromGroup":              true,
		"UpdateAccountPasswordPolicy":      true,

		// S3
		"PutBucketPolicy":                  true,
		"PutBucketVersioning":              true,
		"PutBucketEncryption":              true,
		"DeleteBucketEncryption":           true,
		"PutBucketLogging":                 true,

		// RDS
		"ModifyDBInstance":                 true,
		"ModifyDBCluster":                  true,

		// Lambda
		"UpdateFunctionConfiguration":     true,
		"UpdateFunctionCode":               true,
	}

	return relevantEvents[eventName]
}

// extractResourceID extracts the resource ID from Falco output fields
func (s *Subscriber) extractResourceID(eventName string, fields map[string]string) string {
	// Try different field names based on event type
	idFieldMap := map[string][]string{
		"ModifyInstanceAttribute":      {"ct.request.instanceid", "ct.resource.instanceid"},
		"ModifyVolume":                 {"ct.request.volumeid"},
		"PutBucketPolicy":              {"ct.request.bucket", "ct.resource.bucket"},
		"PutBucketEncryption":          {"ct.request.bucket"},
		"DeleteBucketEncryption":       {"ct.request.bucket"},
		"ModifyDBInstance":             {"ct.request.dbinstanceidentifier"},
		"UpdateFunctionConfiguration":  {"ct.request.functionname"},

		// IAM - Roles
		"PutRolePolicy":                {"ct.request.rolename"},
		"UpdateAssumeRolePolicy":       {"ct.request.rolename"},
		"AttachRolePolicy":             {"ct.request.rolename"},
		"CreateRole":                   {"ct.request.rolename"},
		"DeleteRole":                   {"ct.request.rolename"},

		// IAM - Users
		"PutUserPolicy":                {"ct.request.username"},
		"AttachUserPolicy":             {"ct.request.username"},
		"CreateUser":                   {"ct.request.username"},
		"DeleteUser":                   {"ct.request.username"},
		"CreateAccessKey":              {"ct.request.username"},
		"AddUserToGroup":               {"ct.request.username"},
		"RemoveUserFromGroup":          {"ct.request.username"},

		// IAM - Groups
		"PutGroupPolicy":               {"ct.request.groupname"},
		"AttachGroupPolicy":            {"ct.request.groupname"},

		// IAM - Policies
		"CreatePolicy":                 {"ct.request.policyname"},
		"CreatePolicyVersion":          {"ct.request.policyarn"},
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

// extractChanges extracts the changed attributes from Falco output
func (s *Subscriber) extractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	switch eventName {
	case "ModifyInstanceAttribute":
		if val, ok := fields["ct.request.disableapitermination"]; ok && val != "" {
			changes["disable_api_termination"] = val
		}
		if val, ok := fields["ct.request.instancetype"]; ok && val != "" {
			changes["instance_type"] = val
		}

	case "PutBucketEncryption":
		// Encryption enabled
		if config, ok := fields["ct.request.serversideencryptionconfiguration"]; ok && config != "" {
			changes["server_side_encryption_configuration"] = config
		}

	case "DeleteBucketEncryption":
		// Encryption disabled
		changes["server_side_encryption_configuration"] = nil

	case "UpdateFunctionConfiguration":
		if val, ok := fields["ct.request.timeout"]; ok && val != "" {
			changes["timeout"] = val
		}
		if val, ok := fields["ct.request.memorysize"]; ok && val != "" {
			changes["memory_size"] = val
		}

	case "UpdateAssumeRolePolicy":
		if policy := getStringField(fields, "ct.request.policydocument"); policy != "" {
			var policyDoc map[string]interface{}
			if err := json.Unmarshal([]byte(policy), &policyDoc); err == nil {
				changes["assume_role_policy"] = policyDoc
			}
		}

	// IAM Policy attachments
	case "AttachRolePolicy", "AttachUserPolicy", "AttachGroupPolicy":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["attached_policy_arn"] = policyArn
		}

	// IAM Inline policies
	case "PutRolePolicy", "PutUserPolicy", "PutGroupPolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["inline_policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy creation
	case "CreatePolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy version
	case "CreatePolicyVersion":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["policy_arn"] = policyArn
		}
		if setDefault, ok := fields["ct.request.setasdefault"]; ok && setDefault != "" {
			changes["set_as_default"] = setDefault
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Role creation
	case "CreateRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["role_name"] = roleName
		}
		if assumePolicy := getStringField(fields, "ct.request.assumerolepolicydocument"); assumePolicy != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(assumePolicy), &doc); err == nil {
				changes["assume_role_policy"] = doc
			}
		}

	// IAM User/Role deletion
	case "DeleteRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["deleted_role"] = roleName
		}
	case "DeleteUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["deleted_user"] = userName
		}

	// IAM User creation
	case "CreateUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}

	// IAM Access Key creation
	case "CreateAccessKey":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if accessKeyId := getStringField(fields, "ct.response.accesskey.accesskeyid"); accessKeyId != "" {
			changes["access_key_id"] = accessKeyId
		}

	// IAM Group membership
	case "AddUserToGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}
	case "RemoveUserFromGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}

	// Account password policy
	case "UpdateAccountPasswordPolicy":
		// Extract various password policy fields if available
		if minLength, ok := fields["ct.request.minimumpasswordlength"]; ok && minLength != "" {
			changes["minimum_password_length"] = minLength
		}
		if requireSymbols, ok := fields["ct.request.requiresymbols"]; ok && requireSymbols != "" {
			changes["require_symbols"] = requireSymbols
		}
	}

	return changes
}

// mapEventToResourceType maps a CloudTrail event name to a Terraform resource type
func (s *Subscriber) mapEventToResourceType(eventName string) string {
	mapping := map[string]string{
		// EC2
		"ModifyInstanceAttribute":          "aws_instance",
		"ModifyVolume":                     "aws_ebs_volume",

		// IAM - Roles
		"PutRolePolicy":                    "aws_iam_role_policy",
		"UpdateAssumeRolePolicy":           "aws_iam_role",
		"AttachRolePolicy":                 "aws_iam_role_policy_attachment",
		"CreateRole":                       "aws_iam_role",
		"DeleteRole":                       "aws_iam_role",

		// IAM - Users
		"PutUserPolicy":                    "aws_iam_user_policy",
		"AttachUserPolicy":                 "aws_iam_user_policy_attachment",
		"CreateUser":                       "aws_iam_user",
		"DeleteUser":                       "aws_iam_user",
		"CreateAccessKey":                  "aws_iam_access_key",
		"AddUserToGroup":                   "aws_iam_user_group_membership",
		"RemoveUserFromGroup":              "aws_iam_user_group_membership",

		// IAM - Groups
		"PutGroupPolicy":                   "aws_iam_group_policy",
		"AttachGroupPolicy":                "aws_iam_group_policy_attachment",

		// IAM - Policies
		"CreatePolicy":                     "aws_iam_policy",
		"CreatePolicyVersion":              "aws_iam_policy",

		// IAM - Account
		"UpdateAccountPasswordPolicy":      "aws_iam_account_password_policy",

		// S3
		"PutBucketPolicy":                  "aws_s3_bucket_policy",
		"PutBucketVersioning":              "aws_s3_bucket",
		"PutBucketEncryption":              "aws_s3_bucket",
		"DeleteBucketEncryption":           "aws_s3_bucket",

		// RDS
		"ModifyDBInstance":                 "aws_db_instance",

		// Lambda
		"UpdateFunctionConfiguration":     "aws_lambda_function",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}

// getStringField safely gets a string field from the map
func getStringField(fields map[string]string, key string) string {
	// Try direct lookup
	if val, ok := fields[key]; ok {
		return val
	}

	// Try case-insensitive lookup (Falco might use different casing)
	lowerKey := strings.ToLower(key)
	for k, v := range fields {
		if strings.ToLower(k) == lowerKey {
			return v
		}
	}

	return ""
}
