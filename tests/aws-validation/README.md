# TFDrift-Falco AWS Validation

End-to-end validation of tfdrift-falco in a real AWS environment.
Uses Sysdig dev account (230446364776) via Okta SSO.

## Prerequisites

- AWS CLI v2 + `okta-aws-cli` configured
- Terraform >= 1.5
- kubectl
- Helm 3
- Profile `draios-dev-developer` authenticated

```bash
# Authenticate via Okta
okta-aws-cli web \
  --org-domain sysdig.okta.com \
  --oidc-client-id 0oa8hgzcpqASrDfBP697 \
  --aws-acct-fed-app-id 0oa1r7zeadpxUCPSL697 \
  --profile draios-dev-developer \
  --write-aws-credentials

# Verify
aws sts get-caller-identity --profile draios-dev-developer
```

## Directory Structure

```
tests/aws-validation/
├── phase1-resources/     # VPC, EC2, S3, IAM (drift targets)
├── phase2-eks/           # EKS cluster for Falco
├── phase3-deploy/        # Falco Helm values + tfdrift K8s manifests
├── scripts/
│   ├── create-state-bucket.sh   # One-time: create TF state bucket
│   ├── discover-cloudtrail.sh   # Find existing CloudTrail config
│   ├── trigger-drift.sh         # Introduce intentional drift
│   ├── revert-drift.sh          # Undo drift changes
│   └── cleanup-all.sh           # Destroy everything
└── README.md
```

## Phase 1: Create Test Resources (~5 min)

```bash
# 1. Create TF state bucket (one-time)
./scripts/create-state-bucket.sh

# 2. Deploy test resources
cd phase1-resources
terraform init
terraform apply

# Note the outputs - you'll need vpc_id and subnet_ids for Phase 2
terraform output
```

**Resources created**: VPC, 2 subnets, IGW, SG, EC2 (t4g.micro), S3 bucket, IAM role.
**Estimated cost**: ~$5/month

## Phase 2: Create EKS Cluster (~15 min)

```bash
# 1. Discover CloudTrail config
./scripts/discover-cloudtrail.sh
# Note the CloudTrail S3 bucket name

# 2. Deploy EKS
cd phase2-eks
terraform init
terraform apply \
  -var="vpc_id=$(cd ../phase1-resources && terraform output -raw vpc_id)" \
  -var='subnet_ids='"$(cd ../phase1-resources && terraform output -json subnet_ids)" \
  -var="cloudtrail_bucket=YOUR_CLOUDTRAIL_BUCKET"

# 3. Configure kubectl
$(terraform output -raw kubeconfig_command)
kubectl get nodes
```

**Resources created**: EKS cluster, 1 spot node (t4g.medium), IAM roles.
**Estimated cost**: ~$75/month (EKS control plane $0.10/hr + spot node)

## Phase 3: Deploy Falco + TFDrift-Falco (~10 min)

```bash
cd phase3-deploy

# 1. Install Falco with CloudTrail plugin
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update

# Edit falco-values.yaml: set sqsQueue or s3Bucket for CloudTrail
vim falco-values.yaml

helm install falco falcosecurity/falco \
  -n falco --create-namespace \
  -f falco-values.yaml

# 2. Verify Falco is running
kubectl -n falco get pods
kubectl -n falco logs -l app.kubernetes.io/name=falco --tail=20

# 3. Deploy tfdrift-falco
kubectl apply -f tfdrift-falco.yaml

# 4. Verify
kubectl -n tfdrift get pods
kubectl -n tfdrift logs -l app=tfdrift-falco --tail=20
```

## Phase 4: Trigger & Validate Drift (~20 min)

```bash
# 1. Trigger intentional drift
./scripts/trigger-drift.sh

# 2. Wait 5-15 min for CloudTrail event propagation
# CloudTrail events have a typical delay of 5-15 minutes

# 3. Check Falco logs for CloudTrail events
kubectl -n falco logs -l app.kubernetes.io/name=falco --tail=50 | grep -i "security\|instance\|bucket"

# 4. Check tfdrift-falco for drift detection
kubectl -n tfdrift logs -l app=tfdrift-falco --tail=50

# 5. (Optional) Access tfdrift UI
kubectl -n tfdrift port-forward svc/tfdrift-falco 8080:8080
# Open http://localhost:8080
```

**Expected drift detections**:
1. SG: unauthorized ingress rule on port 8443
2. EC2: monitoring enabled (TF says false)
3. EC2: unexpected `ManualChange` tag
4. S3: versioning Suspended (TF says Enabled)

## Phase 5: Cleanup

```bash
# Revert drift only (keep infra)
./scripts/revert-drift.sh

# Destroy everything
./scripts/cleanup-all.sh
```

## Troubleshooting

### Falco not receiving CloudTrail events
- Verify the CloudTrail S3 bucket name is correct
- Check if SQS notifications are set up for the bucket
- For org-level trails, the bucket may be in a different account
- Try direct S3 polling mode (edit `falco-values.yaml`, remove sqsQueue, set s3Bucket)

### tfdrift-falco can't read TF state
- Verify the node role has S3 access to `tfdrift-validation-state`
- Check IRSA annotation if using service account roles

### EKS nodes not joining
- Check node group status: `aws eks describe-nodegroup --cluster-name tfdrift-val-eks --nodegroup-name tfdrift-val-eks-ng --profile draios-dev-developer`
- Spot capacity may be unavailable — try on-demand by changing `capacity_type` in phase2-eks/main.tf

### Permission denied on CloudTrail bucket
- Org-level CloudTrail buckets may have bucket policies restricting access
- You may need to add the node role ARN to the bucket policy in the org management account
