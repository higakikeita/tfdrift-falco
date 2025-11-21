# GitHub Actions Workflows

This directory contains the CI/CD workflows for TFDrift-Falco. All workflows are automated and run on GitHub Actions.

## Workflow Overview

| Workflow | Trigger | Duration | Purpose |
|----------|---------|----------|---------|
| **test.yml** | Push/PR | ~5 min | Unit tests, race detection, coverage |
| **integration.yml** | Push/PR | ~2 min | Integration tests with mocked dependencies |
| **benchmark.yml** | Push/PR | ~3 min | Performance benchmarks with comparison |
| **e2e.yml** | On-demand/Scheduled | ~30 min | End-to-end tests with real AWS & Falco |

## Workflows

### 1. test.yml - Unit Tests

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

**Jobs:**
- **test**: Run unit tests on Go 1.21, 1.22, 1.23 with race detection
  - Coverage threshold: 78% (goal: 80%+)
  - Uploads coverage to Codecov
- **test-coverage-report**: Generate coverage report and comment on PR
- **test-race**: Run race detector
- **test-summary**: Summarize all test results on PR

**What it tests:**
- All Go packages in the repository
- Data race detection
- Code coverage with threshold enforcement
- Multi-version Go compatibility

### 2. integration.yml - Integration Tests

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`

**Jobs:**
- **integration-tests**: Run integration tests with race detection
  - Tests webhook notifications (Slack, Discord)
  - Tests Falco event parsing
  - Tests state loading and comparison
  - Uses mocked HTTP servers (no real AWS/Falco required)
- **integration-summary**: Comment PR with test results

**What it tests:**
- Component interactions with mocked dependencies
- Notification system (Slack/Discord webhooks)
- Falco gRPC event parsing
- State loading from various backends
- No external dependencies required

**Duration:** ~2 minutes

### 3. benchmark.yml - Performance Benchmarks

**Triggers:**
- Push to `main` or `develop`
- Pull requests to `main` or `develop`
- Manual workflow dispatch

**Jobs:**
- **benchmark**: Run benchmarks and compare against base branch
  - Compares performance against base branch using `benchstat`
  - Flags performance degradations >20%
  - Comments on PR with comparison results
  - Uploads results as artifacts
- **benchmark-baseline**: Update baseline on main branch
  - Runs on every push to `main`
  - Stores baseline results for future comparisons
  - Commits baseline to `.github/benchmark-baselines/`

**What it tests:**
- Event processing throughput (events/sec)
- Memory allocations (bytes/op, allocs/op)
- State comparison performance
- Drift detection for EC2, IAM, S3
- Concurrent event handling

**Performance Targets:**
- ✅ >5,000 events/sec single-threaded
- ✅ <10KB memory per event
- ✅ Linear scaling with concurrent workloads

**Duration:** ~3 minutes

### 4. e2e.yml - End-to-End Tests

**Triggers:**
- Manual workflow dispatch
- Scheduled (weekly on Sundays at 2 AM UTC)
- Pull requests with `run-e2e` or `run-e2e-quick` label

**Jobs:**
- **e2e-tests**: Full E2E test cycle
  - Sets up real AWS infrastructure (Terraform)
  - Starts Falco with Docker Compose
  - Runs E2E tests that modify AWS resources
  - Validates drift detection and alerting
  - Cleans up infrastructure
  - Takes 20-30 minutes due to CloudTrail propagation delays
- **e2e-quick**: Quick E2E tests (no CloudTrail delay)
  - Runs with `-short` flag
  - Skips CloudTrail-dependent tests
  - Takes ~5 minutes

**What it tests:**
- Complete system with real AWS & Falco
- Real CloudTrail event generation
- Real Falco event processing
- Real drift detection and alerting
- Test scenarios:
  1. EC2 termination protection changes
  2. IAM assume role policy changes
  3. S3 bucket encryption changes
  4. Multiple concurrent changes
  5. User context validation

**Requirements:**
- AWS credentials in GitHub secrets:
  - `E2E_AWS_ACCESS_KEY_ID`
  - `E2E_AWS_SECRET_ACCESS_KEY`
  - `E2E_CLOUDTRAIL_BUCKET`

**Duration:** 20-30 minutes (full), 5 minutes (quick)

## Running Tests Locally

### Unit Tests
```bash
# Run all unit tests
make test

# Run with coverage
make test-coverage

# Run with race detector
make test-race
```

### Integration Tests
```bash
# Run integration tests
make test-integration

# Run from test directory
cd tests/integration
go test -v ./...
```

### Benchmark Tests
```bash
# Run benchmarks
make test-benchmark

# Run from test directory
cd tests/benchmark
go test -bench=. -benchmem -benchtime=10s

# Compare against previous results
benchstat old.txt new.txt
```

### E2E Tests
```bash
# Verify prerequisites
cd tests/e2e
make verify

# Start Falco
make docker-up

# Set up infrastructure
make setup

# Run tests
make test

# Clean up
make cleanup
make docker-down

