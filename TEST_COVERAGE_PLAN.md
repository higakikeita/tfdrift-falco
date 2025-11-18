# TFDrift-Falco テストカバレッジ向上計画

**作成日**: 2025-11-18
**最終更新**: 2025-11-18
**目標**: テストカバレッジ 80%達成
**現状**: 36.9% (フェーズ1+2+3: 6/10パッケージ完了)

**進捗サマリー**:
- ✅ **フェーズ1完了** (pkg/types, pkg/config)
- ✅ **フェーズ2完了** (pkg/terraform/state, pkg/detector)
- ✅ **フェーズ3完了** (pkg/diff, pkg/metrics)
- ⏳ 残り4パッケージ: falco, notifier, importer, approval
- 📈 カバレッジ向上: 0% → 36.9% (+36.9%)

---

## 📊 現状分析

### パッケージ別コード量

| パッケージ | 行数 | 複雑度 | 優先度 |
|-----------|------|--------|--------|
| `pkg/diff` | 513 | 高 | 🔴 Critical |
| `pkg/falco` | 471 | 高 | 🔴 Critical |
| `pkg/detector` | 426 | 高 | 🔴 Critical |
| `pkg/notifier` | 236 | 中 | 🟠 High |
| `pkg/terraform/approval` | 235 | 中 | 🟠 High |
| `pkg/terraform/importer` | 205 | 中 | 🟠 High |
| `pkg/config` | 186 | 低 | 🟡 Medium |
| `pkg/terraform/state` | 179 | 中 | 🟠 High |
| `pkg/metrics` | 125 | 低 | 🟡 Medium |
| `pkg/types` | 48 | 低 | 🟢 Low |

**合計**: 2,624行

---

## 🎯 テスト戦略

### フェーズ1: 基盤テスト（Week 1）

#### 1. `pkg/types` - 最優先（48行） ✅ **完了**
**理由**: 他の全パッケージが依存する型定義

**テスト項目**:
- [x] Event構造体の生成と検証
- [x] UserIdentity構造体の生成と検証
- [x] DriftAlert構造体の生成と検証
- [x] UnmanagedResourceAlert構造体の生成と検証
- [x] JSONシリアライズ/デシリアライズ

**実際のカバレッジ**: [no statements] - 構造体のみのため正常
**完了日**: 2025-11-18
**テストファイル**: `pkg/types/types_test.go` (10テストケース、すべてパス)

#### 2. `pkg/config` - 第2優先（186行） ✅ **完了**
**理由**: アプリケーション起動時に必須、テストが書きやすい

**テスト項目**:
- [x] YAMLファイルの読み込み（正常系）
- [x] YAMLファイルの読み込み（異常系：ファイル不存在、パースエラー）
- [x] デフォルト値の設定
- [x] 設定値のバリデーション
- [x] 各Configフィールドの検証
- [x] Save機能のテスト

**実際のカバレッジ**: 90.9% (目標85%を達成！)
**完了日**: 2025-11-18
**テストファイル**: `pkg/config/config_test.go` (17テストケース、すべてパス)
**テストデータ**: `pkg/config/testdata/` (5つのYAMLファイル)

---

### フェーズ2: コアロジックテスト（Week 2）

#### 3. `pkg/terraform/state` - 第3優先（179行） ✅ **完了**
**理由**: ドリフト検出の基盤となるステート管理

**テスト項目**:
- [x] ローカルステートファイルの読み込み
- [x] S3バックエンドからの読み込み（未実装の確認）
- [x] リソースインデックスの構築
- [x] リソースIDの抽出ロジック
- [x] GetResource関数の動作確認
- [x] Refresh機能のテスト
- [x] スレッドセーフティのテスト
- [x] 複数インスタンスの処理

**実際のカバレッジ**: 100% (state.go) / 24.8% (パッケージ全体)
**完了日**: 2025-11-18
**テストファイル**: `pkg/terraform/state_test.go` (17テストケース、すべてパス)
**テストデータ**: `pkg/terraform/testdata/` (4つのtfstateファイル)

#### 4. `pkg/detector` - 第4優先（426行） ✅ **完了**
**理由**: ドリフト検出のコアロジック

**テスト項目**:
- [x] ドリフト検出ロジック（detectDrifts）
- [x] ドリフト検出ルールの適用（evaluateRules）
- [x] 重要度（Severity）の判定（getSeverity）
- [x] 複雑な値の処理
- [x] 各種データ型（Boolean, Numeric, Nil）
- [x] エッジケース（空リソース、空ルール、ケース感度）

**実際のカバレッジ**: 21.1% (コア関数に集中、外部依存は未テスト)
**完了日**: 2025-11-18
**テストファイル**: `pkg/detector/detector_test.go` (20テストケース、すべてパス)
**注**: New(), Start(), handleEvent()は外部依存が多いため未テスト

---

### フェーズ3: 統合機能テスト（Week 3）

#### 5. `pkg/diff` - 第5優先（513行） ✅ **完了**
**理由**: Diffフォーマッターは複雑だが、重要度は中程度

**テスト項目**:
- [x] Console形式の出力
- [x] UnifiedDiff形式の出力
- [x] Markdown形式の出力
- [x] JSON形式の出力
- [x] SideBySide形式の出力
- [x] Unmanaged Resource形式の出力
- [x] カラー出力の有効/無効
- [x] 値のフォーマット（Simple, Complex）
- [x] Terraform構文のフォーマット
- [x] 各種データ型の処理

**実際のカバレッジ**: 96.0% (目標65%を大幅に超える！)
**完了日**: 2025-11-18
**テストファイル**: `pkg/diff/formatter_test.go` (25テストケース、すべてパス)

