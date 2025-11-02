# Service account for API server
resource "google_service_account" "api_server" {
  account_id   = "api-server-sa"
  display_name = "API Server Service Account"
}

# Service account for log analysis server
resource "google_service_account" "log_analysis_server" {
  account_id   = "log-analysis-server-sa"
  display_name = "Log Analysis Server Service Account"
}

# Grant log writer role to API server
resource "google_project_iam_member" "api_server_log_writer" {
  project = var.project_id
  role    = "roles/logging.logWriter"
  member  = "serviceAccount:${google_service_account.api_server.email}"
}

# Grant Pub/Sub subscriber role to log analysis server
resource "google_project_iam_member" "log_analysis_pubsub_subscriber" {
  project = var.project_id
  role    = "roles/pubsub.subscriber"
  member  = "serviceAccount:${google_service_account.log_analysis_server.email}"
}
