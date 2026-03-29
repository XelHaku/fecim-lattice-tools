# Open-Source EDA Tools for FeCIM Chip Design

**Purpose:** Complete reference guide for RTL-to-GDSII open-source EDA tools for designing FeCIM peripheral circuits and digital control logic.

**Audience:** Chip designers, researchers, and students implementing ferroelectric compute-in-memory systems using open-source tools.

**Last Updated:** 2026-01-27

---

## Overview

This document covers the essential open-source EDA (Electronic Design Automation) tools needed to go from Verilog RTL to fabrication-ready GDSII for FeCIM-based chips. The focus is on practical integration with the SkyWater SKY130 130nm process and the Efabless open MPW shuttle program.

### What You Can Build

Using these tools together, you can design:
- **Digital control logic** (FSMs, address decoders, muxes)
- **Analog peripherals** (DAC, ADC, transimpedance amplifiers)
- **Mixed-signal systems** (FeCIM array + surrounding circuits)
- **Complete SoCs** ready for fabrication through Efabless Open MPW

### Key Statistics

- **Active projects using these tools:** 600+ successful tapeouts via OpenLane
- **Open PDKs available:** 3 major processes (SKY130, GF180, IHP130)
- **Free shuttle runs:** Efabless provides free MPW slots (sponsored by Google)
- **Community size:** 6,400+ members in SKY130 community alone

---

## 1. The Complete RTL-to-GDSII Flow

```
Verilog RTL
    ↓
Yosys (Synthesis) → Gate Netlist + Mapped Cells
    ↓
OpenROAD (Place & Route)
    ├─ Floorplanning
    ├─ Placement (cells positioned)
    ├─ Clock Tree Synthesis (CTS)
    └─ Routing (metal layers)
    ↓
Magic (Physical Verification)
    ├─ DRC (Design Rule Checking)
    ├─ Extraction (SPICE netlist from layout)
    └─ Antenna checking
    ↓
Netgen (LVS - Layout vs. Schematic)
    → Verify netlist matches original
    ↓
KLayout (GDSII Viewer/Cleanup)
    ↓
Final GDSII (Ready for Foundry)
```

### Real-World Example: Digital Control FSM

```
control_fsm.v (behavioral Verilog)
    ↓ [Yosys]
→ Control FSM mapped to sky130 cells
    ↓ [OpenROAD]
→ Placed on silicon, routed with metal
    ↓ [Magic]
→ Layout checked against design rules
    ↓ [GDSII]
→ Ready to tape out
```

---

## 2. Core Tools

### 2.1 Yosys - RTL Synthesis

**Repository:** https://github.com/YosysHQ/yosys
**License:** ISC (Open Source)
**Language:** C++
**Maturity:** Production (since 2013)

#### Purpose
Converts behavioral Verilog/SystemVerilog into gate-level netlists using cells from your technology library (SKY130, GF180, etc.).

#### Key Features
- **Multi-language support:** Verilog, SystemVerilog, VHDL (via GHDL integration)
- **Cell mapping:** Automatically selects appropriate standard cells
- **Optimization:** Boolean minimization, gate reduction
- **Technology-agnostic:** Works with any PDK
- **Scripting:** TCL-based flow control

#### Installation

**Ubuntu/Debian:**
```bash
sudo apt-get install yosys
```

**Docker (Recommended):**
```bash
docker pull hpretl/iic-osic-tools:latest
docker run -it hpretl/iic-osic-tools:latest
```

**From Source:**
```bash
git clone https://github.com/YosysHQ/yosys.git
cd yosys
make -j$(nproc)
sudo make install
```

#### Basic Usage

**Simple synthesis script (synth.tcl):**
```tcl
# Read Verilog file
read_verilog control_fsm.v

# Prepare for synthesis
prep -top ControlFSM

# Synthesize to standard cells
synth_sky130 -json synth.json

# Generate netlist
write_netlist -json netlist.json
```

**Run:**
```bash
yosys -m ghdl synth.tcl
```

#### FeCIM-Specific Use Cases

1. **Digital Controllers:** FSM synthesis for row/column addressing
   ```verilog
   // Example: Row selector FSM
   module RowSelector (
     input clk, reset,
     input [3:0] row_addr,
     output reg [31:0] row_select
   );
     // FSM for address decoding
   endmodule
   ```

2. **Mux Arrays:** Multiplexer trees for cell selection
3. **Counters/Dividers:** Timing generation for write pulses

#### Key Parameters

| Parameter | Purpose | FeCIM Relevance |
|-----------|---------|-----------------|
| `synth_sky130 -json` | Output synthesis results as JSON | Integration with OpenROAD |
| `-flatten` | Merge hierarchies for optimization | Improves cell placement |
| `-abc` | Advanced boolean circuit optimization | Better QoR (Quality of Results) |

#### Verification After Synthesis

```bash
# Check for synthesis errors
yosys -m ghdl synth.tcl 2>&1 | grep -i "error\|warning"

# Inspect netlist
cat synth.json | grep '"name"' | head -20
```

