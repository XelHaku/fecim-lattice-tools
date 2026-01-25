# Module 6: OpenLane Integration Guide

**FeCIM Crossbar Macro Integration with OpenLane RTL-to-GDSII Flow**

---

## Quick Links

- **Detailed Integration Guide:** [eda.integration.md](./eda.integration.md)
- **Demo Interface Guide:** [eda.demo.md](./eda.demo.md)
- **Technical Plan:** [module6-technical-plan.md](module6-technical-plan.md)
- **OpenLane Study Notes:** [OPENLANE_STUDY.md](OPENLANE_STUDY.md)
- **Working Examples:** [module6-eda/examples/](../../module6-eda/examples/)

---

## Overview

Module 6 (FeCIM Design Suite) generates files compatible with OpenLane v1.0, enabling automated integration of FeCIM crossbar macros into the open-source ASIC design flow.

```
┌─────────────────────────────────────────────────────────────────┐
│                   FeCIM → OpenLane Pipeline                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Neural Network     FeCIM Design      OpenLane        GDSII     │
│     Weights    ───►    Suite     ───►   Flow    ───►  Output   │
│                                                                 │
│  weights.json       crossbar.v        synthesis      chip.gds   │
│                     crossbar.def      placement                 │
│                     fecim_bit.lef     routing                   │
│                                       signoff                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

| Requirement | Version | Purpose |
|-------------|---------|---------|
| Go | 1.21+ | Build FeCIM Design Suite |
| OpenLane | 1.0+ | RTL-to-GDSII flow |
| SKY130 PDK | 2024.01+ | Process design kit |
| Docker | 20.10+ | OpenLane containerization (optional) |

---

## Quick Start

### Step 1: Generate Design Files

```bash
cd multilayer-ferroelectric-cim-visualizer/module6-eda

# Compile weights to Verilog + DEF
go run ./cmd/eda-cli \
  -input data/sample_weights_8x8.json \
  -rows 8 -cols 8 \
  -verilog=true \
  -def=true \
  -output ./output
```

### Step 2: Configure OpenLane

Create `config.json` for your design:

```json
{
  "DESIGN_NAME": "fecim_crossbar_8x8",
  "VERILOG_FILES": "dir::src/crossbar.v",
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  "SYNTH_ELABORATE_ONLY": 1,
  "PLACEMENT_CURRENT_DEF": "dir::src/crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "DESIGN_IS_CORE": 0,
  "RUN_CTS": 0
}
```

### Step 3: Run OpenLane

```bash
cd ~/OpenLane
./flow.tcl -design fecim_crossbar -tag v1
```

---

## Key Integration Points

### 1. Pre-Placed DEF Injection

FeCIM crossbar cells must maintain their grid positions. Use `PLACEMENT_CURRENT_DEF` to inject the pre-placed DEF:

```json
"PLACEMENT_CURRENT_DEF": "dir::crossbar.def",
"PL_SKIP_INITIAL_PLACEMENT": 1
```

The DEF uses `FIXED` keyword to lock cell positions:

```def
COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 5000 5000 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 5460 5000 ) N ;
  ...
END COMPONENTS
```

### 2. Custom Cell Integration

FeCIM cells are not part of standard cell libraries. Provide custom definitions:

| Variable | File | Purpose |
|----------|------|---------|
| `EXTRA_LEFS` | `fecim_bit.lef` | Abstract view for placement/routing |
| `EXTRA_GDS_FILES` | `fecim_bit.gds` | Physical layout for GDSII merge |
| `EXTRA_LIBS` | `fecim_bit.lib` | Liberty timing model (optional) |
| `VERILOG_FILES_BLACKBOX` | `fecim_bit.v` | Behavioral model for synthesis |

### 3. Structural Netlist Mode

Since FeCIM Design Suite generates structural Verilog (direct cell instantiation), skip logic synthesis:

```json
"SYNTH_ELABORATE_ONLY": 1
```

---

## Working Examples

### Example 1: Basic 8x8 Crossbar

```bash
cd module6-eda
./examples/01-basic-8x8/run.sh
```

### Example 2: MNIST Layer (32x32)

```bash
./examples/02-mnist-layer/run.sh
```

### Example 3: Full OpenLane Integration

```bash
./examples/03-openlane-integration/run_compile.sh
# Then follow README for OpenLane execution
```

---

## Configuration Reference

### Stage-Specific DEF Variables

| Variable | Stage | Description |
|----------|-------|-------------|
| `PLACEMENT_CURRENT_DEF` | Placement | Inject pre-placed DEF |
| `CTS_CURRENT_DEF` | CTS | DEF for clock tree synthesis |
| `ROUTING_CURRENT_DEF` | Routing | DEF for routing stage |

### Placement Control

| Variable | Default | Description |
|----------|---------|-------------|
| `PL_SKIP_INITIAL_PLACEMENT` | 0 | Skip placement (use pre-placed DEF) |
| `PL_BASIC_PLACEMENT` | 0 | Use basic placement for tiny designs |

### Flow Control

| Variable | Default | Description |
|----------|---------|-------------|
| `RUN_CTS` | 1 | Skip for combinational crossbar |
| `DESIGN_IS_CORE` | 1 | Set to 0 for macros |
| `QUIT_ON_MAGIC_DRC` | 1 | Set to 0 during development |

---

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| "EXTRA_LEFS not found" | Path error | Use absolute paths or `dir::` prefix |
| "Unplaced cells remain" | Missing DEF or LEF | Verify all cells in Verilog have LEF definitions |
| "DRC violations" | Stub cells | Expected with stubs; use real cells for tape-out |
| "LVS mismatch" | Pin name mismatch | Verify Verilog ports match LEF pins exactly |

---

## Further Reading

- **Detailed Integration Guide:** [eda.integration.md](./eda.integration.md)
  - Full configuration variable reference
  - DEF format specification
  - Custom cell file structures (LEF, Liberty)
  - Interactive mode commands
  - Complete troubleshooting guide

- **Demo Interface Guide:** [eda.demo.md](./eda.demo.md)
  - 7-tab interface overview
  - Quick start with unified app
  - Usage examples and workflows

- **OpenLane Study Notes:** [OPENLANE_STUDY.md](OPENLANE_STUDY.md)
  - Validated findings from OpenLane source code
  - Stage-by-stage flow explanation
  - Configuration variable sources

- **Technical Plan:** [module6-technical-plan.md](module6-technical-plan.md)
  - Phase-by-phase implementation roadmap
  - Custom cell design workflow
  - Simulation and validation steps

---

## Support

- **Repository Issues:** https://github.com/XelHaku/multilayer-ferroelectric-cim-visualizer/issues
- **OpenLane Documentation:** https://openlane.readthedocs.io/
- **SKY130 PDK:** https://skywater-pdk.readthedocs.io/
