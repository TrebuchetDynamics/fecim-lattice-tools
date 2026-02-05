# Landau-Khalatnikov (L-K) Hysteresis Model (Module 1)

This document covers L-K specific notes for the hysteresis simulator. For the full equation and solver details,
see `docs/hysteresis/HYSTERESIS-equation.md`.

**Discrete Level Mapping (Shared with Preisach)**
- Discrete MLC levels are **evenly spaced in polarization**, not electric field.
- `effectivePs = Ps * rangeFrac`.
- `rangeFrac` comes from `TargetRangeFrac` in the material (defaults to `0.98` via `module1-hysteresis/pkg/gui/gui.go`).
- Spacing: `step = 2 * effectivePs / (NumLevels - 1)` (`module1-hysteresis/pkg/ferroelectric/level_bins.go`).
- Level 1 center: `-effectivePs`.
- Level N center: `+effectivePs`.
- GUI level mapping (used for `a.discreteLevel`) normalizes by the effective range.
- `levelNorm = clamp(P / effectivePs, -1, 1)`.
- `level_index = round((levelNorm + 1) / 2 * (NumLevels - 1))`.
- Stored as 0-based `level_index`; logged as both `level_index` and `level = level_index + 1` (`module1-hysteresis/pkg/gui/physics_engine.go`, `module1-hysteresis/pkg/gui/data_logger.go`).
- The guard band **does not change bin width**; it only shrinks the "safe" region used during verify (`module1-hysteresis/pkg/ferroelectric/level_bins.go`).

**Tuning Note**
If you want the outer levels closer to full saturation, increase `target_range_frac` in `config/materials.yaml`
(e.g., `literature_superlattice.target_range_frac`).
