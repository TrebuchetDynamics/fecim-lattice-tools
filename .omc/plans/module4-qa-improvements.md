# Module 4 QA Improvements Work Plan

**Created:** 2026-01-25
**Objective:** Systematically test, evaluate, and improve Module 4 (Peripheral Circuits) to achieve gold-standard quality for university-level education and research.
**Focus Areas:** UI functionality, physics accuracy, educational value

---

## Context

### Original Request
Systematically test, evaluate, and improve Module 4 (Peripheral Circuits) of the FeCIM Lattice Tools application to achieve gold-standard quality for university-level education and research. Focus on UI and physics accuracy. All sliders and buttons must work correctly.

### Codebase Structure
- **Main app:** `module4-circuits/pkg/gui/app.go` (274 lines)
- **Tab files:**
  - `tab_write.go` (688 lines) - WRITE mode
  - `tab_read.go` (465 lines) - READ mode
  - `tab_compute.go` (438 lines) - COMPUTE mode
  - `tab_comparison.go` (370 lines) - COMPARISON tab
  - `tab_timing.go` (516 lines) - TIMING diagrams
  - `tab_specs.go` (277 lines) - SPECIFICATIONS
- **Support files:** `font.go`, `helpers.go`, `drawing.go`, `theme.go`

---

## Work Objectives

### Core Objective
Fix all identified bugs, implement missing button handlers, and improve physics accuracy to make Module 4 a reliable educational tool.

### Deliverables
1. All buttons functional (no nil handlers)
2. All sliders update their dependent UI elements
3. Physics calculations use dynamic values (not hardcoded)
4. Energy efficiency claims include appropriate disclaimers
5. Educational explanations added for key physics concepts

### Definition of Done
- [ ] All 6 tabs have fully functional controls
- [ ] No hardcoded "magic numbers" in physics calculations
- [ ] Energy claims follow CLAUDE.md accuracy policy
- [ ] All tests pass (`go test ./...`)
- [ ] App builds and runs without errors

---

## Must Have

1. **All buttons must have functional handlers** (no nil handlers)
2. **All sliders must update dependent UI elements**
3. **Physics calculations must use dynamic values** (quantLevels, not "29")
4. **Row/Col dropdowns must update when array size changes**
5. **Energy efficiency claims must have disclaimers**

## Must NOT Have

1. **Do NOT change the overall 6-tab structure**
2. **Do NOT modify the color scheme or theme**
3. **Do NOT add new dependencies**
4. **Do NOT change the peripheral circuit physics model**
5. **Do NOT remove any existing functionality**

---

## Risk Identification

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Breaking existing working features | HIGH | MEDIUM | Test each fix in isolation before committing |
| Inconsistent state after array resize | MEDIUM | HIGH | Add mutex locks around state updates |
| UI thread blocking from goroutines | HIGH | LOW | Use fyne.Do() for all UI updates |
| Physics calculation errors | HIGH | MEDIUM | Cross-check formulas against CLAUDE.md references |

---

## Task Flow and Dependencies

```
Phase 1: Critical Button Fixes (P0)
    |
    +-> Task 1.1: COMPARISON Tab buttons (nil handlers)
    +-> Task 1.2: TIMING Tab buttons (nil handlers)
    +-> Task 1.3: SPECS Tab buttons (nil handlers)
    |
Phase 2: Slider/Config Reactivity (P1)
    |
    +-> Task 2.1: WRITE Tab - array size -> Row/Col dropdowns
    +-> Task 2.2: WRITE Tab - Pulse Width entry -> pulse canvas
    +-> Task 2.3: READ Tab - TIA Gain/ADC Resolution -> calculation display
    +-> Task 2.4: COMPUTE Tab - DAC/ADC Bits handlers
    |
Phase 3: Physics Accuracy (P1)
    |
    +-> Task 3.1: READ Tab - hardcoded "29" -> (levels-1)
    +-> Task 3.2: COMPUTE Tab - math breakdown shows more terms
    +-> Task 3.3: TIMING Tab - consistent timing values
    +-> Task 3.4: SPECS Tab - dynamic System Summary
    |
Phase 4: Educational Enhancements (P2)
    |
    +-> Task 4.1: COMPARISON Tab - energy efficiency disclaimer
    +-> Task 4.2: WRITE Tab - coercive field (Ec) explanation
    +-> Task 4.3: READ Tab - safe voltage explanation
    +-> Task 4.4: COMPUTE Tab - Kirchhoff's law explanation
    |
Phase 5: UI Polish (P2)
    |
    +-> Task 5.1: WRITE Tab - proper table for Level-to-Voltage mapping
    +-> Task 5.2: COMPUTE Tab - show all columns in input (not just first 8)
    +-> Task 5.3: READ Tab - full implementation of READ ALL/VERIFY buttons
```

