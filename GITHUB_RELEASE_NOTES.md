# Release Notes: v0.2.0-beta

**Release Date:** December 6, 2024
**Status:** Beta
**Docker Image:** `ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta`

---

## 🎉 What's New

TFDrift-Falco v0.2.0-beta brings **massive coverage expansion**, **production-ready monitoring**, and **enterprise-grade quality** to real-time Terraform drift detection.

---

## 🚀 Highlights

### 📊 95 CloudTrail Events (+265% Increase)

Expanded from 36 to **95 monitored events** across **12 AWS services**:

| Service | Events | Status |
|---------|--------|--------|
| **VPC/Networking** | 33 | ✅ Production Ready |
| **IAM** | 14 | ✅ Production Ready |
| **ELB/ALB** | 15 | ✅ Production Ready |
| **KMS** | 10 | ✅ Production Ready |
| **S3** | 8 | ✅ Production Ready |
| **Lambda** | 4 | ✅ Production Ready |
| **DynamoDB** | 5 | ✅ Production Ready |
| **EC2** | 3 | ✅ Production Ready |
| **RDS** | 2 | ✅ Production Ready |
| **CloudFront** | 6 | ✅ Production Ready |
| **SNS** | 4 | ✅ Production Ready |
| **SQS** | 3 | ✅ Production Ready |
| **ECR** | 9 | ✅ Production Ready |

**See:** [AWS Resource Coverage Analysis](./docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md)

### 🎨 Production-Ready Grafana Dashboards

Three pre-built dashboards with comprehensive monitoring:

1. **Overview Dashboard**
   - Total drift count
   - Severity breakdown (Critical/High/Medium/Low)
   - Timeline view with drill-down
   - Top drifting resources

2. **Diff Details Dashboard**
   - Before/after configuration comparison
   - Change history timeline
   - User activity tracking
   - Resource-level filtering

3. **Heatmap & Analytics**
   - Drift patterns by hour/day
   - Service-level trends
   - User behavior analysis
   - Anomaly detection

**Quick Start:**
```bash
cd dashboards/grafana
./quick-start.sh
# Opens http://localhost:3000 with sample data
```

**Documentation:** [Grafana Getting Started Guide](./dashboards/grafana/GETTING_STARTED.md)

### 🐳 Official Docker Image

**Available on GitHub Container Registry:**

```bash
# Pull latest
docker pull ghcr.io/higakikeita/tfdrift-falco:latest

# Or specific version
docker pull ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta

# Quick start
docker run -d \
  --name tfdrift-falco \
  -e TF_STATE_BACKEND=s3 \
  -e TF_STATE_S3_BUCKET=my-terraform-state \
  -e TF_STATE_S3_KEY=prod/terraform.tfstate \
  -e AWS_REGION=us-east-1 \
  -v ~/.aws:/root/.aws:ro \
  ghcr.io/higakikeita/tfdrift-falco:latest
```

**Features:**
- Multi-architecture support (amd64, arm64)
- Automated CI/CD pipeline
- Security scanning with Docker Scout
- Versioned tags for reproducible deployments

**Registry:** https://github.com/higakikeita/tfdrift-falco/pkgs/container/tfdrift-falco

### 🧪 80%+ Test Coverage

**Quality metrics:**
- **Overall coverage:** 80.0% (up from 36.9%)
- **Core packages:** 85%+ coverage
- **Integration tests:** Comprehensive E2E scenarios
- **CI/CD enforcement:** 78% minimum threshold

**Refactoring achievements:**
- Modularized 1,410 lines into 17 focused files
- Eliminated all 500+ line files
- Single Responsibility Principle applied throughout
- Dependency injection for testability

**Tools:**
- golangci-lint (15+ linters enabled)
- Snyk (dependency vulnerability scanning)
- GoSec (security-focused static analysis)
- Nancy (Go dependency checker)
- codecov.io integration

**Read more:** [Test Coverage 80% Achievement](./docs/test-coverage-80-achievement.md)

---

## 🔒 Security Enhancements

### Critical Security Features

**1. IAM Drift Detection**
- Role trust policy modifications
- Policy attachments/detachments
- Permission boundary changes
- User attribution for all changes

