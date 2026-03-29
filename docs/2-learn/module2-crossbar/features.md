# Module 2: Crossbar - Features

**Navigation:** [← Module 2 Index](./README.md) | [Physics](./physics.md) | [Architecture](./architecture.md) | [Tools](./tools.md)

---

## Evidence Status

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

---

## What This Module Does

Module 2 simulates crossbar array operations with:

1. **Analog Matrix-Vector Multiply (MVM)** with 30-level discrete conductance
2. **Comprehensive non-ideality simulation:**
   - IR drop (wire resistance)
   - Sneak paths (passive crossbar parasitic currents)
   - Device variation (random + systematic)
   - Conductance drift (time-dependent)
   - Temperature effects (77K to 400K)
3. **Differential arrays** for signed weights [-1, +1]
4. **Write-verify programming** with iterative convergence
5. **Multi-layer neural networks** with hardware-aware training
6. **Extensive visualization:** Heatmaps, IR drop maps, sneak current analysis

---

## Primary Components

### Core Array (`pkg/crossbar/`)
- `array.go` - Array structure, MVM, quantization
- `nonidealities.go` - Enhanced MVM with all physics effects
- `irdrop.go` - Wire resistance and voltage drop
- `sneakpath.go` - Parasitic current analysis
- `drift.go` - Time-dependent conductance drift
- `temperature.go` - Temperature-dependent physics
- `enhanced.go` - Differential arrays, write-verify
- `reference.go` - Low-level reference implementations

### Network Integration (`pkg/network/`)
- `network.go` - Multi-layer neural network inference

### Training (`pkg/training/`)
- `training.go` - Quantization-aware, hardware-aware backpropagation

### Weight Management (`pkg/weights/`)
- `weights.go` - Format definitions
- `serialization.go` - JSON/binary/NumPy export

### Visualization (`pkg/gui/`)
- `app.go` - Main Fyne application
- `heatmap.go` - Conductance/IR drop/sneak visualization
- `controls.go` - Control panels and sliders
- `tabs/` - Specialized analysis tabs

### Evaluation (`pkg/evaluation/`)
- `accuracy.go` - MNIST and inference accuracy metrics

---

## Key Workflows

### 1. Basic MVM Operation

```
1. Create array with configuration
   ├─ Set size (rows × cols)
   ├─ Configure ADC/DAC bits
   ├─ Choose conductance model
   └─ Enable/disable non-idealities

2. Program weights
   ├─ Load from file or generate
   ├─ Quantize to 30 levels
   └─ Apply to cells

3. Run MVM
   ├─ Apply input voltage vector
   ├─ Compute I = G × V
   ├─ Apply non-idealities (optional)
   └─ Quantize output (ADC)

4. Analyze results
   ├─ Compare ideal vs actual output
   ├─ Calculate RMSE, max error
   ├─ Estimate accuracy loss
   └─ Report energy consumption
```

### 2. Non-Ideality Analysis Workflow

```
Enable IR Drop:
  ├─ Set wire resistance parameters
  ├─ Compute voltage distribution
  ├─ Visualize V_eff heatmap
  └─ Report max/avg IR drop

Enable Sneak Paths:
  ├─ Select architecture (0T1R vs 1T1R)
  ├─ Compute parasitic currents
  ├─ Visualize sneak current map
  └─ Report sneak ratio

Enable Variation:
  ├─ Set device sigma, gradients
  ├─ Generate variation map
  ├─ Apply to MVM
  └─ Report variance metrics

Enable Drift:
  ├─ Set drift coefficient, time
  ├─ Compute G(t) for each cell
  ├─ Visualize drift heatmap
  └─ Report accuracy degradation
```

### 3. Write-Verify Programming

```
Target conductance level (0-29):
  ├─ Apply initial write pulse
  ├─ Read back actual level
  ├─ Calculate error
  ├─ If error > tolerance:
  │   ├─ Adjust pulse strength
  │   ├─ Apply corrective pulse
  │   └─ Repeat (max iterations)
  └─ Converged or max iterations reached
```

### 4. Multi-Layer Network Inference

```
Network initialization:
  ├─ Define layer sizes
  ├─ Create crossbar per layer
  └─ Load trained weights

Forward pass:
  ├─ Layer 1: MVM + ReLU
  ├─ Layer 2: MVM + ReLU
  ├─ ...
  └─ Output layer: MVM + Softmax

Accuracy evaluation:
  ├─ Run on test dataset
  ├─ Compare with ground truth
  ├─ Report accuracy metrics
  └─ Analyze degradation sources
```

---

## Feature Matrix

### Conductance Models

| Model | Accuracy | Speed | Use Case |
|-------|----------|-------|----------|
| **Linear** | Moderate | Fast | Prototyping, education |
| **Exponential** | High | Medium | Realistic FeFET simulation |
| **Lookup Table** | Highest | Medium | Calibrated from measurements |

### Non-Idealities

