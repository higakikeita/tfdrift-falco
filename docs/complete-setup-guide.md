# TFDrift-Falco 完全セットアップガイド

## 概要

このガイドでは、AWSのセットアップからTerraform統合、TFDrift-Falcoの設定、そして実際のドリフト検知までの完全な流れを説明します。

**対象読者**: AWS環境でTerraformを使用しており、インフラストラクチャのドリフト（手動変更）をリアルタイムで検知したい方

**所要時間**: 約1-2時間（CloudTrailログの書き込み待ち時間を除く）

---

## アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│                         AWS Cloud                            │
│                                                              │
│  ┌──────────────┐      ┌─────────────────┐                 │
│  │  Terraform   │      │   CloudTrail    │                 │
│  │    State     │◄─────┤   (API Logs)    │                 │
│  │   (S3)       │      │   S3 Bucket     │                 │
│  └──────────────┘      └─────────────────┘                 │
│         │                       │                            │
└─────────┼───────────────────────┼────────────────────────────┘
          │                       │
          │ Read State            │ Read Events
          ▼                       ▼
┌─────────────────────────────────────────────────────────────┐
│                    TFDrift-Falco System                      │
│                                                              │
│  ┌──────────────┐      ┌─────────────────┐                 │
│  │    Falco     │      │    Backend      │                 │
│  │  CloudTrail  │─────►│   API Server    │                 │
│  │   Plugin     │ gRPC │  (Go + Fiber)   │                 │
│  └──────────────┘      └─────────────────┘                 │
│                               │                              │
│                               │ REST API / WebSocket         │
│                               ▼                              │
│                        ┌─────────────────┐                  │
│                        │    Frontend     │                  │
│                        │   (React UI)    │                  │
│                        └─────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

---

## 前提条件

### 必須要件

1. **AWS環境**
   - AWS CLI設定済み（`aws configure`完了）
   - 適切なIAMパーミッション（CloudTrail、S3、IAM）
   - Terraform管理下のリソースが存在

2. **ローカル環境**
   - Docker Desktop（Rosetta有効 - ARM64 Macの場合）
   - Terraform 1.0以上
   - Git
   - 8GB以上のメモリ推奨

3. **ネットワーク**
   - ポート3000（Frontend）が空いている
   - ポート8080（Backend）が空いている
   - ポート5060（Falco gRPC）が空いている

### 推奨環境

- macOS (ARM64): Docker Desktop with Rosetta
- macOS (Intel): Docker Desktop
- Linux: Docker + Docker Compose

---

## Phase 1: AWS CloudTrailのセットアップ

### 1.1 CloudTrailとS3バケットの作成

TFDrift-Falcoは、AWS CloudTrailのログからリアルタイムでAPIコールを検知します。

#### 自動セットアップ（推奨）

プロジェクトに含まれるセットアップスクリプトを使用します。

```bash
cd /path/to/tfdrift-falco
chmod +x scripts/setup-cloudtrail.sh
./scripts/setup-cloudtrail.sh
```

このスクリプトは以下を自動で実行します：
1. S3バケット作成（リージョン固有の命名）
2. CloudTrail用バケットポリシー設定
3. CloudTrailトレイルの作成と有効化
4. マルチリージョン対応の有効化

#### 手動セットアップ

自動スクリプトを使用しない場合は、以下の手順に従います。

**Step 1: S3バケットの作成**

```bash
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
AWS_REGION="us-east-1"  # お好みのリージョン
BUCKET_NAME="tfdrift-cloudtrail-${AWS_ACCOUNT_ID}-${AWS_REGION}"

aws s3api create-bucket \
    --bucket ${BUCKET_NAME} \
    --region ${AWS_REGION}
```

**Step 2: バケットポリシーの設定**

CloudTrailサービスがログを書き込めるようにポリシーを設定します。

