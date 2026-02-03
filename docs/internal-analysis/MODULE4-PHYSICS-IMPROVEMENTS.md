# Module 4 Physics Improvements - Gap Analysis

> **Note:** Internal analysis note. Values are reported/illustrative and not validated by this codebase.

> **Status**: Analysis Complete | **Priority**: Critical gaps identified

## Executive Summary

Module 4's peripheral circuit simulation uses **simplified linear models** where **nonlinear ferroelectric physics** should apply. This document details five key gap areas with specific file references and recommended fixes.

---

## 1. Conductance Model

### Current State

**File**: `module1-hysteresis/pkg/ferroelectric/material.go:487-496`

```go
func (m *HZOMaterial) DiscreteLevel(level int, totalLevels int) float64 {
    normalizedP := -1.0 + 2.0*float64(level)/float64(totalLevels-1)
    Gmin := 1e-6   // 1 µS - HARDCODED
    Gmax := 100e-6 // 100 µS - HARDCODED
    return Gmin + (Gmax-Gmin)*(normalizedP+1)/2  // LINEAR
}
```

### Identified Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| Linear G(P) | **CRITICAL** | Real FeFET conductance is nonlinear with polarization |
| Hardcoded Gmin/Gmax | HIGH | Ignores `conductance_min_s`/`conductance_max_s` from physics.yaml |
| No Preisach integration | HIGH | Preisach model exists but not used for level calculation |
| No temperature dependence | MEDIUM | Material has `PolarizationAtTemp(T)` but conductance ignores it |

### Recommended Fix

```go
func (m *HZOMaterial) DiscreteLevelPhysics(level int, totalLevels int) float64 {
    // 1. Load from physics.yaml
    cfg, _ := physics.Load()
    Gmin := cfg.Crossbar.ConductanceMinS  // 1e-6 S
    Gmax := cfg.Crossbar.ConductanceMaxS  // 100e-6 S

    // 2. Get polarization for this level
    P := m.levelToPolarization(level, totalLevels)

    // 3. FeFET threshold voltage model
    // ΔVth = -2*Pr*t_fe / (ε₀*ε_fe)
    deltaVth := -2.0 * P * m.Thickness / (epsilon0 * m.Epsilon)

    // 4. Exponential I-V (subthreshold region)
    kT := 0.0259 // at 300K
    return Gmin + (Gmax-Gmin) * (1.0 / (1.0 + math.Exp(-deltaVth/kT)))
}
```

---

## 2. Voltage Calculations

### Current State

**File**: `module4-circuits/pkg/gui/device_state.go:184-224`

```go
func (ds *DeviceState) updateVoltageRanges() {
    Vc := ds.material.CoerciveVoltage()  // Ec × thickness - CORRECT

    safeReadMax := ds.calibParams.FieldMinRatio * Vc  // 0.7 × Vc - OK
    writeMin := Vc                                      // At Vc - OK
    writeMax := ds.calibParams.FieldMaxRatio * Vc     // 2.5 × Vc - OK
}
```

### Identified Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| Ec-to-Vc mapping | OK | Correctly uses `CoerciveVoltage() = Ec × Thickness` |
| Static field ratios | LOW | 0.5/2.5 ratios are reasonable defaults |
| No pulse shaping | HIGH | Real FeFET needs V×t product for partial switching |
| No fatigue drift | MEDIUM | Vc should increase with cycling |

### What's Working

The voltage derivation is **mostly correct**:
- FeCIM HZO: `Ec = 1.0e8 V/m`, `thickness = 10nm` → `Vc = 1.0V`
- Read range: 0 to 0.5V (below switching)
- Write range: 1.0V to 2.5V (above Ec)

### Missing Features

1. **Pulse width dependence**: Partial switching = f(V × t)
2. **Temperature-dependent Vc**: `CoerciveFieldAtTemp(T)` exists but unused
3. **Endurance drift**: Ec increases with cycling

---

## 3. Current Summation (MVM)

### Current State

**File**: `module4-circuits/pkg/gui/device_state.go:433-488`

```go
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
    for r := 0; r < ds.rows; r++ {
        totalCurrent := 0.0
        for c := 0; c < ds.cols; c++ {
            voltage := ds.dacVoltages[c]
            conductanceS := ds.material.DiscreteLevel(level, quantLevels)
            current := conductanceS * 1e6 * voltage  // I = G × V
            totalCurrent += current
        }
        ds.rowCurrents[r] = totalCurrent  // No IR drop, no sneak
    }
}
```

