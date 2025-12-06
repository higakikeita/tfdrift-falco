# TFDrift-Falco Grafana - Getting Started Guide

このガイドでは、TFDrift-Falco の Grafana ダッシュボードをゼロから立ち上げて使い始める方法を説明します。

## 前提条件

- Docker Desktop がインストールされている
- TFDrift-Falco プロジェクトをクローン済み

## 📋 5分クイックスタート

### Step 1: Grafana スタックを起動

```bash
# プロジェクトのルートディレクトリに移動
cd /path/to/tfdrift-falco

# Grafana ディレクトリに移動
cd dashboards/grafana

# Docker で起動（バックグラウンド実行）
docker-compose up -d
```

### Step 2: ブラウザでアクセス

```
URL: http://localhost:3000
ユーザー名: admin
パスワード: admin
```

初回ログイン時にパスワード変更を求められます（スキップ可能）。

### Step 3: ダッシュボードを開く

1. 左サイドバーの **Dashboards** をクリック
2. **TFDrift-Falco Overview** を開く

これでサンプルデータが表示されます！

---

## 🔄 実際のデータと連携する

### Option A: Docker Compose で TFDrift-Falco と連携

#### 1. メインの docker-compose.yml を使う

```bash
# プロジェクトルートに戻る
cd /path/to/tfdrift-falco

# TFDrift-Falco と Grafana を同時起動
docker-compose -f docker-compose.yml up -d
```

このコマンドで以下が起動します：
- Falco
- TFDrift-Falco アプリケーション
- Grafana
- Loki
- Promtail

#### 2. TFDrift-Falco にログ出力設定を追加

`config.yaml` を編集：

```yaml
# TFDrift-Falco の設定
output:
  # 標準出力（Docker logs）
  stdout: true

  # ファイル出力（Grafana 連携用）
  file:
    enabled: true
    path: /var/log/tfdrift/drift-events.jsonl
    format: json

# Falco 接続設定
falco:
  hostname: falco
  port: 5060
  tls: false

# AWS 設定
aws:
  region: us-east-1
  profile: default
```

#### 3. docker-compose.yml にログボリュームを追加

プロジェクトルートの `docker-compose.yml` を編集：

```yaml
services:
  tfdrift:
    # ... 既存の設定 ...
    volumes:
      - ./config.yaml:/config/config.yaml:ro
      - ${HOME}/.aws:/root/.aws:ro

      # ログファイル用ボリュームを追加
      - tfdrift-logs:/var/log/tfdrift
    # ... 残りの設定 ...

# Grafana スタックの設定を追加
  grafana:
    extends:
      file: ./dashboards/grafana/docker-compose.yaml
      service: grafana
    depends_on:
      - loki

  loki:
    extends:
      file: ./dashboards/grafana/docker-compose.yaml
      service: loki

  promtail:
    extends:
      file: ./dashboards/grafana/docker-compose.yaml
      service: promtail
    volumes:
      # TFDrift ログを Promtail にマウント
      - tfdrift-logs:/var/log/tfdrift:ro
      - ./dashboards/grafana/promtail-config.yaml:/etc/promtail/config.yml

volumes:
  tfdrift-logs:
    name: tfdrift-logs
```

#### 4. 再起動

```bash
docker-compose down
docker-compose up -d
```

### Option B: 既存の TFDrift-Falco インスタンスと連携

既に TFDrift-Falco が稼働している場合：

#### 1. ログファイルの場所を確認

```bash
# TFDrift-Falco のログファイルパスを確認
grep -A5 "output:" /path/to/your/config.yaml
```

例: `/var/log/tfdrift/drift-events.jsonl`

#### 2. Promtail にログパスをマウント

`dashboards/grafana/docker-compose.yaml` を編集：

```yaml
services:
  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      # 実際のログファイルパスに変更
      - /var/log/tfdrift:/var/log/tfdrift:ro
      - ./promtail-config.yaml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
```

#### 3. Grafana スタックを再起動

```bash
cd dashboards/grafana
docker-compose restart promtail
```

---

## 🎨 ダッシュボードの使い方

### 1. TFDrift-Falco Overview（概要ダッシュボード）

**用途**: 全体像の把握

**主要パネル**:
- **Total Drift Events**: 期間内の総ドリフト数
- **Drift Events by Severity**: 深刻度別の内訳（円グラフ）
- **Drift Events by Resource Type**: リソース種別の内訳
- **Timeline**: 時系列でのドリフト発生状況
- **Recent Drift Events**: 最新のドリフトイベント一覧

**使い方**:
1. 右上の時間範囲を選択（Last 6 hours、Last 24 hours など）
2. Auto-refresh を有効化（5s、30s、1m など）
3. パネルをクリックして詳細を確認

### 2. TFDrift-Falco Diff Details（差分詳細）

**用途**: 設定変更の詳細確認

**主要パネル**:
- **Expected vs Actual**: 期待値と実際の値を比較
- **Changes by Actor**: 誰が変更したか
- **Top 10 Resources**: 最もドリフトが多いリソース

