# Module 4 Improvement Plan

**Peripheral Circuit Physics Enhancement Roadmap**

---

## Overview

| Field | Value |
|-------|-------|
| **Module** | module4-circuits |
| **Goal** | Improve physics accuracy from ~20% error to <5% error |
| **Scope** | Peripheral models (ADC, DAC, TIA, ChargePump) + Device simulation |
| **Estimated Effort** | 6-8 weeks total |

---

## Current State Assessment

### What Works
- Basic peripheral models with proper SI units
- INL/DNL nonlinearity for ADC/DAC
- Material-derived voltage ranges (Vc = Ec × thickness)
- Mode-First UX (READ/WRITE/COMPUTE)
- Architecture toggle (0T1R/1T1R/2T1R)

### What's Missing
- Realistic conductance curves (currently linear)
- Sneak path currents in passive mode
- IR drop in large arrays
- Write disturb modeling
- Temperature effects
- Switching statistics
- Endurance/fatigue tracking

---

## Phase 1: Critical Accuracy Improvements

**Duration:** 2-3 weeks
**Priority:** HIGH
**Goal:** Fix fundamental physics errors that cause >10% simulation inaccuracy

### Task 1.1: Nonlinear Conductance Model

**Files:** `module4-circuits/pkg/gui/device_state.go`

**Problem:**
Current: `G = Gmin + (Gmax-Gmin) × level/29` (linear)
Real: Exponential or saturating G(V) relationship

**Changes:**

1. Add `ConductanceModel` enum type:
   ```go
   type ConductanceModel int
   const (
       ConductanceLinear ConductanceModel = iota
       ConductanceExponential
       ConductancePreisach
   )
   ```

2. Add `conductanceModel` field to `DeviceState`

3. Implement `GetConductance(level, maxLevel int) float64` method:
   - Linear: current behavior
   - Exponential: `Gmin × exp(ln(Gmax/Gmin) × level/maxLevel)`
   - Preisach: call `material.DiscreteLevel()` (already exists)

4. Update `Compute()` to use `GetConductance()` instead of inline calculation

**Acceptance Criteria:**
- [ ] Conductance at level 0 equals Gmin
- [ ] Conductance at level 29 equals Gmax
- [ ] Exponential model: midpoint = geometric mean of Gmin/Gmax
- [ ] GUI selector to switch between models
- [ ] Unit tests pass

---

### Task 1.2: Sneak Path Model for Passive Mode

**Files:** `module4-circuits/pkg/gui/device_state.go`

**Problem:**
Current: Each column current calculated independently
Real: 0T1R arrays have parallel sneak paths adding 5-20% error

**Changes:**

1. Add sneak path tracking fields to `DeviceState`:
   ```go
   sneakFraction []float64  // Per-row sneak current fraction
   sneakEnabled  bool       // Toggle for demonstration
   ```

2. Add `calculateAverageG()` helper method

3. Create `ComputeWithSneakPaths()` method:
   - Calculate direct current (existing logic)
   - Estimate sneak current: `V × Gavg × (rows-1) / 3` per column
   - Apply architecture-dependent attenuation:
     - 0T1R: 1.0 (full sneak)
     - Self-rectifying: 0.01-0.1
     - 1T1R/2T1R: 0.0 (no sneak)
   - Sum direct + sneak currents
   - Store sneak fraction for display

4. Update `Compute()` to call `ComputeWithSneakPaths()` when passive

5. Add sneak percentage display to GUI array visualization

**Acceptance Criteria:**
- [ ] Passive mode shows 5-20% sneak current (configurable)
- [ ] 1T1R mode shows 0% sneak current
- [ ] Sneak percentage displayed per row
- [ ] Toggle to enable/disable for comparison
- [ ] Unit tests verify sneak magnitude in expected range

---

### Task 1.3: IR Drop Model

**Files:** `module4-circuits/pkg/peripherals/analysis.go`, `device_state.go`

**Problem:**
Current: All cells see exact DAC voltage
Real: Word/bit line resistance causes 1-5% voltage drop at array edges

**Changes:**

1. Add to `analysis.go`:
   ```go
   type IRDropAnalysis struct {
       WLResistancePerCell float64
       BLResistancePerCell float64
       EffectiveVoltages   [][]float64
       DropAtFarCorner     float64
   }

   func CalculateIRDrop(rows, cols int, wlR, blR float64,
                        dacVoltages []float64) *IRDropAnalysis
   ```

