<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# FeCIM Lattice Tools

## Purpose

Go monorepo for ferroelectric compute-in-memory (FeCIM) simulation and visualization. Built on published physics (Materlik 2015, Park 2015, Alessandri 2018, Guo 2018) with integrated modules for hysteresis modeling, crossbar array simulation, neural network inference, peripheral circuits, technology comparison, and EDA export. Education phase (simulation-only). Seven integrated modules plus shared infrastructure serve researchers, graduate students, and device engineers.

## Key Files

| File | Description |
|------|-------------|
| `CLAUDE.md` | AI agent instructions, project conventions, and accuracy policy |
| `go.mod` | Go 1.24.0+ module definition with Fyne v2, charmbracelet, Vulkan, GLFW |
| `go.sum` | Dependency hash verification |
| `launch.sh` | Build and run script |
| `README.md` | Project overview, features, physics models, and quick start |
| `CONTRIBUTING.md` | Contribution guidelines |
| `LICENSE` | MIT license |

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| `cmd/` | CLI entry points: `fecim-lattice-tools` (main GUI), `fecim-screenshotter` (headless testing), `latex-svg` (doc rendering) |
| `module1-hysteresis/` | P-E curves, Preisach model, Landau-Khalatnikov solver, ISPP write controller, material presets (HZO, BTO, PZT) |
| `module2-crossbar/` | Crossbar array MVM (matrix-vector multiply), non-idealities (IR drop, sneak paths, conductance drift), device models |
| `module3-mnist/` | End-to-end MNIST inference through CIM pipeline, 80% accuracy baseline, cross-validation with external benchmarks |
| `module4-circuits/` | DAC/ADC/TIA peripheral circuits, read/program paths, front-end behavior, circuit abstractions |
| `module5-comparison/` | Technology comparison views across operating conditions, design assumptions, and performance metrics |
| `module6-eda/` | EDA pipeline with OpenLane integration, SPICE/Verilog/Liberty/DEF/LEF export, netlist generation |
| `module7-docs/` | Integrated documentation viewer, references, physics explanations, educational materials |
| `shared/` | Common packages: theme (Fyne styling), widgets (custom UI components), physics (shared equations), logging, utilities |
| `validation/` | Benchmarks, calibration data, regression test suite, golden files, physics validation harnesses |
| `config/` | Physics configuration files, material property presets, simulation defaults |
| `data/` | Calibration data, crossbar presets, Preisach state files, lookup tables |
| `scripts/` | CI/CD scripts, toolchain setup, build automation |
| `tools/` | External tool integrations, utility binaries |
| `docs/` | Project documentation: architecture, testing guide, GUI notes, Fyne patterns, EDA integration, script reference, video transcripts |
| `experimental-data/` | Real device measurements: HZO, HfO2, crossbar characterization data |
| `cells/` | Standard cell library for EDA export |
| `examples/` | Example projects and use cases |
| `references/` | Academic papers, simulation benchmarks, comparison harnesses |
| `artifacts/` | Build outputs, logs (gitignored) |
| `screenshots/` | UI screenshots and graphics |

## For AI Agents

### Quick Reference

**Full detailed API reference:** `docs/archive/old-structure/development/SCRIPT_REFERENCE.md`

| I need to... | Look in |
|--------------|---------|
| Find a function signature | `docs/archive/old-structure/development/SCRIPT_REFERENCE.md#quick-function-lookups` |
| Understand an error | `docs/archive/old-structure/development/SCRIPT_REFERENCE.md#error-resolution-guide` |
| Implement a new feature | `docs/archive/old-structure/development/SCRIPT_REFERENCE.md#decision-trees` |
| Check thread safety | `docs/archive/old-structure/development/SCRIPT_REFERENCE.md#thread-safety-guide` |
| Fix Fyne GUI issues | `docs/3-develop/gui/FYNE_NOTES.md` |
| Run tests | `docs/3-develop/testing/TESTING.md` |
| Review UI analysis | `docs/3-develop/HYPER_ANALYSIS_REPORT.md` |
| Use EDA pipeline | `docs/eda/README.md` and `docs/eda/guides/integration.md` |
| Access EDA CLI | `docs/eda/references/cli-reference.md` |

If `qmd` emits CUDA build output or starts model downloads, stop using it for that cycle and fall back to `rg`/direct doc reads. Do not stall active validation on local-search bootstrap.

### Working in This Repository

**Build:**
```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./launch.sh
```

**Test:**
```bash
go test ./...                           # Full test suite
go test -race ./...                     # Race condition detection
go test ./module2-crossbar/...          # Module-scoped testing
```

