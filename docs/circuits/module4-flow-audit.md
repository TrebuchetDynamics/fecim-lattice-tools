# Module 4 Circuits Flow Audit (READ / COMPUTE / WRITE)

Audit date: 2026-02-13  
Scope: `module4-circuits/pkg/gui`, `module4-circuits/pkg/arraysim`, `shared/physics`

## Execution map by operation

### READ

1. **Where voltage is applied**
   - `module4-circuits/pkg/gui/tab_unified.go:setOperationMode(OpModeRead)`
     - Grounds all BLs, then applies read voltage only on selected BL.
   - `module4-circuits/pkg/gui/tab_unified_actions.go:onUnifiedRead()`
     - Re-applies selected-row/selected-column read bias (`readRange.Max*0.4`, clamped).

2. **Where currents are solved**
   - `module4-circuits/pkg/gui/tab_unified.go:recomputeAndRefreshNow()`
     - Calls `deviceState.Compute(weights, levels)`.
   - `module4-circuits/pkg/gui/device_state.go:Compute()`
     - Coupled path: `computeWithArraysimLocked()`.
     - Fallback ideal path: `computeIdealLocked()`.
   - `module4-circuits/pkg/gui/device_state.go:computeWithArraysimLocked()`
     - Uses `arraysim.Engine.Solve` (default Tier-A).
   - `module4-circuits/pkg/arraysim/tier_a.go:Solve()` -> `referenceSolveDense(...)`
     - Full WL/BL resistive nodal solve.

3. **Where cell state is read/updated**
   - Read access:
     - `tab_unified_actions.go:onUnifiedRead()` reads `ca.arrayWeights[row][col]` for reporting.
     - `device_state.go:computeWithArraysimLocked()` reads weights -> conductance.
   - No weight update in READ path.

4. **Physics model used**
   - Level->conductance: `device_state.go:levelToConductance()` using `shared/physics.HZOMaterial.DiscreteLevel` + geometry scaling.
   - Coupled currents/voltages: `arraysim/referenceSolveDense` KCL nodal DC network.
   - Sense chain: `arraysim/sensechain.go` via `convertSenseLocked()` (TIA+ADC).

---

### COMPUTE (MVM)

1. **Where voltage is applied**
   - `tab_unified.go:setOperationMode(OpModeCompute)`
     - Enables all rows (unless passive already always-on), maps input vector via `SetDACPreset(DACInputVector)`.
   - `tab_unified_actions.go:onUnifiedCompute()`
     - Forces WL all-on and reapplies DAC input vector.

2. **Where currents are solved**
   - Same compute pipeline as READ:
     - `recomputeAndRefreshNow()` -> `deviceState.Compute()` -> `computeWithArraysimLocked()` / `computeIdealLocked()`.

3. **Where cell state is read/updated**
   - Reads `arrayWeights` to compute conductance.
   - No update to `arrayWeights` in compute flow.

4. **Physics model used**
   - Ohmic branch current `I = G * V` at per-cell level.
   - Crossbar coupling by arraysim nodal solve (Tier-A default, Tier-B optional).
   - Material-aware conductance and geometry scaling from shared physics.

---

### WRITE (Program Cell / ISPP)

1. **Where voltage is applied**
   - Entry: `tab_unified_actions.go:onUnifiedProgram()` -> goroutine `runISPPWithAnimation()`.
   - `tab_unified_voltage.go:applyWriteVoltages(row,col,targetV)`
     - Converts target pulse via DAC quantization (`DeviceState.DACWriteVoltage`).
     - Passive 0T1R: `ApplyHalfSelectWrite` (V/2 scheme on WL/BL).
     - Active 1T1R/2T1R: selected BL driven, others grounded.
   - Phase-level sequencing: `applyWritePhaseVoltages(...)`.

2. **Where currents are solved**
   - Every pulse/verify recomputes through normal compute path:
     - `recomputeAndRefresh()` -> `deviceState.Compute()` -> arraysim solve / ideal fallback.
   - Effective per-cell write voltage for target update comes from
     - `DeviceState.GetEffectiveCellVoltage(row,col)` (coupled voltage when available).

3. **Where cell state is read/updated**
   - Read current state: `runISPPWithAnimation()` reads `ca.arrayWeights[row][col]`.
   - Update target cell:
     - `DeviceState.programLevelFromCoupledVoltage(...)` returns next level.
     - Then writes `ca.arrayWeights[row][col] = nextLevel`.
   - Track hysteresis bookkeeping:
     - `DeviceState.RecordWrite(row,col,finalLevel)`.
   - Half-select neighbor disturb modeling:
     - `tab_unified_voltage.go:applyHalfSelectDisturb(...)` may update non-target cells in passive mode.

4. **Physics model used**
   - ISPP policy math: shared `shared/physics.ISPPCalculator` (`StartISPP`, `ISPPIterate`, overshoot handling).
   - Per-pulse polarization update:
     - `shared/physics.LKSolver` + conductance-polarization transfer (`ConductanceToPolarization`, `PolarizationToConductanceWithParams`).
   - Optional full physics engine:
     - `runISPPWithLK()` + `shared/physics.WriteController`.

---

## Task 2 proof: unified write/ISPP path status

### Finding
Primary Module-4 WRITE path is unified to shared physics (no private voltage-step math in active path):

- Shared ISPP direction and stepping:
  - `device_state.go:StartISPP()` uses `sharedphysics.GetDirection` + `ISPPCalculator.CalculateStartVoltage`.
  - `device_state.go:ISPPIterate()` uses `ISPPCalculator.CheckResult` + `CalculateNextVoltage`.
- Shared ferroelectric dynamics and conductance mapping:
  - `programLevelFromCoupledVoltage()` uses `sharedphysics.NewLKSolver`, `ConductanceToPolarization`, `PolarizationToConductanceWithParams`.
- Shared advanced controller path:
  - `tab_unified_voltage.go:runISPPWithLK()` uses `sharedphysics.NewWriteController`.

### Divergent shortcut check
- `tab_unified_actions.go:writeReadVerifyLoop(...)` contains legacy heuristic level stepping, but no call sites were found (dead helper).
- Active UI program flow (`onUnifiedProgram`) dispatches to `runISPPWithAnimation`/`runISPPWithLK`, both shared-physics-backed.

Conclusion: **operational WRITE path is unified with `shared/physics/`; legacy divergent helper exists but is not on the dispatch path.**
