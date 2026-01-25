# Module 6 EDA Learn Tab UI Fixes

## Context

### Original Request
Fix Module 6 EDA Learn tab UI issues - overlapping text, cramped diagrams, and layout problems across all three tabs.

### Research Findings
Based on screenshot analysis and code review:

**Tab 1 - "What is FeCIM EDA?":**
- `OperationModesVisual()` has fixed coordinates that cause text overflow
- Box width (140px) too narrow for description text like "Flash-like, ~4.9 bits/cell"
- Description positioning uses hardcoded offset (descX = mode.x + 8) without accounting for text width

**Tab 2 - "The Crossbar Architecture":**
- `IsometricCrossbar()` container size (540x400) insufficient for diagram with labels
- `Isometric1T1RCrossbar()` same sizing issues
- `CellComparisonTable()` column widths (100px each) cause text truncation for values like "0.46x2.72um"

**Tab 3 - "EDA Files We Generate":**
- `FileFormatCard()` uses HBox layout causing horizontal cramping
- Cards forced side-by-side when they need vertical stacking for narrow viewports

### Root Causes Identified

| Component | File:Line | Issue |
|-----------|-----------|-------|
| `OperationModesVisual()` | learn_visuals.go:640-750 | Fixed 140px box width, hardcoded text positions |
| `IsometricCrossbar()` | learn_visuals.go:222-357 | 540x400 container too small, legend overlaps diagram |
| `Isometric1T1RCrossbar()` | learn_visuals.go:360-520 | Same sizing issues as passive crossbar |
| `CellComparisonTable()` | learn_visuals.go:527-633 | Column widths hardcoded at 100px each |
| `FileFormatCard()` | learn_visuals.go:757-787 | Card fixed at 360x224, no responsive layout |
| `makeFilesContent()` | learn_tab.go:255-309 | Uses HBox for cards, needs vertical stacking |

---

## Work Objectives

### Core Objective
Fix all visual rendering issues in the Module 6 Learn tab to ensure readable text, properly spaced diagrams, and responsive layouts.

### Deliverables
1. **FYNE_NOTES.md** - Development documentation at `<local-path>`
2. **Fixed learn_visuals.go** - All visual components with proper sizing and spacing
3. **Fixed learn_tab.go** - Layouts using proper Fyne containers

### Definition of Done
- [ ] All text in diagrams is fully visible without truncation
- [ ] No overlapping elements in any diagram
- [ ] Diagrams have adequate padding around edges
- [ ] File cards stack vertically for better readability
- [ ] App builds without errors: `go build -o fecim-visualizer ./cmd/fecim-visualizer`
- [ ] Visual verification confirms all issues resolved

---

## Guardrails

### Must Have
- Use Fyne best practices (VScroll, AdaptiveGrid, Padded containers)
- Maintain existing visual style (colors, fonts, overall aesthetic)
- Preserve all existing content and functionality
- Ensure all changes are backwards compatible

### Must NOT Have
- Breaking changes to exported function signatures
- Removal of any existing visual elements
- Changes to files outside module6-eda/pkg/gui/tabs/
- Hard dependencies on specific window sizes

---

## Task Flow

```
Task 1 (FYNE_NOTES.md)
       |
       v
Task 2 (OperationModesVisual) ---> Task 3 (OpenLaneFlowDiagram)
       |                                    |
       v                                    v
Task 4 (IsometricCrossbar) ----------> Task 5 (Isometric1T1RCrossbar)
       |                                    |
       v                                    v
Task 6 (CellComparisonTable) --------> Task 7 (FileFormatCard)
       |                                    |
       v                                    v
Task 8 (learn_tab.go layouts) -------------|
       |
       v
Task 9 (Build & Visual Verification)
```

---

## Detailed TODOs

### Task 1: Create FYNE_NOTES.md Documentation
**File:** `<local-path>`
**Acceptance Criteria:**
- Document covers Fyne layout best practices
- Includes examples for VScroll, AdaptiveGrid, Padded
- Documents MinSize requirements for canvas-based drawings
- Includes text wrapping patterns

