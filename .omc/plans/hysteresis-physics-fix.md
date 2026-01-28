# Hysteresis Physics Fix - Comprehensive Plan (v5)

## Executive Summary

This plan addresses four critical physics issues in the hysteresis simulation:
1. **Remove broken ISPP skip-reset optimization** - incompatible with Preisach model physics
2. **Add NLS field-dependent switching time** - Merz law dynamics for realistic switching (per-material parameters)
3. **Improve calibration quality** - wider distribution and larger grid for better level resolution
4. **Verify calibration improvements** - concrete acceptance criteria with measurement

**REMOVED from v4**: FerroX Landau coefficients preset - dead code without Landau-Khalatnikov implementation. Will be added in a future plan when LK dynamics are implemented.

## Problem Analysis

### Issue 1: ISPP Skip-Reset Optimization is Fundamentally Broken

**Evidence from logs:**
```
WRD ISPP: SKIP RESET | current=8 target=5 | same branch=false | skipped=1 saved
```

**Root Cause:** The ISPP optimization assumes ferroelectric memory can perform incremental writes without resetting - this is FALSE for the Preisach model:

1. **Preisach model is path-dependent** - the polarization state depends on the FULL history of field extrema, not just current level
2. **Incremental write field is too weak** - using `stepE = Ec * 0.05 * levelDelta` produces fields ~0.2-0.3xEc which are BELOW the switching threshold
3. **Branch assumption is wrong** - being "on the same branch" in terms of level position (above/below midpoint) does NOT mean hysterons are in the right configuration for incremental switching
4. **No physics basis** - real ISPP works with FLASH memory where charge injection is additive; ferroelectric switching requires proper domain nucleation and growth

**Solution:** Remove ISPP optimization entirely from both Manual and WriteRead modes. Always use full RESET-WRITE cycles for reliable level targeting.

### Issue 2: Missing NLS Field-Dependent Switching Time

**Current state:** The Preisach model uses constant switching time (`Tau` from material parameters), ignoring field dependence.

**Physics reality:** Ferroelectric switching follows Merz's law (NLS = Nucleation-Limited Switching):
```
tau(E) = tau_0 * exp(E_a / E)
```

Where:
- `tau_0` = attempt time (material-dependent, typically 1e-10 to 1e-12 s)
- `E_a` = activation field (material-dependent, typically 10-15 MV/cm for HfO2-based)

**Literature:** Merz, W.J. Phys. Rev. 95, 690 (1954) established the law for BaTiO3. For HfO2-based ferroelectrics, the parameters differ:
- HZO: tau0 ~ 1e-10 s, Ea ~ 10-15 MV/cm (Park et al., Adv. Mater. 2015)
- AlScN: tau0 ~ 1e-11 s, Ea ~ 20-25 MV/cm (higher Ec material)

**Solution:** Add NLS parameters as per-material properties in HZOMaterial struct, with sensible defaults. This allows different materials to have appropriate switching dynamics.

### Issue 3: Poor Calibration Quality (23/30 Duplicate E-fields)

**Evidence from logs:**
```
CRITICAL: Calibration has 23/30 duplicate E-fields.
```

**Root Causes:**
1. **Hysteron distribution too narrow:** `sigma = 0.2*Ec` creates a steep sigmoid with flat saturation regions
2. **Grid size too small:** 50x50 = ~1250 hysterons for 30 levels = ~42 hysterons/level (marginal)

**Solution:** Widen sigma to 0.28*Ec (literature range 0.25-0.35) and increase grid to 60x60.

**Acceptance Criteria:** After changes, duplicate E-fields MUST be < 10 (down from 23).

---

## Detailed Implementation Plan

### Task 1: Remove ISPP Skip-Reset Optimization from Manual Mode

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Complete list of ISPP references to remove in Manual mode:**

