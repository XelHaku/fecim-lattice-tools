# IronLattice Research Findings - Part 6 (January 2026)

Additional research on CIM compilers, thermal management, application domains, and implementation analysis.

---

## 1. CIM Compiler and Software Toolchains

### Compilation Challenges for CIM

| Challenge | Description |
|-----------|-------------|
| Weight mapping | Assign NN weights to crossbar cells |
| Tiling | Partition large matrices across arrays |
| Dataflow | Optimize data movement |
| Precision | Handle mixed-precision quantization |
| Non-idealities | Account for hardware errors |

### Major CIM Compilation Frameworks (2024-2025)

| Framework | Year | Key Innovation | Source |
|-----------|------|----------------|--------|
| **CIM-MLC** | 2024 | Multi-level compilation stack | ASPLOS |
| **CMSwitch** | 2025 | Compute-memory mode switching | ASPLOS |
| **CINM (Cinnamon)** | 2024 | LLVM-based CIM migration | ResearchGate |
| **PIMCOMP** | 2023 | End-to-end DNN compiler | arXiv |
| **PIMLC** | 2024 | Bit-serial logic compiler | DATE |
| **CoMN** | 2024 | Algorithm-hardware co-design | TCAD |

### CMSwitch: Dual-Mode Compiler (2025)

**Title:** "Be CIM or Be Memory: A Dual-mode-aware DNN Compiler for CIM Accelerators"

**Key Innovation:** Addresses CIM's capability to switch dynamically between compute and memory modes for LLM workloads.

| Feature | Benefit |
|---------|---------|
| Mode switching | Optimal compute vs memory selection |
| Hardware abstraction | New attribute for CIM modes |
| LLM support | Handles diverse memory needs |

