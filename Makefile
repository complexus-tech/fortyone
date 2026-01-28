# Build all images
build:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build tag=v1.2.3"; exit 1; fi
	@$(MAKE) build-web tag=$(tag)
	@$(MAKE) build-server tag=$(tag)
	@$(MAKE) build-worker tag=$(tag)

# Build individual images
build-web:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-web tag=v1.2.3"; exit 1; fi
	docker build -f deployments/docker/dockerfile.web -t fortyoneapp/web:$(tag) -t fortyoneapp/web:latest .

build-server:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-server tag=v1.2.3"; exit 1; fi
	docker build -f deployments/docker/dockerfile.server -t fortyoneapp/server:$(tag) -t fortyoneapp/server:latest .

build-worker:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-worker tag=v1.2.3"; exit 1; fi
	docker build -f deployments/docker/dockerfile.worker -t fortyoneapp/worker:$(tag) -t fortyoneapp/worker:latest .

# Push all images
push:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make push tag=v1.2.3"; exit 1; fi
	@$(MAKE) push-web tag=$(tag)
	@$(MAKE) push-server tag=$(tag)
	@$(MAKE) push-worker tag=$(tag)

# Push individual images
push-web:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make push-web tag=v1.2.3"; exit 1; fi
	docker push fortyoneapp/web:$(tag)
	docker push fortyoneapp/web:latest

push-server:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make push-server tag=v1.2.3"; exit 1; fi
	docker push fortyoneapp/server:$(tag)
	docker push fortyoneapp/server:latest

push-worker:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make push-worker tag=v1.2.3"; exit 1; fi
	docker push fortyoneapp/worker:$(tag)
	docker push fortyoneapp/worker:latest

# Docker Compose
up:
	docker compose up -d

down:
	docker compose down

.PHONY: build build-web build-server build-worker push push-web push-server push-worker up down