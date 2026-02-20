# Example 03: OpenLane Integration

Educational workflow for integrating a FeCIM crossbar macro with OpenLane.

## Overview

This example shows the path from weights -> Verilog/DEF -> OpenLane (optional). Outputs are illustrative and not signoff-grade.

**Prerequisites (Optional):**
- OpenLane installed
- SKY130 PDK configured
- Docker or native OpenLane setup

## Directory Structure

```
03-openlane-integration/
|-- README.md
|-- weights.json           # 16x16 test weights
|-- config.json            # OpenLane configuration
|-- hooks/                 # Optional post-run checks
`-- run_compile.sh         # Generates Verilog/DEF
```

## Step 1: Compile Weights

```bash
cd module6-eda
./examples/03-openlane-integration/run_compile.sh
```

This creates:
- `output/crossbar.v`
- `output/crossbar.def`

## Step 2: Prepare OpenLane Design (Optional)

```bash
mkdir -p ~/OpenLane/designs/fecim_crossbar/src
mkdir -p ~/OpenLane/designs/fecim_crossbar/cells

# Copy generated files
cp examples/03-openlane-integration/output/* ~/OpenLane/designs/fecim_crossbar/src/
cp examples/03-openlane-integration/config.json ~/OpenLane/designs/fecim_crossbar/
cp -r examples/03-openlane-integration/hooks ~/OpenLane/designs/fecim_crossbar/
```

**Cell Views Required:**
`config.json` references `cells/fecim_bit.lef`, `cells/fecim_bit.lib`, and `cells/fecim_bit.v`. You must supply these (stub or real). The repo includes stub LEF templates in `module6-eda/cells/`.

## Step 3: Run OpenLane (Optional)

```bash
cd ~/OpenLane
make mount
./flow.tcl -design fecim_crossbar -tag v1
```

## Config Snippet

```json
{
  "DESIGN_NAME": "fecim_crossbar_16x16",
  "VERILOG_FILES": "dir::src/crossbar.v",
  "FP_DEF_TEMPLATE": "dir::src/crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1
}
```

## Notes

- OpenLane integration is optional and intended for exploration only.
- Generated artifacts are **not** fabrication-ready.