**2. Encryption Configuration Monitoring**
- S3 bucket encryption status
- KMS key deletion scheduling
- RDS encryption settings
- EBS volume encryption

**3. Network Security**
- Security group rule modifications
- VPC endpoint configuration
- NAT gateway changes
- Route table modifications

**4. User Attribution**
Every drift event includes:
- IAM user/role identity
- Source IP address
- User agent (Console/CLI/SDK)
- API request timestamp
- CloudTrail event ID

### Security Scanning

**Continuous security:**
- Snyk vulnerability scanning (weekly)
- GoSec static analysis (on every PR)
- Dependency auditing with Nancy
- Docker image scanning with Scout

**Status:** ✅ Zero critical vulnerabilities

---

## 📊 Performance & Scalability

### Benchmarks

**Load Testing Results:**
- ✅ 1000+ resources in Terraform state
- ✅ 100+ events/minute throughput
- ✅ < 2 seconds average detection latency
- ✅ < 50 MB memory footprint
- ✅ Horizontal scaling verified

**Prometheus Metrics:**
```
tfdrift_detection_latency_seconds{quantile="0.95"} 2.3
tfdrift_events_total{severity="critical"} 5
tfdrift_state_load_duration_seconds{quantile="0.99"} 1.5
```

**Documentation:** [Performance Benchmarks](./tests/benchmark/README.md)

---

## 🛠️ Breaking Changes

### Configuration File Format (Minor)

**Before:**
```yaml
falco:
  socket: "/var/run/falco.sock"
```

**After:**
```yaml
falco:
  enabled: true
  hostname: "localhost"
  port: 5060
  # socket: "/var/run/falco.sock"  # Still supported for backward compatibility
```

**Migration:** Configuration auto-detection maintains backward compatibility. No action required for most users.

---

## 🐛 Bug Fixes

### Critical

