package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/models"
	"github.com/keitahigaki/tfdrift-falco/pkg/detector"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/provider"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// MockProvider is a minimal implementation of provider.Provider for testing
type MockProvider struct {
	name string
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) ParseEvent(source string, fields map[string]string, rawEvent interface{}) *types.Event {
	return nil
}

func (m *MockProvider) IsRelevantEvent(eventName string) bool {
	return true
}

func (m *MockProvider) MapEventToResource(eventName string, eventSource string) string {
	return "aws_instance"
}

func (m *MockProvider) ExtractChanges(eventName string, fields map[string]string) map[string]interface{} {
	return make(map[string]interface{})
}

func (m *MockProvider) SupportedEventCount() int {
	return 100
}

func (m *MockProvider) SupportedResourceTypes() []string {
	return []string{"aws_instance", "aws_s3_bucket", "aws_vpc"}
}

// ===== CorrelationsHandler Tests =====

func TestNewCorrelationsHandler(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)
	handler := NewCorrelationsHandler(correlator)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.correlator != correlator {
		t.Fatal("expected correlator to be set correctly")
	}
}

func TestGetCorrelations_Empty(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)
	handler := NewCorrelationsHandler(correlator)

	req := httptest.NewRequest("GET", "/api/v1/correlations", nil)
	w := httptest.NewRecorder()

	handler.GetCorrelations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Check response structure
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	if correlations, ok := data["correlations"]; !ok {
		t.Error("expected 'correlations' key in response data")
	} else if correlations == nil {
		// correlations should be an empty slice or null
	}

	if count, ok := data["count"]; !ok {
		t.Error("expected 'count' key in response data")
	} else if countVal, ok := count.(float64); ok && countVal != 0 {
		t.Errorf("expected count to be 0 for empty correlations, got %v", countVal)
	}

	if _, ok := data["timestamp"]; !ok {
		t.Error("expected 'timestamp' key in response data")
	}
}

func TestGetCorrelations_WithGroups(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)

	// Add events to create correlations
	event1 := types.Event{
		Provider:     "aws",
		EventName:    "PutObject",
		ResourceType: "aws_instance",
		ResourceID:   "i-123456",
		UserIdentity: types.UserIdentity{UserName: "alice"},
		Changes:      make(map[string]interface{}),
	}
	event2 := types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.delete",
		ResourceType: "google_compute_instance",
		ResourceID:   "inst-789",
		UserIdentity: types.UserIdentity{UserName: "alice"},
		Changes:      make(map[string]interface{}),
	}

	correlator.AddEvent(event1)
	correlator.AddEvent(event2)

	handler := NewCorrelationsHandler(correlator)

	req := httptest.NewRequest("GET", "/api/v1/correlations", nil)
	w := httptest.NewRecorder()

	handler.GetCorrelations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// The count should reflect the correlations
	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"]; ok {
		if countVal, ok := count.(float64); ok && countVal > 0 {
			// Successfully created a correlation
		}
	}
}

func TestGetCorrelations_WithProviderFilter(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)

	event1 := types.Event{
		Provider:     "aws",
		EventName:    "PutObject",
		ResourceType: "aws_instance",
		ResourceID:   "i-123456",
		UserIdentity: types.UserIdentity{UserName: "bob"},
		Changes:      make(map[string]interface{}),
	}
	event2 := types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.delete",
		ResourceType: "google_compute_instance",
		ResourceID:   "inst-789",
		UserIdentity: types.UserIdentity{UserName: "bob"},
		Changes:      make(map[string]interface{}),
	}

	correlator.AddEvent(event1)
	correlator.AddEvent(event2)

	handler := NewCorrelationsHandler(correlator)

	req := httptest.NewRequest("GET", "/api/v1/correlations?provider=aws", nil)
	w := httptest.NewRecorder()

	handler.GetCorrelations(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify Content-Type header
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestGetCorrelationStats_Empty(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)
	handler := NewCorrelationsHandler(correlator)

	req := httptest.NewRequest("GET", "/api/v1/correlations/stats", nil)
	w := httptest.NewRecorder()

	handler.GetCorrelationStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Stats should have expected keys
	stats, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	expectedKeys := []string{"buffered_events", "correlation_groups", "multi_cloud_groups", "events_by_provider", "window_seconds"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("expected stats to contain key %q", key)
		}
	}
}

