# Test Plan: Module1-Hysteresis Physics and Data Integrity Testing

**Plan ID:** module1-hysteresis-physics-tests
**Created:** 2026-01-28
**Status:** Ready for Review

---

## 1. Requirements Summary

### 1.1 Objective
Create comprehensive end-to-end and unit tests for module1-hysteresis to validate physics accuracy and data integrity against peer-reviewed literature values.

### 1.2 Physics Properties Requiring Validation

| Property | Description | Reference |
|----------|-------------|-----------|
| **Preisach Model Correctness** | Hysteron behavior, distribution, Mayergoyz classical model | Mayergoyz (1991) |
| **P-E Hysteresis Curve Accuracy** | Loop shape, saturation, remanence, coercivity | Park et al. 2015, Nature Commun. 2025 |
| **Multi-Level State Quantization** | 30 discrete states with correct spacing | Dr. Tour COSM 2025, Oh et al. 2017 (32 states) |
| **Temperature Dependence** | Ec(T), Pr(T) scaling with Curie temperature | Standard ferroelectric physics |
| **Field Cycling Behavior** | Wake-up, fatigue, endurance degradation | Pesic et al. 2016 |
| **KAI Switching Dynamics** | Kolmogorov-Avrami-Ishibashi time-domain response | Literature standard (n ~ 2) |

### 1.3 Current Test Coverage Analysis

**Existing Tests (51 total across 4 files):**
- `ferroelectric_test.go`: 21 tests - Basic hysteresis, materials, temperature
- `preisach_advanced_test.go`: 17 tests - Mayergoyz model, minor loops, fatigue
- `engine_test.go`: 13 tests - Thread safety, waveforms, history

**Identified Gaps:**
1. No validation against peer-reviewed physics constants with tolerances
2. No edge case testing (boundary values, extreme fields, numerical stability)
3. No cross-validation between simple and advanced Preisach models
4. No hysteresis loop area/energy calculation tests
5. No Preisach congruency property tests
6. No comprehensive KAI dynamics validation with golden values
7. No data export/format integrity tests
8. No regression tests with golden reference data
9. No imprint field modeling tests
10. No numerical precision/stability tests under extreme conditions
11. No capacitance calculation validation

### 1.4 Relationship Between Existing and New Tests

**Important:** Existing tests use broader tolerance ranges for backward compatibility:
- `TestMaterialParameters` uses Pr range 10-50 uC/cm^2 (ferroelectric_test.go:155-156)
- This is intentionally permissive to allow various HZO compositions

**New physics validation tests** use stricter literature-validated ranges:
- New Pr validation uses 15-34 uC/cm^2 (Nature Commun. 2025)
- Purpose: Validate against peer-reviewed physics, not just "reasonable values"

**Both can coexist:**
- Existing tests = backward compatibility checks (broad ranges)
- New tests = physics validation (strict literature-based ranges)
- A material passing existing tests but failing new tests indicates it may be
  outside the peer-reviewed parameter space for standard HZO

---

## 2. Acceptance Criteria

### 2.1 Physics Tolerance Specifications

All physics validations must use explicit tolerances based on literature variance:

| Parameter | Expected Range | Tolerance | Source |
|-----------|---------------|-----------|--------|
| Pr (HZO) | 15-34 uC/cm^2 (0.15-0.34 C/m^2) | +/- 20% | Nature Commun. 2025 |
| Pr (Cryo 4K) | 75 uC/cm^2 (0.75 C/m^2) | +/- 15% | Adv. Elec. Mat. 2024 |
| Ec (HZO) | 0.6-1.5 MV/cm (0.6e8-1.5e8 V/m) | +/- 25% | Nature Commun. 2025 |
| Ps/Pr ratio | 1.05-1.5 | exact range | Standard physics (1.05 min for CryogenicHZO) |
| Tc (Curie) | 700-750 K | +/- 5% | Literature |
| KAI exponent n | 1.8-2.5 | exact range | Domain growth physics |
| Ec(T) exponent | 0.4-0.6 | +/- 0.1 | Mean-field theory |

### 2.2 Numerical Precision Requirements

| Test Category | Precision Requirement |
|---------------|----------------------|
| Polarization values | 1e-6 C/m^2 absolute |
| Normalized polarization | 1e-4 relative |
| Energy calculations | 1e-15 J (fJ level) |
| Conductance mapping | 1e-9 S (nS level) |
| Temperature scaling | 1 K absolute |

