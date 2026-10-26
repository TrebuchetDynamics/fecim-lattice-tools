# Remaining local work — integration verification

Owner request: **“Merge all main.”** Integration starts from public `b0fb21e2`, brings in local main `f9fea3ef` plus its eight pending source/document changes (recorded in `af67f30e`), and preserves the remaining `autowork/HZO-REF-4` history at `e60a3e4f`. No force push or history rewrite is needed.

## Conflict policy

- Preserve public Go **1.26.10**, patched dependency versions, Astro **7.4.2**, Wayland build packages, the Fyne/GLFW link guard, and screenshotter deferred display cleanup plus regression tests.
- Preserve local configuration editing/root-scoped loading, progress/partial outcomes, scientific display/speech values, legacy terminal presentation, HZO extraction/comparison/export, widget timing correction, scheduled broad physics race lane, and React-primary documentation/skills.
- Keep the stronger integrated `validation/peloop` comparator/baseline from `69eda64c`, not the alternate runtime implementation from `e60a3e4f`. Independent review confirmed its arithmetic fixes are preserved. Restore the alternate branch's unique 28-point huge-residual, finite-wide-span remanence, and five-point numerical regressions in `legacy_branch_regression_test.go`, retaining the existing numeric literals rather than regenerating a golden.
- `go mod tidy` makes the existing patched `x/sys v0.48.0` a direct dependency for the Linux atomic export publisher; no dependency version is lowered.
- Align the newly merged Nightly physics job with public action versions, Go selection from `go.mod`, and Wayland development dependencies. Preserve its 90-minute job budget, 75-minute step, non-short suite, failure-preserving log pipeline and always-uploaded logs. Execution of that full scheduled lane is not implied by a workflow contract test.

## RED → GREEN corrections

1. **New initial-layout fixture:** `xvfb-run -a go test -race ./cmd/fecim-lattice-tools -run '^TestAllModulesInitialLayoutIsStable$' -count=1 -timeout=3m` failed with EDA background status updates racing headless-driver construction/layout/cache operations. Fyne's test driver executes `Do` inline, whereas desktop dispatch queues work. Production already uses `SafeDo` with the shared UI lock. Correct only the new fixture: lock app creation/reset, complete construction/layout/refresh, and close; keep observation time outside the lock. `-count=10` now passes. Independent review reproduced RED and independently verified the correction over three repetitions. This is not a claim that a new production EDA race was fixed.
2. **Digitizer interoperability:** added ascending/descending reversal fixtures in pixel-index order to library and real-command tests. Both initially returned successful reversed physical CSV that the comparator correctly rejected. Require field direction when emitting contiguous indices, without sorting fields or changing reference samples. Focused tests now pass and command failures create no output file.
3. **Restored branch tests:** the three numeric/overflow tests are recovered existing regression coverage, not new model behavior. They pass with the stronger integrated implementation. Reference/data CSV bytes, physics coefficients and tolerances are unchanged.

Implementation commits already contain scoped RED/GREEN evidence for the existing workbench/HZO changes; task evidence TOMLs alone contain abbreviated/placeholder command paths and are not substituted for fresh verification. Documentation, ignore policy, merge-only resolutions and test-fixture corrections do not introduce simulator behavior.

## Combined-tree verification

- `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`: **PASS**.
- `make ci`: **PASS**, including `govulncheck` (**no vulnerabilities found**), short Go suite, build/vet/architecture and research checks.
- `go test -race ./validation/peloop ./cmd/pe-reference ./workbench/... -count=1`: **PASS**.
- `go test -race -short ./shared/... ./validation/...`: **PASS**.
- Display-tagged screenshotter suite: **PASS**, retaining the public all-module and cleanup checks.
- New initial-layout race selector: **PASS**, ten repetitions.
- Adapter generator sync and all eight installer fixture tests: **PASS**.
- `go mod verify`, whitespace checks and Nightly workflow contract: **PASS**.
- Public static-site manifests/configuration and ordinary PR CI workflow are unchanged by this integration. Their security/build/browser evidence from Goal 1 remains applicable; hosted CI rechecks the static build.
- Two independent read-only reviewers found no remaining blockers in core/security/lifecycle and HZO source/data/metric/export scopes after the two corrections. Scientific review independently checked all 26 affine sample conversions and three source receipt hashes. Remaining low-priority coverage note: an already-delivered-record corruption reconciliation case lacks a dedicated test; no defect was demonstrated.

Local logs: `/tmp/fecim-all-main-*`. Before declaring delivery complete, verify every selected branch tip is an ancestor of public main and require green hosted PR/main CI. Merge-only ancestry commits may reuse these checks only when their resulting tree is byte-identical to the verified tree.

## Hosted legacy-fixture follow-up

The first PR #24 CI run (`40264618783`) passed `make ci` but failed the additional legacy Fyne target with a MNIST rich-text panic. A focused existing lifecycle selector reproduced data races within one MNIST instance: direct `fyne.Do` completion/heatmap callbacks bypassed the shared test UI lock and raced other status updates. Per-process isolation alone would not fix that.

Correct only `embedded_module_lifecycle_test.go`: run each original module case in a timeout-bounded child using the same test binary (including race instrumentation), install a test app/driver adapter that routes **all** `DoFromGoroutine` callbacks through the existing re-entrant UI lock, and lock fixture construction/layout/close. Construct only the selected module and avoid `NewTempApp`'s unprotected global-cache reset. Keep the original real lifecycle calls and two-worker leak allowance; `Stop`, observation and leak polling remain outside the UI lock to avoid blocking workers that need it. Process exit contains allowed late workers; it is not evidence that production workers joined.

RED: `xvfb-run -a go test -race ./cmd/fecim-lattice-tools -run '^TestEmbeddedModulesStartStopDoNotLeakWorkers$' -count=3 -timeout=3m` failed. GREEN with the corrected fixture: the same three complete lifecycle runs pass (`-timeout=5m` accommodates seven contained cases per run). Production MNIST lifecycle/dispatch, leak bounds, and application behavior are unchanged. Fresh full xvfb suite, `make test-next-ui`, and `make ci` all pass on this corrected tree. Independent read-only review found no blocker and independently passed the selected MNIST child under the race detector. This follow-up changes the previously frozen tree; the ancestry-only tree identity assertion is not reused for it.

## Scope and non-claims

No release, DOI, JOSS submission, outreach or complete React client is created by this merge. Scientific display strings are not an Orca pass. The HZO comparison retains `experimental_validation=false`; no reference-runner experimental-validation count is promoted. Independent visual/source-rights confirmation remains a prerequisite to public scientific promotion.

The separate `review/concurrent-research-2026-10-25` worktree contains an uncommitted, explicitly stopped literature-research session (`citations/disputed.md` and `citations/reports/`). Its raw tool logs, job lock, local diagnostics and incomplete research outputs are not branch commits and are not silently published by a source-history merge. They remain intact for that workflow's owner review. Ignored local agent/editor state, secrets and generated artifacts are likewise excluded.
