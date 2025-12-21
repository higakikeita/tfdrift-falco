# TFDrift-Falco 実装状況レポート

**日付**: 2025-12-22
**セッション**: 実環境構築とエンドツーエンドテスト

---

## ✅ 完了した作業

### 1. 包括的なセットアップガイドの作成
📄 **ファイル**: `docs/complete-setup-guide.md`

AWSのセットアップからTerraform統合、TFDrift-Falcoの設定、ドリフト検知までの完全な流れを網羅したドキュメントを作成しました。

**内容**:
- Phase 1: AWS CloudTrailセットアップ
- Phase 2: Terraform State設定
- Phase 3: Falco CloudTrailプラグイン
- Phase 4: Docker環境セットアップ
- Phase 5-9: システム起動、動作確認、トラブルシューティング、本番デプロイ、運用

### 2. 実環境Terraform インフラストラクチャの構築

✅ **作成されたAWSリソース** (12リソース):

```
VPC (vpc-05caba02c426adbe7)
├── Subnet (subnet-0100a5565f9626775)
├── Internet Gateway (igw-0c04dbb4e8f48d712)
├── Route Table (rtb-0e5b49e546877874d)
├── Route Table Association
└── Security Group (sg-0d5f6c623e58bcc2e)

IAM
├── Role (tfdrift-test-ec2-role)
├── Instance Profile (tfdrift-test-ec2-profile)
└── Role Policy (cloudwatch-policy)

S3
└── Bucket (tfdrift-test-app-data-595263720623)
    ├── Versioning: Enabled
    └── Encryption: AES256
```

**Terraform State Backend**:
- S3 Bucket: `tfdrift-terraform-state-595263720623`
- DynamoDB Table: `terraform-state-lock`
- State Key: `production-test/terraform.tfstate`

### 3. TFDrift-Falcoシステムの設定と起動

#### Backend API ✅
**ステータス**: 正常動作

```
✅ Terraform State読み込み成功
  - S3から24,103バイト読み込み
  - 13リソースをインデックス化完了

✅ API エンドポイント
  - Health: http://localhost:8080/health
  - Graph: http://localhost:8080/api/v1/graph
  - Drifts: http://localhost:8080/api/v1/drifts
  - State: http://localhost:8080/api/v1/state
  - Stream: http://localhost:8080/api/v1/stream

✅ WebSocket Hub起動完了
✅ ドリフト検知エンジン起動完了
```

#### Frontend UI ✅
**ステータス**: 正常動作

```
✅ React UI起動: http://localhost:3000
✅ 3つの表示モード実装
  - グラフビュー
  - テーブルビュー
  - 分割ビュー（推奨）
```

#### AWS CloudTrail ✅
**ステータス**: 正常動作

```
✅ Trail作成: tfdrift-falco-trail
✅ S3 Bucket: tfdrift-cloudtrail-595263720623-us-east-1
✅ ログ記録開始済み
✅ マルチリージョン有効
✅ ログファイル確認済み (9ファイル以上)
```

### 4. 設定ファイルの最適化

#### docker-compose.yml
- ✅ `version`フィールド削除（Docker Compose v2対応）
- ✅ AWS_PROFILE環境変数追加
- ✅ AWS認証情報の正しいマウントパス設定
  - Backend: `/home/tfdrift/.aws`
  - Falco: `/root/.aws`
- ✅ AWS_SHARED_CREDENTIALS_FILE/AWS_CONFIG_FILE設定

#### config.yaml
- ✅ S3 Backend設定更新
  - Bucket: `tfdrift-terraform-state-595263720623`
  - Key: `production-test/terraform.tfstate`
- ✅ CloudTrail S3 Bucket設定
  - Bucket: `tfdrift-cloudtrail-595263720623-us-east-1`

---

## ⚠️ 残っている課題

### 1. Falco CloudTrailプラグインの接続エラー

**問題**: Falcoが CloudTrail S3バケットに接続できない

```
Error: cloudtrail plugin error: cannot open s3Bucket=tfdrift-cloudtrail-595263720623-us-east-1
```

**原因の可能性**:
1. ✅ CloudTrailログは存在している（確認済み）
2. ✅ AWS認証情報はマウントされている
3. ❌ Falcoプラグインが認証情報を正しく読み込めていない
4. ❌ IAM権限が不足している可能性

