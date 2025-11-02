# Cloud Run service for API server
resource "google_cloud_run_service" "api_server" {
  name     = "api-server"
  location = var.region

  template {
    spec {
      service_account_name = google_service_account.api_server.email

      containers {
        image = var.api_server_image

        ports {
          container_port = 8080
        }

        resources {
          limits = {
            cpu    = "1000m"
            memory = "512Mi"
          }
        }
      }
    }

    metadata {
      annotations = {
        "autoscaling.knative.dev/maxScale" = "10"
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
    google_service_account.api_server
  ]
}

# Allow unauthenticated access to API server (public)
resource "google_cloud_run_service_iam_member" "api_server_public" {
  service  = google_cloud_run_service.api_server.name
  location = var.region
  role     = "roles/run.invoker"
  member   = "allUsers"
}
