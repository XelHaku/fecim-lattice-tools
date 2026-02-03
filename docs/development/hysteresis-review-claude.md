# Hysteresis Module Technical Review

**Date:** January 31, 2026
**Reviewer:** Claude (Opus 4.5)
**Module:** `module1-hysteresis/`
**Scope:** Physics accuracy, algorithm flow, ISPP implementation, materials, UX

> **Note:** Internal review memo. Ratings and claims are subjective and not independently verified.

---

## Executive Summary

The hysteresis module is a ferroelectric simulation system implementing the Preisach hysteresis model for HfO₂-ZrO₂ (HZO) materials. The module demonstrates strong physics foundations with reported parameter ranges, comprehensive test coverage (count may vary), and thread-safe GUI design. However, several areas warrant attention:

| Area | Rating | Key Finding |
|------|--------|-------------|
| **Physics Accuracy** | ★★★★☆ | MayergoyzPreisach is correct; simplified model has history issues |
| **Material Parameters** | ★★★★★ | Parameters compared against reported literature ranges |
| **ISPP Algorithm** | ★★★★☆ | Solid servo control; voltage/field unit confusion |
| **Code Architecture** | ★★★★★ | Excellent separation of concerns, thread-safe |
| **UX/GUI** | ★★★☆☆ | Information overload; needs progressive disclosure |
| **Test Coverage** | ★★★★★ | Test suite includes physics validation (count may vary) |

---

## Table of Contents