| Effect | 0T1R Impact | 1T1R Impact | Configurable |
|--------|-------------|-------------|--------------|
| **IR Drop** | High | Medium | Yes (R_wire) |
| **Sneak Paths** | Very High (5-20% RMSE) | Low (<0.1%) | Yes (architecture) |
| **Variation** | Medium (1-2% RMSE) | Medium | Yes (σ, gradients) |
| **Drift** | Low-Medium (time-dependent) | Low-Medium | Yes (ν, t) |
| **Temperature** | Varies with T | Varies with T | Yes (77K-400K) |

### Architectures

| Architecture | Isolation | Complexity | Use Case |
|--------------|-----------|------------|----------|
| **0T1R** (Passive) | None | Simple | Research, small arrays |
| **1T1R** (1 Transistor) | 1000× | Moderate | Production systems |
| **2T1R** (2 Transistors) | Complete | Complex | Premium applications |
| **FeCAP** (Capacitive) | N/A | Medium | Charge-domain sensing, no sneak paths |

### FeCAP Architecture (Capacitive Crossbar)

FeCAP crossbars eliminate sneak paths and IR drop by operating entirely in the charge domain:

- Charge sensing: `Q[col] = ΣC[row][col] × V[row]` (capacitive summation)
- Displacement current: `I_disp = ΔQ/Δt` (no DC leakage path)
- Implemented in `module2-crossbar/pkg/gui/tabs/fecap_tab.go`
- GUI: Bar chart shows Q[col] per column; I_disp derived from ΔQ/Δt
- See: `docs/4-research/literature-review/crossbar-circuits-literature-review-2025.md` for literature context

### Multi-Hop Sneak Path Analysis

Beyond nearest-neighbor sneak paths, multi-hop sneak current flows through two or more intermediate cells:

- Implemented in `shared/crossbar/sneak_multihop.go`
- Severity increases with array size (scales roughly as N²)
- Only relevant for passive (0T1R) arrays; 1T1R transistors eliminate all sneak paths

### Export Formats

| Format | Use Case | Contents |
|--------|----------|----------|
| **JSON** | Reporting, archiving | Array stats, MVM results, energy |
| **CSV** | Spreadsheet analysis | Weight matrix, conductance values |
| **NumPy** | Python ML integration | Multi-dimensional arrays |
| **SPICE** | Circuit simulation | Netlist with R, G values |

---

## Extension Points

### 1. Adding New Conductance Models

```go
// In array.go
type ConductanceModel int

const (
    ConductanceLinear ConductanceModel = iota
    ConductanceExponential
    ConductanceLookup
    ConductanceCustom  // Add your model
)

func (a *Array) applyCustomConductanceModel(g_norm float64) float64 {
    // Your custom G(g_norm) function
    return customValue
}
```

### 2. New Non-Ideality Effects

```go
// Create new file: pkg/crossbar/myneweffect.go

type MyNewEffectConfig struct {
    Enabled bool
    Param1  float64
    Param2  float64
}

func (a *Array) applyMyNewEffect(current []float64, opts *MVMOptions) []float64 {
    // Modify current based on your effect
    return modifiedCurrent
}
```

### 3. Custom Visualization Tabs

```go
// In pkg/gui/tabs/

type MyAnalysisTab struct {
    array  *crossbar.Array
    canvas *fyne.Container
    // ... widgets
}

func (t *MyAnalysisTab) BuildContent() fyne.CanvasObject {
    // Create custom visualizations
}
```

### 4. Alternative Training Algorithms

```go
// In pkg/training/

type MyTrainer struct {
    config *MyTrainingConfig
}

func (t *MyTrainer) Train(network *Network, dataset Dataset) error {
    // Your custom training loop
}
```

---

## Known Limitations

### Current Implementation

1. **No transistor-level simulation:** Behavioral model only, not SPICE-level
2. **Simplified wire models:** Not process-specific, no parasitic capacitance
3. **Default parameters:** Educational baselines, not production-calibrated
4. **Drift coefficients:** Estimated from retention data, not directly measured
5. **Sneak path optimization:** Simplified model for arrays > 32×32 (trades accuracy for speed)
6. **Thread safety:** Not thread-safe for concurrent operations on same array

### Performance Considerations

- **Large arrays (>256×256):** Sneak path calculation becomes slow (use simplified model)
- **Many layers (>10):** Network inference accumulates quantization errors
- **High-resolution ADC (>8 bits):** Diminishing returns vs silicon area cost
- **Drift simulation:** Long time ranges (years) require many snapshots

### Platform Support

| Platform | GUI | Headless | GPU Acceleration |
|----------|-----|----------|------------------|
| Linux | ✅ | ✅ | ✅ (CUDA/OpenCL) |
| macOS | ✅ | ✅ | ⚠️ (Limited) |
| Windows | ✅ | ✅ | ✅ (CUDA) |

---

## Performance Tips

### For Speed

```bash
# Use linear conductance model (fastest)
# Disable non-idealities during prototyping
# Use smaller arrays for quick iteration
# Enable GPU acceleration for large MVMs

go test -run TestArrayMVM -bench=. ./module2-crossbar/pkg/crossbar
```

### For Accuracy

