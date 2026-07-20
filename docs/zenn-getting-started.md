---
title: "TFDrift-Falco で実現するリアルタイム Terraform Drift 検知"
emoji: "🛰️"
type: "tech"
topics: ["terraform", "aws", "falco", "iac", "devops"]
published: false
---

# TFDrift-Falco とは？

「AWS Console で誰かが設定を変更したけど、Terraform State に反映されていない...」

こんな **Terraform Drift（設定のズレ）** を**リアルタイムで検知**して、即座に Slack 通知してくれる OSS が **TFDrift-Falco** です。

https://github.com/higakikeita/tfdrift-falco

## どんな仕組み？

```
誰かが AWS Console で EC2 の設定を変更
    ↓
CloudTrail イベントを Falco が検知（数秒以内）
    ↓
TFDrift-Falco が Terraform State と比較
    ↓
差分があれば Slack に即座に通知 🚨
```

## 主な特徴

- ⚡ **リアルタイム検知** - CloudTrail イベントをストリーム処理
- 🔍 **差分の詳細表示** - 期待値 vs 実際の値を比較
- 👤 **ユーザー識別** - 誰が変更したか特定可能
- 🔔 **複数の通知チャネル** - Slack、Discord、Webhook
- 📊 **Grafana ダッシュボード** - 可視化とアラート
- 🤖 **Auto-Import 機能** - 管理外リソースの自動取り込み

## 従来ツールとの違い

| 機能 | TFDrift-Falco | terraform plan | driftctl |
|------|--------------|----------------|----------|
| 検知方法 | **リアルタイム** | 手動実行 | 定期スキャン |
| レイテンシ | **数秒** | 手動 | 数分〜数時間 |
| ユーザー識別 | **○** | × | × |
| 通知 | **○** | × | 一部対応 |

---

# セットアップ手順

所要時間: **約40分**

## 前提条件

- Docker Desktop がインストール済み
- AWS CLI が設定済み（`aws configure` 完了）
- Terraform 1.0+ がインストール済み
- Terraform State が存在する（local または S3）

## Phase 1: Falco のセットアップ（15分）

TFDrift-Falco は Falco と連携して動作します。まず Falco をセットアップします。

### Step 1: Falco 設定ファイルを作成

```bash
# ディレクトリ作成
mkdir -p ~/tfdrift-setup/falco
cd ~/tfdrift-setup/falco

# Falco 設定ファイル作成
cat > falco.yaml << 'EOF'
json_output: true
json_include_output_property: true

# gRPC output enabled
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"

# CloudTrail plugin
plugins:
  - name: cloudtrail
    library_path: libcloudtrail.so
    init_config:
      s3DownloadConcurrency: 10

load_plugins: [cloudtrail]

rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
EOF
```

### Step 2: Falco を起動

```bash
docker run -d \
  --name falco \
  --restart unless-stopped \
  -p 5060:5060 \
  -v $(pwd)/falco.yaml:/etc/falco/falco.yaml \
  -v ~/.aws:/root/.aws:ro \
  -e AWS_REGION=us-east-1 \
  falcosecurity/falco:0.37.1 \
  --disable-source syscall
```

### Step 3: 動作確認

```bash
# ログ確認
docker logs falco

# gRPC が起動しているか確認
curl -v http://localhost:5060
# → "method not allowed" が返ればOK ✅
```

:::message
**トラブルシューティング**

- `AWS credentials not found` → `~/.aws/credentials` を確認
- `port already in use` → ポート 5060 を使用している他のプロセスを停止
:::

## Phase 2: TFDrift-Falco のセットアップ（10分）

### Step 1: プロジェクトをクローン

```bash
cd ~/tfdrift-setup
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

### Step 2: 設定ファイルを作成

```bash
cp config.example.yaml config.yaml
vim config.yaml
```

**config.yaml（最小構成）**:

```yaml
# Falco 連携設定
falco:
  enabled: true
  hostname: localhost
  port: 5060
  tls: false

# AWS 設定
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: local
      local_path: /path/to/your/terraform.tfstate  # ← 変更必須

# ドリフトルール
drift_rules:
  - name: "EC2 Configuration Drift"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
      - "tags"
      - "disable_api_termination"
    severity: "high"

  - name: "IAM Policy Drift"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
    watched_attributes:
      - "assume_role_policy"
      - "policy"
    severity: "critical"

  - name: "S3 Bucket Configuration Drift"
    resource_types:
      - "aws_s3_bucket"
    watched_attributes:
      - "acl"
      - "versioning"
      - "logging"
    severity: "high"

