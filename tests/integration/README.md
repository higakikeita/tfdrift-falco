# Integration Tests

Integration tests for TFDrift-Falco components with mocked external dependencies.

## Overview

Integration tests validate component interactions without requiring real AWS or Falco infrastructure. External dependencies are mocked using Go interfaces and test doubles.

## Test Coverage

### Falco Integration (`falco_grpc_test.go`)
- gRPC client connection
- Event stream subscription
- Event parsing
- Reconnection logic
- Error handling

### Webhook Notifications (`*_webhook_test.go`)
- Slack notification formatting
- Discord notification formatting
- Webhook HTTP client behavior
- Retry logic
- Rate limiting

### State Management (`state_loader_test.go`)
- Terraform state parsing
- S3 backend interaction (mocked)
- Resource indexing
- State refresh logic

## Running Tests

```bash
# All integration tests
go test -v ./...

# Specific test
go test -v -run TestFalcoGRPC

# With race detection
go test -v -race ./...

# With coverage
go test -v -coverprofile=coverage.out ./...
```

## Test Structure

Integration tests use:
- **httptest**: Mock HTTP servers for webhooks
- **Mock interfaces**: Fake AWS/Falco clients
- **Test fixtures**: Sample events and configurations

Example:
```go
func TestSlackWebhook(t *testing.T) {
    // Create mock HTTP server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "POST", r.Method)

        // Parse webhook payload
        var payload SlackPayload
        json.NewDecoder(r.Body).Decode(&payload)

        // Verify content
        assert.Contains(t, payload.Text, "drift detected")

        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Test slack notifier with mock server
    notifier := NewSlackNotifier(server.URL)
    err := notifier.Send(testAlert)
    assert.NoError(t, err)
}
```

## Key Differences from E2E Tests

| Aspect | Integration Tests | E2E Tests |
|--------|------------------|-----------|
| Speed | Fast (<1s per test) | Slow (minutes) |
| Dependencies | Mocked | Real AWS/Falco |
| Scope | Component interaction | Full workflow |
| CI/CD | Every PR | Main branch only |

## Adding New Tests

1. Create test file: `component_test.go`
2. Add mock implementations in `mocks.go`
3. Use table-driven tests for multiple scenarios
4. Document expected behavior

---

**Next**: [Benchmark Tests](../benchmark/README.md)
