# driftwire Deployment Guide

This guide covers different deployment methods for driftwire in production environments.

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
docker build -t driftwire:latest .
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
  --name driftwire \
  -v $(pwd)/config.yaml:/config/config.yaml:ro \
  -v ~/.aws:/root/.aws:ro \
  -e AWS_REGION=us-east-1 \
  driftwire:latest \
  --config /config/config.yaml
```

#### Step 4: View Logs

```bash
docker logs -f driftwire
```

---

## Docker Compose Deployment

Docker Compose is the recommended method for running driftwire with all dependencies.

### Architecture

The Docker Compose stack includes:
- **Falco**: Runtime security with CloudTrail plugin
- **driftwire**: Main drift detection service

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
- `config.yaml` - driftwire configuration
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
docker-compose logs -f driftwire
docker-compose logs -f falco
```

### Step 5: Test the Setup

```bash
# Trigger a test CloudTrail event
aws ec2 modify-instance-attribute \
  --instance-id i-1234567890abcdef0 \
  --disable-api-termination

# Check logs for drift detection
docker-compose logs driftwire | grep -i "drift"
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

For Kubernetes environments, deploy driftwire as a Deployment with Falco as a DaemonSet.

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

### Step 2: Create ConfigMap for driftwire Config

```bash
kubectl create configmap driftwire-config \
  --from-file=config.yaml=./config.yaml \
  --namespace driftwire
```

### Step 3: Create Secret for AWS Credentials

```bash
kubectl create secret generic aws-credentials \
  --from-file=credentials=$HOME/.aws/credentials \
  --from-file=config=$HOME/.aws/config \
  --namespace driftwire
```

### Step 4: Deploy driftwire

Create `k8s/deployment.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: driftwire
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: driftwire
  namespace: driftwire
  labels:
    app: driftwire
spec:
  replicas: 1
  selector:
    matchLabels:
      app: driftwire
  template:
    metadata:
      labels:
        app: driftwire
    spec:
      containers:
      - name: driftwire
        image: driftwire:latest
        imagePullPolicy: IfNotPresent
        args:
          - --config
          - /config/config.yaml
        env:
        - name: AWS_REGION
          value: "us-east-1"
        - name: DRIFTWIRE_FALCO_HOSTNAME
          value: "falco-grpc.falco.svc.cluster.local"
        - name: DRIFTWIRE_FALCO_PORT
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
          name: driftwire-config
      - name: aws-credentials
        secret:
          secretName: aws-credentials
---
apiVersion: v1
kind: Service
metadata:
  name: driftwire
  namespace: driftwire
spec:
  selector:
    app: driftwire
  ports:
  - name: http
    port: 8080
    targetPort: 8080
```

Apply the deployment:

```bash
kubectl apply -f k8s/deployment.yaml
```

### Step 5: Verify Deployment

```bash
# Check pod status
kubectl get pods -n driftwire

# View logs
kubectl logs -f deployment/driftwire -n driftwire

# Check Falco connection
kubectl exec -it deployment/driftwire -n driftwire -- /bin/sh
# Inside container:
# nc -zv falco-grpc.falco.svc.cluster.local 5060
```

---

## Systemd Service

For running driftwire as a native systemd service on Linux.

### Step 1: Build and Install Binary

```bash
# Build for Linux
make build-linux

# Install binary
sudo cp bin/driftwire-linux-amd64 /usr/local/bin/driftwire
sudo chmod +x /usr/local/bin/driftwire
```

### Step 2: Create Configuration Directory

```bash
sudo mkdir -p /etc/driftwire
sudo cp config.yaml /etc/driftwire/
sudo chmod 600 /etc/driftwire/config.yaml
```

### Step 3: Create Systemd Service File

Create `/etc/systemd/system/driftwire.service`:

```ini
[Unit]
Description=driftwire Terraform Drift Detection
Documentation=https://github.com/higakikeita/driftwire
After=network.target falco.service
Requires=falco.service

[Service]
Type=simple
User=driftwire
Group=driftwire
WorkingDirectory=/var/lib/driftwire
Environment="AWS_REGION=us-east-1"
Environment="HOME=/var/lib/driftwire"
ExecStart=/usr/local/bin/driftwire --config /etc/driftwire/config.yaml
Restart=on-failure
RestartSec=10s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=driftwire

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/driftwire

[Install]
WantedBy=multi-user.target
```

### Step 4: Create Service User

```bash
sudo useradd -r -s /bin/false -d /var/lib/driftwire driftwire
sudo mkdir -p /var/lib/driftwire/.aws
sudo cp ~/.aws/credentials /var/lib/driftwire/.aws/
sudo cp ~/.aws/config /var/lib/driftwire/.aws/
sudo chown -R driftwire:driftwire /var/lib/driftwire
```

### Step 5: Enable and Start Service

```bash
# Reload systemd
sudo systemctl daemon-reload

# Enable service to start on boot
sudo systemctl enable driftwire

# Start service
sudo systemctl start driftwire

# Check status
sudo systemctl status driftwire

# View logs
sudo journalctl -u driftwire -f
```

---

## Production Considerations

### High Availability

#### Multiple Replicas
- Run multiple driftwire instances for redundancy
- Use different availability zones
- Share state via external storage

#### Load Balancing
- Not required (event-driven architecture)
- Each instance processes events independently

### Monitoring

#### Metrics Collection
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
- Restrict Falco gRPC access to driftwire only
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
  driftwire:
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
- driftwire is stateless (no persistent data)
- Configuration is stored in config.yaml
- Terraform state is external (S3/remote)

#### Recovery Procedures
1. Stop driftwire service
2. Update configuration if needed
3. Restart service
4. Verify Falco connection
5. Test with sample drift event

### Scaling

#### Vertical Scaling
- Increase CPU/memory for high event volumes
- Monitor resource usage

#### Horizontal Scaling
- Run multiple driftwire instances
- Each instance processes all events independently
- No coordination needed

### Maintenance

#### Updates
```bash
# Docker
docker-compose pull
docker-compose up -d

# Systemd
sudo systemctl stop driftwire
sudo cp new-binary /usr/local/bin/driftwire
sudo systemctl start driftwire
```

#### Configuration Changes
```bash
# Validate config
driftwire --config config.yaml --dry-run

# Apply changes
# Docker Compose
docker-compose restart driftwire

# Systemd
sudo systemctl restart driftwire
```

### Troubleshooting

#### Common Issues

**driftwire can't connect to Falco**
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

# Check driftwire logs
docker-compose logs driftwire | grep -i event
```

**High memory usage**
- Reduce event buffer size
- Enable rate limiting
- Check for memory leaks (report issue)

---

## Next Steps

After deployment:
1. [Configure alerts](../examples/config.yaml)
2. [Access the React Dashboard UI](./quickstart.md#step-10-access-the-dashboard-ui-v060)
3. [Review security best practices](./SECURITY.md)
4. [Join the community](https://github.com/higakikeita/driftwire/discussions)

## Support

For deployment issues:
- [GitHub Issues](https://github.com/higakikeita/driftwire/issues)
- [Documentation](https://github.com/higakikeita/driftwire/tree/main/docs)
- [Community Discussions](https://github.com/higakikeita/driftwire/discussions)
