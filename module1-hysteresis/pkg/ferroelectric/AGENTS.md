<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module1-hysteresis/pkg/ferroelectric

## Purpose

Provides material models and hysteresis physics for ferroelectric simulation. Contains Preisach model (with product-form Everett function), material parameter definitions (Pr, Ps, Ec, delta), calibration utilities for matching literature data, and rendering functions for P-E curves. Most core implementations re-export from `shared/physics` for backward compatibility; this package is the module1 facade.

## Key Files

| File | Description |
|------|-------------|
| `material.go` | Re-export of `shared/physics.HZOMaterial` factory functions (DefaultHZO, FeCIMMaterial, LiteratureSuperlattice, etc.). Backward compatibility layer. |
| `preisach.go` | Preisach model with product-form Everett function (9.6KB). Calibration helpers for tuning Delta to match Pr/Ps ratio. |
| `render.go` | P-E curve rendering: canvas drawing, hysteresis loop visualization. |
| `golden_regression_test.go` | Golden file regression test for preisach loop shape (requires `FECIM_UPDATE_PHYSICS_GOLDEN=1` to regenerate). |
| `preisach_test.go` | Comprehensive Preisach model tests: Everett function, minor loops, loop closure. |
| `physics_validation_test.go` | Physics validation against literature: Pr/Ps ratio, coercivity, remanence. |

## For AI Agents

### Working In This Directory

**Critical Physics Fix (Everett Function):**

The Preisach model's Everett function was replaced from factorized-difference form to product form:

- **Old (incorrect)**: `[tanh((α-Ec)/Δ) - tanh((β+Ec)/Δ)] * Ps/2`
  - Goes negative for minor loops within coercive gap (α-β < 2*Ec)
  - Hard-clamped to 0, making all sub-coercive ISPP invisible
  - P stayed frozen during PREP and ISPP, then jumped discontinuously during wipeout

- **New (correct)**: `[1+tanh((α-Ec)/Δ)] * [1-tanh((β+Ec)/Δ)] * Ps/4`
  - Product form is the mathematically correct integral of sech² Preisach density
  - Always non-negative
  - Major loop shape and Pr/Ps ratio are identical
  - Minor loops now produce smooth P changes

**Golden Regression Data:**

- Regenerated when Everett function was fixed
- Location: `validation/testdata/physics_regression/preisach_loop_default_hzo.json`
- Regenerate via: `FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./module1-hysteresis/pkg/ferroelectric -run TestGoldenRegression`
- Do NOT commit new golden files without understanding what changed

**Working on Preisach Model:**

- Read `preisach.go` first to understand Everett calculation
- Material tuning (Delta) should keep Pr/Ps ratio consistent
- Test minor loops thoroughly; they're critical for ISPP convergence
- If P jumps discontinuously during ISPP, check Everett function boundary behavior
- Product-form Everett is stable; factorized-difference form is a historical bug

### Testing Requirements

```bash
# Run all ferroelectric tests
go test ./module1-hysteresis/pkg/ferroelectric -v

# Run Preisach model tests
go test ./module1-hysteresis/pkg/ferroelectric -run TestPreisach -v

# Run golden regression test
go test ./module1-hysteresis/pkg/ferroelectric -run TestGoldenRegression -v

# Regenerate golden regression data (if Everett function changed)
FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./module1-hysteresis/pkg/ferroelectric -run TestGoldenRegression -v

# Run physics validation (slow, compares against literature)
go test ./module1-hysteresis/pkg/ferroelectric -run TestPhysicsValidation -v

# Run hysteresis loop tests
go test ./module1-hysteresis/pkg/ferroelectric -run TestHysteresisLoop -v
```

### Common Patterns

- **Material factory**: Use `DefaultHZO()` for literature-backed baseline behavior, `FeCIMMaterial()` only with conference-baseline simulation-assumption caveats, `LiteratureSuperlattice()` for cited superlattice scenario exploration
- **Preisach tuning**: `tuneDeltaForPr()` adjusts Delta to match target Pr while keeping Ps fixed
- **Loop rendering**: `render.go` provides canvas-based P-E curve visualization
- **Everett calculation**: Always use product form; never revert to factorized-difference
- **Test regression**: Golden file tests ensure loop shape stability across refactoring

## Dependencies

### Internal

- `shared/physics` - Core HZOMaterial, LKSolver, Preisach Everett adapter, quantization
- `shared/logging` - Package-level logging for Preisach calibration
- `module1-hysteresis/pkg/algo` - Generic utilities

### External

- `math` (Go stdlib) - Transcendental functions (tanh, exp, sqrt)

<!-- MANUAL: Last edited 2026-02-13. Everett function fix is stable; no further changes expected. -->
