# Module 4: Peripheral Circuits - Architecture Documentation

## Overview

Module 4 implements a complete peripheral circuit simulation environment for ferroelectric compute-in-memory (FeCIM) systems. It provides physics-accurate models of analog and mixed-signal components (DAC, ADC, TIA, charge pump) integrated with an interactive GUI for visualizing signal flow through a complete CIM read/write/compute pipeline.

**Key Concept**: The module demonstrates how digital input signals are converted to analog control voltages, processed through a ferroelectric crossbar array, and converted back to digital output levels through a complete signal chain.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

## Array Coupling, Fidelity Tiers, and Definitions

If you are looking specifically for **array simulation** concepts (how DAC+architecture selection couples into measurable currents), start here:

- `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`

That page defines **Vcell**, **sneak paths**, **half-select disturb**, and **IR drop**, and proposes a set of explicit **fidelity tiers**:

- **Ideal** (fast, didactic)
- **Approx** (lumped non-idealities)
- **DC nodal / MNA** (sneak + IR drop via circuit solve)
- **Transient** (time-domain settling, bandwidth; planned)

> **HONESTY_AUDIT:** The fidelity-tier selector is a documentation-level plan; implementation status may lag.

### Sense-chain measurable current range (TIA + ADC)

A practical way to reason about “what currents can we measure?” is to combine the TIA transfer and ADC input range:

- TIA (idealized): \(V_{out} = I_{in}\cdot R_{TIA} + V_{offset}\)
- ADC range: \(V_{ref,low} \le V_{in} \le V_{ref,high}\)

This implies an approximate measurable current span:

\[
I_{min} \approx \frac{V_{ref,low}-V_{offset}}{R_{TIA}},\quad
I_{max} \approx \frac{V_{ref,high}-V_{offset}}{R_{TIA}}
\]

…and then apply any explicit TIA input clamp \(I_{max,TIA}\):

\[
I_{max,meas} = \min(I_{max}, I_{max,TIA})
\]

All quantities are in SI units: A, Ω, V.

> **HONESTY_AUDIT:** Exact clamping/rounding depends on the code in `shared/peripherals/*`.

## Project Structure

```
shared/
└── peripherals/              # Physics models (DAC/ADC/TIA/ChargePump + analysis)
    ├── adc.go
    ├── dac.go
    ├── tia.go
    ├── chargepump.go
    ├── analysis.go
    └── peripherals_test.go

module4-circuits/
├── pkg/
│   └── gui/                  # Fyne-based visualization
│       ├── app.go            # Main CircuitsApp struct
│       ├── device_state.go   # DeviceState unified simulation
│       ├── embedded.go       # Embedded interface implementation
│       ├── tab_unified.go    # Mode-First UX
│       ├── tab_comparison.go # FeFET vs CPU/GPU benchmarks
│       ├── tab_reference.go  # Timing diagrams and specs
│       ├── tab_reference_timing.go
│       ├── tab_reference_specs.go
│       ├── drawing.go        # Canvas rendering helpers
│       ├── font.go           # Text rendering utilities
│       └── helpers.go        # UI component builders
└── cmd/
    ├── circuits-gui/         # GUI application entry point
    └── circuits/             # CLI application entry point
```

## Peripherals Package Architecture

**Location:** `shared/peripherals/` (used by Module 4 GUI + CLI)

### Core Components

#### 1. DAC (Digital-to-Analog Converter)

**File**: `shared/peripherals/dac.go`

Converts discrete digital levels to analog voltages for driving the crossbar array.

```go
type DAC struct {
    Bits       int     // Resolution (5 bits = 32 levels, use 30 for demo baseline)
    VrefHigh   float64 // High reference (+1.5V)
    VrefLow    float64 // Low reference (-1.5V)
    INL        float64 // Integral nonlinearity error
    DNL        float64 // Differential nonlinearity error
    SettleTime float64 // Settling time (ns)
}
```

**Key Methods**:
- `Convert(level int) float64` - Maps digital level (0-31) to voltage (-1.5V to +1.5V)
- `ConvertWithNonlinearity()` - Adds realistic INL/DNL errors
- `Resolution()` - Voltage per LSB
- `EnergyPerConversion()` - Power consumption estimate
- `VoltageRange()` - Returns min/max output voltage

