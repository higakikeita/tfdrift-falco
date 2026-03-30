// Package detector provides drift detection and event correlation.
//
// The CrossCloudCorrelator identifies related drift events across multiple
// cloud providers (AWS, GCP, Azure) within configurable time windows.
// It groups events by user identity, time proximity, and resource patterns
// to surface coordinated multi-cloud changes that may indicate larger
// infrastructure drift patterns.
package detector

import (
	"sort"
	"sync"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// timedEvent wraps a types.Event with a receive timestamp for correlation.
type timedEvent struct {
	types.Event
	ReceivedAt time.Time `json:"received_at"`
}

// CorrelationGroup represents a group of related drift events across clouds.
type CorrelationGroup struct {
	ID        string        `json:"id"`
	Events    []types.Event `json:"events"`
	Providers []string      `json:"providers"`
	UserID    string        `json:"user_id,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Score     float64       `json:"correlation_score"` // 0.0-1.0
	Reason    string        `json:"reason"`
}

// CrossCloudCorrelator correlates drift events across cloud providers.
type CrossCloudCorrelator struct {
	mu         sync.RWMutex
	events     []timedEvent
	groups     []CorrelationGroup
	window     time.Duration // Time window for correlation
	maxEvents  int           // Max events to keep in buffer
	groupIDSeq int
}

// NewCrossCloudCorrelator creates a correlator with the given time window.
func NewCrossCloudCorrelator(window time.Duration) *CrossCloudCorrelator {
	if window == 0 {
		window = 10 * time.Minute
	}
	return &CrossCloudCorrelator{
		events:    make([]timedEvent, 0, 1000),
		groups:    make([]CorrelationGroup, 0),
		window:    window,
		maxEvents: 10000,
	}
}

// AddEvent adds a new event and checks for correlations.
func (c *CrossCloudCorrelator) AddEvent(event types.Event) []CorrelationGroup {
	c.mu.Lock()
	defer c.mu.Unlock()

	te := timedEvent{Event: event, ReceivedAt: time.Now()}
	c.events = append(c.events, te)
	c.pruneOldEvents()

	return c.findCorrelations(te)
}

// GetGroups returns all current correlation groups.
func (c *CrossCloudCorrelator) GetGroups() []CorrelationGroup {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]CorrelationGroup, len(c.groups))
	copy(result, c.groups)
	return result
}

// GetGroupsByProvider returns groups involving a specific provider.
func (c *CrossCloudCorrelator) GetGroupsByProvider(provider string) []CorrelationGroup {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]CorrelationGroup, 0)
	for _, g := range c.groups {
		for _, p := range g.Providers {
			if p == provider {
				result = append(result, g)
				break
			}
		}
	}
	return result
}

// Stats returns correlation statistics.
func (c *CrossCloudCorrelator) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providerCounts := make(map[string]int)
	for _, e := range c.events {
		providerCounts[e.Provider]++
	}

	multiCloudGroups := 0
	for _, g := range c.groups {
		if len(g.Providers) > 1 {
			multiCloudGroups++
		}
	}

	return map[string]interface{}{
		"buffered_events":    len(c.events),
		"correlation_groups": len(c.groups),
		"multi_cloud_groups": multiCloudGroups,
		"events_by_provider": providerCounts,
		"window_seconds":     int(c.window.Seconds()),
	}
}

// findCorrelations looks for correlations with the newly added event.
func (c *CrossCloudCorrelator) findCorrelations(newEvent timedEvent) []CorrelationGroup {
	newGroups := make([]CorrelationGroup, 0)

	// Strategy 1: Same user, different providers, within time window
	if group := c.correlateByUser(newEvent); group != nil {
		newGroups = append(newGroups, *group)
	}

	// Strategy 2: Related resource types across clouds (e.g., VPC+VNet+Network)
	if group := c.correlateByResourcePattern(newEvent); group != nil {
		newGroups = append(newGroups, *group)
	}

	// Add new groups
	c.groups = append(c.groups, newGroups...)

	return newGroups
}

// correlateByUser finds events from the same user across different providers.
func (c *CrossCloudCorrelator) correlateByUser(newEvent timedEvent) *CorrelationGroup {
	if newEvent.UserIdentity.UserName == "" {
		return nil
	}

	related := make([]timedEvent, 0)
	providers := make(map[string]bool)

	now := time.Now()
	for _, e := range c.events {
		if e.UserIdentity.UserName == newEvent.UserIdentity.UserName &&
			e.Provider != newEvent.Provider &&
			now.Sub(e.ReceivedAt) <= c.window {
			related = append(related, e)
			providers[e.Provider] = true
		}
	}

	if len(related) == 0 {
		return nil
	}

	// Include the new event itself
	related = append(related, newEvent)
	providers[newEvent.Provider] = true

	if len(providers) < 2 {
		return nil
	}

	providerList := make([]string, 0, len(providers))
	for p := range providers {
		providerList = append(providerList, p)
	}
	sort.Strings(providerList)

	c.groupIDSeq++

	start, end := timeRange(related)
	return &CorrelationGroup{
		ID:        generateGroupID(c.groupIDSeq),
		Events:    extractEvents(related),
		Providers: providerList,
		UserID:    newEvent.UserIdentity.UserName,
		StartTime: start,
		EndTime:   end,
		Score:     calculateScore(related, providers),
		Reason:    "same_user_multi_cloud",
	}
}

// correlateByResourcePattern finds events with related resource types across clouds.
func (c *CrossCloudCorrelator) correlateByResourcePattern(newEvent timedEvent) *CorrelationGroup {
	pattern := resourceCategory(newEvent.ResourceType)
	if pattern == "" {
		return nil
	}

	related := make([]timedEvent, 0)
	providers := make(map[string]bool)

	now := time.Now()
	for _, e := range c.events {
		if e.Provider != newEvent.Provider &&
			resourceCategory(e.ResourceType) == pattern &&
			now.Sub(e.ReceivedAt) <= c.window {
			related = append(related, e)
			providers[e.Provider] = true
		}
	}

	if len(related) == 0 {
		return nil
	}

	related = append(related, newEvent)
	providers[newEvent.Provider] = true

	if len(providers) < 2 {
		return nil
	}

	providerList := make([]string, 0, len(providers))
	for p := range providers {
		providerList = append(providerList, p)
	}
	sort.Strings(providerList)

	c.groupIDSeq++

	start, end := timeRange(related)
	return &CorrelationGroup{
		ID:        generateGroupID(c.groupIDSeq),
		Events:    extractEvents(related),
		Providers: providerList,
		StartTime: start,
		EndTime:   end,
		Score:     calculateScore(related, providers) * 0.7, // Lower score for pattern-based
		Reason:    "related_resource_pattern:" + pattern,
	}
}

// resourceCategory maps Terraform resource types to abstract categories
// for cross-cloud correlation.
func resourceCategory(resourceType string) string {
	categories := map[string][]string{
		"compute": {
			"aws_instance", "google_compute_instance", "azurerm_virtual_machine",
			"azurerm_linux_virtual_machine", "azurerm_windows_virtual_machine",
		},
		"network": {
			"aws_vpc", "google_compute_network", "azurerm_virtual_network",
			"aws_subnet", "google_compute_subnetwork", "azurerm_subnet",
		},
		"firewall": {
			"aws_security_group", "google_compute_firewall", "azurerm_network_security_group",
		},
		"load_balancer": {
			"aws_lb", "aws_alb", "google_compute_backend_service",
			"azurerm_lb", "azurerm_application_gateway",
		},
		"storage": {
			"aws_s3_bucket", "google_storage_bucket", "azurerm_storage_account",
		},
		"database": {
			"aws_db_instance", "aws_rds_cluster", "google_sql_database_instance",
			"azurerm_mssql_server", "azurerm_mysql_flexible_server",
			"azurerm_postgresql_flexible_server",
		},
		"kubernetes": {
			"aws_eks_cluster", "google_container_cluster", "azurerm_kubernetes_cluster",
		},
		"dns": {
			"aws_route53_zone", "google_dns_managed_zone", "azurerm_dns_zone",
			"azurerm_private_dns_zone",
		},
		"iam": {
			"aws_iam_role", "aws_iam_policy", "google_service_account",
			"google_project_iam_binding", "azurerm_role_assignment",
			"azurerm_user_assigned_identity",
		},
		"secrets": {
			"aws_secretsmanager_secret", "google_secret_manager_secret",
			"azurerm_key_vault", "azurerm_key_vault_secret",
		},
	}

	for category, types := range categories {
		for _, t := range types {
			if t == resourceType {
				return category
			}
		}
	}
	return ""
}

func extractEvents(timed []timedEvent) []types.Event {
	result := make([]types.Event, len(timed))
	for i, te := range timed {
		result[i] = te.Event
	}
	return result
}

func generateGroupID(seq int) string {
	return "corr-" + time.Now().Format("20060102-150405") + "-" + itoa(seq)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	result := ""
	for i > 0 {
		result = string(rune('0'+i%10)) + result
		i /= 10
	}
	return result
}

func calculateScore(events []timedEvent, providers map[string]bool) float64 {
	score := 0.0
	// More providers = higher score
	score += float64(len(providers)) * 0.3
	// More events = higher score (diminishing returns)
	if len(events) > 1 {
		score += 0.2
	}
	if len(events) > 3 {
		score += 0.1
	}
	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}
	return score
}

func timeRange(events []timedEvent) (time.Time, time.Time) {
	if len(events) == 0 {
		now := time.Now()
		return now, now
	}
	start := events[0].ReceivedAt
	end := events[0].ReceivedAt
	for _, e := range events[1:] {
		if e.ReceivedAt.Before(start) {
			start = e.ReceivedAt
		}
		if e.ReceivedAt.After(end) {
			end = e.ReceivedAt
		}
	}
	return start, end
}

func (c *CrossCloudCorrelator) pruneOldEvents() {
	if len(c.events) <= c.maxEvents {
		return
	}
	// Keep only the latest half
	c.events = c.events[len(c.events)/2:]
}
