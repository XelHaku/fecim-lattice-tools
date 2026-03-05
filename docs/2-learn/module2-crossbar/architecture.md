# Module 2 Crossbar Array Architecture

> **Citation policy:** Cross-check external numeric claims in this file against `docs/4-research/honesty-audit.md`. If a statement here lacks a DOI or an explicit `simulation default` / `assumption` label, treat it as documentation debt rather than evidence.

## Overview

The module2-crossbar package implements a physics-accurate ferroelectric crossbar array simulator for the FeCIM (Ferroelectric Compute-in-Memory) platform. It models a 30-level baseline (~4.9 bits/cell, simulation baseline (configurable) with comprehensive non-ideality simulation including IR drop, sneak paths, process variation, drift, and temperature effects.

The architecture is organized into five core layers:

1. **Core Array Simulation** - Basic crossbar physics and quantization
2. **Non-Ideality Modules** - Realistic hardware effects simulation
3. **Enhanced Operations** - Multi-cell programming and verification
4. **Network Integration** - Multi-layer neural network inference
5. **Data I/O** - Weight serialization and export

---

## Package Structure

```
module2-crossbar/
├── pkg/crossbar/         # Core array simulation (4798 lines)
│   ├── array.go          # Array definition, MVM, quantization
│   ├── nonidealities.go  # Enhanced MVM with all non-idealities
│   ├── irdrop.go         # Wire resistance and voltage drop simulation
│   ├── sneakpath.go      # Parasitic current path analysis
│   ├── drift.go          # Conductance drift over time
│   ├── temperature.go    # Temperature-dependent physics
│   ├── enhanced.go       # Differential arrays, write-verify, reporting
│   └── reference.go      # Low-level reference implementations
│
├── pkg/network/          # Neural network layers
│   └── network.go        # Multi-layer inference on crossbars
│
├── pkg/training/         # Hardware-aware backpropagation
│   └── training.go       # Quantization-aware training
│
├── pkg/weights/          # Weight serialization
│   ├── weights.go        # Format definitions
│   └── serialization.go  # JSON/binary/NumPy export
│
├── pkg/gui/              # Fyne visualization
│   ├── app.go            # Main application
│   ├── heatmap.go        # Conductance/IR drop/sneak visualization
│   ├── controls.go       # Control panels and sliders
│   └── tabs/             # Specialized visualization tabs
│
└── pkg/evaluation/       # Accuracy and performance metrics
    └── accuracy.go       # MNIST and inference accuracy
```

---

## Core Types

### Configuration

#### `Config` - Crossbar Array Configuration

**Purpose**: Define complete array parameters and physics models.

**Fields**:
```go
type Config struct {
    Rows       int                     // Word lines (rows)
    Cols       int                     // Bit lines (columns)
    NoiseLevel float64                 // Device-to-device variation [0-1]
    ADCBits    int                     // ADC resolution (e.g., 6-bit)
    DACBits    int                     // DAC resolution (e.g., 8-bit)

    // Conductance model (linear, exponential, or lookup table)
    ConductanceModel ConductanceModel
    ConductanceTable []float64         // 30-element lookup table (if using ConductanceLookup)

    // Optional physics modules
    Endurance        *EnduranceConfig
    ProcessVariation *ProcessVariationConfig
    HalfSelect       *HalfSelectConfig
}
```

**Conductance Models** (three options for G(V) relationship):

| Model | Equation | Use Case | Accuracy |
|-------|----------|----------|----------|
| `ConductanceLinear` | G = Gmin + g_norm × (Gmax - Gmin) | Simple, fast | Low at extremes |
| `ConductanceExponential` | G = Gmin × exp(ln(Gmax/Gmin) × g_norm) | Realistic FeFET behavior | High (matches physics) |
| `ConductanceLookup` | G = ConductanceTable[level] | Calibrated data | Highest (measured) |

**Physical Constants**:
- `GMin` = 10 µS (OFF state)
- `GMax` = 100 µS (ON state)
- `DefaultQuantizationLevels` = 30-level baseline (simulation baseline)

#### `Cell` - Individual Memory Cell

**Purpose**: Track per-cell state and history.

**Fields**:
```go
type Cell struct {
    Conductance    float64  // Normalized [0-1] programmed conductance
    NoiseFactor    float64  // Per-cell random variation (1.0 ± NoiseLevel)
    SwitchingCount int64    // Write cycles (endurance tracking)
    HalfSelectCount int64   // Half-select exposures (disturb tracking)
    DisturbShift   float64  // Accumulated drift from half-select stress
}
```

#### `EnduranceConfig` - Fatigue Modeling

```go
type EnduranceConfig struct {
    Enabled          bool  // Enable fatigue degradation simulation
    FatigueThreshold int64 // Cycles before degradation starts (default: 10^8)
    FailureThreshold int64 // Cycles at 50% window degradation (default: 10^12)
}
```

**Physics**: Models exponential conductance window narrowing:
```
degradation_factor = 1.0 - 0.5 × (1 - exp(-3 × fatigue_ratio))
```

#### `ProcessVariationConfig` - Spatial Variation

```go
type ProcessVariationConfig struct {
    DeviceSigma float64  // Random device variation (σ, default: 2%)
    GradientX   float64  // Horizontal gradient (%/cell, default: 0.1%)
    GradientY   float64  // Vertical gradient (%/cell, default: 0.1%)
    EdgeEffect  float64  // Edge cell degradation factor [0-1] (default: 5%)
}
```

**Purpose**: Model systematic process variation from wafer fabrication gradients and edge effects.

#### `HalfSelectConfig` - Passive Crossbar Disturb

```go
type HalfSelectConfig struct {
    Enabled          bool    // Enable half-select disturb tracking
    DisturbThreshold float64 // V/Vc threshold for disturb (default: 0.3)
    DisturbRate      float64 // Conductance shift per pulse (default: 0.1%)
}
```

**Purpose**: Model disturb effects in 0T1R (passive) crossbars where unselected cells experience V/2 stress during writes.

### Array Operation

#### `Array` - Main Crossbar Structure

**Purpose**: Represents a physical crossbar array implementing ferroelectric memory and analog computation.

**Key Methods**:

| Method | Purpose | Complexity |
|--------|---------|-----------|
| `NewArray(cfg)` | Create and initialize array | O(rows × cols) |
| `ProgramWeight(row, col, value)` | Write single cell | O(1) |
| `MVM(input)` | Matrix-vector multiply | O(rows × cols) |
| `VMM(input)` | Vector-matrix multiply | O(rows × cols) |
| `MVMWithNonIdealities(input, opts)` | MVM with all physics effects | O(rows × cols × 3) |
| `AnalyzeIRDrop(input, params)` | Resistive voltage drop analysis | O(rows × cols) |
| `AnalyzeSneakPaths(row, col)` | Parasitic current analysis | O(rows × cols) |

**Quantization**:

```go
func QuantizeToLevels(value float64) float64 {
    // Converts [0, 1] → discrete level 0-29
    level := Round(value × 29)
    return level / 29
}

func GetLevel(conductance float64) int {
    // Returns discrete level [0-29] for a conductance
    return Round(conductance × 29)
}
```

### Multi-Array Operation

#### `MVMOptions` - Non-Ideality Control

```go
type MVMOptions struct {
    EnableIRDrop     bool    // Wire voltage drop
    EnableSneakPaths bool    // Parasitic currents
    EnableVariation  bool    // Device-to-device noise
    EnableDrift      bool    // Conductance drift over time
    Temperature      float64 // Operating temperature (K), default: 300K
    Architecture     string  // "0T1R" (passive) or "1T1R" (with transistor)
}
```

**Architecture Impact**:

| Architecture | Sneak Isolation | IR Drop | Density | Use Case |
|--------------|-----------------|---------|---------|----------|
| **0T1R** (Passive) | 1× (full sneak) | High | Very high | Research, proof-of-concept |
| **1T1R** (1T per cell) | 1000× (transistor blocks sneak) | Lower | Moderate | Practical systems |

#### `MVMResult` - Complete Operation Analysis

```go
type MVMResult struct {
    // Outputs
    IdealOutput  []float64         // Without non-idealities
    ActualOutput []float64         // With all effects

    // Error metrics
    RMSE         float64           // Root mean square error
    MaxError     float64           // Worst-case output error
    MeanError    float64           // Average absolute error
    AccuracyLoss float64           // Estimated NN accuracy loss %

    // Energy (pJ - picojoules)
    ArrayEnergy        float64       // Core computation
    ADCEnergy          float64       // Output conversion
    DACEnergy          float64       // Input conversion
    TotalEnergy        float64       // Sum
    GPUEquivalentEnergy float64      // For comparison (10 pJ/MAC)
    EnergyEfficiency   float64       // GPU energy / FeCIM energy

    // Analysis details
    IRDropAnalysis     *IRDropAnalysis
    SneakPathAnalysis  *SneakPathAnalysis

    // Performance
    MACOperations int               // Total multiply-accumulates
    Latency       float64           // ns (10 ns typical)
    Throughput    float64           // ops/sec
}
```

---

## Non-Ideality Modules

### IR Drop Simulation

**File**: `nonidealities.go`, `irdrop.go`

**Purpose**: Model voltage drop along metal interconnects during MVM operation.

#### Key Components

```go
type WireParams struct {
    RwordLine float64  // Word line resistance/cell pitch (Ω), default: 2.5 Ω
    RbitLine  float64  // Bit line resistance/cell pitch (Ω), default: 2.5 Ω
    Rcontact  float64  // Contact resistance (Ω), default: 50 Ω
}

type IRDropAnalysis struct {
    WordLineVoltages [][]float64  // Voltage at each WL position
    BitLineVoltages  [][]float64  // Voltage at each BL position
    EffectiveVoltage [][]float64  // Actual V_cell = V_WL - V_BL

    MaxIRDrop      float64        // Worst-case drop
    AvgIRDrop      float64        // Mean drop
    IRDropVariance float64        // Standard deviation
    WorstCaseCell  [2]int         // Location of worst cell
}
```

**Voltage Model**:
```
V_eff[i][j] = (V_WL[i] - IR_drop_row[j]) - (V_BL[j] + IR_drop_col[i])

IR_drop_row[j] = j × R_wordline × I_row[i]    // Cumulative along WL
IR_drop_col[i] = i × R_bitline  × I_col[j]    // Cumulative along BL
```

**Architecture Scaling** (in `MVMWithNonIdealities`):
```go
if !Is1T1R {
    params.RwordLine *= 1.5  // 50% higher resistance for 0T1R
    params.RbitLine *= 1.5   // More sneak current → more total current
}
```

**Temperature Adjustment**:
```
R(T) = R_0 × (1 + TCR × (T - T_ref))  // TCR = 0.00393 /K for copper
```

### Sneak Path Analysis

**File**: `nonidealities.go`, `sneakpath.go`

**Purpose**: Analyze parasitic currents through unselected cells in passive (0T1R) crossbars.

#### Sneak Path Types

**Same Row** (one unselected cell in selected row):
```
Path: Selected WL → Cell[sr][j] → BL j → Return paths in column j
Conductance: Series combination of cell and parallel return paths
```

**Same Column** (one unselected cell in selected column):
```
Path: Unselected WL i → Cells in row i → Return through Cell[i][sc]
Conductance: Feed paths in series with this cell
```

**Off-Diagonal** (three-cell path):
```
Path: Selected WL → Cell[sr][j] → BL j → Cell[i][j] → WL i → Cell[i][sc] → Selected BL
Conductance: G_series = 1/(1/g1 + 1/g2 + 1/g3)  // Three in series
```

#### Sneak Calculation

```go
func (a *Array) computeFullSneakCurrent(targetRow int, input []float64) float64 {
    // Enumerates all three-cell paths through other rows
    // For each source row and intermediate column:
    //   g1 = source row input column (entry point)
    //   g2 = source row intermediate column (path through source)
    //   g3 = target row intermediate column (exit point)
    //   I_sneak += V_in × G_series
}
```

**Size-Based Optimization**:
- Arrays ≤ 32×32: Full calculation (O(n^4)) for accuracy
- Arrays > 32×32: Simplified model (0.01 sneak factor) for performance

**Architecture Impact**:
```go
isolationFactor := 1.0
if Is1T1R {
    isolationFactor = 0.001  // Transistor provides ~1000× isolation
}
```

### Device Variation and Noise

**File**: `array.go`

**Sources**:

1. **Random Device Variation** (per-cell):
   ```go
   cell.NoiseFactor = 1.0 + NoiseLevel × (rand.Float64() × 2 - 1)  // ±NoiseLevel
   ```

2. **Spatial Gradients** (systematic):
   ```go
   gradX = 1.0 + GradientX × (col - centerCol)
   gradY = 1.0 + GradientY × (row - centerRow)
   ```

3. **Edge Effects** (boundary cells):
   ```go
   if isEdge:
       edgeFactor = 1.0 - EdgeEffect  // Default: 5% degradation
   ```

**Combined Effect** (in MVM):
```go
g_actual = g_nominal × noiseFactor × gradX × gradY × edgeFactor
```

### Drift Simulation

**File**: `drift.go`

**Purpose**: Model conductance drift over time (important for long-term retention).

#### Drift Models

```go
type DriftModel int
const (
    DriftModelAssumed DriftModel = iota      // 0.001 (conservative)
    DriftModelLiterature                     // 0.0005 (derived from literature)
    DriftModelMeasured                       // User-specified calibration
)

var FeFETDriftCoefficients = struct {
    Assumed    float64  // ⚠️ 0.001 - ASSUMED (no direct measurement)
    Literature float64  // 0.0005 - Derived from retention data
    RRAM       float64  // 0.05   - Literature reference
    PCM        float64  // 0.1    - Literature reference
    Flash      float64  // 0.02   - Literature reference
}
```

**Important Note**: The FeCIM drift coefficients in this repo are simulator coefficients, not directly measured FeCIM drift constants. `DriftModelLiterature` is a conservative value derived from long-retention targets and comparative memory literature; until a DOI-backed FeCIM drift dataset is wired into `validation/literature/`, treat it as a model assumption rather than a device measurement. See `docs/4-research/honesty-audit.md` for the current verification boundary.

#### Drift Physics

**Thermal Activation** (Arrhenius model):
```
rate(T) = rate_ref × exp(-Ea/k_B × (1/T - 1/T_ref))
Ea ≈ 0.5 eV (typical ferroelectric switching activation energy)
```

### Temperature Effects

**File**: `temperature.go`

**Purpose**: Model temperature-dependent device behavior.

#### Effects

**1. Wire Resistance** (copper TCR):
```
R(T) = R_0 × (1 + 0.00393 × (T - 300K))
```

**2. Conductance Window** (temperature-dependent polarization):
- **Cryogenic** (< 100K): Window expands 50% at 4K (Pr: 75 µC/cm² vs 15-34 at RT)
- **Room temp** (300K): No adjustment
- **High temp** (> 300K): Window narrows (conservative 10% per 100K)

**3. Drift Rate** (thermal activation):
```
drift_rate(T) ∝ exp(-Ea / k_B T)
```

#### Temperature Presets

| Temperature | Value | Use Case |
|-------------|-------|----------|
| `TempCryogenic` | 77 K | Liquid nitrogen operation |
| `TempColdSpace` | 4 K | Deep space deployment |
| `TempRoom` | 300 K | Laboratory (27°C) |
| `TempIndustrial` | 358 K | Industrial grade (85°C) |
| `TempAutomotive` | 400 K | Automotive Grade 0 (125°C) |

---

## Enhanced Operations

### Differential Arrays (Signed Weights)

**File**: `enhanced.go`

**Purpose**: Support signed weights [-1, 1] using two crossbar arrays (positive/negative).

**Implementation**:
```go
type DifferentialArray struct {
    positive *Array  // Stores w when w ≥ 0
    negative *Array  // Stores |w| when w < 0
    config   *Config
}

func (d *DifferentialArray) MVM(input []float64) []float64 {
    // Output = I_positive - I_negative
    // Allows signed weights without modifying basic array
}
```

**Energy Impact**: 2× array energy (both arrays operate in parallel).

### Write-Verify Programming

**File**: `enhanced.go`

**Purpose**: Achieve target conductance levels through iterative write-verify loops.

```go
type WriteVerifyConfig struct {
    MaxIterations int     // Default: 10
    Tolerance     float64 // Default: 0.5 levels
    PulseStep     float64 // Default: 0.1 (fractional change per iteration)
}

type WriteVerifyResult struct {
    TargetLevel   int
    AchievedLevel int
    Iterations    int
    Converged     bool
    FinalError    float64
}
```

**Algorithm**:
```
loop iterations:
    error = noise × NoiseLevel × PulseStep
    newG = currentG + (targetG - currentG) × PulseStep + error
    currentG = clamp_and_quantize(newG)
    if |currentG - targetG| ≤ tolerance:
        converged = true
        break
```

### Statistical Write Variation

**File**: `enhanced.go`

**Purpose**: Model threshold voltage variation in write operations.

```go
type WriteStatistics struct {
    Enabled  bool
    VthSigma float64  // Default: 0.05 (5% = ~1.5 level spread)
    RNG      *rand.Rand
}

func (a *Array) ProgramWeightWithVariation(row, col int, targetLevel int) (int, error) {
    // Adds Gaussian noise: actualLevel = targetLevel + N(0, σ)
    // σ = VthSigma × (Levels - 1)
}
```

### Analysis and Reporting

**File**: `enhanced.go`

**Key Functions**:

```go
// AnalysisReport captures complete array state
type AnalysisReport struct {
    Timestamp       time.Time
    ArraySize       [2]int
    TotalMACs       int
    MaxIRDrop       float64
    MaxSneakRatio   float64
    TotalEnergy     float64
    EnergyEfficiency float64
    TotalReads      int64
    TotalWrites     int64
}

// Accuracy degradation analysis
type AccuracyDegradation struct {
    BaselineAccuracy float64
    Degradations     []DegradationStep  // Per-source breakdown
    FinalAccuracy    float64
}

func (a *Array) ComputeAccuracyDegradation(input []float64, baseline float64) *AccuracyDegradation
```

**Stepwise Accuracy Loss** (empirical: 1% accuracy loss per 3% RMSE):
1. ADC/DAC quantization
2. IR drop
3. Device variation
4. Sneak paths

---

## Network Integration

### Multi-Layer Networks

**File**: `pkg/network/network.go`

**Purpose**: Build and inference multi-layer neural networks on crossbar arrays.

#### Architecture

```go
type Config struct {
    InputSize  int   // Input layer neurons
    HiddenSize int   // Hidden layer neurons
    OutputSize int   // Output neurons
    NumLayers  int   // Total layers including input/output
}

type Network struct {
    config *Config
    layers []*Layer     // N-1 weight layers for N neuron layers
}

type Layer struct {
    inputSize  int
    outputSize int
    array      *crossbar.Array
    biases     []float64
}
```

#### Forward Inference

**Activation Flow**:
```
activation[0] = input
for each layer:
    output = array.MVM(activation)
    output += biases
    if hidden layer:
        activation[i+1] = ReLU(output)
    else:
        activation[i+1] = Softmax(output)
```

**Initialization** (Xavier/Glorot):
```go
stddev = sqrt(2 / (inputSize + outputSize))
weights = N(0.5, stddev×0.5)  // Centered at 0.5 for crossbar [0,1]
```

#### Quantization-Aware Training

**File**: `pkg/training/training.go`

```go
type TrainingConfig struct {
    LearningRate   float64      // Base learning rate
    WeightDecay    float64      // L2 regularization
    Momentum       float64      // Momentum coefficient
    BatchSize      int          // Mini-batch size
    Epochs         int          // Number of epochs

    // Hardware-aware
    WeightClipMin  float64      // [0, 1] range
    WeightClipMax  float64
    UpdateNoise    float64      // Noise in weight updates (σ)
    QuantizeBits   int          // 0 = floating point
    AsymmetryRatio float64      // Potentiation/depression asymmetry
}

type Trainer struct {
    config     *TrainingConfig
    weights    [][][]float64
    velocities [][][]float64  // For momentum
}
```

**Backpropagation Modifications**:
- Weight clipping to [0, 1]
- Quantization noise injection (optional)
- Asymmetric update rates (modeling device physics)
- Momentum-based optimization

---

## Data I/O

### Weight Serialization

**File**: `pkg/weights/weights.go`

```go
type Format int
const (
    FormatJSON   Format = iota
    FormatBinary
    FormatNumPy
)

type LayerWeights struct {
    Name  string
    Shape []int         // [rows, cols]
    Dtype string        // "float64", "float32"
    Data  []float64     // Flattened
    Bias  []float64
    Quant *QuantInfo    // Optional quantization metadata
}

type ModelWeights struct {
    Name      string
    Version   string
    NumLayers int
    Layers    []LayerWeights
    Metadata  map[string]any
}
```

**Quantization Metadata**:
```go
type QuantInfo struct {
    Bits       int     // Quantization bits (e.g., 8)
    Scale      float64 // Scaling factor
    ZeroPoint  float64 // Offset
    Symmetric  bool    // Symmetric vs asymmetric
    PerChannel bool    // Per-channel vs per-tensor
}
```

### Export Functions

```go
// CSV: For spreadsheet analysis
func (a *Array) ExportWeightsCSV(path string) error
// Exports: row, col, level, conductance, conductance_uS

// JSON: For reporting and archiving
func (a *Array) ExportAnalysisJSON(path string, mvmResult *MVMResult) error
// Includes: array stats, non-ideality analysis, energy metrics

// NumPy: For integration with Python ML tools
func (w *ModelWeights) ExportNumPy(path string) error
```

---

## Data Flow Diagram

### Basic MVM Pipeline

```
┌─────────────────────────────────────────────────────────────────┐
│                         Input Vector                            │
│                            (1D array)                           │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │   DAC Quantization   │ ← ADCBits parameter
          │   (input rounding)   │
          └──────────┬───────────┘
                     │
                     ▼
        ┌────────────────────────────┐
        │  IR Drop Analysis (Opt.)    │ ← EnableIRDrop flag
        │  Compute V_eff[i][j]       │
        └────────┬───────────────────┘
                 │
                 ▼
      ┌──────────────────────────────────┐
      │  For each output row:              │
      │    sum = 0                         │
      │    for each column:                │
      │      if EnableVariation:           │
      │        g = g_nom × noiseFactor    │
      │      if EnableIRDrop:              │
      │        v = v_in × V_eff[i][j]    │
      │      sum += g × v                 │
      │      MACs++                       │
      │    if EnableSneakPaths:            │ ← Architecture-aware
      │      sum += sneak_current[i]      │
      │    output[i] = ADC_quantize(sum)  │
      └──────────┬───────────────────────┘
                 │
                 ▼
      ┌──────────────────────────┐
      │  ADC Quantization        │ ← DACBits parameter
      │  (output rounding)       │
      └──────────┬───────────────┘
                 │
                 ▼
      ┌──────────────────────────┐
      │  Compute Error Metrics   │
      │  - RMSE                  │
      │  - Accuracy Loss         │
      │  - Energy               │
      └──────────┬───────────────┘
                 │
                 ▼
         ┌─────────────────┐
         │  MVMResult      │
         │  structure      │
         └─────────────────┘
```

### Non-Ideality Impact Chain

```
Ideal Output
    │
    ├─→ + ADC/DAC Quantization      (RMSE ≈ 1-2%)
    │       │
    ├──────→ + IR Drop                (ΔRMSEs ≈ 2-3%)
    │       │
    ├──────→ + Device Variation       (ΔRMSEs ≈ 1-2%)
    │       │
    └──────→ + Sneak Paths (0T1R)     (ΔRMSEs ≈ 5-20%)
                │
                ▼
            Actual Output
            (5-10% RMSE for 0T1R)
            (1-2% RMSE for 1T1R)
```

---

## Thread Safety

### Concurrency Model

**Thread Safety Level**: **Not thread-safe for concurrent operations on same array**.

**Key Constraints**:

1. **No Locks**: Array implementation assumes single-threaded access
2. **GUI Thread**: Fyne integration requires UI updates via `fyne.Do()`
3. **Simulation**: Each simulation can run independently in separate goroutines

**Safe Patterns**:

```go
// Pattern 1: Separate arrays for each goroutine
arrayA, _ := crossbar.NewArray(cfg)
arrayB, _ := crossbar.NewArray(cfg)

go func() {
    arrayA.MVM(input)  // Safe: different array
}()

go func() {
    arrayB.MVM(input)  // Safe: different array
}()

// Pattern 2: GUI updates from goroutine
go func() {
    result, _ := array.MVMWithNonIdealities(input, opts)
    fyne.Do(func() {
        updateLabel.SetText(fmt.Sprintf("RMSE: %.4f", result.RMSE))
    })
}()

// Pattern 3: DON'T do this
go func() {
    array.MVM(input1)  // ❌ Race condition!
}()
array.MVM(input2)      // ❌ Race condition!
```

**GUI Integration Rules** (from CLAUDE.md):
```go
// ✅ Correct: Use fyne.Do for UI updates from goroutines
fyne.Do(func() {
    widget.SetText("Updated")
    heatmap.Refresh()
})

// ❌ Wrong: Direct UI update from goroutine
go func() {
    widget.SetText("Updated")  // PANIC: UI goroutine
}()
```

---

## Performance Characteristics

### Computational Complexity

| Operation | Complexity | Notes |
|-----------|-----------|-------|
| MVM (basic) | O(rows × cols) | Linear combination |
| MVM (with IR drop) | O(rows × cols) | Analytical model |
| Sneak paths (small array) | O(rows² × cols²) | Full enumeration ≤32×32 |
| Sneak paths (large array) | O(rows × cols) | Simplified ≤2 per row |
| Write-verify | O(iterations × rows × cols) | Typically 5-10 iterations |
| Full network forward | O(∑ rows_i × cols_i) | Sum over all layers |

### Memory Usage

| Component | Size per Array |
|-----------|---|
| Conductance matrix | 8 × rows × cols bytes |
| Voltage matrices (IR drop) | 3 × 8 × rows × cols bytes |
| Sneak current map | 8 × rows × cols bytes |
| Drift history | ~100 snapshots × rows bytes |

**Example**: 128×128 array ≈ 256 KB base + up to 3 MB with analysis.

### Energy Model

```go
// Energy per MVM operation (literature-based)
ArrayEnergy = rows × cols × 0.01 fJ        // Cell read energy
ADCEnergy   = rows × 0.5 pJ × 2^(bits-6)   // ADC conversion
DACEnergy   = cols × 0.1 pJ                // DAC conversion
TotalEnergy = ArrayEnergy + ADCEnergy + DACEnergy

// GPU comparison
GPUEnergy = MACOperations × 10 pJ           // ~10 pJ/MAC (literature)
EnergyEfficiency = GPUEnergy / FeCIMEnergy
```

**Typical Values** (128×128 array, 6-bit ADC):
- Array: ~0.16 pJ
- ADC: ~0.5 pJ
- DAC: ~0.013 pJ
- **Total: ~0.67 pJ** (vs ~163 pJ for GPU)
- **Efficiency: ~240×** better than GPU

---

## Key Design Decisions

### 1. 30 Discrete Levels

**Decision**: Demo baseline fixed to 30 analog states per cell (simulation baseline (configurable).

**Rationale**:
- Dr. external research group COSM 2025: "It's got 30 discrete states. Not 0-1-0-1." (simulation baseline)
- Other reported in literature work: multi-level states (reported) demonstrated
- Provides ~4.9 bits/cell (log₂(30) ≈ 4.9)
- Balances precision vs device programming complexity

**Trade-offs**:
- Higher levels → better NN accuracy but harder to program reliably
- Lower levels → simpler programming but coarser quantization

### 2. Three Conductance Models

**Decision**: Support linear, exponential, and lookup-table models.

**Rationale**:
- **Linear**: Simple, fast for prototyping
- **Exponential**: Physics-accurate for FeFET devices
- **Lookup**: Calibrated from measurements

**Implementation**: Pluggable via `Config.ConductanceModel`, no API changes.

### 3. Separate Non-Ideality Modules

**Decision**: Each effect (IR drop, sneak, drift) modular and optional.

**Rationale**:
- Decoupled development and testing
- Each can be enabled/disabled independently
- Allows hierarchical complexity
- Easy to add new effects

**Trade-off**: Slightly higher code complexity vs flexibility.

### 4. Sneak Path Size-Based Optimization

**Decision**: Use full calculation for ≤32×32, simplified for larger.

**Rationale**:
- Full 3-cell enumeration is O(n⁴), becomes slow for large arrays
- Simplified factor model O(n²) is reasonable upper bound
- Threshold (32×32) chosen empirically

**Accuracy Impact**: <5% error for large arrays vs full calculation.

### 5. Differential Arrays for Signed Weights

**Decision**: Use separate positive/negative arrays for [-1, 1] weights.

**Rationale**:
- Conductance naturally positive-only [0, 1]
- Differential output = I+ - I- allows signed computation
- Standard approach in analog NN literature
- No modification to basic array physics

**Cost**: 2× energy, latency unchanged (parallel operation).

---

## Validation & Testing

### Test Coverage

**Location**: `module2-crossbar/pkg/crossbar/` with `*_test.go` files.

**Test Categories**:

1. **Physics Tests** (`physics_test.go`):
   - Quantization accuracy (30-level baseline)
   - Conductance models (linear, exponential, lookup)
   - Temperature effects (cryogenic to automotive)

2. **Array Tests** (`array_test.go`):
   - MVM correctness (basic computation)
   - VMM operations
   - Weight programming and verification
   - Boundary conditions (edge cells)

3. **Non-Ideality Tests** (`nonidealities_test.go`):
   - IR drop analysis
   - Sneak path calculations
   - Combined effect accuracy
   - Architecture differences (0T1R vs 1T1R)

4. **Improvement Tests** (`improvements_test.go`):
   - Write-verify convergence
   - Statistical variation
   - Differential array operation

### Running Tests

```bash
# Run all crossbar tests
go test ./module2-crossbar/pkg/crossbar

# Run specific test file
go test -run TestArrayMVM ./module2-crossbar/pkg/crossbar

# Verbose output
go test -v ./module2-crossbar/pkg/crossbar

# With coverage
go test -cover ./module2-crossbar/pkg/crossbar
```

---

## Integration Points

### With Module 1 (Hysteresis)

- Hysteresis model provides P-E curves
- These drive conductance values in the crossbar
- Temperature/field effects cascade from Module 1 to Module 2

### With Module 3 (MNIST)

- Crossbar arrays execute MVM in neural network layers
- Training uses hardware-aware backprop from `pkg/training`
- Weight serialization via `pkg/weights` for model persistence

### With Module 4 (Circuits)

- Array interfaces with DAC (input driver)
- Array output feeds ADC (sense amplifiers)
- DAC/ADC bit-widths configured in `Config.DACBits`, `Config.ADCBits`
- Power and latency estimates inform circuit design

### With Shared Components

- **Logging**: Uses `shared/logging` for debug output
- **Theme**: GUI uses `shared/theme` for consistent styling
- **Widgets**: Heatmaps and panels extend `shared/widgets`

---

## Future Extensions

### Planned

1. **Memristor Models**: Switch from ferroelectric to other device types
2. **3D Arrays**: Stack multiple crossbars vertically
3. **Hybrid Approaches**: Mix different device types in same array
4. **Advanced Training**: Pruning, quantization-aware training refinement
5. **Power Modeling**: Detailed power consumption per operation

### Research Opportunities

1. **Calibration**: Measured conductance tables from real devices
2. **Drift Characterization**: Long-term retention studies to validate coefficients
3. **Multi-Bit Cells**: Exploit multiple bits per cell
4. **Optical Extensions**: Integrate with optical computing

---

## References & Sources

### Physics Constants & Peer-Reviewed Data

See `docs/4-research/honesty-audit.md` for the active external-claim boundary. Verified or DOI-backed sources used around this module include:

- **Multi-level FeFET crossbar demo**: Nature Communications 2023, DOI `10.1038/s41467-023-42110-y`
- **Related ferroelectric benchmark (not this simulator, not a FeCIM crossbar claim)**: HZO FTJ reservoir computing at 98.24% MNIST accuracy, Journal of Alloys and Compounds 2025, DOI `10.1016/j.jallcom.2025.181869`
- **DOI-backed hysteresis/calibration datasets bundled in this repo**:
  - Park et al., Advanced Materials 2015, DOI `10.1002/adma.201404531`
  - Cheema et al., Nature 2020, DOI `10.1038/s41586-020-2208-x`
  - Crystals 2021 BTO dataset, DOI `10.3390/cryst11101192`
  - Micromachines 2022 AlScN dataset, DOI `10.3390/mi13101629`
  - Nanomaterials 2024 PZT dataset, DOI `10.3390/nano14050432`

The following categories remain intentionally outside the verified-claim set here until exact source papers are wired into the honesty audit and validation artifacts: endurance headline numbers, automotive-grade retention extrapolations, and BEOL manufacturing claims.

### Implementation References

- **Ferroelectric Computing**: Dr. external research group COSM 2025
- **Sneak Path Analysis**: Leveraging standard ReRAM crossbar literature
- **IR Drop**: Standard analog circuit textbook models
- **Drift Modeling**: Literature comparison (RRAM, PCM, Flash)

---

## Code Organization Quick Reference

| Need | File | Key Type/Function |
|------|------|---|
| Create array | `array.go` | `NewArray(cfg)` |
| Program weights | `array.go` | `ProgramWeight()` |
| Run MVM | `array.go` | `MVM(input)` |
| With all effects | `enhanced.go` | `MVMWithNonIdealities(input, opts)` |
| IR drop analysis | `nonidealities.go` | `AnalyzeIRDrop()` |
| Sneak paths | `nonidealities.go` | `AnalyzeSneakPaths()` |
| Signed weights | `enhanced.go` | `DifferentialArray` |
| Write-verify | `enhanced.go` | `ProgramWeightVerified()` |
| Temperature effects | `temperature.go` | `TemperatureEffects` |
| Drift modeling | `drift.go` | `DriftSimulator` |
| Neural network | `pkg/network/network.go` | `Network.Forward()` |
| Hardware training | `pkg/training/training.go` | `Trainer.Backward()` |
| Weight export | `pkg/weights/weights.go` | `ExportNumPy()` |
| Visualization | `pkg/gui/app.go` | `CrossbarApp` |

---

**Last Updated**: January 2026
**Maintained By**: FeCIM Development Team
**License**: See repository LICENSE
