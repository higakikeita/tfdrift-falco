# Versioning Policy

> **このドキュメントはリリース作業の唯一のガイドラインです。**
> バージョン更新時は必ずこのファイルを参照してください。
> Claude（AI）がリリース作業を行う場合も、このドキュメントに従ってください。
>
> This is the single source of truth for release procedures.
> Always consult this document before any version update.

TFDrift-Falco follows [Semantic Versioning 2.0.0](https://semver.org/).

Format: `MAJOR.MINOR.PATCH`

---

## Current Version: 0.9.0

## Version History

| Version | Date | Git Tag | Highlights |
|---------|------|---------|------------|
| 0.9.0 | 2026-03-29 | v0.9.0 | Azure FullProvider, azurerm backend, WebSocket enhancements |
| 0.8.0 | 2026-03-22 | v0.8.0 | Enterprise Foundation (JWT, Helm, OpenAPI, Operations Runbook) |
| 0.7.0 | 2026-03-22 | v0.7.0 | Dashboard UI, versioning policy |
| 0.6.0 | 2026-03-xx | v0.6.0 | UI improvements (colors, Storybook) |
| 0.5.0 | 2025-12-17 | v0.5.0 | Multi-Cloud GCP support |
| 0.2.0-beta | 2025-12-05 | v0.2.0-beta | VPC networking, code quality |
| 0.1.0 | 2024-11-xx | — | Initial release (AWS only) |

---

## SemVer Rules

### PATCH (0.9.x)

Increment for backward-compatible bug fixes:

- Bug fixes that don't change API or behavior
- Security patches
- Dependency updates (non-breaking)
- Documentation typo fixes
- CI/CD fixes

### MINOR (0.x.0)

Increment for backward-compatible feature additions:

- New cloud provider support (e.g., Azure, GCP)
- New Terraform backend support (e.g., azurerm, gcs)
- New API endpoints
- New event types or detection capabilities
- New UI features or pages
- Performance improvements with no API changes

### MAJOR (x.0.0)

Increment for breaking changes:

- Breaking API changes (endpoint removal, response format changes)
- Configuration format changes requiring migration
- Minimum Go/Node version bumps
- Removal of deprecated features
- Database schema changes requiring migration

---

## 1.0.0 GA Criteria

The project will reach 1.0.0 when all of the following are met:

1. **Three cloud providers fully supported** — AWS, GCP, Azure (Discovery + Comparison + Backend) ✅
2. **Stable API** — No breaking changes planned for /api/v1 endpoints
3. **Production validation** — Deployed and tested in at least one production environment
4. **Documentation complete** — Quickstart, architecture, per-provider guides, API reference
5. **Test coverage** — Core packages above 80%
6. **Security hardened** — Authentication, input validation, secret handling reviewed

---

## Release Checklist（リリース時の更新対象ファイル一覧）

バージョンを更新するときは、以下のファイルを **すべて** 更新してください。
漏れ防止のため、PRの説明にこのチェックリストをコピーして使うことを推奨します。

### 1. Source of Truth

| # | File | What to update | Example |
|---|------|----------------|---------|
| 1 | `VERSION` | バージョン番号（単一行） | `0.9.0` |

### 2. Go Source Code

| # | File | What to update | Notes |
|---|------|----------------|-------|
| 2 | `Makefile` | `VERSION?=` のデフォルト値（L5） | `VERSION?=0.9.0` |
| 3 | `pkg/api/websocket/handler.go` | WebSocket welcome message の `"version"` | `"version": "0.9.0"` |
| 4 | `pkg/api/sse/stream.go` | SSE connection event の `"version"` | `"version": "0.9.0"` |

> **Note:** `Dockerfile` は `git describe --tags` を使うため手動更新不要（タグが正しければ自動反映）。

### 3. Documentation（Markdown）

| # | File | What to update |
|---|------|----------------|
| 5 | `README.md` | version badge (`version-X.Y.Z-blue`)、リリースアナウンスセクション |
| 6 | `README.ja.md` | version badge、リリースアナウンスセクション（日本語） |
| 7 | `CHANGELOG.md` | `[Unreleased]` の下に新バージョンセクションを追加 |
| 8 | `VERSIONING.md`（このファイル） | Version History テーブルに新行追加、Current Version 更新 |
| 9 | `docs/architecture.md` | ヘッダーの Version 行、Supported Providers |
| 10 | `docs/overview.md` | ヘッダーの Version / Providers / Status 行 |
| 11 | `docs/release-notes/vX.Y.Z.md` | **新規作成**。過去のリリースノートをテンプレートとして使う |
| 12 | `PROJECT_ROADMAP.md` | Current version セクション |
| 13 | `CONTRIBUTING.md` | Versioning セクション（ルール変更時のみ） |

### 4. API Specification

| # | File | What to update |
|---|------|----------------|
| 14 | `docs/api/openapi.yaml` | `info.version` フィールド |
| 15 | `pkg/api/handlers/openapi.yaml` | `info.version` フィールド（埋め込み用コピー） |

### 5. Helm Chart（Kubernetes）

| # | File | What to update |
|---|------|----------------|
| 16 | `charts/tfdrift-falco/Chart.yaml` | `appVersion` フィールド。`version`（Chart自体）も必要に応じて |

### 6. Website（Next.js — `website/` ディレクトリ）

| # | File | What to update |
|---|------|----------------|
| 17 | `website/app/page.tsx` | リリースバナー文言（L45付近）、`StatCard` の `number` prop（L114付近） |
| 18 | `website/content/blog/` | **新規作成**: `vXYZ-<slug>.mdx`（MINOR以上で推奨） |

### 7. 更新不要なもの

以下は **リリース時に更新しないでください**:

- Go コード内のコメント（`// Added in v0.5.0` 等）→ 歴史的記録であり、リリースバージョンではない
- `go.mod` の module path → バージョンと無関係
- `ui/package.json` の `version` → 現在未使用（将来使う場合はチェックリストに追加）
- `package-lock.json` 内のバージョン → npm 依存関係であり、プロジェクトバージョンではない

---

## Release Process

```bash
# 0. 次のバージョン番号を決める（必ず最初に実行）
echo "Latest tag: $(git tag -l 'v*' | sort -V | tail -1)"
echo "VERSION file: $(cat VERSION)"
# → 上記の大きい方 + 1 が次のバージョン

# 1. リリースブランチ作成
git checkout -b release/vX.Y.Z

# 2. 上記チェックリストの全ファイルを更新

# 3. ビルド & テスト確認
make build
go test $(go list ./... | grep -v 'tests/e2e\|tests/load')

# 4. コミット
git add -A
git commit -m "release: vX.Y.Z — <one-line summary>"

# 5. PR作成 & レビュー & マージ
gh pr create --title "release: vX.Y.Z" --body "$(cat <<'EOF'
## Release vX.Y.Z

### Checklist
- [ ] VERSION file
- [ ] Makefile VERSION
- [ ] WebSocket handler.go version
- [ ] SSE stream.go version
- [ ] README.md badge + announcement
- [ ] README.ja.md badge + announcement
- [ ] CHANGELOG.md
- [ ] VERSIONING.md (this file)
- [ ] docs/architecture.md
- [ ] docs/overview.md
- [ ] docs/release-notes/vX.Y.Z.md (new)
- [ ] PROJECT_ROADMAP.md
- [ ] openapi.yaml (both copies)
- [ ] charts/Chart.yaml appVersion
- [ ] website/app/page.tsx
- [ ] website/content/blog/ (if MINOR+)
EOF
)"

# 6. タグ付け（mainマージ後）
git checkout main && git pull
git tag vX.Y.Z
git push origin vX.Y.Z

# 7. GitHub Release 作成
gh release create vX.Y.Z \
  --title "vX.Y.Z — <summary>" \
  --notes-file docs/release-notes/vX.Y.Z.md
```

---

## Anti-Patterns（やってはいけないこと）

過去にバージョニングの混乱が起きた原因を記録します。再発防止のために確認してください。

### 1. バージョンの巻き戻し禁止

- ❌ v0.8.0 の後に v0.6.0 をリリースする
- ✅ 常に `VERSION` ファイルと `git tag -l 'v*' | sort -V | tail -1` を確認してから次のバージョンを決める

### 2. 複数箇所でのバージョン不整合

- ❌ `VERSION` = 0.9.0、`openapi.yaml` = 0.8.0、`website/page.tsx` = 0.6.0
- ✅ 上記チェックリストを使って全箇所を一括更新する

### 3. バージョンスキップ禁止（原則）

- ❌ 0.2.0 → 0.5.0（0.3.0, 0.4.0 をスキップ）
- ✅ 機能規模に関係なく順番にインクリメントする
- ※ pre-1.0 の過去のスキップは歴史的経緯として許容（修正不要）

### 4. タグなしリリース禁止

- ❌ VERSION ファイルだけ更新して git tag を打たない
- ✅ main マージ後に必ず `git tag vX.Y.Z && git push origin vX.Y.Z` を実行

### 5. AI エージェント（Claude 等）への注意事項

- リリース作業を行う前に、必ず `git tag -l 'v*' | sort -V` で最新タグを確認すること
- `VERSION` ファイルの内容と最新タグが一致していることを確認すること
- 不一致がある場合は人間に確認を取ること
- このドキュメント（`VERSIONING.md`）を読んでからリリース作業を開始すること
- バージョン番号を「推測」せず、必ず上記コマンドで確認すること

---

## Quick Reference

```bash
# 現在のバージョン（source of truth）
cat VERSION

# Git タグ一覧（リリース済みバージョン）
git tag -l 'v*' | sort -V

# 全ファイルのバージョン参照を検索（監査用）
grep -rn "$(cat VERSION)" \
  --include='*.go' --include='*.md' \
  --include='*.yaml' --include='*.tsx' .

# 次のバージョンを決める
LATEST=$(git tag -l 'v*' | sort -V | tail -1)
echo "Latest tag: $LATEST"
echo "VERSION file: $(cat VERSION)"
echo "Next version should be greater than both"
```

---

## Notes

- Pre-1.0 versions (0.x.y) may include minor breaking changes in MINOR releases, documented in CHANGELOG
- The `VERSION` file is the single source of truth for the current version
- All version references across docs must be updated atomically in the release commit
- Git tags must match the `VERSION` file exactly (prefixed with `v`)
- このドキュメント自体もリリース時に更新対象です（Version History テーブル + Current Version）