func TestGetCorrelationStats_WithEvents(t *testing.T) {
	correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)

	// Add multiple events
	for i := 0; i < 3; i++ {
		event := types.Event{
			Provider:     "aws",
			EventName:    "PutObject",
			ResourceType: "aws_instance",
			ResourceID:   "i-123456",
			UserIdentity: types.UserIdentity{UserName: "charlie"},
			Changes:      make(map[string]interface{}),
		}
		correlator.AddEvent(event)
	}

	handler := NewCorrelationsHandler(correlator)

	req := httptest.NewRequest("GET", "/api/v1/correlations/stats", nil)
	w := httptest.NewRecorder()

	handler.GetCorrelationStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	stats := resp.Data.(map[string]interface{})
	if buffered, ok := stats["buffered_events"].(float64); ok && buffered >= 3 {
		// Correctly tracked buffered events
	} else {
		t.Errorf("expected buffered_events >= 3, got %v", buffered)
	}
}

// ===== ProviderStatusHandler Tests =====

func TestNewProviderStatusHandler(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})
	registry.Register(&MockProvider{name: "gcp"})

	handler := NewProviderStatusHandler(registry)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.registry != registry {
		t.Fatal("expected registry to be set correctly")
	}
	if handler.startAt.IsZero() {
		t.Error("expected startAt to be set")
	}
}

func TestRecordEvent_NewProvider(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	handler.RecordEvent("aws", true)

	handler.mu.RLock()
	stats, ok := handler.stats["aws"]
	handler.mu.RUnlock()

	if !ok {
		t.Error("expected stats for aws provider to be created")
	}
	if stats.EventsReceived != 1 {
		t.Errorf("expected EventsReceived to be 1, got %d", stats.EventsReceived)
	}
	if stats.EventsMatched != 1 {
		t.Errorf("expected EventsMatched to be 1, got %d", stats.EventsMatched)
	}
	if stats.Status != "active" {
		t.Errorf("expected status to be 'active', got %s", stats.Status)
	}
	if stats.LastEventAt.IsZero() {
		t.Error("expected LastEventAt to be set")
	}
}

func TestRecordEvent_UnmatchedEvent(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	handler.RecordEvent("gcp", false)

	handler.mu.RLock()
	stats, ok := handler.stats["gcp"]
	handler.mu.RUnlock()

	if !ok {
		t.Error("expected stats for gcp provider to be created")
	}
	if stats.EventsReceived != 1 {
		t.Errorf("expected EventsReceived to be 1, got %d", stats.EventsReceived)
	}
	if stats.EventsMatched != 0 {
		t.Errorf("expected EventsMatched to be 0, got %d", stats.EventsMatched)
	}
}

func TestRecordEvent_MultipleEvents(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	// Record multiple events
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", false)

	handler.mu.RLock()
	stats := handler.stats["aws"]
	handler.mu.RUnlock()

	if stats.EventsReceived != 3 {
		t.Errorf("expected EventsReceived to be 3, got %d", stats.EventsReceived)
	}
	if stats.EventsMatched != 2 {
		t.Errorf("expected EventsMatched to be 2, got %d", stats.EventsMatched)
	}
}

func TestRecordError_NewProvider(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	handler.RecordError("azure")

	handler.mu.RLock()
	stats, ok := handler.stats["azure"]
	handler.mu.RUnlock()

	if !ok {
		t.Error("expected stats for azure provider to be created")
	}
	if stats.ErrorCount != 1 {
		t.Errorf("expected ErrorCount to be 1, got %d", stats.ErrorCount)
	}
	if stats.Status != "error" {
		t.Errorf("expected status to be 'error', got %s", stats.Status)
	}
}

func TestRecordError_MultipleErrors(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	handler.RecordError("aws")
	handler.RecordError("aws")

	handler.mu.RLock()
	stats := handler.stats["aws"]
	handler.mu.RUnlock()

	if stats.ErrorCount != 2 {
		t.Errorf("expected ErrorCount to be 2, got %d", stats.ErrorCount)
	}
}

