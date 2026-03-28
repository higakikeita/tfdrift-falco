# TFDrift-Falco Project Roadmap

**作成日**: 2026-01-10
**最終更新**: 2026-03-29
**現在バージョン**: v0.9.0
**ビジョン**: Real-time, Multi-Cloud Terraform Drift Detection with Security Context

---

## プロジェクトビジョン

### ミッション
Terraform管理下のインフラストラクチャと実際のクラウド環境の差分（Drift）を**リアルタイム**で検知し、セキュリティコンテキストと共に可視化することで、インフラの健全性と安全性を保証する。

### コアバリュー
1. **リアルタイム性** - 定期スキャンではなく、イベント駆動による即座の検知
2. **セキュリティファースト** - Falco統合による脅威分析
3. **マルチクラウド** - AWS, GCP, Azure 3クラウド対応
4. **開発者体験** - 直感的なUI、明確なAPI、豊富なドキュメント
5. **オープンソース** - コミュニティ駆動の発展

---

## 現在地 (v0.9.0)

### 完成度マトリクス

| コンポーネント | 完成度 | 評価 | 備考 |
|---|---|---|---|
| **バックエンドAPI** | 90% | 🟢 | REST API完成、JWT/API Key認証、レートリミット、Provider API追加 |
| **フロントエンドUI** | 85% | 🟢 | Dashboard、Graph、Table、Split View。Provider Status UI未実装 |
| **Falco統合 (AWS)** | 85% | 🟢 | 500+イベント、40+サービス、衝突解決済み |
| **Falco統合 (GCP)** | 80% | 🟢 | 170+イベント、27+サービス、ResourceDiscoverer実装済み |
| **Falco統合 (Azure)** | 80% | 🟢 | 119オペレーション、20+サービス、FullProvider実装済み（Discovery+Comparison） |
| **Terraform Backend** | 90% | 🟢 | local, S3, GCS, Azure Blob Storage 4バックエンド対応 |
| **Drift検知エンジン** | 85% | 🟢 | 3クラウド対応、StateComparator統一インターフェース |
| **リアルタイム通信** | 85% | 🟢 | WebSocket（Provider filtering, 4新イベントタイプ）+ SSE |
| **Webhook統合** | 85% | 🟢 | Slack, Teams, 汎用対応 |
| **ドキュメント** | 85% | 🟢 | OpenAPI、Runbook、VERSIONING.md（24項目チェックリスト） |
| **テスト** | 75% | 🟡 | Go unit tests充実。E2Eテスト未実装、pre-existing failures残存 |
| **CI/CD** | 85% | 🟢 | 統合パイプライン完成 |
| **セキュリティ** | 85% | 🟢 | JWT/API Key認証、レートリミット、セキュリティスキャン |
| **Kubernetes** | 80% | 🟢 | Helm Chart、HPA、NetworkPolicy、ServiceMonitor |

**総合完成度**: **84%**

### v0.9.0 で追加されたもの
- ✅ Azure FullProvider（ResourceDiscoverer + StateComparator）
- ✅ Azure Blob Storage Terraform backend（azurerm）
- ✅ WebSocket Provider filtering + 4新イベントタイプ
- ✅ Provider Capabilities API（`GET /api/v1/providers`）
- ✅ VERSIONING.md（24項目リリースチェックリスト、Anti-Patterns）

### 課題
- ⚠️ E2Eテスト未実装
- ⚠️ pre-existing test failures（GCP dataproc, terraform approval）
- ⚠️ OpenTelemetry未統合
- ⚠️ AsyncAPI仕様（WebSocket/SSE）未作成
- ⚠️ Azure Falco Plugin実環境統合テスト未実施

---

## ロードマップ

### 完了済み Phase

