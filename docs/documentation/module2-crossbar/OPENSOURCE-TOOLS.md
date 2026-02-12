# Module 2: Crossbar - Open-Source Tools

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Used In This Module (Open-Source Dependencies)

- CrossSim (optional external install check; algorithms referenced in tests)
- BadCrossbar (optional external install check)

## Integration Notes

- Crossbar parameters live in `module2-crossbar/pkg/crossbar/array.go`.
- Non-idealities are implemented in `module2-crossbar/pkg/crossbar/nonidealities.go`.
- External circuit simulators and layout tools are not integrated in this repo.
