# Module 4 UI/UX Refactor Plan

## Context

### Original Request
Comprehensive UI/UX improvements for Module 4 (Peripheral Circuits) to align with project-wide design standards and improve educational effectiveness.

### Interview Summary
- **Intent Classification**: Mid-sized Refactoring Task
- **Scope**: 9 files in `module4-circuits/pkg/gui/`
- **Priority**: Visual hierarchy, theme consistency, interactivity
- **Risk Tolerance**: Moderate - must preserve all existing functionality

### Research Findings

**Current State Analysis:**
1. Module 4 has its own local theme in `theme.go` that differs from `shared/theme/theme.go`
2. Data path arrows use plain text `"->"`  instead of styled visual elements
3. No hover states on interactive elements (array cells)
4. Canvas drawing uses raw pixel manipulation without legends
5. Section headers use basic `widget.NewLabel()` without visual distinction
6. Fixed layouts don't respond to window size changes
7. Status feedback labels positioned at bottom, easily missed

**Available Resources:**
- `shared/theme/theme.go` - Comprehensive FeCIM theme with documented colors
- `shared/widgets/color_legend.go` - Reusable ColorLegend widget
- `shared/widgets/adaptive_layout.go` - Responsive layout system
- `shared/widgets/resize_detector.go` - Breakpoint detection utilities

---

## Work Objectives

### Core Objective
Modernize Module 4 UI to match project standards while enhancing educational clarity through improved visual hierarchy, interactive feedback, and consistent theming.

### Deliverables
1. **Theme Migration** - Use `shared/theme` instead of local colors
2. **Enhanced Section Headers** - Bold, colored headers with clear hierarchy
3. **Interactive Array Grid** - Hover states, click feedback, cursor changes
4. **Styled Data Flow Visualization** - Canvas-drawn arrows, component boxes with shadows
5. **Chart Legends** - ColorLegend widgets on all canvases
6. **Educational Tooltips** - Help text for parameters and concepts
7. **Responsive Layout** - Adapt to window sizes using AdaptiveLayout
8. **Prominent Status Feedback** - Toast-style or header-area notifications

### Definition of Done
- [ ] All 6 tabs render without visual defects
- [ ] Theme colors match `shared/theme/theme.go` exactly
- [ ] Array grid responds to hover with visual feedback
- [ ] Data path shows animated/styled arrows (not plain text)
- [ ] All canvases have legends explaining color mapping
- [ ] Help tooltips on key parameters (voltage, levels, bits)
- [ ] Layout adapts gracefully to window resize
- [ ] All existing functionality preserved (verified by manual testing)
- [ ] No compilation errors
- [ ] `go test ./module4-circuits/...` passes

---

## Guardrails

### Must Have
- Backward compatibility with existing functionality
- Consistent use of `shared/theme` colors
- Thread-safe UI updates using `fyne.Do()`
- All canvas operations bounded to image dimensions
- Preserve educational content and accuracy

### Must NOT Have
- New external dependencies
- Breaking changes to EmbeddedCircuitsApp interface
- Animations that can't be disabled
- Hard-coded pixel values without responsive scaling
- Changes to business logic in `pkg/peripherals/`

---

## Task Flow and Dependencies

```
[Phase 1: Theme Foundation]
    |
    +-> Task 1.1: Delete local theme, import shared theme
    |
    +-> Task 1.2: Update color references in all files
    |
    v
[Phase 2: Visual Hierarchy]
    |
    +-> Task 2.1: Create styled section header widget
    |
    +-> Task 2.2: Apply headers to all 6 tabs
    |
    v
[Phase 3: Data Flow Visualization]
    |
    +-> Task 3.1: Create arrow drawing function
    |
    +-> Task 3.2: Create styled component box widget
    |
    +-> Task 3.3: Update WRITE data path
    |
    +-> Task 3.4: Update READ data path
    |
    +-> Task 3.5: Update COMPUTE visualization
    |
    v
[Phase 4: Interactive Array]
    |
    +-> Task 4.1: Add hover state tracking
    |
    +-> Task 4.2: Add visual hover feedback in array drawing
    |
    +-> Task 4.3: Add click feedback animation
    |
    v
[Phase 5: Chart Improvements]
    |
    +-> Task 5.1: Add ColorLegend to WRITE array
    |
    +-> Task 5.2: Add ColorLegend to COMPUTE array
    |
    +-> Task 5.3: Style bar charts in COMPARISON tab
    |
    v
[Phase 6: Tooltips]
    |
    +-> Task 6.1: Add tooltips to configuration inputs
    |
    +-> Task 6.2: Add tooltips to data path components
    |
    v
[Phase 7: Responsive Layout]
    |
    +-> Task 7.1: Wrap each tab content with AdaptiveLayout
    |
    +-> Task 7.2: Define breakpoint behaviors
    |
    v
[Phase 8: Status Feedback]
    |
    +-> Task 8.1: Create prominent status display area
    |
    +-> Task 8.2: Update all status update calls
    |
    v
[Verification]
```

