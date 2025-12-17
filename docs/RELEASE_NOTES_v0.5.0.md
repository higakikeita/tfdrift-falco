# TFDrift-Falco v0.5.0 Release Notes

**Release Date**: December 17, 2025

**Release Type**: Major Feature Release

---

## 🎉 Highlights

### Multi-Cloud Support is Here!

TFDrift-Falco v0.5.0 brings **comprehensive Google Cloud Platform (GCP) support**, enabling real-time drift detection across both AWS and GCP environments **simultaneously**. This is a major milestone in our journey toward true multi-cloud infrastructure drift detection.

### Key Achievements

- ✅ **100+ GCP Events** mapped across 12+ services
- ✅ **GCS Backend** for Terraform state storage
- ✅ **Zero Breaking Changes** - fully backward compatible
- ✅ **Multi-Provider Architecture** ready for future clouds (Azure, etc.)
- ✅ **Production Ready** with comprehensive testing

---

## 🚀 What's New

### 1. GCP Audit Logs Integration

Real-time drift detection for Google Cloud Platform resources using Falco's gcpaudit plugin.

#### Supported GCP Services (12+)

| Service | Events | Key Resources |
|---------|--------|---------------|
| **Compute Engine** | 30+ | Instances, Disks, Machine Types, Metadata, Networks, Firewalls |
| **Cloud Storage** | 15+ | Buckets, Objects, IAM Bindings, ACLs, Lifecycle |
| **Cloud SQL** | 10+ | Instances, Databases, Users, Backups |
| **GKE** | 10+ | Clusters, Node Pools, Workloads |
| **Cloud Run** | 8+ | Services, Revisions, IAM Policies |
| **IAM** | 8+ | Service Accounts, Roles, Bindings, Keys |
| **VPC/Networking** | 10+ | Firewalls, Routes, Subnets, Peering |
| **Cloud Functions** | 5+ | Functions, Triggers, IAM Policies |
| **BigQuery** | 5+ | Datasets, Tables, IAM Policies |
| **Pub/Sub** | 5+ | Topics, Subscriptions, IAM Policies |
| **KMS** | 5+ | Keys, KeyRings, IAM Policies |
| **Secret Manager** | 3+ | Secrets, Versions, IAM Policies |

#### Example Drift Detection Scenarios

**Scenario 1: Compute Instance Metadata Change**
```
Someone adds SSH keys to a GCE instance via Console
    ↓
GCP Audit Log captured by Falco gcpaudit plugin
    ↓
Falco sends event via gRPC to TFDrift-Falco
    ↓
TFDrift-Falco compares with Terraform state
    ↓
Instant Slack alert with user email and metadata changes
```

**Scenario 2: Firewall Rule Modification**
```
Firewall rule source ranges changed manually
    ↓
compute.firewalls.patch event detected
    ↓
Drift detected against Terraform definition
    ↓
Critical severity alert sent to all channels
```

### 2. GCS Backend Support

Load Terraform state files from Google Cloud Storage buckets.

**Features:**
- Application Default Credentials (ADC) support
- Custom credentials file support
- Bucket and prefix configuration
- Automatic error handling and retries

**Configuration Example:**
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - my-project-123
      - my-project-456
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod"
```

### 3. Multi-Provider Architecture

**Intelligent Event Routing:**
- `aws_cloudtrail` events → AWS parser
- `gcpaudit` events → GCP parser
- Extensible design for future providers (Azure, etc.)

**No Breaking Changes:**
- Existing AWS configurations work unchanged
- New GCP fields don't affect AWS events
- Provider-agnostic core fields preserved

### 4. Enhanced Event Types

**New GCP-Specific Fields:**
- `ProjectID` - GCP project identifier
- `ServiceName` - GCP service name (compute.googleapis.com, etc.)
- `Region` - Extracted from zone (us-central1-a → us-central1)

**Preserved AWS Fields:**
- `Region` - AWS region
- `AccountID` - AWS account ID
- All existing AWS user identity fields

---

## 📦 Installation & Upgrade

### Upgrade from v0.4.x

**No breaking changes!** Simply update your binary or Docker image:

```bash
# Binary upgrade
curl -LO https://github.com/keitahigaki/tfdrift-falco/releases/download/v0.5.0/tfdrift-linux-amd64
chmod +x tfdrift-linux-amd64
sudo mv tfdrift-linux-amd64 /usr/local/bin/tfdrift

# Docker upgrade
docker pull ghcr.io/higakikeita/tfdrift-falco:v0.5.0
```

### New GCP Setup

**Step 1: Enable GCP Audit Logs**
```bash
gcloud projects add-iam-policy-binding my-project-123 \
  --member="serviceAccount:falco-sa@my-project-123.iam.gserviceaccount.com" \
  --role="roles/logging.viewer"
