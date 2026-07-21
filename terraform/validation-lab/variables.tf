variable "region" {
  description = "AWS region for the validation lab"
  type        = string
  default     = "ap-northeast-1"
}

variable "name_prefix" {
  description = "Prefix for all resource names"
  type        = string
  default     = "tfdrift-lab"
}

variable "vpc_cidr" {
  description = "CIDR block for the lab VPC"
  type        = string
  default     = "10.20.0.0/16"
}

variable "az_count" {
  description = "Number of Availability Zones to span (EKS needs >= 2)"
  type        = number
  default     = 2
}

# --- EKS ---
variable "eks_cluster_version" {
  description = "Kubernetes version for the EKS control plane"
  type        = string
  default     = "1.31"
}

variable "eks_node_instance_type" {
  description = "Instance type for the single managed node group (kept small for cost)"
  type        = string
  default     = "t3.small"
}

variable "eks_node_desired_size" {
  description = "Desired node count (minimal for a short-lived lab)"
  type        = number
  default     = 1
}

variable "eks_node_max_size" {
  description = "Max node count"
  type        = number
  default     = 2
}

# --- Bastion ---
variable "bastion_instance_type" {
  description = "Instance type for the SSM-managed bastion host"
  type        = string
  default     = "t3.micro"
}
