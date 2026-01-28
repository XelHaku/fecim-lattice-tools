# Module 4 Physics Improvements Implementation Plan

**Document Version:** 1.0
**Date:** 2026-01-27
**Status:** Ready for Implementation

---

## Executive Summary

Module 4 (Peripheral Circuits) simulates FeCIM crossbar array operations but currently has significant physics accuracy gaps. This plan addresses four critical issues:

1. **Hardcoded conductance values** ignore `physics.yaml` configuration
2. **Missing IR drop and sneak path models** for passive (0T1R) architecture
3. **Linear G(P) model** should use nonlinear FeFET threshold voltage model
4. **Unused noise models** - TIA and ADC have noise methods that aren't called

**Expected Outcome:** Reduce simulation error from ~20% to <5% by integrating existing Module 2 physics models and properly using peripheral noise models.

---

## Phase 1: Critical Physics (PRIORITY: CRITICAL)

**Duration:** 1 week
**Goal:** Fix fundamental physics errors causing >10% simulation inaccuracy

### Task 1.1: Load Gmin/Gmax from physics.yaml

**Current Problem:**

In `module1-hysteresis/pkg/ferroelectric/material.go` line 492-495:

```go
// DiscreteLevel returns the conductance for a given state level (0-29).
func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64 {
    // Map to conductance (1-100 uS range)
    Gmin := 1e-6   // 1 uS  <-- HARDCODED
    Gmax := 100e-6 // 100 uS <-- HARDCODED
    return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
}
```

**physics.yaml has correct values (lines 400-402):**
```yaml
crossbar:
  conductance_min_s: 1.0e-6     # 1 uS (Gmin)
  conductance_max_s: 100.0e-6   # 100 uS (Gmax)
```

**Files to Modify:**

| File | Changes |
|------|---------|
| `module1-hysteresis/pkg/ferroelectric/material.go` | Add fields, load from config |
| `config/physics/physics.go` | Ensure `CrossbarConfig` exports Gmin/Gmax |

**Implementation:**

1. Add conductance fields to `HZOMaterial`:
```go
type HZOMaterial struct {
    // ... existing fields ...

    // Conductance range (loaded from physics.yaml or defaults)
    ConductanceMinS float64 // Gmin in Siemens
    ConductanceMaxS float64 // Gmax in Siemens
}
```

2. Update `MaterialFromConfig()` to load conductance:
```go
func MaterialFromConfig(m *physics.Material, cfg *physics.Config) *HZOMaterial {
    mat := &HZOMaterial{
        // ... existing mappings ...
    }

    // Load conductance from crossbar config
    if cfg != nil && cfg.Crossbar != nil {
        mat.ConductanceMinS = cfg.Crossbar.ConductanceMinS
        mat.ConductanceMaxS = cfg.Crossbar.ConductanceMaxS
    } else {
        // Fallback defaults
        mat.ConductanceMinS = 1e-6   // 1 uS
        mat.ConductanceMaxS = 100e-6 // 100 uS
    }

    return mat
}
```

3. Update `DiscreteLevel()` to use loaded values:
```go
func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64 {
    normalizedP := -1.0 + 2.0*float64(level)/float64(totalLevels-1)

    Gmin := m.ConductanceMinS
    Gmax := m.ConductanceMaxS
    if Gmin == 0 {
        Gmin = 1e-6 // Fallback
    }
    if Gmax == 0 {
        Gmax = 100e-6 // Fallback
    }

    return Gmin + (Gmax-Gmin)*(normalizedP+1)/2
}
```

4. Update hardcoded material constructors (`DefaultHZO()`, `FeCIMMaterial()`, etc.) to set default conductance values.

**Test Requirements:**
```go
func TestDiscreteLevelUsesConfiguredConductance(t *testing.T) {
    // Load physics config
    cfg, _ := physics.Load()
    mat := MaterialFromConfig(cfg.GetMaterial("fecim_hzo"), cfg)

    // Level 0 should equal Gmin
    g0 := mat.DiscreteLevel(0, 30)
    assert.InDelta(t, cfg.Crossbar.ConductanceMinS, g0, 1e-9)

    // Level 29 should equal Gmax
    g29 := mat.DiscreteLevel(29, 30)
    assert.InDelta(t, cfg.Crossbar.ConductanceMaxS, g29, 1e-9)
}
```

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] `DiscreteLevel()` uses `physics.yaml` values when available
- [ ] Hardcoded constructors set reasonable defaults
- [ ] Changing `physics.yaml` changes simulation behavior
- [ ] Unit tests pass

---

### Task 1.2: Integrate IR Drop Model from Module 2

**Current Problem:**

`device_state.go` `Compute()` method (lines 433-488) applies DAC voltages directly without accounting for IR drop in word/bit lines. For large arrays (64x64+), this causes 1-5% voltage error at far corners.

