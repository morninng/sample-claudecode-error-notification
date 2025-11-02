output "api_server_url" {
  description = "URL of the API server"
  value       = google_cloud_run_service.api_server.status[0].url
}

output "log_analysis_server_url" {
  description = "URL of the log analysis server (internal only)"
  value       = google_cloud_run_service.log_analysis_server.status[0].url
}

output "pubsub_topic_name" {
  description = "Name of the Pub/Sub topic"
  value       = google_pubsub_topic.error_logs.name
}

output "logging_sink_name" {
  description = "Name of the Cloud Logging sink"
  value       = google_logging_project_sink.error_logs_sink.name
}
