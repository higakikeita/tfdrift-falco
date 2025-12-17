# GCP Support Implementation Design

**Version**: 1.0.0
**Status**: Draft
**Created**: 2025-12-17
**Author**: Keita Higaki

---

## 1. Overview

This document outlines the design and implementation plan for adding Google Cloud Platform (GCP) support to TFDrift-Falco.

### Goals

- Enable real-time drift detection for GCP resources managed by Terraform
- Support GCS backend for Terraform state
- Integrate with Falco's GCP Audit Logs plugin
- Maintain consistency with existing AWS implementation patterns

### Non-Goals (Future Work)

- Multi-project aggregation (Phase 2)
- GCP Organization policy integration
- Cloud Asset Inventory integration

---

## 2. Technical Stack

### Required Dependencies

```go
// GCP SDK
"cloud.google.com/go/storage"           // GCS client
"google.golang.org/api/option"          // Auth options
"google.golang.org/api/cloudresourcemanager/v1" // Project API

// Existing (already in project)
"github.com/falcosecurity/client-go"    // Falco gRPC client
"github.com/spf13/viper"                // Config management
```

### Falco Plugin

**GCP Audit Logs Plugin**:
- Repository: https://github.com/falcosecurity/plugins/tree/main/plugins/gcpaudit
- Status: Official Falco plugin (maintained by Falco Security)
- Input: GCP Audit Logs via Pub/Sub or Cloud Storage

---

## 3. Architecture

### Data Flow

```
GCP Audit Logs
    ↓
Cloud Pub/Sub
    ↓
Falco (gcpaudit plugin)
    ↓
Falco gRPC Output
    ↓
TFDrift-Falco Subscriber
    ↓
GCP Event Parser → Resource Mapper
    ↓
Drift Detector
    ↓
Notification Manager
```

### Component Diagram

```
TFDrift-Falco
├── pkg/
│   ├── gcp/                     # NEW: GCP-specific code
│   │   ├── audit_parser.go      # Parse GCP Audit Log events
│   │   ├── resource_mapper.go   # Map GCP events → Terraform resources
│   │   └── change_extractor.go  # Extract attribute changes
│   ├── terraform/
│   │   └── backend/
│   │       └── gcs_backend.go   # NEW: GCS state backend
│   ├── config/
│   │   └── config.go            # MODIFY: Add GCP config
│   └── detector/
│       └── detector.go          # MODIFY: Support GCP resources
```

---

## 4. GCP Services Coverage

### Phase 1: Core Compute & Networking (MVP)

| Service | Terraform Resource | Priority | Audit Log Events |
|---------|-------------------|----------|------------------|
| **Compute Engine** | google_compute_instance | Critical | compute.instances.insert, delete, setMetadata, setLabels |
| **VPC** | google_compute_network | Critical | compute.networks.insert, delete, patch |
| **Firewall** | google_compute_firewall | Critical | compute.firewalls.insert, delete, update, patch |
| **Subnets** | google_compute_subnetwork | High | compute.subnetworks.insert, delete, patch |
| **IAM** | google_project_iam_binding | Critical | SetIamPolicy |

### Phase 2: Storage & Data

| Service | Terraform Resource | Priority | Audit Log Events |
|---------|-------------------|----------|------------------|
| **Cloud Storage** | google_storage_bucket | High | storage.buckets.create, delete, update |
| **Cloud SQL** | google_sql_database_instance | High | cloudsql.instances.create, delete, update |
| **Persistent Disk** | google_compute_disk | Medium | compute.disks.insert, delete |

### Phase 3: Kubernetes & Serverless

| Service | Terraform Resource | Priority | Audit Log Events |
|---------|-------------------|----------|------------------|
| **GKE** | google_container_cluster | High | container.clusters.create, delete, update |
| **Cloud Run** | google_cloud_run_service | Medium | run.services.create, delete, update |
| **Cloud Functions** | google_cloudfunctions_function | Medium | cloudfunctions.functions.create, delete |

**Total Events (Phase 1)**: ~30 events across 5 services

---

## 5. File Structure

