# Controlled local study fixtures (test-only)

These **invented synthetic values are not physics validation, device measurements,
or a human Orca PASS**. This fixture supplies only local manual tasks 7–9 from the
[Orca acceptance script](../../../../docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md).
It does not supply presentation/edit/conflict/restart/multipage/hosted-Worker/import
fixtures. Unprovided setups and actual Firefox/Linux/Orca, HYS and installed-tool
acceptance remain **NOT RUN**. No production endpoint, flag, schema or default
science changes. All three Go files are excluded unless `orca_fixture` is set.

## Build and bind the candidate

Run in this checkout, from the repository root. Use Go **1.26.10**, Node
**24.19.0**, npm **11.17.0**, and the existing lock file; see
[web setup](../../../../web/README.md). No browser engine, FFmpeg, EDA tool,
model, dataset, PDK or container acquisition is needed for these finite HTTP tests.
Playwright is not used here; real human acceptance requires the separately
provisioned Firefox/Orca environment.

```bash
: "${TMPDIR:?Set a private disk-backed temporary directory}"
: "${AUTOWORK_PORT_BASE:?Set your assigned private port base}"
export FECIM_ORCA_PORT="$AUTOWORK_PORT_BASE"
export FECIM_ORCA_ASSETS="$PWD/internal/browserui/assets"
# Only if dependencies are absent: (cd web && npm ci --cache "$AUTOWORK_CACHE/npm")
(cd web && npm run build:local)
# Under Autowork use autowork-run for builds/suites expected to exceed two minutes.
```

Extract the exact linker inputs, as `web/src/local/build-native.mjs` does.
Verify source freshness, not just agreement between two old manifests:

```bash
node --input-type=module <<'JS' > "$TMPDIR/orca-linker.env"
import {readFileSync} from 'node:fs';
import {execFileSync} from 'node:child_process';
const m = JSON.parse(readFileSync('internal/browserui/assets/local-manifest.json'));
const f = JSON.parse(execFileSync('node', ['src/build/build.mjs', '--fingerprint'], {cwd:'web'}));
if (f.build !== m.identity.build || f.UI !== m.UI || f.API !== m.API)
  throw new Error('Source drift: rebuild local assets');
console.log(`export ORCA_LDFLAGS='-X fecim-lattice-tools/workbench/hysteresis.BuildID=${m.identity.build} -X fecim-lattice-tools/internal/browserui.UIID=${m.UI}'`);
JS
. "$TMPDIR/orca-linker.env"
# Quote the entire GOFLAGS argument; retain the spaces within its linker value.
export GOFLAGS="'-ldflags=$ORCA_LDFLAGS'"
```

The harness uses `os.DirFS(FECIM_ORCA_ASSETS)` and production `New` asset
validation (local profile, Go/core identity, UI, API, engine and every asset hash).
It also rejects a native BuildID/UIID that does not match those actual assets.
Rebuild assets and re-extract linker inputs after **any Go source** change,
including test-only files. Do not substitute dummy assets or a copied old stage.

## Finite offline contracts (no tester, no indefinite server)

From the repository root with the preceding environment:

```bash
env CGO_ENABLED=1 go test -tags=orca_fixture ./workbench/study ./workbench/server -run 'TestOrcaFixtureContract' -count=1
env CGO_ENABLED=1 go test -race -tags=orca_fixture ./workbench/study ./workbench/server -run 'TestOrcaFixtureContract' -count=1
env CGO_ENABLED=1 go test ./workbench/study ./workbench/server -count=1
go list -f '{{.GoFiles}}' ./workbench/study
go list -f '{{.TestGoFiles}}' ./workbench/server
# A narrow repeat of actual-assets start/stop and all terminal controls:
env CGO_ENABLED=1 go test -tags=orca_fixture ./workbench/server -run '^TestOrcaFixtureContract/real-assets-controls$' -count=1
git diff --check
```

Tests create disposable roots under `TMPDIR`, bind only `FECIM_ORCA_PORT`, and
join workers, HTTP handlers and listeners before returning. They launch the same
harness as the interactive entrypoint. The default file lists must exclude
`orca_fixture.go`, `orca_fixture_test.go` and `orca_fixture_contract_test.go`.
The full merged check is the coordinator's gate, not a second worker check.

## Interactive launch

Use a compiled Go test binary so stdin is the controlling terminal (the `go test`
wrapper does not forward interactive stdin). The binary is not the production
application. Run it **from `workbench/server`**, where the fixed manifest lives:

```bash
export FECIM_ORCA_ROOT="$(mktemp -d "$TMPDIR/orca-manual.XXXXXX")"
env CGO_ENABLED=1 go test -c -tags=orca_fixture -o "$TMPDIR/orca-fixture.test" ./workbench/server
(cd workbench/server && FECIM_ORCA_FIXTURE=1 "$TMPDIR/orca-fixture.test" -test.run '^TestOrcaFixtureServer$' -test.v -test.timeout=0)
```

The root must be absolute, existing, empty and disposable. The harness generates
six scenario projects under `generation-1/granted/` and a separate
`outside-ungranted/` directory which is **not** in the grant. Open the printed
single-use launch link in Firefox; navigate to Projects and open the named
synthetic scenario from the granted root. Run with **one worker** (the harness
also forces one) using the normal UI. No terminal command starts/cancels a job
on behalf of the human browser task. Never capture the launch link/token in
speech transcripts, logs, screenshots or committed evidence.

