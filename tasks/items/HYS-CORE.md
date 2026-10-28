+++
schema = 1
id = "HYS-CORE"
title = "Preserve interactive hysteresis state in a UI-free Go session"
short = "Stateful Go hysteresis controls"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 1000
depends_on = []
touches = ["workbench/hysteresis/session.go", "workbench/hysteresis/session_test.go", "workbench/hysteresis/catalog.go", "workbench/hysteresis/view.go", "workbench/presentation/value.go", "workbench/presentation/value_test.go", "module1-hysteresis/pkg/gui/session_parity_test.go", "module1-hysteresis/pkg/gui/sim_loop.go", "module1-hysteresis/pkg/gui/physics_engine.go", "module1-hysteresis/pkg/gui/plot_view.go", "module1-hysteresis/README.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Extract reusable stateful material/model/manual/sine/triangle stepping into workbench/hysteresis over existing solvers, with Go control metadata, reset/invalidation rules, bounded advance/paged snapshots and canonical/display/speech values. Preserve minor-loop history and existing numerical behavior; characterize native GUI before extraction. Implement the CORE public seam and safety bounds in the cited plan; default catalogue selection stays legacy, initial Sine/paused is the delegated UI change."
acceptance = ["Given fixed material/model/input/step sequences, when native session and legacy GUI characterization execute, then E/P/history/reset/engine-control behavior agrees; add TestWorkbenchSessionCharacterization in module1-hysteresis/pkg/gui/session_parity_test.go and TestSession in workbench/hysteresis/session_test.go. No new coefficients or target-specific math.", "Given manual reversals, invalid/overflow settings or a stale page revision, when Apply/Advance/Page runs, then history is retained, invalid state changes are rejected atomically and 32-row revision-bound pages expose every retained sample; TestSession covers these branches and allocation/work limits.", "Given signed/zero/exponent SI E/P and unavailable values, when Go builds its snapshot, then canonical values remain intact and MV/cm uses 1e-8 and microcoulombs per square centimeter uses 100 with correct speech/reasons; extend workbench/presentation/value_test.go::TestHysteresisValues."]
checks = ["go test ./workbench/hysteresis ./workbench/presentation -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/gui -run 'TestWorkbenchSessionCharacterization|TestResetAllStateLocked_IsDeterministic|TestWaveformSwitch|TestLastCompleteCycle' -count=1", "env CGO_ENABLED=1 go test -race ./workbench/hysteresis ./workbench/presentation -count=1"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HYS-CORE.toml"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: hysteresis.NewSession(Config), Session.Apply(Command), Advance(Budget), Page(PageQuery); typed revision-bound Snapshot/Page with catalogue/control metadata, SI and Go presentation. See plan Go seams.
- Consumes: Existing ferroelectric/shared physics constructors, GUI deterministic UpdatePhysics(dt), material_default.go and plot_view.go behavior; no GUI import in new package.
- Repo obligations triggered: ADR 0005 Execution/identity and design Go presentation contract; update module1 README for session API/defaults. Preserve mode_switch/reset/plot characterization guards. Existing FormatMetric unsupported-unit behavior stays for unsupported units. New tests are proposed, not executed. No generated outputs/manifest writes in this task.
- Next action: Add TestWorkbenchSessionCharacterization/TestSession for a signed manual reversal and reset; run targeted RED before adapter implementation.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.
