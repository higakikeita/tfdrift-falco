# TFDrift-Falco Documentation

Welcome to the official documentation for **TFDrift-Falco**, a real-time multi-cloud Terraform drift detection system powered by Falco.

> **Version:** v0.5.0+ | **Status:** Production Ready | **Providers:** AWS + GCP

---

## What is TFDrift-Falco?

TFDrift-Falco detects when your cloud infrastructure changes outside of Terraform by:

1. **Monitoring cloud audit logs** in real-time (AWS CloudTrail, GCP Audit Logs)
2. **Comparing changes against Terraform state** (S3, GCS, or local)
3. **Alerting via Falco** when drift is detected
4. **Visualizing drift in Grafana** dashboards

---

## Key Features

### 🌐 Multi-Cloud Support (v0.5.0+)

#### AWS Coverage
Supports **203+ CloudTrail events** across **19 AWS services**:
- **Compute:** EC2, Lambda, Auto Scaling
- **Networking:** VPC, Security Groups, ELB/ALB
- **Storage:** S3
- **Databases:** RDS, Aurora, DynamoDB
- **Security:** IAM, KMS
- **Containers:** ECS, EKS, ECR
- **Serverless:** API Gateway
- **DNS & CDN:** Route53, CloudFront
- **Messaging:** SNS, SQS

[View AWS Service Coverage →](services/index.md)

#### GCP Coverage (v0.5.0+)
Supports **100+ Audit Log events** across **12+ GCP services**:
- **Compute:** Compute Engine, Disks
- **Networking:** VPC, Firewall, Routes, Routers
- **Storage:** Cloud Storage
- **Databases:** Cloud SQL
- **Security:** IAM, KMS, Secret Manager
- **Containers:** GKE, Cloud Run
- **Serverless:** Cloud Functions
- **Data & Analytics:** BigQuery, Pub/Sub

[View GCP Service Coverage →](services/gcp/index.md)

### ⚡ Real-time Detection

- **Sub-minute latency** from cloud change to alert
- Asynchronous audit log processing (CloudTrail, GCP Audit Logs)
- Parallel multi-cloud service detection
- Event-driven architecture with Falco

### 🔐 Security-Focused

- **IAM policy drift detection** (AWS IAM, GCP IAM)
- **Encryption configuration monitoring** (S3, KMS, Cloud Storage)
- **Firewall rule changes** (Security Groups, GCP Firewall)
- **Key management** (AWS KMS, GCP KMS)
- **Service account modifications** (GCP Service Accounts)

### 📊 Production-Ready Monitoring

- **Grafana dashboards** for multi-cloud visibility
- **Falco rules** with severity levels
- **User attribution** for every change (IAM principals, service accounts)
- **Alert integration** ready (Slack, Discord, webhooks)
- **Multi-cloud unified view** in single dashboard

---

## Quick Start

### Prerequisites

**For AWS:**
- AWS account with CloudTrail enabled
- Falco with cloudtrail plugin

**For GCP (v0.5.0+):**
- GCP project with Audit Logs enabled
- Falco with gcpaudit plugin
- Pub/Sub subscription for Audit Logs

**Common:**
- Terraform state file (S3, GCS, or local backend)
- Kubernetes cluster (for Falco deployment)
- Grafana + Prometheus (optional, for visualization)

### Installation

#### AWS Setup
```bash
# Clone the repository
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco

# Deploy Falco with cloudtrail plugin
kubectl apply -f deployments/falco/

# Configure TFDrift for AWS
vim config.yaml  # Configure AWS provider and S3 state

# Run the detector
./tfdrift --config config.yaml
```

[Full AWS Setup Guide →](falco-setup.md)

#### GCP Setup (v0.5.0+)
```bash
# Quick start (recommended)
./scripts/gcp-quick-start.sh

# Or manual setup
# See full GCP setup guide
```

[Full GCP Setup Guide →](gcp-setup.md)

#### Multi-Cloud Setup
```yaml
# config.yaml - Monitor both AWS and GCP
providers:
  aws:
    enabled: true
    regions: ["us-east-1", "us-west-2"]
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "aws/terraform.tfstate"

  gcp:
    enabled: true
    projects: ["my-gcp-project-123"]
    state:
      backend: "gcs"
      gcs_bucket: "my-terraform-state"
      gcs_prefix: "gcp/terraform.tfstate"
```

[Full Quickstart Guide →](quickstart.md)

---

## Architecture

```
┌─────────────┐
│ AWS Console │ User makes manual change
│   / CLI     │ (e.g., modify EC2 instance type)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ CloudTrail  │ Event: ModifyInstanceAttribute
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  TFDrift    │ 1. Fetch event
│  Detector   │ 2. Load Terraform state
│             │ 3. Compare attributes
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Falco    │ Drift detected!
│    Rules    │ Severity: WARNING
└──────┬──────┘
       │
       ├───────────────┐
       ▼               ▼
┌─────────────┐ ┌─────────────┐
│   Grafana   │ │   Alerting  │
│  Dashboard  │ │ (Slack/PD)  │
└─────────────┘ └─────────────┘
```

