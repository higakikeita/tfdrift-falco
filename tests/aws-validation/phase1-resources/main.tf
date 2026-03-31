# TFDrift-Falco AWS Validation - Phase 1: Test Resources
# Creates VPC, EC2, S3 resources to validate drift detection
#
# Usage:
#   cd tests/aws-validation/phase1-resources
#   terraform init
#   terraform apply
#
# Prerequisites:
#   - AWS credentials configured (profile: draios-dev-developer)
#   - S3 bucket for TF state must exist (create manually or use -backend=false)

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6"
    }
  }

  backend "s3" {
    bucket  = "tfdrift-validation-state"
    key     = "phase1-resources/terraform.tfstate"
    region  = "ap-northeast-1"
    profile = "draios-dev-developer"
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = "tfdrift-falco-validation"
      Environment = "validation"
      ManagedBy   = "terraform"
      Owner       = "keita.higaki"
      Cleanup     = "safe-to-delete"
    }
  }
}

# ---------- Random suffix for unique naming ----------
resource "random_id" "suffix" {
  byte_length = 4
}

locals {
  name_prefix = "tfdrift-val-${random_id.suffix.hex}"
}

# ==========================================================
# VPC + Networking
# ==========================================================
resource "aws_vpc" "main" {
  cidr_block           = "10.99.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = { Name = "${local.name_prefix}-vpc" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${local.name_prefix}-igw" }
}

resource "aws_subnet" "public_a" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.99.1.0/24"
  availability_zone       = "${var.aws_region}a"
  map_public_ip_on_launch = true
  tags                    = { Name = "${local.name_prefix}-public-a" }
}

resource "aws_subnet" "public_c" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.99.2.0/24"
  availability_zone       = "${var.aws_region}c"
  map_public_ip_on_launch = true
  tags                    = { Name = "${local.name_prefix}-public-c" }
}

resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = { Name = "${local.name_prefix}-public-rt" }
}

resource "aws_route_table_association" "public_a" {
  subnet_id      = aws_subnet.public_a.id
  route_table_id = aws_route_table.public.id
}

resource "aws_route_table_association" "public_c" {
  subnet_id      = aws_subnet.public_c.id
  route_table_id = aws_route_table.public.id
}

# ==========================================================
# Security Group (drift detection target)
# ==========================================================
resource "aws_security_group" "webserver" {
  name        = "${local.name_prefix}-webserver-sg"
  description = "Webserver SG for drift validation"
  vpc_id      = aws_vpc.main.id

  # SSH - intentionally open for drift testing
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.my_ip_cidr]
  }

  # HTTP
  ingress {
    description = "HTTP"
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${local.name_prefix}-webserver-sg" }
}

# ==========================================================
# EC2 Instance (drift detection target)
# ==========================================================
data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-arm64"]
  }
  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}

resource "aws_instance" "webserver" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = "t4g.micro" # ARM, cheapest
  subnet_id              = aws_subnet.public_a.id
  vpc_security_group_ids = [aws_security_group.webserver.id]

  # Attributes that are easy to drift manually:
  monitoring                  = false
  disable_api_termination     = false
  instance_initiated_shutdown_behavior = "stop"

  root_block_device {
    volume_type = "gp3"
    volume_size = 8
    encrypted   = true
  }

  tags = {
    Name        = "${local.name_prefix}-webserver"
    Application = "tfdrift-validation"
  }
}

# ==========================================================
# S3 Bucket (drift detection target)
# ==========================================================
resource "aws_s3_bucket" "data" {
  bucket = "${local.name_prefix}-data"
  tags   = { Name = "${local.name_prefix}-data" }
}

resource "aws_s3_bucket_versioning" "data" {
  bucket = aws_s3_bucket.data.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "data" {
  bucket = aws_s3_bucket.data.id
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "data" {
  bucket                  = aws_s3_bucket.data.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ==========================================================
# IAM Role (drift detection target)
# ==========================================================
resource "aws_iam_role" "app" {
  name = "${local.name_prefix}-app-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = { Name = "${local.name_prefix}-app-role" }
}

resource "aws_iam_role_policy" "app_s3" {
  name = "${local.name_prefix}-app-s3-policy"
  role = aws_iam_role.app.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:PutObject", "s3:ListBucket"]
      Resource = [aws_s3_bucket.data.arn, "${aws_s3_bucket.data.arn}/*"]
    }]
  })
}
