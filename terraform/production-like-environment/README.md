# TFDrift-Falco Production-Like Test Environment

本番環境に近い大規模なAWSインフラでTFDrift-Falcoをテストするための完全な構成です。

## 🏗️ アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│ VPC (10.0.0.0/16) - Multi-AZ (3 Availability Zones)        │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Public Subnets (3 AZs)                             │    │
│  │  - NAT Gateway (x3)                                │    │
│  │  - Application Load Balancer                       │    │
│  │  - Internet Gateway                                │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Private Subnets (3 AZs)                            │    │
│  │  - EKS Cluster (Managed Node Groups)               │    │
│  │  - ECS Cluster (Fargate)                           │    │
│  │  - ElastiCache Redis (Multi-AZ)                    │    │
│  └────────────────────────────────────────────────────┘    │
│                                                             │
│  ┌────────────────────────────────────────────────────┐    │
│  │ Database Subnets (3 AZs)                           │    │
│  │  - RDS PostgreSQL (Multi-AZ)                       │    │
│  └────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
```

## 📦 作成されるリソース

### ネットワーク層
- **VPC** - マルチAZ構成（3 AZs）
- **サブネット** - Public x3, Private x3, Database x3 (合計9サブネット)
- **NAT Gateway** - 各AZに1つ（高可用性）
- **Internet Gateway** - 1つ
- **Route Tables** - 複数のルートテーブル
- **VPC Flow Logs** - トラフィック監視用

### コンピューティング層
- **EKS Cluster** - Kubernetes 1.28、Managed Node Groups
- **ECS Cluster** - Fargate & Fargate Spot対応
- **Application Load Balancer** - マルチAZ対応
- **Auto Scaling** - EKS Node Groups

### データベース層
- **RDS PostgreSQL 15.4** - Multi-AZ、暗号化有効
- **ElastiCache Redis 7.0** - クラスターモード、Multi-AZ
- **暗号化** - KMS による at-rest/in-transit 暗号化

### ストレージ層
- **S3 Buckets x3**
  - Application Data (バージョニング有効)
  - Backups (ライフサイクルポリシー)
  - Logs (90日後自動削除)

### セキュリティ層
- **Security Groups x6**
  - ALB用
  - ECS Tasks用
  - EKS追加SG
  - RDS用
  - ElastiCache用
- **IAM Roles & Policies** - 最小権限の原則
- **KMS Keys x3** - EKS, RDS, ElastiCache用
- **Secrets Manager** - RDS/Redis認証情報

### 監視層
- **CloudWatch Log Groups** - ECS, Redis, VPC Flow Logs
- **CloudWatch Alarms** - ALB, RDS監視
- **Performance Insights** - RDS パフォーマンス監視

## 💰 推定コスト

| リソース | 月額コスト (us-east-1) |
|---------|---------------------|
| VPC (基本) | $0.00 |
| NAT Gateway x3 | ~$108.00 |
| EKS Cluster | $73.00 |
| EKS Nodes (t3.medium x2) | ~$60.00 |
| ECS Cluster (基本) | $0.00 |
| ALB | ~$22.00 |
| RDS (db.t3.micro, Multi-AZ) | ~$30.00 |
| ElastiCache (cache.t3.micro x2) | ~$25.00 |
| S3 (1GB x3) | ~$0.07 |
| CloudWatch Logs (5GB) | ~$2.50 |
| **合計** | **~$320/月** |

**重要:**
- これは推定コストです。実際の使用量により変動します
- テスト後は必ずリソースを削除してください
- NAT Gatewayが最もコストが高いです

## 🚀 セットアップ手順

### 前提条件

```bash
# 必要なツール
- AWS CLI (v2.x)
- Terraform (>= 1.0)
- kubectl (EKS操作用)
- jq (JSON処理用、オプション)

# AWS認証情報の確認
aws sts get-caller-identity
```

### ステップ1: S3 Backend用バケットの作成

```bash
# 環境変数を設定
export STATE_BUCKET="tfdrift-prod-state-$(date +%Y%m%d)"
export AWS_REGION="us-east-1"
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)

