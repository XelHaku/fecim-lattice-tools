# Module 6: FeCIM EDA Design Suite

**Purpose:** Generate OpenLane-compatible EDA files for FeCIM crossbar arrays
**Status:** Educational/Research Tool (Work In Progress)
**Last Updated:** 2026-01-29

---

## Documentation Index

### Core Documentation
| Document | Description |
|----------|-------------|
| [API.md](./API.md) | Complete API reference for Module 6 packages |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Detailed architecture and design patterns |
| [WORKFLOW.md](./WORKFLOW.md) | End-to-end RTL to GDSII workflow guide |

### Guides
| Document | Description |
|----------|-------------|
| [guides/eli5.md](./guides/eli5.md) | Beginner-friendly EDA explanation |
| [guides/demo.md](./guides/demo.md) | Demo walkthrough |
| [guides/integration.md](./guides/integration.md) | OpenLane integration guide |
| [guides/zero-to-asic.md](./guides/zero-to-asic.md) | Zero to ASIC practical field guide |
| [guides/spice-format.md](./guides/spice-format.md) | SPICE netlist format guide |
| [guides/fecim-to-wafer.md](./guides/fecim-to-wafer.md) | Complete fabrication workflow |

### References
| Document | Description |
|----------|-------------|
| [references/scientific.md](./references/scientific.md) | Scientific references and citations |
| [references/research-papers.md](./references/research-papers.md) | Research paper collection |
| [references/openlane-study.md](./references/openlane-study.md) | OpenLane source analysis |
| [references/cli-reference.md](./references/cli-reference.md) | OpenLane CLI tool reference |

### PDK Documentation
| Document | Description |
|----------|-------------|
| [pdk/sky130.md](./pdk/sky130.md) | SkyWater 130nm PDK integration |

### Ecosystem
| Document | Description |
|----------|-------------|
| [ecosystem/opensource-eda.md](./ecosystem/opensource-eda.md) | Open-source EDA tools overview |

---

## What This Module Does

Module 6 is an **array builder** that generates EDA file formats compatible with the open-source OpenLane RTL-to-GDSII flow. It does NOT:
- Compile neural network weights (that's conceptual, not implemented)
- Generate validated FeFET device models
- Produce fabrication-ready designs

### Capabilities (Implemented)

| Tab | Function | Output |
|-----|----------|--------|
| 1. Cell Builder | Define FeCIM bitcell dimensions | LEF, Liberty (.lib), Verilog |
| 2. Array Builder | Configure crossbar array size | Array parameters |
| 3. Verilog Export | Generate array netlist | Verilog module |
| 4. DEF Export | Generate placement file | DEF with cell instances |
| 5. Validation | Syntax checking | Yosys validation results |
| 6. Learn | OpenLane tutorial | Educational content |
| 7. Export All | Batch export | All files for OpenLane |

### What Gets Generated

```
output/
├── fecim_bitcell.lef       # Cell abstract (dimensions, pins)
├── fecim_bitcell.lib       # Timing library (PLACEHOLDER VALUES)
├── fecim_bitcell.v         # Behavioral Verilog (pass-through only)
├── fecim_array_NxM.v       # Array instantiation
├── fecim_array_NxM.def     # Placement definition
└── config.json             # OpenLane configuration
```

---

## Critical Disclaimers

### 1. Placeholder Timing Values

All Liberty (.lib) timing parameters are **placeholders**, not characterized values:

```
rise_time: 0.1 ns      <- PLACEHOLDER (not from simulation)
fall_time: 0.1 ns      <- PLACEHOLDER
input_cap: 0.002 pF    <- PLACEHOLDER
leakage:   0.001 nW    <- PLACEHOLDER
```

**Real FeFET characterization requires:**
- SPICE simulation with validated Verilog-A models
- Silicon measurements from test chips
- Liberty characterization flow (e.g., Liberate, OpenSTA)

### 2. Behavioral Model Limitations

The generated Verilog is a **pass-through model** only:

```verilog
// Generated code - does NOT model FeFET physics
module fecim_bitcell (input WL, input BL, output Q);
  assign Q = WL & BL;  // Simplified logic, NOT real behavior
endmodule
```

**What it doesn't model:**
- Polarization states (Pr, Ps)
- Hysteresis (Preisach model)
- 30-level analog states
- Retention, endurance, drift

### 3. No Physical Layout

LEF defines an **abstract view** only (bounding box + pins). There is no:
- Actual transistor layout
- DRC-clean geometry
- LVS-verifiable netlist

### 4. Cell Dimensions

Default dimensions based on SKY130 standard cells:

| Parameter | Value | Source |
|-----------|-------|--------|
| Cell Width | 0.46 um | SKY130 unithd site width |
| Cell Height | 2.72 um | SKY130 standard cell height |
| Site Name | unithd | SKY130 LEF specification |

---

## Summary

**Module 6 is an educational tool** that demonstrates how FeCIM arrays could integrate with open-source EDA flows. It generates syntactically valid files but:

- Uses **placeholder** timing values
- Provides **abstract** cell representations only
- Does **not** model actual FeFET physics
- Is **not** validated for fabrication

For production use, consult the references and work with foundry partners who support ferroelectric processes.

---

*This document aims to be honest about capabilities and limitations.*
