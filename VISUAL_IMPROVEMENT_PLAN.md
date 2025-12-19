# TFDrift-Falco ビジュアル品質改善計画

## 現状の問題
- Cytoscape.jsのレンダリング品質が期待値以下
- アイコンの表示品質が低い
- HTMLオーバーレイの位置ずれ・品質問題

## 提案アーキテクチャ

### Option 1: サーバーサイド図生成 (推奨)
```
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐
│   React UI  │────▶│  Backend API │────▶│ Diagram Service │
└─────────────┘     └──────────────┘     └─────────────────┘
                                                   │
                                                   ▼
                                          ┌─────────────────┐
                                          │ Puppeteer       │
                                          │ + React Render  │
                                          │ → High-Res PNG  │
                                          └─────────────────┘
```

**技術スタック:**
- Backend: Node.js + Express
- Rendering: Puppeteer (Chromium)
- Diagram: D3.js or React Flow
- Storage: S3 for caching
- Icons: Official CDN URLs

**実装ステップ:**
1. `/api/diagram/generate` エンドポイント作成
2. React ComponentをサーバーサイドでHTML化
3. Puppeteerで高解像度スクリーンショット
4. 結果をキャッシュ

### Option 2: React Flow への移行
```bash
npm install reactflow
npm install @xyflow/react
```

**メリット:**
- 現代的なReactベースライブラリ
- 優れたレンダリング品質
- カスタムノード簡単
- TypeScript完全サポート

**移行コード例:**
```typescript
import ReactFlow, { Node, Edge } from 'reactflow';
import 'reactflow/dist/style.css';

const CustomNode = ({ data }) => (
  <div className="bg-white p-4 rounded-lg shadow-xl border-2">
    <img src={data.iconUrl} alt="" className="w-16 h-16" />
    <div className="mt-2 font-semibold">{data.label}</div>
  </div>
);

const nodeTypes = { custom: CustomNode };
```

### Option 3: 公式CDNアイコン使用

**AWS Architecture Icons:**
```
https://d1.awsstatic.com/webteam/architecture-icons/Q1-2025/Arch_AWS-IAM_48.svg
```

**実装:**
```typescript
const OFFICIAL_ICON_URLS = {
  aws_iam_policy: 'https://d1.awsstatic.com/webteam/architecture-icons/Q1-2025/Arch_AWS-IAM_48.svg',
  aws_lambda: 'https://d1.awsstatic.com/webteam/architecture-icons/Q1-2025/Arch_AWS-Lambda_48.svg',
  // ...
};

<img src={OFFICIAL_ICON_URLS[resourceType]} />
```

## 推奨実装順序

### フェーズ1: 即座の改善 (1日)
1. React Flowへ移行
2. 公式CDNアイコン使用
3. ノードスタイル改善

### フェーズ2: サーバーサイド生成 (2-3日)
1. Backend APIセットアップ
2. Puppeteer統合
3. キャッシュ実装

### フェーズ3: 最適化 (継続)
1. パフォーマンス改善
2. アニメーション追加
3. エクスポート機能

## 推奨: まず React Flow で試す

**理由:**
- 最小の変更で大幅な品質向上
- サーバー不要
- 後からサーバーサイド生成追加可能

## 次のアクション

どの提案を試しますか？

1. **React Flow移行** (推奨・最速)
2. **サーバーサイド図生成** (最高品質)
3. **公式CDNアイコンのみ改善** (最小変更)
