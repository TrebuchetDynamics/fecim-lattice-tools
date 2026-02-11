# Module 2 Crossbar Physics Audit (M2-P1, M2-P2)

Date: 2026-02-11  
Scope: Code audit of `module2-crossbar` against `docs/crossbar/reference/PHYSICS.md` for:
- IR drop
- Sneak paths
- Drift
- Temperature effects

## Verdict Summary

| Area | Doc claim status vs code | Evidence |
|---|---|---|
| IR drop iterative solver | ✅ Implemented as documented (iterative relaxed solve, WL/BL coupled drops, worst-case tracking) | `pkg/crossbar/nonidealities.go` (`AnalyzeIRDropIterative`), `pkg/crossbar/physics_test.go` (`TestIRDropOhmsLaw`, `TestIRDropScalesWithResistance`) |
| Sneak path model | ✅ Implemented with three-cell series paths and architecture isolation; off-diagonal-only map logic matches doc comments | `pkg/crossbar/nonidealities.go` (`AnalyzeSneakPathsWithIsolation`), `pkg/crossbar/nonidealities_test.go` (same-row/col excluded, off-diagonal paths present) |
| Drift model | ✅ Implemented with time-dependent simulator + Arrhenius scaling utilities; used in analysis paths and optional MVM approximation | `pkg/crossbar/drift.go`, `pkg/crossbar/temperature.go` (`AdjustedDriftRate`), `pkg/crossbar/physics_test.go` (`TestDriftTemperatureDependence`) |
| Temperature beyond wire resistance | ✅ Implemented and gated in MVM via `TemperatureProfile` (conductance window, variation/noise, drift scaling) | `pkg/crossbar/enhanced.go` (`MVMWithNonIdealities`), `pkg/crossbar/temperature_profile.go`, tests listed below |

## Detailed Findings

### 1) IR Drop

`PHYSICS.md` describes iterative relaxation and coupled voltage/current updates. Code in `AnalyzeIRDropIterative` follows this pattern:
- Initializes WL/BL voltages
- Iterates current solve and voltage updates with damping
- Computes `EffectiveVoltage`, `MaxIRDrop`, `AvgIRDrop`, variance, worst cell

Also, MVM path uses this IR-drop analysis (`MVMWithIRDrop`, `MVMWithNonIdealities` when enabled).

### 2) Sneak Paths

Doc describes three-cell series sneak path and architecture dependence. Code implements:
- three-cell series conductance formula in `AnalyzeSneakPathsWithIsolation`
- architecture isolation factors (0T1R/1T1R/2T1R)
- full vs simplified sneak calculations in MVM for performance scaling (`computeFullSneakCurrent`, `computeSimplifiedSneakCurrent`)

Tests confirm same-row/same-column are excluded in MVM-mode sneak map and off-diagonal paths dominate.

### 3) Drift

Drift utilities and simulator exist and are tested. In MVM path, drift is an optional approximation (gated), while full temporal drift remains a separate simulator path.

This aligns with doc framing that drift is modeled and temperature-accelerated, while keeping baseline MVM fast.

### 4) Temperature scalings beyond wire resistance (M2-P2)

Beyond wire resistance, code now supports:
- Conductance window scaling (`AdjustedConductanceRange`)
- Variation/noise scaling (`AdjustedNoise`)
- Drift-rate scaling (`AdjustedDriftRate`)

These are selectively applied in `MVMWithNonIdealities` when `TemperatureProfile.Enable` is true.

## Test Evidence Added/Updated

### Added in this task

1. `module2-crossbar/pkg/crossbar/temperature_profile_gate_test.go`
   - `TestTemperatureProfileMasterEnableGate`
   - Verifies additional temperature scalings are fully gated by `TemperatureProfile.Enable` (no accidental application when disabled).

2. `module2-crossbar/cmd/crossbar-gui/main_help_test.go`
   - `TestRunGUIHelpTextReflectsImplementedEntryPoints`
   - Locks GUI help text to current, implemented CLI entry points and avoids help drift.

### Existing relevant tests (already in tree)

- `temperature_mvm_scaling_test.go`:
  - conductance-window gating
  - high-T variation/noise impact
  - drift scaling gate and behavior
- `physics_test.go`:
  - IR-drop scaling/worst-case behavior
  - drift temperature dependence
- `nonidealities_test.go`:
  - sneak-path topology behavior

## Notes for PHYSICS.md

Current doc is broadly aligned with implementation, but some sections are intentionally model-level and contain placeholder citation markers. No blocking code/doc contradiction found for the audited areas.
