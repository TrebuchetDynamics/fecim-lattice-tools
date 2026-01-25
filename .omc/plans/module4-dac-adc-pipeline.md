# Module 4 COMPUTE Mode DAC/ADC Pipeline Fix

## Context

### Original Request
Fix Module 4 Circuits COMPUTE mode to have proper DAC/ADC pipeline visualization, matching the data path paradigm of WRITE and READ modes.

### Current Issues
1. **Misleading Output Label**: Line 1100 states "Row currents (uA) -> ADC -> digital" but displays RAW uA values (e.g., "y0: 160.0 uA"), NOT digitized output
2. **No Random Bits Generator**: Only a dropdown with "Manual/Random/Ramp" exists (line 1052-1069), but "Random" sets random 0-255 values - there's no dedicated "RANDOM BITS" button like other modes have
3. **Missing Input Data Path**: No visualization showing DIGITAL(bits) -> DAC -> VOLTAGE conversion chain
4. **Missing Output Data Path**: No visualization showing CURRENT -> TIA -> ADC -> DIGITAL conversion chain
5. **No Peripheral Model Usage**: `onOpsCompute()` (lines 1361-1394) does raw math without using `ca.dac`, `ca.adc`, or `ca.tia`

### Reference Patterns (Working)
- **WRITE mode** (lines 594-606): `DIGITAL -> DAC -> FeFET` with dynamic labels
- **READ mode** (lines 841-851): `FeFET -> TIA -> ADC -> DIGITAL` with static labels

---

## Work Objectives

### Core Objective
Transform COMPUTE mode from displaying raw current values to showing the **FULL 8×8 DAC/ADC PIPELINE** with proper physical access pattern visualization.

### Physical Architecture Understanding (NAND/DRAM/GPU-like CIM)

**Industry-Standard CIM Architecture:**
```
        COLUMNS (Input/Program Lines)
        ↓     ↓     ↓     ↓     ↓     ↓     ↓     ↓
      DAC0  DAC1  DAC2  DAC3  DAC4  DAC5  DAC6  DAC7   ← 8 Column DACs
       x0    x1    x2    x3    x4    x5    x6    x7
        │     │     │     │     │     │     │     │
  ┌─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┴─────┐
  │    [●]   [●]   [●]   [●]   [●]   [●]   [●]   [●]    │──┬─→ TIA0→ADC0 → y0
  │    [●]   [●]   [●]   [●]   [●]   [●]   [●]   [●]    │──┼─→ TIA1→ADC1 → y1
  │    [●]   [●]   [●]   [●]   [●]   [●]   [●]   [●]    │──┼─→ TIA2→ADC2 → y2
  │                    8×8 FeFET ARRAY                  │  │      ...
  │    [●]   [●]   [●]   [●]   [●]   [●]   [●]   [●]    │──┴─→ TIA7→ADC7 → y7
  └─────────────────────────────────────────────────────┘
                                                         ↑
                                              ROWS (Sense Lines) + Switchable Drivers
```

**Switchable Row Driver (per row):**
```
                    ┌─────────────────┐
                    │   MODE SELECT   │
    WRITE mode ────→│ ○ GND/Vref     │──→ Row grounds during programming
    READ/COMPUTE ──→│ ○ TIA → ADC    │──→ Row senses current
                    └─────────────────┘
```

**Operation Modes (Like NAND/DRAM/GPU):**
| Mode | 8 Column DACs | 8 Row Drivers | Like... |
|------|---------------|---------------|---------|
| **WRITE** | Program voltage (2-5V) | GND or Vref | NAND program |
| **READ** | One-hot (single DAC) | TIA→ADC (MUX scan) | NAND/DRAM read |
| **COMPUTE** | Input vector (0-1V) | All 8 TIA→ADC parallel | **GPU MVM** |

**Key Physics:**
- **Cannot access single cell** - voltage goes to ENTIRE column
- **Cannot read single cell** - ADC reads ENTIRE row SUM (KCL)
- **WRITE**: Column DAC programs, row at GND (like NAND)
- **COMPUTE**: All 8 DACs active → 8×8 MVM → All 8 ADCs (like GPU)
- **READ**: One-hot column scan, row TIA senses (like DRAM)

