# Route53 — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| ChangeResourceRecordSets | DNS record added/updated/deleted | ✔ |
| CreateHostedZone | Hosted zone created | ✔ |
| DeleteHostedZone | Hosted zone deleted | ✔ |
| UpdateHostedZoneComment | Hosted zone comment updated | ✔ |
| AssociateVPCWithHostedZone | VPC associated with private zone | ✔ |
| DisassociateVPCFromHostedZone | VPC disassociated | ✔ |

## Monitored Drift Attributes

### Record Sets
- name (DNS name)
- type (A, AAAA, CNAME, MX, TXT, etc.)
- ttl
- records (values)
- alias
  - name
  - zone_id
  - evaluate_target_health
- routing_policy
  - weighted
  - latency
  - failover
  - geolocation
  - geoproximity

### Hosted Zone
- name
- comment
- vpc_association
  - vpc_id
  - vpc_region
- delegation_set_id

## Falco Rule Examples

```yaml
rule: route53_record_modified
condition:
  cloud.service = "route53" and evt.name = "ChangeResourceRecordSets"
output: "Route53 Record Modified (zone=%drift.hosted_zone_id record=%drift.record_name type=%drift.record_type changes=%drift.changes user=%user)"
priority: error

rule: route53_alias_target_changed
condition:
  cloud.service = "route53" and evt.name = "ChangeResourceRecordSets" and
  drift.changes.alias_target != null
output: "Route53 Alias Target Changed (record=%drift.record_name from=%drift.old_alias_target to=%drift.new_alias_target user=%user)"
priority: critical
```

## Example Log Output

```json
{
  "service": "route53",
  "event": "ChangeResourceRecordSets",
  "resource": "Z1234567890ABC",
  "changes": {
    "record_name": "api.example.com",
    "record_type": "A",
    "action": "UPSERT",
    "ttl": ["60", "300"],
    "records": [
      ["192.0.2.1"],
      ["192.0.2.2"]
    ]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- Route53 record changes by zone
- Alias target modifications
- TTL changes
- Hosted zone VPC associations

### Alerts
- Unplanned A/AAAA record changes
- ALIAS → non-ALIAS conversions
- Low TTL values (<60s) for production domains
- Private zone made public

## Known Limitations

- Geo-location routing attributes partially tracked (coordinates not fully parsed)
- ALIAS → A record swap may produce multiple events (CloudTrail batching)
- Health check drift tracked separately (not in ChangeResourceRecordSets)
- DNSSEC configuration drift partial (v0.3.0 planned)
- Traffic policy drift not supported yet

## Security Considerations

Route53 drift detection is **critical for availability**:
- **A/AAAA record changes** → traffic redirection (potential hijack)
- **MX record changes** → email interception risk
- **ALIAS target changes** → service disruption
- **Private zone exposure** → internal DNS leak

**Recommendation**: Set critical priority for production domain record changes.

## Release History

- **v0.2.0-beta**: Base Route53 coverage (6 events)
- **v0.3.0** (planned): DNSSEC, Traffic Policy, Health Check integration
