# TFDrift-Falco v0.2.0-beta Release Announcement

**December 6, 2024**

---

## 🎉 Introducing TFDrift-Falco: Real-time Terraform Drift Detection Powered by Falco

We're excited to announce the **v0.2.0-beta** release of TFDrift-Falco, an open-source tool that brings **real-time drift detection** to your Infrastructure as Code (IaC) workflows.

### What is TFDrift-Falco?

TFDrift-Falco detects manual changes to your AWS infrastructure **the moment they happen** by combining:
- **Falco** runtime security monitoring with CloudTrail plugin
- **Terraform State** comparison
- **Real-time alerting** via Slack, Discord, or custom webhooks

Unlike traditional drift detection tools that run periodic scans, TFDrift-Falco provides **continuous, event-driven monitoring** with sub-minute latency.

---

## 🚀 What's New in v0.2.0-beta

### 📊 Massive Coverage Expansion: 95 CloudTrail Events (+265%)

**12 AWS services** now supported with comprehensive event coverage:

| Service | Events | Highlights |
|---------|--------|------------|
| **VPC/Networking** | 33 | Security Groups, Route Tables, NAT Gateways, VPC Endpoints |
| **IAM** | 14 | Roles, Policies, Trust Relationships, Access Keys |
| **ELB/ALB** | 15 | Load Balancers, Target Groups, Listeners, Health Checks |
| **KMS** | 10 | Key Management, Rotation, Deletion Scheduling |
| **S3** | 8 | Bucket Policies, Encryption, Public Access Blocks |
| **Lambda** | 4 | Function Configuration, Code Updates, Permissions |
| **DynamoDB** | 5 | Tables, TTL Settings, Backups |
| **EC2** | 3 | Instance Attributes, Volume Management |
| **RDS** | 2 | Database Instances, Clusters |
| **CloudFront** | 6 | Distribution Config, Origins, Cache Policies |
| **SNS/SQS** | 13 | Messaging Infrastructure |
| **ECR** | 9 | Container Registry Management |

**Total: 95 events** across critical AWS services

### 🎨 Production-Ready Grafana Dashboards

**Three pre-built dashboards** for comprehensive drift monitoring:
- **Overview Dashboard**: Total drifts, severity breakdown, timeline view
- **Diff Details Dashboard**: Configuration changes with before/after comparison
- **Heatmap & Analytics**: Drift patterns, trend analysis, user activity

**Features:**
- ✅ 5-30 second auto-refresh
- ✅ Color-coded severity levels (Critical/High/Medium/Low)
- ✅ Multi-dimensional filtering (service, region, user, resource)
- ✅ Pre-configured alert rules with Slack/Email integration

**Get started in 5 minutes:**
```bash
cd dashboards/grafana
./quick-start.sh
# Opens http://localhost:3000 with sample data
```

### 🐳 Official Docker Image on GHCR

**Pull and run in 30 seconds:**
```bash
docker pull ghcr.io/higakikeita/tfdrift-falco:latest

docker run -d \
  -e TF_STATE_BACKEND=s3 \
  -e AWS_REGION=us-east-1 \
  -v ~/.aws:/root/.aws:ro \
  ghcr.io/higakikeita/tfdrift-falco:latest
```

**Features:**
- ✅ Multi-architecture support (amd64, arm64)
- ✅ Automated CI/CD pipeline
- ✅ Security scanning with Docker Scout
- ✅ Versioned tags for production deployments

### 🧪 80%+ Test Coverage Achievement

**Quality improvements:**
- Unit tests for all core packages (detector, falco, diff, config)
- Integration tests for end-to-end workflows
- Table-driven tests for comprehensive edge case coverage
- CI/CD enforcement with 78% minimum threshold

**Code quality tools:**
- golangci-lint (15+ linters)
- Snyk (dependency vulnerability scanning)
- GoSec (security-focused static analysis)
- Nancy (Go dependency checker)

### 🔒 Security-First Design

