Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to **headless-first work** unless a GUI change is required for correctness of WRD/ISPP or engine parity.

Objectives

- Explain **why the Landau-Khalatnikov (L-K / "Landau Kalinokov") math is too slow**: quantify where time is spent and which terms/loops dominate.
- Preserve **Frankestein equation** fidelity in `docs/hysteresis/hysteresis-gemini.md` while investigating performance
  (terms, signs, units, and effective viscosity).
- Keep **both physics engines** (L-K and Preisach) supported and coherent. Do not remove Preisach.
- Avoid regressions: WRD/ISPP convergence and autonomous calibration must remain stable.

Non‑Negotiables

- **Both modes must remain**: L-K (dynamic) and Preisach (quasi-static).
- GUI and headless must use the same **L-K solver equation** when L-K is selected.
- GUI Preisach path remains available and stable for fast visualization.
- No GUI updates from goroutines without `fyne.Do(func(){...})`.

Primary Focus (ranked)

1) L-K performance diagnosis (why it's slow)
- Quantify **step counts**, adaptive sub-steps, and `dt` shrink near Ec.
- Identify hot paths: polynomial term evals, allocation churn, logging overhead, GUI sync, and render coupling.
- Separate **math cost** from **framework cost** (GUI vs headless). Preisach is the fast baseline.

2) Frankestein equation fidelity
- Implement exactly the unified L-K + depolarization + series-resistance formulation
  from `docs/hysteresis/hysteresis-gemini.md`.
- Verify all terms, signs, and units in logs.

3) WRD/ISPP convergence and stability (no regressions)
- Strict target hit (exact level match).
- No “stuck at one level” loops.
- Overshoot recovery only when overshoot truly happens.
- Bounds/bisection never collapse or invert.

4) Autonomous calibration (no regressions)
- Trigger only when convergence is poor and run **between targets**.
- Persist per‑engine calibrations (L‑K file suffix `-lk`).

5) Documentation sync
- Keep `docs/hysteresis/hysteresis.demo.md` aligned with WRD/ISPP behavior.
- Update `docs/hysteresis/hysteresis-gemini.md` only if equation handling changes.

Implementation Notes

- **Physics engines**:
  - L-K (dynamic): adaptive sub‑stepping near Ec.
  - Preisach (quasi-static): single step per frame for performance.
- **Logging**:
  - Per‑step physics logs only at `--verbosity trace`.
  - For performance runs, prefer **aggregated counters** (per frame or per second) over per‑step logs.
  - CSV logging is throttled by simulation time.
    - Default: 250ms (approx 4 rows/sec).
    - Override: `FECIM_HYSTERESIS_LOG_INTERVAL_MS` (float, ms).
- **Calibration files**:
  - Preisach: `data/calibrations/<material>.json`
  - L-K: `data/calibrations/<material>-lk.json`

Tasks

1) L-K performance accounting (why it's slow)
- Add/verify counters for **sub-steps per frame**, `dt` min/mean/max, and **time spent in L-K solver** vs render.
- Record whether adaptive stepping is dominating (e.g., long runs of tiny `dt` near Ec).
- Confirm whether overhead is math-bound (`pow`/polynomial evals) or glue-bound (allocations/logging/GUI sync).

2) Frankestein equation (no missing terms)
- Verify: `dP/dt = (E_applied - k_dep*P - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Ensure `rho_eff = rho + (R_series * A / d)` only when `UseEffectiveViscosity=true`.
- Confirm `E_eff = E_applied - k_dep*P` is what the solver actually uses.
- Log (trace only): `E_applied`, `E_dep`, `E_eff`, `dG_dP`, `rho_eff`, `Alpha`, `Beta`, `Gamma`, `K_dep`.

3) ISPP convergence (no regressions)
- If `currentLevel == targetLevel`, exit immediately with success.
- Maintain valid bounds: `VMin < VMax`.
- Reset bounds after overshoot or invalid bracket.
- Add stuck detection: if level does not change for several verifies, boost step size and reset bounds.
- Never saturate unless overshoot recovery requires it.

4) Autonomous recalibration (no regressions)
- Trigger on repeated overshoots or excessive pulses.
- Run between targets to avoid corrupting active state.
- Persist and sync into `CalibrationManager`.

5) Documentation updates
- Update docs only when behavior or equations change.

Validation

- **Headless (always first):**
  - `FECIM_HYSTERESIS_LOG_INTERVAL_MS=250 ./launch.sh --logger --verbosity debug --mode hysteresis`
  - Use `--verbosity trace` only when validating L-K equation terms.
- **Performance evidence:**
  - Capture solver time, sub-step counts, and `dt` stats from the aggregated counters.
  - If pprof is enabled, collect a short CPU profile; otherwise rely on counters.
- **GUI WRD:**
  - `FECIM_HYSTERESIS_LOG_INTERVAL_MS=250 ./launch.sh --logger --verbosity debug`
  - Confirm “TARGET HIT” lines appear in WRD logs.

Frankestein Equation Checklist

- Uses: `E_eff = E_applied - k_dep*P`.
- Uses: `dP/dt = (E_eff - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Uses: `rho_eff = rho + (R_series * A / d)` only if enabled.
- Per‑step logging only when verbosity is `trace`.

WRD/ISPP Correctness Checklist

- Target hit with strict equality for each WRD cycle.
- If `current == target`, success without pulses.
- No convergence to `E~0` due to wrong direction inference.
- Overshoot reset only on true overshoot.
- Bounds never invert; bisection uses valid bracket.
- Stuck detection escalates and breaks deadlocks.

Execution Rules (Autonomous)

- Always run **headless** (`--mode hysteresis`) before GUI checks.
- If the newest headless log is stale, generate a fresh one.
- Inspect newest WRD log when GUI WRD is exercised.
- Prefer minimal, targeted changes; avoid unrelated files.
- If validation fails, report exact error output and last command run.

Deliverable

- Concise report:
  - Why L-K math is slow (dominant hot path + evidence).
  - Frankestein equation verification (what terms/logs confirmed).
  - WRD/ISPP target-hit evidence (log lines).
  - Documentation updates (file paths + summary).
  - Gaps/issues and next iteration target.
- Include validation command and log path.

Baseline (update each run)

- Latest headless log path:
  - logs/2026-02-02_16-58-26-fecim.log (stale; rerun headless)
- Latest WRD log path:
  - logs/2026-02-02_17-49-53-fecim.log
- Latest WRD CSV path:
  - logs/hysteresis-hzo-si-doped-2026-02-02_17-50-01.csv

Recent Changes (2026-02-02)

1) Dual physics engines in GUI: L-K (dynamic) and Preisach (quasi-static) toggle.
2) L-K solver wired into GUI simulation path; Preisach preserved as mode.
3) Per-engine calibration files: L-K uses `-lk` suffix.
4) CSV logging throttled (default 250ms) with env var override.
5) Per-step physics logs moved to `trace` verbosity.
6) Preisach GUI loop uses single step per frame for performance.
7) ISPP now has bounds + bisection and stuck detection.
