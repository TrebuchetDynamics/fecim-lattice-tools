# Hysteresis final-candidate qualification plan

**Goal:** HYS-REACT; reconcile the completed first batch without widening its behavior or declaring the replacement qualified.
**Authority:** HYS-NEXT-PLAN, round 1 of 2; planning only. The owner chose React primary, Hysteresis first and qualification before separately authorized Fyne retirement. The following decomposition is the planner's call, not new owner direction.
**Queue disposition:** Builder-reviewed branch-local records; return this single planning commit to the existing coordinator after validation. No claim, worker dispatch, implementation, deployment or push in this pass. Committed-store selection can be checked only after this commit; canonical consumer delivery/admission is not inferred.
**Delivery shape:** One evidence/documentation integration and its independent review/merged checks, followed by one final planning handoff. Real dual-profile Orca evidence stays with existing owner task HYS-ORCA and is not a dependency of independent automated qualification. No deletion integration is planned.

## Checklist

- [x] Read task, goal, dependencies, accepted plan/ADR and completed reports.
- [x] Compare reported identities and current source, test seams and GUI consumers.
- [x] Choose a bounded frontier and retain owner-only proof separately.
- [x] Source-reading builder review; resolve material findings before publication.
- [x] Allocate/read back records through autowork tasks and validate; one commit is the authorized handoff.

## Evidence and remaining-proof map

Base: `bb176baec848c4566544087e489145b95c48432b`. Goal record: [HYS-REACT](../../tasks/goals/HYS-REACT.md), still **unmet**. Accepted sources: [behavior/seams/qualification plan](../internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md), [ADR 0005](../internals/adr/0005-unified-react-hosted-and-local.md), [manual gate](../internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md).

Finished reports were read in place under the execution common directory `.git/autowork-execution/slot-data/` in the main repository, not copied into this checkout:

| Existing task / report | Retained outcome |
| --- | --- |
| HYS-CORE/attempt-2/report.json; `655cbb9209234fda105218c554278d8b28295db0` | Stateful native-characterized session and Go presentation; scoped/race tests passed. |
| HYS-PROGRAM/attempt-1/report.json; `8d35d47eb6633f93a2fc3f37bb75e244a11f769e` | Programming/calibration, cancellation and native characterization passed. |
| HYS-ANALYSIS/attempt-1/report.json; `b29261005d3a0930bbdcae74746dec0d8bde6110` | GUI PUND/FORC identity/provenance and bounded results passed; no physical validation. |
| HYS-WORKER/attempt-1/report.json; `f85cc9671cd1c813f16ce85e7bea2210a26c494a` | Three-runtime production parity/lifecycle/admission plus preserved example gate passed. |
| HYS-UI/attempt-2/report.json; `525eed4cbfc589eb8f2932f3e7bb9e6599b357c9` | Both static route profiles, budgets/isolation and three-runtime parity passed at core `e102ef610c8d0ed4b006c1720d2f2a275c885a0fb5f3872e0e48f71b0aba512d`. |
| HYS-LOCAL/attempt-1/report.json; `1cb09424cae1af4b914fdeac443c26f9c3eef0db` | Authenticated embedded files/browser/security tests passed at core `f2fc8fa49afeea96b5bb22e7b89ae4820a30a24879bcc1e24d02c5f52a4590c3`, UI `fde1de15c96828f7c7313e64e816a4dc10dc5957ab4f4e66f04bf58289c3c715`. Report does not include a post-LOCAL enabled production parity or full dual-profile route-budget run. |

All six records retain coordinator-verified closure and existing receipts. This plan neither replaces them nor reimplements their features. Prior passing reports are historical evidence, not newly executed checks.

**Current automated qualification gap:** LOCAL changes `workbench/hysteresis/artifacts.go` (extracts shared `analysisArtifact`, adds browser serialization), shared engine/types/components and build scripts. `web/src/build/build.mjs:17–45` and `validation/wasm/hysteresis_parity_test.go:76–100` hash Go sources, dependencies, driver and exact runtime. UI also hashes frontend/build inputs (`build.mjs:46–65`). Thus the final candidate has different actual bound assets, not merely a renamed commit. Existing receipts do not establish combined final parity, payload/admission and complete-route budgets. Re-run existing gates, do not invent a new gate, threshold or model repair.

**Manual proof:** HYS-ORCA remains open, owner-only, Needs decision. `docs/internals/testing/wasm-parity-gate.md:275–278` and LOCAL pending_live explicitly say both profiles' real Orca qualification is NOT RUN. The future manual receipt does not exist. Automated DOM/focus tests are not a substitute. Do not edit HYS-ORCA or attest it.

