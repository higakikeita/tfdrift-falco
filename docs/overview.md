# Overview

TFDrift-Falco is a **real-time Terraform drift detection system** that monitors AWS infrastructure changes and alerts when resources drift from their Terraform-defined state.

---

## The Problem

### Configuration Drift in Cloud Infrastructure

**Configuration drift** occurs when actual infrastructure state diverges from the intended state defined in Infrastructure as Code (IaC).

Common causes:
- 🖱️ Manual changes via AWS Console
- 🔧 Emergency hotfixes bypassing Terraform
- 🤖 Auto Scaling or AWS-managed updates
- 👥 Multiple teams managing the same resources

### Consequences of Drift

1. **Security Risks**
   - Untracked security group rule changes
   - IAM policy modifications
   - Encryption settings disabled

2. **Compliance Violations**
   - Resources no longer meet policy requirements
   - Audit trail inconsistencies
   - Failed compliance scans

3. **Operational Issues**
   - Terraform apply conflicts
   - Unpredictable behavior during deployments
   - Lost tribal knowledge of changes

4. **Cost Implications**
   - Untracked resource provisioning
   - Over-provisioned instances
   - Forgotten resources running indefinitely

---

## The TFDrift-Falco Solution

### How It Works

TFDrift-Falco solves drift detection by combining three powerful technologies:

1. **AWS CloudTrail**: Real-time event stream of all AWS API calls
2. **Terraform State**: Source of truth for intended infrastructure
3. **Falco**: Runtime security engine for alerting

```
AWS Change → CloudTrail Event → TFDrift Detection → Falco Alert → Action
```

### Key Capabilities

#### 1. Real-time Detection (Sub-minute)

Unlike scheduled drift detection tools (Terraform plan runs), TFDrift-Falco detects drift **immediately** when it happens.

**Traditional Approach:**
```
Change happens → Wait 1 hour → Terraform plan runs → Drift found
```

**TFDrift-Falco:**
```
Change happens → CloudTrail event → Drift detected (< 30s)
```

#### 2. Comprehensive Service Coverage

Monitors 150+ AWS CloudTrail events across 12 services:
- Compute: EC2
- Networking: VPC, Route53, CloudFront
- Storage: S3, EBS
- Databases: RDS, Aurora
- Security: IAM, KMS
- Application: API Gateway, SNS, SQS, ECR

[View Full Service Coverage →](services/ec2.md)

#### 3. Intelligent Change Detection

Not all CloudTrail events indicate drift. TFDrift-Falco intelligently distinguishes:

✅ **True Drift** (alert)
- Manual change to instance type via Console
- Security group rule added outside Terraform

❌ **Expected Changes** (no alert)
- Auto Scaling group scaling events
- AWS-managed updates (e.g., RDS maintenance)
- Changes made by Terraform itself

#### 4. User Attribution

Every alert includes:
- **Who**: IAM user or role that made the change
- **What**: Specific attribute(s) changed
- **When**: Timestamp
- **Where**: AWS region and resource ID

Example alert:
```
EC2 Instance Type Changed
- Resource: i-0123456789abcdef0
- Change: t3.micro → t3.small
- User: admin@example.com
- Time: 2025-12-06 07:30:00 UTC
```

---

## Architecture Components

### 1. TFDrift Detector

The core detection engine that:
- Polls CloudTrail for new events
- Fetches corresponding Terraform state
- Compares actual vs. desired state
- Emits drift events to Falco

**Technology:** Go (high performance, low memory)

### 2. Falco Rules

Falco rules define drift policies:
- Which events to monitor
- Severity levels (Critical / Error / Warning)
- Alert output format

Example rule:
```yaml
rule: ec2_instance_type_changed
condition:
  cloud.service = "ec2" and
  evt.name = "ModifyInstanceAttribute" and
  drift.attribute = "instance_type"
output: "EC2 Instance Type Changed (instance=%resource from=%drift.old_value to=%drift.new_value user=%user)"
priority: warning
```

### 3. Grafana Dashboards

Visual monitoring with:
- Service-specific drift timelines
- Resource-level drill-down
- User activity heatmaps
- Alert panels for critical drift

### 4. Alerting Integrations (v0.3.0)

Planned integrations:
- Slack: Channel notifications
- PagerDuty: Incident creation
- Email: SMTP alerts
- Webhook: Custom integrations

---

## Deployment Models

### 1. Single-Account Monitoring

Monitor one AWS account's infrastructure.

**Use case:** Small teams, single production account

```
AWS Account → CloudTrail → TFDrift → Falco
```

### 2. Multi-Account Centralized

Monitor multiple AWS accounts from a central monitoring account.

**Use case:** Organizations with 10+ accounts

```
AWS Account 1 → CloudTrail → \
AWS Account 2 → CloudTrail → → TFDrift (Central) → Falco
AWS Account 3 → CloudTrail → /
```

### 3. Multi-Region

Monitor resources across multiple AWS regions.

**Use case:** Global infrastructure deployments

---

## Comparison to Other Tools

### vs. Terraform Plan

| Feature | Terraform Plan | TFDrift-Falco |
|---------|----------------|---------------|
| Detection Speed | Scheduled (hourly) | Real-time (< 1 min) |
| Resource Consumption | High (full plan) | Low (event-based) |
| Alert Integration | Manual | Automated (Falco) |
| User Attribution | ❌ No | ✅ Yes |

### vs. AWS Config

| Feature | AWS Config | TFDrift-Falco |
|---------|-----------|---------------|
| Terraform-aware | ❌ No | ✅ Yes |
| Cost | $$$$ (per config item) | $ (compute only) |
| Custom Rules | Limited | Unlimited (Falco) |
| Real-time Alerts | ❌ No | ✅ Yes |

### vs. Cloud Custodian

| Feature | Cloud Custodian | TFDrift-Falco |
|---------|-----------------|---------------|
| Drift Detection | Policy-based | Terraform state-based |
| Real-time | ❌ No (Lambda scheduled) | ✅ Yes (event-driven) |
| Remediation | ✅ Yes | Planned (v0.3.0) |
| Complexity | High | Medium |

---

## When to Use TFDrift-Falco

### ✅ Good Fit

- You manage AWS infrastructure with Terraform
- You need **real-time** drift detection
- You have **multi-account** or **multi-region** setups
- You need **user attribution** for compliance
- You want **low-cost** solution (no per-resource charges)

### ⚠️ Consider Alternatives

- You don't use Terraform (use AWS Config instead)
- You need automatic remediation (use Cloud Custodian)
- You're okay with hourly drift checks (use scheduled Terraform plan)

---

## Next Steps

1. [Understand the Architecture →](architecture.md)
2. [Check Service Coverage →](services/ec2.md)
3. [Deploy TFDrift-Falco →](quickstart.md)
