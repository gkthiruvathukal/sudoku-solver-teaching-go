# Repository Guidelines

## Project Structure & Module Organization

This is a single-package Go CLI project. Core source files live at the repository root:

- `main.go` — command-line (`solve`), interactive, and TUI entry points.
- `solver.go` — Sudoku state, backtracking solver, traversal strategies, `SolutionDiff`, and related utilities.
- `tui.go` — Bubble Tea terminal UI, async solve/trace commands, and progress rendering.
- `gset.go` — generic set utilities used for row/column/nonet bookkeeping.

Tests sit beside the code in `solver_test.go`, `gset_test.go`, and `tui_test.go`. Puzzle fixtures and larger sample datasets are in `data/`. Helper scripts `go-sudoku.sh` and `go-sudoku-quick.sh` build the binary and compare both traversal strategies side by side across fixture files.

## Build, Test, and Development Commands

- `go build -o ./sudoku_solver .` builds the local CLI binary.
- `go test ./...` runs all unit tests in the module.
- `go test -cover ./...` reports package coverage.
- `./sudoku_solver solve --puzzle <81-digit-puzzle>` runs the unattended solver (default strategy: `row-major`).
- `./sudoku_solver solve --puzzle <81-digit-puzzle> --strategy nonet-first` runs with the nonet-first strategy.
- `./sudoku_solver interactive --puzzle <81-digit-puzzle>` starts the teaching-oriented interactive CLI.
- `./sudoku_solver tui --puzzle <81-digit-puzzle> [--strategy row-major|nonet-first]` opens the terminal UI.
- `./go-sudoku-quick.sh` builds the binary, runs both strategies on `data/*-sample-*.txt`, and reports wins/losses/ties.
- `./go-sudoku.sh` runs the same comparison on the larger `data/*-[a-b].txt` dataset (no build step).

Use `go mod tidy` after dependency changes and commit resulting `go.mod` or `go.sum` updates.

## Traversal Strategies

The solver separates *position order* from *solver logic*. `solvePositions(positions []int, ...)` is the single generic recursive solver; it takes a pre-computed position list so the strategy decision happens at call time, not inside recursion.

- **`RowMajorPositions()`** — returns positions 0–80 in reading order (left-to-right, top-to-bottom). Easy to reason about when teaching backtracking.
- **`NonetFirstPositions()`** — sorts the nine 3×3 nonets by descending clue count, then emits all cells in each nonet in row-major order. The idea is that denser nonets have fewer candidates, so bad branches are pruned earlier.

The `solve --strategy` flag selects the strategy for the unattended solver. The TUI exposes `/strategy [row-major|nonet-first]` at runtime. The comparison scripts run both strategies on every puzzle and report which finishes faster (by nanosecond count).

Practical note: because the full 9×9 board fits in L1 cache, both strategies perform similarly on modern hardware. The pedagogical value is in comparing placement and backtrack counts, not wall-clock time.

## Coding Style & Naming Conventions

Follow standard Go style. Format all Go files with `gofmt` before committing; tabs are expected for indentation in Go source. Keep package declarations as `package main` unless the project is intentionally split into packages. Use exported names only for API-like concepts already used across files, such as `NewSudoku`, `Solve`, or `NewSet`; keep helpers lower-case when local to the package. Prefer clear, table-driven tests for validation cases.

## Testing Guidelines

Tests use Go’s standard `testing` package. Name test files `*_test.go` and test functions `TestXxx`. Use `t.Run` for grouped edge cases, especially invalid puzzle input, bounds checks, and solver behavior. When changing solver logic, include at least one known puzzle and solution assertion, and run both `go test ./...` and `./go-sudoku-quick.sh` when fixture behavior may be affected.

When adding or changing traversal strategies, add a `TestXxxPositionsCoversAllCells` test that verifies all 81 positions appear exactly once, and a `TestSolveWithOrderXxx` test that confirms the strategy produces a valid solution for a known puzzle. Benchmark tests (`BenchmarkSolveRowMajor`, `BenchmarkSolveNonetFirst`) live in `solver_test.go` and can be run with `go test -bench=. -benchtime=5s`.

## Commit & Pull Request Guidelines

Recent commits use short, imperative subject lines such as `Expand generic set utilities` and `Add solver edge case tests`. Keep commit messages focused on the observable change. Pull requests should include a concise description, commands run, and any relevant puzzle examples or fixture files touched. Link issues when applicable. For CLI output changes, include before/after snippets rather than screenshots.

## Agent-Specific Instructions

Do not rewrite fixture data unless the task explicitly requires it. Keep changes scoped and preserve the pedagogical readability of the solver; this repository is meant to teach recursion and backtracking, not just optimize runtime.