**Key Patterns:**
- **All UI updates from goroutines:** Use `fyne.Do(func() { ... })` (non-blocking, thread-safe)
- **Conductance quantization:** Call `crossbar.QuantizeTo30Levels(value)` for simulation baseline (30 discrete levels, configurable)
- **Module interface:** Every module implements `BuildContent()`, `Start()`, `Stop()` for embedded app integration
- **Physics simulation:** Material presets in `module1-hysteresis/pkg/ferroelectric/material.go`; crossbar defaults in `module2-crossbar/pkg/crossbar/array.go`
- **Write control:** Module 1 and 4 both provide ISPP (In-Situ Pulse Programming) engines; see `MEMORY.md` for architecture

### Critical Behavioral Notes

**From CLAUDE.md:**
- **Simulation baseline, not hardware claim:** The 30-level quantization and all material presets are for education and visualization. Not validated device measurements.
- **Accuracy-first policy:** See `docs/4-research/honesty-audit.md` for verified claims and removed/unverified claims.
- **No blocking UI operations:** Never use `time.Sleep()` or blocking I/O on the Fyne main thread.
- **Commit before pushing:** `go test ./...` must pass.

### Testing Requirements

- **Unit tests:** Each package should have corresponding `*_test.go` files
- **Physics regression golden files:** `validation/testdata/physics_regression/` — compare against existing golden data
- **Integration tests:** Module interaction tests in `cmd/fecim-lattice-tools/mode_*_test.go`
- **Test frameworks:** `testing` (standard), `stretchr/testify` for assertions
- **CI status:** See `.github/workflows/` or latest test run

### Common Debugging Patterns

**Known Bug Patterns (from MEMORY.md):**
1. **Guard-band sign direction flip** — limit guard pulses to 2 max, clamp calcLevel to prevent overshoot
2. **Bounds collapse [VMin, VMax]** — widen minimally using direction info, don't reset to full range
3. **ACCEPT ±1 guard interaction** — skip ACCEPT ±1 when guardActive=true, raise overshoot threshold from 3 to 8
4. **Zero-field bounds reset** — when absField < 0.01*Ec, reset to full [0, MaxField]
5. **Preisach Everett zero-clamp** — use product-form (always non-negative) instead of factorized (goes negative)

**GUI freeze diagnosis:**
- Check `module1-hysteresis/pkg/gui/gui.go` and `simulation.go` for Fyne.Do wrapping
- Check for blocking operations in render loops
- See `docs/3-develop/gui/FYNE_NOTES.md` for Fyne threading model

### Commit Style

```
type: description

Optional longer explanation if needed.

Types: feat, fix, docs, refactor, test, chore
```

Example:
```
test(ispp): add convergence ensemble for LK solver

Tests 9 materials with 50 targets each, verifies ACCEPT ±1
threshold interaction with guard-band logic.
```

## Dependencies

### Go Version
- **Go 1.24.0+** (toolchain 1.24.12)

### External (Direct)
- **fyne.io/fyne/v2** v2.7.2 — Cross-platform GUI framework (OpenGL rendering)
- **github.com/charmbracelet/bubbles** v0.20.0 — TUI component library
- **github.com/charmbracelet/bubbletea** v1.2.4 — TUI framework (CLI support)
- **github.com/charmbracelet/lipgloss** v1.0.0 — TUI styling
- **github.com/go-gl/glfw/v3.3/glfw** v0.0.0-20250301202403-da16c1255728 — Window management
- **github.com/vulkan-go/vulkan** v0.0.0-20221209234627-c0a353ae26c8 — Vulkan GPU acceleration (optional)

### Transitive (Key)
- **gonum.org/v1/gonum** v0.17.0 — Scientific computing (linear algebra, optimization)
- **github.com/yuin/goldmark** v1.7.8 — Markdown rendering (docs)
- **golang.org/x/image** v0.25.0 — Image manipulation
- **golang.org/x/text** v0.23.0 — Unicode/text processing
- **github.com/stretchr/testify** v1.11.1 — Test assertions

### Internal Relationships

```
cmd/fecim-lattice-tools (main entrypoint)
  ↓
shared/ (theme, widgets, physics, logging, utilities)
  ↑
  ├─ module1-hysteresis (P-E, Preisach, LK, ISPP)
  ├─ module2-crossbar (MVM, IR drop, sneak paths)
  ├─ module3-mnist (inference pipeline)
  ├─ module4-circuits (DAC/ADC/TIA, ISPP)
  ├─ module5-comparison (comparative views)
  ├─ module6-eda (SPICE/Verilog/Liberty export)
  └─ module7-docs (documentation viewer)

validation/ (tests against module1, module2, shared)
  ├─ module1-hysteresis (physics regression)
  └─ shared/physics (ISPP, crossbar kernels)
```

**Module 1 provides physics** → used by modules 4, 6, and validation
**Module 2 provides array simulation** → used by modules 3, 5, 6
**Module 3 uses 1+2** → inference pipeline
**Module 4 uses 1** → circuit-level ISPP
**Module 5 uses 1+2** → comparison analysis
**Module 6 uses 1+2** → EDA export
**Module 7 is standalone** → documentation

