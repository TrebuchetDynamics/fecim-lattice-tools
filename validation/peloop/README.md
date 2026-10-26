# Physical P-E comparison

`peloop` compares branch-labelled physical samples with an unchanged educational
model baseline. It does not fit model parameters or establish device validation.
Source admission is separate: see [the provenance reader](provenance.go).

## Library interface

[compare.go](compare.go) provides `Sample`, `ReadSamples(io.Reader)` and
`Compare(reference, simulated []Sample)`. The physical CSV header is exactly:

```csv
branch,point_index,field_MV_cm,polarization_uC_cm2
```

- Branch labels are `ascending` and `descending`. Indices start at zero and are
  consecutive per branch in input order. Fields are strictly increasing on the
  ascending branch and decreasing on the descending branch. Opposite branches
  can contain the same field; duplicate fields within one branch are rejected.
- Values use MV/cm and uC/cm2. The reader rejects empty, malformed and nonfinite
  input, more than 10000 rows, and input larger than 2 MiB; errors return no samples.
- Comparison requires exact branch/index/field equality at each input position.
  It does not reorder, interpolate or match across branches.
- `Metrics.PerBranch` and `Metrics.Pooled` report counts, RMSE and MAE in uC/cm2.
  Pooled errors weight points, not branches equally.
- Each `ErrorMetrics.Normalizer` is labelled `observed max(abs(reference P))`.
  `ObservedMaxAbsReferencePUCcm2` uses the relevant branch or pooled reference
  points, never preset Ps. `NormalizedRMSE` is a dimensionless fraction, or nil
  when the observed normalizer is zero. Unrepresentable errors are rejected.

## Unfitted baseline

[baseline.go](baseline.go) provides `SimulateBaseline(reference []Sample)` and
`BaselineResult{Samples, Material, Settings}`. The result retains the entire
`physics.HZOMaterial` value in native field units (including eV and GPa) and the sampling settings.

The protocol uses a fresh, unchanged `shared/physics.DefaultHZO()` and
`ferroelectric.NewPreisachModel` at constructor defaults **300 K / 1 GPa**:

1. Call `Reset`, then apply the minimum reference field.
2. Run two fixed min → max → min conditioning cycles over the reference extrema,
   with 512 equal steps per half-cycle, excluding each start and including each end.
3. Call quasi-static `Update` at every reference field in input order, without
   resetting between branches. Convert MV/cm to V/m by 1e8 and C/m2 to uC/cm2 by 1e2.

Measured polarization is checked for validity but never drives the model. Numerical
conditioning is not measured wake-up cycling. Constructor temperature/stress are
not inferred measurement conditions. This protocol does **not** reproduce 10 kHz
measurement dynamics or wake-up dynamics, and no residual threshold establishes
success. The [independent extraction](../literature/data/experimental/kim2020_fig3a_ref_digitized/extraction.md)
and the offline `pe-reference compare --profile PATH --data PATH --out-dir PATH`
command now provide a reproducible local comparison bundle. See the
[report workflow and evidence limits](../../docs/research/validation/landing-page-figure-readiness.md#reproducible-local-hzo-comparison-implemented)
for prerequisites, JSON/SVG/caption contents, atomic no-replace publication and
verification. A report is an **unfitted baseline**, not experimental validation.

## Offline verification

From the repository root:

```sh
go test ./validation/peloop -run '^(TestCompareMatchedBranches|TestBaselineUsesUnchangedPreset)$' -count=1
go test ./validation/peloop -count=1
```

The focused tests cover hand-computable branch/pooled errors, exact-match rejection,
zero normalization, deterministic field-only simulation, unchanged material and the
fixed drive protocol. `TestReadSamples` covers physical CSV admission. Recheck this
page when the reader, comparison metrics or baseline protocol changes.
