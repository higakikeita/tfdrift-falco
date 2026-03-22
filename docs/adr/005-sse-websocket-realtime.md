# ADR-005: SSE + WebSocket for Real-Time Communication

## Status

Accepted

## Date

2026-03-22

## Context

TFDrift-Falco's Dashboard UI needs to display drift events, alerts, and system status in real-time. The backend detects drift events continuously, and users expect to see them appear without refreshing the page.

Options considered:

1. **Polling** — Frontend polls API at regular intervals
2. **Server-Sent Events (SSE) only** — Unidirectional server-to-client stream
3. **WebSocket only** — Bidirectional communication channel
4. **SSE + WebSocket hybrid** — SSE for notifications, WebSocket for interactive features

## Decision

We implement a hybrid approach using both SSE and WebSocket:

- **SSE** (`/api/v1/stream`) — Used for push notifications (new drift events, alerts, status updates). The Go backend uses a Broadcaster pattern that fans out events to all connected SSE clients.
- **WebSocket** (`/ws`) — Used for interactive features requiring bidirectional communication (future: live graph updates, collaborative features).

The frontend SSE client (`ui/src/api/sseClient.ts`) manages connection lifecycle, automatic reconnection, and integrates with the Zustand toast store for user notifications.

## Consequences

### Positive

- SSE is simple, uses standard HTTP, works through proxies/load balancers, and auto-reconnects
- SSE is ideal for the primary use case (server pushing notifications to clients)
- WebSocket provides a path for future bidirectional features
- Broadcaster pattern efficiently fans out events to multiple clients
- NotificationPanel provides real-time connection status (live/offline)

### Negative

- Two real-time protocols add complexity to both backend and frontend
- SSE connections are long-lived, consuming server resources per connected client
- WebSocket requires upgrade-aware proxy configuration
- No message persistence — events missed during disconnection are lost

### Neutral

- SSE handles the majority of real-time needs; WebSocket is currently underutilized
- Connection status indicator helps users understand when they're receiving live data
- Future work may consolidate on WebSocket if bidirectional needs grow
