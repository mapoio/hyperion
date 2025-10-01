#!/bin/bash
# Git commit-msg hook for enforcing AngularJS commit conventions

commit_msg_file=$1
commit_msg=$(cat "$commit_msg_file")

# Color codes
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Commit message pattern (AngularJS convention)
# Format: <type>(<scope>): <subject>
# Example: feat(hyperlog): add structured logging support
pattern='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9-]+\))?!?: .{1,100}$'

# Skip merge commits
if [[ $commit_msg =~ ^Merge ]]; then
    exit 0
fi

# Skip revert commits (they have their own format)
if [[ $commit_msg =~ ^Revert ]]; then
    exit 0
fi

# Get the first line of the commit message
first_line=$(echo "$commit_msg" | head -n1)

# Check if the first line matches the pattern
if ! [[ $first_line =~ $pattern ]]; then
    echo -e "${RED}ERROR: Commit message does not follow the AngularJS convention${NC}"
    echo ""
    echo -e "${YELLOW}Commit message format:${NC}"
    echo "  <type>(<scope>): <subject>"
    echo ""
    echo -e "${YELLOW}Types:${NC}"
    echo "  feat:     A new feature"
    echo "  fix:      A bug fix"
    echo "  docs:     Documentation only changes"
    echo "  style:    Changes that do not affect the meaning of the code"
    echo "  refactor: A code change that neither fixes a bug nor adds a feature"
    echo "  perf:     A code change that improves performance"
    echo "  test:     Adding missing tests or correcting existing tests"
    echo "  build:    Changes that affect the build system or external dependencies"
    echo "  ci:       Changes to CI configuration files and scripts"
    echo "  chore:    Other changes that don't modify src or test files"
    echo "  revert:   Reverts a previous commit"
    echo ""
    echo -e "${YELLOW}Scope:${NC} (optional) A noun describing the section of the codebase"
    echo "  Examples: hyperlog, hyperdb, hyperweb, hypergrpc, etc."
    echo ""
    echo -e "${YELLOW}Subject:${NC}"
    echo "  - Use imperative, present tense: 'add' not 'added' nor 'adds'"
    echo "  - Don't capitalize first letter"
    echo "  - No period (.) at the end"
    echo "  - Maximum 100 characters"
    echo ""
    echo -e "${YELLOW}Examples:${NC}"
    echo "  feat(hyperlog): add structured logging support"
    echo "  fix(hyperdb): correct transaction rollback issue"
    echo "  docs: update installation guide"
    echo "  refactor(hyperweb): simplify middleware chain"
    echo ""
    echo -e "${YELLOW}Your commit message:${NC}"
    echo "  $first_line"
    echo ""
    echo -e "${RED}Please fix your commit message and try again.${NC}"
    exit 1
fi

# Check subject line length
subject=$(echo "$first_line" | sed -E 's/^[a-z]+(\([a-z0-9-]+\))?!?: //')
if [ ${#subject} -gt 100 ]; then
    echo -e "${RED}ERROR: Subject line is too long (${#subject} characters, max 100)${NC}"
    echo ""
    echo -e "${YELLOW}Your subject:${NC}"
    echo "  $subject"
    echo ""
    echo -e "${RED}Please shorten your subject line and try again.${NC}"
    exit 1
fi

exit 0