**Default Configuration**:
- 5-bit resolution (32 levels, 30 used)
- Range: -1.5V to +1.5V
- INL: 0.5 LSB, DNL: 0.25 LSB
- Settling time: 10 ns

#### 2. ADC (Analog-to-Digital Converter)

**File**: `shared/peripherals/adc.go`

Converts sensed voltages back to discrete digital levels for data extraction.

```go
type ADC struct {
    Bits           int     // Resolution (5 bits = 32 levels)
    VrefHigh       float64 // High reference (1.0V)
    VrefLow        float64 // Low reference (0.0V)
    INL            float64 // Integral nonlinearity
    DNL            float64 // Differential nonlinearity
    ConversionTime float64 // Conversion time (ns)
    Type           ADCType // SAR, Flash, or Sigma-Delta
}
```

**Key Methods**:
- `Convert(voltage float64) int` - Quantizes voltage to 32 discrete levels
- `ConvertWithNonlinearity()` - Adds INL/DNL errors
- `Resolution()` - Voltage per LSB
- `ENOB()` - Effective Number of Bits (accounting for nonlinearity)
- `TheoreticalSNR()` - Ideal signal-to-noise ratio
- `EffectiveSNR()` - Realistic SNR with nonlinearity
- `EnergyPerConversion()` - Power consumption

**Default Configuration**:
- 5-bit resolution (32 levels)
- Range: 0.0V to 1.0V (safe read zone)
- INL: 0.5 LSB, DNL: 0.25 LSB
- Conversion time: 50 ns (SAR architecture)

#### 3. TIA (Transimpedance Amplifier)

**File**: `shared/peripherals/tia.go`

Converts read column currents to measurable voltages using transimpedance gain.

```go
type TIA struct {
    Gain             float64 // Transimpedance gain (10 kΩ)
    Bandwidth        float64 // -3dB bandwidth (100 MHz)
    InputNoiseRMS    float64 // Input noise (1 pA/sqrt(Hz))
    OutputOffset     float64 // Offset voltage (5 mV)
    MaxInputCurrent  float64 // Maximum input (100 µA)
    MaxOutputVoltage float64 // Maximum output (1.0V)
}
```

**Key Methods**:
- `Convert(current float64) float64` - Vout = Iin * Gain + Offset
- `ConvertWithNoise()` - Adds thermal noise contribution
- `SNR(current float64)` - Signal-to-noise ratio at given current
- `MinDetectableCurrent()` - Noise floor in amperes
- `DynamicRange()` - Range in dB
- `SettlingTime()` - Step response settling to 0.1% accuracy
- `PowerConsumption()` - Estimated power draw

**Default Configuration**:
- Gain: 10 kΩ (converts 1 µA to 10 mV)
- Bandwidth: 100 MHz
- Input noise: 1 pA/sqrt(Hz)
- Max input: 100 µA (saturates above this)
- Offset: 5 mV

**Signal Chain**: Row currents (sum of column currents from active weights) → TIA converts to voltage → ADC quantizes

#### 4. Charge Pump

**File**: `shared/peripherals/chargepump.go`

Generates write voltages (±1.5V) from lower supply voltage (1V) using capacitive charge transfer.

```go
type ChargePump struct {
    InputVoltage   float64 // Supply voltage (1V)
    OutputVoltage  float64 // Target write voltage (±1.5V)
    Stages         int     // Number of pump stages (2 = Dickson)
    DiodeDrop      float64 // Effective diode/switch drop per stage (0.3V)
    ClockFrequency float64 // Pump clock frequency (50 MHz)
    LoadCurrent    float64 // Maximum load current (10 µA)
    FlyCapacitance float64 // Flying capacitor value (100 pF)
    Efficiency     float64 // Conversion efficiency (70%)
}
```

**Key Methods**:
- `IdealOutputVoltage()` - Theoretical max: (N+1) * Vin
- `ActualOutputVoltage()` - Output after diode/IR drops (with regulation clamp)
- `OutputRipple()` - Peak-to-peak ripple voltage
- `BoostFactor()` - Voltage multiplication ratio
- `PowerInput()` / `PowerOutput()` - Power consumption
- `RiseTime()` - Voltage rise time (10-90%)
- `ChargeTransferEfficiency()` - Per-stage efficiency
- `SupportsLevel()` - Can this pump generate voltage for a given level?

