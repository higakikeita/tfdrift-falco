# TFDrift-Falco Grafana - ユーザー向けクイックガイド

## 🚀 すぐに始める（3ステップ）

### 1. Grafana を起動

```bash
cd dashboards/grafana
./quick-start.sh
```

自動でブラウザが開き、Grafana にアクセスできます。

### 2. ログイン

```
URL: http://localhost:3000
ユーザー名: admin
パスワード: admin
```

### 3. ダッシュボードを開く

左サイドバー → **Dashboards** → **TFDrift-Falco Overview**

サンプルデータが表示されます！

---

## 📊 3つのダッシュボード

### 1. TFDrift-Falco Overview（概要）
**用途**: 全体像の把握、日次レビュー

**見るべきポイント**:
- 総ドリフト数
- 深刻度別の内訳（円グラフ）
- 最新のイベント一覧

**使い方**:
- 右上で時間範囲を選択（Last 6 hours、Last 24 hours）
- Auto-refresh を有効化（30s 推奨）

### 2. TFDrift-Falco Diff Details（差分詳細）
**用途**: インシデント調査、設定変更の確認

**見るべきポイント**:
- 期待値 vs 実際の値
- 誰が変更したか
- 最もドリフトが多いリソース

**使い方**:
- 検索バーでリソース ID を検索
- Diff 内容を確認

### 3. TFDrift-Falco Heatmap & Analytics（分析）
**用途**: パターン分析、トレンド把握

**見るべきポイント**:
- 時間帯別のドリフト発生頻度
- リソース種別の活動状況

**使い方**:
- Heatmap で異常な時間帯を特定
- 定期的なパターンを見つける

---

## 🔗 実際のデータと連携

### 最も簡単な方法

TFDrift-Falco のログファイルパスを Promtail にマウントします。

**Step 1**: TFDrift-Falco が JSON ログを出力するよう設定

`config.yaml`:
```yaml
output:
  file:
    enabled: true
    path: /var/log/tfdrift/drift-events.jsonl
    format: json
```

**Step 2**: `docker-compose.yaml` を編集

```yaml
services:
  promtail:
    volumes:
      # 実際のログパスに変更
      - /var/log/tfdrift:/var/log/tfdrift:ro
      - ./promtail-config.yaml:/etc/promtail/config.yml
```

**Step 3**: 再起動

```bash
docker-compose restart promtail
```

数秒後、実際のデータが Grafana に表示されます！

---

## 🚨 アラート設定（オプション）

重要なドリフトを Slack で通知できます。

### Slack Webhook の作成

1. https://api.slack.com/apps にアクセス
2. 新しいアプリを作成 → Incoming Webhooks を有効化
3. Webhook URL をコピー

### 環境変数を設定

```bash
cd dashboards/grafana
echo 'SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL' > .env
docker-compose down && docker-compose up -d
```

### Grafana でアラートを作成

詳細は [ALERTS.md](./ALERTS.md) を参照してください。

**最も重要なアラート**: Critical Drift Detected

1. Grafana → Alerting → Alert rules → + New alert rule
2. 設定:
   - Query: `count_over_time({job="tfdrift-falco"} | json | severity="critical" [5m])`
   - Threshold: `> 1`
   - Contact point: `slack-tfdrift`
3. Save

---

## 💡 よくある使い方

### ユースケース 1: 毎朝のセキュリティレビュー

1. **TFDrift-Falco Overview** を開く
2. 時間範囲を **Last 24 hours** に設定
3. Critical/High を確認
4. 必要に応じて AWS Console で修正

### ユースケース 2: インシデント調査

1. **TFDrift-Falco Diff Details** を開く
2. リソース ID で検索（例: `sg-123456`）
3. 何が変更されたか確認
4. 誰が変更したか特定

### ユースケース 3: 月次監査レポート

1. **TFDrift-Falco Overview** を開く
2. 時間範囲を **Last 30 days** に設定
3. Dashboard → Share → Export → PDF
4. レポートをダウンロード

---

## 🔧 トラブルシューティング

### データが表示されない

```bash
# TFDrift-Falco が動いているか確認
docker-compose ps tfdrift

# ログを確認
docker-compose logs tfdrift

# Promtail がログを収集しているか確認
docker-compose logs promtail | grep -i error
```

### ログインできない

```bash
# Grafana を再起動
docker-compose restart grafana

# デフォルト認証情報
# Username: admin
# Password: admin
```

---

## 📚 詳細ドキュメント

- **[完全版セットアップガイド](./GETTING_STARTED.md)** - 詳細な手順と例
- **[アラート設定ガイド](./ALERTS.md)** - 6つのアラートルール
- **[カスタマイズガイド](./CUSTOMIZATION_GUIDE.md)** - 独自のパネル作成
- **[統合テスト結果](./INTEGRATION_TEST_RESULTS.md)** - 技術詳細

---

## ✨ 次のステップ

1. ✅ サンプルデータで Grafana を試す
2. ✅ 実際の TFDrift-Falco ログを連携
3. ✅ Slack アラートを設定
4. ✅ チーム専用のダッシュボードを作成

**質問がある場合**:
- GitHub Issues: https://github.com/your-org/tfdrift-falco/issues
- Slack: #tfdrift-falco
- Email: support@your-org.com

---

**最終更新**: 2025-12-05
