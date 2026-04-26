# Planned Validation Work

This file tracks validation work that should be visible to collaborators without overstating what is already proven.

## Current Public Baseline

- Keep `bash scripts/reproduce_validation.sh` as the single entry point for broad validation.
- Keep Module 2 KCL conservation executable and artifact-producing.
- Keep README claims tied to simulation, education, and published references.
- Avoid numeric performance or hardware claims unless a reproducible validation artifact exists.

## Next 30 Days

| Module | Work | Target Evidence |
|---|---|---|
| Module 1 hysteresis | Improve Park 2015 HZO overlay report and public plot export | RMSE table, source provenance, side-by-side P-E plot |
| Module 2 crossbar | Promote SPICE comparison from harness to report | ngspice current comparison JSON and plot, target deviation below 1% for small arrays |
| Module 2 crossbar | Add analytical-limit cases | Single-cell Ohm's law and 2x2 no-sneak fixtures |
| Module 3 MNIST | Publish inference artifact | accuracy, per-digit metrics, confusion matrix, quantization settings |
| Module 4 peripherals | Add DAC/ADC/TIA validation report | INL/DNL, quantization error, SNR, and TIA transfer sweep |
| Module 6 EDA | Expand EDA validation report | Yosys log, optional OpenLane result, DRC/LVS status when available |

## Next 90 Days

- Produce a consolidated validation report under `docs/4-research/validation/`.
- Add plots generated from validation artifacts, not hand-edited screenshots.
- Create a reproducibility table mapping every README claim to a validation command.
- Draft an arXiv-style methods paper for FeCIM Lattice Tools as an educational simulation framework.
- Submit the paper or a shorter artifact note to an appropriate EDA, architecture, or devices workshop.

## Claim Gate

Before adding or keeping a public claim, require:

1. A runnable command.
2. A deterministic seed or fixed fixture.
3. A threshold with units.
4. A generated artifact path.
5. A limitation note.
6. A test that fails if the claim regresses.

Claims that do not meet this gate should be labeled as planned, exploratory, or educational.
