# Module 4 COMPUTE Mode UI Revamp Plan

## Context

### Original Request
Redesign the COMPUTE mode panel in Module 4 to fix layout issues, remove redundant UI elements, and integrate DAC/ADC visualization directly into the crossbar array following the physical architecture.

### Interview Summary
- **Intent Classification**: Mid-sized UI Refactoring Task
- **Scope**: Primary file `module4-circuits/pkg/gui/tab_operations.go`, secondary `app.go`
- **Priority**: Layout correctness, physics visualization accuracy, compactness
- **Risk Tolerance**: Moderate - must preserve MVM computation logic

### Research Findings

**Current State Analysis (from `tab_operations.go`):**

1. **Input Vector Layout (lines 1060-1090):**
   - Uses `container.NewGridWithColumns(4)` creating 2x4 grid
   - Should be single horizontal row of 8 entries

2. **Mode Selector (lines 1093-1112):**
   - `modeSelect := widget.NewSelect([]string{"Manual", "Random", "Ramp"}, ...)`
   - RANDOM BITS button exists separately (lines 1133-1141)
   - Mode selector is redundant - only "Random" is useful

3. **INPUT PIPELINE Section (lines 1150-1182):**
   - Creates boxes: `digitalSummaryBox`, `dacSummaryBox`, `columnSummaryBox`
   - Connected with text arrows `->`
   - Redundant visual noise - info is repeated

4. **OUTPUT PIPELINE Section (lines 1184-1226):**
   - Creates boxes: `rowSumBox`, `tiaSummaryBox`, `adcSummaryBox`, `levelSummaryBox`
   - Does NOT show TIA stage properly in the main output display
   - Separated from crossbar visualization

5. **Crossbar Drawing (lines 376-443 in `drawSharedArray`):**
   - DAC labels drawn at TOP of grid (good location)
   - ADC labels drawn at RIGHT of grid (good location)
   - But actual DAC/ADC boxes are in separate section, not integrated

6. **Physics Issue - Voltage Range:**
   - Line 1173: Shows "0-1V (READ-safe)" but needs clarification that this won't exceed Ec

---

## Work Objectives

### Core Objective
Redesign COMPUTE mode panel for a compact, physically-accurate layout where:
- DACs are visually integrated at TOP of crossbar columns
- ADCs are visually integrated at RIGHT of crossbar rows
- Input vector is a single horizontal row
- Redundant elements are removed

### Target Layout (ASCII Mockup)

```
+------------------LEFT PANEL (60%)-------------------+  +-----RIGHT PANEL (40%)------+
|                                                     |  |                            |
|              x0   x1   x2   x3   x4   x5   x6   x7  |  |  INPUT VECTOR (1 row)      |
|             +---+---+---+---+---+---+---+---+       |  |  [x0][x1][x2][x3][x4]...   |
|             |DAC|DAC|DAC|DAC|DAC|DAC|DAC|DAC|       |  |  0.12 0.45 0.78 ...  V     |
|             | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 |       |  |                            |
|             +---+---+---+---+---+---+---+---+       |  |  [RANDOM BITS]             |
|        y0   |   |   |   |   |   |   |   |   |+---+ |  |                            |
|        y1   |   |   |   |   |   |   |   |   ||ADC| |  |  OUTPUT VECTOR             |
|        y2   |   |   |   |   |   |   |   |   || 0 | |  |  y0: 45.2 uA | L12         |
|        y3   | CROSSBAR  ARRAY  (8x8)   |   ||...| |  |  y1: 32.1 uA | L8          |
|        y4   |   |   |   |   |   |   |   |   || 7 | |  |  ...                       |
|        y5   |   |   |   |   |   |   |   |   |+---+ |  |                            |
|        y6   |   |   |   |   |   |   |   |   |      |  |  MATH BREAKDOWN            |
|        y7   |   |   |   |   |   |   |   |   |      |  |  I0 = G00*V0 + G01*V1 +..  |
|             +---+---+---+---+---+---+---+---+      |  |                            |
|                                                    |  |  PERFORMANCE               |
|  Legend: Low [====gradient====] High              |  |  DAC:5ns Array:5ns ADC:10ns|
|  0-1V READ-safe (below Ec coercive field)         |  |  TOTAL: ~20ns for MVM      |
+---------------------------------------------------+  +----------------------------+
```

