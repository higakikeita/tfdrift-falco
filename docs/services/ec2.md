# EC2 — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| RunInstances | Instance launched | ✔ |
| TerminateInstances | Instance terminated | ✔ |
| StopInstances | Instance stopped | ✔ |
| StartInstances | Instance started | ✔ |
| ModifyInstanceAttribute | Instance attribute modified | ✔ |
| CreateTags | Tags added/updated | ✔ |
| DeleteTags | Tags removed | ✔ |
| ModifyNetworkInterfaceAttribute | Network interface modified | ✔ |

## Monitored Drift Attributes

### Instance Configuration
- instance_type
- ami
- monitoring (detailed CloudWatch)
- user_data
- iam_instance_profile
- source_dest_check
- disable_api_termination
- ebs_optimized

### Network
- vpc_security_group_ids
- subnet_id
- associate_public_ip_address
- private_ip

### Storage
- root_block_device
  - volume_size
  - volume_type
  - encrypted
  - delete_on_termination
- ebs_block_device (additional volumes)

## Falco Rule Examples

```yaml
rule: ec2_instance_type_changed
condition:
  cloud.service = "ec2" and evt.name = "ModifyInstanceAttribute" and
  drift.attribute = "instance_type"
output: "EC2 Instance Type Changed (instance=%resource from=%drift.old_value to=%drift.new_value user=%user)"
priority: warning

rule: ec2_instance_terminated_unplanned
condition:
  cloud.service = "ec2" and evt.name = "TerminateInstances" and
  drift.planned = false
output: "Unplanned EC2 Termination (instance=%resource user=%user)"
priority: error
```

## Example Log Output

```json
{
  "service": "ec2",
  "event": "ModifyInstanceAttribute",
  "resource": "i-0123456789abcdef0",
  "changes": {
    "instance_type": ["t3.micro", "t3.small"],
    "monitoring": [false, true]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- EC2 instance type changes (by instance, user, region)
- Unplanned terminations
- Security group modifications
- Tag changes

### Visualizations
- Timeline of instance lifecycle events
- Heatmap of configuration changes by instance
- Alert panel for critical modifications

## Known Limitations

- Spot instance drift may have CloudTrail delay (up to 15 minutes)
- Auto Scaling group actions tracked separately
- EC2 Fleet drift not fully supported yet (v0.3.0 planned)
- Ephemeral instance store changes not tracked (CloudTrail limitation)

## Release History

- **v0.2.0-beta**: Initial EC2 coverage (8 events)
- **v0.3.0** (planned): Enhanced Auto Scaling integration
