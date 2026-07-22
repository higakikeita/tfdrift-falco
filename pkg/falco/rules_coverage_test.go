package falco

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestFalcoRule_MatchesBackendCapability locks the Falco drift rule
// (rules/terraform_drift.yaml) to the backend's authoritative capability set
// (pkg/falco/configs/event_mappings.yaml).
//
// Regression for #324 ("detection ceiling"): the rule used to fire on ~24
// events while the backend could resolve 245+, so representative drift
// (security-group ingress, ModifyVpcAttribute, PutBucketAcl, ModifyDBInstance,
// ...) was mapped in the tables but never selected by Falco — silently never
// inspected. This test fails if the two layers drift apart in either direction.
func TestFalcoRule_MatchesBackendCapability(t *testing.T) {
	var cfg struct {
		RelevantEvents   []string                 `yaml:"relevant_events"`
		ResourceIDFields map[string]interface{}   `yaml:"resource_id_fields"`
	}
	data, err := os.ReadFile(filepath.Join("configs", "event_mappings.yaml"))
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(data, &cfg))

	relevant := make(map[string]bool, len(cfg.RelevantEvents))
	for _, e := range cfg.RelevantEvents {
		relevant[e] = true
	}
	// flow-through = an event that is relevant AND has a resource-ID mapping,
	// i.e. one that can actually reach state comparison (others are dropped in
	// event_parser.extractResourceID and would be pure Falco noise).
	flow := make(map[string]bool)
	for e := range cfg.ResourceIDFields {
		if relevant[e] {
			flow[e] = true
		}
	}

	ruleData, err := os.ReadFile(filepath.Join("..", "..", "rules", "terraform_drift.yaml"))
	require.NoError(t, err)
	// Only event names are quoted in the rule file (outputs/tags/priority are not).
	re := regexp.MustCompile(`"([A-Za-z0-9]+)"`)
	ruleEvents := make(map[string]bool)
	for _, m := range re.FindAllStringSubmatch(string(ruleData), -1) {
		ruleEvents[m[1]] = true
	}

	// 1) The rule must not fire on events the backend cannot resolve to a
	//    resource ID — those would be Falco noise, silently dropped downstream.
	var ruleButNotFlow []string
	for e := range ruleEvents {
		if !flow[e] {
			ruleButNotFlow = append(ruleButNotFlow, e)
		}
	}
	sort.Strings(ruleButNotFlow)
	require.Empty(t, ruleButNotFlow,
		"Falco rule fires on events that lack resource_id_fields (silently dropped downstream): %v", ruleButNotFlow)

	// 2) The rule must cover every flow-through event: if the backend can fully
	//    process an event, Falco must actually select it (no detection ceiling).
	var flowButNotRule []string
	for e := range flow {
		if !ruleEvents[e] {
			flowButNotRule = append(flowButNotRule, e)
		}
	}
	sort.Strings(flowButNotRule)
	require.Empty(t, flowButNotRule,
		"backend can process these events but the Falco rule never fires on them (detection ceiling, #324): %v", flowButNotRule)

	// 3) Representative drift events must be present — a human-readable guard so
	//    a well-meaning refactor can't quietly shrink coverage back down.
	for _, e := range []string{
		"AuthorizeSecurityGroupIngress", "RevokeSecurityGroupIngress",
		"ModifyVpcAttribute", "ModifySubnetAttribute",
		"PutBucketAcl", "PutBucketPolicy", "PutBucketVersioning", "PutBucketPublicAccessBlock",
		"ModifyDBInstance", "PutRolePolicy", "AttachRolePolicy", "ModifySecurityGroupRules",
	} {
		require.Truef(t, ruleEvents[e], "representative drift event %q must be selected by the Falco rule", e)
	}

	require.GreaterOrEqualf(t, len(ruleEvents), 200,
		"Falco rule should cover the broad drift surface (got %d events)", len(ruleEvents))
}
