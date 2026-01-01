# Grafanaを捨ててReactで作り直した話 - テストカバレッジ92%達成までの開発記

## はじめに

こんにちは。今回は、クラウドインフラのセキュリティとドリフト分析を可視化するプロジェクト「TFDrift-Falco」で、**GrafanaベースのUIを完全に捨て、Reactで専用UIを作り直した**経験を共有します。

結果として、開発速度は10倍、テストカバレッジ92%、266件のテストを持つ堅牢なアプリケーションが完成しました。

**TFDrift-Falcoとは？**
- Terraform Drift → IAM → Kubernetes → Falcoの因果関係を追跡
- React Flow + Cytoscape.jsによるインタラクティブなグラフ可視化
- セキュリティリスクの影響範囲を視覚的に分析

この記事では、約2週間で実施した以下の作業について詳しく解説します：
1. UI/UXの全面改装（Phase 1-4）
2. プロジェクトドキュメントの整備
3. テスト基盤の構築
4. 266件のテスト作成とカバレッジ92%達成

## なぜGrafanaから移行したのか？

### 当初の構成：Grafana + Node Graph Panel

プロジェクト開始当初、私たちはGrafanaのNode Graph Panelを使って依存関係グラフを可視化していました。

**Grafanaを選んだ理由：**
- 既存のモニタリング基盤との統合が容易
- 設定ファイルベースで管理できる
- ダッシュボードが素早く作れる
- 運用チームが慣れている

しかし、開発を進めるうちに**決定的な限界**に直面しました。

### Grafanaの限界

#### 1. インタラクティブ性の欠如

```
❌ できなかったこと：
- ノードをクリックして詳細情報を表示
- ダブルクリックでフォーカスビュー
- 右クリックでコンテキストメニュー
- ドラッグ&ドロップでノード位置調整
- 選択状態の保持とハイライト
```

Grafanaのパネルは基本的に**表示専用**です。グラフを見ることはできても、操作して深掘りすることができません。

セキュリティ分析では「このリソースが影響を受けたら、どこまで波及するか？」を**インタラクティブに探索**する必要があります。

#### 2. カスタムUIコンポーネントの制限

```typescript
// ❌ Grafanaではこういうカスタムコンポーネントが作れない
<NodeDetailPanel
  node={selectedNode}
  onClose={() => setSelectedNode(null)}
  metadata={nodeMetadata}
/>

<WelcomeModal
  steps={onboardingSteps}
  onComplete={handleComplete}
/>

<KeyboardShortcutsGuide
  shortcuts={appShortcuts}
/>
```

Grafanaのプラグインシステムでは、細かいUX制御が困難です：
- モーダルダイアログ
- スライドインパネル
- 複数ステップのウィザード
- カスタムキーボードショートカット
- ツールチップの詳細制御

#### 3. レイアウトアルゴリズムの固定化

Grafanaの Node Graph Panel は、レイアウトアルゴリズムが限定的です：

```
利用可能: force-directed (力学モデル)
必要だったもの:
  ✓ Dagre (階層レイアウト)
  ✓ Concentric (放射状レイアウト)
  ✓ Grid (グリッドレイアウト)
  ✓ Custom (カスタムレイアウト)
```

セキュリティの影響範囲分析では、階層構造を明確に示すDagreレイアウトが最適でした。

#### 4. パフォーマンスの問題

```
ノード数: 100+ 、エッジ数: 200+
→ Grafana Node Graph が重くなる
→ ブラウザのメモリ使用量が増大
→ 操作がもたつく
```

大規模なグラフでは、Grafanaのレンダリングエンジンが追いつきませんでした。

#### 5. 開発サイクルの遅さ

Grafanaでの開発：
```bash
1. プラグインコードを変更
2. ビルド (1-2分)
3. Grafanaを再起動
4. ブラウザをリロード
5. 確認
```

React開発：
```bash
1. コードを変更
2. HMR（Hot Module Replacement）で即座に反映
3. 確認
```

開発速度が**10倍以上**変わりました。

### 決断：専用UIへの全面移行

これらの課題を解決するため、**React + React Flow + Cytoscape.js** による専用UIへの全面移行を決断しました。

**移行の決め手となった要因：**

1. **ユーザー中心設計の必要性**
   - セキュリティアナリストが求めるのは「探索的な分析」
   - 静的なダッシュボードでは不十分
   - インタラクティブな操作が必須

2. **機能拡張の柔軟性**
   - 今後追加したい機能（時系列分析、AIによる推奨等）
   - Grafanaの制約では実現困難
   - Reactなら自由に拡張可能

3. **開発生産性の向上**
   - HMRによる高速開発
   - TypeScriptによる型安全性
   - 豊富なエコシステム（Testing Library、Storybook等）

4. **チーム構成の変化**
   - フロントエンドエンジニアの参画
   - モダンなReact開発スキルの活用
   - GrafanaプラグインよりReactの方が採用しやすい

### 移行の結果

| 項目 | Grafana | React専用UI | 改善 |
|-----|---------|------------|------|
| ノードクリック操作 | ❌ 不可 | ✅ 詳細パネル表示 | - |
| ダブルクリック | ❌ 不可 | ✅ フォーカスビュー | - |
| 右クリックメニュー | ❌ 不可 | ✅ コンテキストメニュー | - |
| カスタムレイアウト | ❌ 1種類 | ✅ 5種類 | 5倍 |
| パフォーマンス (100ノード) | △ 遅い | ✅ 高速 | 3倍 |
| 開発サイクル | △ 数分 | ✅ 数秒 | 10倍以上 |
| テスタビリティ | ❌ 困難 | ✅ 266テスト | - |
| オンボーディング | ❌ なし | ✅ 6ステップ | - |

