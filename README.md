# TFDrift-Falco

**Real-time Terraform Drift Detection powered by Falco**

[![Version](https://img.shields.io/badge/version-0.9.0-blue)](https://github.com/higakikeita/tfdrift-falco/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Falco](https://img.shields.io/badge/Falco-Compatible-blue)](https://falco.org/)
[![Docker](https://img.shields.io/badge/Docker-GHCR-2496ED?logo=docker)](https://ghcr.io/higakikeita/tfdrift-falco)

> **v0.9.0** (2026-03-29) — Azure Full Support, azurerm backend, WebSocket enhancements
> [Release Notes](docs/release-notes/v0.9.0.md) | [CHANGELOG](CHANGELOG.md) | [Roadmap](PROJECT_ROADMAP.md)

**[English]** | [日本語](README.ja.md)

---

## What is TFDrift-Falco?

TFDrift-Falco detects infrastructure drift **in real-time** across AWS, GCP, and Azure by combining Falco's runtime security monitoring with Terraform state comparison. Unlike periodic scan tools, it catches changes **the moment they happen** and tells you not just *what* changed, but *who* did it and *when*.

```
Someone modifies a security group via AWS Console
  → Falco captures CloudTrail event in real-time
    → TFDrift-Falco compares with Terraform state
      → Instant alert with user identity and change details
```

---

## Quick Start

```bash
# Clone and configure
git clone https://github.com/higakikeita/tfdrift-falco.git && cd tfdrift-falco
cp config.yaml.example config.yaml  # Edit with your settings

# Launch
docker compose up -d

# Or try demo mode (no cloud credentials needed)
go run ./cmd/tfdrift --demo
```

See [Getting Started Guide](docs/GETTING_STARTED.md) for detailed setup.

---

## Multi-Cloud Support

TFDrift-Falco supports all three major cloud providers through a unified Provider interface.

### AWS

```yaml
providers:
  aws:
    enabled: true
    regions: [us-east-1]
    state:
      backend: s3
      s3_bucket: "your-terraform-state-bucket"
      s3_key: "terraform.tfstate"
```

500+ CloudTrail events across 40+ services. [AWS details](docs/services/index.md)

### GCP

```yaml
providers:
  gcp:
    enabled: true
    project_id: "your-gcp-project"
    state:
      backend: gcs
      gcs_bucket: "your-terraform-state-bucket"
```

170+ Audit Log events across 27+ services. [GCP setup guide](docs/gcp-setup.md)

### Azure

```yaml
providers:
  azure:
    enabled: true
    subscription_id: "your-subscription-id"
    regions: [eastus, westus2]
    state:
      backend: azurerm
      azure_storage_account: "yourstorageaccount"
      azure_container_name: "tfstate"
      azure_blob_name: "terraform.tfstate"
```

119 operations across 20+ services, with full ResourceDiscoverer and StateComparator. [Azure details](docs/release-notes/v0.9.0.md)

### Provider Capabilities

| Capability | AWS | GCP | Azure |
|---|---|---|---|
| Real-time Event Detection | CloudTrail (500+) | Audit Logs (170+) | Activity Logs (119) |
| Resource Discovery | Yes | Yes | Yes |
| State Comparison | Yes | Yes | Yes |
| Terraform Backend | S3 | GCS | Azure Blob |
| Falco Plugin | aws_cloudtrail | gcpaudit | azureaudit |

---

## Key Features

**Real-time Detection** — Event-driven via Falco gRPC, not periodic scans.

**Three-way Drift Analysis** — Detects unmanaged resources (exists in cloud but not in Terraform), missing resources (in Terraform but deleted from cloud), and modified resources (attribute differences).

**Security Context** — Correlates IAM user identity, API keys, service accounts with every change.

**REST API + WebSocket + SSE** — Full API server with real-time streaming, provider-based filtering, and structured JSON events.

**Webhook Notifications** — Slack, Microsoft Teams, PagerDuty, or custom HTTP endpoints with automatic retries.

**Production Ready** — JWT/API Key authentication, rate limiting, OpenAPI 3.0 spec, Kubernetes Helm Chart with HPA and NetworkPolicy.

---

## Architecture

```
 AWS CloudTrail ─┐
 GCP Audit Logs ─┤──→ Falco (gRPC) ──→ Provider Layer ──→ Drift Engine
 Azure Activity ─┘    (plugins)        (parse/map/discover)  (compare)
                                                                  │
                              ┌────────────────┬──────────────────┤
                              ▼                ▼                  ▼
                         GraphDB          Webhook            API Server
                       (in-memory)     (Slack/Teams)     (REST + WS + SSE)
                                                               │
                                                          React UI
                                                     (Graph/Table/Split)
```

See [Architecture Documentation](docs/architecture.md) for details.

---

## API

```bash
# REST endpoints
GET  /api/v1/drifts        # Drift alerts (with filtering)
GET  /api/v1/events        # Falco events
GET  /api/v1/graph         # Causal graph (Cytoscape format)
GET  /api/v1/state         # Terraform state overview
GET  /api/v1/stats         # Statistics
GET  /api/v1/providers     # Provider capabilities
GET  /health               # Health check

# Real-time
GET  /api/v1/stream        # SSE event stream
WS   /ws                   # WebSocket (supports provider filtering)
```

Full specification: [OpenAPI 3.0](docs/api/openapi.yaml)

---

## Deployment

### Docker Compose

```bash
docker compose up -d
# Frontend: http://localhost:3000
# Backend:  http://localhost:8080/api/v1
# WebSocket: ws://localhost:8080/ws
```

### Kubernetes (Helm)

```bash
helm install tfdrift ./charts/tfdrift-falco
```

### Build from Source

```bash
make build    # Binary: ./bin/tfdrift
make test     # Run tests
make lint     # Run linter
```

---

## Documentation

| Document | Description |
|---|---|
| [Getting Started](docs/GETTING_STARTED.md) | Step-by-step setup guide |
| [Architecture](docs/architecture.md) | System design and data flow |
| [GCP Setup](docs/gcp-setup.md) | GCP-specific configuration |
| [API Reference](docs/api/openapi.yaml) | OpenAPI 3.0 specification |
| [Operations Runbook](docs/operations/runbook.md) | Incident playbooks for SRE |
| [Deployment](docs/deployment.md) | Production deployment options |
| [Contributing](CONTRIBUTING.md) | Development workflow |
| [Versioning](VERSIONING.md) | Release policy and checklist (24 items) |
| [Roadmap](PROJECT_ROADMAP.md) | v0.10.0 → v1.0.0 plan |
| [Changelog](CHANGELOG.md) | Version history |

---

## Why Falco?

> Terraform tells you **what should exist**. Falco tells you **what actually happened**.

Traditional drift detection runs periodic scans — by the time you find the change, you've lost *who* did it and *why*. Falco watches cloud audit logs in real-time through its plugin framework, capturing the actor, the action, and the intent at the moment of change.

TFDrift-Falco bridges these two worlds: the blueprint (Terraform) and the witness (Falco).

Read the full story: [Why Falco?](docs/WHY_FALCO_STORY.md)

---

## Contributing

Contributions are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, coding standards, and PR guidelines.

## License

[MIT](LICENSE)
