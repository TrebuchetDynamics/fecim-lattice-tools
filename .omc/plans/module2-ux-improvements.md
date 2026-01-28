# Work Plan: Module2 Crossbar GUI UX/UI Improvements

**Created:** 2026-01-28
**Module:** module2-crossbar/pkg/gui
**Scope:** Fix identified UX/UI inconsistencies between standard and enhanced modes

---

## Context

### Original Request
Analyze module2-crossbar GUI for obvious UX/UI issues, fix them, update documentation, and continuously test the app.

### Research Findings
Code review of the following files revealed 8 actionable UX/UI issues:
- `app.go` (849 lines) - Standard mode layout and controls
- `app_tabs.go` (424 lines) - Enhanced mode layout structure
- `app_controls.go` (336 lines) - Enhanced mode control widgets
- `app_enhanced.go` (242 lines) - Enhanced MVM execution
- `animation.go` (263 lines) - MVM animation flow
- `embedded.go` (100 lines) - Embedded interface

### Key Files Affected
| File | Purpose |
|------|---------|
| `module2-crossbar/pkg/gui/app_controls.go` | Enhanced mode control widgets |
| `module2-crossbar/pkg/gui/app.go` | Standard mode layout + CrossbarApp struct |
| `module2-crossbar/pkg/gui/app_tabs.go` | Enhanced layout structure |
| `module2-crossbar/pkg/gui/embedded.go` | Embedded mode interface |
| `docs/development/GUI/GUI.module2.md` | Documentation |

---

## Work Objectives

### Core Objective
Ensure parity and consistency between standard and enhanced GUI modes, fixing all identified UX issues.

### Deliverables
1. All 8 UX issues fixed
2. Both GUI modes have consistent controls
3. Documentation updated with fix details
4. All tests passing

### Definition of Done
- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes (117 tests)
- [ ] App launches in both standard and enhanced mode without errors
- [ ] All 8 issues verified as fixed via manual testing

---

## Guardrails

### MUST Have
- Run MVM button visible in enhanced mode
- 2T1R architecture button in enhanced mode (parity with standard)
- Consistent ADC label format between modes
- Colormap selector works correctly for all tabs

### MUST NOT Have
- Breaking changes to existing functionality
- New dependencies
- Changes to crossbar physics calculations
- Removal of existing features

### Implementation Note
**Line numbers are approximate guides.** Match code changes by pattern context, not exact line numbers. The codebase may have shifted since this plan was created.

---

## Task Flow

```
TASK-1 (Add Run MVM button)
    |
    v
TASK-2 (Add 2T1R button) --> TASK-3 (Fix ADC label)
    |
    v
TASK-4 (Fix colormap selector)
    |
    v
TASK-5 (Add slider range labels)
    |
    v
TASK-6 (Fix hover info width)
    |
    v
TASK-7 (Set onboarding content)
    |
    v
TASK-8 (Update documentation)
    |
    v
VERIFICATION (go test, manual testing)
```

---

## Detailed Tasks

### TASK-1: Add Run MVM Button to Enhanced Mode Controls
**Priority:** HIGH (P0) - Core functionality missing
**Estimated:** 15 min

**Problem:** Enhanced mode (`app_controls.go`) has no "Run MVM" button. Users cannot manually trigger MVM computation.

**STEP 1: Add struct field in `app.go`**

**Location:** `module2-crossbar/pkg/gui/app.go` line 74 (in the CrossbarApp struct, near `resetButton`)

**Current Code (lines 73-77):**
```go
	// Simple right panel widgets (replacing custom widgets)
	resetButton      *widget.Button
	arraySizeSelect  *widget.Select // Dropdown for array size
	arraySizeLabel   *widget.Label  // Label for slider display
	arraySizeSlider  *widget.Slider // Slider for array size
```

**Add after line 74 (after `resetButton`):**
```go
	runMVMButton     *widget.Button
```

