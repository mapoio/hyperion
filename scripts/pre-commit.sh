#!/bin/bash
# Git pre-commit hook for running code quality checks

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running pre-commit checks...${NC}"

# Check if golangci-lint is installed
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${YELLOW}Warning: golangci-lint is not installed${NC}"
    echo "Install it with: brew install golangci-lint"
    echo "Or download from: https://golangci-lint.run/usage/install/"
    echo ""
    echo "Skipping linter checks..."
else
    echo -e "${GREEN}Running golangci-lint...${NC}"

    # Get list of staged Go files
    staged_go_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

    if [ -n "$staged_go_files" ]; then
        # Run golangci-lint on staged files
        golangci-lint run --new-from-rev=HEAD --config=.golangci.yml $staged_go_files

        if [ $? -ne 0 ]; then
            echo -e "${RED}ERROR: golangci-lint found issues${NC}"
            echo "Please fix the issues above before committing."
            echo ""
            echo "To skip this check (not recommended), use: git commit --no-verify"
            exit 1
        fi

        echo -e "${GREEN}✓ Linter checks passed${NC}"
    else
        echo "No Go files to lint"
    fi
fi

# Check if gofmt is needed
echo -e "${GREEN}Checking code formatting...${NC}"
staged_go_files=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')

if [ -n "$staged_go_files" ]; then
    unformatted_files=$(gofmt -l $staged_go_files)

    if [ -n "$unformatted_files" ]; then
        echo -e "${RED}ERROR: The following files are not formatted:${NC}"
        echo "$unformatted_files"
        echo ""
        echo "Please run: gofmt -w <file>"
        echo "Or run: make fmt"
        echo ""
        echo "To skip this check (not recommended), use: git commit --no-verify"
        exit 1
    fi

    echo -e "${GREEN}✓ Code formatting check passed${NC}"
fi

# Run tests if any Go files are changed
if [ -n "$staged_go_files" ]; then
    echo -e "${GREEN}Running tests...${NC}"

    # Only run tests if test files exist
    if find . -name '*_test.go' -type f | grep -q .; then
        go test -short ./...

        if [ $? -ne 0 ]; then
            echo -e "${RED}ERROR: Tests failed${NC}"
            echo "Please fix failing tests before committing."
            echo ""
            echo "To skip this check (not recommended), use: git commit --no-verify"
            exit 1
        fi

        echo -e "${GREEN}✓ Tests passed${NC}"
    else
        echo "No tests found"
    fi
fi

echo -e "${GREEN}All pre-commit checks passed!${NC}"
exit 0