**使い方**:
1. 特定のリソース ID でフィルタ
2. Diff 内容を JSON ビューアで確認
3. Changed By でユーザーを特定

### 3. TFDrift-Falco Heatmap & Analytics（分析）

**用途**: パターン分析、トレンド把握

**主要パネル**:
- **Drift Frequency Heatmap**: 時間帯別のドリフト発生頻度
- **Activity by Resource Type**: リソース種別の活動状況
- **Hourly Drift Trends**: 時間帯別のトレンド

**使い方**:
1. Heatmap で異常な時間帯を特定
2. Bar chart でリソース種別の傾向を確認
3. 定期的なパターンを見つけて予防措置を検討

---

## 🚨 アラート設定（オプション）

アラートを設定することで、重要なドリフトをリアルタイムで通知できます。

### Step 1: Slack Webhook の設定（推奨）

1. **Slack Webhook を作成**
   - https://api.slack.com/apps にアクセス
   - 新しいアプリを作成
   - Incoming Webhooks を有効化
   - Webhook URL をコピー（例: `https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXX`）

2. **環境変数を設定**

`.env` ファイルを作成：

```bash
cd dashboards/grafana
cat > .env << 'EOF'
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
EOF
```

3. **Grafana を再起動**

```bash
docker-compose down
docker-compose up -d
```

### Step 2: Grafana でアラートを作成

詳細は [ALERTS.md](./ALERTS.md) を参照してください。

**クイックスタート**:

1. Grafana にログイン → **Alerting**（ベルアイコン）
2. **Alert rules** → **+ New alert rule**
3. 以下の設定でアラートを作成：

```
Name: Critical Drift Detected
Query: count_over_time({job="tfdrift-falco"} | json | severity="critical" [5m])
Threshold: > 1
For: 1 minute
Contact point: slack-tfdrift (事前に作成)
```

4. **Save rule and exit**

### Step 3: アラートをテスト

```bash
# テストイベントを生成
cat >> dashboards/grafana/sample-logs/current-drift-events.jsonl << 'EOF'
{"timestamp":"$(date -u +%Y-%m-%dT%H:%M:%SZ)","resource_type":"aws_security_group","resource_id":"sg-test-alert","changed_by":"test-user","severity":"critical","diff":{"ingress":{"expected":["443/tcp"],"actual":["443/tcp","22/tcp"]}},"action":"drift_detected"}
EOF

# 1-2分待つとアラートが発火して Slack に通知が届く
```

---

## 📊 実際の使用例

### ユースケース 1: 日次のセキュリティレビュー

**目的**: 毎朝、前日のドリフトを確認

**手順**:
1. **TFDrift-Falco Overview** を開く
2. 時間範囲を **Last 24 hours** に設定
3. **Drift Events by Severity** で critical/high を確認
4. **Recent Drift Events** で詳細を確認
5. 必要に応じて AWS Console で修正

### ユースケース 2: インシデント調査

**目的**: 特定のリソースでインシデント発生、変更履歴を調査

**手順**:
1. **TFDrift-Falco Diff Details** を開く
2. 検索バーでリソース ID を検索（例: `sg-123456`）
3. **Expected vs Actual** で何が変更されたか確認
4. **Changes by Actor** で誰が変更したか特定
5. Timestamp から変更時刻を確認

### ユースケース 3: コンプライアンス監査

**目的**: 月次監査のためのレポート作成

**手順**:
1. **TFDrift-Falco Overview** を開く
2. 時間範囲を **Last 30 days** に設定
3. **Dashboard** → **Share** → **Export** → **PDF**
4. レポートをダウンロード
5. 監査資料として提出

### ユースケース 4: トレンド分析

**目的**: ドリフトが増加している原因を特定

**手順**:
1. **TFDrift-Falco Heatmap & Analytics** を開く
2. **Drift Frequency Heatmap** で時間帯パターンを確認
3. **Activity by Resource Type** で問題のリソース種別を特定
4. 特定のリソース種別に絞り込んで調査
5. 根本原因を特定して対策を実施

---

## 🔧 トラブルシューティング

### ダッシュボードにデータが表示されない

**原因 1: TFDrift-Falco が動作していない**

```bash
# TFDrift-Falco のステータス確認
docker-compose ps tfdrift

# ログを確認
docker-compose logs tfdrift

# 起動していなければ起動
docker-compose up -d tfdrift
```

**原因 2: ログファイルが生成されていない**

```bash
# ログファイルの存在確認
docker-compose exec tfdrift ls -la /var/log/tfdrift/

# ファイルが無い場合は config.yaml を確認
docker-compose exec tfdrift cat /config/config.yaml
```

**原因 3: Promtail がログを収集していない**

```bash
# Promtail のログを確認
docker-compose logs promtail | grep -i error

# Promtail の設定を確認
docker-compose exec promtail cat /etc/promtail/config.yml
```

### Grafana にログインできない

