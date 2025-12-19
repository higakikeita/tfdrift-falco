# TFDrift-Falco Graph UI Architecture

## 概要

TFDrift-Falcoの中核価値である**因果関係グラフ**を可視化するため、React + Cytoscape.jsベースのGraph UIを導入します。

## なぜGraph UIが必要か

### 現状の問題（Grafanaの限界）

Grafanaは時系列データの可視化に優れていますが、**リソース間の因果関係**を表現できません:

```
❌ Grafanaの表示:
- Drift detected: +1
- Falco event count increased
- IAM change happened at 12:03

❓ 疑問:
- どのIAM?
- そのIAMはどのServiceAccount?
- なぜこのPodに影響?
- その結果、なぜこのFalco Ruleが発火?
```

### TFDrift-Falcoの本質

TFDrift-Falcoは「Terraformの構成ドリフトが、どのリソース → どの権限 → どのRuntimeイベントにつながったか」を説明するツールです。

これは**時系列の話ではなく、因果関係の話**です → **Graph問題**

## アーキテクチャ概要

```
┌─────────────────────────────────────────────────────────────┐
│                    Data Sources                              │
├─────────────────────────────────────────────────────────────┤
│  Terraform State  │  CloudTrail  │  Falco Events  │  K8s API │
└──────────┬──────────────────┬──────────────┬────────────────┘
           │                  │               │
           v                  v               v
┌─────────────────────────────────────────────────────────────┐
│              TFDrift-Falco Core Engine                       │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Resource Relationship Builder                       │   │
│  │  - IAM → ServiceAccount mapping                      │   │
│  │  - ServiceAccount → Pod mapping                      │   │
│  │  - Pod → Container mapping                           │   │
│  │  - Drift → IAM → SA → Pod → Falco Event path        │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────┬──────────────────────────────────────────────────┘
           │
           v
┌─────────────────────────────────────────────────────────────┐
│                    Graph API Layer                           │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  GET  /api/graph/overview                           │   │
│  │  GET  /api/graph/drift/:drift_id                    │   │
│  │  GET  /api/graph/blast-radius/:resource_id          │   │
│  │  GET  /api/graph/path/:from/:to                     │   │
│  │  POST /api/graph/query                              │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────┬──────────────────────────────────────────────────┘
           │
           v
┌─────────────────────────────────────────────────────────────┐
│         Frontend: React + Cytoscape.js UI                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Graph Visualization                                 │   │
│  │  - Node rendering (Drift, IAM, SA, Pod, Falco)     │   │
│  │  - Edge rendering (cause → effect)                  │   │
│  │  - Interactive controls (zoom, pan, select)         │   │
│  │  - Highlighting (blast radius, path)                │   │
│  │  - Layer switching (attack view, ops view)          │   │
│  └─────────────────────────────────────────────────────┘   │
└──────────┬──────────────────────────────────────────────────┘
           │
           v
┌─────────────────────────────────────────────────────────────┐
│              Grafana (Entry Point)                           │
│  - Drift detection trend                                     │
│  - Falco event counts                                        │
│  - Drilldown link → Graph UI                                │
└─────────────────────────────────────────────────────────────┘
```

## グラフデータモデル

### ノードタイプ

```typescript
enum NodeType {
  TERRAFORM_CHANGE = 'terraform_change',
  IAM_POLICY = 'iam_policy',
  IAM_ROLE = 'iam_role',
  SERVICE_ACCOUNT = 'service_account',
  POD = 'pod',
  CONTAINER = 'container',
  FALCO_EVENT = 'falco_event',
  SECURITY_GROUP = 'security_group',
  NETWORK = 'network'
}

interface GraphNode {
  id: string;
  type: NodeType;
  label: string;
  data: {
    resource_type: string;
    resource_name: string;
    timestamp?: string;
    severity?: 'critical' | 'high' | 'medium' | 'low';
    metadata: Record<string, any>;
  };
  style: {
    color?: string;
    size?: number;
    shape?: string;
  };
}
```

### エッジタイプ

