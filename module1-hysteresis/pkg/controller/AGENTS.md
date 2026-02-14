<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module1-hysteresis/pkg/controller

## Purpose

Implements the ISPP (Incremental Step Pulse Programming) write controller state machine (APPLY→WAIT→VERIFY→loop). This package manages ferroelectric device programming with closed-loop binary search on voltage, convergence detection, overshoot recovery, and guard-band correction. Core logic handles Landau-Khalatnikov physics integration and state transitions.

## Key Files

| File | Description |
|------|-------------|
| `writer.go` | Main ISPP WriteController state machine with binary search convergence (33KB). Guard-band logic, overshoot tracking, accept-±1 logic. |
| `ispp_convergence_test.go` | Integration test for convergence behavior across 9 materials × 2 engines. Sensitive to accept-±1 threshold. |
| `ispp_full_cycle_test.go` | Full ISPP cycle validation (APPLY→WAIT→VERIFY→loop→SUCCESS). |
| `writer_stress_test.go` | Stress tests: 1000+ random level targets, edge cases, bounds collapse, overshoot recovery. |
| `writer_extended_test.go` | Extended controller tests: remanent sweep, LK tuning, guard-band direction. |

## For AI Agents

### Working In This Directory

**Critical Bug Patterns (READ FIRST):**

1. **Guard-Band Direction Flip**: Guard logic can override `LastError=0` to `±1` when at target level. If `guardSign` flips direction (ascending→descending), causes catastrophic overshoot. **Fix**: Limit guard pulses to 2 max, clamp `calcLevel` to prevent direction flip.

2. **Bounds Collapse**: Binary search bounds `[VMin, VMax]` can collapse (`VMin >= VMax`) after overshoot recovery. Old code reset to full range `[0, MaxField]`, losing convergence progress. **Fix**: Widen minimally using direction info (`needMore`/`needLess`).

3. **ACCEPT ±1 Guard Interaction**: Accept ±1 logic (accept level within ±1 of target after overshoots) fires prematurely when guard is active. **Fix**: (1) Skip ACCEPT ±1 when `guardActive=true` (error is 0, guard inflated it); (2) Raise threshold from 3 to 8 overshoots so natural convergence finishes first.

4. **Zero-Field Bounds Reset**: During verify after reset shortcut (`nextDir==0`), `CurrentField=0` causes bounds collapse to `[0,0]×Ec`. **Fix**: When `absField < 0.01*Ec`, reset bounds to full `[0, MaxField]` for fresh bisection.

5. **Overshoot Limit as Physics-Limited Convergence**: Materials with sharp switching (fecim_hzo, hzo_custom_14) can't maintain mid-range levels at E=0. Repeated overshoots prove controller bracketed target—it's a physics limitation. **Fix**: `OvershootLimit` (30) triggers `StateSuccess` not `StateFailed`.

**Working on WriteController:**

- Read `writer.go` fully first (state machine structure, phase transitions, error calculations)
- Understand binary search bracketing: `VMin` = "not enough voltage", `VMax` = "too much voltage"
- Guard pulses are temporary voltage nudges during VERIFY; max 2 to prevent direction flip
- `OvershootLimit=30` transitions to SUCCESS (not FAILED) for physics-limited materials
- State transitions are ordered: IDLE→APPLY→WAIT→VERIFY→(loop back to APPLY or SUCCESS)
- Tests in `ispp_convergence_test.go` are regression-critical; rerun after changes to `writer.go`

### Testing Requirements

```bash
# Run all controller tests
go test ./module1-hysteresis/pkg/controller -v

# Run convergence ensemble test (sensitive to accept-±1 threshold)
go test ./module1-hysteresis/pkg/controller -run TestISPPConverges_LandauK_Ensemble_Superlattice -v

# Run full ISPP cycle test
go test ./module1-hysteresis/pkg/controller -run TestISPPFullCycle -v

# Run stress tests (slow, ~30s)
go test ./module1-hysteresis/pkg/controller -run TestStress -v

# Build headless engine test (mode_engine_matrix_test.go in cmd/)
go test ./cmd/fecim-lattice-tools -run TestISPPEngineMatrix -v
```

### Common Patterns

- **Phase transitions**: Check `currentPhase` variable and `WriteState` enum values
- **Binary search**: VMin/VMax updates happen in VERIFY phase; use `needMore`/`needLess` signals
- **Guard pulses**: Only during VERIFY when `|error| <= 1`; max 2 pulses, clamp direction
- **Convergence**: `lastError==0` AND no overshoots in last N iterations = SUCCESS
- **Overshoot detection**: `|currentLevel - targetLevel| > 1` after APPLY pulse
- **Reset shortcut**: When `overshoots >= 3`, attempt rapid reset (set `nextDir=0`) to recover bounds

## Dependencies

### Internal

- `module1-hysteresis/pkg/algo` - Generic binary search, convergence utilities
- `shared/physics` - Landau-Khalatnikov LK solver, material models, quantization
- `shared/logging` - Package-level logging (lazy-initialized)

### External

- `log` (Go stdlib) - Info/error logging
- `math` (Go stdlib) - Floating-point operations

<!-- MANUAL: Last edited 2026-02-13. Critical for hysteresis module ISPP convergence. -->
