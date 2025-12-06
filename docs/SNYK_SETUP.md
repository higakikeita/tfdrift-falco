# Snyk Security Scanning Setup Guide

このガイドでは、TFDrift-FalcoリポジトリでSnykセキュリティスキャンを有効化する手順を説明します。

## 前提条件

- GitHubリポジトリへの管理者アクセス権限
- Snykアカウント（無料でOSS利用可能）

## セットアップ手順

### 1. Snykアカウントの作成

1. [snyk.io](https://snyk.io/) にアクセス
2. "Sign Up Free" をクリック
3. GitHubアカウントで登録（推奨）

### 2. Snyk APIトークンの取得

1. Snykにログイン後、右上のユーザーアイコンをクリック
2. "Account Settings" を選択
3. 左メニューから "General" → "API Token" に移動
4. "KEY" セクションの目隠しアイコンをクリックしてトークンを表示
5. トークンをコピー（形式: `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`）

### 3. GitHubシークレットの設定

#### メインリポジトリ（keitahigaki/tfdrift-falco）の場合

1. GitHubリポジトリページを開く
2. "Settings" タブをクリック
3. 左メニューから "Secrets and variables" → "Actions" を選択
4. "New repository secret" ボタンをクリック
5. 以下を入力：
   - **Name**: `SNYK_TOKEN`
   - **Secret**: コピーしたSnyk APIトークン
6. "Add secret" をクリック

#### フォークの場合

フォークしたリポジトリでSnykを有効化する場合も同様の手順で設定します：

1. 自分のフォークリポジトリの "Settings" に移動
2. "Secrets and variables" → "Actions"
3. `SNYK_TOKEN` を追加

### 4. 動作確認

#### CI/CDでの自動実行

設定が完了すると、以下のタイミングでSnykスキャンが自動実行されます：

- `main` / `develop` ブランチへのプッシュ
- Pull Request作成時
- 毎週月曜日 9:00 AM UTC（スケジュール実行）

#### 手動での確認方法

1. GitHubリポジトリの "Actions" タブを開く
2. "Security" ワークフローを選択
3. 最新の実行結果を確認
4. "Snyk Security Scan" ジョブが成功していることを確認

#### GitHub Code Scanningでの確認

Snykの結果はGitHub Code Scanningにも統合されます：

1. リポジトリの "Security" タブを開く
2. "Code scanning" セクションを確認
3. Snykによって検出された脆弱性が表示されます

### 5. ローカルでのSnyk実行（オプション）

CI/CDでの実行とは別に、ローカルでもSnykを実行できます：

```bash
# Snyk CLIのインストール
npm install -g snyk

# 認証
snyk auth

# 脆弱性スキャン
snyk test

# 自動修正の提案
snyk wizard
```

## トラブルシューティング

### エラー: "Snyk token not found"

**原因**: `SNYK_TOKEN` シークレットが設定されていない

**解決策**:
1. リポジトリの Settings → Secrets and variables → Actions を確認
2. `SNYK_TOKEN` が存在することを確認
3. 存在しない場合は、上記手順3に従って追加

### エラー: "Authentication failed"

**原因**: 無効なトークンまたは期限切れ

**解決策**:
1. Snykアカウントで新しいトークンを生成
2. GitHubシークレットの `SNYK_TOKEN` を更新
3. ワークフローを再実行

### フォークでSnykが動作しない

**原因**: セキュリティ上の理由で、外部フォークからはシークレットにアクセスできない

**解決策**:
1. 自分のフォークリポジトリに `SNYK_TOKEN` を設定
2. または、`.github/workflows/security.yml` の以下の行が正しく設定されていることを確認：
   ```yaml
   if: github.event_name != 'pull_request' || github.event.pull_request.head.repo.full_name == github.repository
   ```

### SARIF ファイルのアップロードエラー

**原因**: SARIF ファイルが生成されていない

**解決策**:
`security.yml` で以下が設定されていることを確認：
```yaml
with:
  args: --severity-threshold=high --sarif-file-output=snyk.sarif
  command: test
```

## Snykの設定カスタマイズ

### 重大度の閾値を変更

`.github/workflows/security.yml` の `args` を編集：

```yaml
# 高・重大のみ検出
args: --severity-threshold=high --sarif-file-output=snyk.sarif

# 中・高・重大を検出
args: --severity-threshold=medium --sarif-file-output=snyk.sarif

# すべての脆弱性を検出
args: --severity-threshold=low --sarif-file-output=snyk.sarif
```

### 特定の脆弱性を無視

プロジェクトルートに `.snyk` ファイルを作成：

```yaml
# .snyk
version: v1.22.0
ignore:
  SNYK-GOLANG-GITHUBCOMAZUREAZURESDKFORGO-1234567:
    - '*':
        reason: False positive
        expires: 2025-12-31T00:00:00.000Z
```

### スキャン対象を限定

```yaml
# 特定のディレクトリのみスキャン
args: --file=go.mod --severity-threshold=high --sarif-file-output=snyk.sarif
```

## ベストプラクティス

### 1. 定期的なトークンローテーション

セキュリティのため、3-6ヶ月ごとにSnyk APIトークンを再生成することを推奨します。

### 2. チーム全体での脆弱性対応

Snykで検出された脆弱性は、GitHub Issuesとして自動作成することも可能です：

1. Snykダッシュボードで "Integrations" を開く
2. "GitHub" を選択
3. "Create automatic fix PRs" を有効化

### 3. 依存関係の定期更新

```bash
# 依存関係の更新確認
go list -u -m all

# 更新の適用
go get -u ./...
go mod tidy
```

### 4. ローカルでの事前チェック

プッシュ前にローカルでセキュリティスキャンを実行：

```bash
./scripts/security-scan.sh
```

## 参考リンク

- [Snyk公式ドキュメント](https://docs.snyk.io/)
- [Snyk GitHub Actions](https://github.com/snyk/actions)
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)
- [プロジェクトのSECURITY.md](../.github/SECURITY.md)

## サポート

質問や問題がある場合：

1. [GitHub Issues](https://github.com/keitahigaki/tfdrift-falco/issues) で質問
2. `.github/SECURITY.md` でセキュリティポリシーを確認
3. Snykサポート: support@snyk.io
