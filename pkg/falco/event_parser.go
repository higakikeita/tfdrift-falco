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

		// VPC - Security Groups (Critical)
		"AuthorizeSecurityGroupIngress": true,
		"AuthorizeSecurityGroupEgress":  true,
		"RevokeSecurityGroupIngress":    true,
		"RevokeSecurityGroupEgress":     true,
		"CreateSecurityGroup":           true,
		"DeleteSecurityGroup":           true,
		"ModifySecurityGroupRules":      true,

		// VPC - Core
		"CreateVpc":           true,
		"DeleteVpc":           true,
		"ModifyVpcAttribute":  true,
		"CreateSubnet":        true,
		"DeleteSubnet":        true,
		"ModifySubnetAttribute": true,

		// VPC - Route Tables (Critical)
		"CreateRoute":         true,
		"DeleteRoute":         true,
		"ReplaceRoute":        true,
		"CreateRouteTable":    true,
		"DeleteRouteTable":    true,
		"AssociateRouteTable": true,

		// VPC - Internet/NAT Gateways
		"AttachInternetGateway": true,
		"DetachInternetGateway": true,
		"CreateNatGateway":      true,
		"DeleteNatGateway":      true,

		// VPC - Network ACLs
		"CreateNetworkAcl":      true,
		"DeleteNetworkAcl":      true,
		"CreateNetworkAclEntry": true,
		"DeleteNetworkAclEntry": true,
		"ReplaceNetworkAclEntry": true,

		// VPC - VPC Endpoints
		"CreateVpcEndpoint": true,
		"DeleteVpcEndpoint": true,
		"ModifyVpcEndpoint": true,

		// ELB/ALB - Load Balancers
		"CreateLoadBalancer":             true,
		"DeleteLoadBalancer":             true,
		"ModifyLoadBalancerAttributes":   true,

		// ELB/ALB - Target Groups
		"CreateTargetGroup":              true,
		"DeleteTargetGroup":              true,
		"ModifyTargetGroup":              true,
		"ModifyTargetGroupAttributes":    true,
		"RegisterTargets":                true,
		"DeregisterTargets":              true,

		// ELB/ALB - Listeners & Rules (Critical)
		"CreateListener":                 true,
		"DeleteListener":                 true,
		"ModifyListener":                 true,
		"CreateRule":                     true,
		"DeleteRule":                     true,
		"ModifyRule":                     true,

		// KMS (Critical)
		"ScheduleKeyDeletion":  true,
		"DisableKey":           true,
		"EnableKey":            true,
		"PutKeyPolicy":         true,
		"CreateKey":            true,
		"CreateAlias":          true,
		"DeleteAlias":          true,
		"UpdateAlias":          true,
		"EnableKeyRotation":    true,
		"DisableKeyRotation":   true,

		// DynamoDB
		"CreateTable":          true,
		"DeleteTable":          true,
		"UpdateTable":          true,
		"UpdateTimeToLive":     true,
		"UpdateContinuousBackups": true,

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
		"PutBucketPolicy":               true,
		"PutBucketVersioning":           true,
		"PutBucketEncryption":           true,
		"DeleteBucketEncryption":        true,
		"PutBucketLogging":              true,
		"PutBucketPublicAccessBlock":    true,
		"DeleteBucketPublicAccessBlock": true,
		"PutBucketAcl":                  true,

		// RDS
		"ModifyDBInstance": true,
		"ModifyDBCluster":  true,

		// Lambda
		"UpdateFunctionConfiguration": true,
		"UpdateFunctionCode":          true,
		"AddPermission":               true,
		"RemovePermission":            true,
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

		// VPC - Security Groups
		"AuthorizeSecurityGroupIngress": {"ct.request.groupid", "ct.request.groupname"},
		"AuthorizeSecurityGroupEgress":  {"ct.request.groupid", "ct.request.groupname"},
		"RevokeSecurityGroupIngress":    {"ct.request.groupid", "ct.request.groupname"},
		"RevokeSecurityGroupEgress":     {"ct.request.groupid", "ct.request.groupname"},
		"CreateSecurityGroup":           {"ct.response.groupid", "ct.request.groupname"},
		"DeleteSecurityGroup":           {"ct.request.groupid", "ct.request.groupname"},
		"ModifySecurityGroupRules":      {"ct.request.groupid"},

		// VPC - Core
		"CreateVpc":            {"ct.response.vpcid", "ct.response.vpc.vpcid"},
		"DeleteVpc":            {"ct.request.vpcid"},
		"ModifyVpcAttribute":   {"ct.request.vpcid"},
		"CreateSubnet":         {"ct.response.subnetid", "ct.response.subnet.subnetid"},
		"DeleteSubnet":         {"ct.request.subnetid"},
		"ModifySubnetAttribute": {"ct.request.subnetid"},

		// VPC - Route Tables
		"CreateRoute":         {"ct.request.routetableid"},
		"DeleteRoute":         {"ct.request.routetableid"},
		"ReplaceRoute":        {"ct.request.routetableid"},
		"CreateRouteTable":    {"ct.response.routetableid", "ct.response.routetable.routetableid"},
		"DeleteRouteTable":    {"ct.request.routetableid"},
		"AssociateRouteTable": {"ct.request.routetableid"},

		// VPC - Gateways
		"AttachInternetGateway": {"ct.request.internetgatewayid"},
		"DetachInternetGateway": {"ct.request.internetgatewayid"},
		"CreateNatGateway":      {"ct.response.natgatewayid", "ct.response.natgateway.natgatewayid"},
		"DeleteNatGateway":      {"ct.request.natgatewayid"},

		// VPC - Network ACLs
		"CreateNetworkAcl":       {"ct.response.networkaclid"},
		"DeleteNetworkAcl":       {"ct.request.networkaclid"},
		"CreateNetworkAclEntry":  {"ct.request.networkaclid"},
		"DeleteNetworkAclEntry":  {"ct.request.networkaclid"},
		"ReplaceNetworkAclEntry": {"ct.request.networkaclid"},

		// VPC - Endpoints
		"CreateVpcEndpoint": {"ct.response.vpcendpointid"},
		"DeleteVpcEndpoint": {"ct.request.vpcendpointid"},
		"ModifyVpcEndpoint": {"ct.request.vpcendpointid"},

		// ELB/ALB
		"CreateLoadBalancer":           {"ct.response.loadbalancers.0.loadbalancerarn"},
		"DeleteLoadBalancer":           {"ct.request.loadbalancerarn"},
		"ModifyLoadBalancerAttributes": {"ct.request.loadbalancerarn"},
		"CreateTargetGroup":            {"ct.response.targetgroups.0.targetgrouparn"},
		"DeleteTargetGroup":            {"ct.request.targetgrouparn"},
		"ModifyTargetGroup":            {"ct.request.targetgrouparn"},
		"ModifyTargetGroupAttributes":  {"ct.request.targetgrouparn"},
		"RegisterTargets":              {"ct.request.targetgrouparn"},
		"DeregisterTargets":            {"ct.request.targetgrouparn"},
		"CreateListener":               {"ct.response.listeners.0.listenerarn"},
		"DeleteListener":               {"ct.request.listenerarn"},
		"ModifyListener":               {"ct.request.listenerarn"},
		"CreateRule":                   {"ct.response.rules.0.rulearn"},
		"DeleteRule":                   {"ct.request.rulearn"},
		"ModifyRule":                   {"ct.request.rulearn"},

		// KMS
		"ScheduleKeyDeletion": {"ct.request.keyid"},
		"DisableKey":          {"ct.request.keyid"},
		"EnableKey":           {"ct.request.keyid"},
		"PutKeyPolicy":        {"ct.request.keyid"},
		"CreateKey":           {"ct.response.keymetadata.keyid"},
		"CreateAlias":         {"ct.request.aliasname"},
		"DeleteAlias":         {"ct.request.aliasname"},
		"UpdateAlias":         {"ct.request.aliasname"},
		"EnableKeyRotation":   {"ct.request.keyid"},
		"DisableKeyRotation":  {"ct.request.keyid"},

		// DynamoDB
		"CreateTable":             {"ct.request.tablename"},
		"DeleteTable":             {"ct.request.tablename"},
		"UpdateTable":             {"ct.request.tablename"},
		"UpdateTimeToLive":        {"ct.request.tablename"},
		"UpdateContinuousBackups": {"ct.request.tablename"},

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

		// S3
		"PutBucketPublicAccessBlock":    {"ct.request.bucket"},
		"DeleteBucketPublicAccessBlock": {"ct.request.bucket"},
		"PutBucketAcl":                  {"ct.request.bucket"},

		// Lambda
		"AddPermission":    {"ct.request.functionname"},
		"RemovePermission": {"ct.request.functionname"},
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
