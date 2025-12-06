# AWS Service Coverage Roadmap

**Last Updated:** December 6, 2024
**Current Version:** v0.2.0-beta
**Planning Horizon:** Q1 2025 - Q4 2025

---

## 📊 Current Coverage Status

### ✅ Fully Supported (v0.2.0-beta)

| Service | Events | Priority | Status |
|---------|--------|----------|--------|
| VPC/Networking | 33 | Critical | ✅ Production |
| IAM | 14 | Critical | ✅ Production |
| ELB/ALB | 15 | High | ✅ Production |
| KMS | 10 | Critical | ✅ Production |
| S3 | 8 | High | ✅ Production |
| Lambda | 4 | Medium | ⚠️ Basic |
| DynamoDB | 5 | Medium | ✅ Production |
| EC2 | 3 | High | ⚠️ Basic |
| RDS | 2 | High | ⚠️ Basic |
| CloudFront | 6 | Medium | ✅ Production |
| SNS | 4 | Medium | ✅ Production |
| SQS | 3 | Medium | ✅ Production |
| ECR | 9 | Medium | ✅ Production |

**Total:** 116 events across 13 services

---

## 🎯 Prioritization Framework

Services are prioritized based on:

1. **Security Impact** - Critical for security/compliance
2. **Usage Frequency** - Commonly used in production
3. **Drift Risk** - High risk of manual changes
4. **Community Demand** - User requests and votes

**Priority Levels:**
- 🔴 **P0 (Critical):** Security-critical, high usage
- 🟠 **P1 (High):** Common infrastructure, moderate security impact
- 🟡 **P2 (Medium):** Standard services, lower drift risk
- 🟢 **P3 (Low):** Specialized services, niche use cases

---

## 📅 Release Plan

### v0.3.0 (Q1 2025 - Target: March 2025)

**Theme:** Enhanced Container & Compute Coverage

#### 🔴 P0: Critical Additions

**ECS (Elastic Container Service)** - 15 events
- `CreateCluster`, `DeleteCluster`, `UpdateCluster`
- `RegisterTaskDefinition`, `DeregisterTaskDefinition`
- `CreateService`, `DeleteService`, `UpdateService`
- `RunTask`, `StopTask`
- `PutClusterCapacityProviders`
- `UpdateClusterSettings`
- `UpdateServicePrimaryTaskSet`
- `TagResource`, `UntagResource`

**Rationale:** ECS is the primary container orchestration service for AWS. Drift in task definitions or service configurations can cause deployment failures.

**EKS (Elastic Kubernetes Service)** - 12 events
- `CreateCluster`, `DeleteCluster`, `UpdateClusterConfig`
- `CreateNodegroup`, `DeleteNodegroup`, `UpdateNodegroupConfig`
- `CreateAddon`, `DeleteAddon`, `UpdateAddon`
- `AssociateEncryptionConfig`
- `TagResource`, `UntagResource`

**Rationale:** Critical for Kubernetes workloads. Node group and addon drift can impact cluster stability.

#### 🟠 P1: High Priority

**Lambda (Enhanced)** - +10 events (total 14)
- `CreateEventSourceMapping`, `DeleteEventSourceMapping`, `UpdateEventSourceMapping`
- `PutFunctionConcurrency`, `DeleteFunctionConcurrency`
- `CreateAlias`, `DeleteAlias`, `UpdateAlias`
- `PublishLayerVersion`, `DeleteLayerVersion`
- `PutFunctionEventInvokeConfig`

**Rationale:** Lambda is heavily used for serverless. Enhanced coverage for event sources and concurrency.

**EC2 (Enhanced)** - +15 events (total 18)
- `CreateVolume`, `DeleteVolume`, `ModifyVolume`
- `AttachVolume`, `DetachVolume`
- `CreateSnapshot`, `DeleteSnapshot`
- `CreateImage`, `DeregisterImage`
- `ModifyImageAttribute`
- `AssociateIamInstanceProfile`, `DisassociateIamInstanceProfile`
- `ModifyInstanceCapacityReservationAttributes`
- `ModifyInstanceCreditSpecification`
- `ModifyInstancePlacement`

**Rationale:** EC2 is fundamental infrastructure. EBS and AMI management are critical.

**RDS (Enhanced)** - +8 events (total 10)
- `CreateDBParameterGroup`, `DeleteDBParameterGroup`, `ModifyDBParameterGroup`
- `CreateDBSnapshot`, `DeleteDBSnapshot`, `RestoreDBInstanceFromDBSnapshot`
- `PromoteReadReplica`
- `RebootDBInstance`

**Rationale:** Database drift is high-risk. Parameter groups affect performance and behavior.

#### 🟡 P2: Medium Priority