**結果として、専用UIへの移行は大成功**でした。

ユーザーからは「探索がスムーズになった」「使いやすくなった」との声をいただき、開発チームの生産性も劇的に向上しました。

### Grafanaは完全に捨てたのか？ - 適材適所の使い分け

**答えは「No」です。** Grafanaは今でもプロジェクトの重要な一部として活躍しています。

重要なのは、**「それぞれのツールを最適な用途で使う」**ことです。

#### Grafanaを使い続けている部分

**1. バックエンドAPIのモニタリング**

```yaml
# Prometheus + Grafana での監視項目
metrics:
  - name: api_response_time
    type: histogram
    description: APIエンドポイントのレスポンスタイム

  - name: drift_detection_count
    type: counter
    description: ドリフト検出数

  - name: graph_query_duration
    type: histogram
    description: GraphDBクエリの実行時間

  - name: error_rate
    type: gauge
    description: エラー発生率
```

**ダッシュボード例：**
- APIレスポンスタイムの推移（P50/P95/P99）
- エンドポイント別のスループット
- エラーレート（5xx/4xx）
- データベース接続プール使用率

**2. ログ分析（Grafana Loki）**

```yaml
# Loki でのログクエリ例
queries:
  # Terraform実行ログ
  - name: terraform_apply_logs
    query: '{job="terraform-scanner"} |= "drift detected"'

  # Falco重大イベント
  - name: falco_critical_events
    query: '{job="falco"} | json | severity="critical"'

  # API異常系ログ
  - name: api_errors
    query: '{job="api"} | json | level="error"'

  # ドリフト検出の詳細
  - name: drift_details
    query: '{job="drift-scanner"} | json | resource_type=~"aws_.*"'
```

**ログダッシュボード例：**
- ドリフト検出イベントのタイムライン
- Falcoアラートの発生頻度
- エラーログの集計と分類
- Terraform実行履歴

**3. ビジネスメトリクス＆トレンド分析**

```
┌─────────────────────────────────────────┐
│ Grafana Dashboard: セキュリティ概要     │
├─────────────────────────────────────────┤
│                                         │
│  📊 ドリフト検出数（24時間）            │
│  ┌───────────────────────────┐         │
│  │  Critical: 15 ↑           │         │
│  │  High:     45 ↓           │         │
│  │  Medium:   123 →          │         │
│  └───────────────────────────┘         │
│                                         │
│  📈 週次トレンド                        │
│  [グラフ: ドリフト検出数の推移]         │
│                                         │
│  🎯 影響範囲トップ5                     │
│  1. IAM Role (影響ノード: 45)          │
│  2. S3 Bucket (影響ノード: 32)         │
│  3. Lambda Function (影響ノード: 28)   │
│  ...                                    │
│                                         │
│  🔔 アラート設定                        │
│  ✓ Critical drift detected > 10/hour   │
│  ✓ Falco events > 50/min               │
│  ✓ API error rate > 5%                 │
└─────────────────────────────────────────┘
```

**4. アラート＆通知設定**

```yaml
# Grafana Alerting Rules
alerts:
  - name: critical_drift_spike
    condition: |
      rate(drift_detection_total{severity="critical"}[5m]) > 10
    annotations:
      summary: "Critical driftが急増しています"
      description: "過去5分間でcritical driftが{{ $value }}件検出されました"
    notify:
      - slack: "#security-alerts"
      - pagerduty: "on-call-team"

  - name: api_high_error_rate
    condition: |
      rate(http_requests_total{status=~"5.."}[5m]) /
      rate(http_requests_total[5m]) > 0.05
    annotations:
      summary: "APIエラー率が5%を超えています"
    notify:
      - slack: "#engineering"

  - name: falco_critical_event
    condition: |
      sum(rate(falco_events_total{severity="critical"}[1m])) > 5
    annotations:
      summary: "Falco critical イベントが頻発しています"
      runbook_url: "https://wiki.example.com/runbooks/falco-critical"
    notify:
      - slack: "#security-incidents"
      - pagerduty: "security-team"
```

**5. システムヘルスモニタリング**

```
監視項目:
┌─────────────────────────────────────┐
│ インフラメトリクス                  │
├─────────────────────────────────────┤
│ ✓ GraphDB CPU/Memory                │
│ ✓ API Server スループット            │
│ ✓ データベース接続数                 │
│ ✓ Redis キャッシュヒット率           │
│                                     │
│ アプリケーションメトリクス          │
├─────────────────────────────────────┤
│ ✓ グラフ生成時間                    │
│ ✓ ドリフトスキャン実行時間           │
│ ✓ Falco処理レイテンシ               │
│ ✓ APIレスポンスタイム（P99）         │
└─────────────────────────────────────┘
```

#### 完全なアーキテクチャ

