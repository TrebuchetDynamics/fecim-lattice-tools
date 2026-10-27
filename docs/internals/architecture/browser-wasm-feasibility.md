# Browser Go/WASM feasibility assessment

Purpose: separate **observed package compilation**, **selected execution**, and **unverified product capabilities** for the accepted [one-React-app decision](../adr/0005-unified-react-hosted-and-local.md). The initial assessment below predates the validation implementation. The subsequent [three-runtime gate receipt](../testing/wasm-parity-gate.md) now records exact example results.json parity and real headless Firefox Worker measurements. This is not a shipped browser application or an Orca acceptance claim.

## Method and scope

Toolchain: Go `go1.26.10 linux/amd64`; target `GOOS=js GOARCH=wasm CGO_ENABLED=0`. Twenty scoped package/command paths were checked with:

```text
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -o /tmp/<unique-scratch>/<package>.a ./<package>
```

No source, build tags, dependencies or manifests were changed. Existing repository WASM build constraints were used as-is. A non-main package output is an object/archive, **not an executable browser module**; only the existing command target links a main. That CLI main is still not a browser worker. This is not an exhaustive compile audit of all seven module trees or all optional integrations.

Initial validation began on HEAD `c5adb475` with an already-running repository runner. Rooted-persistence work may land independently during inspection; it was not implemented by this documentation pass. Repeat relevant checks when those files or dependencies change. Compiler outputs/error logs and the per-target matrix are local artifacts under `/tmp/fecim-wasm-feasibility.tnjWSZ/`; do not mistake these paths for distributable assets or durable release receipts.

## Observed compile matrix

| Package / target | js/wasm result | Browser interpretation |
|---|---|---|
| `shared/physics` | PASS | Reuse CPU solvers/presets; file-based configuration/calibration/export helpers need byte/in-memory adapters or exclusion. |
| `shared/crossbar` | PASS | Existing `gpu_mvm_js.go` disables GPU; CPU MVM/IR-drop/sneak paths remain candidates. File import/export helpers are not browser operations. |
| `shared/peripherals` | PASS | CPU DAC/ADC/TIA and related models are candidates; expose bounded operations through Go worker. |
| `shared/system` | PASS | Area/energy/latency calculations are candidates. |
| `module1-hysteresis/pkg/ferroelectric` | PASS | Existing Go material/Preisach functions compile; bind data, not old UI rendering. |
| `module1-hysteresis/pkg/controller` | PASS | ISPP control code compiles; bounded deterministic stepping/cancellation still needs browser qualification. |
| `module1-hysteresis/pkg/simulation` | PASS | Engine compiles; use controlled stepping/data snapshots, not an unbounded realtime desktop loop. |
| `module1-hysteresis/pkg/headless` | PASS | Compilation is not browser readiness: CLI calibration and CSV logger write files. Keep those paths local; reuse underlying compute, not this CLI as a worker. |
| `module2-crossbar/pkg/crossbar` | NOT A GO PACKAGE | Directory contains instructions but no Go source; build reports “no Go files”. Actual reusable crossbar core is `shared/crossbar`. Not a numerical port failure. |
| `module4-circuits/pkg/arraysim` | PASS | CPU array/peripheral behavior compiles; cap work and control seeds. |
| `module4-circuits/pkg/gpuperiph` | FAIL | Vulkan dependency reports undefined Result/Semaphore/Fence and related native symbols with CGO disabled. Exclude this GPU path from hosted execution; use the existing CPU models, not TypeScript replacement physics. |
| `module5-comparison/pkg/comparison` | PASS | Go architecture/comparison data can be reused; React replaces old rendering. |
| `workbench/project` | PASS | Pure schema/design/expansion logic is reusable; `LoadRoot`/configuration editor still require native filesystem capabilities. Need in-memory byte/data input around common validation, not a copied schema/default resolver. |
| `workbench/experiment` | PASS | Expand, ID, Analyze and result types compile; Run/RunRoot/store/cache/listing use filesystem persistence. Need a bounded in-memory execution path sharing validation, ordering, identities and analysis. |
| `workbench/fecim` | PASS | Current evaluator imports shared peripherals/system, validates design and returns typed results; no direct file access in Evaluate. Strong small-sweep candidate. |
| `workbench/presentation` | PASS | Existing Go display/speakable/availability forms can be used in both builds. |
| `workbench/report` | PASS | WriteJSON/CSV/Markdown are writer-oriented; WriteJSON can serialize to memory. Generate/GenerateRoot write files and stay local. |
| `workbench/study` | PASS | Current orchestration loads/reconciles persistent runs and writes reports. Not the hosted runner as-is. Share the numerical/application pipeline through storage-independent operations. |
| `workbench/edaexport` | PASS | File-generation code can compile, but host filesystem writes remain local-only by product policy. This pass does not move EDA generation or external flows into the browser. |
| `cmd/fecim-lattice-tools` | PASS | Existing non-CGO CLI main links for js/wasm; that does not supply a browser Worker RPC entry point, native server listener, root handles or external processes. A dedicated thin WASM entry point is required. |