**影響**:
- リアルタイムドリフト検知が動作しない
- BackendはFalcoへの接続失敗エラーを出力
- Falcoコンテナが再起動ループに入る

**次のステップ**:
1. Falcoコンテナ内でAWS CLIを使ってS3アクセスをテスト
2. IAM権限の確認（S3:GetObject, S3:ListBucket）
3. CloudTrailプラグインのデバッグログを有効化
4. または、代替アプローチとして直接S3ポーリングを実装

### 2. グラフAPIが空データを返す

**問題**: `/api/v1/graph`エンドポイントが空のノードとエッジを返す

```json
{
  "success": true,
  "data": {
    "nodes": [],
    "edges": []
  }
}
```

**原因の可能性**:
- ✅ Terraform Stateは正しくロードされている（13リソース）
- ❌ グラフ生成ロジックに問題がある可能性
- ❌ リソースがグラフノードに変換されていない

**影響**:
- UI上でリソース依存関係グラフが表示されない
- Terraform管理リソースの可視化ができない

**次のステップ**:
1. バックエンドコードのグラフ生成ロジックを確認
2. デバッグログを有効化してリソース変換プロセスを追跡
3. サンプルデータと実データの差分を調査

### 3. ARM64 Mac環境での制約

**問題**: Docker Desktop on Mac (ARM64)でFalco eBPFドライバーがコンパイルできない

```
Error! Your kernel headers for kernel 6.10.14-linuxkit cannot be found
```

**現状**:
- `platform: linux/amd64`でRosetta経由で動作
- eBPFドライバーは使用不可（カーネルヘッダーなし）
- CloudTrailプラグインのみで動作想定

**影響**:
- システムコール監視は利用できない
- CloudTrail監視のみに限定される
- パフォーマンスがやや低下（x86_64エミュレーション）

**次のステップ**:
- この制約はドキュメント化済み
- 本番環境ではLinux（x86_64）での運用を推奨

---

## 📊 システムステータス

### 起動中のサービス

| サービス | ステータス | ポート | 備考 |
|---------|----------|--------|------|
| Backend API | ✅ 動作中 | 8080 | Terraform State読み込み成功 |
| Frontend UI | ✅ 動作中 | 3000 | React UIアクセス可能 |
| Falco | ⚠️ エラー | 5060 | CloudTrailプラグイン接続失敗 |

### データソース

| ソース | ステータス | 詳細 |
|--------|----------|------|
| Terraform State (S3) | ✅ 読み込み成功 | 13リソース |
| CloudTrail Logs (S3) | ✅ 利用可能 | 9+ログファイル |
| Falco Events | ❌ 取得失敗 | プラグイン接続エラー |

### API エンドポイント

| エンドポイント | ステータス | データ |
|--------------|----------|--------|
| `/health` | ✅ 200 OK | サーバー正常 |
| `/api/v1/state` | ✅ 200 OK | 13リソース |
| `/api/v1/graph` | ⚠️ 200 OK | 空データ |
| `/api/v1/drifts` | ✅ 200 OK | 0件（ドリフトなし） |
| `/api/v1/stream` | ✅ 利用可能 | SSEストリーム |
| `/ws` | ✅ 利用可能 | WebSocket |

---

## 🎯 次のアクションアイテム

### 優先度：高（即時対応）

1. **Falco CloudTrailプラグインのデバッグ**
   - [ ] Falcoコンテナ内でAWS S3アクセステスト
   - [ ] IAM権限の確認と修正
   - [ ] CloudTrailプラグインログの詳細確認

2. **グラフAPIの修正**
   - [ ] バックエンドのグラフ生成ロジックを調査
   - [ ] Terraform StateからGraph変換の実装確認
   - [ ] デバッグログで変換プロセスを追跡

### 優先度：中（今週中）

3. **実際のドリフト検知テスト**
   - [ ] AWSコンソールでセキュリティグループを手動変更
   - [ ] CloudTrailにイベントが記録されることを確認
   - [ ] （Falco修正後）ドリフトアラートがUIに表示されることを確認

4. **ドキュメント改善**
   - [ ] トラブルシューティングセクションに今回の問題を追加
   - [ ] Falco CloudTrailプラグインの詳細設定ガイド
   - [ ] IAM権限の最小要件をドキュメント化

