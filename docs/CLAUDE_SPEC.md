# CLAUDE_SPEC.md

## 📘 プロジェクト概要
本プロジェクトは、GCP（Google Cloud Platform）上で動作する **ログ監視・解析システム** です。  
Cloud Logging Sink → Pub/Sub → Cloud Run（Log Analysis Server）を経由してログを収集・解析し、  
API Server や Slack 通知、Claude Code を用いた自動分析機能を提供します。

---

## 🏗️ システム構成概要

### 全体構成

monorepo/
├── api-server/ # REST API サーバー (Go + Echo)
├── log-analysis-server/ # Pub/Sub 経由でログを受信・分析
├── proto/ # 将来的なgRPC/イベントスキーマ拡張用（現時点では未使用）
├── terraform/ # GCP インフラ構成 (Cloud Run, Pub/Sub, IAM, Logging Sink)
└── docs/ # ドキュメント類（本ファイル含む）



### GCP構成
- **Cloud Logging Sink**  
  → 特定ログ（例: エラーログ、警告ログ）を **Pub/Sub Topic** にエクスポート
- **Pub/Sub Topic**
  → Cloud Run（log-analysis-server）をサブスクライブ
- **Cloud Run: log-analysis-server**
  → 受信したログを解析し、Slack に通知、または Claude Code に渡して構造解析
- **Cloud Run: api-server**
  → 管理UI・設定APIを提供（Slack通知設定など）

---

## 🧩 コンポーネント詳細

### 1️⃣ `api-server`
- **役割**: 通知設定や解析結果を閲覧できる REST API
- **主なエンドポイント**
  | メソッド | パス | 機能 |
  |-----------|------|------|
  | GET | `/healthz` | ヘルスチェック |
  | GET | `/alerts` | 通知履歴取得 |
  | POST | `/settings/slack` | Slack Webhook URL設定 |
  | GET | `/analysis/:id` | Claude解析結果の取得 |
- **使用技術**
  - Go (Echo)
  - Cloud Run (containerized)
  - Firestore / Cloud Storage（設定保存・ログ一時保存）
- **構成例**
api-server/
├── main.go
├── handler/
│ ├── alert_handler.go
│ ├── health_handler.go
│ └── settings_handler.go
├── service/
│ ├── slack_service.go
│ └── claude_service.go
└── go.mod


---

### 2️⃣ `log-analysis-server`
- **役割**: Pub/Sub 経由でログを受信・分類・通知
- **主な処理**
1. Pub/Sub Push エンドポイントでログ受信
2. Cloud Logging の JSON 構造を解析
3. 重大度 (severity) に応じて Slack 通知 or Claude Code 解析
4. 解析結果を Firestore または GCS に保存
- **使用技術**
- Go (net/http)
- Pub/Sub Push Subscription
- Cloud Run
- Slack Webhook 通知
- Claude Code API 呼び出し（例: `/analyze` endpoint）
- **構成例**
log-analysis-server/
├── main.go
├── handler/
│ ├── pubsub_handler.go
│ └── claude_handler.go
├── service/
│ ├── log_parser.go
│ ├── slack_notifier.go
│ └── claude_client.go
└── go.mod


---

### 3️⃣ Claude Code 連携要件
- **目的**: ログ内容をClaudeで構文解析・根本原因特定
- **呼び出し例**
```bash
POST https://api.claude.ai/v1/analyze
Authorization: Bearer $CLAUDE_API_KEY
Body:
{
  "input": "Error: database connection timeout at api-server/main.go:42",
  "context": "log-analysis"
}
```

結果処理

ClaudeがJSONで返した要約・原因分析をGCSに保存

api-serverから取得・閲覧可能

# 4️. Slack 通知仕様

通知メッセージ例:
```
[ERROR] APIサーバーで異常を検出しました
File: api-server/main.go:42
Message: database connection timeout
詳細: https://console.cloud.google.com/logs/query;...
```
通知対象:

severity >= ERROR

Claude Code 解析が完了した場合、そのサマリも追記

# Terraform 構成

terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── cloud_run/
│   ├── api-server.tf
│   └── log-analysis-server.tf
├── pubsub/
│   ├── topic.tf
│   ├── subscription.tf
│   └── iam.tf
├── logging_sink/
│   └── sink.tf
└── iam/
    └── service_account.tf


主要リソース

google_logging_project_sink

google_pubsub_topic

google_pubsub_subscription

google_cloud_run_service

google_service_account

環境変数設定例

env {
  name  = "SLACK_WEBHOOK_URL"
  value = var.slack_webhook_url
}
env {
  name  = "CLAUDE_API_KEY"
  value = var.claude_api_key
}

# コード生成時のガイドライン

すべての Go サービスは main.go にエントリーポイントを持つ

各機能は handler/, service/ に分割

共通構造体（ログ・通知・設定）は internal/model/ に定義

Claude Code 解析は API 経由で非同期実行

Terraform ディレクトリは環境ごとに staging/, prod/ を切り替えられるようにする

🧾 今後の拡張

Pub/Sub メッセージスキーマを .proto 化（Event標準化）

Cloud Scheduler による定期分析

BigQueryへのログ転送

📄 管理メモ

Claude Code の仕様ファイル格納場所: docs/CLAUDE_SPEC.md

AI アシスタント（Claude, GitHub Copilot, etc.）はこのファイルを参照してコード生成を行う

重要キーワード: Cloud Run, Pub/Sub, Slack, Claude Code, Terraform, Go, Echo



---

この `CLAUDE_SPEC.md` をプロジェクトルートの  
`/docs/CLAUDE_SPEC.md` に配置すると、Claude Code / Copilot などが文脈を正確に把握しやすくなります。

---
