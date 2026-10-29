+++
schema = 1
id = "WEB-MNIST-WAIT"
title = "Bind MNIST browser assertions to the matching refreshed result"
short = "Wait for matching MNIST results"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 13000
depends_on = ["WEB-MNIST"]
touches = ["web/test/e2e/mnist.spec.ts"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Harden only web/test/e2e/mnist.spec.ts against prior retained results between terminal job publication and snapshot refresh. Preserve real local science/native oracle, hosted isolation, exact equality/provenance/counts, export, invalid-model retention, one session, lost-poll/start-receipt reload/reconnect/cancellation assertions. Add a deterministic delay of a real terminal snapshot response within the same test, then wait for requested operation/input result rather than generic succeeded. No product code, arbitrary sleeps, weakened assertions or scientific repair."
acceptance = ["Given the existing local MNIST test and a test-only held real terminal snapshot response, when sample-17 inference and three-image evaluation follow retained prior results, then RED exposes premature observation and GREEN waits for the matching mnist-result before reading r. Preserve exact native inference and both confusion-matrix equality, preprocessing none, provenance hashes and table counts.", "Given every original hosted/local boundary, retention/export and recovery assertion in mnist.spec.ts, when all five projects and repeated compact/small-mobile runs execute, then those assertions remain and pass without mocked science, sample reduction, tolerance changes or product edits."]
checks = ["cd web && npm run build:hosted && npm run build:local", "cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/mnist.spec.ts --project=desktop-compact-chromium --project=mobile-small-chromium --repeat-each=3", "cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/mnist.spec.ts", "cd web && npm run test:e2e -- test/e2e/mnist.spec.ts", "cd web && npm test && npm run typecheck", "git diff --check", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["Preserve Go science/defaults/units/speech/identities, original example bytes, tolerances/goldens, CLI and default-built Fyne launch/guards. No hosted upload/persistence/downloads/localhost probing, new product scope, deployment/push, default switch, Fyne deletion, invented qualification/waiver or task receipt.", "No EDA tool/model/dataset/PDK/container acquisition. Use only existing explicitly scoped pinned build/test setup and private assigned ports/roots."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-MNIST-WAIT.toml"
+++
- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#outcome-and-definition-of-done; docs/tasks/WEB-FINAL-PLAN.md#web-mnist-wait--observe-the-matching-retained-result; docs/internals/adr/0005-unified-react-hosted-and-local.md#required-acceptance
- Evidence: WEB-AUTOMATION attempt-1/report.json; docs/internals/testing/all-modules-web-checkpoint.md#failure-and-isolated-rerun-evidence; tested 77e23369ab729e1657ab52624a830a21f4f51462. Compact fails equality then missing cases[0]; small-mobile fails then passes alone. Screen.tsx:130–140 publishes terminal job before await a.refresh(); mnist.spec.ts:142,159,170 waits only generic succeeded. This is not proof of numerical mismatch.
- Planner decisions: Retry result-bound assertions and settle preceding model/config/input actions by their matching observable/re-enabled action. Delay a real route.fetch response, never fabricate scientific replies; finite release/cleanup on failure. Repeat-each=3 is verification, not a product target. Exact seams and conservation list are in the plan.
- Produces: Deterministic existing real MNIST boundary test; WEB-RECHECK consumes the repaired spec.
- Repo obligations triggered: Test-only observation changes, no new interaction/value/API contract, docs/fixtures/generated/CI/manifest/lock edits. The runner owns the full merged check. Newly demonstrated product failures return exact evidence to the same coordinator; no speculative repair.
- Setup: Pinned Go 1.26.10, Node 24.19.0/npm 11.17.0 and locked Playwright per web/README.md. Use private assigned ports and TMPDIR. Existing compatible engines/FFmpeg may need documented pinned build/test provisioning; no optional science/tool acquisition.
- Next action: Reproduce local compact/small-mobile failure and make the real-response delay deterministically expose the premature read RED; implement result-bound waits and run Checks.
- Verification status: Proposed/unrun; prior isolated success is not aggregate qualification.
