# GCP Services Coverage

> **Version:** v0.5.0+
> **Status:** Production Ready
> **Total Events:** 100+
> **Total Services:** 12+

TFDrift-Falco v0.5.0 introduces comprehensive Google Cloud Platform (GCP) support, enabling real-time drift detection for GCP resources through Falco's gcpaudit plugin.

---

## Overview

TFDrift-Falco monitors GCP Audit Logs to detect infrastructure drift in real-time. The system:

1. **Receives** GCP Audit Log events via Falco's gcpaudit plugin
2. **Parses** events to extract resource changes
3. **Compares** changes against Terraform state (GCS, S3, or local)
4. **Alerts** on detected drift via configured channels (Slack, Discord, webhooks)

---

## Supported Services

### Compute (40+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **Compute Engine** | 11 | `google_compute_instance` | ✅ Production |
| **Disks** | 4 | `google_compute_disk` | ✅ Production |
| **Networks** | 3 | `google_compute_network` | ✅ Production |
| **Subnetworks** | 4 | `google_compute_subnetwork` | ✅ Production |
| **Firewalls** | 4 | `google_compute_firewall` | ✅ Production |
| **Routes** | 2 | `google_compute_route` | ✅ Production |
| **Routers** | 4 | `google_compute_router` | ✅ Production |
| **VPN** | 4 | `google_compute_vpn_tunnel`, `google_compute_vpn_gateway` | ✅ Production |
| **Load Balancers** | 12 | `google_compute_backend_service`, `google_compute_health_check` | ✅ Production |

[Learn more →](compute-engine.md)

### Storage (5+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **Cloud Storage** | 3 | `google_storage_bucket` | ✅ Production |
| **Storage Objects** | 2 | `google_storage_bucket_object` | ✅ Production |
| **Storage IAM** | 1 | `google_storage_bucket_iam_binding` | ✅ Production |

[Learn more →](cloud-storage.md)

### Databases (10+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **Cloud SQL Instances** | 4 | `google_sql_database_instance` | ✅ Production |
| **Cloud SQL Databases** | 3 | `google_sql_database` | ✅ Production |
| **Cloud SQL Users** | 3 | `google_sql_user` | ✅ Production |

[Learn more →](cloud-sql.md)

### Security (12+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **IAM Project** | 1 | `google_project_iam_binding` | ✅ Production |
| **IAM Service Accounts** | 3 | `google_service_account` | ✅ Production |
| **KMS Key Rings** | 1 | `google_kms_key_ring` | ✅ Production |
| **KMS Crypto Keys** | 2 | `google_kms_crypto_key` | ✅ Production |
| **Secret Manager** | 2 | `google_secret_manager_secret` | ✅ Production |

[Learn more →](iam.md) | [Learn more →](kms.md) | [Learn more →](secret-manager.md)

### Containers (9+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **GKE Clusters** | 3 | `google_container_cluster` | ✅ Production |
| **GKE Node Pools** | 3 | `google_container_node_pool` | ✅ Production |
| **Cloud Run Services** | 3 | `google_cloud_run_service` | ✅ Production |

[Learn more →](gke.md) | [Learn more →](cloud-run.md)

### Serverless (6+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **Cloud Functions v1** | 3 | `google_cloudfunctions_function` | ✅ Production |
| **Cloud Functions v2** | 3 | `google_cloudfunctions2_function` | ✅ Production |

[Learn more →](cloud-functions.md)

### Data & Analytics (11+ events)

| Service | Event Count | Resources | Status |
|---------|-------------|-----------|--------|
| **BigQuery Datasets** | 3 | `google_bigquery_dataset` | ✅ Production |
| **BigQuery Tables** | 3 | `google_bigquery_table` | ✅ Production |
| **Pub/Sub Topics** | 2 | `google_pubsub_topic` | ✅ Production |
| **Pub/Sub Subscriptions** | 2 | `google_pubsub_subscription` | ✅ Production |

[Learn more →](bigquery.md) | [Learn more →](pubsub.md)

---

## Event Types

### Supported Operations

TFDrift-Falco detects the following types of infrastructure changes:

| Operation Type | GCP Method Pattern | Example |
|----------------|-------------------|---------|
| **Create** | `.insert`, `.create` | `compute.instances.insert` |
| **Update** | `.update`, `.patch` | `compute.firewalls.update` |
| **Delete** | `.delete` | `storage.buckets.delete` |
| **Set Metadata** | `.setMetadata` | `compute.instances.setMetadata` |
| **Set Labels** | `.setLabels` | `compute.disks.setLabels` |
| **Set Tags** | `.setTags` | `compute.instances.setTags` |
| **Set IAM Policy** | `SetIamPolicy` | IAM policy changes |

