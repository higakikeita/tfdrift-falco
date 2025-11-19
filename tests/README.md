# TFDrift-Falco Test Suite

Comprehensive testing infrastructure for TFDrift-Falco including E2E, integration, and benchmark tests.

## 📁 Test Structure

```
tests/
├── README.md                          # This file
├── e2e/                               # End-to-End tests (requires AWS & Falco)
│   ├── README.md                      # E2E setup guide
│   ├── terraform/                     # Test infrastructure
│   ├── fixtures/                      # Test data & configs
│   ├── drift_detection_test.go       # Main drift detection scenarios
│   ├── unmanaged_resource_test.go    # Unmanaged resource detection
│   ├── notification_test.go          # Notification delivery tests
│   ├── falco_reconnect_test.go       # Falco reconnection scenarios
│   ├── helpers.go                     # E2E test helpers
│   ├── docker-compose.e2e.yml        # Falco + TFDrift stack
│   └── Makefile                       # E2E test commands
├── integration/                       # Integration tests (lighter than E2E)
│   ├── README.md                      # Integration test guide
│   ├── falco_grpc_test.go            # Falco gRPC communication
│   ├── slack_webhook_test.go         # Slack notification (mocked)
│   ├── discord_webhook_test.go       # Discord notification (mocked)
│   ├── state_loader_test.go          # Terraform state loading
│   └── helpers.go                     # Integration test helpers
└── benchmark/                         # Performance & benchmark tests
    ├── README.md                      # Benchmark guide
    ├── event_processing_bench_test.go # Event processing throughput
    ├── state_comparison_bench_test.go # State comparison performance
    ├── memory_usage_test.go           # Memory leak detection
    ├── concurrent_events_bench_test.go # Concurrent event handling
    └── fixtures/                      # Benchmark test data
```

## 🎯 Test Types

### 1. E2E Tests (`tests/e2e/`)

**Purpose**: Validate complete workflows with real AWS resources and Falco.

**Requirements**:
- AWS account with credentials
- Falco running (Docker or native)
- S3 bucket for Terraform state
- CloudTrail configured

**Run**:
```bash
cd tests/e2e
make test-e2e
```

**Key Scenarios**:
- EC2 instance drift detection
- IAM policy change detection
- S3 bucket encryption drift
- Unmanaged resource detection
- Notification delivery
- Falco reconnection

### 2. Integration Tests (`tests/integration/`)

**Purpose**: Test component interactions with mocked external dependencies.

**Requirements**:
- No AWS account needed
- No Falco needed
- Mocked gRPC and webhooks

**Run**:
```bash
cd tests/integration
go test -v ./...
```

**Key Tests**:
- Falco gRPC client connection
- Webhook notification formatting
- State file parsing
- Event parsing logic

### 3. Benchmark Tests (`tests/benchmark/`)

**Purpose**: Measure performance and detect regressions.

**Requirements**:
- No external dependencies
- Uses test fixtures

**Run**:
```bash
cd tests/benchmark
go test -bench=. -benchmem -benchtime=10s
```

**Key Metrics**:
- Events processed per second
- State comparison latency
- Memory usage per event
- Concurrent event handling

## 🚀 Quick Start

### Prerequisites

```bash
# Install dependencies
go mod download

# Install test tools
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Set up AWS credentials
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=xxx
export AWS_REGION=us-east-1

# Set up test infrastructure (one-time)
cd tests/e2e/terraform
terraform init
terraform apply
```

### Run All Tests

```bash
# Unit tests (existing)
make test

# Integration tests
make test-integration

# E2E tests (requires AWS & Falco)
make test-e2e

# Benchmark tests
make test-benchmark

# All tests
make test-all
```

## 📊 Test Coverage Goals

| Test Type | Current | Target | Status |
|-----------|---------|--------|--------|
| Unit Tests | 80.0% | 80%+ | ✅ Achieved |
| Integration Tests | 0% | 70%+ | 🔄 In Progress |
| E2E Tests | 0% | 5 scenarios | 🔄 In Progress |
| Benchmark Tests | 0% | 4 benchmarks | 🔄 In Progress |