### Identified Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| No IR drop | **CRITICAL** | Module 2 has model, Module 4 ignores it |
| No sneak paths | **CRITICAL** | 0T1R should show 5-20% error |
| Linear I=G×V | HIGH | FeFET has subthreshold vs linear regions |
| No thermal noise | MEDIUM | TIA noise model exists but unused |

### Evidence of Gap

The UI correctly warns about this (`tab_unified.go:1457`):
```go
helpText = "READ: Passive array - sneak currents add 5-20% error."
```
But `Compute()` doesn't actually apply the error.

### Recommended Fix

```go
func (ds *DeviceState) ComputeWithNonidealities(weights [][]int, quantLevels int) {
    // Import from Module 2
    array := crossbar.NewArray(ds.rows, ds.cols)
    array.SetWeights(weightsToFloat(weights, quantLevels))

    for r := 0; r < ds.rows; r++ {
        totalCurrent := 0.0

        for c := 0; c < ds.cols; c++ {
            // Apply IR drop
            irAnalysis := array.AnalyzeIRDrop(ds.dacVoltages, crossbar.DefaultWireParams())
            effectiveV := irAnalysis.EffectiveVoltage[r][c]

            current := conductance * effectiveV
            totalCurrent += current
        }

        // Add sneak current for passive mode
        if ds.isPassive {
            sneakAnalysis := array.AnalyzeSneakPathsWithArch(r, 0, false)
            totalCurrent += sneakAnalysis.TotalSneak * ds.dacVoltages[0]
        }

        ds.rowCurrents[r] = totalCurrent
    }
}
```

---

## 4. ADC/TIA Model

### Current State

**TIA** (`module4-circuits/pkg/peripherals/tia.go:30-44`):
```go
func (t *TIA) Convert(current float64) float64 {
    output := current*t.Gain + t.OutputOffset  // 10kΩ gain
    if output > t.MaxOutputVoltage { output = t.MaxOutputVoltage }
    return output  // No noise
}
```

**ADC** (`module4-circuits/pkg/peripherals/adc.go:47-65`):
```go
func (a *ADC) Convert(voltage float64) int {
    fraction := (voltage - a.VrefLow) / (a.VrefHigh - a.VrefLow)
    level := int(fraction*float64(a.Levels()-1) + 0.5)  // No INL/DNL
    return level
}
```

### Identified Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| Saturation | OK | TIA saturates at 100µA/1V correctly |
| Noise | HIGH | `ConvertWithNoise()` exists but unused |
| INL/DNL | HIGH | `ConvertWithNonlinearity()` exists but unused |
| Range mismatch | LOW | ADC Vref matches TIA output |

### What's Available But Unused

```go
// In tia.go - EXISTS but not called
func (t *TIA) ConvertWithNoise(current float64) float64

// In adc.go - EXISTS but not called
func (a *ADC) ConvertWithNonlinearity(voltage float64) int
```

### Recommended Fix

In `device_state.go:476-483`:
```go
// Current (too optimistic):
ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6)
ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])

// Recommended (realistic):
ds.rowVoltages[r] = ds.tia.ConvertWithNoise(totalCurrent * 1e-6)
ds.rowLevels[r] = ds.adc.ConvertWithNonlinearity(ds.rowVoltages[r])
```

---

## 5. Material Integration

### Current State

Material selection updates voltage ranges (`tab_unified.go:127-139`):
```go
selector := widget.NewSelect(materialNames, func(selected string) {
    ca.deviceState.SetMaterial(m)
    ca.updateDACPresetLabels()      // Updates voltage labels
    ca.updateDACRangeModeLabel()
})
```

### Identified Gaps

| Gap | Severity | Description |
|-----|----------|-------------|
| Ec/Pr for Vc | OK | Correctly uses `Ec × thickness` |
| NumLevels | OK | Material's NumLevels is used |
| Gmin/Gmax | HIGH | Hardcoded, doesn't use physics.yaml |
| FTJ TER ratio | MEDIUM | 140-state material has `ter_ratio: 834` but unused |

### FTJ Special Case

The 140-state FTJ material in physics.yaml:
```yaml
hzo_ftj_140:
  ter_ratio: 834          # Tunneling electroresistance
  gmax_gmin_ratio: 1.0e5  # 10^5 conductance ratio
```

But `DiscreteLevel()` uses fixed 1-100 µS (100:1 ratio) for ALL materials.

---

## Priority Ranking

