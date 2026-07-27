package aws

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Security-group rule comparison (#338 / #362).
//
// A security group's ingress/egress rules are the highest-value drift to catch
// (AuthorizeSecurityGroupIngress is the demo's headline event), but the AWS API
// shape (IpPermission) and the Terraform-state shape (ingress blocks) differ, and
// rule order is not significant. We normalize both sides to a *sorted set of
// canonical rule strings* so they can be compared with a plain DeepEqual without
// false positives from ordering or representation.
//
// Canonical form: "proto|from|to|cidrs|ipv6|sgrefs" with each list sorted.

func canonicalRuleString(proto string, from, to int64, cidrs, ipv6, sgRefs []string) string {
	sort.Strings(cidrs)
	sort.Strings(ipv6)
	sort.Strings(sgRefs)
	return fmt.Sprintf("%s|%d|%d|%s|%s|%s",
		normalizeProto(proto), from, to,
		strings.Join(cidrs, ","), strings.Join(ipv6, ","), strings.Join(sgRefs, ","))
}

// normalizeProto folds Terraform's "all" onto AWS's "-1" so the two agree.
func normalizeProto(p string) string {
	p = strings.ToLower(strings.TrimSpace(p))
	if p == "all" || p == "" {
		return "-1"
	}
	return p
}

// canonicalizeAWSPermissions normalizes AWS IpPermissions into a sorted canonical
// rule set. Ports are 0 when absent (e.g. protocol -1 / all).
func canonicalizeAWSPermissions(perms []ec2Types.IpPermission) []string {
	out := make([]string, 0, len(perms))
	for _, p := range perms {
		var from, to int64
		if p.FromPort != nil {
			from = int64(*p.FromPort)
		}
		if p.ToPort != nil {
			to = int64(*p.ToPort)
		}
		var cidrs, ipv6, sgRefs []string
		for _, r := range p.IpRanges {
			cidrs = append(cidrs, aws.ToString(r.CidrIp))
		}
		for _, r := range p.Ipv6Ranges {
			ipv6 = append(ipv6, aws.ToString(r.CidrIpv6))
		}
		for _, g := range p.UserIdGroupPairs {
			sgRefs = append(sgRefs, aws.ToString(g.GroupId))
		}
		out = append(out, canonicalRuleString(aws.ToString(p.IpProtocol), from, to, cidrs, ipv6, sgRefs))
	}
	sort.Strings(out)
	return out
}

// canonicalizeTFRules normalizes Terraform-state ingress/egress blocks (a list of
// {from_port,to_port,protocol,cidr_blocks,...}) into the same sorted canonical set.
func canonicalizeTFRules(blocks interface{}) []string {
	out := []string{}
	for _, b := range toSliceOfMaps(blocks) {
		out = append(out, canonicalRuleString(
			toStringVal(b["protocol"]),
			toInt64Val(b["from_port"]),
			toInt64Val(b["to_port"]),
			toStringSliceVal(b["cidr_blocks"]),
			toStringSliceVal(b["ipv6_cidr_blocks"]),
			toStringSliceVal(b["security_groups"]),
		))
	}
	sort.Strings(out)
	return out
}

// --- small, total conversion helpers (Terraform state values are untyped) ---

func toSliceOfMaps(v interface{}) []map[string]interface{} {
	var out []map[string]interface{}
	switch s := v.(type) {
	case []interface{}:
		for _, e := range s {
			if m, ok := e.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
	case []map[string]interface{}:
		out = append(out, s...)
	}
	return out
}

func toStringVal(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func toInt64Val(v interface{}) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	}
	return 0
}

func toStringSliceVal(v interface{}) []string {
	var out []string
	switch s := v.(type) {
	case []interface{}:
		for _, e := range s {
			if str, ok := e.(string); ok {
				out = append(out, str)
			}
		}
	case []string:
		out = append(out, s...)
	}
	return out
}
