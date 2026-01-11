# TFDrift-Falco Production-Like Test Environment
# Large-scale AWS infrastructure for comprehensive drift testing

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "TFDrift-Falco-Production-Test"
      Environment = var.environment
      ManagedBy   = "Terraform"
      Owner       = var.owner
    }
  }
}

# Local variables
locals {
  name_prefix = "${var.environment}-tfdrift"
  azs         = slice(data.aws_availability_zones.available.names, 0, var.az_count)

  tags = {
    Environment = var.environment
    Project     = "TFDrift-Falco"
  }
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_caller_identity" "current" {}

# ==================================================
# VPC - Multi-AZ Configuration
# ==================================================
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${local.name_prefix}-vpc"
  cidr = var.vpc_cidr

  azs              = local.azs
  public_subnets   = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k)]
  private_subnets  = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 10)]
  database_subnets = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 20)]

  # NAT Gateway - One per AZ for high availability
  enable_nat_gateway   = true
  single_nat_gateway   = false
  one_nat_gateway_per_az = true

  # DNS settings
  enable_dns_hostnames = true
  enable_dns_support   = true

  # VPC Flow Logs
  enable_flow_log                      = true
  create_flow_log_cloudwatch_iam_role  = true
  create_flow_log_cloudwatch_log_group = true

  # EKS tags for subnet discovery
  public_subnet_tags = {
    "kubernetes.io/role/elb" = 1
    "kubernetes.io/cluster/${local.name_prefix}-eks" = "shared"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = 1
    "kubernetes.io/cluster/${local.name_prefix}-eks" = "shared"
  }

  tags = local.tags
}

# ==================================================
# Security Groups
# ==================================================

# ALB Security Group
resource "aws_security_group" "alb" {
  name        = "${local.name_prefix}-alb-sg"
  description = "Security group for Application Load Balancer"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "HTTP from internet"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    description = "HTTPS from internet"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-alb-sg"
  })
}

# ECS Task Security Group
resource "aws_security_group" "ecs_tasks" {
  name        = "${local.name_prefix}-ecs-tasks-sg"
  description = "Security group for ECS tasks"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Traffic from ALB"
    from_port       = 0
    to_port         = 65535
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-ecs-tasks-sg"
  })
}

# EKS Cluster Security Group (additional rules)
resource "aws_security_group" "eks_cluster_additional" {
  name        = "${local.name_prefix}-eks-cluster-additional-sg"
  description = "Additional security group for EKS cluster"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "HTTPS from VPC"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-eks-cluster-additional-sg"
  })
}

# RDS Security Group
resource "aws_security_group" "rds" {
  name        = "${local.name_prefix}-rds-sg"
  description = "Security group for RDS database"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "PostgreSQL from ECS tasks"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  ingress {
    description = "PostgreSQL from VPC"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-rds-sg"
  })
}

# ElastiCache Security Group
resource "aws_security_group" "elasticache" {
  name        = "${local.name_prefix}-elasticache-sg"
  description = "Security group for ElastiCache Redis"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description     = "Redis from ECS tasks"
    from_port       = 6379
    to_port         = 6379
    protocol        = "tcp"
    security_groups = [aws_security_group.ecs_tasks.id]
  }

  ingress {
    description = "Redis from VPC"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-elasticache-sg"
  })
}

# ==================================================
# S3 Buckets
# ==================================================

# Application Data Bucket
resource "aws_s3_bucket" "app_data" {
  bucket = "${local.name_prefix}-app-data-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-app-data"
    Type = "application-data"
  })
}

resource "aws_s3_bucket_versioning" "app_data" {
  bucket = aws_s3_bucket.app_data.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "app_data" {
  bucket = aws_s3_bucket.app_data.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Backup Bucket
resource "aws_s3_bucket" "backups" {
  bucket = "${local.name_prefix}-backups-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-backups"
    Type = "backups"
  })
}

resource "aws_s3_bucket_versioning" "backups" {
  bucket = aws_s3_bucket.backups.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "backups" {
  bucket = aws_s3_bucket.backups.id

  rule {
    id     = "transition-to-glacier"
    status = "Enabled"

    filter {}

    transition {
      days          = 30
      storage_class = "GLACIER"
    }

    expiration {
      days = 365
    }
  }
}

# Logs Bucket
resource "aws_s3_bucket" "logs" {
  bucket = "${local.name_prefix}-logs-${data.aws_caller_identity.current.account_id}"

  tags = merge(local.tags, {
    Name = "${local.name_prefix}-logs"
    Type = "logs"
  })
}

resource "aws_s3_bucket_lifecycle_configuration" "logs" {
  bucket = aws_s3_bucket.logs.id

  rule {
    id     = "delete-old-logs"
    status = "Enabled"

    filter {}

    expiration {
      days = 90
    }
  }
}

# ==================================================
# ECS Cluster
# ==================================================
resource "aws_ecs_cluster" "main" {
  name = "${local.name_prefix}-ecs-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  configuration {
    execute_command_configuration {
      logging = "OVERRIDE"

      log_configuration {
        cloud_watch_log_group_name = aws_cloudwatch_log_group.ecs.name
      }
    }
  }

  tags = local.tags
}

