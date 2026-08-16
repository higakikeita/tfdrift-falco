# Testing Guide

driftwire UI の包括的なテスト戦略とツールのガイドです。

## 📋 テスト概要

| テストタイプ | ツール | カバレッジ | 実行方法 |
|------------|--------|----------|---------|
| **単体テスト** | Vitest | 92.32% | `npm test` |
| **E2E テスト** | Playwright | 5 spec ファイル | `npm run test:e2e` |
| **コンポーネントテスト** | Storybook | 33 stories | `npm run storybook` |
| **アクセシビリティ** | Storybook a11y | 全コンポーネント | Storybook UI |
| **パフォーマンス** | Lighthouse CI | Core Web Vitals | `npm run lighthouse` |
| **ビジュアルリグレッション** | Chromatic | 33 stories | CI で自動実行 |

## 🧪 単体テスト (Vitest)

### テスト実行

```bash
# 全テスト実行
npm test

# ウォッチモード
npm run test:watch

# カバレッジ付き
npm run test:coverage

# UI モード
npm run test:ui
```

### テストファイル構造

```
src/
├── api/
│   └── hooks/
│       ├── useGraph.test.tsx      (68 tests)
│       ├── useEvents.test.tsx     (66 tests)
│       └── useDrifts.test.tsx     (52 tests)
├── components/
│   └── reactflow/
│       ├── CustomNode.test.tsx
│       └── NodeDetailPanel.test.tsx
└── __tests__/
    ├── utils/
    │   └── reactQueryTestUtils.tsx  # テストヘルパー
    └── fixtures/
        ├── graphFixtures.ts         # グラフモックデータ
        ├── eventsFixtures.ts        # イベントモックデータ
        └── driftsFixtures.ts        # ドリフトモックデータ
```

### テストユーティリティの使用

```typescript
import { createQueryClientWrapper } from '../../__tests__/utils/reactQueryTestUtils';
import { createMockNode, createLargeGraphData } from '../../__tests__/fixtures/graphFixtures';

describe('My Component', () => {
  it('should render', () => {
    const { result } = renderHook(() => useGraph(), {
      wrapper: createQueryClientWrapper(),
    });

    // テストコード
  });
});
```

### カバレッジ目標

- **全体**: 90%以上
- **API Hooks**: 95%以上
- **コンポーネント**: 85%以上

現在のカバレッジ: **92.32%** (266 tests)

## 🎭 E2E テスト (Playwright)

### テスト実行

```bash
# 全 E2E テスト実行
npm run test:e2e

# UI モード（デバッグに便利）
npm run test:e2e:ui

# デバッグモード
npm run test:e2e:debug

# 特定のブラウザのみ
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
```

### テストファイル

1. **graph-navigation.spec.ts** - グラフビジュアライゼーションと操作
   - グラフ表示
   - ノードクリック → 詳細パネル
   - ズーム・パン操作
   - キーボードショートカット

2. **onboarding.spec.ts** - オンボーディングフロー
   - ウェルカムモーダル
   - チュートリアルナビゲーション
   - ヘルプオーバーレイ

3. **view-modes.spec.ts** - ビューモード切り替え
   - グラフ / テーブル / 分割ビュー
   - デモモード切り替え
   - レイアウト切り替え

4. **drift-table.spec.ts** - ドリフト履歴テーブル
   - テーブル表示
   - フィルタリング・ソート
   - ページネーション
   - ドリフト詳細パネル

5. **features.spec.ts** - 高度な機能
   - Critical Nodes ハイライト
   - テーマ切り替え
   - パスハイライト
   - レスポンシブデザイン

### ブラウザ・デバイス設定

デフォルトで以下の環境でテスト:

- ✅ Desktop Chrome
- ✅ Desktop Firefox
- ✅ Desktop Safari (Webkit)
- ✅ Mobile Chrome (Pixel 5)
- ✅ Mobile Safari (iPhone 12)

### テスト作成のベストプラクティス

```typescript
test('should perform action', async ({ page }) => {
  // 1. ページ遷移
  await page.goto('/');

  // 2. 要素の待機（タイムアウト設定）
  const button = page.locator('button:has-text("Submit")');
  await button.waitFor({ state: 'visible', timeout: 5000 });

  // 3. アクション
  await button.click();

  // 4. アサーション
  const result = page.locator('[data-testid="result"]');
  await expect(result).toBeVisible();
  await expect(result).toHaveText('Success');
});
```

## 📚 Storybook（コンポーネントカタログ）

### 起動方法

