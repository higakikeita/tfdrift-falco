# GCP Compute Engine

> **Service:** Google Compute Engine (GCE)
> **Events Monitored:** 11+
> **Resources:** `google_compute_instance`
> **Status:** ✅ Production Ready

---

## Overview

TFDrift-Falco monitors Google Compute Engine instance changes through GCP Audit Logs, detecting configuration drift in real-time.

---

## Monitored Events

### Instance Lifecycle

| Event Name | Description | Resource Type |
|------------|-------------|---------------|
| `compute.instances.insert` | Instance creation | `google_compute_instance` |
| `compute.instances.delete` | Instance deletion | `google_compute_instance` |
| `compute.instances.start` | Instance start | `google_compute_instance` |
| `compute.instances.stop` | Instance stop | `google_compute_instance` |
| `compute.instances.reset` | Instance reset | `google_compute_instance` |

### Configuration Changes

| Event Name | Description | Resource Type |
|------------|-------------|---------------|
| `compute.instances.setMetadata` | Metadata modification | `google_compute_instance` |
| `compute.instances.setLabels` | Labels modification | `google_compute_instance` |
| `compute.instances.setTags` | Network tags modification | `google_compute_instance` |
| `compute.instances.setMachineType` | Machine type change | `google_compute_instance` |
| `compute.instances.setServiceAccount` | Service account change | `google_compute_instance` |
| `compute.instances.setDeletionProtection` | Deletion protection toggle | `google_compute_instance` |

---

## Example Drift Scenarios

### Scenario 1: Unauthorized Metadata Change

**Terraform State:**
```hcl
resource "google_compute_instance" "web_server" {
  name         = "web-server-01"
  machine_type = "n1-standard-1"
  zone         = "us-central1-a"

  metadata = {
    ssh-keys = "admin:ssh-rsa AAAAB3..."
    env      = "production"
  }
}
```

**Manual Change:**
```bash
gcloud compute instances add-metadata web-server-01 \
  --zone=us-central1-a \
  --metadata=debug=true
```

**TFDrift-Falco Alert:**
```json
{
  "severity": "high",
  "resource_type": "google_compute_instance",
  "resource_id": "web-server-01",
  "event_name": "compute.instances.setMetadata",
  "changes": {
    "metadata": {
      "added": ["debug"]
    }
  },
  "user": "admin@example.com",
  "project": "my-project-123",
  "zone": "us-central1-a"
}
```

### Scenario 2: Machine Type Change

**Terraform State:**
```hcl
resource "google_compute_instance" "app_server" {
  name         = "app-server-01"
  machine_type = "n1-standard-2"
  zone         = "us-central1-a"
}
```

**Manual Change:**
```bash
gcloud compute instances set-machine-type app-server-01 \
  --zone=us-central1-a \
  --machine-type=n1-standard-4
```

**TFDrift-Falco Alert:**
```json
{
  "severity": "critical",
  "resource_type": "google_compute_instance",
  "resource_id": "app-server-01",
  "event_name": "compute.instances.setMachineType",
  "changes": {
    "machine_type": {
      "old": "n1-standard-2",
      "new": "n1-standard-4"
    }
  },
  "user": "devops@example.com",
  "project": "my-project-123",
  "zone": "us-central1-a"
}
```

### Scenario 3: Deletion Protection Disabled

**Terraform State:**
```hcl
resource "google_compute_instance" "database_server" {
  name                = "db-server-01"
  machine_type        = "n1-highmem-4"
  zone                = "us-central1-a"
  deletion_protection = true
}
```

**Manual Change:**
```bash
gcloud compute instances update db-server-01 \
  --zone=us-central1-a \
  --no-deletion-protection
```

**TFDrift-Falco Alert:**
```json
{
  "severity": "critical",
  "resource_type": "google_compute_instance",
  "resource_id": "db-server-01",
  "event_name": "compute.instances.setDeletionProtection",
  "changes": {
    "deletion_protection": {
      "old": true,
      "new": false
    }
  },
  "user": "admin@example.com",
  "project": "my-project-123",
  "zone": "us-central1-a"
}
```

---

## Configuration

### Basic Drift Rule

```yaml
drift_rules:
  - name: "GCE Instance Metadata Change"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "metadata"
    severity: "high"
    actions:
      - type: "alert"
        channels: ["slack"]
```

### Advanced Drift Rule

