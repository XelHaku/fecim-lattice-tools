# Hysteresis Module Improvement Plan

**Date:** January 31, 2026
**Status:** DRAFT
**Based on:** Technical Reviews (Claude Opus 4.5, Sisyphus AI Agent) & Research Meta-Study

---

## 1. Executive Summary

This document outlines a comprehensive plan to elevate the `module1-hysteresis` simulation engine toward more robust, production-grade modeling. The primary focus is correcting fundamental physics inaccuracies in the simplified model, resolving unit confusion in the write controller, and enhancing the user experience during long-running calibration tasks.

## 2. Critical Physics Fixes (Priority 0)

These issues affect the scientific validity of the simulation and must be addressed first.

### 2.1 Fix Simplified Preisach History
**Current Status:** The simplified `PreisachModel` (`pkg/ferroelectric/preisach.go`) only clamps to saturation, missing proper minor loop branch interpolation.
**Impact:** Minor loops are physically incorrect; unsuitable for accurate simulation.
**Implementation Plan:**
- [ ] Implement proper branch interpolation logic in `applyHistoryCorrection`.
- [ ] Use turning point stack to define enclosing major/minor loop bounds (Mayergoyz, 1991).
- [ ] **Critical:** Implement "Wipe-out" property: Pop points from LIFO stack when field exceeds previous extrema to prevent memory leaks.
- [ ] Interpolate based on E-field position within the current branch pair.
- [ ] **Verification:** `TestPreisachLoopShape` and manual verification of minor loop closure; 1000-cycle endurance test.
- [ ] **Reference:** Mayergoyz, I.D. "Mathematical Models of Hysteresis" *IEEE Trans. Magnetics* (1986); Bartic et al. "Preisach model for ferroelectric capacitors" *J. Appl. Phys.* (2001).

