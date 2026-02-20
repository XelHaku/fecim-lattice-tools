# About the Science: Ferroelectric Compute-in-Memory

What if memory could do math? Ferroelectric Compute-in-Memory (FeCIM) eliminates the bottleneck between where data is stored and where computation happens by performing analog matrix-vector multiplication directly inside a crossbar memory array. This tool suite lets you explore the physics, circuits, and system-level implications of that idea -- from a single ferroelectric memory cell all the way to chip layout -- without needing a fabrication lab.

---

## The Core Idea

Modern AI accelerators spend most of their energy moving data between memory and processor. The fundamental insight behind compute-in-memory is that Ohm's law already performs multiplication: when you apply a voltage V across a programmable conductance G, the resulting current is I = V x G. Arrange thousands of these cells into a grid -- a crossbar array -- and Kirchhoff's current law sums the products along each column for free. One voltage pulse in, one current measurement out, and you have a full matrix-vector multiply.

Ferroelectric materials make this practical because they hold their conductance state without power. Hafnium-zirconium-oxide (HZO) and related superlattices can be programmed to many distinct analog levels, CMOS-compatible, and retain state for years. Each cell stores a neural-network weight as a conductance value; inference is then a single analog pass through the array.

The catch is that real hardware is noisy. Wire resistance causes IR drop, parasitic sneak paths corrupt reads in passive arrays, device-to-device variation blurs conductance levels, and peripheral circuits (DAC, TIA, ADC) add their own quantization error. This simulator models all of those effects so you can build intuition about what matters and what can be engineered around.

```
         Digital Input                  Digital Output
         (voltage codes)                (current codes)
              |                              ^
              v                              |
          +-------+                      +-------+
          |  DAC  |                      |  ADC  |
          +---+---+                      +---+---+
              |  V_1  V_2  V_3              |
              v   v    v    v               |
            +-+---+----+----+-+         +---+---+
     WL_0 --| G00 | G01 | G02 |--+---->| TIA_0 |
            +-----+-----+-----+  |     +-------+
     WL_1 --| G10 | G11 | G12 |--+---->| TIA_1 |
            +-----+-----+-----+  |     +-------+
     WL_2 --| G20 | G21 | G22 |--+---->| TIA_2 |
            +-----+-----+-----+        +-------+
              BL_0  BL_1  BL_2

        I_col = SUM( G_row,col x V_row )
        One step. All rows in parallel.
```

---

## Learning Path

The six modules form a progressive curriculum. Each builds on the concepts introduced by the previous one.

```
  [Module 1: Hysteresis]
        |
        | How a single cell remembers
        v
  [Module 2: Crossbar]
        |
        | How an array of cells computes
        v
  [Module 3: MNIST]
        |
        | How two layers recognize digits
        v
  [Module 4: Circuits]
        |
        | The DAC/TIA/ADC signal chain
        v
  [Module 5: Comparison]            [Module 6: EDA]
        |                                 |
        | CPU vs GPU vs FeCIM             | From simulation
        |                                 | to silicon layout
        v                                 v
    Architecture                     Physical design
    trade-offs                       and fabrication
```

Module 6 (EDA) can be studied independently at any time -- it covers chip-design methodology rather than FeCIM physics.

---

## The 6 Modules

**Module 1 -- Hysteresis: The Memory Cell**
Visualize how ferroelectric polarization responds to an applied electric field. The P-E hysteresis loop is the foundation of everything that follows -- it determines how many analog levels a cell can hold, how reliably it retains them, and how much energy each write costs. The demo lets you sweep fields, switch materials, and watch Preisach or Landau-Khalatnikov models in real time.

**Module 2 -- Crossbar: Compute-in-Memory**
Build a crossbar array from the cells you explored in Module 1 and run matrix-vector multiplications through it. Toggle IR drop, sneak paths, device variation, and drift to see how each non-ideality degrades the result. Compare passive (0T1R), single-transistor (1T1R), and dual-transistor (2T1R) architectures side by side.

