# Crossbar Module Improvement Plan

**Based on analysis of badcrossbar and CrossSim**

*Created: January 2026*

---

## Executive Summary

This plan outlines improvements to Module 2 (Crossbar) based on learnings from two production-grade open-source simulators: **badcrossbar** (UCL) and **CrossSim** (Sandia National Labs). The current implementation is well-engineered for FeCIM-specific physics but lacks key features for production-grade simulation.

---

## Current State Assessment

### What We Have (Strengths)

| Feature | Status | Location |
|---------|--------|----------|
| 30-level quantization (demo baseline; simulation baseline) | ✅ Complete | `array.go` |
| Basic MVM (I = G × V) | ✅ Complete | `array.go:450-517` |
| IR drop (damped Jacobi) | ✅ Working | `nonidealities.go:119-229` |
| 3-cell sneak paths | ✅ Working | `enhanced.go:302-357` |
| Architecture modes (0T1R/1T1R/2T1R) | ✅ Complete | `nonidealities.go` |
| Temperature effects | ✅ Complete | `temperature.go` |
| Drift model | ✅ Complete | `drift.go` |
| GPU acceleration (Vulkan) | ✅ Working | `gpu_mvm.go` |
| DAC/ADC quantization | ✅ Basic | `array.go` |
| Unit tests across module | ✅ | `*_test.go` (see CI) |

### What We're Missing (Gaps)

| Feature | Priority | Source |
|---------|----------|--------|
| Iterative SOR solver | **HIGH** | CrossSim |
| Programming error model | **HIGH** | CrossSim |
| Read noise (per-operation) | **MEDIUM** | CrossSim |
| Multi-hop sneak paths | **MEDIUM** | badcrossbar |
| Sparse matrix support | **MEDIUM** | badcrossbar |
| Non-linear device I-V | **LOW** | CrossSim |
| Differential pair cores | **LOW** | CrossSim |
| SAR/Ramp ADC models | **LOW** | CrossSim |

---

## Phase 1: Core Solver Upgrade (HIGH PRIORITY)

### Goal
Replace damped Jacobi iteration with Successive Over-Relaxation (SOR) for better convergence.

### Current Implementation
```go
// nonidealities.go:119-229 - Damped Jacobi
for iter := 0; iter < config.MaxIterations; iter++ {
    // Simple damping: newV = oldV + damping * (targetV - oldV)
}
```

### Target Implementation
```go
// New: SOR solver with adaptive omega
type SORConfig struct {
    MaxIterations   int     // Default: 100
    Tolerance       float64 // Default: 1e-6
    OmegaInitial    float64 // Default: 1.0 (1.0-1.9 for SOR)
    OmegaMin        float64 // Default: 0.01
    OmegaDecay      float64 // Default: 0.98
    AdaptiveOmega   bool    // Auto-tune omega on divergence
}

func (s *Solver) SolveMVMWithParasitics(inputs []float64) (*MVMResult, error) {
    dV := make([][]float64, s.rows)
    omega := s.config.OmegaInitial

    for iter := 0; iter < s.config.MaxIterations; iter++ {
        // 1. Compute device currents: I = G × V
        Ires := s.computeDeviceCurrents(dV)

        // 2. Cumulative currents (CrossSim approach)
        IsumCol := s.cumsum(Ires, AxisColumns)
        IsumRow := s.cumsumReverse(Ires, AxisRows)

        // 3. Parasitic voltage drops
        VdropsCol := s.cumsum(s.multiplyScalar(s.RpCol, IsumCol), AxisColumns)
        VdropsRow := s.cumsum(s.multiplyScalar(s.RpRow, IsumRow), AxisRows)
        Vpar := s.add(VdropsCol, VdropsRow)

        // 4. Voltage error
        VerrMat := s.subtract(s.subtract(s.Vapplied, Vpar), dV)
        Verr := s.maxAbs(VerrMat)

        // 5. Check convergence
        if Verr < s.config.Tolerance {
            return &MVMResult{
                OutputCurrents: s.sumRows(Ires),
                Iterations:     iter,
                Converged:      true,
                MaxError:       Verr,
            }, nil
        }

        // 6. SOR update with adaptive omega
        dV = s.add(dV, s.multiplyScalar(omega, VerrMat))

        // 7. Detect divergence and reduce omega
        if iter > 5 && Verr > prevErr {
            omega *= s.config.OmegaDecay
            if omega < s.config.OmegaMin {
                return nil, ErrConvergenceFailed
            }
        }
        prevErr = Verr
    }
    return nil, ErrMaxIterationsExceeded
}
```

