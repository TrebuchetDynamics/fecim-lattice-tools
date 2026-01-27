---
Module: module4-circuits
Name: Peripheral Circuits Visualizer
Entry: cmd/circuits-gui/main.go
Package: fecim-lattice-tools/module4-circuits/pkg/gui
Theme: FeCIMTheme
Architecture: Unified 3-view design with embedded interface
Last Updated: 2026-01-27
---

## Bugs Summary

### Fixed Bugs
- [x] BUG-M4-002: Array cell click coordinate calculation (FIXED: uses asymmetric margins)
- [x] BUG-M4-003: computeInputRowContainer now initialized in app.go:143
- [x] BUG-M4-005: Race condition in drawSharedArray (FIXED: currentMode read once under lock at line 253)
- [x] BUG-M4-001: Operations panel visibility sync on mode change
- [x] BUG-M4-004: Missing canvas refresh in Start() for shared array

### Open Bugs
(none)

### Physics Issues (from HZO_PARAMETERS.md research) - ALL FIXED
- [x] PHYS-001: WRITE voltage range corrected (derived from material Vc)
- [x] PHYS-002: READ voltage slider max fixed (derived from FieldMinRatio * Vc)
- [x] PHYS-003: COMPUTE voltage note updated (uses read range for compute-safe)
- [x] PHYS-004: Voltage ranges now loaded from physics.yaml calibration section

### UX Issues - ALL FIXED
- [x] UX-001: COMPUTE button redundant (auto-compute implemented on input change)
- [x] UX-002: Export buttons - FIXED (2026-01-26): Now show "Coming soon" dialogs with helpful workarounds
- [x] UX-003: Mode selection refactored (2026-01-27): Mode buttons replace RadioGroup

---

## Recent Changes (2026-01-27)

### Major Refactor: Unified Device Simulation View
- **Replaced** `tab_operations.go` with `tab_unified.go` - single unified device simulation
- **New file** `device_state.go` - DeviceState struct manages all simulation state
- **Mode buttons** replace RadioGroup: READ, WRITE, COMPUTE buttons with visual highlighting
- **Material selector** - dropdown to select ferroelectric material (FeCIM HZO, etc.)
- **Architecture toggle** - PASSIVE/1T1R/2T1R buttons
- **Dynamic voltage ranges** - derived from physics.yaml and material properties (no hardcoded values)

### Voltage Range System
- All voltage thresholds now derived from material properties:
  - **Coercive voltage (Vc)** = Ec × thickness (from material)
  - **Read range**: 0 to FieldMinRatio × Vc (from physics.yaml calibration.field_min_ratio)
  - **Write range**: Vc to FieldMaxRatio × Vc (from physics.yaml calibration.field_max_ratio)
- DAC preset buttons show actual voltage ranges based on selected material
- No hardcoded voltage constants in device_state.go

### Operation Mode System
- **OpMode enum** replaces OperationMode:
  - `OpModeRead`: Single row active, safe voltage (0 to read max)
  - `OpModeWrite`: Single row active, write voltage on selected column
  - `OpModeCompute`: All rows active, input vector (0 to read max for MVM)
- Mode buttons auto-configure WL and DAC settings when clicked

---

## File Structure

| File | Purpose |
|------|---------|
| `app.go` | Main CircuitsApp struct, window setup, view switching |
| `device_state.go` | DeviceState struct, voltage ranges, simulation logic |
| `tab_unified.go` | Unified device simulation view (replaces tab_operations.go) |
| `tab_comparison.go` | FeFET vs GPU vs CPU comparison view |
| `tab_reference.go` | Timing diagrams and specifications reference |
| `tab_reference_timing.go` | Timing diagram drawing functions |
| `tab_reference_specs.go` | Specifications section |
| `drawing.go` | Primitive drawing functions |
| `helpers.go` | DAC/TIA/ADC box drawing helpers |
| `font.go` | Bitmap font patterns for canvas text |
| `embedded.go` | Embedded interface for main app integration |

---

## Screens