---

### 2.2 OpenROAD - Place and Route

**Repository:** https://github.com/The-OpenROAD-Project/OpenROAD
**License:** BSD-3-Clause
**Language:** C++ with Python bindings
**Maturity:** Production (used in 600+ OpenLane tapeouts)

#### Purpose
Handles the physical design phase: taking the gate netlist from Yosys and creating a layout with placement and routing on the silicon.

#### Key Components

| Component | Function |
|-----------|----------|
| **Floorplan** | Define chip area, power distribution |
| **Placement** | Position cells to minimize area/delay |
| **CTS** | Create clock distribution network |
| **Router** | Connect cells with metal wires |
| **Static Timing Analysis** | Verify timing constraints met |

#### Installation

**Docker (Recommended):**
```bash
docker pull hpretl/iic-osic-tools:latest
```

**Ubuntu (via OpenLane package):**
```bash
git clone https://github.com/The-OpenROAD-Project/OpenLane.git
cd OpenLane
make
```

#### Basic Usage

**Standalone flow (Python):**
```python
import openroad as ord

# Create design
db = ord.get_db()
design = ord.get_tech()

# Floorplan
ord.floorplan_tool()
ord.initialize_floorplan(
    die_area=[0, 0, 100, 100],  # μm
    core_area=[10, 10, 90, 90]
)

# Placement
ord.place_tool()
ord.place_cells()

# CTS
ord.cts_tool()
ord.build_cts()

# Routing
ord.route_tool()
ord.route_design()

# Output
ord.write_def("output.def")
```

**TCL script (typical flow):**
```tcl
read_lef sky130.lef
read_lib sky130_library.lib
read_def netlist.def

floorplan -site unithd -density 0.65 -area

place_cells

clock_tree_synthesis -root_buffer sky130_fd_sc_hd__clkbuf_16

route_design

write_def output.def
```

#### FeCIM-Specific Considerations

1. **Custom Macro Placement:** If using hardened FeCIM array
   ```python
   # Place macro at specific location
   macro = db.findInst("fecim_array_macro")
   macro.setLocation(20, 20)  # μm x, y
   ```

2. **Power Distribution:** Critical for mixed-signal designs
   ```tcl
   # Define power domains
   add_power_domain core -power VDD -ground VSS
   add_io_pad VDD -location top
   add_io_pad VSS -location bottom
   ```

3. **Analog-Digital Boundary:** Separation constraints
   ```tcl
   # Keep analog cells away from noisy digital logic
   set_placement_constraint -macro_type ANALOG \
     -loc [list 0 50] -size [list 50 100]
   ```

#### Key Parameters

| Parameter | Default | FeCIM Tuning |
|-----------|---------|--------------|
| Core density | 0.6 (60%) | 0.5-0.65 (lower = more routing space) |
| Power straps width | 0.5 μm | Increase for analog supplies |
| Clock leaf buffer depth | 5 | Increase for lower jitter |

#### Verification

```tcl
# Check timing after CTS
report_timing -path_delay max -digits 3

# Check congestion
report_routing_congestion

# DRC violations before Magic
report_drc -summary
```

---

### 2.3 Magic VLSI - Layout Editor & Verification

**Repository:** https://github.com/RTimothyEdwards/magic
**License:** Open Source
**Language:** C/Tcl
**Age:** Active development since 1983

#### Purpose
Physical layout editor for custom cell design, verification, and extraction. Used for:
- Custom cell design (hand-crafted layouts)
- DRC (Design Rule Checking)
- LVS (Layout vs. Schematic) extraction
- SPICE netlist generation from layout

#### Installation

**Ubuntu/Debian:**
```bash
sudo apt-get install magic
```

**Docker:**
```bash
docker run -it hpretl/iic-osic-tools:latest magic -T sky130A
```

**From Source:**
```bash
git clone https://github.com/RTimothyEdwards/magic.git
cd magic
./configure --prefix=/usr/local
make
sudo make install
```

#### Key Workflow

**1. Open existing cell:**
```bash
magic -T sky130A latch_cell.mag
```

**2. Design custom FeFET cell (command line):**
```tcl
# Create new cell
edit new_cell

# Draw transistors and metal
box 0 0 1 1
paint nmos

# Add metal pins
layer metal1
rect 0 0 0.5 0.5

# Label pin
label input
```

**3. Check Design Rules (DRC):**
```tcl
drc
```

**4. Extract SPICE netlist:**
```tcl
# Define extraction area
box 0 0 10 10

# Extract
extract do local
extract

# Output
ext2spice
```

#### FeCIM Applications

1. **Custom Bitcell Layout:**
   ```tcl
   # Simple ferroelectric capacitor layout
   # (Placeholder - real layout requires PDK support)

   magic -T sky130A

   # Create ferroelectric capacitor
   paint poly
   paint metal1
   # (requires custom ferroelectric layers)
   ```

