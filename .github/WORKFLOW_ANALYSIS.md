# GitHub Actions ワークフロー分析レポート

## 📋 既存ワークフロー一覧

| ファイル | 目的 | 実行タイミング | 状態 |
|---------|------|--------------|------|
| **ci.yml** | メインCI/CD | push, PR, release | ⚠️ 改善必要 |
| **test.yml** | 単体テスト | push, PR | ✅ 良好 |
| **lint.yml** | コード品質チェック | push, PR | ✅ 良好 |
| **security.yml** | セキュリティスキャン | push, PR, 毎週月曜 | ⚠️ 要確認 |
| **e2e.yml** | E2Eテスト | ラベル `run-e2e` | ✅ 良好 |
| **integration.yml** | 統合テスト | push, PR | ✅ 良好 |
| **benchmark.yml** | ベンチマーク | push, PR | ✅ 良好 |
| **docs.yml** | ドキュメント | push (docs/) | ✅ 良好 |
| **publish-ghcr.yml** | イメージ公開 | release | ✅ 良好 |
| **website-security.yml** | Webセキュリティ | 毎週 | ✅ 良好 |

---

## 🔴 重大な問題

### 1. Snyk Token が未設定の可能性

**場所**: `.github/workflows/security.yml:35`

```yaml
env:
  SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}
```

**問題**:
- `SNYK_TOKEN` がGitHub Secretsに設定されていない場合、Snykスキャンが失敗
- `continue-on-error: true` により失敗が隠蔽される

**影響**: セキュリティ脆弱性が検出されない

**対策**:
1. Snyk アカウントを作成: https://snyk.io/
2. トークンを取得
3. GitHub Settings > Secrets に `SNYK_TOKEN` を追加

---

### 2. Frontend CI でエラーが無視される

**場所**: `.github/workflows/ci.yml:107-113`

```yaml
- name: Run linter
  run: npm run lint || true

- name: Run type check
  run: npm run type-check || npx tsc --noEmit || true

- name: Run tests
  run: npm test -- --passWithNoTests || true
```

**問題**:
- すべてのチェックが `|| true` で失敗を無視
- TypeScript エラーやテスト失敗が検出されない

**影響**: 品質が低下したコードがマージされる可能性

**対策**: `|| true` を削除し、エラーを検出できるようにする

---

### 3. ワークフローの重複

**場所**: `ci.yml` と `security.yml`

- **Gosec**: 両方で実行（二重実行）
- **Staticcheck**: `ci.yml` と `lint.yml` で実行（二重実行）

**問題**: CI実行時間が無駄に長くなる

**対策**:
- セキュリティスキャンは `security.yml` に集約
- Staticcheckは `lint.yml` のみで実行

---

## ⚠️ 改善が必要な項目

### 4. UI関連の新規ワークフローが統合されていない

**場所**: `ui/.github/workflows/`

新規作成されたワークフロー：
- `ui/.github/workflows/ci.yml` - UI CI
- `ui/.github/workflows/chromatic.yml` - ビジュアルリグレッション
- `ui/.github/workflows/lighthouse-scheduled.yml` - パフォーマンス監視
- `ui/.github/workflows/storybook-deploy.yml` - Storybook公開

**問題**:
- ルートの `.github/workflows/` ではなく `ui/.github/workflows/` に配置
- GitHub Actionsは **ルートの `.github/workflows/` のみ認識**
- これらのワークフローは **現在動作していない**

**対策**: `ui/.github/workflows/` から `.github/workflows/` に移動

---

### 5. Codecov Token が未設定の可能性

**場所**: 複数箇所

```yaml
env:
  CODECOV_TOKEN: ${{ secrets.CODECOV_TOKEN }}
```

**状態**: `fail_ci_if_error: false` により失敗は無視される

**推奨**:
- Codecov連携を使う場合: トークン設定
- 使わない場合: アップロード処理を削除

---

## ✅ 良好な点

1. **包括的なテストカバレッジ**: 単体、統合、E2E、ベンチマーク
2. **セキュリティスキャン**: Gosec, Nancy, Trivy が設定済み
3. **PR コメント**: カバレッジとテスト結果を自動コメント
4. **キャッシング**: Go modules と npm のキャッシュが適切に設定
5. **マルチプラットフォームビルド**: linux/amd64, linux/arm64 対応

---

## 🔧 推奨される修正

### 優先度: 高 🔴

1. **UI ワークフローの移動**
   ```bash
   mv ui/.github/workflows/*.yml .github/workflows/
   ```

2. **Frontend CI のエラー無視を修正**
   - `ci.yml` の `|| true` を削除
   - または別ファイルに分離

3. **Snyk Token の設定**
   - Snyk アカウント作成
   - トークンをGitHub Secretsに追加

### 優先度: 中 🟡

4. **ワークフローの重複解消**
   - Gosec を `security.yml` のみに
   - Staticcheck を `lint.yml` のみに

5. **Codecov 連携の決定**
   - 使う → トークン設定
   - 使わない → コード削除

### 優先度: 低 🟢

6. **ワークフロー統合の検討**
   - 関連するジョブを1つのファイルにまとめる
   - または明確な責任分担

---

## 📝 必要な GitHub Secrets

以下のシークレットを設定してください：

| Secret名 | 必須 | 用途 | 取得方法 |
|---------|------|------|---------|
| `SNYK_TOKEN` | ⚠️ 推奨 | 脆弱性スキャン | https://snyk.io/ |
| `CODECOV_TOKEN` | オプション | カバレッジ追跡 | https://codecov.io/ |
| `DOCKER_USERNAME` | リリース時 | Docker Hub | Docker Hub設定 |
| `DOCKER_PASSWORD` | リリース時 | Docker Hub | Docker Hub設定 |
| `CHROMATIC_PROJECT_TOKEN` | UI使用時 | ビジュアルテスト | https://www.chromatic.com/ |

---

## 🎯 推奨アクション

### すぐに実行すべき

```bash
# 1. UI ワークフローを正しい場所に移動
cd /path/to/tfdrift-falco
mv ui/.github/workflows/ci.yml .github/workflows/ui-ci.yml
mv ui/.github/workflows/chromatic.yml .github/workflows/ui-chromatic.yml
mv ui/.github/workflows/lighthouse-scheduled.yml .github/workflows/ui-lighthouse.yml
mv ui/.github/workflows/storybook-deploy.yml .github/workflows/ui-storybook.yml
rm -rf ui/.github  # UIディレクトリのGitHub設定を削除
```

### Snyk設定

1. https://snyk.io/ でアカウント作成
2. Settings > General > API Token からトークン取得
3. GitHub リポジトリ > Settings > Secrets and variables > Actions
4. "New repository secret" をクリック
5. Name: `SNYK_TOKEN`, Value: (コピーしたトークン)

---

## 📊 期待される改善効果

| 項目 | 改善前 | 改善後 |
|------|--------|--------|
| **UI CI** | ❌ 動作せず | ✅ 動作 |
| **セキュリティスキャン** | ⚠️ 一部のみ | ✅ 完全 |
| **Frontend 品質** | ⚠️ エラー無視 | ✅ 検出 |
| **CI 実行時間** | 重複あり | 最適化 |
| **ワークフロー整理** | 分散 | 統合 |