**Module 2 has working solution** in `nonidealities.go`:
- `AnalyzeIRDrop()` (lines 43-127) - full IR drop analysis
- `MVMWithIRDrop()` (lines 311-341) - MVM with IR drop correction

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Import Module 2 IR drop, add to Compute() |
| `module4-circuits/pkg/peripherals/analysis.go` | Add wrapper for Module 2 IR drop (optional) |

**Implementation:**

1. Add IR drop fields to `DeviceState`:
```go
type DeviceState struct {
    // ... existing fields ...

    // IR drop tracking
    irDropEnabled   bool
    wireParams      *crossbar.WireParams
    effectiveVolts  [][]float64 // Per-cell effective voltage
    irDropAnalysis  *crossbar.IRDropAnalysis
}
```

2. Add IR drop initialization in `NewDeviceState()`:
```go
ds.wireParams = crossbar.DefaultWireParams()
ds.irDropEnabled = false // Default off for performance
```

3. Update `Compute()` to apply IR drop when enabled:
```go
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
    // Calculate IR drop if enabled
    if ds.irDropEnabled && ds.isPassive {
        ds.calculateIRDrop()
    }

    for r := 0; r < ds.rows; r++ {
        if !ds.activeRows[r] {
            // ... unchanged ...
            continue
        }

        totalCurrent := 0.0
        for c := 0; c < ds.cols; c++ {
            // Get effective voltage (with IR drop if enabled)
            voltage := ds.dacVoltages[c]
            if ds.irDropEnabled && ds.effectiveVolts != nil {
                voltage = ds.effectiveVolts[r][c]
            }

            if voltage < 0.01 {
                continue
            }

            // ... rest of computation unchanged ...
        }

        ds.rowCurrents[r] = totalCurrent
        // ... TIA/ADC unchanged ...
    }
}
```

4. Add helper method:
```go
func (ds *DeviceState) calculateIRDrop() {
    // Create temporary crossbar array for analysis
    // This is a simplified approach - could optimize later

    ds.effectiveVolts = make([][]float64, ds.rows)
    for r := range ds.effectiveVolts {
        ds.effectiveVolts[r] = make([]float64, ds.cols)
    }

    // Simplified analytical model (from nonidealities.go)
    for r := 0; r < ds.rows; r++ {
        rowCurrent := ds.estimateRowCurrent(r)
        for c := 0; c < ds.cols; c++ {
            // Word line drop increases with column (from left driver)
            wlDrop := float64(c) * ds.wireParams.RwordLine * rowCurrent
            wlV := ds.dacVoltages[c] - wlDrop

            // Bit line drop increases with row (from top sense amp)
            colCurrent := ds.estimateColCurrent(c)
            blDrop := float64(r) * ds.wireParams.RbitLine * colCurrent

            // Effective voltage = WL - BL
            ds.effectiveVolts[r][c] = wlV - blDrop
            if ds.effectiveVolts[r][c] < 0 {
                ds.effectiveVolts[r][c] = 0
            }
        }
    }
}

func (ds *DeviceState) estimateRowCurrent(row int) float64 {
    // Estimate current draw for this row
    var current float64
    for c := 0; c < ds.cols; c++ {
        if ds.dacVoltages[c] < 0.01 {
            continue
        }
        // Use approximate average conductance
        avgG := (ds.material.ConductanceMinS + ds.material.ConductanceMaxS) / 2
        current += avgG * ds.dacVoltages[c]
    }
    return current
}

func (ds *DeviceState) estimateColCurrent(col int) float64 {
    // Similar estimation for column current
    var current float64
    if ds.dacVoltages[col] < 0.01 {
        return 0
    }
    for r := 0; r < ds.rows; r++ {
        if !ds.activeRows[r] {
            continue
        }
        avgG := (ds.material.ConductanceMinS + ds.material.ConductanceMaxS) / 2
        current += avgG * ds.dacVoltages[col]
    }
    return current
}
```

5. Add UI toggle:
```go
func (ds *DeviceState) SetIRDropEnabled(enabled bool) {
    ds.irDropEnabled = enabled
}

func (ds *DeviceState) IsIRDropEnabled() bool {
    return ds.irDropEnabled
}

func (ds *DeviceState) GetIRDropAtCell(row, col int) float64 {
    if ds.effectiveVolts == nil || !ds.irDropEnabled {
        return 0
    }
    return ds.dacVoltages[col] - ds.effectiveVolts[row][col]
}
```