2. Add IR drop parameters to `physics.yaml`:
   ```yaml
   crossbar:
     wl_resistance_per_cell: 0.5  # Ω
     bl_resistance_per_cell: 0.5  # Ω
   ```

3. Update `DeviceState.Compute()` to optionally apply IR drop:
   - Load resistance values from config
   - Calculate effective voltage at each cell position
   - Use effective voltage instead of DAC voltage

4. Add IR drop toggle and visualization to GUI

**Acceptance Criteria:**
- [ ] 64×64 array shows 1-5% voltage drop at far corner
- [ ] 8×8 array shows <1% drop (minimal)
- [ ] IR drop scales with array size
- [ ] Heatmap visualization of effective voltages
- [ ] Toggle to enable/disable

---

### Task 1.4: Write Disturb Model

**Files:** `module4-circuits/pkg/gui/device_state.go`

**Problem:**
Current: Write affects only target cell
Real: Half-selected cells (same row or column) see V/2, causing gradual drift

**Changes:**

1. Add cell state tracking:
   ```go
   type CellState struct {
       Level           int
       HalfSelectCount int
       DisturbShift    float64
   }

   cellStates [][]CellState
   ```

2. Implement `ApplyWriteDisturb()` method:
   - Called after each write operation
   - Identify half-selected cells (same row OR same column as target)
   - Calculate disturb factor based on V/2 vs Vc ratio
   - Accumulate small conductance shift (0.1% per pulse typical)
   - When accumulated shift exceeds threshold, bump level

3. Add disturb visualization to GUI:
   - Highlight half-selected cells during write
   - Show cumulative disturb percentage per cell
   - Warning when disturb exceeds 5%

4. Only active in passive mode (1T1R isolates non-target cells)

**Acceptance Criteria:**
- [ ] Half-selected cells highlighted during write animation
- [ ] Disturb accumulates over repeated writes
- [ ] Level shifts after sufficient disturb accumulation
- [ ] 1T1R mode shows no disturb
- [ ] Reset button clears disturb history

---

## Phase 2: Enhanced Realism

**Duration:** 2-3 weeks
**Priority:** MEDIUM
**Goal:** Add environmental and statistical effects for educational value

### Task 2.1: Temperature Model

**Files:** `module4-circuits/pkg/peripherals/temperature.go` (new), `device_state.go`

**Changes:**

1. Create `temperature.go`:
   ```go
   type TemperatureModel struct {
       AmbientK   float64
       ReferenceK float64
   }

   func (t *TemperatureModel) AdjustedEc(Ec0 float64) float64
   func (t *TemperatureModel) AdjustedPr(Pr0 float64) float64
   func (t *TemperatureModel) AdjustedNoise(noise0 float64) float64
   ```

2. Temperature scaling rules:
   - Ec: -10% per +100K (decreases with heat)
   - Pr: -20% per +100K (more sensitive)
   - Pr at cryogenic (<100K): +200-300% (major enhancement)
   - TIA noise: scales with √T

3. Add temperature slider to GUI (77K - 400K)

4. Presets: Cryogenic (77K), Room (300K), Automotive (400K)

5. Update voltage ranges when temperature changes

**Acceptance Criteria:**
- [ ] Temperature slider functional
- [ ] Voltage ranges update with temperature
- [ ] Cryogenic shows enhanced Pr (visible in GUI)
- [ ] Automotive shows degraded performance
- [ ] Preset buttons work

---

### Task 2.2: Switching Statistics

**Files:** `module4-circuits/pkg/gui/device_state.go`

**Changes:**

1. Add switching statistics:
   ```go
   type SwitchingStats struct {
       VthSigma float64    // Standard deviation in V
       Enabled  bool
       RNG      *rand.Rand
   }
   ```

2. Implement `ProgramWithVariation()`:
   - Calculate ideal level from voltage
   - Add Gaussian noise (σ = 30-50 mV typical)
   - Clamp to valid range [0, 29]
   - Return actual programmed level

3. Update write operation to use `ProgramWithVariation()` when enabled

4. Add toggle and σ slider to GUI