### Deliverables
1. "RANDOM BITS" button that generates random 8-bit digital values for all 8 inputs
2. **FULL INPUT PIPELINE**: `8 DIGITAL → 8 DACs → 8 COLUMN VOLTAGES`
3. **FULL OUTPUT PIPELINE**: `8 ROW SUMS → 8 TIAs → 8 ADCs → 8 DIGITAL LEVELS`
4. **Column/row highlighting** in array canvas to show active paths
5. Dual output display: raw uA (row sum) AND ADC-digitized level
6. **IDEAL CROSSBAR DISCLAIMER**: Note that IR drop and sneak paths are not modeled (see Module 2)
7. Use existing peripheral models: `ca.dac`, `ca.adc`, `ca.tia`

### Definition of Done
- [ ] Random Bits button generates random 0-255 values for all 8 inputs
- [ ] Pipeline header shows "8 DACs → 8×8 ARRAY → 8 TIAs → 8 ADCs"
- [ ] Array canvas highlights active columns (input) and rows (output) during compute
- [ ] Output shows row SUM with digitized level (e.g., "y0: 50.0 uA | L16")
- [ ] Output shows saturation indicator "(SAT)" when current > 100 uA
- [ ] Disclaimer displayed: "IDEAL CROSSBAR - no IR drop or sneak paths (see Module 2)"
- [ ] All peripheral conversions use the proper model methods

---

## Guardrails

### MUST Have
- Use `fyne.Do()` for all UI updates from goroutines
- Preserve existing "Manual/Random/Ramp" dropdown functionality
- Match visual style of WRITE/READ data paths (use `createLabeledBoxWithLabel`)
- Thread-safe access to shared state via `ca.mu` locks

### MUST NOT Have
- Break existing COMPUTE functionality
- Remove the mode selector dropdown
- Hardcode peripheral parameters (use `ca.dac`, `ca.adc`, `ca.tia` methods)
- Introduce UI freezes from blocking operations

---

## Task Flow

```
[1] Add struct fields for new UI widgets
          |
          v
[2] Add "RANDOM BITS" button
          |
          v
[3] Create FULL INPUT PIPELINE visualization (8 DACs)
          |
          v
[4] Create FULL OUTPUT PIPELINE visualization (8 TIAs/ADCs) + IDEAL DISCLAIMER
          |
          v
[5] Enhance array canvas with column/row labels (x0-x7, y0-y7)
          |
          v
[6] Update onOpsCompute() to use peripheral models
          |
          v
[7] Add helper functions for data path updates
          |
          v
[8] Update createComputeModePanel() assembly
          |
          v
[9-11] Hook input changes, mode selector, reset handlers
          |
          v
[12] Test and verify visual consistency
```

---

## Detailed TODOs

### TODO 1: Add struct fields for new UI widgets

**File:** `<local-path>`
**Location:** Lines 165-166 (after `opsComputeMathLabel`)

**Add these fields to `CircuitsApp` struct:**
```go
// Compute mode INPUT data path labels
opsComputeInputDigitalLabel  *widget.Label  // Shows "x0: 128\n0b10000000"
opsComputeInputDACLabel      *widget.Label  // Shows "0.50V"
opsComputeInputVoltageLabel  *widget.Label  // Shows "VOLTAGE"

// Compute mode OUTPUT data path labels
opsComputeOutputCurrentLabel *widget.Label  // Shows "50.0 uA" or "160.0 uA (SAT)"
opsComputeOutputTIALabel     *widget.Label  // Shows "0.500 V" or "1.000 V (SAT)"
opsComputeOutputADCLabel     *widget.Label  // Shows "Level 16" or "Level 31 (SAT)"
opsComputeOutputDigitalLabel *widget.Label  // Shows final digital
```

**Acceptance Criteria:**
- Fields compile without error
- No naming conflicts with existing fields

---

### TODO 2: Add "RANDOM BITS" button to COMPUTE mode

**File:** `<local-path>`
**Location:** Line 1091 (in `createComputeModePanel()`, input section)

**Current code (line 1089-1094):**
```go
inputSection := container.NewVBox(
    widget.NewLabelWithStyle("INPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    container.NewHBox(widget.NewLabel("Mode:"), modeSelect),
    widget.NewLabel("Digital inputs (0-255) -> DAC voltages (0-1V):"),
    inputGrid,
)
```