### Main Window (app.go:274-367)
**Purpose**: Top-level window with 3-view architecture
**File**: app.go:274-367
**State**: window, fyneApp, deviceState
**Layout**:
```
Border
├─ Top: Header with inline view selector
│  ├─ Label: "View:"
│  ├─ Select: ["OPERATIONS", "COMPARISON", "REFERENCE"]
│  ├─ Spacer
│  └─ Label: "3 Views | DAC -> FeFET -> TIA -> ADC"
├─ Bottom: Footer
│  └─ Label: "FeCIM Ferroelectric Compute-in-Memory | Based on Published Research"
└─ Center: Stack container
   ├─ OPERATIONS view (visible on start) - from tab_unified.go
   ├─ COMPARISON view (hidden)
   └─ REFERENCE view (hidden)
```

---

### OPERATIONS View (tab_unified.go:29-61)
**Purpose**: Unified device simulation with material selection and mode buttons
**File**: tab_unified.go:29-61
**State**: deviceState (OpMode, VoltageRange, material, activeRows, dacVoltages)
**Layout**:
```
Border
├─ Top: Signal chain header + DAC section
│  ├─ VBox
│  │  ├─ HBox
│  │  │  ├─ Label: "SIGNAL CHAIN: DAC -> Array -> TIA -> ADC" (bold)
│  │  │  ├─ Spacer
│  │  │  ├─ Material selector: [FeCIM HZO, HZO (Si-doped), ...]
│  │  │  ├─ Spacer
│  │  │  ├─ Architecture toggle: [PASSIVE] [1T1R] [2T1R]
│  │  │  ├─ Spacer
│  │  │  └─ operationsStatusLabel
│  │  ├─ operationsModeHelp (mode + architecture help text)
│  │  └─ Separator
│  └─ DAC presets section
│     ├─ HBox
│     │  ├─ Label: "DAC Presets:"
│     │  ├─ dacPresetReadBtn: "Read (0-0.5V)" (dynamic label)
│     │  ├─ dacPresetWriteBtn: "Write (1.0-2.5V)" (dynamic label)
│     │  ├─ Button: "Input Vector"
│     │  ├─ Button: "Random"
│     │  ├─ Spacer
│     │  ├─ dacRangeLabel: "Mode: Read (0-0.5V)"
│     │  ├─ Label: "Set All (V):"
│     │  └─ Entry: allEntry
├─ Bottom: Action buttons
│  ├─ HBox
│  │  ├─ Button: "Write Cell" (HighImportance)
│  │  ├─ Button: "Read/Sense"
│  │  ├─ Button: "Compute MVM"
│  │  ├─ Spacer
│  │  ├─ Button: "Animate"
│  │  ├─ Button: "Random Array"
│  │  └─ Button: "Reset Array"
└─ Center: HSplit (10% WL selector, 90% array)
   ├─ Left: Word Line selector
   │  ├─ Label: "WORD LINES" (bold)
   │  ├─ WL checkboxes: WL0, WL1, ... WL7
   │  ├─ Separator
   │  ├─ Label: "Mode:"
   │  ├─ modeReadBtn: "READ" (highlighted when active)
   │  ├─ modeWriteBtn: "WRITE"
   │  └─ modeComputeBtn: "COMPUTE"
   └─ Right: Array visualization
      ├─ UnifiedTappableCanvas (400x350px)
      │  └─ Raster: drawUnifiedArray
      │     ├─ DAC boxes (top, per column, voltage-colored)
      │     ├─ WL lines (horizontal, orange=active, dim=inactive)
      │     ├─ BL lines (vertical, red=write, blue=read, dim=off)
      │     ├─ Cell grid (color-coded by conductance level)
      │     ├─ TIA+ADC boxes (right, per row)
      │     ├─ 1T1R/2T1R transistors (if active architecture)
      │     ├─ Operation label (top-left): "READ", "WRITE", "COMPUTE (MVM)"
      │     └─ Architecture badge (top-right): "PASSIVE", "1T1R", "2T1R"
      ├─ legendLabel
      ├─ sharedCellInfoLabel: "Cell [r,c]: State N | G=XXµS | BL=X.XXV | Material"
      └─ sharedArrayInfoLabel: "Array: 8x8 | 30 levels"
```

---

