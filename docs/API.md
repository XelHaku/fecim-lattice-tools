# FeCIM Lattice Tools API Reference

This document provides comprehensive API documentation for all public functions and types in the `shared/physics`, `shared/io`, and `shared/widgets` packages.

---

## Table of Contents

- [Package: shared/physics](#package-sharedphysics)
  - [Material Types](#material-types)
  - [Landau-Khalatnikov Solver](#landau-khalatnikov-solver)
  - [Calibration](#calibration)
  - [Conductance](#conductance)
  - [Device Variation](#device-variation)
  - [ISPP Write Controller](#ispp-write-controller)
  - [Preisach Model](#preisach-model)
  - [Quantization](#quantization)
  - [Transfer Functions](#transfer-functions)
  - [Unit Formatting](#unit-formatting)
  - [Write-Verify Statistics](#write-verify-statistics)
- [Package: shared/io](#package-sharedio)
  - [JSON Operations](#json-operations)
  - [File Utilities](#file-utilities)
- [Package: shared/widgets](#package-sharedwidgets)
  - [Layout Components](#layout-components)
  - [Display Widgets](#display-widgets)
  - [Input Widgets](#input-widgets)
  - [Material Selection](#material-selection)
  - [Demo & Status](#demo--status)
  - [Accessibility](#accessibility)
  - [Helper Functions](#helper-functions)

---

## Package: shared/physics

The physics package provides shared physics utilities for FeCIM (Ferroelectric Compute-in-Memory) simulations.

### Material Types

#### HZOMaterial

```go
type HZOMaterial struct {
    Name            string  // Material name/identifier
    Pr              float64 // Remanent polarization (C/m²)
    Ps              float64 // Saturation polarization (C/m²)
    Ec              float64 // Coercive field (V/m)
    Epsilon         float64 // Relative permittivity (high frequency)
    EpsilonLF       float64 // Low frequency permittivity
    Thickness       float64 // Film thickness (m)
    Area            float64 // Active area (m²)
    Tau             float64 // Characteristic switching time (s)
    K_dep           float64 // Depolarization coefficient (V*m/C)
    NumLevels       int     // Number of discrete analog states
    Gmin            float64 // Minimum conductance (S)
    Gmax            float64 // Maximum conductance (S)
    // ... additional fields for Landau-Khalatnikov parameters
}
```

**Material Factory Functions:**

```go
// DefaultHZO returns baseline Si-doped HfO2 parameters
func DefaultHZO() *HZOMaterial

// FeCIMMaterial returns Dr. Tour's demonstrated FeCIM values
func FeCIMMaterial() *HZOMaterial

// FeCIMMaterialTarget returns FeCIM target specifications (not yet achieved)
func FeCIMMaterialTarget() *HZOMaterial

// LiteratureSuperlattice returns academic literature best-case values
func LiteratureSuperlattice() *HZOMaterial

// CryogenicHZO returns HZO parameters at 4K for quantum computing
func CryogenicHZO() *HZOMaterial

// HZOStandard32 returns standard HZO with 32 analog states
func HZOStandard32() *HZOMaterial

// HZOFJT140 returns Ferroelectric Tunnel Junction with 140 states
func HZOFJT140() *HZOMaterial

// AlScN returns Aluminum Scandium Nitride parameters
func AlScN() *HZOMaterial

// AllMaterials returns all available CMOS-compatible materials
func AllMaterials() []*HZOMaterial
```

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    // Get FeCIM material with demonstrated values
    mat := physics.FeCIMMaterial()
    
    // Calculate coercive voltage for this material
    Vc := mat.CoerciveVoltage()
    
    // Get switching energy
    energy := mat.SwitchingEnergy()
    
    // Get temperature-dependent switching time
    tau := mat.SwitchingTime(300.0) // 300K
    
    // Get discrete conductance level
    G := mat.DiscreteLevel(15, 30) // Level 15 of 30
}
```

**Material Methods:**

| Method | Description |
|--------|-------------|
| `GetNumLevels() int` | Returns number of analog states (default: 30) |
| `CoerciveVoltage() float64` | Returns Ec × Thickness |
| `Capacitance() float64` | Returns capacitance (F) |
| `SwitchingEnergy() float64` | Returns energy for complete switching (J) |
| `SwitchingTime(T float64) float64` | Returns temperature-dependent τ (s) |
| `CoerciveFieldAtTemp(T float64) float64` | Returns Ec at temperature T |
| `PolarizationAtTemp(T float64) float64` | Returns Pr at temperature T |
| `EnduranceAtCycles(N float64) float64` | Returns Pr after N cycles |
| `RetentionAtTime(t, T float64) float64` | Returns Pr after time t at temp T |
| `DiscreteLevel(level, total int) float64` | Returns conductance for level |

---

### Landau-Khalatnikov Solver

The `LKSolver` implements the Landau-Khalatnikov equation for ferroelectric polarization dynamics.

```go
type LKSolver struct {
    Beta   float64 // First-order barrier coefficient (Negative)
    Gamma  float64 // Stability coefficient (Positive)
    Rho    float64 // Viscosity / Damping (Ohm-meters)
    K_dep  float64 // Depolarization coefficient
    P      float64 // Current Polarization (C/m²)
    PMax   float64 // Saturation polarization clamp
    // ... additional fields
}
```

**Constructor:**
```go
func NewLKSolver() *LKSolver
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Step(E, dt float64) float64` | Perform one RK4 integration step, returns new P |
| `SetState(P float64)` | Set the current polarization state |
| `GetState() float64` | Get the current polarization |
| `UpdateParams()` | Recalculate Alpha from Temperature/Stress |
| `ConfigureFromMaterial(mat *HZOMaterial)` | Configure solver from material |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    // Create solver with default parameters
    solver := physics.NewLKSolver()
    
    // Configure from material
    mat := physics.FeCIMMaterial()
    solver.ConfigureFromMaterial(mat)
    
    // Simulate polarization under electric field
    E := 1.5e8  // 1.5 MV/cm
    dt := 1e-12 // 1 ps timestep
    
    for i := 0; i < 1000; i++ {
        P := solver.Step(E, dt)
        // P is the new polarization after this step
    }
    
    // Read final state
    finalP := solver.GetState()
}
```

---

### Calibration

The calibration module implements write-verify-retry logic with binary search.

```go
type Calibrator struct {
    NumLevels   int
    Ec          float64 // Coercive field
    MinStep     float64 // Minimum step between levels
    MaxRetries  int     // Maximum retries per target
    Tolerance   int     // Tolerance for success in levels
    Up          []CalibrationState // Ascending calibration
    Down        []CalibrationState // Descending calibration
}
```

**Constructor:**
```go
func NewCalibrator(numLevels int, Ec float64) *Calibrator
```

**Methods:**

| Method | Description |
|--------|-------------|
| `UpdateAscending(levelIdx, levelError int) float64` | Binary search for ascending |
| `UpdateDescending(levelIdx, levelError int) float64` | Binary search for descending |
| `FieldConstraintsAscending(levelIdx int) (min, max float64)` | Get field bounds |
| `FieldConstraintsDescending(levelIdx int) (min, max float64)` | Get field bounds |
| `EnforceGlobalMonotonicity()` | Fix non-monotonic values |
| `CheckVerify(target, read, retry int) VerifyResult` | Check write success |
| `GetAscendingValues() []float64` | Get all ascending values |
| `GetDescendingValues() []float64` | Get all descending values |
| `SetAscendingValues(values []float64)` | Set ascending values |
| `SetDescendingValues(values []float64)` | Set descending values |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    Ec := 1.0e8 // 1 MV/cm
    cal := physics.NewCalibrator(30, Ec)
    
    // Simulate write-verify-retry loop
    targetLevel := 15
    for retry := 0; retry < cal.MaxRetries; retry++ {
        // Write with current field value
        field := cal.Up[targetLevel].Value
        // ... apply field and read back ...
        readLevel := 14 // Simulated read
        
        result := cal.CheckVerify(targetLevel, readLevel, retry)
        if result.Success {
            break
        }
        if result.ShouldRetry {
            // Update field using binary search
            cal.UpdateAscending(targetLevel, result.Error)
        }
    }
}
```

---

### Conductance

The conductance module provides conductance conversion and modeling.

**Constants:**
```go
const (
    GMin   = 10e-6  // 10 µS (OFF state)
    GMax   = 100e-6 // 100 µS (ON state)
    GRatio = GMax / GMin
)
```

**Types:**
```go
type ConductanceModel int

const (
    ConductanceLinear      ConductanceModel = iota // G = Gmin + gNorm*(Gmax-Gmin)
    ConductanceExponential                         // G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
    ConductanceLookup                              // Use calibration table
)
```

**Functions:**

| Function | Description |
|----------|-------------|
| `NormalizedToPhysical(gNorm float64, model ConductanceModel) float64` | Convert [0,1] to Siemens |
| `NormalizedToPhysicalRange(gNorm float64, model ConductanceModel, gMin, gMax float64, table []float64) float64` | With custom range |
| `PhysicalToNormalized(gPhys float64) float64` | Convert Siemens to [0,1] |
| `PhysicalToNormalizedRange(gPhys, gMin, gMax float64) float64` | With custom range |
| `ConductanceToLevel(gPhys float64, levels int) int` | Convert to discrete level |
| `LevelToConductance(level, levels int, model ConductanceModel) float64` | Convert level to conductance |
| `ConductanceWindow() float64` | Returns GMax - GMin |
| `ConductanceMidpoint() float64` | Returns (GMax + GMin) / 2 |
| `ConductanceGeometricMean() float64` | Returns sqrt(GMin * GMax) |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    // Convert normalized [0,1] to physical conductance
    gPhys := physics.NormalizedToPhysical(0.5, physics.ConductanceExponential)
    // For exponential model, 0.5 gives geometric mean ≈ 31.6 µS
    
    // Convert back
    gNorm := physics.PhysicalToNormalized(gPhys)
    
    // Get discrete level
    level := physics.ConductanceToLevel(50e-6, 30) // 50 µS → level ~14
}
```

---

### Device Variation

Models device-to-device variation in Ec and Pr parameters using Gaussian distribution.

```go
type DeviceVariationConfig struct {
    Enable          bool
    EcSigmaRelative float64 // σ_Ec/Ec (default: 0.15 = 15%)
    PrSigmaRelative float64 // σ_Pr/Pr (default: 0.20 = 20%)
    EcPrCorrelation float64 // Correlation coefficient
    Seed            int64   // Random seed (0 = time-based)
}

type DeviceVariation struct {
    EcFactor float64 // Multiplicative factor for Ec
    PrFactor float64 // Multiplicative factor for Pr
}
```

**Constructor:**
```go
func DefaultDeviceVariationConfig() *DeviceVariationConfig
func NewDeviceVariationEngine(config *DeviceVariationConfig) *DeviceVariationEngine
```

**Methods:**

| Method | Description |
|--------|-------------|
| `GetDeviceVariation(row, col int) *DeviceVariation` | Get Ec/Pr factors for device |
| `ApplyToMaterial(base *HZOMaterial, row, col int) *HZOMaterial` | Create varied material |
| `GetArrayVariationStats(rows, cols int) *VariationStats` | Get array statistics |
| `EstimateYield(rows, cols int, maxDeviation float64) float64` | Estimate yield |
| `Reset()` | Clear device cache |
| `SetSeed(seed int64)` | Set new random seed |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    config := physics.DefaultDeviceVariationConfig()
    config.Enable = true
    config.Seed = 42
    
    engine := physics.NewDeviceVariationEngine(config)
    base := physics.FeCIMMaterial()
    
    // Get material with device-specific variation
    varied := engine.ApplyToMaterial(base, 5, 10) // Device at row 5, col 10
    
    // Estimate array yield (devices within ±30% of nominal)
    yield := engine.EstimateYield(64, 64, 0.30)
}
```

---

### ISPP Write Controller

Implements Incremental Step Pulse Programming with binary search.

```go
type WriteController struct {
    Solver        *LKSolver
    Material      *HZOMaterial
    MaxVoltage    float64
    MinVoltage    float64
    PulseWidth    float64
    Tolerance     float64
    MaxIterations int
}
```

**Constructor:**
```go
func NewWriteController(solver *LKSolver, material *HZOMaterial) *WriteController
```

**Methods:**

| Method | Description |
|--------|-------------|
| `WriteTarget(targetG float64) (attempts int, success bool, overshootCount int)` | Write to target conductance |
| `WriteTargetWithReset(targetG float64, reset bool) (int, bool, int)` | Write with optional reset |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    solver := physics.NewLKSolver()
    mat := physics.FeCIMMaterial()
    solver.ConfigureFromMaterial(mat)
    
    ctrl := physics.NewWriteController(solver, mat)
    ctrl.MaxVoltage = 3.0      // 3V max
    ctrl.Tolerance = 0.01      // 1% tolerance
    ctrl.MaxIterations = 20
    
    // Write to 50 µS conductance
    attempts, success, overshoots := ctrl.WriteTarget(50e-6)
    if success {
        // Write succeeded in 'attempts' pulses
    }
}
```

---

### Preisach Model

Implements the Preisach hysteresis model with memory effects.

```go
type EverettFunction interface {
    Calculate(alpha, beta float64) float64
}

type PreisachStack struct {
    Stack       []TurningPoint
    CurrentDir  int
    LastE       float64
    SaturationE float64
    Everett     EverettFunction
}
```

**Constructor:**
```go
func NewPreisachStack(saturationE float64, everett EverettFunction) *PreisachStack
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Update(E float64) float64` | Process new field value, returns new P |
| `ComputePolarization(currentE float64) float64` | Calculate polarization from history |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

// Implement custom Everett function
type MyEverett struct {
    Sat float64
}

func (e MyEverett) Calculate(alpha, beta float64) float64 {
    return (alpha - beta) / (2 * e.Sat)
}

func main() {
    sat := 1.0e8 // Saturation field
    everett := MyEverett{Sat: sat}
    stack := physics.NewPreisachStack(sat, everett)
    
    // Apply triangular field waveform
    for E := -sat; E <= sat; E += sat * 0.01 {
        P := stack.Update(E)
        // P follows hysteresis curve
    }
}
```

---

### Quantization

Utilities for quantizing analog values to discrete FeCIM levels.

**Constants:**
```go
const (
    DefaultLevels = 30   // Default 30 analog states
    BitsPerCell   = 4.91 // log2(30)
)
```

**Functions:**

| Function | Description |
|----------|-------------|
| `QuantizeToLevels(value float64, levels int) float64` | Quantize [0,1] to N levels |
| `QuantizeTo30Levels(value float64) float64` | Quantize to 30 levels |
| `GetLevel(conductance float64, levels int) int` | Get discrete level index |
| `GetLevelFor30(conductance float64) int` | Get level for 30-level system |
| `NormalizeFromLevel(level, levels int) float64` | Convert level to [0,1] |
| `NormalizeFromLevel30(level int) float64` | Convert 30-level to [0,1] |
| `LevelSpacing(levels int) float64` | Get spacing between levels |
| `QuantizationError(levels int) float64` | Get max quantization error |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    // Quantize analog value to 30 levels
    quantized := physics.QuantizeTo30Levels(0.33)
    
    // Get discrete level index (0-29)
    level := physics.GetLevelFor30(0.5) // Returns 15
    
    // Convert level back to normalized value
    normalized := physics.NormalizeFromLevel30(15) // Returns ~0.517
    
    // Get level spacing
    spacing := physics.LevelSpacing(30) // Returns ~0.0345
}
```

---

### Transfer Functions

Convert between polarization and conductance.

```go
// Convert polarization to conductance
func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64

// Convert conductance to polarization
func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64
```

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    Ps := 0.35     // Saturation polarization
    Gmin := 1e-6   // 1 µS
    Gmax := 100e-6 // 100 µS
    
    // Polarization at +Ps gives Gmax
    G := physics.PolarizationToConductance(Ps, Ps, Gmin, Gmax)
    
    // Convert back
    P := physics.ConductanceToPolarization(G, Gmin, Gmax, Ps)
}
```

---

### Unit Formatting

Format physical quantities with appropriate SI prefixes.

| Function | Input | Example Output |
|----------|-------|----------------|
| `FormatEnergy(joules float64) string` | Joules | "1.50 fJ", "2.30 pJ", "4.50 nJ" |
| `FormatConductance(siemens float64) string` | Siemens | "1.00 nS", "50.00 µS", "1.00 mS" |
| `FormatCurrent(amperes float64) string` | Amperes | "1.00 pA", "50.00 nA", "1.00 µA" |
| `FormatVoltage(volts float64) string` | Volts | "1.00 µV", "1.00 mV", "1.50 V" |
| `FormatTime(seconds float64) string` | Seconds | "1.00 ps", "1.00 ns", "1.00 µs" |
| `FormatFrequency(hz float64) string` | Hertz | "1.00 kHz", "1.00 MHz", "1.00 GHz" |
| `FormatResistance(ohms float64) string` | Ohms | "1.00 mΩ", "100.00 Ω", "4.70 kΩ" |
| `FormatCapacitance(farads float64) string` | Farads | "1.00 aF", "1.00 fF", "1.00 pF" |
| `FormatPower(watts float64) string` | Watts | "1.00 fW", "1.00 nW", "1.00 mW" |
| `FormatCharge(coulombs float64) string` | Coulombs | "1.00 fC", "1.00 pC", "1.00 nC" |
| `FormatPolarization(cm2 float64) string` | C/m² | "20.0 µC/cm²", "35.0 µC/cm²" |
| `FormatElectricField(vm float64) string` | V/m | "1.00 MV/cm", "500.00 kV/cm" |

**Example:**
```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/physics"
)

func main() {
    energy := 1.5e-15
    fmt.Println(physics.FormatEnergy(energy)) // "1.50 fJ"
    
    cond := 50e-6
    fmt.Println(physics.FormatConductance(cond)) // "50.00 µS"
    
    field := 1.2e8
    fmt.Println(physics.FormatElectricField(field)) // "1.20 MV/cm"
}
```

---

### Write-Verify Statistics

Track statistics for ISPP write operations.

```go
type WriteVerifyStats struct {
    TotalWrites      int
    SuccessfulWrites int
    FailedWrites     int
    OvershootCount   int
    ResetCount       int
    CycleCount       int
}
```

**Constructor:**
```go
func NewWriteVerifyStats() *WriteVerifyStats
```

**Methods:**

| Method | Description |
|--------|-------------|
| `RecordWrite(targetLevel, pulsesUsed int, success, hadOvershoot bool)` | Record write operation |
| `RecordReset()` | Record reset operation |
| `GetSuccessRate() float64` | Get success rate (0-1) |
| `GetFailureRate() float64` | Get failure rate (0-1) |
| `GetAveragePulses() float64` | Get average pulses per write |
| `GetPulsesHistogram() [10]int` | Get pulses histogram |
| `GetLevelSuccessRates() [256]float64` | Get per-level success rates |
| `GetHardestLevels(n int) []int` | Get n hardest levels to write |
| `GetOvershootRate() float64` | Get overshoot rate |
| `GetSummary() string` | Get human-readable summary |
| `Reset()` | Clear all statistics |

**Example:**
```go
package main

import "fecim-lattice-tools/shared/physics"

func main() {
    stats := physics.NewWriteVerifyStats()
    
    // Record successful write
    stats.RecordWrite(15, 3, true, false)
    
    // Record failed write with overshoot
    stats.RecordWrite(20, 10, false, true)
    
    // Get statistics
    successRate := stats.GetSuccessRate()
    avgPulses := stats.GetAveragePulses()
    summary := stats.GetSummary()
}
```

---

## Package: shared/io

The io package provides shared file I/O utilities.

### JSON Operations

```go
// SaveJSON writes data to a JSON file with pretty formatting.
// Creates parent directories if needed.
func SaveJSON(path string, data interface{}) error

// LoadJSON reads a JSON file and unmarshals into target.
func LoadJSON(path string, target interface{}) error

// SaveJSONCompact writes data to JSON without formatting.
// Useful for large files where size matters.
func SaveJSONCompact(path string, data interface{}) error
```

**Example:**
```go
package main

import "fecim-lattice-tools/shared/io"

type Config struct {
    Name   string `json:"name"`
    Levels int    `json:"levels"`
}

func main() {
    // Save configuration
    config := Config{Name: "HZO", Levels: 30}
    err := io.SaveJSON("config.json", config)
    if err != nil {
        panic(err)
    }
    
    // Load configuration
    var loaded Config
    err = io.LoadJSON("config.json", &loaded)
    if err != nil {
        panic(err)
    }
}
```

### File Utilities

```go
// FileExists checks if a file exists and is not a directory.
func FileExists(path string) bool

// DirExists checks if a directory exists.
func DirExists(path string) bool

// EnsureDir creates a directory and parents if they don't exist.
func EnsureDir(path string) error
```

**Example:**
```go
package main

import "fecim-lattice-tools/shared/io"

func main() {
    if !io.FileExists("config.json") {
        // Create default config
    }
    
    err := io.EnsureDir("output/results")
    if err != nil {
        panic(err)
    }
}
```

---

## Package: shared/widgets

The widgets package provides reusable Fyne GUI components for FeCIM visualizers.

### Layout Components

#### AdaptiveLayout

Provides responsive layout that adapts to screen size.

```go
type AdaptiveLayout struct {
    // Maintains desktop (splits) and mobile (tabs) layouts
}
```

**Constructor:**
```go
func NewAdaptiveLayout(zones []fyne.CanvasObject, tabLabels []string) *AdaptiveLayout
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetDesktopLayout(builder func(zones []fyne.CanvasObject) fyne.CanvasObject)` | Set desktop layout builder |
| `Content() fyne.CanvasObject` | Get the main content container |

#### ResponsiveGridLayout

Adaptive grid layout for material cards and similar components.

```go
type ResponsiveGridLayout struct {
    MinItemWidth float32 // Minimum card width (default: 280)
    MaxItemWidth float32 // Maximum card width (default: 450)
    RowSpacing   float32
    ColSpacing   float32
}
```

**Constructor:**
```go
func NewResponsiveGridLayout() *ResponsiveGridLayout
```

#### ResizeDetector

Fires callbacks when widget size changes.

```go
func NewResizeDetector(onResize func(size fyne.Size)) *ResizeDetector
```

**Example:**
```go
detector := widgets.NewResizeDetector(func(size fyne.Size) {
    if size.Width < 768 {
        // Switch to mobile layout
    }
})
```

---

### Display Widgets

#### ColorLegend

Displays a gradient bar with min/max labels.

```go
func NewColorLegend(minValue, maxValue float64, units string, vertical bool, 
    colorFunc func(float64) color.RGBA) *ColorLegend

func NewColorLegendWithColormap(minValue, maxValue float64, units string, 
    vertical bool, colormapName string) *ColorLegend
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetRange(minValue, maxValue float64)` | Update min/max values |

#### EducationalPanel

Displays educational content with title and scrollable body.

```go
type EducationalPanelConfig struct {
    Title   string
    Content string
    MinSize fyne.Size
}

func NewEducationalPanel(config EducationalPanelConfig) *EducationalPanel
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetContent(title, content string)` | Update content |
| `SetTitle(title string)` | Update title only |
| `AppendContent(text string)` | Append to content |
| `GetContent() (string, string)` | Get title and content |

#### KeyStat

Displays a key statistic prominently (label + value).

```go
type KeyStatConfig struct {
    Label   string
    Value   string
    MinSize fyne.Size
}

func NewKeyStat(config KeyStatConfig) *KeyStat
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetValue(value string)` | Update value |
| `SetLabel(label string)` | Update label |
| `SetLabelAndValue(label, value string)` | Update both |

#### ModeIndicator

Displays operation mode with colored background.

```go
type ModeStyle struct {
    BackgroundColor color.RGBA
    BorderColor     color.RGBA
    Text            string
}

type ModeIndicatorConfig struct {
    MinSize      fyne.Size
    DefaultStyle ModeStyle
    Styles       map[int]ModeStyle
}

func NewModeIndicator(config ModeIndicatorConfig) *ModeIndicator
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetMode(mode int)` | Set current mode |
| `GetMode() int` | Get current mode |
| `SetStyle(mode int, style ModeStyle)` | Set style for mode |

#### OperationLog

Displays timestamped operation history.

```go
type OperationLogConfig struct {
    Title      string
    MaxEntries int
    MinSize    fyne.Size
}

func NewOperationLog(config OperationLogConfig) *OperationLog
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Add(entry string)` | Add entry with timestamp |
| `AddWithPrefix(prefix, entry string)` | Add with custom prefix |
| `Clear()` | Clear all entries |

#### MicLevel

Microphone level indicator widget.

```go
func NewMicLevel() *MicLevel
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetLevel(level int)` | Set level (0-100) |
| `SetPeak(peak int)` | Set peak level (0-100) |
| `SetLevelAndPeak(level, peak int)` | Set both |
| `SetMonitoring(monitoring bool)` | Set monitoring state |
| `SetRecording(recording bool)` | Set recording state |

---

### Input Widgets

#### ArchitectureSelector

Dropdown selector for crossbar architecture (1T1R, 0T1R, 2T1R).

```go
func NewArchitectureSelector(onChanged func(architecture string)) *ArchitectureSelector
```

**Methods:**

| Method | Description |
|--------|-------------|
| `GetArchitecture() string` | Get current selection |
| `SetArchitecture(arch string)` | Set selection |
| `Is1T1R() bool` | Check if 1T1R selected |
| `Is2T1R() bool` | Check if 2T1R selected |

**Constants:**
```go
const (
    Architecture1T1R = "1T1R (Transistor)"
    Architecture0T1R = "0T1R (Passive)"
    Architecture2T1R = "2T1R (Dual Transistor)"
)
```

#### ArchitectureToggle

Button group for architecture selection.

```go
type ArchitectureToggleOptions struct {
    Initial   string
    OnChanged func(architecture string)
}

func NewArchitectureToggle(opts ArchitectureToggleOptions) *ArchitectureToggle
```

---

### Material Selection

#### MaterialPicker

Dialog for selecting ferroelectric materials.

```go
func NewMaterialPicker(onSelected func(string, *physics.Material)) *MaterialPicker
```

#### MaterialCard

Compact material summary for list/grid display.

```go
func NewMaterialCard(materialID string, material *physics.Material, 
    onTapped func(string)) *MaterialCard
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetSelected(selected bool)` | Set selection state |
| `IsSelected() bool` | Get selection state |
| `GetMaterialID() string` | Get material ID |

#### MaterialTable

Displays all properties of a material in tabs.

```go
func NewMaterialTable(material *physics.Material) *MaterialTable
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetMaterial(material *physics.Material)` | Update displayed material |

---

### Demo & Status

#### DemoController

Manages automated demo playback.

```go
type DemoStep struct {
    Name     string
    Duration time.Duration
    Action   func()
}

func NewDemoController(steps []DemoStep) *DemoController
func NewLoopingDemoController(steps []DemoStep) *DemoController
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Start()` | Begin demo playback |
| `Stop()` | Stop demo |
| `Pause()` | Pause demo |
| `Resume()` | Resume demo |
| `SetLoop(loop bool)` | Enable/disable looping |
| `SetOnStart(fn func())` | Set start callback |
| `SetOnStop(fn func())` | Set stop callback |
| `SetOnStepDone(fn func(int, DemoStep))` | Set step callback |

**Example:**
```go
steps := []widgets.DemoStep{
    {Name: "Initialize", Duration: 2*time.Second, Action: func() { 
        // Setup 
    }},
    {Name: "Write", Duration: 3*time.Second, Action: func() { 
        // Perform write 
    }},
    {Name: "Read", Duration: 2*time.Second, Action: func() { 
        // Perform read 
    }},
}

demo := widgets.NewLoopingDemoController(steps)
demo.Start()
```

#### StatusBar

Thread-safe status updates.

```go
func NewStatusBar(prefix string) *StatusBar
func NewStatusBarWithLabel(label *widget.Label, prefix string) *StatusBar
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Update(msg string)` | Set status text (thread-safe) |
| `Updatef(format string, args ...interface{})` | Format and set status |
| `GetLabel() *widget.Label` | Get underlying label |

**Example:**
```go
status := widgets.NewStatusBar("Status: ")

// Safe to call from any goroutine
go func() {
    status.Update("Processing...")
    // ... work ...
    status.Update("Complete")
}()
```

#### EmbeddedApp Interface

Interface for embeddable module applications.

```go
type EmbeddedApp interface {
    BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject
    Start()
    Stop()
}
```

#### EmbeddedAppBase

Base implementation for embedded apps.

```go
type EmbeddedAppBase struct {
    // Common embedded app functionality
}
```

**Methods:**

| Method | Description |
|--------|-------------|
| `Init(fyneApp fyne.App, window fyne.Window)` | Initialize base |
| `GetFyneApp() fyne.App` | Get Fyne app |
| `GetWindow() fyne.Window` | Get parent window |
| `SetStatusBar(status *StatusBar)` | Set status bar |
| `UpdateStatus(msg string)` | Update status |
| `Start()` | Mark as running |
| `Stop()` | Mark as stopped |
| `IsRunning() bool` | Check running state |

---

### Accessibility

#### FocusIndicator

Wraps a widget with visible focus indicator.

```go
func NewFocusIndicator(content fyne.CanvasObject) *FocusIndicator
```

**Methods:**

| Method | Description |
|--------|-------------|
| `SetFocused(focused bool)` | Set focus state |

#### Accessible Colors

Pre-computed WCAG AA compliant colors:

```go
var AccessibleColors = struct {
    HintText    color.RGBA // For placeholder/hint text
    GridLines   color.RGBA // For subtle grid overlays
    DimText     color.RGBA // For secondary text
    SuccessText color.RGBA // Green for success
    WarningText color.RGBA // Yellow for warnings
    ErrorText   color.RGBA // Red for errors
    InfoText    color.RGBA // Blue for info
}
```

---

### Helper Functions

#### Layout Helpers

```go
// CreateLabeledBox creates a styled box with title and value
func CreateLabeledBox(title, value string, bgColor color.Color) *fyne.Container

// CreateLabeledBoxWithLabel creates box with dynamic label
func CreateLabeledBoxWithLabel(title string, valueLbl *widget.Label, 
    bgColor color.Color) *fyne.Container

// CreateSectionDivider creates a horizontal divider line
func CreateSectionDivider(dividerColor color.Color) fyne.CanvasObject

// CreateSectionHeader creates a bold header with separator
func CreateSectionHeader(title string) *fyne.Container
```

#### UI Helpers

```go
// SafeUpdateLabel updates label from any goroutine
func SafeUpdateLabel(label *widget.Label, text string)

// SafeUpdateProgress updates progress bar from any goroutine
func SafeUpdateProgress(progress *widget.ProgressBar, value float64)

// SafeRefresh refreshes canvas object from any goroutine
func SafeRefresh(obj fyne.CanvasObject)

// SafeShow shows canvas object from any goroutine
func SafeShow(obj fyne.CanvasObject)

// SafeHide hides canvas object from any goroutine
func SafeHide(obj fyne.CanvasObject)

// SafeEnable enables widget from any goroutine
func SafeEnable(w fyne.Disableable)

// SafeDisable disables widget from any goroutine
func SafeDisable(w fyne.Disableable)
```

#### Colormap Functions

```go
// GetColormapFunc returns a colormap function by name
func GetColormapFunc(name string) func(float64) color.RGBA
```

Available colormaps: `"viridis"`, `"plasma"`, `"inferno"`, `"magma"`, `"cividis"`, `"turbo"`, `"coolwarm"`, `"spectral"`

---

## Complete Example

Here's a complete example that combines multiple packages:

```go
package main

import (
    "fmt"
    "time"
    
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    
    "fecim-lattice-tools/shared/io"
    "fecim-lattice-tools/shared/physics"
    "fecim-lattice-tools/shared/widgets"
)

func main() {
    // Initialize physics
    mat := physics.FeCIMMaterial()
    solver := physics.NewLKSolver()
    solver.ConfigureFromMaterial(mat)
    ctrl := physics.NewWriteController(solver, mat)
    
    // Track statistics
    stats := physics.NewWriteVerifyStats()
    
    // Create Fyne app
    a := app.New()
    w := a.NewWindow("FeCIM Demo")
    
    // Create widgets
    status := widgets.NewStatusBar("Status: ")
    keyStat := widgets.NewKeyStat(widgets.KeyStatConfig{
        Label: "Success Rate",
        Value: "N/A",
    })
    log := widgets.NewOperationLog(widgets.OperationLogConfig{
        Title:      "Write Operations",
        MaxEntries: 10,
    })
    
    // Create demo controller
    demo := widgets.NewDemoController([]widgets.DemoStep{
        {
            Name:     "Write Level 15",
            Duration: 2 * time.Second,
            Action: func() {
                fyne.Do(func() { status.Update("Writing...") })
                attempts, success, _ := ctrl.WriteTarget(mat.Gmin + (mat.Gmax-mat.Gmin)*0.5)
                stats.RecordWrite(15, attempts, success, false)
                
                fyne.Do(func() {
                    log.Add(fmt.Sprintf("Level 15: %v (%d pulses)", success, attempts))
                    keyStat.SetValue(fmt.Sprintf("%.1f%%", stats.GetSuccessRate()*100))
                    status.Update("Ready")
                })
            },
        },
    })
    
    // Layout
    content := container.NewVBox(
        status.GetLabel(),
        keyStat,
        log,
    )
    
    w.SetContent(content)
    w.Resize(fyne.NewSize(400, 300))
    
    // Start demo
    demo.SetLoop(true)
    demo.Start()
    
    // Save results on close
    w.SetOnClosed(func() {
        demo.Stop()
        io.SaveJSON("results.json", map[string]interface{}{
            "successRate":   stats.GetSuccessRate(),
            "averagePulses": stats.GetAveragePulses(),
        })
    })
    
    w.ShowAndRun()
}
```

---

## See Also

- [README.md](../README.md) - Project overview
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [config/physics.yaml](../config/physics/) - Material configuration
