# TFDrift-Falco API Documentation

Complete API reference for TFDrift-Falco REST API, WebSocket, and SSE endpoints.

## Base URLs

```
REST API: http://localhost:8080/api/v1
WebSocket: ws://localhost:8080/ws
SSE Stream: http://localhost:8080/api/v1/stream
```

## Table of Contents

- [REST API](#rest-api)
  - [Health Check](#health-check)
  - [Graph Endpoints](#graph-endpoints)
  - [Terraform State](#terraform-state)
  - [Falco Events](#falco-events)
  - [Drift Alerts](#drift-alerts)
  - [Statistics](#statistics)
- [WebSocket API](#websocket-api)
- [Server-Sent Events (SSE)](#server-sent-events-sse)
- [Error Handling](#error-handling)

---

## REST API

All REST API endpoints return JSON responses in the following format:

```json
{
  "success": true,
  "data": { ... }
}
```

Error responses:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

### Health Check

Check if the API server is running.

**Endpoint:** `GET /health`

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "ok",
    "timestamp": "2025-01-15T10:30:00Z"
  }
}
```

---

### Graph Endpoints

#### Get Full Graph

Retrieve the complete causal graph with all nodes and edges.

**Endpoint:** `GET /api/v1/graph`

**Response:**
```json
{
  "success": true,
  "data": {
    "nodes": [
      {
        "data": {
          "id": "drift-001",
          "label": "IAM Policy Change",
          "type": "drift",
          "resourceType": "aws_iam_policy",
          "severity": "critical",
          "metadata": {
            "oldValue": "...",
            "newValue": "...",
            "user": "john.doe@example.com"
          }
        }
      }
    ],
    "edges": [
      {
        "data": {
          "id": "edge-001",
          "source": "drift-001",
          "target": "iam-role-001",
          "label": "caused"
        }
      }
    ]
  }
}
```

#### Get Graph Nodes (Paginated)

Retrieve nodes with pagination support.

**Endpoint:** `GET /api/v1/graph/nodes`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 1000) - Items per page

**Response:**
```json
{
  "success": true,
  "data": {
    "nodes": [ ... ],
    "page": 1,
    "limit": 50,
    "total": 150
  }
}
```

#### Get Graph Edges (Paginated)

Retrieve edges with pagination support.

**Endpoint:** `GET /api/v1/graph/edges`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 1000) - Items per page

**Response:**
```json
{
  "success": true,
  "data": {
    "edges": [ ... ],
    "page": 1,
    "limit": 50,
    "total": 120
  }
}
```

---

### Terraform State

#### Get State Metadata

Retrieve Terraform state metadata and summary.

**Endpoint:** `GET /api/v1/state`

**Response:**
```json
{
  "success": true,
  "data": {
    "version": 4,
    "terraform_version": "1.5.0",
    "serial": 42,
    "lineage": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "resource_count": 45,
    "last_modified": "2025-01-15T10:00:00Z"
  }
}
```

#### Get State Resources

List all resources in Terraform state.

**Endpoint:** `GET /api/v1/state/resources`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 1000) - Items per page

**Response:**
```json
{
  "success": true,
  "data": {
    "resources": [
      {
        "address": "aws_iam_policy.example",
        "type": "aws_iam_policy",
        "name": "example",
        "provider": "provider[\"registry.terraform.io/hashicorp/aws\"]",
        "mode": "managed",
        "attributes": { ... }
      }
    ],
    "page": 1,
    "limit": 50,
    "total": 45
  }
}
```

#### Get Single Resource

Retrieve detailed information about a specific resource.

**Endpoint:** `GET /api/v1/state/resources/:address`

**Response:**
```json
{
  "success": true,
  "data": {
    "address": "aws_iam_policy.example",
    "type": "aws_iam_policy",
    "name": "example",
    "attributes": { ... }
  }
}
```

---

### Falco Events

#### Get Events

List Falco security events with filtering.

**Endpoint:** `GET /api/v1/events`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 1000) - Items per page
- `severity` (string) - Filter by severity (critical/high/medium/low)
- `provider` (string) - Filter by cloud provider (aws/gcp/kubernetes)

**Response:**
```json
{
  "success": true,
  "data": {
    "events": [
      {
        "id": "event-001",
        "timestamp": "2025-01-15T10:15:30Z",
        "severity": "critical",
        "rule": "Unauthorized Process Execution",
        "output": "...",
        "output_fields": {
          "container.id": "abc123",
          "proc.name": "bash",
          "user.name": "root"
        }
      }
    ],
    "page": 1,
    "limit": 50,
    "total": 25
  }
}
```

#### Get Single Event

Retrieve detailed information about a specific event.

**Endpoint:** `GET /api/v1/events/:id`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "event-001",
    "timestamp": "2025-01-15T10:15:30Z",
    "severity": "critical",
    "rule": "Unauthorized Process Execution",
    "output": "...",
    "output_fields": { ... },
    "tags": ["runtime", "container"]
  }
}
```

---

### Drift Alerts

#### Get Drift Alerts

List detected drift alerts with filtering.

**Endpoint:** `GET /api/v1/drifts`

**Query Parameters:**
- `page` (int, default: 1) - Page number
- `limit` (int, default: 50, max: 1000) - Items per page
- `severity` (string) - Filter by severity
- `resource_type` (string) - Filter by resource type

**Response:**
```json
{
  "success": true,
  "data": {
    "drifts": [
      {
        "id": "drift-001",
        "timestamp": "2025-01-15T10:00:00Z",
        "severity": "critical",
        "resource_type": "aws_iam_policy",
        "resource_name": "example-policy",
        "resource_id": "ANPAI23HZ27SI6FQMGNQ2",
        "field": "policy_document",
        "old_value": "...",
        "new_value": "...",
        "user_identity": {
          "type": "IAMUser",
          "principal_id": "AIDAI23HZ27SI6FQMGNQ2",
          "arn": "arn:aws:iam::123456789012:user/john.doe",
          "account_id": "123456789012",
          "user_name": "john.doe"
        }
      }
    ],
    "page": 1,
    "limit": 50,
    "total": 12
  }
}
```

#### Get Single Drift

Retrieve detailed information about a specific drift alert.

**Endpoint:** `GET /api/v1/drifts/:id`

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "drift-001",
    "timestamp": "2025-01-15T10:00:00Z",
    "severity": "critical",
    "resource_type": "aws_iam_policy",
    "resource_name": "example-policy",
    "field": "policy_document",
    "old_value": "...",
    "new_value": "...",
    "user_identity": { ... },
    "related_events": [ ... ]
  }
}
```

---

### Statistics

Get comprehensive statistics about the system.

**Endpoint:** `GET /api/v1/stats`

**Response:**
```json
{
  "success": true,
  "data": {
    "graph": {
      "total_nodes": 45,
      "total_edges": 38,
      "node_types": {
        "drift": 12,
        "falco": 15,
        "resource": 18
      }
    },
    "drifts": {
      "total": 12,
      "severity_counts": {
        "critical": 3,
        "high": 5,
        "medium": 3,
        "low": 1
      }
    },
    "events": {
      "total": 15,
      "last_24h": 8
    },
    "severity_breakdown": {
      "critical": 25.0,
      "high": 41.7,
      "medium": 25.0,
      "low": 8.3
    },
    "top_resource_types": [
      {"type": "aws_iam_policy", "count": 5},
      {"type": "aws_iam_role", "count": 4},
      {"type": "kubernetes_service_account", "count": 3}
    ]
  }
}
```

---

## WebSocket API

Real-time bidirectional communication for interactive operations.

**Endpoint:** `ws://localhost:8080/ws`

### Connection

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('Connected');
  // Subscribe to topics
  ws.send(JSON.stringify({
    type: 'subscribe',
    topic: 'drifts'
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log('Received:', message);
};
```

### Message Format

**Client → Server:**
```json
{
  "type": "subscribe" | "unsubscribe" | "ping" | "query",
  "topic": "all" | "drifts" | "events" | "state" | "stats",
  "payload": { ... }
}
```

**Server → Client:**
```json
{
  "type": "drift" | "event" | "state_change" | "pong" | "welcome",
  "topic": "drifts",
  "timestamp": "2025-01-15T10:00:00Z",
  "payload": { ... }
}
```

### Message Types

#### Subscribe
Subscribe to a specific topic for real-time updates.

```json
{
  "type": "subscribe",
  "topic": "drifts"
}
```

**Topics:**
- `all` - Subscribe to all events
- `drifts` - Drift alerts only
- `events` - Falco events only
- `state` - Terraform state changes
- `stats` - Statistics updates

#### Unsubscribe
Unsubscribe from a topic.

```json
{
  "type": "unsubscribe",
  "topic": "drifts"
}
```

#### Ping
Send heartbeat ping (client-initiated).

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

### Server Messages

#### Welcome
Sent immediately after connection is established.

```json
{
  "type": "welcome",
  "payload": {
    "message": "Connected to TFDrift-Falco WebSocket",
    "client_id": "client-uuid-here"
  }
}
```

#### Drift Alert
Real-time drift detection notification.

```json
{
  "type": "drift",
  "topic": "drifts",
  "timestamp": "2025-01-15T10:00:00Z",
  "payload": {
    "severity": "critical",
    "resource_type": "aws_iam_policy",
    "resource_id": "...",
    "old_value": "...",
    "new_value": "..."
  }
}
```

#### Falco Event
Real-time Falco security event.

```json
{
  "type": "falco",
  "topic": "events",
  "timestamp": "2025-01-15T10:15:30Z",
  "payload": {
    "severity": "critical",
    "rule": "Unauthorized Process Execution",
    "output": "..."
  }
}
```

---

## Server-Sent Events (SSE)

Unidirectional real-time streaming from server to client.

**Endpoint:** `GET /api/v1/stream`

### Connection

```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/stream');

eventSource.addEventListener('connected', (event) => {
  const data = JSON.parse(event.data);
  console.log('Stream connected:', data);
});

eventSource.addEventListener('drift', (event) => {
  const data = JSON.parse(event.data);
  console.log('New drift:', data);
});

eventSource.onerror = (error) => {
  console.error('SSE error:', error);
};
```

### Event Types

#### connected
Sent immediately upon successful connection.

```
event: connected
data: {"stream_id":"stream-uuid-here","message":"Connected to TFDrift-Falco SSE stream"}
```

#### drift
Drift alert event.

```
event: drift
data: {"severity":"critical","resource_type":"aws_iam_policy","resource_id":"..."}
```

#### falco
Falco security event.

```
event: falco
data: {"severity":"critical","rule":"Unauthorized Process Execution","output":"..."}
```

#### state_change
Terraform state change notification.

```
event: state_change
data: {"resource":"aws_iam_policy.example","action":"modified"}
```

#### keep-alive
Periodic comment to keep connection alive (every 30 seconds).

```
: keep-alive
```

---

## Error Handling

### HTTP Status Codes

- `200 OK` - Successful request
- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error
- `503 Service Unavailable` - Service temporarily unavailable

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource with ID 'xyz' not found"
  }
}
```

### Common Error Codes

- `VALIDATION_ERROR` - Invalid request parameters
- `NOT_FOUND` - Requested resource doesn't exist
- `INTERNAL_ERROR` - Unexpected server error
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `UNAUTHORIZED` - Authentication required

### WebSocket Error Codes

Standard WebSocket close codes:
- `1000` - Normal closure
- `1001` - Going away (server shutdown)
- `1002` - Protocol error
- `1003` - Unsupported data
- `1006` - Abnormal closure
- `1008` - Policy violation

---

## Rate Limiting

Currently, no rate limiting is enforced. This may be added in future versions.

---

## CORS Configuration

CORS is configured to allow requests from:
- `http://localhost:5173` (Vite dev server)
- `http://localhost:3000` (Alternative dev server)

---

## Examples

### Complete REST API Example

```bash
# Health check
curl http://localhost:8080/health

# Get graph data
curl http://localhost:8080/api/v1/graph

# Get drifts with filtering
curl "http://localhost:8080/api/v1/drifts?severity=critical&page=1&limit=10"

# Get statistics
curl http://localhost:8080/api/v1/stats
```

### Complete WebSocket Example

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  // Subscribe to drift alerts
  ws.send(JSON.stringify({ type: 'subscribe', topic: 'drifts' }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);

  if (message.type === 'drift') {
    console.log('New drift detected:', message.payload);
    // Update UI with new drift
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

### Complete SSE Example

```javascript
const eventSource = new EventSource('http://localhost:8080/api/v1/stream');

eventSource.addEventListener('connected', (event) => {
  console.log('Connected:', JSON.parse(event.data));
});

eventSource.addEventListener('drift', (event) => {
  const drift = JSON.parse(event.data);
  // Handle new drift alert
  notifyUser(drift);
});

eventSource.addEventListener('falco', (event) => {
  const falcoEvent = JSON.parse(event.data);
  // Handle Falco security event
  logSecurityEvent(falcoEvent);
});

eventSource.onerror = (error) => {
  console.error('SSE error:', error);
  // EventSource will auto-reconnect
};
```

---

## Frontend Integration

### TanStack Query Hooks

```typescript
import { useGraph, useDrifts, useStats } from './api/hooks';

function MyComponent() {
  // Fetch graph data
  const { data: graph, isLoading, error } = useGraph();

  // Fetch drifts with filtering
  const { data: drifts } = useDrifts({
    severity: 'critical',
    page: 1,
    limit: 20
  });

  // Fetch stats
  const { data: stats } = useStats();

  return <div>...</div>;
}
```

### WebSocket Hook

```typescript
import { useWebSocket } from './api/websocket';

function RealtimeComponent() {
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
      // Handle drift alert
      console.log('New drift:', lastMessage.payload);
    }
  }, [lastMessage]);

  return <div>Status: {isConnected ? 'Connected' : 'Disconnected'}</div>;
}
```

### SSE Hook

```typescript
import { useSSE, useDriftAlerts } from './api/sse';

function EventStreamComponent() {
  const { isConnected, events } = useSSE();
  const { driftAlerts } = useDriftAlerts();

  return (
    <div>
      <div>Stream: {isConnected ? 'Active' : 'Inactive'}</div>
      <ul>
        {driftAlerts.map(drift => (
          <li key={drift.id}>{drift.severity}: {drift.resource_type}</li>
        ))}
      </ul>
    </div>
  );
}
```

---

## Version

API Version: **v1**
Last Updated: **January 2025**

For issues or questions, please open an issue on GitHub.
