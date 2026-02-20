// pkg/export/sdc.go
// SDC (Synopsys Design Constraints) generator for FeCIM crossbar arrays
//
// SDC is required by OpenLane/LibreLane and OpenSTA for timing analysis.
// For the FeCIM crossbar array itself:
//   - No clock exists (static/ISPP write, capacitive read)
//   - RUN_CTS=0 in config.json skips Clock Tree Synthesis
//   - Timing constraints are trivial (all paths are purely combinational I/O)
//
// If a digital control wrapper is added (FSM, address decoder, etc.),
// uncomment the clock section and set a realistic clock period.
//
// References:
//   OpenSTA TCL commands: https://openroad-flow-scripts.readthedocs.io/
//   SKY130 speed grade:   ~100 MHz at 1.8V typical corner
package export

import (
	"fmt"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GenerateSDC returns an SDC timing constraints file for the FeCIM array.
//
// The FeCIM crossbar has no clock, so constraints are minimal:
//   - Input/output delays set to 0 (no external timing requirements on array pins)
//   - Max transition matching FeFET write speed (~10 ns typical)
//   - Load capacitance from Liberty InputCap
//
// If the design is wrapped with a digital controller, use GenerateSDCWithClock.
func GenerateSDC(cfg config.ArrayConfig) string {
	tech := strings.ToLower(cfg.Technology)
	if tech == "" {
		tech = "sky130"
	}

	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	// FeFET write speed anchor (Trentzsch IEDM 2016: ~50 ns write, ~5 ns read)
	maxTransition := 10.0 // ns (conservative mid-range)
	loadCap := 0.015      // pF (FeFET mid-range input capacitance)

	return fmt.Sprintf(`# FeCIM SDC Timing Constraints
# Generated: %s
# Design:    %s
# Array:     %dx%d %s (%s)
#
# This SDC is appropriate for a pure FeCIM crossbar array with NO clock.
# - RUN_CTS=0 in config.json (no Clock Tree Synthesis)
# - All write/read paths are combinational (set by external DAC timing)
# - No setup/hold violations expected
#
# If you add a digital control wrapper (FSM, address decoder) UNCOMMENT
# the clock section below and set a realistic period.

# ── Current design: No clock ──────────────────────────────────────────────────
# The FeCIM array is driven by DAC outputs (word lines) and sensed
# by TIA/ADC (bit lines). Timing is governed by the peripheral circuits,
# not by the array itself.

# I/O delay: 0 ns (array pins connect directly to peripheral circuits)
set_input_delay  0.0 [all_inputs]
set_output_delay 0.0 [all_outputs]

# Max transition: FeFET write path constraint
# Reference: Trentzsch et al. IEDM 2016 (28nm FDSOI FeFET, ~50 ns write)
# Using conservative 10 ns here (read-dominated timing requirement)
set_max_transition %.1f [all_outputs]

# Load capacitance: FeFET input capacitance (for STA buffer sizing)
# Reference: FeFET mid-range input cap ~0.015 pF
set_load %.4f [all_outputs]

# ── Optional: Digital control wrapper clock ───────────────────────────────────
# Uncomment and set CLK_PERIOD if a control FSM is added around the array.
# SKY130 max speed grade: ~100 MHz at 1.8V typical corner.
# Suggested values for educational designs:
#   10 ns period → 100 MHz (fast, near sky130 limit)
#   20 ns period →  50 MHz (balanced, good timing margin)
#   40 ns period →  25 MHz (conservative, easy closure)
#
# set CLK_PERIOD 20.0
# create_clock -period $CLK_PERIOD -name clk [get_ports clk]
# set_clock_uncertainty 0.25 [all_clocks]
# set_clock_transition  0.15 [all_clocks]
# set_input_delay  [expr {$CLK_PERIOD * 0.15}] -clock clk [get_ports {WL[*]}]
# set_output_delay [expr {$CLK_PERIOD * 0.15}] -clock clk [get_ports {BL[*]}]
`,
		time.Now().Format("2006-01-02"),
		designName,
		cfg.Rows, cfg.Cols, cfg.Architecture, cfg.Technology,
		maxTransition, loadCap)
}

// GenerateOpenSTAScript returns a standalone OpenSTA TCL script for timing
// analysis of the FeCIM design. OpenSTA is embedded in OpenROAD and can
// also be run standalone: opensta < opensta_check.tcl
//
// For a clockless design this will report no paths (expected behavior).
func GenerateOpenSTAScript(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)
	cellName := cellNameFromArch(cfg.Architecture)

	return fmt.Sprintf(`# FeCIM OpenSTA Timing Analysis Script
# Generated: %s
# Design:    %s
# Array:     %dx%d %s
#
# Usage (standalone OpenSTA):
#   opensta < opensta_check.tcl
#
# Usage (embedded in OpenROAD):
#   openroad -no_splash -exit opensta_check.tcl
#
# Expected results for pure FeCIM array:
#   - No timing paths reported (clockless design)
#   - No setup/hold violations
#   - input/output delay = 0 (from constraints.sdc)

# ── Read technology ───────────────────────────────────────────────────────────
# FeCIM bitcell Liberty file (educational timing anchors from Trentzsch 2016)
read_liberty cells/%s/%s.lib

# ── Read netlist ──────────────────────────────────────────────────────────────
read_verilog output/%s.v
link_design fecim_crossbar

# ── Read constraints ──────────────────────────────────────────────────────────
read_sdc constraints.sdc

# ── Timing reports ────────────────────────────────────────────────────────────
puts ""
puts "=== Timing Report: Max paths ==="
report_checks -path_delay max -format full

puts ""
puts "=== Timing Report: Min paths ==="
report_checks -path_delay min -format full

puts ""
puts "=== Summary ==="
report_check_types -max_slew -max_cap -max_fanout

puts ""
puts "=== Cell Power (analytical estimate) ==="
# NOTE: power values come from Liberty leakage_power entries.
# These are educational anchors from Trentzsch 2016 / Muller 2013.
# Production power requires SPICE-based characterization.
if {[catch {report_power} err]} {
    puts "INFO: Power analysis requires clock definition: $err"
}

puts ""
puts "OpenSTA analysis complete for %s"
puts "No timing violations expected for clockless FeCIM array."
`,
		time.Now().Format("2006-01-02"),
		designName, cfg.Rows, cfg.Cols, cfg.Architecture,
		cellName, cellName, designName,
		designName)
}
