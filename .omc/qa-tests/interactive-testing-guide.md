# FeCIM Visualizer Interactive QA Testing Guide

**Session**: $(date +%Y-%m-%d_%H-%M-%S)
**App PID**: Running (verified)

## Testing Protocol

Since tmux is not available, this guide provides manual interaction steps for comprehensive testing.

### Module 5 (Comparison) - Tab-Based Layout Testing

#### Tab Structure Verification
1. **Navigate to Module 5 (Comparison)**
   - [ ] Verify 3 tabs exist with icons:
     - ⚡ Energy Comparison
     - 💰 Market & Strategy
     - 🧮 Calculator

#### Tab 1: Energy Comparison
2. **Hero View Testing**
   - [ ] "1000× LESS ENERGY" headline is prominent and bold
   - [ ] Energy race animation is running smoothly
   - [ ] CPU, GPU, FeCIM bars animate correctly
   - [ ] Memory wall visualization shows data movement
   - [ ] Analog states comparison (30 levels) is visible
   - [ ] Scroll down to see all sections

#### Tab 2: Market & Strategy
3. **Business Case Testing**
   - [ ] Market opportunity chart displays "$721B by 2030"
   - [ ] Chart animation shows market growth
   - [ ] Competitive matrix is readable (rows/columns aligned)
   - [ ] Phased strategy diagram shows 3 stages
   - [ ] Verified claims section at bottom shows:
     - VERIFIED: 30 analog levels, 87% MNIST accuracy
     - ENERGY/MAC: CPU (1000 pJ), GPU (100 pJ), FeCIM (~1.0 pJ)
     - CLAIMED: reduction percentages with TRL4 warning

#### Tab 3: Calculator
4. **Interactive Calculator Testing**
   - [ ] Workload selector shows: MNIST, ResNet-50, BERT-Base, GPT-2, LLM-70B
   - [ ] Select "MNIST" - verify calculations update
   - [ ] Inferences slider adjusts from 100 to 100,000
   - [ ] Move slider - verify labels update in real-time
   - [ ] Click "Calculate" button
   - [ ] Calculator shows energy/inference (µJ) for CPU, GPU, FeCIM
   - [ ] Calculator shows power (W)
   - [ ] Calculator shows monthly cost ($)
   - [ ] Data center transformation shows before/after comparison

#### Mode Selector Testing
5. **Presentation Modes**
   - [ ] Mode selector shows: Manual, Auto Demo, Investor, Engineer
   - [ ] Select "Auto Demo" - verify phase timer appears
   - [ ] Phases cycle every ~10 seconds:
     - Energy Race
     - Market Opportunity
     - Strategy
   - [ ] Select "Investor" - verify educational panel changes
   - [ ] Select "Engineer" - verify technical details appear
   - [ ] Select "Manual" - verify status returns to "Ready"

#### Pause/Resume Testing
6. **Animation Control**
   - [ ] Click "Pause" button
   - [ ] Verify all animations stop
   - [ ] Button text changes to "Resume"
   - [ ] Click "Resume"
   - [ ] Animations restart smoothly

---

### Module 3 (MNIST) - Improved Layout Testing

#### Canvas Size Verification
7. **Drawing Canvas**
   - [ ] Canvas should be approximately 350x350 pixels (larger than before)
   - [ ] Canvas is centered in left panel
   - [ ] Draw a digit (e.g., "3")
   - [ ] Canvas responds smoothly to mouse/touch input
   - [ ] Lines are smooth and continuous

#### Inference Testing
8. **Neural Network Inference**
   - [ ] After drawing, inference starts automatically
   - [ ] Status shows phases:
     - Phase 1: Processing input pixels
     - Phase 2: Hidden layer MVM (784→128)
     - Phase 3: Output layer MVM (128→10)
   - [ ] Layer activation view updates during inference
   - [ ] Output bar chart shows probability distribution
   - [ ] Prediction display shows digit (0-9)
   - [ ] Confidence percentage shown

#### Controls Layout
9. **Button Layout (2x2 Grid)**
   - [ ] Verify 4 buttons in 2x2 grid (not cramped):
     - Row 1: "Clear", "Random"
     - Row 2: "Load Data", "Evaluate"
   - [ ] Click "Clear" - canvas clears
   - [ ] Click "Random" - random test digit loads
   - [ ] Click "Load Data" - status shows loading message
   - [ ] Click "Evaluate" - full evaluation runs

#### Weight Visualization Tabs
10. **New Tabs: Quantization & Energy**
    - [ ] Click on layer activation view
    - [ ] Verify tabs exist:
      - Input Layer
      - Hidden Layer
      - Output Layer
      - **Quantization** (NEW)
      - **Energy** (NEW)
    - [ ] Click "Quantization" tab
      - Shows 30-level quantization visualization
    - [ ] Click "Energy" tab
      - Shows energy consumption per layer

#### Slider Controls
11. **Non-Ideality Controls**
    - [ ] Verify sliders are in 2 rows (not cramped):
      - Row 1: Levels (1-30), Noise (0-5%)
      - Row 2: ADC Bits (4-8), DAC Bits (4-8)
    - [ ] Move "Levels" slider
      - Verify label updates (e.g., "Levels: 15")
    - [ ] Move "Noise" slider
      - Verify label updates (e.g., "Noise: 2.5%")

