# Enable Secret Manager API
resource "google_project_service" "secretmanager" {
  service            = "secretmanager.googleapis.com"
  disable_on_destroy = false
}

# Data sources to access secrets from Secret Manager
# These secrets should be created manually or through a separate process
# before running this Terraform configuration

data "google_secret_manager_secret_version" "github_token" {
  secret = "github-token"

  depends_on = [google_project_service.secretmanager]
}

data "google_secret_manager_secret_version" "anthropic_api_key" {
  secret = "anthropic-api-key"

  depends_on = [google_project_service.secretmanager]
}

data "google_secret_manager_secret_version" "slack_bot_token" {
  secret = "slack-bot-token"

  depends_on = [google_project_service.secretmanager]
}