### 2.3 Behavioral Requirements

| Requirement | Pass Criterion |
|-------------|----------------|
| Hysteresis memory | P(E=0) after positive saturation > 0 |
| Loop closure | Final P within 1% of initial P after full cycle |
| Saturation bounds | -Ps <= P <= Ps at all times |
| Temperature monotonicity | Ec and Pr decrease monotonically with increasing T |
| Congruency property | Minor loops on same branch have parallel edges |
| Wake-up effect | Pr increases for first ~100 cycles then stabilizes |

---

## 3. Implementation Steps

### 3.1 New Test File: `physics_validation_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/physics_validation_test.go`

#### Task 3.1.1: Literature Constants Validation
```
Function: TestLiteraturePolarizationConstants
Purpose: Validate all material Pr/Ps values against peer-reviewed ranges
Coverage: All 8 HZOMaterial variants from material.go:67-315
Acceptance: Each material's Pr must fall within documented literature range
```

#### Task 3.1.2: Coercive Field Validation
```
Function: TestLiteratureCoerciveFieldConstants
Purpose: Validate all material Ec values against peer-reviewed ranges
Coverage: All 8 HZOMaterial variants
Acceptance: Each material's Ec must fall within 0.6-5.0 MV/cm depending on type
```

#### Task 3.1.3: Ps/Pr Ratio Validation
```
Function: TestPolarizationSaturationRatio
Purpose: Verify Ps > Pr with ratio in physically valid range (1.05-1.5)
Coverage: All materials
Acceptance: 1.05 < Ps/Pr < 1.5 for all materials
Note: CryogenicHZO has ratio ~1.07 (Ps=0.80, Pr=0.75) which is valid -
      cryogenic conditions enhance Pr while Ps remains similar, leading to
      tighter Ps/Pr ratios. This is consistent with enhanced polarization
      physics at 4K temperatures.
```

#### Task 3.1.4: Temperature Scaling Validation
```
Function: TestTemperatureScalingPhysics
Purpose: Verify Ec(T) = Ec0 * (1 - T/Tc)^0.5 formula
Coverage: material.go:447-461 (CoerciveFieldAtTemp, PolarizationAtTemp)
Acceptance:
  - At T=0.5*Tc: Ec should be ~0.71*Ec0 (+/- 5%)
  - At T=Tc: Ec = 0
  - Above Tc: Ec = 0
```

#### Task 3.1.5: Imprint Field Validation
```
Function: TestImprintFieldPhysics
Purpose: Verify ImrintField (asymmetric shift) is physically reasonable
Coverage: HZOMaterial.ImprintField field
Acceptance:
  - ImprintField < 0.1*Ec for fresh devices (minimal imprint)
  - ImprintField can grow with cycling/stress
  - Hysteresis loop shifts by ImprintField amount when applied
  - Positive imprint shifts loop toward negative E, negative toward positive E
```

#### Task 3.1.6: Capacitance Calculation Validation
```
Function: TestCapacitanceCalculation
Purpose: Verify capacitance calculation follows parallel plate formula
Coverage: HZOMaterial capacitance-related calculations
Acceptance:
  - C = epsilon_0 * epsilon_r * Area / Thickness
  - Typical values: 1-100 fF for 10nm film, 100nm^2 area
  - C_LF > C_HF (low frequency permittivity higher due to domain contribution)
  - Ratio EpsilonLF/Epsilon should be 1.1-1.5 for HZO materials
```

### 3.2 New Test File: `preisach_physics_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/preisach_physics_test.go`

#### Task 3.2.1: Hysteron Constraint Validation
```
Function: TestHysteronAlphaBetaConstraint
Purpose: Verify all hysterons satisfy alpha > beta (fundamental Preisach constraint)
Coverage: preisach_advanced.go:77-106 (initializeHysterons)
Acceptance: 100% of hysterons must have Alpha > Beta
```

#### Task 3.2.2: Preisach Distribution Normalization
```
Function: TestPreisachDistributionNormalization
Purpose: Verify sum of weighted hysteron contributions equals Ps at saturation
Coverage: preisach_advanced.go:108-146 (initializeDistribution)
Acceptance: |P_max - Ps| < 1% of Ps
```

#### Task 3.2.3: Preisach Congruency Property
```
Function: TestPreisachCongruencyProperty
Purpose: Verify minor loops starting from same P value are congruent (parallel)
Coverage: preisach_advanced.go:265-287 (GetMinorLoop)
Acceptance: Two minor loops starting at same P should have |dP/dE| difference < 5%
```

