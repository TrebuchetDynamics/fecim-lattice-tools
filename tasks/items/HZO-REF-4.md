+++
schema = 1
id = "HZO-REF-4"
title = "Compare HZO data with the unchanged model baseline"
short = "Compare data with the unfitted model"
status = "open"
goal = "HZO-REF"
section = "Now"
rank = 5000
depends_on = []
touches = ["validation/peloop/compare.go", "validation/peloop/compare_test.go", "validation/peloop/baseline.go", "validation/peloop/baseline_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement Sample, ReadSamples, Compare and SimulateBaseline at the interfaces in Sources. Compare exact branch/index/field pairs with per-branch and pooled counts, RMSE/MAE in uC/cm2; optional normalized RMSE uses observed max(abs(reference P)), not fitted Ps. Reuse unchanged DefaultHZO/NewPreisachModel at constructor 300K/1GPa, Reset, two fixed conditioning cycles over reference field extrema with 512 steps per half-cycle, then sample at reference E values. Record the full material snapshot/settings. Never fit measured P or imply 10kHz/wake-up dynamics."
acceptance = ["Given hand-computable matched branch samples, when Compare runs, then per-branch/pooled errors, counts and labelled observed normalizer are exact; TestCompareMatchedBranches proves the numerical contract.", "Given empty/nonfinite or mismatched branch/index/field samples, when compared, then reject without interpolation or cross-branch matching; zero observed normalization has an unavailable normalized metric rather than NaN/Inf; TestCompareMatchedBranches proves these cases.", "Given repeated inputs and reference P changed without changing fields, when SimulateBaseline runs, then simulation remains deterministic and unaffected by measured P, leaves the source preset unchanged and records the fixed constructor/conditioning settings; TestBaselineUsesUnchangedPreset proves this."]
checks = ["go test ./validation/peloop -run '^(TestCompareMatchedBranches|TestBaselineUsesUnchangedPreset)$' -count=1"]
exclusions = ["No changes to existing model equations, coefficients, presets, synthetic/quarantined fixtures or legacy validation counts.", "No unrelated staged/unstaged edits, commits without owner diff approval, dispatch, publication, deployment, credentials, new dependencies, manifests or CI changes."]
+++
Sources: docs/research/validation/landing-page-figure-readiness.md#experimental-hzo-evidence-frontier

Consumes: existing shared/physics.DefaultHZO and module1-hysteresis/pkg/ferroelectric.NewPreisachModel/Reset/Update; shared physical CSV contract, not the actual extracted dataset. Produces: Sample{Branch string, PointIndex int, FieldMVcm float64, PolarizationUCcm2 float64}; ReadSamples(io.Reader) ([]Sample,error); Compare([]Sample,[]Sample) (Metrics,error); SimulateBaseline([]Sample) (BaselineResult,error). BaselineResult records simulated Samples, material snapshot and fixed conditioning/settings. Convert MV/cm to V/m by 1e8 and C/m2 to uC/cm2 by 1e2.

Next action: write TestCompareMatchedBranches and record expected RED, then TestBaselineUsesUnchangedPreset before that behavior. Checks are proposed/unrun. Independent of HZO-REF-1/2/3; their files are disjoint. Do not import test-only simulatePreisachLoop or refactor existing model/test helpers. No residual threshold is a success condition; a mismatch is data, not a reason to fit.
