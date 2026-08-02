//go:build liveaws

// Live-AWS re-verification of the #338 comparison-layer fixes. Runs only with
// `-tags liveaws` against a real account (never in normal CI). It confirms, on
// real infrastructure, that: (1) security-group ingress rules are discovered and
// canonicalized, (2) a rule present in the cloud but not in state is reported as
// ingress drift, (3) no attribute leaks a pointer address, and (4) `missing` is
// scoped to scanned types.
package aws

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLive_ComparatorCorrectness(t *testing.T) {
	region := envOr("TFDRIFT_LIVE_REGION", "ap-northeast-1")
	sgID := envOr("TFDRIFT_LIVE_SG", "sg-056b65b187cd8ece3")

	ctx := context.Background()
	dc, err := NewDiscoveryClient(ctx, region)
	require.NoError(t, err)

	all, err := dc.DiscoverAll(ctx)
	require.NoError(t, err)
	t.Logf("discovered %d resources in %s", len(all), region)

	// 1) The lab SG is discovered and its ingress rules are populated as a
	//    canonical string set (the gap #338 fixed).
	var sg *DiscoveredResource
	for _, r := range all {
		if r.ID == sgID {
			sg = r
		}
	}
	require.NotNil(t, sg, "lab SG %s must be discovered", sgID)
	ingress := toStringSliceVal(sg.Attributes["ingress"])
	require.NotEmpty(t, ingress, "SG ingress rules must be discovered (was silently dropped)")
	t.Logf("SG %s ingress canonical rules: %v", sgID, ingress)

	// 2) Pointer-leak regression: no discovered attribute renders as a 0x address.
	for _, r := range all {
		for k, v := range r.Attributes {
			assert.NotContains(t, fmt.Sprintf("%v", v), "0x",
				"attribute %s.%s leaks a pointer address", r.ID, k)
		}
	}

	// 3) State with no ingress vs a cloud SG that HAS rules → ingress drift.
	tfSG := &terraform.Resource{
		Type: "aws_security_group", Name: "lab",
		Attributes: map[string]interface{}{
			"id":          sgID,
			"vpc_id":      sg.Attributes["vpc_id"],
			"description": sg.Attributes["description"],
			"name":        sg.Attributes["name"],
			// ingress intentionally omitted → cloud rules are drift
		},
	}
	res := CompareStateWithActual([]*terraform.Resource{tfSG}, all)
	var ingressDrift bool
	for _, m := range res.ModifiedResources {
		if m.ResourceID == sgID {
			for _, d := range m.Differences {
				if d.Field == "ingress" {
					ingressDrift = true
					t.Logf("ingress drift reported: tf=%v actual=%v", d.TerraformValue, d.ActualValue)
				}
			}
		}
	}
	assert.True(t, ingressDrift, "cloud SG rules absent from state must be reported as ingress drift")

	// 4) missing scoping: an unscanned type is never reported missing.
	tfIAM := &terraform.Resource{Type: "aws_iam_role", Name: "x", Attributes: map[string]interface{}{"id": "role-not-scanned"}}
	res2 := CompareStateWithActual([]*terraform.Resource{tfIAM}, all)
	for _, m := range res2.MissingResources {
		assert.NotEqual(t, "aws_iam_role", m.Type, "an unscanned type must not be reported missing")
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
