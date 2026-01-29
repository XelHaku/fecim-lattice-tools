# Crossbar Array API Reference

The `module2-crossbar/pkg/crossbar` package provides a complete simulation suite for ferroelectric crossbar arrays with integrated non-ideality analysis. This document details all exported functions, types, and their usage.

**Quick Links:**
- [Array Creation](#array-creation)
- [Quantization](#quantization)
- [Core Operations](#core-operations)
- [Weight Programming](#weight-programming)
- [Non-Ideality Analysis](#non-ideality-analysis)
- [Drift Simulation](#drift-simulation)
- [Temperature Effects](#temperature-effects)
- [Types Reference](#types-reference)
- [Utility Functions](#utility-functions)
- [Common Patterns](#common-patterns)

---

## Array Creation

### NewArray

Creates a new crossbar array with the specified configuration.

**Signature:**
```go
func NewArray(cfg *Config) (*Array, error)
```

**Parameters:**
- `cfg`: Configuration structure specifying array dimensions, quantization, conductance model, and non-ideality settings

**Returns:**
- `*Array`: Initialized array with all cells set to zero conductance
- `error`: Non-nil if configuration is invalid (rows or cols <= 0)

**Example:**
```go
cfg := &Config{
    Rows:       64,
    Cols:       64,
    NoiseLevel: 0.02,
    ADCBits:    8,
    DACBits:    8,
    ConductanceModel: ConductanceExponential,
}

array, err := NewArray(cfg)
if err != nil {
    log.Fatal(err)
}
```

### DefaultEnduranceConfig

Returns default endurance (fatigue) settings based on literature: 10^8 cycles fatigue threshold, 10^12 cycles failure threshold (IEEE IRPS 2022, Nano Letters 2024).

**Signature:**
```go
func DefaultEnduranceConfig() *EnduranceConfig
```

**Returns:**
- `*EnduranceConfig`: Default settings with endurance modeling disabled by default

**Example:**
```go
enduranceCfg := DefaultEnduranceConfig()
enduranceCfg.Enabled = true  // Enable fatigue modeling

cfg := &Config{
    Rows: 64,
    Cols: 64,
    Endurance: enduranceCfg,
}
```

### DefaultProcessVariationConfig

Returns default process variation settings: 2% device-to-device sigma, 0.1% gradient per cell, 5% edge effect.

**Signature:**
```go
func DefaultProcessVariationConfig() *ProcessVariationConfig
```

**Returns:**
- `*ProcessVariationConfig`: Default systematic variation parameters

### DefaultHalfSelectConfig

Returns default half-select disturb settings for passive (0T1R) crossbars.

**Signature:**
```go
func DefaultHalfSelectConfig() *HalfSelectConfig
```

**Returns:**
- `*HalfSelectConfig`: Default settings with V/2 threshold and disturb rate

---

## Quantization

The crossbar uses 30 discrete analog levels (representing 4.9 bits per cell) to model ferroelectric switching. All weight values are automatically quantized to these levels.

### QuantizeToLevels

Quantizes a floating-point value to one of 30 discrete levels (0-29).

**Signature:**
```go
func QuantizeToLevels(value float64) float64
```

**Parameters:**
- `value`: Input value (any range; automatically clamped to [0, 1])

**Returns:**
- Quantized value as one of: 0.0, 1/29, 2/29, ..., 28/29, 1.0

**Example:**
```go
// 0.5 maps to level 15 (middle)
quantized := QuantizeToLevels(0.5)  // Returns ~0.5172 (15/29)

// Extreme values are clamped
quantized := QuantizeToLevels(1.5)  // Returns 1.0 (level 29)
quantized := QuantizeToLevels(-0.5) // Returns 0.0 (level 0)
```

### GetLevel

Converts a normalized conductance value to its discrete level number (0-29).

**Signature:**
```go
func GetLevel(conductance float64) int
```

**Parameters:**
- `conductance`: Normalized conductance [0, 1]

**Returns:**
- Integer level 0-29

**Example:**
```go
level := GetLevel(0.5)  // Returns 15
level := GetLevel(0.0)  // Returns 0 (OFF state)
level := GetLevel(1.0)  // Returns 29 (ON state)
```

---

## Core Operations

### MVM (Matrix-Vector Multiplication)

Performs y = W × x multiplication on the crossbar, applying Ohm's law at each cell: I = G × V.

**Signature:**
```go
func (a *Array) MVM(input []float64) ([]float64, error)
```

**Parameters:**
- `input`: Input vector applied to columns/bit lines (length ≤ array.Cols)

**Returns:**
- `[]float64`: Output vector read from rows/word lines
- `error`: Non-nil if input size exceeds array columns

**Physics:**
- Each cell contributes current: I = G_ij × V_j (Ohm's law)
- Row current summed via Kirchhoff's current law
- Output normalized by theoretical maximum (all weights and inputs = 1.0)
- Both input (DAC) and output (ADC) quantized based on configuration

**Example:**
```go
input := []float64{0.5, 0.3, 0.7, 0.2}
output, err := array.MVM(input)
if err != nil {
    log.Fatal(err)
}
// output contains row-wise accumulation of W*input
```

### MVMWithNonIdealities

Performs MVM with detailed non-ideality effects: IR drop, sneak paths, process variation, and temperature effects.

**Signature:**
```go
func (a *Array) MVMWithNonIdealities(input []float64, opts *MVMOptions) (*MVMResult, error)
```

**Parameters:**
- `input`: Input vector
- `opts`: Options specifying which non-idealities to include (nil uses defaults)

**Returns:**
- `*MVMResult`: Detailed results including ideal/actual outputs, error metrics, and energy estimates
- `error`: Non-nil if input exceeds array columns

**Included Analysis:**
- IR drop voltage distribution (temperature-compensated)
- Sneak path current calculation
- Process variation factors (device, spatial gradient, edge effects)
- Temperature effects on conductance range and wire resistance
- Error metrics: RMSE, max error, mean error
- Energy estimates: array, ADC, DAC, total (pJ)
- GPU energy equivalence factor

**Example:**
```go
opts := &MVMOptions{
    EnableIRDrop:     true,
    EnableSneakPaths: true,
    EnableVariation:  true,
    Temperature:      300.0,
    Architecture:     "0T1R",
}

result, err := array.MVMWithNonIdealities(input, opts)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Ideal output:  %v\n", result.IdealOutput)
fmt.Printf("Actual output: %v\n", result.ActualOutput)
fmt.Printf("RMSE error:    %.4f\n", result.RMSE)
fmt.Printf("Total energy:  %.2f pJ\n", result.TotalEnergy)
```

### VMM (Vector-Matrix Multiplication)

Performs y = x × W multiplication (transpose of MVM).

**Signature:**
```go
func (a *Array) VMM(input []float64) ([]float64, error)
```

**Parameters:**
- `input`: Input vector applied to rows/word lines

**Returns:**
- `[]float64`: Output vector read from columns/bit lines
- `error`: Non-nil if input size exceeds array rows

**Example:**
```go
input := []float64{0.2, 0.4, 0.6}
output, err := array.VMM(input)
// output has length array.Cols
```

---

## Weight Programming

### ProgramWeight

Programs a single weight value to a cell. The value is automatically quantized to one of 30 levels.

**Signature:**
```go
func (a *Array) ProgramWeight(row, col int, weight float64) error
```

**Parameters:**
- `row`: Row index [0, array.Rows)
- `col`: Column index [0, array.Cols)
- `weight`: Weight value (automatically clamped to [0, 1])

**Returns:**
- `error`: Non-nil if indices are out of range

**Side Effects:**
- Increments cell switching count (endurance tracking)
- Updates array total write count

**Example:**
```go
if err := array.ProgramWeight(5, 10, 0.75); err != nil {
    log.Fatal(err)
}
```

### ProgramWeightMatrix

Programs an entire weight matrix to the array at once.

**Signature:**
```go
func (a *Array) ProgramWeightMatrix(weights [][]float64) error
```

**Parameters:**
- `weights`: Matrix with dimensions ≤ array dimensions

**Returns:**
- `error`: Non-nil if matrix exceeds array dimensions

**Example:**
```go
weights := [][]float64{
    {0.1, 0.2, 0.3},
    {0.4, 0.5, 0.6},
    {0.7, 0.8, 0.9},
}

if err := array.ProgramWeightMatrix(weights); err != nil {
    log.Fatal(err)
}
```

### ProgramWeightWithDisturb

Programs a weight with half-select disturb tracking (for passive 0T1R crossbars).

**Signature:**
```go
func (a *Array) ProgramWeightWithDisturb(row, col int, weight float64, isPassive bool) error
```

**Parameters:**
- `row`, `col`: Target cell indices
- `weight`: Weight value to program
- `isPassive`: If true, applies V/2 disturb to cells sharing the row/column

**Returns:**
- `error`: Non-nil if indices out of range

**Disturb Model:**
- Same-row cells: Experience V/2 on word line
- Same-column cells: Experience V/2 on bit line
- Disturb applies if configuration HalfSelect is enabled
- Disturb accumulates conductance shift over multiple programming cycles

**Example:**
```go
// Enable half-select disturb in config
cfg.HalfSelect = DefaultHalfSelectConfig()
cfg.HalfSelect.Enabled = true

array, _ := NewArray(cfg)

// Program with disturb tracking
if err := array.ProgramWeightWithDisturb(5, 10, 0.75, true); err != nil {
    log.Fatal(err)
}
```

---

## Non-Ideality Analysis

Non-ideality analysis functions examine specific effects in isolation. For integrated analysis, use `MVMWithNonIdealities()`.

### AnalyzeIRDrop

Analyzes voltage drop along metal interconnects using an iterative relaxation model.

**Signature:**
```go
func (a *Array) AnalyzeIRDrop(input []float64, params *WireParams) *IRDropAnalysis
```

**Parameters:**
- `input`: Input voltage vector
- `params`: Wire resistance parameters (use `DefaultWireParams()` if nil)

**Returns:**
- `*IRDropAnalysis`: Voltage distribution and drop statistics

**Model:**
- Assumes row drivers on left, column sense amps/ground at top
- Word line voltage drops with column distance (due to word line resistance)
- Bit line voltage rises from ground with row distance (due to bit line resistance)
- Effective cell voltage = word line voltage - bit line voltage

**Example:**
```go
params := DefaultWireParams()
params.RwordLine = 3.0  // Increase to 3Ω per cell for 16nm
params.RbitLine = 3.0

analysis := array.AnalyzeIRDrop(input, params)

fmt.Printf("Max IR drop:     %.2f%%\n", analysis.MaxIRDrop*100)
fmt.Printf("Avg IR drop:     %.2f%%\n", analysis.AvgIRDrop*100)
fmt.Printf("Worst cell:      (%d, %d)\n",
    analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])

// Visualize as heatmap (normalized 0-1)
heatmap := analysis.GetIRDropMap()
```

### DefaultWireParams

Returns typical wire parameters for 45nm technology.

**Signature:**
```go
func DefaultWireParams() *WireParams
```

**Returns:**
- `*WireParams`: Default values: 2.5Ω per cell pitch (WL/BL), 50Ω contact resistance

**Example:**
```go
params := DefaultWireParams()
params.RwordLine = 2.0  // Adjust for different technology node
```

### AnalyzeSneakPaths

Analyzes parasitic current paths from unselected cells.

**Signature:**
```go
func (a *Array) AnalyzeSneakPaths(selectedRow, selectedCol int) *SneakPathAnalysis
```

**Parameters:**
- `selectedRow`, `selectedCol`: Target cell being read

**Returns:**
- `*SneakPathAnalysis`: Sneak current distribution and ratios

**Physics:**
- **Same row**: Current leaks from selected word line through unselected cells to other bit lines
- **Same column**: Current from unselected word lines leaks through unselected cells to selected bit line
- **Off-diagonal**: Three-cell loop paths: WL_selected → cell(selected_row, j) → BL_j → cell(i, j) → WL_i → cell(i, selected_col) → BL_selected

**Returns metrics:**
- `MaxSneakRatio`: Largest sneak/signal ratio
- `AvgSneakRatio`: Average sneak/signal ratio
- `TotalSneak`: Sum of all sneak conductances

**Example:**
```go
analysis := array.AnalyzeSneakPaths(15, 20)

fmt.Printf("Max sneak ratio: %.3f\n", analysis.MaxSneakRatio)
fmt.Printf("Avg sneak ratio: %.3f\n", analysis.AvgSneakRatio)

// Visualize sneak heatmap
heatmap := analysis.GetSneakMap()
```

### AnalyzeSneakPathsWithArch

Architecture-aware sneak path analysis with transistor isolation factor.

**Signature:**
```go
func (a *Array) AnalyzeSneakPathsWithArch(selectedRow, selectedCol int, is1T1R bool) *SneakPathAnalysis
```

**Parameters:**
- `is1T1R`: If true, applies ~1000x (10^-3) isolation factor from transistor

**Returns:**
- `*SneakPathAnalysis`: Sneak analysis with architecture effects

**Isolation Factor:**
- **0T1R** (passive): factor = 1.0 (no isolation)
- **1T1R** (transistor): factor = 0.001 (conservative ~1000x ON/OFF ratio)

**Example:**
```go
// Passive crossbar (0T1R)
sneak0T1R := array.AnalyzeSneakPathsWithArch(10, 15, false)
fmt.Printf("0T1R max sneak ratio: %.2f\n", sneak0T1R.MaxSneakRatio)

// Transistor-isolated crossbar (1T1R)
sneak1T1R := array.AnalyzeSneakPathsWithArch(10, 15, true)
fmt.Printf("1T1R max sneak ratio: %.4f\n", sneak1T1R.MaxSneakRatio)
```

### MVMWithIRDrop

Performs MVM with IR drop effects applied to effective voltages.

**Signature:**
```go
func (a *Array) MVMWithIRDrop(input []float64, params *WireParams) ([]float64, *IRDropAnalysis, error)
```

**Parameters:**
- `input`: Input vector
- `params`: Wire parameters (nil uses defaults)

**Returns:**
- `[]float64`: Output with IR drop effects
- `*IRDropAnalysis`: Voltage drop analysis
- `error`: Non-nil if input exceeds columns

**Effect:**
- For each cell: effective voltage = nominal voltage × relative effective voltage
- Cells with high IR drop see reduced effective voltage → less current
- Worse in bottom-right corner (maximum cumulative drop)

**Example:**
```go
output, analysis, err := array.MVMWithIRDrop(input, nil)
if err != nil {
    log.Fatal(err)
}
```

---

## Drift Simulation

Conductance drift models long-term retention characteristics of the memory device.

### NewDriftSimulator

Creates a drift simulator for analyzing retention and aging effects.

**Signature:**
```go
func NewDriftSimulator(rows, cols int, levels int) *DriftSimulator
```

**Parameters:**
- `rows`, `cols`: Array dimensions
- `levels`: Number of discrete levels (typically 30 for FeCIM)

**Returns:**
- `*DriftSimulator`: Initialized with random conductance levels, default assumed drift model

**Default Configuration:**
- Drift coefficient: 0.001 (assumed, see important note below)
- Temperature: 300K (room temperature)
- Read disturb: 1e-6 per read (very low for FeFET)

**IMPORTANT - Drift Coefficient Source:**
The default FeFET drift coefficient (0.001) is **derived from retention requirements**, not directly measured. Literature shows FeFET excellent >10 year retention at 85°C, which implies low drift, but does not provide explicit drift coefficients. Three models available:
1. **DriftModelAssumed**: Conservative estimate (0.001)
2. **DriftModelLiterature**: Derived from retention studies (0.0005)
3. **DriftModelMeasured**: User-provided from calibration

### NewDriftSimulatorWithModel

Creates a drift simulator with specified drift model source.

**Signature:**
```go
func NewDriftSimulatorWithModel(rows, cols int, levels int, model DriftModel) *DriftSimulator
```

**Parameters:**
- `model`: DriftModelAssumed, DriftModelLiterature, or DriftModelMeasured

**Returns:**
- `*DriftSimulator`: Configured with selected model

**Example:**
```go
// Use literature-derived value
sim := NewDriftSimulatorWithModel(64, 64, 30, DriftModelLiterature)
fmt.Printf("Drift coefficient: %.6f\n", sim.DriftCoeff)  // 0.0005

// Use custom measured value
sim := NewDriftSimulator(64, 64, 30)
sim.SetMeasuredDriftCoeff(0.00015)
```

### SimulateTimeStep

Advances drift simulation by dt seconds.

**Signature:**
```go
func (d *DriftSimulator) SimulateTimeStep(dt float64) error
```

**Parameters:**
- `dt`: Time increment in seconds

**Returns:**
- `error`: Currently always nil (reserved for future validation)

**Physics:**
- Drift model: G(t) = G₀ × (1 + v × ln(t+1) × exp(-E_a/kT))
- v: drift coefficient
- E_a: activation energy (~0.5 eV for ferroelectric)
- T: temperature in Kelvin
- Includes thermal activation via Arrhenius model
- Random component: device-to-device variation

**Example:**
```go
sim := NewDriftSimulator(64, 64, 30)

// Simulate 1 year (365.25 days)
oneYear := 365.25 * 24 * 3600
dt := oneYear / 100.0  // 100 time steps

for i := 0; i < 100; i++ {
    sim.SimulateTimeStep(dt)
}

stats := sim.GetStats()
fmt.Printf("After 1 year: %.2f%% retention\n", stats.RetentionPrediction)
```

### SetConductanceLevel

Sets a cell to a specific discrete level [0, levels).

**Signature:**
```go
func (d *DriftSimulator) SetConductanceLevel(row, col, level int)
```

**Parameters:**
- `row`, `col`: Cell indices
- `level`: Discrete level [0, levels-1]

**Example:**
```go
// Set cell to level 15 (middle)
sim.SetConductanceLevel(10, 20, 15)
```

### SetWeightMatrix

Programs a weight matrix (specified as level integers).

**Signature:**
```go
func (d *DriftSimulator) SetWeightMatrix(weights [][]int)
```

**Parameters:**
- `weights`: Integer weight matrix where each element is a level [0, levels-1]

**Example:**
```go
weights := [][]int{
    {0, 5, 10, 15, 20, 25, 29},
    {10, 15, 20, 25, 29, 0, 5},
}
sim.SetWeightMatrix(weights)
```

### SimulateRead

Simulates read disturb effects on a cell.

**Signature:**
```go
func (d *DriftSimulator) SimulateRead(row, col int, numReads int)
```

**Parameters:**
- `row`, `col`: Cell to read
- `numReads`: Number of read operations

**Returns:**
- No error return; invalid indices are silently ignored

**Example:**
```go
// Simulate reading the same cell 1000 times
sim.SimulateRead(10, 20, 1000)
```

### RefreshCell

Refreshes a cell to its nearest valid discrete level (error correction).

**Signature:**
```go
func (d *DriftSimulator) RefreshCell(row, col int)
```

**Effect:**
- Rounds current conductance to nearest discrete level
- Resets both current and initial conductance to rounded value

### RefreshAll

Refreshes all cells to nearest valid levels.

**Signature:**
```go
func (d *DriftSimulator) RefreshAll()
```

**Example:**
```go
// After 10 years, refresh all cells
sim.SimulateTimeStep(10 * 365.25 * 24 * 3600)
sim.RefreshAll()  // Reset to nearest levels
```

### GetStats

Returns comprehensive drift statistics.

**Signature:**
```go
func (d *DriftSimulator) GetStats() DriftStats
```

**Returns:**
- `DriftStats`: Detailed retention and degradation metrics

**Metrics:**
- `ElapsedTime`: Simulation time (seconds)
- `AvgDrift`: Average conductance shift (Siemens)
- `MaxDrift`: Maximum conductance shift
- `AvgDriftPercent`: Average drift as % of mid-range conductance
- `MaxDriftPercent`: Maximum drift as percentage
- `NumLevelErrors`: Cells that changed discrete level
- `LevelErrorRate`: Percentage of cells with level errors
- `RetentionPrediction`: Predicted 10-year retention at current temperature
- `TechnologyComparison`: Drift comparison with RRAM, PCM, Flash

**Example:**
```go
stats := sim.GetStats()
fmt.Printf("Retention prediction: %.2f%%\n", stats.RetentionPrediction)
fmt.Printf("Level errors: %d (%.2f%%)\n",
    stats.NumLevelErrors, stats.LevelErrorRate)
fmt.Printf("Technology comparison:\n")
fmt.Printf("  FeCIM drift:   %.4f\n", stats.TechnologyComparison.FeFETDrift)
fmt.Printf("  RRAM drift:    %.4f\n", stats.TechnologyComparison.RRAMDrift)
fmt.Printf("  FeCIM advantage: %.1fx vs RRAM\n",
    stats.TechnologyComparison.FeFETAdvantage)
```

### RecordSnapshot

Records current drift state for time-series analysis.

**Signature:**
```go
func (d *DriftSimulator) RecordSnapshot()
```

**Effect:**
- Appends DriftSnapshot to DriftHistory
- Snapshot includes: time, average drift, max drift, level changes, conductance samples

**Example:**
```go
for year := 0; year <= 10; year++ {
    // Simulate one year
    oneYear := 365.25 * 24 * 3600
    for i := 0; i < 100; i++ {
        sim.SimulateTimeStep(oneYear / 100.0)
    }
    sim.RecordSnapshot()
}

// Analyze history
for i, snapshot := range sim.DriftHistory {
    fmt.Printf("Year %d: %.4f%% max drift, %d level errors\n",
        i, snapshot.MaxDrift*100, snapshot.NumLevelChanges)
}
```

---

## Temperature Effects

Temperature significantly affects ferroelectric device physics. All temperature-dependent calculations use Arrhenius activation energy model.

### NewTemperatureEffects

Creates a temperature effects model.

**Signature:**
```go
func NewTemperatureEffects(tempK float64) *TemperatureEffects
```

**Parameters:**
- `tempK`: Operating temperature in Kelvin

**Returns:**
- `*TemperatureEffects`: Configured for specified temperature

**Temperature Presets:**
- `TempColdSpace` (4K): Deep cryogenic operation
- `TempCryogenic` (77K): Liquid nitrogen
- `TempRoom` (300K): Room temperature (27°C) - DEFAULT
- `TempIndustrial` (358K): Industrial grade (85°C)
- `TempAutomotive` (400K): Automotive Grade 0 (125°C)

**Example:**
```go
// Room temperature
tempFx := NewTemperatureEffects(300.0)

// Cryogenic operation
tempFx := NewTemperatureEffects(TempCryogenic)
label := tempFx.GetTemperatureLabel()  // "Cryogenic"
```

### AdjustedWireResistance

Applies temperature coefficient of resistance (TCR) to wire resistance.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedWireResistance(R0 float64) float64
```

**Parameters:**
- `R0`: Base wire resistance at 300K

**Returns:**
- Adjusted resistance accounting for temperature

**Model:**
- Uses copper TCR = 0.00393 /K
- Higher temperature → higher resistance → more IR drop

**Example:**
```go
temp := NewTemperatureEffects(350.0)  // 77°C
adjusted := temp.AdjustedWireResistance(2.5)
// If R0 = 2.5Ω, adjusted ≈ 2.56Ω
```

### AdjustedConductanceRange

Scales minimum and maximum conductance with temperature.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedConductanceRange(gMin, gMax float64) (float64, float64)
```

**Returns:**
- Adjusted (gMin, gMax) pair

**Physics:**
- **Cryogenic (<100K)**: Enhanced ferroelectric polarization. Pr reaches 75 µC/cm² at 4K (vs 15-34 µC/cm² at RT). Window expands 50% at 4K.
- **High temperature (>300K)**: Thermal noise reduces effective window. Conservative 10% degradation per 100K.

**Example:**
```go
cryo := NewTemperatureEffects(4.0)
gmin, gmax := cryo.AdjustedConductanceRange(1e-6, 100e-6)
// At 4K: window ~1.5x wider than room temperature
fmt.Printf("Cryogenic window: %.2e to %.2e S\n", gmin, gmax)
```

### AdjustedDriftRate

Scales drift rate using Arrhenius thermal activation.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedDriftRate(driftCoeff float64) float64
```

**Parameters:**
- `driftCoeff`: Base drift coefficient at 300K

**Returns:**
- Temperature-adjusted drift coefficient

**Model:**
- Drift ∝ exp(-E_a / kT)
- Activation energy: ~0.5 eV
- Higher temperature → exponentially faster drift

**Example:**
```go
// At 85°C (358K): drift accelerates significantly
industrial := NewTemperatureEffects(358.0)
adjusted := industrial.AdjustedDriftRate(0.001)
// adjusted > 0.001 (faster drift)

// At liquid nitrogen (77K): drift dramatically slows
cryo := NewTemperatureEffects(77.0)
adjusted := cryo.AdjustedDriftRate(0.001)
// adjusted << 0.001 (much slower drift)
```

### AdjustedRetention

Estimates retention improvement factor at lower temperatures.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedRetention() float64
```

**Returns:**
- Retention time multiplier (>1 = improved retention)

**Example:**
```go
// Room temperature: baseline
room := NewTemperatureEffects(300.0)
retention := room.AdjustedRetention()  // ~1.0

// Cryogenic: massive improvement
cryo := NewTemperatureEffects(4.0)
retention := cryo.AdjustedRetention()
// Retention could improve 10,000x+ due to exponential Arrhenius scaling
```

### AdjustedNoise

Scales thermal noise with temperature.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedNoise() float64
```

**Returns:**
- Noise multiplier relative to 300K

**Model:**
- RMS voltage noise ∝ sqrt(T)
- Colder → less noise (better signal)

**Example:**
```go
cryo := NewTemperatureEffects(77.0)
noiseFactor := cryo.AdjustedNoise()  // sqrt(77/300) ≈ 0.51
// 77K has ~51% the noise of room temperature
```

### AdjustedSwitchingEnergy

Estimates switching energy scaling with temperature.

**Signature:**
```go
func (t *TemperatureEffects) AdjustedSwitchingEnergy(baseEnergy float64) float64
```

**Returns:**
- Adjusted switching energy

**Example:**
```go
temp := NewTemperatureEffects(350.0)
adjusted := temp.AdjustedSwitchingEnergy(1.0)
// High temperature slightly reduces switching energy
```

### GetTemperatureLabel

Returns a human-readable temperature classification.

**Signature:**
```go
func (t *TemperatureEffects) GetTemperatureLabel() string
```

**Returns:**
- Label: "Deep Cryogenic", "Cryogenic", "Cold", "Room Temperature", "Industrial", "Automotive", or "Extreme Heat"

**Example:**
```go
temp := NewTemperatureEffects(358.0)
fmt.Println(temp.GetTemperatureLabel())  // "Industrial"
```

### GetAdjustedParams

Returns all temperature-adjusted parameters in a single struct.

**Signature:**
```go
func (t *TemperatureEffects) GetAdjustedParams() *TemperatureAdjustedParams
```

**Returns:**
- `*TemperatureAdjustedParams`: All adjustment factors

**Example:**
```go
temp := NewTemperatureEffects(77.0)
params := temp.GetAdjustedParams()

fmt.Printf("Wire resistance factor: %.3f\n", params.WireResistanceFactor)
fmt.Printf("Conductance range: %.2e to %.2e\n",
    params.GminAdjusted, params.GmaxAdjusted)
fmt.Printf("Drift rate factor: %.6f\n", params.DriftRateFactor)
fmt.Printf("Noise factor: %.3f\n", params.NoiseFactor)
fmt.Printf("Retention improvement: %.0fx\n", params.RetentionFactor)
```

---

## Types Reference

### Array

The primary crossbar array type encapsulating cells, configuration, and statistics.

**Fields:**
- `config`: Configuration (private, access via GetConfig())
- `cells`: Cell matrix (private)
- `adcLevels`: ADC quantization levels
- `dacLevels`: DAC quantization levels
- `totalReads`: Cumulative read count
- `totalWrites`: Cumulative write count

**Key Methods:**
- Array creation: `NewArray(cfg)`
- Operations: `MVM()`, `VMM()`, `MVMWithNonIdealities()`
- Weight programming: `ProgramWeight()`, `ProgramWeightMatrix()`, `ProgramWeightWithDisturb()`
- Configuration: `GetConfig()`, `Rows()`, `Cols()`, `SetConductanceModel()`
- Analysis: `AnalyzeIRDrop()`, `AnalyzeSneakPaths()`, `GetCellStats()`
- Maintenance: `ResetDisturbTracking()`, `ResetCycleCounts()`, `AgeCycles()`

### Config

Crossbar array configuration.

**Fields:**
```go
type Config struct {
    Rows       int                    // Number of rows
    Cols       int                    // Number of columns
    NoiseLevel float64                // Device variation [0, 1]
    ADCBits    int                    // ADC resolution
    DACBits    int                    // DAC resolution
    ConductanceModel ConductanceModel // Model: Linear, Exponential, or Lookup
    ConductanceTable []float64        // Calibration table (30 entries)
    Endurance  *EnduranceConfig       // Fatigue/durability modeling
    ProcessVariation *ProcessVariationConfig  // Spatial variation
    HalfSelect *HalfSelectConfig      // Disturb modeling
}
```

**Example:**
```go
cfg := &Config{
    Rows:       128,
    Cols:       128,
    NoiseLevel: 0.02,
    ADCBits:    10,
    DACBits:    10,
    ConductanceModel: ConductanceExponential,
    ConductanceTable: nil,  // Use exponential model, not lookup
    Endurance: DefaultEnduranceConfig(),
    ProcessVariation: DefaultProcessVariationConfig(),
    HalfSelect: DefaultHalfSelectConfig(),
}

array, _ := NewArray(cfg)
```

### Cell

Represents a single ferroelectric memory cell.

**Fields:**
```go
type Cell struct {
    Conductance    float64  // Normalized [0, 1]
    NoiseFactor    float64  // Per-cell device variation
    SwitchingCount int64    // Write cycles (endurance tracking)
    HalfSelectCount int64   // V/2 stress count
    DisturbShift   float64  // Accumulated drift from half-select
}
```

### CellStats

Detailed statistics for a single cell retrieved via `GetCellStats(row, col)`.

**Fields:**
```go
type CellStats struct {
    Row             int
    Col             int
    Conductance     float64 // Normalized [0, 1]
    Level           int     // Discrete level [0-29]
    PhysicalG       float64 // Physical conductance in Siemens
    NoiseFactor     float64
    SwitchingCount  int64
    HalfSelectCount int64
    DisturbShift    float64
    VariationFactor float64 // Process variation factor
}
```

**Example:**
```go
stats, _ := array.GetCellStats(10, 20)
fmt.Printf("Cell [10,20] level: %d, G=%.2e S\n", stats.Level, stats.PhysicalG)
fmt.Printf("Switching count: %d\n", stats.SwitchingCount)
fmt.Printf("Variation factor: %.4f\n", stats.VariationFactor)
```

### EnduranceConfig

Configures endurance/fatigue modeling.

**Fields:**
```go
type EnduranceConfig struct {
    Enabled          bool  // Enable fatigue modeling
    FatigueThreshold int64 // Cycles before degradation (default: 10^8)
    FailureThreshold int64 // 50% window loss (default: 10^12)
}
```

**Sources:**
- Fatigue threshold: IEEE IRPS 2022 (10^9 cycle endurance)
- Failure threshold: Nano Letters 2024 (10^12 cycle endurance for V:HfO₂)

### ProcessVariationConfig

Systematic process variation modeling.

**Fields:**
```go
type ProcessVariationConfig struct {
    DeviceSigma float64 // Device-to-device variation sigma
    GradientX   float64 // Horizontal gradient (%/cell)
    GradientY   float64 // Vertical gradient (%/cell)
    EdgeEffect  float64 // Edge cell degradation [0, 1]
}
```

**Default values:**
- DeviceSigma: 0.02 (2%)
- GradientX: 0.001 (0.1% per cell)
- GradientY: 0.001 (0.1% per cell)
- EdgeEffect: 0.05 (5% degradation at edges)

### HalfSelectConfig

Configuration for half-select disturb in passive crossbars.

**Fields:**
```go
type HalfSelectConfig struct {
    Enabled          bool    // Enable disturb tracking
    DisturbThreshold float64 // V/Vc ratio threshold
    DisturbRate      float64 // Conductance shift per pulse
}
```

### MVMOptions

Options for `MVMWithNonIdealities()`.

**Fields:**
```go
type MVMOptions struct {
    EnableIRDrop     bool
    EnableSneakPaths bool
    EnableVariation  bool
    EnableDrift      bool
    Temperature      float64 // Kelvin
    Architecture     string  // "1T1R" or "0T1R"
}
```

**Methods:**
- `Is1T1R()`: Returns true if 1T1R architecture

### MVMResult

Detailed results from `MVMWithNonIdealities()`.

**Fields:**
```go
type MVMResult struct {
    // Outputs
    IdealOutput  []float64
    ActualOutput []float64

    // Error metrics
    RMSE         float64
    MaxError     float64
    MeanError    float64
    AccuracyLoss float64

    // Energy (pJ)
    ArrayEnergy float64
    ADCEnergy   float64
    DACEnergy   float64
    TotalEnergy float64

    // Performance
    MACOperations int
    Latency       float64
    Throughput    float64

    // Analysis
    IRDropAnalysis    *IRDropAnalysis
    SneakPathAnalysis *SneakPathAnalysis

    // GPU comparison
    GPUEquivalentEnergy float64
    EnergyEfficiency    float64
}
```

### WireParams

Wire resistance parameters for IR drop analysis.

**Fields:**
```go
type WireParams struct {
    RwordLine float64 // Word line resistance per cell
    RbitLine  float64 // Bit line resistance per cell
    Rcontact  float64 // Contact resistance
}
```

**Default values** (45nm):
- RwordLine: 2.5 Ω/cell
- RbitLine: 2.5 Ω/cell
- Rcontact: 50 Ω

### IRDropAnalysis

Results from IR drop analysis.

**Fields:**
```go
type IRDropAnalysis struct {
    WordLineVoltages [][]float64 // Voltage at each position
    BitLineVoltages  [][]float64 // Voltage at each position
    EffectiveVoltage [][]float64 // Cell voltage = WL - BL

    MaxIRDrop      float64 // Maximum drop
    AvgIRDrop      float64 // Average drop
    IRDropVariance float64 // Variance
    WorstCaseCell  [2]int  // Location of worst cell
}
```

**Methods:**
- `GetIRDropMap()`: Returns heatmap normalized to fixed 15% scale

### SneakPathAnalysis

Results from sneak path analysis.

**Fields:**
```go
type SneakPathAnalysis struct {
    SneakCurrents [][]float64

    MaxSneakRatio float64 // Largest sneak/signal ratio
    AvgSneakRatio float64 // Average sneak/signal ratio
    TotalSneak    float64
    TotalSignal   float64
}
```

**Methods:**
- `GetSneakMap()`: Returns heatmap normalized to fixed 2.0 scale

### DriftSimulator

Models conductance drift over time.

**Fields:**
```go
type DriftSimulator struct {
    Rows           int
    Cols           int
    Conductances   [][]float64
    InitialConds   [][]float64
    Levels         int
    GMin           float64
    GMax           float64

    DriftCoeff     float64
    DriftModel     DriftModel
    ReadDisturb    float64
    Temperature    float64
    Time           float64

    DriftHistory   []DriftSnapshot
}
```

### DriftStats

Statistics from `GetStats()`.

**Fields:**
```go
type DriftStats struct {
    ElapsedTime         float64
    AvgDrift            float64
    MaxDrift            float64
    AvgDriftPercent     float64
    MaxDriftPercent     float64
    NumLevelErrors      int
    LevelErrorRate      float64
    RetentionPrediction float64 // Predicted 10-year retention
    TechnologyComparison TechDriftComparison
}
```

### TechDriftComparison

Technology drift comparison in DriftStats.

**Fields:**
```go
type TechDriftComparison struct {
    FeFETDrift      float64 // FeCIM drift coefficient
    RRAMDrift       float64 // RRAM drift coefficient
    PCMDrift        float64 // PCM drift coefficient
    FlashDrift      float64 // Flash drift coefficient
    FeFETAdvantage  float64 // Advantage factor vs RRAM
}
```

### TemperatureEffects

Temperature-dependent physics model.

**Fields:**
```go
type TemperatureEffects struct {
    AmbientK float64 // Operating temperature in Kelvin
}
```

### TemperatureAdjustedParams

All temperature adjustment factors.

**Fields:**
```go
type TemperatureAdjustedParams struct {
    WireResistanceFactor float64
    GminAdjusted         float64
    GmaxAdjusted         float64
    DriftRateFactor      float64
    NoiseFactor          float64
    RetentionFactor      float64
}
```

### DriftSnapshot

Time-series snapshot of drift state.

**Fields:**
```go
type DriftSnapshot struct {
    Time            float64   // Time point
    AvgDrift        float64   // Average drift
    MaxDrift        float64   // Maximum drift
    NumLevelChanges int       // Level errors
    WorstCellRow    int
    WorstCellCol    int
    Conductances    []float64 // Sample (first row)
}
```

---

## Constants and Models

### Conductance Models

```go
const (
    ConductanceLinear      // G = Gmin + gNorm*(Gmax-Gmin)
    ConductanceExponential // G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
    ConductanceLookup      // G = ConductanceTable[level]
)
```

**Selection:**
- **Linear**: Simple, fast, less accurate at extremes
- **Exponential**: Realistic FeFET behavior (recommended)
- **Lookup**: Calibration-based from device measurements

### Drift Models

```go
const (
    DriftModelAssumed     // 0.001 (estimated from retention)
    DriftModelLiterature  // 0.0005 (literature-derived)
    DriftModelMeasured    // User-specified
)
```

### Quantization

```go
const DefaultQuantizationLevels = 30  // FeCIM discrete states
```

### Physical Constants

```go
const (
    GMin = 10e-6  // 10 µS minimum conductance (OFF)
    GMax = 100e-6 // 100 µS maximum conductance (ON)
)
```

---

## Common Patterns

### Complete Simulation with All Non-Idealities

```go
// Create array
cfg := &Config{
    Rows:       64,
    Cols:       64,
    NoiseLevel: 0.02,
    ADCBits:    8,
    DACBits:    8,
    ConductanceModel: ConductanceExponential,
    Endurance: DefaultEnduranceConfig(),
    ProcessVariation: DefaultProcessVariationConfig(),
    HalfSelect: DefaultHalfSelectConfig(),
}
array, _ := NewArray(cfg)

// Program weights
weights := make([][]float64, 64)
// ... fill weights ...
array.ProgramWeightMatrix(weights)

// Run MVM with all non-idealities
opts := &MVMOptions{
    EnableIRDrop:     true,
    EnableSneakPaths: true,
    EnableVariation:  true,
    EnableDrift:      false,
    Temperature:      300.0,
    Architecture:     "0T1R",
}

input := make([]float64, 64)
// ... prepare input ...

result, _ := array.MVMWithNonIdealities(input, opts)

fmt.Printf("Output error: %.4f RMSE\n", result.RMSE)
fmt.Printf("Energy: %.2f pJ\n", result.TotalEnergy)
```

### Temperature and Drift Study

```go
// Temperature effects
temps := []float64{77.0, 300.0, 358.0}
for _, t := range temps {
    tempFx := NewTemperatureEffects(t)
    params := tempFx.GetAdjustedParams()
    fmt.Printf("%.0fK: retention %.0fx, noise %.2fx\n",
        t, params.RetentionFactor, params.NoiseFactor)
}

// Long-term drift simulation
sim := NewDriftSimulatorWithModel(64, 64, 30, DriftModelLiterature)
for year := 0; year <= 10; year++ {
    // Simulate one year
    oneYear := 365.25 * 24 * 3600
    sim.SimulateTimeStep(oneYear)
}

stats := sim.GetStats()
fmt.Printf("10-year retention: %.2f%%\n", stats.RetentionPrediction)
```

### Architecture Comparison (0T1R vs 1T1R)

```go
input := []float64{0.5, 0.3, 0.7}

// Passive crossbar (0T1R)
opts0T1R := &MVMOptions{
    EnableSneakPaths: true,
    Architecture:     "0T1R",
}
result0T1R, _ := array.MVMWithNonIdealities(input, opts0T1R)

// Transistor-isolated (1T1R)
opts1T1R := &MVMOptions{
    EnableSneakPaths: true,
    Architecture:     "1T1R",
}
result1T1R, _ := array.MVMWithNonIdealities(input, opts1T1R)

fmt.Printf("0T1R sneak ratio: %.3f\n",
    result0T1R.SneakPathAnalysis.AvgSneakRatio)
fmt.Printf("1T1R sneak ratio: %.6f\n",
    result1T1R.SneakPathAnalysis.AvgSneakRatio)
```

---

## Utility Functions

### ComputeError

Package-level utility to calculate RMS error between two vectors.

**Signature:**
```go
func ComputeError(ideal, actual []float64) float64
```

**Parameters:**
- `ideal`: Expected output vector
- `actual`: Measured output vector

**Returns:**
- RMS error: sqrt(sum((ideal - actual)^2) / len)
- `math.Inf(1)` if vectors have different lengths

**Example:**
```go
rmse := ComputeError(result.IdealOutput, result.ActualOutput)
fmt.Printf("Output error: %.4f\n", rmse)
```

### CompareTechnologies

Package-level function comparing drift characteristics across memory technologies.

**Signature:**
```go
func CompareTechnologies(rows, cols int, simulationTime float64) map[string]DriftStats
```

**Parameters:**
- `rows`, `cols`: Array dimensions
- `simulationTime`: Simulation duration in seconds

**Returns:**
- `map[string]DriftStats`: Drift statistics for each technology
  - "FeCIM (FeFET)": 0.001 drift coefficient (assumed)
  - "RRAM": 0.05 drift coefficient
  - "PCM": 0.1 drift coefficient
  - "Flash": 0.02 drift coefficient

**Example:**
```go
// Compare 100 days of drift
results := CompareTechnologies(64, 64, 100*24*3600)

for tech, stats := range results {
    fmt.Printf("%s:\n", tech)
    fmt.Printf("  Retention: %.2f%%\n", stats.RetentionPrediction)
    fmt.Printf("  Max drift: %.2e S\n", stats.MaxDrift)
    fmt.Printf("  Level errors: %d\n", stats.NumLevelErrors)
}
```

**IMPORTANT:** FeCIM drift coefficient (0.001) is estimated from retention requirements, not directly measured. This function is for comparative analysis only. Refer to `FeFETDriftCoefficients` for detailed source documentation.

---

## References

**Peer-Reviewed Sources:**
- IEEE IRPS 2022: FeFET endurance characterization (10^9 cycles)
- Nano Letters 2024: V:HfO₂ superlattice 10^12 cycle endurance
- Nature Communications 2025: HZO polarization (15-34 µC/cm² at RT, 75 µC/cm² at 4K)
- Fraunhofer IPMS 2024: >10 year retention at 85°C (AEC-Q100 automotive)

**Implementation Notes:**
- All weights automatically quantize to 30 discrete levels
- Temperature effects use Arrhenius model (E_a ≈ 0.5 eV)
- 0T1R and 1T1R architectures have different sneak/IR drop characteristics
- Drift coefficient for FeFET is estimated (not directly measured) from retention requirements
- Process variation combines random device variation with spatial gradients

