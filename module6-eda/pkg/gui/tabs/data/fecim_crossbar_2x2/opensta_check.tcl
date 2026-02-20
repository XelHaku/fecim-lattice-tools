# FeCIM OpenSTA Timing Analysis Script
# Generated: 2026-02-20
# Design:    fecim_crossbar_2x2
# Array:     2x2 passive
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
read_liberty cells/fecim_bitcell/fecim_bitcell.lib

# ── Read netlist ──────────────────────────────────────────────────────────────
read_verilog fecim_crossbar_2x2.v
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
puts "OpenSTA analysis complete for fecim_crossbar_2x2"
puts "No timing violations expected for clockless FeCIM array."