```bash
cat > /tmp/bucket-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "AWSCloudTrailAclCheck",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "s3:GetBucketAcl",
      "Resource": "arn:aws:s3:::${BUCKET_NAME}"
    },
    {
      "Sid": "AWSCloudTrailWrite",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.amazonaws.com"
      },
      "Action": "s3:PutObject",
      "Resource": "arn:aws:s3:::${BUCKET_NAME}/AWSLogs/${AWS_ACCOUNT_ID}/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "bucket-owner-full-control"
        }
      }
    }
  ]
}
EOF

aws s3api put-bucket-policy \
    --bucket ${BUCKET_NAME} \
    --policy file:///tmp/bucket-policy.json
```

**Step 3: CloudTrailトレイルの作成**

```bash
TRAIL_NAME="tfdrift-falco-trail"

aws cloudtrail create-trail \
    --name ${TRAIL_NAME} \
    --s3-bucket-name ${BUCKET_NAME} \
    --is-multi-region-trail \
    --region ${AWS_REGION}

aws cloudtrail start-logging \
    --name ${TRAIL_NAME} \
    --region ${AWS_REGION}
```

### 1.2 セットアップの確認

CloudTrailが正しく設定されているか確認します。

```bash
# トレイルのステータス確認
aws cloudtrail get-trail-status --name ${TRAIL_NAME} --region ${AWS_REGION}

# 期待される出力: "IsLogging": true
```

**重要**: CloudTrailがS3にログを書き込み始めるまで **5-15分** かかります。この間に次のステップを進めることができます。

### 1.3 CloudTrail設定のメモ

後で使用するため、以下の情報を記録しておきます：

```bash
# これらの値を後でFalco設定に使用します
echo "S3 Bucket: ${BUCKET_NAME}"
echo "Trail Name: ${TRAIL_NAME}"
echo "Region: ${AWS_REGION}"
```

---

## Phase 2: Terraform Stateの設定

TFDrift-Falcoは、Terraform Stateと実際のAWSリソースを比較してドリフトを検知します。

### 2.1 Terraform StateのバックエンドをS3に設定（推奨）

既にTerraformを使用している場合、S3バックエンドを使用することを推奨します。

**terraform/main.tf の例:**

```hcl
terraform {
  backend "s3" {
    bucket = "your-terraform-state-bucket"
    key    = "production/terraform.tfstate"
    region = "us-east-1"
  }
}
```

### 2.2 ローカルStateファイルを使用する場合

S3バックエンドを使用しない場合は、ローカルのStateファイルをマウントすることもできます。

```bash
# Stateファイルのパスを確認
ls -la terraform/production-like-environment/errored.tfstate
```

### 2.3 TFDrift-Falcoのconfig.yaml設定

プロジェクトの`config.yaml`を編集します。

**S3バックエンドを使用する場合:**

```yaml
# config.yaml
state:
  backend: "s3"
  s3_bucket: "your-terraform-state-bucket"
  s3_key: "production/terraform.tfstate"
  s3_region: "us-east-1"
```

**ローカルStateを使用する場合:**

```yaml
# config.yaml
state:
  backend: "local"
  local_path: "/terraform/production-like-environment/errored.tfstate"
```

**重要**: S3バックエンドを使用する場合は、後でAWS認証情報を正しく設定する必要があります。

---

## Phase 3: Falco CloudTrailプラグインのセットアップ

### 3.1 プラグインのダウンロード

FalcoのCloudTrailプラグインをダウンロードします。

```bash
cd deployments/falco/plugins

# プラットフォームに応じたプラグインをダウンロード
# ARM64 Mac の場合は x86_64 版を使用（Rosetta経由）
PLUGIN_VERSION="0.13.0"
PLATFORM="linux-x86_64"  # ARM64の場合もこれを使用

curl -L -o cloudtrail-plugin.tar.gz \
    "https://download.falco.org/plugins/stable/cloudtrail-${PLUGIN_VERSION}-${PLATFORM}.tar.gz"

# 展開
tar -xzf cloudtrail-plugin.tar.gz
rm cloudtrail-plugin.tar.gz

# 確認
ls -lh libcloudtrail.so
```

