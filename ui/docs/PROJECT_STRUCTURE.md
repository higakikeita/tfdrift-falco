# TFDrift-Falco UI - Project Structure

> **MECE原則に基づく体系的なプロジェクト構造ドキュメント**

## 📊 ディレクトリ構造（MECE分類）

### 1. Application Layer（アプリケーション層）

#### 1.1 Entry Points
```
src/
├── main.tsx                 # アプリケーションエントリーポイント
├── App-with-table.tsx       # 【本番環境】分割ビュー（グラフ + テーブル）
└── App-final.tsx            # 【推奨】最新版（フィルター + オンボーディング）
```

**Deprecated（非推奨）:**
- `App.tsx` - 旧グラフのみビュー
- `App-v2.tsx` - 再設計版
- `App-minimal.tsx` - 最小デモ
- `App-test.tsx` - テスト用

#### 1.2 Routing & State
```
src/
└── lib/
    └── queryClient.ts       # React Query設定
```

---

### 2. Data Layer（データ層）

#### 2.1 API Integration
```
src/api/
├── client.ts                # RESTful APIクライアント
├── websocket.ts             # WebSocketクライアント（リアルタイム更新）
├── sse.ts                   # Server-Sent Events（ストリーミング）
└── types.ts                 # APIレスポンス型定義
```

**役割分担:**
- **REST (client.ts)**: 初期データ取得、CRUD操作
- **WebSocket (websocket.ts)**: 双方向リアルタイム通信
- **SSE (sse.ts)**: サーバープッシュ型イベントストリーム

#### 2.2 React Query Hooks
```
src/api/hooks/
├── index.ts                 # フック一括エクスポート
├── useGraph.ts              # グラフデータ取得
├── useDrifts.ts             # ドリフトアラート取得
├── useEvents.ts             # Falcoイベント取得
├── useState.ts              # Terraformステート取得
├── useStats.ts              # 統計情報取得
└── useGraphDB.ts            # GraphDBクエリ（影響範囲、依存関係等）
```

**データフロー:**
```
Backend API → API Client → React Query Hooks → Components
```

---

### 3. Presentation Layer（プレゼンテーション層）

#### 3.1 Component Hierarchy（MECE）

##### A. Container Components（コンテナ）
```
src/
├── App-with-table.tsx       # メインコンテナ（グラフ + テーブル）
└── App-final.tsx            # 最新メインコンテナ（フィルター付き）
```

##### B. Graph Visualization（グラフ可視化）

**React Flow系（推奨）:**
```
src/components/reactflow/
├── ReactFlowGraph.tsx       # メインラッパー（エクスポート機能付き）
├── CustomNode.tsx           # カスタムノード（クラウドアイコン）
├── HierarchicalNodes.tsx    # AWS階層ノード（Region/VPC/AZ/Subnet）
├── NodeDetailPanel.tsx      # ノード詳細パネル
├── ClusterNode.tsx          # クラスターノード
├── LODNode.tsx              # LOD最適化ノード
└── OptimizedGraph.tsx       # パフォーマンス最適化版
```

**Cytoscape系（レガシー）:**
```
src/components/
├── CytoscapeGraph.tsx       # Cytoscapeラッパー
└── GraphWithIcons.tsx       # アイコンオーバーレイ版
```

##### C. Data Display（データ表示）
```
src/components/
├── DriftHistoryTable.tsx    # ドリフト履歴テーブル（ソート・フィルター）
├── DriftDetailPanel.tsx     # ドリフト詳細サイドバー
├── NodeDetailPanel.tsx      # ノード詳細パネル（トップレベル）
├── PatternSearchPanel.tsx   # グラフパターン検索UI
└── ConnectionStatus.tsx     # 接続ステータスインジケーター
```

##### D. Interactive Elements（インタラクティブ要素）
```
src/components/graph/
├── NodeTooltip.tsx          # ノードホバーツールチップ
└── NodeContextMenu.tsx      # ノード右クリックメニュー
```

##### E. UI Components（shadcn/ui）
```
src/components/ui/
├── button.tsx               # ボタンコンポーネント
├── card.tsx                 # カードコンポーネント
└── tabs.tsx                 # タブコンポーネント
```

##### F. Icon Components（アイコン）
```
src/components/icons/
├── OfficialCloudIcons.tsx   # 【推奨】公式プロバイダーアイコン
├── ProviderIcons.tsx        # プロバイダー別アイコン
├── AWSServiceIcons.tsx      # AWSサービスアイコン（旧）
├── GCPServiceIcons.tsx      # GCPサービスアイコン（旧）
└── K8sAndSpecialIcons.tsx   # Kubernetes特殊アイコン（旧）
```