**Retirement:** ADR 0005 requires qualified replacement followed by a separately authorized removal. Current production caller `cmd/fecim-lattice-tools/fyne_app.go:20,167–168` imports the GUI and constructs `demo1gui.NewEmbeddedApp()`. `cmd/fecim-lattice-tools/repo_surface_test.go:16` explicitly requires that package; embedded lifecycle/accessibility tests also instantiate it. Other consumers include `cmd/fecim-lattice-tools-fyne/`, `cmd/fecim-screenshotter-fyne/`, `module1-hysteresis/tests/gui_headless_parity_test.go` and GUI characterization tests. Deletion now would break maintained default builds and lose the behavioral oracle. No deletion task is runnable or justified before qualification and a separately authorized caller/test disposition. Round 2 must re-read this graph; it must not change other modules or launch defaults.

**Secondary documentation observation:** UI's report names stale root README claims (README.md:3,216,234). Current source confirms the unimplemented-workbench/separate-Astro descriptions. These are shared-document closeout outside this maintainer's permitted edit paths and not evidence of an absent implementation. No broad docs sweep or second implementation is authorized here. Retain this observation for coordinator closeout; it is not an owner product question or prerequisite to automated evidence.

## Chosen frontier: final-candidate automated evidence

Planner call: reuse the existing enabled gates in one evidence-producing documentation task. Reject immediate deletion (unmet real manual gate/caller contract), speculative product repairs (no demonstrated failing check), and treating old UI receipts as final LOCAL proof (different actual source/UI identities). No external research or new dependency is needed.

### Building record: refresh combined candidate evidence

**Kind:** implementation, documentation/evidence only. **Section:** Next, after the owner's existing Now priorities. **Dependencies:** HYS-UI and HYS-LOCAL, because it consumes their combined final assets. No dependency on HYS-ORCA: this can proceed without a desktop tester.
**Touches:** `docs/internals/testing/wasm-parity-gate.md` only. Generated assets, browser traces and gate transcripts remain ignored outputs/private worker evidence, not tracked deliverables. No source/test/manifest/lock/fixture writes are authorized.
**Scope:** Rebuild both profiles and execute existing production three-runtime parity, Worker lifecycle/admission/budgets and hosted/local complete-route/workflow/isolation plus authenticated embedded local tests on one unchanged source candidate. Append a source-bound combined checkpoint to the existing gate document, preserving historical measurements and real manual limits. Record failures honestly; failed gates become input to round 2, not permission to alter physics or checks.
**Acceptance:**
1. Record candidate revision, source/core/UI/Go/runtime identities, tool versions, original-byte transcript hashes/lengths, actual gate results, selected projects/skips, private evidence references and complete-route byte/time measurements. The explicit enabled parity command must run; default Go-suite skips are not proof. Bind all results to the same source-built profiles.
2. Existing successful finite scientific parity, independent example coverage, lifecycle/admission and payload/route thresholds remain unchanged and must pass for task qualification. Record pass/fail/incomplete distinctly. A failure or missing engine prevents task qualification and a combined qualification claim; retain exact source-bound evidence for the same coordinator's supported maintenance, never close failed checks as a pass.
3. Keep actual dual-profile Orca marked NOT RUN absent real owner evidence; retain Fyne/default-launch, simulation-only and no-deployment limits. Do not claim the goal met or update ledger receipts.

**Repo obligations triggered:** accepted plan's Verification/manual gates; ADR 0005 Required acceptance 1–5; `validation/wasm/README.md` and `web/README.md` setup/identity/test instructions. Existing document owner is wasm-parity-gate.md. Generated outputs: existing ignored build/browser artifacts only. Fixtures/schema/contracts: none changed. TDD N/A: documentation and execution evidence only, no behavior change.
**Existing assertions — keep all:** `validation/wasm/hysteresis_parity_test.go:21–23,59–62` opt-in production gates; `web/test/e2e/hosted-isolation.spec.ts:164–243` actual complete-route measurements and original raw/compressed/time bounds; all `hysteresis.spec.ts` and `hysteresis-local.spec.ts` workflow/identity/security assertions. No assertion replacements or event-handler changes.

