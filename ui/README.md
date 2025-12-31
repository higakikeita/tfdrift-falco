# TFDrift-Falco UI

> **Cloud Infrastructure Security & Drift Analysis Visualization**
>
> クラウドインフラのドリフトとセキュリティイベントの因果関係を可視化する次世代UIプラットフォーム

[![TypeScript](https://img.shields.io/badge/TypeScript-5.9.3-blue)](https://www.typescriptlang.org/)
[![React](https://img.shields.io/badge/React-19.2.0-61dafb)](https://react.dev/)
[![Vite](https://img.shields.io/badge/Vite-7.2.4-646cff)](https://vitejs.dev/)
[![React Flow](https://img.shields.io/badge/React_Flow-11.11.4-ff69b4)](https://reactflow.dev/)

---

## 📖 目次

- [概要](#概要)
- [主要機能](#主要機能)
- [クイックスタート](#クイックスタート)
- [開発環境構築](#開発環境構築)
- [プロジェクト構造](#プロジェクト構造)
- [技術スタック](#技術スタック)
- [使い方](#使い方)
- [テスト](#テスト)
- [ビルド＆デプロイ](#ビルドデプロイ)
- [トラブルシューティング](#トラブルシューティング)
- [ドキュメント](#ドキュメント)

---

## 🎯 概要

**TFDrift-Falco UI**は、クラウドインフラの変更（Terraform Drift）からセキュリティインシデント（Falco Events）までの因果関係を可視化するWebアプリケーションです。

### 解決する課題

```
Terraform Drift → IAM変更 → ServiceAccount → Pod → Container → Falco Event
```

この因果関係チェーンを**インタラクティブなグラフ**で可視化し、「なぜそのイベントが発生したのか」を瞬時に理解できます。

---

## ✨ 主要機能

### 📊 インタラクティブグラフ可視化
- React Flow based高性能グラフエンジン
- 公式クラウドアイコン使用
- 複数レイアウトアルゴリズム（階層、放射状、力学モデル、AWS階層図）

### 🎯 依存関係追跡
- 依存先/依存元の可視化
- 影響範囲分析（1〜5ホップ）
- パターン検索

### 🔍 高度な検索＆フィルタリング
- テキスト検索
- 深刻度フィルター（Critical/High/Medium/Low）
- リソースタイプフィルター
- リアルタイムフィルタリング

### ⚡ リアルタイム更新
- WebSocket双方向通信
- SSEイベントストリーム
- 自動再接続

### 🎓 オンボーディング＆ヘルプ
- ウェルカムモーダル（6ステップガイド）
- キーボードショートカットガイド
- コンテキストヘルプオーバーレイ

### その他
- 🌙 完全ダークモード対応
- 📤 PNG/SVGエクスポート
- 📈 ドリフトヒストリーテーブル

---

## 🚀 クイックスタート

### 前提条件
- Node.js 20.x以上
- npm 10.x以上
- TFDrift Backend APIが起動していること

### インストール

```bash
# リポジトリのクローン
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco/ui

# 依存関係のインストール
npm install

# 環境変数の設定
cp .env.example .env
# .envファイルを編集してAPIエンドポイントを設定

# 開発サーバーの起動
npm run dev
```

ブラウザで http://localhost:5173 を開きます。

---

## 🛠️ 開発環境構築

### 環境変数の設定

`.env`ファイルを作成：

```bash
# Backend API URL
VITE_API_URL=http://localhost:8080

# WebSocket URL
VITE_WS_URL=ws://localhost:8080/ws

# SSE URL
VITE_SSE_URL=http://localhost:8080/events
```

### 開発サーバーの起動

```bash
npm run dev
```

### ビルド

```bash
# プロダクションビルド
npm run build

# ビルド結果のプレビュー
npm run preview
```

---

## 📁 プロジェクト構造

```
ui/
├── src/
│   ├── api/                    # API統合層
│   ├── components/             # Reactコンポーネント
│   │   ├── reactflow/          # React Flowグラフ
│   │   ├── graph/              # グラフUI要素
│   │   ├── onboarding/         # オンボーディング
│   │   └── ui/                 # 基本UIコンポーネント
│   ├── hooks/                  # カスタムフック
│   ├── types/                  # TypeScript型定義
│   ├── utils/                  # ユーティリティ関数
│   └── App-final.tsx           # メインアプリ
│
├── docs/                       # ドキュメント
│   ├── PROJECT_STRUCTURE.md    # プロジェクト構造詳細
│   └── ARCHITECTURE.md         # アーキテクチャ詳細
│
└── package.json
```

詳細は [`docs/PROJECT_STRUCTURE.md`](./docs/PROJECT_STRUCTURE.md) を参照。

---

## 🏗️ 技術スタック

| Category | Technology | Version |
|----------|-----------|---------|
| **Core** | React | 19.2.0 |
| | TypeScript | 5.9.3 |
| | Vite | 7.2.4 |
| **State** | React Query | 5.90.12 |
| | Zustand | 5.0.9 |
| **UI** | Tailwind CSS | 4.1.18 |
| | shadcn/ui | Latest |
| **Graph** | React Flow | 11.11.4 |
| | Dagre | 1.1.8 |

---

## 📚 使い方

### 基本操作

#### グラフの操作
- **パン**: ドラッグで移動
- **ズーム**: マウスホイール
- **フィット**: `F`キー
- **中央配置**: `C`キー

#### ノードの操作
- **左クリック**: ノード詳細パネル
- **ダブルクリック**: フォーカスビュー
- **右クリック**: コンテキストメニュー

### キーボードショートカット

| キー | 機能 |
|-----|------|
| `F` | グラフ全体をフィット |
| `C` | グラフを中央に配置 |
| `ESC` | 詳細パネルを閉じる |
| `?` | ショートカット一覧 |

---

## 🧪 テスト

```bash
# 全テスト実行
npm run test

# ウォッチモード
npm run test:watch

# カバレッジ
npm run test:coverage
```

目標カバレッジ: **60%以上**

---

## 📦 ビルド＆デプロイ

### ローカルビルド

```bash
npm run build
```

出力: `dist/` ディレクトリ

### Docker

```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

---

## 🔧 トラブルシューティング

### APIに接続できない

**解決策**:
1. Backend APIが起動しているか確認
2. `.env`の`VITE_API_URL`を確認
3. CORS設定を確認

### グラフが表示されない

**解決策**:
1. ブラウザの開発者ツールでエラーを確認
2. Networkタブでデータ取得を確認
3. React Dev Toolsで状態を確認

### パフォーマンスが悪い

**解決策**:
1. ノード数を確認（1000以上はClustering推奨）
2. ブラウザのGPUアクセラレーション有効化
3. React Profilerでボトルネック特定

---

## 📖 ドキュメント

- [`PROJECT_STRUCTURE.md`](./docs/PROJECT_STRUCTURE.md) - プロジェクト構造詳細
- [`ARCHITECTURE.md`](./docs/ARCHITECTURE.md) - システムアーキテクチャ
- [`CONTRIBUTING.md`](./docs/CONTRIBUTING.md) - 貢献ガイドライン（作成予定）

---

## 📄 ライセンス

MIT License - Copyright (c) 2026 TFDrift-Falco Team

---

**Built with ❤️ by the TFDrift-Falco Team**
