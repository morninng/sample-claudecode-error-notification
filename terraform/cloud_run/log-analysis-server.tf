# Cloud Run service for log analysis server
resource "google_cloud_run_service" "log_analysis_server" {
  name     = "log-analysis-server"
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.log_analysis_server.email

      containers {
        image = var.log_analysis_server_image

        ports {
          container_port = 8080
        }

        env {
          name  = "SLACK_BOT_TOKEN"
          value_from {
            secret_key_ref {
              name = data.google_secret_manager_secret_version.slack_bot_token.secret
              key  = "latest"
            }
          }
        }

        env {
          name  = "SLACK_CHANNEL"
          value = var.slack_channel
        }

        env {
          name  = "ANTHROPIC_API_KEY"
          value_from {
            secret_key_ref {
              name = data.google_secret_manager_secret_version.anthropic_api_key.secret
              key  = "latest"
            }
          }
        }

        env {
          name  = "GITHUB_TOKEN"
          value_from {
            secret_key_ref {
              name = data.google_secret_manager_secret_version.github_token.secret
              key  = "latest"
            }
          }
        }

        env {
          name  = "GITHUB_REPOSITORY"
          value = var.github_repository
        }

        resources {
          limits = {
            cpu    = "2000m"
            memory = "1Gi"
          }
        }
      }

      timeout_seconds = 600
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "5"
        "autoscaling.knative.dev/minScale" = "0"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }

  depends_on = [
    google_project_service.required_apis,
    google_service_account.log_analysis_server
  ]
}

# Note: No public access - only Pub/Sub can invoke this service
# IAM is configured in pubsub/iam.tf