---

## Detailed TODOs

### Phase 1: Theme Foundation

#### Task 1.1: Remove Local Theme and Import Shared Theme
**File:** `module4-circuits/pkg/gui/theme.go`
**Lines:** 1-66 (entire file)

**Action:** Delete entire file, it will be replaced by import of shared theme.

**Acceptance Criteria:**
- File `theme.go` no longer exists in module4
- Compilation succeeds after updates in Task 1.2

---

#### Task 1.2: Update All Color References
**Files:** Multiple files need import and reference updates

**File:** `module4-circuits/pkg/gui/app.go`
**Lines:** 18-20, 141-142

**Changes:**
```go
// Add import
import (
    ...
    sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

// Line 142: Change theme assignment
ca.fyneApp.Settings().SetTheme(&sharedtheme.FeCIMTheme{})
```

**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 7, 271-273

**Changes:**
```go
// Add import
import (
    ...
    sharedtheme "multilayer-ferroelectric-cim-visualizer/shared/theme"
)

// Update color references in createWriteDataPathSection (lines 271-273)
digitalBox := ca.createLabeledBoxWithLabel("DIGITAL", ca.writeDigitalLabel, sharedtheme.ColorPrimary)
dacBox := ca.createLabeledBoxWithLabel("DAC", ca.writeDACLabel, sharedtheme.ColorAccent)  // Was colorDAC
fefetBox := ca.createLabeledBoxWithLabel("FeFET", ca.writeFeFETLabel, sharedtheme.ColorInfo) // Was colorArrayCell
```

**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 366-397 (drawWritePulse function)

**Changes:** Replace hardcoded colors with shared theme:
```go
bgColor := sharedtheme.ColorBackground  // Was color.RGBA{0, 40, 80, 255}
cyanColor := sharedtheme.ColorPrimary   // Was color.RGBA{0, 255, 255, 255}
fillColor := sharedtheme.WithAlpha(sharedtheme.ColorPrimary, 100) // Was color.RGBA{0, 100, 150, 200}
threshColor := sharedtheme.ColorWarning  // Was color.RGBA{255, 200, 0, 255}
```

**File:** `module4-circuits/pkg/gui/tab_read.go`

**Changes:** Similar pattern - import shared theme and replace local color vars.

**File:** `module4-circuits/pkg/gui/tab_compute.go`

**Changes:** Similar pattern.

**File:** `module4-circuits/pkg/gui/tab_comparison.go`
**Lines:** 99, 120-121, 141-143

**Changes:** Replace colorCPU, colorGPU, colorFeFET with shared theme colors:
```go
// Define semantic colors for comparison (add near top of function)
cpuColor := sharedtheme.ColorError       // Red-ish
gpuColor := sharedtheme.ColorSuccess     // Green
fefetColor := sharedtheme.ColorPrimary   // Cyan (brand)
```

**File:** `module4-circuits/pkg/gui/tab_timing.go`

**Changes:** Replace hardcoded colors with shared theme.

**Acceptance Criteria:**
- All files compile without errors
- No references to local `color*` variables remain
- App launches and displays with consistent theme

---

### Phase 2: Visual Hierarchy

#### Task 2.1: Create Styled Section Header Helper
**File:** `module4-circuits/pkg/gui/helpers.go`
**Lines:** Add after line 35