```
tfdrift-falco/
├── pkg/
│   ├── gcp/
│   │   ├── audit_parser.go          # NEW: GCP Audit Log parser
│   │   ├── audit_parser_test.go
│   │   ├── resource_mapper.go       # NEW: Event → Resource mapping
│   │   ├── resource_mapper_test.go
│   │   ├── change_extractor.go      # NEW: Extract changes from events
│   │   └── change_extractor_test.go
│   ├── terraform/backend/
│   │   ├── gcs_backend.go           # NEW: GCS backend implementation
│   │   ├── gcs_backend_test.go
│   │   └── factory.go               # MODIFY: Add GCS case
│   ├── config/
│   │   └── config.go                # MODIFY: Add GCPConfig struct
│   └── detector/
│       └── event_handler.go         # MODIFY: Handle GCP events
├── examples/
│   └── config-gcp.yaml              # NEW: GCP example config
├── docs/
│   ├── GCP_IMPLEMENTATION_DESIGN.md # This document
│   ├── gcp-setup.md                 # NEW: Setup guide
│   └── GCP_RESOURCE_COVERAGE.md     # NEW: Coverage analysis
└── tests/
    └── integration/
        └── gcp_test.go              # NEW: Integration tests
```

---

## 6. Detailed Component Design

### 6.1 GCS Backend (`pkg/terraform/backend/gcs_backend.go`)

**Purpose**: Read Terraform state from Google Cloud Storage

**Interface Implementation**:
```go
type GCSBackend struct {
    client     *storage.Client
    bucketName string
    objectKey  string
    projectID  string
}

func (b *GCSBackend) Load(ctx context.Context) ([]byte, error)
func (b *GCSBackend) Name() string
```

**Authentication**:
- Default: Application Default Credentials (ADC)
- Support: Service Account Key JSON file
- Environment: `GOOGLE_APPLICATION_CREDENTIALS`

**Configuration**:
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "my-gcp-project-id"
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod/terraform.tfstate"
```

**Error Handling**:
- Bucket not found → log error, skip GCP monitoring
- Permission denied → log auth error with troubleshooting link
- Network timeout → retry with exponential backoff

---

### 6.2 GCP Audit Log Parser (`pkg/gcp/audit_parser.go`)

**Purpose**: Parse Falco events containing GCP Audit Logs

**Key Functions**:
```go
type AuditParser struct {
    logger *logrus.Logger
}

// Parse Falco event → GCP Audit Log event
func (p *AuditParser) Parse(falcoOutput *client.Response) (*types.Event, error)

// Extract resource information
func (p *AuditParser) extractResource(output *client.Response) (string, string, error)

// Extract user identity
func (p *AuditParser) extractUserIdentity(output *client.Response) string

// Filter relevant events
func (p *AuditParser) isRelevantEvent(methodName string) bool
```

**Input (Falco Output Fields)**:
```json
{
  "output": "GCP Audit Log: compute.instances.setMetadata",
  "output_fields": {
    "gcp.methodName": "compute.instances.setMetadata",
    "gcp.serviceName": "compute.googleapis.com",
    "gcp.resource.name": "projects/123/zones/us-central1-a/instances/vm-1",
    "gcp.authenticationInfo.principalEmail": "user@example.com",
    "gcp.request": "{\"metadata\":{\"items\":[...]}}",
    "gcp.response": "{...}"
  }
}
```

**Output (TFDrift Event)**:
```go
types.Event{
    Provider:     "gcp",
    ResourceType: "google_compute_instance",
    ResourceID:   "vm-1",
    Region:       "us-central1-a",
    ProjectID:    "my-project-123",
    EventName:    "compute.instances.setMetadata",
    User:         "user@example.com",
    Timestamp:    time.Now(),
    Changes: map[string]interface{}{
        "metadata": map[string]string{
            "ssh-keys": "new-value",
        },
    },
}
```

---

### 6.3 Resource Mapper (`pkg/gcp/resource_mapper.go`)

**Purpose**: Map GCP Audit Log method names to Terraform resource types

**Mapping Table**:
```go
var eventToResourceMap = map[string]string{
    // Compute Engine
    "compute.instances.insert":      "google_compute_instance",
    "compute.instances.delete":      "google_compute_instance",
    "compute.instances.setMetadata": "google_compute_instance",
    "compute.instances.setLabels":   "google_compute_instance",

    // VPC
    "compute.networks.insert": "google_compute_network",
    "compute.networks.delete": "google_compute_network",
    "compute.networks.patch":  "google_compute_network",

    // Firewall
    "compute.firewalls.insert": "google_compute_firewall",
    "compute.firewalls.delete": "google_compute_firewall",
    "compute.firewalls.update": "google_compute_firewall",

    // IAM
    "SetIamPolicy": "google_project_iam_binding",

    // Cloud Storage
    "storage.buckets.create": "google_storage_bucket",
    "storage.buckets.update": "google_storage_bucket",

    // Cloud SQL
    "cloudsql.instances.create": "google_sql_database_instance",
    "cloudsql.instances.update": "google_sql_database_instance",
}
```

**Functions**:
```go
func MapEventToResource(methodName string) (string, bool)
func ExtractResourceID(resourceName string) string
func ExtractZone(resourceName string) string
func ExtractProjectID(resourceName string) string
```

---

### 6.4 Change Extractor (`pkg/gcp/change_extractor.go`)

**Purpose**: Extract attribute changes from GCP Audit Log request/response

**Per-Event Extraction Logic**:

```go
func (e *ChangeExtractor) Extract(methodName string, request, response map[string]interface{}) map[string]interface{}