**Acceptance Criteria:**
- [ ] Repeated writes to same voltage produce slightly different levels
- [ ] Distribution is approximately Gaussian
- [ ] σ slider adjusts spread (0-100 mV range)
- [ ] Toggle to enable/disable
- [ ] Histogram visualization of write results (optional)

---

### Task 2.3: Endurance Tracking

**Files:** `module4-circuits/pkg/gui/device_state.go`, `tab_unified.go`

**Changes:**

1. Add endurance tracker:
   ```go
   type EnduranceTracker struct {
       CyclesPerCell [][]int64
       TotalCycles   int64
       FatigueCycles int64  // When degradation starts
   }
   ```

2. Implement `GetEffectivePr()`:
   - Below fatigue threshold: return nominal Pr
   - Above fatigue: exponential decay toward 50%

3. Increment cycle count on each write

4. GUI additions:
   - Total cycle counter display
   - Per-cell cycle heatmap (optional view)
   - "Age 10⁶ cycles" button for demo
   - Warning indicator when approaching fatigue

**Acceptance Criteria:**
- [ ] Cycle counter increments on write
- [ ] Fatigue visible after many cycles (or accelerated aging)
- [ ] Conductance range reduces with fatigue
- [ ] Reset button clears cycle counts
- [ ] Configurable fatigue threshold

---

### Task 2.4: TIA Frequency Response

**Files:** `module4-circuits/pkg/peripherals/tia.go`

**Changes:**

1. Add frequency response to TIA:
   ```go
   func (t *TIA) GainAtFrequency(freq float64) float64 {
       // Single-pole: |H(f)| = Gain / sqrt(1 + (f/f_-3dB)²)
       return t.Gain / math.Sqrt(1 + math.Pow(freq/t.Bandwidth, 2))
   }
   ```

2. Add `ConvertWithBandwidth(current, signalBW float64)` method

3. Optional: Bode plot visualization in reference tab

**Acceptance Criteria:**
- [ ] Gain at DC equals nominal gain
- [ ] Gain at bandwidth equals nominal/√2 (-3dB)
- [ ] High-frequency signals show reduced gain
- [ ] Unit tests verify frequency response

---

## Phase 3: Polish and Edge Cases

**Duration:** 1-2 weeks
**Priority:** LOW
**Goal:** Complete the physics model with remaining effects

### Task 3.1: Charge Pump Transient Response

**Files:** `module4-circuits/pkg/peripherals/chargepump.go`

**Changes:**
- Add `TransientResponse(loadStep, duration float64) []float64`
- Model voltage droop and recovery after load step
- Optional visualization in timing diagrams

---

### Task 3.2: ADC Comparator Kickback

**Files:** `module4-circuits/pkg/peripherals/adc.go`

**Changes:**
- Add `KickbackCharge` parameter
- Implement `ConvertWithKickback()` method
- Small additional noise term during conversion

---

### Task 3.3: SET/RESET Timing Asymmetry

**Files:** `module4-circuits/pkg/peripherals/analysis.go`

**Changes:**
- Separate SET and RESET pulse widths in timing analysis
- RESET typically 20% longer than SET
- Update timing diagrams

---

### Task 3.4: Retention Model

**Files:** `module4-circuits/pkg/gui/device_state.go`

**Changes:**
- Track last write time per cell
- Implement exponential decay toward middle state
- "Simulate 1 year" button for demonstration

---

## Testing Strategy

### Unit Tests

**Location:** `module4-circuits/pkg/peripherals/peripherals_test.go`

```go
// Phase 1 tests
func TestConductanceModels(t *testing.T)
func TestSneakPathMagnitude(t *testing.T)
func TestIRDropScaling(t *testing.T)
func TestWriteDisturbAccumulation(t *testing.T)

// Phase 2 tests
func TestTemperatureScaling(t *testing.T)
func TestSwitchingStatisticsDistribution(t *testing.T)
func TestEnduranceFatigue(t *testing.T)
func TestTIAFrequencyResponse(t *testing.T)
```

### Integration Tests

```go
func TestTransferFunctionAccuracy(t *testing.T)
func TestPassiveModeVs1T1RAccuracy(t *testing.T)
func TestLargeArrayIRDrop(t *testing.T)
```

### Manual Testing