**Module 3 -- MNIST: Neural Network Inference**
Load a trained neural network into two crossbar layers and recognize handwritten digits. The dual-mode view shows floating-point and quantized/noisy CIM inference in parallel so you can directly measure the accuracy cost of going analog. Draw your own digits and watch confidence bars shift.

**Module 4 -- Circuits: The Supporting Cast**
No crossbar works in isolation. This module simulates the peripheral signal chain: DAC (digital voltage in), TIA (analog current out), ADC (digital code out), and charge pump (high-voltage generation for programming). Explore INL/DNL, bandwidth, and how bit-resolution choices ripple through the whole system.

**Module 5 -- Comparison: Why It Matters**
Put FeCIM in context against CPU and GPU architectures. An interactive ROI calculator lets you plug in workload size, batch count, electricity rate, and hardware lifetime to compare energy efficiency (TOPS/W), latency, and total cost of ownership. All FeCIM values are model-based projections, clearly labeled.

**Module 6 -- EDA: From Simulation to Silicon**
Compile a weight matrix into tiled crossbar arrays and export SPICE netlists, Verilog RTL, DEF placement, and LEF macros. Integrate with the open-source OpenLane/OpenROAD flow and Sky130 PDK to explore what a physical FeCIM chip layout looks like. Run design-rule checks and layout-vs-schematic verification.

---

## Key Concepts

| Term | Definition |
|------|-----------|
| Polarization (P) | Net electric dipole moment per unit volume in a ferroelectric. Serves as the state variable for analog storage: higher P corresponds to higher conductance. |
| Hysteresis | Path-dependent response: the polarization depends on the history of applied fields, not just the current value. This is what makes the material a non-volatile memory. |
| Coercive Field (Ec) | The electric field magnitude required to switch the polarization direction. Determines the minimum write voltage. |
| Crossbar Array | A grid of programmable conductances at the intersections of horizontal word lines and vertical bit lines. Performs matrix-vector multiplication in a single step via Ohm's law and Kirchhoff's current law. |
| ISPP | Incremental Step Pulse Programming -- an iterative write-verify loop that converges a cell to a target conductance level using progressively adjusted voltage pulses. |
| Analog Inference | Running a neural network using analog currents through a crossbar rather than digital multiply-accumulate units. Trades precision for massive parallelism and energy efficiency. |
| Preisach Model | A mathematical framework that represents hysteresis as a superposition of elementary bistable switches (hysterons), each with its own switching thresholds. |
| Landau-Khalatnikov (L-K) | A thermodynamic equation of motion for polarization based on the Landau free-energy expansion. Captures time-dependent switching dynamics. |
| IR Drop | Resistive voltage loss along interconnect wires in large arrays. Causes position-dependent programming and read errors. |
| Sneak Path | Parasitic current that flows through unselected cells in passive (0T1R) crossbar arrays, corrupting read and write operations. |

---

## Quick Start

New here? Start with the [Module 1 ELI5](../../2-learn/module1-hysteresis/eli5.md) -- a short read that builds all the intuition you need about ferroelectric memory, no equations required. Then launch the GUI:

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools hysteresis
```

From there, follow the learning path above through each module at your own pace.

---

## Honest Scope Note

This is a simulation and education tool, not a validated silicon product. All material parameters are configurable baselines for learning and design-space exploration. The 30-level conductance quantization is a simulation default, not a measured device specification. Energy-efficiency projections, latency estimates, and cost comparisons are model-based and clearly labeled throughout -- they have not been validated against fabricated hardware. Where peer-reviewed measurements exist, they are cited with DOIs; everything else should be treated as illustrative. Real ferroelectric CIM devices require prototype fabrication, characterization, and system-level benchmarking before any performance claim can be considered demonstrated. See [Honesty Audit](../../4-research/honesty-audit.md) for the full accuracy policy.
