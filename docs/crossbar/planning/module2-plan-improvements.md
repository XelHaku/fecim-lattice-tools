# Module 2 Improvement Plan

**Crossbar Array Physics Enhancement Roadmap**

---

## Overview

| Field | Value |
|-------|-------|
| **Module** | module2-crossbar |
| **Goal** | Improve MVM accuracy and add missing physics models |
| **Scope** | Core crossbar simulation, non-idealities, GUI visualization |
| **Estimated Effort** | 5-7 weeks total |

---

## Current State Assessment

### What Works Well

| Feature | Implementation | Quality |
|---------|----------------|---------|
| **30-level quantization** | `QuantizeToLevels()` | Excellent - matches demo baseline |
| **MVM physics** | Ohm's Law + KCL summation | Correct |
| **DAC/ADC quantization** | Configurable bits (4-12) | Good |
| **IR drop model** | Cumulative WL/BL resistance | Good |
| **Sneak path model** | 3-cell series paths | Good |
| **Architecture awareness** | 0T1R vs 1T1R isolation | Good |
| **Drift simulation** | Log-time thermal model | Basic |
| **Differential arrays** | G+ - G- for signed weights | Complete |
| **Write-verify** | Iterative programming loop | Basic |
| **Energy estimation** | Per-component breakdown | Literature-based |

### What's Missing or Simplified

| Gap | Current State | Impact |
|-----|---------------|--------|
| **Linear conductance** | G = Gmin + k×level | 10-20% error at extremes |
| **No thermal model** | Uniform 300K | Can't show cryo/automotive |
| **Assumed drift coefficient** | 0.001 (no source) | Accuracy unknown |
| **No write disturb** | Only read disturb | Passive write accuracy wrong |
| **No process variation** | Only device noise | Systematic errors ignored |
| **Static sneak analysis** | Single-cell selection | Full MVM sneak not computed |
| **No switching statistics** | Deterministic write | Noise floor understated |
| **No interconnect capacitance** | DC-only | High-speed effects ignored |
| **No half-select model** | Full-select only | Passive mode incomplete |
| **No endurance tracking** | SwitchingCount unused | Fatigue not shown |

---

## Phase 1: Critical Physics Improvements

**Duration:** 2-3 weeks
**Priority:** HIGH
**Goal:** Fix fundamental accuracy issues

### Task 1.1: Nonlinear Conductance Model

**Files:** `module2-crossbar/pkg/crossbar/array.go`

**Problem:**
Current code (line 136-137):
```go
gPhys := (10e-6 + gNorm*90e-6) // 10-100 µS range (LINEAR)
```

Real FeFETs have exponential/logarithmic G(V) relationship.

**Changes:**

1. Add conductance model type:
   ```go
   type ConductanceModel int
   const (
       ConductanceLinear ConductanceModel = iota
       ConductanceExponential
       ConductanceLookup  // From calibration data
   )
   ```

2. Add to `Config`:
   ```go
   ConductanceModel ConductanceModel
   ConductanceTable []float64 // For lookup model
   ```

3. Implement `GetPhysicalConductance(normalized float64) float64`:
   ```go
   func (a *Array) GetPhysicalConductance(gNorm float64) float64 {
       gMin, gMax := 10e-6, 100e-6
       switch a.config.ConductanceModel {
       case ConductanceExponential:
           // G = Gmin * exp(ln(Gmax/Gmin) * gNorm)
           ratio := gMax / gMin
           return gMin * math.Exp(math.Log(ratio) * gNorm)
       case ConductanceLookup:
           level := int(gNorm * 29)
           if level < len(a.config.ConductanceTable) {
               return a.config.ConductanceTable[level]
           }
           fallthrough
       default:
           return gMin + gNorm*(gMax-gMin)
       }
   }
   ```

4. Update `estimateCurrent()` and `estimateColumnCurrent()` to use new function

5. Add GUI selector for conductance model

**Acceptance Criteria:**
- [ ] Level 0 → Gmin, Level 29 → Gmax
- [ ] Exponential midpoint = geometric mean
- [ ] GUI toggle between linear/exponential
- [ ] Unit tests pass

---

### Task 1.2: Full MVM Sneak Path Integration

**Files:** `module2-crossbar/pkg/crossbar/enhanced.go`

**Problem:**
Current `computeSneakCurrentForRow()` uses simplified model with fixed sneak factor.
Does not account for actual array state during full MVM.

**Changes:**

