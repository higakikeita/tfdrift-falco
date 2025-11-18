package detector

import (
	"context"
	"testing"
	"time"

	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/keitahigaki/tfdrift-falco/pkg/diff"
	"github.com/keitahigaki/tfdrift-falco/pkg/falco"
	"github.com/keitahigaki/tfdrift-falco/pkg/notifier"
	"github.com/keitahigaki/tfdrift-falco/pkg/terraform"
	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectDrifts(t *testing.T) {
	d := &Detector{}

	tests := []struct {
		name     string
		resource *terraform.Resource
		changes  map[string]interface{}
		expected int
	}{
		{
			name: "Single drift detected",
			resource: &terraform.Resource{
				Type: "aws_instance",
				Name: "web",
				Attributes: map[string]interface{}{
					"instance_type": "t3.micro",
					"ami":           "ami-123",
				},
			},
			changes: map[string]interface{}{
				"instance_type": "t3.small", // Changed
			},
			expected: 1,
		},
		{
			name: "Multiple drifts detected",
			resource: &terraform.Resource{
				Type: "aws_security_group",
				Name: "web_sg",
				Attributes: map[string]interface{}{
					"description": "Old description",
					"name":        "old-name",
				},
			},
			changes: map[string]interface{}{
				"description": "New description", // Changed
				"name":        "new-name",        // Changed
			},
			expected: 2,
		},
		{
			name: "No drift - same values",
			resource: &terraform.Resource{
				Type: "aws_s3_bucket",
				Name: "data",
				Attributes: map[string]interface{}{
					"bucket": "my-bucket",
				},
			},
			changes: map[string]interface{}{
				"bucket": "my-bucket", // Same value
			},
			expected: 0,
		},
		{
			name: "New attribute added",
			resource: &terraform.Resource{
				Type: "aws_instance",
				Name: "app",
				Attributes: map[string]interface{}{
					"instance_type": "t3.micro",
				},
			},
			changes: map[string]interface{}{
				"instance_type": "t3.micro",
				"tags":          map[string]string{"env": "prod"}, // New attribute
			},
			expected: 1,
		},
		{
			name: "Empty changes",
			resource: &terraform.Resource{
				Type: "aws_instance",
				Name: "app",
				Attributes: map[string]interface{}{
					"instance_type": "t3.micro",
				},
			},
			changes:  map[string]interface{}{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drifts := d.detectDrifts(tt.resource, tt.changes)
			assert.Len(t, drifts, tt.expected)

			if tt.expected > 0 {
				// Verify drift structure
				for _, drift := range drifts {
					assert.NotEmpty(t, drift.Attribute)
					assert.NotNil(t, drift.NewValue)
				}
			}
		})
	}
}

func TestEvaluateRules(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "critical-sg-change",
			ResourceTypes:     []string{"aws_security_group"},
			WatchedAttributes: []string{"ingress", "egress"},
			Severity:          "critical",
		},
		{
			Name:              "instance-type-change",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"instance_type"},
			Severity:          "high",
		},
		{
			Name:              "tag-change",
			ResourceTypes:     []string{"aws_instance", "aws_s3_bucket"},
			WatchedAttributes: []string{"tags"},
			Severity:          "low",
		},
	}

	d := &Detector{
		cfg: &config.Config{
			DriftRules: driftRules,
		},
	}

	tests := []struct {
		name         string
		resourceType string
		attribute    string
		expected     []string
	}{
		{
			name:         "Security group ingress rule matches",
			resourceType: "aws_security_group",
			attribute:    "ingress",
			expected:     []string{"critical-sg-change"},
		},
		{
			name:         "Instance type change matches",
			resourceType: "aws_instance",
			attribute:    "instance_type",
			expected:     []string{"instance-type-change"},
		},
		{
			name:         "Tags match multiple resource types",
			resourceType: "aws_instance",
			attribute:    "tags",
			expected:     []string{"tag-change"},
		},
		{
			name:         "No rule matches",
			resourceType: "aws_lambda_function",
			attribute:    "runtime",
			expected:     nil,
		},
		{
			name:         "Resource type matches but attribute doesn't",
			resourceType: "aws_instance",
			attribute:    "ami",
			expected:     nil,
		},
		{
			name:         "Attribute matches but resource type doesn't",
			resourceType: "aws_rds_instance",
			attribute:    "tags",
			expected:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := d.evaluateRules(tt.resourceType, tt.attribute)
			assert.Equal(t, tt.expected, matched)
		})
	}
}

