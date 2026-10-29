+++
schema = 1
id = "WEB-HYS-STATUS"
title = "Repair the composed Hysteresis late-Apply status observation"
short = "Scope late-Apply status observation"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 12000
depends_on = ["WEB-AVAILABILITY"]
touches = ["web/test/e2e/hysteresis.spec.ts"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Repair only the existing late-Apply regression's ambiguous first status locator in web/test/e2e/hysteresis.spec.ts. The shared research status precedes the Hysteresis lab status. Observe the existing lab-owned status region; keep both product live regions, delayed real Worker reply, Frequency=2 and stale-draft assertions unchanged. No product code or new HYS chain."
acceptance = ["Given both current profiles and all five Playwright projects, when the existing 'late Apply replies preserve newer input drafts' test runs before/after a scoped lab-status locator repair, then retain expected RED and GREEN showing Settings/state applied, Frequency 2 and stale settings. No assertion, delayed reply or shared status behavior is removed/weakened."]
checks = ["cd web && npm run build:hosted && npm run build:local", "cd web && FECIM_WEB_SUITE=shared npm run test:e2e -- test/e2e/hysteresis.spec.ts --grep 'late Apply replies preserve newer input drafts$'", "cd web && FECIM_WEB_PROFILE=local FECIM_WEB_SUITE=shared npm run test:e2e -- test/e2e/hysteresis.spec.ts --grep 'late Apply replies preserve newer input drafts$'", "cd web && npm run test:e2e -- test/e2e/hysteresis.spec.ts", "cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hysteresis.spec.ts", "cd web && npm test && npm run typecheck", "git diff --check", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["Preserve Go science/defaults/units/speech/identities, original example bytes, tolerances/goldens, CLI and default-built Fyne launch/guards. No hosted upload/persistence/downloads/localhost probing, new product scope, deployment/push, default switch, Fyne deletion, invented qualification/waiver or task receipt.", "No EDA tool/model/dataset/PDK/container acquisition. Use only existing explicitly scoped pinned build/test setup and private assigned ports/roots."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-HYS-STATUS.toml"
+++
- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#outcome-and-definition-of-done; docs/tasks/WEB-FINAL-PLAN.md#web-hys-status--restore-the-composed-late-apply-regression; docs/internals/adr/0005-unified-react-hosted-and-local.md#required-acceptance
- Evidence: WEB-AUTOMATION attempt-1/report.json in execution common slot-data; docs/internals/testing/all-modules-web-checkpoint.md#failure-and-isolated-rerun-evidence; tested 77e23369ab729e1657ab52624a830a21f4f51462. All five projects reproduce both profiles. hysteresis.spec.ts:353 first status resolves App.tsx:50 instead of Hysteresis.tsx:85–92. Private old raw logs are unavailable here; reproduce from source.
- Produces: Existing late-Apply test observes .status-line[role="status"] or equivalent existing lab-owned locator; WEB-RECHECK consumes the repaired spec.
- Repo obligations triggered: Existing browser regression only; no runtime contract/doc/generated/manifest/lock/CI change. Preserve all other test assertions. The runner owns the full merged check; do not duplicate it as a worker.
- Setup: Pinned Go 1.26.10, Node 24.19.0/npm 11.17.0 and locked Playwright per web/README.md. Use private assigned ports and TMPDIR. Existing compatible engines/FFmpeg may need documented pinned build/test provisioning; no optional science/tool acquisition.
- Next action: Run the unchanged exact test on rebuilt profiles and record expected locator RED before the minimal test-only edit, then all Checks.
- Verification status: Proposed/unrun repair. Source identity matches the historical failed checkpoint, not a new browser pass.
