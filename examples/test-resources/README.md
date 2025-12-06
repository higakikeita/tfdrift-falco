# TFDrift-Falco Test Resources

This directory contains Terraform configurations for testing drift detection across multiple AWS resource types.

## Prerequisites

1. AWS CLI configured with credentials
2. Terraform 1.0+
3. S3 bucket for Terraform state
4. CloudTrail enabled and configured
5. Falco with CloudTrail plugin running

## Resources Created

This configuration creates the following resources for drift testing:

| Resource Type | Resource | Test Scenario |
|---------------|----------|---------------|
| **EC2** | `aws_instance.test` | Change `disable_api_termination` via console |
| **IAM** | `aws_iam_role.test` | Modify `assume_role_policy` via console |
| **IAM** | `aws_iam_policy.test` | Modify policy document via console |
| **S3** | `aws_s3_bucket.test` | Change bucket tags via console |
| **S3** | `aws_s3_bucket_versioning.test` | Disable versioning via console |
| **S3** | `aws_s3_bucket_public_access_block.test` | Change public access settings via console |
| **VPC** | `aws_security_group.test` | Add/remove ingress rules via console |

## Setup Instructions

### 1. Configure S3 Backend

Edit `main.tf` and replace `YOUR_BUCKET_NAME` with your S3 bucket:

```hcl
backend "s3" {
  bucket = "my-terraform-state-bucket"  # Your bucket name
  key    = "tfdrift-test/terraform.tfstate"
  region = "us-east-1"
}
```

### 2. Initialize and Apply

```bash
cd examples/test-resources

# Initialize Terraform
terraform init

# Review plan
terraform plan

# Apply configuration
terraform apply
```

This will create all test resources and store the state in S3.

### 3. Configure TFDrift-Falco

Create a configuration file `config.yaml`:

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state-bucket"  # Your bucket name
      s3_key: "tfdrift-test/terraform.tfstate"
      s3_region: "us-east-1"

falco:
  enabled: true
  hostname: "localhost"
  port: 5060

drift_rules:
  - name: "EC2 Instance Changes"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "disable_api_termination"
      - "instance_type"
    severity: "high"

  - name: "IAM Policy Changes"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
    watched_attributes:
      - "assume_role_policy"
      - "policy"
    severity: "critical"

  - name: "S3 Security Changes"
    resource_types:
      - "aws_s3_bucket"
      - "aws_s3_bucket_versioning"
      - "aws_s3_bucket_public_access_block"
    watched_attributes:
      - "versioning_configuration"
      - "block_public_acls"
      - "block_public_policy"
    severity: "critical"

  - name: "Security Group Changes"
    resource_types:
      - "aws_security_group"
    watched_attributes:
      - "ingress"
      - "egress"
    severity: "high"

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

  falco_output:
    enabled: true
    priority: "info"

logging:
  level: "info"
  format: "json"
```

### 4. Start TFDrift-Falco

```bash
# Start Falco (if not already running)
cd deployments
docker-compose up -d falco