- [ ] Visual inspection of sneak path highlighting
- [ ] Temperature slider responsiveness
- [ ] Endurance counter incrementing
- [ ] Write disturb visualization
- [ ] Comparison: passive vs 1T1R accuracy

---

## Configuration Changes

### physics.yaml Additions

```yaml
# Add to existing physics.yaml
crossbar:
  wl_resistance_per_cell: 0.5      # Ω per cell pitch
  bl_resistance_per_cell: 0.5      # Ω per cell pitch
  contact_resistance: 100          # Ω (transistor ON resistance)

switching:
  vth_sigma: 0.040                 # 40 mV cycle-to-cycle variation

endurance:
  wakeup_cycles: 10000             # Initial wake-up phase
  fatigue_cycles: 100000000        # 10⁸ cycles to fatigue onset
  failure_cycles: 1000000000000    # 10¹² cycles to 50% Pr loss

temperature:
  reference_k: 300                 # Calibration temperature
  ec_temp_coeff: -0.001            # -0.1% per K
  pr_temp_coeff: -0.002            # -0.2% per K

sneak_path:
  rectification_ratio: 1.0         # 1.0 for 0T1R, 100+ for self-rectifying
```

---

## GUI Changes Summary

### New Controls

| Control | Location | Purpose |
|---------|----------|---------|
| Conductance model selector | Settings panel | Linear/Exponential/Preisach |
| Sneak path toggle | Architecture section | Enable/disable for comparison |
| IR drop toggle | Architecture section | Show voltage drop effects |
| Temperature slider | New section | 77K - 400K range |
| Temperature presets | Temperature section | Cryo/Room/Auto buttons |
| Variation toggle | Write section | Enable switching statistics |
| σ slider | Write section | Adjust variation magnitude |
| Cycle counter | Status bar | Total write cycles |
| Age button | Tools section | Accelerate endurance |

### New Visualizations

| Visualization | Location | Purpose |
|---------------|----------|---------|
| Sneak % per row | Array right side | Show sneak current magnitude |
| IR drop heatmap | Array overlay | Effective voltage per cell |
| Disturb highlight | Array overlay | Half-selected cells |
| Disturb % per cell | Tooltip | Cumulative disturb |
| Fatigue indicator | Cell color | Aged cells darker |

---

## Dependencies

### Internal Dependencies

- `module1-hysteresis/pkg/ferroelectric` - Material properties, Preisach model
- `config/physics` - Configuration loading
- `shared/` - Theme, logging

### No External Dependencies Added

All improvements use standard Go libraries (math, rand).

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Performance regression from sneak calculation | Medium | Make sneak calculation optional, optimize for small arrays |
| Complexity overwhelming users | Medium | Hide advanced features behind "Advanced" toggle |
| Breaking existing behavior | High | Comprehensive unit tests before changes |
| Physics accuracy debates | Low | Document sources, make models configurable |

---

## Success Metrics

| Metric | Current | Target | How to Measure |
|--------|---------|--------|----------------|
| Transfer function error | ~20% | <5% | Compare ADC output vs input level |
| Passive vs 1T1R accuracy gap | Not shown | 10-15% | Compare MVM results |
| Sneak path magnitude | 0% | 5-20% | Measure sneak fraction |
| IR drop at 64×64 corner | 0% | 1-5% | Measure effective voltage |

---

## Timeline

```
Week 1-2:  Phase 1 Tasks 1.1, 1.2 (Conductance, Sneak Paths)
Week 3:    Phase 1 Tasks 1.3, 1.4 (IR Drop, Write Disturb)
Week 4-5:  Phase 2 Tasks 2.1, 2.2 (Temperature, Statistics)
Week 6:    Phase 2 Tasks 2.3, 2.4 (Endurance, TIA)
Week 7:    Phase 3 (Polish tasks)
Week 8:    Testing, documentation, bug fixes
```

---

## Approval

- [ ] Technical review completed
- [ ] Resource allocation confirmed
- [ ] Timeline agreed

---

## References

- `docs/peripheral-circuits/MODULE4-PHYSICS-IMPROVEMENTS.md` - Detailed physics analysis
- `docs/peripheral-circuits/circuits.CIM-fundamentals.md` - CIM background
- `docs/peripheral-circuits/circuits.research.md` - Literature sources
- `docs/development/TESTING.md` - Test framework
