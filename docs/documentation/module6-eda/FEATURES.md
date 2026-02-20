<!-- Category: Features | Module: module6-eda | Reading time: ~4 min -->
# Module 6 Features: EDA Compiler and Export

> Feature inventory for the electronic design automation module.

---

## Compiler

| Feature | Status |
|---------|--------|
| Weight matrix partitioning into tiles | IMPLEMENTED |
| Configurable array dimensions | IMPLEMENTED |
| Storage / Memory / Compute mode selection | IMPLEMENTED |
| Conductance quantization and mapping | IMPLEMENTED |
| Intermediate representation (JSON IR) | IMPLEMENTED |
| Physical coordinate assignment | IMPLEMENTED |

---

## Export Formats

| Format | Status | Notes |
|--------|--------|-------|
| SPICE netlist (.sp) | IMPLEMENTED | Cell-as-resistor model |
| Verilog RTL (.v) | IMPLEMENTED | Structural hierarchy |
| DEF placement (.def) | IMPLEMENTED | Grid-aligned, FIXED cells |
| LEF macro (.lef) | IMPLEMENTED | Cell abstractions |
| Liberty (.lib) | IMPLEMENTED | Timing values are placeholders |
| JSON | IMPLEMENTED | Data interchange |
| CSV | IMPLEMENTED | Data interchange |
| SVG | IMPLEMENTED | Visual layout (via pkg/export API) |

---

## Validation

| Feature | Status |
|---------|--------|
| Basic DRC (design rule checks) | IMPLEMENTED |
| Yosys elaboration check | IMPLEMENTED |
| DEF structural checks | IMPLEMENTED |
| Cross-file consistency validation | IMPLEMENTED |

---

## GUI

| Feature | Status |
|---------|--------|
| Array configuration panel | IMPLEMENTED |
| Mode selection (storage/memory/compute) | IMPLEMENTED |
| Architecture selection (passive/1T1R/2T1R) | IMPLEMENTED |
| Export workflow with file output | IMPLEMENTED |
| Learning visuals and flow diagrams | IMPLEMENTED |

---

## CLI

| Feature | Status |
|---------|--------|
| Compute mode export | IMPLEMENTED |
| Storage mode export | IMPLEMENTED |
| Configurable rows/cols/name | IMPLEMENTED |
| JSON/CSV/SPICE/Verilog/DEF export flags | IMPLEMENTED |
| Passive and 1T1R architecture | IMPLEMENTED |
| 2T1R via API only | PARTIAL |

---

## PDK Integration

| Feature | Status |
|---------|--------|
| SKY130 preset dimensions | IMPLEMENTED |
| Custom PDK configuration (JSON) | IMPLEMENTED |
| Foundry PDK file bundling | NOT INCLUDED |

PDK support refers to preset dimensions and labels. No actual foundry
PDK files are bundled -- the user must supply their own SKY130
installation for downstream tools.

---

## Integration Points

| Integration | Status |
|-------------|--------|
| Module 2 -> Module 6: Crossbar array source | IMPLEMENTED |
| Module 4 -> Module 6: Timing/power back-annotation | PLANNED |
| Module 6 -> OpenLane: Design handoff | IMPLEMENTED (manual) |

---

## Known Limitations

- Liberty timing values are placeholders (need characterization).
- Exports are educational artifacts, not tape-out ready.
- No parasitic extraction in the export path.
- CLI supports passive/1T1R; 2T1R is API-only.

---

## Where It Lives in Code

| Path | Purpose |
|------|---------|
| `module6-eda/pkg/compiler/compiler.go` | Compiler |
| `module6-eda/pkg/compiler/tiling.go` | Tiling |
| `module6-eda/pkg/export/*.go` | All export formats |
| `module6-eda/pkg/validation/*.go` | DRC and validation |
| `module6-eda/pkg/gui/app.go` | GUI entry point |

---

## Status Legend

- **IMPLEMENTED**: Code exists, tests pass, feature is functional
- **PARTIAL**: Core logic exists but incomplete
- **PLANNED**: Specified, not yet coded
- **NOT INCLUDED**: Intentionally not bundled

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
