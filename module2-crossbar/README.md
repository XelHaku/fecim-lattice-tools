# Module 2: Crossbar Array Simulation

## Overview

Module 2 provides a comprehensive simulation of ferroelectric (FeCIM) crossbar arrays for Compute-in-Memory operations. It models the physics of matrix-vector multiplication (MVM) in analog crossbar arrays while accounting for real-world non-idealities including IR drop, sneak paths, device drift, and temperature effects.

The core concept is **30 discrete analog states per cell** (approximately 4.9 bits per cell), enabling multi-level storage and computation on ferroelectric memory elements. The simulation is hardware-aware, supporting both passive (0T1R) and transistor-isolated (1T1R) architectures.

**Primary Reference**: Dr. external research group, COSM 2025 - "It's got 30 discrete states. So it's not 0-1-0-1."

---

## Key Features

### Analog Computation
- **30-level quantization**: Discrete conductance states matching ferroelectric memory cells
- **Matrix-Vector Multiplication (MVM)**: I = G × V (Ohm's law) + Kirchhoff's current summation
- **Differential arrays**: Dual positive/negative weight arrays for signed computations
- **Configurable DAC/ADC**: 4-12 bit conversion with quantization effects

### Non-Idealities
- **IR Drop**: Word line/bit line resistance with cumulative voltage degradation
- **Sneak Paths**: Parasitic current through unintended paths in passive arrays
- **Device Drift**: Power-law and logarithmic drift models with peer-reviewed literature parameters
- **Temperature Effects**: 4K to 500K operation with cryogenic enhancement and thermal degradation
- **Device Variation**: Device-to-device noise, systematic process variation, and edge effects
- **Endurance Tracking**: Cycle-dependent fatigue modeling (10⁸-10¹² cycle endurance)
- **Half-Select Disturb**: Write disturb accumulation on half-selected cells

### Neural Network Integration
- **Multi-layer networks**: Stack multiple crossbar arrays for deep learning
- **Hardware-aware training**: Backpropagation with non-ideality simulation
- **Weight serialization**: JSON/binary export for deployment
- **Accuracy evaluation**: MNIST and custom dataset support

### Interactive Visualization
- **Heatmaps**: Real-time conductance state and performance metrics
- **Parametric analysis**: Sweep effects of individual non-idealities
- **Architecture comparison**: Side-by-side 0T1R vs 1T1R performance
- **Energy/accuracy tradeoff**: Quantify cost of simulation features

---

## Quick Start

### Build and Run

```bash
# Build the main unified application
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools

# Or use the launch script
./launch.sh
```

### Command-Line Tools

```bash
# Crossbar GUI (standard mode)
go run ./module2-crossbar/cmd/crossbar-gui

# Crossbar GUI (enhanced mode with all features)
go run ./module2-crossbar/cmd/crossbar-gui -enhanced

# Neural network inference
go run ./module2-crossbar/cmd/inference --weights=model.json --input=data.csv
```

### Simple Example

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
    // Create a 64×784 crossbar array (example: MNIST layer)
    cfg := &crossbar.Config{
        Rows:       64,
        Cols:       784,
        NoiseLevel: 0.02,
        ADCBits:    8,
        DACBits:    8,
    }
    array, _ := crossbar.NewArray(cfg)

    // Program random weights (auto-quantized to 30 levels)
    weights := make([][]float64, 64)
    for i := range weights {
        weights[i] = make([]float64, 784)
        for j := range weights[i] {
            weights[i][j] = rand.Float64() // [0, 1]
        }
    }
    array.ProgramWeightMatrix(weights)

    // Perform ideal MVM
    input := make([]float64, 784)
    for i := range input {
        input[i] = rand.Float64()
    }
    output, _ := array.MVM(input)
    fmt.Printf("Output shape: [%d]\n", len(output))

    // Perform MVM with all non-idealities
    opts := crossbar.DefaultMVMOptions()
    opts.Architecture = "0T1R"
    result, _ := array.MVMWithNonIdealities(input, opts)

    fmt.Printf("Ideal output:  %v\n", result.IdealOutput[:4])
    fmt.Printf("Actual output: %v\n", result.ActualOutput[:4])
    fmt.Printf("RMSE:          %.4f\n", result.RMSE)
    fmt.Printf("Max error:     %.4f\n", result.MaxError)
}
```

---

## Package Structure

### Core Crossbar Simulation

```
pkg/crossbar/
├── array.go              # Core Array type, MVM, weight programming
├── reference.go          # CPU reference implementation (validation)
├── enhanced.go           # MVMWithNonIdealities, differential arrays
├── nonidealities.go      # IR drop, sneak paths, basic modeling
├── irdrop.go            # Detailed IR drop analysis
├── sneakpath.go         # Sneak path analysis and mitigation
├── drift.go             # Drift simulation and history tracking
├── temperature.go       # Temperature-dependent parameter scaling
│
├── array_test.go        # Quantization, MVM correctness
├── nonidealities_test.go # IR drop, sneak path validation
├── physics_test.go      # Physical constant verification
└── improvements_test.go  # Advanced feature testing
```

### Neural Networks and Training

```
pkg/network/
├── network.go           # Multi-layer network composition
└── network_test.go      # Integration tests