| Line | Code | Action |
|------|------|--------|
| 506-546 | Phase 0 ISPP decision logic | Replace with simple RESET |
| 531 | `a.manualIncremental = true` | Remove |
| 532 | `a.manualSkippedResets++` | Remove |
| 535 | Log with `a.manualSkippedResets` | Remove |
| 540 | `!a.manualIncremental` check | Remove |
| 542 | `a.manualIncremental = false` | Remove |
| 543 | `a.manualFullResets++` | Remove |
| 545 | Log with `a.manualFullResets` | Remove |
| 549 | `if !a.manualIncremental` | Remove (always execute RESET) |
| 611-624 | Phase 2 ISPP incremental write | Remove |
| 684 | `if a.manualIncremental` | Remove |
| 688 | Log with `isppType`, `a.manualSkippedResets`, `a.manualFullResets` | Simplify |
| 691 | `a.manualIncremental = false` | Remove |

**Lines 505-575:** Replace ISPP decision logic in Manual mode phase 0

**Current code (lines 505-575):**
```go
case 0: // RESET decision point - check if ISPP can skip RESET
    midLevel := a.numLevels / 2
    currentLevel := a.discreteLevel + 1 // Convert 0-indexed to 1-indexed

    // ISPP OPTIMIZATION: Check if we can skip RESET (same-branch write)
    // ... (ISPP logic) ...

    if canSkipReset && a.manualPhaseTime == 0 {
        // INCREMENTAL WRITE - skip RESET phases (0 and 1)
        a.manualIncremental = true
        a.manualSkippedResets++
        // ...
    } else if a.manualPhaseTime == 0 && !a.manualIncremental {
        // ... rest of RESET logic
    }
```

**New code:**
```go
case 0: // RESET phase - always saturate to known state before writing
    // NOTE: ISPP skip-reset optimization removed (incompatible with Preisach model).
    // Preisach hysteresis is path-dependent - reliable level targeting requires
    // starting from a known saturation state. Incremental writes with sub-Ec
    // fields do not produce predictable switching.

    var resetE float64
    if targetLevel > startLevel {
        // Going UP: first saturate negative (reach level 1)
        resetE = -1.5 * Ec // Match calibration saturation
    } else {
        // Going DOWN: first saturate positive (reach level N)
        resetE = 1.5 * Ec // Match calibration saturation
    }

    // Ramp to reset field
    diff := resetE - a.electricField
    step := rampRate * dt
    if math.Abs(diff) < step {
        a.electricField = resetE
    } else if diff > 0 {
        a.electricField += step
    } else {
        a.electricField -= step
    }

    // Transition when field reached and held briefly
    if a.manualPhaseTime > phaseDuration*0.3 && math.Abs(a.electricField-resetE) < 0.01*Emax {
        a.manualPhase = 1
        a.manualPhaseTime = 0
    }
```

**Lines 609-624:** Remove ISPP incremental write logic in Manual mode phase 2

**Current code:**
```go
// ISPP: For incremental writes, use smaller step fields
if a.manualIncremental {
    // Incremental write: use smaller step pulse based on level delta
    levelDelta := targetLevel - currentLvl
    // Step field: ~0.05*Ec per level, plus margin
    stepE := Ec * 0.05 * float64(abs(levelDelta)+1)
    if goingUp {
        writeE = stepE
    } else {
        writeE = -stepE
    }
    if a.manualPhaseTime == 0 {
        log.Printf("MANUAL ISPP WRITE: incremental step E=%.4f*Ec for delta=%d levels",
            writeE/Ec, levelDelta)
    }
} else if targetIdx < 0 || ...
```

**New code:**
```go
// Use calibrated fields for reliable level targeting
// (ISPP incremental writes removed - incompatible with Preisach model)
if targetIdx < 0 || ...
```

**Lines 682-691:** Remove ISPP type tracking in animation complete logging

**Current code:**
```go
// Log animation result with detailed state (always log for debugging)
isppType := "FULL"
if a.manualIncremental {
    isppType = "ISPP"
}
log.Printf("ANIMATION COMPLETE [%s]: target=%d, final=%d, error=%d, normalizedP=%.4f, skipped=%d, full=%d",
    isppType, targetLevel, finalLevel, levelError, a.normalizedP, a.manualSkippedResets, a.manualFullResets)

// Reset ISPP flag for next animation
a.manualIncremental = false
```

**New code:**
```go
// Log animation result with detailed state
log.Printf("ANIMATION COMPLETE: target=%d, final=%d, error=%d, normalizedP=%.4f",
    targetLevel, finalLevel, levelError, a.normalizedP)
```