**Source:** [arXiv](https://arxiv.org/html/2502.17006)

### SRAM-CIM Compilation Framework (2024)

**Features:**
- Efficient weight mapping strategies
- Calibration of computation voltage linear error (CCVLE)
- Mitigation of analog-to-digital quantization error (MAQE)

**Source:** [IEEE TCAD](https://dl.acm.org/doi/10.1109/TCAD.2024.3366025)

### Compiler Stack Architecture

```
┌─────────────────────────────────────┐
│       High-Level Framework          │
│    (PyTorch, TensorFlow, ONNX)      │
├─────────────────────────────────────┤
│         Graph Optimization          │
│   (Fusion, Pruning, Quantization)   │
├─────────────────────────────────────┤
│         Weight Mapping              │
│   (Tiling, Placement, Scheduling)   │
├─────────────────────────────────────┤
│        Hardware Abstraction         │
│   (Crossbar, ADC/DAC, Controller)   │
├─────────────────────────────────────┤
│         Runtime/Driver              │
│   (Programming, Inference, Debug)   │
└─────────────────────────────────────┘
```

---

## 2. Thermal Management in Ferroelectric Arrays

### Heat Dissipation Challenge

| Memory Type | Mechanism | Joule Heating |
|-------------|-----------|---------------|
| RRAM/Memristor | Filamentary | **High** |
| PCM | Phase change | **Very High** |
| STT-MRAM | Spin transfer | Moderate |
| **FeFET/FTJ** | Polarization | **Low** |
| **Memcapacitor** | Capacitance | **Near Zero** |

### Why Ferroelectric is Advantaged

**Ferroelectric memcapacitors** circumvent Joule heating issues:
- Displacement current does not produce Joule heating
- Zero static power during read/inference
- Enhanced energy efficiency in large arrays

**Source:** [PMC](https://pmc.ncbi.nlm.nih.gov/articles/PMC10624373/)

### Thermal Comparison

| Metric | Resistive Crossbar | Capacitive Crossbar | Improvement |
|--------|-------------------|---------------------|-------------|
| Energy per MAC | ~100 fJ | **3.8 pJ per VMM** | 14-57× lower |
| Thermal noise | High | Compatible with 3-bit ADC | Better SNR |
| Array scaling | Limited by heat | Less constrained | Larger arrays |

### Design Guidelines

| Consideration | Recommendation |
|---------------|----------------|
| Array size | FCM allows larger arrays than resistive |
| Duty cycle | Lower for high-activity cells |
| Thermal interface | Standard CMOS packaging sufficient |
| Operating temp | Room temp to 85°C validated |

### 3D Integration Thermal Challenges

For 3D stacked arrays:
- Interlayer dielectric thermal conductivity matters
- Heat accumulates in inner layers
- Through-silicon vias (TSVs) help dissipation

---

## 3. Application Domains

### Automotive ADAS

**Market Share:** 27.4% of 2024 neuromorphic chip revenue

| Requirement | FeFET Advantage |
|-------------|-----------------|
| Millisecond response | Low-latency CIM |
| Thermal envelope | Non-volatile, no refresh |
| Safety-critical | Deterministic operation |
| Battery efficiency | Sub-mW inference |

**Source:** [Mordor Intelligence](https://www.mordorintelligence.com/industry-reports/neuromorphic-chip-market)

### Medical Implants & Healthcare

**Market Growth:** 105.4% CAGR (highest vertical)

| Application | FeFET Benefit |
|-------------|---------------|
| Brain-computer interfaces | Sub-milliwatt classification |
| Adaptive neuro-stimulators | Non-volatile state retention |
| Portable diagnostics | Battery-powered edge AI |
| Biosignal processing | Real-time inference |

**Key Research:** "Emerging memory devices for neuromorphic computing in the Internet of Medical Things (IoMT)"

**Source:** [ScienceDirect](https://www.sciencedirect.com/science/article/pii/S2666386425003340)

### Edge AI Applications

| Application | Requirements | FeFET Fit |
|-------------|--------------|-----------|
| Smart earbuds | <10mW, always-on | Excellent |
| Autonomous drones | Real-time vision | Good |
| Industry 4.0 | Predictive maintenance | Excellent |
| Wearables | Ultra-low power | Excellent |

### Application Readiness Matrix

| Domain | TRL | Timeline | Key Challenge |
|--------|-----|----------|---------------|
| IoT sensors | 6-7 | 2025-2026 | Cost |
| Wearables | 5-6 | 2026-2027 | Reliability qualification |
| Automotive | 4-5 | 2027-2028 | Automotive-grade endurance |
| Medical implants | 3-4 | 2028+ | Biocompatibility, FDA approval |

---

## 4. IronLattice Demo Code Analysis

### Demo 1: Hysteresis Visualizer

**Location:** `demo1-hysteresis/`

#### Preisach Model Implementation

**File:** `pkg/ferroelectric/preisach.go`

```go
type PreisachModel struct {
    material      *HZOMaterial
    EcMean        float64   // Mean coercive field
    EcSigma       float64   // Distribution width (25% of Ec)
    turningPoints []float64 // History (LIFO stack)
    polarization  float64
}
```

**Key Methods:**
| Method | Function |
|--------|----------|
| `Update(E)` | Apply field, return polarization |
| `GetHysteresisLoop()` | Generate full P-E curve |
| `DiscreteStates(N)` | Return N analog states |

**Physics Implementation:**
- Hyperbolic tangent switching function (Bo Jiang method)
- Minor loop support via turning point history
- Memory wipe-out for overlapping loops

#### Material Parameters

**File:** `pkg/ferroelectric/material.go` (inferred)

| Parameter | Value | Unit |
|-----------|-------|------|
| Ps | 25 | μC/cm² |
| Ec | 1.0 | MV/cm |
| EcSigma | 25% of Ec | MV/cm |

### Demo 2: Inference Engine

**Location:** `demo2-inference/`

#### Crossbar Array Implementation

**File:** `pkg/crossbar/array.go`

```go
type Array struct {
    config    *Config
    cells     [][]Cell
    adcLevels int
    dacLevels int
}

type Cell struct {
    Conductance    float64 // Normalized 0-1
    NoiseFactor    float64 // Per-cell variation
    SwitchingCount int64   // Write cycle count
}
```

**Key Operations:**
| Method | Function | Complexity |
|--------|----------|------------|
| `MVM(input)` | Matrix-vector multiply | O(rows × cols) in hardware, O(1) in time |
| `VMM(input)` | Vector-matrix multiply | O(rows × cols) in hardware, O(1) in time |
| `ProgramWeight()` | Set cell conductance | O(1) |

**Non-Ideality Modeling:**
- Device-to-device variation (NoiseFactor)
- ADC/DAC quantization
- Read/write statistics tracking

#### Layer Implementations

The `pkg/layers/` directory contains extensive neural network layer implementations:

| Category | Files | Description |
|----------|-------|-------------|
| Core | `activations.go`, `convolution.go`, `normalization.go` | Standard NN layers |
| CIM-specific | `attention_cim.go`, `cam_attention.go` | Attention on CIM |
| Neuromorphic | `snn.go`, `snn_gnn.go`, `neuromorphic_devices.go` | Spiking networks |
| Optimization | `pruning.go`, `quantization.go`, `sparse.go` | Compression |
| Training | `training_loop.go`, `optimizer.go`, `loss.go` | Learning |
| Advanced | `photonic.go`, `reservoir_insensor.go`, `domain_wall.go` | Emerging tech |

### Code Quality Assessment

| Aspect | Status | Notes |
|--------|--------|-------|
| Physics accuracy | Good | Preisach model well-implemented |
| Extensibility | Excellent | Modular package structure |
| Documentation | Adequate | Inline comments present |
| Testing | Unknown | Test files not reviewed |
| Performance | Good | Standard Go patterns |

---

## 5. Additional Papers for Download

### CIM Compilers

| Paper | URL | Year |
|-------|-----|------|
| CMSwitch Dual-Mode | https://arxiv.org/html/2502.17006 | 2025 |
| CINM Cinnamon | https://www.researchgate.net/publication/390679656 | 2024 |
| SRAM-CIM Compilation | https://dl.acm.org/doi/10.1109/TCAD.2024.3366025 | 2024 |
| CoMN Platform | https://dl.acm.org/doi/10.1109/TCAD.2024.3358220 | 2024 |
| GMap Neuromorphic | https://dl.acm.org/doi/10.1145/3589737.3605997 | 2023 |
| Compiler Survey | https://spj.science.org/doi/10.34133/icomputing.0040 | 2023 |

### Thermal Management

| Paper | URL | Year |
|-------|-----|------|
| Ferroelectric Memcapacitor | https://pmc.ncbi.nlm.nih.gov/articles/PMC10624373/ | 2023 |
| Capacitive Crossbar Design | https://ieeexplore.ieee.org/document/9439603/ | 2021 |
| 1Kbit Crossbar Array | https://ieeexplore.ieee.org/document/10044479/ | 2023 |
| FCM Review | https://nanoconvergencejournal.springeropen.com/articles/10.1186/s40580-024-00463-0 | 2024 |
| 3D RRAM Thermal | https://www.nature.com/articles/srep13504 | 2015 |

### Application Domains

| Paper | URL | Year |
|-------|-----|------|
| Neuromorphic Brain Implants | https://www.frontiersin.org/journals/neuroscience/articles/10.3389/fnins.2025.1570104/full | 2025 |
| IoMT Memory Devices | https://www.sciencedirect.com/science/article/pii/S2666386425003340 | 2025 |
| Flexible ITO FeFET | https://www.nature.com/articles/s41467-024-46878-5 | 2024 |
| Reservoir Computing Apps | https://link.springer.com/article/10.1557/s43577-025-00990-z | 2025 |
| Neuromorphic Implantables | https://arxiv.org/html/2506.09599v1 | 2025 |

---

## 6. Implementation Recommendations

### For Demo Enhancement

| Demo | Recommended Addition | Priority |
|------|---------------------|----------|
| Demo 1 | Vulkan GPU acceleration | High |
| Demo 1 | Interactive web UI | Medium |
| Demo 2 | Write-verify simulation | High |
| Demo 2 | Variability injection | Medium |
| Demo 3 | MNIST dataset loader | High |
| Demo 3 | Training visualization | Medium |

### For Compiler Integration

```
Future Architecture:
┌─────────────────────────────────────┐
│          IronLattice API            │
├─────────────────────────────────────┤
│    ONNX/PyTorch Model Import        │
├─────────────────────────────────────┤
│      Weight Quantization (INT4)     │
├─────────────────────────────────────┤
│       Crossbar Mapping Engine       │
├─────────────────────────────────────┤
│      Simulation / Hardware API      │
└─────────────────────────────────────┘
```

### Testing Strategy

| Test Type | Target | Coverage |
|-----------|--------|----------|
| Unit tests | Individual functions | >80% |
| Integration | End-to-end inference | Critical paths |
| Benchmark | Performance metrics | Key operations |
| Validation | Against published results | IronLattice claims |

---

## 7. Key Takeaways

### Research Maturity

| Topic | Papers Found | Maturity |
|-------|-------------|----------|
| CIM compilers | 15+ | Growing rapidly |
| Thermal management | 10+ | Well understood |
| Automotive | 20+ | Commercial stage |
| Medical | 15+ | Early research |

### IronLattice Differentiation

1. **Superlattice structure** - Better endurance than solid-solution HZO
2. **Tour Lab synthesis** - Unique research sample/FWF material preparation
3. **Comprehensive demos** - Physics-accurate Go implementation
4. **Research foundation** - 250+ papers collected and documented

### Next Steps

1. **Vulkan TDGL** - GPU-accelerate Demo 1
2. **Write-verify** - Add to crossbar simulation
3. **Compiler prototype** - ONNX to crossbar mapping
4. **Benchmark suite** - Validate against published CIM results

---

## References

1. CMSwitch Compiler: https://arxiv.org/html/2502.17006
2. Ferroelectric Memcapacitor: https://pmc.ncbi.nlm.nih.gov/articles/PMC10624373/
3. Neuromorphic Chip Market: https://www.mordorintelligence.com/industry-reports/neuromorphic-chip-market
4. IoMT Memory Devices: https://www.sciencedirect.com/science/article/pii/S2666386425003340
5. SRAM-CIM Literature: https://github.com/BUAA-CI-LAB/Literatures-on-SRAM-based-CIM

---

*Last updated: January 2026 - Part 6*
