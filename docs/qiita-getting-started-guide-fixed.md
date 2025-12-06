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
- **CloudTrail が有効化**されており、S3 バケットにログが保存されている

### 推奨

- Linux または macOS（Windows は WSL2 推奨）
- 8GB 以上の RAM
- Slack Webhook URL（通知用）

### CloudTrail の確認

TFDrift-Falco は CloudTrail のログを監視するため、まず CloudTrail が有効化されているか確認します。

```bash
# CloudTrail が有効か確認
aws cloudtrail describe-trails

# S3 バケット名を確認
aws cloudtrail describe-trails | jq -r '.trailList[0].S3BucketName'
# → my-cloudtrail-logs-bucket （この名前を後で使います）
```

---

## セットアップ手順

### Phase 1: プロジェクトのセットアップ（5分）

#### Step 1: プロジェクトをクローン

```bash
cd ~/
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

#### Step 2: 設定ファイルを作成

```bash
# サンプル設定をコピー
cp examples/config.yaml config.yaml

# エディタで編集
vim config.yaml
```

#### Step 3: config.yaml を編集

最小限の設定例（コメントを削除してシンプルに）：

```yaml
# Falco 連携設定
falco:
  enabled: true
  hostname: falco  # Docker Compose のサービス名
  port: 5060

# AWS 設定
providers:
  aws:
    enabled: true
    regions:
      - us-east-1

    # Terraform State の場所（重要！）
    state:
      # ローカルファイルの場合
      backend: local
      local_path: /terraform/terraform.tfstate

      # S3 の場合（推奨）
      # backend: s3
      # s3_bucket: "my-terraform-state-bucket"
      # s3_key: "prod/terraform.tfstate"
      # s3_region: "us-east-1"

# ドリフトルール
drift_rules:
  - name: "EC2 Configuration Drift"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
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
      - "versioning"
      - "server_side_encryption_configuration"
    severity: "high"

# 通知設定
notifications:
  slack:
    enabled: false  # 後で設定
    webhook_url: ""
    channel: "#alerts"

  falco_output:
    enabled: true
    priority: "warning"

# ログ設定
logging:
  level: "info"
  format: "json"
```

**重要な編集ポイント**:

1. **Terraform State のパス**
   - ローカルファイルの場合: `local_path` を設定
   - S3 の場合: `backend: s3` に変更して、バケット名とキーを設定

2. **監視するリソースタイプ**
   - `drift_rules` で監視したいリソースと属性を定義

#### Step 4: 環境変数を設定

```bash
# .env ファイルを作成
cat > .env << 'EOF'
# CloudTrail S3 バケット（必須）
CLOUDTRAIL_S3_BUCKET=my-cloudtrail-logs-bucket

# AWS リージョン
AWS_REGION=us-east-1

# Terraform State ディレクトリ（ローカルの場合）
# Dockerfile でのパスを指定
TERRAFORM_STATE_DIR=/absolute/path/to/your/terraform

# Slack Webhook（オプション、後で設定）
# SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
EOF
```

**CLOUDTRAIL_S3_BUCKET の設定**:
- Phase 1 で確認した CloudTrail の S3 バケット名を入力
- 例: `my-company-cloudtrail-logs`

**TERRAFORM_STATE_DIR の設定**:
- Terraform State ファイルがあるディレクトリの**絶対パス**
- 例: `/Users/yourname/projects/terraform`
- 例: `/home/ubuntu/infrastructure/terraform`

---

### Phase 2: Docker Compose で起動（5分）

TFDrift-Falco のプロジェクトには、Falco と TFDrift-Falco の両方を起動する `docker-compose.yml` が含まれています。

#### Step 1: Docker Compose で起動

```bash
# すべて起動
docker-compose up -d

# 起動確認
docker-compose ps
```

**期待される出力**:

```
NAME                    STATUS
tfdrift-falco-falco     Up (healthy)
tfdrift-falco-app       Up (healthy)
```

#### Step 2: ログを確認

```bash
# すべてのログを確認
docker-compose logs -f

