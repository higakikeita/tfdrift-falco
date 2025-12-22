# Outputs for Production-Like Test Environment

# ==================================================
# VPC Outputs
# ==================================================
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "CIDR block of the VPC"
  value       = module.vpc.vpc_cidr_block
}

output "public_subnets" {
  description = "List of public subnet IDs"
  value       = module.vpc.public_subnets
}

output "private_subnets" {
  description = "List of private subnet IDs"
  value       = module.vpc.private_subnets
}

output "database_subnets" {
  description = "List of database subnet IDs"
  value       = module.vpc.database_subnets
}

output "nat_gateway_ids" {
  description = "List of NAT Gateway IDs"
  value       = module.vpc.natgw_ids
}

# ==================================================
# Security Group Outputs
# ==================================================
output "alb_security_group_id" {
  description = "ID of ALB security group"
  value       = aws_security_group.alb.id
}

output "ecs_tasks_security_group_id" {
  description = "ID of ECS tasks security group"
  value       = aws_security_group.ecs_tasks.id
}

output "rds_security_group_id" {
  description = "ID of RDS security group"
  value       = aws_security_group.rds.id
}

output "elasticache_security_group_id" {
  description = "ID of ElastiCache security group"
  value       = aws_security_group.elasticache.id
}

# ==================================================
# S3 Outputs
# ==================================================
output "app_data_bucket_name" {
  description = "Name of application data bucket"
  value       = aws_s3_bucket.app_data.id
}

output "app_data_bucket_arn" {
  description = "ARN of application data bucket"
  value       = aws_s3_bucket.app_data.arn
}

output "backups_bucket_name" {
  description = "Name of backups bucket"
  value       = aws_s3_bucket.backups.id
}

output "logs_bucket_name" {
  description = "Name of logs bucket"
  value       = aws_s3_bucket.logs.id
}

# ==================================================
# ECS Outputs
# ==================================================
output "ecs_cluster_name" {
  description = "Name of ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "ecs_cluster_arn" {
  description = "ARN of ECS cluster"
  value       = aws_ecs_cluster.main.arn
}

output "ecs_cluster_id" {
  description = "ID of ECS cluster"
  value       = aws_ecs_cluster.main.id
}

# ==================================================
# ALB Outputs
# ==================================================
output "alb_dns_name" {
  description = "DNS name of ALB"
  value       = aws_lb.main.dns_name
}

output "alb_arn" {
  description = "ARN of ALB"
  value       = aws_lb.main.arn
}

output "alb_zone_id" {
  description = "Zone ID of ALB"
  value       = aws_lb.main.zone_id
}

output "alb_target_group_arn" {
  description = "ARN of ALB target group"
  value       = aws_lb_target_group.main.arn
}

# ==================================================
# EKS Outputs
# ==================================================
output "eks_cluster_name" {
  description = "Name of EKS cluster"
  value       = module.eks.cluster_name
}

output "eks_cluster_endpoint" {
  description = "Endpoint for EKS cluster"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_id" {
  description = "ID of EKS cluster"
  value       = module.eks.cluster_id
}

output "eks_cluster_arn" {
  description = "ARN of EKS cluster"
  value       = module.eks.cluster_arn
}

output "eks_cluster_certificate_authority_data" {
  description = "Certificate authority data for EKS cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_cluster_security_group_id" {
  description = "Security group ID of EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "eks_node_security_group_id" {
  description = "Security group ID of EKS nodes"
  value       = module.eks.node_security_group_id
}

output "eks_oidc_provider_arn" {
  description = "ARN of EKS OIDC provider"
  value       = module.eks.oidc_provider_arn
}

# ==================================================
# RDS Outputs
# ==================================================
output "rds_endpoint" {
  description = "Endpoint of RDS instance"
  value       = aws_db_instance.main.endpoint
}

output "rds_address" {
  description = "Address of RDS instance"
  value       = aws_db_instance.main.address
}

output "rds_port" {
  description = "Port of RDS instance"
  value       = aws_db_instance.main.port
}

output "rds_database_name" {
  description = "Name of RDS database"
  value       = aws_db_instance.main.db_name
}

output "rds_username" {
  description = "Username for RDS"
  value       = aws_db_instance.main.username
  sensitive   = true
}

output "rds_password_secret_arn" {
  description = "ARN of Secrets Manager secret for RDS password"
  value       = aws_secretsmanager_secret.rds_password.arn
}

# ==================================================
# ElastiCache Outputs
# ==================================================
output "elasticache_primary_endpoint" {
  description = "Primary endpoint of ElastiCache"
  value       = aws_elasticache_replication_group.main.primary_endpoint_address
}

output "elasticache_reader_endpoint" {
  description = "Reader endpoint of ElastiCache"
  value       = aws_elasticache_replication_group.main.reader_endpoint_address
}

output "elasticache_port" {
  description = "Port of ElastiCache"
  value       = aws_elasticache_replication_group.main.port
}

# ==================================================
# IAM Outputs
# ==================================================
output "ecs_task_execution_role_arn" {
  description = "ARN of ECS task execution role"
  value       = aws_iam_role.ecs_task_execution.arn
}

output "ecs_task_role_arn" {
  description = "ARN of ECS task role"
  value       = aws_iam_role.ecs_task.arn
}

# ==================================================
# Summary Output
# ==================================================
output "deployment_summary" {
  description = "Summary of deployed resources"
  value = {
    vpc = {
      id         = module.vpc.vpc_id
      cidr       = module.vpc.vpc_cidr_block
      nat_gws    = length(module.vpc.natgw_ids)
    }
    ecs = {
      cluster_name = aws_ecs_cluster.main.name
      alb_dns      = aws_lb.main.dns_name
    }
    eks = {
      cluster_name = module.eks.cluster_name
      endpoint     = module.eks.cluster_endpoint
    }
    rds = {
      endpoint     = aws_db_instance.main.endpoint
      multi_az     = aws_db_instance.main.multi_az
    }
    elasticache = {
      endpoint     = aws_elasticache_replication_group.main.primary_endpoint_address
      num_nodes    = var.elasticache_num_nodes
    }
    s3_buckets = {
      app_data = aws_s3_bucket.app_data.id
      backups  = aws_s3_bucket.backups.id
      logs     = aws_s3_bucket.logs.id
    }
  }
}

# ==================================================
# TFDrift Configuration Snippet
# ==================================================
output "tfdrift_config_snippet" {
  description = "Configuration snippet for TFDrift config.yaml"
  value = <<-EOT
  # Add this to your config.yaml:
  terraform:
    backend: s3
    s3:
      bucket: "tfdrift-prod-state-230446364776-20251221"
      key: "production-like-environment/terraform.tfstate"
      region: "${var.aws_region}"
    refresh_interval: 30s
  EOT
}
