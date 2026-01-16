# Sysdig CLI Scanner セットアップガイド

このガイドでは、GitHub ActionsでSysdig CLI Scannerを使用するための設定方法を説明します。

## 📋 概要

Sysdigスキャンワークフロー（`sysdig-scan.yml`）は、Pull Request時に以下のコンテナイメージをスキャンします：

- **バックエンドイメージ**: Goアプリケーション（Alpine Linux base）
- **フロントエンドイメージ**: Node.js/Nginxアプリケーション（Alpine Linux base）

## 🔑 1. Sysdig Secure API Tokenの取得

### 手順:

1. [Sysdig Secure](https://secure.sysdig.com)にログイン
2. 左下のプロフィールアイコンをクリック → **Settings**
3. **User Profile** セクションの **API Tokens** タブに移動
4. **Create API Token** をクリック
   - Token Name: `GitHub Actions - tfdrift-falco`（任意の名前）
   - Permissions: **Scanning API** の権限を選択
5. トークンをコピー（このトークンは一度しか表示されません）

### トークンの権限

必要な権限:
- ✅ **Scanning API**: コンテナイメージのスキャンに必要

## 🔐 2. GitHub Secretsの設定

### 手順:

1. GitHubリポジトリ [higakikeita/tfdrift-falco](https://github.com/higakikeita/tfdrift-falco) を開く
2. **Settings** タブ → **Secrets and variables** → **Actions** に移動
3. **New repository secret** をクリック
4. 以下を入力:
   - **Name**: `SYSDIG_SECURE_API_TOKEN`
   - **Secret**: 先ほどコピーしたSysdig API Token
5. **Add secret** をクリック

## ✅ 3. 動作確認

### テスト実行方法:

#### オプション1: Pull Requestを作成してテスト

```bash
# 新しいブランチを作成
git checkout -b test/sysdig-scan

# 何か変更を加える（例: Dockerfileにコメント追加）
echo "# Test Sysdig scan" >> Dockerfile

# コミットしてプッシュ
git add .
git commit -m "test: Sysdig scan workflow"
git push origin test/sysdig-scan

# GitHubでPull Requestを作成
```

#### オプション2: 手動実行

1. GitHubリポジトリの **Actions** タブを開く
2. 左サイドバーから **Sysdig Container Security Scan** を選択
3. **Run workflow** ボタンをクリック
4. スキャン対象を選択:
   - `both`: バックエンド＋フロントエンド（デフォルト）
   - `backend`: バックエンドのみ
   - `frontend`: フロントエンドのみ
5. **Run workflow** を実行

### 期待される結果:

✅ **成功時**:
- ワークフローが緑色でパス
- PR上にSysdigスキャン結果のコメントが追加される
- Security タブにSARIF形式の脆弱性レポートが表示される

⚠️ **脆弱性検出時**:
- Critical または High レベルの脆弱性が見つかった場合、ワークフローは失敗します
- PRコメントに詳細な脆弱性情報が表示されます
- Security タブで詳細を確認できます

## 🎯 4. スキャン設定のカスタマイズ

### 失敗しきい値を変更する

デフォルトでは Critical と High レベルの脆弱性でワークフローが失敗します。

より厳格にする（Medium以上で失敗）:
```yaml
severities-to-fail: 'critical,high,medium'
```

より緩和する（Critical のみ失敗）:
```yaml
severities-to-fail: 'critical'
```

### スキャンを失敗させない（警告のみ）

```yaml
ignore-failed-scan: true
```

この設定では、脆弱性が見つかってもワークフローは成功し、レポートだけが生成されます。

### スキャン対象のパスを変更

Pull Request時にスキャンをトリガーするファイルパスを変更:

```yaml
on:
  pull_request:
    paths:
      - 'Dockerfile'
      - 'ui/Dockerfile'
      # 他のパスを追加...
```

## 📊 5. スキャン結果の確認方法

### GitHub Actions画面

1. **Actions** タブ → 該当のワークフロー実行を選択
2. **Summary** セクションでスキャン結果のサマリーを確認
3. 各ジョブ（scan-backend, scan-frontend）のログで詳細を確認

### Pull Requestコメント

PR上に自動的に追加されるコメントで以下を確認:
- バックエンドイメージのスキャン結果
- フロントエンドイメージのスキャン結果
- 検出された脆弱性の重要度と数

### Security タブ

1. リポジトリの **Security** タブを開く
2. **Code scanning** → **Sysdig** を選択
3. SARIF形式のレポートで詳細な脆弱性情報を確認:
   - CVE番号
   - 影響を受けるパッケージ
   - 修正バージョン
   - CVSS スコア

## 🔧 6. トラブルシューティング

### エラー: "Error: Unable to locate executable file: sysdig-cli-scanner"

**原因**: Sysdig CLI Scannerのインストールに失敗

**解決策**:
- ワークフローを再実行する
- GitHub Actionsのランナーの問題の可能性があります

### エラー: "Error: Unauthorized. Check your Sysdig Secure API Token"

**原因**: API Tokenが無効または権限不足

**解決策**:
1. Sysdig Secureで新しいAPI Tokenを生成
2. **Scanning API** の権限があることを確認
3. GitHub Secretsを更新

### エラー: "Error: Image not found"

**原因**: Dockerイメージのビルドに失敗

**解決策**:
1. Dockerfileが正しいパスにあることを確認
2. ビルドログを確認して、ビルドエラーがないか確認
3. ローカルで `docker build` を実行してビルドが成功するか確認

### スキャンが遅い

**原因**: イメージサイズが大きい、またはキャッシュが効いていない

**解決策**:
- GitHub Actionsのキャッシュが有効になっていることを確認（既に設定済み）
- マルチステージビルドでイメージサイズを最適化（既に実装済み）

## 📚 7. 追加リソース

- [Sysdig Secure ドキュメント](https://docs.sysdig.com/en/docs/sysdig-secure/)
- [Sysdig CLI Scanner GitHub Action](https://github.com/sysdiglabs/scan-action)
- [コンテナイメージのベストプラクティス](https://docs.sysdig.com/en/docs/sysdig-secure/vulnerabilities/scanning/container-image-scanning/)

## 🎉 完了

これでSysdig CLI Scannerの設定は完了です！

Pull Requestを作成すると、自動的にコンテナイメージがスキャンされ、脆弱性レポートが生成されます。
