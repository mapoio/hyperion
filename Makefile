# Hyperion Framework Makefile

.PHONY: help
help: ## Display this help message
	@echo "Hyperion Framework - Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@chmod +x scripts/*.sh
	@./scripts/install-hooks.sh
	@echo "Installing golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || (echo "Please install golangci-lint from https://golangci-lint.run/usage/install/" && exit 1)
	@echo "✓ Development environment setup complete"

.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting code..."
	@gofmt -w -s .
	@goimports -w -local github.com/mapoio/hyperion .
	@echo "✓ Code formatted"

.PHONY: lint
lint: ## Run linters
	@echo "Running linters..."
	@golangci-lint run --config=.golangci.yml ./...
	@echo "✓ Linting complete"

.PHONY: lint-fix
lint-fix: ## Run linters with auto-fix
	@echo "Running linters with auto-fix..."
	@golangci-lint run --config=.golangci.yml --fix ./...
	@echo "✓ Linting complete"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "✓ Tests complete"

.PHONY: test-short
test-short: ## Run short tests
	@echo "Running short tests..."
	@go test -short -v ./...
	@echo "✓ Short tests complete"

.PHONY: test-coverage
test-coverage: test ## Generate test coverage report
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

.PHONY: build
build: ## Build the application
	@echo "Building..."
	@go build -v ./...
	@echo "✓ Build complete"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache
	@echo "✓ Clean complete"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "✓ Dependencies updated"

.PHONY: verify
verify: fmt lint test ## Verify code (format, lint, test)
	@echo "✓ All verification passed"

.PHONY: check-commit
check-commit: ## Check if code is ready to commit
	@echo "Checking if code is ready to commit..."
	@./scripts/pre-commit.sh
	@echo "✓ Code is ready to commit"

.PHONY: install-hooks
install-hooks: ## Install Git hooks
	@./scripts/install-hooks.sh

.PHONY: uninstall-hooks
uninstall-hooks: ## Uninstall Git hooks
	@echo "Uninstalling Git hooks..."
	@rm -f .git/hooks/pre-commit .git/hooks/commit-msg
	@echo "✓ Git hooks uninstalled"

.PHONY: mod-upgrade
mod-upgrade: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✓ Dependencies upgraded"

.PHONY: mod-vendor
mod-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	@go mod vendor
	@echo "✓ Dependencies vendored"

.PHONY: tools
tools: ## Install development tools
	@echo "Installing development tools..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@echo "✓ Development tools installed"

# Git commit template setup
.PHONY: set-commit-template
set-commit-template: ## Set Git commit message template
	@git config commit.template .commit-msg-template.txt
	@echo "✓ Git commit template set"

.PHONY: ci
ci: deps verify ## Run CI pipeline locally
	@echo "✓ CI pipeline completed successfully"
