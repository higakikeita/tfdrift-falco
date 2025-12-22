# TFDrift-Falco Test Environment

実際のAWS環境でTFDrift-Falcoをテストするための簡単なTerraform構成です。

## 📋 作成されるリソース

1. **S3 Bucket** - 暗号化、バージョニング設定
2. **Security Group** - ingress/egress ルール
3. **IAM Policy** - S3とCloudWatchアクセス
4. **IAM Role** - EC2用のロール
5. **CloudWatch Log Group** - ログ保存
6. **SNS Topic** - ドリフトアラート通知用

すべてのリソースは**ドリフト検知のテストに最適**な設定になっています。

## 🚀 セットアップ手順

### 1. AWS認証情報の設定

```bash
# AWS CLIがインストール済みであることを確認
aws --version

# AWS認証情報を設定
aws configure
```

### 2. S3 Backend用バケットの作成

Terraform stateを保存するS3バケットを作成します：

```bash
# バケット名を決定（グローバルにユニークな名前）
export STATE_BUCKET="tfdrift-test-state-$(date +%Y%m%d)"
export AWS_REGION="us-east-1"

# バケットを作成
aws s3api create-bucket \
  --bucket $STATE_BUCKET \
  --region $AWS_REGION

# バケットのバージョニングを有効化（推奨）
aws s3api put-bucket-versioning \
  --bucket $STATE_BUCKET \
  --versioning-configuration Status=Enabled

# バケットの暗号化を有効化（推奨）
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

### 3. Backend設定の更新

`backend.tf` を編集して、作成したバケット名を指定：

```hcl
terraform {
  backend "s3" {
    bucket = "tfdrift-test-state-20231215"  # ← ここを更新
    key    = "test-environment/terraform.tfstate"
    region = "us-east-1"                    # ← 必要に応じて更新
  }
}
```

### 4. terraform.tfvars の作成

```bash
# サンプルファイルをコピー
cp terraform.tfvars.example terraform.tfvars

# デフォルトVPCのIDを取得
export DEFAULT_VPC=$(aws ec2 describe-vpcs \
  --filters "Name=isDefault,Values=true" \
  --query "Vpcs[0].VpcId" \
  --output text)

echo "Default VPC: $DEFAULT_VPC"

# terraform.tfvars を編集
cat > terraform.tfvars <<EOF
aws_region       = "us-east-1"
environment      = "test"
test_bucket_name = "tfdrift-test-$(date +%Y%m%d)-$(openssl rand -hex 4)"
vpc_id           = "$DEFAULT_VPC"
alert_email      = ""  # 必要に応じてメールアドレスを設定
EOF

cat terraform.tfvars
```

### 5. Terraform実行

```bash
# 初期化
terraform init

# プランの確認
terraform plan

# リソースの作成
terraform apply
```

作成されるリソースを確認し、`yes` を入力してデプロイします。

### 6. 出力値の確認

```bash
# リソース情報を表示
terraform output

# JSONフォーマットで出力
terraform output -json > resources.json
cat resources.json | jq
```

## 🔧 TFDrift設定の更新

### config.yaml の作成

```bash
cd ../../  # プロジェクトルートに戻る

# terraform output から値を取得
export STATE_BUCKET=$(cd terraform/test-environment && terraform output -raw terraform_state_location | cut -d'/' -f3)
export STATE_KEY="test-environment/terraform.tfstate"
export AWS_REGION="us-east-1"

# config.yaml を作成
cat > config.yaml <<EOF
# TFDrift-Falco Configuration

# Terraform State Configuration
terraform:
  backend: s3
  s3:
    bucket: "$STATE_BUCKET"
    key: "$STATE_KEY"
    region: "$AWS_REGION"

  # State refresh interval
  refresh_interval: 30s

# Falco Integration
falco:
  hostname: localhost
  port: 5060
  timeout: 10s

# Detection Rules
rules:
  # S3 Bucket Rules
  - name: s3_encryption_disabled
    resource_type: aws_s3_bucket_server_side_encryption_configuration
    severity: critical
    conditions:
      - attribute: rule.apply_server_side_encryption_by_default.sse_algorithm
        operator: changed

  - name: s3_versioning_disabled
    resource_type: aws_s3_bucket_versioning
    severity: high
    conditions:
      - attribute: versioning_configuration.status
        operator: changed

  - name: s3_public_access_allowed
    resource_type: aws_s3_bucket_public_access_block
    severity: critical
    conditions:
      - attribute: block_public_acls
        operator: changed
      - attribute: block_public_policy
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

# Notification Configuration
notifications:
  # Console output (always enabled)
  console:
    enabled: true
    format: json

  # Slack notifications (optional)
  # slack:
  #   enabled: true
  #   webhook_url: \${SLACK_WEBHOOK_URL}
  #   channel: "#tfdrift-alerts"
  #   username: "TFDrift Bot"

  # SNS notifications (optional)
  # sns:
  #   enabled: true
  #   topic_arn: "arn:aws:sns:us-east-1:123456789012:tfdrift-alerts"

# Logging
logging:
  level: info
  format: json
EOF

echo "✅ config.yaml created"
cat config.yaml
```

## 🧪 ドリフトテストの実行

### テストシナリオ 1: S3バケット暗号化の無効化

```bash
# 1. TFDrift を起動（別のターミナルで）
./tfdrift --config config.yaml

