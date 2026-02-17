# Module 5: Comparison - Physics

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, navigation behavior, and code paths referenced in this page are implemented in this repo and verifiable from source/tests.
- **Modeled:** Equations, defaults, and performance/quality estimates are simulator or documentation models unless explicitly tied to cited measured data.
- **Aspirational:** Any production-scale, silicon-parity, or ecosystem-wide claims are roadmap intent and must not be reported as demonstrated results.

## Prerequisites

- Basic performance metrics
- Power and energy concepts
- Log-scale charts and ratios

## Core Model

- Each architecture is described by compute, memory, and energy parameters.
- Workloads estimate latency, throughput, and energy.
- Comparisons focus on relative differences, not absolute claims.

## Key Equations (Simplified)

```
Energy = Power * Time
Throughput = Ops / Time
Efficiency = Ops / Energy
```

## ROI Calculations (Model Inputs)

All ROI values in Module 5 are **model inputs / model outputs** for visualization only.

GUI calculator (energy per inference in µJ):

```
E_inf(µJ) = MACs * Energy_fJ / 1e9
P(W) = E_inf(µJ) * InferencesPerSec / 1e6
Cost_month($) = (P(W) / 1000) * HoursPerMonth * $/kWh
Savings_year($) = (GPU_cost - FeCIM_cost) * 12 * ServerScale
```

Comparison engine (architecture.go):

```
Energy_mJ = TDP_W * Latency_ms
Energy_kWh = Energy_mJ / 3.6e9
Cost_per_inference($) = Energy_kWh * $/kWh
```

## Parameters And Units

| Symbol | Meaning | Units |
|---|---|---|
| P | Power | Watts |
| t | Time | seconds |
| E | Energy | Joules |
| T | Throughput | ops/s |
| MACs | Multiply-accumulate count per inference | ops |
| Energy_fJ | Energy per MAC (model input) | fJ/op |
| InferencesPerSec | Inference rate (model input) | 1/s |
| HoursPerMonth | Hours per month (model input) | hours |
| $/kWh | Electricity price (model input) | USD/kWh |
| GPU_cost, FeCIM_cost | Monthly cost inputs | USD/month |
| ServerScale | Number of servers (model input) | count |

## Assumptions And Limits

- Modeled numbers depend on configuration assumptions.
- Benchmarks are representative subsets.
- Comparisons should be interpreted with the honesty audit.
- ROI and cost figures are illustrative model outputs, not financial forecasts.

## Where It Lives In Code

- `module5-comparison/pkg/comparison/architecture.go`
- `module5-comparison/pkg/comparison/render.go`
- `module5-comparison/pkg/gui/widgets.go`

## Sources

- `docs/4-research/honesty-audit.md`
- `docs/development/SCRIPT_REFERENCE.md#demo-5-comparison-module5-comparison`
