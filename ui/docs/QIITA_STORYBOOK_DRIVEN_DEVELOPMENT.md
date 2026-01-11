# Storybook駆動開発で変わるフロントエンド開発体験 - TFDrift-Falcoでの実践

## はじめに

フロントエンド開発で「コンポーネントを作ってから動作確認」という流れに疲れていませんか？
アプリ全体を起動し、特定の画面まで遷移し、条件を整えてやっと目的のコンポーネントが表示される...そんな非効率な開発体験を変える手法が **Storybook駆動開発（Storybook-Driven Development: SDD）** です。

本記事では、AWSインフラのDrift検出ツール **TFDrift-Falco** の開発で実践したStorybook駆動開発の導入事例を紹介します。

## TFDrift-Falcoとは

TFDrift-Falcoは、Terraformで管理されているAWSリソースと実際のAWSインフラの差分（Drift）を検出・可視化するOSSツールです。

- **バックエンド**: Go（Terraform State解析、AWS API連携）
- **フロントエンド**: React + TypeScript + Vite
- **可視化**: Cytoscape.js（グラフビュー）+ Storybook

特に、複雑なAWSリソース間の依存関係をグラフで可視化する **CytoscapeGraphコンポーネント** の開発において、Storybook駆動開発が大きな威力を発揮しました。

## Storybook駆動開発とは？

### 従来の開発フロー（Before）

```
1. コンポーネントを実装
2. アプリを起動（npm run dev）
3. 特定の画面まで遷移
4. API呼び出しやデータセットアップ
5. やっとコンポーネントが表示される
6. 「あれ、サイズがおかしい...」
7. コードを修正 → 手順3に戻る（繰り返し）
```

**問題点**:
- フィードバックループが長い（修正→確認に時間がかかる）
- 特定の状態を再現するのが面倒（エラー状態、空データ、大量データなど）
- 複数のバリエーションを同時に確認できない

### Storybook駆動開発（After）

```
1. Storyで仕様を定義（コンポーネントのAPI設計）
2. モックデータを用意
3. コンポーネントを実装
4. Storybookでリアルタイム確認（HMR）
5. 複数のバリエーション（Empty/Error/Large）を並行して確認
6. 完成
```

**メリット**:
- **超高速フィードバック**: 修正が即座に反映（HMR）
- **独立した開発環境**: バックエンドAPIなしで開発可能
- **全バリエーションを一度に確認**: 10個のストーリーを同時に表示
- **自動ドキュメント化**: Storyがそのままドキュメントになる

## 実践：CytoscapeGraphコンポーネントの開発

### Step 1: Story Firstの設計

まず、Storyを書くことでコンポーネントのAPIを設計します。

```typescript
// CytoscapeGraph.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { CytoscapeGraph } from './CytoscapeGraph';
import { mockMediumGraph } from '../mocks/graphData';

const meta = {
  title: 'Components/CytoscapeGraph',
  component: CytoscapeGraph,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  argTypes: {
    layout: {
      control: 'select',
      options: ['fcose', 'dagre', 'cose', 'grid'],
      description: 'グラフレイアウトアルゴリズム'
    }
  }
} satisfies Meta<typeof CytoscapeGraph>;

export default meta;
type Story = StoryObj<typeof meta>;

// デフォルト表示
export const Default: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  }
};
```

この時点で、以下が決まります：
- Propsは `elements` と `layout`
- `layout` は4種類のアルゴリズムをサポート
- フルスクリーン表示が最適

**従来の開発なら、実装してから「あれ、このProps必要だったかも...」となるところを、先に設計できる**のが大きな利点です。

### Step 2: モックデータの構造化

Storyで使うモックデータを別ファイルに切り出します。