func TestGetSeverity(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:     "critical-rule",
			Severity: "critical",
		},
		{
			Name:     "high-rule",
			Severity: "high",
		},
		{
			Name:     "medium-rule",
			Severity: "medium",
		},
		{
			Name:     "low-rule",
			Severity: "low",
		},
	}

	d := &Detector{
		cfg: &config.Config{
			DriftRules: driftRules,
		},
	}

	tests := []struct {
		name         string
		matchedRules []string
		expected     string
	}{
		{
			name:         "Critical severity - highest priority",
			matchedRules: []string{"critical-rule", "high-rule", "medium-rule"},
			expected:     "critical",
		},
		{
			name:         "High severity",
			matchedRules: []string{"high-rule", "medium-rule", "low-rule"},
			expected:     "high",
		},
		{
			name:         "Medium severity",
			matchedRules: []string{"medium-rule", "low-rule"},
			expected:     "medium",
		},
		{
			name:         "Low severity (default)",
			matchedRules: []string{"low-rule"},
			expected:     "low",
		},
		{
			name:         "No matched rules - default to low",
			matchedRules: []string{},
			expected:     "low",
		},
		{
			name:         "Non-existent rule - default to low",
			matchedRules: []string{"non-existent-rule"},
			expected:     "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := d.getSeverity(tt.matchedRules)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestAttributeDrift_Structure(t *testing.T) {
	drift := AttributeDrift{
		Attribute: "instance_type",
		OldValue:  "t3.micro",
		NewValue:  "t3.small",
	}

	assert.Equal(t, "instance_type", drift.Attribute)
	assert.Equal(t, "t3.micro", drift.OldValue)
	assert.Equal(t, "t3.small", drift.NewValue)
}

func TestDetectDrifts_ComplexValues(t *testing.T) {
	d := &Detector{}

	resource := &terraform.Resource{
		Type: "aws_security_group",
		Name: "app_sg",
		Attributes: map[string]interface{}{
			"description": "Old security group",
			"vpc_id":      "vpc-123",
		},
	}

	changes := map[string]interface{}{
		"description": "New security group", // String change
		"vpc_id":      "vpc-456",            // String change
	}

	drifts := d.detectDrifts(resource, changes)

	require.Len(t, drifts, 2)
	// Check that both attributes are present (order not guaranteed from map iteration)
	attributes := []string{drifts[0].Attribute, drifts[1].Attribute}
	assert.Contains(t, attributes, "description")
	assert.Contains(t, attributes, "vpc_id")
	assert.NotNil(t, drifts[0].OldValue)
	assert.NotNil(t, drifts[0].NewValue)
}

func TestDetectDrifts_NilValues(t *testing.T) {
	d := &Detector{}

	resource := &terraform.Resource{
		Type: "aws_instance",
		Name: "web",
		Attributes: map[string]interface{}{
			"user_data": nil,
		},
	}

	changes := map[string]interface{}{
		"user_data": "#!/bin/bash\necho hello",
	}

	drifts := d.detectDrifts(resource, changes)

	require.Len(t, drifts, 1)
	assert.Equal(t, "user_data", drifts[0].Attribute)
	assert.Nil(t, drifts[0].OldValue)
	assert.Equal(t, "#!/bin/bash\necho hello", drifts[0].NewValue)
}

func TestEvaluateRules_MultipleMatches(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "all-resources-tags",
			ResourceTypes:     []string{"aws_instance", "aws_s3_bucket", "aws_rds_instance"},
			WatchedAttributes: []string{"tags"},
			Severity:          "low",
		},
		{
			Name:              "instance-specific-tags",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"tags"},
			Severity:          "medium",
		},
	}

	d := &Detector{
		cfg: &config.Config{
			DriftRules: driftRules,
		},
	}

	// Test that multiple rules can match the same resource type + attribute
	matched := d.evaluateRules("aws_instance", "tags")

	assert.Len(t, matched, 2)
	assert.Contains(t, matched, "all-resources-tags")
	assert.Contains(t, matched, "instance-specific-tags")
}

