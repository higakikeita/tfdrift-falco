// Package falco exported helpers for use by the provider package.
// These functions wrap Subscriber methods to make them accessible
// without requiring a Subscriber instance.
package falco

// GetAWSRelevantEvents returns a map of all AWS event names that trigger drift detection.
func GetAWSRelevantEvents() map[string]bool {
	return buildRelevantEventsMap()
}

// buildRelevantEventsMap returns the complete set of AWS events that are relevant for drift detection.
func buildRelevantEventsMap() map[string]bool {
	// This must stay in sync with Subscriber.isRelevantEvent
	return map[string]bool{
		// EC2
		"ModifyInstanceAttribute":         true,
		"ModifyNetworkInterfaceAttribute": true,
		"ModifyVolume":                    true,
		"RunInstances":                    true,
		"TerminateInstances":              true,
		"StartInstances":                  true,
		"StopInstances":                   true,
		"RebootInstances":                 true,

		// Security Groups
		"AuthorizeSecurityGroupIngress": true,
		"AuthorizeSecurityGroupEgress":  true,
		"RevokeSecurityGroupIngress":    true,
		"RevokeSecurityGroupEgress":     true,
		"CreateSecurityGroup":           true,
		"DeleteSecurityGroup":           true,
		"ModifySecurityGroupRules":      true,

		// VPC
		"CreateVpc":                     true,
		"DeleteVpc":                     true,
		"ModifyVpcAttribute":            true,
		"CreateSubnet":                  true,
		"DeleteSubnet":                  true,
		"ModifySubnetAttribute":         true,
		"CreateRouteTable":              true,
		"DeleteRouteTable":              true,
		"CreateRoute":                   true,
		"DeleteRoute":                   true,
		"ReplaceRoute":                  true,
		"AssociateRouteTable":           true,
		"DisassociateRouteTable":        true,
		"CreateInternetGateway":         true,
		"DeleteInternetGateway":         true,
		"AttachInternetGateway":         true,
		"DetachInternetGateway":         true,
		"CreateNatGateway":              true,
		"DeleteNatGateway":              true,

		// ELB
		"CreateLoadBalancer":            true,
		"DeleteLoadBalancer":            true,
		"ModifyLoadBalancerAttributes":  true,
		"CreateTargetGroup":             true,
		"DeleteTargetGroup":             true,
		"ModifyTargetGroupAttributes":   true,
		"CreateListener":                true,
		"DeleteListener":                true,
		"ModifyListener":                true,

		// S3
		"CreateBucket":                  true,
		"DeleteBucket":                  true,
		"PutBucketPolicy":               true,
		"DeleteBucketPolicy":            true,
		"PutBucketAcl":                  true,
		"PutBucketVersioning":           true,
		"PutBucketEncryption":           true,
		"DeleteBucketEncryption":        true,
		"PutBucketLogging":              true,
		"PutBucketPublicAccessBlock":    true,

		// RDS
		"CreateDBInstance":              true,
		"DeleteDBInstance":              true,
		"ModifyDBInstance":              true,
		"CreateDBSubnetGroup":           true,
		"DeleteDBSubnetGroup":           true,
		"ModifyDBSubnetGroup":           true,
		"CreateDBParameterGroup":        true,
		"DeleteDBParameterGroup":        true,
		"ModifyDBParameterGroup":        true,

		// IAM
		"CreateRole":                    true,
		"DeleteRole":                    true,
		"AttachRolePolicy":              true,
		"DetachRolePolicy":              true,
		"PutRolePolicy":                 true,
		"DeleteRolePolicy":              true,
		"CreateUser":                    true,
		"DeleteUser":                    true,
		"AttachUserPolicy":              true,
		"DetachUserPolicy":              true,
		"CreateGroup":                   true,
		"DeleteGroup":                   true,
		"CreatePolicy":                  true,
		"DeletePolicy":                  true,
		"CreateInstanceProfile":         true,
		"DeleteInstanceProfile":         true,
		"AddRoleToInstanceProfile":      true,
		"RemoveRoleFromInstanceProfile": true,

		// EKS
		"CreateCluster":                 true,
		"DeleteCluster":                 true,
		"UpdateClusterConfig":           true,
		"CreateNodegroup":               true,
		"DeleteNodegroup":               true,
		"UpdateNodegroupConfig":         true,

		// Lambda
		"CreateFunction20150331":        true,
		"DeleteFunction20150331":        true,
		"UpdateFunctionConfiguration20150331v2": true,
		"UpdateFunctionCode20150331v2":          true,
		"AddPermission20150331v2":               true,
		"RemovePermission20150331v2":             true,

		// ElastiCache
		"CreateCacheCluster":               true,
		"DeleteCacheCluster":               true,
		"ModifyCacheCluster":               true,
		"CreateReplicationGroup":            true,
		"DeleteReplicationGroup":            true,
		"ModifyReplicationGroup":            true,
		"CreateCacheSubnetGroup":            true,
		"DeleteCacheSubnetGroup":            true,
		"ModifyCacheSubnetGroup":            true,

		// CloudFront
		"CreateDistribution":             true,
		"DeleteDistribution":             true,
		"UpdateDistribution":             true,

		// SNS
		"CreateTopic":                    true,
		"DeleteTopic":                    true,
		"SetTopicAttributes":             true,
		"Subscribe":                      true,
		"Unsubscribe":                    true,

		// SQS
		"CreateQueue":                    true,
		"DeleteQueue":                    true,
		"SetQueueAttributes":             true,

		// DynamoDB
		"CreateTable":                    true,
		"DeleteTable":                    true,
		"UpdateTable":                    true,

		// KMS
		"CreateKey":                      true,
		"DisableKey":                     true,
		"EnableKey":                      true,
		"ScheduleKeyDeletion":            true,
		"CreateAlias":                    true,
		"DeleteAlias":                    true,
	}
}

// ExtractAWSResourceID extracts the resource ID from AWS CloudTrail event fields.
func ExtractAWSResourceID(eventName string, fields map[string]string) string {
	s := &Subscriber{}
	return s.extractResourceID(eventName, fields)
}

// ExtractAWSChanges extracts attribute changes from AWS CloudTrail event fields.
func ExtractAWSChanges(eventName string, fields map[string]string) map[string]interface{} {
	s := &Subscriber{}
	return s.extractChanges(eventName, fields)
}
