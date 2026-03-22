# GitHub Actions Workflows

CI/CD workflows for TFDrift-Falco. All workflows run on GitHub Actions.

## Workflow Overview

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| **CI/CD Pipeline** | `ci.yml` | Push/PR/Schedule | Build, test, lint, security scan (backend + frontend) |
| **Snyk Security** | `snyk-security.yml` | Schedule/Manual | Comprehensive Snyk scans (OSS, SAST, IaC, Container, License) |
| **Sysdig Scan** | `sysdig-scan.yml` | PR (Dockerfile changes)/Manual | Container image vulnerability scanning |
| **E2E Tests** | `e2e.yml` | Manual/Schedule/Label | End-to-end tests with real AWS & Falco |
| **Publish GHCR** | `publish-ghcr.yml` | Tag/Release/Manual | Build and publish Docker images to GHCR |
| **Deploy Docs** | `docs.yml` | Push (docs/**)/Manual | Build MkDocs site and deploy to GitHub Pages |
| **Deploy Storybook** | `ui-storybook.yml` | Push to main/Manual | Build and deploy Storybook to GitHub Pages |
| **Website Security** | `website-security.yml` | Push (website/**)/Schedule | npm audit + Snyk scan for website |

## ci.yml — Main Pipeline

The primary CI workflow. Runs on every push and PR.

**Jobs:**

| Job | Purpose | Blocking? |
|-----|---------|-----------|
| `changes` | Detect changed files (backend/frontend) | — |
| `backend` | Go build, lint (golangci-lint, gofmt, go vet, staticcheck), test with coverage | Yes (tests, build) |
| `frontend` | Node build, lint (ESLint, tsc), test (Vitest) with coverage | Yes (tests, build) |
| `docker` | Multi-platform Docker builds (amd64, arm64) | No |
| `integration` | Integration tests with mocked dependencies | No |
| `snyk-open-source` | Snyk OSS scan (Go + npm dependencies) | No |
| `snyk-code` | Snyk Code SAST (source code security) | No |
| `snyk-iac` | Snyk IaC (Terraform, Helm, Dockerfile) | No |
| `snyk-container` | Snyk Container (Docker image scan, main only) | No |
| `security-gosec` | GoSec security scanner | No |
| `security-nancy` | Nancy dependency checker | No |
| `benchmark` | Performance benchmarks (label-triggered) | No |

## snyk-security.yml — Dedicated Snyk Dashboard

Runs weekly and on-demand. Provides detailed vulnerability reports.

**Scan types** (selectable via manual dispatch):
- `open-source` — Go & npm dependency SCA with PR comments
- `code` — SAST scanning with SARIF upload
- `iac` — Terraform, Helm charts, Dockerfiles
- `container` — Docker image scanning
- `license` — Dependency license compliance
- `all` — Run everything

## Required Secrets

| Secret | Used By | Status |
|--------|---------|--------|
| `SNYK_TOKEN` | ci.yml, snyk-security.yml, website-security.yml | **Configured** |
| `SNYK_ORG_ID` | ci.yml, snyk-security.yml | **Configured** |
| `SYSDIG_SECURE_API_TOKEN` | sysdig-scan.yml | **Configured** |
| `E2E_AWS_ACCESS_KEY_ID` | e2e.yml | Optional (E2E only) |
| `E2E_AWS_SECRET_ACCESS_KEY` | e2e.yml | Optional (E2E only) |
| `E2E_CLOUDTRAIL_BUCKET` | e2e.yml | Optional (E2E only) |
| `DOCKER_USERNAME` | ci.yml (Docker Hub push) | Optional (releases only) |
| `DOCKER_PASSWORD` | ci.yml (Docker Hub push) | Optional (releases only) |
| `CODECOV_TOKEN` | ci.yml | Optional |

## Running Locally

```bash
# Backend tests
make test                    # Unit tests
make test-race               # Race detection
make test-coverage           # With coverage report
make test-integration        # Integration tests
make test-benchmark          # Performance benchmarks

# Frontend tests
cd ui && npm test            # Unit tests
cd ui && npm run test:coverage  # With coverage

# Security scans
npm install -g snyk && snyk auth
snyk test --file=go.mod      # Go dependency scan
snyk code test               # SAST
snyk iac test terraform/     # IaC scan

# E2E tests (requires AWS credentials)
cd tests/e2e && make all
```

## Troubleshooting

**Backend tests failing**: Run `make test` locally, check for race conditions with `make test-race`.

**Frontend tests failing**: Run `cd ui && npm test`. Common issue: missing mock exports for lucide-react icons or hooks (useSSE, useTheme).

**Snyk scan issues**: Verify `SNYK_TOKEN` and `SNYK_ORG_ID` secrets are set. All Snyk jobs use `continue-on-error: true` so failures are non-blocking.

**Sysdig scan issues**: See [SYSDIG_SETUP.md](SYSDIG_SETUP.md) for setup. Requires `SYSDIG_SECURE_API_TOKEN`.

**E2E tests**: Require real AWS credentials. Estimated cost ~$0.10-0.50/run. Run on-demand with `run-e2e` label.
