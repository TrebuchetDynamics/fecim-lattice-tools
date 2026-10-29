+++
schema = 1
id = "WEB-CIRCUITS"
title = "Port material-aware circuit write read and compute workflows to the web"
short = "Web circuits write read compute"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 4000
depends_on = ["WEB-WORKBENCH"]
touches = ["workbench/modules/circuits/", "shared/viewmodel/circuits/", "module4-circuits/pkg/arraysim/", "module4-circuits/pkg/gui/tab_unified.go", "module4-circuits/pkg/gui/headless_gui_physics_parity_test.go", "cmd/fecim-circuits-wasm/", "web/src/modules/circuits/", "web/test/e2e/circuits.spec.ts", "validation/wasm/circuits_parity_test.go", "shared/viewmodel/e2e_all_modules_workflow_test.go", "module4-circuits/README.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement M4 roadmap with persistent cell/array state, READ/WRITE/COMPUTE/undo/reset, vector/DAC/sense/selected-cell controls, material/node/architecture/Ideal-TierA-TierB, ISPP, stress/SNR/timing/reference/logs and local exports. Use arraysim/peripheral/device cores and actual native GUI parity, replacing synthetic adapter matrices. Preserve circuit-level ISPP as distinct from HYS. Hosted bounded CPU profiles and explicit local larger/file/GPU capabilities; no TypeScript electronics model."
acceptance = ["Given a real material-aware array, when web commands WRITE then READ then COMPUTE including masks/signed inputs/undo/reset, then persistent cell/sense/coupling/timing results agree with native arraysim/GUI oracles, not synthetic fixtures; TestCircuitsWebWorkflow and existing TestHeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware/dense/KCL/sense tests.", "Given valid/invalid sizes/bit/sense/solver parameters and cancellation, when admitted hosted circuits run natively/Node/Firefox, then canonical successful numerical bytes agree exactly, Go caps/yields preserve state and measured per-profile resource/payload gates pass; TestModuleProductionParity circuits cases and failure/transaction tests.", "Given both real deployment profiles, when keyboard users program/select/read/vector-compute and inspect timing/log/evidence/plots or exports, then circuits.spec.ts covers all declared workflows with complete Go-spoken table values and honest unsupported/GPU/file status; hosted isolation and local roots hold, real Orca is not claimed."]
checks = ["go test ./workbench/modules/circuits ./shared/viewmodel/circuits ./module4-circuits/pkg/arraysim -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module4-circuits/pkg/gui -run 'TestHeadlessPhysicsParity_GUIVsHeadless_ReadComputeWriteStep_MaterialAware|TestWriteVerifyRead_LevelRoundTrip|TestComputeSignedInputs|TestFullPipeline_WriteReadCompute' -count=1", "env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -run 'TestModuleProductionParity|TestExampleResultsJSONNativeWASM' -count=1 -v", "env CGO_ENABLED=1 go test -race ./workbench/modules/circuits ./module4-circuits/pkg/arraysim -count=1", "cd web && npm test && npm run typecheck && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/circuits.spec.ts"]
exclusions = ["No implementation/deployment/service change follows from planning; future implementation preserves Go science/CLI, original example-byte parity, existing qualification gates and no hosted scientific upload/persistence/localhost probing.", "No automatic EDA-tool/model/dataset/PDK/container acquisition or installation; explicitly scoped pinned build/test dependency provisioning is allowed only during authorized implementation. No new physics/TS scientific model, fabricated accuracy/provenance/tool pass, premature Fyne deletion/default switch, public release or push."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-CIRCUITS.toml"
+++


- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md; docs/internals/adr/0005-unified-react-hosted-and-local.md; docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md
- Produces: Circuits web port/Screen/CPU Worker with actual persistent device state and parity/resource evidence; native UI extraction helpers remain shared, no new scientific engine.
- Consumes: WEB-WORKBENCH transport/Go field/job/paging capabilities; existing arraysim engine, sense/device/controller and GUI-native characterization. Reuses HYS material/identity infrastructure, not its ISPP semantics.
- Repo obligations triggered: Roadmap M4 native sizes {1,2,4,8,16,32,64,128}, bit/node domains and initial hosted bounded profiles. Add context/resumable/transaction validation where needed inside task-owned solvers, preserve numeric contracts/goldens. Update module README and per-feature tests discovered by the shared CI gates. Synthetic fixture-only adapter tests must no longer claim material/coupling parity. Cold legacy all-module tests may overlap; shared parity harness/CI remain foundation-owned, with no central registry edits.
- Next action: Add TestCircuitsWebWorkflow for one real WRITE→READ→COMPUTE sequence that fails the existing synthetic web path; run RED before extraction/wiring.
- Verification status: All new feature tests/commands are proposed and unrun. Source inspection is not numerical/runtime/manual/tool validation. No implementation or dispatch in this planning occurrence.

- CI ownership: WEB-WORKBENCH owns shared workflow/build discovery once; this task supplies real module tests/Screen/entry artifacts picked up by it, without editing a common CI or route registry. Declared cold all-module test/presentation overlaps still need merge review; no artificial dependency is added.
- Pending live: Real Orca + Firefox/Linux acceptance for both hosted and authenticated local Circuits workflows is NOT RUN. Automated keyboard/focus/axe and runtime checks do not qualify actual speech or human task completion.
