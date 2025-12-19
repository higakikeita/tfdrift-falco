# TFDrift-Falco AWS Terraform Module

This Terraform module deploys TFDrift-Falco on AWS ECS Fargate with all required infrastructure.

## Features

- ✅ **ECS Fargate Deployment** - Serverless container orchestration
- ✅ **IAM Roles & Policies** - Least privilege access to Terraform state and CloudTrail
- ✅ **CloudWatch Logging** - Centralized log management
- ✅ **Security Groups** - Network-level security
- ✅ **CloudWatch Alarms** - CPU and memory monitoring (optional)
- ✅ **Secrets Management** - AWS Secrets Manager integration for Slack webhook

## Prerequisites

- AWS Account with appropriate permissions
- Existing VPC and subnets
- S3 bucket with Terraform state
- S3 bucket with CloudTrail logs
- (Optional) AWS Secrets Manager secret for Slack webhook

## Usage

### Basic Example

```hcl
module "tfdrift_falco" {
  source = "github.com/higakikeita/tfdrift-falco//terraform/aws"

  # Network Configuration
  vpc_id     = "vpc-12345678"
  subnet_ids = ["subnet-12345678", "subnet-87654321"]

  # Terraform State Configuration
  terraform_state_bucket = "my-terraform-state"
  terraform_state_key    = "prod/terraform.tfstate"

  # CloudTrail Configuration
  cloudtrail_bucket = "my-cloudtrail-logs"

  # Tags
  tags = {
    Environment = "production"
    Team        = "security"
  }
}
```

### Advanced Example with All Options

```hcl
module "tfdrift_falco" {
  source = "github.com/higakikeita/tfdrift-falco//terraform/aws"

  # Basic Configuration
  cluster_name = "tfdrift-falco-prod"
  aws_region   = "us-east-1"

  # Network Configuration
  vpc_id           = "vpc-12345678"
  subnet_ids       = ["subnet-12345678", "subnet-87654321"]
  assign_public_ip = false  # Use private subnets with NAT Gateway

  # Terraform State Configuration
  terraform_state_bucket      = "my-terraform-state"
  terraform_state_key         = "prod/terraform.tfstate"
  terraform_state_kms_key_arn = "arn:aws:kms:us-east-1:123456789012:key/abc-123"

  # CloudTrail Configuration
  cloudtrail_bucket        = "my-cloudtrail-logs"
  cloudtrail_sqs_queue_arn = "arn:aws:sqs:us-east-1:123456789012:cloudtrail-events"

  # ECS Task Configuration
  task_cpu      = "1024"  # 1 vCPU
  task_memory   = "2048"  # 2 GB
  desired_count = 2       # High availability with 2 tasks

  # Container Versions
  falco_version   = "0.37.1"
  tfdrift_version = "0.5.0"

  # Notification Configuration
  slack_webhook_secret_arn = aws_secretsmanager_secret.slack_webhook.arn

  # Monitoring Configuration
  enable_cloudwatch_alarms = true
  alarm_sns_topic_arn      = aws_sns_topic.alerts.arn
  log_retention_days       = 30

  # Metrics Access
  metrics_allowed_cidr_blocks = ["10.0.0.0/8"]

  # Tags
  tags = {
    Environment = "production"
    Team        = "security"
    ManagedBy   = "Terraform"
    Application = "TFDrift-Falco"
  }
}

# Store Slack webhook URL in Secrets Manager
resource "aws_secretsmanager_secret" "slack_webhook" {
  name = "tfdrift-falco/slack-webhook"
  tags = {
    Environment = "production"
  }
}

resource "aws_secretsmanager_secret_version" "slack_webhook" {
  secret_id     = aws_secretsmanager_secret.slack_webhook.id
  secret_string = "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
}

# SNS Topic for CloudWatch Alarms
resource "aws_sns_topic" "alerts" {
  name = "tfdrift-falco-alerts"
}

resource "aws_sns_topic_subscription" "alerts_email" {
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "email"
  endpoint  = "security-team@example.com"
}
```

### Multi-Region Deployment

