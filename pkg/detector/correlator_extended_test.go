package detector

import (
	"sync"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestCrossCloudCorrelator_TimeWindowBoundary tests events at the boundary of the time window
func TestCrossCloudCorrelator_TimeWindowBoundary(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Millisecond)

	// Add AWS event
	awsEvent := types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	}
	c.AddEvent(awsEvent)

	// Wait slightly less than the window
	time.Sleep(5 * time.Millisecond)

	// Add Azure event - should correlate
	azureEvent := types.Event{
		Provider:     "azure",
		EventName:    "Microsoft.Compute/virtualMachines/write",
		ResourceType: "azurerm_virtual_machine",
		ResourceID:   "test-vm",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	}
	groups := c.AddEvent(azureEvent)
	assert.NotEmpty(t, groups, "Events within time window should correlate")

	// Wait past the window
	time.Sleep(10 * time.Millisecond)

	// Add GCP event - should not correlate (outside window)
	gcpEvent := types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.patch",
		ResourceType: "google_compute_instance",
		ResourceID:   "gcp-vm",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	}
	groups = c.AddEvent(gcpEvent)
	assert.Empty(t, groups, "Events outside time window should not correlate")
}

func TestCrossCloudCorrelator_EmptyUserName(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// AWS event with no user name
	awsEvent := types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: ""}, // Empty
	}
	groups := c.AddEvent(awsEvent)
	assert.Empty(t, groups, "Event with empty user name should not correlate")
}

func TestCrossCloudCorrelator_GetGroupsByProviderExtended(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add single AWS event (won't create group)
	c.AddEvent(types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		UserIdentity: types.UserIdentity{UserName: "user1"},
	})

	// Get groups by provider - should be empty (single provider)
	awsGroups := c.GetGroupsByProvider("aws")
	assert.Empty(t, awsGroups, "Single AWS event should not create a group")

	// Get groups for non-existent provider
	nonexistentGroups := c.GetGroupsByProvider("nonexistent")
	assert.Empty(t, nonexistentGroups, "Should not find groups for non-existent provider")
}

func TestCrossCloudCorrelator_ConcurrentAddEvent(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Use WaitGroup to synchronize concurrent operations
	var wg sync.WaitGroup

	// Add events concurrently from multiple goroutines
	providers := []string{"aws", "azure", "gcp"}
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				event := types.Event{
					Provider:     providers[idx],
					ResourceType: "resource_" + providers[idx],
					ResourceID:   "id_" + providers[idx],
					UserIdentity: types.UserIdentity{UserName: "concurrent-user"},
				}
				c.AddEvent(event)
			}
		}(i)
	}

	wg.Wait()

	// Verify all events were added and correlations were found
	stats := c.Stats()
	assert.Equal(t, 15, stats["buffered_events"].(int), "All 15 events should be buffered")
}

func TestCrossCloudCorrelator_ConcurrentGetGroups(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add events to create groups
	c.AddEvent(types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		UserIdentity: types.UserIdentity{UserName: "user"},
	})
	c.AddEvent(types.Event{
		Provider:     "azure",
		ResourceType: "azurerm_virtual_machine",
		UserIdentity: types.UserIdentity{UserName: "user"},
	})

	// Call GetGroups concurrently
	var wg sync.WaitGroup
	results := make([][]CorrelationGroup, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = c.GetGroups()
		}(i)
	}

	wg.Wait()

	// All results should be consistent
	firstResult := results[0]
	for i := 1; i < 10; i++ {
		assert.Equal(t, len(firstResult), len(results[i]), "All concurrent reads should return same number of groups")
	}
}

func TestCrossCloudCorrelator_MaxEventsBufferPruning(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)
	// Manually set a small max for testing
	c.maxEvents = 100

	// Add more than maxEvents
	for i := 0; i < 150; i++ {
		event := types.Event{
			Provider:     "aws",
			ResourceType: "aws_instance",
			ResourceID:   "i-" + string(rune(i)),
		}
		c.AddEvent(event)
	}

	// Verify that the buffer was pruned
	stats := c.Stats()
	bufferedCount := stats["buffered_events"].(int)
	assert.LessOrEqual(t, bufferedCount, c.maxEvents, "Buffer size should not exceed maxEvents after pruning")
	assert.Greater(t, bufferedCount, 0, "Should keep at least some events after pruning")
}

