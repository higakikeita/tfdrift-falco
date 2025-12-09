# ECS (Elastic Container Service) — Drift Coverage

## Overview

TFDrift-Falco monitors Amazon ECS for configuration drift by tracking CloudTrail events related to services, task definitions, clusters, and capacity providers. This enables real-time detection of manual changes made outside of Terraform workflows.

## Supported CloudTrail Events

### Services (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateService | ECS service created | WARNING | ✔ |
| UpdateService | ECS service configuration updated | WARNING | ✔ |
| DeleteService | ECS service deleted | CRITICAL | ✔ |

### Task Definitions (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| RegisterTaskDefinition | New task definition registered | WARNING | ✔ |
| DeregisterTaskDefinition | Task definition deregistered | WARNING | ✔ |

### Clusters (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| UpdateCluster | Cluster configuration updated | WARNING | ✔ |
| UpdateClusterSettings | Cluster settings modified | WARNING | ✔ |
| PutClusterCapacityProviders | Capacity providers configured | WARNING | ✔ |
| UpdateContainerInstancesState | Container instance state changed | WARNING | ✔ |

### Capacity Providers (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateCapacityProvider | Capacity provider created | WARNING | ✔ |
| UpdateCapacityProvider | Capacity provider updated | WARNING | ✔ |
| DeleteCapacityProvider | Capacity provider deleted | WARNING | ✔ |

**Total: 13 CloudTrail events**

> **Note**: CreateCluster and DeleteCluster events are context-dependent and shared with EKS and Redshift. ECS-specific cluster management focuses on UpdateCluster and service-level events.

## Supported Terraform Resources

- `aws_ecs_service` — ECS service configuration
- `aws_ecs_task_definition` — Task definition specifications
- `aws_ecs_cluster` — ECS cluster settings
- `aws_ecs_cluster_capacity_providers` — Cluster capacity provider associations
- `aws_ecs_container_instance` — Container instance state
- `aws_ecs_capacity_provider` — Capacity provider configuration

## Monitored Drift Attributes

### ECS Services
- **desired_count** — Number of tasks to run
- **task_definition** — Task definition ARN or family:revision
- **launch_type** — FARGATE, EC2, or EXTERNAL
- **force_new_deployment** — Force new deployment flag
- **enable_execute_command** — ECS Exec enabled/disabled
- **service_name** — Service identifier
- **cluster** — Associated cluster

### Task Definitions
- **family** — Task definition family name
- **container_definitions** — Container configuration (JSON)
- **task_role_arn** — IAM role for task
- **execution_role_arn** — IAM role for ECS agent
- **network_mode** — awsvpc, bridge, host, none
- **cpu** — Task-level CPU units
- **memory** — Task-level memory (MB)
- **requires_compatibilities** — FARGATE, EC2

### Clusters
- **settings** — Container Insights, managed tags
- **capacity_providers** — Configured capacity providers
- **default_capacity_provider_strategy** — Default provider strategy

### Capacity Providers
- **name** — Capacity provider name
- **auto_scaling_group_provider** — ASG configuration (JSON)

## Falco Rule Examples

```yaml
# Service Configuration Changes
- rule: ECS Service Updated
  desc: Detect when an ECS service is updated
  condition: >
    ct.name="UpdateService"
  output: >
    ECS service updated
    (user=%ct.user service=%ct.request.service cluster=%ct.request.cluster
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, ecs, container]

# Critical Deletion Event
- rule: ECS Service Deleted
  desc: Detect when an ECS service is deleted
  condition: >
    ct.name="DeleteService"
  output: >
    ECS service deleted
    (user=%ct.user service=%ct.request.service cluster=%ct.request.cluster
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, ecs, container, security]

# Task Definition Changes
- rule: ECS Task Definition Registered
  desc: Detect when a new ECS task definition is registered
  condition: >
    ct.name="RegisterTaskDefinition"
  output: >
    ECS task definition registered
    (user=%ct.user family=%ct.request.family region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, ecs, container]

# Cluster Configuration
- rule: ECS Cluster Updated
  desc: Detect when an ECS cluster configuration is updated
  condition: >
    ct.name="UpdateCluster" or ct.name="UpdateClusterSettings"
  output: >
    ECS cluster updated
    (user=%ct.user cluster=%ct.request.cluster region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, ecs, container]
```

## Example Drift Scenarios

### Scenario 1: Service Scaling Outside Terraform

**CloudTrail Event:**
```json
{
  "eventName": "UpdateService",
  "requestParameters": {
    "service": "my-service",
    "cluster": "production",
    "desiredCount": 5
  },
  "userIdentity": {
    "principalId": "AIDAI23ABCD4EFGH5IJKL",
    "userName": "ops-admin"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_ecs_service.main
Changed: desired_count = 3 → 5
User: ops-admin (IAM User)
Region: us-east-1
Severity: HIGH
```

### Scenario 2: Task Definition Updated Manually

**CloudTrail Event:**
```json
{
  "eventName": "RegisterTaskDefinition",
  "requestParameters": {
    "family": "web-app",
    "containerDefinitions": [{
      "name": "nginx",
      "image": "nginx:1.21",
      "memory": 512
    }],
    "requiresCompatibilities": ["FARGATE"]
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_ecs_task_definition.web_app
Changed: container_definitions.image = nginx:1.20 → nginx:1.21
User: developer@example.com (Console)
Region: us-east-1
Severity: MEDIUM
```