**Default Configuration** (Dickson pump):
- 2 stages, input 1V → output ±1.5V (boost factor ~1.5)
- Clock: 50 MHz
- Flying caps: 100 pF per stage
- Efficiency: 70%
- Output ripple: ~0.2 mV at 10 µA load (with 1 nF output cap)

#### 5. Analysis Functions

**File**: `shared/peripherals/analysis.go`

System-level analysis combining multiple peripherals.

**Key Analysis Types**:

**INLDNLAnalysis** - Linearity characterization
```go
type INLDNLAnalysis struct {
    Levels     int       // Number of codes
    INLValues  []float64 // INL error at each code (LSB)
    DNLValues  []float64 // DNL error at each code (LSB)
    MaxINL     float64   // Peak INL error
    MaxDNL     float64   // Peak DNL error
    WorstCode  int       // Code with maximum INL
}
```
- `DAC.AnalyzeINLDNL()` - Characterize DAC linearity
- `ADC.AnalyzeINLDNL()` - Characterize ADC linearity

**TimingAnalysis** - Complete signal chain timing
```go
type TimingAnalysis struct {
    DACSettle      float64 // DAC settling
    ArraySettle    float64 // Array RC/sneak settling
    PumpRise       float64 // Charge pump rise time
    WritePulse     float64 // Write pulse width
    WriteTime      float64 // Total write operation time
    TIASettle      float64 // TIA step response
    ADCConvert     float64 // ADC conversion
    ReadTime       float64 // Total read operation time
    CycleTime      float64 // Full read+write cycle
    MaxThroughput  float64 // Operations per second
}
```
- `AnalyzeTiming(dac, adc, tia, pump)` - Complete system timing

**PowerBreakdown** - Energy and power consumption
```go
type PowerBreakdown struct {
    DACPower, ADCPower, TIAPower, PumpPower float64 // Power (W)
    DACEnergy, ADCEnergy, TIAEnergy, PumpEnergy float64 // Energy/op
    TotalPower, TotalEnergy float64
    DACFraction, ADCFraction, TIAFraction, PumpFraction float64
}
```
- `AnalyzePower(dac, adc, tia, pump, timing)` - Power breakdown

**TransferFunction** - Complete signal chain simulation
```go
type TransferFunction struct {
    InputLevels  []int     // Digital input (0-29)
    DACVoltages  []float64 // After DAC
    PumpVoltages []float64 // After charge pump
    TIAVoltages  []float64 // After TIA
    ADCLevels    []int     // Final digital output
    Errors       []int     // Output - Input deviation
}
```
- `ComputeTransferFunction(dac, adc, tia, pump)` - Trace signal through entire chain

## GUI Package Architecture

### Main Application Structure

#### CircuitsApp (463 lines)

**File**: `app.go`

Master application state holding all peripheral components and UI widgets.

**Key Components**:
```go
type CircuitsApp struct {
    // Fyne framework
    fyneApp fyne.App
    window  fyne.Window

    // Peripheral components
    dac  *peripherals.DAC
    adc  *peripherals.ADC
    tia  *peripherals.TIA
    pump *peripherals.ChargePump

    // Configuration state
    arrayRows, arrayCols int  // Array dimensions (8-128)
    quantLevels          int  // 30 for demo baseline
    dacBits, adcBits     int  // Resolution in bits
    vMin, vMax           float64 // Write voltage range
    readVoltage          float64 // Safe read voltage
    tiaGain              float64 // TIA gain (kΩ)
    selectedRow, selectedCol int
    targetLevel          int
    arrayWeights         [][]int // Programmed cell states
    inputVector          []int   // Compute mode input
    outputVector         []float64
    architecture         string  // "0T1R", "1T1R", "2T1R"

    // Device state (unified simulation)
    deviceState          *DeviceState

    // UI widgets for all tabs
    // ... hundreds of widgets for different views
}
```