// Example: compute.instances.setMetadata
func (e *ChangeExtractor) extractMetadataChanges(request map[string]interface{}) map[string]interface{} {
    changes := make(map[string]interface{})

    if metadata, ok := request["metadata"].(map[string]interface{}); ok {
        if items, ok := metadata["items"].([]interface{}); ok {
            for _, item := range items {
                m := item.(map[string]interface{})
                key := m["key"].(string)
                value := m["value"].(string)
                changes[fmt.Sprintf("metadata.%s", key)] = value
            }
        }
    }

    return changes
}
```

---

### 6.5 Configuration (`pkg/config/config.go`)

**Add GCP Config Struct**:

```go
type GCPConfig struct {
    Enabled  bool     `mapstructure:"enabled"`
    Projects []string `mapstructure:"projects"`
    State    StateConfig `mapstructure:"state"`
}

type StateConfig struct {
    Backend    string `mapstructure:"backend"`    // "gcs", "local"
    GCSBucket  string `mapstructure:"gcs_bucket"`
    GCSPrefix  string `mapstructure:"gcs_prefix"`
    LocalPath  string `mapstructure:"local_path"`
}

type ProvidersConfig struct {
    AWS AWSConfig `mapstructure:"aws"`
    GCP GCPConfig `mapstructure:"gcp"` // NEW
}
```

**Example Config**:
```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: my-aws-state
      s3_key: terraform.tfstate

  gcp:
    enabled: true
    projects:
      - my-gcp-project-123
    state:
      backend: gcs
      gcs_bucket: my-gcp-state
      gcs_prefix: prod/terraform.tfstate
```

---

## 7. Implementation Steps

### Step 1: GCS Backend (Week 1)
- [ ] Implement `GCSBackend` struct
- [ ] Add GCP authentication (ADC)
- [ ] Implement `Load()` method with error handling
- [ ] Add retry logic for transient errors
- [ ] Write unit tests (mock GCS client)
- [ ] Update backend factory

### Step 2: Configuration Extension (Week 1)
- [ ] Add `GCPConfig` struct to config
- [ ] Update config parser
- [ ] Add validation for GCP config
- [ ] Create example `config-gcp.yaml`
- [ ] Write config tests

### Step 3: GCP Audit Log Parser (Week 2)
- [ ] Implement `AuditParser` struct
- [ ] Parse Falco output fields (gcp.*)
- [ ] Extract resource information
- [ ] Extract user identity
- [ ] Write parser tests with sample events
- [ ] Document supported event fields

### Step 4: Resource Mapper (Week 2)
- [ ] Create event → resource mapping table (Phase 1: 30 events)
- [ ] Implement `MapEventToResource()`
- [ ] Implement resource ID extraction helpers
- [ ] Write mapper tests
- [ ] Document coverage analysis

### Step 5: Change Extractor (Week 2-3)
- [ ] Implement base `ChangeExtractor`
- [ ] Add event-specific extraction logic
  - [ ] Compute Engine events
  - [ ] VPC events
  - [ ] Firewall events
  - [ ] IAM events
  - [ ] Cloud Storage events
- [ ] Write extraction tests
- [ ] Handle edge cases (empty request/response)

### Step 6: Integration (Week 3)
- [ ] Update detector to support GCP events
- [ ] Modify event handler for GCP resources
- [ ] Add GCP-specific drift rules
- [ ] Test end-to-end flow
- [ ] Fix integration issues

### Step 7: Testing & Documentation (Week 3)
- [ ] Write integration tests
- [ ] Manual testing with real GCP project
- [ ] Write GCP setup guide
- [ ] Update README with GCP support
- [ ] Create demo video/screenshots

---

## 8. Test Plan

### Unit Tests

**Coverage Target**: 80%+

```
pkg/gcp/audit_parser_test.go
- TestParse_ValidEvent
- TestParse_InvalidEvent
- TestExtractResource_ComputeInstance
- TestExtractUserIdentity

pkg/gcp/resource_mapper_test.go
- TestMapEventToResource_AllSupportedEvents
- TestExtractResourceID
- TestExtractProjectID

