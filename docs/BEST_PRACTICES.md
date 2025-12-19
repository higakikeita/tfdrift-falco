# TFDrift-Falco Best Practices

本ドキュメントでは、TFDrift-Falcoを本番環境で運用する際のベストプラクティスを紹介します。

## 📋 目次

1. [Production Deployment](#production-deployment)
2. [Security](#security)
3. [Operational Excellence](#operational-excellence)
4. [Configuration Management](#configuration-management)
5. [Monitoring & Observability](#monitoring--observability)
6. [Performance Tuning](#performance-tuning)
7. [Troubleshooting](#troubleshooting)

---

## Production Deployment

### High Availability Setup

**推奨構成**: Active-Passive with Health Checks

```yaml
# kubernetes deployment example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: tfdrift-falco
spec:
  replicas: 2  # アクティブ-パッシブ構成
  strategy:
    type: Recreate  # 同時実行を防ぐ
  template:
    spec:
      containers:
      - name: tfdrift-falco
        image: ghcr.io/higakikeita/tfdrift-falco:latest
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          exec:
            command: ["/bin/sh", "-c", "pgrep -f tfdrift"]
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          exec:
            command: ["/bin/sh", "-c", "nc -z localhost 5060"]
          initialDelaySeconds: 10
          periodSeconds: 5
```

**なぜActive-Passive?**
- TFDrift-FalcoはFalco gRPCストリームに接続するステートフルなサブスクライバー
- 複数インスタンスが同時にイベントを処理すると、重複通知が発生
- Kubernetes LeaderElectionパターンの使用を推奨（将来のバージョンで実装予定）

### Multi-Region Deployment

**シナリオ**: us-east-1とap-northeast-1を監視

```yaml
# Region 1: us-east-1
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "terraform-state-us-east-1"
      s3_key: "prod/terraform.tfstate"

falco:
  hostname: "falco-us-east-1.internal"
  port: 5060

notifications:
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#drift-us-east-1"
```

```yaml
# Region 2: ap-northeast-1
providers:
  aws:
    enabled: true
    regions:
      - ap-northeast-1
    state:
      backend: "s3"
      s3_bucket: "terraform-state-ap-northeast-1"
      s3_key: "prod/terraform.tfstate"

falco:
  hostname: "falco-ap-northeast-1.internal"
  port: 5060

notifications:
  slack:
    webhook_url: "https://hooks.slack.com/services/..."
    channel: "#drift-ap-northeast-1"
```

**ポイント**:
- リージョンごとに独立したTFDrift-Falcoインスタンスを実行
- CloudTrailログもリージョンごとに処理
- Terraform Stateもリージョンごとに分離

### Resource Sizing

**最小構成** (Small workload, <50 CloudTrail events/min):
```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "256Mi"
    cpu: "250m"
```

**推奨構成** (Medium workload, 50-500 events/min):
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

**Large構成** (Large workload, >500 events/min):
```yaml
resources:
  requests:
    memory: "512Mi"
    cpu: "500m"
  limits:
    memory: "1Gi"
    cpu: "1000m"
```

---

## Security

### IAM Permissions (Principle of Least Privilege)

**最小権限のIAMポリシー例（AWS）**:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "TerraformStateReadOnly",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::my-terraform-state",
        "arn:aws:s3:::my-terraform-state/*"
      ]
    },
    {
      "Sid": "KMSDecryptForState",
      "Effect": "Allow",
      "Action": [
        "kms:Decrypt",
        "kms:DescribeKey"
      ],
      "Resource": "arn:aws:kms:us-east-1:123456789012:key/abc-123-def-456"
    }
  ]
}
```

**GCPのサービスアカウント権限例**:

```bash
# Terraform State読み取り専用
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:tfdrift@my-project.iam.gserviceaccount.com" \
  --role="roles/storage.objectViewer"

# Cloud Auditログ読み取り（Pub/Sub経由）
gcloud projects add-iam-policy-binding my-project \
  --member="serviceAccount:tfdrift@my-project.iam.gserviceaccount.com" \
  --role="roles/pubsub.subscriber"
```

### Network Security

**Falco gRPC接続にmTLS使用**:

```yaml
falco:
  enabled: true
  hostname: "falco.secure.internal"
  port: 5060
  cert_file: "/etc/tfdrift/certs/client.crt"
  key_file: "/etc/tfdrift/certs/client.key"
  ca_root_file: "/etc/tfdrift/certs/ca.crt"
```

**Kubernetesネットワークポリシー**:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: tfdrift-falco-network-policy
spec:
  podSelector:
    matchLabels:
      app: tfdrift-falco
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: prometheus  # メトリクス収集用
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: falco  # Falco gRPCへの接続
    ports:
    - protocol: TCP
      port: 5060
  - to:  # Slack/Webhook通知用
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

### Secrets Management

**❌ 悪い例** - 設定ファイルに平文で記述:

```yaml
notifications:
  slack:
    webhook_url: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX"
```

**✅ 良い例** - Kubernetes Secretsを使用:

```yaml
# config.yaml
notifications:
  slack:
    webhook_url_from_env: "SLACK_WEBHOOK_URL"
```

```bash
# Kubernetes Secret作成
kubectl create secret generic tfdrift-secrets \
  --from-literal=slack-webhook-url="https://hooks.slack.com/services/..."
```

```yaml
# Deployment
env:
- name: SLACK_WEBHOOK_URL
  valueFrom:
    secretKeyRef:
      name: tfdrift-secrets
      key: slack-webhook-url
```

**✅ さらに良い例** - AWS Secrets Manager / GCP Secret Manager:

```yaml
notifications:
  slack:
    webhook_url_from_aws_secret: "prod/tfdrift/slack-webhook"
    # または
    webhook_url_from_gcp_secret: "projects/123/secrets/tfdrift-slack-webhook"
```

---

## Operational Excellence

### Log Retention and Rotation

**推奨**: 構造化ログ（JSON）を外部ロギングシステムに転送

```yaml
# config.yaml
logging:
  level: "info"
  format: "json"  # 本番環境ではJSON推奨
```

**FluentBit統合例**:

```yaml
# fluent-bit-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluent-bit-config
data:
  fluent-bit.conf: |
    [SERVICE]
        Flush        5
        Daemon       Off
        Log_Level    info

    [INPUT]
        Name              tail
        Path              /var/log/tfdrift/*.log
        Parser            json
        Tag               tfdrift.*
        Refresh_Interval  5

    [OUTPUT]
        Name  es
        Match *
        Host  elasticsearch.logging.svc.cluster.local
        Port  9200
        Index tfdrift-logs
```

### Backup and Disaster Recovery

**Terraform State Backup**:

```bash
# 定期的なState Snapshotスクリプト
#!/bin/bash
# backup-terraform-state.sh

DATE=$(date +%Y%m%d-%H%M%S)
BUCKET="my-terraform-state"
KEY="prod/terraform.tfstate"
BACKUP_BUCKET="my-terraform-state-backup"

# S3からダウンロード
aws s3 cp s3://${BUCKET}/${KEY} /tmp/terraform.tfstate

# バックアップバケットにアップロード（バージョニング有効）
aws s3 cp /tmp/terraform.tfstate s3://${BACKUP_BUCKET}/${KEY}.${DATE}

# 90日以上古いバックアップを削除
aws s3 ls s3://${BACKUP_BUCKET}/ | while read -r line; do
  createDate=$(echo $line | awk '{print $1" "$2}')
  createDate=$(date -d "$createDate" +%s)
  olderThan=$(date --date="90 days ago" +%s)
  if [[ $createDate -lt $olderThan ]]; then
    fileName=$(echo $line | awk '{print $4}')
    aws s3 rm s3://${BACKUP_BUCKET}/$fileName
  fi
done
```

**Cron設定** (毎日3AM実行):

```bash
0 3 * * * /usr/local/bin/backup-terraform-state.sh >> /var/log/backup.log 2>&1
```

### Upgrade Procedures

**Zero-Downtime Upgrade (Kubernetes)**:

```bash
# 1. 新しいバージョンをテスト環境で検証
kubectl set image deployment/tfdrift-falco \
  tfdrift-falco=ghcr.io/higakikeita/tfdrift-falco:v0.6.0 \
  -n staging

# 2. テスト環境で動作確認
kubectl logs -f deployment/tfdrift-falco -n staging

# 3. 本番環境にローリングアップデート
kubectl set image deployment/tfdrift-falco \
  tfdrift-falco=ghcr.io/higakikeita/tfdrift-falco:v0.6.0 \
  -n production

# 4. ロールアウト状況を監視
kubectl rollout status deployment/tfdrift-falco -n production

# 5. 問題があればロールバック
kubectl rollout undo deployment/tfdrift-falco -n production
```

---

## Configuration Management

### Drift Rule Design Patterns

**パターン1: Critical Resources Only**

```yaml
drift_rules:
  # IAM関連（最重要）
  - name: "IAM Critical Changes"
    resource_types:
      - "aws_iam_role"
      - "aws_iam_policy"
      - "aws_iam_user"
    watched_attributes:
      - "policy"
      - "assume_role_policy"
      - "inline_policy"
    severity: "critical"

  # セキュリティグループ（重要）
  - name: "Security Group Changes"
    resource_types:
      - "aws_security_group"
      - "aws_security_group_rule"
    watched_attributes:
      - "ingress"
      - "egress"
      - "cidr_blocks"
    severity: "critical"
```

**パターン2: Environment-Specific Rules**

```yaml
drift_rules:
  # 本番環境: 全ての変更を検知
  - name: "Production - All Changes"
    resource_types:
      - "*"  # 全てのリソース
    environment: "production"
    severity: "high"

  # ステージング環境: Critical resourcesのみ
  - name: "Staging - Critical Only"
    resource_types:
      - "aws_iam_*"
      - "aws_security_group*"
      - "aws_kms_*"
    environment: "staging"
    severity: "medium"

  # 開発環境: IAMのみ
  - name: "Dev - IAM Only"
    resource_types:
      - "aws_iam_*"
    environment: "development"
    severity: "low"
```

### Multi-Account Strategy

**アカウント構成例**:
- Production Account (123456789012)
- Staging Account (234567890123)
- Development Account (345678901234)

**推奨デプロイメント**: 各アカウントに個別のTFDrift-Falcoインスタンス

```yaml
# production-config.yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "terraform-state-production"
      s3_key: "terraform.tfstate"

notifications:
  slack:
    webhook_url_from_env: "SLACK_WEBHOOK_PROD"
    channel: "#security-alerts-prod"
```

```yaml
# staging-config.yaml
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: "s3"
      s3_bucket: "terraform-state-staging"
      s3_key: "terraform.tfstate"

notifications:
  slack:
    webhook_url_from_env: "SLACK_WEBHOOK_STAGING"
    channel: "#security-alerts-staging"
```

### Terraform State Backend Setup

**S3バックエンド（推奨設定）**:

```hcl
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true  # 暗号化必須
    kms_key_id     = "arn:aws:kms:us-east-1:123456789012:key/abc-123"
    dynamodb_table = "terraform-state-lock"  # ロック機能

    # バージョニング必須
    versioning     = true
  }
}
```

**S3バケット設定**:

```hcl
resource "aws_s3_bucket" "terraform_state" {
  bucket = "my-terraform-state"

  lifecycle {
    prevent_destroy = true  # 誤削除防止
  }

  tags = {
    Name        = "Terraform State"
    Environment = "Production"
    ManagedBy   = "Terraform"
  }
}

resource "aws_s3_bucket_versioning" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id
  versioning_configuration {
    status = "Enabled"  # バージョニング有効化
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm     = "aws:kms"
      kms_master_key_id = aws_kms_key.terraform_state.arn
    }
  }
}

resource "aws_s3_bucket_public_access_block" "terraform_state" {
  bucket = aws_s3_bucket.terraform_state.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}
```

---

## Monitoring & Observability

### Prometheus Metrics

TFDrift-Falcoは以下のメトリクスを公開します（`/metrics` エンドポイント）:

```prometheus
# 検知したドリフトイベント数
tfdrift_events_total{severity="critical"} 5
tfdrift_events_total{severity="high"} 23

# リソースタイプ別イベント数
tfdrift_events_by_type{type="aws_instance"} 12
tfdrift_events_by_type{type="aws_iam_role"} 8

# 検知レイテンシー（秒）
tfdrift_detection_latency_seconds{quantile="0.5"} 0.8
tfdrift_detection_latency_seconds{quantile="0.95"} 2.3
tfdrift_detection_latency_seconds{quantile="0.99"} 5.1

# Falco接続状態
tfdrift_falco_connected 1

# Terraform State同期時刻（UnixTimestamp）
tfdrift_state_last_sync_timestamp 1705312345
```

### Grafana Alerting

**アラートルール例**:

```yaml
# grafana-alerts.yaml
groups:
  - name: tfdrift-alerts
    interval: 1m
    rules:
      # Criticalドリフトが発生したら即座に通知
      - alert: CriticalDriftDetected
        expr: increase(tfdrift_events_total{severity="critical"}[5m]) > 0
        for: 0m
        labels:
          severity: critical
        annotations:
          summary: "Critical drift detected in {{ $labels.environment }}"
          description: "{{ $value }} critical drift events detected in the last 5 minutes"

      # 検知レイテンシーが10秒を超えたら警告
      - alert: HighDetectionLatency
        expr: tfdrift_detection_latency_seconds{quantile="0.95"} > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High drift detection latency"
          description: "P95 latency is {{ $value }}s (threshold: 10s)"

      # Falco接続が切れたら即座にアラート
      - alert: FalcoConnectionLost
        expr: tfdrift_falco_connected == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Lost connection to Falco"
          description: "TFDrift-Falco cannot connect to Falco gRPC endpoint"

      # State同期が30分以上行われていない場合
      - alert: TerraformStateStale
        expr: (time() - tfdrift_state_last_sync_timestamp) > 1800
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Terraform state not synced for 30+ minutes"
          description: "Last sync: {{ $value | humanizeDuration }}"
```

### Health Checks

**Kubernetes Liveness & Readiness Probes**:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

**ヘルスチェックエンドポイント**:

- `GET /healthz` - 基本的なヘルスチェック（プロセスが生きているか）
- `GET /ready` - 準備状態チェック（Falco接続、State読み込み完了）
- `GET /metrics` - Prometheusメトリクス

---

## Performance Tuning

### Terraform State Refresh Interval

**デフォルト**: 5分ごと

```yaml
providers:
  aws:
    state:
      backend: "s3"
      s3_bucket: "my-terraform-state"
      s3_key: "terraform.tfstate"
      refresh_interval: "5m"  # デフォルト
```

**推奨設定**:
- **Small environments (<100 resources)**: `refresh_interval: "5m"`
- **Medium environments (100-500 resources)**: `refresh_interval: "10m"`
- **Large environments (>500 resources)**: `refresh_interval: "15m"`

**理由**: State読み込みはS3 API呼び出しが発生し、大規模環境ではオーバーヘッドになる

### Event Filtering

**不要なイベントをフィルタリング**:

```yaml
# Falcoルールで事前フィルタリング
- rule: Terraform Relevant CloudTrail Events
  desc: Only process events relevant to Terraform drift detection
  condition: >
    ct.name in (RunInstances, TerminateInstances, ModifyInstanceAttribute, ...)
    and not ct.user startswith "AWSServiceRole"
  output: "Terraform-relevant event detected"
  priority: INFO
```

**TFDrift-Falco側でもフィルタリング**:

```yaml
drift_rules:
  - name: "EC2 Specific Attributes"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"
      - "disable_api_termination"
      # 不要な属性は監視しない（例: tags.Name の変更は無視）
    severity: "high"
```

### Concurrency Settings

```yaml
# config.yaml
performance:
  # 並列処理するイベント数（デフォルト: 10）
  event_worker_count: 10

  # Terraform State解析の並列度（デフォルト: 5）
  state_parser_workers: 5

  # 通知送信の並列度（デフォルト: 3）
  notifier_workers: 3
```

**推奨設定**:
- **Small workload (<50 events/min)**: `event_worker_count: 5`
- **Medium workload (50-200 events/min)**: `event_worker_count: 10`
- **Large workload (>200 events/min)**: `event_worker_count: 20`

---

## Troubleshooting

### Common Issues

#### Issue 1: "Cannot connect to Falco gRPC endpoint"

**症状**:
```
ERROR Failed to connect to Falco gRPC endpoint: connection refused
```

**原因と解決策**:

1. **Falcoが起動していない**
   ```bash
   # Falco起動確認
   systemctl status falco

   # 起動していない場合
   systemctl start falco
   ```

2. **gRPCが有効化されていない**
   ```yaml
   # /etc/falco/falco.yaml
   grpc:
     enabled: true
     bind_address: "0.0.0.0:5060"
     threadiness: 8
   ```

3. **ネットワーク接続問題**
   ```bash
   # Falcoエンドポイントに接続確認
   nc -zv localhost 5060

   # Kubernetesの場合、ServiceのDNS名を確認
   nslookup falco.default.svc.cluster.local
   ```

#### Issue 2: "Terraform state file not found"

**症状**:
```
ERROR Failed to load Terraform state: NoSuchKey
```

**原因と解決策**:

1. **S3バケット/キーが間違っている**
   ```bash
   # S3にファイルが存在するか確認
   aws s3 ls s3://my-terraform-state/prod/terraform.tfstate
   ```

2. **IAM権限不足**
   ```bash
   # IAMポリシーを確認
   aws iam get-policy-version --policy-arn arn:aws:iam::123456789012:policy/TFDriftPolicy --version-id v1
   ```

3. **KMS復号化権限不足**
   ```bash
   # KMSキーへのアクセス確認
   aws kms describe-key --key-id arn:aws:kms:us-east-1:123456789012:key/abc-123
   ```

#### Issue 3: "Too many drift alerts (False Positives)"

**症状**:
- Slackに大量のアラートが送信される
- 実際にはドリフトではない変更も検知される

**原因と解決策**:

1. **watched_attributesが広すぎる**
   ```yaml
   # ❌ 悪い例
   drift_rules:
     - name: "All EC2 Changes"
       resource_types:
         - "aws_instance"
       watched_attributes:
         - "*"  # 全ての属性を監視（Tagsの変更も含む）

   # ✅ 良い例
   drift_rules:
     - name: "Critical EC2 Changes"
       resource_types:
         - "aws_instance"
       watched_attributes:
         - "instance_type"
         - "disable_api_termination"
         - "security_groups"
       # Tags変更は除外
   ```

2. **Terraform管理外のリソースを検知している**
   ```yaml
   # TerraformのStateに含まれるリソースのみ監視
   drift_rules:
     - name: "IAM Changes"
       resource_types:
         - "aws_iam_role"
       terraform_managed_only: true  # Terraform管理リソースのみ
   ```

3. **Auto Scalingによる自動変更を検知している**
   ```yaml
   # ユーザーによる変更のみ検知（AWSサービスロールを除外）
   drift_rules:
     - name: "Manual EC2 Changes"
       resource_types:
         - "aws_instance"
       exclude_users:
         - "AWSServiceRoleForAutoScaling"
         - "AWSServiceRoleForECS"
   ```

#### Issue 4: "High memory usage"

**症状**:
- メモリ使用量が1GB以上
- OOMKillerによるPod再起動

**原因と解決策**:

1. **Terraform Stateが巨大（1000+ resources）**
   ```yaml
   # State読み込み頻度を下げる
   providers:
     aws:
       state:
         refresh_interval: "15m"  # 5m → 15m
   ```

2. **イベントキューが溜まっている**
   ```yaml
   # Worker数を増やす
   performance:
     event_worker_count: 20  # 10 → 20
   ```

3. **リソース制限を増やす**
   ```yaml
   resources:
     limits:
       memory: "1Gi"  # 512Mi → 1Gi
       cpu: "1000m"
   ```

#### Issue 5: "Detection latency is high (>10 seconds)"

**症状**:
- CloudTrailイベント発生からアラート送信まで10秒以上かかる

**原因と解決策**:

1. **Terraform State読み込みが遅い**
   ```bash
   # State読み込み時間を計測
   time aws s3 cp s3://my-terraform-state/terraform.tfstate - | wc -l
   ```

   → S3をVPCエンドポイント経由で接続（レイテンシー削減）

2. **Diff計算が重い**
   ```yaml
   # watched_attributesを減らす
   drift_rules:
     - name: "IAM Changes"
       resource_types:
         - "aws_iam_role"
       watched_attributes:
         - "assume_role_policy"  # 主要な属性のみ
   ```

3. **通知送信が遅い（Webhook timeout）**
   ```yaml
   notifications:
     webhook:
       timeout: 5s  # 10s → 5s
       max_retries: 2  # 5 → 2
   ```

### Debug Logging

**デバッグログの有効化**:

```yaml
# config.yaml
logging:
  level: "debug"  # info → debug
  format: "text"  # 人間が読みやすい形式
```

**特定のコンポーネントのみデバッグ**:

```bash
# 環境変数で制御
export TFDRIFT_LOG_LEVEL=debug
export TFDRIFT_LOG_COMPONENTS="falco,detector"  # FalcoとDetectorのみ

tfdrift --config config.yaml
```

### Performance Profiling

**CPU Profiling**:

```bash
# pprofを有効化
tfdrift --config config.yaml --pprof-port 6060 &

# プロファイリング実行（30秒間）
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

**Memory Profiling**:

```bash
# Heap profileを取得
curl -o heap.out http://localhost:6060/debug/pprof/heap

# 解析
go tool pprof heap.out
```

---

## Appendix

### Checklist: Production Readiness

本番環境にデプロイする前に、以下の項目を確認してください:

- [ ] **High Availability**: 2つ以上のレプリカでデプロイ（Active-Passive構成）
- [ ] **Resource Limits**: CPUとメモリのlimits/requestsを設定
- [ ] **IAM Permissions**: 最小権限の原則に従ったIAMポリシー
- [ ] **Secrets Management**: Webhook URLや認証情報をSecrets/Secret Managerで管理
- [ ] **Network Security**: mTLS有効化、Network Policy設定
- [ ] **Monitoring**: Prometheusメトリクス収集、Grafanaダッシュボード構築
- [ ] **Alerting**: Critical/Highレベルのアラートルール設定
- [ ] **Logging**: JSON形式でログを外部システムに転送
- [ ] **Backup**: Terraform Stateの定期バックアップ
- [ ] **Testing**: ステージング環境で動作確認
- [ ] **Drift Rules**: 環境に適したルール設定（False Positive最小化）
- [ ] **Documentation**: 運用手順書、トラブルシューティングガイド作成

### References

- [AWS CloudTrail User Guide](https://docs.aws.amazon.com/cloudtrail/)
- [Falco Documentation](https://falco.org/docs/)
- [Terraform Backend Configuration](https://developer.hashicorp.com/terraform/language/settings/backends/s3)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [Prometheus Best Practices](https://prometheus.io/docs/practices/)

---

**次のステップ**: [Extending TFDrift-Falco](EXTENDING.md) でカスタムルールや通知チャネルの追加方法を学びましょう。
