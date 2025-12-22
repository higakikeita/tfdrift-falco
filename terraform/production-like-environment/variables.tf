# Variables for Production-Like Test Environment

# ==================================================
# General Configuration
# ==================================================
variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod-test"
}

variable "owner" {
  description = "Owner of the resources"
  type        = string
  default     = ""
}

# ==================================================
# Network Configuration
# ==================================================
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "az_count" {
  description = "Number of Availability Zones to use"
  type        = number
  default     = 3
  validation {
    condition     = var.az_count >= 2 && var.az_count <= 3
    error_message = "AZ count must be between 2 and 3."
  }
}

# ==================================================
# EKS Configuration
# ==================================================
variable "eks_cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.28"
}

variable "eks_node_instance_types" {
  description = "Instance types for EKS managed node group"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "eks_node_desired_size" {
  description = "Desired number of EKS nodes"
  type        = number
  default     = 2
}

variable "eks_node_min_size" {
  description = "Minimum number of EKS nodes"
  type        = number
  default     = 1
}

variable "eks_node_max_size" {
  description = "Maximum number of EKS nodes"
  type        = number
  default     = 4
}

# ==================================================
# RDS Configuration
# ==================================================
variable "rds_instance_class" {
  description = "Instance class for RDS"
  type        = string
  default     = "db.t3.micro"  # Use db.t3.medium or larger for production
}

variable "rds_allocated_storage" {
  description = "Allocated storage for RDS (GB)"
  type        = number
  default     = 20
}

# ==================================================
# ElastiCache Configuration
# ==================================================
variable "elasticache_node_type" {
  description = "Node type for ElastiCache"
  type        = string
  default     = "cache.t3.micro"  # Use cache.r6g.large or larger for production
}

variable "elasticache_num_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 2
  validation {
    condition     = var.elasticache_num_nodes >= 2
    error_message = "Must have at least 2 nodes for automatic failover."
  }
}