### Task 2: Remove ISPP Skip-Reset Optimization from WriteRead Demo Mode

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Complete list of ISPP references to remove in WriteRead mode:**

| Line | Code | Action |
|------|------|--------|
| 844-914 | Phase 0 ISPP decision logic | Replace with simple RESET |
| 867 | `a.wrdIncremental = true` | Remove |
| 868 | `a.wrdSkippedResets++` | Remove |
| 874 | Log with `a.wrdSkippedResets` | Remove |
| 880 | `a.wrdIncremental = false` | Remove |
| 881 | `a.wrdFullResets++` | Remove |
| 892 | Log with `a.wrdFullResets` | Remove |
| 942-958 | Phase 2 ISPP incremental write | Remove |

**Lines 844-914:** Replace ISPP decision logic in WriteRead phase 0

**Current code (lines 844-914):**
```go
case 0: // RESET decision point - check if we can skip
    // Determine if RESET is needed based on branch crossing
    targetInUpperHalf := targetLevel > midLevel
    currentInUpperHalf := currentLevel > midLevel

    // Can skip RESET if:
    // 1. Same half AND moving in correct direction
    // 2. Target is further along the current branch
    canSkipReset := false
    if targetInUpperHalf == currentInUpperHalf {
        // Same half - check if direction is correct
        if targetInUpperHalf {
            canSkipReset = targetLevel >= currentLevel
        } else {
            canSkipReset = targetLevel <= currentLevel
        }
    }

    if canSkipReset {
        // INCREMENTAL WRITE - skip RESET phases
        a.wrdIncremental = true
        a.wrdSkippedResets++
        // ... skip to phase 2
    } else {
        // FULL RESET needed
        // ...
    }
```

**New code:**
```go
case 0: // RESET phase - always saturate to known state
    // NOTE: ISPP skip-reset optimization removed (incompatible with Preisach model).
    // Preisach hysteresis is path-dependent - reliable level targeting requires
    // starting from a known saturation state.

    var resetE float64
    if targetLevel > midLevel {
        // Target in upper half: saturate negative first (reach level 1)
        resetE = -1.5 * Ec
    } else {
        // Target in lower half: saturate positive first (reach level N)
        resetE = 1.5 * Ec
    }
    a.wrdSaturateE = resetE
    log.Printf("WRD RESET: current=%d target=%d | saturating to %.2f*Ec",
        currentLevel, targetLevel, resetE/Ec)

    // Ramp to reset field
    diff := resetE - a.electricField
    step := rampRate * dt
    if math.Abs(diff) < step {
        a.electricField = resetE
    } else if diff > 0 {
        a.electricField += step
    } else {
        a.electricField -= step
    }

    // Transition when field reached and held briefly
    if a.wrdPhaseTimer > phaseDuration*0.25 && math.Abs(a.electricField-resetE) < 0.01*Emax {
        a.wrdResetEndP = a.polarization * 100
        a.wrdResetEndLvl = a.discreteLevel + 1
        log.Printf("WRD PHASE 0->1: RESET done | E=%.3f MV/cm | P=%.2f uC/cm2 | L=%d | target=%d",
            a.electricField/1e8, a.wrdResetEndP, a.wrdResetEndLvl, targetLevel)
        a.wrdPhase = 1
        a.wrdPhaseTimer = 0
    }
```

**Lines 942-958:** Remove ISPP incremental write logic in WriteRead phase 2

**Current code:**
```go
if a.wrdIncremental {
    // INCREMENTAL WRITE: Calculate delta E based on level difference
    levelDelta := targetLevel - currentLvl
    stepE := Ec * 0.05 * float64(abs(levelDelta)+1)
    if goingUp {
        writeE = stepE
    } else {
        writeE = -stepE
    }
    if math.Abs(writeE) > 1.5*Ec {
        writeE = math.Copysign(1.5*Ec, writeE)
    }
    log.Printf("WRD ISPP WRITE: incremental | L%d->L%d | delta=%d | E=%.3f MV/cm",
        currentLvl, targetLevel, levelDelta, writeE/1e8)
} else {
    // FULL WRITE: Use calibration from saturated state
    ...
}
```

