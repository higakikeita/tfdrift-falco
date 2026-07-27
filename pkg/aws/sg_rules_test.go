package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func awsTCP(from, to int32, cidr string) ec2Types.IpPermission {
	return ec2Types.IpPermission{
		IpProtocol: aws.String("tcp"),
		FromPort:   aws.Int32(from),
		ToPort:     aws.Int32(to),
		IpRanges:   []ec2Types.IpRange{{CidrIp: aws.String(cidr)}},
	}
}

// The AWS API shape and the Terraform-state shape for the same rule must produce
// the same canonical string, so an unchanged SG shows no drift.
func TestCanonicalize_AWSAndTF_AgreeOnSameRule(t *testing.T) {
	awsRules := canonicalizeAWSPermissions([]ec2Types.IpPermission{awsTCP(443, 443, "10.0.0.0/8")})
	tfRules := canonicalizeTFRules([]interface{}{
		map[string]interface{}{
			"protocol": "tcp", "from_port": float64(443), "to_port": float64(443),
			"cidr_blocks": []interface{}{"10.0.0.0/8"},
		},
	})
	assert.Equal(t, awsRules, tfRules)
	assert.Len(t, awsRules, 1)
}

// Rule order must not matter (sets, not lists).
func TestCanonicalize_OrderInsensitive(t *testing.T) {
	a := canonicalizeAWSPermissions([]ec2Types.IpPermission{awsTCP(80, 80, "0.0.0.0/0"), awsTCP(443, 443, "0.0.0.0/0")})
	b := canonicalizeAWSPermissions([]ec2Types.IpPermission{awsTCP(443, 443, "0.0.0.0/0"), awsTCP(80, 80, "0.0.0.0/0")})
	assert.Equal(t, a, b)
}

// "all" (Terraform) and "-1" (AWS) are the same protocol.
func TestNormalizeProto_AllEqualsMinusOne(t *testing.T) {
	assert.Equal(t, "-1", normalizeProto("all"))
	assert.Equal(t, "-1", normalizeProto("-1"))
	assert.Equal(t, "tcp", normalizeProto("TCP"))
}

// The money-shot drift: an ingress rule present in the cloud but not in state
// must surface as an `ingress` field diff.
func TestCompareSG_IngressDriftDetected(t *testing.T) {
	tfRes := &terraform.Resource{
		Type: "aws_security_group", Name: "web",
		Attributes: map[string]interface{}{
			"id":     "sg-1",
			"vpc_id": "vpc-1", "description": "web", "name": "web",
			"ingress": []interface{}{
				map[string]interface{}{"protocol": "tcp", "from_port": float64(443), "to_port": float64(443), "cidr_blocks": []interface{}{"10.0.0.0/8"}},
			},
		},
	}
	awsRes := &types.DiscoveredResource{
		Type: "aws_security_group", ID: "sg-1",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-1", "description": "web", "name": "web",
			// drift: an extra rule opened to the world (8443 from 203.0.113.7/32)
			"ingress": canonicalizeAWSPermissions([]ec2Types.IpPermission{
				awsTCP(443, 443, "10.0.0.0/8"),
				awsTCP(8443, 8443, "203.0.113.7/32"),
			}),
			"egress": []string{},
		},
	}

	diffs := compareResourceAttributes(tfRes, awsRes)
	var found bool
	for _, d := range diffs {
		if d.Field == "ingress" {
			found = true
		}
	}
	assert.True(t, found, "an added ingress rule must be reported as ingress drift")
}

// No drift when state and cloud carry the same rule set (no false positive).
func TestCompareSG_NoDriftWhenRulesMatch(t *testing.T) {
	rules := []ec2Types.IpPermission{awsTCP(443, 443, "10.0.0.0/8")}
	tfRes := &terraform.Resource{
		Type: "aws_security_group", Name: "web",
		Attributes: map[string]interface{}{
			"id": "sg-1", "vpc_id": "vpc-1", "description": "web", "name": "web",
			"ingress": []interface{}{
				map[string]interface{}{"protocol": "tcp", "from_port": float64(443), "to_port": float64(443), "cidr_blocks": []interface{}{"10.0.0.0/8"}},
			},
		},
	}
	awsRes := &types.DiscoveredResource{
		Type: "aws_security_group", ID: "sg-1",
		Attributes: map[string]interface{}{
			"vpc_id": "vpc-1", "description": "web", "name": "web",
			"ingress": canonicalizeAWSPermissions(rules),
			"egress":  []string{},
		},
	}
	for _, d := range compareResourceAttributes(tfRes, awsRes) {
		assert.NotEqual(t, "ingress", d.Field, "matching rule sets must not report ingress drift")
	}
}

// `missing` must be scoped to discoverable types: a tf resource of an unscanned
// type (e.g. aws_iam_role) is not "missing", just not scanned (#338).
func TestCompareStateWithActual_MissingScopedToDiscoverableTypes(t *testing.T) {
	tf := []*terraform.Resource{
		{Type: "aws_iam_role", Name: "r", Attributes: map[string]interface{}{"id": "role-x"}},    // unscanned type
		{Type: "aws_instance", Name: "i", Attributes: map[string]interface{}{"id": "i-deleted"}}, // scanned + gone = real missing
	}
	// Cloud discovery found nothing (both absent), but only the aws_instance
	// should be reported missing.
	result := CompareStateWithActual(tf, nil)

	var types_ []string
	for _, m := range result.MissingResources {
		types_ = append(types_, m.Type)
	}
	assert.Contains(t, types_, "aws_instance", "a deleted instance is genuinely missing")
	assert.NotContains(t, types_, "aws_iam_role", "an unscanned type must not be reported missing")
}