**Test Requirements:**
```go
func TestIRDropScalesWithArraySize(t *testing.T) {
    // Small array should have minimal drop
    ds8 := NewDeviceState(8, 8, DefaultTIA(), DefaultADC())
    ds8.SetIRDropEnabled(true)
    ds8.SetAllDACVoltages(1.0)
    ds8.SetWLAll()
    ds8.Compute(weights8, 30)
    drop8 := ds8.GetIRDropAtCell(7, 7)

    // Large array should have significant drop
    ds64 := NewDeviceState(64, 64, DefaultTIA(), DefaultADC())
    ds64.SetIRDropEnabled(true)
    ds64.SetAllDACVoltages(1.0)
    ds64.SetWLAll()
    ds64.Compute(weights64, 30)
    drop64 := ds64.GetIRDropAtCell(63, 63)

    assert.Greater(t, drop64, drop8)
    assert.InDelta(t, 0.01, drop8, 0.01)  // <1% for 8x8
    assert.InDelta(t, 0.05, drop64, 0.03) // 2-8% for 64x64
}
```

**Estimated Complexity:** Medium (M)

**Acceptance Criteria:**
- [ ] IR drop toggle in UI (defaults OFF)
- [ ] 64x64 array shows 2-5% voltage drop at far corner (row 63, col 63)
- [ ] 8x8 array shows <1% drop
- [ ] 1T1R mode: IR drop only affects active row
- [ ] Performance acceptable (no noticeable lag)

---

### Task 1.3: Add Sneak Path Calculation for Passive Mode

**Current Problem:**

`device_state.go` computes current per-row independently. In 0T1R (passive) mode, unselected cells create parallel sneak paths that add 5-20% error to sensed current.

**Module 2 has working solution** in `nonidealities.go`:
- `AnalyzeSneakPaths()` (lines 172-182)
- `AnalyzeSneakPathsWithArch()` (lines 186-308)

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Add sneak path computation to Compute() |

**Implementation:**

1. Add sneak path tracking:
```go
type DeviceState struct {
    // ... existing fields ...

    // Sneak path tracking
    sneakEnabled    bool
    sneakCurrents   []float64 // Per-row sneak current (uA)
    sneakFractions  []float64 // Per-row sneak as fraction of signal
}
```

2. Initialize in `NewDeviceState()`:
```go
ds.sneakEnabled = false
ds.sneakCurrents = make([]float64, rows)
ds.sneakFractions = make([]float64, rows)
```

3. Add sneak path computation:
```go
func (ds *DeviceState) computeSneakPaths(weights [][]int, quantLevels int) {
    if !ds.sneakEnabled || !ds.isPassive {
        // Clear sneak data
        for i := range ds.sneakCurrents {
            ds.sneakCurrents[i] = 0
            ds.sneakFractions[i] = 0
        }
        return
    }

    // For each active row, calculate sneak current from other rows
    for r := 0; r < ds.rows; r++ {
        if !ds.activeRows[r] {
            continue
        }

        signalCurrent := ds.rowCurrents[r]
        if signalCurrent < 1e-9 {
            continue
        }

        // Calculate sneak paths through other rows
        var sneakTotal float64
        for otherRow := 0; otherRow < ds.rows; otherRow++ {
            if otherRow == r || !ds.activeRows[otherRow] {
                continue
            }

            // For each column, calculate sneak contribution
            for c := 0; c < ds.cols; c++ {
                if ds.dacVoltages[c] < 0.01 {
                    continue
                }

                // Sneak path: current row cell -> column -> other row cell -> return
                // Simplified model: parallel path through adjacent cells
                level := 0
                if r < len(weights) && c < len(weights[r]) {
                    level = weights[r][c]
                }
                otherLevel := 0
                if otherRow < len(weights) && c < len(weights[otherRow]) {
                    otherLevel = weights[otherRow][c]
                }

                // Series conductance of two cells
                g1 := ds.material.DiscreteLevel(level, quantLevels)
                g2 := ds.material.DiscreteLevel(otherLevel, quantLevels)

                if g1 > 0 && g2 > 0 {
                    // Series combination: Gseries = g1*g2 / (g1+g2)
                    gSeries := (g1 * g2) / (g1 + g2)
                    sneakCurrent := gSeries * ds.dacVoltages[c] * 1e6 // to uA
                    sneakTotal += sneakCurrent
                }
            }
        }

        // Scale by number of parallel paths (simplified)
        sneakTotal = sneakTotal / float64(ds.rows-1) * 0.3 // Empirical factor

        ds.sneakCurrents[r] = sneakTotal
        ds.sneakFractions[r] = sneakTotal / signalCurrent
    }
}
```

4. Integrate into `Compute()`:
```go
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
    // ... existing computation ...

    // After computing row currents, calculate sneak paths
    ds.computeSneakPaths(weights, quantLevels)

    // Add sneak current to totals when enabled
    if ds.sneakEnabled && ds.isPassive {
        for r := 0; r < ds.rows; r++ {
            ds.rowCurrents[r] += ds.sneakCurrents[r]

            // Re-convert through TIA/ADC with sneak-corrupted current
            if ds.tia != nil {
                ds.rowVoltages[r] = ds.tia.Convert(ds.rowCurrents[r] * 1e-6)
            }
            if ds.adc != nil {
                ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])
            }
        }
    }
}
```

