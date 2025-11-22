.PHONY: test test-go test-js test-all build clean help lint vet pr

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Lima Template Catalog - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

test: vet test-go test-js ## Run linting and all tests (Go + JavaScript)

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

vet: ## Run go vet to check for common errors
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ go vet passed"

lint: ## Run golangci-lint (requires golangci-lint to be installed)
	@echo "🔍 Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "❌ Error: golangci-lint is not installed."; \
		echo ""; \
		echo "Install it with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		echo ""; \
		exit 1; \
	fi
	@golangci-lint run --timeout=5m
	@echo "✅ golangci-lint passed"

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

pr: ## Prepare for PR: rebase on main, run smart tests
	@echo "🚀 Preparing for PR..."
	@echo ""
	@# Check for uncommitted changes
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "❌ Error: You have uncommitted changes"; \
		echo ""; \
		echo "Please commit or stash your changes first:"; \
		echo "  git status"; \
		echo ""; \
		exit 1; \
	fi
	@# Fetch main
	@echo "📡 Fetching origin/main..."
	@git fetch origin main --quiet
	@# Detect changed files vs origin/main
	@CHANGED_FILES=$$(git diff origin/main...HEAD --name-only); \
	HAS_GO_FILES=$$(echo "$$CHANGED_FILES" | grep -E '\.(go|mod|sum)$$' || true); \
	HAS_JS_FILES=$$(echo "$$CHANGED_FILES" | grep -E '\.(js|html|css)$$' || true); \
	echo ""; \
	echo "📝 Changed files:"; \
	echo "$$CHANGED_FILES" | sed 's/^/  /'; \
	echo ""; \
	if [ -n "$$HAS_GO_FILES" ] && [ -n "$$HAS_JS_FILES" ]; then \
		echo "🧪 Running all tests (Go + JS files changed)..."; \
		$(MAKE) test; \
		if command -v golangci-lint >/dev/null 2>&1; then \
			$(MAKE) lint; \
		fi; \
	elif [ -n "$$HAS_GO_FILES" ]; then \
		echo "🧪 Running Go tests only..."; \
		$(MAKE) test-go; \
		if command -v golangci-lint >/dev/null 2>&1; then \
			$(MAKE) lint; \
		fi; \
	elif [ -n "$$HAS_JS_FILES" ]; then \
		echo "🧪 Running JS tests only..."; \
		$(MAKE) test-js; \
	else \
		echo "📄 No code files changed, running all tests to be safe..."; \
		$(MAKE) test; \
	fi
	@# Attempt rebase
	@echo ""
	@echo "🔄 Rebasing on origin/main..."
	@if git rebase origin/main; then \
		echo "✅ Rebase successful!"; \
		echo ""; \
		echo "Next steps:"; \
		echo "  1. Push your branch: git push -f"; \
		echo "  2. Create PR: gh pr create"; \
	else \
		echo ""; \
		echo "❌ Rebase failed - conflicts detected!"; \
		echo ""; \
		echo "To resolve conflicts:"; \
		echo "  1. Fix conflicts in the listed files"; \
		echo "  2. git add <resolved-files>"; \
		echo "  3. git rebase --continue"; \
		echo ""; \
		echo "Or abort the rebase:"; \
		echo "  git rebase --abort"; \
		echo ""; \
		exit 1; \
	fi
