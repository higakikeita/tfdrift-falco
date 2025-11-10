# Auto-Import Guide

TFDrift-Falcoは、Terraformで管理されていないリソースを検出すると、自動的に`terraform import`コマンドを生成・実行できます。

## 機能概要

### 1. コマンド生成のみ（デフォルト）
管理外リソースを検出すると、importコマンドを**表示のみ**します。

```
⚠️  UNMANAGED RESOURCE DETECTED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Resource:
   Type: aws_iam_role
   ID:   production-api-role

💡 Recommendation:
   terraform import aws_iam_role.production_api_role production-api-role
```

### 2. 承認ワークフロー付き自動実行
`auto_import.enabled: true` に設定すると、承認後に自動実行します。

```
🔔 IMPORT APPROVAL REQUIRED
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📦 Resource Type: aws_iam_role
🆔 Resource ID:   production-api-role
📝 Resource Name: production_api_role (auto-generated)
👤 Detected By:   john.doe@company.com
🕐 Detected At:   2025-01-15T14:23:45Z

🔄 Changes:
   role_name: production-api-role
   assume_role_policy: {...}

💻 Import Command:
   terraform import aws_iam_role.production_api_role production-api-role

❓ Approve this import? [y/N]: y
✅ Import approved!
🚀 Executing: terraform import aws_iam_role.production_api_role production-api-role
✅ Import successful!

📄 Generated Terraform code saved to:
   ./infrastructure/imported/aws_iam_role_production_api_role.tf
```

## 設定方法

### config.yaml

```yaml
auto_import:
  # 自動import機能を有効化
  enabled: true

  # Terraformプロジェクトのディレクトリ
  terraform_dir: "./infrastructure"

  # 生成された.tfファイルの保存先
  output_dir: "./infrastructure/imported"

  # 自動importを許可するリソースタイプ（空=全て）
  allowed_resources:
    - "aws_iam_role"
    - "aws_iam_policy"
    - "aws_s3_bucket"
    # EC2インスタンスは危険なので除外
    # - "aws_instance"

  # 承認が必要か（推奨: true）
  require_approval: true
```

## ワークフロー

### オプション1: 手動承認（推奨）

```yaml
auto_import:
  enabled: true
  require_approval: true  # 承認必要
```

**動作:**
1. 管理外リソース検出
2. コンソールに承認プロンプト表示
3. ユーザーが `y` を入力
4. `terraform import` 実行
5. `.tf` ファイル自動生成

**安全性:** ⭐⭐⭐⭐⭐ (最も安全)

### オプション2: 自動承認（特定リソースのみ）

```yaml
auto_import:
  enabled: true
  require_approval: false  # 承認不要
  allowed_resources:       # ホワイトリスト
    - "aws_iam_role"
    - "aws_iam_policy"
```

**動作:**
1. 管理外リソース検出
2. `allowed_resources` をチェック
3. リストにあれば自動実行
4. リストになければスキップ

**安全性:** ⭐⭐⭐⭐ (制限付き自動化)

### オプション3: 完全自動（非推奨）

```yaml
auto_import:
  enabled: true
  require_approval: false
  allowed_resources: []  # 全リソース許可
```

**動作:**
1. 管理外リソース検出
2. 即座に `terraform import` 実行

**安全性:** ⭐⭐ (危険 - 本番環境非推奨)

## 生成されるファイル

### 例: aws_iam_role.production_api_role

`./infrastructure/imported/aws_iam_role_production_api_role.tf`:

```hcl
# Auto-generated resource block for import
# Detected at: 2025-01-15T14:23:45Z
# Detected by: john.doe@company.com
# Resource ID: production-api-role

resource "aws_iam_role" "production_api_role" {
  name = "production-api-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  # TODO: Review and adjust the following attributes
  # path = <value>
  # description = <value>
  # max_session_duration = <value>
  # tags = <value>
}
```

## インタラクティブモード

TFDrift-Falcoをインタラクティブモードで起動:

```bash
tfdrift --config config.yaml --interactive
```

**機能:**
- リアルタイムで承認プロンプト表示
- `y` / `n` で即座に承認/却下
- 承認履歴の表示
- 保留中のリクエスト一覧

### コマンド例

```bash
# 保留中の承認リクエスト一覧
tfdrift approval list

# 特定のリクエストを承認
tfdrift approval approve import-production-api-role-1234567890

# 特定のリクエストを却下
tfdrift approval reject import-test-role-1234567890 --reason "Not needed"

# 期限切れリクエストをクリーンアップ
tfdrift approval cleanup --older-than 24h
```

## API統合

Webhookで外部承認システムと統合:

```yaml
auto_import:
  enabled: true
  require_approval: true
  approval_webhook:
    url: "https://approval-system.company.com/api/approve"
    method: "POST"
    headers:
      Authorization: "Bearer ${APPROVAL_TOKEN}"
```

**Webhook payload:**
```json
{
  "request_id": "import-production-api-role-1234567890",
  "resource_type": "aws_iam_role",
  "resource_id": "production-api-role",
  "import_command": "terraform import aws_iam_role.production_api_role production-api-role",
  "detected_by": "john.doe@company.com",
  "detected_at": "2025-01-15T14:23:45Z",
  "changes": {
    "role_name": "production-api-role",
    "assume_role_policy": {...}
  }
}
```

**承認レスポンス:**
```json
{
  "approved": true,
  "approved_by": "security-team@company.com",
  "reason": "Verified - legitimate role"
}
```

## セキュリティベストプラクティス

### ✅ 推奨設定

1. **承認必須**
   ```yaml
   require_approval: true
   ```

2. **ホワイトリスト使用**
   ```yaml
   allowed_resources:
     - "aws_iam_role"
     - "aws_iam_policy"
   ```

3. **本番環境では無効化**
   ```yaml
   # 本番環境
   auto_import:
     enabled: false  # 表示のみ
   ```

4. **Dry-Runモード**
   ```bash
   tfdrift --config config.yaml --dry-run
   ```

### ❌ 避けるべき設定

1. **全リソース自動承認**
   ```yaml
   # 危険！
   require_approval: false
   allowed_resources: []
   ```

2. **EC2インスタンスの自動import**
   ```yaml
   # 危険！ ステートファイルが巨大になる
   allowed_resources:
     - "aws_instance"
   ```

## トラブルシューティング

### Q: Import が失敗する

**A: Terraformの初期化を確認**
```bash
cd ./infrastructure
terraform init
terraform validate
```

### Q: 生成されたコードが不完全

**A: 手動で編集が必要**

生成されたコードは**基本的な属性のみ**です。以下を手動追加:
- 複雑なネストブロック
- 依存リソース
- タグ
- カスタム属性

### Q: 承認プロンプトが表示されない

**A: インタラクティブモードを確認**
```bash
tfdrift --config config.yaml --interactive
```

## まとめ

| 機能 | 設定 | 安全性 | 用途 |
|------|------|--------|------|
| **表示のみ** | `enabled: false` | ⭐⭐⭐⭐⭐ | 本番環境 |
| **手動承認** | `enabled: true`<br>`require_approval: true` | ⭐⭐⭐⭐⭐ | 推奨 |
| **制限付き自動** | `enabled: true`<br>`allowed_resources: [...]` | ⭐⭐⭐⭐ | 開発環境 |
| **完全自動** | `enabled: true`<br>`require_approval: false`<br>`allowed_resources: []` | ⭐⭐ | テスト環境のみ |

**推奨:** 本番環境では `enabled: false` (表示のみ)、開発環境では `require_approval: true` (手動承認) を使用してください。