### DeviceState (device_state.go)
**Purpose**: Central state management for device simulation
**File**: device_state.go
**Key Fields**:
```go
type DeviceState struct {
    // Dimensions
    rows, cols int

    // Operation mode (READ/WRITE/COMPUTE)
    opMode OpMode

    // WL configuration
    wlMode     WLMode      // WLSingle, WLAll, WLCustom
    activeRows []bool      // true = WL HIGH

    // DAC inputs
    dacVoltages  []float64   // Voltage per column
    dacMode      DACMode     // DACReadPreset, DACWritePreset, etc.
    dacRangeMode DACRangeMode // DACRangeRead, DACRangeWrite

    // Voltage ranges (derived from material + physics.yaml)
    readRange   VoltageRange  // 0 to FieldMinRatio*Vc
    writeRange  VoltageRange  // Vc to FieldMaxRatio*Vc
    calibParams CalibrationParams // From physics.yaml

    // Computed outputs
    rowCurrents []float64   // TIA input (µA)
    rowVoltages []float64   // TIA output (V)
    rowLevels   []int       // ADC output
    saturated   []bool

    // Selection
    selectedRow, selectedCol int

    // Physics
    material *ferroelectric.HZOMaterial
    tia      *peripherals.TIA
    adc      *peripherals.ADC
}
```

**Key Methods**:
| Method | Purpose |
|--------|---------|
| `NewDeviceState(rows, cols, tia, adc)` | Create with dimensions and peripherals |
| `SetMaterial(mat)` | Change material, recalculates voltage ranges |
| `SetOperationMode(mode)` | Set READ/WRITE/COMPUTE mode |
| `SetWLSingle(row)` | Activate only specified row |
| `SetWLAll()` | Activate all rows for MVM |
| `SetDACPreset(preset, params...)` | Apply voltage preset |
| `SetDACVoltageForState(col, level)` | Set write voltage for target state |
| `Compute(weights, levels)` | Run MVM simulation |
| `GetReadRange() / GetWriteRange()` | Get voltage ranges for current material |
| `ClassifyOperation()` | Get operation name string |

---

### Voltage Range Configuration

Voltage ranges are derived from physics.yaml and material properties:

```yaml
# config/physics.yaml
calibration:
  field_min_ratio: 0.5   # Read max = 0.5 * Vc
  field_max_ratio: 2.5   # Write max = 2.5 * Vc
```

**Calculation**:
```go
func (ds *DeviceState) updateVoltageRanges() {
    Vc := ds.material.CoerciveVoltage()  // Vc = Ec × thickness

    // Read range: 0 to FieldMinRatio * Vc (safe, non-destructive)
    safeReadMax := ds.calibParams.FieldMinRatio * Vc

    // Write range: Vc to FieldMaxRatio * Vc (exceeds coercive)
    writeMin := Vc
    writeMax := ds.calibParams.FieldMaxRatio * Vc
}
```

---

### Operation Mode Rules (from docs/peripheral-circuits/circuits.operations.md)

| Mode | WL Selection | DAC Voltage | Effect |
|------|--------------|-------------|--------|
| READ | Single row | 0 to read max | Sense conductance, no change |
| WRITE | Single row | Vc to write max | Program cell state |
| COMPUTE | All rows | 0 to read max (input vector) | MVM multiply, no change |

**Mode Button Behavior** (`setOperationMode()`):
```go
switch mode {
case OpModeRead:
    ca.deviceState.SetWLSingle(selectedRow)
    ca.deviceState.SetDACPreset(DACReadPreset)
case OpModeWrite:
    ca.deviceState.SetWLSingle(selectedRow)
    ca.deviceState.SetDACPreset(DACWritePreset)
case OpModeCompute:
    ca.deviceState.SetWLAll()
    // Keep read range voltages for compute
}
```

---

### Components

| Component | Type | Purpose | File:Line | State |
|-----------|------|---------|-----------|-------|
| UnifiedTappableCanvas | Custom Widget | Clickable array with DAC/TIA/ADC | tab_unified.go:444-522 | sharedArrayCanvas |
| modeReadBtn | Button | Set READ mode | tab_unified.go:264 | opMode |
| modeWriteBtn | Button | Set WRITE mode | tab_unified.go:267 | opMode |
| modeComputeBtn | Button | Set COMPUTE mode | tab_unified.go:270 | opMode |
| dacPresetReadBtn | Button | Apply read voltage preset | tab_unified.go:132 | dacVoltages |
| dacPresetWriteBtn | Button | Apply write voltage preset | tab_unified.go:135 | dacVoltages |
| dacRangeLabel | Label | Shows current DAC range mode | tab_unified.go:147 | dacRangeMode |
| materialSelector | Select | Choose ferroelectric material | tab_unified.go:97-122 | material |
| archPassiveBtn | Button | Select passive (0T1R) architecture | tab_unified.go:1338 | architecture |
| arch1T1RBtn | Button | Select 1T1R architecture | tab_unified.go:1339 | architecture |
| arch2T1RBtn | Button | Select 2T1R architecture | tab_unified.go:1340 | architecture |
| operationsModeHelp | Label | Mode + architecture help text | tab_unified.go:79 | Updated by updateOperationClassification() |

