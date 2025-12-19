# DeepDrift 包括的テスト環境

このTerraform構成は、DeepDriftのドリフト検知機能をテストするための包括的なAWS環境を構築します。

## アーキテクチャ

```
Internet
    ↓
Application Load Balancer (ALB)
    ↓
EC2 Web Server (Public Subnet)
    ↓
RDS PostgreSQL (Private Subnet)

Additional Services:
- Lambda Function (API)
- S3 Buckets (App Data, Logs)
- CloudWatch (Logs, Alarms)
- IAM Roles
```

## 含まれるAWSリソース

1. **ネットワーク**
   - VPC (10.100.0.0/16)
   - Public Subnet (10.100.1.0/24)
   - Private Subnet (10.100.2.0/24)
   - Internet Gateway
   - Route Tables

2. **コンピューティング**
   - EC2 Instance (t3.micro, Apache HTTP Server)
   - Lambda Function (Python 3.11)

3. **データベース**
   - RDS PostgreSQL (db.t3.micro)

4. **ストレージ**
   - S3 Bucket (Application Data)
   - S3 Bucket (Logs)

5. **ネットワーキング**
   - Application Load Balancer (ALB)
   - Target Group
   - Security Groups (Web, Database)

6. **監視**
   - CloudWatch Log Group
   - CloudWatch Alarm (CPU監視)

7. **セキュリティ**
   - IAM Role for Lambda
   - Security Groups

## デプロイ手順

```bash
cd ~/tfdrift-falco/examples/comprehensive-test

# 初期化
terraform init

# プラン確認
terraform plan

# デプロイ
terraform apply -auto-approve

# 出力確認
terraform output
```

## ドリフト検知テストシナリオ

### シナリオ1: EC2インスタンスのタグ変更

```bash
# Terraformでデプロイ後、AWSコンソールまたはCLIで手動変更
aws ec2 create-tags \
  --resources $(terraform output -raw web_server_id) \
  --tags Key=ManuallyAdded,Value=test

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**: タグの追加が検出される

### シナリオ2: Security Groupルールの変更

```bash
# SSH (22番ポート) を全公開に変更
SG_ID=$(aws ec2 describe-security-groups \
  --filters "Name=group-name,Values=deepdrift-test-web-sg" \
  --query 'SecurityGroups[0].GroupId' \
  --output text)

aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**:
- Security Groupルールの追加が検出される
- セキュリティリスク（0.0.0.0/0公開）が警告される

### シナリオ3: EC2インスタンスタイプの変更

```bash
# インスタンスを停止してタイプ変更
INSTANCE_ID=$(terraform output -raw web_server_id)

aws ec2 stop-instances --instance-ids $INSTANCE_ID
aws ec2 wait instance-stopped --instance-ids $INSTANCE_ID

aws ec2 modify-instance-attribute \
  --instance-id $INSTANCE_ID \
  --instance-type t3.small

aws ec2 start-instances --instance-ids $INSTANCE_ID

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**:
- インスタンスタイプの変更が検出される
- Severity: High（コスト・パフォーマンスに影響）

### シナリオ4: S3バケットのバージョニング無効化

```bash
# バケットのバージョニングを無効化
BUCKET_NAME=$(terraform output -raw s3_app_data_bucket)

aws s3api put-bucket-versioning \
  --bucket $BUCKET_NAME \
  --versioning-configuration Status=Suspended

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**: バージョニング設定の変更が検出される

### シナリオ5: RDSインスタンスのバックアップ設定変更

```bash
# バックアップ保持期間を変更
aws rds modify-db-instance \
  --db-instance-identifier deepdrift-test-db \
  --backup-retention-period 14 \
  --apply-immediately

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**: RDS設定の変更が検出される

### シナリオ6: Lambda環境変数の変更

```bash
# Lambda関数の環境変数を変更
aws lambda update-function-configuration \
  --function-name deepdrift-test-api \
  --environment "Variables={ENVIRONMENT=production,NEW_VAR=added}"

# DeepDriftでドリフト検知
curl http://localhost:8002/api/v1/drifts | jq
```

**期待される結果**: Lambda設定の変更が検出される

## リソースグラフの確認

```bash
# SkyGraphでスキャン
curl -X POST http://localhost:8001/api/v1/scan

# グラフデータ取得
curl http://localhost:8001/api/v1/graph | jq

# UIで確認
open http://localhost:3000/ui/
```

## クリーンアップ

```bash
# すべてのリソースを削除
terraform destroy -auto-approve
```

## 注意事項

- RDSの作成には5-10分かかります
- ALBの作成には2-3分かかります
- デフォルトでは`us-east-1`リージョンを使用します
- RDSパスワードは本番環境では必ずSecrets Managerを使用してください
- テスト後は必ずリソースを削除してコストを抑えましょう

## コスト見積もり

おおよその時間単位のコスト（us-east-1）:
- EC2 t3.micro: $0.0104/時間
- RDS db.t3.micro: $0.017/時間
- ALB: $0.0225/時間
- Lambda: 無料枠内
- S3: ほぼ無料（データが少ない場合）

**合計**: 約 $0.05/時間 = $1.20/日