**Implementation:**
```markdown
# Fyne Development Notes

## Layout Best Practices

### Scroll Containers
- Use `container.NewVScroll()` for any content that may exceed viewport
- Set MinSize on scroll container, not child content

### Responsive Grids
- Use `container.NewAdaptiveGrid(columns, objects...)` for card layouts
- Adapts automatically to available width

### Canvas-Based Drawings
- Always call `container.Resize()` with explicit dimensions
- Account for labels/legends in size calculations
- Add 20-40px padding around diagram content

### Text Handling
- Always set `label.Wrapping = fyne.TextWrapWord` for long text
- For canvas.NewText, calculate text width before positioning
- Use `len(text) * averageCharWidth` for approximate centering
```

---

### Task 2: Fix OperationModesVisual()
**File:** `<local-path>`
**Lines:** 640-750
**Acceptance Criteria:**
- Box width increased to fit description text (180px minimum)
- Description text centered properly within boxes
- Container size increased to accommodate larger boxes
- No text overflow or truncation

**Changes Required:**
1. Line 651: Change `boxW := float32(140)` to `boxW := float32(180)`
2. Line 652: Change `boxH := float32(110)` to `boxH := float32(120)`
3. Line 653: Change `spacing := float32(15)` to `spacing := float32(20)`
4. Line 683: Fix centering calculation - use `nameX := mode.x + (boxW - textWidth) / 2` with proper text width estimation
5. Line 691: Change `descX := mode.x + 8` to `descX := mode.x + (boxW - float32(len(mode.description)*6)) / 2`
6. Line 697: Update `circleX` to center under new layout: `circleX := float32(300)`
7. Line 747: Change container size from `fyne.NewSize(560, 260)` to `fyne.NewSize(620, 280)`

---

### Task 3: Fix OpenLaneFlowDiagram()
**File:** `<local-path>`
**Lines:** 41-173
**Acceptance Criteria:**
- All stage boxes visible without cramping
- Arrows properly spaced between boxes
- Labels ("Our LEF", "Our DEF") visible above boxes
- Container size accommodates all elements

**Changes Required:**
1. Line 43: Increase `boxW := float32(140)` to `boxW := float32(150)`
2. Line 46: Increase `spacing := float32(25)` to `spacing := float32(30)`
3. Line 170: Change container size from `fyne.NewSize(720, 300)` to `fyne.NewSize(780, 320)`

---

### Task 4: Fix IsometricCrossbar()
**File:** `<local-path>`
**Lines:** 222-357
**Acceptance Criteria:**
- Diagram content does not overlap legend
- Labels (WL0, WL1, BL0, BL1) fully visible
- Legend positioned below diagram with adequate spacing
- Container large enough for all elements

**Changes Required:**
1. Line 232: Move `startX := float32(220)` to `startX := float32(240)` for more left padding
2. Line 320: Move legend further down: `legendY := float32(320)` (was 280)
3. Line 354: Increase container size: `fyne.NewSize(600, 420)` (was 540x400)

---

### Task 5: Fix Isometric1T1RCrossbar()
**File:** `<local-path>`
**Lines:** 360-520
**Acceptance Criteria:**
- Diagram matches passive crossbar sizing improvements
- SL labels visible without overlap
- Legend properly positioned below diagram

**Changes Required:**
1. Line 369: Change `startX := float32(220)` to `startX := float32(240)`
2. Line 483: Change `legendY := float32(280)` to `legendY := float32(320)`
3. Line 517: Change container size to `fyne.NewSize(600, 420)` (was 540x400)

---

### Task 6: Fix CellComparisonTable()
**File:** `<local-path>`
**Lines:** 527-633
**Acceptance Criteria:**
- All cell text visible without truncation
- Column widths accommodate longest content
- Table properly aligned

**Changes Required:**
1. Line 531: Change column widths from `[]float32{100, 100, 100, 100}` to `[]float32{110, 120, 120, 90}` (total 440px)
2. Line 571: Update row background width from 400 to 440
3. Line 617: Update horizontal line end from 400 to 440
4. Line 625: Update border size from 400 to 440
5. Line 630: Update container size from `fyne.NewSize(420, 200)` to `fyne.NewSize(460, 210)`

