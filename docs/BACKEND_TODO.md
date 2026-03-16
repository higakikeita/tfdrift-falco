# TFDrift-Falco バックエンド TODO 詳細

**最終更新**: 2026-03-16
**現在バージョン**: v0.5.0+
**Phase**: Phase 1 安定化

---

## 🔥 Phase 1: 今すぐやるべき

| # | タスク | 優先度 | 工数見積もり | 担当者 | ステータス |
|---|--------|--------|-------------|--------|-----------|
| 1 | エラーハンドリング強化 | 🔴 最高 | 2h | - | ✅ 完了 |
| 2 | Falcoルール修正・拡充 | 🔴 最高 | 1.5h | - | ✅ 完了 |
| 3 | GET /health 依存サービスチェック追加 | 🔴 高 | 1h | - | 未着手 |
| 4 | ログレベル最適化 | 🟡 中 | 1h | - | 未着手 |
| 5 | GraphDBクエリパフォーマンス改善 | 🔴 高 | 3h | - | 未着手 |
| 6 | GraphDBインデックス追加 | 🔴 高 | 2h | - | 未着手 |

### 1. エラーハンドリング強化 ✅

**対象ファイル**:
- `pkg/api/handlers/discovery.go` — AWS認証エラーの分類（401/403/504）、Terraform State未ロード時の503
- `pkg/api/handlers/graph_query.go` — `models.APIResponse` 統一、`respondError()` 利用に統一
- `pkg/api/handlers/stats.go` — 型アサーション安全化（パニック防止）

**完了した改善**:
- AWS SDK エラーを HTTP ステータスコードに分類する `classifyAWSError()` を追加
- Terraform State が nil/空の場合に適切なエラーレスポンス（503/404）を返却
- `graph_query.go` の全エンドポイントで `models.APIResponse` 形式に統一
- `stats.go` の unsafe な型アサーション `.(map[string]int)` をガード付きに修正

### 2. Falcoルール修正・拡充 ✅

**対象ファイル**: `rules/terraform_drift.yaml`

**完了した改善**:
- `required_engine_version` と `required_plugin_versions` ヘッダーの追加
- `ct.error = ""` 条件の追加（成功したAPIコールのみ検知）
- `ct.srcip`, `ct.accountid` を output に追加（監査追跡性向上）
- 新規ルール追加: ECS, EKS, DynamoDB, VPC/Network, KMS
- Security Group を CRITICAL に昇格（ネットワークセキュリティ直結）
- IAM ルールにポリシーアタッチ系イベントを追加

### 3. GET /health 依存サービスチェック追加

**対象ファイル**: `pkg/api/handlers/health.go`

**やること**:
- GraphDB接続状態の確認
- Terraform State ロード状態の確認
- Falco接続状態の確認（オプショナル）
- レスポンスに `checks` フィールドを追加

**期待するレスポンス例**:
```json
{
  "status": "degraded",
  "version": "v0.5.0",
  "timestamp": "...",
  "checks": {
    "graph_db": "ok",
    "terraform_state": "ok",
    "falco": "unavailable"
  }
}
```

### 4. ログレベル最適化

**対象**: `pkg/api/handlers/` 全体、`pkg/detector/`

**やること**:
- `log.Infof` でのポインタ出力（`%p`）を `log.Debugf` に変更
- リクエストごとの Debug ログを統一
- エラーパスでのログレベルを Error/Warn に適切に使い分け

### 5-6. GraphDB最適化

**対象ファイル**: `pkg/graph/`

**やること**:
- 大規模グラフ（100+ ノード）でのクエリ応答時間を計測
- ノード検索にインデックス（ラベル別、ID 別の map）を追加
- キャッシュ機構の実装（BuildGraph 結果のキャッシュ）

---

## 🔴 Phase 1 継続: 今月中

| # | タスク | 優先度 | 工数見積もり | 担当者 | ステータス |
|---|--------|--------|-------------|--------|-----------|
| 7 | API統合テスト追加 | 🔴 高 | 4h | - | 未着手 |
| 8 | E2Eテスト基盤構築 | 🟡 中 | 2h | - | 未着手 |
| 9 | APIレスポンスタイム改善 | 🟡 中 | 2h | - | 未着手 |
| 10 | Docker Compose改善 | 🟡 中 | 2h | - | 未着手 |
| 11 | 小規模AWS環境構築 | 🔴 高 | 3h | - | 未着手 |

### 7. API統合テスト追加

**対象**: 全 API エンドポイント

**やること**:
- `httptest.NewServer` を使った統合テスト
- 正常系・異常系の網羅
- エラーレスポンスの構造検証
- ペジネーションのテスト

### 8. E2Eテスト基盤構築

**やること**:
- テストフレームワーク選定（Go test + testcontainers）
- CI パイプラインへの組み込み
- モックデータによる基本シナリオ

### 9. APIレスポンスタイム改善

**やること**:
- `BuildGraph()` の結果キャッシュ（TTL付き）
- 大量ノード/エッジ取得時のストリーミング検討
- Prometheus メトリクスでのレイテンシ計測

### 10. Docker Compose改善

**対象ファイル**: `docker-compose.yml`

**やること**:
- `version` フィールド削除（deprecated）
- `AWS_PROFILE` 環境変数追加
- Falco依存関係の条件付き化（`profiles` 利用）
- ヘルスチェック改善

### 11. 小規模AWS環境構築

**対象**: `terraform/minimal-environment/`

**やること**:
- 10-20リソース構成の設計
- VPC, Subnet x2, EKS/ECS, RDS, S3
- 実データでの UI 動作確認

---

## 🟡 Phase 2: 来月（機能拡張）

| # | タスク | 優先度 | 工数見積もり | 担当者 | ステータス |
|---|--------|--------|-------------|--------|-----------|
| 12 | WebSocket完全実装 | 🟡 中 | 6h | - | 未着手 |
| 13 | SSE完全実装 | 🟡 中 | 4h | - | 未着手 |
| 14 | Drift履歴機能 | 🟡 中 | 8h | - | 未着手 |
| 15 | Impact Radius表示 | 🟢 低 | 4h | - | 未着手 |
| 16 | CloudTrail統合ガイド作成 | 🟡 中 | 3h | - | 未着手 |
| 17 | CloudTrailログ処理最適化 | 🟡 中 | 4h | - | 未着手 |

---

## 🟢 Phase 3+: 将来

| # | タスク | 優先度 | 工数見積もり | 担当者 | ステータス |
|---|--------|--------|-------------|--------|-----------|
| 18 | GCP統合テスト完了 | 🟢 低 | 8h | - | 未着手 |
| 19 | Azure対応開始 | 🟢 低 | 40h | - | 未着手 |
| 20 | RBAC実装 | 🟢 低 | 20h | - | 未着手 |
| 21 | マルチテナント対応 | 🟢 低 | 30h | - | 未着手 |
| 22 | 異常検知 (ML) | 🟢 低 | 40h | - | 未着手 |
| 23 | 自動修復提案 | 🟢 低 | 30h | - | 未着手 |
| 24 | バックエンドリファクタリング | 🟢 低 | 10h | - | 未着手 |
| 25 | パッケージ構成最適化 | 🟢 低 | 6h | - | 未着手 |

---

## 変更履歴

| 日付 | 変更内容 |
|------|---------|
| 2026-03-16 | 初版作成。エラーハンドリング強化・Falcoルール修正完了 |
