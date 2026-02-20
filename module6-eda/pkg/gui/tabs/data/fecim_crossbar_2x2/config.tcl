########################################################################
# FeCIM OpenLane v1 Configuration
# Generated: 2026-02-20
# Design:    fecim_crossbar_2x2
# Array:     2x2 cells, arch=passive, tech=sky130
#
# OpenLane v1 native config.tcl format (all vars in ::env namespace).
# For LibreLane (OpenLane v2+), use config.json instead.
#
# Usage:
#   export PDK=sky130A PDK_ROOT=/path/to/pdk
#   cd /path/to/design && flow.tcl -design . -tag run001 -overwrite
#
# Required files (pre-generate from fecim-eda CLI):
#   <DESIGN_NAME>.v    — structural Verilog (array instantiation)
#   fecim_bitcell.lef  — cell abstract (pins, boundary, obstruction)
#   fecim_bitcell.spice — SPICE netlist (for LVS)
#   constraints.sdc    — timing constraints
#
########################################################################

# ── Design identity ──────────────────────────────────────────────────────────
set ::env(DESIGN_NAME) "fecim_crossbar_2x2"
set ::env(DESIGN_DIR) [file dirname [file normalize [info script]]]

# ── PDK ──────────────────────────────────────────────────────────────────────
set ::env(PDK) "sky130A"
set ::env(PDK_VARIANT) "sky130A"
# SKY130 PDK paths (set PDK_ROOT before running):
# export PDK_ROOT=~/.volare
# volare enable --pdk sky130 sky130A
set ::env(STD_CELL_LIBRARY) sky130_fd_sc_hd

# ── Input files ──────────────────────────────────────────────────────────────
set ::env(VERILOG_FILES) "$::env(DESIGN_DIR)/$::env(DESIGN_NAME).v"
set ::env(BASE_SDC_FILE) "$::env(DESIGN_DIR)/constraints.sdc"

# Hard macro (FeCIM cell is a pre-characterized GDS/LEF macro):
set ::env(EXTRA_LEFS) [list \
    "$::env(DESIGN_DIR)/fecim_bitcell.lef" \
]
set ::env(EXTRA_GDS_FILES) [list \
    "$::env(DESIGN_DIR)/fecim_bitcell.gds" \
]

# ── Synthesis ─────────────────────────────────────────────────────────────────
# FeCIM array is a hard macro — only peripheral digital logic is synthesized.
set ::env(SYNTH_READ_BLACKBOX_LIB) 1          ;# FeCIM cell is a blackbox
set ::env(SYNTH_STRATEGY) "AREA 0"            ;# Area-optimized (alternatives: DELAY, balanced)
set ::env(SYNTH_CLOCK_UNCERTAINTY) 0.25       ;# ns — synthesis timing margin
set ::env(SYNTH_CLOCK_TRANSITION) 0.15        ;# ns — clock ramp time
set ::env(SYNTH_FLAT_TOP) 0                   ;# Keep hierarchy
set ::env(SYNTH_EXTRA_MAPPING_FILE) ""        ;# No custom cell mapping

# ── Floorplan ─────────────────────────────────────────────────────────────────
set ::env(FP_SIZING) absolute                 ;# Explicit die area
set ::env(DIE_AREA) "0 0 40.92 45.44"           ;# µm: 2x2 array + margins
set ::env(CORE_AREA) "10.00 10.00 30.92 35.44"    ;# µm: 10 µm core margin
set ::env(FP_CORE_UTIL) 50                    ;# % target core utilization
set ::env(FP_ASPECT_RATIO) 1.0                ;# Square die
set ::env(FP_IO_MODE) 1                       ;# Equidistant I/O placement

# Power distribution network
set ::env(FP_PDN_AUTO_ADJUST) 1
set ::env(FP_PDN_ENABLE_MACROS_GRID) 1        ;# Connect FeCIM macro to PDN
set ::env(FP_PDN_CORE_RING) 0                 ;# No core ring (add if analog supply needed)
set ::env(FP_PDN_ENABLE_RAILS) 1
set ::env(FP_PDN_CHECK_NODES) 1

