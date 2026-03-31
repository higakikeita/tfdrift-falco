#!/usr/bin/env bash
# Complete cleanup of all validation resources
# Run this when done with validation to avoid ongoing costs
set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-draios-dev-developer}"
AWS_REGION="${AWS_REGION:-ap-northeast-1}"
BASE_DIR="$(dirname "$0")/.."

echo "=== TFDrift-Falco Validation Cleanup ==="
echo "This will destroy ALL validation resources."
echo ""
read -p "Are you sure? (yes/no): " confirm
if [ "${confirm}" != "yes" ]; then
  echo "Aborted."
  exit 0
fi

# Phase 3: Remove K8s resources
echo ""
echo "[Phase 3] Removing Kubernetes deployments..."
kubectl delete -f "${BASE_DIR}/phase3-deploy/tfdrift-falco.yaml" --ignore-not-found 2>/dev/null || true
helm uninstall falco -n falco 2>/dev/null || echo "  Falco helm release not found"
kubectl delete namespace tfdrift --ignore-not-found 2>/dev/null || true
kubectl delete namespace falco --ignore-not-found 2>/dev/null || true

# Phase 2: Destroy EKS
echo ""
echo "[Phase 2] Destroying EKS cluster..."
cd "${BASE_DIR}/phase2-eks"
if [ -f "terraform.tfstate" ] || [ -d ".terraform" ]; then
  terraform destroy -auto-approve
else
  echo "  No terraform state found, skipping"
fi

# Phase 1: Destroy test resources
echo ""
echo "[Phase 1] Destroying test resources..."
cd "${BASE_DIR}/phase1-resources"
if [ -f "terraform.tfstate" ] || [ -d ".terraform" ]; then
  terraform destroy -auto-approve
else
  echo "  No terraform state found, skipping"
fi

# State bucket (optional)
echo ""
read -p "Delete TF state bucket (tfdrift-validation-state)? (yes/no): " del_bucket
if [ "${del_bucket}" = "yes" ]; then
  echo "Emptying and deleting state bucket..."
  aws s3 rm "s3://tfdrift-validation-state" --recursive --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null || true
  aws s3api delete-bucket --bucket "tfdrift-validation-state" --profile "${AWS_PROFILE}" --region "${AWS_REGION}" 2>/dev/null || true
  echo "  Deleted"
fi

echo ""
echo "=== Cleanup Complete ==="
echo "Estimated monthly cost saved: ~\$75-100 (EKS) + \$5-10 (EC2/S3)"
