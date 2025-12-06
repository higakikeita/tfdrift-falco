# SNS — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateTopic | Topic created | ✔ |
| DeleteTopic | Topic deleted | ✔ |
| SetTopicAttributes | Topic attributes updated | ✔ |
| Subscribe | Subscription created | ✔ |
| Unsubscribe | Subscription deleted | ✔ |
| SetSubscriptionAttributes | Subscription attributes updated | ✔ |
| AddPermission | Topic policy permission added | ✔ |
| RemovePermission | Topic policy permission removed | ✔ |

## Monitored Drift Attributes

### Topic
- display_name
- delivery_policy
- policy (access policy JSON)
- kms_master_key_id (encryption)
- fifo_topic (for FIFO topics)
- content_based_deduplication
- tags

### Subscription
- protocol (http, https, email, email-json, sms, sqs, lambda, application, firehose)
- endpoint
- raw_message_delivery
- filter_policy (message filtering JSON)
- redrive_policy (DLQ configuration)
- delivery_policy
- subscription_role_arn (for Kinesis Firehose)

## Falco Rule Examples

```yaml
rule: sns_topic_policy_modified
condition:
  cloud.service = "sns" and evt.name in ("SetTopicAttributes","AddPermission","RemovePermission") and
  drift.attribute = "Policy"
output: "SNS Topic Policy Modified (topic=%resource user=%user changes=%drift.changes)"
priority: warning

rule: sns_encryption_disabled
condition:
  cloud.service = "sns" and evt.name = "SetTopicAttributes" and
  drift.changes.kms_master_key_id = null
output: "SNS Topic Encryption Disabled (topic=%resource user=%user)"
priority: critical
```

## Example Log Output

```json
{
  "service": "sns",
  "event": "SetTopicAttributes",
  "resource": "arn:aws:sns:us-east-1:123456789012:MyTopic",
  "changes": {
    "attribute_name": "Policy",
    "old_value": "{\"Version\":\"2012-10-17\",\"Statement\":[...]}",
    "new_value": "{\"Version\":\"2012-10-17\",\"Statement\":[...]}"
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- SNS topic policy changes
- Subscription modifications
- Encryption configuration changes
- Filter policy updates

### Alerts
- Unplanned topic policy changes
- Encryption disabled
- Public topic access granted
- DLQ (redrive policy) removed

## Known Limitations

- SNS delivery failure logs not tracked (requires CloudWatch Logs integration)
- FIFO topic high-throughput mode changes partial
- Message data protection policy drift not fully supported (v0.3.0 planned)
- Cross-account subscription drift requires both accounts' CloudTrail logs

## Security Considerations

SNS drift detection is **important for messaging security**:
- **Policy changes** → unauthorized message publishing
- **Encryption disabled** → sensitive data exposure
- **Public access** → spam/abuse potential
- **Subscription changes** → message interception

**Recommendation**: Set warning/error priority for policy and encryption changes.

## Release History

- **v0.2.0-beta**: Core SNS topic/subscription coverage (8 events)
- **v0.3.0** (planned): Message data protection, archive policy
