# TFDrift-Falco TODO List

**最終更新**: 2026-01-10
**現在バージョン**: v0.5.0+
**フォーカス**: UI改善・Storybook駆動開発

> 📊 **詳細な現状分析**: [STATUS_REPORT_2026-01-10.md](./STATUS_REPORT_2026-01-10.md) を参照

---

## 🔥 高優先度 (今すぐ対応)

### 1. 小規模AWS環境の再構築
**目的**: 実データでの動作確認と視認性確保

**背景**:
- 現在: 119ノードで視認性が低い
- モックデータ (30ノード) は完璧に動作
- 実データでVPC/Subnet階層を確認したい

**タスク**:
- [ ] `terraform/minimal-environment` ディレクトリ作成
- [ ] minimal 環境の設計 (10-20リソース)
  ```
  - VPC x1
  - Subnet x2 (public, private)
  - EKS or ECS x1
  - RDS x1
  - S3 x1-2
  - Security Groups (最小限)
  - IAM Roles (最小限)
  ```
- [ ] main.tf, variables.tf, outputs.tf 作成
- [ ] terraform apply で環境構築
- [ ] Backend API で実データ取得確認
- [ ] UI での表示確認 (VPC/Subnet階層)

**期待される成果**: Storybook と同等の視認性で実データが表示される

---

### 2. ドキュメント整合性の修正

#### 2.1 CHANGELOG.md の更新
**問題**: v0.5.0 が「GCP Support」と記載されているが、実際は「UI改善」

**対応**:
- [ ] v0.5.0+ セクションを追加
- [ ] UI改善内容を記載
  - Storybook駆動開発 (SDD) 実装
  - AWS公式アイコン28個統合
  - ビジュアル改善 (ノードサイズ、VPC/Subnet階層)
  - Drift Detection表示機能
  - DisplayOptions改善 (ドラッグ可能、フィルター)
- [ ] v0.5.0 を v0.5.0-gcp に改名 (将来の参照用)

#### 2.2 TODO.md の更新
- [x] 現状に合わせて全面刷新 ← **このファイル**

#### 2.3 README.md の修正
- [ ] v0.5.0+ の説明を追加
- [ ] UI改善のスクリーンショット追加
- [ ] Storybook のリンク追加

#### 2.4 UI専用ドキュメントの作成
- [ ] `/ui/README.md` 作成
  - 開発環境セットアップ
  - Storybook の使い方
  - モックデータとの切り替え
  - コンポーネント設計思想
- [ ] `/ui/docs/ARCHITECTURE.md` 作成
  - コンポーネント構成図
  - データフロー (API → State → UI)
  - State管理方針
  - CSS/スタイリング

---

### 3. Storybook記事の公開準備

#### 3.1 記事のレビュー
- [ ] `QIITA_STORYBOOK_DRIVEN_DEVELOPMENT.md` の内容確認
- [ ] 技術的な正確性チェック
- [ ] 文章の推敲

#### 3.2 スクリーンショット準備
- [ ] Before/After 比較画像
- [ ] 17 Stories の一覧
- [ ] VPC/Subnet 階層の表示例
- [ ] AWS公式アイコン一覧
- [ ] DisplayOptions パネル

#### 3.3 投稿
- [ ] GitHub リポジトリ URL 確認
- [ ] Qiita に投稿
- [ ] Twitter で共有
- [ ] README.md に記事リンク追加

---

## 📚 中優先度 (今週中)

### 4. UI機能の拡充

#### 4.1 フィルター機能の強化
- [ ] リソースタイプ別フィルター
  - Compute (EKS, ECS, Lambda)
  - Database (RDS, DynamoDB, ElastiCache)
  - Network (VPC, Subnet, SG, ALB)
  - Storage (S3)
- [ ] タグベースフィルター
- [ ] 検索機能 (リソース名、ID)

#### 4.2 ノード詳細表示
- [ ] クリックでリソース詳細パネル表示
  - 基本情報 (ID, Type, Name)
  - Attributes
  - Tags
  - Dependencies (incoming/outgoing)
- [ ] Drift情報の表示 (severity, changes)

#### 4.3 UX改善
- [ ] ズーム・パン操作の最適化
- [ ] ノード選択時のハイライト
- [ ] エクスポート機能 (PNG, SVG, JSON)
- [ ] レイアウトの保存/復元

---

### 5. テストの追加

#### 5.1 ユニットテスト
- [ ] `CytoscapeGraph.test.tsx`
  - Props の正しい処理
  - Filter mode の動作
  - Layout 切り替え
- [ ] `DriftDashboard.test.tsx`
  - データ表示の正確性
  - カウント計算の正確性
- [ ] `graphData.test.ts`
  - モックデータの構造検証
- [ ] `driftData.test.ts`
  - モックデータの型チェック

