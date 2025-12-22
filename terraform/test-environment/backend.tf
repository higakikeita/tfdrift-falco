# Terraform Backend Configuration
# S3 backend for storing state - TFDrift will monitor this state

terraform {
  backend "s3" {
    bucket = "tfdrift-test-state"  # Update with your bucket name
    key    = "test-environment/terraform.tfstate"
    region = "us-east-1"           # Update with your region

    # Optional: Enable state locking with DynamoDB
    # dynamodb_table = "terraform-state-lock"

    # Optional: Enable encryption
    # encrypt = true
    # kms_key_id = "arn:aws:kms:us-east-1:123456789012:key/..."
  }
}