### Files to Modify
- `module2-crossbar/pkg/crossbar/nonidealities.go` - Replace `AnalyzeIRDropIterative`
- `module2-crossbar/pkg/crossbar/solver.go` - New file for SOR solver
- `module2-crossbar/pkg/crossbar/solver_test.go` - Convergence tests

### Validation
```go
func TestSORConvergence(t *testing.T) {
    testCases := []struct {
        name     string
        size     int
        RpRatio  float64 // Rp / Rmin
        expected int     // Max iterations to converge
    }{
        {"8x8 low parasitic", 8, 0.01, 5},
        {"32x32 medium", 32, 0.1, 15},
        {"64x64 high", 64, 1.0, 50},
        {"128x128 extreme", 128, 2.0, 100},
    }
    // Compare against badcrossbar reference
}
```

### Acceptance Criteria
- [ ] Converges for 128×128 arrays with Rp/Rmin = 1.0
- [ ] Output matches badcrossbar within 0.1% tolerance
- [ ] Iteration count < 50 for typical cases
- [ ] Graceful degradation (returns partial result on timeout)

---

## Phase 2: Device Non-Idealities (HIGH PRIORITY)

### Goal
Add realistic programming error and read noise models from CrossSim.

### 2.1 Programming Error Model

```go
// device_errors.go

type ProgrammingErrorConfig struct {
    Enable    bool
    Model     ErrorModel // NormalProportional, NormalIndependent, Uniform
    Sigma     float64    // Standard deviation (fraction of range)
    Symmetric bool       // Apply to both increase/decrease
}

type ErrorModel int
const (
    NormalIndependent ErrorModel = iota // σ fixed
    NormalProportional                   // σ ∝ G
    NormalInverseProportional            // σ ∝ R
    UniformIndependent
    UniformProportional
)

func ApplyProgrammingError(Gtarget float64, config *ProgrammingErrorConfig) float64 {
    if !config.Enable {
        return Gtarget
    }

    var sigma float64
    switch config.Model {
    case NormalIndependent:
        sigma = config.Sigma
    case NormalProportional:
        sigma = config.Sigma * Gtarget
    case NormalInverseProportional:
        sigma = config.Sigma / Gtarget
    }

    noise := rand.NormFloat64() * sigma
    return math.Max(0, math.Min(1, Gtarget + noise))
}
```

### 2.2 Read Noise Model

```go
type ReadNoiseConfig struct {
    Enable    bool
    Model     NoiseModel // White, Pink (1/f), Thermal
    Sigma     float64    // Per-operation noise magnitude
    Persistent bool      // If true, same noise per read; if false, changes each time
}

func ApplyReadNoise(Gprogrammed float64, config *ReadNoiseConfig) float64 {
    if !config.Enable {
        return Gprogrammed
    }

    noise := rand.NormFloat64() * config.Sigma
    return Gprogrammed * (1 + noise)
}
```

### Files to Create
- `module2-crossbar/pkg/crossbar/device_errors.go`
- `module2-crossbar/pkg/crossbar/device_errors_test.go`

### Acceptance Criteria
- [ ] Programming error reduces accuracy by expected amount (illustrative; confirm with tests)
- [ ] Read noise is per-operation (non-persistent)
- [ ] Models match CrossSim distribution shapes
- [ ] Configurable via `Config` struct

---

## Phase 3: Enhanced Sneak Path Analysis (MEDIUM PRIORITY)

### Goal
Extend 3-cell sneak path model to multi-hop analysis for accuracy.

### Current Limitation
```go
// Only considers 3-cell paths: entry → middle → exit
gPath := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
```

### Target: N-hop Sneak Paths
```go
type SneakPathConfig struct {
    MaxHops        int  // Maximum path length (default: 5)
    MinConductance float64 // Prune paths below threshold
    UseGraphSearch bool // BFS/DFS vs exhaustive
}

func (a *Array) AnalyzeSneakPaths(config *SneakPathConfig) *SneakAnalysis {
    // For small arrays (<32x32): exhaustive search
    // For large arrays: BFS with pruning

    if a.config.Rows * a.config.Cols < 1024 {
        return a.exhaustiveSneakSearch(config)
    }
    return a.bfsSneakSearch(config)
}

func (a *Array) bfsSneakSearch(config *SneakPathConfig) *SneakAnalysis {
    // Breadth-first search for sneak paths
    // Prune paths with conductance below threshold
    // Track top-K highest conductance paths
}
```

