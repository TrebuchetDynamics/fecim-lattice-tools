# Scheduled broad physics race checks

The `physics-race` job in [nightly.yml](../../../.github/workflows/nightly.yml)
runs on the existing daily schedule and manual workflow dispatch, independently
of the normal nightly build/tests. It replaces the old broad short-mode race
step embedded near the end of that serial job.

- Job timeout: **90 minutes**.
- Race-test step timeout: **75 minutes**.
- Go per-package timeout: **60 minutes**.
- Scope: `./shared/physics/... ./module1-hysteresis/... ./shared/crossbar/...`.
- Full tests: `-race -count=1`, without `-short`.
- Display variables are empty; native Fyne build dependencies are installed.
- Failures are not masked. The log pipeline enables `pipefail`, and the log
  artifact is uploaded with `if: always()`.

The existing PR short race checks remain in place. The scheduled lane is the
long-running physics/determinism safeguard, not a substitute for focused
race checks when physics or shared-state behavior changes.

Run the same broad command on a suitably provisioned host:

```bash
env -u DISPLAY -u WAYLAND_DISPLAY go test -race -count=1 -timeout=60m \
  ./shared/physics/... ./module1-hysteresis/... ./shared/crossbar/...
```

Workflow contract tests check the trigger, dedicated job, required scope, absence
of short-mode exclusions, longer timeout, and log retention. They do not prove a
hosted CI run passed. No workflow was dispatched during local preparation.
