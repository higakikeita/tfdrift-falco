# AWS Service Coverage Roadmap

## Current Coverage Summary

TFDrift-Falco currently supports monitoring **40+ AWS services** with **500+ CloudTrail events** tracked and integrated into Terraform state drift detection.

### Coverage Statistics

- **Total Supported Services**: 40+
- **Total CloudTrail Events Tracked**: 500+
- **Event Processing Performance**: Sub-second detection latency
- **Multi-Region Support**: Yes (all AWS regions)
- **Organizations Support**: Yes (cross-account monitoring)

## Current Supported Services

### Compute & Containers

- **EC2** - Instances, volumes, security groups, auto-scaling
- **Lambda** - Functions, versions, layers, permissions
- **ECS** - Task definitions, services, clusters
- **EKS** - Kubernetes cluster management
- **Elastic Beanstalk** - Application deployments

### Storage & Databases

- **S3** - Buckets, policies, versioning, encryption
- **EBS** - Volumes, snapshots, encryption
- **RDS** - Database instances, parameter groups, option groups
- **DynamoDB** - Tables, backups, streams
- **Elasticache** - Cache clusters, parameter groups
- **Glacier** - Vaults and policies

### Networking

- **VPC** - Virtual Private Clouds, subnets, routing
- **ELB/ALB/NLB** - Load balancers and target groups
- **CloudFront** - CDN distributions and origins
- **Route 53** - DNS zones and records
- **VPN** - Virtual private gateways
- **Transit Gateway** - Multi-region connectivity

### Security & Identity

- **IAM** - Users, roles, policies, MFA devices (comprehensive)
- **Secrets Manager** - Secret storage and rotation
- **Systems Manager** - Parameter store and automation
- **ACM** - SSL/TLS certificates
- **KMS** - Key management and encryption
- **GuardDuty** - Threat detection integration

### Application Integration

- **SNS** - Simple Notification Service topics and subscriptions
- **SQS** - Message queues and queue policies
- **EventBridge** - Event rules and targets
- **API Gateway** - REST and HTTP APIs
- **AppConfig** - Application configuration
- **Appflow** - Data flows and integration

### Management & Monitoring

- **CloudWatch** - Log groups, alarms, and dashboards
- **CloudTrail** - Audit logging and trails
- **Config** - Configuration compliance
- **Health** - AWS Health events
- **Systems Manager** - OpsCenter and automations

## Monitoring Capabilities by Service

### Event Categories Tracked

1. **Creation Events** - New resource provisioning
2. **Modification Events** - Configuration changes
3. **Deletion Events** - Resource teardown
4. **Permission Changes** - Access control modifications
5. **Encryption Changes** - Security posture updates
6. **Compliance Changes** - Policy and compliance modifications

### Detection Latency

- **CloudTrail Ingestion**: < 5 minutes typical
- **TFDrift-Falco Processing**: < 1 second
- **Alert Delivery**: Real-time via WebSocket/SSE

## Expansion Roadmap

### Phase 1: Q2 2026 (Near-term)

- **Expanded EC2 Coverage**
  - Network interfaces and ENI attachments
  - Placement groups
  - Dedicated hosts
  - Capacity reservations

- **Enhanced RDS**
  - Database proxy support
  - Global database monitoring
  - Custom endpoint tracking

- **Additional Storage Services**
  - EFS (Elastic File System)
  - FSx (Windows/Lustre file systems)
  - DataSync tracking

### Phase 2: Q3 2026 (Medium-term)

- **Advanced Networking**
  - VPC Flow Logs integration
  - Network ACL monitoring
  - Service endpoints
  - Customer gateways

- **Container Services Enhancement**
  - ECR (Elastic Container Registry) policies
  - App Mesh integration
  - Copilot application tracking

- **Message Queue Services**
  - Enhanced SQS monitoring
  - SNS subscription validation
  - Message transformation tracking

### Phase 3: Q4 2026 (Long-term)

- **ML/AI Services**
  - SageMaker endpoints and models
  - Rekognition custom labels
  - Lookout service integration

- **Analytics Services**
  - Redshift cluster monitoring
  - EMR cluster tracking
  - Kinesis stream management

- **Business Applications**
  - Workspaces desktop tracking
  - QuickSight dashboard monitoring
  - Connect contact center management

## Implementation Priorities

### Critical (P0)
- IAM and access control changes (Security)
- Encryption and KMS updates (Security)
- S3 bucket policies and public access (Security)
- Network security groups (Security)

### High (P1)
- Database configuration changes
- Load balancer modifications
- VPC and network topology changes
- Storage service updates

### Medium (P2)
- Application integration service changes
- Management and monitoring updates
- Auto-scaling configuration changes
- Cache service modifications

### Low (P3)
- Non-critical service updates
- Preview/beta service support
- Deprecated service maintenance

## Configuration for AWS Coverage

### Basic Setup

```yaml
# config.yaml
cloud_providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
      - eu-west-1
    services:
      enabled: "all"  # or specify: [ec2, s3, iam, rds, ...]
```

### Advanced Filtering

```yaml
cloud_providers:
  aws:
    services:
      ec2:
        enabled: true
        events:
          - ModifyInstanceAttribute
          - ModifySecurityGroup
      s3:
        enabled: true
        events: "all"
```

## Service Coverage Reference

For detailed coverage information per service, see:
- [AWS Resource Coverage Analysis](AWS_RESOURCE_COVERAGE_ANALYSIS.md)
- [Complete Setup Guide](complete-setup-guide.md)
- [CloudTrail Event Reference](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-log-file-examples.html)

## Contributing to Coverage Expansion

To request support for additional AWS services:

1. Open an issue with the service name and use case
2. Include CloudTrail event references
3. Provide example Terraform resource definitions
4. See [CONTRIBUTING.md](../docs/CONTRIBUTING.md) for more details

## Feedback & Support

For questions about AWS service coverage or to report gaps:
- Create an issue on the project repository
- Refer to the [Use Cases documentation](use-cases.md)
- Check [Troubleshooting guide](best-practices.md#troubleshooting)
