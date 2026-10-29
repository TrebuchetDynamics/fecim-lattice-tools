+++
schema = 1
id = "HYS-PROGRAM"
title = "Preserve hysteresis programming and calibration as bounded Go workflows"
short = "ISPP and level calibration"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 2000
depends_on = ["HYS-CORE"]
touches = ["workbench/hysteresis/programming.go", "workbench/hysteresis/programming_test.go", "workbench/hysteresis/calibration.go", "module1-hysteresis/pkg/gui/sim_loop.go", "module1-hysteresis/pkg/gui/calibration_runtime.go", "module1-hysteresis/pkg/gui/calibration_store.go", "module1-hysteresis/pkg/controller/", "module1-hysteresis/pkg/algo/", "module1-hysteresis/README.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement the PROGRAM seam over HYS-CORE for manual target writes, queued write/read targets, existing ISPP method/domain selection and 2–64 levels/target-range calibration. Extract GUI/controller/algo orchestration without duplicating equations; explicit seeds, drive ownership, bounded advance/cancel, fresh/stale/unavailable calibration and in-memory operation, never auto-save. Time-Resolved Switching is a legacy stub: return legacy_protocol_unavailable with a Go explanation and no fabricated trace; restoring a protocol is excluded."
acceptance = ["Given fixed target/seed/step sequences on both engines, when NewProgrammer/Apply/Advance executes, then phase/branch/level/queued target/guards/reset match existing deterministic controller/GUI behavior and manual field commands cannot override active programming; workbench/hysteresis/programming_test.go::TestProgrammingTrace.", "Given a completed calibration, when material/engine/level-count/target-range/temperature changes, then results become stale and cannot be exported/applied as fresh until recalibration; TestProgrammingCalibration reuses existing headless/shared calibrator expectations without writing files.", "Given cancellation/work cap or a time-resolved mode request, when the job advances/handles it, then terminal status is explicit, no partial result is labelled complete and the legacy stub returns legacy_protocol_unavailable without numerical samples; TestProgrammingCancellation and TestTimeResolvedUnavailable. No new switching protocol, global RNG or file logging."]
checks = ["go test ./workbench/hysteresis -run 'TestProgramming|TestTimeResolved' -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/controller ./module1-hysteresis/pkg/algo ./module1-hysteresis/tests -count=1", "env CGO_ENABLED=1 go test -race ./workbench/hysteresis -count=1"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HYS-PROGRAM.toml"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: NewProgrammer(*Session, ProgramConfig), Apply/Advance/Snapshot; typed phase/target/calibration data and legacy_protocol_unavailable time-resolved status sharing CORE canonical/presentation vocabulary. No operation registry edits.
- Consumes: HYS-CORE NewSession/drive ownership, revision-bound Snapshot and bounded advance API; existing controller/calibration and native GUI parity tests.
- Repo obligations triggered: Plan behavioral inventory level click ownership, 500 ms legacy timer versus explicit in-memory recalibration, scientific 30-level caveat; module1 README explains new workflow. Existing controller convergence/golden guards remain. No schema/manifest or generated outputs.
- Next action: Write TestProgrammingTrace reproducing a target write and reset with explicit dt/seed; confirm RED, then reuse the current controller.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.

- Builder finding resolution: P1 resolved from sim_loop.go:981–985 and cleared-only data arrays: preserve unavailable stub, not invent a new switching protocol.
