package mappings

// ComputeMappings contains CloudTrail event to Terraform resource mappings for compute services
var ComputeMappings = map[string]string{
	// EC2 - Instance Management
	"RunInstances":            "aws_instance",
	"TerminateInstances":      "aws_instance",
	"StartInstances":          "aws_instance",
	"StopInstances":           "aws_instance",
	"ModifyInstanceAttribute": "aws_instance",

	// EC2 - AMI Management
	"CreateImage":     "aws_ami",
	"DeregisterImage": "aws_ami",

	// EC2 - EBS Volume Management
	"CreateVolume":  "aws_ebs_volume",
	"DeleteVolume":  "aws_ebs_volume",
	"AttachVolume":  "aws_volume_attachment",
	"DetachVolume":  "aws_volume_attachment",
	"ModifyVolume":  "aws_ebs_volume",

	// EC2 - Snapshot Management
	"CreateSnapshot": "aws_ebs_snapshot",
	"DeleteSnapshot": "aws_ebs_snapshot",

	// EC2 - Network Interface Management
	"CreateNetworkInterface": "aws_network_interface",
	"DeleteNetworkInterface": "aws_network_interface",
	"AttachNetworkInterface": "aws_network_interface_attachment",

	// Lambda - Function Management
	"CreateFunction":                  "aws_lambda_function",
	"DeleteFunction":                  "aws_lambda_function",
	"UpdateFunctionConfiguration":     "aws_lambda_function",
	"UpdateFunctionCode":              "aws_lambda_function",
	"PublishVersion":                  "aws_lambda_version",
	"PutProvisionedConcurrencyConfig": "aws_lambda_provisioned_concurrency_config",
	"PutFunctionEventInvokeConfig":    "aws_lambda_function_event_invoke_config",
	"DeleteFunctionEventInvokeConfig": "aws_lambda_function_event_invoke_config",

	// Lambda - Event Source Mapping
	"CreateEventSourceMapping": "aws_lambda_event_source_mapping",
	"DeleteEventSourceMapping": "aws_lambda_event_source_mapping",
	"UpdateEventSourceMapping": "aws_lambda_event_source_mapping",

	// Lambda - Permissions
	"AddPermission":    "aws_lambda_permission",
	"RemovePermission": "aws_lambda_permission",

	// Lambda - Concurrency
	"PutFunctionConcurrency":    "aws_lambda_function",
	"DeleteFunctionConcurrency": "aws_lambda_function",

	// ECS - Services
	"CreateService": "aws_ecs_service",
	"UpdateService": "aws_ecs_service",
	"DeleteService": "aws_ecs_service",

	// ECS - Task Definitions
	"RegisterTaskDefinition":   "aws_ecs_task_definition",
	"DeregisterTaskDefinition": "aws_ecs_task_definition",

	// ECS - Clusters
	"UpdateCluster":                 "aws_ecs_cluster",
	"UpdateClusterSettings":         "aws_ecs_cluster",
	"PutClusterCapacityProviders":   "aws_ecs_cluster_capacity_providers",
	"UpdateContainerInstancesState": "aws_ecs_container_instance",

	// ECS - Capacity Providers
	"CreateCapacityProvider": "aws_ecs_capacity_provider",
	"UpdateCapacityProvider": "aws_ecs_capacity_provider",
	"DeleteCapacityProvider": "aws_ecs_capacity_provider",

	// EKS - Clusters
	"CreateCluster":        "aws_eks_cluster",
	"DeleteCluster":        "aws_eks_cluster",
	"UpdateClusterConfig":  "aws_eks_cluster",
	"UpdateClusterVersion": "aws_eks_cluster",

	// EKS - Node Groups
	"CreateNodegroup":        "aws_eks_node_group",
	"DeleteNodegroup":        "aws_eks_node_group",
	"UpdateNodegroupConfig":  "aws_eks_node_group",
	"UpdateNodegroupVersion": "aws_eks_node_group",

	// EKS - Addons
	"CreateAddon": "aws_eks_addon",
	"DeleteAddon": "aws_eks_addon",
	"UpdateAddon": "aws_eks_addon",

	// EKS - Fargate Profiles
	"CreateFargateProfile": "aws_eks_fargate_profile",

	// Auto Scaling - Groups
	"CreateAutoScalingGroup": "aws_autoscaling_group",
	"UpdateAutoScalingGroup": "aws_autoscaling_group",
	"DeleteAutoScalingGroup": "aws_autoscaling_group",
	"SetDesiredCapacity":     "aws_autoscaling_group",

	// Auto Scaling - Launch Configurations
	"CreateLaunchConfiguration": "aws_launch_configuration",
	"DeleteLaunchConfiguration": "aws_launch_configuration",

	// Auto Scaling - Policies
	"PutScalingPolicy": "aws_autoscaling_policy",
	"DeletePolicy":     "aws_autoscaling_policy",

	// Auto Scaling - Scheduled Actions
	"PutScheduledUpdateGroupAction": "aws_autoscaling_schedule",
	"DeleteScheduledAction":         "aws_autoscaling_schedule",

	// Auto Scaling - Load Balancers
	"AttachLoadBalancers":              "aws_autoscaling_attachment",
	"DetachLoadBalancers":              "aws_autoscaling_attachment",
	"AttachLoadBalancerTargetGroups":   "aws_autoscaling_attachment",
	"DetachLoadBalancerTargetGroups":   "aws_autoscaling_attachment",
}
