# Repository Guidelines

## Project Structure & Module Organization

- `cmd/fecim-lattice-tools/` is the unified GUI entry point.
- `module1-hysteresis/` through `module7-docs/` hold the seven feature modules (physics, crossbar, MNIST, circuits, comparison, EDA, docs).
- `module6-eda/` is the design suite and has its own `cmd/eda-gui`, `cmd/eda-cli`, and `Makefile`.
- `shared/` contains common packages (theme, widgets, logging, physics).
- `docs/`, `data/`, `config/`, `cells/` store documentation and static assets; `logs/` is runtime output.

## Build, Test, and Development Commands

```bash
./launch.sh                          # Build and run the main app
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
go test ./...                        # Full test suite
go test -race ./...                  # Race detector (use for concurrency changes)
make -C module6-eda build            # Module 6 build
make -C module6-eda run              # Module 6 GUI
make -C module6-eda cli              # Module 6 CLI sample run
```

## Coding Style & Naming Conventions

- Go code must be `gofmt`-formatted and follow Effective Go conventions.
- Use clear, domain-specific names (e.g., `QuantizeTo30Levels`, `IRDropModel`).
- Tests follow Go conventions: files named `*_test.go` with `TestXxx` functions.
- Fyne UI updates from goroutines must be wrapped in `fyne.Do(func() { ... })`.

## Testing Guidelines

- Run `go test ./...` before opening a PR; scope to a module when iterating (e.g., `go test ./module2-crossbar/...`).
- New features should include tests and maintain >80% coverage where practical.
- Add targeted tests for physics or quantization changes to avoid regressions.

## Commit & Pull Request Guidelines

- Commit messages use `type: subject` (types: `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `chore`).
- Include a short body when the change is non-trivial and mention tests run.
- For new features, open an issue with `[Feature]` prefix for discussion first; for bugs include reproduction steps (`[Bug]`).

## Agent-Specific Notes

- Start with `CLAUDE.md` and `docs/development/scriptReference.md` for task-specific guidance.
- Do not commit generated artifacts (binaries, `logs/`, `output/`, `docs/archive/`).