##### G. Onboarding（オンボーディング）
```
src/components/onboarding/
├── WelcomeModal.tsx         # 初回ユーザーウェルカム
├── HelpOverlay.tsx          # コンテキストヘルプ
└── KeyboardShortcutsGuide.tsx # キーボードショートカット
```

##### H. Utility Components（ユーティリティ）
```
src/components/
└── ThemeToggle.tsx          # ダーク/ライトモード切替
```

**合計: 37コンポーネント**

---

### 4. Business Logic Layer（ビジネスロジック層）

#### 4.1 Custom Hooks
```
src/hooks/
├── useProgressiveGraph.ts   # プログレッシブグラフレンダリング
└── useTheme.ts              # テーマ管理
```

#### 4.2 Utilities
```
src/utils/
├── reactFlowAdapter.ts      # Cytoscape → React Flow変換
├── graphClustering.ts       # グラフクラスタリングアルゴリズム
├── memoryOptimization.ts    # メモリ最適化
├── sampleData.ts            # サンプルグラフデータ生成器
└── sampleDrifts.ts          # サンプルドリフトデータ生成器
```

---

### 5. Type Layer（型定義層）

```
src/types/
├── graph.ts                 # グラフ関連型（NodeType, EdgeType, LayoutType）
└── drift.ts                 # ドリフトイベント型（DriftEvent, DriftSeverity）

src/api/
└── types.ts                 # APIレスポンス型（CytoscapeNode, DriftAlert, Stats）
```

---

### 6. Style Layer（スタイル層）

```
src/styles/
├── cytoscapeStyles.ts       # Cytoscapeグラフスタイル
└── enhanced-cytoscape-styles.ts # 拡張スタイル

src/
├── index.css                # グローバルTailwind CSS
└── App.css                  # カスタムアプリケーションスタイル
```

---

### 7. Configuration Layer（設定層）

```
src/lib/
└── utils.ts                 # Tailwind cn()ヘルパー

Root:
├── tailwind.config.js       # Tailwind設定
├── tsconfig.json            # TypeScript設定
├── vite.config.ts           # Vite設定
└── package.json             # 依存関係
```

---

## 🔄 データフロー図（MECE）

### A. 初期ロード
```
1. main.tsx
   ↓
2. App-with-table.tsx / App-final.tsx
   ↓
3. React Query Hooks (useGraph, useDrifts, etc.)
   ↓
4. API Client (client.ts)
   ↓
5. Backend API
   ↓
6. Components (ReactFlowGraph, DriftHistoryTable)
```

### B. リアルタイム更新
```
1. websocket.ts / sse.ts (接続確立)
   ↓
2. イベント受信
   ↓
3. React Query Cache更新
   ↓
4. コンポーネント自動再レンダリング
```

### C. ユーザーインタラクション
```
1. User Action (クリック、フィルター等)
   ↓
2. Component Event Handler
   ↓
3. State Update (useState, React Query)
   ↓
4. Conditional API Call
   ↓
5. UI Update
```

---

## 📦 技術スタック（MECE分類）

### 1. Core Framework
- **React** 19.2.0 - UIフレームワーク
- **TypeScript** 5.9.3 - 型安全性
- **Vite** 7.2.4 - ビルドツール

### 2. State Management
- **React Query** 5.90.12 - サーバーステート管理
- **Zustand** 5.0.9 - クライアントステート管理

### 3. UI Framework
- **Tailwind CSS** 4.1.18 - スタイリング
- **shadcn/ui** (Radix UI) - UIコンポーネント

### 4. Graph Visualization
- **React Flow** 11.11.4 - 【推奨】モダングラフライブラリ
- **Cytoscape.js** 3.33.1 - 【レガシー】従来型グラフライブラリ

### 5. Icons
- **Lucide React** 0.562.0 - UIアイコン
- **React Icons** 5.5.0 - 汎用アイコン
- **aws-react-icons** 3.2.0 - AWSアイコン

### 6. Layout Algorithms
- **@dagrejs/dagre** 1.1.8 - 階層レイアウト

---

## 🎯 アーキテクチャ決定記録（ADR）

### ADR-001: グラフライブラリの選択

**決定**: React Flowを推奨、Cytoscapeはレガシー扱い

