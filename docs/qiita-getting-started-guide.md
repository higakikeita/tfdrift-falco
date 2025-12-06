# TFDrift-Falco 完全セットアップガイド - リアルタイム Terraform Drift 検知を始めよう

## はじめに

「AWS Console で誰かが設定を変更したけど、Terraform State に反映されていない...」

そんな **Terraform Drift（設定のズレ）** を**リアルタイムで検知**して、即座に Slack 通知してくれるのが **TFDrift-Falco** です。

この記事では、**ゼロから TFDrift-Falco をセットアップして、実際に動かすまで**を丁寧に解説します。

## TFDrift-Falco とは？

TFDrift-Falco は、**Falco のランタイムセキュリティ機能を使って、Terraform で管理されているリソースの設定変更をリアルタイムで検知する**OSS ツールです。

### 仕組み

```
誰かが AWS Console で EC2 の設定を変更
    ↓
CloudTrail イベントを Falco が検知（数秒以内）
    ↓
TFDrift-Falco が Terraform State と比較
    ↓
差分があれば Slack に即座に通知 🚨
```

### 特徴

- ⚡ **リアルタイム検知** - CloudTrail イベントを Falco でストリーム処理
- 🔍 **差分の詳細表示** - 期待値 vs 実際の値を比較
- 🔔 **複数の通知チャネル** - Slack、Discord、Webhook
- 📊 **Grafana ダッシュボード** - 可視化とアラート
- 🤖 **Auto-Import 機能** - 管理外リソースの自動取り込み
- 🐳 **Docker 対応** - コンテナで簡単起動

### 従来のツールとの違い

| 機能 | TFDrift-Falco | terraform plan | driftctl |
|------|--------------|----------------|----------|
| 検知方法 | **リアルタイム** | 手動実行 | 定期スキャン |
| レイテンシ | **数秒** | 手動 | 数分〜数時間 |
| ユーザー識別 | **○**（IAM ユーザー特定） | × | × |
| 通知 | **○** | × | 一部対応 |
| Auto-Import | **○** | × | × |

---

## 前提条件

### 必須

- **Docker Desktop** または **Docker Engine** がインストール済み
- **AWS CLI** が設定済み（`aws configure` 完了）
- **Terraform 1.0+** がインストール済み
- **Terraform State** が存在する（local または S3）

### 推奨

- Linux または macOS（Windows は WSL2 推奨）
- 8GB 以上の RAM
- Slack Webhook URL（通知用）

---

## セットアップ手順

### Phase 1: Falco のセットアップ（15分）

TFDrift-Falco は Falco と連携して動作します。まず Falco をセットアップします。

#### Step 1: Falco を Docker で起動

```bash
# Falco 設定ディレクトリを作成
mkdir -p ~/tfdrift-setup/falco
cd ~/tfdrift-setup/falco

# Falco 設定ファイルを作成
cat > falco.yaml << 'EOF'
# Falco configuration for TFDrift
json_output: true
json_include_output_property: true
json_include_tags_property: true

# gRPC output enabled
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 0

# CloudTrail plugin enabled
plugins:
  - name: cloudtrail
    library_path: libcloudtrail.so
    init_config:
      s3DownloadConcurrency: 10
    open_params: ""

# Load CloudTrail rules
load_plugins: [cloudtrail]

# Rule files
rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
EOF

# Falco を起動
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

#### Step 2: Falco の動作確認

```bash
# Falco のログを確認
docker logs falco

# gRPC が起動しているか確認
curl -v http://localhost:5060
# → "method not allowed" が返ればOK
```

**トラブルシューティング**:
- `AWS credentials not found` → `~/.aws/credentials` を確認
- `port already in use` → ポート 5060 を使用している他のプロセスを停止

---

### Phase 2: TFDrift-Falco のセットアップ（10分）

#### Step 1: プロジェクトをクローン

```bash
cd ~/tfdrift-setup
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

#### Step 2: 設定ファイルを作成

```bash
# サンプル設定をコピー
cp config.example.yaml config.yaml

# エディタで編集
vim config.yaml
```

**config.yaml（最小構成）**:

```yaml
# Falco 連携設定
falco:
  enabled: true
  hostname: localhost  # Docker の場合は "falco"
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
      local_path: /path/to/your/terraform.tfstate

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

# 通知設定（後で設定）
notifications:
  slack:
    enabled: false
    webhook_url: ""
    channel: "#alerts"

  falco_output:
    enabled: true
    priority: "warning"

# ログ設定
logging:
  level: "info"
  format: "text"

# Auto-Import（オプション）
auto_import:
  enabled: false
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"
  require_approval: true
```

