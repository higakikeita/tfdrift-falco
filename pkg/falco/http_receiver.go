package falco

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/falcosecurity/client-go/pkg/api/schema"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// maxFalcoBodyBytes caps the request body Falco's http_output may POST so a
// misconfigured or hostile sender cannot exhaust memory. A single Falco alert
// is a few KB; 1 MiB is generous.
const maxFalcoBodyBytes = 1 << 20

// falcoAlert mirrors the JSON Falco emits via http_output when json_output is
// enabled. Only the fields the parsers consume are decoded (ADR-006).
type falcoAlert struct {
	Time         time.Time         `json:"time"`
	Priority     string            `json:"priority"`
	Rule         string            `json:"rule"`
	Output       string            `json:"output"`
	Source       string            `json:"source"`
	Hostname     string            `json:"hostname"`
	Tags         []string          `json:"tags"`
	OutputFields map[string]string `json:"output_fields"`
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
		OutputFields: a.OutputFields,
	}
	// Falco sends the priority as a name ("Warning"); the proto wants the enum.
	if v, ok := schema.Priority_value[strings.ToUpper(a.Priority)]; ok {
		res.Priority = schema.Priority(v)
	}
	if !a.Time.IsZero() {
		res.Time = &timestamp.Timestamp{
			Seconds: a.Time.Unix(),
			Nanos:   int32(a.Time.Nanosecond()),
		}
	}
	return res
}

// ParseHTTPAlert decodes a single Falco http_output alert body into a TFDrift
// event via the shared parse path. Returns (nil, nil) when the alert is valid
// but not drift-relevant (e.g. a read-only event), mirroring parseFalcoOutput.
// Exposed for unit testing the HTTP transport independently of a live server.
func (s *Subscriber) ParseHTTPAlert(body []byte) (*types.Event, error) {
	var a falcoAlert
	if err := json.Unmarshal(body, &a); err != nil {
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