# 通知設定
notifications:
  slack:
    enabled: false  # 後で設定
    webhook_url: ""

  falco_output:
    enabled: true
    priority: "warning"

# ログ設定
logging:
  level: "info"
  format: "text"
```

:::message alert
**重要**: `state.local_path` を実際の Terraform State ファイルのパスに変更してください。
:::

### Step 3: Docker Compose で起動

```bash
docker-compose up -d
docker-compose logs -f tfdrift
```

**期待される出力**:

```
INFO[2025-12-05 12:00:00] Starting TFDrift-Falco v0.1.0
INFO[2025-12-05 12:00:00] Connected to Falco gRPC: localhost:5060
INFO[2025-12-05 12:00:01] Loaded Terraform state: 42 resources
INFO[2025-12-05 12:00:01] Drift detection started
```

## Phase 3: Slack 通知の設定（5分）

### Step 1: Slack Webhook を作成

1. https://api.slack.com/apps にアクセス
2. **Create New App** → **From scratch**
3. App Name: `TFDrift-Falco`、Workspace を選択
4. **Incoming Webhooks** → **Activate Incoming Webhooks** をオン
5. **Add New Webhook to Workspace**
6. 通知先チャンネル（例: `#alerts`）を選択
7. Webhook URL をコピー

### Step 2: config.yaml を更新

```yaml
notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#alerts"
```

### Step 3: 再起動

```bash
docker-compose restart tfdrift
```

## Phase 4: 動作確認（10分）

実際に AWS リソースを変更して、検知をテストします。

### Step 1: テスト用 EC2 インスタンスを作成

**terraform/main.tf**:

```hcl
terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "test" {
  ami           = "ami-0c55b159cbfafe1f0"
  instance_type = "t2.micro"

  tags = {
    Name        = "tfdrift-test"
    Environment = "development"
    ManagedBy   = "terraform"
  }

  disable_api_termination = true
}

output "instance_id" {
  value = aws_instance.test.id
}
```

```bash
cd terraform
terraform init
terraform apply

# Instance ID をメモ
terraform output instance_id
# → i-0123456789abcdef0
```

### Step 2: AWS Console で変更

1. AWS Console → EC2 → Instances
2. `tfdrift-test` インスタンスを選択
3. **Actions** → **Instance settings** → **Change termination protection**
4. **Disable** を選択 → **Save**

### Step 3: TFDrift-Falco のログを確認

```bash
docker-compose logs -f tfdrift
```

**期待される出力**:

```
INFO[2025-12-05 12:10:23] Drift detected: aws_instance.test
INFO[2025-12-05 12:10:23] Resource: i-0123456789abcdef0
INFO[2025-12-05 12:10:23] Attribute changed: disable_api_termination
INFO[2025-12-05 12:10:23]   Expected: true
INFO[2025-12-05 12:10:23]   Actual:   false
INFO[2025-12-05 12:10:23] Changed by: john.doe@company.com
INFO[2025-12-05 12:10:23] Notification sent to Slack
```

### Step 4: Slack を確認

Slack の `#alerts` チャンネルに通知が届きます：

```
🚨 Terraform Drift Detected

📦 Resource: aws_instance.test (i-0123456789abcdef0)
🔧 Attribute: disable_api_termination
📊 Severity: high

Expected: true
Actual:   false

👤 Changed By: john.doe@company.com
🕐 Detected At: 2025-12-05 12:10:23
```

:::message
**成功！** 🎉
設定変更からわずか数秒で検知・通知されました。
:::

---

# 高度な設定

## ダッシュボード UI

リアルタイムで可視化したい場合は、付属の React Dashboard UI を使います（API サーバー + フロントエンド）。

```bash
docker compose up -d
```

→ http://localhost:3000 でダッシュボードが開きます。ドリフトイベントのライブフィード、トポロジーグラフ、差分詳細を確認できます。

## Auto-Import 機能

管理外リソースを自動で Terraform に取り込む機能です。

**config.yaml**:

```yaml
auto_import:
  enabled: true
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"

  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
    - "aws_s3_bucket"

  require_approval: true  # 承認が必要
```

**動作例**:

```bash
🔔 IMPORT APPROVAL REQUIRED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Resource Type: aws_s3_bucket
🆔 Resource ID:   my-unmanaged-bucket
👤 Detected By:   admin@company.com

💻 Import Command:
   terraform import aws_s3_bucket.my_unmanaged_bucket my-unmanaged-bucket

❓ Approve this import? [y/N]: y

✅ Import successful!
📄 Generated code: ./infrastructure/imported/aws_s3_bucket_my_unmanaged_bucket.tf
```

