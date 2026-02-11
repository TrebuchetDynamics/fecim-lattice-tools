# Landau-Khalatnikov Polydomain Ensemble Remanent Staircase Diagnostic

This diagnostic checks that the Landau-Khalatnikov (LK) solver, when run in *polydomain ensemble* mode, produces a multi-level **remanent staircase** at `E=0` under a write/verify-like loop.

It is a prerequisite for enabling deterministic ISPP convergence testing with LK ensemble (see the skipped `TestISPPConverges_LandauK_Ensemble_Superlattice`).

## Diagnostic Procedure

Implemented by `TestLandauKEnsemble_RemanentStaircase_Superlattice` in `landau_remanent_sweep_test.go`.

Configuration (for the literature superlattice material):

- LK solver with `EnableNoise=false`
- `UseNLS=true` (enables partial switching via the NLS-style ensemble mechanism)
- `EnableEnsemble(numDomains=256, seed=0)`
  - `seed=0` means the solver derives a deterministic seed from `(material name, domain count)` (see `shared/physics/landau.go`).

Sweep loop (positive branch):

- For `k=0..60`, set pulse magnitude `E = (2.5*k/60)*Ec`.
- Start each trial from negative saturation: `SetState(-Ps)`.
- Apply a short pulse at `E` for `pulseSteps` solver steps.
- Relax/verify at `E=0` for `relaxSteps` solver steps.
- Quantize remanent `P(E=0)` into an integer level `1..30` via `levelFromP`.

## Acceptance Metrics

The test computes, logs, and asserts the following metrics:

1. **Distinct-level count threshold**
   - Metric: number of distinct quantized remanent levels observed across the sweep.
   - Acceptance: `distinctLevels >= 6`.

2. **Determinism (fixed seed)**
   - Metric: the full quantized level sequence over the sweep.
   - Acceptance: running the same sweep twice (fresh solver, same material and `seed=0`) must produce an identical level sequence (or identical stable hash).

3. **Monotonic trend tolerance**
   - Metric: number of decreases in the level sequence as pulse magnitude increases, and the worst (largest) single-step drop.
   - Acceptance: at most `2` decreases total, and no decrease larger than `1` level.

4. **Remanent stability after relax-at-`E=0`**
   - Metric: for each `k`, the fractional change in polarization between the final two relax steps at `E=0`:
     - `deltaFrac = |P_end - P_prev| / |Ps|`
   - Acceptance: `maxDeltaFrac <= 1e-3` across the sweep, and the quantized level must not change during the final relax step.

