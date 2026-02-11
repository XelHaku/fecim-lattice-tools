# Module 6: FeCIM Design Suite

**Universal EDA Tool for Ferroelectric Compute-in-Memory Chip Design**

Generate physical chip layouts for FeCIM arrays compatible with OpenLane/OpenROAD flows (educational, not signoff-ready).

## Overview

**Note:** The demo defaults to 30 conductance levels. This is a configurable simulation baseline, not a hardware claim.

The FeCIM Design Suite is a universal chip design tool supporting three distinct FeCIM operation modes:

| Mode | Application | Description |
|------|-------------|-------------|
| **Storage** | Storage-oriented | Retention/endurance focused settings |
| **Memory** | Memory-oriented | Access/bandwidth focused settings |
| **Compute** | AI Accelerator | Analog compute-in-memory for neural network inference |

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FeCIM Design Suite                                │
├────────────────────┬────────────────────┬────────────────────────────┤
│   Storage Mode     │   Memory Mode      │   Compute Mode             │
│   ─────────────    │   ───────────      │   ────────────             │
│   Storage-oriented │   Memory-oriented  │   AI accelerator           │
│   No weights       │   No weights       │   Weights optional         │
│   Configurable     │   Configurable     │   Analog MVM               │
└────────────────────┴────────────────────┴────────────────────────────┘
                              │
                              ▼
                 ┌─────────────────────────┐
                 │   Generated Outputs     │
                 │   - Verilog netlist     │
                 │   - DEF placement       │
                 │   - SPICE netlist       │
                 │   - JSON/CSV data       │
                 └─────────────────────────┘
```

## Quick Start

```bash
# Launch unified app
./launch.sh

# Then select "EDA" tab from the main interface
```

The EDA demo is integrated into the unified FeCIM Visualizer application. Access it through the main tab interface.

## Architecture: 7-Tab Interface

| Tab | Name | Status | Purpose |
|-----|------|--------|---------|
| 1 | **Configure** | Implemented | Array design parameters (mode, size, peripherals) |
| 2 | **Layout** | Implemented | Visual crossbar grid with cell placement |
| 3 | **HDL** | Implemented | Verilog netlist + DEF placement generation |
| 4 | **Explorer** | Placeholder | Design space analysis (area/power/speed) |
| 5 | **Simulate** | Placeholder | ngspice simulation bridge |
| 6 | **Export** | Implemented | Multi-format output (JSON, CSV, SPICE, Verilog, DEF) |
| 7 | **Learn** | Implemented | Interactive OpenLane/OpenROAD documentation |

---

## Tab Details

### Tab 1: Configure Array

Configure FeCIM array parameters for your target application.

**Mode Selection:**

```
┌──────────────────────────────────────────────────────────┐
│ MODE SELECTION                                           │
├─────────────┬────────────────┬────────────────────────────┤
│  Storage    │    Memory      │    Compute                │
│  ─────────  │    ──────      │    ───────                │
│  Storage    │    Memory      │    AI Accelerator         │
│  No weights │    No weights  │    Weights optional       │
│  Levels cfg │    Levels cfg  │    Analog MVM             │
└─────────────┴────────────────┴────────────────────────────┘
```

**Configuration Steps:**
1. **Select mode:** Storage / Memory / Compute
2. **Set array size:** rows × cols (e.g., 256×256)
3. **Choose technology:** SKY130 / GF180MCU / IHP_SG13G2
4. **Select architecture:** passive or 1T1R
5. **Configure peripherals:** DAC bits, ADC bits, TIA gain
6. **[Compute only]** Optionally load pre-trained weights

**Key Point:** For Storage and Memory modes, **NO weights are needed**.
FeFET cells are rewritable — data is programmed during device operation.

**Inputs:**
- Operation mode (Storage, Memory, or Compute)
- Array dimensions (rows × cols)
- Technology selection (SKY130, GF180MCU, IHP_SG13G2)
- Architecture (passive or 1T1R)
- Conductance levels (default: 30)
- Conductance range (G_min, G_max in μS)
- [Compute only] Optional weight matrix for pre-programming

**Outputs:**
- Cell assignments with level, conductance, and resistance
- Design statistics (area, power, throughput)
- For compute with weights: quantization metrics (PSNR, MSE)

**Key Formulas (Compute mode with weights):**
```
Quantization:  level = round((weight + maxWeight) / (2 * maxWeight) × (Levels-1))
Conductance:   G = G_min + (level / (Levels-1)) × (G_max - G_min)  [μS]
Resistance:    R = 1e6 / G  [Ω]
```

### Tab 2: Layout

Interactive crossbar grid visualization.

- **Color coding:** Blue (low G) → Red (high G)
- **Click any cell** to view: row, col, weight, level, conductance, voltage
- **Zoom/pan** for large arrays (128×128+)

### Tab 3: HDL (Verilog + DEF)

Generates hardware description files for OpenLane integration.

**Verilog Output:**
- Structural netlist instantiating FeCIM cells
- Module ports for wordlines (WL), bitlines (BL), and sense lines (SL)
- Compatible with Yosys synthesis (elaborate-only mode)

**DEF Output:**
- Cell placement with FIXED or PLACED keywords
- Row-major ordering with configurable pitch
- Ready for OpenLane's `PLACEMENT_CURRENT_DEF` injection

**Architecture Support:**
- **Passive crossbar:** Simple resistive network
- **1T1R:** Transistor-gated cells for sneak path mitigation

### Tab 6: Export

Multi-format export for different toolchains:

| Format | Extension | Use Case |
|--------|-----------|----------|
| JSON | `.json` | Full mapping with statistics, version control |
| CSV | `.csv` | Spreadsheet analysis, data science |
| SPICE | `.sp` | ngspice/HSPICE simulation |
| Verilog | `.v` | OpenLane synthesis/elaboration |
| DEF | `.def` | OpenLane placement injection |

### Tab 7: Learn

Interactive OpenLane documentation covering:

- **Digital flow stages:** Synthesis → Floorplan → Placement → CTS → Routing → Signoff
- **Tool descriptions:** Yosys, OpenROAD, Magic, KLayout, netgen
- **Configuration variables:** EXTRA_LEFS, EXTRA_GDS_FILES, CURRENT_DEF
- **Custom cell integration:** How to add FeCIM cells to SKY130 PDK

---

## OpenLane Integration

The FeCIM Design Suite generates files compatible with OpenLane v1.0+ flow.

### Integration Strategy

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenLane Flow                            │
├─────────────────────────────────────────────────────────────┤
│  1. Synthesis (Yosys)                                       │
│     └─ SYNTH_ELABORATE_ONLY=1 for structural netlists       │
│                                                             │
│  2. Floorplan                                               │
│     └─ FP_DEF_TEMPLATE: Use our DEF for die area/pins       │
│                                                             │
│  3. Placement                                               │
│     └─ PLACEMENT_CURRENT_DEF: Inject pre-placed DEF ─────┐  │
│     └─ PL_SKIP_INITIAL_PLACEMENT=1                       │  │
│                                              ┌───────────┘  │
│  4. CTS → 5. Routing → 6. Signoff            │              │
│                                              │              │
└──────────────────────────────────────────────│──────────────┘
                                               │
                          ┌────────────────────┘
                          │
              ┌───────────▼───────────┐
              │  FeCIM Design Suite   │
              │  ┌─────────────────┐  │
              │  │ DEF Generator   │  │
              │  │ - FIXED cells   │  │
              │  │ - 1T1R layout   │  │
              │  └─────────────────┘  │
              └───────────────────────┘
```

