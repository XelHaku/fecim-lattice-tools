# CLAUDE.md - FeCIM Visualizer

## Overview

Go-based visualization suite for Ferroelectric Compute-in-Memory (FeCIM) technology demonstrating Dr. external research group's HfO2-ZrO2 superlattice research.

**Core concept**: 30 discrete analog states per cell (~4.9 bits/cell) enabling compute-in-memory where the same device handles both storage and computation.

## Build & Run

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
# Or: ./launch.sh
```

## Project Structure

```
cmd/fecim-visualizer/     # Main unified app entry point
module1-hysteresis/       # P-E curve, Preisach model
module2-crossbar/         # MVM, non-idealities (IR drop, sneak paths, drift)
module3-mnist/            # Neural network digit recognition (87% accuracy)
module4-circuits/         # DAC/ADC/TIA peripherals
module5-comparison/       # Technology comparison
module6-eda/              # EDA tools
shared/                   # Theme, widgets, logging
```

Each module follows: `pkg/gui/embedded.go` (embeddable app), `pkg/gui/app.go` (standalone).

## Key Rules

### Do
- Use `fyne.Do(func() { ... })` for all UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow the embedded app interface pattern (see below)
- Run `go test ./...` before committing

### Don't
- Modify `module2-crossbar/pkg/_layers_experimental/` - archived research code
- Add demos without implementing the embedded interface
- Use blocking operations on the main UI thread
- Commit binaries (fecim-visualizer, crossbar-gui, etc.)

## Embedded App Interface

Every demo must implement:
```go
type EmbeddedXxxApp struct { ... }
func NewEmbeddedXxxApp() *EmbeddedXxxApp
func (app *EmbeddedXxxApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject
func (app *EmbeddedXxxApp) Start()  // Called when tab selected
func (app *EmbeddedXxxApp) Stop()   // Called when tab deselected
```

## Key Files

| Task | File |
|------|------|
| Add demo | `cmd/fecim-visualizer/main.go` |
| Crossbar MVM | `module2-crossbar/pkg/crossbar/array.go` |
| Hysteresis | `module1-hysteresis/pkg/ferroelectric/preisach.go` |
| Theme | `shared/theme/theme.go` |
| Non-idealities | `module2-crossbar/pkg/crossbar/nonidealities.go` |

## Physics Constants

| Parameter | Value | Description |
|-----------|-------|-------------|
| FeCIM Levels | 30 | Discrete states per cell |
| Pr | ~25 µC/cm² | Remanent polarization |
| Ec | ~1 MV/cm | Coercive field |
| Ps | ~30 µC/cm² | Saturation polarization |

## Testing

```bash
go test ./...                            # All tests
go test ./module2-crossbar/pkg/crossbar  # Crossbar only
go test -v -run TestPreisach             # Specific test
```

## Dependencies

- **Fyne v2.7.2** - GUI framework
- **Charm/BubbleTea** - TUI (demo1)
- **go-gl/glfw** - Window management
- **vulkan-go/vulkan** - GPU rendering

## Git Conventions

- Commit messages: `type: description` (feat, fix, docs, refactor, test, chore)
- Keep commits atomic and focused
- Run tests before pushing

## Ignore These Directories

- `logs/` - Runtime logs
- `output/` - Generated exports
- `module2-crossbar/pkg/_layers_experimental/` - Archived research
- `docs/archive/` - Archived documentation
