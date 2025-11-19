# formatter.go リファクタリング計画

## 📊 現状分析

### ファイル情報
- **ファイル**: `pkg/diff/formatter.go`
- **行数**: 513行
- **関数数**: 18個
- **責務**: 5種類のフォーマット（Console, Markdown, JSON, UnifiedDiff, SideBySide）+ ヘルパー関数

### 関数一覧と分類

#### 1. Core/Constructor (1個)
- `NewFormatter` - コンストラクタ

#### 2. Format Methods - DriftAlert (5個)
- `FormatConsole` - コンソール出力（100行）
- `FormatMarkdown` - Markdown形式（53行）
- `FormatJSON` - JSON形式（36行）
- `FormatUnifiedDiff` - Unified diff形式（24行）
- `FormatSideBySide` - 2カラム比較（44行）

#### 3. Format Methods - UnmanagedResourceAlert (2個)
- `FormatUnmanagedResource` - コンソール出力（64行）
- `FormatUnmanagedResourceMarkdown` - Markdown形式（60行）

#### 4. Helper - Value Formatting (5個)
- `formatValueChange` - 値変更のフォーマット（28行）
- `formatValue` - 単一値のフォーマット（18行）
- `formatTerraformValue` - Terraform HCL形式（30行）
- `isComplexType` - 複雑型の判定（12行）
- `indentLines` - インデント処理（11行）

#### 5. Helper - Terraform Code (2個)
- `formatTerraformCode` - Terraformコード生成（15行）
- `formatTerraformResource` - リソースブロック生成（14行）
- `formatRecommendations` - 推奨アクション（15行）

#### 6. Helper - Display (2個)
- `color` - ANSI色付け（8行）
- `getSeverityColor` - severity色マッピング（16行）

## 🎯 分割案

### Option A: 責務ベース分割（推奨）

```
pkg/diff/
├── formatter.go          (80行) - インターフェース + コンストラクタ
├── console_formatter.go  (120行) - Console形式
├── markdown_formatter.go (130行) - Markdown形式
├── json_formatter.go     (50行)  - JSON形式
├── diff_formatter.go     (80行)  - UnifiedDiff + SideBySide
└── helpers.go            (100行) - 共通ヘルパー関数
```

#### ファイル詳細

**1. formatter.go** - インターフェースとコア
```go
package diff

type Formatter interface {
    FormatConsole(alert *types.DriftAlert) string
    FormatMarkdown(alert *types.DriftAlert) string
    FormatJSON(alert *types.DriftAlert) (string, error)
    FormatUnifiedDiff(alert *types.DriftAlert) string
    FormatSideBySide(alert *types.DriftAlert) string
    FormatUnmanagedResource(alert *types.UnmanagedResourceAlert) string
    FormatUnmanagedResourceMarkdown(alert *types.UnmanagedResourceAlert) string
}

type DiffFormatter struct {
    colorEnabled bool
}

func NewFormatter(colorEnabled bool) *DiffFormatter {
    return &DiffFormatter{colorEnabled: colorEnabled}
}

// ANSI color codes
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    // ...
)
```

**2. console_formatter.go** - Console出力
```go
package diff

func (f *DiffFormatter) FormatConsole(alert *types.DriftAlert) string {
    // 現在の FormatConsole 実装
}

func (f *DiffFormatter) FormatUnmanagedResource(alert *types.UnmanagedResourceAlert) string {
    // 現在の FormatUnmanagedResource 実装
}
```

**3. markdown_formatter.go** - Markdown出力
```go
package diff

func (f *DiffFormatter) FormatMarkdown(alert *types.DriftAlert) string {
    // 現在の FormatMarkdown 実装
}

func (f *DiffFormatter) FormatUnmanagedResourceMarkdown(alert *types.UnmanagedResourceAlert) string {
    // 現在の FormatUnmanagedResourceMarkdown 実装
}
```

**4. json_formatter.go** - JSON出力
```go
package diff

func (f *DiffFormatter) FormatJSON(alert *types.DriftAlert) (string, error) {
    // 現在の FormatJSON 実装
}
```

**5. diff_formatter.go** - Diff形式
```go
package diff

func (f *DiffFormatter) FormatUnifiedDiff(alert *types.DriftAlert) string {
    // 現在の FormatUnifiedDiff 実装
}

func (f *DiffFormatter) FormatSideBySide(alert *types.DriftAlert) string {
    // 現在の FormatSideBySide 実装
}
```