---

### Task 7: Fix FileFormatCard()
**File:** `<local-path>`
**Lines:** 757-787
**Acceptance Criteria:**
- Card dimensions allow full code visibility
- Content area has adequate padding
- Cards can stack vertically in parent layout

**Changes Required:**
1. Line 759: Reduce header width for better fit: keep at 360 (or reduce to 340 if needed)
2. This component is OK as-is; the issue is in how cards are laid out in learn_tab.go

---

### Task 8: Fix learn_tab.go Layouts
**File:** `<local-path>`
**Lines:** Multiple sections

**8a. Fix makeFilesContent() card layout (lines 255-309)**
**Acceptance Criteria:**
- Cards stack in 2x2 grid that adapts to width
- Adequate spacing between cards
- All cards fully visible

**Changes Required:**
Replace HBox layout with AdaptiveGrid:
```go
// Replace lines 269-275:
// OLD:
// cardsRow1 := container.NewHBox(lefCard, spacerH1, defCard)
// spacerV := widget.NewLabel("")
// spacerV.Resize(fyne.NewSize(1, 12))
// cardsRow2 := container.NewHBox(verilogCard, spacerH2, libertyCard)

// NEW:
cardsGrid := container.NewAdaptiveGrid(2, lefCard, defCard, verilogCard, libertyCard)
```

Then update the return VBox to use cardsGrid instead of cardsRow1/spacerV/cardsRow2.

**8b. Fix makeCrossbarContent() diagram container sizes (lines 181-253)**
**Acceptance Criteria:**
- Diagram containers sized to match updated visual component sizes
- Adequate vertical spacing between sections

**Changes Required:**
1. Line 194: Update passiveDiagramContainer resize: `fyne.NewSize(620, 440)` (was 560x420)
2. Line 205: Update oneToneRDiagramContainer resize: `fyne.NewSize(620, 440)` (was 560x420)
3. Line 210: Update comparisonContainer resize: `fyne.NewSize(480, 230)` (was 440x220)

**8c. Fix makeIntroContent() visual container sizes (lines 106-178)**
**Acceptance Criteria:**
- Visual containers sized to match updated component sizes

**Changes Required:**
1. Line 115: Update modesContainer resize: `fyne.NewSize(640, 300)` (was 600x280)
2. Line 120: Update flowContainer resize: `fyne.NewSize(800, 340)` (was 750x320)

---

### Task 9: Build and Visual Verification
**Acceptance Criteria:**
- `go build -o fecim-visualizer ./cmd/fecim-visualizer` succeeds
- Launch app and verify each Learn tab visually
- All text readable without truncation
- No overlapping elements
- Diagrams properly spaced

**Verification Steps:**
1. Build: `go build -o fecim-visualizer ./cmd/fecim-visualizer`
2. Launch: `./fecim-visualizer`
3. Navigate to Module 6 (EDA)
4. Check "Learn" tab
5. Verify Tab 1: "What is FeCIM EDA?" - operation modes diagram, OpenLane flow
6. Verify Tab 2: "The Crossbar Architecture" - both crossbar diagrams, comparison table
7. Verify Tab 3: "EDA Files We Generate" - all four file cards visible

---

## Commit Strategy

### Single Commit
```
fix: Module 6 Learn tab UI - spacing, sizing, and layout improvements

- Increase OperationModesVisual box widths for text fit
- Enlarge crossbar diagrams and reposition legends
- Widen comparison table columns
- Convert file cards to adaptive grid layout
- Add FYNE_NOTES.md development documentation

Resolves text truncation, overlapping elements, and cramped layouts.
```

---

## Success Criteria

| Criterion | Verification Method |
|-----------|---------------------|
| Build passes | `go build` exits 0 |
| Tab 1 renders correctly | Visual inspection - no text overflow in operation modes |
| Tab 2 renders correctly | Visual inspection - crossbar legends don't overlap diagrams |
| Tab 3 renders correctly | Visual inspection - all file cards visible in grid |
| No regressions | All three tabs still display all content |