### Deliverables
1. **Compact Input Vector** - Single horizontal row with 8 entries
2. **Remove Mode Selector** - Keep only RANDOM BITS button
3. **Remove INPUT/OUTPUT PIPELINE Boxes** - Redundant visual noise
4. **Integrated DAC Visualization** - Draw DAC boxes at TOP of crossbar columns
5. **Integrated ADC Visualization** - Draw ADC boxes at RIGHT of crossbar rows
6. **Physics Clarification** - Show TIA stage in output, clarify voltage safety
7. **Responsive HSplit Layout** - Left: crossbar+DACs+ADCs, Right: controls+outputs

### Definition of Done
- [ ] Input vector displays as single horizontal row of 8 entries
- [ ] Mode selector (Manual/Random/Ramp) removed
- [ ] RANDOM BITS button retained and functional
- [ ] INPUT PIPELINE boxes removed
- [ ] OUTPUT PIPELINE boxes removed (data shown in output labels instead)
- [ ] DAC boxes drawn integrated at top of crossbar columns
- [ ] ADC boxes drawn integrated at right of crossbar rows
- [ ] TIA conversion shown in output display
- [ ] READ-safe voltage clarification present
- [ ] All existing MVM computation logic preserved
- [ ] `go test ./module4-circuits/...` passes
- [ ] No compilation errors

---

## Guardrails

### Must Have
- Preserve all MVM computation logic in `computeAndUpdateAll()` (lines 1285-1332)
- Preserve input change handlers and auto-compute behavior
- Thread-safe UI updates using `fyne.Do()`
- Maintain backward compatibility with mode switching system

### Must NOT Have
- Breaking changes to `CircuitsApp` struct public interface
- New external dependencies
- Changes to peripheral package logic (`pkg/peripherals/`)
- Removal of existing working features

---

## Task Flow and Dependencies

```
[Phase 1: Remove Redundant Elements]
    |
    +-> Task 1.1: Remove mode selector widget
    |
    +-> Task 1.2: Remove INPUT PIPELINE section
    |
    +-> Task 1.3: Remove OUTPUT PIPELINE section
    |
    v
[Phase 2: Compact Input Layout]
    |
    +-> Task 2.1: Change input grid to horizontal row
    |
    +-> Task 2.2: Restructure input section
    |
    v
[Phase 3: Integrated Crossbar Visualization]
    |
    +-> Task 3.1: Update drawSharedArray for DAC boxes
    |
    +-> Task 3.2: Update drawSharedArray for ADC boxes
    |
    +-> Task 3.3: Add live value display in DAC/ADC boxes
    |
    v
[Phase 4: Output Enhancement]
    |
    +-> Task 4.1: Show TIA stage in output labels
    |
    +-> Task 4.2: Add physics clarification text
    |
    v
[Phase 5: Layout Restructure]
    |
    +-> Task 5.1: Create new compact right panel
    |
    +-> Task 5.2: Adjust HSplit ratio
    |
    v
[Verification]
```

---

## Detailed TODOs

### Phase 1: Remove Redundant Elements

#### Task 1.1: Remove Mode Selector Widget
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1093-1112

**Current Code:**
```go
// Input mode selector
modeSelect := widget.NewSelect([]string{"Manual", "Random", "Ramp"}, func(s string) {
    switch s {
    case "Random":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = rand.Intn(256)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
        ca.computeAndUpdateAll()
    case "Ramp":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
        ca.computeAndUpdateAll()
    }
})
modeSelect.SetSelected("Manual")
```

**Action:** Delete lines 1093-1113. The RANDOM BITS button (lines 1133-1141) already provides the needed functionality.

**Acceptance Criteria:**
- Mode selector no longer appears in COMPUTE panel
- RANDOM BITS button still functions
- No compilation errors

---

#### Task 1.2: Remove INPUT PIPELINE Section
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1150-1182

**Current Code:**
```go
// FULL INPUT PIPELINE: 8 digital values -> 8 DACs -> 8 column voltages
inputPipelineHeader := widget.NewLabelWithStyle(...)
ca.opsComputeInputDigitalLabel = widget.NewLabel("x0-x7\n(0-255)")
ca.opsComputeInputDACLabel = widget.NewLabel("-> 0-1V\neach")
digitalSummaryBox := ca.createLabeledBoxWithLabel(...)
dacSummaryBox := ca.createLabeledBoxWithLabel(...)
columnSummaryBox := ca.createLabeledBox(...)
inputPipelinePath := container.NewHBox(...)
inputPhysicsNote := widget.NewLabel(...)
inputDataPathSection := container.NewVBox(...)
```