**Change to:**
```go
// Create Random Bits button
randomBitsBtn := widget.NewButton("RANDOM BITS", func() {
    ca.mu.Lock()
    for i := range ca.inputVector {
        ca.inputVector[i] = rand.Intn(256)
    }
    ca.mu.Unlock()
    ca.updateOpsComputeInputs()
    ca.updateOpsComputeInputDataPath()
})

inputSection := container.NewVBox(
    widget.NewLabelWithStyle("INPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    container.NewHBox(widget.NewLabel("Mode:"), modeSelect, randomBitsBtn),
    widget.NewLabel("Digital inputs (0-255) -> DAC voltages (0-1V):"),
    inputGrid,
)
```

**Acceptance Criteria:**
- Button appears next to mode dropdown
- Clicking generates random 0-255 values for all 8 inputs
- All input Entry widgets update
- All voltage labels update

---

### TODO 3: Create FULL INPUT PIPELINE visualization

**File:** `<local-path>`
**Location:** After line 1094 (after inputGrid, before outputSection)

**Add FULL 8-input pipeline section:**
```go
// FULL INPUT PIPELINE: 8 digital values -> 8 DACs -> 8 column voltages
// This shows the COMPLETE input path, not just one example

inputPipelineHeader := widget.NewLabelWithStyle(
    "INPUT PIPELINE: 8 DIGITAL → 8 DACs → 8 COLUMNS",
    fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
)

// Summary boxes showing the full pipeline
digitalSummaryBox := ca.createLabeledBox("8× DIGITAL", "x0-x7\n(0-255)", sharedtheme.ColorPrimary)
dacSummaryBox := ca.createLabeledBox("8× DAC", "→ 0-1V\neach", sharedtheme.ColorAccent)
columnSummaryBox := ca.createLabeledBox("8 COLUMNS", "Voltages\napplied", sharedtheme.ColorSuccess)

inputPipelinePath := container.NewHBox(
    digitalSummaryBox, widget.NewLabel("→"),
    dacSummaryBox, widget.NewLabel("→"),
    columnSummaryBox,
)

// Physics note about READ-safe voltages
inputPhysicsNote := widget.NewLabel(
    "⚡ COMPUTE uses 0-1V (READ-safe) - won't disturb programmed cell states",
)
inputPhysicsNote.TextStyle = fyne.TextStyle{Italic: true}

inputDataPathSection := container.NewVBox(
    widget.NewSeparator(),
    inputPipelineHeader,
    inputPipelinePath,
    inputPhysicsNote,
)
```

**Acceptance Criteria:**
- Shows FULL 8-input pipeline (not just example x0)
- Header clearly states "8 DIGITAL → 8 DACs → 8 COLUMNS"
- Physics note explains READ-safe voltage range

---

### TODO 4: Create FULL OUTPUT PIPELINE visualization + IDEAL CROSSBAR DISCLAIMER

**File:** `<local-path>`
**Location:** Line 1097-1102 (replace existing outputSection)

**Current code:**
```go
outputSection := container.NewVBox(
    widget.NewSeparator(),
    widget.NewLabelWithStyle("OUTPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    widget.NewLabel("Row currents (uA) -> ADC -> digital:"),  // MISLEADING!
    outputGrid,
)
```

