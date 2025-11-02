# Sample Claude Code Error Notification

This project is a sample implementation for demonstrating error notification and automated analysis using GCP, Slack, and Claude API.

## Features

1. API server that generates errors based on query parameters
2. Automatic error log detection via Cloud Logging
3. Slack notification when errors occur
4. Automated error analysis using Claude API
5. Analysis results posted to Slack thread

## Architecture

```
api-server (Cloud Run)
    ↓ logs
Cloud Logging
    ↓ (severity >= ERROR)
Cloud Logging Sink
    ↓
Pub/Sub Topic
    ↓
log-analysis-server (Cloud Run)
    ↓
1. Slack notification (get thread_ts)
2. Fetch GitHub repository code
3. Analyze with Claude API
4. Post analysis to Slack thread
```

## Prerequisites

- GCP Project with billing enabled
- Docker and Docker Compose
- Terraform >= 1.0
- gcloud CLI
- Slack Bot Token (with `chat:write` permission)
- Anthropic API Key
- GitHub Personal Access Token

## Project Structure

```
.
├── api-server/           # Simple API server (Go + Echo)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── log-analysis-server/  # Log analysis and notification (Go)
│   ├── main.go
│   ├── go.mod
│   └── Dockerfile
├── terraform/            # Infrastructure as Code
│   ├── main.tf
│   ├── variables.tf
│   ├── outputs.tf
│   ├── cloud_run/
│   ├── pubsub/
│   ├── logging_sink/
│   └── iam/
└── docs/                 # Documentation
```

## Setup Instructions

### 1. Set up Slack Bot

1. Go to [https://api.slack.com/apps](https://api.slack.com/apps)
2. Create a new app
3. Add Bot Token Scopes: `chat:write`, `chat:write.public`
4. Install app to workspace
5. Copy the Bot User OAuth Token

### 2. Set up GitHub Token

1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Generate new token with `repo` scope
3. Copy the token

### 3. Set up Anthropic API Key

1. Go to [https://console.anthropic.com/](https://console.anthropic.com/)
2. Create API key
3. Copy the key

### 4. Configure GCP

```bash
# Set your GCP project
export PROJECT_ID="your-gcp-project-id"
gcloud config set project $PROJECT_ID

# Enable required APIs
gcloud services enable run.googleapis.com
gcloud services enable logging.googleapis.com
gcloud services enable pubsub.googleapis.com
gcloud services enable artifactregistry.googleapis.com

# Create Artifact Registry repository
gcloud artifacts repositories create docker-repo \
    --repository-format=docker \
    --location=asia-northeast1 \
    --description="Docker repository"
```

### 5. Build and Push Docker Images

```bash
# Configure Docker for Artifact Registry
gcloud auth configure-docker asia-northeast1-docker.pkg.dev

# Build and push api-server
cd api-server
docker build -t asia-northeast1-docker.pkg.dev/$PROJECT_ID/docker-repo/api-server:latest .
docker push asia-northeast1-docker.pkg.dev/$PROJECT_ID/docker-repo/api-server:latest

# Build and push log-analysis-server
cd ../log-analysis-server
docker build -t asia-northeast1-docker.pkg.dev/$PROJECT_ID/docker-repo/log-analysis-server:latest .
docker push asia-northeast1-docker.pkg.dev/$PROJECT_ID/docker-repo/log-analysis-server:latest
```

### 6. Deploy with Terraform

```bash
cd terraform

# Create terraform.tfvars file
cat > terraform.tfvars <<EOF
project_id                 = "your-gcp-project-id"
region                     = "asia-northeast1"
slack_bot_token            = "xoxb-your-slack-bot-token"
slack_channel              = "#alert"
anthropic_api_key          = "sk-ant-your-anthropic-key"
github_token               = "ghp_your-github-token"
github_repository          = "owner/repo-name"
api_server_image           = "asia-northeast1-docker.pkg.dev/your-project/docker-repo/api-server:latest"
log_analysis_server_image  = "asia-northeast1-docker.pkg.dev/your-project/docker-repo/log-analysis-server:latest"
EOF

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the configuration
terraform apply
```

### 7. Test the System

```bash
# Get the API server URL
API_URL=$(terraform output -raw api_server_url)

# Test normal request
curl "$API_URL/hello"
# Output: hello world

# Trigger an error
curl "$API_URL/hello?message=error"
# Output: error occured with invalid query message
# This will:
# 1. Generate an ERROR log in Cloud Logging
# 2. Send notification to Slack
# 3. Analyze the error with Claude
# 4. Post analysis result to Slack thread
```

## API Endpoints

### api-server

- `GET /hello` - Returns "hello world"
- `GET /hello?message=error` - Returns 400 error and triggers error notification flow

## Environment Variables

### log-analysis-server

| Variable | Description |
|----------|-------------|
| `SLACK_BOT_TOKEN` | Slack Bot User OAuth Token |
| `SLACK_CHANNEL` | Slack channel for notifications (e.g., #alert) |
| `ANTHROPIC_API_KEY` | Anthropic API key for Claude |
| `GITHUB_TOKEN` | GitHub Personal Access Token |
| `GITHUB_REPOSITORY` | GitHub repository in format "owner/repo" |

## Cleanup

```bash
cd terraform
terraform destroy
```

## Troubleshooting

### Logs not appearing in Slack

1. Check Cloud Logging Sink is active
2. Verify Pub/Sub subscription is working
3. Check log-analysis-server logs:
   ```bash
   gcloud run services logs read log-analysis-server --region=asia-northeast1
   ```

### Claude API errors

1. Verify `ANTHROPIC_API_KEY` is correct
2. Check API quota and rate limits
3. Review log-analysis-server logs for detailed error messages

### GitHub code fetch errors

1. Verify `GITHUB_TOKEN` has correct permissions
2. Check `GITHUB_REPOSITORY` format is "owner/repo"
3. Ensure the repository is accessible with the token

## License

MIT
