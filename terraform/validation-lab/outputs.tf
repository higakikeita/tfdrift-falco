output "region" {
  description = "AWS region"
  value       = var.region
}

output "vpc_id" {
  value       = module.vpc.vpc_id
  description = "Lab VPC id"
}

output "bastion_instance_id" {
  description = "Bastion instance id — connect with: aws ssm start-session --target <id>"
  value       = aws_instance.bastion.id
}

output "bastion_ssm_command" {
  description = "Ready-to-run SSM session command for the bastion"
  value       = "aws ssm start-session --target ${aws_instance.bastion.id} --region ${var.region}"
}

output "eks_cluster_name" {
  value       = module.eks.cluster_name
  description = "EKS cluster name"
}

output "eks_kubeconfig_command" {
  description = "Command to configure kubectl for the EKS cluster"
  value       = "aws eks update-kubeconfig --name ${module.eks.cluster_name} --region ${var.region}"
}

output "ecs_cluster_name" {
  value       = aws_ecs_cluster.lab.name
  description = "ECS Fargate cluster name"
}

output "ecs_service_name" {
  value       = aws_ecs_service.nginx.name
  description = "ECS Fargate service running the nginx workload"
}
