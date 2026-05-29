# ADR 0002: ISPP Convergence Guard Logic — Overshoot Protection Design

The ISPP write controller (`module1-hysteresis/pkg/controller/writer.go`) uses a binary-search pulse sequence to converge ferroelectric cells to target levels. During this convergence, the Preisach model can overshoot the target level and flip polarization direction, causing the controller to oscillate endlessly. We designed a multi-layer guard system (guard-band correction, overshoot acceptance, bounds widening) instead of a single threshold-based limit.

## Context

- Preisach-based ferroelectric cells exhibit sharp switching near mid-range levels where the Everett integral crosses a nonlinear threshold.
- When a pulse overshoots, the cell's remanent polarization at E=0 lands above the target. Simply reducing voltage doesn't help because the Preisach state is already on the wrong hysteresis branch. The controller must reset the cell (reverse pulse) and try again with a lower field.
- Overshoots are not failures — they indicate the controller has bracketed the target field. The question is how many overshoots to tolerate before accepting the current level as physics-limited convergence.

## Decision

- **Guard-band correction**: When the cell is at the exact target level but near a bin edge, up to 2 guard pulses nudge polarization toward the bin center. After 2 pulses, accept convergence (prevents direction flip from overshooting the guard sign).
- **Overshoot acceptance threshold**: After 8 consecutive overshoots with error ≤ ±1, accept the level as converged. This gives natural convergence (via reset shortcut) time to find the exact level first.
- **Skip ACCEPT ±1 when guardActive=true**: The guard artificially sets error to ±1; accepting during guard would mask an incorrect convergence. Guard pulses are for centering, not acceptance.
- **OvershootLimit = 30**: After 30 overshoots, declare SUCCESS (not FAILED). Repeated overshoots at different fields prove the material cannot maintain a stable zero-field remanent state at the exact level — counted as physics-limited convergence, not controller failure.
- **Bounds collapse recovery**: When binary search bounds `[VMin, VMax]` collapse (VMin ≥ VMax), widen directionally using `needMore`/`needLess` signals with minimum separation `minBracketWidthFrac * Ec`. Never reset to full `[0, MaxField]` unless direction is unknown.
- **Zero-field verify**: When `absField < 0.01*Ec`, reset bounds to full range for fresh bisection (stale bounds from a previous verify cycle are worse than no bounds).

## Considered Options

1. **Single overshoot limit with FAIL** — rejected. Sharp-switching materials (FeCIM HZO, HZO Custom 14) naturally overshoot mid-range targets; declaring FAIL would make the controller unusable for these materials.
2. **Accept ±1 immediately on any overshoot** — rejected. Would mask genuine convergence failures where the controller is at the wrong field entirely.
3. **No guard-band correction** — rejected. At the target level, Preisach bin-edge quantization (±1 level) causes oscillation between adjacent levels without a small nudge.

## Consequences

- Materials with sharp switching converge to ±1 rather than exact match, which is acceptable for simulation and education use.
- The guard system adds complexity (~250 lines of state machine logic in writer.go) but the alternative (no guard) makes ISPP convergence unreliable for mid-range targets.
- Stress tests (`writer_stress_test.go`) exercise the guard logic across 9 materials × 2 engines and must remain as offline validation tools until integrated into CI.
