# Hyperion Monorepo Makefile
# This Makefile runs targets across all workspace modules

# All workspace modules (update when adding new modules)
MODULES := hyperion adapter/viper adapter/zap adapter/gorm

.PHONY: help
help: ## Display this help message
	@echo "Hyperion Monorepo - Available targets:"
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
fmt: ## Format Go code across all modules
	@echo "Formatting code across all modules..."
	@for module in $(MODULES); do \
		echo "Formatting $$module..."; \
		(cd $$module && gofmt -w -s . && goimports -w -local github.com/mapoio/hyperion .); \
	done
	@echo "✓ All code formatted"

.PHONY: lint
lint: ## Run linters across all modules
	@echo "Running linters across all modules..."
	@ROOT_DIR=$$(pwd); \
	for module in $(MODULES); do \
		echo "Linting $$module..."; \
		(cd $$module && golangci-lint run --config=$$ROOT_DIR/.golangci.yml ./...) || exit 1; \
	done
	@echo "✓ All linting complete"

.PHONY: lint-fix
lint-fix: ## Run linters with auto-fix across all modules
	@echo "Running linters with auto-fix across all modules..."
	@ROOT_DIR=$$(pwd); \
	for module in $(MODULES); do \
		echo "Linting $$module..."; \
		(cd $$module && golangci-lint run --config=$$ROOT_DIR/.golangci.yml --fix ./...); \
	done
	@echo "✓ All linting complete"

.PHONY: test
test: ## Run tests with coverage across all modules (matches CI)
	@echo "Running tests across all modules..."
	@for module in $(MODULES); do \
		echo "Testing $$module..."; \
		(cd $$module && go test -v -race -coverprofile=coverage.out -covermode=atomic ./...) || exit 1; \
	done
	@echo "✓ All tests complete"

.PHONY: test-short
test-short: ## Run short tests across all modules
	@echo "Running short tests across all modules..."
	@for module in $(MODULES); do \
		echo "Short testing $$module..."; \
		(cd $$module && go test -short -v ./...) || exit 1; \
	done
	@echo "✓ All short tests complete"

.PHONY: test-coverage
test-coverage: test ## Generate test coverage reports
	@echo "Generating coverage reports..."
	@for module in $(MODULES); do \
		if [ -f $$module/coverage.out ]; then \
			echo "Generating coverage for $$module..."; \
			(cd $$module && go tool cover -html=coverage.out -o coverage.html); \
		fi \
	done
	@echo "✓ Coverage reports generated (coverage.html in each module)"

.PHONY: build
build: ## Build all modules
	@echo "Building all modules..."
	@for module in $(MODULES); do \
		echo "Building $$module..."; \
		(cd $$module && go build -v ./...) || exit 1; \
	done
	@echo "✓ All builds complete"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning all modules..."
	@for module in $(MODULES); do \
		echo "Cleaning $$module..."; \
		(cd $$module && rm -f coverage.out coverage.html && go clean -cache -testcache); \
	done
	@rm -f go.work.sum
	@echo "✓ All clean complete"

.PHONY: deps
deps: ## Download dependencies for all modules
	@echo "Downloading dependencies for all modules..."
	@for module in $(MODULES); do \
		echo "Updating dependencies for $$module..."; \
		(cd $$module && go mod download && go mod tidy) || exit 1; \
	done
	@echo "✓ All dependencies updated"

.PHONY: verify
verify: check-format lint test ## Verify code (format check, lint, test)
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
	@echo "Upgrading dependencies for all modules..."
	@for module in $(MODULES); do \
		echo "Upgrading dependencies for $$module..."; \
		(cd $$module && go get -u ./... && go mod tidy) || exit 1; \
	done
	@echo "✓ All dependencies upgraded"

.PHONY: mod-vendor
mod-vendor: ## Vendor dependencies for all modules
	@echo "Vendoring dependencies for all modules..."
	@for module in $(MODULES); do \
		echo "Vendoring dependencies for $$module..."; \
		(cd $$module && go mod vendor) || exit 1; \
	done
	@echo "✓ All dependencies vendored"

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

