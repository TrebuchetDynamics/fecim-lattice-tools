+++
schema = 1
id = "HYS-NEXT-PLAN"
title = "Reconcile Hysteresis qualification and plan only evidence-backed retirement work"
short = "Next Hysteresis planning frontier"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 7000
depends_on = ["HYS-CORE", "HYS-PROGRAM", "HYS-ANALYSIS", "HYS-WORKER", "HYS-UI", "HYS-LOCAL"]
touches = ["docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md", "docs/internals/adr/0005-unified-react-hosted-and-local.md"]
resources = []
kind = "plan"
ownership = "worker"
scope = "For goal HYS-REACT, read completed task reports/current source and remaining module proof after this build batch. Save only evidence-backed qualification repairs and a separately scoped Hysteresis Fyne retirement frontier if prerequisites are known; preserve owner HYS-ORCA if unrun. Inspect current GUI import/caller/test graph before defining deletions. Round 1 of 2: any successor is round 2 of 2, then automatic replanning stops. No other module or global launch-default expansion, no implementation inside planning."
acceptance = ["Given merged building outcomes and current code, when bounded Plan For Me reconciles the goal, then each remaining Hysteresis acceptance/retirement gap maps to complete existing/new records or a precisely scoped owner gap, with source/evidence and real dependencies; run read-only builder check and autowork tasks validate --worktree . before one planning commit.", "Given missing Orca or other qualification evidence, when next tasks are ordered, then no Fyne deletion is called qualified or made runnable before those gates, completed IDs/receipts are preserved and round-2 limit/consumer delivery are explicit. If no remaining building work is justified, save no speculative tasks and report remaining owner evidence."]
checks = ["Read accepted plan, completed HYS task reports, actual source/callers and candidate-bound validation/Orca evidence; perform the planning craft builder check.", "autowork tasks validate --worktree .", "git diff --check"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "manual"
closure_actor = "DOCS-MAINTAINER"
closure_timestamp = "2026-10-28T09:38:01.111282+00:00"
closure_reason = "Round 1 source/report reconciliation saved one final-candidate qualification task and exactly one round-2 successor; real Orca and Fyne retirement remain unqualified. Planning only, no implementation or verification receipt."
closure_evidence = "docs/tasks/HYS-NEXT-PLAN.md"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: One bounded successor planning batch or explicit remaining owner gap; no code/deployment/dispatch in planning.
- Consumes: HYS-CORE/PROGRAM/ANALYSIS/WORKER/UI/LOCAL verified reports and merged source; HYS-ORCA actual evidence if available, never a prerequisite for independent planning.
- Repo obligations triggered: Planning artifact/authority contract and single-pass handoff; queue writes only through coordinator autowork tasks flow, never code-worker queue edits. Preserve unrelated goals/IDs and every live qualification gap.
- Next action: Read finished HYS reports and compare actual feature/import graph with plan inventory; reconcile remaining acceptance without a new product interview.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.
- Plan round: 1 of 2
