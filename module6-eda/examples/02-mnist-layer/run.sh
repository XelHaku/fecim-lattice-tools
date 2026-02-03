#!/bin/bash
# Example 02: MNIST First Layer Compilation
# Run from module6-eda/ directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"
DESIGN_NAME="fecim_array"

echo "=== FeCIM Example 02: MNIST First Layer ==="
echo ""

mkdir -p "$OUTPUT_DIR"

echo "Compiling 32x32 weight matrix..."
go run ./cmd/eda-cli \
  -input "$SCRIPT_DIR/weights.json" \
  -output "$OUTPUT_DIR" \
  -rows 32 \
  -cols 32 \
  -levels 30 \
  -vdd 1.8 \
  -name "$DESIGN_NAME" \
  -json=true \
  -csv=true \
  -spice=true \
  -verilog=true \
  -def=true

echo ""
echo "=== Statistics ==="
DESIGN_JSON="$OUTPUT_DIR/${DESIGN_NAME}_design.json"
if [ -f "$DESIGN_JSON" ]; then
  echo "Extracting compilation stats..."
  python3 -c "
import json
with open('$DESIGN_JSON') as f:
    data = json.load(f)
    stats = data.get('stats', {})
    print(f\"  Total Cells: {stats.get('total_cells', 'N/A')}\")
    print(f\"  Active Cells: {stats.get('active_cells', 'N/A')}\")
    print(f\"  Quant PSNR: {stats.get('quant_psnr_db', 'N/A')} dB\")
" 2>/dev/null || echo "  (Install python3 for detailed stats)"
fi

echo ""
echo "=== Output Files ==="
ls -la "$OUTPUT_DIR"

echo ""
echo "=== Done ==="
echo ""