**New code:**
```go
// Use calibrated fields for reliable level targeting
// (ISPP incremental writes removed - incompatible with Preisach model)
// Bounds check for calibration array access
if wrdTargetIdx < 0 || wrdTargetIdx >= len(a.calibrationUp) {
    // ... existing fallback code
} else if goingUp {
    // ... existing calibration lookup
} else {
    // ... existing calibration lookup
}
```

### Task 3: Remove ISPP State Variables from App Struct

**File:** `module1-hysteresis/pkg/gui/gui.go`

**Lines 96-99:** Remove WriteRead ISPP tracking fields

**Current code:**
```go
// Incremental write optimization (skip RESET when on same branch)
wrdIncremental    bool // True if skipping RESET (same-branch write)
wrdSkippedResets  int  // Count of RESET phases skipped (energy savings)
wrdFullResets     int  // Count of full RESET cycles (crossing midpoint)
```

**Action:** Remove these 4 lines entirely (comment + 3 fields).

**Lines 116-119:** Remove Manual mode ISPP tracking fields

**Current code:**
```go
// Manual mode ISPP optimization (skip RESET when on same branch)
manualIncremental   bool // True if skipping RESET (same-branch write)
manualSkippedResets int  // Count of RESET phases skipped (energy savings)
manualFullResets    int  // Count of full RESET cycles (crossing midpoint)
```

**Action:** Remove these 4 lines entirely (comment + 3 fields).

### Task 4: Add NLS Field-Dependent Switching Time (Merz Law) - Per-Material

**File:** `module1-hysteresis/pkg/ferroelectric/material.go`

**Add NLS parameters to HZOMaterial struct (after line 64, before closing brace):**

```go
// NLS (Nucleation-Limited Switching) parameters for Merz law dynamics
// tau(E) = Tau0NLS * exp(EaNLS / |E|)
// These are per-material since different ferroelectrics have different switching behavior
Tau0NLS float64 // Attempt time for NLS (s), typically 1e-10 to 1e-12 for HfO2
EaNLS   float64 // Activation field for NLS (V/m), typically 10-15 MV/cm for HfO2
```

**Update DefaultHZO() to include NLS parameters (add before closing brace ~line 95):**
```go
Tau0NLS:         1e-10,  // 100 ps attempt time (HfO2 typical)
EaNLS:           12e8,   // 12 MV/cm activation field
```

**Update all other material presets with appropriate NLS values:**

| Material | Tau0NLS | EaNLS | Rationale |
|----------|---------|-------|-----------|
| DefaultHZO | 1e-10 | 12e8 | Standard HZO |
| LiteratureSuperlattice | 0.5e-10 | 10e8 | Faster switching, lower barrier |
| FeCIMMaterial | 1e-10 | 12e8 | Standard HZO base |
| FeCIMMaterialTarget | 0.5e-10 | 10e8 | Optimized target |
| CryogenicHZO | 2e-10 | 15e8 | Slower at cryo, higher barrier |
| HZOStandard32 | 1e-10 | 12e8 | Standard HZO |
| HZOFJT140 | 2e-10 | 10e8 | FTJ tunneling |
| AlScN | 1e-11 | 22e8 | Higher Ec material, faster attempt |

**File:** `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`

**Add new fields to MayergoyzPreisach struct (after line 57):**

```go
// NLS (Nucleation-Limited Switching) parameters for Merz law dynamics
// tau(E) = tau0_NLS * exp(Ea_NLS / |E|)
// Loaded from material, can be overridden with SetNLSParameters()
Tau0NLS float64 // Attempt time for NLS (s)
EaNLS   float64 // Activation field for NLS (V/m)
```

**Update NewMayergoyzPreisach (after line 75):**

```go
// Load NLS parameters from material (with defaults for backward compatibility)
tau0NLS := material.Tau0NLS
if tau0NLS == 0 {
    tau0NLS = 1e-10 // Default: 100 ps
}
eaNLS := material.EaNLS
if eaNLS == 0 {
    eaNLS = 12e8 // Default: 12 MV/cm
}
```

And add to the struct initialization:
```go
Tau0NLS:      tau0NLS,
EaNLS:        eaNLS,
```

**Add new method after GetEffectivePr (after line 388):**