**STEP 2: Create button in `app_controls.go`**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` after line 77 (after resetButton creation)

**Current Code (lines 75-78):**
```go
func (ca *CrossbarApp) createControlWidgets() {
	// Reset button
	ca.resetButton = widget.NewButton("Reset", ca.resetArray)
	ca.resetButton.Importance = widget.MediumImportance
```

**Add after line 78:**
```go
	// Run MVM button - primary action for triggering animated MVM
	ca.runMVMButton = widget.NewButton("Run MVM", ca.runEnhancedMVM)
	ca.runMVMButton.Importance = widget.HighImportance
```

**STEP 3: Add button to action row**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` line 259

**Current Code:**
```go
actionButtons := container.NewGridWithColumns(2, ca.resetButton, exportButton)
```

**Change to:**
```go
actionButtons := container.NewGridWithColumns(3, ca.runMVMButton, ca.resetButton, exportButton)
```

**Acceptance Criteria:**
- [ ] "Run MVM" button visible in enhanced mode right panel
- [ ] Clicking button triggers animated MVM computation
- [ ] Button uses HighImportance styling (visually prominent)

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Navigate to Crossbar tab, verify Run MVM button visible and functional
```

---

### TASK-2: Add 2T1R Architecture Button to Enhanced Mode
**Priority:** HIGH (P0) - Feature parity missing
**Estimated:** 20 min

**Problem:** Standard mode has 3 architecture buttons (PASSIVE, 1T1R, 2T1R) but enhanced mode only has 2.

**STEP 1: Add 2T1R button creation**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` after line 153 (after arch1T1RBtn creation)

**Current Code (lines 152-154):**
```go
	ca.archPassiveBtn = widget.NewButtonWithIcon("PASSIVE", theme.GridIcon(), nil)
	ca.arch1T1RBtn = widget.NewButtonWithIcon("1T1R GATE", theme.VisibilityIcon(), nil)

```

**Add after line 153 (after arch1T1RBtn creation):**
```go
	ca.arch2T1RBtn = widget.NewButtonWithIcon("2T1R", theme.MoreVerticalIcon(), nil)
```

**STEP 2: Update updateArchButtons function**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` lines 156-171

**Current Code (lines 156-171):**
```go
	updateArchButtons := func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			// Selected: high importance, show checkmark in text
			ca.archPassiveBtn.SetText("● PASSIVE")
			ca.archPassiveBtn.Importance = widget.HighImportance
			ca.arch1T1RBtn.SetText("1T1R GATE")
			ca.arch1T1RBtn.Importance = widget.LowImportance
		} else {
			ca.archPassiveBtn.SetText("PASSIVE")
			ca.archPassiveBtn.Importance = widget.LowImportance
			ca.arch1T1RBtn.SetText("● 1T1R GATE")
			ca.arch1T1RBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
	}
```

**Replace with:**
```go
	updateArchButtons := func() {
		// Reset all to unselected state
		ca.archPassiveBtn.SetText("PASSIVE")
		ca.archPassiveBtn.Importance = widget.LowImportance
		ca.arch1T1RBtn.SetText("1T1R GATE")
		ca.arch1T1RBtn.Importance = widget.LowImportance
		ca.arch2T1RBtn.SetText("2T1R")
		ca.arch2T1RBtn.Importance = widget.LowImportance

		// Mark selected button
		switch ca.architecture {
		case sharedwidgets.Architecture0T1R:
			ca.archPassiveBtn.SetText("● PASSIVE")
			ca.archPassiveBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture1T1R:
			ca.arch1T1RBtn.SetText("● 1T1R GATE")
			ca.arch1T1RBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture2T1R:
			ca.arch2T1RBtn.SetText("● 2T1R")
			ca.arch2T1RBtn.Importance = widget.HighImportance
		}

		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
		ca.arch2T1RBtn.Refresh()
	}
```

**STEP 3: Add 2T1R button callback**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` after line 215 (after arch1T1RBtn.OnTapped block)

**Add after line 215:**
```go
	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return // Already selected
		}
		debug.Printf("[ARCH TOGGLE] Switched to: 2T1R")

		ca.stateMu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.stateMu.Unlock()

		updateArchButtons()

		// Update educational content
		title, content := sharedwidgets.ArchitectureInfo(sharedwidgets.Architecture2T1R)
		ca.setEducationalContent(title, content)

		// Re-run MVM
		ca.runEnhancedMVMWithCurrentInput()
	}
```

**STEP 4: Update container to 3 columns**

**Location:** `module2-crossbar/pkg/gui/app_controls.go` line 218

**Current Code:**
```go
	ca.archToggle = container.NewGridWithColumns(2, ca.archPassiveBtn, ca.arch1T1RBtn)
```

**Change to:**
```go
	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)
```

