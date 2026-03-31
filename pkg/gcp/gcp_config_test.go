package gcp

import (
	"testing"
)

// TestLoadResourceConfig tests loading the resource config
func TestLoadResourceConfig(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadResourceConfig returned nil")
	}

	if len(cfg.RelevantEvents) == 0 {
		t.Error("RelevantEvents should not be empty")
	}

	if len(cfg.EventToResourceType) == 0 {
		t.Error("EventToResourceType should not be empty")
	}

	if len(cfg.relevantEventsMap) == 0 {
		t.Error("relevantEventsMap should not be empty")
	}
}

// TestLoadResourceConfigIdempotent tests that LoadResourceConfig is idempotent
func TestLoadResourceConfigIdempotent(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg1, err1 := LoadResourceConfig()
	if err1 != nil {
		t.Fatalf("First LoadResourceConfig failed: %v", err1)
	}

	cfg2, err2 := LoadResourceConfig()
	if err2 != nil {
		t.Fatalf("Second LoadResourceConfig failed: %v", err2)
	}

	// Should return the same instance
	if cfg1 != cfg2 {
		t.Error("LoadResourceConfig should return the same instance on subsequent calls")
	}
}

// TestIsRelevantEvent tests the IsRelevantEvent method
func TestIsRelevantEvent(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	// Test a known relevant event
	if !cfg.IsRelevantEvent("compute.instances.insert") {
		t.Error("compute.instances.insert should be relevant")
	}

	// Test a non-existent event
	if cfg.IsRelevantEvent("nonexistent.event.method") {
		t.Error("nonexistent.event.method should not be relevant")
	}
}

// TestGetRelevantEventsMap tests the GetRelevantEventsMap method
func TestGetRelevantEventsMap(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	eventsMap := cfg.GetRelevantEventsMap()
	if len(eventsMap) == 0 {
		t.Error("GetRelevantEventsMap should return non-empty map")
	}

	// Check if a known event is in the map
	if !eventsMap["compute.instances.insert"] {
		t.Error("compute.instances.insert should be in the events map")
	}
}

// TestGetEventToResourceTypeMap tests the GetEventToResourceTypeMap method
func TestGetEventToResourceTypeMap(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	resourceMap := cfg.GetEventToResourceTypeMap()
	if len(resourceMap) == 0 {
		t.Error("GetEventToResourceTypeMap should return non-empty map")
	}

	// Check if a known mapping exists
	if resourceType, ok := resourceMap["compute.instances.insert"]; !ok || resourceType != "google_compute_instance" {
		t.Errorf("Expected compute.instances.insert to map to google_compute_instance, got %s", resourceType)
	}
}

// TestGetResourceType tests the GetResourceType method
func TestGetResourceType(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	tests := []struct {
		eventName        string
		expectedResource string
	}{
		{"compute.instances.insert", "google_compute_instance"},
		{"compute.firewalls.insert", "google_compute_firewall"},
		{"storage.buckets.create", "google_storage_bucket"},
		{"nonexistent.event", ""},
	}

	for _, tt := range tests {
		t.Run(tt.eventName, func(t *testing.T) {
			resourceType := cfg.GetResourceType(tt.eventName)
			if resourceType != tt.expectedResource {
				t.Errorf("GetResourceType(%s) = %s, expected %s", tt.eventName, resourceType, tt.expectedResource)
			}
		})
	}
}

// TestResourceConfigRelevantEventsContent tests that relevant events are properly loaded
func TestResourceConfigRelevantEventsContent(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	expectedEvents := []string{
		"compute.instances.insert",
		"compute.instances.delete",
		"compute.firewalls.insert",
		"storage.buckets.create",
		"SetIamPolicy",
	}

	for _, expectedEvent := range expectedEvents {
		found := false
		for _, event := range cfg.RelevantEvents {
			if event == expectedEvent {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected event %s not found in RelevantEvents", expectedEvent)
		}
	}
}

// TestIsRelevantEventConsistency tests consistency between RelevantEvents and IsRelevantEvent
func TestIsRelevantEventConsistency(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	// Every event in RelevantEvents should return true from IsRelevantEvent
	for _, event := range cfg.RelevantEvents {
		if !cfg.IsRelevantEvent(event) {
			t.Errorf("IsRelevantEvent returned false for event in RelevantEvents: %s", event)
		}
	}
}

// TestEventToResourceTypeCount tests that we have expected number of mappings
func TestEventToResourceTypeCount(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	// Should have a significant number of mappings (200+)
	if len(cfg.EventToResourceType) < 100 {
		t.Errorf("Expected at least 100 event mappings, got %d", len(cfg.EventToResourceType))
	}
}

// TestRelevantEventsCount tests that we have expected number of relevant events
func TestRelevantEventsCount(t *testing.T) {
	// Reset global config for test
	globalConfig = nil

	cfg, err := LoadResourceConfig()
	if err != nil {
		t.Fatalf("LoadResourceConfig failed: %v", err)
	}

	// Should have a significant number of relevant events (200+)
	if len(cfg.RelevantEvents) < 100 {
		t.Errorf("Expected at least 100 relevant events, got %d", len(cfg.RelevantEvents))
	}
}