# Falco のみ
docker-compose logs -f falco

# TFDrift のみ
docker-compose logs -f tfdrift
```

#### Step 3: 正常起動の確認

**Falco のログ**:

```
Falco initialized with configuration file /etc/falco/falco.yaml
Loading rules from file /etc/falco/rules.d/terraform_drift.yaml
gRPC server threadiness equals to 0, enabling single-threaded mode
Starting gRPC server at 0.0.0.0:5060
Starting internal webserver, listening on port 8765
```

**TFDrift-Falco のログ**:

```json
{
  "level": "info",
  "msg": "Starting TFDrift-Falco v0.1.0",
  "time": "2025-12-05T12:00:00Z"
}
{
  "level": "info",
  "msg": "Connected to Falco gRPC: falco:5060",
  "time": "2025-12-05T12:00:01Z"
}
{
  "level": "info",
  "msg": "Loaded Terraform state: 42 resources",
  "time": "2025-12-05T12:00:02Z"
}
{
  "level": "info",
  "msg": "Drift detection started",
  "time": "2025-12-05T12:00:02Z"
}
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

または、環境変数で設定：

```bash
# .env ファイルに追加
echo 'SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL' >> .env
```

そして config.yaml で環境変数を参照：

```yaml
notifications:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_URL}"
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

まず、Terraform で管理された EC2 インスタンスを作成します。

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

**期待される出力（JSON形式）**:

```json
{
  "level": "info",
  "msg": "Drift detected: aws_instance.test",
  "resource_id": "i-0123456789abcdef0",
  "resource_type": "aws_instance",
  "attribute": "disable_api_termination",
  "expected": true,
  "actual": false,
  "changed_by": "john.doe@company.com",
  "arn": "arn:aws:iam::123456789012:user/john.doe",
  "event_time": "2025-12-05T12:10:23Z",
  "time": "2025-12-05T12:10:25Z"
}
{
  "level": "info",
  "msg": "Notification sent to Slack",
  "time": "2025-12-05T12:10:26Z"
}
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

## トラブルシューティング

### Q1: Falco が CloudTrail S3 バケットにアクセスできない

**エラー**:
```
Error loading CloudTrail events: Access Denied
```

**原因**:
- IAM 権限が不足
- S3 バケット名が間違っている
- `.env` ファイルの `CLOUDTRAIL_S3_BUCKET` が設定されていない

**対策**:

1. **IAM 権限を確認**

```bash
# 現在のユーザーを確認
aws sts get-caller-identity
```

