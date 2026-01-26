# Module 6: Technical Implementation Plan

**FeCIM Design Suite - From Weights to Silicon**

*Last Updated: 2026-01-23*

---

## Executive Summary

Module 6 (FeCIM Design Suite) bridges the gap between neural network training and physical chip fabrication. The module compiles floating-point weight matrices into 30-level ferroelectric memory cell assignments and generates industry-standard output files (Verilog, DEF, LEF, SPICE) compatible with the OpenLane RTL-to-GDSII flow.

**Goals:**
1. Enable automated translation of trained neural networks to FeCIM hardware
2. Generate OpenLane-compatible design files for tape-out
3. Provide educational visualization of the compilation and placement process

**Key Differentiator:** Unlike existing CIM compilers, this tool specifically targets Dr. external research group's 30-level HfO₂-ZrO₂ ferroelectric technology with validated physical parameters.

---

## Implementation Phases

### Phase 0: Custom FeCIM Cell Design

**Objective:** Create a reference FeFET/1T1R cell in Magic for use with SKY130 PDK.

**Timeline:** 1-2 weeks

**Deliverables:**

| File | Purpose | Tool |
|------|---------|------|
| `fecim_bit.mag` | Magic layout source | Magic VLSI |
| `fecim_bit.lef` | Abstract view for placement | Magic → LEF export |
| `fecim_bit.gds` | Physical layout for fabrication | Magic → GDS export |
| `fecim_bit.lib` | Liberty timing model | Manual / Characterization |
| `fecim_bit.spice` | Transistor-level netlist | Magic extraction |

**Cell Specifications:**

```
┌─────────────────────────────────────────┐
│           FeCIM Bit Cell (1T1R)         │
├─────────────────────────────────────────┤
│  Size: 0.46 μm × 2.72 μm                │
│  Site: unithd (SKY130 compatible)       │
│  Pitch: 460 nm × 2720 nm                │
│                                         │
│  Pins:                                  │
│    WL  - Wordline (input, met1)         │
│    BL  - Bitline (output, met2)         │
│    SL  - Sourceline (inout, met1)       │
│    VDD - Power (met1 rail)              │
│    VSS - Ground (met1 rail)             │
│                                         │
│  Layers Used:                           │
│    li1  - Local interconnect            │
│    met1 - Metal 1 (horizontal)          │
│    met2 - Metal 2 (vertical)            │
│    nwell, diff, poly (transistor)       │
└─────────────────────────────────────────┘
```

**Steps:**

1. **Design cell in Magic:**
   ```bash
   magic -d XR -T sky130A fecim_bit.mag
   ```

2. **Extract SPICE netlist:**
   ```tcl
   extract all
   ext2spice lvs
   ext2spice -o fecim_bit.spice
   ```

3. **Generate LEF:**
   ```tcl
   lef write fecim_bit.lef
   ```

4. **Generate GDS:**
   ```tcl
   gds write fecim_bit.gds
   ```

5. **Run DRC:**
   ```tcl
   drc check
   drc count
   ```

**Validation:**
- [ ] DRC clean (0 errors)
- [ ] LVS clean against schematic
- [ ] Pin names match Verilog model
- [ ] Size fits SKY130 row height

---

### Phase 1: Go Macro Generator (Core Compiler)

**Objective:** Compile neural network weights into FeCIM cell assignments.

**Timeline:** 1 week (MVP complete)

**Architecture:**

```
┌─────────────────────────────────────────────────────────────┐
│                     Compiler Pipeline                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  weights.json ──┬──► Quantizer ──► LevelMapper ──► Placer  │
│                 │        │              │            │      │
│  config.json ───┘        ▼              ▼            ▼      │
│                     30 levels      Conductance    (row,col) │
│                     [0..29]        [1..100 μS]    positions │
│                                                             │
│                         └──────────┬─────────────┘          │
│                                    ▼                        │
│                            CellAssignment[]                 │
│                                    │                        │
│                    ┌───────────────┼───────────────┐        │
│                    ▼               ▼               ▼        │
│                .json            .csv           .spice       │
│                .def              .v            .lef         │
└─────────────────────────────────────────────────────────────┘
```

**Key Data Structures:**

