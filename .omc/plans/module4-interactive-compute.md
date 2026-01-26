# Module 4 COMPUTE Mode Interactive Responsiveness

## Context

### Original Request
Module 4 COMPUTE mode needs interactive responsiveness - when input values change, outputs should update automatically without requiring manual COMPUTE button clicks.

### Current State Analysis

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**WRITE mode (lines 589-684)** has full interactivity:
- `opsWriteLevelSlider.OnChanged` (line 594) triggers:
  - `updateOpsWriteDataPath()` - updates DIGITAL/DAC/FeFET boxes
  - `refreshOpsWritePulse()` - updates pulse waveform
  - `updateSharedCellInfo()` - updates cell info label

**COMPUTE mode (lines 1049-1241)** is STATIC:
- `opsComputeInputs[i].OnChanged` (line 1062) only:
  - Updates `inputVector[idx]` value
  - Updates voltage label for that single entry
  - **Does NOT trigger compute or update outputs**

### Key Widgets Identified

| Widget | Location | Current Behavior | Needed Behavior |
|--------|----------|------------------|-----------------|
| `opsComputeInputs[i]` (Entry) | line 1058 | Updates local voltage label only | Trigger full recompute + update all outputs |
| `opsComputeOutputLabels[i]` (Label) | line 1110 | Static until COMPUTE clicked | Live update on input change |
| `opsComputeMathLabel` (Label) | line 1115 | Static placeholder | Live math breakdown |
| Random Bits Button | line 1124 | Updates inputs, no recompute | Update inputs + trigger recompute |
| Mode Select (Random/Ramp) | line 1086 | Updates inputs, no recompute | Update inputs + trigger recompute |
| `sharedArrayCanvas` (Raster) | line 202 | Shows array, no compute highlights | Highlight active columns during compute |

### Existing Data Path Labels (app.go lines 168-174)

**CRITICAL**: These fields ALREADY EXIST in `CircuitsApp` struct:

```go
// Compute mode INPUT data path labels
opsComputeInputDigitalLabel  *widget.Label  // Shows "x0: 128\n0b10000000"
opsComputeInputDACLabel      *widget.Label  // Shows "0.50V"

// Compute mode OUTPUT data path labels
opsComputeOutputCurrentLabel *widget.Label  // Shows "50.0 uA" or "160.0 uA (SAT)"
opsComputeOutputTIALabel     *widget.Label  // Shows "0.500 V" or "1.000 V (SAT)"
opsComputeOutputADCLabel     *widget.Label  // Shows "Level 16" or "Level 31 (SAT)"
```

**PROBLEM**: These labels are NEVER INSTANTIATED in `createComputeModePanel()` (lines 1049-1241). The update functions `updateOpsComputeInputDataPath()` (line 1263) and `updateOpsComputeOutputDataPath()` (line 1288) reference them but they are nil.

---

## Work Objectives

### Core Objective
Make COMPUTE mode fully reactive - any input change should immediately trigger matrix-vector multiplication and update all output displays, similar to how WRITE mode updates pulse diagrams on level changes.

### Deliverables
1. Input entry `OnChanged` handlers trigger automatic recompute
2. Output labels update in real-time as inputs change
3. RANDOM BITS button triggers full pipeline animation/update
4. Data path boxes show LIVE computed values
5. Array canvas highlights active columns/rows during compute (stretch goal)
6. Math breakdown updates dynamically

### Definition of Done
- [ ] Changing any input entry immediately updates all 8 output labels
- [ ] RANDOM BITS button shows animated update of full pipeline
- [ ] Mode selector (Random/Ramp) updates inputs AND computes
- [ ] Math breakdown shows real formula with current values
- [ ] No manual COMPUTE button click required for basic operation
- [ ] COMPUTE button remains available for explicit re-run

---

## Must Have / Must NOT Have

### Must Have (Guardrails)
- Use `fyne.Do()` for all UI updates from background goroutines
- Preserve existing COMPUTE button functionality
- Thread-safe access to shared state via `ca.mu` mutex
- **CRITICAL**: `computeAndUpdateAll()` must NOT call `updateOpsComputeInputs()` to prevent infinite recursion (Entry.OnChanged -> computeAndUpdateAll -> updateOpsComputeInputs -> Entry.SetText -> OnChanged -> ...)

