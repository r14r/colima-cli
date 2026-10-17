# Local quality gates

Install the external checkers and enable Git hooks:

```sh
just tools
just hooks-install
```

Commands:

- `just precommit` — formatting, whitespace, module integrity/tidiness, vet, tests, build and GitHub Actions validation.
- `just prepush` or `just ci` — all pre-commit checks plus race tests and vulnerability scanning.
- `just fmt` — explicitly apply gofmt formatting.
- `just tools` — install the two Go-based analyzers in `$(go env GOPATH)/bin` (add this location to PATH).
- `just hooks-install` — configure this clone to use the version-controlled `.githooks` directory.

The hooks run the same script as GitHub Actions. Checks do not change source files. On failures, fix the reported issue rather than bypassing hooks.

The minimum Go version for this repository is specified in `go.mod`; recent versions of the checking tools may require a newer Go toolchain. Go auto-toolchain selection can handle that if enabled.