1. Implement `computeFullMVMSneak()`:
   ```go
   func (a *Array) computeFullMVMSneak(input []float64, opts *MVMOptions) []float64 {
       sneakPerRow := make([]float64, a.config.Rows)

       if opts.Is1T1R() {
           return sneakPerRow // Zero sneak for 1T1R
       }

       // For each output row, sum sneak from all unintended paths
       for targetRow := 0; targetRow < a.config.Rows; targetRow++ {
           for col := 0; col < len(input); col++ {
               // Three-cell sneak paths through every other row
               for srcRow := 0; srcRow < a.config.Rows; srcRow++ {
                   if srcRow == targetRow {
                       continue
                   }
                   // Path: input[col] → cell[srcRow][col] → cell[srcRow][j] → cell[targetRow][j]
                   for j := 0; j < a.config.Cols; j++ {
                       if j == col {
                           continue
                       }
                       g1 := a.cells[srcRow][col].Conductance
                       g2 := a.cells[srcRow][j].Conductance
                       g3 := a.cells[targetRow][j].Conductance
                       if g1 > 0.01 && g2 > 0.01 && g3 > 0.01 {
                           gPath := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
                           sneakPerRow[targetRow] += input[col] * gPath
                       }
                   }
               }
           }
       }
       return sneakPerRow
   }
   ```

2. Replace simplified model in `MVMWithNonIdealities()`

3. Add sneak calculation toggle (full vs simplified) for performance

**Acceptance Criteria:**
- [ ] Full sneak calculation for small arrays (≤32×32)
- [ ] Simplified fallback for large arrays
- [ ] Sneak magnitude matches literature (5-20% in passive)
- [ ] 1T1R shows ~0% sneak

---

### Task 1.3: Write Disturb Model (Half-Select)

**Files:** `module2-crossbar/pkg/crossbar/array.go`, `enhanced.go`

**Problem:**
Current model ignores half-select disturb during write operations.
In passive arrays, cells sharing row OR column with target see V/2.

**Changes:**

1. Add disturb tracking to `Cell`:
   ```go
   type Cell struct {
       Conductance    float64
       NoiseFactor    float64
       SwitchingCount int64
       HalfSelectCount int64    // NEW
       DisturbShift   float64   // NEW: accumulated drift from V/2 exposure
   }
   ```

2. Implement `ProgramWeightWithDisturb()`:
   ```go
   func (a *Array) ProgramWeightWithDisturb(row, col int, weight float64,
                                            Vc float64, isPassive bool) error {
       // Program target cell
       if err := a.ProgramWeight(row, col, weight); err != nil {
           return err
       }

       if !isPassive {
           return nil // 1T1R: no disturb
       }

       // Half-select disturb on same row (different columns)
       for j := 0; j < a.config.Cols; j++ {
           if j == col {
               continue
           }
           a.cells[row][j].HalfSelectCount++
           // Small disturb if V/2 > 0.3*Vc (disturb threshold)
           halfSelectRatio := 0.5 // V/2 scheme
           if halfSelectRatio > 0.3 {
               disturb := 0.001 * (halfSelectRatio - 0.3) / 0.7
               a.cells[row][j].DisturbShift += disturb
           }
       }

       // Half-select disturb on same column (different rows)
       for i := 0; i < a.config.Rows; i++ {
           if i == row {
               continue
           }
           a.cells[i][col].HalfSelectCount++
           halfSelectRatio := 0.5
           if halfSelectRatio > 0.3 {
               disturb := 0.001 * (halfSelectRatio - 0.3) / 0.7
               a.cells[i][col].DisturbShift += disturb
           }
       }

       return nil
   }
   ```

3. Apply accumulated disturb during MVM reads

4. Add disturb visualization to GUI heatmap

**Acceptance Criteria:**
- [ ] Half-selected cells accumulate disturb count
- [ ] Disturb affects conductance after many writes
- [ ] 1T1R mode shows no disturb
- [ ] GUI visualizes disturb heatmap

---

### Task 1.4: Temperature-Dependent Parameters

**Files:** `module2-crossbar/pkg/crossbar/nonidealities.go`, `drift.go`

**Problem:**
Current code uses fixed 300K temperature. Wire resistance TCR exists but other parameters don't scale.

**Changes:**

