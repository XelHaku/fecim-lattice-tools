# Module 4: Peripheral Circuits

Peripheral circuit simulation for FeCIM (Ferroelectric Compute-in-Memory) crossbar arrays. This module models the signal chain required for integrated analog memory operations: Digital-to-Analog Conversion (DAC), Transimpedance Amplification (TIA), and Analog-to-Digital Conversion (ADC), with interactive Fyne-based visualization.

## Overview

Module 4 provides **simulation models** of peripheral circuits essential for ferroelectric CIM systems. Each cell in a crossbar array requires voltage control (via DAC) for writes and current sensing (via TIA/ADC) for reads. The values below reflect **default parameters** from `shared/peripherals/` and are not device measurements.

- **DAC**: Maps digital codes (0-31) to programmable voltages (-1.5V to +1.5V); demo uses 30 of 32 levels
- **TIA**: Converts crossbar column currents to measurable voltages (10 kOhm gain default)
- **ADC**: Quantizes sensed voltages back to digital levels (5-bit, 32 levels default)
- **Charge Pump**: Boosts 1V CMOS supply to +/-1.5V write voltages (Dickson topology model)

The visualization supports interactive simulation with a Mode-First UX that automatically configures word lines and DAC ranges based on operation (READ, WRITE, or COMPUTE).

## Key Features

### ADC: 5-Bit Successive Approximation Register (Default Model)

- Resolution: 5 bits (32 levels, demo uses 30)
- Input range: 0V to 1.0V (safe sensing below coercive voltage)
- INL: 0.5 LSB (Integral Nonlinearity)
- DNL: 0.25 LSB (Differential Nonlinearity)
- Conversion time: 50 ns (SAR architecture)
- ENOB: 4.87 bits (accounting for nonlinearity)
- Energy: ~25 fJ/conversion (model estimate)

**Modeling**: ADC includes quantization errors via INL/DNL and optional SAR noise modeling.

### DAC: 5-Bit Voltage Generation (Default Model)

- Resolution: 5 bits (32 levels, demo uses 30)
- Output range: -1.5V to +1.5V (write voltage window)
- Settling time: 10 ns
- INL: 0.5 LSB
- DNL: 0.25 LSB
- Energy: ~15 fJ/conversion (model estimate)
- Supports both positive and negative write pulses

**Modeling**: Maps discrete levels to the configured write window.

### TIA: 10 kOhm Transimpedance Amplifier (Default Model)

- Gain: 10 kΩ (10V/µA)
- Bandwidth: 100 MHz
- Input-referred noise: 1 pA/sqrt(Hz)
- Output offset: 5 mV
- Maximum input current: 100 µA
- Maximum output voltage: 1.0V (matches ADC input range)
- Dynamic range: computed from defaults via `TIA.DynamicRange()`

**Modeling**: Current-to-voltage conversion with noise and output clamping.

### Charge Pump: 1V to +/-1.5V Boost (Default Model)

- Topology: 2-stage Dickson pump
- Input: 1V (standard CMOS supply)
- Ideal output: 3.0V (1V × 3 stages)
- Actual output: clamped to target voltage after modeled drops
- Clock frequency: 50 MHz
- Efficiency: 70%
- Output ripple: computed by `ChargePump.OutputRipple()` (default output cap = 10x flying cap)

**Modeling**: Dickson pump charge redistribution enables write voltage generation without off-chip supplies.

## Quick Start

### Build from Source

```bash
cd <local-path>
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools
```

### Run CLI Demo