#### Task 3.2.4: Wiping-Out Property Test
```
Function: TestPreisachWipingOutProperty
Purpose: Verify turning point memory erasure follows Preisach rules
Coverage: preisach.go:107-119 (addTurningPoint)
Acceptance: After returning to a previous extreme, intermediate history is erased
```

#### Task 3.2.5: Cross-Model Consistency
```
Function: TestSimpleVsAdvancedPreisachConsistency
Purpose: Verify simple (tanh) and advanced (Mayergoyz) models produce similar major loops
Coverage: preisach.go vs preisach_advanced.go
Test Conditions:
  - Temperature fixed at 300K for fair comparison
  - Same material (DefaultHZO)
  - Same field amplitude (2*Ec)
  - Same number of points (100)
Acceptance: Loop area difference < 15%, Pr difference < 20%
```

### 3.3 New Test File: `hysteresis_loop_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/hysteresis_loop_test.go`

#### Task 3.3.1: Loop Area Calculation
```
Function: TestHysteresisLoopArea
Purpose: Calculate loop area (energy dissipation) and validate against theory
Coverage: Hysteresis loop generation functions
Acceptance: Area > 0 (clockwise traversal), Area ~ 4*Pr*Ec (order of magnitude)
Formula: Area = integral(P dE) over loop = energy per unit volume per cycle
```

#### Task 3.3.2: Saturation Symmetry
```
Function: TestSaturationSymmetry
Purpose: Verify |+Ps| == |-Ps| within numerical tolerance
Coverage: GetHysteresisLoop in both models
Acceptance: |max(P) + min(P)| < 0.01*Ps
```

#### Task 3.3.3: Loop Closure Test
```
Function: TestHysteresisLoopClosure
Purpose: Verify loop returns to starting point after complete cycle
Coverage: GetHysteresisLoop functions
Acceptance: |P_final - P_initial| < 0.01*Ps for major loop
```

#### Task 3.3.4: Remanent Polarization Extraction
```
Function: TestRemanentPolarizationExtraction
Purpose: Extract Pr from hysteresis loop by finding P at E=0 on descending branch
Coverage: Loop generation and data extraction
Acceptance: Extracted Pr matches material.Pr within 10%
```

#### Task 3.3.5: Coercive Field Extraction
```
Function: TestCoerciveFieldExtraction
Purpose: Extract Ec from hysteresis loop by finding E where P=0
Coverage: Loop generation and interpolation
Acceptance: Extracted Ec matches material.Ec within 15%
```

### 3.4 New Test File: `kai_dynamics_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/kai_dynamics_test.go`

#### Task 3.4.1: KAI Model Mathematical Form
```
Function: TestKAIModelMathematicalForm
Purpose: Verify P(t) = Ps * [1 - exp(-(t/tau)^n)] produces correct S-curve
Coverage: preisach_advanced.go:366-403 (SimulateDomainSwitching)
Acceptance:
  - At t=0: P ~ 0
  - At t=tau: P ~ 0.63*Ps (1 - 1/e)
  - At t=3*tau: P > 0.95*Ps
```

#### Task 3.4.2: KAI Exponent Validation
```
Function: TestKAIExponentPhysics
Purpose: Verify Avrami exponent n ~ 2 (2D domain growth) produces correct kinetics
Coverage: KAI model implementation
Acceptance (quantitative criteria):
  - Inflection point occurs at t=tau within 10% tolerance
  - P(t=tau) reaches 63.2% (1-1/e) of Ps within 5% tolerance
  - P(t=3*tau) reaches 95% of Ps within 2% tolerance
  - Curve is monotonically increasing (dP/dt > 0 for all t > 0)
```

#### Task 3.4.3: Switching Time Constant Validation
```
Function: TestSwitchingTimeTemperatureScaling
Purpose: Verify tau(T) = tau0 * exp(Ea/kT) Arrhenius behavior
Coverage: material.go:439-443 (SwitchingTime)
Acceptance: tau doubles for every ~25K temperature decrease (typical for 0.7eV barrier)
```

### 3.5 New Test File: `state_quantization_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/state_quantization_test.go`

#### Task 3.5.1: 30-Level Quantization Coverage
```
Function: Test30LevelQuantizationFullRange
Purpose: Verify all 30 discrete states span -Ps to +Ps uniformly
Coverage: preisach.go:201-213, preisach_advanced.go:405-437
Acceptance:
  - Level 0: P ~ -Ps
  - Level 15: P ~ 0
  - Level 29: P ~ +Ps
  - Spacing uniform within 1%
```

