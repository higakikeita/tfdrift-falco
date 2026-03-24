#!/usr/bin/env bash
# bump-version.sh — Update version across all project files
#
# Usage:
#   ./scripts/bump-version.sh 0.9.0
#
# Updates:
#   - VERSION                          (plain text)
#   - mkdocs.yml                       (extra.project_version)
#   - ui/package.json                  (version field)
#   - charts/tfdrift-falco/Chart.yaml  (appVersion)

set -euo pipefail

NEW_VERSION="${1:-}"

if [[ -z "$NEW_VERSION" ]]; then
  echo "Usage: $0 <new-version>"
  echo "Example: $0 0.9.0"
  exit 1
fi

# Validate semver format (loose)
if ! [[ "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "Error: '$NEW_VERSION' is not a valid semver (expected X.Y.Z)"
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OLD_VERSION="$(cat "$REPO_ROOT/VERSION" | tr -d '[:space:]')"

echo "Bumping version: $OLD_VERSION → $NEW_VERSION"
echo ""

# 1. VERSION file
echo "$NEW_VERSION" > "$REPO_ROOT/VERSION"
echo "  ✓ VERSION"

# 2. mkdocs.yml — project_version
sed -i.bak "s/project_version: \"$OLD_VERSION\"/project_version: \"$NEW_VERSION\"/" "$REPO_ROOT/mkdocs.yml"
rm -f "$REPO_ROOT/mkdocs.yml.bak"
echo "  ✓ mkdocs.yml (extra.project_version)"

# 3. ui/package.json — version field
sed -i.bak "s/\"version\": \"$OLD_VERSION\"/\"version\": \"$NEW_VERSION\"/" "$REPO_ROOT/ui/package.json"
rm -f "$REPO_ROOT/ui/package.json.bak"
echo "  ✓ ui/package.json"

# 4. Chart.yaml — appVersion
sed -i.bak "s/appVersion: \"$OLD_VERSION\"/appVersion: \"$NEW_VERSION\"/" "$REPO_ROOT/charts/tfdrift-falco/Chart.yaml"
rm -f "$REPO_ROOT/charts/tfdrift-falco/Chart.yaml.bak"
echo "  ✓ charts/tfdrift-falco/Chart.yaml (appVersion)"

echo ""
echo "Done. Files updated:"
echo ""
git -C "$REPO_ROOT" diff --stat
echo ""
echo "Next steps:"
echo "  1. Review changes: git diff"
echo "  2. Commit: git add -A && git commit -m 'chore: bump version to $NEW_VERSION'"
echo "  3. Tag: git tag v$NEW_VERSION"
echo "  4. Push: git push origin main --tags"