## 🔧 Development Workflow

### Adding a New E2E Test

1. Create test scenario in `tests/e2e/`
2. Add required AWS resources to `tests/e2e/terraform/`
3. Update `tests/e2e/fixtures/` with test data
4. Run test: `cd tests/e2e && go test -v -run TestYourScenario`

### Adding a New Benchmark

1. Create benchmark in `tests/benchmark/`
2. Follow naming: `BenchmarkXxx(b *testing.B)`
3. Run: `go test -bench=BenchmarkXxx -benchmem`
4. Compare with baseline: `benchstat old.txt new.txt`

### Debugging Failed Tests

```bash
# Verbose output
go test -v ./tests/e2e/...

# Run specific test
go test -v -run TestDriftDetection ./tests/e2e/

# Enable debug logging
export TFDRIFT_LOG_LEVEL=debug
go test -v ./tests/e2e/...

# Check Falco logs
docker logs tfdrift-falco-falco

# Check TFDrift logs
docker logs tfdrift-falco-app
```

## 📝 Test Data Management

### Fixtures

Test fixtures are located in `*/fixtures/` directories:

```
fixtures/
├── config.yaml           # Test TFDrift config
├── terraform.tfstate     # Sample Terraform state
├── events/               # Sample CloudTrail events
│   ├── ec2_modify.json
│   ├── iam_change.json
│   └── s3_encryption.json
└── expected/             # Expected test outputs
    └── alerts.json
```

### Generating Test Data

```bash
# Capture real CloudTrail event
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=EventName,AttributeValue=ModifyInstanceAttribute \
  --max-results 1 > fixtures/events/ec2_modify.json

# Export Terraform state
terraform show -json > fixtures/terraform.tfstate
```

## 🔒 Security Considerations

### Sensitive Data

- **DO NOT** commit AWS credentials
- **DO NOT** commit real Slack webhooks
- Use `.env` files (gitignored) for secrets
- Use AWS SSM Parameter Store for CI/CD

### Test Cleanup

Always clean up test resources:

```bash
# After E2E tests
cd tests/e2e/terraform
terraform destroy -auto-approve

# Or use the Makefile
make cleanup-e2e
```

## 🤝 CI/CD Integration

### GitHub Actions

E2E tests run on:
- Pull requests (integration tests only)
- Main branch merges (full E2E suite)
- Scheduled runs (nightly benchmarks)

### Test Execution Matrix

| Environment | Unit | Integration | E2E | Benchmark |
|-------------|------|-------------|-----|-----------|
| PR | ✅ | ✅ | ❌ | ❌ |
| Main | ✅ | ✅ | ✅ | ✅ |
| Nightly | ✅ | ✅ | ✅ | ✅ |

## 📖 Additional Resources

- [E2E Test Guide](./e2e/README.md)
- [Integration Test Guide](./integration/README.md)
- [Benchmark Guide](./benchmark/README.md)
- [Troubleshooting](./TROUBLESHOOTING.md)

## 🐛 Troubleshooting

### Common Issues

#### "Falco connection failed"
```bash
# Check Falco is running
docker ps | grep falco

# Check gRPC port
nc -zv localhost 5060

# Check Falco logs
docker logs tfdrift-falco-falco
```

#### "AWS credentials not found"
```bash
# Verify credentials
aws sts get-caller-identity

# Set credentials
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=xxx
```

#### "Test timeout"
```bash
# Increase timeout
go test -timeout 30m ./tests/e2e/...
```

## 🎯 Next Steps

1. ✅ Set up test infrastructure
2. ⏳ Implement E2E tests
3. ⏳ Implement integration tests
4. ⏳ Implement benchmark tests
5. ⏳ Add CI/CD integration

---

**Maintained by**: TFDrift-Falco Team
**Last Updated**: 2025-11-20
