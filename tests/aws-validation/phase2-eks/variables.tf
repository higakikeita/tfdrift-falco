variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-northeast-1"
}

variable "aws_profile" {
  description = "AWS CLI profile"
  type        = string
  default     = "draios-dev-developer"
}

variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "tfdrift-val-eks"
}

variable "k8s_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.29"
}

variable "vpc_id" {
  description = "VPC ID from Phase 1"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs from Phase 1"
  type        = list(string)
}

variable "cloudtrail_bucket" {
  description = "CloudTrail S3 bucket name (org-level)"
  type        = string
  default     = "" # Will be discovered during setup
}

variable "cloudtrail_sqs_arn" {
  description = "SQS queue ARN for CloudTrail notifications (optional)"
  type        = string
  default     = ""
}
