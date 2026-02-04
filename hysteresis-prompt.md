Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives (including tests, goldens, and docs **only when behavior changes**).
- Default to **headless-first work** unless a GUI change is required for correctness of WRD/ISPP or engine parity.
- **Always read `TODO.md` first** to align with current priorities and status (and update it if priorities shift).
- **Manage everything without user input**: tests, test data/golden updates, GUI synchronization, and docs updates as needed.
- **Reference phase‑field examples** in `opensource/hysteresys` (FerroX, ferret) when aligning Landau/LK notes or UI explanations.

Objectives

- Explain **why the Landau-Khalatnikov (L-K)** math is too slow (legacy typo: “Landau Kalinokov” appears in notes): quantify where time is spent and which terms/loops dominate.
- Preserve **Frankenstein equation** fidelity (legacy typo: “Frankestein” in some notes) in `docs/hysteresis/hysteresis-gemini.md` while investigating performance
  (terms, signs, units, and effective viscosity).
- Keep **both physics engines** (L-K and Preisach) supported and coherent. Do not remove Preisach.
- Avoid regressions: WRD/ISPP convergence and autonomous calibration must remain stable.
- **Run headless WRD/ISPP in Preisach mode** for parity with GUI and to debug target/marker mismatches.
- Find and fix the **yellow target mismatch bug** (GUI) with evidence from logs and headless parity.
- Keep the **physics equations UI** accurate for hysteresis ISPP, Preisach, and Landau (labels + links).

Non‑Negotiables

- **Both modes must remain**: L-K (dynamic) and Preisach (quasi-static).
- GUI and headless must use the same **L-K solver equation** when L-K is selected.
- GUI Preisach path remains available and stable for fast visualization.
- Headless must be able to run **Preisach** (default for WRD/ISPP validation).
- No GUI updates from goroutines without `fyne.Do(func(){...})`.
- If tests/goldens must change, update **test data**, **golden version metadata**, and **TODO.md** to reflect the new baseline.

Primary Focus (ranked)

1) L-K performance diagnosis (why it's slow)
- Quantify **step counts**, adaptive sub-steps, and `dt` shrink near Ec.
- Identify hot paths: polynomial term evals, allocation churn, logging overhead, GUI sync, and render coupling.
- Separate **math cost** from **framework cost** (GUI vs headless). Preisach is the fast baseline.

2) Frankenstein equation fidelity
- Implement exactly the unified L-K + depolarization + series-resistance formulation
  from `docs/hysteresis/hysteresis-gemini.md`.
- Verify all terms, signs, and units in logs.

3) WRD/ISPP convergence and stability (no regressions)
- Strict target hit (exact level match).
- No “stuck at one level” loops.
- Overshoot recovery only when overshoot truly happens.
- Bounds/bisection never collapse or invert.
- **Target/marker parity:** GUI yellow target must track the **active** target (not pre-selected next target).
  - If polarization relaxes at E=0, do **not** update the target; target highlight must reflect controller target.

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
  - **Headless WRD/ISPP**: run Preisach to match GUI behavior and isolate controller issues.
  - If a CLI engine flag is missing, **inspect the headless entry point** and wire it up.
- **Physics equations UI** (keep labels/links coherent):
  - `module1-hysteresis/pkg/gui/widgets/physics_equations.go`
  - `module1-hysteresis/pkg/gui/widgets/physics_equations_info.go`
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

2) Frankenstein equation (no missing terms)
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
- Detect **no‑improvement** (error not decreasing) and break deadlocks (boost step + clear bounds).

4) Autonomous recalibration (no regressions)
- Trigger on repeated overshoots or excessive pulses.
- Run between targets to avoid corrupting active state.
- Persist and sync into `CalibrationManager`.

5) Target/marker parity (bug hunt)
- Ensure GUI yellow target reflects the **current** target only.
- Next target must be **queued** and only applied at PREP start.
- If mismatch persists, instrument **phase-boundary logs** to compare:
  - `wrdTargetLevel`
  - `writeController.TargetLevel`
  - `wrdNextTargetLevel`
  - `wrdPhase`
  - `discreteLevel`
  - E/P at the same time
  - (GUI) target label text
- Log only on phase transitions or once per second; avoid per-frame spam.
- If controller target is correct but GUI target is wrong, check UI refresh uses the **same snapshot** as physics step.

6) Documentation updates
- Update docs only when behavior or equations change.
- Keep `TODO.md` and the physics equations UI scope in sync with current tasks.

Validation

- **Headless (always first):**
  - Preisach WRD/ISPP: `FECIM_HYSTERESIS_LOG_INTERVAL_MS=250 ./launch.sh --logger --verbosity debug --mode hysteresis --engine preisach`
  - L-K performance: `FECIM_HYSTERESIS_LOG_INTERVAL_MS=250 ./launch.sh --logger --verbosity debug --mode hysteresis --engine lk`
  - Use `--verbosity trace` only when validating L-K equation terms.
- If `--engine` is not supported, **derive engine selection from code** (config/env/flag) and document it.
- **Performance evidence:**
  - Capture solver time, sub-step counts, and `dt` stats from the aggregated counters.
  - If pprof is enabled, collect a short CPU profile; otherwise rely on counters.
- **GUI WRD:**
  - `FECIM_HYSTERESIS_LOG_INTERVAL_MS=250 ./launch.sh --logger --verbosity debug`
  - Confirm “TARGET HIT” lines appear in WRD logs.
  - Verify GUI target highlight matches `writeController.TargetLevel` through WRITE→DISPLAY.

Frankenstein Equation Checklist

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
- No‑improvement detection breaks deadlocks.
- GUI yellow target == active controller target (no early jump).
- GUI target label text == controller target at PREP/WRITE/VERIFY/DISPLAY boundaries.

Execution Rules (Autonomous)

- Always run **headless** (`--mode hysteresis`) before GUI checks.
- If the newest headless log is stale, generate a fresh one.
- Inspect newest WRD log when GUI WRD is exercised.
- Prefer minimal, targeted changes; avoid unrelated files.
- If validation fails, report exact error output and last command run.
- If GUI behavior changes, update UI labels, physics-equations tabs, and any related golden/test data.

Deliverable

- Concise report:
  - Why L-K math is slow (dominant hot path + evidence).
  - Frankenstein equation verification (what terms/logs confirmed).
  - WRD/ISPP target-hit evidence (log lines).
  - Target/marker parity evidence (log lines showing target alignment).
  - Documentation updates (file paths + summary).
  - Gaps/issues and next iteration target.
- Include validation command and log path.

Baseline (update each run)

- Latest headless log path (refresh when a new headless run is performed):
  - logs/2026-02-03_17-56-17-fecim.log
- Latest headless CSV path (refresh when a new headless CSV is generated):
  - logs/hysteresis-literature-superlattice-2026-02-03_17-56-19.csv
- Note: in this environment, prefer `grep` over `rg` if ripgrep is not installed.

Recent Changes (2026-02-02)

1) Dual physics engines in GUI: L-K (dynamic) and Preisach (quasi-static) toggle.
2) L-K solver wired into GUI simulation path; Preisach preserved as mode.
3) Per-engine calibration files: L-K uses `-lk` suffix.
4) CSV logging throttled (default 250ms) with env var override.
5) Per-step physics logs moved to `trace` verbosity.
6) Preisach GUI loop uses single step per frame for performance.
7) ISPP now has bounds + bisection and stuck detection.
8) WRD target is queued and only applied at PREP start (prevents early target jump).
9) ISPP now detects no‑improvement across verifies and breaks deadlocks.