```bash
# 開発モード
npm run storybook
# → http://localhost:6006

# ビルド
npm run build-storybook
```

### ストーリー一覧

#### ReactFlow Components

**CustomNode.stories.tsx** (11 stories)
- Default
- CriticalSeverity / HighSeverity / MediumSeverity / LowSeverity
- Selected
- LongLabel
- GCPResource
- WithMetadata
- Minimal
- Interactive

**NodeDetailPanel.stories.tsx** (9 stories)
- Default
- CriticalSeverity
- MediumSeverityLambda
- NoSeverity
- MinimalData
- ComplexMetadata
- GCPResource
- LongMetadataValues
- Closed

#### Onboarding Components

**WelcomeModal.stories.tsx** (4 stories)
- Default / Open / Closed
- InteractiveTutorial

**HelpOverlay.stories.tsx** (5 stories)
- Default
- FullyInteractive
- WithoutShortcuts
- WithoutTutorial
- Minimal

**KeyboardShortcutsGuide.stories.tsx** (4 stories)
- Default / Open / Closed
- InteractiveGuide

### Storybook アドオン

- **@storybook/addon-docs** - 自動ドキュメント生成
- **@storybook/addon-a11y** - アクセシビリティ監査
- **@chromatic-com/storybook** - ビジュアルリグレッションテスト

### アクセシビリティチェック

各ストーリーで自動的に WCAG 準拠チェック:

1. Storybook を起動
2. コンポーネントを選択
3. 下部の "Accessibility" タブを確認
4. 検出された問題（Violations）を修正

主なチェック項目:
- ✅ Color contrast (色のコントラスト比)
- ✅ Keyboard navigation (キーボード操作)
- ✅ ARIA attributes (ARIA 属性)
- ✅ Form labels (フォームラベル)
- ✅ Heading hierarchy (見出し構造)

## ⚡ パフォーマンステスト (Lighthouse CI)

### ローカル実行

```bash
# アプリケーションをビルド
npm run build

# Lighthouse CI 実行（全ステップ）
npm run lighthouse

# 個別実行
npm run lighthouse:collect  # データ収集
npm run lighthouse:assert   # 閾値チェック
```

### 設定ファイル: `lighthouserc.js`

#### カテゴリ閾値

- **Performance**: 80% 以上（エラー）
- **Accessibility**: 90% 以上（エラー）
- **Best Practices**: 90% 以上（エラー）
- **SEO**: 80% 以上（エラー）

#### Core Web Vitals

| メトリクス | 目標値 | レベル |
|-----------|--------|--------|
| **FCP** (First Contentful Paint) | < 2000ms | 警告 |
| **LCP** (Largest Contentful Paint) | < 2500ms | 警告 |
| **CLS** (Cumulative Layout Shift) | < 0.1 | 警告 |
| **TBT** (Total Blocking Time) | < 300ms | 警告 |

#### リソースサイズ

- **Total Byte Weight**: < 1MB
- **DOM Size**: < 1500 nodes

### 結果の確認

実行後、以下の形式で結果が表示されます:

```
✅ Performance: 85/100
✅ Accessibility: 92/100
✅ Best Practices: 95/100
✅ SEO: 88/100

⚠️  first-contentful-paint: 2100ms (expected < 2000ms)
✅ largest-contentful-paint: 2300ms
✅ cumulative-layout-shift: 0.05
```

## 🎨 ビジュアルリグレッションテスト (Chromatic)

### セットアップ

