---
title: driftwireを「使いやすく」「つながる」ツールに進化させた話 - v0.3.1→v0.4.1の実装
tags: Terraform, Go, Falco, DevOps, AWS
author: Keita Higaki
slide: false
---

# driftwireを「使いやすく」「つながる」ツールに進化させた話 - v0.3.1→v0.4.1の実装

## はじめに

オープンソースプロジェクト「driftwire」を、**"お手軽に使える"** × **"他システムとつながる"** ツールに進化させた取り組みを紹介します。

3週間で以下の機能を実装しました：

- **v0.3.1**: L1 Semi-Auto Mode（選択的カスタマイズ）
- **v0.4.0**: 構造化イベント出力（NDJSON対応）
- **v0.4.1**: Webhook統合（Slack/Teams対応）

この記事では、設計思想から実装の詳細まで、実際のコードを交えて解説します。

## TL;DR

- 🎯 **設計思想**: "考えなくていいけど、逃げ道はある"
- 🚀 **v0.3.1**: `--auto --region us-west-2` で部分カスタマイズ
- 📊 **v0.4.0**: NDJSON形式でSIEM/SOAR連携可能に
- 🔔 **v0.4.1**: Slack/Teamsへの自動通知（リトライ付き）
- ✅ **テスト**: 全バージョンで包括的なテストケース追加

## driftwireとは

**driftwire**は、Falcoを使ってTerraformのドリフト（手動変更による設定差異）をリアルタイムで検出するツールです。

```bash
# 従来のTerraform drift検出
terraform plan  # 定期的に実行する必要がある

# driftwireの場合
driftwire --config config.yaml  # リアルタイムで検出
```

CloudTrailイベントをFalcoで監視し、Terraform stateと比較することで、**誰が・いつ・何を変更したか**を即座に検知します。

## 課題：「使いづらい」「つながらない」

v0.3.0リリース後、2つの大きな課題が見えてきました。

### 課題1: 設定ファイルが複雑

```yaml
# config.yaml（50行以上の設定が必要）
providers:
  aws:
    enabled: true
    regions:
      - us-east-1
    state:
      backend: s3
      bucket: my-terraform-state
      key: terraform.tfstate
      region: us-east-1

falco:
  hostname: localhost
  port: 5060

notifications:
  slack:
    enabled: true
    webhook_url: https://hooks.slack.com/...
# ... さらに続く
```

**問題点**:
- 初回セットアップのハードルが高い
- ちょっと試したいだけなのに大変
- リージョンだけ変えたい時も全部書く必要がある

### 課題2: 他システムとの連携が困難

```
driftwire (v0.3.0)
    ↓
  ログ出力（人間向けテキスト）
    ↓
  ？？？ ← SIEM/SOARに送りたいけど...
```

**問題点**:
- ログは人間向けで構造化されていない
- Slack/Teamsに通知したいけど自分で実装が必要
- SIEM/SOARへの連携にパーサーが必要

## 解決策：「お手軽」×「つながる」

設計の核となる思想：

> **"考えなくていいけど、逃げ道はある"**

### アプローチ1: 3段階の設定レベル

```
L0（Zero-Config）：考えなくていい
    ↓
L1（Semi-Auto）：一部だけカスタマイズ
    ↓
L2（Full-Config）：完全にコントロール
```

### アプローチ2: イベント駆動アーキテクチャ

```
driftwire
    ↓
構造化イベント（JSON）
    ↓ ↓ ↓
  Slack Teams SIEM/SOAR
```

## v0.3.1: L1 Semi-Auto Mode

**目標**: ゼロコンフィグを保ちつつ、必要な部分だけカスタマイズできるように。

### 実装した機能

```bash
# L0: ゼロコンフィグ
driftwire --auto

# L1: リージョンだけ変更
driftwire --auto --region us-west-2,ap-northeast-1

# L1: Falcoエンドポイントを指定
driftwire --auto --falco-endpoint prod-falco:5061

# L1: ローカルstateファイルを指定
driftwire --auto --state-path ./terraform.tfstate

# L1: バックエンドタイプを強制
driftwire --auto --backend local

# L1: 複数のオプションを組み合わせ
driftwire --auto \
  --region us-west-2 \
  --falco-endpoint localhost:5060 \
  --state-path ./terraform.tfstate
```

