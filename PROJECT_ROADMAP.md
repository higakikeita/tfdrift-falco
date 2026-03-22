# TFDrift-Falco Project Roadmap

**作成日**: 2026-01-10
**最終更新**: 2026-03-22
**現在バージョン**: v0.8.0
**ビジョン**: Real-time, Multi-Cloud Terraform Drift Detection with Security Context

---

## 🎯 プロジェクトビジョン

### ミッション
Terraform管理下のインフラストラクチャと実際のクラウド環境の差分（Drift）を**リアルタイム**で検知し、セキュリティコンテキストと共に可視化することで、インフラの健全性と安全性を保証する。

### コアバリュー
1. **リアルタイム性** - 定期スキャンではなく、イベント駆動による即座の検知
2. **セキュリティファースト** - Falco統合による脅威分析
3. **マルチクラウド** - AWS, GCP, Azure 3クラウド対応
4. **開発者体験** - 直感的なUI、明確なAPI、豊富なドキュメント
5. **オープンソース** - コミュニティ駆動の発展

---

## 📊 現在地 (v0.8.0)

### 完成度マトリクス

| コンポーネント | 完成度 | 評価 | 備考 |
|---|---|---|---|
| **バックエンドAPI** | 90% | 🟢 優秀 | REST API完成、JWT/API Key認証、レートリミット |
| **フロントエンドUI** | 90% | 🟢 優秀 | Dashboard、Events、Topology、Settings 4ページ完成 |
| **Falco統合 (AWS)** | 85% | 🟢 優秀 | 500+イベント、40+サービス、衝突解決済み |
| **Falco統合 (GCP)** | 80% | 🟢 優秀 | 170+イベント、27+サービス |
| **Falco統合 (Azure)** | 70% | 🟡 良好 | 119オペレーション、20+サービス、基本統合完了 |
| **GraphDB** | 65% | 🟡 良好 | 基本実装＋エクスポート機能、最適化余地あり |
| **Drift検知エンジン** | 80% | 🟢 優秀 | コア機能完成、テーブル駆動テスト充実 |
| **リアルタイム通信** | 80% | 🟢 優秀 | SSE通知パネル、WebSocket基盤実装済み |
| **Webhook統合** | 85% | 🟢 優秀 | Slack, Teams, 汎用対応＋設定UI |
| **ドキュメント** | 90% | 🟢 優秀 | OpenAPI仕様書、運用Runbook、README刷新 |
| **テスト** | 85% | 🟢 優秀 | Go 80%+、フロントエンド26テストファイル |
| **CI/CD** | 85% | 🟢 優秀 | 統合パイプライン完成 |
| **セキュリティ** | 85% | 🟢 優秀 | JWT/API Key認証、レートリミット、セキュリティスキャン |
| **Kubernetes** | 80% | 🟢 優秀 | Helm Chart、HPA、NetworkPolicy、ServiceMonitor |

**総合完成度**: **85%**

### 強み
- ✅ 3クラウド対応（AWS 500+, GCP 170+, Azure 119オペレーション）
- ✅ Dashboard UI完全実装（4ページ、リアルタイム通知、ダーク/ライトテーマ）
- ✅ イベント管理ワークフロー（確認/無視/解決/再開）
- ✅ グラフエクスポート（PNG/SVG/JSON）
- ✅ 設定管理UI（Webhooks, Rules, Providers, General）
- ✅ バージョニングポリシー確立（VERSIONING.md）
- ✅ CI/CD自動化、セキュリティスキャン統合
- ✅ JWT/API Key認証 + レートリミット
- ✅ OpenAPI 3.0仕様書 + Swagger UI
- ✅ Kubernetes Helm Chart（HPA, NetworkPolicy, ServiceMonitor）
- ✅ Operations Runbook（5インシデントPlaybook）
- ✅ フロントエンドテスト26ファイル

### 課題
- ⚠️ E2Eテスト未実装
- ⚠️ OpenTelemetry未統合
- ⚠️ AsyncAPI仕様（WebSocket/SSE）未作成

---

## 🗺️ ロードマップ

### Phase 1-4: ✅ 完了済み

| Phase | バージョン | 状態 | ハイライト |
|---|---|---|---|
| **Phase 1** | v0.2.0-beta → v0.5.0 | ✅ 完了 | MVP、AWS CloudTrail、GCP対応、Storybook UI |
| **Phase 2** | v0.6.0 (2026-03-20) | ✅ 完了 | マルチクラウド拡張（AWS 500+, GCP 170+イベント） |
| **Phase 3** | v0.7.0 (2026-03-22) | ✅ 完了 | Dashboard UI、Azure対応、リアルタイム通知、設定管理 |
| **Phase 4** | v0.8.0 (2026-03-22) | ✅ 完了 | JWT/API Key認証、レートリミット、OpenAPI、Helm Chart、Runbook |

