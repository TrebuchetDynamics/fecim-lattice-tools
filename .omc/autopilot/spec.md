# Module4 Device State Testing Specification

## Overview
Comprehensive unit and e2e tests for device_state.go physics and data management.

**Scope:** Test all state management, voltage calculations, ISPP, hysteresis, V/2 visualization.

---

## Test File
`module4-circuits/pkg/gui/device_state_test.go`

## Test Categories (11 categories, 45+ tests)

### 1. DeviceState Initialization (4 tests)
- TestNewDeviceState_Dimensions
- TestNewDeviceState_DefaultMode
- TestNewDeviceState_VoltageRanges
- TestNewDeviceState_NilPeripherals

### 2. Voltage Range Calculation (4 tests)
- TestUpdateVoltageRanges_ReadRange
- TestUpdateVoltageRanges_WriteRange
- TestUpdateVoltageRanges_MaterialCoerciveVoltage
- TestUpdateVoltageRanges_MaxPracticalVoltageClamp

### 3. Per-Level Voltage Calibration (3 tests)
- TestInitVoltageCalibration_LinearInterpolation
- TestGetVoltageForLevel_BoundaryClamping
- TestGetVoltageForLevel_Direction

### 4. Hysteresis Direction Tracking (3 tests)
- TestRecordWrite_AscendingDirection
- TestRecordWrite_DescendingDirection
- TestGetWriteDirection_AllCases

### 5. 4-Phase Write Sequence (4 tests)
- TestStartWriteSequence_Initialization
- TestAdvanceWritePhase_Progression
- TestAdvanceWritePhase_Completion
- TestCancelWriteSequence_Reset

### 6. ISPP State Machine (6 tests)
- TestStartISPP_Initialization
- TestISPPIterate_Verification
- TestISPPIterate_Overshoot
- TestISPPIterate_MaxIterations
- TestHandleOvershoot_Ascending
- TestHandleOvershoot_Descending

### 7. V/2 Half-Select Visualization (5 tests)
- TestEnableHalfSelectVisualization_State
- TestIsHalfSelected_TargetCell
- TestIsHalfSelected_SameRow
- TestIsHalfSelected_SameColumn
- TestDisableHalfSelectVisualization_Clear

### 8. Compute Function (5 tests)
- TestCompute_SingleRow
- TestCompute_AllRowsActive
- TestCompute_InactiveRows
- TestCompute_WithWeights
- TestCompute_Saturation

### 9. DAC Preset Modes (4 tests)
- TestSetDACPreset_ReadPreset
- TestSetDACPreset_WritePreset
- TestSetDACPreset_InputVector
- TestSetDACVoltageForState_AllLevels

### 10. Passive Mode (3 tests)
- TestSetPassiveMode_AllWLsActive
- TestSetWLSingle_IgnoredInPassiveMode
- TestSetWLCustom_IgnoredInPassiveMode

### 11. Edge Cases (6 tests)
- TestResize_SmallToLarge
- TestResize_LargeToSmall
- TestNilMaterial_FallbackBehavior
- TestComputeWithNilMaterial_Fallback
- TestComputeWithNilPeripherals_PartialOutput

---

## Test Helpers Required

```go
const testEpsilon = 1e-6

func newTestDeviceState(rows, cols int) *DeviceState
func newTestDeviceStateNilPeripherals(rows, cols int) *DeviceState
func assertVoltageInRange(t *testing.T, name string, voltage, min, max float64)
func assertFloatEquals(t *testing.T, name string, got, want float64)
func resetGlobalState()
```

## Critical: Global State Isolation

The global singletons must be reset between tests:
- voltageCalibration
- hysteresisState
- writeSequenceState
- isppState
- halfSelectState

---

**EXPANSION_COMPLETE**
