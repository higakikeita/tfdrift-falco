# Go言語プロジェクトのテストカバレッジを0%から65%に向上させた話

## TL;DR

- 🎯 **成果**: テストカバレッジ 0% → 65.0% (6週間)
- 📝 **テスト数**: 265+テストケース、13ファイル
- 🚀 **CI/CD**: GitHub Actions + golangci-lint (17 linters)
- 🛠️ **ツール**: testify, httptest, カスタムモック

## プロジェクト概要

**TFDrift-Falco**: Falcoを使ったTerraformドリフト検出ツール（2,624行）

```
Before:
├── テストコード: 0行
├── カバレッジ: 0%
└── CI/CD: なし

After:
├── テストコード: ~4,150行
├── カバレッジ: 65.0%
└── CI/CD: 完全自動化
```

## 5フェーズ戦略

### Phase 1: 基盤（Week 1）
**対象**: `pkg/types`, `pkg/config`

```go
func TestLoad_ValidConfig(t *testing.T) {
    cfg, err := Load("testdata/valid_config.yaml")
    require.NoError(t, err)
    assert.True(t, cfg.Providers.AWS.Enabled)
}
```

**成果**: 90.9%カバレッジ ✅

### Phase 2: コアロジック（Week 2）
**対象**: `pkg/terraform/state`, `pkg/detector`

スレッドセーフティのテスト：

```go
func TestStateManager_ThreadSafety(t *testing.T) {
    // 10 goroutines で同時アクセス
    for i := 0; i < 10; i++ {
        go func() {
            resource, exists := sm.GetResource("i-123")
            assert.True(t, exists)
            done <- true
        }()
    }
}
```

**成果**: state.goは100%カバレッジ ✅

### Phase 3: 統合機能（Week 3）
**対象**: `pkg/diff`, `pkg/metrics`

5種類のdiffフォーマッター：

```go
tests := []string{"Console", "UnifiedDiff", "Markdown", "JSON", "SideBySide"}
for _, format := range tests {
    t.Run(format, func(t *testing.T) {
        // Test each format
    })
}
```

**成果**: 96.0%カバレッジ ✅

### Phase 4: 外部依存（Week 4）
**対象**: `pkg/falco`, `pkg/notifier`, `pkg/terraform/*`

MockHTTPServerでWebhookテスト：

```go
func TestSend_Slack(t *testing.T) {
    mockServer := testutil.NewMockHTTPServer()
    defer mockServer.Close()

    manager.Send(alert)

    assert.Equal(t, 1, mockServer.GetRequestCount())
    payload := mockServer.GetLastRequestBody()
    // Verify payload
}
```

**成果**: 63-95.5%カバレッジ ✅

### Phase 5: CLI + 統合テスト（Week 5-6）
**対象**: `cmd/tfdrift`, `cmd/test-drift`, `pkg/detector`統合テスト

CLIツールと統合テストを追加：

```go
func TestNewApprovalCmd(t *testing.T) {
    cmd := newApprovalCmd()
    assert.True(t, cmd.HasSubCommands())
    assert.Len(t, cmd.Commands(), 4)
}

func TestStart_Integration(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
    defer cancel()

    err = detector.Start(ctx)
    assert.NoError(t, err)
}
```

**成果**: cmd/tfdrift 47.2%, detector 51.8%→80.1% ✅

## CI/CD構築

### GitHub Actions

```yaml
# .github/workflows/test.yml
jobs:
  test:
    strategy:
      matrix:
        go-version: ['1.21', '1.22', '1.23']
    steps:
      - run: go test -race -coverprofile=coverage.out ./...
      - name: Check threshold
        run: |
          COVERAGE=$(go tool cover -func=coverage.out | grep total)
          if (( COVERAGE < 30.0 )); then exit 1; fi
```

### Makefile

```makefile
ci: deps fmt lint test-coverage-threshold test-race
	@echo "✅ All CI checks passed!"
```

## テストユーティリティ

### pkg/testutil パッケージ

