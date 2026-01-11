terraform {
  backend "s3" {
    bucket = "tfdrift-terraform-state-YOUR-AWS-ACCOUNT-ID"
    key    = "production-test/terraform.tfstate"
    region = "us-east-1"

    # Enable state locking
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}
