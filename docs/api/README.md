# API Documentation

TFDrift-Falco provides comprehensive APIs for interacting with the system, accessing Terraform state, receiving alerts, and monitoring drift detection in real-time.

## API Overview

The TFDrift-Falco API suite includes:

- **REST API**: Standard HTTP endpoints for querying state, retrieving alerts, and accessing system statistics
- **WebSocket API**: Real-time bidirectional communication for live event streaming
- **Server-Sent Events (SSE)**: Server push notifications for drift detection events and alerts

### Base URLs

```
REST API:    http://localhost:8080/api/v1
WebSocket:   ws://localhost:8080/ws
SSE Stream:  http://localhost:8080/api/v1/stream
```

### Authentication

API requests require authentication. Include your API token in the request:

```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  http://localhost:8080/api/v1/health
```

### API Versioning

The current API version is `v1`. Version information is included in REST API URLs (`/api/v1/...`).

### Request/Response Format

All REST API responses follow the standard APIResponse wrapper format:

**Success Response:**
```json
{
  "success": true,
  "data": {
    "message": "Operation completed successfully",
    "result": { /* data */ }
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

## API Documentation

### [REST API Documentation](rest-api.md)

Comprehensive reference for all REST API endpoints, including:
- Health checks and system status
- Graph querying and visualization
- Terraform state management
- Falco event retrieval
- Drift alert queries
- System statistics

### [WebSocket API Documentation](websocket.md)

Real-time event streaming via WebSocket, including:
- Connection and authentication
- Event subscription
- Message formats
- Reconnection handling

### [Server-Sent Events (SSE) Documentation](sse.md)

Server push notifications for drift detection, including:
- SSE endpoint configuration
- Event types and formats
- Retry handling
- Integration patterns

## Common Use Cases

### Monitoring Drift Detection

Use WebSocket or SSE to receive real-time notifications when infrastructure drift is detected:

```bash
# WebSocket connection
wscat -c "ws://localhost:8080/ws" \
  -H "Authorization: Bearer YOUR_API_TOKEN"

# Or use SSE
curl -N -H "Authorization: Bearer YOUR_API_TOKEN" \
  http://localhost:8080/api/v1/stream
```

### Querying Terraform State

Retrieve current infrastructure state and relationships:

```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  http://localhost:8080/api/v1/terraform/state
```

### Getting Drift Alerts

Retrieve past alerts and detection history:

```bash
curl -H "Authorization: Bearer YOUR_API_TOKEN" \
  http://localhost:8080/api/v1/alerts?limit=50
```

## Error Handling

All APIs include consistent error handling with standardized error codes and messages. See individual documentation pages for specific error codes and handling guidelines.

## Rate Limiting

API requests are rate-limited to ensure fair usage. Current limits:
- REST API: 100 requests per minute
- WebSocket: 1 connection per authenticated user
- SSE: Unlimited (push-based)

## Support

For API questions and issues, refer to the main [TFDrift-Falco documentation](../README.md) or create an issue on the project repository.
