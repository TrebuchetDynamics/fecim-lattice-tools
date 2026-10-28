# Local all-work delivery — 2026-10-28

## Authority and candidate

Owner request: **merge all to main**. Local integration only; no push, deployment, publication/DOI, branch deletion, service/model change or manual dispatch. Canonical source at inventory: `53fcd0ed185fdcf6d0737b61debb0e25c03b3285`; fetched origin/main is its ancestor (54 local-only commits, zero remote-only). Root checkout is clean; existing coordinator was idle with zero workers or pending integration. Validate in `.worktrees/merge-all-current`, then fast-forward canonical main under the existing Autowork writer/task locks with source/status rechecks.

## Included and preserved

- `planning/web-refresh` / `0aead65c`: current-source all-module roadmap and the same eight WEB IDs. Preserve completed HYS contracts/receipts, unmet owner-only Orca and final HYS planning stop. The earlier private `planning/react-hysteresis` / `93c1cae0` and `2fa495f8` are superseded by this refresh: do not replay their stale HYS-UI/HYS-LOCAL bodies.
- `planning/citable-release-frontier` / `21056c46`: existing first-release readiness plan, goal CITABLE-RELEASE and IDs T-011-29e9e9a4, T-012-41e51a72, T-013-a31914c8. Preserve priorities/independent offline repair lanes and owner-only Needs decision/live/publication gate. No release, DOI or rights clearance is created by delivery.
- Safe stopped-research documents from the retained review worktree: dispute entry, historical audit and small index. Current source still contains the disputed DOI; the saved exact-DOI metadata agrees with the historical mismatch finding. No new research, source replacement or scientific number change. Raw logs, machine/session/provenance state and unscreened retrieval outputs remain local, not published or promoted into the citation library.

## Explicit exclusions — retained, not discarded

- `autowork/HYS-CORE` at `16e93303` and `autowork/HYS-UI` at `08585c0a`: old initial candidates, not the later coordinator-verified implementations (`655cbb92` and `525eed4c`). Main also contains subsequent programming/analysis/transport/local/combined qualification work. Do not reverse-merge these obsolete alternatives.
- `autowork/T-013-447af517` at `16fdc47b` and `autowork/T-014-b8c042aa` at `dfeafc64`: old attempts superseded by receipt candidates `d19eb3eb` / integration `f220bdab`, and `356383db` / integration `4baa8df7`, respectively. Their source/test diffs are not patch-equivalent to the verified later candidates. Preserve them for recovery, not as competing production implementations.
- Dirty `.worktrees/autowork-HYS-WORKER-set-aside-{1,2}`: incomplete earlier transport/physics retries, superseded by the verified WORKER candidate and later qualification. No arbitrary staging or rollback.
- Dirty `.worktrees/staged-proposals-validation`: historical validation snapshot, not a branch-ready current candidate; its index differs from current main in 273 paths and would remove about 25,000 lines, including delivered features. Preserve index/worktree exactly.
- Raw `.worktrees/concurrent-research-review/citations/reports/fecim-research-2026-10-25/` evidence except the small tracked report index: logs, interrupted/unscreened retrieval and machine/session records are not safe source deliverables. Original dirty research worktree remains untouched; only reviewed Markdown is copied into the delivery candidate with historical qualifiers.

No ignored artifact, credential, binary, cache, code retry, existing receipt, goal-met declaration or default-launch change is staged. These exclusions are not claims that historical attempts were independently requalified or that all local branches now share ancestry.

## Verification

TDD N/A: the combined delivery changes planning, queue metadata and historical research documentation only; production source/tests/fixtures/CI/dependencies remain byte-identical to current main. Task validation, relative links, acyclic dependencies and canonical-ledger conservation are checked on the combined candidate. The mandatory unchanged-code full Go suite is run before committing. Its default skips do not enable production WASM parity or supply Orca evidence. Existing source-bound combined browser/Worker evidence remains attributed to T-015, not rerun here.

**Executed:** `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` PASS, exit 0, on the combined candidate (259 seconds; local transcript `/tmp/fecim-merge-all-fullcheck.log`). `autowork tasks validate --worktree .` PASS: 44 tasks, ten goals, no drafts. `git diff --cached --check` PASS. Independent fresh-context read-only review approved local integration and confirmed all 71 pre-existing task/goal/evidence files byte-identical, original IDs/acyclic prerequisites, resolving relative links, receipt-backed exclusions and no raw/private/executable files staged. Product source/tests/fixtures/CI/manifests/locks remain unchanged. The full-suite run truncated tracked `module6-eda/pkg/gui/tabs/data/fecim_crossbar_2x2/config.json` as a test side effect (initially unchanged, never staged); its exact original canonical bytes were restored after inspection, with no test/production repair or unrelated reset. Final source/status/index/ledger and lock-guarded canonical readback checks are performed at integration; no upstream push or new task dispatch is performed by this delivery.
