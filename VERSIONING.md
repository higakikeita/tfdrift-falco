# Versioning Policy

TFDrift-Falco follows [Semantic Versioning 2.0.0](https://semver.org/).

Format: `MAJOR.MINOR.PATCH`

---

## Current Version: 0.9.0

## Version History

| Version | Date | Highlights |
|---------|------|------------|
| 0.9.0 | 2026-03-29 | Azure FullProvider, azurerm backend, WebSocket enhancements |
| 0.8.0 | 2026-03-22 | Enterprise Foundation (JWT, Helm, OpenAPI, Operations Runbook) |
| 0.7.0 | 2026-03-22 | Dashboard UI, versioning policy |
| 0.6.0 | 2026-03-xx | UI improvements (colors, Storybook) |
| 0.5.0 | 2025-12-17 | Multi-Cloud GCP support |
| 0.2.0-beta | 2025-12-05 | VPC networking, code quality |
| 0.1.0 | 2024-11-xx | Initial release (AWS only) |

---

## Rules

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

## Release Process

1. Create a release branch: `release/vX.Y.Z`
2. Update `VERSION` file
3. Update `CHANGELOG.md` (move Unreleased items to new version section)
4. Update version badges in `README.md` and `README.ja.md`
5. Update `docs/architecture.md` and `docs/overview.md` version references
6. Create release notes in `docs/release-notes/vX.Y.Z.md`
7. Merge to `main` via PR
8. Tag the merge commit: `git tag vX.Y.Z`
9. Create GitHub Release with notes

---

## Notes

- Pre-1.0 versions (0.x.y) may include minor breaking changes in MINOR releases, documented in CHANGELOG
- The `VERSION` file is the single source of truth for the current version
- All version references across docs must be updated atomically in the release commit
- Git tags must match the `VERSION` file exactly (prefixed with `v`)