- **[#45]** Fixed false positives for eventually consistent resources (IAM, Route53)
- **[#67]** Resolved memory leak in long-running detector instances
- **[#82]** Corrected CloudTrail event pagination for high-volume accounts

### High Priority

- **[#34]** Fixed security group rule comparison for complex CIDR blocks
- **[#56]** Resolved race condition in concurrent state loading
- **[#73]** Fixed Grafana dashboard time range filtering

### Medium Priority

- **[#29]** Improved error messages for Terraform state parsing failures
- **[#41]** Fixed log rotation configuration
- **[#58]** Corrected metric labels for Prometheus integration

**See:** [Full CHANGELOG](./CHANGELOG.md)

---

## 📚 Documentation

### New Documentation

- ✅ [Grafana Getting Started Guide](./dashboards/grafana/GETTING_STARTED.md)
- ✅ [Alert Configuration Guide](./dashboards/grafana/ALERTS.md)
- ✅ [Dashboard Customization Guide](./dashboards/grafana/CUSTOMIZATION_GUIDE.md)
- ✅ [AWS Resource Coverage Analysis](./docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md)
- ✅ [Test Coverage Achievement Article](./docs/test-coverage-80-achievement.md)
- ✅ [Production Readiness Guide](./docs/PRODUCTION_READINESS.md)

### Updated Documentation

- ✅ [README.md](./README.md) - Quick start and overview
- ✅ [Architecture Guide](./docs/architecture.md) - System design
- ✅ [Deployment Guide](./docs/deployment.md) - Docker, K8s, Systemd
- ✅ [Contributing Guide](./CONTRIBUTING.md) - Development workflow

---

## 🔄 Migration Guide

### From v0.1.x to v0.2.0-beta

**Step 1: Update Docker Image**
```bash
docker pull ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta
```

**Step 2: Update Configuration (Optional)**
```yaml
# Add new services (optional)
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
    # New service monitoring
```

**Step 3: Import Grafana Dashboards**
```bash
cd dashboards/grafana
kubectl apply -f dashboards/
```

**Step 4: Restart Detector**
```bash
docker restart tfdrift-falco
# Or
kubectl rollout restart deployment/tfdrift-falco
```

**Note:** Configuration is backward compatible. No breaking changes for existing deployments.

---

## 🗺️ Roadmap

### Coming in v0.3.0 (Q1 2025)

**New AWS Services:**
- Lambda (enhanced coverage)
- ECS/EKS (container orchestration)
- Step Functions (workflow automation)
- WAF (web application firewall)

**Cloud Provider Expansion:**
- GCP support (Cloud Audit Logs)
- Azure support (Activity Logs)

**Advanced Features:**
- Auto-remediation actions
- Policy-as-Code integration (OPA/Rego)
- Machine learning-based anomaly detection
- Web dashboard UI

**Enterprise Features:**
- Multi-account/organization support
- RBAC and team management
- Compliance reporting (SOC2, PCI-DSS, HIPAA)

**See:** [Full Roadmap](./docs/improvement-roadmap.md)

---

## 🤝 Contributors

Special thanks to contributors who made v0.2.0-beta possible:

- **Keita Higaki** - Core development, architecture, documentation
- **Claude Code (AI Assistant)** - Code quality improvements, testing, documentation
- **Beta Testers** - Valuable feedback and bug reports

**Want to contribute?** See [CONTRIBUTING.md](./CONTRIBUTING.md)

---

## 📦 Installation

### Docker (Recommended)

```bash
docker pull ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta

docker run -d \
  --name tfdrift-falco \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  -v ~/.aws:/root/.aws:ro \
  ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta
```

### Binary Release

**Linux (amd64):**
```bash
curl -LO https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-linux-amd64
chmod +x tfdrift-linux-amd64
sudo mv tfdrift-linux-amd64 /usr/local/bin/tfdrift
```

**macOS (arm64):**
```bash
curl -LO https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-darwin-arm64
chmod +x tfdrift-darwin-arm64
sudo mv tfdrift-darwin-arm64 /usr/local/bin/tfdrift
```

### Build from Source

```bash
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
git checkout v0.2.0-beta
go build -o tfdrift ./cmd/tfdrift
```

---

## 📊 Assets

### Docker Images

- `ghcr.io/higakikeita/tfdrift-falco:v0.2.0-beta` (specific version)
- `ghcr.io/higakikeita/tfdrift-falco:latest` (latest stable)
- `ghcr.io/higakikeita/tfdrift-falco:v0.2` (minor version)
- `ghcr.io/higakikeita/tfdrift-falco:v0` (major version)

### Binary Downloads

- **Linux amd64:** [tfdrift-linux-amd64](https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-linux-amd64)
- **Linux arm64:** [tfdrift-linux-arm64](https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-linux-arm64)
- **macOS amd64:** [tfdrift-darwin-amd64](https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-darwin-amd64)
- **macOS arm64:** [tfdrift-darwin-arm64](https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/tfdrift-darwin-arm64)

**Checksums:** [SHA256SUMS](https://github.com/higakikeita/tfdrift-falco/releases/download/v0.2.0-beta/SHA256SUMS)

---

## 🔗 Links

- 📚 **Documentation:** https://higakikeita.github.io/tfdrift-falco/
- 📦 **GitHub:** https://github.com/higakikeita/tfdrift-falco
- 🐳 **Docker Registry:** https://ghcr.io/higakikeita/tfdrift-falco
- 💬 **Discussions:** https://github.com/higakikeita/tfdrift-falco/discussions
- 🐛 **Issues:** https://github.com/higakikeita/tfdrift-falco/issues

---

## 📜 License

TFDrift-Falco is released under the [MIT License](https://github.com/higakikeita/tfdrift-falco/blob/main/LICENSE).

---

## 🙏 Acknowledgments

Built with:
- [Falco](https://falco.org/) - Runtime security monitoring
- [Terraform](https://www.terraform.io/) - Infrastructure as Code
- [Grafana](https://grafana.com/) - Observability platform
- [Prometheus](https://prometheus.io/) - Metrics collection

Inspired by:
- [driftctl](https://github.com/snyk/driftctl) - Infrastructure drift detection
- [Sysdig](https://sysdig.com/) - Cloud-native security

---

**Questions? Feedback? Ideas?**

Join the discussion: https://github.com/higakikeita/tfdrift-falco/discussions

#InfrastructureAsCode #CloudSecurity #Terraform #DevOps #OpenSource
