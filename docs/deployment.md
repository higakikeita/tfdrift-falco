# TFDrift-Falco Deployment Guide

This guide covers different deployment methods for TFDrift-Falco in production environments.

## Table of Contents

- [Docker Deployment](#docker-deployment)
- [Docker Compose Deployment](#docker-compose-deployment)
- [Kubernetes Deployment](#kubernetes-deployment)
- [Systemd Service](#systemd-service)
- [Production Considerations](#production-considerations)

---

## Docker Deployment

### Quick Start with Docker

#### Step 1: Build the Docker Image

```bash
# Using Make
make docker-build

# Or directly with Docker
docker build -t tfdrift-falco:latest .
```

#### Step 2: Prepare Configuration

Create a `config.yaml` file (see [examples/config.yaml](../examples/config.yaml)):

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "prod/terraform.tfstate"

falco:
  enabled: true
  hostname: "falco"  # Use Docker service name
  port: 5060

notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
```

#### Step 3: Run the Container

```bash
docker run -d \
  --name tfdrift-falco \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  -v ~/.aws:/root/.aws:ro \
  -e AWS_REGION=us-east-1 \
  tfdrift-falco:latest \
  --config /config/config.yaml
```

#### Step 4: View Logs

```bash
docker logs -f tfdrift-falco
```

---

## Docker Compose Deployment

Docker Compose is the recommended method for running TFDrift-Falco with all dependencies.

### Architecture

The Docker Compose stack includes:
- **Falco**: Runtime security with CloudTrail plugin
- **TFDrift-Falco**: Main drift detection service

### Step 1: Configure Environment Variables

Create a `.env` file:

```bash
# AWS Configuration
AWS_REGION=us-east-1
CLOUDTRAIL_S3_BUCKET=my-cloudtrail-logs
TERRAFORM_STATE_DIR=./terraform

# Slack Webhook (optional)
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL

# Timezone
TZ=America/New_York
```

### Step 2: Prepare Configuration Files

Ensure these files exist:
- `config.yaml` - TFDrift configuration
- `deployments/falco/falco.yaml` - Falco configuration (provided)
- `rules/terraform_drift.yaml` - Falco rules (provided)

Update `config.yaml` to use Docker service names:

```yaml
falco:
  enabled: true
  hostname: "falco"  # Docker Compose service name
  port: 5060
```

### Step 3: Start the Stack

```bash
# Using Make
make docker-compose-up

# Or directly
docker-compose up -d
```

### Step 4: Verify Services

```bash
# Check running containers
make docker-compose-ps

# View logs
make docker-compose-logs

# Or for specific service
docker-compose logs -f tfdrift
docker-compose logs -f falco
```

### Step 5: Test the Setup

```bash
# Trigger a test CloudTrail event
aws ec2 modify-instance-attribute \
  --instance-id i-1234567890abcdef0 \
  --disable-api-termination

# Check logs for drift detection
docker-compose logs tfdrift | grep -i "drift"
```

### Management Commands

```bash
# Stop services
make docker-compose-down

# Restart services
make docker-compose-restart

# Rebuild and restart
make docker-compose-build

# View status
make docker-compose-ps
```

---

## Kubernetes Deployment

For Kubernetes environments, deploy TFDrift-Falco as a Deployment with Falco as a DaemonSet.

### Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Helm 3.x (optional, for Falco installation)

### Step 1: Install Falco via Helm

```bash
# Add Falco Helm repository
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update

# Install Falco with CloudTrail plugin
helm install falco falcosecurity/falco \
  --namespace falco --create-namespace \
  --set falco.grpc.enabled=true \
  --set falco.grpcOutput.enabled=true \
  --set collectors.cloudtrail.enabled=true \
  --set collectors.cloudtrail.s3Bucket=my-cloudtrail-bucket \
  --set collectors.cloudtrail.sqsQueue=my-cloudtrail-queue
```

### Step 2: Create ConfigMap for TFDrift Config

```bash
kubectl create configmap tfdrift-config \
  --from-file=config.yaml=./config.yaml \
  --namespace tfdrift
```

### Step 3: Create Secret for AWS Credentials

```bash
kubectl create secret generic aws-credentials \
  --from-file=credentials=$HOME/.aws/credentials \
  --from-file=config=$HOME/.aws/config \
  --namespace tfdrift
```

### Step 4: Deploy TFDrift-Falco

Create `k8s/deployment.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: tfdrift
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tfdrift-falco
  namespace: tfdrift
  labels:
    app: tfdrift-falco
spec:
  replicas: 1
  selector:
    matchLabels:
      app: tfdrift-falco
  template:
    metadata:
      labels:
        app: tfdrift-falco
    spec:
      containers:
      - name: tfdrift
        image: tfdrift-falco:latest
        imagePullPolicy: IfNotPresent
        args:
          - --config
          - /config/config.yaml
        env:
        - name: AWS_REGION
          value: "us-east-1"
        - name: TFDRIFT_FALCO_HOSTNAME
          value: "falco-grpc.falco.svc.cluster.local"
        - name: TFDRIFT_FALCO_PORT
          value: "5060"
        volumeMounts:
        - name: config
          mountPath: /config
          readOnly: true
        - name: aws-credentials
          mountPath: /root/.aws
          readOnly: true
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
      volumes:
      - name: config
        configMap:
          name: tfdrift-config
      - name: aws-credentials
        secret:
          secretName: aws-credentials
---
apiVersion: v1
kind: Service
metadata:
  name: tfdrift-falco
  namespace: tfdrift
spec:
  selector:
    app: tfdrift-falco
  ports:
  - name: metrics
    port: 9090
    targetPort: 9090
```

Apply the deployment:

```bash
kubectl apply -f k8s/deployment.yaml
```

### Step 5: Verify Deployment

```bash
# Check pod status
kubectl get pods -n tfdrift

# View logs
kubectl logs -f deployment/tfdrift-falco -n tfdrift

# Check Falco connection
kubectl exec -it deployment/tfdrift-falco -n tfdrift -- /bin/sh
# Inside container:
# nc -zv falco-grpc.falco.svc.cluster.local 5060
```

---

## Systemd Service

For running TFDrift-Falco as a native systemd service on Linux.

### Step 1: Build and Install Binary

```bash
# Build for Linux
make build-linux

# Install binary
sudo cp bin/tfdrift-linux-amd64 /usr/local/bin/tfdrift
sudo chmod +x /usr/local/bin/tfdrift
```

### Step 2: Create Configuration Directory

```bash
sudo mkdir -p /etc/tfdrift
sudo cp config.yaml /etc/tfdrift/
sudo chmod 600 /etc/tfdrift/config.yaml
```

### Step 3: Create Systemd Service File

Create `/etc/systemd/system/tfdrift.service`:

```ini
[Unit]
Description=TFDrift-Falco Terraform Drift Detection
Documentation=https://github.com/keitahigaki/tfdrift-falco
After=network.target falco.service
Requires=falco.service

[Service]
Type=simple
User=tfdrift
Group=tfdrift
WorkingDirectory=/var/lib/tfdrift
Environment="AWS_REGION=us-east-1"
Environment="HOME=/var/lib/tfdrift"
ExecStart=/usr/local/bin/tfdrift --config /etc/tfdrift/config.yaml
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=tfdrift

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/tfdrift

[Install]
WantedBy=multi-user.target
```

### Step 4: Create Service User

```bash
sudo useradd -r -s /bin/false -d /var/lib/tfdrift tfdrift
sudo mkdir -p /var/lib/tfdrift/.aws
sudo cp ~/.aws/credentials /var/lib/tfdrift/.aws/
sudo cp ~/.aws/config /var/lib/tfdrift/.aws/
sudo chown -R tfdrift:tfdrift /var/lib/tfdrift
```

### Step 5: Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable tfdrift

# Start service
sudo systemctl start tfdrift

# Check status
sudo systemctl status tfdrift

# View logs
sudo journalctl -u tfdrift -f
```

---

## Production Considerations

### High Availability

#### Multiple Replicas
- Run multiple TFDrift instances for redundancy
- Use different availability zones
- Share state via external storage

#### Load Balancing
- Not required (event-driven architecture)
- Each instance processes events independently

### Monitoring

#### Metrics Collection
- Expose Prometheus metrics (future feature)
- Monitor event processing latency
- Track drift detection rate

#### Logging
- Use structured JSON logging
- Send logs to centralized logging (ELK, Loki)
- Set appropriate log levels

### Security

#### AWS Credentials
- Use IAM roles instead of access keys when possible
- Rotate credentials regularly
- Use least-privilege IAM policies

#### Network Security
- Restrict Falco gRPC access to TFDrift only
- Use TLS/mTLS for Falco gRPC communication
- Place services in private subnets

#### Secrets Management
- Store Slack webhooks in secrets manager
- Use Kubernetes secrets or AWS Secrets Manager
- Never commit secrets to version control

### Performance Tuning

#### Resource Allocation
```yaml
# Docker Compose
services:
  tfdrift:
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 256M
```

#### Event Processing
- Adjust event buffer size in config
- Use SQS for CloudTrail events (lower latency)
- Enable rate limiting for notifications

### Backup and Recovery

#### State Backup
- TFDrift is stateless (no persistent data)
- Configuration is stored in config.yaml
- Terraform state is external (S3/remote)

#### Recovery Procedures
1. Stop TFDrift service
2. Update configuration if needed
3. Restart service
4. Verify Falco connection
5. Test with sample drift event

### Scaling

#### Vertical Scaling
- Increase CPU/memory for high event volumes
- Monitor resource usage

#### Horizontal Scaling
- Run multiple TFDrift instances
- Each instance processes all events independently
- No coordination needed

### Maintenance

#### Updates
```bash
# Docker
docker-compose pull
docker-compose up -d

# Systemd
sudo systemctl stop tfdrift
sudo cp new-binary /usr/local/bin/tfdrift
sudo systemctl start tfdrift
```

#### Configuration Changes
```bash
# Validate config
tfdrift --config config.yaml --dry-run

# Apply changes
# Docker Compose
docker-compose restart tfdrift

# Systemd
sudo systemctl restart tfdrift
```

### Troubleshooting

#### Common Issues

**TFDrift can't connect to Falco**
```bash
# Check Falco is running
docker-compose logs falco
# Or
sudo systemctl status falco

# Verify gRPC port
netstat -tlnp | grep 5060

# Test connection
telnet falco 5060
```

**No drift events detected**
```bash
# Check Falco rules are loaded
docker exec falco falco -L | grep terraform

# Verify CloudTrail events
aws cloudtrail lookup-events --max-results 10

# Check TFDrift logs
docker-compose logs tfdrift | grep -i event
```

**High memory usage**
- Reduce event buffer size
- Enable rate limiting
- Check for memory leaks (report issue)

---

## Next Steps

After deployment:
1. [Configure alerts](../examples/config.yaml)
2. [Set up Grafana dashboards](../dashboards/grafana/README.md)
3. [Review security best practices](./SECURITY.md)
4. [Join the community](https://github.com/keitahigaki/tfdrift-falco/discussions)

## Support

For deployment issues:
- [GitHub Issues](https://github.com/keitahigaki/tfdrift-falco/issues)
- [Documentation](https://github.com/keitahigaki/tfdrift-falco/tree/main/docs)
- [Community Discussions](https://github.com/keitahigaki/tfdrift-falco/discussions)
