# Cloud Logging Sink to export error logs to Pub/Sub
resource "google_logging_project_sink" "error_logs_sink" {
  name        = "error-logs-to-pubsub"
  destination = "pubsub.googleapis.com/projects/${var.project_id}/topics/${google_pubsub_topic.error_logs.name}"

  # Filter for ERROR severity logs
  filter = "severity >= ERROR"

  # Use unique writer identity
  unique_writer_identity = true

  depends_on = [
    google_pubsub_topic.error_logs,
    google_project_service.required_apis
  ]
}
