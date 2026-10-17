#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/.."

run() { printf '\n==> %s\n' "$*"; "$@"; }
need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'Missing required tool: %s (run: just tools)\n' "$1" >&2
    exit 127
  fi
}
format() {
  local unformatted
  unformatted="$(gofmt -l .)"
  if [[ -n "$unformatted" ]]; then
    printf 'Go files need formatting:\n%s\n' "$unformatted" >&2
    exit 1
  fi
  run git diff --check
  if git rev-parse --git-dir >/dev/null 2>&1; then run git diff --cached --check; fi
}
modules() {
  run go mod verify
  local diff
  # go mod tidy -diff does not modify go.mod or go.sum.
  run go mod tidy -diff
}
vet() { run go vet ./...; }
test() { run go test -count=1 ./...; }
race() { run go test -race -count=1 ./...; }
build() { run go build ./...; }
lint() { need actionlint; run actionlint -color=false; }
security() { need govulncheck; run govulncheck ./...; }
precommit() {
  format
  modules
  vet
  test
  build
  lint
}
all() {
  precommit
  race
  security
}
case "${1:-all}" in
  format) format ;;
  modules) modules ;;
  vet) vet ;;
  test) test ;;
  race) race ;;
  build) build ;;
  lint) lint ;;
  security) security ;;
  precommit) precommit ;;
  all|prepush|ci) all ;;
  *) echo "Usage: $0 {format|modules|vet|test|race|build|lint|security|precommit|all}" >&2; exit 2 ;;
esac