#### 5.2 統合テスト
- [ ] API統合テスト
  - モックデータ vs 実データ
  - エラーハンドリング
- [ ] Storybook のビジュアルリグレッションテスト
  - Chromatic または Percy

---

### 6. パフォーマンス最適化

#### 6.1 大規模グラフ対応
- [ ] Level of Detail (LOD) 実装
  - 遠くのノードは簡略表示
- [ ] Clustering 実装
  - 同じSubnet内のノードをグループ化
- [ ] Virtual Scrolling (テーブルビュー)

#### 6.2 レイアウト最適化
- [ ] fcose レイアウトの設定チューニング
- [ ] Web Worker での計算
- [ ] キャッシュ機構

---

## 🚀 低優先度 (将来)

### 7. リアルタイム更新

#### 7.1 WebSocket 統合
- [ ] WebSocket クライアント実装
- [ ] リアルタイムイベント受信
- [ ] グラフの自動更新
- [ ] トースト通知

#### 7.2 SSE (Server-Sent Events) 統合
- [ ] SSE クライアント実装
- [ ] イベントストリーム処理
- [ ] 再接続ロジック

---

### 8. 高度な機能

#### 8.1 Drift 履歴
- [ ] タイムライン表示
- [ ] 変更履歴の diff 表示
- [ ] ロールバック提案

#### 8.2 分析機能
- [ ] リソース依存関係分析
- [ ] Impact Radius (影響範囲) 表示
- [ ] Critical Path 検出

#### 8.3 レポート生成
- [ ] PDF エクスポート
- [ ] Excel エクスポート
- [ ] Drift サマリーレポート

---

## 📦 完了済み (v0.5.0+)

### ✅ Storybook駆動開発 (SDD)
- [x] 17個のStoryを作成
- [x] モックデータ完備 (graphData.ts, driftData.ts)
- [x] SDD指針ドキュメント作成
- [x] Qiita記事下書き作成

### ✅ AWS公式アイコン統合
- [x] 28個のAWS Architecture Icons導入
- [x] `/ui/public/aws-icons/README.md` 作成

### ✅ ビジュアル改善
- [x] ノードサイズ拡大 (45px→60px)
- [x] VPC/Subnet階層の視認性向上
- [x] フォントサイズ改善 (+2-3px)
- [x] アイコンサイズ拡大 (75%→80%)

### ✅ Drift Detection表示
- [x] モックデータ作成
- [x] DriftDashboard コンポーネント動作確認
- [x] Summary Cards 表示
- [x] Resource Type Breakdown 表示

### ✅ DisplayOptions改善
- [x] ドラッグ可能なパネル
- [x] 閉じるボタン (×)
- [x] Layout切り替え
- [x] Filter Mode (All, Drift Only, VPC/Network Only)
- [x] Legend更新 (28サービス、2カラム)

### ✅ AWS環境クリーンアップ
- [x] production-like-environment 削除 (116リソース)

---

## 📋 優先順位サマリー

### 🔴 今日中
1. 小規模AWS環境の設計
2. CHANGELOG.md の更新開始

### 🟡 今週中
3. 小規模環境のデプロイ
4. ドキュメント修正完了
5. Qiita記事投稿

### 🟢 来週以降
6. UI機能拡充
7. テスト追加
8. パフォーマンス最適化

---

## 💡 技術メモ

### Storybook駆動開発の効果
- **開発速度**: 30倍高速化 (フィードバックループ: 2分 → 4秒)
- **品質**: ビジュアルリグレッション防止
- **ドキュメント**: Story自体が生きたドキュメント
- **チーム連携**: デザイナーとの協業が容易

### VPC/Subnet階層表示のポイント
- **Compound Nodes**: Cytoscape.js の parent 機能
- **fcose レイアウト**: compound nodes に最適化
- **視認性**: background opacity、border、padding の調整が重要

### 大規模グラフの課題
- **100+ノード**: フィルター機能が必須
- **レイアウト**: fcose は遅い (3000 iterations)
- **UX**: ズーム・パン、検索機能が重要

### モックデータの利点
- 開発速度の向上
- 一貫性のあるテストデータ
- バックエンド依存なし
- ビジュアル確認が容易

---

## 🔗 関連ドキュメント

- [現状分析レポート](./STATUS_REPORT_2026-01-10.md)
- [Storybook駆動開発指針](/ui/docs/STORYBOOK_DRIVEN_DEVELOPMENT.md)
- [Qiita記事下書き](/ui/docs/QIITA_STORYBOOK_DRIVEN_DEVELOPMENT.md)
- [AWS Icons README](/ui/public/aws-icons/README.md)

---

**作成者**: Keita Higaki
**最終更新**: 2026-01-10
**セッション**: UI改善・Storybook駆動開発