| Priority | Area | Gap | Impact |
|----------|------|-----|--------|
| **CRITICAL** | Conductance | Linear G(P) should be nonlinear | Wrong physics |
| **CRITICAL** | MVM | No IR drop or sneak path | 5-20% error missing |
| HIGH | Conductance | Hardcoded Gmin/Gmax | Material selection broken |
| HIGH | Conductance | Preisach not integrated | Minor loops ignored |
| HIGH | ADC/TIA | Noise/INL models unused | Too optimistic |
| MEDIUM | Voltage | No pulse width dependence | Partial switching wrong |
| MEDIUM | Material | FTJ TER ratio unused | 140-state behaves as 30-state |
| LOW | Voltage | Temperature-dependent Vc | Edge case |

---

## Files Requiring Changes

| File | Changes |
|------|---------|
| `module1-hysteresis/pkg/ferroelectric/material.go` | Nonlinear G(P), load from config |
| `module4-circuits/pkg/gui/device_state.go` | IR drop, sneak paths, noise models |
| `module4-circuits/pkg/gui/tab_unified.go` | "Realistic physics" toggle |
| `config/physics.yaml` | Values correct, just need to use them |

---

## Implementation Phases

See `docs/plans/module4-plan-improvements.md` for detailed tasks.

### Phase 1 - Critical Physics (Week 1)
- Load Gmin/Gmax from physics.yaml
- Integrate IR drop from Module 2
- Add sneak path calculation for passive mode

### Phase 2 - Nonlinear Conductance (Weeks 2-3)
- FeFET threshold voltage model
- Optional Preisach integration
- Temperature dependence

### Phase 3 - Peripheral Accuracy (Week 4)
- Enable TIA noise
- Enable ADC INL/DNL
- Add UI toggle

### Phase 4 - Material Specialization (Week 5)
- FTJ TER ratio handling
- Per-material non-idealities

---

## Implementation References

### Key Papers for Physics Improvements

| Gap | Recommended Paper | DOI | Key Insight |
|-----|-------------------|-----|-------------|
| Nonlinear G(P) | Physical Reality of Preisach Model | [10.1038/s41467-018-06717-w](https://doi.org/10.1038/s41467-018-06717-w) | Domain distribution, minor loops |
| IR Drop | badcrossbar: Crossbar Simulation | [10.1016/j.softx.2020.100617](https://doi.org/10.1016/j.softx.2020.100617) | Nodal analysis, exact voltage drops |
| Sneak Paths | FeCAP Suppresses Sneak Paths | [10.1186/s40580-024-00463-0](https://doi.org/10.1186/s40580-024-00463-0) | Capacitive vs resistive crossbars |
| TIA Noise | NeuroSim Validation | [10.3389/frai.2021.659060](https://doi.org/10.3389/frai.2021.659060) | Circuit-level noise modeling |
| ADC Power | ADC-less Hybrid CIM | See `04-cim-architectures/` | 82% power in ADC |
| FTJ TER | 140 Analog States in SnS₂ | [10.1002/advs.202308588](https://doi.org/10.1002/advs.202308588) | TER ratio 834:1 |

### Simulation Tool Integration

| Tool | What to Import | How to Use |
|------|----------------|------------|
| **CrossSim** | Device models, noise | Reference for validation |
| **badcrossbar** | IR drop algorithm | Port `compute_output()` to Go |
| **Module 2** | `AnalyzeIRDrop()`, `AnalyzeSneakPaths()` | Direct integration |

### Code Examples from Literature

**Nonlinear FeFET Conductance** (based on Physical Reality of Preisach 2018):
```go
// Replace linear model with sigmoid threshold model
func NonlinearConductance(P, Gmin, Gmax, Pth, k float64) float64 {
    // P = polarization state (-1 to +1)
    // Pth = threshold polarization for conductance onset
    // k = steepness factor (typically 5-10)
    sigmoid := 1.0 / (1.0 + math.Exp(-k*(P-Pth)))
    return Gmin + (Gmax-Gmin)*sigmoid
}
```

**IR Drop Integration** (from badcrossbar approach):
```go
// Use Module 2's existing IR drop analysis
func (ds *DeviceState) ComputeWithIRDrop(weights [][]int) {
    array := crossbar.NewArray(ds.rows, ds.cols)
    irResult := array.AnalyzeIRDrop(ds.dacVoltages, crossbar.DefaultWireParams())
    for r := 0; r < ds.rows; r++ {
        for c := 0; c < ds.cols; c++ {
            effectiveV := irResult.EffectiveVoltage[r][c] // Reduced by IR drop
            // ... use effectiveV instead of ds.dacVoltages[c]
        }
    }
}
```

---

## Related Documents

- [circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md) — CIM operation physics
- [cim-circuits.md](cim-circuits.md) — Peripheral circuit specifications
- [crossbar-arrays.md](crossbar-arrays.md) — Crossbar architecture and non-idealities
- [hysteresis-physics.md](hysteresis-physics.md) — Preisach model implementation
