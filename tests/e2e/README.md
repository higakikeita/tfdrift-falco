# E2E Tests for TFDrift-Falco

End-to-end tests that validate complete drift detection workflows with real AWS resources and Falco.

## 🎯 What Gets Tested

### Test Scenarios

1. **EC2 Drift Detection** (`drift_detection_test.go`)
   - Modify `disable_api_termination`
   - Change `instance_type`
   - Verify drift alert with user context

2. **IAM Policy Drift** (`drift_detection_test.go`)
   - Modify IAM role assume policy
   - Attach/detach policies
   - Verify CRITICAL severity alerts

3. **S3 Bucket Drift** (`drift_detection_test.go`)
   - Disable encryption
   - Change versioning
   - Modify bucket policy

4. **Unmanaged Resource Detection** (`unmanaged_resource_test.go`)
   - Create resource outside Terraform
   - Verify unmanaged resource alert
   - Test auto-import suggestions

5. **Notification Delivery** (`notification_test.go`)
   - Slack webhook (mocked)
   - Discord webhook (mocked)
   - Alert formatting

6. **Falco Reconnection** (`falco_reconnect_test.go`)
   - Simulate Falco restart
   - Verify TFDrift reconnects
   - No event loss

## 🛠️ Setup

### Prerequisites

1. **AWS Account**
   ```bash
   export AWS_ACCESS_KEY_ID=xxx
   export AWS_SECRET_ACCESS_KEY=xxx
   export AWS_REGION=us-east-1
   ```

2. **S3 Bucket for Terraform State**
   ```bash
   aws s3 mb s3://tfdrift-test-state-$(date +%s)
   ```

3. **Docker** (for Falco)
   ```bash
   docker --version
   ```

### Infrastructure Setup

1. **Configure Terraform Backend**
   ```bash
   cd terraform/

   # Create backend config
   cat > backend.hcl <<EOF
   bucket = "your-state-bucket"
   key    = "e2e/terraform.tfstate"
   region = "us-east-1"
   EOF
   ```

2. **Initialize Terraform**
   ```bash
   terraform init -backend-config=backend.hcl
   ```

3. **Create Test Resources**
   ```bash
   # Copy example vars
   cp terraform.tfvars.example terraform.tfvars

   # Edit terraform.tfvars with your values
   vim terraform.tfvars

   # Apply infrastructure
   terraform apply
   ```

4. **Export Outputs for Tests**
   ```bash
   terraform output -json > ../fixtures/terraform_outputs.json
   ```

### Falco Setup

1. **Start Falco with Docker Compose**
   ```bash
   # From tests/e2e/
   docker-compose -f docker-compose.e2e.yml up -d falco
   ```

2. **Verify Falco is Running**
   ```bash
   docker logs tfdrift-e2e-falco
   grpcurl -plaintext localhost:5060 list
   ```

### Test Configuration

1. **Create Test Config**
   ```bash
   cp fixtures/config.yaml.example fixtures/config.yaml
   vim fixtures/config.yaml
   ```

2. **Update with Your Values**
   ```yaml
   falco:
     enabled: true
     hostname: "localhost"
     port: 5060

   providers:
     aws:
       enabled: true
       regions: ["us-east-1"]
       state:
         backend: "s3"
         s3_bucket: "your-state-bucket"
         s3_key: "e2e/terraform.tfstate"
   ```

## 🚀 Running Tests

### All E2E Tests

```bash
# Run all E2E tests
go test -v -timeout 30m ./...

# With race detection
go test -v -race -timeout 30m ./...
```

### Specific Test

```bash
# Single test
go test -v -run TestDriftDetection_EC2Termination

# Test suite
go test -v -run TestDriftDetection
```

### Using Makefile

```bash
# Setup infrastructure
make setup

# Run tests
make test

# Cleanup
make cleanup

# Full cycle
make all
```

## 📝 Test Flow

### Typical E2E Test Flow

```go
func TestDriftDetection_EC2(t *testing.T) {
    // 1. Start TFDrift
    detector := startTFDrift(t)
    defer detector.Stop()

    // 2. Wait for initialization
    waitForReady(t, detector)

    // 3. Trigger AWS change
    changeEC2Termination(t, testInstanceID, false)

    // 4. Wait for alert
    alert := waitForAlert(t, detector, 30*time.Second)

    // 5. Verify alert
    assert.Equal(t, "aws_instance", alert.ResourceType)
    assert.Equal(t, testInstanceID, alert.ResourceID)
    assert.Contains(t, alert.Changes, "disable_api_termination")

    // 6. Cleanup - revert change
    changeEC2Termination(t, testInstanceID, true)
}
```

