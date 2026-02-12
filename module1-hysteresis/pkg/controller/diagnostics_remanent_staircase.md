# Landau-Khalatnikov Polydomain Ensemble Remanent Staircase Diagnostic

This diagnostic checks that the Landau-Khalatnikov (LK) solver, when run in *polydomain ensemble* mode, produces a multilevel **remanent staircase** at `E=0` under a write/verify-like loop.

It is a prerequisite for claiming multilevel LK behavior in Module 1.

## Diagnostic Procedure

Implemented by `TestLandauKEnsemble_RemanentStaircase_Superlattice` in `landau_remanent_sweep_test.go`.

Configuration (literature superlattice material):

- LK solver with `EnableNoise=false`
- `UseNLS=true` (partial switching mechanism)
- `EnableEnsemble(numDomains=256, seed=0)`
- pulse/relax sweep settings:
  - `dt = 2e-9 s`
  - `pulseSteps = 6`
  - `relaxSteps = 20` (at `E=0`)
  - sweep `k=0..80` with `E = (3.5*k/80)*Ec`

For each sweep point:

1. Initialize from negative saturation (`SetState(-Ps)`).
2. Apply pulse at sweep field.
3. Relax/verify at `E=0`.
4. Measure `(P_rem, level)` where `level = levelFromP(P_rem, Ps, 30)`.

## Acceptance Metrics

1. **Multilevel claim gate**
   - Metric: number of distinct mapped remanent levels over sweep.
   - Acceptance: `distinctLevels >= 20`.

2. **Determinism (regression mode)**
   - Metric: full level sequence over sweep.
   - Acceptance: two runs with same seed/config match exactly.

3. **Verify-at-`E=0` stability**
   - Metrics:
     - final-step level stability during zero-field relax,
     - `deltaFrac = |P_end - P_prev| / |Ps|`.
   - Acceptance:
     - zero final-step level changes,
     - `maxDeltaFrac <= 1e-3`.

4. **Distribution visibility**
   - The test logs per-level count and mean remanent polarization so the `(P_rem, level)` staircase can be audited numerically.