**New Code:**
```go
// createSectionHeader creates a styled section header with icon-like prefix
func createSectionHeader(title string) *fyne.Container {
    headerLabel := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    // Create accent bar (visual indicator)
    accentBar := canvas.NewRectangle(sharedtheme.ColorPrimary)
    accentBar.SetMinSize(fyne.NewSize(4, 20))
    accentBar.CornerRadius = 2

    return container.NewHBox(
        accentBar,
        layout.NewSpacer(),
        headerLabel,
        layout.NewSpacer(),
    )
}

// createSubsectionHeader creates a lighter header for subsections
func createSubsectionHeader(title string) *widget.Label {
    label := widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    return label
}
```

**Acceptance Criteria:**
- Function compiles
- Returns a container with styled header

---

#### Task 2.2: Apply Section Headers to All Tabs
**Files:** `tab_write.go`, `tab_read.go`, `tab_compute.go`, `tab_comparison.go`, `tab_timing.go`, `tab_specs.go`

**Pattern for each file:**
Replace:
```go
widget.NewLabel("CONFIGURATION"),
```
With:
```go
createSectionHeader("CONFIGURATION"),
```

**Specific locations:**

**tab_write.go lines:** 60, 63, 67, 70, 74, 77, 94
**tab_read.go lines:** 59, 62, 67, 70, 74, 77
**tab_compute.go lines:** 60, 63, 68, 71, 75, 78
**tab_comparison.go lines:** 54-55, 58-59
**tab_timing.go lines:** 55, 57, 59, 61
**tab_specs.go lines:** 54-59

**Acceptance Criteria:**
- All tabs have visually distinct section headers
- Headers use accent color bar
- Hierarchy is clear: Tab -> Section -> Content

---

### Phase 3: Data Flow Visualization

#### Task 3.1: Create Arrow Drawing Function
**File:** `module4-circuits/pkg/gui/drawing.go`
**Lines:** Add after line 23

**New Code:**
```go
// drawArrow draws a styled arrow on an image from (x1,y1) to (x2,y2)
func drawArrow(img *image.RGBA, x1, y1, x2, y2 int, c color.Color, thickness int) {
    // Draw line body
    dx := float64(x2 - x1)
    dy := float64(y2 - y1)
    length := math.Sqrt(dx*dx + dy*dy)
    if length == 0 {
        return
    }

    // Normalize direction
    ux := dx / length
    uy := dy / length

    // Draw thick line
    for t := 0.0; t < length-10; t += 1.0 {
        cx := int(float64(x1) + ux*t)
        cy := int(float64(y1) + uy*t)
        for d := -thickness/2; d <= thickness/2; d++ {
            px := cx + int(float64(d)*(-uy))
            py := cy + int(float64(d)*ux)
            if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
                img.Set(px, py, c)
            }
        }
    }

    // Draw arrowhead
    headSize := float64(thickness * 4)
    tipX := float64(x2)
    tipY := float64(y2)

    // Two points for the arrowhead
    leftX := int(tipX - headSize*ux + headSize*0.5*(-uy))
    leftY := int(tipY - headSize*uy + headSize*0.5*ux)
    rightX := int(tipX - headSize*ux - headSize*0.5*(-uy))
    rightY := int(tipY - headSize*uy - headSize*0.5*ux)

    // Fill triangle
    fillTriangle(img, x2, y2, leftX, leftY, rightX, rightY, c)
}

// fillTriangle fills a triangle given three vertices
func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, c color.Color) {
    // Simple scanline fill
    minY := min(y1, min(y2, y3))
    maxY := max(y1, max(y2, y3))

    for y := minY; y <= maxY; y++ {
        // Find intersections with edges
        var xs []int
        edges := [][4]int{{x1, y1, x2, y2}, {x2, y2, x3, y3}, {x3, y3, x1, y1}}
        for _, e := range edges {
            if (e[1] <= y && e[3] > y) || (e[3] <= y && e[1] > y) {
                t := float64(y-e[1]) / float64(e[3]-e[1])
                x := int(float64(e[0]) + t*float64(e[2]-e[0]))
                xs = append(xs, x)
            }
        }
        if len(xs) >= 2 {
            minX := min(xs[0], xs[1])
            maxX := max(xs[0], xs[1])
            for x := minX; x <= maxX; x++ {
                if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
                    img.Set(x, y, c)
                }
            }
        }
    }
}
```

