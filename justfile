set shell := ["bash", "-euo", "pipefail", "-c"]
version := `tr -d '[:space:]' < VERSION`
default:
    @just --list
run *args:
    go run . {{args}}
fmt:
    gofmt -w .
fmt-check:
    @test -z "$(gofmt -l .)"
vet:
    go vet ./...
test:
    go test -count=1 ./...
precommit: fmt-check vet test
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
release: precommit
    mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -o dist/colima-cli-darwin-amd64 .
    GOOS=darwin GOARCH=arm64 go build -o dist/colima-cli-darwin-arm64 .
    GOOS=linux GOARCH=amd64 go build -o dist/colima-cli-linux-amd64 .
clean:
    rm -rf dist
