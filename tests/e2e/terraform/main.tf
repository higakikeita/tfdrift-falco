# TFDrift-Falco E2E Test Infrastructure
# This creates test resources in AWS for drift detection testing

terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    # Configure via backend config file or environment variables
    # bucket = "tfdrift-test-state"
    # key    = "e2e/terraform.tfstate"
    # region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = "tfdrift-falco"
      Environment = "test"
      ManagedBy   = "Terraform"
      Purpose     = "E2E Testing"
      AutoDelete  = "true" # Signal for cleanup scripts
    }
  }
}

# ==========================================
# Test EC2 Instance
# ==========================================

resource "aws_instance" "test_instance" {
  ami           = data.aws_ami.amazon_linux_2.id
  instance_type = var.instance_type

  disable_api_termination = true # Will be toggled in drift tests

  vpc_security_group_ids = [aws_security_group.test_sg.id]
  subnet_id              = aws_subnet.test_subnet.id

  tags = {
    Name        = "tfdrift-test-instance"
    TestScenario = "drift-detection"
  }

  lifecycle {
    ignore_changes = [
      # These will be changed during tests
      disable_api_termination,
      instance_type,
    ]
  }
}

# ==========================================
# Test IAM Resources
# ==========================================

resource "aws_iam_role" "test_role" {
  name = "tfdrift-test-role"

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

  lifecycle {
    ignore_changes = [
      # Will be changed during tests
      assume_role_policy,
    ]
  }
}

resource "aws_iam_policy" "test_policy" {
  name        = "tfdrift-test-policy"
  description = "Test policy for drift detection"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "s3:GetObject",
        ]
        Effect   = "Allow"
        Resource = "${aws_s3_bucket.test_bucket.arn}/*"
      }
    ]
  })

  lifecycle {
    ignore_changes = [
      # Will be changed during tests
      policy,
    ]
  }
}

resource "aws_iam_role_policy_attachment" "test_attachment" {
  role       = aws_iam_role.test_role.name
  policy_arn = aws_iam_policy.test_policy.arn
}

# ==========================================
# Test S3 Bucket
# ==========================================

resource "aws_s3_bucket" "test_bucket" {
  bucket = "tfdrift-test-bucket-${var.test_id}"

  tags = {
    Name = "tfdrift-test-bucket"
  }

  lifecycle {
    ignore_changes = [
      # Versioning will be changed during tests
      versioning,
    ]
  }
}

resource "aws_s3_bucket_versioning" "test_bucket_versioning" {
  bucket = aws_s3_bucket.test_bucket.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test_bucket_encryption" {
  bucket = aws_s3_bucket.test_bucket.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "test_bucket_pab" {
  bucket = aws_s3_bucket.test_bucket.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ==========================================
# Networking for EC2 Instance
# ==========================================

resource "aws_vpc" "test_vpc" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "tfdrift-test-vpc"
  }
}

resource "aws_subnet" "test_subnet" {
  vpc_id                  = aws_vpc.test_vpc.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false

  tags = {
    Name = "tfdrift-test-subnet"
  }
}

resource "aws_security_group" "test_sg" {
  name        = "tfdrift-test-sg"
  description = "Security group for E2E testing"
  vpc_id      = aws_vpc.test_vpc.id

  # No ingress rules - this is a test instance

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "tfdrift-test-sg"
  }

  lifecycle {
    ignore_changes = [
      # Ingress rules will be changed during tests
      ingress,
    ]
  }
}

# ==========================================
# Data Sources
# ==========================================

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

data "aws_availability_zones" "available" {
  state = "available"
}

# ==========================================
# CloudTrail for E2E Tests (Optional)
# ==========================================

resource "aws_cloudtrail" "test_trail" {
  count = var.create_cloudtrail ? 1 : 0

  name                          = "tfdrift-test-trail"
  s3_bucket_name                = aws_s3_bucket.cloudtrail_bucket[0].id
  include_global_service_events = true
  is_multi_region_trail         = false
  enable_logging                = true

  event_selector {
    read_write_type           = "All"
    include_management_events = true
  }

  tags = {
    Name = "tfdrift-test-trail"
  }
}

resource "aws_s3_bucket" "cloudtrail_bucket" {
  count = var.create_cloudtrail ? 1 : 0

  bucket = "tfdrift-test-cloudtrail-${var.test_id}"

  tags = {
    Name = "tfdrift-test-cloudtrail"
  }
}

resource "aws_s3_bucket_policy" "cloudtrail_bucket_policy" {
  count = var.create_cloudtrail ? 1 : 0

  bucket = aws_s3_bucket.cloudtrail_bucket[0].id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AWSCloudTrailAclCheck"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:GetBucketAcl"
        Resource = aws_s3_bucket.cloudtrail_bucket[0].arn
      },
      {
        Sid    = "AWSCloudTrailWrite"
        Effect = "Allow"
        Principal = {
          Service = "cloudtrail.amazonaws.com"
        }
        Action   = "s3:PutObject"
        Resource = "${aws_s3_bucket.cloudtrail_bucket[0].arn}/*"
        Condition = {
          StringEquals = {
            "s3:x-amz-acl" = "bucket-owner-full-control"
          }
        }
      }
    ]
  })
}