### コア実装：applyConfigOverrides()

**cmd/driftwire/main.go**:

```go
// applyConfigOverrides applies L1 semi-auto mode flag overrides to the config
func applyConfigOverrides(cfg *config.Config) error {
    // Override AWS regions if specified
    if len(regionOverride) > 0 {
        cfg.Providers.AWS.Regions = regionOverride
        log.Infof("✓ Using custom region(s): %v", regionOverride)
    }

    // Override Falco endpoint if specified
    if falcoEndpoint != "" {
        parts := strings.Split(falcoEndpoint, ":")
        if len(parts) != 2 {
            return fmt.Errorf("invalid Falco endpoint format (expected host:port): %s", falcoEndpoint)
        }

        port, err := strconv.Atoi(parts[1])
        if err != nil {
            return fmt.Errorf("invalid port in Falco endpoint: %s", parts[1])
        }

        cfg.Falco.Hostname = parts[0]
        cfg.Falco.Port = uint16(port)
        log.Infof("✓ Using custom Falco endpoint: %s", falcoEndpoint)
    }

    // Override state path if specified
    if statePathOverride != "" {
        cfg.Providers.AWS.State.LocalPath = statePathOverride
        cfg.Providers.AWS.State.Backend = "local"
        log.Infof("✓ Using custom state path: %s", statePathOverride)
    }

    // Override backend type if specified
    if backendTypeOverride != "" {
        if backendTypeOverride != "local" && backendTypeOverride != "s3" {
            return fmt.Errorf("invalid backend type (must be 'local' or 's3'): %s", backendTypeOverride)
        }
        cfg.Providers.AWS.State.Backend = backendTypeOverride
        log.Infof("✓ Using backend type: %s", backendTypeOverride)
    }

    return nil
}
```

### 実行フロー

```
1. --autoフラグ検出
    ↓
2. Terraform stateを自動検出
    ↓
3. デフォルト設定を生成
    ↓
4. applyConfigOverrides()で上書き  ← NEW!
    ↓
5. 検出開始
```

### テスト戦略

**cmd/driftwire/main_test.go**に7つのテストケースを追加：

```go
func TestApplyConfigOverrides_RegionOverride(t *testing.T) {
    // Set region override
    regionOverride = []string{"us-west-2", "ap-northeast-1"}
    defer func() { regionOverride = nil }()

    cfg := &config.Config{
        Providers: config.ProvidersConfig{
            AWS: config.AWSConfig{
                Regions: []string{"us-east-1"},
            },
        },
        Falco: config.FalcoConfig{
            Hostname: "localhost",
            Port:     5060,
        },
    }

    err := applyConfigOverrides(cfg)
    require.NoError(t, err)

    assert.Equal(t, []string{"us-west-2", "ap-northeast-1"}, cfg.Providers.AWS.Regions)
}

func TestApplyConfigOverrides_FalcoEndpoint(t *testing.T) {
    // Set Falco endpoint override
    falcoEndpoint = "prod-falco:5061"
    defer func() { falcoEndpoint = "" }()

    cfg := &config.Config{
        Providers: config.ProvidersConfig{
            AWS: config.AWSConfig{
                Regions: []string{"us-east-1"},
            },
        },
        Falco: config.FalcoConfig{
            Hostname: "localhost",
            Port:     5060,
        },
    }

    err := applyConfigOverrides(cfg)
    require.NoError(t, err)

    assert.Equal(t, "prod-falco", cfg.Falco.Hostname)
    assert.Equal(t, uint16(5061), cfg.Falco.Port)
}
```

**成果**: テストカバレッジが32.3% → 50.8%に向上 ✅

## v0.4.0: Structured Event Output

**目標**: ツールから「プラットフォーム」へ。他システムと連携可能な構造化イベント出力を実装。

### イベントモデルの設計

**pkg/types/drift_event.go**:

