---
title: Go言語プロジェクトのテストカバレッジを0%から60%に向上させた話
tags: Go, Testing, CI/CD, DevOps, テスト
author: [Your Name]
slide: false
---

# Go言語プロジェクトのテストカバレッジを0%から60%に向上させた話

## はじめに

この記事では、オープンソースプロジェクト「TFDrift-Falco」において、テストカバレッジを**0%から59.8%**まで向上させた取り組みを紹介します。

4週間で250以上のテストケースを追加し、CI/CDパイプラインを構築した実践的なアプローチをお伝えします。

## TL;DR

- 🎯 **成果**: テストカバレッジ 0% → 59.8% (4週間)
- 📝 **テスト数**: 250+テストケース、13ファイル、~3,900行
- 🚀 **CI/CD**: GitHub Actions + golangci-lint (17 linters)
- 🛠️ **ツール**: testify, httptest, カスタムモック

## プロジェクト概要

**TFDrift-Falco**は、Falcoを使ってTerraformのドリフト（設定の差異）をリアルタイムで検出するツールです。

```
開始時:
├── 実装コード: 2,624行
├── テストコード: 0行
└── カバレッジ: 0%

完了時:
├── 実装コード: 2,624行
├── テストコード: ~3,900行
└── カバレッジ: 59.8% ✅
```

## 戦略: 4フェーズアプローチ

依存関係と難易度を考慮し、4フェーズに分けて実装しました。

### Phase 1: 基盤（Week 1）- 目標15%

**対象**: `pkg/types`, `pkg/config`

まず、依存関係のない基盤パッケージから着手。設定ファイルの読み込みとバリデーションをテスト：

```go
func TestLoad_ValidConfig(t *testing.T) {
    cfg, err := Load("testdata/valid_config.yaml")
    require.NoError(t, err)

    assert.True(t, cfg.Providers.AWS.Enabled)
    assert.Equal(t, []string{"us-east-1"}, cfg.Providers.AWS.Regions)
}
```

**成果**:
- 17テストケース作成
- **90.9%カバレッジ達成** ✅
- Phase 1完了時: ~15%

### Phase 2: コアロジック（Week 2）- 目標31%

**対象**: `pkg/terraform/state`, `pkg/detector`

Terraform stateの管理とドリフト検出のコアロジックをテスト。スレッドセーフティも検証：

```go
func TestStateManager_ThreadSafety(t *testing.T) {
    sm, err := NewStateManager(cfg)
    require.NoError(t, err)

    err = sm.Load(context.Background())
    require.NoError(t, err)

    // 10 goroutines で同時アクセス
    done := make(chan bool)
    for i := 0; i < 10; i++ {
        go func() {
            resource, exists := sm.GetResource("i-1234567890abcdef0")
            assert.True(t, exists)
            done <- true
        }()
    }

    for i := 0; i < 10; i++ {
        <-done
    }
}
```

**成果**:
- 37テストケース作成
- state.goは**100%カバレッジ** ✅
- Phase 2完了時: 31.2%

### Phase 3: 統合機能（Week 3）- 目標37%

**対象**: `pkg/diff`, `pkg/metrics`

5種類のdiffフォーマッター（Console, UnifiedDiff, Markdown, JSON, SideBySide）を包括的にテスト：

```go
func TestFormatMarkdown(t *testing.T) {
    formatter := NewFormatter(false)

    alert := &types.DriftAlert{
        Severity:     "high",
        ResourceType: "aws_instance",
        ResourceName: "web",
        Attribute:    "instance_type",
        OldValue:     "t3.micro",
        NewValue:     "t3.large",
    }

    result := formatter.FormatMarkdown(alert)

    assert.Contains(t, result, "## 🚨 Drift Alert")
    assert.Contains(t, result, "**Severity:** high")
    assert.Contains(t, result, "`aws_instance.web`")
}
```

**成果**:
- 42テストケース作成
- pkg/diff: **96.0%カバレッジ** ✅
- Phase 3完了時: 36.9%

### Phase 4: 外部依存（Week 4）- 目標52%

**対象**: `pkg/falco`, `pkg/notifier`, `pkg/terraform/importer`, `pkg/terraform/approval`

MockHTTPServerを活用してWebhook送信をテスト：