```typescript
// mocks/graphData.ts
import type { CytoscapeElements } from '../types/graph';

// 小規模グラフ（10ノード）
export const mockSmallGraph: CytoscapeElements = {
  nodes: [
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'prod-vpc' }},
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'subnet-a', parent: 'vpc-1' }},
    { data: { id: 'eks-1', resource_type: 'aws_eks_cluster', label: 'prod-eks', parent: 'subnet-1' }},
    // ...
  ],
  edges: [
    { data: { id: 'e1', source: 'eks-1', target: 'sg-1', label: 'uses' }},
    // ...
  ]
};

// 中規模グラフ（30ノード）
export const mockMediumGraph: CytoscapeElements = { /* ... */ };

// Drift状態を含むグラフ
export const mockGraphWithDrift: CytoscapeElements = {
  nodes: [
    {
      data: {
        id: 'eks-1',
        resource_type: 'aws_eks_cluster',
        label: 'prod-eks',
        severity: 'high'  // Modified
      }
    },
    {
      data: {
        id: 'rds-1',
        resource_type: 'aws_db_instance',
        label: 'prod-db',
        severity: 'critical'  // Missing
      }
    }
  ],
  edges: []
};

// 大規模グラフ生成関数
export function generateLargeGraph(nodeCount: number): CytoscapeElements {
  const nodes: CytoscapeElements['nodes'] = [];
  const edges: CytoscapeElements['edges'] = [];

  // VPC + Subnets + Resources
  for (let i = 1; i <= nodeCount; i++) {
    nodes.push({
      data: {
        id: `resource-${i}`,
        resource_type: 'aws_eks_cluster',
        label: `resource-${i}`
      }
    });
  }

  return { nodes, edges };
}
```

**ポイント**:
- 再利用可能なモックデータ
- 型安全（TypeScript）
- テストでも使える
- プログラマティックにデータ生成可能

### Step 3: 包括的なStoriesの作成

TFDrift-Falcoでは、**17個のStories** を作成しました。

#### 基本Stories

```typescript
// デフォルト表示（10ノード）
export const Default: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose'
  }
};

// 空データ（エラーハンドリング確認）
export const Empty: Story = {
  args: {
    elements: { nodes: [], edges: [] },
    layout: 'fcose'
  }
};
```

#### レイアウトバリエーション

```typescript
// fcoseレイアウト（推奨）
export const LayoutFcose: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: 'VPC/Subnet階層がある場合の推奨レイアウト。'
      }
    }
  }
};

// dagreレイアウト
export const LayoutDagre: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'dagre'
  }
};

// coseレイアウト
export const LayoutCose: Story = { /* ... */ };

// gridレイアウト
export const LayoutGrid: Story = { /* ... */ };
```

#### データ量バリエーション

```typescript
// 小規模（10ノード）
export const SmallGraph: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose'
  }
};

// 中規模（30ノード）
export const MediumGraph: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose'
  }
};

// 大規模（100ノード）
export const LargeGraph: Story = {
  args: {
    elements: generateLargeGraph(100),
    layout: 'fcose'
  }
};

// 超大規模（200ノード）- パフォーマンステスト
export const VeryLargeGraph: Story = {
  args: {
    elements: generateLargeGraph(200),
    layout: 'fcose'
  }
};
```

#### Drift状態の可視化

```typescript
// Drift強調表示
export const DriftHighlighted: Story = {
  args: {
    elements: mockGraphWithDrift,
    layout: 'fcose'
  },
  parameters: {
    docs: {
      description: {
        story: `
Drift状態を持つリソースの表示例。

**Severity:**
- \`critical\`: Missing resources（赤ボーダー）
- \`high\`: Modified resources（オレンジボーダー）
- \`medium\`: Unmanaged resources（黄ボーダー）
        `
      }
    }
  }
};
```

#### インタラクション

```typescript
// ノードクリック
export const WithNodeClick: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose',
    onNodeClick: (nodeId: string, nodeData: any) => {
      console.log('Node clicked:', { nodeId, nodeData });
    }
  }
};

