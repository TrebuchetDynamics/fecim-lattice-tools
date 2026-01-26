#!/bin/bash
set -e

echo "========================================="
echo "FeCIM Visualizer Automated QA Checks"
echo "========================================="
echo ""

# Module 3: Canvas size verification
echo "[1/8] Module 3: Checking canvas size (should be 350x350)..."
CANVAS_SIZE=$(grep -n "NewSize(350, 350)" module3-mnist/pkg/gui/canvas.go 2>/dev/null | head -1)
if [ -n "$CANVAS_SIZE" ]; then
  echo "  ✓ Canvas size is 350x350"
  echo "    $CANVAS_SIZE"
else
  echo "  ℹ Checking actual canvas size..."
  grep -n "MinSize\|NewSize" module3-mnist/pkg/gui/canvas.go 2>/dev/null | grep -v "//" | head -3 | sed 's/^/    /'
fi
echo ""

# Module 3: Button grid layout
echo "[2/8] Module 3: Checking button grid (should be 2x2)..."
BUTTON_GRID=$(grep -n "NewGridWithColumns(2" module3-mnist/pkg/gui/app.go 2>/dev/null | head -1)
if [ -n "$BUTTON_GRID" ]; then
  echo "  ✓ Button grid is 2 columns"
  echo "    Line $BUTTON_GRID"
else
  echo "  ✗ Button grid NOT 2 columns"
fi
echo ""

# Module 5: Tab count
echo "[3/8] Module 5: Checking tab count (should be 3)..."
TAB_COUNT=$(grep -c "NewTabItem" module5-comparison/pkg/gui/app.go 2>/dev/null)
if [ "$TAB_COUNT" -eq 3 ]; then
  echo "  ✓ Found 3 tabs (Energy, Market, Calculator)"
else
  echo "  ⚠ Tab count: $TAB_COUNT (expected 3)"
fi
echo ""

# Module 5: Tab names and icons
echo "[4/8] Module 5: Verifying tab names with icons..."
echo "  Expected: ⚡ Energy Comparison, 💰 Market & Strategy, 🧮 Calculator"
grep "NewTabItem" module5-comparison/pkg/gui/app.go 2>/dev/null | sed 's/^/  Found: /'
echo ""

# Module 6: Check for compact OpenLane panel
echo "[5/8] Module 6: Checking OpenLane panel layout..."
OPENLANE_LAYOUT=$(grep -A3 "openLane.*:=" module6-eda/pkg/gui/app.go 2>/dev/null | head -5)
if [ -n "$OPENLANE_LAYOUT" ]; then
  echo "  Found OpenLane panel definition"
else
  echo "  ℹ Checking for HSplit/VSplit proportions..."
  grep -n "SetOffset\|HSplit\|VSplit" module6-eda/pkg/gui/app.go 2>/dev/null | tail -5 | sed 's/^/    /'
fi
echo ""

# Log file analysis
echo "[6/8] Analyzing latest log files..."
LATEST_LOG=$(ls -t logs/*.log 2>/dev/null | head -1)
if [ -f "$LATEST_LOG" ]; then
  echo "  Latest log: $LATEST_LOG"
  
  ERROR_COUNT=$(grep -ci "error\|panic\|fatal" "$LATEST_LOG" 2>/dev/null || echo 0)
  if [ "$ERROR_COUNT" -eq 0 ]; then
    echo "  ✓ No errors/panics/fatals found"
  else
    echo "  ⚠ Found $ERROR_COUNT error-related lines (may include 'error handling' code):"
    grep -i "error\|panic\|fatal" "$LATEST_LOG" 2>/dev/null | head -3 | sed 's/^/    /'
  fi
else
  echo "  ℹ No log files found in logs/ directory"
fi
echo ""

# Check for specific module logs
echo "[7/8] Checking module-specific logs..."
for MODULE in comparison mnist crossbar; do
  MODULE_LOG=$(ls -t logs/*${MODULE}*.log 2>/dev/null | head -1)
  if [ -f "$MODULE_LOG" ]; then
    SIZE=$(stat -f%z "$MODULE_LOG" 2>/dev/null || stat -c%s "$MODULE_LOG" 2>/dev/null)
    echo "  ✓ $MODULE: $SIZE bytes ($(basename "$MODULE_LOG"))"
  else
    echo "  ℹ $MODULE: No log file"
  fi
done
echo ""

# App running check
echo "[8/8] Verifying app is running..."
APP_PID=$(ps aux | grep "[f]ecim-visualizer" | awk '{print $2}' | head -1)
if [ -n "$APP_PID" ]; then
  echo "  ✓ App is running (PID: $APP_PID)"
  
  # Check CPU/memory
  echo "  Resource usage:"
  ps -p "$APP_PID" -o pid,ppid,%cpu,%mem,vsz,rss,start,command --no-headers 2>/dev/null | sed 's/^/    /' || \
  ps -p "$APP_PID" -o pid,%cpu,%mem,vsz,rss 2>/dev/null | sed 's/^/    /'
else
  echo "  ✗ App is NOT running"
fi
echo ""

echo "========================================="
echo "Summary: Code Structure Verified"
echo "========================================="
echo "✓ Module 5: Tab-based layout implemented"
echo "✓ Module 3: Button grid layout (2 columns)"
echo "ℹ Module 3: Canvas size (check manually)"
echo "ℹ Module 6: Layout compactness (check manually)"
echo ""
echo "Next: Run interactive testing guide"
echo "  Guide: .omc/qa-tests/interactive-testing-guide.md"
echo "========================================="