#### 6. `pkg/falco` - 第6優先（471行）
**理由**: 外部依存（Falco gRPC）があるため、モック化が必要

**テスト項目**:
- [ ] gRPCクライアントの接続（モック）
- [ ] イベントストリームのサブスクライブ
- [ ] イベントのパース
- [ ] 再接続ロジック
- [ ] エラーハンドリング

**推定カバレッジ**: 60%以上（外部依存により制限）

#### 7. `pkg/notifier` - 第7優先（236行）
**理由**: 通知機能、外部API依存

**テスト項目**:
- [ ] Slack通知（モック）
- [ ] Discord通知（モック）
- [ ] Webhook通知（モック）
- [ ] 通知フォーマット
- [ ] リトライロジック
- [ ] 各通知チャネルの有効化/無効化

**推定カバレッジ**: 70%以上

---

### フェーズ4: 補助機能テスト（Week 4）

#### 8. `pkg/terraform/importer` - 第8優先（205行）
**理由**: 自動インポート機能、テストが複雑

**テスト項目**:
- [ ] インポートコマンドの生成
- [ ] リソース名の生成ロジック
- [ ] Terraformコードの生成
- [ ] インポートの実行（モック）
- [ ] バッチインポート

**推定カバレッジ**: 65%以上

#### 9. `pkg/terraform/approval` - 第9優先（235行）
**理由**: 承認ワークフロー、インタラクティブなUIを含む

**テスト項目**:
- [ ] 承認リクエストの生成
- [ ] 承認/拒否の処理
- [ ] 承認サマリーのフォーマット
- [ ] 非インタラクティブモードのテスト

**推定カバレッジ**: 60%以上

#### 10. `pkg/metrics` - 第10優先（125行） ✅ **完了**
**理由**: メトリクス収集、優先度低

**テスト項目**:
- [x] Prometheusメトリクスの登録
- [x] DriftAlertsカウンターの記録
- [x] UnresolvedAlertsゲージの増減
- [x] EventsProcessedカウンターの記録
- [x] DetectionLatencyヒストグラムの記録
- [x] ComponentStatusの設定
- [x] 各種エッジケース（空文字列、ゼロ値）

**実際のカバレッジ**: 81.2% (目標70%を超える！)
**完了日**: 2025-11-18
**テストファイル**: `pkg/metrics/prometheus_test.go` (17テストケース、すべてパス)
**注**: StartMetricsServer()はHTTPサーバー起動のため未テスト

---

## 📈 達成目標

### 週次目標

| 週 | フェーズ | パッケージ | 累積カバレッジ |
|----|---------|-----------|---------------|
| Week 1 | フェーズ1 | types, config | ~15% |
| Week 2 | フェーズ2 | terraform/state, detector | ~45% |
| Week 3 | フェーズ3 | diff, falco, notifier | ~70% |
| Week 4 | フェーズ4 | terraform/importer, approval, metrics | **~80%** |

### 最終目標カバレッジ

```
pkg/types            : 95%
pkg/config           : 85%
pkg/terraform/state  : 80%
pkg/detector         : 75%
pkg/diff             : 70%
pkg/falco            : 65%
pkg/notifier         : 75%
pkg/terraform/importer : 70%
pkg/terraform/approval : 65%
pkg/metrics          : 75%
---
Total                : 80%+
```

---

## 🛠️ テスト実装ガイドライン

### テストファイル命名規則
```
<package>_test.go        # 基本的なユニットテスト
<package>_integration_test.go  # 統合テスト
```

### テストヘルパー
```go
// testutil パッケージの作成を推奨
pkg/testutil/
  ├── fixtures.go      # テストデータ
  ├── mock_config.go   # モック設定
  ├── mock_falco.go    # モックFalcoクライアント
  └── assertions.go    # カスタムアサーション
```

### モックライブラリ
- **gomock**: インターフェースのモック生成
- **testify**: アサーション・スイート

### テストデータ
```
testdata/
  ├── terraform_states/
  │   ├── simple.tfstate
  │   ├── complex.tfstate
  │   └── invalid.tfstate
  ├── configs/
  │   ├── valid.yaml
  │   ├── minimal.yaml
  │   └── invalid.yaml
  └── events/
      ├── ec2_termination.json
      ├── iam_policy_change.json
      └── unmanaged_resource.json
```

---

## 🚀 開始方法

### Step 1: テストインフラ準備
```bash
cd ~/tfdrift-falco

# テスト用ディレクトリ作成
mkdir -p testdata/{terraform_states,configs,events}
mkdir -p pkg/testutil

# 必要なライブラリインストール
go get github.com/stretchr/testify
go get github.com/golang/mock/gomock
```

### Step 2: 最初のテストを書く（pkg/types）
```bash
# テストファイル作成
touch pkg/types/types_test.go

# テスト実行
go test -v ./pkg/types

# カバレッジ確認
go test -cover ./pkg/types
```

### Step 3: CI統合
```yaml
# .github/workflows/test.yml に追加
- name: Run tests with coverage
  run: |
    go test -v -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out

- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 80%"
      exit 1
    fi
```

---

## 📝 次のアクション

**今すぐ実行**:
1. `pkg/types/types_test.go` を作成
2. 基本的な構造体テストを実装
3. `go test -cover ./pkg/types` で確認

**この週末**:
- フェーズ1完了（types, config）
- カバレッジ15%達成

**来週**:
- フェーズ2着手（terraform/state, detector）
- カバレッジ45%を目指す

---

**参考リンク**:
- [Go Testing Best Practices](https://golang.org/doc/tutorial/add-a-test)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [gomock Documentation](https://github.com/golang/mock)
