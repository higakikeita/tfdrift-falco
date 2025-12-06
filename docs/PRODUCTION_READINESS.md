# 本番環境導入チェックリスト

**対象**: TFDrift-Falco を本番環境に導入する前に確認すべき事項
**目的**: 開発環境でのテスト成功 ≠ 本番環境での信頼性の保証

---

## ⚠️ 重要な前提

このドキュメントは、**TFDrift-Falco を本番環境に導入する際の制限事項と検証すべき事項**を明確にするためのものです。

**開発環境での成功は本番環境での成功を保証しません。**

- ✅ 開発環境: サンプルデータ、小規模、単一リージョン、Docker
- ❌ 本番環境: 大量イベント、複数アカウント、複数リージョン、高可用性要件

---

## 1. 既知の制限事項

### 1.1 スケーラビリティ

#### ⚠️ 大規模環境での未検証

**現状**:
- 統合テストは**小規模なサンプルデータ**で実施
- 実際の本番環境での負荷テストは**未実施**

**本番環境で想定される規模**:
```
小規模環境:
- CloudTrail イベント: ~100/分
- Terraform リソース: ~500 個
- リージョン: 1-2 個

中規模環境:
- CloudTrail イベント: ~1,000/分
- Terraform リソース: ~5,000 個
- リージョン: 3-5 個

大規模環境:
- CloudTrail イベント: ~10,000/分 以上
- Terraform リソース: ~50,000 個 以上
- リージョン: 10+ 個
- 複数 AWS アカウント
```

**潜在的な問題**:
- [ ] イベント処理の遅延
- [ ] メモリ使用量の増大
- [ ] Terraform State の読み込み時間
- [ ] Grafana クエリのタイムアウト

**推奨事項**:
1. **段階的な導入**
   - まず単一リージョン・単一アカウントで開始
   - 1週間監視し、メトリクスを収集
   - 問題なければ徐々に拡大

2. **負荷テストの実施**
   ```bash
   # ベンチマークテストを実行
   cd tests/benchmark
   go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

   # プロファイル分析
   go tool pprof cpu.prof
   go tool pprof mem.prof
   ```

3. **リソース制限の設定**
   ```yaml
   # docker-compose.yml
   services:
     tfdrift:
       deploy:
         resources:
           limits:
             cpus: '2'
             memory: 2G
           reservations:
             cpus: '1'
             memory: 1G
   ```

### 1.2 CloudTrail 遅延

#### ⚠️ CloudTrail のログ配信遅延

**問題**:
- CloudTrail ログは S3 に配信されるまで**5-15分の遅延**がある
- SQS 通知を使用しても**1-5分の遅延**がある

**影響**:
- リアルタイム検知ではなく、**準リアルタイム**
- 重大なセキュリティインシデントの即座の検知には不向き

**対策**:
1. **SQS 通知を使用** (推奨)
   ```yaml
   # config.yaml
   providers:
     aws:
       cloudtrail:
         sqs_queue: "arn:aws:sqs:us-east-1:123456789012:cloudtrail-events"
   ```
   - 遅延を 1-5分に短縮

2. **GuardDuty や Security Hub との併用**
   - より即座の脅威検知には GuardDuty
   - TFDrift-Falco は Drift 検知に特化

3. **期待値の調整**
   - "リアルタイム" ではなく "準リアルタイム" と認識
   - SLA を適切に設定 (例: 5分以内に検知)

### 1.3 AWS リソースカバレッジ

#### ⚠️ 全 AWS サービスをカバーしていない

**現在の対応状況**:
- **対応済み**: EC2 (部分)、IAM (充実)、S3 (部分)、RDS (部分)、Lambda (部分)
- **未対応**: VPC、ELB/ALB、KMS、DynamoDB、ECS/EKS、その他多数

**詳細**: `AWS_RESOURCE_COVERAGE_ANALYSIS.md` を参照

**影響**:
- 未対応サービスの変更は**検知されない**
- 例: セキュリティグループの変更 (現在未対応) → **検知漏れ**

**対策**:
1. **カバレッジドキュメントを確認**
   ```bash
   cat docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md
   ```

