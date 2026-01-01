package falco

import (
	"testing"

	"github.com/falcosecurity/client-go/pkg/api/outputs"
	"github.com/keitahigaki/tfdrift-falco/pkg/config"
	"github.com/stretchr/testify/assert"
)

var testCfg = config.FalcoConfig{
	Enabled:  true,
	Hostname: "localhost",
	Port:     5060,
}

// TestParseFalcoOutput_MultiProvider tests parsing events from both AWS and GCP
func TestParseFalcoOutput_MultiProvider(t *testing.T) {
	subscriber, err := NewSubscriber(testCfg)
	assert.NoError(t, err)
	assert.NotNil(t, subscriber)
	assert.NotNil(t, subscriber.gcpParser, "GCP parser should be initialized")

	tests := []struct {
		name         string
		response     *outputs.Response
		wantProvider string
		wantNil      bool
	}{
		{
			name: "AWS CloudTrail Event",
			response: &outputs.Response{
				Source: "aws_cloudtrail",
				OutputFields: map[string]string{
					"ct.name":               "ModifyInstanceAttribute",
					"ct.request.instanceid": "i-1234567890abcdef0",
					"ct.user":               "admin",
					"ct.user.type":          "IAMUser",
					"ct.user.accountid":     "123456789012",
					"ct.user.principalid":   "AIDAI...",
					"ct.user.arn":           "arn:aws:iam::123456789012:user/admin",
					"ct.request.attribute":  "disableApiTermination",
					"ct.request.value":      "true",
				},
			},
			wantProvider: "aws",
			wantNil:      false,
		},
		{
			name: "GCP Audit Log Event",
			response: &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":                        "compute.instances.setMetadata",
					"gcp.resource.name":                     "projects/my-project-123/zones/us-central1-a/instances/vm-1",
					"gcp.serviceName":                       "compute.googleapis.com",
					"gcp.authenticationInfo.principalEmail": "user@example.com",
					"gcp.request":                           `{"metadata": {"items": [{"key": "ssh-keys"}]}}`,
				},
			},
			wantProvider: "gcp",
			wantNil:      false,
		},
		{
			name: "Unknown Source",
			response: &outputs.Response{
				Source: "unknown_source",
				OutputFields: map[string]string{
					"some.field": "some-value",
				},
			},
			wantNil: true,
		},
		{
			name: "GCP Irrelevant Event",
			response: &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":    "storage.objects.get",
					"gcp.resource.name": "projects/my-project/buckets/my-bucket/objects/file.txt",
				},
			},
			wantNil: true,
		},
		{
			name: "AWS Irrelevant Event",
			response: &outputs.Response{
				Source: "aws_cloudtrail",
				OutputFields: map[string]string{
					"ct.name": "GetObject",
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := subscriber.parseFalcoOutput(tt.response)

			if tt.wantNil {
				assert.Nil(t, event, "Expected nil event")
			} else {
				assert.NotNil(t, event, "Expected non-nil event")
				assert.Equal(t, tt.wantProvider, event.Provider)
				assert.NotEmpty(t, event.EventName)
				assert.NotEmpty(t, event.ResourceType)
				assert.NotEmpty(t, event.ResourceID)
			}
		})
	}
}

// TestParseFalcoOutput_GCPSpecificFields tests GCP-specific fields in events
func TestParseFalcoOutput_GCPSpecificFields(t *testing.T) {
	subscriber, err := NewSubscriber(testCfg)
	assert.NoError(t, err)

	res := &outputs.Response{
		Source: "gcpaudit",
		OutputFields: map[string]string{
			"gcp.methodName":                        "compute.instances.setMetadata",
			"gcp.resource.name":                     "projects/my-project-123/zones/us-central1-a/instances/vm-1",
			"gcp.serviceName":                       "compute.googleapis.com",
			"gcp.authenticationInfo.principalEmail": "user@example.com",
			"gcp.request":                           `{"metadata": "test"}`,
		},
	}

	event := subscriber.parseFalcoOutput(res)

	assert.NotNil(t, event)
	assert.Equal(t, "gcp", event.Provider)
	assert.Equal(t, "my-project-123", event.ProjectID, "ProjectID should be extracted")
	assert.Equal(t, "us-central1", event.Region, "Region should be extracted from zone")
	assert.Equal(t, "compute.googleapis.com", event.ServiceName, "ServiceName should be set")
	assert.Equal(t, "user@example.com", event.UserIdentity.UserName, "User email should be set")
}