#### Task 3.5.2: State Separation Margin
```
Function: TestStateSeperationMargin
Purpose: Verify adjacent states are sufficiently separated for reliable sensing
Coverage: DiscreteStates functions
Acceptance: Adjacent state separation > 2*noise_margin (e.g., > 0.02*Ps)
```

#### Task 3.5.3: Conductance Range Mapping
```
Function: TestConductanceRangeMapping
Purpose: Verify 1-100 uS conductance range maps correctly to 30 states
Coverage: preisach_advanced.go:448-458 (polarizationToConductance)
Acceptance:
  - Gmin ~ 1 uS at state 0
  - Gmax ~ 100 uS at state 29
  - Ratio Gmax/Gmin > 50
```

#### Task 3.5.4: Variable Level Count
```
Function: TestVariableLevelCount
Purpose: Test quantization with different level counts (8, 16, 32, 64, 140)
Coverage: DiscreteStates(N) for various N
Acceptance: Uniform spacing verified for all level counts
```

### 3.6 New Test File: `cycling_effects_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/cycling_effects_test.go`

#### Task 3.6.1: Wake-Up Effect Modeling
```
Function: TestWakeUpEffectProgression
Purpose: Verify Pr increases with initial cycling before stabilizing
Coverage: preisach_advanced.go:199-209 (Cycle, wakeup logic)
Acceptance:
  - currentWakeup increases from 0.8 toward 1.0
  - Stabilizes after ~100 cycles (wakeupCycles parameter)
  - Wakeup factor must reach >0.95 after 100 cycles (quantitative threshold)
  - Rate of change decreases monotonically (asymptotic approach)
```

#### Task 3.6.2: Fatigue Degradation Model
```
Function: TestFatigueDegradationModel
Purpose: Verify Pr degrades according to stretched exponential after many cycles
Coverage: material.go:465-468 (EnduranceAtCycles)
Acceptance:
  - At N=0.1*N_endurance: Pr > 90% of initial
  - At N=N_endurance: Pr ~ 70-80% of initial
  - Stretched exponential with beta ~ 0.3
```

#### Task 3.6.3: Endurance Limit Consistency
```
Function: TestEnduranceLimitConsistency
Purpose: Verify endurance limits match documented values for each material
Coverage: All materials' EnduranceCycles field
Acceptance:
  - DefaultHZO: 1e10 (verified IEEE IRPS 2022)
  - FeCIMMaterial: 1e9 (demonstrated)
  - FeCIMMaterialTarget: 1e12 (target only)
```

#### Task 3.6.4: Retention Time Modeling
```
Function: TestRetentionTimeModeling
Purpose: Verify Arrhenius retention loss model at various temperatures
Coverage: material.go:472-485 (RetentionAtTime)
Acceptance:
  - 10-year retention at 85C maintains > 90% Pr
  - Higher temperature accelerates loss exponentially
```

### 3.7 New Test File: `edge_cases_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/edge_cases_test.go`

#### Task 3.7.1: Zero Field Behavior
```
Function: TestZeroFieldBehavior
Purpose: Verify model handles E=0 correctly with memory retention
Coverage: Both Preisach models
Acceptance: P persists at remanent value when E returns to 0
```

#### Task 3.7.2: Extreme Field Values
```
Function: TestExtremeFieldValues
Purpose: Test behavior at E >> Ec (10*Ec, 100*Ec)
Coverage: Update functions in both models
Acceptance: P saturates at +/- Ps, no numerical overflow
```

#### Task 3.7.3: Rapid Field Reversals
```
Function: TestRapidFieldReversals
Purpose: Test many rapid direction changes without instability
Coverage: Model update and history tracking
Acceptance: No memory leaks, bounded history growth, stable output
```

#### Task 3.7.4: Near-Curie Temperature Behavior
```
Function: TestNearCurieTemperatureBehavior
Purpose: Verify smooth transition as T approaches Tc
Coverage: Temperature-dependent functions
Acceptance:
  - Ec and Pr approach 0 smoothly
  - No division by zero or NaN
  - At T > Tc: Ec = 0, Pr = 0
```

#### Task 3.7.5: Negative Temperature Guard
```
Function: TestNegativeTemperatureGuard
Purpose: Verify model rejects or handles T < 0 K gracefully
Coverage: Temperature-dependent functions
Acceptance: Either error return or clamp to T=0
```

