# Quickstart Guide

Get TFDrift-Falco up and running in 30 minutes.

---

## Prerequisites

### Required

- **AWS Account** with CloudTrail enabled
- **Terraform** (v1.0+) managing your infrastructure
- **Kubernetes Cluster** (for Falco deployment)
  - EKS, GKE, AKS, or local (minikube, kind)
- **kubectl** configured
- **Helm 3** installed

### Optional (for monitoring)

- **Grafana** (v9.0+)
- **Prometheus** (v2.40+)

---

## Step 1: Clone the Repository

```bash
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

---

## Step 2: Configure AWS CloudTrail

Ensure CloudTrail is enabled and logging management events.

### Check CloudTrail Status

```bash
aws cloudtrail describe-trails --region us-east-1
```

### Create Trail (if needed)

```bash
aws cloudtrail create-trail \
  --name tfdrift-trail \
  --s3-bucket-name my-cloudtrail-logs \
  --is-multi-region-trail
```

---

## Step 3: Deploy Falco

### Add Falco Helm Repository

```bash
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update
```

### Install Falco

```bash
helm install falco falcosecurity/falco \
  --namespace falco --create-namespace \
  --set falcosidekick.enabled=true \
  --set falcosidekick.webui.enabled=true
```

### Verify Installation

```bash
kubectl get pods -n falco
# Expected output:
# NAME                              READY   STATUS    RESTARTS   AGE
# falco-xxxxx                       1/1     Running   0          30s
# falco-falcosidekick-xxxxx         1/1     Running   0          30s
# falco-falcosidekick-ui-xxxxx      1/1     Running   0          30s
```

---

## Step 4: Deploy TFDrift Rules

### Apply TFDrift Falco Rules

```bash
kubectl apply -f rules/tfdrift-rules.yaml
```

### Verify Rules Loaded

```bash
kubectl logs -n falco -l app.kubernetes.io/name=falco | grep tfdrift
# Expected: "Loaded tfdrift rules successfully"
```

---

## Step 5: Configure TFDrift Detector

### Create Configuration File

```bash
cp config-example.yaml config.yaml
```

### Edit Configuration

```yaml
# config.yaml
aws:
  account_id: "123456789012"
  region: "us-east-1"
  cloudtrail:
    enabled: true
    poll_interval: "1m"

terraform:
  backend:
    type: "s3"
    config:
      bucket: "my-terraform-state"
      key: "prod/terraform.tfstate"
      region: "us-east-1"

services:
  enabled:
    - ec2
    - iam
    - s3
    - vpc
    - rds

falco:
  endpoint: "unix:///var/run/falco/falco.sock"

logging:
  level: "info"
  format: "json"
```

---

## Step 6: Create AWS IAM Policy

TFDrift requires permissions to read CloudTrail and Terraform state.

### Create IAM Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudtrail:LookupEvents",
        "cloudtrail:DescribeTrails"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-terraform-state",
        "arn:aws:s3:::my-terraform-state/*"
      ]
    }
  ]
}
```

### Create IAM Role (for EKS)

```bash
eksctl create iamserviceaccount \
  --name tfdrift-detector \
  --namespace default \
  --cluster my-eks-cluster \
  --attach-policy-arn arn:aws:iam::123456789012:policy/TFDriftPolicy \
  --approve
```

---

## Step 7: Deploy TFDrift Detector

### Create Kubernetes Deployment

```bash
kubectl apply -f deployments/detector/deployment.yaml
```

### Verify Detector Running

```bash
kubectl get pods -l app=tfdrift-detector
# Expected output:
# NAME                                READY   STATUS    RESTARTS   AGE
# tfdrift-detector-xxxxx              1/1     Running   0          30s
```

### Check Logs

```bash
kubectl logs -l app=tfdrift-detector --tail=50
# Expected output:
# {"level":"info","msg":"TFDrift Detector started","version":"v0.2.0-beta"}
# {"level":"info","msg":"Polling CloudTrail","interval":"1m"}
```

---

## Step 8: Test Drift Detection

