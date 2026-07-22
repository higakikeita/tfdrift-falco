package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// TestConfigExample_NoPhantomKeys guards config.yaml.example against
// "claim != reality" drift: every key in the shipped example must map to a
// real field on the Config struct. yaml.v3's KnownFields(true) fails the
// decode on any unknown key, so a phantom block (e.g. a documented option the
// loader silently ignores) breaks this test instead of misleading a user.
//
// Regression for #320: the example previously documented api:, advanced:,
// providers.aws.cloudtrail:, azure subscriptions/azure_key, webhook.method,
// logging.output and other keys that no struct field backed.
func TestConfigExample_NoPhantomKeys(t *testing.T) {
	path := filepath.Join("..", "..", "config.yaml.example")
	data, err := os.ReadFile(path)
	require.NoError(t, err, "config.yaml.example must exist at repo root")

	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)

	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		t.Fatalf("config.yaml.example contains keys that no Config field backs "+
			"(phantom/undocumented options mislead users). Fix the example or add "+
			"the field: %v", err)
	}
}

// TestConfigExample_RealSectionsPresent asserts the example actually exercises
// the sections that exist in the struct, so the file stays a useful reference
// and the header's "documents every section" claim is honest.
func TestConfigExample_RealSectionsPresent(t *testing.T) {
	path := filepath.Join("..", "..", "config.yaml.example")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var cfg Config
	require.NoError(t, yaml.Unmarshal(data, &cfg))

	require.True(t, cfg.Providers.AWS.Enabled, "aws provider should be shown enabled")
	require.NotEmpty(t, cfg.Providers.AWS.Regions, "aws regions should be present")
	require.True(t, cfg.Falco.Enabled, "falco section should be present and enabled")
	require.NotEmpty(t, cfg.DriftRules, "drift_rules should be present")
	require.NotEmpty(t, cfg.Notifications.Slack.WebhookURL, "notifications.slack should be present")
	require.Equal(t, "terraform", cfg.AutoImport.Tool, "auto_import section should be present")

	// Sections that were missing from the old example (#320): telemetry,
	// remediation, github, policy. They must at least be representable.
	require.Equal(t, "tfdrift-falco", cfg.Telemetry.ServiceName, "telemetry section should be documented")
	require.False(t, cfg.Remediation.Enabled, "remediation section should be documented")
	require.Equal(t, "main", cfg.GitHub.Branch, "github section should be documented")
	require.Equal(t, "./policies", cfg.Policy.PolicyDir, "policy section should be documented")
}