```go
// GetSwitchingTime returns the field-dependent switching time using Merz's law.
// This implements NLS (Nucleation-Limited Switching) dynamics:
//   tau(E) = tau0 * exp(Ea / |E|)
//
// At high fields (E >> Ea), switching is fast (~100 ps).
// At low fields (E ~ Ec), switching slows dramatically (~100 ns).
//
// Reference: Merz, W.J. "Domain Formation and Domain Wall Motions in
// Ferroelectric BaTiO3 Single Crystals" Phys. Rev. 95, 690 (1954)
// For HfO2-based materials: Park et al., Adv. Mater. 27, 1811 (2015)
func (m *MayergoyzPreisach) GetSwitchingTime(E float64) float64 {
    absE := math.Abs(E)
    if absE < 1e-6 {
        return math.Inf(1) // No switching at zero field
    }

    // Merz law: tau = tau0 * exp(Ea/E)
    tau := m.Tau0NLS * math.Exp(m.EaNLS/absE)

    // Clamp to reasonable range (100 ps to 1 s)
    // Upper bound of 1 second prevents numerical issues in simulations
    // where very low fields would give astronomically long times
    if tau < 1e-10 {
        tau = 1e-10
    }
    if tau > 1.0 {
        tau = 1.0
    }

    return tau
}

// SetNLSParameters allows customizing the Merz law parameters.
// tau0 is the attempt time (typically 1e-10 to 1e-12 s).
// Ea is the activation field (typically 10-15 MV/cm for HfO2).
func (m *MayergoyzPreisach) SetNLSParameters(tau0, Ea float64) {
    m.Tau0NLS = tau0
    m.EaNLS = Ea
    log.Debug("SetNLSParameters: tau0=%.2e s, Ea=%.2f MV/cm", tau0, Ea/1e8)
}
```

**Update SimulateDomainSwitching to use NLS (lines 392-434):**

**Current code:**
```go
func (m *MayergoyzPreisach) SimulateDomainSwitching(Eapplied float64, duration float64, steps int) ([]float64, []float64, []int) {
    // ...
    tau := m.material.Tau // Switching time constant

    // KAI (Kolmogorov-Avrami-Ishibashi) switching dynamics
    // P(t) = Ps * (1 - exp(-(t/tau)^n))
    n := 2.0 // Avrami exponent for 2D domain growth
```

**New code:**
```go
func (m *MayergoyzPreisach) SimulateDomainSwitching(Eapplied float64, duration float64, steps int) ([]float64, []float64, []int) {
    // ...
    // Use field-dependent switching time (Merz/NLS law)
    tau := m.GetSwitchingTime(Eapplied)

    log.Debug("SimulateDomainSwitching: E=%.2f MV/cm, tau(E)=%.2e s (NLS), duration=%.0f ns",
        Eapplied/1e8, tau, duration*1e9)

    // KAI (Kolmogorov-Avrami-Ishibashi) switching dynamics with NLS time
    // P(t) = Ps * (1 - exp(-(t/tau)^n))
    n := 2.0 // Avrami exponent for 2D domain growth
```

### Task 5: Verify Calibration Quality (Already Partially Done)

**Verify these changes are in place:**

**File:** `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`

**Lines 68-70:** Confirm sigma is 0.28 (not 0.2):
```go
AlphaSigma:    material.Ec * 0.28, // 28% distribution (Mayergoyz literature: 0.25-0.35)
BetaMean:      -material.Ec,      // -Ec
BetaSigma:     material.Ec * 0.28,
```

**File:** `module1-hysteresis/pkg/gui/gui.go`

**Lines 334, 396:** Confirm grid size is 60:
```go
preisachGridSize := 60  // High-resolution physics simulation
```

**File:** `module1-hysteresis/pkg/gui/controls.go`

**Line ~148:** Confirm grid size is 60:
```go
a.preisach = ferroelectric.NewMayergoyzPreisach(a.material, 60)
```

**ACCEPTANCE CRITERIA for calibration:**
- Duplicate E-fields MUST be < 10 (current: 23)
- Measure by checking log output: "CRITICAL: Calibration has X/30 duplicate E-fields"
- If X >= 10, the calibration parameters need further tuning

### Task 6: Add NLS Switching Time Test

**File:** `module1-hysteresis/pkg/ferroelectric/preisach_advanced_test.go`

