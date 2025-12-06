# 負荷テスト・性能テストフレームワーク

**目的**: 本番環境相当の負荷で TFDrift-Falco の性能を検証

---

## テストシナリオ

### シナリオ 1: 小規模環境
- CloudTrail イベント: 100/分
- Terraform リソース: 500個
- リージョン: 1個
- 実行時間: 1時間

### シナリオ 2: 中規模環境
- CloudTrail イベント: 1,000/分
- Terraform リソース: 5,000個
- リージョン: 3個
- 実行時間: 4時間

### シナリオ 3: 大規模環境
- CloudTrail イベント: 10,000/分
- Terraform リソース: 50,000個
- リージョン: 10個
- 実行時間: 8時間

---

## テスト実行方法

### 前提条件
```bash
# Go 1.21+
go version

# Docker
docker --version

# jq, curl
which jq curl
```

### 1. CloudTrail イベントシミュレーター

```bash
cd tests/load

# 小規模環境 (100 events/min)
go run cloudtrail_simulator.go \
  --rate 100 \
  --duration 1h \
  --output /tmp/simulated-cloudtrail-logs

# 中規模環境 (1,000 events/min)
go run cloudtrail_simulator.go \
  --rate 1000 \
  --duration 4h \
  --output /tmp/simulated-cloudtrail-logs

# 大規模環境 (10,000 events/min)
go run cloudtrail_simulator.go \
  --rate 10000 \
  --duration 8h \
  --output /tmp/simulated-cloudtrail-logs
```

### 2. Terraform State ジェネレーター

```bash
# 500 リソース
go run terraform_state_generator.go \
  --resources 500 \
  --output /tmp/terraform.tfstate

# 5,000 リソース
go run terraform_state_generator.go \
  --resources 5000 \
  --output /tmp/terraform.tfstate

# 50,000 リソース
go run terraform_state_generator.go \
  --resources 50000 \
  --output /tmp/terraform.tfstate
```

### 3. 負荷テスト実行

```bash
# テスト環境起動
docker-compose -f docker-compose.load-test.yml up -d

# 負荷テスト実行
go test -v -run=TestLoadScenario1 -timeout=2h

# メトリクス収集
./collect_metrics.sh

# テスト環境停止
docker-compose -f docker-compose.load-test.yml down
```

---

## 性能指標

### 測定項目

1. **イベント処理性能**
   - イベント受信から Drift 判定までの時間 (p50, p95, p99)
   - スループット (events/sec)

2. **リソース使用量**
   - CPU 使用率 (平均, 最大)
   - メモリ使用量 (平均, 最大, リーク有無)
   - ディスク I/O

3. **Terraform State 読み込み**
   - State 読み込み時間
   - State サイズとの相関

4. **エンドツーエンド遅延**
   - CloudTrail イベント発生 → Grafana 表示
   - CloudTrail 遅延を除いた処理時間

5. **エラー率**
   - イベント処理エラー率
   - Falco 接続エラー率

### 合格基準

| 指標 | 小規模 | 中規模 | 大規模 | 測定方法 |
|------|--------|--------|--------|----------|
| イベント処理時間 (p95) | < 100ms | < 500ms | < 1s | Prometheus metrics |
| メモリ使用量 | < 512MB | < 2GB | < 4GB | docker stats |
| CPU 使用率 (平均) | < 10% | < 30% | < 50% | docker stats |
| State 読み込み時間 | < 1s | < 5s | < 30s | ログ |
| エラー率 | < 0.1% | < 1% | < 5% | ログ集計 |

---

## ツール一覧

### 1. `cloudtrail_simulator.go`
CloudTrail イベントをシミュレート

### 2. `terraform_state_generator.go`
大規模な Terraform State を生成

### 3. `load_test.go`
統合負荷テスト

### 4. `collect_metrics.sh`
メトリクス収集スクリプト

### 5. `analyze_results.py`
結果分析・レポート生成

---

## 次のステップ

各ツールの詳細は個別のファイルを参照してください。
