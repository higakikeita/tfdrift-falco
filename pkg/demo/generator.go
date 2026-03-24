// Package demo provides a sample event generator for demo mode.
// It generates realistic-looking drift events without requiring
// real cloud credentials or a running Falco instance.
package demo

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/api/broadcaster"
	"github.com/keitahigaki/tfdrift-falco/pkg/graph"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	log "github.com/sirupsen/logrus"
)

// Generator produces sample drift events for demo mode.
type Generator struct {
	broadcaster *broadcaster.Broadcaster
	graphStore  *graph.Store
	rng         *rand.Rand
}

// NewGenerator creates a new demo event generator.
func NewGenerator(bc *broadcaster.Broadcaster, gs *graph.Store) *Generator {
	return &Generator{
		broadcaster: bc,
		graphStore:  gs,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Start begins generating demo events. It seeds initial data then
// periodically emits new drift events until the context is cancelled.
func (g *Generator) Start(ctx context.Context) error {
	log.Info("[DEMO] Starting demo event generator")
	log.Info("[DEMO] Seeding initial sample events...")

	// Seed initial events
	g.seedInitialEvents()

	log.Infof("[DEMO] Seeded %d events and %d drifts", len(sampleEvents), len(sampleDrifts))
	log.Info("[DEMO] New events will appear every 8-15 seconds")

	// Periodically generate new events
	ticker := time.NewTicker(time.Duration(8+g.rng.Intn(8)) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("[DEMO] Demo event generator stopped")
			return nil
		case <-ticker.C:
			g.emitRandomEvent()
			// Randomize next interval
			ticker.Reset(time.Duration(8+g.rng.Intn(8)) * time.Second)
		}
	}
}

// seedInitialEvents populates the graph store with a batch of realistic events.
func (g *Generator) seedInitialEvents() {
	now := time.Now()

	for i, evt := range sampleEvents {
		// Stagger timestamps over the past 2 hours
		evt.Timestamp = now.Add(-time.Duration(len(sampleEvents)-i) * 8 * time.Minute).Format(time.RFC3339)
		g.graphStore.AddEvent(evt)
	}

	for i, drift := range sampleDrifts {
		drift.Timestamp = now.Add(-time.Duration(len(sampleDrifts)-i) * 10 * time.Minute).Format(time.RFC3339)
		g.graphStore.AddDrift(drift)

		// Also broadcast so SSE/WS clients see them
		g.broadcaster.BroadcastDriftAlert(drift)
	}
}

// emitRandomEvent generates and broadcasts a single random event.
func (g *Generator) emitRandomEvent() {
	now := time.Now().Format(time.RFC3339)

	// 70% chance drift, 30% chance falco event
	if g.rng.Float64() < 0.7 {
		drift := g.randomDrift(now)
		g.graphStore.AddDrift(drift)
		g.broadcaster.BroadcastDriftAlert(drift)
		log.Infof("[DEMO] Generated drift: %s %s (%s)", drift.ResourceType, drift.ResourceID, drift.Severity)
	} else {
		evt := g.randomEvent(now)
		g.graphStore.AddEvent(evt)
		g.broadcaster.BroadcastFalcoEvent(evt)
		log.Infof("[DEMO] Generated event: %s %s", evt.EventName, evt.ResourceID)
	}
}

func (g *Generator) randomDrift(timestamp string) types.DriftAlert {
	scenarios := []struct {
		resourceType string
		resourceID   string
		resourceName string
		attribute    string
		oldValue     interface{}
		newValue     interface{}
		severity     string
		user         string
	}{
		{"aws_security_group", "sg-0a1b2c3d4e5f67890", "web-prod-sg", "ingress.0.cidr_blocks", "10.0.0.0/8", "0.0.0.0/0", "critical", "alice@example.com"},
		{"aws_iam_role", "role-admin-escalated", "admin-role", "assume_role_policy", `{"Effect":"Deny"}`, `{"Effect":"Allow","Principal":"*"}`, "critical", "bob@example.com"},
		{"aws_s3_bucket", "my-prod-data-bucket", "prod-data", "versioning.enabled", true, false, "high", "charlie@example.com"},
		{"aws_rds_instance", "db-prod-main", "prod-database", "publicly_accessible", false, true, "critical", "dave@example.com"},
		{"aws_ec2_instance", "i-0abc123def456789", "api-server-1", "instance_type", "t3.medium", "t3.2xlarge", "medium", "eve@example.com"},
		{"aws_lambda_function", "arn:aws:lambda:us-east-1:123456789:function:payment-processor", "payment-processor", "runtime", "nodejs18.x", "nodejs14.x", "high", "frank@example.com"},
		{"aws_eks_cluster", "eks-prod-main", "prod-cluster", "endpoint_public_access", false, true, "high", "grace@example.com"},
		{"aws_dynamodb_table", "users-table-prod", "users-table", "billing_mode", "PROVISIONED", "PAY_PER_REQUEST", "medium", "heidi@example.com"},
		{"gcp_compute_firewall", "fw-allow-all-ingress", "allow-all-ingress", "source_ranges", "10.128.0.0/9", "0.0.0.0/0", "critical", "ivan@example.com"},
		{"gcp_storage_bucket", "prod-logs-bucket", "prod-logs", "uniform_bucket_level_access", true, false, "high", "judy@example.com"},
		{"aws_cloudwatch_log_group", "/aws/lambda/auth-service", "auth-logs", "retention_in_days", 90, 1, "medium", "kevin@example.com"},
		{"aws_kms_key", "key-prod-encryption", "prod-kms", "key_rotation_enabled", true, false, "high", "laura@example.com"},
	}

	s := scenarios[g.rng.Intn(len(scenarios))]
	return types.DriftAlert{
		Severity:     s.severity,
		ResourceType: s.resourceType,
		ResourceName: s.resourceName,
		ResourceID:   s.resourceID,
		Attribute:    s.attribute,
		OldValue:     s.oldValue,
		NewValue:     s.newValue,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: fmt.Sprintf("AIDA%s", randomID(g.rng, 16)),
			ARN:         fmt.Sprintf("arn:aws:iam::123456789012:user/%s", s.user),
			AccountID:   "123456789012",
			UserName:    s.user,
		},
		MatchedRules: []string{"security-group-open-access", "iam-policy-change", "encryption-disabled"},
		Timestamp:    timestamp,
		AlertType:    "drift",
	}
}

