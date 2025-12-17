# GCP Setup Guide for TFDrift-Falco

This guide walks you through setting up TFDrift-Falco for Google Cloud Platform (GCP) drift detection.

## Table of Contents

- [Quick Start (5 Minutes)](#quick-start-5-minutes) ⭐ **Start Here!**
- [Prerequisites](#prerequisites)
- [Architecture Overview](#architecture-overview)
- [Step 1: Enable GCP Audit Logs](#step-1-enable-gcp-audit-logs)
- [Step 2: Configure Pub/Sub for Audit Logs](#step-2-configure-pubsub-for-audit-logs)
- [Step 3: Install and Configure Falco](#step-3-install-and-configure-falco)
- [Step 4: Configure TFDrift-Falco](#step-4-configure-tfdrift-falco)
- [Step 5: Verify Setup](#step-5-verify-setup)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)
- [Complete Examples](#complete-examples)
- [Advanced Configuration](#advanced-configuration)

---

## Quick Start (5 Minutes)

**Want to try TFDrift-Falco with GCP right now?** This automated script sets everything up for you.

### One-Command Setup

```bash
# Download and run the setup script
curl -fsSL https://raw.githubusercontent.com/higakikeita/tfdrift-falco/main/scripts/gcp-quick-start.sh | bash
```

### Manual Quick Start

If you prefer to run commands manually:

```bash
# 1. Set your project
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# 2. Enable required APIs (30 seconds)
gcloud services enable logging.googleapis.com pubsub.googleapis.com compute.googleapis.com

# 3. Create Pub/Sub infrastructure (30 seconds)
gcloud pubsub topics create tfdrift-audit-logs
gcloud logging sinks create tfdrift-sink \
  pubsub.googleapis.com/projects/$PROJECT_ID/topics/tfdrift-audit-logs \
  --log-filter='protoPayload.serviceName="compute.googleapis.com"'

SINK_SA=$(gcloud logging sinks describe tfdrift-sink --format="value(writerIdentity)")
gcloud pubsub topics add-iam-policy-binding tfdrift-audit-logs \
  --member="$SINK_SA" --role="roles/pubsub.publisher"

gcloud pubsub subscriptions create tfdrift-falco-sub \
  --topic=tfdrift-audit-logs

# 4. Create service account for Falco (30 seconds)
gcloud iam service-accounts create tfdrift-falco \
  --display-name="TFDrift Falco"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"

mkdir -p ~/tfdrift-config
gcloud iam service-accounts keys create ~/tfdrift-config/gcp-key.json \
  --iam-account=tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com

# 5. Run Falco with Docker (1 minute)
cat > ~/tfdrift-config/falco.yaml <<EOF
engine:
  kind: modern_ebpf
plugins:
  - name: gcpaudit
    library_path: /usr/share/falco/plugins/libgcpaudit.so
    init_config:
      project_id: "$PROJECT_ID"
      subscription: "tfdrift-falco-sub"
load_plugins: [gcpaudit]
json_output: true
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8
grpc_output:
  enabled: true
EOF

docker run -d --name falco \
  -p 5060:5060 \
  -v ~/tfdrift-config:/etc/falco \
  -e GOOGLE_APPLICATION_CREDENTIALS=/etc/falco/gcp-key.json \
  falcosecurity/falco:latest \
  -c /etc/falco/falco.yaml

# 6. Create TFDrift-Falco config (30 seconds)
cat > ~/tfdrift-config/config-gcp.yaml <<EOF
providers:
  gcp:
    enabled: true
    projects:
      - "$PROJECT_ID"
    state:
      backend: "local"
      local_path: "./terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  - name: "GCE Instance Change"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
    severity: "high"

notifications:
  slack:
    enabled: false

logging:
  level: "info"
EOF

echo "✅ Setup complete!"
echo ""
echo "Next steps:"
echo "1. Create a test Terraform resource:"
echo "   terraform init && terraform apply"
echo ""
echo "2. Run TFDrift-Falco:"
echo "   tfdrift --config ~/tfdrift-config/config-gcp.yaml"
echo ""
echo "3. Make a manual change in GCP Console to trigger drift detection"
echo "   Example: gcloud compute instances add-metadata INSTANCE_NAME --metadata=test=value"
```

### What This Sets Up

- ✅ GCP Audit Logs → Pub/Sub pipeline
- ✅ Falco with gcpaudit plugin (Docker)
- ✅ TFDrift-Falco configuration
- ✅ Service account with minimal permissions

### Test It

```bash
# 1. Create a simple test resource with Terraform
cat > main.tf <<EOF
resource "google_compute_network" "test" {
  name = "tfdrift-test-network"
  auto_create_subnetworks = false
}
EOF

terraform init
terraform apply -auto-approve

# 2. Run TFDrift-Falco
tfdrift --config ~/tfdrift-config/config-gcp.yaml &

# 3. Make a manual change
gcloud compute networks update tfdrift-test-network \
  --description="Manual change - should trigger drift"

# You should see a drift detection alert!
```

### Clean Up

```bash
# Remove test resources
terraform destroy -auto-approve

# Stop Falco
docker stop falco && docker rm falco

# Delete GCP resources
gcloud pubsub subscriptions delete tfdrift-falco-sub
gcloud pubsub topics delete tfdrift-audit-logs
gcloud logging sinks delete tfdrift-sink
gcloud iam service-accounts delete tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com
```

---

## Prerequisites

### Required Tools

- **GCP Project** with appropriate permissions
- **Terraform 1.0+** managing GCP resources
- **Falco 0.35+** (to be installed)
- **Docker 20.10+** (optional, recommended for Falco)
- **GCP CLI (`gcloud`)** installed and configured

### Pre-flight Checklist

Run these commands to verify your environment is ready:

```bash
# Check gcloud is installed and authenticated
gcloud --version
gcloud auth list

# Verify you have an active project
export PROJECT_ID=$(gcloud config get-value project)
echo "Current project: $PROJECT_ID"

# Check required APIs are enabled
gcloud services list --enabled | grep -E "(logging|pubsub|compute)"

# Enable required APIs if not already enabled
gcloud services enable \
  logging.googleapis.com \
  pubsub.googleapis.com \
  compute.googleapis.com \
  storage-api.googleapis.com

# Check Terraform is installed
terraform version

# Check Docker is running (if using Docker for Falco)
docker ps

# Verify you have sufficient permissions
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:user:$(gcloud config get-value account)" \
  --format="table(bindings.role)"
```

**Expected Result:** All commands should complete successfully without errors.

### Required GCP Permissions

Your account needs these IAM roles:
- `roles/logging.admin` - Create log sinks
- `roles/pubsub.admin` - Create Pub/Sub topics and subscriptions
- `roles/iam.serviceAccountAdmin` - Create service accounts
- `roles/storage.objectViewer` - Read Terraform state from GCS (if using GCS backend)

**Verify permissions:**
```bash
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:user:$(gcloud config get-value account)" \
  --format="table(bindings.role)"
```

### Estimated Time

- **Total setup time**: 20-30 minutes
- Step 1 (Audit Logs): 5 minutes
- Step 2 (Pub/Sub): 5 minutes
- Step 3 (Falco): 10 minutes
- Step 4 (TFDrift-Falco): 5 minutes
- Step 5 (Verification): 5 minutes

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

**⏱️ Estimated time: 5 minutes**

### 1.1 Enable Admin Activity Logs (Enabled by Default)

Admin Activity audit logs are enabled by default and cannot be disabled.

**✅ Verification:**
```bash
# Check if Admin Activity logs are flowing
gcloud logging read "protoPayload.serviceName=compute.googleapis.com" \
  --limit=5 \
  --format=json
```

**Expected output:** You should see recent audit log entries.

### 1.2 Enable Data Access Logs (Optional but Recommended)

For comprehensive drift detection, enable Data Access logs:

> ⚠️ **Warning:** Data Access logs can increase your Cloud Logging costs. Start with Admin Activity logs only for testing, then enable Data Access logs for production monitoring.

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

**⏱️ Estimated time: 5 minutes**

### 2.1 Create Pub/Sub Topic

```bash
# Set project (replace with your actual project ID)
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# Create topic for audit logs
gcloud pubsub topics create tfdrift-audit-logs
```

**✅ Verification:**
```bash
# Verify topic was created
gcloud pubsub topics describe tfdrift-audit-logs
```

**Expected output:**
```
name: projects/YOUR_PROJECT_ID/topics/tfdrift-audit-logs
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

**✅ Verification:**
```bash
# Verify sink was created and is active
gcloud logging sinks describe tfdrift-sink

# Check the filter
gcloud logging sinks describe tfdrift-sink --format="value(filter)"
```

**Expected output:**
```
Created [tfdrift-sink].
writerIdentity: serviceAccount:service-XXXX@gcp-sa-logging.iam.gserviceaccount.com
destination: pubsub.googleapis.com/projects/YOUR_PROJECT_ID/topics/tfdrift-audit-logs
```

> 💡 **Tip:** The `writerIdentity` is automatically created by Google Cloud and will be used to publish messages to Pub/Sub.

### 2.3 Grant Permissions to Sink

```bash
# Grant publish permission to sink service account
gcloud pubsub topics add-iam-policy-binding tfdrift-audit-logs \
  --member="$SINK_SA" \
  --role="roles/pubsub.publisher"
```

**✅ Verification:**
```bash
# Verify IAM policy
gcloud pubsub topics get-iam-policy tfdrift-audit-logs
```

**Expected output:** You should see the sink service account with `roles/pubsub.publisher` role.

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

### Issue 6: Falco Container Crashes or Restarts

**Symptoms:**
- Docker container exits immediately after starting
- Error: `error opening device /dev/host/proc`
- Container restart loop

**Error Messages:**
```
Error: Unable to load gcpaudit plugin: Cannot open library: /usr/share/falco/plugins/libgcpaudit.so
```

**Solutions:**

```bash
# 1. Check container logs for exact error
docker logs falco --tail 50

# 2. Verify plugin exists in container
docker exec falco ls -la /usr/share/falco/plugins/

# 3. Use correct Falco version with gcpaudit plugin (0.37.0+)
docker pull falcosecurity/falco:latest

# 4. Check configuration file syntax
docker run --rm -v ~/tfdrift-config:/config \
  falcosecurity/falco:latest \
  -c /config/falco.yaml --validate

# 5. Verify GOOGLE_APPLICATION_CREDENTIALS path
docker exec falco ls -la $GOOGLE_APPLICATION_CREDENTIALS
```

### Issue 7: Permission Denied on Pub/Sub Subscription

**Symptoms:**
- Error: `PERMISSION_DENIED: User not authorized to perform this action`
- Falco cannot pull messages from subscription

**Error Messages:**
```
ERROR Failed to pull messages: rpc error: code = PermissionDenied
desc = User not authorized to perform this action.
```

**Solutions:**

```bash
# 1. Verify service account has Pub/Sub Subscriber role
gcloud projects get-iam-policy $PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.role:roles/pubsub.subscriber"

# 2. Grant required permission
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"

# 3. Test subscription access with service account
gcloud pubsub subscriptions pull tfdrift-falco-sub \
  --limit=1 \
  --impersonate-service-account=tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com

# 4. Verify key file has correct format
cat ~/tfdrift-config/gcp-key.json | jq .
# Should show valid JSON with private_key, client_email, etc.
```

### Issue 8: Log Sink Service Account Missing Permissions

**Symptoms:**
- Audit logs generated but not appearing in Pub/Sub
- Log sink exists but messages not delivered

**Error Messages:**
```
The caller does not have permission to publish to topic
projects/PROJECT_ID/topics/tfdrift-audit-logs
```

**Solutions:**

```bash
# 1. Get the log sink service account (writer identity)
SINK_SA=$(gcloud logging sinks describe tfdrift-sink \
  --project=$PROJECT_ID \
  --format="value(writerIdentity)")
echo "Sink Service Account: $SINK_SA"

# 2. Grant Publisher role to the sink service account
gcloud pubsub topics add-iam-policy-binding tfdrift-audit-logs \
  --member="$SINK_SA" \
  --role="roles/pubsub.publisher" \
  --project=$PROJECT_ID

# 3. Verify the binding
gcloud pubsub topics get-iam-policy tfdrift-audit-logs \
  --project=$PROJECT_ID

# 4. Trigger a test event and check delivery
gcloud compute instances list  # Triggers compute.instances.list
sleep 30  # Wait for log delivery
gcloud pubsub subscriptions pull tfdrift-falco-sub --limit=1
```

### Issue 9: Invalid Configuration File

**Symptoms:**
- TFDrift-Falco fails to start
- Error: `failed to load configuration`
- YAML parsing errors

**Error Messages:**
```
Error: yaml: unmarshal errors:
  line 12: cannot unmarshal !!str `tfdrift...` into []string
```

**Solutions:**

```bash
# 1. Validate YAML syntax
yamllint ~/tfdrift-config/config-gcp.yaml

# Or use Python
python3 -c "import yaml; yaml.safe_load(open('config-gcp.yaml'))"

# 2. Check common issues:

# ❌ Wrong: projects as string
providers:
  gcp:
    projects: "my-project"  # Wrong

# ✅ Correct: projects as list
providers:
  gcp:
    projects:
      - "my-project"

# ❌ Wrong: watched_attributes with quotes issues
watched_attributes:
  - metadata  # Missing quotes

# ✅ Correct:
watched_attributes:
  - "metadata"
  - "labels"

# 3. Use config validation if available
tfdrift --config ~/tfdrift-config/config-gcp.yaml --validate
```

### Issue 10: Drift Not Detected for Specific Resources

**Symptoms:**
- Some resources show drift, others don't
- Manual changes not triggering alerts
- Logs show "resource not found in state"

**Solutions:**

```bash
# 1. Verify resource exists in Terraform state
terraform show -json | jq -r \
  '.values.root_module.resources[] | select(.type=="google_compute_instance") | .name'

# 2. Check resource name format matches
# GCP Audit Log format: projects/PROJECT_ID/zones/ZONE/instances/NAME
# Terraform resource address: google_compute_instance.NAME

# 3. Enable debug logging to see resource matching
# In config-gcp.yaml:
logging:
  level: "debug"
  format: "json"

# 4. Check drift rule configuration
# Ensure resource type matches exactly:
drift_rules:
  - name: "GCE Instance Drift"
    resource_types:
      - "google_compute_instance"  # Must match Terraform type exactly

# 5. Verify watched attributes exist in Terraform state
terraform show -json | jq \
  '.values.root_module.resources[] |
   select(.type=="google_compute_instance") |
   .values | keys'
```

### Issue 11: Webhook Notifications Not Sending

**Symptoms:**
- Drift detected but no Slack/webhook notification
- No errors in TFDrift-Falco logs

**Solutions:**

```bash
# 1. Test webhook URL manually
curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
  -H 'Content-Type: application/json' \
  -d '{"text":"Test message from TFDrift-Falco"}'

# 2. Check webhook configuration
# In config-gcp.yaml:
notifications:
  slack:
    enabled: true  # Must be true
    webhook_url: "https://hooks.slack.com/services/..."  # Must be valid URL

  webhook:
    enabled: true
    url: "https://your-webhook.example.com/tfdrift"
    method: "POST"
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer YOUR_TOKEN"

# 3. Enable verbose logging for notifications
logging:
  level: "debug"

# 4. Check for network connectivity issues
curl -v https://hooks.slack.com
```

### Issue 12: Multiple Projects Not Loading State

**Symptoms:**
- Only first project's state loaded
- Resources from other projects not detected

**Solutions:**

```yaml
# ❌ Wrong: Single state file for multiple projects
providers:
  gcp:
    projects:
      - "project-1"
      - "project-2"
    state:
      backend: "gcs"
      gcs_bucket: "terraform-state"
      gcs_prefix: "terraform.tfstate"  # Same file for all!

# ✅ Correct: Use {PROJECT_ID} placeholder
providers:
  gcp:
    projects:
      - "project-1"
      - "project-2"
    state:
      backend: "gcs"
      gcs_bucket: "terraform-state"
      gcs_prefix: "{PROJECT_ID}/terraform.tfstate"  # Different per project
```

```bash
# Verify state files exist for all projects
for project in project-1 project-2; do
  gsutil ls gs://terraform-state/$project/terraform.tfstate
done
```

---

## Debug Procedures

### Step-by-Step Debugging Workflow

When TFDrift-Falco is not working as expected, follow this systematic approach:

#### 1. Verify the Complete Pipeline

```bash
#!/bin/bash
# Debug script - save as debug-tfdrift.sh

PROJECT_ID="your-project-id"

echo "==> 1. Checking GCP Audit Logs"
gcloud logging read "protoPayload.serviceName=compute.googleapis.com" \
  --limit=3 \
  --format=json \
  --project=$PROJECT_ID | jq '.[0].protoPayload.methodName'

echo "==> 2. Checking Log Sink"
gcloud logging sinks describe tfdrift-sink --project=$PROJECT_ID

echo "==> 3. Checking Pub/Sub Topic"
gcloud pubsub topics describe tfdrift-audit-logs --project=$PROJECT_ID

echo "==> 4. Checking Pub/Sub Subscription"
gcloud pubsub subscriptions describe tfdrift-falco-sub --project=$PROJECT_ID

echo "==> 5. Pulling Sample Message"
gcloud pubsub subscriptions pull tfdrift-falco-sub \
  --limit=1 \
  --format=json \
  --project=$PROJECT_ID | jq '.[0].message.data' -r | base64 -d | jq .

echo "==> 6. Checking Falco Container"
docker ps | grep falco
docker logs falco --tail 20

echo "==> 7. Testing Falco gRPC"
grpcurl -plaintext localhost:5060 list

echo "==> 8. Testing TFDrift-Falco Connection"
# Run TFDrift-Falco with debug logging
tfdrift --config config-gcp.yaml --log-level=debug
```

#### 2. Enable Maximum Verbosity

```yaml
# config-gcp.yaml - Debug configuration
logging:
  level: "debug"  # trace, debug, info, warn, error
  format: "json"  # json or text

falco:
  enabled: true
  hostname: "localhost"
  port: 5060
  timeout: 30s
  retry:
    max_attempts: 3
    initial_interval: "1s"
```

```yaml
# falco.yaml - Debug configuration
json_output: true
json_include_output_property: true
log_level: debug  # Add this for Falco debug logs

grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8

grpc_output:
  enabled: true
```

#### 3. Test Each Component Independently

```bash
# Test 1: Trigger a known GCP event
echo "Creating test compute instance..."
gcloud compute instances create tfdrift-test-instance \
  --zone=us-central1-a \
  --machine-type=e2-micro \
  --project=$PROJECT_ID

# Wait for audit log (30 seconds typical)
sleep 35

# Test 2: Check if audit log was created
gcloud logging read \
  'protoPayload.methodName="v1.compute.instances.insert" AND
   resource.labels.instance_id="tfdrift-test-instance"' \
  --limit=1 \
  --format=json \
  --project=$PROJECT_ID

# Test 3: Check if it reached Pub/Sub
gcloud pubsub subscriptions pull tfdrift-falco-sub \
  --limit=5 \
  --format=json \
  --project=$PROJECT_ID

# Test 4: Check Falco received it
docker logs falco --tail 50 | grep "compute.instances.insert"

# Cleanup
gcloud compute instances delete tfdrift-test-instance \
  --zone=us-central1-a \
  --quiet \
  --project=$PROJECT_ID
```

#### 4. Isolate Network Issues

```bash
# Test gRPC connectivity from TFDrift-Falco perspective
grpcurl -plaintext localhost:5060 list

# If using remote Falco, test from client machine
grpcurl -plaintext FALCO_HOST:5060 list

# Check Docker network
docker network inspect bridge

# Check port bindings
docker port falco
# Should show: 5060/tcp -> 0.0.0.0:5060

# Test with telnet
telnet localhost 5060
```

---

## Log Analysis Guide

### Understanding GCP Audit Log Structure

GCP Audit Logs delivered via Falco gcpaudit plugin have this structure:

```json
{
  "protoPayload": {
    "serviceName": "compute.googleapis.com",
    "methodName": "v1.compute.instances.setMetadata",
    "resourceName": "projects/123456789/zones/us-central1-a/instances/my-instance",
    "authenticationInfo": {
      "principalEmail": "user@example.com"
    },
    "request": {
      "metadata": {
        "items": [
          {
            "key": "ssh-keys",
            "value": "user:ssh-rsa AAAA..."
          }
        ]
      }
    },
    "response": {
      "operationType": "setMetadata"
    }
  },
  "timestamp": "2025-12-17T10:30:45.123456Z",
  "severity": "NOTICE"
}
```

### Key Fields for Drift Detection

| Field | Purpose | Example |
|-------|---------|---------|
| `serviceName` | Which GCP service | `compute.googleapis.com` |
| `methodName` | What action | `v1.compute.instances.setMetadata` |
| `resourceName` | Which resource | `projects/.../instances/my-instance` |
| `principalEmail` | Who made the change | `user@example.com` |
| `request` | What changed | New metadata values |
| `timestamp` | When | ISO 8601 timestamp |

### Falco gRPC Output Format

When Falco forwards events to TFDrift-Falco via gRPC:

```json
{
  "output": "GCP Audit Log Event",
  "priority": "Notice",
  "rule": "GCP Audit Log",
  "time": "2025-12-17T10:30:45.123456Z",
  "output_fields": {
    "gcp.serviceName": "compute.googleapis.com",
    "gcp.methodName": "v1.compute.instances.setMetadata",
    "gcp.resourceName": "projects/123456789/zones/us-central1-a/instances/my-instance",
    "gcp.principalEmail": "user@example.com",
    "gcp.projectId": "my-project-123"
  },
  "source": "gcpaudit"
}
```

### TFDrift-Falco Event Processing

TFDrift-Falco processes events through these stages:

```
1. Receive gRPC event from Falco
   ↓
2. Parse GCP-specific fields (gcp.serviceName, gcp.methodName, etc.)
   ↓
3. Map methodName to Terraform resource type
   compute.instances.setMetadata → google_compute_instance
   ↓
4. Extract resource identifier from resourceName
   projects/.../instances/my-instance → my-instance
   ↓
5. Load Terraform state for project
   ↓
6. Find matching resource in state
   google_compute_instance.my_instance
   ↓
7. Compare changed attributes with drift rules
   ↓
8. Generate drift alert if mismatch detected
```

### Reading TFDrift-Falco Logs

**Normal Operation:**
```
INFO[2025-12-17T10:30:00Z] Starting TFDrift-Falco v0.5.0
INFO[2025-12-17T10:30:01Z] Connected to Falco gRPC at localhost:5060
INFO[2025-12-17T10:30:01Z] Loaded Terraform state for project my-project-123 (45 resources)
INFO[2025-12-17T10:30:45Z] Received event: compute.instances.setMetadata (user@example.com)
INFO[2025-12-17T10:30:45Z] Mapped to resource: google_compute_instance.my_instance
WARN[2025-12-17T10:30:45Z] Drift detected: metadata changed on google_compute_instance.my_instance
INFO[2025-12-17T10:30:45Z] Sent Slack notification
```

**Debug Mode (--log-level=debug):**
```
DEBUG[2025-12-17T10:30:45Z] Raw Falco event: {output_fields:{gcp.methodName:v1.compute.instances.setMetadata}}
DEBUG[2025-12-17T10:30:45Z] Parsed GCP event: service=compute.googleapis.com method=setMetadata resource=my-instance
DEBUG[2025-12-17T10:30:45Z] Resource mapper: compute.instances → google_compute_instance
DEBUG[2025-12-17T10:30:45Z] State lookup: google_compute_instance.my_instance found
DEBUG[2025-12-17T10:30:45Z] Comparing attributes: [metadata, labels, tags]
DEBUG[2025-12-17T10:30:45Z] Attribute 'metadata' differs: state={...} event={...}
DEBUG[2025-12-17T10:30:45Z] Drift rule matched: GCE Instance Configuration Change (severity: high)
DEBUG[2025-12-17T10:30:45Z] Calling webhook: https://hooks.slack.com/services/...
DEBUG[2025-12-17T10:30:46Z] Webhook response: 200 OK
```

### Common Log Patterns and Meanings

**Pattern: "Resource not found in Terraform state"**
```
WARN[...] Resource not found in Terraform state: google_compute_instance.my_instance
```
**Meaning:** The GCP resource exists and was modified, but it's not managed by Terraform (or state file path is wrong)

**Pattern: "Failed to connect to Falco"**
```
ERROR[...] Failed to connect to Falco gRPC: connection refused
```
**Meaning:** TFDrift-Falco cannot reach Falco on the configured hostname:port

**Pattern: "Failed to load Terraform state"**
```
ERROR[...] Failed to load Terraform state from GCS: storage: object doesn't exist
```
**Meaning:** State file path in configuration doesn't match actual GCS object path

**Pattern: "No drift rules matched"**
```
DEBUG[...] No drift rules matched for resource type: google_compute_firewall
```
**Meaning:** Event received, but no drift rules configured for this resource type

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

## Complete Examples

This section provides production-ready Terraform configurations with corresponding TFDrift-Falco setups.

### Example 1: Basic GCE Instance with Networking

**Scenario:** Monitor a simple compute instance for configuration drift.

**Terraform Configuration** (`main.tf`):
```hcl
# Provider configuration
terraform {
  required_version = ">= 1.0"

  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }

  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "prod/compute"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

# Variables
variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "us-central1"
}

variable "zone" {
  description = "GCP Zone"
  type        = string
  default     = "us-central1-a"
}

# VPC Network
resource "google_compute_network" "main" {
  name                    = "tfdrift-demo-network"
  auto_create_subnetworks = false
  description             = "Network managed by Terraform"
}

# Subnet
resource "google_compute_subnetwork" "main" {
  name          = "tfdrift-demo-subnet"
  ip_cidr_range = "10.0.1.0/24"
  region        = var.region
  network       = google_compute_network.main.id

  log_config {
    aggregation_interval = "INTERVAL_5_SEC"
    flow_sampling        = 0.5
  }
}

# Firewall - Allow SSH
resource "google_compute_firewall" "allow_ssh" {
  name    = "tfdrift-demo-allow-ssh"
  network = google_compute_network.main.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]  # IAP ranges
  target_tags   = ["ssh-enabled"]

  description = "Allow SSH via IAP"
}

# Compute Instance
resource "google_compute_instance" "web" {
  name         = "tfdrift-demo-web-server"
  machine_type = "e2-medium"
  zone         = var.zone

  tags = ["ssh-enabled", "web-server"]

  labels = {
    environment = "production"
    managed_by  = "terraform"
    app         = "web"
  }

  boot_disk {
    initialize_params {
      image = "debian-cloud/debian-11"
      size  = 20
      type  = "pd-standard"
    }
  }

  network_interface {
    subnetwork = google_compute_subnetwork.main.id

    access_config {
      // Ephemeral public IP
    }
  }

  metadata = {
    enable-oslogin = "TRUE"
    startup-script = <<-EOF
      #!/bin/bash
      apt-get update
      apt-get install -y nginx
      systemctl start nginx
      systemctl enable nginx
    EOF
  }

  service_account {
    email  = google_service_account.instance_sa.email
    scopes = ["cloud-platform"]
  }

  scheduling {
    automatic_restart   = true
    on_host_maintenance = "MIGRATE"
  }
}

# Service Account for Instance
resource "google_service_account" "instance_sa" {
  account_id   = "tfdrift-demo-instance-sa"
  display_name = "TFDrift Demo Instance Service Account"
  description  = "Service account for demo web server"
}

# Outputs
output "instance_name" {
  value = google_compute_instance.web.name
}

output "instance_external_ip" {
  value = google_compute_instance.web.network_interface[0].access_config[0].nat_ip
}

output "network_name" {
  value = google_compute_network.main.name
}
```

**TFDrift-Falco Configuration** (`config-demo.yaml`):
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "my-project-123"
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod/compute/terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060
  timeout: 30s

drift_rules:
  # Monitor instance configuration changes
  - name: "GCE Instance Configuration Drift"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
      - "tags"
      - "machine_type"
    severity: "high"

  # Critical: Monitor firewall rules
  - name: "Firewall Rule Modification"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allow"
      - "deny"
      - "source_ranges"
      - "target_tags"
    severity: "critical"

  # Monitor network changes
  - name: "Network Configuration Change"
    resource_types:
      - "google_compute_network"
      - "google_compute_subnetwork"
    watched_attributes:
      - "auto_create_subnetworks"
      - "ip_cidr_range"
    severity: "high"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

logging:
  level: "info"
  format: "text"
```

**Deployment Steps:**
```bash
# 1. Apply Terraform configuration
terraform init
terraform apply -var="project_id=my-project-123"

# 2. Start TFDrift-Falco
tfdrift --config config-demo.yaml

# 3. Trigger drift by making manual changes
gcloud compute instances add-labels tfdrift-demo-web-server \
  --zone=us-central1-a \
  --labels=manual_change=true

# Expected: Drift alert in Slack within 30-60 seconds
```

---

### Example 2: Multi-Tier Web Application

**Scenario:** Complete web application with load balancer, managed instance group, and Cloud SQL.

**Terraform Configuration** (`main.tf`):
```hcl
terraform {
  required_version = ">= 1.0"

  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "prod/webapp"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "us-central1"
}

# Network
resource "google_compute_network" "webapp" {
  name                    = "webapp-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "webapp" {
  name          = "webapp-subnet"
  ip_cidr_range = "10.0.0.0/24"
  region        = var.region
  network       = google_compute_network.webapp.id
}

# Firewall Rules
resource "google_compute_firewall" "allow_lb" {
  name    = "webapp-allow-lb"
  network = google_compute_network.webapp.name

  allow {
    protocol = "tcp"
    ports    = ["80", "443"]
  }

  source_ranges = ["130.211.0.0/22", "35.191.0.0/16"]
  target_tags   = ["web-backend"]
}

# Instance Template
resource "google_compute_instance_template" "webapp" {
  name_prefix  = "webapp-template-"
  machine_type = "e2-medium"

  tags = ["web-backend"]

  labels = {
    environment = "production"
    tier        = "web"
    managed_by  = "terraform"
  }

  disk {
    source_image = "debian-cloud/debian-11"
    auto_delete  = true
    boot         = true
    disk_size_gb = 20
  }

  network_interface {
    subnetwork = google_compute_subnetwork.webapp.id
  }

  metadata = {
    startup-script = templatefile("${path.module}/startup.sh", {
      db_host     = google_sql_database_instance.main.private_ip_address
      db_name     = google_sql_database.webapp.name
      db_user     = google_sql_user.webapp.name
      db_password = random_password.db_password.result
    })
  }

  service_account {
    email  = google_service_account.webapp.email
    scopes = ["cloud-platform"]
  }

  lifecycle {
    create_before_destroy = true
  }
}

# Managed Instance Group
resource "google_compute_region_instance_group_manager" "webapp" {
  name   = "webapp-mig"
  region = var.region

  base_instance_name = "webapp-instance"

  version {
    instance_template = google_compute_instance_template.webapp.id
  }

  target_size = 3

  named_port {
    name = "http"
    port = 80
  }

  auto_healing_policies {
    health_check      = google_compute_health_check.webapp.id
    initial_delay_sec = 300
  }
}

# Health Check
resource "google_compute_health_check" "webapp" {
  name = "webapp-health-check"

  http_health_check {
    port         = 80
    request_path = "/health"
  }

  check_interval_sec  = 10
  timeout_sec         = 5
  healthy_threshold   = 2
  unhealthy_threshold = 3
}

# Backend Service
resource "google_compute_backend_service" "webapp" {
  name                  = "webapp-backend"
  protocol              = "HTTP"
  port_name             = "http"
  timeout_sec           = 30
  enable_cdn            = true
  health_checks         = [google_compute_health_check.webapp.id]
  load_balancing_scheme = "EXTERNAL"

  backend {
    group           = google_compute_region_instance_group_manager.webapp.instance_group
    balancing_mode  = "UTILIZATION"
    capacity_scaler = 1.0
  }

  log_config {
    enable      = true
    sample_rate = 1.0
  }
}

# URL Map
resource "google_compute_url_map" "webapp" {
  name            = "webapp-url-map"
  default_service = google_compute_backend_service.webapp.id
}

# HTTP Proxy
resource "google_compute_target_http_proxy" "webapp" {
  name    = "webapp-http-proxy"
  url_map = google_compute_url_map.webapp.id
}

# Global Forwarding Rule
resource "google_compute_global_forwarding_rule" "webapp" {
  name       = "webapp-forwarding-rule"
  target     = google_compute_target_http_proxy.webapp.id
  port_range = "80"
}

# Cloud SQL Instance
resource "google_sql_database_instance" "main" {
  name             = "webapp-db-instance"
  database_version = "POSTGRES_14"
  region           = var.region

  settings {
    tier              = "db-f1-micro"
    availability_type = "REGIONAL"

    backup_configuration {
      enabled            = true
      start_time         = "03:00"
      point_in_time_recovery_enabled = true
    }

    ip_configuration {
      ipv4_enabled    = false
      private_network = google_compute_network.webapp.id
    }

    maintenance_window {
      day  = 7  # Sunday
      hour = 3
    }

    database_flags {
      name  = "log_connections"
      value = "on"
    }
  }

  deletion_protection = true
}

resource "google_sql_database" "webapp" {
  name     = "webapp"
  instance = google_sql_database_instance.main.name
}

resource "google_sql_user" "webapp" {
  name     = "webapp_user"
  instance = google_sql_database_instance.main.name
  password = random_password.db_password.result
}

resource "random_password" "db_password" {
  length  = 16
  special = true
}

# Service Account
resource "google_service_account" "webapp" {
  account_id   = "webapp-instance-sa"
  display_name = "WebApp Instance Service Account"
}

# IAM Bindings
resource "google_project_iam_member" "webapp_sql_client" {
  project = var.project_id
  role    = "roles/cloudsql.client"
  member  = "serviceAccount:${google_service_account.webapp.email}"
}

# Outputs
output "load_balancer_ip" {
  value = google_compute_global_forwarding_rule.webapp.ip_address
}

output "db_instance_connection" {
  value     = google_sql_database_instance.main.connection_name
  sensitive = true
}
```

**TFDrift-Falco Configuration** (`config-webapp.yaml`):
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "my-project-123"
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod/webapp/terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  # Critical: Database configuration
  - name: "Cloud SQL Configuration Change"
    resource_types:
      - "google_sql_database_instance"
    watched_attributes:
      - "settings"
      - "database_version"
      - "deletion_protection"
    severity: "critical"

  # Critical: IAM changes
  - name: "IAM Binding Modification"
    resource_types:
      - "google_project_iam_member"
      - "google_service_account_iam_binding"
    watched_attributes:
      - "role"
      - "members"
    severity: "critical"

  # High: Load balancer configuration
  - name: "Load Balancer Configuration Change"
    resource_types:
      - "google_compute_backend_service"
      - "google_compute_url_map"
      - "google_compute_target_http_proxy"
    watched_attributes:
      - "backend"
      - "health_checks"
      - "enable_cdn"
    severity: "high"

  # High: Instance template changes
  - name: "Instance Template Modification"
    resource_types:
      - "google_compute_instance_template"
    watched_attributes:
      - "machine_type"
      - "disk"
      - "metadata"
      - "service_account"
    severity: "high"

  # Medium: MIG scaling
  - name: "MIG Target Size Change"
    resource_types:
      - "google_compute_region_instance_group_manager"
    watched_attributes:
      - "target_size"
      - "version"
    severity: "medium"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

  webhook:
    enabled: true
    url: "https://monitoring.example.com/webhooks/tfdrift"
    method: "POST"
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer YOUR_API_TOKEN"

logging:
  level: "info"
  format: "json"
```

**Test Drift Detection:**
```bash
# 1. Deploy infrastructure
terraform apply -var="project_id=my-project-123"

# 2. Start TFDrift-Falco
tfdrift --config config-webapp.yaml

# 3. Trigger various drift scenarios

# Scenario A: Modify database backup settings (CRITICAL)
gcloud sql instances patch webapp-db-instance \
  --backup-start-time=04:00

# Scenario B: Change MIG target size (MEDIUM)
gcloud compute instance-groups managed set-autoscaling webapp-mig \
  --region=us-central1 \
  --max-num-replicas=5

# Scenario C: Modify backend service timeout (HIGH)
gcloud compute backend-services update webapp-backend \
  --global \
  --timeout=60

# Expected: Different severity alerts in Slack
```

---

### Example 3: GKE Cluster with Monitoring

**Terraform Configuration** (`gke-cluster.tf`):
```hcl
terraform {
  backend "gcs" {
    bucket = "my-terraform-state"
    prefix = "prod/gke"
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}

variable "project_id" {
  type = string
}

variable "region" {
  type    = string
  default = "us-central1"
}

variable "cluster_name" {
  type    = string
  default = "prod-gke-cluster"
}

# VPC for GKE
resource "google_compute_network" "gke" {
  name                    = "gke-network"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "gke" {
  name          = "gke-subnet"
  ip_cidr_range = "10.0.0.0/20"
  region        = var.region
  network       = google_compute_network.gke.id

  secondary_ip_range {
    range_name    = "pods"
    ip_cidr_range = "10.4.0.0/14"
  }

  secondary_ip_range {
    range_name    = "services"
    ip_cidr_range = "10.8.0.0/20"
  }
}

# GKE Cluster
resource "google_container_cluster" "primary" {
  name     = var.cluster_name
  location = var.region

  # We can't create a cluster with no node pool defined, but we want to only use
  # separately managed node pools. So we create the smallest possible default
  # node pool and immediately delete it.
  remove_default_node_pool = true
  initial_node_count       = 1

  network    = google_compute_network.gke.name
  subnetwork = google_compute_subnetwork.gke.name

  # IP allocation for VPC-native cluster
  ip_allocation_policy {
    cluster_secondary_range_name  = "pods"
    services_secondary_range_name = "services"
  }

  # Enable Workload Identity
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }

  # Master authorized networks
  master_authorized_networks_config {
    cidr_blocks {
      cidr_block   = "10.0.0.0/8"
      display_name = "Internal"
    }
  }

  # Monitoring and logging
  logging_service    = "logging.googleapis.com/kubernetes"
  monitoring_service = "monitoring.googleapis.com/kubernetes"

  # Addons
  addons_config {
    http_load_balancing {
      disabled = false
    }

    horizontal_pod_autoscaling {
      disabled = false
    }

    network_policy_config {
      disabled = false
    }
  }

  # Network policy
  network_policy {
    enabled = true
  }

  # Maintenance window
  maintenance_policy {
    daily_maintenance_window {
      start_time = "03:00"
    }
  }

  # Binary authorization
  binary_authorization {
    evaluation_mode = "PROJECT_SINGLETON_POLICY_ENFORCE"
  }
}

# Node Pool
resource "google_container_node_pool" "primary_nodes" {
  name       = "primary-node-pool"
  location   = var.region
  cluster    = google_container_cluster.primary.name
  node_count = 3

  node_config {
    machine_type = "e2-medium"

    labels = {
      environment = "production"
      managed_by  = "terraform"
    }

    tags = ["gke-node", "prod"]

    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]

    service_account = google_service_account.gke_nodes.email

    workload_metadata_config {
      mode = "GKE_METADATA"
    }

    shielded_instance_config {
      enable_secure_boot          = true
      enable_integrity_monitoring = true
    }
  }

  autoscaling {
    min_node_count = 2
    max_node_count = 10
  }

  management {
    auto_repair  = true
    auto_upgrade = true
  }
}

# Service Account for GKE nodes
resource "google_service_account" "gke_nodes" {
  account_id   = "gke-node-sa"
  display_name = "GKE Node Service Account"
}

# IAM bindings for nodes
resource "google_project_iam_member" "gke_node_sa_logging" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

resource "google_project_iam_member" "gke_node_sa_monitoring" {
  project = var.project_id
  role    = "roles/monitoring.metricWriter"
  member  = "serviceAccount:${google_service_account.gke_nodes.email}"
}

# Outputs
output "cluster_name" {
  value = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  value     = google_container_cluster.primary.endpoint
  sensitive = true
}

output "cluster_ca_certificate" {
  value     = google_container_cluster.primary.master_auth.0.cluster_ca_certificate
  sensitive = true
}
```

**TFDrift-Falco Configuration** (`config-gke.yaml`):
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "my-project-123"
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "prod/gke/terraform.tfstate"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  # Critical: GKE cluster configuration
  - name: "GKE Cluster Configuration Change"
    resource_types:
      - "google_container_cluster"
    watched_attributes:
      - "master_authorized_networks_config"
      - "workload_identity_config"
      - "binary_authorization"
      - "network_policy"
    severity: "critical"

  # High: Node pool configuration
  - name: "GKE Node Pool Modification"
    resource_types:
      - "google_container_node_pool"
    watched_attributes:
      - "node_config"
      - "autoscaling"
      - "management"
    severity: "high"

  # Medium: Node pool scaling
  - name: "Node Pool Size Change"
    resource_types:
      - "google_container_node_pool"
    watched_attributes:
      - "node_count"
    severity: "medium"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

logging:
  level: "info"
  format: "json"
```

---

### Example 4: Production Multi-Project Setup

**Directory Structure:**
```
terraform/
├── environments/
│   ├── prod/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── terraform.tfvars
│   └── staging/
│       ├── main.tf
│       ├── variables.tf
│       └── terraform.tfvars
├── modules/
│   ├── compute/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   ├── networking/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   └── outputs.tf
│   └── security/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
└── tfdrift-config/
    ├── config-prod.yaml
    └── config-staging.yaml
```

**Module Example** (`modules/compute/main.tf`):
```hcl
resource "google_compute_instance" "this" {
  name         = var.instance_name
  machine_type = var.machine_type
  zone         = var.zone

  tags   = var.tags
  labels = var.labels

  boot_disk {
    initialize_params {
      image = var.boot_disk_image
      size  = var.boot_disk_size
    }
  }

  network_interface {
    subnetwork = var.subnetwork

    dynamic "access_config" {
      for_each = var.enable_external_ip ? [1] : []
      content {
        nat_ip = var.external_ip
      }
    }
  }

  metadata = var.metadata

  service_account {
    email  = var.service_account_email
    scopes = var.service_account_scopes
  }
}
```

**Environment Configuration** (`environments/prod/main.tf`):
```hcl
terraform {
  backend "gcs" {
    bucket = "company-terraform-state"
    prefix = "prod"
  }
}

module "networking" {
  source = "../../modules/networking"

  project_id   = var.project_id
  region       = var.region
  network_name = "prod-network"

  subnets = [
    {
      name          = "prod-subnet-web"
      ip_cidr_range = "10.0.1.0/24"
      region        = "us-central1"
    },
    {
      name          = "prod-subnet-db"
      ip_cidr_range = "10.0.2.0/24"
      region        = "us-central1"
    }
  ]
}

module "web_servers" {
  source = "../../modules/compute"
  count  = 3

  instance_name = "prod-web-${count.index}"
  machine_type  = "n2-standard-2"
  zone          = "us-central1-a"
  subnetwork    = module.networking.subnet_ids["prod-subnet-web"]

  tags = ["web", "prod"]

  labels = {
    environment = "production"
    tier        = "web"
    managed_by  = "terraform"
  }

  service_account_email  = module.security.web_sa_email
  service_account_scopes = ["cloud-platform"]
}

module "security" {
  source = "../../modules/security"

  project_id = var.project_id
}
```

**TFDrift-Falco Production Config** (`tfdrift-config/config-prod.yaml`):
```yaml
providers:
  gcp:
    enabled: true
    projects:
      - "company-prod-123"
      - "company-prod-456"
    state:
      backend: "gcs"
      gcs_bucket: "company-terraform-state"
      gcs_prefix: "{PROJECT_ID}/terraform.tfstate"

falco:
  enabled: true
  hostname: "falco.internal.company.com"
  port: 5060
  tls:
    enabled: true
    ca_cert: "/etc/tfdrift/certs/ca.crt"
    client_cert: "/etc/tfdrift/certs/client.crt"
    client_key: "/etc/tfdrift/certs/client.key"

drift_rules:
  # Production-critical rules
  - name: "Production IAM Changes"
    resource_types:
      - "google_project_iam_member"
      - "google_project_iam_binding"
      - "google_service_account_iam_binding"
    watched_attributes:
      - "role"
      - "members"
    severity: "critical"

  - name: "Production Database Changes"
    resource_types:
      - "google_sql_database_instance"
    watched_attributes:
      - "settings"
      - "deletion_protection"
      - "database_version"
    severity: "critical"

  - name: "Production Network Security"
    resource_types:
      - "google_compute_firewall"
      - "google_compute_security_policy"
    watched_attributes:
      - "allow"
      - "deny"
      - "source_ranges"
    severity: "critical"

  - name: "Production Compute Changes"
    resource_types:
      - "google_compute_instance"
      - "google_compute_instance_template"
    watched_attributes:
      - "machine_type"
      - "metadata"
      - "labels"
      - "service_account"
    severity: "high"

  - name: "Production GKE Cluster Changes"
    resource_types:
      - "google_container_cluster"
      - "google_container_node_pool"
    watched_attributes:
      - "master_authorized_networks_config"
      - "node_config"
      - "autoscaling"
    severity: "high"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/T00/B00/XX"
    channel: "#prod-alerts"
    username: "TFDrift-Falco [PROD]"

  webhook:
    enabled: true
    url: "https://monitoring.company.com/api/v1/alerts/tfdrift"
    method: "POST"
    timeout: "10s"
    retry:
      max_attempts: 3
      initial_interval: "2s"
      max_interval: "10s"
    headers:
      Content-Type: "application/json"
      Authorization: "Bearer ${WEBHOOK_API_TOKEN}"
      X-Environment: "production"

  pagerduty:
    enabled: true
    integration_key: "${PAGERDUTY_INTEGRATION_KEY}"
    severity_mapping:
      critical: "critical"
      high: "error"
      medium: "warning"
      low: "info"

logging:
  level: "info"
  format: "json"
  output: "stdout"

filtering:
  # Ignore read-only operations
  exclude_events:
    - "*.list"
    - "*.get"
    - "*.describe"

  # Focus on specific services
  include_services:
    - "compute.googleapis.com"
    - "container.googleapis.com"
    - "sqladmin.googleapis.com"
    - "iam.googleapis.com"
    - "storage.googleapis.com"
```

**Deployment Script** (`deploy-prod.sh`):
```bash
#!/bin/bash
set -e

ENVIRONMENT="prod"
PROJECT_ID="company-prod-123"

echo "==> Deploying ${ENVIRONMENT} infrastructure..."

cd environments/${ENVIRONMENT}

# Initialize Terraform
terraform init

# Plan
terraform plan -var="project_id=${PROJECT_ID}" -out=tfplan

# Apply with approval
read -p "Apply changes? (yes/no): " APPLY
if [ "$APPLY" = "yes" ]; then
    terraform apply tfplan
    echo "✓ Terraform applied successfully"

    # Start TFDrift-Falco
    echo "==> Starting TFDrift-Falco..."
    tfdrift --config ../../tfdrift-config/config-${ENVIRONMENT}.yaml &
    TFDRIFT_PID=$!
    echo "✓ TFDrift-Falco started (PID: $TFDRIFT_PID)"

    # Save PID
    echo $TFDRIFT_PID > /var/run/tfdrift.pid
else
    echo "Deployment cancelled"
fi
```

---

## Production Best Practices

This section covers best practices for running TFDrift-Falco in production environments.

### 1. Infrastructure Design

#### State Management

**DO:**
- Use remote state backends (GCS) with versioning enabled
- Implement state locking to prevent concurrent modifications
- Organize state files by environment and project
- Use workspace separation for different environments

```yaml
# ✅ Good: Separate state files per project
providers:
  gcp:
    projects:
      - "company-prod-123"
      - "company-prod-456"
    state:
      backend: "gcs"
      gcs_bucket: "company-terraform-state"
      gcs_prefix: "{PROJECT_ID}/prod/terraform.tfstate"
```

**DON'T:**
- Use local state files in production
- Share state files across unrelated resources
- Store state files in public buckets

```yaml
# ❌ Bad: Single state for all projects
state:
  backend: "gcs"
  gcs_prefix: "all-projects.tfstate"  # Too broad!
```

#### Multi-Project Organization

**Recommended Structure:**
```
terraform/
├── shared-services/        # Shared infrastructure
│   ├── networking/
│   ├── security/
│   └── monitoring/
├── environments/
│   ├── prod/              # Production environment
│   │   ├── project-a/
│   │   └── project-b/
│   ├── staging/           # Staging environment
│   └── dev/               # Development environment
└── modules/               # Reusable modules
    ├── compute/
    ├── database/
    └── networking/
```

**TFDrift Configuration per Environment:**
```yaml
# prod/tfdrift-config.yaml
providers:
  gcp:
    projects:
      - "company-prod-project-a"
      - "company-prod-project-b"
    state:
      backend: "gcs"
      gcs_bucket: "company-terraform-state"
      gcs_prefix: "{PROJECT_ID}/prod/terraform.tfstate"

drift_rules:
  - name: "Critical Production Changes"
    resource_types:
      - "google_sql_database_instance"
      - "google_compute_firewall"
      - "google_project_iam_*"
    severity: "critical"
```

#### Resource Naming Conventions

**Consistent Naming Pattern:**
```
{environment}-{project}-{service}-{resource_type}-{identifier}

Examples:
- prod-webapp-lb-frontend
- staging-api-db-primary
- prod-shared-network-main
```

**In Terraform:**
```hcl
locals {
  environment = "prod"
  project     = "webapp"

  naming_prefix = "${local.environment}-${local.project}"
}

resource "google_compute_instance" "web" {
  name = "${local.naming_prefix}-web-${count.index + 1}"

  labels = {
    environment = local.environment
    project     = local.project
    managed_by  = "terraform"
    cost_center = var.cost_center
  }
}
```

---

### 2. Configuration Management

#### Environment Separation

**Use separate configurations for each environment:**

```bash
# Directory structure
configs/
├── dev.yaml          # Development settings
├── staging.yaml      # Staging settings
├── prod.yaml         # Production settings
└── shared.yaml       # Shared base config
```

**Example: Development Config**
```yaml
# configs/dev.yaml
providers:
  gcp:
    enabled: true
    projects:
      - "company-dev-123"
    state:
      backend: "local"  # OK for dev
      local_path: "./terraform.tfstate"

drift_rules:
  - name: "Dev Compute Changes"
    resource_types:
      - "google_compute_instance"
    severity: "medium"  # Lower severity for dev

logging:
  level: "debug"  # More verbose in dev

notifications:
  slack:
    enabled: false  # Don't spam Slack in dev
```

**Example: Production Config**
```yaml
# configs/prod.yaml
providers:
  gcp:
    enabled: true
    projects:
      - "company-prod-123"
      - "company-prod-456"
    state:
      backend: "gcs"  # Remote state required
      gcs_bucket: "company-terraform-state-prod"
      gcs_prefix: "{PROJECT_ID}/terraform.tfstate"

drift_rules:
  - name: "Critical Production IAM Changes"
    resource_types:
      - "google_project_iam_*"
      - "google_service_account_iam_*"
    severity: "critical"

  - name: "Production Database Changes"
    resource_types:
      - "google_sql_database_instance"
    severity: "critical"

logging:
  level: "info"  # Production logging
  format: "json"
  output: "stdout"

notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_PROD}"
    channel: "#prod-alerts"

  pagerduty:
    enabled: true
    integration_key: "${PAGERDUTY_KEY_PROD}"
```

#### Secret Management

**DO:**
- Use environment variables for secrets
- Integrate with secret managers (Secret Manager, Vault)
- Rotate credentials regularly
- Never commit secrets to version control

```bash
# ✅ Good: Environment variables
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/..."
export PAGERDUTY_INTEGRATION_KEY="xxx"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/key.json"

# Run with secrets from environment
tfdrift --config config-prod.yaml
```

```yaml
# ✅ Good: Reference environment variables in config
notifications:
  slack:
    webhook_url: "${SLACK_WEBHOOK_URL}"
  pagerduty:
    integration_key: "${PAGERDUTY_INTEGRATION_KEY}"
```

**DON'T:**
```yaml
# ❌ Bad: Hardcoded secrets
notifications:
  slack:
    webhook_url: "https://hooks.slack.com/services/T00/B00/actual_secret_here"
```

**Integration with GCP Secret Manager:**
```bash
# Store secret
echo -n "https://hooks.slack.com/services/..." | \
  gcloud secrets create tfdrift-slack-webhook \
    --data-file=- \
    --replication-policy="automatic"

# Retrieve secret at runtime
export SLACK_WEBHOOK_URL=$(gcloud secrets versions access latest \
  --secret="tfdrift-slack-webhook")

# Run TFDrift
tfdrift --config config-prod.yaml
```

#### Configuration Validation

**Pre-deployment validation script:**
```bash
#!/bin/bash
# validate-config.sh

CONFIG_FILE=$1

echo "==> Validating TFDrift configuration..."

# 1. Check YAML syntax
yamllint "$CONFIG_FILE" || exit 1

# 2. Check required fields
python3 << EOF
import yaml
import sys

with open("$CONFIG_FILE") as f:
    config = yaml.safe_load(f)

# Check providers
if 'providers' not in config or 'gcp' not in config['providers']:
    print("❌ Missing providers.gcp configuration")
    sys.exit(1)

# Check projects
if not config['providers']['gcp'].get('projects'):
    print("❌ No projects specified")
    sys.exit(1)

# Check state backend
if not config['providers']['gcp'].get('state', {}).get('backend'):
    print("❌ No state backend specified")
    sys.exit(1)

# Check drift rules
if not config.get('drift_rules'):
    print("⚠️  Warning: No drift rules defined")

print("✓ Configuration is valid")
EOF

# 3. Validate environment variables
required_vars=("SLACK_WEBHOOK_URL" "GOOGLE_APPLICATION_CREDENTIALS")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Missing required environment variable: $var"
        exit 1
    fi
done

echo "✓ All validations passed"
```

---

### 3. Monitoring & Alerting

#### Alert Routing Strategy

**Severity-based routing:**

```yaml
# Route alerts based on severity
notifications:
  # Critical: Page on-call engineer
  pagerduty:
    enabled: true
    integration_key: "${PAGERDUTY_KEY}"
    severity_mapping:
      critical: "critical"  # Triggers page
      high: "error"         # Creates incident
      medium: "warning"     # Creates incident (low priority)
      low: "info"           # Notification only

  # High/Critical: Post to Slack immediately
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_PROD}"
    channel: "#prod-alerts"
    severity_filter: ["critical", "high"]

  # All events: Send to SIEM
  webhook:
    enabled: true
    url: "https://siem.company.com/api/events"
    severity_filter: ["critical", "high", "medium", "low"]

  # Critical only: Email leadership
  email:
    enabled: true
    smtp_server: "smtp.company.com"
    to: ["oncall@company.com", "security@company.com"]
    severity_filter: ["critical"]
```

#### Severity Level Guidelines

**Define clear severity criteria:**

| Severity | When to Use | Response Time | Examples |
|----------|-------------|---------------|----------|
| **Critical** | Security-impacting changes, data loss risk | Immediate (< 5 min) | IAM changes, firewall rules, database deletion protection |
| **High** | Service-impacting changes, compliance violations | < 30 minutes | Instance type changes, network config, encryption settings |
| **Medium** | Non-critical config changes, performance impact | < 4 hours | Labels, tags, non-critical metadata |
| **Low** | Informational, tracking purposes | Next business day | Read-only attribute changes |

**Configure in drift rules:**
```yaml
drift_rules:
  # Critical: Zero tolerance
  - name: "IAM Permission Changes"
    resource_types:
      - "google_project_iam_*"
      - "google_service_account_iam_*"
    watched_attributes:
      - "role"
      - "members"
    severity: "critical"
    alert_immediately: true

  # High: Important but not emergency
  - name: "Database Configuration"
    resource_types:
      - "google_sql_database_instance"
    watched_attributes:
      - "settings"
      - "database_version"
    severity: "high"
    alert_immediately: true

  # Medium: Track and review
  - name: "Instance Metadata"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
      - "labels"
    severity: "medium"
    alert_immediately: false
    batch_alerts: true
    batch_interval: "15m"

  # Low: Informational
  - name: "Resource Tags"
    resource_types:
      - "google_compute_*"
    watched_attributes:
      - "tags"
    severity: "low"
    alert_immediately: false
    batch_alerts: true
    batch_interval: "1h"
```

#### Monitoring TFDrift-Falco Itself

**Health Check Endpoint:**
```bash
# Implement health check
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "version": "v0.5.0",
  "uptime_seconds": 3600,
  "last_event_received": "2025-12-17T10:30:45Z",
  "falco_connection": "connected",
  "state_backend": "gcs",
  "projects_monitored": 2
}
```

**Monitoring Script:**
```bash
#!/bin/bash
# monitor-tfdrift.sh

HEALTH_URL="http://localhost:8080/health"
ALERT_WEBHOOK="https://monitoring.company.com/alert"

while true; do
    RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" "$HEALTH_URL")

    if [ "$RESPONSE" != "200" ]; then
        # Alert: TFDrift is down
        curl -X POST "$ALERT_WEBHOOK" \
          -H 'Content-Type: application/json' \
          -d '{
            "severity": "critical",
            "service": "tfdrift-falco",
            "message": "TFDrift health check failed",
            "status_code": "'$RESPONSE'"
          }'
    fi

    sleep 60
done
```

**Log Monitoring:**
```bash
# Monitor for errors in logs
journalctl -u tfdrift-falco -f | grep -E "(ERROR|FATAL)" | while read line; do
    # Alert on errors
    curl -X POST "$ALERT_WEBHOOK" \
      -d "TFDrift Error: $line"
done
```

---

### 4. Operational Excellence

#### Change Management

**Terraform Change Workflow:**
```
1. Developer creates Terraform change
   ↓
2. CI/CD runs terraform plan
   ↓
3. Pull request review
   ↓
4. Approved → terraform apply
   ↓
5. TFDrift-Falco monitors for manual changes
   ↓
6. Alert if drift detected within 1 hour
```

**Handling Drift Alerts:**

1. **Immediate Response (< 5 minutes):**
   - Acknowledge alert
   - Check if change was authorized
   - If unauthorized, investigate and remediate

2. **Investigation:**
   - Who made the change? (check principalEmail)
   - What was changed? (review diff)
   - Why was manual change made? (was it emergency?)

3. **Remediation Options:**
   - **Option A: Revert manual change**
     ```bash
     # Reapply Terraform to revert
     terraform apply
     ```

   - **Option B: Update Terraform state**
     ```bash
     # If change is authorized, update Terraform
     # 1. Update .tf files
     # 2. Run terraform plan to verify
     # 3. Apply changes
     terraform apply
     ```

   - **Option C: Accept drift temporarily**
     ```bash
     # Document in issue tracker
     # Schedule Terraform update
     ```

4. **Post-Incident:**
   - Document incident
   - Update runbooks
   - Review access controls
   - Consider automation improvements

**Incident Response Template:**
```markdown
# Drift Alert Incident Report

**Date:** 2025-12-17
**Severity:** Critical
**Resource:** google_compute_firewall.prod_allow_ssh
**Change Detected:** source_ranges modified

## Timeline
- 10:30:00 - Manual change made via Console
- 10:30:45 - TFDrift alert fired
- 10:31:00 - On-call engineer acknowledged
- 10:35:00 - Change identified as unauthorized
- 10:40:00 - Terraform reapplied, change reverted
- 10:45:00 - Incident resolved

## Root Cause
Engineer made emergency change via Console without following change management process.

## Resolution
1. Reverted unauthorized firewall rule change
2. Reminded team of change management policy
3. Updated documentation

## Action Items
- [ ] Add pre-commit hook to validate Terraform
- [ ] Send change management reminder to team
- [ ] Review firewall rule IAM permissions
```

#### Documentation Standards

**Maintain comprehensive documentation:**

```
docs/
├── runbooks/
│   ├── drift-response.md          # How to respond to drift alerts
│   ├── emergency-procedures.md    # Emergency access procedures
│   └── escalation.md              # Escalation paths
├── architecture/
│   ├── infrastructure-overview.md
│   ├── network-topology.md
│   └── security-controls.md
└── operations/
    ├── deployment-process.md
    ├── monitoring-guide.md
    └── troubleshooting.md
```

**Runbook Example:**
```markdown
# Drift Alert Response Runbook

## Critical IAM Change Alert

**Alert:** `google_project_iam_member` drift detected

### Step 1: Immediate Assessment (< 2 minutes)
1. Check alert details:
   - Who made the change? (principalEmail)
   - What role was granted/revoked?
   - Which project?

2. Verify if authorized:
   - Check change management tickets
   - Contact user if possible
   - Check recent approvals

### Step 2: Containment (< 5 minutes)
If unauthorized:
```bash
# Revoke unauthorized permission immediately
gcloud projects remove-iam-policy-binding PROJECT_ID \
  --member="user:suspicious@example.com" \
  --role="roles/editor"
```

### Step 3: Investigation (< 30 minutes)
1. Review GCP Audit Logs
2. Check for related changes
3. Interview user if needed
4. Document findings

### Step 4: Remediation
Choose appropriate action:
- Revert change via Terraform
- Update Terraform to match (if authorized)
- Escalate to security team

### Step 5: Follow-up
- Update incident tracker
- Notify stakeholders
- Schedule post-mortem if needed
```

---

### 5. Performance & Scalability

#### Event Filtering

**Filter events at multiple levels:**

**Level 1: GCP Log Sink (earliest, most efficient)**
```bash
# Only forward compute and IAM events
gcloud logging sinks update tfdrift-sink \
  --log-filter='
    (protoPayload.serviceName="compute.googleapis.com" OR
     protoPayload.serviceName="iam.googleapis.com") AND
    protoPayload.methodName!~"\.get$|\.list$"
  '
```

**Level 2: Falco Rules (plugin level)**
```yaml
# falco-rules.yaml
- rule: Relevant GCP Changes
  condition: >
    gcp.serviceName in (compute.googleapis.com, iam.googleapis.com, sqladmin.googleapis.com) and
    gcp.methodName matches "(insert|update|patch|delete|set.*)" and
    not gcp.methodName matches "(list|get|describe)"
  output: Relevant GCP change detected
  priority: NOTICE
```

**Level 3: TFDrift Configuration (application level)**
```yaml
# config.yaml
filtering:
  # Ignore read-only operations
  exclude_events:
    - "*.list"
    - "*.get"
    - "*.describe"
    - "*.testIamPermissions"

  # Focus on write operations
  include_methods:
    - "*.insert"
    - "*.update"
    - "*.patch"
    - "*.delete"
    - "*.set*"

  # Only monitor specific services
  include_services:
    - "compute.googleapis.com"
    - "container.googleapis.com"
    - "sqladmin.googleapis.com"
    - "iam.googleapis.com"

  # Ignore automated service accounts
  exclude_principals:
    - "serviceAccount:*@cloudbuild.gserviceaccount.com"
    - "serviceAccount:*@cloudservices.gserviceaccount.com"
```

#### Resource Limits

**Set appropriate limits:**

```yaml
# config.yaml
performance:
  # Maximum events to process per second
  max_events_per_second: 100

  # Maximum concurrent state file loads
  max_concurrent_state_loads: 5

  # Event queue size
  event_queue_size: 10000

  # Worker pool size
  worker_threads: 10

  # Timeout for state backend operations
  state_backend_timeout: "30s"

  # Memory limits
  max_memory_mb: 2048
```

**Docker resource limits:**
```yaml
# docker-compose.yml
services:
  tfdrift-falco:
    image: ghcr.io/higakikeita/tfdrift-falco:v0.5.0
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 2G
        reservations:
          cpus: '1.0'
          memory: 1G
```

#### Load Balancing

**For high-volume environments, run multiple instances:**

```yaml
# docker-compose.yml
services:
  tfdrift-falco-1:
    image: ghcr.io/higakikeita/tfdrift-falco:v0.5.0
    environment:
      - INSTANCE_ID=1
      - PROJECT_FILTER=project-a,project-b
    configs:
      - config-instance-1.yaml

  tfdrift-falco-2:
    image: ghcr.io/higakikeita/tfdrift-falco:v0.5.0
    environment:
      - INSTANCE_ID=2
      - PROJECT_FILTER=project-c,project-d
    configs:
      - config-instance-2.yaml
```

---

### 6. Security

#### IAM Best Practices

**Principle of Least Privilege:**
```bash
# ✅ Good: Minimal permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"  # Read-only for state

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"  # Only subscriber, not admin
```

```bash
# ❌ Bad: Overly broad permissions
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/editor"  # Too broad!
```

**Custom Role for TFDrift:**
```bash
# Create custom role with minimal permissions
gcloud iam roles create tfdriftFalcoRole \
  --project=$PROJECT_ID \
  --title="TFDrift Falco Custom Role" \
  --description="Minimal permissions for TFDrift-Falco" \
  --permissions="\
storage.objects.get,\
storage.objects.list,\
pubsub.subscriptions.consume,\
pubsub.subscriptions.get,\
logging.logEntries.list" \
  --stage=GA

# Assign custom role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@$PROJECT_ID.iam.gserviceaccount.com" \
  --role="projects/$PROJECT_ID/roles/tfdriftFalcoRole"
```

#### Network Security

**Restrict Falco gRPC access:**
```bash
# Only allow TFDrift host to connect to Falco
gcloud compute firewall-rules create allow-tfdrift-to-falco \
  --network=prod-network \
  --allow=tcp:5060 \
  --source-ranges=10.0.1.10/32 \  # TFDrift-Falco IP
  --target-tags=falco-server \
  --description="Allow TFDrift to connect to Falco gRPC"
```

**Use TLS for gRPC:**
```yaml
# config.yaml
falco:
  enabled: true
  hostname: "falco.internal.company.com"
  port: 5060
  tls:
    enabled: true
    ca_cert: "/etc/tfdrift/certs/ca.crt"
    client_cert: "/etc/tfdrift/certs/client.crt"
    client_key: "/etc/tfdrift/certs/client.key"
    verify_server: true
```

#### Audit Logging

**Enable audit logs for TFDrift itself:**
```yaml
# config.yaml
audit:
  enabled: true
  log_file: "/var/log/tfdrift/audit.log"
  log_format: "json"
  log_events:
    - "drift_detected"
    - "state_loaded"
    - "notification_sent"
    - "config_changed"
    - "startup"
    - "shutdown"
```

**Sample audit log entry:**
```json
{
  "timestamp": "2025-12-17T10:30:45Z",
  "event_type": "drift_detected",
  "severity": "critical",
  "user": "user@example.com",
  "resource": "google_compute_firewall.prod_allow_ssh",
  "project": "company-prod-123",
  "changes": {
    "attribute": "source_ranges",
    "old_value": ["10.0.0.0/8"],
    "new_value": ["0.0.0.0/0"]
  },
  "action_taken": "alert_sent",
  "alert_channels": ["slack", "pagerduty"]
}
```

---

### 7. Cost Optimization

#### Log Retention Policies

**Set appropriate retention:**
```bash
# Production: 90 days
gcloud logging sinks update tfdrift-sink \
  --log-filter='...' \
  --retention-days=90

# Development: 7 days
gcloud logging sinks update tfdrift-sink-dev \
  --log-filter='...' \
  --retention-days=7
```

**Cost comparison:**
| Retention | Daily Logs (GB) | Monthly Cost (est.) |
|-----------|-----------------|---------------------|
| 7 days | 10 GB | $5 |
| 30 days | 10 GB | $20 |
| 90 days | 10 GB | $60 |
| 365 days | 10 GB | $240 |

#### Storage Optimization

**Enable lifecycle policies on state bucket:**
```bash
# Delete old state versions after 30 days
cat > lifecycle.json <<EOF
{
  "lifecycle": {
    "rule": [
      {
        "action": {
          "type": "Delete"
        },
        "condition": {
          "age": 30,
          "numNewerVersions": 5
        }
      }
    ]
  }
}
EOF

gsutil lifecycle set lifecycle.json gs://company-terraform-state
```

#### Resource Efficiency

**Right-size Falco deployment:**
```yaml
# Small environment (< 10 projects, < 1000 events/day)
falco:
  resources:
    cpu: "500m"
    memory: "512Mi"

# Medium environment (10-50 projects, 1000-10000 events/day)
falco:
  resources:
    cpu: "1000m"
    memory: "1Gi"

# Large environment (50+ projects, 10000+ events/day)
falco:
  resources:
    cpu: "2000m"
    memory: "2Gi"
```

---

### 8. Disaster Recovery

#### Backup Strategy

**Backup TFDrift configuration:**
```bash
#!/bin/bash
# backup-tfdrift-config.sh

BACKUP_BUCKET="gs://company-backups/tfdrift"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Backup configuration files
gsutil cp -r /etc/tfdrift/config.yaml \
  "${BACKUP_BUCKET}/config_${TIMESTAMP}.yaml"

# Backup service account keys
gsutil cp -r /etc/tfdrift/keys/*.json \
  "${BACKUP_BUCKET}/keys/${TIMESTAMP}/"

# Backup Falco configuration
gsutil cp -r /etc/falco/*.yaml \
  "${BACKUP_BUCKET}/falco/${TIMESTAMP}/"

echo "✓ Backup completed: ${BACKUP_BUCKET}/${TIMESTAMP}"
```

#### Recovery Procedures

**Quick recovery script:**
```bash
#!/bin/bash
# recover-tfdrift.sh

BACKUP_BUCKET="gs://company-backups/tfdrift"
BACKUP_DATE=$1  # Format: 20251217_103000

echo "==> Recovering TFDrift from backup: $BACKUP_DATE"

# Restore configuration
gsutil cp "${BACKUP_BUCKET}/config_${BACKUP_DATE}.yaml" \
  /etc/tfdrift/config.yaml

# Restore keys
gsutil cp -r "${BACKUP_BUCKET}/keys/${BACKUP_DATE}/*" \
  /etc/tfdrift/keys/

# Restore Falco config
gsutil cp -r "${BACKUP_BUCKET}/falco/${BACKUP_DATE}/*" \
  /etc/falco/

# Restart services
docker-compose restart falco
docker-compose restart tfdrift-falco

echo "✓ Recovery completed"
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
