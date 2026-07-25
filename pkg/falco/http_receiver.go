package falco

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/api/schema"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// maxFalcoBodyBytes caps the request body Falco's http_output may POST so a
// misconfigured or hostile sender cannot exhaust memory. A single Falco alert
// is a few KB; 1 MiB is generous.
const maxFalcoBodyBytes = 1 << 20

// falcoAlert mirrors the JSON Falco emits via http_output when json_output is
// enabled (ADR-006). output_fields is deliberately map[string]interface{}: real
// Falco emits mixed value types there — e.g. "ct.request.name": null and
// "evt.time.iso8601": <number> alongside string fields — so decoding straight
// into map[string]string rejects real alerts. coerceFields normalizes it.
type falcoAlert struct {
	Time         time.Time              `json:"time"`
	Priority     string                 `json:"priority"`
	Rule         string                 `json:"rule"`
	Output       string                 `json:"output"`
	Source       string                 `json:"source"`
	Hostname     string                 `json:"hostname"`
	Tags         []string               `json:"tags"`
	OutputFields map[string]interface{} `json:"output_fields"`
}

// coerceFields flattens Falco's mixed-type output_fields to the map[string]string
// the parsers expect: strings pass through, numbers/bools are stringified (via
// json.Number so large integers keep full precision), and nulls are dropped.
func coerceFields(raw map[string]interface{}) map[string]string {
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		switch val := v.(type) {
		case nil:
			// e.g. "ct.request.name": null — omit rather than store ""
		case string:
			out[k] = val
		case json.Number:
			out[k] = val.String()
		case bool:
			out[k] = strconv.FormatBool(val)
		default:
			out[k] = fmt.Sprintf("%v", val)
		}
	}
	return out
}

// toResponse adapts a decoded http_output alert into the gRPC outputs.Response
// the existing parsers already understand, so both transports share one parse
// path instead of duplicating the AWS/GCP/Azure field extraction (ADR-006).
func (a falcoAlert) toResponse() *outputs.Response {
	res := &outputs.Response{
		Rule:         a.Rule,
		Output:       a.Output,
		Source:       a.Source,
		Hostname:     a.Hostname,
		Tags:         a.Tags,
		OutputFields: coerceFields(a.OutputFields),
	}
	// Falco sends the priority as a name ("Warning"); the proto wants the enum.
	if v, ok := schema.Priority_value[strings.ToUpper(a.Priority)]; ok {
		res.Priority = schema.Priority(v)
	}
	if !a.Time.IsZero() {
		// timestamp.Timestamp (what outputs.Response.Time wants) is a type alias
		// for timestamppb.Timestamp, so the modern constructor assigns directly.
		res.Time = timestamppb.New(a.Time)
	}
	return res
}

// ParseHTTPAlert decodes a single Falco http_output alert body into a TFDrift
// event via the shared parse path. Returns (nil, nil) when the alert is valid
// but not drift-relevant (e.g. a read-only event), mirroring parseFalcoOutput.
// Exposed for unit testing the HTTP transport independently of a live server.
func (s *Subscriber) ParseHTTPAlert(body []byte) (*types.Event, error) {
	var a falcoAlert
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber() // keep large integer output_fields (e.g. evt.time.iso8601) exact
	if err := dec.Decode(&a); err != nil {
		return nil, fmt.Errorf("decode Falco alert: %w", err)
	}
	if a.Source == "" {
		return nil, fmt.Errorf("missing source in Falco alert")
	}
	return s.parseFalcoOutput(a.toResponse()), nil
}

// HTTPHandler returns the HTTP counterpart of the gRPC Sub stream (ADR-006):
// Falco's http_output POSTs one JSON alert per request, and each parsed drift
// event is fed onto eventCh — the same channel the gRPC path uses, so the
// downstream detector/broadcaster pipeline is unchanged.
func (s *Subscriber) HTTPHandler(eventCh chan<- types.Event) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, maxFalcoBodyBytes))
		if err != nil {
			http.Error(w, "read body", http.StatusBadRequest)
			return
		}

		event, err := s.ParseHTTPAlert(body)
		if err != nil {
			log.Warnf("Falco http_output: %v", err)
			http.Error(w, "bad alert", http.StatusBadRequest)
			return
		}

		// Receiving a well-formed alert means Falco is alive and reaching us:
		// surface it the same way the gRPC stream does so /health does not
		// report "ok" while silently receiving nothing (pus #9).
		s.connected.Store(true)

		if event != nil {
			select {
			case eventCh <- *event:
				log.Debugf("Falco http_output event queued: %s", event.EventName)
			case <-r.Context().Done():
				http.Error(w, "shutting down", http.StatusServiceUnavailable)
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