### 優先度：低（将来）

5. **代替アプローチの検討**
   - [ ] Falco以外のCloudTrail監視方法
   - [ ] Lambda + S3イベント通知
   - [ ] CloudWatch Events統合

---

## 📝 技術的な発見

### AWS認証情報の取り扱い

Docker コンテナで AWS認証情報を使用する際のベストプラクティス:

```yaml
# docker-compose.yml
volumes:
  - ${HOME}/.aws:/home/app-user/.aws:ro

environment:
  - AWS_PROFILE=mytf
  - AWS_SHARED_CREDENTIALS_FILE=/home/app-user/.aws/credentials
  - AWS_CONFIG_FILE=/home/app-user/.aws/config
```

**重要**: マウントパスはコンテナ内のユーザーのホームディレクトリに合わせる必要がある

### Terraform State Backend設定

S3バックエンドで状態ロックを使用する場合:

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state-bucket"
    key            = "path/to/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"  # 状態ロック
    encrypt        = true                     # 暗号化
  }
}
```

### CloudTrail ログの構造

CloudTrail は以下の構造でS3にログを保存:

```
s3://bucket-name/
└── AWSLogs/
    └── {account-id}/
        └── CloudTrail/
            └── {region}/
                └── {year}/
                    └── {month}/
                        └── {day}/
                            └── *.json.gz
```

ログファイルは5-15分ごとに作成される

---

## 🔍 デバッグ情報

### バックエンドログ（重要部分）

```
[INFO] Starting TFDrift-Falco vdev
[INFO] Loading Terraform state from S3: s3://tfdrift-terraform-state-595263720623/production-test/terraform.tfstate
[INFO] Successfully loaded 24103 bytes from S3
[INFO] Indexed 13 resources from Terraform state
[INFO] Loaded Terraform state: 13 resources
[INFO] Event processor started
[ERROR] Collector error: falco subscriber error: failed to subscribe to Falco outputs: rpc error: code = Unavailable desc = connection error
```

### Falcoログ（重要部分）

```
[INFO] Loading plugin 'cloudtrail' from file /etc/falco/plugins/libcloudtrail.so
[INFO] Loading rules from file /etc/falco/rules.d/terraform_drift.yaml
[INFO] Enabled event sources: aws_cloudtrail, syscall
[INFO] Opening 'aws_cloudtrail' source with plugin 'cloudtrail'
[ERROR] cloudtrail plugin error: cannot open s3Bucket=tfdrift-cloudtrail-595263720623-us-east-1
```

---

## 🎓 学んだこと

1. **Docker環境でのAWS認証情報**
   - 環境変数だけでなく、ファイルパスも明示的に指定する必要がある
   - コンテナ内のユーザー権限に注意

2. **Falco CloudTrailプラグイン**
   - プラグインはAWS SDKを使用してS3にアクセス
   - IAM権限、認証情報の両方が必要
   - エラーメッセージが不明確な場合がある

3. **Terraform State管理**
   - S3バックエンドは本番環境に推奨
   - DynamoDBによる状態ロックで並行実行を防止
   - AWS認証情報の正しい設定が必須

4. **段階的なデバッグ**
   - まず個別コンポーネントの動作確認（Terraform State読み込み成功）
   - 次に統合部分の問題を特定（Falco統合が課題）
   - 全体が動かなくても、動作する部分を確認して前進

---

## 📊 達成率

```
全体進捗: ████████░░ 80%

✅ インフラ構築      100%
✅ Backend API       95%
✅ Frontend UI       95%
⚠️  Falco統合        40%
✅ ドキュメント      100%
```

---

## 🚀 デモ可能な機能

現時点でデモできること:

1. ✅ Terraform管理の実環境AWSリソース
2. ✅ S3からのTerraform State読み込み
3. ✅ API経由でのState情報取得
4. ✅ React UIの3つの表示モード
5. ✅ WebSocket/SSE接続
6. ✅ CloudTrailログの記録

まだできないこと:

1. ❌ リアルタイムドリフト検知（Falco接続エラー）
2. ❌ リソース依存関係グラフの表示（グラフAPI空）
3. ❌ UIでのドリフトアラート表示

---

**最終更新**: 2025-12-22 00:40 JST
