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
# Override DB_URL inline or set environment variables
DB_URL ?= "postgresql://$(APP_DB_USER):$(APP_DB_PASSWORD)@$(APP_DB_HOST):$(APP_DB_PORT)/$(APP_DB_NAME)?sslmode=disable"
MIGRATE ?= ~/go/bin/migrate

# Create a new migration: make migrate-create name=create_users_table
migrate-create:
	$(MIGRATE) create -ext sql -dir cmd/migrations -seq $(name)

# Apply all pending migrations
migrate-up:
	$(MIGRATE) -path cmd/migrations -database $(DB_URL) up

# Rollback last N migrations (default 1): make migrate-down n=1
migrate-down:
	$(MIGRATE) -path cmd/migrations -database $(DB_URL) down $(or $(n),1)

# Show current migration version
migrate-version:
	$(MIGRATE) -path cmd/migrations -database $(DB_URL) version

# Force set migration version: make migrate-force v=2
migrate-force:
	$(MIGRATE) -path cmd/migrations -database $(DB_URL) force $(v)