```

**Step 2: Configure Falco gcpaudit Plugin**

See comprehensive setup guide: [docs/gcp-setup.md](../gcp-setup.md)

**Step 3: Update TFDrift-Falco Configuration**
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - my-project-123
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod"

drift_rules:
  - name: "GCP Compute Instance Modification"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
    severity: "high"
```

**Step 4: Start Monitoring**
```bash
tfdrift --config config.yaml
```

---

## 🧪 Testing & Quality

### Test Coverage

- **34 GCP-specific tests** covering all parser functionality
- **Integration tests** for multi-provider scenarios
- **Resource type mapping validation** for 100+ events
- **100% test pass rate**

### Tested Scenarios

✅ GCP Audit Log event parsing
✅ Multi-provider event routing (AWS + GCP)
✅ GCS backend state loading
✅ Resource type mapping for all services
✅ User identity extraction
✅ Change tracking and correlation

---

## 📚 Documentation

### New Documentation

- **[GCP Setup Guide](../gcp-setup.md)** - Complete setup instructions (500+ lines)
  - Prerequisites and architecture
  - Falco gcpaudit plugin configuration
  - GCP Audit Logs and Pub/Sub setup
  - TFDrift-Falco configuration
  - Troubleshooting (5 common issues)
  - Advanced configuration
  - Security best practices

- **[GCP Configuration Example](../examples/config-gcp.yaml)** - Production-ready config template

### Updated Documentation

- **README.md** - Updated with GCP support announcements
- **CHANGELOG.md** - Comprehensive v0.5.0 changes
- **Architecture diagrams** - Updated to show GCP integration

---

## 🔄 Migration Guide

### From v0.4.x to v0.5.0

**No breaking changes!** Your existing AWS configurations will continue to work without modifications.

#### Option 1: AWS Only (No Changes Needed)

Keep your existing configuration - everything works as before.

#### Option 2: Add GCP Support

Add GCP configuration alongside existing AWS config:

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "prod/terraform.tfstate"

  gcp:  # NEW - Add this section
    enabled: true
    projects:
      - my-project-123
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod"
```

#### Option 3: GCP Only

Start fresh with GCP-only configuration:

```yaml
providers:
  gcp:
    enabled: true
    projects:
      - my-project-123
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  - name: "GCP Compute Instance Modification"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
      - "deletion_protection"
    severity: "high"
```

---

## 🐛 Known Limitations

### GCP-Specific Limitations

1. **New Feature** - GCP support is new, production validation recommended
2. **Audit Log Latency** - GCP Audit Logs have 30 seconds to 5 minutes delivery latency via Pub/Sub
3. **Multi-Project** - Multi-project environments require per-project configuration
4. **Coverage** - Some advanced GCP features may not be fully covered yet

### General Limitations

1. **Large Scale** - Environments with 50,000+ resources require performance tuning
2. **Multi-Account** - AWS multi-account setups need additional validation
3. **CloudTrail Latency** - AWS CloudTrail has 5-15 minutes latency (S3), 1-5 minutes (SQS)

See [Production Readiness Guide](PRODUCTION_READINESS.md) for comprehensive limitations and best practices.

---

## 🔐 Security Considerations

### GCP Credentials

**Recommended: Application Default Credentials (ADC)**
```bash
gcloud auth application-default login
```

**Alternative: Service Account Key**
```yaml
providers:
  gcp:
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      credentials_file: "/path/to/service-account-key.json"  # Optional
```

### Least Privilege IAM

**Minimum Required Permissions:**
- `roles/storage.objectViewer` - Read Terraform state from GCS
- `roles/logging.viewer` - Read Audit Logs (for Falco plugin)

**Recommended Service Account:**
```bash
gcloud iam service-accounts create tfdrift-falco \
  --display-name="TFDrift-Falco Service Account"

gcloud projects add-iam-policy-binding my-project-123 \
  --member="serviceAccount:tfdrift-falco@my-project-123.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

gcloud projects add-iam-policy-binding my-project-123 \
  --member="serviceAccount:tfdrift-falco@my-project-123.iam.gserviceaccount.com" \
  --role="roles/logging.viewer"
```

---

## 🎯 Use Cases

### 1. Multi-Cloud Drift Detection

Monitor infrastructure drift across AWS and GCP simultaneously.

**Example:**
- AWS EC2 instances managed by Terraform
- GCP Compute Engine instances managed by Terraform
- Single TFDrift-Falco instance monitoring both
- Unified Slack alerts for all drift events

### 2. GCP-Only Environments

Organizations using only GCP can now benefit from real-time drift detection.

**Example:**
- 100% GCP infrastructure
- Terraform managing GKE clusters, Cloud SQL, Cloud Run
- Real-time alerts on console-based changes
- Compliance enforcement for Infrastructure-as-Code

### 3. Hybrid Cloud Security

Ensure all infrastructure changes follow IaC workflows regardless of cloud provider.

**Example:**
- Multi-cloud environment (AWS + GCP)
- Centralized security monitoring
- Consistent drift detection policies
- Unified audit trail across clouds

---

## 📊 Technical Architecture

### Event Flow

```
GCP Console/API Change
    ↓
