# TFDrift-Falco Dashboard Design

> **実装状況（2026-07）**: 出荷している UI は **組み込みの React + Cytoscape UI**
> （グラフトポロジ＋ドリフトフィード、下記「オプションA」）です。本ドキュメントの
> **Grafana / Prometheus / Kibana（オプションB・C）は未実装の設計代替案**であり、
> 現行機能ではありません。機能一覧やホームページで Grafana を「機能」として挙げない
> こと（主張≠実態の回避、#366）。

## 現在の状態

### 出力方式
現在、TFDrift-Falcoは以下の形式で出力しています:

1. **Console出力** (ANSI色付き)
   - ターミナルでの視覚的な表示
   - Severity別の色分け (CRITICAL=赤, HIGH=黄, etc)

2. **Slack/Discord通知**
   - Markdown形式
   - リアルタイムアラート

3. **JSON出力**
   - SIEM統合用
   - ログ収集システム連携

### 制限事項
- ✅ リアルタイムアラートは機能
- ❌ 履歴の可視化なし
- ❌ ダッシュボードなし
- ❌ トレンド分析なし
- ❌ フィルタリング/検索機能なし

## ダッシュボード構想

### アーキテクチャオプション

#### オプションA: 組み込みWebダッシュボード
```
TFDrift-Falco
  ├── Detector (既存)
  └── Web Server (新規)
       ├── REST API
       ├── WebSocket (リアルタイム更新)
       └── Static Files (React/Vue frontend)
```

**技術スタック:**
- Backend: Go標準 `net/http` または Gin/Echo
- Frontend: React/Vue.js
- Storage: SQLite (軽量) または PostgreSQL
- Realtime: WebSocket

**メリット:**
- オールインワンデプロイ
- 追加インフラ不要
- シンプルな構成

**デメリット:**
- TFDrift-Falcoに多くの機能追加
- 単一障害点

#### オプションB: 外部ダッシュボード統合
```
TFDrift-Falco → Metrics Export → Grafana/Kibana
                  ↓
                Prometheus/Elasticsearch
```

**技術スタック:**
- Metrics: Prometheus exporter
- Logs: Elasticsearch/Loki
- Dashboard: Grafana/Kibana

**メリット:**
- 既存のエコシステム活用
- スケーラブル
- 多機能

**デメリット:**
- 複雑な構成
- 追加インフラが必要

#### オプションC: ハイブリッド (推奨)
```
TFDrift-Falco
  ├── Detector
  ├── Simple Web UI (内蔵)
  └── Metrics/Logs Export
       ├── Prometheus metrics
       ├── JSON logs
       └── Grafana integration
```

**両方のいいとこ取り:**
- 簡易的な内蔵ダッシュボード（設定確認、最新アラート）
- 詳細分析は Grafana/Kibana

## UIデザインコンセプト

### トップページ
```
┌─────────────────────────────────────────────────────────────┐
│  🛰️ TFDrift-Falco Dashboard                    [Settings] │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  📊 Statistics (Last 24h)                                    │
│  ┌────────────┬────────────┬────────────┬────────────┐      │
│  │  CRITICAL  │    HIGH    │   MEDIUM   │    LOW     │      │
│  │     12     │     45     │     23     │     8      │      │
│  └────────────┴────────────┴────────────┴────────────┘      │
│                                                               │
│  🔴 Recent Alerts                                [Filter ▼]  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ 🚨 CRITICAL │ 2 min ago                              │  │
│  │ IAM Role Trust Policy Modified                        │  │
│  │ aws_iam_role.lambda_execution_role                    │  │
│  │ User: john.doe@company.com                            │  │
│  │                                      [Details] [Ack]  │  │
│  ├───────────────────────────────────────────────────────┤  │
│  │ ⚠️  HIGH    │ 15 min ago                             │  │
│  │ EC2 Instance Type Changed                             │  │
│  │ aws_instance.web_server                               │  │
│  │ t2.micro → t3.large                                   │  │
│  │                                      [Details] [Ack]  │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  📈 Drift Trends                                             │
│  [7d] [30d] [90d]                                            │
│  ┌───────────────────────────────────────────────────────┐  │
│  │        ██                                              │  │
│  │     ███████                                            │  │
│  │  ████████████                                          │  │
│  └───────────────────────────────────────────────────────┘  │
│   Mon  Tue  Wed  Thu  Fri  Sat  Sun                         │
│                                                               │
│  🎯 Top Drifted Resources                                    │
│  1. aws_iam_role.*              45 changes                   │
│  2. aws_instance.*              23 changes                   │
│  3. aws_s3_bucket.*             18 changes                   │
└─────────────────────────────────────────────────────────────┘
```

