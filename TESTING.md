# TFDrift-Falco Testing Guide

## Quick Reference

```bash
make test                    # Run unit tests
make test-coverage           # Run with coverage report (generates coverage.html)
make test-coverage-threshold # Run with 60% minimum coverage gate
make test-coverage-by-pkg    # Show per-package coverage breakdown
make test-race               # Run with Go race detector
make ci                      # Full CI pipeline locally (lint + test + coverage + race)
```

## Test Structure

```
tests/
  benchmark/       # Performance benchmarks (go test -bench)
  e2e/             # End-to-end tests (require Docker/cloud credentials)
  integration/     # Integration tests (mocked external services)
  load/            # Load testing

pkg/
  <package>/
    <file>.go
    <file>_test.go  # Unit tests co-located with source
```

## Coverage Requirements

The project enforces a **60% minimum** total coverage threshold in both CI and the `make test-coverage-threshold` target. This threshold is deliberately set to be achievable while still catching regressions.

Current coverage targets by package priority:

| Priority | Package | Target | Notes |
|----------|---------|--------|-------|
| P0 | `pkg/detector` | 85%+ | Core drift detection logic |
| P0 | `pkg/diff` | 95%+ | State comparison — correctness critical |
| P0 | `pkg/falco` | 85%+ | Event parsing from Falco |
| P1 | `pkg/provider` | 80%+ | Provider adapters |
| P1 | `pkg/terraform` | 85%+ | Terraform state/approval logic |
| P1 | `pkg/config` | 90%+ | Configuration validation |
| P2 | `pkg/output` | 80%+ | Output formatting |
| P2 | `pkg/notifier` | 90%+ | Notification delivery |
| P2 | `pkg/api/handlers` | 70%+ | HTTP API handlers |
| P3 | `cmd/` | 40%+ | CLI entry points |

## Writing Tests

### Unit Tests

Place `_test.go` files alongside source files. Use table-driven tests for comprehensive input coverage:

```go
func TestResourceCategory(t *testing.T) {
    tests := []struct {
        resourceType string
        expected     string
    }{
        {"aws_instance", "compute"},
        {"unknown_resource", ""},
    }
    for _, tt := range tests {
        t.Run(tt.resourceType, func(t *testing.T) {
            assert.Equal(t, tt.expected, resourceCategory(tt.resourceType))
        })
    }
}
```

### HTTP Handler Tests

Use `net/http/httptest` with real dependency instances (not mocks) where possible:

```go
func TestGetCorrelations(t *testing.T) {
    correlator := detector.NewCrossCloudCorrelator(10 * time.Minute)
    handler := handlers.NewCorrelationsHandler(correlator)

    req := httptest.NewRequest("GET", "/api/v1/correlations", nil)
    w := httptest.NewRecorder()
    handler.GetCorrelations(w, req)

    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Functions Requiring Cloud APIs

For functions that call real AWS/GCP/Azure APIs (DiscoverResources, CompareState), test:
- Nil/empty input handling
- Error path behavior
- Helper/utility functions exhaustively
- Mark integration-level tests with build tags: `//go:build e2e`

## Development Cycle

### Before Committing

```bash
make fmt                     # Format code
make test                    # Run unit tests
make test-coverage-by-pkg    # Check coverage didn't regress
```

### Before Opening a PR

```bash
make ci                      # Full pipeline: fmt + lint + coverage threshold + race
```

### After Adding New Code

Every new exported function should have at least one test. For new packages, create a `<package>_test.go` file with baseline tests before merging.

Packages currently missing test files (tech debt backlog):
- `pkg/api/broadcaster`
- `pkg/api/middleware`
- `pkg/api/sse`
- `pkg/api/websocket`
- `pkg/aws`
- `pkg/graph`
- `pkg/falco/mappings`

## CI Pipeline

The GitHub Actions CI workflow (`ci.yml`) enforces:

1. `golangci-lint` — static analysis
2. `gofmt` — formatting check
3. `go vet` — suspicious constructs
4. `go mod tidy` — dependency hygiene
5. Unit tests with `-race` flag
6. **Coverage threshold: 60% minimum** (fails the build if below)
7. Codecov upload for trend tracking

## Running Specific Tests

```bash
# Single package
go test -v ./pkg/detector/...

# Single test function
go test -v -run TestCrossCloudCorrelator_SameUserMultiCloud ./pkg/detector/...

# With coverage for one package
go test -coverprofile=pkg_cov.out ./pkg/provider/...
go tool cover -func=pkg_cov.out

# Benchmarks
go test -bench=. -benchmem ./tests/benchmark/...
```
