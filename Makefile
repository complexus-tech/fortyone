otel:
	docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

stop-otel:
	docker stop jaeger

dev:
	~/go/bin/air    

worker:
	go run cmd/worker/main.go

develop:
	go run cmd/api/main.go 

tidy:
	go mod tidy

# =============================================================================
# Database Migrations (golang-migrate)
# =============================================================================
# Load environment variables from .env if it exists
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Override DB_URL inline or set environment variables
APP_DB_NAME ?= complexus
APP_DB_USER ?= postgres
APP_DB_PASSWORD ?= password
APP_DB_HOST ?= localhost
APP_DB_PORT ?= 5432

DB_URL ?= "postgresql://$(APP_DB_USER):$(APP_DB_PASSWORD)@$(APP_DB_HOST):$(APP_DB_PORT)/$(APP_DB_NAME)?sslmode=disable"
MIGRATE ?= ~/go/bin/migrate

# Create a new migration: make migrate-create name=create_users_table
migrate-create:
	$(MIGRATE) create -ext sql -dir internal/migrations -seq $(name)

# Apply all pending migrations
migrate-up:
	$(MIGRATE) -path internal/migrations -database $(DB_URL) up

# Rollback last N migrations (default 1): make migrate-down n=1
migrate-down:
	$(MIGRATE) -path internal/migrations -database $(DB_URL) down $(or $(n),1)

# Show current migration version
migrate-version:
	$(MIGRATE) -path internal/migrations -database $(DB_URL) version

# Force set migration version: make migrate-force v=2
migrate-force:
	$(MIGRATE) -path internal/migrations -database $(DB_URL) force $(v)

# =============================================================================
# Database Seeding
# =============================================================================

# Run database seeding: make seed name="My Workspace" slug="my-workspace" email="admin@example.com"
seed:
	go run cmd/seed/main.go --name "$(or $(name),Development)" --slug "$(or $(slug),dev)" --email "$(or $(email),admin@example.com)" --fullname "$(or $(fullname),Admin User)" --disable-tls=$(or $(disable-tls),true)
