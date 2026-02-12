# Module 1: Hysteresis - Open-Source Tools

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Used In This Module (Open-Source Dependencies)

- Go toolchain (build/runtime for the simulator)
- Fyne (GUI rendering)
- Bubble Tea (TUI mode)
- Vulkan-go (optional GPU renderer)

## Integration Notes

- Material presets live in `module1-hysteresis/pkg/ferroelectric/material.go` and are simulation defaults.
- Exported data should keep units consistent with `PHYSICS.md`.
- External solvers or data science stacks are not integrated in this repo.
