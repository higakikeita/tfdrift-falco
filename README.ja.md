# TFDrift-Falco

**Falcoを活用したリアルタイムTerraformドリフト検知**

[![Version](https://img.shields.io/badge/version-0.9.0-blue)](https://github.com/higakikeita/tfdrift-falco/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)
[![Falco](https://img.shields.io/badge/Falco-Compatible-blue)](https://falco.org/)
[![Docker](https://img.shields.io/badge/Docker-GHCR-2496ED?logo=docker)](https://ghcr.io/higakikeita/tfdrift-falco)

> **v0.9.0** (2026-03-29) — Azure完全対応、azurermバックエンド、WebSocket強化
> [リリースノート](docs/release-notes/v0.9.0.md) | [CHANGELOG](CHANGELOG.md) | [ロードマップ](PROJECT_ROADMAP.md)

[English](README.md) | **[日本語]**

---

## TFDrift-Falcoとは？

TFDrift-Falcoは、AWS・GCP・Azureにまたがるインフラのドリフトを**リアルタイム**で検知するツールです。Falcoのランタイムセキュリティ監視とTerraformステート比較を組み合わせ、定期スキャンでは得られない「**誰が**、**いつ**、**何を**変更したか」を変更の瞬間に捕捉します。

```
AWSコンソールでセキュリティグループを変更
  → FalcoがCloudTrailイベントをリアルタイムで検知
    → TFDrift-FalcoがTerraformステートと比較
      → ユーザー情報と変更詳細を含む即時アラート
```

---

## クイックスタート

```bash
# クローンして設定
git clone https://github.com/higakikeita/tfdrift-falco.git && cd tfdrift-falco
cp config.yaml.example config.yaml  # 設定を編集

# 起動
docker compose up -d

# デモモード（クラウド認証不要）
go run ./cmd/tfdrift --demo
```

詳細は [Getting Started Guide](docs/GETTING_STARTED.md) を参照。

---

## マルチクラウド対応

統一されたProviderインターフェースで3大クラウドをサポートしています。

### AWS

```yaml
providers:
  aws:
    enabled: true
    regions: [us-east-1]
    state:
      backend: s3
      s3_bucket: "your-terraform-state-bucket"
      s3_key: "terraform.tfstate"
```

500以上のCloudTrailイベント、40以上のサービス対応。[詳細](docs/services/index.md)

### GCP

```yaml
providers:
  gcp:
    enabled: true
    project_id: "your-gcp-project"
    state:
      backend: gcs
      gcs_bucket: "your-terraform-state-bucket"
```

170以上のAudit Logイベント、27以上のサービス対応。[セットアップガイド](docs/gcp-setup.md)

### Azure

```yaml
providers:
  azure:
    enabled: true
    subscription_id: "your-subscription-id"
    regions: [eastus, westus2]
    state:
      backend: azurerm
      azure_storage_account: "yourstorageaccount"
      azure_container_name: "tfstate"
      azure_blob_name: "terraform.tfstate"
```

119オペレーション、20以上のサービス対応。ResourceDiscovererとStateComparator完全実装。[詳細](docs/release-notes/v0.9.0.md)

### プロバイダー機能比較

| 機能 | AWS | GCP | Azure |
|---|---|---|---|
| リアルタイムイベント検知 | CloudTrail (500+) | Audit Logs (170+) | Activity Logs (119) |
| リソース探索 | 対応 | 対応 | 対応 |
| ステート比較 | 対応 | 対応 | 対応 |
| Terraformバックエンド | S3 | GCS | Azure Blob |
| Falcoプラグイン | aws_cloudtrail | gcpaudit | azureaudit |

---

## 主な機能

**リアルタイム検知** — Falco gRPCによるイベント駆動。定期スキャンではない。

**3方向ドリフト分析** — 未管理リソース（クラウドにあるがTerraformにない）、欠落リソース（Terraformにあるがクラウドから削除済み）、変更リソース（属性差分）を検知。

**セキュリティコンテキスト** — 全ての変更にIAMユーザー、APIキー、サービスアカウントを紐付け。

**REST API + WebSocket + SSE** — フルAPIサーバー。リアルタイムストリーミング、プロバイダーフィルタリング、構造化JSONイベント。

**Webhook通知** — Slack、Microsoft Teams、PagerDuty、カスタムHTTPエンドポイント対応。自動リトライ付き。

**本番対応** — JWT/APIキー認証、レートリミット、OpenAPI 3.0仕様、Kubernetes Helm Chart（HPA、NetworkPolicy対応）。

---

## アーキテクチャ

```
 AWS CloudTrail ─┐
 GCP Audit Logs ─┤──→ Falco (gRPC) ──→ Provider Layer ──→ Drift Engine
 Azure Activity ─┘    (plugins)        (parse/map/discover)  (compare)
                                                                  │
                              ┌────────────────┬──────────────────┤
                              ▼                ▼                  ▼
                         GraphDB          Webhook            API Server
                       (in-memory)     (Slack/Teams)     (REST + WS + SSE)
                                                               │
                                                          React UI
                                                     (Graph/Table/Split)
```

詳細は [アーキテクチャドキュメント](docs/architecture.md) を参照。

---

## API

```bash
# RESTエンドポイント
GET  /api/v1/drifts        # ドリフトアラート（フィルタリング対応）
GET  /api/v1/events        # Falcoイベント
GET  /api/v1/graph         # 因果グラフ（Cytoscape形式）
GET  /api/v1/state         # Terraformステート概要
GET  /api/v1/stats         # 統計
GET  /api/v1/providers     # プロバイダー機能
GET  /health               # ヘルスチェック

# リアルタイム
GET  /api/v1/stream        # SSEイベントストリーム
WS   /ws                   # WebSocket（プロバイダーフィルタリング対応）
```

完全な仕様: [OpenAPI 3.0](docs/api/openapi.yaml)

---

## デプロイ

### Docker Compose

```bash
docker compose up -d
# フロントエンド: http://localhost:3000
# バックエンド:  http://localhost:8080/api/v1
# WebSocket: ws://localhost:8080/ws
```

### Kubernetes (Helm)

```bash
helm install tfdrift ./charts/tfdrift-falco
```

### ソースからビルド

```bash
make build    # バイナリ: ./bin/tfdrift
make test     # テスト実行
make lint     # Lint実行
```

---

## ドキュメント

| ドキュメント | 説明 |
|---|---|
| [Getting Started](docs/GETTING_STARTED.md) | ステップバイステップのセットアップガイド |
| [アーキテクチャ](docs/architecture.md) | システム設計とデータフロー |
| [GCPセットアップ](docs/gcp-setup.md) | GCP固有の設定 |
| [APIリファレンス](docs/api/openapi.yaml) | OpenAPI 3.0仕様 |
| [運用Runbook](docs/operations/runbook.md) | SRE向けインシデントPlaybook |
| [デプロイ](docs/deployment.md) | 本番デプロイオプション |
| [Contributing](CONTRIBUTING.md) | 開発ワークフロー |
| [バージョニング](VERSIONING.md) | リリースポリシーとチェックリスト（24項目） |
| [ロードマップ](PROJECT_ROADMAP.md) | v0.10.0 → v1.0.0 計画 |
| [Changelog](CHANGELOG.md) | バージョン履歴 |

---

## なぜFalcoか？

> Terraformは「**何があるべきか**」を知っている。Falcoは「**何が起きたか**」を知っている。

従来のドリフト検知は定期スキャン — 変更を見つけた時には「誰が」「なぜ」は失われています。Falcoはクラウド監査ログをプラグインフレームワークでリアルタイム監視し、変更の瞬間に行為者・行為・意図を捕捉します。

TFDrift-Falcoは設計図（Terraform）と証人（Falco）の2つの世界を橋渡しします。

全文はこちら: [Why Falco?](docs/WHY_FALCO_STORY.md)

---

## コントリビューション

コントリビューションを歓迎します。開発環境のセットアップ、コーディング規約、PRガイドラインは [CONTRIBUTING.md](CONTRIBUTING.md) を参照してください。

## ライセンス

[MIT](LICENSE)
