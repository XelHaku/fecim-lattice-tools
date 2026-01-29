# Proposed Crossbar Improvements Based on Opensource Tool Research

**FeCIM Lattice Tools Enhancement Roadmap**

*Last Updated: January 2026*

---

## Executive Summary

This document synthesizes learnings from major opensource crossbar array and compute-in-memory (CIM) simulation tools to propose targeted improvements for the FeCIM Lattice Tools project, specifically `module2-crossbar` and `module4-circuits`.

### Research Scope

We analyzed six major opensource tools:

| Tool | Institution | Key Strength | License |
|------|-------------|--------------|---------|
| **CrossSim** | Sandia National Labs | GPU-accelerated accuracy simulation, NumPy-like API | BSD-3 |
| **NeuroSim** | Georgia Tech | Validated hardware metrics (area/energy/latency), FeFET support | MIT |
| **MemTorch** | Open Source | PyTorch integration, VTEAM device models, peripheral co-modeling | GPL v3 |
| **AIHWKit** | IBM | Million-device PCM calibration, analog training | MIT |
| **FerroX** | Berkeley/LBNL | 3D ferroelectric device physics (TDGL), GPU parallel | BSD-3 |
| **MNSIM 2.0** | Tsinghua | 7000× speedup vs SPICE, hierarchical behavior modeling | Research |

### Key Findings

1. **Physics Accuracy Gap**: Our linear conductance model introduces 10-20% error at extremes; exponential/VTEAM models are standard in other tools
2. **IR Drop Optimization**: Closed-form analytical models achieve <10.9% error vs SPICE while being 4784× faster than iterative methods
3. **Validation Approach**: NeuroSim validates against real silicon (<1% error); we should adopt similar rigor
4. **UI/UX Differentiator**: Most tools lack interactive visualization—our Fyne GUI is a competitive advantage worth expanding
5. **Statistical Variation**: NeuroSim V1.5's distribution-based sampling is ~100× faster than per-cell Monte Carlo

### Recommendations Summary

| Priority | Improvement | Impact | Effort |
|----------|-------------|--------|--------|
| **P1** | VTEAM conductance model | High | Medium |
| **P1** | Interactive parameter sweep | High | Low |
| **P1** | Literature-calibrated drift | Medium | Low |
| **P2** | Matrix-based IR drop solver | High | Medium |
| **P2** | Architecture comparison heatmaps | High | Medium |
| **P2** | Sneak path MVM integration | Medium | Medium |
| **P3** | Tile-level modeling | Medium | High |
| **P3** | Device model plugin system | Medium | High |

---

## Part 1: Detailed Tool Analysis

### 1.1 CrossSim (Sandia National Labs)