**Initialization**:
```go
func NewCircuitsApp() *CircuitsApp {
    ca := &CircuitsApp{
        arrayRows:    8,           // Default 8x8 array
        arrayCols:    8,
        quantLevels:  30,          // Demo baseline (simulation baseline)
        dacBits:      5,
        adcBits:      5,
        vMin:         1.2,         // Write range
        vMax:         1.5,
        readVoltage:  0.5,
        tiaGain:      10.0,
        architecture: "0T1R",      // Default: passive mode
    }
    ca.dac = peripherals.DefaultDAC()
    ca.adc = peripherals.DefaultADC()
    ca.tia = peripherals.DefaultTIA()
    ca.pump = peripherals.DefaultChargePump()
    ca.initializeArray()
    return ca
}
```

#### DeviceState (615 lines)

**File**: `device_state.go`

Unified simulation state for read/write/compute operations.

**Key Enums**:

```go
// Operation modes
type OpMode int
const (
    OpModeRead    OpMode = iota // Single row, safe voltage (0-0.5V)
    OpModeWrite                 // Single row, write voltage (Vc to 1.3*Vc)
    OpModeCompute               // All rows, input vector (0-1V)
)

// Word line selection
type WLMode int
const (
    WLSingle WLMode = iota // One row (program/read operations)
    WLAll                  // All rows (matrix-vector multiply)
    WLCustom               // User-defined pattern
)

// DAC preset modes
type DACMode int
const (
    DACManual       DACMode = iota
    DACReadPreset            // All columns at readVoltage
    DACWritePreset           // Selected column at write voltage
    DACInputVector           // From digital input (0-255 -> 0-1V)
    DACRandom
)
```

**Core State**:
```go
type DeviceState struct {
    rows, cols int
    isPassive bool // 0T1R mode: all WLs always ON

    opMode OpMode  // Current operation
    wlMode WLMode  // Word line pattern
    activeRows []bool

    dacVoltages  []float64 // Per-column voltages
    dacMode      DACMode
    dacRangeMode DACRangeMode // Read vs Write range

    readRange   VoltageRange  // 0 to min(FieldMinRatio*Vc, 1.0V), floor 0.1V
    writeRange  VoltageRange  // Vc to FieldMaxRatio*Vc
    calibParams CalibrationParams // From physics.yaml

    rowCurrents []float64 // Computed output currents (µA)
    rowVoltages []float64 // TIA output voltages (V)
    rowLevels   []int     // ADC output levels
    saturated   []bool    // Saturation flags

    material *ferroelectric.HZOMaterial
    tia      *peripherals.TIA
    adc      *peripherals.ADC
}
```

**Voltage Range Calculation**:
- **Read Range**: 0 to `min(FieldMinRatio * Vc, 1.0V)` (floor 0.1V, safe zone)
- **Write Range**: `Vc` to `FieldMaxRatio * Vc` (exceeds coercive voltage for switching)
- Values derived from `physics.yaml` calibration parameters
- Fallback defaults: `FieldMinRatio=0.7`, `FieldMaxRatio=2.5`

**Key Methods**:
```go
// Mode control
SetPassiveMode(passive bool)      // Force 0T1R (all WLs on)
SetWLSingle(row int)              // Single row active
SetWLAll()                         // All rows for MVM
SetWLCustom(pattern []bool)        // Custom pattern

// Voltage control
SetDACVoltage(col int, voltage)    // Manual voltage
SetDACPreset(preset DACMode)       // Preset patterns
SetDACVoltageForState(col, level)  // Map level to write voltage

// Computation
Compute(weights [][]int, quantLevels int) // Run simulation

// Accessors
GetRowCurrent(row), GetRowVoltage(row), GetRowLevel(row) int
IsRowActive(row), IsSaturated(row) bool
GetDACVoltage(col) float64
ClassifyOperation() string // "READ", "WRITE", "COMPUTE"
```

### Mode-First UX (tab_unified.go - 1841 lines)

The unified view implements a "mode-first" design where operation mode is the primary control, with mode-specific panels hidden/shown dynamically.

#### View Structure

