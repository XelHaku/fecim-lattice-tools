# Module 4 Physics Improvements Proposal

**Comprehensive Review and Enhancement Plan for FeCIM Peripheral Circuit Simulation**

*Last Updated: January 2026*

---

## Executive Summary

This document analyzes the current physics implementation in Module 4 (peripheral circuits) and proposes improvements based on peer-reviewed CIM research (2023-2026). The analysis identifies **12 critical gaps** and proposes **prioritized enhancements** to improve simulation accuracy.

**Current State:** Good foundational models with proper SI units and basic nonlinearity. Missing key real-world effects that impact accuracy by 10-30%.

**Goal:** Achieve <5% error between simulation and real FeFET CIM hardware behavior.

---

## 1. Current Implementation Analysis

### 1.1 What's Working Well

| Component | Implementation | Quality | Notes |
|-----------|----------------|---------|-------|
| **ADC Model** | SAR/Flash/Σ-Δ types, INL/DNL | Good | Proper ENOB calculation |
| **DAC Model** | 5-bit, ±1.5V range, INL/DNL | Good | Monotonicity check included |
| **TIA Model** | 10kΩ gain, 100MHz BW, noise | Good | SNR and settling time |
| **Charge Pump** | 2-stage Dickson, 70% efficiency | Good | Rise time, ripple modeled |
| **Material Integration** | Vc from Ec×thickness | Excellent | Dynamic voltage ranges |
| **Mode-First UX** | READ/WRITE/COMPUTE modes | Excellent | Proper constraint enforcement |

### 1.2 Critical Physics Gaps

| Gap | Current Model | Real Hardware | Impact on Accuracy |
|-----|---------------|---------------|-------------------|
| **1. Linear Conductance** | G = Gmin + k×level | Exponential/saturating G(V) | 10-20% error at extremes |
| **2. No Sneak Paths** | Independent columns | 5-20% leakage current (0T1R) | Underestimates read error |
| **3. No IR Drop** | Zero line resistance | 1-5% drop in 64×64 arrays | Large arrays inaccurate |
| **4. No Write Disturb** | Perfect isolation | V/2 on half-selected cells | Passive mode error ignored |
| **5. Temperature Effects** | Room temp only (300K) | -40°C to +125°C variation | ±10-15% Ec/Pr drift |
| **6. No Switching Statistics** | Deterministic Vth | Gaussian cycle-to-cycle noise | Noise floor understated |
| **7. No Endurance Model** | Infinite cycles | Fatigue after 10⁸-10¹² cycles | Reliability not shown |
| **8. TIA Frequency Response** | Ideal integrator | Single-pole roll-off | High-speed read errors |
| **9. Charge Pump Dynamics** | Steady-state only | Transient load regulation | Write pulse distortion |
| **10. ADC Kickback Noise** | No comparator kickback | 0.5-2 LSB kickback during SAR | Conversion noise |
| **11. SET/RESET Asymmetry** | Symmetric operations | 20% slower erase | Programming time wrong |
| **12. Retention Loss** | Perfect storage | 1-10% drift over months | Long-term accuracy |

---

## 2. Proposed Improvements (Priority Order)

### 2.1 HIGH PRIORITY - Critical for Accuracy

#### Improvement #1: Nonlinear Conductance Model

**Problem:** Current linear model: `G = Gmin + (Gmax-Gmin) × level/29`

Real FeFETs have exponential/logarithmic conductance:

```
Real G(V) curve:
        │
  Gmax ─┼─────────────────────●●●
        │                ●●●
  G     │            ●●●
        │        ●●●
        │     ●●
  Gmin ─┼─●●●
        └────────────────────────→ Level
         0                      29

Linear (current) vs. Exponential (real)
```

**Proposed Implementation:**

