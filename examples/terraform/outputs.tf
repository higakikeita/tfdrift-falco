output "vpc_id" {
  description = "ID of the test VPC"
  value       = aws_vpc.test.id
}

output "ec2_instance_id" {
  description = "ID of the test EC2 instance"
  value       = aws_instance.webserver.id
}

output "ec2_instance_public_ip" {
  description = "Public IP of the test EC2 instance"
  value       = aws_instance.webserver.public_ip
}

output "s3_bucket_name" {
  description = "Name of the test S3 bucket"
  value       = aws_s3_bucket.data.id
}

output "s3_bucket_arn" {
  description = "ARN of the test S3 bucket"
  value       = aws_s3_bucket.data.arn
}

output "iam_role_name" {
  description = "Name of the test IAM role"
  value       = aws_iam_role.app.name
}

output "iam_role_arn" {
  description = "ARN of the test IAM role"
  value       = aws_iam_role.app.arn
}

output "security_group_id" {
  description = "ID of the webserver security group"
  value       = aws_security_group.webserver.id
}