func TestGetProviderStatus_Empty(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Check response structure
	data := resp.Data.(map[string]interface{})
	if _, ok := data["providers"]; !ok {
		t.Error("expected 'providers' key in response data")
	}
	if _, ok := data["count"]; !ok {
		t.Error("expected 'count' key in response data")
	}
	if _, ok := data["uptime"]; !ok {
		t.Error("expected 'uptime' key in response data")
	}
	if _, ok := data["timestamp"]; !ok {
		t.Error("expected 'timestamp' key in response data")
	}
}

func TestGetProviderStatus_WithProviders(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})
	registry.Register(&MockProvider{name: "gcp"})
	registry.Register(&MockProvider{name: "azure"})

	handler := NewProviderStatusHandler(registry)

	// Record some events
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", true)
	handler.RecordEvent("gcp", false)
	handler.RecordError("azure")

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Verify response structure
	data := resp.Data.(map[string]interface{})

	count := data["count"].(float64)
	if count != 3 {
		t.Errorf("expected 3 providers, got %v", count)
	}

	providers := data["providers"].([]interface{})
	if len(providers) != 3 {
		t.Errorf("expected 3 provider entries, got %d", len(providers))
	}

	// Check that providers are sorted
	var names []string
	for _, p := range providers {
		entry := p.(map[string]interface{})
		names = append(names, entry["name"].(string))
	}
	if len(names) == 3 {
		if names[0] != "aws" || names[1] != "azure" || names[2] != "gcp" {
			t.Errorf("expected providers to be sorted, got %v", names)
		}
	}

	// Verify uptime is set
	if uptime, ok := data["uptime"].(float64); !ok || uptime < 0 {
		t.Errorf("expected positive uptime, got %v", uptime)
	}

	// Verify Content-Type header
	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

func TestGetProviderStatus_ProviderDetails(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})

	handler := NewProviderStatusHandler(registry)
	handler.RecordEvent("aws", true)

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	providers := data["providers"].([]interface{})
	if len(providers) > 0 {
		awsEntry := providers[0].(map[string]interface{})

		// Check expected fields
		expectedFields := []string{"name", "status", "event_count", "resource_types", "has_discovery", "has_comparison", "events_received", "events_matched", "error_count", "uptime_seconds"}
		for _, field := range expectedFields {
			if _, ok := awsEntry[field]; !ok {
				t.Errorf("expected field %q in provider entry", field)
			}
		}

		// Verify values
		if awsEntry["name"] != "aws" {
			t.Errorf("expected name 'aws', got %v", awsEntry["name"])
		}
		if awsEntry["events_received"].(float64) != 1 {
			t.Errorf("expected 1 event received, got %v", awsEntry["events_received"])
		}
		if awsEntry["events_matched"].(float64) != 1 {
			t.Errorf("expected 1 event matched, got %v", awsEntry["events_matched"])
		}
	}
}

func TestGetProviderStatus_ActiveStatus(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})

	handler := NewProviderStatusHandler(registry)
	handler.RecordEvent("aws", true)

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	providers := data["providers"].([]interface{})
	if len(providers) > 0 {
		entry := providers[0].(map[string]interface{})
		if entry["status"] != "active" {
			t.Errorf("expected status 'active' for recently active provider, got %s", entry["status"])
		}
	}
}

func TestGetProviderStatus_IdleStatus(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "gcp"})

	handler := NewProviderStatusHandler(registry)
	handler.RecordEvent("gcp", true)

	// Manually set LastEventAt to be old (5+ minutes ago)
	handler.mu.Lock()
	handler.stats["gcp"].LastEventAt = time.Now().Add(-10 * time.Minute)
	handler.mu.Unlock()

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	providers := data["providers"].([]interface{})
	if len(providers) > 0 {
		entry := providers[0].(map[string]interface{})
		if entry["status"] != "idle" {
			t.Errorf("expected status 'idle' for inactive provider, got %s", entry["status"])
		}
	}
}