Summary: **18 successful targets, one Vulkan-dependent failure, one source-free legacy directory**. All counts refer to this exact scoped matrix, not project-wide coverage or completed UI migration.

## Why filesystem imports can still compile

The statement “filesystem packages cannot compile to WASM” is too broad. Go's js/wasm platform includes stubs/host-mediated interfaces, so packages importing os/io can compile and even execute with Node's filesystem. A normal browser Worker does not thereby receive arbitrary on-disk paths, directory handles, root-confined native access, subprocesses or Vulkan.

Consequently compilation and Node tests do **not** justify calling native Load/RunRoot/Generate/Export from the hosted app or using a virtual native filesystem merely to disguise server-only features. Keep browser studies/results in memory; supply explicitly selected bytes/bundled assets through a data adapter. Stored-run verification, project writes/reuse, EDA and all application-requested file writes remain server capabilities.

Relevant source:

- `shared/crossbar/gpu_mvm_js.go`: js build implementation returns GPU unavailable; native `gpu_mvm.go` is `cgo && !js`.
- `shared/physics/material_config.go:82`, `preisach_calibration.go:219`, `calibration_studio.go`: selected config/calibration paths read/write files; solver availability does not make those paths browser-safe.
- `module1-hysteresis/pkg/headless/data_logger.go`, `cli_calibration.go`: native CSV/calibration persistence and time metadata.
- `shared/logging/logging.go:130-140`: logger is no-op unless file logging is enabled; never enable native file logging in hosted worker execution.
- `shared/crossbar/array.go`, `shared/physics/device_variation.go`, `landau_util.go`: stochastic paths exist. Expose explicit seeds/state and disable wall-clock defaults for reproducible worker requests; do not assume that the deterministic current workbench evaluator proves every stochastic model deterministic.
- `workbench/fecim/evaluator.go:16-70`: CPU evaluation with explicit model version/result metadata; seed is presently unused by this deterministic evaluator, not evidence that all simulations may ignore seeds.
- `workbench/experiment/sweep.go`, `identity.go`, `analyze.go` and `runner.go`: reusable scientific steps coexist with persisted-run orchestration and private successful-result validation. New memory execution must reuse/expose the common validation rather than bypassing it.
- `workbench/report/report.go`: writer serializers are distinct from file generation; report data includes repository revision, so native/WASM parity needs identical provenance, not just equal numerical inputs.

## Selected execution evidence

Existing tests were run unchanged on native Go and through Go's Node js/wasm runner, for `workbench/fecim`, `workbench/experiment` and `workbench/report`.

Common selector:

```text
^(TestEvaluateHZOProducesTraceableMetrics|TestEvaluateRejectsUnsupportedMaterialAsDesignFailure|TestEvaluateHonorsCancellation|TestExpandDeterministicCartesianOrder|TestExpandRejectsInvalidDefinitions|TestIDStableAndBehaviorSensitive|TestIDSortsInputsByPath|TestAnalyzeClassifiesFailuresConstraintsAndPareto|TestAnalyzeMarksMissingAndMismatchedMetricsUnusable|TestAnalyzeHonorsMaximizeAndKeepsEqualPoints|TestAnalyzeConstraintOperators|TestReportsAreDeterministicAndExcludeRuntimeMetadata)$
```