```go
// DriftEvent represents a structured drift detection event
type DriftEvent struct {
    EventType    string    `json:"event_type"`    // "terraform_drift"
    Provider     string    `json:"provider"`      // "aws"
    ResourceType string    `json:"resource_type"` // "aws_security_group"
    ResourceID   string    `json:"resource_id"`   // "sg-12345"
    ChangeType   string    `json:"change_type"`   // "created", "modified", "deleted"
    DetectedAt   time.Time `json:"detected_at"`   // RFC3339 timestamp
    Source       string    `json:"source"`        // "falco"
    Severity     string    `json:"severity"`      // "critical", "high", "medium", "low"

    // Optional fields
    Region          string `json:"region,omitempty"`
    User            string `json:"user,omitempty"`
    CloudTrailEvent string `json:"cloudtrail_event,omitempty"`
    RequestID       string `json:"request_id,omitempty"`

    // Metadata
    Version string `json:"version"` // Schema version: "1.0.0"
}
```

**設計のポイント**:
- 不変なスキーマバージョン（`version: "1.0.0"`）
- オプショナルフィールドは`omitempty`
- RFC3339形式のタイムスタンプ
- 自動的な重要度判定

### Builder Pattern実装

```go
// NewDriftEvent creates a new drift event
func NewDriftEvent(provider, resourceType, resourceID, changeType string) *DriftEvent {
    event := &DriftEvent{
        EventType:    "terraform_drift",
        Provider:     provider,
        ResourceType: resourceType,
        ResourceID:   resourceID,
        ChangeType:   changeType,
        DetectedAt:   time.Now(),
        Source:       "falco",
        Version:      "1.0.0",
    }

    // Auto-determine severity
    event.Severity = DetermineSeverity(resourceType, changeType)

    return event
}

// Chainable builder methods
func (e *DriftEvent) WithSeverity(severity string) *DriftEvent {
    e.Severity = severity
    return e
}

func (e *DriftEvent) WithRegion(region string) *DriftEvent {
    e.Region = region
    return e
}

func (e *DriftEvent) WithUser(user string) *DriftEvent {
    e.User = user
    return e
}

func (e *DriftEvent) WithCloudTrailEvent(eventName, requestID string) *DriftEvent {
    e.CloudTrailEvent = eventName
    e.RequestID = requestID
    return e
}
```

### 自動重要度判定

```go
// DetermineSeverity automatically determines event severity
func DetermineSeverity(resourceType, changeType string) string {
    // Critical resources
    criticalResources := map[string]bool{
        "aws_iam_role":             true,
        "aws_iam_policy":           true,
        "aws_security_group":       true,
        "aws_security_group_rule":  true,
        "aws_kms_key":              true,
    }

    if criticalResources[resourceType] {
        return SeverityCritical
    }

    // Deletions are high severity
    if changeType == ChangeTypeDeleted {
        return SeverityHigh
    }

    // Default to medium
    return SeverityMedium
}
```

### NDJSON出力の実装

**pkg/output/json.go**:

```go
// JSONOutput writes drift events as newline-delimited JSON
type JSONOutput struct {
    writer io.Writer
    mu     sync.Mutex
}

func NewJSONOutput(writer io.Writer) *JSONOutput {
    return &JSONOutput{
        writer: writer,
    }
}

func (j *JSONOutput) Write(event *types.DriftEvent) error {
    j.mu.Lock()
    defer j.mu.Unlock()

    jsonStr, err := event.ToJSONString()
    if err != nil {
        return fmt.Errorf("failed to serialize event: %w", err)
    }

    _, err = fmt.Fprintln(j.writer, jsonStr)
    return err
}
```

**NDJSON形式の利点**:
- 1行1イベント = ストリーミング処理に最適
- jqで簡単にフィルタリング可能
- Fluent Bit/Fluentdで直接取り込み可能

### 出力モード管理

**pkg/output/manager.go**:

