variable "project_id" {
  description = "GCP Project ID"
  type        = string
}

variable "region" {
  description = "GCP Region"
  type        = string
  default     = "asia-northeast1"
}

variable "slack_channel" {
  description = "Slack Channel for notifications"
  type        = string
  default     = "#alert"
}

variable "github_repository" {
  description = "GitHub Repository (owner/repo)"
  type        = string
}

variable "api_server_image" {
  description = "Docker image for api-server"
  type        = string
}

variable "log_analysis_server_image" {
  description = "Docker image for log-analysis-server"
  type        = string
}