### アラート詳細ページ
```
┌─────────────────────────────────────────────────────────────┐
│  ← Back to Dashboard                                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  🚨 DRIFT ALERT DETAILS                                      │
│                                                               │
│  Severity: CRITICAL                                          │
│  Timestamp: 2025-01-15 14:23:45 JST                          │
│  Status: [Open ▼] [Mark as Resolved] [Create Ticket]        │
│                                                               │
│  📦 Resource                                                  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Type:     aws_iam_role                                 │  │
│  │ Name:     lambda_execution_role                        │  │
│  │ ID:       arn:aws:iam::123456789:role/lambda-exec      │  │
│  │ Region:   us-east-1                                    │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  👤 User Identity                                             │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ User:        john.doe@company.com                      │  │
│  │ Principal:   arn:aws:iam::123456789:user/john.doe      │  │
│  │ IP Address:  203.0.113.45                              │  │
│  │ User Agent:  AWS Console                               │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  🔄 Changes                                                   │
│  ┌───────────────────────────────────────────────────────┐  │
│  │ Attribute: assume_role_policy                          │  │
│  │                                                         │  │
│  │ Before (Terraform State):                              │  │
│  │ {                                                       │  │
│  │   "Version": "2012-10-17",                             │  │
│  │   "Statement": [{                                       │  │
│  │     "Effect": "Allow",                                  │  │
│  │     "Principal": {                                      │  │
│  │ -     "Service": "lambda.amazonaws.com"                 │  │
│  │     },                                                  │  │
│  │     "Action": "sts:AssumeRole"                          │  │
│  │   }]                                                    │  │
│  │ }                                                       │  │
│  │                                                         │  │
│  │ After (Current State):                                 │  │
│  │ {                                                       │  │
│  │   "Version": "2012-10-17",                             │  │
│  │   "Statement": [{                                       │  │
│  │     "Effect": "Allow",                                  │  │
│  │     "Principal": {                                      │  │
│  │ +     "Service": [                                      │  │
│  │ +       "lambda.amazonaws.com",                         │  │
│  │ +       "ec2.amazonaws.com"                             │  │
│  │ +     ]                                                 │  │
│  │     },                                                  │  │
│  │     "Action": "sts:AssumeRole"                          │  │
│  │   }]                                                    │  │
│  │ }                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
│  📋 Matched Falco Rules                                      │
│  • IAM Role Trust Policy Modified                            │
│                                                               │
│  💬 Comments                                                  │
│  [Add comment...]                                            │
│                                                               │
│  📊 Actions                                                   │
│  [Revert via Terraform] [Ignore] [Create JIRA Ticket]       │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

### サイドバーナビゲーション
```
┌────────────────┐
│ 🛰️ TFDrift     │
├────────────────┤
│ 📊 Dashboard   │
│ 🔴 Alerts      │
│ 📈 Analytics   │
│ 🎯 Resources   │
│ 🦅 Falco Rules │
│ ⚙️  Settings   │
│ 📖 Docs        │
└────────────────┘
```

## 機能要件

### Phase 1: MVP (Minimum Viable Product)
- [ ] リアルタイムアラートリスト
- [ ] アラート詳細表示
- [ ] Severity別フィルタリング
- [ ] 最近24時間の統計
- [ ] WebSocket でリアルタイム更新

### Phase 2: 分析機能
- [ ] 時系列グラフ (7日/30日/90日)
- [ ] リソース別ドリフト集計
- [ ] ユーザー別アクティビティ
- [ ] エクスポート機能 (CSV/JSON)

### Phase 3: インタラクティブ機能
- [ ] アラートステータス管理 (Open/Acknowledged/Resolved)
- [ ] コメント機能
- [ ] Terraform リバート提案
- [ ] JIRA/GitHub Issues 連携

### Phase 4: 高度な機能
- [ ] ドリフトパターン検出 (ML)
- [ ] 異常検知
- [ ] コンプライアンスレポート
- [ ] 変更承認ワークフロー

## 技術実装案

### Backend API設計

```go
// pkg/dashboard/server.go
package dashboard

type Server struct {
    detector *detector.Detector
    storage  Storage
    hub      *websocket.Hub
}

