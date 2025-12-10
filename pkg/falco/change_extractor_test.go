package falco

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractChanges(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name      string
		eventName string
		fields    map[string]string
		wantKeys  []string
	}{
		{
			name:      "ModifyInstanceAttribute - Instance Type",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.instancetype": "t3.medium",
			},
			wantKeys: []string{"instance_type"},
		},
		{
			name:      "ModifyInstanceAttribute - API Termination",
			eventName: "ModifyInstanceAttribute",
			fields: map[string]string{
				"ct.request.disableapitermination": "true",
			},
			wantKeys: []string{"disable_api_termination"},
		},
		{
			name:      "PutBucketEncryption",
			eventName: "PutBucketEncryption",
			fields: map[string]string{
				"ct.request.serversideencryptionconfiguration": "AES256",
			},
			wantKeys: []string{"server_side_encryption_configuration"},
		},
		{
			name:      "DeleteBucketEncryption",
			eventName: "DeleteBucketEncryption",
			fields:    map[string]string{},
			wantKeys:  []string{"server_side_encryption_configuration"},
		},
		{
			name:      "UpdateFunctionConfiguration - Timeout and Memory",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.timeout":    "60",
				"ct.request.memorysize": "512",
			},
			wantKeys: []string{"timeout", "memory_size"},
		},
		{
			name:      "AttachRolePolicy",
			eventName: "AttachRolePolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/ReadOnlyAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "PutRolePolicy",
			eventName: "PutRolePolicy",
			fields: map[string]string{
				"ct.request.policyname":     "inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "CreateRole",
			eventName: "CreateRole",
			fields: map[string]string{
				"ct.request.rolename":                 "my-role",
				"ct.request.assumerolepolicydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"role_name", "assume_role_policy"},
		},
		{
			name:      "DeleteUser",
			eventName: "DeleteUser",
			fields: map[string]string{
				"ct.request.username": "old-user",
			},
			wantKeys: []string{"deleted_user"},
		},
		{
			name:      "CreateAccessKey",
			eventName: "CreateAccessKey",
			fields: map[string]string{
				"ct.request.username":               "service-user",
				"ct.response.accesskey.accesskeyid": "AKIAIOSFODNN7EXAMPLE",
			},
			wantKeys: []string{"user_name", "access_key_id"},
		},
		{
			name:      "AddUserToGroup",
			eventName: "AddUserToGroup",
			fields: map[string]string{
				"ct.request.username":  "john",
				"ct.request.groupname": "developers",
			},
			wantKeys: []string{"user_name", "group_name"},
		},
		{
			name:      "UpdateAccountPasswordPolicy",
			eventName: "UpdateAccountPasswordPolicy",
			fields: map[string]string{
				"ct.request.minimumpasswordlength": "14",
				"ct.request.requiresymbols":        "true",
			},
			wantKeys: []string{"minimum_password_length", "require_symbols"},
		},
		{
			name:      "DeleteRole",
			eventName: "DeleteRole",
			fields: map[string]string{
				"ct.request.rolename": "obsolete-role",
			},
			wantKeys: []string{"deleted_role"},
		},
		{
			name:      "CreateUser",
			eventName: "CreateUser",
			fields: map[string]string{
				"ct.request.username": "new-user",
			},
			wantKeys: []string{"user_name"},
		},
		{
			name:      "RemoveUserFromGroup",
			eventName: "RemoveUserFromGroup",
			fields: map[string]string{
				"ct.request.username":  "jane",
				"ct.request.groupname": "admins",
			},
			wantKeys: []string{"user_name", "group_name"},
		},
		{
			name:      "CreatePolicy",
			eventName: "CreatePolicy",
			fields: map[string]string{
				"ct.request.policyname":     "new-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17","Statement":[]}`,
			},
			wantKeys: []string{"policy_name", "policy_document"},
		},
		{
			name:      "CreatePolicyVersion",
			eventName: "CreatePolicyVersion",
			fields: map[string]string{
				"ct.request.policyarn":      "arn:aws:iam::123:policy/my-policy",
				"ct.request.setasdefault":   "true",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"policy_arn", "set_as_default", "policy_document"},
		},
		{
			name:      "AttachUserPolicy",
			eventName: "AttachUserPolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/PowerUserAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "AttachGroupPolicy",
			eventName: "AttachGroupPolicy",
			fields: map[string]string{
				"ct.request.policyarn": "arn:aws:iam::aws:policy/ReadOnlyAccess",
			},
			wantKeys: []string{"attached_policy_arn"},
		},
		{
			name:      "PutUserPolicy",
			eventName: "PutUserPolicy",
			fields: map[string]string{
				"ct.request.policyname":     "user-inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "PutGroupPolicy",
			eventName: "PutGroupPolicy",
			fields: map[string]string{
				"ct.request.policyname":     "group-inline-policy",
				"ct.request.policydocument": `{"Version":"2012-10-17"}`,
			},
			wantKeys: []string{"inline_policy_name", "policy_document"},
		},
		{
			name:      "UpdateFunctionConfiguration - Only Timeout",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.timeout": "30",
			},
			wantKeys: []string{"timeout"},
		},
		{
			name:      "UpdateFunctionConfiguration - Only Memory",
			eventName: "UpdateFunctionConfiguration",
			fields: map[string]string{
				"ct.request.memorysize": "1024",
			},
			wantKeys: []string{"memory_size"},
		},
		{
			name:      "Unknown Event",
			eventName: "UnknownEvent",
			fields: map[string]string{
				"some.field": "value",
			},
			wantKeys: []string{},
		},
		// ECS - Services
		{
			name:      "CreateService",
			eventName: "CreateService",
			fields: map[string]string{
				"ct.request.servicename":   "my-service",
				"ct.request.cluster":       "my-cluster",
				"ct.request.taskdefinition": "my-task:1",
				"ct.request.desiredcount":  "2",
				"ct.request.launchtype":    "FARGATE",
			},
			wantKeys: []string{"service_name", "cluster", "task_definition", "desired_count", "launch_type"},
		},
		{
			name:      "UpdateService",
			eventName: "UpdateService",
			fields: map[string]string{
				"ct.request.desiredcount":        "3",
				"ct.request.taskdefinition":      "my-task:2",
				"ct.request.forcenewdeployment":  "true",
				"ct.request.enableexecutecommand": "true",
			},
			wantKeys: []string{"desired_count", "task_definition", "force_new_deployment", "enable_execute_command"},
		},
		{
			name:      "DeleteService",
			eventName: "DeleteService",
			fields: map[string]string{
				"ct.request.service": "my-service",
				"ct.request.force":   "true",
			},
			wantKeys: []string{"deleted_service", "force"},
		},
		// ECS - Task Definitions
		{
			name:      "RegisterTaskDefinition",
			eventName: "RegisterTaskDefinition",
			fields: map[string]string{
				"ct.request.family":                  "my-task",
				"ct.request.containerdefinitions":    `[{"name":"app","image":"nginx:latest"}]`,
				"ct.request.taskrolearn":             "arn:aws:iam::123456789012:role/task-role",
				"ct.request.executionrolearn":        "arn:aws:iam::123456789012:role/execution-role",
				"ct.request.networkmode":             "awsvpc",
				"ct.request.cpu":                     "256",
				"ct.request.memory":                  "512",
				"ct.request.requirescompatibilities": "FARGATE",
			},
			wantKeys: []string{"family", "container_definitions", "task_role_arn", "execution_role_arn", "network_mode", "cpu", "memory", "requires_compatibilities"},
		},
		{
			name:      "DeregisterTaskDefinition",
			eventName: "DeregisterTaskDefinition",
			fields: map[string]string{
				"ct.request.taskdefinition": "my-task:1",
			},
			wantKeys: []string{"deregistered_task_definition"},
		},
		// ECS - Clusters
		{
			name:      "UpdateCluster",
			eventName: "UpdateCluster",
			fields: map[string]string{
				"ct.request.settings": `[{"name":"containerInsights","value":"enabled"}]`,
			},
			wantKeys: []string{"settings"},
		},
		{
			name:      "UpdateClusterSettings",
			eventName: "UpdateClusterSettings",
			fields: map[string]string{
				"ct.request.settings": `[{"name":"containerInsights","value":"disabled"}]`,
			},
			wantKeys: []string{"settings"},
		},
		{
			name:      "PutClusterCapacityProviders",
			eventName: "PutClusterCapacityProviders",
			fields: map[string]string{
				"ct.request.capacityproviders":             `["FARGATE","FARGATE_SPOT"]`,
				"ct.request.defaultcapacityproviderstrategy": `[{"capacityProvider":"FARGATE","weight":1}]`,
			},
			wantKeys: []string{"capacity_providers", "default_capacity_provider_strategy"},
		},
		// ECS - Container Instances
		{
			name:      "UpdateContainerInstancesState",
			eventName: "UpdateContainerInstancesState",
			fields: map[string]string{
				"ct.request.status": "DRAINING",
			},
			wantKeys: []string{"status"},
		},
		// ECS - Capacity Providers
		{
			name:      "CreateCapacityProvider",
			eventName: "CreateCapacityProvider",
			fields: map[string]string{
				"ct.request.name":                    "my-provider",
				"ct.request.autoscalinggroupprovider": `{"autoScalingGroupArn":"arn:aws:autoscaling:us-east-1:123456789012:autoScalingGroup:abc123"}`,
			},
			wantKeys: []string{"name", "auto_scaling_group_provider"},
		},
		{
			name:      "UpdateCapacityProvider",
			eventName: "UpdateCapacityProvider",
			fields: map[string]string{
				"ct.request.autoscalinggroupprovider": `{"managedScaling":{"status":"ENABLED"}}`,
			},
			wantKeys: []string{"auto_scaling_group_provider"},
		},
		{
			name:      "DeleteCapacityProvider",
			eventName: "DeleteCapacityProvider",
			fields: map[string]string{
				"ct.request.capacityprovider": "my-provider",
			},
			wantKeys: []string{"deleted_capacity_provider"},
		},
		// EKS - Clusters
		{
			name:      "CreateCluster",
			eventName: "CreateCluster",
			fields: map[string]string{
				"ct.request.name":               "my-eks-cluster",
				"ct.request.version":            "1.28",
				"ct.request.rolearn":            "arn:aws:iam::123456789012:role/eks-cluster-role",
				"ct.request.resourcesvpcconfig": `{"subnetIds":["subnet-abc123","subnet-def456"]}`,
			},
			wantKeys: []string{"cluster_name", "version", "role_arn", "resources_vpc_config"},
		},
		{
			name:      "DeleteCluster",
			eventName: "DeleteCluster",
			fields: map[string]string{
				"ct.request.name": "my-eks-cluster",
			},
			wantKeys: []string{"deleted_cluster"},
		},
		{
			name:      "UpdateClusterConfig",
			eventName: "UpdateClusterConfig",
			fields: map[string]string{
				"ct.request.resourcesvpcconfig": `{"endpointPublicAccess":true}`,
				"ct.request.logging":            `{"clusterLogging":[{"types":["api","audit"],"enabled":true}]}`,
			},
			wantKeys: []string{"resources_vpc_config", "logging"},
		},
		{
			name:      "UpdateClusterVersion",
			eventName: "UpdateClusterVersion",
			fields: map[string]string{
				"ct.request.version": "1.29",
			},
			wantKeys: []string{"version"},
		},
		// EKS - Node Groups
		{
			name:      "CreateNodegroup",
			eventName: "CreateNodegroup",
			fields: map[string]string{
				"ct.request.nodegroupname": "my-nodegroup",
				"ct.request.clustername":   "my-eks-cluster",
				"ct.request.noderole":      "arn:aws:iam::123456789012:role/eks-node-role",
				"ct.request.subnets":       `["subnet-abc123","subnet-def456"]`,
				"ct.request.scalingconfig": `{"minSize":1,"maxSize":3,"desiredSize":2}`,
				"ct.request.instancetypes": `["t3.medium"]`,
				"ct.request.amitype":       "AL2_x86_64",
				"ct.request.disksize":      "20",
			},
			wantKeys: []string{"nodegroup_name", "cluster_name", "node_role_arn", "subnets", "scaling_config", "instance_types", "ami_type", "disk_size"},
		},
		{
			name:      "DeleteNodegroup",
			eventName: "DeleteNodegroup",
			fields: map[string]string{
				"ct.request.nodegroupname": "my-nodegroup",
			},
			wantKeys: []string{"deleted_nodegroup"},
		},
		{
			name:      "UpdateNodegroupConfig",
			eventName: "UpdateNodegroupConfig",
			fields: map[string]string{
				"ct.request.scalingconfig": `{"minSize":2,"maxSize":5,"desiredSize":3}`,
				"ct.request.labels":        `{"environment":"production"}`,
				"ct.request.taints":        `[{"key":"dedicated","value":"backend","effect":"NoSchedule"}]`,
			},
			wantKeys: []string{"scaling_config", "labels", "taints"},
		},
		{
			name:      "UpdateNodegroupVersion",
			eventName: "UpdateNodegroupVersion",
			fields: map[string]string{
				"ct.request.version":        "1.29",
				"ct.request.releaseversion": "1.29.0-20240129",
			},
			wantKeys: []string{"version", "release_version"},
		},
		// EKS - Addons
		{
			name:      "CreateAddon",
			eventName: "CreateAddon",
			fields: map[string]string{
				"ct.request.addonname":             "vpc-cni",
				"ct.request.clustername":           "my-eks-cluster",
				"ct.request.addonversion":          "v1.15.1-eksbuild.1",
				"ct.request.serviceaccountrolearn": "arn:aws:iam::123456789012:role/eks-addon-vpc-cni",
			},
			wantKeys: []string{"addon_name", "cluster_name", "addon_version", "service_account_role_arn"},
		},
		{
			name:      "DeleteAddon",
			eventName: "DeleteAddon",
			fields: map[string]string{
				"ct.request.addonname": "vpc-cni",
			},
			wantKeys: []string{"deleted_addon"},
		},
		{
			name:      "UpdateAddon",
			eventName: "UpdateAddon",
			fields: map[string]string{
				"ct.request.addonversion":          "v1.16.0-eksbuild.1",
				"ct.request.serviceaccountrolearn": "arn:aws:iam::123456789012:role/eks-addon-vpc-cni-v2",
				"ct.request.resolveconflicts":      "OVERWRITE",
			},
			wantKeys: []string{"addon_version", "service_account_role_arn", "resolve_conflicts"},
		},
		// EKS - Fargate Profiles
		{
			name:      "CreateFargateProfile",
			eventName: "CreateFargateProfile",
			fields: map[string]string{
				"ct.request.fargateprofilename": "my-fargate-profile",
				"ct.request.clustername":        "my-eks-cluster",
				"ct.request.podexecutionrolearn": "arn:aws:iam::123456789012:role/eks-fargate-pod-execution-role",
				"ct.request.subnets":            `["subnet-abc123","subnet-def456"]`,
				"ct.request.selectors":          `[{"namespace":"default"},{"namespace":"kube-system"}]`,
			},
			wantKeys: []string{"fargate_profile_name", "cluster_name", "pod_execution_role_arn", "subnets", "selectors"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.extractChanges(tt.eventName, tt.fields)

			// Check that all expected keys are present
			for _, key := range tt.wantKeys {
				assert.Contains(t, got, key, "Missing key: %s", key)
			}

			// Check that we don't have unexpected keys (except for special cases)
			if len(tt.wantKeys) > 0 {
				assert.Len(t, got, len(tt.wantKeys), "Unexpected number of keys")
			}
		})
	}
}

func TestExtractChanges_JSONParsing(t *testing.T) {
	sub := &Subscriber{}

	t.Run("Valid JSON Policy Document", func(t *testing.T) {
		fields := map[string]string{
			"ct.request.policydocument": `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:*","Resource":"*"}]}`,
		}

		changes := sub.extractChanges("UpdateAssumeRolePolicy", fields)

		assert.Contains(t, changes, "assume_role_policy")
		policy, ok := changes["assume_role_policy"].(map[string]interface{})
		require.True(t, ok, "Policy should be a map")
		assert.Equal(t, "2012-10-17", policy["Version"])
	})

	t.Run("Invalid JSON Policy Document", func(t *testing.T) {
		fields := map[string]string{
			"ct.request.policydocument": `{invalid json}`,
		}

		changes := sub.extractChanges("UpdateAssumeRolePolicy", fields)

		// Invalid JSON should not be added to changes
		assert.NotContains(t, changes, "assume_role_policy")
	})
}
