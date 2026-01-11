# TFDrift-Falco UI - TODO List

## 📅 次の優先タスク

### 🎨 ビジュアル改善（継続）

#### 1. AWS公式アイコンの統合 ✅ **完了**
- [x] AWS Architecture Iconsをダウンロード（aws-iconsパッケージ使用）
- [x] `/public/aws-icons/`に主要サービスのアイコンを配置
- [x] cytoscapeStyles.tsでローカルアイコンパスに更新
- [x] 各AWSサービスタイプに対応するアイコンマッピング完成

#### 2. VPC/Subnet階層の可視化 ✅ **完了**
- [x] コンパウンドノード（親子関係）の実装
- [x] fcoseレイアウトの統合
- [x] VPC枠（破線）とSubnet枠（実線）の表示
- [x] リソースのSubnet内配置

#### 3. 構成図サイジング ✅ **完了**
- [x] 全ノードを構成図標準サイズに変更（40-50px）
- [x] スタイル統一（アイコン75%、フォント8-11px、ボーダー2-3px）
- [x] ノードスケール調整機能追加（スライダー＋プリセット）

#### 4. Graph Viewレイアウト改善（一部完了）
- [x] fcoseレイアウト最適化（コンパウンドノード対応）
- [x] ノードスケール調整機能
- [ ] エッジのラベル位置とフォントサイズ改善
- [ ] ノードのラベル表示方法改善（長い名前の省略表示）

#### 3. インタラクション改善
- [ ] ノードホバー時の詳細情報ツールチップ追加
- [ ] ノードクリック時のサイドパネルで詳細情報表示
- [ ] Drift状態のフィルタリング機能
  - [ ] 「Driftのみ表示」トグル
  - [ ] 重要度別フィルタ（Critical/High/Medium/Low）
- [ ] ノード検索機能（リソース名で検索）

#### 4. Legend（凡例）の改善
- [ ] より詳細なサービス分類
- [ ] インタラクティブな凡例（クリックでフィルタリング）
- [ ] 折りたたみ可能なLegend UI

### 📊 データ可視化強化

#### 5. Drift情報の統合改善
- [ ] Unmanagedリソースの明確な視覚化
- [ ] Modifiedリソースの差分表示
- [ ] Missingリソースの表示方法
- [ ] Drift検出時刻の表示

#### 6. グラフ分析機能
- [ ] Critical Pathの可視化
- [ ] Impact Radiusの表示（特定ノードの影響範囲）
- [ ] 依存関係の深さ表示
- [ ] 孤立ノードの検出と表示

### 🎭 Storybook強化

#### 7. CytoscapeGraphコンポーネントのStories作成
- [ ] `CytoscapeGraph.stories.tsx`作成
- [ ] 基本的なグラフ表示Story
- [ ] Drift状態のあるノードStory
- [ ] 大規模グラフ（100+ nodes）のStory
- [ ] 各種レイアウト（dagre, concentric, cose, grid）のStory

#### 8. DriftDashboardのStories改善
- [ ] モックデータでの表示確認
- [ ] ローディング状態のStory
- [ ] エラー状態のStory
- [ ] 空データ状態のStory

### 🔧 技術的改善

#### 9. パフォーマンス最適化
- [ ] 大規模グラフ（500+ nodes）でのパフォーマンステスト
- [ ] 仮想化/クラスタリングの検討
- [ ] レンダリング最適化

#### 10. TypeScript型定義の強化
- [ ] CytoscapeElements型の厳密化
- [ ] Drift関連型定義の整理
- [ ] API Response型の統合

## 📋 その他の改善項目（中〜低優先度）

### UI/UX
- [ ] ダークモード対応
- [ ] レスポンシブデザイン改善（モバイル対応）
- [ ] キーボードショートカット追加
- [ ] エクスポート機能（PNG, SVG, JSON）

### 機能追加
- [ ] グラフの保存/ロード機能
- [ ] 比較機能（前回のスキャンとの差分）
- [ ] アラート設定機能
- [ ] レポート生成機能

