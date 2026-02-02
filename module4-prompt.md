Role

  - You are an expert software engineer and mixed-signal ferroelectrics circuits scientist.
  - Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
  - If an ambiguity remains, choose the most reasonable default and proceed; document the choice.

Objective

  - Ensure the Module 4 peripheral circuits implementation fully matches the equations and behaviors in
    docs/peripheral-circuits/PHYSICS.md (and supporting module4 docs) when running module4-circuits.
  - Make any required code + documentation updates to achieve fidelity and verify via CLI output and logs.
  - Improve Module 4 documentation quality and ensure referenced papers are downloaded into the repo's
    research-papers area when possible.

Tasks

  1. Physics fidelity (no approximations unless explicitly called out)

  - Verify DAC, ADC, TIA, and charge pump equations, ranges, nonlinearities, noise, timing, and power.
  - Cross-check variable names, units, and parameter mappings between code and docs.
  - Identify missing terms, approximations, or implicit assumptions.
  - If gaps are found, implement fixes and update docs accordingly.

  2. Signal-chain and mode correctness

  - Validate READ/WRITE/COMPUTE mode behavior: DAC ranges, word-line control, and charge pump usage.
  - Confirm passive vs 1T1R behavior, half-select (V/2) rules, and calibration usage align with docs.
  - Ensure end-to-end signal flow matches the documented pipeline (DAC -> Array -> TIA -> ADC).

  3. Architecture documentation

  - Update docs/peripheral-circuits/ARCHITECTURE.md and docs/development/ARCHITECTURE.md to reflect
    Module 4 data flow, responsibilities, and interfaces.
  - Update docs/development/GUI/GUI.module4.md with diagrams and related visual explanations as needed.

  4. Analysis outputs and multi-architecture support

  - Ensure CLI analysis outputs (linearity, timing, power) match documented formulas and defaults.
  - Confirm multiple architectures (passive, 1T1R) and multi-row compute behavior work end-to-end.
  - If missing, implement a minimal end-to-end path and validate.

Validation

  - Run: go run ./module4-circuits/cmd/circuits -all
  - Run: go test ./module4-circuits/...
  - If GUI verification is required: ./launch.sh --logger --verbosity debug --module circuits
  - Use logs to confirm signal-chain and mode behavior.
  - If any command fails, fix and re-run until it succeeds or a clear blocker exists.

Execution Rules (Autonomous)

  - No human intermediaries: run commands, inspect logs, make edits, and validate independently.
  - Always check logs in logs/ for the most recent run and quote key evidence in the report.
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
