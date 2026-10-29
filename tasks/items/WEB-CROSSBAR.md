+++
schema = 1
id = "WEB-CROSSBAR"
title = "Port real crossbar models and interactive matrices to the web"
short = "Web crossbar physics workflows"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 3000
depends_on = ["WEB-WORKBENCH"]
touches = ["workbench/modules/crossbar/", "shared/viewmodel/crossbar/", "shared/crossbar/", "module2-crossbar/pkg/gui/app_analysis.go", "module2-crossbar/pkg/gui/animation.go", "cmd/fecim-crossbar-wasm/", "web/src/modules/crossbar/", "web/test/e2e/crossbar.spec.ts", "validation/wasm/crossbar_parity_test.go", "shared/viewmodel/e2e_all_modules_action_catalog_test.go", "shared/viewmodel/e2e_all_modules_workflow_test.go", "module2-crossbar/README.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement M2 scope in roadmap: actual conductance/matrix/IR/sneak/input-output/ideal-actual/accuracy/FeCAP/stack/literature workflows, controls, selection and local files. Use shared/crossbar models with explicit normalized/enhanced/parasitic/charge-domain profiles, seeded data and typed units. Correct the plain-dot-product adapter and IR toggle bug under TDD; do not reproduce fabricated throughput/latency/hardware accuracy. Hosted Go CPU caps versus local full existing GUI sizes, real progress/error/cancel and accessible cell/vector alternatives."
acceptance = ["Given fixed programmed matrix/vector/config/seed and basic/enhanced/FeCAP/parasitic profiles, when web adapter executes, then actual core vectors/cell/convergence outputs and declared normalization/units agree, toggling IR affects IR not SneakPaths and invalid shape/nonfinite state is rejected atomically; TestCrossbarWebWorkflow plus existing KCL/KVL/enhanced/charge tests.", "Given admitted hosted profiles and identical native inputs/versions, when real Node/Firefox feature Worker runs, then successful canonical scientific bytes exactly match native, Go allocation/work caps and cancellation hold and measured payload/default-result gate passes; extend TestModuleProductionParity for crossbar, no tolerance or empty/failure-only equivalence.", "Given real transports in both builds, when programming/selecting a cell, reading matrix/input/output, switching analysis modes and using file-only actions, then crossbar.spec.ts reaches every declared native workflow through keyboard/tables with Go values and honest unavailable states, hosted isolation holds and local file operations remain rooted; not a manual Orca pass."]
checks = ["go test ./workbench/modules/crossbar ./shared/viewmodel/crossbar ./shared/crossbar -count=1", "env CGO_ENABLED=1 go test -race ./workbench/modules/crossbar ./shared/viewmodel/crossbar -count=1", "env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run 'TestModuleProductionParity|TestExampleResultsJSONNativeWASM' -count=1 -v", "cd web && npm test && npm run typecheck && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/crossbar.spec.ts", "go test ./shared/viewmodel -run TestAllModulesE2E -count=1"]
exclusions = ["No implementation/deployment/service change follows from planning; future implementation preserves Go science/CLI, original example-byte parity, existing qualification gates and no hosted scientific upload/persistence/localhost probing.", "No automatic EDA-tool/model/dataset/PDK/container acquisition or installation; explicitly scoped pinned build/test dependency provisioning is allowed only during authorized implementation. No new physics/TS scientific model, fabricated accuracy/provenance/tool pass, premature Fyne deletion/default switch, public release or push."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-CROSSBAR.toml"
+++


- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md; docs/internals/adr/0005-unified-react-hosted-and-local.md; docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md
- Produces: Crossbar web-port implementation behind the common factory/capability check, module Screen, dedicated CPU Worker and actual local/hosted parity/browser evidence. Serialization is pure Go; I/O goes through existing local capability services.
- Consumes: WEB-WORKBENCH protocol/job/client/result components/route slots; existing shared/crossbar core/native GUI behavior, seeded matrices and Kirchhoff/FeCAP oracles. No dependency on a new MNIST/Circuits UI.
- Repo obligations triggered: Roadmap M2 profiles, 32x32 initial hosted versus existing 128x128 local UI cap, strict vector shape, no WebGPU and preserved scientific evidence. Scope includes task-owned constructor/validation/context/yield and adapter-truth repairs in listed core/VM owners; no coefficient/golden/tolerance retuning. Update module README and module tests in the preinstalled shared CI gate. Cold legacy all-module test overlap is declared, not a fake dependency; per-feature parity files are separately owned. Central route/factory registry stays WEB-WORKBENCH-owned.
- Next action: Write TestCrossbarWebWorkflow demonstrating real IR action/core output and preservation on shape failure; run RED before replacing plain dot-product path.
- Verification status: All new feature tests/commands are proposed and unrun. Source inspection is not numerical/runtime/manual/tool validation. No implementation or dispatch in this planning occurrence.

- CI ownership: WEB-WORKBENCH owns shared workflow/build discovery once; this task supplies real module tests/Screen/entry artifacts picked up by it, without editing a common CI or route registry. Declared cold all-module test/presentation overlaps still need merge review; no artificial dependency is added.
- Pending live: Human Orca + Firefox/Linux acceptance for hosted and authenticated local crossbar workflows remains NOT RUN; automated keyboard/table/viewport checks do not qualify screen-reader speech.
