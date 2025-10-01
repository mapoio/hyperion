#!/bin/bash
# Script to install Git hooks

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
GIT_DIR="$(git rev-parse --git-dir 2>/dev/null)"

if [ -z "$GIT_DIR" ]; then
    echo "Error: Not in a git repository"
    exit 1
fi

HOOKS_DIR="$GIT_DIR/hooks"

echo -e "${GREEN}Installing Git hooks...${NC}"

# Install pre-commit hook
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    echo -e "${YELLOW}Warning: pre-commit hook already exists, backing up...${NC}"
    mv "$HOOKS_DIR/pre-commit" "$HOOKS_DIR/pre-commit.backup"
fi

ln -sf "$SCRIPT_DIR/pre-commit.sh" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"
echo -e "${GREEN}✓ Installed pre-commit hook${NC}"

# Install commit-msg hook
if [ -f "$HOOKS_DIR/commit-msg" ]; then
    echo -e "${YELLOW}Warning: commit-msg hook already exists, backing up...${NC}"
    mv "$HOOKS_DIR/commit-msg" "$HOOKS_DIR/commit-msg.backup"
fi

ln -sf "$SCRIPT_DIR/commit-msg.sh" "$HOOKS_DIR/commit-msg"
chmod +x "$HOOKS_DIR/commit-msg"
echo -e "${GREEN}✓ Installed commit-msg hook${NC}"

echo ""
echo -e "${GREEN}Git hooks installed successfully!${NC}"
echo ""
echo "Hooks installed:"
echo "  - pre-commit:  Runs linter, formatter, and tests"
echo "  - commit-msg:  Validates commit message format"
echo ""
echo "To bypass hooks (not recommended), use: git commit --no-verify"
