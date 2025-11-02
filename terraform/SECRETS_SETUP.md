# GCP Secret Manager Setup

This document explains how to create the required secrets in GCP Secret Manager before deploying the infrastructure with Terraform.

## Prerequisites

- Google Cloud SDK (`gcloud`) installed and configured
- Appropriate permissions to create secrets in your GCP project
- The secret values ready (Slack Bot Token, Anthropic API Key, GitHub Token)

## Required Secrets

The following secrets must be created in GCP Secret Manager:

1. **slack-bot-token**: Slack Bot OAuth Token (starts with `xoxb-`)
2. **anthropic-api-key**: Anthropic API Key for Claude (starts with `sk-ant-`)
3. **github-token**: GitHub Personal Access Token (starts with `ghp_` or `github_pat_`)

## Creating Secrets

### Method 1: Using gcloud CLI

Replace `YOUR_PROJECT_ID` with your actual GCP project ID and the placeholder values with your actual secret values.

```bash
# Set your project ID
PROJECT_ID="your-gcp-project-id"

# Enable Secret Manager API (if not already enabled)
gcloud services enable secretmanager.googleapis.com --project=$PROJECT_ID

# Create Slack Bot Token secret
echo -n "xoxb-your-actual-slack-bot-token" | \
  gcloud secrets create slack-bot-token \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Create Anthropic API Key secret
echo -n "sk-ant-your-actual-anthropic-api-key" | \
  gcloud secrets create anthropic-api-key \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID

# Create GitHub Token secret
echo -n "ghp_your-actual-github-token" | \
  gcloud secrets create github-token \
  --data-file=- \
  --replication-policy="automatic" \
  --project=$PROJECT_ID
```

### Method 2: Using GCP Console

1. Navigate to [Secret Manager](https://console.cloud.google.com/security/secret-manager) in the GCP Console
2. Click "CREATE SECRET"
3. For each secret:
   - **Name**: Use the exact names: `slack-bot-token`, `anthropic-api-key`, `github-token`
   - **Secret value**: Paste the actual secret value
   - **Regions**: Choose "Automatic" for replication
   - Click "CREATE SECRET"

## Verifying Secrets

After creating the secrets, verify they exist:

```bash
gcloud secrets list --project=$PROJECT_ID
```

You should see all three secrets listed:
- slack-bot-token
- anthropic-api-key
- github-token

## Updating Secret Values

If you need to update a secret value:

```bash
# Add a new version to the secret
echo -n "new-secret-value" | \
  gcloud secrets versions add SECRET_NAME \
  --data-file=- \
  --project=$PROJECT_ID
```

The Terraform configuration automatically uses the latest version of each secret.

## Security Best Practices

1. **Limit Access**: Grant Secret Manager access only to the service accounts that need it
2. **Audit Logs**: Enable audit logging for Secret Manager to track secret access
3. **Rotation**: Regularly rotate your secrets and add new versions
4. **Never Commit**: Never commit secret values to version control

## IAM Permissions

The Terraform configuration automatically grants the log-analysis-server service account the `roles/secretmanager.secretAccessor` role, which allows it to read secret values at runtime.

## Troubleshooting

If Cloud Run cannot access the secrets:

1. Verify secrets exist: `gcloud secrets list --project=$PROJECT_ID`
2. Check the service account has the correct permissions
3. Review Cloud Run logs for permission errors
4. Ensure the secret names in [secrets.tf](secrets.tf) match exactly with the created secrets
