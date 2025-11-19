# テストカバレッジ80%達成からリファクタリングまで - TFDrift-Falcoプロジェクト

## 🎯 はじめに

本記事では、オープンソースプロジェクト「TFDrift-Falco」において、テストカバレッジを**36.9%から80.0%まで向上**させ、さらに**3つの主要ファイル（計1,410行）を責任ごとに分割するリファクタリング**を実施した経験を共有します。

単なる数値改善ではなく、**テストによる安全性を確保しながら、保守性の高いコードベースを構築する実践的なアプローチ**を紹介します。

### この記事で得られること

- ✅ テストカバレッジを段階的に向上させる具体的な手法
- ✅ テーブル駆動テストやモックを活用した効率的なテスト作成
- ✅ テストカバレッジ80%達成後のリファクタリング実践例
- ✅ golangci-lintを使った静的解析とコード品質向上
- ✅ Single Responsibility Principleに基づくファイル分割戦略

### TFDrift-Falcoとは？

Falcoのランタイムセキュリティ監視とTerraform状態管理を組み合わせた、リアルタイムなInfrastructure as Code（IaC）ドリフト検出ツールです。

- **リポジトリ**: https://github.com/higakikeita/tfdrift-falco
- **言語**: Go 1.21+
- **総コード行数**: 約2,600行
- **実施期間**: テストカバレッジ向上（1日）+ リファクタリング（1日）

## 📊 プロジェクトの初期状態

### スタート地点

```
開始時のカバレッジ: 36.9%
目標: 80.0%
期間: 約1日（実質作業時間）
```

### パッケージ別の初期状態

| パッケージ | 開始時 | 主な課題 |
|-----------|--------|---------|
| pkg/types | 0% | 型定義のみでテストなし |
| pkg/config | 約40% | バリデーションロジックが未テスト |
| pkg/detector | 約52% | コアロジックの一部のみカバー |
| pkg/diff | 約85% | フォーマッター関数が部分的 |
| pkg/metrics | 約75% | HTTPサーバー起動がテスト困難 |
| pkg/testutil | 0% | テストヘルパー自体が未テスト |

## 🎬 第1フェーズ: 基盤固め（36.9% → 60%台）

### 戦略: 依存関係の下から攻める

まず依存される側のパッケージ（`pkg/types`, `pkg/config`）から着手しました。

#### pkg/types - 型定義の完全テスト

最初に取り組んだのが型定義パッケージ。他の全パッケージが依存するため、ここの安定性が最重要でした。

```go
// pkg/types/types_test.go
func TestEvent_JSONSerialization(t *testing.T) {
    event := &Event{
        Provider:     "aws",
        ResourceType: "aws_instance",
        ResourceID:   "i-12345",
        EventName:    "ModifyInstanceAttribute",
        Changes: map[string]interface{}{
            "instance_type": "t3.large",
        },
    }

    // JSONシリアライズ
    data, err := json.Marshal(event)
    require.NoError(t, err)

    // デシリアライズ
    var decoded Event
    err = json.Unmarshal(data, &decoded)
    require.NoError(t, err)

    assert.Equal(t, event.ResourceID, decoded.ResourceID)
}
```

**成果**: pkg/types を 0% → ほぼ100% に（型定義なので実行文が少ない）

#### pkg/config - バリデーションロジックの網羅

設定ファイルのバリデーションは、本番環境でのトラブルを未然に防ぐ重要な部分です。

```go
func TestValidate_NoProviderEnabled(t *testing.T) {
    cfg := &Config{
        Providers: ProvidersConfig{
            AWS: AWSConfig{Enabled: false},
            GCP: GCPConfig{Enabled: false},
        },
        Falco: FalcoConfig{Enabled: true},
    }

    err := cfg.Validate()
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "at least one provider must be enabled")
}
```

**学び**: エラーケースを網羅することで、実際に役立つテストになる

## 🚀 第2フェーズ: コアロジックの強化（60%台 → 70%台）

### pkg/detector - ドリフト検出エンジン

プロジェクトの心臓部であるdetectorパッケージ。複雑なイベント処理とゴルーチンの制御が課題でした。

#### 課題1: 非同期処理のテスト

```go
func TestProcessEvents_MultipleEvents(t *testing.T) {
    detector := &Detector{
        cfg:          cfg,
        stateManager: stateManager,
        formatter:    formatter,
        eventCh:      make(chan types.Event, 10),
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // ゴルーチンでイベント処理を開始
    go detector.processEvents(ctx)

    // 複数イベントを送信
    for i := 0; i < 3; i++ {
        event := types.Event{
            ResourceID: "i-0cea65ac652556767",
            // ...
        }
        detector.eventCh <- event
    }

    // 処理完了を待機
    time.Sleep(200 * time.Millisecond)

    cancel()
    time.Sleep(50 * time.Millisecond)

    assert.True(t, true) // パニックしなければ成功
}
```