- [ ] Inspect repo instructions and current fingerprints; provision only documented locked local test prerequisites (Node 24+, exact go.mod Go version, npm ci, Playwright 1.60.0 browser install/cache per validation/wasm/README.md). No service or host-wide network changes. Use assigned private ports and temporary granted roots.
- [ ] Run `cd web && npm ci && npm test && npm run typecheck && npm run build:hosted && npm run build:local`.
- [ ] Run `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -v` with `FECIM_WASM_ARTIFACT_DIR` set to retained worker evidence and the documented Playwright cache/module if needed; retain actual output and native/Node/Firefox transcript hashes. Do not run only the cheap default skipped gate.
- [ ] Run `cd web && npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/hosted-isolation.spec.ts test/e2e/hysteresis-local.spec.ts` against hosted assets, all configured projects. The embedded test constructs the real local CLI with temporary grants.
- [ ] Run `cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hysteresis.spec.ts`, all configured projects; then `cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hosted-isolation.spec.ts --grep 'complete real route transfer and cold Go sine result meet the production guardrails$' --project=desktop-chromium`. Only that existing budget case applies to local-static assets: the other test requires hosted instruction text/no API requests, whereas the local adapter intentionally requests `/api/v1/roots`. Authenticated files are separately covered above. No test edits are needed.
- [ ] Run `cd web && node src/build/build.mjs --fingerprint`; compare both `dist/engine/manifest.json` and `dist-local/engine/manifest.json` with current generated identity, and embedded asset identity per the existing staging checks. Bind results to those identities, not stale copied outputs.
- [ ] Append the combined checkpoint to wasm-parity-gate.md, with actual failures/skips and NOT RUN boundaries; check relative links and `git diff --check`. Expected outputs are recorded evidence, not a predetermined pass.
- [ ] Review/stage only the owned document and commit once, after applicable repository precommit obligations; report failures without product repairs or task-file edits. The coordinator, not this worker, owns merged verification `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` (validation/wasm/README.md:24); retain that full gate and its default-build/GUI guards in the task Checks.

**Stop/recovery:** A genuine gate failure is a result, not a reason to spend repeated rework on an unknown defect. Record exact failing operation/source/evidence for the successor planner. Setup failure is incomplete, never a pass. Do not relax thresholds, regenerate goldens, alter algorithms, drop modes, fabricate human speech, repair unrelated failures or grant credentials/spending. Source changes while collecting evidence invalidate affected comparisons; rebuild/re-run them or explicitly record incompleteness.

### Successor record: final reconciliation, round 2 of 2

One open plan task under HYS-REACT, depending on the new building record only. Read its report/current source and all remaining goal proof; preserve HYS-ORCA and completed receipts. Map demonstrated failures to bounded repairs, never speculative work; no implementation inside planning. Re-read GUI callers/tests before proposing separately authorized retirement. At round 2 save no successor plan; if the goal is not proved met under ledger rules, preserve an open Needs decision task and write one stop question under `Q-<round-2-ID>` in BLOCKERS.md asking whether/how to continue, naming the remaining proof. No automatic third round.

## Review focus

1. LOCAL fingerprint differs from UI: compare rebuilt identities, never repin old measurements.
2. Expensive Go parity is opt-in: record enabled test results, not default skips.
3. Local complete-route budgets live in hosted-isolation.spec.ts: run its exact budget selector for local, not its hosted-only isolation workflow.
4. Static-local tests do not prove authenticated embedding: run the real embedded local suite too.
5. Automated success cannot supply Orca speech or authorize Fyne removal: retain HYS-ORCA and existing caller/default guards.

## Builder dispositions and planning verification

Fresh-context source-reading review supported this one evidence task rather than speculative repairs or no task: LOCAL changed shared serialization and bound assets after the last full production qualification. Material findings were resolved: local-static must exclude the hosted-only isolation workflow; the full merged check stays coordinator-owned. A second pass found that Playwright matches project/file prefixes as well as title, so the budget selector must not have a leading caret. Read-only `--list` with the corrected local selector and desktop-chromium selected exactly one existing test at hosted-isolation.spec.ts:164. No test implementation was changed. Failed focused checks remain unmet, not an inspection-success loophole.

Saved records: [T-015-ba27c6bd](../../tasks/items/T-015-ba27c6bd.md) (building evidence/documentation), then [T-016-64f7b2b2](../../tasks/items/T-016-64f7b2b2.md) (the only successor plan, round 2 of 2, depends on the entire new building batch). [HYS-ORCA](../../tasks/items/HYS-ORCA.md) is unchanged. Two new records total, both under HYS-REACT; no owner Now priority moved.

Executed planning checks: task consumer reports prepare_api 4 and read-only preparation READY with no in-flight task entries or conversion actions; existing unrelated goal/legacy-reference diagnostics were not a mandate for this scoped batch. Full unchanged-source check `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` passed (exit 0; `/tmp/fecim-hys-next-plan-fullcheck.log`). That default full run does not execute the opt-in browser parity gates; new task's qualification runs remain unrun. TDD N/A: only task documents/metadata change, no production behavior. Final fresh-context review of all four saved artifacts: Approved, no material blockers. Task validation passed with 33 tasks, 8 goals and no drafts; pending entries are the three intentional uncommitted task changes, not qualification evidence. Relative-link/readback/diff checks are run before the one planning commit.

## Handoff limits

At most two new records (one building, one final plan); no unrelated goal edits. After commit the authorized existing runner can continue independently; saving this batch starts no runner and grants no execution authority. Stop the chain by moving its successor planning task to Needs decision through autowork tasks place. Canonical integration/selection/admission remains unassessed until the coordinator consumes this checkout's commit. Planning does not produce a verification receipt or mark HYS-REACT met.
