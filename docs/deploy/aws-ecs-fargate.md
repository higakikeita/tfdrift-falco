# AWS ECS/Fargate Deployment Guide

This guide covers deploying TFDrift-Falco on AWS ECS using AWS Fargate for a serverless container experience.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Architecture Overview](#architecture-overview)
3. [Task Definition](#task-definition)
4. [Service Configuration](#service-configuration)
5. [Load Balancer Setup](#load-balancer-setup)
6. [Environment Variables](#environment-variables)
7. [Secrets Management](#secrets-management)
8. [Auto Scaling](#auto-scaling)
9. [CloudWatch Monitoring](#cloudwatch-monitoring)
10. [Health Checks](#health-checks)
11. [Infrastructure as Code Examples](#infrastructure-as-code-examples)
12. [Deployment Steps](#deployment-steps)
13. [Troubleshooting](#troubleshooting)

## Prerequisites

### AWS Account and Permissions

- AWS Account with appropriate IAM permissions
- IAM user/role with permissions for:
  - ECS (task definition, service, cluster)
  - ECR (container registry)
  - CloudWatch Logs
  - AWS Secrets Manager
  - Elastic Load Balancing (ALB)
  - VPC and networking
  - Auto Scaling

### AWS Resources

- VPC with at least 2 subnets in different availability zones
- Security groups for:
  - ALB (allow HTTP/HTTPS traffic)
  - ECS tasks (allow traffic from ALB)
  - Database access (if using RDS)
- ECR repository (create via `aws ecr create-repository --repository-name tfdrift-falco`)
- CloudWatch Log Group (create via AWS console or CLI)

### Local Tools

- AWS CLI v2+
- Docker (for building and pushing images)
- Terraform or AWS CloudFormation (optional, for IaC)

## Architecture Overview

```
Internet → ALB (port 80/443) → ECS Service → Fargate Tasks
                ↓
         CloudWatch Logs
         ↓
         CloudWatch Metrics → Auto Scaling
```

The architecture consists of:
- **ALB**: Routes traffic to healthy tasks
- **ECS Service**: Manages task lifecycle and scaling
- **Fargate Tasks**: Containerized application instances
- **CloudWatch**: Centralized logging and metrics

## Task Definition

Create an ECS task definition for TFDrift-Falco with recommended settings.

### Memory and CPU Configuration

| Configuration | Memory | CPU | Use Case |
|---|---|---|---|
| Small | 512 MB | 256 | Development/Testing |
| Standard | 1 GB | 512 | Single drift detection |
| Medium | 2 GB | 1024 | Multiple drifts + API |
| Large | 4 GB | 2048 | High-load production |

### Sample Task Definition (JSON)

```json
{
  "family": "tfdrift-falco",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "tfdrift-falco",
      "image": "123456789.dkr.ecr.us-east-1.amazonaws.com/tfdrift-falco:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "hostPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "TFDRIFT_FALCO_PORT",
          "value": "8080"
        },
        {
          "name": "TFDRIFT_FALCO_LOG_LEVEL",
          "value": "info"
        },
        {
          "name": "TFDRIFT_FALCO_LOG_FORMAT",
          "value": "json"
        }
      ],
      "secrets": [
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:tfdrift-falco/jwt-secret:jwt_secret::"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/tfdrift-falco",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ],
  "executionRoleArn": "arn:aws:iam::123456789:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789:role/ecsTaskRole"
}
```

### IAM Roles

**Execution Role** (ecsTaskExecutionRole) - Allows ECS to pull images and write logs:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecr:GetAuthorizationToken",
        "ecr:BatchGetImage",
        "ecr:GetDownloadUrlForLayer"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:us-east-1:123456789:log-group:/ecs/tfdrift-falco:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "secretsmanager:GetSecretValue"
      ],
      "Resource": "arn:aws:secretsmanager:us-east-1:123456789:secret:tfdrift-falco/*"
    }
  ]
}
```

**Task Role** (ecsTaskRole) - Allows application to access AWS services:

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
        "arn:aws:s3:::terraform-state-bucket",
        "arn:aws:s3:::terraform-state-bucket/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudtrail:LookupEvents"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": "arn:aws:iam::*:role/TerraformAuditRole"
    }
  ]
}
```

## Service Configuration

### ECS Service Settings

```json
{
  "serviceName": "tfdrift-falco-service",
  "cluster": "production",
  "taskDefinition": "tfdrift-falco:1",
  "desiredCount": 3,
  "launchType": "FARGATE",
  "platformVersion": "LATEST",
  "networkConfiguration": {
    "awsvpcConfiguration": {
      "subnets": [
        "subnet-12345678",
        "subnet-87654321"
      ],
      "securityGroups": [
        "sg-tfdrift-falco-tasks"
      ],
      "assignPublicIp": "DISABLED"
    }
  },
  "loadBalancers": [
    {
      "targetGroupArn": "arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco/abcd1234",
      "containerName": "tfdrift-falco",
      "containerPort": 8080
    }
  ],
  "deploymentConfiguration": {
    "maximumPercent": 200,
    "minimumHealthyPercent": 100,
    "deploymentCircuitBreaker": {
      "enable": true,
      "rollback": true
    }
  },
  "enableECSManagedTags": true
}
```

## Load Balancer Setup

### Application Load Balancer (ALB) Configuration

```bash
# Create ALB
aws elbv2 create-load-balancer \
  --name tfdrift-falco-alb \
  --subnets subnet-12345678 subnet-87654321 \
  --security-groups sg-alb-tfdrift \
  --scheme internet-facing \
  --type application

# Create target group
aws elbv2 create-target-group \
  --name tfdrift-falco-targets \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-12345678 \
  --health-check-protocol HTTP \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3

# Create listener
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/tfdrift-falco-alb/1234567890abcdef \
  --protocol HTTP \
  --port 80 \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco-targets/1234567890abcdef
```

### HTTPS Configuration

```bash
# Create HTTPS listener (requires SSL certificate)
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:us-east-1:123456789:loadbalancer/app/tfdrift-falco-alb/1234567890abcdef \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=arn:aws:acm:us-east-1:123456789:certificate/12345678-1234-1234-1234-123456789012 \
  --default-actions Type=forward,TargetGroupArn=arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco-targets/1234567890abcdef

# Redirect HTTP to HTTPS
aws elbv2 modify-listener \
  --listener-arn arn:aws:elasticloadbalancing:us-east-1:123456789:listener/app/tfdrift-falco-alb/1234567890abcdef/1234567890abcdef \
  --default-actions Type=redirect,RedirectConfig='{Protocol=HTTPS,Port=443,StatusCode=HTTP_301}'
```

## Environment Variables

### Required Environment Variables

```bash
# Core Configuration
TFDRIFT_FALCO_PORT=8080
TFDRIFT_FALCO_LOG_LEVEL=info
TFDRIFT_FALCO_LOG_FORMAT=json

# Falco Integration
TFDRIFT_FALCO_HOSTNAME=falco.example.com
TFDRIFT_FALCO_PORT_GRPC=5060

# AWS Configuration
AWS_REGION=us-east-1
AWS_S3_BUCKET=terraform-state-bucket
AWS_S3_KEY=prod/terraform.tfstate

# Frontend URLs
VITE_API_BASE_URL=https://api.example.com
VITE_WS_URL=wss://api.example.com/ws
```

### Optional Environment Variables

```bash
# Monitoring
PROMETHEUS_ENABLED=true
METRICS_PORT=9090

# Slack Integration
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
SLACK_ENABLED=false

# Feature Flags
FEATURE_ADVANCED_FILTERING=true
FEATURE_HISTORICAL_ANALYSIS=true
```

## Secrets Management

Use AWS Secrets Manager to manage sensitive data.

### Create Secrets

```bash
# Create JWT secret
aws secretsmanager create-secret \
  --name tfdrift-falco/jwt-secret \
  --description "JWT signing secret for TFDrift-Falco" \
  --secret-string "$(openssl rand -base64 32)"

# Create database credentials (if using RDS)
aws secretsmanager create-secret \
  --name tfdrift-falco/db-password \
  --description "Database password" \
  --secret-string '{"username":"dbuser","password":"secure-password"}'

# Create API keys
aws secretsmanager create-secret \
  --name tfdrift-falco/api-keys \
  --description "External API keys" \
  --secret-string '{"slack_webhook":"https://...","datadog_api_key":"..."}'
```

### Reference Secrets in Task Definition

Secrets are injected as environment variables via the task definition:

```json
"secrets": [
  {
    "name": "JWT_SECRET",
    "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:tfdrift-falco/jwt-secret:jwt_secret::"
  },
  {
    "name": "DB_PASSWORD",
    "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789:secret:tfdrift-falco/db-password:password::"
  }
]
```

### Rotate Secrets

```bash
# Update a secret (creates new version)
aws secretsmanager put-secret-value \
  --secret-id tfdrift-falco/jwt-secret \
  --secret-string "$(openssl rand -base64 32)"

# Force tasks to refresh and pick up new secret
aws ecs update-service \
  --cluster production \
  --service tfdrift-falco-service \
  --force-new-deployment
```

## Auto Scaling

### Application Auto Scaling Configuration

```bash
# Register scalable target
aws application-autoscaling register-scalable-target \
  --service-namespace ecs \
  --resource-id service/production/tfdrift-falco-service \
  --scalable-dimension ecs:service:DesiredCount \
  --min-capacity 3 \
  --max-capacity 10

# Create scaling policy (CPU-based)
aws application-autoscaling put-scaling-policy \
  --policy-name tfdrift-falco-cpu-scaling \
  --service-namespace ecs \
  --resource-id service/production/tfdrift-falco-service \
  --scalable-dimension ecs:service:DesiredCount \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration '{
    "TargetValue": 70.0,
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ECSServiceAverageCPUUtilization"
    },
    "ScaleOutCooldown": 60,
    "ScaleInCooldown": 300
  }'

# Create scaling policy (Memory-based)
aws application-autoscaling put-scaling-policy \
  --policy-name tfdrift-falco-memory-scaling \
  --service-namespace ecs \
  --resource-id service/production/tfdrift-falco-service \
  --scalable-dimension ecs:service:DesiredCount \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration '{
    "TargetValue": 80.0,
    "PredefinedMetricSpecification": {
      "PredefinedMetricType": "ECSServiceAverageMemoryUtilization"
    },
    "ScaleOutCooldown": 60,
    "ScaleInCooldown": 300
  }'

# Create scaling policy (ALB Request Count)
aws application-autoscaling put-scaling-policy \
  --policy-name tfdrift-falco-request-scaling \
  --service-namespace ecs \
  --resource-id service/production/tfdrift-falco-service \
  --scalable-dimension ecs:service:DesiredCount \
  --policy-type TargetTrackingScaling \
  --target-tracking-scaling-policy-configuration '{
    "TargetValue": 1000.0,
    "CustomizedMetricSpecification": {
      "MetricName": "RequestCountPerTarget",
      "Namespace": "AWS/ApplicationELB",
      "Statistic": "Sum",
      "Unit": "Count"
    },
    "ScaleOutCooldown": 60,
    "ScaleInCooldown": 300
  }'
```

### Scaling Configuration Recommendations

```yaml
# Recommended scaling parameters
MinCapacity: 3              # Minimum instances
MaxCapacity: 10             # Maximum instances
DesiredCount: 5             # Initial task count
CPUTarget: 70%              # Target CPU utilization
MemoryTarget: 80%           # Target memory utilization
ScaleOutCooldown: 60s       # Wait before scaling out again
ScaleInCooldown: 300s       # Wait before scaling in again
```

## CloudWatch Monitoring

### Create Log Group

```bash
# Create log group
aws logs create-log-group --log-group-name /ecs/tfdrift-falco

# Set retention policy (30 days)
aws logs put-retention-policy \
  --log-group-name /ecs/tfdrift-falco \
  --retention-in-days 30
```

### CloudWatch Metrics

```bash
# Enable Container Insights (provides more detailed metrics)
aws ecs update-cluster-settings \
  --cluster production \
  --settings name=containerInsights,value=enabled
```

### Create Alarms

```bash
# Alarm for high CPU
aws cloudwatch put-metric-alarm \
  --alarm-name tfdrift-falco-high-cpu \
  --alarm-description "Alert when CPU exceeds 80%" \
  --metric-name CPUUtilization \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 80 \
  --comparison-operator GreaterThanThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:us-east-1:123456789:alerts

# Alarm for task failures
aws cloudwatch put-metric-alarm \
  --alarm-name tfdrift-falco-task-failures \
  --alarm-description "Alert when tasks fail" \
  --metric-name RunningCount \
  --namespace AWS/ECS \
  --statistic Average \
  --period 300 \
  --threshold 2 \
  --comparison-operator LessThanThreshold \
  --evaluation-periods 1 \
  --alarm-actions arn:aws:sns:us-east-1:123456789:alerts

# Alarm for ALB target health
aws cloudwatch put-metric-alarm \
  --alarm-name tfdrift-falco-unhealthy-targets \
  --alarm-description "Alert when targets become unhealthy" \
  --metric-name UnHealthyHostCount \
  --namespace AWS/ApplicationELB \
  --statistic Average \
  --period 60 \
  --threshold 1 \
  --comparison-operator GreaterThanOrEqualToThreshold \
  --evaluation-periods 2 \
  --alarm-actions arn:aws:sns:us-east-1:123456789:alerts
```

### CloudWatch Dashboards

```bash
# Create dashboard with key metrics
aws cloudwatch put-dashboard \
  --dashboard-name TFDrift-Falco-Production \
  --dashboard-body file://dashboard.json
```

Sample `dashboard.json`:
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          [ "AWS/ECS", "CPUUtilization", { "stat": "Average" } ],
          [ ".", "MemoryUtilization", { "stat": "Average" } ],
          [ "AWS/ApplicationELB", "RequestCount", { "stat": "Sum" } ],
          [ ".", "TargetResponseTime", { "stat": "Average" } ],
          [ ".", "HTTPCode_Target_5XX_Count", { "stat": "Sum" } ]
        ],
        "period": 300,
        "stat": "Average",
        "region": "us-east-1",
        "title": "TFDrift-Falco Performance"
      }
    }
  ]
}
```

## Health Checks

### ECS Health Check Configuration

The health check validates that the application is running:

```json
"healthCheck": {
  "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
  "interval": 30,
  "timeout": 5,
  "retries": 3,
  "startPeriod": 60
}
```

### Health Endpoint Implementation

The application should expose a `/health` endpoint:

```go
// Example Go implementation
func healthHandler(w http.ResponseWriter, r *http.Request) {
  // Check critical dependencies
  if !isDatabaseConnected() || !isFalcoConnected() {
    w.WriteHeader(http.StatusServiceUnavailable)
    json.NewEncoder(w).Encode(map[string]string{
      "status": "unhealthy",
      "reason": "database or falco unavailable",
    })
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusOK)
  json.NewEncoder(w).Encode(map[string]string{
    "status": "healthy",
    "timestamp": time.Now().UTC().String(),
  })
}
```

### Target Group Health Check Settings

```bash
aws elbv2 modify-target-group \
  --target-group-arn arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco/1234567890abcdef \
  --health-check-protocol HTTP \
  --health-check-path /health \
  --health-check-interval-seconds 30 \
  --health-check-timeout-seconds 5 \
  --healthy-threshold-count 2 \
  --unhealthy-threshold-count 3
```

## Infrastructure as Code Examples

### AWS CloudFormation Template

See `terraform/` directory for complete CloudFormation templates.

### Terraform Configuration

```hcl
# Provider
provider "aws" {
  region = var.aws_region
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "production"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# Task Definition
resource "aws_ecs_task_definition" "app" {
  family                   = "tfdrift-falco"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "512"
  memory                   = "1024"

  execution_role_arn = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn      = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode([
    {
      name      = "tfdrift-falco"
      image     = "${aws_ecr_repository.app.repository_url}:latest"
      essential = true

      portMappings = [
        {
          containerPort = 8080
          hostPort      = 8080
          protocol      = "tcp"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.app.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])
}

# ECS Service
resource "aws_ecs_service" "app" {
  name            = "tfdrift-falco-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 3
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnets
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.app.arn
    container_name   = "tfdrift-falco"
    container_port   = 8080
  }

  deployment_configuration {
    maximum_percent         = 200
    minimum_healthy_percent = 100

    deployment_circuit_breaker {
      enable   = true
      rollback = true
    }
  }
}

# Auto Scaling
resource "aws_appautoscaling_target" "ecs_target" {
  max_capacity       = 10
  min_capacity       = 3
  resource_id        = "service/${aws_ecs_cluster.main.name}/${aws_ecs_service.app.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "ecs_policy_cpu" {
  policy_type            = "TargetTrackingScaling"
  resource_id            = aws_appautoscaling_target.ecs_target.resource_id
  scalable_dimension     = aws_appautoscaling_target.ecs_target.scalable_dimension
  service_namespace      = aws_appautoscaling_target.ecs_target.service_namespace

  target_tracking_scaling_policy_configuration {
    target_value = 70.0

    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }

    scale_out_cooldown = 60
    scale_in_cooldown  = 300
  }
}
```

## Deployment Steps

### 1. Build and Push Docker Image

```bash
# Build image
docker build -t tfdrift-falco:latest .

# Tag for ECR
docker tag tfdrift-falco:latest 123456789.dkr.ecr.us-east-1.amazonaws.com/tfdrift-falco:latest

# Push to ECR
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/tfdrift-falco:latest
```

### 2. Create Task Definition

```bash
aws ecs register-task-definition --cli-input-json file://task-definition.json
```

### 3. Create or Update Service

```bash
# Create service
aws ecs create-service \
  --cluster production \
  --service-name tfdrift-falco-service \
  --task-definition tfdrift-falco:1 \
  --desired-count 3 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345678,subnet-87654321],securityGroups=[sg-tfdrift],assignPublicIp=DISABLED}" \
  --load-balancers "targetGroupArn=arn:aws:...,containerName=tfdrift-falco,containerPort=8080"

# Or update existing service
aws ecs update-service \
  --cluster production \
  --service tfdrift-falco-service \
  --task-definition tfdrift-falco:2 \
  --force-new-deployment
```

### 4. Verify Deployment

```bash
# Check service status
aws ecs describe-services \
  --cluster production \
  --services tfdrift-falco-service

# Check running tasks
aws ecs list-tasks \
  --cluster production \
  --service-name tfdrift-falco-service

# Check target health
aws elbv2 describe-target-health \
  --target-group-arn arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco/1234567890abcdef

# Watch logs
aws logs tail /ecs/tfdrift-falco --follow
```

## Troubleshooting

### Tasks Not Starting

**Problem**: Tasks stuck in PROVISIONING or PENDING state

```bash
# Check task details
aws ecs describe-tasks \
  --cluster production \
  --tasks arn:aws:ecs:us-east-1:123456789:task/production/task-id

# Common causes:
# 1. Insufficient capacity - increase task size or availability
# 2. Security group blocking traffic - check inbound/outbound rules
# 3. Missing IAM permissions - verify execution role has required permissions
```

### Tasks Failing Health Checks

**Problem**: Healthy count staying below desired count

```bash
# Check target group health
aws elbv2 describe-target-health \
  --target-group-arn arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco/1234567890abcdef

# Check logs for errors
aws logs tail /ecs/tfdrift-falco --follow | grep -i error

# Common causes:
# 1. Health check endpoint not responding - verify /health endpoint
# 2. Application not binding to correct port - check environment variables
# 3. Missing dependencies - verify Falco connectivity, database access
```

### High CPU/Memory Usage

**Problem**: Tasks frequently restarting due to resource exhaustion

```bash
# Increase task memory/CPU in task definition
aws ecs register-task-definition \
  --cli-input-json file://task-definition-larger.json

# Update service to use new task definition
aws ecs update-service \
  --cluster production \
  --service tfdrift-falco-service \
  --task-definition tfdrift-falco:3
```

### Connection to Falco Failing

**Problem**: "Connection refused" errors in logs

```bash
# Verify Falco connectivity
aws ecs execute-command \
  --cluster production \
  --task task-id \
  --container tfdrift-falco \
  --interactive \
  --command "/bin/sh -c 'nc -zv falco.example.com 5060'"

# Check security group rules
aws ec2 describe-security-groups --group-ids sg-tfdrift-tasks
```

### Deployment Stuck During Rollout

**Problem**: Service stuck with old and new tasks unable to complete deployment

```bash
# Check deployment status
aws ecs describe-services \
  --cluster production \
  --services tfdrift-falco-service

# Rollback to previous task definition
aws ecs update-service \
  --cluster production \
  --service tfdrift-falco-service \
  --task-definition tfdrift-falco:1 \
  --force-new-deployment

# Increase deployment timeout
aws ecs update-service \
  --cluster production \
  --service tfdrift-falco-service \
  --deployment-configuration maximumPercent=300,minimumHealthyPercent=50
```

For additional support, check CloudWatch logs and the [AWS ECS documentation](https://docs.aws.amazon.com/ecs/).