### Make a Manual Change

```bash
# Example: Change EC2 instance type
aws ec2 modify-instance-attribute \
  --instance-id i-0123456789abcdef0 \
  --instance-type t3.small
```

### Verify Drift Detected

```bash
# Check TFDrift logs
kubectl logs -l app=tfdrift-detector --tail=20
# Expected output:
# {"level":"warning","msg":"Drift detected","service":"ec2","event":"ModifyInstanceAttribute","resource":"i-0123456789abcdef0"}

# Check Falco logs
kubectl logs -n falco -l app.kubernetes.io/name=falco --tail=20
# Expected output:
# 07:30:00.000000000: Warning EC2 Instance Type Changed (instance=i-0123456789abcdef0 from=t3.micro to=t3.small user=admin)
```

---

## Step 9: Set Up Grafana (Optional)

### Deploy Grafana

```bash
helm install grafana grafana/grafana \
  --namespace monitoring --create-namespace \
  --set adminPassword=admin
```

### Access Grafana

```bash
kubectl port-forward -n monitoring svc/grafana 3000:80
# Open browser: http://localhost:3000
# Login: admin / admin
```

### Import TFDrift Dashboard

1. In Grafana, go to **Dashboards** → **Import**
2. Upload `dashboards/grafana-tfdrift-overview.json`
3. Select Prometheus data source
4. Click **Import**

---

## Step 10: Configure Alerting (Optional)

### Slack Alerts

```bash
kubectl create secret generic falcosidekick-config \
  --from-literal=slack.webhookurl=https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
  -n falco

helm upgrade falco falcosecurity/falco \
  --namespace falco \
  --set falcosidekick.config.slack.webhookurl=https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
  --reuse-values
```

### Test Alert

Make another manual change and verify Slack notification.

---

## Troubleshooting

### TFDrift Detector Not Starting

**Symptom:** Pod in `CrashLoopBackOff`

**Solution:**
```bash
# Check logs for errors
kubectl logs -l app=tfdrift-detector --previous

# Common issues:
# 1. AWS credentials not configured
kubectl describe pod -l app=tfdrift-detector
# Check for ServiceAccount annotation

# 2. Config file invalid
kubectl get configmap tfdrift-config -o yaml
# Verify YAML syntax
```

### No Drift Detected

**Symptom:** Manual changes not triggering alerts

**Solution:**
```bash
# 1. Verify CloudTrail is logging events
aws cloudtrail lookup-events --max-results 10

# 2. Check TFDrift is polling CloudTrail
kubectl logs -l app=tfdrift-detector | grep "Polling CloudTrail"

# 3. Verify Falco rules loaded
kubectl exec -n falco -it falco-xxxxx -- falco --list
# Should show tfdrift rules
```

### Terraform State Not Found

**Symptom:** Error "Terraform state not found"

**Solution:**
```bash
# 1. Verify S3 bucket exists and is accessible
aws s3 ls s3://my-terraform-state/prod/terraform.tfstate

# 2. Check IAM permissions
aws sts get-caller-identity
# Ensure IAM role has S3 read permissions
```

---

## Next Steps

### Production Deployment

1. [Configure multi-region monitoring →](deployment.md#multi-region)
2. [Set up high availability →](deployment.md#high-availability)
3. [Enable Prometheus metrics →](deployment.md#metrics)

### Customize Rules

1. [Adjust Falco rule severity →](falco-setup.md#severity-levels)
2. [Add custom service coverage →](contributing.md#adding-services)
3. [Tune false positive filtering →](falco-setup.md#filters)

### Advanced Features

1. [Multi-account monitoring →](deployment.md#multi-account)
2. [Export drift history →](deployment.md#drift-history)
3. [Integrate with CI/CD →](deployment.md#cicd-integration)

---

## Resources

- [Full Deployment Guide →](deployment.md)
- [Falco Setup Guide →](falco-setup.md)
- [Troubleshooting Guide →](deployment.md#troubleshooting)
- [GitHub Issues →](https://github.com/higakikeita/tfdrift-falco/issues)