```bash
# Show all peripheral details
go run ./cmd/fecim-lattice-tools circuits cli -all

# Show all details with file logging
go run ./cmd/fecim-lattice-tools circuits cli -all -logger -verbosity 2

# Show specific circuit
go run ./cmd/fecim-lattice-tools circuits cli -dac
go run ./cmd/fecim-lattice-tools circuits cli -adc
go run ./cmd/fecim-lattice-tools circuits cli -tia
go run ./cmd/fecim-lattice-tools circuits cli -pump

# Show detailed analysis
go run ./cmd/fecim-lattice-tools circuits cli -linearity
go run ./cmd/fecim-lattice-tools circuits cli -timing
go run ./cmd/fecim-lattice-tools circuits cli -power

# Run ISPP write/verify demo (shared hysteresis physics)
go run ./cmd/fecim-lattice-tools circuits cli -ispp
```

### Run GUI Visualization

```bash
# Start unified application with Module 4 tab
go run ./cmd/fecim-lattice-tools/main.go

# Or run Module 4 GUI via subcommand
go run ./cmd/fecim-lattice-tools circuits
```

## Package Structure

### `shared/peripherals/` — Physics Models (Shared)

Core peripheral circuit implementations with realistic behavior:

- **adc.go** (123 lines)
  - `ADC` struct: 5-bit SAR configuration
  - `Convert()`: Ideal quantization
  - `ConvertWithNonlinearity()`: Realistic INL/DNL errors
  - `ENOB()`: Effective number of bits
  - `EffectiveSNR()`: Signal-to-noise accounting for errors
  - Methods: `Levels()`, `Resolution()`, `EnergyPerConversion()`, `TheoreticalSNR()`

- **dac.go** (90 lines)
  - `DAC` struct: 5-bit converter
  - `Convert()`: Level to voltage mapping
  - `ConvertWithNonlinearity()`: With INL/DNL errors
  - `VoltageRange()`: Min/max output
  - Energy and resolution calculations

- **tia.go** (101 lines)
  - `TIA` struct: Transimpedance amplifier
  - `Convert()`: Current to voltage with offset/clamping
  - `ConvertWithNoise()`: Thermal noise injection
  - `SNR()`: Signal-to-noise ratio for input current
  - Performance: `MinDetectableCurrent()`, `DynamicRange()`, `SettlingTime()`, `PowerConsumption()`

- **chargepump.go** (127 lines)
  - `ChargePump` struct: Dickson topology
  - `IdealOutputVoltage()`: Theoretical (N+1)×Vin
  - `ActualOutputVoltage()`: With threshold drops and IR losses
  - Efficiency analysis: `BoostFactor()`, `PowerInput()`, `PowerOutput()`, `PowerLoss()`
  - Dynamic behavior: `OutputRipple()`, `RiseTime()`, `MaxCurrentCapability()`

- **analysis.go** — Advanced Analysis
  - `AnalyzeINLDNL()`: Linearity metrics for DAC/ADC
  - `AnalyzeTiming()`: Critical path analysis
  - `AnalyzePower()`: Energy breakdown across components

### `pkg/gui/` — Fyne Visualization

Interactive simulation interface built with Fyne v2:

- **embedded.go** (99 lines)
  - `EmbeddedCircuitsApp`: Pluggable app for unified visualizer
  - `BuildContent()`: Create tab content
  - `Start()`: Resume simulation
  - `Stop()`: Pause simulation

- **device_state.go** (616 lines)
  - `DeviceState`: Unified simulation state
  - `OpMode`: READ/WRITE/COMPUTE operation modes
  - `VoltageRange`: Material-derived safe voltage regions
  - `CalibrationParams`: Loaded from physics.yaml
  - **Passive Mode (0T1R)**: Forces all WLs on; ignores manual WL selection
  - **1T1R Mode**: Individual row control via word lines
  - `Compute()`: MVM computation with TIA/ADC signal chain

- **tab_unified.go** (1000+ lines)
  - `createUnifiedView()`: Mode-First UX layout
  - `createModeBar()`: READ/WRITE/COMPUTE buttons
  - `createWriteModePanel()`: Target level selection
  - `createComputeModePanel()`: Input vector controls
  - `createDACInputSection()`: Voltage presets with material calibration
  - `createMainSimSection()`: Array visualization
  - Dynamic voltage ranges from physics.yaml