### Scenario 3: Cluster Capacity Providers Modified

**CloudTrail Event:**
```json
{
  "eventName": "PutClusterCapacityProviders",
  "requestParameters": {
    "cluster": "production",
    "capacityProviders": ["FARGATE", "FARGATE_SPOT"],
    "defaultCapacityProviderStrategy": [{
      "capacityProvider": "FARGATE_SPOT",
      "weight": 1
    }]
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_ecs_cluster_capacity_providers.main
Changed: capacity_providers = ["FARGATE"] → ["FARGATE", "FARGATE_SPOT"]
User: admin (Assumed Role)
Region: us-east-1
Severity: HIGH
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "ECS Service Configuration"
    resource_types:
      - "aws_ecs_service"
    watched_attributes:
      - "desired_count"
      - "task_definition"
      - "launch_type"
      - "enable_execute_command"
    severity: "high"

  - name: "ECS Task Definition Changes"
    resource_types:
      - "aws_ecs_task_definition"
    watched_attributes:
      - "container_definitions"
      - "task_role_arn"
      - "execution_role_arn"
      - "cpu"
      - "memory"
    severity: "medium"

  - name: "ECS Cluster Settings"
    resource_types:
      - "aws_ecs_cluster"
      - "aws_ecs_cluster_capacity_providers"
    watched_attributes:
      - "settings"
      - "capacity_providers"
      - "default_capacity_provider_strategy"
    severity: "high"
```

## Grafana Dashboard Metrics

### Service Metrics
- ECS service updates by cluster
- Desired count changes over time
- Service deployments (forced vs planned)
- Task definition version changes

### Cluster Metrics
- Cluster configuration changes
- Capacity provider modifications
- Container instance state transitions

### User Activity
- Top users making ECS changes
- Changes by source (Console, CLI, API)
- Changes by time of day

## Known Limitations

### 1. CreateCluster/DeleteCluster Events
- These events are shared between ECS, EKS, and Redshift
- Context-specific detection requires additional CloudTrail fields
- ECS cluster lifecycle tracking relies on UpdateCluster events

### 2. Container Definition Complexity
- Full container definition comparison may generate verbose diffs
- Consider monitoring specific container attributes (image, memory, cpu)

### 3. Service Discovery Integration
- Changes to Service Discovery (Cloud Map) configurations are not tracked
- Monitor Route53 events separately for DNS changes

### 4. Task Execution
- RunTask and StartTask events are not tracked (focus on configuration drift)
- Monitor CloudWatch Logs for runtime behavior

## Best Practices

### 1. Service Deployment Strategy
```hcl
# Terraform - Use deployment_controller
resource "aws_ecs_service" "app" {
  name            = "my-app"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.app.arn
  desired_count   = 3

  deployment_controller {
    type = "ECS"  # or "CODE_DEPLOY" for blue/green
  }

  lifecycle {
    ignore_changes = [desired_count]  # Allow auto-scaling
  }
}
```

### 2. Task Definition Management
```hcl
# Use terraform_data to trigger updates
resource "terraform_data" "app_image" {
  input = var.app_image_tag
}

resource "aws_ecs_task_definition" "app" {
  family = "my-app"

  container_definitions = jsonencode([{
    name  = "app"
    image = "my-repo:${terraform_data.app_image.output}"
  }])

  # Track all changes
  lifecycle {
    create_before_destroy = true
  }
}
```

### 3. Capacity Provider Strategy
```hcl
# Define explicit capacity provider strategy
resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 4
    base              = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}
```

### 4. ECS Exec Security
```hcl
# Monitor enable_execute_command changes
resource "aws_ecs_service" "app" {
  name    = "my-app"
  cluster = aws_ecs_cluster.main.id

  enable_execute_command = false  # Disable for production

  # Alert on any changes to this setting
}
```

## Security Considerations

### 1. IAM Permissions
- Monitor changes to task_role_arn and execution_role_arn
- Alert on privilege escalation via role changes

### 2. Network Configuration
- Track changes to network_mode (especially awsvpc → bridge/host)
- Monitor security group associations

### 3. Container Image Sources
- Track unauthorized image registries
- Alert on image tag changes (e.g., latest → specific version)

### 4. ECS Exec Access
- Monitor enable_execute_command flag changes
- Track usage via CloudTrail ExecuteCommand events

## Troubleshooting

### High Alert Volume
- **Problem**: Too many alerts for routine scaling operations
- **Solution**: Use lifecycle.ignore_changes for desired_count with auto-scaling

### Missing Task Definition Changes
- **Problem**: Task definition updates not detected
- **Solution**: Verify RegisterTaskDefinition events are in CloudTrail
- Check Falco plugin configuration for ECS event filtering

### Cluster Updates Not Detected
- **Problem**: UpdateCluster alerts not appearing
- **Solution**: Ensure cluster name/ARN extraction is working
- Verify ct.request.cluster field in CloudTrail events

## Related Documentation

- [AWS ECS CloudTrail Events](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/logging-using-cloudtrail.html)
- [Terraform aws_ecs_service](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_service)
- [Terraform aws_ecs_task_definition](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_task_definition)
- [Terraform aws_ecs_cluster](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ecs_cluster)

## Version History

- **v0.3.0** (2025 Q1) - Initial ECS support with 13 CloudTrail events
  - Services, Task Definitions, Clusters, Capacity Providers
  - Comprehensive drift detection for container orchestration