**Replace with:**
```go
// FULL OUTPUT PIPELINE: 8 row sums -> 8 TIAs -> 8 ADCs -> 8 digital levels
outputPipelineHeader := widget.NewLabelWithStyle(
    "OUTPUT PIPELINE: 8 ROWS → 8 TIAs → 8 ADCs → 8 LEVELS",
    fyne.TextAlignLeading, fyne.TextStyle{Bold: true},
)

// Summary boxes for full output pipeline
rowSumBox := ca.createLabeledBox("8× ROW SUM", "y0-y7\n(KCL)", sharedtheme.ColorWarning)
tiaSummaryBox := ca.createLabeledBox("8× TIA", "I→V\n10kΩ", sharedtheme.ColorInfo)
adcSummaryBox := ca.createLabeledBox("8× ADC", "5-bit\n0-31", sharedtheme.ColorSuccess)
levelSummaryBox := ca.createLabeledBox("8× LEVEL", "Digital\noutput", sharedtheme.ColorPrimary)

outputPipelinePath := container.NewHBox(
    rowSumBox, widget.NewLabel("→"),
    tiaSummaryBox, widget.NewLabel("→"),
    adcSummaryBox, widget.NewLabel("→"),
    levelSummaryBox,
)

// Physics note about row sums
outputPhysicsNote := widget.NewLabel(
    "⚡ Each y_i = Σ(G[i,j] × V_j) - sum of 8 cell currents via KCL",
)
outputPhysicsNote.TextStyle = fyne.TextStyle{Italic: true}

// IDEAL CROSSBAR DISCLAIMER
idealDisclaimer := widget.NewLabel(
    "⚠️ IDEAL CROSSBAR: No IR drop or sneak paths modeled (see Module 2 for non-idealities)",
)
idealDisclaimer.TextStyle = fyne.TextStyle{Bold: true}

outputDataPathSection := container.NewVBox(
    widget.NewSeparator(),
    outputPipelineHeader,
    outputPipelinePath,
    outputPhysicsNote,
    idealDisclaimer,
)

outputSection := container.NewVBox(
    widget.NewSeparator(),
    widget.NewLabelWithStyle("OUTPUT VECTOR (Row Sums)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    widget.NewLabel("Each output = sum of 8 cell currents, digitized by TIA+ADC:"),
    outputGrid,
    outputDataPathSection,
)
```

**Acceptance Criteria:**
- Shows FULL 8-output pipeline (8 ROWS → 8 TIAs → 8 ADCs → 8 LEVELS)
- Header clearly shows parallel nature of crossbar readout
- KCL physics note explains row sum calculation
- **IDEAL CROSSBAR DISCLAIMER** visible to user
- Shows "(SAT)" indicator when TIA saturates (current > 100 uA)

---

### TODO 5: Enhance array canvas with column/row highlighting

**File:** `<local-path>`
**Location:** Lines 280-410 (in `drawSharedArray()` function, ModeCompute section)

**Current COMPUTE mode visualization (lines 280-303):**
The code already draws column and row lines, but they're static. We need to:
1. Add column headers showing DAC labels (x0-x7)
2. Add row labels showing TIA/ADC outputs (y0-y7)
3. Highlight active paths with brighter colors

**Enhance existing ModeCompute drawing:**
```go
case ModeCompute:
    // Draw column voltage indicators at TOP (8 DACs)
    dacLabelColor := color.RGBA{150, 200, 255, 255}  // Light blue for input
    for c := 0; c < min(8, cols); c++ {
        x := offsetX + c*cellSize + cellSize/2
        // Draw column line
        for y := offsetY - 25; y < offsetY+gridH+5; y++ {
            if y >= 0 && y < h {
                img.Set(x, y, color.RGBA{100, 100, 200, 180})
            }
        }
        // Draw "x0", "x1", etc. labels at top
        drawSimpleText(img, fmt.Sprintf("x%d", c), x-6, offsetY-28, dacLabelColor)
    }

    // Draw row output indicators at RIGHT (8 TIAs/ADCs)
    adcLabelColor := color.RGBA{255, 200, 150, 255}  // Light orange for output
    for r := 0; r < min(8, rows); r++ {
        y := offsetY + r*cellSize + cellSize/2
        xStart := offsetX + gridW + 5
        // Draw row line
        for x := offsetX - 5; x < xStart + 25; x++ {
            if x >= 0 && x < w {
                img.Set(x, y, color.RGBA{200, 100, 100, 180})
            }
        }
        // Draw "y0", "y1", etc. labels at right
        drawSimpleText(img, fmt.Sprintf("y%d", r), xStart+8, y-3, adcLabelColor)
    }

    // Draw "DACs" label at top
    drawSimpleText(img, "8 DACs", offsetX + gridW/2 - 20, offsetY - 40, dacLabelColor)

    // Draw "ADCs" label at right
    drawSimpleText(img, "8 ADCs", offsetX + gridW + 35, offsetY + gridH/2 - 5, adcLabelColor)
```

**Acceptance Criteria:**
- Column headers show x0-x7 (DAC inputs)
- Row labels show y0-y7 (TIA/ADC outputs)
- "8 DACs" label at top, "8 ADCs" label at right
- Visual shows full parallel MVM operation

---

### TODO 6: Update onOpsCompute() to use peripheral models

**File:** `<local-path>`
**Location:** Lines 1361-1394 (entire `onOpsCompute()` function)