2. **DRC Verification for Custom Cells:**
   ```bash
   # Check all cells against PDK rules
   for cell in *.mag; do
     magic -T sky130A -Batch -r $cell.drc << EOF
     drc on
     drc area
     EOF
   done
   ```

3. **Extraction for SPICE Simulation:**
   ```tcl
   # Extract realistic parasitic R, C
   extract all
   ext2spice lvs
   ```

#### Key Features

| Feature | Command | Use Case |
|---------|---------|----------|
| Layer painting | `paint nmos` | Cell design |
| DRC checking | `drc on` | Rule verification |
| LVS extraction | `extract` | Netlist comparison |
| Parasitic extraction | `ext2spice` | Accurate simulation |

#### Limitations & FeCIM Considerations

- **Standard PDKs don't include ferroelectric layers** - requires custom layer definitions
- **No physics simulation** - tool is geometric only
- **Manual layout required** for custom ferroelectric structures

**Workaround:** Use Magic for CMOS periphery, add ferroelectric layers via custom design or post-processing.

---

### 2.4 Netgen - LVS Verification

**Repository:** https://github.com/RTimothyEdwards/netgen
**License:** Open Source
**Language:** C/Tcl

#### Purpose
Layout vs. Schematic (LVS) verification: compares the extracted netlist from a layout against the original schematic to ensure they match.

#### Installation

```bash
sudo apt-get install netgen
```

**From Source:**
```bash
git clone https://github.com/RTimothyEdwards/netgen.git
cd netgen
./configure
make
sudo make install
```

#### Basic Usage

**LVS Script (compare.tcl):**
```tcl
#!/usr/bin/env netgen

# Read technology
load sky130A

# Read schematic netlist (from Yosys)
readnet spice schematic.spi toplevel

# Read extracted layout netlist (from Magic)
readnet spice layout.spi toplevel

# Compare
compare

# Generate report
report lvs comparison.rpt
```

**Run:**
```bash
netgen -batch compare.tcl
```

#### Key Output

```
Circuit 1: schematic
Circuit 2: layout

Netlists do not match:
  Cells in Circuit 1: 150
  Cells in Circuit 2: 150
  Pins match: YES
  Properties match: YES

(Detailed discrepancy report)
```

#### FeCIM Integration

For mixed-signal designs:
```tcl
# Define analog blocks to ignore
add_ignore_class resistor
add_ignore_class capacitor
add_ignore_class inductor

# Compare digital logic only
compare
```

---

### 2.5 KLayout - GDSII Viewer & Editor

**URL:** https://www.klayout.de/
**License:** GPL (and commercial)
**Language:** C++ with Python/Ruby scripting

#### Purpose
Industrial-strength layout viewer, editor, and scripter. Final GDSII inspection and cleanup before tape-out.

#### Installation

**Ubuntu/Debian:**
```bash
sudo apt-get install klayout
```

**Docker:**
```bash
docker run -e DISPLAY=$DISPLAY -v /tmp/.X11-unix:/tmp/.X11-unix \
  hpretl/iic-osic-tools:latest klayout
```

#### Key Features

1. **GDSII Viewing & Inspection:**
   ```bash
   klayout output.gds
   ```

2. **Python Scripting:**
   ```python
   import pya

   # Load GDSII
   layout = pya.Layout()
   layout.read("chip.gds")

   # Inspect cell
   cell = layout.find_cell("fecim_array")

   # Count instances
   print(f"Instances: {len(cell.insts())}")
   ```

3. **DRC/LVS Running:**
   ```bash
   # Run DRC rules on GDSII
   klayout -r sky130_drc.lylm -i output.gds -o drc_report.txt
   ```

#### FeCIM Use Cases

1. **Pre-submission GDSII check:**
   ```bash
   klayout -r full_drc.lylm output.gds
   ```

2. **Layer stack inspection:**
   - Verify metal routing on correct layers
   - Check via density
   - Validate power delivery network

3. **Custom layer visualization:**
   ```python
   # Add custom ferroelectric layer colors
   # (requires layer mapping file)
   ```

---

## 3. Process Design Kits (PDKs)

### 3.1 SkyWater SKY130

**GitHub:** https://github.com/google/skywater-pdk
**Process:** 130nm CMOS
**Metal Layers:** 5
**Features:** Most mature, largest community

#### Quick Facts

| Specification | Value |
|---------------|-------|
| **Technology Node** | 130 nm |
| **Transistor Count** | ~5 million/mm² |
| **Metal Layers** | 5 |
| **Vdd** | 1.8 V |
| **Fmax** | ~100 MHz digital |
| **Leakage** | ~1 pW/μm² (standby) |

#### What's Included

1. **Standard Cell Library (sky130_fd_sc_hd)**
   - 600+ logic cells
   - Flip-flops, latches, buffers
   - Characterization: SPICE simulations + Liberty timing

2. **I/O Library (sky130_fd_io)**
   - ESD protection cells
   - Analog I/O pads
   - Biasing circuits

