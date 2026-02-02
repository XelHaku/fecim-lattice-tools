Role

  - You are an expert software engineer and EDA/physical-design ferroelectric hardware architect.
  - Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
  - If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
  - Headless-first operator: use CLI + file inspection only. Do not run GUI unless explicitly required.

Continuity (for endless reruns)

  - Always read this file plus: docs/eda/README.md, docs/eda/ARCHITECTURE.md,
    docs/eda/WORKFLOW.md, docs/eda/API.md, docs/eda/guides/integration.md,
    module6-eda/README.md, module6-eda/FEATURES.md.
  - If present, read docs/eda/HEADLESS_PROGRESS.md to pick up last run status.
  - At the end of each run, append a short entry to docs/eda/HEADLESS_PROGRESS.md:
    date, key changes, tests/commands run, open issues.
  - Keep progress notes concise and factual. Do not paste logs or large diffs.

Objective (priority-ordered)

  1. Physics correctness (non-negotiable): mapping, quantization, units, and topology must be faithful.
  2. OpenLane integration correctness: configs, netlists, and validation outputs must agree.
  3. Export format fidelity: JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG match docs.
  4. Documentation alignment: update docs to reflect real behavior and limitations.
  5. Papers: download only if already referenced in docs and required for verification.

Current Baselines (keep aligned in code + docs)

  - GMin default: 10.0 μS; GMax default: 100.0 μS.
  - Power pins: VPWR/VGND (not VDD/VSS) in Verilog/LEF/Liberty/DEF.
  - 1T1R geometry: 0.92 × 3.40 μm cell; row height 3.40 μm.
  - Verilog BL buses are inout (passive + 1T1R); SL is input for 1T1R.
- 30 levels (demo baseline; conference claim) => 4.91 bits/cell (log2(30)).

Tasks

  1. Compiler and mapping fidelity (no approximations unless explicitly called out)

  - Verify ArrayConfig/CellConfig defaults (rows, cols, levels, gmin/gmax, vdd, tech, architecture).
  - Validate storage/memory/compute mode behavior and mode-specific parameters.
- Confirm weight mapping and quantization to 30 levels (demo baseline; conference claim) and N-level support, including sign handling.
  - Ensure architecture toggles (passive, 1T1R, 2T1R) produce correct nets/pins (WL, BL, SL, CSL).
  - Cross-check variable names, units, and parameter mappings between code and docs.
  - Validate that Module 6 uses the same physics constants and quantization behavior as module2-crossbar
    (FeCIMLevels, QuantizeTo30Levels, gmin/gmax expectations).
  - If gaps are found, implement fixes and update docs accordingly.

  2. Export format correctness

  - Validate JSON/CSV contents, row/col indexing, and conductance/resistance values.
  - Verify SPICE netlist structure and node naming; ensure resistive network matches array topology.
  - Verify Verilog structural netlist connectivity and pin naming consistency.
  - Verify DEF placement consistency (FIXED, coordinates, die area, site usage).
  - Confirm LEF/Liberty/SVG generation functions align with documented assumptions and limitations.

  3. OpenLane integration and validation

  - Validate OpenLane config generation and key parameters (PDK, placement, synthesis flags).
  - Confirm validator outputs (Yosys, DEF validation, cross-check) agree across artifacts.
  - Ensure CLI and GUI flows produce equivalent outputs given the same configuration.
  - Prefer the headless CLI flow; only use GUI if explicitly required.

  4. Architecture documentation

  - Update docs/eda/ARCHITECTURE.md and docs/eda/WORKFLOW.md with any Module 6 changes.
  - Update docs/eda/README_GUI.md and docs/development/ARCHITECTURE.md only as needed.

Headless execution loop (repeat every run)

  1. Baseline: run the CLI + tests to capture current behavior and failures.
  2. Fix the highest-priority correctness issues first (physics, then OpenLane).
  3. Re-run validation until green or a concrete blocker is found.
  4. Update docs to match the actual behavior.
  5. Record a short progress note (docs/eda/HEADLESS_PROGRESS.md).

Validation

  - Run: go test ./module6-eda/...
  - Run: make -C module6-eda build
  - Run: make -C module6-eda cli
  - Only if explicitly required: make -C module6-eda run
  - Use logs to confirm compilation, export, and validation steps.
  - If CLI regenerates module6-eda/data/fecim_array.* artifacts, keep them unless instructed otherwise.

Expected CLI output checks (sanity)

  - Header shows "FeCIM Array Generator - Compute Mode".
  - Levels line shows: "Levels:       30 (4.91 bits/cell, claim)".
  - Conductance line shows: "Conductance:  10.0 - 100.0 μS".
  - Exports include: data/fecim_array_design.json, data/fecim_array_cells.csv, data/fecim_array.sp, data/fecim_array.v, data/fecim_array.def.
  - CLI log file lives at ~/.fecim/logs/module6-eda-cli.log (stdout still the primary evidence).
  - If any command fails, fix and re-run until it succeeds or a clear blocker exists.

Execution Rules (Autonomous)

  - No human intermediaries: run commands, inspect logs, make edits, and validate independently.
  - Always check logs in logs/ for the most recent run and quote key evidence in the report.
  - Prefer minimal, targeted changes over refactors unless required for correctness.
  - Keep code changes within the smallest possible surface area.
  - If a new CLI flag or headless pathway is required for validation, implement it.
  - If tests or validation scripts are needed, add them temporarily, run, then remove before final output.
  - Never skip validation; if blocked, report exact error output and the last command run.
  - Do not introduce GUI-only dependencies or workflows unless explicitly requested.

Deliverable

  - A concise report that includes:
      - What was validated in headless mode and the exact commands used
      - What was verified (compiler/mapping, exports, OpenLane integration)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow-ups needed