```go
func TestSend_Slack(t *testing.T) {
    mockServer := testutil.NewMockHTTPServer()
    defer mockServer.Close()

    cfg := config.NotificationsConfig{
        Slack: config.SlackConfig{
            Enabled:    true,
            WebhookURL: mockServer.URL(),
            Channel:    "#alerts",
        },
    }

    manager, err := NewManager(cfg)
    require.NoError(t, err)

    alert := testutil.CreateTestDriftAlert()
    err = manager.Send(alert)
    assert.NoError(t, err)

    // Verify request was sent
    assert.Equal(t, 1, mockServer.GetRequestCount())

    // Verify payload
    body := mockServer.GetLastRequestBody()
    var payload map[string]interface{}
    json.Unmarshal([]byte(body), &payload)

    assert.Equal(t, "#alerts", payload["channel"])
    assert.Contains(t, payload, "blocks")
}
```

**成果**:
- 104テストケース作成
- pkg/notifier: **95.5%カバレッジ** ✅
- Phase 4完了時: **52.2%** 🎉

### Phase 5: コマンドラインテスト（Week 5）- 目標60%

**対象**: `cmd/tfdrift`, `cmd/test-drift`, `pkg/detector`追加テスト

コマンドラインツールとDetectorの残りの関数をテスト：

```go
func TestNewApprovalCmd(t *testing.T) {
    cmd := newApprovalCmd()

    assert.NotNil(t, cmd)
    assert.Equal(t, "approval", cmd.Use)
    assert.True(t, cmd.HasSubCommands())

    // Verify all subcommands are present
    subcommands := cmd.Commands()
    assert.Len(t, subcommands, 4)
}

func TestNew_WithAutoImport(t *testing.T) {
    cfg := &config.Config{
        AutoImport: config.AutoImportConfig{
            Enabled:         true,
            TerraformDir:    ".",
            RequireApproval: true,
        },
    }

    detector, err := New(cfg)

    assert.NoError(t, err)
    assert.NotNil(t, detector.importer)
    assert.NotNil(t, detector.approvalManager)
}
```

**成果**:
- cmd/tfdrift: **47.2%カバレッジ** ✅
- pkg/detector追加テスト: **51.8%カバレッジ** ✅
- Phase 5完了時: **59.8%** 🎉

## CI/CD環境の構築

### GitHub Actions ワークフロー

マトリックステストとカバレッジチェックを自動化：

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go ${{ matrix.go-version }}
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Check coverage threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
          THRESHOLD=30.0
          if (( $(echo "$COVERAGE < $THRESHOLD" | bc -l) )); then
            echo "❌ Coverage ${COVERAGE}% is below threshold"
            exit 1
          fi
```

### golangci-lint設定

17個のlinterを有効化：

```yaml
# .golangci.yml
linters:
  enable:
    - errcheck      # Unchecked errors
    - gosimple      # Simplify code
    - govet         # Go vet
    - staticcheck   # Static analysis
    - unused        # Unused code
    - gofmt         # Formatting
    - goimports     # Import formatting
    - misspell      # Spelling
    - revive        # Fast linter
    - gosec         # Security
    - gocritic      # Extensible linter
    - unparam       # Unused parameters
```

### Makefile

ローカルでCIチェックを実行できるように：

```makefile
# Run all CI checks locally
ci: deps fmt lint test-coverage-threshold test-race
	@echo "✅ All CI checks passed!"

# Quick CI checks without race detector
ci-local: fmt lint test-coverage
	@echo "✅ Local CI checks passed!"
```

## テストユーティリティの整備

### pkg/testutil パッケージ

再利用可能なヘルパーを作成：

```go
// fixtures.go - テストデータ生成
func CreateTestDriftAlert() *types.DriftAlert {
    return &types.DriftAlert{
        Severity:     "high",
        ResourceType: "aws_instance",
        ResourceName: "web",
        ResourceID:   "i-1234567890abcdef0",
        Attribute:    "instance_type",
        OldValue:     "t3.micro",
        NewValue:     "t3.small",
        UserIdentity: types.UserIdentity{
            Type:     "IAMUser",
            UserName: "admin",
        },
        MatchedRules: []string{"instance_type_change"},
    }
}

// mock_http.go - HTTPサーバーモック
type MockHTTPServer struct {
    Server        *httptest.Server
    requests      []*http.Request
    requestBodies []string
    statusCode    int
}

func NewMockHTTPServer() *MockHTTPServer {
    mock := &MockHTTPServer{
        statusCode: http.StatusOK,
    }
    mock.Server = httptest.NewServer(http.HandlerFunc(mock.handleRequest))
    return mock
}
```

## 遭遇した課題と解決策

### 1. Prometheus メトリクス重複登録エラー

**問題**:
```
panic: duplicate metrics collector registration attempted
```

テスト実行時に同じメトリクスを複数回登録しようとしてpanicが発生。

**解決策**: Singleton patternを採用

```go
var testMetrics *Metrics

