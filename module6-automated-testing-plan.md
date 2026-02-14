# Module 6 Automated Testing Plan — Research-Grade EDA Validation

## 1. Purpose

Build a **research-grade** automated test pipeline for Module 6 (`module6-eda`) that validates EDA compiler and export formats at publication quality. Module 6 compiles high-level FeCIM array descriptions to industry-standard formats (SPICE, Verilog, DEF, LEF, Liberty, SVG) for integration with OpenLane/commercial EDA tools.

### Scope

| Grade | Definition | Status |
|-------|------------|--------|
| **Simulation-grade** | Internally consistent, exports valid syntax | Current (77.9% coverage, 326 tests) |
| **Research-grade** | Format-correct, physics-preserving, cross-validated, tool-compatible | Target |

### Current State

| Metric | Value |
|--------|-------|
| Source files | 49 (.go) |
| Test files | 42 (*_test.go) |
| Test functions | 326 |
| Coverage | 77.9% |
| Packages | 10 (all PASS) |

---

## 2. Architecture Overview

### EDA Flow

```
High-Level IR (array spec, materials, topology)
    │
    ▼
Compiler → Netlist IR (cells, nets, instances)
    │
    ├─→ SPICE (.sp) ────────→ Analog simulation (ngspice, HSPICE)
    ├─→ Verilog (.v) ───────→ Logic synthesis (Yosys)
    ├─→ DEF (.def) ─────────→ Floorplan/placement
    ├─→ LEF (.lef) ─────────→ Physical library
    ├─→ Liberty (.lib) ─────→ Timing/power characterization
    ├─→ SVG (.svg) ─────────→ Visual layout
    ├─→ CSV (.csv) ─────────→ Data export
    └─→ JSON (.json) ───────→ Metadata/config
```

### Key Source Files

| File | Domain | Lines |
|------|--------|-------|
| `compiler/compiler.go` | IR → netlist compilation | ~1800 |
| `export/spice.go` | SPICE netlist generation | ~900 |
| `export/verilog.go` | Verilog RTL generation | ~230 |
| `export/def.go` | DEF physical layout | ~520 |
| `export/lef.go` | LEF library cells | ~380 |
| `export/liberty.go` | Liberty timing/power | ~650 |
| `export/svg.go` | SVG visual layout | ~240 |
| `validation/physics_test.go` | Physics model validation | ~400 |
| `openlane/flow.go` | OpenLane TCL generation | ~240 |

---

## 3. Test Plan — 7 Phases

### Phase 1: Compiler Correctness (P0 — foundational)

**Goal**: Verify compiler produces correct netlist IR from high-level specs.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-COMP-01 | `compiler_ir_correctness_test.go` | Net conservation | Every input/output connected, no dangling nets |
| M6-COMP-02 | `compiler_ir_correctness_test.go` | Instance count | N×M array → N×M cell instances + parasitic R/C |
| M6-COMP-03 | `compiler_topology_test.go` | 0T1R/1T1R/2T1R | Correct device count per architecture |
| M6-COMP-04 | `compiler_material_test.go` | Material propagation | Capacitance/resistance from material params |
| M6-COMP-05 | `compiler_determinism_test.go` | Same input → same output | Bit-identical netlists |

**Acceptance**: Net conservation, instance counts exact, deterministic output.

### Phase 2: SPICE Export Validation (P0)

**Goal**: Verify SPICE netlist syntax, physics, and tool compatibility.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-SPICE-01 | `spice_syntax_test.go` | Parser round-trip | Export → parse → compare |
| M6-SPICE-02 | `spice_subcircuit_test.go` | 1T1R subcircuit | Correct FET + FeCap topology |
| M6-SPICE-03 | `spice_capacitance_test.go` | C_fe values | Match material εᵣ × A / d |
| M6-SPICE-04 | `spice_ngspice_test.go` | ngspice compatibility | `.op` analysis runs without error |
| M6-SPICE-05 | `spice_power_test.go` | Power model | P = V × I integrated over time |

**Acceptance**: Valid SPICE syntax, ngspice-compatible, physics parameters correct.

### Phase 3: Verilog & Digital Export (P0)

**Goal**: Verify Verilog RTL syntax and logic correctness.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-VER-01 | `verilog_syntax_test.go` | Yosys parse | `read_verilog` succeeds |
| M6-VER-02 | `verilog_mvm_test.go` | MVM behavioral model | Correct multiply-accumulate |
| M6-VER-03 | `verilog_array_test.go` | Array instantiation | N×M cells → N×M module instances |
| M6-VER-04 | `verilog_ports_test.go` | Port consistency | Input/output widths match spec |

**Acceptance**: Yosys-compatible, behavioral model correct, port widths exact.

### Phase 4: Physical Layout (DEF/LEF) (P1)

**Goal**: Verify DEF/LEF physical correctness and DRC compliance.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-DEF-01 | `def_syntax_test.go` | Parser round-trip | Export → parse → compare |
| M6-DEF-02 | `def_placement_test.go` | Non-overlapping cells | No geometry collisions |
| M6-DEF-03 | `def_routing_test.go` | Net connectivity | All nets routed, no opens |
| M6-DEF-04 | `def_area_test.go` | Die area | Area = N×M × cell_area × routing_overhead |
| M6-LEF-01 | `lef_syntax_test.go` | Parser round-trip | Export → parse → compare |
| M6-LEF-02 | `lef_pins_test.go` | Pin geometry | Pins on routing grid |
| M6-LEF-03 | `lef_obstruction_test.go` | OBS layers | Metal blockages for dense cells |

**Acceptance**: Valid DEF/LEF syntax, no overlaps, routing complete, area correct.

### Phase 5: Timing & Power (Liberty) (P1)