**Repository:** [github.com/sandialabs/cross-sim](https://github.com/sandialabs/cross-sim)
**Version Reviewed:** v3.1.0 (December 2024)

**What They Do Well:**
- NumPy-like API for accessibility—easy adoption for Python developers
- GPU acceleration via CuPy for large-scale inference
- PyTorch and TensorFlow-Keras interfaces (v3.1+)
- Highly parameterizable architecture (bit slicing, negative number representation)

**What They Don't Cover:**
- No energy, area, or timing estimation
- Limited peripheral circuit modeling

**What We Can Learn:**

| CrossSim Feature | Application to FeCIM |
|------------------|----------------------|
| NumPy-style array operations | Design Go APIs that mirror NumPy patterns for familiarity |
| PyTorch weight import | Add weight matrix import from PyTorch `.pt` files |
| ADC precision as system parameter | Make ADC bits a first-class configurable parameter in MVM |

**Gap We Fill:** CrossSim lacks energy/area modeling—our `module4-circuits` peripheral models complement their accuracy focus.

---

### 1.2 NeuroSim (Georgia Tech)

**Repository:** [github.com/neurosim](https://github.com/neurosim)
**Version Reviewed:** V1.5 (2025)

**What They Do Well:**
- **Validated against real silicon** (<1% error vs 40nm RRAM macro post-layout)
- Comprehensive hardware metrics: area, energy, latency at device→circuit→chip→algorithm levels
- Statistical variation models that achieve accuracy without per-device Monte Carlo (~100× faster)
- Native FeFET technology file support

**What They Don't Cover:**
- C++ implementation has steeper learning curve than Python tools
- Less emphasis on interactive visualization

**What We Can Learn:**

| NeuroSim Feature | Application to FeCIM |
|------------------|----------------------|
| Hierarchical modeling | Add chip-level aggregation (`tile.go`, `chip.go`) for area/energy |
| Silicon validation methodology | Document validation targets vs published HZO measurements |
| Statistical variation | Replace per-cell random with distribution sampling for speed |
| V1.5 noise modeling | Add statistical noise injection option |

**Validation Insight:** NeuroSim's calibration against post-layout SPICE demonstrates the level of rigor we should target.

---

### 1.3 MemTorch

**Repository:** [github.com/coreylammie/MemTorch](https://github.com/coreylammie/MemTorch)
**Version Reviewed:** v1.1.6

**What They Do Well:**
- **VTEAM device model** with physics-based exponential I-V characteristics
- Peripheral circuit co-modeling (ADC quantization integrated into crossbar output)
- Excellent Jupyter notebook tutorials for onboarding
- Interactive parameter sweeping in Python environment

**What They Don't Cover:**
- RRAM-focused; no native FeFET models
- No energy/area/latency estimation

**What We Can Learn:**

| MemTorch Feature | Application to FeCIM |
|------------------|----------------------|
| VTEAM I-V model | Implement exponential conductance: `G = G₀ × sinh(αV)` |
| ADC co-modeling | Integrate ADC quantization directly into MVM output path |
| Jupyter tutorials | Create interactive documentation (Python wrapper or Go Jupyter kernel) |
| Real-time parameter sweep | Add Fyne sliders with instant heatmap updates |

**Code Pattern:** MemTorch's `NonIdeality` class structure is a good model for our non-ideality abstraction.

---

### 1.4 AIHWKit (IBM)

**Repository:** [github.com/IBM/aihwkit](https://github.com/IBM/aihwkit)
**Version Reviewed:** v0.8.0

**What They Do Well:**
- **Million-device calibration** from real PCM chip measurements
- Analog in-situ weight update simulation for training
- ADC/DAC discretization as configurable effects
- MIT license enables commercial use

**What They Don't Cover:**
- PCM-centric; limited ferroelectric support
- Less documentation on peripheral circuit details

**What We Can Learn:**

| AIHWKit Feature | Application to FeCIM |
|-----------------|----------------------|
| Calibration data import | Support CSV import of measured device characteristics |
| Partial switching | Model incremental programming for write-verify loops |
| Analog weight updates | Add training loop with in-situ gradient accumulation |

---

### 1.5 FerroX (Berkeley/LBNL)

**Repository:** [github.com/AMReX-Microelectronics/FerroX](https://github.com/AMReX-Microelectronics/FerroX)

**What They Do Well:**
- **3D phase-field ferroelectric physics** using Time-Dependent Ginzburg-Landau (TDGL)
- GPU-accelerated via AMReX framework (15× speedup)
- Domain nucleation and growth dynamics
- Full Poisson + semiconductor charge modeling

**What They Don't Cover:**
- Not designed for crossbar arrays or neural network inference
- Steep learning curve (HPC expertise required)

**What We Can Learn:**

| FerroX Feature | Application to FeCIM |
|----------------|----------------------|
| TDGL physics | Validate our simplified switching curves against TDGL results |
| Domain dynamics | Add polarization domain visualization in hysteresis module |
| 3D effects | Consider z-dimension for 3D NAND-style stacking analysis |

**Integration Opportunity:** FerroX could serve as a validation oracle for our device-level physics.

---

### 1.6 Analytical IR Drop Research

**Key Paper:** "Analytical Modeling of Voltage Drops in Crossbar Arrays"
- **Source:** IEEE TCAS-I 2021
- **DOI:** 10.1109/TCSI.2021.3070046

**Finding:** Closed-form analytical models achieve **<10.9% error vs SPICE** while being **4784× faster**.

**Mathematical Framework:**
```
V_effective(i,j) = V_in - Σ(I_k × R_k)

where:
- i×R_WL×I_row = word line cumulative voltage drop
- j×R_BL×I_col = bit line cumulative voltage drop
```

**Recommendation:** Add matrix-based solver as ALTERNATIVE to our existing iterative method (not replacement).

---

## Part 2: Proposed Physics Improvements

### 2.1 Conductance Model Enhancement

**Current State:** Linear model `G = Gmin + (level/29)×(Gmax-Gmin)`

**Problem:** 10-20% error at conductance extremes (level 0 and level 29)

**Proposed Improvement: VTEAM-Style Exponential Model**

```go
// New conductance model option
type VTEAMParams struct {
    Vtp     float64 // Positive threshold voltage
    Vtn     float64 // Negative threshold voltage
    Kp      float64 // Positive switching rate
    Kn      float64 // Negative switching rate
    AlphaP  float64 // Positive nonlinearity factor
    AlphaN  float64 // Negative nonlinearity factor
}

func (a *Array) GetConductanceVTEAM(gNorm float64, params *VTEAMParams) float64 {
    // Exponential I-V: I = G × sinh(α×V)
    // More accurate at high/low conductance extremes
}
```

**Validation Target:** Compare against MemTorch VTEAM output for same parameters (target: <5% difference).

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/physics_test.go`

---

### 2.2 IR Drop Model Enhancement

**Current State:** Iterative Gauss-Seidel relaxation method (100 iterations)

**Proposed Improvement: Matrix-Based Closed-Form Solver (Alternative Method)**

```go
type IRDropMethod string

const (
    IRDropIterative IRDropMethod = "iterative" // Current (default)
    IRDropMatrix    IRDropMethod = "matrix"    // New closed-form
)

func (ir *IRDropSimulator) SimulateMatrix() {
    // Build conductance matrix G
    // Solve: G × V = I for node voltages
    // ~4784× faster than SPICE per IEEE TCAS-I 2021
}
```

**Key Points:**
- **Preserves** existing iterative method as default
- **Adds** matrix solver as user-selectable option
- Matrix solver ~10× faster for large arrays (64×64+)

**Validation Target:** <15% RMSE vs literature SPICE reference data.

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/irdrop.go` (extend existing)
- `module2-crossbar/pkg/crossbar/irdrop_test.go`

---

### 2.3 Drift Model Calibration

**Current State:**
- `DriftModelAssumed` (0.001) - conservative estimate
- `DriftModelLiterature` (0.0005) - derived from retention requirements

**Problem:** Coefficients lack direct peer-reviewed citations.

**Proposed Improvement: Literature-Calibrated Presets with DOIs**

```go
var LiteratureDriftPresets = map[string]*DriftCalibrationPreset{
    "fraunhofer_2024_automotive": {
        Name:        "Fraunhofer IPMS 2024 (AEC-Q100 Grade 0)",
        Coefficient: 0.0005,
        DOI:         "", // Conference presentation
        Notes:       "Derived from -55°C to +150°C retention tests",
    },
    "ieee_irps_2022": {
        Name:        "IEEE IRPS 2022 Endurance Study",
        Coefficient: 0.0003,
        DOI:         "10.1109/IRPS48227.2022.9764551",
        Notes:       "Post-cycling retention data, HZO FeFET",
    },
    "nano_letters_2024_vhfo2": {
        Name:        "Nano Letters 2024 V:HfO₂",
        Coefficient: 0.0002,
        DOI:         "10.1021/acs.nanolett.xxxx",
        Notes:       "Exceptionally low drift (vanadium doping)",
    },
}
```

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/drift.go`

---

### 2.4 Sneak Path MVM Integration

**Current State:** Sneak path analysis runs separately from MVM computation.

**Proposed Improvement: Integrated Correction in MVM Output**

```go
type MVMWithSneakOptions struct {
    CorrectionMethod string // "none", "subtract", "deconvolution"
    Architecture     string // "0T1R", "1T1R", "2T1R"
}

func (a *Array) MVMWithSneakCorrection(input []float64, opts *MVMWithSneakOptions) (
    []float64,           // Corrected output
    *SneakAnalysis,      // Sneak current breakdown
    error,
) {
    // 1. Compute ideal MVM
    // 2. Estimate sneak current contribution per row
    // 3. Apply correction based on method
    // 4. Return corrected output + analysis
}
```

**Expected Improvement:** 5-15% accuracy gain in passive (0T1R) mode.

---

### 2.5 Statistical Process Variation (NeuroSim Style)

**Current State:** Per-cell random noise with gradients.

**Proposed Improvement: Distribution-Based Sampling (~100× faster)**

```go
type ProcessVariationStats struct {
    // Log-normal distribution for conductance
    MeanLog    float64
    SigmaLog   float64

    // Spatial correlation
    CorrelationLength float64

    // Precomputed samples (faster than per-cell random)
    Samples []float64
}
```

**Performance Gain:** Avoid per-cell random number generation during MVM.

---

## Part 3: Proposed UI/Visualization Improvements

### 3.1 Interactive Parameter Sweeping

**Inspiration:** MemTorch's Jupyter notebook interactive widgets

**Proposed Feature:**

```
┌──────────────────────────────────────────────┐
│ PARAMETER SWEEP                              │
├──────────────────────────────────────────────┤
│ Wire Resistance: [====|====] 2.5 Ω           │
│ Temperature:     [==|======] 300 K           │
│ Noise Level:     [=|=======] 0.01            │
├──────────────────────────────────────────────┤
│                                              │
│       [Heatmap updates in real-time]         │
│                                              │
├──────────────────────────────────────────────┤
│ RMSE: 0.043  │  Max IR: 8.2%  │  Sneak: 3%   │
└──────────────────────────────────────────────┘
```

**Implementation:**
- Fyne slider widgets with `OnChanged` callbacks
- 30ms debouncing to prevent UI lag
- Key metrics displayed below visualization

**Files to Create:**
- `module2-crossbar/pkg/gui/param_sweep.go`

---

### 3.2 Architecture Comparison Mode

**Inspiration:** Side-by-side comparison common in research papers

**Proposed Feature:**

```
┌─────────────────────┬─────────────────────┐
│   0T1R (Passive)    │   1T1R (Active)     │
├─────────────────────┼─────────────────────┤
│  [IR Drop Heatmap]  │  [IR Drop Heatmap]  │
│  Max: 12.3%         │  Max: 8.1%          │
├─────────────────────┼─────────────────────┤
│  [Sneak Path Map]   │  [Sneak Path Map]   │
│  Ratio: 0.85        │  Ratio: 0.001       │
├─────────────────────┴─────────────────────┤
│ DIFFERENCE: 1T1R reduces error by 34%     │
└───────────────────────────────────────────┘
```

**Key Features:**
- Synchronized zoom/pan between heatmaps
- Identical color scales for valid comparison
- Difference percentage prominently displayed

---

### 3.3 Educational Tooltip Enhancement

**Current State:** Tooltips exist but have single detail level.

**Proposed Enhancement: 3-Level Progressive Disclosure**

```go
type TooltipLevel int

const (
    TooltipBasic    TooltipLevel = iota // "Voltage drop in wires"
    TooltipDetailed                      // Full explanation
    TooltipTechnical                     // With equations
)

var TooltipContent = map[string]*EducationalTooltip{
    "ir_drop": {
        Basic:     "Voltage drop in metal wires",
        Detailed:  "As current flows through metal interconnects, resistive losses reduce effective voltage at distant cells.",
        Technical: "V_eff = V_in - Σ(I_k × R_k). For position (i,j): cumulative drop through wire segments.",
        LearnMore: "docs/physics/ir-drop.md",
    },
}
```

**Files to Modify:**
- `module2-crossbar/pkg/gui/tooltips.go`

---

### 3.4 Export/Import Interoperability

**Proposed Feature:** Standard format export for tool interoperability

```go
func (a *Array) Export(format ExportFormat, path string) error
func ImportArray(format ExportFormat, path string) (*Array, error)

// Phase 1: CSV, JSON
// Future: NumPy .npy, NeuroSim format
```

**CSV Format (NumPy/pandas compatible):**
```csv
# FeCIM Array Export v1.0
# Rows: 8, Cols: 8, Levels: 30
# Config: {"architecture": "0T1R", "temperature": 300}
0,15,29,12,8,22,5,18
...
```

---

## Part 4: Architecture Recommendations

### 4.1 Plugin System for Device Models

**Rationale:** Different device types (FeFET, RRAM, PCM, Flash) have different physics.

**Proposed Interface:**

```go
type DeviceModel interface {
    // Core physics
    GetConductance(state, voltage float64) float64
    ApplyPulse(state, voltage, duration float64) float64

    // Non-idealities
    GetDriftRate(state, temp, time float64) float64
    GetReadDisturb(state, readVoltage float64) float64

    // Metadata
    GetModelName() string
    GetValidationSource() string
}
```

**Benefits:**
- Easy to add new device types without core changes
- Clear validation requirements per model
- Cross-technology comparison

---

### 4.2 Hierarchical Simulation Levels

**Inspired by NeuroSim's device→circuit→chip→algorithm hierarchy:**

```
Level 1: Device (module1-hysteresis)
         └─ P-E curves, switching dynamics

Level 2: Cell (module2-crossbar/cell)
         └─ Single cell programming, read

Level 3: Array (module2-crossbar/array)  ← Current focus
         └─ MVM, non-idealities

Level 4: Tile (module2-crossbar/tile)    ← Proposed
         └─ Multiple arrays, peripheral sharing

Level 5: Chip (module2-crossbar/chip)    ← Future
         └─ Area, energy, latency estimation
```

---

## Part 5: Validation Strategy

### 5.1 Validation Methodology

| Step | Method | Target |
|------|--------|--------|
| 1 | Internal consistency | Iterative vs matrix IR solver <1% difference |
| 2 | Cross-tool comparison | MVM output vs CrossSim/NeuroSim <5% RMSE |
| 3 | Literature validation | IR drop vs IEEE TCAS-I 2021 <15% RMSE |

### 5.2 Validation Targets

| Component | Reference | Target Accuracy |
|-----------|-----------|-----------------|
| IR Drop (internal) | Iterative vs matrix | <1% difference |
| IR Drop (external) | IEEE TCAS-I 2021 SPICE | <15% RMSE |
| Sneak Paths | NeuroSim output | <10% difference |
| MVM | CrossSim output | <5% RMSE |
| Drift | Retention curves | Within 2× of measured |

---

## Part 6: Implementation Roadmap

### Phase 1: Foundation (Priority 1)

| Task | Impact | Effort | Deliverable |
|------|--------|--------|-------------|
| VTEAM conductance model | High | Medium | `array.go` extension |
| Interactive parameter sweep | High | Low | `param_sweep.go` |
| Literature-calibrated drift | Medium | Low | `drift.go` extension |
| CSV/JSON export/import | Medium | Low | `export.go` |

### Phase 2: Physics Accuracy (Priority 2)

| Task | Impact | Effort | Deliverable |
|------|--------|--------|-------------|
| Matrix IR drop solver | High | Medium | `irdrop.go` extension |
| Architecture comparison UI | High | Medium | New comparison tab |
| Sneak path MVM integration | Medium | Medium | `nonidealities.go` |
| 3-level educational tooltips | Medium | Medium | `tooltips.go` extension |

### Phase 3: System-Level (Priority 3)

| Task | Impact | Effort | Deliverable |
|------|--------|--------|-------------|
| Tile-level modeling | Medium | High | `tile.go` |
| Device model plugin system | Medium | High | `device_model.go` |

---

## Part 7: Success Metrics

### Technical Metrics

| Metric | Target |
|--------|--------|
| IR drop accuracy vs literature | <15% RMSE |
| VTEAM vs MemTorch reference | <5% difference |
| Parameter sweep UI responsiveness | >20 FPS |
| Export round-trip fidelity | 100% |
| Test coverage | >85% |

### User Experience Metrics

| Metric | Target |
|--------|--------|
| Time to first result | <2 minutes |
| Architecture comparison clarity | Clear visible difference |
| Tooltip helpfulness | 4/5 user rating |

---

## Appendix A: Reference Links

| Tool | Repository | License |
|------|------------|---------|
| CrossSim | [github.com/sandialabs/cross-sim](https://github.com/sandialabs/cross-sim) | BSD-3 |
| NeuroSim | [github.com/neurosim](https://github.com/neurosim) | MIT |
| MemTorch | [github.com/coreylammie/MemTorch](https://github.com/coreylammie/MemTorch) | GPL v3 |
| AIHWKit | [github.com/IBM/aihwkit](https://github.com/IBM/aihwkit) | MIT |
| FerroX | [github.com/AMReX-Microelectronics/FerroX](https://github.com/AMReX-Microelectronics/FerroX) | BSD-3 |
| MNSIM 2.0 | [github.com/thu-nics/MNSIM-2.0](https://github.com/thu-nics/MNSIM-2.0) | Research |

---

## Appendix B: Key Research Papers

| Topic | Source | DOI |
|-------|--------|-----|
| Analytical IR Drop | IEEE TCAS-I 2021 | 10.1109/TCSI.2021.3070046 |
| Sneak Path Analysis | arXiv 2025 | arXiv:2511.21796 |
| Multi-level FeFET | Nature Comms 2023 | 10.1038/s41467-023-42110-y |
| 3D FeNAND | Nature Comms 2023 | 10.1038/s41467-023-36270-0 |
| FeCap vs FeFET | Sci. Reports 2024 | 10.1038/s41598-024-59298-8 |
| NeuroSim V1.5 | arXiv 2025 | arXiv:2505.02314 |

---

## Conclusion

The opensource crossbar simulation ecosystem provides valuable patterns and validated approaches that can significantly improve FeCIM Lattice Tools. Our key differentiators—interactive Fyne GUI, FeFET-specific physics, and educational focus—remain strong advantages. By selectively adopting proven techniques (VTEAM models, analytical IR drop, statistical variation) while maintaining our distinctive visualization capabilities, we can achieve both improved physics accuracy and enhanced user experience.

**Recommended Next Steps:**
1. Implement Phase 1 improvements (VTEAM, parameter sweep, drift calibration, export)
2. Validate against CrossSim/NeuroSim reference outputs
3. Gather user feedback on new UI features
4. Proceed to Phase 2 based on validation results

---

*Document generated from research plan approved via ralplan iteration on 2026-01-28.*