1. Create `temperature.go`:
   ```go
   type TemperatureEffects struct {
       AmbientK float64
   }

   // AdjustedWireResistance applies TCR to wire resistance
   func (t *TemperatureEffects) AdjustedWireResistance(R0 float64) float64 {
       TCR := 0.00393 // Copper
       return R0 * (1.0 + TCR*(t.AmbientK-300.0))
   }

   // AdjustedConductanceRange scales Gmin/Gmax with temperature
   func (t *TemperatureEffects) AdjustedConductanceRange(gMin, gMax float64) (float64, float64) {
       // Higher temp: narrower window (thermal noise increases)
       // Cryogenic: wider window (enhanced properties)
       if t.AmbientK < 100 {
           // Cryogenic enhancement
           factor := 1.0 + 0.5*(100-t.AmbientK)/100
           return gMin / factor, gMax * factor
       }
       // High temp degradation
       factor := 1.0 - 0.1*(t.AmbientK-300)/100
       return gMin * factor, gMax * factor
   }

   // AdjustedDriftRate scales drift with temperature
   func (t *TemperatureEffects) AdjustedDriftRate(driftCoeff float64) float64 {
       // Arrhenius: rate ∝ exp(-Ea/kT)
       kB := 1.38e-23
       Ea := 0.5 * 1.6e-19 // 0.5 eV in Joules
       refRate := math.Exp(-Ea / (kB * 300))
       newRate := math.Exp(-Ea / (kB * t.AmbientK))
       return driftCoeff * (newRate / refRate)
   }
   ```

2. Add temperature to `MVMOptions` (already exists, expand usage)

3. Integrate temperature effects into IR drop and drift calculations

4. Add temperature slider to GUI with presets:
   - Cryogenic (77K)
   - Room (300K)
   - Automotive (400K)

**Acceptance Criteria:**
- [ ] Wire resistance scales with temperature
- [ ] Drift rate scales with Arrhenius model
- [ ] Cryogenic mode shows enhanced performance
- [ ] Temperature slider in GUI

---

## Phase 2: Enhanced Realism

**Duration:** 2-3 weeks
**Priority:** MEDIUM
**Goal:** Add secondary physics effects for educational value

### Task 2.1: Switching Statistics

**Files:** `module2-crossbar/pkg/crossbar/array.go`

**Changes:**

1. Add statistics parameters:
   ```go
   type WriteStatistics struct {
       Enabled  bool
       VthSigma float64 // Threshold voltage sigma (normalized)
       RNG      *rand.Rand
   }
   ```

2. Implement `ProgramWeightWithVariation()`:
   ```go
   func (a *Array) ProgramWeightWithVariation(row, col int, targetLevel int,
                                               stats *WriteStatistics) (int, error) {
       if !stats.Enabled {
           return targetLevel, a.ProgramWeight(row, col, float64(targetLevel)/29.0)
       }

       // Add Gaussian noise to target level
       noise := stats.RNG.NormFloat64() * stats.VthSigma * 29
       actualLevel := int(math.Round(float64(targetLevel) + noise))

       // Clamp
       if actualLevel < 0 {
           actualLevel = 0
       }
       if actualLevel > 29 {
           actualLevel = 29
       }

       err := a.ProgramWeight(row, col, float64(actualLevel)/29.0)
       return actualLevel, err
   }
   ```

3. Add toggle and σ slider to GUI

**Acceptance Criteria:**
- [ ] Repeated writes produce distribution of levels
- [ ] σ slider adjusts spread (0-3 levels)
- [ ] Can be disabled for deterministic operation

---

### Task 2.2: Endurance Model Integration

**Files:** `module2-crossbar/pkg/crossbar/array.go`, `drift.go`

**Problem:**
`SwitchingCount` is tracked but not used. No fatigue modeling.

**Changes:**

1. Add endurance parameters:
   ```go
   type EnduranceConfig struct {
       FatigueThreshold int64   // Cycles before degradation starts
       FailureThreshold int64   // Cycles at 50% degradation
       Enabled          bool
   }
   ```

2. Implement `GetFatigueAdjustedConductance()`:
   ```go
   func (a *Array) GetFatigueAdjustedConductance(row, col int) float64 {
       cell := &a.cells[row][col]
       G := cell.Conductance

       if !a.config.Endurance.Enabled {
           return G
       }

       cycles := cell.SwitchingCount
       if cycles < a.config.Endurance.FatigueThreshold {
           return G // No degradation yet
       }

       // Exponential degradation
       fatigueRatio := float64(cycles-a.config.Endurance.FatigueThreshold) /
                       float64(a.config.Endurance.FailureThreshold-a.config.Endurance.FatigueThreshold)
       degradation := 1.0 - 0.5*(1-math.Exp(-3*fatigueRatio))

       // Degradation narrows the conductance window
       gMid := 0.5
       return gMid + (G-gMid)*degradation
   }
   ```

