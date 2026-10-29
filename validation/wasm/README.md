# Native / Node / Firefox WASM qualification

This package has two independent gates. `TestExampleResultsJSONNativeWASM` preserves the existing eight-point persisted-study report check. `TestHysteresisProductionParity`, `TestHysteresisWorkerLifecycle` and `TestHysteresisWorkerBudgets` exercise the actual [production Go entry](../../cmd/fecim-hysteresis-wasm/main_js.go), [Go bridge](../../workbench/hysteresis/bridge.go) and [dedicated Worker/client](../../web/src/worker/hysteresis-worker.mjs). JavaScript transports original JSON strings; it does not compute science or parse/stringify scientific requests/results for comparison. Signed zero is included in the production trace.

## Shared-module qualification helper

[Module support](module_parity_support_test.go), [module_harness.mjs](module_harness.mjs) and the private [module runner](testdata/module_runner/main.go) reuse original Go-generated envelopes, dedicated Workers and original scientific bytes across native, Node and Firefox. Feature producers own their separate parity files/fixtures and may call `qualifyModule` with their real production entries. There is no repeated central module-test registry. Unavailable future ports are not numerical passes.

`TestModuleResearchNativeWASM` runs the actual production research bridge on an admitted eight-point 16/32-row fixture and compares `report.WriteJSON` bytes against persisted native results. It requires eight successful runs with finite metrics. The harness additionally requests real Worker cancellation and checks authoritative settled counts. `TestModuleHostedDependencyBoundary` rejects native filesystem/tool/UI dependencies in the hosted science import graph. The original `TestExampleResultsJSONNativeWASM` remains a separate unchanged gate using the checked-in 64-row example; hosted production caps were not raised to fit it.