func init() {
    testMetrics = NewMetrics("tfdrift_test")
}

func TestRecordDriftAlert(t *testing.T) {
    m := testMetrics  // 全テストで同じインスタンスを使用
    assert.NotPanics(t, func() {
        m.RecordDriftAlert("critical", "aws_instance", "aws")
    })
}
```

### 2. JSON unmarshalの型変換

**問題**: Goの`json.Unmarshal`はすべての数値を`float64`に変換

```go
// Before (失敗):
assert.Equal(t, 0xFF0000, embed["color"])

// After (成功):
assert.Equal(t, float64(0xFF0000), embed["color"])
```

### 3. nil vs 空スライス

**問題**: Goでは`[]string{}`と`[]string(nil)`は異なる

```go
// Before (失敗):
expected: []string{},

// After (成功):
expected: nil,  // Go では nil == empty slice
```

## 最終結果

| パッケージ | カバレッジ | 評価 |
|-----------|-----------|------|
| pkg/diff | 96.0% | ⭐⭐⭐ |
| pkg/notifier | 95.5% | ⭐⭐⭐ |
| pkg/config | 90.9% | ⭐⭐⭐ |
| pkg/metrics | 81.2% | ⭐⭐ |
| pkg/terraform | 77.2% | ⭐⭐ |
| pkg/falco | 63.0% | ⭐ |
| pkg/detector | 51.8% | ⭐ |
| cmd/tfdrift | 47.2% | ⭐ |
| **全体** | **59.8%** | **✅** |

## ベストプラクティス

### 1. Table-Driven Tests

複数のテストケースを効率的に管理：

```go
func TestGenerateResourceName(t *testing.T) {
    tests := []struct {
        name       string
        resourceID string
        want       string
    }{
        {
            name:       "EC2 Instance ID",
            resourceID: "i-123",
            want:       "i_123",
        },
        {
            name:       "IAM Role ARN",
            resourceID: "arn:aws:iam::123:role/MyRole",
            want:       "arn_aws_iam__123_role_MyRole",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := generateResourceName(tt.resourceID)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### 2. t.Helper() の活用

スタックトレースをクリーンに保つ：

```go
func setupTest(t *testing.T) (*Config, func()) {
    t.Helper()  // この関数をスタックトレースから除外

    config := &Config{...}
    cleanup := func() {
        // cleanup logic
    }

    return config, cleanup
}
```

### 3. 段階的アプローチ

```
Week 1: 基盤（簡単）     → 15%
Week 2: コア（中程度）   → 31%
Week 3: 統合（やや難）   → 37%
Week 4: 外部依存（難）   → 52%
Week 5: CLI + 追加（仕上げ） → 60%
```

一度に全てをテストしようとせず、優先度をつけて段階的に進める。

## 学んだこと

### ✅ Do's

- **依存関係の少ないパッケージから始める**
  - types → config → state → detector の順で
- **テストユーティリティを早めに整備**
  - 再利用可能なモックやヘルパーを作成
- **CI/CDを同時に構築**
  - テストが書けたらすぐCI化
- **モックは必要最小限に**
  - 複雑なモックより単純なロジックテスト

### ❌ Don'ts

- **全てを一度にテストしようとしない**
  - 完璧主義は進捗を妨げる
- **複雑なモックを作りすぎない**
  - テストがメンテナンス不能になる
- **カバレッジだけを追わない**
  - 質の高いテストを書く
- **テストのメンテナンスを怠らない**
  - リファクタ時は必ずテストも更新

## まとめ

テストカバレッジ向上は単なる数値目標ではなく、**開発文化の変革**です：

- 🎯 コードの信頼性向上
- 🚀 安心してリファクタリング
- 🐛 バグの早期発見
- 📚 実行可能なドキュメント
- 🤝 チーム開発の基盤

重要なのは、**完璧を目指さず、段階的に改善し続けること**です。

この記事が、テストカバレッジ向上に取り組む方の参考になれば幸いです！

## 参考リンク

- [GitHubリポジトリ](https://github.com/higakikeita/tfdrift-falco)
- [詳細記事（英語版）](https://github.com/higakikeita/tfdrift-falco/blob/main/docs/test-coverage-improvement-journey.md)
- [Go Testing Documentation](https://golang.org/doc/tutorial/add-a-test)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

---

この記事は、Claude Code（Anthropic製AI開発アシスタント）との協力により作成されました。
