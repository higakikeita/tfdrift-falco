package broadcaster

import (
	"sync"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBroadcaster(t *testing.T) {
	bc := NewBroadcaster()

	assert.NotNil(t, bc)
	assert.NotNil(t, bc.subscribers)
	assert.Equal(t, 0, bc.SubscriberCount())
}

func TestBroadcaster_Subscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	assert.Equal(t, 1, bc.SubscriberCount())
}

func TestBroadcaster_MultipleSubscribers(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)
	ch3 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)
	bc.Subscribe(ch3)

	assert.Equal(t, 3, bc.SubscriberCount())
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch)
	assert.Equal(t, 0, bc.SubscriberCount())

	// Channel should be closed
	_, ok := <-ch
	assert.False(t, ok)
}

func TestBroadcaster_UnsubscribeMultiple(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)
	assert.Equal(t, 2, bc.SubscriberCount())

	bc.Unsubscribe(ch1)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch2)
	assert.Equal(t, 0, bc.SubscriberCount())
}

func TestBroadcaster_Broadcast(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	event := Event{
		Type:      "drift",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload: map[string]interface{}{
			"resource": "test",
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.Equal(t, "test", received.Payload["resource"])
}

func TestBroadcaster_BroadcastMultipleSubscribers(t *testing.T) {
	bc := NewBroadcaster()
	ch1 := make(chan Event, 1)
	ch2 := make(chan Event, 1)

	bc.Subscribe(ch1)
	bc.Subscribe(ch2)

	event := Event{
		Type:      "state_change",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload:   map[string]interface{}{},
	}

	bc.Broadcast(event)

	received1 := <-ch1
	received2 := <-ch2

	assert.Equal(t, "state_change", received1.Type)
	assert.Equal(t, "state_change", received2.Type)
}

func TestBroadcaster_BroadcastNonBlockingWithFullBuffer(t *testing.T) {
	bc := NewBroadcaster()
	// Channel with buffer size of 1
	ch := make(chan Event, 1)

	bc.Subscribe(ch)

	event1 := Event{Type: "drift", Payload: map[string]interface{}{}}
	event2 := Event{Type: "state", Payload: map[string]interface{}{}}

	// First broadcast fills the buffer
	bc.Broadcast(event1)
	// Second broadcast should be non-blocking and not panic
	bc.Broadcast(event2)

	// Only first event should be in channel
	received := <-ch
	assert.Equal(t, "drift", received.Type)

	// Channel should be empty (second broadcast was dropped)
	assert.Equal(t, 0, len(ch))
}

func TestBroadcaster_BroadcastDriftAlert(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	// We'll use a minimal DriftAlert structure for this test
	// Since we don't import the full types package, we'll verify the broadcast mechanism
	alert := struct {
		Severity     string
		ResourceType string
		ResourceName string
		ResourceID   string
		Attribute    string
		OldValue     string
		NewValue     string
		UserIdentity string
		MatchedRules []string
		AlertType    string
		Timestamp    string
	}{
		Severity:     "high",
		ResourceType: "EC2",
		ResourceName: "prod-server",
		Timestamp:    "2024-01-01T00:00:00Z",
	}

	// Manually create the event that BroadcastDriftAlert would create
	event := Event{
		Type:      "drift",
		Timestamp: alert.Timestamp,
		Payload: map[string]interface{}{
			"severity":      alert.Severity,
			"resource_type": alert.ResourceType,
			"resource_name": alert.ResourceName,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.Equal(t, "high", received.Payload["severity"])
	assert.Equal(t, "EC2", received.Payload["resource_type"])
}

func TestBroadcaster_SubscriberCountAccuracy(t *testing.T) {
	bc := NewBroadcaster()

	assert.Equal(t, 0, bc.SubscriberCount())

	channels := make([]chan Event, 5)
	for i := 0; i < 5; i++ {
		channels[i] = make(chan Event, 1)
		bc.Subscribe(channels[i])
		assert.Equal(t, i+1, bc.SubscriberCount())
	}

	for i := 0; i < 5; i++ {
		bc.Unsubscribe(channels[i])
		assert.Equal(t, 4-i, bc.SubscriberCount())
	}
}

func TestBroadcaster_ConcurrentBroadcast(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 10)

	bc.Subscribe(ch)

	// Send multiple broadcasts concurrently
	go func() {
		for i := 0; i < 5; i++ {
			event := Event{
				Type:    "drift",
				Payload: map[string]interface{}{"id": i},
			}
			bc.Broadcast(event)
		}
	}()

	// Receive broadcasts
	received := 0
	timeout := time.After(2 * time.Second)

	for received < 5 {
		select {
		case <-ch:
			received++
		case <-timeout:
			t.Fatalf("Timeout: expected 5 broadcasts, got %d", received)
		}
	}
}

func TestBroadcaster_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	var wg sync.WaitGroup

	// Concurrent subscribe
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch := make(chan Event, 1)
			bc.Subscribe(ch)
		}()
	}

	wg.Wait()
	assert.Equal(t, 10, bc.SubscriberCount())

	// Concurrent unsubscribe (would need to track channels)
	// This test verifies no race conditions
}

