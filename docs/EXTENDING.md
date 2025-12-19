# Extending TFDrift-Falco

TFDrift-Falcoは拡張性を重視して設計されており、カスタムルール、通知チャネル、イベントハンドラーを簡単に追加できます。

## 📋 目次

1. [カスタムFalcoルールの追加](#カスタムfalcoルールの追加)
2. [カスタム通知チャネルの追加](#カスタム通知チャネルの追加)
3. [カスタムドリフトルールの追加](#カスタムドリフトルールの追加)
4. [カスタムリソースマッパーの追加](#カスタムリソースマッパーの追加)
5. [プラグインアーキテクチャ](#プラグインアーキテクチャ)

---

## カスタムFalcoルールの追加

Falcoルールを追加することで、TFDrift-Falcoが検知するCloudTrail/GCP Audit Logイベントをカスタマイズできます。

### 基本的なFalcoルール構造

Falcoルールは以下の要素で構成されます:

```yaml
- rule: ルール名
  desc: ルールの説明
  condition: イベント検知条件（Falco条件式）
  output: アラート出力形式
  priority: 優先度（DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY）
  tags: [タグ1, タグ2, ...]
```

### Example 1: S3バケット削除の検知

**シナリオ**: S3バケットが削除されたら即座にアラート

```yaml
# /etc/falco/falco_rules.local.yaml
- rule: S3 Bucket Deletion Detected
  desc: Detect when an S3 bucket is deleted via CloudTrail
  condition: >
    ct.name = "DeleteBucket"
    and not ct.user startswith "AWSServiceRole"
  output: >
    S3 Bucket Deletion Detected
    (user=%ct.user
     bucket=%ct.request.bucketname
     region=%ct.region
     source_ip=%ct.srcip
     event_time=%ct.time)
  priority: CRITICAL
  tags: [terraform, drift, s3, deletion]

- rule: S3 Bucket Public Access Enabled
  desc: Detect when S3 bucket is made public
  condition: >
    ct.name in (PutBucketAcl, PutBucketPolicy)
    and ct.response.publicaccessblock.blockpublicacls = false
  output: >
    S3 Bucket Made Public
    (user=%ct.user
     bucket=%ct.request.bucketname
     action=%ct.name
     region=%ct.region)
  priority: CRITICAL
  tags: [terraform, drift, s3, security]
```

### Example 2: IAM Admin権限付与の検知

**シナリオ**: IAMユーザーにAdministratorAccessポリシーがアタッチされたら警告

```yaml
- rule: IAM Administrator Access Granted
  desc: Detect when AdministratorAccess policy is attached to IAM user or role
  condition: >
    ct.name in (AttachUserPolicy, AttachRolePolicy)
    and ct.request.policyarn contains "AdministratorAccess"
  output: >
    IAM Administrator Access Granted
    (user=%ct.user
     target=%ct.request.username
     policy=%ct.request.policyarn
     region=%ct.region
     event_time=%ct.time)
  priority: CRITICAL
  tags: [terraform, drift, iam, security, privilege-escalation]

- rule: IAM Root Account Usage
  desc: Detect root account usage (should always use IAM users)
  condition: >
    ct.user = "root"
    and ct.useridentity.type = "Root"
  output: >
    Root Account Used (AVOID!)
    (action=%ct.name
     region=%ct.region
     source_ip=%ct.srcip
     event_time=%ct.time)
  priority: ALERT
  tags: [terraform, drift, iam, security, compliance]
```

### Example 3: RDS Encryption無効化の検知

```yaml
- rule: RDS Encryption Disabled
  desc: Detect when RDS encryption is disabled
  condition: >
    ct.name in (CreateDBInstance, ModifyDBInstance)
    and ct.request.storageencrypted = false
  output: >
    RDS Instance Created/Modified Without Encryption
    (user=%ct.user
     db_instance=%ct.request.dbinstanceidentifier
     action=%ct.name
     region=%ct.region)
  priority: HIGH
  tags: [terraform, drift, rds, security, encryption]
```

### Example 4: Security Group Port 22/3389 公開の検知

```yaml
- rule: Security Group Public SSH/RDP Access
  desc: Detect security group rules allowing SSH (22) or RDP (3389) from 0.0.0.0/0
  condition: >
    ct.name in (AuthorizeSecurityGroupIngress, CreateSecurityGroup)
    and (ct.request.ipprotocol = "tcp")
    and (ct.request.fromport = 22 or ct.request.fromport = 3389)
    and ct.request.cidrip contains "0.0.0.0/0"
  output: >
    Security Group Public SSH/RDP Access Allowed
    (user=%ct.user
     sg_id=%ct.request.groupid
     port=%ct.request.fromport
     cidr=%ct.request.cidrip
     region=%ct.region)
  priority: CRITICAL
  tags: [terraform, drift, security-group, network-security]
```

### Falcoルールのテスト

```bash
# Falco設定構文チェック
falco --validate /etc/falco/falco_rules.local.yaml

# ルールが正しく読み込まれるか確認
falco --list | grep "S3 Bucket Deletion Detected"

# ドライランモードでテスト（イベント処理のみ、アラートなし）
falco -c /etc/falco/falco.yaml --dry-run
```

### Falcoルールの有効化

```bash
# 1. カスタムルールファイルを配置
sudo cp custom_rules.yaml /etc/falco/falco_rules.local.yaml

# 2. Falco設定でカスタムルールを有効化
# /etc/falco/falco.yaml
rules_file:
  - /etc/falco/falco_rules.yaml
  - /etc/falco/falco_rules.local.yaml  # カスタムルール追加

# 3. Falcoを再起動
sudo systemctl restart falco

# 4. ログで確認
journalctl -u falco -f
```

---

## カスタム通知チャネルの追加

TFDrift-Falcoは複数の通知チャネル（Slack, Discord, Webhook）をサポートしていますが、カスタムチャネルを追加することも可能です。

### Architecture Overview

```
Drift Event
    ↓
NotificationManager (pkg/notifier/manager.go)
    ↓
Notifier Interface (pkg/notifier/notifier.go)
    ↓
    ├── SlackNotifier
    ├── DiscordNotifier
    ├── WebhookNotifier
    └── [Your Custom Notifier]  ← ここを追加
```

### Step 1: Notifierインターフェースの実装

**ファイル**: `pkg/notifier/custom_notifier.go`

```go
package notifier

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

// CustomNotifier は独自の通知チャネル実装
type CustomNotifier struct {
	apiEndpoint string
	apiKey      string
	timeout     time.Duration
}

// NewCustomNotifier creates a new custom notifier
func NewCustomNotifier(endpoint, apiKey string, timeout time.Duration) *CustomNotifier {
	return &CustomNotifier{
		apiEndpoint: endpoint,
		apiKey:      apiKey,
		timeout:     timeout,
	}
}

// Notify sends a drift event notification to the custom channel
func (n *CustomNotifier) Notify(ctx context.Context, event *DriftEvent) error {
	log.Infof("Sending notification to custom channel: %s", event.ResourceType)

	// 通知ペイロード作成
	payload := n.buildPayload(event)

	// HTTPリクエスト送信
	client := &http.Client{Timeout: n.timeout}
	req, err := http.NewRequestWithContext(ctx, "POST", n.apiEndpoint, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 認証ヘッダー追加
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.apiKey))
	req.Header.Set("Content-Type", "application/json")

	// リクエスト送信
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	// レスポンス確認
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notification failed with status %d: %s", resp.StatusCode, string(body))
	}

	log.Infof("Notification sent successfully to custom channel")
	return nil
}

// buildPayload creates the notification payload
func (n *CustomNotifier) buildPayload(event *DriftEvent) []byte {
	payload := map[string]interface{}{
		"event_type":    "terraform_drift",
		"resource_type": event.ResourceType,
		"resource_id":   event.ResourceID,
		"change_type":   event.ChangeType,
		"severity":      event.Severity,
		"user":          event.User,
		"region":        event.Region,
		"timestamp":     event.DetectedAt.Format(time.RFC3339),
		"details":       event.Details,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Errorf("Failed to marshal payload: %v", err)
		return []byte("{}")
	}

	return data
}

// Name returns the notifier name
func (n *CustomNotifier) Name() string {
	return "custom"
}
```

### Step 2: 設定構造体の追加

**ファイル**: `pkg/config/config.go`

```go
// NotificationConfig contains notification settings
type NotificationConfig struct {
	Slack   *SlackConfig   `yaml:"slack"`
	Discord *DiscordConfig `yaml:"discord"`
	Webhook *WebhookConfig `yaml:"webhook"`
	Custom  *CustomConfig  `yaml:"custom"`  // 追加
}

// CustomConfig contains custom notifier settings
type CustomConfig struct {
	Enabled     bool   `yaml:"enabled"`
	APIEndpoint string `yaml:"api_endpoint"`
	APIKey      string `yaml:"api_key"`
	Timeout     string `yaml:"timeout"`  // "10s", "30s" など
}
```

### Step 3: Notification Managerへの統合

**ファイル**: `pkg/notifier/manager.go`

```go
// NewNotificationManager creates a new notification manager
func NewNotificationManager(cfg *config.Config) *NotificationManager {
	notifiers := []Notifier{}

	// Slack
	if cfg.Notifications.Slack != nil && cfg.Notifications.Slack.Enabled {
		notifiers = append(notifiers, NewSlackNotifier(cfg.Notifications.Slack))
	}

	// Discord
	if cfg.Notifications.Discord != nil && cfg.Notifications.Discord.Enabled {
		notifiers = append(notifiers, NewDiscordNotifier(cfg.Notifications.Discord))
	}

	// Webhook
	if cfg.Notifications.Webhook != nil && cfg.Notifications.Webhook.Enabled {
		notifiers = append(notifiers, NewWebhookNotifier(cfg.Notifications.Webhook))
	}

	// Custom (追加)
	if cfg.Notifications.Custom != nil && cfg.Notifications.Custom.Enabled {
		timeout, _ := time.ParseDuration(cfg.Notifications.Custom.Timeout)
		if timeout == 0 {
			timeout = 10 * time.Second
		}
		notifiers = append(notifiers, NewCustomNotifier(
			cfg.Notifications.Custom.APIEndpoint,
			cfg.Notifications.Custom.APIKey,
			timeout,
		))
	}

	return &NotificationManager{
		notifiers: notifiers,
	}
}
```

### Step 4: 設定ファイルで有効化

**ファイル**: `config.yaml`

```yaml
notifications:
  slack:
    enabled: true
    webhook_url: "https://hooks.slack.com/services/..."

  custom:
    enabled: true
    api_endpoint: "https://your-custom-api.example.com/notifications"
    api_key: "${CUSTOM_API_KEY}"  # 環境変数から読み込み
    timeout: "30s"
```

### Example: PagerDuty統合

```go
// pkg/notifier/pagerduty_notifier.go
package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

type PagerDutyNotifier struct {
	integrationKey string
	timeout        time.Duration
}

func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		timeout:        10 * time.Second,
	}
}

func (n *PagerDutyNotifier) Notify(ctx context.Context, event *DriftEvent) error {
	// PagerDuty Events API v2 ペイロード
	payload := map[string]interface{}{
		"routing_key":  n.integrationKey,
		"event_action": "trigger",
		"payload": map[string]interface{}{
			"summary":  fmt.Sprintf("Drift Detected: %s", event.ResourceType),
			"severity": n.mapSeverity(event.Severity),
			"source":   "tfdrift-falco",
			"custom_details": map[string]interface{}{
				"resource_type": event.ResourceType,
				"resource_id":   event.ResourceID,
				"user":          event.User,
				"region":        event.Region,
				"change_type":   event.ChangeType,
			},
		},
	}

	data, _ := json.Marshal(payload)

	client := &http.Client{Timeout: n.timeout}
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://events.pagerduty.com/v2/enqueue", bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create PagerDuty request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send PagerDuty notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("PagerDuty API returned status %d", resp.StatusCode)
	}

	log.Info("PagerDuty incident created successfully")
	return nil
}

func (n *PagerDutyNotifier) mapSeverity(severity string) string {
	switch severity {
	case "critical":
		return "critical"
	case "high":
		return "error"
	case "medium":
		return "warning"
	default:
		return "info"
	}
}

func (n *PagerDutyNotifier) Name() string {
	return "pagerduty"
}
```

---

## カスタムドリフトルールの追加

TFDrift-Falcoのドリフト検知ルールは完全にカスタマイズ可能です。

### ドリフトルールの構造

```yaml
drift_rules:
  - name: "ルール名"
    resource_types:
      - "aws_instance"
      - "aws_security_group"
    watched_attributes:
      - "instance_type"
      - "disable_api_termination"
    severity: "high"  # critical, high, medium, low
    exclude_users:
      - "AWSServiceRoleForAutoScaling"
    environment: "production"  # production, staging, development
```

### Example 1: 環境固有のルール

```yaml
drift_rules:
  # 本番環境: 全ての変更を検知（最も厳格）
  - name: "Production - All Changes"
    resource_types:
      - "*"
    watched_attributes:
      - "*"
    severity: "critical"
    environment: "production"
    exclude_users:
      - "terraform-automation"

  # ステージング環境: セキュリティ関連のみ
  - name: "Staging - Security Changes"
    resource_types:
      - "aws_security_group*"
      - "aws_iam_*"
      - "aws_kms_*"
    watched_attributes:
      - "*"
    severity: "high"
    environment: "staging"

  # 開発環境: IAMのみ（開発者の自由度を確保）
  - name: "Development - IAM Only"
    resource_types:
      - "aws_iam_*"
    watched_attributes:
      - "policy"
      - "assume_role_policy"
    severity: "medium"
    environment: "development"
```

### Example 2: 属性レベルの細かい制御

```yaml
drift_rules:
  # EC2インスタンス: Criticalな属性のみ
  - name: "EC2 Critical Attributes"
    resource_types:
      - "aws_instance"
    watched_attributes:
      - "instance_type"          # コスト影響大
      - "disable_api_termination" # セキュリティ重要
      - "security_groups"         # セキュリティ重要
      - "iam_instance_profile"    # セキュリティ重要
    # Tags変更は除外（ノイズが多い）
    severity: "high"

  # RDS: 暗号化とバックアップ設定
  - name: "RDS Security Settings"
    resource_types:
      - "aws_db_instance"
    watched_attributes:
      - "storage_encrypted"
      - "backup_retention_period"
      - "deletion_protection"
      - "publicly_accessible"
    severity: "critical"

  # S3: Public Access設定のみ
  - name: "S3 Public Access"
    resource_types:
      - "aws_s3_bucket"
      - "aws_s3_bucket_public_access_block"
    watched_attributes:
      - "block_public_acls"
      - "block_public_policy"
      - "ignore_public_acls"
      - "restrict_public_buckets"
    severity: "critical"
```

### Example 3: 時間帯ベースのルール（将来機能）

```yaml
drift_rules:
  # 営業時間外の変更は全て検知
  - name: "After Hours Changes"
    resource_types:
      - "*"
    watched_attributes:
      - "*"
    severity: "critical"
    schedule:
      allowed_hours: "09:00-18:00"  # 許可時間帯
      timezone: "Asia/Tokyo"
      weekdays_only: true  # 平日のみ

  # 営業時間内は重要なリソースのみ
  - name: "Business Hours - Critical Only"
    resource_types:
      - "aws_iam_*"
      - "aws_kms_*"
    watched_attributes:
      - "*"
    severity: "high"
    schedule:
      allowed_hours: "09:00-18:00"
      timezone: "Asia/Tokyo"
```

---

## カスタムリソースマッパーの追加

新しいCloudTrailイベントやリソースタイプをサポートするには、リソースマッパーを拡張します。

### リソースマッパーの構造

**ファイル**: `pkg/falco/resource_mapper.go`

```go
func (s *Subscriber) mapEventToResourceType(eventName string, eventSource string) string {
	// イベント名の衝突解決（eventSource使用）
	switch eventName {
	case "CreateAlias", "DeleteAlias", "UpdateAlias":
		if eventSource == "lambda.amazonaws.com" {
			return "aws_lambda_alias"
		}
		if eventSource == "kms.amazonaws.com" || eventSource == "" {
			return "aws_kms_alias"
		}
	}

	// 標準マッピング
	mapping := map[string]string{
		// EC2
		"RunInstances":       "aws_instance",
		"TerminateInstances": "aws_instance",

		// 新しいサービスを追加
		"CreateCluster":      "aws_ecs_cluster",
		"DeleteCluster":      "aws_ecs_cluster",
	}

	if resourceType, ok := mapping[eventName]; ok {
		return resourceType
	}

	return "unknown"
}
```

### Example: 新しいAWSサービスの追加（AWS App Runner）

```go
// pkg/falco/resource_mapper.go

mapping := map[string]string{
	// ... existing mappings ...

	// AWS App Runner (新規追加)
	"CreateService":        "aws_apprunner_service",
	"DeleteService":        "aws_apprunner_service",
	"UpdateService":        "aws_apprunner_service",
	"CreateAutoScalingConfiguration": "aws_apprunner_auto_scaling_configuration_version",
	"DeleteAutoScalingConfiguration": "aws_apprunner_auto_scaling_configuration_version",
	"CreateConnection":     "aws_apprunner_connection",
	"DeleteConnection":     "aws_apprunner_connection",
	"CreateVpcConnector":   "aws_apprunner_vpc_connector",
	"DeleteVpcConnector":   "aws_apprunner_vpc_connector",
}
```

### Example: GCPサービスの追加（Cloud Spanner）

```go
// pkg/gcp/resource_mapper.go

func mapGCPEventToResourceType(methodName string, resourceType string) string {
	mapping := map[string]string{
		// ... existing mappings ...

		// Cloud Spanner (新規追加)
		"google.spanner.admin.instance.v1.InstanceAdmin.CreateInstance": "google_spanner_instance",
		"google.spanner.admin.instance.v1.InstanceAdmin.DeleteInstance": "google_spanner_instance",
		"google.spanner.admin.instance.v1.InstanceAdmin.UpdateInstance": "google_spanner_instance",
		"google.spanner.admin.database.v1.DatabaseAdmin.CreateDatabase": "google_spanner_database",
		"google.spanner.admin.database.v1.DatabaseAdmin.DropDatabase":   "google_spanner_database",
	}

	if rt, ok := mapping[methodName]; ok {
		return rt
	}

	return "unknown"
}
```

### テストの追加

**ファイル**: `pkg/falco/resource_mapper_test.go`

```go
func TestMapEventToResourceType_AppRunner(t *testing.T) {
	sub := &Subscriber{}

	tests := []struct {
		name        string
		eventName   string
		eventSource string
		want        string
	}{
		{"App Runner Service Create", "CreateService", "apprunner.amazonaws.com", "aws_apprunner_service"},
		{"App Runner Service Delete", "DeleteService", "apprunner.amazonaws.com", "aws_apprunner_service"},
		{"App Runner AutoScaling", "CreateAutoScalingConfiguration", "apprunner.amazonaws.com", "aws_apprunner_auto_scaling_configuration_version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sub.mapEventToResourceType(tt.eventName, tt.eventSource)
			assert.Equal(t, tt.want, got)
		})
	}
}
```

---

## プラグインアーキテクチャ

TFDrift-Falcoは将来的にプラグインアーキテクチャをサポート予定です（v0.6.0+）。

### プラグインインターフェース（計画中）

```go
// pkg/plugin/plugin.go
package plugin

import (
	"context"
)

// Plugin is the interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Initialize initializes the plugin with configuration
	Initialize(config map[string]interface{}) error

	// OnDriftDetected is called when a drift is detected
	OnDriftDetected(ctx context.Context, event *DriftEvent) error

	// OnStateRefresh is called when Terraform state is refreshed
	OnStateRefresh(ctx context.Context, state *TerraformState) error

	// Cleanup cleans up plugin resources
	Cleanup() error
}

// DriftEvent represents a detected drift event
type DriftEvent struct {
	ResourceType string
	ResourceID   string
	ChangeType   string
	Severity     string
	User         string
	Region       string
	DetectedAt   time.Time
	Details      map[string]interface{}
}
```

### プラグイン実装例: Auto-Remediation

```go
// plugins/auto_remediation/plugin.go
package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	log "github.com/sirupsen/logrus"
)

type AutoRemediationPlugin struct {
	ec2Client *ec2.Client
	enabled   bool
}

func (p *AutoRemediationPlugin) Name() string {
	return "auto-remediation"
}

func (p *AutoRemediationPlugin) Version() string {
	return "0.1.0"
}

func (p *AutoRemediationPlugin) Initialize(cfg map[string]interface{}) error {
	p.enabled = cfg["enabled"].(bool)

	// AWS SDK初期化
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	p.ec2Client = ec2.NewFromConfig(awsCfg)
	log.Info("Auto-remediation plugin initialized")
	return nil
}

func (p *AutoRemediationPlugin) OnDriftDetected(ctx context.Context, event *DriftEvent) error {
	if !p.enabled {
		return nil
	}

	log.Infof("Auto-remediation triggered for %s", event.ResourceType)

	// EC2インスタンスのTermination Protection無効化を自動修正
	if event.ResourceType == "aws_instance" && event.ChangeType == "disable_api_termination" {
		return p.reEnableTerminationProtection(ctx, event.ResourceID)
	}

	return nil
}

func (p *AutoRemediationPlugin) reEnableTerminationProtection(ctx context.Context, instanceID string) error {
	_, err := p.ec2Client.ModifyInstanceAttribute(ctx, &ec2.ModifyInstanceAttributeInput{
		InstanceId: &instanceID,
		DisableApiTermination: &types.AttributeBooleanValue{
			Value: aws.Bool(true),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to re-enable termination protection: %w", err)
	}

	log.Infof("Termination protection re-enabled for instance %s", instanceID)
	return nil
}

func (p *AutoRemediationPlugin) OnStateRefresh(ctx context.Context, state *TerraformState) error {
	// State refresh時の処理（必要に応じて実装）
	return nil
}

func (p *AutoRemediationPlugin) Cleanup() error {
	log.Info("Auto-remediation plugin cleaned up")
	return nil
}

// プラグインのエクスポート（Go 1.8+ plugin system）
var Plugin AutoRemediationPlugin
```

### プラグインの使用（設定ファイル）

```yaml
# config.yaml
plugins:
  - name: "auto-remediation"
    enabled: true
    config:
      dry_run: false  # true: 実際には修正しない（ログのみ）
      allowed_resources:
        - "aws_instance"
        - "aws_security_group"
      excluded_accounts:
        - "123456789012"  # 本番アカウントは除外

  - name: "cost-calculator"
    enabled: true
    config:
      currency: "USD"
      region: "us-east-1"
```

---

## Contributing Your Extension

カスタム拡張を作成したら、ぜひコミュニティに共有してください！

### プルリクエストの手順

1. **フォークとクローン**
   ```bash
   git clone https://github.com/YOUR_USERNAME/tfdrift-falco.git
   cd tfdrift-falco
   ```

2. **ブランチ作成**
   ```bash
   git checkout -b feature/add-pagerduty-notifier
   ```

3. **実装とテスト**
   ```bash
   # 実装
   vim pkg/notifier/pagerduty_notifier.go

   # テスト追加
   vim pkg/notifier/pagerduty_notifier_test.go

   # テスト実行
   go test ./pkg/notifier/...
   ```

4. **ドキュメント更新**
   ```bash
   # README更新
   vim README.md

   # 拡張ドキュメント更新
   vim docs/EXTENDING.md
   ```

5. **コミットとプッシュ**
   ```bash
   git add .
   git commit -m "feat(notifier): add PagerDuty integration"
   git push origin feature/add-pagerduty-notifier
   ```

6. **プルリクエスト作成**
   - GitHubでPRを作成
   - 変更内容、テスト結果、使用例を記載

### コーディング規約

- **Go Formatting**: `gofmt` と `goimports` を使用
- **Linting**: `golangci-lint` でエラーなし
- **Testing**: 80%+ のカバレッジ
- **Documentation**: 全ての公開関数にGoDocコメント
- **Commit Messages**: [Conventional Commits](https://www.conventionalcommits.org/) 形式

---

## Examples Repository

コミュニティによる拡張例を集めたリポジトリ（計画中）:

- **tfdrift-falco-plugins** - 公式プラグインコレクション
- **tfdrift-falco-rules** - カスタムFalcoルール集
- **tfdrift-falco-notifiers** - サードパーティ通知チャネル

---

## Support

拡張機能の開発でサポートが必要な場合:

- **GitHub Discussions**: https://github.com/higakikeita/tfdrift-falco/discussions
- **Slack Community**: [Join Slack](https://join.slack.com/t/tfdrift-falco/...)
- **Email**: support@tfdrift-falco.io

---

**次のステップ**:
- [Use Cases](USE_CASES.md) - 実際の使用例を確認
- [Best Practices](BEST_PRACTICES.md) - 本番運用のベストプラクティス
- [Contributing Guide](../CONTRIBUTING.md) - コントリビューションガイドライン
