# 100% Coverage Boundary for FeCIM Validation

This document defines what can be fully validated internally and what requires external evidence.

## A) What CAN be fully tested internally

These areas can reach effectively complete software-test coverage within this repository:

- Physics equation implementation correctness (unit tests and regression checks)
- Deterministic unit/integration tests across modules
- Golden-file regressions for modeled outputs
- ISPP controller convergence behavior under modeled conditions
- Sense chain software path consistency (DAC -> array model -> TIA -> ADC)
- Export format correctness (JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG structure)
- CI reproducibility for pinned software/tool versions when tools are present

These support strong claims about **code correctness relative to implemented models**.

## B) What REQUIRES external data/tools

The following cannot be claimed as fully validated from internal tests alone:

- Fabricated-device measurements (real Pr/Ec variability, endurance, retention)
- Comparator agreement against independent simulators (Heracles, CrossSim, ngspice)
- Real analog non-idealities and process spread beyond modeled assumptions
- Device characterization under true environmental/stress conditions
- Signoff-grade circuit behavior across full PVT without external flow/tool evidence

These require either external simulators, measured datasets, or both.

## C) Scientific-claim boundary

Use the following boundary language in reports/docs:

- **Allowed internal claim**: "Model implementation passes internal regression and consistency tests."
- **Requires external claim support**: "Model is quantitatively accurate to physical hardware/tool references."

If no external evidence is attached, claims must be framed as:

- modeled/simulated behavior,
- educational/engineering approximation,
- not a fabrication-validated result.
