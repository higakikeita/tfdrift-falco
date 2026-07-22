package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
)

// envelope mirrors respondJSON's wrapper ({success, data}).
type healthEnvelope struct {
	Success bool                  `json:"success"`
	Data    models.HealthResponse `json:"data"`
}

func decodeHealth(t *testing.T, h *HealthHandler) models.HealthResponse {
	t.Helper()
	w := httptest.NewRecorder()
	h.GetHealth(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("liveness must stay 200, got %d", w.Code)
	}
	var env healthEnvelope
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v\n%s", err, w.Body.String())
	}
	return env.Data
}

// #312: /health must surface a degraded signal when the real processing path
// (Falco subscription) is down, instead of always reporting "ok".
func TestHealthHandler_ReadinessDegraded(t *testing.T) {
	h := NewHealthHandlerWithReadiness("1.0.0", func() (bool, map[string]string) {
		return false, map[string]string{"falco": "disconnected"}
	})
	got := decodeHealth(t, h)
	if got.Status != "degraded" {
		t.Errorf("Status = %q, want degraded", got.Status)
	}
	if got.Checks["falco"] != "disconnected" {
		t.Errorf("Checks[falco] = %q, want disconnected", got.Checks["falco"])
	}
}

func TestHealthHandler_ReadinessOK(t *testing.T) {
	h := NewHealthHandlerWithReadiness("1.0.0", func() (bool, map[string]string) {
		return true, map[string]string{"falco": "connected"}
	})
	got := decodeHealth(t, h)
	if got.Status != "ok" {
		t.Errorf("Status = %q, want ok", got.Status)
	}
	if got.Checks["falco"] != "connected" {
		t.Errorf("Checks[falco] = %q, want connected", got.Checks["falco"])
	}
}

// Backward-compat: the probe-less constructor still reports plain ok, no checks.
func TestHealthHandler_NoProbeStillOK(t *testing.T) {
	got := decodeHealth(t, NewHealthHandler("1.0.0"))
	if got.Status != "ok" || got.Checks != nil {
		t.Errorf("probe-less health = %+v, want ok with no checks", got)
	}
}