```
┌────────────────────────────────────────────────────────┐
│                   TFDrift-Falco System                 │
├────────────────────────────────────────────────────────┤
│                                                        │
│  ┌─────────────────┐    ┌──────────────────┐         │
│  │  Frontend UI    │    │  Grafana         │         │
│  │  (React専用)    │    │  Dashboards      │         │
│  ├─────────────────┤    ├──────────────────┤         │
│  │ ✓ グラフ可視化  │    │ ✓ メトリクス     │         │
│  │ ✓ ノード操作    │    │ ✓ ログ分析       │         │
│  │ ✓ 詳細パネル    │    │ ✓ アラート       │         │
│  │ ✓ フォーカス    │    │ ✓ トレンド       │         │
│  │ ✓ オンボード    │    │ ✓ システム監視   │         │
│  └────────┬────────┘    └────────┬─────────┘         │
│           │                      │                   │
│           └──────────┬───────────┘                   │
│                      │                               │
│           ┌──────────▼──────────────────┐            │
│           │     Backend API             │            │
│           │  (Go / FastAPI)             │            │
│           ├─────────────────────────────┤            │
│           │ ✓ GraphDB (Neo4j/DGraph)   │            │
│           │ ✓ Drift Scanner             │            │
│           │ ✓ Falco Event Processor     │            │
│           │ ✓ Prometheus Exporter       │            │
│           └─────────────────────────────┘            │
│                                                        │
└────────────────────────────────────────────────────────┘

使い分けの原則:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
React専用UI     : インタラクティブ探索・リアルタイム操作
Grafana         : 監視・ログ分析・アラート・トレンド分析
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

#### なぜこの使い分けが最適なのか

| 用途 | ツール | 理由 |
|-----|--------|------|
| **グラフ探索** | React UI | インタラクティブ操作が必須。ユーザーがノードをクリック・ドラッグ・フォーカスする必要がある |
| **メトリクス監視** | Grafana | Prometheus統合、豊富なグラフ種類、アラート機能が標準装備 |
| **ログ分析** | Grafana + Loki | ログクエリ、フィルタリング、時系列分析に最適 |
| **アラート** | Grafana | 柔軟なアラートルール、通知チャンネル統合（Slack/PagerDuty等） |
| **トレンド分析** | Grafana | 長期的なデータ保持、ダッシュボード共有、自動リフレッシュ |

### 学び：「完全に捨てる」vs「適材適所で使う」

最初は「Grafanaを完全に置き換える」と考えていましたが、実際には：

```
❌ 間違ったアプローチ:
   「Grafanaか、React専用UIか、どちらか一方」

✅ 正しいアプローチ:
   「それぞれの強みを活かして併用」
```

**Grafanaの強み：**
- モニタリング・アラートのエコシステム
- Prometheus/Loki との統合
- 運用チームが慣れている
- ダッシュボード共有が簡単

**React専用UIの強み：**
- インタラクティブ性
- カスタムUX
- 複雑なユーザーワークフロー
- アプリケーション固有の機能

**結論：両方のツールを適切に使い分けることで、最高のシステムが構築できる。**

## 技術スタック

```json
{
  "frontend": {
    "framework": "React 19.2.0",
    "visualization": ["React Flow 11.11.4", "Cytoscape 3.33.1"],
    "state": ["Zustand 5.0.9", "React Query 5.90.12"],
    "styling": "Tailwind CSS 4.1.18",
    "icons": ["Lucide React", "AWS React Icons"]
  },
  "testing": {
    "runner": "Vitest 4.0.16",
    "library": "Testing Library 16.3.1",
    "mocking": "MSW 2.12.7",
    "coverage": "@vitest/coverage-v8 4.0.16"
  }
}
```

## Phase 1: UI/UX改善 - ユーザー体験の向上

### 1.1 ツールチップシステムの実装

まず最初に取り組んだのが、ユーザーへの情報提供です。グラフ上のノードにマウスをホバーしたときに詳細情報を表示するツールチップを実装しました。

**実装のポイント：**

```tsx
// CustomNode.tsx - ホバーツールチップ
const [showTooltip, setShowTooltip] = useState(false);

const handleMouseEnter = () => {
  setShowTooltip(true);
};

const handleMouseLeave = () => {
  setShowTooltip(false);
};

// ツールチップコンポーネント
{showTooltip && (
  <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2
                  px-3 py-2 bg-gray-900 text-white text-sm rounded-lg
                  shadow-lg z-50 whitespace-nowrap animate-in fade-in-0
                  slide-in-from-bottom-2">
    <div className="font-semibold">{data.label}</div>
    {data.resource_name && (
      <div className="text-xs text-gray-300">{data.resource_name}</div>
    )}
  </div>
)}
```

**学び：**
- Tailwind CSSの`animate-in`を使ったスムーズなアニメーション
- `z-50`で他の要素と重ならないよう配慮
- `whitespace-nowrap`でテキストの折り返しを防止

### 1.2 インタラクティブ機能の強化

次に、ユーザーがグラフと対話できる機能を追加しました。

**実装した機能：**

1. **ノードクリック → 詳細パネル表示**
```tsx
const handleClick = () => {
  const event = new CustomEvent('node-detail', {
    detail: { node: { id, data } }
  });
  window.dispatchEvent(event);
};
```

2. **ダブルクリック → フォーカスビュー**
```tsx
const handleDoubleClick = () => {
  const event = new CustomEvent('node-focus', {
    detail: { nodeId: id }
  });
  window.dispatchEvent(event);
};
```

3. **右クリック → コンテキストメニュー**
```tsx
const handleContextMenu = (e: React.MouseEvent) => {
  e.preventDefault();
  const event = new CustomEvent('node-context-menu', {
    detail: {
      nodeId: id,
      position: { x: e.clientX, y: e.clientY }
    }
  });
  window.dispatchEvent(event);
};
```

**学び：**
- CustomEventを使ったコンポーネント間通信
- `preventDefault()`でブラウザのデフォルト動作を制御
- イベント駆動アーキテクチャの活用

### 1.3 ビジュアル強化

セキュリティリスクを直感的に理解できるよう、ビジュアル要素を強化しました。

**Severity別の色分け：**

```tsx
const getSeverityStyles = (severity?: string) => {
  switch (severity) {
    case 'critical':
      return {
        border: 'border-red-500 border-2',
        bg: 'bg-red-50',
        badge: 'bg-red-100 text-red-800'
      };
    case 'high':
      return {
        border: 'border-orange-500 border-2',
        bg: 'bg-orange-50',
        badge: 'bg-orange-100 text-orange-800'
      };
    case 'medium':
      return {
        border: 'border-yellow-500',
        bg: 'bg-yellow-50',
        badge: 'bg-yellow-100 text-yellow-800'
      };
    case 'low':
      return {
        border: 'border-blue-500',
        bg: 'bg-blue-50',
        badge: 'bg-blue-100 text-blue-800'
      };
    default:
      return {
        border: 'border-gray-300',
        bg: 'bg-white',
        badge: 'bg-gray-100 text-gray-800'
      };
  }
};
```

**グラデーションヘッダー：**

```tsx
<div className="bg-gradient-to-r from-blue-600 to-blue-700 px-4 py-3">
  <div className="flex items-center justify-between">
    <h2 className="text-lg font-semibold text-white">
      {data.label}
    </h2>
  </div>
