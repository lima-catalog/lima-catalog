.PHONY: test test-go test-js test-all build clean help

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Lima Template Catalog - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: test-go test-js ## Run all tests (Go + JavaScript)

test-go: ## Run Go unit tests
	@echo "🧪 Running Go tests..."
	@go test ./... -v

test-js: ## Run JavaScript tests (requires Node.js)
	@echo "🧪 Running JavaScript tests..."
	@if ! command -v node >/dev/null 2>&1; then \
		echo "❌ Error: Node.js is not installed. Please install Node.js to run JavaScript tests."; \
		exit 1; \
	fi
	@node test.js

test-all: test ## Run all tests (alias for 'test')

test-integration: ## Run integration test (requires GITHUB_TOKEN)
	@echo "🧪 Running integration tests..."
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo ""; \
		echo "⚠️  WARNING: GITHUB_TOKEN environment variable is not set"; \
		echo ""; \
		echo "Integration tests require a GitHub Personal Access Token to run."; \
		echo ""; \
		echo "To set the token:"; \
		echo "  export GITHUB_TOKEN=your_token_here"; \
		echo ""; \
		echo "To create a token:"; \
		echo "  https://github.com/settings/tokens"; \
		echo ""; \
		exit 1; \
	fi
	@./scripts/test-integration.sh

build: ## Build the lima-catalog CLI tool
	@echo "🔨 Building lima-catalog..."
	@go build -o lima-catalog ./cmd/lima-catalog
	@echo "✅ Build complete: ./lima-catalog"

clean: ## Remove build artifacts and test data
	@echo "🧹 Cleaning up..."
	@rm -f lima-catalog
	@rm -rf test_data test_incremental
	@echo "✅ Clean complete"

.PHONY: check-token
check-token: ## Check if GITHUB_TOKEN is set
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "❌ GITHUB_TOKEN is not set"; \
		echo ""; \
		echo "Set it with:"; \
		echo "  export GITHUB_TOKEN=your_token_here"; \
		echo ""; \
		exit 1; \
	else \
		echo "✅ GITHUB_TOKEN is set"; \
	fi
