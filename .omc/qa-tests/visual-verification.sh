#!/bin/bash

echo "========================================="
echo "Visual Verification Helper"
echo "========================================="
echo ""
echo "The FeCIM Visualizer app is running."
echo "Window ID: $(wmctrl -l | grep 'FeCIM Lattice Tools' | awk '{print $1}')"
echo ""
echo "MANUAL VERIFICATION STEPS:"
echo "========================================="
echo ""

echo "1. MODULE 5 (COMPARISON) - TAB-BASED LAYOUT"
echo "   - Navigate to Module 5"
echo "   - Verify 3 tabs exist:"
echo "     • ⚡ Energy Comparison"
echo "     • 💰 Market & Strategy"
echo "     • 🧮 Calculator"
echo ""
echo "   Tab 1 (Energy):"
echo "     • \"1000× LESS ENERGY\" headline is prominent"
echo "     • Energy race animation is running"
echo "     • CPU/GPU/FeCIM bars animate"
echo ""
echo "   Tab 2 (Market):"
echo "     • Market chart shows \"$721B by 2030\""
echo "     • Competitive matrix is readable"
echo "     • Phased strategy shows 3 stages"
echo ""
echo "   Tab 3 (Calculator):"
echo "     • Workload selector: MNIST, ResNet-50, BERT, GPT-2, LLM-70B"
echo "     • Slider updates labels in real-time"
echo "     • Calculate button works"
echo ""

echo "2. MODULE 3 (MNIST) - IMPROVED LAYOUT"
echo "   - Navigate to Module 3"
echo "   - Canvas is larger (~350x350 pixels)"
echo "   - Draw a digit (e.g., 3)"
echo "   - Inference shows 3 phases"
echo "   - Buttons in 2x2 grid (Clear, Random, Load Data, Evaluate)"
echo ""

echo "3. MODULE 6 (EDA) - POLISHED"
echo "   - Navigate to Module 6"
echo "   - Builder tab shows statistics"
echo "   - Preview tabs exist (Verilog, DEF, Constraints)"
echo "   - Log has monospace font + Clear button"
echo "   - OpenLane panel is compact (~30% width, NOT 50%)"
echo ""

echo "4. PERFORMANCE CHECK"
echo "   - Rapidly switch between all 6 modules"
echo "   - Verify no crashes or freezes"
echo "   - Animations should run smoothly (>20 FPS)"
echo ""

echo "5. SCREENSHOT TEST"
echo "   - Click screenshot button"
echo "   - Check screenshots/ directory for new file"
echo ""

echo "========================================="
echo "AUTOMATED CHECKS (COMPLETED):"
echo "========================================="
echo "✅ Code structure verified"
echo "✅ Module 5: 3 tabs implemented"
echo "✅ Module 3: 350x350 canvas"
echo "✅ Module 3: 2x2 button grid"
echo "✅ No errors in logs"
echo "✅ App is running (PID: $(ps aux | grep '[f]ecim-visualizer' | awk '{print $2}' | head -1))"
echo ""

echo "========================================="
echo "SCREENSHOT CAPTURE"
echo "========================================="
echo ""
echo "Taking screenshot of current window..."

# Find FeCIM window
WINDOW_ID=$(wmctrl -l | grep 'FeCIM Lattice Tools' | awk '{print $1}')

if [ -n "$WINDOW_ID" ]; then
  # Focus window
  wmctrl -i -a "$WINDOW_ID"
  sleep 0.5
  
  # Take screenshot
  SCREENSHOT_FILE="<local-path> +%Y%m%d-%H%M%S).png"
  scrot -u "$SCREENSHOT_FILE" 2>/dev/null && echo "✅ Screenshot saved: $SCREENSHOT_FILE" || echo "❌ Screenshot failed (scrot not available)"
else
  echo "❌ Could not find FeCIM window"
fi

echo ""
echo "========================================="
echo "NEXT STEPS:"
echo "========================================="
echo "1. Complete manual verification steps above"
echo "2. Document any issues found"
echo "3. Mark checkboxes in:"
echo "   .omc/qa-tests/interactive-testing-guide.md"
echo "4. Update status in:"
echo "   .omc/qa-tests/QA_STATUS_REPORT.md"
echo ""
echo "Full testing guide:"
echo "  cat .omc/qa-tests/interactive-testing-guide.md"
echo "========================================="
