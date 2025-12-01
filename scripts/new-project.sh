#!/usr/bin/env bash
#
# new-project.sh - Scaffold a new project from hexagonal-go template
#
# Usage:
#   ./scripts/new-project.sh <new-module-name> [target-directory]
#
# Examples:
#   ./scripts/new-project.sh github.com/myorg/my-saas
#   ./scripts/new-project.sh github.com/myorg/my-saas ~/projects/my-saas
#
set -euo pipefail

# ============================================================================
# Configuration
# ============================================================================

OLD_MODULE="github.com/0xsj/hexagonal-go"
OLD_APP_NAME="hexagonal-go"

# ============================================================================
# Colors
# ============================================================================

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ============================================================================
# Helper Functions
# ============================================================================

info() {
    echo -e "${BLUE}ℹ${RESET} $1"
}

success() {
    echo -e "${GREEN}✓${RESET} $1"
}

warn() {
    echo -e "${YELLOW}⚠${RESET} $1"
}

error() {
    echo -e "${RED}✗${RESET} $1" >&2
}

die() {
    error "$1"
    exit 1
}

# ============================================================================
# Validation
# ============================================================================

if [[ $# -lt 1 ]]; then
    echo -e "${BOLD}Usage:${RESET} $0 <new-module-name> [target-directory]"
    echo ""
    echo -e "${BOLD}Examples:${RESET}"
    echo "  $0 github.com/myorg/my-saas"
    echo "  $0 github.com/myorg/my-saas ~/projects/my-saas"
    echo ""
    echo -e "${BOLD}Arguments:${RESET}"
    echo "  new-module-name   The Go module name (e.g., github.com/myorg/my-app)"
    echo "  target-directory  Optional. Where to create the project (default: current directory)"
    exit 1
fi

NEW_MODULE="$1"
TARGET_DIR="${2:-.}"

# Validate module name format
if [[ ! "$NEW_MODULE" =~ ^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$ ]]; then
    die "Invalid module name format. Expected: domain/org/project (e.g., github.com/myorg/my-app)"
fi

# Extract app name from module (last segment)
NEW_APP_NAME=$(basename "$NEW_MODULE")

# ============================================================================
# Main
# ============================================================================

echo ""
echo -e "${BOLD}${CYAN}Hexagonal Go - Project Scaffolding${RESET}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
info "Old module: ${OLD_MODULE}"
info "New module: ${NEW_MODULE}"
info "App name:   ${NEW_APP_NAME}"
info "Target:     ${TARGET_DIR}"
echo ""

# ============================================================================
# Copy to Target Directory (if specified)
# ============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$(dirname "$SCRIPT_DIR")"

if [[ "$TARGET_DIR" != "." ]]; then
    if [[ -d "$TARGET_DIR" ]]; then
        die "Target directory already exists: $TARGET_DIR"
    fi
    
    info "Copying template to ${TARGET_DIR}..."
    mkdir -p "$TARGET_DIR"
    
    # Copy everything except .git, tmp, bin, vendor
    rsync -a \
        --exclude='.git' \
        --exclude='tmp' \
        --exclude='bin' \
        --exclude='vendor' \
        --exclude='coverage.out' \
        --exclude='*.log' \
        "$TEMPLATE_DIR/" "$TARGET_DIR/"
    
    success "Template copied"
    
    cd "$TARGET_DIR"
else
    # Working in-place
    if [[ ! -f "go.mod" ]]; then
        die "go.mod not found. Run this script from project root or specify target directory."
    fi
    
    warn "Working in-place (current directory)"
fi

# ============================================================================
# Replace Module in Go Files
# ============================================================================

info "Updating Go imports..."

# Find all .go files and replace import paths
GO_FILES=$(find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./tmp/*")
FILE_COUNT=0

for file in $GO_FILES; do
    if grep -q "$OLD_MODULE" "$file" 2>/dev/null; then
        # Use different sed syntax for macOS vs Linux
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s|${OLD_MODULE}|${NEW_MODULE}|g" "$file"
        else
            sed -i "s|${OLD_MODULE}|${NEW_MODULE}|g" "$file"
        fi
        ((FILE_COUNT++)) || true
    fi
done

success "Updated ${FILE_COUNT} Go files"

# ============================================================================
# Update go.mod
# ============================================================================

info "Updating go.mod..."

if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^module ${OLD_MODULE}|module ${NEW_MODULE}|g" go.mod
else
    sed -i "s|^module ${OLD_MODULE}|module ${NEW_MODULE}|g" go.mod
fi

success "Updated go.mod"

# ============================================================================
# Update .env
# ============================================================================

if [[ -f ".env" ]]; then
    info "Updating .env..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|^APP_NAME=.*|APP_NAME=${NEW_APP_NAME}|g" .env
    else
        sed -i "s|^APP_NAME=.*|APP_NAME=${NEW_APP_NAME}|g" .env
    fi
    
    success "Updated .env"
fi

# ============================================================================
# Update docker-compose.yml
# ============================================================================

DOCKER_COMPOSE="deployments/docker/docker-compose.yml"

if [[ -f "$DOCKER_COMPOSE" ]]; then
    info "Updating docker-compose.yml..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" "$DOCKER_COMPOSE"
    else
        sed -i "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" "$DOCKER_COMPOSE"
    fi
    
    success "Updated docker-compose.yml"
fi

# ============================================================================
# Update .air.toml
# ============================================================================

if [[ -f ".air.toml" ]]; then
    info "Updating .air.toml..."
    # No changes needed unless app-specific
    success "Checked .air.toml"
fi

# ============================================================================
# Update Makefile
# ============================================================================

if [[ -f "Makefile" ]]; then
    info "Updating Makefile..."
    
    # Update any hardcoded references
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" Makefile
    else
        sed -i "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" Makefile
    fi
    
    success "Updated Makefile"
fi

# ============================================================================
# Update README
# ============================================================================

if [[ -f "README.md" ]]; then
    info "Updating README.md..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|${OLD_MODULE}|${NEW_MODULE}|g" README.md
        sed -i '' "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" README.md
    else
        sed -i "s|${OLD_MODULE}|${NEW_MODULE}|g" README.md
        sed -i "s|${OLD_APP_NAME}|${NEW_APP_NAME}|g" README.md
    fi
    
    success "Updated README.md"
fi

# ============================================================================
# Update Swagger Docs
# ============================================================================

SWAGGER_MAIN="cmd/api/main.go"

if [[ -f "$SWAGGER_MAIN" ]]; then
    info "Updating Swagger annotations..."
    
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|Hexagonal Go API|${NEW_APP_NAME} API|g" "$SWAGGER_MAIN"
    else
        sed -i "s|Hexagonal Go API|${NEW_APP_NAME} API|g" "$SWAGGER_MAIN"
    fi
    
    success "Updated Swagger annotations"
fi

# ============================================================================
# Clean Up
# ============================================================================

info "Cleaning up..."

# Remove old .git if exists
if [[ -d ".git" ]]; then
    rm -rf .git
    success "Removed old .git directory"
fi

# Remove generated files that need regeneration
rm -rf tmp/ bin/ vendor/ 2>/dev/null || true
rm -f docs/swagger/swagger.json docs/swagger/swagger.yaml docs/swagger/docs.go 2>/dev/null || true

success "Cleaned generated files"

# ============================================================================
# Initialize New Git Repo
# ============================================================================

info "Initializing new git repository..."

git init -q
git add .
git commit -q -m "Initial commit: scaffolded from hexagonal-go template"

success "Git repository initialized"

# ============================================================================
# Run go mod tidy
# ============================================================================

info "Running go mod tidy..."

if go mod tidy 2>/dev/null; then
    success "Dependencies resolved"
else
    warn "go mod tidy had warnings (this is usually fine)"
fi

# ============================================================================
# Generate Wire & Swagger
# ============================================================================

info "Regenerating wire code..."

if command -v wire &> /dev/null; then
    if make wire 2>/dev/null; then
        success "Wire code generated"
    else
        warn "Wire generation had issues (run 'make wire' manually)"
    fi
else
    warn "Wire not installed. Run 'go install github.com/google/wire/cmd/wire@latest'"
fi

info "Regenerating swagger docs..."

if command -v swag &> /dev/null; then
    if make swagger 2>/dev/null; then
        success "Swagger docs generated"
    else
        warn "Swagger generation had issues (run 'make swagger' manually)"
    fi
else
    warn "Swag not installed. Run 'go install github.com/swaggo/swag/cmd/swag@latest'"
fi

# ============================================================================
# Done
# ============================================================================

echo ""
echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "${BOLD}${GREEN}  Project scaffolded successfully!${RESET}"
echo -e "${BOLD}${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo ""
echo -e "${BOLD}Next steps:${RESET}"
echo ""
echo "  1. Update .env with your configuration"
echo "  2. Start infrastructure:"
echo "     ${CYAN}make docker-up${RESET}"
echo ""
echo "  3. Run migrations:"
echo "     ${CYAN}make migrate-up${RESET}"
echo ""
echo "  4. Start the server:"
echo "     ${CYAN}make dev${RESET}"
echo ""
echo -e "${BOLD}Project:${RESET} ${NEW_MODULE}"
if [[ "$TARGET_DIR" != "." ]]; then
    echo -e "${BOLD}Location:${RESET} ${TARGET_DIR}"
fi
echo ""