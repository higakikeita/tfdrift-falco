# Versioning Policy

TFDrift-Falco follows [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html) with the following project-specific guidelines.

## Pre-1.0 Versioning (Current)

While the project is in the `0.x.y` phase, the following rules apply:

### Patch (`0.x.Y`) — Bug Fixes & Minor Improvements

Increment for changes that do not add new features or change behavior.

- Bug fixes
- Documentation typos and corrections
- Minor refactoring with no behavior change
- Dependency updates (non-breaking)
- CI/CD pipeline fixes
- Test additions or improvements (no production code change)

Examples: `v0.6.0 → v0.6.1`

### Minor (`0.X.0`) — New Features & Enhancements

Increment for any of the following:

- New user-facing feature (UI page, API endpoint, CLI flag)
- New cloud provider or service support (AWS services, GCP services, Azure)
- Major UI overhaul or new dashboard capability
- New integration (webhook type, notification channel)
- Breaking API changes (acceptable in pre-1.0)
- Significant architecture changes

Examples: `v0.6.0 → v0.7.0`

### When to Bump

| Change Type | Version Bump | Example |
|---|---|---|
| Fix crash in events API | Patch | `0.7.0 → 0.7.1` |
| Update README / docs only | Patch | `0.7.0 → 0.7.1` |
| Add graph export feature | Minor | `0.7.0 → 0.8.0` |
| Add Azure provider support | Minor | `0.7.0 → 0.8.0` |
| Add 10 new AWS service mappings | Minor | `0.7.0 → 0.8.0` |
| Fix typo in error message | Patch | `0.7.0 → 0.7.1` |
| Add detector unit tests | Patch | `0.7.0 → 0.7.1` |

## 1.0.0 Criteria

The project will be promoted to `1.0.0` when all of the following are met:

- [ ] Falco gRPC integration stable and battle-tested
- [ ] Multi-cloud support: AWS, GCP, and Azure all functional
- [ ] End-to-end test suite covering critical paths
- [ ] Dashboard UI feature-complete with production usage
- [ ] API stability (no breaking changes expected)
- [ ] Documentation comprehensive and up to date
- [ ] At least one production deployment reference

After `1.0.0`, standard semver applies strictly:

- **Major**: Breaking API or configuration changes
- **Minor**: New features, backward-compatible
- **Patch**: Bug fixes, backward-compatible

## Post-1.0 Versioning

### Major (`X.0.0`)

- Breaking changes to REST API contracts
- Breaking changes to `config.yaml` schema
- Breaking changes to CLI flags or behavior
- Removal of deprecated features

### Minor (`x.Y.0`)

- New features (backward-compatible)
- New cloud provider support
- New API endpoints (additive)
- New configuration options (with defaults)

### Patch (`x.y.Z`)

- Bug fixes
- Security patches
- Performance improvements
- Documentation updates

## Release Process

1. Update `CHANGELOG.md` with all changes under the new version
2. Update version references in code (`README.md` badge, `pkg/version/version.go`, etc.)
3. Create a PR titled `release: vX.Y.Z`
4. Merge to `main`
5. Tag with `git tag vX.Y.Z` and push the tag
6. GitHub Actions publishes the Docker image automatically

## Version History

| Version | Date | Highlights |
|---|---|---|
| `v0.2.0-beta` | 2025-12 | MVP: AWS CloudTrail, Falco gRPC, Slack |
| `v0.3.0` | 2025-12 | Enhanced AWS (203 events, 19 services) |
| `v0.4.0` | 2026-01 | NDJSON output, structured events |
| `v0.4.1` | 2026-01 | Webhook integration (Slack, Teams, custom) |
| `v0.5.0` | 2026-01 | GCP Audit Logs, UI improvements |
| `v0.6.0` | 2026-03-20 | Multi-cloud expansion (AWS 500+, GCP 170+) |
| `v0.7.0` | 2026-03-22 | Dashboard UI: events mgmt, notifications, graph export, settings |
