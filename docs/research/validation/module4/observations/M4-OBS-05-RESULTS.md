# M4-OBS-05 — 0T1R passive behavior validation + disclosure

Date: 2026-02-12

## Scope checked
- `module4-circuits/pkg/gui/device_state.go`
- `module4-circuits/pkg/gui/tab_unified_voltage.go`
- `module2-crossbar/pkg/crossbar/enhanced.go`
- `module4-circuits/pkg/arraysim/tier_a.go`

## Findings
1. **0T1R V/2 write biasing is modeled in `ApplyHalfSelectWrite`**:
   - target WL gets `+Vwrite/2`, target BL gets `-Vwrite/2`
   - target cell sees full `Vwrite`
   - same-row and same-col cells see `Vwrite/2`
   - diagonal/unselected cells see `0V`

2. **`applyHalfSelectDisturb` coverage**:
   - Iterates all cells; updates based on effective `Vcell = WL[r]-BL[c]` in passive mode.
   - Residue accumulation is applied to half-selected neighbors (same row XOR same col).
   - Row/col neighbors are therefore all covered for disturb residue.

3. **Disturb magnitude**:
   - Electrical half-select exposure is `Vwrite/2` at device-state level.
   - Disturb residue increment is a simplified UI/test hook (`0.01 * halfSelectDisturbRate`), not a direct numeric `V/2` transfer.

4. **Sneak current behavior**:
   - `module2-crossbar/pkg/crossbar/enhanced.go` models 0T1R sneak paths through multi-cell paths (`computeFullSneakCurrent`) and simplified scaling for larger arrays.
   - 1T1R/2T1R paths are strongly suppressed by architecture isolation factors.

## Architecture truth table

| Architecture | Target Cell | Same-Row Cells | Same-Col Cells | Other Cells | Sneak Current |
|-------------|-------------|----------------|----------------|-------------|---------------|
| 0T1R | Full V_write | V/2 disturb | V/2 disturb | 0V | Yes, through all paths |
| 1T1R | Full V_write | Suppressed (Vth) | Suppressed (Vth) | Isolated | Minimal (Ioff) |
| 2T1R | Full V_write | Isolated | Isolated | Isolated | None |

## Changes made
- Added UI disclosure text: **"0T1R model: simplified half-select (row+col neighbors only)"**.
- Added FEATURES section-4 note + truth table.
- Added test: `module4-circuits/pkg/gui/passive_behavior_test.go`.