2. **自組織で使用するサービスを特定**
   ```bash
   # Terraform State から使用中のリソースタイプを抽出
   terraform show -json | jq -r '.values.root_module.resources[].type' | sort -u
   ```

3. **カバレッジギャップを特定**
   - 使用中だが未対応のサービスをリスト化
   - 重要なサービスの追加実装を検討

4. **代替手段の検討**
   - 未対応サービスは AWS Config Rules や GuardDuty で補完

### 1.4 マルチアカウント・マルチリージョン

#### ⚠️ マルチアカウント環境での未検証

**現状**:
- 単一アカウント・単一リージョンでのテストのみ

**課題**:
1. **CloudTrail の組織トレイル**
   - AWS Organizations の組織トレイルに対応しているか未検証

2. **Terraform State の管理**
   - 複数アカウントの State をどう管理するか
   - S3 バックエンドでアカウントごとに異なるバケット

3. **リージョンごとのリソース重複**
   - グローバルリソース (IAM, Route53) の重複排除

**推奨事項**:
1. **アカウントごとに TFDrift-Falco インスタンスを起動**
   ```
   Account A: TFDrift-Falco Instance A
   Account B: TFDrift-Falco Instance B
   Account C: TFDrift-Falco Instance C
   ```

2. **Grafana で集約**
   - 各インスタンスから Grafana Loki にログを集約
   - アカウント情報をラベルに追加

3. **組織トレイルの場合**
   - メンバーアカウントの CloudTrail ログもマスターアカウントの S3 に保存される
   - TFDrift-Falco をマスターアカウントで起動し、全アカウントを監視可能
   - ただし、各アカウントの Terraform State へのアクセス権限が必要

### 1.5 Terraform State の同期

#### ⚠️ State の更新タイミング

**問題**:
- Terraform State は `terraform apply` 時に更新される
- State 更新前の CloudTrail イベントは**誤検知の可能性**

**シナリオ**:
```
1. 12:00:00 - terraform apply 開始
2. 12:00:30 - AWS リソースが変更される (CloudTrail イベント発生)
3. 12:01:00 - terraform apply 完了、State 更新
4. 12:01:30 - TFDrift-Falco が State をリロード
5. 12:02:00 - CloudTrail ログが S3 に配信 (5分遅延)
6. 12:07:00 - TFDrift-Falco がイベントを受信し、Drift 判定
```

**問題点**:
- 12:00:30 のイベントが 12:07:00 に処理される
- その時点で State は既に更新済み → **Drift なしと判定**
- つまり、**正常な terraform apply による変更は検知されない** ✅

**別のシナリオ (誤検知)**:
```
1. 12:00:00 - 手動で AWS コンソールからリソース変更
2. 12:00:30 - CloudTrail イベント発生
3. 12:05:00 - CloudTrail ログが S3 に配信
4. 12:06:00 - TFDrift-Falco がイベント受信し、Drift 検知 ✅
5. 12:10:00 - 管理者が terraform import でリソースを State に追加
6. 12:11:00 - TFDrift-Falco が State をリロード
7. 12:15:00 - (新しいイベントがないため、Drift アラートは継続)
```

**対策**:
1. **State 更新頻度の設定**
   ```yaml
   # config.yaml
   advanced:
     state_refresh_interval: "5m"  # デフォルト
   ```
   - より頻繁に更新すると誤検知が減る
   - ただし State 読み込みの負荷が増大

2. **terraform apply のフック**
   - CI/CD で `terraform apply` 後に TFDrift-Falco に State リロードを指示
   ```bash
   terraform apply
   curl -X POST http://tfdrift-api:8080/api/v1/reload-state
   ```
   - ただし、現在 TFDrift-Falco に API エンドポイントはない (要実装)

3. **誤検知の許容**
   - 完全な同期は困難
   - 一定の誤検知は許容し、手動で確認

### 1.6 Grafana アラートの自動化

#### ⚠️ YAML ベースのアラート設定が動作しない

**問題**:
- Grafana 10.x 以降、YAML ベースのアラートプロビジョニングが**動作しない**
- `dashboards/grafana/provisioning/alerting/alerts.yaml` は参考用のみ

**影響**:
- アラート設定を手動で UI から実施する必要がある
- Infrastructure as Code (IaC) での管理が困難