Commands (PATTERN denotes the literal selector above, not a missing repository file):

```bash
CGO_ENABLED=0 go test ./workbench/fecim ./workbench/experiment ./workbench/report -run "$PATTERN" -count=1
GOOS=js GOARCH=wasm CGO_ENABLED=0 go test \
  -exec="$(go env GOROOT)/lib/wasm/go_js_wasm_exec" \
  ./workbench/fecim ./workbench/experiment ./workbench/report -run "$PATTERN" -count=1
```

Observed: **all three packages passed in both executions**. These are existing pure evaluator/expansion/identity/analysis/writer checks. They are not complete package tests, actual Firefox/Web Worker execution, a cross-target byte comparison, browser performance, root-confinement proof, or Orca acceptance. No new test or fixture was implemented.

## Required changes before hosted capabilities can ship

1. **Portable data boundary:** reuse current Go schemas/default validation with bytes/decoded inputs. Bundle approved static citations/material data as build assets or embedded Go data; no fetch of scientific defaults from an uncontrolled service, path-based loader or Typescript copy. User-selected result bundles are bounded, read-only input. Schema checks do not imply stored-run verification.
2. **Portable in-memory study:** share Expand → run ID → Evaluate → validate result → Analyze → WriteJSON with the native path, without disk reuse/commit/report generation. Preserve typed failures, warnings, units, evidence and numerical ordering. Avoid a second scientific runner with divergent validation/defaults. Native file persistence remains its adapter.
3. **Thin WASM main and Worker protocol:** same Go models, typed requests/responses, request IDs, progress snapshots, error categories, deterministic seeds and Go-owned summary/speech values. Main thread only renders and transports data. Go goroutines are not a promise of browser multicore speed. Yield/cooperate on bounded chunks; terminate/recreate worker for a hard abort of ephemeral work. Expose no local-server URL in the hosted transport.
4. **Resource limits:** enforce point count, dimensions, step/sample counts, input sizes and memory-relevant bounds in Go before allocating/starting. Measure and document conservative supported limits that admit the checked-in example; do not invent performance numbers here. Over-limit requests return a readable server-required explanation, not silent downsampling or changed physics.
5. **One React distribution:** two build profiles from the same routes/components. Hosted profile includes only browser adapters and same-origin static assets; local profile uses the same UI plus relative same-origin server APIs. Match compiler/WASM runtime shim/frontend/worker contracts. Replace Astro, do not keep a parallel landing app; no deployment is performed now.
6. **Parity gate:** `validation/wasm/parity_test.go::TestExampleResultsJSONNativeWASM` is now implemented and explicitly run: the native persisted example report, Node Go/WASM and headless Firefox Go/WASM Worker outputs were byte-identical. See the [run instructions](../../../validation/wasm/README.md) and [receipt/budgets](../testing/wasm-parity-gate.md). It compares original bytes, requires successful numerical example results, and diagnoses exact first-field differences without rounding/stripping. This validation-only entry point receives native-resolved input; production browser input/Worker integration and HZO-loop parity remain separate work. The explicit gate must be provisioned/invoked in CI rather than treating its default integration-test skip as a pass. Future relevant source changes require a fresh run.
7. **Dual-build browser/accessibility/security checks:** validate hosted no-localhost/no-upload/no-write behavior, responsiveness/cancellation, local authentication/root scope, and Orca/Firefox task scripts for both capability profiles. A browser-only feature remains unavailable until its bounded execution and accessible equivalents pass; the feature list is accepted scope, not a current support claim.

## Sources and update triggers

Primary runtime guidance: [Go WebAssembly documentation](https://go.dev/wiki/WebAssembly) explains the difference between package objects and executable modules, browser support files, and the Node test runner. Use the matching pinned Go runtime support with the compiler; do not assume a Node filesystem proves browser capabilities.

Re-run this matrix when Go/build tags/dependencies, solver imports, GPU routing, portable input/run adapters or serializers change. Re-run exact report parity for every relevant numerical/identity/serialization/default change and both builds' manual gate when navigation/value speech/focus/capabilities change. No current feasibility finding authorizes native tool installation, code changes, task dispatch, commits or deployment.