### 2.2 Resolve Voltage vs. Field Unit Confusion
**Current Status:** `WriteController` (`pkg/controller/writer.go`) and ISPP logic ambiguously mix Voltage (V) and Electric Field (MV/cm). e.g., `EcEst` (field) added to `CurrentVoltage`.
**Impact:** Potential for massive scaling errors (orders of magnitude) if thickness is not exactly 1cm (it's 10nm!).
**Implementation Plan:**
- [ ] Refactor `WriteController` to use explicit types or naming: `TargetField` (V/m) vs `AppliedVoltage` (V).
- [ ] Enforce relationship: `V = E * thickness` everywhere.
- [ ] Audit `ISPPCalculator` for similar unit mixing.
- [ ] **Verification:** Code review and unit tests with non-standard thickness (e.g., 20nm) to ensure scaling holds.

### 2.3 Correct KAI Dynamics (NLS Integration)
**Current Status:** KAI model uses constant `tau` (time constant).
**Impact:** Switching speed does not scale with overdrive voltage, contradicting Merz's Law.
**Implementation Plan:**
- [ ] Implement Nucleation-Limited Switching (NLS) model: `tau(E) = tau0 * exp(Ea / |E|)` (Merz's Law).
- [ ] **Critical:** Implement **Adaptive Time-Stepping**. Shrink `dt` when field is near `Ec` to prevent numerical oscillation in NLS dynamics.
- [ ] Update `SimulateDomainSwitching` in `preisach_advanced.go`.
- [ ] **Verification:** Plot switching time vs. voltage; ensure it follows exponential law.
- [ ] **Reference:** "Domain Wall Dynamics in Ferroelectric Materials" (arXiv); *Ferroelectric Domain Switching Dynamics*.

---

## 3. Algorithm & Feature Enhancements (Priority 1)

These improvements extend the capabilities of the module.

### 3.1 Adaptive/Background Calibration
**Current Status:** Calibration blocks the main UI thread; no feedback.
**Implementation Plan:**
- [ ] Move `CalibrationManager` execution to a goroutine ("Worker").
- [ ] Add `progress` channel to report percentage completion.
- [ ] Update GUI to show a progress bar/spinner during calibration.
- [ ] **Verification:** Run calibration, ensure UI remains responsive (window moves/resizes).

### 3.2 Temperature Control & Visualization
**Current Status:** Temperature scaling exists in backend but no GUI override.
**Implementation Plan:**
- [ ] Add `TemperatureSlider` (Standard: 233K - 423K) to Controls panel.
- [ ] Bind slider to `material.SetTemperature()`.
- [ ] Display real-time `Ec(T)` and `Pr(T)` values in Metrics panel.
- [ ] Add **State Stability Indicator**: Calculate Boltzmann probability (`exp(-Eb/kT)`) of state retention. Warn user if thermal noise threatens 30-level stability.
- [ ] **Verification:** Change temp, observe P-E loop shrinking (approaching Tc).
- [ ] **Reference:** Böscke et al. (2011) "Ferroelectricity in hafnium oxide" (Tc ~450°C); Park et al. (2015) (Temperature stability).

### 3.3 Preisach Plane Visualization
**Current Status:** Internal state exists but is invisible.
**Implementation Plan:**
- [ ] Add "Debug View" or separate window for Preisach Plane.
- [ ] Render 2D heatmap/grid of Hysteron states (+1 red, -1 blue).
- [ ] **Verification:** Visual check: as field increases, "switch" front should move across the plane.

---

## 4. UX/GUI Polishing (Priority 2)

Improvements to usability and visual clarity.

### 4.1 Input Validation & Feedback
- [ ] Add visual error state (red border/text) for invalid inputs (e.g., negative frequency).
- [ ] Clamp slider values to safe ranges but visually indicate clamping.

### 4.2 Layout Optimization
- [ ] Group controls into collapsible sections ("Waveform", "Material", "Physics").
- [ ] Reduce visual noise in the left Info panel.
- [ ] Add Tooltips to complex parameters (Ec, Pr, Alpha, Beta).

### 4.3 P-E Plot Interactions
- [ ] Implement Mouse Wheel Zoom for P-E plot.
- [ ] Add "Reset View" button.

---

## 5. Implementation Roadmap & Task breakdown

### Phase 1: Core Physics (Days 1-2)
1. Unit Audit: Rename variables in `writer.go` and `ispp.go`.
2. Fix `PreisachModel` history logic (Implement Stack Management).
3. Integrate field-dependent NLS/KAI tau.
4. Implement Adaptive Time-Stepping for NLS stability (dt scaling).
5. *Milestone: Physics Verification Pass* (`go test ./pkg/ferroelectric -v`)

### Phase 2: Async Architecture (Day 3)
1. Create `CalibrationWorker` struct.
2. Refactor `gui/controls.go` to use channels for calibration.
3. Add ProgressDialog widget.

### Phase 3: UI Features (Days 4-5)
1. Add Temperature Slider & Bindings.
2. Implement Collapsible Accordion for Controls.
3. Add Input Validation hints.
4. *Milestone: UX Review*

---

## 6. Verification Plan

### Automated Tests
Run the following existing test suites after changes:
```bash
# Physics Engine Integrity
go test ./module1-hysteresis/pkg/ferroelectric -run Physics -v

# Literature Validation (Park 2015, etc)
go test ./module1-hysteresis/pkg/ferroelectric -run Literature -v

# ISPP Algorithm correctness
go test ./module1-hysteresis/pkg/ferroelectric -run ISPP -v

# Endurance/Memory Leak Check
go test ./module1-hysteresis/pkg/ferroelectric -run Endurance
```

### Manual Validation Checklist
1. **Unit Consistency:** Set thickness to 100nm. Verify Ec is 1/10th of voltage compared to 10nm setting (for same Field).
2. **Hysteresis Shape:** Run Sine wave. Ensure closed loops with correct Coercivity.
3. **Calibration:** Click "Calibrate". UI should not freeze. Progress bar should appear.
4. **Temp Effect:** Slide temp to 400K. Loop should narrow significantly.

## 7. Key References

1. **Mayergoyz, I.D.** (1986). "Mathematical Models of Hysteresis". *IEEE Transactions on Magnetics*, 22(5), 603-608. (Foundation of the hysteron summation model).
2. **Bartic, A.T., et al.** (2001). "Preisach model for ferroelectric capacitors". *Journal of Applied Physics*, 89(6), 3420. (Tanh distribution adaptation).
3. **Park, M.H., et al.** (2015). "Ferroelectricity and Antiferroelectricity of Doped Thin HfO2-Based Films". *Advanced Materials*. (Source for Pr, Ec, and Temperature parameters).
4. **Cheema, S.S., et al.** (2020). "Enhanced ferroelectricity in ultrathin films grown directly on silicon". *Nature*. (Superlattice 30-state feasibility).
5. **Tour Lab Papers** (Various). *Ferroelectric Analog Switching* & *HZO Switching Pathways*. (Multi-level cell quantization strategies).