```go
type OutputMode string

const (
    OutputModeHuman OutputMode = "human" // 人間向け（stderr）
    OutputModeJSON  OutputMode = "json"  // JSON（stdout）
    OutputModeBoth  OutputMode = "both"  // 両方
)

type Manager struct {
    mode       OutputMode
    jsonOutput *JSONOutput
    humanOut   io.Writer
}

func (m *Manager) EmitDriftEvent(event *types.DriftEvent) error {
    // JSON to stdout
    if m.mode == OutputModeJSON || m.mode == OutputModeBoth {
        if err := m.jsonOutput.Write(event); err != nil {
            return err
        }
    }

    // Human to stderr
    if m.mode == OutputModeHuman || m.mode == OutputModeBoth {
        humanMsg := m.formatHumanMessage(event)
        fmt.Fprintln(m.humanOut, humanMsg)
    }

    return nil
}

func (m *Manager) formatHumanMessage(event *types.DriftEvent) string {
    emoji := getSeverityEmoji(event.Severity)

    msg := fmt.Sprintf("%s [%s] %s: %s (%s)",
        emoji,
        event.Severity,
        event.ChangeType,
        event.ResourceType,
        event.ResourceID,
    )

    if event.User != "" {
        msg += fmt.Sprintf(" by %s", event.User)
    }

    if event.Region != "" {
        msg += fmt.Sprintf(" in %s", event.Region)
    }

    return msg
}
```

### 使用例

```bash
# Human-readable output (default)
driftwire --auto
# Output to stderr:
# 🚨 [critical] modified: aws_security_group (sg-12345) by admin in us-west-2

# JSON output only
driftwire --auto --output json
# Output to stdout (NDJSON):
# {"event_type":"terraform_drift","provider":"aws","resource_type":"aws_security_group",...}

# Both outputs
driftwire --auto --output both
# stderr: 🚨 [critical] modified: aws_security_group (sg-12345)
# stdout: {"event_type":"terraform_drift",...}

# Pipeline to jq
driftwire --auto --output json | jq 'select(.severity == "critical")'

# Pipeline to Fluent Bit
driftwire --auto --output json | fluent-bit -c fluent-bit.conf
```

### テスト戦略

**pkg/types/drift_event_test.go**（25+テストケース）:

```go
func TestNewDriftEvent(t *testing.T) {
    event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified)

    assert.Equal(t, "terraform_drift", event.EventType)
    assert.Equal(t, "aws", event.Provider)
    assert.Equal(t, "aws_security_group", event.ResourceType)
    assert.Equal(t, "sg-12345", event.ResourceID)
    assert.Equal(t, types.ChangeTypeModified, event.ChangeType)
    assert.Equal(t, "falco", event.Source)
    assert.Equal(t, "1.0.0", event.Version)
    assert.Equal(t, types.SeverityCritical, event.Severity) // Auto-determined
}

func TestDriftEvent_BuilderPattern(t *testing.T) {
    event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated).
        WithRegion("us-west-2").
        WithUser("admin@example.com").
        WithCloudTrailEvent("RunInstances", "req-123").
        WithSeverity(types.SeverityHigh)

    assert.Equal(t, "us-west-2", event.Region)
    assert.Equal(t, "admin@example.com", event.User)
    assert.Equal(t, "RunInstances", event.CloudTrailEvent)
    assert.Equal(t, "req-123", event.RequestID)
    assert.Equal(t, types.SeverityHigh, event.Severity)
}

func TestDriftEvent_ToJSON(t *testing.T) {
    event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeModified)

    jsonBytes, err := event.ToJSON()
    require.NoError(t, err)

    // Unmarshal and verify
    var decoded types.DriftEvent
    err = json.Unmarshal(jsonBytes, &decoded)
    require.NoError(t, err)

    assert.Equal(t, event.ResourceType, decoded.ResourceType)
    assert.Equal(t, event.Version, decoded.Version)
}
```

**成果**: pkg/types で100%カバレッジ達成 ✅

## v0.4.1: Webhook Integration

**目標**: Slack/Teamsへの即座通知と、カスタムWebhook統合。

### Webhook設定

**pkg/output/webhook.go**:

