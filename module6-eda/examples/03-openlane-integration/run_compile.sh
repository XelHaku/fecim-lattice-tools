#!/bin/bash
# Example 03: Compile weights for OpenLane integration
# Run from module6-eda/ directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
DESIGN_NAME="crossbar"

echo "=== FeCIM Example 03: OpenLane Integration ==="
echo ""

mkdir -p "$OUTPUT_DIR"

echo "Step 1: Compiling 16x16 crossbar..."
go run ../cmd/fecim-lattice-tools eda cli \
  -input "$SCRIPT_DIR/weights.json" \
  -output "$OUTPUT_DIR" \
  -rows 16 \
  -cols 16 \
  -levels 30 \
  -vdd 1.8 \
  -name "$DESIGN_NAME" \
  -verilog=true \
  -def=true

echo ""
echo "Step 2: Verifying output files..."

if [ -f "$OUTPUT_DIR/${DESIGN_NAME}.v" ]; then
  echo "  ✓ ${DESIGN_NAME}.v created"
  MODULES=$(grep -c "fecim_bit cell_" "$OUTPUT_DIR/${DESIGN_NAME}.v" || true)
  echo "    Contains $MODULES cell instantiations"
fi

if [ -f "$OUTPUT_DIR/${DESIGN_NAME}.def" ]; then
  echo "  ✓ ${DESIGN_NAME}.def created"
  FIXED=$(grep -c "+ FIXED" "$OUTPUT_DIR/${DESIGN_NAME}.def" || true)
  echo "    Contains $FIXED FIXED placements"
fi

echo ""
echo "=== Compilation Complete ==="
echo ""
echo "Next steps:"
echo "  1. Copy output/ to OpenLane design directory"
echo "  2. Copy cells/ for custom cell definitions"
echo "  3. Run: ./flow.tcl -design fecim_crossbar"
echo ""
echo "See README.md for detailed instructions."