func TestGetSeverity_PriorityOrdering(t *testing.T) {
	driftRules := []config.DriftRule{
		{Name: "low-rule", Severity: "low"},
		{Name: "medium-rule", Severity: "medium"},
		{Name: "high-rule", Severity: "high"},
		{Name: "critical-rule", Severity: "critical"},
	}

	d := &Detector{
		cfg: &config.Config{
			DriftRules: driftRules,
		},
	}

	// Test that critical always wins, regardless of order
	tests := []struct {
		name         string
		matchedRules []string
		expected     string
	}{
		{
			name:         "Critical first",
			matchedRules: []string{"critical-rule", "low-rule"},
			expected:     "critical",
		},
		{
			name:         "Critical last",
			matchedRules: []string{"low-rule", "critical-rule"},
			expected:     "critical",
		},
		{
			name:         "High without critical",
			matchedRules: []string{"medium-rule", "high-rule", "low-rule"},
			expected:     "high",
		},
		{
			name:         "Medium without high or critical",
			matchedRules: []string{"low-rule", "medium-rule"},
			expected:     "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			severity := d.getSeverity(tt.matchedRules)
			assert.Equal(t, tt.expected, severity)
		})
	}
}

func TestDetectDrifts_EmptyResource(t *testing.T) {
	d := &Detector{}

	resource := &terraform.Resource{
		Type:       "aws_instance",
		Name:       "empty",
		Attributes: map[string]interface{}{},
	}

	changes := map[string]interface{}{
		"instance_type": "t3.micro",
		"ami":           "ami-123",
	}

	drifts := d.detectDrifts(resource, changes)

	// All changes should be detected as drifts since resource has no attributes
	assert.Equal(t, 2, len(drifts))
}

func TestEvaluateRules_EmptyRulesList(t *testing.T) {
	d := &Detector{
		cfg: &config.Config{
			DriftRules: []config.DriftRule{}, // No rules
		},
	}

	matched := d.evaluateRules("aws_instance", "instance_type")

	assert.Empty(t, matched)
}

func TestDetectDrifts_BooleanValues(t *testing.T) {
	d := &Detector{}

	resource := &terraform.Resource{
		Type: "aws_s3_bucket",
		Name: "data",
		Attributes: map[string]interface{}{
			"versioning_enabled": false,
		},
	}

	changes := map[string]interface{}{
		"versioning_enabled": true,
	}

	drifts := d.detectDrifts(resource, changes)

	require.Len(t, drifts, 1)
	assert.Equal(t, "versioning_enabled", drifts[0].Attribute)
	assert.Equal(t, false, drifts[0].OldValue)
	assert.Equal(t, true, drifts[0].NewValue)
}

func TestDetectDrifts_NumericValues(t *testing.T) {
	d := &Detector{}

	resource := &terraform.Resource{
		Type: "aws_rds_instance",
		Name: "db",
		Attributes: map[string]interface{}{
			"allocated_storage": 100,
		},
	}

	changes := map[string]interface{}{
		"allocated_storage": 200,
	}

	drifts := d.detectDrifts(resource, changes)

	require.Len(t, drifts, 1)
	assert.Equal(t, "allocated_storage", drifts[0].Attribute)
	assert.Equal(t, 100, drifts[0].OldValue)
	assert.Equal(t, 200, drifts[0].NewValue)
}

