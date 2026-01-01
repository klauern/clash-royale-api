#!/bin/bash
# setup-githooks.sh - Install git hooks for the clash-royale-api project
# This script configures git to automatically format code before pushing

set -e

# Get the repository root directory
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKS_DIR="$REPO_ROOT/.git/hooks"

echo "Installing git hooks..."

# Create the pre-push hook
cat > "$HOOKS_DIR/pre-push" << 'EOF'
#!/bin/bash
# pre-push hook - Auto-format code before pushing to prevent CI failures

set -e

# Check if there are any Go files that need formatting
echo "Running gofumpt to check formatting..."

# Temporarily stash any uncommitted changes
STASH_RESULT=$(git stash push --keep-index --include-untracked --quiet 2>&1 || true)

# Run gofumpt to format all Go files
if command -v gofumpt &> /dev/null; then
    gofumpt -w . 2>/dev/null || true
else
    echo "gofumpt not found, skipping auto-format"
fi

# Check if any files were changed by formatting
CHANGED=$(git diff --name-only | grep -E '\.go$' || true)

if [ -n "$CHANGED" ]; then
    echo "The following files were auto-formatted:"
    echo "$CHANGED"
    echo ""

    # Add the formatted files
    git add $CHANGED

    # Check if we should commit automatically
    # Only commit if we're on a non-main branch and the user hasn't disabled it
    CURRENT_BRANCH=$(git branch --show-current)
    if [ "$CURRENT_BRANCH" != "main" ] && [ "$CURRENT_BRANCH" != "master" ]; then
        if [ -z "$DISABLE_AUTO_COMMIT" ]; then
            echo "Auto-committing formatting changes..."
            git commit -m "style: auto-format with gofumpt [skip ci]" --no-verify 2>/dev/null || true
            echo "Formatting changes committed. Please push again."
            exit 1
        else
            echo "DISABLE_AUTO_COMMIT is set - files staged but not committed"
            exit 1
        fi
    else
        echo "On main/master branch - files staged but not committed"
        echo "Please review the changes and commit manually, or set DISABLE_AUTO_COMMIT=1"
        exit 1
    fi
fi

# Restore stashed changes if any
if [ -n "$STASH_RESULT" ]; then
    git stash pop --quiet 2>/dev/null || true
fi

echo "No formatting changes needed"
EOF

chmod +x "$HOOKS_DIR/pre-push"

echo "Git hooks installed successfully!"
echo ""
echo "The pre-push hook will:"
echo "  1. Run gofumpt on all Go files before pushing"
echo "  2. Automatically commit formatting changes on feature branches"
echo "  3. Prevent push if formatting changes are needed on main/master"
echo ""
echo "To disable auto-commit on feature branches:"
echo "  export DISABLE_AUTO_COMMIT=1"
echo ""
echo "To bypass the hook entirely:"
echo "  git push --no-verify"