Terminal commands, each followed by Enter:

```text
status
advance success
release success
advance cancel-resume
release cancel-resume
renew
reset
shutdown
```

`advance SCENARIO` permits one evaluation; at most one advance can be pending.
`status` reports the zero-based point index and whether its checkpoint is held,
plus actual job state/counts; it also writes safe identity evidence. Wait for the
next held checkpoint before advancing again. `release SCENARIO` permanently opens
that scenario's gates for the current generation, even after context cancellation.
`renew` alone prints a fresh secret launch link; the evidence writer never does.
`reset` first revokes admission, cancels jobs, releases checkpoints and joins all
workers/requests/streams, then creates a fresh generation and server session.
Sign in again with `renew`. Old project bytes and generation identity records are
retained, not overwritten. `shutdown` and stdin EOF perform the same settled
cleanup; neither deletes evidence. Inspect evidence before removing **only your
own** disposable root and test binary.

## Scenarios and independently fixed expectations

[`expected.json`](expected.json) is the oracle, not an Analyze-generated golden.
The synthetic evaluator is separate arithmetic. Sweep order is conductance
levels `[16,30]`, rows `[32,64]`, ADC bits `[4,6]`; the last axis varies fastest.
Point `i` has seed `17+i`. Model versions are `orca-fixture-v1-SCENARIO` so
scenario records cannot be confused/reused across roots. Fixed run IDs were
independently SHA-256 calculated from the canonical schema-1 identity JSON
(design, seed, model/evaluator versions, `inputs:null` for the empty input set),
not collected from a run.

Energy is `i+1 pJ`, latency `i+10 ns`, area is `(i+1)*100 um2` for indices 0–5,
then `6000` and `7000 um2`. Both objectives minimize; area must be `<=5000 um2`.
Thus indices 0–5 are feasible and only index 0 is Pareto. Every successful metric
has model `SYNTHETIC ORCA FIXTURE v1`, evidence `simulation-default` and assumption
`Invented acceptance values; not physics or measurements`.

| Project | Preparation and expected terminal result |
| --- | --- |
| `success` | Start in browser, then `release success`: 8 committed, 6 feasible, 1 Pareto, 0 failed; succeeded/complete. |
| `design-failure` | `release design-failure`: index 7 commits an intentional `synthetic-domain` failure with no fabricated metrics; 8 committed, 7 successful, 6 feasible, 1 infeasible, 1 Pareto, 1 failed; job succeeded/complete. |
| `cancel-resume` | Start in browser. Enter `advance cancel-resume` three times, waiting for held indices 1, 2, then 3. At held index 3 exactly indices 0–2 have committed. Cancel in browser: acknowledgement is `cancelling`, **not terminal**. Enter `release cancel-resume`; settlement is cancelled/partial, 3 completed, 5 cancelled, 0 failed, 3 feasible, 1 Pareto. Resume via a fresh browser Run: 8 completed, exactly 3 reused, 6 feasible, 1 Pareto, 0 failed. |
| `completion-wins` | `release completion-wins`, wait for succeeded/complete before requesting cancel. The terminal success remains, without a cancellation request. This is the deterministic completion-first ordering; `cancel-resume` is the cancellation-first ordering. |
| `systemic-failure` | `release systemic-failure`: evaluator errors at index 3; failed/partial with three genuine records, 3 feasible, 1 Pareto, **zero** scientific failures/cancelled points. |
| `report-failure` | `release report-failure`: all eight commits finish, then the injected writer puts a regular file at `reports` and invokes the actual `report.GenerateRoot`, whose directory creation fails. Job failed but results complete: 8 successful, 6 feasible, 1 Pareto, 0 failed. No report JSON is fabricated. Reset for a fresh writer scenario. |

These scenarios use real auth, rooted commits/integrity checks, reuse,
reconciliation, report generation/failure, SSE and production terminal job
projection. They never construct replacement Job snapshots. Unresolved work is
not relabeled as a failed design point.

## Runtime identity and evidence

Run `status` after each observation and before reset. `identity.json` at the root
and in the current `generation-N/` records fixture revision/expected-manifest
hash, candidate Git revision, repository-relative tracked fixture diff hash,
exact hashes of all five fixture sources (including untracked files), native test
binary hash, full local manifest (core/Go/UI/API and asset hashes), original
project document digests, expected scientific IDs, actual committed records and
real job snapshots with input/results revisions. The current root record follows
the latest generation; each older generation keeps its final settled record.
Random project/job handles in that record are not launch credentials. The
record contains neither root paths nor cookies/session/bootstrap tokens. Keep
private raw evidence outside Git; do not commit generated assets, test binaries,
launch links or workspace paths. On a dirty checkout, use source and binary hashes,
not the Git revision alone, to identify the tested candidate. The
`tracked_fixture_diff_sha256` field covers only the five named tracked fixture
paths against HEAD; Git omits untracked files. `fixture_source_sha256` binds their
actual bytes regardless of index state. Git evidence command failures fail the
harness rather than producing a success-shaped empty digest.

Actual human tasks must still record tester, OS/Firefox/Orca/synthesizer settings,
keyboard layout, speech, focus and per-task outcome. Automated fixture contracts
only establish engineering setup, never speech or accessibility qualification.
