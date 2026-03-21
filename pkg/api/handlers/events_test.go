package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupEventsTestStore() *graph.Store {
	store := graph.NewStore()
	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "ModifySecurityGroupIngress",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-001",
		Timestamp:    "2026-03-21T10:00:00Z",
		Severity:     "critical",
		Status:       types.EventStatusOpen,
		Region:       "us-east-1",
		UserIdentity: types.UserIdentity{UserName: "admin"},
	})
	store.AddEvent(types.Event{
		Provider:     "gcp",
		EventName:    "compute.firewalls.patch",
		ResourceType: "google_compute_firewall",
		ResourceID:   "fw-002",
		Timestamp:    "2026-03-21T11:00:00Z",
		Severity:     "high",
		Status:       types.EventStatusOpen,
		Region:       "us-central1",
		UserIdentity: types.UserIdentity{UserName: "developer"},
	})
	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "PutBucketTagging",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "my-bucket",
		Timestamp:    "2026-03-21T12:00:00Z",
		Severity:     "low",
		Status:       types.EventStatusAcknowledged,
		Region:       "eu-west-1",
		UserIdentity: types.UserIdentity{UserName: "ci-bot"},
	})
	return store
}

func TestGetEvents_NoFilters(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)

	// Check paginated response
	data, ok := resp.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(3), data["total"])
}

func TestGetEvents_FilterByProvider(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events?provider=aws", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(2), data["total"])
}

func TestGetEvents_FilterBySeverity(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events?severity=critical", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}

func TestGetEvents_FilterByStatus(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events?status=acknowledged", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}

func TestGetEvents_Search(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events?search=admin", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"])
}

func TestGetEvents_TimeRange(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events?from=2026-03-21T10:30:00Z&to=2026-03-21T11:30:00Z", nil)
	rr := httptest.NewRecorder()
	handler.GetEvents(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	data := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["total"]) // only gcp event at 11:00
}

func TestGetEvent_Found(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Get("/api/v1/events/{id}", handler.GetEvent)

	req := httptest.NewRequest("GET", "/api/v1/events/sg-001", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.True(t, resp.Success)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "sg-001", data["resource_id"])
	assert.Equal(t, "aws", data["provider"])
}

func TestGetEvent_NotFound(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Get("/api/v1/events/{id}", handler.GetEvent)

	req := httptest.NewRequest("GET", "/api/v1/events/nonexistent", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestUpdateEventStatus_Acknowledge(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Patch("/api/v1/events/{id}", handler.UpdateEventStatus)

	body := `{"status": "acknowledged"}`
	req := httptest.NewRequest("PATCH", "/api/v1/events/sg-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp models.APIResponse
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.True(t, resp.Success)

	data := resp.Data.(map[string]interface{})
	assert.Equal(t, "acknowledged", data["status"])
}

func TestUpdateEventStatus_IgnoreRequiresReason(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Patch("/api/v1/events/{id}", handler.UpdateEventStatus)

	// Without reason - should fail
	body := `{"status": "ignored"}`
	req := httptest.NewRequest("PATCH", "/api/v1/events/sg-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// With reason - should succeed
	body = `{"status": "ignored", "reason": "Known test environment change"}`
	req = httptest.NewRequest("PATCH", "/api/v1/events/sg-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestUpdateEventStatus_InvalidStatus(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Patch("/api/v1/events/{id}", handler.UpdateEventStatus)

	body := `{"status": "invalid_status"}`
	req := httptest.NewRequest("PATCH", "/api/v1/events/sg-001", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestUpdateEventStatus_NotFound(t *testing.T) {
	store := setupEventsTestStore()
	handler := NewEventsHandler(store)

	r := chi.NewRouter()
	r.Patch("/api/v1/events/{id}", handler.UpdateEventStatus)

	body := `{"status": "acknowledged"}`
	req := httptest.NewRequest("PATCH", "/api/v1/events/nonexistent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