3. Apply fatigue during MVM reads

4. Add endurance visualization (cycle count heatmap, warning indicators)

5. Add "Age 10⁶ cycles" button for demo

**Acceptance Criteria:**
- [ ] Cycle count increments on write
- [ ] Conductance window narrows with fatigue
- [ ] Visual warning when approaching fatigue
- [ ] Reset button clears cycle counts

---

### Task 2.3: Validated Drift Model

**Files:** `module2-crossbar/pkg/crossbar/drift.go`

**Problem:**
Line 69-70 notes: "FeFET drift coefficient 0.001 is an assumed value for simulation. No reported in literature source."

**Changes:**

1. Research and document actual FeFET drift from literature:
   - Find reported in literature drift measurements for HZO FeFET
   - Document source and conditions
   - Update coefficient or mark as "estimated"

2. Add drift model selection:
   ```go
   type DriftModel int
   const (
       DriftModelAssumed DriftModel = iota // Current 0.001
       DriftModelLiterature                 // From reported in literature source
       DriftModelMeasured                   // From user calibration
   )
   ```

3. Add documentation of assumptions in code comments

4. Add "Assumed Value" warning in GUI when using unvalidated parameters

**Acceptance Criteria:**
- [ ] Drift coefficient documented with source (or marked assumed)
- [ ] GUI shows warning for assumed values
- [ ] Allow user to input custom drift coefficient

---

### Task 2.4: Process Variation Model

**Files:** `module2-crossbar/pkg/crossbar/array.go`

**Problem:**
Current `NoiseFactor` is device-to-device variation. Missing systematic process variation (gradients, edge effects).

**Changes:**

1. Add systematic variation:
   ```go
   type ProcessVariation struct {
       // Device-to-device (random)
       DeviceSigma float64

       // Systematic (spatial correlation)
       GradientX   float64 // Horizontal gradient (%/cell)
       GradientY   float64 // Vertical gradient (%/cell)
       EdgeEffect  float64 // Edge cell degradation (%)
   }
   ```

2. Implement `GetVariationFactor(row, col int) float64`:
   ```go
   func (a *Array) GetVariationFactor(row, col int) float64 {
       pv := a.config.ProcessVariation

       // Random component (existing)
       random := a.cells[row][col].NoiseFactor

       // Systematic gradient
       centerRow := float64(a.config.Rows) / 2
       centerCol := float64(a.config.Cols) / 2
       gradX := 1.0 + pv.GradientX*(float64(col)-centerCol)
       gradY := 1.0 + pv.GradientY*(float64(row)-centerRow)

       // Edge effect
       edgeFactor := 1.0
       if row == 0 || row == a.config.Rows-1 ||
          col == 0 || col == a.config.Cols-1 {
           edgeFactor = 1.0 - pv.EdgeEffect
       }

       return random * gradX * gradY * edgeFactor
   }
   ```

3. Apply systematic variation in MVM

4. Visualize process variation as heatmap overlay

**Acceptance Criteria:**
- [ ] Systematic gradient visible in heatmap
- [ ] Edge cells show degradation
- [ ] Can disable for ideal comparison

---

## Phase 3: GUI and Visualization Improvements

**Duration:** 1-2 weeks
**Priority:** LOW
**Goal:** Better visualization of physics effects

### Task 3.1: Before/After Comparison View

**Files:** `module2-crossbar/pkg/gui/tabs/*.go`

**Changes:**
- Side-by-side ideal vs actual output display
- Difference heatmap showing error distribution
- Toggle between views

---

### Task 3.2: Accuracy Waterfall Chart

**Files:** `module2-crossbar/pkg/gui/widgets_analysis.go`

**Changes:**
- Visualize `ComputeAccuracyDegradation()` results
- Show stepwise accuracy loss from each non-ideality
- Interactive: click on bar to see details

---

### Task 3.3: Sneak Path Animation

**Files:** `module2-crossbar/pkg/gui/animation.go`

**Changes:**
- Animate sneak current flow during read
- Show parasitic paths as faded lines
- Highlight worst-case sneak contributors

---

### Task 3.4: Parameter Sensitivity Dashboard