```go
// Enhanced conductance model in device_state.go
type ConductanceModel int

const (
    ConductanceLinear      ConductanceModel = iota // Current: G = Gmin + k*level
    ConductanceExponential                         // G = Gmin * exp(k*level)
    ConductancePreisach                            // From hysteresis module
    ConductanceMeasured                            // Lookup table from calibration
)

// GetConductance returns physics-accurate conductance for a given level
func (ds *DeviceState) GetConductance(level, maxLevel int) float64 {
    switch ds.conductanceModel {
    case ConductanceExponential:
        // Exponential model: G = Gmin * exp(ln(Gmax/Gmin) * level/maxLevel)
        ratio := ds.Gmax / ds.Gmin
        return ds.Gmin * math.Exp(math.Log(ratio) * float64(level) / float64(maxLevel))

    case ConductancePreisach:
        // Use Module 1 hysteresis for accurate P-E curve
        if ds.material != nil {
            return ds.material.DiscreteLevel(level, maxLevel)
        }
        fallthrough

    case ConductanceMeasured:
        // Lookup from calibration data
        if level < len(ds.calibratedG) {
            return ds.calibratedG[level]
        }
        fallthrough

    default: // Linear fallback
        return ds.Gmin + (ds.Gmax - ds.Gmin) * float64(level) / float64(maxLevel)
    }
}
```

**Expected Improvement:** +10-15% accuracy at high/low conductance states.

---

#### Improvement #2: Sneak Path Current Model (Passive Mode)

**Problem:** Current model assumes independent column currents. In 0T1R passive arrays, sneak paths add 5-20% error.

**Physics:**
```
Three-cell sneak path model:

    BL_j (read column)     BL_k (other column)
         │                      │
WL_i ────●──────────────────────●────→ Target cell
         │      ↗               │
         │    Sneak path        │
         │   through            │
WL_m ────●──────────────────────●────
         │   unselected         │
         │   cells              │

Sneak conductance:
G_sneak = Σ[1 / (1/G_a + 1/G_b + 1/G_c)]  for each parallel path

Total sneak from N×N array:
G_sneak_total ≈ G_avg × (N-1) / 3  (empirical)
```

**Proposed Implementation:**

```go
// In device_state.go - Compute method
func (ds *DeviceState) ComputeWithSneakPaths(weights [][]int, quantLevels int) {
    if !ds.isPassive {
        ds.Compute(weights, quantLevels) // 1T1R: no sneak paths
        return
    }

    // Calculate average array conductance
    Gavg := ds.calculateAverageG(weights, quantLevels)

    for r := 0; r < ds.rows; r++ {
        // Direct current (as before)
        directCurrent := ds.computeDirectCurrent(r, weights, quantLevels)

        // Sneak current estimate
        // Sneak ≈ V × Gavg × (rows-1) / 3 per active column
        activeColumns := ds.countActiveColumns()
        sneakFactor := float64(ds.rows-1) / 3.0
        sneakCurrent := 0.0

        for c := 0; c < ds.cols; c++ {
            if ds.dacVoltages[c] > 0.01 {
                sneakCurrent += ds.dacVoltages[c] * Gavg * sneakFactor
            }
        }

        // Attenuate sneak based on array size and material
        sneakAttenuation := ds.getSneakAttenuation() // 1.0 for pure 0T1R, 0.1 for self-rectifying
        totalCurrent := directCurrent + sneakCurrent * sneakAttenuation

        ds.rowCurrents[r] = totalCurrent
        ds.sneakFraction[r] = sneakCurrent * sneakAttenuation / totalCurrent
        // ... rest of TIA/ADC conversion
    }
}

// Sneak attenuation based on architecture
func (ds *DeviceState) getSneakAttenuation() float64 {
    switch ds.architecture {
    case Arch0T1R:
        return 1.0 // Full sneak paths
    case Arch0T1R_SelfRectifying:
        return 0.01 // 100:1 rectification ratio
    case Arch1T1R, Arch2T1R:
        return 0.0 // No sneak paths
    default:
        return 1.0
    }
}
```

**Expected Improvement:** +15-20% accuracy in passive mode MVM.

---

#### Improvement #3: IR Drop Model for Large Arrays

**Problem:** Current model ignores word line and bit line resistance. Real arrays have IR drop:

```
Voltage profile across 64-column array:

V_actual │
         │ ●
  V_nom ─┼──●───────────────────────────────────────────
         │    ●●●
         │       ●●●●
         │          ●●●●●
         │              ●●●●●●
   V_far─┼─────────────────────●●●●●●●●●●●●●●●●●●●●●●●●●
         └────────────────────────────────────────────→
         0          Column Index                     63

IR drop = I × R_line × column_index
Typical: 1-5% drop at far end of 64×64 array
```

