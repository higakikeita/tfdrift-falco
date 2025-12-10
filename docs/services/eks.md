# EKS (Elastic Kubernetes Service) — Drift Coverage

## Overview

TFDrift-Falco monitors Amazon EKS for configuration drift by tracking CloudTrail events related to clusters, node groups, addons, and Fargate profiles. This enables real-time detection of manual changes made outside of Terraform workflows.

## Supported CloudTrail Events

### Clusters (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateCluster | EKS cluster created | WARNING | ✔ |
| DeleteCluster | EKS cluster deleted | CRITICAL | ✔ |
| UpdateClusterConfig | Cluster configuration updated | WARNING | ✔ |
| UpdateClusterVersion | Kubernetes version upgraded | WARNING | ✔ |

### Node Groups (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateNodegroup | Node group created | WARNING | ✔ |
| DeleteNodegroup | Node group deleted | CRITICAL | ✔ |
| UpdateNodegroupConfig | Node group configuration updated | WARNING | ✔ |
| UpdateNodegroupVersion | Node group version updated | WARNING | ✔ |

### Addons (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateAddon | EKS addon created | WARNING | ✔ |
| DeleteAddon | EKS addon deleted | WARNING | ✔ |
| UpdateAddon | EKS addon updated | WARNING | ✔ |

### Fargate Profiles (1 event)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateFargateProfile | Fargate profile created | WARNING | ✔ |

**Total: 12 CloudTrail events**

> **Note**: CreateCluster and DeleteCluster events are context-dependent and shared with ECS and Redshift. EKS-specific detection requires additional fields to distinguish between services.

## Supported Terraform Resources

- `aws_eks_cluster` — EKS cluster configuration
- `aws_eks_node_group` — Managed node group specifications
- `aws_eks_addon` — EKS addons (VPC-CNI, CoreDNS, kube-proxy, etc.)
- `aws_eks_fargate_profile` — Fargate profile configuration

## Monitored Drift Attributes

### EKS Clusters
- **name** — Cluster identifier
- **version** — Kubernetes version (e.g., 1.28, 1.29)
- **role_arn** — IAM role for cluster service account
- **resources_vpc_config** — VPC configuration (subnets, security groups, endpoint access)
- **logging** — Control plane logging configuration
- **encryption_config** — Secrets encryption with KMS

### Node Groups
- **nodegroup_name** — Node group identifier
- **cluster_name** — Associated cluster
- **node_role_arn** — IAM role for worker nodes
- **subnets** — Subnet IDs for node placement
- **scaling_config** — Min, max, desired size
- **instance_types** — EC2 instance types
- **ami_type** — AMI type (AL2_x86_64, AL2_ARM_64, etc.)
- **disk_size** — Root volume size (GB)
- **labels** — Kubernetes labels
- **taints** — Kubernetes taints
- **version** — Kubernetes version
- **release_version** — AMI release version

### Addons
- **addon_name** — Addon identifier (vpc-cni, coredns, kube-proxy, etc.)
- **cluster_name** — Associated cluster
- **addon_version** — Addon version
- **service_account_role_arn** — IAM role for service account (IRSA)
- **resolve_conflicts** — Conflict resolution strategy (OVERWRITE, PRESERVE)

### Fargate Profiles
- **fargate_profile_name** — Profile identifier
- **cluster_name** — Associated cluster
- **pod_execution_role_arn** — IAM role for Fargate pod execution
- **subnets** — Subnet IDs for Fargate pods
- **selectors** — Pod selectors (namespace, labels)

## Falco Rule Examples

```yaml
# Cluster Lifecycle
- rule: EKS Cluster Created
  desc: Detect when an EKS cluster is created
  condition: >
    ct.name="CreateCluster" and ct.request.name exists
  output: >
    EKS cluster created
    (user=%ct.user cluster=%ct.request.name version=%ct.request.version
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, eks, kubernetes]

- rule: EKS Cluster Deleted
  desc: Detect when an EKS cluster is deleted
  condition: >
    ct.name="DeleteCluster"
  output: >
    EKS cluster deleted
    (user=%ct.user cluster=%ct.request.name region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, eks, kubernetes, security]

# Configuration Changes
- rule: EKS Cluster Version Upgraded
  desc: Detect when an EKS cluster Kubernetes version is upgraded
  condition: >
    ct.name="UpdateClusterVersion"
  output: >
    EKS cluster version upgraded
    (user=%ct.user cluster=%ct.request.name version=%ct.request.version
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, eks, kubernetes]

# Node Group Management
- rule: EKS Node Group Configuration Updated
  desc: Detect when an EKS node group configuration is updated
  condition: >
    ct.name="UpdateNodegroupConfig"
  output: >
    EKS node group configuration updated
    (user=%ct.user nodegroup=%ct.request.nodegroupName cluster=%ct.request.clusterName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, eks, kubernetes]
```

