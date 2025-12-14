# Lambda (AWS Lambda) — Drift Coverage

## Overview

TFDrift-Falco monitors AWS Lambda for configuration drift by tracking CloudTrail events related to functions, permissions, event source mappings, and concurrency settings. This enables real-time detection of manual changes made outside of Terraform workflows.

## Supported CloudTrail Events

### Function Management (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateFunction | Lambda function created | WARNING | ✔ |
| DeleteFunction | Lambda function deleted | CRITICAL | ✔ |
| UpdateFunctionCode | Function code updated | WARNING | ✔ |
| UpdateFunctionConfiguration | Function configuration updated | WARNING | ✔ |

### Permissions (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| AddPermission | Permission added to function | WARNING | ✔ |
| RemovePermission | Permission removed from function | WARNING | ✔ |

### Event Source Mappings (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateEventSourceMapping | Event source mapping created | WARNING | ✔ |
| DeleteEventSourceMapping | Event source mapping deleted | WARNING | ✔ |
| UpdateEventSourceMapping | Event source mapping updated | WARNING | ✔ |

### Concurrency (1 event)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| PutFunctionConcurrency | Concurrency limit configured | WARNING | ✔ |

**Total: 10 CloudTrail events**

> **Note**: Lambda alias events (CreateAlias, DeleteAlias, UpdateAlias) share the same event names with KMS aliases and cannot be distinguished without the eventSource field in the current implementation.

## Supported Terraform Resources

- `aws_lambda_function` — Lambda function configuration and code
- `aws_lambda_permission` — Resource-based policy for function invocation
- `aws_lambda_event_source_mapping` — Event source configuration (SQS, Kinesis, DynamoDB Streams, etc.)
- `aws_lambda_alias` — Function version aliases (limited support due to KMS event name collision)

## Monitored Drift Attributes

### Lambda Functions
- **function_name** — Function identifier
- **runtime** — Runtime environment (python3.12, nodejs20.x, etc.)
- **handler** — Function entry point
- **role** — IAM role ARN for execution
- **memory_size** — Memory allocation (MB)
- **timeout** — Execution timeout (seconds)
- **environment** — Environment variables
- **vpc_config** — VPC configuration (subnets, security groups)
- **dead_letter_config** — DLQ configuration
- **tracing_config** — X-Ray tracing mode
- **kms_key_arn** — Environment variable encryption key
- **layers** — Lambda layer ARNs
- **reserved_concurrent_executions** — Concurrency limit

### Lambda Permissions
- **statement_id** — Permission identifier
- **action** — Allowed action (lambda:InvokeFunction)
- **principal** — Service or account principal
- **source_arn** — Event source ARN (e.g., API Gateway, S3 bucket)
- **source_account** — AWS account ID

### Event Source Mappings
- **event_source_arn** — Source ARN (SQS queue, Kinesis stream, DynamoDB table)
- **function_name** — Target Lambda function
- **enabled** — Mapping enabled state
- **batch_size** — Number of records per batch
- **maximum_batching_window_in_seconds** — Batching window
- **starting_position** — Stream reading position (LATEST, TRIM_HORIZON)
- **filter_criteria** — Event filtering patterns

## Falco Rule Examples