GCP Audit Logs
    ↓
Cloud Pub/Sub
    ↓
Falco gcpaudit Plugin
    ↓
Falco Rules Engine
    ↓
Falco gRPC Output
    ↓
TFDrift-Falco Subscriber (Event Router)
    ↓
┌─────────────┬──────────────┐
│ AWS Parser  │  GCP Parser  │
└─────────────┴──────────────┘
    ↓
Drift Detection Engine
    ↓
Terraform State Comparison (GCS/S3/Local)
    ↓
Notification Channels (Slack/Webhook/etc.)
```

### Code Architecture

```
pkg/
├── gcp/                    # NEW - GCP support
│   ├── audit_parser.go     # Parse GCP Audit Log events
│   ├── resource_mapper.go  # Map events to Terraform resources
│   └── *_test.go           # Comprehensive tests
├── terraform/backend/
│   ├── gcs.go              # NEW - GCS backend
│   └── gcs_test.go
├── falco/
│   ├── subscriber.go       # UPDATED - GCP parser initialization
│   └── event_parser.go     # UPDATED - Multi-provider routing
├── types/
│   └── types.go            # UPDATED - GCP-specific fields
└── config/
    └── config.go           # UPDATED - GCP configuration
```

---

## 🙏 Acknowledgments

This release was made possible by:

- **Falco Community** - For the excellent gcpaudit plugin
- **Google Cloud** - For comprehensive Audit Logs documentation
- **Terraform Community** - For GCS backend specifications
- **Our Users** - For feature requests and feedback

---

## 🔗 Resources

### Documentation
- [GCP Setup Guide](../gcp-setup.md)
- [AWS Setup Guide](../falco-setup.md)
- [Production Readiness Guide](PRODUCTION_READINESS.md)
- [Architecture Overview](architecture.md)

### Examples
- [GCP Configuration Example](../examples/config-gcp.yaml)
- [AWS Configuration Example](../examples/config.yaml)
- [Multi-Cloud Configuration](../examples/config-multi-cloud.yaml)

### Community
- [GitHub Repository](https://github.com/keitahigaki/tfdrift-falco)
- [Issue Tracker](https://github.com/keitahigaki/tfdrift-falco/issues)
- [Discussions](https://github.com/keitahigaki/tfdrift-falco/discussions)

---

## 📞 Support

### Getting Help

- 📖 **Documentation**: Start with [GCP Setup Guide](../gcp-setup.md)
- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/keitahigaki/tfdrift-falco/issues)
- 💬 **Questions**: [GitHub Discussions](https://github.com/keitahigaki/tfdrift-falco/discussions)
- 📧 **Security Issues**: security@example.com

### Troubleshooting

Common issues and solutions are documented in:
- [GCP Setup Guide - Troubleshooting Section](../gcp-setup.md#troubleshooting)
- [Production Readiness Guide](PRODUCTION_READINESS.md)

---

## 🚀 What's Next?

### Roadmap

**Phase 3: Advanced Features**
- [ ] Web dashboard UI
- [ ] Azure Activity Logs support
- [ ] Machine learning-based anomaly detection
- [ ] Auto-remediation actions
- [ ] Policy-as-Code integration (OPA/Rego)

**Phase 4: Enterprise Features**
- [ ] Multi-account/multi-org support
- [ ] RBAC and team management
- [ ] Compliance reporting (SOC2, PCI-DSS, HIPAA)
- [ ] Integration marketplace

See [Roadmap](../README.md#roadmap) for detailed plans.

---

## 📊 Statistics

### Code Changes

- **Files Added**: 6 (audit_parser.go, resource_mapper.go, gcs.go, + tests)
- **Files Modified**: 6 (subscriber.go, event_parser.go, types.go, config.go, factory.go, README.md)
- **Lines Added**: ~2,000
- **Test Coverage**: 34 new tests, 100% pass rate
- **Documentation**: 500+ lines of new documentation

### Event Coverage

| Provider | Events | Services | Status |
|----------|--------|----------|--------|
| AWS | 203 | 19 | ✅ Stable (v0.3.0) |
| GCP | 100+ | 12+ | ✅ New (v0.5.0) |
| Azure | - | - | 🚧 Planned |

---

**Thank you for using TFDrift-Falco!** 🎉

We're excited to bring multi-cloud drift detection to the community. Please share your feedback and help us make TFDrift-Falco even better.

---

*Made with ❤️ by the TFDrift-Falco Team*

*Follow us: [Twitter](https://x.com/keitah0322) | [GitHub](https://github.com/keitahigaki)*