**Proposed Implementation:**

```go
// In analysis.go - Add IR drop calculation
type IRDropAnalysis struct {
    WLResistancePerCell float64 // Ω per cell pitch
    BLResistancePerCell float64 // Ω per cell pitch
    DropAtFarCorner     float64 // V
    EffectiveVoltages   [][]float64 // Actual V at each cell
}

func CalculateIRDrop(rows, cols int, wlR, blR float64,
                     dacVoltages []float64, weights [][]int) *IRDropAnalysis {
    ir := &IRDropAnalysis{
        WLResistancePerCell: wlR,
        BLResistancePerCell: blR,
        EffectiveVoltages:   make([][]float64, rows),
    }

    for r := 0; r < rows; r++ {
        ir.EffectiveVoltages[r] = make([]float64, cols)

        // Accumulated WL current up to this row
        cumulativeWLCurrent := 0.0

        for c := 0; c < cols; c++ {
            // BL IR drop (from driver at top to this row)
            blDrop := cumulativeWLCurrent * blR * float64(r)

            // WL IR drop (from driver at left to this column)
            wlDrop := dacVoltages[c] * wlR * float64(c) / 1000.0 // Simplified

            ir.EffectiveVoltages[r][c] = dacVoltages[c] - blDrop - wlDrop

            if ir.EffectiveVoltages[r][c] < 0 {
                ir.EffectiveVoltages[r][c] = 0
            }
        }
    }

    // Record worst-case drop
    ir.DropAtFarCorner = dacVoltages[0] - ir.EffectiveVoltages[rows-1][cols-1]

    return ir
}
```

**Configuration in physics.yaml:**
```yaml
crossbar:
  wl_resistance_per_cell: 0.5  # Ω (22nm metal)
  bl_resistance_per_cell: 0.5  # Ω
  contact_resistance: 100      # Ω (access transistor ON)
```

**Expected Improvement:** +3-5% accuracy for 64×64+ arrays.

---

#### Improvement #4: Write Disturb Model

**Problem:** In passive mode write, half-selected cells receive V/2, causing small conductance shifts.

**Physics:**
```
V/2 Scheme for passive write:

Target: Cell (1,1) at +1.5V

       BL0=0.75V  BL1=0V   BL2=0.75V
           │        │         │
WL0=0.75V ─●────────●─────────●─── Cells see ±0.75V (half-select)
           │        │         │
WL1=1.5V ──●────────●─────────●─── Target sees 1.5V, others see 0.75V
           │   ✓    │         │
           │ TARGET │         │
WL2=0.75V ─●────────●─────────●─── Cells see ±0.75V (half-select)

Half-select disturb:
- V/2 < Vc, so no full switching
- BUT: Repeated V/2 pulses cause gradual drift
- Typical: 0.1-1% conductance change per 100 pulses at V/2
```

**Proposed Implementation:**

```go
// In device_state.go
type CellState struct {
    Level           int
    HalfSelectCount int     // Number of V/2 exposures
    DisturbShift    float64 // Accumulated conductance drift (%)
}

// Track disturb during write
func (ds *DeviceState) ApplyWriteDisturb(targetRow, targetCol int,
                                          voltage float64, weights [][]int) {
    if !ds.isPassive || voltage < ds.writeRange.Min {
        return // No disturb in 1T1R or below threshold
    }

    halfSelectV := voltage / 2.0
    disturbThreshold := ds.material.CoerciveVoltage() * 0.5

    if halfSelectV < disturbThreshold {
        return // V/2 too low to cause any disturb
    }

    // Calculate disturb probability based on how close V/2 is to Vc
    disturbFactor := (halfSelectV - disturbThreshold) /
                     (ds.material.CoerciveVoltage() - disturbThreshold)

    // Apply small shift to half-selected cells
    for r := 0; r < ds.rows; r++ {
        for c := 0; c < ds.cols; c++ {
            if r == targetRow && c == targetCol {
                continue // Target cell - full program
            }

            isHalfSelected := (r == targetRow || c == targetCol)
            if isHalfSelected {
                ds.cellStates[r][c].HalfSelectCount++

                // Cumulative disturb (0.1% per pulse typical)
                disturb := 0.001 * disturbFactor
                ds.cellStates[r][c].DisturbShift += disturb

                // Apply to conductance (shifts level slightly)
                if ds.cellStates[r][c].DisturbShift > 0.05 {
                    // 5% accumulated shift = bump level by 1
                    shift := int(ds.cellStates[r][c].DisturbShift / 0.05)
                    weights[r][c] = min(weights[r][c] + shift, 29)
                    ds.cellStates[r][c].DisturbShift -= float64(shift) * 0.05
                }
            }
        }
    }
}
```