| Phase | バージョン | 状態 | ハイライト |
|---|---|---|---|
| **Phase 1** | v0.2.0-beta → v0.5.0 | ✅ 完了 | MVP、AWS CloudTrail、GCP対応、Storybook UI |
| **Phase 2** | v0.6.0 (2026-03-20) | ✅ 完了 | マルチクラウド拡張（AWS 500+, GCP 170+イベント） |
| **Phase 3** | v0.7.0 (2026-03-22) | ✅ 完了 | Dashboard UI、Azure基本対応、リアルタイム通知 |
| **Phase 4** | v0.8.0 (2026-03-22) | ✅ 完了 | JWT/API Key認証、OpenAPI、Helm Chart、Runbook |
| **Phase 5** | v0.9.0 (2026-03-29) | ✅ 完了 | Azure FullProvider、azurerm backend、WebSocket v0.6.0強化 |

---

### Phase 6: テスト基盤 & 品質強化 (v0.10.0)

**目標**: リグレッション防止のためのテスト基盤確立

- [ ] **E2Eテスト環境構築**
  - Docker Compose統合テスト（Falcoモック + クラウドモック）
  - CI/CDパイプラインにE2Eステージ追加
  - GitHub Actionsワークフロー統合

- [ ] **既存テスト修正 & カバレッジ向上**
  - pre-existing test failures修正（GCP dataproc nil pointer、terraform approval）
  - golangci-lint警告ゼロ化
  - codecov連携修正・カバレッジレポート自動化

- [ ] **APIテスト**
  - REST APIエンドポイント統合テスト
  - WebSocket/SSEイベント配信テスト
  - Provider Capabilities APIテスト

---

### Phase 7: Azure Falco統合 & クロスクラウド (v0.11.0)

**目標**: Azure実環境統合とクラウド横断分析

- [ ] **Azure Activity Log → Falco Plugin統合テスト**
  - Azure Event Hub → Falco azureaudit plugin のE2E検証
  - Azure Activity Logイベントパーシング検証
  - Azure環境セットアップガイド（`docs/azure-setup.md`）

- [ ] **クロスクラウドドリフト相関**
  - 複数プロバイダーのドリフト結果を統合ビューで表示
  - クラウド横断リソースマッピング（例：同一サービスのAWS/Azure版）
  - 統合ダッシュボードUI

- [ ] **Provider Status UI**
  - `GET /api/v1/providers` のフロントエンド実装
  - プロバイダー別ヘルスチェック・ケイパビリティ表示
  - Grafanaダッシュボードテンプレート（Azure対応版）

---

### Phase 8: Drift自動修復 & Policy-as-Code (v0.12.0)

**目標**: 検知から修復までの自動化

- [ ] **Drift自動修復提案**
  - 検知したドリフトに対する `terraform plan` 自動生成
  - unmanaged resource → `terraform import` コマンド提案
  - GitHub PR自動作成（修復用Terraformコード）

- [ ] **Policy-as-Code**
  - OPA/Regoによるドリフトポリシー定義
  - 「許容するドリフト」「即時アラート」「自動修復」のポリシー分類
  - ポリシーバイオレーション通知

- [ ] **OpenTelemetry統合**
  - 分散トレーシング（イベント受信→パース→検知→通知の全フロー）
  - メトリクス公開（Prometheus形式 + OTLP）
  - Grafana Tempoとの連携

---

### Phase 9: v1.0.0 GA (Production Ready)

**目標**: 本番環境対応の完成

**GA Criteria** (VERSIONING.md より):
1. ✅ 3クラウドプロバイダー完全対応（AWS, GCP, Azure）
2. [ ] 安定API — /api/v1 エンドポイントの破壊的変更なし
3. [ ] 本番環境検証 — 最低1環境での本番デプロイ実績
4. [ ] ドキュメント完備 — 全プロバイダーガイド、APIリファレンス
5. [ ] テストカバレッジ — コアパッケージ80%以上
6. [ ] セキュリティ強化 — 認証、入力検証、シークレット管理のレビュー完了

- [ ] **パフォーマンス最適化**
  - 10,000+ ノード対応（GraphDB最適化、Web Worker活用）
  - AsyncAPI仕様書（WebSocket/SSE）

