package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

// TestAWSClientImplementsInterfaces verifies that real AWS SDK clients implement our interfaces
func TestAWSClientImplementsInterfaces(t *testing.T) {
	// These assignments will fail at compile time if the interfaces don't match
	var _ EC2API = (*ec2.Client)(nil)
	var _ RDSAPI = (*rds.Client)(nil)
	var _ EKSAPI = (*eks.Client)(nil)
	var _ ElastiCacheAPI = (*elasticache.Client)(nil)
	var _ ELBAPI = (*elasticloadbalancingv2.Client)(nil)

	t.Log("All AWS SDK clients correctly implement their respective interfaces")
}
