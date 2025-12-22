# TFDrift-Falco 本番環境デプロイガイド

## 目次

1. [前提条件](#前提条件)
2. [クイックスタート](#クイックスタート)
3. [環境変数設定](#環境変数設定)
4. [Docker Composeデプロイ](#docker-composeデプロイ)
5. [Kubernetesデプロイ](#kubernetesデプロイ)
6. [トラブルシューティング](#トラブルシューティング)

---

## 前提条件

### 必須

- Docker 20.10+
- Docker Compose 2.0+
- AWS credentials（CloudTrail統合の場合）
- Terraform State（S3またはローカル）

### オプション

- Kubernetes cluster（K8sデプロイの場合）
- Helm 3.0+
- Slack Webhook URL（通知の場合）

---

## クイックスタート

### 1. リポジトリクローン

```bash
git clone https://github.com/higakikeita/tfdrift-falco.git
cd tfdrift-falco
```

### 2. 環境変数設定

```bash
# .envファイルを作成
cat > .env <<EOF
AWS_REGION=us-east-1
TERRAFORM_STATE_DIR=/path/to/terraform
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
TZ=Asia/Tokyo
EOF
```

### 3. 起動

```bash
docker-compose up -d
```

### 4. アクセス

- **Frontend UI**: http://localhost:3000
- **Backend API**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:9090

---

## 環境変数設定

### 必須変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `AWS_REGION` | AWS region | `us-east-1` |
| `TFDRIFT_FALCO_HOSTNAME` | Falco hostname | `falco` |
| `TFDRIFT_FALCO_PORT` | Falco gRPC port | `5060` |

### オプション変数

| 変数名 | 説明 | デフォルト値 |
|--------|------|-------------|
| `TERRAFORM_STATE_DIR` | Terraform state directory | `./terraform` |
| `SLACK_WEBHOOK_URL` | Slack webhook URL | - |
| `TZ` | Timezone | `UTC` |
| `VITE_API_BASE_URL` | Frontend API URL | `http://backend:8080/api/v1` |
| `VITE_WS_URL` | WebSocket URL | `ws://backend:8080/ws` |
| `VITE_SSE_URL` | SSE URL | `http://backend:8080/api/v1/stream` |

---

## Docker Composeデプロイ

### 基本デプロイ

```bash
# すべてのサービスを起動
docker-compose up -d

# ログを確認
docker-compose logs -f

# 状態確認
docker-compose ps
```

### サービス個別操作

```bash
# バックエンドのみ再起動
docker-compose restart backend

# フロントエンドのみビルド&起動
docker-compose up -d --build frontend

# 特定のサービスのログ
docker-compose logs -f backend
```

### スケーリング

```bash
# バックエンドを2インスタンスに
docker-compose up -d --scale backend=2
```

### 停止と削除

```bash
# サービス停止
docker-compose stop

# サービス停止&削除
docker-compose down

# ボリュームも含めて削除
docker-compose down -v
```

---

## Kubernetesデプロイ

### Helmチャート（準備中）

```bash
# Helm chartをインストール
helm install tfdrift ./charts/tfdrift-falco \
  --set backend.image.tag=latest \
  --set frontend.image.tag=latest \
  --set aws.region=us-east-1
```

### kubectl

```bash
# ConfigMapとSecretを作成
kubectl create configmap tfdrift-config --from-file=config.yaml

kubectl create secret generic aws-credentials \
  --from-file=credentials=$HOME/.aws/credentials \
  --from-file=config=$HOME/.aws/config

# デプロイ
kubectl apply -f k8s/

# 状態確認
kubectl get pods -l app=tfdrift
kubectl logs -f deployment/tfdrift-backend
```

---

## 本番環境最適化

### 1. リソース制限

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

### 2. ログローテーション

```yaml
services:
  backend:
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

### 3. 自動再起動

```yaml
services:
  backend:
    restart: unless-stopped
```

### 4. セキュリティ

- 非rootユーザーで実行（既に設定済み）
- read-onlyファイルシステム（必要に応じて）
- seccomp/AppArmorプロファイル適用

---

## トラブルシューティング

### フロントエンドが表示されない

**原因**: バックエンドAPIに接続できない

**解決策**:
```bash
# バックエンドの健全性確認
curl http://localhost:8080/health

# ネットワーク接続確認
docker-compose exec frontend ping backend

# 環境変数確認
docker-compose exec frontend env | grep VITE
```

### Falcoが起動しない

**原因**: 特権モードが必要

**解決策**:
```yaml
falco:
  privileged: true  # これが設定されているか確認
```

### Driftが検知されない

**原因1**: AWS credentials未設定

**解決策**:
```bash
# credentials確認
docker-compose exec backend aws sts get-caller-identity
```

**原因2**: Terraform Stateが見つからない

**解決策**:
```bash
# State pathを確認
docker-compose exec backend ls -la /terraform
```

### WebSocketが接続できない

**原因**: nginxプロキシ設定

**解決策**:
```nginx
# ui/nginx.conf を確認
location /ws {
    proxy_pass http://backend:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "Upgrade";
}
```

---

## ログ確認コマンド

```bash
# すべてのログ
docker-compose logs -f

# 最新100行
docker-compose logs --tail=100

# 特定のサービス
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f falco

# タイムスタンプ付き
docker-compose logs -f -t
```

---

## メトリクス監視

### Prometheus

```bash
# Prometheusエンドポイント
curl http://localhost:9090/metrics
```

### ヘルスチェック

```bash
# バックエンド
curl http://localhost:8080/health

# フロントエンド
curl http://localhost:3000/health
```

---

## バックアップと復元

### データボリュームバックアップ

```bash
# バックアップ
docker run --rm \
  -v tfdrift-data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/tfdrift-data-backup.tar.gz /data

# 復元
docker run --rm \
  -v tfdrift-data:/data \
  -v $(pwd):/backup \
  alpine tar xzf /backup/tfdrift-data-backup.tar.gz -C /
```

---

## アップグレード手順

### 1. バックアップ

```bash
# データをバックアップ
docker-compose down
# データボリュームバックアップ（上記参照）
```

### 2. 最新イメージ取得

```bash
# イメージ更新
docker-compose pull
```

### 3. 再起動

```bash
# 新バージョンで起動
docker-compose up -d
```

### 4. 動作確認

```bash
# ログ確認
docker-compose logs -f

# ヘルスチェック
curl http://localhost:8080/health
```

---

## セキュリティベストプラクティス

1. **環境変数でシークレット管理**
   ```bash
   # .envファイルを.gitignoreに追加
   echo ".env" >> .gitignore
   ```

2. **非rootユーザーで実行**（既に設定済み）

3. **最小権限原則**
   ```yaml
   volumes:
     - ${HOME}/.aws:/root/.aws:ro  # read-only
   ```

4. **定期的なアップデート**
   ```bash
   docker-compose pull
   docker-compose up -d
   ```

5. **ログ監視**
   ```bash
   # 異常検知のためのログ監視
   docker-compose logs -f | grep -i error
   ```

---

## サポート

問題が解決しない場合:

1. **GitHub Issues**: https://github.com/higakikeita/tfdrift-falco/issues
2. **ドキュメント**: https://higakikeita.github.io/tfdrift-falco/
3. **FAQ**: [docs/FAQ.md](FAQ.md)

---

## 関連ドキュメント

- [Getting Started](GETTING_STARTED.md)
- [API Documentation](api/rest-api.md)
- [Architecture](architecture.md)
- [Best Practices](best-practices.md)