```
SIGNAL CHAIN: DAC -> Array -> TIA -> ADC
Architecture Toggle (Passive/1T1R/2T1R) | Material Selector
─────────────────────────────────────────────────────────
[MODE BAR] READ | WRITE | COMPUTE
─────────────────────────────────────────────────────────
[MODE-SPECIFIC PANEL - Hidden/Shown Based on Mode]
  WRITE MODE: Level slider + voltage display
  COMPUTE MODE: Input vector entry boxes
─────────────────────────────────────────────────────────
[DAC INPUT SECTION]
  Preset buttons (Read/Write) | Range label | Manual input
─────────────────────────────────────────────────────────
[MAIN SIMULATION AREA - Center]
  Array visualization (colors by level)
  Cell info display
  Input/output data paths
─────────────────────────────────────────────────────────
[ACTION BUTTONS] Program | Read | Verify | Compute | Reset
```

#### Key Sections

**1. Signal Chain Header**
- Shows architecture toggle (Passive/1T1R/2T1R)
- Material selector (changes voltage ranges)
- Signal path diagram

**2. Mode Bar**
- Three radio-button style selectors: READ, WRITE, COMPUTE
- Determines which preset/panel is shown
- Controls WL selection behavior

**3. Mode-Specific Panels**

**WRITE MODE**:
- Level slider (0-30): maps to write voltage
- Voltage display: shows DAC output in volts
- Target cell display: "Target: Row X, Col Y"
- Pulse visualization canvas

**COMPUTE MODE**:
- Input vector entry boxes (one per column)
- Voltage display per column
- Output level display per row
- Matrix multiplication indicator

**4. DAC Input Section**
- Preset buttons:
  - **Read Preset**: All columns at safe read voltage (material-derived)
  - **Write Preset**: Selected column at mid-write voltage
  - Actual voltage range shown: e.g., "0.0-0.5V" (read) or "0.6-1.5V" (write)
- Range mode label: "Read Mode" or "Write Mode"

**5. Main Simulation Area**
- **Array Canvas**: Visual representation
  - Cell colors: gradient from dark (level 0) to bright (level 30)
  - Cell borders highlight selected cell
  - Click-to-select for row/column
- **Cell Info Label**: Shows selected cell level, status
- **Data Path Labels**: Digital → DAC → Array → TIA → ADC with intermediate values
- **Status Label**: Operation feedback

**6. Action Section**
- **Program**: Set selected cell to target level (WRITE mode)
- **Program Random**: Fill array with random levels
- **Read**: Read selected cell and display output
- **Verify**: Write then immediately read back
- **Compute**: MVM with input vector
- **Animate**: Show signal flow step-by-step
- **Reset**: Clear array to mid-level

#### Passive Mode Enforcement (0T1R)

In passive (0T1R) architecture:
- All word lines are always ON (no transistor gating)
- All columns selected simultaneously
- Current sums across all rows: `I_out[r] = sum_c(G[r,c] * V[c])`
- WL checkboxes are disabled (cannot be toggled)
- Help text: "Passive mode: all rows always active"

#### Dynamic Voltage Ranges

Voltage ranges are calculated from ferroelectric material properties:

```go
safeReadMax := FieldMinRatio * Vc  // e.g., 0.7 * 1.2V = 0.84V
if safeReadMax > 1.0 {
    safeReadMax = 1.0
}
if safeReadMax < 0.1 {
    safeReadMax = 0.1
}
readRange.Max = safeReadMax
writeRange.Min = Vc                 // Coercive voltage (1.2V)
writeRange.Max = FieldMaxRatio * Vc // e.g., 2.5 * 1.2V = 3.0V
```

Material Selector updates these ranges in real-time.

### Supporting GUI Files

#### embedded.go

Implements the embedded application interface for use in unified visualizer.

```go
type EmbeddedCircuitsApp struct {
    *CircuitsApp
}

// BuildContent creates UI for embedding in a tab
func (e *EmbeddedCircuitsApp) BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject

// Start/Stop for lifecycle management
func (e *EmbeddedCircuitsApp) Start()
func (e *EmbeddedCircuitsApp) Stop()
```

#### tab_comparison.go

Compares FeCIM against CPU/GPU benchmarks with canvas visualizations.

#### tab_reference.go (+ timing/specs)

Timing diagrams and detailed specifications:
- **Timing Tab**: Waveforms showing read/write/compute pulse timing
- **Specs Tab**: System parameters table (array size, resolution, power, etc.)

#### drawing.go & font.go

Low-level graphics utilities:
- `drawRect()` - Fill rectangles with color
- `drawRectBorder()` - Rectangle outlines
- `drawRoundedRect()` - Rounded rectangles
- Text measurement and rendering

