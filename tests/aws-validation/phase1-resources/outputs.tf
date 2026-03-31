output "vpc_id" {
  description = "VPC ID for Phase 2 EKS"
  value       = aws_vpc.main.id
}

output "subnet_ids" {
  description = "Public subnet IDs for EKS"
  value       = [aws_subnet.public_a.id, aws_subnet.public_c.id]
}

output "instance_id" {
  description = "EC2 instance ID (drift target)"
  value       = aws_instance.webserver.id
}

output "instance_public_ip" {
  description = "EC2 public IP"
  value       = aws_instance.webserver.public_ip
}

output "security_group_id" {
  description = "Security group ID (drift target)"
  value       = aws_security_group.webserver.id
}

output "s3_bucket_name" {
  description = "S3 bucket name (drift target)"
  value       = aws_s3_bucket.data.id
}

output "iam_role_arn" {
  description = "IAM role ARN (drift target)"
  value       = aws_iam_role.app.arn
}

output "name_prefix" {
  description = "Resource name prefix for reference"
  value       = local.name_prefix
}
