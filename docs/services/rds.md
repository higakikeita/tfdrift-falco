# RDS / Aurora — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateDBInstance | DB instance created | ✔ |
| DeleteDBInstance | DB instance deleted | ✔ |
| ModifyDBInstance | Instance config updated | ✔ |
| CreateDBCluster | Aurora cluster created | ✔ |
| DeleteDBCluster | Aurora cluster deleted | ✔ |
| ModifyDBCluster | Cluster config updated | ✔ |
| AddTagsToResource | Tags added/updated | ✔ |
| RemoveTagsFromResource | Tags removed | ✔ |
| DeleteDBSnapshot | Snapshot deleted | ✔ |
| ModifyDBParameterGroup | Parameter group updated | ✔ |
| ModifyDBSubnetGroup | Subnet group modified | ✔ |

## Monitored Drift Attributes

### DB Instance
- instance_class (e.g., db.t3.micro → db.t3.small)
- allocated_storage
- engine_version
- storage_encrypted
- deletion_protection
- multi_az
- monitoring_interval
- backup_retention_period
- backup_window
- maintenance_window
- publicly_accessible
- iam_database_authentication_enabled

### DB Cluster (Aurora)
- engine_version
- storage_encrypted
- kms_key_id
- preferred_backup_window
- preferred_maintenance_window
- backup_retention_period
- deletion_protection
- enable_http_endpoint (Aurora Serverless)

### Parameter Groups
- parameter values (configuration drift)

## Falco Rule Examples

```yaml
rule: rds_instance_modified
condition:
  cloud.service = "rds" and evt.name = "ModifyDBInstance"
output: "RDS Instance Modified (instance=%resource changes=%drift.changes user=%user)"
priority: warning

rule: rds_deletion_protection_disabled
condition:
  cloud.service = "rds" and evt.name in ("ModifyDBInstance","ModifyDBCluster") and
  drift.changes.deletion_protection = false
output: "RDS Deletion Protection Disabled (resource=%resource user=%user)"
priority: critical
```

## Example Log Output

```json
{
  "service": "rds",
  "event": "ModifyDBInstance",
  "resource": "db-instance-1",
  "changes": {
    "engine_version": ["14.6", "14.7"],
    "multi_az": [false, true],
    "backup_retention_period": [7, 14]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- RDS instance class changes
- Storage scaling events
- Engine version upgrades
- Multi-AZ configuration changes

### Alerts
- Unplanned deletion protection removal
- Encryption disabled
- Public accessibility enabled
- Backup retention reduced

## Known Limitations

- Cross-region replication may have CloudTrail delay (eventual consistency)
- Aurora Serverless v2 auto-scaling not tracked in real-time (AWS limitation)
- RDS Proxy drift tracked separately (v0.3.0 planned)
- Performance Insights configuration changes partial
- Blue/Green deployment drift not fully supported yet

## Security Considerations

RDS drift detection is **critical for data security**:
- **Encryption removal** → compliance violation
- **Public accessibility** → data exposure risk
- **Deletion protection** → accidental data loss prevention
- **Backup retention** → disaster recovery capability

**Recommendation**: Set critical priority for encryption and deletion protection rules.

## Release History

- **v0.2.0-beta**: Base RDS/Aurora coverage (11 events)
- **v0.3.0** (planned): RDS Proxy, Performance Insights, Aurora Global Database
