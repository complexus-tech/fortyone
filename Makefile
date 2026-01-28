# Default tag if not provided
TAG ?= latest

build-web:
	docker build -f deployments/docker/dockerfile.web -t fortyoneapp/web:$(TAG) .

build-server:
	docker build -f deployments/docker/dockerfile.server -t fortyoneapp/server:$(TAG) .

build-worker:
	docker build -f deployments/docker/dockerfile.worker -t fortyoneapp/worker:$(TAG) .

build-all: build-web build-server build-worker
	@echo "All images built with tag: $(TAG)"

push-web:
	docker push fortyoneapp/web:$(TAG)

push-server:
	docker push fortyoneapp/server:$(TAG)

push-worker:
	docker push fortyoneapp/worker:$(TAG)

push-all: push-web push-server push-worker
	@echo "All images pushed with tag: $(TAG)"

up:
	docker compose up -d

down:
	docker compose down

.PHONY: build-web build-server build-worker build-all push-web push-server push-worker push-all up down