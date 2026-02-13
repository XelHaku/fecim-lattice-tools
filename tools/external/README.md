# External Open-Source Tool Inventory

This document tracks external open-source tooling used for FeCIM validation and EDA flows.

| Tool | Version Pin | License | Install Method | FeCIM Use-Case |
|---|---:|---|---|---|
| Heracles (Univ. Groningen) | `v0.4.0` (pin in validation docs; binary optional) | MIT (project-reported) | Build from source (`go build`) or release binary when available | Ferroelectric compact-model reference for P-E loop comparison harness |
| CrossSim (Sandia) | `v3.1.0` | BSD-3-Clause | Python package (`pip install cross-sim==3.1.0`) | Crossbar array behavioral comparison (trend-level and architecture sensitivity) |
| ngspice | `42` | BSD-3-Clause | distro pkg (`apt install ngspice`) / source tarball | SPICE netlist round-trip and sanity simulation for Module 6 exports |
| Icarus Verilog (`iverilog`) | `12.0` | GPL-2.0-or-later | distro pkg (`apt install iverilog`) | RTL syntax/simulation checks for exported Verilog |
| Verilator | `5.028` | LGPL-3.0-or-later OR Artistic-2.0 | distro pkg (`apt install verilator`) / source build | Fast lint/simulation for generated digital wrappers |
| OpenROAD | `v2.0-2025.02` | BSD-3-Clause | OpenROAD-flow-scripts / source build | Place-and-route validation path for FeCIM digital integration |
| OpenLane2 | `v2.1.2` | Apache-2.0 | Python package (`pip install openlane==2.1.2`) or container | End-to-end open-source RTL→GDS flow orchestration |
| Python scientific stack (`numpy`, `scipy`) | `numpy==2.1.3`, `scipy==1.14.1` | BSD-3-Clause | `pip install numpy==2.1.3 scipy==1.14.1` | Helper scripts for curve fitting, interpolation, and metric post-processing |
| Go toolchain | `1.24.x` | BSD-style | official tarball / distro | Build and test FeCIM source + validation harnesses |

## Notes

- **Required tools for CI/toolchain checks**: `go`, `ngspice`.
- **Optional tools**: Heracles, CrossSim, iverilog, Verilator, OpenROAD, OpenLane2, Python stack.
- Version pins are explicit to keep comparator output reproducible across environments.
