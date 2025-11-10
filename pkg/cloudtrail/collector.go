package cloudtrail

import (
	"context"
	"encoding/json"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Collector collects CloudTrail events
type Collector struct {
	cfg config.CloudTrailConfig
}

// CloudTrailEvent represents a CloudTrail event
type CloudTrailEvent struct {
	EventVersion string    `json:"eventVersion"`
	UserIdentity UserIdentity `json:"userIdentity"`
	EventTime    time.Time `json:"eventTime"`
	EventSource  string    `json:"eventSource"`
	EventName    string    `json:"eventName"`
	AWSRegion    string    `json:"awsRegion"`
	SourceIPAddress string `json:"sourceIPAddress"`
	UserAgent    string    `json:"userAgent"`
	RequestParameters map[string]interface{} `json:"requestParameters"`
	ResponseElements  map[string]interface{} `json:"responseElements"`
}

// UserIdentity represents the user who made the API call
type UserIdentity struct {
	Type        string `json:"type"`
	PrincipalID string `json:"principalId"`
	ARN         string `json:"arn"`
	AccountID   string `json:"accountId"`
	UserName    string `json:"userName"`
}

// NewCollector creates a new CloudTrail collector
func NewCollector(cfg config.CloudTrailConfig) (*Collector, error) {
	return &Collector{
		cfg: cfg,
	}, nil
}

// Start starts collecting CloudTrail events
func (c *Collector) Start(ctx context.Context, eventCh chan<- types.Event) error {
	log.Info("CloudTrail collector started")

	// TODO: Implement actual SQS polling or S3 event processing
	// For now, this is a placeholder that demonstrates the structure

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("CloudTrail collector stopping...")
			return nil

		case <-ticker.C:
			// Poll for events
			events, err := c.pollEvents(ctx)
			if err != nil {
				log.Errorf("Failed to poll CloudTrail events: %v", err)
				continue
			}

			// Process events
			for _, ctEvent := range events {
				event := c.convertToDetectorEvent(ctEvent)
				if event != nil {
					eventCh <- *event
				}
			}
		}
	}
}

// pollEvents polls for new CloudTrail events
func (c *Collector) pollEvents(ctx context.Context) ([]CloudTrailEvent, error) {
	// TODO: Implement SQS queue polling or S3 bucket scanning
	// This is a placeholder
	return []CloudTrailEvent{}, nil
}

// convertToDetectorEvent converts a CloudTrail event to a detector event
func (c *Collector) convertToDetectorEvent(ctEvent CloudTrailEvent) *types.Event {
	// Filter events we care about
	if !c.isRelevantEvent(ctEvent) {
		return nil
	}

	resourceID := c.extractResourceID(ctEvent)
	if resourceID == "" {
		return nil
	}

	return &types.Event{
		Provider:     "aws",
		EventName:    ctEvent.EventName,
		ResourceType: c.mapEventToResourceType(ctEvent.EventName),
		ResourceID:   resourceID,
		UserIdentity: types.UserIdentity{
			Type:        ctEvent.UserIdentity.Type,
			PrincipalID: ctEvent.UserIdentity.PrincipalID,
			ARN:         ctEvent.UserIdentity.ARN,
			AccountID:   ctEvent.UserIdentity.AccountID,
			UserName:    ctEvent.UserIdentity.UserName,
		},
		Changes: c.extractChanges(ctEvent),
		RawEvent: ctEvent,
	}
}

// isRelevantEvent checks if an event is relevant for drift detection
func (c *Collector) isRelevantEvent(event CloudTrailEvent) bool {
	relevantEvents := map[string]bool{
		// EC2
		"ModifyInstanceAttribute":     true,
		"ModifyNetworkInterfaceAttribute": true,
		"ModifyVolume":                true,

		// IAM
		"PutUserPolicy":               true,
		"PutRolePolicy":               true,
		"UpdateAssumeRolePolicy":      true,
		"AttachUserPolicy":            true,
		"AttachRolePolicy":            true,

		// S3
		"PutBucketPolicy":             true,
		"PutBucketVersioning":         true,
		"PutBucketEncryption":         true,
		"PutBucketLogging":            true,

		// RDS
		"ModifyDBInstance":            true,
		"ModifyDBCluster":             true,

		// Lambda
		"UpdateFunctionConfiguration": true,
		"UpdateFunctionCode":          true,
	}

	return relevantEvents[event.EventName]
}

// extractResourceID extracts the resource ID from the event
func (c *Collector) extractResourceID(event CloudTrailEvent) string {
	// Try different fields based on event type
	params := event.RequestParameters
	if params == nil {
		return ""
	}

	// Common ID fields
	idFields := []string{
		"instanceId",
		"volumeId",
		"bucketName",
		"functionName",
		"dBInstanceIdentifier",
		"roleName",
		"userName",
		"policyArn",
	}

	for _, field := range idFields {
		if id, ok := params[field].(string); ok && id != "" {
			return id
		}
	}

	// Try ARN from response
	if event.ResponseElements != nil {
		if arn, ok := event.ResponseElements["arn"].(string); ok {
			return arn
		}
	}

	return ""
}

// extractChanges extracts the changed attributes from the event
func (c *Collector) extractChanges(event CloudTrailEvent) map[string]interface{} {
	changes := make(map[string]interface{})

	// Extract changes based on event type
	switch event.EventName {
	case "ModifyInstanceAttribute":
		if val, ok := event.RequestParameters["disableApiTermination"]; ok {
			changes["disable_api_termination"] = val
		}
		if val, ok := event.RequestParameters["instanceType"]; ok {
			changes["instance_type"] = val
		}

	case "PutBucketPolicy":
		if policy, ok := event.RequestParameters["bucketPolicy"].(string); ok {
			var policyDoc map[string]interface{}
			json.Unmarshal([]byte(policy), &policyDoc)
			changes["policy"] = policyDoc
		}

	case "UpdateFunctionConfiguration":
		if val, ok := event.RequestParameters["timeout"]; ok {
			changes["timeout"] = val
		}
		if val, ok := event.RequestParameters["memorySize"]; ok {
			changes["memory_size"] = val
		}
	}

	return changes
}

// mapEventToResourceType maps an event name to a Terraform resource type
func (c *Collector) mapEventToResourceType(eventName string) string {
	mapping := map[string]string{
		"ModifyInstanceAttribute":          "aws_instance",
		"ModifyVolume":                     "aws_ebs_volume",
		"PutUserPolicy":                    "aws_iam_user_policy",
		"PutRolePolicy":                    "aws_iam_role_policy",
		"UpdateAssumeRolePolicy":           "aws_iam_role",
		"PutBucketPolicy":                  "aws_s3_bucket_policy",
		"PutBucketVersioning":              "aws_s3_bucket",
		"ModifyDBInstance":                 "aws_db_instance",
		"UpdateFunctionConfiguration":     "aws_lambda_function",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