### 3.2 Falco設定ファイルの編集

`deployments/falco/falco-simple.yaml`を編集し、CloudTrailプラグインを有効化します。

```yaml
# deployments/falco/falco-simple.yaml

# CloudTrail Plugin Configuration
plugins:
  - name: cloudtrail
    library_path: /etc/falco/plugins/libcloudtrail.so
    init_config: ""
    open_params: "s3Bucket=tfdrift-cloudtrail-595263720623-us-east-1"  # ← あなたのバケット名

# Load plugins
load_plugins: [cloudtrail]
```

**注意**: `s3Bucket=`の後には、Phase 1で作成したS3バケット名を指定します。

### 3.3 Falcoルールの確認

`rules/terraform_drift.yaml`に、ドリフト検知ルールが含まれていることを確認します。

**簡易版ルールの例:**

```yaml
# rules/terraform_drift.yaml

- rule: AWS EC2 Modification
  desc: Detect EC2 instance modifications
  condition: >
    ct.name in ("ModifyInstanceAttribute", "StartInstances", "StopInstances",
    "TerminateInstances", "RebootInstances")
  output: >
    AWS EC2 modification detected (event=%ct.name user=%ct.user.identity.principalid region=%ct.region)
  priority: WARNING
  source: aws_cloudtrail
  tags: [terraform, drift, ec2]

- rule: AWS IAM Modification
  desc: Detect IAM policy and role changes
  condition: >
    ct.name in ("PutRolePolicy", "DeleteRolePolicy", "UpdateAssumeRolePolicy",
    "PutUserPolicy", "DeleteUserPolicy", "CreateRole", "DeleteRole")
  output: >
    AWS IAM modification detected (event=%ct.name user=%ct.user.identity.principalid)
  priority: CRITICAL
  source: aws_cloudtrail
  tags: [terraform, drift, iam, security]
```

**ルール作成のポイント:**
- `condition`: CloudTrailのイベント名（API操作）
- `output`: アラート出力フォーマット
- `priority`: WARNING, CRITICAL, ERROR, DEBUG
- `source`: aws_cloudtrail を指定
- `tags`: 分類タグ

---

## Phase 4: Docker環境のセットアップ

### 4.1 docker-compose.ymlの修正

#### ステップ1: versionフィールドを削除

Docker Compose v2では`version`フィールドが廃止されました。

```yaml
# 削除する
# version: '3.8'

services:
  # ...
```

#### ステップ2: AWS認証情報の設定

バックエンドコンテナにAWS認証情報を渡すため、環境変数を追加します。

```yaml
# docker-compose.yml

services:
  falco:
    image: falcosecurity/falco:0.37.1
    container_name: tfdrift-falco-falco
    platform: linux/amd64  # ARM64 Macの場合は必須
    # ... その他の設定 ...
    environment:
      - AWS_REGION=${AWS_REGION:-us-east-1}
      - AWS_PROFILE=${AWS_PROFILE:-default}  # 追加
    volumes:
      - ${HOME}/.aws:/root/.aws:ro

  backend:
    # ... その他の設定 ...
    environment:
      - AWS_REGION=${AWS_REGION:-us-east-1}
      - AWS_PROFILE=${AWS_PROFILE:-default}  # 追加
      - TZ=${TZ:-UTC}
      - TFDRIFT_FALCO_HOSTNAME=falco
      - TFDRIFT_FALCO_PORT=5060
    volumes:
      - ./config.yaml:/config/config.yaml:ro
      - ${HOME}/.aws:/root/.aws:ro  # AWS認証情報
      - ./terraform:/terraform:ro   # Terraform State
```

### 4.2 環境変数の設定

使用するAWSプロファイルとリージョンを設定します。

