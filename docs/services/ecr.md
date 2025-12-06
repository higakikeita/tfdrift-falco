# ECR — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateRepository | Repository created | ✔ |
| DeleteRepository | Repository deleted | ✔ |
| PutImageScanningConfiguration | Image scanning config updated | ✔ |
| PutImageTagMutability | Tag mutability updated | ✔ |
| PutLifecyclePolicy | Lifecycle policy added/updated | ✔ |
| DeleteLifecyclePolicy | Lifecycle policy deleted | ✔ |
| SetRepositoryPolicy | Repository policy updated | ✔ |
| DeleteRepositoryPolicy | Repository policy deleted | ✔ |
| PutReplicationConfiguration | Replication config updated | ✔ |

## Monitored Drift Attributes

### Repository
- name
- image_tag_mutability (MUTABLE / IMMUTABLE)
- image_scanning_configuration
  - scan_on_push
- encryption_configuration
  - encryption_type (AES256 / KMS)
  - kms_key
- tags

### Repository Policy
- policy (access policy JSON)

### Lifecycle Policy
- lifecycle_policy (image retention rules JSON)

### Replication Configuration
- replication_configuration
  - rules
    - destinations (region, registry ID)
    - repository_filters

## Falco Rule Examples

```yaml
rule: ecr_image_tag_mutability_enabled
condition:
  cloud.service = "ecr" and evt.name = "PutImageTagMutability" and
  drift.changes.image_tag_mutability = "MUTABLE"
output: "ECR Image Tag Mutability Enabled (repository=%resource user=%user)"
priority: warning

rule: ecr_repository_policy_modified
condition:
  cloud.service = "ecr" and evt.name in ("SetRepositoryPolicy","DeleteRepositoryPolicy")
output: "ECR Repository Policy Modified (repository=%resource user=%user)"
priority: error
```

## Example Log Output

```json
{
  "service": "ecr",
  "event": "PutImageTagMutability",
  "resource": "my-app-repo",
  "changes": {
    "image_tag_mutability": ["IMMUTABLE", "MUTABLE"]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- ECR repository policy changes
- Image scanning configuration updates
- Tag mutability changes
- Lifecycle policy modifications

### Alerts
- Image tag mutability enabled (IMMUTABLE → MUTABLE)
- Repository policy made public
- Image scanning disabled
- Lifecycle policy deleted (retention risk)

## Known Limitations

- Image push/pull events not tracked by TFDrift (use CloudTrail data events if needed)
- ECR Public repositories have separate API (not covered yet)
- Vulnerability scan results not included in drift logs
- Cross-region replication status not real-time (eventual consistency)

## Security Considerations

ECR drift detection is **important for container security**:
- **Tag mutability** → supply chain attack risk (image overwrite)
- **Policy changes** → unauthorized image access
- **Scanning disabled** → vulnerability detection loss
- **Public repository** → intellectual property exposure

**Recommendation**: Set warning/error priority for tag mutability and policy changes.

## Release History

- **v0.2.0-beta**: Core ECR repository configuration (9 events)
- **v0.3.0** (planned): ECR Public, pull-through cache configuration
