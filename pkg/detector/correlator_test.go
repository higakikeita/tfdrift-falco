package detector

import (
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCrossCloudCorrelator_SameUserMultiCloud(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// AWS event from user@example.com
	awsEvent := types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	}
	groups := c.AddEvent(awsEvent)
	assert.Empty(t, groups, "Single provider event should not create a group")

	// Azure event from same user
	azureEvent := types.Event{
		Provider:     "azure",
		EventName:    "Microsoft.Compute/virtualMachines/write",
		ResourceType: "azurerm_virtual_machine",
		ResourceID:   "test-vm",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	}
	groups = c.AddEvent(azureEvent)
	assert.Len(t, groups, 2, "Same user across 2 clouds should create correlation groups (user + pattern)")

	// Verify the user-based group
	var userGroup *CorrelationGroup
	for i, g := range groups {
		if g.Reason == "same_user_multi_cloud" {
			userGroup = &groups[i]
			break
		}
	}
	assert.NotNil(t, userGroup, "Should have user-based correlation")
	assert.Equal(t, "user@example.com", userGroup.UserID)
	assert.Contains(t, userGroup.Providers, "aws")
	assert.Contains(t, userGroup.Providers, "azure")
	assert.Greater(t, userGroup.Score, 0.0)
}

func TestCrossCloudCorrelator_ResourcePattern(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// AWS security group change
	c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "AuthorizeSecurityGroupIngress",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-123",
		UserIdentity: types.UserIdentity{UserName: "alice"},
	})

	// Azure NSG change from different user
	groups := c.AddEvent(types.Event{
		Provider:     "azure",
		EventName:    "Microsoft.Network/networkSecurityGroups/write",
		ResourceType: "azurerm_network_security_group",
		ResourceID:   "test-nsg",
		UserIdentity: types.UserIdentity{UserName: "bob"},
	})

	// Should correlate by resource pattern (firewall category)
	var patternGroup *CorrelationGroup
	for i, g := range groups {
		if g.Reason == "related_resource_pattern:firewall" {
			patternGroup = &groups[i]
			break
		}
	}
	assert.NotNil(t, patternGroup, "Should have pattern-based correlation for firewall resources")
	assert.Contains(t, patternGroup.Providers, "aws")
	assert.Contains(t, patternGroup.Providers, "azure")
}

func TestCrossCloudCorrelator_NoCorrelationSameProvider(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	// Two AWS events from same user - should NOT correlate (same provider)
	c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "ModifyInstanceAttribute",
		ResourceType: "aws_instance",
		ResourceID:   "i-123",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	})

	groups := c.AddEvent(types.Event{
		Provider:     "aws",
		EventName:    "AuthorizeSecurityGroupIngress",
		ResourceType: "aws_security_group",
		ResourceID:   "sg-456",
		UserIdentity: types.UserIdentity{UserName: "user@example.com"},
	})

	assert.Empty(t, groups, "Same provider events should not create cross-cloud correlation")
}

func TestCrossCloudCorrelator_Stats(t *testing.T) {
	c := NewCrossCloudCorrelator(10 * time.Minute)

	c.AddEvent(types.Event{Provider: "aws", ResourceType: "aws_instance"})
	c.AddEvent(types.Event{Provider: "gcp", ResourceType: "google_compute_instance"})

	stats := c.Stats()
	assert.Equal(t, 2, stats["buffered_events"])
	assert.Equal(t, 10*60, stats["window_seconds"])

	byProvider := stats["events_by_provider"].(map[string]int)
	assert.Equal(t, 1, byProvider["aws"])
	assert.Equal(t, 1, byProvider["gcp"])
}

func TestResourceCategory(t *testing.T) {
	tests := []struct {
		resourceType string
		expected     string
	}{
		{"aws_instance", "compute"},
		{"google_compute_instance", "compute"},
		{"azurerm_virtual_machine", "compute"},
		{"aws_security_group", "firewall"},
		{"azurerm_network_security_group", "firewall"},
		{"aws_s3_bucket", "storage"},
		{"google_storage_bucket", "storage"},
		{"azurerm_storage_account", "storage"},
		{"aws_eks_cluster", "kubernetes"},
		{"google_container_cluster", "kubernetes"},
		{"azurerm_kubernetes_cluster", "kubernetes"},
		{"unknown_resource", ""},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			assert.Equal(t, tt.expected, resourceCategory(tt.resourceType))
		})
	}
}
