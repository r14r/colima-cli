set shell := ["bash", "-euo", "pipefail", "-c"]
version := `tr -d '[:space:]' < VERSION`
default:
    @just --list

run *args:
    go run . {{args}}

fmt:
    gofmt -w .

fmt-check:
    bash scripts/checks.sh format

vet:
    bash scripts/checks.sh vet

test:
    bash scripts/checks.sh test

race:
    bash scripts/checks.sh race

lint:
    bash scripts/checks.sh lint

security:
    bash scripts/checks.sh security

tools:
    go install github.com/rhysd/actionlint/cmd/actionlint@latest
    go install golang.org/x/vuln/cmd/govulncheck@latest

hooks-install:
    chmod +x .githooks/pre-commit .githooks/pre-push
    git config --local core.hooksPath .githooks
    @echo "Git pre-commit and pre-push hooks installed"

precommit:
    bash scripts/checks.sh precommit

prepush:
    bash scripts/checks.sh all

check: precommit

ci: prepush

build: precommit
    mkdir -p dist
    go build -trimpath -ldflags "-s -w" -o dist/colima-cli .

install: build
    mkdir -p "${INSTALL_DIR:-/Users/Shared/CLOUD/DeveloperTools/bin}"
    install -m 0755 dist/colima-cli "${INSTALL_DIR:-/Users/Shared/CLOUD/DeveloperTools/bin}/colima-cli"

deploy: install

status:
    go run . status

doctor:
    go run . doctor

release: prepush
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -o dist/colima-cli-darwin-amd64 .
    GOOS=darwin GOARCH=arm64 go build -o dist/colima-cli-darwin-arm64 .
    GOOS=linux GOARCH=amd64 go build -o dist/colima-cli-linux-amd64 .

clean:
    rm -rf dist