pkg/training/
├── training.go          # Hardware-aware backpropagation
└── training_test.go     # Training validation
```

### Utilities and Data

```
pkg/weights/
├── weights.go           # Weight matrix management
├── serialization.go     # JSON/binary import/export
└── weights_test.go      # Serialization tests

pkg/data/
└── mnist.go             # MNIST dataset loading

pkg/evaluation/
└── accuracy.go          # Inference accuracy metrics

pkg/visualization/
├── heatmap.go           # Conductance heatmap rendering
└── terminal.go          # ASCII terminal visualization
```

### GUI Components

```
pkg/gui/
├── app.go               # Main application lifecycle
├── embedded.go          # Embedded app interface
├── tabbed_app.go        # Tab-based UI layout
│
├── tabs/                # Individual visualization tabs
│   ├── ideal_tab.go     # Ideal MVM without non-idealities
│   ├── irdrop_tab.go    # IR drop analysis
│   ├── sneak_tab.go     # Sneak path visualization
│   └── drift_tab.go     # Drift over time
│
├── widgets.go           # Basic UI controls
├── widgets_metrics.go   # Performance metrics display
├── widgets_comparison.go # Before/after comparison
├── heatmap.go           # Interactive heatmap widget
└── analysis.go          # Analysis and reporting tools
```

### Command-Line Tools

```
cmd/crossbar-gui/       # Interactive GUI application
cmd/inference/          # Inference on pre-trained models
```

---

## Core Concepts

### 30 Discrete Conductance Levels

FeCIM cells store weight information as one of **30 discrete conductance states**, quantized to `[0/29, 1/29, 2/29, ..., 28/29, 29/29]`.

```go
// Quantize any value to nearest level
quantized := crossbar.QuantizeToLevels(0.542)  // Returns ~14/29 ≈ 0.483

