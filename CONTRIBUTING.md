# Contributing to godexer

Thanks for your interest in improving godexer! This document is a lightweight starter you can extend. It outlines the basics to report issues, propose changes, and develop locally.

## Code of Conduct

Be respectful and constructive. Treat others as you’d like to be treated. Harassment or discrimination is not tolerated.

## Ways to contribute

- Report bugs and request features via GitHub Issues
- Improve documentation (README, examples, comments)
- Triage issues, reproduce problems, and share workarounds
- Submit pull requests for fixes and enhancements

## Reporting issues

Before filing a new issue:
- Search existing issues to avoid duplicates
- Include minimal, reproducible examples (code snippets and the scenario YAML if relevant)
- Provide environment details: OS, Go version (`go version`), and the library version (commit or tag)
- Share expected vs actual behavior and logs if possible

Link: https://github.com/go-extras/godexer/issues

## Development setup

Requirements:
- Go 1.25+ (CI runs on 1.25.x)
- golangci-lint (optional, recommended)

Common tasks:

```sh
# Ensure dependencies are in good shape
go mod tidy

# Run tests (race detector and coverage recommended)
go test -race -v ./...

# Lint (if you have golangci-lint installed)
golangci-lint run
```
## Dependencies

- Use `go get` to add or update dependencies; do not edit `go.mod` by hand.
- Run `go mod tidy` and commit both `go.mod` and `go.sum`.


Examples:

```sh
# Run the local example
go run ./example/local

# See example/ssh/main.go for flags used to run the SSH example
```

## Testing guidelines

- Prefer table-driven tests
- Use frankban/quicktest with the alias `qt` (import as `qt`).
- In table-driven tests, use `t.Run()` for subtests; within each subtest initialize a new quicktest instance: `c := qt.New(t)`.
- Prefer `qt.IsNotNil(value)` over `qt.Not(qt.IsNil(value))` for nil checks.
- Keep happy-path and error-path cases clearly separated (dedicated subtests or distinct test functions).
- Write tests in `*_test.go` files and the `_test` package when practical.

## Style and conventions

- Go formatting: `gofmt` or `goimports`.
- Keep public APIs documented with GoDoc comments.
- Write all code comments and documentation in English.
- Maintain clear, minimal, and focused changes.
- Add comments for non-obvious logic.

## Branching and pull requests

- Create feature branches from `master`.
- Keep pull requests small and focused; include tests and docs updates.
- Ensure `go test ./...` passes locally; run `go mod tidy` and address lint warnings.
- CI runs build, unit tests with race and coverage (>=80% enforced), uploads coverage to Codecov, and runs lint (golangci-lint) and govulncheck; PRs must pass all checks.
- Use descriptive PR titles and summaries (include context and rationale).

## Commit messages

- Use clear, imperative-style messages (e.g., "Add X", "Fix Y")
- Reference issues when applicable (e.g., "Fixes #123")

## Versioning and releases

- The project follows semantic versioning where practical; current status is Alpha and may include breaking changes

## Security

If you discover a security issue, please report it privately via issues marked as security or contact a maintainer directly. Avoid public disclosure until a fix is available.

## License

By contributing, you agree that your contributions will be licensed under the project’s MIT License.

