.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# ============================================================================
# Development
# ============================================================================

.PHONY: install-tools
install-tools: ## Install development tools (air, wire, migrate)
	@echo "Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install github.com/google/wire/cmd/wire@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "✓ Tools installed"

.PHONY: wire
wire: ## Generate Wire dependency injection code
	@echo "Generating Wire code..."
	@cd cmd/api && wire
	@echo "✓ Wire generation complete"

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@air

.PHONY: run
run: wire ## Run the API server
	@go run ./cmd/api

# ============================================================================
# Docker
# ============================================================================

.PHONY: docker-up
docker-up: ## Start all Docker services
	@echo "Starting Docker services..."
	@docker-compose -f deployments/docker/docker-compose.yml up -d
	@echo "✓ Docker services started"
	@echo ""
	@echo "Services:"
	@echo "  PostgreSQL:        localhost:5436"
	@echo "  Redis:             localhost:6383"
	@echo "  Jaeger UI:         http://localhost:16686"
	@echo "  Prometheus:        http://localhost:9092"
	@echo "  Grafana:           http://localhost:3002 (admin/admin)"
	@echo "  OTEL Collector:    localhost:4317 (gRPC), localhost:4318 (HTTP)"

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "Stopping Docker services..."
	@docker-compose -f deployments/docker/docker-compose.yml down
	@echo "✓ Docker services stopped"

.PHONY: docker-logs
docker-logs: ## Show Docker logs
	@docker-compose -f deployments/docker/docker-compose.yml logs -f

.PHONY: docker-clean
docker-clean: ## Remove all Docker volumes (WARNING: deletes data)
	@echo "⚠️  This will delete all data. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@docker-compose -f deployments/docker/docker-compose.yml down -v
	@echo "✓ Docker volumes removed"

.PHONY: docker-restart
docker-restart: docker-down docker-up ## Restart Docker services

# ============================================================================
# Database
# ============================================================================

.PHONY: db-connect
db-connect: ## Connect to PostgreSQL database
	@docker exec -it hexagonal-postgres psql -U hexagonal -d hexagonal_identity

.PHONY: db-logs
db-logs: ## Show PostgreSQL logs
	@docker logs -f hexagonal-postgres

.PHONY: redis-cli
redis-cli: ## Connect to Redis CLI
	@docker exec -it hexagonal-redis redis-cli

# ============================================================================
# Build
# ============================================================================

.PHONY: build
build: wire ## Build the API binary
	@echo "Building..."
	@go build -o bin/api ./cmd/api
	@echo "✓ Built to bin/api"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ tmp/
	@echo "✓ Clean complete"

# ============================================================================
# Testing
# ============================================================================

.PHONY: test
test: ## Run all tests
	@go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out

# ============================================================================
# Code Quality
# ============================================================================

.PHONY: lint
lint: ## Run linter
	@golangci-lint run ./...

.PHONY: fmt
fmt: ## Format code
	@go fmt ./...
	@goimports -w .

# ============================================================================
# Database Migrations
# ============================================================================

.PHONY: migrate-up
migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path migrations -database "postgresql://hexagonal:hexagonal_dev_pass@localhost:5436/hexagonal_identity?sslmode=disable" up
	@echo "✓ Migrations complete"

.PHONY: migrate-down
migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@migrate -path migrations -database "postgresql://hexagonal:hexagonal_dev_pass@localhost:5436/hexagonal_identity?sslmode=disable" down 1
	@echo "✓ Rollback complete"

.PHONY: migrate-reset
migrate-reset: ## Reset all migrations (WARNING: deletes all data)
	@echo "⚠️  This will delete all data. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@migrate -path migrations -database "postgresql://hexagonal:hexagonal_dev_pass@localhost:5435/hexagonal_identity?sslmode=disable" down -all
	@migrate -path migrations -database "postgresql://hexagonal:hexagonal_dev_pass@localhost:5435/hexagonal_identity?sslmode=disable" up
	@echo "✓ Database reset complete"

.PHONY: migrate-status
migrate-status: ## Show migration status
	@migrate -path migrations -database "postgresql://hexagonal:hexagonal_dev_pass@localhost:5436/hexagonal_identity?sslmode=disable" version