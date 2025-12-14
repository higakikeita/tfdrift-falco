# React2Shell脆弱性（CVE-2025-55182他）への対応とSnyk導入による継続的セキュリティ監視

## はじめに

TFDrift-Falcoのプロジェクトサイト公開直後、Reactに複数の重大な脆弱性が発見されました。この記事では、**React2Shell（CVE-2025-55182）**をはじめとする4つの脆弱性への対応と、Snykを使った継続的なセキュリティ監視体制の構築について紹介します。

- **プロジェクトサイト**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco

## 発見された脆弱性

### CVE-2025-55182: React2Shell（Critical - CVSS 10.0）

2025年12月3日に公開された、**最高深刻度の脆弱性**です。

```
脆弱性: React Server Componentsでの安全でないデシリアライゼーション
影響: 認証なしでリモートコード実行（RCE）が可能
攻撃: 中国系脅威グループが公開後数時間で悪用開始
```

#### 技術的詳細

- **原因**: Flight protocolにおける安全でないデシリアライゼーション
- **攻撃方法**: 悪意のあるHTTPリクエストを送信
- **成功率**: ほぼ100%、デフォルト設定で脆弱
- **影響範囲**: React Server Components使用の全アプリ

### その他の関連脆弱性

#### CVE-2025-55184: サービス拒否（High - CVSS 7.5）

```yaml
問題: 悪意のあるHTTPリクエストで無限ループを引き起こす
影響: サービスがダウンし、利用不可になる
```

#### CVE-2025-67779: 不完全な修正（High - CVSS 7.5）

```yaml
問題: CVE-2025-55184の初回修正が不完全
影響: React 19.0.2, 19.1.3, 19.2.2が依然として脆弱
```

#### CVE-2025-55183: ソースコード露出（Medium - CVSS 5.3）

```yaml
問題: サーバー関数のソースコードが露出
影響: APIキーなどのハードコードされた機密情報が漏洩
```

## 脆弱性の発見

### 1. 気づいたきっかけ

プロジェクトサイト公開後、セキュリティニュースで**React2Shell**の情報を目にしました。

```bash
# 使用していたバージョンを確認
cat website/package.json | grep react
```

```json
{
  "react": "19.2.1",        // ← 脆弱！
  "react-dom": "19.2.1",    // ← 脆弱！
  "next": "16.0.10"         // ← これは安全
}
```

### 2. npm auditの結果

```bash
cd website
npm audit
```

結果：
```
found 0 vulnerabilities
```

**驚くべきことに、npm auditは検出しませんでした。**

これは：
- npm auditのデータベースが最新でない
- 新しい脆弱性の登録に時間がかかる

という問題を示しています。

### 3. 手動での確認

公式情報を確認：

