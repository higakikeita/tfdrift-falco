// Package falco provides integration with Falco for event processing.
package falco

import (
	"encoding/json"
)

// extractChanges extracts the changed attributes from Falco output
func (s *Subscriber) extractChanges(eventName string, fields map[string]string) map[string]interface{} {
	changes := make(map[string]interface{})

	switch eventName {
	case "ModifyInstanceAttribute":
		if val, ok := fields["ct.request.disableapitermination"]; ok && val != "" {
			changes["disable_api_termination"] = val
		}
		if val, ok := fields["ct.request.instancetype"]; ok && val != "" {
			changes["instance_type"] = val
		}

	case "PutBucketEncryption":
		// Encryption enabled
		if config, ok := fields["ct.request.serversideencryptionconfiguration"]; ok && config != "" {
			changes["server_side_encryption_configuration"] = config
		}

	case "DeleteBucketEncryption":
		// Encryption disabled
		changes["server_side_encryption_configuration"] = nil

	case "UpdateFunctionConfiguration":
		if val, ok := fields["ct.request.timeout"]; ok && val != "" {
			changes["timeout"] = val
		}
		if val, ok := fields["ct.request.memorysize"]; ok && val != "" {
			changes["memory_size"] = val
		}

	case "UpdateAssumeRolePolicy":
		if policy := getStringField(fields, "ct.request.policydocument"); policy != "" {
			var policyDoc map[string]interface{}
			if err := json.Unmarshal([]byte(policy), &policyDoc); err == nil {
				changes["assume_role_policy"] = policyDoc
			}
		}

	// IAM Policy attachments
	case "AttachRolePolicy", "AttachUserPolicy", "AttachGroupPolicy":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["attached_policy_arn"] = policyArn
		}

	// IAM Inline policies
	case "PutRolePolicy", "PutUserPolicy", "PutGroupPolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["inline_policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy creation
	case "CreatePolicy":
		if policyName, ok := fields["ct.request.policyname"]; ok && policyName != "" {
			changes["policy_name"] = policyName
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Policy version
	case "CreatePolicyVersion":
		if policyArn, ok := fields["ct.request.policyarn"]; ok && policyArn != "" {
			changes["policy_arn"] = policyArn
		}
		if setDefault, ok := fields["ct.request.setasdefault"]; ok && setDefault != "" {
			changes["set_as_default"] = setDefault
		}
		if policyDoc := getStringField(fields, "ct.request.policydocument"); policyDoc != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(policyDoc), &doc); err == nil {
				changes["policy_document"] = doc
			}
		}

	// IAM Role creation
	case "CreateRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["role_name"] = roleName
		}
		if assumePolicy := getStringField(fields, "ct.request.assumerolepolicydocument"); assumePolicy != "" {
			var doc map[string]interface{}
			if err := json.Unmarshal([]byte(assumePolicy), &doc); err == nil {
				changes["assume_role_policy"] = doc
			}
		}

	// IAM User/Role deletion
	case "DeleteRole":
		if roleName, ok := fields["ct.request.rolename"]; ok && roleName != "" {
			changes["deleted_role"] = roleName
		}
	case "DeleteUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["deleted_user"] = userName
		}

	// IAM User creation
	case "CreateUser":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}

	// IAM Access Key creation
	case "CreateAccessKey":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if accessKeyId := getStringField(fields, "ct.response.accesskey.accesskeyid"); accessKeyId != "" {
			changes["access_key_id"] = accessKeyId
		}

	// IAM Group membership
	case "AddUserToGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}
	case "RemoveUserFromGroup":
		if userName, ok := fields["ct.request.username"]; ok && userName != "" {
			changes["user_name"] = userName
		}
		if groupName, ok := fields["ct.request.groupname"]; ok && groupName != "" {
			changes["group_name"] = groupName
		}

	// Account password policy
	case "UpdateAccountPasswordPolicy":
		// Extract various password policy fields if available
		if minLength, ok := fields["ct.request.minimumpasswordlength"]; ok && minLength != "" {
			changes["minimum_password_length"] = minLength
		}
		if requireSymbols, ok := fields["ct.request.requiresymbols"]; ok && requireSymbols != "" {
			changes["require_symbols"] = requireSymbols
		}

	// ECS - Services
	case "CreateService":
		if serviceName, ok := fields["ct.request.servicename"]; ok && serviceName != "" {
			changes["service_name"] = serviceName
		}
		if cluster, ok := fields["ct.request.cluster"]; ok && cluster != "" {
			changes["cluster"] = cluster
		}
		if taskDefinition, ok := fields["ct.request.taskdefinition"]; ok && taskDefinition != "" {
			changes["task_definition"] = taskDefinition
		}
		if desiredCount, ok := fields["ct.request.desiredcount"]; ok && desiredCount != "" {
			changes["desired_count"] = desiredCount
		}
		if launchType, ok := fields["ct.request.launchtype"]; ok && launchType != "" {
			changes["launch_type"] = launchType
		}

	case "UpdateService":
		if desiredCount, ok := fields["ct.request.desiredcount"]; ok && desiredCount != "" {
			changes["desired_count"] = desiredCount
		}
		if taskDefinition, ok := fields["ct.request.taskdefinition"]; ok && taskDefinition != "" {
			changes["task_definition"] = taskDefinition
		}
		if launchType, ok := fields["ct.request.launchtype"]; ok && launchType != "" {
			changes["launch_type"] = launchType
		}
		if forceNewDeployment, ok := fields["ct.request.forcenewdeployment"]; ok && forceNewDeployment != "" {
			changes["force_new_deployment"] = forceNewDeployment
		}
		if enableExecuteCommand, ok := fields["ct.request.enableexecutecommand"]; ok && enableExecuteCommand != "" {
			changes["enable_execute_command"] = enableExecuteCommand
		}

	case "DeleteService":
		if service, ok := fields["ct.request.service"]; ok && service != "" {
			changes["deleted_service"] = service
		}
		if force, ok := fields["ct.request.force"]; ok && force != "" {
			changes["force"] = force
		}

	// ECS - Task Definitions
	case "RegisterTaskDefinition":
		if family, ok := fields["ct.request.family"]; ok && family != "" {
			changes["family"] = family
		}
		if containerDefs := getStringField(fields, "ct.request.containerdefinitions"); containerDefs != "" {
			var containers []interface{}
			if err := json.Unmarshal([]byte(containerDefs), &containers); err == nil {
				changes["container_definitions"] = containers
			}
		}
		if taskRoleArn, ok := fields["ct.request.taskrolearn"]; ok && taskRoleArn != "" {
			changes["task_role_arn"] = taskRoleArn
		}
		if executionRoleArn, ok := fields["ct.request.executionrolearn"]; ok && executionRoleArn != "" {
			changes["execution_role_arn"] = executionRoleArn
		}
		if networkMode, ok := fields["ct.request.networkmode"]; ok && networkMode != "" {
			changes["network_mode"] = networkMode
		}
		if cpu, ok := fields["ct.request.cpu"]; ok && cpu != "" {
			changes["cpu"] = cpu
		}
		if memory, ok := fields["ct.request.memory"]; ok && memory != "" {
			changes["memory"] = memory
		}
		if requiresCompatibilities := getStringField(fields, "ct.request.requirescompatibilities"); requiresCompatibilities != "" {
			changes["requires_compatibilities"] = requiresCompatibilities
		}

	case "DeregisterTaskDefinition":
		if taskDef, ok := fields["ct.request.taskdefinition"]; ok && taskDef != "" {
			changes["deregistered_task_definition"] = taskDef
		}

	// ECS - Clusters
	// Note: CreateCluster/DeleteCluster are context-dependent (ECS, EKS, Redshift)
	case "UpdateCluster":
		if settings := getStringField(fields, "ct.request.settings"); settings != "" {
			var settingsArray []interface{}
			if err := json.Unmarshal([]byte(settings), &settingsArray); err == nil {
				changes["settings"] = settingsArray
			}
		}

	case "UpdateClusterSettings":
		if settings := getStringField(fields, "ct.request.settings"); settings != "" {
			var settingsArray []interface{}
			if err := json.Unmarshal([]byte(settings), &settingsArray); err == nil {
				changes["settings"] = settingsArray
			}
		}

	case "PutClusterCapacityProviders":
		if capacityProviders := getStringField(fields, "ct.request.capacityproviders"); capacityProviders != "" {
			var providers []interface{}
			if err := json.Unmarshal([]byte(capacityProviders), &providers); err == nil {
				changes["capacity_providers"] = providers
			}
		}
		if defaultStrategy := getStringField(fields, "ct.request.defaultcapacityproviderstrategy"); defaultStrategy != "" {
			var strategy []interface{}
			if err := json.Unmarshal([]byte(defaultStrategy), &strategy); err == nil {
				changes["default_capacity_provider_strategy"] = strategy
			}
		}

	// ECS - Container Instances
	case "UpdateContainerInstancesState":
		if status, ok := fields["ct.request.status"]; ok && status != "" {
			changes["status"] = status
		}

	// ECS - Capacity Providers
	case "CreateCapacityProvider":
		if name, ok := fields["ct.request.name"]; ok && name != "" {
			changes["name"] = name
		}
		if autoScalingConfig := getStringField(fields, "ct.request.autoscalinggroupprovider"); autoScalingConfig != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(autoScalingConfig), &config); err == nil {
				changes["auto_scaling_group_provider"] = config
			}
		}

	case "UpdateCapacityProvider":
		if autoScalingConfig := getStringField(fields, "ct.request.autoscalinggroupprovider"); autoScalingConfig != "" {
			var config map[string]interface{}
			if err := json.Unmarshal([]byte(autoScalingConfig), &config); err == nil {
				changes["auto_scaling_group_provider"] = config
			}
		}

	case "DeleteCapacityProvider":
		if capacityProvider, ok := fields["ct.request.capacityprovider"]; ok && capacityProvider != "" {
			changes["deleted_capacity_provider"] = capacityProvider
		}

	// EKS - Clusters
	case "CreateCluster":
		if name, ok := fields["ct.request.name"]; ok && name != "" {
			changes["cluster_name"] = name
		}
		if version, ok := fields["ct.request.version"]; ok && version != "" {
			changes["version"] = version
		}
		if roleArn, ok := fields["ct.request.rolearn"]; ok && roleArn != "" {
			changes["role_arn"] = roleArn
		}
		if resourcesVpcConfig := getStringField(fields, "ct.request.resourcesvpcconfig"); resourcesVpcConfig != "" {
			var vpcConfig map[string]interface{}
			if err := json.Unmarshal([]byte(resourcesVpcConfig), &vpcConfig); err == nil {
				changes["resources_vpc_config"] = vpcConfig
			}
		}

	case "DeleteCluster":
		if name, ok := fields["ct.request.name"]; ok && name != "" {
			changes["deleted_cluster"] = name
		}

	case "UpdateClusterConfig":
		if resourcesVpcConfig := getStringField(fields, "ct.request.resourcesvpcconfig"); resourcesVpcConfig != "" {
			var vpcConfig map[string]interface{}
			if err := json.Unmarshal([]byte(resourcesVpcConfig), &vpcConfig); err == nil {
				changes["resources_vpc_config"] = vpcConfig
			}
		}
		if logging := getStringField(fields, "ct.request.logging"); logging != "" {
			var loggingConfig map[string]interface{}
			if err := json.Unmarshal([]byte(logging), &loggingConfig); err == nil {
				changes["logging"] = loggingConfig
			}
		}

	case "UpdateClusterVersion":
		if version, ok := fields["ct.request.version"]; ok && version != "" {
			changes["version"] = version
		}

	// EKS - Node Groups
	case "CreateNodegroup":
		if nodegroupName, ok := fields["ct.request.nodegroupname"]; ok && nodegroupName != "" {
			changes["nodegroup_name"] = nodegroupName
		}
		if clusterName, ok := fields["ct.request.clustername"]; ok && clusterName != "" {
			changes["cluster_name"] = clusterName
		}
		if nodeRole, ok := fields["ct.request.noderole"]; ok && nodeRole != "" {
			changes["node_role_arn"] = nodeRole
		}
		if subnets := getStringField(fields, "ct.request.subnets"); subnets != "" {
			var subnetList []interface{}
			if err := json.Unmarshal([]byte(subnets), &subnetList); err == nil {
				changes["subnets"] = subnetList
			}
		}
		if scalingConfig := getStringField(fields, "ct.request.scalingconfig"); scalingConfig != "" {
			var scaling map[string]interface{}
			if err := json.Unmarshal([]byte(scalingConfig), &scaling); err == nil {
				changes["scaling_config"] = scaling
			}
		}
		if instanceTypes := getStringField(fields, "ct.request.instancetypes"); instanceTypes != "" {
			var types []interface{}
			if err := json.Unmarshal([]byte(instanceTypes), &types); err == nil {
				changes["instance_types"] = types
			}
		}
		if amiType, ok := fields["ct.request.amitype"]; ok && amiType != "" {
			changes["ami_type"] = amiType
		}
		if diskSize, ok := fields["ct.request.disksize"]; ok && diskSize != "" {
			changes["disk_size"] = diskSize
		}

	case "DeleteNodegroup":
		if nodegroupName, ok := fields["ct.request.nodegroupname"]; ok && nodegroupName != "" {
			changes["deleted_nodegroup"] = nodegroupName
		}

	case "UpdateNodegroupConfig":
		if scalingConfig := getStringField(fields, "ct.request.scalingconfig"); scalingConfig != "" {
			var scaling map[string]interface{}
			if err := json.Unmarshal([]byte(scalingConfig), &scaling); err == nil {
				changes["scaling_config"] = scaling
			}
		}
		if labels := getStringField(fields, "ct.request.labels"); labels != "" {
			var labelsMap map[string]interface{}
			if err := json.Unmarshal([]byte(labels), &labelsMap); err == nil {
				changes["labels"] = labelsMap
			}
		}
		if taints := getStringField(fields, "ct.request.taints"); taints != "" {
			var taintsList []interface{}
			if err := json.Unmarshal([]byte(taints), &taintsList); err == nil {
				changes["taints"] = taintsList
			}
		}

	case "UpdateNodegroupVersion":
		if version, ok := fields["ct.request.version"]; ok && version != "" {
			changes["version"] = version
		}
		if releaseVersion, ok := fields["ct.request.releaseversion"]; ok && releaseVersion != "" {
			changes["release_version"] = releaseVersion
		}

	// EKS - Addons
	case "CreateAddon":
		if addonName, ok := fields["ct.request.addonname"]; ok && addonName != "" {
			changes["addon_name"] = addonName
		}
		if clusterName, ok := fields["ct.request.clustername"]; ok && clusterName != "" {
			changes["cluster_name"] = clusterName
		}
		if addonVersion, ok := fields["ct.request.addonversion"]; ok && addonVersion != "" {
			changes["addon_version"] = addonVersion
		}
		if serviceAccountRoleArn, ok := fields["ct.request.serviceaccountrolearn"]; ok && serviceAccountRoleArn != "" {
			changes["service_account_role_arn"] = serviceAccountRoleArn
		}

	case "DeleteAddon":
		if addonName, ok := fields["ct.request.addonname"]; ok && addonName != "" {
			changes["deleted_addon"] = addonName
		}

	case "UpdateAddon":
		if addonVersion, ok := fields["ct.request.addonversion"]; ok && addonVersion != "" {
			changes["addon_version"] = addonVersion
		}
		if serviceAccountRoleArn, ok := fields["ct.request.serviceaccountrolearn"]; ok && serviceAccountRoleArn != "" {
			changes["service_account_role_arn"] = serviceAccountRoleArn
		}
		if resolveConflicts, ok := fields["ct.request.resolveconflicts"]; ok && resolveConflicts != "" {
			changes["resolve_conflicts"] = resolveConflicts
		}

	// EKS - Fargate Profiles
	case "CreateFargateProfile":
		if fargateProfileName, ok := fields["ct.request.fargateprofilename"]; ok && fargateProfileName != "" {
			changes["fargate_profile_name"] = fargateProfileName
		}
		if clusterName, ok := fields["ct.request.clustername"]; ok && clusterName != "" {
			changes["cluster_name"] = clusterName
		}
		if podExecutionRoleArn, ok := fields["ct.request.podexecutionrolearn"]; ok && podExecutionRoleArn != "" {
			changes["pod_execution_role_arn"] = podExecutionRoleArn
		}
		if subnets := getStringField(fields, "ct.request.subnets"); subnets != "" {
			var subnetList []interface{}
			if err := json.Unmarshal([]byte(subnets), &subnetList); err == nil {
				changes["subnets"] = subnetList
			}
		}
		if selectors := getStringField(fields, "ct.request.selectors"); selectors != "" {
			var selectorsList []interface{}
			if err := json.Unmarshal([]byte(selectors), &selectorsList); err == nil {
				changes["selectors"] = selectorsList
			}
		}
	}

	return changes
}