func (g *Generator) randomEvent(timestamp string) types.Event {
	events := []struct {
		provider     string
		eventName    string
		resourceType string
		resourceID   string
		user         string
		region       string
		severity     string
	}{
		{"aws", "AuthorizeSecurityGroupIngress", "aws_security_group", "sg-demo-" + randomID(g.rng, 8), "ops-user@example.com", "us-east-1", "high"},
		{"aws", "PutBucketPolicy", "aws_s3_bucket", "demo-bucket-" + randomID(g.rng, 6), "dev-user@example.com", "us-west-2", "medium"},
		{"aws", "ModifyDBInstance", "aws_rds_instance", "db-demo-" + randomID(g.rng, 6), "dba@example.com", "eu-west-1", "high"},
		{"aws", "UpdateFunctionConfiguration", "aws_lambda_function", "demo-function-" + randomID(g.rng, 6), "developer@example.com", "us-east-1", "low"},
		{"aws", "CreateRole", "aws_iam_role", "role-demo-" + randomID(g.rng, 6), "admin@example.com", "global", "high"},
		{"gcp", "compute.firewalls.patch", "gcp_compute_firewall", "fw-demo-" + randomID(g.rng, 6), "gcp-admin@example.com", "us-central1", "high"},
		{"gcp", "storage.buckets.update", "gcp_storage_bucket", "gcs-demo-" + randomID(g.rng, 6), "gcp-dev@example.com", "us-east4", "medium"},
		{"aws", "RunInstances", "aws_ec2_instance", "i-demo" + randomID(g.rng, 12), "ec2-user@example.com", "ap-northeast-1", "medium"},
	}

	e := events[g.rng.Intn(len(events))]
	return types.Event{
		Provider:     e.provider,
		EventName:    e.eventName,
		ResourceType: e.resourceType,
		ResourceID:   e.resourceID,
		UserIdentity: types.UserIdentity{
			Type:        "IAMUser",
			PrincipalID: fmt.Sprintf("AIDA%s", randomID(g.rng, 16)),
			ARN:         fmt.Sprintf("arn:aws:iam::123456789012:user/%s", e.user),
			AccountID:   "123456789012",
			UserName:    e.user,
		},
		Changes:   map[string]interface{}{"demo": true},
		Timestamp: timestamp,
		Severity:  e.severity,
		Status:    types.EventStatusOpen,
		Region:    e.region,
	}
}