**重要**: `state.local_path` を実際の Terraform State ファイルのパスに変更してください。

#### Step 3: Docker Compose で起動

```bash
# Docker Compose で起動
docker-compose up -d

# ログを確認
docker-compose logs -f tfdrift
```

**期待される出力**:

```
INFO[2025-12-05 12:00:00] Starting TFDrift-Falco v0.1.0
INFO[2025-12-05 12:00:00] Connected to Falco gRPC: localhost:5060
INFO[2025-12-05 12:00:01] Loaded Terraform state: 42 resources
INFO[2025-12-05 12:00:01] Drift detection started
```

---

### Phase 3: Slack 通知の設定（5分）

#### Step 1: Slack Webhook を作成

1. https://api.slack.com/apps にアクセス
2. **Create New App** → **From scratch**
3. App Name: `TFDrift-Falco`、Workspace を選択
4. **Incoming Webhooks** → **Activate Incoming Webhooks** をオン
5. **Add New Webhook to Workspace**
6. 通知先チャンネル（例: `#alerts`）を選択
7. Webhook URL をコピー（`https://hooks.slack.com/services/...`）

#### Step 2: config.yaml を更新

```yaml
notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#alerts"
```

#### Step 3: 再起動

```bash
docker-compose restart tfdrift
```

---

### Phase 4: 動作確認（10分）

実際に AWS リソースを変更して、TFDrift-Falco が検知するかテストします。

#### Step 1: テスト用 EC2 インスタンスを作成

**terraform/main.tf**:

```hcl
terraform {
  required_version = ">= 1.0"
}

provider "aws" {
  region = "us-east-1"
}

resource "aws_instance" "test" {
  ami           = "ami-0c55b159cbfafe1f0"  # Amazon Linux 2
  instance_type = "t2.micro"

  tags = {
    Name        = "tfdrift-test"
    Environment = "development"
    ManagedBy   = "terraform"
  }

  # 終了保護を有効化
  disable_api_termination = true
}

output "instance_id" {
  value = aws_instance.test.id
}
```

```bash
# Terraform で作成
cd terraform
terraform init
terraform apply

# Instance ID をメモ
terraform output instance_id
# → i-0123456789abcdef0
```

#### Step 2: AWS Console で変更

1. AWS Console → EC2 → Instances
2. `tfdrift-test` インスタンスを選択
3. **Actions** → **Instance settings** → **Change termination protection**
4. **Disable** を選択 → **Save**

#### Step 3: TFDrift-Falco のログを確認

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
INFO[2025-12-05 12:10:23] Changed by: john.doe@company.com (arn:aws:iam::123456789012:user/john.doe)
INFO[2025-12-05 12:10:23] Notification sent to Slack
```

#### Step 4: Slack を確認

Slack の `#alerts` チャンネルに以下のような通知が届きます：

```
🚨 Terraform Drift Detected

📦 Resource: aws_instance.test (i-0123456789abcdef0)
🔧 Attribute: disable_api_termination
📊 Severity: high

Expected: true
Actual:   false

👤 Changed By: john.doe@company.com
🕐 Detected At: 2025-12-05 12:10:23
🔗 CloudTrail Event: ModifyInstanceAttribute
```

**成功！** 🎉

---

## 高度な設定

### 1. Grafana ダッシュボードの追加

リアルタイムで可視化したい場合は、Grafana 統合を有効化します。

```bash
# Grafana スタックを起動
cd dashboards/grafana
./quick-start.sh
```

→ http://localhost:3000 で Grafana ダッシュボードが開きます（admin/admin）

