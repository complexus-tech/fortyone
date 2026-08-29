# Go entrypoints load .env with the application config parser. Never parse it as
# Make syntax: passwords may contain Make metacharacters, and SQLC validation
# must receive SQLC_DATABASE_URL explicitly from its calling environment.

# Air v1.64.5 is the latest reviewed release whose module contract remains on
# the API's Go 1.25 toolchain. Later releases require Go 1.26.
AIR_VERSION := v1.64.5
AIR := $(CURDIR)/.tools/bin/air
MIGRATE_VERSION := v4.18.3
MIGRATE := $(CURDIR)/.tools/bin/migrate

dev: air-bootstrap
	@$(AIR)

worker:
	go run ./cmd/worker

develop:
	go run ./cmd/api

tidy:
	go mod tidy

.PHONY: dev worker develop tidy bootstrap-tools tool-versions air-bootstrap migrate-bootstrap generate generated-check check-fast ci test test-integration test-race test-fuzz format-check staticcheck workflow-check security-bootstrap security-check g104-check gitleaks-check quality-check architecture-check architecture-baseline-generate oapi-bootstrap oasdiff-bootstrap openapi-generate openapi-check openapi-breaking-check sdk-generate sdk-check

test:
	@go test -count=1 ./...

test-integration:
	@go test -tags=integration -count=1 ./...

test-race:
	@go test -count=1 -race ./...

# Keep fuzz smoke deterministic in scope and bounded in time. Add a line here
# when a new fuzz target becomes part of the governed pull-request surface.
test-fuzz:
	@go test -run '^$$' -fuzz '^FuzzCursorCodecDecode$$' -fuzztime=5s ./internal/platform/pagination
	@go test -run '^$$' -fuzz '^FuzzParseKeyAndDigest$$' -fuzztime=5s ./internal/platform/idempotency
	@go test -run '^$$' -fuzz '^FuzzHashRequestUsesExactBytes$$' -fuzztime=5s ./internal/platform/idempotency
	@go test -run '^$$' -fuzz '^FuzzInvitationTokenLookupNeverPanics$$' -fuzztime=5s ./internal/modules/invitations/service
	@go test -run '^$$' -fuzz '^FuzzParseURLFormNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzDecodeWithLimitNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzUUIDPathParameterDoesNotExposeInput$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzOptionalQueryParameterNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzOptionalTextQueryParameterNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzOptionalListQueryParameterNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzOptionalCommaSeparatedUUIDQueryParameterNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzOptionalIntegerQueryParameterNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzReadBoundedBodyNeverPanics$$' -fuzztime=5s ./pkg/web
	@go test -run '^$$' -fuzz '^FuzzInt64ToInt32$$' -fuzztime=5s ./internal/platform/safecast

# =============================================================================
# Required API quality and security checks
# =============================================================================

STATICCHECK_VERSION := v0.7.0
GOVULNCHECK_VERSION := v1.7.0
GOSEC_VERSION := v2.29.0
ACTIONLINT_VERSION := v1.7.7
OASDIFF_VERSION := v1.28.0
OASDIFF := $(CURDIR)/.tools/bin/oasdiff
GITLEAKS := $(CURDIR)/.tools/bin/gitleaks

format-check:
	@files="$$(gofmt -l .)"; if [ -n "$$files" ]; then printf 'gofmt required:\n%s\n' "$$files" >&2; exit 1; fi

staticcheck:
	@go run honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION) ./...