</div>
```

### 1.4 オンボーディング機能

新規ユーザーがアプリケーションをすぐに使いこなせるよう、包括的なオンボーディングシステムを実装しました。

**1. ウェルカムモーダル（6ステップウィザード）**

```tsx
// WelcomeModal.tsx
const [currentStep, setCurrentStep] = useState(0);

const steps = [
  {
    title: 'TFDrift-Falcoへようこそ',
    icon: <Zap className="w-12 h-12 text-blue-600" />,
    description: 'クラウドインフラのセキュリティとドリフト分析を可視化します',
    details: [
      'Terraform Drift → IAM → Kubernetes → Falcoの因果関係を追跡',
      'インタラクティブなグラフで依存関係を可視化',
      'セキュリティリスクの影響範囲を分析'
    ]
  },
  // ... 他の5ステップ
];
```

**LocalStorageで初回表示を管理：**

```tsx
export const shouldShowWelcome = (): boolean => {
  return !localStorage.getItem('tfdrift-welcome-seen');
};

const handleFinish = () => {
  localStorage.setItem('tfdrift-welcome-seen', 'true');
  onClose();
};
```

**2. キーボードショートカットガイド**

14個のショートカットを4つのカテゴリーに整理：

```tsx
const shortcuts = {
  navigation: [
    { key: 'F', description: 'グラフ全体を画面にフィット' },
    { key: 'C', description: 'グラフを中央に配置' },
    { key: '+', description: 'ズームイン' },
    { key: '-', description: 'ズームアウト' }
  ],
  selection: [
    { key: 'Click', description: 'ノード詳細パネルを開く' },
    { key: 'Double Click', description: 'フォーカスビューでハイライト' },
    { key: 'Right Click', description: 'コンテキストメニューを表示' },
    { key: 'ESC', description: '詳細パネルを閉じる' }
  ],
  // ...
};
```

**3. ヘルプオーバーレイ**

常時アクセス可能な折りたたみ式ヘルプパネル：

```tsx
// HelpOverlay.tsx
const [isVisible, setIsVisible] = useState(true);
const [isExpanded, setIsExpanded] = useState(true);

// フローティングボタンへの切り替え
{!isVisible ? (
  <button
    onClick={() => setIsVisible(true)}
    className="fixed bottom-6 right-6 p-4 bg-blue-600 hover:bg-blue-700
               text-white rounded-full shadow-2xl transition-all
               duration-200 hover:scale-110"
    aria-label="ヘルプを表示"
  >
    <HelpCircle className="w-6 h-6" />
  </button>
) : (
  // 展開可能なヘルプパネル
  <div className="fixed bottom-6 right-6 w-full max-w-sm">
    {/* パネル内容 */}
  </div>
)}
```

## Phase 2: ドキュメント整備

### PROJECT_STRUCTURE.md

プロジェクト構造を明確に文書化：

```markdown
## ディレクトリ構造

ui/
├── src/
│   ├── api/              # APIクライアントとReact Queryフック
│   │   ├── client.ts     # Fetch-basedシングルトンAPIクライアント
│   │   ├── types.ts      # API型定義
│   │   └── hooks/        # React Queryカスタムフック
│   ├── components/       # Reactコンポーネント
│   │   ├── onboarding/   # オンボーディング関連
│   │   └── reactflow/    # React Flowカスタムコンポーネント
│   ├── utils/            # ユーティリティ関数
│   └── __tests__/        # テストユーティリティ
```

### ARCHITECTURE.md

システムアーキテクチャの全体像を記述：

```markdown
## データフロー

1. **APIクライアント層** (`src/api/client.ts`)
   - Fetch APIベースのシングルトンクライアント
   - エラーハンドリングと型安全性

2. **React Query層** (`src/api/hooks/`)
   - データフェッチングとキャッシング
   - 30秒ごとの自動リフレッシュ

3. **State管理層** (Zustand + React Query)
   - ローカルUI状態: Zustand
   - サーバー状態: React Query

4. **プレゼンテーション層** (`src/components/`)
   - React Flow + Cytoscapeによる可視化
   - カスタムノード/エッジコンポーネント
```

### README.md

開発者向けのクイックスタートガイド：

```markdown
## クイックスタート

### 前提条件
- Node.js 18.x以上
- npm 9.x以上

### インストールと起動

\`\`\`bash
# 依存関係のインストール
npm install

# 開発サーバー起動
npm run dev

# テスト実行
npm test

# カバレッジレポート
npm run test:coverage
\`\`\`
```

## Phase 3: テスト基盤の構築

