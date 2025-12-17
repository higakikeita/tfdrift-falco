# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.5.0] - 2025-12-17

### 🎉 Major Release - Multi-Cloud Support (GCP)

This release brings comprehensive Google Cloud Platform (GCP) support to TFDrift-Falco, enabling real-time drift detection across both AWS and GCP environments simultaneously.

### Added

#### GCP Audit Logs Integration
- **GCP Audit Parser** (`pkg/gcp/audit_parser.go`)
  - Full parsing of GCP Audit Log events from Falco gcpaudit plugin
  - Extraction of resource details (project ID, zone, region)
  - User identity correlation (principal email, service accounts)
  - Change tracking with request/response capture
  - Comprehensive validation and error handling

- **GCP Resource Mapper** (`pkg/gcp/resource_mapper.go`)
  - 100+ event-to-Terraform-resource mappings
  - Coverage across 12+ GCP services:
    - Compute Engine (30+ events): Instances, Disks, Machine Types, Metadata, Networks, Firewalls
    - Cloud Storage (15+ events): Buckets, Objects, IAM Bindings, ACLs, Lifecycle
    - Cloud SQL (10+ events): Instances, Databases, Users, Backups
    - GKE (10+ events): Clusters, Node Pools, Workloads
    - Cloud Run (8+ events): Services, Revisions, IAM Policies
    - IAM (8+ events): Service Accounts, Roles, Bindings, Keys
    - VPC/Networking (10+ events): Firewalls, Routes, Subnets, Peering
    - Cloud Functions (5+ events): Functions, Triggers, IAM Policies
    - BigQuery (5+ events): Datasets, Tables, IAM Policies
    - Pub/Sub (5+ events): Topics, Subscriptions, IAM Policies
    - KMS (5+ events): Keys, KeyRings, IAM Policies
    - Secret Manager (3+ events): Secrets, Versions, IAM Policies
  - Intelligent action detection (create, update, delete, setIamPolicy)
  - Service-based resource type inference

#### GCS Backend Support
- **Google Cloud Storage Backend** (`pkg/terraform/backend/gcs.go`)
  - Load Terraform state from GCS buckets
  - Application Default Credentials (ADC) support
  - Custom credentials file support
  - Bucket and prefix configuration
  - Comprehensive error handling

#### Multi-Provider Architecture
- **Event Router** - Source-based routing in `parseFalcoOutput()`
  - `aws_cloudtrail` → AWS parser
  - `gcpaudit` → GCP parser
  - Extensible design for future providers (Azure, etc.)

- **Extended Event Type** (`pkg/types/types.go`)
  - GCP-specific fields: `ProjectID`, `ServiceName`
  - Preserved AWS-specific fields: `Region`, `AccountID`
  - Provider-agnostic core fields

#### Testing & Documentation
- **Comprehensive Test Coverage**
  - 34 GCP parser tests covering all functionality
  - Integration tests for multi-provider scenarios
  - Resource type mapping validation
  - All tests passing (100% pass rate)

- **GCP Setup Guide** (`docs/gcp-setup.md`)
  - Step-by-step Falco gcpaudit plugin configuration
  - GCP Audit Logs and Pub/Sub setup
  - TFDrift-Falco configuration examples
  - Troubleshooting guide with 5 common issues
  - Advanced configuration (multi-project, custom rules, regional deployment)
  - Security best practices

- **Example Configuration** (`examples/config-gcp.yaml`)
  - Complete GCP configuration with drift rules
  - Multi-project setup examples
  - GCS backend configuration

### Changed

#### Configuration Updates
- **Extended Provider Config** (`pkg/config/config.go`)
  - New `GCPConfig` structure with projects and state configuration
  - GCS backend fields: `GCSBucket`, `GCSPrefix`
  - Backward compatible with existing AWS configurations

- **Backend Factory** (`pkg/terraform/backend/factory.go`)
  - Added GCS backend case to factory method
  - Supports `backend: "gcs"` in configuration
  - Context propagation for GCS client initialization

#### Falco Integration
- **Subscriber Enhancement** (`pkg/falco/subscriber.go`)
  - Initialized GCP parser in `NewSubscriber()`
  - Multi-provider event processing

- **Event Parser Refactoring** (`pkg/falco/event_parser.go`)
  - Extracted AWS parsing into `parseAWSEvent()` method
  - Added GCP parsing via `gcpParser.Parse()`
  - Clean separation of provider-specific logic

### Dependencies
- Added `cloud.google.com/go/storage` v1.58.0
- Added GCP SDK dependencies for authentication and storage access