詳細: [Grafana セットアップガイド](https://github.com/higakikeita/tfdrift-falco/blob/main/dashboards/grafana/GETTING_STARTED.md)

### 2. Auto-Import の有効化

管理外リソースを自動で Terraform に取り込みたい場合：

**config.yaml**:

```yaml
auto_import:
  enabled: true
  terraform_dir: "./infrastructure"
  output_dir: "./infrastructure/imported"

  # 許可するリソースタイプ
  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
    - "aws_s3_bucket"

  # 承認が必要（推奨）
  require_approval: true
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

### 3. S3 Backend の使用

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

### 4. 複数リージョンの監視

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
      - us-west-2
      - ap-northeast-1
```

Falco の設定でも対応するリージョンを追加してください。

---

## 本番環境での運用

### 推奨構成

```
┌─────────────────────────────────────────┐
│         AWS Account (Production)        │
│                                          │
│  ┌────────────────┐                     │
│  │  EC2 Instance  │                     │
│  │  (App Server)  │                     │
│  └────────────────┘                     │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │         ECS Cluster                │ │
│  │                                    │ │
│  │  ┌──────────┐  ┌───────────────┐ │ │
│  │  │  Falco   │  │ TFDrift-Falco │ │ │
│  │  │  Task    │→ │     Task      │ │ │
│  │  └──────────┘  └───────────────┘ │ │
│  │                        ↓          │ │
│  └────────────────────────┼──────────┘ │
│                            ↓             │
└────────────────────────────┼─────────────┘
                             ↓
                    ┌────────────────┐
                    │  Slack/Email   │
                    └────────────────┘
```

### ECS での実行例

**task-definition.json**:

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
      "essential": true,
      "command": ["--disable-source", "syscall"],
      "portMappings": [
        {
          "containerPort": 5060,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "AWS_REGION",
          "value": "us-east-1"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/tfdrift-falco",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "falco"
        }
      }
    },
    {
      "name": "tfdrift",
      "image": "ghcr.io/higakikeita/tfdrift-falco:latest",
      "essential": true,
      "dependsOn": [
        {
          "containerName": "falco",
          "condition": "START"
        }
      ],
      "environment": [
        {
          "name": "TFDRIFT_FALCO_HOSTNAME",
          "value": "localhost"
        }
      ],
      "secrets": [
        {
          "name": "TFDRIFT_SLACK_WEBHOOK_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789012:secret:tfdrift/slack-webhook"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/tfdrift-falco",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "tfdrift"
        }
      }
    }
  ]
}
```

### セキュリティのベストプラクティス

1. **IAM Role の最小権限**

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
      "Action": [
        "s3:GetObject"
      ],
      "Resource": "arn:aws:s3:::my-terraform-state/*"
    }
  ]
}
```

2. **Slack Webhook の保護**

- Secrets Manager または Parameter Store で管理
- 環境変数として注入
- ハードコードしない

3. **ネットワーク分離**

- VPC 内で実行
- Security Group で 5060 ポートのアクセス制限
- Private Subnet での実行を推奨

---

## トラブルシューティング

### Q1: Falco に接続できない

**エラー**:
```
ERRO[2025-12-05] Failed to connect to Falco gRPC: connection refused
```

**対策**:
```bash
# Falco が起動しているか確認
docker ps | grep falco

# Falco のログを確認
docker logs falco | grep gRPC

# ポートが開いているか確認
netstat -an | grep 5060
```

### Q2: Terraform State が読み込めない

**エラー**:
```
ERRO[2025-12-05] Failed to load Terraform state: file not found
```

**対策**:
```bash
# パスを確認
ls -la /path/to/terraform.tfstate

# config.yaml のパスを絶対パスに変更
state:
  backend: local
  local_path: /absolute/path/to/terraform.tfstate
```

### Q3: AWS 認証エラー

**エラー**:
```
ERRO[2025-12-05] AWS authentication failed: no credentials found
```

**対策**:
```bash
# AWS CLI が設定されているか確認
aws sts get-caller-identity

# Docker で AWS 認証情報をマウント
docker-compose.yaml:
  volumes:
    - ~/.aws:/root/.aws:ro
```

### Q4: ドリフトが検知されない

**チェックリスト**:

1. Falco が CloudTrail イベントを受信しているか？
   ```bash
   docker logs falco | grep cloudtrail
   ```

2. リソースタイプがルールに含まれているか？
   ```yaml
   drift_rules:
     - name: "Test"
       resource_types:
         - "aws_instance"  # ← これが含まれているか
   ```

3. watched_attributes が正しいか？
   ```yaml
   watched_attributes:
     - "disable_api_termination"  # ← 属性名が正しいか
   ```

4. Terraform State に該当リソースが存在するか？
   ```bash
   terraform state list | grep aws_instance.test
   ```

---

## 実用例

### ユースケース 1: セキュリティチームの監視

**シナリオ**: IAM ロールや S3 バケットの設定変更を即座に検知