### Must NOT Have
- Remove COMPUTE button (keep it for explicit control)
- Break existing ANIMATE button functionality
- Change the MVM algorithm/physics model
- Add new dependencies

---

## Task Flow

```
[TODO-1: Create unified compute+update function]
       |
       v
[TODO-2: Wire input entry OnChanged to auto-compute]
       |
       v
[TODO-3: Wire RANDOM BITS to auto-compute]
       |
       v
[TODO-4: Wire mode selector to auto-compute]
       |
       v
[TODO-5: Instantiate existing data path labels in createComputeModePanel]
       |
       v
[TODO-6: Array highlight on compute] (STRETCH)
       |
       v
[TODO-7: Refactor onOpsCompute to use computeAndUpdateAll]
       |
       v
[TODO-8: Initial compute on mode switch]
       |
       v
[TODO-9: (OPTIONAL) Add debounce for rapid typing]
       |
       v
[VERIFY: Test reactive behavior]
```

---

## Detailed TODOs

### TODO-1: Create unified compute-and-update function

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** After `updateOpsComputeInputs()` (around line 1261)

**Action:** Create new function `computeAndUpdateAll()` that extracts MVM computation logic from `onOpsCompute` and calls all update functions.

**CRITICAL**: This function must NOT call `updateOpsComputeInputs()` to prevent Entry->OnChanged recursion.

**Complete Function Body:**
```go
// computeAndUpdateAll performs MVM and updates all output displays
// Called by: input changes, RANDOM BITS, mode selector, COMPUTE button
// IMPORTANT: Does NOT call updateOpsComputeInputs() to prevent Entry->OnChanged recursion
func (ca *CircuitsApp) computeAndUpdateAll() {
    // 1. MVM computation (extracted from onOpsCompute lines 1546-1559)
    ca.mu.Lock()
    rows := min(8, ca.arrayRows)
    cols := min(8, ca.arrayCols)

    // MVM: output = weights * input
    for r := 0; r < rows && r < len(ca.arrayWeights); r++ {
        sum := 0.0
        for c := 0; c < cols && c < len(ca.arrayWeights[r]); c++ {
            conductance := 1.0 + float64(ca.arrayWeights[r][c])/29.0*99.0
            voltage := float64(ca.inputVector[c]) / 255.0
            sum += conductance * voltage
        }
        ca.outputVector[r] = sum
    }
    ca.mu.Unlock()

    // 2. Update output labels (extracted from onOpsCompute lines 1562-1589)
    ca.mu.RLock()
    for i := 0; i < 8 && i < len(ca.outputVector); i++ {
        if ca.opsComputeOutputLabels[i] != nil {
            rawCurrent := ca.outputVector[i] // uA

            // TIA conversion: current (uA) -> voltage (V)
            tiaVoltage := ca.tia.Convert(rawCurrent * 1e-6) // Convert uA to A for TIA

            // ADC conversion: voltage -> digital level (5-bit: 0-31)
            adcLevel := ca.adc.Convert(tiaVoltage)

            // Check for TIA saturation (current > 100 uA causes clamp to 1V)
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
    ca.mu.RUnlock()

    // 3. Update math breakdown
    ca.updateOpsComputeMath()

    // 4. Update data path displays (these functions already exist at lines 1263 and 1288)
    ca.updateOpsComputeInputDataPath()
    ca.updateOpsComputeOutputDataPath()
}
```

**Acceptance Criteria:**
- [ ] Function exists and compiles
- [ ] Performs same computation as onOpsCompute
- [ ] Does not call updateOpsComputeInputs() (no recursion risk)
- [ ] Does not duplicate code (refactor onOpsCompute to call this)

---

### TODO-2: Wire input entry OnChanged to auto-compute

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** Lines 1062-1074 (inside `createComputeModePanel`)

**Current Code (lines 1062-1074):**
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

