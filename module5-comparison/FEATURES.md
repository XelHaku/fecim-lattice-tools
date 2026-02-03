# Module 5: Comparison - Features

FeCIM vs CPU/GPU business-case visualization (model-based, not measured).

---

## Features

- **Architecture Comparison** - CPU+DRAM vs GPU+HBM vs FeCIM CIM
- **Workload Library** - MNIST, ResNet-50, BERT-Base, GPT-2, LLM-70B
- **Data-Center Scaling** - Power, cost, CO2 projections from model inputs
- **Animated Visuals** - Energy race, market chart, ROI summary
- **Presentation Modes** - Manual, Auto Demo, Investor, Engineer

---

## Default Architecture Models (From Code)

| Architecture | Node | TDP | Peak TOPS | Notes |
|---|---:|---:|---:|---|
| CPU+DRAM | 5 nm | 125 W | 1 | Baseline model |
| GPU+HBM | 4 nm | 400 W | 100 | Baseline model |
| FeCIM CIM | 45 nm | 5 W | 50 | **Estimated** (visualization only) |

> FeCIM values are explicitly marked **estimated** in code and should not be treated as measured device specs.

---

## Market Chart

- Market segments and totals are **scenario inputs** used for visualization.
- The code comments cite WSTS/Gartner as sources; verify before external use.

---

## Known Limitations

- Cross-architecture comparisons depend on assumed parameters.
- No hardware benchmarking or silicon validation is performed.
