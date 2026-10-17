# colima-cli developer commands
set shell := ["bash", "-euo", "pipefail", "-c"]

version := `tr -d '[:space:]' < VERSION`
binary := "colima-cli"
install_dir := env_var_or_default("INSTALL_DIR", "/usr/local/bin")

# -----------------------------------------------------------------------------
# Help and CLI usage
# -----------------------------------------------------------------------------

# List available development commands.
default:
    @just --list

# Run the CLI directly from source.
run *args:
    go run . {{args}}

# Show Colima status.
status:
    go run . status

# Diagnose Colima and Docker.
doctor:
    go run . doctor

# -----------------------------------------------------------------------------
# Project setup and dependencies
# -----------------------------------------------------------------------------

# Set up Python tooling, Go modules, analyzers and Git hooks.
bootstrap:
    @command -v go >/dev/null || { echo "Go is required" >&2; exit 1; }
    @command -v python3 >/dev/null || { echo "Python 3 is required" >&2; exit 1; }
    @command -v git >/dev/null || { echo "Git is required" >&2; exit 1; }
    python3 -m venv .venv
    .venv/bin/python -m pip install --upgrade pip pre-commit
    go mod download
    just tools
    just hooks-install

# Download Go modules.
deps:
    go mod download

# Install external Go quality-check tools.
tools:
    go install github.com/rhysd/actionlint/cmd/actionlint@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest

# Set up version-controlled Git hooks for this clone.
hooks-install:
    chmod +x .githooks/pre-commit .githooks/pre-push
    git config --local core.hooksPath .githooks
    @echo "Enabled Git hooks from .githooks"

# Restore Git's default hooks directory.
hooks-uninstall:
    git config --local --unset core.hooksPath || true

# -----------------------------------------------------------------------------
# Formatting, validation and tests
# -----------------------------------------------------------------------------

# Format Go source.
fmt:
    gofmt -w .

# Verify Go formatting without changing files.
fmt-check:
    bash scripts/checks.sh format

# Run go vet.
vet:
    bash scripts/checks.sh vet

# Run unit tests without cached results.
test:
    bash scripts/checks.sh test

# Run tests under the Go race detector.
test-race:
    bash scripts/checks.sh race

# Alias for test-race.
race: test-race

# Validate GitHub Actions workflows.
lint:
    bash scripts/checks.sh lint

# Check dependencies for known vulnerabilities.
security:
    bash scripts/checks.sh security

# Run the checks used before a commit.
precommit:
    bash scripts/checks.sh precommit

# Alternative spelling, consistent with other CLI repositories.
pre-commit: precommit

# Run all quality checks before a push.
prepush:
    bash scripts/checks.sh all

# Execute the complete CI quality gate locally.
precommit-all: prepush

# Run normal local validation.
check: precommit

# Run the same checks as GitHub Actions.
ci: prepush

# -----------------------------------------------------------------------------
# Build, installation and release
# -----------------------------------------------------------------------------

# Build the production binary after validation.
build: precommit
    mkdir -p dist
    go build -trimpath -ldflags "-s -w" -o dist/{{binary}} .

# Build and install into /usr/local/bin (override with INSTALL_DIR).
install: build
    @dir="{{install_dir}}"; \
    if [ -w "$dir" ]; then \
      install -m 0755 dist/{{binary}} "$dir/{{binary}}"; \
    else \
      sudo mkdir -p "$dir"; \
      sudo install -m 0755 dist/{{binary}} "$dir/{{binary}}"; \
    fi; \
    echo "Installed {{binary}} {{version}} to $dir/{{binary}}"

# Alias for install.
deploy: install

# Build platform-specific macOS and Linux binaries.
release: prepush
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -o dist/{{binary}}-darwin-amd64 .
    GOOS=darwin GOARCH=arm64 go build -o dist/{{binary}}-darwin-arm64 .
    GOOS=linux GOARCH=amd64 go build -o dist/{{binary}}-linux-amd64 .

# Remove generated binaries.
clean:
    rm -rf dist