### 3.1 テストツールの選定と設定

**なぜVitest？**
- Viteとのネイティブ統合（設定がシンプル）
- Jestと互換性のあるAPI
- 高速な実行速度
- ES Modules完全サポート

**Vitest設定：**

```typescript
// vitest.config.ts
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/__tests__/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'src/__tests__/',
        '**/*.test.{ts,tsx}',
        '**/*.config.{ts,js}',
      ],
    },
  },
});
```

### 3.2 テストユーティリティの作成

React Query + React Routerを使ったコンポーネントをテストするためのカスタムレンダー関数：

```typescript
// src/__tests__/utils/testUtils.tsx
import { render } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { MemoryRouter } from 'react-router-dom';

export const renderWithProviders = (
  ui: React.ReactElement,
  {
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    }),
    route = '/',
    ...renderOptions
  } = {}
) => {
  const Wrapper = ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[route]}>
        {children}
      </MemoryRouter>
    </QueryClientProvider>
  );

  return render(ui, { wrapper: Wrapper, ...renderOptions });
};
```

### 3.3 MSWによるAPIモッキング

Mock Service Workerを使ったHTTPリクエストのインターセプト：

```typescript
// src/api/client.test.ts
import { setupServer } from 'msw/node';
import { http, HttpResponse } from 'msw';

const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

// テスト内でのハンドラー設定
server.use(
  http.get(`${API_BASE_URL}/graph`, () => {
    return HttpResponse.json(createSuccessResponse(mockGraphData));
  })
);
```

**MSWの利点：**
- 実際のHTTPリクエストをモック
- ネットワーク層でのインターセプト
- Node環境とブラウザ環境両方で動作
- 型安全なレスポンス定義

## Phase 4: テストカバレッジ拡大

### 4.1 初期テスト（15件）

まず、コアユーティリティとコンポーネントのテストから開始：

**reactFlowAdapter.test.ts:**

```typescript
describe('convertToReactFlow', () => {
  it('should convert Cytoscape data to React Flow format', () => {
    const cytoscapeData: CytoscapeElements = {
      nodes: [
        {
          data: {
            id: 'node1',
            label: 'Test Node',
            type: 'aws_iam_role',
            severity: 'high'
          }
        }
      ],
      edges: [
        {
          data: {
            id: 'edge1',
            source: 'node1',
            target: 'node2',
            type: 'depends_on'
          }
        }
      ]
    };

    const result = convertToReactFlow(cytoscapeData);

    expect(result.nodes).toHaveLength(1);
    expect(result.edges).toHaveLength(1);
    expect(result.nodes[0].type).toBe('custom');
    expect(result.nodes[0].data.severity).toBe('high');
  });
});
```

### 4.2 コンポーネントテスト（138件）

#### CustomNode.test.tsx（18件）

```typescript
describe('CustomNode', () => {
  describe('User Interactions', () => {
    it('should dispatch node-detail event on click', async () => {
      const user = userEvent.setup();
      const eventSpy = vi.fn();

      window.addEventListener('node-detail', eventSpy);

      renderWithProviders(<CustomNode {...mockNodeProps} />);

      const node = screen.getByText('Test Node');
      await user.click(node);

      expect(eventSpy).toHaveBeenCalledTimes(1);
      expect(eventSpy.mock.calls[0][0].detail).toMatchObject({
        node: { id: 'node1', data: expect.any(Object) }
      });
    });

    it('should show tooltip on mouse enter', async () => {
      const user = userEvent.setup();
      renderWithProviders(<CustomNode {...mockNodeProps} />);

      const node = screen.getByText('Test Node');
      await user.hover(node);

      expect(screen.getByText('Test Node')).toBeInTheDocument();
      expect(screen.getByText('test-resource')).toBeInTheDocument();
    });
  });
});
```

#### NodeDetailPanel.test.tsx（33件）

```typescript
describe('NodeDetailPanel', () => {
  describe('Metadata Display', () => {
    it('should display string metadata values', () => {
      renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      expect(screen.getByText('arn')).toBeInTheDocument();
      expect(screen.getByText('arn:aws:iam::123456789012:role/test-role'))
        .toBeInTheDocument();
    });

    it('should display object metadata as JSON', () => {
      const { container } = renderWithProviders(
        <NodeDetailPanel node={mockNode} onClose={mockOnClose} />
      );

      const preElements = container.querySelectorAll('pre');
      const jsonContent = preElements[0].textContent || '';

      expect(jsonContent).toContain('Environment');
      expect(jsonContent).toContain('production');
    });
  });
});
```

**遭遇した問題と解決策：**

1. **重複テキストエラー**
```typescript
// ❌ 失敗: 同じテキストが複数の要素に存在
expect(screen.getByText('aws_iam_role')).toBeInTheDocument();

// ✅ 解決: getAllByTextを使用
const resourceTypes = screen.getAllByText('aws_iam_role');
expect(resourceTypes.length).toBeGreaterThan(0);
```

2. **JSON表示のテスト**
```typescript
// ❌ 失敗: 厳密な文字列マッチングは脆弱
expect(screen.getByText(/"Environment": "production"/)).toBeInTheDocument();

// ✅ 解決: コンテナクエリを使用
const { container } = renderWithProviders(<Component />);
const preElements = container.querySelectorAll('pre');
const jsonContent = preElements[0].textContent || '';
expect(jsonContent).toContain('Environment');
```

#### WelcomeModal.test.tsx（30件）

