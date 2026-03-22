# TFDrift-Falco Operations Runbook

**Version**: v0.8.0
**Last Updated**: 2026-03-22
**Audience**: SRE / Platform Engineering / On-Call Engineers

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Health Monitoring](#health-monitoring)
3. [Incident Response Playbooks](#incident-response-playbooks)
4. [Troubleshooting Guide](#troubleshooting-guide)
5. [Scaling Guidelines](#scaling-guidelines)
6. [Backup and Recovery](#backup-and-recovery)
7. [Log Analysis](#log-analysis)
8. [Common Alert Responses](#common-alert-responses)
9. [Maintenance Procedures](#maintenance-procedures)
10. [Escalation Matrix](#escalation-matrix)

---

## System Overview

### Components

| Component | Description | Port | Health Endpoint |
|-----------|------------|------|-----------------|
| API Server | REST API (Chi router) | 8080 | `GET /health` |
| SSE Handler | Server-Sent Events stream | 8080 | `/api/v1/stream` |
| WebSocket | Bidirectional real-time | 8080 | `/ws` |
| Falco Subscriber | gRPC client to Falco | - | Falco gRPC port 5060 |
| GraphDB | In-memory graph store | - | Internal |
| UI (React) | Dashboard frontend | 5173 (dev) | Served by API or CDN |

### Dependencies

| Dependency | Required | Purpose |
|-----------|----------|---------|
| Falco (v0.35+) | Yes | Event source via gRPC |
| Terraform State | Yes | Resource baseline |
| AWS/GCP/Azure API | Yes | Cloud resource metadata |
| Prometheus | No | Metrics collection |

### Key Environment Variables

```
TFDRIFT_API_AUTH_JWT_SECRET  - JWT signing key (required if auth enabled)
AWS_PROFILE                 - AWS credentials profile
AWS_REGION                  - Default AWS region
GOOGLE_APPLICATION_CREDENTIALS - GCP service account key path
AZURE_SUBSCRIPTION_ID       - Azure subscription
```

---

## Health Monitoring

### Health Check Endpoint

```bash
# Basic health check
curl -s http://localhost:8080/health | jq .

# Expected response
{
  "status": "healthy",
  "version": "0.8.0",
  "timestamp": "2026-03-22T10:00:00Z"
}
```

### Key Metrics to Monitor

| Metric | Warning Threshold | Critical Threshold | Action |
|--------|------------------|-------------------|--------|
| API Response Time (p95) | > 500ms | > 2s | Check GraphDB size, scale up |
| Memory Usage | > 70% | > 90% | Check for memory leaks, restart |
| CPU Usage | > 70% | > 90% | Scale horizontally |
| SSE Active Connections | > 100 | > 500 | Scale horizontally |
| Event Processing Lag | > 30s | > 5m | Check Falco connection |
| Error Rate (5xx) | > 1% | > 5% | Check logs, investigate |
| Falco gRPC Connection | Reconnecting | Disconnected > 5m | Check Falco service |

### Kubernetes Health Probes

```yaml
# Already configured in Helm chart
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
```

---

## Incident Response Playbooks

### INC-001: API Server Down

**Severity**: P1
**Symptoms**: Health check failing, 502/503 from ingress, no API responses

**Steps**:
1. Check pod status: `kubectl get pods -l app.kubernetes.io/name=tfdrift-falco`
2. Check pod logs: `kubectl logs -l app.kubernetes.io/name=tfdrift-falco --tail=100`
3. Check pod events: `kubectl describe pod <pod-name>`
4. If OOMKilled: increase memory limits in values.yaml, redeploy
5. If CrashLoopBackOff: check config.yaml validity, check Falco connectivity
6. If ImagePullBackOff: verify image tag, check registry credentials

**Resolution**: Restart pod or fix underlying issue, verify with health check.

---

### INC-002: High Memory Usage

**Severity**: P2
**Symptoms**: Memory > 90%, OOM kills, slow responses

**Steps**:
1. Check memory: `kubectl top pod -l app.kubernetes.io/name=tfdrift-falco`
2. Check graph size: `curl -s http://localhost:8080/api/v1/graph/stats | jq .`
3. If graph is very large (>10K nodes): review event ingestion rate
4. Check for memory leaks in Go runtime:
   ```bash
   # If pprof enabled
   curl http://localhost:8080/debug/pprof/heap > heap.prof
   go tool pprof heap.prof
   ```
5. Restart pod if immediate relief needed: `kubectl rollout restart deployment <name>`

**Resolution**: Scale up memory limits, optimize graph cleanup, add TTL for old events.

---

### INC-003: Falco Connection Lost

**Severity**: P2
**Symptoms**: No new events being detected, Falco subscriber errors in logs

**Steps**:
1. Check Falco pod: `kubectl get pods -n falco`
2. Check Falco gRPC port:
   ```bash
   kubectl exec -it <tfdrift-pod> -- nc -zv falco 5060
   ```
3. Check NetworkPolicy allows egress to Falco namespace
4. Verify Falco config has gRPC output enabled:
   ```yaml
   # /etc/falco/falco.yaml
   grpc:
     enabled: true
     bind_address: "0.0.0.0:5060"
   grpc_output:
     enabled: true
   ```
5. Check TFDrift-Falco logs for reconnection attempts

**Resolution**: Fix Falco connectivity, TFDrift-Falco will auto-reconnect.

---

### INC-004: Cloud API Authentication Failure

**Severity**: P2
**Symptoms**: Discovery endpoints returning errors, state refresh failing

**Steps**:
1. Check service account permissions:
   ```bash
   # AWS
   kubectl exec -it <pod> -- aws sts get-caller-identity
   # GCP
   kubectl exec -it <pod> -- gcloud auth list
   ```
2. Verify IRSA/Workload Identity annotations on ServiceAccount
3. Check if credentials have expired (temporary tokens)
4. Review IAM policy for required permissions

**Resolution**: Update credentials/IAM configuration, restart pod.

---

### INC-005: Rate Limiting Blocking Legitimate Traffic

**Severity**: P3
**Symptoms**: 429 responses for legitimate API clients

**Steps**:
1. Check current rate limit config:
   ```bash
   curl -s http://localhost:8080/api/v1/config | jq '.data.api.rate_limit'
   ```
2. Check headers on 429 response: `Retry-After`, `X-RateLimit-Remaining`
3. Identify affected client (IP or API key name from logs)
4. Increase limits in config.yaml if needed:
   ```yaml
   api:
     rate_limit:
       requests_per_minute: 120  # Increase from default 60
       burst_size: 20
   ```
5. Redeploy with new config

**Resolution**: Adjust rate limits, add separate limits for CI/automation keys.

---

## Troubleshooting Guide

### Decision Tree

```
Problem: API not responding
├── Can you reach the pod?
│   ├── No → Check k8s networking, service, ingress
│   └── Yes → Is /health returning 200?
│       ├── No → Check pod logs for startup errors
│       └── Yes → Is the specific endpoint failing?
│           ├── Auth error (401) → Check JWT/API key config
│           ├── Rate limited (429) → Wait for Retry-After or increase limits
│           ├── Server error (500) → Check logs for stack trace
│           └── Timeout → Check if GraphDB query is too large

Problem: No events being detected
├── Is Falco running and healthy?
│   ├── No → Fix Falco first
│   └── Yes → Can TFDrift reach Falco gRPC?
│       ├── No → Check NetworkPolicy, DNS, port
│       └── Yes → Are events appearing in Falco?
│           ├── No → Check Falco rules and plugin config
│           └── Yes → Check TFDrift subscriber logs for parse errors

Problem: Drift not detected for known change
├── Is the cloud provider enabled in config?
├── Is the event type mapped in resource_mapper?
├── Is the resource in Terraform state?
└── Check event parser logs for the specific event name
```

### Common Log Messages

| Log Message | Meaning | Action |
|-------------|---------|--------|
| `Connected to Falco gRPC` | Successful connection | None (info) |
| `Falco connection lost, reconnecting...` | gRPC disconnected | Check Falco health |
| `Failed to parse event` | Unknown event format | Check event source field |
| `Rate limit exceeded for client` | Client hit rate limit | May need to increase limits |
| `JWT validation failed` | Invalid/expired token | Client needs new token |
| `State refresh failed` | Can't read Terraform state | Check state backend access |

---

## Scaling Guidelines

### Vertical Scaling

| Load Profile | CPU | Memory | Notes |
|-------------|-----|--------|-------|
| Small (< 100 events/min) | 100m-250m | 128Mi-256Mi | Default settings |
| Medium (100-1000 events/min) | 250m-500m | 256Mi-512Mi | Increase GraphDB capacity |
| Large (1000+ events/min) | 500m-1000m | 512Mi-1Gi | Consider horizontal scaling |

### Horizontal Scaling (HPA)

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 75
```

Note: Each replica maintains its own in-memory graph. For shared state across replicas, an external store (Redis/PostgreSQL) would be needed (future enhancement).

---

## Backup and Recovery

### Configuration Backup

```bash
# Backup ConfigMap
kubectl get configmap <release>-tfdrift-falco-config -o yaml > config-backup.yaml

# Backup Secret
kubectl get secret <release>-tfdrift-falco-secret -o yaml > secret-backup.yaml

# Backup Helm values
helm get values <release> > values-backup.yaml
```

### Recovery Procedure

```bash
# From Helm values backup
helm upgrade <release> charts/tfdrift-falco -f values-backup.yaml

# Or from scratch
helm install <release> charts/tfdrift-falco \
  --set config.auth.jwtSecret="<secret>" \
  --set config.providers.aws.state.s3Bucket="<bucket>"
```

### State Recovery

TFDrift-Falco uses in-memory graph storage. After a restart:
1. Terraform state is reloaded automatically from the configured backend
2. Graph is rebuilt from current state
3. Historical events are lost (they exist in Falco/SIEM logs)
4. Real-time event detection resumes immediately after Falco reconnection

---

## Log Analysis

### Log Format (JSON)

```json
{
  "level": "info",
  "msg": "Drift detected",
  "time": "2026-03-22T10:00:00Z",
  "provider": "aws",
  "resource_type": "aws_security_group",
  "resource_id": "sg-12345",
  "severity": "critical",
  "user": "arn:aws:iam::123456789012:user/admin"
}
```

### Useful Log Queries

```bash
# All errors in last hour
kubectl logs -l app.kubernetes.io/name=tfdrift-falco --since=1h | jq 'select(.level == "error")'

# Drift events by severity
kubectl logs -l app.kubernetes.io/name=tfdrift-falco --since=24h | jq 'select(.msg == "Drift detected") | {severity, resource_type, user}'

# Auth failures
kubectl logs -l app.kubernetes.io/name=tfdrift-falco --since=1h | jq 'select(.msg | contains("auth")) | {msg, time}'

# Rate limit events
kubectl logs -l app.kubernetes.io/name=tfdrift-falco --since=1h | jq 'select(.msg | contains("rate limit"))'
```

---

## Common Alert Responses

### Alert: TFDriftFalcoDown

**Condition**: Health check failing for > 2 minutes

**Response**:
1. Follow INC-001 playbook
2. Escalate to P1 if not resolved in 15 minutes

### Alert: TFDriftFalcoHighMemory

**Condition**: Memory > 80% for > 5 minutes

**Response**:
1. Follow INC-002 playbook
2. Consider immediate scale-up

### Alert: TFDriftFalcoHighErrorRate

**Condition**: 5xx error rate > 5% for > 5 minutes

**Response**:
1. Check logs for recurring errors
2. Check recent deployments (rollback if needed)
3. Check cloud API status pages

### Alert: TFDriftFalcoCriticalDrift

**Condition**: Critical severity drift event detected

**Response**:
1. This is a security-relevant infrastructure change
2. Verify with the user identified in the event
3. If unauthorized: initiate security incident response
4. If authorized but unplanned: update Terraform code

---

## Maintenance Procedures

### Rolling Update

```bash
# Update image tag
helm upgrade <release> charts/tfdrift-falco --set image.tag=v0.8.1

# Verify rollout
kubectl rollout status deployment/<release>-tfdrift-falco

# Rollback if needed
kubectl rollout undo deployment/<release>-tfdrift-falco
```

### Config Change

```bash
# Update values and apply
helm upgrade <release> charts/tfdrift-falco -f values.yaml

# Force pod restart for config changes
kubectl rollout restart deployment/<release>-tfdrift-falco
```

### API Key Rotation

```bash
# Generate new API key via API
curl -X POST http://localhost:8080/api/v1/auth/api-keys \
  -H "X-API-Key: <admin-key>" \
  -H "Content-Type: application/json" \
  -d '{"name": "new-ci-key", "scopes": ["read"]}'

# Revoke old API key
curl -X DELETE http://localhost:8080/api/v1/auth/api-keys \
  -H "X-API-Key: <admin-key>" \
  -H "Content-Type: application/json" \
  -d '{"name": "old-ci-key"}'
```

### JWT Secret Rotation

1. Generate new secret
2. Update Kubernetes secret or values.yaml
3. Rolling restart (old tokens will be invalidated)
4. All clients need to re-authenticate

---

## Escalation Matrix

| Level | Condition | Response Time | Contact |
|-------|-----------|---------------|---------|
| P1 | Service down, no drift detection | 15 min | On-call SRE |
| P2 | Degraded (high latency, partial failure) | 1 hour | SRE team |
| P3 | Non-critical (UI issues, rate limiting) | 4 hours | Platform team |
| P4 | Enhancement request, documentation | Next sprint | Engineering |

> **Note**: Update the contact information above with your organization's actual escalation contacts.
