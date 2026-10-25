+++
schema = 1
id = "HZO-REF-2"
title = "Convert HZO plot pixels into physical units"
short = "Convert plot pixels to physical units"
status = "done"
goal = "HZO-REF"
section = "Next"
rank = 2000
depends_on = ["HZO-REF-1"]
touches = ["validation/peloop/digitize.go", "validation/peloop/digitize_test.go", "cmd/pe-reference/main.go", "cmd/pe-reference/main_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement Digitize(SourceProfile, io.Reader) ([]byte,error) and the offline pe-reference digitize command using two independent affine axes. Convert branch,point_index,x_px,y_px to branch,point_index,field_MV_cm,polarization_uC_cm2 without changing reference values to match a model. Use ascending/descending labels, zero-based per-branch indices, original raster coordinates, 10000-row and 2MiB bounds. No internet access from the command."
acceptance = ["Given hand-computable pixel/tick anchors including inverted y, when Digitize converts the CSV, then exact physical values and branch/index ordering match the expected CSV; TestDigitizeCalibratedPixels proves it.", "Given invalid/nonfinite/out-of-bounds points, degenerate axes, missing/repeated indices, duplicate within-branch fields or over-limit input, when converted, then fail rather than smoothing, extrapolating or completing points; TestDigitizeCalibratedPixels proves the negative cases.", "Given valid versus invalid input, when the digitize command runs, then it produces exact deterministic CSV versus nonzero failure with no success output; TestDigitizeCommand proves the public seam."]
checks = ["go test ./validation/peloop -run '^TestDigitizeCalibratedPixels$' -count=1", "go test ./cmd/pe-reference -run '^TestDigitizeCommand$' -count=1"]
exclusions = ["No changes to existing model equations, coefficients, presets, synthetic/quarantined fixtures or legacy validation counts.", "No unrelated staged/unstaged edits, commits without owner diff approval, dispatch, publication, deployment, credentials, new dependencies, manifests or CI changes."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HZO-REF-2.toml"
+++
Sources: docs/research/validation/landing-page-figure-readiness.md#experimental-hzo-evidence-frontier

Consumes: HZO-REF-1 SourceProfile and ReadSourceProfile. Produces: Digitize and pe-reference digitize --profile PATH --pixels PATH --out PATH. Headers and semantics are fixed in Sources, including rejecting duplicate fields within each branch while permitting the same field on opposite branches.

Next action: write TestDigitizeCalibratedPixels and record expected RED, then the command case before its implementation. All checks are proposed/unrun. Use clearly synthetic, hand-computable unit inputs; they are not measured reference fixtures. Protect existing output paths from overwrite.
