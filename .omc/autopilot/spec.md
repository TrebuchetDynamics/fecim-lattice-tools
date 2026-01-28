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
// ISPPState: Active, Iteration, MaxIterations, TargetLevel, CurrentLevel, Voltages, Verified
// HysteresisDirection enum: DirectionUnknown, DirectionAscending, DirectionDescending
// HysteresisState: LastWrittenLevel, Direction (per cell)
// PerLevelVoltageCalibration: AscendingVoltages[30], DescendingVoltages[30]
// HalfSelectVisualization: Enabled, FullVoltage, HalfVoltage, Selected/HalfSelect rows/cols
```

### New Methods

**device_state.go (15 methods):**
- `InitVoltageCalibration()` - Linear interpolation per level
- `GetVoltageForLevel(level, direction)` - Calibrated voltage lookup
- `StartWriteSequence(row, col, targetLevel)` - Begin 4-phase sequence
- `AdvanceWritePhase()` - Move to next phase
- `GetWritePhaseInfo()` - Current phase for UI
- `CancelWriteSequence()` - Abort sequence
- `StartISPP(row, col, targetLevel)` - Begin ISPP loop
- `ISPPIterate()` - One write-verify iteration
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