5. Add accessors:
```go
func (ds *DeviceState) SetSneakPathEnabled(enabled bool) {
    ds.sneakEnabled = enabled
}

func (ds *DeviceState) IsSneakPathEnabled() bool {
    return ds.sneakEnabled
}

func (ds *DeviceState) GetSneakFraction(row int) float64 {
    if row >= 0 && row < ds.rows {
        return ds.sneakFractions[row]
    }
    return 0
}

func (ds *DeviceState) GetSneakCurrent(row int) float64 {
    if row >= 0 && row < ds.rows {
        return ds.sneakCurrents[row]
    }
    return 0
}
```

**Test Requirements:**
```go
func TestSneakPathMagnitude(t *testing.T) {
    ds := NewDeviceState(8, 8, DefaultTIA(), DefaultADC())
    ds.SetPassiveMode(true)
    ds.SetSneakPathEnabled(true)
    ds.SetAllDACVoltages(1.0)

    // Create uniform weight matrix
    weights := make([][]int, 8)
    for r := range weights {
        weights[r] = make([]int, 8)
        for c := range weights[r] {
            weights[r][c] = 15 // Mid-level
        }
    }

    ds.Compute(weights, 30)

    // Sneak fraction should be 5-20% in passive mode
    for r := 0; r < 8; r++ {
        fraction := ds.GetSneakFraction(r)
        assert.Greater(t, fraction, 0.05, "Sneak too low")
        assert.Less(t, fraction, 0.20, "Sneak too high")
    }
}

func TestNoSneakIn1T1R(t *testing.T) {
    ds := NewDeviceState(8, 8, DefaultTIA(), DefaultADC())
    ds.SetPassiveMode(false) // 1T1R mode
    ds.SetSneakPathEnabled(true)

    // ... setup ...
    ds.Compute(weights, 30)

    for r := 0; r < 8; r++ {
        assert.Equal(t, 0.0, ds.GetSneakFraction(r))
    }
}
```

**Estimated Complexity:** Medium (M)

**Acceptance Criteria:**
- [ ] Sneak path toggle in UI (defaults OFF)
- [ ] 0T1R mode shows 5-20% sneak current fraction
- [ ] 1T1R mode shows 0% sneak (transistors isolate)
- [ ] Sneak fraction displayed per row in visualization
- [ ] Combined IR drop + sneak path scenario works

---

## Phase 2: Nonlinear Conductance (PRIORITY: HIGH)

**Duration:** 1-2 weeks
**Goal:** Replace linear G(P) with physics-accurate FeFET model

### Task 2.1: Implement FeFET Threshold Voltage Model

**Current Problem:**

`DiscreteLevel()` uses linear interpolation between Gmin and Gmax:
```go
return Gmin + (Gmax-Gmin)*(normalizedP+1)/2  // Linear
```

Real FeFET conductance is exponential with threshold voltage:
```
G(Vth) = Gmin * exp((Vgs - Vth) / (n * Vt))
```

where Vth shifts with polarization state.

**Files to Modify:**

| File | Changes |
|------|---------|
| `module1-hysteresis/pkg/ferroelectric/material.go` | Add `DiscreteLevelFeFET()` method |
| `module4-circuits/pkg/gui/device_state.go` | Add conductance model selector |

**Implementation:**

1. Add FeFET-specific parameters to `HZOMaterial`:
```go
type HZOMaterial struct {
    // ... existing fields ...

    // FeFET parameters
    VthMin         float64 // Threshold voltage at -Pr (high G state)
    VthMax         float64 // Threshold voltage at +Pr (low G state)
    SubthresholdN  float64 // Subthreshold slope factor (1.0-2.5)
    ThermalVoltage float64 // kT/q at operating temperature (26mV at 300K)
}
```