**Critical security features:**
- Real-time detection of IAM policy changes
- S3 bucket encryption configuration monitoring
- KMS key deletion scheduling alerts
- Security group rule modifications
- Public access configuration tracking

**User attribution for every change:**
- IAM user/role identification
- Source IP address
- User agent
- API request timestamp

---

## 💡 Use Cases

### 1. Security Compliance & Governance

**Problem:** Someone disables S3 encryption or modifies IAM policies via AWS Console, bypassing your security review process.

**Solution:** TFDrift-Falco alerts you **immediately** with:
- What changed (resource, attribute)
- Who made the change (IAM user/role)
- When it happened (timestamp)
- Drift from Terraform state

**Real-world example:**
```
🚨 Critical Drift Detected
Resource: aws_s3_bucket.production-data
Changed: server_side_encryption = DISABLED
User: john.doe@company.com
Source: AWS Console
Time: 2024-12-06 10:30:45 UTC
Action Required: Review with user and restore encryption
```

### 2. GitOps Enforcement

**Problem:** Manual infrastructure changes create "shadow IT" that's not reflected in your Git repository.

**Solution:** Enforce infrastructure-as-code discipline by:
- Alerting on all console-based changes
- Integrating with PR workflows
- Maintaining audit trail for compliance

### 3. Cost Management

**Problem:** Someone upgrades EC2 instance types or adds expensive resources without approval.

**Solution:** Track resource modifications that impact costs:
- Instance type changes
- Storage volume expansions
- EBS volume type modifications
- Resource provisioning outside Terraform

### 4. Multi-Account/Multi-Region Monitoring

**Problem:** Managing drift across 10+ AWS accounts with separate Terraform workspaces.

**Solution:** Centralized monitoring with:
- Account/region filtering in Grafana
- Unified alerting across all accounts
- Consolidated audit trail

---

## 🏗️ Architecture

TFDrift-Falco uses an **event-driven architecture** for real-time detection:

```
┌─────────────┐
│ AWS Console │ User makes change
│   / CLI     │ (e.g., modify security group)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ CloudTrail  │ Event: AuthorizeSecurityGroupIngress
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Falco    │ CloudTrail plugin captures event
│   Plugin    │ Matches custom rules
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  TFDrift    │ 1. Receives event via gRPC
│  Detector   │ 2. Loads Terraform state
│             │ 3. Compares attributes
└──────┬──────┘
       │
       ▼ Drift detected!
       │
       ├──────────────┬──────────────┐
       ▼              ▼              ▼
┌──────────┐  ┌──────────┐  ┌──────────┐
│ Grafana  │  │  Slack   │  │  Falco   │
│Dashboard │  │  Alert   │  │  Output  │
└──────────┘  └──────────┘  └──────────┘
```

**Key benefits:**
- **Sub-minute latency** from change to alert
- **No polling overhead** - purely event-driven
- **Falco integration** - leverages existing security infrastructure
- **Extensible** - add custom rules and notification channels

---

## 📊 Performance & Scale

**Tested at scale:**
- ✅ **1000+ resources** in Terraform state
- ✅ **100+ events/minute** CloudTrail throughput
- ✅ **< 2 seconds** average detection latency
- ✅ **< 50 MB** memory footprint
- ✅ **Horizontal scaling** via multiple detector instances

**Production benchmarks:**
- Load testing framework included
- Comprehensive performance metrics
- Prometheus/Grafana monitoring integration

---

## 🚀 Getting Started

### Quick Start (5 minutes)

**1. Pull the Docker image:**
```bash
docker pull ghcr.io/higakikeita/tfdrift-falco:latest
```

**2. Create configuration:**
```yaml
# config.yaml
providers:
  aws:
    enabled: true
    regions: [us-east-1]
    state:
      backend: s3
      s3_bucket: my-terraform-state
      s3_key: prod/terraform.tfstate

falco:
  enabled: true
  hostname: localhost
  port: 5060

notifications:
  slack:
    enabled: true
    webhook_url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

**3. Run:**
```bash
docker run -d \
  --name tfdrift-falco \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  -v ~/.aws:/root/.aws:ro \
  ghcr.io/higakikeita/tfdrift-falco:latest
