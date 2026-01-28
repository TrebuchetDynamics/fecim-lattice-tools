# Module4 Voltage Rules Implementation Specification

## Overview
Implement voltage rules from VOLTAGE_RULES.md into the module4-circuits GUI.

**Scope:** Visual demo of voltage concepts, not physics-accurate simulation.

---

## Requirements Summary

### Functional Requirements
1. **Multi-Level Write Voltage Calibration** - Per-level voltage arrays (30 levels)
2. **Program-Verify Loop (ISPP)** - Write→read→verify→adjust→retry (max 5 iterations)
3. **V/2 Half-Select Biasing** - Visual indication for passive (0T1R) mode
4. **4-Phase Write Sequence** - RESET→HOLD→WRITE→HOLD animation
5. **Architecture-Specific Voltage UI** - Different displays for 0T1R vs 1T1R/2T1R
6. **Hysteresis Path Awareness** - Track ascending vs descending direction

### Non-Functional Requirements
- 20 FPS animation (50ms frame delay)
- Thread-safe state updates (use existing ca.mu mutex)
- Responsive UI (no blocking operations)

### Out of Scope
- Full Preisach model integration
- Temperature-dependent calibration
- Physics-accurate switching simulation

---

## Design Simplifications

This specification intentionally uses simplified models for demonstration purposes:

### Linear Calibration Curve
**This implementation uses LINEAR interpolation as a simplified demo.** The actual physics (non-linear Preisach model with domain-by-domain switching, temperature dependence, and history-dependent behavior) is out of scope per design intent.

The linear calibration maps:
- Level 0 → WriteRange.Min voltage
- Level 29 → WriteRange.Max voltage
- Intermediate levels → Linear interpolation between min and max

This is sufficient to demonstrate the UI concepts (4-phase timing, ISPP iteration, V/2 visualization) without requiring physics-accurate switching models.

---

## Existing Code Integration

### Relationship with `writeReadVerifyLoop()`

The existing `writeReadVerifyLoop()` method at `tab_unified.go:1186-1292` already implements basic ISPP functionality:
- `maxIterations = 5`
- Write → Read → Verify → Adjust voltage loop
- Step-based level adjustment (1-2 levels per pulse)
- Voltage adjustment for undershoot/overshoot
- 300ms iteration delay with UI status updates

**Integration Approach: ENHANCE (not replace)**

The new `runISPPWithAnimation()` method will:
1. **Call into** the new state machine methods (`StartISPP()`, `ISPPIterate()`, etc.)
2. **Preserve** the existing `writeReadVerifyLoop()` as a fallback/legacy path
3. **Add** new capabilities on top:
   - 4-phase sequence animation within each ISPP iteration
   - Calibrated per-level voltage lookup
   - Hysteresis direction tracking
   - V/2 visualization for 0T1R mode

The existing code remains functional. New features are additive.

---

## Overshoot Handling

### The Overshoot Problem
When writing to a target level, the physical device may overshoot (e.g., targeting level 15, but reaching level 17). In real FeCIM devices, you cannot simply "decrease" the level with a reverse pulse - you must RESET to saturation and re-approach.

### `HandleOvershoot(row, col int)` Method Specification

```go
// HandleOvershoot performs RESET-to-saturation when write overshoots target
// Returns true if reset was performed, false if no overshoot detected
func (ds *DeviceState) HandleOvershoot(row, col int) bool
```

**Behavior:**
- **Ascending direction** (writing to higher level):
  - If `currentLevel > targetLevel`: Overshoot detected
  - Action: RESET to level 0 (negative saturation)
  - Then restart ISPP from level 0 ascending to target

- **Descending direction** (writing to lower level):
  - If `currentLevel < targetLevel`: Overshoot detected
  - Action: RESET to level 29 (positive saturation)
  - Then restart ISPP from level 29 descending to target

**Integration with ISPP:**
```go
func (ds *DeviceState) ISPPIterate() ISPPResult {
    // ... perform write pulse ...
    // ... read back currentLevel ...

    if ds.isppState.Direction == DirectionAscending && currentLevel > targetLevel {
        ds.HandleOvershoot(row, col)
        return ISPPResultOvershoot
    }
    if ds.isppState.Direction == DirectionDescending && currentLevel < targetLevel {
        ds.HandleOvershoot(row, col)
        return ISPPResultOvershoot
    }
    // ... continue normal ISPP ...
}
```

---

## Integration Flow

### Write Operation Call Sequence

