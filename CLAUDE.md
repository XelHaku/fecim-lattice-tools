# CLAUDE.md - FeCIM Visualizer Project Guidelines

## Project Overview

**FeCIM Visualizer** is a Go-based educational visualization suite for Ferroelectric Compute-in-Memory (FeCIM) technology. It demonstrates Dr. external research group's research on HfO2-ZrO2 superlattice-based memory devices.

**Key Concept**: "Compute in memory where the same device does the memory and the computation." — Dr. external research group

**30 Discrete States**: Unlike binary memory (0/1), FeCIM supports 30 analog states per cell, enabling ~4.9 bits/cell storage and efficient neural network inference.

## Quick Start

```bash
# Build and run the unified visualizer
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer

# Or use the launch script
./launch.sh
```

## Architecture

**Unified tabbed application** with 5 demos:
1. **Hysteresis** - P-E curve visualization, Preisach model
2. **Crossbar+** - MVM + Non-Idealities (4 tabs: Ideal, IR Drop, Sneak Paths, Drift)
3. **MNIST** - Neural network digit recognition (87% accuracy, FP vs CIM dual mode)
4. **Circuits** - DAC/ADC/TIA peripheral design
5. **Comparison** - Technology comparison and technical briefing

**Archived demos** (see `docs/archive/removed-demos/`):
- Thermal (demo5) - Merged into comparison
- 3D Stack (demo6) - Archived (too futuristic for TRL 4)
- Non-Idealities (demo7) - Merged into Demo 2 as tabs

## Code Conventions

### File Structure Pattern
```
demo{N}-{name}/
├── cmd/{name}/main.go       # Standalone entry point (optional)
├── pkg/{core}/              # Core logic (simulation, models)
│   └── {feature}.go
├── pkg/gui/
│   ├── app.go               # Standalone app (XxxApp)
│   └── embedded.go          # Embeddable version (EmbeddedXxxApp)
└── shaders/                 # GPU shaders if needed
```

### Embedded App Interface
Every demo implements this pattern for the unified GUI:
```go
type EmbeddedXxxApp struct { ... }
func NewEmbeddedXxxApp() *EmbeddedXxxApp
func (app *EmbeddedXxxApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject
func (app *EmbeddedXxxApp) Start()  // Called when tab selected
func (app *EmbeddedXxxApp) Stop()   // Called when tab deselected
```

### GUI Framework: Fyne v2
- Use `fyne.Do(func() { ... })` for UI updates from goroutines
- Theme: FeCIM blue background (#003264), cyan accents
- Widgets: Prefer standard Fyne widgets, custom widgets extend `widget.BaseWidget`

### Key Constants
```go
const FeCIMLevels = 30  // Always use for quantization
```

### Quantization
Always quantize to 30 levels for FeCIM simulation:
```go
import "multilayer-ferroelectric-cim-visualizer/demo2-crossbar/pkg/crossbar"
quantized := crossbar.QuantizeTo30Levels(value)  // value in [0,1]
```

## Important Files Reference

| Task | File | Key Functions |
|------|------|---------------|
| Add new demo | `cmd/fecim-visualizer/main.go` | Add to DemoApp struct, create tab |
| Modify crossbar | `demo2-crossbar/pkg/crossbar/array.go` | MVM(), VMM(), ProgramWeight() |
| Modify hysteresis | `demo1-hysteresis/pkg/ferroelectric/preisach.go` | Update(), GetHysteresisLoop() |
| Change theme | `shared/theme/theme.go` | Color constants |
| Add logging | `shared/logging/logging.go` | NewLogger() |

See **scriptReference.md** for complete file structure and function index.

## Physics Background

### Ferroelectric Hysteresis
- **Pr (Remanent Polarization)**: ~25 µC/cm² - polarization at zero field
- **Ec (Coercive Field)**: ~1 MV/cm - field to switch polarization
- **Ps (Saturation Polarization)**: ~30 µC/cm² - maximum polarization

### Preisach Model
Hysteron-based model tracking turning points. Key methods:
- `Update(E)` - Apply field, get polarization with memory
- `GetHysteresisLoop()` - Full P-E curve for plotting

### Crossbar MVM
Matrix-vector multiply in O(1) by Kirchhoff's law:
```
I_out[i] = Σ G[i,j] × V_in[j]
```
Where G is conductance matrix (weights), V_in is input voltage.

## Testing

```bash
go test ./...                          # All tests
go test ./demo2-crossbar/pkg/crossbar  # Crossbar tests
go test -v -run TestPreisach           # Specific test
```

## Common Tasks

### Add a New Demo
1. Create `demo{N}-{name}/pkg/gui/app.go` with `XxxApp`
2. Create `demo{N}-{name}/pkg/gui/embedded.go` with `EmbeddedXxxApp`
3. Import in `cmd/fecim-visualizer/main.go`
4. Add to `DemoApp` struct and tabs
5. Add to `launcher.go` demo list

### Modify Crossbar Behavior
Edit `demo2-crossbar/pkg/crossbar/array.go`:
- `MVM()` for matrix-vector multiply
- `quantizeDAC()`/`quantizeADC()` for quantization
- `ProgramWeight()` for cell programming

### Add Non-Ideality
Edit `demo2-crossbar/pkg/crossbar/nonidealities.go`:
- Add analysis struct (e.g., `DriftAnalysis`)
- Add method (e.g., `Array.AnalyzeDrift()`)
- Update GUI in `demo2-crossbar/pkg/gui/app.go`

## Dependencies

- **Fyne v2.7.2** - Cross-platform GUI
- **Charm/BubbleTea** - Terminal UI (demo1 TUI mode)
- **go-gl/glfw** - Window management
- **vulkan-go/vulkan** - GPU rendering (demo1 Vulkan mode)

## Error Handling

- Crossbar array operations return errors for bounds checking
- GUI operations use `fyne.Do()` for thread safety
- Logging to `logs/` directory with timestamps

## Performance Notes

- Heatmaps use `SetData()` for batch updates
- Animation uses `time.Sleep()` in goroutines with `fyne.Do()` callbacks
- Large arrays (128x128) may need rate limiting on updates

## Dr. Tour's Key Specs

| Metric | Value | Source |
|--------|-------|--------|
| Discrete Levels | 30 | Tour presentation |
| Bits/Cell | 4.91 | log2(30) |
| Energy vs NAND | 10M× better | Tour claims |
| Energy vs DRAM | 1000× better | Tour claims |
| TRL | 4 | Tour presentation |

## Experimental Code

The `demo2-crossbar/pkg/_layers_experimental/` directory contains research code for advanced features (attention, transformers, federated learning, etc.). This code is not integrated into the main demos.