**Current implementation (simplified):**
```go
func (ca *CircuitsApp) onOpsCompute() {
    // ...MVM calculation...
    sum += conductance * voltage
    // ...
    ca.opsComputeOutputLabels[i].SetText(fmt.Sprintf("y%d: %.1f uA", i, val))
}
```

**Replace with:**
```go
func (ca *CircuitsApp) onOpsCompute() {
    ca.mu.Lock()
    rows := min(8, ca.arrayRows)
    cols := min(8, ca.arrayCols)

    // MVM: output = weights * input
    for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
        sum := 0.0
        for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
            // Convert digital input through DAC to voltage
            digitalInput := ca.inputVector[c]
            voltage := float64(digitalInput) / 255.0  // Normalized to 0-1V range

            // Cell conductance based on programmed level
            conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0  // uS

            // Current = G * V (in uA since G is in uS and V is in V)
            sum += conductance * voltage
        }
        ca.outputVector[r] = sum
    }
    ca.mu.Unlock()

    // Update output labels with BOTH raw current AND digitized level
    // PHYSICS NOTE: TIA saturates at MaxInputCurrent (100 uA) -> clamps output to 1.0V
    ca.mu.RLock()
    for i := 0; i < 8 && i < len(ca.outputVector); i++ {
        if ca.opsComputeOutputLabels[i] != nil {
            rawCurrent := ca.outputVector[i]  // uA

            // TIA conversion: current (uA) -> voltage (V)
            // TIA gain is 10 kOhm, MaxInputCurrent = 100 uA, MaxOutputVoltage = 1.0V
            // If rawCurrent > 100 uA, TIA clamps to 1.0V (SATURATION)
            tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)  // Convert uA to A for TIA

            // ADC conversion: voltage -> digital level
            // 5-bit ADC (0-1V range): 0.5V -> level 16, 1.0V -> level 31
            adcLevel := ca.adc.Convert(tiaVoltage)

            // Check for TIA saturation (current > 100 uA causes clamp to 1V)
            isSaturated := rawCurrent > 100.0  // 100 uA is TIA max input

            idx := i
            current := rawCurrent
            level := adcLevel
            sat := isSaturated
            fyne.Do(func() {
                if sat {
                    // Show saturation indicator - ADC level will be 31 (max)
                    ca.opsComputeOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1f uA | L%d (SAT)", idx, current, level))
                } else {
                    ca.opsComputeOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1f uA | L%d", idx, current, level))
                }
            })
        }
    }
    ca.mu.RUnlock()

    // Update output data path visualization (example: y0)
    ca.updateOpsComputeOutputDataPath()

    // Update math breakdown
    ca.updateOpsComputeMath()

    ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
}
```

**Acceptance Criteria:**
- Uses `ca.tia.Convert()` for current-to-voltage
- Uses `ca.adc.Convert()` for voltage-to-level
- Output shows both raw uA AND digitized level
- Thread-safe with proper fyne.Do() usage

---

### TODO 7: Add helper functions for data path updates

**File:** `<local-path>`
**Location:** After line 1146 (after `updateOpsComputeInputs()`)