```typescript
enum EdgeType {
  CAUSED_BY = 'caused_by',        // Drift → IAM change
  GRANTS_ACCESS = 'grants_access', // IAM → ServiceAccount
  USED_BY = 'used_by',            // SA → Pod
  CONTAINS = 'contains',          // Pod → Container
  TRIGGERED = 'triggered'         // Container → Falco Event
}

interface GraphEdge {
  id: string;
  source: string;
  target: string;
  type: EdgeType;
  label: string;
  data: {
    relationship: string;
    metadata: Record<string, any>;
  };
  style: {
    color?: string;
    width?: number;
    line_style?: 'solid' | 'dashed' | 'dotted';
  };
}
```

### グラフ全体

```typescript
interface CausalGraph {
  nodes: GraphNode[];
  edges: GraphEdge[];
  metadata: {
    generated_at: string;
    root_cause?: string;
    blast_radius_size?: number;
  };
}
```

## API仕様

### 1. Overview Graph

全体の因果関係グラフを取得

```http
GET /api/graph/overview?time_range=1h&severity=high,critical

Response:
{
  "nodes": [...],
  "edges": [...],
  "metadata": {
    "generated_at": "2025-12-19T17:50:00Z",
    "total_nodes": 42,
    "total_edges": 56
  }
}
```

### 2. Drift-specific Graph

特定のDriftから始まる因果関係パスを取得

```http
GET /api/graph/drift/:drift_id

Response:
{
  "root_cause": {
    "id": "drift-123",
    "type": "terraform_change",
    "label": "IAM Policy modification"
  },
  "path": [
    {
      "from": "drift-123",
      "to": "iam-policy-456",
      "relationship": "caused_by"
    },
    {
      "from": "iam-policy-456",
      "to": "sa-789",
      "relationship": "grants_access"
    },
    ...
  ],
  "graph": {
    "nodes": [...],
    "edges": [...]
  }
}
```

### 3. Blast Radius

特定リソースの影響範囲を取得

```http
GET /api/graph/blast-radius/:resource_id

Response:
{
  "center": {
    "id": "iam-policy-456",
    "type": "iam_policy"
  },
  "affected_resources": [
    {
      "id": "sa-789",
      "type": "service_account",
      "distance": 1
    },
    {
      "id": "pod-101",
      "type": "pod",
      "distance": 2
    },
    ...
  ],
  "graph": {
    "nodes": [...],
    "edges": [...]
  }
}
```

### 4. Path Query

2つのリソース間のパスを検索

```http
GET /api/graph/path/:from/:to

Response:
{
  "paths": [
    {
      "length": 4,
      "nodes": ["drift-123", "iam-456", "sa-789", "pod-101", "falco-202"],
      "edges": [...]
    }
  ]
}
```

## フロントエンド構造

```
ui/
├── src/
│   ├── components/
│   │   ├── Graph/
│   │   │   ├── CytoscapeGraph.tsx       # メイングラフコンポーネント
│   │   │   ├── NodeRenderer.tsx         # ノード描画
│   │   │   ├── EdgeRenderer.tsx         # エッジ描画
│   │   │   ├── GraphControls.tsx        # ズーム/パン/リセット
│   │   │   ├── GraphLegend.tsx          # 凡例
│   │   │   └── GraphToolbar.tsx         # ツールバー
│   │   ├── Filters/
│   │   │   ├── NodeTypeFilter.tsx       # ノードタイプフィルタ
│   │   │   ├── SeverityFilter.tsx       # 重要度フィルタ
│   │   │   ├── TimeRangeFilter.tsx      # 時間範囲フィルタ
│   │   │   └── SearchFilter.tsx         # 検索フィルタ
│   │   ├── Panels/
│   │   │   ├── NodeDetailPanel.tsx      # ノード詳細パネル
│   │   │   ├── PathExplorerPanel.tsx    # パス探索パネル
│   │   │   ├── BlastRadiusPanel.tsx     # 影響範囲パネル
│   │   │   └── RecommendationPanel.tsx  # 修正推奨パネル
│   │   └── Layouts/
│   │       ├── AttackViewLayout.tsx     # 攻撃視点レイアウト
│   │       └── OpsViewLayout.tsx        # 運用視点レイアウト
│   ├── hooks/
│   │   ├── useGraphData.ts              # グラフデータ取得
│   │   ├── useCytoscape.ts              # Cytoscape制御
│   │   ├── useGraphInteraction.ts       # インタラクション管理
│   │   └── useGraphLayout.ts            # レイアウト管理
│   ├── services/
│   │   ├── graphApi.ts                  # Graph API client
│   │   └── graphProcessor.ts            # グラフデータ処理
│   ├── styles/
│   │   ├── cytoscapeStyles.ts           # Cytoscapeスタイル定義
│   │   └── theme.ts                     # UIテーマ
│   ├── types/
│   │   └── graph.ts                     # 型定義
│   └── App.tsx
├── package.json
└── tsconfig.json
```

