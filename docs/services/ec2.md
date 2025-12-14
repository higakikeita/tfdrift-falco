# EC2 (Elastic Compute Cloud) — Drift Coverage

## Overview

TFDrift-Falco monitors Amazon EC2 for configuration drift by tracking CloudTrail events related to instances, AMIs, EBS volumes, snapshots, and network interfaces. This enables real-time detection of manual changes made outside of Terraform workflows.

## Supported CloudTrail Events

### Instance Management (5 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| RunInstances | EC2 instance created | CRITICAL | ✔ |
| TerminateInstances | EC2 instance terminated | CRITICAL | ✔ |
| StartInstances | EC2 instance started | WARNING | ✔ |
| StopInstances | EC2 instance stopped | WARNING | ✔ |
| ModifyInstanceAttribute | Instance attribute modified | WARNING | ✔ |

### AMI Management (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateImage | AMI created from instance | WARNING | ✔ |
| DeregisterImage | AMI deregistered | WARNING | ✔ |

### EBS Volume Management (5 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateVolume | EBS volume created | WARNING | ✔ |
| DeleteVolume | EBS volume deleted | WARNING | ✔ |
| AttachVolume | Volume attached to instance | WARNING | ✔ |
| DetachVolume | Volume detached from instance | WARNING | ✔ |
| ModifyVolume | Volume configuration modified | WARNING | ✔ |

### Snapshot Management (2 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateSnapshot | EBS snapshot created | WARNING | ✔ |
| DeleteSnapshot | EBS snapshot deleted | WARNING | ✔ |

### Network Interface Management (3 events)
| Event | Description | Priority | Status |
|-------|-------------|----------|--------|
| CreateNetworkInterface | Network interface (ENI) created | WARNING | ✔ |
| DeleteNetworkInterface | Network interface deleted | WARNING | ✔ |
| AttachNetworkInterface | Network interface attached | WARNING | ✔ |

**Total: 17 CloudTrail events**

## Supported Terraform Resources

- `aws_instance` — EC2 instance configuration
- `aws_ami` — Amazon Machine Images
- `aws_ebs_volume` — EBS volume configuration
- `aws_volume_attachment` — EBS volume attachments
- `aws_ebs_snapshot` — EBS snapshots
- `aws_network_interface` — Elastic Network Interfaces (ENI)
- `aws_network_interface_attachment` — ENI attachments

## Monitored Drift Attributes

### EC2 Instances
- **instance_type** — Instance size (t3.micro, m5.large, etc.)
- **ami** — Amazon Machine Image ID
- **availability_zone** — AZ placement
- **subnet_id** — VPC subnet
- **vpc_security_group_ids** — Security groups
- **iam_instance_profile** — IAM role
- **user_data** — Initialization script
- **monitoring** — Detailed CloudWatch monitoring
- **ebs_optimized** — EBS optimization flag
- **root_block_device** — Root volume configuration
- **ebs_block_device** — Additional EBS volumes
- **network_interface** — Network interface configuration
- **tags** — Resource tags

### AMIs
- **name** — AMI name
- **description** — AMI description
- **virtualization_type** — HVM or paravirtual
- **root_device_name** — Root device identifier
- **ebs_block_device** — EBS snapshot mappings
- **tags** — Resource tags

### EBS Volumes
- **availability_zone** — AZ placement
- **size** — Volume size (GB)
- **type** — Volume type (gp3, io2, st1, sc1)
- **iops** — Provisioned IOPS
- **throughput** — Throughput (MB/s) for gp3
- **encrypted** — Encryption status
- **kms_key_id** — KMS key for encryption
- **snapshot_id** — Source snapshot
- **tags** — Resource tags

### Volume Attachments
- **device_name** — Device name (/dev/sdf, /dev/sdg, etc.)
- **instance_id** — Attached instance
- **volume_id** — EBS volume ID
- **force_detach** — Force detachment flag
- **skip_destroy** — Skip on destroy

### EBS Snapshots
- **volume_id** — Source volume
- **description** — Snapshot description
- **storage_tier** — Standard or archive
- **tags** — Resource tags

