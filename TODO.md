# TFDrift-Falco TODO List

## 🔥 Critical Issues (今すぐ修正)

### 1. Docker Compose設定の修正
- [ ] `docker-compose.yml`の`version`フィールドを削除（廃止された）
- [ ] `backend`サービスにAWS_PROFILE環境変数を追加
- [ ] Falcoの依存関係を条件付きにする（Falco無効時もバックエンドが起動できるように）

```yaml
# 修正例
environment:
  - AWS_PROFILE=${AWS_PROFILE:-mytf}
  - AWS_REGION=${AWS_REGION:-us-east-1}
```

### 2. AWS認証情報の設定
- [ ] バックエンドコンテナにAWS認証情報を正しく渡す
- [ ] `~/.aws`マウント時のパーミッション確認
- [ ] AWS_PROFILEのデフォルト値設定

**現在のエラー:**
```
NoCredentialProviders: no valid providers in chain
```

### 3. Falcoルールの修正
- [ ] `rules/terraform_drift.yaml`の構文エラーを修正
- [ ] CloudTrailプラグインの正しいフィールド名を使用
- [ ] Priority値を正しい値に修正（HIGH → WARNING）

**問題箇所:**
- Line 13, 32, 54: `account=%ct.account)` - 閉じ括弧が多い
- Line 44: `priority: HIGH` - 無効な値

---

## 📚 ドキュメント更新 (高優先度)

### 4. ARM64 Mac対応ドキュメント作成
- [ ] `docs/setup-arm64-mac.md`を作成
- [ ] docker-composeで`platform: linux/amd64`指定の説明
- [ ] Rosetta経由での動作に関する注意事項
- [ ] パフォーマンス影響について記載

### 5. CloudTrail統合ガイド作成
- [ ] `docs/cloudtrail-integration-guide.md`を作成
- [ ] ステップバイステップのセットアップ手順
- [ ] AWS認証情報の設定方法
- [ ] CloudTrailプラグインのトラブルシューティング
- [ ] S3バケット作成からFalco統合までの完全な手順

### 6. トラブルシューティングガイド作成
- [ ] `docs/troubleshooting.md`を作成
- [ ] よくあるエラーと解決方法
  - NoCredentialProviders エラー
  - Falco eBPFドライバーのコンパイル失敗
  - CloudTrailプラグインのS3接続エラー
  - ポート競合エラー

### 7. READMEの更新
- [ ] Quick Startセクションに前提条件を明記
  - Docker Desktop (Rosetta有効)
  - AWS CLI設定済み
  - Terraform 1.0+
- [ ] ARM64 Macでのセットアップ手順リンク追加
- [ ] 環境変数の説明セクション追加

---

## 🔧 コード改善 (中優先度)

### 8. バックエンドの改善
- [ ] AWS認証情報エラー時の適切なフォールバック
- [ ] Terraform state読み込み失敗時にサンプルデータで起動
- [ ] エラーメッセージの改善（ユーザーフレンドリーに）
- [ ] ヘルスチェックエンドポイントをGETに変更（現在HEAD 405エラー）

```go
// 改善例
if err := loadTerraformState(); err != nil {
    log.Warn("Failed to load Terraform state, using sample data")
    loadSampleData()
}
```

### 9. Falco設定の改善
- [ ] `falco-simple.yaml`を環境別に分割
  - `falco-docker-desktop.yaml` (eBPFなし)
  - `falco-production.yaml` (CloudTrail統合)
- [ ] CloudTrailプラグイン設定の環境変数化

### 10. セットアップスクリプトの改善
- [ ] `setup-cloudtrail.sh`にAWS_PROFILE自動検出機能追加
- [ ] エラーハンドリングの強化
- [ ] ロールバック機能の追加
- [ ] 冪等性の保証（複数回実行可能に）

---

## 🎨 UI改善 (低優先度)

### 11. エラー表示の改善
- [ ] APIエラー時にユーザーフレンドリーなメッセージ表示
- [ ] ローディング状態の改善
- [ ] オフライン時の表示

### 12. 実環境データの表示
- [ ] Terraform Stateから読み込んだリソースのグラフ表示
- [ ] 空データ時のプレースホルダー表示
- [ ] サンプルデータとの切り替え機能

---

## 🚀 機能追加 (将来)

### 13. CloudTrail統合の完成
- [ ] AWS認証情報の設定を完了
- [ ] CloudTrailログが利用可能になるまで待機
- [ ] 実際のAWSリソース変更でドリフト検知テスト
- [ ] WebSocket/SSEでリアルタイム通知

### 14. 自動テストの追加
- [ ] バックエンドの統合テスト
- [ ] UIのE2Eテスト
- [ ] docker-composeのヘルスチェックテスト

---

## 📋 優先順位サマリー

### 今日中に完了すべきもの
1. ✅ UIの起動確認（完了）
2. AWS認証情報の設定
3. docker-compose.ymlの修正
4. Falcoルールの修正

### 今週中に完了すべきもの
5. CloudTrail統合ガイドの作成
6. トラブルシューティングガイドの作成
7. ARM64 Mac対応ドキュメント

### 来週以降
8. バックエンドのエラーハンドリング改善
9. 自動テストの追加
10. CloudTrail統合の完成

---

## 📝 メモ

### 判明した技術的制約
- ARM64 Mac環境ではFalco eBPFドライバーがコンパイルできない
- Docker Desktop (linuxkit) ではカーネルヘッダーが利用不可
- CloudTrailプラグインはx86_64のみサポート（Rosetta経由で動作）
- CloudTrailログのS3書き込みには5-15分かかる

### 現在動作している機能
- ✅ Backend API (port 8080)
- ✅ Frontend UI (port 3000)
- ✅ WebSocket/SSE準備完了
- ✅ React UIのグラフ・テーブル表示
- ⚠️ Terraform State読み込み（AWS認証情報エラー）
- ❌ Falco統合（一時的に無効化）

### 次のステップ
1. AWS認証情報を設定
2. Terraform Stateを正常に読み込み
3. グラフに実環境のリソースを表示
4. CloudTrail統合を完成させてリアルタイムドリフト検知
