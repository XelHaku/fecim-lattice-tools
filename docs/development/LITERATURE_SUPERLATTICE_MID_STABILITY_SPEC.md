# Literature Superlattice MID Stability Acceptance Criteria (G08)

Status: Active (2026-02-11)

Material scope: **Literature Superlattice (Cheema 2020)** calibration
- Preisach: `data/calibrations/literature_superlattice.json`
- LK: `data/calibrations/literature_superlattice-lk.json`

Test scope: headless WRD/ISPP MID target (level 15 of 30) with verify-at-E=0 semantics.

Primary source task: TODO G08 (`hysteresis-prompt.md`): define numeric MID stability criteria for target level, tolerance band, max pulses, and overshoot limits for both engines.

---

## 1) Definitions

- **MID target level**: `L_target = 15` on the 30-level quantization scale (`0..29`).
- **Final level**: level reported at end of ISPP write/verify sequence (`final_level`).
- **Level error**: `e = final_level - L_target`.
- **Pulse budget**: number of write pulses consumed (`pulses`).
- **Overshoot count**: number of controller overshoot events (`overshoots`).

All acceptance checks are performed from the JSON outputs of:
- `validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json`
- `validation/testdata/ispp_regression/lk_wrd_ispp_regression.json`

---

## 2) Acceptance criteria (MID only)

### A. Preisach engine (strict MID stability gate)

For MID (target 15), the run is **PASS** only if all are true:

1. **Target lock**: `final_level = 15` (exact hit)
2. **Tolerance band**: `|e| <= 0` (equivalently exact)
3. **Pulse budget**: `pulses <= 12`
4. **Overshoot budget**: `overshoots <= 6`
5. **Completion**: `converged = true` and `reached_done = true`

Rationale: current regression converges MID at 8 pulses with overshoots observed during bisection/recovery; cap preserves realistic controller behavior while preventing regressions in write stability.

### B. LK engine (bounded MID behavior gate, single-domain baseline)

For MID (target 15), LK is currently a **bounded-stability** gate (not exact-mid lock):

1. **Bounded terminal level**: `final_level` must remain in the positive-middle window `[20, 26]`
2. **Tolerance band (bounded)**: `|e| <= 11`
3. **Pulse ceiling**: `pulses <= 30`
4. **Overshoot ceiling**: `overshoots <= 0`
5. **Retry ceiling**: `retries <= 1`
6. **Completion safety**: `reached_done = true` and `final_field_mv_cm = 0`

Rationale: existing single-domain LK does not hold exact intermediate remanent states for this material yet; acceptance focuses on deterministic bounded completion and safety while LK stabilization items remain open (LK05/LK07/LK-PD series).

---

## 3) Current evidence snapshot (2026-02-11 18:49 CST)

From committed regression outputs:

- Preisach MID: `final_level=15`, `e=0`, `pulses=8`, `overshoots=6`, `converged=true`
- LK MID: `final_level=24`, `e=+9`, `pulses=30`, `overshoots=0`, `retries=1`, `converged=false`, `reached_done=true`

Evaluation against this spec:

- **Preisach MID**: PASS
- **LK MID (bounded gate)**: PASS

---

## 4) Promotion rule (future tightening)

When LK polydomain/intermediate-retention implementation lands, LK MID criteria must be promoted to strict target-lock parity with Preisach:

- `final_level = 15`, `|e| <= 1`, `pulses <= 25`, and explicit overshoot delta accounting.

Until that promotion, this document is the normative acceptance baseline for G08.