**Acceptance Criteria:**
- [ ] 2T1R button visible in enhanced mode
- [ ] Clicking 2T1R updates architecture and re-runs MVM
- [ ] Educational panel shows 2T1R info when selected
- [ ] Selection indicator (●) shows on active architecture

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Test: Click each architecture button, verify selection indicator and educational content updates
```

---

### TASK-3: Standardize ADC Label Format
**Priority:** MEDIUM (P1) - Inconsistent labeling
**Estimated:** 5 min

**Problem:** Standard mode shows "ADC Bits: 6" but enhanced mode just shows "6".

**Locations:**
- `app.go:320` shows "ADC Bits: %d"
- `app_controls.go:111` shows "%d"

**Fix in `app_controls.go` line 111:**
Change:
```go
ca.adcBitsLabel.SetText(fmt.Sprintf("%d", bits))
```
To:
```go
ca.adcBitsLabel.SetText(fmt.Sprintf("%d bits", bits))
```

Also update initial value at line 104:
```go
ca.adcBitsLabel = widget.NewLabel("6 bits")
```

**Note:** The row label already says "ADC:" so we just need "6 bits" not "ADC Bits: 6".

**Acceptance Criteria:**
- [ ] ADC label shows "6 bits" format in enhanced mode
- [ ] Updates correctly when slider moves

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Move ADC slider, verify label shows "X bits" format
```

---

### TASK-4: Fix Colormap Selector for Non-Heatmap Tabs
**Priority:** MEDIUM (P1) - Unexpected behavior
**Estimated:** 10 min

**Problem:** When on "Ideal vs Actual" or "Accuracy Analysis" tabs, colormap changes fall through to conductance heatmap.

**Location:** `module2-crossbar/pkg/gui/app_controls.go` lines 116-143

**IMPORTANT:** `BeforeAfterToggle` does NOT have a `SetColormap()` method. The widget only has `SetMode()` and `SetData()`. The heatmaps inside BeforeAfterToggle use their own colormaps that are set during mode changes.

**Current Code (lines 116-143):**
```go
ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
	// Change colormap for the currently active tab and store the selection
	if ca.tabs != nil {
		switch ca.tabs.Selected().Text {
		case "Conductance":
			ca.conductanceHeatmap.SetColormap(s)
			ca.condLegend.SetColormap(s)
			ca.condColormap = s
		case "IR Drop":
			ca.irDropHeatmap.SetColormap(s)
			ca.irLegend.SetColormap(s)
			ca.irColormap = s
		case "Sneak Paths":
			ca.sneakPathHeatmap.SetColormap(s)
			ca.sneakLegend.SetColormap(s)
			ca.sneakColormap = s
		default:
			// For other tabs, default to conductance
			ca.conductanceHeatmap.SetColormap(s)
			ca.condLegend.SetColormap(s)
			ca.condColormap = s
		}
	} else {
		ca.conductanceHeatmap.SetColormap(s)
		ca.condLegend.SetColormap(s)
		ca.condColormap = s
	}
})
```

**Fix - Replace default case to ignore non-heatmap tabs:**
```go
ca.colormapSelect = widget.NewSelect([]string{"fecim", "viridis", "plasma", "coolwarm"}, func(s string) {
	// Change colormap for the currently active tab and store the selection
	if ca.tabs != nil {
		// Get base tab name (strip badge suffix if present)
		tabName := ca.getBaseTabName(ca.tabs.Selected().Text)
		switch tabName {
		case "Conductance":
			ca.conductanceHeatmap.SetColormap(s)
			ca.condLegend.SetColormap(s)
			ca.condColormap = s
		case "IR Drop":
			ca.irDropHeatmap.SetColormap(s)
			ca.irLegend.SetColormap(s)
			ca.irColormap = s
		case "Sneak Paths":
			ca.sneakPathHeatmap.SetColormap(s)
			ca.sneakLegend.SetColormap(s)
			ca.sneakColormap = s
		case "Ideal vs Actual", "Accuracy Analysis", "Input/Output":
			// These tabs don't have user-controllable colormaps
			// BeforeAfterToggle manages its own colormaps based on mode
			// Ignore colormap changes for these tabs
			return
		default:
			// Unknown tab - ignore to avoid unexpected side effects
			return
		}
	} else {
		ca.conductanceHeatmap.SetColormap(s)
		ca.condLegend.SetColormap(s)
		ca.condColormap = s
	}
})
```