- **helpers.go**, **drawing.go**, **font.go** — UI Utilities
  - Drawing primitives for arrays, waveforms
  - Color scheme management
  - Layout helpers

## Mode-First UX

The GUI presents three operation modes, each with automatically configured word lines and DAC ranges:

### READ Mode

- **Word Lines**: Single row active (user selectable)
- **DAC Voltage**: Safe read range (0 to ~0.5V, derived from material)
- **Purpose**: Non-destructive sense of cell state
- **Signal Chain**: DAC → Array → TIA → ADC
- **Read voltage** must stay below coercive voltage Vc to avoid polarization switching

### WRITE Mode

- **Word Lines**: Single row active (user selectable)
- **DAC Voltage**: Write range (Vc to ~2.5V, derived from material)
- **Purpose**: Program cell to target state via ferroelectric polarization
- **Signal Chain**: DAC → ChargePump → Array
- **Write voltage** must exceed Vc to switch polarization
- **Target Level** slider maps 0-29 to write voltage range

### COMPUTE Mode

- **Word Lines**: All rows active
- **DAC Voltage**: Safe compute range (0 to ~0.5V)
- **Purpose**: Matrix-vector multiplication across all rows
- **Signal Chain**: DAC → Array → TIA → ADC (per row)
- Each column gets independent input voltage from input vector
- Produces one output per active row

## Architecture Support

Module 4 supports multiple crossbar array architectures:

### Passive (0T1R)

- **Topology**: Direct FeFET gate-to-bit-line, no access transistor
- **Density**: 4F² (minimum, highest density)
- **Word Lines**: Always driven (no control)
- **Sneak Paths**: 5-20% error; mitigated via design
- **Best For**: Edge inference, mobile compute, ultra-high density
- **Mode Behavior**: All WLs always on; cannot change manually

### 1T1R

- **Topology**: 1 Transistor + 1 FeFET per cell
- **Density**: 8-12F² (lower than 0T1R)
- **Word Lines**: Gated by pass transistor; precise control
- **Sneak Paths**: ~0% (transistor isolates)
- **Best For**: Training, precise programming, neural networks
- **Mode Behavior**: WL selection controls active rows

### 2T1R (Planned)

- **Topology**: 2 Transistors + 1 FeFET
- **Features**: Read and write gates separate for decoupling
- **Support**: Will be added in future

## Signal Chain

Complete signal path from digital command to digital readout:

```
WRITE OPERATION:
  Digital Level (0-29)
       ↓
    DAC (5-bit)
       ↓
  Charge Pump (1V → ±1.5V)
       ↓
  Ferroelectric Cell
       ↓
  Polarization State Change

READ OPERATION:
  Digital Command (row/col select)
       ↓
    DAC (sense voltage ~0.1-0.5V)
       ↓
  Ferroelectric Cell
       ↓
  Column Current (picoamps to nanoamps)
       ↓
    TIA (10 kΩ gain)
       ↓
  Voltage (0-1V range)
       ↓
    ADC (5-bit quantization)
       ↓
  Digital Level (0-29)

COMPUTE OPERATION (MVM):
  Input Vector [0-255 × N cols]
       ↓
    DAC × N (maps to read range)
       ↓
  Crossbar Array [M rows × N cols]
       ↓
  Row Currents = Sum(DAC[c] × G[r][c])
       ↓
    TIA × M (current-to-voltage)
       ↓
  Row Voltages
       ↓
    ADC × M (quantize)
       ↓
  Output Vector [0-29 × M rows]
```

## Material Calibration

Voltage ranges are automatically derived from material physics via `physics.yaml`:

### Configuration (physics.yaml)

```yaml
calibration:
  field_min_ratio: 0.5      # Read max = 0.5 × Vc (safe sensing)
  field_max_ratio: 2.5      # Write max = 2.5 × Vc (program window)
```

