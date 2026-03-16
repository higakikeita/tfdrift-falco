# TFDrift-Falco Backend TODO

**最終更新**: 2026-03-16
**現在バージョン**: v0.5.0+
**フォーカス**: Phase 1 安定化

> 📊 **詳細な現状分析**: [STATUS_REPORT_2026-01-10.md](./STATUS_REPORT_2026-01-10.md) を参照
> 🗺️ **ロードマップ**: [PROJECT_ROADMAP.md](./PROJECT_ROADMAP.md) を参照

---

## 🔥 今すぐやるべき（Phase 1）

### API改善
- [x] エラーハンドリング強化（AWS認証エラー、Terraform State失敗時の適切なレスポンス）
- [ ] GET /health エンドポイント改善（依存サービスのヘルスチェック追加）
- [ ] ログレベル最適化（Debug/Info/Warn/Error の使い分け統一）
- [ ] API レスポンスタイム改善

### Falcoルール修正
- [x] rules/terraform_drift.yaml 構文エラー修正
- [x] CloudTrailフィールド名修正（ct.name, ct.user, ct.region）
- [x] Priority値確認（IAM=CRITICAL, その他=WARNING）
- [ ] Falcoルールのテスト（falco -V による検証）
- [ ] ルールの網羅性向上（EKS, ECS, DynamoDB等の追加）

### GraphDB最適化
- [ ] クエリパフォーマンス改善（大規模グラフ 100+ノード対応）
- [ ] インデックス追加
- [ ] キャッシュ機構の実装

---

## 🔴 今月中（Phase 1 継続）

### テスト強化
- [ ] API統合テスト追加（全エンドポイント）
- [ ] E2Eテスト基盤構築
- [ ] UIユニットテスト（CytoscapeGraph, DriftDashboard）

### ドキュメント整備
- [ ] CHANGELOG.md にv0.5.0+ UI改善を追記
- [ ] README.md のバージョン説明更新
- [ ] /ui/README.md 作成
- [ ] /ui/docs/ARCHITECTURE.md 作成
- [ ] docs/ 配下の古いドキュメント整理

### 小規模AWS環境構築
- [ ] terraform/minimal-environment 設計（10-20リソース）
- [ ] main.tf, variables.tf, outputs.tf 作成
- [ ] 実データでのUI動作確認

### Docker Compose改善
- [ ] version フィールド削除
- [ ] AWS_PROFILE環境変数追加
- [ ] Falco依存関係の条件付き化
- [ ] ヘルスチェック改善

---

## 🟡 来月（Phase 2: 機能拡張）

### リアルタイム通信
- [ ] WebSocketクライアント完全実装
- [ ] SSE (Server-Sent Events) 完全実装
- [ ] リアルタイムイベント配信
- [ ] 再接続ロジック

### 高度な分析機能
- [ ] Drift履歴タイムライン表示
- [ ] 変更履歴の diff 表示
- [ ] Impact Radius（影響範囲）表示
- [ ] Critical Path 検出

### UI機能拡充
- [ ] フィルター機能強化（リソースタイプ別、タグベース）
- [ ] ノード詳細パネル
- [ ] エクスポート機能（PNG, SVG, JSON）
- [ ] ズーム・パン最適化

---

## 🟢 将来（Phase 3+）

### マルチクラウド
- [ ] GCP統合テスト完了・安定化
- [ ] Azure対応開始（Activity Logs パーサー）
- [ ] マルチクラウドUI

### エンタープライズ機能
- [ ] RBAC
- [ ] マルチテナント対応
- [ ] 高度なレポート機能
- [ ] 監査ログ

### AI機能
- [ ] 異常検知（Anomaly Detection）
- [ ] 自動修復提案
- [ ] 予測分析

---

## ✅ 完了済み

- [x] Storybook駆動開発（SDD）実装 - 17 stories
- [x] AWS公式アイコン統合 - 28個
- [x] ビジュアル改善（ノードサイズ、VPC/Subnet階層）
- [x] Drift Detection表示機能
- [x] DisplayOptions改善（ドラッグ、フィルター）
- [x] AWS環境クリーンアップ（116リソース削除）
- [x] CloudTrailプラグインフィールド名修正
- [x] バックエンドエラーハンドリング強化

---

**関連ドキュメント**:
- [現状分析レポート](./STATUS_REPORT_2026-01-10.md)
- [プロジェクトロードマップ](./PROJECT_ROADMAP.md)
- [バックエンドTODO詳細](./docs/BACKEND_TODO.md)