### Files to Modify
- `module2-crossbar/pkg/crossbar/sneakpath.go` - Extend analysis
- `module2-crossbar/pkg/crossbar/enhanced.go` - Update integration

### Acceptance Criteria
- [ ] Detects 5-hop paths for 16×16 arrays
- [ ] Graceful degradation (falls back to 3-hop for large arrays)
- [ ] Performance: <100ms for 64×64 analysis
- [ ] Matches badcrossbar's implicit sneak calculation within 5%

---

## Phase 4: ADC/DAC Enhancements (MEDIUM PRIORITY)

### Goal
Add realistic ADC models beyond simple quantization.

### 4.1 Quantizer ADC (Current - Enhanced)
```go
type QuantizerADC struct {
    Bits            int
    Signed          bool
    RangeMin        float64
    RangeMax        float64
    StochasticRound bool // Probabilistic rounding
}
```

### 4.2 SAR ADC Model (New)
```go
type SARADC struct {
    Bits              int
    CapacitorMismatch float64 // σ for cap variation
    ComparatorOffset  float64 // mV offset
    SettlingTime      float64 // ns per bit
}

func (adc *SARADC) Convert(current float64) (code int, latency float64) {
    // Successive approximation with non-idealities
    code = 0
    latency = 0

    for bit := adc.Bits - 1; bit >= 0; bit-- {
        // Add capacitor mismatch
        threshold := adc.getThreshold(bit)
        threshold += rand.NormFloat64() * adc.CapacitorMismatch

        // Add comparator offset
        effectiveCurrent := current + adc.ComparatorOffset

        if effectiveCurrent > threshold {
            code |= (1 << bit)
        }

        latency += adc.SettlingTime
    }
    return
}
```

### Files to Create
- `module2-crossbar/pkg/crossbar/adc_models.go`
- `module2-crossbar/pkg/crossbar/adc_models_test.go`

### Acceptance Criteria
- [ ] SAR ADC matches CrossSim's INL/DNL patterns
- [ ] Latency calculation reflects bit-serial operation
- [ ] Configurable non-idealities (mismatch, offset)

---

## Phase 5: GUI Integration (LOW PRIORITY)

### Goal
Expose new features in Module 2 Fyne GUI.

### New Controls
```go
// gui/controls.go additions

// Non-ideality panel
type NonIdealityControls struct {
    ProgrammingErrorSlider *widget.Slider // 0-10% σ
    ReadNoiseSlider        *widget.Slider // 0-5% σ
    ADCBitsSelect          *widget.Select // 4, 6, 8, 10 bits
    SORToleranceEntry      *widget.Entry  // Scientific notation
}

// Visualization panel
type VisualizationControls struct {
    ShowDeviceCurrents  *widget.Check
    ShowVoltageDropMap  *widget.Check
    ShowSneakPathsOverlay *widget.Check
    AnimateConvergence  *widget.Check // Show SOR iterations
}
```

### New Visualizations
- **Convergence Animation**: Show SOR iterations converging
- **Sneak Path Overlay**: Highlight major sneak paths on heatmap
- **Error Distribution**: Histogram of programming errors
- **Accuracy Waterfall**: Show degradation sources

### Files to Modify
- `module2-crossbar/pkg/gui/controls.go`
- `module2-crossbar/pkg/gui/visualization.go`
- `module2-crossbar/pkg/gui/app.go`

---

## Phase 6: Validation Framework (ONGOING)

### Goal
Ensure accuracy against reference implementations.

### 6.1 badcrossbar Validation
```go
// validation/badcrossbar_test.go

func TestAgainstBadcrossbar(t *testing.T) {
    // Generate test case
    voltages, resistances := generateTestCase(8, 8)

    // Run badcrossbar via Python subprocess
    expected := runBadcrossbar(voltages, resistances, 0.5)

    // Run our solver
    actual := solver.SolveMVMWithParasitics(voltages)

    // Compare
    assert.InDeltaSlice(t, expected.OutputCurrents, actual.OutputCurrents, 0.001)
}

func runBadcrossbar(voltages, resistances [][]float64, r_i float64) *MVMResult {
    // Python subprocess call
    script := `
