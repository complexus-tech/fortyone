# Default tag if not provided
TAG ?= latest

docker-build-web:
	docker build -f deployments/docker/dockerfile.web -t fortyoneapp/web:$(TAG) .

docker-build-server:
	docker build -f deployments/docker/dockerfile.server -t fortyoneapp/server:$(TAG) .

docker-build-worker:
	docker build -f deployments/docker/dockerfile.worker -t fortyoneapp/worker:$(TAG) .

docker-build-all: docker-build-web docker-build-server docker-build-worker
	@echo "All images built with tag: $(TAG)"

docker-up:
	docker compose up -d

docker-down:
	docker compose down

.PHONY: docker-build-web docker-build-server docker-build-worker docker-build-all docker-up docker-down