## Example Drift Scenarios

### Scenario 1: Cluster Configuration Changed Outside Terraform

**CloudTrail Event:**
```json
{
  "eventName": "UpdateClusterConfig",
  "requestParameters": {
    "name": "production-cluster",
    "resourcesVpcConfig": {
      "endpointPublicAccess": true,
      "endpointPrivateAccess": false
    },
    "logging": {
      "clusterLogging": [{
        "types": ["api", "audit"],
        "enabled": true
      }]
    }
  },
  "userIdentity": {
    "principalId": "AIDAI23ABCD4EFGH5IJKL",
    "userName": "ops-admin"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_eks_cluster.production
Changed:
  - resources_vpc_config.endpoint_public_access = false → true
  - logging.cluster_logging[0].enabled = false → true
User: ops-admin (IAM User)
Region: us-east-1
Severity: HIGH
```

### Scenario 2: Node Group Scaling Modified Manually

**CloudTrail Event:**
```json
{
  "eventName": "UpdateNodegroupConfig",
  "requestParameters": {
    "clusterName": "production-cluster",
    "nodegroupName": "general-workers",
    "scalingConfig": {
      "minSize": 5,
      "maxSize": 20,
      "desiredSize": 10
    }
  },
  "userIdentity": {
    "type": "AssumedRole",
    "principalId": "AIDAI23ABCD4EFGH5IJKL:admin-session",
    "arn": "arn:aws:sts::123456789012:assumed-role/AdminRole/admin-session"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_eks_node_group.general_workers
Changed:
  - scaling_config.min_size = 3 → 5
  - scaling_config.max_size = 10 → 20
  - scaling_config.desired_size = 5 → 10
User: AdminRole (Assumed Role)
Region: us-east-1
Severity: MEDIUM
```

### Scenario 3: Addon Version Updated

**CloudTrail Event:**
```json
{
  "eventName": "UpdateAddon",
  "requestParameters": {
    "clusterName": "production-cluster",
    "addonName": "vpc-cni",
    "addonVersion": "v1.16.0-eksbuild.1",
    "resolveConflicts": "OVERWRITE"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_eks_addon.vpc_cni
Changed:
  - addon_version = v1.15.1-eksbuild.1 → v1.16.0-eksbuild.1
  - resolve_conflicts = PRESERVE → OVERWRITE
User: developer@example.com (Console)
Region: us-east-1
Severity: MEDIUM
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "EKS Cluster Configuration"
    resource_types:
      - "aws_eks_cluster"
    watched_attributes:
      - "version"
      - "resources_vpc_config"
      - "logging"
    severity: "high"

  - name: "EKS Node Group Settings"
    resource_types:
      - "aws_eks_node_group"
    watched_attributes:
      - "scaling_config"
      - "instance_types"
      - "version"
      - "labels"
      - "taints"
    severity: "high"

  - name: "EKS Addon Management"
    resource_types:
      - "aws_eks_addon"
    watched_attributes:
      - "addon_version"
      - "service_account_role_arn"
      - "resolve_conflicts"
    severity: "medium"
```

## Grafana Dashboard Metrics

### Cluster Metrics
- EKS cluster version upgrades over time
- Cluster configuration changes by type
- Endpoint access changes (public/private)
- Logging configuration changes

### Node Group Metrics
- Node group scaling events
- Instance type changes
- AMI version updates
- Label and taint modifications

### Addon Metrics
- Addon version updates by addon name
- IRSA role changes
- Conflict resolution strategy changes

### User Activity
- Top users making EKS changes
- Changes by source (Console, CLI, Terraform)
- Changes by time of day and cluster

## Known Limitations

### 1. CreateCluster/DeleteCluster Event Ambiguity
- These events are shared between EKS, ECS, and Redshift
- Requires additional CloudTrail fields (e.g., `ct.request.name`) to distinguish EKS clusters
- Context-specific detection implemented in Falco rules

### 2. Fargate Profile Deletion
- `DeleteFargateProfile` event is not tracked (not in v0.3.0 scope)
- Fargate profile lifecycle tracking is limited to creation
- Monitor AWS Config for comprehensive Fargate profile tracking

### 3. Node Group Launch Templates
- Changes to launch templates are not directly tracked
- Monitor EC2 `CreateLaunchTemplateVersion` events separately
- Node group updates may reference new launch template versions

### 4. Self-Managed Node Groups
- Self-managed node groups (via ASG/EC2) are not covered
- Monitor EC2 and Auto Scaling events separately
- Focus is on EKS-managed node groups

### 5. Kubernetes API Changes
- Changes made via kubectl/Kubernetes API are not tracked
- CloudTrail only captures EKS control plane API changes
- Consider using Kubernetes audit logs for in-cluster changes