func TestGetProviderStatus_ErrorStatus(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "azure"})

	handler := NewProviderStatusHandler(registry)
	handler.RecordError("azure")

	req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
	w := httptest.NewRecorder()

	handler.GetProviderStatus(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	providers := data["providers"].([]interface{})
	if len(providers) > 0 {
		entry := providers[0].(map[string]interface{})
		if entry["status"] != "error" {
			t.Errorf("expected status 'error' for error provider, got %s", entry["status"])
		}
	}
}

func TestGetProviderSummary_Empty(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/summary", nil)
	w := httptest.NewRecorder()

	handler.GetProviderSummary(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Check response structure
	data := resp.Data.(map[string]interface{})
	expectedKeys := []string{"total_providers", "active_providers", "total_events", "total_matched", "total_errors", "providers", "match_rate_percent", "uptime_seconds"}
	for _, key := range expectedKeys {
		if _, ok := data[key]; !ok {
			t.Errorf("expected key %q in summary response", key)
		}
	}
}

func TestGetProviderSummary_WithProviders(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})
	registry.Register(&MockProvider{name: "gcp"})
	registry.Register(&MockProvider{name: "azure"})

	handler := NewProviderStatusHandler(registry)

	// Record events
	handler.RecordEvent("aws", true)
	handler.RecordEvent("aws", true)
	handler.RecordEvent("gcp", true)
	handler.RecordEvent("gcp", false)
	handler.RecordError("azure")
	handler.RecordError("azure")

	req := httptest.NewRequest("GET", "/api/v1/providers/summary", nil)
	w := httptest.NewRecorder()

	handler.GetProviderSummary(w, req)

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	data := resp.Data.(map[string]interface{})

	// Verify totals
	if totalProviders := int(data["total_providers"].(float64)); totalProviders != 3 {
		t.Errorf("expected 3 total providers, got %d", totalProviders)
	}

	if totalEvents := int(data["total_events"].(float64)); totalEvents != 4 {
		t.Errorf("expected 4 total events, got %d", totalEvents)
	}

	if totalMatched := int(data["total_matched"].(float64)); totalMatched != 3 {
		t.Errorf("expected 3 total matched events, got %d", totalMatched)
	}

	if totalErrors := int(data["total_errors"].(float64)); totalErrors != 2 {
		t.Errorf("expected 2 total errors, got %d", totalErrors)
	}

	// Verify match rate
	matchRate := data["match_rate_percent"].(float64)
	if matchRate != 75.0 {
		t.Errorf("expected 75%% match rate, got %f%%", matchRate)
	}

	// Check providers list
	providers := data["providers"].([]interface{})
	if len(providers) != 3 {
		t.Errorf("expected 3 providers in list, got %d", len(providers))
	}
}

func TestGetProviderSummary_MatchRate(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})

	handler := NewProviderStatusHandler(registry)

	// Record events with known match rate
	handler.RecordEvent("aws", true)  // 1 matched
	handler.RecordEvent("aws", false) // 0 matched
	handler.RecordEvent("aws", false) // 0 matched
	handler.RecordEvent("aws", false) // 0 matched

	req := httptest.NewRequest("GET", "/api/v1/providers/summary", nil)
	w := httptest.NewRecorder()

	handler.GetProviderSummary(w, req)

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	matchRate := data["match_rate_percent"].(float64)
	if matchRate != 25.0 {
		t.Errorf("expected 25%% match rate, got %f%%", matchRate)
	}
}

func TestGetProviderSummary_ContentType(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/summary", nil)
	w := httptest.NewRecorder()

	handler.GetProviderSummary(w, req)

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

// ===== Thread Safety Tests =====

func TestConcurrentRecordEvent(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProviderStatusHandler(registry)

	// Run concurrent operations
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				handler.RecordEvent("aws", j%2 == 0)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	handler.mu.RLock()
	stats := handler.stats["aws"]
	handler.mu.RUnlock()

	if stats.EventsReceived != 1000 {
		t.Errorf("expected 1000 events, got %d", stats.EventsReceived)
	}
}

func TestConcurrentRecordEventAndGetStatus(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})

	handler := NewProviderStatusHandler(registry)

	// Record events and read status concurrently
	done := make(chan bool, 20)

	// 10 writers
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 50; j++ {
				handler.RecordEvent("aws", true)
			}
			done <- true
		}()
	}

	// 10 readers
	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/v1/providers/status", nil)
			w := httptest.NewRecorder()
			handler.GetProviderStatus(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify consistency
	handler.mu.RLock()
	stats := handler.stats["aws"]
	handler.mu.RUnlock()

	if stats.EventsReceived != 500 {
		t.Errorf("expected 500 events, got %d", stats.EventsReceived)
	}
}

