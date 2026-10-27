# Native / Node-WASM / Firefox-WASM parity gate

This validation-only gate runs the checked-in eight-point example through the existing native `study.Run` persisted-report path and through the same Go numerical/identity/analysis/report functions compiled to WASM, once in Node and in a real dedicated headless Firefox Worker. It compares the original `results.json` bytes, not parsed/rounded JS numbers, and reports the first differing JSON field. It also records raw/gzip/Brotli sizes and Worker startup/run timings.

No React application, production Worker API, browser project loader or physics implementation is added here. The private `testdata/runner` receives a bundle already resolved by the real native Go loader. It tests cross-runtime numerical/report execution; browser YAML/input validation remains separate work. The successful example must still contain eight successful runs with its four finite numerical metrics; identical failures or empty reports are not a numerical parity pass.

## Run the required gate explicitly

Supported gate host: Linux with readable `/proc`, Go matching `go.mod`, Node, the locked Playwright dependency from `web/`, and an installed Playwright-compatible Firefox. No new dependency installation was performed by the gate author. To provision a fresh test host, review/install the existing web lockfile dependencies and Playwright Firefox first; do not launch a deploy script.

```bash
FECIM_WASM_PARITY=1 \
FECIM_WASM_ARTIFACT_DIR="$PWD/artifacts/wasm-parity" \
CGO_ENABLED=1 go test ./validation/wasm -count=1 -v
```

Optional overrides:

- `FECIM_FIREFOX_EXECUTABLE`: an explicitly selected existing Firefox executable. Its actual path/version are recorded. Default is Playwright's matching bundled Firefox. An override is not a claim that every driver/browser combination is supported.
- `FECIM_PLAYWRIGHT_MODULE`: an existing Playwright `index.mjs`; default `web/node_modules/playwright/index.mjs` relative to the repository root.
- `FECIM_WASM_ARTIFACT_DIR`: parent directory for retained per-run receipts. Without it, artifacts use Go's temporary test directory and are removed after the test, including on failure.

Missing enabled prerequisites fail; they do not turn into passing skips. In ordinary `go test ./...`, the expensive three-runtime test is explicitly **skipped** unless `FECIM_WASM_PARITY=1`; its diagnostic/numerical-coverage/process-ownership regressions still run. A normal full-suite pass therefore does not establish parity. CI must invoke the explicit gate with browser provisioning before enabling or shipping hosted numerical features; this change does not silently edit CI or claim it is already enforced there.

## Evidence and limits

Each retained run contains:

- `native-project/`: source YAML copy and actual native committed runs/reports;
- `input.json`: resolved bundle plus identical repository revision for both WASM executions;
- `native-results.json`, `node-results.json`, `firefox-results.json`: unmodified comparison bytes;
- `study.wasm`: optimized validation entry point (`-trimpath -ldflags='-s -w'`);
- `measurements.json`: actual runtime/browser/tool versions, host, payload sizes and timings.

The JS driver only transports JSON, launches the browser/Worker and measures bytes/time. Physics, defaults, run IDs, feasibility/Pareto and report serialization stay in Go. An equality result is not an independent scientific correctness proof. This bounded example does not certify all models/materials, actual HZO-loop parity, production admission/cancellation, schema security or Orca behavior.

The Linux command wrapper owns subprocess descendants, including detached Firefox groups, on timeout; a focused regression verifies that a deliberately detached child does not survive cancellation. This ownership is limited to commands launched by this gate and never targets existing user browsers. Linux process enumeration is best-effort cleanup, not adversarial process containment or a guarantee against every possible crash/reparenting race. The Worker has a separate timeout and ordinary cleanup. The gate currently fails clearly rather than pretending this process-tree contract exists on other operating systems.

See [implementation receipt, GPU audit and proposed budgets](../../docs/internals/testing/wasm-parity-gate.md) and the [HZO-only first-slice plan](../../docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md). No commits are authorized until owner review.