**New Code:**
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
    // NEW: Auto-compute on input change
    // NOTE: computeAndUpdateAll() does NOT call updateOpsComputeInputs(),
    // so there is no Entry->OnChanged recursion risk
    ca.computeAndUpdateAll()
}
```

**Acceptance Criteria:**
- [ ] Typing in any x0-x7 entry immediately updates y0-y7 outputs
- [ ] No infinite recursion on input change
- [ ] No UI lag during rapid typing (see TODO-9 for optional debounce)

---

### TODO-3: Wire RANDOM BITS button to auto-compute

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** Lines 1124-1131 (RANDOM BITS button handler)

**Current Code (lines 1124-1131):**
```go
randomBitsBtn := widget.NewButton("RANDOM BITS", func() {
    ca.mu.Lock()
    for i := range ca.inputVector {
        ca.inputVector[i] = rand.Intn(256)
    }
    ca.mu.Unlock()
    ca.updateOpsComputeInputs()
})
```

**New Code:**
```go
randomBitsBtn := widget.NewButton("RANDOM BITS", func() {
    ca.mu.Lock()
    for i := range ca.inputVector {
        ca.inputVector[i] = rand.Intn(256)
    }
    ca.mu.Unlock()
    ca.updateOpsComputeInputs()
    // NEW: Auto-compute after randomizing
    ca.computeAndUpdateAll()
})
```

**Acceptance Criteria:**
- [ ] Clicking RANDOM BITS updates inputs AND immediately shows computed outputs
- [ ] Full pipeline displays update (input entries, voltage labels, output labels, math)

---

### TODO-4: Wire mode selector to auto-compute

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** Lines 1086-1103 (mode selector callback)

**Current Code (lines 1086-1103):**
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
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
    }
})
```

