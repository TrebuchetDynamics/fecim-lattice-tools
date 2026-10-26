+++
schema = 1
id = "HZO-REF-1"
title = "Verify the HZO source and calibrate its plot axes"
short = "Verify source and calibrate axes"
status = "done"
goal = "HZO-REF"
section = "Now"
rank = 4000
depends_on = []
touches = ["validation/peloop/provenance.go", "validation/peloop/provenance_test.go", "validation/literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json", "citations/papers/kim2020_tin_hzo_electrodes.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement a fail-closed SourceProfile reader and record the actual Kim 2020 DOI 10.3390/ma13132968 Figure 3(a) LEFT red Ref. marker source, calibration and attribution. Use the original 3629x1210 publisher PNG hash pinned in Sources. Capture tick-centre anchors at labelled -3/+3 MV/cm and -45/+45 uC/cm2 from original pixels. Metadata describes TiN 90nm/HZO 10nm/TiN 90nm, 10kHz and 1e5 wake-up cycles at 2.5MV/cm; preserve unreported measurement temperature/waveform/exact device diameter as not reported. Do not digitize or generate polarization values in this task."
acceptance = ["Given the qualified primary-source profile, when ReadSourceProfile is called, then it preserves exact DOI/figure/trace, source SHA/dimensions, license evidence, conditions and two anchors per axis; TestSourceProfileAdmission proves admission of that contract.", "Given missing/wrong trace, missing license/hash/conditions/calibration, nonfinite/degenerate anchors or unsupported units, when read, then reject rather than borrowing model defaults; the table-driven negative cases in TestSourceProfileAdmission prove this.", "Given measurement fields not reported in the inspected source, when admitted, then retain explicit not-reported status and a documented pixel selection bound, never anneal-as-measurement temperature or a fitted uncertainty; TestSourceProfileAdmission proves preservation."]
checks = ["go test ./validation/peloop -run '^TestSourceProfileAdmission$' -count=1"]
exclusions = ["No changes to existing model equations, coefficients, presets, synthetic/quarantined fixtures or legacy validation counts.", "No unrelated staged/unstaged edits, commits without owner diff approval, dispatch, publication, deployment, credentials, new dependencies, manifests or CI changes."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HZO-REF-1.toml"
+++
Sources: docs/research/validation/landing-page-figure-readiness.md#experimental-hzo-evidence-frontier

Produces: SourceProfile; ReadSourceProfile(io.Reader) (SourceProfile, error), exactly as specified in Sources. The dataset ID is kim2020_fig3a_ref_digitized, not the legacy mdpi2020 ID.

Next action: write TestSourceProfileAdmission first and record expected RED before production changes. Checks are proposed/unrun. The source raster may be reacquired read-only from the pinned public URL into temporary scratch; verify SHA before using pixels. If the asset changed or rights cannot be established, stop and record the evidence blocker; do not silently change source.

Qualification: automated profile checks do not certify visual trace fidelity or authorize public promotion. No original image/PDF is committed. Implementation/commit permission remains separately gated.
- Pending live: Independent human visual source/trace-fidelity qualification against the pinned original raster remains required before public promotion of a future extracted curve. Automated metadata admission and the read-only code review do not certify marker selection or experimental agreement. No polarization samples were extracted in this task.