When WRITE mode is active and user clicks "Write Cell":

```
User clicks "Write Cell" button
    │
    ▼
onUnifiedWrite() [existing entry point]
    │
    ├─► Validate selection (row, col, targetLevel)
    ├─► Save undo history
    │
    ▼
[NEW] Check architecture mode
    │
    ├─► If 0T1R (passive): EnableHalfSelectVisualization(row, col, voltage)
    │
    ▼
[NEW] Get calibrated voltage
    │
    └─► voltage = GetVoltageForLevel(targetLevel, direction)
    │
    ▼
[NEW] Start 4-phase sequence (if animation enabled)
    │
    └─► StartWriteSequence(row, col, targetLevel)
    │
    ▼
go runISPPWithAnimation(row, col, targetLevel, voltage) [goroutine]
    │
    ├─► StartISPP(row, col, targetLevel)
    │
    └─► Loop (max 5 iterations):
        │
        ├─► AdvanceWritePhase() through RESET→HOLD1→WRITE→HOLD2
        ├─► Update UI via updateWriteSequenceUI()
        ├─► ISPPIterate() - perform write and verify
        ├─► Check for overshoot → HandleOvershoot() if needed
        ├─► updateISPPUI() - show iteration status
        │
        └─► If verified: break
    │
    ▼
[NEW] DisableHalfSelectVisualization()
    │
    ▼
RecordWrite(row, col, finalLevel) - update hysteresis state
```

### V/2 Activation Trigger

**When does `EnableHalfSelectVisualization()` get called?**

Automatically when ALL of:
1. Architecture is 0T1R (passive crossbar)
2. Mode is WRITE
3. A write operation begins (via "Write Cell" button or ISPP start)

```go
// In onUnifiedWrite() or runISPPWithAnimation():
if ds.GetArchitectureType() == Architecture0T1R && ds.GetOperationMode() == ModeWrite {
    ds.EnableHalfSelectVisualization(row, col, writeVoltage)
}
```

**Disable trigger:**
- When write operation completes (success or max iterations)
- When mode changes away from WRITE
- When architecture changes away from 0T1R
- When user cancels operation

---

## Visual Acceptance Criteria

### What the User Should SEE

#### 4-Phase Timing Diagram
- **Location:** Right side panel during write operation
- **Content:**
  - 4 labeled phases: "RESET", "HOLD", "WRITE", "HOLD"
  - Voltage waveform showing step changes between phases
  - Current phase highlighted (bright) vs completed (dim)
  - Phase duration labels (e.g., "100ns", "50ns", "200ns", "50ns")
- **Animation:** Phase highlight moves left-to-right as sequence progresses

#### V/2 Half-Select Overlay
- **When visible:** Only when architecture=0T1R AND mode=WRITE AND operation in progress
- **Target cell:** Colored in Bright Gold (255, 200, 0)
- **Half-selected row cells:** Colored in Amber (255, 165, 0) - same row, different columns
- **Half-selected column cells:** Colored in Amber (255, 165, 0) - same column, different rows
- **Non-selected cells:** Remain in normal color (not highlighted)
- **Label:** "V/2 Bias Active" indicator visible during overlay

#### ISPP Status Display
- **Location:** Status area below array visualization
- **Content:**
  - Iteration counter: "Iteration 3/5"
  - Current level: "Current: Level 12"
  - Target level: "Target: Level 15"
  - Voltage being applied: "V = 2.35V"
  - Direction indicator: "↑ Ascending" (green) or "↓ Descending" (red)
- **On overshoot:** Display "OVERSHOOT - Resetting to saturation..."
- **On success:** Display "SUCCESS - Target reached in N iterations"
- **On max iterations:** Display "PARTIAL - Reached level X (target was Y)"

#### Hysteresis Direction Indicator
- **Location:** Near cell selection or in write panel
- **Ascending (level increasing):** Green arrow pointing up (↑) with "Ascending" label
- **Descending (level decreasing):** Red arrow pointing down (↓) with "Descending" label
- **Updates:** Automatically when target level is selected

#### Architecture-Specific Panels
- **0T1R mode:** Shows "Passive Crossbar Voltage Panel" with V/2 bias information
- **1T1R/2T1R mode:** Shows "Active Transistor Voltage Panel" with direct cell access info
- **Panel switches:** Automatically when architecture dropdown changes

---

## Technical Specification