# ECS Capacity Providers
resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
    base              = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 4
  }
}

# CloudWatch Log Group for ECS
resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${local.name_prefix}"
  retention_in_days = 7

  tags = local.tags
}

# ==================================================
# Application Load Balancer
# ==================================================
resource "aws_lb" "main" {
  name               = "${local.name_prefix}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = module.vpc.public_subnets

  enable_deletion_protection = false
  enable_http2              = true
  enable_cross_zone_load_balancing = true

  access_logs {
    bucket  = aws_s3_bucket.logs.bucket
    prefix  = "alb"
    enabled = false  # Set to true in production
  }

  tags = local.tags
}

# ALB Target Group
resource "aws_lb_target_group" "main" {
  name        = "${local.name_prefix}-tg"
  port        = 80
  protocol    = "HTTP"
  vpc_id      = module.vpc.vpc_id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    interval            = 30
    matcher             = "200"
    path                = "/health"
    port                = "traffic-port"
    protocol            = "HTTP"
    timeout             = 5
    unhealthy_threshold = 3
  }

  deregistration_delay = 30

  tags = local.tags
}

# ALB Listener (HTTP)
resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.main.arn
  }

  tags = local.tags
}

# ==================================================
# EKS Cluster
# ==================================================
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 19.0"

  cluster_name    = "${local.name_prefix}-eks"
  cluster_version = var.eks_cluster_version

  vpc_id                         = module.vpc.vpc_id
  subnet_ids                     = module.vpc.private_subnets
  cluster_endpoint_public_access = true

  # Encryption
  cluster_encryption_config = {
    resources        = ["secrets"]
    provider_key_arn = aws_kms_key.eks.arn
  }

  # Additional security groups
  cluster_additional_security_group_ids = [aws_security_group.eks_cluster_additional.id]

  # EKS Managed Node Groups
  eks_managed_node_groups = {
    general = {
      name           = "${local.name_prefix}-general"
      desired_size   = var.eks_node_desired_size
      min_size       = var.eks_node_min_size
      max_size       = var.eks_node_max_size
      instance_types = var.eks_node_instance_types

      labels = {
        Environment = var.environment
        NodeGroup   = "general"
      }

      tags = local.tags
    }
  }

  # Cluster access
  manage_aws_auth_configmap = false

  tags = local.tags
}

# KMS Key for EKS encryption
resource "aws_kms_key" "eks" {
  description             = "EKS Secret Encryption Key for ${local.name_prefix}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.tags
}

resource "aws_kms_alias" "eks" {
  name          = "alias/${local.name_prefix}-eks"
  target_key_id = aws_kms_key.eks.key_id
}

# ==================================================
# RDS PostgreSQL - Multi-AZ
# ==================================================
resource "aws_db_subnet_group" "main" {
  name       = "${local.name_prefix}-db-subnet-group"
  subnet_ids = module.vpc.database_subnets

  tags = local.tags
}

resource "aws_db_instance" "main" {
  identifier = "${local.name_prefix}-postgres"

  engine               = "postgres"
  engine_version       = "15.15"
  instance_class       = var.rds_instance_class
  allocated_storage    = var.rds_allocated_storage
  storage_type         = "gp3"
  storage_encrypted    = true
  kms_key_id           = aws_kms_key.rds.arn

  db_name  = "appdb"
  username = "dbadmin"
  password = random_password.rds_password.result

  # Multi-AZ for high availability
  multi_az               = true
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  # Backups
  backup_retention_period = 7
  backup_window           = "03:00-04:00"
  maintenance_window      = "mon:04:00-mon:05:00"

  # Performance Insights
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  performance_insights_enabled    = true
  performance_insights_kms_key_id = aws_kms_key.rds.arn

  # Protection
  deletion_protection = false  # Set to true in production
  skip_final_snapshot = true   # Set to false in production

  tags = local.tags
}

# KMS Key for RDS encryption
resource "aws_kms_key" "rds" {
  description             = "RDS Encryption Key for ${local.name_prefix}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.tags
}

