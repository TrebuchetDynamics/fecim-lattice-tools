+++
schema = 1
id = "HZO-REF-5"
title = "Export a reproducible HZO comparison report"
short = "Export a reproducible HZO report"
status = "open"
goal = "HZO-REF"
section = "Next"
rank = 4000
depends_on = ["HZO-REF-2", "HZO-REF-3", "HZO-REF-4"]
touches = ["cmd/pe-reference/compare.go", "cmd/pe-reference/compare_test.go", "cmd/pe-reference/main.go", "docs/research/validation/landing-page-figure-readiness.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Integrate pe-reference compare --profile PATH --data PATH --out-dir PATH with the measured extraction and baseline comparator. Publish metrics.json, overlay.svg and caption.txt together into a non-existing caller-selected directory only after successful validation. Include source/data/profile/model fingerprints, Git revision plus dirty state, material/settings, counts, units, exact-field policy and all source/model limitations. Use measured markers and separate simulated lines. comparison_kind is unfitted-baseline and experimental_validation remains false even with low RMSE. Do not deploy or add a headline experimental overlay."
acceptance = ["Given the new measured extraction and unchanged baseline, when compare runs, then the bundle contains reproducible metrics, hashes, branch counts, RMSE units/normalizer and the same scientific values in JSON/SVG/caption; TestCompareCommandEvidenceBoundary proves the end-to-end consistency.", "Given any residual size, when reporting, then preserve the mismatch, unfitted-baseline and experimental_validation=false plus unknown measurement temperature/waveform and quasi-static/10kHz limits; TestCompareCommandEvidenceBoundary proves that low error never promotes predictive validity.", "Given invalid input or an existing output path, when compare runs, then return nonzero, leave no newly published partial/success bundle and preserve pre-existing paths; TestCompareCommandEvidenceBoundary proves the rejection boundary."]
checks = ["go test ./cmd/pe-reference -run '^TestCompareCommandEvidenceBoundary$' -count=1", "go test ./validation/peloop ./cmd/pe-reference -count=1", "go test -race ./validation/peloop ./cmd/pe-reference -count=1"]
exclusions = ["No changes to existing model equations, coefficients, presets, synthetic/quarantined fixtures or legacy validation counts.", "No unrelated staged/unstaged edits, commits without owner diff approval, dispatch, publication, deployment, credentials, new dependencies, manifests or CI changes."]
+++
Sources: docs/research/validation/landing-page-figure-readiness.md#experimental-hzo-evidence-frontier

Consumes: HZO-REF-2 command seam; HZO-REF-3 source/pixel/physical extraction artifacts; HZO-REF-4 Sample/ReadSamples/Compare/SimulateBaseline and results. Produces: the single-command local bundle and evidence-bound readiness update. Validate/generate into a sibling temporary directory, then publish the new output directory; never remove or overwrite an existing destination.

Next action: write TestCompareCommandEvidenceBoundary and record expected RED. All checks are proposed/unrun. main.go is also owned by HZO-REF-2, already ordered by the real command-seam dependency. Keep model/data fingerprints honest for dirty runs; HEAD alone is insufficient. No raw source images/PDFs enter the bundle.

Qualification: update readiness only from actual extraction/review/run evidence; no automatic README/deployment, experimental-validation counter increment, research-use assertion or JOSS eligibility claim. Repository full-suite and required independent review run once at eventual integrated closeout under separately authorized execution. Focused package/race tests do not substitute for that full gate.
