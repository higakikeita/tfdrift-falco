# TFDrift-Falco Use Cases

This document provides detailed, real-world use cases for TFDrift-Falco with concrete examples and implementation patterns.

## Table of Contents

- [Security & Compliance](#security--compliance)
- [Cost Management](#cost-management)
- [Audit & Governance](#audit--governance)
- [GitOps Enforcement](#gitops-enforcement)
- [Incident Response](#incident-response)
- [Multi-Cloud Scenarios](#multi-cloud-scenarios)

---

## Security & Compliance

### Use Case 1: Detecting Unauthorized Security Group Changes

**Scenario**: A developer accidentally opens port 22 (SSH) to 0.0.0.0/0 via AWS Console, bypassing your IaC workflow and creating a security vulnerability.

**Detection Flow**:
```
1. Developer modifies security group in AWS Console
   ↓
2. CloudTrail captures AuthorizeSecurityGroupIngress event
   ↓
3. Falco CloudTrail plugin processes event
   ↓
4. TFDrift-Falco compares with Terraform state
   ↓
5. Drift detected: ingress rule added (port 22, 0.0.0.0/0)
   ↓
6. Critical alert sent to Slack #security-alerts within 5 seconds
```

**Configuration**:
```yaml
# config.yaml
drift_rules:
  - name: "Security Group Ingress Rule Violation"
    resource_types:
      - "aws_security_group"
      - "aws_security_group_rule"
    watched_attributes:
      - "ingress"
      - "cidr_blocks"
    severity: "critical"

notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_SECURITY_WEBHOOK}"
    channel: "#security-alerts"

  webhook:
    enabled: true
    url: "https://your-siem.example.com/security-events"
    headers:
      Authorization: "Bearer ${SIEM_TOKEN}"
```

**Alert Example**:
```
🚨 CRITICAL DRIFT DETECTED

Resource: aws_security_group.web_servers
Change: Unauthorized ingress rule added

Details:
  Protocol: tcp
  Port: 22 (SSH)
  Source: 0.0.0.0/0 (PUBLIC INTERNET)

User: developer@example.com
Time: 2025-01-20T14:32:15Z
Region: us-east-1

Action Required:
  1. Immediately revert change in AWS Console
  2. Contact developer for explanation
  3. Run `terraform plan` to sync state
  4. Update Terraform code if change is intentional
```

**Integration with Security Tools**:
```yaml
# Integration with Sysdig Secure
notifications:
  webhook:
    enabled: true
    url: "https://secure.sysdig.com/api/v1/events"
    headers:
      Authorization: "Bearer ${SYSDIG_TOKEN}"
    payload_template: |
      {
        "event": {
          "name": "Security Group Exposure Detected",
          "description": "Port 22 opened to 0.0.0.0/0",
          "severity": "high",
          "scope": "{{ .ResourceID }}",
          "tags": {
            "user": "{{ .User }}",
            "resource_type": "{{ .ResourceType }}"
          }
        }
      }
```

---

### Use Case 2: IAM Policy Drift Detection

**Scenario**: Someone adds excessive permissions to an IAM role outside of Terraform, violating least-privilege principle.

**Terraform State**:
```hcl
resource "aws_iam_role_policy" "app_policy" {
  role = aws_iam_role.app_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["s3:GetObject"]
        Resource = "arn:aws:s3:::my-app-bucket/*"
      }
    ]
  })
}
```

**Manual Change** (via AWS Console):
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["s3:*"],  # Changed from GetObject to wildcard!
      "Resource": "*"       # Changed from specific bucket to wildcard!
    }
  ]
}
```

**TFDrift-Falco Alert**:
```json
{
  "event_type": "terraform_drift_detected",
  "provider": "aws",
  "resource_type": "aws_iam_role_policy",
  "resource_id": "app_role:app_policy",
  "severity": "critical",
  "change_type": "modified",
  "detected_at": "2025-01-20T10:15:30Z",
  "user": "admin@example.com",
  "cloudtrail_event": "PutRolePolicy",
  "changes": {
    "policy": {
      "expected": "{\"Statement\":[{\"Action\":[\"s3:GetObject\"],\"Resource\":\"arn:aws:s3:::my-app-bucket/*\"}]}",
      "actual": "{\"Statement\":[{\"Action\":[\"s3:*\"],\"Resource\":\"*\"}]}",
      "diff": "Excessive permissions granted: s3:GetObject → s3:* on all resources"
    }
  }
}
```

**Automated Response** (via webhook):
```yaml
# config.yaml
notifications:
  webhook:
    enabled: true
    url: "https://your-api.example.com/iam-violation"
    headers:
      Authorization: "Bearer ${API_TOKEN}"
      X-Action: "create-incident"
```

---

### Use Case 3: Encryption Settings Disabled

**Scenario**: S3 bucket encryption is disabled manually, violating compliance requirements (SOC2, HIPAA, etc.).

**Detection**:
```yaml
drift_rules:
  - name: "S3 Encryption Disabled"
    resource_types:
      - "aws_s3_bucket"
    watched_attributes:
      - "server_side_encryption_configuration"
    severity: "critical"
    tags:
      - "compliance"
      - "encryption"
      - "soc2"
```

**Alert Integration with PagerDuty**:
```yaml
notifications:
  webhook:
    enabled: true
    url: "https://events.pagerduty.com/v2/enqueue"
    headers:
      Authorization: "Token token=${PAGERDUTY_TOKEN}"
      Content-Type: "application/json"
    payload_template: |
      {
        "routing_key": "${PAGERDUTY_ROUTING_KEY}",
        "event_action": "trigger",
        "payload": {
          "summary": "S3 Encryption Disabled - {{ .ResourceID }}",
          "severity": "critical",
          "source": "tfdrift-falco",
          "custom_details": {
            "resource": "{{ .ResourceID }}",
            "user": "{{ .User }}",
            "change": "{{ .ChangeType }}"
          }
        }
      }
```

---

## Cost Management

### Use Case 4: Detecting Unexpected Instance Type Changes

**Scenario**: An EC2 instance is upgraded from t3.micro ($0.0104/hr) to m5.8xlarge ($1.536/hr), increasing monthly costs from $7.50 to $1,107.

**Configuration**:
```yaml
drift_rules:
  - name: "EC2 Instance Type Change"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
    severity: "high"

  - name: "High-Cost Instance Upgrade"
    resource_types:
      - "aws_instance"
    filters:
      # Detect upgrades to expensive instance families
      instance_type_pattern: "^(m5|c5|r5)\\.(8|12|16|24)xlarge$"
    severity: "critical"
```

**Cost Impact Analysis Alert**:
```
⚠️ HIGH-COST DRIFT DETECTED

Resource: aws_instance.app_server
Change: instance_type = "t3.micro" → "m5.8xlarge"

Cost Impact:
  Before: $7.50/month ($0.0104/hr)
  After: $1,107/month ($1.536/hr)
  Increase: +$1,099.50/month (+14,660%)

User: devops-team@example.com
Justification: [Request justification]

Actions:
  1. Confirm if intentional upgrade
  2. Evaluate if m5.4xlarge ($0.768/hr) is sufficient
  3. Update Terraform code if approved
  4. Set budget alert for this resource
```

**Integration with Cloud Cost Management**:
```yaml
notifications:
  webhook:
    enabled: true
    url: "https://cost-management.example.com/api/alerts"
    headers:
      Authorization: "Bearer ${COST_MGMT_TOKEN}"
    payload_template: |
      {
        "alert_type": "unexpected_cost_increase",
        "resource_id": "{{ .ResourceID }}",
        "estimated_monthly_increase": 1099.50,
        "user": "{{ .User }}",
        "timestamp": "{{ .DetectedAt }}"
      }
```

---

### Use Case 5: Storage Volume Expansion

**Scenario**: EBS volume size is manually increased from 100GB to 1TB, increasing storage costs.

**Alert**:
```
💰 Storage Cost Drift Detected

Resource: aws_ebs_volume.app_data
Change: size = 100 GB → 1000 GB

Cost Impact (GP3):
  Before: $8.00/month ($0.08/GB)
  After: $80.00/month
  Increase: +$72.00/month (+900%)

User: backend-team@example.com
CloudTrail Event: ModifyVolume
Time: 2025-01-20T16:45:00Z

Recommendation:
  - Verify data growth justifies expansion
  - Consider using lifecycle policies for old data
  - Evaluate cheaper storage tiers (S3, EFS)
```

---

## Audit & Governance

### Use Case 6: Comprehensive Change Tracking

**Scenario**: Track all infrastructure changes with complete audit trail for compliance reporting (SOC2, ISO 27001, etc.).

**Configuration**:
```yaml
# Export all drift events to audit log system
output:
  mode: "json"  # NDJSON format for log aggregation

notifications:
  webhook:
    enabled: true
    url: "https://audit-logs.example.com/api/events"
    headers:
      Authorization: "Bearer ${AUDIT_TOKEN}"
      X-Event-Type: "infrastructure-change"
```

**Audit Log Entry**:
```json
{
  "event_id": "evt_1234567890abcdef",
  "event_type": "terraform_drift_detected",
  "timestamp": "2025-01-20T12:00:00Z",
  "provider": "aws",
  "account_id": "123456789012",
  "region": "us-east-1",
  "resource_type": "aws_iam_role",
  "resource_id": "arn:aws:iam::123456789012:role/AdminRole",
  "change_type": "modified",
  "severity": "high",

  "user_identity": {
    "type": "IAMUser",
    "principal_id": "AIDACKCEVSQ6C2EXAMPLE",
    "arn": "arn:aws:iam::123456789012:user/admin",
    "account_id": "123456789012",
    "username": "admin@example.com"
  },

  "changes": {
    "assume_role_policy": {
      "expected": "...",
      "actual": "...",
      "diff": "Added new trusted entity: 111122223333"
    }
  },

  "cloudtrail": {
    "event_id": "a1b2c3d4-5678-90ab-cdef-1234567890ab",
    "event_name": "UpdateAssumeRolePolicy",
    "source_ip": "203.0.113.42",
    "user_agent": "console.amazonaws.com"
  },

  "terraform": {
    "workspace": "production",
    "module": "modules/iam",
    "file": "roles.tf",
    "line": 45
  }
}
```

**Compliance Reporting Query** (via SIEM):
```sql
-- Query all IAM changes in last 30 days
SELECT
  timestamp,
  user_identity.username,
  resource_id,
  change_type,
  changes
FROM audit_logs
WHERE event_type = 'terraform_drift_detected'
  AND resource_type LIKE 'aws_iam_%'
  AND timestamp >= NOW() - INTERVAL '30 days'
ORDER BY timestamp DESC;
```

---

### Use Case 7: Multi-Account Governance

**Scenario**: Monitor multiple AWS accounts (dev, staging, prod) from a central security account.

**Architecture**:
```
┌─────────────────┐
│  Security Acct  │
│  (Monitoring)   │
│                 │
│  Falco          │
│  TFDrift-Falco  │
└────────┬────────┘
         │
         ├──────────┬──────────┬──────────┐
         │          │          │          │
    ┌────▼────┐ ┌───▼────┐ ┌───▼────┐ ┌───▼────┐
    │  Dev    │ │ Staging│ │  Prod  │ │  Prod  │
    │ Account │ │ Account│ │ Account│ │ Account│
    │         │ │        │ │  (US)  │ │  (EU)  │
    └─────────┘ └────────┘ └────────┘ └────────┘
       123         456         789         101
```

**Configuration**:
```yaml
# config.yaml - Multi-account monitoring
providers:
  aws:
    enabled: true
    accounts:
      - account_id: "123456789012"
        name: "development"
        regions: ["us-east-1"]
        state:
          backend: "s3"
          s3_bucket: "dev-terraform-state"
          s3_key: "terraform.tfstate"
          role_arn: "arn:aws:iam::123456789012:role/TFDriftMonitor"

      - account_id: "456789012345"
        name: "staging"
        regions: ["us-east-1"]
        state:
          backend: "s3"
          s3_bucket: "staging-terraform-state"
          s3_key: "terraform.tfstate"
          role_arn: "arn:aws:iam::456789012345:role/TFDriftMonitor"

      - account_id: "789012345678"
        name: "production-us"
        regions: ["us-east-1", "us-west-2"]
        state:
          backend: "s3"
          s3_bucket: "prod-us-terraform-state"
          s3_key: "terraform.tfstate"
          role_arn: "arn:aws:iam::789012345678:role/TFDriftMonitor"

      - account_id: "101112131415"
        name: "production-eu"
        regions: ["eu-west-1", "eu-central-1"]
        state:
          backend: "s3"
          s3_bucket: "prod-eu-terraform-state"
          s3_key: "terraform.tfstate"
          role_arn: "arn:aws:iam::101112131415:role/TFDriftMonitor"

notifications:
  slack:
    enabled: true
    channels:
      - name: "#dev-alerts"
        filter:
          accounts: ["123456789012"]
      - name: "#staging-alerts"
        filter:
          accounts: ["456789012345"]
      - name: "#prod-critical"
        filter:
          accounts: ["789012345678", "101112131415"]
          severity: ["critical", "high"]
```

---

## GitOps Enforcement

### Use Case 8: Enforcing Infrastructure-as-Code Discipline

**Scenario**: Ensure ALL infrastructure changes go through Git workflow (branch → PR → review → merge → CI/CD → apply).

**Workflow**:
```
❌ Console Change (Blocked Pattern)
   Developer → AWS Console → Modify Resource
                                    ↓
                            TFDrift Alert
                                    ↓
                            Auto-Revert (Optional)

✅ Proper GitOps Workflow
   Developer → Git Branch → Terraform Code → PR
                                            ↓
                                    Code Review
                                            ↓
                                    Merge to Main
                                            ↓
                                    CI/CD Pipeline
                                            ↓
                                    terraform apply
                                            ↓
                              No Drift (Expected Change)
```

**Configuration**:
```yaml
drift_rules:
  - name: "Console Change Detection"
    resource_types:
      - "*"  # Monitor all resources
    exclude_users:
      - "terraform-ci-cd@example.com"  # Allow CI/CD user
      - "arn:aws:iam::123456789012:role/TerraformRole"
    severity: "high"
    actions:
      - type: "alert"
        channels: ["slack", "email"]
      - type: "create_ticket"
        system: "jira"
        project: "INFRA"
        issue_type: "Task"
        summary: "Manual Infrastructure Change Detected"
      # - type: "auto_revert"  # Optional: automatic rollback
      #   enabled: false

notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK}"
    message_template: |
      ⛔ GitOps Violation Detected

      User {{ .User }} made manual change to {{ .ResourceType }}

      This violates our Infrastructure-as-Code policy.
      All changes must go through Git workflow.

      Resource: {{ .ResourceID }}
      Change: {{ .ChangeType }}
      Time: {{ .DetectedAt }}

      Action Required:
      1. Revert the manual change
      2. Create Git branch with equivalent Terraform code
      3. Submit PR for review
      4. Contact #platform-team for guidance
```

---

## Incident Response

### Use Case 9: Real-Time Security Incident Detection

**Scenario**: Detect potential security breach when suspicious infrastructure changes occur (e.g., backdoor security group rules, IAM role escalation).

**Detection Rules**:
```yaml
drift_rules:
  # Backdoor Detection
  - name: "Potential Backdoor - Public SSH Access"
    resource_types:
      - "aws_security_group"
    conditions:
      - attribute: "ingress"
        pattern: "port=22.*cidr_blocks.*0\\.0\\.0\\.0/0"
    severity: "critical"
    tags:
      - "security-incident"
      - "potential-breach"
    actions:
      - type: "alert"
        priority: "p1"
      - type: "webhook"
        url: "https://incident-response.example.com/trigger"

  # Privilege Escalation Detection
  - name: "IAM Privilege Escalation Attempt"
    resource_types:
      - "aws_iam_role_policy"
      - "aws_iam_user_policy"
    conditions:
      - attribute: "policy"
        contains: "iam:PassRole"
      - attribute: "policy"
        contains: "sts:AssumeRole"
    severity: "critical"
    tags:
      - "security-incident"
      - "privilege-escalation"

  # Data Exfiltration Risk
  - name: "S3 Bucket Made Public"
    resource_types:
      - "aws_s3_bucket_public_access_block"
    conditions:
      - attribute: "block_public_acls"
        value: false
    severity: "critical"
    tags:
      - "security-incident"
      - "data-exfiltration-risk"
```

**Incident Response Integration**:
```yaml
notifications:
  # Immediate PagerDuty alert for security incidents
  webhook:
    enabled: true
    url: "https://events.pagerduty.com/v2/enqueue"
    headers:
      Authorization: "Token token=${PAGERDUTY_TOKEN}"
    filter:
      tags: ["security-incident"]
    payload_template: |
      {
        "routing_key": "${PAGERDUTY_SECURITY_KEY}",
        "event_action": "trigger",
        "dedup_key": "{{ .ResourceID }}-{{ .DetectedAt }}",
        "payload": {
          "summary": "🚨 SECURITY INCIDENT: {{ .ResourceType }} modified",
          "severity": "critical",
          "source": "tfdrift-falco",
          "custom_details": {
            "resource": "{{ .ResourceID }}",
            "user": "{{ .User }}",
            "change": "{{ .ChangeType }}",
            "tags": {{ .Tags }},
            "cloudtrail_event": "{{ .CloudTrailEvent }}",
            "source_ip": "{{ .SourceIP }}"
          }
        }
      }

  # Create SIEM event
  webhook:
    enabled: true
    url: "https://siem.example.com/api/security-events"
    headers:
      Authorization: "Bearer ${SIEM_TOKEN}"
      X-Event-Category: "infrastructure-security"
    filter:
      tags: ["security-incident"]
```

---

## Multi-Cloud Scenarios

### Use Case 10: Hybrid AWS + GCP Infrastructure Monitoring

**Scenario**: Monitor both AWS and GCP resources from a single TFDrift-Falco instance.

**Configuration**:
```yaml
# config.yaml - Multi-cloud monitoring
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "aws/terraform.tfstate"

  gcp:
    enabled: true
    projects:
      - my-project-123
      - my-project-456
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "gcp"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  # AWS Rules
  - name: "AWS Security Group Change"
    resource_types:
      - "aws_security_group"
    watched_attributes:
      - "ingress"
      - "egress"
    severity: "high"

  # GCP Rules
  - name: "GCP Firewall Rule Change"
    resource_types:
      - "google_compute_firewall"
    watched_attributes:
      - "allowed"
      - "denied"
      - "source_ranges"
    severity: "high"

  # Cross-Cloud Policy
  - name: "Public Internet Exposure (Any Cloud)"
    resource_types:
      - "aws_security_group"
      - "google_compute_firewall"
    conditions:
      - attribute: "*"
        pattern: "0\\.0\\.0\\.0/0"
    severity: "critical"

notifications:
  slack:
    enabled: true
    channels:
      - name: "#aws-alerts"
        filter:
          provider: "aws"
      - name: "#gcp-alerts"
        filter:
          provider: "gcp"
      - name: "#multi-cloud-security"
        filter:
          severity: ["critical"]
```

**Multi-Cloud Alert**:
```
🌐 Multi-Cloud Drift Detected

AWS:
  Resource: aws_security_group.web
  Change: Port 80 opened to 0.0.0.0/0
  User: aws-admin@example.com
  Region: us-east-1

GCP:
  Resource: google_compute_firewall.web
  Change: Port 443 opened to 0.0.0.0/0
  User: gcp-admin@example.com
  Project: my-project-123

Similar changes detected across both clouds.
Possible coordinated attack or policy violation.

Action: Review both changes immediately.
```

---

## Summary

TFDrift-Falco provides comprehensive drift detection across multiple scenarios:

| Use Case | Detection Time | Severity | Integration |
|----------|---------------|----------|-------------|
| Security Group Changes | < 5 seconds | Critical | Slack, SIEM, PagerDuty |
| IAM Policy Drift | < 10 seconds | Critical | SIEM, Audit Logs |
| Cost Changes | < 15 seconds | High | Cost Management Tools |
| Compliance Violations | Real-time | Variable | Audit Systems, Compliance Platforms |
| GitOps Enforcement | Real-time | High | Slack, Email, Ticketing |
| Security Incidents | < 3 seconds | Critical | PagerDuty, SOAR, SIEM |
| Multi-Cloud | Real-time | Variable | Unified Dashboard, SIEM |

---

## Next Steps

- [Best Practices](BEST_PRACTICES.md) - Operational best practices for production deployments
- [Extending TFDrift-Falco](EXTENDING.md) - Add custom rules and notification channels
- [Architecture](architecture.md) - Detailed system architecture