1. [Chromatic](https://www.chromatic.com/) でアカウント作成
2. プロジェクトを作成
3. プロジェクトトークンを取得
4. GitHub Secrets に `CHROMATIC_PROJECT_TOKEN` を設定

### 使用方法

**自動実行（推奨）**:
- プルリクエストを作成すると自動的に実行
- Chromatic が全ストーリーのスクリーンショットを撮影
- 前回のベースラインと比較
- 変更があればプルリクエストにコメント

**手動実行**:
```bash
npx chromatic --project-token=<token>
```

### レビューフロー

1. プルリクエストを作成
2. Chromatic チェックが実行される
3. ビジュアル変更が検出されたら Chromatic UI でレビュー
4. 意図した変更なら「Accept」
5. バグなら「Deny」して修正

### ベストプラクティス

- コンポーネントに小さな変更を加えるたびにコミット
- 大きなリファクタリングは複数のPRに分割
- 重要なコンポーネントは複数の状態でストーリーを作成

## 🔄 CI/CD での自動実行

### GitHub Actions でのテスト実行

すべてのテストは GitHub Actions で自動実行されます:

**トリガー**:
- プッシュ（main, develop ブランチ）
- プルリクエスト

**実行されるテスト**:
1. ✅ ESLint + TypeScript チェック
2. ✅ Vitest 単体テスト（カバレッジ付き）
3. ✅ Playwright E2E テスト
4. ✅ Storybook ビルド検証
5. ✅ Lighthouse CI パフォーマンス監査
6. ✅ Chromatic ビジュアルテスト（PR のみ）

**定期実行**:
- Lighthouse CI: 毎日 9:00 AM UTC

詳細は [.github/README.md](.github/README.md) を参照。

## 📊 テスト戦略

### テストピラミッド

```
        🔺 E2E Tests (Playwright)
         │  - Critical user flows
         │  - Cross-browser testing
         │  - 5 spec files
        ┌┴┐
       🔷 Component Tests (Storybook)
        │  - UI component isolation
        │  - Visual regression
        │  - 33 stories
       ┌┴┐
      🔶 Unit Tests (Vitest)
       │  - Business logic
       │  - API hooks
       │  - 266 tests, 92.32% coverage
      └┴┘
```

### テストの責任分担

| テストタイプ | 何をテストするか | 何をテストしないか |
|------------|----------------|------------------|
| **単体テスト** | - API フック<br>- ユーティリティ関数<br>- データ変換ロジック | - UI の見た目<br>- ブラウザ統合 |
| **コンポーネントテスト** | - UI の状態<br>- Props による表示変化<br>- ユーザーインタラクション | - API 通信<br>- ルーティング |
| **E2E テスト** | - ユーザーフロー全体<br>- ページ遷移<br>- ブラウザ互換性 | - エッジケース<br>- 詳細なロジック |

## 🐛 デバッグ Tips

### Vitest デバッグ

```bash
# UI モードで視覚的にデバッグ
npm run test:ui

# 特定のテストファイルのみ実行
npm test -- useGraph.test

# デバッグログ出力
DEBUG=* npm test
```

### Playwright デバッグ

```bash
# UI モード（推奨）
npm run test:e2e:ui

# デバッグモード（ステップ実行）
npm run test:e2e:debug

# 特定のテストのみ実行
npx playwright test graph-navigation.spec.ts

# ヘッドフルモード（ブラウザを表示）
npx playwright test --headed

# スクリーンショット・ビデオ確認
# 失敗時に自動保存される
ls test-results/
ls playwright-report/
```

### Storybook デバッグ

1. ブラウザの DevTools を開く
2. React DevTools で Props を確認
3. Storybook の "Actions" タブでイベントログ確認
4. "Controls" タブで Props を動的に変更

## 📈 継続的改善

### テストカバレッジ向上

現在のカバレッジ（92.32%）をさらに向上させる:

1. 未テストのコンポーネントを特定
   ```bash
   npm run test:coverage
   # coverage/index.html を開く
   ```

2. カバレッジが低いファイルにテストを追加

3. エッジケースのテストを追加

### E2E テストの拡充

現在の E2E テストに追加できる項目:

- [ ] エラーハンドリングフロー
- [ ] API エラー時の動作
- [ ] ネットワークオフライン対応
- [ ] 複雑なマルチステップフロー
- [ ] データ永続化（LocalStorage）
- [ ] 多言語対応

### パフォーマンス最適化

Lighthouse で低スコアの項目を改善:

1. 画像最適化（WebP 形式、遅延読み込み）
2. 未使用 JavaScript の削除
3. CSS の最小化
4. コード分割（React.lazy）
5. CDN の活用

## 🎯 品質目標

- ✅ 単体テストカバレッジ: **92.32%** (目標: 90%以上)
- ✅ E2E テストスイート: **5 spec ファイル** (目標: 主要フロー網羅)
- ✅ Storybook ストーリー: **33 stories** (目標: 全 UI コンポーネント)
- ✅ Lighthouse Performance: **目標 80%以上**
- ✅ Lighthouse Accessibility: **目標 90%以上**
- ✅ ビジュアルリグレッション: **Chromatic 統合済み**

## 🔗 参考リンク

- [Vitest ドキュメント](https://vitest.dev/)
- [Playwright ドキュメント](https://playwright.dev/)
- [Storybook ドキュメント](https://storybook.js.org/)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
- [Chromatic ドキュメント](https://www.chromatic.com/docs/)
- [Web.dev - Testing Guide](https://web.dev/testing/)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
