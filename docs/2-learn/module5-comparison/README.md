# Module 5: Technology Comparison

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md)

---

## Overview

Module 5 provides architecture-level comparison between CPU, GPU, and FeCIM technologies across various AI workloads. It visualizes relative performance, energy efficiency, and cost trade-offs using model-based estimates.

**Key Concept:** Compare compute architectures on metrics that matter—energy per inference, throughput, and total cost of ownership. All values are model inputs for educational visualization.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - Why compare architectures?

### For Developers
- **[Physics Reference](./physics.md)** - Performance equations, ROI calculations
- **[Features](./features.md)** - Workload library, visualization

### For Researchers
- **[Honesty Audit](../../4-research/honesty-audit.md)** - Validation status of all claims

---

## Module Contents

```
module5-comparison/
├── pkg/comparison/
│   ├── architecture.go       # CPU/GPU/FeCIM models, workloads, ROI
│   ├── cacti_baseline.go     # CACTI-normalized DRAM/SRAM baselines
│   └── render.go             # Chart generation
├── pkg/gui/
│   ├── app.go                # Main comparison app + calculator
│   ├── widgets.go            # Comparison visualizations
│   ├── evidence_model.go     # Evidence-based claim tracking
│   └── fabrication_reality.go # Fab reality checks
└── cmd/comparison/
    └── main.go               # CLI entry point
```

---

## Quick Start

### GUI Mode
```bash
fecim-lattice-tools comparison
```

### Headless Report
```bash
fecim-lattice-tools --mode comparison --workload llm-70b --export report.json
```

---

## What You'll Learn

1. **Architecture Trade-offs**
   - CPU: Flexible, high latency
   - GPU: Parallel, high power
   - FeCIM: Energy-efficient, analog

2. **Performance Metrics**
   - TOPS (Tera-Operations Per Second)
   - TOPS/W (Energy efficiency)
   - Latency per inference

3. **Cost Analysis**
   - Energy cost ($/kWh × usage)
   - Hardware cost amortization
   - Total Cost of Ownership (TCO)

4. **Workload Characteristics**
   - Compute-bound vs memory-bound
   - Batch size impact
   - Model size scaling

---

## Comparison Framework

### Architecture Models (Configurable Inputs)

| Architecture | TOPS | Power (W) | TOPS/W | Cost |
|--------------|------|-----------|--------|------|
| **CPU** (Xeon) | 0.5 | 150 | 0.003 | $5K |
| **GPU** (A100) | 312 | 400 | 0.78 | $15K |
| **FeCIM** (Modeled) | 50 | 10 | 5.0 | $2K* |

*Model-based estimate, not production hardware

### Workload Library

```
MNIST (Small CNN):
  - MACs: 100K
  - Batch: 1
  - Latency target: <10ms

ResNet-50 (Image Classification):
  - MACs: 4B
  - Batch: 32
  - Latency target: <100ms

LLM-70B (Large Language Model):
  - MACs: 70B
  - Batch: 1
  - Latency target: <1s
```

---

## Key Visualizations

### 1. Energy Efficiency Chart

```
TOPS/W (higher is better):
CPU:   ███ 0.003
GPU:   ████████████████ 0.78
FeCIM: ██████████████████████████ 5.0 (modeled)
```

### 2. Latency Comparison

```
Time per inference (lower is better):
         MNIST  ResNet  LLM-70B
CPU:     10ms   2s      1min
GPU:     1ms    50ms    5s
FeCIM:   0.5ms  20ms    2s (modeled)
```

### 3. Cost Breakdown

```
Monthly Operating Cost (1M inferences/day):
         Energy  Hardware  Total
CPU:     $450    $42       $492
GPU:     $1200   $125      $1325
FeCIM:   $30     $17       $47 (modeled)
```

---

## ROI Calculator

### Inputs (User-Configurable)

```
- Workload: MNIST / ResNet / LLM
- Inferences per day: 1M
- Electricity rate: $0.12/kWh
- Hardware lifetime: 3 years
- Server count: 10
```

### Outputs (Model-Based)

```
Annual Savings (FeCIM vs GPU):
  Energy: $14,040
  Hardware: $1,080
  Total: $15,120 per year

Payback period: 18 months (modeled)
```

---

## Features

- **Architecture library:** CPU, GPU, FeCIM with configurable parameters
- **Workload library:** MNIST, ImageNet, LLM benchmarks
- **Interactive charts:** Energy, latency, cost comparisons
- **ROI calculator:** Customize inputs, see projected savings
- **Scalability analysis:** 1 to 1000 servers
- **Export reports:** JSON, CSV, PDF

---

## Important Disclaimers

### Model-Based Estimates

All FeCIM values are **model-based projections**, not measured hardware:

- **Energy efficiency:** Estimated from device physics + simulator
- **Latency:** Based on analytical models, not silicon
- **Cost:** Projected manufacturing cost, not retail pricing
- **Reliability:** Assumed from literature, not long-term testing

### Validation Status

See `docs/4-research/honesty-audit.md` for complete validation status of all claims.

**Demonstrated:** Simulation framework, comparison methodology
**Modeled:** Performance estimates, energy calculations
**Aspirational:** Production deployment, market adoption

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | Why compare? | Beginners |
| [physics.md](./physics.md) | Equations, ROI math | Developers |
| [features.md](./features.md) | Workloads, visualizations | All |
| [honesty-audit.md](../../4-research/honesty-audit.md) | Validation status | Researchers |

---

## Related Modules

- **[Module 2: Crossbar](../module2-crossbar/README.md)** - FeCIM energy model source
- **[Module 3: MNIST](../module3-mnist/README.md)** - Benchmark workload
- **[Module 6: EDA](../module6-eda/README.md)** - Area estimates

---

## Testing

```bash
go test ./module5-comparison/pkg/comparison
```

---

**Last Updated:** 2026-02-16