- [React公式ブログ](https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components)
- [Next.js Security Update](https://nextjs.org/blog/security-update-2025-12-11)

**結論**: React 19.2.1は脆弱、19.2.3が必要

## Snykの導入

npm auditでは検出できなかったため、より高度なセキュリティツール**Snyk**を導入することにしました。

### Snykとは？

- **開発者向けセキュリティプラットフォーム**
- オープンソースの脆弱性データベース
- npm auditより検出精度が高い
- GitHub Actionsとの統合が容易

### セットアップ手順

#### 1. Snykアカウント作成

```bash
# 1. https://snyk.io/ にアクセス
# 2. GitHubアカウントでサインアップ
# 3. APIトークンを取得
```

#### 2. GitHub Secretsに登録

```bash
# GitHubリポジトリ → Settings → Secrets and variables → Actions
# New repository secret
Name: SNYK_TOKEN
Secret: (SnykのAPIトークン)
```

#### 3. GitHub Actionsワークフローを作成

```yaml
# .github/workflows/website-security.yml
name: Website Security Scan

on:
  push:
    branches: [main]
    paths:
      - 'website/**'
  pull_request:
    branches: [main]
    paths:
      - 'website/**'
  schedule:
    - cron: '0 9 * * 1' # 毎週月曜 9:00 UTC
  workflow_dispatch:

jobs:
  security:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: website/package-lock.json

      - name: Install dependencies
        working-directory: ./website
        run: npm ci

      - name: Run Snyk to check for vulnerabilities
        uses: snyk/actions/node@master
        continue-on-error: true
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
        with:
          args: --severity-threshold=high --sarif-file-output=snyk.sarif
          command: test
          working-directory: website

      - name: Upload Snyk results to GitHub Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        if: always()
        with:
          sarif_file: snyk.sarif
          category: website-security
```

#### 4. カスタムセキュリティチェックスクリプト

Snykに加えて、即座に確認できるスクリプトも作成：

```bash
#!/bin/bash
# check-security.sh

echo "🔍 Security Check"
echo "================="

REACT_VERSION=$(node -p "require('./package.json').dependencies.react")
REACT_VER_NUM=$(echo $REACT_VERSION | sed 's/[\^~]//g')

# React 19.2.1, 19.2.2, 19.1.3, 19.0.2 は脆弱
if [[ "$REACT_VER_NUM" == "19.2.1" ]] || [[ "$REACT_VER_NUM" == "19.2.2" ]]; then
    echo "⚠️  VULNERABLE: React $REACT_VER_NUM"
    echo "   CVE-2025-55182: RCE (Critical)"
    echo "   CVE-2025-55183: Source Code Exposure (Medium)"
    echo "   CVE-2025-55184: DoS (High)"
    echo "   CVE-2025-67779: Incomplete fix (High)"
    echo ""
    echo "🔧 Fix: npm install react@19.2.3 react-dom@19.2.3"
    exit 1
elif [[ "$REACT_VER_NUM" == "19.2.3" ]]; then
    echo "✅ SAFE: React $REACT_VER_NUM (patched)"
    exit 0
fi
```

実行：
```bash
chmod +x check-security.sh
./check-security.sh
```

## 脆弱性の修正

### 1. Reactのアップデート

```bash
cd website

# パッチ済みバージョンにアップデート
npm install react@19.2.3 react-dom@19.2.3
```

### 2. バージョン確認

```bash
# package.json を確認
cat package.json | grep react
```

```json
{
  "react": "^19.2.3",      // ✅ 安全
  "react-dom": "^19.2.3"   // ✅ 安全
}
```

### 3. ビルド確認

```bash
npm run build
```

```
✓ Compiled successfully
✓ Generating static pages (8/8)

Route (app)
┌ ○ /
├ ○ /blog
├ ● /blog/[slug]
└ ○ /releases

✓ Build completed successfully
```

### 4. セキュリティ再チェック

```bash
./check-security.sh
```

```
🔍 Security Check
=================
📦 Current Versions:
  React: ^19.2.3
  Next.js: 16.0.10

✅ SAFE: React 19.2.3 (patched)
✅ SAFE: Next.js 16.0.10

✅ No known vulnerabilities detected
```

### 5. コミット＆デプロイ

```bash
git add package.json package-lock.json check-security.sh
git commit -m "security: Fix critical React vulnerabilities (CVE-2025-55182/55183/55184/67779)"
git push origin main
```

Vercelが自動的に再デプロイ → **数分で本番環境に反映**

## 継続的なセキュリティ監視体制

### 1. 自動スキャンのタイミング

```yaml
# GitHub Actionsが以下のタイミングで自動実行
on:
  push:              # mainブランチへのプッシュ時
  pull_request:      # PR作成時
  schedule:          # 毎週月曜 9:00 UTC
  workflow_dispatch: # 手動実行
```

### 2. GitHub Code Scanningとの統合

Snykの結果は**GitHub Code Scanning**に統合されます：

```
リポジトリ → Security → Code scanning alerts
```

脆弱性が検出されると：
- 自動的にアラート作成
- 影響を受けるファイルを表示
- 修正方法を提案

### 3. 通知設定

```
GitHub → Settings → Notifications → Security alerts
```

以下を有効化：
- Dependabot alerts
- Code scanning alerts
- Secret scanning alerts

### 4. ローカルでの定期チェック

```bash
# 週次でローカル確認（開発者の習慣化）
cd website
./check-security.sh

# Snyk CLIでも確認
npm install -g snyk
snyk auth
snyk test
```

## 学んだこと

### 1. npm auditは万能ではない

```
npm audit: 0 vulnerabilities ❌
実際: 4つの重大な脆弱性 ⚠️
```

**教訓**: 複数のツールを組み合わせる

### 2. 公開直後でも脆弱性は発生する

```
ウェブサイト公開: 2025-12-14
React2Shell公開: 2025-12-03（わずか11日前）
```

**教訓**: 継続的な監視が必須

### 3. 自動化の重要性

手動チェックだけでは：
- チェック漏れが発生
- 対応が遅れる
- 人的リソースを消費

**教訓**: CI/CDに組み込む

### 4. 多層防御の原則

```
1. Snyk（高精度な検出）
2. npm audit（基本的なチェック）
3. カスタムスクリプト（即座の確認）
4. GitHub Code Scanning（可視化）
5. 手動での情報収集（最新情報）
```

**教訓**: 1つのツールに依存しない

## セキュリティベストプラクティス

### 1. 依存関係の定期更新

```bash
# 月次で実行
npm outdated
npm update
npm audit fix
```

### 2. Dependabotの活用

```yaml
# .github/dependabot.yml
version: 2
updates:
  - package-ecosystem: "npm"
    directory: "/website"
    schedule:
      interval: "weekly"
    open-pull-requests-limit: 10
```

### 3. セキュリティポリシーの明文化

```markdown
# SECURITY.md

## 脆弱性報告
security@example.com

## サポートバージョン
| Version | Supported |
| ------- | --------- |
| 1.x.x   | ✅        |
| 0.x.x   | ❌        |
```

### 4. 環境変数の適切な管理

```bash
# ❌ ソースコードにハードコード
const API_KEY = "sk-1234567890"

# ✅ 環境変数を使用
const API_KEY = process.env.API_KEY
```

### 5. 最小権限の原則

```yaml
# GitHub Actions
permissions:
  contents: read        # 読み取りのみ
  security-events: write # セキュリティイベントのみ書き込み
```

## 対応のタイムライン

```
2025-12-03: React2Shell (CVE-2025-55182) 公開
2025-12-11: 追加の脆弱性公開 (CVE-2025-55183/55184/67779)
2025-12-14: TFDrift-Falcoサイト公開（脆弱なバージョン）
2025-12-14: 脆弱性を認識
2025-12-14: Snyk導入開始
2025-12-14: React 19.2.3にアップデート
2025-12-14: セキュリティチェック自動化完了

対応時間: 発見から修正完了まで約2時間
```

## コスト

すべて**無料**で実現できています：

- **Snyk**: オープンソースプロジェクトは無料
- **GitHub Actions**: 月2000分まで無料
- **GitHub Code Scanning**: パブリックリポジトリは無料
- **Vercel**: Hobbyプランは無料

## 今後の改善計画

### 短期（1ヶ月以内）

1. **Dependabot有効化**
   - 自動PR作成
   - 定期的な依存関係更新

2. **SBOM生成**
   - ソフトウェア部品表の作成
   - 依存関係の可視化

3. **セキュリティドキュメント整備**
   - SECURITY.md
   - 脆弱性対応フロー

### 中期（3ヶ月以内）

1. **Container Scanning追加**
   - Dockerイメージのスキャン
   - ベースイメージの脆弱性チェック

2. **SAST/DAST導入**
   - 静的解析（SAST）
   - 動的解析（DAST）

3. **セキュリティ教育**
   - チーム内での情報共有
   - セキュアコーディングガイドライン

### 長期（6ヶ月以内）

1. **Bug Bountyプログラム**
   - セキュリティ研究者からのフィードバック

2. **ペネトレーションテスト**
   - 定期的な外部監査

3. **インシデント対応計画**
   - 緊急時の対応フロー
   - バックアップ・復旧計画

## まとめ

React2Shell（CVE-2025-55182）への対応を通じて学んだこと：

✅ **複数のセキュリティツールを併用する**
- npm audit
- Snyk
- カスタムスクリプト
- GitHub Code Scanning

✅ **自動化による継続的監視**
- CI/CDパイプラインに統合
- 定期実行（スケジュール）
- リアルタイムアラート

✅ **迅速な対応体制**
- 発見から修正まで2時間
- 自動デプロイで即反映

✅ **多層防御の実践**
- 検出・対応・監視の3層構造
- 1つのツールに依存しない

✅ **コミュニティとの連携**
- 公式情報の定期確認
- セキュリティコミュニティへの参加

セキュリティは**一度やって終わりではなく、継続的なプロセス**です。今回構築した体制により、今後の脆弱性にも迅速に対応できるようになりました。

## 参考リンク

### 公式情報
- [React - Critical Security Vulnerability in React Server Components](https://react.dev/blog/2025/12/03/critical-security-vulnerability-in-react-server-components)
- [React - Denial of Service and Source Code Exposure](https://react.dev/blog/2025/12/11/denial-of-service-and-source-code-exposure-in-react-server-components)
- [Next.js Security Update: December 11, 2025](https://nextjs.org/blog/security-update-2025-12-11)
- [Vercel Security Bulletin](https://vercel.com/kb/bulletin/security-bulletin-cve-2025-55184-and-cve-2025-55183)

### セキュリティ解析
- [AWS Security Blog - React2Shell](https://aws.amazon.com/blogs/security/china-nexus-cyber-threat-groups-rapidly-exploit-react2shell-vulnerability-cve-2025-55182/)
- [Qualys - React2Shell Decoding](https://blog.qualys.com/product-tech/2025/12/10/react2shell-decoding-cve-2025-55182-the-silent-threat-in-react-server-components)

### ツール
- [Snyk](https://snyk.io/)
- [GitHub Code Scanning](https://docs.github.com/en/code-security/code-scanning)

### プロジェクト
- **TFDrift-Falco**: https://tfdrift-falco.vercel.app/
- **GitHub**: https://github.com/higakikeita/tfdrift-falco

---

**セキュリティは継続的な取り組みです。一緒に安全なソフトウェアを作りましょう！**

質問やフィードバックは[GitHub Discussions](https://github.com/higakikeita/tfdrift-falco/discussions)でお待ちしています。