// API Endpoints
GET  /api/alerts              // アラート一覧
GET  /api/alerts/:id          // アラート詳細
POST /api/alerts/:id/ack      // アラート承認
GET  /api/stats               // 統計情報
GET  /api/resources           // リソース一覧
GET  /ws                      // WebSocket接続
```

### データストレージ

```go
// pkg/storage/sqlite.go
type Alert struct {
    ID           string    `json:"id"`
    Timestamp    time.Time `json:"timestamp"`
    Severity     string    `json:"severity"`
    ResourceType string    `json:"resource_type"`
    ResourceName string    `json:"resource_name"`
    ResourceID   string    `json:"resource_id"`
    Attribute    string    `json:"attribute"`
    OldValue     string    `json:"old_value"`
    NewValue     string    `json:"new_value"`
    UserIdentity string    `json:"user_identity"`
    Status       string    `json:"status"` // open, acked, resolved
    Comments     []Comment `json:"comments"`
}
```

### WebSocket リアルタイム更新

```javascript
// frontend/src/websocket.js
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onmessage = (event) => {
  const alert = JSON.parse(event.data);
  // Update UI with new alert
  addAlertToList(alert);
  showNotification(alert);
};
```

## UI/UXデザイン参考

### カラースキーム
```
CRITICAL: #FF4444 (赤)
HIGH:     #FFA500 (オレンジ)
MEDIUM:   #FFD700 (黄)
LOW:      #90EE90 (緑)

Background: #1E1E1E (ダークテーマ)
Text:       #E0E0E0
Accent:     #4A90E2 (青)
```

### フォント
- Primary: Inter, SF Pro, Segoe UI
- Monospace: Fira Code, JetBrains Mono (コードブロック用)

### アイコン
- Material Design Icons または Lucide Icons

## Grafana統合サンプル

```yaml
# grafana-dashboard.json
{
  "dashboard": {
    "title": "TFDrift-Falco Monitoring",
    "panels": [
      {
        "title": "Drift Events by Severity",
        "type": "piechart",
        "targets": [
          {
            "expr": "sum by(severity) (tfdrift_alerts_total)"
          }
        ]
      },
      {
        "title": "Drifts Over Time",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(tfdrift_alerts_total[5m])"
          }
        ]
      },
      {
        "title": "Top Drifted Resources",
        "type": "table",
        "targets": [
          {
            "expr": "topk(10, sum by(resource_type) (tfdrift_alerts_total))"
          }
        ]
      }
    ]
  }
}
```

## 実装優先順位

### 短期 (1-2週間)
1. ✅ 基本的なメトリクスエクスポート (Prometheus)
2. ✅ JSON logs output (既存機能の拡張)
3. ⬜ 簡易的なWebダッシュボード (読み取り専用)

### 中期 (1-2ヶ月)
1. ⬜ SQLite永続化
2. ⬜ WebSocketリアルタイム更新
3. ⬜ フィルタリング/検索機能

### 長期 (3-6ヶ月)
1. ⬜ ステータス管理機能
2. ⬜ 外部連携 (JIRA, Slack詳細)
3. ⬜ MLベースの異常検知

## デモ用モックデータ

開発時に使用するサンプルアラート:

```json
[
  {
    "id": "alert-001",
    "timestamp": "2025-01-15T14:23:45Z",
    "severity": "critical",
    "resource_type": "aws_iam_role",
    "resource_name": "lambda_execution_role",
    "attribute": "assume_role_policy",
    "user_identity": "john.doe@company.com",
    "status": "open"
  },
  {
    "id": "alert-002",
    "timestamp": "2025-01-15T14:08:12Z",
    "severity": "high",
    "resource_type": "aws_instance",
    "resource_name": "web_server",
    "attribute": "instance_type",
    "old_value": "t2.micro",
    "new_value": "t3.large",
    "user_identity": "admin@company.com",
    "status": "acknowledged"
  }
]
```

## 参考プロジェクト

- **Falco UI**: https://github.com/falcosecurity/falcosidekick-ui
- **Prometheus Alertmanager**: シンプルなアラート管理UI
- **Grafana**: 高機能ダッシュボード
- **Sentry**: エラートラッキングのUX参考
- **Datadog**: セキュリティアラートのUI参考

## まとめ

ダッシュボードは **Phase 1 (MVP)** から段階的に実装するのが現実的です。

**推奨アプローチ:**
1. まず Prometheus metrics export 追加 → Grafana連携
2. 次に簡易的な内蔵Webダッシュボード (読み取り専用)
3. 最後にインタラクティブ機能追加

これにより、機能を段階的に追加しながら、早期にダッシュボードの価値を提供できます。
