// pkg/export/openlane_tcl.go
// OpenLane v1 TCL configuration file generator for FeCIM crossbar arrays.
//
// OpenLane v1 (https://github.com/The-OpenROAD-Project/OpenLane) uses a
// TCL-based flow orchestrated by flow.tcl. All configuration is set via
// `set ::env(VARIABLE) value` syntax sourced into the TCL environment.
//
// LibreLane (v2+) uses JSON config files — see openlane_config.go.
// This file generates config.tcl specifically for OpenLane v1 native format.
//
// Flow steps controlled by this config:
//   1. Synthesis      (Yosys) — RTL to gate-level netlist
//   2. Floorplan      (OpenROAD) — die + core area, PDN
//   3. Placement      (OpenROAD) — global + detailed placement
//   4. CTS            (OpenROAD) — clock tree synthesis
//   5. Routing        (OpenROAD TritonRoute) — global + detailed routing
//   6. Signoff        (Magic DRC, Netgen LVS) — DRC/LVS verification
//   7. GDS generation (Magic + KLayout) — final layout
//
// References:
//   OpenLane v1 configuration: https://openlane.readthedocs.io/en/latest/reference/configuration/
//   flow.tcl variables: /opensource/OpenLane/flow.tcl
//   PDK config hierarchy: /configuration/*.tcl
package export

