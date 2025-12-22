# Outputs for TFDrift-Falco Test Environment

output "s3_bucket_name" {
  description = "Name of the test S3 bucket"
  value       = aws_s3_bucket.test.id
}

output "s3_bucket_arn" {
  description = "ARN of the test S3 bucket"
  value       = aws_s3_bucket.test.arn
}

output "security_group_id" {
  description = "ID of the test security group"
  value       = aws_security_group.test.id
}

output "security_group_name" {
  description = "Name of the test security group"
  value       = aws_security_group.test.name
}

output "iam_policy_arn" {
  description = "ARN of the test IAM policy"
  value       = aws_iam_policy.test.arn
}

output "iam_role_name" {
  description = "Name of the test IAM role"
  value       = aws_iam_role.test.name
}

output "iam_role_arn" {
  description = "ARN of the test IAM role"
  value       = aws_iam_role.test.arn
}

output "cloudwatch_log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = aws_cloudwatch_log_group.test.name
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for drift alerts"
  value       = aws_sns_topic.drift_alerts.arn
}

output "terraform_state_location" {
  description = "Location of Terraform state (for TFDrift config)"
  value       = "s3://${terraform.backend.config.bucket}/${terraform.backend.config.key}"
}

output "resources_summary" {
  description = "Summary of created resources for easy reference"
  value = {
    s3_bucket       = aws_s3_bucket.test.id
    security_group  = aws_security_group.test.id
    iam_policy      = aws_iam_policy.test.arn
    iam_role        = aws_iam_role.test.name
    log_group       = aws_cloudwatch_log_group.test.name
    sns_topic       = aws_sns_topic.drift_alerts.arn
  }
}
