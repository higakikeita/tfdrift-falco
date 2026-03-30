package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// TestIntegrationHealthEndpoint tests GET /health using httptest
func TestIntegrationHealthEndpoint(t *testing.T) {
	// Set up router with health handler
	r := chi.NewRouter()
	healthHandler := NewHealthHandler("v1.0.0")
	r.Get("/health", healthHandler.GetHealth)

	// Create test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make request
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Parse response
	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	// Verify response data
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be map[string]interface{}, got %T", apiResp.Data)
	}

	if status, ok := data["status"].(string); !ok || status != "ok" {
		t.Errorf("expected status 'ok', got %v", status)
	}

	if version, ok := data["version"].(string); !ok || version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got %v", version)
	}

	if timestamp, ok := data["timestamp"].(string); !ok || timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

// TestIntegrationStatsEndpoint tests GET /api/v1/stats with real router
func TestIntegrationStatsEndpoint(t *testing.T) {
	// Set up store with test data
	store := graph.NewStore()
	addTestDrifts(store)
	addTestEvents(store)

	// Set up router
	r := chi.NewRouter()
	statsHandler := NewStatsHandler(store)
	r.Get("/api/v1/stats", statsHandler.GetStats)

	// Create test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make request
	resp, err := http.Get(ts.URL + "/api/v1/stats")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Verify status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Parse response
	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	// Verify response has stats data
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be map[string]interface{}, got %T", apiResp.Data)
	}

	// Check for expected keys
	if _, ok := data["drifts"]; !ok {
		t.Error("expected 'drifts' key in stats")
	}
	if _, ok := data["events"]; !ok {
		t.Error("expected 'events' key in stats")
	}
}

// TestIntegrationDriftsListEndpoint tests GET /api/v1/drifts with pagination
func TestIntegrationDriftsListEndpoint(t *testing.T) {
	// Set up store with test drifts
	store := graph.NewStore()
	for i := 0; i < 10; i++ {
		store.AddDrift(types.DriftAlert{
			Severity:     "high",
			ResourceType: "aws_instance",
			ResourceName: fmt.Sprintf("test-instance-%d", i),
			ResourceID:   fmt.Sprintf("i-{%08d}", i),
			Attribute:    "tags.Name",
			OldValue:     "old-name",
			NewValue:     "new-name",
			Timestamp:    time.Now().Format(time.RFC3339),
		})
	}

	// Set up router
	r := chi.NewRouter()
	driftsHandler := NewDriftsHandler(store)
	r.Get("/api/v1/drifts", driftsHandler.GetDrifts)

	// Create test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test 1: Get first page with default limit
	resp, err := http.Get(ts.URL + "/api/v1/drifts")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	// Verify paginated response structure
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be map[string]interface{}, got %T", apiResp.Data)
	}

	// Check pagination metadata
	if page, ok := data["page"].(float64); !ok || page != 1 {
		t.Errorf("expected page 1, got %v", page)
	}
	if limit, ok := data["limit"].(float64); !ok || limit != 50 {
		t.Errorf("expected limit 50, got %v", limit)
	}
	if total, ok := data["total"].(float64); !ok || total != 10 {
		t.Errorf("expected total 10, got %v", total)
	}

	// Test 2: Get drifts with custom pagination
	resp2, err := http.Get(ts.URL + "/api/v1/drifts?page=1&limit=5")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp2.Body.Close()

	var apiResp2 models.APIResponse
	if err := json.NewDecoder(resp2.Body).Decode(&apiResp2); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	data2, ok := apiResp2.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be map[string]interface{}, got %T", apiResp2.Data)
	}

	if limit, ok := data2["limit"].(float64); !ok || limit != 5 {
		t.Errorf("expected limit 5, got %v", limit)
	}

	// Verify data array is sliced correctly
	driftArray, ok := data2["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data to be []interface{}, got %T", data2["data"])
	}

	if len(driftArray) != 5 {
		t.Errorf("expected 5 drifts in response, got %d", len(driftArray))
	}
}

// TestIntegrationEventsListEndpoint tests GET /api/v1/events
func TestIntegrationEventsListEndpoint(t *testing.T) {
	// Set up store with test events
	store := graph.NewStore()
	for i := 0; i < 3; i++ {
		store.AddEvent(types.Event{
			Provider:     "aws",
			EventName:    "RunInstances",
			ResourceType: "aws_instance",
			ResourceID:   fmt.Sprintf("i-event{%08d}", i),
			UserIdentity: types.UserIdentity{
				Type:        "IAMUser",
				PrincipalID: "AIDAI123456789ABCDEF",
				UserName:    "test-user",
			},
			Changes: map[string]interface{}{
				"tag:Name": "test-instance",
			},
			Region: "us-east-1",
		})
	}

	// Set up router
	r := chi.NewRouter()
	eventsHandler := NewEventsHandler(store)
	r.Get("/api/v1/events", eventsHandler.GetEvents)

	// Create test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Make request
	resp, err := http.Get(ts.URL + "/api/v1/events")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify Content-Type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Parse response
	var apiResp models.APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !apiResp.Success {
		t.Error("expected success to be true")
	}

	// Verify paginated response
	data, ok := apiResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be map[string]interface{}, got %T", apiResp.Data)
	}

	if total, ok := data["total"].(float64); !ok || total != 3 {
		t.Errorf("expected total 3, got %v", total)
	}

	if page, ok := data["page"].(float64); !ok || page != 1 {
		t.Errorf("expected page 1, got %v", page)
	}
}

