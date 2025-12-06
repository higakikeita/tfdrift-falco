# セットアップ手順の検証レポート

## 現状の問題点

### 1. config.example.yaml が存在しない

**記事で提案している手順**:
```bash
cp config.example.yaml config.yaml
```

**実際のプロジェクト構造**:
- `config.example.yaml` は存在しない ❌
- `examples/config.yaml` が存在する ✅

**修正案**:
```bash
cp examples/config.yaml config.yaml
```

### 2. Falco の設定が複雑

**記事で提案している設定**:
- 簡易的な falco.yaml を手動作成

**実際のプロジェクト**:
- `deployments/falco/falco.yaml` が既に存在
- CloudTrail S3 バケットの環境変数が必要
- より複雑な構成

**修正案**:
プロジェクトの既存の docker-compose.yml を使用する方法を推奨

### 3. Docker Compose の構成が異なる

**記事で提案している手順**:
1. Falco を単独で起動
2. TFDrift-Falco を別途起動

**実際のプロジェクト**:
- `docker-compose.yml` に Falco と TFDrift-Falco の両方が含まれている
- 一度に起動可能

**修正案**:
既存の docker-compose.yml を活用する手順に変更

### 4. 必要な環境変数

**不足している情報**:
- `CLOUDTRAIL_S3_BUCKET` 環境変数が必要
- CloudTrail ログが保存されている S3 バケット名

## 正しいセットアップ手順（修正版）

### Phase 1: 前提条件の確認（5分）

1. **CloudTrail の確認**
   ```bash
   # CloudTrail が有効か確認
   aws cloudtrail describe-trails

   # S3 バケット名を確認
   aws cloudtrail describe-trails | jq -r '.trailList[0].S3BucketName'
   ```

2. **必要なツールの確認**
   ```bash
   docker --version
   docker-compose --version
   terraform --version
   aws --version
   ```

### Phase 2: プロジェクトのセットアップ（10分）

1. **プロジェクトをクローン**
   ```bash
   cd ~/
   git clone https://github.com/higakikeita/tfdrift-falco.git
   cd tfdrift-falco
   ```

2. **設定ファイルを作成**
   ```bash
   # サンプル設定をコピー
   cp examples/config.yaml config.yaml

   # エディタで編集
   vim config.yaml
   ```

3. **config.yaml の編集ポイント**
   ```yaml
   # 1. Falco 接続設定（Docker Compose 使用時）
   falco:
     enabled: true
     hostname: falco  # Docker Compose のサービス名
     port: 5060
     tls: false

   # 2. Terraform State のパス（重要！）
   providers:
     aws:
       enabled: true
       regions:
         - us-east-1
       state:
         backend: local
         local_path: /terraform/terraform.tfstate  # コンテナ内のパス

   # 3. Slack Webhook（オプション）
   notifications:
     slack:
       enabled: true
       webhook_url: "YOUR_WEBHOOK_URL"
       channel: "#alerts"
   ```

4. **環境変数を設定**
   ```bash
   # .env ファイルを作成
   cat > .env << 'EOF'
   # CloudTrail S3 バケット（必須）
   CLOUDTRAIL_S3_BUCKET=your-cloudtrail-bucket-name

   # AWS リージョン
   AWS_REGION=us-east-1

   # Terraform State ディレクトリ（ローカルの場合）
   TERRAFORM_STATE_DIR=/path/to/your/terraform

   # Slack Webhook（オプション）
   SLACK_WEBHOOK_URL=https://hooks.slack.com/services/...
   EOF
   ```

### Phase 3: 起動（5分）

1. **Docker Compose で起動**
   ```bash
   docker-compose up -d
   ```

2. **ログを確認**
   ```bash
   # すべてのログ
   docker-compose logs -f

   # Falco のみ
   docker-compose logs -f falco

   # TFDrift のみ
   docker-compose logs -f tfdrift
   ```

3. **期待される出力**

   **Falco**:
   ```
   Falco initialized with configuration file /etc/falco/falco.yaml
   Loading rules from file /etc/falco/rules.d/terraform_drift.yaml
   gRPC server threadiness equals to 0, enabling single-threaded mode
   Starting gRPC server at 0.0.0.0:5060
   ```

   **TFDrift-Falco**:
   ```
   INFO[2025-12-05] Starting TFDrift-Falco v0.1.0
   INFO[2025-12-05] Connected to Falco gRPC: falco:5060
   INFO[2025-12-05] Loaded Terraform state: 42 resources
   INFO[2025-12-05] Drift detection started
   ```

### Phase 4: 動作確認（10分）

記事の「Phase 4: 動作確認」と同じ

## トラブルシューティング

### 問題1: Falco が CloudTrail S3 バケットにアクセスできない

**エラー**:
```
Error loading CloudTrail events: Access Denied
```

**原因**:
- IAM 権限が不足
- S3 バケット名が間違っている

**対策**:
```bash
# IAM ポリシーを確認
aws iam get-user-policy --user-name $(aws sts get-caller-identity --query 'Arn' --output text | cut -d'/' -f2) --policy-name CloudTrailRead

# 必要な権限
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
    }
  ]
}
```

### 問題2: Terraform State が読み込めない

**エラー**:
```
Failed to load Terraform state: no such file or directory
```

**原因**:
- `TERRAFORM_STATE_DIR` のパスが間違っている
- Docker ボリュームマウントの問題

**対策**:
```bash
# ローカルの Terraform State パスを確認
ls -la /path/to/your/terraform/terraform.tfstate

# docker-compose.yml で正しくマウントされているか確認
docker-compose config | grep -A5 volumes
```

### 問題3: Falco と TFDrift-Falco が接続できない

**エラー**:
```
Failed to connect to Falco gRPC: connection refused
```

**原因**:
- Falco がまだ起動していない
- ネットワーク設定の問題

**対策**:
```bash
# Falco のステータスを確認
docker-compose ps falco

# Falco のヘルスチェックを確認
docker-compose exec falco grpc_health_probe -addr=:5060

# ネットワークを確認
docker network inspect tfdrift-network
```

## 推奨する記事の修正内容

### 1. Phase 1 を簡略化

**修正前**:
- Falco 設定ファイルを手動作成
- Docker で Falco を単独起動

**修正後**:
- CloudTrail S3 バケットの確認
- 環境変数の設定

### 2. Phase 2 を統合

**修正前**:
- 別々のセットアップ手順

**修正後**:
- docker-compose.yml を使用した統合セットアップ

### 3. 環境変数の明確化

**追加すべき情報**:
```bash
# 必須の環境変数
CLOUDTRAIL_S3_BUCKET=xxx
TERRAFORM_STATE_DIR=xxx

# オプション
CLOUDTRAIL_USE_SQS=true
CLOUDTRAIL_SQS_QUEUE=xxx
SLACK_WEBHOOK_URL=xxx
```

### 4. 前提条件の追加

**追加すべき情報**:
- CloudTrail が有効化されていること
- CloudTrail ログが S3 に保存されていること
- IAM 権限が適切に設定されていること

## 結論

**現在の記事では、ユーザーはセットアップに失敗する可能性が高い**

主な理由:
1. `config.example.yaml` が存在しない
2. Falco の CloudTrail 設定が複雑
3. 環境変数の設定が不足
4. docker-compose.yml の構造が記事と異なる

**修正が必要な項目**:
- [ ] Phase 1 を CloudTrail 確認と環境変数設定に変更
- [ ] Phase 2 を docker-compose 統合セットアップに変更
- [ ] 必要な環境変数を明記
- [ ] トラブルシューティングを拡充