1. [Module Architecture](#1-module-architecture)
2. [Physics Analysis](#2-physics-analysis)
3. [ISPP Algorithm Review](#3-ispp-algorithm-review)
4. [Material Parameters Verification](#4-material-parameters-verification)
5. [UX/GUI Analysis](#5-uxgui-analysis)
6. [Recommendations](#6-recommendations)
7. [References](#7-references)

---

## 1. Module Architecture

### Package Structure

```
module1-hysteresis/
├── cmd/hysteresis/main.go          # Standalone entry point
├── pkg/
│   ├── ferroelectric/              # Physics layer
│   │   ├── preisach.go             # Simplified Preisach model
│   │   ├── preisach_advanced.go    # Full Mayergoyz implementation
│   │   ├── material.go             # HZOMaterial re-exports
│   │   └── render.go               # ASCII visualization
│   ├── algo/
│   │   └── calibration.go          # Level-to-voltage mapping
│   ├── controller/
│   │   └── writer.go               # ISPP state machine
│   ├── simulation/
│   │   └── engine.go               # Time-stepping simulation
│   ├── gui/
│   │   ├── gui.go                  # Main UI construction
│   │   ├── simulation.go           # Calibration & WRD logic
│   │   ├── controls.go             # Control panel
│   │   ├── info.go                 # Information panel
│   │   ├── embedded.go             # Unified visualizer interface
│   │   └── widgets/                # Custom Fyne widgets
│   │       ├── peplot.go           # P-E curve plotter
│   │       ├── cell.go             # Memory cell visualizer
│   │       ├── level.go            # 30-level indicator
│   │       ├── phase.go            # State machine indicator
│   │       └── ispp_visualization.go
│   ├── render/                     # Vulkan/ASCII rendering
│   └── tui/                        # Terminal UI alternative
└── shaders/                        # GPU compute shaders
```

### Key Data Structures

| Structure | Location | Purpose |
|-----------|----------|---------|
| `MayergoyzPreisach` | `preisach_advanced.go:36-92` | Full hysteron-based model with temperature/strain |
| `HZOMaterial` | `shared/physics/material.go:36-93` | Material parameters (Pr, Ps, Ec, dynamics) |
| `WriteController` | `controller/writer.go:25-63` | ISPP state machine |
| `CalibrationManager` | `algo/calibration.go:9-25` | Binary search voltage calibration |
| `Engine` | `simulation/engine.go` | Thread-safe simulation loop |

### Data Flow

```
User Input → Waveform Generation → Preisach.Update(E) → Polarization
    ↓                                                       ↓
Mode Selection                                      Quantize to Level
    ↓                                                       ↓
ISPP Controller ← CalibrationManager ← Level Error ← GUI Display
```

### Thread Safety

- `sync.RWMutex` protects all shared state in Engine and App
- All UI updates wrapped in `fyne.Do()` for main thread execution
- State copies returned from getters to prevent data races

---

## 2. Physics Analysis

### 2.1 Preisach Model Implementation

#### MayergoyzPreisach (Full Model) - CORRECT

**Location:** `pkg/ferroelectric/preisach_advanced.go`

| Aspect | Implementation | Status |
|--------|----------------|--------|
| Hysteron constraint α > β | Lines 199-208 | ✅ Correct per Mayergoyz (1991) |
| Gaussian distribution | Lines 234-270 | ✅ Proper 2D bivariate formulation |
| Lorentzian distribution | Lines 278-310 | ✅ Correct implementation |
| Hysteron switching | Lines 421-428 | ✅ UP when E≥α, DOWN when E≤β |
| Normalization | Lines 265-270 | ✅ Weights sum to Ps |
| Temperature dependence | Lines 312-328 | ✅ Ec(T) = Ec₀(1-T/Tc)^0.5 |
| Substrate strain | Lines 354-397 | ✅ Electrostrictive coupling |

**Physics Extensions Implemented:**
1. **KAI (Kolmogorov-Avrami-Ishibashi)** dynamics (line 725): `P(t) = Ps(1 - exp(-(t/τ)^n))`
2. **Merz Law (NLS)** switching (lines 666-695): `τ(E) = τ₀ exp(Ea/|E|)`
3. **Wake-up/Fatigue** degradation (lines 257-259)
4. **State export/import** for continuity

#### PreisachModel (Simplified) - HAS ISSUES

**Location:** `pkg/ferroelectric/preisach.go`

| Issue | Location | Severity | Description |
|-------|----------|----------|-------------|
| History correction oversimplified | Lines 157-189 | HIGH | Only clamps to saturation, no branch interpolation |
| effectiveCoerciveField returns constant | Lines 136-140 | HIGH | Returns `p.EcMean` instead of integrating distribution |
| Turning point wipe-out logic | Lines 144-155 | MEDIUM | May be backwards for some ascending cases |

**Impact:** The simplified model does not produce physically correct minor loops. It should only be used for fast visualization, not accurate physics simulation.

### 2.2 Coercive Field (Ec) Modeling

```go
// Temperature correction - CORRECT
Ec_eff = Ec0 * math.Pow(1-T/Tc, TempExponent)  // TempExponent = 0.5

// Strain correction - CORRECT
Ec_eff = Ec_temp * (1 + strainFactor * strain)
```

**References checked:** Haun et al. 1987, Materlik et al. 2015

### 2.3 Remanent Polarization (Pr) Calculation

**Distribution width from Pr/Ps ratio (lines 119-134):**
```go
sigma = Ec * (1.2 - Pr/Ps)  // Narrower distribution for higher squareness
```

This physics-based approach correctly produces:
- High Pr/Ps → narrow distribution → square loops
- Low Pr/Ps → wide distribution → slanted loops

### 2.4 KAI Dynamics Issue

**Location:** `preisach_advanced.go:715-726`

**Problem:** `SimulateDomainSwitching` uses constant `tau := m.material.Tau` but KAI's tau should be field-dependent via Merz law:

```go
// Current (constant tau - only valid at fixed high field):
tau := m.material.Tau

// Should be (field-dependent):
tauNLS := m.Tau0NLS * math.Exp(m.EaNLS / math.Abs(E))
```

---

## 3. ISPP Algorithm Review

### 3.1 State Machine

**Location:** `pkg/controller/writer.go`

```
StateIdle → StateApply → StateWait → StateVerify
                ↑                         ↓
                └──── StateAdjust ←───────┘
                            ↓
                   StateSuccess / StateFailed
```

### 3.2 Servo Control Features

| Feature | Location | Description |
|---------|----------|-------------|
| Oscillation damping | Lines 239-244 | 0.5x dampening on sign change |
| Stuck-state recovery | Lines 252-264 | Aggressive "kick" when below threshold |
| Slope estimation | Lines 266-280 | Two-point extrapolation for faster convergence |

### 3.3 Voltage/Field Unit Confusion

**Critical Issue** in `writer.go`:

```go
// Line 254-260 - EcEst calculated from Emax
EcEst := wc.Emax * 0.4  // This is a field estimate

// But CurrentVoltage is named as voltage
wc.CurrentVoltage = wc.Emax + EcEst  // Mixing concepts
```

**Clarification needed:** If `CurrentVoltage` represents voltage (V), then comparisons against `Ec` (V/m) are incorrect. The proper relationship is:
```
Vc = Ec × thickness
```

### 3.4 Calibration Algorithm

**Location:** `pkg/algo/calibration.go`

| Feature | Implementation | Status |
|---------|----------------|--------|
| Binary search | Lines 69-71 | ✅ Midpoint bounds convergence |
| Oscillation damping | Lines 189-191 | ✅ 0.7 old + 0.3 new EMA |
| Monotonicity cascade | Lines 98-127 | ✅ Enforces ordering between levels |
| Relaxation compensation | Lines 19-22 | ✅ Accounts for post-pulse drift |

---

## 4. Material Parameters Verification

### 4.1 HZO Parameters vs Literature

| Parameter | Code Value | Literature | Source | Status |
|-----------|------------|------------|--------|--------|
| Pr (RT) | 25 µC/cm² | 15-34 µC/cm² | Nature Commun. 2025 | ✅ |
| Ps (RT) | 30 µC/cm² | 25-40 µC/cm² | Various | ✅ |
| Ec (RT) | 1.2 MV/cm | 0.6-1.5 MV/cm | Nature Commun. 2025 | ✅ |
| Pr (4K) | 75 µC/cm² | ~75 µC/cm² | Adv. Elec. Mat. 2024 | ✅ |
| Ec (4K) | 1.5 MV/cm | 1.2-2.0 MV/cm | Measured | ✅ |
| Tc | 723 K | 450-500°C | Standard HZO | ✅ |
| KAI n | 2.0 | 1.8-2.5 | 2D growth theory | ✅ |
| τ₀ NLS | 1e-10 to 1e-12 s | 100fs-1ps | Typical | ✅ |

### 4.2 Material Variants Supported

| Material | Factory Function | Notes |
|----------|------------------|-------|
| Default HZO | `DefaultHZO()` | Si-doped, standard |
| High-Pr HZO | `HighPrHZO()` | Optimized Pr |
| Cryogenic | `CryogenicHZO()` | 4K operation |
| Low-Power | `LowPowerHZO()` | Reduced Ec |
| High-Endurance | `HighEnduranceHZO()` | 10¹² cycles |

### 4.3 YAML Configuration

Materials can also be loaded from `config/physics/physics.yaml` for runtime flexibility.

---

## 5. UX/GUI Analysis

### 5.1 Layout Assessment

**Three-column adaptive layout:**
```
┌─────────────┬─────────────────────────────────┬──────────────┐
│   Info      │                                 │   Controls   │
│   Panel     │         P-E Plot                │              │
│   (scroll)  │                                 │   (scroll)   │
│             │   ┌─────────────────────────┐   │              │
│  - Stats    │   │    Level Indicator      │   │  - Waveform  │
│  - Log      │   │    (30 levels)          │   │  - Material  │
│  - ISPP     │   └─────────────────────────┘   │  - Levels    │
└─────────────┴─────────────────────────────────┴──────────────┘
```

### 5.2 High-Priority UX Issues

| Issue | Location | Impact | Recommendation |
|-------|----------|--------|----------------|
| **No calibration feedback** | `controls.go:204-211` | Users wait without indication | Add spinner/progress |
| **Silent input validation** | `controls.go:161-166` | Confusion when clamped | Show validation message |
| **Information overload** | `gui.go:616-626` | 7+ widgets in left panel | Add collapsible sections |
| **Calibration blocks UI** | `controls.go:203-211` | Unresponsive feel | Run in goroutine |

### 5.3 Medium-Priority UX Issues

| Issue | Location | Recommendation |
|-------|----------|----------------|
| Phase indicator text 8pt | `widgets/phase.go:214` | Increase to 10-11pt |
| Controls lack grouping | `controls.go:364-378` | Add visual separators |
| No P-E plot zoom | `widgets/peplot.go` | Add mouse wheel zoom |
| ELI5 button not prominent | `controls.go:282-286` | Move to header icon |
| ISPP widget expensive refresh | `widgets/ispp_visualization.go:322` | Cache histogram bars |

### 5.4 Fyne-Specific Issues

| Issue | Severity | Notes |
|-------|----------|-------|
| `fyne.Do()` usage | OK | Correctly applied throughout |
| Animation cleanup | LOW | `polBarAnim` should stop when hidden |
| Calibration cache unbounded | LOW | Consider LRU eviction |
| Hard-coded pixel sizes | LOW | May not scale on HiDPI |

### 5.5 UX Strengths

- **Educational content:** ELI5 dialog provides excellent explanations
- **Visual feedback:** Phase indicator, stability indicator, color-coded levels
- **Thread safety:** All UI updates properly wrapped
- **Adaptive layout:** Handles window resizing gracefully

---

## 6. Recommendations

### 6.1 High Priority

| # | Recommendation | Effort | Impact | Location |
|---|----------------|--------|--------|----------|
| 1 | **Fix simplified Preisach history** | Medium | High | `preisach.go:157-189` |
| 2 | **Clarify voltage vs field units** | Low | Medium | `writer.go` throughout |
| 3 | **Add calibration loading indicator** | Low | High | `controls.go:204-211` |
| 4 | **Add input validation feedback** | Low | High | `controls.go:161-166` |
| 5 | **Move calibration to background** | Medium | High | `controls.go:203-211` |

### 6.2 Medium Priority

| # | Recommendation | Effort | Impact |
|---|----------------|--------|--------|
| 6 | Make KAI tau field-dependent | Medium | Medium |
| 7 | Group controls into sections | Low | Medium |
| 8 | Add tooltips to complex widgets | Low | Medium |
| 9 | Optimize ISPP widget refresh | Medium | Medium |
| 10 | Add zoom/pan to P-E plot | Medium | Medium |

### 6.3 Low Priority

| # | Recommendation | Effort | Impact |
|---|----------------|--------|--------|
| 11 | Add Everett integral option | High | Low |
| 12 | Support light theme | Low | Low |
| 13 | Animation cleanup on destroy | Low | Low |
| 14 | Add physics validation summary doc | Low | Medium |

### 6.4 Implementation Guidance

**To fix simplified Preisach history (`preisach.go:157-189`):**
```go
// Current: only clamps to saturation
// Should: interpolate between branches
func (p *PreisachModel) applyHistoryCorrection(P, E float64) float64 {
    // Find enclosing branch pair from turning points
    // Interpolate based on E position within minor loop
    // Use FORCs (First Order Reversal Curves) for accurate interpolation
}
```

**To clarify voltage/field in ISPP:**
```go
// Option A: Rename to clarify these are fields
type WriteController struct {
    CurrentField float64  // V/m, not V
    // ...
}

// Option B: Add explicit thickness and convert
currentVoltage := currentField * thickness  // V = (V/m) * m
```

---

## 7. References

### Code References

| File | Key Lines | Content |
|------|-----------|---------|
| `preisach_advanced.go` | 36-92 | MayergoyzPreisach struct |
| `preisach_advanced.go` | 119-134 | Distribution width calculation |
| `preisach_advanced.go` | 312-328 | Temperature/strain Ec correction |
| `preisach_advanced.go` | 421-428 | Hysteron switching logic |
| `preisach_advanced.go` | 715-726 | KAI dynamics |
| `preisach.go` | 157-189 | Simplified history (needs fix) |
| `writer.go` | 14-22 | ISPP state definitions |
| `writer.go` | 254-260 | EcEst calculation (unit confusion) |
| `calibration.go` | 46-155 | Binary search calibration |
| `material.go` | 101-130 | DefaultHZO parameters |

### Literature References

1. Mayergoyz, I.D. (1991). *Mathematical Models of Hysteresis*. Springer.
2. Park et al. (2015). *Ferroelectricity in Si-doped HfO₂*. Applied Physics Letters.
3. Nature Communications (2025). HZO P-E measurements.
4. Advanced Electronic Materials (2024). Cryogenic HZO properties.
5. IEEE IRPS (2022). Endurance characterization.
6. Nano Letters (2024). V:HfO₂ 10¹² cycle endurance.

### Test Coverage

- **179 test functions** across 20 test files
- Key validation tests:
  - `preisach_physics_test.go` - Comprehensive physics validation
  - `literature_validation_test.go` - Literature comparison
  - `golden_regression_test.go` - Golden file regression

---

## Appendix A: Algorithm Flowcharts

### A.1 ISPP Write Flow

```
START
  ↓
Load calibration voltage for target level
  ↓
Apply pulse (StateApply)
  ↓
Wait for settling (StateWait)
  ↓
Read back polarization (StateVerify)
  ↓
┌──────────────────────────────┐
│ Level within tolerance?      │
│                              │
│  YES → StateSuccess → END    │
│                              │
│  NO  → Calculate error       │
│         ↓                    │
│     Oscillation detected?    │
│       YES → Dampen step      │
│       NO  → Binary search    │
│         ↓                    │
│     Update voltage           │
│         ↓                    │
│     Max iterations?          │
│       YES → StateFailed      │
│       NO  → StateApply       │
└──────────────────────────────┘
```

### A.2 Preisach Update Flow

```
Input: E (electric field)
  ↓
For each hysteron h in grid:
  ├── If E >= h.Alpha AND h.State == -1:
  │     h.State = +1 (switch UP)
  │
  ├── If E <= h.Beta AND h.State == +1:
  │     h.State = -1 (switch DOWN)
  │
  └── Accumulate: P += h.Weight * h.State
  ↓
Apply fatigue factor
  ↓
Clamp to [-Ps, +Ps]
  ↓
Output: P (polarization)
```

---

## Appendix B: Test Commands

```bash
# Run all hysteresis tests
go test ./module1-hysteresis/...

# Run physics validation tests
go test ./module1-hysteresis/pkg/ferroelectric -run Physics

# Run with verbose output
go test ./module1-hysteresis/pkg/ferroelectric -v

# Run golden regression tests
go test ./module1-hysteresis/pkg/ferroelectric -run Golden

# Check coverage
go test ./module1-hysteresis/... -cover
```

---

*End of Review*
