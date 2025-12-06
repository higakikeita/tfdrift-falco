# KMS — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateKey | KMS key created | ✔ |
| ScheduleKeyDeletion | Key deletion scheduled | ✔ |
| CancelKeyDeletion | Key deletion cancelled | ✔ |
| DisableKey | Key disabled | ✔ |
| EnableKey | Key enabled | ✔ |
| PutKeyPolicy | Key policy updated | ✔ |
| CreateAlias | Key alias created | ✔ |
| DeleteAlias | Key alias deleted | ✔ |
| UpdateAlias | Alias target updated | ✔ |
| EnableKeyRotation | Automatic key rotation enabled | ✔ |
| DisableKeyRotation | Automatic key rotation disabled | ✔ |
| TagResource | Tags added/updated | ✔ |
| UntagResource | Tags removed | ✔ |

## Monitored Drift Attributes

### KMS Key
- description
- key_usage (ENCRYPT_DECRYPT / SIGN_VERIFY / GENERATE_VERIFY_MAC)
- customer_master_key_spec (key spec)
- key_state (Enabled / Disabled / PendingDeletion)
- deletion_date (when scheduled for deletion)
- enable_key_rotation
- policy (key policy JSON)
- multi_region (true/false)
- tags

### Key Alias
- alias_name
- target_key_id

## Falco Rule Examples

```yaml
rule: kms_key_deletion_scheduled
condition:
  cloud.service = "kms" and evt.name = "ScheduleKeyDeletion"
output: "KMS Key Deletion Scheduled (key=%resource deletion_date=%drift.deletion_date user=%user)"
priority: critical

rule: kms_key_rotation_disabled
condition:
  cloud.service = "kms" and evt.name = "DisableKeyRotation"
output: "KMS Key Rotation Disabled (key=%resource user=%user)"
priority: error

rule: kms_key_policy_modified
condition:
  cloud.service = "kms" and evt.name = "PutKeyPolicy"
output: "KMS Key Policy Modified (key=%resource user=%user)"
priority: critical
```

## Example Log Output

```json
{
  "service": "kms",
  "event": "PutKeyPolicy",
  "resource": "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
  "changes": {
    "policy_name": "default",
    "policy": {
      "added_principals": ["arn:aws:iam::123456789012:role/NewRole"],
      "removed_principals": []
    }
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- KMS key policy changes
- Key deletion schedules
- Key rotation status changes
- Key state transitions (Enabled/Disabled)

### Alerts
- Unplanned key deletion scheduled
- Key rotation disabled
- Key policy grant added to external account
- Key disabled unexpectedly

## Known Limitations

- KMS key usage (Encrypt/Decrypt operations) not tracked by TFDrift (requires CloudTrail data events)
- Grant drift tracked but grant tokens not parsed
- Multi-region key replica drift partial (eventual consistency)
- Custom key store (CloudHSM) drift not fully supported yet
- Key material import drift not tracked

## Security Considerations

KMS drift detection is **CRITICAL for security and compliance**:
- **Key deletion** → data loss, service disruption
- **Policy changes** → unauthorized decryption access
- **Key disabled** → application failures
- **Rotation disabled** → compliance violation (PCI-DSS, HIPAA)

**Recommendation**:
- Set CRITICAL priority for key deletion and policy changes
- Enable CloudWatch alarms for all KMS drift events
- Review KMS key policies quarterly

## Release History

- **v0.2.0-beta**: Core KMS key/alias/policy coverage (13 events)
- **v0.3.0** (planned): Multi-region key enhancements, custom key store support