```bash
# .envファイルを作成（またはexport）
cat > .env <<EOF
AWS_PROFILE=mytf
AWS_REGION=us-east-1
TZ=Asia/Tokyo
EOF
```

または、直接exportする場合：

```bash
export AWS_PROFILE=mytf
export AWS_REGION=us-east-1
```

### 4.3 ARM64 Mac対応（該当する場合）

ARM64 Mac（Apple Silicon）を使用している場合、以下の設定が必要です。

#### Docker Desktopの設定確認

1. Docker Desktop を開く
2. Settings → General
3. "Use Rosetta for x86_64/amd64 emulation on Apple Silicon" にチェック
4. Docker Desktop を再起動

#### docker-compose.ymlでプラットフォーム指定

```yaml
services:
  falco:
    platform: linux/amd64  # この行が重要
```

**理由**: CloudTrailプラグインはx86_64版のみ提供されているため、ARM64環境ではRosetta経由でエミュレーションが必要です。

---

## Phase 5: TFDrift-Falcoの起動

### 5.1 イメージのビルド

```bash
# プロジェクトルートで実行
cd /path/to/tfdrift-falco

# バックエンドとフロントエンドをビルド
docker-compose build
```

**ビルド時間**: 初回は5-10分程度かかります。

### 5.2 サービスの起動

```bash
# すべてのサービスを起動
docker-compose up -d

# ログを確認
docker-compose logs -f
```

### 5.3 起動確認

各サービスが正常に起動しているか確認します。

```bash
# コンテナステータス確認
docker-compose ps

# 期待される出力:
# NAME                   STATUS    PORTS
# tfdrift-backend        Up        0.0.0.0:8080->8080/tcp
# tfdrift-frontend       Up        0.0.0.0:3000->8080/tcp
# tfdrift-falco-falco    Up        0.0.0.0:5060->5060/tcp
```

#### ヘルスチェック

```bash
# Backend API
curl http://localhost:8080/health
# 期待される出力: {"status":"ok"}

# Frontend
curl http://localhost:3000/health
# 期待される出力: OK

# Falco gRPC（grpcurlが必要）
grpcurl -plaintext localhost:5060 list
```

---

## Phase 6: 動作確認とドリフト検知テスト

### 6.1 UIへのアクセス

ブラウザで以下のURLにアクセスします。

```
http://localhost:3000
```

**期待される画面:**
- グラフビュー: Terraform管理リソースの依存関係グラフ
- テーブルビュー: ドリフトイベント履歴
- 分割ビュー: グラフとテーブルを同時表示（推奨）

### 6.2 初期データの確認

UIが起動したら、以下を確認します：

1. **グラフビュー**: Terraform Stateから読み込んだリソースが表示される
2. **テーブルビュー**: 初期状態では空またはサンプルデータ
3. **API接続**: 右上にWebSocket接続ステータスが表示される

### 6.3 リアルタイムドリフト検知のテスト

実際にAWSコンソールでリソースを変更し、ドリフトが検知されることを確認します。

#### テストシナリオ1: EC2インスタンスの停止/起動

```bash
# Terraform管理下のEC2インスタンスIDを確認
terraform show | grep instance_id

# AWSコンソールまたはCLIでインスタンスを停止
aws ec2 stop-instances --instance-ids i-xxxxxxxxxxxxx --region us-east-1

# 数秒後、TFDrift-Falco UIでアラートが表示される
```

**期待される動作:**
1. CloudTrailが`StopInstances` APIコールを記録
2. FalcoがCloudTrailログから検知
3. BackendがFalcoからイベント受信
4. UIにリアルタイムでアラート表示

#### テストシナリオ2: セキュリティグループの変更

```bash
# セキュリティグループのルールを追加
aws ec2 authorize-security-group-ingress \
    --group-id sg-xxxxxxxxxxxxx \
    --protocol tcp \
    --port 22 \
    --cidr 0.0.0.0/0 \
    --region us-east-1
```

