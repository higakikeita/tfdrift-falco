package detector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// managedInstanceState is a Terraform state with one managed aws_instance, so a
// CloudTrail change to it is a *managed-resource* drift (not "unmanaged").
const managedInstanceState = `{
  "version": 4, "terraform_version": "1.9.0", "serial": 1,
  "lineage": "00000000-0000-0000-0000-000000000000", "outputs": {},
  "resources": [
    { "mode": "managed", "type": "aws_instance", "name": "web",
      "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
      "instances": [ { "attributes": {
        "id": "i-0moneyshotdemo01", "instance_type": "t3.micro"
      } } ] }
  ]
}`

// realFalcoModifyAlert is the money-shot Falco 0.43 http_output body: a
// ModifyInstanceAttribute whose resource id lives *only* inside the aggregate
// ct.request JSON (the cloudtrail plugin 0.17.1 does not expose
// ct.request.instanceid). Extracting the id from ct.request is the #362 fix.
const realFalcoModifyAlert = `{
  "priority": "Warning",
  "rule": "Terraform Managed Resource Modified",
  "source": "aws_cloudtrail",
  "time": "2026-07-23T09:40:55Z",
  "output_fields": {
    "ct.name": "ModifyInstanceAttribute",
    "ct.region": "ap-northeast-1",
    "ct.request": "{\"instanceId\":\"i-0moneyshotdemo01\",\"instanceType\":{\"value\":\"t3.xlarge\"}}",
    "ct.user": "ADMIN_OKTA_TEST",
    "ct.user.accountid": "230446364776"
  }
}`

// falcoAlertNoResourceID has a ct.request with no id key — a valid alert that
// must NOT produce a drift (the id is genuinely unknown), proving we don't
// silently attribute drift to the wrong/empty resource.
const falcoAlertNoResourceID = `{
  "priority": "Warning", "rule": "Terraform Managed Resource Modified",
  "source": "aws_cloudtrail",
  "output_fields": { "ct.name": "ModifyInstanceAttribute", "ct.request": "{\"clientToken\":\"abc\"}" }
}`

// startDetectorWithState boots a detector over a local-state fixture with the
// HTTP Falco transport — no AWS, no Falco, no plugin — and returns the mounted
// receiver handler plus the graph store drifts land in.
func startDetectorWithState(t *testing.T, stateJSON string) (http.HandlerFunc, *graph.Store) {
	t.Helper()
	dir := t.TempDir()
	statePath := filepath.Join(dir, "terraform.tfstate")
	require.NoError(t, os.WriteFile(statePath, []byte(stateJSON), 0o600))

	cfg := &config.Config{}
	cfg.Providers.AWS.Enabled = true
	cfg.Providers.AWS.Regions = []string{"ap-northeast-1"}
	cfg.Providers.AWS.State.Backend = "local"
	cfg.Providers.AWS.State.LocalPath = statePath
	cfg.Falco.Enabled = true
	cfg.Falco.Transport = "http"

	det, err := New(cfg)
	require.NoError(t, err)

	gs := graph.NewStore()
	det.SetGraphStore(gs)

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() { _ = det.Start(ctx) }()

	// Wait until state is loaded so managed-resource lookups succeed.
	require.Eventually(t, func() bool {
		return det.GetStateManager() != nil && det.GetStateManager().ResourceCount() > 0
	}, 3*time.Second, 20*time.Millisecond, "terraform state should load")

	return det.FalcoHTTPHandler(), gs
}

func postAlert(t *testing.T, handler http.HandlerFunc, body string) int {
	t.Helper()
	rr := httptest.NewRecorder()
	handler(rr, httptest.NewRequest(http.MethodPost, "/api/v1/falco/events", strings.NewReader(body)))
	return rr.Code
}

// TestE2E_HTTPAlertToResourceAttributedDrift is the AWS-free reproduction of the
// money-shot (#365, merge gate for #363): a real Falco http_output alert whose
// id lives only in ct.request flows receiver → parser → detector and emits a
// drift attributed to the real resource id — not resource=<NA>.
func TestE2E_HTTPAlertToResourceAttributedDrift(t *testing.T) {
	handler, gs := startDetectorWithState(t, managedInstanceState)

	require.Equal(t, http.StatusAccepted, postAlert(t, handler, realFalcoModifyAlert))

	require.Eventually(t, func() bool {
		return len(gs.GetDrifts()) > 0
	}, 3*time.Second, 20*time.Millisecond, "a drift should be emitted for the managed resource")

	drifts := gs.GetDrifts()
	require.NotEmpty(t, drifts)
	assert.Equal(t, "i-0moneyshotdemo01", drifts[0].ResourceID,
		"resource id must be extracted from ct.request, not <NA>")
	assert.Equal(t, "aws_instance", drifts[0].ResourceType)
}

// TestE2E_HTTPAlertWithoutResourceIDEmitsNoDrift is the negative case: a valid
// alert whose ct.request carries no id must not fabricate a drift.
func TestE2E_HTTPAlertWithoutResourceIDEmitsNoDrift(t *testing.T) {
	handler, gs := startDetectorWithState(t, managedInstanceState)

	require.Equal(t, http.StatusAccepted, postAlert(t, handler, falcoAlertNoResourceID))

	// Give the processor time to (not) produce a drift.
	time.Sleep(500 * time.Millisecond)
	assert.Empty(t, gs.GetDrifts(), "no resource id → no drift (not a silent misattribution)")
}