**Files:** `module2-crossbar/pkg/gui/` (new file)

**Changes:**
- Sliders for all physics parameters
- Real-time RMSE/accuracy update
- Sensitivity analysis: which parameter matters most?

---

## Testing Strategy

### Unit Tests

**Location:** `module2-crossbar/pkg/crossbar/*_test.go`

```go
// Phase 1 tests
func TestExponentialConductance(t *testing.T)
func TestFullMVMSneakMagnitude(t *testing.T)
func TestHalfSelectDisturb(t *testing.T)
func TestTemperatureScaling(t *testing.T)

// Phase 2 tests
func TestSwitchingStatisticsDistribution(t *testing.T)
func TestEnduranceFatigue(t *testing.T)
func TestProcessVariationGradient(t *testing.T)
```

### Integration Tests

```go
func TestMVMAccuracyWithAllNonIdealities(t *testing.T)
func TestPassiveVs1T1RAccuracyGap(t *testing.T)
func TestLargeArrayScaling(t *testing.T)
```

### Benchmark Tests

```go
func BenchmarkMVMWithSneakPaths(b *testing.B)
func BenchmarkFullSneakCalculation(b *testing.B)
```

---

## Configuration Changes

### physics.yaml Additions

```yaml
crossbar:
  # Existing
  wl_resistance_per_cell: 2.5
  bl_resistance_per_cell: 2.5
  contact_resistance: 50

  # NEW: Conductance model
  conductance_model: "exponential"  # linear, exponential, lookup
  conductance_table: []             # For lookup model

  # NEW: Process variation
  process_variation:
    device_sigma: 0.02
    gradient_x: 0.001
    gradient_y: 0.001
    edge_effect: 0.05

  # NEW: Write statistics
  write_statistics:
    enabled: true
    vth_sigma: 0.05  # 5% level variation

  # NEW: Endurance
  endurance:
    enabled: true
    fatigue_threshold: 100000000   # 10^8
    failure_threshold: 1000000000000  # 10^12

  # NEW: Half-select disturb
  half_select:
    enabled: true
    disturb_threshold: 0.3  # Fraction of Vc
    disturb_rate: 0.001     # Per pulse
```

---

## Accuracy Targets

| Metric | Current | Target |
|--------|---------|--------|
| MVM RMSE (ideal) | 0% | 0% |
| MVM RMSE (0T1R + all effects) | ~5-10% | <10% (realistic) |
| MVM RMSE (1T1R + all effects) | ~2-5% | <5% (realistic) |
| Sneak path magnitude (0T1R) | Simplified | 5-20% (literature-matched) |
| Sneak path magnitude (1T1R) | ~0.001% | ~0.001% |

---

## Dependencies

### Internal
- `config/physics` - Configuration loading
- `module1-hysteresis` - Material properties (optional integration)
- `shared/` - Theme, logging

### External
- None new (standard Go libraries)

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Full sneak calculation slow | High | Simplified fallback for large arrays |
| Drift coefficient unvalidated | Medium | Document assumptions, allow override |
| Breaking existing behavior | High | Comprehensive unit tests |
| GUI complexity | Medium | Hide advanced features behind toggle |

---

## Timeline

```
Week 1:    Phase 1 Tasks 1.1, 1.2 (Conductance, Full Sneak)
Week 2:    Phase 1 Tasks 1.3, 1.4 (Disturb, Temperature)
Week 3-4:  Phase 2 Tasks 2.1, 2.2, 2.3 (Statistics, Endurance, Drift)
Week 5:    Phase 2 Task 2.4 + Phase 3 GUI (Process Variation, Visualizations)
Week 6:    Testing, documentation, bug fixes
```

---

## File Summary

| File | Changes |
|------|---------|
| `array.go` | Conductance model, disturb tracking, statistics |
| `nonidealities.go` | Temperature scaling |
| `enhanced.go` | Full sneak calculation, disturb integration |
| `drift.go` | Validated coefficients, documentation |
| `temperature.go` | NEW: Temperature effects |
| `process_variation.go` | NEW: Systematic variation |
| `gui/tabs/*.go` | Enhanced visualizations |
| `physics.yaml` | New configuration parameters |

---

## Related Documentation

- `docs/crossbar/../educational/crossbar.research.md` - Crossbar array research
- `docs/peripheral-circuits/circuits.CIM-fundamentals.md` - CIM physics
- `docs/development/TESTING.md` - Test framework
- `CLAUDE.md` - Project conventions