```go
// pkg/compiler/types.go

type CompileConfig struct {
    ArrayRows int     // Target array dimensions
    ArrayCols int
    Levels    int     // Quantization levels (default: 30)
    GMin      float64 // Conductance range [μS]
    GMax      float64
    VDD       float64 // Supply voltage
    CellPitch float64 // Cell-to-cell spacing [nm]
}

type CellAssignment struct {
    Row         int
    Col         int
    Weight      float64 // Original weight
    Level       int     // Quantized level [0..29]
    Conductance float64 // Physical conductance [μS]
    Resistance  float64 // 1/G [Ω]
    VProg       float64 // Programming voltage [V]
}

type CrossbarMapping struct {
    Config CompileConfig
    Cells  []CellAssignment
    Stats  CompilationStats
}

type CompilationStats struct {
    TotalCells      int
    UtilizedCells   int
    Utilization     float64
    PSNR            float64 // Peak signal-to-noise ratio
    LevelHistogram  [30]int
    MaxWeight       float64
    MinWeight       float64
}
```

**Core Algorithm:**

```go
// pkg/compiler/compiler.go

func Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error) {
    // 1. Validate input
    if len(weights) > config.ArrayRows || len(weights[0]) > config.ArrayCols {
        return nil, fmt.Errorf("weights exceed array size")
    }

    // 2. Find weight range for symmetric quantization
    maxAbs := findMaxAbsWeight(weights)

    // 3. Quantize and map each weight
    cells := make([]CellAssignment, 0, len(weights)*len(weights[0]))
    for row, rowWeights := range weights {
        for col, weight := range rowWeights {
            // Symmetric quantization: [-maxAbs, +maxAbs] → [0, Levels-1]
            normalized := (weight + maxAbs) / (2 * maxAbs)
            level := int(math.Round(normalized * float64(config.Levels-1)))
            level = clamp(level, 0, config.Levels-1)

            // Map level to physical conductance
            conductance := config.GMin +
                float64(level)/float64(config.Levels-1) * (config.GMax - config.GMin)

            cells = append(cells, CellAssignment{
                Row:         row,
                Col:         col,
                Weight:      weight,
                Level:       level,
                Conductance: conductance,
                Resistance:  1e6 / conductance, // Convert μS to Ω
            })
        }
    }

    // 4. Calculate statistics
    stats := calculateStats(weights, cells, config)

    return &CrossbarMapping{
        Config: config,
        Cells:  cells,
        Stats:  stats,
    }, nil
}
```

**Testing:**

```go
// pkg/compiler/compiler_test.go

func TestCompile_8x8Matrix(t *testing.T) {
    weights := [][]float64{
        {0.5, -0.3, 0.8, -0.1, 0.2, -0.4, 0.6, -0.7},
        // ... 8 rows
    }

    config := DefaultConfig()
    config.ArrayRows = 8
    config.ArrayCols = 8

    mapping, err := Compile(weights, config)
    require.NoError(t, err)

    assert.Equal(t, 64, len(mapping.Cells))
    assert.Equal(t, 64, mapping.Stats.TotalCells)

    // Check quantization bounds
    for _, cell := range mapping.Cells {
        assert.GreaterOrEqual(t, cell.Level, 0)
        assert.LessOrEqual(t, cell.Level, 29)
        assert.GreaterOrEqual(t, cell.Conductance, config.GMin)
        assert.LessOrEqual(t, cell.Conductance, config.GMax)
    }
}

func TestCompile_PSNR(t *testing.T) {
    // PSNR should be high for fine quantization
    weights := generateRandomWeights(32, 32)
    mapping, _ := Compile(weights, DefaultConfig())

    assert.Greater(t, mapping.Stats.PSNR, 30.0) // >30 dB is good
}
```

---

### Phase 2: Export Generators

**Objective:** Generate industry-standard output files for simulation and fabrication.

**Timeline:** 1 week

**Deliverables:**

| Format | File | Purpose | Downstream Tool |
|--------|------|---------|-----------------|
| JSON | `mapping.json` | Full data export | Version control, analysis |
| CSV | `cells.csv` | Spreadsheet analysis | Excel, pandas |
| SPICE | `crossbar.sp` | Analog simulation | ngspice, HSPICE |
| Verilog | `crossbar.v` | Digital simulation | iverilog, Yosys |
| DEF | `crossbar.def` | Physical placement | OpenLane, OpenROAD |
| LEF | `crossbar.lef` | Abstract view | OpenLane, OpenROAD |