**Add after TestPECurveSmoothness:**

```go
// TestNLSSwitchingTime verifies the Merz law switching time calculation.
func TestNLSSwitchingTime(t *testing.T) {
    material := DefaultHZO()
    model := NewMayergoyzPreisach(material, 50)

    Ec := material.Ec

    // Test cases: field -> expected tau range
    testCases := []struct {
        field    float64
        tauMin   float64
        tauMax   float64
        desc     string
    }{
        {2.0 * Ec, 1e-10, 1e-8, "High field (2*Ec)"},
        {1.5 * Ec, 1e-9, 1e-7, "Moderate field (1.5*Ec)"},
        {1.1 * Ec, 1e-8, 1e-5, "Near threshold (1.1*Ec)"},
        {0.5 * Ec, 1e-6, 1.0, "Below Ec (0.5*Ec)"},
    }

    for _, tc := range testCases {
        tau := model.GetSwitchingTime(tc.field)
        if tau < tc.tauMin || tau > tc.tauMax {
            t.Errorf("%s: tau=%.2e, expected [%.2e, %.2e]", tc.desc, tau, tc.tauMin, tc.tauMax)
        } else {
            t.Logf("%s: tau=%.2e s (OK)", tc.desc, tau)
        }
    }
}

// TestNLSFieldDependence verifies switching time increases as field decreases.
func TestNLSFieldDependence(t *testing.T) {
    material := DefaultHZO()
    model := NewMayergoyzPreisach(material, 50)

    Ec := material.Ec
    fields := []float64{2.0 * Ec, 1.5 * Ec, 1.2 * Ec, 1.0 * Ec}

    var prevTau float64 = 0
    for _, E := range fields {
        tau := model.GetSwitchingTime(E)
        if prevTau > 0 && tau <= prevTau {
            t.Errorf("Switching time should increase as field decreases: E=%.2f*Ec gave tau=%.2e (prev=%.2e)",
                E/Ec, tau, prevTau)
        }
        t.Logf("E=%.2f*Ec -> tau=%.2e s", E/Ec, tau)
        prevTau = tau
    }
}

// TestNLSPerMaterial verifies different materials have different NLS parameters.
func TestNLSPerMaterial(t *testing.T) {
    hzo := DefaultHZO()
    alscn := AlScN()

    modelHZO := NewMayergoyzPreisach(hzo, 50)
    modelAlScN := NewMayergoyzPreisach(alscn, 50)

    // At same normalized field (1.5*Ec), AlScN should have different tau
    fieldHZO := 1.5 * hzo.Ec
    fieldAlScN := 1.5 * alscn.Ec

    tauHZO := modelHZO.GetSwitchingTime(fieldHZO)
    tauAlScN := modelAlScN.GetSwitchingTime(fieldAlScN)

    t.Logf("HZO at 1.5*Ec: tau=%.2e s (Tau0NLS=%.2e, EaNLS=%.2e)", tauHZO, modelHZO.Tau0NLS, modelHZO.EaNLS)
    t.Logf("AlScN at 1.5*Ec: tau=%.2e s (Tau0NLS=%.2e, EaNLS=%.2e)", tauAlScN, modelAlScN.Tau0NLS, modelAlScN.EaNLS)

    // They should be different (AlScN has higher EaNLS but faster Tau0NLS)
    if tauHZO == tauAlScN {
        t.Errorf("Expected different switching times for different materials")
    }
}
```

---

## Acceptance Criteria

1. **ISPP removed:** No `canSkipReset`, `wrdIncremental`, or `manualIncremental` logic in simulation.go
2. **State variables removed:** No `wrdIncremental`, `wrdSkippedResets`, `wrdFullResets`, `manualIncremental`, `manualSkippedResets`, `manualFullResets` in gui.go
3. **All writes use RESET:** Every level transition goes through saturation phase first
4. **NLS implemented:** `GetSwitchingTime(E)` returns field-dependent tau following Merz law
5. **NLS is per-material:** `Tau0NLS` and `EaNLS` fields exist in HZOMaterial and are used
6. **Tests pass:** All existing tests plus new NLS tests pass
7. **Level accuracy:** Write operations achieve target level within +/-2 (verified in Manual mode)
8. **Calibration quality:** Duplicate E-fields < 10 (verified in logs after startup)

