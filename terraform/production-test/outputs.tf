output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "subnet_id" {
  description = "ID of the public subnet"
  value       = aws_subnet.public.id
}

output "security_group_id" {
  description = "ID of the web security group"
  value       = aws_security_group.web.id
}

output "iam_role_arn" {
  description = "ARN of the EC2 IAM role"
  value       = aws_iam_role.ec2_role.arn
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.app_data.id
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = aws_s3_bucket.app_data.arn
}

# output "instance_id" {
#   description = "ID of the EC2 instance"
#   value       = aws_instance.web.id
# }
#
# output "instance_public_ip" {
#   description = "Public IP of the EC2 instance"
#   value       = aws_instance.web.public_ip
# }