**Verilog Generator:**

```go
// pkg/export/verilog.go

func GenerateVerilog(mapping *CrossbarMapping) string {
    var sb strings.Builder

    rows := mapping.Config.ArrayRows
    cols := mapping.Config.ArrayCols

    sb.WriteString(fmt.Sprintf(`// FeCIM Crossbar Array - Auto-generated
// Date: %s
// Size: %dx%d
// Levels: %d

module fecim_crossbar_%dx%d (
    input  wire [%d:0] WL,     // Wordlines
    output wire [%d:0] BL,     // Bitlines
    inout  wire VDD,
    inout  wire VSS
);

`, time.Now().Format("2006-01-02"),
        rows, cols, mapping.Config.Levels,
        rows, cols,
        rows-1, cols-1))

    // Instantiate cells
    for _, cell := range mapping.Cells {
        sb.WriteString(fmt.Sprintf(
            "    fecim_bit cell_%d_%d (\n"+
            "        .WL(WL[%d]),\n"+
            "        .BL(BL[%d]),\n"+
            "        .VDD(VDD),\n"+
            "        .VSS(VSS)\n"+
            "    );  // Level: %d, G: %.2f μS\n\n",
            cell.Row, cell.Col, cell.Row, cell.Col,
            cell.Level, cell.Conductance))
    }

    sb.WriteString("endmodule\n")
    return sb.String()
}
```

**DEF Generator:**

```go
// pkg/export/def.go

func GenerateDEF(mapping *CrossbarMapping) string {
    var sb strings.Builder

    rows := mapping.Config.ArrayRows
    cols := mapping.Config.ArrayCols
    pitch := int(mapping.Config.CellPitch) // in nm, convert to DEF units

    // DEF header
    sb.WriteString(`VERSION 5.8 ;
DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;
`)
    sb.WriteString(fmt.Sprintf("DESIGN fecim_crossbar_%dx%d ;\n", rows, cols))
    sb.WriteString("UNITS DISTANCE MICRONS 1000 ;\n\n")

    // Die area (with margin)
    dieWidth := cols * pitch + 10000  // 10μm margin
    dieHeight := rows * 2720 + 10000  // 2.72μm cell height + margin
    sb.WriteString(fmt.Sprintf("DIEAREA ( 0 0 ) ( %d %d ) ;\n\n", dieWidth, dieHeight))

    // Components
    sb.WriteString(fmt.Sprintf("COMPONENTS %d ;\n", len(mapping.Cells)))
    for _, cell := range mapping.Cells {
        x := 5000 + cell.Col*pitch  // 5μm offset from edge
        y := 5000 + cell.Row*2720
        sb.WriteString(fmt.Sprintf(
            "  - cell_%d_%d fecim_bit + FIXED ( %d %d ) N ;\n",
            cell.Row, cell.Col, x, y))
    }
    sb.WriteString("END COMPONENTS\n\n")

    // Pins
    sb.WriteString(fmt.Sprintf("PINS %d ;\n", rows+cols))
    for i := 0; i < rows; i++ {
        sb.WriteString(fmt.Sprintf(
            "  - WL[%d] + NET WL[%d] + DIRECTION INPUT + USE SIGNAL\n"+
            "    + LAYER met3 ( -140 0 ) ( 140 400 )\n"+
            "    + FIXED ( 0 %d ) N ;\n",
            i, i, 5000+i*2720+1360))
    }
    for i := 0; i < cols; i++ {
        sb.WriteString(fmt.Sprintf(
            "  - BL[%d] + NET BL[%d] + DIRECTION OUTPUT + USE SIGNAL\n"+
            "    + LAYER met4 ( 0 -140 ) ( 400 140 )\n"+
            "    + FIXED ( %d %d ) N ;\n",
            i, i, 5000+i*pitch+pitch/2, dieHeight))
    }
    sb.WriteString("END PINS\n\n")

    sb.WriteString("END DESIGN\n")
    return sb.String()
}
```

---

### Phase 3: Simulation Integration

**Objective:** Validate generated designs with ngspice and iverilog.

**Timeline:** 1 week

**ngspice Workflow:**

