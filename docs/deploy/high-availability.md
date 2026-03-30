# High Availability Deployment Guide for TFDrift-Falco

This guide covers configuring TFDrift-Falco for high availability across multiple replicas, availability zones, and cloud regions.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Multi-Replica Deployment](#multi-replica-deployment)
3. [Session Affinity](#session-affinity)
4. [Health Checks and Readiness Probes](#health-checks-and-readiness-probes)
5. [Graceful Shutdown](#graceful-shutdown)
6. [Database Considerations](#database-considerations)
7. [Caching Strategy](#caching-strategy)
8. [Monitoring and Alerting](#monitoring-and-alerting)
9. [Disaster Recovery](#disaster-recovery)
10. [Load Testing](#load-testing)

## Architecture Overview

```
Internet
   │
   ├─→ Load Balancer (Multi-Zone)
   │
   ├─→ Zone A: Node Pool
   │    ├─→ Pod 1 (Replica 1)
   │    ├─→ Pod 2 (Replica 2)
   │
   ├─→ Zone B: Node Pool
   │    ├─→ Pod 3 (Replica 3)
   │
   └─→ Zone C: Node Pool
        └─→ [Additional Pods]

   All Pods Connect to:
   ├─→ Shared Database (RDS/Azure Database/Cloud SQL)
   ├─→ Shared Cache (Redis/Memcached)
   └─→ Falco Cluster
```

## Multi-Replica Deployment

### Kubernetes Configuration

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tfdrift-falco
  namespace: tfdrift-falco
spec:
  # Minimum replicas for HA
  replicas: 3

  # Rolling update strategy
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # Add one pod at a time
      maxUnavailable: 0  # Don't kill pods until new one is ready

  # Graceful termination period
  template:
    spec:
      terminationGracePeriodSeconds: 60

      # Pod priority (prevent eviction)
      priorityClassName: high-priority

      # Pod topology spread
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: topology.kubernetes.io/zone
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              app: tfdrift-falco
        - maxSkew: 2
          topologyKey: kubernetes.io/hostname
          whenUnsatisfiable: ScheduleAnyway
          labelSelector:
            matchLabels:
              app: tfdrift-falco

      # Anti-affinity for spreading across nodes
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: app
                    operator: In
                    values:
                      - tfdrift-falco
              topologyKey: kubernetes.io/hostname

        podAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - cache
                topologyKey: kubernetes.io/hostname

      containers:
        - name: tfdrift-falco
          image: tfdrift-falco:latest

          # Resource guarantees
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 1Gi

          # Environment
          env:
            - name: REPLICA_ID
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
```

### Pod Disruption Budget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: tfdrift-falco-pdb
  namespace: tfdrift-falco
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: tfdrift-falco
  unhealthyPodEvictionPolicy: AlwaysAllow
```

## Session Affinity

For WebSocket connections and stateful interactions, session affinity ensures requests from the same client go to the same pod.

### Load Balancer Session Affinity

```yaml
# Kubernetes Service with session affinity
apiVersion: v1
kind: Service
metadata:
  name: tfdrift-falco
  namespace: tfdrift-falco
spec:
  type: ClusterIP
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 10800  # 3 hours
  ports:
    - name: http
      port: 8080
      targetPort: 8080
      protocol: TCP
  selector:
    app: tfdrift-falco
```

### AWS ALB Session Stickiness

```bash
aws elbv2 modify-target-group-attributes \
  --target-group-arn arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/tfdrift-falco/1234567890abcdef \
  --attributes \
    Key=stickiness.enabled,Value=true \
    Key=stickiness.type,Value=lb_cookie \
    Key=stickiness.lb_cookie.duration_seconds,Value=86400
```

### Azure Application Gateway Affinity

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tfdrift-falco
  annotations:
    appgw.ingress.kubernetes.io/cookie-based-affinity: "enabled"
    appgw.ingress.kubernetes.io/cookie-based-affinity-primary: "tfdrift-session"
spec:
  rules:
    - host: tfdrift-falco.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: tfdrift-falco
                port:
                  number: 8080
```

### GCP Cloud Load Balancing Session Affinity

```bash
gcloud compute backend-services update tfdrift-falco-backend \
  --session-affinity CLIENT_IP \
  --global
```

## Health Checks and Readiness Probes

### Application Health Check Endpoint

The application should expose a health check endpoint:

```go
package main

import (
  "net/http"
  "sync/atomic"
  "time"
)

var (
  readyFlag       int32
  shutdownStarted int32
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
  // Health check (always ready)
  w.Header().Set("Content-Type", "application/json")

  if atomic.LoadInt32(&shutdownStarted) == 1 {
    w.WriteHeader(http.StatusServiceUnavailable)
    w.Write([]byte(`{"status":"shutting-down"}`))
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"healthy"}`))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
  // Readiness check (dependencies ready)
  w.Header().Set("Content-Type", "application/json")

  if atomic.LoadInt32(&readyFlag) == 0 {
    w.WriteHeader(http.StatusServiceUnavailable)
    w.Write([]byte(`{"status":"not-ready"}`))
    return
  }

  if !isDatabaseConnected() || !isFalcoConnected() {
    w.WriteHeader(http.StatusServiceUnavailable)
    w.Write([]byte(`{"status":"dependencies-unavailable"}`))
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"status":"ready"}`))
}

func markReady() {
  atomic.StoreInt32(&readyFlag, 1)
}

func initiateShutdown() {
  atomic.StoreInt32(&shutdownStarted, 1)
}
```

### Kubernetes Probe Configuration

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: tfdrift-falco
spec:
  containers:
    - name: tfdrift-falco
      image: tfdrift-falco:latest

      # Startup probe (check if application started)
      startupProbe:
        httpGet:
          path: /health
          port: 8080
        failureThreshold: 30
        periodSeconds: 10
        timeoutSeconds: 5

      # Liveness probe (restart if dead)
      livenessProbe:
        httpGet:
          path: /health
          port: 8080
        initialDelaySeconds: 30
        periodSeconds: 10
        timeoutSeconds: 5
        failureThreshold: 3

      # Readiness probe (traffic only if ready)
      readinessProbe:
        httpGet:
          path: /ready
          port: 8080
        initialDelaySeconds: 10
        periodSeconds: 5
        timeoutSeconds: 3
        failureThreshold: 1
        successThreshold: 1

      # Pre-stop hook for graceful shutdown
      lifecycle:
        preStop:
          exec:
            command: ["/bin/sh", "-c", "sleep 15"]
```

## Graceful Shutdown

### Application Shutdown Handler

```go
package main

import (
  "context"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "time"
)

func main() {
  server := &http.Server{
    Addr:         ":8080",
    Handler:      mux,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  60 * time.Second,
  }

  // Graceful shutdown
  go func() {
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
    <-sigChan

    // Signal shutdown has started
    initiateShutdown()

    // Stop accepting new connections
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := server.Shutdown(ctx); err != nil {
      logger.Errorf("Server shutdown error: %v", err)
    }
  }()

  // Start server
  if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    logger.Fatalf("Server error: %v", err)
  }
}
```

### Kubernetes Termination Configuration

```yaml
spec:
  template:
    spec:
      # Time to allow graceful shutdown
      terminationGracePeriodSeconds: 60

      containers:
        - name: tfdrift-falco
          lifecycle:
            preStop:
              exec:
                # Give time for connections to drain
                command: ["/bin/sh", "-c", "sleep 15 && kill -TERM 1"]
```

## Database Considerations

### Connection Pooling

```go
// Database connection pool configuration
db := &sql.DB{
  // Use connection pool
  MaxOpenConns: 25,        // Max concurrent connections
  MaxIdleConns: 5,         // Keep-alive connections
  ConnMaxLifetime: 5 * time.Minute,
  ConnMaxIdleTime: 2 * time.Minute,
}
```

### High Availability Database

#### AWS RDS Multi-AZ

```bash
aws rds modify-db-instance \
  --db-instance-identifier tfdrift-falco-db \
  --multi-az \
  --backup-retention-period 30 \
  --preferred-backup-window "02:00-03:00" \
  --apply-immediately
```

#### Azure Database for PostgreSQL High Availability

```bash
az postgres server create \
  --name tfdrift-falco-db \
  --resource-group rg-tfdrift-falco \
  --sku-name B_Gen5_2 \
  --geo-redundant-backup Enabled \
  --backup-retention 30
```

#### GCP Cloud SQL High Availability

```bash
gcloud sql instances create tfdrift-falco-db \
  --database-version=POSTGRES_13 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --availability-type=REGIONAL \
  --backup \
  --backup-start-time=02:00
```

### Read Replicas

```bash
# AWS RDS Read Replica
aws rds create-db-instance-read-replica \
  --db-instance-identifier tfdrift-falco-db-replica \
  --source-db-instance-identifier tfdrift-falco-db \
  --db-instance-class db.t3.micro

# Azure PostgreSQL Read Replica
az postgres server replica create \
  --name tfdrift-falco-db-replica \
  --source-server tfdrift-falco-db \
  --resource-group rg-tfdrift-falco

# GCP Cloud SQL Read Replica
gcloud sql instances create tfdrift-falco-db-replica \
  --master-instance-name=tfdrift-falco-db \
  --tier=db-f1-micro \
  --region=us-east1
```

## Caching Strategy

### Redis Cache Configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: redis-config
  namespace: tfdrift-falco
data:
  redis.conf: |
    # Memory management
    maxmemory 1gb
    maxmemory-policy allkeys-lru

    # Persistence
    save 900 1
    save 300 10
    save 60 10000

    # Replication
    repl-diskless-sync yes
    repl-diskless-sync-delay 5

    # Cluster
    cluster-enabled yes
    cluster-node-timeout 15000
```

### Cache Invalidation Strategy

```go
type CacheManager struct {
  client *redis.Client
}

// Invalidate on resource changes
func (cm *CacheManager) InvalidateOnDriftChange(ctx context.Context, resourceID string) error {
  patterns := []string{
    fmt.Sprintf("drift:%s:*", resourceID),
    fmt.Sprintf("resource:%s:*", resourceID),
    "drift:list:*",
  }

  for _, pattern := range patterns {
    keys, err := cm.client.Keys(ctx, pattern).Result()
    if err != nil {
      return err
    }
    if len(keys) > 0 {
      if err := cm.client.Del(ctx, keys...).Err(); err != nil {
        return err
      }
    }
  }
  return nil
}

// Cache with TTL
func (cm *CacheManager) CacheWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
  return cm.client.Set(ctx, key, value, ttl).Err()
}
```

## Monitoring and Alerting

### Key Metrics to Monitor

```yaml
# Prometheus scrape config
scrape_configs:
  - job_name: 'tfdrift-falco'
    kubernetes_sd_configs:
      - role: pod
        namespaces:
          names:
            - tfdrift-falco
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
```

### Alert Rules

```yaml
groups:
  - name: tfdrift-falco
    interval: 30s
    rules:
      # Pod availability
      - alert: PodDownReplicaCountBelow
        expr: kube_deployment_status_replicas_available{deployment="tfdrift-falco"} < 2
        for: 5m
        annotations:
          summary: "TFDrift-Falco has fewer than 2 replicas"

      # Response time
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        annotations:
          summary: "Response time P95 > 1s"

      # Error rate
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 2m
        annotations:
          summary: "Error rate > 5%"

      # Database connection issues
      - alert: DatabaseConnectionErrors
        expr: rate(db_connection_errors_total[5m]) > 0.1
        for: 5m
        annotations:
          summary: "Database connection errors detected"
```

### Health Dashboard

Key dashboard panels:
- Pod count and status
- Request rate and latency
- Error rate
- CPU and memory usage
- Database connection pool
- Cache hit/miss ratio
- Active WebSocket connections

## Disaster Recovery

### Backup Strategy

```bash
# Kubernetes resource backups
kubectl get all -n tfdrift-falco -o yaml > backup-$(date +%Y%m%d).yaml

# Database backup
pg_dump -h <db-host> tfdrift_falco > db-backup-$(date +%Y%m%d).sql

# Helm values backup
helm get values tfdrift-falco > helm-values-backup.yaml
```

### Recovery Procedures

```bash
# Restore from backup
kubectl apply -f backup-20240101.yaml

# Restore database
psql -h <db-host> tfdrift_falco < db-backup-20240101.sql

# Redeploy with Helm
helm install tfdrift-falco charts/tfdrift-falco \
  --namespace tfdrift-falco \
  --values helm-values-backup.yaml
```

### Failover Testing

```bash
# Simulate pod failure
kubectl delete pod <pod-name> -n tfdrift-falco

# Verify auto-recovery
kubectl get pods -n tfdrift-falco --watch

# Simulate node failure
kubectl drain <node> --ignore-daemonsets --delete-emptydir-data

# Verify pod rescheduling
kubectl get pods -n tfdrift-falco -o wide
```

## Load Testing

### Capacity Testing with k6

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
};

export default function() {
  // API load
  let res = http.get('http://tfdrift-falco.example.com/api/v1/drifts');

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  // WebSocket load
  let ws = new WebSocket('ws://tfdrift-falco.example.com/ws');

  check(ws, {
    'WebSocket connected': (w) => w.readyState === WebSocket.OPEN,
  });

  ws.close();
  sleep(1);
}
```

### Testing Commands

```bash
# Run load test
k6 run load-test.js

# Run with different VUs
k6 run -u 100 -d 10m load-test.js
```

This comprehensive guide ensures TFDrift-Falco can handle production workloads with high availability and graceful degradation.