---

### Data Flow

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| Mode button click | modeReadBtn/modeWriteBtn/modeComputeBtn | opMode, WL config, DAC config, button highlighting | tab_unified.go:295-328 |
| Material selection | materialSelector | material, voltage ranges, DAC labels | tab_unified.go:104-115 |
| Architecture change | archPassiveBtn/arch1T1RBtn/arch2T1RBtn | architecture, WL state, transistor display | tab_unified.go:1366-1406 |
| DAC preset button | dacPresetReadBtn/dacPresetWriteBtn | dacVoltages, dacRangeMode, range label | tab_unified.go:935-954 |
| Cell click | UnifiedTappableCanvas.Tapped() | selectedRow, selectedCol, WL (if single mode) | tab_unified.go:1023-1035 |
| Write Cell button | programBtn | arrayWeights[row][col] | tab_unified.go:1042-1073 |
| Compute MVM button | computeBtn | WL all, recompute | tab_unified.go:1087-1094 |

---

### COMPARISON View (tab_comparison.go:20-71)
**Purpose**: Compare FeFET vs GPU vs CPU architectures
**File**: tab_comparison.go:20-71
**State**: compArraySize (8, 16, 32, 64)
*(Layout unchanged from previous version)*

---

### REFERENCE View (tab_reference.go:25-53)
**Purpose**: Timing diagrams + specifications reference
**File**: tab_reference.go:25-53
**State**: timingOpSelect, specArraySizeSelect
*(Layout unchanged from previous version)*

---

## State Machine

### OpMode State Transitions
```
Initial State: OpModeRead

OpModeRead
  └─> "WRITE" button clicked -> OpModeWrite
  └─> "COMPUTE" button clicked -> OpModeCompute

OpModeWrite
  └─> "READ" button clicked -> OpModeRead
  └─> "COMPUTE" button clicked -> OpModeCompute

OpModeCompute
  └─> "READ" button clicked -> OpModeRead
  └─> "WRITE" button clicked -> OpModeWrite
```

**Actions on Mode Change**:
1. Update `deviceState.opMode`
2. Configure WL (single for READ/WRITE, all for COMPUTE)
3. Configure DAC preset (read vs write range)
4. Update mode button highlighting
5. Update WL checkboxes
6. Update DAC range label
7. Refresh array canvas
8. Update operation classification help text

---

## Key Patterns

### 1. Unified Tappable Canvas Pattern
```go
type UnifiedTappableCanvas struct {
    widget.BaseWidget
    raster *canvas.Raster
    onTap  func(row, col int)
    ca     *CircuitsApp
}

func (t *UnifiedTappableCanvas) Tapped(e *fyne.PointEvent) {
    // Convert screen coordinates to grid coordinates
    col := (int(e.Position.X) - offsetX) / cellSize
    row := (int(e.Position.Y) - offsetY) / cellSize
    t.onTap(row, col)
}
```

### 2. Material-Derived Voltage Ranges
```go
// No hardcoded values - all derived from material + config
Vc := material.CoerciveVoltage()  // Ec × thickness
readMax := calibParams.FieldMinRatio * Vc
writeMax := calibParams.FieldMaxRatio * Vc
```

### 3. Mode Button Highlighting
```go
func (ca *CircuitsApp) updateModeButtons() {
    // Reset all to low importance
    ca.modeReadBtn.Importance = widget.LowImportance
    ca.modeWriteBtn.Importance = widget.LowImportance
    ca.modeComputeBtn.Importance = widget.LowImportance

    // Highlight active mode
    switch ca.deviceState.GetOperationMode() {
    case OpModeRead:
        ca.modeReadBtn.Importance = widget.HighImportance
    case OpModeWrite:
        ca.modeWriteBtn.Importance = widget.HighImportance
    case OpModeCompute:
        ca.modeComputeBtn.Importance = widget.HighImportance
    }
}
```

