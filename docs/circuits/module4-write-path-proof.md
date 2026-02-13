# Module 4 WRITE/ISPP Shared-Physics Proof

Date: 2026-02-13

## Result
Module 4 operational write dispatch is unified with `shared/physics`.

## Dispatch chain
- `tab_unified_actions.go:onUnifiedProgram()`
  - launches `runISPPWithAnimation(selectedRow, selectedCol, targetLevel)`.
- `tab_unified_voltage.go:runISPPWithAnimation()`
  - level-engine path uses `DeviceState.StartISPP` / `ISPPIterate`.
  - LK-engine path uses `runISPPWithLK()`.

## Shared physics usage evidence

### Level-engine ISPP path
- `device_state.go:StartISPP()`:
  - `sharedphysics.GetDirection(...)`
  - `sharedphysics.ISPPCalculator.CalculateStartVoltage(...)`
- `device_state.go:ISPPIterate()`:
  - `sharedphysics.ISPPCalculator.CheckResult(...)`
  - `sharedphysics.ISPPCalculator.CalculateNextVoltage(...)`
- `device_state.go:programLevelFromCoupledVoltage()`:
  - `sharedphysics.NewLKSolver()`
  - `sharedphysics.ConductanceToPolarization(...)`
  - `sharedphysics.PolarizationToConductanceWithParams(...)`

### LK-engine ISPP path
- `tab_unified_voltage.go:runISPPWithLK()`:
  - `sharedphysics.NewLKSolver()`
  - `sharedphysics.NewWriteController(...)`
  - `WriteController.WriteTargetWithReset(...)`

### Conductance model source
- `device_state.go:levelToConductance(...)` delegates to `shared/physics` material model (`HZOMaterial.DiscreteLevel`) and geometry scaling.

## Divergent shortcut check
- `tab_unified_actions.go:writeReadVerifyLoop(...)` contains legacy heuristic stepping.
- No runtime call site found from current UI dispatch path.

Conclusion: active write/ISPP path is shared-physics-backed; legacy helper remains non-dispatched.