### Voltage Calculation

For a material with coercive voltage Vc = 2.0V:

| Parameter | Formula | Value |
|-----------|---------|-------|
| Read Min | 0 | 0V |
| Read Max | field_min_ratio × Vc | 1.0V |
| Write Min | Vc | 2.0V |
| Write Max | field_max_ratio × Vc | 5.0V |
| Practical Max | Hardware limit | 3.0V (clamped) |

The DAC automatically switches between these ranges based on operation mode:
- READ/COMPUTE: Uses 0V to ~0.5-1.0V (safe sensing)
- WRITE: Uses Vc to ~1.5-2.5V (sufficient for switching)

## Testing

All peripheral circuits include comprehensive tests:

```bash
# Run all Module 4 tests
go test ./module4-circuits/...

# Run specific test suite
go test ./shared/peripherals -v

# Run with coverage
go test ./shared/peripherals -cover
```

### Test Coverage

- **DAC Tests**: Monotonicity, level mapping, round-trip with ADC
- **ADC Tests**: Quantization, clamping, ENOB calculation
- **TIA Tests**: Linearity, output clamping, noise behavior
- **ChargePump Tests**: Voltage boost, efficiency, power balance
- **Integration Tests**: Complete signal chain round-trips

Run `go test ./shared/peripherals/peripherals_test.go` to see all 10+ test cases.

## Related Documentation

For deeper understanding of ferroelectric physics, architecture trade-offs, and design decisions:

- **[Crossbar Operations Guide](../docs/peripheral-circuits/circuits.operations.md)** — 0T1R vs 1T1R architecture details, sneak path analysis
- **[CIM Fundamentals](../docs/peripheral-circuits/circuits.CIM-fundamentals.md)** — Core concepts of compute-in-memory
- **[Research Papers](../docs/peripheral-circuits/circuits.research.md)** — References for physics models and circuit design
- **[ELI5 Explanation](../docs/peripheral-circuits/circuits.ELI5.md)** — Beginner-friendly introduction to peripheral circuits
- **[Physics Configuration](../config/physics.yaml)** — Material properties and calibration parameters used for voltage range calculation

## Development Notes

### Key Conventions

- All voltages in volts (V), currents in amperes (A)
- Use `fyne.Do()` for all UI updates from simulation threads
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Test before committing: `go test ./...`

### Design Decisions

1. **5-Bit DAC/ADC**: Minimal resolution for 30-level FeCIM (uses 30 of 32 codes)
2. **Dickson Charge Pump**: Simple, efficient topology suitable for CMOS integration
3. **10 kΩ TIA Gain**: Balances speed (100 MHz) and noise (1 pA/√Hz)
4. **Material-Derived Ranges**: Voltage limits calculated from Ec and thickness, not hardcoded
5. **Passive Mode Lock**: 0T1R architecture forces all WLs on; prevents misconfiguration

### Adding New Peripherals

To add a new circuit model (e.g., multiplexer):

1. Create `shared/peripherals/component.go` with struct and methods
2. Add tests in `shared/peripherals/component_test.go`
3. Create GUI in `pkg/gui/tab_component.go` with Fyne visualization
4. Integrate into `device_state.go` if part of signal chain
5. Add to README in this section

### Debugging Simulation

Enable detailed logging in `pkg/gui/helpers.go`:

```go
// In device_state.Compute():
if debugMode {
    log.Printf("Row %d: I=%.2e A, V=%.3f V, Level=%d", r, rowCurrent, rowVoltage, rowLevel)
}
```

## License

Part of fecim-lattice-tools. See repository LICENSE for details.

## Questions?

- Check `docs/development/GUI/FYNE_NOTES.md` for Fyne-specific issues
- See `docs/development/scriptReference.md` for function lookups
- Review `docs/comparison/HONESTY_AUDIT.md` for physics verification
