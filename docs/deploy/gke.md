# Google Kubernetes Engine (GKE) Deployment Guide

This guide covers deploying TFDrift-Falco on Google Kubernetes Engine (GKE) using Helm charts for production-ready deployments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [GKE Cluster Setup](#gke-cluster-setup)
3. [Helm Chart Deployment](#helm-chart-deployment)
4. [Production Values Configuration](#production-values-configuration)
5. [Workload Identity Setup](#workload-identity-setup)
6. [Cloud Armor Integration](#cloud-armor-integration)
7. [Cloud Monitoring Setup](#cloud-monitoring-setup)
8. [Network Configuration](#network-configuration)
9. [Deployment Verification](#deployment-verification)
10. [Scaling and Auto-Scaling](#scaling-and-auto-scaling)
11. [Troubleshooting](#troubleshooting)

## Prerequisites

### GCP Project and IAM

- Active GCP project with billing enabled
- Appropriate IAM roles:
  - Kubernetes Engine Admin
  - Compute Admin
  - Service Account Admin
  - Cloud Monitoring Admin (for metrics)
  - Cloud Logging Admin (for logs)

### Local Tools

```bash
# Install gcloud CLI
curl https://sdk.cloud.google.com | bash

# Install kubectl
gcloud components install kubectl

# Install Helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Verify installations
gcloud --version
kubectl version --client
helm version
```

### GCP Service Account

```bash
# Create service account for TFDrift-Falco
gcloud iam service-accounts create tfdrift-falco \
  --display-name="TFDrift-Falco Service Account"

# Grant permissions
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/compute.viewer"

gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"
```

## GKE Cluster Setup

### Create GKE Cluster

```bash
# Set project
gcloud config set project PROJECT_ID

# Create cluster
gcloud container clusters create tfdrift-falco \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2 \
  --enable-stackdriver-kubernetes \
  --enable-ip-alias \
  --enable-autorepair \
  --enable-autoupgrade \
  --enable-autoscaling \
  --min-nodes 3 \
  --max-nodes 10 \
  --enable-workload-identity \
  --workload-pool=PROJECT_ID.svc.id.goog \
  --addons HorizontalPodAutoscaling,HttpLoadBalancing,GcePersistentDiskCsiDriver

# Get cluster credentials
gcloud container clusters get-credentials tfdrift-falco --zone us-central1-a
```

### Verify Cluster Access

```bash
# Check cluster info
kubectl cluster-info

# Check nodes
kubectl get nodes

# Check default namespaces
kubectl get namespace
```

### Create Namespace

```bash
# Create namespace for TFDrift-Falco
kubectl create namespace tfdrift-falco

# Set as default namespace
kubectl config set-context --current --namespace=tfdrift-falco

# Verify
kubectl get namespace tfdrift-falco
```

## Helm Chart Deployment

### Add Helm Repository

```bash
# If hosting chart on a repository
helm repo add tfdrift https://charts.example.com
helm repo update

# Or use local chart (from project root)
cd /path/to/tfdrift-falco
```

### Deploy Using Helm

```bash
# Create values file (see section below)
cp charts/tfdrift-falco/values.yaml values-production.yaml

# Edit for your environment
# See Production Values Configuration section

# Deploy with Helm
helm install tfdrift-falco charts/tfdrift-falco \
  --namespace tfdrift-falco \
  --values values-production.yaml

# Or upgrade existing release
helm upgrade --install tfdrift-falco charts/tfdrift-falco \
  --namespace tfdrift-falco \
  --values values-production.yaml

# Check deployment status
helm status tfdrift-falco --namespace tfdrift-falco

# List releases
helm list --namespace tfdrift-falco
```

### Verify Deployment

```bash
# Check pods
kubectl get pods -n tfdrift-falco

# Check services
kubectl get svc -n tfdrift-falco

# Check deployments
kubectl get deployments -n tfdrift-falco

# Watch deployment progress
kubectl rollout status deployment/tfdrift-falco -n tfdrift-falco

# View logs
kubectl logs -n tfdrift-falco -l app=tfdrift-falco -f
```

## Production Values Configuration

### Production values.yaml

```yaml
# Helm chart values for production GKE deployment

replicaCount: 3

image:
  repository: gcr.io/PROJECT_ID/tfdrift-falco
  tag: "1.0.0"
  pullPolicy: IfNotPresent

# Google Cloud specific settings
imagePullSecrets: []

serviceAccount:
  create: true
  annotations:
    iam.gke.io/gcp-service-account: tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com
  name: tfdrift-falco

# Security context
podSecurityContext:
  fsGroup: 1000
  seccompProfile:
    type: RuntimeDefault

securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  readOnlyRootFilesystem: true
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL

# Pod disruption budget for high availability
podDisruptionBudget:
  enabled: true
  minAvailable: 2

# Pod annotations for monitoring
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"

# Service configuration
service:
  type: ClusterIP
  port: 8080
  annotations:
    cloud.google.com/neg: '{"ingress": true}'

# Ingress for external access
ingress:
  enabled: true
  className: "gce"
  annotations:
    kubernetes.io/ingress.class: "gce"
    kubernetes.io/ingress.global-static-ip-name: "tfdrift-falco-ip"
    networking.gke.io/managed-certificates: "tfdrift-falco-cert"
    kubernetes.io/ingress.allow-http: "false"
  hosts:
    - host: tfdrift-falco.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: tfdrift-falco-tls
      hosts:
        - tfdrift-falco.example.com

# Managed Certificate (Google Cloud)
managedCertificate:
  enabled: true
  domains:
    - tfdrift-falco.example.com

# Network Policy
networkPolicy:
  enabled: true
  ingressNamespaces: []

# Resource requests and limits
resources:
  requests:
    cpu: 250m
    memory: 256Mi
  limits:
    cpu: 1000m
    memory: 1Gi

# Horizontal Pod Autoscaler
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80

# Pod Disruption Budget
pdb:
  enabled: true
  minAvailable: 2

# Node affinity
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
      - weight: 100
        podAffinityTerm:
          labelSelector:
            matchExpressions:
              - key: app
                operator: In
                values:
                  - tfdrift-falco
          topologyKey: kubernetes.io/hostname

# Health checks
healthCheck:
  enabled: true
  path: /health
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

# Application configuration
config:
  port: 8080
  logLevel: info
  logFormat: json

  auth:
    enabled: true
    jwtIssuer: "tfdrift-falco"
    jwtExpiry: "24h"

  rateLimit:
    enabled: true
    requestsPerMinute: 1000
    burstSize: 100

  falco:
    hostname: "falco.falco"
    port: 5060

  providers:
    gcp:
      enabled: true
      projects:
        - "PROJECT_ID"
      state:
        backend: "gcs"
        gcsBucket: "terraform-state-PROJECT_ID"
        gcsPrefix: "prod"

# Secrets management
existingSecret: ""
googleSecretManager:
  enabled: true
  projectId: "PROJECT_ID"
  secrets:
    - name: jwt-secret
      key: tfdrift-falco-jwt-secret

# ServiceMonitor for Prometheus
serviceMonitor:
  enabled: true
  interval: 30s
  labels:
    prometheus: kube-prometheus

# Pod Monitor
podMonitor:
  enabled: true
  interval: 30s

# Vertical Pod Autoscaler
vpa:
  enabled: true
  updateMode: "auto"

# Backup configuration
backup:
  enabled: false
  schedule: "0 2 * * *"
  retentionDays: 30
```

## Workload Identity Setup

Workload Identity allows your GKE pods to authenticate to GCP services without managing keys.

### Configure Workload Identity

```bash
# Link Kubernetes ServiceAccount to GCP ServiceAccount
gcloud iam service-accounts add-iam-policy-binding \
  tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:PROJECT_ID.svc.id.goog[tfdrift-falco/tfdrift-falco]"

# Annotate Kubernetes ServiceAccount
kubectl annotate serviceaccount tfdrift-falco \
  --namespace tfdrift-falco \
  iam.gke.io/gcp-service-account=tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com
```

### Verify Workload Identity

```bash
# Test from pod
kubectl run -it --image google/cloud-sdk:slim \
  --serviceaccount tfdrift-falco \
  --namespace tfdrift-falco \
  test-workload-identity -- bash

# Inside pod, verify credentials
gcloud auth list

# Access GCS bucket
gsutil ls gs://terraform-state-PROJECT_ID
```

## Cloud Armor Integration

Cloud Armor protects your application with advanced DDoS and WAF capabilities.

### Create Cloud Armor Policy

```bash
# Create policy
gcloud compute security-policies create tfdrift-falco-policy \
  --description="Cloud Armor policy for TFDrift-Falco"

# Allow traffic from known IPs
gcloud compute security-policies rules create 1000 \
  --security-policy tfdrift-falco-policy \
  --action allow \
  --expression "origin.ip in ['203.0.113.0/24']"

# Block traffic from suspicious IPs
gcloud compute security-policies rules create 2000 \
  --security-policy tfdrift-falco-policy \
  --action deny-403 \
  --expression "evaluatePreconfiguredExpr('xss-stable')"

# Rate limiting
gcloud compute security-policies rules create 3000 \
  --security-policy tfdrift-falco-policy \
  --action rate-based-ban \
  --rate-limit-options "rate-limit-threshold-count=100,rate-limit-threshold-interval-sec=60" \
  --ban-duration-sec=600

# Default rule
gcloud compute security-policies rules create 65000 \
  --security-policy tfdrift-falco-policy \
  --action allow
```

### Link to Load Balancer

```bash
# Create backend service with Cloud Armor
gcloud compute backend-services update tfdrift-falco-backend \
  --security-policy tfdrift-falco-policy \
  --global
```

## Cloud Monitoring Setup

### Enable Cloud Monitoring

```bash
# Cloud Monitoring is automatically enabled with GKE
# Verify Prometheus metrics are scraped
kubectl get servicemonitor -n tfdrift-falco
```

### Create Monitoring Dashboard

```bash
# Create dashboard
gcloud monitoring dashboards create --config-from-file=- <<'EOF'
{
  "displayName": "TFDrift-Falco GKE",
  "mosaicLayout": {
    "columns": 12,
    "tiles": [
      {
        "width": 6,
        "height": 4,
        "widget": {
          "title": "CPU Usage",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "metric.type=\"kubernetes.io/container/cpu/core_usage_time\" resource.type=\"k8s_container\" metadata.system_labels.top_level_controller_name=\"tfdrift-falco\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_RATE"
                    }
                  }
                }
              }
            ]
          }
        }
      },
      {
        "xPos": 6,
        "width": 6,
        "height": 4,
        "widget": {
          "title": "Memory Usage",
          "xyChart": {
            "dataSets": [
              {
                "timeSeriesQuery": {
                  "timeSeriesFilter": {
                    "filter": "metric.type=\"kubernetes.io/container/memory/used_bytes\" resource.type=\"k8s_container\" metadata.system_labels.top_level_controller_name=\"tfdrift-falco\"",
                    "aggregation": {
                      "alignmentPeriod": "60s",
                      "perSeriesAligner": "ALIGN_MEAN"
                    }
                  }
                }
              }
            ]
          }
        }
      }
    ]
  }
}
EOF
```

### Create Alerts

```bash
# Alert for pod failures
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="TFDrift-Falco Pod Failures" \
  --condition-display-name="Pods Not Running" \
  --condition-threshold-value=2 \
  --condition-threshold-duration=300s \
  --condition-threshold-filter='resource.type="k8s_pod" AND metadata.user_labels.app="tfdrift-falco" AND metric.type="kubernetes.io/pod/running"'

# Alert for high CPU
gcloud alpha monitoring policies create \
  --notification-channels=CHANNEL_ID \
  --display-name="TFDrift-Falco High CPU" \
  --condition-display-name="CPU > 80%" \
  --condition-threshold-value=80 \
  --condition-threshold-duration=300s \
  --condition-threshold-filter='resource.type="k8s_container" AND metadata.system_labels.top_level_controller_name="tfdrift-falco" AND metric.type="kubernetes.io/container/cpu/request_utilization"'
```

## Network Configuration

### Network Policy

```yaml
# Example network policy
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tfdrift-falco-network-policy
  namespace: tfdrift-falco
spec:
  podSelector:
    matchLabels:
      app: tfdrift-falco
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              name: falco
      ports:
        - protocol: TCP
          port: 5060
    - to:
        - podSelector: {}
      ports:
        - protocol: TCP
          port: 53
        - protocol: UDP
          port: 53
```

### Firewall Rules

```bash
# Allow traffic to cluster
gcloud compute firewall-rules create allow-tfdrift-falco \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags kubernetes-cluster

# Restrict egress
gcloud compute firewall-rules create deny-all-egress \
  --action DENY \
  --direction EGRESS \
  --priority 1000
```

## Deployment Verification

### Check Deployment Status

```bash
# Check all resources
kubectl get all -n tfdrift-falco

# Check Pod status
kubectl get pods -n tfdrift-falco -o wide

# Describe pod for details
kubectl describe pod <pod-name> -n tfdrift-falco

# Check logs
kubectl logs -n tfdrift-falco -l app=tfdrift-falco --tail=50 -f

# Check events
kubectl get events -n tfdrift-falco --sort-by='.lastTimestamp'
```

### Test Connectivity

```bash
# Port forward for testing
kubectl port-forward -n tfdrift-falco svc/tfdrift-falco 8080:8080

# Test health endpoint
curl http://localhost:8080/health

# Test API
curl http://localhost:8080/api/v1/drifts
```

## Scaling and Auto-Scaling

### Manual Scaling

```bash
# Scale replicas
kubectl scale deployment tfdrift-falco \
  --namespace tfdrift-falco \
  --replicas=5

# Check scaling
kubectl get deployment tfdrift-falco -n tfdrift-falco
```

### Horizontal Pod Autoscaling (HPA)

HPA is configured in the Helm values. Monitor it with:

```bash
# Check HPA status
kubectl get hpa -n tfdrift-falco

# Describe HPA
kubectl describe hpa tfdrift-falco -n tfdrift-falco

# Watch HPA behavior
kubectl get hpa -n tfdrift-falco --watch
```

### Vertical Pod Autoscaling (VPA)

If VPA is enabled, it recommends resource requests:

```bash
# Check VPA recommendations
kubectl describe vpa tfdrift-falco -n tfdrift-falco

# Apply recommendations
kubectl patch vpa tfdrift-falco \
  --namespace tfdrift-falco \
  --type merge \
  --patch '{"spec":{"updatePolicy":{"updateMode":"Auto"}}}'
```

### Node Pool Scaling

```bash
# Check node pools
gcloud container node-pools list --cluster tfdrift-falco

# Update node pool autoscaling
gcloud container node-pools update default-pool \
  --cluster tfdrift-falco \
  --min-nodes 3 \
  --max-nodes 10 \
  --enable-autoscaling
```

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl describe pod <pod-name> -n tfdrift-falco

# Check image pull
kubectl describe pod <pod-name> -n tfdrift-falco | grep -A 5 "Events:"

# Verify image exists in GCR
gcloud container images list --repository=gcr.io/PROJECT_ID

# Check logs
kubectl logs <pod-name> -n tfdrift-falco
```

### Workload Identity Issues

```bash
# Verify annotation
kubectl get serviceaccount tfdrift-falco -n tfdrift-falco -o yaml | grep gcp-service-account

# Check IAM binding
gcloud iam service-accounts get-iam-policy tfdrift-falco@PROJECT_ID.iam.gserviceaccount.com

# Test from pod
kubectl run -it --image google/cloud-sdk:slim \
  --serviceaccount tfdrift-falco \
  --namespace tfdrift-falco \
  test -- gcloud auth list
```

### Connection to Falco Failing

```bash
# Check if Falco namespace/pod exists
kubectl get pods -n falco

# Test connectivity
kubectl run -it --image alpine:latest \
  --namespace tfdrift-falco \
  test -- nc -zv falco.falco 5060
```

### GKE Cluster Issues

```bash
# Check cluster status
gcloud container clusters describe tfdrift-falco --zone us-central1-a

# Check node status
kubectl get nodes -o wide

# View cluster logs
gcloud logging read "resource.type=k8s_cluster" --limit 50

# Diagnose cluster
gcloud container operations describe <operation-id> --zone us-central1-a
```

For more information, refer to the [GKE documentation](https://cloud.google.com/kubernetes-engine/docs).