**Add new functions:**
```go
// updateOpsComputeInputDataPath updates the input data path display (shows x0 as example)
func (ca *CircuitsApp) updateOpsComputeInputDataPath() {
    ca.mu.RLock()
    defer ca.mu.RUnlock()

    if len(ca.inputVector) == 0 {
        return
    }

    // Show x0 as the example
    digitalVal := ca.inputVector[0]
    voltage := float64(digitalVal) / 255.0

    if ca.opsComputeInputDigitalLabel != nil {
        fyne.Do(func() {
            ca.opsComputeInputDigitalLabel.SetText(fmt.Sprintf("x0: %d\n0b%08b", digitalVal, digitalVal))
        })
    }
    if ca.opsComputeInputDACLabel != nil {
        fyne.Do(func() {
            ca.opsComputeInputDACLabel.SetText(fmt.Sprintf("%.2fV", voltage))
        })
    }
}

// updateOpsComputeOutputDataPath updates the output data path display (shows y0 as example)
func (ca *CircuitsApp) updateOpsComputeOutputDataPath() {
    ca.mu.RLock()
    defer ca.mu.RUnlock()

    if len(ca.outputVector) == 0 {
        return
    }

    // Show y0 as the example
    rawCurrent := ca.outputVector[0]  // uA

    // TIA conversion (saturates at 100 uA -> 1.0V output)
    // PHYSICS: Gain = 10 kOhm, MaxInputCurrent = 100 uA, MaxOutputVoltage = 1.0V
    tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)  // uA to A

    // ADC conversion (5-bit: 0V->0, 1V->31)
    adcLevel := ca.adc.Convert(tiaVoltage)

    // Check for TIA saturation
    isSaturated := rawCurrent > 100.0  // uA

    satSuffix := ""
    if isSaturated {
        satSuffix = " (SAT)"
    }

    if ca.opsComputeOutputCurrentLabel != nil {
        fyne.Do(func() {
            ca.opsComputeOutputCurrentLabel.SetText(fmt.Sprintf("%.1f uA%s", rawCurrent, satSuffix))
        })
    }
    if ca.opsComputeOutputTIALabel != nil {
        fyne.Do(func() {
            ca.opsComputeOutputTIALabel.SetText(fmt.Sprintf("%.3f V%s", tiaVoltage, satSuffix))
        })
    }
    if ca.opsComputeOutputADCLabel != nil {
        fyne.Do(func() {
            ca.opsComputeOutputADCLabel.SetText(fmt.Sprintf("Level %d%s", adcLevel, satSuffix))
        })
    }
}
```

**Acceptance Criteria:**
- `updateOpsComputeInputDataPath()` shows x0's digital value and DAC voltage
- `updateOpsComputeOutputDataPath()` shows y0's current, TIA voltage, and ADC level
- Both use proper `fyne.Do()` for UI updates
- Both show "(SAT)" indicator when TIA saturates

---

### TODO 8: Update createComputeModePanel() assembly

**File:** `<local-path>`
**Location:** Lines 1120-1126 (end of `createComputeModePanel()`)

**Variable Scoping Note:** The `inputDataPathSection` variable created in TODO 3 must be available in this TODO's scope. Both are within the same function `createComputeModePanel()`, so the variable is naturally in scope.

**Current code:**
```go
ca.computeConfigPanel = container.NewVBox(
    inputSection,
    outputSection,
    mathSection,
    perfSection,
)
```

**Update to include data path sections:**
```go
// NOTE: inputDataPathSection is created earlier in this function (TODO 3)
ca.computeConfigPanel = container.NewVBox(
    inputSection,
    inputDataPathSection,  // NEW: Input data path (from TODO 3)
    outputSection,         // Already includes outputDataPathSection
    mathSection,
    perfSection,
)

// Initialize data path values
ca.updateOpsComputeInputDataPath()
```

**Acceptance Criteria:**
- Input data path appears between input grid and output section
- Output data path appears within output section (after output grid)
- Initial values populated on panel creation

---

### TODO 9: Hook input changes to data path update

**File:** `<local-path>`
**Location:** Lines 1028-1040 (input entry OnChanged handler)

**Current code:**
```go
ca.opsComputeInputs[i].OnChanged = func(s string) {
    var v int
    fmt.Sscanf(s, "%d", &v)
    if v > 255 {
        v = 255
    }
    ca.mu.Lock()
    ca.inputVector[idx] = v
    ca.mu.Unlock()
    if ca.opsComputeVoltageLabels[idx] != nil {
        ca.opsComputeVoltageLabels[idx].SetText(fmt.Sprintf("%.2fV", float64(v)/255.0))
    }
}
```

**Add call to update input data path:**
```go
ca.opsComputeInputs[i].OnChanged = func(s string) {
    var v int
    fmt.Sscanf(s, "%d", &v)
    if v > 255 {
        v = 255
    }
    ca.mu.Lock()
    ca.inputVector[idx] = v
    ca.mu.Unlock()
    if ca.opsComputeVoltageLabels[idx] != nil {
        ca.opsComputeVoltageLabels[idx].SetText(fmt.Sprintf("%.2fV", float64(v)/255.0))
    }
    // Update input data path (shows x0 example)
    if idx == 0 {
        ca.updateOpsComputeInputDataPath()
    }
}
```

**Acceptance Criteria:**
- Changing x0 input updates the input data path visualization
- Other inputs don't trigger unnecessary updates

---

### TODO 10: Update mode selector Random/Ramp handlers

**File:** `<local-path>`
**Location:** Lines 1052-1069 (modeSelect OnChanged handler)

