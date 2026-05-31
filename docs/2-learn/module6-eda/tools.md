# Module 6: EDA - Open-Source Tools

## Evidence Status (Demonstrated vs Modeled vs Aspirational)

- **Demonstrated:** Repository structure, export code paths, and generated artifact formats described here are implemented in this repo.
- **Modeled:** Timing/area/power implications and large-scale PPA outcomes depend on downstream tool runs and characterized libraries.
- **Aspirational:** Production tapeout, silicon parity, and foundry acceptance are outside current demonstrated scope.

## Toolchain Components (Open Source)

- **FeCIM Module 6 exporter** (`module6-eda/pkg/export/`): Generates Verilog/DEF/SPICE/JSON/CSV.
- **OpenLane**: Flow orchestration (synthesis → floorplan → place/route → signoff).
- **OpenROAD**: Physical implementation engine used by OpenLane.
- **Yosys**: Synthesis/elaboration for structural Verilog handling.
- **Magic / netgen / KLayout / OpenSTA**: DRC, LVS, layout inspection, timing checks.

## OpenLane/OpenROAD Integration Path (Concrete)

This is the practical path from FeCIM array export to OpenLane execution.

### Step 1 — Generate FeCIM artifacts

Example:

```bash
go run ./cmd/fecim-lattice-tools eda cli \
  -mode compute \
  -input data/sample_weights_8x8.json \
  -rows 8 -cols 8 \
  -name fecim_array_8x8 \
  -output ./output \
  -verilog=true -def=true -spice=true -json=true -csv=true
```

Expected key files:

- `output/fecim_array_8x8.v`
- `output/fecim_array_8x8.def`
- optional cell collateral: `.lef`, `.gds`, `.lib`, blackbox `.v`

### Step 2 — Create OpenLane design folder

```bash
mkdir -p "$OPENLANE_ROOT/designs/fecim_array/src" "$OPENLANE_ROOT/designs/fecim_array/cells"
cp output/fecim_array_8x8.v "$OPENLANE_ROOT/designs/fecim_array/src/"
cp output/fecim_array_8x8.def "$OPENLANE_ROOT/designs/fecim_array/"
# Copy custom cell files when available
# cp cells/fecim_bit.lef cells/fecim_bit.gds cells/fecim_bit.lib cells/fecim_bit.v \
#   "$OPENLANE_ROOT/designs/fecim_array/cells/"
```

### Step 3 — Configure OpenLane injection points

Use `config.json` (or `config.tcl`) in the OpenLane design directory:

```json
{
  "DESIGN_NAME": "fecim_array_8x8",
  "VERILOG_FILES": "dir::src/*.v",
  "SYNTH_ELABORATE_ONLY": 1,

  "FP_DEF_TEMPLATE": "dir::fecim_array_8x8.def",
  "PLACEMENT_CURRENT_DEF": "dir::fecim_array_8x8.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "EXTRA_LIBS": "dir::cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v"
}
```

### Step 4 — Run OpenLane flow

```bash
cd "$OPENLANE_ROOT"
./flow.tcl -design fecim_array
```

Optional interactive debug:

```bash
./flow.tcl -design fecim_array -interactive
```

### Step 5 — Verify handoff integrity before signoff claims

Minimum checks:

1. DEF import succeeded and intended instances are `FIXED` where required.
2. Missing LEF/pin checks pass.
3. DRC/LVS/STA results are recorded from tool logs (not inferred).
4. Any unavailable custom `.lef/.gds/.lib` collateral is documented as a blocker, not hidden.

## Handoff Boundaries (What Module 6 does vs OpenLane/OpenROAD does)

- **Module 6 does:** Array mapping, structural export, placement template emission.
- **OpenLane/OpenROAD does:** Legalization/placement flow control, routing, clocking, signoff automation.
- **Custom cell authoring does:** PDK-legal LEF/GDS/LIB/SPICE definitions and characterization.

## References

- `docs/2-learn/module6-eda/README.md`
- `docs/4-research/validation/policies/eda-trust-boundary.md`
- `module6-eda/pkg/export/`
- `module6-eda/pkg/compiler/`