## Cytoscapeスタイル定義

```typescript
export const cytoscapeStylesheet = [
  // Terraform Change nodes
  {
    selector: 'node[type="terraform_change"]',
    style: {
      'background-color': '#ff6b6b',
      'label': 'data(label)',
      'shape': 'hexagon',
      'width': 80,
      'height': 80,
      'font-size': 12,
      'text-valign': 'center',
      'text-halign': 'center',
      'border-width': 3,
      'border-color': '#c92a2a'
    }
  },

  // IAM Policy nodes
  {
    selector: 'node[type="iam_policy"]',
    style: {
      'background-color': '#4dabf7',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70,
      'font-size': 11
    }
  },

  // Service Account nodes
  {
    selector: 'node[type="service_account"]',
    style: {
      'background-color': '#51cf66',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 70,
      'height': 70
    }
  },

  // Pod nodes
  {
    selector: 'node[type="pod"]',
    style: {
      'background-color': '#ffd43b',
      'label': 'data(label)',
      'shape': 'round-rectangle',
      'width': 60,
      'height': 60
    }
  },

  // Falco Event nodes
  {
    selector: 'node[type="falco_event"]',
    style: {
      'background-color': '#f06595',
      'label': 'data(label)',
      'shape': 'diamond',
      'width': 70,
      'height': 70,
      'font-size': 10
    }
  },

  // Critical severity highlight
  {
    selector: 'node[severity="critical"]',
    style: {
      'border-width': 5,
      'border-color': '#c92a2a',
      'background-color': '#ff6b6b'
    }
  },

  // Edges - caused_by
  {
    selector: 'edge[type="caused_by"]',
    style: {
      'width': 3,
      'line-color': '#ff6b6b',
      'target-arrow-color': '#ff6b6b',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier',
      'label': 'data(label)',
      'font-size': 9,
      'text-rotation': 'autorotate'
    }
  },

  // Edges - grants_access
  {
    selector: 'edge[type="grants_access"]',
    style: {
      'width': 2,
      'line-color': '#4dabf7',
      'target-arrow-color': '#4dabf7',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier'
    }
  },

  // Edges - used_by
  {
    selector: 'edge[type="used_by"]',
    style: {
      'width': 2,
      'line-color': '#51cf66',
      'target-arrow-color': '#51cf66',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier'
    }
  },

  // Edges - triggered
  {
    selector: 'edge[type="triggered"]',
    style: {
      'width': 3,
      'line-color': '#f06595',
      'target-arrow-color': '#f06595',
      'target-arrow-shape': 'triangle',
      'curve-style': 'bezier'
    }
  },

  // Highlighted path
  {
    selector: '.highlighted',
    style: {
      'background-color': '#ffd43b',
      'line-color': '#ffd43b',
      'target-arrow-color': '#ffd43b',
      'width': 5,
      'z-index': 999
    }
  },

  // Selected node
  {
    selector: ':selected',
    style: {
      'border-width': 6,
      'border-color': '#228be6',
      'z-index': 999
    }
  }
];
```

## レイアウト戦略

### 1. Hierarchical Layout（階層レイアウト）

因果関係を上から下に表示（デフォルト）

```typescript
const hierarchicalLayout = {
  name: 'dagre',
  rankDir: 'TB',  // Top to Bottom
  nodeSep: 100,
  rankSep: 150,
  animate: true,
  animationDuration: 500
};
```

### 2. Radial Layout（放射状レイアウト）

Blast Radiusビュー用

```typescript
const radialLayout = {
  name: 'concentric',
  concentric: (node) => node.data('distance'),
  levelWidth: () => 2,
  animate: true
};
```