### 4. Dynamic DAC Label Updates
```go
func (ca *CircuitsApp) updateDACPresetLabels() {
    readRange := ca.deviceState.GetReadRange()
    writeRange := ca.deviceState.GetWriteRange()
    ca.dacPresetReadBtn.SetText(fmt.Sprintf("Read (0-%.1fV)", readRange.Max))
    ca.dacPresetWriteBtn.SetText(fmt.Sprintf("Write (%.1f-%.1fV)", writeRange.Min, writeRange.Max))
}
```

### 5. Architecture-Aware WL Handling
```go
// Passive mode: all WLs always active (no transistor gating)
if ca.architecture == sharedwidgets.Architecture0T1R {
    ca.deviceState.SetWLAll()
}
```

---

## Thread Safety

### Mutex Protection
All shared state accessed via ca.mu (RWMutex):
- arrayWeights (read/write)
- inputVector, outputVector (read/write)
- architecture (read/write)
- animationStep, animationActive (read/write)

### DeviceState Thread Safety
DeviceState methods should be called under appropriate locking in CircuitsApp.

### Canvas Refresh Pattern
All canvas refresh calls wrapped in fyne.Do():
```go
fyne.Do(func() {
    ca.sharedArrayCanvas.Refresh()
})
```

---

## Physics Constants (Now Dynamic)

| Parameter | Source | Calculation |
|-----------|--------|-------------|
| Coercive Voltage (Vc) | material.CoerciveVoltage() | Ec × thickness |
| Read Max Voltage | physics.yaml + material | FieldMinRatio × Vc |
| Write Min Voltage | material | Vc |
| Write Max Voltage | physics.yaml + material | FieldMaxRatio × Vc |
| Max Practical Voltage | device_state.go:92 | 3.0V (hardware limit) |
| FeCIM Levels | app.go:25 | 30 |
| Default Array Size | app.go:27 | 8×8 |

---

## External Dependencies

### Fyne GUI Framework
- fyne.io/fyne/v2
- fyne.io/fyne/v2/app
- fyne.io/fyne/v2/canvas
- fyne.io/fyne/v2/container
- fyne.io/fyne/v2/layout
- fyne.io/fyne/v2/widget
- fyne.io/fyne/v2/driver/desktop (for Cursor() interface)

### Internal Packages
- fecim-lattice-tools/module4-circuits/pkg/peripherals (DAC, ADC, TIA, ChargePump)
- fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric (HZOMaterial, AllMaterials)
- fecim-lattice-tools/config/physics (Load physics.yaml config)
- fecim-lattice-tools/shared/theme (FeCIMTheme)
- fecim-lattice-tools/shared/widgets (DebugInteraction, Architecture constants)

---

## Notes

1. **Unified Device Simulation**: The OPERATIONS view is now a true device simulator. Configure WL and DAC inputs, see outputs in real-time. No artificial "modes" - the hardware is the same, only inputs differ.

2. **Material Selection**: The material selector loads all materials from `ferroelectric.AllMaterials()`. Changing material updates voltage ranges and recalculates outputs.

3. **Architecture Toggle**: Switches between PASSIVE (0T1R), 1T1R, and 2T1R. Passive mode always has all WLs active. 1T1R/2T1R draw transistor symbols.

4. **Dynamic Voltage Ranges**: All voltage thresholds are derived from physics.yaml calibration parameters and material properties. No hardcoded values. DAC preset button labels update automatically.

5. **Mode Buttons vs RadioGroup**: The new mode buttons (READ/WRITE/COMPUTE) replace the old RadioGroup. They provide better visual feedback with button importance highlighting.

6. **Calibration from physics.yaml**: The `CalibrationParams` struct loads `field_min_ratio` and `field_max_ratio` from `config/physics.yaml`. These define operating regions relative to coercive voltage.

7. **Operation Classification Help**: The `operationsModeHelp` label shows mode-specific guidance including voltage ranges and architecture-specific notes (e.g., sneak paths in passive mode).

8. **Embedded Interface**: Implements BuildContent(), Start(), Stop() for integration with main visualizer (cmd/fecim-lattice-tools).