# Margins (in units of standard cell height)
set ::env(BOTTOM_MARGIN_MULT) 4
set ::env(TOP_MARGIN_MULT) 4
set ::env(LEFT_MARGIN_MULT) 12
set ::env(RIGHT_MARGIN_MULT) 12

# ── Placement ─────────────────────────────────────────────────────────────────
set ::env(PL_ROUTABILITY_DRIVEN) 1            ;# Minimize routing congestion
set ::env(PL_TIME_DRIVEN) 1                   ;# Timing-aware placement
set ::env(PL_RESIZER_DESIGN_OPTIMIZATIONS) 1  ;# Buffer/gate sizing
set ::env(PL_RESIZER_TIMING_OPTIMIZATIONS) 1  ;# Timing optimization
set ::env(PL_OPTIMIZE_MIRRORING) 1
set ::env(PL_MACRO_HALO) "0 0"               ;# No halo around FeCIM macro
set ::env(PL_MACRO_CHANNEL) "0 0"
set ::env(PL_WIRELENGTH_COEF) 0.25

# ── Clock Tree Synthesis ──────────────────────────────────────────────────────
set ::env(CLOCK_PORT) "clk"                   ;# Peripheral control clock
set ::env(CLOCK_PERIOD) "10.0"                ;# ns — 100 MHz
set ::env(RUN_CTS) 0                         ;# 0=disabled (passive array has no clock) for CTS

# CTS tuning (only relevant if RUN_CTS=1)
set ::env(CTS_TOLERANCE) 100
set ::env(CTS_SINK_CLUSTERING_SIZE) 25
set ::env(CTS_CLK_MAX_WIRE_LENGTH) 0

# ── Routing ───────────────────────────────────────────────────────────────────
set ::env(GLOBAL_ROUTER) fastroute
set ::env(DETAILED_ROUTER) tritonroute
set ::env(GRT_ADJUSTMENT) 0.3
set ::env(GRT_OVERFLOW_ITERS) 50
set ::env(GRT_ESTIMATE_PARASITICS) 1
set ::env(DRT_OPT_ITERS) 64
set ::env(GRT_MACRO_EXTENSION) 0

# ── Signoff (DRC / LVS) ───────────────────────────────────────────────────────
set ::env(RUN_MAGIC_DRC) 1
set ::env(RUN_KLAYOUT_DRC) 0                  ;# Optional; set 1 for double-check
set ::env(RUN_LVS) 1
set ::env(LVS_INSERT_POWER_PINS) 1
set ::env(MAGIC_GENERATE_GDS) 1
set ::env(MAGIC_DRC_USE_GDS) 1                ;# Run DRC on final GDS
set ::env(MAGIC_EXT_USE_GDS) 1                ;# Extract netlist from GDS

# ── Timing & Power ───────────────────────────────────────────────────────────
set ::env(STA_REPORT_POWER) 1
set ::env(RUN_SPEF_EXTRACTION) 1
set ::env(RUN_IRDROP_REPORT) 1

# ── Error control ─────────────────────────────────────────────────────────────
set ::env(QUIT_ON_UNMAPPED_CELLS) 1
set ::env(QUIT_ON_SYNTH_CHECKS) 1
set ::env(QUIT_ON_TIMING_VIOLATIONS) 0        ;# Relax — FeCIM write path is slow by design
set ::env(QUIT_ON_LVS_ERROR) 1
set ::env(QUIT_ON_MAGIC_DRC) 1

# ── Standard cell library ────────────────────────────────────────────────────
set ::env(STD_CELL_LIBRARY) "sky130_fd_sc_hd"

# ── FeCIM-specific notes ─────────────────────────────────────────────────────
# The FeCIM array is placed as a hard macro at the center of the die.
# Peripheral digital logic (row/column decoders, ADC interface) is synthesized.
# Ferroelectric write operations are slow (~50 ns); timing violations in the
# write path are expected and benign — QUIT_ON_TIMING_VIOLATIONS is disabled.
