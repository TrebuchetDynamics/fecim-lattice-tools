#!/bin/bash
# Example 02: MNIST First Layer Compilation
# Run from module6-eda/ directory

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OUTPUT_DIR="$SCRIPT_DIR/output"

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
  -json=true \
  -csv=true \
  -spice=true \
  -verilog=true \
  -def=true

echo ""
echo "=== Statistics ==="
if [ -f "$OUTPUT_DIR/mapping.json" ]; then
  echo "Extracting compilation stats..."
  python3 -c "
import json
with open('$OUTPUT_DIR/mapping.json') as f:
    data = json.load(f)
    stats = data.get('stats', {})
    print(f\"  Total Cells: {stats.get('total_cells', 'N/A')}\")
    print(f\"  Utilization: {stats.get('utilization', 0)*100:.1f}%\")
    print(f\"  PSNR: {stats.get('psnr_db', 'N/A'):.1f} dB\")
" 2>/dev/null || echo "  (Install python3 for detailed stats)"
fi

echo ""
echo "=== Output Files ==="
ls -la "$OUTPUT_DIR"

echo ""
echo "=== Done ==="
echo ""
echo "Next: Run ngspice simulation with testbench.sp"
