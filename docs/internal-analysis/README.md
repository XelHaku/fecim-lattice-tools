# Internal Analysis Documents

> **Purpose**: Synthesized research notes for the FeCIM Lattice Tools project.
> These documents summarize reported literature and are **not** independently verified.

## Document Index

| Document | Topic | Key Content |
|----------|-------|-------------|
| [hysteresis-physics.md](hysteresis-physics.md) | Ferroelectric Physics | P-E curves, Preisach model, HfO₂-ZrO₂ materials |
| [crossbar-arrays.md](crossbar-arrays.md) | Array Architecture | MVM, IR drop, sneak paths, 0T1R/1T1R/2T1R |
| [cim-circuits.md](cim-circuits.md) | Peripheral Circuits | DAC/ADC/TIA, energy efficiency, CMOS |
| [eda-chip-design.md](eda-chip-design.md) | Chip Design | OpenLane, PDKs, RTL-to-GDSII flow |
| [circuits.CIM-fundamentals.md](circuits.CIM-fundamentals.md) | CIM Operations | READ/WRITE/COMPUTE physics |
| [module2-vs-module4-physics-comparison.md](module2-vs-module4-physics-comparison.md) | Module Comparison | Crossbar vs Circuit abstraction |
| [MODULE4-PHYSICS-IMPROVEMENTS.md](MODULE4-PHYSICS-IMPROVEMENTS.md) | Gap Analysis | Physics improvements roadmap |

## Relationship to Research Papers

```
docs/research-papers/          ← External literature (230+ papers)
    by-topic/                  ← 25 topic directories with PDFs + READMEs
    _tools/downloaded/         ← Recently downloaded papers by source
    paper_metadata.json        ← Machine-readable paper index
    RESEARCH_GAP_ANALYSIS.md   ← Coverage assessment (A+ grade: 97/100)

docs/internal-analysis/        ← Synthesized analysis (this folder)
    *.md                       ← Topic syntheses with extracted data
```

## Key Physics Constants (Reported Examples)

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 (demo baseline) | Simulator default (configurable) |
| Multi-level states | multi-level (reported) | Literature summaries |
| Pr (RT) | reported ranges | Literature summaries |
| Pr (4K) | reported ranges | Literature summaries |
| Ec | reported ranges | Literature summaries |
| Endurance | reported ranges | Literature summaries |
| MNIST accuracy | reported ranges | Literature summaries |
| Energy efficiency | reported ranges | Literature summaries |

## Accuracy & Honesty

All claims in these documents are **reported** or **illustrative** and should not be treated as verified by this project. For current policy, see `docs/4-research/honesty-audit.md`.

## Contributing

When adding new analysis documents:

1. Include DOIs for all key claims
2. Reference specific papers from `/docs/research-papers/`
3. Mark any unverified simulation baselines
4. Add entry to this README