必要な IAM ポリシー：

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::your-cloudtrail-bucket/*",
        "arn:aws:s3:::your-cloudtrail-bucket"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "cloudtrail:LookupEvents"
      ],
      "Resource": "*"
    }
  ]
}
```

2. **環境変数を確認**

```bash
# .env ファイルを確認
cat .env | grep CLOUDTRAIL_S3_BUCKET

# docker-compose で環境変数が読み込まれているか確認
docker-compose config | grep CLOUDTRAIL_S3_BUCKET
```

### Q2: Terraform State が読み込めない

**エラー**:
```
Failed to load Terraform state: no such file or directory
```

**原因**:
- `TERRAFORM_STATE_DIR` のパスが間違っている
- `config.yaml` の `local_path` が間違っている
- Docker ボリュームマウントの問題

**対策**:

1. **ローカルのパスを確認**

```bash
# Terraform State ファイルの存在確認
ls -la /path/to/your/terraform/terraform.tfstate
```

2. **docker-compose.yml の設定を確認**

```bash
# volumes が正しくマウントされているか確認
docker-compose config | grep -A5 volumes

# コンテナ内でファイルが見えるか確認
docker-compose exec tfdrift ls -la /terraform/
```

3. **config.yaml のパスを確認**

```yaml
state:
  backend: local
  local_path: /terraform/terraform.tfstate  # コンテナ内のパス
```

**重要**: `local_path` は**コンテナ内**のパスを指定します。
- ホスト側: `/Users/yourname/projects/terraform` → `TERRAFORM_STATE_DIR` に設定
- コンテナ内: `/terraform/terraform.tfstate` → `local_path` に設定

### Q3: Falco と TFDrift-Falco が接続できない

**エラー**:
```
Failed to connect to Falco gRPC: connection refused
```

**原因**:
- Falco がまだ起動していない
- Falco のヘルスチェックが失敗している
- ネットワーク設定の問題

**対策**:

```bash
# Falco のステータスを確認
docker-compose ps falco

# Falco のログを確認
docker-compose logs falco | grep -i error

# Falco のヘルスチェックを確認
docker-compose logs falco | grep -i grpc

# ネットワークを確認
docker network inspect tfdrift-network

# Falco を再起動
docker-compose restart falco

# すべてを再起動
docker-compose down
docker-compose up -d
```

### Q4: ドリフトが検知されない

**チェックリスト**:

1. **Falco が CloudTrail イベントを受信しているか？**
   ```bash
   docker-compose logs falco | grep cloudtrail
   ```

2. **リソースタイプがルールに含まれているか？**
   ```yaml
   drift_rules:
     - name: "Test"
       resource_types:
         - "aws_instance"  # ← これが含まれているか
   ```

3. **watched_attributes が正しいか？**
   ```yaml
   watched_attributes:
     - "disable_api_termination"  # ← 属性名が正しいか
   ```

4. **Terraform State に該当リソースが存在するか？**
   ```bash
   terraform state list | grep aws_instance.test
   ```

5. **CloudTrail のイベントが S3 に到達しているか？**
   ```bash
   aws s3 ls s3://your-cloudtrail-bucket/AWSLogs/ --recursive | tail -10
   ```

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

**config.yaml に追加**:

```yaml
# Auto-Import 設定（config.yaml の最後に追加）
auto_import:
  enabled: true
  terraform_dir: "/terraform"
  output_dir: "/terraform-imports"

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
📄 Generated code: /terraform-imports/aws_s3_bucket_my_unmanaged_bucket.tf
```

詳細: [Auto-Import ガイド](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/auto-import-guide.md)

### 3. S3 Backend の使用（推奨）

本番環境では、Terraform State を S3 に保存することを推奨します。

**config.yaml**:

```yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      s3_bucket: "my-terraform-state-bucket"
      s3_key: "prod/terraform.tfstate"
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
      - eu-west-1
```

**注意**: 各リージョンに CloudTrail を設定し、すべてのログが同じ S3 バケットに保存されている必要があります。

---

## 本番環境での運用

### 推奨構成

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

### セキュリティのベストプラクティス

1. **IAM Role の最小権限**

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::cloudtrail-bucket/*",
        "arn:aws:s3:::cloudtrail-bucket"
      ]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:GetObject"],
      "Resource": "arn:aws:s3:::terraform-state-bucket/*"
    }
  ]
}
```

2. **Secrets Manager で認証情報管理**

```bash
# Slack Webhook を Secrets Manager に保存
aws secretsmanager create-secret \
  --name tfdrift/slack-webhook \
  --secret-string "https://hooks.slack.com/services/..."
```

3. **ネットワーク分離**

- VPC 内で実行
- Security Group で 5060 ポートのアクセス制限
- Private Subnet での実行を推奨

---

## パフォーマンスとコスト

### リソース使用量

| コンポーネント | CPU | メモリ | ディスク |
|---------------|-----|--------|----------|
| Falco | 1-5% | 150MB | 100MB |
| TFDrift-Falco | 1-3% | 100MB | 50MB |
| **合計** | **<10%** | **250MB** | **150MB** |

t3.small インスタンス（$0.0208/時間）で十分動作します。

### 月間コスト試算

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
      remote_hostname: "app.terraform.io"
      remote_organization: "my-org"
      remote_workspace: "production"
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

A: CloudTrail イベント発生から通知まで、通常 **3-10 秒**です。ただし、CloudTrail がログを S3 に書き込むまでに数分かかる場合があります。SQS を使用すると遅延を短縮できます。

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
