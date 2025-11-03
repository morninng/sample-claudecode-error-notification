# Quick Start - Local Testing

This is a quick reference for running log-analysis-server locally.

## Setup (First Time)

```bash
# 1. Navigate to the directory
cd log-analysis-server

# 2. Install dependencies
go mod download

# 3. Create environment file
cp .env.local.example .env.local

# 4. Edit .env.local with your actual credentials
# You need:
# - SLACK_BOT_TOKEN (from Slack App settings)
# - SLACK_CHANNEL (e.g., #alert)
# - ANTHROPIC_API_KEY (from Anthropic console)
# - GITHUB_TOKEN (from GitHub settings)
# - GITHUB_REPOSITORY (e.g., owner/repo-name)
```

## Run the Server

```bash
# Load environment variables
export $(cat .env.local | xargs)

# Start the server
go run main.go
```

Expected output:
```
Starting log-analysis-server on port 8080
```

## Test with Sample Message

Open a new terminal and run:

```bash
cd log-analysis-server
./test-pubsub.sh
```