**6. helpers.go** - 共通ヘルパー
```go
package diff

// Value formatting
func (f *DiffFormatter) formatValueChange(oldValue, newValue interface{}) string
func (f *DiffFormatter) formatValue(value interface{}) string
func (f *DiffFormatter) formatTerraformValue(value interface{}) string
func (f *DiffFormatter) isComplexType(value interface{}) bool

// Terraform code generation
func (f *DiffFormatter) formatTerraformCode(alert *types.DriftAlert) string
func (f *DiffFormatter) formatTerraformResource(alert *types.DriftAlert, value interface{}) string
func (f *DiffFormatter) formatRecommendations(alert *types.DriftAlert) string

// Display helpers
func (f *DiffFormatter) color(colorCode, text string) string
func (f *DiffFormatter) getSeverityColor(severity string) string
func (f *DiffFormatter) indentLines(text string, spaces int, color string) string
```

### Option B: Alert Type別分割

```
pkg/diff/
├── formatter.go                    (60行) - インターフェース
├── drift_alert_formatter.go        (300行) - DriftAlert用
├── unmanaged_resource_formatter.go (150行) - UnmanagedResourceAlert用
└── helpers.go                      (100行) - 共通ヘルパー
```

### Option C: Format Type別分割

```
pkg/diff/
├── formatter.go      (80行) - インターフェース
├── console.go        (200行) - Console系
├── text.go           (150行) - Markdown + UnifiedDiff + SideBySide
├── json.go           (50行)  - JSON系
└── helpers.go        (100行) - ヘルパー
```

## ✅ 推奨: Option A（責務ベース分割）

### メリット

1. **単一責任の原則（SRP）**
   - 各ファイルが1つの出力形式に責任を持つ
   - 変更の影響範囲が明確

2. **保守性の向上**
   - Console形式を変更しても、Markdown形式に影響しない
   - 新しい形式追加が容易（新ファイル追加のみ）

3. **テストの整理**
   - テストも同じ構造で分割可能
   - `console_formatter_test.go`, `markdown_formatter_test.go`...

4. **並行開発の容易さ**
   - 複数人で異なる形式を同時に開発可能
   - マージコンフリクトの減少

5. **読みやすさ**
   - ファイルサイズが適切（80-130行）
   - 各ファイルの目的が明確

### デメリット

- ファイル数が増える（1 → 6ファイル）
  - **対策**: パッケージとしてまとまっているため、影響は小さい

## 🔧 実装手順

### Phase 1: 準備（テスト確認）
```bash
# 現在のテストが全てパスすることを確認
go test ./pkg/diff/... -v

# カバレッジ確認
go test ./pkg/diff/... -coverprofile=cover.out
go tool cover -func=cover.out
```

### Phase 2: helpers.go 作成
1. `helpers.go` を作成
2. 全ヘルパー関数を移動
3. テスト実行（パスすることを確認）

### Phase 3: 各フォーマッター分割
1. `console_formatter.go` 作成
   - `FormatConsole` と `FormatUnmanagedResource` を移動
   - テスト実行

2. `markdown_formatter.go` 作成
   - `FormatMarkdown` と `FormatUnmanagedResourceMarkdown` を移動
   - テスト実行

3. `json_formatter.go` 作成
   - `FormatJSON` を移動
   - テスト実行

4. `diff_formatter.go` 作成
   - `FormatUnifiedDiff` と `FormatSideBySide` を移動
   - テスト実行

### Phase 4: 元ファイルのクリーンアップ
1. `formatter.go` に残すのは：
   - インターフェース定義（追加）
   - `DiffFormatter` 構造体
   - `NewFormatter` コンストラクタ
   - ANSI color定数

2. 最終テスト
```bash
go test ./pkg/diff/... -v
go test ./... -coverprofile=coverage.out
```

### Phase 5: テストファイルの整理（オプション）
```bash
pkg/diff/
├── console_formatter_test.go
├── markdown_formatter_test.go
├── json_formatter_test.go
├── diff_formatter_test.go
└── helpers_test.go
```

## 📋 チェックリスト

### 分割前
- [ ] 現在のテストが全てパス
- [ ] カバレッジが98.2%であることを確認
- [ ] gitブランチ作成 `git checkout -b refactor/split-formatter`

