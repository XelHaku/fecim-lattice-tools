Role

  - You are an expert software engineer and ferroelectrics scientist.
  - Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
  - If an ambiguity remains, choose the most reasonable default and proceed; document the choice.

Objective

  - Ensure the hysteresis module fully implements the complete equation in docs/hysteresis/hysteresis-gemini.md when
    running in hysteresis mode.
  - Make any required code + documentation updates to achieve fidelity and verify via logs.
  - Improve hysteresis documentation quality and ensure referenced papers (e.g., those cited in hysteresis-gemini.md)
    are downloaded into the repo’s research-papers area when possible.

Tasks

  1. Equation fidelity (no approximations unless explicitly called out)

  - Verify every term, unit, parameter mapping, and sign convention is implemented.
  - Cross‑check variable names and units between code and the doc.
  - Identify any missing terms, approximations, or implicit assumptions.
  - If gaps are found, implement fixes and update docs accordingly.

  2. Architecture documentation

  - Update docs/development to reflect the new architecture: modules, data flow, responsibilities, and key
    interfaces.

  3. ISPP documentation

  - Document the ISPP method used in the read/write demo, including:
      - Step sequencing and termination criteria
      - Parameter choices and their physical meaning
      - Any constraints or limits applied

  4. Multi‑step ISPP support

  - Confirm the implementation supports multiple ISPP steps end‑to‑end.
  - If it does not, implement a minimal end‑to‑end multi‑step path and validate.

Validation

  - Run: ./launch.sh --logger --verbosity debug --mode hysteresis
  - Use logs to confirm equation terms are exercised and ISPP runs across multiple steps.
  - If the command fails, fix and re‑run until it succeeds or a clear blocker exists.

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
      - What was verified (equation, ISPP, multi‑step behavior)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow‑ups needed
