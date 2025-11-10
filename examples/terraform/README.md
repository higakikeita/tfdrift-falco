# TFDrift-Falco Test Environment

This Terraform configuration creates a test AWS environment for validating TFDrift-Falco drift detection capabilities.

## Resources Created

- **VPC** with public subnet and internet gateway
- **EC2 Instance** (t2.micro, Amazon Linux 2023)
  - Termination protection enabled
  - Monitoring disabled
- **S3 Bucket** with:
  - Versioning enabled
  - Server-side encryption (AES256)
  - Public access blocked
- **IAM Role** with EC2 assume role policy
- **Security Group** with SSH and HTTP access

## Prerequisites

1. AWS CLI configured with credentials
2. Terraform 1.0+
3. Appropriate AWS permissions to create resources

## Usage

### Initialize Terraform

```bash
cd examples/terraform
terraform init
```

### Plan the deployment

```bash
terraform plan
```

### Apply the configuration

```bash
terraform apply
```

### Get outputs

```bash
terraform output
```

## Testing Drift Detection

After applying this configuration, you can test TFDrift-Falco by manually changing resources:

### Test Case 1: EC2 Termination Protection

```bash
# Get instance ID
INSTANCE_ID=$(terraform output -raw ec2_instance_id)

# Disable termination protection (will trigger drift alert)
aws ec2 modify-instance-attribute \
  --instance-id $INSTANCE_ID \
  --no-disable-api-termination
```

### Test Case 2: S3 Encryption

```bash
# Get bucket name
BUCKET_NAME=$(terraform output -raw s3_bucket_name)

# Remove encryption configuration (will trigger drift alert)
aws s3api delete-bucket-encryption \
  --bucket $BUCKET_NAME
```

### Test Case 3: Instance Type Change

```bash
# Stop instance
aws ec2 stop-instances --instance-ids $INSTANCE_ID
aws ec2 wait instance-stopped --instance-ids $INSTANCE_ID

# Change instance type (will trigger drift alert)
aws ec2 modify-instance-attribute \
  --instance-id $INSTANCE_ID \
  --instance-type t2.small

# Start instance
aws ec2 start-instances --instance-ids $INSTANCE_ID
```

### Test Case 4: IAM Role Policy

```bash
# Get role name
ROLE_NAME=$(terraform output -raw iam_role_name)

# Modify assume role policy (will trigger drift alert)
aws iam update-assume-role-policy \
  --role-name $ROLE_NAME \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Principal": {
        "Service": ["ec2.amazonaws.com", "lambda.amazonaws.com"]
      },
      "Action": "sts:AssumeRole"
    }]
  }'
```

## Running TFDrift-Falco

From the project root:

```bash
# Update test-config.yaml to point to the new state file
cd ../..

# Update the state path in test-config.yaml
# local_path: "./examples/terraform/terraform.tfstate"

# Run the detector
go run ./cmd/tfdrift/main.go --config test-config.yaml
```

## Cleanup

To destroy all resources:

```bash
terraform destroy
```

## Cost Estimate

- EC2 t2.micro: Free tier eligible (~$0/month with free tier, ~$8.50/month without)
- S3 Bucket: Minimal storage costs (~$0.023/GB/month)
- Data Transfer: Minimal
- **Total: ~$0-10/month**

## State Management

This example uses **local state** for simplicity. For production use, consider:

```hcl
terraform {
  backend "s3" {
    bucket = "your-terraform-state-bucket"
    key    = "tfdrift-falco/test/terraform.tfstate"
    region = "us-east-1"
  }
}
```

## Notes

- All resources are tagged with `ManagedBy = "Terraform"` for easy identification
- The EC2 instance uses the latest Amazon Linux 2023 AMI
- Security group allows SSH and HTTP from anywhere (0.0.0.0/0) - **for testing only**
- Termination protection is enabled by default to test drift detection