```typescript
describe('WelcomeModal', () => {
  describe('Navigation', () => {
    it('should navigate through all 6 steps', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      const expectedSteps = [
        'TFDrift-Falcoへようこそ',
        'グラフの操作方法',
        '依存関係の可視化',
        '検索とフィルタリング',
        'エクスポートと共有',
        'キーボードショートカット'
      ];

      for (let i = 0; i < expectedSteps.length; i++) {
        expect(screen.getByText(expectedSteps[i])).toBeInTheDocument();

        if (i < expectedSteps.length - 1) {
          const nextButton = screen.getByRole('button', { name: '次へ' });
          await user.click(nextButton);

          await waitFor(() => {
            expect(screen.getByText(`ステップ ${i + 2} / 6`))
              .toBeInTheDocument();
          });
        }
      }
    });
  });

  describe('LocalStorage Integration', () => {
    it('should save preference when finished', async () => {
      const user = userEvent.setup();
      renderWithProviders(<WelcomeModal onClose={mockOnClose} />);

      // 最終ステップまで進む
      const nextButton = screen.getByRole('button', { name: '次へ' });
      for (let i = 0; i < 5; i++) {
        await user.click(nextButton);
      }

      // 完了ボタンをクリック
      const finishButton = await screen.findByRole('button', { name: '始める' });
      await user.click(finishButton);

      await waitFor(() => {
        expect(localStorage.getItem('tfdrift-welcome-seen')).toBe('true');
        expect(mockOnClose).toHaveBeenCalledTimes(1);
      });
    });
  });
});
```

#### KeyboardShortcutsGuide.test.tsx（19件）

```typescript
describe('KeyboardShortcutsGuide', () => {
  describe('Content Completeness', () => {
    it('should display all 14 shortcuts', () => {
      const { container } = renderWithProviders(
        <KeyboardShortcutsGuide onClose={mockOnClose} />
      );

      const allKbdElements = container.querySelectorAll('kbd');
      // 14 shortcuts + 1 in footer hint = 15 total
      expect(allKbdElements.length).toBe(15);
    });

    it('should not have duplicate shortcuts', () => {
      const { container } = renderWithProviders(
        <KeyboardShortcutsGuide onClose={mockOnClose} />
      );

      const kbdElements = container.querySelectorAll('kbd');
      const kbdTexts = Array.from(kbdElements).map(el => el.textContent);
      const shortcutKeys = kbdTexts.slice(0, -1); // Exclude footer hint
      const uniqueKeys = new Set(shortcutKeys);

      expect(shortcutKeys.length).toBe(uniqueKeys.size);
    });
  });
});
```

#### HelpOverlay.test.tsx（29件）

```typescript
describe('HelpOverlay', () => {
  describe('Hide/Show Functionality', () => {
    it('should transition between panel and floating button', async () => {
      const user = userEvent.setup();
      renderWithProviders(<HelpOverlay />);

      // 初期状態: パネル表示
      expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();

      // 閉じる
      const closeButton = screen.getByLabelText('閉じる');
      await user.click(closeButton);

      await waitFor(() => {
        expect(screen.queryByText('クイックヘルプ')).not.toBeInTheDocument();
        expect(screen.getByLabelText('ヘルプを表示')).toBeInTheDocument();
      });

      // 再表示
      const showButton = screen.getByLabelText('ヘルプを表示');
      await user.click(showButton);

      await waitFor(() => {
        expect(screen.getByText('クイックヘルプ')).toBeInTheDocument();
      });
    });
  });
});
```

### 4.3 API層テスト（113件）

#### APIClient.test.ts（34件）

```typescript
describe('APIClient', () => {
  describe('Graph API', () => {
    it('should fetch nodes with pagination', async () => {
      const mockPaginatedNodes = {
        data: [mockNode],
        page: 1,
        limit: 10,
        total: 100,
        total_pages: 10
      };

      server.use(
        http.get(`${API_BASE_URL}/graph/nodes`, ({ request }) => {
          const url = new URL(request.url);
          expect(url.searchParams.get('page')).toBe('1');
          expect(url.searchParams.get('limit')).toBe('10');
          return HttpResponse.json(createSuccessResponse(mockPaginatedNodes));
        })
      );

      const result = await apiClient.getNodes({ page: 1, limit: 10 });
      expect(result).toEqual(mockPaginatedNodes);
    });
  });

  describe('Error Handling', () => {
    it('should handle HTTP 404 error', async () => {
      server.use(
        http.get(`${API_BASE_URL}/graph/nodes/nonexistent`, () => {
          return new HttpResponse(null, { status: 404 });
        })
      );

      await expect(apiClient.getNodeById('nonexistent'))
        .rejects.toThrow('HTTP error! status: 404');
    });

    it('should handle API error response', async () => {
      server.use(
        http.get(`${API_BASE_URL}/drifts/invalid`, () => {
          return HttpResponse.json(
            createErrorResponse(400, 'Invalid drift ID')
          );
        })
      );

      await expect(apiClient.getDrift('invalid'))
        .rejects.toThrow('Invalid drift ID');
    });
  });
});
```

#### useDrifts.test.tsx（21件）