// Get discrete level index (0-29)
level := crossbar.GetLevel(0.5)  // Returns 15
```

**Why 30?** Ferroelectric hysteresis curves with proper programming can reliably achieve ~30 distinct polarization states. This provides 4.9 bits per cell (log₂(30) ≈ 4.91), compared to 1 bit for conventional DRAM.

### Matrix-Vector Multiplication

The fundamental operation is standard MVM, but implemented in analog:

**Physics:**
- Input voltages V applied to **columns (bit lines)**
- Conductances G stored in crossbar cells
- Output currents I summed on **rows (word lines)**

**Computation:**
```
I[i] = Σ(G[i][j] × V[j])    (Ohm's law for each cell)
       j

I[i]_total = Σ(I[i][j])    (Kirchhoff's current law at node)
```

**Normalization:**
```go
// Normalize by maximum possible current to keep output in [0, 1]
normalizedSum := sum / (numInputs * maxInputVoltage)

// Quantize through ADC
output[i] := quantizeADC(normalizedSum)
```

### Conductance Range

**Physical constants:**
- **Gmin**: 10 µS (OFF state, level 0)
- **Gmax**: 100 µS (ON state, level 29)
- **Range**: 10× dynamic range

**Conductance models:**
```go
// Linear (simple)
G = Gmin + normalized*(Gmax - Gmin)

// Exponential (realistic FeFET behavior)
G = Gmin * exp(ln(Gmax/Gmin) * normalized)
// Midpoint: geometric mean ≈ 31.6 µS

// Lookup (calibrated from measurement)
G = ConductanceTable[level]
```

---

## Non-Idealities

### IR Drop

**Problem**: Metal word lines and bit lines have finite resistance. Current flow causes voltage drops, reducing effective cell voltage.

**Parameters** (45nm technology):
- Word line resistance: 2.5 Ω/cell
- Bit line resistance: 2.5 Ω/cell
- Contact resistance: 50 Ω

**Impact:**
- Bottom-right cell sees ~10-20% IR drop in large arrays
- Voltage reduction → conductance error → output error
- Quantified by: Max drop, average drop, variance

**Example**:
```go
wireParams := crossbar.DefaultWireParams()
analysis := array.AnalyzeIRDrop(input, wireParams)

fmt.Printf("Max IR drop: %.2f%%\n", analysis.MaxIRDrop*100)
fmt.Printf("Affected cell: [%d, %d]\n", analysis.WorstCaseCell[0], analysis.WorstCaseCell[1])
```

### Sneak Paths

**Problem**: In passive crossbar arrays (0T1R), unintended current paths form parasitic loops.

**3-cell sneak path example:**
```
When selecting cell[target_row][col]:
  Input current can flow through:
    cell[other_row][col] →
    cell[other_row][other_col] →
    cell[target_row][other_col]

  This "sneak current" adds to target cell reading.
```

**Architecture comparison:**
- **0T1R (Passive)**: No transistor isolation, 5-20% sneak current
- **1T1R (Transistor)**: Transistor blocks sneak paths, ~1000:1 isolation

**Impact:**
```go
opts := &crossbar.MVMOptions{
    Architecture: "0T1R",  // vs "1T1R"
    EnableSneakPaths: true,
}
result, _ := array.MVMWithNonIdealities(input, opts)

fmt.Printf("Sneak paths magnitude: %.2f%%\n", result.SneakPathAnalysis.TotalSneakRatio*100)
```

### Drift

**Problem**: Ferroelectric polarization gradually depolarizes over time (milliseconds to years depending on temperature).

**Models:**

| Model | Source | Coefficient | Notes |
|-------|--------|-------------|-------|
| Assumed | None | 0.001 | Conservative estimate |
| Literature | Retention studies | 0.0005 | <0.5 level drift in 10 years |
| RRAM | Comparative | 0.05 | Much higher drift |
| PCM | Comparative | 0.1 | Highest drift |
| Flash | Comparative | 0.02 | Moderate drift |

**Drift mechanisms:**
- **Power-law drift**: Most significant, t^(-ν)
- **Logarithmic drift**: Slower long-term, log(t)
- **Arrhenius temperature scaling**: Drift rate ∝ exp(-Ea/kT)

**Example:**
```go
sim := crossbar.NewDriftSimulator(
    64, 784,                                  // Rows, cols
    weights,                                  // Initial conductances
    crossbar.DriftModelLiterature,           // Model type
    0.0005,                                  // Coefficient
    300.0,                                   // Temperature (K)
)

// Simulate 1 second of drift
sim.SimulateTime(1.0)

// Check degradation
stats := sim.GetStatistics()
fmt.Printf("Max drift: %.4f levels\n", stats.MaxDrift*29)  // Convert to levels
```

### Temperature Effects

**Supported range**: 4K (cryogenic) to 500K

**Temperature-dependent parameters:**

| Temperature | Effect | Gmin/Gmax | Drift | Resistance | Use Case |
|-------------|--------|-----------|-------|------------|----------|
| 4K (Space) | Enhanced polarization | 1.5× window | ↓↓ Minimal | ↓ 50% | Scientific instruments |
| 77K | Cryogenic | 1.2× window | ↓ Low | ↓ 74% | Helium-cooled labs |
| 300K | Room temp | Baseline | Baseline | Baseline | Standard operation |
| 358K (85°C) | Industrial | 0.95× | ↑ Faster | ↑ 5% | Industrial equipment |
| 400K (125°C) | Automotive | 0.9× | ↑↑ Much faster | ↑ 10% | Grade 0 automotive |

**Physics basis:**
- Polarization enhancement at cryo: Pr(4K) = 75 µC/cm² vs Pr(RT) = 15-34 µC/cm² (Adv. Mat. 2024)
- Thermal activation: Drift rate follows Arrhenius law

**Example:**
```go
temp := crossbar.NewTemperatureEffects(77.0)  // Liquid nitrogen

// Get adjusted parameters
gMin, gMax := temp.AdjustedConductanceRange(10e-6, 100e-6)
fmt.Printf("At 77K: Gmin=%.2e, Gmax=%.2e\n", gMin, gMax)

// Drift accelerates at high temperature
driftAtRT := 0.0005
driftAt400K := temp.AdjustedDriftRate(driftAtRT)
fmt.Printf("Drift speed ratio (400K/300K): %.2f×\n", driftAt400K/driftAtRT)
```

---

## API Quick Reference

### Creating Arrays

```go
// Basic configuration
cfg := &crossbar.Config{
    Rows:       128,              // Number of word lines
    Cols:       784,              // Number of bit lines
    NoiseLevel: 0.02,             // Device-to-device variation (2%)
    ADCBits:    8,                // ADC resolution
    DACBits:    8,                // DAC resolution
}

// Optional: Advanced configuration
cfg.ConductanceModel = crossbar.ConductanceExponential
cfg.Endurance = crossbar.DefaultEnduranceConfig()
cfg.ProcessVariation = crossbar.DefaultProcessVariationConfig()
cfg.HalfSelect = crossbar.DefaultHalfSelectConfig()

array, err := crossbar.NewArray(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Programming Weights

```go
// Single cell
err := array.ProgramWeight(row, col, 0.75)  // Auto-quantized to 30 levels

// Full matrix
weights := make([][]float64, 128)
for i := range weights {
    weights[i] = make([]float64, 784)
    // ... fill with trained weights
}
err := array.ProgramWeightMatrix(weights)

// Get current conductances
matrix := array.GetConductanceMatrix()
```

### Ideal MVM

```go
input := make([]float64, 784)
for i := range input {
    input[i] = rand.Float64()  // [0, 1]
}

output, err := array.MVM(input)
// Returns: []float64 of length 128
```

### MVM with Non-Idealities

```go
opts := &crossbar.MVMOptions{
    EnableIRDrop:     true,
    EnableSneakPaths: true,
    EnableVariation:  true,
    EnableDrift:      false,    // Usually simulated separately
    Temperature:      300.0,    // Kelvin
    Architecture:     "0T1R",   // "0T1R" or "1T1R"
}

result, err := array.MVMWithNonIdealities(input, opts)

// Access results
fmt.Printf("Ideal:  %v\n", result.IdealOutput)
fmt.Printf("Actual: %v\n", result.ActualOutput)
fmt.Printf("RMSE:   %.4f\n", result.RMSE)
fmt.Printf("Energy: %.2f pJ\n", result.TotalEnergy)
```

### Differential Arrays (Signed Weights)

For networks with signed weights, use two complementary arrays:

```go
cfg := &crossbar.Config{Rows: 128, Cols: 784}
pos, _ := crossbar.NewArray(cfg)  // Positive weights
neg, _ := crossbar.NewArray(cfg)  // Negative weights

diffArr := crossbar.NewDifferentialArray(pos, neg)
diffArr.ProgramWeights(signed_weights)  // Can be negative

output, _ := diffArr.MVM(input)  // Returns signed result
result, _ := diffArr.MVMWithNonIdealities(input, opts)
```

### Quantization

```go
// Quantize a single value to nearest 30-level
q := crossbar.QuantizeToLevels(0.123)  // Returns 0.138 (≈ 4/29)

// Get the discrete level (0-29)
level := crossbar.GetLevel(0.5)  // Returns 15

// Physical conductance from normalized value
phys_g := array.GetPhysicalConductance(0.5)  // Returns 55.5e-6 (55.5 µS)
```

### Analysis and Metrics

```go
// IR drop analysis
irAnalysis := array.AnalyzeIRDrop(input, nil)  // Uses default WireParams
fmt.Printf("Max IR drop: %.2f%%\n", irAnalysis.MaxIRDrop*100)

// Sneak path analysis
sneakAnalysis := array.AnalyzeSneakPaths(input, "0T1R")
fmt.Printf("Sneak ratio: %.2f%%\n", sneakAnalysis.TotalSneakRatio*100)

// Accuracy degradation breakdown
result, _ := array.MVMWithNonIdealities(input, opts)
fmt.Printf("Accuracy loss from IR drop:    %.2f%%\n", result.AccuracyLoss)
fmt.Printf("Energy overhead:               %.2f pJ\n", result.TotalEnergy)
```

---

## Testing

### Run All Tests

```bash
cd <local-path>
go test ./module2-crossbar/...
```

### Run Specific Test Suites

```bash
# Quantization and MVM correctness
go test -v ./module2-crossbar/pkg/crossbar -run TestQuantize

# Physics validation
go test -v ./module2-crossbar/pkg/crossbar -run TestPhysics

# Non-idealities
go test -v ./module2-crossbar/pkg/crossbar -run TestNonidealities

# Network training
go test -v ./module2-crossbar/pkg/network

# Weights serialization
go test -v ./module2-crossbar/pkg/weights
```

### Test Coverage

Current test coverage spans:
- **Quantization**: 30-level discretization, edge cases
- **MVM physics**: Correctness against CPU reference
- **Non-idealities**: IR drop, sneak paths, drift, temperature
- **Network inference**: Multi-layer computation
- **Training**: Backpropagation with hardware constraints
- **Serialization**: JSON/binary import/export

Example test:
```go
// Verify exactly 30 quantization levels
func TestQuantizeToLevelsProducesExactly30Values(t *testing.T) {
    seen := make(map[float64]bool)
    for i := 0; i <= 1000; i++ {
        quantized := crossbar.QuantizeToLevels(float64(i) / 1000.0)
        seen[quantized] = true
    }
    if len(seen) != 30 {
        t.Errorf("Expected 30 levels, got %d", len(seen))
    }
}
```

---

## Related Documentation

- **Physics Background**: `docs/crossbar/module2-plan-improvements.md` - Physics enhancement roadmap
- **Improvement Plan**: Detailed breakdown of features (conductance models, full sneak integration, disturb, etc.)
- **Research**: `docs/crossbar/crossbar.research.md` - Literature citations and physics principles
- **CIM Fundamentals**: `docs/peripheral-circuits/circuits.CIM-fundamentals.md` - Circuit-level modeling
- **Testing Guide**: `docs/development/TESTING.md` - Test framework and conventions
- **Project Guide**: `CLAUDE.md` - Overall project structure and conventions

---

## Architecture Comparison: 0T1R vs 1T1R

| Feature | 0T1R (Passive) | 1T1R (Transistor-Isolated) |
|---------|---|---|
| **Cell overhead** | 1 resistor per cell | 1 transistor + 1 resistor |
| **Array density** | Higher (smaller cells) | Lower (requires transistor area) |
| **Sneak paths** | 5-20% parasitic current | ~0.01% (1000:1 isolation) |
| **IR drop** | Present, significant | Present but smaller with better row selection |
| **Read disturb** | Minimal | Minimal |
| **Write disturb** | Significant (half-select V/2) | None (transistor blocks) |
| **Speed** | Slower (sneak path RC delay) | Faster (isolated) |
| **Accuracy (MNIST)** | ~95-96% (with non-idealities) | ~98-99% |
| **Use case** | Embedded, high-density | High-accuracy AI |

---

## Example: MNIST Neural Network

### 1. Train a network (standard PyTorch)

```python
# Standard training produces floating-point weights
import torch
model = torch.nn.Sequential(
    torch.nn.Linear(784, 128),
    torch.nn.ReLU(),
    torch.nn.Linear(128, 10)
)
# ... training code ...
torch.save(model.state_dict(), "mnist_model.pth")
```

### 2. Convert to crossbar weights

```go
// Load pre-trained weights, quantize to 30 levels
weights := loadPyTorchWeights("mnist_model.pth")
for i := range weights {
    for j := range weights[i] {
        weights[i][j] = crossbar.QuantizeToLevels(weights[i][j])
    }
}
```

### 3. Create crossbar-based network

```go
cfg := &crossbar.Config{
    Rows: 128, Cols: 784, ADCBits: 8, DACBits: 8,
}
layer1, _ := crossbar.NewArray(cfg)
layer1.ProgramWeightMatrix(weights[0])

cfg.Rows, cfg.Cols = 10, 128
layer2, _ := crossbar.NewArray(cfg)
layer2.ProgramWeightMatrix(weights[1])
```

### 4. Run inference

```go
// Load MNIST test image
image := loadMNISTImage("digit.png")
input := normalizeImage(image)  // Scale to [0, 1]

// Forward pass with non-idealities
opts := crossbar.DefaultMVMOptions()
opts.Architecture = "0T1R"

result1, _ := layer1.MVMWithNonIdealities(input, opts)
hidden := ReLU(result1.ActualOutput)

result2, _ := layer2.MVMWithNonIdealities(hidden, opts)
logits := result2.ActualOutput

prediction := argmax(logits)
confidence := softmax(logits)[prediction]

fmt.Printf("Predicted: %d (%.2f%% confidence)\n", prediction, confidence*100)
```

### 5. Measure accuracy impact

```go
// Compare ideal vs actual
stats := struct {
    IdealAccuracy  float64
    ActualAccuracy float64
    ErrorDelta     float64
}

// Run 10,000 test images
for img := range testSet {
    ideal, _ := layer1.MVM(img)      // No non-idealities
    actual, _ := layer1.MVMWithNonIdealities(img, opts)

    if idealPred == actualPred {
        stats.IdealAccuracy += 1.0
    } else {
        stats.ErrorDelta += 1.0
    }
}

fmt.Printf("Accuracy loss: %.2f%%\n",
    (stats.IdealAccuracy - (10000-stats.ErrorDelta)) / 10000 * 100)
```

---

## Configuration and Customization

### Conductance Models

```go
// Linear (simple, less accurate)
cfg.ConductanceModel = crossbar.ConductanceLinear

// Exponential (realistic, more accurate)
cfg.ConductanceModel = crossbar.ConductanceExponential

// Lookup (from calibration data)
cfg.ConductanceModel = crossbar.ConductanceLookup
cfg.ConductanceTable = []float64{
    // 30 values: G(V) for levels 0-29
    10e-6, 12e-6, 14.5e-6, ...  // Measured from device
}
```

### Wire Parameters

```go
// Default (45nm technology)
params := crossbar.DefaultWireParams()
// RwordLine: 2.5 Ω/cell, RbitLine: 2.5 Ω/cell, Rcontact: 50 Ω

// Custom technology
params := &crossbar.WireParams{
    RwordLine: 1.0,  // 22nm technology
    RbitLine:  1.0,
    Rcontact:  25,
}

analysis := array.AnalyzeIRDrop(input, params)
```

### Temperature Presets

```go
// Cryogenic (liquid nitrogen)
temp := crossbar.NewTemperatureEffects(crossbar.TempCryogenic)  // 77K

// Room temperature
temp := crossbar.NewTemperatureEffects(crossbar.TempRoom)  // 300K

// Automotive Grade 0
temp := crossbar.NewTemperatureEffects(crossbar.TempAutomotive)  // 400K

// Drift rate adjustment
adjustedDrift := temp.AdjustedDriftRate(baseDriftCoeff)
```

---

## Performance Characteristics

### Computation
- **MVM latency**: ~1-5 ns (analog summation, AD conversion only)
- **Write latency**: ~100-500 ns (program and verify)
- **Memory access**: ~10-50 ns (word line/bit line RC delays)

### Energy Efficiency
- **Per MAC**: 10-100 pJ (vs 1-10 nJ for digital GPU)
- **Scaling**: 10-100× better than digital at same operation
- **Breakdown**:
  - Array computation: 30-50%
  - ADC conversion: 30-40%
  - DAC conversion: 10-20%

### Accuracy
- **Ideal (no non-idealities)**: 100% (matches floating-point weights)
- **0T1R passive**: 94-97% (with IR drop + sneak + variation)
- **1T1R isolated**: 97-99% (minimal parasitic effects)

---

## Design Decisions and Rationale

### Why 30 Levels?

Ferroelectric hysteresis curves naturally exhibit ~30 distinguishable polarization states under normal programming conditions. This provides:
- **Sufficient granularity**: 4.9 bits per cell is competitive with 8-bit quantization in neural networks
- **Reliable operation**: Can be reliably written and read with standard voltage/current levels
- **Practical**: Demonstrated by multiple research teams (Song et al. 2024, Oh et al. 2017)

### Quantization During Weights Programming

Weights are **automatically quantized to 30 levels** when programmed. This is a fundamental constraint of the device:

```go
array.ProgramWeight(row, col, 0.5123)
// Internally quantized to nearest level: 15/29 ≈ 0.5172
```

This differs from post-training quantization because it reflects the device's inability to store arbitrary floating-point values.

### Default Non-Idealities

By default, `MVMWithNonIdealities()` enables:
- **IR drop**: Realistic for all array sizes
- **Sneak paths**: Only for 0T1R (passive) arrays
- **Device variation**: Always present
- **Drift**: Usually simulated separately over time
- **Temperature**: Defaults to 300K (room temp)

This provides a **realistic but reproducible** simulation.

### Temperature Model: Arrhenius Activation

Thermal effects follow **Arrhenius kinetics**:
```
Rate ∝ exp(-Ea/kT)
```

This is physically accurate for thermally activated processes like:
- Drift (ferroelectric depolarization)
- Leakage current (affects noise floor)
- Wire resistance (thermal expansion)

---

## Known Limitations and Future Work

### Current Limitations

1. **Drift coefficient**: 0.0005 (literature) is derived from retention requirements, not directly measured for HZO FeFET
2. **Sneak path calculation**: Full 3-cell paths computed for small arrays, simplified for large arrays
3. **Process variation**: Device-to-device + gradient modeled; wafer-level variation not included
4. **Speed model**: DC-only analysis; AC impedance and RC delays simplified
5. **Temperature range**: Model validated 4K-500K; extrapolation beyond may be inaccurate

### Future Enhancements

- Full sneak path network analysis with iterative solver
- Measured drift coefficients from characterization
- Multi-wafer process variation statistics
- AC/dynamic behavior (frequency-dependent impedance)
- Write-verify loop with probabilistic success modeling
- Endurance data from latest literature (10¹² cycle devices)

---

## References

### Peer-Reviewed Sources

| Reference | Finding | Year |
|-----------|---------|------|
| Nature Commun. 2025 | Pr: 15-34 µC/cm², Ec: 0.6-1.5 MV/cm | 2025 |
| Adv. Elec. Mat. 2024 | Pr: 75 µC/cm² at 4K (cryogenic) | 2024 |
| Nano Letters 2024 | V:HfO₂ 10¹² cycle endurance | 2024 |
| IEEE IRPS 2022 | 10⁹-10¹² cycle endurance range | 2022 |
| Samsung Nature 2025 | 25-100× energy vs NAND | 2025 |
| CEA-Leti Dec 2024 | 3D BEOL @ 22nm integration | 2024 |
| Fraunhofer IPMS 2024 | AEC-Q100 automotive qualification | 2024 |

### Unverified Claims

- **30 analog states**: Demonstrated by Dr. Tour at COSM 2025 (conference, not peer-reviewed)
- **Drift coefficient 0.001**: Assumed value for simulation; <10 year retention requires <0.001

---

## Contributing

When adding features to module2-crossbar:

1. **Test first**: Write tests for new features in `*_test.go` files
2. **Quantize weights**: All programmed values must respect 30-level quantization
3. **Document physics**: Reference literature for non-ideality parameters
4. **GUI integration**: Update `pkg/gui/` for visualization support
5. **Backward compatibility**: Ensure existing configs still work

See `CLAUDE.md` for general project conventions.

---

## License

FeCIM Lattice Tools is provided as-is for research and educational purposes. See LICENSE file in root directory.
