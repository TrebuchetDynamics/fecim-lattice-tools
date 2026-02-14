<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module1-hysteresis

## Purpose

Ferroelectric hysteresis simulation module. Implements P-E (polarization-electric field) curve visualization, Preisach model for memory effects with product-form Everett function, and ISPP (Incremental Step Pulse Programming) write controller for multi-level cell programming.

## Key Files

| File | Purpose |
|------|---------|
| `pkg/ferroelectric/material.go` | Re-exports HZOMaterial types from shared/physics (backward compatibility) |
| `pkg/controller/writer.go` | ISPP write controller state machine (APPLY→WAIT→VERIFY→loop) |
| `pkg/gui/gui.go` | Fyne GUI entry point for hysteresis visualization |
| `pkg/simulation/engine.go` | P-E curve simulation engine |
| `pkg/render/render.go` | Canvas rendering for hysteresis plots |

## Subdirectories

| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `cmd/hysteresis/` | Standalone CLI entry point | `main.go` |
| `pkg/algo/` | Binary search and convergence algorithms | `bsearch.go`, `convergence.go` |
| `pkg/controller/` | ISPP write controller and state machine | `writer.go` (9 test files) |
| `pkg/ferroelectric/` | Material models, Preisach model, Landau-Khalatnikov solver | Re-exports from shared/physics |
| `pkg/gui/` | Fyne GUI for hysteresis visualization | `gui.go`, `widgets/`, `simulation.go` |
| `pkg/render/` | Rendering utilities for plot canvas | `render.go` |
| `pkg/simulation/` | Simulation engine and multi-cell support | `engine.go`, `multicell.go` |
| `pkg/tui/` | Terminal UI (legacy) | TUI entry point |
| `shaders/` | GPU compute shaders (future GPU acceleration) | `.glsl` shader files |

## For AI Agents

### Working In This Directory

- **Core physics module**: Changes to ferroelectric models, Preisach, or ISPP controller affect all downstream modules (2, 3, 4).
- **Preisach model**: Uses **product-form Everett function** (NOT factorized-difference). Fix applied in commit history to prevent "teleportation" bugs.
- **ISPP controller**: State machine with 8 states (IDLE, APPLY, WAIT, VERIFY, HOLD, SUCCESS, FAILED, FORCE_RESET).
- **Guard-band logic**: Can flip sign direction. Limit to 2 max pulses to prevent catastrophic overshoot.
- **ACCEPT ±1 threshold**: Raised from 3 to 8 overshoots. Skip when guardActive=true.

### Testing Requirements

```bash
go test ./module1-hysteresis/...
go test ./module1-hysteresis/pkg/controller/...
go test ./module1-hysteresis/pkg/ferroelectric/...
```

Key test files:
- `pkg/controller/ispp_convergence_test.go` - Ensemble ISPP convergence (sensitive to ACCEPT ±1 tuning)
- `pkg/controller/writer.go` - 9 test files covering state machine, edge cases, stress
- `cmd/fecim-lattice-tools/mode_engine_matrix_test.go` - Headless ISPP tests (9 materials × 2 engines)
- `cmd/fecim-lattice-tools/mode_preisach_target_progression_test.go` - Preisach target progression validation
- Physics golden regression: `validation/testdata/physics_regression/preisach_loop_default_hzo.json`

### Common Patterns

- Material presets: `shared/physics/hzo_materials.go` (DefaultHZO, FeCIMMaterial, HZOStandard32, etc.)
- State machine: `StateApply` → `StateWait` → `StateVerify` → loop or `StateSuccess`
- Binary search convergence: `pkg/algo/bsearch.go` with bracket widening on overshoot
- Bounds reset logic: When absField < 0.01*Ec, reset to full [0, MaxField] for fresh bisection
- Overshoot handling: OvershootLimit (30) triggers StateSuccess, not StateFailure

### Critical Bug Patterns (Do Not Re-Introduce)

1. **Guard Sign Direction Flip**: Guard pulses can reverse search direction. Fixed: Limit to 2 max, clamp calcLevel.
2. **Bounds Collapse**: Binary search brackets [VMin, VMax] can collapse. Fixed: Widen minimally using direction info.
3. **ACCEPT ±1 Guard Interaction**: Guard sets absErr=1 even when actual error=0. Fixed: Skip ACCEPT ±1 when guardActive=true.
4. **Zero-Field Bounds Reset**: Setting VMax=0 causes catastrophic collapse. Fixed: Reset bounds to full range when absField < 0.01*Ec.
5. **Preisach Everett Zero-Clamp**: Factorized-difference form goes negative. Fixed: Use product-form Everett (mathematically correct).

## Dependencies

### Internal

- `shared/physics/` - Landau-Khalatnikov solver, HZOMaterial definitions
- `shared/widgets/` - GUI components (sliders, charts, buttons)
- `shared/theme/` - Styling and color scheme
- `shared/logging/` - Structured logging

### External

- Fyne v2 (`fyne.io/fyne/v2`) - Cross-platform native GUI
- gonum (`gonum.org/v1/gonum`) - Numerical computation (matrix operations, special functions)

## Configuration

Material parameters are defined in `shared/physics/hzo_materials.go`:
- `DefaultHZO()` - Typical Si-doped HfO2 (Hf0.5Zr0.5O2)
- `FeCIMMaterial()` - FeCIM specifications
- `HZOStandard32()` - 32 analog states (conference demonstration)
- Custom presets: `HZOCustom14()`, `CryogenicHZO()`, `HZOFJT140()`

ISPP configuration in `WriteController`:
- `NumLevels` - Target quantization levels (1-30)
- `MaxRetries` - Max pulses before reset (prevents infinite loops)
- `OvershootLimit` - Max consecutive overshoots (30 is typical)
- `MinStep` - Hard lower bound on voltage step size

## Performance Notes

- ISPP convergence: ~5-20 pulses per level (material-dependent)
- Preisach curve computation: ~1-2ms per point (depends on material sharpness)
- GUI update rate: ~30-60 Hz (synced to Fyne render cycle)
- Memory: ~10-50MB for typical plots (16×16 crossbar + history)

<!-- MANUAL: -->