3. **SRAM Compiler**
   - Generates custom SRAM arrays
   - Configurable word size, depth

4. **Analog Cells**
   - Op-amps
   - Comparators
   - Resistor/capacitor arrays

#### Installation

**Minimal (for tool use):**
```bash
export PDK_ROOT=$(pwd)
git clone https://github.com/google/skywater-pdk.git
cd skywater-pdk
make sky130A  # Downloads ~2GB
```

**Full installation (10GB+):**
```bash
make all
```

#### Directory Structure

```
skywater-pdk/
├── sky130A/                # Process variant A (most common)
│   ├── libs.tech/          # Technology files for tools
│   │   ├── magic/          # Magic VLSI setup
│   │   ├── klayout/        # KLayout definitions
│   │   └── xyce/           # Xyce SPICE
│   ├── libs.ref/           # Reference libraries
│   │   ├── sky130_fd_sc_hd/    # Standard cells
│   │   ├── sky130_fd_io/       # I/O cells
│   │   └── sky130_sram_macros/ # SRAM
│   └── doc/                # Documentation
├── sky130B/                # Higher-speed variant
└── rules/                  # Design rules documentation
```

#### FeCIM Integration

For custom ferroelectric cells alongside SKY130:
```tcl
# config.tcl (for OpenLane)
set ::env(PDK) "sky130A"
set ::env(STD_CELL_LIBRARY) "sky130_fd_sc_hd"

# Add custom ferroelectric macro
set ::env(EXTRA_LEFS) "${::env(DESIGN_DIR)}/fecim_array.lef"
set ::env(EXTRA_LIBS) "${::env(DESIGN_DIR)}/fecim_array.lib"
```

---

### 3.2 GlobalFoundries GF180MCU

**GitHub:** https://github.com/google/gf180mcu-pdk
**Process:** 180nm MCU (Mixed-signal Compute Unit)
**Metal Layers:** 6
**Best For:** Analog-heavy designs, higher voltage operation

#### Key Differences from SKY130

| Parameter | SKY130 | GF180MCU |
|-----------|--------|----------|
| **Node** | 130 nm | 180 nm |
| **Vdd (digital)** | 1.8 V | 1.8 V |
| **Vdd (analog)** | 1.8 V | 3.3 V / 5 V |
| **Power density** | Higher | Lower |
| **Max freq** | ~100 MHz | ~50 MHz |

#### Use Cases

- **Higher voltage analog blocks** (DAC, ADC needing wider range)
- **Lower power consumption** (larger transistors, lower leakage)
- **Less aggressive scaling** (better for analog)

#### Installation

```bash
git clone https://github.com/google/gf180mcu-pdk.git
export PDK_ROOT=$(pwd)/gf180mcu-pdk
```

---

### 3.3 IHP SG13G2 (IHP130 BiCMOS)

**GitHub:** https://github.com/IHP-GmbH/IHP-Open-PDK
**Process:** 130nm BiCMOS
**Metal Layers:** 7
**Special Feature:** High-speed analog, bipolar transistors

#### Unique Features

- **Bipolar transistors:** Faster gain-bandwidth than MOSFETs
- **Precision resistors and capacitors:** For analog circuits
- **RRAM support:** Basic resistance RAM elements
- **Academic focus:** Designed for research

#### Best For

- **High-frequency analog circuits** (>1 GHz)
- **Research prototyping** (early-stage devices)
- **RF/Analog co-design**

---

## 4. Integration Platforms

### 4.1 OpenLane - Automated RTL-to-GDSII

**GitHub:** https://github.com/The-OpenROAD-Project/OpenLane
**License:** Apache 2.0
**Maturity:** Production (600+ tapeouts)

#### What Is OpenLane?

OpenLane is an automated RTL-to-GDSII flow that orchestrates all the tools above:
- Yosys → Synthesis
- OpenROAD → Place & Route
- Magic → DRC
- Netgen → LVS
- KLayout → GDS cleanup

#### Installation

**Docker (Recommended - All-in-one):**
```bash
docker pull efabless/openlane:latest
docker run -it -v $(pwd):/openLANE_flow/designs \
  efabless/openlane:latest
```

**Native (Ubuntu 20.04+):**
```bash
git clone https://github.com/The-OpenROAD-Project/OpenLane.git
cd OpenLane
make
make test  # Verify installation
```

#### Basic Flow

**1. Create design directory:**
```bash
mkdir designs/my_fecim_controller
cd designs/my_fecim_controller
mkdir -p src config
```

**2. Create config.tcl:**
```tcl
# config.tcl
set ::env(DESIGN_NAME) "controller"
set ::env(VERILOG_FILES) "./src/controller.v"
set ::env(CLOCK_PORT) "clk"
set ::env(CLOCK_PERIOD) "10"  # 10ns = 100MHz
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 500 500"  # 500x500 μm die
```