**期待される動作:**
- `AWS Security Group Modification` アラートが発火
- Priority: WARNING
- ドリフト履歴テーブルに新しい行が追加

#### テストシナリオ3: IAMロールの変更（高リスク）

```bash
# IAMロールポリシーを変更
aws iam put-role-policy \
    --role-name MyTerraformManagedRole \
    --policy-name TestPolicy \
    --policy-document file://test-policy.json
```

**期待される動作:**
- `AWS IAM Modification` アラートが発火
- Priority: CRITICAL（赤色表示）
- セキュリティタグ付き

### 6.4 ドリフト詳細の確認

テーブルビューで任意のドリフトイベントをクリックすると、詳細パネルが表示されます。

**表示内容:**
- イベント種別（CloudTrailイベント名）
- 実行ユーザー（AWS IAMプリンシパル）
- タイムスタンプ
- リージョン
- 変更前後の値（該当する場合）
- 推奨アクション

---

## Phase 7: トラブルシューティング

### 7.1 よくあるエラーと解決方法

#### エラー1: AWS認証情報エラー

**症状:**
```
NoCredentialProviders: no valid providers in chain
failed to load terraform state
```

**原因**: バックエンドコンテナがAWS認証情報にアクセスできない

**解決策:**

1. AWS_PROFILEが正しく設定されているか確認
   ```bash
   echo $AWS_PROFILE
   ```

2. `~/.aws/credentials`ファイルが存在するか確認
   ```bash
   ls -la ~/.aws/credentials
   ```

3. docker-compose.ymlで認証情報がマウントされているか確認
   ```yaml
   volumes:
     - ${HOME}/.aws:/root/.aws:ro
   environment:
     - AWS_PROFILE=${AWS_PROFILE:-default}
   ```

4. コンテナ内で認証情報を確認
   ```bash
   docker exec -it tfdrift-backend ls -la /root/.aws
   docker exec -it tfdrift-backend cat /root/.aws/credentials
   ```

#### エラー2: Falco eBPFドライバーのコンパイル失敗

**症状:**
```
Error! Your kernel headers for kernel 6.10.14-linuxkit cannot be found
failed: failed to compile the module
```

**原因**: Docker Desktop on Macではカーネルヘッダーが利用できない

**解決策:**

1. `falco-simple.yaml`で`modern_ebpf`を使用する設定になっているか確認
   ```yaml
   engine:
     kind: modern_ebpf
   ```

2. それでも失敗する場合は、一時的にFalcoを無効化してUI/Backendのみテスト
   ```yaml
   # docker-compose.yml
   services:
     backend:
       # depends_on:
       #   falco:
       #     condition: service_healthy
   ```

3. Falcoなしでバックエンドを起動
   ```bash
   docker-compose up -d backend frontend
   ```

#### エラー3: CloudTrailプラグインのS3接続エラー

**症状:**
```
cloudtrail plugin error: cannot open s3Bucket=...
```

**原因**:
- AWS認証情報が設定されていない
- CloudTrailログがまだS3に書き込まれていない（初回は5-15分かかる）

**解決策:**

1. CloudTrailのロギングが有効か確認
   ```bash
   aws cloudtrail get-trail-status --name tfdrift-falco-trail --region us-east-1
   ```

2. S3バケットにログが書き込まれているか確認
   ```bash
   aws s3 ls s3://tfdrift-cloudtrail-xxxxx/AWSLogs/ --recursive
   ```

3. 5-15分待ってから再試行

4. それでも解決しない場合は、一時的にCloudTrailプラグインを無効化
   ```yaml
   # falco-simple.yaml
   # plugins:
   #   - name: cloudtrail
   #     ...
   # load_plugins: [cloudtrail]
   ```

#### エラー4: ポート競合

**症状:**
```
Error: bind: address already in use
```

**解決策:**

