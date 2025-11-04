# Allow Pub/Sub to invoke log analysis server
resource "google_cloud_run_service_iam_member" "pubsub_invoker" {
  service  = google_cloud_run_service.log_analysis_server.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "serviceAccount:${google_service_account.log_analysis_server.email}"
}

# Allow logging sink to publish to Pub/Sub
resource "google_pubsub_topic_iam_member" "log_sink_publisher" {
  topic  = google_pubsub_topic.error_logs.name
  role   = "roles/pubsub.publisher"
  member = google_logging_project_sink.error_logs_sink.writer_identity
}