**3. Create Verilog:**
```verilog
// src/controller.v
module controller (
  input clk, reset,
  input [3:0] cmd,
  output reg [7:0] addr
);

  always @(posedge clk) begin
    if (reset) addr <= 0;
    else addr <= addr + 1;
  end

endmodule
```

**4. Run OpenLane:**
```bash
flow.tcl -design designs/my_fecim_controller
```

**5. Check results:**
```bash
# Generated outputs
designs/my_fecim_controller/runs/RUN_*/
├── results/
│   ├── synthesis/         # Synthesis reports
│   ├── placement/         # Placement results
│   ├── cts/              # Clock tree
│   ├── routing/          # Final routing
│   └── final/            # Final GDSII
├── reports/              # Detailed logs
└── logs/                 # Tool outputs
```

#### Configuration Parameters for FeCIM

```tcl
# Performance
set ::env(CLOCK_PERIOD) "10"           # 100 MHz
set ::env(SETUP_SLACK_MARGIN) "0.5"    # ns margin

# Area & Power
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 1000 1000"    # μm
set ::env(FP_CORE_UTIL) "0.55"         # 55% utilization

# Routing
set ::env(ROUTING_CORES) "8"           # Parallel routing
set ::env(DETAILED_ROUTER) "tritonRoute"

# Mixed-signal (if used)
set ::env(MACRO_PLACEMENT_CFG) "macro_placement.cfg"
```

#### Full Example: FeCIM Row Decoder

**controller.v:**
```verilog
module fecim_row_decoder (
  input clk, reset,
  input [4:0] row_addr,
  output reg [31:0] row_select
);

  always @(posedge clk or posedge reset) begin
    if (reset)
      row_select <= 32'b0;
    else
      row_select <= (32'b1) << row_addr;
  end

endmodule
```

**config.tcl:**
```tcl
set ::env(DESIGN_NAME) "fecim_row_decoder"
set ::env(VERILOG_FILES) "./src/decoder.v"
set ::env(CLOCK_PORT) "clk"
set ::env(CLOCK_PERIOD) "10"
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 300 300"
set ::env(FP_CORE_UTIL) "0.60"
```

**Run:**
```bash
flow.tcl -design designs/fecim_row_decoder 2>&1 | tee flow.log
```

#### Output Inspection

```bash
# View synthesis report
cat results/synthesis/synthesis.rpt

# View timing
cat results/final/timing_summary.rpt

# View GDSII
klayout results/final/*.gds

# View DEF
gvim results/final/*.def
```

---

### 4.2 Caravel Harness & Efabless Platform

**GitHub:** https://github.com/efabless/caravel
**Platform:** https://platform.efabless.com
**Cost:** FREE (via Google/Efabless Open MPW)

#### What Is Caravel?

Caravel is a "wrapper" or harness that includes:
- **User project area** (~20,000 cells) for custom designs
- **GPIO I/O** (38 pins)
- **Power distribution** (VDD, VSS, VDDA, VSSA)
- **Clock distribution** (40 MHz reference)
- **Wishbone bus** for control interface
- **Built-in tests** and characterization circuits

#### Integration Steps

**1. Clone Caravel:**
```bash
git clone https://github.com/efabless/caravel.git
cd caravel
```

**2. Create user project:**
```bash
# Generate template
python3 venv/bin/python3 -m mcu.magic.gen_project \
  --project-name my_fecim_chip \
  --template simple
```

**3. Implement your design:**
```bash
# Replace user_proj_example with your controller
cp my_design.v openlane/user_proj_example/src/

# Update config
cat > openlane/user_proj_example/config.tcl << EOF
set ::env(DESIGN_NAME) "my_fecim_chip"
set ::env(VERILOG_FILES) "./src/my_design.v"
...
EOF
```

**4. Build user project:**
```bash
cd openlane
make user_proj_example
```

**5. Integrate into Caravel:**
```bash
make all  # Builds full chip
```

**6. Submit to Efabless:**
```bash
# Follow submission guidelines
# https://platform.efabless.com/
```

#### Caravel Block Diagram

```
┌─────────────────────────────────────────┐
│           Caravel Harness               │
├─────────────────────────────────────────┤
│  ┌──────────────────────────────────┐   │
│  │  Your FeCIM Design (40×20 sites) │   │
│  │  (implements here)                │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  Built-in Self-Test Circuits     │   │
│  └──────────────────────────────────┘   │
│  ┌──────────────────────────────────┐   │
│  │  Power Delivery & Distribution   │   │
│  └──────────────────────────────────┘   │
│  GPIO Pads (38), Clock, Power Rails     │
└─────────────────────────────────────────┘
```

---

### 4.3 IIC-OSIC-TOOLS Docker Container

**GitHub:** https://github.com/iic-jku/IIC-OSIC-TOOLS
**Type:** Pre-integrated container with all tools

#### What's Included