**Acceptance Criteria:**
- Function draws directional arrows
- Arrowheads are visible and properly oriented

---

#### Task 3.2: Create Styled Component Box Widget
**File:** `module4-circuits/pkg/gui/helpers.go`

**New Code (add to helpers.go):**
```go
// ComponentBox represents a styled data path component
type ComponentBox struct {
    Title      string
    ValueLabel *widget.Label
    BgColor    color.Color
    container  *fyne.Container
}

// NewComponentBox creates a styled component box for data paths
func NewComponentBox(title string, initialValue string, bgColor color.Color) *ComponentBox {
    titleLbl := widget.NewLabelWithStyle(title, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
    valueLbl := widget.NewLabel(initialValue)
    valueLbl.Alignment = fyne.TextAlignCenter

    // Background with corner radius and subtle shadow effect
    bg := canvas.NewRectangle(bgColor)
    bg.SetMinSize(fyne.NewSize(110, 70))
    bg.CornerRadius = 8

    // Border for definition
    border := canvas.NewRectangle(color.RGBA{255, 255, 255, 40})
    border.SetMinSize(fyne.NewSize(110, 70))
    border.CornerRadius = 8
    border.StrokeWidth = 1
    border.StrokeColor = color.RGBA{255, 255, 255, 80}

    content := container.NewVBox(
        container.NewCenter(titleLbl),
        container.NewCenter(valueLbl),
    )

    box := &ComponentBox{
        Title:      title,
        ValueLabel: valueLbl,
        BgColor:    bgColor,
        container:  container.NewStack(bg, border, container.NewCenter(content)),
    }

    return box
}

// Container returns the fyne container for this component box
func (cb *ComponentBox) Container() *fyne.Container {
    return cb.container
}

// SetValue updates the displayed value
func (cb *ComponentBox) SetValue(value string) {
    cb.ValueLabel.SetText(value)
}
```

**Acceptance Criteria:**
- Component boxes have rounded corners
- Boxes have subtle border for definition
- Value labels are updateable

---

#### Task 3.3: Update WRITE Data Path
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 265-292 (createWriteDataPathSection function)

**Changes:**
- Replace plain text arrows with canvas-drawn styled arrows
- Use ComponentBox instead of createLabeledBoxWithLabel
- Add animation support (optional, stretch goal)

**New Implementation:**
```go
func (ca *CircuitsApp) createWriteDataPathSection() fyne.CanvasObject {
    // Create component boxes with shared theme colors
    ca.writeDigitalBox = NewComponentBox("DIGITAL", "Level:15\n01111", sharedtheme.ColorPrimary)
    ca.writeDACBox = NewComponentBox("DAC", "3.55V", sharedtheme.ColorAccent)
    ca.writeFeFETBox = NewComponentBox("FeFET", "[3,5]\n52.2uS", sharedtheme.ColorInfo)

    // Store label references for updates
    ca.writeDigitalLabel = ca.writeDigitalBox.ValueLabel
    ca.writeDACLabel = ca.writeDACBox.ValueLabel
    ca.writeFeFETLabel = ca.writeFeFETBox.ValueLabel

    // Create arrow canvas
    arrowCanvas := canvas.NewRaster(func(w, h int) image.Image {
        img := image.NewRGBA(image.Rect(0, 0, w, h))
        // Draw arrows between boxes
        arrowColor := sharedtheme.ColorWarning
        y := h / 2
        // Arrow 1: after first box
        drawArrow(img, 115, y, 135, y, arrowColor, 3)
        // Arrow 2: after second box
        drawArrow(img, 255, y, 275, y, arrowColor, 3)
        return img
    })
    arrowCanvas.SetMinSize(fyne.NewSize(400, 80))

    // Overlay boxes on arrow canvas
    ca.writeDataPath = container.NewStack(
        arrowCanvas,
        container.NewHBox(
            ca.writeDigitalBox.Container(),
            layout.NewSpacer(),
            ca.writeDACBox.Container(),
            layout.NewSpacer(),
            ca.writeFeFETBox.Container(),
        ),
    )

    ca.updateWriteDataPath()

    helperText := widget.NewLabel("Data path: Digital level -> DAC voltage conversion -> FeFET polarization")
    helperText.TextStyle = fyne.TextStyle{Italic: true}

    return container.NewVBox(ca.writeDataPath, helperText)
}
```