**Expected Improvement:** Realistic passive mode write accuracy (-5-10% from ideal).

---

### 2.2 MEDIUM PRIORITY - Improved Realism

#### Improvement #5: Temperature-Dependent Parameters

**Problem:** Current model uses fixed 300K parameters. Real FeFETs show significant temperature variation:

| Parameter | 300K | 400K | 77K (Cryo) | Source |
|-----------|------|------|------------|--------|
| Ec | 1.0 MV/cm | 0.9 MV/cm | 1.1 MV/cm | Literature |
| Pr | 25 µC/cm² | 20 µC/cm² | 75 µC/cm² | Adv. Elec. Mat. 2024 |
| TIA Noise | 1 pA/√Hz | 1.2 pA/√Hz | 0.3 pA/√Hz | Theory |

**Proposed Implementation:**

```go
// In peripherals/temperature.go
type TemperatureModel struct {
    AmbientK      float64 // Operating temperature (K)
    ReferenceK    float64 // Calibration temperature (K)
}

// Temperature-adjusted coercive field
func (t *TemperatureModel) AdjustedEc(Ec0 float64) float64 {
    // Ec decreases ~10% per 100K increase (typical HZO)
    tempFactor := 1.0 - 0.001 * (t.AmbientK - t.ReferenceK)
    return Ec0 * tempFactor
}

// Temperature-adjusted Pr
func (t *TemperatureModel) AdjustedPr(Pr0 float64) float64 {
    // Pr decreases ~20% per 100K increase (more sensitive)
    tempFactor := 1.0 - 0.002 * (t.AmbientK - t.ReferenceK)
    if t.AmbientK < 100 {
        // Cryogenic: Pr increases 2-3× below 100K
        tempFactor = 1.0 + 2.0 * (100 - t.AmbientK) / 100
    }
    return Pr0 * tempFactor
}

// Temperature-adjusted TIA noise
func (t *TemperatureModel) AdjustedNoise(noise0 float64) float64 {
    // Thermal noise ∝ √T
    return noise0 * math.Sqrt(t.AmbientK / t.ReferenceK)
}
```

**GUI Addition:** Temperature slider (77K - 400K) with presets:
- Cryogenic (77K) - Shows enhanced Pr
- Room Temp (300K) - Default
- Automotive High (400K) - Shows degraded performance

---

#### Improvement #6: Switching Statistics (Cycle-to-Cycle Noise)

**Problem:** Current write is deterministic. Real FeFETs show Gaussian Vth variation:

```
Vth distribution for repeated SET operations:

Count │
      │       ████
      │      ██████
      │     ████████
      │    ██████████
      │   ████████████
      │  ██████████████
      └──────────────────→ Vth
       µ-2σ  µ  µ+2σ

Typical σ(Vth) = 30-50 mV for HZO FeFET
```

**Proposed Implementation:**

```go
// In device_state.go
type SwitchingStats struct {
    VthMean   float64 // Mean threshold voltage
    VthSigma  float64 // Standard deviation (V)
    RNG       *rand.Rand
}

// ProgramWithVariation applies realistic write with statistical variation
func (ds *DeviceState) ProgramWithVariation(row, col int, targetLevel int,
                                             voltage float64) int {
    // Base programming (ideal)
    idealLevel := ds.voltageTolevel(voltage)

    // Add cycle-to-cycle variation
    sigma := ds.switchingStats.VthSigma / ds.writeRange.StepSize // In levels
    noise := ds.switchingStats.RNG.NormFloat64() * sigma

    actualLevel := idealLevel + int(math.Round(noise))

    // Clamp to valid range
    if actualLevel < 0 {
        actualLevel = 0
    }
    if actualLevel > 29 {
        actualLevel = 29
    }

    return actualLevel
}
```