### Key Configuration Variables

```tcl
# In OpenLane config.json or config.tcl:

# Custom cell definitions
"EXTRA_LEFS": "/path/to/fecim_cell.lef",
"EXTRA_GDS_FILES": "/path/to/fecim_cell.gds",
"EXTRA_LIBS": "/path/to/fecim_cell.lib",

# Use FeCIM DEF as template
"FP_DEF_TEMPLATE": "/path/to/fecim_crossbar.def",

# Or inject at placement stage
"PLACEMENT_CURRENT_DEF": "/path/to/fecim_crossbar.def",
"PL_SKIP_INITIAL_PLACEMENT": 1,

# For structural netlists
"SYNTH_ELABORATE_ONLY": 1,
"VERILOG_FILES_BLACKBOX": "/path/to/fecim_cell.v"
```

See [eda.integration.md](./integration.md) for detailed OpenLane integration guide.

---

## Key Concepts

### Default Quantization (Configurable)

The demo uses a 30-level baseline by default. You can change the number of levels in the configuration:

```
Level 0  → G_min (lowest conductance, highest resistance)
Level 15 → G_mid (middle state)
Level 29 → G_max (highest conductance, lowest resistance)
```

### DEF File Format

The DEF (Design Exchange Format) output uses:

- **FIXED:** Cells that placement tools must not move
- **PLACED:** Cells that may be adjusted during optimization

```def
COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 0 0 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 460 0 ) N ;
  ...
END COMPONENTS
```

### Verilog Netlist

Structural netlist instantiating FeCIM cells:

```verilog
module fecim_crossbar_8x8 (
    input  [7:0] WL,    // Wordlines
    output [7:0] BL,    // Bitlines
    inout  VPWR,
    inout  VGND
);
    fecim_bit cell_0_0 (.WL(WL[0]), .BL(BL[0]), .VPWR(VPWR), .VGND(VGND));
    fecim_bit cell_0_1 (.WL(WL[0]), .BL(BL[1]), .VPWR(VPWR), .VGND(VGND));
    // ...
endmodule
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [eda.integration.md](./integration.md) | OpenLane integration guide |
| [plan-demo6.md](./demo.md) | Implementation plan with code templates |
| [eda.opensource.md](../ecosystem/opensource-eda.md) | Open-source EDA ecosystem analysis |
| [eda.eli5.md](./eli5.md) | Beginner-friendly EDA explanation |
| [README.md](../README.md) | Module 6 overview with disclaimers |

---

## Roadmap

### Implemented
- [x] **Three operation modes** (Storage, Memory, Compute)
- [x] Weight-to-conductance compiler (compute mode)
- [x] Array design generation (all modes)
- [x] Visual crossbar layout
- [x] Verilog/DEF generation
- [x] Multi-format export (JSON, CSV, SPICE)
- [x] OpenLane documentation (Learn tab)

### In Progress
- [ ] OpenLane flow integration testing
- [ ] Custom FeCIM cell LEF/GDS (Magic layout)
- [ ] Liberty timing model generation

### Planned
- [ ] Design space explorer (area/power/throughput estimation)
- [ ] ngspice simulation bridge
- [ ] Automated DRC/LVS validation
- [ ] Multi-layer stacked crossbar support

---

## Related Resources

- [FeCIM Design Suite Examples](../../../module6-eda/examples/) - Sample designs and test cases
- [OpenLane Documentation](https://openlane.readthedocs.io/) - Official OpenLane resources
- [SKY130 PDK Guide](../pdk/sky130.md) - SkyWater 130nm process integration

---

**Part of the FeCIM Lattice Tools educational suite** - See [../../README.md](../../README.md) for project overview.
