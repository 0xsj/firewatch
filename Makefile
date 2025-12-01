# ============================================================================
# Hexagonal Go - Makefile
# ============================================================================
#
# Usage:
#   make run                    # Loads .env, runs API
#   make run ENV=api2           # Loads .env.api2, runs API
#   SERVER_PORT=8081 make run   # Override single var
#
# ============================================================================

# ============================================================================
# Environment Configuration
# ============================================================================

# Load .env file by default, or .env.{ENV} if ENV is specified

ENV_FILE := $(if $(ENV),.env.$(ENV),.env)

ifneq (,$(wildcard $(ENV_FILE)))
  include $(ENV_FILE)
  export $(shell sed 's/=.*//' $(ENV_FILE) | grep -v '^\#')
else
  $(warning Warning: $(ENV_FILE) not found, using defaults)
endif

# ============================================================================
# Defaults (used if not in .env)
# ============================================================================

DATABASE_HOST     ?= localhost
DATABASE_PORT     ?= 5436
DATABASE_USER     ?= hexagonal
DATABASE_PASSWORD ?= hexagonal_dev_pass
DATABASE_NAME     ?= hexagonal_go
DATABASE_SSL_MODE ?= disable

SERVER_PORT  ?= 8080
METRICS_PORT ?= 9090

# Derived
DATABASE_URL := postgresql://$(DATABASE_USER):$(DATABASE_PASSWORD)@$(DATABASE_HOST):$(DATABASE_PORT)/$(DATABASE_NAME)?sslmode=$(DATABASE_SSL_MODE)