```bash
# Use exponential or lookup table conductance
# Enable all non-idealities
# Use 1T1R architecture (reduces sneak paths)
# Increase ADC/DAC resolution
# Use write-verify programming

go test -v ./module2-crossbar/pkg/crossbar
```

---

## Data Export

### JSON Export Example

```json
{
  "array_size": [128, 128],
  "architecture": "1T1R",
  "mvm_result": {
    "rmse": 0.0234,
    "max_error": 0.0891,
    "total_energy_pJ": 0.673,
    "gpu_equivalent_pJ": 163.84,
    "efficiency": 243.5
  },
  "ir_drop": {
    "max_drop_V": 0.043,
    "avg_drop_V": 0.021
  },
  "sneak_analysis": {
    "max_sneak_ratio": 0.0012,
    "isolation_factor": 0.001
  }
}
```

### CSV Export Example

```csv
row,col,level,conductance_norm,conductance_uS
0,0,15,0.517,56.7
0,1,23,0.793,82.3
...
```

### NumPy Export

```python
import numpy as np

# Load weights
weights = np.load('crossbar_weights.npy')
print(weights.shape)  # (128, 128)

# Analyze
mean_weight = weights.mean()
std_weight = weights.std()
```

---

## Testing

### Unit Tests

```bash
# All crossbar tests
go test ./module2-crossbar/pkg/crossbar

# Specific test categories
go test -run TestPhysics ./module2-crossbar/pkg/crossbar
go test -run TestNonIdealities ./module2-crossbar/pkg/crossbar
go test -run TestImprovement ./module2-crossbar/pkg/crossbar

# With coverage
go test -cover ./module2-crossbar/pkg/crossbar

# Verbose output
go test -v ./module2-crossbar/pkg/crossbar
```

### Integration Tests

```bash
# Network inference test
go test -run TestNetworkForward ./module2-crossbar/pkg/network

# Training test
go test -run TestTrainer ./module2-crossbar/pkg/training

# Weight export test
go test -run TestExport ./module2-crossbar/pkg/weights
```

### Benchmarks

```bash
# MVM performance
go test -bench=BenchmarkMVM ./module2-crossbar/pkg/crossbar

# Non-ideality overhead
go test -bench=BenchmarkNonIdealities ./module2-crossbar/pkg/crossbar
```

---

## Integration with Other Modules

### Module 1 (Hysteresis)

```go
// Get conductance from hysteresis model
level := hysteresis.GetDiscreteLevel()  // 0-29
g_norm := float64(level) / 29.0
crossbar.ProgramWeight(row, col, g_norm)
```

### Module 3 (MNIST)

```go
// Neural network layer uses crossbar
layer := network.NewLayer(784, 128)  // Input to hidden
layer.LoadWeights(trainedWeights)
output := layer.Forward(mnistImage)  // Uses crossbar.MVM
```

### Module 4 (Circuits)

```go
// Configure ADC/DAC to match circuit specs
cfg := crossbar.Config{
    DACBits: 8,   // 8-bit input DAC
    ADCBits: 6,   // 6-bit output ADC
}

// Energy estimates inform circuit design
result, _ := array.MVMWithNonIdealities(input, opts)
fmt.Printf("ADC energy: %.2f pJ\n", result.ADCEnergy)
```

### Module 6 (EDA)

```go
// Export for layout generation
array.ExportSPICE("netlist.sp")     // For SPICE simulation
array.ExportCSV("layout_params.csv") // For layout tool
```

---

## Troubleshooting

### High RMSE in MVM

**Check:**
1. Architecture (0T1R has high sneak paths)
2. Wire resistance (high R_wire → large IR drop)
3. Device variation (high σ → more noise)
4. Array size (larger arrays → more IR drop)

**Fix:**
- Switch to 1T1R architecture
- Reduce wire resistance
- Lower device variation
- Use smaller array or tile large arrays

### Write-Verify Not Converging

**Check:**
1. Target level is within [0, 29]
2. Tolerance is reasonable (default: 0.5 levels)
3. Max iterations is sufficient (default: 10)
4. Write noise is not too high

**Fix:**
- Increase max iterations
- Relax tolerance slightly
- Reduce write noise
- Check if target is achievable with current physics

### Slow Sneak Path Calculation

**Check:**
1. Array size (full calculation for ≤32×32)
2. Architecture (0T1R requires full sneak analysis)

**Fix:**
- Use simplified sneak model for large arrays
- Switch to 1T1R (much faster, minimal sneak)
- Disable sneak path analysis if not critical

---

## Future Enhancements (Roadmap)

- [ ] Multi-bit cells (exploit >30 levels)
- [ ] 3D stacked arrays
- [ ] Advanced drift models (measured calibration)
- [ ] Memristor device types (beyond FeFET)
- [ ] Hybrid analog-digital architectures
- [ ] Online learning with in-situ training
- [ ] Power modeling with dynamic switching
- [ ] Optical crossbar integration

---

## See Also

- **[Physics Reference](./physics.md)** - Equations and models
- **[Architecture](./architecture.md)** - Code structure and types
- **[Tools](./tools.md)** - External tool integration

---

**Last Updated:** 2026-02-16