**Acceptance Criteria:**
- Arrows are canvas-drawn, not text
- Component boxes are styled
- Layout adapts to container width

---

#### Task 3.4: Update READ Data Path
**File:** `module4-circuits/pkg/gui/tab_read.go`
**Lines:** 202-242 (createReadDataPathSection function)

**Changes:** Same pattern as WRITE - use ComponentBox and canvas arrows.

**Acceptance Criteria:**
- READ data path matches WRITE styling
- 4 components connected by arrows

---

#### Task 3.5: Update COMPUTE Visualization
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 203-300 (createComputeVizSection and drawComputeViz functions)

**Changes:**
- Add visual separation between DAC, Array, and ADC sections
- Draw connecting lines/arrows
- Use shared theme colors

**Acceptance Criteria:**
- DAC -> Array -> ADC flow is visually clear
- Components have distinct colors from theme

---

### Phase 4: Interactive Array

#### Task 4.1: Add Hover State Tracking
**File:** `module4-circuits/pkg/gui/app.go`
**Lines:** Add after line 57 (in CircuitsApp struct)

**New Fields:**
```go
// Hover state for array interaction
hoveredRow int
hoveredCol int
isHovering bool
```

**Acceptance Criteria:**
- Fields added to struct
- Initialized to -1, -1, false

---

#### Task 4.2: Add Visual Hover Feedback
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 488-575 (drawWriteArray function)

**Changes:**
- Add hover highlight color (brighter version of cell color)
- Draw hover indicator when isHovering && row==hoveredRow && col==hoveredCol

**New code in draw loop:**
```go
isHovered := ca.isHovering && r == ca.hoveredRow && c == ca.hoveredCol

var cr, cg, cb uint8
if isSelected {
    cr, cg, cb = 255, 200, 50 // Bright yellow for selection
} else if isHovered {
    // Lighter version of normal color for hover
    cr = uint8(min(255, int(intensity*200)+50))
    cg = uint8(min(255, int(50+(1-intensity)*100)+50))
    cb = uint8(min(255, int((1-intensity)*200)+50))
} else {
    cr = uint8(intensity * 200)
    cg = uint8(50 + (1-intensity)*100)
    cb = uint8((1 - intensity) * 200)
}
```

**Acceptance Criteria:**
- Hovered cell is visually distinct
- Hover doesn't interfere with selection

---

#### Task 4.3: Add Click Feedback
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 482-486 (createWriteArraySection function)

**Changes:**
- Wrap raster in a container that handles taps
- Convert tap position to cell coordinates
- Update selection and show feedback

**New Implementation:**
```go
func (ca *CircuitsApp) createWriteArraySection() fyne.CanvasObject {
    ca.writeArrayCanvas = canvas.NewRaster(ca.drawWriteArray)
    ca.writeArrayCanvas.SetMinSize(fyne.NewSize(500, 350))

    // Create tappable container
    tappable := &ArrayTappable{
        canvas:  ca.writeArrayCanvas,
        onTap:   ca.onArrayCellTapped,
        onHover: ca.onArrayCellHover,
        ca:      ca,
    }

    return container.NewStack(tappable, ca.writeArrayCanvas)
}

// ArrayTappable handles tap and hover events on the array
type ArrayTappable struct {
    widget.BaseWidget
    canvas  *canvas.Raster
    onTap   func(row, col int)
    onHover func(row, col int, hovering bool)
    ca      *CircuitsApp
}

func (at *ArrayTappable) Tapped(e *fyne.PointEvent) {
    row, col := at.positionToCell(e.Position)
    if row >= 0 && col >= 0 {
        at.onTap(row, col)
    }
}

func (at *ArrayTappable) MouseIn(e *desktop.MouseEvent) {
    row, col := at.positionToCell(e.Position)
    at.onHover(row, col, true)
}

func (at *ArrayTappable) MouseMoved(e *desktop.MouseEvent) {
    row, col := at.positionToCell(e.Position)
    at.onHover(row, col, true)
}

func (at *ArrayTappable) MouseOut() {
    at.onHover(-1, -1, false)
}

func (at *ArrayTappable) positionToCell(pos fyne.Position) (int, int) {
    // Convert position to cell coordinates
    // Uses same logic as drawing
    size := at.canvas.Size()
    at.ca.mu.RLock()
    rows := at.ca.arrayRows
    cols := at.ca.arrayCols
    at.ca.mu.RUnlock()

    margin := float32(40)
    cellW := (size.Width - 2*margin) / float32(cols)
    cellH := (size.Height - 2*margin) / float32(rows)
    cellSize := cellW
    if cellH < cellSize {
        cellSize = cellH
    }
    if cellSize > 40 {
        cellSize = 40
    }

    gridW := float32(cols) * cellSize
    gridH := float32(rows) * cellSize
    offsetX := (size.Width - gridW) / 2
    offsetY := (size.Height - gridH) / 2

    col := int((pos.X - offsetX) / cellSize)
    row := int((pos.Y - offsetY) / cellSize)

    if row < 0 || row >= rows || col < 0 || col >= cols {
        return -1, -1
    }
    return row, col
}
```