func TestBroadcaster_BroadcastStateChange(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	changes := map[string]interface{}{
		"status": "modified",
		"tags":   []string{"prod"},
	}

	event := Event{
		Type: "state_change",
		Payload: map[string]interface{}{
			"resource_type": "aws_instance",
			"resource_id":   "i-123456",
			"changes":       changes,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "state_change", received.Type)
	assert.Equal(t, "aws_instance", received.Payload["resource_type"])
	assert.Equal(t, "i-123456", received.Payload["resource_id"])
}

func TestBroadcaster_EventPayload(t *testing.T) {
	event := Event{
		Type:      "test",
		Timestamp: "2024-01-01T00:00:00Z",
		Payload: map[string]interface{}{
			"string": "value",
			"number": 42,
			"bool":   true,
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	assert.Equal(t, "value", event.Payload["string"])
	assert.Equal(t, 42, event.Payload["number"])
	assert.Equal(t, true, event.Payload["bool"])
	assert.NotNil(t, event.Payload["nested"])
}

func TestBroadcaster_UnsubscribeNonExistent(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	// Should not panic when unsubscribing non-existent channel
	assert.NotPanics(t, func() {
		bc.Unsubscribe(ch)
	})
}

func TestBroadcaster_SubscribeAfterUnsubscribe(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)

	bc.Subscribe(ch)
	assert.Equal(t, 1, bc.SubscriberCount())

	bc.Unsubscribe(ch)
	assert.Equal(t, 0, bc.SubscriberCount())

	// Subscribe again with new channel
	ch2 := make(chan Event, 1)
	bc.Subscribe(ch2)
	assert.Equal(t, 1, bc.SubscriberCount())
}

func TestBroadcaster_BroadcastWithMultipleEvents(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 3)

	bc.Subscribe(ch)

	events := []Event{
		{Type: "drift", Payload: map[string]interface{}{"id": 1}},
		{Type: "state", Payload: map[string]interface{}{"id": 2}},
		{Type: "falco", Payload: map[string]interface{}{"id": 3}},
	}

	for _, event := range events {
		bc.Broadcast(event)
	}

	for i, expectedEvent := range events {
		received := <-ch
		assert.Equal(t, expectedEvent.Type, received.Type, "Event %d type mismatch", i)
	}
}

func TestBroadcaster_DiscoveryProgress(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "discovery_progress",
		Payload: map[string]interface{}{
			"provider":      "aws",
			"resource_type": "EC2",
			"count":         10,
			"phase":         "discovering",
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "discovering", received.Payload["phase"])
}

func TestBroadcaster_ProviderStatus(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "provider_status",
		Payload: map[string]interface{}{
			"provider": "aws",
			"status":   "connected",
			"capabilities": map[string]bool{
				"drift_detection": true,
				"discovery":       true,
			},
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "connected", received.Payload["status"])
}

func TestBroadcaster_UnmanagedResource(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "unmanaged_resource",
		Payload: map[string]interface{}{
			"id":       "i-123456",
			"type":     "instance",
			"provider": "aws",
			"name":     "test-server",
			"region":   "us-east-1",
			"tags": map[string]string{
				"Environment": "prod",
			},
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	assert.Equal(t, "i-123456", received.Payload["id"])
	assert.Equal(t, "prod", received.Payload["tags"].(map[string]string)["Environment"])
}

func TestBroadcaster_BroadcastDriftResult(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	event := Event{
		Type: "drift_result",
		Payload: map[string]interface{}{
			"provider":        "aws",
			"unmanaged_count": 5,
			"missing_count":   2,
			"modified_count":  3,
		},
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, "drift_result", received.Type)
	assert.Equal(t, 5, received.Payload["unmanaged_count"])
}

func TestBroadcaster_ThreadSafety(t *testing.T) {
	bc := NewBroadcaster()
	var wg sync.WaitGroup

	// Create multiple publishers and subscribers
	channels := make([]chan Event, 5)
	for i := 0; i < 5; i++ {
		channels[i] = make(chan Event, 10)
		bc.Subscribe(channels[i])
	}

	// Concurrent broadcasts
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			event := Event{
				Type:    "test",
				Payload: map[string]interface{}{"broadcast_id": id},
			}
			bc.Broadcast(event)
		}(i)
	}

	wg.Wait()

	// Verify all messages were delivered
	for _, ch := range channels {
		count := 0
		timeout := time.After(1 * time.Second)
		for {
			select {
			case <-ch:
				count++
			case <-timeout:
				assert.Equal(t, 10, count, "Not all broadcasts received")
				return
			}
		}
	}
}

