# WEB-WORKBENCH: CI contract review finding

## Disposition

Restate the existing [WEB-WORKBENCH](../../tasks/items/WEB-WORKBENCH.md) task; do not split or close it. Preserve its Scope, all three Acceptance items, Dependencies, section/rank, resources and authorization. Add only the missing CI test path to Touches, focused/full regression Checks and an explicit review-finding note. This is the same shared CI/generated-contract outcome, not a new product requirement or a MAIN repair task.

At inspection, the common-Git-directory `autowork-execution/slots/WEB-WORKBENCH.json` reports `held`, cause `validation`, not a running or integration-ready candidate. Its recorded worker PID is absent. The retained candidate and failed integration remain untouched; no attempt, hold, review budget or runtime state is reset by this documentation change.

## Evidence and diagnosis

- The [accepted roadmap](../internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#protocol-and-registration-ownership) assigns generated OpenAPI/TypeScript contracts and shared CI to WEB-WORKBENCH. Its Resource, build and security guardrails require regeneration/drift rejection and both profiles. The [browser design](../internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md#stack-contracts-and-build-policy) and [ADR 0005](../internals/adr/0005-unified-react-hosted-and-local.md#required-acceptance) agree. No owner product decision is needed to reconcile the matching CI regression test.
- Worker report: common-Git-directory `autowork-execution/slot-data/WEB-WORKBENCH/attempt-2/report.json`, candidate `cd35b12adb7af4d1f9ead00eda023fef45a28f40`. It reports focused workbench, browser/build and parity checks passing, but does not report the full repository check or `TestCIRunsWebTestsAndProductionBuild`. Those claims are retained evidence, not qualification by this maintainer.
- Review finding: runner-state `review-findings/WEB-WORKBENCH.json` identifies failed integration `49f35e2e0071898b77ac8506e150402b5a6385af`. Runner-state paths here are relative to the existing `bc334d016fa9179d` state directory, not repository files to copy or commit.
- Runner-state `full-check/tree-3fa700f47d3b6b5fbea084c0972ad603fbcb68e1.json` records the full check failing only `TestCIRunsWebTestsAndProductionBuild`, versus passing baseline `b8be83f73d671aef339b7565d6cc185f1ff9ba7a`. `full-check/WEB-WORKBENCH-merged.log:88–92` reports `ci_contract_test.go:334: web CI command="npm ci && npm run generate:api && npm test && npm run build" working-directory="web"`. Reruns JSON is incomplete with no flaky classification; its rerun log is empty. Do not call this a flaky test.
- At the baseline, `scripts/ci_contract_test.go:325–343` parses `.github/workflows/ci.yml` and requires exactly `npm ci && npm test && npm run build`, working directory `web`, and no `continue-on-error`. The candidate adds `npm run generate:api` between install and tests in that step, plus a required generated-drift step, without changing this Go test. It therefore fails deterministically even though the required install/test/build operations remain present. `scripts/ci_contract_test.go` is missing from the task's original Touches.
- Fresh read-only reproduction at the clean retained candidate: `env CGO_ENABLED=1 go -C <retained-worktree> test ./scripts -run '^TestCIRunsWebTestsAndProductionBuild$' -count=1` exits 1 with the same command mismatch. No source file was edited to reproduce it.

## Worker continuation

1. Retain the complete original task contract and all existing checks. Reproduce the named CI test failure on the implementation tree before changing tests or workflow.
2. Reconcile `scripts/ci_contract_test.go` with the accepted generated-API CI pipeline. Preserve required locked installation, real unit tests and production build in `web`, fail-fast behavior and no failure masking. Keep generation and generated-drift rejection, both profile builds and independent HYS/scientific gates. Do not delete/skip the test or remove required CI behavior just to recover its old literal string.
3. If replacing literal equality with a semantic helper, add failing mutation tests before implementing it: missing install/test/build, wrong directory and optional/masked failures must still be rejected. An exact updated pipeline assertion is also sufficient for this bounded mismatch; do not invent a general shell parser.
4. Run `env CGO_ENABLED=1 go test ./scripts -count=1`, all original task Checks and `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`. Record RED/GREEN and the full result. Source inspection or this task edit does not certify the implementation or manual Orca gate.

## Ledger proposal conservation

No saved ledger proposal is identified for this item: the retained slot has `ledger_edit_dropped: null`, and `git diff cd35b12a^ cd35b12a -- tasks/items/ tasks/goals/` is empty. The candidate's older parent predates the baseline's T-012 closure, so a baseline-to-candidate diff would misleadingly include unrelated task/evidence differences. Do not replay that diff or reopen the coordinator-verified T-012 record. No candidate task patch is applied; current closures, planning and goal status are preserved.

## Maintainer verification

TDD: N/A — task metadata and this source-backed note only; no production, test or workflow code changed. The future worker must follow RED/GREEN for its CI regression repair.

- `tasks validate --worktree .`: PASS, 44 tasks, 10 goals, no drafts; only WEB-WORKBENCH is a pending modified record before commit.
- `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`: PASS (exit 0) on this documentation-only checkout based on `b8be83f7`, including `fecim-lattice-tools/scripts`. This is baseline regression verification, **not** GREEN for the retained candidate; its focused reproduction remains RED until a worker repairs it.
- Conservation/readback: PASS. Parsed headers differ only by the appended Touches path and two Checks; every original item and prior note is preserved. All other tracked files, including goals and closed task records, are unchanged. All local Markdown targets/anchors resolve; `git diff --check` passes. No runner evidence bundle is copied into the checkout.
- Independent fresh-context read-only review: no blocker; ready for this task/documentation commit, not acceptance of the retained implementation. Candidate product behavior, scientific parity, browser/accessibility qualification and deployment were outside that review and remain unqualified here. A passing baseline check cannot qualify the failed candidate.
