# Falco Setup Guide for TFDrift-Falco

This guide explains how to set up Falco with the CloudTrail plugin to work with TFDrift-Falco.

## Prerequisites

- Linux system (Ubuntu 20.04+, Debian 11+, RHEL 8+, or similar)
- AWS CloudTrail enabled and logging to S3
- AWS credentials configured (for CloudTrail plugin to access S3)
- Root or sudo access

## Quick Setup (Recommended)

### Step 1: Install Falco

#### Option A: Using the install script (easiest)

```bash
curl -s https://falco.org/repo/falcosecurity-packages.asc | \
  sudo gpg --dearmor -o /usr/share/keyrings/falco-archive-keyring.gpg

echo "deb [signed-by=/usr/share/keyrings/falco-archive-keyring.gpg] https://download.falco.org/packages/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/falcosecurity.list

sudo apt-get update
sudo apt-get install -y falco
```

#### Option B: Using Docker

```bash
docker pull falcosecurity/falco:latest
```

### Step 2: Install Falco CloudTrail Plugin

The CloudTrail plugin is required for TFDrift-Falco to receive AWS events.

```bash
# Download CloudTrail plugin
sudo mkdir -p /usr/share/falco/plugins
sudo curl -L -o /usr/share/falco/plugins/libcloudtrail.so \
  https://download.falco.org/plugins/stable/cloudtrail-latest-x86_64.tar.gz

# Extract the plugin
cd /usr/share/falco/plugins
sudo tar -xzf libcloudtrail.so cloudtrail-latest-x86_64.tar.gz
```

### Step 3: Configure Falco for TFDrift

Create or edit `/etc/falco/falco.yaml`:

```yaml
# Enable gRPC output (required for TFDrift-Falco)
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 0

# Load CloudTrail plugin
plugins:
  - name: cloudtrail
    library_path: /usr/share/falco/plugins/libcloudtrail.so
    init_config:
      # AWS S3 bucket where CloudTrail logs are stored
      s3_bucket: "your-cloudtrail-logs-bucket"

      # AWS region
      aws_region: "us-east-1"

      # Optional: SQS queue for CloudTrail events (recommended for lower latency)
      # sqs_queue: "arn:aws:sqs:us-east-1:123456789012:cloudtrail-events"

      # Optional: Use AWS profile
      # aws_profile: "default"
    open_params: ''

# Load CloudTrail rules for Terraform drift detection
load_plugins: [cloudtrail]

rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
  - /etc/falco/rules.d
  # Add TFDrift-Falco rules
  - /path/to/tfdrift-falco/rules/terraform_drift.yaml

# Output configuration
json_output: true
json_include_output_property: true

# Logging
log_stderr: true
log_syslog: false
log_level: info
```

### Step 4: Copy TFDrift Rules to Falco

```bash
sudo cp /path/to/tfdrift-falco/rules/terraform_drift.yaml /etc/falco/rules.d/
```

### Step 5: Configure AWS Credentials

Falco's CloudTrail plugin needs AWS credentials to access CloudTrail logs:

```bash
# Option 1: IAM instance role (recommended for EC2)
# Attach an IAM role with CloudTrail read permissions to your EC2 instance

# Option 2: AWS credentials file
sudo mkdir -p /root/.aws
sudo cat > /root/.aws/credentials <<EOF
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
EOF
sudo chmod 600 /root/.aws/credentials
```

Required IAM permissions for the CloudTrail plugin:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::your-cloudtrail-bucket",
        "arn:aws:s3:::your-cloudtrail-bucket/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "sqs:ReceiveMessage",
        "sqs:DeleteMessage",
        "sqs:GetQueueAttributes"
      ],
      "Resource": "arn:aws:sqs:*:*:cloudtrail-*"
    }
  ]
}
```

### Step 6: Start Falco

```bash
# Using systemd
sudo systemctl enable falco
sudo systemctl start falco
sudo systemctl status falco

# Check logs
sudo journalctl -u falco -f
```

### Step 7: Verify Falco gRPC is Running

```bash
# Check if Falco gRPC is listening
sudo netstat -tlnp | grep 5060

# Or using ss
sudo ss -tlnp | grep 5060

# Expected output:
# tcp   LISTEN 0      128    0.0.0.0:5060      0.0.0.0:*    users:(("falco",pid=1234,fd=10))
```

### Step 8: Test CloudTrail Plugin

```bash
# Trigger a test CloudTrail event (modify an EC2 instance)
aws ec2 modify-instance-attribute \
  --instance-id i-1234567890abcdef0 \
  --disable-api-termination

# Check Falco logs for the event
sudo journalctl -u falco | grep "EC2 termination protection"
```

## Running with Docker

If you prefer to run Falco in Docker:

### docker-compose.yaml

```yaml
version: '3.8'

