# TFDrift-Falco Documentation Structure Analysis (MECE)

**Date:** 2025-01-15
**Status:** MECE Restructuring Proposal

---

## Current State Analysis

### Inventory (75 files)

**Published (in mkdocs.yml):** 29 files
- Home: 1
- Getting Started: 7
- AWS Services: 13
- GCP Services: 12
- Release Notes: 4
- Community: 3

**Unpublished:** 46 files (重複・散在・未分類)

---

## MECE Issues Identified

### 1. **Mutually Exclusive (相互排他性) 違反**

**重複コンテンツ:**
- `quickstart.md` vs `GETTING_STARTED.md` (両方とも初期設定)
- `deployment.md` vs `PRODUCTION_READINESS.md` (デプロイ情報)
- `architecture.md` vs `GRAPH_UI_ARCHITECTURE.md` (アーキテクチャ)
- 複数のqiita/zenn記事 (外部記事の複製)

**命名不統一:**
- 大文字: `API.md`, `BEST_PRACTICES.md`, `USAGE.md`
- 小文字: `quickstart.md`, `deployment.md`, `overview.md`
- ハイフン: `how-it-works.md`, `falco-setup.md`
- アンダースコア: `GETTING_STARTED.md`, `USE_CASES.md`

### 2. **Collectively Exhaustive (網羅性) 違反**

**欠落カテゴリ:**
- ❌ **API Reference** - 新規作成したAPI.mdが未公開
- ❌ **Architecture Documentation** - IMPLEMENTATION_SUMMARY.mdが未公開
- ❌ **Best Practices** - BEST_PRACTICES.mdが未公開
- ❌ **Use Cases** - USE_CASES.mdが未公開
- ❌ **Configuration Reference** - 設定リファレンスが散在
- ❌ **Troubleshooting** - トラブルシューティングが不足
- ❌ **Development Guide** - EXTENDING.mdが未公開

---

## Proposed MECE Structure

### Level 1: Top Categories (Mutually Exclusive)

```
1. Home (ホーム)
2. Getting Started (導入)
3. User Guide (ユーザーガイド)
4. Reference (リファレンス)
5. Operations (運用)
6. Development (開発)
7. Release Notes (リリース情報)
8. Community (コミュニティ)
```

### Level 2: Detailed Breakdown

#### 1. **Home**
**Purpose:** プロジェクト概要とナビゲーション起点

- `index.md` - トップページ
- Features overview
- Quick links to key sections

#### 2. **Getting Started** (初心者向け)
**Purpose:** 0→1の導入、初回セットアップ

- `overview.md` - プロジェクト概要
- `how-it-works.md` - 仕組みの理解
- `quickstart.md` - 5分クイックスタート ⭐
- `installation.md` - インストール手順 (新規)
- `first-drift-detection.md` - 最初のドリフト検知 (新規)

**統合すべきファイル:**
- ❌ 削除: `GETTING_STARTED.md` (quickstart.mdに統合)
- ❌ 削除: `zenn-getting-started.md`, `qiita-getting-started-*` (外部記事)

#### 3. **User Guide** (使い方・操作方法)
**Purpose:** 日常的な使用方法、ユースケース

```
3.1 Basic Usage
  - usage.md (USAGE.mdをリネーム)
  - configuration.md (設定ファイル解説)
  - filtering-and-alerting.md (フィルタリング・通知)

3.2 Use Cases
  - use-cases/index.md (USE_CASES.mdをリネーム)
  - use-cases/compliance-monitoring.md (コンプライアンス監視)
  - use-cases/incident-response.md (インシデント対応)
  - use-cases/security-auditing.md (セキュリティ監査)

3.3 Best Practices
  - best-practices.md (BEST_PRACTICES.mdをリネーム)
  - performance-tuning.md (パフォーマンスチューニング)
  - false-positive-reduction.md (誤検知削減)
```

#### 4. **Reference** (リファレンス)
**Purpose:** 詳細な技術情報、API、設定

```
4.1 API Reference
  - api/rest-api.md (API.mdをリネーム) ⭐
  - api/websocket.md (WebSocket API)
  - api/sse.md (Server-Sent Events)
  - api/authentication.md (認証)

4.2 Architecture
  - architecture/overview.md (architecture.mdをリネーム)
  - architecture/system-design.md (IMPLEMENTATION_SUMMARY.mdから抽出) ⭐
  - architecture/data-flow.md (データフロー)
  - architecture/ui-architecture.md (GRAPH_UI_ARCHITECTURE.mdをリネーム)

4.3 Configuration
  - config/reference.md (設定リファレンス完全版)
  - config/environment-variables.md (環境変数)
  - config/falco-rules.md (Falcoルール)

4.4 Service Coverage
  - services/index.md (維持)
  - [AWS Services] (既存維持)
  - [GCP Services] (既存維持)
```

#### 5. **Operations** (運用・デプロイ)
**Purpose:** 本番環境での運用