**Goal**: Verify Liberty timing/power models and STA compatibility.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-LIB-01 | `liberty_syntax_test.go` | Parser round-trip | Export → parse → compare |
| M6-LIB-02 | `liberty_timing_test.go` | NLDM tables | 7×7 delay/slew tables populated |
| M6-LIB-03 | `liberty_power_test.go` | Dynamic power | Energy per transition > 0 |
| M6-LIB-04 | `liberty_corners_test.go` | PVT corners | FF/TT/SS × T=-40/25/125°C |
| M6-LIB-05 | `liberty_capacitance_test.go` | Input caps | Match SPICE netlist caps |

**Acceptance**: Valid Liberty syntax, NLDM tables complete, multi-corner coverage.

### Phase 6: Cross-Format Consistency (P1)

**Goal**: Verify same design → different formats agree on physics/topology.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-CROSS-01 | `cross_spice_liberty_test.go` | Capacitance agreement | SPICE C_fe = Liberty input_cap |
| M6-CROSS-02 | `cross_verilog_def_test.go` | Instance count | Verilog modules = DEF components |
| M6-CROSS-03 | `cross_def_lef_test.go` | Cell size | DEF component size = LEF macro size |
| M6-CROSS-04 | `cross_power_test.go` | Energy consistency | SPICE P×t = Liberty energy |

**Acceptance**: All cross-format checks within 1% tolerance.

### Phase 7: OpenLane & Tool Integration (P2)

**Goal**: Verify OpenLane flow compatibility and tool invocation.

| ID | Test File | Tests | Description |
|----|-----------|-------|-------------|
| M6-OL-01 | `openlane_tcl_test.go` | TCL syntax | Valid TCL, no parse errors |
| M6-OL-02 | `openlane_flow_test.go` | Config.tcl generation | All required vars set |
| M6-OL-03 | `openlane_drc_test.go` | DRC clean (if tool available) | 0 violations |

**Acceptance**: Valid TCL, config complete, DRC clean (when tools present).

---

## 4. Automation Scripts

### Fast Gate (CI — < 30s)

```bash
#!/bin/bash
# scripts/run_module6_fast_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M6 Fast Gate ==="
go build ./module6-eda/...
go vet ./module6-eda/...
go test -short -count=1 ./module6-eda/...
echo "=== M6 Fast Gate PASS ==="
```

### Full Gate (Nightly — < 5min)

```bash
#!/bin/bash
# scripts/run_module6_full_gate.sh
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"
echo "=== M6 Full Gate ==="
go build ./module6-eda/...
go vet ./module6-eda/...
go test -race -count=1 ./module6-eda/...
echo "=== M6 Full Gate PASS ==="
```

---

## 5. Evidence Requirements

Every test must produce:
1. **Exact file sizes** for exports (not "it works")
2. **Parser round-trip success** (export → parse → exact match)
3. **Physics values with units** (capacitance in fF, resistance in Ω)
4. **Cross-format agreement** (< 1% delta)

### Reporting Format

```
M6-SPICE-03: C_fe=15.9 fF (expected 15.9 fF from εᵣ=30, A=100nm², d=10nm) — PASS
M6-CROSS-01: SPICE C_fe=15.9 fF, Liberty input_cap=15.9 fF, delta=0.0% — PASS
```

---

## 6. Priority & Timeline

| Phase | Priority | Est. Tests | Description |
|-------|----------|-----------|-------------|
| Phase 1 | P0 | 10 | Compiler correctness — IR validation |
| Phase 2 | P0 | 10 | SPICE export — analog netlist |
| Phase 3 | P0 | 8 | Verilog export — digital RTL |
| Phase 4 | P1 | 14 | DEF/LEF — physical layout |
| Phase 5 | P1 | 10 | Liberty — timing/power |
| Phase 6 | P1 | 8 | Cross-format consistency |
| Phase 7 | P2 | 6 | OpenLane integration |
| **Total** | | **66** | |

Phase 1-3 (P0) are mandatory for research claims.
Phase 4-6 (P1) are required for publication.
Phase 7 (P2) is for toolchain integration.

---

## 7. Cross-Module Validation Chain

```
M1 (Hysteresis) → Material C_fe → M6 (EDA) SPICE capacitor
M2 (Crossbar) → Array topology → M6 (EDA) DEF layout
M4 (Circuits) → Timing model → M6 (EDA) Liberty delays
```

---

## 8. Literature & Standards

- **SPICE**: HSPICE Reference Manual (Synopsys)
- **Verilog**: IEEE 1364-2005 Verilog HDL
- **DEF**: Cadence DEF 5.8 Specification
- **LEF**: Cadence LEF/DEF Language Reference
- **Liberty**: Synopsys Liberty User Guide
- **OpenLane**: efabless/openlane GitHub repo

---

## 9. Coverage Target

**Goal**: 90%+ coverage on compiler and export packages.

**Baseline**: 77.9% overall
- `compiler/`: improve to 90%+
- `export/`: improve to 95%+ (syntax-heavy, deterministic)
- `validation/`: 45.1% → 80%+ (add physics regression tests)
- `openlane/`: 39.8% → 70%+ (add flow integration tests)

---

## 10. Known Gaps (Current State)

| Gap | Impact | Remediation |
|-----|--------|-------------|
| No ngspice round-trip | SPICE netlists untested in real tool | Add M6-SPICE-04 with skip-if-missing |
| No Yosys validation | Verilog syntax untested | Add M6-VER-01 with skip-if-missing |
| Liberty power uncalibrated | Energy values are estimates | Cross-validate with SPICE transient |
| DEF routing stub | Routing topology not validated | Add M6-DEF-03 net connectivity check |
| No multi-corner Liberty | Only TT corner exported | Add M6-LIB-04 FF/SS corner generation |

All gaps addressed in Phases 1-7.