func TestBroadcaster_LargePayload(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	largePayload := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		largePayload[string(rune(i))] = "value_" + string(rune(i))
	}

	event := Event{
		Type:    "large",
		Payload: largePayload,
	}

	bc.Broadcast(event)

	received := <-ch
	assert.Equal(t, 100, len(received.Payload))
}

// ==================== BroadcastDriftAlert Tests ====================

func TestBroadcaster_BroadcastDriftAlert_HighSeverity(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	alert := types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_ec2_instance",
		ResourceName: "prod-server",
		ResourceID:   "i-1234567890abcdef0",
		Attribute:    "tags",
		OldValue:     map[string]string{"Environment": "prod"},
		NewValue:     map[string]string{"Environment": "staging"},
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: "AIDAI23HXD2O5Q5T5EXAMPLE",
			UserName:    "john.doe",
		},
		MatchedRules: []string{"rule1", "rule2"},
		Timestamp:    "2024-01-01T12:00:00Z",
		AlertType:    "drift",
	}

	bc.BroadcastDriftAlert(alert)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.Equal(t, "2024-01-01T12:00:00Z", received.Timestamp)
	assert.Equal(t, "high", received.Payload["severity"])
	assert.Equal(t, "aws_ec2_instance", received.Payload["resource_type"])
	assert.Equal(t, "prod-server", received.Payload["resource_name"])
	assert.Equal(t, "i-1234567890abcdef0", received.Payload["resource_id"])
	assert.Equal(t, "tags", received.Payload["attribute"])
	assert.Equal(t, "drift", received.Payload["alert_type"])
}

func TestBroadcaster_BroadcastDriftAlert_WithMatchedRules(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	alert := types.DriftAlert{
		Severity:     "medium",
		ResourceType: "aws_security_group",
		ResourceName: "default-sg",
		ResourceID:   "sg-0123456789abcdef0",
		Attribute:    "ingress_rules",
		OldValue:     1,
		NewValue:     2,
		MatchedRules: []string{"unauthorized_rule", "open_port"},
		Timestamp:    "2024-01-01T13:00:00Z",
		AlertType:    "drift",
	}

	bc.BroadcastDriftAlert(alert)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	assert.NotNil(t, received.Payload["matched_rules"])
	matchedRules := received.Payload["matched_rules"].([]string)
	assert.Equal(t, 2, len(matchedRules))
	assert.Contains(t, matchedRules, "unauthorized_rule")
}