**対策**:
1. **UI での手動設定** (現状の推奨)
   - `dashboards/grafana/ALERTS.md` の手順に従う
   - スクリーンショット付きガイド

2. **Terraform Provider for Grafana** (将来的な解決策)
   ```hcl
   resource "grafana_rule_group" "tfdrift_alerts" {
     name             = "TFDrift Alerts"
     folder_uid       = grafana_folder.tfdrift.uid
     interval_seconds = 60

     rule {
       name      = "Critical Drift Detected"
       condition = "C"
       # ...
     }
   }
   ```
   - ただし、設定が複雑

3. **Grafana HTTP API** (スクリプトで自動化)
   ```bash
   curl -X POST http://grafana:3000/api/ruler/grafana/api/v1/rules/tfdrift \
     -H "Authorization: Bearer $GRAFANA_API_KEY" \
     -d @alerts.json
   ```

### 1.7 セキュリティ・アクセス制御

#### ⚠️ ログの改竄・漏洩リスク

**潜在的なリスク**:
1. **Falco → TFDrift-Falco 間の通信**
   - デフォルトは TLS なし
   - 中間者攻撃のリスク

2. **TFDrift-Falco → Loki 間の通信**
   - 認証なし
   - ログの改竄リスク

3. **Grafana へのアクセス**
   - デフォルトは `admin/admin`
   - 不正アクセスのリスク

4. **CloudTrail ログへのアクセス**
   - TFDrift-Falco に S3 読み取り権限が必要
   - 過大な権限付与のリスク

**対策**:

1. **Falco gRPC の mTLS 有効化**
   ```yaml
   # config.yaml
   falco:
     enabled: true
     hostname: "falco"
     port: 5061
     cert_file: "/certs/client-cert.pem"
     key_file: "/certs/client-key.pem"
     ca_root_file: "/certs/ca-root.pem"
   ```
   - `deployments/falco/certs/` に証明書を配置

2. **Loki の認証有効化**
   ```yaml
   # loki-config.yaml
   auth_enabled: true
   ```
   - Basic Auth または OAuth2

3. **Grafana のセキュリティ強化**
   ```yaml
   # grafana.ini
   [security]
   admin_user = admin
   admin_password = <strong-password>

   [auth.anonymous]
   enabled = false

   [auth.basic]
   enabled = true
   ```