- [ ] **本番デプロイガイド**
  - AWS ECS/Fargate対応
  - GKE対応
  - AKS対応
  - HA (High Availability) 構成ガイド

- [ ] **セキュリティ監査**
  - 脆弱性スキャン最終確認
  - シークレット管理レビュー

---

## アーキテクチャ (v0.9.0)

```
┌──────────────────────────────────────────────────────────────────────┐
│                       TFDrift-Falco v0.9.0                           │
├──────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────────────────┐  │
│  │ AWS           │  │ GCP          │  │ Azure                    │  │
│  │ CloudTrail    │  │ Audit Logs   │  │ Activity Logs            │  │
│  │ (500+ events) │  │ (170+ events)│  │ (119 operations)         │  │
│  │ S3 backend    │  │ GCS backend  │  │ Blob Storage backend     │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬───────────────┘  │
│         │                  │                      │                   │
│         └─────────┬────────┴──────────────────────┘                   │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  Falco gRPC       │  ← Falco Plugin Framework              │
│          │  Subscriber       │    (aws_cloudtrail, gcpaudit,          │
│          └────────┬──────────┘     azureaudit)                        │
│                   │                                                    │
│          ┌────────▼──────────┐  ┌─────────────────────────┐          │
│          │  Provider Layer   │  │  Resource Discovery     │          │
│          │  (unified iface)  │  │  AWS / GCP / Azure      │          │
│          │  Event Parser     │  │  + StateComparator      │          │
│          │  Resource Mapper  │  │  (unmanaged/missing/    │          │
│          └────────┬──────────┘  │   modified detection)   │          │
│                   │              └────────────┬────────────┘          │
│          ┌────────▼──────────────────────────▼──┐                    │
│          │  Drift Detection Engine               │                    │
│          │  Rule engine + severity scoring        │                    │
│          └────────┬──────────────────────────────┘                    │
│                   │                                                    │
│          ┌────────▼──────────┐  ┌────────────────────┐               │
│          │  In-Memory Graph  │  │  Webhook Notifier  │               │
│          │  Store (GraphDB)  │  │  (Slack/Teams/HTTP) │               │
│          └────────┬──────────┘  └────────────────────┘               │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  API Server       │  ← Chi Router                          │
│          │  REST + SSE + WS  │    /api/v1/* + /api/v1/providers       │
│          │  (provider filter)│    /api/v1/stream (SSE)                │
│          └────────┬──────────┘    /ws (WebSocket + provider filter)   │
│                   │                                                    │
│          ┌────────▼──────────┐                                        │
│          │  React 19 UI      │  ← Vite + TypeScript + Tailwind       │
│          │  Graph / Table /  │    React Query + Zustand               │
│          │  Split View       │                                        │
│          └───────────────────┘                                        │
│                                                                      │
└──────────────────────────────────────────────────────────────────────┘
```

---

## バージョン履歴

| バージョン | 日付 | ハイライト |
|---|---|---|
| v0.2.0-beta | 2025-12 | MVP: AWS CloudTrail + Falco基本統合 |
| v0.3.0 | 2026-01 | Storybook UI、GraphDB、CI/CD |
| v0.4.0 | 2026-01 | Drift Detection Engine、Webhook統合 |
| v0.5.0 | 2026-01 | GCPサポート、API Server、React UI |
| v0.6.0 | 2026-03-20 | マルチクラウド拡張（AWS 500+, GCP 170+イベント） |
| v0.7.0 | 2026-03-22 | Dashboard UI完成、Azure基本対応、リアルタイム通知 |
| v0.8.0 | 2026-03-22 | エンタープライズ基盤：認証、OpenAPI、Helm Chart、Runbook |
| v0.9.0 | 2026-03-29 | Azure FullProvider、azurerm backend、WebSocket強化 |

詳細は [CHANGELOG.md](CHANGELOG.md) および [VERSIONING.md](VERSIONING.md) を参照。

---

**作成者**: Keita Higaki
**最終更新**: 2026-03-29
**次回レビュー**: v0.10.0 リリース時
