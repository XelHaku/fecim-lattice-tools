<!-- Category: Physics | Module: module5-comparison | Reading time: ~4 min -->
# Module 5 Physics: Comparison Models

> Equations and models used to compare CIM architectures against
> conventional computing approaches.

All values are **model inputs or model outputs** for visualization
and education. They are not measured hardware results or financial
forecasts.

---

## Core Metrics

Three fundamental metrics drive all comparisons:

```
Energy      = Power * Time                    [Joules]
Throughput  = Operations / Time               [ops/s]
Efficiency  = Operations / Energy             [ops/J or TOPS/W]
```

---

## Architecture Comparison Engine

**Code location:** `module5-comparison/pkg/comparison/architecture.go`

Each architecture (CPU, GPU, FeCIM) is described by a parameter set:

| Parameter | Meaning | Units |
|-----------|---------|-------|
| TDP | Thermal design power | W |
| Latency | Time per inference | ms |
| TOPS | Peak throughput | tera-ops/s |
| TOPS_per_W | Energy efficiency | TOPS/W |

Derived quantities:

```
Energy_mJ          = TDP_W * Latency_ms
Energy_kWh         = Energy_mJ / 3.6e9
Cost_per_inference  = Energy_kWh * price_per_kWh
```

---

## ROI Calculator

**Code location:** `module5-comparison/pkg/gui/widgets.go`

The GUI provides a configurable ROI calculator for data-center scale
projections:

```
E_inf (uJ)       = MACs * Energy_fJ / 1e9
P (W)            = E_inf (uJ) * InferencesPerSec / 1e6
Cost_month ($)   = (P / 1000) * HoursPerMonth * price_per_kWh
Savings_year ($) = (GPU_cost - FeCIM_cost) * 12 * ServerScale
```

| Parameter | Meaning | Units |
|-----------|---------|-------|
| MACs | Multiply-accumulate operations per inference | ops |
| Energy_fJ | Energy per MAC (model input) | fJ/op |
| InferencesPerSec | Inference rate | 1/s |
| HoursPerMonth | Operating hours | hours |
| price_per_kWh | Electricity price | USD/kWh |
| ServerScale | Number of servers | count |

These are **model inputs** -- changing any assumption changes the
output. The calculator is a "what-if" tool, not a financial forecast.

---

## What Gets Compared

The module focuses on relative differences across these axes:

```
  Energy efficiency:  How many operations per joule?
  Latency:           How fast is one inference?
  Throughput:        How many inferences per second at scale?
  Cost:              What does it cost to run at data-center scale?
```

Comparisons are presented as ratios or side-by-side tables, not
absolute claims.

---

## Key Assumptions to Watch

1. **CIM precision:** The comparison assumes 4-bit analog precision.
   Higher precision requires more ADC bits, which changes the energy
   and area equations.

2. **ADC energy dominance:** In CIM systems, the ADC typically
   consumes 40-60% of total system energy. The comparison engine
   includes ADC energy as part of the CIM efficiency calculation.

3. **Digital baseline:** GPU numbers use published TDP and throughput
   figures as model inputs. They represent a specific generation of
   hardware at a specific operating point.

4. **No system overhead:** The comparison does not include memory
   controller overhead, host CPU interaction, or data preprocessing.

---

## Assumptions and Limits

- All modeled numbers depend on configuration assumptions.
- Benchmarks are representative subsets, not exhaustive.
- Comparisons should be interpreted alongside the honesty audit.
- ROI and cost figures are illustrative, not financial advice.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module5-comparison/pkg/comparison/architecture.go` | Architecture profiles and comparison engine |
| `module5-comparison/pkg/comparison/render.go` | Chart and table rendering |
| `module5-comparison/pkg/gui/widgets.go` | ROI calculator and GUI |
| `module5-comparison/pkg/gui/app.go` | Module entry point |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