```yaml
# Function Lifecycle
- rule: Lambda Function Created
  desc: Detect when a Lambda function is created
  condition: >
    ct.name="CreateFunction"
  output: >
    Lambda function created
    (user=%ct.user function=%ct.request.functionName runtime=%ct.request.runtime
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, serverless]

# Critical Deletion Event
- rule: Lambda Function Deleted
  desc: Detect when a Lambda function is deleted
  condition: >
    ct.name="DeleteFunction"
  output: >
    Lambda function deleted
    (user=%ct.user function=%ct.request.functionName
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, serverless, security]

# Code Deployment
- rule: Lambda Function Code Updated
  desc: Detect when Lambda function code is updated
  condition: >
    ct.name="UpdateFunctionCode"
  output: >
    Lambda function code updated
    (user=%ct.user function=%ct.request.functionName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, deployment]

# Configuration Changes
- rule: Lambda Function Configuration Changed
  desc: Detect when Lambda function configuration is updated
  condition: >
    ct.name="UpdateFunctionConfiguration"
  output: >
    Lambda function configuration changed
    (user=%ct.user function=%ct.request.functionName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, configuration]

# Permission Changes
- rule: Lambda Permission Added
  desc: Detect when permission is added to Lambda function
  condition: >
    ct.name="AddPermission"
  output: >
    Lambda permission added
    (user=%ct.user function=%ct.request.functionName principal=%ct.request.principal
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, security, iam]

# Event Source Mapping
- rule: Lambda Event Source Mapping Created
  desc: Detect when event source mapping is created
  condition: >
    ct.name="CreateEventSourceMapping"
  output: >
    Lambda event source mapping created
    (user=%ct.user function=%ct.request.functionName source=%ct.request.eventSourceArn
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, integration]

# Concurrency Settings
- rule: Lambda Concurrency Limit Set
  desc: Detect when concurrency limit is configured
  condition: >
    ct.name="PutFunctionConcurrency"
  output: >
    Lambda concurrency limit configured
    (user=%ct.user function=%ct.request.functionName limit=%ct.request.reservedConcurrentExecutions
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, lambda, performance]
```

## Example Drift Scenarios

### Scenario 1: Function Memory Size Changed in Console

**CloudTrail Event:**
```json
{
  "eventName": "UpdateFunctionConfiguration",
  "requestParameters": {
    "functionName": "process-orders",
    "memorySize": 1024,
    "timeout": 30
  },
  "userIdentity": {
    "principalId": "AIDAI23ABCD4EFGH5IJKL",
    "userName": "ops-engineer"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_lambda_function.process_orders
Changed: memory_size = 512 → 1024
Changed: timeout = 10 → 30
User: ops-engineer (IAM User)
Region: us-east-1
Severity: HIGH
```

### Scenario 2: Unauthorized Permission Added

**CloudTrail Event:**
```json
{
  "eventName": "AddPermission",
  "requestParameters": {
    "functionName": "secure-function",
    "statementId": "AllowS3Invoke",
    "action": "lambda:InvokeFunction",
    "principal": "s3.amazonaws.com",
    "sourceArn": "arn:aws:s3:::untrusted-bucket"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_lambda_permission.secure_function
Added: statement_id = AllowS3Invoke
Added: source_arn = arn:aws:s3:::untrusted-bucket
User: developer@example.com (Console)
Region: us-east-1
Severity: CRITICAL
```

### Scenario 3: Event Source Mapping Modified

**CloudTrail Event:**
```json
{
  "eventName": "UpdateEventSourceMapping",
  "requestParameters": {
    "uuid": "12345678-1234-1234-1234-123456789012",
    "functionName": "stream-processor",
    "batchSize": 500,
    "enabled": false
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_lambda_event_source_mapping.stream_processor
Changed: batch_size = 100 → 500
Changed: enabled = true → false
User: admin (Assumed Role)
Region: us-east-1
Severity: HIGH
```

### Scenario 4: Concurrency Limit Added

**CloudTrail Event:**
```json
{
  "eventName": "PutFunctionConcurrency",
  "requestParameters": {
    "functionName": "high-traffic-api",
    "reservedConcurrentExecutions": 50
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_lambda_function.high_traffic_api
Changed: reserved_concurrent_executions = null → 50
User: sre-team (Assumed Role)
Region: us-east-1
Severity: MEDIUM
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "Lambda Function Configuration"
    resource_types:
      - "aws_lambda_function"
    watched_attributes:
      - "runtime"
      - "handler"
      - "memory_size"
      - "timeout"
      - "environment"
      - "vpc_config"
      - "role"
    severity: "high"

  - name: "Lambda Permissions"
    resource_types:
      - "aws_lambda_permission"
    watched_attributes:
      - "principal"
      - "source_arn"
      - "action"
    severity: "critical"

  - name: "Lambda Event Sources"
    resource_types:
      - "aws_lambda_event_source_mapping"
    watched_attributes:
      - "event_source_arn"
      - "enabled"
      - "batch_size"
      - "filter_criteria"
    severity: "medium"
```

## Grafana Dashboard Metrics

### Function Metrics
- Lambda function updates by region
- Memory size changes over time
- Runtime version distribution
- Function deletion events

### Permission Metrics
- Permission additions by principal
- Unauthorized permission attempts
- Cross-account access grants