#### Task 3.7.6: Numerical Precision at Boundaries
```
Function: TestNumericalPrecisionAtBoundaries
Purpose: Test polarization calculations at P ~= Ps and P ~= -Ps
Coverage: All calculation paths
Acceptance: No precision loss, values stay bounded
```

### 3.8 New Test File: `golden_regression_test.go`

**Location:** `module1-hysteresis/pkg/ferroelectric/golden_regression_test.go`

#### Task 3.8.1: Golden Loop Reference Data
```
Function: TestGoldenHysteresisLoopData
Purpose: Compare generated loops against stored golden reference
Coverage: Major loop generation
Acceptance: RMS error < 2% of Ps compared to golden data
           (2% tolerance to handle platform/floating-point variations)
Data: Store golden_loop_default_hzo.json in testdata/
```

#### Task 3.8.2: Golden Temperature Sweep Data
```
Function: TestGoldenTemperatureSweepData
Purpose: Compare Ec(T), Pr(T) against stored golden reference
Coverage: Temperature scaling functions
Acceptance: Max error < 2% across automotive range (233K-423K)
Data: Store golden_temp_sweep.json in testdata/
```

#### Task 3.8.3: Golden Quantization Data
```
Function: TestGolden30StateQuantization
Purpose: Compare discrete state values against stored golden reference
Coverage: DiscreteStates function
Acceptance: Exact match for all 30 levels (within 1e-10)
Data: Store golden_30_states.json in testdata/
```

### 3.9 Test Data Files

**Location:** `module1-hysteresis/pkg/ferroelectric/testdata/`

#### Task 3.9.1: Create Golden Reference Data
```
Files to create:
  - golden_loop_default_hzo.json: Reference hysteresis loop (E, P arrays)
  - golden_temp_sweep.json: Ec(T), Pr(T) for T = 100K to 700K
  - golden_30_states.json: All 30 discrete state values
  - golden_kai_dynamics.json: P(t) for domain switching simulation
Format: JSON with version field for tracking changes
```

---

## 4. Risk Identification

### 4.1 Technical Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Physics tolerances too tight | Medium | Tests fail unnecessarily | Use literature variance as guide, not arbitrary tight bounds |
| Golden data invalidated by legitimate model improvements | High | Regression tests fail | Version golden data, update with documented rationale |
| Cross-model consistency may be poor due to different approximations | Medium | False test failures | Set realistic 15-20% tolerance for model comparison |
| Temperature edge cases (T~Tc) may have numerical issues | Low | Test failures near Curie temp | Use explicit guards in code, test with margin from Tc |

### 4.2 Implementation Risks

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Large number of new tests may have long runtime | Low | Slow CI | Use table-driven tests, parallelize where possible |
| Testdata files may grow large | Low | Repo bloat | Keep golden data minimal, compress if needed |
| Existing tests may conflict with new stricter tests | Medium | CI failures | Review existing tests first, align tolerances |

---

## 5. Verification Steps

### 5.1 Pre-Implementation Verification

1. Review existing test coverage with `go test -cover ./module1-hysteresis/...`
2. Identify any existing failures before adding new tests
3. Document current physics constant values in code vs. literature

### 5.2 Test Implementation Verification

For each new test file:

1. **Compile Check:** `go build ./module1-hysteresis/pkg/ferroelectric/`
2. **Test Run:** `go test -v -run TestXxx ./module1-hysteresis/pkg/ferroelectric/`
3. **Coverage Report:** `go test -coverprofile=coverage.out && go tool cover -html=coverage.out`

### 5.3 Post-Implementation Verification

1. Full test suite: `go test ./module1-hysteresis/...`
2. Race detection: `go test -race ./module1-hysteresis/...`
3. Benchmark stability: `go test -bench=. ./module1-hysteresis/...`
4. Coverage target: Achieve > 80% coverage for ferroelectric package

---

## 6. Commit Strategy

### 6.1 Commit Sequence

1. **Commit 1:** Add `physics_validation_test.go` - Literature constant validation (includes Imprint + Capacitance)
2. **Commit 2:** Add `preisach_physics_test.go` - Preisach model physics tests
3. **Commit 3:** Add `hysteresis_loop_test.go` - Loop property tests
4. **Commit 4:** Add `kai_dynamics_test.go` - Switching dynamics tests
5. **Commit 5:** Add `state_quantization_test.go` - 30-level quantization tests
6. **Commit 6:** Add `cycling_effects_test.go` - Wake-up/fatigue tests
7. **Commit 7:** Add `edge_cases_test.go` - Boundary condition tests
8. **Commit 8:** Add `golden_regression_test.go` and `testdata/` - Regression tests

