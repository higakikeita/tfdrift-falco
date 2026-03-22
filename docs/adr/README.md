# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the TFDrift-Falco project.

ADRs are short documents that capture important architectural decisions made along with their context and consequences. We follow the [Michael Nygard format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions).

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [ADR-000](000-use-adrs.md) | Use Architecture Decision Records | Accepted | 2026-03-22 |
| [ADR-001](001-event-driven-falco-grpc.md) | Event-driven architecture with Falco gRPC | Accepted | 2026-03-22 |
| [ADR-002](002-in-memory-graph-store.md) | In-memory graph store over external database | Accepted | 2026-03-22 |
| [ADR-003](003-multi-cloud-provider-abstraction.md) | Multi-cloud provider abstraction pattern | Accepted | 2026-03-22 |
| [ADR-004](004-dual-auth-jwt-apikey.md) | JWT + API Key dual authentication | Accepted | 2026-03-22 |
| [ADR-005](005-sse-websocket-realtime.md) | SSE + WebSocket for real-time communication | Accepted | 2026-03-22 |

## Template

Use [ADR template](template.md) when creating new ADRs.

## Statuses

- **Proposed** — Under discussion
- **Accepted** — Decision has been made and is in effect
- **Deprecated** — No longer valid, superseded by another ADR
- **Superseded** — Replaced by a newer ADR (link to replacement)