### Event Source Metrics
- Event source mapping changes
- Batch size adjustments
- Disabled/enabled transitions

### User Activity
- Top users making Lambda changes
- Changes by source (Console, CLI, CloudFormation)
- Changes by time of day

## Known Limitations

### 1. Lambda Alias Events
- CreateAlias, DeleteAlias, UpdateAlias events are shared with KMS
- Context-specific detection requires eventSource field
- Lambda aliases are mapped to KMS aliases in current implementation
- Future versions may add context-aware mapping

### 2. Environment Variable Visibility
- Environment variable values are omitted from CloudTrail logs for security
- Only detects that environment configuration changed, not specific values
- Use Terraform state comparison for detailed environment diff

### 3. Lambda Layer Versions
- PublishLayerVersion event not currently tracked
- Layer version changes detected via UpdateFunctionConfiguration
- Consider monitoring layer ARN changes in function configuration

### 4. Function URL Configuration
- CreateFunctionUrlConfig, UpdateFunctionUrlConfig events not tracked
- Planned for future enhancement
- Monitor function URL changes via Terraform state

### 5. Code Package Content
- ZipFile parameter omitted from CloudTrail logs
- Code changes detected by event, not by content diff
- Use S3 versioning or ECR tags for code version tracking

## Best Practices

### 1. Function Configuration Management
```hcl
# Terraform - Separate code from configuration
resource "aws_lambda_function" "app" {
  function_name = "my-app"
  role          = aws_iam_role.lambda.arn

  # Configuration
  runtime       = "python3.12"
  handler       = "main.handler"
  memory_size   = 512
  timeout       = 30

  # Code (managed separately)
  s3_bucket         = aws_s3_bucket.lambda_code.id
  s3_key            = "app-${var.version}.zip"
  source_code_hash  = data.aws_s3_object.app_code.version_id

  # Environment variables (use sensitive values from SSM/Secrets Manager)
  environment {
    variables = {
      STAGE       = "production"
      LOG_LEVEL   = "INFO"
      DB_HOST_SSM = "/app/db/host"  # Reference, not actual value
    }
  }

  # Track all changes except source_code_hash
  lifecycle {
    ignore_changes = [source_code_hash]
  }
}
```

### 2. Permission Management
```hcl
# Explicit permission resources
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "apigateway.amazonaws.com"

  # Scope to specific API
  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/*"
}

resource "aws_lambda_permission" "s3" {
  statement_id  = "AllowS3Invoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.app.function_name
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.uploads.arn

  # Ensure bucket exists first
  depends_on = [aws_s3_bucket.uploads]
}
```

### 3. Event Source Mapping
```hcl
# SQS trigger with filtering
resource "aws_lambda_event_source_mapping" "queue" {
  event_source_arn = aws_sqs_queue.orders.arn
  function_name    = aws_lambda_function.processor.arn

  batch_size                         = 10
  maximum_batching_window_in_seconds = 5

  # Event filtering
  filter_criteria {
    filter {
      pattern = jsonencode({
        body = {
          orderType = ["premium", "express"]
        }
      })
    }
  }

  # Error handling
  function_response_types = ["ReportBatchItemFailures"]
}

# Kinesis stream trigger
resource "aws_lambda_event_source_mapping" "stream" {
  event_source_arn  = aws_kinesis_stream.events.arn
  function_name     = aws_lambda_function.analytics.arn
  starting_position = "LATEST"

  batch_size                         = 100
  maximum_batching_window_in_seconds = 10
  parallelization_factor             = 10

  # Retry and error handling
  maximum_retry_attempts             = 3
  maximum_record_age_in_seconds      = 86400
  bisect_batch_on_function_error     = true
  destination_config {
    on_failure {
      destination_arn = aws_sqs_queue.dlq.arn
    }
  }
}
```

### 4. Concurrency Management
```hcl
# Reserve concurrency for critical functions
resource "aws_lambda_function" "critical_api" {
  function_name = "critical-payment-processor"
  runtime       = "nodejs20.x"
  handler       = "index.handler"
  role          = aws_iam_role.lambda.arn

  # Reserve dedicated capacity
  reserved_concurrent_executions = 100

  # Ensure consistent performance
  memory_size = 1024
  timeout     = 60
}

# Unreserved concurrency for non-critical functions
resource "aws_lambda_function" "background_job" {
  function_name = "background-cleanup"
  runtime       = "python3.12"
  handler       = "main.handler"
  role          = aws_iam_role.lambda.arn

  # Share account concurrency pool
  reserved_concurrent_executions = -1  # No reservation
}
```