# Start TFDrift-Falco
cd ../..
./bin/tfdrift --config config.yaml
```

## Drift Testing Scenarios

### Test 1: EC2 Termination Protection

1. Get EC2 instance ID from output:
   ```bash
   terraform output ec2_instance_id
   # Output: i-0123456789abcdef0
   ```

2. Change termination protection via AWS Console:
   - EC2 Console → Instances → Select instance
   - Actions → Instance settings → Change termination protection
   - **Disable** termination protection

3. **Expected Result**:
   - TFDrift-Falco detects drift within 5-15 minutes
   - Alert sent to Slack
   - Falco output shows drift event

### Test 2: IAM Role Policy

1. Get IAM role name:
   ```bash
   terraform output iam_role_name
   # Output: tfdrift-test-role
   ```

2. Modify assume role policy via AWS Console:
   - IAM Console → Roles → Select role
   - Trust relationships → Edit trust policy
   - Change `"Service": "ec2.amazonaws.com"` to `"Service": "*"`

3. **Expected Result**:
   - Critical drift alert
   - Shows old vs new assume role policy

### Test 3: S3 Bucket Versioning

1. Get S3 bucket name:
   ```bash
   terraform output s3_bucket_name
   # Output: tfdrift-test-bucket-a1b2c3d4
   ```

2. Disable versioning via AWS Console:
   - S3 Console → Buckets → Select bucket
   - Properties → Bucket Versioning → **Suspend**

3. **Expected Result**:
   - Critical drift alert
   - Shows versioning change from Enabled to Suspended

### Test 4: S3 Public Access Block

1. Using the same bucket, change public access settings via AWS Console:
   - S3 Console → Buckets → Select bucket
   - Permissions → Block public access → Edit
   - **Uncheck** "Block all public access"

2. **Expected Result**:
   - Critical drift alert
   - Shows public access block changes

### Test 5: Security Group Rules

1. Get security group ID:
   ```bash
   terraform output security_group_id
   # Output: sg-0123456789abcdef0
   ```

2. Add a new ingress rule via AWS Console:
   - EC2 Console → Security Groups → Select SG
   - Inbound rules → Edit inbound rules → Add rule
   - Add SSH (22) from 0.0.0.0/0

3. **Expected Result**:
   - High severity drift alert
   - Shows added ingress rule

### Test 6: IAM Policy Document

1. Get IAM policy ARN:
   ```bash
   terraform output iam_policy_arn
   # Output: arn:aws:iam::123456789012:policy/tfdrift-test-policy
   ```

2. Modify policy document via AWS Console:
   - IAM Console → Policies → Select policy
   - Edit policy → Add new action (e.g., `s3:DeleteObject`)

3. **Expected Result**:
   - Critical drift alert
   - Shows policy diff

## Verification

### Check CloudTrail Events

```bash
# View recent CloudTrail events
aws cloudtrail lookup-events \
  --lookup-attributes AttributeKey=ResourceName,AttributeValue=<resource-id> \
  --max-results 5
```

### Check Falco Logs

```bash
# View Falco logs
docker logs -f tfdrift-falco
```

### Check TFDrift-Falco Logs

```bash
# If running in foreground, logs will appear in terminal
# If using JSON logging, you can filter:
./bin/tfdrift --config config.yaml | jq 'select(.action == "drift_detected")'
```

## Expected Timing

- **CloudTrail Delay**: 5-15 minutes
- **Falco Processing**: < 1 second
- **TFDrift Detection**: < 1 second
- **Total Time**: ~5-15 minutes from console change to alert

## Cleanup

```bash
# Destroy all test resources
terraform destroy

# This will remove:
# - EC2 instance
# - IAM role and policy
# - S3 bucket (must be empty)
# - Security group
# - VPC
```

## Troubleshooting

### No Drift Detected

1. **Check Terraform state is in S3**:
   ```bash
   aws s3 ls s3://YOUR_BUCKET/tfdrift-test/
   ```

2. **Check TFDrift loaded state**:
   Look for log line: `Loaded Terraform state: X resources`

3. **Check CloudTrail is working**:
   ```bash
   aws cloudtrail get-event-selectors --trail-name YOUR_TRAIL
   ```

4. **Check Falco is receiving events**:
   ```bash
   docker logs tfdrift-falco | grep cloudtrail
   ```

### Drift Not Matching

1. **Check resource ID format**: Terraform state IDs must match CloudTrail resource IDs
2. **Check drift rules**: Ensure resource types and attributes are configured
3. **Check attribute names**: Use exact attribute names from Terraform state

### State Loading Errors

1. **S3 Access Denied**:
   - Check IAM permissions for S3 bucket access
   - Ensure TFDrift-Falco has correct AWS credentials

2. **State Not Found**:
   - Verify S3 bucket and key path
   - Check region configuration

## Notes

- All resources have `lifecycle.ignore_changes` to prevent Terraform from reverting manual changes
- Resources are tagged with `ManagedBy = "terraform"` for easy identification
- Use `terraform state list` to see all resources in state
- Use `terraform show` to see full state details