// パスハイライト
export const WithHighlightedPath: Story = {
  args: {
    elements: mockSmallGraph,
    layout: 'fcose',
    highlightedPath: ['eks-1', 'sg-1', 'iam-1']
  }
};
```

#### Playground

```typescript
// 全機能を自由にテストできるStory
export const Playground: Story = {
  args: {
    elements: mockMediumGraph,
    layout: 'fcose',
    onNodeClick: (nodeId, nodeData) => console.log('Node:', nodeId),
    onEdgeClick: (edgeId, edgeData) => console.log('Edge:', edgeId),
    highlightedPath: [],
    className: ''
  }
};
```

### Step 4: 実装とリアルタイム確認

Storybookを起動すると、**すべてのStoriesがリアルタイムで表示されます**。

```bash
npm run storybook
# → http://localhost:6006/
```

実装中の様子：

1. **ノードサイズを調整** → 即座に17個すべてのStoryに反映
2. **VPC階層の透明度を変更** → リアルタイムで確認
3. **エッジのスタイル変更** → 全レイアウトで同時確認

従来なら「アプリ起動 → 画面遷移 → 確認 → 修正 → 再起動」を繰り返していたところが、**Storybookなら1秒でフィードバック** が得られます。

## TFDrift-Falcoで遭遇した問題とStorybookによる解決

### 問題1: ノードサイズの最適化

**状況**: AWSリソースノードのサイズが大きすぎて、構成図として見づらい

**従来の開発なら**:
1. アプリ起動
2. データ取得（API呼び出し）
3. グラフ表示
4. 「サイズおかしい...」
5. 修正してリロード（繰り返し）

**Storybookなら**:
```typescript
// SmallGraph (10 nodes)
// MediumGraph (30 nodes)
// LargeGraph (100 nodes)
// VeryLargeGraph (200 nodes)
```
を **同時に表示** して、すべてのデータ量で最適なサイズを一度に確認。

**結果**: 90px → 45px に縮小が最適と判明（5分で決定）

### 問題2: VPC/Subnet階層の可視化

**状況**: コンパウンドノード（親子関係）がうまく表示されない

**Storybookでデバッグ**:
```typescript
export const WithVPCHierarchy: Story = {
  args: {
    elements: mockMediumGraph,  // VPC + 3 Subnets
    layout: 'fcose'
  }
};
```

Chromeデベロッパーツールで確認 → 背景の透明度が低すぎることが判明

**修正**:
```typescript
// cytoscapeStyles.ts
{
  selector: 'node[resource_type="aws_vpc"]',
  style: {
    'background-color': 'rgba(235, 245, 251, 0.6)',  // 0.3 → 0.6
    'padding': '50px'  // 30px → 50px
  }
}
```

Storybookでリアルタイム確認 → 即座に「これだ！」と確定。

### 問題3: 4種類のレイアウトアルゴリズムの比較

TFDrift-Falcoは4種類のレイアウトをサポート：
- **fcose**: Force-directed compound spring embedder（推奨）
- **dagre**: 階層的有向グラフ
- **cose**: 物理シミュレーション
- **grid**: グリッド配置

**Storybookなら**:
```typescript
export const LayoutFcose: Story = { /* ... */ };
export const LayoutDagre: Story = { /* ... */ };
export const LayoutCose: Story = { /* ... */ };
export const LayoutGrid: Story = { /* ... */ };
```

4つのStoryを並べて表示 → **一目でfcoseが最適** と判断できる。

従来なら、アプリで切り替えながら何度も確認する必要があったところが、**10秒で結論が出た**。

## Storybook駆動開発のベストプラクティス

TFDrift-Falcoで実践して効果的だったルールを紹介します。

### 1. Story First（コンポーネントより先にStoryを書く）

```typescript
// ❌ Bad: いきなりコンポーネントを作る
export const CytoscapeGraph = ({ elements, layout }) => { ... }

// ✅ Good: まずStoryで仕様を定義
export const Default: Story = {
  args: {
    elements: mockGraphData,
    layout: 'fcose'
  }
}
// → ここでコンポーネントのAPIが明確になる
```

### 2. 最低3つのバリエーション（Default/Empty/Error）

```typescript
export const Default: Story = { /* 正常系 */ };
export const Empty: Story = { /* 空データ */ };
export const Error: Story = { /* エラー状態 */ };
```

**理由**: エッジケースを忘れがちだが、Storyにすることで強制的に考える

### 3. モックデータは別ファイルに分離

```typescript
// ❌ Bad: Storyファイル内にべた書き
export const Default: Story = {
  args: {
    elements: {
      nodes: [{ data: { id: 'vpc-1', ... } }],
      edges: [...]
    }
  }
};

// ✅ Good: 再利用可能なモックデータ
import { mockMediumGraph } from '../mocks/graphData';