### 3. Force-directed Layout（力学レイアウト）

複雑な関係性の探索用

```typescript
const forceLayout = {
  name: 'cose',
  nodeRepulsion: 8000,
  idealEdgeLength: 100,
  animate: true
};
```

## インタラクティブ機能

### 1. ノードクリック

- クリック → 詳細パネル表示
- ダブルクリック → Blast Radius表示
- 右クリック → コンテキストメニュー

### 2. パスハイライト

```typescript
function highlightPath(startNodeId: string, endNodeId: string) {
  const path = cy.elements().aStar({
    root: `#${startNodeId}`,
    goal: `#${endNodeId}`
  });

  path.path.addClass('highlighted');
}
```

### 3. フィルタリング

```typescript
function filterByNodeType(types: NodeType[]) {
  cy.nodes().forEach(node => {
    if (!types.includes(node.data('type'))) {
      node.style('display', 'none');
    } else {
      node.style('display', 'element');
    }
  });
}
```

### 4. Blast Radius計算

```typescript
function calculateBlastRadius(nodeId: string, maxDepth: number = 3) {
  const startNode = cy.$id(nodeId);
  const affected = startNode.successors().filter(node => {
    const distance = getDistance(startNode, node);
    return distance <= maxDepth;
  });

  return affected;
}
```

## Grafana統合

### Grafanaパネルからのドリルダウン

```javascript
// Grafana Panel Link設定
{
  "links": [
    {
      "title": "View Causality Graph",
      "url": "http://localhost:3000/graph?drift_id=${__data.fields.drift_id}"
    }
  ]
}
```

### データソース連携

```typescript
// Grafanaクエリパラメータを解析
const urlParams = new URLSearchParams(window.location.search);
const driftId = urlParams.get('drift_id');
const timeRange = urlParams.get('from');

// Graph APIを呼び出し
const graphData = await fetchDriftGraph(driftId, timeRange);
```

## 実装フェーズ

### Phase 1: 基盤構築（Week 1）
- ✅ React + TypeScript + Vite セットアップ
- ✅ Cytoscape.js 統合
- ✅ 基本的なグラフ描画機能
- ✅ サンプルデータでの動作確認

### Phase 2: バックエンド統合（Week 2）
- ⏳ Graph API実装
- ⏳ Resource Relationship Builder
- ⏳ Drift → Falco因果パス生成ロジック
- ⏳ APIとフロントエンドの統合

### Phase 3: UI/UX強化（Week 3）
- ⏳ ノード詳細パネル
- ⏳ Blast Radiusビュー
- ⏳ パス探索機能
- ⏳ フィルタリング機能
- ⏳ レイアウト切り替え

### Phase 4: Grafana統合（Week 4）
- ⏳ Grafanaドリルダウンリンク
- ⏳ 時系列データとの連携
- ⏳ 統合テスト
- ⏳ ドキュメント作成

## 技術スタック

### フロントエンド
- **React 18** - UIフレームワーク
- **TypeScript** - 型安全性
- **Cytoscape.js** - グラフビジュアライゼーション
- **Vite** - ビルドツール
- **TanStack Query (React Query)** - データフェッチング
- **Zustand** - 状態管理
- **Tailwind CSS** - スタイリング

### バックエンド
- **Go** - 既存のTFDrift-Falcoコア
- **Gin** - HTTPフレームワーク（Graph API）
- **GraphQL (オプション)** - 柔軟なクエリ

### インフラ
- **Docker** - コンテナ化
- **Nginx** - リバースプロキシ
- **Grafana** - 時系列ダッシュボード（入口）

## 成功指標

1. **視認性**: 因果関係が一目で理解できる
2. **パフォーマンス**: 1000ノード以下のグラフを1秒以内に描画
3. **インタラクティブ性**: クリック→詳細表示が0.5秒以内
4. **統合性**: GrafanaからシームレスにGraph UIへ遷移

## まとめ

TFDrift-Falcoの真価は**「なぜそれが起きたか」を説明すること**にあります。

- **Grafana** = 「いつ・何回」（入口）
- **Graph UI** = 「なぜ・どこを直せば」（核心）

この2つの組み合わせで、完全なObservability（可観測性）と Actionability（行動可能性）を実現します。
