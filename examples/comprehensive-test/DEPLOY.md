# 🚀 包括的テスト環境デプロイガイド

## 含まれるAWSサービス（全23リソース）

### ネットワーキング (9リソース)
- ✅ VPC (10.100.0.0/16)
- ✅ Public Subnets x2 (AZ-a, AZ-b)
- ✅ Private Subnets x2 (AZ-a, AZ-b)
- ✅ Internet Gateway
- ✅ **NAT Gateway** (Elastic IP付き)
- ✅ Route Tables x2 (Public, Private)

### コンピューティング (5リソース)
- ✅ EC2 Web Server (t3.micro)
- ✅ **EKS Cluster** (v1.28)
- ✅ **EKS Node Group** (t3.medium x2)
- ✅ Lambda Function (Python 3.11)

### データベース (1リソース)
- ✅ RDS PostgreSQL (db.t3.micro)

### ストレージ (2リソース)
- ✅ S3 Bucket (Application Data)
- ✅ S3 Bucket (Logs)

### ロードバランシング (2リソース)
- ✅ Application Load Balancer (ALB)
- ✅ Target Group

### セキュリティ (6リソース)
- ✅ Security Groups x5 (Web, DB, EKS Cluster, EKS Nodes, ALB)
- ✅ **WAF Web ACL** (6ルール)
  - Rate Limiting (2000 req/5min)
  - AWS Managed Core Rule Set
  - Known Bad Inputs Protection
  - SQL Injection Protection
  - Geographic Blocking
  - IP Blacklist
- ✅ IAM Roles x4 (Lambda, EKS Cluster, EKS Nodes, App)

### 監視 (3リソース)
- ✅ CloudWatch Log Groups x2 (App, WAF)
- ✅ CloudWatch Alarms x2 (High CPU, WAF Blocked Requests)

## 📊 アーキテクチャ図

```
Internet
    ↓
[WAF] → [ALB] → [Public Subnet 1a]
              ↓     ├─ EC2 Web Server
              ↓     └─ [NAT Gateway]
              ↓            ↓
              ↓     [Private Subnet 1a]
              ↓     ├─ RDS PostgreSQL
              ↓     ├─ EKS Nodes (x2)
              ↓     └─ Lambda Function
              ↓
              └──→ [Public Subnet 1b]
                        ↓
                   [Private Subnet 1b]
                        └─ EKS Nodes (backup)
```

## 💰 コスト見積もり

| サービス | タイプ | 時間単価 | 日単価 |
|---------|-------|---------|--------|
| EC2 | t3.micro | $0.0104 | $0.25 |
| EKS Cluster | 固定 | $0.10 | $2.40 |
| EKS Nodes | t3.medium x2 | $0.0832 | $2.00 |
| RDS | db.t3.micro | $0.017 | $0.41 |
| ALB | 固定 | $0.0225 | $0.54 |
| NAT Gateway | 固定 | $0.045 | $1.08 |
| WAF | 固定 | $0.60 | $14.40 |
| **合計** | - | **$0.88/時間** | **$21.08/日** |

⚠️ **重要**: EKSとWAFが高コストなので、テスト後は必ず削除してください！

## 🔧 デプロイ手順

### 1. 前提条件

```bash
# AWS認証確認
aws sts get-caller-identity --profile draios-dev-developer

# Terraform バージョン確認 (>= 1.0)
terraform version

# 必要な権限
# - EKS操作権限
# - WAF操作権限
# - その他フル管理者権限
```

### 2. デプロイ実行

```bash
cd ~/tfdrift-falco/examples/comprehensive-test

# 初期化
terraform init

# プラン確認（約30リソース作成）
terraform plan

# デプロイ（約15-20分かかります）
terraform apply -auto-approve

# 主要な待ち時間:
# - RDS作成: 5-8分
# - EKS Cluster作成: 10-12分
# - EKS Node Group作成: 3-5分
# - その他: 1-2分
```

### 3. デプロイ状況の確認

```bash
# すべての出力を表示
terraform output

# EKS クラスタ設定
aws eks update-kubeconfig \
  --name $(terraform output -raw eks_cluster_name) \
  --profile draios-dev-developer

# EKS ノード確認
kubectl get nodes

# ALB DNS確認
echo "ALB URL: http://$(terraform output -raw alb_dns_name)"

# WAF確認
echo "WAF ACL: $(terraform output -raw waf_web_acl_id)"
```

