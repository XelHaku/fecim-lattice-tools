# Module 1: Hysteresis - Features

## What This Module Does

- Simulates ferroelectric hysteresis loops (Preisach) with a selectable dynamic engine (Landau-Khalatnikov).
- Exposes discrete polarization levels for memory-state modeling.
- Provides GUI visualization, calibration, and write/read/verify flows.

## Primary Components

- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `module1-hysteresis/pkg/gui/physics_engine.go`
- `module1-hysteresis/pkg/gui/simulation.go`
- `module1-hysteresis/pkg/render/plot.go`

## Key Workflows

- Sweep electric field to generate a P-E loop.
- Program and read discrete levels with ISPP-style verification.
- Calibrate per temperature and interpolate between cached calibrations.

## Extension Points

- Add new material parameter sets in `shared/physics/material.go`.
- Extend hysteron distributions in `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`.
- Add alternative renderers or plotting styles.

## Known Limitations

- Landau-Khalatnikov engine is educational, not device-calibrated.
- Wake-up/fatigue indicators are UI-level placeholders.
- GPU renderer is specialized and may not cover all platforms.