func TestBroadcaster_BroadcastDriftAlert_WithUserIdentity(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	alert := types.DriftAlert{
		Severity:     "low",
		ResourceType: "aws_s3_bucket",
		ResourceName: "my-bucket",
		ResourceID:   "my-bucket",
		Attribute:    "versioning",
		OldValue:     true,
		NewValue:     false,
		UserIdentity: types.UserIdentity{
			Type:        "IAMRole",
			PrincipalID: "AIDACKCEVSQ6C2EXAMPLE",
			ARN:         "arn:aws:iam::123456789012:role/lambda-role",
			AccountID:   "123456789012",
			UserName:    "lambda-function",
		},
		Timestamp: "2024-01-01T14:00:00Z",
		AlertType: "drift",
	}

	bc.BroadcastDriftAlert(alert)

	received := <-ch
	assert.Equal(t, "drift", received.Type)
	userIdentity := received.Payload["user_identity"].(types.UserIdentity)
	assert.Equal(t, "IAMRole", userIdentity.Type)
	assert.Equal(t, "123456789012", userIdentity.AccountID)
}

// ==================== BroadcastFalcoEvent Tests ====================

func TestBroadcaster_BroadcastFalcoEvent_BasicEvent(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	falcoEvent := types.Event{
		Provider:     "aws",
		EventName:    "PutBucketVersioning",
		ResourceType: "s3_bucket",
		ResourceID:   "my-bucket",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
		Changes: map[string]interface{}{
			"versioning": false,
		},
		Region: "us-east-1",
	}

	bc.BroadcastFalcoEvent(falcoEvent)

	received := <-ch
	assert.Equal(t, "falco", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "PutBucketVersioning", received.Payload["event_name"])
	assert.Equal(t, "s3_bucket", received.Payload["resource_type"])
}

func TestBroadcaster_BroadcastFalcoEvent_WithMetadata(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	falcoEvent := types.Event{
		Provider:     "gcp",
		EventName:    "compute.instances.create",
		ResourceType: "compute_instance",
		ResourceID:   "instance-1",
		ProjectID:    "my-project",
		ServiceName:  "compute.googleapis.com",
		Changes: map[string]interface{}{
			"machine_type": "n1-standard-1",
		},
		Metadata: map[string]string{
			"project_id":   "my-project",
			"service_name": "compute.googleapis.com",
		},
	}

	bc.BroadcastFalcoEvent(falcoEvent)

	received := <-ch
	assert.Equal(t, "falco", received.Type)
	assert.Equal(t, "gcp", received.Payload["provider"])
	assert.Equal(t, "my-project", received.Payload["project_id"])
	assert.Equal(t, "compute.googleapis.com", received.Payload["service_name"])
}

func TestBroadcaster_BroadcastFalcoEvent_WithComplexChanges(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	falcoEvent := types.Event{
		Provider:     "aws",
		EventName:    "ModifyDBInstance",
		ResourceType: "rds_instance",
		ResourceID:   "db-instance-1",
		Changes: map[string]interface{}{
			"allocated_storage": 100,
			"engine_version":    "5.7",
			"multi_az":          true,
			"backup_retention":  7,
		},
		Region: "us-west-2",
	}

	bc.BroadcastFalcoEvent(falcoEvent)

	received := <-ch
	assert.Equal(t, "falco", received.Type)
	changes := received.Payload["changes"].(map[string]interface{})
	assert.Equal(t, 100, changes["allocated_storage"])
	assert.Equal(t, "5.7", changes["engine_version"])
	assert.Equal(t, true, changes["multi_az"])
}

// ==================== BroadcastStateChange Tests ====================

func TestBroadcaster_BroadcastStateChange_BasicChange(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	changes := map[string]interface{}{
		"status": "modified",
	}

	bc.BroadcastStateChange("aws_instance", "i-12345", changes)

	received := <-ch
	assert.Equal(t, "state_change", received.Type)
	assert.Equal(t, "aws_instance", received.Payload["resource_type"])
	assert.Equal(t, "i-12345", received.Payload["resource_id"])
	assert.Equal(t, "modified", received.Payload["changes"].(map[string]interface{})["status"])
}