**ElastiCache** - 12 events
- `CreateCacheCluster`, `DeleteCacheCluster`, `ModifyCacheCluster`
- `CreateReplicationGroup`, `DeleteReplicationGroup`, `ModifyReplicationGroup`
- `CreateCacheParameterGroup`, `DeleteCacheParameterGroup`, `ModifyCacheParameterGroup`
- `CreateCacheSubnetGroup`, `DeleteCacheSubnetGroup`, `ModifyCacheSubnetGroup`

**Rationale:** Common for caching layers. Parameter drift affects performance.

**Auto Scaling** - 10 events
- `CreateAutoScalingGroup`, `DeleteAutoScalingGroup`, `UpdateAutoScalingGroup`
- `PutScalingPolicy`, `DeleteScalingPolicy`
- `PutScheduledUpdateGroupAction`, `DeleteScheduledAction`
- `AttachInstances`, `DetachInstances`
- `SetDesiredCapacity`

**Rationale:** Auto Scaling drift can cause cost issues and availability problems.

**Estimated Total for v0.3.0:** +82 events (198 total)

---

### v0.4.0 (Q2 2025 - Target: June 2025)

**Theme:** Application & Integration Services

#### 🔴 P0: Critical Additions

**Secrets Manager** - 8 events
- `CreateSecret`, `DeleteSecret`, `UpdateSecret`
- `PutSecretValue`, `DeleteResourcePolicy`, `PutResourcePolicy`
- `RotateSecret`, `CancelRotateSecret`

**Rationale:** Critical for security. Secret drift can cause authentication failures.

**Systems Manager (Parameter Store)** - 8 events
- `PutParameter`, `DeleteParameter`, `DeleteParameters`
- `AddTagsToResource`, `RemoveTagsFromResource`
- `LabelParameterVersion`, `UpdateMaintenanceWindow`
- `RegisterPatchBaselineForPatchGroup`

**Rationale:** Widely used for configuration management. Parameter drift common.

#### 🟠 P1: High Priority

**Step Functions** - 8 events
- `CreateStateMachine`, `DeleteStateMachine`, `UpdateStateMachine`
- `CreateActivity`, `DeleteActivity`
- `TagResource`, `UntagResource`
- `UpdateMapRun`

**Rationale:** Critical for workflow orchestration. State machine changes affect business logic.

**EventBridge** - 10 events
- `PutRule`, `DeleteRule`, `DisableRule`, `EnableRule`
- `PutTargets`, `RemoveTargets`
- `PutEvents`, `PutPartnerEvents`
- `CreateEventBus`, `DeleteEventBus`

**Rationale:** Central to event-driven architectures. Rule drift affects automation.

**API Gateway (Enhanced)** - +12 events (total 21)
- `CreateDomainName`, `DeleteDomainName`, `UpdateDomainName`
- `CreateBasePathMapping`, `DeleteBasePathMapping`, `UpdateBasePathMapping`
- `CreateUsagePlan`, `DeleteUsagePlan`, `UpdateUsagePlan`
- `CreateApiKey`, `DeleteApiKey`, `UpdateApiKey`

**Rationale:** API management is critical. Usage plans and keys affect rate limiting.

#### 🟡 P2: Medium Priority

**CloudWatch** - 12 events
- `PutMetricAlarm`, `DeleteAlarms`
- `PutDashboard`, `DeleteDashboards`
- `CreateLogGroup`, `DeleteLogGroup`
- `PutRetentionPolicy`, `DeleteRetentionPolicy`
- `PutMetricFilter`, `DeleteMetricFilter`
- `PutSubscriptionFilter`, `DeleteSubscriptionFilter`

**Rationale:** Monitoring infrastructure. Alarm drift affects incident response.

**Glue** - 10 events
- `CreateDatabase`, `DeleteDatabase`, `UpdateDatabase`
- `CreateTable`, `DeleteTable`, `UpdateTable`
- `CreateJob`, `DeleteJob`, `UpdateJob`
- `StartCrawler`

**Rationale:** ETL infrastructure. Schema and job drift affect data pipelines.

**Estimated Total for v0.4.0:** +68 events (266 total)

---

### v0.5.0 (Q3 2025 - Target: September 2025)

**Theme:** Security & Compliance Services

#### 🔴 P0: Critical Additions

**WAF (Web Application Firewall)** - 15 events
- `CreateWebACL`, `DeleteWebACL`, `UpdateWebACL`
- `CreateRuleGroup`, `DeleteRuleGroup`, `UpdateRuleGroup`
- `CreateIPSet`, `DeleteIPSet`, `UpdateIPSet`
- `CreateRegexPatternSet`, `DeleteRegexPatternSet`, `UpdateRegexPatternSet`
- `AssociateWebACL`, `DisassociateWebACL`
- `PutLoggingConfiguration`, `DeleteLoggingConfiguration`

**Rationale:** Critical security service. Rule drift can expose vulnerabilities.