# 2. AWSコンソールまたはCLIでS3バケットの暗号化を無効化
export BUCKET_NAME=$(cd terraform/test-environment && terraform output -raw s3_bucket_name)

aws s3api delete-bucket-encryption --bucket $BUCKET_NAME

# 3. TFDriftのログで検知を確認
# 期待される出力:
# {
#   "level": "warn",
#   "severity": "critical",
#   "resource": "aws_s3_bucket_server_side_encryption_configuration.test",
#   "rule": "s3_encryption_disabled",
#   "message": "Drift detected: encryption configuration changed"
# }
```

### テストシナリオ 2: Security Groupルールの変更

```bash
# 1. Security Group IDを取得
export SG_ID=$(cd terraform/test-environment && terraform output -raw security_group_id)

# 2. SSHポートを全世界に開放（意図的にセキュリティリスクを作成）
aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

# 3. TFDriftのログで検知を確認
# 期待される出力:
# {
#   "level": "warn",
#   "severity": "critical",
#   "resource": "aws_security_group.test",
#   "rule": "security_group_open_to_world",
#   "message": "Drift detected: SSH open to 0.0.0.0/0"
# }
```

### テストシナリオ 3: IAM Policyの変更

```bash
# 1. IAM Policy ARNを取得
export POLICY_ARN=$(cd terraform/test-environment && terraform output -raw iam_policy_arn)

# 2. 新しいバージョンを作成（権限を拡大）
aws iam create-policy-version \
  --policy-arn $POLICY_ARN \
  --policy-document '{
    "Version": "2012-10-17",
    "Statement": [{
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }]
  }' \
  --set-as-default

# 3. TFDriftのログで検知を確認
```

### テストシナリオ 4: タグの追加/削除

```bash
# 1. S3バケットにタグを追加
aws s3api put-bucket-tagging \
  --bucket $BUCKET_NAME \
  --tagging 'TagSet=[{Key=Unauthorized,Value=true}]'

# 2. TFDriftのログで検知を確認
```

## 🔄 ドリフトの修復

### オプション1: Terraform Apply で修復

```bash
cd terraform/test-environment

# 現在の状態を確認
terraform plan

# ドリフトを修復
terraform apply -auto-approve
```

### オプション2: 手動で元に戻す

各テストシナリオの逆操作を実行：

```bash
# S3暗号化を再有効化
aws s3api put-bucket-encryption \
  --bucket $BUCKET_NAME \
  --server-side-encryption-configuration '{
    "Rules": [{
      "ApplyServerSideEncryptionByDefault": {
        "SSEAlgorithm": "AES256"
      }
    }]
  }'

# Security Groupルールを削除
aws ec2 revoke-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0
```

## 📊 UIでの確認

TFDrift Web UIでドリフトを可視化：

```bash
# UIを起動（既にDockerで起動済みの場合はスキップ）
docker-compose up -d frontend backend

# ブラウザでアクセス
open http://localhost:3000
```

UI上で以下を確認：
- ドリフトイベントの一覧
- 因果関係グラフ
- 変更前後の値の比較
- ユーザー情報とタイムスタンプ

## 🧹 クリーンアップ

テスト完了後、リソースを削除してコストを削減：

```bash
cd terraform/test-environment

# すべてのリソースを削除
terraform destroy

# State bucket も削除（必要に応じて）
aws s3 rb s3://$STATE_BUCKET --force
```

## 💰 コスト見積もり

このテスト環境の推定コスト（us-east-1）：

| リソース | 月額コスト |
|---------|-----------|
| S3 Bucket (1GB) | $0.02 |
| Security Group | $0.00 |
| IAM Policy/Role | $0.00 |
| CloudWatch Logs (1GB) | $0.50 |
| SNS (1000 emails) | $0.00 |
| **合計** | **~$0.52/月** |

短期テスト（1日）の場合：約 **$0.02**

## 🔒 セキュリティベストプラクティス

1. **最小権限の原則** - 必要最小限のIAM権限のみ付与
2. **State暗号化** - S3 backend で暗号化を有効化
3. **アクセス制限** - Security Group のCIDRを信頼できる範囲に制限
4. **定期的なクリーンアップ** - テスト後は必ずリソースを削除
5. **本番環境との分離** - 必ず別のAWSアカウントまたはリージョンでテスト

## 🐛 トラブルシューティング

### State bucket が見つからない

```bash
# bucket の存在確認
aws s3 ls | grep tfdrift

# bucket を再作成
aws s3api create-bucket --bucket YOUR_BUCKET_NAME --region us-east-1
```

### VPC が見つからない

```bash
# デフォルトVPCの確認
aws ec2 describe-vpcs --filters "Name=isDefault,Values=true"

# デフォルトVPCが無い場合は作成
aws ec2 create-default-vpc
```

### TFDrift が state を読めない

```bash
# AWS認証情報の確認
aws sts get-caller-identity

# S3バケットへのアクセス確認
aws s3 ls s3://YOUR_BUCKET_NAME/test-environment/
```

## 📚 参考リンク

- [TFDrift-Falco Documentation](https://higakikeita.github.io/tfdrift-falco/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS CloudTrail](https://aws.amazon.com/cloudtrail/)
- [Falco Documentation](https://falco.org/docs/)