```
5.1 Deployment
  - deployment/docker.md (deployment.mdから分割)
  - deployment/kubernetes.md (deployment.mdから分割)
  - deployment/production.md (PRODUCTION_READINESS.mdをリネーム) ⭐
  - deployment/cloud-providers.md (AWS/GCP/Azure)

5.2 Setup
  - setup/falco-aws.md (falco-setup.mdをリネーム)
  - setup/falco-gcp.md (gcp-setup.mdをリネーム)
  - setup/auto-import.md (auto-import-guide.mdをリネーム)

5.3 Monitoring
  - monitoring/metrics.md (メトリクス監視)
  - monitoring/logging.md (ログ管理)
  - monitoring/alerting.md (アラート設定)
  - monitoring/dashboards.md (Grafanaダッシュボード)

5.4 Troubleshooting
  - troubleshooting/common-issues.md (一般的な問題)
  - troubleshooting/debugging.md (デバッグ)
  - troubleshooting/faq.md (FAQ)
```

#### 6. **Development** (開発者向け)
**Purpose:** コードベースへの貢献、拡張

```
6.1 Contributing
  - CONTRIBUTING.md (維持)
  - development-setup.md (開発環境セットアップ)
  - coding-standards.md (コーディング規約)

6.2 Extending
  - extending/overview.md (EXTENDING.mdをリネーム)
  - extending/custom-providers.md (カスタムプロバイダー)
  - extending/custom-rules.md (カスタムルール)
  - extending/plugins.md (プラグイン開発)

6.3 Testing
  - testing/unit-tests.md (ユニットテスト)
  - testing/integration-tests.md (統合テスト)
  - testing/coverage.md (test-coverage-*を統合)

6.4 Design Documents
  - design/gcp-implementation.md (GCP_IMPLEMENTATION_DESIGN.mdをリネーム)
  - design/diff-formats.md (diff-formats.mdをリネーム)
```

#### 7. **Release Notes** (既存維持)
**Purpose:** バージョン履歴

- v0.5.0
- v0.3.0
- v0.2.0-beta
- Architecture Changes

#### 8. **Community** (既存維持)
**Purpose:** コミュニティ・ガバナンス

- CONTRIBUTING.md
- CODE_OF_CONDUCT.md
- SECURITY.md

---

## Migration Plan

### Phase 1: Critical Files (即座に対応)

**新規追加:**
- ✅ `api/rest-api.md` (API.mdをリネーム・移動)
- ✅ `architecture/system-design.md` (IMPLEMENTATION_SUMMARY.mdから抽出)
- ✅ `deployment/production.md` (本番環境デプロイ)

**統合:**
- `quickstart.md` ← `GETTING_STARTED.md`
- `best-practices.md` ← `BEST_PRACTICES.md` (リネーム)
- `use-cases/index.md` ← `USE_CASES.md` (移動)

### Phase 2: Reorganization (段階的に対応)

**ディレクトリ作成:**
```
docs/
├── api/                  # API Reference
├── architecture/         # Architecture docs
├── config/              # Configuration
├── deployment/          # Deployment guides
├── setup/              # Setup guides
├── monitoring/         # Monitoring & observability
├── troubleshooting/    # Troubleshooting
├── development/        # Development guides
├── use-cases/          # Use case examples
└── design/             # Design documents
```

**ファイル移動:**
- Move `API.md` → `api/rest-api.md`
- Move `architecture.md` → `architecture/overview.md`
- Move `BEST_PRACTICES.md` → `best-practices.md`
- Move `USE_CASES.md` → `use-cases/index.md`

### Phase 3: Cleanup (不要ファイル整理)

**削除対象 (外部記事):**
- `qiita-*.md` (13ファイル)
- `zenn-*.md` (2ファイル)
- `PR_v0.2.0-beta.md`
- `test-coverage-improvement-*.md` (開発ログ)

**アーカイブ対象:**
- `docs/archive/` ディレクトリに移動
- 履歴として保持、ただしmkdocs.ymlからは除外

---

## File Count Comparison

### Before
- Total: 75 files
- Published: 29 files
- Unpublished: 46 files

### After (Proposed)
- Total: ~50 files (重複削除後)
- Published: ~45 files (カバレッジ大幅向上)
- Archived: ~25 files (外部記事・開発ログ)

**Coverage Improvement:** 29 → 45 files (+55%)

---

## Benefits of MECE Structure

### ✅ Mutually Exclusive (相互排他的)
- No content duplication
- Clear boundaries between categories
- Single source of truth for each topic

### ✅ Collectively Exhaustive (網羅的)
- All user needs covered
- API documentation included
- Production deployment covered
- Development guidelines included

### ✅ User Journey Optimized
- **Beginner:** Getting Started → User Guide
- **Operator:** Operations → Monitoring
- **Developer:** Development → Reference
- **Architect:** Architecture → Design Docs

### ✅ Maintainability
- Clear file organization
- Consistent naming conventions
- Logical grouping
- Easy to find and update

---

## Next Steps

1. Create new directory structure
2. Move/rename critical files
3. Update mkdocs.yml navigation
4. Create missing documents
5. Archive external articles
6. Update cross-references
7. Publish to GitHub Pages

---

**Prepared by:** Claude Code
**Review Required:** Yes
**Implementation Priority:** High
