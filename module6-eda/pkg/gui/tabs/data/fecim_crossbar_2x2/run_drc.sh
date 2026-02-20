#!/usr/bin/env bash
# FeCIM Magic DRC Script
# Generated: 2026-02-20
# Array: 2x2, Architecture: passive, Technology: sky130
#
# Runs Magic DRC (Design Rule Check) on the FeCIM cell layout.
# DRC errors must be resolved before tapeout.
#
# Prerequisites:
#   - magic installed (apt: magic, or build from source)
#   - PDK technology file available (set PDK_ROOT env var)
#   - Layout file: cells/fecim_bitcell/fecim_bitcell.mag (Magic layout format)
#   - OR GDS file:  cells/fecim_bitcell/fecim_bitcell.gds (for GDS-based DRC)
#
# Usage:
#   export PDK_ROOT=/path/to/pdk
#   bash run_drc.sh
#
# NOTE: This is for abstract/educational layouts. Production tapeout requires
# full transistor-level layout with foundry DRC sign-off.

set -e

CELL="fecim_bitcell"
DESIGN="fecim_crossbar_2x2"
TECH_FILE="${PDK_ROOT}/sky130A/libs.tech/magic/sky130A.tech"
MAG_FILE="cells/${CELL}/${CELL}.mag"
GDS_FILE="cells/${CELL}/${CELL}.gds"
DRC_LOG="output/drc_report.txt"

echo "=== FeCIM Magic DRC ==="
echo "Cell:      ${CELL}"
echo "Design:    ${DESIGN}"
echo "Tech file: ${TECH_FILE}"
echo ""

# Check that PDK tech file is available
if [[ ! -f "${TECH_FILE}" ]]; then
    echo "WARNING: PDK tech file not found: ${TECH_FILE}"
    echo "  For SKY130: export PDK_ROOT=~/.volare && volare enable --pdk sky130 sky130A"
    echo "  For IHP:    export IHP_PDK_ROOT=~/ihp-sg13g2-pdk"
    echo "  Falling back to generic DRC (no foundry rules)"
    TECH_FILE=""
fi

mkdir -p output

# Prefer .mag layout; fall back to .gds
if [[ -f "${MAG_FILE}" ]]; then
    INPUT_FILE="${MAG_FILE}"
    INPUT_FORMAT="mag"
elif [[ -f "${GDS_FILE}" ]]; then
    INPUT_FILE="${GDS_FILE}"
    INPUT_FORMAT="gds"
else
    echo "WARNING: No layout file found (${MAG_FILE} or ${GDS_FILE})"
    echo "  The abstract LEF/GDS stub from module6 is sufficient for P&R"
    echo "  Full DRC requires transistor-level Magic layout."
    echo ""
    echo "Educational DRC skip: abstract FeCIM layout has no transistor geometry."
    echo "  For full DRC: create transistor-level layout in Magic, then run:"
    echo "  magic -T ${TECH_FILE:-sky130A} -rcfile magicrc cells/${CELL}/${CELL}.mag"
    exit 0
fi

echo "Running Magic DRC on: ${INPUT_FILE}"

# Run Magic in batch mode with Tcl DRC script
# Unquoted heredoc: bash expands ${INPUT_FILE}; Tcl vars (drc_count) are escaped as \$
magic -T "${TECH_FILE:-sky130A}" -noc -dnull << EOF
drc on
load ${INPUT_FILE}
drc check
set drc_count [drc list count total]
puts "DRC violations: \${drc_count}"
if {\${drc_count} > 0} {
    puts "=== DRC Errors ==="
    drc listall why
}
quit
EOF

echo ""
echo "DRC check complete — see output for violations."
echo "Report: ${DRC_LOG}"
