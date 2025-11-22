terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  # S3 backend configuration
  backend "s3" {
    bucket = "tfdrift-falco-state-test"
    key    = "tfdrift-test/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
  default     = "tfdrift-test"
}

# ======================
# EC2 Instance
# ======================

resource "aws_instance" "test" {
  ami                     = data.aws_ami.amazon_linux_2.id
  instance_type           = "t2.micro"
  disable_api_termination = true

  tags = {
    Name        = "${var.project_name}-instance"
    Project     = var.project_name
    ManagedBy   = "terraform"
    Environment = "test"
  }

  lifecycle {
    # Ignore changes that we'll make manually for drift testing
    ignore_changes = [
      disable_api_termination,
      instance_type,
      tags,
    ]
  }
}

data "aws_ami" "amazon_linux_2" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

# ======================
# IAM Role
# ======================

resource "aws_iam_role" "test" {
  name = "${var.project_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })

  tags = {
    Name      = "${var.project_name}-role"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  lifecycle {
    ignore_changes = [assume_role_policy]
  }
}

resource "aws_iam_policy" "test" {
  name        = "${var.project_name}-policy"
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
        Resource = "*"
      }
    ]
  })

  tags = {
    Name      = "${var.project_name}-policy"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  lifecycle {
    ignore_changes = [policy]
  }
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}

# ======================
# S3 Bucket
# ======================

resource "aws_s3_bucket" "test" {
  bucket = "${var.project_name}-bucket-${random_id.bucket_suffix.hex}"

  tags = {
    Name      = "${var.project_name}-bucket"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  lifecycle {
    ignore_changes = [tags]
  }
}

resource "random_id" "bucket_suffix" {
  byte_length = 4
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }

  lifecycle {
    ignore_changes = [versioning_configuration]
  }
}

resource "aws_s3_bucket_public_access_block" "test" {
  bucket = aws_s3_bucket.test.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true

  lifecycle {
    ignore_changes = [
      block_public_acls,
      block_public_policy,
      ignore_public_acls,
      restrict_public_buckets,
    ]
  }
}

# ======================
# Security Group
# ======================

resource "aws_security_group" "test" {
  name        = "${var.project_name}-sg"
  description = "Test security group for drift detection"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "HTTP"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound"
  }

  tags = {
    Name      = "${var.project_name}-sg"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  lifecycle {
    ignore_changes = [ingress, egress, tags]
  }
}

# ======================
# VPC (for Security Group)
# ======================

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name      = "${var.project_name}-vpc"
    Project   = var.project_name
    ManagedBy = "terraform"
  }
}

# ======================
# Outputs
# ======================

output "ec2_instance_id" {
  description = "EC2 Instance ID"
  value       = aws_instance.test.id
}

output "iam_role_name" {
  description = "IAM Role Name"
  value       = aws_iam_role.test.name
}

output "iam_role_arn" {
  description = "IAM Role ARN"
  value       = aws_iam_role.test.arn
}

output "iam_policy_arn" {
  description = "IAM Policy ARN"
  value       = aws_iam_policy.test.arn
}

output "s3_bucket_name" {
  description = "S3 Bucket Name"
  value       = aws_s3_bucket.test.id
}

output "security_group_id" {
  description = "Security Group ID"
  value       = aws_security_group.test.id
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.test.id
}

# ======================
# AWS WAF (Regional)
# ======================

# WAF IP Set - for IP-based restrictions
resource "aws_wafv2_ip_set" "test" {
  name               = "${var.project_name}-ip-set"
  description        = "Test IP set for drift detection"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"

  # Initial allowed IPs (we'll modify these via console for testing)
  addresses = [
    "203.0.113.0/24",  # Example IP range
  ]

  tags = {
    Name      = "${var.project_name}-ip-set"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  lifecycle {
    # Ignore changes to addresses so Terraform doesn't revert manual changes
    ignore_changes = [addresses, tags]
  }
}

# WAF Web ACL with IP-based rule
resource "aws_wafv2_web_acl" "test" {
  name        = "${var.project_name}-web-acl"
  description = "Test Web ACL for drift detection"
  scope       = "REGIONAL"

  default_action {
    block {}  # Block by default, only allow IPs in the IP set
  }

  # Rule to allow IPs from the IP set
  rule {
    name     = "AllowSpecificIPs"
    priority = 1

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AllowSpecificIPs"
      sampled_requests_enabled   = true
    }
  }

  # Rule to block SQL injection
  rule {
    name     = "BlockSQLInjection"
    priority = 2

    action {
      block {}
    }

    statement {
      sqli_match_statement {
        field_to_match {
          body {
            oversize_handling = "CONTINUE"
          }
        }

        text_transformation {
          priority = 0
          type     = "URL_DECODE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "BlockSQLInjection"
      sampled_requests_enabled   = true
    }
  }

  tags = {
    Name      = "${var.project_name}-web-acl"
    Project   = var.project_name
    ManagedBy = "terraform"
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "${var.project_name}-web-acl"
    sampled_requests_enabled   = true
  }

  lifecycle {
    # Ignore changes to rules and default action for testing
    ignore_changes = [rule, default_action, tags]
  }
}

# Optional: Associate WAF with ALB (if you want to test association changes)
# Uncomment if you create an ALB
# resource "aws_wafv2_web_acl_association" "test" {
#   resource_arn = aws_lb.test.arn
#   web_acl_arn  = aws_wafv2_web_acl.test.arn
# }

output "waf_ip_set_id" {
  description = "WAF IP Set ID"
  value       = aws_wafv2_ip_set.test.id
}

output "waf_ip_set_arn" {
  description = "WAF IP Set ARN"
  value       = aws_wafv2_ip_set.test.arn
}

output "waf_web_acl_id" {
  description = "WAF Web ACL ID"
  value       = aws_wafv2_web_acl.test.id
}

output "waf_web_acl_arn" {
  description = "WAF Web ACL ARN"
  value       = aws_wafv2_web_acl.test.arn
}
