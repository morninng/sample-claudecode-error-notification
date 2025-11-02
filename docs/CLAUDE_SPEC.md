
# プロジェクト概要
本プロジェクトは、実証実験用のサンプルです。
## 検証したい機能
(1) api serverでエラーがきたときに、slackにエラーを通知する
(2) エラー内容と、ソースコードをclaude codeに渡し、claude codeでエラーの原因を解析する
(3) 解析したエラー原因を、(1)で通知したslackのスレッドに、追加する。

## 動作

 GCP（Google Cloud Platform）上で動作する
 - api serverが、api機能を提供するが、サンプルとしてエラーを出したいだけなので、最低限の機能としてhello worldおよび、queryの値によってerrorをだすだけの機能でよい。
 - api serverはcloud runで実装
 - api serverのログをcloud logにて取得する。
 - cloud logで取得したログのうち、errorのみをCloud Logging Sinkを用いて、pubsubで、log analysis serverに通知する
 - log analysis serverは、ログをslackで通知し、claude codeへの解析を行い、再度解析結果をslackに通知する
  - log analysis serverはin はapi serverのみ、
  - api serverはpublic accessibleとする。 


 - cloud run(api server) -> Cloud Logging Sink → Pub/Sub → Cloud Run（Log Analysis Server）を経由してログを収集・解析する

---

## 🏗️ システム構成概要

### 全体構成

monorepo/
├── api-server/ # REST API サーバー (Go + Echo)
├── log-analysis-server/ # Pub/Sub 経由でログを受信・分析
├── terraform/ # GCP インフラ構成 (Cloud Run, Pub/Sub, IAM, Logging Sink)
└── docs/ # ドキュメント類（本ファイル含む）



### GCP構成
- **Cloud Run: api-server**
  → API server
- **Cloud Logging Sink**  
  → 特定ログ（severity >= ERROR）のみをfilterし **Pub/Sub Topic** にエクスポート
- **Pub/Sub Topic**
  → Cloud Run（log-analysis-server）がサブスクライブ
- **Cloud Run: log-analysis-server**
  → 受信したログを解析し、Slack に通知、または Claude Code に渡して構造解析
- **Region**
  asia-northeast1

---

## 🧩 コンポーネント詳細

### 1️⃣ `api-server`
- **役割**: シンプルな REST API。hello worldのみ
- **主なエンドポイント**
  | メソッド | パス | 機能 |
  |-----------|------|------|
  | GET | `/hello` | シンプルなhello worldを返すだけのapi |
- **動作詳細**
 /hello?messag=error で "error occured with invalid query message" status code 400を返す
 それ以外では常に hello worldをstatus code 200で返す

- **使用技術**
  - Go (Echo)
  - Cloud Run (containerized)
- **構成例**
api-server/
├── main.go
├── handler/
│ ├── hello.go
└── go.mod


---

### 2️⃣ `log-analysis-server`
- **役割**: Pub/Sub 経由でログを受信・通知
- **処理**
1. Cloud LoggingからPub/Sub Push エンドポイントでログ受信
2. 受信したデータをJSON 構造を解析
3. error logを受信し、環境変数SLACK_BOT_TOKENを用いてSlackのSLACK_CHANNELに通知を行い、thread_tsを取得
  -     Slack Notifier は chat.postMessage
    https://docs.slack.dev/reference/methods/chat.postMessage/
   thread_tsに関してはこちらを参照

4. github上のコードを環境変数のGITHUB_REPOSITORYとGITHUB_TOKENを用いて取得する。
5. github上のコードと、error logを用いてpromptを生成し、claude code apiに送信
　https://docs.claude.com/en/api/messages
　claude codeにはこちらのapiにあるmessage機能を用いてください

  githubのcodeに関しては全コードを送信してください
　
5. claude codeから取得したデータを 3.のthread_tsに対して、slack通知をスレッドとして付与する。

- **環境変数**
 - SLACK_BOT_TOKEN
 - SLACK_CHANNEL
 - GITHUB_REPOSITORY
 - GITHUB_TOKEN

- **使用技術**
- Go (net/http)
- Pub/Sub Push Subscription
- Cloud Run
- Slack Webhook 通知
- Claude Code API 呼び出し

- **構成例**
log-analysis-server/
├── main.go
├── handler/
│ └── log.go
└── go.mod



---


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
  name  = "SLACK_BOT_TOKEN"
  value = var.slack_bot_token
}
env {
  name  = "ANTHROPIC_API_KEY"
  value = var.anthropic_api_key
}
env {
  name  = "GITHUB_TOKEN"
  value = var.github_token
}

env {
  name  = "GITHUB_REPOSITORY"
  value = var.github_repository
}


env {
  name  = "SLACK_CHANNEL"
  value = "#alert"
}