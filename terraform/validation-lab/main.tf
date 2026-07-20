# TFDrift-Falco validation lab: bastion (SSM) + EKS + ECS Fargate.
# Deliberately lean and short-lived — apply, generate real CloudTrail drift,
# verify tfdrift detection, then `terraform destroy`. Cost drivers are the EKS
# control plane (~$0.10/h), one NAT gateway, one small node, and Fargate tasks.

data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  name = var.name_prefix
  azs  = slice(data.aws_availability_zones.available.names, 0, var.az_count)
}

# ---------------------------------------------------------------------------
# Networking — 2 AZ, single NAT gateway (cost-optimized for a short-lived lab)
# ---------------------------------------------------------------------------
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${local.name}-vpc"
  cidr = var.vpc_cidr
  azs  = local.azs

  public_subnets  = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k)]
  private_subnets = [for k, v in local.azs : cidrsubnet(var.vpc_cidr, 8, k + 10)]

  enable_nat_gateway   = true
  single_nat_gateway   = true # one NAT for the whole VPC — cheapest workable option
  enable_dns_hostnames = true
  enable_dns_support   = true

  # Tags required for EKS subnet auto-discovery
  public_subnet_tags  = { "kubernetes.io/role/elb" = "1" }
  private_subnet_tags = { "kubernetes.io/role/internal-elb" = "1" }
}

# ---------------------------------------------------------------------------
# Bastion — SSM Session Manager only (no inbound SSH, no key pair, no open :22)
# ---------------------------------------------------------------------------
data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }
}

resource "aws_iam_role" "bastion" {
  name = "${local.name}-bastion"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "bastion_ssm" {
  role       = aws_iam_role.bastion.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_instance_profile" "bastion" {
  name = "${local.name}-bastion"
  role = aws_iam_role.bastion.name
}

# Egress-only SG: the bastion reaches AWS APIs / clusters outbound; SSM handles
# the inbound session over the AWS backbone, so no ingress rule is needed.
resource "aws_security_group" "bastion" {
  name        = "${local.name}-bastion"
  description = "Bastion (SSM) - egress only"
  vpc_id      = module.vpc.vpc_id

  egress {
    description = "All outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${local.name}-bastion" }
}

resource "aws_instance" "bastion" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = var.bastion_instance_type
  subnet_id              = module.vpc.private_subnets[0] # private: reachable via SSM only
  iam_instance_profile   = aws_iam_instance_profile.bastion.name
  vpc_security_group_ids = [aws_security_group.bastion.id]

  metadata_options {
    http_tokens = "required" # IMDSv2
  }

  tags = { Name = "${local.name}-bastion" }
}

# ---------------------------------------------------------------------------
# EKS — single small managed node group
# ---------------------------------------------------------------------------
module "eks" {
  source  = "terraform-aws-modules/eks/aws"
  version = "~> 20.0"

  cluster_name    = "${local.name}-eks"
  cluster_version = var.eks_cluster_version

  cluster_endpoint_public_access           = true
  enable_cluster_creator_admin_permissions = true

  vpc_id     = module.vpc.vpc_id
  subnet_ids = module.vpc.private_subnets

  eks_managed_node_groups = {
    default = {
      instance_types = [var.eks_node_instance_type]
      min_size       = 1
      max_size       = var.eks_node_max_size
      desired_size   = var.eks_node_desired_size
    }
  }
}

# ---------------------------------------------------------------------------
# ECS Fargate — cluster + one minimal nginx service to have a live workload
# whose manual changes generate CloudTrail drift events.
# ---------------------------------------------------------------------------
resource "aws_ecs_cluster" "lab" {
  name = "${local.name}-ecs"

  setting {
    name  = "containerInsights"
    value = "disabled" # keep cost down
  }
}

resource "aws_ecs_cluster_capacity_providers" "lab" {
  cluster_name       = aws_ecs_cluster.lab.name
  capacity_providers = ["FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}

resource "aws_iam_role" "ecs_execution" {
  name = "${local.name}-ecs-exec"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "ecs-tasks.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/${local.name}"
  retention_in_days = 1
}

resource "aws_security_group" "ecs_service" {
  name        = "${local.name}-ecs-svc"
  description = "ECS Fargate service - egress only"
  vpc_id      = module.vpc.vpc_id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${local.name}-ecs-svc" }
}

resource "aws_ecs_task_definition" "nginx" {
  family                   = "${local.name}-nginx"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn

  container_definitions = jsonencode([{
    name         = "nginx"
    image        = "public.ecr.aws/nginx/nginx:stable"
    essential    = true
    portMappings = [{ containerPort = 80, protocol = "tcp" }]
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "nginx"
      }
    }
  }])
}

resource "aws_ecs_service" "nginx" {
  name            = "${local.name}-nginx"
  cluster         = aws_ecs_cluster.lab.id
  task_definition = aws_ecs_task_definition.nginx.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = module.vpc.private_subnets
    security_groups  = [aws_security_group.ecs_service.id]
    assign_public_ip = false
  }
}