1. 使用中のポートを確認
   ```bash
   lsof -i :8080
   lsof -i :3000
   lsof -i :5060
   ```

2. 競合するプロセスを停止するか、docker-compose.ymlでポートを変更
   ```yaml
   ports:
     - "8081:8080"  # ホスト側のポートを変更
   ```

#### エラー5: Terraform State読み込み失敗

**症状:**
```
failed to load state from s3 backend
```

**解決策:**

1. S3バケットとキーが正しいか確認
   ```bash
   aws s3 ls s3://your-terraform-state-bucket/production/
   ```

2. IAMポリシーでS3読み取り権限があるか確認
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": ["s3:GetObject"],
         "Resource": "arn:aws:s3:::your-terraform-state-bucket/*"
       }
     ]
   }
   ```

3. ローカルStateファイルを使用する場合は、パスが正しいか確認
   ```bash
   ls -la ./terraform/production-like-environment/errored.tfstate
   ```

### 7.2 ログの確認方法

問題を診断するため、各サービスのログを確認します。

```bash
# すべてのサービスのログ
docker-compose logs -f

# 特定のサービスのログ
docker-compose logs -f backend
docker-compose logs -f falco
docker-compose logs -f frontend

# 最新100行のログ
docker-compose logs --tail=100 backend
```

### 7.3 デバッグモード

より詳細なログが必要な場合は、デバッグモードを有効化します。

```yaml
# docker-compose.yml
services:
  backend:
    command: ["--server", "--api-port", "8080", "--config", "/config/config.yaml", "--debug"]
```

```yaml
# falco-simple.yaml
log_level: debug
```

---

## Phase 8: 本番環境へのデプロイ

### 8.1 本番環境の考慮事項

#### セキュリティ

1. **Falco gRPC通信のTLS証明書**
   ```bash
   # 本番用の証明書を生成
   openssl req -x509 -newkey rsa:4096 -keyout server.key -out server.crt -days 365 -nodes
   ```

2. **API認証の追加**
   - Backendにベーシック認証またはJWTを実装
   - Frontendにログイン機能を追加

3. **ネットワークセグメンテーション**
   - FalcoとBackendは内部ネットワーク
   - FrontendのみPublicアクセス可能

#### スケーラビリティ

1. **複数リージョン対応**
   - リージョンごとにCloudTrailを設定
   - S3バケットは各リージョンに配置

2. **ロードバランシング**
   - 複数のBackendインスタンスを起動
   - Nginx/HAProxyでロードバランシング

3. **データ永続化**
   - PostgreSQLまたはMySQLでドリフト履歴を保存
   - Elasticsearchでログ検索機能

#### 監視とアラート

1. **Slackインテグレーション**
   ```yaml
   # docker-compose.yml
   environment:
     - TFDRIFT_SLACK_WEBHOOK=https://hooks.slack.com/services/xxx
   ```

2. **Prometheusメトリクス**
   - Backend: `http://localhost:9090/metrics`
   - アラート数、処理時間などを監視

3. **ヘルスチェック**
   - Kubernetes Liveness/Readiness Probes
   - AWS ECS Health Checks

### 8.2 Kubernetesデプロイ（参考）

```yaml
# k8s-deployment.yaml (簡易版)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tfdrift-falco
spec:
  replicas: 2
  selector:
    matchLabels:
      app: tfdrift-falco
  template:
    metadata:
      labels:
        app: tfdrift-falco
    spec:
      containers:
      - name: falco
        image: falcosecurity/falco:0.37.1
        volumeMounts:
        - name: aws-credentials
          mountPath: /root/.aws
          readOnly: true
      - name: backend
        image: tfdrift-falco:latest
        env:
        - name: AWS_REGION
          value: "us-east-1"
      volumes:
      - name: aws-credentials
        secret:
          secretName: aws-credentials
```

---

## Phase 9: 運用とメンテナンス

### 9.1 定期的なメンテナンス

