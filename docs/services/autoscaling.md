# Auto Scaling (Amazon EC2 Auto Scaling) — Drift Coverage

## Overview

TFDrift-Falco monitors Amazon EC2 Auto Scaling for configuration drift by tracking CloudTrail events related to Auto Scaling groups, launch configurations, scaling policies, and scheduled actions. This enables real-time detection of manual changes made outside of Terraform workflows.

## Supported CloudTrail Events

### Auto Scaling Group Management (4 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateAutoScalingGroup | Auto Scaling group created | CRITICAL | ✔ |
| DeleteAutoScalingGroup | Auto Scaling group deleted | CRITICAL | ✔ |
| UpdateAutoScalingGroup | Auto Scaling group configuration updated | WARNING | ✔ |
| SetDesiredCapacity | Desired capacity manually set | WARNING | ✔ |

### Launch Configuration Management (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateLaunchConfiguration | Launch configuration created | WARNING | ✔ |
| DeleteLaunchConfiguration | Launch configuration deleted | WARNING | ✔ |

### Scaling Policy Management (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| PutScalingPolicy | Scaling policy created/updated | WARNING | ✔ |
| DeletePolicy | Scaling policy deleted | WARNING | ✔ |

### Scheduled Action Management (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| PutScheduledUpdateGroupAction | Scheduled action created/updated | WARNING | ✔ |
| DeleteScheduledAction | Scheduled action deleted | WARNING | ✔ |

**Total: 10 CloudTrail events**

## Supported Terraform Resources

- `aws_autoscaling_group` — Auto Scaling group configuration
- `aws_launch_configuration` — Launch configuration (EC2-Classic or default VPC)
- `aws_autoscaling_policy` — Scaling policies (target tracking, step, simple)
- `aws_autoscaling_schedule` — Scheduled scaling actions

## Monitored Drift Attributes

### Auto Scaling Groups
- **name** — Auto Scaling group name
- **launch_configuration** — Launch configuration name
- **launch_template** — Launch template specification
- **min_size** — Minimum capacity
- **max_size** — Maximum capacity
- **desired_capacity** — Desired capacity
- **default_cooldown** — Default cooldown period (seconds)
- **health_check_type** — Health check type (EC2, ELB)
- **health_check_grace_period** — Grace period (seconds)
- **vpc_zone_identifier** — VPC subnet IDs
- **availability_zones** — Availability zones
- **load_balancers** — Classic load balancer names
- **target_group_arns** — ALB/NLB target group ARNs
- **termination_policies** — Instance termination policies
- **enabled_metrics** — CloudWatch metrics collection
- **suspended_processes** — Suspended scaling processes
- **tags** — Resource tags

### Launch Configurations
- **name** — Launch configuration name
- **image_id** — AMI ID
- **instance_type** — EC2 instance type
- **key_name** — SSH key pair name
- **security_groups** — Security group IDs
- **user_data** — User data script
- **iam_instance_profile** — IAM instance profile
- **ebs_optimized** — EBS optimization flag
- **enable_monitoring** — Detailed monitoring
- **spot_price** — Spot instance max price
- **root_block_device** — Root volume configuration
- **ebs_block_device** — Additional EBS volumes

### Scaling Policies
- **name** — Policy name
- **autoscaling_group_name** — Target Auto Scaling group
- **policy_type** — Policy type (TargetTrackingScaling, StepScaling, SimpleScaling)
- **adjustment_type** — Adjustment type (ChangeInCapacity, PercentChangeInCapacity, ExactCapacity)
- **scaling_adjustment** — Scaling adjustment value
- **cooldown** — Cooldown period (seconds)
- **target_tracking_configuration** — Target tracking settings
- **step_adjustments** — Step scaling adjustments
- **estimated_instance_warmup** — Instance warmup time

### Scheduled Actions
- **scheduled_action_name** — Schedule name
- **autoscaling_group_name** — Target Auto Scaling group
- **min_size** — Scheduled minimum size
- **max_size** — Scheduled maximum size
- **desired_capacity** — Scheduled desired capacity
- **start_time** — Start time (one-time)
- **end_time** — End time
- **recurrence** — Cron expression (recurring)
- **time_zone** — Time zone