#### Preset Buttons
12. **Preset Layout (2 Rows)**
    - [ ] Verify preset buttons in 2 rows (not 5-column cramped):
      - Row 1: "Ideal", "Real", "Stressed"
      - Row 2: "Extreme", "Reset"
    - [ ] Click "Ideal" - sliders jump to ideal values
    - [ ] Click "Stressed" - noise increases
    - [ ] Click "Reset" - returns to defaults

---

### Module 6 (EDA) - Polished Layout Testing

#### Builder Tab
13. **Statistics Panel**
    - [ ] Verify statistics show:
      - Cells: 256 (16x16)
      - Density: calculated percentage
      - Utilization: calculated percentage
    - [ ] Click "Generate Verilog" button
    - [ ] Verify log shows generation progress
    - [ ] Statistics update after generation

#### Preview Tabs
14. **Preview Before Generation**
    - [ ] Verify preview tabs exist:
      - Verilog Preview
      - DEF Preview
      - Constraints Preview
    - [ ] Click "Verilog Preview"
      - Shows example Verilog code BEFORE generation
      - Code is syntax-highlighted or monospace
    - [ ] Click "DEF Preview"
      - Shows example DEF content
    - [ ] Click "Constraints Preview"
      - Shows example SDC constraints

#### Log Panel
15. **Operation Log**
    - [ ] Verify log has:
      - Monospace font (readable code snippets)
      - "Clear" button at top
      - Scroll capability
    - [ ] Generate Verilog
    - [ ] Log shows timestamped messages
    - [ ] Click "Clear" - log empties

#### OpenLane Panel
16. **Compact Layout**
    - [ ] Verify OpenLane panel is NOT 50% width (should be ~30%)
    - [ ] Panel shows:
      - Platform selector (sky130A, asap7)
      - "Run OpenLane" button
      - "View Results" button
    - [ ] Panel does not dominate the screen

#### Learn Tab
17. **Diagram Rendering**
    - [ ] Click "Learn" tab
    - [ ] Verify diagrams render correctly:
      - Ferroelectric hysteresis loop
      - Crossbar architecture
      - DAC/ADC peripherals
    - [ ] Scroll through content - no rendering issues

---

### All Modules - General Testing

#### No Crashes or Freezes
18. **Stability**
    - [ ] Switch between all 6 modules rapidly
    - [ ] No crashes observed
    - [ ] No UI freezes (animations continue)
    - [ ] No error popups

#### Animation Smoothness
19. **Performance**
    - [ ] Module 1 (Hysteresis): P-E loop animates at ~30 FPS
    - [ ] Module 2 (Crossbar): MVM computation animates smoothly
    - [ ] Module 3 (MNIST): Inference phases transition smoothly
    - [ ] Module 4 (Circuits): DAC/ADC waveforms animate
    - [ ] Module 5 (Comparison): Energy race animates without stuttering

#### Screenshot Feature
20. **Screenshot Button**
    - [ ] Verify screenshot button exists in toolbar
    - [ ] Click screenshot button
    - [ ] Notification appears with file path
    - [ ] Check screenshots/ directory for new PNG file
    - [ ] Open PNG - verify it shows current module view

---

## Automated Checks (Code-Based)

### Layout Constants Verification
Run these checks to verify code matches expected layout:

```bash
# Module 3: Canvas size should be 350x350
grep -n "350" module3-mnist/pkg/gui/digit_canvas.go

# Module 3: Button grid should be 2x2
grep -n "NewGridWithColumns(2" module3-mnist/pkg/gui/app.go

# Module 5: Tab count should be 3
grep -n "NewTabItem" module5-comparison/pkg/gui/app.go | wc -l

# Module 6: OpenLane panel width (should be SetMinSize ~300-400, not 50%)
grep -n "SetMinSize" module6-eda/pkg/gui/app.go
```

### Log File Analysis
Check logs for errors:

```bash
# Latest log file
LATEST_LOG=$(ls -t logs/*.log | head -1)
echo "Analyzing: $LATEST_LOG"

# Check for errors
grep -i "error\|panic\|fatal" "$LATEST_LOG"

# Check for layout warnings
grep -i "layout\|resize\|cascade" "$LATEST_LOG"

# Count inference operations (Module 3)
grep -c "onDigitChanged" "$LATEST_LOG"
```

---

## Issue Reporting Template

If you find an issue, report it in this format:

```markdown
### Issue: [Brief description]

**Module**: [1-6]
**Severity**: [Critical/High/Medium/Low]

**Steps to Reproduce**:
1. Step 1
2. Step 2
3. Step 3

**Expected Behavior**:
[What should happen]

**Actual Behavior**:
[What actually happens]

**Screenshot**: [Path to screenshot if applicable]

**Logs** (if relevant):
```
[paste relevant log lines]
```
```

---

## Success Criteria

- [ ] All checkboxes above are marked
- [ ] No crashes or freezes observed
- [ ] Animations run smoothly (>20 FPS perceived)
- [ ] UI is responsive (buttons/sliders react <100ms)
- [ ] No layout issues (overlapping text, cropped widgets)
- [ ] Screenshot feature works correctly

**Test completed**: ________________ (date/time)
**Tester**: ________________
**Result**: PASS / FAIL
**Issues found**: _____ (count)
