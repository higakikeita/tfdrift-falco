# TFDrift-Falco Terraform Module Variables

variable "cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
  default     = "tfdrift-falco"
}

variable "aws_region" {
  description = "AWS region to deploy TFDrift-Falco"
  type        = string
  default     = "us-east-1"
}

variable "vpc_id" {
  description = "VPC ID where TFDrift-Falco will be deployed"
  type        = string
}

variable "subnet_ids" {
  description = "List of subnet IDs for ECS tasks"
  type        = list(string)
}

variable "assign_public_ip" {
  description = "Assign public IP to ECS tasks (required if using public subnets)"
  type        = bool
  default     = false
}

# Terraform State Configuration
variable "terraform_state_bucket" {
  description = "S3 bucket containing Terraform state"
  type        = string
}

variable "terraform_state_key" {
  description = "S3 key for Terraform state file"
  type        = string
  default     = "terraform.tfstate"
}

variable "terraform_state_kms_key_arn" {
  description = "KMS key ARN for encrypted Terraform state (optional)"
  type        = string
  default     = ""
}

# CloudTrail Configuration
variable "cloudtrail_bucket" {
  description = "S3 bucket containing CloudTrail logs"
  type        = string
}

variable "cloudtrail_sqs_queue_arn" {
  description = "SQS queue ARN for CloudTrail notifications (optional)"
  type        = string
  default     = ""
}

# ECS Task Configuration
variable "task_cpu" {
  description = "CPU units for ECS task (256, 512, 1024, 2048, 4096)"
  type        = string
  default     = "512"
}

variable "task_memory" {
  description = "Memory for ECS task in MB (512, 1024, 2048, 4096, 8192)"
  type        = string
  default     = "1024"
}

variable "desired_count" {
  description = "Number of ECS tasks to run"
  type        = number
  default     = 1
}

# Container Versions
variable "falco_version" {
  description = "Falco container image version"
  type        = string
  default     = "0.37.1"
}

variable "tfdrift_version" {
  description = "TFDrift-Falco container image version"
  type        = string
  default     = "latest"
}

# Notification Configuration
variable "slack_webhook_secret_arn" {
  description = "AWS Secrets Manager ARN containing Slack webhook URL (optional)"
  type        = string
  default     = ""
}

# Monitoring Configuration
variable "enable_cloudwatch_alarms" {
  description = "Enable CloudWatch alarms for CPU and memory"
  type        = bool
  default     = true
}

variable "alarm_sns_topic_arn" {
  description = "SNS topic ARN for CloudWatch alarm notifications (optional)"
  type        = string
  default     = ""
}

variable "log_retention_days" {
  description = "CloudWatch Logs retention in days"
  type        = number
  default     = 30
}

variable "metrics_allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access metrics endpoint"
  type        = list(string)
  default     = ["10.0.0.0/8"]
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Terraform   = "true"
    Application = "TFDrift-Falco"
  }
}