### Network Interfaces
- **subnet_id** — VPC subnet
- **private_ips** — Private IP addresses
- **security_groups** — Security group IDs
- **source_dest_check** — Source/destination check
- **attachment** — Attachment configuration
- **tags** — Resource tags

## Falco Rule Examples

```yaml
# Instance Lifecycle
- rule: EC2 Instance Created
  desc: Detect when an EC2 instance is launched
  condition: >
    ct.name="RunInstances"
  output: >
    EC2 instance created
    (user=%ct.user ami=%ct.request.imageId instance_type=%ct.request.instanceType
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, ec2, compute]

- rule: EC2 Instance Terminated
  desc: Detect when an EC2 instance is terminated
  condition: >
    ct.name="TerminateInstances"
  output: >
    EC2 instance terminated
    (user=%ct.user instance=%ct.request.instancesSet.items.instanceId
     region=%ct.region account=%ct.account)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, ec2, compute, security]

# Volume Management
- rule: EBS Volume Created
  desc: Detect when an EBS volume is created
  condition: >
    ct.name="CreateVolume"
  output: >
    EBS volume created
    (user=%ct.user size=%ct.request.size type=%ct.request.volumeType az=%ct.request.availabilityZone
     region=%ct.region account=%ct.account)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, ec2, ebs, storage]
```

## Configuration Example

```yaml
# config.yaml
drift_rules:
  - name: "EC2 Instance Configuration"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
      - "ami"
      - "vpc_security_group_ids"
      - "iam_instance_profile"
    severity: "critical"

  - name: "EBS Volume Management"
    resource_types:
      - "aws_ebs_volume"
      - "aws_volume_attachment"
    watched_attributes:
      - "size"
      - "type"
      - "encrypted"
    severity: "high"
```

## Best Practices

### 1. Instance Configuration Management
```hcl
# Terraform - Immutable infrastructure pattern
resource "aws_instance" "app" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = var.instance_type

  subnet_id                   = aws_subnet.private.id
  vpc_security_group_ids      = [aws_security_group.app.id]
  iam_instance_profile        = aws_iam_instance_profile.app.name

  monitoring = true

  root_block_device {
    volume_type           = "gp3"
    volume_size           = 20
    encrypted             = true
    delete_on_termination = true
  }

  lifecycle {
    create_before_destroy = true
  }

  tags = {
    Name        = "app-server"
    Environment = "production"
  }
}
```

### 2. EBS Volume Management
```hcl
resource "aws_ebs_volume" "data" {
  availability_zone = aws_instance.app.availability_zone
  size              = 100
  type              = "gp3"
  encrypted         = true
  kms_key_id        = aws_kms_key.ebs.arn

  tags = {
    Name   = "app-data"
    Backup = "daily"
  }
}

resource "aws_volume_attachment" "data" {
  device_name = "/dev/sdf"
  volume_id   = aws_ebs_volume.data.id
  instance_id = aws_instance.app.id
}
```

## Known Limitations

### 1. Instance State vs Configuration
- StartInstances/StopInstances track state changes, not configuration
- Use lifecycle.ignore_changes for instance state with auto-scaling

### 2. AMI Snapshot References
- CreateImage creates implicit EBS snapshots
- AMI deregistration doesn't auto-delete underlying snapshots

### 3. Network Interface Secondary IPs
- Secondary private IP assignments not tracked via CloudTrail management events

## Related Documentation

- [AWS EC2 CloudTrail Events](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/monitor-with-cloudtrail.html)
- [Terraform aws_instance](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/instance)
- [Terraform aws_ebs_volume](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ebs_volume)
- [Terraform aws_ebs_snapshot](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/ebs_snapshot)

## Version History

- **v0.2.0-beta** — Initial EC2 coverage (2 events)
- **v0.3.0** (2025 Q1) — EC2 Enhanced support with 17 CloudTrail events
  - Instance management (run, terminate, start, stop, modify)
  - AMI management (create, deregister)
  - EBS volume management (create, delete, attach, detach, modify)
  - Snapshot management (create, delete)
  - Network interface management (create, delete, attach)
