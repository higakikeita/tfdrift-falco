# GCP Setup Guide for TFDrift-Falco

This guide walks you through setting up TFDrift-Falco for Google Cloud Platform (GCP) drift detection.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Architecture Overview](#architecture-overview)
- [Step 1: Enable GCP Audit Logs](#step-1-enable-gcp-audit-logs)
- [Step 2: Configure Pub/Sub for Audit Logs](#step-2-configure-pubsub-for-audit-logs)
- [Step 3: Install and Configure Falco](#step-3-install-and-configure-falco)
- [Step 4: Configure TFDrift-Falco](#step-4-configure-tfdrift-falco)
- [Step 5: Verify Setup](#step-5-verify-setup)
- [Troubleshooting](#troubleshooting)
- [Advanced Configuration](#advanced-configuration)

---

## Prerequisites

- **GCP Project** with appropriate permissions
- **Terraform** managing GCP resources
- **Falco 0.35+** (to be installed)
- **Docker** (optional, recommended for Falco)
- **GCP CLI (`gcloud`)** installed and configured

**Required GCP Permissions:**
- `logging.logEntries.list`
- `logging.sinks.create`
- `pubsub.topics.create`
- `pubsub.subscriptions.create`
- `storage.buckets.get` (for GCS state backend)

---

## Architecture Overview

```
┌─────────────────┐
│  GCP Resources  │
│  (Terraform)    │
└────────┬────────┘
         │
         │ Manual Changes
         │ (Console/CLI)
         ▼
┌─────────────────┐
│  GCP Audit Logs │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Cloud Pub/Sub  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Falco          │
│  (gcpaudit)     │
└────────┬────────┘
         │ gRPC
         ▼
┌─────────────────┐
│  TFDrift-Falco  │
│  + GCS Backend  │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Notifications  │
│  (Slack/etc)    │
└─────────────────┘
```

---

## Step 1: Enable GCP Audit Logs

### 1.1 Enable Admin Activity Logs (Enabled by Default)

Admin Activity audit logs are enabled by default and cannot be disabled.

### 1.2 Enable Data Access Logs (Optional but Recommended)

For comprehensive drift detection, enable Data Access logs:

```bash
# Create audit config
cat > audit-config.yaml <<EOF
auditConfigs:
- auditLogConfigs:
  - logType: ADMIN_READ
  - logType: DATA_READ
  - logType: DATA_WRITE
  service: compute.googleapis.com
- auditLogConfigs:
  - logType: ADMIN_READ
  - logType: DATA_WRITE
  service: storage.googleapis.com
- auditLogConfigs:
  - logType: ADMIN_READ
  - logType: DATA_WRITE
  service: sqladmin.googleapis.com
EOF

# Apply audit config to project
gcloud projects set-iam-policy PROJECT_ID audit-config.yaml
```

### 1.3 Verify Audit Logs

```bash
# List recent audit logs
gcloud logging read "protoPayload.serviceName=compute.googleapis.com" \
  --limit 10 \
  --format json
```

---

## Step 2: Configure Pub/Sub for Audit Logs

### 2.1 Create Pub/Sub Topic

```bash
# Set project
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# Create topic for audit logs
gcloud pubsub topics create tfdrift-audit-logs
```

### 2.2 Create Log Sink

Route audit logs to Pub/Sub:

```bash
# Create log sink
gcloud logging sinks create tfdrift-sink \
  pubsub.googleapis.com/projects/$PROJECT_ID/topics/tfdrift-audit-logs \
  --log-filter='
    protoPayload.serviceName="compute.googleapis.com" OR
    protoPayload.serviceName="storage.googleapis.com" OR
    protoPayload.serviceName="sqladmin.googleapis.com" OR
    protoPayload.serviceName="container.googleapis.com"
  '

# Get sink service account
SINK_SA=$(gcloud logging sinks describe tfdrift-sink --format="value(writerIdentity)")
echo "Sink Service Account: $SINK_SA"
```

### 2.3 Grant Permissions to Sink

```bash
# Grant publish permission to sink service account
gcloud pubsub topics add-iam-policy-binding tfdrift-audit-logs \
  --member="$SINK_SA" \
  --role="roles/pubsub.publisher"
```

### 2.4 Create Subscription for Falco

```bash
# Create pull subscription
gcloud pubsub subscriptions create tfdrift-falco-sub \
  --topic=tfdrift-audit-logs \
  --ack-deadline=60
```

---

## Step 3: Install and Configure Falco

### 3.1 Install Falco with Docker (Recommended)

```bash
# Pull Falco image
docker pull falcosecurity/falco:latest

# Create Falco config directory
mkdir -p ~/falco-config
```

### 3.2 Create Falco Configuration

```bash
cat > ~/falco-config/falco.yaml <<EOF
# Falco Configuration for GCP Audit Logs

# Enable gRPC output
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8

grpc_output:
  enabled: true

# Disable kernel module (not needed for cloud audit logs)
engine:
  kind: modern_ebpf
  modern_ebpf:
    cpus_for_each_buffer: 2

# Load GCP audit plugin
plugins:
  - name: gcpaudit
    library_path: /usr/share/falco/plugins/libgcpaudit.so
    init_config:
      project_id: "$PROJECT_ID"
      subscription: "tfdrift-falco-sub"
    open_params: ""

# Load rules for GCP
load_plugins: [gcpaudit]

# Output configuration
json_output: true
json_include_output_property: true
EOF
```

### 3.3 Create GCP Credentials Secret

```bash
# Create service account for Falco
gcloud iam service-accounts create tfdrift-falco \
  --display-name="TFDrift Falco Service Account"

# Grant Pub/Sub subscriber role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"

# Download key
gcloud iam service-accounts keys create ~/falco-config/gcp-key.json \
  --iam-account=tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com
```

### 3.4 Run Falco

```bash
# Run Falco with GCP plugin
docker run -d \
  --name falco \
  -p 5060:5060 \
  -v ~/falco-config:/etc/falco \
  -e GOOGLE_APPLICATION_CREDENTIALS=/etc/falco/gcp-key.json \
  falcosecurity/falco:latest \
  -c /etc/falco/falco.yaml
```

### 3.5 Verify Falco is Running

```bash
# Check Falco logs
docker logs falco

# You should see:
# "Falco initialized with GCP Audit Log plugin"
# "gRPC server listening on 0.0.0.0:5060"
```

---

## Step 4: Configure TFDrift-Falco

### 4.1 Create GCS Bucket for Terraform State (If Using GCS Backend)

```bash
# Create bucket
gsutil mb -p $PROJECT_ID -l us-central1 gs://tfdrift-terraform-state

# Enable versioning
gsutil versioning set on gs://tfdrift-terraform-state
```

### 4.2 Create TFDrift-Falco Configuration

```bash
cat > config-gcp.yaml <<EOF
# TFDrift-Falco GCP Configuration

providers:
  aws:
    enabled: false

  gcp:
    enabled: true
    projects:
      - "$PROJECT_ID"
    state:
      backend: "gcs"
      gcs_bucket: "tfdrift-terraform-state"
      gcs_prefix: "terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"  # or Falco container IP
  port: 5060

drift_rules:
  - name: "GCE Instance Configuration Change"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
      - "tags"
      - "machine_type"
      - "service_account"
    severity: "high"

  - name: "Firewall Rule Modification"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allow"
      - "deny"
      - "source_ranges"
      - "target_tags"
    severity: "critical"

  - name: "Cloud Storage Security Settings"
    resource_types:
      - "google_storage_bucket"
    watched_attributes:
      - "encryption"
      - "public_access_prevention"
      - "uniform_bucket_level_access"
    severity: "critical"

  - name: "Cloud SQL Instance Change"
    resource_types:
      - "google_sql_database_instance"
    watched_attributes:
      - "settings"
      - "database_version"
      - "deletion_protection"
    severity: "high"

  - name: "IAM Policy Change"
    resource_types:
      - "google_project_iam_binding"
      - "google_project_iam_member"
    watched_attributes:
      - "role"
      - "members"
    severity: "critical"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#gcp-security-alerts"

  webhook:
    enabled: false
    url: "https://your-siem.example.com/webhook"

logging:
  level: "info"
  format: "json"
EOF
```

### 4.3 Run TFDrift-Falco

#### Option A: Docker

```bash
docker run -d \
  --name tfdrift-falco \
  --network host \
  -v $(pwd)/config-gcp.yaml:/config/config.yaml:ro \
  -e GOOGLE_APPLICATION_CREDENTIALS=/config/gcp-key.json \
  -v ~/falco-config/gcp-key.json:/config/gcp-key.json:ro \
  ghcr.io/higakikeita/tfdrift-falco:latest \
  --config /config/config.yaml
```

#### Option B: Binary

```bash
# Set GCP credentials
export GOOGLE_APPLICATION_CREDENTIALS=~/falco-config/gcp-key.json

# Run TFDrift-Falco
./tfdrift --config config-gcp.yaml
```

---

## Step 5: Verify Setup

### 5.1 Test Drift Detection

Manually modify a Terraform-managed resource:

```bash
# Example: Add metadata to a GCE instance
gcloud compute instances add-metadata INSTANCE_NAME \
  --zone=us-central1-a \
  --metadata=test-key=test-value
```

### 5.2 Check TFDrift-Falco Logs

```bash
# Docker
docker logs -f tfdrift-falco

# You should see:
# "Drift Detected: google_compute_instance.INSTANCE_NAME"
# "Changed: metadata.test-key = null → test-value"
```

### 5.3 Check Slack Notification

You should receive a Slack alert with:
- 🚨 Resource: `google_compute_instance.INSTANCE_NAME`
- Changed attribute: `metadata.test-key`
- User: `your-email@example.com`
- Project: `your-gcp-project-id`

---

## Troubleshooting

### Issue 1: Falco Not Receiving Audit Logs

**Symptoms:**
- Falco starts but no events appear
- `docker logs falco` shows no audit log entries

**Solutions:**

```bash
# 1. Verify Pub/Sub subscription
gcloud pubsub subscriptions describe tfdrift-falco-sub

# 2. Check messages in subscription
gcloud pubsub subscriptions pull tfdrift-falco-sub --limit=5

# 3. Verify log sink
gcloud logging sinks describe tfdrift-sink

# 4. Check service account permissions
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:serviceAccount:tfdrift-falco@*"
```

### Issue 2: TFDrift-Falco Cannot Connect to Falco

**Symptoms:**
- Error: "failed to connect to Falco"
- gRPC connection refused

**Solutions:**

```bash
# 1. Check Falco gRPC port
docker exec falco netstat -tlnp | grep 5060

# 2. Test gRPC connection
grpcurl -plaintext localhost:5060 list

# 3. Check firewall rules (if using remote Falco)
gcloud compute firewall-rules create allow-falco-grpc \
  --allow=tcp:5060 \
  --source-ranges=YOUR_CLIENT_IP/32
```

### Issue 3: GCS Backend Authentication Fails

**Symptoms:**
- Error: "failed to read object from GCS"
- Permission denied

**Solutions:**

```bash
# 1. Verify service account has Storage Object Viewer role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# 2. Test GCS access manually
gsutil ls gs://tfdrift-terraform-state/

# 3. Verify GOOGLE_APPLICATION_CREDENTIALS is set
echo $GOOGLE_APPLICATION_CREDENTIALS
```

### Issue 4: Events Not Matching Terraform Resources

**Symptoms:**
- Audit logs received but no drift detected
- "Resource not found in Terraform state"

**Solutions:**

```bash
# 1. Verify Terraform state path
gsutil ls gs://tfdrift-terraform-state/terraform.tfstate

# 2. Check resource naming in Terraform vs GCP
terraform show -json | jq '.values.root_module.resources[] | {type, name}'

# 3. Enable debug logging
# In config.yaml:
logging:
  level: "debug"
```

### Issue 5: High Volume of Irrelevant Events

**Symptoms:**
- Too many events being processed
- Performance degradation

**Solutions:**

Update log sink filter to be more specific:

```bash
gcloud logging sinks update tfdrift-sink \
  --log-filter='
    protoPayload.serviceName="compute.googleapis.com" AND
    protoPayload.methodName=~"compute\.(instances|firewalls|networks)\.(insert|delete|update|patch|set.*)"
  '
```

---

## Advanced Configuration

### Multi-Project Setup

Monitor multiple GCP projects:

```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "project-1"
      - "project-2"
      - "project-3"
    state:
      backend: "gcs"
      gcs_bucket: "tfdrift-terraform-state"
      gcs_prefix: "project-{PROJECT_ID}/terraform.tfstate"
```

### Custom Falco Rules

Create custom rules for specific GCP events:

```yaml
# falco-custom-rules.yaml
- rule: Terraform Managed GCE Instance Modified
  desc: Detect modifications to Terraform-managed GCE instances
  condition: >
    gcp.methodName in (compute.instances.setMetadata,
                        compute.instances.setLabels,
                        compute.instances.setTags) and
    not gcp.authenticationInfo.principalEmail startswith "terraform-"
  output: >
    GCE instance modified outside Terraform
    (user=%gcp.authenticationInfo.principalEmail
     instance=%gcp.resource.name
     method=%gcp.methodName)
  priority: WARNING
  tags: [gcp, terraform, drift]
```

### Regional Deployment

Deploy TFDrift-Falco per region:

```yaml
# config-us-central1.yaml
providers:
  gcp:
    enabled: true
    projects:
      - "my-project"
    state:
      backend: "gcs"
      gcs_bucket: "tfdrift-state-us-central1"
      gcs_prefix: "terraform.tfstate"

# Separate config for each region
```

### Integration with SIEM

Send events to your SIEM:

```yaml
notifications:
  webhook:
    enabled: true
    url: "https://splunk.example.com/services/collector"
    headers:
      Authorization: "Splunk YOUR_HEC_TOKEN"
      Content-Type: "application/json"
```

---

## Performance Tuning

### Falco Configuration

```yaml
# falco.yaml
grpc:
  threadiness: 16  # Increase for high volume

plugins:
  - name: gcpaudit
    init_config:
      # Batch size for Pub/Sub pulls
      max_messages: 100
      # Subscription timeout
      timeout: 60s
```

### TFDrift-Falco Configuration

```yaml
# Filter irrelevant events at TFDrift level
drift_rules:
  - name: "High Priority Only"
    resource_types:
      - "google_compute_firewall"
      - "google_project_iam_binding"
    watched_attributes:
      - "*"
    severity: "critical"
```

---

## Security Best Practices

1. **Least Privilege**: Grant minimal IAM roles
2. **Service Account Keys**: Rotate regularly (every 90 days)
3. **Network Security**: Restrict Falco gRPC access
4. **Audit Logs Retention**: Configure log retention policies
5. **Encryption**: Enable encryption at rest for GCS state bucket

```bash
# Enable customer-managed encryption
gsutil encryption set \
  -k projects/$PROJECT_ID/locations/global/keyRings/tfdrift/cryptoKeys/state \
  gs://tfdrift-terraform-state
```

---

## Next Steps

- [Back to Main Documentation](../README.md)
- [Deployment Guide](./deployment.md)
- [AWS Setup Guide](./falco-setup.md)
- [Troubleshooting Guide](./troubleshooting.md)

---

## Support

- **Issues**: https://github.com/higakikeita/tfdrift-falco/issues
- **Discussions**: https://github.com/higakikeita/tfdrift-falco/discussions
- **Documentation**: https://tfdrift-falco.readthedocs.io

---

**Last Updated**: 2025-12-17
**TFDrift-Falco Version**: v0.5.0+