### Target Files
1. `module4-circuits/pkg/gui/device_state.go` - State management (~250-300 lines added)
2. `module4-circuits/pkg/gui/tab_unified.go` - UI and animation (~400-500 lines added)
3. `module4-circuits/pkg/gui/app.go` - New widget fields (~20 lines added)

### New Data Structures

**device_state.go:**
```go
// WritePhase enum: PhaseIdle, PhaseReset, PhaseHold1, PhaseWrite, PhaseHold2
// WriteSequenceState: Active, Phase, TargetRow/Col/Level, CurrentLevel, Timing, Progress
// ISPPState: Active, Iteration, MaxIterations, TargetLevel, CurrentLevel, Voltages, Verified, Direction
// ISPPResult enum: ISPPResultContinue, ISPPResultVerified, ISPPResultOvershoot, ISPPResultMaxIterations
// HysteresisDirection enum: DirectionUnknown, DirectionAscending, DirectionDescending
// HysteresisState: LastWrittenLevel, Direction (per cell)
// PerLevelVoltageCalibration: AscendingVoltages[30], DescendingVoltages[30]
// HalfSelectVisualization: Enabled, FullVoltage, HalfVoltage, Selected/HalfSelect rows/cols
```

### New Methods

**device_state.go (16 methods):**
- `InitVoltageCalibration()` - Linear interpolation per level
- `GetVoltageForLevel(level, direction)` - Calibrated voltage lookup
- `StartWriteSequence(row, col, targetLevel)` - Begin 4-phase sequence
- `AdvanceWritePhase()` - Move to next phase
- `GetWritePhaseInfo()` - Current phase for UI
- `CancelWriteSequence()` - Abort sequence
- `StartISPP(row, col, targetLevel)` - Begin ISPP loop
- `ISPPIterate()` - One write-verify iteration (returns ISPPResult)
- `HandleOvershoot(row, col)` - RESET to saturation on overshoot (see Overshoot Handling section)
- `GetISPPStatus()` - Current ISPP state for UI
- `CancelISPP()` - Abort ISPP
- `RecordWrite(row, col, newLevel)` - Update hysteresis state
- `GetWriteDirection(row, col, current, target)` - Determine direction
- `EnableHalfSelectVisualization(row, col, voltage)` - Enable V/2 overlay
- `DisableHalfSelectVisualization()` - Disable V/2 overlay
- `GetHalfSelectState()` - Current V/2 state

**tab_unified.go (12 methods):**
- `drawWriteSequenceTimingDiagram()` - 4-phase timing diagram
- `animateWriteSequence()` - Run 4-phase animation
- `updateWriteSequenceUI()` - Refresh write display
- `runISPPWithAnimation()` - ISPP with visual feedback
- `updateISPPUI()` - Refresh ISPP display
- `drawHalfSelectOverlay()` - V/2 overlay on array canvas
- `updateHalfSelectVisualization()` - Enable/disable V/2
- `createPassiveVoltagePanel()` - V/2 panel for 0T1R
- `createActiveVoltagePanel()` - Direct panel for 1T1R/2T1R
- `updateArchitectureSpecificUI()` - Show/hide panels
- `updateHysteresisDirectionUI()` - Direction indicator
- `createEnhancedWriteModePanel()` - Replace existing write panel

### Constants

| Constant | Value |
|----------|-------|
| PhaseResetDurationNs | 100 |
| PhaseHold1DurationNs | 50 |
| PhaseWriteDurationNs | 200 |
| PhaseHold2DurationNs | 50 |
| ISPPMaxIterations | 5 |
| ISPPToleranceLevels | 0 (exact match) |
| HalfSelectVoltageRatio | 0.5 |
| AnimationFrameDelayMs | 50 |

### UI Color Scheme

| Element | Color |
|---------|-------|
| Full write voltage | Bright Gold (255, 200, 0) |
| V/2 half-select | Amber (255, 165, 0) |
| Zero voltage | Dim Gray (50, 50, 60) |
| Ascending direction | Green (100, 220, 120) |
| Descending direction | Red (220, 100, 100) |

---

## Implementation Priority

1. Per-Level Voltage Calibration (Low complexity)
2. Hysteresis Direction Tracking (Low)
3. 4-Phase Write Sequence State (Medium)
4. 4-Phase Animation UI (Medium)
5. ISPP State Machine (Medium)
6. ISPP Animation UI (Medium)
7. V/2 Visualization State (Low)
8. V/2 Overlay Drawing (Medium)
9. Architecture-Specific Panels (Low)

---

**EXPANSION_COMPLETE**