4. **IAM 権限の最小化**
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
           "arn:aws:s3:::my-cloudtrail-bucket/*",
           "arn:aws:s3:::my-cloudtrail-bucket"
         ]
       },
       {
         "Effect": "Allow",
         "Action": [
           "s3:GetObject"
         ],
         "Resource": [
           "arn:aws:s3:::my-terraform-state-bucket/prod/terraform.tfstate"
         ]
       },
       {
         "Effect": "Allow",
         "Action": [
           "sqs:ReceiveMessage",
           "sqs:DeleteMessage",
           "sqs:GetQueueAttributes"
         ],
         "Resource": [
           "arn:aws:sqs:us-east-1:123456789012:cloudtrail-events"
         ]
       }
     ]
   }
   ```

5. **ログ保持ポリシー**
   ```yaml
   # loki-config.yaml
   limits_config:
     retention_period: 30d  # 30日後に自動削除
   ```

6. **ネットワークセグメンテーション**
   - TFDrift-Falco を private subnet に配置
   - Security Group で最小限のアクセスのみ許可

---

## 2. 本番導入前の検証チェックリスト

### 2.1 機能検証

- [ ] **使用中の AWS サービスのカバレッジ確認**
  ```bash
  # 使用中のリソースタイプを抽出
  terraform show -json | jq -r '.values.root_module.resources[].type' | sort -u > /tmp/used_resources.txt

  # 対応状況を確認
  cat docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md | grep -f /tmp/used_resources.txt
  ```

- [ ] **サンプルイベントでの動作確認**
  ```bash
  # EC2 インスタンス属性変更をシミュレート
  aws ec2 modify-instance-attribute --instance-id i-xxxxx --no-disable-api-termination

  # 5-10分待機 (CloudTrail 遅延)

  # Grafana でイベントを確認
  # ログクエリ: {job="tfdrift-falco"} | json | resource_type="aws_instance"
  ```

- [ ] **アラート通知の確認**
  - Slack 通知が届くか
  - 通知内容が適切か
  - 誤検知がないか

### 2.2 性能検証

- [ ] **メモリ使用量の監視**
  ```bash
  docker stats tfdrift
  ```
  - 1時間、1日、1週間の推移を記録

- [ ] **CPU 使用率の監視**
  - 通常時: 5-10% 以下
  - イベント処理時: 30% 以下

- [ ] **イベント処理遅延の測定**
  ```bash
  # CloudTrail イベント発生時刻と TFDrift-Falco 処理時刻の差
  # Grafana で測定:
  # (now() - timestamp(event))
  ```

- [ ] **Terraform State 読み込み時間**
  ```bash
  # ログから測定
  docker logs tfdrift | grep "Loaded Terraform state"
  ```
  - 5秒以内が理想
  - 30秒以上の場合は State の分割を検討

### 2.3 セキュリティ検証

- [ ] **IAM 権限の最小化確認**
  ```bash
  # 実際に使用されている権限をログから確認
  aws cloudtrail lookup-events \
    --lookup-attributes AttributeKey=Username,AttributeValue=tfdrift-role \
    --max-results 100
  ```

- [ ] **TLS/mTLS の有効化確認**
  ```bash
  # Falco gRPC の TLS 確認
  openssl s_client -connect falco:5061 -CAfile /certs/ca-root.pem
  ```

- [ ] **Grafana アクセス制御の確認**
  - デフォルトパスワードが変更されているか
  - 匿名アクセスが無効化されているか
  - RBAC (Role-Based Access Control) が設定されているか

- [ ] **ログの暗号化**
  - Loki の保存先 (S3) が暗号化されているか
  - 転送中の暗号化 (TLS) が有効か

### 2.4 可用性検証

- [ ] **Falco 停止時の挙動**
  ```bash
  docker stop falco
  # TFDrift-Falco のログを確認
  docker logs -f tfdrift
  # 期待: 再接続を試行、エラーログ出力
  ```

- [ ] **CloudTrail 遅延時の挙動**
  - 15分以上イベントがない場合の動作

- [ ] **Terraform State 読み込み失敗時**
  ```bash
  # State ファイルを一時的に削除
  mv terraform.tfstate terraform.tfstate.bak
  # TFDrift-Falco のログを確認
  docker logs tfdrift
  ```

- [ ] **メモリ不足時の挙動**
  ```bash
  # メモリ制限を設定
  docker update --memory 512m tfdrift
  # 大量イベントを送信
  ```

### 2.5 運用検証

- [ ] **ログローテーション設定**
  ```yaml
  # docker-compose.yml
  services:
    tfdrift:
      logging:
        driver: "json-file"
        options:
          max-size: "100m"
          max-file: "3"
  ```

- [ ] **バックアップ・リストア手順**
  - Grafana ダッシュボード設定のエクスポート
  - Loki ログのバックアップ
  - 設定ファイルのバージョン管理

- [ ] **アラート対応手順 (Runbook)**
  - アラートが発生した際の対応手順を文書化
  - エスカレーションフロー

- [ ] **メトリクス収集**
  ```bash
  # Prometheus メトリクスの確認
  curl http://localhost:9090/metrics | grep tfdrift
  ```

---

## 3. アラート閾値のチューニングガイド

### 3.1 組織ごとの閾値調整が必須

**デフォルト値は参考値**:
- 組織の規模、変更頻度、リスク許容度によって最適な閾値は異なる

**チューニングプロセス**:

1. **ベースライン測定 (1-2週間)**
   ```
   目的: 通常時のイベント発生頻度を把握

   測定項目:
   - 1日あたりの Drift イベント数
   - 時間帯別の傾向
   - サービス別の傾向
   - 重要度別の分布
   ```

2. **閾値の仮設定**
   ```
   Critical アラート: ベースライン平均 + 3σ (標準偏差)
   High アラート:     ベースライン平均 + 2σ
   Medium アラート:   ベースライン平均 + 1σ
   ```

3. **1週間の試行**
   - 誤検知率を測定
   - 見逃し (False Negative) がないか確認

4. **閾値の調整**
   - 誤検知が多い → 閾値を上げる
   - 見逃しが多い → 閾値を下げる

### 3.2 閾値設定例

#### 小規模組織 (月間 terraform apply 10回程度)

```yaml
# Grafana アラートルール

