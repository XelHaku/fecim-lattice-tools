# Open Source Crossbar Simulation Tools

**Analysis of badcrossbar and CrossSim for FeCIM improvement**

*Last Updated: January 2026*

---

## Executive Summary

We analyzed two open-source crossbar simulation projects to improve our FeCIM implementation:

| Project | Developer | Purpose | Best For |
|---------|-----------|---------|----------|
| **badcrossbar** | UCL (Academic) | Pure nodal analysis | Validation, education |
| **CrossSim** | Sandia National Labs | Production simulator | GPU acceleration, neural networks |

**Key Recommendation:** Use CrossSim's iterative solver for performance, badcrossbar for validation.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

---

## 1. badcrossbar

**GitHub:** https://github.com/joksas/badcrossbar
**License:** MIT
**Language:** Python (~617 SLOC)
**Publication:** [SoftwareX 2020](https://doi.org/10.1016/j.softx.2020.100617)

### Core Algorithm: Nodal Analysis via KCL

Solves the sparse linear system: `g · v = i`

```python
# 1. Build conductance matrix g (sparse lil_matrix)
g_matrix = fill.g(resistances, r_i)

# 2. Build current injection vector i
i_vector = fill.i(applied_voltages, resistances, r_i)

# 3. Solve sparse system (SuperLU via scipy)
v_solution = scipy.sparse.linalg.spsolve(g.tocsc(), i)

# 4. Extract branch currents from node voltages
device_currents = (V_wordline - V_bitline) / R_device
output_currents = sum(device_currents, axis=row)
```

### KCL Matrix Assembly

For interior nodes (columns 1 to n-2):
```
g[idx, idx] = 2·g_interconnect + g_device[i,j]  # Self-conductance
g[idx, idx-1] = -g_interconnect                 # Left neighbor
g[idx, idx+1] = -g_interconnect                 # Right neighbor
g[idx, idx+crossbar_size] = -g_device[i,j]      # Connected bit line node
```

### API Design

```python
import badcrossbar

solution = badcrossbar.compute(
    applied_voltages,   # (m, p) array
    resistances,        # (m, n) array
    r_i=0.5,            # Scalar or tuple (r_i_word, r_i_bit)
)

# Access results
output_current = solution.currents.output[0, 2]
device_current = solution.currents.device[1, 3, 0]
voltage = solution.voltages.bit_line[-1, 1]
```

### Visualization

Publication-quality vector PDFs via Cairo:
```python
badcrossbar.plot.branches(currents=solution.currents, filename="currents")
badcrossbar.plot.nodes(voltages=solution.voltages, filename="voltages")
```

### Strengths
- Mathematically exact (within floating-point precision)
- Clean separation of concerns (KCL assembly vs solving)
- Handles arbitrary interconnect resistance (word line ≠ bit line)
- Self-documenting named tuples

### Limitations
- No device non-idealities (only resistances)
- No ADC/DAC quantization
- No read noise or programming error
- Single-threaded (no GPU)
- Scales poorly: O(N³) for dense N×N array

### Scalability

| Array Size | Performance |
|------------|-------------|
| ≤64×64 | <100ms |
| 128×128 | ~1-10s |
| ≥256×256 | >1 minute (memory-limited) |

---

## 2. CrossSim v3.1

**GitHub:** https://github.com/sandialabs/cross-sim
**License:** BSD-3 (U.S. Government work)
**Language:** Python (~4,300+ SLOC)
**GPU Support:** CuPy (CUDA 12.3)

### Core Algorithm: Iterative Circuit Solver (SOR)

Successive over-relaxation with adaptive γ:

```python
dV = applied_voltages  # Initial guess
Ires = matrix * dV     # Device currents (Ohm's law)

while Verr > threshold and iters < max_iters:
    # 1. Compute cumulative currents
    Isum_col = cumsum(Ires, axis=columns)
    Isum_row = cumsum(Ires[::-1], axis=rows)

    # 2. Compute parasitic voltage drops
    Vdrops_col = cumsum(Rp_col * Isum_col[:,::-1], axis=1)[:,::-1]
    Vdrops_row = cumsum(Rp_row * Isum_row, axis=0)
    Vpar = Vdrops_col + Vdrops_row

    # 3. Calculate voltage error
    VerrMat = V_applied - Vpar - dV
    Verr = max(abs(VerrMat))

    # 4. Update device voltages (successive under-relaxation)
    dV += gamma * VerrMat
    Ires = matrix * dV

output = sum(Ires, axis=rows)
```

### Array Topologies

| Topology | Current Source | Use Case |
|----------|---------------|----------|
| `NonInterleaved_InputSource` | Input line | Standard VMM/MVM |
| `Interleaved_InputSource` | Input line | Local current cancellation |
| `NonInterleaved_SepSource` | Dedicated source | 1T1R with bit slicing |
| `Interleaved_SepSource` | Dedicated source | 1T1R + cancellation |

### API Design

**Level 1: Numpy Replacement**
```python
from simulator.cores import AnalogCore
from simulator.parameters import CrossSimParameters

params = CrossSimParameters.from_file("configs/default.json")
core = AnalogCore(matrix, params)

output = core @ input_vector        # MVM
output = input_vector @ core        # VMM
```

**Level 2: Neural Network Layers**
```python
from simulator.algorithms.dnn.keras import replace_keras_layers

model = keras.models.load_model("resnet50.h5")
analog_model = replace_keras_layers(model, params)
accuracy = analog_model.evaluate(test_data)
```

**Level 3: Direct Core Access**
```python
balanced = BalancedCore(clipper_factory, params)
balanced.set_matrix(weights)
output = balanced.run_mvm(inputs, adc=True, dac=True)
conductance_matrix = balanced.core_pos.matrix
```

### Parameter Hierarchy

```
CrossSimParameters
├── core (CoreParameters)
│   ├── style: BALANCED | BITSLICED | OFFSET
│   ├── rows_max, cols_max, weight_bits
│   └── mapping (weights, inputs)
├── simulation (SimulationParameters)
│   ├── useGPU, gpu_id
│   ├── Niters_max_parasitics, Verr_th_mvm
│   └── relaxation_gamma, fast_matmul
└── xbar (XbarParameters)
    ├── device (DeviceParameters)
    │   ├── cell_bits, Rmin, Rmax
    │   ├── read_noise, programming_error, drift_error
    ├── array (parasitics: Rp_row, Rp_col)
    ├── adc (6 models: Ideal, Quantizer, SAR, Ramp, Pipeline, Cyclic)
    └── dac (Ideal, Quantizer)
```

### Device Non-Ideality Models

| Effect | Model Options |
|--------|---------------|
| Quantization | `cell_bits` → 2^N levels |
| Programming Error | Normal/Uniform × Independent/Proportional/Inverse |
| Read Noise | Per-operation white noise |
| Drift | Time-dependent error |
| ADC Non-idealities | Capacitor mismatch, comparator offset, finite gain |

**Built-in Device Models:**
- **SONOS** (Sandia, IEDM 2022)
- **PCM_Joshi** (IBM, Nat. Commun. 2020)
- **RRAM_Milo** (Politecnico, IRPS 2021)
- **RRAM_Wan** (Tsinghua, Nature 2022)

### Performance

| Model | CPU | GPU | Speedup |
|-------|-----|-----|---------|
| MNIST CNN6 | ~100ms | ~10ms | 10× |
| ResNet50 | ~120ms | ~8ms | 15× |
| VGG19 | ~200ms | ~15ms | 13× |

**GPU Optimizations:**
- CuPy backend (automatic numpy/cupy switch)
- `fast_matmul`: Batches MVMs into single matmul
- Tensor core exploitation on modern GPUs

### Strengths
- Production-grade (validated against SPICE)
- GPU acceleration (CuPy)
- Extensive device models
- Neural network APIs (Keras, PyTorch)
- Multi-level abstraction

### Limitations
- Linear device assumption (R = constant)
- Complex parameter hierarchy
- Documentation scattered across tutorials
- No real-time visualization

---

## 3. Feature Comparison

| Feature | badcrossbar | CrossSim | FeCIM Target |
|---------|-------------|----------|--------------|
| **IR Drop** | ✅ Exact | ✅ Converged | ✅ |
| **Sneak Paths** | ✅ Implicit | ✅ Implicit | ✅ |
| **Device Quantization** | ❌ | ✅ cell_bits | ✅ 30 levels |
| **Programming Error** | ❌ | ✅ 4 models | ✅ |
| **Read Noise** | ❌ | ✅ | ✅ |
| **Drift** | ❌ | ✅ | ⚠️ Later |
| **ADC/DAC** | ❌ | ✅ 6 models | ✅ Quantizer |
| **GPU Acceleration** | ❌ | ✅ CuPy | ✅ Vulkan |
| **Visualization** | ✅ Vector PDF | ⚠️ Manual | ✅ Fyne |
| **Code Clarity** | ✅ Excellent | ⚠️ Complex | ✅ |
| **Scalability** | ⚠️ Poor | ✅ Good | ✅ |

---

## 4. Physics Equations

### Parasitic Resistance (from CrossSim)

Normalized by device Rmin:
```
Rp_row_norm = Rp_row / Rmin
Rp_col_norm = Rp_col / Rmin
```

Parasitic voltage drop (cumulative sum):
```
Isum_col[i,j] = Σ(Ires[i, 0:j+1])
Vdrop_col[i,j] = Σ(Rp_col × Isum_col[i, j:end])
```

### KCL at Node (from badcrossbar)

```
(V[i,j] - V[i-1,j])/Rp_row + (V[i,j] - V[i+1,j])/Rp_row
+ (V[i,j] - V[i,j-1])/Rp_col + (V[i,j] - V[i,j+1])/Rp_col
+ (V[i,j] - V_bitline[i,j])/R_device = 0
```

### Device Error Models

**Gaussian Programming Error:**
```
G_programmed = G_target × (1 + N(0, σ))
```

**Read Noise (per-operation):**
```
G_read = G_programmed × (1 + N(0, σ_read))
```

**ADC Quantization:**
```
levels = 2^bits
code = round((I - I_min) / (I_max - I_min) × (levels - 1))
```

---

## 5. Implementation Roadmap for FeCIM

### Phase 1: Core Solver (MVP)
**Goal:** Match badcrossbar's IR drop accuracy

```go
type CrossbarArray struct {
    Conductances [][]float64  // Normalized (Gmin=0, Gmax=1)
    RpRow        float64       // Normalized row parasitic
    RpCol        float64       // Normalized col parasitic
}

type MVMResult struct {
    OutputCurrents []float64
    DeviceCurrents [][]float64
    Iterations     int
    Converged      bool
}
```

**Key Algorithm (port from CrossSim):**
```go
func (s *Solver) MVM(inputs []float64) (*MVMResult, error) {
    dV := inputs // Initial guess
    gamma := 1.0

    for iters := 0; iters < maxIters; iters++ {
        Ires := s.matmul(s.conductances, dV)

        // Cumulative currents
        IsumCol := s.cumsum(Ires, axisColumns)
        IsumRow := s.cumsumReverse(Ires, axisRows)

        // Parasitic drops
        VdropsCol := s.cumsum(multiply(s.RpCol, IsumCol), axisColumns)
        VdropsRow := s.cumsum(multiply(s.RpRow, IsumRow), axisRows)
        Vpar := add(VdropsCol, VdropsRow)

        // Voltage error
        VerrMat := subtract(subtract(s.Vapplied, Vpar), dV)
        Verr := max(abs(VerrMat))

        if Verr < threshold {
            return &MVMResult{Converged: true, Iterations: iters}, nil
        }

        // SOR update
        dV = add(dV, multiply(gamma, VerrMat))
    }
    return nil, ErrConvergenceFailed
}
```

### Phase 2: Device Non-Idealities
**Goal:** Add quantization, programming error, read noise

```go
type DeviceConfig struct {
    CellBits          int     // 0 = continuous, 5 = 32 levels
    ProgrammingError  float64 // σ as fraction of range
    ReadNoise         float64 // σ per operation
}

func QuantizeConductances(G [][]float64, bits int) [][]float64 {
    levels := math.Pow(2, float64(bits))
    // Round to nearest level in [Gmin, Gmax]
}

func AddProgrammingError(G [][]float64, sigma float64) [][]float64 {
    // G_prog = G_target × (1 + N(0, σ))
}
```

### Phase 3: ADC/DAC Models
**Goal:** Quantization + range optimization

```go
type ADC struct {
    Bits   int
    Signed bool
    Min    float64
    Max    float64
}

func (adc *ADC) Convert(currents []float64) []float64 {
    levels := math.Pow(2, float64(adc.Bits))
    // Uniform quantization
}
```

### Phase 4: GUI Integration
**Goal:** Real-time visualization in Fyne

- Integrate into existing Module 2
- Add sliders: parasitic R, error σ, ADC bits
- Visualize: device currents, output distributions
- Metrics: SNR degradation, accuracy loss

---

## 6. Validation Strategy

### Cross-Check with badcrossbar

```go
func TestParasiticResistance(t *testing.T) {
    // Run badcrossbar via Python subprocess
    expected := runBadcrossbar(voltages, resistances, r_i)

    // Run FeCIM Go solver (no device errors)
    actual := fecim.SolveParasitics(voltages, resistances, r_i)

    // Compare (tolerance: 0.1%)
    assert.InDelta(t, expected, actual, 0.001)
}
```

### SPICE Validation (Small Arrays)

```spice
* 4x4 Crossbar Cell Model
.title FeCIM Crossbar Validation

* Word line driver
VWL0 WL0 0 DC 2V

* Parasitic resistance per cell
RWL0_para WL0 WL0_internal 1

* Cell resistance (30-level quantized)
Rcell00 WL0_internal BL0_internal 10k

.op
.print dc i(Rcell00)
.end
```

### Test Cases

| Test | Array | Rp | Expected |
|------|-------|-----|----------|
| Zero parasitic | 8×8 | 0 | Ideal MVM |
| High parasitic | 8×8 | 10Ω | 10-20% drop |
| Edge case | 1×1 | 0 | Direct |
| Large array | 64×64 | 2.5Ω | Converges <10 iters |

---

## 7. Licensing & Attribution

### badcrossbar Citation
```bibtex
@article{JoksasMehonic2020,
  author = {Joksas, Dovydas and Mehonic, Adnan},
  title = {badcrossbar: A Python tool for computing currents and voltages
           in passive crossbar arrays},
  journal = {SoftwareX},
  volume = {12},
  pages = {100617},
  year = {2020},
  doi = {10.1016/j.softx.2020.100617}
}
```

### CrossSim Citation
```bibtex
@software{CrossSim,
  author = {Feinberg, Ben and Xiao, T. Patrick and Brinker, Curtis J.
            and Bennett, Christopher H. and Marinella, Matthew J.
            and Agarwal, Sapan},
  title = {CrossSim: accuracy simulation of analog in-memory computing},
  url = {https://github.com/sandialabs/cross-sim}
}
```

---

## 8. Related Documentation

- **[circuit-analysis-libraries.md](circuit-analysis-libraries.md)** - Full library analysis
- **[circuit-simulation-tools.md](circuit-simulation-tools.md)** - SPICE tools
- **[../crossbar/reference/VOLTAGE_RULES.md](../crossbar/reference/VOLTAGE_RULES.md)** - Voltage specifications
- **[../crossbar/educational/../educational/crossbar.physics.md](../crossbar/educational/../educational/crossbar.physics.md)** - Physics models

---

## Quick Reference: Algorithm Selection

| Use Case | Tool | Reason |
|----------|------|--------|
| IR drop validation | badcrossbar | Mathematically exact |
| Large arrays (>128×128) | CrossSim | Iterative solver scales |
| GPU acceleration | CrossSim | CuPy backend |
| Neural network inference | CrossSim | Keras/PyTorch APIs |
| Educational demos | badcrossbar | Clean code, visualizations |
| Production simulation | CrossSim | Industry-validated |

**Bottom Line:** Port CrossSim's iterative solver to Go, validate against badcrossbar, integrate into Module 2 GUI.
