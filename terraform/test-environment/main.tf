# TFDrift-Falco Test Environment
# Simple AWS resources for testing drift detection

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "TFDrift-Falco-Test"
      Environment = "test"
      ManagedBy   = "Terraform"
    }
  }
}

# ==================================================
# 1. S3 Bucket - Easy to drift (encryption, versioning)
# ==================================================
resource "aws_s3_bucket" "test" {
  bucket = var.test_bucket_name

  tags = {
    Name        = "TFDrift Test Bucket"
    Description = "Bucket for testing drift detection"
  }
}

# S3 Bucket Versioning
resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

# S3 Bucket Server-side Encryption
resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# S3 Bucket Public Access Block
resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ==================================================
# 2. Security Group - Easy to drift (ingress rules)
# ==================================================
resource "aws_security_group" "test" {
  name        = "${var.environment}-tfdrift-test-sg"
  description = "Security group for TFDrift testing"
  vpc_id      = var.vpc_id

  # SSH access (restricted to specific CIDR)
  ingress {
    description = "SSH from trusted network"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  # HTTPS access
  ingress {
    description = "HTTPS"
    from_port   = 443
    to_port     = 443
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/8"]
  }

  # HTTP access (restricted)
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    description = "All outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-tfdrift-test-sg"
  }
}

# ==================================================
# 3. IAM Policy - Easy to drift (policy document)
# ==================================================
resource "aws_iam_policy" "test" {
  name        = "${var.environment}-tfdrift-test-policy"
  description = "Test policy for drift detection"

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
          "${aws_s3_bucket.test.arn}",
          "${aws_s3_bucket.test.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:${var.aws_region}:*:*"
      }
    ]
  })

  tags = {
    Name = "${var.environment}-tfdrift-test-policy"
  }
}

# ==================================================
# 4. IAM Role - For testing
# ==================================================
resource "aws_iam_role" "test" {
  name = "${var.environment}-tfdrift-test-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Name = "${var.environment}-tfdrift-test-role"
  }
}

# Attach policy to role
resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

# ==================================================
# 5. CloudWatch Log Group - For testing
# ==================================================
resource "aws_cloudwatch_log_group" "test" {
  name              = "/tfdrift/test/${var.environment}"
  retention_in_days = 7

  tags = {
    Name = "TFDrift Test Logs"
  }
}

# ==================================================
# 6. SNS Topic - For drift notifications (optional)
# ==================================================
resource "aws_sns_topic" "drift_alerts" {
  name = "${var.environment}-tfdrift-alerts"

  tags = {
    Name = "TFDrift Alert Topic"
  }
}

# SNS Topic Subscription (optional - add your email)
resource "aws_sns_topic_subscription" "drift_alerts_email" {
  count     = var.alert_email != "" ? 1 : 0
  topic_arn = aws_sns_topic.drift_alerts.arn
  protocol  = "email"
  endpoint  = var.alert_email
}
