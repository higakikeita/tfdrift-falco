variable "aws_region" {
  description = "AWS region for validation resources"
  type        = string
  default     = "ap-northeast-1"
}

variable "aws_profile" {
  description = "AWS CLI profile to use"
  type        = string
  default     = "draios-dev-developer"
}

variable "my_ip_cidr" {
  description = "Your IP address in CIDR notation for SSH access (e.g., 1.2.3.4/32)"
  type        = string
  default     = "0.0.0.0/0" # Override with your actual IP
}
