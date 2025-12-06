# SQS — Drift Coverage

## Supported CloudTrail Events
| Event | Description | Status |
|-------|-------------|--------|
| CreateQueue | Queue created | ✔ |
| DeleteQueue | Queue deleted | ✔ |
| SetQueueAttributes | Queue attributes updated | ✔ |
| AddPermission | Queue policy permission added | ✔ |
| RemovePermission | Queue policy permission removed | ✔ |
| TagQueue | Tags added/updated | ✔ |
| UntagQueue | Tags removed | ✔ |

## Monitored Drift Attributes

### Queue Configuration
- visibility_timeout_seconds
- message_retention_seconds
- max_message_size
- delay_seconds (delivery delay)
- receive_wait_time_seconds (long polling)
- policy (access policy JSON)
- redrive_policy (DLQ configuration)
  - deadLetterTargetArn
  - maxReceiveCount
- kms_master_key_id (encryption)
- kms_data_key_reuse_period_seconds
- fifo_queue (for FIFO queues)
- content_based_deduplication
- deduplication_scope (FIFO only)
- fifo_throughput_limit (FIFO only)

## Falco Rule Examples

```yaml
rule: sqs_queue_policy_modified
condition:
  cloud.service = "sqs" and evt.name in ("SetQueueAttributes","AddPermission","RemovePermission") and
  drift.attribute = "Policy"
output: "SQS Queue Policy Modified (queue=%resource user=%user)"
priority: warning

rule: sqs_dlq_removed
condition:
  cloud.service = "sqs" and evt.name = "SetQueueAttributes" and
  drift.changes.redrive_policy = null
output: "SQS Dead Letter Queue Removed (queue=%resource user=%user)"
priority: error
```

## Example Log Output

```json
{
  "service": "sqs",
  "event": "SetQueueAttributes",
  "resource": "https://sqs.us-east-1.amazonaws.com/123456789012/MyQueue",
  "changes": {
    "visibility_timeout_seconds": ["30", "60"],
    "message_retention_seconds": ["345600", "1209600"]
  },
  "user": "admin@example.com",
  "timestamp": "2025-12-06T07:30:00Z"
}
```

## Grafana Dashboard Examples

### Metrics
- SQS queue attribute changes
- Policy modifications
- DLQ configuration changes
- Retention period updates

### Alerts
- Unplanned queue policy changes
- DLQ removed
- Encryption disabled
- Public queue access granted

## Known Limitations

- SQS message-level drift not tracked (CloudTrail doesn't log SendMessage/ReceiveMessage by default)
- High-throughput FIFO queue mode changes partial
- Server-side encryption (SSE-SQS vs SSE-KMS) transition not fully parsed
- Cross-account access drift requires both accounts' logs

## Security Considerations

SQS drift detection is **important for messaging reliability**:
- **Policy changes** → unauthorized message access
- **DLQ removed** → message loss risk
- **Encryption disabled** → sensitive data exposure
- **Retention reduced** → compliance violation

**Recommendation**: Set error priority for DLQ and encryption changes.

## Release History

- **v0.2.0-beta**: Core SQS queue configuration (7 events)
- **v0.3.0** (planned): High-throughput FIFO enhancements