```
IIC-OSIC-TOOLS container
├── Yosys (synthesis)
├── OpenROAD (place & route)
├── OpenLane (automated flow)
├── Magic (layout editor)
├── Netgen (LVS)
├── KLayout (viewer)
├── ngspice (SPICE simulator)
├── Xyce (parallel SPICE)
├── PDKs (SKY130, GF180, IHP130)
└── JupyterLab (interactive notebooks)
```

#### Installation & Usage

**Pull and run:**
```bash
docker pull hpretl/iic-osic-tools:latest
docker run -d -p 8888:8888 -p 80:80 \
  -v $(pwd):/root/designs \
  hpretl/iic-osic-tools:latest

# Access JupyterLab at http://localhost:8888
```

**All tools available in container:**
```bash
# Inside container
yosys
openroad
magic -T sky130A
ngspice
klayout
```

#### Advantages

- **No installation hassle** - everything pre-configured
- **Multiple PDKs** - supports SKY130, GF180, IHP130 simultaneously
- **All tools integrated** - perfect for learning
- **Jupyter environment** - interactive tutorials included

---

## 5. ngspice - Circuit Simulation

**URL:** https://ngspice.sourceforge.io/
**License:** BSD-3-Clause
**Language:** C

#### Purpose
SPICE circuit simulator for analog blocks (DAC, ADC, op-amps). Supports ferroelectric device models.

#### Installation

```bash
sudo apt-get install ngspice
```

**From Source:**
```bash
git clone git://ngspice.git.sourceforge.net/gitroot/ngspice/ngspice
cd ngspice
./configure --prefix=/usr/local
make
sudo make install
```

#### FeCIM Applications

**1. DAC Verification:**
```spice
* DAC design
.title 8-bit DAC with FeCIM crossbar

.include sky130_fd_pr/models/sky130.lib.spice typical

* Reference ladder
R1 vref gnd 10k
R2 gnd vout 10k
* Crossbar array (behavioral model)
X1 vref bit7 bit6 ... bit0 vout fecim_dac_array

.control
* Run transient analysis
transient 0 1u 10n
plot v(vout)
.endc

.end
```

**2. Read-path simulation (TIA):**
```spice
* Transimpedance Amplifier
.include sky130_fd_pr/models/sky130.lib.spice

* Input current source (from FeCIM array)
Iin 1 gnd dc 0 ac 1u pulse(0 1u 0 1n 1n 5n 10n)

* Feedback network
Rf 1 2 100k
Cf 1 2 10p

* Ideal op-amp
E1 out_p out_n 1 gnd 100k
* Output buffer
Rout out_p vout 1k

.control
ac 1 10 100
plot vdb(vout)
.endc

.end
```

#### Key Parameters for Simulation

| Parameter | Purpose |
|-----------|---------|
| `.include sky130_*.lib` | Load transistor models |
| `.param` | Define design variables |
| `.meas` | Measure results (gain, BW, etc.) |
| `.control` | Simulation commands (ac, transient, etc.) |

---

## 6. Complete Workflow: From Concept to GDSII

### Step 1: Specification (1-2 hours)

Define your FeCIM peripheral:
- **Clock frequency:** (e.g., 100 MHz)
- **I/O requirements:** (GPIO, ADC precision, DAC resolution)
- **Area budget:** (mm² or number of cells)
- **Power budget:** (mW)

**Example: Row Address Decoder**
- Input: 5-bit row address
- Output: 32 one-hot signals (for 32-row array)
- Clock: 100 MHz
- Area: < 100 μm × 100 μm

### Step 2: RTL Design (2-4 hours)

Write behavioral Verilog.

```verilog
module row_decoder_32x5 (
  input clk, reset,
  input [4:0] addr,
  output reg [31:0] row_sel
);

  always @(posedge clk or posedge reset) begin
    if (reset)
      row_sel <= 32'b0;
    else
      row_sel <= (32'b1 << addr);
  end

endmodule
```

### Step 3: Simulation & Verification (1-2 hours)

Write testbench, verify functionality.

```verilog
`timescale 1ns/1ps

module tb_row_decoder;
  reg clk, reset;
  reg [4:0] addr;
  wire [31:0] row_sel;

  row_decoder_32x5 dut(.clk(clk), .reset(reset),
                        .addr(addr), .row_sel(row_sel));

  initial begin
    clk = 0;
    forever #5 clk = ~clk;
  end

  initial begin
    reset = 1; #20 reset = 0;
    for (int i = 0; i < 32; i++) begin
      addr = i; #10;
      assert(row_sel == (1 << i)) else
        $error("Mismatch at addr=%d", i);
    end
    $display("All tests passed!");
    $finish;
  end

endmodule
```

**Run with Yosys:**
```bash
yosys -m ghdl -p "read_verilog decoder.v; read_verilog tb.v; \
  opt -full; proc; show" -o sim.vcd