# バケットを作成
aws s3api create-bucket \
  --bucket $STATE_BUCKET \
  --region $AWS_REGION

# バージョニングを有効化
aws s3api put-bucket-versioning \
  --bucket $STATE_BUCKET \
  --versioning-configuration Status=Enabled

# 暗号化を有効化
aws s3api put-bucket-encryption \
  --bucket $STATE_BUCKET \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

echo "✅ State bucket created: $STATE_BUCKET"
```

### ステップ2: DynamoDB State Lock Table作成（オプション）

```bash
# State lockingのためのDynamoDBテーブルを作成
aws dynamodb create-table \
  --table-name terraform-state-lock \
  --attribute-definitions \
    AttributeName=LockID,AttributeType=S \
  --key-schema \
    AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region $AWS_REGION

echo "✅ DynamoDB lock table created"
```

### ステップ3: Backend設定の更新

```bash
# backend.tf を編集
cat > backend.tf <<EOF
terraform {
  backend "s3" {
    bucket         = "$STATE_BUCKET"
    key            = "production-like-environment/terraform.tfstate"
    region         = "$AWS_REGION"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}
EOF
```

### ステップ4: 変数ファイルの作成

```bash
# terraform.tfvars を作成
cat > terraform.tfvars <<EOF
aws_region  = "$AWS_REGION"
environment = "prod-test"
owner       = "$(whoami)"

# Network
vpc_cidr = "10.0.0.0/16"
az_count = 3

# EKS
eks_cluster_version     = "1.28"
eks_node_instance_types = ["t3.medium"]
eks_node_desired_size   = 2
eks_node_min_size       = 1
eks_node_max_size       = 4

# RDS
rds_instance_class    = "db.t3.micro"
rds_allocated_storage = 20

# ElastiCache
elasticache_node_type = "cache.t3.micro"
elasticache_num_nodes = 2
EOF

echo "✅ terraform.tfvars created"
cat terraform.tfvars
```

### ステップ5: Terraformの実行

```bash
# 初期化
terraform init

# プランの確認
terraform plan -out=tfplan

# 作成時間: 約20-30分
terraform apply tfplan

# 完了後、出力を確認
terraform output
```

### ステップ6: EKS kubeconfig設定

```bash
# EKSクラスターに接続
export CLUSTER_NAME=$(terraform output -raw eks_cluster_name)

aws eks update-kubeconfig \
  --name $CLUSTER_NAME \
  --region $AWS_REGION

# クラスター確認
kubectl get nodes
kubectl get pods --all-namespaces
```

### ステップ7: TFDrift設定の作成

```bash
cd ../../  # プロジェクトルートに戻る

# terraform outputから値を取得
cd terraform/production-like-environment
export STATE_BUCKET=$(cd terraform/production-like-environment && \
  grep 'bucket' backend.tf | awk '{print $3}' | tr -d '"')
export STATE_KEY="production-like-environment/terraform.tfstate"

# 包括的なconfig.yamlを作成
cat > ../../config-production.yaml <<'EOF'
# TFDrift-Falco Production Test Configuration

# Terraform State
terraform:
  backend: s3
  s3:
    bucket: "$STATE_BUCKET"
    key: "$STATE_KEY"
    region: "$AWS_REGION"
  refresh_interval: 30s

# Falco Integration
falco:
  hostname: localhost
  port: 5060
  timeout: 10s

# Detection Rules - Production-Like
rules:
  # VPC Rules
  - name: vpc_flow_logs_disabled
    resource_type: aws_vpc
    severity: high
    conditions:
      - attribute: enable_flow_log
        operator: changed

  # NAT Gateway Rules
  - name: nat_gateway_deleted
    resource_type: aws_nat_gateway
    severity: critical
    conditions:
      - attribute: state
        operator: not_equals
        value: "available"

  # EKS Rules
  - name: eks_public_endpoint_exposed
    resource_type: aws_eks_cluster
    severity: critical
    conditions:
      - attribute: vpc_config.endpoint_public_access
        operator: equals
        value: true

  - name: eks_secrets_encryption_disabled
    resource_type: aws_eks_cluster
    severity: critical
    conditions:
      - attribute: encryption_config
        operator: changed

  - name: eks_version_downgraded
    resource_type: aws_eks_cluster
    severity: high
    conditions:
      - attribute: version
        operator: changed

  # ECS Rules
  - name: ecs_container_insights_disabled
    resource_type: aws_ecs_cluster
    severity: medium
    conditions:
      - attribute: setting.container_insights
        operator: equals
        value: "disabled"

  # ALB Rules
  - name: alb_access_logs_disabled
    resource_type: aws_lb
    severity: medium
    conditions:
      - attribute: access_logs.enabled
        operator: equals
        value: false

  - name: alb_deletion_protection_disabled
    resource_type: aws_lb
    severity: high
    conditions:
      - attribute: enable_deletion_protection
        operator: equals
        value: false

  # S3 Rules
  - name: s3_encryption_disabled
    resource_type: aws_s3_bucket_server_side_encryption_configuration
    severity: critical
    conditions:
      - attribute: rule.apply_server_side_encryption_by_default
        operator: changed

  - name: s3_versioning_disabled
    resource_type: aws_s3_bucket_versioning
    severity: high
    conditions:
      - attribute: versioning_configuration.status
        operator: not_equals
        value: "Enabled"

  - name: s3_public_access_allowed
    resource_type: aws_s3_bucket_public_access_block
    severity: critical
    conditions:
      - attribute: block_public_acls
        operator: equals
        value: false

  - name: s3_lifecycle_policy_changed
    resource_type: aws_s3_bucket_lifecycle_configuration
    severity: medium
    conditions:
      - attribute: rule
        operator: changed

  # Security Group Rules
  - name: security_group_ingress_changed
    resource_type: aws_security_group
    severity: critical
    conditions:
      - attribute: ingress
        operator: changed

  - name: security_group_open_to_world
    resource_type: aws_security_group
    severity: critical
    conditions:
      - attribute: ingress.cidr_blocks
        operator: contains
        value: "0.0.0.0/0"

  - name: security_group_all_ports_open
    resource_type: aws_security_group
    severity: critical
    conditions:
      - attribute: ingress.from_port
        operator: equals
        value: 0
      - attribute: ingress.to_port
        operator: equals
        value: 65535

  # RDS Rules
  - name: rds_multi_az_disabled
    resource_type: aws_db_instance
    severity: critical
    conditions:
      - attribute: multi_az
        operator: equals
        value: false

  - name: rds_encryption_disabled
    resource_type: aws_db_instance
    severity: critical
    conditions:
      - attribute: storage_encrypted
        operator: equals
        value: false

  - name: rds_backup_retention_reduced
    resource_type: aws_db_instance
    severity: high
    conditions:
      - attribute: backup_retention_period
        operator: less_than
        value: 7

  - name: rds_deletion_protection_disabled
    resource_type: aws_db_instance
    severity: critical
    conditions:
      - attribute: deletion_protection
        operator: equals
        value: false

  - name: rds_public_access_enabled
    resource_type: aws_db_instance
    severity: critical
    conditions:
      - attribute: publicly_accessible
        operator: equals
        value: true

  # ElastiCache Rules
  - name: elasticache_encryption_disabled
    resource_type: aws_elasticache_replication_group
    severity: critical
    conditions:
      - attribute: at_rest_encryption_enabled
        operator: equals
        value: false
      - attribute: transit_encryption_enabled
        operator: equals
        value: false

  - name: elasticache_automatic_failover_disabled
    resource_type: aws_elasticache_replication_group
    severity: high
    conditions:
      - attribute: automatic_failover_enabled
        operator: equals
        value: false

  - name: elasticache_multi_az_disabled
    resource_type: aws_elasticache_replication_group
    severity: high
    conditions:
      - attribute: multi_az_enabled
        operator: equals
        value: false

  # IAM Rules
  - name: iam_policy_modified
    resource_type: aws_iam_policy
    severity: high
    conditions:
      - attribute: policy
        operator: changed

  - name: iam_role_assume_policy_changed
    resource_type: aws_iam_role
    severity: high
    conditions:
      - attribute: assume_role_policy
        operator: changed

  - name: iam_role_policy_attachment_changed
    resource_type: aws_iam_role_policy_attachment
    severity: medium
    conditions:
      - attribute: policy_arn
        operator: changed

  # KMS Rules
  - name: kms_key_rotation_disabled
    resource_type: aws_kms_key
    severity: high
    conditions:
      - attribute: enable_key_rotation
        operator: equals
        value: false

  - name: kms_key_deletion_window_reduced
    resource_type: aws_kms_key
    severity: medium
    conditions:
      - attribute: deletion_window_in_days
        operator: less_than
        value: 7

  # CloudWatch Rules
  - name: cloudwatch_log_retention_reduced
    resource_type: aws_cloudwatch_log_group
    severity: medium
    conditions:
      - attribute: retention_in_days
        operator: changed

  - name: cloudwatch_alarm_disabled
    resource_type: aws_cloudwatch_metric_alarm
    severity: medium
    conditions:
      - attribute: actions_enabled
        operator: equals
        value: false

# Notifications
notifications:
  console:
    enabled: true
    format: json

  # Slack通知（オプション）
  # slack:
  #   enabled: true
  #   webhook_url: ${SLACK_WEBHOOK_URL}
  #   channel: "#tfdrift-prod-alerts"

# Logging
logging:
  level: info
  format: json
  output: stdout
EOF

# 環境変数を展開
envsubst < ../../config-production.yaml > ../../config-production-final.yaml
mv ../../config-production-final.yaml ../../config-production.yaml

echo "✅ config-production.yaml created"
```

## 🧪 ドリフトテストシナリオ

### シナリオ1: NAT Gatewayの削除（Critical）

```bash
export NAT_GW_ID=$(terraform output -json nat_gateway_ids | jq -r '.[0]')

# NAT Gatewayを削除（意図的）
aws ec2 delete-nat-gateway --nat-gateway-id $NAT_GW_ID

# TFDriftで検知確認
# 期待される出力: Critical alert - nat_gateway_deleted
```

### シナリオ2: EKS Public Endpoint の公開（Critical）

```bash
export CLUSTER_NAME=$(terraform output -raw eks_cluster_name)

# Public endpoint accessを有効化
aws eks update-cluster-config \
  --name $CLUSTER_NAME \
  --resources-vpc-config endpointPublicAccess=true,endpointPrivateAccess=true

# TFDriftで検知確認
```

### シナリオ3: RDS Multi-AZの無効化（Critical）

```bash
export RDS_ID=$(terraform output -json rds_endpoint | jq -r 'split(":")[0]')

# Single-AZに変更（ダウンタイムあり）
aws rds modify-db-instance \
  --db-instance-identifier $RDS_ID \
  --no-multi-az \
  --apply-immediately

# TFDriftで検知確認
```

### シナリオ4: S3バケット暗号化の無効化（Critical）

```bash
export BUCKET_NAME=$(terraform output -raw app_data_bucket_name)

# 暗号化を削除
aws s3api delete-bucket-encryption --bucket $BUCKET_NAME

# TFDriftで検知確認
```

### シナリオ5: Security Group ルールの変更（Critical）

```bash
export ALB_SG_ID=$(terraform output -raw alb_security_group_id)

# 全ポートを開放（危険）
aws ec2 authorize-security-group-ingress \
  --group-id $ALB_SG_ID \
  --protocol tcp \
  --port 0-65535 \
  --cidr 0.0.0.0/0

# TFDriftで検知確認
```

### シナリオ6: ElastiCache 暗号化の無効化（Critical）

```bash
# 注: ElastiCacheの暗号化は作成後変更不可
# 代わりに、クラスターを削除して再作成

export REPLICATION_GROUP_ID=$(cd terraform/production-like-environment && \
  terraform show -json | jq -r '.values.root_module.resources[] | select(.type=="aws_elasticache_replication_group") | .values.replication_group_id')

echo "Replication Group: $REPLICATION_GROUP_ID"

# TFDriftで検知確認
```

### シナリオ7: IAM Policyの権限拡大（High）

```bash
export POLICY_ARN=$(terraform output -json ecs_task_role_arn | jq -r)

# 新しいポリシーをアタッチ（意図的な権限エスカレーション）
aws iam attach-role-policy \
  --role-name $(echo $POLICY_ARN | rev | cut -d'/' -f1 | rev) \
  --policy-arn arn:aws:iam::aws:policy/AdministratorAccess

# TFDriftで検知確認
```

## 🔄 ドリフトの修復

### オプション1: Terraform Applyで修復

```bash
cd terraform/production-like-environment

# 現在の状態を確認
terraform plan

# ドリフトを修復
terraform apply -auto-approve
```

### オプション2: AWS CLIで手動修復

各シナリオの逆操作を実行（例: NAT Gatewayの再作成）

## 📊 UIでの確認

```bash
# UIが起動していない場合
cd ../../
docker-compose up -d frontend backend

# ブラウザでアクセス
open http://localhost:3000
```

## 🧹 クリーンアップ

**重要:** テスト完了後は必ずリソースを削除してください

```bash
cd terraform/production-like-environment

# すべてのリソースを削除（20-30分）
terraform destroy

# State bucketも削除
aws s3 rb s3://$STATE_BUCKET --force

# DynamoDB lock table削除
aws dynamodb delete-table \
  --table-name terraform-state-lock \
  --region $AWS_REGION
```

## 📈 モニタリング

### CloudWatch Dashboard

```bash
# ダッシュボードURLを取得
echo "https://console.aws.amazon.com/cloudwatch/home?region=$AWS_REGION#dashboards:"
```

### コスト確認

```bash
# 現在のコストを確認
aws ce get-cost-and-usage \
  --time-period Start=$(date -d '1 day ago' +%Y-%m-%d),End=$(date +%Y-%m-%d) \
  --granularity DAILY \
  --metrics BlendedCost \
  --group-by Type=TAG,Key=Project
```

## 🔒 セキュリティのベストプラクティス

1. **最小権限の原則** - 必要最小限のIAM権限のみ付与
2. **暗号化** - すべてのデータを at-rest/in-transit で暗号化
3. **Multi-AZ** - 高可用性のため複数AZ構成
4. **Secrets管理** - Secrets Managerで認証情報を管理
5. **VPC分離** - Public/Private/Database サブネット分離
6. **State暗号化** - Terraform state を暗号化
7. **定期的なクリーンアップ** - 使わないリソースは即削除

## 🐛 トラブルシューティング

### EKS Node Groupが起動しない

```bash
# ノードグループの状態確認
aws eks describe-nodegroup \
  --cluster-name $CLUSTER_NAME \
  --nodegroup-name prod-test-tfdrift-general

# IAM role確認
aws iam get-role --role-name <node-role-name>
```

### RDS接続エラー

```bash
# Security Groupルール確認
aws ec2 describe-security-groups \
  --group-ids $(terraform output -raw rds_security_group_id)

# RDSエンドポイント確認
aws rds describe-db-instances \
  --db-instance-identifier $(terraform output -raw rds_endpoint | cut -d':' -f1)
```

### ElastiCache接続エラー

```bash
# Redis cluster確認
aws elasticache describe-replication-groups \
  --replication-group-id $(terraform show -json | jq -r '.values.root_module.resources[] | select(.type=="aws_elasticache_replication_group") | .values.replication_group_id')
```

## 📚 参考リンク

- [TFDrift-Falco Documentation](https://higakikeita.github.io/tfdrift-falco/)
- [AWS EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
- [AWS ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/bestpracticesguide/intro.html)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