func TestEvaluateRules_CaseSensitivity(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "test-rule",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"tags"},
			Severity:          "medium",
		},
	}

	d := &Detector{
		cfg: &config.Config{
			DriftRules: driftRules,
		},
	}

	// Test case sensitivity - should NOT match
	matched := d.evaluateRules("AWS_INSTANCE", "tags")
	assert.Empty(t, matched)

	matched = d.evaluateRules("aws_instance", "TAGS")
	assert.Empty(t, matched)

	// Exact match should work
	matched = d.evaluateRules("aws_instance", "tags")
	assert.Len(t, matched, 1)
}

func TestNew_Success(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.stateManager)
	assert.NotNil(t, detector.falcoSubscriber)
	assert.NotNil(t, detector.notifier)
	assert.NotNil(t, detector.formatter)
	assert.NotNil(t, detector.eventCh)
	assert.Nil(t, detector.importer)
	assert.Nil(t, detector.approvalManager)
}

func TestNew_WithAutoImport(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
		AutoImport: config.AutoImportConfig{
			Enabled:         true,
			TerraformDir:    ".",
			RequireApproval: true,
		},
	}

	detector, err := New(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestNew_WithAutoImportAutoApproval(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance", "aws_s3_bucket"},
		},
	}

	detector, err := New(cfg)

	assert.NoError(t, err)
	assert.NotNil(t, detector)
	assert.NotNil(t, detector.importer)
	assert.NotNil(t, detector.approvalManager)
}

func TestHandleEvent_ResourceNotFound(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	// Create minimal detector with mock state manager
	stateManager := &terraform.StateManager{}
	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-nonexistent",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic when resource not found
	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleEvent_DriftDetected(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "instance-type-change",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"instance_type"},
			Severity:          "high",
		},
	}

	cfg := &config.Config{
		DriftRules: driftRules,
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	// Create state manager with test resource
	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	// Load the state
	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"instance_type": "t3.large", // Changed from t3.micro in state
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic when processing drift
	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleEvent_NoDrift(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	// Load the state
	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "DescribeInstances",
		Changes:      map[string]interface{}{
			// No actual changes
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic or produce alerts
	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestSendAlert_DryRun(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
	}

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
	}

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-0cea65ac652556767",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic in dry-run mode
	assert.NotPanics(t, func() {
		detector.sendAlert(alert)
	})
}

func TestSendUnmanagedResourceAlert_DryRun(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
	}

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic in dry-run mode
	assert.NotPanics(t, func() {
		detector.sendUnmanagedResourceAlert(event)
	})
}

func TestHandleEvent_WithTimestamp(t *testing.T) {
	driftRules := []config.DriftRule{
		{
			Name:              "instance-type-change",
			ResourceTypes:     []string{"aws_instance"},
			WatchedAttributes: []string{"instance_type"},
			Severity:          "high",
		},
	}

	cfg := &config.Config{
		DriftRules: driftRules,
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	// Load the state
	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
	}

	// Event with timestamp in RawEvent
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"instance_type": "t3.large",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
		RawEvent: map[string]interface{}{
			"eventTime": "2024-01-15T10:30:00Z",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestHandleAutoImport(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance"},
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-unmanaged-123",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_NotAllowed(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_s3_bucket"}, // Not aws_instance
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance", // Not in allowed list
		ResourceID:   "i-unmanaged-456",
		EventName:    "RunInstances",
		Changes:      map[string]interface{}{},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic even if not allowed
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleAutoImport_EmptyAllowList(t *testing.T) {
	cfg := &config.Config{
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{}, // Empty = allow all
			OutputDir:        "/tmp",
		},
	}

	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		importer:        importer,
		approvalManager: approvalManager,
	}

	event := &types.Event{
		Provider:     "aws",
		ResourceType: "aws_iam_role",
		ResourceID:   "role-123",
		EventName:    "CreateRole",
		Changes:      map[string]interface{}{},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
			ARN:      "arn:aws:iam::123456789012:user/admin",
		},
	}

	// Should not panic with empty allow list
	assert.NotPanics(t, func() {
		detector.handleAutoImport(context.Background(), event)
	})
}