import badcrossbar
import json
import sys

data = json.load(sys.stdin)
solution = badcrossbar.compute(data['voltages'], data['resistances'], r_i=data['r_i'])
print(json.dumps({'output': solution.currents.output.tolist()}))
`
    // Execute and parse
}
```

### 6.2 SPICE Validation (Small Arrays)
```spice
* validation/spice/crossbar_4x4.cir
.title 4x4 Crossbar Validation

* Word line drivers
VWL0 WL0 0 DC 1.0V
VWL1 WL1 0 DC 0.5V
VWL2 WL2 0 DC 0.8V
VWL3 WL3 0 DC 0.3V

* Parasitic resistance (1 Ohm/cell)
RWL0_0 WL0 N0_0 1
RWL0_1 N0_0 N0_1 1
...

* Crossbar cells (conductances from test case)
R0_0 N0_0 M0_0 {1/G[0,0]}
...

.op
.print dc i(VWL0) i(VWL1) i(VWL2) i(VWL3)
.end
```

### Test Matrix

| Test | Array | Rp | Non-idealities | Expected Match |
|------|-------|-----|----------------|----------------|
| Basic MVM | 8×8 | 0 | None | 100% |
| IR drop | 16×16 | 0.5Ω | None | 99.9% |
| High parasitic | 32×32 | 2.0Ω | None | 99% |
| With noise | 16×16 | 0.5Ω | σ=5% | Statistical |
| Full non-ideal | 64×64 | 1.0Ω | All | 95% |

---

## Implementation Timeline

| Phase | Duration | Dependencies | Deliverables |
|-------|----------|--------------|--------------|
| **Phase 1: SOR Solver** | 1 week | None | `solver.go`, tests |
| **Phase 2: Device Errors** | 1 week | None | `device_errors.go`, tests |
| **Phase 3: Sneak Paths** | 1 week | Phase 1 | Enhanced `sneakpath.go` |
| **Phase 4: ADC Models** | 1 week | None | `adc_models.go`, tests |
| **Phase 5: GUI** | 2 weeks | Phases 1-4 | UI updates |
| **Phase 6: Validation** | Ongoing | All | Test suite |

**Total Estimated Effort:** 7-8 weeks

---

## Success Metrics

| Metric | Current | Target |
|--------|---------|--------|
| badcrossbar match (IR drop) | N/A | >99% |
| Convergence rate (128×128) | ~80% | >99% |
| Iteration count (64×64, Rp=1Ω) | ~100 | <30 |
| Test coverage | 158 tests | 200+ tests |
| Max array size (real-time) | 64×64 | 128×128 |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| SOR doesn't converge for extreme cases | Medium | High | Fallback to damped Jacobi |
| Performance regression | Low | Medium | Benchmark suite |
| Breaking existing tests | Medium | Low | Incremental refactoring |
| badcrossbar API changes | Low | Low | Pin version |

---

## References

1. **badcrossbar**: Joksas & Mehonic, SoftwareX 2020, [DOI](https://doi.org/10.1016/j.softx.2020.100617)
2. **CrossSim**: Sandia National Labs, [GitHub](https://github.com/sandialabs/cross-sim)
3. **SOR Algorithm**: Young, D.M. (1971). Iterative Solution of Large Linear Systems
4. **FeCIM Physics**: Archival conference reference (not validated)

---

## Appendix: Code Snippets from Reference Implementations

### badcrossbar KCL Assembly
```python
# kcl.py - Conductance matrix assembly
g[idx, idx] = 2*g_interconnect + g_device[i,j]
g[idx, idx-1] = -g_interconnect
g[idx, idx+1] = -g_interconnect
g[idx, idx+crossbar_size] = -g_device[i,j]
```

### CrossSim SOR Solver
```python
# NonInterleaved_InputSource.py
while Verr > threshold and iters < max_iters:
    Isum_col = cumsum(Ires, axis=columns)
    Vdrops_col = cumsum(Rp_col * Isum_col[:,::-1], axis=1)[:,::-1]
    VerrMat = V_applied - Vpar - dV
    dV += gamma * VerrMat
```

### CrossSim Device Error
```python
# generic_error.py
class NormalProportionalDevice:
    def programming_error(self, G_target):
        sigma = self.sigma * G_target
        return G_target + np.random.normal(0, sigma)
```
