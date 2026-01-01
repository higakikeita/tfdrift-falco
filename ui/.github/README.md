# GitHub Actions CI/CD Configuration

このディレクトリには、TFDrift-Falco UIの継続的インテグレーション（CI）と継続的デプロイメント（CD）のワークフローが含まれています。

## 📋 ワークフロー一覧

### 1. `ci.yml` - メインCI パイプライン

**トリガー**: プッシュ、プルリクエスト（main、develop ブランチ）

**ジョブ**:
- **lint-and-typecheck**: ESLint と TypeScript コンパイラチェック
- **unit-tests**: Vitest による単体テスト実行（カバレッジ付き）
- **e2e-tests**: Playwright による E2E テスト（Chromium のみ）
- **storybook**: Storybook ビルド検証
- **lighthouse**: Lighthouse CI によるパフォーマンス監査

**成果物**:
- カバレッジレポート（Codecov へアップロード）
- Playwright テストレポート
- Storybook ビルド

### 2. `lighthouse-scheduled.yml` - 定期パフォーマンスモニタリング

**トリガー**:
- スケジュール: 毎日 9:00 AM UTC
- 手動実行可能（workflow_dispatch）

**目的**:
- 日次のパフォーマンス監視
- パフォーマンス劣化の早期検出
- Core Web Vitals のトレンド追跡

### 3. `chromatic.yml` - ビジュアルリグレッションテスト

**トリガー**: プッシュ、プルリクエスト（main、develop ブランチ）

**機能**:
- Storybook コンポーネントのビジュアル差分検出
- 変更されたストーリーのみテスト（onlyChanged: true）
- main ブランチでは自動承認

**必要な設定**:
- `CHROMATIC_PROJECT_TOKEN` シークレット

### 4. `storybook-deploy.yml` - Storybook 公開

**トリガー**:
- main ブランチへのプッシュ
- 手動実行可能

**デプロイ先**: GitHub Pages

**URL**: `https://<organization>.github.io/<repository>/`

## 🔧 セットアップ手順

### 必要なシークレットの設定

GitHub リポジトリの Settings > Secrets and variables > Actions で以下を設定:

1. **CHROMATIC_PROJECT_TOKEN** (オプション)
   - [Chromatic](https://www.chromatic.com/) でプロジェクトを作成
   - プロジェクトトークンをコピーして設定

2. **LHCI_GITHUB_APP_TOKEN** (オプション)
   - Lighthouse CI Server を使用する場合
   - 現在は temporary-public-storage を使用

3. **CODECOV_TOKEN** (オプション)
   - [Codecov](https://codecov.io/) でリポジトリを連携
   - トークンを設定してカバレッジレポートをアップロード

### GitHub Pages の有効化

Storybook をデプロイするには:

1. リポジトリの Settings > Pages
2. Source: "GitHub Actions" を選択
3. 保存

最初のデプロイ後、Storybook は `https://<org>.github.io/<repo>/` でアクセス可能になります。

## 📊 CI 実行結果の確認方法

### 単体テストカバレッジ

- Actions タブで `unit-tests` ジョブを確認
- Codecov（設定済みの場合）でカバレッジトレンドを確認

### E2E テスト結果

- Actions タブで `e2e-tests` ジョブを確認
- 失敗時は Artifacts から `playwright-report` をダウンロード
- レポート内の `index.html` をブラウザで開く

### Lighthouse 結果

- Actions タブで `lighthouse` ジョブを確認
- Lighthouse CI ダッシュボード（設定済みの場合）で履歴確認
- Job Summary に主要メトリクスが表示される

### Chromatic ビジュアルテスト

- Actions タブで `chromatic` ジョブを確認
- Chromatic ダッシュボードでビジュアル差分を確認
- プルリクエストにコメントで結果が投稿される

## 🚀 ローカルでの実行方法

CI で実行されるのと同じコマンドをローカルで実行できます:

```bash
# Lint とタイプチェック
npm run lint
npx tsc -b --noEmit

# 単体テスト（カバレッジ付き）
npm run test:coverage

# E2E テスト
npm run build
npm run test:e2e

# Storybook ビルド
npm run build-storybook

# Lighthouse CI
npm run build
npm run lighthouse
```

## 🔄 ワークフローのカスタマイズ

### E2E テストのブラウザを追加

`ci.yml` の `e2e-tests` ジョブで:

```yaml
- name: Install Playwright browsers
  run: npx playwright install --with-deps chromium firefox webkit
```

複数ブラウザでテストすると実行時間が増加するため、CI では Chromium のみ実行しています。

### Lighthouse の閾値調整

`lighthouserc.js` で閾値をカスタマイズ:

```javascript
assertions: {
  'categories:performance': ['error', { minScore: 0.8 }], // 80%以上
  'categories:accessibility': ['error', { minScore: 0.9 }], // 90%以上
}
```

### 定期実行のスケジュール変更

`lighthouse-scheduled.yml` の cron 式を編集:

```yaml
schedule:
  - cron: '0 9 * * *'  # 毎日 9:00 AM UTC
```

[Cron 式の参考](https://crontab.guru/)

## 📈 CI パフォーマンス最適化

### 並列実行

現在のワークフローはジョブを並列実行しています:
- lint-and-typecheck
- unit-tests
- e2e-tests
- storybook
- lighthouse

### キャッシュ

`actions/setup-node@v4` が `npm` キャッシュを自動管理:

```yaml
- uses: actions/setup-node@v4
  with:
    node-version: '20'
    cache: 'npm'  # package-lock.json をキャッシュ
```

### Artifacts の保持期間

ストレージコストを削減するため:

- Playwright レポート: 30日
- Storybook ビルド: 7日

## 🐛 トラブルシューティング

### E2E テストが失敗する

1. ローカルで再現するか確認: `npm run test:e2e`
2. タイムアウトの可能性: `playwright.config.ts` で timeout を調整
3. CI 環境特有の問題: Playwright レポートをダウンロードして確認

### Lighthouse が失敗する

1. ビルドが成功しているか確認
2. `lighthouserc.js` の閾値が厳しすぎないか確認
3. プレビューサーバーが正常に起動しているか確認

### Chromatic デプロイエラー

1. `CHROMATIC_PROJECT_TOKEN` が正しく設定されているか確認
2. Chromatic の利用制限（スナップショット数）を確認
3. Chromatic ダッシュボードでエラー詳細を確認

## 📚 参考リンク

- [GitHub Actions ドキュメント](https://docs.github.com/en/actions)
- [Playwright ドキュメント](https://playwright.dev/)
- [Lighthouse CI ドキュメント](https://github.com/GoogleChrome/lighthouse-ci)
- [Chromatic ドキュメント](https://www.chromatic.com/docs/)
- [Codecov ドキュメント](https://docs.codecov.com/)
