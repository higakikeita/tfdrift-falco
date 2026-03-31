package falco

import "encoding/json"

// ChangeData represents a generic change event that can hold various types of structured data
type ChangeData struct {
	// Simple string fields
	StringValue string `json:"string_value,omitempty"`

	// For complex JSON structures, use RawMessage for flexibility
	RawJSON json.RawMessage `json:"raw_json,omitempty"`

	// Common fields across multiple change types
	Name      string `json:"name,omitempty"`
	ARN       string `json:"arn,omitempty"`
	Status    string `json:"status,omitempty"`
	Version   string `json:"version,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

// InstanceChange represents EC2 instance modifications
type InstanceChange struct {
	DisableAPITermination *bool  `json:"disable_api_termination,omitempty"`
	InstanceType          string `json:"instance_type,omitempty"`
}

// BucketEncryptionChange represents S3 bucket encryption changes
type BucketEncryptionChange struct {
	Config json.RawMessage `json:"config,omitempty"`
}

// LambdaFunctionChange represents Lambda function configuration changes
type LambdaFunctionChange struct {
	Timeout    *int32 `json:"timeout,omitempty"`
	MemorySize *int32 `json:"memory_size,omitempty"`
}

// IAMPolicyChange represents IAM policy modifications
type IAMPolicyChange struct {
	PolicyName     string          `json:"policy_name,omitempty"`
	PolicyDocument json.RawMessage `json:"policy_document,omitempty"`
	PolicyARN      string          `json:"policy_arn,omitempty"`
	AssumePolicy   json.RawMessage `json:"assume_role_policy,omitempty"`
}

// IAMRoleChange represents IAM role modifications
type IAMRoleChange struct {
	RoleName     string          `json:"role_name,omitempty"`
	AssumePolicy json.RawMessage `json:"assume_role_policy,omitempty"`
}

// IAMUserChange represents IAM user modifications
type IAMUserChange struct {
	UserName    string `json:"user_name,omitempty"`
	AccessKeyID string `json:"access_key_id,omitempty"`
}

// IAMGroupChange represents IAM group membership changes
type IAMGroupChange struct {
	UserName  string `json:"user_name,omitempty"`
	GroupName string `json:"group_name,omitempty"`
}

// PasswordPolicyChange represents account password policy changes
type PasswordPolicyChange struct {
	MinimumLength  *int32 `json:"minimum_password_length,omitempty"`
	RequireSymbols *bool  `json:"require_symbols,omitempty"`
}

// ECSServiceChange represents ECS service modifications
type ECSServiceChange struct {
	ServiceName      string `json:"service_name,omitempty"`
	Cluster          string `json:"cluster,omitempty"`
	TaskDefinition   string `json:"task_definition,omitempty"`
	DesiredCount     *int32 `json:"desired_count,omitempty"`
	LaunchType       string `json:"launch_type,omitempty"`
	ForceNewDeploy   *bool  `json:"force_new_deployment,omitempty"`
	EnableExecuteCmd *bool  `json:"enable_execute_command,omitempty"`
}

// ECSTaskDefinitionChange represents ECS task definition changes
type ECSTaskDefinitionChange struct {
	Family                  string          `json:"family,omitempty"`
	ContainerDefinitions    json.RawMessage `json:"container_definitions,omitempty"`
	TaskRoleARN             string          `json:"task_role_arn,omitempty"`
	ExecutionRoleARN        string          `json:"execution_role_arn,omitempty"`
	NetworkMode             string          `json:"network_mode,omitempty"`
	CPU                     string          `json:"cpu,omitempty"`
	Memory                  string          `json:"memory,omitempty"`
	RequiresCompatibilities string          `json:"requires_compatibilities,omitempty"`
}

// ECSClusterChange represents ECS cluster modifications
type ECSClusterChange struct {
	Settings                json.RawMessage `json:"settings,omitempty"`
	CapacityProviders       json.RawMessage `json:"capacity_providers,omitempty"`
	DefaultCapacityStrategy json.RawMessage `json:"default_capacity_provider_strategy,omitempty"`
}

// ECSCapacityProviderChange represents ECS capacity provider changes
type ECSCapacityProviderChange struct {
	Name                   string          `json:"name,omitempty"`
	AutoScalingGroupConfig json.RawMessage `json:"auto_scaling_group_provider,omitempty"`
}

// EKSClusterChange represents EKS cluster modifications
type EKSClusterChange struct {
	ClusterName        string          `json:"cluster_name,omitempty"`
	Version            string          `json:"version,omitempty"`
	RoleARN            string          `json:"role_arn,omitempty"`
	ResourcesVPCConfig json.RawMessage `json:"resources_vpc_config,omitempty"`
	Logging            json.RawMessage `json:"logging,omitempty"`
}

// EKSNodeGroupChange represents EKS node group modifications
type EKSNodeGroupChange struct {
	NodegroupName  string          `json:"nodegroup_name,omitempty"`
	ClusterName    string          `json:"cluster_name,omitempty"`
	NodeRoleARN    string          `json:"node_role_arn,omitempty"`
	Subnets        json.RawMessage `json:"subnets,omitempty"`
	ScalingConfig  json.RawMessage `json:"scaling_config,omitempty"`
	InstanceTypes  json.RawMessage `json:"instance_types,omitempty"`
	AMIType        string          `json:"ami_type,omitempty"`
	DiskSize       *int32          `json:"disk_size,omitempty"`
	Labels         json.RawMessage `json:"labels,omitempty"`
	Taints         json.RawMessage `json:"taints,omitempty"`
	ReleaseVersion string          `json:"release_version,omitempty"`
}

// EKSAddonChange represents EKS addon modifications
type EKSAddonChange struct {
	AddonName             string `json:"addon_name,omitempty"`
	ClusterName           string `json:"cluster_name,omitempty"`
	AddonVersion          string `json:"addon_version,omitempty"`
	ServiceAccountRoleARN string `json:"service_account_role_arn,omitempty"`
	ResolveConflicts      string `json:"resolve_conflicts,omitempty"`
}

// EKSFargateProfileChange represents EKS Fargate profile changes
type EKSFargateProfileChange struct {
	FargateProfileName  string          `json:"fargate_profile_name,omitempty"`
	ClusterName         string          `json:"cluster_name,omitempty"`
	PodExecutionRoleARN string          `json:"pod_execution_role_arn,omitempty"`
	Subnets             json.RawMessage `json:"subnets,omitempty"`
	Selectors           json.RawMessage `json:"selectors,omitempty"`
}

// ContainerInstanceStateChange represents ECS container instance state changes
type ContainerInstanceStateChange struct {
	Status string `json:"status,omitempty"`
}
