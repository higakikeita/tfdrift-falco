terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.region

  default_tags {
    tags = {
      Project     = "tfdrift-falco"
      Environment = "validation-lab"
      ManagedBy   = "terraform"
      Purpose     = "drift-detection-e2e"
    }
  }
}