**Action:** Delete entire INPUT PIPELINE section (lines 1150-1182). Keep only the physics note about READ-safe voltages, move it elsewhere.

**Acceptance Criteria:**
- INPUT PIPELINE boxes no longer appear
- Physics safety note preserved (moved to input section)
- No compilation errors

---

#### Task 1.3: Remove OUTPUT PIPELINE Section
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1184-1226

**Current Code:**
```go
// FULL OUTPUT PIPELINE: 8 row sums -> 8 TIAs -> 8 ADCs -> 8 digital levels
outputPipelineHeader := widget.NewLabelWithStyle(...)
ca.opsComputeOutputCurrentLabel = widget.NewLabel("y0-y7\n(KCL)")
ca.opsComputeOutputTIALabel = widget.NewLabel("I->V\n10k")
ca.opsComputeOutputADCLabel = widget.NewLabel("5-bit\n0-31")
rowSumBox := ca.createLabeledBoxWithLabel(...)
tiaSummaryBox := ca.createLabeledBoxWithLabel(...)
adcSummaryBox := ca.createLabeledBoxWithLabel(...)
levelSummaryBox := ca.createLabeledBox(...)
outputPipelinePath := container.NewHBox(...)
outputPhysicsNote := widget.NewLabel(...)
idealDisclaimer := widget.NewLabel(...)
outputDataPathSection := container.NewVBox(...)
```

**Action:** Delete entire OUTPUT PIPELINE section (lines 1184-1226). Keep the ideal crossbar disclaimer, move it to output section.

**Acceptance Criteria:**
- OUTPUT PIPELINE boxes no longer appear
- Ideal crossbar disclaimer preserved (moved to output section)
- No compilation errors

---

### Phase 2: Compact Input Layout

#### Task 2.1: Change Input Grid to Horizontal Row
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1060-1090

**Current Code:**
```go
inputGrid := container.NewGridWithColumns(4)  // Creates 2x4 grid
maxDisplay := min(8, ca.arrayCols)
for i := 0; i < maxDisplay; i++ {
    ca.opsComputeInputs[i] = widget.NewEntry()
    // ... entry setup ...
    inputGrid.Add(container.NewVBox(
        widget.NewLabel(fmt.Sprintf("x%d", i)),
        ca.opsComputeInputs[i],
        ca.opsComputeVoltageLabels[i],
    ))
}
```

**New Code:**
```go
// Create horizontal input row with compact entries
inputRow := container.NewHBox()
maxDisplay := min(8, ca.arrayCols)
for i := 0; i < maxDisplay; i++ {
    ca.opsComputeInputs[i] = widget.NewEntry()
    ca.opsComputeInputs[i].SetText(fmt.Sprintf("%d", ca.inputVector[i]))
    ca.opsComputeInputs[i].Resize(fyne.NewSize(45, 30))  // Compact width

    idx := i
    ca.opsComputeInputs[i].OnChanged = func(s string) {
        var v int
        fmt.Sscanf(s, "%d", &v)
        if v > 255 {
            v = 255
        }
        ca.mu.Lock()
        ca.inputVector[idx] = v
        ca.mu.Unlock()
        ca.computeAndUpdateAll()
    }

    // Compact column: label on top, entry below, voltage below
    ca.opsComputeVoltageLabels[i] = widget.NewLabel(fmt.Sprintf("%.2fV", float64(ca.inputVector[i])/255.0))
    ca.opsComputeVoltageLabels[i].TextStyle = fyne.TextStyle{Monospace: true}

    col := container.NewVBox(
        widget.NewLabel(fmt.Sprintf("x%d", i)),
        ca.opsComputeInputs[i],
    )
    inputRow.Add(col)
}
```

**Acceptance Criteria:**
- Input vector displays as single horizontal row (x0 through x7)
- Each entry shows label above, entry below
- Entries are compact width (~45px)
- Auto-compute still triggers on input change

---

#### Task 2.2: Restructure Input Section
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1143-1148

