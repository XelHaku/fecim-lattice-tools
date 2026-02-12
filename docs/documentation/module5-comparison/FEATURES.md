# Module 5: Comparison - Features

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## What This Module Does

- Builds architecture profiles for CPU, GPU, and FeCIM.
- Runs workload estimates and compares modeled metrics.
- Renders charts and summary advantages.

## Primary Components

- `module5-comparison/pkg/comparison/architecture.go`
- `module5-comparison/pkg/comparison/render.go`
- `module5-comparison/pkg/gui/app.go`

## Key Workflows

- Select a workload and compare architectures.
- Default GUI inputs are model inputs and may change; check current config in code (e.g., GPT-2 at 10,000 inferences/sec).
- Scale results to data-center assumptions.
- Render charts for energy, latency, and throughput.

## Extension Points

- Add new workloads or benchmark definitions.
- Extend architecture parameters with new fields.
- Add new chart styles or visual summaries.

## Known Limitations

- Values are modeled, not measured.
- Cross-architecture comparisons depend on assumptions.
- Not a replacement for benchmarking on real hardware.
