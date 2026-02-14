<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module4-circuits/pkg/gui

## Purpose

Provides complex Fyne-based GUI for peripheral circuit simulation. Displays unified circuit view with DAC/ADC/TIA components, ISPP state machine for write programming, overlay modes (voltage, conduction, disturbance), sense-amplifier chains, array timing analysis, and circuit metrics. Deeply integrated with array simulation and supports multiple operation modes (read, write, compute). Large codebase with careful state management.

## Key Files

| File | Description |
|------|-------------|
| `device_state.go` | Unified device state (68KB). Manages ISPP state machine (level-based), circuit parameters, array simulation state. |
| `app.go` | Main application entry point (17KB). Window layout, tab structure. |
| `tab_unified.go` | Primary unified display tab (68KB). Array visualization with overlays. |
| `tab_unified_voltage.go` | Voltage waveform view and time-series analysis (29KB). |
| `tab_unified_drawing.go` | Canvas rendering for heatmaps and waveforms (26KB). |
| `keyboard.go` | Keyboard controls for device operation (8.9KB). |
| `embedded.go` | Implements embedded app interface. |
| Subdirectories | `unified/`, `display/`, `ispp/`, `overlay/`, `sense/` with specialized components |

## For AI Agents

### Working In This Directory

**Critical: ISPP State Machine in device_state.go**

Module4 has its own level-based ISPP state machine (different from Module1's WriteController):

- States: `OpStateIdle`, `OpStateISPPProgram`, `OpStateVerify`, `OpStateSuccess`, `OpStateFailed`
- Called by `runISPPWithAnimation()` in `tab_unified_voltage.go`
- Dispatches to either level-based engine or Landau-Khalatnikov WriteController (via `ispp_write.go` in shared/physics)
- Must track program pulses, sense operations, convergence

**Device State Structure (device_state.go):**

Massive state struct (68KB) tracking:
- Array dimensions, conductance values
- ISPP parameters: target level, pulse width, convergence tolerance
- DAC settings: voltage for each column (0-1V for read, up to 1.3V for write)
- Sense amplifier chain state (in `sensechain.go`)
- Temperature and device variation
- Overlay mode (voltage, conduction, disturbance)
- Operation mode (OpModeRead, OpModeWrite, OpModeCompute)
- Word-line selection (single row, all rows, or custom pattern)

**Multi-Tab Integration:**

- `tab_unified.go` - Primary unified view with array heatmap
- `tab_unified_voltage.go` - Voltage time-series with ISPP waveforms
- `tab_comparison.go` - Technology comparison metrics
- `tab_reference_specs.go` - Reference circuit specifications
- Tabs share `device_state.go`; changes in one tab affect others

**Subdirectory Structure:**

- `unified/` - High-level unified display components
- `display/` - Low-level canvas drawing and rendering utilities
- `ispp/` - ISPP-specific widgets and dialogs
- `overlay/` - Overlay mode selectors and visualizers
- `sense/` - Sense-amplifier chain visualization

### Testing Requirements

```bash
# Run all module4 GUI tests
go test ./module4-circuits/pkg/gui -v

# Run device_state tests (large test file, 47KB)
go test ./module4-circuits/pkg/gui -run TestDeviceState -v

# Run ISPP engine tests
go test ./module4-circuits/pkg/gui -run TestISPPEngine -v

# Run unified tab tests
go test ./module4-circuits/pkg/gui -run TestUnifiedTab -v

# Run voltage tab tests (includes ISPP animation)
go test ./module4-circuits/pkg/gui -run TestVoltageTab -v

# Run module1-module4 integration test
go test ./module4-circuits/pkg/gui -run TestModule1Module4Integration -v

# Run physics and coupling tests
go test ./module4-circuits/pkg/gui -run TestPhysics -v

# Run keyboard controls test
go test ./module4-circuits/pkg/gui -run TestKeyboard -v
```

### Common Patterns

- **Device state initialization**: NewDeviceState() with array dimensions and parameters
- **ISPP program loop**: Call `runISPPWithAnimation()` which handles state transitions
- **Array heatmap**: Render conductance values using color gradient
- **Overlay modes**: Switch visualization between voltage, conduction, disturbance
- **DAC voltage update**: Modify `deviceState.DACVoltages[col]` for each column
- **Sense chain**: Use `sensechain.go` to compute sense-amplifier output voltage
- **Keyboard control**: Check key press events and update device state accordingly
- **Tab switching**: All tabs access same `device_state` instance

## Dependencies

### Internal

- `module4-circuits/pkg/arraysim` - Array simulation for circuits module
- `shared/physics` - Landau-Khalatnikov WriteController, material models, units
- `shared/peripherals` - DAC, ADC, TIA component models
- `shared/widgets` - Custom Fyne widgets
- `shared/theme` - Application theme
- `shared/logging` - Operation logging

### External

- `fyne.io/fyne/v2` - GUI framework
- `image/color` (Go stdlib) - Heatmap colors
- `math` (Go stdlib) - Numerical calculations
- `sync` (Go stdlib) - Thread-safe state access

### Known Issues

- Pre-existing test failures: Some `TestUnifiedTabISPPEngine` and `TestUnifiedActionButtons` tests fail even on clean main (likely setup issues; not critical for feature development)
- Module4 is the most complex module; large test file (47KB) and monolithic state struct

<!-- MANUAL: Last edited 2026-02-13. Most complex module; device_state.go is 68KB and manages all circuit state. -->