```go
// WebhookConfig contains webhook configuration
type WebhookConfig struct {
    URL         string            `yaml:"url" json:"url"`
    Method      string            `yaml:"method" json:"method"`           // POST, PUT (default: POST)
    Headers     map[string]string `yaml:"headers" json:"headers"`         // Custom headers
    Timeout     time.Duration     `yaml:"timeout" json:"timeout"`         // Request timeout (default: 10s)
    MaxRetries  int               `yaml:"max_retries" json:"max_retries"` // Max retry attempts (default: 3)
    RetryDelay  time.Duration     `yaml:"retry_delay" json:"-"`           // Initial retry delay (default: 1s)
    ContentType string            `yaml:"content_type" json:"content_type"` // Content-Type header
}
```

### Exponential Backoff リトライ

```go
// sendWithRetry sends the request with exponential backoff retry
func (w *WebhookOutput) sendWithRetry(jsonData []byte) error {
    var lastErr error

    for attempt := 0; attempt <= w.config.MaxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff: delay * 2^(attempt-1)
            // Attempt 1: 1s, 2: 2s, 3: 4s, 4: 8s
            delay := w.config.RetryDelay * time.Duration(1<<uint(attempt-1))
            log.Debugf("Webhook retry attempt %d/%d after %v", attempt, w.config.MaxRetries, delay)
            time.Sleep(delay)
        }

        err := w.send(jsonData)
        if err == nil {
            if attempt > 0 {
                log.Infof("Webhook succeeded after %d retries", attempt)
            }
            return nil
        }

        lastErr = err
        log.Warnf("Webhook attempt %d failed: %v", attempt+1, err)
    }

    return fmt.Errorf("webhook failed after %d attempts: %w", w.config.MaxRetries+1, lastErr)
}
```

### Slackフォーマッター

```go
// FormatSlackPayload formats a drift event as a Slack message
func FormatSlackPayload(event *types.DriftEvent) map[string]interface{} {
    severity := getSeverityColor(event.Severity)

    text := fmt.Sprintf("*Terraform Drift Detected*\n"+
        "Resource: `%s` (%s)\n"+
        "Change: %s\n"+
        "Severity: %s",
        event.ResourceType,
        event.ResourceID,
        event.ChangeType,
        event.Severity)

    if event.Region != "" {
        text += fmt.Sprintf("\nRegion: %s", event.Region)
    }
    if event.User != "" {
        text += fmt.Sprintf("\nUser: %s", event.User)
    }
    if event.CloudTrailEvent != "" {
        text += fmt.Sprintf("\nCloudTrail: %s", event.CloudTrailEvent)
    }

    return map[string]interface{}{
        "attachments": []map[string]interface{}{
            {
                "color":       severity,
                "text":        text,
                "footer":      "driftwire",
                "footer_icon": "https://falco.org/img/brand/falco-logo.png",
                "ts":          event.DetectedAt.Unix(),
            },
        },
    }
}

// getSeverityColor returns a color code for Slack attachments
func getSeverityColor(severity string) string {
    switch severity {
    case types.SeverityCritical:
        return "danger" // Red
    case types.SeverityHigh:
        return "warning" // Orange
    case types.SeverityMedium:
        return "#439FE0" // Blue
    case types.SeverityLow:
        return "good" // Green
    default:
        return "#808080" // Gray
    }
}
```

### Microsoft Teamsフォーマッター

```go
// FormatTeamsPayload formats a drift event as a Microsoft Teams message
func FormatTeamsPayload(event *types.DriftEvent) map[string]interface{} {
    title := fmt.Sprintf("Terraform Drift Detected: %s", event.ResourceType)

    text := fmt.Sprintf("**Resource ID**: %s\n\n"+
        "**Change Type**: %s\n\n"+
        "**Severity**: %s",
        event.ResourceID,
        event.ChangeType,
        event.Severity)

    if event.Region != "" {
        text += fmt.Sprintf("\n\n**Region**: %s", event.Region)
    }
    if event.User != "" {
        text += fmt.Sprintf("\n\n**User**: %s", event.User)
    }

    return map[string]interface{}{
        "@type":      "MessageCard",
        "@context":   "https://schema.org/extensions",
        "summary":    title,
        "title":      title,
        "text":       text,
        "themeColor": getTeamsColor(event.Severity),
    }
}

// getTeamsColor returns a color code for Microsoft Teams
func getTeamsColor(severity string) string {
    switch severity {
    case types.SeverityCritical:
        return "FF0000" // Red
    case types.SeverityHigh:
        return "FFA500" // Orange
    case types.SeverityMedium:
        return "0078D7" // Blue
    case types.SeverityLow:
        return "28A745" // Green
    default:
        return "808080" // Gray
    }
}
```

