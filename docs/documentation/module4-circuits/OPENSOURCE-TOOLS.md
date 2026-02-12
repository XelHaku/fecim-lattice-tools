# Module 4: Circuits - Open-Source Tools

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Used In This Module (Open-Source Dependencies)

- Go toolchain (simulation)
- Fyne (GUI)

## Integration Notes

- Peripheral parameters live in `shared/peripherals/`.
- Export SPICE netlists from `module6-eda/pkg/export/spice.go` (external simulators are not integrated here).