# Critical Drift (即座に対応が必要)
- name: Critical Drift Detected
  expr: 'count_over_time({job="tfdrift-falco"} | json | severity="critical" [5m]) > 0'
  # 閾値: 1件でもアラート (厳しめ)

# High Severity Drift (30分以内に対応)
- name: High Severity Drift
  expr: 'count_over_time({job="tfdrift-falco"} | json | severity="high" [10m]) > 2'
  # 閾値: 10分で3件以上

# Excessive Drift Rate (異常な変更頻度)
- name: Excessive Drift Rate
  expr: 'count_over_time({job="tfdrift-falco"} | json [1h]) > 10'
  # 閾値: 1時間で10件以上
```

#### 大規模組織 (月間 terraform apply 100回以上)

```yaml
# Critical Drift
- name: Critical Drift Detected
  expr: 'count_over_time({job="tfdrift-falco"} | json | severity="critical" [10m]) > 3'
  # 閾値: 10分で4件以上 (緩め)

# High Severity Drift
- name: High Severity Drift
  expr: 'count_over_time({job="tfdrift-falco"} | json | severity="high" [30m]) > 10'
  # 閾値: 30分で11件以上

# Excessive Drift Rate
- name: Excessive Drift Rate
  expr: 'count_over_time({job="tfdrift-falco"} | json [1h]) > 50'
  # 閾値: 1時間で50件以上
```

### 3.3 メンテナンスウィンドウの設定

**問題**:
- 計画的なメンテナンス時にアラートが大量発生

**対策**:
```yaml
# config.yaml
advanced:
  maintenance_windows:
    - name: "Weekly Deployment"
      days: ["Sunday"]
      start_time: "02:00"
      end_time: "06:00"
      timezone: "Asia/Tokyo"

    - name: "Emergency Maintenance"
      start_datetime: "2025-12-10T20:00:00Z"
      end_datetime: "2025-12-10T23:00:00Z"
```

**Grafana での実装**:
- Silence (ミュート) 機能を使用
- スケジュールされた Silence を事前に設定

---

## 4. 本番環境アーキテクチャ推奨構成

### 4.1 小規模環境 (単一アカウント・リージョン)

```
┌─────────────────────────────────────────┐
│ AWS Account (Production)               │
│                                          │
│  ┌─────────────┐      ┌──────────────┐ │
│  │ CloudTrail  │─────>│ S3 Bucket    │ │
│  └─────────────┘      │ (Logs)       │ │
│                        └──────────────┘ │
│                               │          │
│                               v          │
│  ┌────────────────────────────────────┐ │
│  │ EC2 Instance / ECS Task            │ │
│  │                                    │ │
│  │  ┌──────────┐    ┌─────────────┐ │ │
│  │  │  Falco   │───>│ TFDrift-    │ │ │
│  │  │ (Docker) │    │ Falco       │ │ │
│  │  └──────────┘    └─────────────┘ │ │
│  │                         │          │ │
│  │                         v          │ │
│  │  ┌─────────────────────────────┐  │ │
│  │  │ Grafana Stack               │  │ │
│  │  │ - Loki                      │  │ │
│  │  │ - Promtail                  │  │ │
│  │  │ - Grafana                   │  │ │
│  │  └─────────────────────────────┘  │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```

**特徴**:
- シンプルな構成
- 単一インスタンスで完結
- コスト効率的

**制限**:
- 単一障害点 (SPOF)
- スケールアウト不可

### 4.2 中規模環境 (複数リージョン・高可用性)

```
┌─────────────────────────────────────────────────────────┐
│ AWS Account (Production)                                │
│                                                          │
│  ┌──────────────┐        ┌──────────────┐              │
│  │ CloudTrail   │───────>│ S3 Bucket    │              │
│  │ (All Regions)│        │ (us-east-1)  │              │
│  └──────────────┘        └──────────────┘              │
│                                 │                        │
│                   ┌─────────────┴───────────────┐       │
│                   │                             │       │
│                   v                             v       │
│  ┌────────────────────────────┐  ┌────────────────────────────┐
│  │ us-east-1                  │  │ us-west-2 (DR)             │
│  │ ┌───────────────────────┐  │  │ ┌───────────────────────┐  │
│  │ │ ECS Fargate Cluster   │  │  │ │ ECS Fargate Cluster   │  │
│  │ │                       │  │  │ │ (Standby)             │  │
│  │ │ - Falco Task (x2)     │  │  │ │ - Falco Task (x1)     │  │
│  │ │ - TFDrift Task (x2)   │  │  │ │ - TFDrift Task (x1)   │  │
│  │ └───────────────────────┘  │  │ └───────────────────────┘  │
│  │           │                │  │           │                │
│  │           v                │  │           v                │
│  │ ┌───────────────────────┐  │  │ ┌───────────────────────┐  │
│  │ │ Grafana Cloud Loki    │  │  │ │ Grafana Cloud Loki    │  │
│  │ │ (Multi-tenant)        │  │  │ │ (Same endpoint)       │  │
│  │ └───────────────────────┘  │  │ └───────────────────────┘  │
│  └────────────────────────────┘  └────────────────────────────┘
│                                                          │
│  ┌────────────────────────────────────────────────────┐ │
│  │ Grafana Cloud (SaaS)                               │ │
│  │ - Centralized Dashboards                           │ │
│  │ - Alerting                                         │ │
│  └────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