### テスト
- [ ] CytoscapeGraphコンポーネントのUnit Test
- [ ] Integration Test（API連携）
- [ ] E2E Test（Playwrightなど）

## ✅ 完了済み

### 🎉 v0.5.0 リリース（2026-01-03）

#### 主要機能追加:
- ✅ **VPC/Subnet階層の可視化**（最重要機能）
  - コンパウンドノード実装（親子関係のネスト表示）
  - VPC枠（破線、50pxパディング）
  - Subnet枠（実線、35pxパディング）
  - リソースのSubnet内配置
  - メタデータベースの親子関係構築（vpc_id, subnet_id活用）

- ✅ **fcoseレイアウト統合**
  - cytoscape-fcoseパッケージ追加
  - コンパウンドノード専用レイアウトエンジン
  - 階層構造に最適化されたパラメータ設定

- ✅ **ノードスケール調整機能**
  - スライダーコントロール（0.5x〜2.0x）
  - プリセットボタン（小0.7x、標準1.0x、大1.3x）
  - リアルタイム拡大縮小

#### 設計改善:
- ✅ **構成図サイジング**（全体再設計）
  - 全ノードを40-50pxに縮小（AWS構成図標準）
  - Small (40x40): Route Table, IAM Policy, CloudWatch, KMS, Secrets Manager
  - Medium (45x45): Security Group, RDS, ElastiCache, IAM Role, Gateways
  - Large (50x50): EKS, ECS, Load Balancer
  - デフォルトノード: 45x45

- ✅ **スタイル統一**
  - アイコンサイズ: 75%統一
  - フォントサイズ: 8-11px階層化（Small=8px, Medium=9px, Large=9px, VPC=11px）
  - ボーダー幅: 2-3px統一
  - テキストスタイル: text-background統一（text-outline廃止）

#### バグ修正:
- ✅ **黒塗りアイコン問題の解決**
  - デフォルトterraform_resourceスタイルを黒背景→白背景に変更
  - 不足リソースタイプ追加: aws_kms_alias, aws_route, aws_eks_node_group, aws_eks_addon

- ✅ **レイアウト選択機能の修正**
  - Layout radio buttonが選択できない問題を修正
  - currentLayout状態管理を追加

- ✅ **コンパウンドノードの透明度改善**
  - VPC背景: 30% → 60%
  - Subnet背景: 40% → 70%

### 2026-01-02
- ✅ **AWS公式アイコンの統合**（MUST要件）
  - aws-iconsパッケージ（v3.2.0）のインストール
  - 15種類のAWSサービスアイコンを/public/aws-icons/に配置
  - cytoscapeStyles.tsを更新してbackground-imageで公式アイコン表示

### 2026-01-01
- ✅ Graph View基本実装
- ✅ APIデータの直接使用（Cytoscape形式）
- ✅ Drift情報の統合（unmanaged, modified, missing）
- ✅ AWSサービス別のカラースキーム
- ✅ Legend（凡例）の追加
- ✅ Edge validation（無効なエッジの除外）

## 🐛 既知の問題

- ✅ ~~AWS CDN アイコンURL 403エラー~~ → **解決済み**（ローカルアイコン使用）
- ✅ ~~黒塗りアイコン問題~~ → **解決済み**（デフォルトスタイル改善）
- ⚠️ 一部のノードラベルが長すぎて読みにくい → 省略表示機能追加予定

## 📝 メモ

- Storybookは既にインストール済み（v10.1.11）
- CytoscapeGraphのStoriesはまだ未作成
- AWS Architecture Iconsのライセンス確認済み（商用利用可能）
- fcoseレイアウトがVPC/Subnet階層に最適

---

**最終更新**: 2026-01-03
**現在のバージョン**: v0.5.0
**今回の成果**: VPC/Subnet階層可視化・構成図サイジング・スケール調整機能
**次回セッション**: インタラクション改善（ツールチップ、詳細パネル）
