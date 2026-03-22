# ADR-001: Event-Driven Architecture with Falco gRPC

## Status

Accepted

## Date

2026-03-22

## Context

TFDrift-Falco needs to detect infrastructure drift in real-time. The traditional approach is periodic polling — scanning Terraform state and comparing it against cloud resources on a schedule. However, polling introduces latency (minutes to hours between scans) and wastes resources scanning unchanged infrastructure.

Falco is an open-source runtime security tool that can monitor cloud API audit logs (CloudTrail, GCP Audit Logs, Azure Activity Logs) via its plugin framework. It provides a gRPC output API that streams detected events in real-time.

We needed to decide between:

1. **Periodic polling** — Scheduled scans comparing Terraform state to cloud reality
2. **Event-driven via Falco gRPC** — Subscribe to Falco's gRPC stream for real-time cloud API events
3. **Direct cloud API integration** — Subscribe directly to CloudTrail/Pub/Sub/Event Grid

## Decision

We adopt an event-driven architecture using Falco's gRPC output API as the primary event source. TFDrift-Falco subscribes to Falco's gRPC stream (`pkg/falco/subscriber.go`) and processes events through provider-specific parsers (AWS, GCP, Azure) that map cloud API calls to Terraform resource types.

The event pipeline is: Cloud API → Audit Logs → Falco Plugin → gRPC Stream → TFDrift-Falco Subscriber → Event Parser → Resource Mapper → Drift Detection Engine.

## Consequences

### Positive

- Real-time drift detection (seconds, not minutes/hours)
- Resource-efficient — only processes actual changes, not full infrastructure scans
- Leverages Falco's mature plugin ecosystem for multi-cloud audit log parsing
- Security context — Falco provides threat detection alongside drift detection
- Extensible — new cloud providers can be added via Falco plugins

### Negative

- Dependency on Falco — requires Falco deployment alongside TFDrift-Falco
- Complexity — gRPC connection management, reconnection logic, backpressure handling
- Audit log delivery latency varies by provider (CloudTrail: 5-15 min via S3, GCP: 30s-5min via Pub/Sub)
- Cannot detect drift that occurred before TFDrift-Falco started (no historical scan)

### Neutral

- Falco's plugin framework handles the complexity of parsing provider-specific audit log formats
- gRPC provides efficient binary protocol with streaming support
