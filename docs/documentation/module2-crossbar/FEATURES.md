# Module 2: Crossbar - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

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