## Best Practices

### 1. Cluster Version Management
```hcl
# Terraform - Explicit version pinning
resource "aws_eks_cluster" "main" {
  name     = "production-cluster"
  version  = "1.28"
  role_arn = aws_iam_role.cluster.arn

  vpc_config {
    subnet_ids              = aws_subnet.private[*].id
    endpoint_private_access = true
    endpoint_public_access  = false
  }

  enabled_cluster_log_types = ["api", "audit", "authenticator"]

  lifecycle {
    # Prevent accidental version downgrades
    prevent_destroy = true
  }
}
```

### 2. Node Group Auto-Scaling
```hcl
# Allow auto-scaling to manage desired_size
resource "aws_eks_node_group" "workers" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "general-workers"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = aws_subnet.private[*].id

  scaling_config {
    min_size     = 3
    max_size     = 10
    desired_size = 5
  }

  lifecycle {
    # Allow Cluster Autoscaler to modify desired_size
    ignore_changes = [scaling_config[0].desired_size]
  }
}
```

### 3. Addon Management with IRSA
```hcl
# Use IAM Roles for Service Accounts (IRSA)
resource "aws_eks_addon" "vpc_cni" {
  cluster_name             = aws_eks_cluster.main.name
  addon_name               = "vpc-cni"
  addon_version            = "v1.15.1-eksbuild.1"
  service_account_role_arn = aws_iam_role.vpc_cni.arn
  resolve_conflicts        = "PRESERVE"

  lifecycle {
    # Explicit version management
    ignore_changes = []
  }
}
```

### 4. Fargate Profile for Specific Workloads
```hcl
# Fargate for serverless Kubernetes pods
resource "aws_eks_fargate_profile" "app" {
  cluster_name           = aws_eks_cluster.main.name
  fargate_profile_name   = "app-profile"
  pod_execution_role_arn = aws_iam_role.fargate_pod_execution.arn
  subnet_ids             = aws_subnet.private[*].id

  selector {
    namespace = "production"
    labels = {
      workload = "serverless"
    }
  }

  selector {
    namespace = "staging"
  }
}
```

## Security Considerations

### 1. Cluster Endpoint Access
- **Private-only access**: Recommended for production clusters
- **Public access**: Monitor `UpdateClusterConfig` for unauthorized changes
- Alert on transitions from private to public endpoint access

### 2. IAM Role Changes
- Track changes to `role_arn` (cluster service role)
- Monitor `node_role_arn` (worker node IAM role)
- Alert on privilege escalation via role modifications

### 3. Logging Configuration
- Ensure control plane logging is enabled (api, audit, authenticator)
- Alert on logging disablement via `UpdateClusterConfig`
- Monitor CloudWatch Logs for audit trail

### 4. Addon Security
- Track IRSA role changes (`service_account_role_arn`)
- Monitor addon version updates for security patches
- Alert on `resolveConflicts=OVERWRITE` (may override custom configurations)

### 5. Network Security
- Monitor VPC configuration changes (security groups, subnets)
- Alert on unauthorized subnet additions
- Track security group association changes

## Troubleshooting

### High Alert Volume from Auto-Scaling
- **Problem**: Too many alerts for node group scaling events
- **Solution**: Use `lifecycle.ignore_changes = [scaling_config[0].desired_size]` in Terraform
- Configure Cluster Autoscaler to work with Terraform-managed node groups

### Missing Cluster Creation/Deletion Events
- **Problem**: CreateCluster/DeleteCluster events not detected
- **Solution**: Verify Falco rule includes `ct.request.name exists` condition
- Check CloudTrail for correct event format and service context

### Addon Update Conflicts
- **Problem**: Addon updates fail with conflict errors
- **Solution**: Review `resolveConflicts` setting (PRESERVE vs OVERWRITE)
- Check for manual changes to addon configurations

### Version Upgrade Detection Issues
- **Problem**: Kubernetes version upgrades not detected
- **Solution**: Ensure `UpdateClusterVersion` events are in CloudTrail
- Verify Falco plugin is processing EKS events correctly

## Related Documentation

- [AWS EKS CloudTrail Events](https://docs.aws.amazon.com/eks/latest/userguide/logging-using-cloudtrail.html)
- [Terraform aws_eks_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_cluster)
- [Terraform aws_eks_node_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_node_group)
- [Terraform aws_eks_addon](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_addon)
- [Terraform aws_eks_fargate_profile](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/eks_fargate_profile)

## Version History

- **v0.3.0** (2025 Q1) - Initial EKS support with 12 CloudTrail events
  - Clusters, Node Groups, Addons, Fargate Profiles
  - Comprehensive drift detection for Kubernetes control plane