詳細: [Auto-Import ガイド](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/auto-import-guide.md)

## S3 Backend の使用

Terraform State が S3 にある場合：

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: "my-terraform-state"
      s3_key: "production/terraform.tfstate"
      s3_region: "us-east-1"
```

## 複数リージョンの監視

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
      - ap-northeast-1
```

---

# 本番環境での運用

## 推奨アーキテクチャ

```
┌─────────────────────────────────────────┐
│         AWS Account (Production)        │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │         ECS Cluster                │ │
│  │                                    │ │
│  │  ┌──────────┐  ┌───────────────┐ │ │
│  │  │  Falco   │  │ TFDrift-Falco │ │ │
│  │  │  Task    │→ │     Task      │ │ │
│  │  └──────────┘  └───────────────┘ │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
                    ↓
           ┌────────────────┐
           │  Slack/Email   │
           └────────────────┘
```

## ECS Task Definition（抜粋）

```json
{
  "family": "tfdrift-falco",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "falco",
      "image": "falcosecurity/falco:0.37.1",
      "portMappings": [
        {
          "containerPort": 5060,
          "protocol": "tcp"
        }
      ]
    },
    {
      "name": "tfdrift",
      "image": "ghcr.io/higakikeita/tfdrift-falco:latest",
      "dependsOn": [
        {
          "containerName": "falco",
          "condition": "START"
        }
      ],
      "secrets": [
        {
          "name": "TFDRIFT_SLACK_WEBHOOK_URL",
          "valueFrom": "arn:aws:secretsmanager:..."
        }
      ]
    }
  ]
}
```

## セキュリティのベストプラクティス