**Acceptance Criteria:**
- [ ] Colormap selector changes Conductance heatmap when on "Conductance" tab
- [ ] Colormap selector changes IR Drop heatmap when on "IR Drop" tab
- [ ] Colormap selector changes Sneak Paths heatmap when on "Sneak Paths" tab
- [ ] Colormap selector is ignored when on "Ideal vs Actual", "Accuracy Analysis", or "Input/Output" tabs
- [ ] No unexpected heatmap changes on non-heatmap tabs

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Switch to each tab, change colormap, verify expected behavior
```

---

### TASK-5: Add Array Size Slider Min/Max Labels
**Priority:** LOW (P2) - Usability improvement
**Estimated:** 10 min

**Problem:** Users can't see the range limits (8-128) without moving the slider.

**Location:** `module2-crossbar/pkg/gui/app_controls.go` lines 228-233

**Current Code:**
```go
arraySizeRow := container.NewBorder(
	nil, nil,
	widget.NewLabel("Array:"),
	ca.arraySizeLabel,
	ca.arraySizeSlider,
)
```

**Fix:** Add min/max labels flanking the slider:
```go
minLabel := widget.NewLabel("8")
minLabel.TextStyle = fyne.TextStyle{Monospace: true}
maxLabel := widget.NewLabel("128")
maxLabel.TextStyle = fyne.TextStyle{Monospace: true}

sliderWithLabels := container.NewBorder(
	nil, nil,
	minLabel,
	maxLabel,
	ca.arraySizeSlider,
)

arraySizeRow := container.NewBorder(
	nil, nil,
	widget.NewLabel("Array:"),
	ca.arraySizeLabel,
	sliderWithLabels,
)
```

**Acceptance Criteria:**
- [ ] Min (8) and Max (128) labels visible flanking slider
- [ ] Labels use monospace font for alignment

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Verify min/max labels visible next to array size slider
```

---

### TASK-6: Make Hover Info Width Responsive
**Priority:** LOW (P2) - Small screen compatibility
**Estimated:** 10 min

**Problem:** Fixed width `450` for hover info doesn't adapt to narrow windows.

**Location:** `module2-crossbar/pkg/gui/app_controls.go` line 324

**Current Code:**
```go
hoverInfoContainer := container.NewGridWrap(fyne.NewSize(450, 20), ca.hoverInfoLabel)
```

**IMPORTANT:** `Label.SetMinSize()` does NOT exist in Fyne. We MUST use container wrapping only.

**Fix - Use smaller fixed width that works on more screen sizes:**
```go
hoverInfoContainer := container.NewGridWrap(fyne.NewSize(300, 20), ca.hoverInfoLabel)
```

**Alternative (more flexible):** Use HBox with spacer:
```go
hoverInfoContainer := container.NewHBox(ca.hoverInfoLabel, layout.NewSpacer())
```

**Note:** The second approach allows natural sizing but may cause layout shifts. The fixed 300px approach is safer and still provides room for hover text while fitting narrow windows.

**Acceptance Criteria:**
- [ ] Hover info doesn't break layout on narrow windows
- [ ] Text truncation still works (ellipsis)

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# Resize window to narrow width, verify hover info doesn't overflow
```

---

### TASK-7: Set Proper Onboarding Content on Enhanced Mode Start
**Priority:** LOW (P2) - First-time user experience
**Estimated:** 5 min

**Problem:** Initial content in `app_tabs.go:231` is just generic placeholder text.

**Location:** `module2-crossbar/pkg/gui/app_tabs.go:231-232`

**Current Code:**
```go
ca.eduContentLabel = widget.NewLabel("CROSSBAR MVM\n\nClick a button to start\na demonstration.")
```

**Analysis:** The proper onboarding content IS set in `app.go:221-235` after `createEnhancedMainLayout()` returns. This is working correctly for `RunWithLayout(true)` but NOT for embedded mode.

**Fix in `embedded.go` after line 65:**

**Current Code (lines 63-67):**
```go
	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	return content
