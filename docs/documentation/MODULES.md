# Module Reference Card

Quick-reference summary for every module in the FeCIM Lattice Tools suite. For the full learning narrative, see [About the Science](about/About.Science.md).

---

## Module 1: Hysteresis

**In one line:** Simulate and visualize ferroelectric P-E hysteresis loops with Preisach and Landau-Khalatnikov engines.

**Physics:**
- Polarization vs electric field (P-E) curves with path-dependent memory (hysteresis)
- Two physics engines: Preisach (hysteron superposition) and Landau-Khalatnikov (thermodynamic free-energy solver)

**Key files:** `docs/2-learn/module1-hysteresis/eli5.md`, `physics.md`, `features.md`
**Source:** `module1-hysteresis/pkg/ferroelectric/` (models), `pkg/controller/` (ISPP), `pkg/gui/` (visualization)

**Demo shows:** Real-time P-E loop responding to manual, sine, triangle, square, and random-walk waveforms. Switch between HZO, superlattice, AlScN, and cryogenic material presets. Run ISPP write-verify to program discrete conductance levels.

---

## Module 2: Crossbar

**In one line:** Build crossbar arrays and run matrix-vector multiplications with full non-ideality modeling.

**Physics:**
- Ohm's law multiplication (I = G x V) and Kirchhoff current-law accumulation across rows
- Non-idealities: IR drop (wire resistance), sneak paths (parasitic currents), device variation, drift, and temperature effects

**Key files:** `docs/2-learn/module2-crossbar/eli5.md`, `physics.md`, `features.md`, `architecture.md`
**Source:** `module2-crossbar/pkg/crossbar/` (core), `shared/crossbar/` (shared array logic)

**Demo shows:** Interactive heatmap of an N x N crossbar. Toggle IR drop, sneak paths, and variation on or off. Compare passive (0T1R), 1T1R, and 2T1R architectures. Visualize MVM output error vs ideal.

---

## Module 3: MNIST

**In one line:** Recognize handwritten digits through two crossbar layers, comparing floating-point and quantized CIM inference.

**Physics:**
- Weight and activation quantization to 30 discrete levels (configurable)
- Multiplicative Gaussian noise model for device variation (sigma/mu)

**Key files:** `docs/2-learn/module3-mnist/eli5.md`, `physics.md`, `features.md`
**Source:** `module3-mnist/pkg/core/` (inference engine, quantization), `pkg/gui/` (dual-mode UI)

**Demo shows:** Draw a digit on the canvas and see FP vs CIM predictions side by side. Adjust quantization levels (8/16/30/64) and noise (0-20%) to watch accuracy degrade gracefully. Activation heatmaps and confidence bars update in real time.

---

## Module 4: Circuits

**In one line:** Simulate the DAC, TIA, ADC, and charge-pump peripherals that interface digital control with the analog crossbar.

**Physics:**
- DAC: code-to-voltage transfer function, LSB resolution, INL/DNL nonlinearity
- ADC: voltage-to-code quantization with +/-0.5 LSB error; TIA: current-to-voltage gain with offset and noise

**Key files:** `docs/2-learn/module4-circuits/eli5.md`, `physics.md`, `features.md`
**Source:** `shared/peripherals/` (circuit models), `module4-circuits/pkg/gui/` (visualization), `pkg/arraysim/` (integration)

**Demo shows:** Full signal-chain walkthrough from digital input through DAC, across a crossbar cell, through TIA sensing, to ADC output. Voltage-zone visualization for 0T1R/1T1R/2T1R. ISPP programming convergence trace.

---

## Module 5: Comparison

**In one line:** Compare CPU, GPU, and FeCIM architectures on energy efficiency, latency, and total cost of ownership.

**Physics:**
- TOPS/W efficiency modeling from device-level energy per MAC through system-level fleet costs
- Workload characterization: compute-bound vs memory-bound, batch-size scaling

**Key files:** `docs/2-learn/module5-comparison/eli5.md`, `physics.md`, `features.md`
**Source:** `module5-comparison/pkg/comparison/` (models, workloads, ROI calculator), `pkg/gui/` (charts)

**Demo shows:** Bar charts comparing TOPS/W, latency per inference, and monthly operating cost for CPU (Xeon), GPU (A100), and FeCIM (modeled). Interactive ROI calculator with user-configurable workload, electricity rate, and fleet size. All FeCIM values clearly labeled as model-based projections.

---

## Module 6: EDA

**In one line:** Compile weight matrices into tiled crossbar layouts and export industry-standard SPICE, Verilog, DEF, and LEF files.

**Physics:**
- Algorithm-to-layout abstraction: weights to tiles to RTL to placement to GDS
- Design-rule checking (DRC) and layout-vs-schematic (LVS) validation

**Key files:** `docs/2-learn/module6-eda/eli5.md`, `physics.md`, `features.md`
**Source:** `module6-eda/pkg/compiler/` (tiling), `pkg/export/` (SPICE, Verilog, DEF, LEF), `pkg/validation/` (DRC)

**Demo shows:** Tile a 512x512 weight matrix into 128x128 crossbar arrays. Export SPICE netlists and DEF placement files. Integrate with the open-source OpenLane/OpenROAD flow and Sky130 PDK. Run design-rule checks against target technology constraints.

---

## Cross-Module Connections

```
Module 1 (cell physics)
    |
    | conductance levels feed into
    v
Module 2 (array computation) <----> Module 4 (peripheral circuits)
    |                                     |
    | quantized weights load into         | DAC/ADC bit-width choices
    v                                     | affect inference accuracy
Module 3 (neural inference) <-------------+
    |
    | accuracy/energy metrics compared in
    v
Module 5 (architecture benchmarks)

Module 6 (physical design) takes array specs from Modules 2 and 4
```