workflow-check:
	@go run github.com/rhysd/actionlint/cmd/actionlint@$(ACTIONLINT_VERSION) ../../.github/workflows/*.yml

security-bootstrap:
	@./scripts/install-gitleaks.sh

security-check:
	@go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...
	@$(MAKE) g104-check

# The complete gosec report is classified in one pass: every rule remains
# blocking while the only approved finding is an AST-verified standalone
# pkg/web response write for G104. Scanner stderr is withheld because
# diagnostics can contain source snippets; malformed or incomplete JSON still
# fails the guard. Do not add -quiet: gosec omits its JSON report when no issues
# are found in quiet mode.
g104-check:
	@go run github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION) -exclude-generated -fmt=json ./... 2>/dev/null | go run ./internal/tools/g104guard -root .

gitleaks-check: security-bootstrap
	@GITLEAKS_BIN="$(GITLEAKS)" ./scripts/check-gitleaks.sh

quality-check: format-check
	@go mod verify
	@go mod tidy -diff
	@go vet ./...
	@$(MAKE) staticcheck
	@$(MAKE) workflow-check

architecture-check:
	@go test -count=1 ./internal/bootstrap/architecture

# This command replaces the debt ceiling with the exact current snapshot. Run
# it only after reviewing intentional reductions; never use it to accept growth.
architecture-baseline-generate:
	@FORTYONE_UPDATE_ARCHITECTURE_BASELINE=1 go test -count=1 -run '^TestUpdateArchitectureDebtBaseline$$' ./internal/bootstrap/architecture
	@$(MAKE) architecture-check

bootstrap-tools: air-bootstrap migrate-bootstrap sqlc-bootstrap oapi-bootstrap oasdiff-bootstrap security-bootstrap

air-bootstrap:
	@if [ ! -x "$(AIR)" ]; then \
		mkdir -p "$(dir $(AIR))"; \
		GOTOOLCHAIN=go1.25.0 GOBIN="$(dir $(AIR))" go install github.com/air-verse/air@$(AIR_VERSION); \
	fi

migrate-bootstrap:
	@if [ ! -x "$(MIGRATE)" ]; then \
		actual_version="$$(go list -m -f '{{.Version}}' github.com/golang-migrate/migrate/v4)"; \
		if [ "$$actual_version" != "$(MIGRATE_VERSION)" ]; then \
			printf 'golang-migrate module version is %s, want %s\n' "$$actual_version" "$(MIGRATE_VERSION)" >&2; \
			exit 1; \
		fi; \
		mkdir -p "$(dir $(MIGRATE))"; \
		GOTOOLCHAIN=go1.25.0 go build -tags postgres -ldflags '-X main.Version=$(MIGRATE_VERSION)' -o "$(MIGRATE)" github.com/golang-migrate/migrate/v4/cmd/migrate; \
	fi

tool-versions:
	@go version
	@printf 'air %s\n' '$(AIR_VERSION)'
	@printf 'golang-migrate %s\n' '$(MIGRATE_VERSION)'
	@printf 'sqlc %s\n' 'v1.31.1'
	@printf 'oapi-codegen %s\n' 'v2.8.0'
	@printf 'oapi-codegen runtime %s\n' 'v1.6.0'
	@printf 'openapi-typescript %s\n' 'v7.13.0'
	@printf 'openapi-fetch %s\n' 'v0.17.0'
	@printf 'SDK TypeScript %s\n' 'v5.9.3'
	@printf 'staticcheck %s\n' 'v0.7.0'
	@printf 'govulncheck %s\n' 'v1.7.0'
	@printf 'gosec %s\n' 'v2.29.0'
	@printf 'actionlint %s\n' 'v1.7.7'
	@printf 'oasdiff %s\n' 'v1.28.0'
	@printf 'gitleaks %s\n' 'v8.27.2'

# Intentional writers. The architecture baseline is excluded because accepting
# its diff always requires a separate, explicit review.
generate: sqlc-generate openapi-generate sdk-generate config-generate inventory-generate migration-docs

check-fast: format-check
	@go mod verify
	@go mod tidy -diff
	@go vet ./...
	@go test -count=1 ./...
	@$(MAKE) generated-check
	@$(MAKE) architecture-check

ci: check-fast
	@$(MAKE) staticcheck
	@$(MAKE) workflow-check
	@$(MAKE) security-check
	@$(MAKE) gitleaks-check

# =============================================================================
# Typed SQL generation (sqlc)
# =============================================================================

SQLC := $(CURDIR)/.tools/bin/sqlc

.PHONY: sqlc-bootstrap sqlc-generate sqlc-check sqlc-vet config-generate config-check inventory-generate inventory-check generated-check

sqlc-bootstrap:
	@./scripts/install-sqlc.sh

sqlc-generate:
	@SQLC_BIN="$(SQLC)" ./scripts/generate-sqlc.sh

sqlc-check:
	@SQLC_BIN="$(SQLC)" ./scripts/check-sqlc.sh
	@go test -run '^$$' ./...

sqlc-vet:
	@SQLC_BIN="$(SQLC)" ./scripts/vet-sqlc.sh

config-generate:
	@go run ./internal/tools/configreference -write

config-check:
	@go run ./internal/tools/configreference

inventory-generate:
	@go run ./internal/tools/apiinventory -write

inventory-check:
	@go run ./internal/tools/apiinventory

generated-check: sqlc-check openapi-check sdk-check config-check inventory-check migration-check

# =============================================================================
# Public API contract generation (OpenAPI)
# =============================================================================

oapi-bootstrap:
	@./scripts/install-oapi-codegen.sh

oasdiff-bootstrap:
	@./scripts/install-oasdiff.sh

openapi-generate:
	@OAPI_CODEGEN_BIN="$(CURDIR)/.tools/bin/oapi-codegen" ./scripts/generate-openapi.sh

openapi-check:
	@OAPI_CODEGEN_BIN="$(CURDIR)/.tools/bin/oapi-codegen" ./scripts/check-openapi.sh

# OPENAPI_BASE_REF must identify the reviewed base commit (normally the pull
# request base SHA). The script resolves it to a commit before extracting the
# split spec, so neither a moving branch nor an option-shaped ref reaches the
# comparison tool.
openapi-breaking-check: oasdiff-bootstrap
	@OASDIFF_VERSION="$(OASDIFF_VERSION)" OASDIFF_BIN="$(OASDIFF)" ./scripts/check-openapi-breaking.sh

sdk-generate:
	@OAPI_CODEGEN_BIN="$(CURDIR)/.tools/bin/oapi-codegen" ./scripts/generate-sdk.sh

sdk-check:
	@OAPI_CODEGEN_BIN="$(CURDIR)/.tools/bin/oapi-codegen" ./scripts/check-sdk.sh

# =============================================================================
# Database Migrations (golang-migrate)
# =============================================================================
# Override DB_URL inline or use the APP_DB_* environment contract. migrate.sh
# loads .env without Make parsing and builds the latter through the same
# validated DSN code as the API and worker.

.PHONY: migration-docs migration-check migrate-create migrate-up migrate-down migrate-version migrate-force

migration-docs:
	@go run ./internal/tools/migrationcheck -write

migration-check:
	@go run ./internal/tools/migrationcheck

# Create a new migration: make migrate-create name=create_users_table
migrate-create: migrate-bootstrap
	$(MIGRATE) create -ext sql -dir internal/migrations -seq $(name)

# Apply all pending migrations
migrate-up: migrate-bootstrap
	@MIGRATE_BIN="$(MIGRATE)" ./scripts/migrate.sh up

# Rollback last N migrations (default 1): make migrate-down n=1
migrate-down: migrate-bootstrap
	@MIGRATE_BIN="$(MIGRATE)" ./scripts/migrate.sh down "$(or $(n),1)"

# Show current migration version
migrate-version: migrate-bootstrap
	@MIGRATE_BIN="$(MIGRATE)" ./scripts/migrate.sh version

# Force set migration version: make migrate-force v=2
migrate-force: migrate-bootstrap
	@MIGRATE_BIN="$(MIGRATE)" ./scripts/migrate.sh force "$(v)"

# =============================================================================
# Database Seeding
# =============================================================================

# Run database seeding: make seed name="My Workspace" slug="my-workspace" email="admin@example.com"
seed:
	go run ./cmd/seed --name "$(or $(name),Development)" --slug "$(or $(slug),dev)" --email "$(or $(email),admin@example.com)" --fullname "$(or $(fullname),Admin User)" $(if $(strip $(disable-tls)),--disable-tls=$(disable-tls),)