```bash
# 1. Generate SPICE from Module 6
go run ./cmd/eda-cli -input weights.json -spice=true -output ./sim

# 2. Create testbench
cat > sim/testbench.sp << 'EOF'
* FeCIM Crossbar Testbench
.include crossbar.sp
.include fecim_bit.spice

* Power supply
VDD vdd 0 DC 1.8
VSS vss 0 DC 0

* Input vectors (wordline activation)
VWL0 wl0 0 PULSE(0 1.8 0 1n 1n 10n 100n)
VWL1 wl1 0 PULSE(0 1.8 20n 1n 1n 10n 100n)

* Load capacitance on bitlines
CBL0 bl0 0 10f
CBL1 bl1 0 10f

* Transient analysis
.tran 1n 200n

* Measure bitline currents
.measure tran ibl0_peak MAX i(VBL0)
.measure tran ibl1_peak MAX i(VBL1)

.end
EOF

# 3. Run simulation
ngspice -b sim/testbench.sp -o sim/results.log

# 4. Plot results
ngspice -b -c 'source sim/testbench.sp; plot v(bl0) v(bl1)'
```

**iverilog Workflow:**

```bash
# 1. Generate Verilog from Module 6
go run ./cmd/eda-cli -input weights.json -verilog=true -output ./sim

# 2. Create testbench
cat > sim/tb_crossbar.v << 'EOF'
`timescale 1ns/1ps

module tb_crossbar;
    reg [7:0] WL;
    wire [7:0] BL;
    wire VDD = 1'b1;
    wire VSS = 1'b0;

    fecim_crossbar_8x8 uut (
        .WL(WL),
        .BL(BL),
        .VDD(VDD),
        .VSS(VSS)
    );

    initial begin
        $dumpfile("sim/crossbar.vcd");
        $dumpvars(0, tb_crossbar);

        WL = 8'b00000000;
        #10 WL = 8'b00000001;  // Activate row 0
        #10 WL = 8'b00000010;  // Activate row 1
        #10 WL = 8'b00000100;  // Activate row 2
        #10 WL = 8'b11111111;  // Activate all rows
        #10 $finish;
    end
endmodule
EOF

# 3. Compile and run
iverilog -o sim/crossbar_tb sim/tb_crossbar.v sim/crossbar.v sim/fecim_bit.v
vvp sim/crossbar_tb

# 4. View waveforms
gtkwave sim/crossbar.vcd
```

---

### Phase 4: OpenLane Hardening

**Objective:** Integrate FeCIM macro with OpenLane for tape-out.

**Timeline:** 2 weeks

**OpenLane Directory Structure:**

```
designs/fecim_crossbar/
├── config.json           # OpenLane configuration
├── src/
│   └── crossbar.v        # Generated Verilog
├── cells/
│   ├── fecim_bit.lef     # Custom cell abstract
│   ├── fecim_bit.gds     # Custom cell layout
│   ├── fecim_bit.lib     # Timing model
│   └── fecim_bit.v       # Behavioral model
├── placement/
│   └── crossbar.def      # Pre-placed cells
└── hooks/
    └── post_run.py       # Validation script
```

**Configuration (config.json):**

```json
{
  "DESIGN_NAME": "fecim_crossbar_32x32",
  "VERILOG_FILES": "dir::src/crossbar.v",
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",

  "PDK": "sky130A",
  "STD_CELL_LIBRARY": "sky130_fd_sc_hd",

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "EXTRA_LIBS": "dir::cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  "SYNTH_ELABORATE_ONLY": 1,

  "FP_SIZING": "absolute",
  "DIE_AREA": "0 0 200 200",
  "DESIGN_IS_CORE": 0,

  "PLACEMENT_CURRENT_DEF": "dir::placement/crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "FP_PDN_ENABLE_RAILS": 0,
  "FP_PDN_MACRO_HOOKS": "fecim_array vccd1 vssd1 VDD VSS",

  "RUN_CTS": 0,
  "RUN_FILL_INSERTION": 1,

  "QUIT_ON_MAGIC_DRC": 0,
  "QUIT_ON_LVS_ERROR": 0
}
```

**Execution:**

```bash
cd ~/OpenLane

# Full automated flow
./flow.tcl -design fecim_crossbar -tag v1.0

# Or interactive for debugging
./flow.tcl -design fecim_crossbar -interactive
```

**Interactive Session:**

```tcl
% package require openlane
% prep -design fecim_crossbar -tag test -overwrite