### 分割中
- [ ] helpers.go 作成・移動
  - [ ] テストパス確認
- [ ] console_formatter.go 作成・移動
  - [ ] テストパス確認
- [ ] markdown_formatter.go 作成・移動
  - [ ] テストパス確認
- [ ] json_formatter.go 作成・移動
  - [ ] テストパス確認
- [ ] diff_formatter.go 作成・移動
  - [ ] テストパス確認
- [ ] formatter.go クリーンアップ
  - [ ] テストパス確認

### 分割後
- [ ] 全テストパス
- [ ] カバレッジ維持（98.2%）
- [ ] golangci-lint チェック
- [ ] コミット + PR作成

## 🎨 コード例

### 分割前（formatter.go - 513行）
```go
// 全てが1ファイルに
type DiffFormatter struct { ... }
func NewFormatter() { ... }
func FormatConsole() { ... }  // 100行
func FormatMarkdown() { ... } // 53行
func FormatJSON() { ... }     // 36行
// ... 全18関数
```

### 分割後（formatter.go - 80行）
```go
package diff

// Formatter defines the interface for formatting drift alerts
type Formatter interface {
    FormatConsole(alert *types.DriftAlert) string
    FormatMarkdown(alert *types.DriftAlert) string
    FormatJSON(alert *types.DriftAlert) (string, error)
    FormatUnifiedDiff(alert *types.DriftAlert) string
    FormatSideBySide(alert *types.DriftAlert) string
    FormatUnmanagedResource(alert *types.UnmanagedResourceAlert) string
    FormatUnmanagedResourceMarkdown(alert *types.UnmanagedResourceAlert) string
}

// DiffFormatter implements the Formatter interface
type DiffFormatter struct {
    colorEnabled bool
}

// NewFormatter creates a new diff formatter
func NewFormatter(colorEnabled bool) *DiffFormatter {
    return &DiffFormatter{colorEnabled: colorEnabled}
}

// ANSI color codes
const (
    ColorReset  = "\033[0m"
    ColorRed    = "\033[31m"
    ColorGreen  = "\033[32m"
    ColorYellow = "\033[33m"
    ColorBlue   = "\033[34m"
    ColorPurple = "\033[35m"
    ColorCyan   = "\033[36m"
    ColorGray   = "\033[37m"
    ColorBold   = "\033[1m"
)
```

### 新規ファイル例（console_formatter.go）
```go
package diff

import (
    "fmt"
    "strings"

    "github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// FormatConsole formats the drift for console output with colors
func (f *DiffFormatter) FormatConsole(alert *types.DriftAlert) string {
    var b strings.Builder

    // Header
    severityColor := f.getSeverityColor(alert.Severity)
    b.WriteString(f.color(severityColor, "━━━━━━━━━..."))

    // ... 現在の実装をそのまま

    return b.String()
}

// FormatUnmanagedResource formats an unmanaged resource alert for console
func (f *DiffFormatter) FormatUnmanagedResource(alert *types.UnmanagedResourceAlert) string {
    // ... 現在の実装をそのまま
}
```

## 📈 期待される効果

### Before
```
formatter.go: 513行, 18関数
- 複雑度: 高
- 保守性: 中
- 並行開発: 困難
```

### After
```
formatter.go:          80行, 3関数  ← インターフェース
console_formatter.go:  120行, 2関数
markdown_formatter.go: 130行, 2関数
json_formatter.go:     50行, 1関数
diff_formatter.go:     80行, 2関数
helpers.go:            100行, 10関数
---
合計: 560行（+47行、コメント増）, 20関数
```

**メリット**:
- ✅ 各ファイルが150行以下
- ✅ 単一責任の原則
- ✅ テスト整理が容易
- ✅ 保守性向上

## 🚀 次のアクション

1. **このrefactoring計画をレビュー**
   - 分割方針の確認
   - 懸念点のヒアリング

2. **実装開始**
   ```bash
   git checkout -b refactor/split-formatter
   ```

3. **Phase by Phaseで進行**
   - 各Phaseごとにテスト
   - 問題があれば即座にロールバック

4. **PR作成**
   - タイトル: `refactor: split formatter.go into smaller files`
   - 説明: この計画書をベースに

---

**質問や懸念点があれば、実装前に確認しましょう！**