**New Code:**
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
        ca.computeAndUpdateAll()  // NEW
    case "Ramp":
        ca.mu.Lock()
        for i := range ca.inputVector {
            ca.inputVector[i] = i * 255 / max(1, len(ca.inputVector)-1)
        }
        ca.mu.Unlock()
        ca.updateOpsComputeInputs()
        ca.computeAndUpdateAll()  // NEW
    }
})
```

**Acceptance Criteria:**
- [ ] Selecting "Random" updates inputs AND computes outputs
- [ ] Selecting "Ramp" updates inputs AND computes outputs
- [ ] "Manual" selection does not trigger compute

---

### TODO-5: Instantiate existing data path labels in createComputeModePanel

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** Inside `createComputeModePanel()` function (around lines 1145-1149 and 1177-1180)

**Problem:** The existing struct fields in app.go (lines 168-174) are NEVER INSTANTIATED:
- `opsComputeInputDigitalLabel`
- `opsComputeInputDACLabel`
- `opsComputeOutputCurrentLabel`
- `opsComputeOutputTIALabel`
- `opsComputeOutputADCLabel`

The existing update functions at lines 1263-1329 reference these but they are nil.

**Action:** Replace static `createLabeledBox` calls with `createLabeledBoxWithLabel` using the existing struct fields.

**STEP 1: Instantiate input labels (insert BEFORE line 1147):**
```go
// Instantiate input data path labels (struct fields already exist in app.go)
ca.opsComputeInputDigitalLabel = widget.NewLabel("x0-x7\n(0-255)")
ca.opsComputeInputDACLabel = widget.NewLabel("→ 0-1V\neach")
```

**STEP 2: Replace input pipeline boxes (lines 1147-1149):**

**Current:**
```go
digitalSummaryBox := ca.createLabeledBox("8× DIGITAL", "x0-x7\n(0-255)", sharedtheme.ColorPrimary)
dacSummaryBox := ca.createLabeledBox("8× DAC", "→ 0-1V\neach", sharedtheme.ColorAccent)
columnSummaryBox := ca.createLabeledBox("8 COLUMNS", "Voltages\napplied", sharedtheme.ColorSuccess)
```

**New:**
```go
digitalSummaryBox := ca.createLabeledBoxWithLabel("8× DIGITAL", ca.opsComputeInputDigitalLabel, sharedtheme.ColorPrimary)
dacSummaryBox := ca.createLabeledBoxWithLabel("8× DAC", ca.opsComputeInputDACLabel, sharedtheme.ColorAccent)
columnSummaryBox := ca.createLabeledBox("8 COLUMNS", "Voltages\napplied", sharedtheme.ColorSuccess)  // Keep static
```

**STEP 3: Instantiate output labels (insert BEFORE line 1177):**
```go
// Instantiate output data path labels (struct fields already exist in app.go)
ca.opsComputeOutputCurrentLabel = widget.NewLabel("y0-y7\n(KCL)")
ca.opsComputeOutputTIALabel = widget.NewLabel("I→V\n10kΩ")
ca.opsComputeOutputADCLabel = widget.NewLabel("5-bit\n0-31")
```

**STEP 4: Replace output pipeline boxes (lines 1177-1180):**

**Current:**
```go
rowSumBox := ca.createLabeledBox("8× ROW SUM", "y0-y7\n(KCL)", sharedtheme.ColorWarning)
tiaSummaryBox := ca.createLabeledBox("8× TIA", "I→V\n10kΩ", sharedtheme.ColorInfo)
adcSummaryBox := ca.createLabeledBox("8× ADC", "5-bit\n0-31", sharedtheme.ColorSuccess)
levelSummaryBox := ca.createLabeledBox("8× LEVEL", "Digital\noutput", sharedtheme.ColorPrimary)
```

**New:**
```go
rowSumBox := ca.createLabeledBoxWithLabel("8× ROW SUM", ca.opsComputeOutputCurrentLabel, sharedtheme.ColorWarning)
tiaSummaryBox := ca.createLabeledBoxWithLabel("8× TIA", ca.opsComputeOutputTIALabel, sharedtheme.ColorInfo)
adcSummaryBox := ca.createLabeledBoxWithLabel("8× ADC", ca.opsComputeOutputADCLabel, sharedtheme.ColorSuccess)
levelSummaryBox := ca.createLabeledBox("8× LEVEL", "Digital\noutput", sharedtheme.ColorPrimary)  // Keep static
```

**Acceptance Criteria:**
- [ ] Labels are instantiated with initial placeholder text
- [ ] `createLabeledBoxWithLabel` used for dynamic boxes
- [ ] Existing update functions now work (labels are no longer nil)
- [ ] Data path boxes show actual computed values, not placeholders
- [ ] Values update live when inputs change

---

### TODO-6: Array canvas highlights active columns during compute (STRETCH GOAL)

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** `drawSharedArray()` function (lines 226-447)

**Current Behavior:** In ModeCompute (lines 376-443), draws static input/output arrows and labels.

**Enhancement:** Add visual indication of which columns/rows are "active" based on non-zero input values.

**Implementation:**
1. Add field `computeHighlightActive bool` to CircuitsApp
2. In `computeAndUpdateAll()`, set flag before compute, clear after
3. In `drawSharedArray()` ModeCompute case, check input values:
   - If `inputVector[c] > 0`, draw column with brighter highlight
   - Highlight rows proportional to output current

**Acceptance Criteria:**
- [ ] Columns with non-zero inputs show visual emphasis
- [ ] Rows show intensity based on output magnitude
- [ ] Animation optional (can be static highlight)

---

### TODO-7: Refactor onOpsCompute to use computeAndUpdateAll

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** `onOpsCompute()` function (lines 1544-1596)

**Action:** Refactor to call shared function, only add status message

**New Code:**
```go
func (ca *CircuitsApp) onOpsCompute() {
    ca.computeAndUpdateAll()
    ca.operationsStatusLabel.SetText("Compute complete in ~20ns")
}
```

**Acceptance Criteria:**
- [ ] COMPUTE button still works
- [ ] Behavior identical to before
- [ ] No code duplication

---

### TODO-8: Initial compute on mode switch to COMPUTE

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Location:** `onModeChanged()` function (lines 458-476)

**Action:** Trigger initial compute when switching to COMPUTE mode so outputs are populated.

**Current Code (lines 458-476):**
```go
func (ca *CircuitsApp) onModeChanged(mode string) {
    ca.mu.Lock()
    switch mode {
    case "WRITE":
        ca.currentMode = ModeWrite
    case "READ":
        ca.currentMode = ModeRead
    case "COMPUTE":
        ca.currentMode = ModeCompute
    }
    ca.mu.Unlock()

    ca.updateOperationsPanels()
    ca.updateModeHelp()
    ca.refreshSharedArray()
    ca.updateSharedCellInfo()
}
```

**New Code:**
```go
func (ca *CircuitsApp) onModeChanged(mode string) {
    ca.mu.Lock()
    switch mode {
    case "WRITE":
        ca.currentMode = ModeWrite
    case "READ":
        ca.currentMode = ModeRead
    case "COMPUTE":
        ca.currentMode = ModeCompute
    }
    ca.mu.Unlock()

    ca.updateOperationsPanels()
    ca.updateModeHelp()
    ca.refreshSharedArray()
    ca.updateSharedCellInfo()

    // NEW: Auto-compute when entering COMPUTE mode
    if mode == "COMPUTE" {
        ca.computeAndUpdateAll()
    }
}
```

**Acceptance Criteria:**
- [ ] Switching to COMPUTE mode shows computed outputs immediately
- [ ] No need to click COMPUTE button after mode switch

---

### TODO-9: (OPTIONAL) Add debounce for rapid typing

**File:** `module4-circuits/pkg/gui/tab_operations.go`

**Status:** OPTIONAL - only implement if UI becomes sluggish during rapid typing

**Location:** Add new field to CircuitsApp struct and modify `computeAndUpdateAll()` call sites

**Implementation:**
1. Add debounce timer field: `computeDebounceTimer *time.Timer`
2. Wrap `computeAndUpdateAll()` calls in TODO-2 (OnChanged handler) with debounce:
```go
// In OnChanged handler:
if ca.computeDebounceTimer != nil {
    ca.computeDebounceTimer.Stop()
}
ca.computeDebounceTimer = time.AfterFunc(50*time.Millisecond, func() {
    ca.computeAndUpdateAll()
})
```

**Acceptance Criteria:**
- [ ] Rapid typing does not cause UI lag
- [ ] Final value is always computed after typing stops
- [ ] Debounce delay is short enough to feel responsive (50-100ms)

---

## Commit Strategy

### Commit 1: Core reactive compute function
- Add `computeAndUpdateAll()` function (TODO-1)
- Refactor `onOpsCompute()` to use it (TODO-7)

### Commit 2: Wire all input triggers
- Update input entry OnChanged handlers (TODO-2)
- Update RANDOM BITS button (TODO-3)
- Update mode selector callbacks (TODO-4)
- Add initial compute on mode switch (TODO-8)

### Commit 3: Live data path displays
- Instantiate existing data path label fields (TODO-5)
- Replace static boxes with dynamic ones
- Verify update functions now work

### Commit 4: Array highlights (STRETCH - skip if time constrained)
- Add highlight state (TODO-6)
- Update draw function
- Wire to compute cycle

### Commit 5: (OPTIONAL) Debounce
- Add debounce timer if needed (TODO-9)

---

## Success Criteria

| Criteria | Test |
|----------|------|
| Input changes trigger recompute | Type "128" in x0, see y0-y7 update immediately |
| RANDOM BITS updates pipeline | Click button, see all 8 inputs AND outputs change |
| Mode selector triggers compute | Select "Ramp", see ramp values AND computed outputs |
| Math breakdown is live | Change x0, see formula update with new values |
| No manual compute required | All above work without clicking COMPUTE button |
| COMPUTE button still works | Click COMPUTE, see status message + outputs |
| No infinite recursion | Typing in entry does not freeze UI |
| Data path boxes show live values | Input/output pipeline boxes update with real values |
| No UI lag | Rapid typing doesn't freeze interface |

---

## Files to Modify

| File | Changes |
|------|---------|
| `module4-circuits/pkg/gui/tab_operations.go` | Add computeAndUpdateAll(), wire OnChanged handlers, update mode switch, instantiate data path labels |
| `module4-circuits/pkg/gui/app.go` | No changes needed - fields already exist at lines 168-174 |

---

## Recursion Safety Analysis

**Potential Recursion Chain (BLOCKED):**
```
Entry.OnChanged (user types)
  -> computeAndUpdateAll()
    -> updateOpsComputeInputDataPath()  [OK - updates different labels]
    -> updateOpsComputeOutputDataPath() [OK - updates different labels]
    -> updateOpsComputeMath()           [OK - updates math label]
    -> DOES NOT call updateOpsComputeInputs() [SAFE]
```

**If computeAndUpdateAll() called updateOpsComputeInputs() (DANGEROUS):**
```
Entry.OnChanged (user types)
  -> computeAndUpdateAll()
    -> updateOpsComputeInputs()
      -> Entry.SetText()
        -> Entry.OnChanged (TRIGGERED AGAIN!)
          -> computeAndUpdateAll()
            -> INFINITE LOOP!
```

**Resolution:** The `computeAndUpdateAll()` function explicitly does NOT call `updateOpsComputeInputs()`. The only places that call `updateOpsComputeInputs()` are:
- RANDOM BITS button (calls it BEFORE computeAndUpdateAll)
- Mode selector (calls it BEFORE computeAndUpdateAll)

This is safe because those call sites explicitly SET the inputVector values first, then update the UI, then compute. The Entry.OnChanged won't re-trigger because the value hasn't changed.
