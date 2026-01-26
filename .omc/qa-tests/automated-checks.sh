#!/bin/bash
set -e

echo "========================================="
echo "FeCIM Visualizer Automated QA Checks"
echo "========================================="
echo ""

# Module 3: Canvas size verification
echo "[1/8] Module 3: Checking canvas size (should be 350x350)..."
CANVAS_SIZE=$(grep -n "NewSize(350, 350)" module3-mnist/pkg/gui/digit_canvas.go | head -1)
if [ -n "$CANVAS_SIZE" ]; then
  echo "  ✓ Canvas size is 350x350"
  echo "    $CANVAS_SIZE"
else
  echo "  ✗ Canvas size NOT 350x350"
fi
echo ""

# Module 3: Button grid layout
echo "[2/8] Module 3: Checking button grid (should be 2x2)..."
BUTTON_GRID=$(grep -n "NewGridWithColumns(2" module3-mnist/pkg/gui/app.go | head -1)
if [ -n "$BUTTON_GRID" ]; then
  echo "  ✓ Button grid is 2 columns"
  echo "    $BUTTON_GRID"
else
  echo "  ✗ Button grid NOT 2 columns"
fi
echo ""

# Module 5: Tab count
echo "[3/8] Module 5: Checking tab count (should be 3)..."
TAB_COUNT=$(grep -c "NewTabItem" module5-comparison/pkg/gui/app.go)
if [ "$TAB_COUNT" -eq 3 ]; then
  echo "  ✓ Found 3 tabs"
else
  echo "  ✗ Tab count mismatch: $TAB_COUNT (expected 3)"
fi
echo ""

# Module 5: Tab names
echo "[4/8] Module 5: Checking tab names..."
grep "NewTabItem" module5-comparison/pkg/gui/app.go | sed 's/^/  /'
echo ""

# Module 6: OpenLane panel width
echo "[5/8] Module 6: Checking OpenLane panel width..."
PANEL_WIDTH=$(grep -n "SetMinSize.*openLane" module6-eda/pkg/gui/app.go)
if [ -n "$PANEL_WIDTH" ]; then
  echo "  Found: $PANEL_WIDTH"
else
  echo "  ℹ No explicit SetMinSize for OpenLane panel"
fi
echo ""

# Log file analysis
echo "[6/8] Analyzing latest log files..."
LATEST_LOG=$(ls -t logs/*.log | head -1)
echo "  Latest log: $LATEST_LOG"

ERROR_COUNT=$(grep -i "error\|panic\|fatal" "$LATEST_LOG" 2>/dev/null | wc -l)
if [ "$ERROR_COUNT" -eq 0 ]; then
  echo "  ✓ No errors/panics/fatals found"
else
  echo "  ⚠ Found $ERROR_COUNT error lines:"
  grep -i "error\|panic\|fatal" "$LATEST_LOG" | head -5 | sed 's/^/    /'
fi
echo ""

# Layout warnings
echo "[7/8] Checking for layout warnings..."
LAYOUT_WARNINGS=$(grep -i "layout\|resize\|cascade" "$LATEST_LOG" 2>/dev/null | wc -l)
if [ "$LAYOUT_WARNINGS" -eq 0 ]; then
  echo "  ✓ No layout warnings"
else
  echo "  ℹ Found $LAYOUT_WARNINGS layout-related messages"
fi
echo ""

# App running check
echo "[8/8] Verifying app is running..."
APP_PID=$(ps aux | grep fecim-lattice-tools | grep -v grep | awk '{print $2}' | head -1)
if [ -n "$APP_PID" ]; then
  echo "  ✓ App is running (PID: $APP_PID)"
  
  # Check uptime
  APP_START=$(ps -p "$APP_PID" -o lstart= 2>/dev/null)
  echo "  Started: $APP_START"
  
  # Check CPU/memory
  APP_STATS=$(ps -p "$APP_PID" -o %cpu,%mem,vsz,rss --no-headers)
  echo "  Stats: CPU $APP_STATS"
else
  echo "  ✗ App is NOT running"
fi
echo ""

echo "========================================="
echo "Automated checks complete"
echo "========================================="