func TestHandleEvent_WithAutoImport(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
		AutoImport: config.AutoImportConfig{
			Enabled:          true,
			TerraformDir:     ".",
			RequireApproval:  false,
			AllowedResources: []string{"aws_instance"},
		},
	}

	stateManager := &terraform.StateManager{}
	formatter := diff.NewFormatter(false)
	importer := terraform.NewImporter(".", true)
	approvalManager := terraform.NewApprovalManager(importer, false)

	detector := &Detector{
		cfg:             cfg,
		stateManager:    stateManager,
		formatter:       formatter,
		importer:        importer,
		approvalManager: approvalManager,
	}

	// Unmanaged resource should trigger auto-import
	event := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-new-resource",
		EventName:    "RunInstances",
		Changes: map[string]interface{}{
			"instance_type": "t3.micro",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	assert.NotPanics(t, func() {
		detector.handleEvent(event)
	})
}

func TestSendAlert_WithNotifier(t *testing.T) {
	cfg := &config.Config{
		DryRun: false, // Not dry-run
		Notifications: config.NotificationsConfig{
			Slack: config.SlackConfig{
				Enabled:    true,
				WebhookURL: "http://invalid-url",
				Channel:    "#test",
			},
		},
	}

	formatter := diff.NewFormatter(false)
	// Create notifier (will fail to send but shouldn't panic)
	notifierMgr, err := notifier.NewManager(cfg.Notifications)
	require.NoError(t, err)

	detector := &Detector{
		cfg:       cfg,
		formatter: formatter,
		notifier:  notifierMgr,
	}

	alert := &types.DriftAlert{
		Severity:     "high",
		ResourceType: "aws_instance",
		ResourceName: "web",
		ResourceID:   "i-123",
		Attribute:    "instance_type",
		OldValue:     "t3.micro",
		NewValue:     "t3.large",
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	// Should not panic even if notification fails
	assert.NotPanics(t, func() {
		detector.sendAlert(alert)
	})
}

func TestProcessEvents(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{
			{
				Name:              "instance-type-change",
				ResourceTypes:     []string{"aws_instance"},
				WatchedAttributes: []string{"instance_type"},
				Severity:          "high",
			},
		},
		DryRun: true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)

	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
		eventCh:      make(chan types.Event, 10),
	}

	// Start processEvents in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go detector.processEvents(ctx)

	// Send a test event
	testEvent := types.Event{
		Provider:     "aws",
		ResourceType: "aws_instance",
		ResourceID:   "i-0cea65ac652556767",
		EventName:    "ModifyInstanceAttribute",
		Changes: map[string]interface{}{
			"instance_type": "t3.large",
		},
		UserIdentity: types.UserIdentity{
			Type:     "IAMUser",
			UserName: "admin",
		},
	}

	detector.eventCh <- testEvent

	// Give it time to process
	time.Sleep(100 * time.Millisecond)

	// Cancel context to stop processEvents
	cancel()

	// Give it time to shutdown
	time.Sleep(50 * time.Millisecond)

	// Test passed if no panic occurred
	assert.True(t, true)
}

func TestProcessEvents_Cancellation(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{},
		DryRun:     true,
	}

	detector := &Detector{
		cfg:     cfg,
		eventCh: make(chan types.Event, 10),
	}

	// Start processEvents with a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should return quickly without panic
	done := make(chan bool)
	go func() {
		detector.processEvents(ctx)
		done <- true
	}()

	select {
	case <-done:
		// Success - processEvents returned
		assert.True(t, true)
	case <-time.After(1 * time.Second):
		t.Fatal("processEvents did not return after context cancellation")
	}
}

