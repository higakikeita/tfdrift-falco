package falco

// ExtractAWSChanges extracts attribute changes from AWS CloudTrail event fields.
// This is exported for use by the provider abstraction layer.
func ExtractAWSChanges(eventName string, fields map[string]string) map[string]interface{} {
	s := &Subscriber{} // We only need the method, no state
	return s.extractChanges(eventName, fields)
}

// ExtractAWSResourceID extracts the resource ID from AWS CloudTrail event fields.
// This is exported for use by the provider abstraction layer.
func ExtractAWSResourceID(eventName string, fields map[string]string) string {
	s := &Subscriber{}
	return s.extractResourceID(eventName, fields)
}

// GetAWSRelevantEvents returns the map of relevant AWS CloudTrail event names.
// This is exported for use by the provider abstraction layer.
// NOTE: This must be kept in sync with Subscriber.isRelevantEvent.
func GetAWSRelevantEvents() map[string]bool {
	return buildRelevantEventsMap()
}

// buildRelevantEventsMap builds the map of relevant events.
// This duplicates the logic in Subscriber.isRelevantEvent to provide
// a standalone function that doesn't require checking each event individually.
func buildRelevantEventsMap() map[string]bool {
	return map[string]bool{
		// EC2
		"RunInstances":                    true,
		"TerminateInstances":              true,
		"StopInstances":                   true,
		"StartInstances":                  true,
		"ModifyInstanceAttribute":         true,
		"ModifyNetworkInterfaceAttribute": true,
		"ModifyVolume":                    true,

		// VPC - Security Groups
		"AuthorizeSecurityGroupIngress": true,
		"AuthorizeSecurityGroupEgress":  true,
		"RevokeSecurityGroupIngress":    true,
		"RevokeSecurityGroupEgress":     true,
		"CreateSecurityGroup":           true,
		"DeleteSecurityGroup":           true,
		"ModifySecurityGroupRules":      true,

		// VPC - Core
		"CreateVpc":             true,
		"DeleteVpc":             true,
		"ModifyVpcAttribute":    true,
		"CreateSubnet":          true,
		"DeleteSubnet":          true,
		"ModifySubnetAttribute": true,

		// VPC - Route Tables
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
		"CreateNetworkAcl":       true,
		"DeleteNetworkAcl":       true,
		"CreateNetworkAclEntry":  true,
		"DeleteNetworkAclEntry":  true,
		"ReplaceNetworkAclEntry": true,

		// VPC - VPC Endpoints
		"CreateVpcEndpoint": true,
		"DeleteVpcEndpoint": true,
		"ModifyVpcEndpoint": true,

		// S3
		"CreateBucket":        true,
		"DeleteBucket":        true,
		"PutBucketPolicy":     true,
		"DeleteBucketPolicy":  true,
		"PutBucketAcl":        true,
		"PutBucketVersioning": true,
		"PutBucketEncryption": true,
		"PutBucketLogging":    true,

		// IAM
		"CreateRole":          true,
		"DeleteRole":          true,
		"AttachRolePolicy":    true,
		"DetachRolePolicy":    true,
		"PutRolePolicy":       true,
		"DeleteRolePolicy":    true,
		"CreateUser":          true,
		"DeleteUser":          true,
		"AttachUserPolicy":    true,
		"DetachUserPolicy":    true,
		"CreatePolicy":        true,
		"DeletePolicy":        true,
		"CreatePolicyVersion": true,

		// RDS
		"CreateDBInstance": true,
		"DeleteDBInstance": true,
		"ModifyDBInstance": true,
		"CreateDBCluster":  true,
		"DeleteDBCluster":  true,
		"ModifyDBCluster":  true,

		// ELB
		"CreateLoadBalancer":           true,
		"DeleteLoadBalancer":           true,
		"ModifyLoadBalancerAttributes": true,
		"CreateTargetGroup":            true,
		"DeleteTargetGroup":            true,
		"ModifyTargetGroup":            true,

		// Lambda
		"CreateFunction20150331":                true,
		"DeleteFunction20150331":                true,
		"UpdateFunctionCode20150331v2":          true,
		"UpdateFunctionConfiguration20150331v2": true,

		// EKS
		"CreateCluster":       true,
		"DeleteCluster":       true,
		"UpdateClusterConfig": true,

		// ElastiCache
		"CreateReplicationGroup": true,
		"DeleteReplicationGroup": true,
		"ModifyReplicationGroup": true,
		"CreateCacheCluster":     true,
		"DeleteCacheCluster":     true,
		"ModifyCacheCluster":     true,
	}
}