**Acceptance Criteria:**
- Clicking cell selects it
- Hover shows visual feedback
- Mouse out clears hover state

---

### Phase 5: Chart Improvements

#### Task 5.1: Add ColorLegend to WRITE Array
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 482-486 (createWriteArraySection)

**Changes:**
```go
func (ca *CircuitsApp) createWriteArraySection() fyne.CanvasObject {
    ca.writeArrayCanvas = canvas.NewRaster(ca.drawWriteArray)
    ca.writeArrayCanvas.SetMinSize(fyne.NewSize(500, 350))

    // Add color legend
    legend := sharedwidgets.NewColorLegend(0, 29, "Level", true, func(t float64) color.RGBA {
        // Match the color mapping in drawWriteArray
        cr := uint8(t * 200)
        cg := uint8(50 + (1-t)*100)
        cb := uint8((1 - t) * 200)
        return color.RGBA{cr, cg, cb, 255}
    })

    return container.NewBorder(nil, nil, nil, legend, ca.writeArrayCanvas)
}
```

**Acceptance Criteria:**
- Legend shows low-to-high color mapping
- Labels show "Level" units
- Legend positioned to right of array

---

#### Task 5.2: Add ColorLegend to COMPUTE Array
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 203-207 (createComputeVizSection)

**Changes:** Same pattern as WRITE array.

**Acceptance Criteria:**
- COMPUTE array has matching legend
- Consistent with WRITE array styling

---

#### Task 5.3: Style Bar Charts in COMPARISON Tab
**File:** `module4-circuits/pkg/gui/tab_comparison.go`
**Lines:** 160-238 (drawCompTiming), 246-309 (drawCompEnergy)

**Changes:**
- Add axis titles
- Add grid lines
- Use rounded bar ends
- Add value labels on bars

**Acceptance Criteria:**
- Charts have clear axis labels
- Grid lines aid value reading
- Bars are visually distinct per technology

---

### Phase 6: Tooltips

#### Task 6.1: Add Tooltips to Configuration Inputs
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 100-202 (createWriteConfigSection)

**Changes:**
- Wrap key inputs with tooltip-enabled containers
- Add explanatory text for voltage ranges, pulse width, quantization

**Example:**
```go
// Create tooltip wrapper
vMinEntry := widget.NewEntry()
vMinEntry.SetText("2.0")
vMinTooltip := &widget.Button{
    Text: "?",
    OnTapped: func() {
        dialog.ShowInformation("Min Write Voltage",
            "The minimum voltage needed to switch ferroelectric polarization.\n"+
            "Must exceed the coercive field (Ec ~1.0-1.5 MV/cm).\n"+
            "Typical range: 2.0V - 3.0V",
            ca.window)
    },
}
vMinRow := container.NewHBox(vMinEntry, vMinTooltip, widget.NewLabel("V min"))
```

**Acceptance Criteria:**
- Key parameters have "?" help buttons
- Dialogs explain physics significance
- Non-intrusive to main workflow