**Configuration:**
```yaml
switching:
  vth_sigma: 0.040  # 40 mV standard deviation
  cycle_correlation: 0.3  # Some correlation between cycles
```

---

#### Improvement #7: Endurance Tracking

**Problem:** Current model assumes infinite endurance. Real devices degrade:

```
Endurance curve:

Pr/Pr0 │
  1.0  ─┼●●●●●●●●●●●●●●●●●●●●●
       │                      ●●●
  0.9  ─┼                        ●●●
       │                           ●●
  0.8  ─┼                            ●●
       │                              ●
       └─────────────────────────────────→ Cycles
        10⁶    10⁸    10¹⁰   10¹²

Wake-up: Pr increases first 10⁴ cycles
Fatigue: Pr decreases after 10⁸-10¹⁰ cycles
```

**Proposed Implementation:**

```go
// In device_state.go
type EnduranceTracker struct {
    CyclesPerCell [][]int64 // Write cycles per cell
    TotalCycles   int64

    WakeupCycles  int64   // Cycles for wake-up phase
    FatigueCycles int64   // Cycles where fatigue begins
    FailureCycles int64   // Cycles at 50% Pr loss
}

// GetEffectivePr returns Pr adjusted for endurance
func (e *EnduranceTracker) GetEffectivePr(row, col int, Pr0 float64) float64 {
    cycles := e.CyclesPerCell[row][col]

    if cycles < e.WakeupCycles {
        // Wake-up: Pr increases to 1.1× then back to 1.0×
        wakeupFactor := 1.0 + 0.1 * math.Sin(math.Pi * float64(cycles) / float64(e.WakeupCycles))
        return Pr0 * wakeupFactor
    }

    if cycles < e.FatigueCycles {
        // Stable phase
        return Pr0
    }

    // Fatigue: exponential decay
    fatigueProgress := float64(cycles - e.FatigueCycles) / float64(e.FailureCycles - e.FatigueCycles)
    fatigueFactor := math.Exp(-2.0 * fatigueProgress) // 50% at failure

    return Pr0 * fatigueFactor
}

// IncrementCycles called after each write
func (e *EnduranceTracker) IncrementCycles(row, col int) {
    e.CyclesPerCell[row][col]++
    e.TotalCycles++
}
```

**GUI Addition:**
- Endurance counter display (total cycles, per-cell heatmap)
- "Accelerated aging" button to simulate 10⁶ cycles
- Warning when approaching fatigue threshold

---

#### Improvement #8: TIA Frequency Response

**Problem:** Current TIA assumes ideal gain at all frequencies. Real TIAs have bandwidth-limited response:

```
Frequency Response:

Gain │
     │●●●●●●●●●●●●●●●●●
(dB) │                 ●●●
     │                    ●●●
     │                       ●●●
-3dB ┼─────────────────────────●●●
     │                            ●●●
     └────────────────────────────────→ Frequency
                               f_-3dB
```

**Proposed Implementation:**

```go
// In tia.go - Enhanced TIA model
type TIAEnhanced struct {
    TIA

    // Frequency response
    DominantPole   float64 // Hz (typically = Bandwidth)
    SecondaryPole  float64 // Hz (optional, for 2-pole model)

    // Parasitic capacitance
    InputCapacitance float64 // F (affects BW with source impedance)
}

// FrequencyResponse returns gain at given frequency
func (t *TIAEnhanced) FrequencyResponse(freq float64) complex128 {
    // Single-pole transfer function: H(s) = Gain / (1 + s/ωp)
    s := complex(0, 2*math.Pi*freq)
    wp := complex(2*math.Pi*t.DominantPole, 0)

    H := complex(t.Gain, 0) / (1 + s/wp)
    return H
}

// ConvertWithBandwidth applies frequency-dependent gain
func (t *TIAEnhanced) ConvertWithBandwidth(current float64, signalBW float64) float64 {
    // Effective gain reduced at high signal frequencies
    gainReduction := 1.0 / math.Sqrt(1 + math.Pow(signalBW/t.DominantPole, 2))
    effectiveGain := t.Gain * gainReduction

    return current * effectiveGain + t.OutputOffset
}
```

---

### 2.3 LOW PRIORITY - Nice to Have

#### Improvement #9: Charge Pump Transient Response

**Problem:** Current model is steady-state. Real charge pumps have load-dependent transients:

