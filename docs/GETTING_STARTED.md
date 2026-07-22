# Getting Started with TFDrift-Falco

このガイドでは、TFDrift-Falcoを**5分でセットアップ**して、すぐにTerraformドリフト検知を開始する方法を説明します。

## 📋 目次

1. [前提条件](#前提条件)
2. [3コマンドでセットアップ（推奨）](#3コマンドでセットアップ推奨)
3. [手動セットアップ（詳細版）](#手動セットアップ詳細版)
4. [動作確認](#動作確認)
5. [トラブルシューティング](#トラブルシューティング)

---

## 前提条件

以下がインストールされていることを確認してください：

### 必須
- **Docker** 20.10+ - [インストールガイド](https://docs.docker.com/get-docker/)
- **Docker Compose** 2.0+ - [インストールガイド](https://docs.docker.com/compose/install/)
- **AWS credentials** - `~/.aws/credentials` に設定済み

### 任意（推奨）
- **make** - Makefileコマンドを使用する場合
- **git** - リポジトリをクローンする場合

### AWS Credentials 確認

```bash
# AWS credentials が設定されているか確認
cat ~/.aws/credentials

# 出力例:
# [default]
# aws_access_key_id = AKIAIOSFODNN7EXAMPLE
# aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

もし設定されていない場合：

```bash
# AWS CLIで設定
aws configure

# または手動で作成
mkdir -p ~/.aws
cat > ~/.aws/credentials <<EOF
[default]
aws_access_key_id = YOUR_ACCESS_KEY
aws_secret_access_key = YOUR_SECRET_KEY
EOF

cat > ~/.aws/config <<EOF
[default]
region = us-east-1
output = json
EOF
```

---

## 3コマンドでセットアップ（推奨）

最も簡単な方法です。5分で完了します。

### ステップ 1: リポジトリをクローン

```bash
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

### ステップ 2: Quick Start スクリプトを実行

```bash
./quick-start.sh
# または
make quick-start
```

**スクリプトが自動的に行うこと：**
- ✅ Docker / Docker Compose のチェック
- ✅ AWS credentials の確認（未設定の場合は対話的に設定）
- ✅ Falco設定ファイルの生成 (`deployments/falco/falco.yaml`)
- ✅ Falcoルールファイルの生成 (`rules/terraform_drift.yaml`)
- ✅ TFDrift-Falco設定ファイルの生成 (`config.yaml`)
- ✅ 対話的な設定（AWS Region、Terraform State Backend、Slack Webhook）

**対話的な質問例：**

```
AWS Region to monitor (default: us-east-1): us-east-1 ⏎

Terraform State Backend:
  1) S3 (recommended for production)
  2) Local file (for testing)
Select backend (1-2, default: 2): 1 ⏎

S3 bucket name: my-terraform-state ⏎
S3 key (e.g., prod/terraform.tfstate): prod/terraform.tfstate ⏎

Slack webhook URL (optional, press Enter to skip): https://hooks.slack.com/... ⏎
```

### ステップ 3: TFDrift-Falcoを起動

```bash
docker compose up -d
# または
make start
```

**起動完了！** 🎉

---

## 手動セットアップ（詳細版）

Quick Startスクリプトを使わず、手動で設定する場合の手順です。

### 1. ディレクトリ構造の準備

```bash
mkdir -p deployments/falco
mkdir -p rules
mkdir -p examples/terraform
```

### 2. Falco設定ファイルを作成

**ファイル**: `deployments/falco/falco.yaml`

```yaml
# Falco configuration for TFDrift-Falco
watch_config_files: true
time_format_iso_8601: true

# Rules
rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml
  - /etc/falco/rules.d

# gRPC output (required for TFDrift-Falco)
grpc:
  enabled: true
  bind_address: "0.0.0.0:5060"
  threadiness: 8

grpc_output:
  enabled: true

# JSON output for CloudTrail events
json_output: true
json_include_output_property: true
json_include_tags_property: true

# Logging
log_stderr: true
log_syslog: false
log_level: info

# CloudTrail plugin configuration
plugins:
  - name: cloudtrail
    library_path: libcloudtrail.so
    init_config:
      s3DownloadConcurrency: 64
      sqsDelete: false
      useAsync: true
    open_params: ""

# Load CloudTrail plugin
load_plugins: [cloudtrail]
```

### 3. TFDrift Falcoルールを作成

**ファイル**: `rules/terraform_drift.yaml`

```yaml
# TFDrift-Falco Rules
- rule: Terraform Managed Resource Modified
  desc: Detect modifications to resources that should be managed by Terraform
  condition: >
    evt.type = aws_api_call and
    ct.name in (
      ModifyInstanceAttribute, ModifyDBInstance, ModifySecurityGroupRules,
      PutBucketPolicy, PutBucketEncryption, UpdateFunctionConfiguration,
      PutRolePolicy, UpdateAssumeRolePolicy, AttachRolePolicy,
      AuthorizeSecurityGroupIngress, RevokeSecurityGroupIngress
    )
  output: >
    Potential Terraform drift detected
    (user=%ct.user event=%ct.name resource=%ct.request.instanceid region=%ct.region)
  priority: WARNING
  tags: [terraform, drift, iac]
  source: aws_cloudtrail

- rule: Critical Infrastructure Change
  desc: Detect critical infrastructure changes that bypass IaC workflows
  condition: >
    evt.type = aws_api_call and
    ct.name in (
      TerminateInstances, DeleteDBInstance, DeleteBucket,
      DeleteSecurityGroup, ScheduleKeyDeletion, DeleteRole
    )
  output: >
    Critical infrastructure deletion detected
    (user=%ct.user event=%ct.name resource=%ct.request.instanceid region=%ct.region)
  priority: CRITICAL
  tags: [terraform, drift, deletion, critical]
  source: aws_cloudtrail
```

### 4. TFDrift-Falco設定ファイルを作成

**ファイル**: `config.yaml`

```yaml
# TFDrift-Falco Configuration

# Cloud Provider Configuration
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"  # または "local"
      s3_bucket: "my-terraform-state"
      s3_key: "prod/terraform.tfstate"

# Falco Integration
falco:
  enabled: true
  hostname: "falco"
  port: 5060

# Drift Detection Rules
drift_rules:
  - name: "EC2 Instance Modification"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
      - "disable_api_termination"
      - "security_groups"
    severity: "high"

  - name: "Security Group Changes"
    resource_types:
      - "aws_security_group"
    watched_attributes:
      - "ingress"
      - "egress"
    severity: "critical"

  - name: "IAM Policy Changes"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
    watched_attributes:
      - "policy"
      - "assume_role_policy"
    severity: "critical"

# Notification Channels
notifications:
  slack:
    enabled: false  # Slackを使う場合はtrueに変更
    webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
    channel: "#tfdrift-alerts"

  falco_output:
    enabled: true
    priority: "warning"

# Logging
logging:
  level: "info"
  format: "json"
```

### 5. Docker Composeで起動

```bash
docker compose up -d
```

---

## 動作確認

### 1. サービスの起動状態を確認

```bash
# 起動状態確認
docker compose ps
# または
make status

# 期待される出力:
# NAME                    STATUS              PORTS
# tfdrift-falco-app       Up 30 seconds       0.0.0.0:8080->8080/tcp
# tfdrift-falco-falco     Up 30 seconds       0.0.0.0:5060->5060/tcp
```

### 2. ログを確認

```bash
# TFDrift-Falcoのログ
docker compose logs -f tfdrift

# Falcoのログ
docker compose logs -f falco

# 両方同時に
docker compose logs -f
# または
make logs
```

**正常起動時のログ例：**

```
tfdrift-falco-app    | INFO  Starting TFDrift-Falco v0.5.0
tfdrift-falco-app    | INFO  Loaded Terraform state: 142 resources
tfdrift-falco-app    | INFO  Connected to Falco gRPC endpoint: falco:5060
tfdrift-falco-app    | INFO  Listening for CloudTrail events...

tfdrift-falco-falco  | Falco initialized with configuration file: /etc/falco/falco.yaml
tfdrift-falco-falco  | Loading plugin cloudtrail from /usr/share/falco/plugins/libcloudtrail.so
tfdrift-falco-falco  | gRPC server listening on 0.0.0.0:5060
```

### 3. ドリフトを手動でテストする

**EC2インスタンスの属性を変更してテスト：**

```bash
# AWS CLIでEC2インスタンスのTermination Protectionを変更
aws ec2 modify-instance-attribute \
  --instance-id i-1234567890abcdef0 \
  --no-disable-api-termination

# TFDrift-Falcoのログを確認
docker compose logs -f tfdrift
```

**期待されるアラート：**

```
tfdrift-falco-app | ALERT Drift Detected!
tfdrift-falco-app | ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
tfdrift-falco-app | Resource:     aws_instance.web_server
tfdrift-falco-app | Type:         Manual Modification
tfdrift-falco-app | Severity:     HIGH
tfdrift-falco-app |
tfdrift-falco-app | Changed Attribute:
tfdrift-falco-app |   disable_api_termination: true → false
tfdrift-falco-app |
tfdrift-falco-app | Context:
tfdrift-falco-app |   User:         admin@example.com
tfdrift-falco-app |   Region:       us-east-1
tfdrift-falco-app |   Timestamp:    2025-12-19T10:35:10Z
```

### 4. Slackで通知を受け取る

Slack Webhookを設定した場合、Slackチャネルにアラートが届きます：

```
🚨 Drift Detected: aws_instance.web_server

Changed: disable_api_termination = true → false

User: admin@example.com
Region: us-east-1
Severity: HIGH

CloudTrail EventID: a1b2c3d4-5678-90ab-cdef-1234567890ab
```

---

## トラブルシューティング

### 問題 1: "Cannot connect to Falco gRPC endpoint"

**症状:**
```
ERROR Failed to connect to Falco gRPC endpoint: connection refused
```

**解決策:**

1. **Falcoが起動しているか確認**
   ```bash
   docker compose ps falco
   ```

2. **Falco gRPCポートが開いているか確認**
   ```bash
   docker compose logs falco | grep "gRPC"
   # 期待される出力: gRPC server listening on 0.0.0.0:5060
   ```

3. **Falcoを再起動**
   ```bash
   docker compose restart falco
   docker compose logs -f falco
   ```

4. **ネットワーク接続を確認**
   ```bash
   docker network inspect tfdrift-network
   ```

### 問題 2: "Terraform state file not found"

**症状:**
```
ERROR Failed to load Terraform state: NoSuchKey
```

**解決策:**

1. **S3バケットとキーを確認**
   ```bash
   aws s3 ls s3://my-terraform-state/prod/terraform.tfstate
   ```

2. **IAM権限を確認**
   ```bash
   aws iam get-user
   # Terraform StateバケットへのGetObject権限が必要
   ```

3. **config.yamlのs3_bucketとs3_keyが正しいか確認**
   ```bash
   cat config.yaml | grep -A 3 "state:"
   ```

4. **ローカルバックエンドに切り替えてテスト**
   ```yaml
   state:
     backend: "local"
     local_path: "/terraform/terraform.tfstate"
   ```

### 問題 3: "Too many drift alerts (False Positives)"

**症状:**
- Slackに大量のアラートが送信される
- 実際にはドリフトではない変更も検知される

**解決策:**

1. **watched_attributesを絞る**
   ```yaml
   drift_rules:
     - name: "EC2 Critical Changes"
       resource_types:
         - "aws_instance"
       watched_attributes:
         - "instance_type"
         - "disable_api_termination"
       # Tags変更は除外
   ```

2. **AWSサービスロールを除外**
   ```yaml
   drift_rules:
     - name: "Manual EC2 Changes"
       resource_types:
         - "aws_instance"
       exclude_users:
         - "AWSServiceRoleForAutoScaling"
         - "AWSServiceRoleForECS"
   ```

3. **重要度をmediumに下げる**
   ```yaml
   drift_rules:
     - name: "Non-Critical Changes"
       severity: "medium"  # critical → medium
   ```

### 問題 4: CloudTrailイベントが届かない

**症状:**
```
INFO  Listening for CloudTrail events...
# その後、何も表示されない
```

**解決策:**

1. **CloudTrailが有効か確認**
   ```bash
   aws cloudtrail describe-trails
   aws cloudtrail get-trail-status --name my-trail
   ```

2. **CloudTrail設定でデータイベントが有効か確認**
   ```bash
   aws cloudtrail get-event-selectors --trail-name my-trail
   # ReadWriteType: All または WriteOnly が必要
   ```

3. **AWS Regionが一致しているか確認**
   ```bash
   # config.yamlのregionsとCloudTrailのregionが一致している必要がある
   cat config.yaml | grep regions
   ```

4. **Falco CloudTrailプラグインのログを確認**
   ```bash
   docker compose logs falco | grep cloudtrail
   ```

### 問題 5: メモリ使用量が高い

**症状:**
- TFDrift-FalcoコンテナがOOMKillerで再起動される
- メモリ使用量が1GB以上

**解決策:**

1. **Terraform State refresh間隔を延長**
   ```yaml
   providers:
     aws:
       state:
         refresh_interval: "15m"  # デフォルト5m → 15m
   ```

2. **Docker Composeのメモリ制限を増やす**
   ```yaml
   tfdrift:
     deploy:
       resources:
         limits:
           memory: 1G  # 512M → 1G
   ```

3. **watched_attributesを減らす**
   ```yaml
   drift_rules:
     - name: "Critical Only"
       watched_attributes:
         - "instance_type"  # 主要な属性のみ
   ```

---

## 次のステップ

セットアップが完了したら、以下のドキュメントを参照してください：

1. **[Use Cases](USE_CASES.md)** - 実際の使用例とシナリオ
2. **[Best Practices](BEST_PRACTICES.md)** - 本番環境での運用ベストプラクティス
3. **[Extending TFDrift-Falco](EXTENDING.md)** - カスタムルール・通知チャネルの追加方法
4. **[Configuration Guide](../README.md#-configuration)** - 詳細な設定オプション

---

## サポート

問題が解決しない場合：

- **GitHub Issues**: https://github.com/higakikeita/tfdrift-falco/issues
- **Discussions**: https://github.com/higakikeita/tfdrift-falco/discussions
- **Slack Community**: [Join Slack](https://join.slack.com/t/tfdrift-falco/...)

---

**ハッピードリフト検知！** 🎉