**Current Code:**
```go
inputSection := container.NewVBox(
    widget.NewLabelWithStyle("INPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    container.NewHBox(widget.NewLabel("Mode:"), modeSelect, randomBitsBtn),
    widget.NewLabel("Digital inputs (0-255) -> DAC voltages (0-1V):"),
    inputGrid,
)
```

**New Code:**
```go
// Compact input section
inputHeader := container.NewHBox(
    widget.NewLabelWithStyle("INPUT VECTOR (0-255)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    layout.NewSpacer(),
    randomBitsBtn,
)

physicsNote := widget.NewLabel("0-1V READ-safe (below Ec)")
physicsNote.TextStyle = fyne.TextStyle{Italic: true}

inputSection := container.NewVBox(
    inputHeader,
    inputRow,
    physicsNote,
)
```

**Acceptance Criteria:**
- Input section is compact
- RANDOM BITS button is inline with header
- Physics note appears below inputs
- Mode selector removed

---

### Phase 3: Integrated Crossbar Visualization

#### Task 3.1: Update drawSharedArray for DAC Boxes
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 411-426 (inside ModeCompute case of drawSharedArray)

**Current Code:**
```go
// Add column labels at TOP (x0-x7) - light blue for inputs
inputLabelColor := color.RGBA{100, 150, 255, 255}
for c := 0; c < min(8, cols); c++ {
    x := offsetX + c*cellSize + cellSize/2 - 6
    y := offsetY - 30
    if y >= 5 {
        drawSimpleText(img, fmt.Sprintf("x%d", c), x, y, inputLabelColor)
    }
}

// Add "8 DACs" label centered above grid
dacLabelX := offsetX + gridW/2 - 20
dacLabelY := offsetY - 40
if dacLabelY >= 5 {
    drawSimpleText(img, "8 DACs", dacLabelX, dacLabelY, inputLabelColor)
}
```

**New Code:**
```go
// Draw DAC boxes at TOP of each column (integrated visualization)
dacBoxHeight := 25
dacBoxWidth := cellSize - 4
dacY := offsetY - dacBoxHeight - 10
dacColor := color.RGBA{100, 80, 180, 255}  // Purple for DACs

// Define input label color for column labels (light blue for inputs)
inputLabelColor := color.RGBA{100, 150, 255, 255}

// OPTIMIZATION: Copy input vector data once before loop to avoid RLock per iteration
dacColCount := min(8, cols)
inputVectorCopy := make([]int, dacColCount)
ca.mu.RLock()
copy(inputVectorCopy, ca.inputVector[:dacColCount])
ca.mu.RUnlock()

for c := 0; c < dacColCount; c++ {
    dacX := offsetX + c*cellSize + 2

    // Draw DAC box
    drawRect(img, dacX, dacY, dacBoxWidth, dacBoxHeight, dacColor)

    // Draw border
    borderColor := color.RGBA{150, 130, 220, 255}
    drawRectBorder(img, dacX, dacY, dacBoxWidth, dacBoxHeight, borderColor)

    // Use pre-copied input value (no lock needed)
    inputVal := inputVectorCopy[c]
    voltage := float64(inputVal) / 255.0

    // Show voltage value
    voltageText := fmt.Sprintf("%.1fV", voltage)
    textX := dacX + dacBoxWidth/2 - len(voltageText)*3
    textY := dacY + dacBoxHeight/2 - 3
    drawSimpleText(img, voltageText, textX, textY, color.RGBA{255, 255, 255, 255})

    // Draw column label below DAC box
    labelX := offsetX + c*cellSize + cellSize/2 - 6
    labelY := dacY - 12
    drawSimpleText(img, fmt.Sprintf("x%d", c), labelX, labelY, inputLabelColor)
}
```

**Add helper function to `module4-circuits/pkg/gui/helpers.go` at line 21 (after `drawRect` function):**
```go
// drawRectBorder draws only the border (outline) of a rectangular region.
// It performs boundary checks to ensure all pixels are within image bounds.
func drawRectBorder(img *image.RGBA, x, y, rectW, rectH int, c color.Color) {
    bounds := img.Bounds()
    maxX := bounds.Dx()
    maxY := bounds.Dy()

    // Top edge
    if y >= 0 && y < maxY {
        for px := x; px < x+rectW; px++ {
            if px >= 0 && px < maxX {
                img.Set(px, y, c)
            }
        }
    }
    // Bottom edge
    bottomY := y + rectH - 1
    if bottomY >= 0 && bottomY < maxY {
        for px := x; px < x+rectW; px++ {
            if px >= 0 && px < maxX {
                img.Set(px, bottomY, c)
            }
        }
    }
    // Left edge
    if x >= 0 && x < maxX {
        for py := y; py < y+rectH; py++ {
            if py >= 0 && py < maxY {
                img.Set(x, py, c)
            }
        }
    }
    // Right edge
    rightX := x + rectW - 1
    if rightX >= 0 && rightX < maxX {
        for py := y; py < y+rectH; py++ {
            if py >= 0 && py < maxY {
                img.Set(rightX, py, c)
            }
        }
    }
}
```