**Current code:**
```go
modeSelect := widget.NewSelect([]string{"Manual", "Random", "Ramp"}, func(s string) {
    switch s {
    case "Random":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = rand.Intn(256)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
    case "Ramp":
        // ...
        ca.updateOpsComputeInputs()
    }
})
```

**Add data path update calls:**
```go
modeSelect := widget.NewSelect([]string{"Manual", "Random", "Ramp"}, func(s string) {
    switch s {
    case "Random":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = rand.Intn(256)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
        ca.updateOpsComputeInputDataPath()  // ADD THIS
    case "Ramp":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
        ca.updateOpsComputeInputDataPath()  // ADD THIS
    }
})
```

**Acceptance Criteria:**
- Random mode updates input data path
- Ramp mode updates input data path

---

### TODO 11: Update onOpsReset() to clear data paths

**File:** `<local-path>`
**Location:** Lines 1454-1485 (`onOpsReset()` function)

**Add at end of function (before status update):**
```go
// Reset data path displays
if ca.opsComputeInputDigitalLabel != nil {
    fyne.Do(func() {
        ca.opsComputeInputDigitalLabel.SetText("x0: 0\n0b00000000")
    })
}
if ca.opsComputeInputDACLabel != nil {
    fyne.Do(func() {
        ca.opsComputeInputDACLabel.SetText("0.00V")
    })
}
if ca.opsComputeOutputCurrentLabel != nil {
    fyne.Do(func() {
        ca.opsComputeOutputCurrentLabel.SetText("-- uA")
    })
}
if ca.opsComputeOutputTIALabel != nil {
    fyne.Do(func() {
        ca.opsComputeOutputTIALabel.SetText("-- V")
    })
}
if ca.opsComputeOutputADCLabel != nil {
    fyne.Do(func() {
        ca.opsComputeOutputADCLabel.SetText("Level --")
    })
}

ca.operationsStatusLabel.SetText("Reset complete")
```

**Acceptance Criteria:**
- Reset clears all data path labels
- Shows placeholder values after reset

---

## Commit Strategy

### Single Commit (Recommended)
All changes are tightly coupled and should be committed together:

```
feat(module4): Add proper DAC/ADC pipeline to COMPUTE mode

- Add RANDOM BITS button for generating random input vectors
- Add INPUT data path visualization (DIGITAL -> DAC -> VOLTAGE)
- Add OUTPUT data path visualization (CURRENT -> TIA -> ADC -> DIGITAL)
- Update output display to show both raw uA and ADC-digitized levels
- Fix misleading "Row currents (uA) -> ADC -> digital" label
- Use existing peripheral models (ca.tia, ca.adc) for conversions

Fixes compute mode UI to match WRITE/READ data path paradigms.
```

---

## Success Criteria

1. **Visual Parity**: COMPUTE mode data paths match WRITE/READ mode visual style
2. **Accurate Labels**: No misleading claims about conversions that don't happen
3. **Functional Pipeline**: Output shows actual ADC-converted levels, not just raw currents
4. **Complete UI**: Random Bits button works, all data path boxes update appropriately
5. **No Regressions**: Existing Manual/Random/Ramp modes still work
6. **Thread Safety**: All UI updates use `fyne.Do()`, all state access uses `ca.mu`

---