### 1. IAM Role の最小権限

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "cloudtrail:LookupEvents",
        "s3:GetObject"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::my-terraform-state/*"
    }
  ]
}
```

### 2. Secrets Manager で認証情報管理

```bash
# Slack Webhook を Secrets Manager に保存
aws secretsmanager create-secret \
  --name tfdrift/slack-webhook \
  --secret-string "https://hooks.slack.com/services/..."

# ECS Task で参照
"secrets": [
  {
    "name": "TFDRIFT_SLACK_WEBHOOK_URL",
    "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789012:secret:tfdrift/slack-webhook"
  }
]
```

### 3. ネットワーク分離

- VPC 内で実行
- Security Group で 5060 ポートのアクセス制限
- Private Subnet での実行を推奨

---

# トラブルシューティング

## Falco に接続できない

**エラー**: `Failed to connect to Falco gRPC: connection refused`

**対策**:

```bash
# Falco が起動しているか確認
docker ps | grep falco

# Falco のログを確認
docker logs falco | grep gRPC

# ポートが開いているか確認
netstat -an | grep 5060
```

## Terraform State が読み込めない

**エラー**: `Failed to load Terraform state: file not found`

**対策**:

```bash
# パスを確認
ls -la /path/to/terraform.tfstate

# 絶対パスで指定
state:
  backend: local
  local_path: /absolute/path/to/terraform.tfstate
```

## ドリフトが検知されない

**チェックリスト**:

- [ ] Falco が CloudTrail イベントを受信している？
  ```bash
  docker logs falco | grep cloudtrail
  ```

- [ ] リソースタイプがルールに含まれている？
  ```yaml
  drift_rules:
    - name: "Test"
      resource_types:
        - "aws_instance"  # ← これが含まれているか
  ```

- [ ] watched_attributes が正しい？
  ```yaml
  watched_attributes:
    - "disable_api_termination"  # ← 属性名が正しいか
  ```

- [ ] Terraform State に該当リソースが存在する？
  ```bash
  terraform state list | grep aws_instance.test
  ```

---

# 実用例

## ユースケース 1: セキュリティ監視

**シナリオ**: IAM ロールや S3 バケットの設定変更を即座に検知

```yaml
drift_rules:
  - name: "Critical Security Configuration"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
      - "aws_s3_bucket"
      - "aws_security_group"
    watched_attributes:
      - "assume_role_policy"
      - "policy"
      - "acl"
      - "ingress"
      - "egress"
    severity: "critical"

notifications:
  slack:
    channel: "#security-alerts"
```

**効果**:
- 不正なアクセス権限変更を数秒で検知
- 誰が変更したか特定可能
- インシデント対応時間を大幅短縮

## ユースケース 2: 本番環境の変更管理

**シナリオ**: 本番環境への手動変更を禁止し、IaC 経由のみを許可

```yaml
drift_rules:
  - name: "Production Environment Protection"
    resource_types:
      - "aws_instance"
      - "aws_rds_instance"
      - "aws_elasticache_cluster"
      - "aws_lambda_function"
    watched_attributes:
      - "*"  # すべての属性を監視
    severity: "critical"

auto_import:
  enabled: true
  require_approval: true
```

**効果**:
- 手動変更を検知して即座に通知
- 変更内容を自動で Terraform コード化
- 承認プロセスを経て State に反映

## ユースケース 3: マルチアカウント監視

**構成**:

```
Account A (Production)
  → TFDrift-Falco Instance A → Slack #prod-alerts

Account B (Staging)
  → TFDrift-Falco Instance B → Slack #staging-alerts

Account C (Development)
  → TFDrift-Falco Instance C → Slack #dev-alerts
```

各アカウントで独立して実行し、それぞれ異なる Slack チャンネルに通知。

---

# パフォーマンスとコスト

## リソース使用量

| コンポーネント | CPU | メモリ | ディスク |
|---------------|-----|--------|----------|
| Falco | 1-5% | 150MB | 100MB |
| TFDrift-Falco | 1-3% | 100MB | 50MB |
| **合計** | **<10%** | **250MB** | **150MB** |

t3.small インスタンス（$0.0208/時間）で十分動作します。

## 月間コスト試算

```
ECS Fargate (0.5 vCPU, 1GB メモリ):
  $0.04856 × 24時間 × 30日 = $35/月

t3.small EC2 (2 vCPU, 2GB メモリ):
  $0.0208 × 24時間 × 30日 = $15/月

CloudTrail:
  無料枠（最初の 100,000 イベント）
  追加イベント: $2.00/100,000イベント
```

**合計**: 月額 $15-50 程度で運用可能

---

# FAQ

## Terraform Cloud に対応していますか？

はい。Terraform Cloud の Remote State に対応しています。

```yaml
providers:
  aws:
    state:
      backend: remote
      organization: "my-org"
      workspace: "production"
```

API Token は環境変数 `TF_CLOUD_TOKEN` で設定してください。

## 既存の terraform plan との違いは？

| 比較項目 | terraform plan | TFDrift-Falco |
|---------|---------------|--------------|
| 実行タイミング | 手動 | **リアルタイム** |
| 検知速度 | 数分〜数時間 | **数秒** |
| ユーザー特定 | × | **○** |
| 自動通知 | × | **○** |

TFDrift-Falco は `terraform plan` を置き換えるものではなく、**補完する**ツールです。

## CloudTrail の費用が心配です

CloudTrail は最初の 100,000 イベント/月が無料です。通常の利用であれば追加費用はほとんど発生しません。

## GCP や Azure に対応していますか？

現在は AWS のみ対応。GCP、Azure は Phase 2 で対応予定です（2025年 Q2 予定）。

## 検知の遅延はどのくらいですか？

CloudTrail イベント発生から通知まで、通常 **3-10 秒**です。

---

# まとめ

TFDrift-Falco を使えば：

✅ **リアルタイムで Drift を検知** - 手動変更を見逃さない
✅ **誰が変更したか特定** - インシデント対応が迅速化
✅ **自動通知で即座に対応** - Slack で関係者に通知
✅ **Grafana で可視化** - トレンド分析とダッシュボード
✅ **Auto-Import で自動化** - 管理外リソースを自動取り込み

特に、**セキュリティ重視の環境**や**変更管理が厳格な本番環境**で威力を発揮します。

## 次のステップ

1. [GitHub リポジトリ](https://github.com/higakikeita/tfdrift-falco) を Star ⭐
2. サンプル環境で試してみる
3. Slack 通知を設定
4. Grafana ダッシュボードを追加
5. 本番環境へのデプロイ

## リンク

- **GitHub**: https://github.com/higakikeita/tfdrift-falco
- **Grafana セットアップ**: [dashboards/grafana/GETTING_STARTED.md](https://github.com/higakikeita/tfdrift-falco/blob/main/dashboards/grafana/GETTING_STARTED.md)
- **Auto-Import ガイド**: [docs/auto-import-guide.md](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/auto-import-guide.md)
- **Issue / 質問**: https://github.com/higakikeita/tfdrift-falco/issues

フィードバックや機能リクエストがあれば、ぜひ [GitHub Issues](https://github.com/higakikeita/tfdrift-falco/issues) でお知らせください！