2. Add FeFET conductance model:
```go
// DiscreteLevelFeFET returns conductance using FeFET exponential model.
// More accurate than linear model for actual FeFET behavior.
//
// Model: G(P) = Gmin * exp((Vgs - Vth(P)) / (n * Vt))
// where Vth(P) varies linearly with polarization state
func (m *HZOMaterial) DiscreteLevelFeFET(level int, totalLevels int, Vgs float64) float64 {
    // Get material conductance limits
    Gmin := m.ConductanceMinS
    Gmax := m.ConductanceMaxS
    if Gmin == 0 {
        Gmin = 1e-6
    }
    if Gmax == 0 {
        Gmax = 100e-6
    }

    // Default FeFET parameters if not set
    vthMin := m.VthMin
    vthMax := m.VthMax
    n := m.SubthresholdN
    vt := m.ThermalVoltage

    if vthMin == 0 && vthMax == 0 {
        // Derive from Gmax/Gmin ratio
        // Gmax/Gmin = exp((VthMax - VthMin) / (n * Vt))
        // Typical memory window: 1V
        vthMin = 0.3 // V (high conductance state)
        vthMax = 1.3 // V (low conductance state)
    }
    if n == 0 {
        n = 1.5 // Typical subthreshold factor
    }
    if vt == 0 {
        vt = 0.026 // 26mV at 300K
    }
    if Vgs == 0 {
        Vgs = 0.8 // Default read voltage (mid-threshold)
    }

    // Calculate threshold voltage for this level
    // Level 0 = -Pr (Vth = VthMin, high G)
    // Level max = +Pr (Vth = VthMax, low G)
    fraction := float64(level) / float64(totalLevels-1)
    vth := vthMin + (vthMax-vthMin)*fraction

    // Exponential conductance model
    exponent := (Vgs - vth) / (n * vt)

    // Clamp exponent to avoid overflow
    if exponent > 20 {
        exponent = 20
    }
    if exponent < -20 {
        exponent = -20
    }

    G := Gmin * math.Exp(exponent)

    // Clamp to physical limits
    if G < Gmin {
        G = Gmin
    }
    if G > Gmax {
        G = Gmax
    }

    return G
}
```

3. Add conductance model enum to `DeviceState`:
```go
type ConductanceModel int

const (
    ConductanceLinear ConductanceModel = iota
    ConductanceExponential
    ConductanceFeFET
)

type DeviceState struct {
    // ... existing fields ...
    conductanceModel ConductanceModel
    readVoltage      float64 // Vgs for FeFET model
}
```

4. Update `Compute()` to use selected model:
```go
func (ds *DeviceState) getConductance(level, totalLevels int) float64 {
    switch ds.conductanceModel {
    case ConductanceLinear:
        return ds.material.DiscreteLevel(level, totalLevels)
    case ConductanceExponential:
        // Geometric interpolation
        Gmin := ds.material.ConductanceMinS
        Gmax := ds.material.ConductanceMaxS
        if Gmin == 0 {
            Gmin = 1e-6
        }
        if Gmax == 0 {
            Gmax = 100e-6
        }
        ratio := math.Log(Gmax / Gmin)
        fraction := float64(level) / float64(totalLevels-1)
        return Gmin * math.Exp(ratio * fraction)
    case ConductanceFeFET:
        return ds.material.DiscreteLevelFeFET(level, totalLevels, ds.readVoltage)
    default:
        return ds.material.DiscreteLevel(level, totalLevels)
    }
}
```

**Test Requirements:**
```go
func TestFeFETConductanceIsNonlinear(t *testing.T) {
    mat := FeCIMMaterial()
    mat.VthMin = 0.3
    mat.VthMax = 1.3
    mat.SubthresholdN = 1.5
    mat.ThermalVoltage = 0.026
    mat.ConductanceMinS = 1e-6
    mat.ConductanceMaxS = 100e-6

    // Get conductances at levels 0, 15, 29
    g0 := mat.DiscreteLevelFeFET(0, 30, 0.8)
    g15 := mat.DiscreteLevelFeFET(15, 30, 0.8)
    g29 := mat.DiscreteLevelFeFET(29, 30, 0.8)

    // Should be exponentially spaced, not linear
    // Linear midpoint: (g0 + g29) / 2
    // Geometric midpoint: sqrt(g0 * g29)
    linearMid := (g0 + g29) / 2
    geoMid := math.Sqrt(g0 * g29)

    // FeFET should be closer to geometric
    assert.Less(t, math.Abs(g15-geoMid), math.Abs(g15-linearMid))
}
```

**Estimated Complexity:** Medium (M)

**Acceptance Criteria:**
- [ ] FeFET model produces exponential G(level) curve
- [ ] Conductance model selector in UI (Linear/Exponential/FeFET)
- [ ] Level 0 still maps to Gmin
- [ ] Level 29 still maps to Gmax
- [ ] Mid-levels are geometrically spaced (not linearly)

---

### Task 2.2: Integrate Preisach Model for Level Calculation

**Current Problem:**

Level calculation uses simple voltage-to-level mapping. The Preisach model in Module 1 provides history-dependent polarization switching that affects actual programmed level.

**Module 1 has Preisach model** in `preisach.go`:
- `PreisachModel.Update(E float64) float64` returns polarization for applied field
- `PreisachModel.DiscreteStates(N int)` returns N discrete polarization values

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Optional Preisach integration for write modeling |

**Implementation:**

1. Add optional Preisach model per cell:
```go
type CellPreisach struct {
    model      *ferroelectric.PreisachModel
    lastWriteE float64
}

type DeviceState struct {
    // ... existing fields ...
    preisachEnabled bool
    cellPreisach    [][]*CellPreisach // Per-cell Preisach models
}
```

