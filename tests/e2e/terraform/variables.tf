variable "aws_region" {
  description = "AWS region for test resources"
  type        = string
  default     = "us-east-1"
}

variable "test_id" {
  description = "Unique identifier for this test run (for resource naming)"
  type        = string
  default     = "default"

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.test_id))
    error_message = "test_id must contain only lowercase letters, numbers, and hyphens"
  }
}

variable "instance_type" {
  description = "EC2 instance type for test instance"
  type        = string
  default     = "t3.micro"
}

variable "create_cloudtrail" {
  description = "Whether to create CloudTrail (set to false if using existing trail)"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Additional tags for all resources"
  type        = map(string)
  default     = {}
}