**ポイント**:
- contextによる適切なキャンセル処理
- time.Sleepは最小限に（将来的にはsync.WaitGroupやchannelで改善）

#### 課題2: Terraform状態との整合性

実際のTerraform状態ファイルを使ったテストで、リアルな条件を再現：

```go
func TestHandleEvent_WithTerraformState(t *testing.T) {
    stateConfig := config.TerraformStateConfig{
        Backend:   "local",
        LocalPath: "testdata/terraform.tfstate",
    }
    stateManager, err := terraform.NewStateManager(stateConfig)
    require.NoError(t, err)

    err = stateManager.Load(context.Background())
    require.NoError(t, err)

    // 実際の状態ファイルに存在するリソースIDを使用
    event := types.Event{
        ResourceID: "i-0cea65ac652556767", // ← 実際に存在
        Changes: map[string]interface{}{
            "instance_type": "t3.large",
        },
    }

    detector.handleEvent(event)
    // ドリフト検出の検証
}
```

**成果**: pkg/detector を 52% → 89.2% に

## 🎨 第3フェーズ: 細部の磨き上げ（70%台 → 80.0%）

### pkg/metrics - 100%達成の挑戦

最も印象的だったのが、metricsパッケージの100%達成でした。

#### HTTPサーバーのテスト

Prometheusメトリクスを公開するHTTPサーバーのテスト：

```go
func TestStartMetricsServer(t *testing.T) {
    addr := "localhost:19090"

    // サーバーをバックグラウンドで起動
    go func() {
        _ = StartMetricsServer(addr)
    }()

    // 起動を待機
    time.Sleep(100 * time.Millisecond)

    // メトリクスエンドポイントにリクエスト
    resp, err := http.Get("http://" + addr + "/metrics")
    if err != nil {
        t.Logf("Could not connect (port conflict): %v", err)
        return
    }
    defer resp.Body.Close()

    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

**工夫点**:
- ポート衝突を考慮したエラーハンドリング
- 実際のHTTPリクエストで動作を検証

### pkg/testutil - テストヘルパーのテスト

ironicですが、テストを助けるコード自体もテストが必要です。

```go
func TestAssertJSONEqual(t *testing.T) {
    tests := []struct {
        name     string
        expected string
        actual   string
    }{
        {
            name:     "Equal JSON objects",
            expected: `{"name": "test", "value": 123}`,
            actual:   `{"value": 123, "name": "test"}`,
        },
        {
            name:     "Equal arrays",
            expected: `[1, 2, 3]`,
            actual:   `[1, 2, 3]`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            AssertJSONEqual(t, tt.expected, tt.actual)
        })
    }
}
```

**成果**: pkg/testutil を 28.3% → 56.1% に（+27.8%）

### 最後の0.1%を追うドラマ

79.9%から80.0%への最後の0.1%は、意外にも`pkg/diff`の小さな改善でした：

```go
func TestFormatSideBySide_ComplexValues(t *testing.T) {
    formatter := NewFormatter(false)

    alert := &types.DriftAlert{
        OldValue: map[string]interface{}{
            "key1": "value1",
            "key2": "value2",
        },
        NewValue: map[string]interface{}{
            "key1": "changed",
            "key3": "new",
        },
    }

    result := formatter.FormatSideBySide(alert)

    assert.Contains(t, result, "Terraform State")
    assert.Contains(t, result, "Actual Configuration")
    assert.Contains(t, result, "│")
}
```

複雑なデータ構造のフォーマットテストを追加して、ついに80.0%達成！

## 📈 最終結果

### 数値の推移

```
36.9% (開始)
  ↓ フェーズ1: 基盤固め
60.5% (+23.6%)
  ↓ フェーズ2: コアロジック
73.7% (+13.2%)
  ↓ フェーズ3: 最後の磨き
