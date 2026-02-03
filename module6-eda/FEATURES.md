# Module 6: EDA Tools - Features

Educational design-suite for FeCIM arrays (exploration, not signoff).

---

## Features

- **Three Operation Modes** - Storage, Memory, Compute
- **Export Formats** - JSON, CSV, SPICE, Verilog, DEF
- **Programmatic Generators** - LEF, Liberty, SVG via `pkg/export`
- **Architecture Support** - Passive, 1T1R, 2T1R (API); CLI supports passive + 1T1R
- **PDK Support** - SKY130, GF180MCU, IHP_SG13G2
- **OpenLane Integration** - DEF + OpenLane config helpers
- **Validation Tools** - Yosys checks, DEF validation, cross-file consistency
- **GUI** - Builder/validation tab + learning visuals

---

## Key Defaults (From Code)

| Parameter | Default | Notes |
|---|---:|---|
| Levels | 30 | Demo baseline |
| Gmin / Gmax | 10 / 100 uS | Conductance range |
| Vprog | 2.0-5.0 V | Programming window |
| Cell pitch / height | 0.46 / 2.72 um | SKY130 defaults |

---

## Limitations

- Liberty timing values are placeholders (require SPICE characterization).
- Exports are educational artifacts, not tape-out ready.