services:
  falco:
    image: falcosecurity/falco:latest
    container_name: falco
    privileged: true
    ports:
      - "5060:5060"
    volumes:
      - /var/run/docker.sock:/host/var/run/docker.sock
      - /dev:/host/dev
      - /proc:/host/proc:ro
      - ./rules/terraform_drift.yaml:/etc/falco/rules.d/terraform_drift.yaml:ro
      - ./falco.yaml:/etc/falco/falco.yaml:ro
      - ~/.aws:/root/.aws:ro
    environment:
      - AWS_REGION=us-east-1
    command:
      - /usr/bin/falco
      - -c
      - /etc/falco/falco.yaml

  tfdrift-falco:
    image: tfdrift-falco:latest
    container_name: tfdrift
    depends_on:
      - falco
    volumes:
      - ./config.yaml:/config.yaml:ro
      - ~/.aws:/root/.aws:ro
    environment:
      - AWS_REGION=us-east-1
    command:
      - --config
      - /config.yaml
```

Start the stack:

```bash
docker-compose up -d
```

## Troubleshooting

### Falco gRPC Not Starting

**Symptom**: TFDrift-Falco fails to connect with "connection refused"

**Solutions**:
```bash
# Check if gRPC is enabled in falco.yaml
sudo grep -A 3 "grpc:" /etc/falco/falco.yaml

# Check Falco logs for errors
sudo journalctl -u falco -n 100

# Verify Falco is running
sudo systemctl status falco
```

### CloudTrail Plugin Not Loading

**Symptom**: No CloudTrail events in Falco output

**Solutions**:
```bash
# Check plugin configuration
sudo grep -A 10 "cloudtrail" /etc/falco/falco.yaml

# Verify plugin file exists
ls -l /usr/share/falco/plugins/libcloudtrail.so

# Check AWS credentials
sudo -u root aws s3 ls s3://your-cloudtrail-bucket

# Check CloudTrail plugin logs in Falco output
sudo journalctl -u falco | grep cloudtrail
```

### No Events Received

**Symptom**: Falco is running but no events are received

**Solutions**:
```bash
# Verify CloudTrail is enabled
aws cloudtrail describe-trails

# Check CloudTrail logs are being written
aws s3 ls s3://your-cloudtrail-bucket/ --recursive | tail

# Verify SQS queue (if using)
aws sqs get-queue-attributes --queue-url YOUR_QUEUE_URL

# Test with a manual AWS action
aws ec2 describe-instances --max-results 1
```

### Permission Denied Errors

**Symptom**: Falco CloudTrail plugin reports S3 access denied

**Solutions**:
```bash
# Verify IAM permissions
aws iam get-user

# Test S3 access manually
aws s3 ls s3://your-cloudtrail-bucket/

# Check if bucket policy blocks access
aws s3api get-bucket-policy --bucket your-cloudtrail-bucket
```

## Advanced Configuration

### Using SQS for Real-Time Events (Recommended)

For lower latency, configure CloudTrail to send events to SQS:

1. Create an SQS queue:
```bash
aws sqs create-queue --queue-name cloudtrail-events
```

2. Configure CloudTrail to send events to SQS:
```bash
aws cloudtrail put-event-selectors \
  --trail-name my-trail \
  --event-selectors '[{"ReadWriteType":"All","IncludeManagementEvents":true}]' \
  --advanced-event-selectors '[{"Name":"Log all management events","FieldSelectors":[{"Field":"eventCategory","Equals":["Management"]}]}]'
```

3. Update Falco CloudTrail plugin config to use SQS:
```yaml
plugins:
  - name: cloudtrail
    init_config:
      sqs_queue: "arn:aws:sqs:us-east-1:123456789012:cloudtrail-events"
      aws_region: "us-east-1"
```

### Filtering Specific Events

To reduce noise, you can filter Falco rules to only Terraform-managed resources:

Edit `/etc/falco/rules.d/terraform_drift.yaml` and uncomment the macro-based rules that check for Terraform tags.

### Performance Tuning

For high-volume environments:

```yaml
# In falco.yaml
grpc:
  threadiness: 4  # Increase for better throughput

plugins:
  - name: cloudtrail
    init_config:
      # Process events in batches
      batch_size: 100

      # Increase worker threads
      num_workers: 4
```

## Next Steps

After Falco is set up and running:

1. Configure TFDrift-Falco: Edit `config.yaml` with your Falco gRPC endpoint
2. Run TFDrift-Falco: `tfdrift --config config.yaml`
3. Test drift detection: Make a manual change to a Terraform-managed resource
4. Check alerts: Verify you receive notifications (Slack, Discord, etc.)

## References

- [Falco Official Documentation](https://falco.org/docs/)
- [Falco CloudTrail Plugin](https://github.com/falcosecurity/plugins/tree/master/plugins/cloudtrail)
- [AWS CloudTrail Documentation](https://docs.aws.amazon.com/cloudtrail/)
- [TFDrift-Falco Repository](https://github.com/keitahigaki/tfdrift-falco)