## Falco Rule Examples

```yaml
# Auto Scaling Group Lifecycle
- rule: Auto Scaling Group Created
  desc: Detect when an Auto Scaling group is created
  condition: >
    ct.name="CreateAutoScalingGroup"
  output: >
    Auto Scaling group created
    (user=%ct.user asg=%ct.request.autoScalingGroupName min=%ct.request.minSize max=%ct.request.maxSize
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, compute]

- rule: Auto Scaling Group Deleted
  desc: Detect when an Auto Scaling group is deleted
  condition: >
    ct.name="DeleteAutoScalingGroup"
  output: >
    Auto Scaling group deleted
    (user=%ct.user asg=%ct.request.autoScalingGroupName
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, compute, security]

# Configuration Changes
- rule: Auto Scaling Group Updated
  desc: Detect when Auto Scaling group configuration is updated
  condition: >
    ct.name="UpdateAutoScalingGroup"
  output: >
    Auto Scaling group updated
    (user=%ct.user asg=%ct.request.autoScalingGroupName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, configuration]

# Manual Scaling
- rule: Auto Scaling Desired Capacity Changed
  desc: Detect when desired capacity is manually set
  condition: >
    ct.name="SetDesiredCapacity"
  output: >
    Auto Scaling desired capacity changed
    (user=%ct.user asg=%ct.request.autoScalingGroupName capacity=%ct.request.desiredCapacity
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, scaling]

# Scaling Policy Management
- rule: Auto Scaling Policy Created
  desc: Detect when a scaling policy is created or updated
  condition: >
    ct.name="PutScalingPolicy"
  output: >
    Auto Scaling policy created/updated
    (user=%ct.user policy=%ct.request.policyName asg=%ct.request.autoScalingGroupName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, policy]

# Scheduled Actions
- rule: Auto Scaling Scheduled Action Created
  desc: Detect when a scheduled action is created or updated
  condition: >
    ct.name="PutScheduledUpdateGroupAction"
  output: >
    Auto Scaling scheduled action created/updated
    (user=%ct.user action=%ct.request.scheduledActionName asg=%ct.request.autoScalingGroupName
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, autoscaling, schedule]
```

## Example Drift Scenarios

### Scenario 1: Desired Capacity Manually Changed

**CloudTrail Event:**
```json
{
  "eventName": "SetDesiredCapacity",
  "requestParameters": {
    "autoScalingGroupName": "web-asg",
    "desiredCapacity": 10,
    "honorCooldown": false
  },
  "userIdentity": {
    "principalId": "AIDAI23ABCD4EFGH5IJKL",
    "userName": "ops-engineer"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_autoscaling_group.web_asg
Changed: desired_capacity = 5 → 10
User: ops-engineer (IAM User)
Region: us-east-1
Severity: MEDIUM
```

### Scenario 2: Auto Scaling Group Deleted

**CloudTrail Event:**
```json
{
  "eventName": "DeleteAutoScalingGroup",
  "requestParameters": {
    "autoScalingGroupName": "prod-asg",
    "forceDelete": true
  },
  "userIdentity": {
    "type": "AssumedRole",
    "principalId": "AROAI23ABCD4EFGH5IJKL:admin"
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_autoscaling_group.prod_asg
DELETED: Auto Scaling group removed with force
User: admin (Assumed Role)
Region: us-east-1
Severity: CRITICAL
```

### Scenario 3: Scaling Policy Modified

**CloudTrail Event:**
```json
{
  "eventName": "PutScalingPolicy",
  "requestParameters": {
    "autoScalingGroupName": "api-asg",
    "policyName": "cpu-scale-out",
    "policyType": "TargetTrackingScaling",
    "targetTrackingConfiguration": {
      "predefinedMetricSpecification": {
        "predefinedMetricType": "ASGAverageCPUUtilization"
      },
      "targetValue": 80.0
    }
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_autoscaling_policy.cpu_scale_out
Changed: target_tracking_configuration.target_value = 70.0 → 80.0
User: sre-team (Assumed Role)
Region: us-east-1
Severity: MEDIUM
```