## Module Structure (Pattern)

Each major module follows:
```
module{N}-{name}/
  ├─ cmd/          (optional: CLI entry)
  ├─ pkg/          (main packages)
  ├─ shaders/      (optional: GPU code)
  ├─ README.md     (module-specific docs)
  └─ *_test.go     (tests alongside source)
```

**Shared interface:**
```go
type Tab interface {
    BuildContent() fyne.CanvasObject
    Start()
    Stop()
}
```

Every module tab in the GUI implements this interface for lifecycle management.

## Physics Models Summary

**Hysteresis (Module 1):**
- Preisach model with Everett function (product-form, non-negative)
- Landau-Khalatnikov solver for switching dynamics
- Material presets: HZO (hafnium zirconium oxide), BTO (barium titanate), PZT (lead zirconate titanate)
- Published parameters: Materlik 2015, Park 2015, Alessandri 2018

**Crossbar Array (Module 2):**
- Parallel conductance update with quantization (30 levels default)
- IR drop simulation (Ohm's law across array resistance)
- Sneak-path currents (Kirchhoff-law verification)
- Conductance drift over time
- Word-line select logic for row/column addressing

**ISPP Write Control (Modules 1 & 4):**
- Binary search convergence with pulse amplitude tuning
- Guard-band protection against overshoot
- State machine: APPLY → WAIT → VERIFY → loop
- Two engines: waveform-based (Module 1) and level-based (Module 4)
- Shared L-K solver in `shared/physics/ispp_write.go`

**MNIST (Module 3):**
- End-to-end inference: input quantization → crossbar MVM → output mapping
- 80% baseline accuracy on MNIST test set
- Validated against external benchmarks (HZO FTJ paper)

## Testing Infrastructure

**Test Files:**
- `*_test.go` files alongside source code
- `cmd/fecim-lattice-tools/mode_*_test.go` — integration tests
- `validation/testdata/` — golden files for regression

**Test Patterns:**
- Physics regression: compare computed states to golden JSON
- ISPP convergence: ensemble tests across materials
- Crossbar validation: Kirchhoff-law verification
- GUI tests: render cycle and state machine checks

**Run Tests:**
```bash
go test ./...              # All tests
go test -v ./...           # Verbose
go test -race ./...        # Race detection
go test -cover ./...       # Coverage
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...  # Regenerate golden
```

## Common Tasks

### Add a New Feature
1. **Plan:** Read `docs/archive/old-structure/development/SCRIPT_REFERENCE.md#decision-trees`
2. **Implement:** Follow module pattern; use `fyne.Do()` for UI updates
3. **Test:** Add `*_test.go` with unit and integration tests
4. **Verify:** `go test ./...` passes, `go test -race ./...` passes
5. **Commit:** `type: description` with test evidence

### Debug a GUI Freeze
1. Check `shared/physics/` and module `pkg/gui/` for blocking operations
2. Verify all UI updates use `fyne.Do(func() { ... })`
3. See `docs/3-develop/gui/FYNE_NOTES.md`
4. Review `docs/3-develop/HYPER_ANALYSIS_REPORT.md` for UI critique

### Fix a Physics Test Failure
1. Read the test output — identify which material/level failed
2. Check golden file in `validation/testdata/physics_regression/`
3. Run with `FECIM_UPDATE_PHYSICS_GOLDEN=1` if regression is intended
4. Otherwise, debug the solver in `module1-hysteresis/pkg/ferroelectric/`

### Export to EDA Tools
1. See `module6-eda/` and `docs/eda/guides/integration.md`
2. Module 6 generates SPICE (circuit), Verilog (logic), Liberty (timing), DEF/LEF (physical)
3. Integration with OpenLane via CLI in `docs/eda/references/cli-reference.md`

## CI/CD

**Pre-commit:**
```bash
go test ./...
go test -race ./...
go fmt ./...
```

**Build artifacts:** `.github/workflows/` (if GitHub Actions configured)

## Accuracy & Honesty Policy

**Verified claims (peer-reviewed):**
- 98.24% MNIST accuracy in HZO FTJ reservoir computing (Journal of Alloys and Compounds 2025) — NOT a FeCIM device claim

**Simulation defaults (educational):**
- 30 analog conductance levels (configurable)
- Material parameters (HZO, BTO, PZT) from published physics, explicitly marked
- Crossbar non-idealities (IR drop, sneak paths) from literature

**Removed/unverified claims:**
- 30 analog states for Tour device (conference-only reference, not reported in literature)
- 87% MNIST accuracy (conference-only reference)
- Energy multipliers vs NAND/GPUs without published measurement evidence

**Full audit:** `docs/4-research/honesty-audit.md`

## License

MIT License — see `LICENSE`

---

<!-- MANUAL: Any manually added notes below this line are preserved on regeneration -->
