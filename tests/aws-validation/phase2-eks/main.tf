# TFDrift-Falco AWS Validation - Phase 2: EKS Cluster
# Creates a minimal EKS cluster for Falco + tfdrift-falco deployment
#
# Usage:
#   cd tests/aws-validation/phase2-eks
#   terraform init
#   terraform apply -var="vpc_id=vpc-xxx" -var='subnet_ids=["subnet-aaa","subnet-bbb"]'
#
# Or use Phase 1 outputs:
#   terraform apply \
#     -var="vpc_id=$(cd ../phase1-resources && terraform output -raw vpc_id)" \
#     -var="subnet_ids=$(cd ../phase1-resources && terraform output -json subnet_ids)"

terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket  = "tfdrift-validation-state"
    key     = "phase2-eks/terraform.tfstate"
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

# ==========================================================
# EKS Cluster
# ==========================================================
resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  role_arn = aws_iam_role.eks_cluster.arn
  version  = var.k8s_version

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_public_access  = true
    endpoint_private_access = true
    security_group_ids      = [aws_security_group.eks_cluster.id]
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_vpc_resource_controller,
  ]

  tags = { Name = var.cluster_name }
}

# ==========================================================
# EKS Managed Node Group (minimal: 1 node)
# ==========================================================
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.cluster_name}-ng"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = var.subnet_ids

  instance_types = ["t4g.medium"] # ARM, cost-effective
  capacity_type  = "SPOT"         # Use spot for validation (cheap)
  ami_type       = "AL2023_ARM_64_STANDARD"

  scaling_config {
    desired_size = 1
    max_size     = 2
    min_size     = 1
  }

  update_config {
    max_unavailable = 1
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_ecr_policy,
  ]

  tags = { Name = "${var.cluster_name}-ng" }
}

# ==========================================================
# IAM: EKS Cluster Role
# ==========================================================
resource "aws_iam_role" "eks_cluster" {
  name = "${var.cluster_name}-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "eks.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster.name
}

resource "aws_iam_role_policy_attachment" "eks_vpc_resource_controller" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSVPCResourceController"
  role       = aws_iam_role.eks_cluster.name
}

# ==========================================================
# IAM: EKS Node Role
# ==========================================================
resource "aws_iam_role" "eks_node" {
  name = "${var.cluster_name}-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "eks_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_node.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_node.name
}

resource "aws_iam_role_policy_attachment" "eks_ecr_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_node.name
}

# CloudTrail S3 read access for Falco on nodes
resource "aws_iam_role_policy" "eks_node_cloudtrail" {
  name = "${var.cluster_name}-node-cloudtrail"
  role = aws_iam_role.eks_node.id

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
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = var.cloudtrail_sqs_arn != "" ? [var.cloudtrail_sqs_arn] : ["*"]
      }
    ]
  })
}

# TF state read access for tfdrift-falco on nodes
resource "aws_iam_role_policy" "eks_node_tfstate" {
  name = "${var.cluster_name}-node-tfstate"
  role = aws_iam_role.eks_node.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject", "s3:ListBucket"]
      Resource = [
        "arn:aws:s3:::tfdrift-validation-state",
        "arn:aws:s3:::tfdrift-validation-state/*"
      ]
    }]
  })
}

# ==========================================================
# Security Group for EKS Cluster
# ==========================================================
resource "aws_security_group" "eks_cluster" {
  name        = "${var.cluster_name}-cluster-sg"
  description = "EKS cluster security group"
  vpc_id      = var.vpc_id

  ingress {
    description = "Allow all from VPC"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["10.99.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.cluster_name}-cluster-sg" }
}