**特徴**:
- 高可用性 (Multi-AZ / Multi-Region)
- ECS Fargate でサーバーレス運用
- Grafana Cloud でマネージド Loki

**利点**:
- SPOF なし
- 自動スケーリング
- 運用負荷低減

### 4.3 大規模環境 (複数アカウント・Organizations)

```
┌─────────────────────────────────────────────────────────────┐
│ AWS Organizations                                           │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Management Account                                   │  │
│  │ ┌─────────────────────────────────────────────────┐  │  │
│  │ │ Organization CloudTrail                          │  │  │
│  │ │ (All Accounts, All Regions)                      │  │  │
│  │ └──────────────────┬───────────────────────────────┘  │  │
│  └────────────────────┼──────────────────────────────────┘  │
│                       │                                      │
│                       v                                      │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ Centralized S3 Bucket (Logs)                         │  │
│  │ arn:aws:s3:::org-cloudtrail-logs                     │  │
│  └──────────────────┬───────────────────────────────────┘  │
│                     │                                        │
│      ┌──────────────┼──────────────┬──────────────┐         │
│      │              │               │              │         │
│      v              v               v              v         │
│  ┌────────┐    ┌────────┐     ┌────────┐    ┌────────┐    │
│  │Account │    │Account │     │Account │    │Account │    │
│  │A (Dev) │    │B (Stg) │     │C (Prod)│    │D (Sec) │    │
│  └────────┘    └────────┘     └────────┘    └────┬───┘    │
│                                                    │         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Security Account (Monitoring)                       │   │
│  │                                                      │   │
│  │  ┌─────────────────────────────────────────────┐   │   │
│  │  │ EKS Cluster (Multi-node)                    │   │   │
│  │  │                                              │   │   │
│  │  │  ┌────────────┐  ┌────────────┐            │   │   │
│  │  │  │ Falco      │  │ TFDrift-   │            │   │   │
│  │  │  │ Daemonset  │  │ Falco Pods │            │   │   │
│  │  │  │            │  │ (x5)       │            │   │   │
│  │  │  └────────────┘  └────────────┘            │   │   │
│  │  │         │                │                   │   │   │
│  │  │         └────────────────┼───────────────┐  │   │   │
│  │  └──────────────────────────┼───────────────┼──┘   │   │
│  │                              │               │      │   │
│  │  ┌───────────────────────────v───────────────v──┐  │   │
│  │  │ Centralized Grafana Stack                   │  │   │
│  │  │ - Loki (S3 backend)                         │  │   │
│  │  │ - Grafana (with LDAP/SSO)                   │  │   │
│  │  │ - Prometheus                                │  │   │
│  │  └─────────────────────────────────────────────┘  │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

**特徴**:
- 組織トレイルで全アカウント監視
- 専用の Security Account で集中監視
- Kubernetes (EKS) でスケーラブル運用

**構成要素**:
1. **Management Account**: 組織トレイル設定
2. **Centralized S3**: 全アカウントのログを集約
3. **Security Account**: 監視基盤を集約
4. **EKS Cluster**: TFDrift-Falco を Kubernetes で運用
5. **Grafana Stack**: 全アカウントのログを可視化

---

## 5. トラブルシューティング

### 5.1 よくある問題

#### 問題: イベントが検知されない

**チェックリスト**:
1. [ ] CloudTrail が有効か
   ```bash
   aws cloudtrail describe-trails
   ```

2. [ ] Falco が CloudTrail ログを読んでいるか
   ```bash
   docker logs falco | grep CloudTrail
   ```

3. [ ] TFDrift-Falco が Falco に接続できているか
   ```bash
   docker logs tfdrift | grep "Connected to Falco"
   ```

4. [ ] 対応している CloudTrail イベントか
   - `docs/AWS_RESOURCE_COVERAGE_ANALYSIS.md` を確認

5. [ ] Terraform State にリソースが存在するか
   ```bash
   terraform show | grep <resource-id>
   ```

#### 問題: 誤検知が多い

**原因**:
1. Terraform State の更新頻度が低い
2. 手動変更後に `terraform import` していない
3. 閾値が低すぎる

**対策**:
1. State 更新頻度を上げる
   ```yaml
   advanced:
     state_refresh_interval: "3m"
   ```

2. terraform apply 後に State リロード
   ```bash
   # CI/CD に追加
   terraform apply && sleep 60  # State 反映待ち
   ```

3. 閾値を調整

#### 問題: メモリ使用量が高い

**原因**:
1. Terraform State が大きい (10,000+ リソース)
2. イベント処理のバッファが溜まっている

**対策**:
1. State を分割
   ```
   ワークスペース別:
   - terraform.tfstate.network
   - terraform.tfstate.compute
   - terraform.tfstate.database
   ```

2. メモリ制限を緩和
   ```yaml
   # docker-compose.yml
   services:
     tfdrift:
       deploy:
         resources:
           limits:
             memory: 4G
   ```

3. ガベージコレクション設定
   ```bash
   # 環境変数
   GOGC=50  # より頻繁に GC
   ```

---

## 6. まとめ

### TFDrift-Falco の適切な使用シーン

✅ **適している**:
- IAM ポリシー変更の監視
- S3 バケットポリシーの監視
- RDS インスタンス変更の監視
- 小〜中規模環境 (リソース < 5,000個)
- 単一アカウント・単一リージョン

⚠️ **注意が必要**:
- 大規模環境 (リソース > 10,000個) → 性能検証必須
- マルチアカウント環境 → 個別検証必須
- 未対応 AWS サービスの監視 → カバレッジ確認必須

❌ **適していない**:
- ミリ秒単位のリアルタイム検知 (CloudTrail 遅延あり)
- 全 AWS サービスの完全な監視 (未対応サービスあり)
- 高信頼性が必要なコンプライアンス監視 (補助ツールとして使用)

### 最終チェックリスト

本番導入前に、以下をすべて確認してください:

- [ ] AWS リソースカバレッジを確認し、使用中のサービスが対応済みか確認
- [ ] サンプルイベントで動作確認
- [ ] 1週間のベースライン測定を実施
- [ ] 閾値を組織に合わせて調整
- [ ] IAM 権限を最小化
- [ ] TLS/mTLS を有効化
- [ ] Grafana のアクセス制御を設定
- [ ] アラート通知先を設定 (Slack/Email)
- [ ] Runbook (対応手順書) を作成
- [ ] バックアップ・リストア手順を確認
- [ ] 障害シナリオのテスト (Falco 停止、メモリ不足など)

**これらを完了すれば、本番環境での導入準備が整います。**

---

**参考ドキュメント**:
- `AWS_RESOURCE_COVERAGE_ANALYSIS.md`: AWS リソース対応状況
- `dashboards/grafana/GETTING_STARTED.md`: Grafana セットアップ
- `dashboards/grafana/ALERTS.md`: アラート設定
- `docs/qiita-getting-started-guide-fixed.md`: 基本セットアップ
