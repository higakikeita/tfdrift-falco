# AWS Service Coverage

> **Version:** v0.2.0+ (v0.5.0 adds GCP support)
> **Events:** 203+ CloudTrail events
> **Services:** 19 AWS services
> **Status:** Production Ready

TFDrift-Falco monitors **203+ CloudTrail events** across **19 AWS services**. This page provides an overview of AWS service coverage.

For GCP service coverage, see [GCP Services →](gcp/index.md).

---

## Coverage Summary

| Service | Events | Status | Documentation |
|---------|--------|--------|---------------|
| EC2 | 8 | ✅ Full | [View →](ec2.md) |
| VPC | 19 | ✅ Full | [View →](vpc.md) |
| S3 | 12 | ✅ Full | [View →](s3.md) |
| RDS/Aurora | 11 | ✅ Full | [View →](rds.md) |
| IAM | 14 | ✅ Full | [View →](iam.md) |
| KMS | 13 | ✅ Full | [View →](kms.md) |
| API Gateway | 9 | ⚠️ Partial | [View →](api-gateway.md) |
| Route53 | 6 | ✅ Full | [View →](route53.md) |
| CloudFront | 6 | ✅ Full | [View →](cloudfront.md) |
| SNS | 8 | ✅ Full | [View →](sns.md) |
| SQS | 7 | ✅ Full | [View →](sqs.md) |
| ECR | 9 | ✅ Full | [View →](ecr.md) |
| **Total** | **122** | | |

---

## Service Categories

### Compute

#### EC2
- Instance lifecycle management
- Instance attribute modifications
- Network interface changes
- Tagging operations

[Full EC2 Coverage →](ec2.md)

---

### Networking

#### VPC
- Security group rules (ingress/egress)
- Route table modifications
- NAT Gateway operations
- Internet Gateway attachments

[Full VPC Coverage →](vpc.md)

#### Route53
- DNS record changes
- Hosted zone management
- VPC associations

[Full Route53 Coverage →](route53.md)

#### CloudFront
- Distribution configuration
- Origin modifications
- Cache policy updates

[Full CloudFront Coverage →](cloudfront.md)

---

### Storage

#### S3
- Bucket configuration
- Encryption settings
- Bucket policies
- Public access blocks
- Lifecycle rules

[Full S3 Coverage →](s3.md)

---

### Databases

#### RDS / Aurora
- Instance and cluster configuration
- Engine version upgrades
- Multi-AZ changes
- Backup settings
- Parameter groups

[Full RDS Coverage →](rds.md)

---

### Security & Identity

#### IAM
- Role trust policies
- Policy attachments/detachments
- Inline policy modifications
- User management

[Full IAM Coverage →](iam.md)

#### KMS
- Key management
- Key policies
- Key rotation
- Key deletion scheduling

[Full KMS Coverage →](kms.md)

---

### Application Services

#### API Gateway
- REST API configuration
- Authorizer management
- Stage deployments
- Throttling settings

[Full API Gateway Coverage →](api-gateway.md)

---

### Messaging

#### SNS
- Topic configuration
- Subscription management
- Topic policies
- Encryption settings

[Full SNS Coverage →](sns.md)

#### SQS
- Queue configuration
- Dead letter queues
- Queue policies
- Encryption settings

[Full SQS Coverage →](sqs.md)

---

### Containers

#### ECR
- Repository management
- Image scanning configuration
- Tag mutability
- Lifecycle policies

[Full ECR Coverage →](ecr.md)

---

## Coverage Status Legend

| Icon | Status | Description |
|------|--------|-------------|
| ✅ | Full | All major events covered |
| ⚠️ | Partial | Core events covered, some advanced features pending |
| 🚧 | In Progress | Under development |
| 📅 | Planned | Planned for future release |

---

## Planned Service Additions (v0.3.0)

The following services are planned for v0.3.0:

- **Lambda**: Function configuration, triggers, environment variables
- **ECS**: Task definitions, services, cluster configuration
- **EKS**: Cluster configuration, node groups, add-ons
- **Step Functions**: State machine definitions
- **WAF**: Web ACL rules, rate limiting
- **CodePipeline**: Pipeline configuration

[View v0.3.0 Release Plan →](../release-notes/v0.3.0.md)

---

## Requesting New Service Coverage

Need coverage for a service not listed here?

1. **Check existing issues**: [GitHub Issues](https://github.com/higakikeita/tfdrift-falco/issues?q=is%3Aissue+label%3Aservice-request)
2. **Open a new request**: [Request Service Coverage](https://github.com/higakikeita/tfdrift-falco/issues/new?template=service-request.md)
3. **Contribute**: See [Contributing Guide](../CONTRIBUTING.md)

---

## Service-Specific Limitations

Each service has known limitations documented in its respective page. Common limitations include:

- **Eventual consistency**: Some AWS services (IAM, Route53) have eventual consistency
- **CloudTrail delays**: Regional events may have 5-15 minute delay
- **Complex attributes**: Nested resource configurations may have partial coverage

Refer to individual service documentation for details.