func TestCrossCloudCorrelator_ZeroWindowDefaultsTo10Minutes(t *testing.T) {
	c := NewCrossCloudCorrelator(0) // Zero window

	stats := c.Stats()
	windowSeconds := stats["window_seconds"].(int)
	assert.Equal(t, 600, windowSeconds, "Zero window should default to 10 minutes (600 seconds)")
}

func TestCrossCloudCorrelator_DefaultWindow(t *testing.T) {
	c := NewCrossCloudCorrelator(5 * time.Minute)

	stats := c.Stats()
	windowSeconds := stats["window_seconds"].(int)
	assert.Equal(t, 300, windowSeconds, "Window should be set to 5 minutes (300 seconds)")
}

func TestCrossCloudCorrelator_CorrelationScoring(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add events with same user across multiple providers
	c.AddEvent(types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		UserIdentity: types.UserIdentity{UserName: "admin"},
	})

	c.AddEvent(types.Event{
		Provider:     "azure",
		ResourceType: "azurerm_virtual_machine",
		UserIdentity: types.UserIdentity{UserName: "admin"},
	})

	groups := c.GetGroups()
	assert.NotEmpty(t, groups, "Should have correlation groups")

	// Verify score is between 0.0 and 1.0
	for _, group := range groups {
		assert.GreaterOrEqual(t, group.Score, 0.0, "Score should be non-negative")
		assert.LessOrEqual(t, group.Score, 1.0, "Score should not exceed 1.0")
	}
}

func TestCrossCloudCorrelator_ResourcePatternScoreMultiplier(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Add firewall-type resources
	c.AddEvent(types.Event{
		Provider:     "aws",
		ResourceType: "aws_security_group",
	})

	groups := c.AddEvent(types.Event{
		Provider:     "azure",
		ResourceType: "azurerm_network_security_group",
	})

	// Find the pattern-based group
	var patternGroup *CorrelationGroup
	for i, g := range groups {
		if g.Reason == "related_resource_pattern:firewall" {
			patternGroup = &groups[i]
			break
		}
	}

	if patternGroup != nil {
		// Pattern-based score should be multiplied by 0.7
		assert.Less(t, patternGroup.Score, 1.0, "Pattern-based score should be reduced from max")
	}
}

func TestResourceCategory_AllCategories(t *testing.T) {
	categories := map[string][]string{
		"compute": {"aws_instance", "google_compute_instance", "azurerm_virtual_machine"},
		"network": {"aws_vpc", "google_compute_network", "azurerm_virtual_network"},
		"firewall": {"aws_security_group", "google_compute_firewall", "azurerm_network_security_group"},
		"load_balancer": {"aws_lb", "google_compute_backend_service"},
		"storage": {"aws_s3_bucket", "google_storage_bucket", "azurerm_storage_account"},
		"database": {"aws_db_instance", "google_sql_database_instance"},
		"kubernetes": {"aws_eks_cluster", "google_container_cluster"},
		"dns": {"aws_route53_zone", "google_dns_managed_zone"},
		"iam": {"aws_iam_role", "google_service_account", "azurerm_role_assignment"},
		"secrets": {"aws_secretsmanager_secret", "google_secret_manager_secret"},
	}

	for expectedCategory, resources := range categories {
		for _, resource := range resources {
			category := resourceCategory(resource)
			assert.Equal(t, expectedCategory, category, "Resource %s should be in %s category", resource, expectedCategory)
		}
	}
}