resource "aws_kms_alias" "rds" {
  name          = "alias/${local.name_prefix}-rds"
  target_key_id = aws_kms_key.rds.key_id
}

# Random password for RDS
resource "random_password" "rds_password" {
  length  = 32
  special = true
}

# Store RDS password in Secrets Manager
resource "aws_secretsmanager_secret" "rds_password" {
  name = "${local.name_prefix}-rds-password"

  tags = local.tags
}

resource "aws_secretsmanager_secret_version" "rds_password" {
  secret_id     = aws_secretsmanager_secret.rds_password.id
  secret_string = random_password.rds_password.result
}

# ==================================================
# ElastiCache Redis - Cluster Mode
# ==================================================
resource "aws_elasticache_subnet_group" "main" {
  name       = "${local.name_prefix}-redis-subnet-group"
  subnet_ids = module.vpc.private_subnets

  tags = local.tags
}

resource "aws_elasticache_replication_group" "main" {
  replication_group_id = "${local.name_prefix}-redis"
  description          = "Redis cluster for ${local.name_prefix}"

  engine               = "redis"
  engine_version       = "7.0"
  node_type            = var.elasticache_node_type
  num_cache_clusters   = var.elasticache_num_nodes
  parameter_group_name = "default.redis7"
  port                 = 6379

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.elasticache.id]

  # Encryption
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = random_password.redis_auth.result
  kms_key_id                 = aws_kms_key.elasticache.arn

  # Automatic failover
  automatic_failover_enabled = true
  multi_az_enabled           = true

  # Backups
  snapshot_retention_limit = 5
  snapshot_window          = "03:00-05:00"
  maintenance_window       = "mon:05:00-mon:07:00"

  # Logs
  log_delivery_configuration {
    destination      = aws_cloudwatch_log_group.redis.name
    destination_type = "cloudwatch-logs"
    log_format       = "json"
    log_type         = "slow-log"
  }

  tags = local.tags
}

# KMS Key for ElastiCache encryption
resource "aws_kms_key" "elasticache" {
  description             = "ElastiCache Encryption Key for ${local.name_prefix}"
  deletion_window_in_days = 7
  enable_key_rotation     = true

  tags = local.tags
}

# Random auth token for Redis
resource "random_password" "redis_auth" {
  length  = 32
  special = false
}

# CloudWatch Log Group for Redis
resource "aws_cloudwatch_log_group" "redis" {
  name              = "/elasticache/${local.name_prefix}/redis"
  retention_in_days = 7

  tags = local.tags
}

# ==================================================
# IAM Roles and Policies
# ==================================================

# ECS Task Execution Role
resource "aws_iam_role" "ecs_task_execution" {
  name = "${local.name_prefix}-ecs-task-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution" {
  role       = aws_iam_role.ecs_task_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ECS Task Role
resource "aws_iam_role" "ecs_task" {
  name = "${local.name_prefix}-ecs-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })

  tags = local.tags
}

# Policy for ECS tasks to access S3
resource "aws_iam_policy" "ecs_task_s3_access" {
  name        = "${local.name_prefix}-ecs-task-s3-access"
  description = "Allow ECS tasks to access S3 buckets"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.app_data.arn,
          "${aws_s3_bucket.app_data.arn}/*"
        ]
      }
    ]
  })

  tags = local.tags
}

resource "aws_iam_role_policy_attachment" "ecs_task_s3" {
  role       = aws_iam_role.ecs_task.name
  policy_arn = aws_iam_policy.ecs_task_s3_access.arn
}

# ==================================================
# CloudWatch Alarms
# ==================================================

# ALB Target Health Alarm
resource "aws_cloudwatch_metric_alarm" "alb_unhealthy_targets" {
  alarm_name          = "${local.name_prefix}-alb-unhealthy-targets"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "UnHealthyHostCount"
  namespace           = "AWS/ApplicationELB"
  period              = "60"
  statistic           = "Average"
  threshold           = "0"
  alarm_description   = "Alert when ALB has unhealthy targets"

  dimensions = {
    LoadBalancer = aws_lb.main.arn_suffix
    TargetGroup  = aws_lb_target_group.main.arn_suffix
  }

  tags = local.tags
}

# RDS CPU Alarm
resource "aws_cloudwatch_metric_alarm" "rds_cpu_high" {
  alarm_name          = "${local.name_prefix}-rds-cpu-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "CPUUtilization"
  namespace           = "AWS/RDS"
  period              = "300"
  statistic           = "Average"
  threshold           = "80"
  alarm_description   = "Alert when RDS CPU is high"

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.main.id
  }

  tags = local.tags
}
