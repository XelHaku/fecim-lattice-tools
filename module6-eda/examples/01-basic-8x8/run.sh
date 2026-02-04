#!/bin/bash
# Example 01: Basic 8x8 Crossbar Compilation
# Run from module6-eda/ directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"

echo "=== FeCIM Example 01: Basic 8x8 Crossbar ==="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Run compiler
echo "Compiling weights..."
go run ../cmd/fecim-lattice-tools eda cli \
  -input "$SCRIPT_DIR/weights.json" \
  -output "$OUTPUT_DIR" \
  -rows 8 \
  -cols 8 \
  -levels 30 \
  -vdd 1.8 \
  -json=true \
  -csv=true \
  -spice=true \
  -verilog=true \
  -def=true

echo ""
echo "=== Output Files ==="
ls -la "$OUTPUT_DIR"

echo ""
echo "=== Validation ==="

# Check JSON
if [ -f "$OUTPUT_DIR/mapping.json" ]; then
  echo "✓ mapping.json created"
  CELLS=$(grep -o '"row"' "$OUTPUT_DIR/mapping.json" | wc -l)
  echo "  Contains $CELLS cells"
fi

# Check CSV
if [ -f "$OUTPUT_DIR/cells.csv" ]; then
  echo "✓ cells.csv created"
  LINES=$(wc -l < "$OUTPUT_DIR/cells.csv")
  echo "  Contains $((LINES-1)) data rows"
fi

# Check Verilog (if iverilog available)
if [ -f "$OUTPUT_DIR/crossbar.v" ]; then
  echo "✓ crossbar.v created"
  if command -v iverilog &> /dev/null; then
    if iverilog -o /dev/null "$OUTPUT_DIR/crossbar.v" 2>/dev/null; then
      echo "  Verilog syntax valid"
    else
      echo "  WARNING: Verilog syntax errors"
    fi
  fi
fi

# Check DEF
if [ -f "$OUTPUT_DIR/crossbar.def" ]; then
  echo "✓ crossbar.def created"
  COMPONENTS=$(grep -c "fecim_bit" "$OUTPUT_DIR/crossbar.def" || true)
  echo "  Contains $COMPONENTS cell placements"
fi

# Check SPICE
if [ -f "$OUTPUT_DIR/crossbar.sp" ]; then
  echo "✓ crossbar.sp created"
fi

echo ""
echo "=== Done ==="