```go
// In chargepump.go
func (c *ChargePump) TransientResponse(loadStep float64, duration float64) []float64 {
    // Model voltage droop and recovery after load step
    samples := int(duration / 1e-9) // 1ns resolution
    response := make([]float64, samples)

    initialDroop := loadStep / (c.FlyCapacitance * c.ClockFrequency)
    tau := c.OutputCapacitance / (c.ClockFrequency * c.FlyCapacitance)

    for i := 0; i < samples; i++ {
        t := float64(i) * 1e-9
        // Exponential recovery
        response[i] = c.OutputVoltage - initialDroop * math.Exp(-t/tau)
    }

    return response
}
```

---

#### Improvement #10: ADC Comparator Kickback

**Problem:** SAR ADC comparators inject charge back to DAC during comparison:

```go
// In adc.go
type ADCEnhanced struct {
    ADC
    KickbackCharge float64 // Coulombs (typical 1-10 fF × Vdd)
}

func (a *ADCEnhanced) ConvertWithKickback(voltage float64, sourceImpedance float64) int {
    // Kickback causes transient on input
    kickbackV := a.KickbackCharge / (1e-12 + sourceImpedance * 1e-12)

    // This decays before comparison settles (usually)
    // But adds noise floor
    kickbackNoise := kickbackV * 0.1 // 10% residual

    return a.ConvertWithNonlinearity(voltage + kickbackNoise)
}
```

---

#### Improvement #11: SET/RESET Asymmetry

**Problem:** Erase (RESET) is typically 20% slower than program (SET):

```go
// In device_state.go or timing model
type AsymmetricTiming struct {
    SetPulseWidth   float64 // ns (typical 100ns)
    ResetPulseWidth float64 // ns (typical 120ns)
    SetVoltage      float64 // V (positive coercive)
    ResetVoltage    float64 // V (negative coercive, often higher magnitude)
}

func (a *AsymmetricTiming) GetPulseWidth(isReset bool) float64 {
    if isReset {
        return a.ResetPulseWidth
    }
    return a.SetPulseWidth
}
```

---

#### Improvement #12: Retention Loss Model

**Problem:** Stored states drift over time due to depolarization:

```go
// In device_state.go
type RetentionModel struct {
    LastWriteTime [][]time.Time // When each cell was last written
    RetentionTau  float64       // Time constant (typical 10⁸ seconds at 85°C)
}

func (r *RetentionModel) GetRetainedLevel(row, col int, originalLevel int, now time.Time) int {
    elapsed := now.Sub(r.LastWriteTime[row][col]).Seconds()

    // Exponential decay toward middle state (level 15)
    decayFactor := math.Exp(-elapsed / r.RetentionTau)
    middleLevel := 15.0

    retainedLevel := middleLevel + (float64(originalLevel) - middleLevel) * decayFactor
    return int(math.Round(retainedLevel))
}
```

---

## 3. Implementation Roadmap

### Phase 1: Critical Accuracy (2-3 weeks)

| Task | Files | Priority |
|------|-------|----------|
| Nonlinear conductance model | `device_state.go` | P1 |
| Sneak path calculation | `device_state.go` | P1 |
| IR drop for large arrays | `analysis.go`, `device_state.go` | P1 |
| Write disturb tracking | `device_state.go` | P1 |

**Acceptance Criteria:**
- Transfer function error < 10% vs. literature
- Passive mode shows realistic sneak path impact
- 64×64 array shows measurable IR drop

### Phase 2: Enhanced Realism (2-3 weeks)

| Task | Files | Priority |
|------|-------|----------|
| Temperature slider | `device_state.go`, `tab_unified.go` | P2 |
| Switching statistics | `device_state.go` | P2 |
| Endurance tracking | `device_state.go`, `tab_unified.go` | P2 |
| TIA frequency response | `tia.go` | P2 |

**Acceptance Criteria:**
- Temperature effects visible in GUI
- Repeated writes show statistical variation
- Endurance counter tracks cycles

### Phase 3: Polish (1-2 weeks)

| Task | Files | Priority |
|------|-------|----------|
| Charge pump transients | `chargepump.go` | P3 |
| ADC kickback | `adc.go` | P3 |
| SET/RESET asymmetry | Timing model | P3 |
| Retention model | `device_state.go` | P3 |