**Acceptance Criteria:**
- DAC boxes appear at TOP of each column
- Each DAC box shows current voltage value
- Column labels (x0-x7) appear above DAC boxes
- Boxes update when input values change

---

#### Task 3.2: Update drawSharedArray for ADC Boxes
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 428-443 (inside ModeCompute case)

**Current Code:**
```go
// Add row labels at RIGHT (y0-y7) - light orange for outputs
outputLabelColor := color.RGBA{255, 180, 100, 255}
for r := 0; r < min(8, rows); r++ {
    x := offsetX + gridW + 20
    y := offsetY + r*cellSize + cellSize/2 - 3
    if x < w-15 {
        drawSimpleText(img, fmt.Sprintf("y%d", r), x, y, outputLabelColor)
    }
}

// Add "8 ADCs" label to the right of grid
adcLabelX := offsetX + gridW + 20
adcLabelY := offsetY + gridH + 10
if adcLabelX < w-35 && adcLabelY < h-5 {
    drawSimpleText(img, "8 ADCs", adcLabelX, adcLabelY, outputLabelColor)
}
```

**New Code:**
```go
// Draw ADC boxes at RIGHT of each row (integrated visualization)
adcBoxWidth := 45
adcBoxHeight := cellSize - 4
adcX := offsetX + gridW + 8
adcColor := color.RGBA{80, 150, 100, 255}  // Green for ADCs

// Define output label color for row labels (light orange for outputs)
outputLabelColor := color.RGBA{255, 180, 100, 255}

// OPTIMIZATION: Copy output vector data once before loop to avoid RLock per iteration
adcRowCount := min(8, rows)
outputVectorCopy := make([]float64, adcRowCount)
ca.mu.RLock()
copy(outputVectorCopy, ca.outputVector[:min(adcRowCount, len(ca.outputVector))])
ca.mu.RUnlock()

for r := 0; r < adcRowCount; r++ {
    adcY := offsetY + r*cellSize + 2

    // Draw ADC box
    drawRect(img, adcX, adcY, adcBoxWidth, adcBoxHeight, adcColor)

    // Draw border
    borderColor := color.RGBA{130, 200, 150, 255}
    drawRectBorder(img, adcX, adcY, adcBoxWidth, adcBoxHeight, borderColor)

    // Use pre-copied output value (no lock needed)
    outputVal := outputVectorCopy[r]

    // Show ADC level (after TIA+ADC conversion)
    tiaVoltage := ca.tia.Convert(outputVal * 1e-6)
    adcLevel := ca.adc.Convert(tiaVoltage)

    levelText := fmt.Sprintf("L%d", adcLevel)
    textX := adcX + adcBoxWidth/2 - len(levelText)*3
    textY := adcY + adcBoxHeight/2 - 3
    drawSimpleText(img, levelText, textX, textY, color.RGBA{255, 255, 255, 255})

    // Draw row label to right of ADC box
    labelX := adcX + adcBoxWidth + 5
    labelY := offsetY + r*cellSize + cellSize/2 - 3
    drawSimpleText(img, fmt.Sprintf("y%d", r), labelX, labelY, outputLabelColor)
}
```

**Acceptance Criteria:**
- ADC boxes appear at RIGHT of each row
- Each ADC box shows current ADC level (L0-L31)
- Row labels (y0-y7) appear to right of ADC boxes
- Boxes update when computation completes

---

#### Task 3.3: Expand Canvas Size for Integrated Components
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 200-201 (createSharedArraySection)

**Current Code:**
```go
tappableArray := NewTappableArrayCanvas(ca, ca.drawSharedArray, ca.onArrayCellTapped)
tappableArray.SetMinSize(fyne.NewSize(400, 350))
```