### Example Events

```yaml
# Compute Instance Metadata Change
gcp.methodName: "compute.instances.setMetadata"
gcp.resource.name: "projects/my-project/zones/us-central1-a/instances/web-server"
gcp.authenticationInfo.principalEmail: "admin@example.com"

# Firewall Rule Update
gcp.methodName: "compute.firewalls.update"
gcp.resource.name: "projects/my-project/global/firewalls/allow-http"
gcp.authenticationInfo.principalEmail: "terraform@my-project.iam.gserviceaccount.com"

# GCS Bucket IAM Change
gcp.methodName: "storage.buckets.setIamPolicy"
gcp.resource.name: "projects/_/buckets/my-app-data"
gcp.authenticationInfo.principalEmail: "admin@example.com"
```

---

## Coverage Comparison

### TFDrift-Falco v0.5.0 GCP Coverage

| Category | Event Count | Service Count | Coverage Level |
|----------|-------------|---------------|----------------|
| **Compute & Networking** | 40+ | 8 | ⭐⭐⭐⭐⭐ Excellent |
| **Storage** | 5+ | 1 | ⭐⭐⭐⭐ Good |
| **Databases** | 10+ | 1 | ⭐⭐⭐⭐ Good |
| **Security** | 12+ | 3 | ⭐⭐⭐⭐ Good |
| **Containers** | 9+ | 3 | ⭐⭐⭐⭐ Good |
| **Serverless** | 6+ | 1 | ⭐⭐⭐ Moderate |
| **Data & Analytics** | 11+ | 2 | ⭐⭐⭐ Moderate |
| **Total** | **100+** | **12+** | ⭐⭐⭐⭐ Production Ready |

### AWS vs GCP Coverage

| Metric | AWS (v0.2.0+) | GCP (v0.5.0+) |
|--------|---------------|---------------|
| **Events** | 203+ | 100+ |
| **Services** | 19+ | 12+ |
| **Resources** | 60+ | 40+ |
| **Maturity** | Production (v0.2.0) | Production (v0.5.0) |

---

## Roadmap

### Phase 1: Foundation (v0.5.0) ✅ Complete

- ✅ Core GCP services (Compute, Storage, SQL)
- ✅ GCS backend support
- ✅ Falco gcpaudit plugin integration
- ✅ Multi-provider architecture
- ✅ Comprehensive testing

### Phase 2: Expansion (v0.6.0) 🔄 Planned

- 🔄 Cloud Dataflow
- 🔄 Cloud Dataproc
- 🔄 Cloud Composer
- 🔄 Memorystore (Redis, Memcached)
- 🔄 Cloud Spanner

### Phase 3: Advanced (v0.7.0) 📋 Future

- 📋 Cloud Armor
- 📋 Cloud CDN
- 📋 Cloud DNS
- 📋 Cloud Load Balancing (advanced)
- 📋 Cloud Logging & Monitoring

---

## Getting Started

### Prerequisites

1. **GCP Project** with Audit Logs enabled
2. **Falco** with gcpaudit plugin installed
3. **Pub/Sub** subscription for Audit Logs
4. **Terraform state** in GCS, S3, or local

### Quick Start

```bash
# Run the GCP quick-start script
./scripts/gcp-quick-start.sh

# Or follow the manual setup guide
# See: docs/gcp-setup.md
```

### Configuration Example

```yaml
providers:
  gcp:
    enabled: true
    projects:
      - my-gcp-project-123
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod/terraform.tfstate"

drift_rules:
  - name: "GCP Firewall Rule Modification"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allowed"
      - "source_ranges"
    severity: "critical"

  - name: "GCS Bucket IAM Change"
    resource_types:
      - "google_storage_bucket_iam_binding"
    watched_attributes:
      - "members"
    severity: "high"
```

---

## Documentation

- [GCP Setup Guide](../../gcp-setup.md) - Complete setup instructions
- [Architecture](../../architecture.md) - Multi-cloud architecture details
- [Release Notes v0.5.0](../../release-notes/v0.5.0.md) - Full release details
- [Troubleshooting](../../gcp-setup.md#troubleshooting) - Common issues and solutions

---

## Support

- **GitHub Issues:** [Report bugs and request features](https://github.com/higakikeita/tfdrift-falco/issues)
- **GitHub Discussions:** [Ask questions and share ideas](https://github.com/higakikeita/tfdrift-falco/discussions)
- **Documentation:** [Complete documentation](https://higakikeita.github.io/tfdrift-falco/)

---

**Last Updated:** 2025-01-18
**Version:** v0.5.0
