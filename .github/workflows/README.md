# GitHub Actions Workflows

CI/CD workflows for TFDrift-Falco. All workflows run on GitHub Actions.

## Workflow Overview

| Workflow | File | Trigger | Purpose |
|----------|------|---------|---------|
| **CI/CD Pipeline** | `ci.yml` | Push/PR/Release | Build & test (Go, React, Docker, Integration) |
| **Security Scanning** | `security.yml` | Push/PR/Schedule/Manual | Unified security (Snyk 5 types + GoSec + Nancy) |
| **Benchmarks** | `benchmark.yml` | Label/Manual/Main | Performance benchmarks with baseline tracking |
| **Sysdig Scan** | `sysdig-scan.yml` | PR (Dockerfile changes)/Manual | Container image vulnerability scanning |
| **E2E Tests** | `e2e.yml` | Manual/Schedule/Label | End-to-end tests with real AWS & Falco |
| **Publish GHCR** | `publish-ghcr.yml` | Tag/Release/Manual | Build and publish Docker images to GHCR |
| **Deploy Docs** | `docs.yml` | Push (docs/**)/Manual | Build MkDocs site and deploy to GitHub Pages |
| **Deploy Storybook** | `ui-storybook.yml` | Push to main/Manual | Build and deploy Storybook to GitHub Pages |
| **Website Security** | `website-security.yml` | Push (website/**)/Schedule | npm audit + Snyk scan for website |

## ci.yml — Build & Test Only

The primary CI workflow. Runs on every push and PR. **Focuses on build and test only** — security scans moved to `security.yml`.

**Jobs:**

| Job | Purpose | Blocking? |
|-----|---------|-----------|
| `changes` | Detect changed files (backend/frontend) | — |
| `backend` | Go build, lint (golangci-lint, gofmt, go vet, staticcheck), test with coverage | Yes (tests, build) |
| `frontend` | Node build, lint (ESLint, tsc), test (Vitest) with coverage | Yes (tests, build) |
| `docker` | Multi-platform Docker builds (amd64, arm64) | No |
| `integration` | Integration tests with mocked dependencies | No |

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`
- Release (published)
- Manual dispatch (`workflow_dispatch`)

## security.yml — Unified Security Scanning

All security scans in one place. Runs independently so failures don't block CI/CD.

**Jobs & Scan Types** (all use `continue-on-error: true`):

| Job | Type | Triggers | PR Comment? |
|-----|------|----------|------------|
| `snyk-open-source` | SCA (Go + npm deps) | Push/PR/Schedule | Yes |
| `snyk-code` | SAST (source code) | Push/PR/Schedule | Yes |
| `snyk-iac` | IaC (Terraform, Helm, Docker) | Push/PR/Schedule | Yes |
| `snyk-container` | Container (Docker image) | Push/PR/Schedule | Yes |
| `snyk-license` | License compliance | Push/PR/Schedule | No |
| `gosec` | Go security scanner | Push/PR/Schedule | Yes |
| `nancy` | Go dependency checker | Push/PR/Schedule | No |

**Triggers:**
- Push to `main` or `develop`
- All pull requests to `main` or `develop`
- Weekly schedule: Monday 6 AM UTC
- Manual dispatch with `scan_type` selector (all/snyk/gosec/nancy)

**Key Features:**
- Unified severity thresholds (medium for OSS/IaC, high for container)
- All jobs upload SARIF to GitHub Security tab
- Path-based filtering on push/PR (only run relevant scans)
- Schedule runs everything
- PR comments summarize findings
- No impact on build status

## benchmark.yml — Performance Tracking

Isolates performance benchmarks from the main CI pipeline. Runs on label, manual dispatch, or main branch updates.

**Jobs:**

| Job | Purpose | Trigger |
|-----|---------|---------|
| `benchmark` | Run benchmarks, compare vs baseline, track regressions | Label (`benchmark`), Manual dispatch, Push to main |

**Baseline Tracking:**
- On push to main: Auto-commits new baseline
- On PR with `benchmark` label: Compares vs base branch
- Detects >20% performance regressions and warns (non-blocking)

## Required Secrets

| Secret | Used By | Status |
|--------|---------|--------|
| `SNYK_TOKEN` | security.yml, website-security.yml | **Configured** |
| `SNYK_ORG_ID` | security.yml | **Configured** |
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

## Workflow Architecture Notes

### Separation of Concerns
- **ci.yml**: Build + test only (fast, blocking)
- **security.yml**: All security scans (independent, non-blocking, informational)
- **benchmark.yml**: Performance tracking (separate triggers, baseline managed)

### Security Scanning Details
- All Snyk + GoSec + Nancy jobs use `continue-on-error: true` — they're informational only
- No security scan failures will block PRs or builds
- All jobs upload SARIF to GitHub Security tab
- PR comments provide quick summary
- Schedule (Monday 6 AM UTC) runs comprehensive scans with no path filtering

## Troubleshooting

**Backend tests failing**: Run `make test` locally, check for race conditions with `make test-race`.

**Frontend tests failing**: Run `cd ui && npm test`. Common issue: missing mock exports for lucide-react icons or hooks (useSSE, useTheme).

**Security scans not showing in GitHub**: Check that the workflow has permission `security-events: write`. All scans upload SARIF under distinct categories (snyk-oss-go, snyk-code, gosec, etc.).

**Snyk scans taking too long**: They run in parallel and are non-blocking. Set manual dispatch `scan_type` to reduce scope (e.g., `snyk` runs only Snyk jobs, `gosec` runs only GoSec).

**Sysdig scan issues**: See [SYSDIG_SETUP.md](SYSDIG_SETUP.md) for setup. Requires `SYSDIG_SECURE_API_TOKEN`.

**E2E tests**: Require real AWS credentials. Estimated cost ~$0.10-0.50/run. Run on-demand with `run-e2e` label.

**Benchmark degradation warnings**: False positives are possible. Review the PR comments for details and re-run if needed.
