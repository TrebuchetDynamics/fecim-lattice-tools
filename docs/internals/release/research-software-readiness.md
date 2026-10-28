# Research software release and JOSS readiness

Status: release preparation and planning only; no release, DOI registration, or submission performed by this planning pass.
The original preparation evidence was checked 25 October 2026; the post-merge frontier below is bound to `258e8ab6`. Recheck live policies before publication.

The [first-pass verification receipt](public-preparation-verification.md) records the original GUI and adapter failures. The [follow-up correction audit](../../research/validation/park-reference-correction.md) records the Park evidence reclassification and BTO exclusion. Recheck the [follow-up verification and full diff](reference-correction-verification.md) before release; local preparation does not establish scientific validation or JOSS eligibility.

## What is established

- MIT license: [LICENSE](../../../LICENSE).
- Citation metadata: [CITATION.cff](../../../CITATION.cff); no invented version, release date, ORCID, or DOI.
- An existing local `v1.0.0-rc1` tag points to `e29dd222` (14 February 2026). Preserve it. The owner's subsequent ordered goals explicitly select **`v0.1.0`** as the first citable release; that choice is settled, not a new version interview.
- GitHub's releases endpoint returned an empty list during this review.
- GitHub reports repository creation on 26 April 2026. This does **not** prove when it first became public.
- The eight-run example demonstrates reproducibility, not calibrated physics or measured performance.
- [Internal notes](../README.md#internal-notes-and-publication-preparation) have moved; instruction entry points and CI-required generated adapters remain.
- A [JOSS manuscript draft](../../research/paper/joss/paper.md) exists separately from the older LaTeX manuscript. It is not submission-ready.

Read-only checks used:

```bash
git tag --list
git show -s --format='%h %as %s' v1.0.0-rc1
gh api repos/TrebuchetDynamics/fecim-lattice-tools/releases
gh api repos/TrebuchetDynamics/fecim-lattice-tools --jq '{created_at,visibility}'
```

## Release gate — maintainer approval required

- [x] Version intent settled by the owner: `v0.1.0`; preserve the earlier `v1.0.0-rc1` tag. This is not a published release or reserved DOI.
- [ ] Review the complete candidate diff, including pre-existing work; confirm authorship and affiliation in the citation file.
- [ ] Record full-suite, applicable race, CLI reproduction, and external-tool verification against the exact candidate commit. Old status notes are not fresh evidence.
- [ ] Review third-party data redistribution and the existing [public release boundary](PUBLIC_BOUNDARY.md).
- [ ] Approve the release notes and exact commit, then create the tag and GitHub release as separate authorized actions.
- [ ] Add actual version and release date to citation metadata as part of release preparation; keep them consistent with the published release.

## Zenodo gate — account owner action required

Enable the repository in Zenodo **before** publishing the release intended for automatic archival. Follow the official [repository enablement](https://help.zenodo.org/docs/github/enable-repository/) and [release archival](https://help.zenodo.org/docs/github/archive-software/github-upload/) instructions.

- [ ] Account owner authorizes the GitHub integration and enables this repository.
- [ ] Validate `CITATION.cff` and confirm archive metadata.
- [ ] Publish the approved GitHub release and verify that its archive record resolves.
- [ ] Record the actual version DOI and distinguish it from the concept DOI.
- [ ] Add the real DOI to citation metadata and README only after verification.

A Zenodo software DOI provides a citable archive; it does not imply peer review or scientific validation.

## JOSS gate — eligibility is not yet established

The current [submission requirements](https://joss.readthedocs.io/en/latest/submitting.html) require more than tests, documentation, and an open license:

- [ ] Establish **more than six months of public development**, with distributed, iterative activity. On the verification date, even the reported repository creation date was less than six calendar months ago; creation date is not public-history evidence.
- [ ] Document **actual research use**, at least by the developers: a named research workflow, question, inputs, outputs, and evidence of use. A deterministic example or intended audience alone does not establish research impact.
- [ ] Review open-source workflow, support expectations, governance, contribution guidance, CI, changelog, and release evidence. Stars and forks are not eligibility gates.
- [ ] Complete and human-review the [draft and bibliography](../../research/paper/joss/README.md), including a defensible comparison with related tools and why this project is a separate contribution.
- [ ] Provide the complete AI tool/model/version inventory and assistance scope for code, tests, docs, and manuscript. The human author must personally verify and attest to review, validation, and core design decisions.
- [ ] Confirm authorship, affiliations, funding, and conflicts of interest; do not infer these from Git history.
- [ ] Render the manuscript with JOSS tooling and check citations, word count (currently 750–1750), and all required sections.
- [ ] Obtain explicit approval before submitting. JOSS's current policy prohibits AI-generated author/editor/reviewer conversational interactions except translation.

The [paper format](https://joss.readthedocs.io/en/latest/paper.html) requires Summary, Statement of need, State of the field, Software design, Research impact statement, AI usage disclosure, Acknowledgements, and References. A validation plot can support the software's evidence story, but does not substitute for eligibility.

## Validation and outreach gate

- [ ] Resolve the [figure provenance issues](../../research/validation/landing-page-figure-readiness.md) before adding a headline experimental overlay.
- [ ] Keep fitting/calibration and held-out validation distinct, and report units, sample conditions, branch alignment, extraction uncertainty, RMSE definition, and reproducible generation command.
- [ ] Refresh Tour/IronLattice primary-source research immediately before drafting outreach. Company marketing and interviews are not measurement publications; an HZO model must not be described as calibrated to an unspecified IronLattice device.
- [ ] Approve any email separately; nothing in this preparation sends messages.

Update this checklist when JOSS policy, release state, figure evidence, or research-use evidence changes.

## First citable release frontier

Status: Accepted

**Authority and size:** Standalone Plan For Me discovery, producing one small release frontier inside the owner's already accepted seven-goal order. The owner chose `v0.1.0`, a real DOI and matching `CITATION.cff`, not new physics or a new product direction. The two engineering repairs below are planner decisions, not separately claimed owner design choices. Planning saves documents/tasks and one local planning commit only; it does not implement, push, publish, contact anyone or dispatch workers.

**Isolation:** Base `258e8ab6610de17ece9f5cd76bd8f94e5ebf4c36`. The actual consumer's runner is running, reads `tasks/`, and had zero active workers and 15 closed tasks at inspection. Save this batch on `planning/citable-release-frontier` in `.worktrees/planning-citable-release`, not on consumer `main`: installing new worker records on that continuous runner would risk execution beyond this standalone request. The committed branch is a durable pending-integration handoff, not an automatically delivered consumer queue. No service is stopped or reconfigured, no lease is cleared, and no hard sandbox is claimed.

### Project and accepted-outcome inventory

FeCIM is a local Go scientific simulation/reproducibility workbench for researchers, graduate students and device engineers. Plain project YAML, Go-owned calculations/evidence/run identities and immutable run/report files are authoritative. CLI studies and seven retained Fyne expert workspaces currently run. React owns the accepted future browser presentation/API boundary; `web/` is a separate static Astro marketing site, not that client.

| Owner outcome / class | Implementation and evidence at this base | Disposition |
|---|---|---|
| 1. Land reviewed preparation — Accepted | Public root cleanup/Park correction delivered; all local branch history merged; main CI run `40267125983` passed at this exact base. See `all-main-integration-verification.md`. | Delivered; no duplicate shipping tasks. |
| 2. First citable `v0.1.0` release — Accepted | Tutorial timing repair, headless lifecycle fixes and long scheduled physics-race job are implemented. GitHub releases API still returns `[]`; `CITATION.cff` has no version DOI. Existing release/DOI gate remains unmet. | Highest-priority remaining outcome; this three-task frontier. Configuration of the scheduled job is not proof that its full run passed. |
| 3. First real experimental reference — Accepted | `HZO-REF-1..5` are closed with coordinator receipts; Kim extraction and unfitted export exist. Export explicitly says `experimental_validation=false`; legacy runner reports zero experimental validations. Topics 1–3/specific-check re-research and independent visual/source qualification remain distinct. | Preserve completed IDs; do not repeat digitization or promote comparison into validation. Follow the accepted goal order after release. Separate stopped Guo research worktree remains untouched. |
| 4. Every physics constant cited or labeled — Accepted | `make research-audit` checks two registered claims, not exhaustive constant coverage. No zero-unclassified-constant certificate is established by it. | Later evidence-dependent constants frontier; no model/constant edits here. |
| 5. Accessible browser workbench — Accepted | `BWC-FOUNDATIONS`, `BWC-CONFIG`, `REACT-UI-POLICY` and `RDW-LIFECYCLE` have 10 closed implementation tasks with valid receipts. Core reads/edits/progress/speech primitives exist; React/API/security integration and Orca manual tasks are not delivered. | Keep their task IDs/status history; no reimplementation. A working Go speech string is not an Orca pass. |
| 6. JOSS submission — Accepted | Draft exists; actual submission/author attestations, current eligibility/research-use and final tagged-release checks remain unmet/unverified. | Deferred to its named goal; do not submit or invent eligibility now. |
| 7. Outreach — Accepted | Requires goals 2 **and** 3; no sent-email evidence. | Preserve dependency and publication/contact boundary. No email or outreach-copy task in this frontier. |

The five existing goal records still say unmet/partial despite their 15 closed tasks. `prepare` reports `GOAL_WITHOUT_WORK`; this is not permission to mark broader goals met or invent verification. Leave those source/history records unchanged in this narrow frontier. Its legacy `TODO.md` diagnostics mostly name preserved historical documents or negative layout tests; inspect meaning rather than mechanically replacing every token. The closed T-001..T-010 and HZO-REF-1..5 records remain byte-identical.

**Non-goals / anti-goals:** no Fyne retirement/default UI switch, hosted compute, WASM demo, new physics modules before goals 3/4, separate arXiv paper, fabricated DOI, weakened boundary/CI gate, republished unreviewed research-session state or implied experimental validation. Do not resurrect root agent/editor folders. No strategic pivot is proposed.

### Source observations and confidence

- **OBSERVED, high:** `launch.sh:29` hardcodes `CGO_ENABLED=0`, while `cmd/fecim-lattice-tools/fyne_nocgo.go:16–20` rejects desktop startup. Thus its advertised default Fyne launch cannot work after that build. This is static control-flow evidence; the destructive/building launcher was **not run** during planning.
- **OBSERVED, high:** the public boundary checker and PDF-audit bootstrap read removed `docs/public-release/THIRD_PARTY_PDF_AUDIT.csv`. The real owner is `docs/internals/release/THIRD_PARTY_PDF_AUDIT.csv`.
- **OBSERVED, high:** the real audit contains **79 unique rows** (51 keep / 28 conditional), while Git tracks **76 PDFs**. Exact current-path matches are zero; the exact `docs/4-research/` → `docs/research/` prefix migration matches all 76, with three historical extra rows. This says nothing new about the correctness of their rights evidence.
- **OBSERVED, high:** inspected read-only `PYTHONDONTWRITEBYTECODE=1 python3 scripts/public-release/check_public_release_boundary.py` exits **1**, reporting 76 missing audit rows and `rg` exit 2 for removed `docs/2-learn`, `docs/3-develop` and `docs/4-research`. No fix, bootstrap rewrite or test generation was run.
- **DOCUMENTED:** the public boundary permits source, audited third-party material and release-policy/audit documents; it does not grant a blanket MIT license to third-party PDFs. Existing rights decisions/obligations must survive path repairs unchanged.
- **UNKNOWN:** Zenodo account/repository enablement, final release candidate, author confirmation and disposition of any genuine boundary findings after repair. These block only publication, not the two offline repairs.

Input SHA-256 bindings: `launch.sh` = `6234222331d0324ebf21a5a85d825268028e9b7d71240759dab093f7f24a9d7c`; non-CGO adapter = `34c329c226d94cfb22c84a5fb95c9dd5d6920d28fc75259eed79701231d239f7`; boundary checker = `682c60aa5220f98ef03bd36302aeffd4f651ac63db8946d27ba09fc2a04fa70c`; bootstrap = `abbd1aeb4d39417ed4817928f605bc2189cb43253dee018c0ab236552a59392d`; audit CSV = `5fb67855d8f96c0682fd3f8c4cffa937288918aa793da8ccc27ae64494d89768`.

### Decisions and behavior contracts

| Choice | Planner's call and reason | Rejected alternative / boundary |
|---|---|---|
| Desktop launcher | Default an unset/empty `CGO_ENABLED` to `1`, preserving an explicit caller value, including `0` for headless commands. Keep command/argument forwarding, `--clear` scope, build failure behavior and application exit status. Use subprocess tests against a copied launcher and a fake Go executable in a temporary tree. | Do not force CGO off, add a browser launch, silently fall back to a headless build, open a real display in unit tests, or broaden cleanup paths. |
| Audit relocation | Move audit consumers to the canonical owner; migrate only the exact legacy research prefix in CSV path cells. Preserve every other cell, all 79 decisions and historical rows; reject ambiguous duplicate normalized keys. Bootstrap preserves existing records and defaults genuinely new PDFs to `replace-with-link`. | Do not alias old directories, regenerate every decision from defaults, guess matches by basename, delete PDFs, or invent redistribution evidence. |
| Boundary scan | Use existing canonical guide/module/internal/research roots and equivalent migrated blocked paths. Disallowed tracked paths/PDF checks always apply. The policy definition `PUBLIC_BOUNDARY.md` and removal-policy log may be excluded **only from phrase scanning** because they state prohibited terms; fixture tests must prove an otherwise identical ordinary document is still blocked. | No blanket skip of internal/research docs, no dropping `rg` errors or actual violations. Other genuine content/rights findings remain visible and keep the release blocked; correcting policy or science is not smuggled into this task. |
| Publication metadata | Use the existing `CITATION.cff`, not a new `.zenodo.json` override. Preserve `v1.0.0-rc1`; publish only the chosen `v0.1.0` candidate once qualified. Enable Zenodo before the intended GitHub release; verify the resulting **version DOI**, separately recording concept DOI. Add the actual DOI to main only after the record resolves and its version/source match. | Never synthesize a DOI, rewrite the old tag, infer current account enablement, or claim that archival means peer review. Automatic initial archival can precede the post-release DOI metadata commit; do not pretend that minted DOI was already in the tagged archive. |
| Frontier size/order | Two independent offline vertical repairs plus one owner-only publication/metadata task. Both repairs precede candidate qualification; the live task is pending, not an execution permission prompt. | No artificial launcher↔audit dependency, CI/manifest overhaul, full research re-run or automatic successor-planning daemon. |

Primary external evidence inspected in this pass: [enable a repository](https://help.zenodo.org/docs/github/enable-repository/), [archive a GitHub release](https://help.zenodo.org/docs/github/archive-software/github-upload/), and [CITATION.cff metadata](https://help.zenodo.org/docs/github/describe-software/citation-file/). Zenodo documents repository enablement before new-release ingestion, CFF validation, and `.zenodo.json` precedence over CFF. No account integration was attempted.

### Saved frontier and verification

| Saved ID | Placement / ownership | Outcome / real prerequisites |
|---|---|---|
| `T-011-29e9e9a4` (A) | Now / worker | Compatible desktop launcher; independent of B. |
| `T-012-41e51a72` (B) | Now / worker | Canonical fail-closed boundary/audit gate; independent of A. |
| `T-013-a31914c8` (C) | Needs decision / owner, `live` denied | Qualified `v0.1.0` archival and actual DOI metadata; verified dependencies on A and B plus account/candidate/rights inputs. |

All serve `CITABLE-RELEASE` (priority 1; no focus takeover). Existing priorities/holds/receipts stay unchanged. The complete requested discovery frontier is these three records, not a claim to have decomposed every later product milestone; no successor-planning task is needed for this finite batch.

- **A — proposed/unrun:** `scripts/launch_contract_test.go::TestLaunchPreservesDesktopAndHeadlessModes` executes the actual copied shell script with controlled inputs. It covers default/explicit CGO settings, spaced arguments/clear handling, build failure without launch, and application exit propagation. Required checks include that selector, `bash -n launch.sh`, and the existing non-CGO experiment/report end-to-end check. Update current build/run docs, not an invented React command.
- **B — proposed/unrun:** `scripts/public-release/test_public_release_boundary.py::TestPublicReleaseBoundary` exercises real checker/bootstrap behavior in temporary Git repositories: migrated audit/scan roots, missing/conflicting/denied/conditional decisions, exact rights-cell/history preservation, policy-definition self-exemption, real blocked content/generated paths and scan errors. Passing fixture checks prove the gate works, **not that this repository has redistribution clearance**. Run the actual gate separately and record its unmodified exit/findings in this owner; remaining genuine violations become publication blockers, not masked check successes.
- **C — owner-only:** qualify the exact candidate with full/race/CLI/security/boundary evidence and actual metadata; enable Zenodo, publish the selected tag/release and verify archive/DOI only under resumed external-operation authority. Keep authorship/account/rights findings on this task. Coordinator-owned changelog entry is part of release closeout; no worker contract writes `CHANGELOG.md` or `tasks/`.

The default full/affected execution guards remain the repository's AGENTS/CI requirements. New feature tests above do not exist yet and were **not run** in planning. No baseline score, speedup, legal clearance or passing new selector is fabricated. Implementers begin with the named RED test and make the minimum correction; a new test/fixture is their deliverable, not an external dependency.

**Pre-mortem:** caller CGO or quoted arguments are lost → A's real-shell subprocess cases; canonicalization erases rights or broadens exemptions → B's byte/cell preservation and negative ordinary-document cases; concept DOI or stale source is cited → C's explicit version/SHA comparison and schema/live checks. Global invariants: scientific calculations/units/evidence and existing runnable Fyne/CLI surfaces remain unchanged; no release is inferred from an offline gate.

**Builder disposition:** a fresh-context read-only builder check returned **“nothing important left.”** It confirmed disjoint A/B ownership, B's fixture-versus-clearance distinction, and C's verified prerequisites and separate owner/live verdict. No material questions were left pending on A/B; C's real external gates remain pending by design. The same dispositions are saved in the three task records. No implementation, installation, new feature test or runtime restart was performed.

**Handoff:** validate the final batch and commit once. Report branch-local selection separately from the actual consumer, which still has no new task until the planning commit is deliberately integrated under later execution authority. No implementation or workflow restart follows this standalone handoff.