## State Management & Data Flow

### Read Operation Flow

```
User Interaction
    ↓
Select cell (row, col)
    ↓
Set ReadPreset (DAC to safe voltage, e.g., 0.3V)
    ↓
Set WLSingle(row) to enable target row only
    ↓
Compute():
    DAC outputs voltage V[col]
    ↓ (for active rows only)
    Array: conductance G[row,col] from weight level
    ↓
    Current I = G * V (sum across columns)
    ↓
    TIA: Vout = I * Gain (e.g., 100µA * 10kΩ = 1V)
    ↓
    ADC: Quantize voltage to level (0-31)
    ↓
Display results (level, current, voltage)
```

### Write Operation Flow

```
User Interaction
    ↓
Select target cell and level
    ↓
Set WritePreset (DAC to write voltage in write range)
    ↓
Set WLSingle(row) for selected row
    ↓
Apply voltage pulse (simulated, not physically written)
    ↓
Update cell state: weights[row][col] = targetLevel
    ↓
Refresh array visualization (color update)
```

### Compute (MVM) Flow

```
Set OpModeCompute
    ↓
Set WLAll() (all rows active)
    ↓
DAC SetInputVector(x[0..cols-1])
    ↓
For each row:
    Sum: I[row] = sum_c(G[row,c] * x[c])
    ↓
    Apply TIA: V[row] = I[row] * Gain
    ↓
    Apply ADC: y[row] = ADC(V[row])
    ↓
Output vector y[0..rows-1] is complete MVM result
```

## Architecture Design Decisions

### 1. Unified Device State

**Why**: DeviceState centralizes all simulation logic (WL control, voltage ranges, computation) separate from GUI concerns. Enables:
- Easy testing without GUI
- Consistent behavior across all views
- Material-aware voltage range calculation

### 2. Mode-First UX

**Why**: Operation mode (READ/WRITE/COMPUTE) should be the primary conceptual model:
- Users think in operations, not individual parameters
- Mode determines allowed configurations (WL patterns, voltage ranges)
- UI updates automatically: wrong configurations impossible

### 3. Physics-Based Calibration

**Why**: Voltage ranges derived from material properties (Ec, thickness) via physics.yaml:
- Safe read voltage = min(FieldMinRatio * Vc, 1.0V) (floor 0.1V, non-destructive)
- Write voltage = FieldMaxRatio * Vc (exceeds Vc for switching)
- Automatic when material selected, not hardcoded

### 4. Passive Mode Enforcement

**Why**: 0T1R architecture (all WLs always on) fundamentally changes device behavior:
- UI enforces this constraint (WL checkboxes disabled)
- Compute becomes true parallel: all rows active simultaneously
- Simplifies teaching (no complex row selection logic needed)

### 5. Separate Peripherals Package

**Why**: Peripheral physics models (DAC/ADC/TIA/ChargePump) are:
- Independent of GUI (can be used in CLI tools, other apps)
- Testable in isolation
- Composable (analysis functions combine multiple peripherals)
- Documented with typical parameters from published research

## Signal Chain Details

### Complete Path: Write → Read → Compute

```
WRITE OPERATION:
Digital Level (0-30)
    ↓ [DAC: Level → Voltage]
Analog Voltage (-1.5 to +1.5V)
    ↓ [Charge Pump: Supply boost]
Boosted Voltage (higher for MOS/BJT drivers)
    ↓ [FeFET Gate: Pulse applied]
Polarization Change (level updated)

READ OPERATION:
Selected Cell (conductance G)
    ↓ [DAC: 0-0.5V safe read voltage]
Analog Voltage (0-0.5V max)
    ↓ [Crossbar: Sum currents]
Row Current (sum of G*V across columns)
    ↓ [TIA: Current → Voltage]
    V_out = I_in * 10kΩ + 5mV offset
Analog Voltage (0-1.0V max)
    ↓ [ADC: 5-bit SAR]
Digital Level (0-31)

COMPUTE (MVM):
Input Vector x[0..cols-1]
    ↓ [DAC: All columns to input vector]
Voltage Vector (0-0.5V per column)
    ↓ [Crossbar: Weight matrix]
Current Sums per row: I[r] = Σ_c(G[r,c] * x[c])
    ↓ [TIA: Current vectors → Voltage vectors]
Output Voltage Sums (0-1.0V per row)
    ↓ [ADC: All rows in parallel]
Output Vector y[0..rows-1]
    = Quantized(W * x) where W is weight matrix
```