80.0% (+6.3%) ✨
```

### パッケージ別最終結果

| パッケージ | 開始 | 最終 | 改善幅 | 評価 |
|-----------|------|------|--------|------|
| **pkg/metrics** | 75% | **100.0%** | +25.0% | 🏆 |
| pkg/diff | 85% | 98.2% | +13.2% | ⭐ |
| pkg/terraform | - | 97.6% | - | ⭐ |
| pkg/notifier | - | 95.5% | - | ⭐ |
| pkg/config | 40% | 93.9% | +53.9% | ⭐ |
| pkg/detector | 52% | 91.0% | +39.0% | ⭐ |
| pkg/falco | - | 79.3% | - | ✓ |
| pkg/testutil | 0% | 56.1% | +56.1% | ✓ |
| cmd/tfdrift | - | 55.6% | - | ✓ |

### コード追加量

- **新規テストファイル**: 2個
  - `pkg/testutil/assertions_test.go` (225行)
  - `pkg/testutil/fixtures_test.go` (136行)
- **既存テスト強化**: 7ファイル
- **合計追加行数**: 929行

## 💡 学んだこと

### 1. テーブル駆動テストの威力

Go言語のテーブル駆動テスト（Table-Driven Tests）は、効率的にカバレッジを上げる強力な手法でした：

```go
func TestExtractChanges(t *testing.T) {
    tests := []struct {
        name      string
        eventName string
        fields    map[string]string
        wantKeys  []string
    }{
        {
            name:      "AttachUserPolicy",
            eventName: "AttachUserPolicy",
            fields: map[string]string{
                "ct.request.username":  "john",
                "ct.request.policyarn": "arn:aws:iam::123:policy/Admin",
            },
            wantKeys: []string{"user_name", "policy_arn"},
        },
        {
            name:      "CreateRole",
            eventName: "CreateRole",
            fields: map[string]string{
                "ct.request.rolename": "new-role",
                "ct.request.assumerolePolicyDocument": "{}",
            },
            wantKeys: []string{"role_name", "assume_role_policy"},
        },
        // 12個以上のテストケース...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := extractChanges(tt.eventName, tt.fields)

            for _, key := range tt.wantKeys {
                assert.Contains(t, result, key)
            }
        })
    }
}
```

**メリット**:
- 1つのテスト関数で複数のシナリオをカバー
- 新しいケース追加が容易
- 可読性が高い

### 2. テストしやすいコード設計

カバレッジ向上の過程で、「テストしにくいコード」に直面しました：

#### Before（テストしにくい）
```go
func Start() error {
    // グローバル変数に直接アクセス
    client = grpc.Dial(GlobalConfig.Host)
    // ...
}
```

#### After（テストしやすい）
```go
func (s *Subscriber) Start(ctx context.Context, eventCh chan<- types.Event) error {
    // 依存性を注入
    config := s.cfg
    // contextで制御可能
    // channelでテスト可能
}
```

**原則**:
- 依存性の注入（DI）
- contextによる制御
- channelによる通信
- グローバル変数の回避

### 3. 100%を目指さない勇気

以下は意図的にテストしませんでした：

```go
// cmd/tfdrift/main.go
func main() {
    // os.Exit()を呼ぶため、テスト困難
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

// pkg/falco/subscriber.go
func (s *Subscriber) Start(ctx context.Context) error {
    // 実際のgRPCサーバーが必要
    c, err := client.NewForConfig(ctx, clientConfig)
    // ...
}
```

**判断基準**:
- テストコストが高い（モックサーバー構築など）
- ビジネスロジックでない（main関数など）
- 外部依存が強い（実ネットワーク通信など）

**80%で十分な理由**:
- 業界標準は70-85%
- コアロジックは90%以上カバー済み
- 残り20%は上記の「テスト困難」な部分

### 4. テストは開発速度を上げる

当初「テスト書くと遅くなる」と思っていましたが、実際は逆でした：

**Before テスト不足の時**:
```
実装 → 手動テスト → バグ発見 → 修正 → 手動テスト...
（無限ループ）
```

**After テスト充実後**:
```
実装 → go test → バグ即座に発見 → 修正 → go test ✓
（数秒で完結）
```

**具体例**: detectorの修正
- テスト前: 1つの修正に30分（手動確認が必要）
- テスト後: 1つの修正に5分（`go test`で即座に検証）

## 🛠 使用したツールとテクニック

### 1. カバレッジ測定

```bash
# カバレッジ測定
go test ./... -coverprofile=coverage.out

# 詳細表示
go tool cover -func=coverage.out

# HTML表示
go tool cover -html=coverage.out
```

### 2. テストヘルパーの活用

```go
// pkg/testutil/fixtures.go
func CreateTestConfig() *config.Config {
    return &config.Config{
        Providers: ProvidersConfig{
            AWS: AWSConfig{
                Enabled: true,
                Regions: []string{"us-east-1"},
            },
        },
        Falco: FalcoConfig{
            Enabled:  true,
            Hostname: "localhost",
            Port:     5060,
        },
    }
}
```

テストデータ生成を共通化することで、テスト作成が劇的に楽になりました。

### 3. testifyライブラリ

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    // require: 失敗時に即座に停止
    require.NoError(t, err)

    // assert: 失敗しても続行
    assert.Equal(t, expected, actual)
    assert.Contains(t, result, "substring")
}
```

## 🚧 直面した課題と解決策

### 課題1: テストの実行順序による失敗

**症状**:
```bash
go test ./pkg/metrics  # ✓ 成功
go test ./...          # ✗ 失敗
```

**原因**: グローバルなPrometheusレジストリの重複登録

**解決策**:
```go
// pkg/metrics/prometheus_test.go
var testMetrics *Metrics

func init() {
    // テスト全体で1つのインスタンスを共有
    testMetrics = NewMetrics("tfdrift_test")
}

func TestNewMetrics(t *testing.T) {
    m := testMetrics  // 共有インスタンスを使用
    require.NotNil(t, m)
}
```

### 課題2: contextのデッドロック

**症状**: ゴルーチンのテストでタイムアウト

**原因**: contextのキャンセルタイミングの問題

**解決策**:
```go
func TestProcessEvents(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()  // 必ずクリーンアップ

    done := make(chan bool)
    go func() {
        detector.processEvents(ctx)
        done <- true
    }()

    // イベント送信...

    cancel()  // 明示的にキャンセル

    select {
    case <-done:
        // 正常終了
    case <-time.After(1 * time.Second):
        t.Fatal("timeout")
    }
}
```

### 課題3: ファイルシステムのテスト

**問題**: 一時ファイルの管理が煩雑

**解決策**:
```go
func TestWithTempDir(t *testing.T) {
    tmpDir := t.TempDir()  // Go 1.15+の便利機能
    // テスト終了時に自動削除される

    testFile := filepath.Join(tmpDir, "test.txt")
    err := os.WriteFile(testFile, []byte("data"), 0644)
    require.NoError(t, err)
}
```

## 🔄 次のステップ: リファクタリング実施報告

80%達成後、コード品質をさらに向上させるため、計画的にリファクタリングを実施しました。

### リファクタリング前の状態

```bash
$ find . -name "*.go" -not -path "./vendor/*" | xargs wc -l | sort -rn | head -10

1669 ./pkg/detector/detector_test.go
 671 ./pkg/falco/subscriber_test.go
 513 ./pkg/diff/formatter.go          ← 18関数、責任が混在
 471 ./pkg/falco/subscriber.go        ← イベント処理が肥大化
 426 ./pkg/detector/detector.go       ← 複数の責任が混在
```

### 実施したリファクタリング

#### ✅ 1. golangci-lint による静的解析（完了）

```bash
golangci-lint run
```

**修正内容**:
- 未チェックエラー16件を修正（特に`defer c.Close()`のエラーハンドリング）
- 到達不能コード（unreachable code）の削除
- セキュリティ警告（gosec）への対応
- 未使用変数の削除

**成果**: 重要な警告0件達成

#### ✅ 2. formatter.goの分割（513行 → 6ファイル）

```bash
pkg/diff/
├── formatter.go          (27行)   # インターフェース定義のみ
├── helpers.go           (177行)   # 共通ヘルパー関数
├── console_formatter.go (138行)   # コンソール出力
├── markdown_formatter.go (85行)   # Markdown形式
├── json_formatter.go     (43行)   # JSON形式
└── diff_formatter.go     (76行)   # Unified/Side-by-side Diff
```

**成果**:
- 513行のモノリシックファイルを、責任ごとに分割
- 各フォーマッターが独立してメンテナンス可能に

#### ✅ 3. detector.goの分割（426行 → 6ファイル）

```bash
pkg/detector/
├── detector.go        (87行)   # コア構造体と初期化
├── lifecycle.go       (68行)   # Start/Stop ライフサイクル管理
├── event_handler.go   (66行)   # イベント受信と処理
├── drift_detector.go  (88行)   # ドリフト検出アルゴリズム
├── alert_sender.go    (80行)   # アラート送信ロジック
└── auto_import.go     (76行)   # 自動インポート機能
```

**成果**:
- Single Responsibility Principleに従った設計
- 各機能が独立してテスト可能に

#### ✅ 4. subscriber.goの分割（473行 → 89行、81%削減）

```bash
pkg/falco/
├── subscriber.go        (89行)   # コアSubscriber構造体とStart()
├── event_parser.go      (81行)   # Falcoイベントのパース処理
├── resource_mapper.go   (56行)   # EventName → ResourceType マッピング
├── change_extractor.go (158行)   # CloudTrailイベントの変更抽出
└── helpers.go           (23行)   # ユーティリティ関数
```

**成果**:
- 473行 → 89行（81%削減）
- 各ファイルが単一の責任を持つ明確な構造
- イベントタイプ追加時の変更箇所が明確化

### リファクタリング後の状態

```bash
# 500行を超える実装ファイル
before: 3個 (formatter.go, subscriber.go, detector.go)
after:  0個

# 最大ファイルサイズ（実装ファイル）
before: 513行 (formatter.go)
after:  177行 (diff/helpers.go)

# golangci-lint 重要警告
before: 16件
after:  0件

# テストカバレッジ
before: 80.0%
after:  80.0% (維持)
```

### 定量的な改善指標

| 指標 | リファクタリング前 | リファクタリング後 | 改善率 |
|-----|-----------------|-----------------|-------|
| 500行超ファイル数 | 3個 | 0個 | 100%削減 |
| 最大ファイル行数 | 513行 | 177行 | 65%削減 |
| 平均ファイル行数 | 181行 | 92行 | 49%削減 |
| golangci-lint警告 | 16件 | 0件 | 100%解消 |

### 残りの課題

テストカバレッジ80%達成とコード品質改善により、プロジェクトの基盤は大幅に強化されました。今後の課題：

#### 優先度: 低

**1. 大きなテストファイルの分割**
- `pkg/detector/detector_test.go` (1669行)
- `pkg/falco/subscriber_test.go` (671行)

これらはテストケースが豊富な証拠でもあり、緊急性は低い

**2. エラーハンドリングの統一**
```go
// 現在: 標準のerror wrapping
return fmt.Errorf("failed to load config: %w", err)

// 将来: カスタムエラー型の導入検討
return errors.Wrap(err, "failed to load config")
```

**3. パッケージドキュメントの充実**
```go
// Package detector provides real-time infrastructure drift detection
// by comparing Falco runtime events with Terraform state.
package detector
```

### 達成した目標

✅ golangci-lint の重要警告 0件
✅ 500行超のファイル 0個（実装ファイル）
✅ テストカバレッジ 80.0% 維持
✅ Single Responsibility Principle に従った構造
✅ 継続的なテスト基盤の確立

## 🎓 まとめ

### テストカバレッジ向上のベストプラクティス

1. **依存関係の下から攻める**
   - 型定義 → 設定 → コアロジック の順

2. **テーブル駆動テストを活用**
   - 効率的にケースを網羅

3. **テストしやすいコード設計**
   - DI、context、channelの活用

4. **100%を目指さない**
   - 80%で実用上十分
   - コストと効果のバランス

5. **継続的な測定**
   - go test -cover を習慣化
   - CIで自動チェック

### 所感

このプロジェクトを通じて、**テストカバレッジ向上とリファクタリングは車の両輪**だと実感しました。

#### テストカバレッジ80%達成で得られたもの

- ✅ テストを書くことで、自分のコードの問題点に気づけた
- ✅ リファクタリングの安全性が確保された
- ✅ 新機能追加時の不安が大幅に減った
- ✅ コードレビュー時の品質チェックポイントが明確に

#### リファクタリング実施で得られたもの

- ✅ 500行超のファイルが0個に（実装ファイル）
- ✅ Single Responsibility Principleに従った明確な構造
- ✅ 機能追加時の変更箇所が特定しやすい設計
- ✅ golangci-lintによる静的解析で品質基準を維持

**重要な気づき**:
- 80%という数値は通過点であり、目的はより保守しやすく、信頼性の高いコードベースを作ること
- テストがあるからこそ、安心してリファクタリングできる
- リファクタリング後もテストカバレッジ80%を維持できたことで、品質の継続性を確認

この強固なテスト基盤とクリーンなコード構造があれば、今後の機能追加やコントリビューション受け入れにも自信を持って対応できます。

## 🔗 参考リンク

- **リポジトリ**: https://github.com/higakikeita/tfdrift-falco
- **Go Testing**: https://golang.org/doc/tutorial/add-a-test
- **Table-Driven Tests**: https://dave.cheney.net/2019/05/07/prefer-table-driven-tests
- **testify**: https://github.com/stretchr/testify

---

**この記事が、テストカバレッジ向上に取り組む皆さんの参考になれば幸いです！**

質問やフィードバックは、GitHubのIssueまたはDiscussionsでお待ちしています。

---

*本記事のテストコードは、すべてMITライセンスで公開されています。*