## 🧹 Cleanup

### After Tests

```bash
# Destroy test infrastructure
cd terraform/
terraform destroy -auto-approve

# Stop Falco
docker-compose -f docker-compose.e2e.yml down

# Clean up state bucket (optional)
aws s3 rm s3://your-state-bucket/e2e/ --recursive
```

### Automatic Cleanup

All resources are tagged with `AutoDelete: true`. You can use the cleanup script:

```bash
# From tests/e2e/
./scripts/cleanup.sh
```

## 🐛 Troubleshooting

### Test Failures

#### "Falco connection refused"
```bash
# Check Falco status
docker ps | grep falco
docker logs tfdrift-e2e-falco

# Restart Falco
docker-compose -f docker-compose.e2e.yml restart falco
```

#### "AWS credentials not found"
```bash
# Verify credentials
aws sts get-caller-identity

# Check environment
env | grep AWS
```

#### "Terraform state not found"
```bash
# Verify state exists
aws s3 ls s3://your-state-bucket/e2e/

# Re-apply infrastructure
cd terraform && terraform apply
```

#### "Alert not received within timeout"
```bash
# Check CloudTrail delay (can be 5-15 minutes)
# Increase timeout
go test -v -run TestName -timeout 30m

# Check TFDrift logs
docker logs tfdrift-e2e-app

# Manually verify CloudTrail event
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=EventName,AttributeValue=ModifyInstanceAttribute
```

### Debug Mode

```bash
# Enable debug logging
export TFDRIFT_LOG_LEVEL=debug
export FALCO_LOG_LEVEL=debug

# Run with verbose output
go test -v -run TestDriftDetection

# Collect logs
docker logs tfdrift-e2e-falco > falco.log
docker logs tfdrift-e2e-app > tfdrift.log
```

## 📊 Expected Results

### Successful Test Run

```
=== RUN   TestDriftDetection_EC2Termination
--- PASS: TestDriftDetection_EC2Termination (45.23s)
=== RUN   TestDriftDetection_IAMPolicy
--- PASS: TestDriftDetection_IAMPolicy (38.91s)
=== RUN   TestUnmanagedResource_EC2
--- PASS: TestUnmanagedResource_EC2 (52.15s)
=== RUN   TestNotification_Slack
--- PASS: TestNotification_Slack (5.12s)
=== RUN   TestFalcoReconnect
--- PASS: TestFalcoReconnect (15.34s)
PASS
ok      github.com/keitahigaki/tfdrift-falco/tests/e2e    156.750s
```

### Performance Expectations

| Test | Expected Duration | Notes |
|------|------------------|-------|
| EC2 Drift | 30-60s | Includes CloudTrail delay |
| IAM Drift | 30-60s | Includes CloudTrail delay |
| S3 Drift | 30-60s | Includes CloudTrail delay |
| Unmanaged Resource | 30-60s | Depends on polling interval |
| Notification | <10s | Mock webhook |
| Falco Reconnect | 10-20s | Connection retry |

## 🔐 Security

### Credentials Management

- Use IAM roles when running on EC2
- Never commit credentials to git
- Use AWS SSM Parameter Store for CI/CD
- Rotate credentials regularly

### Resource Isolation

- Use unique `test_id` to avoid conflicts
- Resources are tagged for easy identification
- Automatic cleanup scripts available

### Cost Management

- All resources use minimal instance sizes
- CloudTrail can reuse existing trail
- Terraform state in S3 (low cost)
- Estimated cost: <$1/day

## 📚 Additional Resources

- [Main Test Suite README](../README.md)
- [Integration Tests](../integration/README.md)
- [Benchmark Tests](../benchmark/README.md)
- [Troubleshooting Guide](../TROUBLESHOOTING.md)

## 🤝 Contributing

When adding new E2E tests:

1. Add corresponding AWS resources to `terraform/main.tf`
2. Create test scenario in appropriate `*_test.go` file
3. Update this README with test description
4. Add expected duration to performance table
5. Test locally before creating PR

---

**Last Updated**: 2025-11-20
