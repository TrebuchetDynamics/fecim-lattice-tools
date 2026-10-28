# WEB-WORKBENCH overlap disposition

DOCS-MAINTAINER inspection at `dba35de3ce3549f69d9e6995a7a88e57030ba09b`.

## Decision

Preserve the reported wait for `T-011-29e9e9a4` on `RELEASE.md`. No task contract, dependency, placement, closure or authorization changes are justified. This is scheduler admission, not a missing product prerequisite. Reassess admission after the launcher work lands; this note does not promise dispatch or bypass resource/protected-path admission.

## Source-backed diagnosis

- [WEB-WORKBENCH](../../tasks/items/WEB-WORKBENCH.md) declares `RELEASE.md` and owns current shared availability/build guidance. The [accepted roadmap, Current-source discovery reconciliation](../internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md#current-source-discovery-reconciliation--53fcd0ed) explicitly assigns README/RELEASE/CLAUDE/AGENTS factual closeout to this unstarted task. Removing that path or replacing it with a directory/glob would conceal required work.
- [T-011-29e9e9a4](../../tasks/items/T-011-29e9e9a4.md) also declares `RELEASE.md` for compatible desktop/headless launch guidance. Its [accepted release frontier](../internals/release/research-software-readiness.md#first-citable-release-frontier) requires current build/run documentation alongside the launcher repair. At initial inspection, its common-Git-directory `autowork-execution/slots/T-011-29e9e9a4.json` reported `running` without a completed report. During review it advanced to `integrating`. The now-available `slot-data/T-011-29e9e9a4/attempt-1/report.json` identifies candidate `25d8f98de8155d17fcf8c0c1c002ed14d43a7f93`, explicitly lists `RELEASE.md` among its five changed files and reports focused checks passing with no blockers. This corroborates the actual overlap; it does not qualify integration or authorize editing the live entry. No WEB-WORKBENCH slot/report exists at inspection.
- [T-013-a31914c8](../../tasks/items/T-013-a31914c8.md) is the third unfinished contract naming literal `RELEASE.md`. It remains owner-only in Needs decision with resource `live`. It counts toward file heat but is not itself running work reserving that file; its real verified launcher/boundary prerequisites remain intact.
- The supplied scheduler policy counts literal files named by at least three unfinished contracts, including noncandidates and unlanded live work. Installed scheduler `Src/pi_autowork/engine/parallel_plan.py:329–342,365–387` implements that count and the hot-file wait against active work. The three contracts above therefore explain the reported wait. Held/noncandidate ownership alone is not a file reservation.
- WEB-WORKBENCH's existing dependencies remain `HYS-UI` and `HYS-LOCAL`; both records are `done` with `coordinator_verified` closure. Neither its Sources nor accepted roadmap identifies the launcher as a producer needed for its implementation. Shared documentation alone does not justify adding that dependency. No source-backed cohesion, conflict-cost or testability evidence justifies splitting or merging these tasks solely to remove overlap.

## Verification and limits

`tasks validate --worktree .` passed again after review: 44 tasks, 10 goals, no drafts or pending records. The repository full check, `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`, exited 0 on this checkout. Local Markdown link targets resolve and `git diff --check` passed. An independent read-only reviewer found no blocker or justified contract repair. No task-file patch was applied, no completed HYS record reopened, and no goal qualification inferred. Task entries and their priority order remain unchanged.

TDD: N/A — this audit note changes no product behavior or task execution contract. Product, browser, scientific and manual qualification are not established by this diagnosis or ledger validation.
