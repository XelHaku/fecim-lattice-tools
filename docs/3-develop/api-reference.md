# FeCIM Lattice Tools API Reference

> **Note:** This file was previously located at `docs/API.md` / `docs/API_REFERENCE.md`. It has moved to `docs/3-develop/api-reference.md`.

Comprehensive API guide for the core simulation packages and GUI components.

## Scope

- `shared/physics` - Core physics models and utilities
- `shared/peripherals` - Peripheral models (DAC, ADC, TIA, charge pump)
- `shared/io` - File I/O utilities
- `shared/widgets` - GUI components and helpers
- `module1-hysteresis/pkg/ferroelectric` - Hysteresis and Preisach modeling
- `module2-crossbar/pkg/crossbar` - Crossbar array simulation
- `module3-mnist/pkg/core` - MNIST neural network simulation

---

## Table of Contents

1. [shared/physics](#sharedphysics)
2. [shared/peripherals](#sharedperipherals)
3. [shared/io](#sharedio)
4. [shared/widgets](#sharedwidgets)
5. [module1-hysteresis/pkg/ferroelectric](#module1-hysteresispkgferroelectric)
6. [module2-crossbar/pkg/crossbar](#module2-crossbarpkgcrossbar)
7. [module3-mnist/pkg/core](#module3-mnistpkgcore)

---

## shared/physics

Import path:

```go
import "fecim-lattice-tools/shared/physics"
```

Core physics models for FeCIM (Ferroelectric Compute-in-Memory) simulations.

### Key Types

#### HZOMaterial

Canonical ferroelectric material model with physical, thermal, and reliability parameters.

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

#### LKSolver

Landau-Khalatnikov dynamic solver for polarization state evolution.

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

#### WriteController

Closed-loop write engine that iteratively drives target conductance/polarization.

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

#### PreisachStack

Generic Preisach hysteresis memory stack with wipe-out/turning-point logic.

```go
type PreisachStack struct {
    Stack       []TurningPoint
    CurrentDir  int
    LastE       float64
    SaturationE float64
    Everett     EverettFunction
}
```

#### Additional Types

- `type ConductanceModel int` - Mapping mode (Linear, Exponential, Lookup)
- `type Calibrator struct { ... }` - Level-wise calibration with monotonic constraints
- `type WriteVerifyStats struct { ... }` - Aggregated write/verify metrics
- `type DeviceVariationEngine struct { ... }` - Spatially-correlated device variation
- `type CellGeometry struct { ... }` - Field/charge/current conversion helper

### Material Presets

```go
// Baseline Si-doped HfO2 parameters
func DefaultHZO() *HZOMaterial

// Demonstrated FeCIM values
func FeCIMMaterial() *HZOMaterial

// Target/aspirational FeCIM specifications
func FeCIMMaterialTarget() *HZOMaterial

// Academic literature best-case values
func LiteratureSuperlattice() *HZOMaterial

// HZO parameters at 4K for quantum computing
func CryogenicHZO() *HZOMaterial

// Standard HZO with 32 analog states
func HZOStandard32() *HZOMaterial

// Ferroelectric Tunnel Junction with 140 states
func HZOFJT140() *HZOMaterial

// 14-level compact preset
func HZOCustom14() *HZOMaterial

// Aluminum Scandium Nitride parameters
func AlScN() *HZOMaterial

// All available CMOS-compatible materials
func AllMaterials() []*HZOMaterial
```

### Material Methods

| Method | Description |
|--------|-------------|
| `GetNumLevels() int` | Returns number of analog states |
| `CoerciveVoltage() float64` | Returns Ec × Thickness |
| `Capacitance() float64` | Returns capacitance (F) |
| `SwitchingEnergy() float64` | Returns energy for complete switching (J) |
| `SwitchingTime(T float64) float64` | Returns temperature-dependent τ (s) |
| `CoerciveFieldAtTemp(T float64) float64` | Returns Ec at temperature T |
| `PolarizationAtTemp(T float64) float64` | Returns Pr at temperature T |
| `EnduranceAtCycles(N float64) float64` | Returns Pr after N cycles |
| `RetentionAtTime(t, T float64) float64` | Returns Pr after time t at temp T |
| `DiscreteLevel(level, totalLevels int) float64` | Returns conductance for level |

### LK Solver Methods

| Method | Description |
|--------|-------------|
| `Step(E, dt float64) float64` | Perform one RK4 integration step |
| `SetState(P float64)` | Set the current polarization state |
| `GetState() float64` | Get the current polarization |
| `UpdateParams()` | Recalculate Alpha from Temperature/Stress |
| `ConfigureFromMaterial(mat *HZOMaterial)` | Configure solver from material |

### WriteController Methods

| Method | Description |
|--------|-------------|
| `WriteTarget(targetG float64)` | Returns (attempts, success, overshootCount) |
| `WriteTargetWithReset(targetG float64, reset bool)` | Optional pre-reset before write |

### Calibration Methods

| Method | Description |
|--------|-------------|
| `UpdateAscending(levelIdx, levelError int) float64` | Binary search for ascending |
| `UpdateDescending(levelIdx, levelError int) float64` | Binary search for descending |
| `FieldConstraintsAscending(levelIdx int) (min, max float64)` | Get field bounds |
| `FieldConstraintsDescending(levelIdx int) (min, max float64)` | Get field bounds |
| `EnforceGlobalMonotonicity()` | Fix non-monotonic values |
| `CheckVerify(target, read, retry int) VerifyResult` | Check write success |

### Conductance Functions

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
    ConductanceLinear      ConductanceModel = iota
    ConductanceExponential
    ConductanceLookup
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

### Device Variation API

```go
type DeviceVariationConfig struct {
    Enable          bool
    EcSigmaRelative float64 // σ_Ec/Ec (default: 0.15 = 15%)
    PrSigmaRelative float64 // σ_Pr/Pr (default: 0.20 = 20%)
    EcPrCorrelation float64 // Correlation coefficient
    Seed            int64   // Random seed (0 = time-based)
}
```

**Methods:**

| Method | Description |
|--------|-------------|
| `GetDeviceVariation(row, col int) *DeviceVariation` | Get Ec/Pr factors for device |
| `ApplyToMaterial(base *HZOMaterial, row, col int) *HZOMaterial` | Create varied material |
| `GetArrayVariationStats(rows, cols int) *VariationStats` | Get array statistics |
| `EstimateYield(rows, cols int, maxDeviation float64) float64` | Estimate yield |

### Preisach Methods

| Method | Description |
|--------|-------------|
| `Update(E float64) float64` | Process new field value, returns new P |
| `ComputePolarization(currentE float64) float64` | Calculate polarization from history |

### Quantization Functions

**Constants:**

```go
const (
    DefaultLevels = 30   // Default 30 analog states (simulation baseline)
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

### Transfer Functions

```go
// Convert polarization to conductance
func PolarizationToConductance(P, Ps, Gmin, Gmax float64) float64

// Convert conductance to polarization
func ConductanceToPolarization(G, Gmin, Gmax, Ps float64) float64
```

### Unit Formatting

Format physical quantities with appropriate SI prefixes.

| Function | Input | Example Output |
|----------|-------|----------------|
| `FormatEnergy(joules float64) string` | Joules | "1.50 fJ", "2.30 pJ" |
| `FormatConductance(siemens float64) string` | Siemens | "50.00 µS", "1.00 mS" |
| `FormatCurrent(amperes float64) string` | Amperes | "50.00 nA", "1.00 µA" |
| `FormatVoltage(volts float64) string` | Volts | "1.00 mV", "1.50 V" |
| `FormatTime(seconds float64) string` | Seconds | "1.00 ps", "1.00 ns" |
| `FormatFrequency(hz float64) string` | Hertz | "1.00 kHz", "1.00 MHz" |
| `FormatResistance(ohms float64) string` | Ohms | "100.00 Ω", "4.70 kΩ" |
| `FormatCapacitance(farads float64) string` | Farads | "1.00 fF", "1.00 pF" |
| `FormatPower(watts float64) string` | Watts | "1.00 nW", "1.00 mW" |
| `FormatCharge(coulombs float64) string` | Coulombs | "1.00 pC", "1.00 nC" |
| `FormatPolarization(cm2 float64) string` | C/m² | "20.0 µC/cm²", "35.0 µC/cm²" |
| `FormatElectricField(vm float64) string` | V/m | "1.00 MV/cm", "500.00 kV/cm" |

### Write-Verify Statistics

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

### World-Class Characterization APIs

These APIs implement standard ferroelectric characterization techniques. All are in `shared/physics/worldclass_*.go`.

#### PUND Measurement (`worldclass_pund.go`)

Separates switching charge from non-switching (linear) charge using the P/U/N/D pulse sequence.

```go
type PUNDResult struct {
    QP_C                float64 // Program pulse charge (C)
    QU_C                float64 // Up baseline pulse charge (C)
    QN_C                float64 // Negative pulse charge (C)
    QD_C                float64 // Down baseline pulse charge (C)
    SwitchingPositive_C float64 // QP - QU (ferroelectric contribution)
    SwitchingNegative_C float64 // QN - QD
}

func AnalyzePUND(programP, upU, negativeN, downD []PulseSample) (PUNDResult, error)
func IntegrateCurrent(samples []PulseSample) (float64, error)
```

#### Retention Model (`worldclass_retention.go`)

Power-law and exponential retention decay models.

```go
type RetentionPoint struct {
    TimeS           float64 // Hold time (s)
    Polarization_Cm float64 // P at hold time (C/m²)
}

// P(t) = P0 * (t/t0)^(-beta).  beta ≈ 0.01-0.05 for HZO.
func SimulateRetentionPowerLaw(P0_Cm, t0S, beta float64, timesS []float64) ([]RetentionPoint, error)

// P(t) = Pinf + (P0-Pinf)*exp(-t/tau)
func SimulateRetentionExponential(initialP_Cm, asymptoticP_Cm, tauS float64, timesS []float64) ([]RetentionPoint, error)

func GenerateLogTimeSweep(tMinS, tMaxS float64, points int) ([]float64, error)
```

#### Wake-Up and Fatigue (`worldclass_wakeup.go`)

Two-phase Pr(N) model: initial wake-up then slow fatigue.

```go
type WakeUpModelConfig struct {
    PrInitial_Cm2      float64 // Pr before any cycling
    WakeUpGainFraction float64 // Fractional Pr gain at full wake-up
    WakeUpTauCycles    float64 // Wake-up characteristic cycle count
    FatigueOnsetCycles float64 // Fatigue starts after this cycle count
    FatigueTauCycles   float64 // Fatigue characteristic decay
}

func WakeUpPolarization(cycles float64, cfg WakeUpModelConfig) (float64, error)
```

#### FORC Analysis (`worldclass_forc.go`)

First-order reversal curve (FORC) measurement and Preisach density extraction.

```go
type FORCResult struct {
    Emax_Vm           float64
    ReversalFields_Vm []float64
    Curves            []FORCCurve
    PreisachDensity   [][]float64  // rho(Ea,Eb) grid
    ReversalPairs     [][2]float64
}

// Run FORC sweep on a PreisachStack model.
func RunFORCSweep(model *PreisachStack, Emax float64, numReversals int) (FORCResult, error)

// Extract Preisach density via -0.5 * d²P/(dEa dEb).
func ComputeFORCDensity(results FORCResult) ([][]float64, [][2]float64)
```

#### Cycle-to-Cycle Variation (`worldclass_c2c.go`)

State-dependent C2C conductance noise (higher noise at lower G).

```go
type StateDepC2CConfig struct {
    AbsoluteNoiseSigma float64 // σ at G = G_ref
    G_ref              float64 // Reference conductance (G_max)
}

// σ(G) = AbsoluteNoiseSigma * G_ref / G  (clamped to 0.5*G)
func (c StateDepC2CConfig) C2CSigma(G float64) (float64, error)

// Draw noisy sample for nominal conductance G.
func ApplyStateDepC2C(G float64, cfg StateDepC2CConfig, rng *rand.Rand) (float64, error)

// DefaultStateDepC2CConfig returns 5% noise at G_max, calibrated for HZO FeFET.
func DefaultStateDepC2CConfig(G_max float64) StateDepC2CConfig
```

#### DCC Write (`dcc_write.go`)

Direct Cell Control: single-pulse write without ISPP feedback loop.

```go
type DCCResult struct {
    TargetPolarization_Cm float64
    ActualPolarization_Cm float64
    PulseAmplitude_Vm     float64
    PulseWidth_s          float64
    Success               bool
}

func NewDCCWriter(solver *LKSolver, mat *HZOMaterial) *DCCWriter
func (w *DCCWriter) Write(targetLevel int) (DCCResult, error)
```

### Usage Examples

#### Example 1: Material + LK solver + write control

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/physics"
)

func main() {
    mat := physics.FeCIMMaterial()
    solver := physics.NewLKSolver()
    solver.ConfigureFromMaterial(mat)

    wc := physics.NewWriteController(solver, mat)
    attempts, ok, overshoot := wc.WriteTarget(60e-6) // target 60 µS

    fmt.Printf("write ok=%v attempts=%d overshoot=%d\n", ok, attempts, overshoot)
}
```

#### Example 2: Preisach stack with custom Everett function

```go
package main

import "fecim-lattice-tools/shared/physics"

type linearEverett struct{}

func (linearEverett) Calculate(alpha, beta float64) float64 {
    if alpha <= beta {
        return 0
    }
    return (alpha - beta) / (2 * 2e8)
}

func main() {
    ps := physics.NewPreisachStack(2e8, linearEverett{})
    _ = ps.Update(-1e8)
    _ = ps.Update(0)
    _ = ps.Update(1e8)
}
```

#### Example 3: Quantization and transfer mapping

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/physics"
)

func main() {
    p := 0.18 // C/m²
    g := physics.PolarizationToConductance(p, 0.30, physics.GMin, physics.GMax)
    lvl := physics.ConductanceToLevel(g, 30)

    fmt.Printf("g=%s level=%d\n", physics.FormatConductance(g), lvl)
}
```

---

## shared/peripherals

Import path:

```go
import "fecim-lattice-tools/shared/peripherals"
```

Peripheral models for DAC, ADC, TIA, charge pump, and system-level analysis.

### Key Types

#### DAC

Write-path digital-to-analog converter with INL/DNL and energy models.

```go
type DAC struct {
    // Configuration fields
}
```

#### ADC

Read-path analog-to-digital converter with optional SAR-specific noise model.

```go
type ADC struct {
    // Configuration fields
}
```

**SAR Noise Extensions:**

```go
type SARNoiseConfig struct {
    // Thermal noise, metastability, reference drift configuration
}
```

#### TIA

Transimpedance amplifier model for current-to-voltage conversion.

```go
type TIA struct {
    // Configuration fields
}
```

#### ChargePump

Voltage booster model used for write pulses.

```go
type ChargePump struct {
    // Configuration fields
}
```

#### Additional Types

- `type SampleAndHold struct { ... }` - Front-end sample/hold model
- `type VoltageRegulator struct { ... }` - Supply regulator model
- `type ProcessCorner string` - Process corner enum
- `type INLDNLAnalysis struct { ... }` - Code-level linearity report
- `type PVTINLDNLAnalysis struct { ... }` - Linearity report under PVT condition
- `type ProcessCornerAnalysis struct { ... }` - Cross-corner linearity comparison
- `type TimingAnalysis struct { ... }` - Latency/throughput breakdown
- `type PowerBreakdown struct { ... }` - Energy and power split
- `type TransferFunction struct { ... }` - End-to-end DAC→TIA→ADC behavior

### Constructors and Defaults

```go
func DefaultDAC() *DAC
func DefaultADC() *ADC
func DefaultTIA() *TIA
func DefaultChargePump() *ChargePump
func NegativePump() *ChargePump
func DefaultSampleAndHold() *SampleAndHold
func DefaultVoltageRegulator() *VoltageRegulator
func DefaultSARNoiseConfig() *SARNoiseConfig
```

### DAC API

| Method | Description |
|--------|-------------|
| `Levels() int` | Number of output codes |
| `Convert(level int) float64` | Ideal code→voltage conversion |
| `ConvertWithNonlinearity(level int) float64` | Includes INL/DNL effects |
| `ConvertWithCondition(level int, tempK float64, corner ProcessCorner) float64` | PVT-aware conversion |
| `Resolution() float64` | LSB size |
| `VoltageRange() (min, max float64)` | DAC full-scale range |
| `EnergyPerConversion() float64` | Energy estimate per conversion |
| `AnalyzeINLDNL() *INLDNLAnalysis` | Detailed linearity report |

### ADC API

| Method | Description |
|--------|-------------|
| `Levels() int` | Number of quantization levels |
| `Convert(voltage float64) int` | Ideal voltage→code conversion |
| `ConvertWithNonlinearity(voltage float64) int` | Includes INL/DNL effects |
| `ConvertWithCondition(voltage float64, tempK float64, corner ProcessCorner) int` | PVT-aware conversion |
| `Resolution() float64` | LSB size |
| `ENOB() float64` | Effective number of bits |
| `TheoreticalSNR() float64` | Theoretical signal-to-noise ratio |
| `EffectiveSNR() float64` | Effective SNR with non-idealities |
| `EnergyPerConversion() float64` | Energy estimate per conversion |
| `AnalyzeINLDNL() *INLDNLAnalysis` | Detailed linearity report |

**SAR-specific methods:**

| Method | Description |
|--------|-------------|
| `EnableSARNoise()` | Enable SAR thermal noise model |
| `DisableSARNoise()` | Disable SAR thermal noise |
| `SetTemperature(tempK float64)` | Set operating temperature |
| `GetEffectiveVref() (vrefLow, vrefHigh float64)` | Get effective reference voltages |
| `GetThermalNoiseVoltage() float64` | Get thermal noise voltage |
| `GetMetastabilityErrorRate(inputVoltage, thresholdVoltage float64) float64` | Metastability error rate |
| `ConvertWithSARNoise(voltage float64, seed int64) int` | Convert with noise injection |
| `GetSARNoiseReport() map[string]float64` | Detailed noise report |

### TIA / ChargePump / Other Peripheral APIs

| Method | Description |
|--------|-------------|
| `(t *TIA) Convert(current float64) float64` | Current→voltage conversion |
| `(t *TIA) ConvertWithNoise(current float64) float64` | With noise added |
| `(t *TIA) SNR(current float64) float64` | Signal-to-noise ratio |
| `(t *TIA) SettlingTime() float64` | Settling time constant |
| `(t *TIA) PowerConsumption() float64` | Power draw estimate |
| `(c *ChargePump) IdealOutputVoltage() float64` | Ideal boost voltage |
| `(c *ChargePump) ActualOutputVoltage() float64` | With ripple effects |
| `(c *ChargePump) OutputRipple() float64` | Output voltage ripple |
| `(c *ChargePump) EnergyPerOperation(pulseDuration float64) float64` | Energy per pulse |
| `(c *ChargePump) MaxCurrentCapability() float64` | Max current capability |
| `(s *SampleAndHold) SettledFraction(tSeconds float64) float64` | Settling progress |
| `(s *SampleAndHold) HoldDroop(tSeconds float64) float64` | Voltage droop during hold |
| `(r *VoltageRegulator) Regulate(vin, loadCurrent float64) float64` | Regulated output |
| `(r *VoltageRegulator) SupplyNoiseTransfer(vinRipple float64) float64` | Ripple transfer |

### System-Level Analysis APIs

```go
func EffectiveINLDNL(inl, dnl, tempK float64, corner ProcessCorner) (float64, float64)
func AnalyzeINLDNLAtCondition(dac *DAC, adc *ADC, temperatureK float64, corner ProcessCorner) *PVTINLDNLAnalysis
func AnalyzeProcessCorners(dac *DAC, adc *ADC, temperatureK float64) *ProcessCornerAnalysis
func AnalyzeTiming(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TimingAnalysis
func AnalyzePower(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump, timing *TimingAnalysis) *PowerBreakdown
func ComputeTransferFunction(dac *DAC, adc *ADC, tia *TIA, pump *ChargePump) *TransferFunction
func BuildBehavioralSpiceSubcircuits(dac *DAC, adc *ADC, tia *TIA, sh *SampleAndHold, vr *VoltageRegulator) string
```

### Usage Examples

#### Example 1: DAC + ADC round-trip

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    dac := peripherals.DefaultDAC()
    adc := peripherals.DefaultADC()

    v := dac.Convert(12)
    code := adc.Convert(v)
    fmt.Printf("level=12 -> %.4f V -> code=%d\n", v, code)
}
```

#### Example 2: PVT and timing/power analysis

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    dac := peripherals.DefaultDAC()
    adc := peripherals.DefaultADC()
    tia := peripherals.DefaultTIA()
    pump := peripherals.DefaultChargePump()

    timing := peripherals.AnalyzeTiming(dac, adc, tia, pump)
    power := peripherals.AnalyzePower(dac, adc, tia, pump, timing)
    fmt.Printf("cycle(ns)=%.2f total_energy(J)=%.3e\n", timing.CycleTime, power.TotalEnergy)
}
```

#### Example 3: Generate behavioral SPICE stubs

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/shared/peripherals"
)

func main() {
    netlist := peripherals.BuildBehavioralSpiceSubcircuits(
        peripherals.DefaultDAC(),
        peripherals.DefaultADC(),
        peripherals.DefaultTIA(),
        peripherals.DefaultSampleAndHold(),
        peripherals.DefaultVoltageRegulator(),
    )
    fmt.Println(netlist)
}
```

---

## shared/io

Import path:

```go
import "fecim-lattice-tools/shared/io"
```

File I/O utilities for JSON and directory operations.

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

## shared/widgets

Import path:

```go
import "fecim-lattice-tools/shared/widgets"
```

Reusable Fyne GUI components for FeCIM visualizers.

### Layout Components

#### AdaptiveLayout

Provides responsive layout that adapts to screen size.

```go
type AdaptiveLayout struct {
    // Maintains desktop (splits) and mobile (tabs) layouts
}

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

#### UI Helpers (Thread-Safe)

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

## module1-hysteresis/pkg/ferroelectric

Import path:

```go
import "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
```

Hysteresis and Preisach modeling for ferroelectric materials.

### Key Types

```go
type HZOMaterial = sharedphysics.HZOMaterial
    // Backward-compatible alias of shared material model

type PreisachModel struct { ... }
    // High-level hysteresis model wrapper around shared Preisach engine

type TanhEverett struct {
    Ps    float64 // Saturation polarization
    Ec    float64 // Coercive field
    Delta float64 // Transition width
}
    // Everett kernel implementation using tanh major-loop approximation

type DiscreteState struct { ... }
    // One programmable state with level, polarization, voltage, conductance

type LevelBins struct { ... }
    // Quantization bin helper with guard-band awareness

type PERenderer struct {
    Width  int
    Height int
    Color  bool
}
    // ASCII renderer for loops/domain/switching plots
```

### Material Presets (Re-exported)

```go
func DefaultHZO() *HZOMaterial
func FeCIMMaterial() *HZOMaterial
func FeCIMMaterialTarget() *HZOMaterial
func LiteratureSuperlattice() *HZOMaterial
func CryogenicHZO() *HZOMaterial
func HZOStandard32() *HZOMaterial
func HZOFJT140() *HZOMaterial
func HZOCustom14() *HZOMaterial
func AlScN() *HZOMaterial
func AllMaterials() []*HZOMaterial
```

### Preisach Model API

| Method | Description |
|--------|-------------|
| `NewPreisachModel(material *HZOMaterial) *PreisachModel` | Creates hysteresis model |
| `Update(E float64) float64` | Applies field step, returns polarization |
| `Polarization() float64` | Current polarization |
| `NormalizedPolarization() float64` | Polarization normalized by saturation |
| `Reset()` | Reset to negative saturation state |
| `GetHysteresisLoop(Emax float64, points int) ([]float64, []float64)` | Generates major loop |
| `SetTemperature(tempK float64)` | Applies thermal parameter scaling |
| `SetStress(stressGPa float64)` | Applies mechanical stress coupling |
| `GetEffectiveEc() float64` | Returns current effective coercive field |
| `DiscreteStates(n int) []DiscreteState` | Samples n evenly spaced states |

### Binning and Rendering API

| Function | Description |
|----------|-------------|
| `NewLevelBins(ps float64, numLevels int, rangeFrac float64, guardFrac float64) LevelBins` | Create bin helper |
| `(b LevelBins) EffectivePs() float64` | Effective saturation polarization |
| `(b LevelBins) Step() float64` | Bin step size |
| `(b LevelBins) LevelForP(P float64) (level int, inError bool, delta float64)` | Map P to level |
| `NewPERenderer() *PERenderer` | Create renderer |
| `(r *PERenderer) RenderPELoop(E, P []float64, material *HZOMaterial) string` | Render P-E loop |
| `(r *PERenderer) RenderDomainStates(alphas, betas []float64, states []int) string` | Render domain states |
| `(r *PERenderer) RenderDiscreteStates(states []DiscreteState) string` | Render discrete states |
| `(r *PERenderer) RenderSwitchingDynamics(times, pols []float64, switched []int, material *HZOMaterial) string` | Render switching |
| `(r *PERenderer) RenderTemperatureDependence(material *HZOMaterial) string` | Render temperature effects |
| `(r *PERenderer) RenderMaterialComparison() string` | Render material comparison |

### Usage Examples

#### Example 1: Generate and inspect a P-E loop

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
    m := ferroelectric.NewPreisachModel(ferroelectric.FeCIMMaterial())
    E, P := m.GetHysteresisLoop(2e8, 200)
    fmt.Printf("points=%d firstP=%g lastP=%g\n", len(E), P[0], P[len(P)-1])
}
```

#### Example 2: Quantize polarization into guarded bins

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
    bins := ferroelectric.NewLevelBins(0.30, 30, 0.90, 0.20)
    level, inError, delta := bins.LevelForP(0.12)
    fmt.Printf("level=%d guard=%v delta=%g\n", level, inError, delta)
}
```

#### Example 3: Render loop as ASCII

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

func main() {
    model := ferroelectric.NewPreisachModel(ferroelectric.DefaultHZO())
    E, P := model.GetHysteresisLoop(2e8, 120)
    plot := ferroelectric.NewPERenderer().RenderPELoop(E, P, ferroelectric.DefaultHZO())
    fmt.Println(plot)
}
```

---

## module2-crossbar/pkg/crossbar

Import path:

```go
import "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
```

Crossbar array simulation with non-ideality models and parasitic solvers.

### Key Types

#### Array Management

```go
type Config struct { ... }
    // Top-level array config (dimensions, quantization, non-idealities)

type Array struct { ... }
    // Main crossbar array model (programming, MVM/VMM, statistics, exports)

type Cell struct { ... }
    // Per-cell state and metadata

type CellStats struct { ... }
    // Per-cell derived write/disturb stats
```

#### Configuration and Effects

```go
type EnduranceConfig struct { ... }
    // Endurance-fatigue knobs

type ProcessVariationConfig struct { ... }
    // Process-gradient/edge variation controls

type HalfSelectConfig struct { ... }
    // Half-select disturb coupling config

type MVMOptions struct { ... }
    // Advanced MVM pipeline options (IR drop, sneak, temperature profile)

type MVMResult struct { ... }
    // Rich MVM output incl. error and energy metrics
```

#### Analysis and Solvers

```go
type AnalysisReport struct { ... }
    // High-level report snapshot for a run

type AccuracyDegradation struct { ... }
type DegradationStep struct { ... }
    // Accuracy drop progression under cumulative non-idealities

type SORConfig struct { ... }
    // Iterative parasitic solver settings

type ParasiticSolver struct { ... }
    // SOR-based IR-drop/parasitic MVM solver

type OptimizedParasiticSolver struct { ... }
    // Allocation-optimized SOR solver variant

type ParasiticMVMResult struct { ... }
    // Detailed solve result (iterations, convergence, effective outputs)

type IRDropSimulator struct { ... }
    // Time-domain IR drop simulator

type SneakPathAnalyzer struct { ... }
    // Dedicated sneak-path analysis engine

type DriftSimulator struct { ... }
    // Conductance drift evolution simulator

type DeviceErrorEngine struct { ... }
    // Programming/read error injector

type WriteDisturbEngine struct { ... }
    // Half-select stress/disturb tracker
```

#### GPU Acceleration

```go
type GPUAccelerator struct { ... }
    // Optional GPU MVM backend
```

#### Thermal and Reliability

```go
type TemperatureEffects struct { ... }
type ThermalPhysicsModel struct { ... }
    // Thermal scaling and reliability modeling utilities
```

### Array Lifecycle and Core Operations

| Method | Description |
|--------|-------------|
| `NewArray(cfg *Config) (*Array, error)` | Create array |
| `(a *Array) Destroy()` | Cleanup resources |
| `(a *Array) Rows() int` | Get row count |
| `(a *Array) Cols() int` | Get column count |
| `(a *Array) GetConfig() Config` | Get configuration |
| `(a *Array) GetStats() (reads, writes int64)` | Get operation counts |

### Programming Methods

| Method | Description |
|--------|-------------|
| `(a *Array) ProgramWeight(row, col int, weight float64) error` | Program single cell |
| `(a *Array) ProgramWeightMatrix(weights [][]float64) error` | Program full matrix |
| `(a *Array) ProgramWeightWithDisturb(row, col int, weight float64, isPassive bool) error` | With disturb effects |
| `(a *Array) ProgramWeightWithVariation(row, col int, targetLevel int, stats *WriteStatistics) (int, error)` | With device variation |

### Compute Methods

| Method | Description |
|--------|-------------|
| `(a *Array) MVM(input []float64) ([]float64, error)` | Ideal matrix-vector multiply |
| `(a *Array) VMM(input []float64) ([]float64, error)` | Vector-matrix multiply |
| `(a *Array) MVMWithNonIdealities(input []float64, opts *MVMOptions) (*MVMResult, error)` | Non-ideal MVM |

### Conductance Access

| Method | Description |
|--------|-------------|
| `(a *Array) GetConductanceMatrix() [][]float64` | Get programmed conductances |
| `(a *Array) GetEffectiveConductanceMatrix() [][]float64` | Get effective conductances |
| `(a *Array) GetPhysicalConductance(gNorm float64) float64` | Normalize to physical |
| `(a *Array) GetPhysicalConductanceForCell(row, col int) float64` | Get physical conductance |

### Quantization and Compatibility

| Function | Description |
|----------|-------------|
| `QuantizeToLevels(value float64) float64` | 30-level normalized quantizer |
| `GetLevel(conductance float64) int` | 30-level index conversion |

### Non-Ideal Analysis and Reporting

| Method | Description |
|--------|-------------|
| `(a *Array) AnalyzeRCDelay(params *WireParams, inputVoltage float64) *RCDelayAnalysis` | RC delay analysis |
| `(a *Array) AnalyzeIRDrop(input []float64, params *WireParams) *IRDropAnalysis` | IR drop analysis |
| `(a *Array) AnalyzeIRDropIterative(input []float64, params *WireParams, config *IRDropSolverConfig) *IRDropAnalysis` | Iterative IR drop |
| `(a *Array) AnalyzeSneakPaths(selectedRow, selectedCol int) *SneakPathAnalysis` | Sneak path analysis |
| `(a *Array) AnalyzeSneakPathsWithArch(selectedRow, selectedCol int, is1T1R bool) *SneakPathAnalysis` | Architecture-aware sneak |
| `(a *Array) GenerateMVMSneakTrace(input []float64, opts *MVMOptions, maxPaths int) *MVMSneakTraceReport` | MVM sneak trace |
| `(a *Array) AnalyzeSneakContributions(targetRow int, input []float64, maxPaths int) []SneakPathContribution` | Sneak contributions |

### Reporting and Export

| Method | Description |
|--------|-------------|
| `(a *Array) GenerateReport(mvmResult *MVMResult) *AnalysisReport` | Generate analysis report |
| `(a *Array) ExportWeightsCSV(path string) error` | Export to CSV |
| `(a *Array) ExportAnalysisJSON(path string, mvmResult *MVMResult) error` | Export to JSON |
| `(a *Array) ComputeAccuracyDegradation(input []float64, baselineAccuracy float64) (*AccuracyDegradation, error)` | Compute degradation |
| `(a *Array) ComputeAccuracyDegradationWithOptions(input []float64, baselineAccuracy float64, opts *MVMOptions) (*AccuracyDegradation, error)` | With options |

### Parasitic Solver API

| Function/Method | Description |
|-----------------|-------------|
| `DefaultSORConfig() *SORConfig` | Get default SOR config |
| `NewParasiticSolver(rows, cols int, config *SORConfig) (*ParasiticSolver, error)` | Create solver |
| `(s *ParasiticSolver) SetConductances(g [][]float64)` | Set conductance matrix |
| `(s *ParasiticSolver) SetParasitics(rpRow, rpCol float64)` | Set parasitic resistances |
| `(s *ParasiticSolver) SolveMVM(appliedVoltages []float64) (*ParasiticMVMResult, error)` | Solve with parasitics |
| `(s *ParasiticSolver) SolveMVMWithFallback(appliedVoltages []float64) (*ParasiticMVMResult, error)` | With fallback |
| `(s *ParasiticSolver) ComputeIdealMVM(appliedVoltages []float64) []float64` | Ideal MVM |
| `(s *ParasiticSolver) AnalyzeParasiticImpact(appliedVoltages []float64) (*ParasiticImpact, error)` | Analyze impact |
| `NewOptimizedParasiticSolver(rows, cols int, config *SORConfig) (*OptimizedParasiticSolver, error)` | Optimized solver |
| `(s *OptimizedParasiticSolver) SolveMVM(appliedVoltages []float64) (*ParasiticMVMResult, error)` | Optimized solve |
| `(s *OptimizedParasiticSolver) SolveMVMFast(appliedVoltages []float64) ([]float64, int, error)` | Fast solve |

### IR Drop / Sneak Path / Drift / Error Engines

| Function/Method | Description |
|-----------------|-------------|
| `NewIRDropSimulator(rows, cols int) *IRDropSimulator` | Create IR drop simulator |
| `(ir *IRDropSimulator) SetConductance(row, col int, g float64)` | Set cell conductance |
| `(ir *IRDropSimulator) SetAllInputs(voltages []float64)` | Set input voltages |
| `(ir *IRDropSimulator) Simulate(iterations int)` | Run simulation |
| `(ir *IRDropSimulator) GetOutputCurrents() []float64` | Get outputs |
| `(ir *IRDropSimulator) GetStats() IRDropStats` | Get statistics |
| `NewSneakPathAnalyzer(rows, cols int) *SneakPathAnalyzer` | Create analyzer |
| `(sp *SneakPathAnalyzer) AnalyzeTarget(targetRow, targetCol int, voltage float64)` | Analyze target |
| `(sp *SneakPathAnalyzer) GetStats(voltage float64) SneakPathStats` | Get stats |
| `NewDriftSimulator(rows, cols int, levels int) *DriftSimulator` | Create drift simulator |
| `NewDriftSimulatorWithModel(rows, cols int, levels int, model DriftModel) *DriftSimulator` | With custom model |
| `(d *DriftSimulator) SimulateTimeStep(dt float64)` | Simulate time step |
| `(d *DriftSimulator) GetStats() DriftStats` | Get statistics |
| `CompareTechnologies(rows, cols int, simulationTime float64) map[string]DriftStats` | Compare techs |
| `NewDeviceErrorEngine(progConfig *ProgrammingErrorConfig, readConfig *ReadNoiseConfig) *DeviceErrorEngine` | Create error engine |
| `(e *DeviceErrorEngine) ApplyProgrammingError(gTarget float64) float64` | Apply prog error |
| `(e *DeviceErrorEngine) ApplyReadNoise(gProgrammed float64, row, col int) float64` | Apply read noise |
| `ComputeErrorStatistics(target, actual [][]float64) *ErrorStatistics` | Compute stats |
| `NewWriteDisturbEngine(rows, cols int, config *WriteDisturbConfig) *WriteDisturbEngine` | Create disturb engine |
| `(e *WriteDisturbEngine) RecordWrite(targetRow, targetCol int)` | Record write |
| `(e *WriteDisturbEngine) ApplyDisturbEffects(conductances [][]float64, levels int) int` | Apply effects |
| `(e *WriteDisturbEngine) GetStressStats() WriteDisturbStats` | Get stress stats |

### Standalone Utility Functions

```go
func ComputeError(ideal, actual []float64) float64
func ComputeAccuracyLoss(ideal, actual []float64) float64
func SimulateAccuracyDegradation(progSigma, readSigma float64, arraySize int) float64
func RecommendErrorBudget(targetAccuracy float64, arraySize int) (progSigma, readSigma float64)
func EstimateDisturbRate(writesPerCell float64, config *WriteDisturbConfig) float64
func CompareArchitectures(writesPerCell float64) (passiveRate, activeRate float64)
func HalfSelectVoltage(writeVoltage float64, scheme string) float64
func IsDisturbCritical(halfSelectV, coerciveV, safetyMargin float64) bool
func NewGPUAccelerator(maxRows, maxCols int) (*GPUAccelerator, error)
```

### Usage Examples

#### Example 1: Program matrix and run MVM

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
    cfg := &crossbar.Config{Rows: 2, Cols: 3}
    arr, err := crossbar.NewArray(cfg)
    if err != nil {
        panic(err)
    }
    defer arr.Destroy()

    _ = arr.ProgramWeightMatrix([][]float64{
        {0.2, 0.7, 0.1},
        {0.9, 0.4, 0.6},
    })

    y, _ := arr.MVM([]float64{1.0, 0.5})
    fmt.Println(y)
}
```

#### Example 2: Non-ideality-aware MVM

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
    arr, _ := crossbar.NewArray(&crossbar.Config{Rows: 4, Cols: 4})
    defer arr.Destroy()

    opts := crossbar.DefaultMVMOptions()
    res, _ := arr.MVMWithNonIdealities([]float64{1, 0, 1, 0}, opts)
    fmt.Printf("rmse=%.4f energy=%.3e\n", res.RMSE, res.EnergyJ)
}
```

#### Example 3: Parasitic solver use

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
    s := crossbar.DefaultSORConfig()
    solver, _ := crossbar.NewParasiticSolver(8, 8, s)
    solver.SetParasitics(2.5, 2.5)

    result, err := solver.SolveMVM(make([]float64, 8))
    if err != nil {
        panic(err)
    }
    fmt.Printf("iters=%d converged=%v\n", result.Iterations, result.Converged)
}
```

---

## module3-mnist/pkg/core

Import path:

```go
import "fecim-lattice-tools/module3-mnist/pkg/core"
```

MNIST neural network simulation with dual floating-point and CIM modes.

### Key Types

#### Network Configuration

```go
type NetworkConfig struct { ... }
    // Runtime configuration for quantization, noise, ADC/DAC bits, mode options

type DualModeNetwork struct { ... }
    // Main MNIST inference model exposing FP and CIM paths

type InferenceResult struct { ... }
    // Per-sample FP vs CIM outputs, predictions, agreement, energy metrics
```

#### Data and Serialization

```go
type WeightsFile struct { ... }
    // JSON schema for stored weights (+ optional quant metadata)
```

#### Energy and Noise Modeling

```go
type EnergyEstimate struct { ... }
    // Detailed energy accounting for one inference

type QuantizationStats struct { ... }
    // Statistical quality metrics comparing original vs quantized weights

type CIMNoiseComponents struct { ... }
    // Decomposed physical noise components for CIM path

type TIAModel struct { ... }
    // Lightweight TIA transfer/bandwidth helper used by CIM physics path
```

#### Metrics and Analysis

```go
type ModeMetrics struct { ... }
    // Confusion-matrix + aggregate metrics for one inference mode

type DualModeDatasetMetrics struct { ... }
    // Dataset-level FP/CIM metrics and agreement summary

type RandomSource struct { ... }
    // Reproducible RNG wrapper for quantization/noise
```

#### Interfaces

```go
type Inferer interface { ... }
type WeightLoader interface { ... }
type WeightProvider interface { ... }
type NetworkConfigurer interface { ... }
type DataLoader interface { ... }
type Network interface { ... }
```

### Network Creation and Configuration

| Method | Description |
|--------|-------------|
| `DefaultNetworkConfig() *NetworkConfig` | Get default config |
| `NewDualModeNetwork(inputSize, hiddenSize, outputSize int) *DualModeNetwork` | Create network |
| `(net *DualModeNetwork) SetNumLevels(levels int)` | Set quantization levels |
| `(net *DualModeNetwork) GetNumLevels() int` | Get quantization levels |
| `(net *DualModeNetwork) SetPerLayerQuant(enabled bool)` | Enable per-layer quant |
| `(net *DualModeNetwork) IsPerLayerQuant() bool` | Check per-layer quant |
| `(net *DualModeNetwork) SetPerLayerLevels(layer1, layer2 int)` | Set per-layer levels |
| `(net *DualModeNetwork) GetPerLayerQuantInfo() (enabled bool, l1Levels, l2Levels int)` | Get per-layer info |
| `(net *DualModeNetwork) SetNoiseLevel(noise float64)` | Set noise level |
| `(net *DualModeNetwork) SetADCBits(bits int)` | Set ADC bits |
| `(net *DualModeNetwork) SetDACBits(bits int)` | Set DAC bits |
| `(net *DualModeNetwork) SetSingleLayer(enabled bool)` | Set single-layer mode |
| `(net *DualModeNetwork) IsSingleLayer() bool` | Check single-layer mode |

### Inference API

| Method | Description |
|--------|-------------|
| `(net *DualModeNetwork) Infer(input []float64) *InferenceResult` | Run both FP and CIM |
| `(net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64)` | FP inference only |
| `(net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64)` | CIM inference only |
| `EvaluateDualModeDataset(net *DualModeNetwork, images [][]float64, labels []int) DualModeDatasetMetrics` | Evaluate dataset |

### Weights and Quantization

| Method | Description |
|--------|-------------|
| `(net *DualModeNetwork) LoadWeights(filename string) error` | Load weights |
| `(net *DualModeNetwork) LoadWeightsForLevel(dataDir string, levels int) error` | Load for level |
| `(net *DualModeNetwork) RequantizeWeights()` | Requantize weights |
| `ScanAvailableQATLevels(dataDir string) []int` | Scan available levels |
| `GetWeightsFilename(dataDir string, levels int) string` | Get filename |
| `GetBestMatchingWeightsLevel(dataDir string, targetLevels int) int` | Find best match |
| `QuantizeWeights(fpWeights [][]float64, levels int) ([][]float64, error)` | Quantize weights |
| `QuantizeBias(fpBias []float64, levels int) ([]float64, error)` | Quantize bias |
| `ComputeQuantizationStats(original, quantized [][]float64) QuantizationStats` | Compute stats |
| `(net *DualModeNetwork) GetQuantizationStats() (layer1Stats, layer2Stats QuantizationStats)` | Get stats |
| `(net *DualModeNetwork) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64)` | Get FP weights |
| `(net *DualModeNetwork) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64)` | Get quantized weights |

### Noise and Energy Modeling

| Method | Description |
|--------|-------------|
| `AddGaussianNoise(values []float64, noiseLevel float64, rng *RandomSource) []float64` | Add Gaussian noise |
| `AddGaussianNoiseInPlace(values []float64, noiseLevel float64, rng *RandomSource)` | Add noise in-place |
| `NewRandomSource(seed uint64) *RandomSource` | Create RNG |
| `(n CIMNoiseComponents) TotalSigma() float64` | Get total noise sigma |
| `EnergyPerMACJ(levels int) float64` | Energy per MAC |
| `EstimateInferenceEnergyJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) EnergyEstimate` | Estimate energy |
| `EstimateInferenceEnergyMicroJ(cfg *NetworkConfig, inputSize, hiddenSize, outputSize int) float64` | Energy in microJ |

### GPU and Notifications

| Method | Description |
|--------|-------------|
| `InitGPU()` | Initialize GPU |
| `IsGPUAvailable() bool` | Check GPU availability |
| `DestroyGPU()` | Cleanup GPU |
| `(net *DualModeNetwork) SetUseGPU(use bool)` | Enable/disable GPU |
| `(net *DualModeNetwork) UseGPU() bool` | Check GPU usage |
| `(net *DualModeNetwork) SetNotificationHandler(handler func(message string))` | Set notifications |

### Usage Examples

#### Example 1: Run dual-mode inference

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
    net := core.NewDualModeNetwork(784, 128, 10)
    net.SetNumLevels(30)
    net.SetNoiseLevel(0.03)

    input := make([]float64, 784) // replace with normalized MNIST sample
    res := net.Infer(input)
    if res == nil {
        panic("invalid input")
    }
    fmt.Printf("fp=%d cim=%d agree=%v\n", res.FPPrediction, res.CIMPrediction, res.Agreement)
}
```

