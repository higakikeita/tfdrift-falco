# Server-Sent Events (SSE) API

Unidirectional real-time streaming from server to client.

## Overview

Server-Sent Events provide a simple HTTP-based protocol for server-to-client streaming, ideal for real-time notifications and updates without the overhead of WebSocket.

**Endpoint:** `GET http://localhost:8080/api/v1/stream`

## Features

- ✅ **HTTP-based** - Works through firewalls and proxies
- ✅ **Automatic Reconnection** - Built into EventSource API
- ✅ **Text-based Protocol** - Simple to debug and implement
- ✅ **Multiple Event Types** - Typed events for different notifications
- ✅ **Low Latency** - < 100ms from event to client

---

## Connection

### JavaScript/TypeScript

```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/stream');

// Connected event
eventSource.addEventListener('connected', (event) => {
  const data = JSON.parse(event.data);
  console.log('Stream connected:', data);
});

// Drift alerts
eventSource.addEventListener('drift', (event) => {
  const drift = JSON.parse(event.data);
  console.log('New drift detected:', drift);
  showNotification(drift);
});

// Falco events
eventSource.addEventListener('falco', (event) => {
  const falcoEvent = JSON.parse(event.data);
  console.log('Security event:', falcoEvent);
  logSecurityEvent(falcoEvent);
});

// State changes
eventSource.addEventListener('state_change', (event) => {
  const change = JSON.parse(event.data);
  console.log('State updated:', change);
});

// Error handling
eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  // EventSource automatically reconnects
};
```

### Using React Hook

```typescript
import { useSSE, useDriftAlerts } from './api/sse';

function MyComponent() {
  const { isConnected, events } = useSSE();
  const { driftAlerts, clearAlerts } = useDriftAlerts();

  return (
    <div>
      <div>Stream: {isConnected ? 'Active' : 'Inactive'}</div>
      <ul>
        {driftAlerts.map(drift => (
          <li key={drift.id}>
            {drift.severity}: {drift.resource_type}
          </li>
        ))}
      </ul>
    </div>
  );
}
```

---

## Event Types

### connected

Sent immediately upon successful connection.

```
event: connected
data: {"stream_id":"stream-uuid-here","message":"Connected to TFDrift-Falco SSE stream"}
```

### drift

Drift alert event.

```
event: drift
data: {"severity":"critical","resource_type":"aws_iam_policy","resource_id":"...","old_value":"...","new_value":"..."}
```

### falco

Falco security event.

```
event: falco
data: {"severity":"critical","rule":"Unauthorized Process Execution","output":"...","container_id":"abc123"}
```

### state_change

Terraform state change notification.

```
event: state_change
data: {"resource":"aws_iam_policy.example","action":"modified","serial":43}
```

### keep-alive

Periodic comment to keep connection alive (every 30 seconds).

```
: keep-alive
```

---

## Event Data Format

All events use JSON format:

```json
{
  "severity": "critical",
  "resource_type": "aws_iam_policy",
  "resource_id": "ANPAI23HZ27SI6FQMGNQ2",
  "old_value": "...",
  "new_value": "...",
  "user_identity": {
    "type": "IAMUser",
    "user_name": "john.doe",
    "arn": "arn:aws:iam::123456789012:user/john.doe"
  },
  "timestamp": "2025-01-15T10:00:00Z"
}
```

---

## Connection Management

### Automatic Reconnection

EventSource API automatically reconnects on connection loss. No manual implementation needed.

**Reconnection Behavior:**
- Initial connection: Immediate
- After disconnect: 3 seconds default
- Exponential backoff: Handled by browser

### Manual Reconnection Control

```javascript
let eventSource;

function connect() {
  eventSource = new EventSource('http://localhost:8080/api/v1/stream');

  eventSource.addEventListener('connected', handleConnected);
  eventSource.addEventListener('drift', handleDrift);

  eventSource.onerror = (error) => {
    console.error('Connection error:', error);
    if (eventSource.readyState === EventSource.CLOSED) {
      console.log('Connection closed, will auto-reconnect');
    }
  };
}

function disconnect() {
  if (eventSource) {
    eventSource.close();
    eventSource = null;
  }
}
```

---

## Error Handling

### Connection Errors

```javascript
eventSource.onerror = (error) => {
  console.error('SSE error:', error);

  // Check connection state
  switch (eventSource.readyState) {
    case EventSource.CONNECTING:
      console.log('Reconnecting...');
      break;
    case EventSource.OPEN:
      console.log('Connection open');
      break;
    case EventSource.CLOSED:
      console.log('Connection closed');
      break;
  }
};
```

### Event Parsing Errors

```javascript
eventSource.addEventListener('drift', (event) => {
  try {
    const data = JSON.parse(event.data);
    handleDrift(data);
  } catch (error) {
    console.error('Failed to parse drift event:', error);
  }
});
```

---

## Performance

### Memory Management

```javascript
// Limit event history
const MAX_EVENTS = 100;
const eventHistory = [];

eventSource.addEventListener('drift', (event) => {
  const data = JSON.parse(event.data);
  eventHistory.push(data);

  // Keep only last 100 events
  if (eventHistory.length > MAX_EVENTS) {
    eventHistory.shift();
  }
});
```

### Resource Cleanup

```javascript
// React component cleanup
useEffect(() => {
  const eventSource = new EventSource('/api/v1/stream');

  // ... event listeners

  return () => {
    eventSource.close(); // Important: close on unmount
  };
}, []);
```

---

## Comparison: SSE vs WebSocket

| Feature | SSE | WebSocket |
|---------|-----|-----------|
| Direction | Server → Client | Bidirectional |
| Protocol | HTTP | ws:// |
| Reconnection | Automatic | Manual |
| Browser Support | Excellent | Excellent |
| Firewall | No issues | May be blocked |
| Complexity | Simple | More complex |
| Use Case | Notifications | Interactive |

**When to use SSE:**
- Real-time notifications
- Event streams
- Simple one-way communication
- Through restrictive firewalls

**When to use WebSocket:**
- Interactive features
- Client needs to send data
- Binary data transfer
- Low latency required

---

## Best Practices

1. **Always close connections on cleanup**
2. **Limit event history in memory**
3. **Handle JSON parsing errors**
4. **Use EventSource for browser compatibility**
5. **Monitor connection state**
6. **Log connection events for debugging**
7. **Implement graceful degradation**

---

## Troubleshooting

### Connection Refused

```bash
# Check server is running
curl http://localhost:8080/health

# Check SSE endpoint
curl -N http://localhost:8080/api/v1/stream
```

### No Events Received

```bash
# Verify events are being generated
docker-compose logs backend | grep "Broadcasting"

# Check Falco is sending events
docker-compose logs falco
```

### CORS Issues

If accessing from different origin, ensure CORS is configured:

```go
// Backend CORS configuration
cors.New(cors.Options{
    AllowedOrigins: []string{"http://localhost:3000"},
    AllowedMethods: []string{"GET", "OPTIONS"},
})
```

---

## Related Documentation

- [REST API](API.md) - HTTP-based API endpoints
- [WebSocket API](websocket.md) - Bidirectional communication
- [Frontend Integration](../IMPLEMENTATION_SUMMARY.md#sse-client) - React hook usage

---

**Last Updated:** January 2025
