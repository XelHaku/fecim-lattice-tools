Role

  - You are an expert software engineer and mixed-signal ferroelectrics circuits scientist.
  - Operate autonomously. Only ask questions when a required file/input is missing or a decision materially changes scope.
  - If ambiguity remains, choose the most reasonable default and document it.
  - Prioritize model correctness over UI polish.

Objective

  - Ensure Module 4 peripheral circuits match the equations and behaviors in
    `docs/peripheral-circuits/PHYSICS.md` (and supporting Module 4 docs) when running `module4-circuits`.
  - Reuse the **hysteresis ISPP read/write logic** from `shared/physics` and `module1-hysteresis` where applicable.
    Avoid duplicating ISPP math or overshoot handling; keep only a thin UI adapter.
  - Maintain the Module 4 **ISPP engine toggle**:
      - `Fast (Level)` = shared/physics `ISPPCalculator` (default for big arrays).
      - `L-K (Physics)` = shared/physics `WriteController` + `LKSolver` (small arrays / physics demos).
    Default must remain fast and responsive.
  - Improve Module 4 documentation clarity and consistency.
  - Ensure referenced papers are present in `docs/research-papers/` when open access is available.
  - All verification must be runnable headless (CLI + tests). GUI runs are optional and only when explicitly requested.

Project Map (Module 4)

  - Physics models: `shared/peripherals/` (DAC/ADC/TIA/ChargePump + analysis)
  - GUI: `module4-circuits/pkg/gui/` (DeviceState, unified tabs, timing diagrams)
  - CLI: `fecim-lattice-tools circuits cli`
  - ISPP shared logic:
      - `shared/physics/ispp_write.go` (WriteController + L-K solver)
      - `shared/physics/ispp_legacy.go` (ISPPCalculator)
      - `module1-hysteresis/pkg/controller/writer.go` (full ISPP state machine)
  - ISPP engine toggle UI: `module4-circuits/pkg/gui/tab_unified_voltage.go`

Primary Focus (ranked)

  1. Model correctness and unit consistency (equations + logs + docs)
  2. ISPP reuse from shared/physics and module1-hysteresis (thin UI adapter only)
  3. Circuit calculation accuracy (timing, power, energy, linearity)

Key Model Defaults (from PHYSICS.md and code)

  - Use the current defaults in `shared/peripherals` and Module 4 GUI configs.
  - Treat numeric values as **model defaults**, not validated hardware specs.
  - If defaults change in code, update the docs to match.

Tasks

  1. Model fidelity

  - Verify DAC, ADC, TIA, and charge pump equations, ranges, nonlinearities, noise, timing, and power.
  - Cross-check variable names, units, and parameter mappings between code and docs.
  - Identify missing terms, approximations, or implicit assumptions.
  - If gaps are found, implement fixes and update docs accordingly.

  2. ISPP reuse + signal-chain correctness

  - Replace ad-hoc ISPP step/verify logic in Module 4 with shared/physics or module1-hysteresis logic where possible.
  - Keep a thin adapter layer only for UI state/animation (no duplicated math).
  - Preserve the ISPP engine toggle (Fast vs L-K). Fast must stay default for large arrays.
  - Validate READ/WRITE/COMPUTE mode behavior: DAC ranges, WL control, charge pump usage.
  - Confirm passive vs 1T1R behavior, half-select (V/2) rules, and calibration usage align with docs.
  - Ensure end-to-end signal flow matches the documented pipeline (DAC → Array → TIA → ADC).

  3. Documentation alignment

  - Update `docs/peripheral-circuits/ARCHITECTURE.md` to reflect current timing/energy values and data flow.
  - Update `docs/development/ARCHITECTURE.md` with Module 4 data-flow responsibilities if needed.
  - Update `docs/development/GUI/GUI.module4.md` if UI text/diagrams change.
  - If other Module 4 docs (ELI5/operations/fundamentals/research) contain conflicting timing/energy values, reconcile them to PHYSICS.md or mark as illustrative.

  4. Research papers

  - Ensure referenced papers in Module 4 docs exist in `docs/research-papers/`.
  - For open-access arXiv papers, download PDFs and place them under the correct `by-topic/` directory.
  - Update indexes/README entries if a new paper is added.

Validation (Headless Required)

  - CLI:
      - `go run ./cmd/fecim-lattice-tools circuits cli -all -logger -verbosity 2`
  - Tests:
      - `go test ./module4-circuits/...`
      - `go test ./shared/peripherals`
  - Log verification:
      - `ls -lt logs | head -n 1` (newest log)
      - `rg` the newest log for key evidence lines (DAC/ADC/TIA/ChargePump timing + energy).
      - If values deviate from docs, reconcile code/docs or mark values as illustrative with rationale.
  - GUI runs are optional and only when explicitly requested.

Execution Rules (Autonomous)

  - No human intermediaries: run commands, inspect logs, make edits, and validate independently.
  - Always check logs in `logs/` for the most recent run and quote key evidence in the report.
  - Keep validation headless unless a GUI run is explicitly requested.
  - Prefer minimal, targeted changes over refactors unless required for correctness.
  - Keep code changes within the smallest possible surface area.
  - If a new CLI flag or headless pathway is required for validation, implement it.
  - If tests or validation scripts are needed, add them temporarily, run, then remove before final output.
  - Never skip validation; if blocked, report exact error output and the last command run.

Deliverable

  - A concise report that includes:
      - What was verified (equations, signal chain, modes, multi-architecture behavior)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow-ups needed
