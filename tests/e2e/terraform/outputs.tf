output "test_instance_id" {
  description = "ID of the test EC2 instance"
  value       = aws_instance.test_instance.id
}

output "test_instance_arn" {
  description = "ARN of the test EC2 instance"
  value       = aws_instance.test_instance.arn
}

output "test_role_name" {
  description = "Name of the test IAM role"
  value       = aws_iam_role.test_role.name
}

output "test_role_arn" {
  description = "ARN of the test IAM role"
  value       = aws_iam_role.test_role.arn
}

output "test_policy_name" {
  description = "Name of the test IAM policy"
  value       = aws_iam_policy.test_policy.name
}

output "test_policy_arn" {
  description = "ARN of the test IAM policy"
  value       = aws_iam_policy.test_policy.arn
}

output "test_bucket_name" {
  description = "Name of the test S3 bucket"
  value       = aws_s3_bucket.test_bucket.id
}

output "test_bucket_arn" {
  description = "ARN of the test S3 bucket"
  value       = aws_s3_bucket.test_bucket.arn
}

output "test_vpc_id" {
  description = "ID of the test VPC"
  value       = aws_vpc.test_vpc.id
}

output "test_subnet_id" {
  description = "ID of the test subnet"
  value       = aws_subnet.test_subnet.id
}

output "test_security_group_id" {
  description = "ID of the test security group"
  value       = aws_security_group.test_sg.id
}

output "cloudtrail_bucket_name" {
  description = "Name of the CloudTrail S3 bucket (if created)"
  value       = var.create_cloudtrail ? aws_s3_bucket.cloudtrail_bucket[0].id : null
}

output "cloudtrail_name" {
  description = "Name of the CloudTrail (if created)"
  value       = var.create_cloudtrail ? aws_cloudtrail.test_trail[0].name : null
}

# Output for test configuration
output "test_config" {
  description = "Configuration to use in E2E tests"
  value = {
    aws_region      = var.aws_region
    instance_id     = aws_instance.test_instance.id
    role_name       = aws_iam_role.test_role.name
    policy_name     = aws_iam_policy.test_policy.name
    bucket_name     = aws_s3_bucket.test_bucket.id
    security_group_id = aws_security_group.test_sg.id
  }
}
