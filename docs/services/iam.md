# IAM — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateRole | IAM role created | ✔ |
| DeleteRole | IAM role deleted | ✔ |
| UpdateAssumeRolePolicy | Trust policy updated | ✔ |
| AttachRolePolicy | Managed policy attached | ✔ |
| DetachRolePolicy | Managed policy detached | ✔ |
| PutRolePolicy | Inline policy added/updated | ✔ |
| DeleteRolePolicy | Inline policy deleted | ✔ |
| CreatePolicy | Managed policy created | ✔ |
| CreatePolicyVersion | Policy version updated | ✔ |
| DeletePolicy | Managed policy deleted | ✔ |
| CreateUser | IAM user created | ✔ |
| DeleteUser | IAM user deleted | ✔ |
| AttachUserPolicy | User policy attached | ✔ |
| DetachUserPolicy | User policy detached | ✔ |

## Monitored Drift Attributes

### IAM Role
- assume_role_policy (trust relationship)
- managed_policy_arns
- inline_policies
- max_session_duration
- permissions_boundary
- description
- tags

### IAM Policy
- policy_document
- description
- tags

### IAM User
- path
- permissions_boundary
- tags
- attached policies

## Falco Rule Examples

```yaml
rule: iam_role_trust_policy_modified
condition:
  cloud.service = "iam" and evt.name = "UpdateAssumeRolePolicy"
output: "IAM Role Trust Policy Modified (role=%resource user=%user changes=%drift.changes)"
priority: critical

rule: iam_admin_policy_attached
condition:
  cloud.service = "iam" and evt.name in ("AttachRolePolicy","AttachUserPolicy") and
  drift.policy_arn contains "AdministratorAccess"
output: "Admin Policy Attached (resource=%resource policy=%drift.policy_arn user=%user)"
priority: error
```

## Example Log Output

```json
{
  "service": "iam",
  "event": "UpdateAssumeRolePolicy",
  "resource": "MyAppRole",
  "changes": {
    "assume_role_policy": {
      "added_principals": ["arn:aws:iam::123456789012:role/NewTrustedRole"],
      "removed_principals": []
    }
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- IAM role trust policy changes
- Policy attachments/detachments
- Admin privilege escalations
- Cross-account trust modifications

### Alerts
- Unplanned IAM admin policy attachments
- IAM role trust policy modifications
- Service-to-service access changes

## Known Limitations

- IAM policy condition evaluation not analyzed (CloudTrail limitation)
- Permission boundary changes may have eventual consistency delay
- Service control policies (SCPs) tracked separately at AWS Organizations level
- SAML/OIDC provider drift partial (v0.3.0 planned)

## Security Considerations

IAM drift detection is **critical for security**. Unplanned changes to:
- Trust policies → potential privilege escalation
- Admin policies → unauthorized access
- Cross-account roles → data exfiltration risk

**Recommendation**: Set Falco priority to `CRITICAL` for IAM rules.

## Release History

- **v0.2.0-beta**: Core IAM role/policy/user coverage (14 events)
- **v0.3.0** (planned): IAM Identity Center, permission sets