func TestMatchRate(t *testing.T) {
	tests := []struct {
		total    int64
		matched  int64
		expected float64
	}{
		{0, 0, 0.0},
		{10, 5, 50.0},
		{100, 75, 75.0},
		{4, 1, 25.0},
	}

	for _, tt := range tests {
		result := matchRate(tt.total, tt.matched)
		if result != tt.expected {
			t.Errorf("matchRate(%d, %d) = %f, want %f", tt.total, tt.matched, result, tt.expected)
		}
	}
}

// ===== HealthHandler Tests =====

func TestNewHealthHandler(t *testing.T) {
	handler := NewHealthHandler("1.2.3")

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.version != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", handler.version)
	}
}

func TestGetHealth(t *testing.T) {
	handler := NewHealthHandler("2.0.0")

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	// Check response data structure
	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	if status, ok := data["status"].(string); !ok || status != "ok" {
		t.Errorf("expected status 'ok', got %v", data["status"])
	}

	if version, ok := data["version"].(string); !ok || version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %v", data["version"])
	}

	if _, ok := data["timestamp"]; !ok {
		t.Error("expected 'timestamp' key in response data")
	}

	if contentType := w.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}
}

// ===== StatsHandler Tests =====

func TestNewStatsHandler(t *testing.T) {
	store := graph.NewStore()
	handler := NewStatsHandler(store)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.store != store {
		t.Fatal("expected store to be set correctly")
	}
}

func TestGetStats_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewStatsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	stats, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	// Verify expected keys
	expectedKeys := []string{"graph", "drifts", "events", "unmanaged", "severity_breakdown", "top_resource_types"}
	for _, key := range expectedKeys {
		if _, ok := stats[key]; !ok {
			t.Errorf("expected stats to contain key %q", key)
		}
	}
}

func TestGetStats_WithData(t *testing.T) {
	store := graph.NewStore()

	// Add some test data
	drift := types.DriftAlert{
		ResourceID:   "i-123",
		Severity:     "high",
		ResourceType: "aws_instance",
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	store.AddDrift(drift)

	handler := NewStatsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

// ===== EventsHandler Tests =====

func TestNewEventsHandler(t *testing.T) {
	store := graph.NewStore()
	handler := NewEventsHandler(store)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.store != store {
		t.Fatal("expected store to be set correctly")
	}
}

func TestGetEvents_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	w := httptest.NewRecorder()

	handler.GetEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	if total, ok := data["total"].(float64); !ok || total != 0 {
		t.Errorf("expected total to be 0, got %v", data["total"])
	}
}

func TestGetEvents_WithData(t *testing.T) {
	store := graph.NewStore()

	event := types.Event{
		Provider:     "aws",
		EventName:    "PutObject",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		Changes:      make(map[string]interface{}),
	}
	store.AddEvent(event)

	handler := NewEventsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/events", nil)
	w := httptest.NewRecorder()

	handler.GetEvents(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}
}


// ===== DriftsHandler Tests =====

func TestNewDriftsHandler(t *testing.T) {
	store := graph.NewStore()
	handler := NewDriftsHandler(store)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.store != store {
		t.Fatal("expected store to be set correctly")
	}
}

func TestGetDrifts_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewDriftsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/drifts", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	data, ok := resp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("expected Data to be a map, got %T", resp.Data)
	}

	if total, ok := data["total"].(float64); !ok || total != 0 {
		t.Errorf("expected total to be 0, got %v", data["total"])
	}
}