**理由**:
- React Flowの方がReactとの統合が自然
- パフォーマンスが優れている
- コミュニティサポートが活発
- TypeScript型定義が充実

**影響**:
- 新機能はReact Flowで実装
- Cytoscapeは既存コード保守のみ
- 段階的にReact Flowへ移行

### ADR-002: 状態管理の分離

**決定**: サーバーステート（React Query）とUIステート（Zustand）を分離

**理由**:
- サーバーデータのキャッシュ・再検証はReact Queryが最適
- UIの一時的な状態（モーダル表示等）はZustandでシンプルに管理
- 関心の分離により保守性向上

### ADR-003: コンポーネント配置規則

**決定**: 機能別ディレクトリ構造（/reactflow、/graph、/onboarding等）

**理由**:
- 関連コンポーネントをグループ化
- インポートパスが明確
- スケーラビリティ向上

---

## 🚨 現在の技術的負債

### 1. 重複コンポーネント（High Priority）
- [ ] 6つのApp variants → 1つに統合
- [ ] NodeDetailPanel × 2 → 1つに統合
- [ ] アイコンコンポーネント × 7 → 標準化

### 2. 巨大ファイル（High Priority）
- [ ] sampleData.ts (19,836行) → モジュール分割

### 3. テスト不足（Critical）
- [ ] テストカバレッジ: 0% → 目標60%
- [ ] Vitestセットアップ
- [ ] コンポーネント単体テスト
- [ ] API統合テスト

### 4. ドキュメント不足（High Priority）
- [ ] README.md更新（テンプレートからの脱却）
- [ ] ARCHITECTURE.md作成
- [ ] CONTRIBUTING.md作成

---

## 📝 コーディング規約

### 1. ファイル命名
- コンポーネント: PascalCase.tsx
- フック: camelCase.ts (use〜)
- ユーティリティ: camelCase.ts
- 型定義: camelCase.ts

### 2. インポート順序
```typescript
// 1. 外部ライブラリ
import React from 'react';
import { useQuery } from '@tanstack/react-query';

// 2. 内部モジュール（絶対パス）
import { Button } from '@/components/ui/button';
import { useGraph } from '@/api/hooks';

// 3. 相対パス
import { NodeTooltip } from './NodeTooltip';

// 4. 型定義
import type { Node } from './types';

// 5. スタイル
import './styles.css';
```

### 3. コンポーネント構造
```typescript
/**
 * Component description
 */

// 1. インポート
import React from 'react';

// 2. 型定義
interface Props {
  // ...
}

// 3. 定数
const CONSTANT = 'value';

// 4. ヘルパー関数
const helperFunction = () => {};

// 5. コンポーネント
export const Component: React.FC<Props> = ({ prop }) => {
  // 5.1 フック
  const [state, setState] = useState();

  // 5.2 ハンドラー
  const handleClick = () => {};

  // 5.3 レンダー
  return <div>...</div>;
};
```

---

## 🔍 推奨開発フロー

### 1. 新機能追加
```bash
1. Issue作成 → 要件定義
2. ブランチ作成 (feature/xxx)
3. 型定義作成 (/types)
4. APIフック作成 (/api/hooks)
5. コンポーネント作成 (/components)
6. テスト作成 (*.test.tsx)
7. ドキュメント更新
8. PR作成 → レビュー
```

### 2. バグ修正
```bash
1. Issue作成 → 再現手順記載
2. ブランチ作成 (fix/xxx)
3. テストケース作成（失敗を確認）
4. 修正実装
5. テスト通過確認
6. PR作成 → レビュー
```

### 3. リファクタリング
```bash
1. Issue作成 → 改善提案
2. ブランチ作成 (refactor/xxx)
3. テスト作成（既存動作保証）
4. リファクタリング実施
5. テスト通過確認
6. パフォーマンス測定
7. PR作成 → レビュー
```

---

## 📚 参考リソース

### 公式ドキュメント
- [React Flow](https://reactflow.dev/)
- [React Query](https://tanstack.com/query/latest)
- [Tailwind CSS](https://tailwindcss.com/)
- [shadcn/ui](https://ui.shadcn.com/)

### 内部ドキュメント
- [ARCHITECTURE.md](./ARCHITECTURE.md) - システムアーキテクチャ
- [CONTRIBUTING.md](./CONTRIBUTING.md) - 開発ガイドライン
- [TESTING.md](./TESTING.md) - テスト戦略

---

**最終更新**: 2026-01-01
**メンテナー**: TFDrift-Falco Team
