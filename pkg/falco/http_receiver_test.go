package falco

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleFalcoHTTPAlert is a real Falco http_output body (json_output: true) for
// a drift-relevant CloudTrail event — the same field set the gRPC path parses,
// so it proves the two transports share one parse path (ADR-006).
const sampleFalcoHTTPAlert = `{
  "time": "2026-07-22T14:24:41.000000000Z",
  "priority": "Warning",
  "rule": "Terraform Managed Resource Modified",
  "source": "aws_cloudtrail",
  "hostname": "falco",
  "tags": ["terraform","drift","iac"],
  "output_fields": {
    "ct.name": "ModifyInstanceAttribute",
    "ct.request.instanceid": "i-1234567890abcdef0",
    "ct.request.instancetype": "t3.medium",
    "ct.user.type": "IAMUser",
    "ct.user.principalid": "AIDAI123456789",
    "ct.user.arn": "arn:aws:iam::123456789012:user/admin",
    "ct.user.accountid": "123456789012",
    "ct.user": "admin"
  }
}`

func TestParseHTTPAlert_ValidDriftEvent(t *testing.T) {
	sub := NewSubscriberWithDefaults()

	event, err := sub.ParseHTTPAlert([]byte(sampleFalcoHTTPAlert))
	require.NoError(t, err)
	require.NotNil(t, event, "a drift-relevant alert must yield an event")

	assert.Equal(t, "i-1234567890abcdef0", event.ResourceID)
	assert.Equal(t, "ModifyInstanceAttribute", event.EventName)
}

// realFalcoAuthorizeAlert is a verbatim Falco 0.43 http_output body captured on
// real infra (#360 verification). Note output_fields mixes types: "ct.request.name"
// is null and "evt.time.iso8601" is a large integer. Decoding this straight into
// map[string]string returned HTTP 400 — this locks that regression.
const realFalcoAuthorizeAlert = `{"hostname":"cca3f82eb5b6","output":"Warning Potential Terraform drift detected","output_fields":{"ct.name":"AuthorizeSecurityGroupIngress","ct.region":"ap-northeast-1","ct.request.name":null,"ct.srcip":"221.113.81.87","ct.user":"ADMIN_OKTA_TEST","ct.user.accountid":"230446364776","evt.time.iso8601":1784730281000000000},"priority":"Warning","rule":"Terraform Managed Resource Modified","source":"aws_cloudtrail","tags":["drift","iac","terraform"],"time":"2026-07-22T14:24:41.000000000Z"}`

func TestParseHTTPAlert_RealFalcoMixedTypeFields(t *testing.T) {
	sub := NewSubscriberWithDefaults()

	// Must not error on null / numeric output_fields (the HTTP 400 regression).
	_, err := sub.ParseHTTPAlert([]byte(realFalcoAuthorizeAlert))
	require.NoError(t, err)
}

func TestCoerceFields_MixedTypes(t *testing.T) {
	got := coerceFields(map[string]interface{}{
		"str":  "keep",
		"null": nil,
		"num":  json.Number("1784730281000000000"),
		"bool": true,
	})
	assert.Equal(t, "keep", got["str"])
	assert.Equal(t, "1784730281000000000", got["num"], "large ints must keep precision")
	assert.Equal(t, "true", got["bool"])
	_, hasNull := got["null"]
	assert.False(t, hasNull, "null fields must be dropped, not stored as empty")
}

func TestParseHTTPAlert_Errors(t *testing.T) {
	sub := NewSubscriberWithDefaults()

	_, err := sub.ParseHTTPAlert([]byte(`{not json`))
	assert.Error(t, err, "malformed JSON must error")

	_, err = sub.ParseHTTPAlert([]byte(`{"rule":"x","output_fields":{}}`))
	assert.Error(t, err, "missing source must error")
}

func TestHTTPHandler_PostDeliversEventToChannel(t *testing.T) {
	sub := NewSubscriberWithDefaults()
	eventCh := make(chan types.Event, 1)

	srv := httptest.NewServer(sub.HTTPHandler(eventCh))
	defer srv.Close()

	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(sampleFalcoHTTPAlert))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	select {
	case got := <-eventCh:
		assert.Equal(t, "i-1234567890abcdef0", got.ResourceID)
	default:
		t.Fatal("expected a drift event on the channel")
	}

	// Receiving a well-formed alert marks the subscriber connected (pus #9).
	assert.True(t, sub.Connected())
}

func TestHTTPHandler_RejectsBadRequests(t *testing.T) {
	sub := NewSubscriberWithDefaults()
	eventCh := make(chan types.Event, 1)
	handler := sub.HTTPHandler(eventCh)

	// Non-POST is rejected.
	rr := httptest.NewRecorder()
	handler(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// Malformed body is a 400, not a panic.
	rr = httptest.NewRecorder()
	handler(rr, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{bad`)))
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestHTTPHandler_IrrelevantEventAcceptedButNotQueued(t *testing.T) {
	sub := NewSubscriberWithDefaults()
	eventCh := make(chan types.Event, 1)
	handler := sub.HTTPHandler(eventCh)

	// A read-only event is a valid alert but not drift — accepted, not queued.
	readOnly := `{"source":"aws_cloudtrail","output_fields":{"ct.name":"DescribeInstances"}}`
	rr := httptest.NewRecorder()
	handler(rr, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(readOnly)))
	assert.Equal(t, http.StatusAccepted, rr.Code)
	assert.Len(t, eventCh, 0, "read-only events must not be queued as drift")
}