```yaml
drift_rules:
  - name: "GCE Instance Critical Changes"
    resource_types:
      - "google_compute_instance"
    conditions:
      - attribute: "deletion_protection"
        operator: "changed"
        from: true
        to: false
      - attribute: "machine_type"
        operator: "changed"
    severity: "critical"
    filters:
      - user_identity:
          principal_email: "*-terraform@*.iam.gserviceaccount.com"
        action: "skip"
    actions:
      - type: "alert"
        channels: ["slack", "pagerduty"]
      - type: "webhook"
        url: "https://webhook.example.com/gcp-drift"
```

---

## Terraform State Mapping

### Attributes Monitored

| Terraform Attribute | GCP Audit Log Field | Event Type |
|---------------------|---------------------|------------|
| `name` | `gcp.resource.name` | `*.insert`, `*.delete` |
| `machine_type` | `gcp.request.machineType` | `*.setMachineType` |
| `metadata` | `gcp.request.metadata` | `*.setMetadata` |
| `labels` | `gcp.request.labels` | `*.setLabels` |
| `tags` | `gcp.request.tags` | `*.setTags` |
| `service_account` | `gcp.request.serviceAccount` | `*.setServiceAccount` |
| `deletion_protection` | `gcp.request.deletionProtection` | `*.setDeletionProtection` |

### State Correlation

TFDrift-Falco extracts resource identifiers from GCP Audit Logs and correlates them with Terraform state:

```
GCP Audit Log:
  gcp.resource.name: "projects/my-project-123/zones/us-central1-a/instances/web-server-01"

Extracted:
  project: "my-project-123"
  zone: "us-central1-a"
  resource_id: "web-server-01"

Terraform State Match:
  google_compute_instance.web_server {
    name = "web-server-01"
    project = "my-project-123"
    zone = "us-central1-a"
  }
```

---

## Best Practices

### 1. Monitor Critical Attributes

Focus on attributes that impact security and cost:

```yaml
drift_rules:
  - name: "GCE Security-Critical Changes"
    resource_types:
      - "google_compute_instance"
    watched_attributes:
      - "deletion_protection"
      - "service_account"
      - "tags"  # Firewall rules
    severity: "critical"
```

### 2. Exclude Terraform Service Accounts

Prevent alerts for legitimate Terraform changes:

```yaml
drift_rules:
  - name: "GCE Manual Changes"
    resource_types:
      - "google_compute_instance"
    filters:
      - user_identity:
          principal_email: "terraform@my-project.iam.gserviceaccount.com"
        action: "skip"
```

### 3. Environment-Specific Rules

Different severity for different environments:

```yaml
drift_rules:
  - name: "GCE Prod Changes"
    resource_types:
      - "google_compute_instance"
    conditions:
      - attribute: "labels.env"
        operator: "equals"
        value: "production"
    severity: "critical"

  - name: "GCE Dev Changes"
    resource_types:
      - "google_compute_instance"
    conditions:
      - attribute: "labels.env"
        operator: "equals"
        value: "development"
    severity: "medium"
```

---

## Troubleshooting

### Issue: Instance changes not detected

**Cause:** Falco gcpaudit plugin not receiving events

**Solution:**
1. Verify Pub/Sub subscription is active:
   ```bash
   gcloud pubsub subscriptions describe falco-gcp-audit-logs
   ```

2. Check Falco logs:
   ```bash
   kubectl logs -n falco -l app=falco | grep gcpaudit
   ```

3. Verify Audit Logs configuration:
   ```bash
   gcloud logging sinks describe audit-logs-to-pubsub
   ```

### Issue: False positives from automated systems

**Cause:** Automated tools making legitimate changes

**Solution:** Add service account filters:

```yaml
drift_rules:
  - name: "GCE Instance Changes"
    resource_types:
      - "google_compute_instance"
    filters:
      - user_identity:
          principal_email: "*@cloudservices.gserviceaccount.com"
        action: "skip"
      - user_identity:
          principal_email: "*@developer.gserviceaccount.com"
        action: "skip"
```

---

## Related Services

- [VPC & Firewall](vpc.md) - Network configuration drift detection
- [Disks](disks.md) - Persistent disk changes
- [IAM](iam.md) - Service account and IAM policy changes

---

## Additional Resources

- [GCP Compute Engine Documentation](https://cloud.google.com/compute/docs)
- [Terraform google_compute_instance](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/compute_instance)
- [GCP Audit Logs](https://cloud.google.com/logging/docs/audit)

---

**Last Updated:** 2025-01-18
**Version:** v0.5.0