**New Code:**
```go
tappableArray := NewTappableArrayCanvas(ca, ca.drawSharedArray, ca.onArrayCellTapped)
// Larger size to accommodate integrated DAC (top) and ADC (right) boxes
tappableArray.SetMinSize(fyne.NewSize(480, 420))
```

**Also update margin and grid position calculation in drawSharedArray (lines 252-277):**

**Current Code:**
```go
// Calculate cell size (use square cells)
margin := 40
cellW := (w - 2*margin) / cols
cellH := (h - 2*margin) / rows
cellSize := cellW
if cellH < cellSize {
    cellSize = cellH
}
if cellSize > 40 {
    cellSize = 40
}
if cellSize < 8 {
    cellSize = 8
}

// Store cell geometry for click detection
ca.mu.Lock()
ca.sharedArrayCellSize = cellSize
ca.sharedArrayOffsetX = (w - cols*cellSize) / 2
ca.sharedArrayOffsetY = (h - rows*cellSize) / 2
ca.mu.Unlock()

// Center the grid
gridW := cols * cellSize
gridH := rows * cellSize
offsetX := (w - gridW) / 2
offsetY := (h - gridH) / 2
```

**New Code (COMPLETE MARGIN REFACTOR):**
```go
// Calculate asymmetric margins for integrated DAC (top) and ADC (right)
topMargin := 70    // Space for DAC boxes + labels above grid
rightMargin := 70  // Space for ADC boxes + labels right of grid
bottomMargin := 30
leftMargin := 30

// Calculate available area for grid
availableW := w - leftMargin - rightMargin
availableH := h - topMargin - bottomMargin

// Calculate cell size (use square cells)
cellW := availableW / cols
cellH := availableH / rows
cellSize := cellW
if cellH < cellSize {
    cellSize = cellH
}
if cellSize > 40 {
    cellSize = 40
}
if cellSize < 8 {
    cellSize = 8
}

// Calculate grid dimensions
gridW := cols * cellSize
gridH := rows * cellSize

// Calculate offset to center grid within available area
offsetX := leftMargin + (availableW-gridW)/2
offsetY := topMargin + (availableH-gridH)/2

// Store cell geometry for click detection
ca.mu.Lock()
ca.sharedArrayCellSize = cellSize
ca.sharedArrayOffsetX = offsetX
ca.sharedArrayOffsetY = offsetY
ca.mu.Unlock()
```

**Also update TappableArrayCanvas.Tapped() (lines 78-100) to use same margin logic:**

**Current Code:**
```go
// Recalculate cell geometry using same logic as drawSharedArray
w := int(size.Width)
h := int(size.Height)
margin := 40
cellW := (w - 2*margin) / cols
cellH := (h - 2*margin) / rows
// ... cellSize calculations ...

// Calculate grid size and centering offset
gridW := cols * cellSize
gridH := rows * cellSize
offsetX := (w - gridW) / 2
offsetY := (h - gridH) / 2
```

**New Code:**
```go
// Recalculate cell geometry using same logic as drawSharedArray
w := int(size.Width)
h := int(size.Height)

// Use same asymmetric margins as drawSharedArray
topMargin := 70
rightMargin := 70
bottomMargin := 30
leftMargin := 30

availableW := w - leftMargin - rightMargin
availableH := h - topMargin - bottomMargin

cellW := availableW / cols
cellH := availableH / rows

// Cell size calculations (COMPLETE - matching drawSharedArray exactly)
cellSize := cellW
if cellH < cellSize {
    cellSize = cellH
}
if cellSize > 40 {
    cellSize = 40
}
if cellSize < 8 {
    cellSize = 8
}
if cellSize <= 0 {
    return
}

// Calculate grid size and offset
gridW := cols * cellSize
gridH := rows * cellSize
offsetX := leftMargin + (availableW-gridW)/2
offsetY := topMargin + (availableH-gridH)/2
```

**Acceptance Criteria:**
- Canvas large enough to show DAC boxes at top
- Canvas large enough to show ADC boxes at right
- Grid positioned with asymmetric margins (more space top/right for DAC/ADC)
- Click detection uses same offset calculations as drawing
- Cell selection still works correctly after margin changes

---

### Phase 4: Output Enhancement

#### Task 4.1: Show TIA Stage in Output Labels
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1302-1322 (inside computeAndUpdateAll)