Production hosted studies use `project.LoadBytesLimited` and `study.RunMemory`: <=8 MiB input bytes, <=1 MiB per YAML, <=256 points and <=32×32 arrays, with dimension/work preflight before expansion. Canonical results use the original serializer, not parsed/reformatted JSON. Go identity/schema/toolchain/build fields remain authoritative. Source builds discover feature-owned `src/modules/<id>/wasm` assets and bind them with the shared OpenAPI digest in the manifest; see [web conventions](../../web/README.md#shared-module-transport-and-extension-conventions).

These checks do not qualify future module science, deployment, human Orca speech or hardware. They use no hosted upload/persistence/loopback fallback. Real Orca + Firefox/Linux acceptance for both profiles remains **NOT RUN**.

## Run

Supported qualification host: Linux with readable `/proc`, Go matching `go.mod`, Node, the locked `web/` Playwright dependency and its compatible Firefox. Missing enabled prerequisites fail rather than skip. Ordinary tests skip only the browser gates unless explicitly enabled; ordinary full-suite success is not a parity receipt.

```bash
# Provision only if missing; use a worker-owned cache, not a deployment script.
PLAYWRIGHT_BROWSERS_PATH="$AUTOWORK_CACHE/playwright-1.60.0" \
  node web/node_modules/playwright/cli.js install firefox

PLAYWRIGHT_BROWSERS_PATH="$AUTOWORK_CACHE/playwright-1.60.0" \
FECIM_WASM_PARITY=1 FECIM_WASM_ARTIFACT_DIR="$TMPDIR/wasm-evidence" \
CGO_ENABLED=1 go test ./validation/wasm -count=1 -v

go test ./workbench/hysteresis -count=1
CGO_ENABLED=1 go test -race ./validation/wasm ./workbench/hysteresis -count=1
go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric -count=1
go test ./validation -run '^TestPhysicsRegressionCurves' -count=1
```

Use `autowork-run timeout 600 ...` for long checks under Autowork. The merged-tree full check belongs to the coordinator, not this worker. Private servers use `AUTOWORK_PORT_BASE` and `AUTOWORK_PORT_BASE+1` when set; standalone runs without it use ephemeral loopback ports. No host-wide networking changes occur.

Overrides:

- `FECIM_PLAYWRIGHT_MODULE`: absolute path to an existing locked Playwright `index.mjs`; default `web/node_modules/playwright/index.mjs` from the repository root.
- `FECIM_FIREFOX_EXECUTABLE`: explicitly selected Playwright-compatible Firefox executable; actual path/version are recorded. An arbitrary system Firefox need not support Playwright's protocol.
- `PLAYWRIGHT_BROWSERS_PATH`: installed browser cache directory.
- `FECIM_WASM_ARTIFACT_DIR`: retained evidence parent; otherwise Go removes temporary artifacts at test completion.

The Linux command wrapper owns only its subprocess descendants, including detached Firefox groups, on timeout. This is best-effort cleanup, not adversarial containment. Each browser client has startup/request deadlines and explicit termination. Nothing targets an existing user browser.

## Production protocol and ownership

The authoritative schemas and **sole operation registry** are `Envelope`, `Reply`, `Identity` and `Operations()` in [bridge.go](../../workbench/hysteresis/bridge.go). `session.new` returns the Go defaults/catalogue, registry and outer work/envelope limits. Producer payloads retain their Go field names and SI units; do not add UI conversions/defaults or a second registry.

Every envelope requires `identity`, `session`, `request`, `revision`, `operation` and `input`. Identity has `protocol=fecim-hysteresis/1`, `model=hysteresis-producers/1`, source-derived `build` and exact `go` toolchain version. Request IDs are positive monotonic safe JSON integers; initial transport revision is zero. Session IDs have 1–128 bytes. Seeds above `2^53-1` are rejected, not rounded. `input` is `null` for payload-free operations.

| Operation | Input / result |
|---|---|
| `session.new` | Complete `Config`, or `{}`/`null` for Go defaults; returns core snapshot, registry and limits. One session per Worker. |
| `core.snapshot`, `core.apply`, `core.advance` | `null`, `Command`, or `{Steps, DT}`; core snapshot. |
| `core.page` | `{Offset}`; Go binds the producer revision to the admitted transport revision and returns at most 32 rows. |
| `program.new`, `program.apply`, `program.advance`, `program.snapshot` | `ProgramConfig` (or `{}` defaults), `ProgramCommand`, `{Steps, DT}`, or `null`; programming snapshot. |
| `analysis.new` | `{Kind, FORC, References}`; selected context is copied from Go's current session, never caller-supplied material parameters/snapshot. |
| `analysis.advance` | `{Steps:1}`; bounded progress with state/completed/total. |
| `analysis.page`, `analysis.result` | `{Table, Offset}` or `null`; revision-bound rows or complete independent diagnostic. Incomplete result is an error. |
| `job.cancel` | `null`; truthful cancelled active analysis/program/core snapshot. |

Each admitted reply increments the **transport** revision, including reads. Producer revisions inside snapshots remain scientific state identities. Rejected invalid requests leave state, transport revision and replay barrier unchanged. A producer terminal error that changes state (for example work-limit exhaustion) returns `error`, the retained terminal `result`, an incremented revision and `committed=true`; never treat it as scientific success. All successful replies have `committed=true`. Errors without mutation have `committed=false`. Missing/null initial revision, unknown/duplicate fields, unsupported operations, malformed/trailing/non-finite JSON, stale IDs and oversize requests are rejected by Go before producer state allocation/mutation. Reference-array token preflight enforces at most 32 branches before decoding the typed slice, so many tiny/null entries cannot amplify allocation past that known cap.

Only one active computation is admitted. Pause/cancel core/program before starting an independent analysis; cancel a running analysis before replacing it or driving core/program. Read-only pages remain available. Analysis reports its actual diagnostic model/profile/applied inputs separately from selected loop context. Time-resolved programming returns `legacy_protocol_unavailable`, an explanation and no fabricated switching plot. References remain `imported_unverified`, not measured/verified device evidence.

The client is `globalThis.HysteresisWorkerClient` after importing the Worker file on Window. Construct it with same-origin `workerURL`, `wasmURL`, `runtimeURL`, `workerSHA256`, `wasmSHA256`, `runtimeSHA256`, `identity` and `session`; await `ready`. `request(operation,input)` builds transport metadata; `send(originalJSONString)` preserves exact caller bytes. Both serialize outstanding requests and return the original Go reply string, including Go errors. `terminate()` discards ephemeral state and rejects pending work. `recreate()` starts a new verified Worker; explicitly create a new Go session afterward. Sends bind their generation before awaiting readiness, including same-turn cancel/recreate at revision zero. Retired generations and stale/mismatched replies cannot occupy or publish into the replacement.

Worker, module and runtime bytes are verified before use; cross-origin assets and redirects are rejected. Verified Worker/runtime bytes execute via Blob URLs, so the host CSP must permit `worker-src blob:` and runtime scripts from `blob:` plus its Go/WASM compilation policy. No `eval`, scientific upload, persistence, localhost discovery or native fallback is provided. Transport `ready/running/failed/cancelled` is not a substitute for the producer's scientific state; the UI must render the returned state and stale-result identities.

## Build coherence and artifacts

Unpinned builds (`BuildID=unqualified`) reject operations. The qualification builder in [hysteresis_parity_test.go](hysteresis_parity_test.go) hashes sorted tracked/untracked non-ignored Go source paths/bytes plus `go.mod`, `go.sum`, `config/constants.yaml`, `config/materials.yaml`, the production Worker bytes and matching toolchain `wasm_exec.js` bytes (portable label `wasm_exec.js`), the generated module OpenAPI, embedded research YAML and module/API build scripts. It uses `-buildvcs=false -trimpath -ldflags='-s -w -X fecim-lattice-tools/workbench/hysteresis.BuildID=DIGEST'` for **both** the native and `GOOS=js GOARCH=wasm CGO_ENABLED=0` entries. VCS-derived module/version stamps are excluded so committing or merging unchanged sources cannot silently change the pinned assets. No target-specific input/output normalization is permitted. Reuse that byte inventory for packaging; a repository revision alone does not identify a dirty build. Use `go env GOROOT` → `lib/wasm/wasm_exec.js` from the same Go toolchain, never a downloaded unrelated shim. `measurements-*.json` records all three asset hashes and the full identity for a host manifest.

Production catalogue bytes are compiled from the existing canonical material/constants YAML by [config/hysteresis.go](../../config/hysteresis.go). This avoids native filesystem overrides versus older WASM fallback defaults. CLI/Fyne filesystem loaders and presets are unchanged. The bounded shared repair uses the same portable Go exponential evaluation in native/WASM Preisach/NLS, reversible Tanh and GUI FORC Gaussian weight; equations, parameters, diagnostic profiles, history/guard semantics, tolerances and goldens remain unchanged.

Retained production runs contain:

- `native-input.json`: exact request array used by the native entry; `input.json`: the same original request strings plus Go-defined cold Sine schedule for JS transport.
- `hysteresis-native`, `hysteresis.wasm`: real production entries, not the old private example runner.
- `native-results.json`, `node-results.json`, `firefox-results.json`: original byte-comparison transcripts, with a guard requiring complete programming/calibration, finite diagnostics, maximum FORC, signed rows, unavailable status and independent profile/reference identities.
- `measurements-parity.json`, `measurements-lifecycle.json`, or `measurements-budgets.json`: identity, asset hashes/sizes, host/tools, per-dispatch inputs/timings, lifecycle outcomes and cold-transfer records.

The separate example run still retains its resolved bundle, native project/reports, private `study.wasm`, three original `results.json` outputs and `measurements.json`. It requires eight successful runs and four finite metrics each; matching failures/empty reports do not pass.

## Budgets and limits

`TestHysteresisWorkerBudgets` gates production module/runtime/Worker totals at **6 MiB raw, 1.5 MiB Brotli and 2 MiB gzip**. Compression is per emitted asset (Brotli quality 11, gzip level 9). The Worker-driver cold marker starts before importing the client and ends after one complete Go-default signed Sine cycle (64 physical steps), explicit pause and its first 32-row Go table page. Three fresh Firefox contexts/Workers, `no-store`, actual Brotli responses, 100 ms added request RTT and one aggregate **10,000,000 bit/s** server-owned chunk schedule are used. Concurrent assets share bandwidth. The byte filter includes client and Worker script fetches, module and runtime; excludes only the empty HTML navigation before the marker. Each cold result must finish within **5 seconds**.

The production budget run also gates per-dispatch wall time at **1,000 ms**, cancellation at **500 ms**, recording every registry operation. These engineering admission bounds are not the old example-only 250 ms budget, whole-job completion limits or real-time guarantees under arbitrary host scheduling. The measured schedule covers 256 outer core/program steps, 128 programming domains, 100,000-sample allocation, one analysis unit/call and complete 64-reversal/256-resolution FORC. Go also caps supplementary plots at 2,000 points, pages at 32 rows, request/reference envelope at 1 MiB and lifetime solver/calibration work using producer limits. LK calibration and continuous write/read are measured for bounded advance/cancel, not falsely described as complete converged jobs; completed target-write and Preisach calibration are separately required.

No CDN, low-end-device, React route, paint, accessibility or hardware claim follows from these private measurements. HYS-UI must independently gate the actual complete React route (including its 2 MiB Brotli transfer target); real Orca/keyboard qualification remains separate. See the [production receipt and historical example evidence](../../docs/internals/testing/wasm-parity-gate.md).