### 5. VPC Configuration
```hcl
# VPC-enabled Lambda with monitoring
resource "aws_lambda_function" "vpc_app" {
  function_name = "vpc-app"
  runtime       = "python3.12"
  handler       = "app.handler"
  role          = aws_iam_role.lambda_vpc.arn

  # VPC configuration
  vpc_config {
    subnet_ids         = aws_subnet.private[*].id
    security_group_ids = [aws_security_group.lambda.id]
  }

  # Track VPC changes
  lifecycle {
    prevent_destroy = true
  }
}

# Security group for Lambda
resource "aws_security_group" "lambda" {
  name        = "lambda-app-sg"
  description = "Security group for Lambda function"
  vpc_id      = aws_vpc.main.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "lambda-app"
  }
}
```

## Security Considerations

### 1. IAM Role Permissions
- Monitor changes to execution role ARN
- Alert on privilege escalation via role changes
- Use least-privilege IAM policies
- Track AssumeRolePolicyDocument modifications

### 2. Function Permissions (Resource Policy)
- Monitor AddPermission events for unauthorized access
- Restrict principal to known services/accounts
- Use source_arn to scope permissions
- Alert on wildcard principals or source ARNs

### 3. Environment Variables
- Never store secrets in environment variables
- Use AWS Secrets Manager or SSM Parameter Store
- Monitor environment configuration changes
- Encrypt environment variables with KMS

### 4. VPC Configuration
- Track changes to VPC, subnets, security groups
- Alert on transitions from VPC to no-VPC
- Monitor security group rule changes
- Ensure private subnet routing for RDS/ElastiCache access

### 5. Code Source
- Use S3 versioning or ECR for code storage
- Monitor UpdateFunctionCode events
- Verify code integrity with source_code_hash
- Implement code signing for production functions

### 6. Concurrency Limits
- Monitor PutFunctionConcurrency events
- Alert on concurrency limit removal
- Track regional concurrency pool exhaustion
- Use reserved concurrency for critical functions

## Troubleshooting

### High Alert Volume
- **Problem**: Too many alerts for frequent code deployments
- **Solution**: Use lifecycle.ignore_changes for source_code_hash
- Consider filtering UpdateFunctionCode events in CI/CD deployments

### Missing Configuration Changes
- **Problem**: UpdateFunctionConfiguration events not detected
- **Solution**: Verify CloudTrail is logging Lambda events
- Check Falco plugin configuration for Lambda event filtering
- Ensure eventName extraction is working correctly

### Permission Drift Not Detected
- **Problem**: AddPermission events not generating alerts
- **Solution**: Verify Lambda permissions are managed by Terraform
- Check that statement_id matches between Terraform and CloudTrail
- Ensure resource policy changes are monitored

### Event Source Mapping Issues
- **Problem**: Event source changes not appearing
- **Solution**: Verify UUID extraction from CloudTrail events
- Check that mapping is managed by Terraform (not console-created)
- Ensure eventSourceArn matches Terraform configuration

### Alias Event Collisions
- **Problem**: Lambda alias changes detected as KMS changes
- **Solution**: Current limitation due to shared event names
- Use Terraform state diff to identify actual resource type
- Future enhancement will add eventSource-aware mapping

## Related Documentation

- [AWS Lambda CloudTrail Events](https://docs.aws.amazon.com/lambda/latest/dg/logging-using-cloudtrail.html)
- [Terraform aws_lambda_function](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_function)
- [Terraform aws_lambda_permission](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_permission)
- [Terraform aws_lambda_event_source_mapping](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_event_source_mapping)
- [Terraform aws_lambda_alias](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/lambda_alias)

## Version History

- **v0.3.0** (2025 Q1) - Lambda Enhanced support with 10 CloudTrail events
  - Function management (create, delete, code update, configuration)
  - Permissions (add, remove)
  - Event source mappings (create, update, delete)
  - Concurrency management
  - Comprehensive drift detection for serverless applications