---

## Verification Commands

```bash
# Run all physics tests
go test ./module1-hysteresis/pkg/ferroelectric/... -v

# Run specific NLS tests
go test ./module1-hysteresis/pkg/ferroelectric/... -v -run "NLS"

# Delete old calibration to force recalibration with new parameters
rm -f data/hysteresis_calibration.json

# Build and test application
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools

# In the app:
# 1. Open Hysteresis module
# 2. Wait for calibration - CHECK LOGS: "duplicate E-fields" MUST be < 10
# 3. Switch to Manual mode
# 4. Click levels: 1 -> 15 -> 30 -> 5 -> 25
# 5. Verify each target hit within +/-2 levels
# 6. Check logs: NO "ISPP" or "incremental" messages should appear
```

---

## File Summary

| File | Lines | Action | Description |
|------|-------|--------|-------------|
| `simulation.go` | 505-575 | Replace | Remove ISPP from Manual phase 0 |
| `simulation.go` | 611-624 | Remove | Delete ISPP incremental write logic |
| `simulation.go` | 684-691 | Simplify | Remove ISPP type tracking in logs |
| `simulation.go` | 844-914 | Replace | Remove ISPP from WriteRead phase 0 |
| `simulation.go` | 942-958 | Remove | Delete ISPP incremental write in WriteRead |
| `gui.go` | 96-99 | Remove | Delete wrdIncremental/wrdSkippedResets/wrdFullResets fields (4 lines) |
| `gui.go` | 116-119 | Remove | Delete manualIncremental/manualSkippedResets/manualFullResets fields (4 lines) |
| `material.go` | ~64 | Add | NLS parameters (Tau0NLS, EaNLS) to HZOMaterial struct |
| `material.go` | all presets | Add | NLS parameter values to each material function |
| `preisach_advanced.go` | ~57 | Add | NLS parameters (Tau0NLS, EaNLS) to MayergoyzPreisach struct |
| `preisach_advanced.go` | ~75 | Add | Initialize NLS parameters from material in constructor |
| `preisach_advanced.go` | ~388 | Add | GetSwitchingTime() and SetNLSParameters() methods |
| `preisach_advanced.go` | ~400 | Edit | Use NLS tau in SimulateDomainSwitching |
| `preisach_advanced_test.go` | EOF | Add | TestNLSSwitchingTime, TestNLSFieldDependence, TestNLSPerMaterial |

---

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Removing ISPP breaks energy tracking | Low | Energy is still calculated per cycle; just no "skipped" count |
| NLS slows simulation at low fields | Medium | Clamp tau to 1s max; upper bound prevents numerical issues |
| Calibration file invalidated | Low | Auto-recalibration on mismatch already implemented |
| Line number drift | Medium | Use search patterns, verify context before editing |

---

## Commit Strategy

Single commit with message:
```
fix(hysteresis): remove broken ISPP optimization, add NLS switching

- Remove ISPP skip-reset logic from Manual and WriteRead modes
  (incompatible with Preisach path-dependent physics)
- Add Merz law field-dependent switching time (NLS dynamics)
- NLS parameters (Tau0NLS, EaNLS) are now per-material properties
- Verify calibration parameters (sigma=0.28, grid=60) for quality

The ISPP optimization assumed incremental writes work like FLASH
memory, but ferroelectric Preisach model requires full reset-write
cycles for reliable level targeting. All level transitions now
properly saturate before applying calibrated write field.

Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

---

## Removed from This Plan (Future Work)

**FerroX Landau Coefficients:** The v4 plan included adding `LandauAlpha`, `LandauBeta`, `LandauGamma` fields to HZOMaterial and a `FerroXHZO()` preset. This has been REMOVED because:

1. These coefficients are not consumed by the current Preisach model
2. Full Landau-Khalatnikov dynamics implementation is out of scope
3. Adding unused fields creates confusion and dead code

**When to add Landau coefficients:**
- When implementing Landau-Khalatnikov (LK) dynamics for domain switching
- When adding free energy landscape visualization
- When implementing temperature-dependent phase transitions via Landau theory

This will be tracked as a separate future enhancement.

---

PLAN_READY: .omc/plans/hysteresis-physics-fix.md
