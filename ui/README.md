# TFDrift-Falco UI

**Real-time Terraform Drift Detection - React Web UI**

[![Version](https://img.shields.io/badge/version-0.5.0-blue)](../CHANGELOG.md)
[![React](https://img.shields.io/badge/React-19.2-61DAFB?logo=react)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.9-3178C6?logo=typescript)](https://www.typescriptlang.org/)
[![Vite](https://img.shields.io/badge/Vite-7.2-646CFF?logo=vite)](https://vite.dev/)
[![Storybook](https://img.shields.io/badge/Storybook-10.1-FF4785?logo=storybook)](https://storybook.js.org/)

> 🎉 **v0.5.0+** - Storybook駆動開発、AWS公式アイコン28個統合、VPC/Subnet階層表示を実現！

A modern React web UI for TFDrift-Falco, featuring interactive graph visualization with Cytoscape.js, real-time drift detection dashboard, and Storybook-driven development.

---

## 📑 目次

- [概要](#-概要)
- [クイックスタート](#-クイックスタート)
- [開発環境構築](#️-開発環境構築)
- [Storybook駆動開発](#-storybook駆動開発)
- [主要コンポーネント](#-主要コンポーネント)
- [モックデータ](#-モックデータ)
- [ディレクトリ構造](#-ディレクトリ構造)
- [技術スタック](#-技術スタック)
- [スクリプト一覧](#-スクリプト一覧)
- [トラブルシューティング](#-トラブルシューティング)

---

## 🌟 概要

TFDrift-Falco UIは、Terraform Drift検知を可視化するインタラクティブなWebインターフェースを提供します。

### 主要機能

- **📊 インタラクティブグラフ可視化**
  - Cytoscape.js ベースのグラフレンダリング
  - VPC/Subnet階層表示（Compound Nodes）
  - 28種類のAWS公式アイコン
  - 5種類のレイアウトアルゴリズム (fcose, dagre, concentric, cose, grid)
  - ズーム、パン、ノード選択

- **🎯 Drift検知ダッシュボード**
  - 全体Drift状態表示
  - サマリーカード (Terraform管理リソース、未管理、不足、変更)
  - リソースタイプ別の内訳表示
  - カラーコーディング
  - リアルタイム更新対応 (WebSocket/SSE - 準備完了)

- **⚙️ DisplayOptionsパネル**
  - ドラッグ可能パネル（位置カスタマイズ）
  - レイアウト切り替え
  - フィルターモード (全リソース、Driftのみ、VPC/ネットワークのみ)
  - レジェンド表示 (28種類のAWSサービス)

- **🚀 Storybook駆動開発**
  - 17個の包括的なStory
  - 30倍高速なフィードバックループ (2分 → 4秒)
  - モックデータ統合
  - 生きたドキュメント

---

## 🚀 クイックスタート

### 前提条件

- **Node.js**: v20+ (推奨 v22+)
- **npm**: v10+ または **pnpm**: v9+
- **Backend API**: `http://localhost:8080` で起動 (開発時はオプション)

### インストール

```bash
# UIディレクトリに移動
cd ui

# 依存関係のインストール
npm install

# 開発サーバー起動
npm run dev
```

開発サーバーは **http://localhost:5173/** で起動します。

### Storybookへのアクセス

```bash
# Storybook起動 (別ターミナル)
npm run storybook
```

Storybookは **http://localhost:6006/** で利用できます。

---

## 🛠️ 開発環境構築

### 開発モード

```bash
npm run dev
```

Vite開発サーバーが起動し、以下の機能が有効になります：
- Hot Module Replacement (HMR)
- Fast Refresh
- TypeScript型チェック
- ESLint リンティング

### モックデータの使用

UIは、バックエンドなしで開発できるモックデータをサポートしています。

**`src/App-drift.tsx` で設定:**

```typescript
// モックデータを有効化
const USE_MOCK_GRAPH_DATA = true;
const USE_MOCK_DRIFT_DATA = true;
```

モックデータの場所：
- `src/mocks/graphData.ts` - グラフ可視化モックデータ (282行)
- `src/mocks/driftData.ts` - Drift検知モックデータ (366行)

### バックエンド統合

実際のAPIデータを使用する場合：

```typescript
// モックデータを無効化
const USE_MOCK_GRAPH_DATA = false;
const USE_MOCK_DRIFT_DATA = false;
```

バックエンドAPIを起動：

```bash
# プロジェクトルートから
cd ../
docker-compose up -d backend
```

APIエンドポイント：
- `GET /api/v1/discovery/drift` - Drift検知データ
- `GET /api/v1/discovery/drift/summary` - Driftサマリー
- `GET /api/v1/graph` - グラフデータ (Cytoscape形式)

---

## 📚 Storybook駆動開発

TFDrift-Falco UIは、**Storybook駆動開発 (SDD)** 手法を採用しています。

### なぜStorybookなのか？

- **30倍高速なフィードバックループ**: 2分 → 4秒
- **生きたドキュメント**: Storyそのものがドキュメント
- **分離されたコンポーネント開発**: バックエンド不要
- **ビジュアルリグレッションテスト**: Chromatic/Percy対応準備完了

### 利用可能なStory (全17個)

#### 基本Story
1. **Default** - デフォルトグラフ（30ノード）
2. **Empty** - 空の状態
3. **VPC Hierarchy** - VPC/Subnet階層（Compound Nodes）

#### レイアウトStory
4. **Layout: fcose** (デフォルト)
5. **Layout: dagre**
6. **Layout: cose**
7. **Layout: grid**

#### サイズStory
8. **Small Graph** - 10ノード
9. **Medium Graph** - 30ノード
10. **Large Graph** - 100ノード
11. **Very Large Graph** - 200ノード

#### Drift Story
12. **Drift Highlighted** - Driftノードを赤色表示

#### AWSアイコンStory
13. **All AWS Services** - 28種類のAWSサービスタイプ紹介

#### インタラクティブStory
14. **Interactive: Node Click** - ノードクリック動作
15. **Interactive: Edge Click** - エッジクリック動作
16. **Interactive: Path Highlighting** - パスハイライト

#### Playground
17. **Playground** - ライブコントロールで実験

### Storybookの実行

```bash
# Storybook開発サーバー起動
npm run storybook

# 静的Storybook構築
npm run build-storybook
```

### SDDガイドライン

詳細なSDD実践方法については以下を参照：
- [Storybook駆動開発ガイド](docs/STORYBOOK_DRIVEN_DEVELOPMENT.md)
- [Qiita記事 (下書き)](docs/QIITA_STORYBOOK_DRIVEN_DEVELOPMENT.md)

---

## 🧩 主要コンポーネント

### メインコンポーネント

#### **CytoscapeGraph** (`src/components/CytoscapeGraph.tsx`)

Cytoscape.jsを使用したコアグラフ可視化コンポーネント。

**Props:**
- `elements` - グラフ要素（ノードとエッジ）
- `layout` - レイアウトアルゴリズム ('fcose', 'dagre', など)
- `filterMode` - フィルターモード ('all', 'driftOnly', 'vpcOnly')
- `highlightDriftNodes` - Driftノードのハイライト有無
- `onNodeClick` - ノードクリックハンドラ
- `onEdgeClick` - エッジクリックハンドラ

**機能:**
- VPC/Subnet階層 (Compound Nodes)
- 28種類のAWS公式アイコン
- ドラッグ可能なDisplayOptionsパネル
- リアルタイムフィルタリング
- レジェンド表示

**Storyファイル:** `src/components/CytoscapeGraph.stories.tsx` (459行)

#### **DriftDashboard** (`src/components/DriftDashboard.tsx`)

Drift検知状態ダッシュボード。

**機能:**
- 全体Drift状態
- サマリーカード
- リソースタイプ別内訳
- 色分けされた重要度表示

### サポートコンポーネント

- **DisplayOptions** - ドラッグ可能オプションパネル (CytoscapeGraphに埋め込み)
- **Legend** - AWSサービスレジェンド (DisplayOptionsに埋め込み)

---

## 🗂️ モックデータ

### グラフモックデータ (`src/mocks/graphData.ts`)

グラフ可視化テスト用の再利用可能なモックデータ。

**利用可能なモックデータ:**
- `mockGraphDataSmall` - 10ノード
- `mockGraphDataDefault` - 30ノード (VPC + Subnet + リソース)
- `mockGraphDataLarge` - 100ノード
- `mockGraphDataVeryLarge` - 200ノード
- `mockGraphDataVPCHierarchy` - VPC/Subnet重点
- `mockGraphDataAllAWSServices` - 28種類のAWSサービスタイプ

**使用方法:**
```typescript
import { mockGraphDataDefault } from '@/mocks/graphData';

const elements = mockGraphDataDefault;
```

### Driftモックデータ (`src/mocks/driftData.ts`)

Drift検知シナリオ用のモックデータ。

**利用可能なモックデータ:**
- `mockDriftSummaryWithDrift` - Drift検知シナリオ
- `mockDriftSummaryClean` - Driftなしシナリオ
- `mockDriftDetectionWithDrift` - 詳細Driftデータ (8未管理, 3不足, 5変更)
- `mockDriftDetectionClean` - クリーン状態

**使用方法:**
```typescript
import { mockDriftSummaryWithDrift, mockDriftDetectionWithDrift } from '@/mocks/driftData';

const summary = mockDriftSummaryWithDrift;
const detection = mockDriftDetectionWithDrift;
```

---

## 📁 ディレクトリ構造

```
ui/
├── public/                     # 静的アセット
│   └── aws-icons/              # 28個のAWS公式SVGアイコン
│       └── README.md           # アイコンリストとライセンス情報
├── src/
│   ├── components/             # Reactコンポーネント
│   │   ├── CytoscapeGraph.tsx  # メイングラフコンポーネント
│   │   ├── CytoscapeGraph.stories.tsx  # 17個のStorybook Story
│   │   └── DriftDashboard.tsx  # Driftダッシュボード
│   ├── mocks/                  # 開発用モックデータ
│   │   ├── graphData.ts        # グラフモックデータ (282行)
│   │   └── driftData.ts        # Driftモックデータ (366行)
│   ├── styles/                 # スタイリング
│   │   └── cytoscapeStyles.ts  # Cytoscapeスタイル (28 AWSサービス)
│   ├── App-drift.tsx           # メインアプリ (Drift検知)
│   ├── App.tsx                 # メインアプリエントリーポイント
│   ├── main.tsx                # Reactエントリーポイント
│   └── index.css               # グローバルスタイル (Tailwind)
├── docs/                       # ドキュメント
│   ├── STORYBOOK_DRIVEN_DEVELOPMENT.md  # SDDガイドライン (289行)
│   ├── QIITA_STORYBOOK_DRIVEN_DEVELOPMENT.md  # Qiita記事下書き
│   └── ARCHITECTURE.md         # アーキテクチャ (作成予定)
├── .storybook/                 # Storybook設定
├── package.json                # NPM依存関係
├── tsconfig.json               # TypeScript設定
├── vite.config.ts              # Vite設定
├── tailwind.config.ts          # Tailwind CSS設定
└── README.md                   # このファイル
```

---

## 🔧 技術スタック

### コア技術

| 技術 | バージョン | 用途 |
|---|---|---|
| **React** | 19.2 | UIフレームワーク |
| **TypeScript** | 5.9 | 型安全性 |
| **Vite** | 7.2 | ビルドツール & 開発サーバー |
| **Tailwind CSS** | 4.1 | ユーティリティファーストCSS |

### グラフ可視化

| ライブラリ | バージョン | 用途 |
|---|---|---|
| **Cytoscape.js** | 3.33 | グラフレンダリング |
| **cytoscape-fcose** | 2.2 | 力学的レイアウト |
| **cytoscape-dagre** | 2.5 | 階層的レイアウト |
| **AWS Icons** | 3.2 | AWS公式アイコン |

### 状態管理 & データ取得

| ライブラリ | バージョン | 用途 |
|---|---|---|
| **React Query** | 5.90 | APIデータ管理 |
| **Zustand** | 5.0 | グローバル状態管理 |

### 開発ツール

| ツール | バージョン | 用途 |
|---|---|---|
| **Storybook** | 10.1 | コンポーネント開発 |
| **Vitest** | 4.0 | ユニットテスト |
| **Playwright** | 1.57 | E2Eテスト |
| **ESLint** | 9.39 | リンティング |
| **Lighthouse** | 13.0 | パフォーマンス監査 |

### UIコンポーネント

| ライブラリ | 用途 |
|---|---|
| **Radix UI** | アクセシブルなコンポーネントプリミティブ |
| **Lucide React** | アイコンライブラリ |
| **class-variance-authority** | バリアントスタイル |
| **clsx** / **tailwind-merge** | クラスユーティリティ |

---

## 📜 スクリプト一覧

### 開発

```bash
# 開発サーバー起動
npm run dev

# Storybook起動
npm run storybook
```

### テスト

```bash
# ユニットテスト実行 (Vitest)
npm run test

# ウォッチモードでテスト実行
npm run test:watch

# カバレッジ付きテスト実行
npm run test:coverage

# UIでテスト実行
npm run test:ui

# E2Eテスト実行 (Playwright)
npm run test:e2e

# UIでE2Eテスト実行
npm run test:e2e:ui

# E2Eテストデバッグ
npm run test:e2e:debug
```

### ビルド

```bash
# プロダクションビルド
npm run build

# プロダクションビルドプレビュー
npm run preview

# Storybookビルド
npm run build-storybook
```

### コード品質

```bash
# リンティング
npm run lint

# Lighthouse監査実行
npm run lighthouse

# Lighthouseデータ収集
npm run lighthouse:collect

# Lighthouseバジェット確認
npm run lighthouse:assert
```

---

## 🐛 トラブルシューティング

### よくある問題

#### 1. **ポート5173が既に使用中**

```bash
# ポート5173を使用しているプロセスを終了
lsof -ti:5173 | xargs kill -9

# または vite.config.ts でポート変更
export default defineConfig({
  server: { port: 5174 }
})
```

#### 2. **Storybookポート6006が既に使用中**

```bash
# プロセスを終了
lsof -ti:6006 | xargs kill -9

# または別ポートを指定
npm run storybook -- -p 6007
```

#### 3. **API接続失敗 (CORSエラー)**

バックエンドAPIが起動し、CORSが設定されていることを確認：

```bash
# バックエンドログを確認
docker-compose logs backend
```

バックエンドは `http://localhost:5173` originを許可する必要があります。

#### 4. **グラフが表示されない**

ブラウザコンソールでエラーを確認してください。よくある原因：
- Cytoscape.jsスタイルの欠落
- 無効なグラフデータ構造
- レイアウトアルゴリズムが読み込まれていない

**解決方法:**
```typescript
// 必要なレイアウトをインポート
import fcose from 'cytoscape-fcose';
import dagre from 'cytoscape-dagre';
cytoscape.use(fcose);
cytoscape.use(dagre);
```

#### 5. **AWSアイコンが表示されない**

`public/aws-icons/` にアイコンが存在することを確認：

```bash
ls -la public/aws-icons/
```

アイコンはSVGファイルである必要があります。欠落している場合、`aws-icons` パッケージから再生成してください。

#### 6. **IDEでTypeScriptエラー**

```bash
# TypeScriptサーバー再起動 (VSCode)
Cmd+Shift+P → "TypeScript: Restart TS Server"

# またはTypeScriptプロジェクト再構築
npm run build
```

#### 7. **モックデータが読み込まれない**

`App-drift.tsx` でフラグが正しく設定されているか確認：

```typescript
const USE_MOCK_GRAPH_DATA = true;
const USE_MOCK_DRIFT_DATA = true;
```

#### 8. **グラフレンダリングが遅い (100+ノード)**

最適化方法：
- fcoseレイアウトのイテレーション数を減らす
- よりシンプルなレイアウト (dagre) を使用
- フィルタリングを実装
- Level of Detail (LOD) を有効化

```typescript
const layoutOptions = {
  name: 'fcose',
  numIter: 1000,  // 3000から削減
  nodeRepulsion: 4500,
  idealEdgeLength: 100
};
```

---

## 📚 追加リソース

- [メインプロジェクトREADME](../README.md)
- [Storybook駆動開発ガイド](docs/STORYBOOK_DRIVEN_DEVELOPMENT.md)
- [アーキテクチャドキュメント](docs/ARCHITECTURE.md) (作成予定)
- [プロジェクトロードマップ](../PROJECT_ROADMAP.md)
- [CHANGELOG](../CHANGELOG.md)

---

## 📝 貢献

貢献を歓迎します！以下を参照してください：
- [貢献ガイド](../docs/CONTRIBUTING.md)
- [行動規範](../CODE_OF_CONDUCT.md) (存在する場合)

---

## 📄 ライセンス

MITライセンス - 詳細は [../LICENSE](../LICENSE) を参照

---

## 🙋 サポート

- **Issues**: [GitHub Issues](https://github.com/higakikeita/tfdrift-falco/issues)
- **ドキュメント**: [docs/](../docs/)
- **Discussions**: [GitHub Discussions](https://github.com/higakikeita/tfdrift-falco/discussions)

---

---

**作成者**: Keita Higaki
**Built with ❤️ using Storybook-Driven Development**
