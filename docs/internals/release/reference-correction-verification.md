# Reference correction — verification and review

Date: 25 October 2026. Base HEAD: `816d17c7`. No commit, tag, release, DOI registration, submission, or email was authorized or performed.

## Review scope

See [the scientific correction inventory](../../research/validation/park-reference-correction.md) and [the dated honesty-audit entry](../../research/honesty-audit.md#corrected-claims--2026-10-25).

- Park CSV and its direct/derived consumers are explicitly synthetic regression, not measured literature agreement. A second distinct Park-attributed JSON and tanh-fit test are also synthetic/non-experimental; their historical origin is not presumed identical to the CSV.
- Unsupported Park scalar extraction and the disputed BTO trace are quarantined; no reference samples were smoothed or retuned.
- Active reference artifacts number eight, with zero experimental validations. Historical BTO and duplicate legacy Park outputs cannot inflate the runner summary.
- Full screenshot smoke test is opt-in via `display`; ordinary PNG tests remain default. Current generated adapters were restored. `.pi/` and `.vscode/` are ignored; local files remain intact.
- Publication documents remain drafts. The current official [JOSS submission requirements](https://joss.readthedocs.io/en/latest/submitting.html) were reread directly; no eligibility or acceptance assertion was made.

## RED → GREEN evidence

Expected failures were observed before the corresponding implementation:

- `go test ./scripts -run TestFullScreenshotSmokeRequiresDisplayBuildTag -count=1`: GUI smoke was selected without a tag; now passes for both default and `display` file selection.
- `go test ./validation/literature -run 'TestParkFixtureDeclaresSyntheticProvenance|TestExcludedBTODigitizationCannotPassProvenanceGate|TestHZOForcReferenceDeclaresSyntheticTargets' -count=1`: synthetic fixture absent, BTO incorrectly accepted, and FORC targets not classified synthetic; now pass.
- `go test ./shared/physics -run 'TestParkPresetNamesDeclareSyntheticReference|TestLoadPELoopCSVSupportsSyntheticProvenanceHeader' -count=1`: preset names claimed measured calibration and reader rejected provenance comments; now pass.
- `go test ./cmd/calibrate-preisach -run TestParseCSVLoopAcceptsSyntheticHeader -count=1` and generator header test: missing comment/header support; now pass.
- `go test ./validation/calibration -run TestDefaultKilohertzTargetIsSimulationDefault -count=1`: target incorrectly cited unsupported extraction; now passes.
- Artifact-family contract: expected eight active references, observed nine before BTO exclusion; now passes and requires explicit Park synthetic/non-experimental metadata.
- Summary regression test: first failed on obsolete `rmse_full_scale`, then exposed duplicate legacy/current Park counting; now passes, reports eight references, excludes stale BTO, and suppresses duplicate Park output.
- Additional JSON consumer classification assertion failed against the old purported published-data fixture before relabeling. `TestFitSyntheticReference_RMSBelow10PercentScale` now passes with explicit metadata and unchanged numeric samples.
- Adapter sync initially failed for seven missing current adapters; regeneration restored all seven. Obsolete GoGPU adapter has no canonical skill or sync requirement; its pre-existing deletion remains untouched.

TDD: N/A for prose, citation metadata, generated-adapter restoration and metadata-only snapshot corrections. No physics coefficients or reference numeric arrays changed.

## Final checks

- `env -u DISPLAY -u WAYLAND_DISPLAY go test ./...`: **PASS**, exit 0. Log: `artifacts/reference-correction/go-test-final.log`.
- `env -u WAYLAND_DISPLAY xvfb-run -a go test -tags=display ./cmd/fecim-screenshotter-fyne -run '^TestFullScreenshotterCapturesEveryModuleSequentially$' -count=1 -v`: **PASS**, all-module smoke. Log: `artifacts/reference-correction/display-smoke.log`.
- `go test -race ./workbench/... ./scripts`: **PASS**.
- Focused race checks for synthetic-reader/fitting tests in `shared/physics` and `module1-hysteresis/pkg/ferroelectric`: **PASS**.
- Broad race invocations including complete physics packages exceeded a 180-second command limit; no successful full-physics race result is claimed for this final state. They were replaced by the focused changed-behavior checks, not silently counted as passes.
- `scripts/install-fecim-skills.sh --check`: **PASS**. Installer's eight fixture tests: **PASS**.
- `make test-research`: **PASS**. `make research-audit`: **PASS**, two claims, zero errors. An initial audit rejected the workbench-style `simulation-default` status; it was corrected to the citation system's existing `not-validated` status and rerun.
- Reference runner: **PASS**, `reference_artifacts=8 experimental_validations=0`; stale artifacts are excluded or superseded.
- Park CSV numeric body, BTO CSV bytes, second JSON numeric trace/legacy weights, and Park snapshot numeric metrics were compared with HEAD: **unchanged**. Misleading uncertainty-confidence metadata was removed, not numeric measurement samples.
- New local Markdown targets: **PASS** (229 checked in the final documentation scan); heading anchors were not exhaustively checked.
- `CITATION.cff`: CFF 1.2.0 schema validation **PASS**. JOSS draft YAML and word count checked; official PDF rendering remains unperformed.
- `git diff --check` and `git diff --cached --check`: **PASS**.

One full run under concurrent load failed the pre-existing `shared/widgets/TestTutorialControllerProgress` timing assertion (75 ms sleep, observed progress zero). A focused rerun and the subsequent complete headless suite passed. Its source was not changed; retain this as an intermittent-test observation, not a resolved widget bug. Log: `artifacts/reference-correction/go-test-transient-widget-failure.log`.

A fresh-context reviewer found two missed evidence paths (the second JSON and facts index); both were corrected and re-reviewed with no remaining blocker within this scope. This is code review assistance, not human scientific peer review.

## Diff and workspace handling

Complete local review artifacts, including untracked files and rename detection, are saved under ignored `artifacts/reference-correction/`:

- `full-diff-summary.txt`: full worktree stat and name-status listing against HEAD.
- `full-worktree.patch`: complete binary-capable diff for inspection.

These include earlier cleanup, pre-existing user edits, and concurrent research material; they are **not a proposed all-files commit**. Concurrent changes in `citations/disputed.md` and `citations/reports/fecim-research-2026-10-25/` belong to a separate research workflow and were preserved, not authored or resolved by this correction. Their separate Guo-source finding still requires its owning review.

Only the three previously staged local-state untracking deletions remain in the real index. The complete diff was assembled using a temporary index, without staging the remaining changes or altering the real index. Review before deciding any commit scope.