# Synthesis (elaborate only)
% run_synthesis

# Floorplan with our die area
% run_floorplan

# Inject pre-placed DEF
% set_def $::env(PLACEMENT_CURRENT_DEF)

# Skip CTS (no clock), go to routing
% run_routing

# Signoff
% run_magic
% run_klayout
% run_lvs
% run_magic_drc

# Generate reports
% generate_final_summary_report
```

---

## Validation Checklist

### Pre-Submission (Before Friday Email)

- [ ] **Code compiles:** `go build ./...`
- [ ] **Tests pass:** `go test ./... -v`
- [ ] **GUI launches:** `go run ./cmd/eda-gui`
- [ ] **CLI works:** `go run ./cmd/eda-cli -input data/sample_weights_8x8.json`
- [ ] **Exports valid:** Open generated .json, .csv, .sp, .v, .def files
- [ ] **Documentation complete:** README, INTEGRATION guide, examples

### Post-Submission (Next Week)

- [ ] **ngspice simulation:** Testbench runs without errors
- [ ] **iverilog simulation:** VCD generated, waveforms correct
- [ ] **OpenLane flow:** Completes without fatal errors
- [ ] **DRC:** Magic reports acceptable violations
- [ ] **LVS:** Netgen shows matching

---

## Risk Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Custom cell DRC violations | High | Medium | Use conservative design rules, iterate |
| OpenLane version mismatch | Medium | High | Pin to v1.0, document version |
| Liberty file accuracy | Medium | Medium | Use placeholder timing, refine later |
| Large array routing congestion | Medium | Medium | Start with 8x8, scale gradually |

---

## Dependencies

| Component | Version | Purpose |
|-----------|---------|---------|
| Go | 1.21+ | Core language |
| Fyne | 2.7.2 | GUI framework |
| OpenLane | 1.0 | RTL-to-GDSII flow |
| SKY130 PDK | 2024.01+ | Process design kit |
| Magic | 8.3.460+ | Layout editing |
| ngspice | 42+ | SPICE simulation |
| iverilog | 12+ | Verilog simulation |
| KLayout | 0.29+ | GDS viewing/DRC |

---

## References

1. OpenLane Documentation: https://openlane.readthedocs.io/
2. SKY130 PDK: https://skywater-pdk.readthedocs.io/
3. Magic VLSI: http://opencircuitdesign.com/magic/
4. Dr. Tour FeFET paper: [Tour Research Group publications]
5. DEF/LEF specification: SI2 OpenAccess

---

## Appendix A: DEF Format Quick Reference

```def
VERSION 5.8 ;
DESIGN <name> ;
UNITS DISTANCE MICRONS <value> ;

DIEAREA ( llx lly ) ( urx ury ) ;

COMPONENTS <count> ;
  - <name> <cell> + <status> ( x y ) <orient> ;
  # status: FIXED | PLACED | UNPLACED
  # orient: N | S | E | W | FN | FS | FE | FW
END COMPONENTS

PINS <count> ;
  - <name> + NET <net> + DIRECTION <dir> + USE <use>
    + LAYER <layer> ( llx lly ) ( urx ury )
    + <status> ( x y ) <orient> ;
END PINS

NETS <count> ;
  - <name> ( <pin1> ) ( <pin2> ) ... ;
END NETS

END DESIGN
```

---

## Appendix B: OpenLane Variables Quick Reference

| Variable | Default | Purpose |
|----------|---------|---------|
| `EXTRA_LEFS` | None | Custom cell LEF files |
| `EXTRA_GDS_FILES` | None | Custom cell GDS files |
| `EXTRA_LIBS` | None | Custom cell Liberty files |
| `VERILOG_FILES_BLACKBOX` | None | Black-box Verilog for synthesis |
| `SYNTH_ELABORATE_ONLY` | 0 | Skip logic mapping |
| `PLACEMENT_CURRENT_DEF` | None | Pre-placed DEF |
| `PL_SKIP_INITIAL_PLACEMENT` | 0 | Skip placement algorithm |
| `FP_DEF_TEMPLATE` | None | Template DEF for die/pins |
| `DESIGN_IS_CORE` | 1 | 0 for macros |

---

*Document Version: 1.0*
*Author: XelHaku*
*Status: Draft for Friday Email*