## Files Modified

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/app.go` | Add 7 new widget.Label fields for data path displays |
| `module4-circuits/pkg/gui/tab_operations.go` | Major: createComputeModePanel(), onOpsCompute(), new helper functions |

---

## Technical Notes

### CRITICAL: Physics-Based Voltage Ranges

**The voltage ranges depend on the OPERATION TYPE based on hysteresis physics:**

| Operation | Voltage Range | Why |
|-----------|--------------|-----|
| **WRITE** (programming) | 2.0V - 5.0V | Must exceed coercive field Ec (~1.0-1.5 MV/cm) to switch FeFET polarization |
| **READ** (sensing) | 0.1V - 1.0V | SAFE zone - below Ec to avoid disturbing stored state |
| **COMPUTE** (MVM) | 0.0V - 1.0V | Same as READ - input voltages must NOT alter cell states |

From `module1-hysteresis/pkg/ferroelectric/material.go`:
- FeCIM HZO: Ec = 1.0e8 V/m (1.0 MV/cm)
- With 10nm film: Coercive voltage = Ec × thickness = 1.0V
- Safe READ margin: < 1.0V (below Ec)
- WRITE minimum: > 2.0V (well above Ec for reliable switching)

**COMPUTE mode uses 0-1V range because:**
1. MVM inputs are for READING conductance values, NOT writing
2. Higher voltages would reprogram the cells during computation (BAD!)
3. This matches the READ mode safe zone shown in the Voltage Zones diagram

### Peripheral Model Usage

**DAC (Digital-to-Analog Converter):**
- WRITE mode: Maps levels 0-29 to 2.0V-5.0V (programming voltages)
- COMPUTE mode: Maps bits 0-255 to 0.0V-1.0V (read/sense voltages)
- Method: `ca.dac.Convert(level) float64` (returns VrefLow to VrefHigh)
- Note: Default DAC VrefLow=-1.5V, VrefHigh=+1.5V is for WRITE mode
- For COMPUTE: Use normalized 0-1V mapping: `float64(bits) / 255.0`

**TIA (Transimpedance Amplifier):**
- Input: Current in Amps (convert uA to A: `rawCurrent * 1e-6`)
- Output: Voltage in Volts
- Method: `ca.tia.Convert(currentInAmps) float64`
- Default gain: 10 kOhm
- **CRITICAL: TIA SATURATION BEHAVIOR**
  - MaxInputCurrent: 100 uA (100e-6 A)
  - MaxOutputVoltage: 1.0V
  - When input current > 100 uA, output CLAMPS to 1.0V
  - Example: 160 uA input -> would be 1.6V, but clamps to 1.0V
  - UI should show "(SAT)" indicator when saturation occurs

**ADC (Analog-to-Digital Converter):**
- Input: Voltage in Volts (0 to 1V range)
- Output: Digital level (0 to 31 for 5-bit)
- Method: `ca.adc.Convert(voltage) int`
- Default: 5-bit SAR ADC
- **Level Calculation:** `level = round(voltage / 1.0 * 31)`
  - 0.0V -> Level 0
  - 0.5V -> Level 16 (round(0.5 * 31) = 16)
  - 1.0V -> Level 31
- When TIA saturates at 1.0V, ADC always reads Level 31

### Current Scale Consideration

The crossbar outputs are in micro-amps (uA) because:
- Conductance G is in micro-Siemens (uS): 1-100 uS range
- Voltage V is in Volts: 0-1V range
- Current I = G * V gives uA

When passing to TIA, must convert: `tia.Convert(current_uA * 1e-6)`

### Example Calculations (For UI Display)

**Example 1: Normal Operation (below saturation)**
```
Input: 50 uA row current
TIA:   50e-6 A * 10e3 Ohm = 0.5 V (below 1.0V max, no saturation)
ADC:   0.5V / 1.0V * 31 = 15.5 -> Level 16
Display: "y0: 50.0 uA | L16"
```

**Example 2: TIA Saturation (above 100 uA)**
```
Input: 160 uA row current
TIA:   160e-6 A * 10e3 Ohm = 1.6 V (EXCEEDS 1.0V max!)
       -> CLAMPS to 1.0V (saturation)
ADC:   1.0V / 1.0V * 31 = 31 -> Level 31
Display: "y0: 160.0 uA | L31 (SAT)"
```

**Example 3: Edge case at 100 uA**
```
Input: 100 uA row current (exactly at limit)
TIA:   100e-6 A * 10e3 Ohm = 1.0 V (exactly at max, no saturation indicator)
ADC:   1.0V / 1.0V * 31 = 31 -> Level 31
Display: "y0: 100.0 uA | L31"
```

### Variable Scoping Reference

```
createComputeModePanel() function:
|
+-- inputSection (local var, contains input grid)
|
+-- inputDataPathSection (local var, created in TODO 3)
|       |
|       +-- inputDigitalBox, inputDACBox, inputVoltageBox
|
+-- outputSection (local var, contains outputDataPathSection)
|       |
|       +-- outputDataPathSection (embedded, from TODO 4)
|             |
|             +-- outputCurrentBox, outputTIABox, outputADCBox, outputDigitalBox
|
+-- ca.computeConfigPanel = container.NewVBox(...)  [TODO 7 uses all above]
```

All local variables are scoped within `createComputeModePanel()` and remain valid throughout the function.
