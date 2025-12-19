# TFDrift-Falco AWS Infrastructure Module
# This module sets up all required AWS resources for TFDrift-Falco

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ECS Cluster for TFDrift-Falco
resource "aws_ecs_cluster" "tfdrift" {
  name = var.cluster_name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = merge(
    var.tags,
    {
      Name = var.cluster_name
    }
  )
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "tfdrift" {
  name              = "/ecs/${var.cluster_name}"
  retention_in_days = var.log_retention_days

  tags = var.tags
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "ecs_task_execution" {
  name = "${var.cluster_name}-ecs-task-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# Attach AWS managed policy for ECS task execution
resource "aws_iam_role_policy_attachment" "ecs_task_execution_policy" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# IAM Role for ECS Task (TFDrift-Falco application)
resource "aws_iam_role" "ecs_task" {
  name = "${var.cluster_name}-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

# IAM Policy for TFDrift-Falco (Terraform State access)
resource "aws_iam_policy" "tfdrift_state_access" {
  name        = "${var.cluster_name}-state-access"
  description = "Allow TFDrift-Falco to read Terraform state from S3"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.terraform_state_bucket}",
          "arn:aws:s3:::${var.terraform_state_bucket}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "kms:Decrypt",
          "kms:DescribeKey"
        ]
        Resource = var.terraform_state_kms_key_arn != "" ? [var.terraform_state_kms_key_arn] : []
      }
    ]
  })

  tags = var.tags
}

# Attach State Access Policy
resource "aws_iam_role_policy_attachment" "tfdrift_state_access" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = aws_iam_policy.tfdrift_state_access.arn
}

# IAM Policy for CloudTrail access
resource "aws_iam_policy" "tfdrift_cloudtrail_access" {
  name        = "${var.cluster_name}-cloudtrail-access"
  description = "Allow Falco to read CloudTrail logs"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          "arn:aws:s3:::${var.cloudtrail_bucket}",
          "arn:aws:s3:::${var.cloudtrail_bucket}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = var.cloudtrail_sqs_queue_arn != "" ? [var.cloudtrail_sqs_queue_arn] : []
      }
    ]
  })

  tags = var.tags
}

# Attach CloudTrail Access Policy
resource "aws_iam_role_policy_attachment" "tfdrift_cloudtrail_access" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = aws_iam_policy.tfdrift_cloudtrail_access.arn
}

# Security Group for ECS Tasks
resource "aws_security_group" "tfdrift" {
  name        = "${var.cluster_name}-sg"
  description = "Security group for TFDrift-Falco ECS tasks"
  vpc_id      = var.vpc_id

  # Falco gRPC internal communication
  ingress {
    description = "Falco gRPC"
    from_port   = 5060
    to_port     = 5060
    protocol    = "tcp"
    self        = true
  }

  # Metrics endpoint (optional)
  ingress {
    description = "Prometheus metrics"
    from_port   = 9090
    to_port     = 9090
    protocol    = "tcp"
    cidr_blocks = var.metrics_allowed_cidr_blocks
  }

  # Outbound traffic
  egress {
    description = "Allow all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.cluster_name}-sg"
    }
  )
}

# ECS Task Definition
resource "aws_ecs_task_definition" "tfdrift" {
  family                   = var.cluster_name
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([
    {
      name      = "falco"
      image     = "falcosecurity/falco:${var.falco_version}"
      essential = true

      environment = [
        {
          name  = "AWS_REGION"
          value = var.aws_region
        }
      ]

      portMappings = [
        {
          containerPort = 5060
          protocol      = "tcp"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.tfdrift.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "falco"
        }
      }
    },
    {
      name      = "tfdrift"
      image     = "ghcr.io/higakikeita/tfdrift-falco:${var.tfdrift_version}"
      essential = true

      environment = [
        {
          name  = "AWS_REGION"
          value = var.aws_region
        },
        {
          name  = "TF_STATE_BACKEND"
          value = "s3"
        },
        {
          name  = "TF_STATE_S3_BUCKET"
          value = var.terraform_state_bucket
        },
        {
          name  = "TF_STATE_S3_KEY"
          value = var.terraform_state_key
        }
      ]

      secrets = var.slack_webhook_secret_arn != "" ? [
        {
          name      = "SLACK_WEBHOOK_URL"
          valueFrom = var.slack_webhook_secret_arn
        }
      ] : []

      portMappings = [
        {
          containerPort = 9090
          protocol      = "tcp"
        }
      ]

      dependsOn = [
        {
          containerName = "falco"
          condition     = "START"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.tfdrift.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "tfdrift"
        }
      }
    }
  ])

  tags = var.tags
}

# ECS Service
resource "aws_ecs_service" "tfdrift" {
  name            = var.cluster_name
  cluster         = aws_ecs_cluster.tfdrift.id
  task_definition = aws_ecs_task_definition.tfdrift.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.subnet_ids
    security_groups  = [aws_security_group.tfdrift.id]
    assign_public_ip = var.assign_public_ip
  }

  tags = var.tags
}

# CloudWatch Alarms (optional)
resource "aws_cloudwatch_metric_alarm" "ecs_cpu_high" {
  count               = var.enable_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.cluster_name}-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "TFDrift-Falco CPU utilization is too high"

  dimensions = {
    ClusterName = aws_ecs_cluster.tfdrift.name
    ServiceName = aws_ecs_service.tfdrift.name
  }

  alarm_actions = var.alarm_sns_topic_arn != "" ? [var.alarm_sns_topic_arn] : []

  tags = var.tags
}

resource "aws_cloudwatch_metric_alarm" "ecs_memory_high" {
  count               = var.enable_cloudwatch_alarms ? 1 : 0
  alarm_name          = "${var.cluster_name}-memory-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "MemoryUtilization"
  namespace           = "AWS/ECS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "TFDrift-Falco memory utilization is too high"

  dimensions = {
    ClusterName = aws_ecs_cluster.tfdrift.name
    ServiceName = aws_ecs_service.tfdrift.name
  }

  alarm_actions = var.alarm_sns_topic_arn != "" ? [var.alarm_sns_topic_arn] : []

  tags = var.tags
}
