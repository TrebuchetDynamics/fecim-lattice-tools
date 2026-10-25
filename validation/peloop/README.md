# Physical P-E sample comparison

`peloop` provides an offline, unfitted comparison between branch-labelled reference
samples and the existing HZO Preisach model. It does not admit or digitize a
published source, certify trace fidelity, or establish a device-validation claim.

## Input and matching

`ReadSamples(io.Reader)` reads this exact physical CSV header:

```csv
branch,point_index,field_MV_cm,polarization_uC_cm2
```

Branches are `ascending` or `descending`. Each branch starts at index zero and
increments by one in input order. Fields and polarizations must be finite; fields
must be distinct within each branch. The same field on opposite branches is
allowed. Empty input, invalid rows, more than 10000 samples, and CSV input larger
than 2 MiB are rejected. The reader preserves values and order; it does not infer
branch direction, smooth, sort, or complete missing points.

`Compare(reference, simulated)` requires equal sample counts and exact
branch/index/field equality at every position. No tolerance, interpolation or
cross-branch matching is applied. It returns `Metrics.PerBranch` and
`Metrics.Pooled` with counts, RMSE and MAE in uC/cm2. Pooled errors weight each point
equally, not each branch equally.

Each result labels its normalizer as `observed max(abs(reference P))` and records
that value in `ObservedMaxAbsReferencePUCcm2`. A branch uses its own reference
observations; pooled normalization uses all reference observations. `NormalizedRMSE`
is RMSE divided by that observed normalizer (a fraction, not a percentage). It is
`nil` when the observed normalizer is zero. It never uses fitted saturation
polarization. Unrepresentable residuals, accumulated errors or normalized errors
are rejected.

## Unchanged model baseline

`SimulateBaseline(reference)` constructs a fresh model using
[`physics.DefaultHZO`](../../shared/physics/material_presets.go) and
[`ferroelectric.NewPreisachModel`](../../module1-hysteresis/pkg/ferroelectric/preisach.go).
It leaves constructor temperature/stress at 300 K / 1 GPa and does not modify the
preset, equations, coefficients or reference samples.

The protocol is fixed:

1. Call `Reset`, then apply the minimum reference field.
2. Apply two minimum → maximum → minimum conditioning cycles over the reference
   field extrema, with 512 equal field intervals per half-cycle. Each half-cycle
   excludes its starting point and includes its endpoint.
3. Call quasi-static `Update` once at each reference field, in input order, without
   resetting between branches. The caller supplies the physical traversal order.
4. Convert MV/cm to V/m by `1e8` and C/m2 to uC/cm2 by `1e2`.

`BaselineResult` records simulated `Samples`, a complete `Material` value snapshot
in native SI units, and the constructor, conditioning, sampling and conversion
`Settings`. Reference polarization is checked for finiteness but never drives the
model or settings. No residual threshold is a success condition: disagreement is
reported, not fitted away.

Conditioning prepares numerical history only. This quasi-static baseline is **not**
a simulation of a 10 kHz measurement or experimental wake-up dynamics. Constructor
conditions are model defaults, not inferred measurement conditions.

## Verification

From the repository root:

```bash
go test ./validation/peloop -count=1
```

`TestCompareMatchedBranches` covers hand-computable metrics, observed and zero
normalization, and invalid/unmatched samples. `TestReadSamples` covers physical CSV
parsing and bounds. `TestBaselineUsesUnchangedPreset` covers deterministic sampling,
independence from measured P, the complete preset snapshot, fixed settings and
unit conversion. Test inputs are synthetic unit examples, not measured fixtures.