export const Default: Story = {
  args: { elements: mockMediumGraph }
};
```

### 4. autodocs tagで自動ドキュメント生成

```typescript
const meta = {
  title: 'Components/CytoscapeGraph',
  component: CytoscapeGraph,
  tags: ['autodocs'],  // ← これだけでドキュメント生成
} satisfies Meta<typeof CytoscapeGraph>;
```

### 5. Controlsで主要Propsをインタラクティブに操作

```typescript
const meta = {
  argTypes: {
    layout: {
      control: 'select',
      options: ['fcose', 'dagre', 'cose', 'grid']
    },
    nodeScale: {
      control: { type: 'range', min: 0.5, max: 2, step: 0.1 }
    }
  }
} satisfies Meta<typeof CytoscapeGraph>;
```

Storybook UIで **リアルタイムにプロパティを変更** → 即座に反映

## 導入効果

### 定量的な効果

| 指標 | Before | After | 改善率 |
|------|--------|-------|--------|
| コンポーネント確認時間 | 30秒 | 1秒 | **30倍** |
| バリエーション確認 | 1個ずつ切り替え | 17個同時表示 | **17倍** |
| バグ発見率 | - | - | **↑↑** |
| ドキュメント作成時間 | 手動で作成 | 自動生成 | **無限大** |

### 定性的な効果

#### 開発体験の向上
- **ストレスフリー**: アプリ全体を起動する必要がない
- **集中できる**: コンポーネント開発に集中
- **自信がつく**: すべてのバリエーションを確認してからリリース

#### チーム開発への影響
- **レビューしやすい**: PRにStorybookのURLを貼るだけ
- **引き継ぎが楽**: Storyがそのまま仕様書
- **デザイナーとの協業**: Storybookを見せながら議論

#### 品質向上
- **エッジケースの考慮**: Empty/Error/Largeを最初から作る
- **回帰テストが楽**: Storybookで全バリエーションを一覧確認
- **一貫性**: 同じモックデータを使うため、ブレがない

## プロジェクト構成

TFDrift-Falcoのディレクトリ構成：

```
ui/
├── src/
│   ├── components/
│   │   ├── CytoscapeGraph.tsx         # コンポーネント本体
│   │   └── CytoscapeGraph.stories.tsx # 17個のStories
│   ├── mocks/
│   │   └── graphData.ts               # 再利用可能なモックデータ
│   ├── types/
│   │   └── graph.ts                   # 型定義
│   └── styles/
│       └── cytoscapeStyles.ts         # スタイル定義
├── docs/
│   └── STORYBOOK_DRIVEN_DEVELOPMENT.md # SDDガイドライン
└── package.json
```

### 依存関係

```json
{
  "devDependencies": {
    "storybook": "^10.1.11",
    "@storybook/react-vite": "^10.1.11",
    "@storybook/addon-essentials": "^10.1.11",
    "@storybook/addon-interactions": "^10.1.11"
  }
}
```

### 起動コマンド

```bash
# Storybook起動
npm run storybook  # http://localhost:6006/

# Storybookビルド（デプロイ用）
npm run build-storybook
```

## チームへの展開方法

### Step 1: ガイドラインドキュメント作成

```markdown
# docs/STORYBOOK_DRIVEN_DEVELOPMENT.md

## 原則
1. Story First
2. Isolated Development
3. Visual Testing
4. Documentation as Code

## チェックリスト
- [ ] まずStoryを書く
- [ ] 最低3つのバリエーション
- [ ] モックデータを別ファイルに
- [ ] autodocs tagを追加
```

### Step 2: テンプレート提供

```typescript
// template.stories.tsx
import type { Meta, StoryObj } from '@storybook/react';
import { YourComponent } from './YourComponent';

const meta = {
  title: 'Components/YourComponent',
  component: YourComponent,
  tags: ['autodocs'],
} satisfies Meta<typeof YourComponent>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {}
};

export const Empty: Story = {
  args: {}
};
```

### Step 3: CI/CDへの統合

```yaml
# .github/workflows/storybook.yml
name: Storybook Deploy
on: [push]
jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - run: npm ci
      - run: npm run build-storybook
      - uses: peaceiris/actions-gh-pages@v3
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./storybook-static
```

## まとめ

### Storybook駆動開発の本質

Storybook駆動開発は単なる「Storybookを使う」という話ではありません。

**本質は「コンポーネントを独立した環境で設計・開発・テスト・ドキュメント化する」という思想です。**

### TFDrift-Falcoでの成果

- ✅ **17個のStories** で全バリエーションをカバー
- ✅ **開発速度30倍**: フィードバックループの短縮
- ✅ **自動ドキュメント化**: Storyが仕様書に
- ✅ **品質向上**: エッジケースの網羅

### これから始める人へ

1. **小さく始める**: まず1コンポーネントだけStory化
2. **Story First**: コンポーネントより先にStoryを書く
3. **3つのバリエーション**: Default/Empty/Errorから
4. **チームで共有**: Storybook URLをPRに貼る

### 参考リンク

- [TFDrift-Falco GitHub](https://github.com/higakikeita/tfdrift-falco)
- [Storybook公式ドキュメント](https://storybook.js.org/)
- [Component-Driven Development](https://www.componentdriven.org/)

---

**Storybook駆動開発は、フロントエンド開発の常識を変えるゲームチェンジャーです。**

ぜひ、あなたのプロジェクトでも試してみてください。

**Happy Coding with Storybook!**
