# TFDrift-Falco Project Roadmap

**作成日**: 2026-01-10
**最終更新**: 2026-07-19
**現在バージョン**: v0.13.0
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

## 現在地 (v0.13.0)

### 完成度マトリクス

| コンポーネント | 完成度 | 評価 | 備考 |
|---|---|---|---|
| **バックエンドAPI** | 90% | 🟢 | REST API完成、JWT/API Key認証、レートリミット、Provider API追加 |
| **フロントエンドUI** | 85% | 🟢 | Dashboard、Graph、Table、Split View。Provider Status UI未実装 |
| **Falco統合 (AWS)** | 85% | 🟢 | 500+イベント、40+サービス、衝突解決済み |
| **Falco統合 (GCP)** | 80% | 🟢 | 170+イベント、27+サービス、ResourceDiscoverer実装済み |
| **Falco統合 (Azure)** | 40% | 🟡 | Activity Log イベント検知のみ。**discovery/comparison は未実装（Azure SDK 未統合）**＝#326 |
| **Terraform Backend** | 90% | 🟢 | local, S3, GCS, Azure Blob Storage 4バックエンド対応 |
| **Drift検知エンジン** | 85% | 🟢 | 3クラウド対応、StateComparator統一インターフェース |
| **リアルタイム通信** | 85% | 🟢 | WebSocket（Provider filtering, 4新イベントタイプ）+ SSE |
| **Webhook統合** | 85% | 🟢 | Slack, Teams, 汎用対応 |
| **ドキュメント** | 85% | 🟢 | OpenAPI、Runbook、VERSIONING.md（24項目チェックリスト） |
| **テスト** | 80% | 🟢 | Go unit tests充実、E2E 6シナリオ（`-tags e2e`）、lint警告ゼロ、カバレッジゲート70% |
| **CI/CD** | 85% | 🟢 | 統合パイプライン完成 |
| **セキュリティ** | 85% | 🟢 | JWT/API Key認証、レートリミット、セキュリティスキャン |
| **Kubernetes** | 80% | 🟢 | Helm Chart、HPA、NetworkPolicy |

**総合完成度**: **84%**

