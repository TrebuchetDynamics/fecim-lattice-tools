# Diagnostics Remanent Staircase

## Problem Description

The Preisach Everett integral was historically implemented with a factorized-difference form:

```
E(α,β) = [tanh((α-Ec)/Δ) - tanh((β+Ec)/Δ)] * Ps/2
```

This form produces a **negative Everett integral** for minor loops within the coercive gap (α-β < 2*Ec). The polarization P then becomes negative in the Preisach integration, which is unphysical for ferroelectrics (P must remain ≥ 0 for the Landau free energy minimum).

## Root Cause

The Preisach density ρ(α,β) must be non-negative over the entire (α,β) half-plane for the Everett integral to stay ≥ 0. However, the original factorized difference form violates this by hard-clamping to zero outside the coercive window and allowing negative contributions inside.

The **product-form Everett** solves this:

```
E(α,β) = [1 + tanh((α-Ec)/Δ)] * [1 - tanh((β+Ec)/Δ)] * Ps/4
```

This ensures ρ(α,β) ≥ 0 everywhere, so P-E loops stay non-negative during minor loop excursions.

## Detection

This issue was surfaced in 2025-02 physics validation when:
- `module1-hysteresis/pkg/ferroelectric/preisach.go` material tests showed negative P for minor loop sweeps
- `shared/physics/landau.go` revealed that the Preisach integrals flipped sign at sub-coercive fields
- The golden regression data "golden_loop__Landau_Khalatnikov__Everett__preisach.json" originally captured this bug

## Verification

When `FECIM_UPDATE_PHYSICS_GOLDEN=1` is set, physics golden files regenerate and the test suite validates against the new product-form. Without this flag, the golden regression tests will fail against the old buggy factorization.

## Resolution

The Everett function was corrected from factorized-difference to product-form in:
- `module1-hysteresis/pkg/ferroelectric/preisach.go` (2025-02 commit series)
- `shared/physics/landau.go` (dxBXvVxS3N-enLandau-Everett regression fix)

The change ensures:
- Preisach integrals remain ≥ 0 for all field histories
- Minor loops produce smooth P-E transitions without staircasing
- The Preisach-Khalatnikov convergence tests pass full ensemble sweeps

## Long-Term Impact

- **Material calibration**: The correct Everett form enables accurate Preisach→Pr/Ps ratio tuning
- **ISPP convergence**: Guard-band logic now handles overshoot without false-fail transitions
- **Device simulation**: Multi-cell arrays maintain physical consistency with non-negative Everett
- **EDA export**: Preisach-based Compact Model exports use normalized Everett integrals

## Documentation References

- `shared/physics/physics_validation_test.go` — physics golden regression verification
- `module1-hysteresis/cmd/hysteresis/main.go` — CLI regression runner
- `docs/4-research/hysteresis-in-feCIM.md` — problem history documentation
- `AGENTS.md` — task coordination for this workflow