**GuardDuty** - 6 events
- `CreateDetector`, `DeleteDetector`, `UpdateDetector`
- `CreateIPSet`, `DeleteIPSet`, `UpdateIPSet`

**Rationale:** Threat detection service. Configuration drift affects security monitoring.

**Security Hub** - 8 events
- `EnableSecurityHub`, `DisableSecurityHub`, `UpdateSecurityHubConfiguration`
- `EnableImportFindingsForProduct`, `DisableImportFindingsForProduct`
- `CreateActionTarget`, `DeleteActionTarget`, `UpdateActionTarget`

**Rationale:** Centralized security management. Configuration critical for compliance.

#### 🟠 P1: High Priority

**ACM (Certificate Manager)** - 6 events
- `RequestCertificate`, `DeleteCertificate`, `RenewCertificate`
- `ImportCertificate`, `AddTagsToCertificate`, `RemoveTagsFromCertificate`

**Rationale:** SSL/TLS certificates. Drift affects application availability.

**CloudTrail (Meta!)** - 8 events
- `CreateTrail`, `DeleteTrail`, `UpdateTrail`
- `StartLogging`, `StopLogging`
- `PutEventSelectors`, `PutInsightSelectors`
- `AddTags`, `RemoveTags`

**Rationale:** Audit infrastructure. Trail configuration affects compliance.

**Config** - 10 events
- `PutConfigurationRecorder`, `DeleteConfigurationRecorder`
- `PutDeliveryChannel`, `DeleteDeliveryChannel`
- `PutConfigRule`, `DeleteConfigRule`
- `PutAggregationAuthorization`, `DeleteAggregationAuthorization`
- `PutConformancePack`, `DeleteConformancePack`

**Rationale:** Compliance monitoring. Rule drift affects audit readiness.

#### 🟡 P2: Medium Priority

**Backup** - 8 events
- `CreateBackupVault`, `DeleteBackupVault`, `PutBackupVaultAccessPolicy`
- `CreateBackupPlan`, `DeleteBackupPlan`, `UpdateBackupPlan`
- `CreateBackupSelection`, `DeleteBackupSelection`

**Rationale:** DR infrastructure. Backup plan drift affects recovery capability.

**Transit Gateway** - 10 events
- `CreateTransitGateway`, `DeleteTransitGateway`, `ModifyTransitGateway`
- `CreateTransitGatewayVpcAttachment`, `DeleteTransitGatewayVpcAttachment`, `ModifyTransitGatewayVpcAttachment`
- `CreateTransitGatewayRoute`, `DeleteTransitGatewayRoute`
- `CreateTransitGatewayRouteTable`, `DeleteTransitGatewayRouteTable`

**Rationale:** Network infrastructure. Routing drift affects connectivity.

**Estimated Total for v0.5.0:** +71 events (337 total)

---

### v0.6.0+ (Q4 2025 - Target: December 2025)

**Theme:** Specialized Services & Long Tail

#### Planned Services (Priority TBD)

**Networking & Content Delivery:**
- Global Accelerator (6 events)
- Route 53 Resolver (8 events)
- VPC Lattice (10 events)
- CloudFront Functions (4 events)

**Database & Analytics:**
- Redshift (12 events)
- Athena (6 events)
- Kinesis (10 events)
- DMS (Database Migration Service) (8 events)

**Machine Learning:**
- SageMaker (15 events)
- Bedrock (8 events)

**Application Integration:**
- AppSync (10 events)
- MQ (8 events)
- MSK (Managed Kafka) (10 events)

**Developer Tools:**
- CodePipeline (10 events)
- CodeBuild (8 events)
- CodeDeploy (8 events)
- CodeCommit (6 events)

**Management & Governance:**
- Organizations (8 events)
- Control Tower (6 events)
- Service Catalog (8 events)
- Resource Groups (4 events)

**Storage:**
- EFS (Elastic File System) (8 events)
- FSx (10 events)
- Storage Gateway (8 events)

**Estimated Total for v0.6.0+:** +190 events (527+ total)

---

## 🗺️ Long-Term Vision (2026+)

### Multi-Cloud Expansion

**GCP (Google Cloud Platform)**
- Cloud Audit Logs integration
- Terraform state comparison
- 100+ GCP services

**Azure**
- Activity Log integration
- Terraform state comparison
- 100+ Azure services

### Advanced Features

**Auto-Remediation**
- Automatic terraform apply for approved changes
- Self-healing infrastructure
- Policy-based remediation

**ML-Powered Anomaly Detection**
- Pattern recognition for drift
- Predictive alerts
- Risk scoring

**Policy as Code Integration**
- OPA (Open Policy Agent) rules
- Rego policy enforcement
- Compliance-as-Code

---

## 📊 Coverage Metrics Goals