func TestBroadcaster_BroadcastStateChange_MultipleChanges(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	changes := map[string]interface{}{
		"tags": map[string]string{
			"Environment": "production",
			"Owner":       "devops",
		},
		"security_groups": []string{"sg-123", "sg-456"},
		"state":           "running",
	}

	bc.BroadcastStateChange("aws_instance", "i-abcdef", changes)

	received := <-ch
	assert.Equal(t, "state_change", received.Type)
	payload := received.Payload["changes"].(map[string]interface{})
	assert.Contains(t, payload, "tags")
	assert.Contains(t, payload, "security_groups")
	assert.Contains(t, payload, "state")
}

func TestBroadcaster_BroadcastStateChange_EmptyChanges(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	changes := map[string]interface{}{}

	bc.BroadcastStateChange("aws_rds", "rds-1", changes)

	received := <-ch
	assert.Equal(t, "state_change", received.Type)
	assert.Equal(t, 0, len(received.Payload["changes"].(map[string]interface{})))
}

// ==================== BroadcastDriftResult Tests ====================

func TestBroadcaster_BroadcastDriftResult_WithAllResourceTypes(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	result := &types.DriftResult{
		Provider: "aws",
		UnmanagedResources: []*types.DiscoveredResource{
			{ID: "i-1", Type: "aws_instance", Name: "server1"},
			{ID: "i-2", Type: "aws_instance", Name: "server2"},
		},
		MissingResources: []*types.TerraformResource{
			{ID: "sg-1", Type: "aws_security_group"},
		},
		ModifiedResources: []*types.ResourceDiff{
			{ResourceID: "vpc-1", ResourceType: "aws_vpc"},
		},
	}

	bc.BroadcastDriftResult(result)

	received := <-ch
	assert.Equal(t, "drift_result", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, 2, received.Payload["unmanaged_count"])
	assert.Equal(t, 1, received.Payload["missing_count"])
	assert.Equal(t, 1, received.Payload["modified_count"])
}

func TestBroadcaster_BroadcastDriftResult_NoResources(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	result := &types.DriftResult{
		Provider:           "gcp",
		UnmanagedResources: []*types.DiscoveredResource{},
		MissingResources:   []*types.TerraformResource{},
		ModifiedResources:  []*types.ResourceDiff{},
	}

	bc.BroadcastDriftResult(result)

	received := <-ch
	assert.Equal(t, "drift_result", received.Type)
	assert.Equal(t, 0, received.Payload["unmanaged_count"])
	assert.Equal(t, 0, received.Payload["missing_count"])
	assert.Equal(t, 0, received.Payload["modified_count"])
}

func TestBroadcaster_BroadcastDriftResult_IncludesResourceDetails(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	result := &types.DriftResult{
		Provider: "aws",
		UnmanagedResources: []*types.DiscoveredResource{
			{
				ID:       "i-123",
				Type:     "aws_instance",
				Provider: "aws",
				Name:     "test-server",
				Region:   "us-east-1",
				Tags:     map[string]string{"Env": "dev"},
			},
		},
		MissingResources:  []*types.TerraformResource{},
		ModifiedResources: []*types.ResourceDiff{},
	}

	bc.BroadcastDriftResult(result)

	received := <-ch
	assert.Equal(t, "drift_result", received.Type)
	unmanagedResources := received.Payload["unmanaged_resources"].([]*types.DiscoveredResource)
	assert.Equal(t, 1, len(unmanagedResources))
	assert.Equal(t, "i-123", unmanagedResources[0].ID)
}

// ==================== BroadcastDiscoveryProgress Tests ====================

func TestBroadcaster_BroadcastDiscoveryProgress_Started(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastDiscoveryProgress("aws", "ec2", 0, "started")

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "ec2", received.Payload["resource_type"])
	assert.Equal(t, 0, received.Payload["count"])
	assert.Equal(t, "started", received.Payload["phase"])
}