### Scenario 4: Scheduled Action Added

**CloudTrail Event:**
```json
{
  "eventName": "PutScheduledUpdateGroupAction",
  "requestParameters": {
    "autoScalingGroupName": "batch-asg",
    "scheduledActionName": "scale-up-morning",
    "recurrence": "0 8 * * MON-FRI",
    "minSize": 10,
    "maxSize": 20,
    "desiredCapacity": 15
  }
}
```

**TFDrift-Falco Alert:**
```
🚨 Drift Detected: aws_autoscaling_schedule (unmanaged)
NEW: Scheduled action created outside Terraform
name: scale-up-morning
recurrence: 0 8 * * MON-FRI
User: developer@example.com (Console)
Region: us-east-1
Severity: LOW
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "Auto Scaling Group Configuration"
    resource_types:
      - "aws_autoscaling_group"
    watched_attributes:
      - "min_size"
      - "max_size"
      - "desired_capacity"
      - "launch_configuration"
      - "launch_template"
      - "vpc_zone_identifier"
    severity: "critical"

  - name: "Scaling Policies"
    resource_types:
      - "aws_autoscaling_policy"
    watched_attributes:
      - "policy_type"
      - "target_tracking_configuration"
      - "adjustment_type"
      - "scaling_adjustment"
    severity: "high"

  - name: "Scheduled Actions"
    resource_types:
      - "aws_autoscaling_schedule"
    watched_attributes:
      - "min_size"
      - "max_size"
      - "desired_capacity"
      - "recurrence"
    severity: "medium"
```

## Best Practices

### 1. Auto Scaling Group with Target Tracking
```hcl
# Terraform - ASG with CPU-based target tracking
resource "aws_autoscaling_group" "web" {
  name                = "web-asg"
  launch_configuration = aws_launch_configuration.web.name
  min_size            = 2
  max_size            = 10
  desired_capacity    = 3
  health_check_type   = "ELB"
  health_check_grace_period = 300
  vpc_zone_identifier = aws_subnet.private[*].id
  target_group_arns   = [aws_lb_target_group.web.arn]

  # Termination policies
  termination_policies = [
    "OldestLaunchConfiguration",
    "Default"
  ]

  # Metrics
  enabled_metrics = [
    "GroupDesiredCapacity",
    "GroupInServiceInstances",
    "GroupMinSize",
    "GroupMaxSize"
  ]

  # Important: Ignore desired_capacity for policy-based scaling
  lifecycle {
    ignore_changes = [desired_capacity]
  }

  tag {
    key                 = "Name"
    value               = "web-server"
    propagate_at_launch = true
  }
}

# Target tracking scaling policy
resource "aws_autoscaling_policy" "cpu_tracking" {
  name                   = "cpu-target-tracking"
  autoscaling_group_name = aws_autoscaling_group.web.name
  policy_type            = "TargetTrackingScaling"

  target_tracking_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ASGAverageCPUUtilization"
    }
    target_value = 70.0
  }
}
```

### 2. Launch Configuration
```hcl
# Terraform - Launch configuration for ASG
resource "aws_launch_configuration" "web" {
  name_prefix          = "web-"
  image_id             = data.aws_ami.ubuntu.id
  instance_type        = "t3.medium"
  key_name             = var.key_name
  security_groups      = [aws_security_group.web.id]
  iam_instance_profile = aws_iam_instance_profile.web.name

  # User data
  user_data = base64encode(templatefile("${path.module}/user_data.sh", {
    environment = "production"
  }))

  # EBS optimization
  ebs_optimized = true

  # Enable detailed monitoring
  enable_monitoring = true

  # Root volume
  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    encrypted             = true
    delete_on_termination = true
  }

  # Create new before destroying old
  lifecycle {
    create_before_destroy = true
  }
}
```

