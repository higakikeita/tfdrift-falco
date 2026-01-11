# Storybook駆動開発（SDD）ガイドライン

## 🎯 Storybook駆動開発とは

**Storybook-Driven Development (SDD)** は、UIコンポーネントの開発において、Storybookを中心に据えた開発手法です。

### 原則:
1. **Story First**: コンポーネントを作る前にStoryを書く
2. **Isolated Development**: 独立した環境で開発
3. **Visual Testing**: 目で見て確認できる状態を保つ
4. **Documentation as Code**: Storyがそのままドキュメントになる

---

## 🔄 TFDrift-Falco UIでの開発フロー

### 1. Story作成（設計フェーズ）

```typescript
// ❌ Bad: いきなりコンポーネントを作る
export const MyComponent = () => { ... }

// ✅ Good: まずStoryで仕様を定義
export const Default: Story = {
  args: {
    elements: mockGraphData,
    layout: 'fcose'
  }
}
```

**理由**: Storyを先に書くことで、コンポーネントのAPIを設計できる

---

### 2. コンポーネント実装

Storyで定義したPropsに従ってコンポーネントを実装：

```typescript
interface CytoscapeGraphProps {
  elements: CytoscapeElements;
  layout?: LayoutType;
  onNodeClick?: (nodeId: string) => void;
  // Storyで必要だとわかったProps
}
```

---

### 3. ビジュアル確認

```bash
npm run storybook
```

- ブラウザで http://localhost:6006 を開く
- 各Storyを見ながら調整
- リアルタイムでプレビュー

---

### 4. バリエーション追加

```typescript
// 正常系
export const Default: Story = { ... }

// 異常系
export const EmptyState: Story = {
  args: { elements: { nodes: [], edges: [] } }
}

// エッジケース
export const LargeGraph: Story = {
  args: { elements: generate100Nodes() }
}
```

---

## 📋 Story命名規則

### パターン別命名:

```typescript
// 状態別
export const Default: Story           // デフォルト状態
export const Loading: Story           // ローディング中
export const Error: Story             // エラー状態
export const Empty: Story             // 空データ

// バリエーション別
export const WithVPCHierarchy: Story  // VPC階層あり
export const DriftHighlighted: Story  // Drift強調表示
export const SmallScale: Story        // 小サイズ
export const LargeScale: Story        // 大サイズ

// インタラクション別
export const WithTooltip: Story       // ツールチップ付き
export const WithSelection: Story     // 選択状態

// データ量別
export const SmallGraph: Story        // 10 nodes
export const MediumGraph: Story       // 50 nodes
export const LargeGraph: Story        // 100+ nodes
```

---

## 🧪 Story作成のベストプラクティス

### 1. **モックデータを用意**

```typescript
// src/mocks/graphData.ts
export const mockVPCHierarchy: CytoscapeElements = {
  nodes: [
    { data: { id: 'vpc-1', resource_type: 'aws_vpc', label: 'prod-vpc' }},
    { data: { id: 'subnet-1', resource_type: 'aws_subnet', label: 'subnet-a', parent: 'vpc-1' }},
    // ...
  ],
  edges: [...]
}
```

### 2. **args vs render**

```typescript
// ✅ シンプルな場合: args
export const Default: Story = {
  args: { elements: mockData }
}

// ✅ 複雑な場合: render
export const WithInteraction: Story = {
  render: (args) => {
    const [selected, setSelected] = useState(null);
    return <CytoscapeGraph {...args} onNodeClick={setSelected} />
  }
}
```

### 3. **Controls（インタラクティブ設定）**

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

### 4. **Play関数（自動インタラクション）**

```typescript
export const WithNodeSelection: Story = {
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement);
    const node = await canvas.findByTestId('node-vpc-1');
    await userEvent.click(node);
  }
}
```

---

## 🎨 TFDrift-Falcoで重視するStory

### CytoscapeGraphコンポーネント:

#### ✅ 必須Stories:
1. **Default** - 基本表示
2. **WithVPCHierarchy** - VPC/Subnet階層
3. **Empty** - 空データ
4. **Loading** - ローディング状態

#### ✅ レイアウトバリエーション:
5. **LayoutFcose** - fcoseレイアウト
6. **LayoutDagre** - dagreレイアウト
7. **LayoutCose** - coseレイアウト
8. **LayoutGrid** - gridレイアウト

#### ✅ スケールバリエーション:
9. **SmallScale** - 0.7x
10. **NormalScale** - 1.0x
11. **LargeScale** - 1.3x

#### ✅ データ量バリエーション:
12. **SmallGraph** - 10-20 nodes
13. **MediumGraph** - 50 nodes
14. **LargeGraph** - 100+ nodes

#### ✅ 状態バリエーション:
15. **DriftHighlighted** - Driftのあるノード強調
16. **AllResourceTypes** - 全AWSサービスタイプ表示

---

## 📊 Storybook Addons活用

### インストール済み推奨Addons:

```bash
# 既にインストール済み
@storybook/addon-essentials  # 基本機能
@storybook/addon-interactions # インタラクションテスト
@storybook/addon-links       # Story間リンク
```

### 使い方:

```typescript
// Docs（自動生成ドキュメント）
export default {
  tags: ['autodocs']  // 自動でDocs生成
} satisfies Meta;

// Actions（イベントログ）
export const WithActions: Story = {
  args: {
    onNodeClick: fn(),  // クリックイベントをログ表示
  }
}
```

---

## 🚀 開発ワークフロー例

### 新機能「ノードツールチップ」の追加:

```bash
# 1. Storybook起動
npm run storybook

# 2. Story作成（仕様定義）
# src/components/CytoscapeGraph.stories.tsx
export const WithTooltip: Story = {
  args: {
    showTooltip: true,
    elements: mockData
  }
}

# 3. コンポーネント実装
# src/components/CytoscapeGraph.tsx
# showTooltip propを追加、ツールチップ機能実装

# 4. Storybookで確認しながら調整
# → リアルタイムでプレビュー更新

# 5. 完成したらコミット
git add .
git commit -m "feat: Add node tooltip feature"
```

---

## 📖 参考リソース

- [Storybook公式ドキュメント](https://storybook.js.org/)
- [Component-Driven Development](https://www.componentdriven.org/)
- [Visual Testing Handbook](https://storybook.js.org/tutorials/visual-testing-handbook/)

---

## ✅ チェックリスト

新しいコンポーネントを作る時:

- [ ] まずStoryを書く
- [ ] 最低3つのバリエーション（Default, Empty, Error）
- [ ] Controlsで主要Propsを操作可能にする
- [ ] モックデータを別ファイルに分離
- [ ] autodocs tagを追加
- [ ] README.mdにStoryへのリンクを追加

---

**最終更新**: 2026-01-03
**対象バージョン**: v0.5.0以降