**設定**:

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
    enabled: true
    webhook_url: "${SECURITY_TEAM_WEBHOOK}"
    channel: "#security-alerts"
```

**効果**:
- 不正なアクセス権限変更を数秒で検知
- 誰が変更したか特定可能
- インシデント対応時間を大幅短縮

### ユースケース 2: 本番環境の変更管理

**シナリオ**: 本番環境への手動変更を禁止し、IaC 経由のみを許可

**設定**:

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

### ユースケース 3: マルチアカウント監視

**シナリオ**: 複数の AWS アカウントを一元監視

**構成**:

```
Account A (Production)
  → TFDrift-Falco Instance A → Slack #prod-alerts

Account B (Staging)
  → TFDrift-Falco Instance B → Slack #staging-alerts

Account C (Development)
  → TFDrift-Falco Instance C → Slack #dev-alerts
```

各アカウントで独立して TFDrift-Falco を実行し、それぞれ異なる Slack チャンネルに通知。

---

## パフォーマンスとコスト

### リソース使用量

| コンポーネント | CPU | メモリ | ディスク |
|---------------|-----|--------|----------|
| Falco | 1-5% | 150MB | 100MB |
| TFDrift-Falco | 1-3% | 100MB | 50MB |
| **合計** | **<10%** | **250MB** | **150MB** |

t3.small インスタンス（$0.0208/時間）で十分動作します。

### 月間コスト（参考）

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

## よくある質問（FAQ）

### Q: Terraform Cloud に対応していますか？

A: はい。Terraform Cloud の Remote State に対応しています。

```yaml
providers:
  aws:
    state:
      backend: remote
      organization: "my-org"
      workspace: "production"
```

API Token は環境変数 `TF_CLOUD_TOKEN` で設定してください。

### Q: 既存の terraform plan との違いは？

A:

| 比較項目 | terraform plan | TFDrift-Falco |
|---------|---------------|--------------|
| 実行タイミング | 手動 | **リアルタイム** |
| 検知速度 | 数分〜数時間 | **数秒** |
| ユーザー特定 | × | **○** |
| 自動通知 | × | **○** |

TFDrift-Falco は `terraform plan` を置き換えるものではなく、**補完する**ツールです。

### Q: CloudTrail の費用が心配です

A: CloudTrail は最初の 100,000 イベント/月が無料です。通常の利用であれば追加費用はほとんど発生しません。

### Q: GCP や Azure に対応していますか？

A: 現在は AWS のみ対応。GCP、Azure は Phase 2 で対応予定です（2025年 Q2 予定）。

### Q: 検知の遅延はどのくらいですか？

A: CloudTrail イベント発生から通知まで、通常 **3-10 秒**です。

---

## まとめ

TFDrift-Falco を使えば：

✅ **リアルタイムで Drift を検知** - 手動変更を見逃さない
✅ **誰が変更したか特定** - インシデント対応が迅速化
✅ **自動通知で即座に対応** - Slack で関係者に通知
✅ **Grafana で可視化** - トレンド分析とダッシュボード
✅ **Auto-Import で自動化** - 管理外リソースを自動取り込み

特に、**セキュリティ重視の環境**や**変更管理が厳格な本番環境**で威力を発揮します。

## 次のステップ

1. ✅ [GitHub リポジトリ](https://github.com/higakikeita/tfdrift-falco) を Star ⭐
2. ✅ サンプル環境で試してみる
3. ✅ Slack 通知を設定
4. ✅ Grafana ダッシュボードを追加
5. ✅ 本番環境へのデプロイ

## リンク

- **GitHub**: https://github.com/higakikeita/tfdrift-falco
- **Grafana セットアップガイド**: [dashboards/grafana/GETTING_STARTED.md](https://github.com/higakikeita/tfdrift-falco/blob/main/dashboards/grafana/GETTING_STARTED.md)
- **Auto-Import ガイド**: [docs/auto-import-guide.md](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/auto-import-guide.md)
- **Issue / 質問**: https://github.com/higakikeita/tfdrift-falco/issues

## フィードバック募集中！

使ってみた感想や、機能リクエストがあれば、ぜひ [GitHub Issues](https://github.com/higakikeita/tfdrift-falco/issues) でお知らせください！

---

**タグ**: #Terraform #AWS #Falco #IaC #DevOps #CloudSecurity #OSS #InfrastructureAsCode