---

#### Task 6.2: Add Tooltips to Data Path Components
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 265-292

**Changes:**
- Make component boxes tappable for info
- Show educational popup on tap

**Acceptance Criteria:**
- Tapping component shows explanation
- Explains role in data path

---

### Phase 7: Responsive Layout

#### Task 7.1: Wrap Tab Content with AdaptiveLayout
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 20-98 (createWriteTab)

**Changes:**
- Define zones for different layout regions
- Use AdaptiveLayout for responsive behavior
- Mobile layout: vertical scroll with sections
- Desktop layout: horizontal three-panel

**Example:**
```go
func (ca *CircuitsApp) createWriteTab() fyne.CanvasObject {
    // ... create sections as before ...

    // Define zones
    zones := []fyne.CanvasObject{
        leftPanel,
        centerPanel,
        rightPanel,
        arraySection,
    }
    tabLabels := []string{"Config", "Data Path", "Mapping", "Array"}

    adaptive := sharedwidgets.NewAdaptiveLayout(zones, tabLabels)
    adaptive.SetDesktopLayout(func(zones []fyne.CanvasObject) fyne.CanvasObject {
        // Desktop: three columns on top, array below
        topRow := container.NewHBox(
            container.NewPadded(zones[0]),
            widget.NewSeparator(),
            container.NewPadded(zones[1]),
            widget.NewSeparator(),
            container.NewPadded(zones[2]),
        )
        return container.NewBorder(
            container.NewVBox(headerLabel, widget.NewSeparator(), topRow),
            container.NewVBox(widget.NewSeparator(), buttonBox),
            nil, nil,
            zones[3],
        )
    })

    return adaptive.Content()
}
```

**Acceptance Criteria:**
- Desktop shows three-panel layout
- Mobile shows tabbed interface
- Transitions smoothly

---

#### Task 7.2: Define Breakpoint Behaviors
**File:** `module4-circuits/pkg/gui/app.go`

**Changes:**
- Add callback for breakpoint changes
- Adjust canvas sizes based on breakpoint

**Acceptance Criteria:**
- Canvases resize appropriately
- No content overflow
- Maintains usability at all sizes

---

### Phase 8: Status Feedback

#### Task 8.1: Create Prominent Status Display Area
**File:** `module4-circuits/pkg/gui/app.go`
**Lines:** 186-273 (createMainLayout)

**Changes:**
- Add status bar at top of content area
- Use colored background for status messages
- Auto-fade after timeout (optional)

**New Code:**
```go
// StatusDisplay shows prominent feedback messages
type StatusDisplay struct {
    container *fyne.Container
    label     *widget.Label
    bg        *canvas.Rectangle
}

func NewStatusDisplay() *StatusDisplay {
    label := widget.NewLabel("")
    label.Alignment = fyne.TextAlignCenter

    bg := canvas.NewRectangle(sharedtheme.ColorSurface)
    bg.SetMinSize(fyne.NewSize(0, 30))
    bg.CornerRadius = 4

    sd := &StatusDisplay{
        label: label,
        bg:    bg,
        container: container.NewStack(bg, container.NewCenter(label)),
    }
    sd.container.Hide()

    return sd
}

func (sd *StatusDisplay) Show(message string, statusType string) {
    switch statusType {
    case "success":
        sd.bg.FillColor = sharedtheme.ColorSuccess
    case "error":
        sd.bg.FillColor = sharedtheme.ColorError
    case "warning":
        sd.bg.FillColor = sharedtheme.ColorWarning
    default:
        sd.bg.FillColor = sharedtheme.ColorInfo
    }
    sd.label.SetText(message)
    sd.container.Show()
    sd.bg.Refresh()
}
```

**Acceptance Criteria:**
- Status messages visible at top
- Color-coded by type (success, error, info)
- Does not obscure content

---

#### Task 8.2: Update All Status Update Calls
**Files:** All tab files

**Changes:**
- Replace `ca.writeStatusLabel.SetText()` with `ca.statusDisplay.Show()`
- Add status type parameter

**Acceptance Criteria:**
- All status updates use new display
- Consistent feedback across tabs

---

