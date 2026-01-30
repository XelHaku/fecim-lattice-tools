# OpenLane & Open-Source EDA Tools: Comprehensive CLI Reference

**Purpose:** Complete reference for OpenLane flow and its component tools
**Last Updated:** 2026-01-26
**Sources:** Official documentation, GitHub repositories, community resources

---

## Table of Contents

1. [OpenLane Overview](#1-openlane-overview)
2. [Yosys - Synthesis](#2-yosys---synthesis)
3. [OpenROAD - Physical Design](#3-openroad---physical-design)
4. [Magic VLSI - Layout](#4-magic-vlsi---layout)
5. [KLayout - GDSII Viewer/Editor](#5-klayout---gdsii-viewereditor)
6. [Netgen - LVS](#6-netgen---lvs)
7. [OpenSTA - Timing Analysis](#7-opensta---timing-analysis)
8. [PDK Installation](#8-pdk-installation)
9. [Validation & Analysis CLI](#9-validation--analysis-cli)
10. [Quick Reference Tables](#10-quick-reference-tables)
11. [**Module 6 CLI Cheatsheet**](#11-module-6-cli-cheatsheet)

---

## 1. OpenLane Overview

OpenLane is an automated RTL-to-GDSII flow based on OpenROAD, Yosys, Magic, Netgen, CVC, KLayout, and custom scripts.

**Official Resources:**
- Documentation: https://openlane2.readthedocs.io/
- GitHub (OpenLane 2): https://github.com/efabless/openlane2
- GitHub (OpenLane 1): https://github.com/The-OpenROAD-Project/OpenLane
- PyPI: https://pypi.org/project/openlane/

### 1.1 OpenLane 1 vs OpenLane 2

| Aspect | OpenLane 1 | OpenLane 2 |
|--------|------------|------------|
| Language | Tcl scripts | Python with Tcl compatibility |
| Configuration | config.tcl | JSON, YAML, or Tcl |
| Architecture | Procedural scripts | Modular step-based |
| API | Minimal | Well-documented Python API |
| Type Checking | None | Built-in validation |
| Entry Point | `./flow.tcl` | `openlane` CLI |

### 1.2 Installation Methods

#### Nix-based (Recommended)
```bash
# Install Nix and Cachix first
nix-shell
openlane --smoke-test
```

#### Docker-based
```bash
cd $HOME/OpenLane
make mount
# Inside container:
./flow.tcl -design spm
```

#### PyPI (OpenLane 2)
```bash
pip install openlane
openlane --dockerized --smoke-test
```

### 1.3 OpenLane CLI Commands

#### Basic Flow Execution
```bash
# Run complete flow (OpenLane 2)
openlane --pdk-root /path/to/pdk config.json

# Run with Docker
openlane --dockerized --pdk-root /path/to/pdk config.json

# Smoke test
openlane --log-level ERROR --condensed --show-progress-bar --smoke-test

# Open results in KLayout
openlane --last-run --flow OpenInKLayout config.json

# Open results in Magic
openlane --last-run --flow OpenInMagic config.json
```

#### OpenLane 1 flow.tcl Commands
```bash
# Run autonomous flow
./flow.tcl -design <design_name>

# Interactive mode
./flow.tcl -interactive

# Initialize new design
./flow.tcl -design <design_name> -init_design_config

# Synthesis exploration
./flow.tcl -design <design_name> -synth_explore

# Specify PDK
./flow.tcl -design <design_name> -pdk sky130A
```

### 1.4 Interactive Mode Commands (OpenLane 1)

```tcl
# Must run first
package require openlane 0.9
prep -design <design_name>

# Individual steps
run_synthesis
run_floorplan
run_placement
run_cts
run_routing
run_magic
run_magic_spice_export
run_lvs
run_antenna_check
```

### 1.5 OpenLane 2 Flow Steps (Classic Flow)

The Classic flow executes these steps in sequence:

```
Yosys.JsonHeader → Yosys.Synthesis → Checker.YosysUnmappedCells →
Checker.YosysSynthChecks → OpenROAD.CheckSDCFiles → OpenROAD.Floorplan →
OpenROAD.TapEndcapInsertion → OpenROAD.GeneratePDN → OpenROAD.IOPlacement →
OpenROAD.GlobalPlacement → OpenROAD.DetailedPlacement →
OpenROAD.GlobalRouting → OpenROAD.DetailedRouting → OpenROAD.FillInsertion →
Magic.StreamOut → Magic.DRC → Magic.SpiceExtraction → Netgen.LVS
```

### 1.6 Key Configuration Variables

#### Synthesis
| Variable | Description | Default |
|----------|-------------|---------|
| `SYNTH_STRATEGY` | Optimization strategy (AREA/DELAY) | AREA 0 |
| `SYNTH_NO_FLAT` | Disable hierarchy flattening | 0 |
| `SYNTH_SIZING` | Enable gate sizing | 0 |
| `SYNTH_BUFFERING` | Enable buffering | 1 |
| `MAX_FANOUT_CONSTRAINT` | Maximum fanout | 10 |

#### Timing
| Variable | Description | Default |
|----------|-------------|---------|
| `CLOCK_PERIOD` | Clock period (ns) | 10 |
| `CLOCK_PORT` | Clock pin name | clk |
| `CLOCK_NET` | Clock net name | - |

#### Floorplan
| Variable | Description | Default |
|----------|-------------|---------|
| `FP_CORE_UTIL` | Core utilization (%) | 50 |
| `FP_ASPECT_RATIO` | Aspect ratio | 1 |
| `FP_SIZING` | absolute/relative | relative |
| `FP_PDN_CORE_RING` | Add power ring | 0 |

#### Placement
| Variable | Description | Default |
|----------|-------------|---------|
| `PL_TARGET_DENSITY` | Target density | 0.55 |
| `PL_ROUTABILITY_DRIVEN` | Enable routability | 1 |
| `PL_TIME_DRIVEN` | Enable timing-driven | 1 |

---

## 2. Yosys - Synthesis

Yosys is the open-source RTL synthesis suite that transforms Verilog to gate-level netlists.

**Official Resources:**
- GitHub: https://github.com/YosysHQ/yosys
- Documentation: https://yosyshq.readthedocs.io/projects/yosys/en/latest/
- Man page: https://www.mankier.com/1/yosys

### 2.1 CLI Options

```bash
# Basic usage
yosys [OPTIONS] [INFILES]

# Key options
-b, --backend <backend>      # Output backend (e.g., verilog, json)
-f, --frontend <frontend>    # Input frontend (e.g., verilog, liberty)
-s, --scriptfile <file>      # Execute script file
-p, --commands <cmds>        # Execute commands (semicolon-separated)
-c, --tcl-scriptfile <file>  # Execute TCL script
-r, --top <module>           # Specify top module
-m, --plugin <plugin>        # Load plugin module
-D, --define <name>[=val]    # Set Verilog define
-S, --synth                  # Run default synth command
-o, --outfile <file>         # Write design to file on exit
-q, --quiet                  # Quiet operation
-v, --verbose <level>        # Verbosity level (0-9)
-l, --logfile <file>         # Write log to file
-Q                           # Suppress banner
-T                           # Suppress footer
-V, --version                # Print version
```

### 2.2 Essential Synthesis Commands

#### Reading/Writing Files
```tcl
# Read Verilog
read_verilog design.v
read_verilog -defer design.v         # Defer elaboration
read_verilog -sv design.sv           # SystemVerilog

# Read Liberty library
read_liberty -lib cells.lib

# Write outputs
write_verilog synth.v                # Verilog netlist
write_json design.json               # JSON netlist
write_blif design.blif               # BLIF format
write_edif design.edif               # EDIF format
```

#### Hierarchy & Elaboration
```tcl
hierarchy -check -top <module>       # Elaborate with top module
hierarchy -auto-top                  # Auto-detect top
flatten                              # Flatten hierarchy
```

#### High-Level Transforms
```tcl
proc                                 # Process procedures
opt                                  # General optimization
fsm                                  # FSM extraction/optimization
memory                               # Memory inference
```

#### Technology Mapping
```tcl
techmap                              # Map to internal library
dfflibmap -liberty cells.lib         # Map flip-flops
abc -liberty cells.lib               # Map combinational logic
abc -lut 4                           # Map to 4-LUTs (FPGA)
abc9 -lut 4                          # Enhanced LUT mapping
```

#### Cleanup
```tcl
clean                                # Remove unused
opt_clean                            # Clean + optimize
```

### 2.3 Complete Synthesis Script Example

```tcl
# ASIC synthesis with Liberty library
read_verilog design.v
hierarchy -check -top top_module

# High-level synthesis
proc; opt; fsm; opt; memory; opt

# Technology mapping
techmap; opt
dfflibmap -liberty mycells.lib
abc -liberty mycells.lib

# Cleanup and output
clean
write_verilog synth.v
stat
```

### 2.4 Platform-Specific Synth Commands

```tcl
# Generic synthesis
synth -top <module>

# Specific targets
synth_xilinx -top <module>           # Xilinx FPGAs
synth_ice40 -top <module>            # Lattice iCE40
synth_ecp5 -top <module>             # Lattice ECP5
synth_gowin -top <module>            # Gowin FPGAs
synth_intel -top <module>            # Intel/Altera FPGAs

# Generic (for verification)
prep -top <module>                   # Coarse-grain only
```

### 2.5 ABC Options in Yosys

```tcl
# Basic ABC with Liberty
abc -liberty cells.lib

# With timing constraints
abc -liberty cells.lib -D 1000       # 1000 ps delay target
abc -liberty cells.lib -constr constraints.sdc

# LUT mapping (FPGA)
abc -lut 6                           # 6-input LUTs
abc9 -lut 4:6                        # Variable LUT sizes

# Options
abc -liberty cells.lib -dff          # Include flip-flops
abc -liberty cells.lib -keepff       # Keep FF outputs
abc -liberty cells.lib -nocleanup    # Debug: keep temp files
```

---

## 3. OpenROAD - Physical Design

OpenROAD is the unified application for physical design, handling floorplanning through routing.

**Official Resources:**
- GitHub: https://github.com/The-OpenROAD-Project/OpenROAD
- Documentation: https://openroad.readthedocs.io/en/latest/
- Flow Scripts: https://github.com/The-OpenROAD-Project/OpenROAD-flow-scripts

### 3.1 CLI Options

```bash
# Basic usage
openroad [OPTIONS] [script.tcl]

# Key options
-help                    # Show help
-version                 # Show version
-no_init                 # Skip .openroad init file
-no_splash               # Suppress startup message
-exit                    # Exit after script
-gui                     # Launch GUI
-log <file>              # Log file
-metrics <file>          # Metrics output file
```

### 3.2 Floorplan Commands

```tcl
# Initialize floorplan
initialize_floorplan \
  -die_area {0 0 1000 1000} \
  -core_area {100 100 900 900} \
  -site unithd

# Or by utilization
initialize_floorplan \
  -utilization 50 \
  -aspect_ratio 1.0 \
  -core_space 10 \
  -site unithd

# Add routing tracks
make_tracks metal1 -x_offset 0.17 -x_pitch 0.34 -y_offset 0.17 -y_pitch 0.34
make_tracks metal2 -x_offset 0.23 -x_pitch 0.46 -y_offset 0.23 -y_pitch 0.46

# Tapcell and endcap insertion
tapcell \
  -distance 14 \
  -tapcell_master TAPCELL_X1 \
  -endcap_master ENDCAP_X1

# Power distribution network
pdngen pdn.cfg
```

### 3.3 Placement Commands

```tcl
# Global placement
global_placement \
  -density 0.6 \
  -pad_left 2 \
  -pad_right 2

# IO placement
place_pins \
  -hor_layers metal3 \
  -ver_layers metal2

# Detailed placement
detailed_placement \
  -max_displacement {100 100}

# Filler cell insertion
filler_placement [list FILL1 FILL2 FILL4]

# Remove fillers (if needed)
remove_fillers

# Optimize placement for timing
repair_design
repair_timing
```

### 3.4 Clock Tree Synthesis (CTS)

```tcl
# Configure CTS
set_wire_rc -clock -layer metal3

# Run CTS
clock_tree_synthesis \
  -root_buf CLKBUF_X3 \
  -buf_list {CLKBUF_X1 CLKBUF_X2 CLKBUF_X3} \
  -wire_unit 20

# Repair clock tree
repair_clock_nets
repair_clock_inverters
```

### 3.5 Routing Commands

```tcl
# Set routing layers
set_routing_layers -signal metal1-metal5 -clock metal3-metal5

# Global routing
global_route \
  -guide_file route.guide \
  -congestion_iterations 50 \
  -verbose

# Detailed routing (TritonRoute)
detailed_route \
  -guide route.guide \
  -output_drc drc.rpt \
  -output_maze maze.log \
  -verbose 1

# Antenna repair
repair_antennas -iterations 5

# Fill insertion
density_fill \
  -rules fill_rules.json
```

### 3.6 Analysis & Reports

```tcl
# Timing reports
report_checks -path_delay max
report_checks -path_delay min
report_tns
report_wns

# Power reports
report_power

# Design statistics
report_design_area
report_cell_usage

# DRC
check_placement -verbose
check_routing
```

### 3.7 File I/O

```tcl
# Read files
read_lef tech.lef
read_lef cells.lef
read_def design.def
read_liberty cells.lib
read_sdc constraints.sdc
read_verilog design.v
link_design top_module

# Write files
write_def output.def
write_verilog output.v
write_db design.odb
```

---

## 4. Magic VLSI - Layout

Magic is the open-source VLSI layout editor with DRC, extraction, and LVS capabilities.

**Official Resources:**
- Website: http://opencircuitdesign.com/magic/
- GitHub: https://github.com/RTimothyEdwards/magic
- Command Reference: http://opencircuitdesign.com/magic/userguide.html
- Tutorials: http://opencircuitdesign.com/magic/tutorials/

### 4.1 CLI Options

```bash
# Basic usage
magic [OPTIONS] [cellname]

# Key options
-noconsole               # No console window
-dnull                   # No graphics (batch mode)
-T <techfile>            # Specify technology file
-rcfile <file>           # Use specific startup file
-norcfile                # Skip .magicrc
-d <display>             # Graphics driver (X11, OGL, NULL)
```

### 4.2 File Operations

```tcl
# Load/save cells
load cellname
save cellname
writeall                 # Save all modified cells

# Read GDSII
gds read design.gds
gds readonly true
gds flatten true

# Write GDSII
gds write design.gds

# Read/write CIF
cif read design.cif
cif write design.cif

# Read DEF/LEF
lef read cells.lef
def read design.def
def write design.def
```

### 4.3 Layout Commands

```tcl
# Box operations
box 0 0 100 100           # Set box coordinates
box width 50              # Set box width
box height 50             # Set box height
box move right 10         # Move box

# Paint/erase
paint metal1              # Paint layer in box
paint m1                  # Shorthand
erase metal1              # Erase layer in box
erase *                   # Erase all layers

# Wire tool
wire type metal1
wire width 0.5
wire horizontal
wire vertical

# Polygon
polygon metal1 0 0 10 0 10 10 5 15 0 10

# Labels
label "signal_name" center metal1
port make input
port class input
```

### 4.4 Selection Commands

```tcl
# Select operations
select area              # Select in box
select cell cellname     # Select cell instances
select clear             # Clear selection
select top cell          # Select top cell
select visible           # Select visible layers

# Move/copy
move right 10
move up 5
copy
```

### 4.5 DRC (Design Rule Checking)

```tcl
# Run DRC
drc check                # Check current cell
drc catchup              # Complete DRC on cell
drc find                 # Find next error
drc find [nth]           # Find nth error
drc why                  # Explain error in box
drc count                # Count errors

# DRC settings
drc on                   # Enable continuous DRC
drc off                  # Disable continuous DRC
drc style drc(full)      # Full DRC
drc euclidean on         # Euclidean distance checks
```

### 4.6 Extraction & SPICE

```tcl
# Extract parasitic netlist
extract all              # Extract all cells
extract unique           # Extract with unique names
extract no all           # Clear extraction

# Generate SPICE
ext2spice lvs            # Setup for LVS
ext2spice                # Generate SPICE file

# Options
ext2spice cthresh 0.01   # Capacitance threshold
ext2spice rthresh 1      # Resistance threshold
ext2spice hierarchy on   # Hierarchical extraction
ext2spice subcircuit top on  # Top cell as subckt

# Resistor extraction
extresist all
extresist tolerance 10

# Combined flow
extract all
ext2spice lvs
ext2spice -o design.spice
```

### 4.7 Batch Mode Script Example

```bash
#!/bin/bash
magic -dnull -noconsole << EOF
tech load sky130A
gds read design.gds
load topcell
select top cell
extract all
ext2spice lvs
ext2spice -o design.spice
quit
EOF
```

### 4.8 Common Command Reference

| Command | Description |
|---------|-------------|
| `help` | List all commands |
| `help <cmd>` | Help for specific command |
| `tech load <name>` | Load technology |
| `cellname list children` | List sub-cells |
| `cellname list parents` | List parent cells |
| `property` | View/set cell properties |
| `cif ostyle` | Set CIF output style |
| `gds ordering on` | Preserve GDS ordering |

### 4.9 Image Generation (PNG/SVG)

Magic can export layout images via `plot pnm` (then convert to PNG) or `plot svg`.

#### PNM Export (Batch Mode)

```bash
magic -dnull -noconsole << 'EOF'
tech load sky130A
gds read design.gds
load topcell
select top cell
box 0 0 1000 1000
plot pnm output.pnm 1500
quit
EOF

# Convert PNM to PNG with ImageMagick
convert output.pnm output.png
```

#### Plot PNM Command Syntax

```tcl
# Basic syntax (width in pixels, default 1500)
plot pnm <filename> [width] [layers]

# Examples
plot pnm layout.pnm                    # Default 1500px width
plot pnm layout.pnm 3000               # 3000px width
plot pnm layout.pnm 2000 "metal1,metal2"  # Specific layers
```

#### SVG Export (Requires Cairo)

```bash
# Must use Cairo graphics (-d XR)
magic -d XR << 'EOF'
tech load sky130A
load topcell
plot svg output.svg
quit
EOF
```

#### Batch Script Example

```bash
#!/bin/bash
# generate_layout_image.sh
INPUT_GDS=$1
OUTPUT_PNG=$2
CELL_NAME=${3:-topcell}

magic -dnull -noconsole << EOF
tech load sky130A
gds read $INPUT_GDS
load $CELL_NAME
select top cell
plot pnm /tmp/layout_temp.pnm 2000
quit
EOF

convert /tmp/layout_temp.pnm $OUTPUT_PNG
rm /tmp/layout_temp.pnm
```

| Format | Command | Notes |
|--------|---------|-------|
| PNM | `plot pnm file [width]` | Best for large layouts, needs ImageMagick to convert |
| SVG | `plot svg file` | Requires `-d XR` (Cairo), vector format |
| PostScript | `plot postscript file` | Legacy format |

---

## 5. KLayout - GDSII Viewer/Editor

KLayout is a high-performance layout viewer and editor supporting GDSII, OASIS, and other formats.

**Official Resources:**
- Website: https://www.klayout.de/
- Command Args: https://www.klayout.de/command_args.html
- Documentation: https://www.klayout.de/doc.html

### 5.1 CLI Options

```bash
# Basic usage
klayout [OPTIONS] [files...]

# Key options
-b                       # Batch mode (no GUI)
-zz                      # Non-GUI mode (no display)
-e                       # Edit mode
-ne                      # Non-edit mode (view only)
-r <script>              # Run script and exit
-rm <script>             # Run script then continue
-rd <var>=<value>        # Define variable for script
-l <file>                # Layer properties file
-u <file>                # Session file
-s                       # Sync with other instance
-p <plugin>              # Load plugin
-j <threads>             # Number of threads
-t                       # Enable undo/redo
-nn <tech>               # Technology name
-n <tech>                # Technology file
```

### 5.2 Batch Mode Operations

```bash
# Run DRC in batch mode
klayout -b -r drc_rules.drc

# With variables
klayout -b \
  -rd input=design.gds \
  -rd report=drc_report.lyrdb \
  -r my_drc.drc

# Convert formats
klayout -b -r convert.rb \
  -rd input=design.gds \
  -rd output=design.oas

# Run LVS
klayout -b \
  -rd schematic=design.spice \
  -rd layout=design.gds \
  -rd report=lvs_report.lvsdb \
  -r lvs_rules.lvs
```

### 5.3 DRC Script Example (.drc)

```ruby
# DRC script (Ruby-based DSL)
source($input)  # Variable from -rd input=...
report($report) # Variable from -rd report=...

# Define layers
metal1 = input(68, 20)
metal2 = input(69, 20)
via1 = input(68, 44)

# DRC rules
metal1.width(0.14).output("M1 width < 0.14um")
metal1.space(0.14).output("M1 space < 0.14um")
metal2.width(0.14).output("M2 width < 0.14um")
metal2.space(0.14).output("M2 space < 0.14um")

# Enclosure
metal1.enclosing(via1, 0.03).output("M1 via enclosure")
metal2.enclosing(via1, 0.03).output("M2 via enclosure")
```

### 5.4 LVS Script Example (.lvs)

```ruby
# LVS script (Ruby-based DSL)
deep

# Source layout and schematic
source($layout)
schematic($schematic)

# Define device recognition layers
nwell = input(64, 20)
diff = input(65, 20)
poly = input(66, 20)
nsdm = input(93, 44)
psdm = input(94, 20)

# Define devices
nmos = nsdm & diff
pmos = psdm & diff

# Extract and compare
extract
compare
report($report)
```

### 5.5 Python/Ruby Scripting

```python
# Python script for KLayout
import pya

# Load layout
layout = pya.Layout()
layout.read("design.gds")

# Access top cell
top_cell = layout.top_cell()

# Iterate shapes
layer = layout.layer(68, 20)
for shape in top_cell.shapes(layer).each():
    print(shape.bbox)

# Save
layout.write("output.gds")
```

```ruby
# Ruby script for KLayout
layout = RBA::Layout.new
layout.read("design.gds")

top_cell = layout.top_cell
layer = layout.layer(68, 20)

top_cell.shapes(layer).each do |shape|
  puts shape.bbox.to_s
end

layout.write("output.gds")
```

### 5.6 Environment Variables

| Variable | Description |
|----------|-------------|
| `KLAYOUT_PATH` | Search paths (: separated on Linux) |
| `KLAYOUT_HOME` | Home directory for config |

### 5.7 Image Generation (PNG)

KLayout provides robust PNG export via Ruby/Python scripting. Use `-z` flag (not `-zz`) as it requires a main window object.

#### Basic CLI Command

```bash
klayout -z -r screenshot.rb -rd input=design.gds -rd output=design.png
```

#### Ruby Script (`screenshot.rb`)

```ruby
# Basic screenshot script
mw = RBA::Application::instance.main_window
mw.load_layout($input, 0)
view = mw.current_view

# Configure view
view.max_hier                                    # Show full hierarchy
view.set_config("background-color", "#ffffff")   # White background
view.set_config("grid-visible", "false")         # Hide grid

# Save image (width x height in pixels)
view.save_image($output, 2000, 2000)
```

#### Ruby Script with Zoom Control

```ruby
# Screenshot with specific region
input_file = $input
output_file = $output
width = 1000
height = 1000

# Rectangle to capture (left, bottom, right, top in microns)
rect = RBA::DBox::new(-100.0, -200.0, 100.0, 200.0)

mw = RBA::Application::instance.main_window
mw.load_layout(input_file, 0)
view = mw.current_view

view.max_hier
view.zoom_box(rect)
view.save_image(output_file, width, height)
```

#### Python Script

```python
import pya

# Load layout
app = pya.Application.instance()
mw = app.main_window()
mw.load_layout("design.gds", 0)

# Configure view
lv = mw.current_view()
lv.max_hier()

# Zoom to specific region (left, bottom, right, top in µm)
lv.zoom_box(pya.DBox(-100.0, -200.0, 100.0, 200.0))

# Save image
lv.save_image("output.png", 1920, 1080)
```

#### Layer Properties Configuration

```ruby
# Load layer properties file for consistent styling
mw = RBA::Application::instance.main_window
mw.load_layout($input, 0)
view = mw.current_view

# Load .lyp file for layer colors/visibility
view.load_layer_props("layers.lyp", false)

view.max_hier
view.save_image($output, 2000, 2000)
```

#### Batch Script Example

```bash
#!/bin/bash
# generate_gds_image.sh
INPUT_GDS=$1
OUTPUT_PNG=$2

cat > /tmp/screenshot.rb << 'RUBY'
mw = RBA::Application::instance.main_window
mw.load_layout($input, 0)
view = mw.current_view
view.max_hier
view.set_config("background-color", "#ffffff")
view.set_config("grid-visible", "false")
view.save_image($output, 2000, 2000)
RUBY

klayout -z -r /tmp/screenshot.rb -rd input="$INPUT_GDS" -rd output="$OUTPUT_PNG"
```

| Option | Description |
|--------|-------------|
| `view.max_hier` | Show all hierarchy levels |
| `view.zoom_fit` | Fit entire layout in view |
| `view.zoom_box(DBox)` | Zoom to specific region |
| `view.save_image(file, w, h)` | Save PNG (max ~500M pixels) |
| `view.set_config(key, val)` | Set view configuration |
| `view.load_layer_props(lyp)` | Load layer properties file |

---

## 6. Netgen - LVS

Netgen is the LVS (Layout vs. Schematic) verification tool.

**Official Resources:**
- Website: http://opencircuitdesign.com/netgen/
- GitHub: https://github.com/RTimothyEdwards/netgen
- Reference: http://opencircuitdesign.com/netgen/reference.html
- Tutorial: http://opencircuitdesign.com/netgen/tutorial/tutorial.html

### 6.1 CLI Options

```bash
# Basic usage
netgen [OPTIONS] [script.tcl]

# Key options
-noconsole               # No console
-batch                   # Batch mode
-log <file>              # Log file
```

### 6.2 LVS Commands

```tcl
# Simple LVS comparison
lvs "layout.spice subckt" "schematic.spice subckt" setup.tcl output.txt

# Full form
lvs layout.spice schematic.spice sky130A_setup.tcl lvs_results.log

# If setup.tcl doesn't exist, uses defaults
lvs layout.spice schematic.spice

# Read netlists separately
readnet spice layout.spice
readnet spice schematic.spice
# Then compare
compare
```

### 6.3 Common Commands

```tcl
# Read netlists
readnet spice design.spice
readnet verilog design.v

# Setup comparison
equate classes                # Equate device classes
equate pins                   # Equate pin names
property                      # Check properties
ignore                        # Ignore specific elements

# Run comparison
compare
run converge                  # Iterate until stable

# Reports
summary                       # Print summary
nodes                         # Print node info
elements                      # Print element info
print                         # Print full report
```

### 6.4 LVS Script Example

```tcl
#!/usr/bin/env netgen -batch source

# Load PDK setup
source /path/to/sky130A_setup.tcl

# Read netlists
readnet spice extracted.spice
readnet spice schematic.spice

# Run LVS
lvs "extracted.spice topcell" "schematic.spice topcell" \
    /path/to/setup.tcl \
    lvs_output.log

# Check results
if {[info exists lvs_result]} {
    if {$lvs_result == 0} {
        puts "LVS CLEAN"
    } else {
        puts "LVS ERRORS: $lvs_result"
    }
}
```

### 6.5 PDK Setup Files

Each PDK provides a setup file (e.g., `sky130A_setup.tcl`):

```tcl
# Example setup file content
permute default
property default
equate class {nfet_01v8 nfet_01v8_lvt}
equate class {pfet_01v8 pfet_01v8_hvt}
```

---

## 7. OpenSTA - Timing Analysis

OpenSTA is the static timing analysis engine used in OpenROAD.

**Official Resources:**
- GitHub: https://github.com/The-OpenROAD-Project/OpenSTA
- Manual: https://github.com/parallaxsw/OpenSTA/blob/master/doc/OpenSTA.pdf

### 7.1 CLI Options

```bash
# Basic usage
sta [OPTIONS] [script.tcl]

# Runs in TCL interpreter mode
```

### 7.2 Setup Commands

```tcl
# Read libraries
read_liberty -corner fast fast.lib
read_liberty -corner slow slow.lib
read_liberty cells.lib

# Read design
read_verilog design.v
link_design top_module

# Read constraints
read_sdc design.sdc

# Read parasitics
read_spef design.spef
# Or SDF
read_sdf design.sdf
```

### 7.3 SDC Constraint Commands

```tcl
# Create clock
create_clock -name clk -period 10.0 [get_ports clk]
create_clock -name clk -period 10 -waveform {0 5} [get_ports clk]

# Generated clocks
create_generated_clock -name clk_div2 \
    -source [get_ports clk] \
    -divide_by 2 \
    [get_pins divider/Q]

# Input/output delays
set_input_delay -clock clk 2.0 [get_ports {data_in[*]}]
set_output_delay -clock clk 2.0 [get_ports {data_out[*]}]

# Input transition
set_input_transition 0.5 [get_ports data_in]

# Output load
set_load 0.1 [get_ports data_out]

# Clock uncertainty
set_clock_uncertainty -setup 0.5 [get_clocks clk]
set_clock_uncertainty -hold 0.3 [get_clocks clk]

# False paths
set_false_path -from [get_clocks clk1] -to [get_clocks clk2]

# Multi-cycle paths
set_multicycle_path 2 -setup -from [get_pins reg1/Q] -to [get_pins reg2/D]

# Max delay
set_max_delay 5.0 -from [get_ports in] -to [get_ports out]
```

### 7.4 Reporting Commands

```tcl
# Timing reports
report_checks                     # All timing checks
report_checks -path_delay max     # Setup (max) paths
report_checks -path_delay min     # Hold (min) paths
report_checks -to [get_pins reg/D]  # To specific pin
report_checks -through [get_nets net1]  # Through net
report_checks -group_path_count 10  # Top 10 paths
report_checks -digits 4           # 4 decimal places

# Slack reports
report_tns                        # Total negative slack
report_wns                        # Worst negative slack

# Power analysis
report_power                      # Power report
report_power -instances           # Per-instance power

# Clock reports
report_clocks                     # Clock summary
report_clock_skew                 # Clock skew

# Design info
report_design                     # Design summary
report_units                      # Unit definitions
```

### 7.5 Multi-Corner Analysis

```tcl
# Define corners
define_corners slow fast

# Read libraries per corner
read_liberty -corner slow slow.lib
read_liberty -corner fast fast.lib

# Apply derating
set_timing_derate -early 0.95 -corner slow
set_timing_derate -late 1.05 -corner slow

# Report per corner
report_checks -corner slow
report_checks -corner fast
```

### 7.6 Complete STA Script Example

```tcl
# Read libraries
read_liberty cells.lib

# Read design
read_verilog synth.v
link_design top

# Read constraints
read_sdc constraints.sdc

# Read parasitics
read_spef design.spef

# Report timing
report_checks -path_delay max -format full_clock_expanded > setup.rpt
report_checks -path_delay min -format full_clock_expanded > hold.rpt
report_tns > tns.rpt
report_wns > wns.rpt

# Exit
exit
```

---

## 8. PDK Installation

### 8.1 SkyWater SKY130 PDK

**Official Resources:**
- GitHub: https://github.com/google/skywater-pdk
- Open_PDKs: https://github.com/RTimothyEdwards/open_pdks
- Documentation: https://skywater-pdk.readthedocs.io/

#### Installation via open_pdks

```bash
# Clone open_pdks
git clone https://github.com/RTimothyEdwards/open_pdks.git
cd open_pdks

# Configure for SKY130
./configure \
    --prefix=/usr \
    --enable-sky130-pdk \
    --enable-sram-sky130

# Build and install
make
sudo make install

# Set environment
export PDK_ROOT=/usr/share/pdk
export PDK=sky130A
```

#### Minimal Installation (Analog Only)

```bash
./configure \
    --enable-sky130-pdk \
    --enable-sram-sky130 \
    --disable-sc-hs-sky130 \
    --disable-sc-ms-sky130 \
    --disable-sc-ls-sky130 \
    --disable-sc-lp-sky130 \
    --disable-sc-hd-sky130 \
    --disable-sc-hdll-sky130 \
    --disable-sc-hvl-sky130
make
sudo make install
```

#### PDK Variants

| Variant | Description |
|---------|-------------|
| sky130A | Standard digital (most common) |
| sky130B | With ReRAM option |

### 8.2 GlobalFoundries GF180MCU PDK

**Official Resources:**
- GitHub: https://github.com/google/gf180mcu-pdk
- Open_PDKs support included

#### Installation

```bash
# Clone open_pdks
git clone https://github.com/RTimothyEdwards/open_pdks.git
cd open_pdks

# Configure for GF180MCU
./configure \
    --prefix=/usr \
    --enable-gf180mcu-pdk

# Build and install
make
sudo make install

# Set environment
export PDK_ROOT=/usr/share/pdk
export PDK=gf180mcuD
```

#### GF180MCU Variants

| Variant | Metal Stack | Description |
|---------|-------------|-------------|
| gf180mcuA | 3 metal | Basic |
| gf180mcuB | 4 metal | Standard |
| gf180mcuC | 5 metal | 0.9um thick top metal |
| gf180mcuD | 5 metal | 1.1um thick top metal (shuttles) |

### 8.3 Conda Installation (Alternative)

```bash
# Sky130
conda install -c litex-hub open_pdks.sky130A

# GF180MCU
conda install -c litex-hub open_pdks.gf180mcuC
```

---

## 9. Validation & Analysis CLI

This section covers CLI commands specifically useful for validating generated EDA files (LEF, Liberty, Verilog, DEF) - essential for tools like Module 6 that generate OpenLane-compatible files.

### 9.1 Verilog Validation

#### Yosys Syntax Check

```bash
# Basic syntax check (read and elaborate)
yosys -p "read_verilog design.v; hierarchy -check -top top_module"

# With detailed error reporting
yosys -p "read_verilog design.v; hierarchy -check -top top_module; check -assert"

# Check for unmapped cells after synthesis
yosys -p "read_verilog design.v; synth -top top; check -mapped -assert"
```

#### Yosys `check` Command Options

```tcl
check                    # Basic check for problems
check -assert            # Error if problems found
check -noinit            # Check 'init' attributes
check -initdrv           # Check init-driven wires
check -mapped            # Check for unmapped cells
check -allow-tbuf        # Allow $_TBUF_ cells
```

#### Verilator Lint (Recommended Pre-Check)

```bash
# Lint-only mode (no simulation)
verilator --lint-only design.v

# With specific warnings
verilator --lint-only -Wall design.v

# Specify top module
verilator --lint-only --top-module top_module design.v
```

#### Icarus Verilog Syntax Check

```bash
# Parse only (no elaboration)
iverilog -t null design.v

# With top module
iverilog -t null -s top_module design.v
```

### 9.2 Liberty (.lib) Validation

#### OpenSTA Liberty Check

```bash
# Create validation script
cat > check_liberty.tcl << 'EOF'
read_liberty -lib cells.lib
report_units
exit
EOF

# Run validation
sta check_liberty.tcl
```

#### OpenSTA Liberty Analysis

```tcl
# Read and report library info
read_liberty -lib cells.lib

# Report units (timing, capacitance, etc.)
report_units

# Report cell info
report_cell <cell_name>

# List all cells
report_lib <lib_name>
```

#### Liberty Syntax Check with Yosys

```bash
# Yosys can also validate Liberty files
yosys -p "read_liberty -lib cells.lib"
```

### 9.3 LEF Validation

#### Magic LEF Check

```bash
magic -dnull -noconsole << 'EOF'
tech load sky130A
lef read cells.lef
quit
EOF
```

If the LEF has errors, Magic will report them.

#### OpenROAD LEF Check

```bash
openroad << 'EOF'
read_lef tech.lef
read_lef cells.lef
exit
EOF
```

### 9.4 DEF Validation

#### Magic DEF Check

```bash
magic -dnull -noconsole << 'EOF'
tech load sky130A
lef read tech.lef
lef read cells.lef
def read design.def
quit
EOF
```

#### OpenROAD DEF Check

```bash
openroad << 'EOF'
read_lef tech.lef
read_lef cells.lef
read_def design.def
check_placement -verbose
exit
EOF
```

### 9.5 Design Statistics

#### Yosys `stat` Command

```bash
# Basic statistics
yosys -p "read_verilog design.v; synth -top top; stat"

# With Liberty for area estimation
yosys -p "read_verilog design.v; synth -top top; stat -liberty cells.lib"

# JSON output for parsing
yosys -p "read_verilog design.v; synth -top top; stat -json" > stats.json

# Technology-specific estimation
yosys -p "read_verilog design.v; synth -top top; stat -tech cmos"
```

#### Yosys `stat` Output Includes

- Wire count and bits
- Port count and bits
- Cell count by type
- Area (with Liberty file)
- Module hierarchy

#### OpenROAD Design Reports

```tcl
# Design area
report_design_area

# Cell usage
report_cell_usage

# Instance count
report_instance_count

# Power grid analysis
report_pdn
```

### 9.6 Complete Validation Script

```bash
#!/bin/bash
# validate_eda_files.sh - Validate all generated EDA files

LEF_FILE=${1:-cells.lef}
LIB_FILE=${2:-cells.lib}
VERILOG_FILE=${3:-design.v}
DEF_FILE=${4:-design.def}
TOP_MODULE=${5:-top}

echo "=== Validating Verilog ==="
verilator --lint-only --top-module $TOP_MODULE $VERILOG_FILE && \
yosys -q -p "read_verilog $VERILOG_FILE; hierarchy -check -top $TOP_MODULE; check -assert"

echo "=== Validating Liberty ==="
cat > /tmp/check_lib.tcl << EOF
read_liberty -lib $LIB_FILE
report_units
exit
EOF
sta /tmp/check_lib.tcl

echo "=== Validating LEF ==="
yosys -q -p "read_liberty -lib $LIB_FILE"

echo "=== Validating with Yosys stat ==="
yosys -p "read_verilog $VERILOG_FILE; synth -top $TOP_MODULE; stat -liberty $LIB_FILE"

echo "=== All validations complete ==="
```

### 9.7 Batch Image Generation

Generate images for all cells in a library:

```bash
#!/bin/bash
# generate_cell_images.sh

GDS_FILE=$1
OUTPUT_DIR=${2:-./images}
mkdir -p $OUTPUT_DIR

# Get list of cells from GDS
klayout -b -r - << 'RUBY' -rd gds=$GDS_FILE
layout = RBA::Layout.new
layout.read($gds)
layout.each_cell { |cell| puts cell.name }
RUBY | while read cell; do
    klayout -z -r /tmp/screenshot.rb \
        -rd input="$GDS_FILE" \
        -rd cell="$cell" \
        -rd output="$OUTPUT_DIR/${cell}.png"
done
```

### 9.8 Format Conversion Reference

| From | To | Tool | Command |
|------|-----|------|---------|
| Verilog | JSON | Yosys | `write_json out.json` |
| Verilog | BLIF | Yosys | `write_blif out.blif` |
| Verilog | EDIF | Yosys | `write_edif out.edif` |
| GDS | OASIS | KLayout | `klayout -b -r convert.rb` |
| OASIS | GDS | KLayout | `klayout -b -r convert.rb` |
| GDS | CIF | Magic | `cif write out.cif` |
| CIF | GDS | Magic | `gds write out.gds` |
| DEF | GDS | Magic | `gds write out.gds` |
| Layout | SPICE | Magic | `ext2spice` |
| Layout | PNM | Magic | `plot pnm out.pnm` |
| GDS | PNG | KLayout | `view.save_image()` |

### 9.9 CI/CD Integration Examples

#### GitHub Actions Validation

```yaml
# .github/workflows/validate-eda.yml
name: Validate EDA Files
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install tools
        run: |
          sudo apt-get install -y yosys verilator
      - name: Validate Verilog
        run: |
          verilator --lint-only --top-module top design.v
          yosys -p "read_verilog design.v; check -assert"
```

#### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Validate all .v files
for f in $(git diff --cached --name-only | grep '\.v$'); do
    verilator --lint-only "$f" || exit 1
done
```

### 9.10 Circuit Visualization (Yosys)

Yosys can generate circuit diagrams via Graphviz.

#### `show` Command - Schematic Generation

```bash
# Generate SVG schematic
yosys -p "read_verilog design.v; show -format svg -prefix schematic"

# Generate DOT file only
yosys -p "read_verilog design.v; show -format dot -prefix circuit"

# With library for proper cell display
yosys -p "read_liberty -lib cells.lib; read_verilog design.v; show -format svg"
```

#### `show` Command Options

```tcl
show                              # Open in xdot viewer (interactive)
show -format svg                  # Generate SVG file
show -format png                  # Generate PNG file
show -format pdf                  # Generate PDF file
show -format dot                  # Generate DOT file only
show -prefix <name>               # Output file prefix
show -viewer none                 # Don't open viewer
show -lib <verilog_file>          # Use library for cell display
show -stretch                     # Inputs left, outputs right
show -width                       # Annotate word widths
show -signed                      # Mark signed signals
show -colors <N>                  # Use N random colors
show -notitle                     # Suppress title
```

#### `viz` Command - Data Flow Graph

```tcl
# Visualize selected wires/signals
select -set path sig_a sig_b sig_c
viz -set path

# Generate to file
viz -format svg -prefix dataflow
```

#### Batch Schematic Generation

```bash
#!/bin/bash
# generate_schematics.sh
for v in *.v; do
    base=$(basename "$v" .v)
    yosys -q -p "read_verilog $v; hierarchy -auto-top; show -format svg -prefix ${base}_schematic -viewer none"
done
```

### 9.11 OpenROAD Image Export

OpenROAD can export layout images without GUI display.

```tcl
# Save layout image (works in batch mode)
save_image output.png

# With resolution control (microns per pixel)
save_image -resolution 0.1 output.png

# Specific area (x0 y0 x1 y1 in microns)
save_image -area {0 0 100 100} output.png

# With width specification
save_image -width 2000 output.png

# With display options
save_image -display_option {Layers metal1} output.png
```

#### Animated GIF Generation

```tcl
# Create animated GIF of flow stages
save_animated_gif -start -delay 500 flow_animation.gif
# ... run placement steps ...
save_animated_gif -add
# ... run routing steps ...
save_animated_gif -add
save_animated_gif -end
```

#### Timing Histogram Image

```tcl
save_histogram_image timing_hist.png -width 800 -height 600
```

### 9.12 KLayout Layer Operations (Batch)

Useful for post-processing generated layouts.

#### Boolean Operations Script

```ruby
# boolean_ops.rb - Run with: klayout -b -r boolean_ops.rb -rd input=design.gds -rd output=result.gds
layout = RBA::Layout.new
layout.read($input)

cell = layout.top_cell
layer1 = layout.layer(1, 0)
layer2 = layout.layer(2, 0)
result_layer = layout.layer(10, 0)

# Get regions
r1 = RBA::Region.new(cell.begin_shapes_rec(layer1))
r2 = RBA::Region.new(cell.begin_shapes_rec(layer2))

# Boolean AND
result = r1 & r2
cell.shapes(result_layer).insert(result)

layout.write($output)
```

#### Layer Sizing/Biasing

```ruby
# Size (expand) a layer
region = RBA::Region.new(cell.begin_shapes_rec(layer))
sized = region.sized(100)  # 100 database units (typically nm)
cell.shapes(output_layer).insert(sized)

# Shrink
shrunk = region.sized(-50)
```

#### Merge Overlapping Shapes

```ruby
region = RBA::Region.new(cell.begin_shapes_rec(layer))
merged = region.merged
cell.clear(layer)
cell.shapes(layer).insert(merged)
```

#### DRC Script Boolean Operations

```ruby
# In .drc file
metal1 = input(68, 20)
metal2 = input(69, 20)
via = input(68, 44)

# Boolean operations
overlap = metal1 & metal2                    # AND
diff = metal1 - metal2                       # NOT
xor_result = metal1 ^ metal2                 # XOR

# Sizing
expanded = metal1.sized(0.1.um)              # Expand by 100nm
shrunk = metal1.sized(-0.05.um)              # Shrink by 50nm

# Output results
overlap.output(100, 0)
expanded.output(101, 0)
```

### 9.13 Magic DRC Batch Reports

Generate DRC reports in batch mode.

```bash
#!/bin/bash
# drc_report.sh
magic -dnull -noconsole << 'EOF'
tech load sky130A
gds read design.gds
load topcell
select top cell
drc catchup
drc count total
drc listall why
quit
EOF
```

#### DRC Commands Reference

```tcl
drc on                    # Enable continuous DRC
drc off                   # Disable continuous DRC
drc check                 # Force recheck of box area
drc catchup               # Complete all pending DRC
drc find                  # Move box to next error
drc find [n]              # Move to nth error
drc why                   # Explain errors in box
drc count                 # Count errors (cell, count)
drc count total           # Total error count only
drc listall why           # List all errors with explanations
drc style drc(full)       # Full DRC mode
drc euclidean on          # Enable Euclidean distance
```

### 9.14 JSON Netlist Analysis (Yosys)

Export and analyze netlists programmatically.

```bash
# Generate JSON netlist
yosys -p "read_verilog design.v; synth -top top; write_json netlist.json"

# With AIG models
yosys -p "read_verilog design.v; synth -top top; write_json -aig netlist.json"
```

#### Python Analysis Script

```python
#!/usr/bin/env python3
# analyze_netlist.py
import json
import sys

with open(sys.argv[1]) as f:
    netlist = json.load(f)

for mod_name, mod in netlist['modules'].items():
    print(f"Module: {mod_name}")
    print(f"  Ports: {len(mod.get('ports', {}))}")
    print(f"  Cells: {len(mod.get('cells', {}))}")

    # Count cell types
    cell_types = {}
    for cell_name, cell in mod.get('cells', {}).items():
        t = cell['type']
        cell_types[t] = cell_types.get(t, 0) + 1

    print("  Cell types:")
    for t, count in sorted(cell_types.items()):
        print(f"    {t}: {count}")
```

#### Use Cases for Module 6

```bash
# Validate array structure
yosys -p "
  read_verilog fecim_array.v
  hierarchy -check -top fecim_array_4x4
  check -assert
  stat
  write_json -aig array_netlist.json
"

# Compare before/after synthesis
yosys -p "read_verilog design.v; write_json pre_synth.json"
yosys -p "read_verilog design.v; synth -top top; write_json post_synth.json"
# Then diff the JSON files
```

### 9.15 Useful One-Liners

```bash
# Quick Verilog syntax check
yosys -p "read_verilog design.v" 2>&1 | grep -i error

# Count cells after synthesis
yosys -p "read_verilog design.v; synth -top top; stat" | grep "cells:"

# Generate all format exports
yosys -p "read_verilog design.v; synth -top top; \
  write_verilog synth.v; write_json synth.json; \
  write_blif synth.blif; stat -json" > report.json

# Quick Liberty validation
sta -exit -f <(echo "read_liberty -lib cells.lib; report_units")

# Generate GDS thumbnail
klayout -z -r - -rd in=design.gds -rd out=thumb.png << 'RUBY'
mw = RBA::Application.instance.main_window
mw.load_layout($in, 0)
mw.current_view.max_hier
mw.current_view.save_image($out, 256, 256)
RUBY

# Magic: Quick area calculation
magic -dnull -noconsole << 'EOF' | grep "Total area"
load topcell
select top cell
box
quit
EOF
```

---

## 10. Quick Reference Tables

### 10.1 File Formats

| Format | Extension | Tool | Purpose |
|--------|-----------|------|---------|
| Verilog | .v | Yosys, OpenROAD | RTL/netlist |
| Liberty | .lib | Yosys, OpenSTA | Timing library |
| LEF | .lef | OpenROAD, Magic | Cell abstracts |
| DEF | .def | OpenROAD, Magic | Placement/routing |
| GDSII | .gds | Magic, KLayout | Physical layout |
| OASIS | .oas | KLayout | Compressed layout |
| SPICE | .spice, .sp | Netgen, Magic | Circuit netlist |
| SDC | .sdc | OpenSTA, OpenROAD | Timing constraints |
| SPEF | .spef | OpenSTA | Parasitics |
| SDF | .sdf | OpenSTA | Delay file |

### 10.2 Common CLI Patterns

| Task | OpenLane 2 | OpenLane 1 |
|------|------------|------------|
| Run flow | `openlane config.json` | `./flow.tcl -design name` |
| Interactive | `openlane -i config.json` | `./flow.tcl -interactive` |
| Smoke test | `openlane --smoke-test` | N/A |
| Docker | `openlane --dockerized` | `make mount` |

### 10.3 OpenLane Directory Structure

```
designs/
└── mydesign/
    ├── config.json          # OpenLane 2 config
    ├── config.tcl           # OpenLane 1 config
    ├── src/
    │   └── design.v         # RTL source
    └── runs/
        └── RUN_*/
            ├── logs/        # Tool logs
            ├── reports/     # Analysis reports
            ├── results/     # Output files
            └── tmp/         # Intermediate files
```

### 10.4 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PDK_ROOT` | PDK installation path | `/usr/share/pdk` |
| `PDK` | Active PDK variant | `sky130A` |
| `STD_CELL_LIBRARY` | Standard cell library | `sky130_fd_sc_hd` |
| `OPENLANE_ROOT` | OpenLane installation | `~/OpenLane` |
| `KLAYOUT_PATH` | KLayout search paths | `~/.klayout` |

---

## 11. Module 6 CLI Cheatsheet

**Essential CLI commands for the FeCIM Array Builder workflow.**

Module 6 generates: `fecim_bitcell.lef`, `fecim_bitcell.lib`, `fecim_bitcell.v`, `fecim_array_NxM.v`, `fecim_array_NxM.def`, and `config.json`.

---

### 11.1 Validate Generated Files

#### Verilog Validation (Cell + Array)

```bash
# Quick syntax check
yosys -p "read_verilog cells/fecim_bitcell/fecim_bitcell.v"
yosys -p "read_verilog data/lattice.v"

# Full validation with hierarchy check
yosys -p "
  read_verilog cells/fecim_bitcell/fecim_bitcell.v
  read_verilog data/lattice.v
  hierarchy -check -top fecim_array_4x4
  check -assert
"

# Pre-check with Verilator (recommended)
verilator --lint-only cells/fecim_bitcell/fecim_bitcell.v
verilator --lint-only data/lattice.v
```

#### Liberty (.lib) Validation

```bash
# OpenSTA validation
sta << 'EOF'
read_liberty -lib cells/fecim_bitcell/fecim_bitcell.lib
report_units
report_lib fecim_bitcell_lib
exit
EOF

# Quick one-liner
sta -exit -f <(echo "read_liberty -lib cells/fecim_bitcell/fecim_bitcell.lib; report_units")
```

#### LEF Validation

```bash
# With Magic
magic -dnull -noconsole << 'EOF'
lef read cells/fecim_bitcell/fecim_bitcell.lef
quit
EOF

# With OpenROAD
openroad << 'EOF'
read_lef cells/fecim_bitcell/fecim_bitcell.lef
exit
EOF
```

#### DEF Validation

```bash
# Validate DEF with LEF dependency
magic -dnull -noconsole << 'EOF'
lef read cells/fecim_bitcell/fecim_bitcell.lef
def read data/placement.def
quit
EOF
```

---

### 11.2 Design Statistics

```bash
# Cell statistics
yosys -p "
  read_verilog cells/fecim_bitcell/fecim_bitcell.v
  stat
"

# Array statistics with Liberty
yosys -p "
  read_liberty -lib cells/fecim_bitcell/fecim_bitcell.lib
  read_verilog cells/fecim_bitcell/fecim_bitcell.v
  read_verilog data/lattice.v
  hierarchy -top fecim_array_4x4
  stat -liberty cells/fecim_bitcell/fecim_bitcell.lib
"

# JSON output for programmatic analysis
yosys -p "
  read_verilog data/lattice.v
  hierarchy -top fecim_array_4x4
  stat -json
" > array_stats.json
```

---

### 11.3 Generate Documentation Images

#### Array Schematic (Yosys → SVG)

```bash
# Generate SVG schematic of array
yosys -p "
  read_verilog cells/fecim_bitcell/fecim_bitcell.v
  read_verilog data/lattice.v
  hierarchy -top fecim_array_4x4
  show -format svg -prefix docs/array_schematic -viewer none
"

# Cell-only schematic
yosys -p "
  read_verilog cells/fecim_bitcell/fecim_bitcell.v
  show -format svg -prefix docs/cell_schematic -viewer none
"
```

#### Layout Image (KLayout → PNG)

```bash
# Generate PNG from DEF (requires GDS or visual DEF)
cat > /tmp/def_screenshot.rb << 'RUBY'
mw = RBA::Application::instance.main_window
# Load LEF first for cell definitions
options = RBA::LoadLayoutOptions::new
lefdef = RBA::LEFDEFReaderConfiguration::new
lefdef.read_lef_with_def = true
options.lefdef_config = lefdef
mw.load_layout($input, options, 0)
view = mw.current_view
view.max_hier
view.set_config("background-color", "#ffffff")
view.save_image($output, 2000, 2000)
RUBY

klayout -z -r /tmp/def_screenshot.rb \
  -rd input=data/placement.def \
  -rd output=docs/array_layout.png
```

---

### 11.4 Test OpenLane Integration

#### Syntax Test with OpenLane Tools

```bash
# Test full file set loads correctly
openroad << 'EOF'
read_lef cells/fecim_bitcell/fecim_bitcell.lef
read_liberty cells/fecim_bitcell/fecim_bitcell.lib
read_verilog data/lattice.v
link_design fecim_array_4x4
report_design_area
exit
EOF
```

#### Simulate OpenLane Config

```bash
# Validate config.json structure
python3 -c "
import json
with open('data/config.json') as f:
    cfg = json.load(f)
    print('Design:', cfg.get('DESIGN_NAME', 'NOT SET'))
    print('Verilog:', cfg.get('VERILOG_FILES', 'NOT SET'))
    print('Clock:', cfg.get('CLOCK_PORT', 'NOT SET'))
"
```

---

### 11.5 Complete Validation Script

```bash
#!/bin/bash
# validate_module6_output.sh
# Run from fecim-lattice-tools root

set -e
echo "=== Module 6 Output Validation ==="

CELL_DIR="cells/fecim_bitcell"
GEN_DIR="data"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

# 1. Verilog syntax
echo -e "\n[1/5] Verilog Validation..."
yosys -q -p "read_verilog $CELL_DIR/fecim_bitcell.v" && pass "Cell Verilog OK" || fail "Cell Verilog FAILED"
yosys -q -p "read_verilog $GEN_DIR/lattice.v" && pass "Array Verilog OK" || fail "Array Verilog FAILED"

# 2. Liberty validation
echo -e "\n[2/5] Liberty Validation..."
sta -exit -f <(echo "read_liberty -lib $CELL_DIR/fecim_bitcell.lib; exit") 2>/dev/null && \
  pass "Liberty OK" || fail "Liberty FAILED"

# 3. LEF validation
echo -e "\n[3/5] LEF Validation..."
magic -dnull -noconsole << EOF 2>/dev/null && pass "LEF OK" || fail "LEF FAILED"
lef read $CELL_DIR/fecim_bitcell.lef
quit
EOF

# 4. DEF validation
echo -e "\n[4/5] DEF Validation..."
magic -dnull -noconsole << EOF 2>/dev/null && pass "DEF OK" || fail "DEF FAILED"
lef read $CELL_DIR/fecim_bitcell.lef
def read $GEN_DIR/placement.def
quit
EOF

# 5. Full integration test
echo -e "\n[5/5] Integration Test..."
yosys -q -p "
  read_liberty -lib $CELL_DIR/fecim_bitcell.lib
  read_verilog $CELL_DIR/fecim_bitcell.v
  read_verilog $GEN_DIR/lattice.v
  hierarchy -check
  check -assert
" && pass "Integration OK" || fail "Integration FAILED"

echo -e "\n${GREEN}=== All validations passed ===${NC}"
```

---

### 11.6 Quick Reference Card

| Task | Command |
|------|---------|
| **Validate cell Verilog** | `yosys -p "read_verilog fecim_bitcell.v"` |
| **Validate array Verilog** | `yosys -p "read_verilog lattice.v; hierarchy -check"` |
| **Validate Liberty** | `sta -exit -f <(echo "read_liberty -lib file.lib")` |
| **Validate LEF** | `magic -dnull -noconsole -c "lef read file.lef; quit"` |
| **Validate DEF** | `magic -dnull -noconsole -c "lef read cell.lef; def read file.def; quit"` |
| **Cell count** | `yosys -p "read_verilog design.v; stat"` |
| **Generate schematic SVG** | `yosys -p "read_verilog design.v; show -format svg -prefix out"` |
| **Check hierarchy** | `yosys -p "read_verilog *.v; hierarchy -check -top top"` |
| **JSON netlist** | `yosys -p "read_verilog design.v; write_json out.json"` |
| **Full lint** | `verilator --lint-only -Wall design.v` |

---

### 11.7 Troubleshooting Common Issues

#### "Module not found" in Yosys
```bash
# Ensure cell module is read before array
yosys -p "
  read_verilog cells/fecim_bitcell/fecim_bitcell.v  # Cell first
  read_verilog data/lattice.v                   # Array second
  hierarchy -check -top fecim_array_4x4
"
```

#### "Unknown cell type" in Liberty/LEF
```bash
# Check cell name matches across files
grep "^cell(" cells/fecim_bitcell/fecim_bitcell.lib
grep "^MACRO" cells/fecim_bitcell/fecim_bitcell.lef
grep "^module" cells/fecim_bitcell/fecim_bitcell.v
# All should show: fecim_bitcell
```

#### "Port mismatch" errors
```bash
# Compare ports across formats
yosys -p "read_verilog fecim_bitcell.v; ls" | grep -A20 "module fecim_bitcell"
grep "PIN" fecim_bitcell.lef
grep "pin(" fecim_bitcell.lib
```

#### DEF placement issues
```bash
# Check COMPONENTS section references correct cell
grep "COMPONENTS" data/placement.def -A5
# Should show: fecim_bitcell (matching MACRO name in LEF)
```

---

### 11.8 Automated CI/CD for Module 6

```yaml
# .github/workflows/validate-eda-output.yml
name: Validate EDA Output
on:
  push:
    paths:
      - 'cells/**'
      - 'data/**'
      - 'module6-eda/**'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install EDA tools
        run: |
          sudo apt-get update
          sudo apt-get install -y yosys verilator

      - name: Validate Verilog
        run: |
          verilator --lint-only cells/fecim_bitcell/fecim_bitcell.v
          yosys -p "read_verilog cells/fecim_bitcell/fecim_bitcell.v; check -assert"

      - name: Validate Array
        run: |
          yosys -p "
            read_verilog cells/fecim_bitcell/fecim_bitcell.v
            read_verilog data/lattice.v
            hierarchy -check
            check -assert
          "

      - name: Generate Stats
        run: |
          yosys -p "
            read_verilog data/lattice.v
            stat -json
          " > stats.json
          cat stats.json
```

---

## Sources

- [OpenLane 2 Documentation](https://openlane2.readthedocs.io/)
- [OpenLane GitHub](https://github.com/The-OpenROAD-Project/OpenLane)
- [Yosys Documentation](https://yosyshq.readthedocs.io/projects/yosys/en/latest/)
- [Yosys GitHub](https://github.com/YosysHQ/yosys)
- [OpenROAD Documentation](https://openroad.readthedocs.io/en/latest/)
- [OpenROAD Flow Scripts](https://openroad-flow-scripts.readthedocs.io/en/latest/)
- [Magic VLSI](http://opencircuitdesign.com/magic/)
- [Magic GitHub](https://github.com/RTimothyEdwards/magic)
- [KLayout](https://www.klayout.de/)
- [Netgen](http://opencircuitdesign.com/netgen/)
- [Netgen GitHub](https://github.com/RTimothyEdwards/netgen)
- [OpenSTA GitHub](https://github.com/The-OpenROAD-Project/OpenSTA)
- [SkyWater PDK](https://github.com/google/skywater-pdk)
- [Open_PDKs](https://github.com/RTimothyEdwards/open_pdks)
- [GF180MCU PDK](https://github.com/google/gf180mcu-pdk)
- [OpenLane WOSET Paper](https://woset-workshop.github.io/PDFs/2020/a21.pdf)
- [TritonRoute](https://github.com/The-OpenROAD-Project/TritonRoute)
- [ABC Toolbox](https://yosyshq.readthedocs.io/projects/yosys/en/latest/using_yosys/synthesis/abc.html)

---

*Document generated from web research on 2026-01-26*
