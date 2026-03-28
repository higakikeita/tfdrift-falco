# TFDrift-Falco Local E2E Test Environment

This directory contains a Docker Compose-based end-to-end test environment for TFDrift-Falco that works completely locally without requiring real cloud resources.

## Overview

The local E2E environment includes:

- **Mock Falco gRPC Server** (`mock_falco_server.go`): A lightweight Go server that implements the Falco outputs gRPC service and provides HTTP endpoints to trigger CloudTrail-like events
- **TFDrift-Falco API Server**: Runs in Docker and consumes events from the mock Falco server
- **Terraform State Fixture** (`fixtures/simple.tfstate`): A minimal terraform state file with sample AWS resources
- **Docker Compose Orchestration** (`docker-compose.yml`): Manages container startup, networking, and health checks
- **E2E Tests** (`docker_compose_test.go`): Go test suite with build tag `//go:build e2e` that exercises the full system

## Quick Start

```bash
# Build Docker images
make build

# Start all services
make up

# In another terminal, follow logs
make logs

# In another terminal, run E2E tests
make test

# Stop services
make down
```

## File Structure

```
tests/e2e/local/
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile.mock-falco           # Build image for mock Falco server
├── mock_falco_server.go            # Mock Falco gRPC server implementation
├── docker_compose_test.go           # E2E test suite (tagged: e2e)
├── config.yaml                      # TFDrift configuration for local testing
├── Makefile                         # Development and testing shortcuts
├── README.md                        # This file
└── fixtures/
    └── simple.tfstate               # Sample Terraform state file
```

## Make Targets

- `make up` - Start all services with docker-compose
- `make down` - Stop and remove services
- `make logs` - Follow logs from all services
- `make logs-falco` - Follow logs from mock Falco server
- `make logs-tfdrift` - Follow logs from TFDrift API
- `make test` - Run the E2E test suite
- `make build` - Build Docker images
- `make clean` - Remove all containers and volumes
- `make ps` - Show running containers
- `make health` - Check service health
- `make full-test` - Build, start, test, and cleanup all in one command

## How It Works

### Mock Falco Server

The `mock_falco_server.go` implements a full Falco outputs service server with both gRPC and HTTP interfaces:

**gRPC Service:**
- Implements `outputs.Service` (Falco's output service)
- Listens on port 5060 (configurable)
- Maintains subscriptions for connected clients

**HTTP Control API:**
- Port 8081 (configurable)
- `/health` - Health check endpoint
- `/trigger-event` - Generic event trigger
- `/trigger-ec2-change` - Trigger EC2 instance modification event
- `/trigger-sg-change` - Trigger security group modification event
- `/trigger-s3-change` - Trigger S3 bucket modification event

### Event Triggering Flow

1. Test calls HTTP endpoint on mock Falco server (e.g., `/trigger-ec2-change`)
2. Server creates a fake CloudTrail event with gRPC `outputs.Response` structure
3. Event is broadcast to all connected gRPC clients
4. TFDrift-Falco consumes the event via gRPC subscription
5. Detector processes the event and detects drift
6. Test verifies drift was detected via TFDrift API

### Docker Network

Services communicate via the `tfdrift-network` bridge network:
- `mock-falco` - Mock Falco server (hostname in compose)
- `tfdrift` - TFDrift API server

TFDrift is configured to connect to `mock-falco:5060` (internal Docker network).

## Configuration

### `config.yaml`

The local configuration file is pre-set for the test environment:
- AWS provider enabled with single region (us-east-1)
- Falco connection to `mock-falco:5060`
- Local Terraform state at `/data/terraform.tfstate`
- Drift rules for EC2, security groups, and S3
- All notifications enabled except Slack

To customize, edit `config.yaml` and restart services with `make down && make up`.

### `docker-compose.yml`

Services configuration:
- **mock-falco**: Built from `Dockerfile.mock-falco`, exposes ports 5060 (gRPC) and 8081 (HTTP)
- **tfdrift**: Built from project Dockerfile, exposes port 8080 (API)
- Both have health checks configured
- TFDrift depends on mock-falco service

## Running Tests

### Run all E2E tests:
```bash
make test
```

### Run only a specific test:
```bash
cd ../.. && go test -v -tags e2e -count=1 ./tests/e2e/local/... -run TestLocalE2E/TriggerEC2DriftEvent
```

### View test output with logs:
```bash
# Terminal 1: Follow logs
make logs

# Terminal 2: Run tests
make test
```

## Debugging

### Health checks:
```bash
make health
```

### View logs from specific service:
```bash
make logs-falco    # Mock Falco server
make logs-tfdrift  # TFDrift API
```

### Open shell in container:
```bash
make shell-falco
make shell-tfdrift
```

### Manual event triggering:
```bash
# Trigger EC2 event
curl -X POST http://localhost:8081/trigger-ec2-change

# Check TFDrift API
curl http://localhost:8080/health
```

### View running containers:
```bash
make ps
```

## Limitations

The local E2E environment is designed for testing drift detection logic, not for comprehensive AWS cloud behavior:

- **Mock Falco Server**: Provides synthetic CloudTrail events; does not track actual infrastructure state changes
- **Terraform State**: Static fixture file; not updated when events are triggered
- **No Real AWS Resources**: All events are simulated; does not require AWS credentials
- **Drift Detection**: Works as designed but based on static state vs. simulated events

For testing against real AWS, see the AWS-based E2E tests in the parent `tests/e2e/` directory.

## Troubleshooting

### Services fail to start
```bash
# Check Docker setup
docker --version
docker-compose --version

# Rebuild images
make clean
make build
make up
```

### Services start but health checks fail
```bash
# View detailed logs
make logs

# Test connectivity manually
curl http://localhost:8081/health  # Falco
curl http://localhost:8080/health  # TFDrift
```

### Tests fail with "connection refused"
Ensure services are fully healthy before running tests:
```bash
make health  # Verify both endpoints respond
make test    # Then run tests
```

### Port conflicts
If ports 5060, 8080, or 8081 are in use, modify `docker-compose.yml`:
```yaml
ports:
  - "5061:5060"    # Change host port from 5060 to 5061
```

Then update test constants in `docker_compose_test.go`.

## Development Tips

### Iterative Development

1. **Make changes to `mock_falco_server.go`**:
```bash
make build-mock
make down
make build
make up
```

2. **Make changes to test file**:
```bash
# Services stay running, just rebuild and run tests
make test
```

3. **Make changes to `config.yaml`**:
```bash
make down
make up  # Services restart and load new config
```

### Adding New Test Scenarios

1. Add a handler method in `MockFalcoServer` or HTTP endpoint
2. Add corresponding HTTP endpoint in `HTTPServer`
3. Add a test function in `docker_compose_test.go`
4. Add test to the `TestLocalE2E` test suite

Example:
```go
// mock_falco_server.go
func (h *HTTPServer) handleNewEventType(w http.ResponseWriter, r *http.Request) {
    // Create event and publish
}

// docker_compose_test.go
func testNewScenario(t *testing.T) {
    // Trigger event and verify
}
```

### Performance Considerations

The mock Falco server uses goroutines to handle streaming:
- Event queue: buffered channel (capacity 100)
- Dropped events logged if queue is full
- Consider increasing capacity if running high-volume tests

## Related Documentation

- [TFDrift Main README](../../../README.md)
- [AWS E2E Tests](../README.md)
- [Falco Client-Go](https://github.com/falcosecurity/client-go)
- [Docker Compose Docs](https://docs.docker.com/compose/)