2. Add Preisach-aware write method:
```go
func (ds *DeviceState) WriteWithPreisach(row, col int, writeVoltage float64) int {
    if !ds.preisachEnabled || ds.cellPreisach == nil {
        // Fallback to simple level calculation
        return ds.calculateLevelFromVoltage(writeVoltage)
    }

    // Get cell's Preisach model
    cp := ds.cellPreisach[row][col]
    if cp == nil {
        cp = &CellPreisach{
            model: ferroelectric.NewPreisachModel(ds.material),
        }
        ds.cellPreisach[row][col] = cp
    }

    // Convert voltage to electric field
    E := writeVoltage / ds.material.Thickness

    // Update Preisach model
    P := cp.model.Update(E)
    cp.lastWriteE = E

    // Convert polarization to level
    normalizedP := P / ds.material.Ps
    level := int((normalizedP + 1) / 2 * float64(ds.material.GetNumLevels()-1))

    if level < 0 {
        level = 0
    }
    if level >= ds.material.GetNumLevels() {
        level = ds.material.GetNumLevels() - 1
    }

    return level
}
```

**Note:** Full Preisach integration is optional for Phase 2. The primary benefit is demonstrating minor loop effects during repeated partial writes.

**Estimated Complexity:** Medium-Large (M-L)

**Acceptance Criteria:**
- [ ] Preisach toggle available (defaults OFF for performance)
- [ ] Minor loop behavior visible when enabled
- [ ] Write history affects subsequent writes to same cell
- [ ] Reset clears Preisach history

---

### Task 2.3: Add Temperature-Dependent Parameters

**Current Problem:**

All conductance calculations use room temperature values. Real FeFETs show significant temperature dependence.

**physics.yaml has temperature parameters (lines 104-107):**
```yaml
temp_coeff_ec: -2.0e5       # dEc/dT (V/m/K)
temp_coeff_pr: -5.0e-5      # dPr/dT (C/m²/K)
```

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Add temperature field, apply to conductance |
| `module1-hysteresis/pkg/ferroelectric/material.go` | Add `ConductanceAtTemp()` method |

**Implementation:**

1. Add temperature methods to `HZOMaterial`:
```go
// ConductanceAtTemp returns temperature-adjusted conductance.
// Higher temperature: increased leakage (Gmin increases)
// Lower temperature: sharper switching
func (m *HZOMaterial) ConductanceAtTemp(G, T float64) float64 {
    Tref := 300.0 // Reference temperature (K)

    // Leakage increases exponentially with temperature
    // Typical activation energy for leakage: 0.3-0.5 eV
    Ea := 0.4 // eV
    kB := 8.617e-5 // eV/K

    factor := math.Exp(Ea / kB * (1/Tref - 1/T))

    // Only affects Gmin significantly
    // Gmax dominated by channel, less temp sensitive
    return G * (1 + 0.1*(factor-1)) // 10% of exponential effect
}
```

2. Add temperature to `DeviceState`:
```go
type DeviceState struct {
    // ... existing fields ...
    temperatureK float64 // Operating temperature (default 300K)
}

func (ds *DeviceState) SetTemperature(T float64) {
    if T < 4 {
        T = 4 // Minimum: cryogenic
    }
    if T > 500 {
        T = 500 // Maximum: high temp
    }
    ds.temperatureK = T
    ds.updateVoltageRanges() // Temperature affects Ec, Pr
}
```

3. Update `getConductance()` to apply temperature:
```go
func (ds *DeviceState) getConductance(level, totalLevels int) float64 {
    G := ds.material.DiscreteLevel(level, totalLevels) // or FeFET model

    if ds.temperatureK != 0 && ds.temperatureK != 300 {
        G = ds.material.ConductanceAtTemp(G, ds.temperatureK)
    }

    return G
}
```

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] Temperature slider in UI (77K to 400K)
- [ ] Presets: Cryogenic (77K), Room (300K), Automotive (400K)
- [ ] Higher temperature shows increased leakage
- [ ] Cryogenic shows enhanced Pr (via existing material methods)

---

## Phase 3: Peripheral Accuracy (PRIORITY: HIGH)

**Duration:** 1 week
**Goal:** Use existing noise models in simulation path

### Task 3.1: Use TIA Noise Model in Simulation

**Current Problem:**

`device_state.go` line 476-478:
```go
if ds.tia != nil {
    ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6) // Uses ideal Convert()
}
```

TIA has `ConvertWithNoise()` that adds thermal noise, but it's never called.

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Add noise toggle, use ConvertWithNoise() |

**Implementation:**

1. Add noise toggle:
```go
type DeviceState struct {
    // ... existing fields ...
    realisticNoiseEnabled bool
}
```

2. Update TIA call in `Compute()`:
```go
if ds.tia != nil {
    if ds.realisticNoiseEnabled {
        ds.rowVoltages[r] = ds.tia.ConvertWithNoise(totalCurrent * 1e-6)
    } else {
        ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6)
    }
}
```

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] Noise toggle in UI (defaults OFF)
- [ ] Noise visible in output values when enabled
- [ ] SNR calculation displayed