func TestItoa_Numbers(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{9, "9"},
		{10, "10"},
		{100, "100"},
		{999, "999"},
		{1234, "1234"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := itoa(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateScore_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		events    []timedEvent
		providers map[string]bool
		check     func(t *testing.T, score float64)
	}{
		{
			name:      "No events",
			events:    []timedEvent{},
			providers: make(map[string]bool),
			check: func(t *testing.T, score float64) {
				assert.Equal(t, 0.0, score)
			},
		},
		{
			name:      "One provider with two events",
			events:    []timedEvent{{}, {}},
			providers: map[string]bool{"aws": true},
			check: func(t *testing.T, score float64) {
				// 0.3 for provider + 0.2 for more than 1 event = 0.5
				assert.Equal(t, 0.5, score)
			},
		},
		{
			name:      "Two providers with two events",
			events:    []timedEvent{{}, {}},
			providers: map[string]bool{"aws": true, "azure": true},
			check: func(t *testing.T, score float64) {
				// 0.3*2 for 2 providers + 0.2 for more than 1 event = 0.8
				assert.Equal(t, 0.8, score)
			},
		},
		{
			name:      "Three providers with four events",
			events:    []timedEvent{{}, {}, {}, {}},
			providers: map[string]bool{"aws": true, "azure": true, "gcp": true},
			check: func(t *testing.T, score float64) {
				// Score should be capped at 1.0
				assert.LessOrEqual(t, score, 1.0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := calculateScore(tt.events, tt.providers)
			tt.check(t, score)
		})
	}
}

func TestTimeRange_EdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		events []timedEvent
		check  func(t *testing.T, start, end time.Time)
	}{
		{
			name:   "Empty events",
			events: []timedEvent{},
			check: func(t *testing.T, start, end time.Time) {
				// Should return valid times (both will be now)
				assert.NotZero(t, start)
				assert.NotZero(t, end)
				assert.Equal(t, start, end)
			},
		},
		{
			name: "Single event",
			events: []timedEvent{
				{ReceivedAt: time.Unix(1000, 0)},
			},
			check: func(t *testing.T, start, end time.Time) {
				assert.Equal(t, start, end)
				assert.Equal(t, time.Unix(1000, 0), start)
			},
		},
		{
			name: "Multiple events in order",
			events: []timedEvent{
				{ReceivedAt: time.Unix(1000, 0)},
				{ReceivedAt: time.Unix(2000, 0)},
				{ReceivedAt: time.Unix(3000, 0)},
			},
			check: func(t *testing.T, start, end time.Time) {
				assert.Equal(t, time.Unix(1000, 0), start)
				assert.Equal(t, time.Unix(3000, 0), end)
			},
		},
		{
			name: "Multiple events out of order",
			events: []timedEvent{
				{ReceivedAt: time.Unix(3000, 0)},
				{ReceivedAt: time.Unix(1000, 0)},
				{ReceivedAt: time.Unix(2000, 0)},
			},
			check: func(t *testing.T, start, end time.Time) {
				assert.Equal(t, time.Unix(1000, 0), start)
				assert.Equal(t, time.Unix(3000, 0), end)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end := timeRange(tt.events)
			tt.check(t, start, end)
		})
	}
}

func TestGenerateGroupID_Uniqueness(t *testing.T) {
	ids := make(map[string]bool)

	for i := 1; i <= 100; i++ {
		id := generateGroupID(i)
		assert.NotEmpty(t, id, "Generated ID should not be empty")
		assert.NotContains(t, ids, id, "Generated ID should be unique")
		ids[id] = true
	}

	assert.Equal(t, 100, len(ids), "All 100 IDs should be unique")
}

func TestCrossCloudCorrelator_CorrelationGroupStructure(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	c.AddEvent(types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		UserIdentity: types.UserIdentity{UserName: "testuser"},
	})

	groups := c.AddEvent(types.Event{
		Provider:     "azure",
		ResourceType: "azurerm_virtual_machine",
		UserIdentity: types.UserIdentity{UserName: "testuser"},
	})

	assert.NotEmpty(t, groups)
	group := groups[0]

	// Verify all fields are populated
	assert.NotEmpty(t, group.ID, "ID should be set")
	assert.NotEmpty(t, group.Events, "Events should be populated")
	assert.NotEmpty(t, group.Providers, "Providers should be populated")
	assert.Equal(t, "testuser", group.UserID, "UserID should be set")
	assert.False(t, group.StartTime.IsZero(), "StartTime should be set")
	assert.False(t, group.EndTime.IsZero(), "EndTime should be set")
	assert.Greater(t, group.Score, 0.0, "Score should be greater than 0")
	assert.NotEmpty(t, group.Reason, "Reason should be set")
}

func TestExtractEvents_PreservesEventData(t *testing.T) {
	timedEvents := []timedEvent{
		{
			Event: types.Event{
				Provider:     "aws",
				ResourceType: "aws_instance",
				ResourceID:   "i-123",
			},
			ReceivedAt: time.Now(),
		},
		{
			Event: types.Event{
				Provider:     "azure",
				ResourceType: "azurerm_vm",
				ResourceID:   "vm-456",
			},
			ReceivedAt: time.Now(),
		},
	}

	events := extractEvents(timedEvents)

	assert.Len(t, events, 2)
	assert.Equal(t, "aws", events[0].Provider)
	assert.Equal(t, "aws_instance", events[0].ResourceType)
	assert.Equal(t, "i-123", events[0].ResourceID)
	assert.Equal(t, "azure", events[1].Provider)
}
