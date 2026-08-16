# Terraform Backend Configuration
# S3 backend for storing state - driftwire will monitor this state

terraform {
  backend "s3" {
    bucket = "driftwire-prod-state-230446364776-20251221"
    key    = "production-like-environment/terraform.tfstate"
    region = "us-east-1"
    encrypt = true
  }
}