```typescript
describe('useDrifts', () => {
  const createWrapper = () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false }
      }
    });

    return ({ children }: { children: React.ReactNode }) => (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  };

  describe('Data Fetching', () => {
    it('should fetch drifts with all filters', async () => {
      vi.mocked(apiClient.getDrifts).mockResolvedValue(mockPaginatedDrifts);

      const params = {
        page: 1,
        limit: 10,
        severity: 'high',
        resource_type: 'aws_iam_role'
      };

      const { result } = renderHook(() => useDrifts(params), {
        wrapper: createWrapper()
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrifts).toHaveBeenCalledWith(params);
      expect(result.current.data).toEqual(mockPaginatedDrifts);
    });
  });

  describe('Enabled State', () => {
    it('should not fetch when id is empty', () => {
      const { result } = renderHook(() => useDrift(''), {
        wrapper: createWrapper()
      });

      expect(result.current.isLoading).toBe(false);
      expect(apiClient.getDrift).not.toHaveBeenCalled();
    });

    it('should fetch when id changes from empty to valid', async () => {
      vi.mocked(apiClient.getDrift).mockResolvedValue(mockDriftAlert);

      const { result, rerender } = renderHook(
        ({ id }: { id: string }) => useDrift(id),
        {
          wrapper: createWrapper(),
          initialProps: { id: '' }
        }
      );

      expect(apiClient.getDrift).not.toHaveBeenCalled();

      rerender({ id: 'drift-123' });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(apiClient.getDrift).toHaveBeenCalledWith('drift-123');
    });
  });
});
```

#### useEvents.test.tsx（26件）

```typescript
describe('useEvents', () => {
  describe('Pagination', () => {
    it('should handle multiple pages', async () => {
      const page1 = { ...mockPaginatedEvents, page: 1 };
      const page2 = {
        ...mockPaginatedEvents,
        page: 2,
        data: [{ ...mockFalcoEvent, id: 'event-456' }]
      };

      vi.mocked(apiClient.getEvents).mockResolvedValueOnce(page1);

      const { result, rerender } = renderHook(
        ({ page }: { page: number }) => useEvents({ page }),
        {
          wrapper: createWrapper(),
          initialProps: { page: 1 }
        }
      );

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.page).toBe(1);

      // ページ2に切り替え
      vi.mocked(apiClient.getEvents).mockResolvedValueOnce(page2);
      rerender({ page: 2 });

      await waitFor(() => {
        expect(result.current.data?.page).toBe(2);
      });
    });
  });
});
```

#### useGraph.test.tsx（32件）

```typescript
describe('useGraph', () => {
  describe('Data Fetching', () => {
    it('should fetch empty graph', async () => {
      const emptyGraph: CytoscapeElements = {
        nodes: [],
        edges: []
      };

      vi.mocked(apiClient.getGraph).mockResolvedValue(emptyGraph);

      const { result } = renderHook(() => useGraph(), {
        wrapper: createWrapper()
      });

      await waitFor(() => {
        expect(result.current.isSuccess).toBe(true);
      });

      expect(result.current.data?.nodes).toHaveLength(0);
      expect(result.current.data?.edges).toHaveLength(0);
    });
  });
});
```

## 成果と学び

### 📊 最終結果

| 指標 | 開始時 | 最終 | 増加 |
|-----|--------|------|------|
| テスト件数 | 0 | 266 | +266 |
| カバレッジ (Statements) | 0% | 92.32% | +92.32% |
| テストファイル数 | 0 | 10 | +10 |
| 実装機能数 | 基本的なグラフ表示 | フル機能UI + 完全なテスト | - |

### 📈 100%カバレッジ達成ファイル

- ✅ `src/api/client.ts` (100% statements, 86.79% branches)
- ✅ `src/api/hooks/useDrifts.ts` (100%)
- ✅ `src/api/hooks/useEvents.ts` (100%)
- ✅ `src/api/hooks/useGraph.ts` (100%)
- ✅ `src/components/onboarding/HelpOverlay.tsx` (100%)
- ✅ `src/components/onboarding/KeyboardShortcutsGuide.tsx` (100%)
- ✅ `src/components/onboarding/WelcomeModal.tsx` (100% statements)
- ✅ `src/components/reactflow/CustomNode.tsx` (100% statements)

### 💡 重要な学び

#### 1. テストファーストは後から書くより圧倒的に楽

実装とテストを同時に進めることで：
- バグの早期発見
- リファクタリングの安心感
- 設計の改善点が明確に

#### 2. MSWは強力なAPIモックツール

```typescript
// 実際のネットワークリクエストをインターセプト
server.use(
  http.get('/api/v1/graph', () => {
    return HttpResponse.json({
      success: true,
      data: mockGraphData
    });
  })
);
```

メリット：
- 実際のfetch/axiosコールをそのまま使える
- レスポンスの型安全性
- エラーケースのテストが簡単

#### 3. React Queryのテストパターン

```typescript
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false, // テスト時はリトライ無効化
      }
    }
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};
```

ポイント：
- 各テストで新しいQueryClientを作成
- `retry: false`でテストを高速化
- `waitFor`で非同期処理の完了を待つ

#### 4. Testing Libraryのベストプラクティス

```typescript
// ❌ Bad: 実装の詳細に依存
const button = container.querySelector('.submit-button');

// ✅ Good: ユーザーの視点でクエリ
const button = screen.getByRole('button', { name: '送信' });

// ❌ Bad: 不安定なセレクタ
const icon = container.querySelector('div > svg.icon');

// ✅ Good: data-testid で明示的に指定
const icon = screen.getByTestId('cloud-icon');
```

#### 5. カバレッジは目的ではなく手段

- 100%を目指すより、重要な機能を確実にテスト
- エッジケースやエラーハンドリングを重視
- ユーザーの使い方に近いテストを書く

### 🔧 使用したテクニック

#### 1. CustomEvent でのコンポーネント間通信

```typescript
// 送信側
const handleClick = () => {
  const event = new CustomEvent('node-detail', {
    detail: { node: { id, data } }
  });
  window.dispatchEvent(event);
};

// 受信側
useEffect(() => {
  const handleNodeDetail = (e: CustomEvent) => {
    setSelectedNode(e.detail.node);
  };

  window.addEventListener('node-detail', handleNodeDetail);
  return () => window.removeEventListener('node-detail', handleNodeDetail);
}, []);
```