## 🧪 ドリフト検知テストシナリオ

### シナリオ1: EKS Node Group のスケーリング変更

```bash
# Desired Sizeを変更 (2 → 3)
aws eks update-nodegroup-config \
  --cluster-name $(terraform output -raw eks_cluster_name) \
  --nodegroup-name deepdrift-test-node-group \
  --scaling-config desiredSize=3

# ドリフト検知
curl http://localhost:8002/api/v1/drifts | jq '.drifts[] | select(.resource_type == "eks_node_group")'
```

### シナリオ2: WAF ルールの無効化

```bash
# Rate Limitルールを無効化
WAF_ID=$(terraform output -raw waf_web_acl_id)

# AWS Console または CLI でルールを無効化
# Expected: WAF設定の変更を検出
```

### シナリオ3: NAT Gateway の削除（危険）

```bash
# NAT Gatewayを手動削除
NAT_ID=$(terraform output -raw nat_gateway_id)
aws ec2 delete-nat-gateway --nat-gateway-id $NAT_ID

# ドリフト検知
# Expected: Severity=Critical, 重大なインフラ変更を検出
```

### シナリオ4: セキュリティグループの変更

```bash
# EKS Cluster SGに不要なルールを追加
SG_ID=$(terraform output -raw eks_cluster_security_group_id)

aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 22 \
  --cidr 0.0.0.0/0

# ドリフト検知
# Expected: セキュリティリスク（SSH全公開）を警告
```

### シナリオ5: EKS クラスタのログ無効化

```bash
# Cluster loggingを無効化
aws eks update-cluster-config \
  --name $(terraform output -raw eks_cluster_name) \
  --logging '{"clusterLogging":[{"types":["api","audit","authenticator"],"enabled":false}]}'

# ドリフト検知
# Expected: ログ設定の変更を検出
```

## 📈 リソースグラフの確認

### SkyGraphでスキャン

```bash
# AWSリソースをスキャン
curl -X POST http://localhost:8001/api/v1/scan

# 結果確認（約30リソース）
curl http://localhost:8001/api/v1/graph | jq '{
  node_count: (.nodes | length),
  resource_types: (.nodes | group_by(.type) | map({type: .[0].type, count: length}))
}'
```

### DeepDriftで構成図比較

```bash
# Terraform意図構成
curl http://localhost:8002/api/v1/graph/intended | jq '{node_count: (.nodes | length)}'

# 実際のAWS構成
curl http://localhost:8002/api/v1/graph | jq '{node_count: (.nodes | length)}'

# UIで視覚的に確認
open http://localhost:3000/ui/
```

## 🧹 クリーンアップ

```bash
# 手動で変更したリソースを戻す（必要に応じて）
terraform refresh

# すべて削除（約10-15分）
terraform destroy -auto-approve

# 確認
terraform show
```

## ⚠️ 注意事項

1. **EKS削除時の注意**
   - Node Groupが完全に削除されるまで待つ
   - ENI（Elastic Network Interface）が残る場合がある
   - VPC削除前にENIを手動削除する必要がある場合がある

2. **NAT Gateway削除**
   - Elastic IPの削除を忘れずに
   - 削除には数分かかる

3. **WAF削除**
   - ALBとの関連付けを先に解除
   - ログ設定も削除される

4. **コスト管理**
   - EKSとWAFは高コスト
   - テスト後は必ず`terraform destroy`を実行
   - CloudWatchアラームで予算超過を監視

## 🎯 期待される結果

デプロイ後:
- ✅ 約30個のAWSリソースが作成される
- ✅ EKSクラスタが稼働し、Nodeが2つ起動
- ✅ WAFがALBを保護
- ✅ NAT Gatewayを通じてPrivate SubnetがInternet接続
- ✅ すべてのリソースがSkyGraphでスキャンされる
- ✅ Terraform意図構成と実構成が一致する

ドリフト作成後:
- ✅ DeepDriftがすべての変更を検出
- ✅ セキュリティリスクが赤色で強調
- ✅ UIで意図vs実際の差分が可視化される