// TestParseFalcoOutput_AWSSpecificFields tests AWS-specific fields are preserved
func TestParseFalcoOutput_AWSSpecificFields(t *testing.T) {
	subscriber, err := NewSubscriber(testCfg)
	assert.NoError(t, err)

	res := &outputs.Response{
		Source: "aws_cloudtrail",
		OutputFields: map[string]string{
			"ct.name":               "ModifyInstanceAttribute",
			"ct.request.instanceid": "i-1234567890abcdef0",
			"ct.user":               "admin",
			"ct.user.type":          "IAMUser",
			"ct.user.accountid":     "123456789012",
			"ct.user.principalid":   "AIDAI123456",
			"ct.user.arn":           "arn:aws:iam::123456789012:user/admin",
			"ct.region":             "us-east-1",
		},
	}

	event := subscriber.parseFalcoOutput(res)

	assert.NotNil(t, event)
	assert.Equal(t, "aws", event.Provider)
	assert.Equal(t, "IAMUser", event.UserIdentity.Type)
	assert.Equal(t, "AIDAI123456", event.UserIdentity.PrincipalID)
	assert.Equal(t, "arn:aws:iam::123456789012:user/admin", event.UserIdentity.ARN)
	assert.Equal(t, "123456789012", event.UserIdentity.AccountID)
}

// TestSubscriberInitialization tests subscriber initialization with GCP parser
func TestSubscriberInitialization(t *testing.T) {
	tests := []struct {
		name    string
		config  config.FalcoConfig
		wantErr bool
	}{
		{
			name: "Valid Configuration",
			config: config.FalcoConfig{
				Enabled:  true,
				Hostname: "localhost",
				Port:     5060,
			},
			wantErr: false,
		},
		{
			name: "With TLS Configuration",
			config: config.FalcoConfig{
				Enabled:    true,
				Hostname:   "falco.example.com",
				Port:       5060,
				CertFile:   "/path/to/cert.pem",
				KeyFile:    "/path/to/key.pem",
				CARootFile: "/path/to/ca.pem",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subscriber, err := NewSubscriber(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, subscriber)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, subscriber)
				assert.NotNil(t, subscriber.gcpParser, "GCP parser must be initialized")
				assert.Equal(t, tt.config, subscriber.cfg)
			}
		})
	}
}

// TestParseFalcoOutput_GCPResourceTypes tests various GCP resource type mappings
func TestParseFalcoOutput_GCPResourceTypes(t *testing.T) {
	subscriber, err := NewSubscriber(testCfg)
	assert.NoError(t, err)

	tests := []struct {
		name             string
		methodName       string
		resourceName     string
		wantResourceType string
	}{
		{
			"Compute Instance",
			"compute.instances.setMetadata",
			"projects/proj-1/zones/us-central1-a/instances/vm-1",
			"google_compute_instance",
		},
		{
			"Firewall Rule",
			"compute.firewalls.insert",
			"projects/proj-1/global/firewalls/allow-ssh",
			"google_compute_firewall",
		},
		{
			"Cloud Storage Bucket",
			"storage.buckets.create",
			"projects/_/buckets/my-bucket",
			"google_storage_bucket",
		},
		{
			"Cloud SQL Instance",
			"cloudsql.instances.create",
			"projects/proj-1/instances/db-1",
			"google_sql_database_instance",
		},
		{
			"GKE Cluster",
			"container.clusters.create",
			"projects/proj-1/zones/us-central1-a/clusters/cluster-1",
			"google_container_cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &outputs.Response{
				Source: "gcpaudit",
				OutputFields: map[string]string{
					"gcp.methodName":                        tt.methodName,
					"gcp.resource.name":                     tt.resourceName,
					"gcp.serviceName":                       "compute.googleapis.com",
					"gcp.authenticationInfo.principalEmail": "user@example.com",
				},
			}

			event := subscriber.parseFalcoOutput(res)

			assert.NotNil(t, event, "Event should not be nil for %s", tt.methodName)
			assert.Equal(t, "gcp", event.Provider)
			assert.Equal(t, tt.wantResourceType, event.ResourceType)
		})
	}
}
