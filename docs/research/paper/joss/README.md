# JOSS working draft

Start with [paper.md](paper.md) and [paper.bib](paper.bib). This is an **unsubmitted
working draft**, not an acceptance claim or a finished eligibility assessment.
It is separate from the parent directory's LaTeX tools-paper scaffold.

## Submission blockers

Follow the [release and JOSS readiness checklist](../../../internals/release/research-software-readiness.md).
In particular, establish public development history and actual research use;
confirm author/affiliation, funding and conflicts; reconcile the complete AI
tool/model/version history; obtain the author's own review attestation; and
render the paper with the official JOSS tooling. Do not remove these caveats
by substituting unsupported claims.

The current paper cites the inspected official CrossSim and AIHWKit documentation.
A final related-work review should add the appropriate software publications with
verified bibliographic metadata. The source records in the repository are leads,
not automatically verified bibliography entries.

## Source map for manuscript review

| Claim | Live source |
|---|---|
| Strict project loading and contained inputs | [project/load.go](../../../../workbench/project/load.go#L23) |
| Run identity fields | [experiment/identity.go](../../../../workbench/experiment/identity.go#L12) |
| Bounded execution and stored failures | [experiment/runner.go](../../../../workbench/experiment/runner.go#L36), [failure test](../../../../workbench/experiment/runner_test.go#L905) |
| Worker determinism | [runner test](../../../../workbench/experiment/runner_test.go#L966) |
| Verified run storage and reuse | [experiment/store.go](../../../../workbench/experiment/store.go#L30) |
| Current evaluator and its unused seed | [fecim/evaluator.go](../../../../workbench/fecim/evaluator.go#L16) |
| Direction-aware analysis | [experiment/analyze.go](../../../../workbench/experiment/analyze.go#L28) |
| Reports and determinism checks | [report/report.go](../../../../workbench/report/report.go#L45), [report test](../../../../workbench/report/report_test.go#L45) |
| Shared study operations | [study/study.go](../../../../workbench/study/study.go#L172) |
| Figure evidence limits | [figure readiness review](../../validation/landing-page-figure-readiness.md) |

Public format and policy references:
[JOSS paper format](https://joss.readthedocs.io/en/latest/paper.html) and
[submission requirements](https://joss.readthedocs.io/en/latest/submitting.html).

Line anchors are navigation aids, not pinned citations. Recheck this source map
whenever the evaluator, storage schema, study workflow, or manuscript changes.
