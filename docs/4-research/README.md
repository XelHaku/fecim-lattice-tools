# Research & Validation

**Scientific foundations, literature review, and accuracy assessment for FeCIM simulation.**

---

## ⚠️ Read This First

This repository is **simulation-only**. All external scientific claims require explicit verification.

**Rule:** If a claim is not listed in [honesty-audit.md](honesty-audit.md) as verified, treat it as **unverified**.

See the full policy: [honesty-audit.md](honesty-audit.md)

---

## 📖 Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [honesty-audit.md](honesty-audit.md) | Claims verification status | All readers |
| [physics-validation.md](physics-validation.md) | Physics model accuracy | Researchers |
| [papers/](papers/) | 230+ indexed research papers | Researchers |
| [literature-review/](literature-review/) | Synthesized literature reviews | All |
| [opensource-tools/](opensource-tools/) | Related open-source tools | Developers |
| [validation/](validation/) | Model validation data | Researchers |

---

## 🔬 Scientific Accuracy Policy

### What Is Verified

Two external claims are currently verified in this project:

1. **98.24% MNIST accuracy** — reported for HZO ferroelectric tunnel junction (FTJ) reservoir computing in *Journal of Alloys and Compounds* (2025), DOI: `10.1016/j.jallcom.2025.181869`.
   - This is **not** a FeCIM device claim.
   - It must not be attributed to this simulator.

2. **97% MNIST accuracy with a current limiter, vs 9.8% without it** — reported for a 28 nm HKMG-based current-limited FeFET crossbar array in *IEEE Transactions on Electron Devices* (2022), DOI: `10.1109/TED.2022.3216973`.
   - This is an external hardware-paper benchmark.
   - It supports the importance of current limiting in FeFET arrays, not a blanket performance claim for this repository.

### What Is Simulation Baseline

The following are **simulation defaults, not measured hardware**:

- 30 analog states per cell (configurable baseline)
- Material parameters (Pr, Ec, endurance values)
- Energy estimates (physics-based projections)
- MNIST accuracy (simulator, not fabricated device)

### What Is Unverified

These appear in historical docs but are **not verified**:

- 30 discrete states for any specific device
- Energy multipliers vs NAND or GPUs (specific numbers)
- TRL levels or manufacturing readiness claims
- Automotive qualification (AEC-Q100)
- Cryogenic operation specifics

Full audit: [honesty-audit.md](honesty-audit.md)

---

## 📚 Research Papers (230+)

### Coverage: 23 Topics

The research library covers all major areas of ferroelectric CIM technology:

| # | Topic | Papers |
|---|-------|--------|
| 01 | Ferroelectric Materials | HZO, FeFET, FTJ fundamentals |
| 02 | Write/Erase Mechanisms | Switching dynamics, endurance |
| 03 | Simulation Tools | Open-source and commercial tools |
| 04 | Crossbar Non-Idealities | IR drop, sneak paths, drift |
| 05 | Neuromorphic Computing | Spiking neural networks |
| 06 | Quantization Methods | ADC/DAC optimization |
| 07 | Multi-Level Cells | State density, retention |
| 08 | Thermal Effects | Temperature dependence |
| 09 | Process Variation | Device-to-device spread |
| 10 | CIM Architectures | 1T1R, 0T1R, FeCAP |
| 11 | Reservoir Computing | Physical neural networks |
| 12 | On-Chip Training | In-memory learning |
| 13 | Benchmark Workloads | MNIST, ImageNet, etc. |
| 14 | CMOS Integration | BEOL compatibility |
| 15 | Energy Analysis | Efficiency projections |
| 16 | Photonic Hybrids | Optical-ferroelectric systems |
| 17 | Read Circuits | Sense amplifiers, TIA |
| 18 | Write Circuits | Drivers, charge pumps |
| 19 | Reliability | Retention, endurance limits |
| 20 | Manufacturing | Process integration |
| 21 | 3D Stacking | Monolithic 3D integration |
| 22 | Automotive/Harsh Env | High-temperature operation |
| 23 | Cryogenic Operation | Sub-4K operation |

Browse: [papers/by-topic/](papers/by-topic/)

### New Papers (2026)

Recent additions:
- [papers/NEW_PAPERS_2026-02-01.md](papers/NEW_PAPERS_2026-02-01.md)
- [papers/NEW_PAPERS_2026-01-29.md](papers/NEW_PAPERS_2026-01-29.md)
- [papers/NEW_PAPERS_2026-01-26.md](papers/NEW_PAPERS_2026-01-26.md)

### Strategic Topic Analysis

For prioritized research gaps and recommendations:
→ [papers/TOPIC_SUMMARIES.md](papers/TOPIC_SUMMARIES.md)
→ [papers/RESEARCH_GAP_ANALYSIS.md](papers/RESEARCH_GAP_ANALYSIS.md)

---

## 🔬 Literature Reviews

Synthesized analyses of key technical areas:

| Review | Topic | Key Findings |
|--------|-------|-------------|
| [literature-review/crossbar-circuits.md](literature-review/) | Crossbar circuits | FeCAP eliminates sneak paths; 4-bit DAC/ADC is optimal |
| [papers/TOPIC_SUMMARIES.md](papers/TOPIC_SUMMARIES.md) | All 23 topics | Strategic summary per topic |

### Key Findings from 2025-2026 Review