---

## Detailed TODOs

### Phase 1: Critical Button Fixes (P0)

#### Task 1.1: COMPARISON Tab - Fix nil button handlers
**File:** `module4-circuits/pkg/gui/tab_comparison.go`
**Lines:** 39-40

**Current Code:**
```go
animateBtn := widget.NewButton("ANIMATE", nil)
scaleBtn := widget.NewButton("SCALE UP", nil)
```

**Fix:** Implement `onAnimateComparison()` and `onScaleUpComparison()` methods

**Acceptance Criteria:**
- [ ] ANIMATE button triggers step-by-step animation showing CPU vs GPU vs FeFET timing
- [ ] SCALE UP button allows changing array size (8x8 -> 16x16 -> 32x32) and updates comparison

**Verification:** Click both buttons; no crashes, status label updates

---

#### Task 1.2: TIMING Tab - Fix nil button handlers
**File:** `module4-circuits/pkg/gui/tab_timing.go`
**Lines:** 37-38

**Current Code:**
```go
animateBtn := widget.NewButton("ANIMATE", nil)
exportBtn := widget.NewButton("EXPORT SVG", nil)
```

**Fix:** Implement `onAnimateTiming()` and `onExportTimingSVG()` methods

**Acceptance Criteria:**
- [ ] ANIMATE button highlights each signal phase in sequence
- [ ] EXPORT SVG saves current timing diagram to file (or shows "not implemented" message)

**Verification:** Click both buttons; no crashes, appropriate feedback shown

---

#### Task 1.3: SPECS Tab - Fix nil button handlers
**File:** `module4-circuits/pkg/gui/tab_specs.go`
**Lines:** 41-42

**Current Code:**
```go
exportBtn := widget.NewButton("EXPORT SPECS", nil)
compareBtn := widget.NewButton("COMPARE TO GPU", nil)
```

**Fix:** Implement `onExportSpecs()` and `onCompareToGPU()` methods

**Acceptance Criteria:**
- [ ] EXPORT SPECS saves specifications to JSON/text file (or shows "not implemented")
- [ ] COMPARE TO GPU switches to COMPARISON tab with pre-filled values

**Verification:** Click both buttons; no crashes, appropriate feedback shown

---

### Phase 2: Slider/Config Reactivity (P1)

#### Task 2.1: WRITE Tab - Array size should update Row/Col dropdowns
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 103-123 (rowSelect and colSelect handlers)

**Current Issue:** When array size changes via `rowSelect`/`colSelect`, the Row/Col selection dropdowns in `createWriteCellSection()` are not updated to reflect new valid range.

**Fix:** Add `refreshCellSelectOptions()` method called from array size change handlers

**Implementation:**
```go
func (ca *CircuitsApp) refreshCellSelectOptions() {
    // Update writeRowSelect options to 0..(arrayRows-1)
    // Update writeColSelect options to 0..(arrayCols-1)
    // Reset selection if out of bounds
}
```

**Acceptance Criteria:**
- [ ] Changing array size to 4x4 updates Row dropdown to show 0-3
- [ ] Changing array size to 64x64 updates Row dropdown to show 0-63
- [ ] If selected row/col is now out of bounds, reset to valid value