# Docker
DOCKER_COMPOSE_FILE ?= deployments/docker/docker-compose.yml

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Show this help message
	@echo ''
	@echo '\033[1mHexagonal Go\033[0m'
	@echo ''
	@echo '\033[36mUsage:\033[0m'
	@echo '  make \033[33m<target>\033[0m [ENV=name] [VAR=value]'
	@echo ''
	@echo '\033[36mExamples:\033[0m'
	@echo '  make run                     # Load .env, run API'
	@echo '  make run ENV=api2            # Load .env.api2, run API'
	@echo '  make dev SERVER_PORT=8081    # Override port'
	@echo ''
	@echo '\033[36mTargets:\033[0m'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[33m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ''
	@echo '\033[36mCurrent Configuration:\033[0m'
	@echo '  ENV_FILE:      $(ENV_FILE)'
	@echo '  SERVER_PORT:   $(SERVER_PORT)'
	@echo '  METRICS_PORT:  $(METRICS_PORT)'
	@echo '  DATABASE_HOST: $(DATABASE_HOST):$(DATABASE_PORT)'
	@echo '  DATABASE_NAME: $(DATABASE_NAME)'
	@echo ''

# ============================================================================
# Development
# ============================================================================

.PHONY: install-tools
install-tools: ## Install development tools (air, wire, migrate, swag)
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/google/wire/cmd/wire@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "✓ Tools installed"

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@echo "Starting dev server ($(ENV_FILE))..."
	@sed -i '' 's/dotenv = .*/dotenv = "$(ENV_FILE)"/' .air.toml 2>/dev/null || true
	@air

.PHONY: run
run: wire ## Run the API server
	@echo "Starting API server ($(ENV_FILE))..."
	@go run ./cmd/api

.PHONY: worker
worker: ## Run the background worker
	@echo "Starting worker ($(ENV_FILE))..."
	@go run ./cmd/worker

# ============================================================================
# Build
# ============================================================================

.PHONY: build
build: wire ## Build the API binary
	@echo "Building..."
	@go build -o bin/api ./cmd/api
	@echo "✓ Built to bin/api"

.PHONY: build-worker
build-worker: ## Build the worker binary
	@echo "Building worker..."
	@go build -o bin/worker ./cmd/worker
	@echo "✓ Built to bin/worker"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ tmp/ coverage.out
	@echo "✓ Clean complete"

# ============================================================================
# Docker
# ============================================================================

.PHONY: docker-up
docker-up: ## Start all Docker services
	@echo "Starting Docker services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) up -d
	@echo "✓ Docker services started"
	@echo ""
	@echo "\033[36mServices:\033[0m"
	@echo "  PostgreSQL:      localhost:5436"
	@echo "  Redis:           localhost:6383"
	@echo "  Mailpit SMTP:    localhost:1025"
	@echo "  Mailpit UI:      http://localhost:8025"
	@echo "  Jaeger UI:       http://localhost:16686"
	@echo "  Prometheus:      http://localhost:9092"
	@echo "  Grafana:         http://localhost:3002 (admin/admin)"
	@echo "  OTEL Collector:  localhost:4317 (gRPC), localhost:4318 (HTTP)"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "Stopping Docker services..."
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down
	@echo "✓ Docker services stopped"

.PHONY: docker-logs
docker-logs: ## Show Docker logs (follow)
	@docker-compose -f $(DOCKER_COMPOSE_FILE) logs -f

.PHONY: docker-clean
docker-clean: ## Remove all Docker volumes (WARNING: deletes data)
	@echo "⚠️  This will delete all data. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@docker-compose -f $(DOCKER_COMPOSE_FILE) down -v
	@echo "✓ Docker volumes removed"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker services

# ============================================================================
# Database
# ============================================================================

.PHONY: db-connect
db-connect: ## Connect to PostgreSQL database
	@docker exec -it hexagonal-go-postgres psql -U $(DATABASE_USER) -d $(DATABASE_NAME)

.PHONY: db-logs
db-logs: ## Show PostgreSQL logs
	@docker logs -f hexagonal-go-postgres

.PHONY: redis-cli
redis-cli: ## Connect to Redis CLI
	@docker exec -it hexagonal-go-redis redis-cli

# ============================================================================
# Migrations
# ============================================================================

.PHONY: migrate-up
migrate-up: ## Run all pending migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✓ Migrations complete"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@migrate -path migrations -database "$(DATABASE_URL)" down 1
	@echo "✓ Rollback complete"

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations (WARNING: deletes all data)
	@echo "⚠️  This will delete all data. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@migrate -path migrations -database "$(DATABASE_URL)" down -all
	@migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✓ Database reset complete"

.PHONY: migrate-status
migrate-status: ## Show current migration version
	@migrate -path migrations -database "$(DATABASE_URL)" version

.PHONY: migrate-create
migrate-create: ## Create new migration (NAME=create_foo_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		echo "Usage: make migrate-create NAME=create_foo_table"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✓ Created migration: $(NAME)"

# ============================================================================
# Code Generation
# ============================================================================

.PHONY: wire
wire: ## Generate Wire dependency injection code
	@echo "Generating wire code..."
	@find . \( -name "wire.go" -o -name "provider.go" \) -not -path "./vendor/*" -not -path "./tmp/*" | while read -r file; do \
		if grep -q "wireinject" "$$file" 2>/dev/null; then \
			dir=$$(dirname "$$file"); \
			echo "  → $$dir"; \
			(cd "$$dir" && wire) || exit 1; \
		fi \
	done
	@echo "✓ Wire generation complete!"

.PHONY: wire-check
wire-check: ## Verify Wire configuration
	@echo "Checking wire configuration..."
	@find . \( -name "wire.go" -o -name "provider.go" \) -not -path "./vendor/*" -not -path "./tmp/*" | while read -r file; do \
		if grep -q "wireinject" "$$file" 2>/dev/null; then \
			dir=$$(dirname "$$file"); \
			echo "  → $$dir"; \
			(cd "$$dir" && wire check) || exit 1; \
		fi \
	done
	@echo "✓ Wire check complete!"

.PHONY: swagger
swagger: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@swag init \
		--generalInfo cmd/api/main.go \
		--output docs/swagger \
		--parseDependency \
		--parseInternal \
		--parseDepth 1
	@echo "✓ Swagger docs generated at docs/swagger/"

.PHONY: swagger-fmt
swagger-fmt: ## Format Swagger comments
	@echo "Formatting Swagger comments..."
	@swag fmt

.PHONY: generate
generate: wire swagger ## Run all code generation (wire + swagger)

# ============================================================================
# Testing & Quality
# ============================================================================

.PHONY: test
test: ## Run all tests
	@go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run tests (short mode, skip slow tests)
	@go test -v -short ./...

.PHONY: test-coverage
test-coverage: test ## Run tests and open coverage report
	@go tool cover -html=coverage.out

.PHONY: lint
lint: ## Run linter
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@go fmt ./...
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@go vet ./...

.PHONY: check
check: fmt vet lint test-short ## Run all checks (fmt, vet, lint, test)

# ============================================================================
# Utilities
# ============================================================================

.PHONY: env-example
env-example: ## Create .env.example from current .env (strips values)
	@if [ -f .env ]; then \
		sed 's/=.*/=/' .env > .env.example; \
		echo "✓ Created .env.example"; \
	else \
		echo "Error: .env not found"; \
		exit 1; \
	fi

.PHONY: env-create
env-create: ## Create new env file (NAME=api2 creates .env.api2)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		echo "Usage: make env-create NAME=api2"; \
		exit 1; \
	fi
	@if [ -f .env.$(NAME) ]; then \
		echo "Error: .env.$(NAME) already exists"; \
		exit 1; \
	fi
	@cp .env .env.$(NAME)
	@echo "✓ Created .env.$(NAME) (edit to customize)"

.DEFAULT_GOAL := help