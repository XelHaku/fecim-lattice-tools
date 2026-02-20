# FeCIM Lattice Tools -- Documentation Index

This documentation covers the six interactive modules that make up the FeCIM Lattice Tools simulator. Each module teaches a different layer of the ferroelectric compute-in-memory stack, from single-cell physics to chip layout. They are designed to be studied in order, though any module can be explored independently.

| # | Module | What You Learn | Key Demo |
|---|--------|---------------|----------|
| 1 | [Hysteresis](../2-learn/module1-hysteresis/README.md) | How ferroelectric materials remember: polarization, P-E loops, Preisach and L-K models, multi-level storage | Real-time P-E curve with material switching, waveform modes, and ISPP write-verify |
| 2 | [Crossbar](../2-learn/module2-crossbar/README.md) | How a grid of cells computes: Ohm's law multiplication, Kirchhoff summation, IR drop, sneak paths, device variation | Interactive crossbar MVM with toggleable non-idealities and architecture comparison (0T1R/1T1R/2T1R) |
| 3 | [MNIST](../2-learn/module3-mnist/README.md) | How two crossbar layers recognize handwritten digits: quantization, noise injection, FP vs CIM accuracy | Draw-your-own-digit canvas with dual-mode floating-point and CIM inference side by side |
| 4 | [Circuits](../2-learn/module4-circuits/README.md) | The peripheral signal chain: DAC input encoding, TIA current sensing, ADC output quantization, charge pump programming | Signal-chain walkthrough showing INL/DNL, voltage zones, and ISPP convergence |
| 5 | [Comparison](../2-learn/module5-comparison/README.md) | Architecture trade-offs: CPU vs GPU vs FeCIM on energy, latency, and cost across real AI workloads | Interactive ROI calculator with configurable workload, electricity rate, and fleet size |
| 6 | [EDA](../2-learn/module6-eda/README.md) | From simulation to silicon: tiling, SPICE/Verilog/DEF/LEF export, OpenLane integration, DRC | Compile a weight matrix into tiled layout files and run design-rule checks |

## Additional Resources

| Resource | Path |
|----------|------|
| Science overview (start here) | [About the Science](about/About.Science.md) |
| Module reference card | [MODULES.md](MODULES.md) |
| Full ELI5 guide | [eli5-overview.md](../2-learn/eli5-overview.md) |
| Technical glossary | [GLOSSARY.md](../GLOSSARY.md) |
| Accuracy and honesty policy | [honesty-audit.md](../4-research/honesty-audit.md) |

Use F1 or the **?** button to return to [About the Science](about/About.Science.md).