# Or run full cycle
make all
```

## CI/CD Best Practices

### For Contributors

1. **All PRs trigger**:
   - Unit tests (test.yml)
   - Integration tests (integration.yml)
   - Benchmark tests (benchmark.yml)

2. **E2E tests are optional**:
   - Add label `run-e2e` to PR to trigger full E2E tests
   - Add label `run-e2e-quick` for quick E2E tests without CloudTrail delay
   - E2E tests run automatically on schedule (weekly)

3. **Performance regressions**:
   - Benchmark workflow flags >20% degradations
   - Review benchmark comparison in PR comments
   - Investigate and optimize if performance degrades

4. **Coverage requirements**:
   - Minimum: 78%
   - Goal: 80%+
   - Coverage report posted as PR comment

### For Maintainers

1. **Scheduled E2E tests**:
   - Run weekly on Sundays at 2 AM UTC
   - Create GitHub issue if tests fail
   - Requires AWS credentials in secrets

2. **Benchmark baselines**:
   - Automatically updated on push to `main`
   - Stored in `.github/benchmark-baselines/`
   - Used for PR comparisons

3. **Adding new secrets**:
   ```
   E2E_AWS_ACCESS_KEY_ID       - AWS access key for E2E tests
   E2E_AWS_SECRET_ACCESS_KEY   - AWS secret key for E2E tests
   E2E_CLOUDTRAIL_BUCKET       - S3 bucket with CloudTrail logs
   CODECOV_TOKEN               - Codecov upload token (optional)
   ```

4. **Workflow maintenance**:
   - Update Go versions in test.yml matrix
   - Adjust coverage thresholds as needed
   - Review E2E test costs (AWS resources)
   - Monitor workflow execution times

## Troubleshooting

### Unit Tests Failing

1. Check test output in workflow logs
2. Run locally: `make test`
3. Check for race conditions: `make test-race`
4. Verify coverage: `make test-coverage`

### Integration Tests Failing

1. Check if mock servers are working correctly
2. Run locally: `make test-integration`
3. Review webhook payload formats
4. Check Falco event parsing logic

### Benchmark Degradation

1. Review benchmark comparison in PR comment
2. Run locally: `make test-benchmark`
3. Profile hot paths: `go test -bench=. -cpuprofile=cpu.prof`
4. Analyze allocations: `go test -bench=. -memprofile=mem.prof`

### E2E Tests Failing

1. **AWS credentials**: Verify secrets are configured
2. **Falco not starting**: Check Docker Compose logs
3. **CloudTrail delay**: E2E tests take 20-30 minutes
4. **Infrastructure issues**: Check Terraform state
5. **Cleanup failed**: Manually run `make cleanup` in tests/e2e/

**Common issues:**
```bash
# Check Falco logs
docker logs tfdrift-e2e-falco

# Verify AWS credentials
aws sts get-caller-identity

# Check Terraform state
cd tests/e2e/terraform
terraform show

# Manual cleanup
cd tests/e2e
make cleanup
```

### Coverage Below Threshold

1. Check which packages need more tests
2. See coverage report in PR comment
3. Generate HTML report: `make test-coverage`
4. Add tests for uncovered code paths

## Workflow Dependencies

### Required GitHub Actions
- `actions/checkout@v4` - Checkout code
- `actions/setup-go@v5` - Set up Go
- `actions/upload-artifact@v4` - Upload artifacts
- `actions/github-script@v7` - Run JavaScript in workflows
- `codecov/codecov-action@v4` - Upload coverage
- `aws-actions/configure-aws-credentials@v4` - AWS credentials
- `hashicorp/setup-terraform@v3` - Set up Terraform

### External Tools
- **benchstat**: `go install golang.org/x/perf/cmd/benchstat@latest`
- **golangci-lint**: For linting (not in CI yet)
- **Terraform**: For E2E infrastructure
- **Docker Compose**: For Falco in E2E tests

## Workflow Status Badges

Add these to your README.md:

```markdown
[![Test](https://github.com/your-org/tfdrift-falco/actions/workflows/test.yml/badge.svg)](https://github.com/your-org/tfdrift-falco/actions/workflows/test.yml)
[![Integration Tests](https://github.com/your-org/tfdrift-falco/actions/workflows/integration.yml/badge.svg)](https://github.com/your-org/tfdrift-falco/actions/workflows/integration.yml)
[![Benchmark](https://github.com/your-org/tfdrift-falco/actions/workflows/benchmark.yml/badge.svg)](https://github.com/your-org/tfdrift-falco/actions/workflows/benchmark.yml)
[![E2E Tests](https://github.com/your-org/tfdrift-falco/actions/workflows/e2e.yml/badge.svg)](https://github.com/your-org/tfdrift-falco/actions/workflows/e2e.yml)
```

## Cost Considerations

### E2E Tests

E2E tests create real AWS resources:
- EC2 instances (t2.micro)
- IAM roles and policies
- S3 buckets
- VPC and networking
- CloudTrail (if not already enabled)

**Estimated cost per run:** ~$0.10-0.50 (mostly EC2 and CloudTrail)

**Optimization:**
- Run E2E tests on-demand (label-triggered)
- Schedule weekly instead of daily
- Use spot instances for EC2
- Clean up resources immediately after tests

### GitHub Actions Minutes

Free tier: 2,000 minutes/month for private repos

**Estimated usage per month:**
- Unit tests: ~5 min × 30 runs = 150 min
- Integration tests: ~2 min × 30 runs = 60 min
- Benchmark tests: ~3 min × 30 runs = 90 min
- E2E tests: ~30 min × 4 runs = 120 min
- **Total:** ~420 minutes/month

**Optimization:**
- Use `if` conditions to skip unnecessary runs
- Cache Go modules and build artifacts
- Run E2E tests only on schedule or label

## Future Enhancements

1. **Linting workflow**: Add golangci-lint
2. **Security scanning**: Add CodeQL, gosec
3. **Docker image builds**: Automate image publishing
4. **Release automation**: Automate GitHub releases
5. **Performance tracking**: Historical benchmark tracking
6. **Multi-region E2E**: Test in multiple AWS regions
7. **Load testing**: Add load test workflow
8. **Dependency updates**: Dependabot integration

## Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Codecov Documentation](https://docs.codecov.com/)
- [Terraform GitHub Actions](https://developer.hashicorp.com/terraform/tutorials/automation/github-actions)
- [AWS Actions for GitHub](https://github.com/aws-actions)