func randomID(rng *rand.Rand, length int) string {
	const charset = "abcdef0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}

// sampleEvents is the initial seed of Falco events.
var sampleEvents = []types.Event{
	{
		Provider: "aws", EventName: "AuthorizeSecurityGroupIngress",
		ResourceType: "aws_security_group", ResourceID: "sg-0a1b2c3d4e5f67890",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "alice@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/alice"},
		Severity: "critical", Status: types.EventStatusOpen, Region: "us-east-1",
		Changes: map[string]interface{}{"ingress": map[string]interface{}{"from_port": 0, "to_port": 65535, "cidr": "0.0.0.0/0"}},
	},
	{
		Provider: "aws", EventName: "PutBucketVersioning",
		ResourceType: "aws_s3_bucket", ResourceID: "my-prod-data-bucket",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "charlie@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/charlie"},
		Severity: "high", Status: types.EventStatusOpen, Region: "us-east-1",
		Changes: map[string]interface{}{"versioning": map[string]interface{}{"status": "Suspended"}},
	},
	{
		Provider: "aws", EventName: "ModifyDBInstance",
		ResourceType: "aws_rds_instance", ResourceID: "db-prod-main",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "dave@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/dave"},
		Severity: "critical", Status: types.EventStatusOpen, Region: "us-east-1",
		Changes: map[string]interface{}{"publicly_accessible": true},
	},
	{
		Provider: "aws", EventName: "UpdateAssumeRolePolicy",
		ResourceType: "aws_iam_role", ResourceID: "role-admin-escalated",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "bob@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/bob"},
		Severity: "critical", Status: types.EventStatusAcknowledged, Region: "global",
		Changes: map[string]interface{}{"assume_role_policy": map[string]interface{}{"effect": "Allow", "principal": "*"}},
	},
	{
		Provider: "gcp", EventName: "compute.firewalls.patch",
		ResourceType: "gcp_compute_firewall", ResourceID: "fw-allow-all-ingress",
		UserIdentity: types.UserIdentity{Type: "ServiceAccount", UserName: "ivan@example.com", AccountID: "my-gcp-project"},
		Severity: "critical", Status: types.EventStatusOpen, ProjectID: "my-gcp-project", ServiceName: "compute.googleapis.com",
		Changes: map[string]interface{}{"source_ranges": []string{"0.0.0.0/0"}},
	},
	{
		Provider: "aws", EventName: "RunInstances",
		ResourceType: "aws_ec2_instance", ResourceID: "i-0abc123def456789",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "eve@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/eve"},
		Severity: "medium", Status: types.EventStatusOpen, Region: "us-west-2",
		Changes: map[string]interface{}{"instance_type": "t3.2xlarge"},
	},
	{
		Provider: "aws", EventName: "UpdateFunctionConfiguration",
		ResourceType: "aws_lambda_function", ResourceID: "arn:aws:lambda:us-east-1:123456789:function:payment-processor",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "frank@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/frank"},
		Severity: "high", Status: types.EventStatusOpen, Region: "us-east-1",
		Changes: map[string]interface{}{"runtime": "nodejs14.x"},
	},
	{
		Provider: "aws", EventName: "DisableKeyRotation",
		ResourceType: "aws_kms_key", ResourceID: "key-prod-encryption",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "laura@example.com", AccountID: "123456789012", ARN: "arn:aws:iam::123456789012:user/laura"},
		Severity: "high", Status: types.EventStatusOpen, Region: "us-east-1",
		Changes: map[string]interface{}{"key_rotation_enabled": false},
	},
}

// sampleDrifts is the initial seed of drift alerts.
var sampleDrifts = []types.DriftAlert{
	{
		Severity: "critical", ResourceType: "aws_security_group", ResourceName: "web-prod-sg",
		ResourceID: "sg-0a1b2c3d4e5f67890", Attribute: "ingress.0.cidr_blocks",
		OldValue: "10.0.0.0/8", NewValue: "0.0.0.0/0",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "alice@example.com", AccountID: "123456789012"},
		MatchedRules: []string{"security-group-open-access"}, AlertType: "drift",
	},
	{
		Severity: "critical", ResourceType: "aws_iam_role", ResourceName: "admin-role",
		ResourceID: "role-admin-escalated", Attribute: "assume_role_policy",
		OldValue: `{"Effect":"Deny","Principal":"arn:aws:iam::123456789012:root"}`,
		NewValue: `{"Effect":"Allow","Principal":"*"}`,
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "bob@example.com", AccountID: "123456789012"},
		MatchedRules: []string{"iam-privilege-escalation"}, AlertType: "drift",
	},
	{
		Severity: "high", ResourceType: "aws_s3_bucket", ResourceName: "prod-data",
		ResourceID: "my-prod-data-bucket", Attribute: "versioning.enabled",
		OldValue: true, NewValue: false,
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "charlie@example.com", AccountID: "123456789012"},
		MatchedRules: []string{"s3-versioning-required"}, AlertType: "drift",
	},
	{
		Severity: "critical", ResourceType: "aws_rds_instance", ResourceName: "prod-database",
		ResourceID: "db-prod-main", Attribute: "publicly_accessible",
		OldValue: false, NewValue: true,
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "dave@example.com", AccountID: "123456789012"},
		MatchedRules: []string{"rds-no-public-access"}, AlertType: "drift",
	},
	{
		Severity: "critical", ResourceType: "gcp_compute_firewall", ResourceName: "allow-all-ingress",
		ResourceID: "fw-allow-all-ingress", Attribute: "source_ranges",
		OldValue: "10.128.0.0/9", NewValue: "0.0.0.0/0",
		UserIdentity: types.UserIdentity{Type: "ServiceAccount", UserName: "ivan@example.com", AccountID: "my-gcp-project"},
		MatchedRules: []string{"firewall-open-access"}, AlertType: "drift",
	},
	{
		Severity: "medium", ResourceType: "aws_ec2_instance", ResourceName: "api-server-1",
		ResourceID: "i-0abc123def456789", Attribute: "instance_type",
		OldValue: "t3.medium", NewValue: "t3.2xlarge",
		UserIdentity: types.UserIdentity{Type: "IAMUser", UserName: "eve@example.com", AccountID: "123456789012"},
		MatchedRules: []string{"instance-type-change"}, AlertType: "drift",
	},
}