---

### Phase 5: AIによる自動化 (v0.9.0) - 2ヶ月

**目標**: AI/MLによるインテリジェント機能

#### v0.9.0 - AI機能
- [ ] **異常検知 (Anomaly Detection)**
  - MLモデルによる異常パターン検出
  - 自動アラート生成
  - False Positive削減

- [ ] **自動修復提案**
  - Driftの自動修復コード生成
  - Terraform Plan生成
  - Pull Request自動作成

- [ ] **予測分析**
  - Driftリスク予測
  - リソース変更の影響予測
  - コスト最適化提案

---

### Phase 6: v1.0.0 リリース (Production Ready) - 1ヶ月

**目標**: 本番環境対応の完成

#### v1.0.0 - Production Ready
- [ ] **完全なドキュメント**
  - 全機能のドキュメント完備
  - チュートリアル動画
  - API完全ドキュメント

- [ ] **パフォーマンス最適化**
  - 10,000+ ノード対応
  - クラスタリング・LOD実装
  - Web Worker活用

- [ ] **セキュリティ監査**
  - 第三者セキュリティ監査
  - 脆弱性スキャン
  - ペネトレーションテスト

- [ ] **本番環境デプロイガイド**
  - Kubernetes Helm Chart
  - AWS ECS/Fargate対応
  - GKE対応
  - HA (High Availability) 構成

---

## 📊 成功指標 (KPI)

### v0.8.0 完了時 ✅
- [x] API認証実装済み（JWT + API Key）
- [x] OpenAPI仕様書公開（Swagger UI統合）
- [x] Helm Chart利用可能（HPA, NetworkPolicy, ServiceMonitor）
- [x] フロントエンドテスト26ファイル
- [x] Operations Runbook完備（5インシデントPlaybook）

---

## 🏗️ アーキテクチャ (v0.8.0)

```
┌──────────────────────────────────────────────────────────────────────┐
│                       TFDrift-Falco v0.8.0                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ AWS           │  │ GCP          │  │ Azure                    │  │
│  │ CloudTrail    │  │ Audit Logs   │  │ Activity Logs            │  │
│  │ (500+ events) │  │ (170+ events)│  │ (119 operations)         │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────────┘  │
│         │                  │                      │                   │
│         └─────────┬────────┴──────────────────────┘                   │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  Falco gRPC       │  ← Falco Plugin Framework              │
│          │  Subscriber       │    (aws_cloudtrail, gcpaudit,          │
│          └────────┬──────────┘     azureaudit)                        │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  Event Parser     │  ← Provider-specific parsers           │
│          │  + Resource Mapper│    (AWS 40+, GCP 27+, Azure 20+       │
│          └────────┬──────────┘     Terraform resource types)          │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  Drift Detection  │  ← Rule engine + severity scoring      │
│          │  Engine           │    + unmanaged resource detection       │
│          └────────┬──────────┘                                        │
│                   │                                                    │
│          ┌────────▼──────────┐  ┌────────────────────┐               │
│          │  In-Memory Graph  │  │  Webhook Notifier  │               │
│          │  Store (GraphDB)  │  │  (Slack/Teams/HTTP) │               │
│          └────────┬──────────┘  └────────────────────┘               │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  API Server       │  ← Chi Router                          │
│          │  REST + SSE + WS  │    /api/v1/* (events, graph, config)   │
│          └────────┬──────────┘    /api/v1/stream (SSE)                │
│                   │               /ws (WebSocket)                      │
│          ┌────────▼──────────┐                                        │
│          │  React 19 UI      │  ← Vite + TypeScript + Tailwind       │
│          │  Dashboard        │    React Query + Zustand               │
│          │  (4 pages)        │    Dashboard / Events / Topology /     │
│          └───────────────────┘    Settings                            │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## 📝 バージョン履歴

| バージョン | 日付 | ハイライト |
|---|---|---|
| v0.2.0-beta | 2025-12 | MVP: AWS CloudTrail + Falco基本統合 |
| v0.3.0 | 2026-01 | Storybook UI、GraphDB、CI/CD |
| v0.4.0 | 2026-01 | Drift Detection Engine、Webhook統合 |
| v0.5.0 | 2026-01 | GCPサポート、API Server、React UI |
| v0.6.0 | 2026-03-20 | マルチクラウド拡張（AWS 500+, GCP 170+イベント） |
| v0.7.0 | 2026-03-22 | Dashboard UI完成、Azure対応、リアルタイム通知 |
| v0.8.0 | 2026-03-22 | エンタープライズ基盤：認証、OpenAPI、Helm Chart、Runbook |

詳細は [CHANGELOG.md](CHANGELOG.md) および [VERSIONING.md](VERSIONING.md) を参照。

---

**作成者**: Keita Higaki
**最終更新**: 2026-03-22
**次回レビュー**: v0.8.0 リリース時
