<!-- Category: ELI5 | Module: module6-eda | Reading time: ~4 min -->
# Module 6 ELI5: Electronic Design Automation

> How do chips go from idea to silicon? EDA tools are the software
> that makes chip design possible.

---

## The Chip Design Flow

Designing a chip is like building a house. You start with a blueprint,
then work through increasingly detailed plans until you have something
a factory can build.

```
  Idea (weight matrix)
    |
    v
  Architecture (tiled crossbar arrays)
    |
    v
  RTL (structural Verilog -- the "wiring diagram")
    |
    v
  Synthesis (Yosys -- map to standard cells)
    |
    v
  Place & Route (OpenROAD -- lay out on silicon)
    |
    v
  GDS (final mask layout -- ready for fabrication)
```

Each step adds more detail. Module 6 handles the top half of this
flow: taking a weight matrix, partitioning it into crossbar-sized
tiles, and exporting the design in industry-standard file formats.

---

## Tiling: Cutting the Matrix to Fit

A neural network weight matrix might be 512x512, but a single
crossbar array might only be 128x128. The compiler partitions the
big matrix into tiles that fit:

```
  512x512 weight matrix, 128x128 arrays:

  TilesRows = ceil(512 / 128) = 4
  TilesCols = ceil(512 / 128) = 4
  Total tiles = 16

  +------+------+------+------+
  | T00  | T01  | T02  | T03  |
  +------+------+------+------+
  | T10  | T11  | T12  | T13  |
  +------+------+------+------+
  | T20  | T21  | T22  | T23  |
  +------+------+------+------+
  | T30  | T31  | T32  | T33  |
  +------+------+------+------+

  Each tile = one 128x128 crossbar array
```

Each tile gets placed at a physical coordinate and connected to its
neighbors through routing.

---

## Export Formats: Speaking the Industry Language

The chip industry uses specific file formats at each stage. Module 6
exports several:

| Format | What It Describes | Who Uses It |
|--------|-------------------|-------------|
| **SPICE** (.sp) | Electrical circuit netlist | Circuit simulators (ngspice, Xyce) |
| **Verilog** (.v) | Structural hardware description | Synthesis tools (Yosys) |
| **DEF** (.def) | Physical placement of cells | Place-and-route (OpenROAD) |
| **LEF** (.lef) | Cell physical abstractions | Place-and-route (OpenROAD) |
| **Liberty** (.lib) | Timing and power models | Static timing analysis (OpenSTA) |
| **JSON / CSV** | Data interchange | Analysis scripts |

---

## What PDKs Are and Why SKY130 Matters

A PDK (Process Design Kit) is a package of rules and models from a
semiconductor foundry. It tells your EDA tools:

- How small wires can be (minimum width, spacing)
- What transistors look like (device models)
- What metal layers are available
- Design rules the factory enforces

**SKY130** is the first fully open-source PDK, released by SkyWater
Technology and Google. It describes a 130 nm process and lets anyone
design chips without signing an restricted access or paying license fees.

Module 6 uses SKY130-compatible defaults for cell dimensions and
design rule checks.

---

## Where This CIM Chip Fits

```
  Module 6 handles these steps:
  +-----------------------------------------+
  | Weight matrix --> Tiled arrays           |  Compiler
  | Tiled arrays --> SPICE / Verilog / DEF   |  Exporter
  | Basic DRC checks                         |  Validator
  +-----------------------------------------+

  OpenLane / OpenROAD handles these steps:
  +-----------------------------------------+
  | Synthesis (Verilog --> gates)            |
  | Floorplanning                            |
  | Placement and routing                    |
  | Signoff (DRC, LVS, STA)                 |
  +-----------------------------------------+

  Foundry handles the rest:
  +-----------------------------------------+
  | Mask fabrication                         |
  | Wafer processing                         |
  | Testing and packaging                    |
  +-----------------------------------------+
```

The exports from Module 6 are **educational artifacts** -- they
demonstrate the flow but are not tape-out ready. Production designs
would require characterized cell libraries, full parasitic extraction,
and foundry sign-off.

---

## Three Operating Modes

Module 6 supports three array configurations:

| Mode | Optimized For | Use Case |
|------|---------------|----------|
| **Storage** | Density and retention | Memory arrays |
| **Memory** | Read speed and low power | Cache replacement |
| **Compute** | Throughput and efficiency | Neural network accelerator |

---

## What the Simulator Simplifies

- Placement is rule-based, not fully optimized.
- No parasitic RC extraction in the export path.
- Liberty timing values are placeholders (not characterized).
- No foundry PDK files are bundled -- SKY130 dimensions are presets.

---

## Next Steps

- [PHYSICS.md](PHYSICS.md) -- abstraction levels and mapping model.
- [FEATURES.md](FEATURES.md) -- compiler, exports, and validation.
- [OPENSOURCE-TOOLS.md](OPENSOURCE-TOOLS.md) -- OpenROAD, KLayout, and more.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