**Current Code:**
```go
for i := 0; i < 8 && i < len(ca.outputVector); i++ {
    if ca.opsComputeOutputLabels[i] != nil {
        rawCurrent := ca.outputVector[i]
        tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)
        adcLevel := ca.adc.Convert(tiaVoltage)
        isSaturated := rawCurrent > 100.0

        idx := i
        current := rawCurrent
        level := adcLevel
        sat := isSaturated
        fyne.Do(func() {
            if sat {
                ca.opsComputeOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1f uA | L%d (SAT)", idx, current, level))
            } else {
                ca.opsComputeOutputLabels[idx].SetText(fmt.Sprintf("y%d: %.1f uA | L%d", idx, current, level))
            }
        })
    }
}
```

**New Code:**
```go
for i := 0; i < 8 && i < len(ca.outputVector); i++ {
    if ca.opsComputeOutputLabels[i] != nil {
        rawCurrent := ca.outputVector[i]
        tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6)
        adcLevel := ca.adc.Convert(tiaVoltage)
        isSaturated := rawCurrent > 100.0

        idx := i
        current := rawCurrent
        tiaV := tiaVoltage
        level := adcLevel
        sat := isSaturated
        fyne.Do(func() {
            // Show full pipeline: Current -> TIA Voltage -> ADC Level
            satSuffix := ""
            if sat {
                satSuffix = " SAT"
            }
            ca.opsComputeOutputLabels[idx].SetText(
                fmt.Sprintf("y%d: %.1fuA -> %.2fV -> L%d%s", idx, current, tiaV, level, satSuffix))
        })
    }
}
```

**Acceptance Criteria:**
- Output labels show full conversion path: current -> TIA voltage -> ADC level
- Saturation indicator still shown when applicable
- Labels are compact but informative

---

#### Task 4.2: Add Physics Clarification Text
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1229-1235 (outputSection)

**Current Code:**
```go
outputSection := container.NewVBox(
    widget.NewSeparator(),
    widget.NewLabelWithStyle("OUTPUT VECTOR (Row Sums)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    widget.NewLabel("Each output = sum of 8 cell currents, digitized by TIA+ADC:"),
    outputGrid,
    outputDataPathSection,
)
```

**New Code:**
```go
// Ideal crossbar disclaimer (moved from removed section)
idealDisclaimer := widget.NewLabel(
    "IDEAL CROSSBAR: No IR drop or sneak paths (see Module 2)")
idealDisclaimer.TextStyle = fyne.TextStyle{Italic: true}

outputSection := container.NewVBox(
    widget.NewSeparator(),
    widget.NewLabelWithStyle("OUTPUT VECTOR", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
    widget.NewLabel("I_row -> TIA (10k) -> ADC (5-bit):"),
    outputGrid,
    idealDisclaimer,
)
```

**Acceptance Criteria:**
- Output section header is clear
- TIA stage mentioned in description
- Ideal crossbar disclaimer preserved

---

### Phase 5: Layout Restructure

#### Task 5.1: Create New Compact Right Panel
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 1253-1259 (computeConfigPanel assembly)

**Current Code:**
```go
ca.computeConfigPanel = container.NewVBox(
    inputSection,
    inputDataPathSection,
    outputSection,
    mathSection,
    perfSection,
)
```

**New Code:**
```go
// Compact right panel without redundant pipeline sections
ca.computeConfigPanel = container.NewVBox(
    inputSection,      // Horizontal input row + RANDOM BITS
    outputSection,     // Output with TIA info
    mathSection,       // Math breakdown
    perfSection,       // Performance timing
)
```

**Acceptance Criteria:**
- Right panel is compact
- No redundant pipeline sections
- All essential information preserved

---

#### Task 5.2: Adjust HSplit Ratio
**File:** `module4-circuits/pkg/gui/tab_operations.go`
**Lines:** 158-162 (in createOperationsView)

**Current Code:**
```go
mainContent := container.NewHSplit(
    arraySection,
    rightPanel,
)
mainContent.SetOffset(0.4) // Array gets 40% width
```

**New Code:**
```go
mainContent := container.NewHSplit(
    arraySection,
    rightPanel,
)
mainContent.SetOffset(0.55) // Array gets 55% width (more space for integrated DAC/ADC)
```

**Acceptance Criteria:**
- Left panel (crossbar) gets more space for integrated components
- Right panel (controls) remains usable but more compact

