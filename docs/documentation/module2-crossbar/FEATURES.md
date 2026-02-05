# Module 2: Crossbar - Features

## What This Module Does

- Simulates crossbar array MVM/VMM with discrete conductance levels.
- Provides IR drop, sneak-path, and drift analyses.
- Visualizes conductance heatmaps and live MVM steps.

## Primary Components

- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/gui/tabs/irdrop_tab.go`
- `module2-crossbar/pkg/gui/tabs/sneak_tab.go`

## Key Workflows

- Program weights, run MVM, compare ideal vs non-ideal.
- Sweep wire parameters to study IR-drop sensitivity.
- Visualize drift over time and quantify error.

## Extension Points

- Add new non-ideality models in `pkg/crossbar`.
- Extend GUI tabs for additional analyses.
- Enable GPU acceleration via config (`UseGPU`).

## Known Limitations

- No transistor-level or SPICE transient simulation.
- Wire models are simplified and not process-specific.
- Default parameters are for teaching, not calibration.
- Drift and variation coefficients are assumed unless cited (DOI: (add)).