```

### Step 4: Synthesis (30 minutes)

Convert Verilog to gate-level netlist.

```tcl
# synth.tcl
read_verilog row_decoder.v
prep -top row_decoder_32x5
synth_sky130 -json synth.json
write_netlist netlist.json
```

```bash
yosys synth.tcl 2>&1 | tee synth.log
```

**Check results:**
```bash
# Cell count
grep "Number of cells" synth.log

# Cell types
grep "cell" netlist.json | head -20
```

### Step 5: Physical Design (1-2 hours)

Use OpenLane for place & route.

```tcl
# config.tcl
set ::env(DESIGN_NAME) "row_decoder_32x5"
set ::env(VERILOG_FILES) "./src/row_decoder.v"
set ::env(CLOCK_PORT) "clk"
set ::env(CLOCK_PERIOD) "10"  # 100 MHz
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 150 150"
set ::env(FP_CORE_UTIL) "0.60"
```

```bash
flow.tcl -design designs/row_decoder 2>&1 | tee pnr.log
```

### Step 6: Verification & Signoff (1-2 hours)

```bash
# DRC check
magic -T sky130A -Batch -r drc.tcl

# LVS verification
netgen -batch compare.tcl

# GDSII inspection
klayout results/final/row_decoder.gds

# Timing report
cat results/final/timing_summary.rpt
```

### Step 7: Prepare for Fabrication (1-2 hours)

**Final GDSII:**
```bash
cp results/final/*.gds final_tapeout.gds
```

**Submission package:**
```
├── final_tapeout.gds        # GDSII ready for fabrication
├── design_summary.txt       # Design specs
├── schematic.v              # RTL reference
├── simulation_results/      # Verification evidence
└── README.md               # Design notes
```

**Total effort: ~10 hours for a simple decoder**

---

## 7. Practical Example: Complete FeCIM DAC Design

This example shows a complete DAC design flow using open-source tools.

### Specification

- **Resolution:** 8-bit (256 levels)
- **Reference:** 1.8 V (SKY130 Vdd)
- **Clock:** 100 MHz
- **Output:** Analog voltage (0-1.8V)

### Design Files

**dac_8bit.v - Digital Controller:**
```verilog
module dac_8bit_controller (
  input clk, reset,
  input [7:0] digital_in,
  output reg [7:0] code_out,
  output reg code_valid
);

  always @(posedge clk or posedge reset) begin
    if (reset) begin
      code_out <= 8'b0;
      code_valid <= 1'b0;
    end else begin
      code_out <= digital_in;
      code_valid <= 1'b1;
    end
  end

endmodule
```

**dac_analog.spi - Analog Section (SPICE):**
```spice
* 8-bit Resistor Ladder DAC
* Outputs proportional to digital input

.title 8-bit DAC using R-ladder

.include sky130_fd_pr/models/sky130.lib.spice typical

* Reference voltage divider
Vref vref 0 dc 1.8

* Resistor ladder (simplified 3-bit example)
R0 vref n1 10k
R1 n1 n2 10k
R2 n2 n3 10k
R3 n3 gnd 10k

* Output buffer (unity gain)
E_out vout 0 n1 gnd 1

* Load resistor
Rload vout 0 100k

* Analysis
.control
dc Vref 0 1.8 0.1
plot vout
quit
.endc

.end
```

**tb_dac.v - Testbench:**
```verilog
`timescale 1ns/1ps

module tb_dac;
  reg clk, reset;
  reg [7:0] digital_in;
  wire [7:0] code_out;
  wire code_valid;

  dac_8bit_controller dut (
    .clk(clk), .reset(reset),
    .digital_in(digital_in),
    .code_out(code_out),
    .code_valid(code_valid)
  );

  initial begin
    clk = 0;
    forever #5 clk = ~clk;
  end

  initial begin
    reset = 1; #20 reset = 0;

    for (int i = 0; i < 256; i++) begin
      digital_in = i;
      #10;
      $display("Input: %3d, Output: %3d, Valid: %b",
        digital_in, code_out, code_valid);
    end

    $finish;
  end

endmodule
```

**OpenLane config.tcl:**
```tcl
set ::env(DESIGN_NAME) "dac_8bit_controller"
set ::env(VERILOG_FILES) "./src/dac_8bit.v"
set ::env(CLOCK_PORT) "clk"
set ::env(CLOCK_PERIOD) "10"
set ::env(FP_SIZING) "absolute"
set ::env(DIE_AREA) "0 0 200 200"
set ::env(FP_CORE_UTIL) "0.55"
set ::env(ROUTING_CORES) "4"
```

**Build:**
```bash
# Directory structure
designs/fecim_dac/
├── config.tcl
├── src/
│   ├── dac_8bit.v
│   └── tb_dac.v
└── sim/
    └── dac_analog.spi

# Synthesize
cd designs/fecim_dac
flow.tcl 2>&1 | tee build.log

# Results
# runs/RUN_*/results/final/dac_8bit_controller.gds
```

---

## 8. Troubleshooting Common Issues

### Issue 1: "Cannot find sky130 PDK"

**Symptom:**
```
Error: Cannot find PDK sky130A
```

**Solution:**
```bash
export PDK_ROOT=$(pwd)/skywater-pdk
export PDK=sky130A
export STD_CELL_LIBRARY=sky130_fd_sc_hd

