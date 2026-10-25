+++
schema = 1
id = "HZO-REF-3"
title = "Independently digitize the published HZO reference markers"
short = "Digitize published reference markers"
status = "open"
goal = "HZO-REF"
section = "Next"
rank = 3000
depends_on = ["HZO-REF-1", "HZO-REF-2"]
touches = ["validation/literature/data/experimental/kim2020_fig3a_ref_digitized/points.pixels.csv", "validation/literature/data/experimental/kim2020_fig3a_ref_digitized/loop.csv", "validation/literature/data/experimental/kim2020_fig3a_ref_digitized/extraction.md", "validation/literature/hzo_measured_data_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Independently select visible red Ref. marker centres from Kim 2020 Figure 3(a) LEFT subpanel, record their original pixel coordinates and branch indices, regenerate the physical CSV with HZO-REF-2, and document point selection, bounds/uncertainty and reproducibility hashes. Select no blue curve, panel 3(b), connecting line or previous MDPI/Park samples. Each branch must preserve the visible span/crossings with at least 10 finite distinct-field markers. Missing or merged markers are disclosed, not guessed. No smoothing or interpolation of reference samples."
acceptance = ["Given the committed raw pixel CSV and admitted profile, when regenerated, then loop.csv is byte-identical and profile/pixel/physical hashes and counts agree; TestHZOIndependentDigitization proves the reproducibility contract.", "Given each branch, when validated, then indices/order are consistent, there are at least 10 finite distinct-field markers with both signs of E/P and preserved zero crossings, and invalid coverage is rejected; TestHZOIndependentDigitization proves the structural contract.", "Given this new dataset, when its integrity test runs, then the source ID/hash remains the independently pinned primary source and no old synthetic/unverified sample file is used or rewritten; TestHZOIndependentDigitization proves the input boundary."]
checks = ["go test ./validation/literature -run '^TestHZOIndependentDigitization$' -count=1"]
exclusions = ["No changes to existing model equations, coefficients, presets, synthetic/quarantined fixtures or legacy validation counts.", "No unrelated staged/unstaged edits, commits without owner diff approval, dispatch, publication, deployment, credentials, new dependencies, manifests or CI changes."]
+++
Sources: docs/research/validation/landing-page-figure-readiness.md#experimental-hzo-evidence-frontier

Consumes: qualified HZO-REF-1 source profile and HZO-REF-2 digitizer. Produces: the new raw-pixel/physical dataset and extraction.md, consumed by HZO-REF-5. Do not modify source.provenance.json without reconciling its producer's actual evidence; record point counts/hashes/selection details in extraction.md. Do not put sidecar JSON beneath recursively ingested experimental-data/.

Next action: write TestHZOIndependentDigitization for missing/unproven new input and record expected RED before adding numeric fixtures. Checks are proposed/unrun. Extract lower branch in ascending E and upper in descending E; disclose that direction is conventional quasi-static assignment, not measured timing. Stop on ambiguous evidence rather than silently changing to another source or making up points.

Non-automated qualification: independent source-image review must compare markers/calibration against the original before any public measured-data promotion. Automated structure/hash checks do not establish trace fidelity; record actual review evidence, never invented human approval. Required review and public promotion remain existing gates, not new worker authority.