From the crossbar circuits literature review:

1. **FeCAP architecture** eliminates sneak paths and IR drop entirely (capacitive crossbars)
2. **4-bit DAC/ADC** is literature-optimal (not 5-bit as previously assumed)
3. **ADC dominates** 40-60% of system energy
4. **State-dependent C2C variation** is not yet modeled in this simulator
5. **Non-linear I-V curves** are missing from current crossbar model
6. **Charge-domain sensing** is needed for FeCAP mode

These findings drive the priority backlog. See [papers/RESEARCH_GAP_ANALYSIS.md](papers/RESEARCH_GAP_ANALYSIS.md).

---

## 🧮 Physics Validation

### What the Models Are

The simulator uses two physics engines:

**Preisach Model:**
- Mathematical hysteresis using Tanh Everett function
- Product-form Everett: `[1+tanh((α-Ec)/Δ)] × [1-tanh((β+Ec)/Δ)] × Ps/4`
- This is the mathematically correct form (integral of sech² density)
- Major loop shape and Pr/Ps ratio match standard measurements

**Landau-Khalatnikov (L-K) Model:**
- Dynamic polarization solver using 4th-order Runge-Kutta
- Models switching time and viscosity
- Used for ISPP write controller physics

### Validation Status

| Model | Status | Notes |
|-------|--------|-------|
| Preisach hysteresis | Physics-correct | Product-form Everett, verified against L-K |
| L-K dynamics | Physics-correct | RK4 integration, compared to Preisach |
| Crossbar MVM | Exact | Ohm's Law + KCL |
| IR drop (SOR solver) | Approximate | Iterative, configurable tolerance |
| DAC/ADC models | Literature-based | INL/DNL from standard model |
| Energy estimates | Physics-based | Not device-measured |

Full report: [physics-validation.md](physics-validation.md)

---

## 🛠️ Open-Source Tools Survey

Analysis of related open-source simulation tools:

| Document | Content |
|----------|---------|
| [opensource-tools/ferroelectric-simulation-tools.md](opensource-tools/) | Available ferroelectric simulators |
| [opensource-tools/tool-comparison-matrix.md](opensource-tools/) | Side-by-side comparison |
| [opensource-tools/data-acquisition-tools.md](opensource-tools/) | Measurement software |

Key finding: No existing open-source tool covers the full FeCIM design flow from device physics to system-level simulation. This project fills that gap for education.

---

## 🔎 Material-Specific Research

### Superlattice Materials

Analysis of HZO superlattice parameters for the `literature_superlattice` material preset:

- [validation/](validation/) - Calibration validation data
- Data source: Literature values from peer-reviewed publications

---

## 📊 How to Cite

### Citing This Simulator

```bibtex
@software{fecim_lattice_tools,
  title = {FeCIM Lattice Tools: Educational Ferroelectric CIM Simulator},
  year = {2026},
  note = {Simulation tool only. Results not validated on fabricated hardware.}
}
```

### Citing the Physics Models

For the Preisach model:
```bibtex
@article{preisach1935,
  author = {Preisach, F.},
  title = {Über die magnetische Nachwirkung},
  journal = {Zeitschrift für Physik},
  year = {1935}
}
```

For Landau-Khalatnikov:
```bibtex
@article{lk_model,
  title = {On the anomalous heat capacity of a ferroelectric crystal},
  author = {Khalatnikov, I. M.},
  journal = {Zh. Eksp. Teor. Fiz},
  year = {1954}
}
```

For the verified MNIST accuracy claim:
```bibtex
@article{hzo_ftj_reservoir_2025,
  title = {Reservoir computing with HZO FTJ},
  journal = {Journal of Alloys and Compounds},
  year = {2025},
  doi = {10.1016/j.jallcom.2025.181869}
}
```

---

## 🔗 Priority Research Gaps

Based on the 2026 literature review, these areas need attention:

### Priority 0 (Blocking Accuracy)
- Switch to 4-bit default DAC/ADC (literature-optimal)
- Add exponential conductance default (non-linear I-V)

### Priority 1 (Important Features)
- Multiple ADC types (Flash, SAR, Slope)
- ADC sharing across columns
- State-dependent cycle-to-cycle variation

### Priority 2 (Advanced)
- FeCAP architecture (capacitive crossbar, no sneak paths)
- Charge-domain sensing
- Non-linear I-V device model

Full analysis: [papers/RESEARCH_GAP_ANALYSIS.md](papers/RESEARCH_GAP_ANALYSIS.md)

---

## 🔗 Quick Links

**Scientific Accuracy:**
- [Honesty Audit](honesty-audit.md) - What is/isn't verified
- [Physics Validation](physics-validation.md) - Model accuracy

**Research Library:**
- [All Papers by Topic](papers/by-topic/) - 23 topic categories
- [Topic Summaries](papers/TOPIC_SUMMARIES.md) - Strategic overview
- [Research Gaps](papers/RESEARCH_GAP_ANALYSIS.md) - What's missing

**Technical Analysis:**
- [Internal Analysis](internal-analysis/) - In-depth technical docs
- [Literature Reviews](literature-review/) - Synthesized findings
- [Open-Source Tools](opensource-tools/) - Tool comparison

---

**Last Updated:** 2026-02-16
**Papers Indexed:** 230+ (23 topics)
**Honesty Policy:** Verified claims only presented as facts