```bash
# Grafana のステータス確認
docker-compose ps grafana

# ログを確認
docker-compose logs grafana

# 再起動
docker-compose restart grafana

# デフォルト認証情報
# Username: admin
# Password: admin
```

### アラートが発火しない

1. **データが Loki に届いているか確認**

```bash
curl -s "http://localhost:3100/loki/api/v1/labels" | jq
# "job": "tfdrift-falco" が表示されるはず
```

2. **アラートルールが正しく作成されているか確認**

Grafana → Alerting → Alert rules で確認

3. **クエリをテスト**

Grafana → Explore でクエリを実行してデータが返るか確認

### パフォーマンスが遅い

```bash
# リソース使用量を確認
docker stats grafana-grafana-1 grafana-loki-1 grafana-promtail-1

# メモリが不足している場合は docker-compose.yaml に追加
services:
  grafana:
    mem_limit: 512m
  loki:
    mem_limit: 1g
```

---

## 📚 次のステップ

### 基本をマスターしたら

1. **[カスタマイズガイド](./CUSTOMIZATION_GUIDE.md)** を読んで独自のパネルを作成
2. **[アラート設定ガイド](./ALERTS.md)** で通知を設定
3. **ダッシュボード変数** を使って柔軟なフィルタリングを実装
4. **チーム専用ダッシュボード** を作成

### さらに詳しく知りたい場合

- [Integration Test Results](./INTEGRATION_TEST_RESULTS.md) - テスト結果と技術詳細
- [Customization Guide](./CUSTOMIZATION_GUIDE.md) - カスタマイズ方法
- [Alert Configuration Guide](./ALERTS.md) - アラート設定の詳細
- [Grafana公式ドキュメント](https://grafana.com/docs/grafana/latest/)
- [Loki公式ドキュメント](https://grafana.com/docs/loki/latest/)

---

## 💡 よくある質問（FAQ）

### Q1: サンプルデータを削除して実データだけを表示したい

```bash
# サンプルデータを削除
rm dashboards/grafana/sample-logs/*.jsonl

# Promtail を再起動
docker-compose restart promtail
```

### Q2: 複数の環境（prod、staging、dev）を同じ Grafana で監視したい

`docker-compose.yaml` で環境ラベルを追加：

```yaml
services:
  promtail:
    volumes:
      - ./sample-logs:/var/log/tfdrift
      - ./promtail-config.yaml:/etc/promtail/config.yml
    environment:
      - ENVIRONMENT=production
```

クエリで環境を指定：
```logql
{job="tfdrift-falco", environment="production"} | json
```

### Q3: ダッシュボードを他のチームと共有したい

**方法 1: JSON エクスポート**

1. Dashboard → Share → Export
2. JSON ファイルを保存
3. 相手に送信
4. 相手側で Dashboards → Import → Upload JSON

**方法 2: Dashboard スナップショット**

1. Dashboard → Share → Snapshot
2. Expire: Never
3. Publish snapshot
4. リンクを共有

### Q4: データ保持期間を設定したい

Loki の設定ファイルを作成（30日間保持の例）：

```yaml
# dashboards/grafana/loki-config.yaml
schema_config:
  configs:
    - from: 2020-10-24
      store: boltdb-shipper
      object_store: filesystem
      schema: v11
      index:
        prefix: index_
        period: 24h

limits_config:
  retention_period: 720h  # 30 days
```

`docker-compose.yaml` で設定をマウント：

```yaml
loki:
  image: grafana/loki:2.9.0
  volumes:
    - ./loki-config.yaml:/etc/loki/local-config.yaml
  command: -config.file=/etc/loki/local-config.yaml
```

### Q5: 本番環境で使うためのセキュリティ設定は？

1. **パスワードを変更**
   ```yaml
   grafana:
     environment:
       - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_PASSWORD}
   ```

2. **SSL/TLS を有効化**（Nginx リバースプロキシ経由）

3. **認証を設定**（OAuth、LDAP など）
   ```yaml
   grafana:
     environment:
       - GF_AUTH_GOOGLE_ENABLED=true
       - GF_AUTH_GOOGLE_CLIENT_ID=${GOOGLE_CLIENT_ID}
       - GF_AUTH_GOOGLE_CLIENT_SECRET=${GOOGLE_CLIENT_SECRET}
   ```

4. **永続化ボリュームを設定**
   ```yaml
   volumes:
     - grafana-data:/var/lib/grafana
     - loki-data:/loki
   ```

---

## 🆘 サポート

問題が解決しない場合：

1. **ログを確認**
   ```bash
   docker-compose logs grafana
   docker-compose logs loki
   docker-compose logs promtail
   docker-compose logs tfdrift
   ```

2. **Issue を作成**
   - GitHub: https://github.com/your-org/tfdrift-falco/issues
   - 以下の情報を含めてください：
     - エラーメッセージ
     - `docker-compose logs` の出力
     - 環境情報（OS、Docker バージョン）

3. **コミュニティに質問**
   - Slack: #tfdrift-falco
   - Email: support@your-org.com

---

**最終更新**: 2025-12-05
**バージョン**: 1.0.0