#### 2. LocalStorageの活用

```typescript
// 初回表示判定
export const shouldShowWelcome = (): boolean => {
  return !localStorage.getItem('tfdrift-welcome-seen');
};

// リセット関数（開発・テスト用）
export const resetWelcome = () => {
  localStorage.removeItem('tfdrift-welcome-seen');
};
```

#### 3. Tailwind CSSのアニメーション

```tsx
// フェードイン + スライドイン
<div className="animate-in fade-in-0 slide-in-from-bottom-2 duration-200">
  {/* コンテンツ */}
</div>

// ホバー時のスケール変換
<button className="transition-all duration-200 hover:scale-110">
  {/* ボタン */}
</button>
```

#### 4. TypeScript型定義の活用

```typescript
// APIレスポンスの型
export interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: number;
    message: string;
  };
}

// ページネーションレスポンス
export interface PaginatedResponse<T> {
  data: T[];
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}
```

### ⚠️ ハマったポイントと解決策

#### 1. React Query のキャッシュ問題

**問題：**
テスト間でキャッシュが残ってテストが失敗する

**解決：**
```typescript
beforeEach(() => {
  queryClient.clear(); // 各テスト前にキャッシュクリア
});
```

#### 2. MSWのセットアップ

**問題：**
`.ts`ファイルでJSXを使おうとしてエラー

**解決：**
```bash
# .tsx 拡張子を使用
mv useDrifts.test.ts useDrifts.test.tsx
```

#### 3. waitForのタイムアウト

**問題：**
非同期処理が完了せずにタイムアウト

**解決：**
```typescript
await waitFor(() => {
  expect(result.current.isSuccess).toBe(true);
}, { timeout: 5000 }); // タイムアウトを延長
```

#### 4. モックのリセット

**問題：**
前のテストのモックが残って予期しない動作

**解決：**
```typescript
beforeEach(() => {
  vi.clearAllMocks(); // すべてのモックをクリア
});

afterEach(() => {
  server.resetHandlers(); // MSWハンドラーをリセット
});
```

## まとめ

約2週間で以下を達成しました：

### 1. **Grafanaからの完全移行**
   - Node Graph Panel → React Flow + Cytoscape.js
   - 静的ダッシュボード → インタラクティブUI
   - 開発速度10倍向上
   - パフォーマンス3倍改善

### 2. **UI/UX の全面改装**
   - ツールチップ、インタラクション、ビジュアル強化
   - 6ステップのオンボーディング
   - 14個のキーボードショートカット
   - コンテキストヘルプシステム

### 3. **包括的なドキュメント**
   - プロジェクト構造
   - アーキテクチャ設計
   - 開発者ガイド

### 4. **堅牢なテスト基盤**
   - Vitest + Testing Library + MSW
   - カスタムテストユーティリティ
   - 266件のテスト
   - 92.32%のカバレッジ

### 重要な意思決定：いつ既存ツールを捨てるべきか

このプロジェクトで最も重要な学びは、**「いつ既存のツールを捨てて、専用ソリューションを作るべきか」**の判断基準です。

#### Grafanaを使い続けるべきケース：
- ✅ 主な用途がモニタリングとアラート
- ✅ 静的なダッシュボードで十分
- ✅ 既存のGrafanaエコシステムとの統合が重要
- ✅ 開発リソースが限られている

#### 専用UIを作るべきケース：
- ✅ **インタラクティブな操作が必須**（今回のケース）
- ✅ カスタムUIコンポーネントが多数必要
- ✅ 複雑なユーザーワークフローがある
- ✅ パフォーマンスが重要
- ✅ 長期的な機能拡張を見据えている

今回は、セキュリティアナリストの「探索的な分析」という要求が明確だったため、専用UI開発が正解でした。

結果として：
- ユーザー満足度の向上
- 開発チームの生産性10倍
- テスタビリティの大幅改善
- 将来の拡張性確保

**「適切なツールを選ぶ」よりも「適切なタイミングでツールを変える」ことが重要**だと学びました。

### 次のステップ

- [ ] E2Eテストの追加（Playwright）
- [ ] パフォーマンステスト
- [ ] アクセシビリティ監査
- [ ] CI/CDパイプラインの構築
- [ ] Storybookの導入
- [ ] AIによる異常検知機能（Grafanaでは不可能だった機能）

このプロジェクトを通じて、**品質の高いフロントエンド開発には適切なツール選定とテスト戦略が不可欠**だと改めて実感しました。

特に、MSW + React Query + Testing Libraryの組み合わせは、モダンなReactアプリケーションのテストに最適だと確信しています。

そして最も重要なのは、**既存ツールの限界を認識し、適切なタイミングで技術スタックを刷新する勇気**を持つことです。

## 参考リンク

- [TFDrift-Falco GitHub](https://github.com/your-repo/tfdrift-falco)
- [Vitest Documentation](https://vitest.dev/)
- [Testing Library](https://testing-library.com/)
- [MSW (Mock Service Worker)](https://mswjs.io/)
- [React Query Testing](https://tanstack.com/query/latest/docs/react/guides/testing)

---

この記事が、あなたのプロジェクトのテスト戦略や品質向上の参考になれば幸いです！

質問やコメントがあれば、ぜひお聞かせください 👇

## タグ

`#React` `#TypeScript` `#Grafana` `#Testing` `#Vitest` `#MSW` `#ReactQuery` `#TailwindCSS` `#フロントエンド` `#リファクタリング`