### 設定例

**config.yaml**:

```yaml
# Slack通知
notifications:
  webhooks:
    - url: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
      method: POST
      timeout: 10s
      max_retries: 3
      retry_delay: 1s

# Microsoft Teams通知
notifications:
  webhooks:
    - url: https://outlook.office.com/webhook/YOUR_TEAMS_WEBHOOK
      method: POST

# カスタムAPI
notifications:
  webhooks:
    - url: https://api.example.com/drift-events
      method: POST
      headers:
        Authorization: "Bearer YOUR_TOKEN"
        X-Custom-Header: "custom-value"
      timeout: 5s
      max_retries: 5
```

### 使用例

```bash
# Slackに通知
driftwire --config config-slack.yaml

# Microsoft Teamsに通知
driftwire --config config-teams.yaml

# カスタムWebhook + JSON出力
driftwire --config config-custom.yaml --output both
```

### テスト戦略

**pkg/output/webhook_test.go**（15テストケース）:

```go
func TestWebhookOutput_Write_Success(t *testing.T) {
    // Create test server
    var receivedEvent types.DriftEvent
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify request
        assert.Equal(t, "POST", r.Method)
        assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

        // Decode event
        err := json.NewDecoder(r.Body).Decode(&receivedEvent)
        require.NoError(t, err)

        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Create webhook
    config := WebhookConfig{
        URL: server.URL,
    }
    webhook := NewWebhookOutput(config)
    defer webhook.Close()

    // Send event
    event := types.NewDriftEvent("aws", "aws_security_group", "sg-12345", types.ChangeTypeModified)
    err := webhook.Write(event)
    require.NoError(t, err)

    // Verify received event
    assert.Equal(t, "aws_security_group", receivedEvent.ResourceType)
    assert.Equal(t, "sg-12345", receivedEvent.ResourceID)
}

func TestWebhookOutput_Write_Retry(t *testing.T) {
    attempts := 0

    // Create test server that fails first 2 attempts
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        attempts++
        if attempts < 3 {
            w.WriteHeader(http.StatusInternalServerError)
            return
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Create webhook with fast retries for testing
    config := WebhookConfig{
        URL:        server.URL,
        MaxRetries: 3,
        RetryDelay: 10 * time.Millisecond,
    }
    webhook := NewWebhookOutput(config)
    defer webhook.Close()

    // Send event
    event := types.NewDriftEvent("aws", "aws_db_instance", "db-12345", types.ChangeTypeDeleted)
    err := webhook.Write(event)
    require.NoError(t, err)

    // Should have retried twice before success
    assert.Equal(t, 3, attempts)
}

func TestWebhookOutput_Write_CustomHeaders(t *testing.T) {
    // Create test server
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Verify custom headers
        assert.Equal(t, "Bearer secret-token", r.Header.Get("Authorization"))
        assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))

        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    // Create webhook with custom headers
    config := WebhookConfig{
        URL: server.URL,
        Headers: map[string]string{
            "Authorization":   "Bearer secret-token",
            "X-Custom-Header": "custom-value",
        },
    }
    webhook := NewWebhookOutput(config)
    defer webhook.Close()

    // Send event
    event := types.NewDriftEvent("aws", "aws_instance", "i-12345", types.ChangeTypeCreated)
    err := webhook.Write(event)
    require.NoError(t, err)
}
```

**成果**: pkg/output で95%+カバレッジ達成 ✅

## アーキテクチャの進化

### Before（v0.3.0）: ツール

