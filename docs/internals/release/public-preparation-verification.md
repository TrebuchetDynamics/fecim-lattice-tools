# Local public-preparation verification

Date: 25 October 2026. Starting HEAD: `816d17c7`.

**Historical first-pass receipt:** later corrections reclassified the Park fixture, excluded BTO, gated the display smoke test, and restored current adapters. See [the follow-up correction inventory](../../research/validation/park-reference-correction.md). The failures below describe the first pass, not an assertion that the corrected default suite still fails.
This is verification of a dirty working tree, not a release certificate.
Existing unrelated changes were preserved; no commit or publication was made.

## Scope

Relocated root internal notes and Fyne research under `docs/internals/`;
preserved instruction-discovery entry points and CI-required generated adapters;
untracked local editor/Pi state while retaining its local files; clarified
reproducibility limits; added citation metadata, an unsubmitted JOSS draft,
and evidence/readiness reviews.

Only the three local-state untracking deletions are staged:
`.pi/development-loop.json`, `.pi/development-loop/logs.jsonl`,
and `.vscode/settings.json`. Other changes remain unstaged/untracked.
This staging distinction is intentional; no commit authorization is implied.

## RED

Before moving notes or changing ingestion manifests:

```bash
go test ./scripts -run 'TestRepositoryRootKeepsCollateralUnderPurposeDirectories|TestKnowledgeIngestionIncludesRelocatedInternalNotes' -count=1
```

Exit 1, expected:
- Root-layout contract rejected `fyne-guide/` at the root.
- Ingestion manifests still named root `status.md`/`TODO.md` and the new input paths were absent.

## GREEN and focused verification

After relocation and manifest updates:
- Same focused command: exit 0.
- `go test ./scripts -count=1`: exit 0.
- `go test -race ./workbench/... ./scripts`: exit 0.
- `bash -n scripts/agent-test-loop.sh scripts/check-architecture.sh`: exit 0.
- Python AST parsing of both Cognee scripts: pass; Cognee itself was not imported or run.
- `uvx --from cffconvert==2.0.0 cffconvert --validate --infile CITATION.cff`: exit 0; CFF 1.2.0 schema valid.
- Local-link scan across edited navigation, moved notes, and new publication documents: 197 links checked, no broken local targets. Markdown heading anchors were not exhaustively validated.
- Draft YAML, bibliography keys, all eight required section headings, and body word count (932): pass. Official JOSS PDF rendering was not performed.
- `git diff --check` and `git diff --cached --check`: pass.
- SHA-256 checks of the 16 preserved unrelated modified/untracked files: pass. The existing README prerequisite edit was also retained.

TDD: N/A for prose, metadata, comments, and generated-file relocation; these do not change model or UI behavior. The ingestion-path/layout change used the RED/GREEN cycle above.

## CLI reproducibility check

Built with `go build -o artifacts/public-preparation/fecim-lattice-tools ./cmd/fecim-lattice-tools`, then copied the starter project to a fresh temporary directory.

```console
project validate: valid: hzo-study
experiment run -workers 2: total=8 reused=0 failed=0 feasible=6 pareto=1
experiment run -workers 4: total=8 reused=8 failed=0 feasible=6 pareto=1
report generate: generated reports
```

SHA-256 checks of all three report files were unchanged after worker-count reuse
and after standalone report regeneration. This verifies computational reproducibility,
not independent physics validation.

## Full verification is not green

`go test ./...` returned exit 1:

- `cmd/fecim-screenshotter-fyne/TestFullScreenshotterCapturesEveryModuleSequentially` failed while capturing hysteresis.
- X11 reported `BadValue` for `X_ConfigureWindow`.
- A focused rerun with `-count=1` reproduced the same failure.
- The screenshotter code/test were already modified before this preparation and remained checksum-identical. No claim is made about whether the cause is code, display environment, or their interaction.

`scripts/install-fecim-skills.sh --check` returned exit 1 because seven expected Claude adapters were already absent at session start. Those pre-existing deletions were not restored or attributed to this cleanup.

Do not tag or publish from this receipt as though these gates passed. The next action
is to resolve the GUI failure and adapter drift in their owning work, rerun the full
candidate verification, then obtain explicit release approval.

## Evidence and limitations

Disposable logs, preserved-file hashes, initial diff snapshot, CLI outputs, and the
fresh Park curve are under ignored `artifacts/public-preparation/`.
A fresh-context read-only reviewer found no blocker in the preparation scope; this
is not human peer review or JOSS acceptance. The literature research worker returned
useful findings but reported tool limitations; the parent independently inspected
the cited official JOSS, Zenodo, related-tool, Tour, IronLattice, and BTO article pages.

The [figure evidence review](../../research/validation/landing-page-figure-readiness.md)
records why neither inspected fixture is promoted to a headline experimental
validation plot. Public-history, research-use, complete AI disclosure, author
approval, and manuscript rendering remain submission gates in the
[readiness checklist](research-software-readiness.md).