**Verification:** Change array size, verify dropdown options update

---

#### Task 2.2: WRITE Tab - Pulse Width entry should update pulse canvas
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 162-171 (pulseEntry.OnChanged handler)

**Current Code:**
```go
pulseEntry.OnChanged = func(s string) {
    var pw float64
    fmt.Sscanf(s, "%f", &pw)
    ca.mu.Lock()
    ca.pulseWidth = pw
    ca.mu.Unlock()
    // MISSING: ca.refreshWritePulse()
}
```

**Fix:** Add `ca.refreshWritePulse()` call to update the pulse visualization

**Acceptance Criteria:**
- [ ] Changing pulse width from 50 to 100 ns updates the pulse visualization
- [ ] The pulse duration label/display reflects the new value

**Verification:** Enter different pulse widths, verify canvas updates

---

#### Task 2.3: READ Tab - TIA Gain and ADC Resolution should update calculation
**File:** `module4-circuits/pkg/gui/tab_read.go`
**Lines:** 117-137 (adcSelect and tiaSelect handlers)

**Current Issue:** Changing TIA Gain or ADC Resolution does not refresh the calculation display (`readCalcLabel`).

**Fix:** Add refresh calls to both select handlers

**Implementation:**
```go
adcSelect := widget.NewSelect(adcOptions, func(s string) {
    var bits int
    fmt.Sscanf(s, "%d", &bits)
    ca.mu.Lock()
    ca.adcBits = bits
    ca.mu.Unlock()
    ca.refreshReadCalculation() // ADD THIS
})
```

**Acceptance Criteria:**
- [ ] Changing TIA Gain updates the V_tia calculation in real-time
- [ ] Changing ADC Resolution updates the ADC formula display

**Verification:** Change TIA Gain from 10 to 100, verify calculation updates

---

#### Task 2.4: COMPUTE Tab - DAC/ADC Bits handlers
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 100-116 (dacSelect and adcSelect are created but have nil handlers)

**Current Code:**
```go
dacSelect := widget.NewSelect(dacBitsOptions, nil)
adcSelect := widget.NewSelect(adcBitsOptions, nil)
```

**Fix:** Add handlers to store the DAC/ADC bits values and optionally affect compute results

**Acceptance Criteria:**
- [ ] Changing DAC bits updates internal state
- [ ] Changing ADC bits updates internal state
- [ ] Values are used in compute calculations (if applicable)

**Verification:** Change DAC bits, verify state is stored

---

### Phase 3: Physics Accuracy (P1)

#### Task 3.1: READ Tab - Replace hardcoded "29" with (levels-1)
**File:** `module4-circuits/pkg/gui/tab_read.go`
**Line:** 445

**Current Code:**
```go
"Level = ADC/Max   = %d / 255 × 29  = %d",
```

**Fix:** Use `ca.quantLevels - 1` instead of hardcoded 29

**Corrected Code:**
```go
"Level = ADC/Max   = %d / 255 × %d  = %d",
adcRaw, ca.quantLevels-1, decodedLevel,
```

**Acceptance Criteria:**
- [ ] Changing quantization levels changes the formula display
- [ ] Formula is mathematically consistent with actual calculation

**Verification:** Set quantization to 32 levels, verify formula shows "× 31"

---

#### Task 3.2: COMPUTE Tab - Math breakdown should show more terms
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 374 (only shows 4 terms)

**Current Code:**
```go
cols := min(4, len(ca.arrayWeights[0]))
```

**Fix:** Show at least 6 terms (with ellipsis if more), or dynamically based on array size

**Implementation:**
```go
cols := min(6, len(ca.arrayWeights[0]))
```

**Acceptance Criteria:**
- [ ] Math breakdown shows at least 6 terms for 8x8 array
- [ ] Ellipsis ("+ ...") indicates there are more terms
- [ ] Formula is readable and educational

**Verification:** Run compute on 8x8 array, verify 6 terms shown

---