### v0.3.0 Goals (Q1 2025)
- 📊 **200+ events** covered
- 🎯 **20+ AWS services**
- 🔒 **Critical security services** 100% covered
- 📈 **Container services** comprehensive coverage

### v0.4.0 Goals (Q2 2025)
- 📊 **270+ events** covered
- 🎯 **30+ AWS services**
- 🔧 **Application services** comprehensive coverage
- 🔗 **Integration services** well covered

### v0.5.0 Goals (Q3 2025)
- 📊 **340+ events** covered
- 🎯 **40+ AWS services**
- 🛡️ **Security services** comprehensive coverage
- ✅ **Compliance services** well covered

### v0.6.0+ Goals (Q4 2025)
- 📊 **500+ events** covered
- 🎯 **60+ AWS services**
- 🌐 **Long tail services** covered
- 🚀 **Multi-cloud** foundation

---

## 🤝 Community Input

### How to Influence the Roadmap

**Vote on Services:**
- 👍 React to issues with 👍 for services you need
- 💬 Comment with your use case
- 🌟 Star the repo to show general support

**Request New Services:**
1. Check [existing issues](https://github.com/higakikeita/tfdrift-falco/issues?q=is%3Aissue+label%3Aservice-request)
2. Open new issue with [Service Request template](https://github.com/higakikeita/tfdrift-falco/issues/new?template=service-request.md)
3. Provide:
   - Service name
   - Key events to monitor
   - Your use case
   - Priority justification

**Contribute:**
- 🔧 Implement service coverage (see [CONTRIBUTING.md](../CONTRIBUTING.md))
- 📝 Write documentation
- 🧪 Test beta features
- 📊 Share usage data

### Top Community Requests

Track community-requested services here:
https://github.com/higakikeita/tfdrift-falco/discussions/categories/service-requests

Current top requests (will be updated based on votes):
1. ECS/EKS (containers)
2. Secrets Manager (security)
3. Step Functions (workflows)
4. WAF (security)
5. ElastiCache (performance)

---

## 📈 Success Criteria

### Service Coverage Targets

**Critical Services (Must Have):**
- ✅ All security services (IAM, KMS, WAF, GuardDuty, etc.)
- ✅ Core compute (EC2, ECS, EKS, Lambda)
- ✅ Core networking (VPC, Route53, CloudFront)
- ✅ Core storage (S3, EBS, RDS)

**High-Priority Services (Should Have):**
- ✅ Application integration (EventBridge, SNS, SQS, Step Functions)
- ✅ Developer tools (CodePipeline, CodeBuild, CodeDeploy)
- ✅ Monitoring (CloudWatch, X-Ray)
- ✅ Compliance (Config, Security Hub, CloudTrail)

**Medium-Priority Services (Nice to Have):**
- ⚠️ Analytics (Athena, Glue, Kinesis, Redshift)
- ⚠️ ML/AI (SageMaker, Bedrock)
- ⚠️ Specialized networking (Transit Gateway, Global Accelerator)
- ⚠️ Specialized storage (EFS, FSx)

### Quality Targets

Each new service must meet:
- ✅ 80%+ test coverage
- ✅ Comprehensive documentation
- ✅ Falco rules with examples
- ✅ Grafana dashboard integration
- ✅ Integration tests
- ✅ Performance benchmarks

---

## 🚀 Getting Started with Service Development

### Want to Implement a Service?

**1. Check the roadmap** (this document)
**2. Claim an issue** or create one
**3. Read the guide:** [Adding New AWS Services](../CONTRIBUTING.md#adding-new-aws-services)
**4. Follow the template:** `pkg/detector/service_template.go`
**5. Submit PR** with tests and docs

### Service Implementation Checklist

- [ ] CloudTrail event list identified
- [ ] Detector implementation (`pkg/detector/service_name.go`)
- [ ] Unit tests (80%+ coverage)
- [ ] Integration tests
- [ ] Falco rules (`rules/tfdrift-service-name.yaml`)
- [ ] Documentation (`docs/services/service-name.md`)
- [ ] Grafana dashboard updates
- [ ] CHANGELOG entry
- [ ] Example configuration

---

## 📞 Questions?

- 💬 [GitHub Discussions](https://github.com/higakikeita/tfdrift-falco/discussions)
- 🐛 [Issue Tracker](https://github.com/higakikeita/tfdrift-falco/issues)
- 📧 Email: keita.higaki@example.com
- 🐦 Twitter: [@keitah0322](https://x.com/keitah0322)

---

**This roadmap is a living document and will be updated based on:**
- Community feedback
- Security trends
- AWS service launches
- Enterprise requirements
- Technical feasibility

**Last Major Update:** December 6, 2024
**Next Review:** January 15, 2025

---

⭐ **Star the repo** to stay updated on new service releases!

📦 **GitHub:** https://github.com/higakikeita/tfdrift-falco
