+++
schema = 1
id = "WEB-COMPARISON"
title = "Port comparison scenarios with dimensionally honest metrics and exports"
short = "Web comparison and honest units"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 5000
depends_on = ["WEB-WORKBENCH"]
touches = ["workbench/modules/comparison/", "shared/viewmodel/comparison/", "module5-comparison/pkg/comparison/", "module5-comparison/pkg/gui/scenario_profiles.go", "module5-comparison/pkg/gui/evidence_model.go", "module5-comparison/pkg/gui/export.go", "module5-comparison/pkg/gui/app.go", "module5-comparison/pkg/gui/calculations_test.go", "module5-comparison/pkg/gui/data_confidence_interval_test.go", "module5-comparison/pkg/gui/data_sensitivity_test.go", "cmd/fecim-comparison-wasm/", "web/src/modules/comparison/", "web/test/e2e/comparison.spec.ts", "validation/wasm/comparison_parity_test.go", "module5-comparison/README.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement full M5 roadmap: workloads/rate, review/presentation, scenarios, energy/market/ROI/fabrication context, source evidence, diff/sensitivity and local export. Extract GUI-bound Go calculations into reusable model and web port; existing comparison VM rejects every action. Keep architecture-demo and scenario protocols distinct. Normative decisions: inference rate per second, correct day/J/kWh/million-USD conversions, versioned artifacts, withhold currency-based CO2 and statistical confidence claims, finite validation and deterministic tie breaks. No fitting/new device-performance claims."
acceptance = ["Given hand-computable MAC/fJ/rate examples, when calculation/export uses canonical Go quantities, then inferences/s→daily count, J→kWh and USD→million USD follow roadmap dimensional identities and metadata versions; TestComparisonWebUnits fails legacy /3600 and /1000 bugs, retains explicit 730h/month and $0.10/kWh assumptions only for their declared protocol.", "Given each workload/scenario and invalid/missing denominator/source, when calculate/diff/sensitivity runs, then real Go actions work, unavailable CO2 has a reason, heuristic confidence is not a statistical interval, evidence remains simulation/assumption and sorting is deterministic; TestComparisonWebWorkflow plus native architecture/scenario compatibility tests.", "Given both builds and identical native/Node/Firefox requests, when scenario controls/tables/source/exports are used, then successful canonical numeric bytes and actual browser results match, admitted Worker budget/isolation passes and local exports are rooted; comparison.spec.ts plus TestModuleProductionParity comparison cases, not mocked or manual acceptance."]
checks = ["go test ./workbench/modules/comparison ./shared/viewmodel/comparison ./module5-comparison/pkg/comparison -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module5-comparison/pkg/gui -count=1", "env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run 'TestModuleProductionParity|TestExampleResultsJSONNativeWASM' -count=1 -v", "cd web && npm test && npm run typecheck && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/comparison.spec.ts"]
exclusions = ["No implementation/deployment/service change follows from planning; future implementation preserves Go science/CLI, original example-byte parity, existing qualification gates and no hosted scientific upload/persistence/localhost probing.", "No automatic EDA-tool/model/dataset/PDK/container acquisition or installation; explicitly scoped pinned build/test dependency provisioning is allowed only during authorized implementation. No new physics/TS scientific model, fabricated accuracy/provenance/tool pass, premature Fyne deletion/default switch, public release or push."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-COMPARISON.toml"
+++


- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md; docs/internals/adr/0005-unified-react-hosted-and-local.md; docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md
- Produces: Comparison Go model/web port/Screen/Worker and versioned local serializers, actual scenario/diff/sensitivity/evidence and corrected dimensional tests; no unsupported calibrated CO2 output.
- Consumes: WEB-WORKBENCH fields/Go presentation/transport/paging/artifact boundaries; native scenario/architecture protocols and source ledger. Does not pretend to aggregate new M2/M3/M4 screen results.
- Repo obligations triggered: Roadmap M5 science adjudication is binding, not a worker design interview. Source lines app.go:631–665, evidence_model.go:76–93 and export.go:76–121 explain rebasing old incorrect assertions. Scope allows minimal extraction/unit/provenance fixes in named GUI/core and affected tests, not coefficient tuning or paper promotion. Add README limitations and feature tests discovered by the shared CI gate; future tests unrun. Central route registry unchanged.
- Next action: Write TestComparisonWebUnits with a one-Joule-per-inference rate case and TestComparisonWebWorkflow with unavailable CO2; verify RED before fixing/extracting.
- Verification status: All new feature tests/commands are proposed and unrun. Source inspection is not numerical/runtime/manual/tool validation. No implementation or dispatch in this planning occurrence.

- CI ownership: WEB-WORKBENCH owns shared workflow/build discovery once; this task supplies real module tests/Screen/entry artifacts picked up by it, without editing a common CI or route registry. Declared cold all-module test/presentation overlaps still need merge review; no artificial dependency is added.
- Pending live: Real Orca + Firefox/Linux keyboard/speech qualification of Comparison in hosted and authenticated local profiles remains NOT RUN; requires a human operator and is not an offline acceptance claim.
