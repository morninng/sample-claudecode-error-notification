.PHONY: help build push deploy test clean

PROJECT_ID ?= $(shell gcloud config get-value project)
REGION ?= asia-northeast1
REGISTRY := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/docker-repo

help:
	@echo "Available commands:"
	@echo "  make setup          - Set up GCP project and enable APIs"
	@echo "  make build          - Build Docker images"
	@echo "  make push           - Push Docker images to Artifact Registry"
	@echo "  make deploy         - Deploy infrastructure with Terraform"
	@echo "  make test           - Test the API server"
	@echo "  make logs-api       - View api-server logs"
	@echo "  make logs-analysis  - View log-analysis-server logs"
	@echo "  make clean          - Destroy infrastructure"

setup:
	@echo "Setting up GCP project..."
	gcloud config set project $(PROJECT_ID)
	gcloud services enable run.googleapis.com
	gcloud services enable logging.googleapis.com
	gcloud services enable pubsub.googleapis.com
	gcloud services enable artifactregistry.googleapis.com
	gcloud artifacts repositories create docker-repo \
		--repository-format=docker \
		--location=$(REGION) \
		--description="Docker repository" || true
	gcloud auth configure-docker $(REGION)-docker.pkg.dev

build:
	@echo "Building Docker images..."
	docker build -t $(REGISTRY)/api-server:latest ./api-server
	docker build -t $(REGISTRY)/log-analysis-server:latest ./log-analysis-server

push: build
	@echo "Pushing Docker images..."
	docker push $(REGISTRY)/api-server:latest
	docker push $(REGISTRY)/log-analysis-server:latest

deploy:
	@echo "Deploying infrastructure..."
	cd terraform && terraform init && terraform apply

test:
	@echo "Testing API server..."
	$(eval API_URL := $(shell cd terraform && terraform output -raw api_server_url))
	@echo "Normal request:"
	curl "$(API_URL)/hello"
	@echo "\n\nError request (this will trigger notification):"
	curl "$(API_URL)/hello?message=error"

logs-api:
	gcloud run services logs read api-server --region=$(REGION) --limit=50

logs-analysis:
	gcloud run services logs read log-analysis-server --region=$(REGION) --limit=50

clean:
	@echo "Destroying infrastructure..."
	cd terraform && terraform destroy