---

### Task 3.2: Use ADC INL/DNL Model in Simulation

**Current Problem:**

`device_state.go` line 481-483:
```go
if ds.adc != nil {
    ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r]) // Uses ideal Convert()
}
```

ADC has `ConvertWithNonlinearity()` that adds INL/DNL errors, but it's never called.

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/device_state.go` | Use ConvertWithNonlinearity() when realistic mode |

**Implementation:**

Update ADC call in `Compute()`:
```go
if ds.adc != nil {
    if ds.realisticNoiseEnabled {
        ds.rowLevels[r] = ds.adc.ConvertWithNonlinearity(ds.rowVoltages[r])
    } else {
        ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])
    }
}
```

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] Same toggle as TIA noise
- [ ] ADC levels show occasional off-by-one errors when enabled
- [ ] ENOB displayed in reference panel

---

### Task 3.3: Add "Ideal vs Realistic" UI Toggle

**Files to Modify:**

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/tab_unified.go` | Add toggle to control panel |
| `module4-circuits/pkg/gui/tab_operations.go` | Add toggle to operations view |

**Implementation:**

1. Create unified control:
```go
func (ca *CircuitsApp) createRealisticToggle() fyne.CanvasObject {
    toggle := widget.NewCheck("Realistic Mode", func(checked bool) {
        ca.mu.Lock()
        ca.deviceState.SetRealisticNoiseEnabled(checked)
        ca.deviceState.SetIRDropEnabled(checked && ca.deviceState.IsPassiveMode())
        ca.deviceState.SetSneakPathEnabled(checked && ca.deviceState.IsPassiveMode())
        ca.mu.Unlock()
        ca.refreshComputation()
    })

    help := widget.NewLabel("Adds TIA noise, ADC nonlinearity, IR drop, sneak paths")
    help.TextStyle.Italic = true

    return container.NewVBox(toggle, help)
}
```

2. Add to settings panel or mode selector area.

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] Single "Realistic Mode" toggle
- [ ] Enables: TIA noise, ADC INL/DNL, IR drop (if passive), sneak paths (if passive)
- [ ] Visual indication of which effects are active
- [ ] Performance acceptable

---

## Phase 4: Material Specialization (PRIORITY: MEDIUM)

**Duration:** 3-5 days
**Goal:** Handle special material properties correctly

### Task 4.1: Handle FTJ TER Ratio for 140-State Material

**Current Problem:**

`hzo_ftj_140` material in `physics.yaml` has 140 analog states and special TER (Tunneling Electroresistance) ratio of 834. Current code doesn't use TER for conductance calculation.

**physics.yaml (lines 324-327):**
```yaml
ter_ratio: 834              # Tunneling electroresistance ratio
gmax_gmin_ratio: 1.0e5      # >10^5 conductance ratio
```

**Files to Modify:**

| File | Changes |
|------|---------|
| `module1-hysteresis/pkg/ferroelectric/material.go` | Add FTJ-specific conductance model |
| `config/physics/physics.go` | Expose TER ratio in config struct |

**Implementation:**

1. Add FTJ fields to `HZOMaterial`:
```go
type HZOMaterial struct {
    // ... existing fields ...
    TERRatio     float64 // Tunneling electroresistance ratio (FTJ only)
    IsFTJ        bool    // Is this a Ferroelectric Tunnel Junction?
}
```

2. Add FTJ conductance model:
```go
// DiscreteLevelFTJ returns conductance using FTJ tunneling model.
// FTJ uses polarization-dependent tunneling barrier height.
func (m *HZOMaterial) DiscreteLevelFTJ(level int, totalLevels int) float64 {
    if !m.IsFTJ || m.TERRatio == 0 {
        return m.DiscreteLevel(level, totalLevels) // Fallback
    }

    // FTJ: exponential dependence on polarization
    // TER = G_high / G_low = exp(2 * sqrt(2m*phi) * d / hbar * delta_phi)
    // Simplified: G(P) = Gmin * TER^(P/Ps)

    Gmin := m.ConductanceMinS
    if Gmin == 0 {
        Gmin = 1e-9 // FTJ Gmin is typically lower
    }

    normalizedP := float64(level) / float64(totalLevels-1)

    // Exponential interpolation using TER ratio
    G := Gmin * math.Pow(m.TERRatio, normalizedP)

    // Clamp to reasonable limits
    Gmax := m.ConductanceMaxS
    if Gmax == 0 {
        Gmax = Gmin * m.TERRatio
    }
    if G > Gmax {
        G = Gmax
    }

    return G
}
```

3. Update `MaterialFromConfig()` to set FTJ flag:
```go
// Check for FTJ-specific parameters
if m.TERRatio > 0 {
    mat.TERRatio = m.TERRatio
    mat.IsFTJ = true
}
```