```
┌─────────────────┐
│  driftwire  │
└────────┬────────┘
         │
         ↓
   Human Logs
```

### After（v0.4.1）: プラットフォーム

```
┌─────────────────┐
│  driftwire  │
└────────┬────────┘
         │
         ↓
  DriftEvent (JSON)
         │
    ┌────┴────┐
    ↓         ↓
  Stdout    Stderr
    │         │
    ↓         ↓
  NDJSON    Human
    │
┌───┴────┐
│Webhook │
└───┬────┘
    │
  ┌─┴──┬──┬───┐
  ↓    ↓  ↓   ↓
Slack Teams Custom SIEM
```

## パフォーマンスと信頼性

### リトライロジックの動作

```
Attempt 0: Immediate
    ↓ ❌ (500 error)
Wait 1s
    ↓
Attempt 1: After 1s
    ↓ ❌ (timeout)
Wait 2s
    ↓
Attempt 2: After 2s
    ↓ ❌ (connection refused)
Wait 4s
    ↓
Attempt 3: After 4s
    ↓ ✅ Success!

Total time: ~7s
Total attempts: 4
```

### スレッドセーフティ

```go
// JSONOutput and WebhookOutput are thread-safe
type JSONOutput struct {
    writer io.Writer
    mu     sync.Mutex  // Protects concurrent writes
}

func (j *JSONOutput) Write(event *types.DriftEvent) error {
    j.mu.Lock()
    defer j.mu.Unlock()

    // Safe concurrent access
    jsonStr, err := event.ToJSONString()
    if err != nil {
        return fmt.Errorf("failed to serialize event: %w", err)
    }

    _, err = fmt.Fprintln(j.writer, jsonStr)
    return err
}
```

## 学んだこと

### 設計面

**✅ Do's**:
- **段階的な設定レベル（L0/L1/L2）**: ユーザーが選べる自由度
- **stdout/stderrの分離**: パイプライン処理を容易に
- **不変なスキーマバージョン**: 互換性を保証
- **Builder Pattern**: 柔軟なイベント構築

**❌ Don'ts**:
- **完璧主義を追わない**: まずMVPをリリース
- **複雑な設定を強要しない**: デフォルトで動くように
- **ログと構造化データを混ぜない**: 明確に分離

### 実装面

**✅ Do's**:
- **HTTPモックサーバー**: 外部依存のテストに最適
- **Table-Driven Tests**: 複数シナリオを効率的にカバー
- **Exponential Backoff**: ネットワークエラーに対する標準的手法
- **Concurrent-Safe設計**: sync.Mutexで保護

**❌ Don'ts**:
- **テストなしで実装しない**: TDD的アプローチが安全
- **エラーハンドリングを省略しない**: リトライやタイムアウトは必須
- **ハードコードを避ける**: 設定可能にする

## まとめ

3週間で実装した機能：

| バージョン | 機能 | 成果 |
|-----------|------|------|
| v0.3.1 | L1 Semi-Auto Mode | 選択的カスタマイズ対応 |
| v0.4.0 | Structured Events | NDJSON出力、SIEM連携 |
| v0.4.1 | Webhook Integration | Slack/Teams通知 |

**設計思想の実現**:
```
"考えなくていいけど、逃げ道はある"
    ↓
driftwire --auto                    ← L0: 考えない
driftwire --auto --region us-west-2 ← L1: 一部カスタマイズ
driftwire --config config.yaml      ← L2: 完全コントロール
```

**プラットフォーム化の達成**:
```
ツール（v0.3.0） → プラットフォーム（v0.4.1）
    単体で動く          他システムと連携
```

この記事が、オープンソースツールを「使いやすく」「つながる」ものに進化させる参考になれば幸いです！

## 参考リンク

- [GitHubリポジトリ](https://github.com/higakikeita/driftwire)
- [ホームページ](https://tfdrift-falco.vercel.app/)
- [README](https://github.com/higakikeita/driftwire/blob/main/README.md)

---

この記事は、Claude Code（Anthropic製AI開発アシスタント）との協力により作成されました。
