# ADR-006: Migrate Falco event ingestion from gRPC output to HTTP output

## Status

Accepted (supersedes the transport choice in [ADR-001](001-event-driven-falco-grpc.md))

## Date

2026-07-23

## Context

[ADR-001](001-event-driven-falco-grpc.md) adopted an event-driven architecture that
consumes Falco alerts over Falco's **gRPC output** API (`pkg/falco/subscriber.go`
subscribes to `outputs.Sub` via `falcosecurity/client-go`). Real-machine verification
of the real-time path (#311) surfaced two problems that make the gRPC transport a
dead end:

1. **Falco removed gRPC output.** Falco `0.44+` removed `grpc_output` entirely.
   `0.43.0` still ships it but logs *"deprecated as consequence of gRPC output
   deprecation"* on startup and dropped `private_key`/`cert_chain`/`root_certs` from
   the grpc config schema (they still function but emit a schema-validation warning).
   We pinned `docker-compose.yml` to `0.43.0` precisely because we could not go newer
   (#357). Staying on gRPC pins the whole project to an EOL Falco.

2. **The mTLS gRPC surface is operationally heavy.** It requires generating and
   distributing a CA + server + client cert set, and the `root_certs` foot-gun (must
   be `ca.crt`, not `server.crt`) already cost us real debugging time (#357).

The event *ingestion* decision of ADR-001 (Cloud API → audit logs → Falco plugin →
TFDrift → detector) is still correct and unchanged. Only the **Falco → TFDrift
transport** needs to move off the removed gRPC path.

Note: this ADR does **not** cover the separate cloudtrail-plugin continuous-ingest
emit gap (the plugin consumes SQS-live messages but only emits reliably in S3-batch
mode); that is tracked under the same issue (#360) and is orthogonal to the transport.

### Options considered

1. **Falco `http_output` → TFDrift HTTP receiver.** Falco POSTs each alert as JSON
   to a URL. TFDrift already runs an HTTP API server, and `http_output` blocks
   already exist (disabled) in every `deployments/falco/*.yaml`. Supported on all
   current and future Falco versions.
2. **Falcosidekick in between.** Falco → Falcosidekick → (many outputs). Adds a
   second service to operate. Falcosidekick's value is *fan-out* to external sinks
   (Slack, SIEM, OpenCTI) — a concern that belongs to the donation/observability
   track, not to TFDrift's own event consumption.
3. **Keep gRPC / vendor an old Falco.** Rejected: pins us to EOL Falco and keeps the
   mTLS operational burden.

## Decision

Adopt **Falco `http_output` → a TFDrift HTTP receiver** as the primary Falco→TFDrift
transport, superseding the gRPC output choice in ADR-001.

- Falco is configured with `http_output.enabled: true` and `url` pointing at a new
  TFDrift endpoint (default `POST /api/v1/falco/events`), one JSON alert per request
  (`json_output: true`).
- TFDrift gains an HTTP receiver that parses the Falco alert JSON into the existing
  `types.Event` and feeds the **same** downstream pipeline (parser → resource mapper →
  detector → broadcaster). The parsing/extraction logic in `subscriber.go`
  (`ParseFalcoOutput`, `ExtractChanges`, `ExtractResourceID`) is reused, not rewritten.
- The gRPC subscriber is kept for one release as an opt-in fallback
  (`falco.transport: grpc|http`, default `http`) and then removed, so existing 0.43
  deployments are not broken on upgrade.
- Falcosidekick is explicitly out of scope here and left to the donation track.

The updated pipeline: Cloud API → audit logs → Falco plugin → **Falco http_output** →
**TFDrift HTTP receiver** → event parser → resource mapper → drift detector →
broadcaster → UI.

## Consequences

### Positive

- Works on current and future Falco (unpins us from EOL `0.43.0`; unblocks `0.44+`).
- Drops the mTLS cert set and the `root_certs` foot-gun for the default path; auth
  becomes a standard HTTP concern (shared secret / bearer, aligned with the API's
  own auth work in #341).
- Reuses the existing HTTP server and the existing parse/extract code — small blast
  radius.
- One less streaming/reconnect surface to own (the reconnect logic added in #312 was
  gRPC-stream-specific).

### Negative

- Push model: Falco must reach TFDrift's URL (network/DNS/ingress), where gRPC was a
  pull subscription. Mitigated by co-locating in the same compose/Helm network.
- A brief two-transport window (grpc + http) until the gRPC path is removed.
- The HTTP receiver must be hardened (auth, body size limits, source restriction) —
  folded into the API-auth hardening (#341) rather than solved twice.

### Follow-ups

- Implement the receiver + `falco.transport` switch (#360).
- Flip `http_output` on and `grpc_output` off in `deployments/**` and
  `docker-compose.yml`; re-verify real-time end-to-end on real AWS before closing #311.
- Remove the gRPC subscriber one release later.
