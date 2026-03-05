# Repository Guidelines

These guidelines describe how to navigate, build, test, and contribute to `code-clip`. Keep changes small, focused, and aligned with the CLI-first design.

## Project Structure & Module Organization

- `cmd/code-clip/main.go`: CLI entry point and flag parsing.
- `pkg/walker/`: filesystem traversal and ignore handling.
- `pkg/formatter/`: output formats (markdown/plain/xml).
- `pkg/walker/walker_test.go`: unit tests for traversal behavior.
- `bin/`: local build output created by `make build`.
- Root docs: `README.md` for user usage and `TODO.md` for planned work. Misc `test_*.go` files are local experiments and should not be expanded without intent.

## Build, Test, and Development Commands

- `make build`: compile the CLI to `bin/code-clip`.
- `make run`: build and run the local binary.
- `make test`: run all Go tests with verbose output.
- `make install`: install `cmd/code-clip` into `~/go/bin`.
- `go test ./... -v`: direct test runner (same as `make test`).
- Go version is defined in `go.mod` (currently `go 1.24.1`).

## Coding Style & Naming Conventions

- Use standard Go formatting (`gofmt`); tabs are the default indentation.
- Package names stay lowercase (e.g., `walker`, `formatter`).
- Exported identifiers use `CamelCase`; unexported use `mixedCase`.
- Tests live in `*_test.go` with `TestXxx` naming.
- New flags should follow the short/long pattern in `cmd/code-clip/main.go` (e.g., `-o`/`--output-format`).

## Testing Guidelines

- Framework: Go `testing` package only.
- Focus tests on filesystem behavior, ignore rules, and traversal limits.
- Run `make test` or `go test ./... -v` before PRs.
- No coverage threshold is defined; add tests for new behavior.

## Commit & Pull Request Guidelines

- Git history only contains `First commit`, so no established convention exists.
- Use concise, imperative subjects (e.g., `Add xml formatter header`).
- PRs should include: a clear summary, tests run, and example CLI output for user-facing changes (commands and snippets).
- Link related issues when applicable.

## Configuration & Ignore Files

- Traversal respects `.gitignore`, `.ignore`, and `.cursorignore` in the current and ancestor directories.
- Use `--ignore`/`-i` for ad-hoc exclusions during local runs; do not commit secrets or generated artifacts.