[Learn More About Architecture →](architecture.md)

---

## Use Cases

### 1. Detect Unplanned Changes

**Problem:** Someone modifies infrastructure via cloud console or CLI, bypassing Terraform.

**Solution:** TFDrift-Falco alerts you immediately with:
- **What changed** - Resource type and modified attributes
- **Who made the change** - IAM user/role (AWS) or principal email (GCP)
- **When it happened** - Precise timestamp with timezone
- **Where** - Account/project, region/zone

**Examples:**
- AWS: EC2 instance type changed via AWS Console
- GCP: Compute Engine metadata modified via gcloud CLI

### 2. Security Compliance

**Problem:** Security configurations modified without approval across multiple clouds.

**Solution:** Critical severity alerts for security-related drift:

**AWS:**
- IAM role trust policy changes
- S3 bucket made public
- Security group rules opened to 0.0.0.0/0
- KMS key deletion scheduled

**GCP:**
- Firewall rules allowing public access
- GCS bucket IAM policy changes
- Service account key creation
- KMS crypto key rotation disabled

### 3. Multi-Cloud Governance

**Problem:** Managing drift across multiple cloud providers, accounts, and projects.

**Solution:** Unified monitoring with multi-cloud support:
- **AWS:** Monitor 10+ accounts across multiple regions
- **GCP:** Monitor multiple projects and organizations
- **Hybrid:** Single dashboard for AWS + GCP resources
- **Filtering:** Account/project, region/zone, service-level filtering

---

## Documentation Sections

### Getting Started
- [Overview](overview.md)
- [How It Works](how-it-works.md)
- [Architecture](architecture.md)
- [Quickstart](quickstart.md)
- [Deployment Guide](deployment.md)
- [Falco Setup (AWS)](falco-setup.md)
- [GCP Setup](gcp-setup.md) - **New in v0.5.0**

### AWS Service Coverage (203+ events, 19 services)
- [AWS Services Overview](services/index.md)
- **Compute:** [EC2](services/ec2.md) | [Lambda](services/lambda.md)
- **Networking:** [VPC](services/vpc.md) | [ELB/ALB](services/elb.md)
- **Storage:** [S3](services/s3.md)
- **Databases:** [RDS](services/rds.md) | [DynamoDB](services/dynamodb.md)
- **Security:** [IAM](services/iam.md) | [KMS](services/kms.md)
- **Containers:** [ECS](services/ecs.md) | [EKS](services/eks.md) | [ECR](services/ecr.md)
- **Messaging:** [SNS](services/sns.md) | [SQS](services/sqs.md)
- [All AWS Services →](services/index.md)

### GCP Service Coverage (100+ events, 12+ services) - **New in v0.5.0**
- [GCP Services Overview](services/gcp/index.md)
- **Compute:** [Compute Engine](services/gcp/compute-engine.md) | [Disks](services/gcp/disks.md)
- **Networking:** [VPC & Firewall](services/gcp/vpc.md) | [Routes](services/gcp/routes.md)
- **Storage:** [Cloud Storage](services/gcp/cloud-storage.md)
- **Databases:** [Cloud SQL](services/gcp/cloud-sql.md)
- **Security:** [IAM](services/gcp/iam.md) | [KMS](services/gcp/kms.md)
- **Containers:** [GKE](services/gcp/gke.md) | [Cloud Run](services/gcp/cloud-run.md)
- [All GCP Services →](services/gcp/index.md)

### Release Notes
- [v0.5.0 - Multi-Cloud Support](release-notes/v0.5.0.md) - **Latest (2025-12-17)**
- [v0.2.0-beta](release-notes/v0.2.0-beta.md)
- [v0.3.0 (planned)](release-notes/v0.3.0.md)
- [Architecture Changes](release-notes/architecture-changes.md)

---

## Community

- **GitHub**: [higakikeita/tfdrift-falco](https://github.com/higakikeita/tfdrift-falco)
- **Issues**: [Report bugs or request features](https://github.com/higakikeita/tfdrift-falco/issues)
- **Discussions**: [Ask questions](https://github.com/higakikeita/tfdrift-falco/discussions)
- **Contributing**: [CONTRIBUTING.md](https://github.com/higakikeita/tfdrift-falco/blob/main/CONTRIBUTING.md)

---

## License

TFDrift-Falco is open source under the [MIT License](https://github.com/higakikeita/tfdrift-falco/blob/main/LICENSE).

---

## Next Steps

1. [Understand how TFDrift-Falco works →](how-it-works.md)
2. [Check service coverage for your infrastructure →](services/ec2.md)
3. [Deploy TFDrift-Falco →](quickstart.md)
