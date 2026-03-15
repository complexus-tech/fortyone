# Build all images
build:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build tag=v1.2.3"; exit 1; fi
	@$(MAKE) build-web tag=$(tag)
	@$(MAKE) build-server tag=$(tag)
	@$(MAKE) build-worker tag=$(tag)

# Build individual images
build-web:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-web tag=v1.2.3"; exit 1; fi
	docker build --platform linux/amd64,linux/arm64 -f deployments/docker/dockerfile.web -t fortyoneapp/web:$(tag) -t fortyoneapp/web:latest .

build-server:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-server tag=v1.2.3"; exit 1; fi
	docker build --platform linux/amd64,linux/arm64 -f deployments/docker/dockerfile.server -t fortyoneapp/server:$(tag) -t fortyoneapp/server:latest .

build-worker:
	@if [ -z "$(tag)" ]; then echo "ERROR: tag required. Usage: make build-worker tag=v1.2.3"; exit 1; fi
	docker build --platform linux/amd64,linux/arm64 -f deployments/docker/dockerfile.worker -t fortyoneapp/worker:$(tag) -t fortyoneapp/worker:latest .

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

dev:
	@set -e; \
	pnpm dev & \
	pnpm_pid=$$!; \
	$(MAKE) -C apps/server dev & \
	server_pid=$$!; \
	trap 'kill $$pnpm_pid $$server_pid 2>/dev/null || true' INT TERM EXIT; \
	while kill -0 $$pnpm_pid 2>/dev/null && kill -0 $$server_pid 2>/dev/null; do \
		sleep 1; \
	done; \
	kill $$pnpm_pid $$server_pid 2>/dev/null || true; \
	wait $$pnpm_pid 2>/dev/null || pnpm_status=$$?; \
	wait $$server_pid 2>/dev/null || server_status=$$?; \
	if [ "$${pnpm_status:-0}" -ne 0 ] && [ "$${pnpm_status:-0}" -ne 130 ] && [ "$${pnpm_status:-0}" -ne 143 ]; then \
		exit $$pnpm_status; \
	fi; \
	if [ "$${server_status:-0}" -ne 0 ] && [ "$${server_status:-0}" -ne 130 ] && [ "$${server_status:-0}" -ne 143 ]; then \
		exit $$server_status; \
	fi

.PHONY: build build-web build-server build-worker push push-web push-server push-worker up down dev