import (
	"fmt"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateOpenLaneTCLConfig returns an OpenLane v1 config.tcl for the FeCIM array.
//
// The generated config handles three FeCIM architectures:
//   - passive (0T1R): hard macro, no transistors, skip CTS
//   - 1t1r:  hard macro with selector transistor, minimal CTS
//   - 2t1r:  hard macro with dual selectors, minimal CTS
//
// All FeCIM arrays are treated as hard macros (GDS/LEF pre-provided).
// Digital peripheral logic (address decoders, control FSM) is synthesized.
func GenerateOpenLaneTCLConfig(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	cellName := cellNameFromArch(cfg.Architecture)

	tech := strings.ToLower(cfg.Technology)
	isPassive := strings.ToLower(cfg.Architecture) == "passive"

	// PDK-specific parameters (from local IHP and GF180 PDK exploration)
	var pdkName, pdkVariant, scLibrary string
	var clockPeriod float64
	var pdkSetup string

	switch {
	case strings.Contains(tech, "ihp") || strings.Contains(tech, "sg13"):
		pdkName = "sg13g2"
		pdkVariant = "sg13g2"
		scLibrary = "sg13g2_stdcell"
		clockPeriod = 10.0 // 100 MHz; ferroelectric write is slow (~50 ns)
		pdkSetup = `# IHP SG13G2 PDK paths (set IHP_PDK_ROOT before running):
# export IHP_PDK_ROOT=~/ihp-sg13g2-pdk
# git clone https://github.com/IHP-GmbH/IHP-Open-PDK.git ~/ihp-sg13g2-pdk
set ::env(PDK_ROOT) $::env(IHP_PDK_ROOT)
set ::env(STD_CELL_LIBRARY) sg13g2_stdcell`
	case strings.Contains(tech, "gf180"):
		pdkName = "gf180mcuD"
		pdkVariant = "gf180mcuD"
		scLibrary = "gf180mcu_fd_sc_mcu7t5v0"
		clockPeriod = 10.0
		pdkSetup = `# GF180MCU PDK paths (set PDK_ROOT before running):
# export PDK_ROOT=~/.volare
# volare enable --pdk gf180mcuD gf180mcuD
set ::env(STD_CELL_LIBRARY) gf180mcu_fd_sc_mcu7t5v0`
	default: // sky130
		pdkName = "sky130A"
		pdkVariant = "sky130A"
		scLibrary = "sky130_fd_sc_hd"
		clockPeriod = 10.0
		pdkSetup = `# SKY130 PDK paths (set PDK_ROOT before running):
# export PDK_ROOT=~/.volare
# volare enable --pdk sky130 sky130A
set ::env(STD_CELL_LIBRARY) sky130_fd_sc_hd`
	}

	// Estimate die area from array footprint + 20 µm margin on each side
	arrayW := float64(cfg.Cols) * cfg.CellWidth
	arrayH := float64(cfg.Rows) * cfg.CellHeight
	margin := 20.0 // µm
	dieW := arrayW + 2*margin
	dieH := arrayH + 2*margin
	coreMargin := 10.0 // µm
	coreX1 := coreMargin
	coreY1 := coreMargin
	coreX2 := dieW - coreMargin
	coreY2 := dieH - coreMargin

	// Core utilization: FeCIM crossbar is mostly the hard macro
	coreUtil := 60
	if isPassive {
		coreUtil = 50 // More routing headroom for passive (sneak-path sensitive)
	}

	// CTS: skip for fully passive arrays (no clock)
	runCTS := 1
	if isPassive {
		runCTS = 0
	}

	// IR drop: check is useful for any design
	runIRDrop := 1

	return fmt.Sprintf(`########################################################################
# FeCIM OpenLane v1 Configuration
# Generated: %s
# Design:    %s
# Array:     %dx%d cells, arch=%s, tech=%s
#
# OpenLane v1 native config.tcl format (all vars in ::env namespace).
# For LibreLane (OpenLane v2+), use config.json instead.
#
# Usage:
#   export PDK=%s PDK_ROOT=/path/to/pdk
#   cd /path/to/design && flow.tcl -design . -tag run001 -overwrite
#
# Required files (pre-generate from fecim-eda CLI):
#   design.v           — structural Verilog (array instantiation)
#   %s.lef  — cell abstract (pins, boundary, obstruction)
#   %s.spice — SPICE netlist (for LVS)
#   constraints.sdc    — timing constraints
#
########################################################################

# ── Design identity ──────────────────────────────────────────────────────────
set ::env(DESIGN_NAME) "%s"
set ::env(DESIGN_DIR) [file dirname [file normalize [info script]]]

# ── PDK ──────────────────────────────────────────────────────────────────────
set ::env(PDK) "%s"
set ::env(PDK_VARIANT) "%s"
%s

# ── Input files ──────────────────────────────────────────────────────────────
set ::env(VERILOG_FILES) "$::env(DESIGN_DIR)/design.v"
set ::env(BASE_SDC_FILE) "$::env(DESIGN_DIR)/constraints.sdc"

# Hard macro (FeCIM cell is a pre-characterized GDS/LEF macro):
set ::env(EXTRA_LEFS) [list \
    "$::env(DESIGN_DIR)/%s.lef" \
]
set ::env(EXTRA_GDS_FILES) [list \
    "$::env(DESIGN_DIR)/%s.gds" \
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
set ::env(DIE_AREA) "0 0 %.2f %.2f"           ;# µm: %dx%d array + margins
set ::env(CORE_AREA) "%.2f %.2f %.2f %.2f"    ;# µm: 10 µm core margin
set ::env(FP_CORE_UTIL) %d                    ;# %% target core utilization
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
set ::env(CLOCK_PERIOD) "%.1f"                ;# ns — %.0f MHz
set ::env(RUN_CTS) %d                         ;# %s for CTS

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
set ::env(RUN_IRDROP_REPORT) %d

# ── Error control ─────────────────────────────────────────────────────────────
set ::env(QUIT_ON_UNMAPPED_CELLS) 1
set ::env(QUIT_ON_SYNTH_CHECKS) 1
set ::env(QUIT_ON_TIMING_VIOLATIONS) 0        ;# Relax — FeCIM write path is slow by design
set ::env(QUIT_ON_LVS_ERROR) 1
set ::env(QUIT_ON_MAGIC_DRC) 1

# ── Standard cell library ────────────────────────────────────────────────────
set ::env(STD_CELL_LIBRARY) "%s"

# ── FeCIM-specific notes ─────────────────────────────────────────────────────
# The FeCIM array is placed as a hard macro at the center of the die.
# Peripheral digital logic (row/column decoders, ADC interface) is synthesized.
# Ferroelectric write operations are slow (~50 ns); timing violations in the
# write path are expected and benign — QUIT_ON_TIMING_VIOLATIONS is disabled.
`,
		// Header
		time.Now().Format("2006-01-02"),
		designName,
		cfg.Rows, cfg.Cols, cfg.Architecture, cfg.Technology,
		// PDK and usage
		pdkName,
		cellName,
		cellName,
		// Design name
		designName,
		// PDK vars
		pdkName, pdkVariant,
		pdkSetup,
		// Input files (LEF + GDS)
		cellName,
		cellName,
		// Die area (comment args: rows×cols for conventional notation)
		dieW, dieH, cfg.Rows, cfg.Cols,
		// Core area
		coreX1, coreY1, coreX2, coreY2,
		// Core util
		coreUtil,
		// Clock
		clockPeriod, 1000.0/clockPeriod,
		runCTS,
		map[bool]string{true: "1=enabled", false: "0=disabled (passive array has no clock)"}[runCTS == 1],
		// IR drop
		runIRDrop,
		// Std cell library
		scLibrary,
	)
}

// GenerateOpenLaneTCLMacroPlacement returns a macro placement constraint file
// (macros.cfg) that pins the FeCIM array to the center of the die.
// OpenLane reads this via MACRO_PLACEMENT_CFG.
func GenerateOpenLaneTCLMacroPlacement(cfg config.ArrayConfig) string {
	cellName := cellNameFromArch(cfg.Architecture)
	// Place macro at center of die (die origin at 0,0)
	margin := 20.0
	centerX := float64(cfg.Cols)*cfg.CellWidth/2 + margin
	centerY := float64(cfg.Rows)*cfg.CellHeight/2 + margin

	return fmt.Sprintf(`# FeCIM Macro Placement Constraints
# Generated: %s
# Format: <instance_name> <X_µm> <Y_µm> <orientation>
# Orientations: N, S, E, W, FN, FS, FE, FW
#
# Place the FeCIM crossbar array at the die center.
# Set in config.tcl: set ::env(MACRO_PLACEMENT_CFG) "macros.cfg"

%s_0 %.2f %.2f N
`, time.Now().Format("2006-01-02"), cellName, centerX, centerY)
}