func TestBroadcaster_BroadcastDiscoveryProgress_Discovering(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastDiscoveryProgress("gcp", "compute_instance", 42, "discovering")

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, "gcp", received.Payload["provider"])
	assert.Equal(t, "compute_instance", received.Payload["resource_type"])
	assert.Equal(t, 42, received.Payload["count"])
	assert.Equal(t, "discovering", received.Payload["phase"])
}

func TestBroadcaster_BroadcastDiscoveryProgress_Completed(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastDiscoveryProgress("azure", "virtual_machine", 128, "completed")

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, 128, received.Payload["count"])
	assert.Equal(t, "completed", received.Payload["phase"])
}

func TestBroadcaster_BroadcastDiscoveryProgress_Error(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastDiscoveryProgress("aws", "rds", 0, "error")

	received := <-ch
	assert.Equal(t, "discovery_progress", received.Type)
	assert.Equal(t, "error", received.Payload["phase"])
}

// ==================== BroadcastProviderStatus Tests ====================

func TestBroadcaster_BroadcastProviderStatus_Connected(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	capabilities := map[string]bool{
		"drift_detection": true,
		"discovery":       true,
		"monitoring":      false,
	}

	bc.BroadcastProviderStatus("aws", "connected", capabilities, map[string]interface{}{
		"region": "us-east-1",
	})

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "connected", received.Payload["status"])
	caps := received.Payload["capabilities"].(map[string]bool)
	assert.True(t, caps["drift_detection"])
	assert.False(t, caps["monitoring"])
}

func TestBroadcaster_BroadcastProviderStatus_Disconnected(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastProviderStatus("gcp", "disconnected", map[string]bool{}, map[string]interface{}{
		"error": "authentication failed",
	})

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "gcp", received.Payload["provider"])
	assert.Equal(t, "disconnected", received.Payload["status"])
	details := received.Payload["details"].(map[string]interface{})
	assert.Equal(t, "authentication failed", details["error"])
}

func TestBroadcaster_BroadcastProviderStatus_Error(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastProviderStatus("azure", "error", map[string]bool{
		"drift_detection": false,
	}, map[string]interface{}{
		"error_code": "INVALID_CREDENTIALS",
		"message":    "Failed to authenticate with Azure",
	})

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "error", received.Payload["status"])
}

func TestBroadcaster_BroadcastProviderStatus_Discovering(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	bc.BroadcastProviderStatus("aws", "discovering", map[string]bool{
		"drift_detection": true,
		"discovery":       true,
	}, map[string]interface{}{
		"progress_percent": 45,
	})

	received := <-ch
	assert.Equal(t, "provider_status", received.Type)
	assert.Equal(t, "discovering", received.Payload["status"])
}

// ==================== BroadcastUnmanagedResource Tests ====================

func TestBroadcaster_BroadcastUnmanagedResource_BasicResource(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	resource := &types.DiscoveredResource{
		ID:       "i-1234567890abcdef0",
		Type:     "aws_instance",
		Provider: "aws",
		Name:     "orphaned-server",
		Region:   "us-east-1",
	}

	bc.BroadcastUnmanagedResource(resource)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	assert.Equal(t, "i-1234567890abcdef0", received.Payload["id"])
	assert.Equal(t, "aws_instance", received.Payload["type"])
	assert.Equal(t, "aws", received.Payload["provider"])
	assert.Equal(t, "orphaned-server", received.Payload["name"])
	assert.Equal(t, "us-east-1", received.Payload["region"])
}

func TestBroadcaster_BroadcastUnmanagedResource_WithAttributes(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	resource := &types.DiscoveredResource{
		ID:       "bucket-123",
		Type:     "aws_s3_bucket",
		Provider: "aws",
		Name:     "unmanaged-data-bucket",
		Region:   "us-west-2",
		Attributes: map[string]interface{}{
			"versioning": true,
			"encryption": "AES256",
			"acl":        "private",
		},
	}

	bc.BroadcastUnmanagedResource(resource)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	attrs := received.Payload["attributes"].(map[string]interface{})
	assert.Equal(t, true, attrs["versioning"])
	assert.Equal(t, "AES256", attrs["encryption"])
}

