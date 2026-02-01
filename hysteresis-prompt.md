Role

  - You are an expert software engineer and ferroelectrics scientist. Work autonomously; no questions unless blocked.

  Objective

  - Ensure the hysteresis module fully implements the complete equation in docs/hysteresis/hysteresis-gemini.md when
    running in hysteresis mode.

  Tasks

  1. Equation fidelity

  - Verify every term, unit, parameter mapping, and sign convention is implemented.
  - Cross‑check variable names and units between code and the doc.
  - Identify any missing terms, approximations, or implicit assumptions.

  2. Architecture documentation

  - Update docs/developmnet to reflect the new architecture: modules, data flow, responsibilities, and key
    interfaces.

  3. ISPP documentation

  - Document the ISPP method used in the read/write demo, including:
      - Step sequencing and termination criteria
      - Parameter choices and their physical meaning
      - Any constraints or limits applied

  4. Multi‑step ISPP support

  - Confirm the implementation supports multiple ISPP steps end‑to‑end.

  Validation

  - Run: ./launch.sh --logger --verbosity debug --mode hysteresis
  - Use logs to confirm equation terms are exercised and ISPP runs across multiple steps.

  Deliverable

  - A concise report that includes:
      - What was verified (equation, ISPP, multi‑step behavior)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow‑ups needed