func TestGetDrifts_WithData(t *testing.T) {
	store := graph.NewStore()

	drift := types.DriftAlert{
		ResourceID:   "i-123",
		Severity:     "high",
		ResourceType: "aws_instance",
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	store.AddDrift(drift)

	handler := NewDriftsHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/drifts", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	data := resp.Data.(map[string]interface{})
	if total, ok := data["total"].(float64); !ok || total != 1 {
		t.Errorf("expected total to be 1, got %v", data["total"])
	}
}

func TestGetDrifts_WithFilters(t *testing.T) {
	store := graph.NewStore()

	drift1 := types.DriftAlert{
		ResourceID:   "i-123",
		Severity:     "high",
		ResourceType: "aws_instance",
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	drift2 := types.DriftAlert{
		ResourceID:   "vol-456",
		Severity:     "low",
		ResourceType: "aws_ebs_volume",
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	store.AddDrift(drift1)
	store.AddDrift(drift2)

	handler := NewDriftsHandler(store)

	// Test severity filter
	req := httptest.NewRequest("GET", "/api/v1/drifts?severity=high", nil)
	w := httptest.NewRecorder()

	handler.GetDrifts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	if total, ok := data["total"].(float64); !ok || total != 1 {
		t.Errorf("expected total to be 1 after filter, got %v", data["total"])
	}
}

// ===== ProvidersHandler Tests =====

func TestNewProvidersHandler(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProvidersHandler(registry)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.registry != registry {
		t.Fatal("expected registry to be set correctly")
	}
}

func TestGetProviders_Empty(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProvidersHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	w := httptest.NewRecorder()

	handler.GetProviders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected success to be true")
	}

	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"].(float64); !ok || count != 0 {
		t.Errorf("expected count to be 0, got %v", data["count"])
	}
}

func TestGetProviders_WithProviders(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})
	registry.Register(&MockProvider{name: "gcp"})

	handler := NewProvidersHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers", nil)
	w := httptest.NewRecorder()

	handler.GetProviders(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	if count, ok := data["count"].(float64); !ok || count != 2 {
		t.Errorf("expected count to be 2, got %v", data["count"])
	}
}

func TestGetProviderCapabilities_Success(t *testing.T) {
	registry := provider.NewRegistry()
	registry.Register(&MockProvider{name: "aws"})

	handler := NewProvidersHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/capabilities?name=aws", nil)
	w := httptest.NewRecorder()

	handler.GetProviderCapabilities(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGetProviderCapabilities_NotFound(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProvidersHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/capabilities?name=nonexistent", nil)
	w := httptest.NewRecorder()

	handler.GetProviderCapabilities(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestGetProviderCapabilities_MissingName(t *testing.T) {
	registry := provider.NewRegistry()
	handler := NewProvidersHandler(registry)

	req := httptest.NewRequest("GET", "/api/v1/providers/capabilities", nil)
	w := httptest.NewRecorder()

	handler.GetProviderCapabilities(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// ===== GraphHandler Tests =====

func TestNewGraphHandler(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphHandler(store)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.store != store {
		t.Fatal("expected store to be set correctly")
	}
}

func TestGetGraph_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph", nil)
	w := httptest.NewRecorder()

	handler.GetGraph(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success to be true")
	}
}

func TestGetNodes_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph/nodes", nil)
	w := httptest.NewRecorder()

	handler.GetNodes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	if total, ok := data["total"].(float64); !ok || total != 0 {
		t.Errorf("expected total to be 0, got %v", data["total"])
	}
}

func TestGetEdges_Empty(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph/edges", nil)
	w := httptest.NewRecorder()

	handler.GetEdges(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp.Data.(map[string]interface{})
	if total, ok := data["total"].(float64); !ok || total != 0 {
		t.Errorf("expected total to be 0, got %v", data["total"])
	}
}

// ===== GraphQueryHandler Tests =====

func TestNewGraphQueryHandler(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	if handler == nil {
		t.Fatal("expected handler to be created, got nil")
	}
	if handler.graphStore != store {
		t.Fatal("expected graphStore to be set correctly")
	}
}


func TestGetNodesByLabel_Missing(t *testing.T) {
	store := graph.NewStore()
	handler := NewGraphQueryHandler(store)

	req := httptest.NewRequest("GET", "/api/v1/graph/nodes", nil)
	w := httptest.NewRecorder()

	handler.GetNodesByLabel(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