### Nonlinearity & Error Sources

1. **DAC Nonlinearity** (INL/DNL)
   - Integral nonlinearity (code-dependent offset)
   - Differential nonlinearity (step size variation)

2. **ADC Quantization**
   - 5-bit resolution: ~1/32 of full scale
   - Rounding errors at code boundaries

3. **TIA Noise**
   - Thermal noise: 1 pA/sqrt(Hz) input-referred
   - Output offset: ±5mV

4. **Charge Pump Losses**
   - Diode drops (0.3V per stage)
   - IR drop (load-dependent)
   - Output ripple from finite capacitance

5. **Array Effects** (modeled in higher modules)
   - Sneak paths (parasitic currents)
   - Drift (state changes over time)
   - Non-ideal conductance curves

## Thread Safety & Concurrency

### GUI Thread Safety

**Key Rule**: All UI updates must use `fyne.Do(func() { ... })`

This ensures updates run on the Fyne main thread:

```go
// Wrong: potential race condition
ca.writeArrayCanvas.Refresh()

// Correct: thread-safe
fyne.Do(func() {
    ca.writeArrayCanvas.Refresh()
})
```

### Mutex Protection

CircuitsApp uses `sync.RWMutex` for multi-threaded access:

```go
type CircuitsApp struct {
    mu sync.RWMutex
    // All state fields below
    arrayRows, arrayCols int
    ...
}
```

### Compute Goroutine Pattern

For long operations (animations, mass programming):

```go
go func() {
    // Compute in background
    for step := 0; step < numSteps; step++ {
        // Computation...

        // Update UI from background thread
        fyne.Do(func() {
            ca.refreshCanvas()
        })
    }
}()
```

## Testing

### Unit Tests

**File**: `shared/peripherals/peripherals_test.go` and `shared/peripherals/analysis_test.go`

Test individual components:
- DAC/ADC conversion accuracy
- Nonlinearity modeling
- TIA saturation behavior
- Charge pump efficiency

### Integration Tests

Test complete signal chains:
- Write → Read → Verify cycle
- Input vector → Output vector (MVM)
- Saturation handling
- Voltage range clamping

**Run Tests**:
```bash
go test ./shared/peripherals/...
go test ./module4-circuits/...
```

## Key Constants & Configuration

### Physical Constants

```go
FeCIMLevels    = 30  // Demo baseline (simulation baseline)
MaxArraySize   = 128 // Maximum supported array dimension
DefaultSize    = 8   // Default 8x8 demo array
DefaultDACBits = 5   // 32 levels, use 30
DefaultADCBits = 5   // 32 levels, use 30
```

### Voltage Ranges (Typical)

```go
// Write voltages (application-specific)
vMin = 1.2V  // Just above coercive voltage
vMax = 1.5V  // Safe upper limit

// Read voltage (safe sensing)
readVoltage = 0.5V  // Non-destructive

// TIA gain
tiaGain = 10.0 kΩ  // Converts µA → mV
```

### Material Physics (from physics.yaml)

```yaml
calibration:
    field_min_ratio: 0.7    # Read max = 0.7 * Vc
    field_max_ratio: 2.5    # Write max = 2.5 * Vc
```

## Architecture Toggle (1T1R / 0T1R / 2T1R)

### Passive Mode (0T1R)

- All WLs always ON
- No transistor gating
- Simplified but fundamental change in operation
- WL checkboxes disabled in GUI

### Active Mode (1T1R)

- Standard 1-transistor-1-ferroelectric
- Single row selection via WL (typical mode)
- Checkboxes enabled for row selection

### Future: 2T1R

- Two-transistor architecture
- Additional control (not yet implemented)
- Placeholder for educational expansion

## Performance Metrics

### Typical Timing (from AnalyzeTiming)

