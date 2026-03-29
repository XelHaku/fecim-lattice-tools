<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module2-crossbar/pkg/gui

## Purpose

Provides Fyne-based GUI for crossbar module. Visualizes array conductance values, controls MVM operations, selects/deselects rows and columns, displays result outputs, manages operation history, and enables non-ideality control (IR drop, sneak paths, drift, temperature). Embedded as a tabbed module within the main application.

## Key Files

| File | Description |
|------|-------------|
| `embedded.go` | Implements embedded app interface: `BuildContent()`, `Start()`, `Stop()`. |
| `app.go` | Main application window layout and tab structure (17KB). |
| `keyboard.go` | Keyboard controls for row/column selection and operation triggers (3.7KB). |
| GUI test files | Module integration tests, embedded interface validation. |

## For AI Agents

### Working In This Directory

**Embedded App Interface:**

Must implement (see `embedded.go`):

```go
type EmbeddedApp interface {
    BuildContent() fyne.CanvasObject
    Start()
    Stop()
}
```

**Array Visualization:**

- Conductance displayed as heatmap (darker = lower conductance)
- Selected rows/columns highlighted
- Result vector displayed as bar chart or histogram

**Operation Controls:**

- Row selector (single or multi-select)
- Column selector (single or multi-select)
- Input vector entry (DAC values or fixed preset)
- MVM execute button
- Reset array button
- Operation history log

**Non-ideality Controls:**

- Toggle IR drop on/off
- Toggle sneak paths on/off
- Control temperature setting
- Drift simulation enabled/disabled
- Device variation (noise level) slider

**Thread Safety:**

- Array access protected by crossbar package; GUI just reads results
- MVM operations must run in goroutine; UI updates via `fyne.Do()`
- Operation history thread-safe list

### Testing Requirements

```bash
# Run all crossbar GUI tests
go test ./module2-crossbar/pkg/gui -v

# Run embedded interface test
go test ./module2-crossbar/pkg/gui -run TestEmbedded -v

# Run keyboard control tests
go test ./module2-crossbar/pkg/gui -run TestKeyboard -v

# Run integration with array simulation
go test ./module2-crossbar/pkg/gui -run TestArrayIntegration -v
```

### Common Patterns

- **Row/column selection**: Bitmask or index list
- **Input vector**: DAC voltages or conductance targets
- **MVM execution**: Queue operation in background goroutine, update UI when done
- **Non-ideality toggle**: Modify array config struct before next MVM
- **Heatmap rendering**: Use color gradient from min to max conductance

## Dependencies

### Internal

- `module2-crossbar/pkg/crossbar` - Array simulation
- `shared/physics` - Conductance models, units
- `shared/widgets` - Custom Fyne widgets (tooltips, color legends, etc.)
- `shared/theme` - Application theme

### External

- `fyne.io/fyne/v2` - GUI framework
- `image/color` (Go stdlib) - Heatmap colors

<!-- MANUAL: Last edited 2026-02-13. Embedded module for crossbar visualization. -->