### 6.2 Commit Message Format

```
test(module1): add {category} physics validation tests

- TestXxx: validates {property} against {source}
- TestYyy: verifies {behavior} with {tolerance}
...

Coverage: {before}% -> {after}%
```

---

## 7. Success Criteria

### 7.1 Quantitative Criteria

| Metric | Target |
|--------|--------|
| New tests added | 27-37 (includes Imprint + Capacitance tests) |
| Code coverage increase | +10-15% |
| Physics constants validated | 100% of materials |
| Edge cases covered | All identified (6+) |
| Golden regression tests | 4 files with versioned data |

### 7.2 Qualitative Criteria

- All tests pass with `-race` flag
- No flaky tests (deterministic behavior)
- Test names clearly indicate physics property being tested
- Tolerances documented with literature references
- Golden data versioned for reproducibility

---

## 8. Dependencies

### 8.1 Required Before Implementation

- None (tests only, no code changes required)

### 8.2 Optional Enhancements (Future Work)

- Integration with CI/CD pipeline
- Automated golden data regeneration tool
- Physics documentation auto-generation from tests

---

## 9. Appendix: Physics Reference Values

### 9.1 HZO Material Constants (from physics.yaml and literature)

| Material | Pr (C/m^2) | Ec (V/m) | Tc (K) | Levels |
|----------|------------|----------|--------|--------|
| DefaultHZO | 0.25 | 1.2e8 | 723 | 30 |
| FeCIMMaterial | 0.30 | 1.0e8 | 723 | 30 |
| LiteratureSuperlattice | 0.50 | 0.85e8 | 773 | 64 |
| CryogenicHZO | 0.75 | 1.5e8 | 723 | 48 |
| HZOStandard32 | 0.20 | 1.0e8 | 723 | 32 |
| HZOFJT140 | 0.25 | 1.2e8 | 723 | 140 |
| AlScN | 1.20 | 5.0e8 | 1273 | 12 |

### 9.2 Temperature Scaling Formulas

```
Ec(T) = Ec0 * (1 - T/Tc)^0.5    (mean-field approximation)
Pr(T) = Pr0 * (1 - T/Tc)^0.5    (same scaling for Pr)
tau(T) = tau0 * exp(Ea/kT)      (Arrhenius switching time)
```

### 9.3 KAI Model

```
P(t) = Ps * [1 - exp(-(t/tau)^n)]
n ~ 2 for 2D domain growth (typical HZO)
n ~ 3 for 3D nucleation-limited growth
```

---

**Plan Status:** REVISED (Critic Feedback Addressed)

---

## 10. Revision History

### Revision 1 (2026-01-28) - Critic Feedback

**Critical Issues Fixed:**

1. **Ps/Pr ratio tolerance mismatch** - Changed from "1.1-1.4" to "1.05-1.5" consistently
   - Section 2.1 tolerance table: Updated
   - Task 3.1.3: Updated with note explaining CryogenicHZO physics

2. **Task 3.4.2 non-measurable criterion** - Changed from qualitative "curve shape" to quantitative:
   - Inflection point at t=tau within 10%
   - 63.2% of Ps at t=tau within 5%
   - 95% of Ps by t=3*tau within 2%
   - Monotonicity check added

3. **Existing test tolerance conflict** - Added Section 1.4 explaining coexistence:
   - Existing tests = backward compatibility (broad ranges: 10-50 uC/cm^2)
   - New tests = physics validation (strict ranges: 15-34 uC/cm^2)

**Minor Issues Fixed:**

1. **Added Imprint field tests** - Task 3.1.5 added
2. **Added Capacitance calculation tests** - Task 3.1.6 added
3. **Golden RMS tolerance** - Increased from 1% to 2% (Task 3.8.1)
4. **Wake-up quantitative threshold** - Added ">0.95 after 100 cycles" (Task 3.6.1)

**Architect Questions Answered:**

1. CryogenicHZO included with 1.05 minimum (valid physics)
2. Cross-model consistency test fixed at 300K (Task 3.2.5 updated)

PLAN_READY: .omc/plans/module1-hysteresis-physics-tests.md