pkg/terraform/backend/gcs_backend_test.go
- TestLoad_Success
- TestLoad_BucketNotFound
- TestLoad_PermissionDenied
- TestLoad_NetworkError
```

### Integration Tests

```
tests/integration/gcp_test.go
- TestGCPDriftDetection_ComputeInstanceMetadata
- TestGCPDriftDetection_FirewallRule
- TestGCPDriftDetection_IAMBinding
- TestGCSBackend_LoadState
```

### Manual Testing Checklist

- [ ] Deploy Falco with gcpaudit plugin
- [ ] Configure GCP Audit Logs → Pub/Sub
- [ ] Deploy TFDrift-Falco with GCP config
- [ ] Create Terraform-managed GCP resources
- [ ] Manually modify resource via Console
- [ ] Verify drift detected in real-time
- [ ] Check Slack notification
- [ ] Verify JSON output format

---

## 9. Success Criteria

### Functional Requirements

✅ **Must Have (MVP)**:
- GCS backend successfully loads Terraform state
- Parse GCP Audit Logs from Falco
- Map 30+ GCP events to Terraform resources
- Detect drift for Compute Engine instances
- Detect drift for VPC firewall rules
- Detect drift for IAM bindings
- Send notifications (Slack/Discord/Webhook)

✅ **Should Have**:
- Support 50+ GCP events across 8 services
- JSON output mode for GCP events
- Auto-detection for GCS backend

⚠️ **Could Have (Future)**:
- Multi-project support
- GCP Organization policies
- Cloud Asset Inventory integration

### Non-Functional Requirements

- **Performance**: Process GCP events in <500ms
- **Reliability**: Handle network errors gracefully
- **Security**: Use ADC (no hardcoded credentials)
- **Test Coverage**: 80%+ for new GCP code
- **Documentation**: Complete setup guide

---

## 10. Risks & Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Falco gcpaudit plugin instability | High | Low | Test with stable Falco version; document known issues |
| GCP Audit Log field changes | Medium | Low | Version detection; graceful degradation |
| Authentication complexity | Medium | Medium | Comprehensive docs; support multiple auth methods |
| Resource mapping incompleteness | Medium | High | Phased rollout; document coverage gaps |
| GCS rate limiting | Low | Medium | Implement exponential backoff; caching |

---

## 11. Future Enhancements (Post-MVP)

### Phase 2: Advanced GCP Features
- Multi-project aggregation
- GCP Organization-level monitoring
- Cloud Asset Inventory integration
- Service Account impersonation

### Phase 3: GCP-Specific Features
- GKE cluster drift detection
- Cloud Run service drift
- BigQuery dataset/table drift
- Cloud Functions configuration drift

---

## 12. References

### Official Documentation
- [Falco GCP Audit Plugin](https://github.com/falcosecurity/plugins/tree/main/plugins/gcpaudit)
- [GCP Audit Logs](https://cloud.google.com/logging/docs/audit)
- [Terraform GCP Provider](https://registry.terraform.io/providers/hashicorp/google/latest/docs)
- [GCS Go Client](https://pkg.go.dev/cloud.google.com/go/storage)

### Similar Projects
- [driftctl GCP support](https://docs.driftctl.com/latest/providers/google/)
- [CloudQuery GCP](https://www.cloudquery.io/docs/plugins/sources/gcp/overview)

---

## Appendix A: GCP Audit Log Event Example

```json
{
  "protoPayload": {
    "@type": "type.googleapis.com/google.cloud.audit.AuditLog",
    "authenticationInfo": {
      "principalEmail": "user@example.com"
    },
    "methodName": "compute.instances.setMetadata",
    "resourceName": "projects/my-project/zones/us-central1-a/instances/vm-1",
    "serviceName": "compute.googleapis.com",
    "request": {
      "@type": "type.googleapis.com/compute.instances.setMetadata",
      "metadata": {
        "items": [
          {
            "key": "ssh-keys",
            "value": "user:ssh-rsa AAAA..."
          }
        ]
      }
    }
  }
}
```

---

## Appendix B: Terraform State Example (GCP)

```json
{
  "version": 4,
  "terraform_version": "1.5.0",
  "resources": [
    {
      "mode": "managed",
      "type": "google_compute_instance",
      "name": "vm-1",
      "provider": "provider[\"registry.terraform.io/hashicorp/google\"]",
      "instances": [
        {
          "attributes": {
            "id": "projects/my-project/zones/us-central1-a/instances/vm-1",
            "name": "vm-1",
            "zone": "us-central1-a",
            "metadata": {
              "ssh-keys": "user:ssh-rsa ORIGINAL..."
            }
          }
        }
      ]
    }
  ]
}
```

---

**End of Design Document**