- DAC settling: ~10 ns
- Charge pump rise: ~88 ns
- Write pulse: ~100 ns
- TIA settling: ~11 ns
- ADC conversion (5-bit SAR): 50 ns
- Total read cycle: ~76 ns
- Total write cycle: ~203 ns
- Full write+read cycle: ~279 ns

### Throughput

- Max operations: ~3.6 million operations/second (1/279ns)
- Parallel read: 128 columns simultaneously
- Parallel compute: 128x128 MVM in single cycle

### Power Estimates

- DAC: ~14.4 fJ per conversion
- ADC: ~25 fJ per conversion
- TIA: ~83 nW dynamic power (~6.3 fJ per read window)
- Charge pump: ~2.14 pJ per write (input energy, pump-dominated)

## Key Files & Line Counts

| File | Lines | Purpose |
|------|-------|---------|
| app.go | 463 | Main CircuitsApp struct |
| device_state.go | 615 | Unified simulation state |
| tab_unified.go | 1841 | Mode-First UX (READ/WRITE/COMPUTE) |
| tab_comparison.go | ~400 | Benchmark comparison view |
| tab_reference.go | ~300 | Timing diagrams + specs |
| shared/peripherals/dac.go | 90 | DAC model |
| shared/peripherals/adc.go | 123 | ADC model |
| shared/peripherals/tia.go | 101 | TIA model |
| shared/peripherals/chargepump.go | 127 | Charge pump model |
| shared/peripherals/analysis.go | 265 | System analysis tools |
| **Total GUI** | ~4400 | Complete GUI package |
| **Total Peripherals** | ~700 | Physics models + tests |

## Integration with Other Modules

### Module 1 (Hysteresis)

DeviceState uses `ferroelectric.HZOMaterial` for physics-accurate conductance:
```go
conductanceS = material.DiscreteLevel(level, quantLevels)
```

Material properties loaded from: `config/physics/physics.yaml`

### Module 2 (Crossbar)

Peripheral circuits provide:
- DAC → Crossbar drivers
- Crossbar → TIA/ADC sensing
- Can integrate full crossbar + peripherals simulation

### Module 3 (MNIST)

Network input → DAC as compute mode input vector
ADC output → Network layer computation

## Future Extensions

1. **Nonlinearity Modeling**
   - Add realistic P-V hysteresis to write operations
   - State-dependent drift over time
   - Sneak path currents in array

2. **Temperature Effects**
   - Update material properties with temperature
   - TIA noise vs. temperature
   - Coercive voltage drift

3. **Multi-Row Access Patterns**
   - Arbitrary WL combinations (beyond single/all)
   - Interactive WL selection
   - Row-stacking analysis

4. **Peripheral Optimization**
   - DAC/ADC bit resolution sweep
   - Charge pump stage count analysis
   - Power vs. performance tradeoff

5. **Silicon Integration**
   - Parasitic routing resistance
   - Clock distribution impact
   - Thermal management

## Common Pitfalls & Gotchas

### 1. Voltage Range Confusion

- **Read range**: 0 to `min(FieldMinRatio * Vc, 1.0V)` (floor 0.1V, safe)
- **Write range**: `Vc` to `FieldMaxRatio * Vc` (must exceed Vc to switch)
- Crossing Vc in read mode will destructively change state!

### 2. Passive Mode Limitations

In 0T1R mode:
- Cannot select individual rows
- All rows always active
- Compute is always parallel MVM
- Read still measures individual cells (via column sensing)

### 3. Quantization Loss

30 levels per 1.5V write range = 50 mV per level
- Write error: ±25 mV typical
- Compounds with ADC quantization on read
- Total error budget ~5-10% of full scale

### 4. Saturation Handling

TIA saturates above 100 µA:
- Indicates overcurrent (too many active weights)
- ADC level clamps to 31
- Check weight distribution if saturating

### 5. Material Physics Integration

DeviceState automatically updates voltage ranges when material changes:
```go
ds.SetMaterial(newMaterial)
// Internally recalculates readRange and writeRange
```
Ensure physics.yaml has correct calibration parameters!

---

**Last Updated**: January 2026
**Related Documentation**:
- `MODULE4-PHYSICS-IMPROVEMENTS.md` - Physics enhancements
- `/docs/development/GUI/FYNE_NOTES.md` - Fyne framework notes
- `/config/physics/physics.yaml` - Material calibration parameters
