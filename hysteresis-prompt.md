Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to **headless-only work** unless a GUI change is required for correctness of WRD/ISPP.

Objective

- Make the **Write/Read ISPP demo** hit its target levels reliably (no convergence to Ec=0, no infinite loops).
- Ensure the **Frankestein equation** in `docs/hysteresis/hysteresis-gemini.md` is correctly understood and implemented
  (terms, signs, units, and effective viscosity).
- Keep calibration autonomous during runtime so WRD converges quickly without manual intervention.
- **Run headless first, every iteration** and keep improving the ISPP + equation implementation based on those logs.
- Update docs only when behavior or equations change.
- Use the new **full-resolution CSV logging** to debug ISPP state transitions and numerical stability.

Primary Focus (ranked)

1) WRD/ISPP target accuracy (highest priority)
- Read/Write demo must **hit targets** with strict equality (level match).
- No "stuck at E=0" convergence or endless binary search loops.
- Direction logic must use **current vs target** (not stale initial state).
- Overshoot reset only when overshoot truly occurs; no unnecessary saturations between targets.

2) Frankestein equation fidelity
- Implement exactly the unified L-K + depolarization + series-resistance formulation
  from `docs/hysteresis/hysteresis-gemini.md`.
- Verify all terms, signs, and units in logs.

3) Autonomous calibration
- Runtime recalibration should trigger only when convergence is poor and run between targets.
- Calibrations must persist and update calibration manager state coherently.

4) Documentation sync
- Keep `docs/hysteresis/hysteresis.demo.md` aligned with WRD/ISPP behavior.
- Update `docs/hysteresis/hysteresis-gemini.md` only if equation handling changes.

Tasks