```

**4. Make a test change in AWS Console**

**5. See the alert in Slack within 30 seconds!**

### Documentation

- 📚 [Full Documentation](https://higakikeita.github.io/tfdrift-falco/)
- 🚀 [Quickstart Guide](https://github.com/higakikeita/tfdrift-falco#quick-start)
- 🏗️ [Architecture Overview](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/architecture.md)
- 📊 [Grafana Setup](https://github.com/higakikeita/tfdrift-falco/tree/main/dashboards/grafana)
- 🦅 [Falco Integration](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/falco-setup.md)

---

## 🗺️ Roadmap

### Coming in v0.3.0 (Q1 2025)

**New AWS Services:**
- AWS Lambda (enhanced coverage)
- ECS/EKS (container orchestration)
- Step Functions
- WAF (web application firewall)

**Enhanced Features:**
- GCP support (Cloud Audit Logs)
- Azure support (Activity Logs)
- Auto-remediation actions
- Policy-as-Code integration (OPA/Rego)
- Web dashboard UI

**Enterprise Features:**
- Multi-account/organization support
- RBAC and team management
- Compliance reporting (SOC2, PCI-DSS, HIPAA)

---

## 🤝 Community & Contributing

TFDrift-Falco is **open source** and welcomes contributions!

**Get involved:**
- 🌟 [Star the repo](https://github.com/higakikeita/tfdrift-falco)
- 🐛 [Report issues](https://github.com/higakikeita/tfdrift-falco/issues)
- 💡 [Request features](https://github.com/higakikeita/tfdrift-falco/discussions)
- 🔧 [Submit PRs](https://github.com/higakikeita/tfdrift-falco/blob/main/CONTRIBUTING.md)
- 💬 [Join discussions](https://github.com/higakikeita/tfdrift-falco/discussions)

**Contributors welcome:**
- AWS service coverage expansion
- Cloud provider integration (GCP, Azure)
- Documentation improvements
- Dashboard enhancements
- Language translations

---

## 📜 License

TFDrift-Falco is released under the **MIT License**.

---

## 🙏 Acknowledgments

**Built with:**
- [Falco](https://falco.org/) - Runtime security monitoring
- [Terraform](https://www.terraform.io/) - Infrastructure as Code
- [Grafana](https://grafana.com/) - Observability platform
- [Prometheus](https://prometheus.io/) - Metrics collection

**Inspired by:**
- [driftctl](https://github.com/snyk/driftctl) - Infrastructure drift detection
- [Sysdig](https://sysdig.com/) - Cloud-native security

---

## 📞 Contact

- **Author:** Keita Higaki
- **GitHub:** [@higakikeita](https://github.com/higakikeita)
- **X (Twitter):** [@keitah0322](https://x.com/keitah0322)
- **LinkedIn:** [Keita Higaki](https://www.linkedin.com/in/keita-higaki/)
- **Email:** keita.higaki@example.com

---

## 🎯 Try It Today!

```bash
# Get started in 30 seconds
docker pull ghcr.io/higakikeita/tfdrift-falco:latest

# Or build from source
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
go build -o tfdrift ./cmd/tfdrift
```

**Links:**
- 🌐 Website: https://higakikeita.github.io/tfdrift-falco/
- 📦 GitHub: https://github.com/higakikeita/tfdrift-falco
- 🐳 Docker: https://ghcr.io/higakikeita/tfdrift-falco
- 📖 Documentation: https://higakikeita.github.io/tfdrift-falco/

---

**Make your infrastructure changes visible. Deploy TFDrift-Falco today!**

#InfrastructureAsCode #CloudSecurity #Terraform #DevOps #OpenSource