```

**Add onboarding content call after line 65:**
```go
	// Initialize displays
	e.updateConductanceDisplay()
	e.updateStatus("Ready. Program weights and run MVM operations.")

	// Set first-load onboarding content (same as standalone mode)
	e.setEducationalContent("Getting Started",
		"Welcome to Crossbar MVM!\n\n"+
			"Quick Start:\n"+
			"1. Hover over cells to see\n"+
			"   conductance values\n"+
			"2. Click cells for details\n"+
			"3. Adjust controls on right\n"+
			"4. Explore IR Drop & Sneak\n"+
			"   Path analysis tabs\n\n"+
			"Key Concepts:\n"+
			"* 30 analog levels/cell\n"+
			"* MVM = Matrix x Vector\n"+
			"* All rows compute in 1 step")

	return content
```

**Acceptance Criteria:**
- [ ] Embedded mode shows proper onboarding content
- [ ] Content matches standalone mode

**Test Verification:**
```bash
go build -o /tmp/test-app ./cmd/fecim-lattice-tools && /tmp/test-app
# In unified app, switch to Crossbar tab, verify educational panel shows proper onboarding
```

---

### TASK-8: Update GUI Documentation
**Priority:** MEDIUM (P1) - Documentation accuracy
**Estimated:** 15 min

**Location:** `docs/development/GUI/GUI.module2.md`

**Updates Required:**
1. Add new UX issues to bug list with FIXED status
2. Update widget documentation for new buttons
3. Document architectural parity between modes

**New Bugs to Document:**
```yaml
Bugs:
  - [x] BUG-M2-006: Missing Run MVM button in enhanced mode - FIXED (2026-01-28)
  - [x] BUG-M2-007: Missing 2T1R architecture button in enhanced mode - FIXED (2026-01-28)
  - [x] BUG-M2-008: Inconsistent ADC label format between modes - FIXED (2026-01-28)
  - [x] BUG-M2-009: Colormap selector falls through on non-heatmap tabs - FIXED (2026-01-28)
  - [x] BUG-M2-010: Array size slider missing min/max labels - FIXED (2026-01-28)
  - [x] BUG-M2-011: Hover info fixed width causes layout issues - FIXED (2026-01-28)
  - [x] BUG-M2-012: Embedded mode missing onboarding content - FIXED (2026-01-28)
```

**Acceptance Criteria:**
- [ ] All 7 new bugs documented
- [ ] Widget documentation reflects new buttons
- [ ] Last Updated date changed to 2026-01-28

---

## Commit Strategy

### Commit 1: Core functionality fixes
```
fix(gui/crossbar): add Run MVM button and 2T1R architecture to enhanced mode

- Add Run MVM button to enhanced mode controls panel
- Add 2T1R architecture button for feature parity with standard mode
- Wire up callbacks for new buttons
```

### Commit 2: Label and selector fixes
```
fix(gui/crossbar): standardize labels and fix colormap selector behavior

- Standardize ADC bits label format across modes
- Fix colormap selector to handle non-heatmap tabs correctly
- Add min/max labels to array size slider
```

### Commit 3: Responsive and onboarding fixes
```
fix(gui/crossbar): improve responsiveness and onboarding experience

- Make hover info width responsive for narrow windows
- Add proper onboarding content to embedded mode
```

### Commit 4: Documentation
```
docs(gui): update module2 GUI documentation with UX fixes

- Document 7 new bugs as fixed
- Update widget documentation
- Update last modified date
```

---

## Success Criteria

| Criterion | Verification Method |
|-----------|---------------------|
| Build succeeds | `go build ./...` |
| Tests pass | `go test ./...` |
| Run MVM button visible | Manual: Enhanced mode, right panel |
| 2T1R button works | Manual: Click button, check educational panel |
| ADC label consistent | Manual: Move slider, check format |
| Colormap selector correct | Manual: Each tab, change colormap |
| Slider labels visible | Manual: Check array size row |
| Hover info responsive | Manual: Resize to narrow width |
| Onboarding content shows | Manual: Embedded mode start |
| Docs updated | Review GUI.module2.md |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking existing functionality | Low | High | Run full test suite after each change |
| Layout regressions | Medium | Medium | Test on multiple window sizes |
| Callback wiring errors | Low | Medium | Test each button click manually |

---

## Notes

- All changes are isolated to the GUI layer - no physics changes
- Enhanced mode is the default for embedded use, so fixes there have higher priority
- Standard mode changes (if needed) should mirror enhanced mode patterns
- **TASK-4 CRITICAL FIX:** BeforeAfterToggle does NOT have SetColormap() - must use return statement for non-heatmap tabs
- **TASK-6 CRITICAL FIX:** Fyne Labels do NOT have SetMinSize() - must use container wrapping only
