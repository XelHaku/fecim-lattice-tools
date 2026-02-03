# Module 6: FeCIM EDA Design Suite

**Purpose:** Generate OpenLane-compatible artifacts for FeCIM crossbar arrays
**Status:** Educational/Research Tool (simulation-only)
**Last Updated:** 2026-02-03

---

## Documentation Index

### Core Documentation
| Document | Description |
|---|---|
| [API.md](./API.md) | API reference for Module 6 packages |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | Package layout and data flow |
| [WORKFLOW.md](./WORKFLOW.md) | End-to-end workflow guide |

### Guides
| Document | Description |
|---|---|
| [guides/eli5.md](./guides/eli5.md) | Beginner-friendly explanation |
| [guides/demo.md](./guides/demo.md) | Demo walkthrough |
| [guides/integration.md](./guides/integration.md) | OpenLane integration guide |
| [guides/zero-to-asic.md](./guides/zero-to-asic.md) | Practical field guide |
| [guides/spice-format.md](./guides/spice-format.md) | SPICE netlist format |
| [guides/fecim-to-wafer.md](./guides/fecim-to-wafer.md) | Fabrication workflow overview |

### References
| Document | Description |
|---|---|
| [references/scientific.md](./references/scientific.md) | Scientific references |
| [references/research-papers.md](./references/research-papers.md) | Research paper collection |
| [references/openlane-study.md](./references/openlane-study.md) | OpenLane source analysis |
| [references/cli-reference.md](./references/cli-reference.md) | OpenLane CLI reference |

### PDK Documentation
| Document | Description |
|---|---|
| [pdk/sky130.md](./pdk/sky130.md) | SkyWater 130nm PDK integration |

---

## What This Module Does

Module 6 is an **array builder** that generates EDA file formats compatible with OpenLane-style flows. It can optionally map weights in compute mode. It does **not**:

- Produce fabrication-ready layouts
- Include validated FeFET compact models
- Replace signoff flows

### Capabilities (Implemented)

- **Array configuration** (size, mode, technology, architecture)
- **Export formats**: JSON, CSV, SPICE, Verilog, DEF
- **Programmatic generators**: LEF, Liberty, SVG
- **Validation**: Yosys checks, DEF validation, cross-file consistency
- **GUI**: Builder & Validation view + Learn view

### What Gets Generated

```
output/
|-- fecim_array_design.json   # Design metadata
|-- fecim_array_cells.csv     # Per-cell assignments
|-- fecim_array.sp            # SPICE netlist
|-- fecim_array.v             # Verilog netlist
`-- fecim_array.def           # DEF placement
```

---

## Critical Disclaimers

### 1) Placeholder Timing Values

Liberty timing parameters are placeholders and require characterization.

### 2) Behavioral Model Limits

Generated Verilog is structural and **does not** model ferroelectric physics.

### 3) No Physical Layout

LEF defines abstract geometry only; there is no DRC/LVS-clean layout.

### 4) Default Cell Dimensions

Defaults are SKY130-style placeholders:

| Parameter | Value | Source |
|---|---:|---|
| Cell Width | 0.46 um | SKY130 unithd site width |
| Cell Height | 2.72 um | SKY130 standard cell height |

---

## Summary

Module 6 is an educational tool for exploring FeCIM array flows. It generates syntactically valid artifacts but is **not** suitable for fabrication without characterization and foundry support.
