# Polydomain Landau-Khalatnikov (LK) Target Behavior Specification

Status: Drafted for LK-PD-1 (acceptance spec)

## Purpose

Define the target behavior for **polydomain LK** in FeCIM Module 1 so that write/verify at zero field (`E=0`) supports true multilevel operation.

This spec explicitly rejects binary-only behavior (two stable wells) as sufficient for multilevel ISPP claims.

## Scope

- Module: `module1-hysteresis`
- Engine: Landau-Khalatnikov with ensemble/polydomain behavior enabled
- Verification condition: remanence measured after relax/verify at `E=0`
- Level mapping: use existing level quantizer (`levelFromP`) to map `P_rem -> level` in `1..30`

## Definitions

- `P_rem`: remanent polarization after write pulse and zero-field relaxation.
- Verify-at-`E=0`: post-pulse relaxation/measurement phase with field set to zero.
- Level mapping: quantization of normalized polarization `P/|Ps|` into 30 controller levels.
- Stable level: mapped level does not change between final consecutive `E=0` relax steps; final-step drift is bounded.

## Required Target Behavior

### R1. Remanent multilevel stability (core requirement)

For polydomain LK, verify-at-`E=0` must yield a **staircase of stable remanent levels** via level mapping. Behavior must not collapse to only low/high well occupancy.

Acceptance target:

- Across the full operational pulse sweep, the mapped remanent output should cover the 30-level ladder (or near-full practical coverage under finite sweep resolution), rather than only 2 levels.

### R2. Stability at verify point (`E=0`)

For each sweep point, after relaxation at zero field:

- mapped level must remain unchanged between final consecutive relax steps;
- normalized last-step drift must satisfy:
  - `|P_end - P_prev| / |Ps| <= 1e-3`.

### R3. Determinism for regression mode

With fixed seed and noise disabled:

- remanent level sequence must be deterministic run-to-run.

## Diagnostic/Regression Criteria

Multilevel claim gate for diagnostic runs:

- Distinct remanent mapped levels observed in sweep must be `>=20`.

Rationale:

- `>=20` is a practical floor indicating broad multilevel behavior before claiming robust polydomain support.
- Full 30-level achievement remains the long-term target behavior for production calibration.

## Non-Goals

- This spec does not prescribe a unique microscopic domain model implementation.
- This spec does not replace material calibration/citation work (handled separately).

## Traceability

- TODO: `LK-PD-1` — behavior spec definition (this file)
- TODO: `LK-PD-2` — remanent staircase sweep diagnostic enforcing `>=20` distinct levels
- Test implementation: `module1-hysteresis/pkg/controller/landau_remanent_sweep_test.go`