// TestIntegrationPaginationBehavior tests pagination edge cases
func TestIntegrationPaginationBehavior(t *testing.T) {
	// Set up store with exactly 25 drifts
	store := graph.NewStore()
	for i := 0; i < 25; i++ {
		store.AddDrift(types.DriftAlert{
			ResourceID: fmt.Sprintf("resource-%d", i),
		})
	}

	// Set up router
	r := chi.NewRouter()
	driftsHandler := NewDriftsHandler(store)
	r.Get("/api/v1/drifts", driftsHandler.GetDrifts)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name      string
		query     string
		wantPage  int
		wantLimit int
		wantTotal int
		wantCount int
	}{
		{"No params", "", 1, 50, 25, 25},
		{"Page 1, limit 10", "?page=1&limit=10", 1, 10, 25, 10},
		{"Page 2, limit 10", "?page=2&limit=10", 2, 10, 25, 10},
		{"Page 3, limit 10", "?page=3&limit=10", 3, 10, 25, 5},
		{"Page 4, limit 10", "?page=4&limit=10", 4, 10, 25, 0},
		{"Invalid page defaults to 1", "?page=-5&limit=10", 1, 10, 25, 10},
		{"Zero limit defaults to 50", "?page=1&limit=0", 1, 50, 25, 25},
		{"Limit > max (1000)", "?page=1&limit=2000", 1, 1000, 25, 25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + "/api/v1/drifts" + tt.query)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			var apiResp models.APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			data := apiResp.Data.(map[string]interface{})

			if page := int(data["page"].(float64)); page != tt.wantPage {
				t.Errorf("expected page %d, got %d", tt.wantPage, page)
			}
			if limit := int(data["limit"].(float64)); limit != tt.wantLimit {
				t.Errorf("expected limit %d, got %d", tt.wantLimit, limit)
			}
			if total := int(data["total"].(float64)); total != tt.wantTotal {
				t.Errorf("expected total %d, got %d", tt.wantTotal, total)
			}

			driftArray := data["data"].([]interface{})
			if len(driftArray) != tt.wantCount {
				t.Errorf("expected %d drifts in response, got %d", tt.wantCount, len(driftArray))
			}
		})
	}
}

// TestIntegrationDriftFiltering tests filtering by severity and resource type
func TestIntegrationDriftFiltering(t *testing.T) {
	// Set up store with drifts of different severities
	store := graph.NewStore()
	store.AddDrift(types.DriftAlert{
		ResourceID:   "r-1",
		Severity:     "critical",
		ResourceType: "aws_instance",
	})
	store.AddDrift(types.DriftAlert{
		ResourceID:   "r-2",
		Severity:     "high",
		ResourceType: "aws_instance",
	})
	store.AddDrift(types.DriftAlert{
		ResourceID:   "r-3",
		Severity:     "high",
		ResourceType: "aws_s3_bucket",
	})

	// Set up router
	r := chi.NewRouter()
	driftsHandler := NewDriftsHandler(store)
	r.Get("/api/v1/drifts", driftsHandler.GetDrifts)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test filtering by severity
	resp, err := http.Get(ts.URL + "/api/v1/drifts?severity=critical")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var apiResp models.APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	data := apiResp.Data.(map[string]interface{})
	total := int(data["total"].(float64))

	if total != 1 {
		t.Errorf("expected 1 critical drift, got %d", total)
	}

	// Test filtering by resource type
	resp2, err := http.Get(ts.URL + "/api/v1/drifts?resource_type=aws_s3_bucket")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp2.Body.Close()

	var apiResp2 models.APIResponse
	json.NewDecoder(resp2.Body).Decode(&apiResp2)
	data2 := apiResp2.Data.(map[string]interface{})
	total2 := int(data2["total"].(float64))

	if total2 != 1 {
		t.Errorf("expected 1 s3 bucket drift, got %d", total2)
	}
}