```go
// fixtures.go - テストデータ生成
func CreateTestDriftAlert() *types.DriftAlert { ... }
func CreateTestConfig() *config.Config { ... }

// mock_http.go - HTTPモック
type MockHTTPServer struct {
    Server        *httptest.Server
    requests      []*http.Request
    requestBodies []string
}

// mock_falco.go - Falcoクライアントモック
type MockFalcoClient struct {
    events []*types.Event
}
```

## 遭遇した課題

### 1. Prometheus重複登録エラー

**問題**:
```
panic: duplicate metrics collector registration attempted
```

**解決**: Singleton pattern

```go
var testMetrics *Metrics
func init() { testMetrics = NewMetrics("test") }

func TestRecordDriftAlert(t *testing.T) {
    m := testMetrics // 全テストで同じインスタンス
}
```

### 2. JSON unmarshalの型変換

**問題**: `int` → `float64` への変換

```go
// Before (失敗):
assert.Equal(t, 0xFF0000, embed["color"])

// After (成功):
assert.Equal(t, float64(0xFF0000), embed["color"])
```

### 3. nil vs 空スライス

```go
// Before:
expected: []string{},

// After:
expected: nil,  // Go では nil == empty slice
```

## 最終結果

| パッケージ | カバレッジ | 評価 |
|-----------|-----------|------|
| pkg/diff | 96.0% | ⭐⭐⭐ |
| pkg/notifier | 95.5% | ⭐⭐⭐ |
| pkg/config | 90.9% | ⭐⭐⭐ |
| pkg/detector | 86.7% | ⭐⭐⭐ |
| pkg/metrics | 81.2% | ⭐⭐ |
| pkg/terraform | 77.2% | ⭐⭐ |
| pkg/falco | 63.0% | ⭐ |
| cmd/tfdrift | 47.2% | ⭐ |
| **全体** | **65.0%** | **✅** |

## ベストプラクティス

### 1. Table-Driven Tests

```go
tests := []struct {
    name string
    input string
    want string
}{
    {"case1", "input1", "output1"},
    {"case2", "input2", "output2"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        got := function(tt.input)
        assert.Equal(t, tt.want, got)
    })
}
```

### 2. t.Helper() の活用

```go
func setupTest(t *testing.T) (*Config, func()) {
    t.Helper()  // スタックトレースから除外
    // setup logic
    return config, cleanup
}
```

### 3. 段階的アプローチ

```
Week 1: 基盤（簡単）           → 15%
Week 2: コア（中程度）         → 31%
Week 3: 統合（やや難）         → 37%
Week 4: 外部依存（難）         → 52%
Week 5: CLI + 追加             → 60%
Week 6: 統合テスト + handleEvent → 65%
```

## 学んだこと

### ✅ Do's

- 依存関係の少ないパッケージから始める
- テストユーティリティを早めに整備
- CI/CDを同時に構築
- モックは必要最小限に

### ❌ Don'ts

- 全てを一度にテストしようとしない
- 複雑なモックを作りすぎない
- カバレッジだけを追わない
- テストのメンテナンスを怠らない

## 次のステップ

### 短期（1-2ヶ月）
- [x] pkg/detector: 21% → 86.7% ✅
- [x] cmd/: 0% → 47.2% ✅
- [x] 統合テスト追加 ✅
- [x] handleEvent関数のカバレッジ向上（36.4% → 95.5%）✅
- [x] カバレッジ65%達成 ✅
- [ ] カバレッジ70%達成（次フェーズ）

### 中期（3-6ヶ月）
- [ ] カバレッジ75%達成
- [ ] パフォーマンステスト
- [ ] Fuzzing導入

### 長期（6ヶ月+）
- [ ] Mutation testing
- [ ] カオスエンジニアリング

## まとめ

テストカバレッジ向上は単なる数値目標ではなく、**開発文化の変革**です：

- 🎯 コードの信頼性向上
- 🚀 安心してリファクタリング
- 🐛 バグの早期発見
- 📚 実行可能なドキュメント
- 🤝 チーム開発の基盤

**重要なのは、完璧を目指さず、段階的に改善し続けること**です。

## コード例

完全なコード例は以下を参照：
- [GitHub Repository](https://github.com/yourusername/tfdrift-falco)
- [詳細記事](./test-coverage-improvement-journey.md)

---

**執筆**: 2025年11月18日
**Tags**: #golang #testing #cicd #devops