### Architecture Improvements
- Multi-provider support without breaking changes
- Interface-based design for future cloud providers
- Comprehensive logging and error handling
- Production-ready GCP integration

### Migration Guide

No breaking changes in this release. To enable GCP support:

1. **Update Configuration**:
   ```yaml
   providers:
     gcp:
       enabled: true
       projects:
         - my-project-123
       state:
         backend: "gcs"
         gcs_bucket: "my-terraform-state"
         gcs_prefix: "prod"
   ```

2. **Setup Falco gcpaudit Plugin**: Follow [GCP Setup Guide](./docs/gcp-setup.md)

3. **Configure GCP Credentials**: Use Application Default Credentials or specify credentials file

### Known Limitations

- GCP support is new - production validation recommended
- Multi-project environments require additional configuration
- GCP Audit Log delivery latency: 30 seconds to 5 minutes (via Pub/Sub)
- Some advanced GCP features may not be fully covered yet

### Contributors

This release brings comprehensive GCP support enabling true multi-cloud drift detection. Special thanks to the community for feature requests and feedback.

---

## [Unreleased]

### Added
- **MkDocs Documentation Site** (PR #3) - Professional project documentation
  - Material for MkDocs theme with dark/light mode toggle
  - Complete documentation structure: Getting Started, Configuration, Development, API
  - AWS service coverage matrix (54+ services documented)
  - Automated link checking in CI
  - GitHub Pages deployment configured

- **Official Docker Image Support** (PR #4) - GHCR container registry integration
  - Multi-architecture builds (linux/amd64, linux/arm64)
  - Automated publishing to `ghcr.io/higakikeita/tfdrift-falco`
  - Version tagging: `latest`, `vX.Y.Z`, `vX.Y`, `vX`
  - SHA-based tags for immutable deployments
  - Updated Dockerfile with optimized Alpine 3.21 base

- **Backend Package Tests** - Comprehensive test coverage for Terraform backend abstraction layer
  - Local filesystem backend: validation, error handling, file operations
  - S3 backend: configuration validation, region defaults
  - Factory pattern tests for backend selection
  - Coverage improved from 0.0% to 67.4%

- **Benchmark Test Suite** - Performance baseline establishment
  - Event processing benchmarks: ~44μs/op, 9.5KB/op, 117 allocs/op (22,000 events/sec capable)
  - State comparison benchmarks: ~4ns/op (cached lookups)
  - Concurrent event handling benchmarks
  - Memory usage tests with leak detection (4 tests)
  - Created test helper methods (HandleEventForTest, GetStateManagerForTest)

- **Load Test Implementation** - Production-scale testing framework
  - TestLoadScenario1_Small: 100 events/min, 500 resources, 1h duration
  - TestLoadScenario2_Medium: 1,000 events/min, 5,000 resources, 4h duration
  - TestLoadScenario3_Large: 10,000 events/min, 50,000 resources, 8h duration
  - TestLoadTest_QuickSmoke: Infrastructure validation (~15s)
  - Automated setup, execution, validation, and cleanup
  - Integration with CloudTrail simulator and Terraform state generator

- **Security Infrastructure**
  - Snyk workflow with proper SARIF output configuration
  - Local security scanning script (`scripts/security-scan.sh`)
  - Comprehensive security policy documentation (`.github/SECURITY.md`)
  - GoSec, Nancy, and govulncheck integration
  - Security scanning section in README

### Security
- **Critical Vulnerability Fixes** (PR #5) - Updated Go stdlib and dependencies
  - Fixed GO-2025-4175: crypto/x509 DNS name constraint verification (CVSS 7.5)
  - Fixed GO-2025-4155: crypto/x509 excessive resource consumption (CVSS 5.3)
  - Updated Go toolchain: 1.23.0 → 1.24.0/1.25.5
  - Updated dependencies: grpc v1.77.0, cobra v1.10.2, viper v1.21.0
  - Updated Docker base images to latest secure versions

### Fixed
- **Test Expectation Updates** (PR #6) - Fixed failing unit tests
  - Fixed pkg/terraform state backend error handling tests
  - Fixed pkg/config validation message case sensitivity
  - Fixed pkg/falco RDS event relevance detection
  - All tests now passing locally

- Fixed Snyk SARIF file generation with `--sarif-file-output` flag
- Fixed benchmark test API compatibility issues (HandleEvent → HandleEventForTest)
- Fixed memory test uint64 underflow by using TotalAlloc instead of Alloc
- Fixed memory leak detection growth calculation

### Documentation
- Added comprehensive MkDocs documentation site with AWS service coverage
- Added `SECURITY_FIXES.md` documenting crypto/x509 vulnerability remediation
- Added `docs/v0.2.0-beta-quality-improvements-diary.md` - Development diary for post-release improvements
- Updated README with Docker usage instructions and GHCR registry
- Updated README with security scanning section
- Added security policy and vulnerability reporting process
- Copied community health files to docs/ for MkDocs integration

## [0.2.0-beta] - 2025-12-05

### 🎉 Major Release - Production Readiness & VPC Support

This release dramatically expands AWS service coverage and adds comprehensive production readiness tooling.

### Added

#### New Service Coverage (+265% event coverage)
- **VPC/Networking (33 events)** - Addresses #1 priority gap
  - Security Groups: AuthorizeSecurityGroupIngress/Egress, RevokeSecurityGroupIngress/Egress, CreateSecurityGroup, DeleteSecurityGroup
  - VPC Core: CreateVpc, DeleteVpc, ModifyVpcAttribute, CreateSubnet, DeleteSubnet, ModifySubnetAttribute
  - Route Tables: CreateRoute, DeleteRoute, ReplaceRoute, CreateRouteTable, DeleteRouteTable, AssociateRouteTable
  - Gateways: AttachInternetGateway, DetachInternetGateway, CreateNatGateway, DeleteNatGateway
  - Network ACLs: CreateNetworkAcl, DeleteNetworkAcl, CreateNetworkAclEntry, DeleteNetworkAclEntry, ReplaceNetworkAclEntry
  - VPC Endpoints: CreateVpcEndpoint, DeleteVpcEndpoint, ModifyVpcEndpoint

- **ELB/ALB (15 events)** - Load balancer drift detection
  - Load Balancers: CreateLoadBalancer, DeleteLoadBalancer, ModifyLoadBalancerAttributes
  - Target Groups: CreateTargetGroup, DeleteTargetGroup, ModifyTargetGroup, RegisterTargets, DeregisterTargets
  - Listeners & Rules: CreateListener, DeleteListener, ModifyListener, CreateRule, DeleteRule, ModifyRule

- **KMS (10 events)** - Encryption key monitoring
  - Key Management: ScheduleKeyDeletion, DisableKey, EnableKey, PutKeyPolicy, CreateKey
  - Aliases: CreateAlias, DeleteAlias, UpdateAlias
  - Rotation: EnableKeyRotation, DisableKeyRotation

- **DynamoDB (5 events)** - NoSQL table monitoring
  - Table Lifecycle: CreateTable, DeleteTable, UpdateTable
  - Features: UpdateTimeToLive, UpdateContinuousBackups

- **S3 Enhancements (3 events)**
  - PutBucketPublicAccessBlock, DeleteBucketPublicAccessBlock, PutBucketAcl

- **Lambda Enhancements (2 events)**
  - AddPermission, RemovePermission

#### Documentation & Analysis
- **Production Readiness Assessment** (`docs/PRODUCTION_READINESS.md`)
  - Known limitations and constraints
  - Multi-account/multi-region considerations
  - CloudTrail latency expectations
  - Security best practices
  - Recommended architectures (small/medium/large scale)
  - Comprehensive troubleshooting guide

- **AWS Resource Coverage Analysis** (`docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md`)
  - Complete service-by-service coverage analysis
  - Priority matrix with scoring
  - Implementation roadmap (Phase 1-3)
  - Gap analysis and recommendations

- **Setup Guides**
  - Corrected Qiita setup guide with actual project structure
  - Zenn article for Japanese community
  - Setup verification report

#### Load Testing Framework
- **CloudTrail Event Simulator** (`tests/load/cloudtrail_simulator.go`)
  - Generates realistic CloudTrail events at configurable rates (100-10,000 events/min)
  - Supports 15 AWS service event types with weighted distribution
  - Hourly log file rotation matching CloudTrail behavior

- **Terraform State Generator** (`tests/load/terraform_state_generator.go`)
  - Generates states with 500 to 50,000 resources
  - 17 AWS resource types with realistic attributes
  - Weighted resource distribution

- **Metrics Collection** (`tests/load/collect_metrics.sh`)
  - Docker container CPU/memory monitoring
  - Prometheus metrics collection
  - Loki event count tracking
  - Automatic summary generation

- **Integrated Test Runner** (`tests/load/run_load_test.sh`)
  - Three scenarios: small (1h), medium (4h), large (8h)
  - Automated setup, execution, and teardown
  - Generates comprehensive performance reports

#### Grafana Enhancements
- **Alert Configuration**
  - 6 pre-configured alert rules (Critical, High, Medium severity)
  - 4 notification channels (Slack, Email, Webhook, Default)
  - Alert setup guide (`dashboards/grafana/ALERTS.md`)

- **Integration Testing**
  - 9 test scenarios covering Docker, services, data ingestion, queries
  - Automated test script (`tests/integration/test_grafana.sh`)
  - Test results documentation

- **Improved Promtail Configuration**
  - JSON pipeline stages for proper label extraction
  - Timestamp parsing
  - Field extraction for severity, resource_type, action

- **User Guides**
  - Comprehensive getting started guide
  - Dashboard customization guide with 15+ query examples
  - Quick-start script for one-command setup

### Changed

#### Event Coverage Statistics
- **Before**: 26 events across 5 services (EC2, IAM, S3, RDS, Lambda)
- **After**: 95 events across 10 services (+ VPC, ELB/ALB, KMS, DynamoDB, enhanced S3/Lambda)
- **Increase**: +265% event coverage

#### Resource Type Mappings
Added 40+ new Terraform resource type mappings:
- `aws_security_group`, `aws_security_group_rule`
- `aws_vpc`, `aws_subnet`, `aws_route`, `aws_route_table`, `aws_route_table_association`
- `aws_nat_gateway`, `aws_internet_gateway_attachment`
- `aws_network_acl`, `aws_network_acl_rule`, `aws_vpc_endpoint`
- `aws_lb`, `aws_lb_target_group`, `aws_lb_listener`, `aws_lb_listener_rule`, `aws_lb_target_group_attachment`
- `aws_kms_key`, `aws_kms_alias`
- `aws_dynamodb_table`
- `aws_s3_bucket_public_access_block`, `aws_s3_bucket_acl`
- `aws_lambda_permission`

### Performance Validation

#### Tested Scenarios
- **Small**: 100 events/min, 500 resources, 1 hour
- **Medium**: 1,000 events/min, 5,000 resources, 4 hours
- **Large**: 10,000 events/min, 50,000 resources, 8 hours

#### Acceptance Criteria
| Metric | Small | Medium | Large |
|--------|-------|--------|-------|
| Event Processing (p95) | < 100ms | < 500ms | < 1s |
| Memory Usage | < 512MB | < 2GB | < 4GB |
| CPU Usage (avg) | < 10% | < 30% | < 50% |
| State Load Time | < 1s | < 5s | < 30s |
| Error Rate | < 0.1% | < 1% | < 5% |

### Known Limitations

#### Scale & Performance
- Large-scale environments (50,000+ resources) not yet validated in production
- Multi-account/multi-region setups require additional validation
- CloudTrail log delivery latency: 5-15 minutes (S3), 1-5 minutes (SQS)

#### Service Coverage
- Still missing: CloudFormation, Route53, CloudFront, API Gateway, ECS/EKS, SNS/SQS, Secrets Manager
- See `docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md` for complete gap analysis

#### Tooling
- Grafana 10.x YAML-based alert provisioning not working (use UI or Terraform)
- Integration tests use sample data; real production validation recommended

### Migration Guide

No breaking changes in this release. Configuration format remains compatible with v0.1.x.

To use new services, ensure your Terraform state includes the corresponding resources.

### Contributors

This release brings comprehensive production readiness analysis and dramatically expanded AWS service coverage based on real-world deployment needs.

---

## [0.1.0] - 2024-11-xx

### Added
- Initial release with core drift detection functionality
- Support for EC2, IAM, S3, RDS, Lambda (basic coverage)
- Falco gRPC integration
- CloudTrail event processing
- Slack notifications
- Basic Grafana dashboards
- Docker Compose deployment

### Features
- Real-time drift detection from CloudTrail events
- Terraform state comparison
- Auto-import command generation
- Multiple Terraform backend support (local, S3, remote)
- Configurable drift rules with severity levels
- JSON structured logging

---

## Release Notes

### v0.2.0-beta Highlights

🎯 **Main Achievement**: Comprehensive production readiness with VPC/Networking support

📊 **Event Coverage**: 26 → 95 events (+265%)

🔒 **Security**: Critical security events now monitored (Security Groups, KMS)

⚡ **Performance**: Load testing framework with 3 validated scenarios

📖 **Documentation**: 10,000+ words of production guidance

🚀 **Next Steps**:
- Run load tests in your environment
- Review production readiness checklist
- Validate multi-account setup if applicable
- Consider implementing missing services based on your needs

For detailed implementation plans, see `docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md`.
