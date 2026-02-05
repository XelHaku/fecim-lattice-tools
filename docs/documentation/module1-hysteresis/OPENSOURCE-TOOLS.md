# Module 1: Hysteresis - Open-Source Tools

## Used In This Module (Open-Source Dependencies)

- Go toolchain (build/runtime for the simulator)
- Fyne (GUI rendering)
- Bubble Tea (TUI mode)
- Vulkan-go (optional GPU renderer)

## Integration Notes

- Material presets live in `module1-hysteresis/pkg/ferroelectric/material.go` and are simulation defaults.
- Exported data should keep units consistent with `PHYSICS.md`.
- External solvers or data science stacks are not integrated in this repo.
