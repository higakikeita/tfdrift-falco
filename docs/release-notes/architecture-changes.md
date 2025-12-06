# Architecture Evolution

This document tracks major architectural changes across TFDrift-Falco releases.

---

## v0.2.0-beta (November 2024)

### Service Layer Refactoring

**Change:** Introduced modular service-specific detectors

**Before (v0.1.x):**
```
detector/
  ├── detector.go (monolithic, 2000+ lines)
  └── cloudtrail.go
```

**After (v0.2.0-beta):**
```
detector/
  ├── detector.go (orchestrator, 300 lines)
  ├── service/
  │   ├── ec2.go
  │   ├── iam.go
  │   ├── s3.go
  │   └── [12 service files]
  └── cloudtrail.go
```

**Benefits:**
- Easier to add new services
- Better test coverage (80%+ achieved)
- Parallel service processing

---

### State Comparison Algorithm

**Change:** Moved from full state diff to incremental change detection

**Old Approach:**
1. Load full Terraform state (can be 100MB+)
2. Compare every resource attribute
3. Generate diff for entire state

**New Approach:**
1. Load only resources mentioned in CloudTrail events
2. Compare specific attributes changed in the event
3. Generate targeted diff

**Performance Impact:**
- **Processing time**: 5s → 0.5s (for typical deployments)
- **Memory usage**: 500MB → 50MB
- **CloudTrail API calls**: Reduced by 60%

---

### Falco Output Format

**Change:** Standardized structured output format

**Old Format (v0.1.x):**
```
Drift detected in EC2 instance i-123456
```

**New Format (v0.2.0-beta):**
```json
{
  "service": "ec2",
  "event": "ModifyInstanceAttribute",
  "resource": "i-123456",
  "changes": {
    "instance_type": ["t3.micro", "t3.small"]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

**Benefits:**
- Machine-parsable for alerting systems
- Includes user attribution
- Contains detailed change information

---

### Grafana Dashboard Architecture

**Change:** Moved from single monolithic dashboard to service-specific panels

**Old Architecture:**
- 1 dashboard with 50+ panels (slow loading)
- Hard to customize per-service

**New Architecture:**
- 12 service-specific dashboards
- 1 overview dashboard with links to service dashboards
- Shared template variables (account, region, time range)

**Loading Performance:**
- Dashboard load time: 10s → 2s
- Query performance: 5s → 1s (indexed by service)

---

## v0.3.0 (Planned - Q1 2025)

### Event Processing Pipeline

**Planned Change:** Introduce event queue for asynchronous processing

**Current Architecture (v0.2.0):**
```
CloudTrail → Detector → Falco → Grafana
(synchronous, blocking)
```

**Planned Architecture (v0.3.0):**
```
CloudTrail → Queue (SQS/Kafka) → Workers → Falco → Grafana
                                  ↓
                               Alerting (Slack, PagerDuty)
```

**Benefits:**
- Handle high-volume CloudTrail events (1000+/min)
- Retry failed events automatically
- Scale workers independently

---

### Multi-Account State Management

**Planned Change:** Centralized state store with account isolation

**Current (v0.2.0):**
```
Single Terraform state file
Limited to one AWS account
```

**Planned (v0.3.0):**
```
state_store/
  ├── account_123456789012/
  │   ├── us-east-1/
  │   │   └── terraform.tfstate
  │   └── eu-west-1/
  │       └── terraform.tfstate
  └── account_234567890123/
      └── us-east-1/
          └── terraform.tfstate
```

**Configuration:**
```yaml
accounts:
  - id: "123456789012"
    state_backend: s3://my-bucket/prod/terraform.tfstate
  - id: "234567890123"
    state_backend: s3://my-bucket/staging/terraform.tfstate
```

---

### Rule Engine Redesign

**Planned Change:** Plugin-based rule engine for custom drift logic

**Current (v0.2.0):**
- Falco rules defined in YAML (static)
- Hard to add custom business logic

**Planned (v0.3.0):**
```go
type DriftRule interface {
    Match(event CloudTrailEvent, state TerraformState) bool
    Severity() string
    Output(ctx Context) string
}

// Custom rule example
type NoPublicS3Rule struct{}

func (r *NoPublicS3Rule) Match(event CloudTrailEvent, state TerraformState) bool {
    if event.EventName == "PutBucketAcl" {
        // Custom logic: Check if ACL contains "public-read"
        return checkPublicACL(event)
    }
    return false
}
```

**Benefits:**
- Custom drift policies per organization
- Dynamic rule loading
- Testable rule logic

---

## Design Principles

Throughout the evolution of TFDrift-Falco, we maintain these principles:

### 1. Modularity
- Each AWS service is self-contained
- Easy to add/remove services
- Clear interfaces between components

### 2. Performance
- Optimize for large-scale deployments (1000+ resources)
- Minimize CloudTrail API calls
- Efficient state comparison

### 3. Extensibility
- Plugin-based architecture
- Configuration over code
- Community contributions welcome

### 4. Observability
- Structured logging
- Prometheus metrics
- Grafana dashboards

---

## Migration Guides

### From v0.1.x to v0.2.0-beta
See [v0.2.0-beta Release Notes](v0.2.0-beta.md#upgrade-guide)

### From v0.2.0-beta to v0.3.0
(Will be published when v0.3.0 is released)

---

## Resources

- [Architecture Documentation](../architecture.md)
- [CHANGELOG](../../CHANGELOG.md)
- [GitHub Discussions - Architecture](https://github.com/higakikeita/tfdrift-falco/discussions/categories/architecture)