#### Example 2: Load level-matched weights and evaluate energy

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
    cfg := core.DefaultNetworkConfig()
    cfg.NumLevels = 30

    uJ := core.EstimateInferenceEnergyMicroJ(cfg, 784, 128, 10)
    fmt.Printf("estimated energy = %.4f µJ\n", uJ)
}
```

#### Example 3: Quantize custom weight tensor

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module3-mnist/pkg/core"
)

func main() {
    w := [][]float64{{-0.5, 0.0, 0.8}, {0.3, -0.2, 0.1}}
    q, err := core.QuantizeWeights(w, 30)
    if err != nil {
        panic(err)
    }
    stats := core.ComputeQuantizationStats(w, q)
    fmt.Printf("mse=%g max_abs=%g\n", stats.MSE, stats.MaxAbsError)
}
```

---

## Complete Example

Here's a complete example combining multiple packages:

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

## Notes and Conventions

- This document intentionally focuses on **public, high-value API surfaces** and omits internal/private helpers.
- All snippets are minimal and intended as starting points; production code should include validation, error handling, and deterministic seeds where applicable.
- For package internals and implementation details, inspect source files directly in the corresponding package directories.
- Widget APIs use Fyne 2.x primitives (`fyne.CanvasObject`, `fyne.App`, etc.)
- Physics APIs follow physics conventions: fields in V/m, polarization in C/m², conductance in Siemens, time in seconds.

---

## See Also

- [README.md](../../README.md) - Project overview
- [CONTRIBUTING.md](../../CONTRIBUTING.md) - Contribution guidelines
- [status.md](../../status.md) - Project status
- [docs/4-research/honesty-audit.md](../4-research/honesty-audit.md) - Accuracy and honesty audit