---

## 4. Validation Strategy

### 4.1 Unit Tests

```go
// Test nonlinear conductance matches literature
func TestConductanceExponential(t *testing.T) {
    ds := NewDeviceStateWithModel(ConductanceExponential)

    // Level 0 should be Gmin
    assert.InDelta(t, ds.Gmin, ds.GetConductance(0, 29), 1e-9)

    // Level 29 should be Gmax
    assert.InDelta(t, ds.Gmax, ds.GetConductance(29, 29), 1e-9)

    // Midpoint should be geometric mean (exponential property)
    midG := math.Sqrt(ds.Gmin * ds.Gmax)
    assert.InDelta(t, midG, ds.GetConductance(15, 29), midG*0.1)
}

// Test sneak path magnitude matches literature
func TestSneakPathMagnitude(t *testing.T) {
    ds := NewDeviceState(64, 64, DefaultTIA(), DefaultADC())
    ds.SetPassiveMode(true)

    // Fill array with mid-level conductance
    weights := makeUniformArray(64, 64, 15)

    ds.ComputeWithSneakPaths(weights, 30)

    // Sneak should be 5-20% of signal
    for r := 0; r < 64; r++ {
        sneakFrac := ds.sneakFraction[r]
        assert.True(t, sneakFrac > 0.05 && sneakFrac < 0.20,
            "Sneak fraction should be 5-20%%, got %.1f%%", sneakFrac*100)
    }
}
```

### 4.2 Integration Tests

```go
// Test end-to-end transfer function
func TestTransferFunctionAccuracy(t *testing.T) {
    // Configure realistic peripherals
    dac := DefaultDAC()
    adc := DefaultADC()
    tia := DefaultTIA()
    pump := DefaultChargePump()

    tf := ComputeTransferFunction(dac, adc, tia, pump)

    // Count errors (output != input)
    errors := 0
    for i := 0; i < 30; i++ {
        if tf.ADCLevels[i] != i {
            errors++
        }
    }

    // Allow max 2 level errors (93% accuracy)
    assert.LessOrEqual(t, errors, 2,
        "Transfer function should have <10%% level errors")
}
```

### 4.3 Literature Validation

Compare simulation results against published data:

| Metric | Target | Source |
|--------|--------|--------|
| MNIST accuracy (passive 64×64) | 85-92% | Literature |
| MNIST accuracy (1T1R 64×64) | 96-98% | Literature |
| Sneak path error (passive) | 5-20% | Crossbar_Sneak_Path_Analysis |
| Write disturb rate (V/2) | <1% | Multi_Level_FeFET_Programming |

---

## 5. Summary

### What Changes

| Current | Proposed | Impact |
|---------|----------|--------|
| Linear G(level) | Exponential/Preisach | +10-15% accuracy |
| No sneak paths | 3-cell model | +15-20% passive accuracy |
| No IR drop | Per-cell calculation | +3-5% large array accuracy |
| No write disturb | V/2 tracking | Realistic passive write |
| Room temp only | 77K-400K range | Cryogenic/automotive demo |
| Deterministic write | Gaussian Vth | Realistic noise |
| Infinite endurance | Fatigue model | Reliability visualization |

### Expected Outcomes

1. **Simulation-to-silicon correlation:** <5% error (vs. current ~20%)
2. **Passive mode realism:** Proper sneak path impact shown
3. **Temperature sweep:** Cryogenic benefits visible
4. **Endurance tracking:** Wear-out visualization

### Files to Modify

| File | Changes |
|------|---------|
| `device_state.go` | Conductance model, sneak paths, disturb, endurance |
| `analysis.go` | IR drop, enhanced transfer function |
| `tia.go` | Frequency response |
| `adc.go` | Kickback noise |
| `chargepump.go` | Transient response |
| `tab_unified.go` | Temperature slider, endurance display |
| `physics.yaml` | New parameters for all models |

---

## Related Documentation

- **[circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md)** - CIM physics background
- **[circuits.operations.md](circuits.operations.md)** - 0T1R vs 1T1R operations
- **[circuits.research.md](circuits.research.md)** - Peripheral circuit research
- **[../development/TESTING.md](../development/TESTING.md)** - Test framework

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
