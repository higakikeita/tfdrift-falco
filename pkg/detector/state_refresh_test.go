package detector

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const stateOneResource = `{
  "version": 4, "terraform_version": "1.9.0", "serial": 1, "lineage": "l", "outputs": {},
  "resources": [
    { "mode": "managed", "type": "aws_instance", "name": "a",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [ { "attributes": { "id": "i-aaa", "instance_type": "t3.micro" } } ] }
  ]
}`

const stateTwoResources = `{
  "version": 4, "terraform_version": "1.9.0", "serial": 2, "lineage": "l", "outputs": {},
  "resources": [
    { "mode": "managed", "type": "aws_instance", "name": "a",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [ { "attributes": { "id": "i-aaa", "instance_type": "t3.micro" } } ] },
    { "mode": "managed", "type": "aws_instance", "name": "b",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [ { "attributes": { "id": "i-bbb", "instance_type": "t3.small" } } ] }
  ]
}`

// TestRefreshAllState_PicksUpAppliedChanges proves the fix for #331: after a
// legitimate `terraform apply` adds a resource to the state file, a running
// detector that refreshes sees it — instead of comparing against the startup
// snapshot forever (and flagging the new resource as unmanaged drift).
func TestRefreshAllState_PicksUpAppliedChanges(t *testing.T) {
	dir := t.TempDir()
	statePath := filepath.Join(dir, "terraform.tfstate")
	require.NoError(t, os.WriteFile(statePath, []byte(stateOneResource), 0o600))

	cfg := &config.Config{}
	cfg.Providers.AWS.Enabled = true
	cfg.Providers.AWS.Regions = []string{"ap-northeast-1"}
	cfg.Providers.AWS.State.Backend = "local"
	cfg.Providers.AWS.State.LocalPath = statePath
	cfg.Falco.Enabled = false

	det, err := New(cfg)
	require.NoError(t, err)

	ctx := context.Background()
	require.NoError(t, det.GetStateManager().Load(ctx))
	require.Equal(t, 1, det.GetStateManager().ResourceCount())

	// Simulate a `terraform apply` that adds a second managed resource.
	require.NoError(t, os.WriteFile(statePath, []byte(stateTwoResources), 0o600))

	// Without a refresh the detector would still see only 1 resource.
	det.refreshAllState(ctx)

	assert.Equal(t, 2, det.GetStateManager().ResourceCount(),
		"refresh must re-read state so applied changes are visible (#331)")
	_, exists := det.GetStateManager().GetResource("i-bbb")
	assert.True(t, exists, "the newly applied resource must be known after refresh")
}
