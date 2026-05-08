# Repository Guidelines

## Project Structure & Module Organization

This is a small Go CLI module named `git-wth` and is a Go rewrite of `git-wtf`. The command entry point and most application flow live in `main.go`, with option parsing in `options.go` and configuration loading in `config.go`. Git command abstractions are under `git/`, terminal color handling is under `ui/`, and shared helpers are under `util/`. Embedded user-facing text assets live in `text/*.txt` and are loaded through `//go:embed`; keep those paths stable unless you update the embed pattern. Tests currently live beside the main package in `main_test.go`. When backwards-incompatible changes or other functionality differences from `git-wtf` are introduced, document them in `README.md`.

## Build, Test, and Development Commands

- `go test ./...`: runs all package tests.
- `go test ./... -cover`: runs tests with coverage reporting.
- `go build -o git-wth .`: builds the local CLI binary.
- `./git-wth --help`: checks the built command's help output.
- `gofmt -w *.go git/*.go ui/*.go util/*.go`: formats Go source files before committing.

## Coding Style & Naming Conventions

Use standard Go formatting with tabs as produced by `gofmt`. Keep package names short and lowercase, matching existing directories such as `git`, `ui`, and `util`. Exported names should be reserved for cross-package interfaces and types, such as `git.Git` and `git.Ref`; keep package-local helpers unexported. Prefer clear table or focused unit tests over broad integration tests when behavior can be exercised without invoking real Git commands.

## Testing Guidelines

Use Go's built-in `testing` package. Name tests `TestXxx` and place them in `*_test.go` files near the code under test. Existing tests focus on option parsing, YAML-like config parsing, and formatting helpers. When adding behavior that shells out to Git, test through the `git.Git` interface with fakes rather than depending on a particular repository state.

## Commit & Pull Request Guidelines

Use concise, imperative commit subjects, for example `Add config parsing tests` or `Fix branch relation output`. Keep each commit scoped to one logical change and include tests when behavior changes. Pull requests should describe the user-visible impact, list validation commands run, and call out changes to CLI flags, config keys such as `integration-branches` or `max_commits`, and embedded text in `text/`.

## Security & Configuration Tips

Do not commit local `.git-wthrc` or `.git-wtfrc` files unless they are intentional examples. Treat Git remote URLs and branch names as user-controlled data in output paths and messages.