## Commit Strategy

### Commit 1: Theme Migration
```
feat(module4): migrate to shared theme system

- Remove local theme.go file
- Import shared/theme in all gui files
- Update color references to use sharedtheme.*
- Maintains visual consistency with other modules
```

### Commit 2: Visual Hierarchy
```
feat(module4): add styled section headers

- Create section header helper with accent bar
- Apply consistent headers to all 6 tabs
- Improves scanability and navigation
```

### Commit 3: Data Flow Visualization
```
feat(module4): enhance data path visualization

- Add canvas-drawn arrows with arrowheads
- Create styled component box widget
- Update WRITE, READ, COMPUTE data paths
- Better illustrates signal flow
```

### Commit 4: Interactive Array
```
feat(module4): add array interaction feedback

- Add hover state tracking
- Visual hover highlight on cells
- Click-to-select with feedback
- Improves discoverability
```

### Commit 5: Chart Improvements
```
feat(module4): add chart legends and styling

- Add ColorLegend to array visualizations
- Improve bar chart styling in COMPARISON
- Add axis labels and grid lines
```

### Commit 6: Tooltips
```
feat(module4): add educational tooltips

- Add help buttons to configuration inputs
- Component boxes show explanations on tap
- Enhances learning experience
```

### Commit 7: Responsive Layout
```
feat(module4): implement responsive layout

- Wrap tabs with AdaptiveLayout
- Define desktop and mobile layouts
- Smooth breakpoint transitions
```

### Commit 8: Status Feedback
```
feat(module4): improve status feedback display

- Add prominent status bar at top
- Color-coded status messages
- Replace bottom-corner labels
```

---

## Success Criteria

### Functional
- [ ] All 6 tabs render correctly
- [ ] All interactive elements work (buttons, sliders, selects)
- [ ] Array click-to-select functions properly
- [ ] Calculations and simulations produce correct results
- [ ] No regressions from current functionality

### Visual
- [ ] Theme colors match `shared/theme/theme.go`
- [ ] Section headers visually distinct
- [ ] Data paths use styled arrows and boxes
- [ ] Array has hover feedback
- [ ] Charts have legends
- [ ] Status messages are prominent

### Code Quality
- [ ] No compilation errors
- [ ] All tests pass
- [ ] No new linter warnings
- [ ] Thread-safe UI updates
- [ ] Follows Go idioms and project conventions

### Performance
- [ ] No noticeable lag on UI interactions
- [ ] Canvas renders complete within 16ms (60fps)
- [ ] No memory leaks on tab switching

---

## Risk Assessment

### High Risk
| Risk | Mitigation |
|------|------------|
| Canvas drawing performance | Profile with larger arrays, optimize hot paths |
| Thread safety on hover updates | Use mutex, batch updates with fyne.Do() |
| Breaking EmbeddedCircuitsApp interface | Verify interface compliance after each phase |

### Medium Risk
| Risk | Mitigation |
|------|------------|
| Responsive layout edge cases | Test at 320px, 768px, 1400px widths |
| Color contrast accessibility | Verify WCAG AA ratios for text on colors |
| Tooltip dialog UX | Test with keyboard users |

### Low Risk
| Risk | Mitigation |
|------|------------|
| Theme import paths | Verify module path in go.mod |
| Arrow drawing math errors | Unit test with known coordinates |

---

## Verification Steps

1. **Build Verification**
   ```bash
   go build ./cmd/fecim-visualizer
   ```

2. **Test Verification**
   ```bash
   go test ./module4-circuits/...
   ```

3. **Visual Verification**
   - Launch app, navigate to Module 4
   - Check each tab for visual correctness
   - Verify all colors match shared theme
   - Test array hover and click
   - Check status feedback visibility

4. **Responsive Verification**
   - Resize window to mobile width (<768px)
   - Verify layout adapts
   - Resize back to desktop
   - Verify no layout glitches

5. **Functionality Verification**
   - WRITE: Program a cell, verify array updates
   - READ: Read a cell, verify calculation display
   - COMPUTE: Run MVM, verify output
   - COMPARISON: Run comparison, verify charts
   - TIMING: Check all timing diagrams
   - SPECS: Verify all specs display correctly