4. Update `getConductance()` in `DeviceState`:
```go
func (ds *DeviceState) getConductance(level, totalLevels int) float64 {
    if ds.material.IsFTJ {
        return ds.material.DiscreteLevelFTJ(level, totalLevels)
    }

    // ... existing logic ...
}
```

**Estimated Complexity:** Small-Medium (S-M)

**Acceptance Criteria:**
- [ ] FTJ 140-state material uses TER-based model
- [ ] Conductance ratio matches TER (Gmax/Gmin ~ 834)
- [ ] Non-FTJ materials unchanged
- [ ] 140 states visualized correctly in GUI

---

### Task 4.2: Add Material-Specific Non-Idealities

**Current Problem:**

All materials use same non-ideality parameters. AlScN with high Ec (5 MV/cm) needs different handling than HZO.

**Files to Modify:**

| File | Changes |
|------|---------|
| `config/physics.yaml` | Add per-material non-ideality parameters |
| `config/physics/physics.go` | Parse material-specific parameters |

**Implementation:**

Add per-material non-ideality section:
```yaml
materials:
  alscn:
    # ... existing params ...
    non_idealities:
      switching_sigma_v: 0.1      # Higher sigma for high-Ec material
      endurance_factor: 1.0       # Same endurance
      retention_factor: 1.2       # Better retention
```

**Estimated Complexity:** Small (S)

**Acceptance Criteria:**
- [ ] AlScN shows wider switching distribution
- [ ] FTJ shows different noise characteristics
- [ ] Material-specific parameters load correctly

---

## Dependencies

```
Phase 1:
  Task 1.1 (Gmin/Gmax) ─┬─→ Task 1.2 (IR Drop) ─┬─→ Task 1.3 (Sneak)
                        │                        │
                        v                        v
                   Task 2.1 (FeFET)         Task 3.1 (TIA Noise)
                        │                        │
                        v                        v
                   Task 2.3 (Temp)          Task 3.2 (ADC INL)
                                                 │
                                                 v
                                            Task 3.3 (UI Toggle)

Phase 4:
  Task 4.1 (FTJ) ─→ Task 4.2 (Material Non-Idealities)
```

**No cross-module dependencies** - all changes are within Module 4 and Module 1.

---

## Risk Analysis

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Performance regression from IR drop/sneak calculation | Medium | Medium | Make features optional, optimize for small arrays first |
| Breaking existing UI | Medium | High | Comprehensive unit tests, feature flags |
| Physics accuracy debates | Low | Low | Document sources in code comments, make models configurable |
| Complexity overwhelming users | Medium | Medium | Hide advanced features behind "Realistic Mode" toggle |
| Integration with Module 2 models | Low | Medium | Use wrapper functions, don't modify Module 2 directly |

---

## Success Criteria

| Phase | Metric | Current | Target | How to Measure |
|-------|--------|---------|--------|----------------|
| 1 | Conductance uses config | No | Yes | Unit test verifies physics.yaml values used |
| 1 | IR drop at 64x64 corner | 0% | 2-5% | Measure effective voltage vs DAC voltage |
| 1 | Sneak fraction in passive mode | 0% | 5-20% | Compare row current with/without sneak |
| 2 | G(level) curve shape | Linear | Exponential | Compare midpoint to geometric mean |
| 3 | TIA noise visible | No | Yes | Observe output variation |
| 3 | ADC INL/DNL effects | No | Yes | Observe occasional level errors |
| 4 | FTJ 140 states work | Untested | Yes | All 140 levels produce unique conductance |

---

## Commit Strategy

Each phase should be committed separately:

1. **Phase 1.1:** `fix(material): load Gmin/Gmax from physics.yaml`
2. **Phase 1.2:** `feat(device): add IR drop model to device simulation`
3. **Phase 1.3:** `feat(device): add sneak path calculation for passive mode`
4. **Phase 2.1:** `feat(material): add FeFET threshold voltage conductance model`
5. **Phase 2.3:** `feat(device): add temperature-dependent conductance`
6. **Phase 3.1-3.3:** `feat(device): add realistic mode toggle for noise/nonlinearity`
7. **Phase 4.1:** `feat(material): add FTJ TER-based conductance model`

---

## References

- Module 2 IR Drop: `module2-crossbar/pkg/crossbar/nonidealities.go` lines 43-127
- Module 2 Sneak Paths: `module2-crossbar/pkg/crossbar/nonidealities.go` lines 159-308
- Preisach Model: `module1-hysteresis/pkg/ferroelectric/preisach.go`
- TIA Noise: `module4-circuits/pkg/peripherals/tia.go` lines 47-59
- ADC Nonlinearity: `module4-circuits/pkg/peripherals/adc.go` lines 68-81
- Physics Config: `config/physics.yaml`
- Existing Improvement Plan: `docs/peripheral-circuits/module4-plan-improvements.md`