1) Frankestein equation (no missing terms)
- Verify: `dP/dt = (E_applied - k_dep*P - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Ensure `rho_eff = rho + (R_series * A / d)` only when `UseEffectiveViscosity=true`.
- Confirm `E_eff = E_applied - k_dep*P` is what the solver actually uses.
- Log: `E_applied`, `E_dep` or `k_dep*P`, `E_eff`, `dG_dP`, `rho_eff`, `Alpha`, `Beta`, `Gamma`, `K_dep`.

2) Numerical method fidelity (time stepping + integration)
- Simulation uses **adaptive sub-stepping**: `dtNominal=1e-4`, `dtMin=1e-6` when `|E - Ec| < 0.1 MV/cm`, `dtMax=0.025`.
- Confirm sub-step size changes in the CSV (`dt_s`) and that no frame skips occur near switching.
- Keep integration explicit per sub-step (no hidden state jumps); verify `sim_time_s` monotonicity.

3) WRD/ISPP target hit guarantee
- Fix direction inference for `target == initial` (use current vs target).
- If `currentLevel == targetLevel`, **exit immediately** with success (no pulses).
- Prevent binary search from collapsing to `VMax=0` when the direction is wrong.
- Keep pre-biasing (+/-Ec) but avoid full saturation unless overshoot is detected.
- Ensure retry logic does not spin indefinitely; failures should be explicit and rare.

4) Autonomous recalibration
- Trigger on repeated overshoots or too many pulses.
- Run recalibration **between targets** to avoid corrupting active state.
- Persist calibration file and sync into `CalibrationManager`.

5) Docs
- Update WRD/ISPP sequencing and calibration behavior in `docs/hysteresis/hysteresis.demo.md`.
- If equation handling changes, update `docs/hysteresis/hysteresis-gemini.md` accordingly.

Validation

- **Universal command (always use this):** `./launch.sh --logger --verbosity debug`
  - Headless physics: add `--mode hysteresis` when validating the L-K equation + ISPP diagnostics.
  - GUI WRD: run without `--mode` and use the latest WRD log to verify target hits.
  - Evidence must include "TARGET HIT" lines (WRD) and no "Unexpected state ... VMax=0" loops.
  - Use the CSV (`logs/hysteresis-<material>-<timestamp>.csv`) to inspect per-step states and ISPP transitions.
  - Key CSV columns for ISPP debugging: `wrd_phase_name`, `controller_state`, `controller_current_field_mv_cm`, `controller_vmin_ec`, `controller_vmax_ec`, `controller_last_error`, `controller_reset_direction`.

Frankestein Equation Checklist (must satisfy each run)

- Uses: `E_eff = E_applied - k_dep*P`.
- Uses: `dP/dt = (E_eff - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Uses: `rho_eff = rho + (R_series * A / d)` only if enabled.
- Logs show all terms at debug verbosity.

WRD/ISPP Correctness Checklist

- Target hit with strict equality for each WRD cycle.
- If `current == target`, success without pulses.
- No convergence to `E~0` caused by wrong direction inference.
- Overshoot reset only on true overshoot.
- No forced saturation between targets unless overshoot recovery requires it.
- Auto-recalibration occurs between targets and is logged.
- CSV must show coherent state transitions: `controller_state` moves APPLY→WAIT→VERIFY and only to RESETTING on overshoot.

ISPP Debugging Checklist (short)

- Confirm APPLY→WAIT→VERIFY sequence before any RESETTING.
- Verify `controller_last_error` sign matches `target - level` (negative = undershoot).
- Check `controller_vmin_ec`/`controller_vmax_ec` shrink toward target without collapsing to 0.
- On overshoot, confirm `controller_reset_direction` is set and reset target polarity is correct.
- Use CSV to ensure `sim_time_s` is monotonic and `dt_s` shrinks near switching.

Regression Guardrails

- If WRD success rate drops or failures appear, treat as regression and fix immediately.
- If binary search collapses to zero or loops > MaxRetries, fix direction/bounds logic.
- Keep a **baseline** with latest WRD log path + key success/failure stats.

Execution Rules (Autonomous)

- Always run **headless** (`--mode hysteresis`) before any GUI checks.
- Always inspect the **newest** headless log under `logs/` for equation + ISPP signals.
- If the newest headless log is older than the current run, **generate a fresh headless log** immediately.
- Inspect the newest WRD log file under `logs/` when GUI WRD is exercised.
- If the newest WRD log is stale, **rerun GUI with logging** before reporting.
- Prefer minimal, targeted changes; avoid unrelated files.
- If validation fails, report exact error output and last command run.
- GUI changes are allowed only to fix WRD/ISPP correctness.

Deliverable

- Concise report:
  - Frankestein equation verification (what terms/logs confirmed).
  - WRD/ISPP target-hit evidence (log lines).
  - Documentation updates (file paths + summary).
  - Gaps/issues and next iteration target.
- Include validation command and log path.

Baseline (update each run)

- Latest headless log path (always newest under `logs/`):
  - logs/2026-02-02_14-09-35-fecim.log
- Headless status:
  - LK term logs present (E_applied/E_dep/E_eff/dG_dP/rho_eff/Alpha/Beta/Gamma/K_dep). ISPP targets did **not** hit: step 1 (pos‑1) timed out before WRITE, step 2 (pos‑2) hit MaxRetries with level stuck at 21 vs target 27, step 3 (neg‑1) timed out with failures.
- Latest WRD log path (always newest under `logs/`):
  - logs/2026-02-02_14-09-35-fecim.log
- Latest WRD CSV path (always newest under `logs/`):
  - logs/hysteresis-fecim-hzo-2026-02-02_14-09-35.csv
- WRD status:
  - Headless WRD CSV shows `wrd_phase=RESET` stuck for target 23 (no WRITE reached). Need to fix headless phase progression and target sequencing before rerun.

Next run (resume here)

- Investigate why headless WRD stays in phase 0 for target 23 (check phaseDuration vs dt, prep ramp thresholds, and `fromSaturation` gating).
- Verify that headless target steps correctly update `wrd_target_level` per step in CSV.
- Rerun: `./launch.sh --logger --verbosity debug --mode hysteresis` and refresh baseline entries above.