### v0.13.0 で追加されたもの（v0.9.0 以降の一括清算リリース）
- ✅ クロスクラウドドリフト相関エンジン / Provider Status API + Grafana
- ✅ Policy-as-Code（OPA/Rego）(#136)・Drift自動修復 + GitHub PR作成 (#135)・OpenTelemetry統合 (#137)
- ✅ AWS SDK v2 移行、Sprint 3〜6 の品質改善（lint警告ゼロ、UIテスト拡充）
- ✅ CI/E2E/GHCR パイプライン修復、E2Eテスト再有効化（`-tags e2e`）(#278)
- ✅ Falcoルールをプラグイン実フィールドで検証・修正 (#280)
- ℹ️ v0.10.0〜v0.12.0 タグは Release/GHCR 未公開のまま残っていたため、本リリースに統合（ADR「前進で清算」）

### 課題
- ⚠️ E2E定常実行は停止中（`E2E_AWS_*` シークレット未設定。設定後に schedule 復帰）
- ⚠️ Provider Status UI（フロントエンド）未実装 — バックエンドAPIのみ
- ⚠️ AsyncAPI仕様（WebSocket/SSE）未作成
- ⚠️ Azure Falco Plugin実環境統合テスト未実施（`docs/azure-setup.md` も未作成）
- ⚠️ ESLint 10系 dependabot PR は typescript-eslint 対応待ちで保留

---

## ロードマップ

### 完了済み Phase

| Phase | バージョン | 状態 | ハイライト |
|---|---|---|---|
| **Phase 1** | v0.2.0-beta → v0.5.0 | ✅ 完了 | MVP、AWS CloudTrail、GCP対応、Storybook UI |
| **Phase 2** | v0.6.0 (2026-03-20) | ✅ 完了 | マルチクラウド拡張（AWS 500+, GCP 170+イベント） |
| **Phase 3** | v0.7.0 (2026-03-22) | ✅ 完了 | Dashboard UI、Azure基本対応、リアルタイム通知 |
| **Phase 4** | v0.8.0 (2026-03-22) | ✅ 完了 | JWT/API Key認証、OpenAPI、Helm Chart、Runbook |
| **Phase 5** | v0.9.0 (2026-03-29) | ✅ 完了 | Azure イベント検知（discovery/comparison は未実装＝#326）、azurerm backend、WebSocket v0.6.0強化 |
| **Phase 6** | v0.13.0 (2026-07-19) | ✅ 完了 | E2E再有効化、lint警告ゼロ、カバレッジゲート、CI/GHCR修復 |
| **Phase 7** | v0.13.0 (2026-07-19) | 🟡 一部完了 | クロスクラウド相関 + Provider Status API。UI と Azure実環境検証は残 |
| **Phase 8** | v0.13.0 (2026-07-19) | ✅ 完了 | Policy-as-Code (OPA)、Drift自動修復、OpenTelemetry |

---

### Phase 6: テスト基盤 & 品質強化 — ✅ 完了（v0.13.0 で正式リリース）

**目標**: リグレッション防止のためのテスト基盤確立

- [x] **E2Eテスト環境構築** — `-tags e2e` の6シナリオ + e2e.yml ワークフロー (#278)。定常実行は `E2E_AWS_*` シークレット設定後に schedule 復帰
- [x] **既存テスト修正 & カバレッジ向上** — pre-existing failures 解消、golangci-lint 警告ゼロ、カバレッジゲート70% + codecov 連携 (#278)
- [x] **APIテスト** — Sprint 6 で統合テスト追加（tests/integration/）

---

### Phase 7: Azure Falco統合 & クロスクラウド — 🟡 一部完了

**目標**: Azure実環境統合とクラウド横断分析

- [x] **クロスクラウドドリフト相関** — 相関エンジン実装（`pkg/detector/correlator.go`）
- [~] **Grafanaダッシュボードテンプレート**（Azure対応版含む） — 後に撤去。可視化は専用 React Web UI に一本化（理由: `ui/docs/qiita-article.md`）
- [ ] **Azure Activity Log → Falco Plugin統合テスト**（残）
  - Azure Event Hub → Falco azureaudit plugin のE2E検証
  - Azure環境セットアップガイド（`docs/azure-setup.md`）
- [ ] **Provider Status UI**（残 — バックエンド `GET /api/v1/providers` は実装済み、フロントエンド未着手）

---

### Phase 8: Drift自動修復 & Policy-as-Code — ✅ 完了（v0.13.0 で正式リリース）

**目標**: 検知から修復までの自動化

- [x] **Drift自動修復提案** — `terraform import` 提案 + GitHub PR自動作成 (#135)
- [x] **Policy-as-Code** — OPA/Rego によるポリシー分類（許容/アラート/自動修復）(#136)
- [x] **OpenTelemetry統合** — 分散トレーシング + OTLPメトリクス (#137)。Grafana Tempo との連携検証は残

---

### Phase 9: v1.0.0 GA (Production Ready)

**目標**: 本番環境対応の完成

**GA Criteria** (VERSIONING.md より):
1. [ ] 3クラウドプロバイダー対応 — AWS/GCP は検知＋discovery/comparison（core types）。**Azure は検知のみ、discovery/comparison 未実装（#326）**
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

## アーキテクチャ (v0.13.0)

```
┌──────────────────────────────────────────────────────────────────────┐
│                       TFDrift-Falco v0.13.0                          │
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
| v0.10.0〜v0.12.0 | 2026-03-29 | ⚠️ タグのみ（Release/GHCR 未公開の幽霊リリース）。内容は v0.13.0 に統合 |
| v0.13.0 | 2026-07-19 | 一括清算: 相関エンジン、Policy-as-Code、自動修復、OTel、CI/E2E/GHCR修復 |

詳細は [CHANGELOG.md](CHANGELOG.md) および [VERSIONING.md](VERSIONING.md) を参照。

---

**作成者**: Keita Higaki
**最終更新**: 2026-07-19
**次回レビュー**: v0.13.0 ベンチマーク到達時（方向性再議論とあわせて）