### 3. Step Scaling Policy
```hcl
# Terraform - Step scaling for high traffic
resource "aws_autoscaling_policy" "scale_out" {
  name                   = "scale-out-step"
  autoscaling_group_name = aws_autoscaling_group.api.name
  policy_type            = "StepScaling"
  adjustment_type        = "PercentChangeInCapacity"

  # Warmup time
  estimated_instance_warmup = 300

  step_adjustment {
    metric_interval_lower_bound = 0
    metric_interval_upper_bound = 10
    scaling_adjustment          = 10
  }

  step_adjustment {
    metric_interval_lower_bound = 10
    scaling_adjustment          = 20
  }
}

# CloudWatch alarm to trigger scaling
resource "aws_cloudwatch_metric_alarm" "high_cpu" {
  alarm_name          = "api-high-cpu"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    AutoScalingGroupName = aws_autoscaling_group.api.name
  }

  alarm_actions = [aws_autoscaling_policy.scale_out.arn]
}
```

### 4. Scheduled Scaling
```hcl
# Terraform - Scale up for business hours
resource "aws_autoscaling_schedule" "scale_up_morning" {
  scheduled_action_name  = "scale-up-morning"
  autoscaling_group_name = aws_autoscaling_group.batch.name
  min_size               = 5
  max_size               = 20
  desired_capacity       = 10
  recurrence             = "0 8 * * MON-FRI"
  time_zone              = "America/New_York"
}

# Scale down for off-hours
resource "aws_autoscaling_schedule" "scale_down_evening" {
  scheduled_action_name  = "scale-down-evening"
  autoscaling_group_name = aws_autoscaling_group.batch.name
  min_size               = 2
  max_size               = 5
  desired_capacity       = 2
  recurrence             = "0 18 * * MON-FRI"
  time_zone              = "America/New_York"
}
```

## Security Considerations

### 1. Launch Configuration Security
- Use encrypted EBS volumes
- Apply least-privilege IAM instance profiles
- Avoid embedding secrets in user_data
- Use AWS Secrets Manager or SSM Parameter Store
- Monitor image_id changes for AMI drift

### 2. Network Security
- Deploy ASG instances in private subnets
- Use security groups to restrict access
- Monitor vpc_zone_identifier changes
- Track security group associations

### 3. Scaling Policies
- Set appropriate min/max boundaries
- Monitor SetDesiredCapacity for manual overrides
- Use target tracking for predictable workloads
- Track cooldown period changes

### 4. Access Control
- Restrict who can modify ASG configurations
- Monitor UpdateAutoScalingGroup events
- Alert on DeleteAutoScalingGroup operations
- Track suspended_processes changes

## Known Limitations

### 1. Desired Capacity Management
- SetDesiredCapacity is normal during auto-scaling
- Use lifecycle.ignore_changes for desired_capacity
- Distinguish between manual and automatic changes
- Policy-based scaling triggers frequent updates

### 2. Launch Template vs Launch Configuration
- Launch templates not tracked in current implementation
- Only launch configurations have dedicated events
- Launch template changes tracked via UpdateAutoScalingGroup
- Consider separate monitoring for launch templates

### 3. Mixed Instances Policy
- Mixed instances policy changes tracked via UpdateAutoScalingGroup
- Detailed instance type changes not granularly tracked
- Use Terraform state comparison for detailed drift

### 4. Lifecycle Hooks
- Lifecycle hook changes not tracked in current implementation
- Planned for future enhancement
- Monitor via UpdateAutoScalingGroup events

## Related Documentation

- [AWS Auto Scaling CloudTrail Logging](https://docs.aws.amazon.com/autoscaling/ec2/userguide/logging-using-cloudtrail.html)
- [Terraform aws_autoscaling_group](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_group)
- [Terraform aws_launch_configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/launch_configuration)
- [Terraform aws_autoscaling_policy](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_policy)
- [Terraform aws_autoscaling_schedule](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/autoscaling_schedule)
- [Terraform Auto Scaling Tutorial](https://developer.hashicorp.com/terraform/tutorials/aws/aws-asg)

## Version History

- **v0.3.0** (2025 Q1) - Initial Auto Scaling support with 10 CloudTrail events
  - Auto Scaling group management (create, delete, update, set capacity)
  - Launch configuration management (create, delete)
  - Scaling policy management (put, delete)
  - Scheduled action management (put, delete)
  - Comprehensive drift detection for auto-scaling infrastructure