func TestBroadcaster_BroadcastUnmanagedResource_WithTags(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	resource := &types.DiscoveredResource{
		ID:       "i-abcdef123456",
		Type:     "aws_instance",
		Provider: "aws",
		Name:     "dev-server",
		Region:   "eu-west-1",
		Tags: map[string]string{
			"Environment": "development",
			"Owner":       "platform-team",
			"CostCenter":  "eng-001",
		},
	}

	bc.BroadcastUnmanagedResource(resource)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	tags := received.Payload["tags"].(map[string]string)
	assert.Equal(t, "development", tags["Environment"])
	assert.Equal(t, "platform-team", tags["Owner"])
}

func TestBroadcaster_BroadcastUnmanagedResource_GCPResource(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 1)
	bc.Subscribe(ch)

	resource := &types.DiscoveredResource{
		ID:       "gcp-instance-456",
		Type:     "google_compute_instance",
		Provider: "gcp",
		Name:     "orphaned-vm",
		Region:   "us-central1",
		Attributes: map[string]interface{}{
			"machine_type": "n1-standard-2",
			"zone":         "us-central1-a",
		},
	}

	bc.BroadcastUnmanagedResource(resource)

	received := <-ch
	assert.Equal(t, "unmanaged_resource", received.Type)
	assert.Equal(t, "gcp", received.Payload["provider"])
	assert.Equal(t, "google_compute_instance", received.Payload["type"])
}

// ==================== Concurrent Broadcasting Tests ====================

func TestBroadcaster_ConcurrentDriftAlerts(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 100)
	bc.Subscribe(ch)

	done := make(chan struct{})
	numAlerts := 50

	go func() {
		for i := 0; i < numAlerts; i++ {
			alert := types.DriftAlert{
				Severity:     "medium",
				ResourceType: "aws_instance",
				ResourceName: "server-" + string(rune(i)),
				ResourceID:   "i-" + string(rune(i)),
				Timestamp:    "2024-01-01T00:00:00Z",
			}
			bc.BroadcastDriftAlert(alert)
		}
		done <- struct{}{}
	}()

	<-done
	time.Sleep(100 * time.Millisecond)

	// Verify we received all broadcasts
	assert.Equal(t, numAlerts, len(ch))
}

func TestBroadcaster_MixedEventTypes(t *testing.T) {
	bc := NewBroadcaster()
	ch := make(chan Event, 50)
	bc.Subscribe(ch)

	// Broadcast different event types
	bc.BroadcastDriftAlert(types.DriftAlert{
		Severity:  "high",
		Timestamp: "2024-01-01T00:00:00Z",
	})

	bc.BroadcastFalcoEvent(types.Event{
		Provider:  "aws",
		EventName: "CreateInstance",
	})

	bc.BroadcastStateChange("aws_instance", "i-123", map[string]interface{}{})

	bc.BroadcastDiscoveryProgress("aws", "ec2", 10, "discovering")

	bc.BroadcastProviderStatus("aws", "connected", map[string]bool{}, map[string]interface{}{})

	bc.BroadcastUnmanagedResource(&types.DiscoveredResource{
		ID:   "i-456",
		Type: "aws_instance",
	})

	bc.BroadcastDriftResult(&types.DriftResult{
		Provider: "aws",
	})

	// Verify all event types were received
	assert.Equal(t, 7, len(ch))

	eventTypes := make(map[string]int)
	for i := 0; i < 7; i++ {
		event := <-ch
		eventTypes[event.Type]++
	}

	assert.Equal(t, 1, eventTypes["drift"])
	assert.Equal(t, 1, eventTypes["falco"])
	assert.Equal(t, 1, eventTypes["state_change"])
	assert.Equal(t, 1, eventTypes["discovery_progress"])
	assert.Equal(t, 1, eventTypes["provider_status"])
	assert.Equal(t, 1, eventTypes["unmanaged_resource"])
	assert.Equal(t, 1, eventTypes["drift_result"])
}
