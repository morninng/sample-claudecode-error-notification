# Pub/Sub topic for error logs
resource "google_pubsub_topic" "error_logs" {
  name = "error-logs-topic"

  depends_on = [google_project_service.required_apis]
}