#### Task 3.3: TIMING Tab - Consistent timing values
**File:** `module4-circuits/pkg/gui/tab_timing.go`

**Current Issues:**
- Write timing shows "70ns total" (line 192)
- Compute timing phases: DAC 5ns + ARRAY 5ns + ADC 10ns = 20ns, but signals extend past this
- Phase labels don't match actual signal durations

**Fix:** Ensure timing values are consistent:
- Write: 70ns total (correct)
- Read: 20ns total (correct)
- Compute: 20ns total with phases that add up correctly

**Acceptance Criteria:**
- [ ] All timing diagrams have consistent total durations
- [ ] Phase durations match the actual signal patterns
- [ ] Labels are accurate

**Verification:** Visual inspection of all three timing diagrams

---

#### Task 3.4: SPECS Tab - Dynamic System Summary
**File:** `module4-circuits/pkg/gui/tab_specs.go`
**Lines:** 261-276 (createSpecSummarySection)

**Current Issue:** Summary shows hardcoded values (1,024 cells, 32 DACs) regardless of configuration.

**Fix:** Calculate values from `ca.specArraySizeSelect` and update on change

**Implementation:**
- Add handler to `specArraySizeSelect.OnChanged`
- Calculate: cells = rows × cols
- Update: MACs = cells, throughput = cells / 20ns

**Acceptance Criteria:**
- [ ] Changing array size to 64x64 updates summary to show 4,096 cells
- [ ] Throughput calculation updates accordingly
- [ ] Efficiency calculation updates accordingly

**Verification:** Change array size, verify summary updates

---

### Phase 4: Educational Enhancements (P2)

#### Task 4.1: COMPARISON Tab - Energy efficiency disclaimer
**File:** `module4-circuits/pkg/gui/tab_comparison.go`
**Lines:** 293-294 (energy savings annotation)

**Current Code:**
```go
drawSimpleText(img, "20000x savings!", w-120, fefetY+8, color.RGBA{0, 255, 200, 255})
```

**Fix:** Add footnote per CLAUDE.md accuracy policy (unverified claim)

**Implementation:**
- Add small footnote text: "*Simulated estimate; see CLAUDE.md for peer-reviewed comparisons"
- Or change to more conservative claim: "~10-100x savings (demonstrated in literature)"

**Acceptance Criteria:**
- [ ] Energy efficiency claim includes disclaimer
- [ ] Disclaimer is visible but not distracting
- [ ] Complies with CLAUDE.md "Accuracy & Honesty Policy"

**Verification:** Visual inspection of COMPARISON tab

---

#### Task 4.2: WRITE Tab - Coercive field (Ec) explanation
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 140-148 (voltage range entries)

**Current Help Text:** "Minimum write voltage (V) - must exceed coercive field"

**Enhancement:** Add tooltip or inline explanation

**Implementation:**
Add helper text explaining:
- Ec (Coercive Field): ~1.0-1.5 MV/cm for HZO
- Voltage must exceed Ec to switch polarization
- Below Ec = non-destructive read; Above Ec = write

**Acceptance Criteria:**
- [ ] User understands why 2.0V is the minimum write voltage
- [ ] Connection to coercive field concept is clear

**Verification:** Read the explanation, verify clarity

---

#### Task 4.3: READ Tab - Safe voltage explanation
**File:** `module4-circuits/pkg/gui/tab_read.go`
**Lines:** 110-114 (warning labels)

**Current Code:**
```go
warningLabel := widget.NewLabel("SAFE ZONE: 0.1V - 1.0V")
dangerLabel := widget.NewLabel("DANGER: > 2.0V (will modify cell!)")
```

**Enhancement:** Explain WHY 0.5V is safe

**Implementation:**
Add: "Read voltage must be below coercive field (~1.5V) to avoid disturbing the stored polarization state. Using 0.5V provides margin."

**Acceptance Criteria:**
- [ ] User understands the physics of non-destructive read
- [ ] Connection to write threshold is clear

