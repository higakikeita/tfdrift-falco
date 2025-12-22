# WebSocket API

Real-time bidirectional communication for interactive operations.

## Overview

WebSocket provides full-duplex communication between the client and TFDrift-Falco backend, enabling real-time updates and interactive queries.

**Endpoint:** `ws://localhost:8080/ws`

## Features

- ✅ **Topic-based Subscriptions** - Subscribe to specific event types
- ✅ **Real-time Updates** - Instant drift and event notifications
- ✅ **Bidirectional Communication** - Client can query, server can push
- ✅ **Automatic Reconnection** - Client libraries handle reconnection
- ✅ **Heartbeat Mechanism** - Ping/Pong to keep connection alive

---

## Connection

### JavaScript/TypeScript

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected to TFDrift WebSocket');

  // Subscribe to drift alerts
  ws.send(JSON.stringify({
    type: 'subscribe',
    topic: 'drifts'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);

  if (message.type === 'drift') {
    handleDriftAlert(message.payload);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Disconnected');
  // Implement reconnection logic
};
```

### Using React Hook

```typescript
import { useWebSocket } from './api/websocket';

function MyComponent() {
  const {
    isConnected,
    lastMessage,
    subscribe,
    unsubscribe
  } = useWebSocket();

  useEffect(() => {
    subscribe('drifts');
    return () => unsubscribe('drifts');
  }, []);

  useEffect(() => {
    if (lastMessage?.type === 'drift') {
      console.log('New drift:', lastMessage.payload);
    }
  }, [lastMessage]);

  return <div>Status: {isConnected ? 'Connected' : 'Disconnected'}</div>;
}
```

---

## Message Format

### Client → Server

```json
{
  "type": "subscribe" | "unsubscribe" | "ping" | "query",
  "topic": "all" | "drifts" | "events" | "state" | "stats",
  "payload": { ... }
}
```

### Server → Client

```json
{
  "type": "drift" | "event" | "state_change" | "pong" | "welcome",
  "topic": "drifts",
  "timestamp": "2025-01-15T10:00:00Z",
  "payload": { ... }
}
```

---

## Message Types

### Subscribe

Subscribe to a specific topic for real-time updates.

```json
{
  "type": "subscribe",
  "topic": "drifts"
}
```

**Available Topics:**
- `all` - All events
- `drifts` - Drift alerts only
- `events` - Falco events only
- `state` - Terraform state changes
- `stats` - Statistics updates

### Unsubscribe

```json
{
  "type": "unsubscribe",
  "topic": "drifts"
}
```

### Ping (Heartbeat)

```json
{
  "type": "ping"
}
```

**Response:**
```json
{
  "type": "pong",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

---

## Server Events

### Welcome

Sent immediately after connection.

```json
{
  "type": "welcome",
  "payload": {
    "message": "Connected to TFDrift-Falco WebSocket",
    "client_id": "client-uuid-here"
  }
}
```

### Drift Alert

Real-time drift detection notification.

```json
{
  "type": "drift",
  "topic": "drifts",
  "timestamp": "2025-01-15T10:00:00Z",
  "payload": {
    "severity": "critical",
    "resource_type": "aws_iam_policy",
    "resource_id": "ANPAI23HZ27SI6FQMGNQ2",
    "old_value": "...",
    "new_value": "...",
    "user": "john.doe@example.com"
  }
}
```

### Falco Event

Real-time Falco security event.

```json
{
  "type": "falco",
  "topic": "events",
  "timestamp": "2025-01-15T10:15:30Z",
  "payload": {
    "severity": "critical",
    "rule": "Unauthorized Process Execution",
    "output": "...",
    "container_id": "abc123"
  }
}
```

### State Change

Terraform state modification notification.

```json
{
  "type": "state_change",
  "topic": "state",
  "timestamp": "2025-01-15T10:20:00Z",
  "payload": {
    "resource": "aws_iam_policy.example",
    "action": "modified",
    "serial": 43
  }
}
```

---

## Connection Management

### Heartbeat

The server sends ping messages every 60 seconds. Clients should respond with pong within 54 seconds to maintain the connection.

**Server → Client:**
```
PING
```

**Client → Server:**
```json
{
  "type": "ping"
}
```

### Reconnection

Implement exponential backoff for reconnection:

```javascript
let reconnectAttempts = 0;
const maxAttempts = 5;

function connect() {
  const ws = new WebSocket('ws://localhost:8080/ws');

  ws.onclose = () => {
    if (reconnectAttempts < maxAttempts) {
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts), 30000);
      console.log(`Reconnecting in ${delay}ms...`);
      setTimeout(connect, delay);
      reconnectAttempts++;
    }
  };

  ws.onopen = () => {
    reconnectAttempts = 0; // Reset on successful connection
  };
}
```

---

## Error Handling

### Connection Errors

```javascript
ws.onerror = (error) => {
  console.error('WebSocket error:', error);
  // Log error details
  // Attempt reconnection
};
```

### Close Codes

Standard WebSocket close codes:
- `1000` - Normal closure
- `1001` - Going away (server shutdown)
- `1002` - Protocol error
- `1003` - Unsupported data
- `1006` - Abnormal closure
- `1008` - Policy violation

---

## Best Practices

1. **Always implement reconnection logic**
2. **Handle ping/pong for connection health**
3. **Subscribe only to needed topics**
4. **Buffer messages during disconnection**
5. **Implement exponential backoff**
6. **Clean up subscriptions on unmount**
7. **Log connection events for debugging**

---

## Related Documentation

- [REST API](./rest-api.md) - HTTP-based API endpoints
- [Server-Sent Events](./sse.md) - Unidirectional streaming
- [Frontend Integration](../IMPLEMENTATION_SUMMARY.md#websocket-client) - React hook usage

---

**Last Updated:** January 2025
