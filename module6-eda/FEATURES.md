# Module 6: EDA Tools - Features

## Features

- **Three Operation Modes** — Storage (NAND), Memory (DRAM), Compute (AI)
- **8 Export Formats** — JSON, CSV, SPICE, Verilog, DEF, LEF, Liberty, SVG
- **Architecture Support** — Passive (0T1R), 1T1R, 2T1R
- **PDK Support** — SKY130, GF180MCU, IHP_SG13G2
- **OpenLane Integration** — RTL-to-GDSII flow ready
- **Weight Mapping** — Quantize trained weights to 30 conductance levels
- **Validation Tools** — Yosys, DEF checker, cross-file consistency

## Operation Modes

| Mode | Purpose | Key Feature |
|------|---------|-------------|
| **Storage** | NAND replacement | 30 levels = 4.9 bits/cell, ECC support |
| **Memory** | DRAM replacement | Zero refresh, 10ns target access |
| **Compute** | AI accelerator | MVM with optional pre-programmed weights |

## Export Formats

| Format | Tool Target |
|--------|-------------|
| Verilog (.v) | Yosys, synthesis |
| SPICE (.sp) | ngspice, HSPICE |
| DEF (.def) | OpenLane, place-and-route |
| LEF (.lef) | Abstract cell views |
| Liberty (.lib) | STA (placeholder timing) |
| SVG (.svg) | Visual layout |
| JSON/CSV | Data exchange |

## Key Parameters

| Parameter | Value | Notes |
|-----------|-------|-------|
| Levels | 30 | FeCIM standard |
| Conductance | 1-100 µS | Gmin/Gmax range |
| Prog Voltage | 2-5V | Configurable |
| Cell Pitch | 0.46 µm | SKY130 compatible |
| Cell Height | 2.72 µm | Standard cell |
| Max Array | 512×512 | Tested |

## Architecture Comparison

| Arch | Nets | Density | Sneak Paths |
|------|------|---------|-------------|
| Passive | WL, BL | Highest | Susceptible |
| 1T1R | WL, BL, SL | Medium | Mitigated |
| 2T1R | WL, BL, SL, CSL | Lower | Eliminated |

## Limitations

- Liberty timing values are placeholders (need SPICE characterization)
- LEF provides abstract views only (need Magic for tape-out)
- Ready for: Research, algorithm validation
- Not ready for: Production tape-out without characterization