**Verification:** Read the explanation, verify clarity

---

#### Task 4.4: COMPUTE Tab - Kirchhoff's Current Law explanation
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 21-24 (header description)

**Current Description:** "...summed as currents in each row via Kirchhoff's law..."

**Enhancement:** Make KCL explanation more explicit

**Implementation:**
Update header or add helper text:
"Kirchhoff's Current Law: Currents from all columns in a row sum at the row line. This analog summation performs the dot product in a single physical operation."

**Acceptance Criteria:**
- [ ] KCL concept is clearly explained
- [ ] Connection to dot product computation is clear

**Verification:** Read the explanation, verify clarity

---

### Phase 5: UI Polish (P2)

#### Task 5.1: WRITE Tab - Proper table for Level-to-Voltage mapping
**File:** `module4-circuits/pkg/gui/tab_write.go`
**Lines:** 585-659 (createWriteMappingSection and getMappingText)

**Current Issue:** Uses plain text label for table-like data

**Options:**
1. Use `widget.Table` (Fyne widget)
2. Use formatted grid with `container.NewGridWithColumns`
3. Keep text but improve formatting

**Recommended Fix:** Use `container.NewGridWithColumns(4)` with proper alignment

**Acceptance Criteria:**
- [ ] Table columns are aligned
- [ ] Current target level is highlighted
- [ ] Table updates when target level changes

**Verification:** Visual inspection of mapping table

---

#### Task 5.2: COMPUTE Tab - Show all columns in input (not just first 8)
**File:** `module4-circuits/pkg/gui/tab_compute.go`
**Lines:** 157 (hardcoded 8 columns)

**Current Code:**
```go
for i := 0; i < min(8, ca.arrayCols); i++ {
```

**Issue:** Only shows first 8 input columns regardless of array size

**Fix:** Use scrollable container or show inputs based on array size (up to reasonable limit)

**Implementation:**
- For arrays <= 16 columns: show all
- For arrays > 16 columns: show first 8 with "... +N more" indicator
- Or use horizontal scroll

**Acceptance Criteria:**
- [ ] 4x4 array shows 4 input columns
- [ ] 16x16 array shows up to 16 input columns (or scrollable)
- [ ] 32x32 array shows reasonable subset with indicator

**Verification:** Change array size, verify input columns adapt

---

#### Task 5.3: READ Tab - Full implementation of READ ALL/VERIFY buttons
**File:** `module4-circuits/pkg/gui/tab_read.go`
**Lines:** 453-464 (stub implementations)

**Current Code:**
```go
func (ca *CircuitsApp) onReadAllCells() {
    ca.readStatusLabel.SetText("Reading all cells...")
    // In a real implementation, this would iterate through all cells
    ca.readStatusLabel.SetText(fmt.Sprintf("Read all %d cells", ca.arrayRows*ca.arrayCols))
}

func (ca *CircuitsApp) onVerifyArray() {
    ca.readStatusLabel.SetText("Verifying array...")
    // Simplified verification
    errors := 0
    ca.readStatusLabel.SetText(fmt.Sprintf("Verification complete: %d errors", errors))
}
```

**Fix:** Implement actual read-all and verification logic

**Implementation:**
- `onReadAllCells()`: Actually read each cell and update a summary
- `onVerifyArray()`: Compare read levels with programmed levels, count mismatches

**Acceptance Criteria:**
- [ ] READ ALL CELLS iterates through array and shows aggregate results
- [ ] VERIFY ARRAY compares read vs. written values
- [ ] Mismatches are counted and reported

**Verification:** Click READ ALL, verify it processes all cells

---

## Commit Strategy

### Commit 1: Critical Button Fixes
```
fix(module4): Implement nil button handlers in COMPARISON, TIMING, SPECS tabs

- Add onAnimateComparison() and onScaleUpComparison()
- Add onAnimateTiming() and onExportTimingSVG()
- Add onExportSpecs() and onCompareToGPU()

Fixes buttons that previously did nothing when clicked.
```