flow.tcl -design ...
```

### Issue 2: Synthesis fails with "No cells mapped"

**Symptom:**
```
Warning: Cell type 'my_cell' not found in library
```

**Solution:**
- Verify cell exists in Liberty library
- Check library file syntax
- Use `-json` flag to inspect cell mapping

```bash
yosys -p "read_verilog design.v; synth_sky130; show"
```

### Issue 3: Place & Route fails due to high congestion

**Symptom:**
```
Routing congestion exceeded 100%
```

**Solution:**
```tcl
# Increase core area
set ::env(DIE_AREA) "0 0 300 300"  # Increase from 200x200
set ::env(FP_CORE_UTIL) "0.45"     # Lower utilization

# Increase routing layers
set ::env(DETAILED_ROUTER) "tritonRoute"
```

### Issue 4: Timing violations

**Symptom:**
```
Worst timing path: +2.3 ns slack (setup violation)
```

**Solution:**
```tcl
# Increase clock period
set ::env(CLOCK_PERIOD) "15"  # From 10 ns to 15 ns

# Enable optimization
set ::env(SYNTH_STRATEGY) "DELAY 0"
set ::env(PL_TARGET_DENSITY) "0.50"

# Increase buffer depth for timing
set ::env(CTS_CLK_BUFFER_LIST) "sky130_fd_sc_hd__clkbuf_16"
```

### Issue 5: DRC violations

**Symptom:**
```
Magic DRC: 150 violations found
```

**Check violations:**
```bash
magic -T sky130A -Batch << EOF
load final/design.gds
drc on
drc area
EOF
```

**Common issues:**
- Metal spacing too close (increase routing layer width)
- Via density (reduce routing pitch)
- Contact-to-gate distance (layout issue)

---

## 9. Resources & References

### Official Documentation

| Tool | URL |
|------|-----|
| Yosys | https://yosyshq.github.io/yosys/ |
| OpenROAD | https://openroad.readthedocs.io/ |
| OpenLane | https://openlane.readthedocs.io/ |
| Magic | http://opencircuitdesign.com/magic/ |
| SkyWater SKY130 | https://skywater-pdk.readthedocs.io/ |
| Caravel | https://caravel-user-project.readthedocs.io/ |

### Tutorials & Workshops

1. **Zero-to-ASIC (Caravel + SKY130):** https://github.com/mattvenn/caravel_layouts
2. **OpenROAD Tutorials:** https://github.com/The-OpenROAD-Project/OpenROAD-Tutorials
3. **Magic VLSI Guide:** http://opencircuitdesign.com/magic/userguide.html
4. **Efabless OSIC Guide:** https://github.com/efabless/open_source_chip_design

### Community

- **Efabless Forum:** https://groups.google.com/forum/#!forum/chipIgnite
- **SKY130 Community:** 6,400+ members on various platforms
- **OpenROAD Slack:** https://openroad.readthedocs.io/en/latest/
- **GitHub Issues:** Each tool has active issue tracking

### Related FeCIM Documentation

- [Module 6: EDA Design Suite](../README.md)
- [Module 6 EDA](../../2-learn/module6-eda/README.md) — EDA module overview and tools

---

## 10. Key Takeaways

### For Quick Prototyping
1. Use **Docker** (IIC-OSIC-TOOLS) - saves installation time
2. Start with **OpenLane** - fully automated flow
3. Target **SKY130** - most community support

### For Production Designs
1. Use **native tools** for fine control
2. Implement **comprehensive verification** (DRC/LVS/timing)
3. Plan **integration** with Caravel or custom harness

### For FeCIM Specifically
1. **Digital logic** → Design with SKY130 standard cells
2. **Analog periphery** → Use GF180MCU for higher voltage
3. **Custom FeFET cells** → Design in Magic, integrate as macros
4. **Mixed-signal verification** → Combine ngspice + SPICE extraction

### Timeline Estimate
- **Simple decoder:** 10 hours (concept → GDSII)
- **DAC/ADC:** 20-40 hours (including SPICE verification)
- **Full FeCIM array + control:** 100+ hours (new PDK/process support)

---

## Summary

Open-source EDA tools now enable complete chip design from RTL to fabrication:

- **Yosys** synthesizes RTL to gates
- **OpenROAD** places and routes on silicon
- **Magic** provides layout verification
- **SKY130 PDK** gives proven, free process design
- **OpenLane** automates the entire flow
- **Efabless/Caravel** enables free chip fabrication

For FeCIM chip design, these tools allow digital control logic and standard analog peripherals. Custom ferroelectric cells require process partnerships but can be integrated as hardened macros.

---

**Document maintained by:** FeCIM Documentation Team
**Last updated:** 2026-01-27
**Status:** Comprehensive reference guide
