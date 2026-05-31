# Verox Feasibility Assessment

## Context and interpretation

The term **"Verox"** is not currently defined in this repository or tool inventory. In context, it most likely refers to **Verilog verification flow tooling** (lint/sim/formal-adjacent checks).

For FeCIM Lattice Tools, the practical interpretation is:

- verify generated Verilog artifacts from Module 6
- run syntax/lint/smoke simulation checks with open-source tools

## Proposed target tool/package

Primary candidate stack (already aligned with tool inventory):

1. **Icarus Verilog (`iverilog`)** for compile/sim smoke tests
2. **Verilator** for lint/static elaboration and fast simulation checks

If "Verox" later resolves to a specific package, map it to this same interface contract.

## IO format and contract

### Inputs

- Generated Verilog from `module6-eda` export pipeline
- Optional testbench stubs and vector files
- Optional synthesis constraints metadata (non-blocking)

### Outputs

- pass/fail status
- compiler/lint logs
- optional waveform (`.vcd`) for smoke vectors
- summary JSON or markdown record for CI artifacts

## Validation scope (recommended)

- Syntax validity of generated Verilog
- Module/interface consistency (ports, widths, parameter wiring)
- Basic behavioral sanity on representative vectors
- Non-goals (for now): timing signoff, analog equivalence, PVT closure

## Feasibility conclusion

**Feasible with current infrastructure** for syntax/lint/smoke simulation scope.

Rationale:

- Tool pins already exist (`iverilog`, `verilator`)
- CI can treat these as optional tools with explicit skip messaging
- Exported Verilog path already exists in module6

Constraints:

- Full signoff-equivalent verification is **not** feasible without broader ASIC flow setup and calibrated hardware references.
- Any claim beyond syntax/behavior smoke must remain explicitly scoped as non-signoff.
