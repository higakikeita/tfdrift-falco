# TFDrift-Falco Documentation

Welcome to the official documentation for **TFDrift-Falco**, a real-time Terraform drift detection system powered by AWS CloudTrail and Falco.

---

## What is TFDrift-Falco?

TFDrift-Falco detects when your AWS infrastructure changes outside of Terraform by:

1. **Monitoring AWS CloudTrail events** in real-time
2. **Comparing changes against Terraform state**
3. **Alerting via Falco** when drift is detected
4. **Visualizing drift in Grafana** dashboards

---

## Key Features

### 🚀 Comprehensive AWS Coverage

Supports **150+ CloudTrail events** across **12 AWS services**:
- EC2, VPC, Security Groups
- S3, RDS, Aurora
- IAM, KMS
- API Gateway, Route53, CloudFront
- SNS, SQS, ECR

[View Service Coverage →](services/ec2.md)

### ⚡ Real-time Detection

- **Sub-minute latency** from AWS change to alert
- Asynchronous CloudTrail processing
- Parallel service detection

### 🔐 Security-Focused

- IAM policy drift detection
- Encryption configuration monitoring
- Security group rule changes
- KMS key policy modifications

### 📊 Production-Ready Monitoring

- **Grafana dashboards** for 12 services
- **Falco rules** with severity levels
- **User attribution** for every change
- **Alert integration** ready

---

## Quick Start

### Prerequisites

- AWS account with CloudTrail enabled
- Terraform state file (S3 backend recommended)
- Kubernetes cluster (for Falco deployment)
- Grafana + Prometheus (for visualization)

### Installation

```bash
# Clone the repository
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco

# Deploy Falco with TFDrift rules
kubectl apply -f deployments/falco/

# Configure TFDrift detector
cp config-example.yaml config.yaml
vim config.yaml  # Edit with your AWS account ID, S3 state location

# Run the detector
./tfdrift --config config.yaml
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

**Problem:** Someone modifies infrastructure via AWS Console, bypassing Terraform.

**Solution:** TFDrift-Falco alerts you immediately with:
- What changed (resource, attribute)
- Who made the change (IAM user/role)
- When it happened (timestamp)

### 2. Security Compliance

**Problem:** IAM policies, S3 encryption, or security groups modified without approval.

**Solution:** Critical severity alerts for security-related drift:
- IAM role trust policy changes
- S3 bucket made public
- KMS key deletion scheduled

### 3. Multi-Account Governance

**Problem:** Managing drift across 10+ AWS accounts with separate Terraform workspaces.

**Solution:** Centralized monitoring with account/region filtering in Grafana.

---

## Documentation Sections

### Getting Started
- [Overview](overview.md)
- [Architecture](architecture.md)
- [Quickstart](quickstart.md)
- [Deployment Guide](deployment.md)

### AWS Service Coverage
- [EC2](services/ec2.md) | [IAM](services/iam.md) | [S3](services/s3.md) | [VPC](services/vpc.md)
- [RDS](services/rds.md) | [API Gateway](services/api-gateway.md) | [Route53](services/route53.md)
- [CloudFront](services/cloudfront.md) | [SNS](services/sns.md) | [SQS](services/sqs.md)
- [ECR](services/ecr.md) | [KMS](services/kms.md)

### Release Notes
- [v0.2.0-beta](release-notes/v0.2.0-beta.md) (Current)
- [v0.3.0](release-notes/v0.3.0.md) (Planned)
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