#### 週次タスク

1. **ドリフト履歴のレビュー**
   - 頻繁に発生するドリフトのパターン確認
   - Terraformコードへの反映が必要か判断

2. **Falcoルールの調整**
   - False Positiveの削減
   - 新しいAWSサービスへの対応

#### 月次タスク

1. **CloudTrailログのクリーンアップ**
   ```bash
   # 90日以上古いログを削除（S3ライフサイクルポリシー）
   aws s3api put-bucket-lifecycle-configuration \
       --bucket tfdrift-cloudtrail-xxx \
       --lifecycle-configuration file://lifecycle-policy.json
   ```

2. **Dockerイメージの更新**
   ```bash
   docker-compose pull
   docker-compose up -d --build
   ```

### 9.2 ドリフト対応のワークフロー

#### ドリフト検知時の対応フロー

```
ドリフト検知
    │
    ▼
┌─────────────────┐
│  重要度の判定    │  Priority: CRITICAL / WARNING / DEBUG
└─────────────────┘
    │
    ▼
┌─────────────────┐
│  原因の調査      │  誰が？なぜ？どのリソース？
└─────────────────┘
    │
    ├──→ 意図的な変更：Terraformコードに反映
    │
    ├──→ 緊急対応：許可済み（ログに記録）
    │
    └──→ 不正な変更：元に戻す + インシデント報告
```

#### Terraformコードへの反映手順

1. ドリフトをTerraformコードにインポート
   ```bash
   terraform import aws_instance.example i-xxxxxxxxxxxxx
   ```

2. コードを更新
   ```hcl
   resource "aws_instance" "example" {
     # ドリフトした設定を反映
     instance_type = "t3.medium"  # 変更された値
   }
   ```

3. Terraform planで確認
   ```bash
   terraform plan
   # No changes expected
   ```

4. PR/コミット/レビュー

### 9.3 監視ダッシュボードの例

**Grafanaダッシュボード（参考）**

```
┌─────────────────────────────────────────────────────────┐
│  TFDrift-Falco Monitoring Dashboard                     │
├─────────────────────────────────────────────────────────┤
│  Total Drifts: 127        Today: 8       Critical: 2    │
├─────────────────────────────────────────────────────────┤
│  Drift by Service                                       │
│  ██████████ EC2 (45)                                    │
│  ██████ IAM (22)                                        │
│  ████ S3 (18)                                           │
│  ███ RDS (12)                                           │
├─────────────────────────────────────────────────────────┤
│  Drift Timeline (Last 24h)                              │
│  [グラフ]                                                │
└─────────────────────────────────────────────────────────┘
```

---

## まとめ

### 完了したこと

- ✅ AWS CloudTrailのセットアップ
- ✅ Terraform State連携
- ✅ Falco CloudTrailプラグインのインストール
- ✅ TFDrift-Falcoシステムの起動
- ✅ リアルタイムドリフト検知のテスト

### 次のステップ

1. **チームへの展開**: 他のチームメンバーがセットアップできるようドキュメント共有
2. **ルールのカスタマイズ**: 組織固有のポリシーに合わせてFalcoルールを調整
3. **アラート統合**: Slack/PagerDuty/Emailなどへの通知設定
4. **ダッシュボード改善**: より詳細な分析とレポート機能の追加

### 参考リンク

- [Falco公式ドキュメント](https://falco.org/docs/)
- [CloudTrailプラグイン](https://github.com/falcosecurity/plugins/tree/master/plugins/cloudtrail)
- [TFDrift-Falco GitHubリポジトリ](https://github.com/yourusername/tfdrift-falco)
- [AWS CloudTrailドキュメント](https://docs.aws.amazon.com/cloudtrail/)
- [Terraformドキュメント](https://www.terraform.io/docs)

---

**最終更新**: 2025-12-22

**著者**: TFDrift-Falco プロジェクト

**ライセンス**: MIT