func TestStartCollectors_Integration(t *testing.T) {
	// This is a minimal test since startCollectors calls Falco subscriber
	// which requires a real Falco connection
	cfg := &config.Config{
		Falco: config.FalcoConfig{
			Enabled:  false, // Disabled to avoid connection
			Hostname: "localhost",
			Port:     5060,
		},
	}

	falcoSub, err := falco.NewSubscriber(cfg.Falco)
	require.NoError(t, err)

	detector := &Detector{
		cfg:             cfg,
		falcoSubscriber: falcoSub,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should return error since Falco is not running, but shouldn't panic
	err = detector.startCollectors(ctx)
	// Error is expected since we're not running Falco
	assert.Error(t, err)
}

func TestStart_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/terraform.tfstate",
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false, // Disabled to avoid connection
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
		AutoImport: config.AutoImportConfig{
			Enabled: false,
		},
	}

	detector, err := New(cfg)
	require.NoError(t, err)
	require.NotNil(t, detector)

	// Start with a short timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// Start returns nil (starts goroutines) even if Falco is not running
	// The collector error is logged but doesn't stop the detector
	err = detector.Start(ctx)
	assert.NoError(t, err) // Start() succeeds, errors are logged in goroutines

	// Wait for context timeout
	<-ctx.Done()

	// Give goroutines time to clean up
	time.Sleep(50 * time.Millisecond)
}

func TestStart_StateLoadError(t *testing.T) {
	cfg := &config.Config{
		Providers: config.ProvidersConfig{
			AWS: config.AWSConfig{
				Enabled: true,
				Regions: []string{"us-east-1"},
				State: config.TerraformStateConfig{
					Backend:   "local",
					LocalPath: "testdata/nonexistent.tfstate", // Invalid path
				},
			},
		},
		Falco: config.FalcoConfig{
			Enabled:  false,
			Hostname: "localhost",
			Port:     5060,
		},
		Notifications: config.NotificationsConfig{},
		DriftRules:    []config.DriftRule{},
		DryRun:        true,
	}

	detector, err := New(cfg)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Should return error immediately due to state load failure
	err = detector.Start(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "terraform state")
}

func TestProcessEvents_MultipleEvents(t *testing.T) {
	cfg := &config.Config{
		DriftRules: []config.DriftRule{
			{
				Name:              "test-rule",
				ResourceTypes:     []string{"aws_instance"},
				WatchedAttributes: []string{"instance_type"},
				Severity:          "high",
			},
		},
		DryRun: true,
	}

	stateConfig := config.TerraformStateConfig{
		Backend:   "local",
		LocalPath: "testdata/terraform.tfstate",
	}
	stateManager, err := terraform.NewStateManager(stateConfig)
	require.NoError(t, err)
	err = stateManager.Load(context.Background())
	require.NoError(t, err)

	formatter := diff.NewFormatter(false)

	detector := &Detector{
		cfg:          cfg,
		stateManager: stateManager,
		formatter:    formatter,
		eventCh:      make(chan types.Event, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go detector.processEvents(ctx)

	// Send multiple events
	for i := 0; i < 3; i++ {
		event := types.Event{
			Provider:     "aws",
			ResourceType: "aws_instance",
			ResourceID:   "i-0cea65ac652556767",
			EventName:    "ModifyInstanceAttribute",
			Changes: map[string]interface{}{
				"instance_type": "t3.large",
			},
			UserIdentity: types.UserIdentity{
				Type:     "IAMUser",
				UserName: "admin",
			},
		}
		detector.eventCh <- event
	}

	// Give time to process all events
	time.Sleep(200 * time.Millisecond)

	cancel()
	time.Sleep(50 * time.Millisecond)

	assert.True(t, true) // No panic = success
}