```hcl
# US-East-1
module "tfdrift_falco_us_east_1" {
  source = "github.com/higakikeita/tfdrift-falco//terraform/aws"

  cluster_name               = "tfdrift-falco-us-east-1"
  aws_region                 = "us-east-1"
  vpc_id                     = data.aws_vpc.us_east_1.id
  subnet_ids                 = data.aws_subnets.us_east_1_private.ids
  terraform_state_bucket     = "terraform-state-us-east-1"
  terraform_state_key        = "us-east-1/terraform.tfstate"
  cloudtrail_bucket          = "cloudtrail-logs-us-east-1"
}

# AP-Northeast-1
module "tfdrift_falco_ap_northeast_1" {
  source = "github.com/higakikeita/tfdrift-falco//terraform/aws"

  cluster_name               = "tfdrift-falco-ap-northeast-1"
  aws_region                 = "ap-northeast-1"
  vpc_id                     = data.aws_vpc.ap_northeast_1.id
  subnet_ids                 = data.aws_subnets.ap_northeast_1_private.ids
  terraform_state_bucket     = "terraform-state-ap-northeast-1"
  terraform_state_key        = "ap-northeast-1/terraform.tfstate"
  cloudtrail_bucket          = "cloudtrail-logs-ap-northeast-1"
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| cluster_name | Name of the ECS cluster | string | "tfdrift-falco" | no |
| aws_region | AWS region to deploy TFDrift-Falco | string | "us-east-1" | no |
| vpc_id | VPC ID where TFDrift-Falco will be deployed | string | - | yes |
| subnet_ids | List of subnet IDs for ECS tasks | list(string) | - | yes |
| assign_public_ip | Assign public IP to ECS tasks | bool | false | no |
| terraform_state_bucket | S3 bucket containing Terraform state | string | - | yes |
| terraform_state_key | S3 key for Terraform state file | string | "terraform.tfstate" | no |
| terraform_state_kms_key_arn | KMS key ARN for encrypted Terraform state | string | "" | no |
| cloudtrail_bucket | S3 bucket containing CloudTrail logs | string | - | yes |
| cloudtrail_sqs_queue_arn | SQS queue ARN for CloudTrail notifications | string | "" | no |
| task_cpu | CPU units for ECS task | string | "512" | no |
| task_memory | Memory for ECS task in MB | string | "1024" | no |
| desired_count | Number of ECS tasks to run | number | 1 | no |
| falco_version | Falco container image version | string | "0.37.1" | no |
| tfdrift_version | TFDrift-Falco container image version | string | "latest" | no |
| slack_webhook_secret_arn | AWS Secrets Manager ARN containing Slack webhook URL | string | "" | no |
| enable_cloudwatch_alarms | Enable CloudWatch alarms for CPU and memory | bool | true | no |
| alarm_sns_topic_arn | SNS topic ARN for CloudWatch alarm notifications | string | "" | no |
| log_retention_days | CloudWatch Logs retention in days | number | 30 | no |
| metrics_allowed_cidr_blocks | CIDR blocks allowed to access metrics endpoint | list(string) | ["10.0.0.0/8"] | no |
| tags | Tags to apply to all resources | map(string) | {} | no |

## Outputs

| Name | Description |
|------|-------------|
| ecs_cluster_name | Name of the ECS cluster |
| ecs_cluster_arn | ARN of the ECS cluster |
| ecs_service_name | Name of the ECS service |
| ecs_task_definition_arn | ARN of the ECS task definition |
| security_group_id | Security group ID for TFDrift-Falco ECS tasks |
| ecs_task_role_arn | IAM role ARN for ECS tasks |
| ecs_task_execution_role_arn | IAM role ARN for ECS task execution |
| cloudwatch_log_group_name | CloudWatch log group name |
| cloudwatch_log_group_arn | CloudWatch log group ARN |

## Viewing Logs

### Using AWS Console

1. Navigate to CloudWatch → Log groups
2. Select `/ecs/tfdrift-falco` (or your custom cluster name)
3. View log streams for `falco` and `tfdrift`

### Using AWS CLI

```bash
# View TFDrift-Falco logs
aws logs tail /ecs/tfdrift-falco --follow --filter-pattern "tfdrift"

# View Falco logs
aws logs tail /ecs/tfdrift-falco --follow --filter-pattern "falco"
```

## Monitoring

### CloudWatch Metrics

The module creates CloudWatch alarms for:
- **CPU Utilization** - Alert when > 80% for 10 minutes
- **Memory Utilization** - Alert when > 80% for 10 minutes

### Custom Metrics

TFDrift-Falco exposes Prometheus metrics on port 9090:
- `tfdrift_events_total` - Total drift events detected
- `tfdrift_events_by_type` - Events by resource type
- `tfdrift_detection_latency_seconds` - Detection latency

## Cost Estimation

Typical costs for running TFDrift-Falco on AWS ECS Fargate:

| Resource | Configuration | Estimated Monthly Cost (us-east-1) |
|----------|---------------|-----------------------------------|
| ECS Fargate Task | 0.5 vCPU, 1 GB RAM, 1 task | $15-20 |
| CloudWatch Logs | 5 GB/month, 30-day retention | $2-3 |
| Data Transfer | Minimal (mainly S3 reads) | $1-2 |
| **Total** | | **~$18-25/month** |

For high availability (2 tasks), multiply by 2: **~$36-50/month**

## Troubleshooting

### Issue: ECS task fails to start

**Check IAM permissions:**
```bash
aws iam get-role --role-name tfdrift-falco-ecs-task
aws iam list-attached-role-policies --role-name tfdrift-falco-ecs-task
```

### Issue: Cannot access Terraform state

**Verify S3 bucket permissions:**
```bash
aws s3 ls s3://my-terraform-state/terraform.tfstate
```

**Check KMS key permissions (if using encryption):**
```bash
aws kms describe-key --key-id YOUR_KMS_KEY_ID
```

### Issue: No CloudTrail events received

**Verify CloudTrail is logging to the correct bucket:**
```bash
aws cloudtrail describe-trails
aws cloudtrail get-trail-status --name my-trail
```

## Upgrading

To upgrade TFDrift-Falco version:

```hcl
module "tfdrift_falco" {
  # ...

  tfdrift_version = "0.6.0"  # Update version
}
```

Then apply:
```bash
terraform plan
terraform apply
```

ECS will perform a rolling update with zero downtime.

## Security Considerations

1. **Use private subnets** - Deploy ECS tasks in private subnets with NAT Gateway
2. **Enable encryption** - Use KMS-encrypted S3 buckets for Terraform state
3. **Rotate secrets** - Regularly rotate Slack webhook URLs and AWS credentials
4. **Restrict metrics access** - Limit `metrics_allowed_cidr_blocks` to monitoring systems only
5. **Enable CloudTrail** - Ensure CloudTrail is enabled in all monitored regions
6. **Use VPC endpoints** - Use VPC endpoints for S3 and Secrets Manager to avoid public internet traffic

## Support

For issues or questions:
- GitHub Issues: https://github.com/higakikeita/tfdrift-falco/issues
- Documentation: https://github.com/higakikeita/tfdrift-falco/docs

## License

MIT License - see [LICENSE](../../LICENSE) for details
