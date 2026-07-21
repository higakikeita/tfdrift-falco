# TFDrift-Falco Validation Lab (bastion + EKS + ECS Fargate)

A **lean, short-lived** AWS environment for end-to-end verification of TFDrift-Falco
drift detection against real CloudTrail events. Deliberately minimal — provision it,
generate a few manual (non-IaC) changes, confirm tfdrift detects them, then destroy.

> This is intentionally separate from `terraform/production-like-environment/`
> (a 157-resource behemoth with RDS/ElastiCache/ALB and a stale `errored.tfstate`).
> Use this lab for routine verification; it's cheaper and safer to tear down.

## What it provisions

| Component | Detail |
|---|---|
| VPC | 2 AZ, **no NAT gateway** — workloads run in public subnets with public IPs (avoids consuming an Elastic IP; the shared account is at its 5/5 EIP limit) |
| Bastion | EC2 `t3.micro` in a public subnet, reachable **only via SSM Session Manager** — no key pair, no inbound SSH, no open `:22` |
| EKS | 1 managed node (`t3.small`), public API endpoint, creator = cluster admin |
| ECS Fargate | Cluster + one minimal `nginx` service (a live workload to drift) |

## ⚠️ Cost & lifecycle

This spends real money while it exists. Rough run-rate (ap-northeast-1):

- EKS control plane: **~$0.10/hour** (~$2.4/day)
- NAT gateway: ~$0.062/hour + data
- 1× t3.small node + 1× t3.micro bastion + Fargate task: a few $/day

**Plan on ~$5–8/day.** The intended flow is **apply → verify → `terraform destroy` the same day.** Do not leave it running.

## Prerequisites

- AWS credentials for the target account (`aws configure` / SSO). Verify: `aws sts get-caller-identity`
- Terraform ≥ 1.5 (repo uses 1.13.3)
- For bastion access: AWS CLI Session Manager plugin

## Usage

```bash
cd terraform/validation-lab
cp terraform.tfvars.example terraform.tfvars   # optional; defaults are fine

terraform init
terraform plan            # review — EKS + node group can take ~15 min to apply
terraform apply

# Connect
aws ssm start-session --target "$(terraform output -raw bastion_instance_id)"
eval "$(terraform output -raw eks_kubeconfig_command)" && kubectl get nodes

# TEAR DOWN when done (important — stops the meter)
terraform destroy
```

## Generating drift to verify tfdrift

Once tfdrift-falco is watching this account's CloudTrail, make manual changes the
rules in `rules/terraform_drift.yaml` watch for, e.g.:

```bash
# Security-group ingress change (AuthorizeSecurityGroupIngress)
aws ec2 authorize-security-group-ingress --group-id <ecs-svc-sg> \
  --protocol tcp --port 8080 --cidr 10.20.0.0/16

# ECS service scale (UpdateService) / task-def change
aws ecs update-service --cluster tfdrift-lab-ecs --service tfdrift-lab-nginx --desired-count 2

# EKS logging toggle, IAM policy attach, etc.
```

Each should surface in tfdrift as a drift event (who / when / what) — the exact
signal the flagship demo shows.

## Notes

- State is local (`terraform.tfstate`, gitignored). For a shared/long-lived env, add an S3 backend.
- The bastion uses SSM (not SSH) on purpose: no `0.0.0.0/0:22`, matching the security posture expected of a security tool's own infra.