.PHONY: check-format
check-format: ## Check code formatting (matches CI)
	@echo "Checking code formatting..."
	@for module in $(MODULES); do \
		echo "Checking format for $$module..."; \
		unformatted=$$(cd $$module && gofmt -l .); \
		if [ -n "$$unformatted" ]; then \
			echo "❌ The following files are not formatted:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi \
	done
	@echo "✓ All code is properly formatted"

.PHONY: check-coverage
check-coverage: ## Check coverage threshold (80%, matches CI)
	@echo "Checking coverage threshold (80%)..."
	@for module in $(MODULES); do \
		if [ -f $$module/coverage.out ]; then \
			coverage=$$(cd $$module && go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
			echo "$$module coverage: $$coverage%"; \
			if [ $$(echo "$$coverage < 80" | bc -l) -eq 1 ]; then \
				echo "❌ $$module coverage $$coverage% is below threshold 80%"; \
				exit 1; \
			fi \
		fi \
	done
	@echo "✓ All modules meet coverage threshold"

.PHONY: security
security: ## Run security scan across all modules (matches CI)
	@echo "Running security scan..."
	@command -v gosec >/dev/null 2>&1 || (echo "Installing gosec..." && go install github.com/securego/gosec/v2/cmd/gosec@latest)
	@for module in $(MODULES); do \
		echo "Security scanning $$module..."; \
		(cd $$module && gosec -no-fail ./...) || exit 1; \
	done
	@echo "✓ Security scan complete"

.PHONY: vuln-check
vuln-check: ## Check for vulnerable dependencies (matches CI)
	@echo "Checking for vulnerable dependencies..."
	@command -v govulncheck >/dev/null 2>&1 || (echo "Installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	@for module in $(MODULES); do \
		echo "Checking vulnerabilities in $$module..."; \
		(cd $$module && govulncheck ./...); \
	done
	@echo "✓ Vulnerability check complete"

.PHONY: quality-tools
quality-tools: ## Install code quality tools
	@echo "Installing code quality tools..."
	@command -v gocyclo >/dev/null 2>&1 || (echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest)
	@command -v gocognit >/dev/null 2>&1 || (echo "Installing gocognit..." && go install github.com/uudashr/gocognit/cmd/gocognit@latest)
	@command -v dupl >/dev/null 2>&1 || (echo "Installing dupl..." && go install github.com/mibk/dupl@latest)
	@echo "✓ Code quality tools installed"

.PHONY: check-cyclo
check-cyclo: ## Check cyclomatic complexity (threshold: 15, matches CI)
	@echo "Checking cyclomatic complexity..."
	@command -v gocyclo >/dev/null 2>&1 || (echo "Installing gocyclo..." && go install github.com/fzipp/gocyclo/cmd/gocyclo@latest)
	@total_funcs=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 0 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	high_complexity=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 15 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	echo "Total functions: $$total_funcs"; \
	echo "High complexity (>15): $$high_complexity"; \
	if [ $$high_complexity -gt 0 ]; then \
		echo "❌ Found $$high_complexity functions with cyclomatic complexity > 15:"; \
		find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 15 {} \; 2>/dev/null; \
		exit 1; \
	else \
		echo "✓ All functions meet cyclomatic complexity threshold"; \
	fi

.PHONY: check-cognit
check-cognit: ## Check cognitive complexity (threshold: 20, matches CI)
	@echo "Checking cognitive complexity..."
	@command -v gocognit >/dev/null 2>&1 || (echo "Installing gocognit..." && go install github.com/uudashr/gocognit/cmd/gocognit@latest)
	@total_funcs=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 0 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	high_cognit=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 20 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	echo "Total functions: $$total_funcs"; \
	echo "High cognitive complexity (>20): $$high_cognit"; \
	if [ $$high_cognit -gt 0 ]; then \
		echo "❌ Found $$high_cognit functions with cognitive complexity > 20:"; \
		find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 20 {} \; 2>/dev/null; \
		exit 1; \
	else \
		echo "✓ All functions meet cognitive complexity threshold"; \
	fi

.PHONY: check-dupl
check-dupl: ## Check code duplication (threshold: 50 tokens, matches CI)
	@echo "Checking code duplication..."
	@command -v dupl >/dev/null 2>&1 || (echo "Installing dupl..." && go install github.com/mibk/dupl@latest)
	@dupl_output=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" | xargs dupl -threshold 50 2>/dev/null || echo ""); \
	if echo "$$dupl_output" | grep -q "Found total 0 clone groups"; then \
		echo "✓ No significant code duplication detected (threshold: 50 tokens)"; \
	elif [ -n "$$dupl_output" ] && ! echo "$$dupl_output" | grep -q "Found total 0 clone groups"; then \
		dupl_count=$$(echo "$$dupl_output" | grep -oE 'Found total [0-9]+' | grep -oE '[0-9]+' || echo 0); \
		echo "❌ Found $$dupl_count clone groups of duplicated code:"; \
		echo "$$dupl_output" | head -50; \
		exit 1; \
	else \
		echo "✓ No significant code duplication detected (threshold: 50 tokens)"; \
	fi

.PHONY: quality
quality: check-cyclo check-cognit check-dupl ## Run all code quality checks (matches CI)
	@echo "✓ All code quality checks passed"

.PHONY: quality-report
quality-report: ## Generate detailed code quality report (matches CI)
	@echo "━━━━━━━━━━━━━━━━━━━━━━ Code Quality Report ━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📊 Cyclomatic Complexity:"
	@command -v gocyclo >/dev/null 2>&1 || go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	@total_funcs=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 0 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	high_complexity=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 15 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	echo "  Total functions: $$total_funcs"; \
	echo "  High complexity (>15): $$high_complexity"; \
	echo "  Top 10 most complex:"; \
	find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocyclo -over 0 {} \; 2>/dev/null | sort -rn | head -10 | sed 's/^/    /' || echo "    No functions found"
	@echo ""
	@echo "🧠 Cognitive Complexity:"
	@command -v gocognit >/dev/null 2>&1 || go install github.com/uudashr/gocognit/cmd/gocognit@latest
	@total_funcs=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 0 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	high_cognit=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 20 {} \; 2>/dev/null | wc -l | tr -d ' '); \
	echo "  Total functions: $$total_funcs"; \
	echo "  High cognitive complexity (>20): $$high_cognit"; \
	echo "  Top 10 most complex:"; \
	find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" -exec gocognit -over 0 {} \; 2>/dev/null | sort -rn | head -10 | sed 's/^/    /' || echo "    No functions found"
	@echo ""
	@echo "🔄 Code Duplication:"
	@command -v dupl >/dev/null 2>&1 || go install github.com/mibk/dupl@latest
	@dupl_output=$$(find ./hyperion ./adapter -name "*.go" -not -name "*_test.go" | xargs dupl -threshold 50 2>/dev/null || echo ""); \
	if echo "$$dupl_output" | grep -q "Found total 0 clone groups"; then \
		echo "  ✓ No significant code duplication detected (threshold: 50 tokens)"; \
	elif [ -n "$$dupl_output" ] && ! echo "$$dupl_output" | grep -q "Found total 0 clone groups"; then \
		dupl_count=$$(echo "$$dupl_output" | grep -oE 'Found total [0-9]+' | grep -oE '[0-9]+' || echo 0); \
		echo "  ⚠️  Found $$dupl_count clone groups"; \
		echo "$$dupl_output" | head -30 | sed 's/^/    /'; \
	else \
		echo "  ✓ No significant code duplication detected (threshold: 50 tokens)"; \
	fi
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# GitHub-Agnostic CI Checks (used by both local dev and CI/CD)
# These targets can be used with any CI provider (GitHub, GitLab, etc.)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: check-workspace
check-workspace: ## Verify Go workspace configuration
	@echo "Verifying Go workspace configuration..."
	@go work sync
	@echo "✓ Workspace verified"

.PHONY: mod-verify
mod-verify: ## Verify module dependencies
	@echo "Verifying dependencies for all modules..."
	@for module in $(MODULES); do \
		echo "Verifying dependencies for $$module..."; \
		(cd $$module && go mod download && go mod verify) || exit 1; \
	done
	@echo "✓ All dependencies verified"

.PHONY: check-large-files
check-large-files: ## Check for large files (>1MB)
	@echo "Checking for large files (>1MB)..."
	@large_files=$$(find . -type f -size +1M -not -path "./.git/*" -not -path "./vendor/*" -not -path "./.idea/*" -not -path "./.vscode/*" 2>/dev/null || true); \
	if [ -n "$$large_files" ]; then \
		echo "❌ Large files detected (>1MB):"; \
		echo "$$large_files"; \
		exit 1; \
	fi
	@echo "✓ No large files detected"

.PHONY: check-conflicts
check-conflicts: ## Check for merge conflict markers
	@echo "Checking for merge conflict markers..."
	@conflicts=$$(grep -rE '(<{7} HEAD|>{7})' . --exclude-dir=.git --exclude-dir=vendor --exclude-dir=.idea --exclude-dir=.vscode 2>/dev/null || true); \
	if [ -n "$$conflicts" ]; then \
		echo "❌ Merge conflict markers detected:"; \
		echo "$$conflicts"; \
		exit 1; \
	fi
	@echo "✓ No merge conflicts detected"

.PHONY: lint-commits
lint-commits: ## Validate commit messages (conventional commits)
	@echo "Validating commit messages..."
	@if git rev-parse --verify origin/main >/dev/null 2>&1; then \
		commits=$$(git log --format=%s --no-merges origin/main..HEAD 2>/dev/null || git log --format=%s --no-merges -10); \
		echo "$$commits" | while IFS= read -r commit; do \
			[ -z "$$commit" ] && continue; \
			if ! echo "$$commit" | grep -qE '^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?!?: .+'; then \
				echo "❌ Invalid commit message: $$commit"; \
				echo "Expected format: type(scope): description or type(scope)!: description"; \
				exit 1; \
			fi; \
		done; \
	else \
		echo "⚠️  No origin/main branch found, skipping commit message validation"; \
	fi
	@echo "✓ All commit messages are valid"

.PHONY: coverage-report
coverage-report: test ## Generate detailed coverage report with statistics
	@echo "━━━━━━━━━━━━━━━━━━━━━━ Coverage Report ━━━━━━━━━━━━━━━━━━━━━━"
	@total_coverage=0; \
	module_count=0; \
	for module in $(MODULES); do \
		if [ -f $$module/coverage.out ]; then \
			coverage=$$(cd $$module && go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
			echo "📊 $$module: $$coverage%"; \
			total_coverage=$$(echo "$$total_coverage + $$coverage" | bc); \
			module_count=$$((module_count + 1)); \
		fi \
	done; \
	if [ $$module_count -gt 0 ]; then \
		avg_coverage=$$(echo "scale=2; $$total_coverage / $$module_count" | bc); \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
		echo "📈 Average Coverage: $$avg_coverage%"; \
		echo "🎯 Threshold: 80%"; \
		if [ $$(echo "$$avg_coverage >= 80" | bc -l) -eq 1 ]; then \
			echo "✅ PASS: Coverage meets threshold"; \
		else \
			echo "❌ FAIL: Coverage below threshold"; \
		fi; \
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"; \
	fi

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# CI Pipeline Targets (Platform Agnostic)
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

.PHONY: ci-pre
ci-pre: check-workspace mod-verify ## Run pre-CI checks (workspace and dependencies)
	@echo "✓ Pre-CI checks complete"

.PHONY: ci-test
ci-test: test check-coverage ## Run CI test suite with coverage validation
	@echo "✓ CI tests complete"

.PHONY: ci-lint
ci-lint: check-format lint ## Run CI linting checks
	@echo "✓ CI linting complete"

.PHONY: ci-security
ci-security: security vuln-check ## Run CI security scans
	@echo "✓ CI security checks complete"

.PHONY: ci-quality
ci-quality: quality ## Run CI code quality checks
	@echo "✓ CI quality checks complete"

.PHONY: ci-pr
ci-pr: check-large-files check-conflicts lint-commits ## Run PR-specific checks
	@echo "✓ PR checks complete"

.PHONY: ci
ci: ci-pre ci-lint ci-test ci-security ci-quality build ## Run complete CI pipeline (matches GitHub Actions)
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Complete CI pipeline passed successfully"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

.PHONY: ci-full
ci-full: ci ci-pr coverage-report quality-report ## Run full CI pipeline with detailed reports
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Full CI pipeline with reports completed"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
