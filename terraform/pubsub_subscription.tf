# Push subscription to log analysis server
resource "google_pubsub_subscription" "error_logs_push" {
  name  = "error-logs-subscription"
  topic = google_pubsub_topic.error_logs.name

  push_config {
    push_endpoint = google_cloud_run_service.log_analysis_server.status[0].url

    oidc_token {
      service_account_email = google_service_account.log_analysis_server.email
    }
  }

  ack_deadline_seconds = 600

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  depends_on = [
    google_cloud_run_service.log_analysis_server,
    google_project_service.required_apis
  ]
}
