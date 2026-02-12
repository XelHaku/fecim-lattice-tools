# Module 1: Hysteresis - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## What This Module Does

- Simulates ferroelectric hysteresis loops (Preisach) with an optional, simplified Landau-Khalatnikov engine.
- Exposes discrete polarization levels for memory-state modeling (simulation baseline).
- Provides GUI visualization plus simulated calibration and write/read/verify flows.

## Primary Components

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/gui/physics_engine.go`
- `module1-hysteresis/pkg/gui/simulation.go`
- `module1-hysteresis/pkg/render/plot.go`

See also: [RUN_MODES.md](./RUN_MODES.md) for GUI/TUI/headless/Vulkan behavior and Preisach vs L-K defaults.

## Key Workflows

- Sweep electric field to generate a P-E loop.
- Program and read discrete levels with ISPP-style verification (simulated).
- Calibrate per temperature and interpolate between cached calibrations (heuristic).

## Extension Points

- Add new material parameter sets in `shared/physics/material.go`.
- Extend hysteron distributions in `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`.
- Add alternative renderers or plotting styles.

## Known Limitations

- Landau-Khalatnikov engine is educational, not device-calibrated.
- Wake-up/fatigue indicators are UI-level placeholders.
- GPU renderer is specialized and may not cover all platforms.
- Material presets include illustrative or unverified values; cite before external use (DOI: (add)).
