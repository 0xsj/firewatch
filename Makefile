.PHONY: fmt
fmt: ## Format all Go code
	@echo "Formatting Go code..."
	@go fmt ./...
	@echo "✓ Done"

.PHONY: help
help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: wire
wire: ## Generate Wire code
	@echo "Generating Wire code..."
	@cd cmd/api && wire
	@cd internal/identity && wire
	@echo "✓ Wire generation complete"