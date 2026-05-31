# Repository Guidelines

## Project Structure & Module Organization

This is a single-package Go CLI project. Core source files live at the repository root: `main.go` contains command-line and interactive modes, `solver.go` contains Sudoku state and solving logic, and `gset.go` contains generic set utilities. Tests sit beside the code in `solver_test.go` and `gset_test.go`. Puzzle fixtures and larger sample datasets are in `data/`. Helper scripts `go-sudoku.sh` and `go-sudoku-quick.sh` build and run solver checks against fixture files.

## Build, Test, and Development Commands

- `go build -o ./sudoku_solver .` builds the local CLI binary.
- `go test ./...` runs all unit tests in the module.
- `go test -cover ./...` reports package coverage.
- `./sudoku_solver solve --puzzle <81-digit-puzzle>` runs the unattended solver.
- `./sudoku_solver interactive --puzzle <81-digit-puzzle>` starts the teaching-oriented interactive CLI.
- `./go-sudoku-quick.sh` builds the binary and validates sample puzzles from `data/*-sample-*.txt`.

Use `go mod tidy` after dependency changes and commit resulting `go.mod` or `go.sum` updates.

## Coding Style & Naming Conventions

Follow standard Go style. Format all Go files with `gofmt` before committing; tabs are expected for indentation in Go source. Keep package declarations as `package main` unless the project is intentionally split into packages. Use exported names only for API-like concepts already used across files, such as `NewSudoku`, `Solve`, or `NewSet`; keep helpers lower-case when local to the package. Prefer clear, table-driven tests for validation cases.

## Testing Guidelines

Tests use Go’s standard `testing` package. Name test files `*_test.go` and test functions `TestXxx`. Use `t.Run` for grouped edge cases, especially invalid puzzle input, bounds checks, and solver behavior. When changing solver logic, include at least one known puzzle and solution assertion, and run both `go test ./...` and `./go-sudoku-quick.sh` when fixture behavior may be affected.

## Commit & Pull Request Guidelines

Recent commits use short, imperative subject lines such as `Expand generic set utilities` and `Add solver edge case tests`. Keep commit messages focused on the observable change. Pull requests should include a concise description, commands run, and any relevant puzzle examples or fixture files touched. Link issues when applicable. For CLI output changes, include before/after snippets rather than screenshots.

## Agent-Specific Instructions

Do not rewrite fixture data unless the task explicitly requires it. Keep changes scoped and preserve the pedagogical readability of the solver; this repository is meant to teach recursion and backtracking, not just optimize runtime.
