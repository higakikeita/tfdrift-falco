# Variables for TFDrift-Falco Test Environment

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name (e.g., test, dev, staging)"
  type        = string
  default     = "test"
}

variable "test_bucket_name" {
  description = "Name of the S3 bucket for testing (must be globally unique)"
  type        = string
  # Example: "tfdrift-test-bucket-20231215"
}

variable "vpc_id" {
  description = "VPC ID where security group will be created"
  type        = string
  # Use your default VPC or create a new one
}

variable "alert_email" {
  description = "Email address for drift alert notifications (leave empty to skip)"
  type        = string
  default     = ""
}
