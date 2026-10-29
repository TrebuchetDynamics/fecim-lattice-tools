# WEB-CIRCUITS / WEB-MNIST overlap disposition

DOCS-MAINTAINER inspection at `a535c0c8a2a2d13c4b4d6ebce49a2071998af9a7`.

## Checklist

- [x] Read both task records, their three Sources and the accepted module roadmap.
- [x] Inspect worker evidence and live-slot ownership without changing runtime state.
- [x] Compare declared overlap with actual assertions and installed admission policy.
- [x] Preserve contracts where no source-backed repair is justified.
- [x] Validate the ledger, documentation links, unchanged contracts and repository full check; obtain independent review.

## Decision

Leave [WEB-CIRCUITS](../../tasks/items/WEB-CIRCUITS.md) and [WEB-MNIST](../../tasks/items/WEB-MNIST.md) unchanged. Neither requires the new Crossbar web port as a product prerequisite. Both legitimately declare `shared/viewmodel/e2e_all_modules_workflow_test.go`; removing it or disguising it as a directory/glob would hide required work, not repair a contract. Preserve `depends_on = ["WEB-WORKBENCH"]`, every Acceptance/Check, placement, resources and authorization. No split, merge, supersession or closure is justified by overlap alone.

The reported wait is reproducible from the installed scheduler's file classification and the live Crossbar slot. It is not evidence that these contracts are unrunnable or require a new producer. This audit does not bypass admission or promise dispatch; admission must be reconsidered when the conflicting work lands.

## Evidence

- The [accepted roadmap](../internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#evidence-and-verification-matrix) explicitly requires updating affected `shared/viewmodel/e2e_all_modules_*` assertions when replacing placeholder behavior, preserving action/catalogue/resilience/evidence invariants. Its delivery map and Sources/trade-offs section make module ports parallel after the shared foundation: source reuse is not a dependency on another module's new screen. Its historical description of these overlaps as cold does not override installed runtime admission.
- The actual [shared workflow test](../../shared/viewmodel/e2e_all_modules_workflow_test.go#L17) contains separate Crossbar, MNIST and Circuits rows in one workflow matrix (lines 30–32). MNIST currently invokes inference/sweeps and requires accuracy/correct metrics; Circuits invokes READ/WRITE/COMPUTE; Crossbar invokes MVM/IR. Those are exactly the adapter behaviors the roadmap requires repairing. The file also holds composition/action/error guards; preserving them matters. No source supports removing the file from either contract or relocating tests merely to lower heat.
- [WEB-CROSSBAR](../../tasks/items/WEB-CROSSBAR.md) is the third unfinished record naming that literal path. At inspection, common-Git-directory `autowork-execution/slots/WEB-CROSSBAR.json` reported `running`, attempt 1, baseline `a535c0c8`; its attempt directory had no completed `report.json`. Neither WEB-CIRCUITS nor WEB-MNIST had a slot or slot-data directory, so neither has a worker report supplying a narrower required file set. Do not edit the running Crossbar entry or infer completion from its live log.
- Installed release `c82163f5806498c6d7559f1e8feb5384d46a5845`, `Src/pi_autowork/engine/parallel_plan.py:13–23,329–342,365–387`, counts distinct unfinished contracts, sets the hot threshold to three and waits on shared hot files against active/unlanded work. Its `test_module` exemption recognizes `tests/` directories and Python test names, **not Go `_test.go` names**. Thus this file is hot under the installed classifier despite being a test and despite the roadmap's cold-overlap wording. This classification discrepancy cannot be repaired honestly by narrowing task Touches; scheduler code is outside this docs-only remit. Cold-path ranking, held-peer nonreservation, genuine prerequisites and resource/protected-path admission remain separate rules.
- [WEB-WORKBENCH](../../tasks/items/WEB-WORKBENCH.md) is `done` with `coordinator_verified` closure and its retained receipt. `tasks why WEB-CIRCUITS .` and `tasks why WEB-MNIST .` each reported committed-store `eligible: true` at this baseline, explicitly with `admission_evaluated: false` and `runtime_evaluated: false`. Eligibility is not dispatch or admission proof. No source-backed dependency change follows from the shared path.

## Verification and limits

Final `tasks validate --worktree .`: PASS, 44 tasks, 10 goals, no drafts or pending records. All six local Markdown link targets resolve; `git diff --exit-code a535c0c8a2a2d13c4b4d6ebce49a2071998af9a7 -- tasks/` confirms task/goal/evidence bytes are unchanged. `git diff --check`: PASS. Repository full check `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`: PASS, exit 0. A fresh-context read-only reviewer reproduced both installed hot-file waits and found no material blocker or justified contract repair. Live-slot observations describe inspection time only; they are not durable ownership or completion receipts.

TDD: N/A — this audit note changes no product behavior, task execution contract or scheduler code. No product, scientific, browser or manual qualification is inferred from documentation review or ledger validation. No task-file patch was applied, newer closure undone, runtime state edited or goal marked verified.