### Commit 2: Slider/Config Reactivity
```
fix(module4): Make UI controls update their dependent elements

- WRITE: Array size updates Row/Col dropdowns
- WRITE: Pulse Width entry updates pulse canvas
- READ: TIA Gain/ADC Resolution update calculation display
- COMPUTE: DAC/ADC Bits selects have handlers
```

### Commit 3: Physics Accuracy
```
fix(module4): Use dynamic values in physics calculations

- READ: Replace hardcoded "29" with (levels-1)
- COMPUTE: Show 6 terms in math breakdown
- TIMING: Ensure consistent timing values
- SPECS: Dynamic System Summary calculation
```

### Commit 4: Educational Enhancements
```
docs(module4): Add physics explanations and accuracy disclaimers

- COMPARISON: Add disclaimer for energy efficiency claims
- WRITE: Explain coercive field (Ec) concept
- READ: Explain safe voltage reasoning
- COMPUTE: Clarify Kirchhoff's Current Law application
```

### Commit 5: UI Polish
```
feat(module4): Improve UI elements and complete stub implementations

- WRITE: Use proper table widget for Level-to-Voltage mapping
- COMPUTE: Adapt input columns to array size
- READ: Full implementation of READ ALL and VERIFY buttons
```

---

## Success Criteria

### Functional
- [ ] All 6 tabs load without errors
- [ ] All buttons trigger their intended actions
- [ ] All sliders/inputs update their dependent UI elements
- [ ] Physics calculations use correct dynamic values

### Educational
- [ ] Energy claims include appropriate disclaimers
- [ ] Key physics concepts (Ec, KCL, safe read voltage) are explained
- [ ] Formulas are accurate and use dynamic values

### Technical
- [ ] `go build` succeeds with no errors
- [ ] `go test ./...` passes all tests
- [ ] No data races (verified with `-race` flag)
- [ ] App runs on Wayland/Sway without layout issues

---

## Verification Steps

### Per-Task Verification
Each task includes specific verification steps in its acceptance criteria.

### Integration Verification
1. Build: `go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
2. Run: `./fecim-lattice-tools`
3. Navigate to Module 4 (Peripheral Circuits)
4. Test each tab systematically:
   - WRITE: Program cells, verify array updates
   - READ: Read cells, verify calculations
   - COMPUTE: Run MVM, verify outputs
   - COMPARISON: Run comparison, click all buttons
   - TIMING: View diagrams, click all buttons
   - SPECS: Change config, verify summary updates

### Regression Testing
```bash
go test ./module4-circuits/... -v
go test ./... -race  # Check for data races
```

---

## File Reference Index

| File | Line Range | What's There |
|------|------------|--------------|
| `app.go` | 32-120 | CircuitsApp struct definition |
| `app.go` | 157-171 | initializeArray() |
| `tab_write.go` | 103-123 | Array size select handlers |
| `tab_write.go` | 162-171 | Pulse width entry handler |
| `tab_write.go` | 204-263 | Cell selection section |
| `tab_write.go` | 585-659 | Mapping table generation |
| `tab_read.go` | 117-137 | ADC/TIA select handlers |
| `tab_read.go` | 440-450 | Calculation display update |
| `tab_read.go` | 453-464 | READ ALL/VERIFY stubs |
| `tab_compute.go` | 100-116 | DAC/ADC bits selects |
| `tab_compute.go` | 157 | Input column loop |
| `tab_compute.go` | 374 | Math breakdown term count |
| `tab_comparison.go` | 39-40 | ANIMATE/SCALE UP buttons |
| `tab_comparison.go` | 293-294 | Energy savings annotation |
| `tab_timing.go` | 37-38 | ANIMATE/EXPORT buttons |
| `tab_timing.go` | 192 | Write timing total |
| `tab_specs.go` | 41-42 | EXPORT/COMPARE buttons |
| `tab_specs.go` | 261-276 | System Summary section |

---

**PLAN_READY: .omc/plans/module4-qa-improvements.md**