// TestIntegrationNotFoundEndpoint tests 404 handling
func TestIntegrationNotFoundEndpoint(t *testing.T) {
	r := chi.NewRouter()
	healthHandler := NewHealthHandler("v1.0.0")
	r.Get("/health", healthHandler.GetHealth)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/nonexistent")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

// TestIntegrationFullRouterStack tests multiple endpoints on the same router
func TestIntegrationFullRouterStack(t *testing.T) {
	// Set up store
	store := graph.NewStore()
	addTestDrifts(store)
	addTestEvents(store)

	// Set up full router with multiple handlers
	r := chi.NewRouter()
	healthHandler := NewHealthHandler("v1.0.0")
	statsHandler := NewStatsHandler(store)
	driftsHandler := NewDriftsHandler(store)
	eventsHandler := NewEventsHandler(store)

	r.Get("/health", healthHandler.GetHealth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler.GetHealth)
		r.Get("/stats", statsHandler.GetStats)
		r.Get("/drifts", driftsHandler.GetDrifts)
		r.Get("/events", eventsHandler.GetEvents)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test multiple endpoints
	tests := []struct {
		name     string
		path     string
		wantCode int
	}{
		{"Health at root", "/health", http.StatusOK},
		{"Health at v1", "/api/v1/health", http.StatusOK},
		{"Stats", "/api/v1/stats", http.StatusOK},
		{"Drifts", "/api/v1/drifts", http.StatusOK},
		{"Events", "/api/v1/events", http.StatusOK},
		{"Not found", "/api/v1/invalid", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, resp.StatusCode)
			}
		})
	}
}

// TestIntegrationResponseContentType tests all endpoints return application/json
func TestIntegrationResponseContentType(t *testing.T) {
	store := graph.NewStore()
	addTestDrifts(store)

	r := chi.NewRouter()
	healthHandler := NewHealthHandler("v1.0.0")
	statsHandler := NewStatsHandler(store)
	driftsHandler := NewDriftsHandler(store)

	r.Get("/health", healthHandler.GetHealth)
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/stats", statsHandler.GetStats)
		r.Get("/drifts", driftsHandler.GetDrifts)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	endpoints := []string{"/health", "/api/v1/stats", "/api/v1/drifts"}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			resp, err := http.Get(ts.URL + endpoint)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			contentType := resp.Header.Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", contentType)
			}
		})
	}
}

// TestIntegrationEmptyDataSet tests endpoints with no data
func TestIntegrationEmptyDataSet(t *testing.T) {
	store := graph.NewStore()
	// Store is empty

	r := chi.NewRouter()
	driftsHandler := NewDriftsHandler(store)
	eventsHandler := NewEventsHandler(store)

	r.Get("/drifts", driftsHandler.GetDrifts)
	r.Get("/events", eventsHandler.GetEvents)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name string
		path string
	}{
		{"Empty drifts", "/drifts"},
		{"Empty events", "/events"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := http.Get(ts.URL + tt.path)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
			}

			var apiResp models.APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			data := apiResp.Data.(map[string]interface{})
			if total, ok := data["total"].(float64); !ok || int(total) != 0 {
				t.Errorf("expected total 0, got %v", total)
			}
		})
	}
}

// TestIntegrationTotalPages tests total_pages calculation in paginated responses
func TestIntegrationTotalPages(t *testing.T) {
	store := graph.NewStore()
	// Add 35 drifts
	for i := 0; i < 35; i++ {
		store.AddDrift(types.DriftAlert{
			ResourceID: fmt.Sprintf("r-%d", i),
		})
	}

	r := chi.NewRouter()
	driftsHandler := NewDriftsHandler(store)
	r.Get("/drifts", driftsHandler.GetDrifts)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// With limit 10: 35 items = 4 pages (10 + 10 + 10 + 5)
	resp, err := http.Get(ts.URL + "/drifts?limit=10")
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	var apiResp models.APIResponse
	json.NewDecoder(resp.Body).Decode(&apiResp)
	data := apiResp.Data.(map[string]interface{})

	totalPages := int(data["total_pages"].(float64))
	if totalPages != 4 {
		t.Errorf("expected 4 total pages, got %d", totalPages)
	}
}

// Helper function to add test drifts
func addTestDrifts(store *graph.Store) {
	store.AddDrift(types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "test-instance",
		ResourceID:   "i-123456",
		Attribute:    "tags.Name",
		OldValue:     "old-value",
		NewValue:     "new-value",
		Timestamp:    time.Now().Format(time.RFC3339),
	})
	store.AddDrift(types.DriftAlert{
		Severity:     "medium",
		ResourceType: "aws_s3_bucket",
		ResourceName: "test-bucket",
		ResourceID:   "bucket-123",
		Attribute:    "versioning",
		OldValue:     true,
		NewValue:     false,
		Timestamp:    time.Now().Format(time.RFC3339),
	})
}

// Helper function to add test events
func addTestEvents(store *graph.Store) {
	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "RunInstances",
		ResourceType: "aws_instance",
		ResourceID:   "i-event123",
		Region:       "us-east-1",
	})
	store.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "PutBucketVersioning",
		ResourceType: "aws_s3_bucket",
		ResourceID:   "bucket-event123",
		Region:       "us-west-2",
	})
}