---

## Commit Strategy

### Commit 1: Remove Redundant Elements
```
refactor(module4): remove redundant compute mode UI elements

- Remove Manual/Random/Ramp mode selector (RANDOM BITS button sufficient)
- Remove INPUT PIPELINE visualization boxes
- Remove OUTPUT PIPELINE visualization boxes
- Preserve physics notes and disclaimers
```

### Commit 2: Compact Input Layout
```
feat(module4): compact input vector to single horizontal row

- Change input grid from 2x4 to 1x8 horizontal layout
- Move RANDOM BITS button inline with header
- Add READ-safe voltage note below inputs
```

### Commit 3: Integrated DAC/ADC Visualization
```
feat(module4): integrate DAC/ADC boxes into crossbar visualization

- Draw DAC boxes at top of each column showing voltage
- Draw ADC boxes at right of each row showing level
- Add drawRectBorder helper function
- Expand canvas size for integrated components
- Update margin calculations
```

### Commit 4: Output Enhancement
```
feat(module4): show full TIA conversion path in output labels

- Update output labels: current -> TIA voltage -> ADC level
- Move ideal crossbar disclaimer to output section
- Add TIA stage description to output header
```

### Commit 5: Layout Polish
```
feat(module4): adjust compute mode layout ratios

- Increase left panel width for integrated visualization
- Remove empty space from right panel
- Final layout polish
```

---

## Success Criteria

### Functional
- [ ] MVM computation produces correct results
- [ ] Input changes trigger auto-compute
- [ ] RANDOM BITS button randomizes all inputs
- [ ] Output labels update with computation
- [ ] Mode switching between WRITE/READ/COMPUTE works

### Visual
- [ ] Input vector is single horizontal row (x0-x7)
- [ ] DAC boxes integrated at top of crossbar columns
- [ ] ADC boxes integrated at right of crossbar rows
- [ ] No redundant INPUT/OUTPUT PIPELINE boxes
- [ ] TIA stage visible in output display
- [ ] Physics safety notes present

### Code Quality
- [ ] No compilation errors
- [ ] `go test ./module4-circuits/...` passes
- [ ] Thread-safe updates preserved
- [ ] No memory leaks on repeated computation

### Performance
- [ ] Canvas renders in <16ms (60fps)
- [ ] No lag on input changes
- [ ] Auto-compute responsive

---

## Risk Assessment

### High Risk
| Risk | Mitigation |
|------|------------|
| Breaking MVM computation | Preserve `computeAndUpdateAll()` logic exactly |
| Canvas drawing performance | Profile with 8x8 grid, optimize if needed |
| Thread safety on draw | Use RLock for reading outputVector in draw |

### Medium Risk
| Risk | Mitigation |
|------|------------|
| Layout sizing issues | Test at multiple window sizes |
| Text overflow in DAC/ADC boxes | Use short formats (L##, #.#V) |

### Low Risk
| Risk | Mitigation |
|------|------------|
| Color consistency | Use existing color patterns from draw code |

---

## Verification Steps

1. **Build Verification**
   ```bash
   go build ./cmd/fecim-lattice-tools
   ```

2. **Test Verification**
   ```bash
   go test ./module4-circuits/...
   ```

3. **Visual Verification**
   - Launch app, navigate to Module 4 OPERATIONS
   - Select COMPUTE mode
   - Verify input vector is horizontal row
   - Verify DAC boxes at top of columns show voltages
   - Verify ADC boxes at right of rows show levels
   - Click RANDOM BITS, verify all update
   - Check output labels show TIA conversion

4. **Computation Verification**
   - Set known input values (e.g., all 128)
   - Verify output currents are reasonable
   - Verify ADC levels match expected conversion

---

## Files Modified

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/tab_operations.go` | Remove redundant UI, update layout, integrate DAC/ADC, update margin calculations |
| `module4-circuits/pkg/gui/helpers.go` | Add `drawRectBorder()` helper function at line 21 (after `drawRect`) |

## Estimated LOC Changes

| Type | Lines |
|------|-------|
| Deleted | ~120 (redundant pipeline sections) |
| Modified | ~100 (layout restructuring, margin refactor) |
| Added | ~90 (integrated DAC/ADC drawing, drawRectBorder helper) |
| **Net** | **~70 more lines** |
