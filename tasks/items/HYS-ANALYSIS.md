+++
schema = 1
id = "HYS-ANALYSIS"
title = "Expose real hysteresis diagnostics and honest reference context in Go"
short = "PUND FORC and evidence results"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 3000
depends_on = ["HYS-CORE"]
touches = ["workbench/hysteresis/analysis.go", "workbench/hysteresis/analysis_test.go", "workbench/hysteresis/reference.go", "workbench/hysteresis/artifacts.go", "shared/physics/worldclass_pund.go", "shared/physics/worldclass_forc.go", "shared/viewmodel/hysteresis/viewmodel.go", "module1-hysteresis/pkg/gui/literature_overlay.go", "module1-hysteresis/pkg/gui/forc_panel.go", "module1-hysteresis/pkg/gui/pund_panel.go", "module1-hysteresis/README.md", "module1-hysteresis/pkg/gui/analysis_parity_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement ANALYSIS seam preserving actual GUI PUND/FORC diagnostic profiles plus read-only cited/imported context. The plan Diagnostic protocol decision is normative: fixed legacy-pund-demo parameters; FORC uses extracted guiFORCEverett with selected material Ps and GUI Emax/reversal/grid defaults, not headless TanhEverett. Both identify independent Preisach diagnostics even when selected loop is L-K; copy loop revision/context but never claim its model/T/stress was applied. Pure byte parsers/Go serializers, no fabricated evidence/fitting/implicit I/O."
acceptance = ["Given any selected loop engine/material, when legacy-pund-demo is run, then fixed Ec=3e7 V/m, Ps=0.25 C/m², Delta=0.3Ec, 5Ec drive, 100 ns width, 5 ns dt and 1e-12 m² area produce the existing GUI/shared pulse/charge results with independent Preisach/HZO-demo identity, not the selected loop model; TestAnalysisPUND and native GUI TestAnalysisGUIParity.", "Given selected-material Ps and GUI FORC defaults Emax=1 V/m, reversals=21, resolution=21, when bounded analysis completes/cancels, then it preserves guiFORCEverett's weighted expression/clamps and actual sweep/density/grid metadata, never silently substitutes headless TanhEverett or reports L-K execution; TestAnalysisFORC and TestAnalysisGUIParity. Validate bounds before allocation and never call partial matrices complete.", "Given cited/imported, malformed or illustrative reference data, when context is displayed/serialized, then source/branch and imported-unverified/simulation-default status remain, unknown comparisons are unavailable with reasons and illustrative Pr/Ec bands are not experimental evidence; TestAnalysisEvidence."]
checks = ["go test ./workbench/hysteresis -run TestAnalysis -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/gui -run TestAnalysisGUIParity -count=1", "go test ./shared/physics -run 'TestPUND|TestFORC' -count=1", "go test ./shared/viewmodel/hysteresis ./validation/peloop -count=1"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HYS-ANALYSIS.toml"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: NewAnalysis(AnalysisRequest), AnalysisJob.Advance/Page/Result plus bounded typed diagnostic/artifact data and evidence notices. Shared opaque loops gain bounded resumable/context-aware work where required. No Worker registry edits.
- Consumes: HYS-CORE immutable input snapshot/catalogue and Go presentation; GUI runPUND/runFORCWorkflow scientific oracles, shared underlying routines, headless compatibility guards (not equivalent FORC oracle), and reference evidence rules.
- Repo obligations triggered: ADR scientific identity/evidence and honesty-audit requirements; module1 README documents diagnostics/import labels. Preserve existing false experimental-validation flags, source units and branch semantics. No new experimental fixture or synthetic measurement assertion; test inputs may be labelled synthetic.
- Next action: Add native TestAnalysisGUIParity and TestAnalysisPUND for the actual fixed GUI PUND profile, plus guiFORCEverett FORC characterization; confirm RED before extraction.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.

- Builder finding resolution: P2 resolved: actual GUI diagnostic profiles/units are the oracle; Preisach model and applied versus contextual inputs are explicitly separate. Worker budgets use only private shaping and do not preclaim the React-route measurement.
